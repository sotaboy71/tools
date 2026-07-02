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

## Commands

| Command       | Purpose                                                        |
|---------------|----------------------------------------------------------------|
| `portscan`    | Concurrent TCP connect scan across a range/list of ports       |
| `banner`      | Grab the service banner exposed on a single TCP port           |
| `httpheaders` | Fetch a URL and report on security-relevant HTTP headers       |
| `dns`         | Resolve A/AAAA/MX/NS/TXT/CNAME records for a domain             |
| `tlsinfo`     | Inspect the TLS version, cipher, and certificate chain of a host |

Run `pentoolkit <command> -h` for the flags of any command.

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
```

## Notes

- `portscan` performs a full TCP handshake ("connect scan"), which is easy for
  the target to log. It is intentionally not stealthy.
- `httpheaders` honors standard `HTTP(S)_PROXY` / `NO_PROXY` environment
  variables.
- `tlsinfo` deliberately skips certificate verification so that misconfigured,
  expired, or self-signed certificates can still be inspected.
