# Sota's Pen Testing cheat-sheet & glossary

A one-page reference. For the full tutorial see [LEARNING.md](LEARNING.md).

## Command cheat-sheet

| Goal | Command |
|------|---------|
| See all commands | `sotaserver help` |
| Guided walkthrough | `sotaserver I need help` |
| Launch the app (web UI) | `sotaserver serve` → open `http://127.0.0.1:8787` |
| Resolve DNS records | `sotaserver dns -domain example.com` |
| Find subdomains | `sotaserver subenum -domain example.com` |
| Find subdomains (own list) | `sotaserver subenum -domain example.com -wordlist subs.txt` |
| Scan common ports | `sotaserver portscan -host TARGET -ports 1-1024` |
| Scan specific ports | `sotaserver portscan -host TARGET -ports 22,80,443` |
| Identify a service | `sotaserver banner -host TARGET -port 22` |
| Probe a web service | `sotaserver banner -host TARGET -port 80 -probe "HEAD / HTTP/1.0\r\n\r\n"` |
| Inspect a TLS cert | `sotaserver tlsinfo -host TARGET -port 443` |
| Audit web headers | `sotaserver httpheaders -url https://TARGET` |
| Content discovery | `sotaserver httpprobe -url https://TARGET` |
| Threat reference | `sotaserver attacks` / `sotaserver attacks -name phishing` |
| JSON output (any tool) | add `-json`, e.g. `sotaserver dns -domain x -json \| jq` |
| Flags for a command | `sotaserver <command> -h` |

Safe practice targets: `127.0.0.1` (yourself), `scanme.nmap.org` (Nmap's
practice host), `example.com` (docs domain), local Docker apps (Juice Shop /
DVWA). Only test what you own or are authorized to test.

## Setup toolbox — the programs you use to get things running

These are the general-purpose tools you'll type commands into or install other
things with. You don't need all of them — see "What sotaserver needs" at the
end.

- **Terminal** — the text window where you type commands (Terminal on macOS,
  Command Prompt/PowerShell on Windows, any terminal app on Linux, Termux on
  Android).
- **Shell** (`sh`, `bash`, `zsh`) — the program *inside* the terminal that
  actually runs what you type. When a guide says "run this in a shell / `sh`,"
  it just means type it in the terminal. A `.sh` file is a saved list of shell
  commands.
- **`./sotaserver`** — the leading `./` means "run the program in this folder."
  On Windows it's just `sotaserver.exe`.
- **Go** (`go build`) — the programming language sotaserver is written in.
  `go build -o sotaserver ./cmd/sotaserver` turns the source code into the
  runnable `sotaserver` program. **This is the only thing sotaserver itself
  needs.** Get Go from <https://go.dev/dl/>.
- **Package manager** — the "app store" you install software with from the
  terminal. Which one depends on your system:
  - Debian/Ubuntu Linux: `sudo apt install golang git`
  - macOS (Homebrew): `brew install go git`
  - Android (Termux): `pkg install golang git`
  - Windows: `winget install GoLang.Go` or download from the site
- **Python** & **`pip`** — Python is a different programming language; `pip` is
  Python's package installer (`pip install <name>`). sotaserver does **not** use
  Python, but tons of *other* security tools do, so you'll meet `pip` a lot. The
  tutorial uses one handy Python trick — `python3 -m http.server 8000` — to
  start a tiny test website on your own machine to practice against.
- **Docker** (`docker run`) — runs a whole pre-packaged app in an isolated box
  without you installing all its pieces. Used here to launch practice targets,
  e.g. `docker run --rm -p 3000:3000 bkimminich/juice-shop`.
- **git** (`git clone`) — downloads a copy of a code project (a "repository") to
  your machine, e.g. `git clone https://github.com/OWNER/REPO.git`.
- **jq** — a small tool that pretty-prints and filters JSON; pair it with
  `-json`, e.g. `sotaserver dns -domain x -json | jq`.
- **curl** — fetches a URL straight from the terminal; handy for quick checks.
- **PATH** — the list of folders your shell searches for programs. Copy
  `sotaserver` into one (e.g. `/usr/local/bin` or Termux's `$PREFIX/bin`) to run
  it from anywhere by name instead of `./sotaserver`.

**What sotaserver needs:** just **Go** to build it (or a prebuilt binary someone
hands you). Everything else above is optional and only for the wider learning:
**Docker** for practice apps, **Python** for the quick test-server trick, **jq**
for prettier output, **git** to download source code.

## Glossary

- **Host** — a single machine on a network, named (`example.com`) or numbered
  (`93.184.216.34`).
- **IP address** — a machine's numeric address. IPv4 looks like `10.0.0.5`;
  IPv6 like `2606:4700::1`.
- **Port** — a numbered "door" on a host where a specific service listens
  (22=SSH, 80=HTTP, 443=HTTPS). 0–65535.
- **TCP** — the most common way two machines open a reliable connection. The
  "handshake" is the quick back-and-forth that starts it; a *connect scan*
  completes that handshake to prove a port is open.
- **Open / closed port** — open = a program is there and accepting connections;
  closed = nothing is listening.
- **Service / daemon** — the program sitting behind a port (e.g. OpenSSH on 22).
  "Daemon" (say "demon") is just another word for a background service.
- **Banner** — the first bytes a service sends when you connect; often reveals
  its name and version.
- **DNS** — the phone book of the internet: turns names into IP addresses.
  - **A / AAAA record** — the IPv4 / IPv6 address(es) for a name.
  - **MX record** — which mail servers handle a domain's email.
  - **NS record** — the servers officially in charge of answering DNS questions
    for a domain.
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
- **CSP (Content-Security-Policy)** — a header that limits what a page is allowed
  to load, which helps block XSS (cross-site scripting, a common web attack that
  injects malicious scripts into a page).
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
- **Wordlist** — a plain text file with one guess per line (subdomain names,
  page paths, …) that a tool tries one by one. Pass one with `-wordlist`.
- **localhost / `127.0.0.1`** — a name/number that always means "this same
  computer." Perfect (and always allowed) for practice.
- **404** — a web server's "page not found" reply. `httpprobe` hides these by
  default so you only see pages that actually exist.
