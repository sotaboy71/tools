# pentoolkit

A small, self-contained toolkit of **reconnaissance** utilities for authorized
penetration testing and security education. It is written in pure Go (standard
library only) and adds no dependencies to the module.

> ⚠️ **Authorized use only.** Every command here talks to remote systems. Only
> run them against hosts you own or have **explicit, written permission** to
> test. Unauthorized scanning may be illegal in your jurisdiction.

All commands are **read-only and non-destructive** — they fingerprint and
report, they do not attempt to exploit anything.

## Install

```
go build -o pentoolkit ./cmd/pentoolkit
# or
go install golang.org/x/tools/cmd/pentoolkit@latest
```

## New here?

Get a step-by-step walkthrough of a full recon workflow:

```
pentoolkit I need help
# or
pentoolkit guide
```

## App / web UI (works on your phone)

Prefer buttons over a terminal? Launch the built-in web app:

```
pentoolkit serve
```

Then open <http://127.0.0.1:8787> in a browser. It's a mobile-friendly page
with a tool picker, a form for each tool, a JSON-output toggle, and a Help tab
containing these instructions.

To use it from your **phone** on the same Wi-Fi, bind to all interfaces and
browse to your computer's LAN IP:

```
pentoolkit serve -addr 0.0.0.0:8787
# then on your phone: http://<your-computer-ip>:8787
```

The server binds to `127.0.0.1` (this machine only) by default, and never runs
a shell — each tool is executed as a separate, argument-safe subprocess.

## Commands

| Command       | Purpose                                                        |
|---------------|----------------------------------------------------------------|
| `portscan`    | Concurrent TCP connect scan across a range/list of ports       |
| `banner`      | Grab the service banner exposed on a single TCP port           |
| `httpheaders` | Fetch a URL and report on security-relevant HTTP headers       |
| `dns`         | Resolve A/AAAA/MX/NS/TXT/CNAME records for a domain             |
| `tlsinfo`     | Inspect the TLS version, cipher, and certificate chain of a host |
| `subenum`     | Discover live subdomains of a domain via DNS resolution        |
| `httpprobe`   | Probe a URL for common/interesting paths (content discovery)   |
| `guide`       | Print a step-by-step walkthrough of a recon workflow           |
| `serve`       | Launch the mobile-friendly web UI for all tools                |

Run `pentoolkit <command> -h` for the flags of any command. Every command
accepts `-json` to emit machine-readable output for piping into other tools
(e.g. `jq`).

## Examples

```sh
# Scan the well-known ports of an authorized target
pentoolkit portscan -host 10.0.0.5 -ports 1-1024

# Scan a specific set of ports
pentoolkit portscan -host 10.0.0.5 -ports 22,80,443,8080

# Identify a service by its banner
pentoolkit banner -host 10.0.0.5 -port 22

# Send an HTTP probe and read the response
pentoolkit banner -host 10.0.0.5 -port 80 -probe "HEAD / HTTP/1.0\r\n\r\n"

# Audit a site's security headers
pentoolkit httpheaders -url https://example.com

# Footprint a domain's DNS records
pentoolkit dns -domain example.com

# Inspect a TLS certificate (works on expired/self-signed certs too)
pentoolkit tlsinfo -host example.com -port 443

# Enumerate subdomains (built-in wordlist, or supply your own)
pentoolkit subenum -domain example.com
pentoolkit subenum -domain example.com -wordlist subdomains.txt

# Content discovery: probe common paths and show non-404 responses
pentoolkit httpprobe -url https://example.com
pentoolkit httpprobe -url https://example.com -wordlist paths.txt -all

# Any command can emit JSON for scripting
pentoolkit dns -domain example.com -json | jq '.a'
```

## Notes

- `portscan` performs a full TCP handshake ("connect scan"), which is easy for
  the target to log. It is intentionally not stealthy.
- `httpheaders` and `httpprobe` honor standard `HTTP(S)_PROXY` / `NO_PROXY`
  environment variables.
- `tlsinfo` deliberately skips certificate verification so that misconfigured,
  expired, or self-signed certificates can still be inspected.
- `subenum` performs active DNS enumeration (it sends real DNS queries for each
  candidate label). It is non-destructive but not silent.
- `httpprobe` sends plain `GET`/`HEAD` requests and hides `404`s by default;
  pass `-all` to see every result.
