// Command sotaserver (Sota's Pen Testing) is a small, self-contained toolkit of reconnaissance
// utilities for authorized penetration testing and security education.
//
// It bundles several read-only, non-destructive commands that are commonly
// used during the information-gathering phase of an authorized engagement:
//
//	portscan    TCP connect scan of a host over a range/list of ports
//	banner      Grab the service banner exposed on a single TCP port
//	httpheaders Fetch a URL and report on security-relevant HTTP headers
//	dns         Resolve A/AAAA/MX/NS/TXT/CNAME records for a domain
//	tlsinfo     Inspect the TLS certificate chain presented by a host
//	subenum     Discover live subdomains of a domain via DNS
//	httpprobe   Probe common paths on a URL for content discovery
//
// IMPORTANT: Only run these tools against systems you own or have explicit,
// written permission to test. Unauthorized scanning may be illegal.
package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"
)

// command describes a single subcommand of the toolkit.
type command struct {
	name    string
	summary string
	run     func(ctx context.Context, args []string) error
}

var commands = []command{
	{"portscan", "TCP connect scan of a host across a set of ports", runPortscan},
	{"banner", "Grab the service banner from a single TCP port", runBanner},
	{"httpheaders", "Report on security-relevant HTTP response headers", runHTTPHeaders},
	{"dns", "Resolve DNS records (A, AAAA, MX, NS, TXT, CNAME)", runDNS},
	{"tlsinfo", "Inspect the TLS certificate presented by a host", runTLSInfo},
	{"subenum", "Discover live subdomains of a domain via DNS", runSubenum},
	{"httpprobe", "Probe common paths on a URL for content discovery", runHTTPProbe},
	{"guide", "Print a step-by-step walkthrough of a recon workflow", runGuide},
	{"attacks", "Defender reference: detect & defend the 12 common attacks", runAttacks},
	{"toolbox", "Directory of well-known tools (Nmap, Metasploit, ...) & where to get them", runToolbox},
	{"serve", "Launch the mobile-friendly web UI for all tools", runServe},
}

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}

	name := os.Args[1]
	switch name {
	case "-h", "--help", "help":
		usage()
		return
	}

	// Friendly natural-language entry point: `sotaserver I need help`
	// (in any capitalization) prints the step-by-step guide.
	if strings.EqualFold(strings.Join(os.Args[1:], " "), "i need help") {
		runGuide(context.Background(), nil)
		return
	}

	for _, c := range commands {
		if c.name == name {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			// Give every command a generous default deadline so a hung
			// target cannot wedge the process indefinitely.
			ctx, cancel2 := context.WithTimeout(ctx, 10*time.Minute)
			defer cancel2()
			if err := c.run(ctx, os.Args[2:]); err != nil {
				fmt.Fprintf(os.Stderr, "sotaserver %s: %v\n", name, err)
				os.Exit(1)
			}
			return
		}
	}

	fmt.Fprintf(os.Stderr, "sotaserver: unknown command %q\n\n", name)
	usage()
	os.Exit(2)
}

func usage() {
	fmt.Fprintf(os.Stderr, `Sota's Pen Testing is a recon toolkit for authorized penetration testing.

Usage:
	sotaserver <command> [flags]

Commands:
`)
	for _, c := range commands {
		fmt.Fprintf(os.Stderr, "\t%-12s %s\n", c.name, c.summary)
	}
	fmt.Fprintf(os.Stderr, `
Run "sotaserver <command> -h" for details on a command.
New here? Run "sotaserver I need help" for a step-by-step walkthrough.

Only use these tools against systems you are authorized to test.
`)
}
