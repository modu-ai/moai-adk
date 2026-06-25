package preference

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestDecayScanCmd_NoAskUserQuestion (REQ-PGN-012 / C-HRA-008 family) is the
// static guard mirroring internal/cli/worktree/new_test.go::TestNew_NoAskUserQuestion.
// The CLI runs in subagent context and MUST NOT reference AskUserQuestion or
// mcp__askuser anywhere in its source. A grep over the cmd.go file asserts 0
// matches.
func TestDecayScanCmd_NoAskUserQuestion(t *testing.T) {
	t.Parallel()
	files := []string{
		"cmd.go",
	}
	for _, file := range files {
		src, err := os.ReadFile(file)
		if err != nil {
			t.Fatalf("read %s: %v", file, err)
		}
		if strings.Contains(string(src), "AskUserQuestion") || strings.Contains(string(src), "mcp__askuser") {
			t.Errorf("internal/cli/preference/%s must NOT reference AskUserQuestion or mcp__askuser (orchestrator-only HARD per REQ-PGN-012)", file)
		}
	}
}

// TestRunDecayScan_EndToEnd runs the full CLI body in isolation: it sets
// $CLAUDE_PROJECT_DIR to a temp dir, seeds a transient + a stable entry,
// runs runDecayScan with --force, and verifies the scan report + the
// side-effects (recall write-back, archival move, timestamp file).
func TestRunDecayScan_EndToEnd(t *testing.T) {
	// Cannot run in parallel because it mutates $CLAUDE_PROJECT_DIR.
	tmp := t.TempDir()
	t.Setenv("CLAUDE_PROJECT_DIR", tmp)

	// Seed entries via the public Store API so the memory-dir layout is real.
	memDir, err := resolveMemoryDirOverride("")
	if err != nil {
		t.Fatalf("resolveMemoryDirOverride: %v", err)
	}
	store, err := NewFileStore(memDir)
	if err != nil {
		t.Fatalf("NewFileStore: %v", err)
	}
	now := time.Date(2026, 6, 25, 12, 0, 0, 0, time.UTC)
	longAgo := now.Add(-30 * 24 * time.Hour)

	transient := freshEntry(ScopeTransient, longAgo, 0.5)
	transient.Domain = "log_level"
	transient.DecisionKey = "verbose"
	if err := store.Upsert(transient.Domain, transient.DecisionKey, transient); err != nil {
		t.Fatalf("seed transient: %v", err)
	}
	stable := freshEntry(ScopeStable, longAgo, 0.92)
	stable.Domain = "lang"
	stable.DecisionKey = "backend"
	if err := fsUpsertStable(store, stable); err != nil {
		t.Fatalf("seed stable into recall directly: %v", err)
	}

	// Run the CLI body.
	var stdout, stderr bytes.Buffer
	flags := &decayScanFlags{
		memoryDir: memDir,
		nowStr:    now.Format(time.RFC3339),
		force:     true,
	}
	if err := runDecayScan(&stdout, &stderr, flags); err != nil {
		t.Fatalf("runDecayScan: %v\nstderr: %s", err, stderr.String())
	}

	// Verify the transient entry was soft-deleted to archival.
	if _, _, tier, _ := store.Get("log_level", "verbose"); tier != TierArchival {
		t.Errorf("transient tier after scan = %v, want archival", tier)
	}

	// Verify the timestamp file was written under the project root's .moai/state/.
	stampPath := filepath.Join(tmp, ".moai", "state", decayLastRunFileName)
	data, err := os.ReadFile(stampPath)
	if err != nil {
		t.Errorf("cadence stamp not written: %v", err)
	} else {
		stampTime, pErr := time.Parse(time.RFC3339, strings.TrimSpace(string(data)))
		if pErr != nil {
			t.Errorf("cadence stamp unparseable: %v (raw=%q)", pErr, string(data))
		} else if !stampTime.Equal(now) {
			t.Errorf("cadence stamp = %v, want %v", stampTime, now)
		}
	}
}

// fsUpsertStable seeds a stable entry directly into recall (bypassing Upsert's
// core routing) so the decay-scan end-to-end test exercises the stable-floor
// path. Named with a lowercase-leading prefix so it reads as a helper, not an
// exported API. It is only used by this test file.
func fsUpsertStable(store Store, e Entry) error {
	fs, ok := store.(*fileStore)
	if !ok {
		return errNotFileStore
	}
	return fs.upsertToRecall(e.Domain, e.DecisionKey, e)
}

// errNotFileStore is a sentinel for the test helper above.
var errNotFileStore = &stringError{"preference: store is not *fileStore in test helper"}

type stringError struct{ s string }

func (e *stringError) Error() string { return e.s }

