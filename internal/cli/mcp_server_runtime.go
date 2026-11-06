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
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
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
	// REQ-MN-015: the stamp must not CREATE a .moai tree in an arbitrary cwd —
	// a directory that is not already a moai project has no record to write.
	if info, err := os.Stat(filepath.Join(projectDir, ".moai")); err != nil || !info.IsDir() {
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

// ─── per-call staleness (SPEC-MCP-NEUTRALITY-001 REQ-MN-013) ──────────────────

// binaryIdentity is the executable identity captured ONCE at server start:
// a later in-place reinstall of the same path changes what a fresh os.Stat
// of that path reads, so the staleness comparison must run against the
// identity the server STARTED with — the AGENTS.md §14 reconnect contract
// rests on the stale server being able to recognize itself (card-review
// P1-3).
type binaryIdentity struct {
	Path     string
	Size     int64
	ModTime  time.Time
	Captured bool
}

// servingBinaryIdentity is captured at process start, before serving begins.
// Captured is false for a non-moai executable (the test process) or a capture
// failure — the check is then INERT, exactly as the old same-file reading was
// for a test binary.
var servingBinaryIdentity = captureExecutableIdentity()

// moaiLookPath is the installed-binary resolution seam (tests swap it).
var moaiLookPath = exec.LookPath

func captureExecutableIdentity() binaryIdentity {
	path, err := os.Executable()
	if err != nil {
		return binaryIdentity{}
	}
	if base := filepath.Base(path); !strings.HasPrefix(base, "moai") {
		// The test binary (or any non-moai executable): no identity, inert check.
		return binaryIdentity{}
	}
	if resolved, rerr := filepath.EvalSymlinks(path); rerr == nil {
		path = resolved
	}
	fi, err := os.Stat(path)
	if err != nil {
		return binaryIdentity{}
	}
	return binaryIdentity{Path: path, Size: fi.Size(), ModTime: fi.ModTime(), Captured: true}
}

// serverStaleness compares the identity the EXECUTABLE SERVING this call
// captured AT START against the installed `moai` binary on PATH: a different
// file (size or mtime moved since start) is the post-reinstall skew — the
// serving process keeps running the old build while the host's next spawn
// would use the new one. Comparing against the START-TIME identity (not a
// fresh stat of the serving path) is what an in-place reinstall cannot fool:
// both end-of-call stats of the same path read the new file and would call a
// stale server fresh (card-review P1-3). The judgment is BUILD identity
// (size+mtime), not location (t1632): the same build installed at a second
// PATH location is the same build — only a moved size or mtime is the skew.
// The check is INERT when no start-time identity was captured.
type serverStaleness struct {
	Stale     bool   `json:"stale_server"`
	Serving   string `json:"serving"`
	Installed string `json:"installed"`
	Version   string `json:"version"`
	Commit    string `json:"commit"`
}

func serverStalenessInfo() serverStaleness {
	info := serverStaleness{Version: version.GetVersion(), Commit: version.GetCommit()}
	if !servingBinaryIdentity.Captured {
		// Non-moai executable (the test process) or a capture failure: inert.
		return info
	}
	info.Serving = servingBinaryIdentity.Path
	installed, err := moaiLookPath("moai")
	if err != nil {
		return info
	}
	if resolved, err := filepath.EvalSymlinks(installed); err == nil {
		installed = resolved
	}
	info.Installed = installed
	installedStat, err := os.Stat(installed)
	if err != nil {
		return info
	}
	// t1632: freshness is judged on content identity — size+mtime, the data
	// the start-time capture already carries — never on the PATH alone. Why
	// size+mtime and not a content hash or a build stamp: the identity is
	// already captured at start (no extra binary read per call), it is the
	// same signal the same-path case above always trusted, a hash would
	// re-read the binary on every tool call, and a stamp would require
	// executing the installed binary. An in-place reinstall still moves size
	// or mtime, so the skew this check exists for stays detected.
	if installedStat.Size() == servingBinaryIdentity.Size &&
		installedStat.ModTime().Equal(servingBinaryIdentity.ModTime) {
		return info
	}
	info.Stale = true
	return info
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
