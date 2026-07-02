# pentoolkit

## What is this? (in plain English)

pentoolkit is a set of small tools for **looking at** computers and websites to
understand how they're set up — which "doors" (ports) are open, what software is
running, whether a site's security settings are in place, and so on. Think of it
as a flashlight and a checklist, **not** a crowbar: it looks and reports, it
never breaks in or damages anything.

People use tools like these to check their own systems, to learn how security
works, and to practice on legal training sites like HackTheBox.

> ⚠️ **Only use it on things you're allowed to.** These tools reach out to real
> computers over the network. Only point them at machines you own or have clear,
> written permission to test (or purpose-built practice sites). Scanning
> other people's systems without permission can be illegal.

### What each tool does, in one line

- **dns** — look up a website's address and mail/name-server info.
- **subenum** — find extra sites hidden under a domain (like `mail.` or `dev.`).
- **portscan** — check which "doors" (ports) on a machine are open.
- **banner** — ask an open door what program is behind it (and its version).
- **tlsinfo** — check a site's HTTPS certificate (is it valid? expiring? modern?).
- **httpheaders** — check whether a website has good security settings turned on.
- **httpprobe** — look for common hidden pages on a site (like `/admin`).
- **attacks** — a study guide: the 12 common attack types and how to spot/stop them.
- **guide** / **serve** — a walkthrough, and the point-and-click app version.

## What you need to run it (requirements) — and why

**To build and run pentoolkit itself, you need exactly two things:**

1. **Go** (version 1.11 or newer) — **because** pentoolkit is written in the Go
   programming language, and its source code is just human-readable text until
   something turns it into a program your computer can actually run. Go's build
   command does that. Without Go there's nothing to turn the code into a
   runnable app (unless someone hands you an already-built copy). Download:
   <https://go.dev/dl/>
2. **A terminal** — **because** pentoolkit is a command-line program: you start
   it and read its results by typing commands, so you need the text window that
   lets you do that. Every computer already has one.

That's the whole list. pentoolkit uses **only Go's standard library**, so there
is nothing else to install — no `pip`, no `npm`, no extra packages.

**Install Go** with your system's package manager:

```sh
sudo apt install golang      # Debian / Ubuntu Linux
brew install go              # macOS (Homebrew)
winget install GoLang.Go     # Windows (or download from go.dev/dl)
pkg install golang           # Android (Termux)
```

**Then build it:**

```sh
go build -o pentoolkit ./cmd/pentoolkit
```

**Runs on:** Linux, macOS, Windows, and Android (via Termux) — it's a single
self-contained binary.

### "Do I need Python?" — No.

