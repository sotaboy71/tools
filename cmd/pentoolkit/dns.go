package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"sort"
	"strings"
)

// runDNS resolves common DNS record types for a domain. It is a convenience
// wrapper around the standard library resolver, useful for the footprinting
// stage of an authorized assessment.
func runDNS(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("dns", flag.ContinueOnError)
	domain := fs.String("domain", "", "domain to resolve, e.g. example.com - required")
	fs.Usage = func() {
		fmt.Fprintf(fs.Output(), "Usage: pentoolkit dns -domain DOMAIN\n\n")
		fmt.Fprintf(fs.Output(), "Resolve A/AAAA/MX/NS/TXT/CNAME records. Example:\n  pentoolkit dns -domain example.com\n\nFlags:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *domain == "" {
		fs.Usage()
		return fmt.Errorf("missing required -domain")
	}

	d := strings.TrimSuffix(strings.TrimSpace(*domain), ".")
	r := net.DefaultResolver

	fmt.Printf("DNS records for %s\n\n", d)

	// A / AAAA
	if ips, err := r.LookupIPAddr(ctx, d); err != nil {
		fmt.Printf("A/AAAA: lookup failed: %v\n", err)
	} else {
		var v4, v6 []string
		for _, ip := range ips {
			if ip.IP.To4() != nil {
				v4 = append(v4, ip.IP.String())
			} else {
				v6 = append(v6, ip.IP.String())
			}
		}
		sort.Strings(v4)
		sort.Strings(v6)
		printList("A", v4)
		printList("AAAA", v6)
	}

	// CNAME
	if cname, err := r.LookupCNAME(ctx, d); err == nil && cname != "" && strings.TrimSuffix(cname, ".") != d {
		printList("CNAME", []string{strings.TrimSuffix(cname, ".")})
	}

	// MX
	if mxs, err := r.LookupMX(ctx, d); err == nil && len(mxs) > 0 {
		var out []string
		for _, mx := range mxs {
			out = append(out, fmt.Sprintf("%d %s", mx.Pref, strings.TrimSuffix(mx.Host, ".")))
		}
		sort.Strings(out)
		printList("MX", out)
	}

	// NS
	if nss, err := r.LookupNS(ctx, d); err == nil && len(nss) > 0 {
		var out []string
		for _, ns := range nss {
			out = append(out, strings.TrimSuffix(ns.Host, "."))
		}
		sort.Strings(out)
		printList("NS", out)
	}

	// TXT
	if txts, err := r.LookupTXT(ctx, d); err == nil && len(txts) > 0 {
		printList("TXT", txts)
	}

	return nil
}

func printList(label string, values []string) {
	if len(values) == 0 {
		return
	}
	fmt.Printf("%-6s:\n", label)
	for _, v := range values {
		fmt.Printf("  %s\n", v)
	}
}
