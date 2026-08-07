// Package template — Navigator SessionStart hook tests (TDD driver).
//
// Exercises handle-session-start-navigator.sh (the ambient auto-brief hook):
// AC-PN-009 (bounded additionalContext ≤500 tokens, frontier+next-task+link),
// AC-PN-010 (fail-open on missing file / timeout / missing script),
// AC-PN-012 (staleness advisory when >3 sync cycles behind HEAD).
package template

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

const navigatorHookScript = "templates/.claude/hooks/moai/handle-session-start-navigator.sh"

// runHook invokes the SessionStart Navigator hook with CLAUDE_PROJECT_DIR=dir
// and returns (stdout, exitCode).
func runHook(t *testing.T, dir string) (stdout string, exitCode int) {
	t.Helper()
	scriptAbs, err := filepath.Abs(navigatorHookScript)
	if err != nil {
		t.Fatalf("resolve hook abs: %v", err)
	}
	if _, err := os.Stat(scriptAbs); err != nil {
		t.Fatalf("hook script not found at %s — has M3 landed it?", scriptAbs)
	}
	c := exec.Command("bash", scriptAbs)
	c.Dir = dir
	c.Env = append(os.Environ(), "CLAUDE_PROJECT_DIR="+dir)
	var out, errBuf bytes.Buffer
	c.Stdout = &out
	c.Stderr = &errBuf
	err = c.Run()
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			return out.String(), ee.ExitCode()
		}
		t.Fatalf("invoke hook: %v", err)
	}
	return out.String(), 0
}

// parseAdditionalContext extracts hookSpecificOutput.additionalContext from
// stdout JSON. Returns ("", false) if stdout is empty/not JSON.
func parseAdditionalContext(t *testing.T, stdout string) (string, bool) {
	t.Helper()
	stdout = strings.TrimSpace(stdout)
	if stdout == "" {
		return "", false
	}
	var doc struct {
		HookSpecificOutput struct {
			HookEventName     string `json:"hookEventName"`
			AdditionalContext string `json:"additionalContext"`
		} `json:"hookSpecificOutput"`
	}
	if err := json.Unmarshal([]byte(stdout), &doc); err != nil {
		return "", false
	}
	return doc.HookSpecificOutput.AdditionalContext, true
}

// approxTokens returns a rough token count (~4 chars/token heuristic).
func approxTokens(s string) int { return len(s) / 4 }

// TestACPN009_AmbientBriefBounded verifies AC-PN-009: the hook emits an
// additionalContext with frontier + next-task + link, ≤500 tokens.
func TestACPN009_AmbientBriefBounded(t *testing.T) {
	dir := t.TempDir()
	initFixtureRepo(t, dir)
	writeSPEC(t, dir, "SPEC-ACTIVE-001", "Active", "in-progress", "v1.0.0", "internal/active")
	if err := gitRun(dir, "add", ".moai/specs"); err != nil {
		t.Fatal(err)
	}
	if err := gitRun(dir, "commit", "-m", "active"); err != nil {
		t.Fatal(err)
	}
	// Regenerate the Navigator so the hook has something to read.
	if _, err := runRegen(t, dir); err != nil {
		t.Fatal(err)
	}

	stdout, code := runHook(t, dir)
	if code != 0 {
		t.Fatalf("AC-PN-009: hook exited %d (must be 0); stdout=%s", code, stdout)
	}
	ac, ok := parseAdditionalContext(t, stdout)
	if !ok || ac == "" {
		t.Fatalf("AC-PN-009: hook emitted no additionalContext; stdout=%s", stdout)
	}
	if !strings.Contains(ac, "SPEC-ACTIVE-001") {
		t.Errorf("AC-PN-009: additionalContext missing the frontier SPEC reference:\n%s", ac)
	}
	if !strings.Contains(ac, "navigator.md") {
		t.Errorf("AC-PN-009: additionalContext missing link to navigator.md:\n%s", ac)
	}
	if tok := approxTokens(ac); tok > 500 {
		t.Errorf("AC-PN-009: additionalContext exceeds 500-token ceiling (~%d tokens):\n%s", tok, ac)
	}
}

// TestACPN010_FailOpen verifies AC-PN-010: the hook exits 0 and emits no
// additionalContext (or an empty one) when navigator.md is absent.
func TestACPN010_FailOpenMissingFile(t *testing.T) {
	dir := t.TempDir()
	initFixtureRepo(t, dir)
	// No Navigator regeneration — navigator.md does not exist.
	stdout, code := runHook(t, dir)
	if code != 0 {
		t.Fatalf("AC-PN-010: hook exited %d on missing navigator.md (must be 0)", code)
	}
	ac, ok := parseAdditionalContext(t, stdout)
	if ok && ac != "" {
		t.Errorf("AC-PN-010: hook emitted non-empty additionalContext on missing file: %q", ac)
	}
}

