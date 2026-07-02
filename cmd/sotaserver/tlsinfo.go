package main

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
)

// tlsVersionName maps a TLS version constant to a human-readable name.
func tlsVersionName(v uint16) string {
	switch v {
	case tls.VersionTLS10:
		return "TLS 1.0"
	case tls.VersionTLS11:
		return "TLS 1.1"
	case tls.VersionTLS12:
		return "TLS 1.2"
	case tls.VersionTLS13:
		return "TLS 1.3"
	}
	return fmt.Sprintf("0x%04x", v)
}

// runTLSInfo connects to a host over TLS and reports on the negotiated
// protocol version, cipher suite, and the presented certificate chain
// (subjects, issuers, validity, and SANs). Certificate verification is
// skipped so that expired or self-signed certificates can still be inspected.
func runTLSInfo(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("tlsinfo", flag.ContinueOnError)
	host := fs.String("host", "", "target host (name or IP) - required")
	port := fs.Int("port", 443, "target TLS port")
	timeout := fs.Duration("timeout", 10*time.Second, "connection timeout")
	server := fs.String("servername", "", "SNI server name (defaults to -host)")
	jsonOut := fs.Bool("json", false, "emit results as JSON")
	fs.Usage = func() {
		fmt.Fprintf(fs.Output(), "Usage: sotaserver tlsinfo -host HOST [-port 443] [flags]\n\n")
		fmt.Fprintf(fs.Output(), "Inspect a TLS certificate chain. Example:\n  sotaserver tlsinfo -host example.com\n\nFlags:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *host == "" {
		fs.Usage()
		return fmt.Errorf("missing required -host")
	}
	sni := *server
	if sni == "" {
		sni = *host
	}

	addr := net.JoinHostPort(*host, strconv.Itoa(*port))
	dialer := &net.Dialer{Timeout: *timeout}
	rawConn, err := dialer.DialContext(ctx, "tcp", addr)
	if err != nil {
		return fmt.Errorf("connect: %w", err)
	}
	defer rawConn.Close()

	// Skip verification on purpose: we want to inspect whatever the server
	// presents, including expired or self-signed certificates.
	conn := tls.Client(rawConn, &tls.Config{
		ServerName:         sni,
		InsecureSkipVerify: true,
	})
	rawConn.SetDeadline(time.Now().Add(*timeout))
	if err := conn.HandshakeContext(ctx); err != nil {
		return fmt.Errorf("TLS handshake: %w", err)
	}

	state := conn.ConnectionState()
	now := time.Now()

	out := tlsReport{
		Address:     addr,
		SNI:         sni,
		Version:     tlsVersionName(state.Version),
		CipherSuite: tls.CipherSuiteName(state.CipherSuite),
		ALPN:        state.NegotiatedProtocol,
		Weak:        state.Version < tls.VersionTLS12,
	}
	for _, cert := range state.PeerCertificates {
		ci := certInfo{
			Subject:   cert.Subject.String(),
			Issuer:    cert.Issuer.String(),
			NotBefore: cert.NotBefore.UTC().Format(time.RFC3339),
			NotAfter:  cert.NotAfter.UTC().Format(time.RFC3339),
			SANs:      cert.DNSNames,
		}
		switch {
		case now.Before(cert.NotBefore):
			ci.Status = "not_yet_valid"
		case now.After(cert.NotAfter):
			ci.Status = "expired"
		default:
			ci.Status = "valid"
			ci.ExpiresInDays = int(cert.NotAfter.Sub(now).Hours() / 24)
		}
		out.Chain = append(out.Chain, ci)
	}

	return emit(*jsonOut, out, func() {
		fmt.Printf("TLS connection to %s (SNI %q)\n\n", addr, sni)
		fmt.Printf("Version:      %s\n", out.Version)
		fmt.Printf("Cipher suite: %s\n", out.CipherSuite)
		if out.ALPN != "" {
			fmt.Printf("ALPN:         %s\n", out.ALPN)
		}
		if out.Weak {
			fmt.Printf("  [!] negotiated protocol is older than TLS 1.2\n")
		}
		fmt.Printf("\nCertificate chain (%d):\n", len(out.Chain))
		for i, ci := range out.Chain {
			fmt.Printf("\n[%d] Subject: %s\n", i, ci.Subject)
			fmt.Printf("    Issuer:  %s\n", ci.Issuer)
			fmt.Printf("    Valid:   %s -> %s\n", ci.NotBefore, ci.NotAfter)
			switch ci.Status {
			case "not_yet_valid":
				fmt.Printf("    [!] not yet valid\n")
			case "expired":
				fmt.Printf("    [!] EXPIRED\n")
			default:
				fmt.Printf("    Expires in %d day(s)\n", ci.ExpiresInDays)
			}
			if len(ci.SANs) > 0 {
				fmt.Printf("    SANs:    %s\n", strings.Join(ci.SANs, ", "))
			}
		}
	})
}

// tlsReport is the JSON shape of the tlsinfo command.
type tlsReport struct {
	Address     string     `json:"address"`
	SNI         string     `json:"sni"`
	Version     string     `json:"version"`
	CipherSuite string     `json:"cipher_suite"`
	ALPN        string     `json:"alpn,omitempty"`
	Weak        bool       `json:"weak_protocol"`
	Chain       []certInfo `json:"certificate_chain"`
}

// certInfo describes a single certificate in the chain.
type certInfo struct {
	Subject       string   `json:"subject"`
	Issuer        string   `json:"issuer"`
	NotBefore     string   `json:"not_before"`
	NotAfter      string   `json:"not_after"`
	Status        string   `json:"status"`
	ExpiresInDays int      `json:"expires_in_days,omitempty"`
	SANs          []string `json:"sans,omitempty"`
}
