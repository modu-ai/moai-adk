package kanban

// foreman_queue_statement_test.go — SPEC-TODO-QUEUE-HOME-CANON-001 AC-003a:
// the docs-match-resolution guard. The kanban-foreman skill's queue-watch
// script documents where the backlog queue lives; this test pins that
// documentation to the resolver by EXECUTING the skill's own shell fragment
// inside a git-repository fixture and comparing the directory it computes
// against StateDirForRoot for the same fixture.
//
// A parse-only check ("the doc mentions the home layout") would pass while the
// script kept watching a directory the resolver never uses — the exact
// silent-no-op defect this SPEC repairs. Execution is the honest equality
// check, and it fails in either drift direction: a doc edit that recomputes
// the path wrongly, or a resolver change the doc no longer mirrors.
//
// Fixture discipline (acceptance.md edge cases): a t.TempDir() root IS a
// temporary origin, so the DEFAULT resolution for it is project-local by
// design. The test mirrors SPEC-TODO-HOME-TEMP-GUARD-001's control pattern —
// it sets MOAI_HOME to a fixture-local ABSOLUTE path, which the resolver
// honors for any origin (state_dir.go override branch), and pins HOME to a
// canary so neither side can touch the operator's home. Executing sh keeps
// this test POSIX-only; it is skipped where sh does not exist.

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// foremanSkillRelPath is the skill's path from this package directory. The
// TEMPLATE source tree is the subject: the deployed local copy is byte-identical
// by parity test elsewhere, and the template is what every future project gets.
var foremanSkillRelPath = filepath.Join("..", "template", "templates", ".claude", "skills", "moai-kanban-foreman", "SKILL.md")

// extractForemanWatchResolution pulls the sh fence out of the skill and returns
// its resolution prologue — the lines that compute the queue directory $d,
// up to (not including) the watch loop's `last=init` sentinel. If the fence or
// the sentinel is missing the skill changed shape and the guard must say so
// rather than execute the wrong text.
func extractForemanWatchResolution(t *testing.T, src string) []string {
	t.Helper()
	lines := strings.Split(src, "\n")
	start := -1
	for i, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "```sh") {
			start = i + 1
			break
		}
	}
	if start < 0 {
		t.Fatalf("no ```sh fenced block in %s — the queue-watch script moved", foremanSkillRelPath)
	}
	end := -1
	for i := start; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "```" {
			end = i
			break
		}
	}
	if end < 0 {
		t.Fatalf("unterminated ```sh fence in %s", foremanSkillRelPath)
	}
	body := lines[start:end]
	boundary := -1
	for i, line := range body {
		if strings.TrimSpace(line) == "last=init" {
			boundary = i
			break
		}
	}
	if boundary < 0 {
		t.Fatalf("the queue-watch script in %s lost its resolution/watch boundary (the `last=init` sentinel) — re-point this guard at the new shape", foremanSkillRelPath)
	}
	return body[:boundary]
}

func TestForemanQueueWatchResolvesCanonicalStateDir(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("the queue-watch script is a POSIX sh fragment; verified on the darwin/linux legs")
	}
	if _, err := exec.LookPath("sh"); err != nil {
		t.Skipf("no sh on PATH: %v", err)
	}
	src, err := os.ReadFile(filepath.Clean(foremanSkillRelPath))
	if err != nil {
		t.Fatalf("read the skill: %v", err)
	}
	if strings.Contains(string(src), "state/todo") {
		t.Fatalf("the skill names the project-local .moai/state/todo path — the dead watch target this SPEC removed")
	}
	resolution := extractForemanWatchResolution(t, string(src))
	script := strings.Join(resolution, "\n") + "\nprintf '%s\\n' \"$d\"\n"

	root := t.TempDir()
	if out, err := exec.Command("git", "init", root).CombinedOutput(); err != nil {
		t.Fatalf("git init the fixture: %v (%s)", err, out)
	}
	canaryHome := t.TempDir()
	moaiHome := t.TempDir()

	scriptPath := filepath.Join(t.TempDir(), "watch-resolution.sh")
	if err := os.WriteFile(scriptPath, []byte(script), 0o700); err != nil {
		t.Fatalf("write the extracted script: %v", err)
	}
	cmd := exec.Command("sh", scriptPath)
	cmd.Dir = root
	env := envWithout(t, "MOAI_HOME", "HOME")
	env = append(env, "HOME="+canaryHome, "MOAI_HOME="+moaiHome)
	cmd.Env = env
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("the skill's resolution prologue failed inside the fixture: %v\n%s", err, out)
	}
	got := strings.TrimSpace(string(out))
	if got == "" {
		t.Fatal("the skill's resolution prologue printed no queue directory — it no longer assigns $d")
	}

	t.Setenv("MOAI_HOME", moaiHome)
	want := StateDirForRoot(root)
	if got != want {
		t.Errorf("the skill's watch directory %q does not match the resolver %q — the documented queue path has drifted from StateDirForRoot", got, want)
	}
	if want == projectStateDirForRoot(root) {
		t.Errorf("the resolver answered project-local %q for the standard git-repo case — the fixture lost its home resolution (MOAI_HOME override broken?)", want)
	}
	if !strings.HasPrefix(got, moaiHome+string(os.PathSeparator)) {
		t.Errorf("the skill's watch directory %q is not beneath the moai home %q", got, moaiHome)
	}
	if got == filepath.Join(root, ".moai", "state", "todo") {
		t.Errorf("the skill's watch directory is the retired project-local path %q", got)
	}
}

// envWithout copies os.Environ() minus the named variables, so the fixture
// cannot inherit a developer's MOAI_HOME or HOME.
func envWithout(t *testing.T, names ...string) []string {
	t.Helper()
	drop := make(map[string]bool, len(names))
	for _, n := range names {
		drop[n] = true
	}
	var env []string
	for _, kv := range os.Environ() {
		name, _, _ := strings.Cut(kv, "=")
		if drop[name] {
			continue
		}
		env = append(env, kv)
	}
	return env
}
