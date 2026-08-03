package hook

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// These tests exercise the M2 shell-trim changes (SPEC-STOPCHAIN-TRIM-001):
//   - AC-001: handle-stop-goal.sh must NOT exec the moai binary when the goal
//     state file for the current session is absent (the shell-layer
//     precondition short-circuits before any cold-start).
//   - AC-002: sync-phase-quality-gate.sh must NOT run go vet / go build when
//     the HEAD commit subject is not a sync-phase commit (the existing
//     once-per-commit sentinel — A10 second clause, "이미 있음 보강").
//
// The tests use a counting fake-binary approach: a stub `moai` (and `go`) is
// placed first in PATH inside a temp dir, HOME is redirected to a temp dir so
// the 3-tier fallback chain has no real binary to find, and the hook script is
// invoked with controlled stdin. The counter file then asserts whether the
// real binary path was reached.

// requireBash skips when bash is unavailable (Windows CI without Git Bash).
func requireBash(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("bash"); err != nil {
		t.Skipf("bash not in PATH: %v", err)
	}
}

// writeCountingStub writes a stub binary at dir/name that increments the counter
// file at counterPath and exits 0. Returns the absolute stub path.
func writeCountingStub(t *testing.T, dir, name, counterPath string) string {
	t.Helper()
	stub := filepath.Join(dir, name)
	body := "#!/bin/bash\n" +
		"echo \"1\" >> \"" + counterPath + "\"\n" +
		"exit 0\n"
	if runtime.GOOS == "windows" {
		body = "#!/bin/sh\n" + body
	}
	if err := os.WriteFile(stub, []byte(body), 0o755); err != nil {
		t.Fatalf("write stub %s: %v", stub, err)
	}
	return stub
}

// runHookScript runs the named hook script with the given stdin JSON, a custom
// PATH (stubBin first), and HOME pointed at homeDir so the 3-tier moai fallback
// chain finds nothing real. Returns the script's combined exit code.
func runHookScript(t *testing.T, scriptRel string, stdinJSON string, stubBin, homeDir, projectDir string) int {
	t.Helper()
	script := filepath.Join(projectDir, scriptRel)
	if _, err := os.Stat(script); err != nil {
		t.Fatalf("hook script not found: %s (%v)", script, err)
	}
	// Build a PATH that puts stubBin first and keeps system dirs for git/grep.
	sysPath := os.Getenv("PATH")
	path := stubBin + string(os.PathListSeparator) + sysPath

	cmd := exec.Command("bash", script)
	cmd.Stdin = strings.NewReader(stdinJSON)
	cmd.Dir = projectDir
	cmd.Env = []string{
		"PATH=" + path,
		"HOME=" + homeDir,
		"CLAUDE_PROJECT_DIR=" + projectDir,
	}
	out, err := cmd.CombinedOutput()
	code := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			code = exitErr.ExitCode()
		} else {
			t.Fatalf("run %s: %v; output=%s", scriptRel, err, out)
		}
	}
	return code
}

// TestAC001_GoalAbsentSkipsMoaiBinary (A10 / AC-STOPCHAIN-TRIM-001):
// a session with NO armed goal must pay zero moai-binary cold-starts at
// turn-end. The shell precondition must short-circuit before exec'ing moai.
func TestAC001_GoalAbsentSkipsMoaiBinary(t *testing.T) {
	requireBash(t)
	requireGit(t)
	projectDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	// Locate the repo root (two levels up from internal/hook) so the hook path
	// resolves the same way it ships.
	repoRoot := filepath.Join(projectDir, "..", "..")

	tmp := t.TempDir()
	stubBin := filepath.Join(tmp, "bin")
	if err := os.MkdirAll(stubBin, 0o755); err != nil {
		t.Fatal(err)
	}
	homeDir := filepath.Join(tmp, "home")
	if err := os.MkdirAll(homeDir, 0o755); err != nil {
		t.Fatal(err)
	}
	counter := filepath.Join(tmp, "moai.count")
	writeCountingStub(t, stubBin, "moai", counter)

	// Session has NO goal state file. The state dir does not even exist.
	stdin := `{"session_id":"sess-no-goal-001","last_assistant_message":"done"}`
	_ = runHookScript(t, ".claude/hooks/moai/handle-stop-goal.sh", stdin, stubBin, homeDir, repoRoot)

	got := readCounter(t, counter)
	if got != 0 {
		t.Fatalf("AC-001 REGRESSION: moai binary invoked %d time(s) on a goal-less session; expected 0 (shell precondition missing?)", got)
	}

	// Positive case: arm a goal (write the state file) → moai IS invoked exactly once.
	stateDir := filepath.Join(repoRoot, ".moai", "state", "goal")
	if err := os.MkdirAll(stateDir, 0o755); err != nil {
		t.Fatal(err)
	}
	stateFile := filepath.Join(stateDir, "sess-no-goal-001.json")
	if err := os.WriteFile(stateFile, []byte(`{"armed":true}`), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Remove(stateFile) })

	_ = runHookScript(t, ".claude/hooks/moai/handle-stop-goal.sh", stdin, stubBin, homeDir, repoRoot)
	gotPositive := readCounter(t, counter)
	if gotPositive != 1 {
		t.Fatalf("AC-001 positive case: moai binary invoked %d time(s) when a goal WAS armed; expected exactly 1", gotPositive)
	}
}

