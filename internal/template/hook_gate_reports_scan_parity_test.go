// hook_gate_reports_scan_parity_test.go: SPEC-CODEX-GATE-SCOPING-001
// card-review repair round 2 — the sync gate's find-based checker sweeps must
// exclude the same parked .moai/reports tree its content key excludes.
//
// The M4 exclusion removed .moai/reports from worktree_content_id (the cache
// key), but the find-based checkers (Ruby/PHP/Java/Kotlin/C++/Scala/R) still
// sweep it: a broken fixture parked there fails the checker, the gate records
// the failure keyed WITHOUT that tree, and fixing the fixture then changes
// neither HEAD nor the key — the stored failure replays stale-fail forever.
// The sweep arms must prune .moai/reports so the checks read the same trees
// the key digests.
//
// The real ruby is replaced by a stub on PATH (the reportsGateStub precedent)
// that fails exactly when the file it is asked to check carries a BROKEN
// marker token — the refusal keys on the fixture's brokenness, never on a
// path substring. The observation is two gate runs against one fixture: run 1
// with the broken parked fixture (must not block, and the stub must have
// checked the root file), then the fixture fixed (run 2 must still not block —
// pre-repair the stored failure replays here).
//
// Sentinel on failure: SYNC_GATE_REPORTS_SCAN_PARITY
package template_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// rubyScanStub logs every invocation and fails (exit 1) exactly when a file
// argument carries the broken marker token.
const rubyScanStub = `#!/bin/sh
echo "stub ruby $*" >> "$STUB_LOG"
for a in "$@"; do
    case "$a" in
        -c) ;;
        *) if [ -f "$a" ] && grep -q SYNC_GATE_R2_BROKEN "$a" 2>/dev/null; then
               echo "stub ruby: syntax error in $a" >&2
               exit 1
           fi ;;
    esac
done
exit 0
`

// rubyScanRun builds a two-commit sync-phase Ruby fixture carrying a parked
// broken file under .moai/reports/lab, runs the gate script once, optionally
// FIXES the parked fixture (removing only its brokenness — a change the key
// must be indifferent to, since the key excludes reports), runs the gate
// again, and returns both stdouts and the stub log.
func rubyScanRun(t *testing.T, script string, fixLab bool) (string, string, string) {
	t.Helper()
	dir := t.TempDir()
	bin := t.TempDir()
	stubLog := filepath.Join(t.TempDir(), "stub.log")

	if err := os.WriteFile(filepath.Join(bin, "ruby"), []byte(rubyScanStub), 0o755); err != nil {
		t.Fatalf("write stub ruby: %v", err)
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
	write("Gemfile", "source 'https://rubygems.org'\n")
	write(".moai/reports/lab/broken.rb", "SYNC_GATE_R2_BROKEN = true\ndef broken(; end\n")

	git := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", append([]string{"-c", "user.email=t1404@example.invalid", "-c", "user.name=t1404"}, args...)...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, out)
		}
	}
	git("init", "--quiet")
	git("add", "Gemfile")
	git("commit", "--quiet", "-m", "docs: sync-phase scan parity fixture")
	// The sync commit's own diff carries the root Ruby change, so the ruby
	// checker runs even though the parked fixture is untracked.
	write("app.rb", "puts 'ok'\n")
	git("add", "app.rb")
	git("commit", "--quiet", "-m", "docs: sync-phase ruby check input")

	run := func() string {
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
		return string(out)
	}
	out1 := run()
	if fixLab {
		// The repair lands: the parked fixture becomes valid. Only its content
		// under .moai/reports changes — neither HEAD nor the content key moves.
		write(".moai/reports/lab/broken.rb", "puts 'fixed'\n")
	}
	out2 := run()
	log, _ := os.ReadFile(stubLog)
	return out1, out2, string(log)
}

