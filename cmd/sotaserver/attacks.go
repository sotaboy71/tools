package main

import (
	"context"
	"flag"
	"fmt"
	"strings"
)

// attackInfo is a defender-oriented reference entry for a common attack type.
// It intentionally describes what the attack is and how to DETECT and DEFEND
// against it — it is educational reference material, not an attack tool.
type attackInfo struct {
	ID       string   `json:"id"`
	Name     string   `json:"name"`
	Summary  string   `json:"summary"`
	Category string   `json:"category"`
	Detect   []string `json:"detect"`
	Defend   []string `json:"defend"`
}

// attackCatalog covers the 12 most common attack types, framed for defenders.
var attackCatalog = []attackInfo{
	{
		ID: "phishing", Name: "Phishing", Category: "Social",
		Summary: "Deceptive emails, messages, or websites that trick people into revealing credentials or financial details.",
		Detect: []string{
			"Mail gateway flags look-alike/spoofed sender domains and newly registered domains",
			"Users report suspicious messages via a one-click 'report phish' button",
			"Web proxy/DNS logs show clicks to newly seen or typosquatted domains",
		},
		Defend: []string{
			"Enforce phishing-resistant MFA (FIDO2/passkeys) so stolen passwords aren't enough",
			"Deploy DMARC, SPF, and DKIM to reduce spoofing of your domain",
			"Run recurring awareness training and easy reporting; filter mail at the gateway",
		},
	},
	{
		ID: "malware", Name: "Malware", Category: "Endpoint",
		Summary: "Malicious software (viruses, spyware, Trojans) that infiltrates, damages, or monitors devices.",
		Detect: []string{
			"EDR/AV alerts on known signatures and anomalous process/behavior",
			"Unexpected outbound connections to command-and-control infrastructure",
			"New autostart entries, scheduled tasks, or services appearing on hosts",
		},
		Defend: []string{
			"Keep OS and software patched; use application allow-listing",
			"Run modern EDR with behavioral detection and least-privilege accounts",
			"Restrict macros and block risky email attachment types at the gateway",
		},
	},
	{
		ID: "ransomware", Name: "Ransomware", Category: "Endpoint",
		Summary: "Malware that encrypts a victim's data and demands payment for the decryption key.",
		Detect: []string{
			"Spikes in file rename/modification rates and mass file-extension changes",
			"Shadow-copy deletion attempts and backup tampering",
			"Canary/honey files being modified",
		},
		Defend: []string{
			"Maintain tested, offline/immutable backups (follow 3-2-1) and rehearse restores",
			"Segment networks and enforce least privilege to limit blast radius",
			"Patch internet-facing services and require MFA on remote access",
		},
	},
	{
		ID: "dos", Name: "Denial-of-Service (DoS/DDoS)", Category: "Network",
		Summary: "Flooding a system or network with traffic so legitimate users can't use it.",
		Detect: []string{
			"Sudden traffic spikes far above baseline from many or spoofed sources",
			"Rising latency, connection-table exhaustion, and upstream saturation",
			"Netflow/telemetry showing traffic concentrated on one service or port",
		},
		Defend: []string{
			"Use an upstream DDoS-mitigation/CDN service and rate limiting",
			"Provision autoscaling and capacity headroom; drop malformed traffic early",
			"Have an incident runbook with your ISP/mitigation provider contacts ready",
		},
	},
	{
		ID: "mitm", Name: "Man-in-the-Middle (MitM)", Category: "Network",
		Summary: "An attacker secretly intercepts (and may alter) communications between two parties.",
		Detect: []string{
			"Unexpected TLS certificate changes or validation warnings (see 'tlsinfo')",
			"ARP table anomalies and duplicate MAC/IP bindings on the LAN",
			"Rogue access points or gateways advertising your SSID",
		},
		Defend: []string{
			"Enforce TLS everywhere with HSTS and certificate pinning where practical",
			"Use encrypted DNS and a VPN on untrusted networks",
			"Enable dynamic ARP inspection / DHCP snooping on managed switches",
		},
	},
	{
		ID: "credential", Name: "Password / Credential Attacks", Category: "Identity",
		Summary: "Guessing, stealing, or using automated tools (e.g. credential stuffing) to obtain valid logins.",
		Detect: []string{
			"Bursts of failed logins, impossible-travel, and logins from new geos/ASNs",
			"High authentication volume against many accounts from few sources",
			"Sign-ins using credentials seen in known breach corpora",
		},
		Defend: []string{
			"Require MFA and move toward passwordless (passkeys)",
			"Lockout/throttle and use breached-password screening; ban reuse",
			"Monitor identity logs and alert on anomalies; rotate exposed secrets",
		},
	},
	{
		ID: "social", Name: "Social Engineering", Category: "Social",
		Summary: "Psychological manipulation that tricks people into security mistakes or disclosures.",
		Detect: []string{
			"Reports of unusual urgent requests (gift cards, wire transfers, MFA prompts)",
			"MFA-fatigue push spikes and help-desk password-reset anomalies",
			"Pretexting patterns flagged by trained staff",
		},
		Defend: []string{
			"Define out-of-band verification for money/credential/access requests",
			"Number-matching MFA and strict help-desk identity-proofing procedures",
			"Continuous awareness training and a blameless reporting culture",
		},
	},
	{
		ID: "sqli", Name: "SQL Injection (SQLi)", Category: "AppSec",
		Summary: "Injecting malicious SQL into data-driven apps to read or manipulate the database.",
		Detect: []string{
			"WAF alerts and SQL syntax/errors appearing in application logs",
			"Anomalous query patterns or spikes in database errors",
			"Unexpected bulk reads from sensitive tables",
		},
		Defend: []string{
			"Use parameterized queries / prepared statements and ORMs correctly",
			"Apply least-privilege database accounts and strict input validation",
			"Add a WAF and review code; test in authorized assessments before release",
		},
	},
	{
		ID: "zeroday", Name: "Zero-Day Exploits", Category: "Vulnerability",
		Summary: "Attacks against newly discovered vulnerabilities before a patch exists.",
		Detect: []string{
			"Behavioral/anomaly detection rather than signatures (unknown by definition)",
			"Unexpected exploitation patterns and new IOCs from threat intel feeds",
			"Crash/exploit telemetry on exposed services",
		},
		Defend: []string{
			"Defense-in-depth: segmentation, least privilege, exploit mitigations (ASLR/DEP)",
			"Rapid patch pipeline and virtual patching/WAF rules when fixes lag",
			"Subscribe to vendor and CISA advisories; reduce attack surface",
		},
	},
	{
		ID: "insider", Name: "Insider Threats", Category: "Governance",
		Summary: "Breaches or leaks caused by people inside the organization, malicious or negligent.",
		Detect: []string{
			"UEBA flags abnormal access, large downloads, or off-hours activity",
			"DLP alerts on sensitive data leaving via mail, USB, or cloud",
			"Access to systems unrelated to a user's role",
		},
		Defend: []string{
			"Least privilege plus periodic access reviews and joiner/mover/leaver process",
			"Separation of duties and audit logging on sensitive systems",
			"DLP, offboarding checklists, and an insider-risk program",
		},
	},
	{
		ID: "supplychain", Name: "Supply Chain Attacks", Category: "Governance",
		Summary: "Compromising a less-secure vendor, supplier, or dependency to reach a larger target.",
		Detect: []string{
			"Unexpected changes in dependencies, build artifacts, or update signatures",
			"New/anomalous behavior after a third-party update",
			"SBOM diffs and integrity-check failures",
		},
		Defend: []string{
			"Maintain an SBOM; pin and verify dependencies with signatures/hashes",
			"Vet vendors, enforce least privilege for third-party access",
			"Secure the build pipeline (signed builds, provenance/SLSA)",
		},
	},
	{
		ID: "spoofing", Name: "Spoofing", Category: "Identity",
		Summary: "Impersonating a trusted person, device, or network to gain access or data.",
		Detect: []string{
			"Email auth failures (SPF/DKIM/DMARC) and display-name mismatches",
			"ARP/DNS/IP anomalies and unexpected certificate issuers",
			"GPS/caller-ID inconsistencies in relevant contexts",
		},
		Defend: []string{
			"Enforce DMARC/SPF/DKIM and mutual authentication where possible",
			"Use DNSSEC, validated TLS, and network anti-spoofing controls",
			"Verify identities out-of-band for sensitive actions",
		},
	},
}

