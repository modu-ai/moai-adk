// hook_gate_exit_status_test.go: behavioural guard proving that a failing
// per-language c1 check in the sync-phase quality gate reaches the decision.
//
// The gate wraps several checkers in shell pipelines or per-file loops. Two
// shapes discard the checker's exit status (card t602 / hooks audit H07):
//
//	A. `checker … 2>&1 | head -20` — the pipeline's status is head's, so a
//	   failed compile is recorded as c1=0.
//	B. `find … -exec checker {} \;` and `… | while read f; do checker; done` —
//	   find ignores each invocation's status, and a loop reports only its
//	   last iteration, so one broken file among several passes.
//
// The real toolchains are replaced by stubs on PATH so every branch is
// exercised on any POSIX host: a stub exits 7 when STUB_FAIL=1 or when one of
// its arguments names a "bad" file, and 0 otherwise.
//
// Sentinel on failure: SYNC_GATE_EXIT_STATUS_LOST
package template_test

import (
	"maps"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

const gateStubScript = `#!/bin/sh
echo "stub $(basename "$0") $*" >> "$STUB_LOG"
[ "$STUB_FAIL" = "1" ] && exit 7
for a in "$@"; do
    case "$a" in *bad*) echo "stub failure: $a" >&2; exit 7 ;; esac
done
exit 0
`

type gateExitCase struct {
	lang  string
	tools []string          // stubbed executables (guard name + invoked name)
	files map[string]string // marker + sources; "bad" in a name makes the stub fail
}

func gateExitCases() []gateExitCase {
	return []gateExitCase{
		{"java", []string{"javac"}, map[string]string{"pom.xml": "<project/>\n", "A.java": "class A {}\n"}},
		{"kotlin", []string{"kotlinc"}, map[string]string{"build.gradle.kts": "plugins { kotlin(\"jvm\") }\n", "A.kt": "class A\n"}},
		{"scala", []string{"scalac"}, map[string]string{"build.sbt": "name := \"f\"\n", "A.scala": "object A\n"}},
		{"ruby", []string{"ruby"}, map[string]string{"Gemfile": "source \"x\"\n", "a.rb": "1\n"}},
		{"php", []string{"php"}, map[string]string{"composer.json": "{}\n", "a.php": "<?php\n"}},
		{"r", []string{"R", "Rscript"}, map[string]string{"DESCRIPTION": "Package: f\n", "a.R": "1\n"}},
		{"elixir", []string{"mix"}, map[string]string{"mix.exs": "defmodule F do end\n", "a.ex": "1\n"}},
		{"csharp", []string{"dotnet"}, map[string]string{"f.csproj": "<Project/>\n", "a.cs": "class A {}\n"}},
		{"flutter", []string{"dart"}, map[string]string{"pubspec.yaml": "name: f\n", "a.dart": "void main() {}\n"}},
		{"swift", []string{"swift"}, map[string]string{"Package.swift": "// f\n", "a.swift": "let a = 1\n"}},
	}
}

func gateExitRequire(t *testing.T) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("sync-phase-quality-gate.sh is a POSIX shell script")
	}
	for _, tool := range []string{"git", "bash", "find", "head"} {
		if _, err := exec.LookPath(tool); err != nil {
			t.Skipf("%s not on PATH", tool)
		}
	}
}

