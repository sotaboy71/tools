# Learning pentoolkit — a hands-on lab guide

This guide teaches you how to use every tool in pentoolkit by practicing on
targets that are **safe and legal to test**. Work through it top to bottom; each
tool builds on the last.

## First: where you're allowed to practice

Scanning random hosts on the internet can be illegal even if you mean no harm.
Practice only on these:

| Target | What it's for | Notes |
|--------|---------------|-------|
| **Your own machine** (`127.0.0.1` / `localhost`) | Everything | Always allowed |
| **scanme.nmap.org** | portscan, banner, tlsinfo | The Nmap project runs this host *specifically* for people to practice scanning. Don't hammer it — a scan or two. |
| **example.com** | dns, tlsinfo, httpheaders | Reserved documentation domain; fine for read-only lookups |
| **Local vulnerable apps** (DVWA, OWASP Juice Shop) | httpheaders, httpprobe | You run them yourself in Docker (below) |
| **HackTheBox / TryHackMe** | Everything, realistically | Legal, sandboxed practice ranges — see the last section |

Rule of thumb: if you don't own it and it isn't on a list like this, don't scan
it.

## Build the toolkit

```sh
go build -o pentoolkit ./cmd/pentoolkit
./pentoolkit help          # see all commands
./pentoolkit I need help   # the quick workflow walkthrough
```

You can also do all of this from the web app — run `./pentoolkit serve` and open
<http://127.0.0.1:8787>. The CLI and the app do exactly the same thing.

---

## Tool 1 — `dns`: what a domain points at

Start here because it's read-only and harmless.

```sh
./pentoolkit dns -domain example.com
```

**What you're seeing:**
- `A` / `AAAA` — the IPv4 / IPv6 addresses the name resolves to.
- `MX` — mail servers for the domain.
- `NS` — the authoritative name servers.
- `TXT` — free-form records; look for `v=spf1...` (email anti-spoofing).

**Try:** run it on a few domains you use. Notice how big sites return many `A`
records (load balancing) and lots of `TXT` records.

## Tool 2 — `subenum`: find subdomains

```sh
./pentoolkit subenum -domain example.com
```

It tries a built-in list of common names (`www`, `mail`, `api`, `dev`, …) and
reports the ones that resolve. Real hosts often expose `dev`/`staging`/`vpn`
subdomains you wouldn't guess from the main site.

**Try:** make your own list and pass it:
```sh
printf "www\nmail\napi\nblog\nshop\n" > subs.txt
./pentoolkit subenum -domain example.com -wordlist subs.txt
```

## Tool 3 — `portscan`: which ports are open

Practice on your own machine and on Nmap's practice host.

```sh
./pentoolkit portscan -host 127.0.0.1 -ports 1-1024
./pentoolkit portscan -host scanme.nmap.org -ports 20-100,443
```

**What you're seeing:** an open port means a service is listening there. The
tool labels well-known ports (22=ssh, 80=http, 443=https, …).

**Understand the trade-offs:**
- `-ports` accepts lists (`22,80,443`) and ranges (`1-1024`).
- `-workers` controls concurrency (higher = faster, noisier).
- This is a *connect* scan — it completes the full TCP handshake, so the target
  can easily log it. That's fine for authorized practice.

**Try:** start a local server (`python3 -m http.server 8000`) in another
terminal, then scan `127.0.0.1 -ports 8000-8010` and watch it show up.

## Tool 4 — `banner`: identify a service

Once you know a port is open, see what's actually running there.

```sh
# against the python server you started above:
./pentoolkit banner -host 127.0.0.1 -port 8000 -probe "HEAD / HTTP/1.0\r\n\r\n"
# against a real SSH server (e.g. scanme):
./pentoolkit banner -host scanme.nmap.org -port 22
```

Many services announce their name and version in the first bytes they send
(SSH does automatically; HTTP needs the `-probe` above to prompt a reply).
Knowing exact versions is how you'd later check for known vulnerabilities.

## Tool 5 — `tlsinfo`: inspect a certificate

```sh
./pentoolkit tlsinfo -host example.com -port 443
```

**What to look for:**
- `Version` — you want TLS 1.2+; the tool flags anything older as weak.
- `Valid ... -> ...` and `Expires in N days` — expired/soon-to-expire certs.
- `SANs` — every hostname the certificate is valid for (often reveals other
  domains an org owns).

**Try:** point it at `scanme.nmap.org:443` and at your own sites. Compare a
modern site (TLS 1.3) with anything older you can find.

## Tool 6 — `httpheaders`: audit a site's security headers

Best practiced against a vulnerable app you run locally (next section), but it
works on any site you're allowed to fetch.

```sh
./pentoolkit httpheaders -url http://localhost:3000
```

Each `[-]` is a missing protective header (HSTS, CSP, X-Frame-Options, …) and
each `[!]` is a header that leaks software versions. This is how you'd assess a
web app's baseline hardening.

## Tool 7 — `httpprobe`: content discovery

```sh
./pentoolkit httpprobe -url http://localhost:3000
```

It requests a list of common/interesting paths (`/admin`, `/robots.txt`,
`/.git/config`, `/.env`, …) and shows anything that isn't a 404. This is how you
find login panels, exposed config, or leftover files.

**Try:** `-all` to see every result, and `-wordlist paths.txt` for your own list.

---

## Set up your own vulnerable web app (safe, local)

The best way to practice the web tools is against apps that are *designed* to be
tested. Run one locally with Docker:

```sh
# OWASP Juice Shop (modern, lots to explore)
docker run --rm -p 3000:3000 bkimminich/juice-shop
# then, in another terminal:
./pentoolkit httpheaders -url http://localhost:3000
./pentoolkit httpprobe   -url http://localhost:3000
```

```sh
# DVWA (classic, deliberately vulnerable)
docker run --rm -p 8080:80 vulnerables/web-dvwa
./pentoolkit httpprobe -url http://localhost:8080
```

These run entirely on your machine, so you can poke at them all day. They're
also where you'd practice the *manual* techniques (SQLi, XSS, etc.) that go
beyond what pentoolkit does — each app ships with its own lessons.

---

## Suggested first session (30–45 min)

1. `./pentoolkit I need help` — read the workflow.
2. `dns` and `subenum` on `example.com`.
3. Start `python3 -m http.server 8000`; `portscan` and `banner` it on
   `127.0.0.1`.
4. `portscan`, `banner`, and `tlsinfo` against `scanme.nmap.org`.
5. `docker run ... juice-shop`; run `httpheaders` and `httpprobe` against it.
6. Re-run anything with `-json` and pipe to `jq` to see structured output.

## When you're ready: HackTheBox / TryHackMe

These are legal, sandboxed platforms built for exactly this — you connect to an
isolated network and attack machines you're *authorized* to attack.

- **TryHackMe** (<https://tryhackme.com>) — the gentler on-ramp; guided rooms
  like "Nmap", "Network Services", and "Web Fundamentals" map directly onto the
  pentoolkit tools you just learned.
- **HackTheBox** (<https://hackthebox.com>) — start with *Starting Point*, then
  easy retired machines. Their Academy has structured modules.

Your pentoolkit workflow is the same on those ranges: enumerate DNS/subdomains,
scan ports, grab banners, check TLS, then probe web services — the difference is
you're allowed to go further and actually exploit the (intentionally vulnerable)
targets.

## The one rule

Only test systems you own or are explicitly authorized to test (your own lab,
the practice hosts above, or a range that grants permission). That's not just
etiquette — unauthorized scanning/access is illegal in most places. Keep it in
the sandbox and you can learn everything safely.
