package cli

// SPEC-CLIFIX-CRITICAL-001 M1 — Reproduction tests (RED phase).
//
// Each Test*_Repro function reproduces one of the 8 Critical defects from the
// CLI audit. They are written BEFORE the corresponding fix (Reproduction-First,
// REQ-CRIT-001-009) and are intended to FAIL on the unmodified codebase.
//
// Canonical test-name tokens (acceptance.md §D): SettingsLocalPreserve,
// ClaimTaskAppend, HarnessMutePreserve, UpdateLock, MigrateRollbackPreexisting,
// MigrateSymlinkSkip, TierPromotionHighWater. RemoveHarnessBoundary lives in
// internal/cli/harness/v4lifecycle_boundary_repro_test.go (package harness).

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// --- helpers ---

func mustWriteFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdirall %s: %v", path, err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func readFileStr(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(b)
}

// =========================================================================
// Defect a — closed SettingsLocal struct RMW wipes non-struct top-level keys.
// AC-CRIT-001-001 / AC-CRIT-001-001b
// =========================================================================

func TestSettingsLocalPreserve_Repro(t *testing.T) {
	tmp := t.TempDir()
	settingsPath := filepath.Join(tmp, "settings.local.json")

	// Fixture carries top-level keys beyond the closed SettingsLocal struct
	// (hooks, outputStyle, model) plus modeled keys (env, teammateMode).
	original := `{
  "hooks": {
    "Stop": [{"hooks": [{"command": "echo done"}]}]
  },
  "outputStyle": "moai",
  "model": "claude-opus-4-8",
  "env": {"FOO": "bar"},
  "teammateMode": "auto"
}`
	mustWriteFile(t, settingsPath, original)

	if err := ensureSettingsLocalJSON(settingsPath); err != nil {
		t.Fatalf("ensureSettingsLocalJSON: %v", err)
	}

	after := readFileStr(t, settingsPath)
	var m map[string]any
	if err := json.Unmarshal([]byte(after), &m); err != nil {
		t.Fatalf("parse after: %v", err)
	}

	// Every pre-existing top-level key must survive the write-back.
	for _, key := range []string{"hooks", "outputStyle", "model", "env"} {
		if _, ok := m[key]; !ok {
			t.Errorf("top-level key %q was wiped by the RMW round-trip (defect a)", key)
		}
	}
	if tm, _ := m["teammateMode"].(string); tm != "tmux" {
		t.Errorf("teammateMode = %q, want %q", tm, "tmux")
	}
}

// =========================================================================
// Defect b — ClaimTask opens O_RDWR and writes at offset 0, overwriting the
// ledger head. AC-CRIT-001-002 / AC-CRIT-001-002b
// =========================================================================

func TestClaimTaskAppend_Repro(t *testing.T) {
	tmp := t.TempDir()
	teamID := "team-repro"
	teamDir := filepath.Join(tmp, "team", teamID)
	mustWriteFile(t, filepath.Join(teamDir, "tasklist.md"), "") // mkdir via helper
	tasklistPath := filepath.Join(teamDir, "tasklist.md")

	header := "# Team Tasklist\n\n"
	taskLine := "- **TASK-001** implement X - Status: pending\n"
	original := header + taskLine
	mustWriteFile(t, tasklistPath, original)

	if err := ClaimTask(tmp, teamID, "teammate-1", "TASK-001"); err != nil {
		t.Fatalf("ClaimTask: %v", err)
	}

	after := readFileStr(t, tasklistPath)

	// The original content MUST be a prefix of the result (append-only).
	if !strings.HasPrefix(after, original) {
		t.Errorf("ledger head was overwritten (defect b):\nwant prefix=%q\ngot           =%q", original, after)
	}

	// The CLAIMED line must land at the tail.
	if !strings.Contains(after, "CLAIMED") {
		t.Errorf("claim line missing from ledger tail: %q", after)
	}
}

// =========================================================================
// Defect c — saveWorkflowMuteConfig minimal-struct YAML round-trip wipes
// workflow.yaml sibling keys. AC-CRIT-001-003
// =========================================================================

func TestHarnessMutePreserve_Repro(t *testing.T) {
	tmp := t.TempDir()
	wfPath := filepath.Join(tmp, "workflow.yaml")
	original := `agentic_loop:
  enabled: true
team:
  enabled: false
harness:
  proposal:
    mode: auto
    mute:
      categories: []
`
	mustWriteFile(t, wfPath, original)

	cfg, err := loadWorkflowMuteConfig(wfPath)
	if err != nil {
		t.Fatalf("loadWorkflowMuteConfig: %v", err)
	}
	cfg.Harness.Proposal.Mute.Categories = append(cfg.Harness.Proposal.Mute.Categories, "naming")

	if err := saveWorkflowMuteConfig(wfPath, cfg); err != nil {
		t.Fatalf("saveWorkflowMuteConfig: %v", err)
	}

	after := readFileStr(t, wfPath)
	// Sibling keys must survive the round-trip.
	for _, key := range []string{"agentic_loop", "team", "naming"} {
		if !strings.Contains(after, key) {
			t.Errorf("sibling key %q was wiped by saveWorkflowMuteConfig (defect c)", key)
		}
	}
}

// =========================================================================
// Defect e — acquireUpdateLock has zero production callers; moai update runs
// lockless. AC-CRIT-001-005
// =========================================================================

