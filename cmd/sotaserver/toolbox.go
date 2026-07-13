package main

import (
	"context"
	"flag"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

// extTool is a reference entry for a well-known, industry-standard security
// tool. sotaserver does not bundle these — this is a directory that tells you
// what each one is, where to get it, and (with -check) whether it's already
// installed on this machine.
type extTool struct {
	Name   string `json:"name"`
	What   string `json:"what"`
	URL    string `json:"url"`
	MapsTo string `json:"maps_to,omitempty"` // the sotaserver command that's a lightweight version
	Bin    string `json:"bin,omitempty"`     // command name to look for on PATH (empty = GUI/OS/platform)
	Pkg    string `json:"pkg,omitempty"`     // common package name for install hints
}

// toolCategory groups related tools.
type toolCategory struct {
	Name  string    `json:"category"`
	Tools []extTool `json:"tools"`
}

// toolboxCatalog is a curated map of the common tools people graduate to. All
// are legitimate tools used in authorized security work; use them only on
// systems you own or are permitted to test.
var toolboxCatalog = []toolCategory{
	{"Reconnaissance & scanning", []extTool{
		{"Nmap", "The classic network/port scanner and service/OS detector.", "https://nmap.org", "portscan, banner", "nmap", "nmap"},
		{"Masscan", "An extremely fast, internet-scale port scanner.", "https://github.com/robertdavidgraham/masscan", "portscan", "masscan", "masscan"},
		{"Amass", "In-depth subdomain enumeration and attack-surface mapping.", "https://github.com/owasp-amass/amass", "subenum", "amass", "amass"},
		{"Subfinder", "Fast passive subdomain discovery.", "https://github.com/projectdiscovery/subfinder", "subenum", "subfinder", "subfinder"},
	}},
	{"Web application testing", []extTool{
		{"Burp Suite", "The industry-standard intercepting proxy for testing web apps.", "https://portswigger.net/burp", "httpheaders, httpprobe", "", ""},
		{"OWASP ZAP", "A free, open-source web app scanner and proxy.", "https://www.zaproxy.org", "httpheaders, httpprobe", "", ""},
		{"ffuf", "Fast web fuzzer for content/parameter discovery.", "https://github.com/ffuf/ffuf", "httpprobe", "ffuf", "ffuf"},
		{"Gobuster", "Directory, DNS, and vhost brute-forcing.", "https://github.com/OJ/gobuster", "httpprobe, subenum", "gobuster", "gobuster"},
		{"Nikto", "Web server scanner for known issues and misconfigurations.", "https://github.com/sullo/nikto", "httpheaders", "nikto", "nikto"},
		{"sqlmap", "Automated detection and testing of SQL injection.", "https://sqlmap.org", "", "sqlmap", "sqlmap"},
	}},
	{"Exploitation frameworks", []extTool{
		{"Metasploit Framework", "The best-known framework for developing and running exploits and post-exploitation, used in authorized engagements.", "https://www.metasploit.com", "", "msfconsole", "metasploit-framework"},
	}},
	{"Passwords & hashes", []extTool{
		{"Hashcat", "GPU-accelerated password-hash recovery (for auditing your own hashes).", "https://hashcat.net/hashcat/", "", "hashcat", "hashcat"},
		{"John the Ripper", "Long-standing password-hash cracker/auditor.", "https://www.openwall.com/john/", "", "john", "john"},
		{"Hydra", "Network login testing tool for many protocols.", "https://github.com/vanhauser-thc/thc-hydra", "", "hydra", "hydra"},
	}},
	{"Wireless (Wi-Fi)", []extTool{
		{"Aircrack-ng", "A suite for auditing the security of Wi-Fi networks you own.", "https://www.aircrack-ng.org", "", "aircrack-ng", "aircrack-ng"},
		{"Kismet", "Wireless network detector, sniffer, and monitor.", "https://www.kismetwireless.net", "", "kismet", "kismet"},
		{"Wifite", "Automated wrapper around common wireless-auditing tools.", "https://github.com/kimocoder/wifite2", "", "wifite", "wifite"},
	}},
	{"Traffic capture & network", []extTool{
		{"Wireshark", "The standard GUI packet capture and analysis tool.", "https://www.wireshark.org", "", "wireshark", "wireshark"},
		{"tcpdump", "Command-line packet capture.", "https://www.tcpdump.org", "", "tcpdump", "tcpdump"},
		{"Bettercap", "A framework for network monitoring and MitM research.", "https://www.bettercap.org", "tlsinfo", "bettercap", "bettercap"},
	}},
	{"All-in-one OS distributions", []extTool{
		{"Kali Linux", "A Linux distro that comes with most of these tools preinstalled.", "https://www.kali.org", "", "", ""},
		{"Parrot Security OS", "A security-focused distro, alternative to Kali.", "https://parrotsec.org", "", "", ""},
	}},
	{"Legal practice platforms", []extTool{
		{"HackTheBox", "Sandboxed, authorized machines and challenges to practice on.", "https://www.hackthebox.com", "", "", ""},
		{"TryHackMe", "Guided, beginner-friendly rooms and learning paths.", "https://tryhackme.com", "", "", ""},
		{"VulnHub", "Downloadable vulnerable VMs you run in your own lab.", "https://www.vulnhub.com", "", "", ""},
	}},
}

// runToolbox prints the reference catalog of well-known external tools, or with
// -check reports which of them are already installed on this machine.
func runToolbox(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("toolbox", flag.ContinueOnError)
	category := fs.String("category", "", "show only a category (partial match), e.g. web, wireless, recon")
	check := fs.Bool("check", false, "check which of these tools are installed on this machine (PATH lookup)")
	osHint := fs.String("os", "", "install-hint style for -check: apt, dnf, pacman, brew, pkg, winget (default: auto-detect)")
	jsonOut := fs.Bool("json", false, "emit output as JSON")
	fs.Usage = func() {
		fmt.Fprintf(fs.Output(), "Usage: sotaserver toolbox [-category NAME] [-check] [-json]\n\n")
		fmt.Fprintf(fs.Output(), "A directory of well-known security tools (Nmap, Metasploit, Wireshark,\n")
		fmt.Fprintf(fs.Output(), "aircrack-ng, ...): what each is and where to get it. sotaserver does not\n")
		fmt.Fprintf(fs.Output(), "bundle these; install them from their official sources or via Kali Linux.\n")
		fmt.Fprintf(fs.Output(), "Use -check to see which are already on your machine.\n\nFlags:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}

	selected := toolboxCatalog
	if *category != "" {
		selected = nil
		q := strings.ToLower(strings.TrimSpace(*category))
		for _, c := range toolboxCatalog {
			if strings.Contains(strings.ToLower(c.Name), q) {
				selected = append(selected, c)
			}
		}
		if len(selected) == 0 {
			return fmt.Errorf("no category matches %q (try: sotaserver toolbox)", *category)
		}
	}

	if *check {
		mgr := *osHint
		if mgr == "" {
			mgr = detectPkgMgr()
		}
		return runToolboxCheck(selected, mgr, *jsonOut)
	}

	return emit(*jsonOut, selected, func() {
		fmt.Println("Sota's Pen Testing — external tools reference")
		fmt.Println("Well-known tools you'll graduate to. sotaserver does NOT include these;")
		fmt.Println("install them from their official sites (or use Kali Linux, which bundles many).")
		fmt.Println("Run with -check to see which are already installed here.")
		fmt.Println("Use every tool only on systems you own or are authorized to test.")
		fmt.Println()
		for _, c := range selected {
			fmt.Printf("== %s ==\n", c.Name)
			for _, t := range c.Tools {
				fmt.Printf("  %-22s %s\n", t.Name, t.What)
				fmt.Printf("  %-22s %s\n", "", t.URL)
				if t.MapsTo != "" {
					fmt.Printf("  %-22s (sotaserver equivalent: %s)\n", "", t.MapsTo)
				}
			}
			fmt.Println()
		}
	})
}

// checkResult reports whether one tool was found on PATH.
type checkResult struct {
	Name      string `json:"name"`
	Bin       string `json:"bin"`
	Installed bool   `json:"installed"`
	Path      string `json:"path,omitempty"`
	URL       string `json:"url"`
	Install   string `json:"install,omitempty"` // suggested install command (missing tools)
}

// detectPkgMgr picks the package manager most likely available on this system.
func detectPkgMgr() string {
	if runtime.GOOS == "windows" {
		if _, err := exec.LookPath("winget"); err == nil {
			return "winget"
		}
		return ""
	}
	// Order matters: prefer the system's native manager over generic "pkg".
	for _, m := range []string{"apt", "dnf", "pacman", "brew", "pkg"} {
		if _, err := exec.LookPath(m); err == nil {
			return m
		}
	}
	return ""
}

// installHint returns the command to install pkg with the given manager.
func installHint(mgr, pkg string) string {
	if pkg == "" || mgr == "" {
		return ""
	}
	switch mgr {
	case "apt":
		return "sudo apt install " + pkg
	case "dnf":
		return "sudo dnf install " + pkg
	case "pacman":
		return "sudo pacman -S " + pkg
	case "brew":
		return "brew install " + pkg
	case "pkg":
		return "pkg install " + pkg
	case "winget":
		return "winget install " + pkg
	}
	return ""
}

// toolboxCheckResults looks up each tool's command on PATH and returns the
// results (installed path or a per-OS install hint). Shared by the CLI and the
// web UI.
func toolboxCheckResults(cats []toolCategory, mgr string) []checkResult {
	var results []checkResult
	for _, c := range cats {
		for _, t := range c.Tools {
			if t.Bin == "" {
				continue // GUI/OS/platform, not a CLI we can detect
			}
			r := checkResult{Name: t.Name, Bin: t.Bin, URL: t.URL}
			if path, err := exec.LookPath(t.Bin); err == nil {
				r.Installed = true
				r.Path = path
			} else {
				r.Install = installHint(mgr, t.Pkg)
			}
			results = append(results, r)
		}
	}
	return results
}

// runToolboxCheck looks up each tool's command on PATH and reports the results,
// including a per-OS install hint for the ones that are missing.
func runToolboxCheck(cats []toolCategory, mgr string, jsonOut bool) error {
	results := toolboxCheckResults(cats, mgr)
	installed, total := 0, len(results)
	for _, r := range results {
		if r.Installed {
			installed++
		}
	}

	return emit(jsonOut, results, func() {
		label := mgr
		if label == "" {
			label = "unknown — pass -os apt|dnf|pacman|brew|pkg|winget"
		}
		fmt.Printf("Installed CLI tools (%d of %d found on PATH). Package manager: %s\n\n", installed, total, label)
		for _, r := range results {
			if r.Installed {
				fmt.Printf("  [+] %-16s %s\n", r.Name, r.Path)
			}
		}
		fmt.Println("\nNot found — install with:")
		for _, r := range results {
			if r.Installed {
				continue
			}
			if r.Install != "" {
				fmt.Printf("  [-] %-16s %s\n", r.Name, r.Install)
			} else {
				fmt.Printf("  [-] %-16s %s\n", r.Name, r.URL)
			}
		}
		fmt.Println("\nHints assume the package name is the common one; it can vary by distro,")
		fmt.Println("and a few tools (masscan, subfinder, ffuf, ...) may need their own repo or")
		fmt.Println("`go install`. Kali Linux ships them all. GUI apps/platforms aren't listed.")
	})
}
