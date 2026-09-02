package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/pkg/version"
)

// writeRuntimeRecord is a test helper stamping an arbitrary record for a project.
func writeRuntimeRecord(t *testing.T, projectDir string, rec mcpServerRuntimeRecord) string {
	t.Helper()
	dir := mcpServerRuntimeDir(projectDir)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	data, err := json.Marshal(rec)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	path := filepath.Join(dir, strconv.Itoa(rec.PID)+".json")
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	return path
}

// TestWriteMCPServerRuntimeRecord_RoundTrip asserts the server's startup stamp
// lands under .moai/state/mcp-server/<pid>.json carrying this build's identity,
// and that the exit path removes it.
func TestWriteMCPServerRuntimeRecord_RoundTrip(t *testing.T) {
	projectDir := t.TempDir()

	path, err := writeMCPServerRuntimeRecord(projectDir)
	if err != nil {
		t.Fatalf("writeMCPServerRuntimeRecord: %v", err)
	}
	wantPath := filepath.Join(projectDir, ".moai", "state", "mcp-server", strconv.Itoa(os.Getpid())+".json")
	if path != wantPath {
		t.Errorf("record path = %q, want %q", path, wantPath)
	}

	records := readMCPServerRuntimeRecords(projectDir)
	if len(records) != 1 {
		t.Fatalf("read %d records, want 1", len(records))
	}
	rec := records[0]
	if rec.PID != os.Getpid() {
		t.Errorf("PID = %d, want %d", rec.PID, os.Getpid())
	}
	if rec.Version != version.GetVersion() {
		t.Errorf("Version = %q, want %q", rec.Version, version.GetVersion())
	}
	if rec.Commit != version.GetCommit() {
		t.Errorf("Commit = %q, want %q", rec.Commit, version.GetCommit())
	}
	if rec.StartedAt == "" {
		t.Error("StartedAt is empty; the stamp must record when the server started")
	}

	removeMCPServerRuntimeRecord(path)
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Errorf("record still present after remove: %v", err)
	}
}

// TestWriteMCPServerRuntimeRecord_EmptyProjectDir asserts the stamp is skipped
// rather than written to a relative path when the project dir cannot be
// resolved — the server must still serve.
func TestWriteMCPServerRuntimeRecord_EmptyProjectDir(t *testing.T) {
	path, err := writeMCPServerRuntimeRecord("")
	if err == nil {
		t.Error("expected an error for an empty project dir")
	}
	if path != "" {
		t.Errorf("path = %q, want empty", path)
	}
}

// TestReadMCPServerRuntimeRecords_SkipsMalformed asserts a corrupt or
// non-JSON file is skipped, not surfaced as an error — a bad stamp must not
// break the diagnostic.
func TestReadMCPServerRuntimeRecords_SkipsMalformed(t *testing.T) {
	projectDir := t.TempDir()
	dir := mcpServerRuntimeDir(projectDir)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "999999.json"), []byte("{not json"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "notes.txt"), []byte("ignored"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	writeRuntimeRecord(t, projectDir, mcpServerRuntimeRecord{PID: os.Getpid(), Commit: "abc123"})

	records := readMCPServerRuntimeRecords(projectDir)
	if len(records) != 1 {
		t.Fatalf("read %d records, want 1 (malformed + non-JSON skipped)", len(records))
	}
	if records[0].Commit != "abc123" {
		t.Errorf("Commit = %q, want abc123", records[0].Commit)
	}
}

// TestReadMCPServerRuntimeRecords_MissingDir asserts an absent record
// directory reads as no records rather than an error.
func TestReadMCPServerRuntimeRecords_MissingDir(t *testing.T) {
	if records := readMCPServerRuntimeRecords(t.TempDir()); len(records) != 0 {
		t.Errorf("read %d records from a project with no state dir, want 0", len(records))
	}
	if records := readMCPServerRuntimeRecords(""); len(records) != 0 {
		t.Errorf("read %d records for an empty project dir, want 0", len(records))
	}
}

// TestLiveMCPServerRuntimeRecords_SplitsByLiveness asserts a dead PID's stamp
// is reported as stale (so the caller can prune it) while this process's own
// stamp is reported live.
func TestLiveMCPServerRuntimeRecords_SplitsByLiveness(t *testing.T) {
	projectDir := t.TempDir()
	writeRuntimeRecord(t, projectDir, mcpServerRuntimeRecord{PID: os.Getpid(), Commit: "live"})
	// PID 0 never names a live process and is rejected by the record reader,
	// so use a PID that is valid-looking but almost certainly dead.
	deadPath := writeRuntimeRecord(t, projectDir, mcpServerRuntimeRecord{PID: 4194303, Commit: "dead"})

	live, stalePaths := liveMCPServerRuntimeRecords(projectDir)
	if len(live) != 1 || live[0].Commit != "live" {
		t.Fatalf("live = %+v, want exactly the running process's record", live)
	}
	if len(stalePaths) != 1 || stalePaths[0] != deadPath {
		t.Fatalf("stalePaths = %v, want [%s]", stalePaths, deadPath)
	}
}

// TestMoaiMCPServerInstructions_NamesBuildAndRestart asserts the initialize
// instructions carry the build stamp and state the restart requirement —
// the operator-facing half of the skew signal.
func TestMoaiMCPServerInstructions_NamesBuildAndRestart(t *testing.T) {
	got := moaiMCPServerInstructions()
	if !strings.Contains(got, version.GetFullVersion()) {
		t.Errorf("instructions %q do not name the build %q", got, version.GetFullVersion())
	}
	if !strings.Contains(got, "reconnect") {
		t.Errorf("instructions %q do not state the reconnect requirement", got)
	}
	if !strings.Contains(got, "moai doctor") {
		t.Errorf("instructions %q do not point at the doctor check", got)
	}
}
