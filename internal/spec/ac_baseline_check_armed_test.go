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

	// A same-length substitution defeats any comparison weaker than byte
	// equality (length-only, prefix-only) — t1206 audit D2.
	t.Run("stale/same_length_substitution", func(t *testing.T) {
		r := armedRepo(t)
		want := r.git("config", "--get", checkArmedKey)
		mutated := strings.Replace(want, "check-staged.sh", "check-stagex.sh", 1)
		if mutated == want || len(mutated) != len(want) {
			t.Fatalf("fixture premise: substitution must change bytes and keep length")
		}
		r.git("config", "--local", checkArmedKey, mutated)
		out, errs, code := r.checkArmed(r.dir)
		wantStale(t, out, errs, code)
	})

	// The installer is resolved against the work-tree top, not the caller's
	// cwd; a SessionStart hook does not promise to start at the top (D3).
	t.Run("stale/from_subdirectory", func(t *testing.T) {
		r := armedRepo(t)
		r.git("config", "--local", checkArmedKey, "true")
		out, errs, code := r.checkArmed(filepath.Join(r.dir, "scripts"))
		wantStale(t, out, errs, code)
	})

	// An empty command value is an unarmed guard wherever the report runs
	// from, including without a work tree (t1206 audit D1).
	for _, where := range []string{"top", "git_dir"} {
		t.Run("empty_command/"+where, func(t *testing.T) {
			r := armedRepo(t)
			r.git("config", "--local", checkArmedKey, "")
			dir := r.dir
			if where == "git_dir" {
				dir = filepath.Join(r.dir, ".git")
			}
			out, errs, code := r.checkArmed(dir)
			if code != 0 || out != "NOT-ARMED 1" || len(errs) != 1 || !strings.Contains(errs[0], "command is unset") {
				t.Fatalf("want one command-unset line and NOT-ARMED 1; got out=%q exit=%d stderr=%q", out, code, errs)
			}
		})
	}

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
