package cli

// Shared fixture for the SPEC-DUAL-HARNESS-HOOK-PARITY-001 M2d Stop-chain
// tests: the AC-HPR-002 goldens, the receipt producers, AC-HPR-003, AC-HPR-005
// and the AC-HPR-016 timing leg.

import (
	"bufio"
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/codexadapter"
	"github.com/modu-ai/moai-adk/internal/hook"
)

// stopFixture is a scratch git project: go.mod + main.go, HEAD a sync-phase
// commit with a code delta, .moai/ git-ignored so hook state never moves the
// working-tree digest, and a private bin directory holding a fake `go`.
type stopFixture struct {
	root string
	bin  string
	path string // PATH for both harness paths: the fake bin first
}

func newStopFixture(t *testing.T) *stopFixture {
	t.Helper()
	return newStopFixtureAt(t, t.TempDir())
}

// newStopFixtureAt builds the fixture inside an existing directory (the
// AC-HPR-003 golden deploys a gpt-profile project there first).
func newStopFixtureAt(t *testing.T, root string) *stopFixture {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("the Claude sync-gate script is bash; the golden needs a POSIX shell")
	}
	f := &stopFixture{root: root, bin: t.TempDir()}
	var dirs []string
	for _, tool := range []string{"git", "bash", "find", "grep", "shasum", "cat", "mktemp"} {
		p, err := exec.LookPath(tool)
		if err != nil {
			t.Fatalf("golden needs %s on PATH: %v", tool, err)
		}
		dirs = append(dirs, filepath.Dir(p))
	}
	f.path = strings.Join(append([]string{f.bin}, append(dirs, "/usr/bin", "/bin")...), string(os.PathListSeparator))
	t.Setenv("PATH", f.path)
	t.Setenv("MOAI_SYNC_GATE_BLOCKING", "")
	t.Setenv("MOAI_AUTONOMY_TIER", "")
	if _, err := os.Stat(filepath.Join(root, ".git")); err != nil {
		f.git(t, "init", "-q", "-b", "main")
	}
	f.write(t, ".gitignore", ".moai/\n")
	f.write(t, "go.mod", "module example.com/fx\n\ngo 1.22\n")
	f.write(t, "main.go", "package main\n\nfunc main() {}\n")
	f.commit(t, "feat: base")
	f.write(t, "main.go", "package main\n\nfunc main() { _ = 1 }\n")
	f.commit(t, "docs(SPEC-FX-001): sync-phase close")
	f.setFakeGo(t, 0)
	return f
}

func (f *stopFixture) git(t *testing.T, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", f.root, "-c", "user.name=fx", "-c", "user.email=fx@example.com", "-c", "commit.gpgsign=false"}, args...)...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
	return strings.TrimSpace(string(out))
}

func (f *stopFixture) write(t *testing.T, rel, content string) {
	t.Helper()
	p := filepath.Join(f.root, rel)
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func (f *stopFixture) commit(t *testing.T, subject string) {
	t.Helper()
	f.git(t, "add", "-A")
	f.git(t, "commit", "-q", "--allow-empty", "-m", subject)
}

// setFakeGo installs a `go` whose vet exits vetExit; build always passes.
func (f *stopFixture) setFakeGo(t *testing.T, vetExit int) {
	t.Helper()
	script := "#!/bin/sh\ncase \"$1\" in vet) echo 'fake vet'; exit " + strconv.Itoa(vetExit) + " ;; build) exit 0 ;; *) exit 0 ;; esac\n"
	if err := os.WriteFile(filepath.Join(f.bin, "go"), []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
}

// enableReviewGates writes the project's review-gate switches (git-ignored).
func (f *stopFixture) enableReviewGates(t *testing.T, codexGate, multiGate bool) {
	t.Helper()
	b := func(v bool) string {
		if v {
			return "true"
		}
		return "false"
	}
	f.write(t, ".moai/config/sections/workflow.yaml",
		"workflow:\n  codex:\n    review_gate:\n      enabled: "+b(codexGate)+"\n  multi:\n    review_gate:\n      enabled: "+b(multiGate)+"\n")
}

// dirty leaves an uncommitted, reviewable change in the tree.
func (f *stopFixture) dirty(t *testing.T, content string) {
	t.Helper()
	f.write(t, "extra.go", "package main\n\n// "+content+"\n")
}

// runClaudeSyncGate runs the distributed Claude sync-gate script, unmodified,
// on the fixture and returns its normalized decision.
func (f *stopFixture) runClaudeSyncGate(t *testing.T, payload string, env ...string) codexadapter.Decision {
	t.Helper()
	script, err := filepath.Abs(filepath.Join("..", "template", "templates", ".claude", "hooks", "moai", "sync-phase-quality-gate.sh"))
	if err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("bash", script)
	cmd.Dir = f.root
	cmd.Env = append([]string{"PATH=" + f.path, "HOME=" + os.Getenv("HOME"), "CLAUDE_PROJECT_DIR=" + f.root}, env...)
	cmd.Stdin = strings.NewReader(payload)
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("claude sync gate: %v\n%s", err, out)
	}
	if strings.Contains(string(out), `"decision":"block"`) {
		return codexadapter.DecisionDeny
	}
	return codexadapter.DecisionAllow
}

func stopInput(session string, active bool) *hook.HookInput {
	return &hook.HookInput{HookEventName: string(hook.EventStop), SessionID: session, StopHookActive: active}
}

func claudeDecision(out *hook.HookOutput) codexadapter.Decision {
	if out != nil && out.Decision == hook.DecisionBlock {
		return codexadapter.DecisionDeny
	}
	return codexadapter.DecisionAllow
}

func sinkRecords(t *testing.T, root string) []codexadapter.Discard {
	t.Helper()
	fh, err := os.Open(filepath.Join(root, codexadapter.DiagnosticSinkRel))
	if err != nil {
		return nil
	}
	defer func() { _ = fh.Close() }()
	var out []codexadapter.Discard
	sc := bufio.NewScanner(fh)
	for sc.Scan() {
		var d codexadapter.Discard
		if json.Unmarshal(sc.Bytes(), &d) == nil {
			out = append(out, d)
		}
	}
	return out
}

// fakeCodexVersion pins the reviewer version both the receipt producer and the
// Stop chain read, so a test does not spawn a real codex binary.
func fakeCodexVersion(t *testing.T, v string) {
	t.Helper()
	prev := codexVersionProbe
	codexVersionProbe = func(context.Context, string) (string, error) { return v, nil }
	t.Cleanup(func() { codexVersionProbe = prev })
}

// wideStopBudget widens every member's internal budget for a decision test,
// so a loaded machine cannot turn it into a timing test. The declared budgets
// are measured by the AC-HPR-016 timing leg instead.
func wideStopBudget(int) time.Duration { return time.Minute }

func countDiscards(recs []codexadapter.Discard, keyPart string) int {
	n := 0
	for _, d := range recs {
		if strings.Contains(d.Key, keyPart) {
			n++
		}
	}
	return n
}
