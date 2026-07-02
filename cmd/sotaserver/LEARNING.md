# Learning Sota's Pen Testing — a hands-on lab guide

This guide teaches you how to use every tool in sotaserver by trying it on
targets that are **safe and legal to practice on**. No experience needed — go
top to bottom and each tool builds on the last. Copy a command, run it, and read
the "What you're seeing" note to understand the result.

**The fastest way to see it work:** open a terminal, run `python3 -m http.server
8000` (that starts a tiny website on your own computer), then in another
terminal run `./sotaserver portscan -host 127.0.0.1 -ports 8000-8005`. You just
scanned a machine (your own) and found the open door. That's the whole idea —
the rest is variations on it.

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
go build -o sotaserver ./cmd/sotaserver
./sotaserver help          # see all commands
./sotaserver I need help   # the quick workflow walkthrough
```

You can also do all of this from the web app — run `./sotaserver serve` and open
<http://127.0.0.1:8787>. The CLI and the app do exactly the same thing.

---

## Tool 1 — `dns`: what a domain points at

DNS is the internet's phone book: it turns a name like `example.com` into the
numeric address of the computer behind it. This tool just looks names up, so
it's completely harmless — a great place to start.

```sh
./sotaserver dns -domain example.com
```

**What you're seeing:**
- `A` / `AAAA` — the actual address(es) of the site (the newer `AAAA` kind is
  just the longer, modern format).
- `MX` — which servers handle that domain's email.
- `NS` — the servers that are "in charge of" answering questions about the
  domain.
- `TXT` — notes attached to the domain. A `v=spf1...` line is an anti-spam
  setting that says which servers are allowed to send the domain's email.

**Try:** run it on a few sites you use. Big sites often return several addresses
(so traffic can be spread across many computers) and lots of `TXT` notes.

## Tool 2 — `subenum`: find subdomains

```sh
./sotaserver subenum -domain example.com
```

A subdomain is a site that lives "under" a main domain — like
`mail.example.com` or `dev.example.com` under `example.com`. This tool tries a
built-in list of common names (`www`, `mail`, `api`, `dev`, …) and tells you
which ones actually exist. Organizations often have `dev`, `staging`, or `vpn`
sites you'd never find just by visiting the main page.

**Try:** make your own list and pass it:
```sh
printf "www\nmail\napi\nblog\nshop\n" > subs.txt
./sotaserver subenum -domain example.com -wordlist subs.txt
```

## Tool 3 — `portscan`: which ports are open

Practice on your own machine and on Nmap's practice host.

```sh
./sotaserver portscan -host 127.0.0.1 -ports 1-1024
./sotaserver portscan -host scanme.nmap.org -ports 20-100,443
```

Think of a computer as a building with thousands of numbered doors (ports).
Each program that accepts connections sits behind one door. This tool knocks on
the doors you ask about and tells you which ones open.

**What you're seeing:** an "open" port means some program is listening there and
willing to talk. The tool labels the famous doors for you (22 = remote login /
SSH, 80 = web / HTTP, 443 = secure web / HTTPS, …).

**Good to know:**
- `-ports` takes a list (`22,80,443`) or a range (`1-1024`).
- `-workers` is how many doors it knocks on at once — higher is faster but more
  noticeable.
- This kind of scan knocks politely and fully, so the target can easily see it
  in its logs. That's completely fine when you're testing your own systems or a
  practice host.

**Try:** start a tiny local website with `python3 -m http.server 8000` in
another terminal, then scan `127.0.0.1 -ports 8000-8010` and watch door 8000
show up as open.

## Tool 4 — `banner`: identify a service

Once you know a port is open, see what's actually running there.

```sh
# against the python server you started above:
./sotaserver banner -host 127.0.0.1 -port 8000 -probe "HEAD / HTTP/1.0\r\n\r\n"
# against a real SSH server (e.g. scanme):
./sotaserver banner -host scanme.nmap.org -port 22
```

When you connect to a program, it often introduces itself — sometimes with its
exact name and version. That first hello is called a "banner." Some programs
say hello on their own (like SSH); a web server needs you to ask a question
first, which is what the `-probe` bit does. Knowing the exact version matters
because that's how you (or an attacker) would look up whether it has known,
already-published security holes.

## Tool 5 — `tlsinfo`: inspect a certificate

```sh
./sotaserver tlsinfo -host example.com -port 443
```

TLS is the lock behind the padlock icon in your browser — the encryption that
makes a site "HTTPS." Every secure site presents a certificate: a signed ID card
proving it's really that site. This tool shows you that ID card.

**What to look for:**
- `Version` — the age of the encryption. You want TLS 1.2 or 1.3; the tool warns
  you if it's older (weaker).
- `Valid ... -> ...` and `Expires in N days` — when the ID card is good until. An
  expired one causes browser warnings.
- `SANs` — the full list of site names this one certificate covers. It often
  reveals other domains the same organization owns.

**Try:** point it at `scanme.nmap.org:443` and at your own sites, and compare a
modern site (TLS 1.3) with anything older you can find.

## Tool 6 — `httpheaders`: audit a site's security headers

Best practiced against a vulnerable app you run locally (next section), but it
works on any site you're allowed to fetch.

```sh
./sotaserver httpheaders -url http://localhost:3000
```

Every time a website responds, it sends along some hidden settings called
"headers." A few of them are safety features (force HTTPS, block certain
attacks, etc.). This tool checks whether those safety features are switched on.

- A `[-]` means a recommended safety setting is **missing**.
- A `[!]` means the site is **giving away** which software and version it runs
  (helpful info for an attacker).

This is a quick way to gauge how well a website has been locked down.

## Tool 7 — `httpprobe`: content discovery

```sh
./sotaserver httpprobe -url http://localhost:3000
```

Websites often have pages that aren't linked anywhere but still exist if you
know the address — an admin login, a settings file, a forgotten backup. This
tool tries a list of common ones (`/admin`, `/robots.txt`, `/.git/config`,
`/.env`, …) and shows you which ones are actually there. (A "404" just means
"page not found," so the tool hides those by default and shows you the hits.)

**Try:** add `-all` to see every result, or `-wordlist paths.txt` to use your
own list of pages to check.

---

## Set up your own vulnerable web app (safe, local)

The best way to practice the web tools is against apps that are *designed* to be
tested. Run one locally with Docker:

```sh
# OWASP Juice Shop (modern, lots to explore)
docker run --rm -p 3000:3000 bkimminich/juice-shop
# then, in another terminal:
./sotaserver httpheaders -url http://localhost:3000
./sotaserver httpprobe   -url http://localhost:3000
```

```sh
# DVWA (classic, deliberately vulnerable)
docker run --rm -p 8080:80 vulnerables/web-dvwa
./sotaserver httpprobe -url http://localhost:8080
```

These run entirely on your machine, so you can poke at them all day. They're
also where you'd practice the *manual* techniques (SQLi, XSS, etc.) that go
beyond what sotaserver does — each app ships with its own lessons.

---

## Suggested first session (30–45 min)

1. `./sotaserver I need help` — read the workflow.
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
  sotaserver tools you just learned.
- **HackTheBox** (<https://hackthebox.com>) — start with *Starting Point*, then
  easy retired machines. Their Academy has structured modules.

Your sotaserver workflow is the same on those ranges: enumerate DNS/subdomains,
scan ports, grab banners, check TLS, then probe web services — the difference is
you're allowed to go further and actually exploit the (intentionally vulnerable)
targets.

## The one rule

Only test systems you own or are explicitly authorized to test (your own lab,
the practice hosts above, or a range that grants permission). That's not just
etiquette — unauthorized scanning/access is illegal in most places. Keep it in
the sandbox and you can learn everything safely.
