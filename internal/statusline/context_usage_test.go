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
func readRecord(t *testing.T, path string) *SessionTelemetryRecord {
	t.Helper()
	rec, err := ReadSessionTelemetry(path)
	if err != nil {
		t.Fatalf("ReadSessionTelemetry(%q) error: %v", path, err)
	}
	return rec
}

// usagePath is the on-disk record path for one session under projDir. It
// mirrors what the writer does rather than re-spelling the layout.
func usagePath(projDir, sessionID string) string {
	return SessionTelemetryPath(filepath.Join(projDir, ".moai", "state"), sessionID)
}

// TestWriteContextUsage_ConfigIndependent — AC-THRESHOLD-007.
// The write happens regardless of any handoff config (there is no config param);
// it depends only on Memory availability.
func TestWriteContextUsage_ConfigIndependent(t *testing.T) {
	t.Parallel()

	proj := t.TempDir()
	m := MemoryData{ContextWindowSize: 256_000, TokensUsed: 230_400, Available: true}
	writeContextUsage(proj, "sess-1", 4242, m, handoffStageSoft)

	rec := readRecord(t, usagePath(proj, "sess-1"))
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

	path := usagePath(proj, "sess-atomic")
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("session telemetry record not created: %v", err)
	}
	// Valid JSON (ReadSessionTelemetry parses it).
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
	if _, err := os.Stat(usagePath(proj, "s")); !os.IsNotExist(err) {
		t.Errorf("no file expected when Memory.Available == false")
	}

	// (c) non-positive window → skip (avoids div-by-zero).
	proj2 := t.TempDir()
	writeContextUsage(proj2, "s", 1, MemoryData{ContextWindowSize: 0, Available: true}, handoffStageNone)
	if _, err := os.Stat(usagePath(proj2, "s")); !os.IsNotExist(err) {
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

	raw, err := os.ReadFile(usagePath(proj, "sess-schema"))
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

	rec := readRecord(t, usagePath(proj, "sess-schema"))
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
	path := usagePath(proj, "sess-throttle")
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

// TestReadContextUsage_Corrupt — corrupt JSON surfaces an error so the caller
// falls back (throttle → write, reader → heuristics).
func TestReadContextUsage_Corrupt(t *testing.T) {
	t.Parallel()
	proj := t.TempDir()
	stateDir := filepath.Join(proj, ".moai", "state")
	if err := os.MkdirAll(stateDir, 0o755); err != nil {
		t.Fatal(err)
	}
	path := usagePath(proj, "sess-recover")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("{not json"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := ReadSessionTelemetry(path); err == nil {
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

// TestWriteContextUsage_TemplateSourceGuard — the write MUST NOT create
// context-usage.json when projDir resolves into the moai-adk-go template embed
// source tree (internal/template/templates). The //go:embed all:templates
// directive in internal/template/embed.go includes dot-prefixed dirs (.moai/),
// so a write here would leak a runtime artifact into the distributed binary.
func TestWriteContextUsage_TemplateSourceGuard(t *testing.T) {
	t.Parallel()

	m := MemoryData{ContextWindowSize: 256_000, TokensUsed: 230_400, Available: true}

	// (a) projDir IS the templates dir.
	root := t.TempDir()
	templatesDir := filepath.Join(root, "internal", "template", "templates")
	if err := os.MkdirAll(templatesDir, 0o755); err != nil {
		t.Fatal(err)
	}
	writeContextUsage(templatesDir, "sess-guard", 1, m, handoffStageSoft)
	if _, err := os.Stat(usagePath(templatesDir, "sess-guard")); !os.IsNotExist(err) {
		t.Errorf("(a) no file expected inside templates source dir, got: %v", err)
	}

	// (b) projDir is a deep subdir of templates (mimics a hook path).
	deep := filepath.Join(templatesDir, ".claude", "hooks", "moai")
	if err := os.MkdirAll(deep, 0o755); err != nil {
		t.Fatal(err)
	}
	writeContextUsage(deep, "sess-guard", 1, m, handoffStageSoft)
	if _, err := os.Stat(usagePath(deep, "sess-guard")); !os.IsNotExist(err) {
		t.Errorf("(b) no file expected inside templates source subdir, got: %v", err)
	}
}

// TestWriteContextUsage_TemplateSourceGuard_InertForNormalDir — the guard must
// NOT trigger for a normal project dir; the existing write behavior is preserved.
func TestWriteContextUsage_TemplateSourceGuard_InertForNormalDir(t *testing.T) {
	t.Parallel()

	proj := t.TempDir()
	m := MemoryData{ContextWindowSize: 256_000, TokensUsed: 230_400, Available: true}
	writeContextUsage(proj, "sess-normal", 1, m, handoffStageSoft)
	if _, err := os.Stat(usagePath(proj, "sess-normal")); err != nil {
		t.Fatalf("file expected for normal dir (guard must be inert), got: %v", err)
	}
}

// TestWriteContextUsage_TemplateSourceGuard_SubstringNoMatch — a dir whose name
// merely CONTAINS the segment characters but is NOT a real
// internal/template/templates path component must still get the write (guard
// does not falsely trigger on substring matches).
func TestWriteContextUsage_TemplateSourceGuard_SubstringNoMatch(t *testing.T) {
	t.Parallel()

	m := MemoryData{ContextWindowSize: 256_000, TokensUsed: 230_400, Available: true}
	root := t.TempDir()

	// (a) segments fused with X — not a path component.
	fused := filepath.Join(root, "internalXtemplateXtemplates")
	if err := os.MkdirAll(fused, 0o755); err != nil {
		t.Fatal(err)
	}
	writeContextUsage(fused, "sess-fused", 1, m, handoffStageSoft)
	if _, err := os.Stat(usagePath(fused, "sess-fused")); err != nil {
		t.Errorf("(a) file expected for substring-only dir, got: %v", err)
	}

	// (b) real component path but with a _bar suffix on the last segment.
	barDir := filepath.Join(root, "internal", "template", "templates_bar")
	if err := os.MkdirAll(barDir, 0o755); err != nil {
		t.Fatal(err)
	}
	writeContextUsage(barDir, "sess-bar", 1, m, handoffStageSoft)
	if _, err := os.Stat(usagePath(barDir, "sess-bar")); err != nil {
		t.Errorf("(b) file expected for _bar suffix dir, got: %v", err)
	}
}

// TestIsTemplateSourceDir — direct unit test for the directory-component
// boundary matcher. Covers the exact-match, subdir, substring-non-match, and
// suffix-non-match cases across absolute and relative inputs.
func TestIsTemplateSourceDir(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		dir  string
		want bool
	}{
		{"empty", "", false},
		{"exact templates dir (absolute)", "/home/user/repo/internal/template/templates", true},
		{"subdir of templates (deep path)", "/home/user/repo/internal/template/templates/.claude/hooks/moai", true},
		{"root-level templates dir", "/internal/template/templates", true},
		{"substring fused no match", "/tmp/internalXtemplateXtemplates", false},
		{"suffix _bar no match", "/tmp/internal/template/templates_bar", false},
		{"normal project dir", "/home/user/myproject", false},
		{"normal dir with internal subdir (not template tree)", "/home/user/myproject/internal/cli", false},
		{"normal dir ending in template (singular)", "/home/user/repo/internal/template", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := isTemplateSourceDir(tt.dir); got != tt.want {
				t.Errorf("isTemplateSourceDir(%q) = %v, want %v", tt.dir, got, tt.want)
			}
		})
	}
}

// TestSessionTelemetryReaderIsExported — SPEC-SESSION-TELEMETRY-001 AC-ST-005.
// The package exports exactly one reader for the session telemetry record, named
// ReadSessionTelemetry and returning the exported SessionTelemetryRecord, so a
// cross-package consumer can name both. A compile error here means the reader is
// still unexported or carries a different identifier than the SPEC pins.
var _ func(string) (*SessionTelemetryRecord, error) = ReadSessionTelemetry
