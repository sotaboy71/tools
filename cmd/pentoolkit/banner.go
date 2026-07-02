package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
	"unicode"
)

// runBanner connects to a single TCP port and prints whatever the service
// sends first, optionally after writing a probe string. This is useful for
// fingerprinting services (SSH, SMTP, HTTP, etc.) during authorized recon.
func runBanner(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("banner", flag.ContinueOnError)
	host := fs.String("host", "", "target host (name or IP) - required")
	port := fs.Int("port", 0, "target TCP port - required")
	timeout := fs.Duration("timeout", 5*time.Second, "connection and read timeout")
	probe := fs.String("probe", "", "optional string to send before reading (\\r\\n and \\n are expanded)")
	maxBytes := fs.Int("max-bytes", 2048, "maximum bytes to read from the banner")
	fs.Usage = func() {
		fmt.Fprintf(fs.Output(), "Usage: pentoolkit banner -host HOST -port PORT [flags]\n\n")
		fmt.Fprintf(fs.Output(), "Grab a service banner. Examples:\n")
		fmt.Fprintf(fs.Output(), "  pentoolkit banner -host example.com -port 22\n")
		fmt.Fprintf(fs.Output(), "  pentoolkit banner -host example.com -port 80 -probe \"HEAD / HTTP/1.0\\r\\n\\r\\n\"\n\nFlags:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *host == "" || *port == 0 {
		fs.Usage()
		return fmt.Errorf("both -host and -port are required")
	}
	if *port < 1 || *port > 65535 {
		return fmt.Errorf("port %d out of range 1-65535", *port)
	}

	addr := net.JoinHostPort(*host, strconv.Itoa(*port))
	dialer := &net.Dialer{Timeout: *timeout}
	conn, err := dialer.DialContext(ctx, "tcp", addr)
	if err != nil {
		return fmt.Errorf("connect: %w", err)
	}
	defer conn.Close()

	conn.SetDeadline(time.Now().Add(*timeout))

	if *probe != "" {
		expanded := strings.NewReplacer(`\r\n`, "\r\n", `\n`, "\n", `\t`, "\t").Replace(*probe)
		if _, err := conn.Write([]byte(expanded)); err != nil {
			return fmt.Errorf("write probe: %w", err)
		}
	}

	buf := make([]byte, *maxBytes)
	n, err := conn.Read(buf)
	if n == 0 && err != nil {
		return fmt.Errorf("read banner: %w", err)
	}

	fmt.Printf("Banner from %s (%d bytes):\n", addr, n)
	fmt.Println(strings.Repeat("-", 40))
	fmt.Println(printable(buf[:n]))
	fmt.Println(strings.Repeat("-", 40))
	return nil
}

// printable renders bytes for terminal display, replacing non-printable
// characters (other than common whitespace) with '.'.
func printable(b []byte) string {
	var sb strings.Builder
	for _, r := range string(b) {
		switch {
		case r == '\n' || r == '\r' || r == '\t':
			sb.WriteRune(r)
		case unicode.IsPrint(r):
			sb.WriteRune(r)
		default:
			sb.WriteByte('.')
		}
	}
	return strings.TrimRight(sb.String(), "\r\n")
}
