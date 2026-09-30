// hook_gate_multi_language_test.go: behavioural guard for multi-language check
// execution in the sync-phase quality gate.
//
// Observed-RED pair, same input both cells: a node+python fixture whose sync
// commit changes both languages, with a failing eslint stub and a passing ruff
// stub. While the gate executed only the first detected candidate, python came
// first in detection order, ruff alone ran, and the gate recorded pass while
// the configured-to-fail eslint never ran. The gate must run every CHANGED
// language's checks, fold the worst exit into the decision, and leave an
// absent tool surfaced as a skip — never counted as a pass.
//
// Sentinel on failure: SYNC_GATE_SINGLE_LANGUAGE_CHECK
package template_test

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// gateMultiRun builds a sync-phase fixture from base (committed first, may be
// empty) plus sync (the sync-phase commit), stubs every tool in fail's key set
// plus "go" on PATH with the given failing state, runs the gate, and returns
// stdout, the audit log, and the stub invocation log.
func gateMultiRun(t *testing.T, base, sync map[string]string, fail map[string]bool) (string, string, string) {
	t.Helper()
	dir := t.TempDir()
	bin := t.TempDir()
	stubLog := filepath.Join(t.TempDir(), "stub.log")

	tools := []string{"eslint", "ruff", "go"}
	for _, name := range tools {
		code := "0"
		if fail[name] {
			code = "7"
		}
		script := fmt.Sprintf("#!/bin/sh\necho \"stub %s $*\" >> \"$STUB_LOG\"\nexit %s\n", name, code)
		if err := os.WriteFile(filepath.Join(bin, name), []byte(script), 0o755); err != nil {
			t.Fatalf("write stub %s: %v", name, err)
		}
	}
	write := func(files map[string]string) {
		t.Helper()
		for name, body := range files {
			if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
				t.Fatalf("write %s: %v", name, err)
			}
		}
	}
	git := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", append([]string{"-c", "user.email=t1379@example.invalid", "-c", "user.name=t1379"}, args...)...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, out)
		}
	}
	git("init", "--quiet")
	if len(base) > 0 {
		write(base)
		git("add", ".")
		git("commit", "--quiet", "-m", "chore: base fixture")
	}
	write(sync)
	git("add", ".")
	git("commit", "--quiet", "-m", "docs: sync-phase fixture")

	gate := filepath.Join(hocProjectRoot(t), "internal", "template", "templates",
		".claude", "hooks", "moai", "sync-phase-quality-gate.sh")
	cmd := exec.Command("bash", gate)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(),
		"PATH="+bin+string(os.PathListSeparator)+os.Getenv("PATH"),
		"CLAUDE_PROJECT_DIR="+dir,
		"MOAI_SYNC_GATE_BLOCKING=1",
		"MOAI_AUTONOMY_TIER=",
		"STUB_LOG="+stubLog,
	)
	out, _ := cmd.CombinedOutput() // the gate always exits 0; its verdict rides stdout
	audit, _ := os.ReadFile(filepath.Join(dir, ".moai", "logs", "sync-quality-gate.log"))
	stub, _ := os.ReadFile(stubLog)
	return string(out), string(audit), string(stub)
}

// The regression itself: python is detected before node, so a
// first-candidate-only gate ran ruff (passing) and never invoked the failing
// eslint. Both checkers must run and the failure must block.
func TestSyncGateMultiLanguage_FailingCheckerInLaterLanguageBlocks(t *testing.T) {
	gateExitRequire(t)
	files := map[string]string{
		"package.json":   "{}\n",
		"a.js":           "console.log(1);\n",
		"pyproject.toml": "[project]\nname = \"f\"\n",
		"a.py":           "print(1)\n",
	}
	out, audit, stub := gateMultiRun(t, nil, files, map[string]bool{"eslint": true})
	if !strings.Contains(stub, "stub ruff") || !strings.Contains(stub, "stub eslint") {
		t.Fatalf("premise: both checkers must be invoked.\nstub log: %q\naudit: %q", stub, audit)
	}
	if !strings.Contains(audit, "language=python,node ") {
		t.Fatalf("premise: both changed languages must be recorded.\naudit: %q", audit)
	}
	if !strings.Contains(out, `"decision":"block"`) {
		t.Errorf("SYNC_GATE_SINGLE_LANGUAGE_CHECK: eslint exited 7 but the gate allowed — only one language was checked.\naudit: %q\nout: %q", audit, out)
	}
}

// The control: the same fixture with both checkers passing must allow, so the
// case above cannot be satisfied by a gate that blocks unconditionally.
func TestSyncGateMultiLanguage_PassingCheckersAllow(t *testing.T) {
	gateExitRequire(t)
	files := map[string]string{
		"package.json":   "{}\n",
		"a.js":           "console.log(1);\n",
		"pyproject.toml": "[project]\nname = \"f\"\n",
		"a.py":           "print(1)\n",
	}
	out, audit, stub := gateMultiRun(t, nil, files, nil)
	if !strings.Contains(stub, "stub ruff") || !strings.Contains(stub, "stub eslint") {
		t.Fatalf("premise: both checkers must be invoked.\nstub log: %q", stub)
	}
	if strings.Contains(out, `"decision":"block"`) {
		t.Errorf("passing checkers in every changed language must not block.\naudit: %q\nout: %q", audit, out)
	}
}

// The gate checks what the sync commit changed, not every marker the project
// carries: a python-only sync in a go+python project runs ruff and never the
// go checker.
func TestSyncGateMultiLanguage_OnlyChangedLanguagesChecked(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("sync-phase-quality-gate.sh is a POSIX shell script")
	}
	base := map[string]string{"go.mod": "module f\n\ngo 1.23\n"}
	sync := map[string]string{
		"pyproject.toml": "[project]\nname = \"f\"\n",
		"a.py":           "print(1)\n",
	}
	out, audit, stub := gateMultiRun(t, base, sync, nil)
	if !strings.Contains(stub, "stub ruff") {
		t.Fatalf("premise: the ruff checker must run for a python change.\nstub log: %q\naudit: %q", stub, audit)
	}
	if strings.Contains(stub, "stub go") {
		t.Errorf("SYNC_GATE_SINGLE_LANGUAGE_CHECK inverse: the go checker ran for a python-only sync — checks must follow the changed languages.\nstub log: %q", stub)
	}
	if strings.Contains(out, `"decision":"block"`) {
		t.Errorf("a passing python-only sync must not block.\naudit: %q\nout: %q", audit, out)
	}
}