// TestAC002_NonSyncHeadSkipsVetBuild (A10 / AC-STOPCHAIN-TRIM-002):
// a non-sync HEAD commit subject must short-circuit sync-phase-quality-gate.sh
// before running go vet / go build / lint. Verified by counting stubbed `go`
// invocations.
func TestAC002_NonSyncHeadSkipsVetBuild(t *testing.T) {
	requireBash(t)
	requireGit(t)
	repoRoot := filepath.Join(mustGetwd(t), "..", "..")

	tmp := t.TempDir()
	stubBin := filepath.Join(tmp, "bin")
	if err := os.MkdirAll(stubBin, 0o755); err != nil {
		t.Fatal(err)
	}
	homeDir := filepath.Join(tmp, "home")
	if err := os.MkdirAll(homeDir, 0o755); err != nil {
		t.Fatal(err)
	}
	goCounter := filepath.Join(tmp, "go.count")
	writeCountingStub(t, stubBin, "go", goCounter)

	// We cannot mutate the real repo's HEAD. Instead, copy the gate into a
	// throwaway git fixture so we control the commit subject. detect_language
	// needs a go.mod present to classify as "go".
	fixture := t.TempDir()
	fixtureRepo := initGoFixtureRepo(t, fixture, "feat(scope): run-phase edit — not a sync commit")

	// Run the gate script sourced from the repo, but with CWD = fixture so
	// git/detect_language resolve against the fixture.
	script := filepath.Join(repoRoot, ".claude", "hooks", "moai", "sync-phase-quality-gate.sh")
	cmd := exec.Command("bash", script)
	cmd.Stdin = strings.NewReader(`{}`)
	cmd.Dir = fixtureRepo
	cmd.Env = []string{
		"PATH=" + stubBin + string(os.PathListSeparator) + os.Getenv("PATH"),
		"HOME=" + homeDir,
		"CLAUDE_PROJECT_DIR=" + fixtureRepo,
	}
	if out, err := cmd.CombinedOutput(); err != nil {
		// exit 0 is expected on the skip path; a non-zero is fine as long as
		// go was not invoked — but log it for diagnosis.
		t.Logf("gate output (non-sync HEAD): %s", out)
		_ = err
	}

	if n := readCounter(t, goCounter); n != 0 {
		t.Fatalf("AC-002 REGRESSION: go invoked %d time(s) on a non-sync HEAD; expected 0", n)
	}

	// Positive case: flip HEAD subject to a sync commit → go IS invoked.
	resetHeadSubject(t, fixtureRepo, "docs(SPEC-X): sync-phase artifacts")
	goCounter2 := filepath.Join(tmp, "go2.count")
	writeCountingStub(t, stubBin, "go", goCounter2)
	// Clear the once-per-commit sentinel so the gate re-runs.
	_ = os.RemoveAll(filepath.Join(fixtureRepo, ".moai", "state"))
	cmd2 := exec.Command("bash", script)
	cmd2.Stdin = strings.NewReader(`{}`)
	cmd2.Dir = fixtureRepo
	cmd2.Env = []string{
		"PATH=" + stubBin + string(os.PathListSeparator) + os.Getenv("PATH"),
		"HOME=" + homeDir,
		"CLAUDE_PROJECT_DIR=" + fixtureRepo,
	}
	_, _ = cmd2.CombinedOutput()
	if n := readCounter(t, goCounter2); n == 0 {
		t.Fatalf("AC-002 positive case: go NOT invoked on a sync-phase HEAD; expected >=1 (early-exit over-firing?)")
	}
}

// --- helpers ---

func readCounter(t *testing.T, path string) int {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return 0
		}
		t.Fatalf("read counter %s: %v", path, err)
	}
	lines := strings.Split(strings.TrimSpace(string(b)), "\n")
	n := 0
	for _, ln := range lines {
		if strings.TrimSpace(ln) != "" {
			n++
		}
	}
	return n
}

func mustGetwd(t *testing.T) string {
	t.Helper()
	d, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	return d
}

func initGoFixtureRepo(t *testing.T, dir, headSubject string) string {
	t.Helper()
	mustRunGit(t, dir, "init")
	mustRunGit(t, dir, "config", "user.email", "stopchain-test@example.com")
	mustRunGit(t, dir, "config", "user.name", "Stopchain Test")
	mustRunGit(t, dir, "config", "core.hooksPath", "/dev/null")
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module fixture\n\ngo 1.23\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main\nfunc main() {}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	mustRunGit(t, dir, "add", "go.mod", "main.go")
	mustRunGit(t, dir, "commit", "-m", headSubject)
	return dir
}

func resetHeadSubject(t *testing.T, dir, subject string) {
	t.Helper()
	// Amend the existing commit's subject without changing the tree.
	mustRunGit(t, dir, "commit", "--amend", "-m", subject)
}

// jsonStdin is a tiny helper for ad-hoc stdin construction in other ACs.
func jsonStdin(t *testing.T, v map[string]any) string {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}
