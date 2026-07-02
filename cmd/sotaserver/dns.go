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
	jsonOut := fs.Bool("json", false, "emit results as JSON")
	fs.Usage = func() {
		fmt.Fprintf(fs.Output(), "Usage: sotaserver dns -domain DOMAIN\n\n")
		fmt.Fprintf(fs.Output(), "Resolve A/AAAA/MX/NS/TXT/CNAME records. Example:\n  sotaserver dns -domain example.com\n\nFlags:\n")
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

	var rec dnsRecords
	rec.Domain = d

	// A / AAAA
	if ips, err := r.LookupIPAddr(ctx, d); err == nil {
		for _, ip := range ips {
			if ip.IP.To4() != nil {
				rec.A = append(rec.A, ip.IP.String())
			} else {
				rec.AAAA = append(rec.AAAA, ip.IP.String())
			}
		}
		sort.Strings(rec.A)
		sort.Strings(rec.AAAA)
	}

	// CNAME
	if cname, err := r.LookupCNAME(ctx, d); err == nil {
		if c := strings.TrimSuffix(cname, "."); c != "" && c != d {
			rec.CNAME = c
		}
	}

	// MX
	if mxs, err := r.LookupMX(ctx, d); err == nil {
		for _, mx := range mxs {
			rec.MX = append(rec.MX, fmt.Sprintf("%d %s", mx.Pref, strings.TrimSuffix(mx.Host, ".")))
		}
		sort.Strings(rec.MX)
	}

	// NS
	if nss, err := r.LookupNS(ctx, d); err == nil {
		for _, ns := range nss {
			rec.NS = append(rec.NS, strings.TrimSuffix(ns.Host, "."))
		}
		sort.Strings(rec.NS)
	}

	// TXT
	if txts, err := r.LookupTXT(ctx, d); err == nil {
		rec.TXT = txts
	}

	return emit(*jsonOut, rec, func() {
		fmt.Printf("DNS records for %s\n\n", d)
		printList("A", rec.A)
		printList("AAAA", rec.AAAA)
		if rec.CNAME != "" {
			printList("CNAME", []string{rec.CNAME})
		}
		printList("MX", rec.MX)
		printList("NS", rec.NS)
		printList("TXT", rec.TXT)
	})
}

// dnsRecords is the JSON shape of the dns command.
type dnsRecords struct {
	Domain string   `json:"domain"`
	A      []string `json:"a,omitempty"`
	AAAA   []string `json:"aaaa,omitempty"`
	CNAME  string   `json:"cname,omitempty"`
	MX     []string `json:"mx,omitempty"`
	NS     []string `json:"ns,omitempty"`
	TXT    []string `json:"txt,omitempty"`
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
