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
	fs.Usage = func() {
		fmt.Fprintf(fs.Output(), "Usage: pentoolkit tlsinfo -host HOST [-port 443] [flags]\n\n")
		fmt.Fprintf(fs.Output(), "Inspect a TLS certificate chain. Example:\n  pentoolkit tlsinfo -host example.com\n\nFlags:\n")
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
	fmt.Printf("TLS connection to %s (SNI %q)\n\n", addr, sni)
	fmt.Printf("Version:      %s\n", tlsVersionName(state.Version))
	fmt.Printf("Cipher suite: %s\n", tls.CipherSuiteName(state.CipherSuite))
	if state.NegotiatedProtocol != "" {
		fmt.Printf("ALPN:         %s\n", state.NegotiatedProtocol)
	}
	if state.Version < tls.VersionTLS12 {
		fmt.Printf("  [!] negotiated protocol is older than TLS 1.2\n")
	}

	now := time.Now()
	fmt.Printf("\nCertificate chain (%d):\n", len(state.PeerCertificates))
	for i, cert := range state.PeerCertificates {
		fmt.Printf("\n[%d] Subject: %s\n", i, cert.Subject)
		fmt.Printf("    Issuer:  %s\n", cert.Issuer)
		fmt.Printf("    Valid:   %s -> %s\n",
			cert.NotBefore.UTC().Format(time.RFC3339),
			cert.NotAfter.UTC().Format(time.RFC3339))
		switch {
		case now.Before(cert.NotBefore):
			fmt.Printf("    [!] not yet valid\n")
		case now.After(cert.NotAfter):
			fmt.Printf("    [!] EXPIRED\n")
		default:
			days := int(cert.NotAfter.Sub(now).Hours() / 24)
			fmt.Printf("    Expires in %d day(s)\n", days)
		}
		if len(cert.DNSNames) > 0 {
			fmt.Printf("    SANs:    %s\n", strings.Join(cert.DNSNames, ", "))
		}
	}
	return nil
}
