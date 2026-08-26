// Package cli — mcp_server_runtime.go
//
// Build-identity record for a RUNNING `moai mcp-server` process.
//
// The MCP server is spawned by the host (Claude Code, Cursor, ...) and lives
// for the whole host session. Reinstalling the binary mid-session does NOT
// restart it, so the host keeps talking to the previously-installed build —
// a silent version skew: newly added tools are absent from tools/list and
// fixed handlers keep running the old code path, with no signal anywhere.
//
// The server therefore stamps its own build identity to disk at startup and
// removes the stamp on exit. `moai doctor` reads the live stamps and compares
// them against the currently-installed binary (doctor_mcp_version.go), which
// turns the skew from invisible into a WARN with restart guidance.
//
// One file per PID, so two hosts running a server for the same project each
// get their own record and neither clobbers the other.
package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/modu-ai/moai-adk/pkg/version"
)

// mcpServerRuntimeRecord is the on-disk build identity of one running
// `moai mcp-server` process. Fields mirror pkg/version's build-time vars plus
// the process identity needed to tell a live server from a stale record.
type mcpServerRuntimeRecord struct {
	PID       int    `json:"pid"`
	Version   string `json:"version"`
	Commit    string `json:"commit"`
	BuildDate string `json:"build_date"`
	StartedAt string `json:"started_at"`
	// Executable is the resolved path of the running binary, best-effort
	// (empty when os.Executable fails). Diagnostic only — the commit is the
	// comparison key.
	Executable string `json:"executable,omitempty"`
}

// mcpServerRuntimeDir returns the directory holding per-PID server records for
// a project. Empty projectDir yields an empty path (caller skips the write).
func mcpServerRuntimeDir(projectDir string) string {
	if projectDir == "" {
		return ""
	}
	return filepath.Join(projectDir, ".moai", "state", "mcp-server")
}

// currentMCPServerRuntimeRecord builds the record describing THIS process.
func currentMCPServerRuntimeRecord() mcpServerRuntimeRecord {
	exe, err := os.Executable()
	if err != nil {
		exe = ""
	}
	return mcpServerRuntimeRecord{
		PID:        os.Getpid(),
		Version:    version.GetVersion(),
		Commit:     version.GetCommit(),
		BuildDate:  version.GetDate(),
		StartedAt:  time.Now().UTC().Format(time.RFC3339),
		Executable: exe,
	}
}

// writeMCPServerRuntimeRecord stamps this process's build identity under the
// project's state directory and returns the written path.
//
// Best-effort by contract: a read-only or missing state directory must never
// prevent the server from serving, so every failure returns ("", err) and the
// caller ignores it.
func writeMCPServerRuntimeRecord(projectDir string) (string, error) {
	dir := mcpServerRuntimeDir(projectDir)
	if dir == "" {
		return "", os.ErrInvalid
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	rec := currentMCPServerRuntimeRecord()
	data, err := json.MarshalIndent(rec, "", "  ")
	if err != nil {
		return "", err
	}
	path := filepath.Join(dir, strconv.Itoa(rec.PID)+".json")
	if err := os.WriteFile(path, append(data, '\n'), 0o644); err != nil {
		return "", err
	}
	return path, nil
}

// removeMCPServerRuntimeRecord deletes a stamp written by
// writeMCPServerRuntimeRecord. Best-effort: an empty path or a failed remove
// is ignored, because a leftover record is pruned by the doctor check's
// liveness probe anyway.
func removeMCPServerRuntimeRecord(path string) {
	if path == "" {
		return
	}
	_ = os.Remove(path)
}

// readMCPServerRuntimeRecords returns every parseable record in the project's
// server-record directory. Unreadable or malformed files are skipped rather
// than reported: a corrupt stamp must not turn a diagnostic into an error.
func readMCPServerRuntimeRecords(projectDir string) []mcpServerRuntimeRecord {
	dir := mcpServerRuntimeDir(projectDir)
	if dir == "" {
		return nil
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var records []mcpServerRuntimeRecord
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".json" {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			continue
		}
		var rec mcpServerRuntimeRecord
		if err := json.Unmarshal(data, &rec); err != nil || rec.PID <= 0 {
			continue
		}
		records = append(records, rec)
	}
	return records
}

// liveMCPServerRuntimeRecords splits records into those whose PID is still
// alive and the paths of those that are not, so the caller can prune the dead
// stamps a hard-killed server left behind.
func liveMCPServerRuntimeRecords(projectDir string) (live []mcpServerRuntimeRecord, stalePaths []string) {
	dir := mcpServerRuntimeDir(projectDir)
	for _, rec := range readMCPServerRuntimeRecords(projectDir) {
		if isProcessAlive(rec.PID) {
			live = append(live, rec)
			continue
		}
		stalePaths = append(stalePaths, filepath.Join(dir, strconv.Itoa(rec.PID)+".json"))
	}
	return live, stalePaths
}