// TestACPN010_FailOpenEmptyProject verifies AC-PN-010 on an empty Navigator
// (placeholder form) — the hook still exits 0 and emits nothing harmful.
func TestACPN010_FailOpenEmptyProject(t *testing.T) {
	dir := t.TempDir()
	initFixtureRepo(t, dir)
	// Regenerate with zero SPECs → placeholder navigator.md.
	if _, err := runRegen(t, dir); err != nil {
		t.Fatal(err)
	}
	stdout, code := runHook(t, dir)
	if code != 0 {
		t.Fatalf("AC-PN-010 (empty): hook exited %d (must be 0)", code)
	}
	// Either no output OR output that is benign (no crash).
	if strings.Contains(stdout, "panic") || strings.Contains(stdout, "error") {
		t.Errorf("AC-PN-010 (empty): hook surfaced an error/panic: %s", stdout)
	}
}

// TestACPN012_StalenessAdvisory verifies AC-PN-012: when the Navigator's
// last-regen-commit is more than 3 sync cycles behind HEAD, the hook's
// additionalContext includes a staleness advisory.
func TestACPN012_StalenessAdvisory(t *testing.T) {
	dir := t.TempDir()
	initFixtureRepo(t, dir)
	writeSPEC(t, dir, "SPEC-S-001", "S", "draft", "v1.0.0", "internal/s")
	if err := gitRun(dir, "add", ".moai/specs"); err != nil {
		t.Fatal(err)
	}
	if err := gitRun(dir, "commit", "-m", "s"); err != nil {
		t.Fatal(err)
	}
	// Regenerate so last-regen-commit.txt = HEAD (call it C0).
	if _, err := runRegen(t, dir); err != nil {
		t.Fatal(err)
	}
	// Advance HEAD by 4 commits (each with a distinct file) so the Navigator
	// is >3 sync cycles behind.
	for i := 0; i < 4; i++ {
		if err := os.WriteFile(filepath.Join(dir, "advance"+string(rune('0'+i))+".txt"),
			[]byte("x\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := gitRun(dir, "add", "advance"+string(rune('0'+i))+".txt"); err != nil {
			t.Fatal(err)
		}
		if err := gitRun(dir, "commit", "-m", "advance "+string(rune('0'+i))); err != nil {
			t.Fatal(err)
		}
	}

	stdout, code := runHook(t, dir)
	if code != 0 {
		t.Fatalf("AC-PN-012: hook exited %d", code)
	}
	ac, ok := parseAdditionalContext(t, stdout)
	if !ok {
		t.Fatalf("AC-PN-012: no additionalContext; stdout=%s", stdout)
	}
	if !strings.Contains(strings.ToLower(ac), "stale") {
		t.Errorf("AC-PN-012: additionalContext missing staleness advisory (>3 commits behind HEAD):\n%s", ac)
	}
}

// TestACPN012_StalenessOverride verifies the navigator.staleness_cycles
// override: writing the config key at .moai/config/sections/navigator.yaml
// with staleness_cycles: 1 causes the advisory to fire at 2 commits behind.
func TestACPN012_StalenessOverride(t *testing.T) {
	dir := t.TempDir()
	initFixtureRepo(t, dir)
	writeSPEC(t, dir, "SPEC-T-001", "T", "draft", "v1.0.0", "internal/t")
	if err := gitRun(dir, "add", ".moai/specs"); err != nil {
		t.Fatal(err)
	}
	if err := gitRun(dir, "commit", "-m", "t"); err != nil {
		t.Fatal(err)
	}
	if _, err := runRegen(t, dir); err != nil {
		t.Fatal(err)
	}
	// Override: staleness_cycles: 1 (advisory fires at 2+ commits behind).
	cfgDir := filepath.Join(dir, ".moai", "config", "sections")
	if err := os.MkdirAll(cfgDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(cfgDir, "navigator.yaml"),
		[]byte("navigator:\n  staleness_cycles: 1\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	// Advance by 2 commits — under default N=3 but over the override N=1.
	for i := 0; i < 2; i++ {
		if err := os.WriteFile(filepath.Join(dir, "ov"+string(rune('0'+i))+".txt"),
			[]byte("x\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := gitRun(dir, "add", "ov"+string(rune('0'+i))+".txt"); err != nil {
			t.Fatal(err)
		}
		if err := gitRun(dir, "commit", "-m", "ov"+string(rune('0'+i))); err != nil {
			t.Fatal(err)
		}
	}
	stdout, _ := runHook(t, dir)
	ac, _ := parseAdditionalContext(t, stdout)
	if !strings.Contains(strings.ToLower(ac), "stale") {
		t.Errorf("AC-PN-012 override: expected staleness advisory at N=1 with 2 commits behind, got:\n%s", ac)
	}
}