// runAttacks prints the defender-oriented threat reference. With no argument it
// lists all entries; with a name/id it shows detail for the matching one(s).
func runAttacks(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("attacks", flag.ContinueOnError)
	name := fs.String("name", "", "show detail for a single attack (id or partial name), e.g. phishing")
	jsonOut := fs.Bool("json", false, "emit the catalog as JSON")
	fs.Usage = func() {
		fmt.Fprintf(fs.Output(), "Usage: sotaserver attacks [-name NAME] [-json]\n\n")
		fmt.Fprintf(fs.Output(), "Defender reference for the 12 most common attack types:\n")
		fmt.Fprintf(fs.Output(), "what each is, how to DETECT it, and how to DEFEND against it.\n\nFlags:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	// Allow a bare positional argument as the name too.
	if *name == "" && fs.NArg() > 0 {
		*name = strings.Join(fs.Args(), " ")
	}

	var selected []attackInfo
	if *name == "" {
		selected = attackCatalog
	} else {
		q := strings.ToLower(strings.TrimSpace(*name))
		for _, a := range attackCatalog {
			if a.ID == q || strings.Contains(strings.ToLower(a.Name), q) {
				selected = append(selected, a)
			}
		}
		if len(selected) == 0 {
			return fmt.Errorf("no attack matches %q (try: sotaserver attacks)", *name)
		}
	}

	return emit(*jsonOut, selected, func() {
		if *name == "" {
			fmt.Println("Sota's Pen Testing — threat reference (defender view)")
			fmt.Println("The 12 most common attack types. Run with -name <id> for detail.")
			fmt.Println()
			for _, a := range attackCatalog {
				fmt.Printf("  %-13s %s\n", a.ID, a.Name)
			}
			fmt.Println("\nExample: sotaserver attacks -name phishing")
			return
		}
		for _, a := range selected {
			printAttack(a)
		}
	})
}

func printAttack(a attackInfo) {
	fmt.Printf("== %s (%s) ==\n", a.Name, a.Category)
	fmt.Printf("%s\n\n", a.Summary)
	fmt.Println("Detect:")
	for _, d := range a.Detect {
		fmt.Printf("  - %s\n", d)
	}
	fmt.Println("\nDefend:")
	for _, d := range a.Defend {
		fmt.Printf("  - %s\n", d)
	}
	fmt.Println()
}
