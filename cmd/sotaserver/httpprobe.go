package main

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"
)

// defaultPaths is a small built-in list of interesting paths to probe. Supply
// -wordlist to use your own.
var defaultPaths = []string{
	"/", "/robots.txt", "/sitemap.xml", "/.well-known/security.txt",
	"/admin", "/admin/login", "/login", "/wp-login.php", "/wp-admin/",
	"/administrator", "/phpmyadmin/", "/server-status", "/.git/config",
	"/.git/HEAD", "/.env", "/.env.local", "/config.php", "/config.json",
	"/backup", "/backup.zip", "/backup.tar.gz", "/db.sql", "/dump.sql",
	"/api", "/api/", "/api/v1", "/api/v2", "/swagger.json",
	"/swagger-ui.html", "/openapi.json", "/actuator", "/actuator/health",
	"/metrics", "/health", "/status", "/debug", "/test", "/.DS_Store",
	"/web.config", "/.htaccess", "/crossdomain.xml", "/console", "/manager/html",
}

// probeResult is the outcome of probing a single path.
type probeResult struct {
	Path     string `json:"path"`
	URL      string `json:"url"`
	Status   int    `json:"status"`
	Length   int64  `json:"length"`
	Location string `json:"location,omitempty"`
	Error    string `json:"error,omitempty"`
}

// httpprobeOutput is the JSON shape of the httpprobe command.
type httpprobeOutput struct {
	Base      string        `json:"base"`
	Tried     int           `json:"tried"`
	Displayed []probeResult `json:"displayed"`
}

// runHTTPProbe performs content discovery by requesting a list of common paths
// against a base URL and reporting the response status for each. Requests are
// plain GET/HEAD calls; nothing is modified on the target.
func runHTTPProbe(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("httpprobe", flag.ContinueOnError)
	base := fs.String("url", "", "base URL, e.g. https://example.com - required")
	wordlist := fs.String("wordlist", "", "path to a newline-separated list of paths (defaults to a built-in list)")
	method := fs.String("method", "GET", "HTTP method (GET or HEAD)")
	workers := fs.Int("workers", 20, "number of concurrent requests")
	timeout := fs.Duration("timeout", 10*time.Second, "per-request timeout")
	insecure := fs.Bool("insecure", false, "skip TLS certificate verification")
	showAll := fs.Bool("all", false, "show every result, not just 'interesting' status codes")
	jsonOut := fs.Bool("json", false, "emit results as JSON")
	fs.Usage = func() {
		fmt.Fprintf(fs.Output(), "Usage: sotaserver httpprobe -url URL [flags]\n\n")
		fmt.Fprintf(fs.Output(), "Probe common paths for content discovery. Example:\n  sotaserver httpprobe -url https://example.com\n\nFlags:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *base == "" {
		fs.Usage()
		return fmt.Errorf("missing required -url")
	}
	if *workers < 1 {
		return fmt.Errorf("-workers must be >= 1")
	}
	m := strings.ToUpper(*method)
	if m != "GET" && m != "HEAD" {
		return fmt.Errorf("-method must be GET or HEAD")
	}

	root := strings.TrimRight(*base, "/")
	paths := defaultPaths
	if *wordlist != "" {
		l, err := readWordlist(*wordlist)
		if err != nil {
			return err
		}
		paths = normalizePaths(l)
	}

	// Do not follow redirects so we can report 3xx locations directly.
	client := &http.Client{
		Timeout: *timeout,
		Transport: &http.Transport{
			Proxy:           http.ProxyFromEnvironment,
			TLSClientConfig: &tls.Config{InsecureSkipVerify: *insecure},
		},
		CheckRedirect: func(*http.Request, []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	if !*jsonOut {
		fmt.Printf("Probing %d path(s) under %s (%s, %d workers)...\n\n",
			len(paths), root, m, *workers)
	}

	pathsCh := make(chan string)
	resultsCh := make(chan probeResult)
	var wg sync.WaitGroup

	for i := 0; i < *workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for p := range pathsCh {
				full := root + p
				res := probeResult{Path: p, URL: full}
				req, err := http.NewRequestWithContext(ctx, m, full, nil)
				if err != nil {
					res.Error = err.Error()
					resultsCh <- res
					continue
				}
				req.Header.Set("User-Agent", "sotaserver/1.0 (authorized-testing)")
				resp, err := client.Do(req)
				if err != nil {
					res.Error = err.Error()
					resultsCh <- res
					continue
				}
				res.Status = resp.StatusCode
				res.Length = resp.ContentLength
				res.Location = resp.Header.Get("Location")
				resp.Body.Close()
				resultsCh <- res
			}
		}()
	}

	go func() {
		defer close(pathsCh)
		for _, p := range paths {
			select {
			case <-ctx.Done():
				return
			case pathsCh <- p:
			}
		}
	}()

	go func() {
		wg.Wait()
		close(resultsCh)
	}()

	var shown []probeResult
	for r := range resultsCh {
		if !*showAll && !interesting(r) {
			continue
		}
		shown = append(shown, r)
		if !*jsonOut {
			printProbe(r)
		}
	}
	sort.Slice(shown, func(i, j int) bool { return shown[i].Path < shown[j].Path })

	out := httpprobeOutput{Base: root, Tried: len(paths), Displayed: shown}
	return emit(*jsonOut, out, func() {
		fmt.Printf("\n%d/%d path(s) shown.\n", len(shown), len(paths))
	})
}

// interesting reports whether a probe result is worth showing by default:
// anything that is not a plain 404 or a connection error.
func interesting(r probeResult) bool {
	if r.Error != "" {
		return false
	}
	return r.Status != http.StatusNotFound
}

func printProbe(r probeResult) {
	marker := "[+]"
	switch {
	case r.Status >= 500:
		marker = "[!]"
	case r.Status >= 400:
		marker = "[-]"
	case r.Status >= 300:
		marker = "[>]"
	}
	line := fmt.Sprintf("  %s %-3d %s", marker, r.Status, r.Path)
	if r.Location != "" {
		line += " -> " + r.Location
	}
	fmt.Println(line)
}

// normalizePaths ensures each path starts with a single leading slash.
func normalizePaths(in []string) []string {
	out := make([]string, 0, len(in))
	for _, p := range in {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if !strings.HasPrefix(p, "/") {
			p = "/" + p
		}
		out = append(out, p)
	}
	return out
}
