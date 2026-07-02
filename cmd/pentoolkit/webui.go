package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"time"
)

// knownTools is the set of subcommands the web UI is allowed to invoke.
var knownTools = map[string]bool{
	"portscan":    true,
	"banner":      true,
	"httpheaders": true,
	"dns":         true,
	"tlsinfo":     true,
	"subenum":     true,
	"httpprobe":   true,
}

// runServe starts a small local web server that presents a mobile-friendly UI
// for every tool in the kit, plus built-in "how to run" help. Each run shells
// out to this same binary, so the web UI and CLI share identical behavior.
func runServe(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("serve", flag.ContinueOnError)
	addr := fs.String("addr", "127.0.0.1:8787", "address to listen on (host:port)")
	fs.Usage = func() {
		fmt.Fprintf(fs.Output(), "Usage: pentoolkit serve [-addr HOST:PORT]\n\n")
		fmt.Fprintf(fs.Output(), "Start the web UI. Defaults to localhost only.\n")
		fmt.Fprintf(fs.Output(), "To reach it from your phone on the same network, use -addr 0.0.0.0:8787\n\nFlags:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}

	self, err := os.Executable()
	if err != nil {
		return fmt.Errorf("locate own binary: %w", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, indexHTML)
	})
	mux.HandleFunc("/api/run", func(w http.ResponseWriter, r *http.Request) {
		handleRun(w, r, self)
	})

	srv := &http.Server{Addr: *addr, Handler: mux}

	fmt.Printf("pentoolkit web UI listening on http://%s\n", *addr)
	fmt.Println("Open that URL in your browser (including your phone's). Press Ctrl+C to stop.")

	// Shut down cleanly when the context is cancelled (e.g. Ctrl+C).
	go func() {
		<-ctx.Done()
		shutCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		srv.Shutdown(shutCtx)
	}()

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

// runRequest is the JSON body posted to /api/run.
type runRequest struct {
	Tool string   `json:"tool"`
	Args []string `json:"args"`
	JSON bool     `json:"json"`
}

// runResponse is returned from /api/run.
type runResponse struct {
	OK     bool   `json:"ok"`
	Output string `json:"output"`
	Error  string `json:"error,omitempty"`
}

func handleRun(w http.ResponseWriter, r *http.Request, self string) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		writeJSON(w, runResponse{Error: "use POST"})
		return
	}
	var req runRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, runResponse{Error: "bad request: " + err.Error()})
		return
	}
	if !knownTools[req.Tool] {
		writeJSON(w, runResponse{Error: "unknown tool: " + req.Tool})
		return
	}

	// Build the argument list. exec.Command with an explicit arg slice never
	// invokes a shell, so user input cannot cause command injection.
	cmdArgs := append([]string{req.Tool}, req.Args...)
	if req.JSON {
		cmdArgs = append(cmdArgs, "-json")
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Minute)
	defer cancel()

	cmd := exec.CommandContext(ctx, self, cmdArgs...)
	out, err := cmd.CombinedOutput()
	resp := runResponse{OK: err == nil, Output: string(out)}
	if err != nil {
		resp.Error = err.Error()
	}
	writeJSON(w, resp)
}

func writeJSON(w http.ResponseWriter, v interface{}) {
	json.NewEncoder(w).Encode(v)
}
