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
// on the untracked arm (③) AND the tracked arms (① the sync commit's diff and
// ② the tracked uncommitted diff), in the deployed hook AND the template
// mirror; and the control proves a broken file OUTSIDE .moai/reports still
// gates (REQ-CGSC-010).
//
// The real toolchain is replaced by a stub on PATH (the
// hook_gate_exit_status_test.go precedent) so the observation is hermetic. The
// stub refuses a module root carrying a BROKEN marker file — the parked lab
// and the control's root fixture each carry one — so the refusal keys on the
// fixture's brokenness, never on a path substring: a control module whose
// path merely resembles reports stays vetted, and the exclusion under test
// stays the only thing that can silence the lab (plan-audit iteration-2 D6:
// the previous substring-keyed stub made the old control impossible under the
// correct fix, because the only block source was the reports sweep itself).
//
// Authored BEFORE the fix on tree 2de0a2cb613b04765a1554f86685a3b48e0be806
// (branch WT-codex-gate-scope) and observed RED there; revised at HEAD
// 9ef1cbedc3082f97ac74cf6162bc1e2fd0a55324 per plan-audit iteration-2 (D6:
// control redefined onto a non-reports broken change; D8-③: the tracked
// fixture now carries real ① and ② input — a two-commit fixture plus an
// uncommitted modification of a committed reports file) and re-observed RED
// on the revised file. manager-develop's M4 turns it GREEN (template mirror
// first, make build, then the tracked twin in the same commit).
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

// reportsGateStub logs every invocation and fails (exit 7) exactly when the
// vetted module root — the directory named by go's -C flag — carries a BROKEN
// marker file. Everything else passes.
const reportsGateStub = `#!/bin/sh
echo "stub go $*" >> "$STUB_LOG"
dir=""
prev=""
for a in "$@"; do
    if [ "$prev" = "-C" ]; then dir="$a"; fi
    prev="$a"
done
if [ -n "$dir" ] && [ -f "$dir/BROKEN" ]; then
    echo "stub go: refusing broken module at $dir" >&2
    exit 7
fi
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

// reportsGateRun builds a two-commit sync-phase Go fixture carrying a parked
// module under .moai/reports/lab, runs one gate script against it with the stub
// go on PATH, and returns stdout and the stub log.
//
// tracked: when true the reports fixture is committed as a SECOND commit — the
// sync commit's own diff (①, HEAD~1..HEAD) carries it — and a committed
// reports file is then modified without a further commit, so the tracked
// uncommitted diff (②, git diff HEAD) carries it too (plan-audit iteration-2
// D8-③: an all-committed fixture left arm ② empty and exercised only ①).
// When false the reports fixture stays untracked (the ③ walk arm).
//
// goOnly: when true the fixture carries no root .go file, so the only delta Go
// file lives under .moai/reports — the parked-fixture shape.
//
// rootBroken: when true an UNTRACKED broken Go file is written at the repo
// root — a non-reports change no M4 exclusion names, so the delta keeps a
// root module to vet. The control's premise and its expected block both ride
// this file (plan-audit iteration-2 D6).
func reportsGateRun(t *testing.T, script string, tracked, goOnly, rootBroken bool) (string, string) {
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
	if rootBroken {
		// The marker is what the stub refuses on, so the control's root module
		// blocks under the CORRECT fix too — the block must not depend on the
		// reports sweep surviving (plan-audit iteration-2 D6).
		write("BROKEN", "the root module's working-tree change does not compile\n")
		write("broken_extra.go", "package main\n\nfunc Extra() { thisIsBrokenToo() }\n")
	}
	write(".moai/reports/lab/go.mod", "module example.com/reports-lab\n\ngo 1.22\n")
	write(".moai/reports/lab/broken.go", "package lab\n\nfunc F() { thisIsNotDefined() }\n")
	// The parked lab is broken — the marker is what the stub refuses on, so
	// the refusal follows the fixture's brokenness rather than its path text.
	write(".moai/reports/lab/BROKEN", "the parked audit lab does not compile\n")

	git := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", append([]string{"-c", "user.email=t1404@example.invalid", "-c", "user.name=t1404"}, args...)...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, out)
		}
	}
	git("init", "--quiet")
	git("add", "go.mod")
	if !goOnly {
		git("add", "main.go")
	}
	git("commit", "--quiet", "-m", "docs: sync-phase fixture")
	if tracked {
		git("add", ".moai")
		// The gate applies only to a sync-phase-shaped LAST commit subject —
		// commit 2 must carry the shape too or the gate silently passes
		// (observed: the first draft's "docs: park …" subject exited the gate
		// before any collection).
		git("commit", "--quiet", "-m", "docs: sync-phase park audit lab under .moai/reports")
		// ② input: a tracked-but-uncommitted modification of a committed
		// reports file — git diff HEAD now names the reports path.
		write(".moai/reports/lab/broken.go",
			"package lab\n\nfunc F() { thisIsNotDefined() }\n\nfunc G() { thisIsBrokenUncommitted() }\n")
	}

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
// vetted module roots — on the untracked arm and the tracked arms, in the
// template mirror and the deployed twin; and the control proves a broken file
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
		out, stub := reportsGateRun(t, deployed, false, true, false)
		if !strings.Contains(out, `"decision":"block"`) {
			return // the post-fix expectation; the RED face below records today's sweep
		}
		t.Errorf("SYNC_GATE_REPORTS_SWEEP: an untracked Go fixture under .moai/reports drove a spurious block.\nstub log: %q\nout: %q", stub, out)
	})
	t.Run("tracked reports fixture is not collected", func(t *testing.T) {
		out, stub := reportsGateRun(t, deployed, true, true, false)
		if !strings.Contains(out, `"decision":"block"`) {
			return
		}
		t.Errorf("SYNC_GATE_REPORTS_SWEEP: a committed Go fixture under .moai/reports drove a spurious block.\nstub log: %q\nout: %q", stub, out)
	})
	t.Run("mirror carries the same exclusion", func(t *testing.T) {
		out, stub := reportsGateRun(t, mirror, false, true, false)
		if !strings.Contains(out, `"decision":"block"`) {
			return
		}
		t.Errorf("SYNC_GATE_REPORTS_SWEEP: the template mirror sweeps .moai/reports too.\nstub log: %q\nout: %q", stub, out)
	})
	t.Run("mirror carries the tracked arms too", func(t *testing.T) {
		out, stub := reportsGateRun(t, mirror, true, true, false)
		if !strings.Contains(out, `"decision":"block"`) {
			return
		}
		t.Errorf("SYNC_GATE_REPORTS_SWEEP: the template mirror sweeps .moai/reports on the tracked arms too.\nstub log: %q\nout: %q", stub, out)
	})
	t.Run("root fixture still gates", func(t *testing.T) {
		out, stub := reportsGateRun(t, deployed, true, false, true)
		// Premise: the gate vetted the ROOT module — a path no M4 exclusion
		// names — so the block below is the gate gating a non-reports change,
		// not a vacuous pass over a tree it never inspected.
		rootVetted := false
		for _, line := range strings.Split(stub, "\n") {
			if strings.Contains(line, "vet") && !strings.Contains(line, ".moai/reports") {
				rootVetted = true
			}
		}
		if !rootVetted {
			t.Fatalf("premise: the stub go was never invoked to vet the root module (a path the fix does not exclude).\nstub log: %q", stub)
		}
		if !strings.Contains(out, `"decision":"block"`) {
			t.Errorf("SYNC_GATE_REPORTS_SWEEP control: a broken non-reports Go change must still gate, out: %q", out)
		}
	})
}
