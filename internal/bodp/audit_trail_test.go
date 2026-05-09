package bodp

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// sampleEntry returns a fully-populated AuditEntry for fixture reuse.
func sampleEntry() AuditEntry {
	return AuditEntry{
		Timestamp:     time.Date(2026, 5, 9, 12, 0, 0, 0, time.UTC),
		EntryPoint:    EntryPlanBranch,
		CurrentBranch: "chore/translation-batch-b",
		NewBranch:     "feat/SPEC-Y-001",
		Decision: BODPDecision{
			SignalA:     false,
			SignalB:     false,
			SignalC:     false,
			Recommended: ChoiceMain,
			Rationale:   "현재 브랜치와 무관한 새 작업이므로 main 분기를 권장합니다.",
			BaseBranch:  "origin/main",
		},
		UserChoice:  ChoiceMain,
		ExecutedCmd: "git fetch origin main && git checkout -B feat/SPEC-Y-001 origin/main",
	}
}

// TestWriteDecision_CreatesFile: WriteDecision creates the expected file with
// frontmatter + body sections.
func TestWriteDecision_CreatesFile(t *testing.T) {
	repoRoot := t.TempDir()
	if err := os.MkdirAll(filepath.Join(repoRoot, auditTrailDir), 0o755); err != nil {
		t.Fatalf("pre-create dir: %v", err)
	}

	entry := sampleEntry()
	if err := WriteDecision(repoRoot, entry); err != nil {
		t.Fatalf("WriteDecision: %v", err)
	}

	expected := filepath.Join(repoRoot, auditTrailDir, "feat-SPEC-Y-001.md")
	content, err := os.ReadFile(expected)
	if err != nil {
		t.Fatalf("read expected file: %v", err)
	}
	body := string(content)

	frontmatterChecks := []string{
		"timestamp: 2026-05-09",
		"entry_point: plan-branch",
		"current_branch: chore/translation-batch-b",
		"new_branch: feat/SPEC-Y-001",
		"user_choice: main",
	}
	for _, want := range frontmatterChecks {
		if !strings.Contains(body, want) {
			t.Errorf("expected frontmatter to contain %q, got:\n%s", want, body)
		}
	}

	bodyChecks := []string{
		"## Signals",
		"## Decision",
		"## Executed",
		entry.ExecutedCmd,
		"Signal (a)",
		"Signal (b)",
		"Signal (c)",
		entry.Decision.Rationale,
	}
	for _, want := range bodyChecks {
		if !strings.Contains(body, want) {
			t.Errorf("expected body to contain %q, got:\n%s", want, body)
		}
	}
}

// TestWriteDecision_NormalizesBranchNameForFilename: slash → dash conversion
// produces a filesystem-safe filename.
func TestWriteDecision_NormalizesBranchNameForFilename(t *testing.T) {
	repoRoot := t.TempDir()
	if err := os.MkdirAll(filepath.Join(repoRoot, auditTrailDir), 0o755); err != nil {
		t.Fatalf("pre-create dir: %v", err)
	}
	entry := sampleEntry()
	entry.NewBranch = "feat/SPEC-A/sub-feature"

	if err := WriteDecision(repoRoot, entry); err != nil {
		t.Fatalf("WriteDecision: %v", err)
	}

	expected := filepath.Join(repoRoot, auditTrailDir, "feat-SPEC-A-sub-feature.md")
	if _, err := os.Stat(expected); err != nil {
		t.Errorf("expected normalized filename %q, stat err: %v", expected, err)
	}
}

// TestWriteDecision_CreatesDirectoryIfAbsent: parent dir absent → MkdirAll
// invoked transparently.
func TestWriteDecision_CreatesDirectoryIfAbsent(t *testing.T) {
	repoRoot := t.TempDir()
	// Note: deliberately not pre-creating auditTrailDir.

	entry := sampleEntry()
	if err := WriteDecision(repoRoot, entry); err != nil {
		t.Fatalf("WriteDecision: %v", err)
	}

	expected := filepath.Join(repoRoot, auditTrailDir, "feat-SPEC-Y-001.md")
	if _, err := os.Stat(expected); err != nil {
		t.Errorf("expected file %q to exist after auto-mkdir, stat err: %v", expected, err)
	}
}

// TestHasAuditTrail_DetectsExisting: pre-existing markdown → true.
func TestHasAuditTrail_DetectsExisting(t *testing.T) {
	repoRoot := t.TempDir()
	dir := filepath.Join(repoRoot, auditTrailDir)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	target := filepath.Join(dir, "feat-X.md")
	if err := os.WriteFile(target, []byte("# stub\n"), 0o644); err != nil {
		t.Fatalf("seed audit trail: %v", err)
	}

	if !HasAuditTrail(repoRoot, "feat/X") {
		t.Errorf("HasAuditTrail returned false for existing file %q", target)
	}
}

// TestHasAuditTrail_AbsentBranch: dir present, file absent → false.
func TestHasAuditTrail_AbsentBranch(t *testing.T) {
	repoRoot := t.TempDir()
	if err := os.MkdirAll(filepath.Join(repoRoot, auditTrailDir), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if HasAuditTrail(repoRoot, "feat/missing") {
		t.Errorf("HasAuditTrail returned true for absent file")
	}
}

// TestHasAuditTrail_DirAbsentReturnsFalse: dir absent → false (and no error).
// W7-T05 reminder relies on this behaviour to avoid false-positive on fresh
// projects (acceptance.md AC-CIAUT-024 Failure Mode).
func TestHasAuditTrail_DirAbsentReturnsFalse(t *testing.T) {
	repoRoot := t.TempDir()
	if HasAuditTrail(repoRoot, "feat/anything") {
		t.Errorf("HasAuditTrail returned true for absent directory")
	}
}
