package sync

import (
	"fmt"
	"os"
	"path/filepath"
)

// appendLog appends a single line to logPath, fail-open on any error.
// Mirrors `internal/cli/navigator_enrich.go:152 appendLog` (REQ-NS-017 sink).
func appendLog(logPath, line string) {
	if logPath == "" {
		return
	}
	_ = os.MkdirAll(filepath.Dir(logPath), 0o755)
	f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return
	}
	defer func() { _ = f.Close() }()
	_, _ = fmt.Fprintln(f, line)
}

// relOrRoot returns path relative to root when resolvable, else path verbatim.
func relOrRoot(root, path string) string {
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return path
	}
	return rel
}
