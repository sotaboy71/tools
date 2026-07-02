# pentoolkit cheat-sheet & glossary

A one-page reference. For the full tutorial see [LEARNING.md](LEARNING.md).

## Command cheat-sheet

| Goal | Command |
|------|---------|
| See all commands | `pentoolkit help` |
| Guided walkthrough | `pentoolkit I need help` |
| Launch the app (web UI) | `pentoolkit serve` → open `http://127.0.0.1:8787` |
| Resolve DNS records | `pentoolkit dns -domain example.com` |
| Find subdomains | `pentoolkit subenum -domain example.com` |
| Find subdomains (own list) | `pentoolkit subenum -domain example.com -wordlist subs.txt` |
| Scan common ports | `pentoolkit portscan -host TARGET -ports 1-1024` |
| Scan specific ports | `pentoolkit portscan -host TARGET -ports 22,80,443` |
| Identify a service | `pentoolkit banner -host TARGET -port 22` |
| Probe a web service | `pentoolkit banner -host TARGET -port 80 -probe "HEAD / HTTP/1.0\r\n\r\n"` |
| Inspect a TLS cert | `pentoolkit tlsinfo -host TARGET -port 443` |
| Audit web headers | `pentoolkit httpheaders -url https://TARGET` |
| Content discovery | `pentoolkit httpprobe -url https://TARGET` |
| Threat reference | `pentoolkit attacks` / `pentoolkit attacks -name phishing` |
| JSON output (any tool) | add `-json`, e.g. `pentoolkit dns -domain x -json \| jq` |
| Flags for a command | `pentoolkit <command> -h` |

Safe practice targets: `127.0.0.1` (yourself), `scanme.nmap.org` (Nmap's
practice host), `example.com` (docs domain), local Docker apps (Juice Shop /
DVWA). Only test what you own or are authorized to test.

## Glossary

- **Host** — a single machine on a network, named (`example.com`) or numbered
  (`93.184.216.34`).
- **IP address** — a machine's numeric address. IPv4 looks like `10.0.0.5`;
  IPv6 like `2606:4700::1`.
- **Port** — a numbered "door" on a host where a specific service listens
  (22=SSH, 80=HTTP, 443=HTTPS). 0–65535.
- **TCP** — the connection-based protocol most services use. A *connect scan*
  finishes the full TCP handshake to tell if a port is open.
- **Open / closed port** — open = something is listening and accepting
  connections; closed = nothing there.
- **Service / daemon** — the program listening on a port (e.g. OpenSSH on 22).
- **Banner** — the first bytes a service sends when you connect; often reveals
  its name and version.
- **DNS** — the phone book of the internet: turns names into IP addresses.
  - **A / AAAA record** — the IPv4 / IPv6 address(es) for a name.
  - **MX record** — which mail servers handle a domain's email.
  - **NS record** — the authoritative name servers for a domain.
  - **TXT record** — free-form text; often holds SPF/verification data.
  - **CNAME** — an alias pointing one name at another.
- **Subdomain** — a name under a domain (`api.example.com` under
  `example.com`). *Enumeration* = discovering which ones exist.
- **TLS / SSL** — the encryption behind HTTPS. TLS 1.2/1.3 are current; older
  versions are weak.
- **Certificate** — the signed document proving a site's identity for TLS.
  - **SAN (Subject Alternative Name)** — the list of hostnames a certificate is
    valid for.
  - **Issuer** — the authority that signed the certificate.
- **HTTP header** — metadata in a web response. *Security headers* (HSTS, CSP,
  X-Frame-Options, …) harden a site; missing ones are weaknesses.
- **HSTS** — a header that forces browsers to use HTTPS.
- **CSP (Content-Security-Policy)** — a header that limits what a page can load,
  reducing XSS risk.
- **Content discovery** — probing for paths that exist but aren't linked
  (`/admin`, `/.git/config`, `/.env`).
- **Recon (reconnaissance)** — the information-gathering phase: map, enumerate,
  scan, identify — before any deeper testing.
- **Payload / probe** — data you send to a service to elicit a response (e.g. an
  HTTP request to make a web server reply).
- **Authorized testing** — testing you have explicit permission to do. The only
  kind that's legal.
- **Blue team / red team** — defenders / authorized simulated attackers.
- **Enumeration** — systematically listing what exists (ports, subdomains,
  paths, users) to build a picture of the target.
