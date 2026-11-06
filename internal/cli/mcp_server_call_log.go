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
	cur  string // the CURRENT segment file name (card-review P2-9) — empty until the first row
}

// mcpCallLog is the per-process call log. nil (tests constructing servers
// directly, or an unresolvable project dir) disables it — logging must never
// block a result, and a direct-construction test must never touch the real
// project's state tree.
var mcpCallLog *mcpCallLogWriter

// newMCPCallLogWriter resolves the log directory from the server's project
// dir; an unresolvable dir leaves the logger disabled. Card-review P2-11: a
// cwd that is not ALREADY a moai project (no .moai directory) disables the
// logger too — the log must never be the thing that creates a .moai tree in
// an arbitrary cwd (the REQ-MN-015 runtime-dir discipline).
func newMCPCallLogWriter() *mcpCallLogWriter {
	root := resolveProjectDir()
	if root == "" {
		return &mcpCallLogWriter{}
	}
	if info, err := os.Stat(filepath.Join(root, ".moai")); err != nil || !info.IsDir() {
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
	// Card-review P2-9: rotation reads the CURRENT segment's size, tracked on
	// the writer — re-statting the initial calls.jsonl on every row made each
	// post-rotation row rotate again (a segment per row), because the initial
	// file stayed over the limit forever.
	if w.cur == "" {
		w.cur = "calls.jsonl"
		w.size = fileSizeIfPresent(filepath.Join(w.dir, w.cur))
	}
	if w.size > mcpCallLogFileLimit {
		w.seq++
		w.cur = "calls." + strconv.Itoa(w.seq) + ".jsonl"
		w.size = fileSizeIfPresent(filepath.Join(w.dir, w.cur))
	}
	path := filepath.Join(w.dir, w.cur)
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return
	}
	defer func() { _ = f.Close() }()
	_, _ = f.Write(append(data, '\n'))
	w.size += int64(len(data)) + 1
}

// fileSizeIfPresent stats a segment file, 0 when absent.
func fileSizeIfPresent(path string) int64 {
	if info, err := os.Stat(path); err == nil {
		return info.Size()
	}
	return 0
}
