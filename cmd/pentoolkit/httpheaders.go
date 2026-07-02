package main

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"
)

// securityHeader describes a response header that is relevant to a site's
// security posture, along with guidance shown when it is missing.
type securityHeader struct {
	name    string
	missing string
}

var securityHeaders = []securityHeader{
	{"Strict-Transport-Security", "no HSTS: connections may be downgraded to HTTP"},
	{"Content-Security-Policy", "no CSP: reduced protection against XSS/injection"},
	{"X-Frame-Options", "missing: page may be embeddable (clickjacking)"},
	{"X-Content-Type-Options", "missing: MIME-type sniffing not disabled"},
	{"Referrer-Policy", "missing: referrer may leak to third parties"},
	{"Permissions-Policy", "missing: browser feature access not restricted"},
}

// runHTTPHeaders fetches a URL and reports on security-relevant response
// headers, flagging which recommended headers are present or missing. It also
// highlights headers that disclose server/software versions.
func runHTTPHeaders(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("httpheaders", flag.ContinueOnError)
	url := fs.String("url", "", "target URL, e.g. https://example.com - required")
	method := fs.String("method", "GET", "HTTP method to use")
	timeout := fs.Duration("timeout", 15*time.Second, "request timeout")
	follow := fs.Bool("follow", true, "follow redirects")
	insecure := fs.Bool("insecure", false, "skip TLS certificate verification")
	all := fs.Bool("all", false, "print all response headers, not just security-relevant ones")
	fs.Usage = func() {
		fmt.Fprintf(fs.Output(), "Usage: pentoolkit httpheaders -url URL [flags]\n\n")
		fmt.Fprintf(fs.Output(), "Inspect HTTP security headers. Example:\n  pentoolkit httpheaders -url https://example.com\n\nFlags:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *url == "" {
		fs.Usage()
		return fmt.Errorf("missing required -url")
	}

	client := &http.Client{
		Timeout: *timeout,
		Transport: &http.Transport{
			Proxy:           http.ProxyFromEnvironment,
			TLSClientConfig: &tls.Config{InsecureSkipVerify: *insecure},
		},
	}
	if !*follow {
		client.CheckRedirect = func(*http.Request, []*http.Request) error {
			return http.ErrUseLastResponse
		}
	}

	req, err := http.NewRequestWithContext(ctx, strings.ToUpper(*method), *url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "pentoolkit/1.0 (authorized-testing)")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	fmt.Printf("%s %s -> %s\n\n", req.Method, *url, resp.Status)

	if *all {
		fmt.Println("All response headers:")
		names := make([]string, 0, len(resp.Header))
		for k := range resp.Header {
			names = append(names, k)
		}
		sort.Strings(names)
		for _, k := range names {
			for _, v := range resp.Header[k] {
				fmt.Printf("  %s: %s\n", k, v)
			}
		}
		fmt.Println()
	}

	fmt.Println("Security headers:")
	for _, h := range securityHeaders {
		if v := resp.Header.Get(h.name); v != "" {
			fmt.Printf("  [+] %-28s %s\n", h.name+":", v)
		} else {
			fmt.Printf("  [-] %-28s %s\n", h.name+":", h.missing)
		}
	}

	fmt.Println("\nInformation disclosure:")
	disclosed := false
	for _, name := range []string{"Server", "X-Powered-By", "X-AspNet-Version", "X-AspNetMvc-Version", "Via"} {
		if v := resp.Header.Get(name); v != "" {
			fmt.Printf("  [!] %-28s %s\n", name+":", v)
			disclosed = true
		}
	}
	if !disclosed {
		fmt.Println("  none of the common version-disclosing headers are set")
	}
	return nil
}