pentoolkit does **not** need Python at all. Python only ever comes up for one
optional thing: the learning tutorial uses a single Python command
(`python3 -m http.server 8000`) as a quick way to start a tiny test website on
your own computer so you have something safe to practice scanning. Skip that
exercise and you never touch Python. (You'll still meet Python eventually
because *many other* security tools are written in it — but pentoolkit isn't.)

### Optional extras (only for the practice/learning parts — NOT needed to run pentoolkit)

| Tool | Why you'd use it (the "because") | Install |
|------|----------------------------------|---------|
| **git** | **Because** it copies this project's code from the internet onto your machine in one command. Alternative: just download a ZIP. | `apt/brew/pkg install git` |
| **Docker** | **Because** the safe practice sites (Juice Shop, DVWA) are big apps with many parts — Docker runs them pre-packaged in one command instead of you installing every piece. | <https://docs.docker.com/get-docker/> |
| **Python 3** | **Because** the tutorial's `python3 -m http.server` trick instantly creates a small test website to scan. Only needed for that exercise. | Usually preinstalled; else `apt/brew/pkg install python3` |
| **jq** | **Because** pentoolkit's `-json` output is compact; jq reformats it to be readable and lets you pull out specific fields. | `apt/brew/pkg install jq` |

## Quick start (easiest path)

```sh
# 1. build it once (needs Go installed)
go build -o pentoolkit ./cmd/pentoolkit

# 2. open the point-and-click app
./pentoolkit serve
#    then open http://127.0.0.1:8787 in your browser

# or, if you like the command line, start with the guided walkthrough:
./pentoolkit I need help
```

New to the terms (port, DNS, TLS…)? See the plain-English glossary in
[CHEATSHEET.md](CHEATSHEET.md), and the step-by-step tutorial in
[LEARNING.md](LEARNING.md).

## Install

```
go build -o pentoolkit ./cmd/pentoolkit
# or
go install golang.org/x/tools/cmd/pentoolkit@latest
```

All commands are **read-only and non-destructive** — they look and report, they
do not attempt to break into or exploit anything. Written in pure Go (standard
library only), so there's nothing extra to install.

## New here?

Get a step-by-step walkthrough of a full recon workflow:

```
pentoolkit I need help
# or
pentoolkit guide
```

**Want to actually learn the tools?** [LEARNING.md](LEARNING.md) is a hands-on
lab guide: how to practice every tool on safe, legal targets (your own machine,
local vulnerable apps, Nmap's practice host) and how to move on to
HackTheBox/TryHackMe. For a one-page command reference and a glossary of terms,
see [CHEATSHEET.md](CHEATSHEET.md).

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

**On Android?** You can build and run this whole thing on your phone with
Termux — no PC needed. See [TERMUX.md](TERMUX.md). For a real launcher icon /
full-screen app, there's also a WebView wrapper in
[android/](android/) that displays this same UI.

## Commands

| Command       | In plain English                                               | Technical name |
|---------------|----------------------------------------------------------------|----------------|
| `dns`         | Look up a domain's addresses and mail/name servers             | DNS record lookup |
| `subenum`     | Find extra sites under a domain (`mail.`, `dev.`, …)           | subdomain enumeration |
| `portscan`    | See which ports (doors) are open on a machine                  | TCP connect scan |
| `banner`      | Ask an open port what program/version is behind it             | banner grab |
| `tlsinfo`     | Check a site's HTTPS certificate and encryption                | TLS/cert inspection |
| `httpheaders` | Check whether a website's security settings are turned on      | HTTP security-header audit |
| `httpprobe`   | Look for common hidden pages (`/admin`, `/.env`, …)            | content discovery |
| `attacks`     | Study guide: 12 common attacks and how to spot/stop them       | defender reference |
| `toolbox`     | Directory of the big-name tools (Nmap, Metasploit, …) + links  | tool reference |
| `guide`       | A step-by-step walkthrough of how to use everything            | workflow help |
| `serve`       | The point-and-click app version (opens in your browser)        | web UI |

Run `pentoolkit <command> -h` to see the options for any command. Add `-json` to
any command to get the results as data (handy for feeding into other tools).

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

# Learn to detect and defend the 12 most common attack types
pentoolkit attacks                 # list them
pentoolkit attacks -name phishing  # detail: what it is, detect, defend
```

## The bigger toolkit (`toolbox`)

pentoolkit is a small, safe **learner's** toolkit. The professionals use a wider
set of programs — and `pentoolkit toolbox` is a built-in directory of them:

```sh
pentoolkit toolbox                 # list everything, grouped by category
pentoolkit toolbox -category web   # just one category
pentoolkit toolbox -json           # as data
```

It covers the big names you've probably heard of — **Nmap** (scanning),
**Metasploit** (exploitation), **Burp Suite** / **OWASP ZAP** (web),
**aircrack-ng** (Wi-Fi), **Wireshark** (traffic), **hashcat** / **John the
Ripper** / **Hydra** (passwords), and more — plus the all-in-one **Kali Linux**
distro that bundles most of them, and legal practice sites (HackTheBox,
TryHackMe, VulnHub).

**Important:** pentoolkit does **not** ship these — they're separate projects.
The directory tells you what each is, links to its official source, and (where
relevant) which pentoolkit command is a lightweight version of it. The easiest
way to get them all at once is to install Kali Linux. As always, use every one
of them only on systems you own or are authorized to test.

## Threat reference (defender view)

`pentoolkit attacks` is a built-in, read-only reference covering the 12 most
common attack types (phishing, malware, ransomware, DoS/DDoS, MitM, credential
attacks, social engineering, SQLi, zero-day, insider, supply chain, spoofing).
For each one it explains what it is, **how to detect it**, and **how to defend
against it**. It is educational/blue-team material — the toolkit does not
perform any of these attacks. The same content appears in the web UI's
**Threats** tab.

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
