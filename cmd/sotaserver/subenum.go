package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"net"
	"os"
	"sort"
	"strings"
	"sync"
)

// defaultSubdomains is a small built-in wordlist of common subdomain labels.
// Supply -wordlist to use a larger list.
var defaultSubdomains = []string{
	"www", "mail", "smtp", "pop", "imap", "webmail", "ns1", "ns2", "dns",
	"ftp", "sftp", "vpn", "remote", "portal", "api", "api-v1", "dev", "staging",
	"stage", "test", "qa", "uat", "beta", "demo", "admin", "administrator",
	"cpanel", "whm", "blog", "shop", "store", "app", "apps", "mobile", "m",
	"secure", "login", "auth", "sso", "git", "gitlab", "jenkins", "ci",
	"docker", "registry", "db", "database", "sql", "mysql", "postgres",
	"redis", "cache", "cdn", "static", "assets", "img", "images", "media",
	"files", "download", "downloads", "docs", "wiki", "help", "support",
	"status", "monitor", "grafana", "kibana", "prometheus", "internal",
	"intranet", "corp", "vpn2", "gw", "gateway", "proxy", "mx", "mx1", "mx2",
	"autodiscover", "owa", "exchange", "lync", "sip", "voip", "pbx",
}

// subResult is one resolved subdomain and its addresses.
type subResult struct {
	Host      string   `json:"host"`
	Addresses []string `json:"addresses"`
}

// subenumOutput is the JSON shape of the subenum command.
type subenumOutput struct {
	Domain     string      `json:"domain"`
	Tried      int         `json:"tried"`
	Discovered []subResult `json:"discovered"`
}

// runSubenum discovers live subdomains of a domain by resolving each candidate
// label from a wordlist via DNS. This is active enumeration (it sends DNS
// queries) but is non-destructive.
func runSubenum(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("subenum", flag.ContinueOnError)
	domain := fs.String("domain", "", "base domain, e.g. example.com - required")
	wordlist := fs.String("wordlist", "", "path to a newline-separated subdomain wordlist (defaults to a small built-in list)")
	workers := fs.Int("workers", 50, "number of concurrent resolvers")
	jsonOut := fs.Bool("json", false, "emit results as JSON")
	fs.Usage = func() {
		fmt.Fprintf(fs.Output(), "Usage: sotaserver subenum -domain DOMAIN [flags]\n\n")
		fmt.Fprintf(fs.Output(), "Discover live subdomains via DNS resolution. Example:\n  sotaserver subenum -domain example.com -wordlist subs.txt\n\nFlags:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *domain == "" {
		fs.Usage()
		return fmt.Errorf("missing required -domain")
	}
	if *workers < 1 {
		return fmt.Errorf("-workers must be >= 1")
	}

	base := strings.TrimSuffix(strings.TrimSpace(*domain), ".")
	labels := defaultSubdomains
	if *wordlist != "" {
		l, err := readWordlist(*wordlist)
		if err != nil {
			return err
		}
		labels = l
	}

	if !*jsonOut {
		fmt.Printf("Enumerating %d candidate subdomain(s) of %s with %d workers...\n\n",
			len(labels), base, *workers)
	}

	resolver := net.DefaultResolver
	namesCh := make(chan string)
	resultsCh := make(chan subResult)
	var wg sync.WaitGroup

	for i := 0; i < *workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for host := range namesCh {
				ips, err := resolver.LookupHost(ctx, host)
				if err != nil || len(ips) == 0 {
					continue
				}
				sort.Strings(ips)
				resultsCh <- subResult{Host: host, Addresses: ips}
			}
		}()
	}

	go func() {
		defer close(namesCh)
		for _, label := range labels {
			host := label + "." + base
			select {
			case <-ctx.Done():
				return
			case namesCh <- host:
			}
		}
	}()

	go func() {
		wg.Wait()
		close(resultsCh)
	}()

	var found []subResult
	for r := range resultsCh {
		found = append(found, r)
		if !*jsonOut {
			fmt.Printf("  [+] %-40s %s\n", r.Host, strings.Join(r.Addresses, ", "))
		}
	}
	sort.Slice(found, func(i, j int) bool { return found[i].Host < found[j].Host })

	out := subenumOutput{Domain: base, Tried: len(labels), Discovered: found}
	return emit(*jsonOut, out, func() {
		fmt.Printf("\n%d/%d subdomain(s) resolved.\n", len(found), len(labels))
	})
}

// readWordlist loads non-empty, non-comment lines from a file.
func readWordlist(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var words []string
	seen := make(map[string]bool)
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		w := strings.TrimSpace(sc.Text())
		if w == "" || strings.HasPrefix(w, "#") {
			continue
		}
		w = strings.TrimPrefix(w, ".")
		if !seen[w] {
			seen[w] = true
			words = append(words, w)
		}
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	if len(words) == 0 {
		return nil, fmt.Errorf("wordlist %q is empty", path)
	}
	return words, nil
}
