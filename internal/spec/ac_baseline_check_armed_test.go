// ac_baseline_check_armed_test.go — regression tests for
// scripts/ac-baseline/check-armed.sh, the read-only report that says when the
// commit-time AC-snapshot guard is not armed (cards t1161, t1197, t1206).
//
// Every scenario runs in a throwaway repo from newGuardRepo with the isolated
// git environment of guardEnv; nothing here reads or writes the repository's
// own git config. The installer is copied in and run only inside that repo.
package spec

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const (
	checkArmedRel   = "scripts/ac-baseline/check-armed.sh"
	checkArmedKey   = "hook.ac-baseline-guard.command"
	checkArmedStale = "ac-baseline-guard: STALE:"

	// preFailOpenCommand is the command the first installer (before the checker
	// learned to fail open on its own death) wrote. A config still carrying it
	// runs a real guard, so only a byte comparison can tell it is out of date.
	preFailOpenCommand = `if [ -f scripts/ac-baseline/check-staged.sh ]; then sh scripts/ac-baseline/check-staged.sh; else echo "ac-baseline-guard: NOT CHECKED (scripts/ac-baseline/check-staged.sh absent in this tree; it is armed once the tree absorbs develop)" >&2; fi`
)

// armedRepo returns a scratch repo holding the three guard scripts, armed by
// the copied installer.
func armedRepo(t *testing.T) *guardRepo {
	t.Helper()
	requireRealCommitGit(t)
	r := newGuardRepo(t)
	for _, rel := range []string{guardCheckerRel, guardInstallRel, checkArmedRel} {
		raw, err := os.ReadFile(filepath.Join(repoRoot(t), rel))
		if err != nil {
			t.Fatalf("read %s: %v", rel, err)
		}
		r.write(rel, string(raw))
	}
	if _, errb, code := r.runIn(r.dir, nil, "sh", guardInstallRel); code != 0 {
		t.Fatalf("install: exit=%d stderr=%q", code, errb)
	}
	return r
}

// checkArmed runs the report from dir and returns stdout (trimmed), the
// non-empty stderr lines, and the exit code.
func (r *guardRepo) checkArmed(dir string) (string, []string, int) {
	r.t.Helper()
	out, errb, code := r.runIn(dir, nil, "sh", filepath.Join(r.dir, filepath.FromSlash(checkArmedRel)))
	return strings.TrimSpace(out), nonEmptyLines(errb), code
}

func TestACBaselineCheckArmed(t *testing.T) {
	// wantArmed and wantStale pin both halves of the output contract: the
	// stdout marker AND the stderr lines, so a report that says ARMED while
	// also warning (or warns about the wrong thing) cannot pass.
	wantArmed := func(t *testing.T, out string, errs []string, code int) {
		t.Helper()
		if code != 0 || out != "ARMED" || len(errs) != 0 {
			t.Fatalf("want ARMED, exit 0, no stderr; got out=%q exit=%d stderr=%q", out, code, errs)
		}
	}
	wantStale := func(t *testing.T, out string, errs []string, code int) {
		t.Helper()
		if code != 0 || out != "NOT-ARMED 1" || len(errs) != 1 || !strings.HasPrefix(errs[0], checkArmedStale) {
			t.Fatalf("want one %q line and NOT-ARMED 1, exit 0; got out=%q exit=%d stderr=%q", checkArmedStale, out, code, errs)
		}
	}

	t.Run("armed", func(t *testing.T) {
		r := armedRepo(t)
		out, errs, code := r.checkArmed(r.dir)
		wantArmed(t, out, errs, code)
	})

	t.Run("stale/command_true", func(t *testing.T) {
		r := armedRepo(t)
		r.git("config", "--local", checkArmedKey, "true")
		out, errs, code := r.checkArmed(r.dir)
		wantStale(t, out, errs, code)
	})

	t.Run("stale/pre_fail_open_command", func(t *testing.T) {
		r := armedRepo(t)
		r.git("config", "--local", checkArmedKey, preFailOpenCommand)
		out, errs, code := r.checkArmed(r.dir)
		wantStale(t, out, errs, code)
	})

	// t1197 F1: "byte identical" must include trailing bytes. $(...) strips
	// every trailing newline, so without care a value with an extra newline
	// reads as the installer's command.
	t.Run("stale/trailing_newline", func(t *testing.T) {
		r := armedRepo(t)
		want := r.git("config", "--get", checkArmedKey)
		r.git("config", "--local", checkArmedKey, want+"\n")
		out, errs, code := r.checkArmed(r.dir)
		wantStale(t, out, errs, code)
	})

	t.Run("repair/reinstall_clears_stale", func(t *testing.T) {
		r := armedRepo(t)
		r.git("config", "--local", checkArmedKey, "true")
		if _, errb, code := r.runIn(r.dir, nil, "sh", guardInstallRel); code != 0 {
			t.Fatalf("reinstall: exit=%d stderr=%q", code, errb)
		}
		out, errs, code := r.checkArmed(r.dir)
		wantArmed(t, out, errs, code)
	})

	t.Run("installer_unreadable", func(t *testing.T) {
		r := armedRepo(t)
		if err := os.Remove(filepath.Join(r.dir, filepath.FromSlash(guardInstallRel))); err != nil {
			t.Fatalf("remove installer: %v", err)
		}
		out, errs, code := r.checkArmed(r.dir)
		if code != 0 || out != "NOT-ARMED 1" || len(errs) != 1 || !strings.Contains(errs[0], "cannot read the expected hook command") {
			t.Fatalf("want one cannot-read line and NOT-ARMED 1; got out=%q exit=%d stderr=%q", out, code, errs)
		}
	})

	t.Run("keys_unset", func(t *testing.T) {
		r := armedRepo(t)
		r.git("config", "--local", "--unset", checkArmedKey)
		r.git("config", "--local", "--unset", "hook.ac-baseline-guard.event")
		out, errs, code := r.checkArmed(r.dir)
		if code != 0 || out != "NOT-ARMED 2" || len(errs) != 2 {
			t.Fatalf("want two NOT ARMED lines; got out=%q exit=%d stderr=%q", out, code, errs)
		}
		for _, e := range errs {
			if !strings.Contains(e, "NOT ARMED:") || !strings.Contains(e, "is unset") {
				t.Fatalf("unexpected line %q", e)
			}
		}
	})

	// Without a work tree the byte comparison has no installer to read, so it
	// is skipped exactly as the checker-existence step is (t1197 F2).
	t.Run("no_work_tree_skips_comparison", func(t *testing.T) {
		r := armedRepo(t)
		r.git("config", "--local", checkArmedKey, "true")
		out, errs, code := r.checkArmed(filepath.Join(r.dir, ".git"))
		wantArmed(t, out, errs, code)
	})

	t.Run("not_a_git_repository", func(t *testing.T) {
		r := armedRepo(t)
		out, errs, code := r.checkArmed(t.TempDir())
		wantArmed(t, out, errs, code)
	})
}
