// hook_gate_reports_exclude_test.go: SPEC-CODEX-GATE-SCOPING-001 plan-phase
// RED observation for Facet 3 — a parked Go fixture under .moai/reports joins
// the sync gate's delta collection and drives a spurious block.
//
// The gate's WCI_EXCLUDES set excludes state/log/worktree trees and dependency
// dirs, but not .moai/reports: a Go module parked there (an audit lab with its
// own go.mod) enters SYNC_DELTA_FILES and becomes a vetted GO_ROOT, so a
// docs-adjacent sync turn blocks on a fixture no lane owns (the measured
// disposition record #5 of card t1395). The test pins the post-fix behavior —
// no block, and the stub log carries no invocation naming the reports module —
// on BOTH delta arms that can carry the fixture in (the untracked walk and the
// tracked/committed diff), plus the control that a broken file OUTSIDE
// .moai/reports still gates (REQ-CGSC-010).
//
// The real toolchain is replaced by a stub on PATH (the
// hook_gate_exit_status_test.go precedent) so the observation is hermetic: the
// stub fails exactly when it is invoked against a module root under "reports",
// so a block whose reason names the stub IS the observation that the gate
// reached into .moai/reports.
//
// Authored BEFORE the fix on tree 2de0a2cb613b04765a1554f86685a3b48e0be806
// (branch WT-codex-gate-scope); observed RED there. manager-develop's M4 turns
// it GREEN (template mirror first, make build, then the tracked twin in the
// same commit).
//
// Sentinel on failure: SYNC_GATE_REPORTS_SWEEP
package template_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// reportsGateStub fails only when invoked against a module root under
// "reports" — everything else (the fixture root module) passes.
const reportsGateStub = `#!/bin/sh
echo "stub go $*" >> "$STUB_LOG"
for a in "$@"; do
    case "$a" in *reports*) echo "stub go: refusing module under $a" >&2; exit 7 ;; esac
done
exit 0
`

func reportsGateRequire(t *testing.T) {
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

// reportsGateRun builds a one-commit sync-phase Go fixture carrying a parked
// module under .moai/reports/lab, runs one gate script against it with the stub
// go on PATH, and returns stdout and the stub log.
//
// tracked: when true the reports fixture is COMMITTED (the sync commit's own
// diff carries it — the ①/② delta arms); when false it stays untracked (the ③
// walk arm). goOnly: when true the fixture carries no root .go file, so the
// only delta Go file lives under .moai/reports — the parked-fixture shape.
func reportsGateRun(t *testing.T, script string, tracked, goOnly bool) (string, string) {
	t.Helper()
	dir := t.TempDir()
	bin := t.TempDir()
	stubLog := filepath.Join(t.TempDir(), "stub.log")

	if err := os.WriteFile(filepath.Join(bin, "go"), []byte(reportsGateStub), 0o755); err != nil {
		t.Fatalf("write stub go: %v", err)
	}
	write := func(rel, body string) {
		t.Helper()
		p := filepath.Join(dir, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", filepath.Dir(p), err)
		}
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatalf("write %s: %v", rel, err)
		}
	}
	write("go.mod", "module example.com/reportsfixture\n\ngo 1.22\n")
	if !goOnly {
		write("main.go", "package main\n\nfunc main() {}\n")
	}
	write(".moai/reports/lab/go.mod", "module example.com/reports-lab\n\ngo 1.22\n")
	write(".moai/reports/lab/broken.go", "package lab\n\nfunc F() { thisIsNotDefined() }\n")

	git := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", append([]string{"-c", "user.email=t1404@example.invalid", "-c", "user.name=t1404"}, args...)...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, out)
		}
	}
	git("init", "--quiet")
	if tracked {
		git("add", ".")
	} else {
		git("add", "go.mod")
		if !goOnly {
			git("add", "main.go")
		}
	}
	git("commit", "--quiet", "-m", "docs: sync-phase fixture")

	cmd := exec.Command("bash", script)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(),
		"PATH="+bin+string(os.PathListSeparator)+os.Getenv("PATH"),
		"CLAUDE_PROJECT_DIR="+dir,
		"MOAI_SYNC_GATE_BLOCKING=1",
		"MOAI_AUTONOMY_TIER=",
		"STUB_LOG="+stubLog,
	)
	out, _ := cmd.CombinedOutput() // the gate always exits 0; its verdict rides stdout
	log, _ := os.ReadFile(stubLog)
	return string(out), string(log)
}

// TestSyncPhaseGateExcludesReportsGoFixture pins REQ-CGSC-009: a parked Go
// fixture under .moai/reports must reach neither the delta collection nor the
// vetted module roots, on the untracked arm AND the tracked arm, in the
// template mirror AND the deployed twin; and the control proves a broken file
// outside .moai/reports still gates.
//
// RED on the pre-fix tree: every reports arm BLOCKs, and the stub log names the
// go invocation against .moai/reports/lab — the sweep, observed.
func TestSyncPhaseGateExcludesReportsGoFixture(t *testing.T) {
	reportsGateRequire(t)
	root := hocProjectRoot(t)
	mirror := filepath.Join(root, "internal", "template", "templates",
		".claude", "hooks", "moai", "sync-phase-quality-gate.sh")
	deployed := filepath.Join(root, ".claude", "hooks", "moai", "sync-phase-quality-gate.sh")

	t.Run("untracked reports fixture is not collected", func(t *testing.T) {
		out, stub := reportsGateRun(t, deployed, false, true)
		if !strings.Contains(out, `"decision":"block"`) {
			return // the post-fix expectation; the RED face below records today's sweep
		}
		t.Errorf("SYNC_GATE_REPORTS_SWEEP: an untracked Go fixture under .moai/reports drove a spurious block.\nstub log: %q\nout: %q", stub, out)
	})
	t.Run("tracked reports fixture is not collected", func(t *testing.T) {
		out, stub := reportsGateRun(t, deployed, true, true)
		if !strings.Contains(out, `"decision":"block"`) {
			return
		}
		t.Errorf("SYNC_GATE_REPORTS_SWEEP: a committed Go fixture under .moai/reports drove a spurious block.\nstub log: %q\nout: %q", stub, out)
	})
	t.Run("mirror carries the same exclusion", func(t *testing.T) {
		out, stub := reportsGateRun(t, mirror, false, true)
		if !strings.Contains(out, `"decision":"block"`) {
			return
		}
		t.Errorf("SYNC_GATE_REPORTS_SWEEP: the template mirror sweeps .moai/reports too.\nstub log: %q\nout: %q", stub, out)
	})
	t.Run("root fixture still gates", func(t *testing.T) {
		out, stub := reportsGateRun(t, deployed, true, false)
		if !strings.Contains(stub, "stub go") || !strings.Contains(stub, "vet") {
			t.Fatalf("premise: the stub go was never invoked for the root delta.\nstub log: %q", stub)
		}
		if !strings.Contains(out, `"decision":"block"`) {
			t.Errorf("SYNC_GATE_REPORTS_SWEEP control: a broken root-level Go change must still block, out: %q", out)
		}
	})
}
