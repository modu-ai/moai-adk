package hook

// subagent_write_guard_newfile_test.go — card t1499 M3.
//
// A stat failure on the target was always logged as
// `fail-open reason="existing file unreadable"`, which also covers a brand-new
// file (ENOENT): 6,122 of 6,947 audit rows were new files, drowning the real
// fail-opens. A missing file cannot be destructively shrunk, so it is an
// allow with its own reason; any other stat error stays fail-open.

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestSubagentWriteGuardNewFileIsAllowNotFailOpen(t *testing.T) {
	repo, _ := swgSetupTrackedRepo(t, 20000)
	missing := filepath.Join(repo, "brand-new.txt")

	input := swgPayload("agent-abc123", "general-purpose", missing, strings.Repeat("x", 100))
	input.CWD = repo

	ev := evaluateSubagentWrite(input, true)
	if ev.Decision != swgDecisionAllow {
		t.Fatalf("decision = %q, want allow for a file that does not exist (reason %q)", ev.Decision, ev.Reason)
	}
	if ev.Reason != "new file" {
		t.Errorf("reason = %q, want %q", ev.Reason, "new file")
	}
	if strings.Contains(ev.Reason, "unreadable") {
		t.Errorf("a missing file must not be reported as unreadable: %q", ev.Reason)
	}
}

func TestSubagentWriteGuardNewFileAuditRowIsAllowWithReason(t *testing.T) {
	repo, _ := swgSetupTrackedRepo(t, 20000)
	projectDir := t.TempDir()
	h := swgHandlerWithConfig(true, projectDir)

	input := swgPayload("agent-abc123", "general-purpose", filepath.Join(repo, "brand-new.txt"), "hello")
	input.CWD = repo
	if reason := h.checkSubagentDestructiveWrite(input); reason != "" {
		t.Fatalf("want allow, got deny %q", reason)
	}
	rows := swgReadAuditRows(t, projectDir)
	if len(rows) != 1 {
		t.Fatalf("rows = %v, want exactly 1", rows)
	}
	if !strings.Contains(rows[0], "decision=allow") || !strings.Contains(rows[0], `reason="new file"`) {
		t.Errorf("row = %q, want decision=allow and reason=\"new file\"", rows[0])
	}
	if strings.Contains(rows[0], "decision=fail-open") {
		t.Errorf("a new file must not be recorded as fail-open: %q", rows[0])
	}
}

// A relative path to a file that does not exist resolves against the project
// root first and then hits the same ENOENT branch.
func TestSubagentWriteGuardNewRelativeFileIsAllow(t *testing.T) {
	repo, _ := swgSetupTrackedRepo(t, 20000)
	input := swgPayload("agent-abc123", "general-purpose", "sub/brand-new.txt", "hello")
	input.CWD = repo
	t.Setenv("CLAUDE_PROJECT_DIR", repo)

	ev := evaluateSubagentWrite(input, true)
	if ev.Decision != swgDecisionAllow || ev.Reason != "new file" {
		t.Fatalf("decision=%q reason=%q, want allow / new file", ev.Decision, ev.Reason)
	}
}

// Any stat error other than "does not exist" stays fail-open, and its reason
// says the stat failed rather than claiming the file is missing.
func TestSubagentWriteGuardStatErrorOtherThanMissingStaysFailOpen(t *testing.T) {
	// chmod 0o000 on a directory does not block access on Windows, and
	// Geteuid() is -1 there, so the root check below cannot catch it.
	if runtime.GOOS == "windows" {
		t.Skip("directory permission bits do not deny stat on Windows")
	}
	if os.Geteuid() == 0 {
		t.Skip("root bypasses directory permissions")
	}
	base := t.TempDir()
	locked := filepath.Join(base, "locked")
	if err := os.Mkdir(locked, 0o755); err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(locked, "file.txt")
	if err := os.WriteFile(target, []byte(strings.Repeat("a", 20000)), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(locked, 0o000); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(locked, 0o755) })

	input := swgPayload("agent-abc123", "general-purpose", target, strings.Repeat("x", 100))
	input.CWD = base

	ev := evaluateSubagentWrite(input, true)
	if ev.Decision != swgDecisionFailOpen {
		t.Fatalf("decision = %q, want fail-open for a permission-denied stat (reason %q)", ev.Decision, ev.Reason)
	}
	if !strings.Contains(ev.Reason, "stat failed") {
		t.Errorf("reason = %q, want it to say the stat failed", ev.Reason)
	}
	if ev.Reason == "new file" {
		t.Errorf("a permission error must not read as a new file")
	}
}

// The repair must not weaken the guard: a real shrink of an existing tracked
// file is still denied.
func TestSubagentWriteGuardRealShrinkStillDeniedAfterNewFileRepair(t *testing.T) {
	repo, filePath := swgSetupTrackedRepo(t, 20000)
	input := swgPayload("agent-abc123", "general-purpose", filePath, strings.Repeat("x", 300))
	input.CWD = repo

	ev := evaluateSubagentWrite(input, true)
	if ev.Decision != swgDecisionDeny {
		t.Fatalf("decision = %q, want deny for a real shrink (reason %q)", ev.Decision, ev.Reason)
	}
}
