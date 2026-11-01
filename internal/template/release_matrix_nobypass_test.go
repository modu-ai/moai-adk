// release_matrix_nobypass_test.go: the 3-OS matrix gate of scripts/release.sh
// (Validation 9b) has no off switch (card t2012). Every test here runs the real
// script in a scratch repository with the stub gh of release_fixture_test.go.
//
// The release gate is an operator-mandated full verification: no flag, no
// environment variable and no other gate's override turns it off. The only
// bypass a test may use lives in the gh double, which reports a green matrix.
package template_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// rlsRetiredBypass is the flag an earlier draft of the gate offered to switch
// it off. It is spelled by concatenation so that the retired name stays out of
// the production script and of every grep for it; the script must reject it
// as an unknown flag.
const rlsRetiredBypass = "--skip-" + "matrix-run"

// rlsGreenMatrix makes the stub gh report one dispatched, completed, green
// matrix run (3 OS x 6 shards) on sha. This is the test-side double the other
// release-gate tests use to get past Validation 9b without touching the script.
func rlsGreenMatrix(t *testing.T, s *rlsStub, sha string) {
	t.Helper()
	s.setRuns(t, rlsDispatched(t, s, 901, sha, nil))
}

// rlsReleaseGreenMatrix runs the script with a green matrix on the clone's
// current HEAD.
func rlsReleaseGreenMatrix(t *testing.T, r *rlsRepo, s *rlsStub, args ...string) rlsResult {
	t.Helper()
	rlsGreenMatrix(t, s, r.git("rev-parse", "HEAD"))
	return rlsRelease(t, r, s, args...)
}

type rlsRealTagResult struct {
	out        string
	exit       int
	localTag   string
	remoteTags string
}

// rlsRealTagRun runs the script WITHOUT --dry-run, on the path that creates the
// tag. A git shim blocks `git push`, so nothing can leave the scratch clone;
// the harness marker and an affirmative stdin let the script reach the tag.
func rlsRealTagRun(t *testing.T, r *rlsRepo, s *rlsStub, extraEnv []string, args ...string) rlsRealTagResult {
	t.Helper()
	realGit, err := exec.LookPath("git")
	if err != nil {
		t.Skip("git not available")
	}
	shim := t.TempDir()
	rlsWriteScript(t, shim, "git", fmt.Sprintf(`#!/usr/bin/env bash
case "${1:-}" in
  push) echo "git shim: push blocked" >&2; exit 1 ;;
esac
exec %q "$@"
`, realGit))
	stubPath := ""
	for _, kv := range s.env {
		if v, ok := strings.CutPrefix(kv, "PATH="); ok {
			stubPath = v
		}
	}
	env := append(append([]string{}, s.env...), "MOAI_RELEASE_VIA_HARNESS=1", "PATH="+shim+string(os.PathListSeparator)+stubPath)
	env = append(env, extraEnv...)

	script := filepath.Join(findProjectRootForMirrorTest(t), filepath.FromSlash(rlsReleaseScriptRel))
	ctx, cancel := context.WithTimeout(context.Background(), rlsDeadline)
	defer cancel()
	cmd := exec.CommandContext(ctx, "bash", append([]string{script}, args...)...)
	cmd.Dir, cmd.Env, cmd.Stdin = r.work, env, strings.NewReader("y\n")
	out, runErr := cmd.CombinedOutput()
	if ctx.Err() != nil {
		t.Fatalf("release.sh exceeded %s\n%s", rlsDeadline, out)
	}
	res := rlsRealTagResult{out: rlsNorm(string(out))}
	var exitErr *exec.ExitError
	switch {
	case runErr == nil:
	case errors.As(runErr, &exitErr):
		res.exit = exitErr.ExitCode()
	default:
		t.Fatalf("start release.sh: %v", runErr)
	}
	res.localTag = strings.TrimSpace(rlsRun(t, r.work, rlsEnv(), "git", "rev-parse", "--verify", "--quiet", rlsRcTag+"^{commit}").out)
	res.remoteTags = strings.TrimSpace(rlsRun(t, r.work, rlsEnv(), "git", "ls-remote", "--tags", "origin", "refs/tags/"+rlsRcTag).out)
	return res
}

func rlsMatrixScenario(t *testing.T, name string) matrixScenario {
	t.Helper()
	for _, sc := range matrixScenarios() {
		if sc.name == name {
			return sc
		}
	}
	t.Fatalf("matrix scenario %q not found", name)
	return matrixScenario{}
}