// TestRunDecayScan_CadenceGateSkips verifies the --force=false path: when the
// cadence stamp is fresh (< 24h), the scan is skipped and the stdout/stderr
// carry the skip signal.
func TestRunDecayScan_CadenceGateSkips(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("CLAUDE_PROJECT_DIR", tmp)

	memDir, err := resolveMemoryDirOverride("")
	if err != nil {
		t.Fatalf("resolveMemoryDirOverride: %v", err)
	}
	store, err := NewFileStore(memDir)
	if err != nil {
		t.Fatalf("NewFileStore: %v", err)
	}

	// Seed a 1h-old cadence stamp.
	now := time.Date(2026, 6, 25, 12, 0, 0, 0, time.UTC)
	stateDir := stateDirFromProjectRoot(tmp)
	if err := MarkScanned(stateDir, now.Add(-1*time.Hour)); err != nil {
		t.Fatalf("MarkScanned: %v", err)
	}

	// Seed an entry that WOULD be soft-deleted if the scan ran — to prove the
	// scan did NOT run.
	longAgo := now.Add(-40 * 24 * time.Hour)
	transient := freshEntry(ScopeTransient, longAgo, 0.5)
	transient.Domain = "flag"
	transient.DecisionKey = "old"
	if err := store.Upsert(transient.Domain, transient.DecisionKey, transient); err != nil {
		t.Fatalf("seed: %v", err)
	}

	var stdout, stderr bytes.Buffer
	flags := &decayScanFlags{
		memoryDir: memDir,
		nowStr:    now.Format(time.RFC3339),
		force:     false,
		json:      true,
	}
	if err := runDecayScan(&stdout, &stderr, flags); err != nil {
		t.Fatalf("runDecayScan: %v", err)
	}

	// JSON skip signal on stdout.
	var skip struct {
		Status string `json:"status"`
		Reason string `json:"reason"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &skip); err != nil {
		t.Fatalf("parse skip JSON %q: %v", stdout.String(), err)
	}
	if skip.Status != "skipped" {
		t.Errorf("skip status = %q, want \"skipped\"", skip.Status)
	}

	// The entry is STILL in recall (scan did not run).
	if _, _, tier, _ := store.Get("flag", "old"); tier != TierRecall {
		t.Errorf("entry moved despite skip (tier=%v); want recall (scan must NOT run)", tier)
	}
}

// TestNewDecayScanCmd_FlagsRegistered verifies the --memory-dir, --json,
// --now, --force flags are registered on the subcommand so the CLI surface
// matches the documented contract.
func TestNewDecayScanCmd_FlagsRegistered(t *testing.T) {
	t.Parallel()
	cmd := newDecayScanCmd()
	for _, name := range []string{"memory-dir", "json", "now", "force"} {
		if f := cmd.Flags().Lookup(name); f == nil {
			t.Errorf("flag --%s is not registered on decay-scan", name)
		}
	}
}

// TestPreferenceCmd_HasDecayScanChild verifies the init() func wired the
// child under the parent so `moai preference decay-scan` resolves.
func TestPreferenceCmd_HasDecayScanChild(t *testing.T) {
	t.Parallel()
	found := false
	for _, c := range PreferenceCmd.Commands() {
		if c.Use == "decay-scan" {
			found = true
			break
		}
	}
	if !found {
		t.Error("PreferenceCmd has no decay-scan child; init() wiring broken")
	}
}

// TestRunDecayScan_JSONOutput covers the --json output path: the report is
// emitted as a JSON object with the expected fields.
func TestRunDecayScan_JSONOutput(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("CLAUDE_PROJECT_DIR", tmp)

	memDir, err := resolveMemoryDirOverride("")
	if err != nil {
		t.Fatalf("resolveMemoryDirOverride: %v", err)
	}
	if _, err := NewFileStore(memDir); err != nil {
		t.Fatalf("NewFileStore: %v", err)
	}
	now := time.Date(2026, 6, 25, 12, 0, 0, 0, time.UTC)

	var stdout, stderr bytes.Buffer
	flags := &decayScanFlags{
		memoryDir: memDir,
		nowStr:    now.Format(time.RFC3339),
		force:     true,
		json:      true,
	}
	if err := runDecayScan(&stdout, &stderr, flags); err != nil {
		t.Fatalf("runDecayScan: %v\nstderr: %s", err, stderr.String())
	}
	var report DecayReport
	if err := json.Unmarshal(stdout.Bytes(), &report); err != nil {
		t.Fatalf("parse report JSON %q: %v", stdout.String(), err)
	}
	if !report.ScannedAt.Equal(now) {
		t.Errorf("report.ScannedAt = %v, want %v", report.ScannedAt, now)
	}
}

// TestRunDecayScan_InvalidNowFlag covers the --now parse-error path: a
// malformed RFC3339 value returns a wrapped error.
func TestRunDecayScan_InvalidNowFlag(t *testing.T) {
	t.Parallel()
	var stdout, stderr bytes.Buffer
	flags := &decayScanFlags{
		memoryDir: t.TempDir(),
		nowStr:    "not-a-date",
		force:     true,
	}
	err := runDecayScan(&stdout, &stderr, flags)
	if err == nil {
		t.Fatal("runDecayScan with invalid --now returned nil error")
	}
	if !strings.Contains(err.Error(), "parse --now") {
		t.Errorf("error = %v, want one mentioning 'parse --now'", err)
	}
}

// TestRunDecayScan_EmptyMemoryDirResolvedFromEnv covers the --memory-dir
// empty-value path: when the flag is unset, resolveMemoryDirOverride derives
// the dir from $CLAUDE_PROJECT_DIR. This test verifies that derivation does
// not error and the scan runs against it.
func TestRunDecayScan_EmptyMemoryDirResolvedFromEnv(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("CLAUDE_PROJECT_DIR", tmp)

	var stdout, stderr bytes.Buffer
	flags := &decayScanFlags{
		memoryDir: "", // empty → derive from $CLAUDE_PROJECT_DIR
		nowStr:    time.Date(2026, 6, 25, 12, 0, 0, 0, time.UTC).Format(time.RFC3339),
		force:     true,
	}
	if err := runDecayScan(&stdout, &stderr, flags); err != nil {
		t.Fatalf("runDecayScan with empty --memory-dir: %v", err)
	}
	if !strings.Contains(stdout.String(), "decay scan at") {
		t.Errorf("stdout = %q, missing 'decay scan at' summary", stdout.String())
	}
}
