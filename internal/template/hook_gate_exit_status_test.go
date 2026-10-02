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
	"bytes"
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

// --- card t1395: an absent-tool skip must reach the user ---
//
// Four language checks (csharp dotnet, elixir mix, flutter dart, swift swift)
// wrapped run_step in an external pipeline (`run_step … | head -N`), so the
// absent-tool branch's SKIPPED_TOOLS update landed in the pipeline's subshell
// and was lost: the user saw an empty stdout, the audit recorded a bare
// skipped_tools= with an empty value, and the allow was cached as a verified
// pass. These tests drive the gate with the language's tool genuinely ABSENT
// from a hermetic PATH, making every surface of the skip observable: the
// stdout notice, the audit value, the journal-derived trail, and the t1385
// pass-record contract.
//
// Sentinel on failure: SYNC_GATE_SKIP_LOST

type gateSkipCase struct {
	lang  string
	tool  string
	files map[string]string
}

func gateSkipCases() []gateSkipCase {
	return []gateSkipCase{
		{"csharp", "dotnet", map[string]string{"f.csproj": "<Project/>\n", "a.cs": "class A {}\n"}},
		{"elixir", "mix", map[string]string{"mix.exs": "defmodule F do end\n", "a.ex": "1\n"}},
		{"flutter", "dart", map[string]string{"pubspec.yaml": "name: f\n", "a.dart": "void main() {}\n"}},
		{"swift", "swift", map[string]string{"Package.swift": "// f\n", "a.swift": "let a = 1\n"}},
	}
}

// gateSkipScrubTools: every external binary the gate script invokes on these
// four language paths — write_state_file's mv included, whose absence is
// silently swallowed (2>/dev/null) and would falsify the record assertions.
// The fixture PATH contains exactly these (symlinked from the host) and
// nothing else, so tool presence is fully determined by the fixture — no
// ambient PATH axis, lane env included, can leak in.
var gateSkipScrubTools = []string{
	"git", "bash", "sh", "find", "head", "grep", "awk", "sed", "stat", "date",
	"shasum", "sort", "tr", "cut", "cat", "mkdir", "rm", "tail", "cksum",
	"mktemp", "env", "mv", "dirname",
}

// gateSkipBin builds that hermetic bin directory, excluding the given tool
// name (which must not itself be a gate dependency).
func gateSkipBin(t *testing.T, exclude string) string {
	t.Helper()
	bin := filepath.Join(t.TempDir(), "bin")
	if err := os.MkdirAll(bin, 0o755); err != nil {
		t.Fatalf("mkdir scrub bin: %v", err)
	}
	for _, tool := range gateSkipScrubTools {
		host, err := exec.LookPath(tool)
		if err != nil {
			t.Skipf("%s not on PATH (host lacks a gate dependency)", tool)
		}
		link := filepath.Join(bin, tool)
		if err := os.Symlink(host, link); err != nil {
			b, rerr := os.ReadFile(host)
			if rerr != nil {
				t.Fatalf("copy %s: %v", host, rerr)
			}
			if err := os.WriteFile(link, b, 0o755); err != nil {
				t.Fatalf("write %s: %v", link, err)
			}
		}
	}
	if _, err := os.Stat(filepath.Join(bin, exclude)); err == nil {
		t.Fatalf("scrub failed: %s is itself a gate dependency", exclude)
	}
	return bin
}

// lookPathIn resolves tool the way exec.LookPath would inside an environment
// whose PATH is exactly dir — the in-test analog of the RED harness's pc1.
func lookPathIn(dir, tool string) (string, bool) {
	full := filepath.Join(dir, tool)
	if fi, err := os.Stat(full); err == nil && !fi.IsDir() && fi.Mode()&0o111 != 0 {
		return full, true
	}
	return "", false
}

