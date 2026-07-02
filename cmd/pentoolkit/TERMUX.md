# Running pentoolkit on Android with Termux

You can build and run `pentoolkit` directly on your phone — no PC required.
Termux gives you a Linux shell on Android, and Go compiles a native ARM binary
right on the device.

> ⚠️ **Authorized use only.** Running this on your phone does not change the
> rules: only scan hosts you own or have written permission to test. Note that
> many mobile carriers and networks prohibit scanning — use your own lab or an
> authorized target.

---

## TL;DR — paste this into Termux

Once Termux is installed (step 1 below), this grabs the source we built and gets
it running in one go:

```sh
pkg update && pkg upgrade -y
pkg install -y golang git
git clone https://github.com/sotaboy71/tools.git
cd tools
git checkout claude/pen-testing-toolkit-lay32c
go build -o pentoolkit ./cmd/pentoolkit
./pentoolkit I need help          # confirms it works + shows the walkthrough
```

Then jump to [step 5](#5a-run-the-web-app-easiest) to open the app. The sections
below explain each line.

---

## 1. Install Termux (from F-Droid, not the Play Store)

The Google Play version of Termux is outdated and broken. Install from
**F-Droid** or the official GitHub releases instead:

- F-Droid: <https://f-droid.org/packages/com.termux/>
- GitHub: <https://github.com/termux/termux-app/releases>

Open Termux after installing.

## 2. Update packages and install Go + git

```sh
pkg update && pkg upgrade -y
pkg install -y golang git
```

Verify Go is installed:

```sh
go version
```

## 3. Get the code (the source we made)

Clone your repository and switch to the branch that has pentoolkit on it:

```sh
git clone https://github.com/sotaboy71/tools.git
cd tools
git checkout claude/pen-testing-toolkit-lay32c
```

`git clone` downloads a full copy of the project into a new `tools` folder;
`cd tools` moves into it; `git checkout` switches to the branch where we added
`cmd/pentoolkit`. (If the code later gets merged into the main branch, you can
skip the `git checkout` line.)

## 4. Build the binary

Because the toolkit uses only the Go standard library, this builds offline once
Go is installed:

```sh
go build -o pentoolkit ./cmd/pentoolkit
```

You now have a single executable named `pentoolkit` in the current folder.

## 5a. Run the web app (easiest)

```sh
./pentoolkit serve
```

Leave that running, then open your phone's browser (Chrome, Firefox, etc.) at:

```
http://127.0.0.1:8787
```

`127.0.0.1`/`localhost` works because the server and the browser are on the same
device. Pick a tool, fill in the target, and tap **Run**. Use the **Help** tab
for the built-in instructions.

> Tip: keep Termux from being killed in the background — some phones aggressively
> stop background apps. Running `termux-wake-lock` first helps. Splitting the
> screen (Termux + browser) also keeps it alive.

## 5b. Or use the command line directly

```sh
./pentoolkit I need help          # step-by-step walkthrough
./pentoolkit dns -domain example.com
./pentoolkit portscan -host example.com -ports 1-1024
./pentoolkit tlsinfo -host example.com -port 443
./pentoolkit attacks              # the 12 attack types: detect & defend
./pentoolkit toolbox              # directory of the big-name tools
```

## Put it to work — a safe first win (all on your phone)

Prove the whole loop works against a target that's 100% yours — the phone
itself. In one Termux session start a tiny local website:

```sh
python3 -m http.server 8000       # (pkg install -y python  if needed)
```

Open a **second** Termux session (swipe from the left edge → New session) and
scan it:

```sh
cd tools
./pentoolkit portscan -host 127.0.0.1 -ports 8000-8005
./pentoolkit banner -host 127.0.0.1 -port 8000 -probe "HEAD / HTTP/1.0\r\n\r\n"
```

You just scanned a machine and identified the service running on it — the core
of the whole toolkit — without touching anyone else's system.

## Optional: install the bigger tools in Termux

`toolbox -check` will tell you which well-known tools you have and give you the
Termux install command for the rest:

```sh
./pentoolkit toolbox -check -os pkg
```

Many (e.g. `nmap`, `hydra`, `nikto`) install straight from Termux:

```sh
pkg install -y nmap
```

Not every desktop tool exists for Android/Termux, but `-check` shows you what's
available and hands you the commands.

## Optional: run pentoolkit from anywhere

Copy the binary onto your `PATH` so you can type `pentoolkit` from any folder:

```sh
mkdir -p $PREFIX/bin
cp pentoolkit $PREFIX/bin/
pentoolkit help
```

## Optional: use your own wordlists

To read a wordlist file from your phone's shared storage:

```sh
termux-setup-storage           # grant storage permission (one time)
./pentoolkit subenum -domain example.com -wordlist ~/storage/downloads/subs.txt
```

---

## Troubleshooting

- **`go: command not found`** — reopen Termux, or run `pkg install -y golang`
  again.
- **Build is slow the first time** — Go compiles the standard library once and
  caches it; later builds are fast.
- **Browser can't reach `127.0.0.1:8787`** — make sure `./pentoolkit serve` is
  still running in Termux (don't close the tab), and that nothing else is using
  that port (change it with `-addr 127.0.0.1:9000`).
- **Want to reach it from another device** — run
  `./pentoolkit serve -addr 0.0.0.0:8787` and browse to your phone's Wi-Fi IP
  from the other device (shown by `ip addr` / `ifconfig`).
