// mcp_server_call_log.go — SPEC-MCP-NEUTRALITY-001 M4 (card t1632,
// REQ-MN-015): one JSONL row per tool call under the project's
// .moai/state/mcp-server/<pid>/ directory, rotated by size. The row carries
// the tool, the duration, and the refused/error classification — the M1
// refused marking reused, so the call log distinguishes a policy refusal
// from a malfunction exactly as the wire does.
package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
)

// mcpCallLogFileLimit bounds one calls.jsonl segment before rotation.
const mcpCallLogFileLimit = 4 * 1024 * 1024

// mcpCallLogEntry is one tool-call row.
type mcpCallLogEntry struct {
	Timestamp string `json:"timestamp"`
	Tool      string `json:"tool"`
	Duration  string `json:"duration_ms"`
	Outcome   string `json:"outcome"` // ok | refused | error
	Stale     bool   `json:"stale,omitempty"`
}

// mcpCallLogWriter writes one JSONL row per tool call under the project's
// .moai/state/mcp-server/<pid>/ directory, rotated by size.
type mcpCallLogWriter struct {
	mu   sync.Mutex
	dir  string
	pid  int
	seq  int
	size int64
}

// mcpCallLog is the per-process call log. nil (tests constructing servers
// directly, or an unresolvable project dir) disables it — logging must never
// block a result, and a direct-construction test must never touch the real
// project's state tree.
var mcpCallLog *mcpCallLogWriter

// newMCPCallLogWriter resolves the log directory from the server's project
// dir; an unresolvable dir leaves the logger disabled.
func newMCPCallLogWriter() *mcpCallLogWriter {
	root := resolveProjectDir()
	if root == "" {
		return &mcpCallLogWriter{}
	}
	dir := filepath.Join(root, ".moai", "state", "mcp-server", strconv.Itoa(os.Getpid()))
	return &mcpCallLogWriter{dir: dir, pid: os.Getpid()}
}

// record writes one row. Fail-open: every error is swallowed — a call-log
// failure must never reach the wire.
func (w *mcpCallLogWriter) record(tool string, d time.Duration, res *mcp.CallToolResult, handlerErr error) {
	if w == nil || w.dir == "" {
		return
	}
	outcome := "ok"
	if handlerErr != nil || (res != nil && res.IsError) {
		outcome = "error"
	} else if res != nil {
		for _, c := range res.Content {
			if tc, ok := c.(mcp.TextContent); ok && strings.Contains(tc.Text, `"refused":true`) {
				outcome = "refused"
			}
		}
	}
	stale := false
	if res != nil && res.StructuredContent != nil {
		if body, err := json.Marshal(res.StructuredContent); err == nil && strings.Contains(string(body), `"stale_server":true`) {
			stale = true
		}
	}
	entry := mcpCallLogEntry{
		Timestamp: time.Now().UTC().Format(time.RFC3339Nano),
		Tool:      tool,
		Duration:  strconv.FormatInt(d.Milliseconds(), 10),
		Outcome:   outcome,
		Stale:     stale,
	}
	data, err := json.Marshal(entry)
	if err != nil {
		return
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	if err := os.MkdirAll(w.dir, 0o755); err != nil {
		return
	}
	path := w.currentPath()
	info, statErr := os.Stat(path)
	if statErr == nil && info.Size() > mcpCallLogFileLimit {
		w.seq++
		path = w.rotatedPath(w.seq)
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return
	}
	defer f.Close()
	_, _ = f.Write(append(data, '\n'))
	w.size += int64(len(data)) + 1
}

func (w *mcpCallLogWriter) currentPath() string {
	return filepath.Join(w.dir, "calls.jsonl")
}

func (w *mcpCallLogWriter) rotatedPath(seq int) string {
	return filepath.Join(w.dir, "calls."+strconv.Itoa(seq)+".jsonl")
}
