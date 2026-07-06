// Package statusline tests for SPEC-HANDOFF-THRESHOLD-001 (Handoff-v2 M4) D3:
// context-usage.json persistence — atomic best-effort write, write-if-changed
// throttle, schema, session_id guard, fallback-UUID freshness, and the
// concurrent empty-session_id writer_pid discriminator.
package statusline

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestMain keeps the source tree clean. builder.Build writes
// <cwd>/.moai/state/context-usage.json as a best-effort side effect (D3).
// Existing Build tests that pass stdin without workspace info resolve the
// project dir to os.Getwd() (this package directory), which would otherwise
// leave a runtime artifact under version control (context-usage.json is a
// runtime artifact, never committed — plan.md §B8). We record whether ".moai"
// pre-existed and remove it after the run if the tests created it. No chdir is
// performed, so the os.Getwd()-basename assertions elsewhere in the package are
// preserved. Tests that need a real write target t.TempDir().
func TestMain(m *testing.M) {
	_, preErr := os.Stat(".moai")
	moaiPreExisted := preErr == nil

	code := m.Run()

	if !moaiPreExisted {
		_ = os.RemoveAll(".moai")
	}
	os.Exit(code)
}

// AC-THRESHOLD-007 static assertion: writeContextUsage takes NO HandoffConfig /
// cfg parameter — the state-file write is a pure function of usage, never gated
// by Mode/Guide. A compile error here means REQ-THRESHOLD-007 was violated.
var _ func(string, string, int, MemoryData, handoffStage) = writeContextUsage

// readRecord is a test helper that reads + parses the on-disk record.
func readRecord(t *testing.T, path string) *contextUsageRecord {
	t.Helper()
	rec, err := readContextUsage(path)
	if err != nil {
		t.Fatalf("readContextUsage(%q) error: %v", path, err)
	}
	return rec
}

func usagePath(projDir string) string {
	return filepath.Join(projDir, ".moai", "state", "context-usage.json")
}

// TestWriteContextUsage_ConfigIndependent — AC-THRESHOLD-007.
// The write happens regardless of any handoff config (there is no config param);
// it depends only on Memory availability.
func TestWriteContextUsage_ConfigIndependent(t *testing.T) {
	t.Parallel()

	proj := t.TempDir()
	m := MemoryData{ContextWindowSize: 256_000, TokensUsed: 230_400, Available: true}
	writeContextUsage(proj, "sess-1", 4242, m, handoffStageSoft)

	rec := readRecord(t, usagePath(proj))
	if rec.SessionID != "sess-1" {
		t.Errorf("session_id = %q, want sess-1", rec.SessionID)
	}
	if rec.Stage != "soft" {
		t.Errorf("stage = %q, want soft", rec.Stage)
	}
}

// TestWriteContextUsage_Atomic — AC-THRESHOLD-009 (happy path).
// A normal call writes a valid JSON record via temp+rename; no stray .tmp file
// is left behind.
func TestWriteContextUsage_Atomic(t *testing.T) {
	t.Parallel()

	proj := t.TempDir()
	m := MemoryData{ContextWindowSize: 1_000_000, TokensUsed: 500_000, Available: true}
	writeContextUsage(proj, "sess-atomic", 100, m, handoffStageSoft)

	path := usagePath(proj)
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("context-usage.json not created: %v", err)
	}
	// Valid JSON (readContextUsage parses it).
	_ = readRecord(t, path)
	// No leftover temp file.
	if _, err := os.Stat(path + ".tmp"); !os.IsNotExist(err) {
		t.Errorf("temp file %q should not exist after atomic rename", path+".tmp")
	}
}

