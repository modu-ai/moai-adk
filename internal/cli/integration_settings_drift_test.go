package cli

// integration_settings_drift_test.go — SPEC-PREMERGE-SETTINGS-DRIFT-001 M4/M5.
//
// AC-PSD-009 through AC-PSD-013 live here: they are assertions about the two
// CLI surfaces (the `acquire` precondition and the `preflight` verb), and none
// of them can be made in internal/kanban because the config-gated refusal and
// the three-state JSON only exist at this layer.
//
// No verdict below is read from a process exit code (AC-PSD-006 keeps that
// honest mechanically).

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/kanban"
	"github.com/spf13/cobra"
)

func driftGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("fixture: git %s in %s: %v\n%s", strings.Join(args, " "), dir, err, out)
	}
}

// newDriftWorktree builds a repository with a committed .claude/settings.json.
// `-b main` is pinned per SPEC-GIT-STATUS-FIXTURE-001: an ambient
// init.defaultBranch makes CI and a local machine disagree for a reason that
// has nothing to do with this card.
func newDriftWorktree(t *testing.T) string {
	t.Helper()
	dir, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatalf("fixture: resolve temp dir: %v", err)
	}
	driftGit(t, dir, "init", "-b", "main")
	driftGit(t, dir, "config", "user.name", "t488 fixture")
	driftGit(t, dir, "config", "user.email", "t488@example.invalid")
	driftGit(t, dir, "config", "commit.gpgsign", "false")
	if err := os.MkdirAll(filepath.Join(dir, ".claude"), 0o755); err != nil {
		t.Fatalf("fixture: mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".claude", "settings.json"), []byte("{\"committed\":true}\n"), 0o644); err != nil {
		t.Fatalf("fixture: write: %v", err)
	}
	driftGit(t, dir, "add", ".claude/settings.json")
	driftGit(t, dir, "commit", "-m", "fixture: baseline")
	return dir
}

// dirtyDriftWorktree is newDriftWorktree plus the modification that makes the
// predicate hit.
func dirtyDriftWorktree(t *testing.T) string {
	t.Helper()
	dir := newDriftWorktree(t)
	if err := os.WriteFile(filepath.Join(dir, ".claude", "settings.json"), []byte("{\"drifted\":true}\n"), 0o644); err != nil {
		t.Fatalf("fixture: dirty: %v", err)
	}
	return dir
}

// writeDriftGateConfig turns the REFUSAL layer on for the project rooted at
// root. The default is false, so a refusal fixture must opt in explicitly —
// which is itself the property AC-PSD-013 pins from the other side.
func writeDriftGateConfig(t *testing.T, root string, enabled bool) {
	t.Helper()
	sections := filepath.Join(root, ".moai", "config", "sections")
	if err := os.MkdirAll(sections, 0o750); err != nil {
		t.Fatalf("fixture: mkdir sections: %v", err)
	}
	value := "false"
	if enabled {
		value = "true"
	}
	yaml := "workflow:\n  settings_drift_gate:\n    enabled: " + value + "\n"
	if err := os.WriteFile(filepath.Join(sections, "workflow.yaml"), []byte(yaml), 0o600); err != nil {
		t.Fatalf("fixture: write workflow.yaml: %v", err)
	}
}

// runIntegrationIn is runIntegration with the process cwd moved into worktree
// for the duration of the call. `acquire` measures the CALLER's tree, so the
// only faithful way to exercise that path is to be standing in it.
func runIntegrationIn(t *testing.T, worktree, root string, args ...string) (string, error) {
	t.Helper()
	prev, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(worktree); err != nil {
		t.Fatalf("chdir %s: %v", worktree, err)
	}
	defer func() { _ = os.Chdir(prev) }()

	t.Setenv("CLAUDE_PROJECT_DIR", root)
	cmd := newIntegrationCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs(args)
	walk(cmd, func(c *cobra.Command) { c.SilenceUsage = true; c.SilenceErrors = true })
	execErr := cmd.Execute()
	return out.String(), execErr
}

func lockPathFor(root string) string {
	return filepath.Join(root, ".moai", "state", kanban.IntegrationLockFileName)
}

