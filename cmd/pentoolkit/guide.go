package main

import (
	"context"
	"fmt"
)

// runGuide prints a beginner-friendly, step-by-step walkthrough of a typical
// reconnaissance workflow using the toolkit's commands. It is also shown when
// the user runs `pentoolkit I need help`.
func runGuide(ctx context.Context, args []string) error {
	fmt.Print(guideText)
	return nil
}

const guideText = `pentoolkit — step-by-step guide
================================

Before you start: only test systems you OWN or have EXPLICIT WRITTEN
permission to test. Everything below talks to live hosts.

Replace TARGET with your authorized host or domain (e.g. 10.0.0.5 or
example.com) as you go.

Step 1 — Map the domain (DNS footprint)
    See what a domain points at and who handles its mail/name service:

        pentoolkit dns -domain TARGET

Step 2 — Find subdomains
    Discover other hosts that belong to the same domain:

        pentoolkit subenum -domain TARGET

    Have your own list? Point at it with -wordlist:

        pentoolkit subenum -domain TARGET -wordlist my-subdomains.txt

Step 3 — Scan for open ports
    See which services a host exposes. Start with the common ports:

        pentoolkit portscan -host TARGET -ports 1-1024

    Or scan a specific set:

        pentoolkit portscan -host TARGET -ports 22,80,443,8080

Step 4 — Identify each open service
    For every open port from Step 3, grab its banner to learn what it is:

        pentoolkit banner -host TARGET -port 22

Step 5 — Inspect TLS (for any HTTPS/TLS port)
    Check the protocol version and certificate (flags weak TLS and expiry):

        pentoolkit tlsinfo -host TARGET -port 443

Step 6 — Audit web security headers
    For a web server, see which protective headers are missing:

        pentoolkit httpheaders -url https://TARGET

Step 7 — Look for interesting paths (content discovery)
    Probe common paths like /admin, /.git/config, /.env, /robots.txt:

        pentoolkit httpprobe -url https://TARGET

Tips
    * Add -json to ANY command to get machine-readable output:
          pentoolkit dns -domain TARGET -json | jq '.a'
    * Run "pentoolkit <command> -h" to see all flags for that command.
    * Run "pentoolkit help" for the full list of commands.

That's the whole loop: map -> enumerate -> scan -> identify -> inspect.
`