// gateSkipRun commits the language's marker fixture, runs the gate with a
// hermetic PATH that excludes the language's tool, and returns stdout, the
// audit log, and the record file body.
func gateSkipRun(t *testing.T, c gateSkipCase) (string, string, string) {
	t.Helper()
	dir := t.TempDir()
	bin := gateSkipBin(t, c.tool)
	if p, ok := lookPathIn(bin, c.tool); ok {
		t.Fatalf("PC1 failed: %s resolvable in the scrubbed env at %s", c.tool, p)
	}
	for name, body := range c.files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}
	git := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", append([]string{"-c", "user.email=t1395@example.invalid", "-c", "user.name=t1395"}, args...)...)
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
	cmd := exec.Command("bash", gate)
	cmd.Dir = dir
	cmd.Env = []string{
		"PATH=" + bin,
		"HOME=" + t.TempDir(),
		"TMPDIR=" + t.TempDir(),
		"CLAUDE_PROJECT_DIR=" + dir,
		"MOAI_SYNC_GATE_BLOCKING=1",
		"MOAI_AUTONOMY_TIER=",
	}
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("run gate: %v\nstderr: %s", err, stderr.String())
	}
	audit, _ := os.ReadFile(filepath.Join(dir, ".moai", "logs", "sync-quality-gate.log"))
	record, _ := os.ReadFile(filepath.Join(dir, ".moai", "state", "sync-quality-gate.last"))
	return stdout.String(), string(audit), strings.TrimRight(string(record), "\n")
}

// Every absent-tool skip must reach the user. Four surfaces per language:
//
//	stdout  — the systemMessage notice names the absent checker (the pipe
//	          subshell used to swallow it: stdout came back empty)
//	audit   — the decision line carries skipped_tools=<tool> (it used to read
//	          skipped_tools= with an empty value)
//	journal — the per-step trail records the skip (that line always landed in
//	          the log; only the variable feeding the notice was lost)
//	record  — the t1385 contract: the allow still writes its pass record
func TestSyncGateSkipNotice_ToolAbsentNotifies_AllFourLanguages(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("sync-phase-quality-gate.sh is a POSIX shell script")
	}
	for _, c := range gateSkipCases() {
		t.Run(c.lang, func(t *testing.T) {
			out, audit, record := gateSkipRun(t, c)
			lines := strings.Split(strings.TrimRight(audit, "\n"), "\n")
			if len(lines) == 0 || lines[len(lines)-1] == "" {
				t.Fatalf("premise: audit log has no decision line.\naudit: %q", audit)
			}
			last := lines[len(lines)-1]
			if !strings.Contains(last, "language="+c.lang+" ") {
				t.Fatalf("premise: gate did not detect %s.\naudit: %q", c.lang, audit)
			}
			if !strings.Contains(last, "decision=allow") {
				t.Errorf("a skipped check must allow (it never ran).\naudit line: %q", last)
			}
			if !strings.Contains(last, "skipped_tools= "+c.tool) {
				t.Errorf("SYNC_GATE_SKIP_LOST: audit decision line does not name the absent %s.\naudit line: %q", c.tool, last)
			}
			if !strings.Contains(audit, "tool="+c.tool+" skipped") {
				t.Errorf("SYNC_GATE_SKIP_LOST: journal-derived step trail missing the %s skip.\naudit: %q", c.tool, audit)
			}
			if out == "" {
				t.Errorf("SYNC_GATE_SKIP_LOST: stdout is empty — the absent-%s notice never reached the user", c.tool)
			}
			if !strings.Contains(out, `"systemMessage"`) {
				t.Errorf("SYNC_GATE_SKIP_LOST: stdout lacks the systemMessage notice.\nstdout: %q", out)
			}
			if !strings.Contains(out, c.tool) {
				t.Errorf("SYNC_GATE_SKIP_LOST: the notice does not name the absent checker %s.\nstdout: %q", c.tool, out)
			}
			if strings.Contains(out, `"decision"`) {
				t.Errorf("the skip notice must ride the advisory channel only, never a decision payload.\nstdout: %q", out)
			}
			fields := strings.Fields(record)
			if len(fields) < 2 || fields[1] != "pass" {
				t.Errorf("t1385 contract: allow-with-skip must still write its pass record; record = %q", record)
			}
		})
	}
}