func TestUpdateLock_ContendedFailsFast_Repro(t *testing.T) {
	tmp := t.TempDir()
	if err := os.MkdirAll(filepath.Join(tmp, ".moai"), 0o755); err != nil {
		t.Fatalf("mkdir .moai: %v", err)
	}

	// First acquisition succeeds.
	release, err := acquireUpdateLock(tmp)
	if err != nil {
		t.Fatalf("first acquireUpdateLock: %v", err)
	}
	defer release()

	// Second concurrent acquisition must fail fast with ErrUpdateLockHeld.
	_, err2 := acquireUpdateLock(tmp)
	if !errors.Is(err2, ErrUpdateLockHeld) {
		t.Errorf("second acquireUpdateLock err = %v, want ErrUpdateLockHeld (defect e primitive)", err2)
	}
}

// =========================================================================
// Defect f — rollback unconditionally removes recorded paths including
// pre-existing dirs. AC-CRIT-001-006
// =========================================================================

func TestMigrateRollbackPreexisting_Repro(t *testing.T) {
	tmp := t.TempDir()
	src := filepath.Join(tmp, "src")
	mustWriteFile(t, filepath.Join(src, "a.txt"), "from-src")

	// dst pre-exists with precious user content.
	dst := filepath.Join(tmp, "dst")
	if err := os.MkdirAll(dst, 0o755); err != nil {
		t.Fatalf("mkdir dst: %v", err)
	}
	mustWriteFile(t, filepath.Join(dst, "precious.txt"), "do-not-delete")

	r := &migrateAgencyRunner{}
	tx := &transactionLog{}
	if err := r.copyDir(src, dst, tx); err != nil {
		t.Fatalf("copyDir: %v", err)
	}

	// Simulate a phase failure → rollback.
	tx.rollback()

	// The pre-existing dir + its original content must survive rollback.
	if _, err := os.Stat(filepath.Join(dst, "precious.txt")); err != nil {
		t.Errorf("pre-existing precious.txt did not survive rollback (defect f): %v", err)
	}
}

// =========================================================================
// Defect g — os.Stat follows symlinks → ModeSymlink guard is dead code; symlink
// targets outside the tree get copied. AC-CRIT-001-007
// =========================================================================

func TestMigrateSymlinkSkip_Repro(t *testing.T) {
	tmp := t.TempDir()
	// Out-of-tree target with secret content.
	outside := filepath.Join(tmp, "outside")
	if err := os.MkdirAll(outside, 0o755); err != nil {
		t.Fatalf("mkdir outside: %v", err)
	}
	secretPath := filepath.Join(outside, "secret.txt")
	mustWriteFile(t, secretPath, "OUT-OF-TREE-SECRET")

	// Symlink inside src pointing outside the tree.
	src := filepath.Join(tmp, "src")
	if err := os.MkdirAll(src, 0o755); err != nil {
		t.Fatalf("mkdir src: %v", err)
	}
	linkPath := filepath.Join(src, "link.txt")
	if err := os.Symlink(secretPath, linkPath); err != nil {
		t.Fatalf("symlink: %v", err)
	}

	dst := filepath.Join(tmp, "dst", "link.txt")
	r := &migrateAgencyRunner{}
	tx := &transactionLog{}
	if err := r.copyFile(linkPath, dst, tx); err != nil {
		t.Fatalf("copyFile: %v", err)
	}

	// The symlink must be skipped — no out-of-tree content copied into dst.
	got, err := os.ReadFile(dst)
	if err == nil {
		if strings.Contains(string(got), "OUT-OF-TREE-SECRET") {
			t.Errorf("out-of-tree symlink target was copied into dst (defect g): %q", string(got))
		}
	}
}

// =========================================================================
// Defect h — Stop-hook auto-classify re-appends promotions each session,
// unbounded tier-promotions.jsonl. AC-CRIT-001-008
// =========================================================================

func TestTierPromotionHighWater_Repro(t *testing.T) {
	tmp := t.TempDir()
	logDir := filepath.Join(tmp, ".moai", "harness")
	histDir := filepath.Join(logDir, "learning-history")
	if err := os.MkdirAll(histDir, 0o755); err != nil {
		t.Fatalf("mkdir hist: %v", err)
	}

	// usage-log.jsonl with one eligible pattern at a stable count.
	usageLog := filepath.Join(logDir, "usage-log.jsonl")
	var lines []string
	for i := 0; i < 5; i++ {
		entry := `{"timestamp":"2026-07-10T00:00:00Z","event_type":"tool_use","subject":"/moai plan","context_hash":"ctx1"}`
		lines = append(lines, entry)
	}
	mustWriteFile(t, usageLog, strings.Join(lines, "\n")+"\n")

	// First classify call establishes the baseline promotion.
	if _, _, err := classifyHarnessPatterns(tmp); err != nil {
		t.Fatalf("first classifyHarnessPatterns: %v", err)
	}
	promoPath := filepath.Join(histDir, "tier-promotions.jsonl")
	before := countLines(t, promoPath)
	if before == 0 {
		t.Fatalf("baseline: expected >=1 promotion line, got 0 (setup issue)")
	}

	// Second classify call over identical logs — high-water mark must suppress.
	if _, _, err := classifyHarnessPatterns(tmp); err != nil {
		t.Fatalf("second classifyHarnessPatterns: %v", err)
	}
	after := countLines(t, promoPath)

	if after != before {
		t.Errorf("duplicate promotion appended (defect h): before=%d after=%d (expected no growth)", before, after)
	}
}

func countLines(t *testing.T, path string) int {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		return 0
	}
	s := strings.TrimSpace(string(b))
	if s == "" {
		return 0
	}
	return len(strings.Split(s, "\n"))
}
