// hook_cpp_gate_behavior_test.go: behavioural guard for the sync-phase quality
// gate's C++ (c1) check.
//
// These tests EXECUTE the gate against throwaway git repositories rather than
// grepping its source, because the defect this file guards against was invisible
// to source inspection: the command looked correct and silently checked nothing.
//
// Two independent failures were repaired together (card t603 / hooks audit H08);
// each gets its own case, because either one alone still leaves the gate inert:
//
//	A. `-name "*.cpp" -o -name "*.cc" -exec …` parses as `*.cpp -o ( *.cc -a -exec )`,
//	   so a .cpp file short-circuits the -o and never reaches the compiler.
//	B. `-exec … \;` does not propagate the utility's exit status, so even a .cc
//	   file whose compilation failed was recorded as c1=0 — a pass.
//
// Sentinel on failure: SYNC_GATE_CPP_INERT
//
// @MX:ANCHOR: [AUTO] C++ sync-gate behavioural guard — the only test that proves the gate can block.
// @MX:REASON: Source-level assertions cannot distinguish "checked and passed" from "checked
// nothing"; both emit exit 0 with no output. Only an executed control experiment separates them.
package template_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// cppGateRequire skips the test unless the tools the gate shells out to are
// present. The gate is a POSIX shell script, so Windows is out of scope.
func cppGateRequire(t *testing.T) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("sync-phase-quality-gate.sh is a POSIX shell script")
	}
	for _, tool := range []string{"git", "bash", "g++"} {
		if _, err := exec.LookPath(tool); err != nil {
			t.Skipf("%s not on PATH", tool)
		}
	}
}

// cppGateRunFixture builds a one-commit git repository containing files, then
// runs the gate against it and returns its stdout.
//
// The commit subject must match the gate's sync-phase detection, and the commit
// must carry a C++ code-file delta — otherwise the gate exits early and the test
// would pass while measuring nothing.
func cppGateRunFixture(t *testing.T, files map[string]string) string {
	t.Helper()
	_, out := cppGateRunFixtureIn(t, files)
	return out
}

// cppGateRunFixtureIn is cppGateRunFixture plus the fixture directory, for the
// cases that need to inspect what the gate wrote inside it.
func cppGateRunFixtureIn(t *testing.T, files map[string]string) (string, string) {
	t.Helper()
	dir := t.TempDir()

	for name, body := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}

	git := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, out)
		}
	}
	git("init", "--quiet")
	git("-c", "user.email=t603@example.invalid", "-c", "user.name=t603", "add", ".")
	git("-c", "user.email=t603@example.invalid", "-c", "user.name=t603",
		"commit", "--quiet", "-m", "docs: sync-phase fixture")

	gate := filepath.Join(hocProjectRoot(t), "internal", "template", "templates",
		".claude", "hooks", "moai", "sync-phase-quality-gate.sh")

	cmd := exec.Command("bash", gate)
	cmd.Dir = dir
	// A minimal environment: the two variables below steer the gate's mode, and
	// inheriting the developer's values would make the verdict depend on who ran
	// the test rather than on the code.
	cmd.Env = append(os.Environ(),
		"CLAUDE_PROJECT_DIR="+dir,
		"MOAI_SYNC_GATE_BLOCKING=1",
		"MOAI_AUTONOMY_TIER=",
	)
	out, _ := cmd.CombinedOutput() // the gate always exits 0; its verdict rides stdout
	return dir, string(out)
}

const (
	brokenCPP = "int main() { this is not valid c++ ; }\n"
	validCPP  = "int main() { return 0; }\n"
	cmakeFile = "cmake_minimum_required(VERSION 3.10)\nproject(fixture)\n"
)

// A broken .cpp file MUST block. Before the repair this was the silent case:
// the compiler was never invoked, c1 was 0, and the gate allowed the turn.
func TestSyncGateCpp_BrokenCppBlocks(t *testing.T) {
	cppGateRequire(t)
	out := cppGateRunFixture(t, map[string]string{
		"CMakeLists.txt": cmakeFile,
		"bad.cpp":        brokenCPP,
	})
	if !strings.Contains(out, `"decision":"block"`) {
		t.Errorf("SYNC_GATE_CPP_INERT: a broken .cpp did not block the gate.\ngate output: %q", out)
	}
}

// A broken .cc file MUST block. The compiler DID run here even before the
// repair, and reported the error — but find discarded its exit status, so the
// gate allowed the turn anyway. This case guards the exit-status propagation.
func TestSyncGateCpp_BrokenCcBlocks(t *testing.T) {
	cppGateRequire(t)
	out := cppGateRunFixture(t, map[string]string{
		"CMakeLists.txt": cmakeFile,
		"bad.cc":         brokenCPP,
	})
	if !strings.Contains(out, `"decision":"block"`) {
		t.Errorf("SYNC_GATE_CPP_INERT: a broken .cc did not block the gate.\ngate output: %q", out)
	}
}

// The control that keeps the two cases above honest: a gate that blocked
// unconditionally would satisfy both of them. Valid sources must pass.
func TestSyncGateCpp_ValidSourcePasses(t *testing.T) {
	cppGateRequire(t)
	out := cppGateRunFixture(t, map[string]string{
		"CMakeLists.txt": cmakeFile,
		"ok.cpp":         validCPP,
	})
	if strings.Contains(out, `"decision":"block"`) {
		t.Errorf("valid C++ sources must not block the gate.\ngate output: %q", out)
	}
}

// A C++ project whose commit touched no compilable translation unit (headers
// only) must say so. "Checked nothing" and "checked everything and it passed"
// are both silent exit 0s, and conflating them is what let the original defect
// survive unnoticed — so the zero-target case is reported in words.
func TestSyncGateCpp_ZeroTargetsReportedDistinctly(t *testing.T) {
	cppGateRequire(t)
	dir, out := cppGateRunFixtureIn(t, map[string]string{
		"CMakeLists.txt": cmakeFile,
		"api.h":          "#pragma once\nint f();\n",
	})
	if strings.Contains(out, `"decision":"block"`) {
		t.Errorf("a header-only commit must not block.\ngate output: %q", out)
	}
	auditLog, err := os.ReadFile(filepath.Join(dir, ".moai", "logs", "sync-quality-gate.log"))
	if err != nil {
		t.Fatalf("gate audit log not written: %v", err)
	}
	if !strings.Contains(string(auditLog), "cpp_targets=0") {
		t.Errorf("zero compiled targets must be recorded distinctly from a passing check.\naudit log: %q", auditLog)
	}
}

// Both copies of the hook must stay byte-identical: the local one under
// .claude/ is what this repository runs, the template one is what ships to
// users, and a repair applied to only one of them is a silent half-fix.
func TestSyncGateCpp_LocalAndTemplateCopiesIdentical(t *testing.T) {
	root := hocProjectRoot(t)
	rel := filepath.Join(".claude", "hooks", "moai", "sync-phase-quality-gate.sh")

	local, err := os.ReadFile(filepath.Join(root, rel))
	if err != nil {
		t.Skipf("local hook copy not present: %v", err)
	}
	shipped, err := os.ReadFile(filepath.Join(root, "internal", "template", "templates", rel))
	if err != nil {
		t.Fatalf("template hook copy not present: %v", err)
	}
	if string(local) != string(shipped) {
		t.Error("sync-phase-quality-gate.sh differs between the local and template copies; edit both together")
	}
}
