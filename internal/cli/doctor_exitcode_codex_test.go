package cli

// t508 (SPEC-CODEX-ENABLED-FATAL-001) — the PROCESS-level exit-code contract
// for the Codex Wiring check's fatal grade (AC-CEF-008, AC-CEF-009,
// AC-CEF-014).
//
// [HARD] These cases read a REAL process exit status. Asserting that
// doctorExitStatus(1) returns a non-nil *exitCodeError — the form
// internal/cli/exitcode_contract_test.go:30,38 and
// internal/cli/binary_lag_test.go:102 already take in-package — does NOT
// satisfy them: a function's return value is a hypothesis about what the
// process would do, and the callers this contract exists for (hooks, CI
// wrappers) read the process's status, not the function's return.
//
// Authored BEFORE the severity axis lands, so its RED is observed against the
// pre-change tree (plan.md §F M4 item 1).

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/cli/uikit"
	"github.com/modu-ai/moai-adk/internal/codexwiring"
)

// buildMoaiForDoctorExit builds the moai binary under the test's own temp dir
// and returns its path.
//
// The package path is given as an IMPORT path rather than a relative one, so
// the build does not depend on how far `internal/cli` sits from the module
// root — a relative "./cmd/moai" would be wrong the moment this file moved.
func buildMoaiForDoctorExit(t *testing.T) string {
	t.Helper()
	bin := filepath.Join(t.TempDir(), "moai")
	if runtime.GOOS == "windows" {
		bin += ".exe"
	}
	cmd := exec.Command("go", "build", "-o", bin, "github.com/modu-ai/moai-adk/cmd/moai")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("go build moai: %v\n%s", err, out)
	}
	return bin
}

// runDoctorProcess executes `moai doctor` in projRoot with CODEX_HOME pinned
// to codexHome, and returns the OBSERVED process exit status plus the combined
// output.
//
// The inherited environment is filtered rather than merely appended to: a
// developer with CODEX_HOME already exported would otherwise leave two
// entries in the child's environment, and which one wins is a libc detail no
// test should rest on.
func runDoctorProcess(t *testing.T, bin, projRoot, codexHome string) (int, string) {
	t.Helper()
	env := make([]string, 0, len(os.Environ())+1)
	for _, kv := range os.Environ() {
		if strings.HasPrefix(kv, codexHomeEnvVar+"=") {
			continue
		}
		env = append(env, kv)
	}
	env = append(env, codexHomeEnvVar+"="+codexHome)

	cmd := exec.Command(bin, "doctor")
	cmd.Dir = projRoot
	cmd.Env = env
	out, err := cmd.CombinedOutput()
	if err == nil {
		return 0, string(out)
	}
	var ee *exec.ExitError
	if errors.As(err, &ee) {
		return ee.ExitCode(), string(out)
	}
	t.Fatalf("running %s doctor: %v\n%s", bin, err, out)
	return -1, ""
}

// codexHomeDirOf turns the home root writeCodexHomeConfig returns into the
// CODEX_HOME value the resolver expects — the directory HOLDING config.toml,
// not its parent.
func codexHomeDirOf(home string) string { return filepath.Join(home, codexHomeDirName) }

// TestDoctorExitCode_CodexEnabledFatal — a config carrying the fatal shape
// (an entry whose path exists but which declares no `enabled` key) must make
// the `moai doctor` PROCESS exit 1. Codex itself exits 1 on this config; a
// doctor that exits 0 tells every CI wrapper the machine is fine.
func TestDoctorExitCode_CodexEnabledFatal(t *testing.T) {
	bin := buildMoaiForDoctorExit(t)
	root := wireProjectForDoctor(t)
	home := writeCodexHomeConfig(t, []codexSkillEntrySpec{
		{Path: liveSkillFile(t)}, // no EnabledKey: the fatal shape
	})

	code, out := runDoctorProcess(t, bin, root, codexHomeDirOf(home))

	if code != 1 {
		t.Errorf("moai doctor exit status = %d, want 1 — codex cannot start on this config\n%s", code, out)
	}
}

// TestDoctorExitCode_CodexCleanStaysZero is AC-CEF-008's paired control: the
// same wiring with a bare boolean `enabled` must keep the process at 0.
// Without it, a doctor that exits 1 unconditionally passes the case above.
func TestDoctorExitCode_CodexCleanStaysZero(t *testing.T) {
	bin := buildMoaiForDoctorExit(t)
	root := wireProjectForDoctor(t)
	home := writeCodexHomeConfig(t, []codexSkillEntrySpec{
		{Path: liveSkillFile(t), EnabledKey: "true"},
	})

	code, out := runDoctorProcess(t, bin, root, codexHomeDirOf(home))

	if code != 0 {
		t.Errorf("moai doctor exit status = %d, want 0 on a clean config\n%s", code, out)
	}
}

// TestDoctorExitCode_CodexAdvisoryOnlyStaysZero (AC-CEF-014) is the case
// AC-CEF-009 cannot cover. A clean fixture produces len(problems) == 0 and
// takes the CheckOK branch, exercising NO warn finding at all — so a severity
// fold that re-graded every advisory to fatal would still pass it, and would
// still exit 1 on every real advisory machine.
//
// This fixture produces a non-`enabled` advisory finding (hooks.json removed
// from a wired project, a `problems = append(...)` site this SPEC does not
// modify) and no fatal one, then asserts BOTH halves: the in-package status is
// CheckWarn, and the process exits 0.
func TestDoctorExitCode_CodexAdvisoryOnlyStaysZero(t *testing.T) {
	bin := buildMoaiForDoctorExit(t)
	root := wireProjectForDoctor(t)
	if err := os.Remove(filepath.Join(root, codexwiring.HooksRelPath)); err != nil {
		t.Fatalf("removing hooks.json from the wired fixture: %v", err)
	}
	home := writeCodexHomeConfig(t, []codexSkillEntrySpec{
		{Path: liveSkillFile(t), EnabledKey: "false"},
	})

	// In-package half: warn, not OK and not Fail, and the advisory finding's
	// own text survives into the message-plus-detail.
	stubCodexLookup(t, true, true)
	stubCodexHome(t, home)
	check := checkCodexWiring(root, false)
	if check.Status != uikit.CheckWarn {
		t.Errorf("status = %v, want CheckWarn on an advisory-only config: %+v", check.Status, check)
	}
	if text := check.Message + " " + codexDetailText(check); !strings.Contains(text, "hooks.json missing") {
		t.Errorf("the advisory finding's text did not survive: %+v", check)
	}

	// Process half: exit 0.
	code, out := runDoctorProcess(t, bin, root, codexHomeDirOf(home))
	if code != 0 {
		t.Errorf("moai doctor exit status = %d, want 0 on an advisory-only config\n%s", code, out)
	}
}