// TestReleaseScriptRetiredBypassIsAnUnknownFlag: the flag that once switched
// the gate off is refused like any unknown flag, with no tag anywhere, even
// when the matrix itself is green.
func TestReleaseScriptRetiredBypassIsAnUnknownFlag(t *testing.T) {
	rlsRequireTools(t, "jq")
	t.Run("dry_run", func(t *testing.T) {
		r := rlsBuildReleaseRepo(t, rlsRcChangelog)
		s := rlsNewStub(t)
		rlsGreenMatrix(t, s, r.git("rev-parse", "HEAD"))
		res := rlsRelease(t, r, s, rlsRcTag, "--dry-run", rlsRetiredBypass)
		out := rlsNorm(res.out)
		if res.exit != 1 {
			t.Errorf("exit = %d, want 1\n%s", res.exit, out)
		}
		rlsMustContain(t, out, "Unknown flag: "+rlsRetiredBypass)
	})
	t.Run("real_tag_path", func(t *testing.T) {
		r := rlsBuildReleaseRepo(t, rlsRcChangelog)
		s := rlsNewStub(t)
		rlsGreenMatrix(t, s, r.git("rev-parse", "HEAD"))
		res := rlsRealTagRun(t, r, s, nil, rlsRcTag, rlsRetiredBypass)
		if res.exit != 1 {
			t.Errorf("exit = %d, want 1\n%s", res.exit, res.out)
		}
		rlsMustContain(t, res.out, "Unknown flag: "+rlsRetiredBypass)
		if res.localTag != "" || res.remoteTags != "" {
			t.Errorf("a tag exists after the refusal: local=%q remote=%q", res.localTag, res.remoteTags)
		}
	})
}

// TestReleaseScriptRealTagPathRefusesWithoutAGreenMatrix: on the path that
// creates the tag (no --dry-run), a missing, red or wrong-SHA matrix refuses
// before any tag, local or remote. The green control proves the same harness
// does reach tag creation, so the refusals are not vacuous.
func TestReleaseScriptRealTagPathRefusesWithoutAGreenMatrix(t *testing.T) {
	rlsRequireTools(t, "jq")
	for _, name := range []string{"b_no_run", "c_green_run_on_another_sha", "d_failed_run_on_the_target_sha"} {
		sc := rlsMatrixScenario(t, name)
		t.Run(name, func(t *testing.T) {
			r := rlsBuildReleaseRepo(t, rlsRcChangelog)
			s := rlsNewStub(t)
			s.setRuns(t, sc.build(t, s, r.git("rev-parse", "HEAD")))
			res := rlsRealTagRun(t, r, s, nil, rlsRcTag)
			if res.exit != 1 {
				t.Errorf("exit = %d, want 1\n%s", res.exit, res.out)
			}
			rlsMustContain(t, res.out, "Matrix gate:", sc.wantReason)
			if res.localTag != "" || res.remoteTags != "" {
				t.Errorf("a tag exists after the refusal: local=%q remote=%q", res.localTag, res.remoteTags)
			}
		})
	}
	t.Run("control_green_matrix_reaches_the_tag", func(t *testing.T) {
		r := rlsBuildReleaseRepo(t, rlsRcChangelog)
		s := rlsNewStub(t)
		sha := r.git("rev-parse", "HEAD")
		rlsGreenMatrix(t, s, sha)
		res := rlsRealTagRun(t, r, s, nil, rlsRcTag)
		if res.localTag != sha {
			t.Errorf("green matrix: local tag = %q, want %q (the harness never reached tag creation)\n%s", res.localTag, sha, res.out)
		}
		if res.remoteTags != "" {
			t.Errorf("remote tag exists although push is blocked: %q", res.remoteTags)
		}
	})
}

// TestReleaseScriptNoSwitchTurnsTheMatrixOff: neither the environment nor
// another gate's override reaches the matrix gate. With no matrix run on the
// commit every row is refused at Validation 9b.
func TestReleaseScriptNoSwitchTurnsTheMatrixOff(t *testing.T) {
	rlsRequireTools(t, "jq")
	rows := []struct {
		name     string
		env      []string
		args     []string
		wantAlso string
	}{
		{name: "env_require_matrix_run_false", env: []string{"REQUIRE_MATRIX_RUN=false"}, args: []string{rlsRcTag, "--dry-run"}},
		{name: "env_matrix_gate_off_spellings", env: []string{"REQUIRE_MATRIX_RUN=0", "MATRIX_GATE=off", "SKIP_MATRIX_RUN=true"}, args: []string{rlsRcTag, "--dry-run"}},
		{name: "skip_ci_check_does_not_skip_the_matrix", args: []string{rlsRcTag, "--dry-run", "--skip-ci-check"}, wantAlso: "CI gate OVERRIDDEN (--skip-ci-check)"},
		{name: "hotfix_does_not_skip_the_matrix", args: []string{rlsRcTag, "--dry-run", "--hotfix"}},
		{name: "require_matrix_run_is_a_noop_flag", args: []string{rlsRcTag, "--dry-run", "--require-matrix-run"}},
	}
	for _, row := range rows {
		t.Run(row.name, func(t *testing.T) {
			r := rlsBuildReleaseRepo(t, rlsRcChangelog)
			s := rlsNewStub(t) // no matrix run on the commit
			script := filepath.Join(findProjectRootForMirrorTest(t), filepath.FromSlash(rlsReleaseScriptRel))
			env := append(append([]string{}, s.env...), row.env...)
			res := rlsRun(t, r.work, env, "bash", append([]string{script}, row.args...)...)
			out := rlsNorm(res.out)
			if res.exit != 1 {
				t.Errorf("exit = %d, want 1 (refused at the matrix gate)\n%s", res.exit, out)
			}
			rlsMustContain(t, out, "Matrix gate:", "no run of "+rlsMatrixFile+" found for")
			if row.wantAlso != "" {
				rlsMustContain(t, out, row.wantAlso)
			}
		})
	}
}