// TestSyncPhaseGateCheckerSweepParityWithReportsKey pins the round-2 N2
// contract on the deployed hook AND the template mirror: the find-based
// checker sweep reads no parked .moai/reports fixture (run 1 does not block,
// and the stub checked the root file rather than the parked one), so fixing
// the fixture cannot replay a stored failure (run 2 does not block either).
func TestSyncPhaseGateCheckerSweepParityWithReportsKey(t *testing.T) {
	reportsGateRequire(t)
	root := hocProjectRoot(t)
	mirror := filepath.Join(root, "internal", "template", "templates",
		".claude", "hooks", "moai", "sync-phase-quality-gate.sh")
	deployed := filepath.Join(root, ".claude", "hooks", "moai", "sync-phase-quality-gate.sh")

	t.Run("deployed sweep never reads the parked fixture", func(t *testing.T) {
		out1, out2, stub := rubyScanRun(t, deployed, true)
		if !strings.Contains(stub, "app.rb") {
			t.Fatalf("premise: the stub ruby was never handed the root app.rb — the ruby check never ran.\nstub log: %q", stub)
		}
		if strings.Contains(stub, ".moai/reports/lab/broken.rb") {
			t.Errorf("SYNC_GATE_REPORTS_SCAN_PARITY: the checker sweep read the parked reports fixture — a verdict the content key cannot invalidate.\nstub log: %q", stub)
		}
		if strings.Contains(out1, `"decision":"block"`) {
			t.Errorf("SYNC_GATE_REPORTS_SCAN_PARITY: run 1 blocked on the parked fixture the key excludes.\nrun1: %q", out1)
		}
		if strings.Contains(out2, `"decision":"block"`) {
			t.Errorf("SYNC_GATE_REPORTS_SCAN_PARITY: run 2 replayed a stored failure after the fixture was fixed — the sweep and the key disagree.\nrun2: %q", out2)
		}
	})

	t.Run("mirror carries the same sweep parity", func(t *testing.T) {
		out1, out2, stub := rubyScanRun(t, mirror, true)
		if !strings.Contains(stub, "app.rb") {
			t.Fatalf("premise: the stub ruby was never handed the root app.rb — the ruby check never ran.\nstub log: %q", stub)
		}
		if strings.Contains(stub, ".moai/reports/lab/broken.rb") {
			t.Errorf("SYNC_GATE_REPORTS_SCAN_PARITY: the template mirror's sweep reads the parked fixture too.\nstub log: %q", stub)
		}
		if strings.Contains(out1, `"decision":"block"`) || strings.Contains(out2, `"decision":"block"`) {
			t.Errorf("SYNC_GATE_REPORTS_SCAN_PARITY: the template mirror blocked across the fix-the-fixture sequence.\nrun1: %q\nrun2: %q", out1, out2)
		}
	})

	t.Run("broken root ruby still gates", func(t *testing.T) {
		// The prune must not defang the checker: a BROKEN root file (a path no
		// exclusion names) still fails the gate, observed through the stub
		// being handed that file and the block emitted.
		dir := t.TempDir()
		bin := t.TempDir()
		stubLog := filepath.Join(t.TempDir(), "stub.log")
		if err := os.WriteFile(filepath.Join(bin, "ruby"), []byte(rubyScanStub), 0o755); err != nil {
			t.Fatalf("write stub ruby: %v", err)
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
		write("Gemfile", "source 'https://rubygems.org'\n")
		git := func(args ...string) {
			t.Helper()
			cmd := exec.Command("git", append([]string{"-c", "user.email=t1404@example.invalid", "-c", "user.name=t1404"}, args...)...)
			cmd.Dir = dir
			if out, err := cmd.CombinedOutput(); err != nil {
				t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, out)
			}
		}
		git("init", "--quiet")
		git("add", "Gemfile")
		git("commit", "--quiet", "-m", "docs: sync-phase scan parity control")
		write("app.rb", "SYNC_GATE_R2_BROKEN = true\ndef broken(; end\n")
		git("add", "app.rb")
		git("commit", "--quiet", "-m", "docs: sync-phase ruby check input")

		cmd := exec.Command("bash", deployed)
		cmd.Dir = dir
		cmd.Env = append(os.Environ(),
			"PATH="+bin+string(os.PathListSeparator)+os.Getenv("PATH"),
			"CLAUDE_PROJECT_DIR="+dir,
			"MOAI_SYNC_GATE_BLOCKING=1",
			"MOAI_AUTONOMY_TIER=",
			"STUB_LOG="+stubLog,
		)
		out, _ := cmd.CombinedOutput()
		log, _ := os.ReadFile(stubLog)
		if !strings.Contains(string(log), "app.rb") {
			t.Fatalf("premise: the stub ruby was never handed the broken root app.rb.\nstub log: %q", string(log))
		}
		if !strings.Contains(string(out), `"decision":"block"`) {
			t.Errorf("SYNC_GATE_REPORTS_SCAN_PARITY control: a broken non-reports Ruby change must still gate, out: %q", string(out))
		}
	})
}