// TestWriteContextUsage_SilentFail — AC-THRESHOLD-009 (best-effort).
// Unwritable / unresolvable targets and absent source signals never panic and
// never surface an error (the function has no error return); no file is written.
func TestWriteContextUsage_SilentFail(t *testing.T) {
	t.Parallel()

	m := MemoryData{ContextWindowSize: 256_000, TokensUsed: 230_400, Available: true}

	// (a) empty projDir → skip.
	writeContextUsage("", "s", 1, m, handoffStageSoft)

	// (b) Memory.Available == false → skip (no source signal).
	proj := t.TempDir()
	writeContextUsage(proj, "s", 1, MemoryData{Available: false}, handoffStageNone)
	if _, err := os.Stat(usagePath(proj)); !os.IsNotExist(err) {
		t.Errorf("no file expected when Memory.Available == false")
	}

	// (c) non-positive window → skip (avoids div-by-zero).
	proj2 := t.TempDir()
	writeContextUsage(proj2, "s", 1, MemoryData{ContextWindowSize: 0, Available: true}, handoffStageNone)
	if _, err := os.Stat(usagePath(proj2)); !os.IsNotExist(err) {
		t.Errorf("no file expected when ContextWindowSize <= 0")
	}

	// (d) projDir whose parent is a regular file → MkdirAll fails → silent.
	fileParent := filepath.Join(t.TempDir(), "not-a-dir")
	if err := os.WriteFile(fileParent, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	blocked := filepath.Join(fileParent, "sub") // fileParent is a file, so MkdirAll fails
	writeContextUsage(blocked, "s", 1, m, handoffStageSoft)
	// No panic reaching here == pass; nothing to assert on the blocked path.
}

// TestContextUsage_Schema — AC-THRESHOLD-010.
// The record carries all 9 required fields with the expected value shapes.
func TestContextUsage_Schema(t *testing.T) {
	t.Parallel()

	proj := t.TempDir()
	m := MemoryData{ContextWindowSize: 256_000, TokensUsed: 245_760, Available: true}
	writeContextUsage(proj, "sess-schema", 7777, m, handoffStageHard)

	raw, err := os.ReadFile(usagePath(proj))
	if err != nil {
		t.Fatal(err)
	}
	var obj map[string]any
	if err := json.Unmarshal(raw, &obj); err != nil {
		t.Fatalf("record is not valid JSON: %v", err)
	}
	for _, field := range []string{
		"schema_version", "session_id", "writer_pid", "captured_at",
		"context_window_size", "tokens_used", "raw_pct", "stage", "band",
	} {
		if _, ok := obj[field]; !ok {
			t.Errorf("required field %q missing from record", field)
		}
	}

	rec := readRecord(t, usagePath(proj))
	if rec.SchemaVersion != contextUsageSchemaVersion {
		t.Errorf("schema_version = %d, want %d", rec.SchemaVersion, contextUsageSchemaVersion)
	}
	if rec.WriterPID != 7777 {
		t.Errorf("writer_pid = %d, want 7777", rec.WriterPID)
	}
	if rec.Stage != "hard" {
		t.Errorf("stage = %q, want hard", rec.Stage)
	}
	if rec.Band != "standard" { // 256K < 500K cutoff
		t.Errorf("band = %q, want standard", rec.Band)
	}
	// large band label at/above the cutoff.
	if got := bandLabel(500_000); got != "large" {
		t.Errorf("bandLabel(500K) = %q, want large", got)
	}
}

// TestWriteContextUsage_ThrottleSkipUnchanged — AC-THRESHOLD-012.
// A second write with an unchanged semantic payload is skipped (mtime frozen);
// a changed payload (different int raw_pct) triggers a fresh write.
func TestWriteContextUsage_ThrottleSkipUnchanged(t *testing.T) {
	t.Parallel()

	proj := t.TempDir()
	path := usagePath(proj)
	m := MemoryData{ContextWindowSize: 256_000, TokensUsed: 230_400, Available: true} // 90%

	writeContextUsage(proj, "sess-throttle", 11, m, handoffStageSoft)

	// Force an old mtime so a skip (no rewrite) is detectable.
	old := time.Now().Add(-2 * time.Hour)
	if err := os.Chtimes(path, old, old); err != nil {
		t.Fatal(err)
	}

	// Same semantic payload → skip (mtime must stay old).
	writeContextUsage(proj, "sess-throttle", 22, m, handoffStageSoft) // writer_pid differs (22) — still throttled
	st1, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if !st1.ModTime().Equal(old) {
		t.Errorf("unchanged payload should be throttled (mtime changed: %v != %v)", st1.ModTime(), old)
	}

	// Changed payload (different tokens → different int raw_pct) → write occurs.
	m2 := MemoryData{ContextWindowSize: 256_000, TokensUsed: 200_000, Available: true} // ~78%
	writeContextUsage(proj, "sess-throttle", 11, m2, handoffStageNone)
	st2, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if st2.ModTime().Equal(old) {
		t.Errorf("changed payload should trigger a fresh write (mtime still old)")
	}
}

// TestSessionGuard_MismatchStale — AC-THRESHOLD-013.
// A record stamped with session_id A is stale for a reader in session B, and
// valid for a reader in session A. writer_pid is irrelevant on the UUID path.
func TestSessionGuard_MismatchStale(t *testing.T) {
	t.Parallel()

	rec := &contextUsageRecord{
		SessionID:  "sess-A",
		WriterPID:  1001,
		CapturedAt: time.Now().Format(time.RFC3339Nano),
	}
	if isFreshForSession(rec, "sess-B", 1001) {
		t.Errorf("session_id mismatch (A vs B) must be stale")
	}
	if !isFreshForSession(rec, "sess-A", 1001) {
		t.Errorf("session_id match (A) must be valid")
	}
	// UUID path ignores writer_pid: mismatched writer, same UUID → still valid.
	if !isFreshForSession(rec, "sess-A", 9999) {
		t.Errorf("UUID path must ignore writer_pid (same UUID → valid)")
	}
}

// TestFallbackUUID_FreshnessValidation — AC-THRESHOLD-014.
// With no real UUID on either side (empty session_id), validity is decided by
// captured_at freshness (with a matching single-session writer_pid). A mixed
// UUID-vs-empty pair is conservatively stale. The write still happens when the
// session_id is empty (primary path survives for the common single-session case).
func TestFallbackUUID_FreshnessValidation(t *testing.T) {
	t.Parallel()

	// (a) write still occurs with an empty session_id.
	proj := t.TempDir()
	m := MemoryData{ContextWindowSize: 200_000, TokensUsed: 180_000, Available: true}
	writeContextUsage(proj, "", 5150, m, handoffStageSoft)
	if _, err := os.Stat(usagePath(proj)); err != nil {
		t.Fatalf("empty session_id must still write the record: %v", err)
	}

	const pid = 5150
	fresh := &contextUsageRecord{SessionID: "", WriterPID: pid, CapturedAt: time.Now().Format(time.RFC3339Nano)}
	expired := &contextUsageRecord{SessionID: "", WriterPID: pid, CapturedAt: time.Now().Add(-13 * time.Hour).Format(time.RFC3339Nano)}

	// (b) both empty + fresh + matching writer → valid (single-session survives).
	if !isFreshForSession(fresh, "", pid) {
		t.Errorf("empty/empty + fresh + same writer must be valid")
	}
	// captured_at expired → stale.
	if isFreshForSession(expired, "", pid) {
		t.Errorf("empty/empty + expired must be stale")
	}
	// mixed: record UUID vs empty current → conservatively stale.
	uuidRec := &contextUsageRecord{SessionID: "sess-X", WriterPID: pid, CapturedAt: time.Now().Format(time.RFC3339Nano)}
	if isFreshForSession(uuidRec, "", pid) {
		t.Errorf("UUID-vs-empty mix must be conservatively stale")
	}
	// mixed the other way: empty record vs UUID current → conservatively stale.
	if isFreshForSession(fresh, "sess-X", pid) {
		t.Errorf("empty-vs-UUID mix must be conservatively stale")
	}
}

// TestConcurrentEmptyID_WriterPIDGuard — AC-THRESHOLD-018.
// Two concurrent empty-session_id records distinguished only by writer_pid: a
// reader with curWriterID matching recA reads recA as its own (valid) but reads
// recB (different writer) as stale — closing the cross-read hole that captured_at
// freshness alone would leave open. table: writer_pid match/mismatch × fresh/stale.
func TestConcurrentEmptyID_WriterPIDGuard(t *testing.T) {
	t.Parallel()

	now := time.Now().Format(time.RFC3339Nano)
	stale := time.Now().Add(-13 * time.Hour).Format(time.RFC3339Nano)

	tests := []struct {
		name       string
		recPID     int
		recCapture string
		curPID     int
		want       bool
	}{
		{"match+fresh → valid", 1001, now, 1001, true},
		{"mismatch+fresh → stale (cross-read blocked)", 1002, now, 1001, false},
		{"match+stale → stale (expired)", 1001, stale, 1001, false},
		{"mismatch+stale → stale", 1002, stale, 1001, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			rec := &contextUsageRecord{SessionID: "", WriterPID: tt.recPID, CapturedAt: tt.recCapture}
			if got := isFreshForSession(rec, "", tt.curPID); got != tt.want {
				t.Errorf("isFreshForSession(empty, pid=%d) = %v, want %v", tt.curPID, got, tt.want)
			}
		})
	}

	// UUID path stays independent of writer_pid even when writer_pid differs.
	uuid := &contextUsageRecord{SessionID: "sess-U", WriterPID: 1002, CapturedAt: now}
	if !isFreshForSession(uuid, "sess-U", 1001) {
		t.Errorf("UUID match must remain valid regardless of writer_pid (AC-013 preserved)")
	}
}

