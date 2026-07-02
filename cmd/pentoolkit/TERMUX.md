# Running pentoolkit on Android with Termux

You can build and run `pentoolkit` directly on your phone — no PC required.
Termux gives you a Linux shell on Android, and Go compiles a native ARM binary
right on the device.

> ⚠️ **Authorized use only.** Running this on your phone does not change the
> rules: only scan hosts you own or have written permission to test. Note that
> many mobile carriers and networks prohibit scanning — use your own lab or an
> authorized target.

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

## 3. Get the code

Clone the repository that contains `cmd/pentoolkit` (use your own fork's URL):

```sh
git clone https://github.com/sotaboy71/tools.git
cd tools
git checkout claude/pen-testing-toolkit-lay32c
```

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
```

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