// gateExitRun builds a one-commit sync-phase fixture, runs the gate with the
// stubs first on PATH, and returns stdout, the audit log, and the stub log.
func gateExitRun(t *testing.T, c gateExitCase, files map[string]string, stubFail bool) (string, string, string) {
	t.Helper()
	dir := t.TempDir()
	bin := t.TempDir()
	stubLog := filepath.Join(t.TempDir(), "stub.log")

	for _, tool := range c.tools {
		if err := os.WriteFile(filepath.Join(bin, tool), []byte(gateStubScript), 0o755); err != nil {
			t.Fatalf("write stub %s: %v", tool, err)
		}
	}
	for name, body := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}
	git := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", append([]string{"-c", "user.email=t602@example.invalid", "-c", "user.name=t602"}, args...)...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, out)
		}
	}
	git("init", "--quiet")
	git("add", ".")
	git("commit", "--quiet", "-m", "docs: sync-phase fixture")

	gate := filepath.Join(hocProjectRoot(t), "internal", "template", "templates",
		".claude", "hooks", "moai", "sync-phase-quality-gate.sh")
	fail := "0"
	if stubFail {
		fail = "1"
	}
	cmd := exec.Command("bash", gate)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(),
		"PATH="+bin+string(os.PathListSeparator)+os.Getenv("PATH"),
		"CLAUDE_PROJECT_DIR="+dir,
		"MOAI_SYNC_GATE_BLOCKING=1",
		"MOAI_AUTONOMY_TIER=",
		"STUB_LOG="+stubLog,
		"STUB_FAIL="+fail,
	)
	out, _ := cmd.CombinedOutput() // the gate always exits 0; its verdict rides stdout
	audit, _ := os.ReadFile(filepath.Join(dir, ".moai", "logs", "sync-quality-gate.log"))
	stub, _ := os.ReadFile(stubLog)
	return string(out), string(audit), string(stub)
}

// A failing checker MUST block. Every case first asserts the stub was invoked
// and the intended language was detected, so a pass cannot come from a gate
// that exited early and measured nothing.
func TestSyncGateExitStatus_FailingCheckerBlocks(t *testing.T) {
	gateExitRequire(t)
	for _, c := range gateExitCases() {
		t.Run(c.lang, func(t *testing.T) {
			out, audit, stub := gateExitRun(t, c, c.files, true)
			if stub == "" {
				t.Fatalf("premise: stub %v was never invoked.\naudit: %q\nout: %q", c.tools, audit, out)
			}
			if !strings.Contains(audit, "language="+c.lang+" ") {
				t.Fatalf("premise: gate did not detect %s.\naudit: %q", c.lang, audit)
			}
			if !strings.Contains(out, `"decision":"block"`) {
				t.Errorf("SYNC_GATE_EXIT_STATUS_LOST: %s checker exited 7 but the gate allowed.\naudit: %q\nout: %q", c.lang, audit, out)
			}
		})
	}
}

// The control: a gate that blocked unconditionally would satisfy the case above.
func TestSyncGateExitStatus_PassingCheckerAllows(t *testing.T) {
	gateExitRequire(t)
	for _, c := range gateExitCases() {
		t.Run(c.lang, func(t *testing.T) {
			out, audit, stub := gateExitRun(t, c, c.files, false)
			if stub == "" {
				t.Fatalf("premise: stub %v was never invoked.\naudit: %q", c.tools, audit)
			}
			if strings.Contains(out, `"decision":"block"`) {
				t.Errorf("a passing %s checker must not block.\naudit: %q\nout: %q", c.lang, audit, out)
			}
		})
	}
}

// Per-file checkers must aggregate: one broken file among several blocks,
// regardless of the order find happens to visit them in.
func TestSyncGateExitStatus_OneBadFileAmongManyBlocks(t *testing.T) {
	gateExitRequire(t)
	ext := map[string]string{"ruby": ".rb", "php": ".php", "r": ".R"}
	for _, c := range gateExitCases() {
		e, ok := ext[c.lang]
		if !ok {
			continue
		}
		t.Run(c.lang, func(t *testing.T) {
			files := maps.Clone(c.files)
			files["bad"+e] = "1\n"
			files["z_ok"+e] = "1\n"
			files["zz_ok"+e] = "1\n"
			out, audit, stub := gateExitRun(t, c, files, false)
			if !strings.Contains(stub, "bad"+e) {
				t.Fatalf("premise: the broken file was never checked.\nstub log: %q", stub)
			}
			if !strings.Contains(out, `"decision":"block"`) {
				t.Errorf("SYNC_GATE_EXIT_STATUS_LOST: a broken %s file among valid ones did not block.\naudit: %q\nstub log: %q", c.lang, audit, stub)
			}
		})
	}
}
