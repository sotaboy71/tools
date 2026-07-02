package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

// runPortscan performs a concurrent TCP connect scan against a single host.
//
// A connect scan completes the full TCP handshake, so it is easy to detect and
// intentionally non-stealthy. It only reports whether a port accepts
// connections; it does not attempt to exploit any service it finds.
func runPortscan(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("portscan", flag.ContinueOnError)
	host := fs.String("host", "", "target host (name or IP) - required")
	portspec := fs.String("ports", "1-1024", "ports to scan, e.g. \"22,80,443\" or \"1-1024\"")
	workers := fs.Int("workers", 100, "number of concurrent connection attempts")
	timeout := fs.Duration("timeout", 2*time.Second, "per-port connection timeout")
	openOnly := fs.Bool("open-only", true, "only report open ports")
	jsonOut := fs.Bool("json", false, "emit results as JSON")
	fs.Usage = func() {
		fmt.Fprintf(fs.Output(), "Usage: pentoolkit portscan -host HOST [-ports SPEC] [flags]\n\n")
		fmt.Fprintf(fs.Output(), "TCP connect scan of a host. Example:\n  pentoolkit portscan -host scanme.example -ports 1-1024\n\nFlags:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *host == "" {
		fs.Usage()
		return fmt.Errorf("missing required -host")
	}
	if *workers < 1 {
		return fmt.Errorf("-workers must be >= 1")
	}

	ports, err := parsePorts(*portspec)
	if err != nil {
		return err
	}

	if !*jsonOut {
		fmt.Printf("Scanning %s across %d port(s) with %d workers (timeout %s)...\n",
			*host, len(ports), *workers, *timeout)
	}

	type result struct {
		port int
		open bool
	}

	portsCh := make(chan int)
	resultsCh := make(chan result)
	var wg sync.WaitGroup

	dialer := &net.Dialer{Timeout: *timeout}
	for i := 0; i < *workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for port := range portsCh {
				addr := net.JoinHostPort(*host, strconv.Itoa(port))
				conn, err := dialer.DialContext(ctx, "tcp", addr)
				open := err == nil
				if conn != nil {
					conn.Close()
				}
				resultsCh <- result{port: port, open: open}
			}
		}()
	}

	go func() {
		defer close(portsCh)
		for _, p := range ports {
			select {
			case <-ctx.Done():
				return
			case portsCh <- p:
			}
		}
	}()

	go func() {
		wg.Wait()
		close(resultsCh)
	}()

	var open []int
	for r := range resultsCh {
		if r.open {
			open = append(open, r.port)
		} else if !*openOnly && !*jsonOut {
			fmt.Printf("  %5d/tcp closed\n", r.port)
		}
	}
	sort.Ints(open)

	type openPort struct {
		Port    int    `json:"port"`
		Service string `json:"service,omitempty"`
	}
	out := struct {
		Host  string     `json:"host"`
		Ports []openPort `json:"open_ports"`
	}{Host: *host}
	for _, p := range open {
		out.Ports = append(out.Ports, openPort{Port: p, Service: commonService(p)})
	}

	if err := emit(*jsonOut, out, func() {
		fmt.Printf("\n%d open port(s):\n", len(open))
		for _, p := range open {
			if svc := commonService(p); svc != "" {
				fmt.Printf("  %5d/tcp open   %s\n", p, svc)
			} else {
				fmt.Printf("  %5d/tcp open\n", p)
			}
		}
	}); err != nil {
		return err
	}
	return ctx.Err()
}

// parsePorts turns a spec like "22,80,443" or "1-1024" (or a mix) into a
// deduplicated, ordered list of port numbers.
func parsePorts(spec string) ([]int, error) {
	seen := make(map[int]bool)
	var ports []int
	add := func(p int) error {
		if p < 1 || p > 65535 {
			return fmt.Errorf("port %d out of range 1-65535", p)
		}
		if !seen[p] {
			seen[p] = true
			ports = append(ports, p)
		}
		return nil
	}

	for _, part := range strings.Split(spec, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if i := strings.IndexByte(part, '-'); i >= 0 {
			lo, err := strconv.Atoi(strings.TrimSpace(part[:i]))
			if err != nil {
				return nil, fmt.Errorf("invalid range %q: %v", part, err)
			}
			hi, err := strconv.Atoi(strings.TrimSpace(part[i+1:]))
			if err != nil {
				return nil, fmt.Errorf("invalid range %q: %v", part, err)
			}
			if lo > hi {
				lo, hi = hi, lo
			}
			for p := lo; p <= hi; p++ {
				if err := add(p); err != nil {
					return nil, err
				}
			}
			continue
		}
		p, err := strconv.Atoi(part)
		if err != nil {
			return nil, fmt.Errorf("invalid port %q: %v", part, err)
		}
		if err := add(p); err != nil {
			return nil, err
		}
	}
	if len(ports) == 0 {
		return nil, fmt.Errorf("no ports parsed from %q", spec)
	}
	return ports, nil
}

// commonService returns a friendly name for well-known ports, or "" if unknown.
func commonService(port int) string {
	switch port {
	case 21:
		return "ftp"
	case 22:
		return "ssh"
	case 23:
		return "telnet"
	case 25:
		return "smtp"
	case 53:
		return "dns"
	case 80:
		return "http"
	case 110:
		return "pop3"
	case 143:
		return "imap"
	case 443:
		return "https"
	case 445:
		return "smb"
	case 3306:
		return "mysql"
	case 3389:
		return "rdp"
	case 5432:
		return "postgresql"
	case 6379:
		return "redis"
	case 8080:
		return "http-alt"
	case 8443:
		return "https-alt"
	}
	return ""
}