// TestIsFreshForSession_NilRecord — nil record is never fresh.
func TestIsFreshForSession_NilRecord(t *testing.T) {
	t.Parallel()
	if isFreshForSession(nil, "", 1) {
		t.Errorf("nil record must be stale")
	}
}

// TestContextUsageFresh_Unparseable — a malformed captured_at is not fresh
// (conservative → heuristics fallback).
func TestContextUsageFresh_Unparseable(t *testing.T) {
	t.Parallel()
	if contextUsageFresh("not-a-timestamp") {
		t.Errorf("unparseable captured_at must be treated as not fresh")
	}
	if !contextUsageFresh(time.Now().Format(time.RFC3339Nano)) {
		t.Errorf("a just-now timestamp must be fresh")
	}
}

// TestReadContextUsage_Corrupt — corrupt JSON surfaces an error so the caller
// falls back (throttle → write, reader → heuristics).
func TestReadContextUsage_Corrupt(t *testing.T) {
	t.Parallel()
	proj := t.TempDir()
	stateDir := filepath.Join(proj, ".moai", "state")
	if err := os.MkdirAll(stateDir, 0o755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(stateDir, "context-usage.json")
	if err := os.WriteFile(path, []byte("{not json"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := readContextUsage(path); err == nil {
		t.Errorf("corrupt JSON must return an error")
	}
	// A corrupt on-disk record must NOT block a subsequent write (throttle
	// read fails → write proceeds).
	m := MemoryData{ContextWindowSize: 256_000, TokensUsed: 230_400, Available: true}
	writeContextUsage(proj, "sess-recover", 9, m, handoffStageSoft)
	rec := readRecord(t, path)
	if rec.SessionID != "sess-recover" {
		t.Errorf("write must overwrite a corrupt record; session_id = %q", rec.SessionID)
	}
}

// TestResolveProjectDir_Chain — the workspace/CWD resolution chain and the
// nil-input CWD fallback (design §D.2).
func TestResolveProjectDir_Chain(t *testing.T) {
	t.Parallel()

	// workspace.current_dir wins.
	if got := resolveProjectDir(&StdinData{Workspace: &WorkspaceInfo{CurrentDir: "/ws/cur"}, CWD: "/legacy"}); got != "/ws/cur" {
		t.Errorf("workspace.current_dir should win, got %q", got)
	}
	// legacy cwd when workspace absent.
	if got := resolveProjectDir(&StdinData{CWD: "/legacy"}); got != "/legacy" {
		t.Errorf("legacy cwd fallback, got %q", got)
	}
	// nil input → os.Getwd() (non-empty in a normal test env).
	if got := resolveProjectDir(nil); got == "" {
		t.Errorf("nil input should fall back to os.Getwd(), got empty")
	}
}
