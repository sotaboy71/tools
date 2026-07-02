package main

import (
	"encoding/json"
	"os"
)

// emit renders a command's result. When jsonOut is true it prints result as
// indented JSON; otherwise it calls text to produce human-readable output.
func emit(jsonOut bool, result interface{}, text func()) error {
	if jsonOut {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(result)
	}
	text()
	return nil
}
