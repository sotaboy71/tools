package main

import (
	"context"
	"flag"
	"fmt"
	"os/exec"
	"strings"
)

// extTool is a reference entry for a well-known, industry-standard security
// tool. pentoolkit does not bundle these — this is a directory that tells you
// what each one is, where to get it, and (with -check) whether it's already
// installed on this machine.
type extTool struct {
	Name   string `json:"name"`
	What   string `json:"what"`
	URL    string `json:"url"`
	MapsTo string `json:"maps_to,omitempty"` // the pentoolkit command that's a lightweight version
	Bin    string `json:"bin,omitempty"`     // command name to look for on PATH (empty = GUI/OS/platform)
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
		{"Nmap", "The classic network/port scanner and service/OS detector.", "https://nmap.org", "portscan, banner", "nmap"},
		{"Masscan", "An extremely fast, internet-scale port scanner.", "https://github.com/robertdavidgraham/masscan", "portscan", "masscan"},
		{"Amass", "In-depth subdomain enumeration and attack-surface mapping.", "https://github.com/owasp-amass/amass", "subenum", "amass"},
		{"Subfinder", "Fast passive subdomain discovery.", "https://github.com/projectdiscovery/subfinder", "subenum", "subfinder"},
	}},
	{"Web application testing", []extTool{
		{"Burp Suite", "The industry-standard intercepting proxy for testing web apps.", "https://portswigger.net/burp", "httpheaders, httpprobe", ""},
		{"OWASP ZAP", "A free, open-source web app scanner and proxy.", "https://www.zaproxy.org", "httpheaders, httpprobe", ""},
		{"ffuf", "Fast web fuzzer for content/parameter discovery.", "https://github.com/ffuf/ffuf", "httpprobe", "ffuf"},
		{"Gobuster", "Directory, DNS, and vhost brute-forcing.", "https://github.com/OJ/gobuster", "httpprobe, subenum", "gobuster"},
		{"Nikto", "Web server scanner for known issues and misconfigurations.", "https://github.com/sullo/nikto", "httpheaders", "nikto"},
		{"sqlmap", "Automated detection and testing of SQL injection.", "https://sqlmap.org", "", "sqlmap"},
	}},
	{"Exploitation frameworks", []extTool{
		{"Metasploit Framework", "The best-known framework for developing and running exploits and post-exploitation, used in authorized engagements.", "https://www.metasploit.com", "", "msfconsole"},
	}},
	{"Passwords & hashes", []extTool{
		{"Hashcat", "GPU-accelerated password-hash recovery (for auditing your own hashes).", "https://hashcat.net/hashcat/", "", "hashcat"},
		{"John the Ripper", "Long-standing password-hash cracker/auditor.", "https://www.openwall.com/john/", "", "john"},
		{"Hydra", "Network login testing tool for many protocols.", "https://github.com/vanhauser-thc/thc-hydra", "", "hydra"},
	}},
	{"Wireless (Wi-Fi)", []extTool{
		{"Aircrack-ng", "A suite for auditing the security of Wi-Fi networks you own.", "https://www.aircrack-ng.org", "", "aircrack-ng"},
		{"Kismet", "Wireless network detector, sniffer, and monitor.", "https://www.kismetwireless.net", "", "kismet"},
		{"Wifite", "Automated wrapper around common wireless-auditing tools.", "https://github.com/kimocoder/wifite2", "", "wifite"},
	}},
	{"Traffic capture & network", []extTool{
		{"Wireshark", "The standard GUI packet capture and analysis tool.", "https://www.wireshark.org", "", "wireshark"},
		{"tcpdump", "Command-line packet capture.", "https://www.tcpdump.org", "", "tcpdump"},
		{"Bettercap", "A framework for network monitoring and MitM research.", "https://www.bettercap.org", "tlsinfo", "bettercap"},
	}},
	{"All-in-one OS distributions", []extTool{
		{"Kali Linux", "A Linux distro that comes with most of these tools preinstalled.", "https://www.kali.org", "", ""},
		{"Parrot Security OS", "A security-focused distro, alternative to Kali.", "https://parrotsec.org", "", ""},
	}},
	{"Legal practice platforms", []extTool{
		{"HackTheBox", "Sandboxed, authorized machines and challenges to practice on.", "https://www.hackthebox.com", "", ""},
		{"TryHackMe", "Guided, beginner-friendly rooms and learning paths.", "https://tryhackme.com", "", ""},
		{"VulnHub", "Downloadable vulnerable VMs you run in your own lab.", "https://www.vulnhub.com", "", ""},
	}},
}

// runToolbox prints the reference catalog of well-known external tools, or with
// -check reports which of them are already installed on this machine.
func runToolbox(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("toolbox", flag.ContinueOnError)
	category := fs.String("category", "", "show only a category (partial match), e.g. web, wireless, recon")
	check := fs.Bool("check", false, "check which of these tools are installed on this machine (PATH lookup)")
	jsonOut := fs.Bool("json", false, "emit output as JSON")
	fs.Usage = func() {
		fmt.Fprintf(fs.Output(), "Usage: pentoolkit toolbox [-category NAME] [-check] [-json]\n\n")
		fmt.Fprintf(fs.Output(), "A directory of well-known security tools (Nmap, Metasploit, Wireshark,\n")
		fmt.Fprintf(fs.Output(), "aircrack-ng, ...): what each is and where to get it. pentoolkit does not\n")
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
			return fmt.Errorf("no category matches %q (try: pentoolkit toolbox)", *category)
		}
	}

	if *check {
		return runToolboxCheck(selected, *jsonOut)
	}

	return emit(*jsonOut, selected, func() {
		fmt.Println("pentoolkit — external tools reference")
		fmt.Println("Well-known tools you'll graduate to. pentoolkit does NOT include these;")
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
					fmt.Printf("  %-22s (pentoolkit equivalent: %s)\n", "", t.MapsTo)
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
}

// runToolboxCheck looks up each tool's command on PATH and reports the results.
func runToolboxCheck(cats []toolCategory, jsonOut bool) error {
	var results []checkResult
	var installed, total int
	for _, c := range cats {
		for _, t := range c.Tools {
			if t.Bin == "" {
				continue // GUI/OS/platform, not a CLI we can detect
			}
			total++
			r := checkResult{Name: t.Name, Bin: t.Bin, URL: t.URL}
			if path, err := exec.LookPath(t.Bin); err == nil {
				r.Installed = true
				r.Path = path
				installed++
			}
			results = append(results, r)
		}
	}

	return emit(jsonOut, results, func() {
		fmt.Printf("Installed CLI tools (%d of %d found on PATH):\n\n", installed, total)
		for _, r := range results {
			if r.Installed {
				fmt.Printf("  [+] %-16s %s\n", r.Name, r.Path)
			}
		}
		fmt.Println("\nNot found (install from the official site, or use Kali Linux):")
		for _, r := range results {
			if !r.Installed {
				fmt.Printf("  [-] %-16s %s\n", r.Name, r.URL)
			}
		}
		fmt.Println("\nNote: GUI apps and platforms (Burp, ZAP, Kali, HackTheBox, ...) aren't")
		fmt.Println("PATH commands, so they're not checked here — see 'pentoolkit toolbox'.")
	})
}
