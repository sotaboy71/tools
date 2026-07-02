package main

import (
	"context"
	"flag"
	"fmt"
	"strings"
)

// extTool is a reference entry for a well-known, industry-standard security
// tool. pentoolkit does not bundle these — this is a directory that tells you
// what each one is and where to get it from its official source.
type extTool struct {
	Name   string `json:"name"`
	What   string `json:"what"`
	URL    string `json:"url"`
	MapsTo string `json:"maps_to,omitempty"` // the pentoolkit command that's a lightweight version
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
		{"Nmap", "The classic network/port scanner and service/OS detector.", "https://nmap.org", "portscan, banner"},
		{"Masscan", "An extremely fast, internet-scale port scanner.", "https://github.com/robertdavidgraham/masscan", "portscan"},
		{"Amass", "In-depth subdomain enumeration and attack-surface mapping.", "https://github.com/owasp-amass/amass", "subenum"},
		{"Subfinder", "Fast passive subdomain discovery.", "https://github.com/projectdiscovery/subfinder", "subenum"},
	}},
	{"Web application testing", []extTool{
		{"Burp Suite", "The industry-standard intercepting proxy for testing web apps.", "https://portswigger.net/burp", "httpheaders, httpprobe"},
		{"OWASP ZAP", "A free, open-source web app scanner and proxy.", "https://www.zaproxy.org", "httpheaders, httpprobe"},
		{"ffuf", "Fast web fuzzer for content/parameter discovery.", "https://github.com/ffuf/ffuf", "httpprobe"},
		{"Gobuster", "Directory, DNS, and vhost brute-forcing.", "https://github.com/OJ/gobuster", "httpprobe, subenum"},
		{"Nikto", "Web server scanner for known issues and misconfigurations.", "https://github.com/sullo/nikto", "httpheaders"},
		{"sqlmap", "Automated detection and testing of SQL injection.", "https://sqlmap.org", ""},
	}},
	{"Exploitation frameworks", []extTool{
		{"Metasploit Framework", "The best-known framework for developing and running exploits and post-exploitation, used in authorized engagements.", "https://www.metasploit.com", ""},
	}},
	{"Passwords & hashes", []extTool{
		{"Hashcat", "GPU-accelerated password-hash recovery (for auditing your own hashes).", "https://hashcat.net/hashcat/", ""},
		{"John the Ripper", "Long-standing password-hash cracker/auditor.", "https://www.openwall.com/john/", ""},
		{"Hydra", "Network login testing tool for many protocols.", "https://github.com/vanhauser-thc/thc-hydra", ""},
	}},
	{"Wireless (Wi-Fi)", []extTool{
		{"Aircrack-ng", "A suite for auditing the security of Wi-Fi networks you own.", "https://www.aircrack-ng.org", ""},
		{"Kismet", "Wireless network detector, sniffer, and monitor.", "https://www.kismetwireless.net", ""},
		{"Wifite", "Automated wrapper around common wireless-auditing tools.", "https://github.com/kimocoder/wifite2", ""},
	}},
	{"Traffic capture & network", []extTool{
		{"Wireshark", "The standard GUI packet capture and analysis tool.", "https://www.wireshark.org", ""},
		{"tcpdump", "Command-line packet capture.", "https://www.tcpdump.org", ""},
		{"Bettercap", "A framework for network monitoring and MitM research.", "https://www.bettercap.org", "tlsinfo"},
	}},
	{"All-in-one OS distributions", []extTool{
		{"Kali Linux", "A Linux distro that comes with most of these tools preinstalled.", "https://www.kali.org", ""},
		{"Parrot Security OS", "A security-focused distro, alternative to Kali.", "https://parrotsec.org", ""},
	}},
	{"Legal practice platforms", []extTool{
		{"HackTheBox", "Sandboxed, authorized machines and challenges to practice on.", "https://www.hackthebox.com", ""},
		{"TryHackMe", "Guided, beginner-friendly rooms and learning paths.", "https://tryhackme.com", ""},
		{"VulnHub", "Downloadable vulnerable VMs you run in your own lab.", "https://www.vulnhub.com", ""},
	}},
}

// runToolbox prints the reference catalog of well-known external tools.
func runToolbox(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("toolbox", flag.ContinueOnError)
	category := fs.String("category", "", "show only a category (partial match), e.g. web, wireless, recon")
	jsonOut := fs.Bool("json", false, "emit the catalog as JSON")
	fs.Usage = func() {
		fmt.Fprintf(fs.Output(), "Usage: pentoolkit toolbox [-category NAME] [-json]\n\n")
		fmt.Fprintf(fs.Output(), "A directory of well-known security tools (Nmap, Metasploit, Wireshark,\n")
		fmt.Fprintf(fs.Output(), "aircrack-ng, ...): what each is and where to get it. pentoolkit does not\n")
		fmt.Fprintf(fs.Output(), "bundle these; install them from their official sources or via Kali Linux.\n\nFlags:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}

	var selected []toolCategory
	if *category == "" {
		selected = toolboxCatalog
	} else {
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

	return emit(*jsonOut, selected, func() {
		fmt.Println("pentoolkit — external tools reference")
		fmt.Println("Well-known tools you'll graduate to. pentoolkit does NOT include these;")
		fmt.Println("install them from their official sites (or use Kali Linux, which bundles many).")
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