func readLockRecord(t *testing.T, root string) kanban.IntegrationLock {
	t.Helper()
	data, err := os.ReadFile(lockPathFor(root))
	if err != nil {
		t.Fatalf("read lock record: %v", err)
	}
	var lock kanban.IntegrationLock
	if err := json.Unmarshal(data, &lock); err != nil {
		t.Fatalf("lock record is not JSON: %v", err)
	}
	return lock
}

func driftLedgerLines(t *testing.T, root string) []string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(kanban.SettingsDriftDir(root), kanban.SettingsDriftLedgerName))
	if err != nil {
		t.Fatalf("read ledger: %v", err)
	}
	var lines []string
	for _, l := range strings.Split(strings.TrimRight(string(data), "\n"), "\n") {
		if strings.TrimSpace(l) != "" {
			lines = append(lines, l)
		}
	}
	return lines
}

func sha256Of(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

// ---------------------------------------------------------------------------
// AC-PSD-009 — the refusal does not take the window.
// ---------------------------------------------------------------------------

// TestAcquireRefusesOnDriftWithGateEnabled — AC-PSD-009.
//
// This is the catcher for the inverted-verdict mutant. A predicate-level
// fixture cannot see that inversion: the predicate's own match count is
// unchanged whichever way the gate compares it. What changes is whether the
// window gets recorded, which is what this test reads.
func TestAcquireRefusesOnDriftWithGateEnabled(t *testing.T) {
	worktree := dirtyDriftWorktree(t)
	root := t.TempDir()
	writeDriftGateConfig(t, root, true)

	out, err := runIntegrationIn(t, worktree, root, "acquire", "--session", "sess-lane8", "--card", "t488")

	// Positive control: the command ran and produced the refusal report. An
	// absence assertion under a command that never ran asserts nothing.
	if err == nil {
		t.Fatalf("acquire succeeded on a drifted tree with the refusal layer on\noutput: %s", out)
	}
	if strings.TrimSpace(out) == "" {
		t.Fatalf("refusal produced no report; the assertions below would be about an empty string")
	}
	source := filepath.Join(worktree, ".claude", "settings.json")
	wantSHA := sha256Of(t, source)
	if !strings.Contains(out, wantSHA) {
		t.Errorf("refusal report does not carry the sha256 %s\noutput: %s", wantSHA, out)
	}
	preserved := preservedCopyPath(t, root)
	if !strings.Contains(out, preserved) {
		t.Errorf("refusal report does not carry the preserved path %s\noutput: %s", preserved, out)
	}

	// The assertion this test exists for.
	if _, statErr := os.Stat(lockPathFor(root)); statErr == nil {
		t.Errorf("the refusal path recorded the window at %s", lockPathFor(root))
	}
}

// TestAcquireRefusalLeavesAnExistingLockByteIdentical — AC-PSD-009, the
// variant where a record already exists. The identity assertion stands on a
// content control: an empty "before" would make it vacuously true.
func TestAcquireRefusalLeavesAnExistingLockByteIdentical(t *testing.T) {
	worktree := newDriftWorktree(t)
	root := t.TempDir()
	writeDriftGateConfig(t, root, true)

	if _, err := runIntegrationIn(t, worktree, root, "acquire", "--session", "sess-lane8"); err != nil {
		t.Fatalf("seed acquire on a clean tree: %v", err)
	}
	before, err := os.ReadFile(lockPathFor(root))
	if err != nil {
		t.Fatalf("read seeded lock: %v", err)
	}
	if len(before) == 0 || !strings.Contains(string(before), "sess-lane8") {
		t.Fatalf("control: the seeded lock is empty or does not name the holder: %q", before)
	}

	// Now dirty the tree and let a DIFFERENT session try to take the window.
	if err := os.WriteFile(filepath.Join(worktree, ".claude", "settings.json"), []byte("{\"drifted\":true}\n"), 0o644); err != nil {
		t.Fatalf("dirty: %v", err)
	}
	if _, err := runIntegrationIn(t, worktree, root, "acquire", "--session", "sess-lane5"); err == nil {
		t.Fatal("a drifted tree took the window with the refusal layer on")
	}

	after, err := os.ReadFile(lockPathFor(root))
	if err != nil {
		t.Fatalf("read lock after refusal: %v", err)
	}
	if string(after) != string(before) {
		t.Errorf("the refusal rewrote the lock record\nbefore: %s\nafter:  %s", before, after)
	}
}

// preservedCopyPath returns the single preserved copy under root, failing when
// there is not exactly one.
func preservedCopyPath(t *testing.T, root string) string {
	t.Helper()
	entries, err := os.ReadDir(kanban.SettingsDriftDir(root))
	if err != nil {
		t.Fatalf("read preserve dir: %v", err)
	}
	var found []string
	for _, e := range entries {
		if e.Name() != kanban.SettingsDriftLedgerName {
			found = append(found, filepath.Join(kanban.SettingsDriftDir(root), e.Name()))
		}
	}
	if len(found) != 1 {
		t.Fatalf("preserved copies: got %d (%v), want 1", len(found), found)
	}
	return found[0]
}

// ---------------------------------------------------------------------------
// AC-PSD-010 — the bypass is recorded, and --force is not it.
// ---------------------------------------------------------------------------

func TestAcquireBypassIsRecorded(t *testing.T) {
	worktree := dirtyDriftWorktree(t)
	root := t.TempDir()
	writeDriftGateConfig(t, root, true)

	out, err := runIntegrationIn(t, worktree, root,
		"acquire", "--session", "sess-lane8", "--card", "t488", "--allow-settings-drift")
	if err != nil {
		t.Fatalf("acquire with --allow-settings-drift: %v\noutput: %s", err, out)
	}

	lock := readLockRecord(t, root)
	if !lock.SettingsDriftBypass {
		t.Errorf("lock record does not carry the bypass marker: %+v", lock)
	}
	preserved := preservedCopyPath(t, root)
	if lock.SettingsDriftPreserved != preserved {
		t.Errorf("lock settings_drift_preserved: got %q, want %q", lock.SettingsDriftPreserved, preserved)
	}
	if !strings.Contains(out, preserved) {
		t.Errorf("output does not carry the preserved path\noutput: %s", out)
	}
	if !strings.Contains(strings.ToLower(out), "bypass") {
		t.Errorf("output does not say the gate was bypassed\noutput: %s", out)
	}
}

// TestAcquireForceIsNotASettingsDriftBypass — the second half of AC-PSD-010.
// The two flags are different axes; one must not silently do the other's job.
func TestAcquireForceIsNotASettingsDriftBypass(t *testing.T) {
	worktree := dirtyDriftWorktree(t)
	root := t.TempDir()
	writeDriftGateConfig(t, root, true)

	out, err := runIntegrationIn(t, worktree, root, "acquire", "--session", "sess-lane8", "--force")
	if err == nil {
		t.Fatalf("--force bypassed the settings-drift refusal\noutput: %s", out)
	}
	if _, statErr := os.Stat(lockPathFor(root)); statErr == nil {
		t.Errorf("--force recorded the window past a drift refusal")
	}
}

// ---------------------------------------------------------------------------
// AC-PSD-011 / AC-PSD-012 — the preflight verdict surface.
// ---------------------------------------------------------------------------

// TestPreflightUndeterminedOnPredicateFailure — AC-PSD-011, fixture F5.
// The assertion is positive ("status is exactly undetermined"), not negative
// ("status is not clean"): with only two states the negative form would be
// satisfied by `drift`, turning a failed measurement into a false positive.
func TestPreflightUndeterminedOnPredicateFailure(t *testing.T) {
	dir, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "loose.txt"), []byte("not a repository\n"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	root := t.TempDir()

	out, _ := runIntegration(t, root, "preflight", "--json", dir)

	var raw map[string]any
	if err := json.Unmarshal([]byte(out), &raw); err != nil {
		t.Fatalf("preflight --json is not JSON (%v): %s", err, out)
	}
	if raw["status"] != string(kanban.SettingsDriftUndetermined) {
		t.Errorf("status: got %v, want %q", raw["status"], kanban.SettingsDriftUndetermined)
	}
	if _, present := raw["match_count"]; present {
		t.Errorf("match_count is present under undetermined (%v); a 0 there reads as a pass", raw["match_count"])
	}
	if msg, ok := raw["error"].(string); !ok || strings.TrimSpace(msg) == "" {
		t.Errorf("error field is missing or empty: %v", raw["error"])
	}
}

// TestPreflightKeepsDriftVerdictWhenPreservationFails — AC-PSD-012.
// The assertion is on the verdict field, not on the lock file: with the
// refusal layer off by default, no lock-file assertion would run at all.
func TestPreflightKeepsDriftVerdictWhenPreservationFails(t *testing.T) {
	worktree := dirtyDriftWorktree(t)
	root := t.TempDir()

	// Occupy the preserve directory's path with a regular file: MkdirAll then
	// fails for a reason no permission model waves through, a root-owned test
	// runner included.
	if err := os.MkdirAll(filepath.Join(root, ".moai", "state"), 0o755); err != nil {
		t.Fatalf("mkdir state: %v", err)
	}
	if err := os.WriteFile(kanban.SettingsDriftDir(root), []byte("not a directory\n"), 0o644); err != nil {
		t.Fatalf("occupy preserve dir: %v", err)
	}

	out, _ := runIntegration(t, root, "preflight", "--json", worktree)

	var raw map[string]any
	if err := json.Unmarshal([]byte(out), &raw); err != nil {
		t.Fatalf("preflight --json is not JSON (%v): %s", err, out)
	}
	if raw["status"] != string(kanban.SettingsDriftDetected) {
		t.Errorf("status: got %v, want %q (a preservation failure must not flip the verdict)", raw["status"], kanban.SettingsDriftDetected)
	}
	if mc, ok := raw["match_count"].(float64); !ok || int(mc) != 1 {
		t.Errorf("match_count: got %v, want 1", raw["match_count"])
	}
	if msg, ok := raw["preserve_error"].(string); !ok || strings.TrimSpace(msg) == "" {
		t.Errorf("preserve_error is missing or empty: %v", raw["preserve_error"])
	}
}

// TestPreflightReportsCleanAndDrift — AC-PSD-001/002 through the read surface
// REQ-PSD-011 requires, and the `status` carrier REQ-PSD-012 requires.
func TestPreflightReportsCleanAndDrift(t *testing.T) {
	clean := newDriftWorktree(t)
	root := t.TempDir()

	out, err := runIntegration(t, root, "preflight", "--json", clean)
	if err != nil {
		t.Fatalf("preflight on a clean tree errored: %v\noutput: %s", err, out)
	}
	var cleanRaw map[string]any
	if jsonErr := json.Unmarshal([]byte(out), &cleanRaw); jsonErr != nil {
		t.Fatalf("preflight --json is not JSON (%v): %s", jsonErr, out)
	}
	if cleanRaw["status"] != string(kanban.SettingsDriftClean) {
		t.Errorf("clean status: got %v, want %q", cleanRaw["status"], kanban.SettingsDriftClean)
	}
	if mc, ok := cleanRaw["match_count"].(float64); !ok || int(mc) != 0 {
		t.Errorf("clean match_count: got %v, want 0", cleanRaw["match_count"])
	}

	dirty := dirtyDriftWorktree(t)
	root2 := t.TempDir()
	out2, _ := runIntegration(t, root2, "preflight", "--json", "--card", "t488", dirty)
	var dirtyRaw map[string]any
	if jsonErr := json.Unmarshal([]byte(out2), &dirtyRaw); jsonErr != nil {
		t.Fatalf("preflight --json is not JSON (%v): %s", jsonErr, out2)
	}
	if dirtyRaw["status"] != string(kanban.SettingsDriftDetected) {
		t.Errorf("dirty status: got %v, want %q", dirtyRaw["status"], kanban.SettingsDriftDetected)
	}
	if mc, ok := dirtyRaw["match_count"].(float64); !ok || int(mc) != 1 {
		t.Errorf("dirty match_count: got %v, want 1", dirtyRaw["match_count"])
	}
	if p, ok := dirtyRaw["preserved"].(string); !ok || p == "" {
		t.Errorf("preserved path missing from the drift report: %v", dirtyRaw["preserved"])
	}
}

// ---------------------------------------------------------------------------
// AC-PSD-013 — the default posture: no refusal, but the observation layer runs.
// ---------------------------------------------------------------------------

// TestAcquireDefaultPostureObservesWithoutRefusing — AC-PSD-013 (a)-(d).
//
// The kill switch is untouched here, so it resolves to false. An
// implementation that gated the WHOLE gate rather than the refusal layer would
// pass (a) and fail (b), (c) and (d) — which is the only thing keeping the
// operator's stated reason for choosing default-OFF from being prose.
func TestAcquireDefaultPostureObservesWithoutRefusing(t *testing.T) {
	worktree := dirtyDriftWorktree(t)
	root := t.TempDir() // no workflow.yaml at all: the key is absent, not false

	out, err := runIntegrationIn(t, worktree, root, "acquire", "--session", "sess-lane8", "--card", "t488")
	if err != nil {
		t.Fatalf("(a) the default posture refused: %v\noutput: %s", err, out)
	}
	if _, statErr := os.Stat(lockPathFor(root)); statErr != nil {
		t.Fatalf("(a) the window was not recorded: %v", statErr)
	}

	// (b) the preserved copy exists and matches the original byte for byte.
	source := filepath.Join(worktree, ".claude", "settings.json")
	original, readErr := os.ReadFile(source)
	if readErr != nil {
		t.Fatalf("read source: %v", readErr)
	}
	if len(original) == 0 {
		t.Fatalf("control: the source file is empty; byte equality would assert nothing")
	}
	preserved := preservedCopyPath(t, root)
	got, readErr := os.ReadFile(preserved)
	if readErr != nil {
		t.Fatalf("read preserved: %v", readErr)
	}
	if string(got) != string(original) {
		t.Errorf("(b) preserved copy differs\n got: %q\nwant: %q", got, original)
	}

	// (c) exactly one ledger row.
	if lines := driftLedgerLines(t, root); len(lines) != 1 {
		t.Errorf("(c) ledger rows: got %d, want 1", len(lines))
	}

	// (d) the output names the drift and the preserved path.
	if !strings.Contains(out, preserved) {
		t.Errorf("(d) output does not carry the preserved path\noutput: %s", out)
	}
	if !strings.Contains(strings.ToLower(out), "drift") {
		t.Errorf("(d) output does not report drift\noutput: %s", out)
	}

	// The lock record must NOT claim a bypass: nothing was bypassed, because
	// with the refusal layer off there was no refusal to bypass.
	lock := readLockRecord(t, root)
	if lock.SettingsDriftBypass {
		t.Errorf("lock record claims a bypass under the default posture: %+v", lock)
	}
}

// TestAcquireDefaultPostureWithAllowFlagIsIdentical — AC-PSD-013(e).
func TestAcquireDefaultPostureWithAllowFlagIsIdentical(t *testing.T) {
	worktree := dirtyDriftWorktree(t)
	root := t.TempDir()

	out, err := runIntegrationIn(t, worktree, root,
		"acquire", "--session", "sess-lane8", "--card", "t488", "--allow-settings-drift")
	if err != nil {
		t.Fatalf("(e) acquire failed: %v\noutput: %s", err, out)
	}
	if _, statErr := os.Stat(lockPathFor(root)); statErr != nil {
		t.Fatalf("(e)(a) the window was not recorded: %v", statErr)
	}
	preserved := preservedCopyPath(t, root)
	if _, statErr := os.Stat(preserved); statErr != nil {
		t.Fatalf("(e)(b) no preserved copy: %v", statErr)
	}
	if lines := driftLedgerLines(t, root); len(lines) != 1 {
		t.Errorf("(e)(c) ledger rows: got %d, want 1", len(lines))
	}
	if !strings.Contains(out, preserved) {
		t.Errorf("(e)(d) output does not carry the preserved path\noutput: %s", out)
	}

	// The record must not lie: with the refusal layer off, the flag bypassed
	// nothing, so recording it as a bypass would make the record say something
	// that did not happen.
	lock := readLockRecord(t, root)
	if lock.SettingsDriftBypass {
		t.Errorf("(e) lock record claims a bypass with the refusal layer off: %+v", lock)
	}
}

// TestAcquireCleanTreeIsUnchanged pins that the precondition is additive: a
// clean tree acquires exactly as it did before this card, and no preservation
// directory is created.
func TestAcquireCleanTreeIsUnchanged(t *testing.T) {
	worktree := newDriftWorktree(t)
	root := t.TempDir()

	out, err := runIntegrationIn(t, worktree, root, "acquire", "--session", "sess-lane8")
	if err != nil {
		t.Fatalf("acquire on a clean tree: %v\noutput: %s", err, out)
	}
	if !strings.Contains(out, "acquired") {
		t.Errorf("acquire output changed shape on the clean path: %s", out)
	}
	if _, statErr := os.Stat(kanban.SettingsDriftDir(root)); statErr == nil {
		t.Errorf("a clean tree created the preservation directory")
	}
}
