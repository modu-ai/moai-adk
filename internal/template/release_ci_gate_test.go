// release_ci_gate_test.go: guards for the CI gate in scripts/release.sh (card
// t1833). A pushed v* tag cannot be deleted or moved, so the script refuses to
// place one on a commit whose CI has not passed on origin. Every test runs the
// real script with `--dry-run` in a scratch repository with a stub `gh`
// (release_fixture_test.go); nothing here tags or pushes.
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

const (
	// The green page names the check runs the repository CI workflow (.github/
	// workflows/ci.yml, name: CI) publishes on a push to main; the gate demands
	// them by name (fix round 1, F6).
	rlsCheckRunsGreen = `{"total_count":4,"check_runs":[` +
		`{"name":"Lint","status":"completed","conclusion":"success","app":{"slug":"github-actions"}},` +
		`{"name":"Test (ubuntu-latest)","status":"completed","conclusion":"success","app":{"slug":"github-actions"}},` +
		`{"name":"Build (linux/amd64)","status":"completed","conclusion":"success","app":{"slug":"github-actions"}},` +
		`{"name":"docs","status":"completed","conclusion":"skipped","app":{"slug":"github-actions"}}]}`
	// Green, completed, and from an unrelated app: no CI workflow check run.
	rlsCheckRunsUnrelatedGreen = `{"total_count":1,"check_runs":[` +
		`{"name":"CodeRabbit","status":"completed","conclusion":"success","app":{"slug":"coderabbitai"}}]}`
	// Lint and Test reported, the Build leg never did.
	rlsCheckRunsBuildMissing = `{"total_count":3,"check_runs":[` +
		`{"name":"Lint","status":"completed","conclusion":"success","app":{"slug":"github-actions"}},` +
		`{"name":"Test (ubuntu-latest)","status":"completed","conclusion":"success","app":{"slug":"github-actions"}},` +
		`{"name":"CodeRabbit","status":"completed","conclusion":"success","app":{"slug":"coderabbitai"}}]}`
	// Fix round 2 (N2): all three required names, but one is published by another
	// app. A name alone does not prove the CI workflow ran.
	rlsCheckRunsRightNameWrongApp = `{"total_count":3,"check_runs":[` +
		`{"name":"Lint","status":"completed","conclusion":"success","app":{"slug":"github-actions"}},` +
		`{"name":"Test (ubuntu-latest)","status":"completed","conclusion":"success","app":{"slug":"some-other-app"}},` +
		`{"name":"Build (linux/amd64)","status":"completed","conclusion":"success","app":{"slug":"github-actions"}}]}`
	// The app object absent (an app-less check run): treated as not from Actions.
	rlsCheckRunsNoApp = `{"total_count":3,"check_runs":[` +
		`{"name":"Lint","status":"completed","conclusion":"success"},` +
		`{"name":"Test (ubuntu-latest)","status":"completed","conclusion":"success"},` +
		`{"name":"Build (linux/amd64)","status":"completed","conclusion":"success"}]}`
	rlsCheckRunsPending = `{"total_count":2,"check_runs":[` +
		`{"name":"build","status":"completed","conclusion":"success"},` +
		`{"name":"test","status":"in_progress","conclusion":null}]}`
	rlsCheckRunsFailed = `{"total_count":2,"check_runs":[` +
		`{"name":"build","status":"completed","conclusion":"success"},` +
		`{"name":"test","status":"completed","conclusion":"failure"}]}`
	rlsCheckRunsFailedAndPending = `{"total_count":2,"check_runs":[` +
		`{"name":"test","status":"completed","conclusion":"cancelled"},` +
		`{"name":"lint","status":"queued","conclusion":null}]}`
	rlsCheckRunsAbsent = `{"total_count":0,"check_runs":[]}`
)

func (s *rlsStub) setCheckRuns(t *testing.T, json string) {
	t.Helper()
	if err := os.WriteFile(s.checkRuns, []byte(json), 0o644); err != nil {
		t.Fatalf("write stub check runs: %v", err)
	}
}

// TestReleaseScriptCIGate: the tag is placed only on a commit whose check runs
// on origin are all completed and green; an absent, pending, failed or
// unreadable verdict refuses it before the tag, naming the condition.
func TestReleaseScriptCIGate(t *testing.T) {
	rlsRequireTools(t, "jq")

	rows := []struct {
		name     string
		runs     string
		apiFail  bool
		args     []string
		wantExit int
		wantOut  []string
	}{
		{name: "green", runs: rlsCheckRunsGreen, wantExit: 0,
			wantOut: []string{"CI status on HEAD: success (4 check runs", "CI gate:     verified"}},
		{name: "pending_is_refused", runs: rlsCheckRunsPending, wantExit: 1,
			wantOut: []string{"CI gate: 1 of 2 check run(s)", "not completed"}},
		{name: "failed_is_refused", runs: rlsCheckRunsFailed, wantExit: 1,
			wantOut: []string{"CI gate: 1 of 2 check run(s)", "did not succeed"}},
		{name: "failed_beats_pending", runs: rlsCheckRunsFailedAndPending, wantExit: 1,
			wantOut: []string{"did not succeed"}},
		{name: "absent_is_refused", runs: rlsCheckRunsAbsent, wantExit: 1,
			wantOut: []string{"CI gate: no check runs"}},
		{name: "unrelated_green_check_run_alone_is_refused", runs: rlsCheckRunsUnrelatedGreen, wantExit: 1,
			wantOut: []string{"CI gate: expected check run(s) not reported", "Lint", "Test (ubuntu-latest)", "Build (linux/amd64)"}},
		{name: "right_name_from_another_app_is_refused", runs: rlsCheckRunsRightNameWrongApp, wantExit: 1,
			wantOut: []string{"CI gate: expected check run(s) not reported", "Test (ubuntu-latest)", "github-actions"}},
		{name: "right_names_without_an_app_are_refused", runs: rlsCheckRunsNoApp, wantExit: 1,
			wantOut: []string{"CI gate: expected check run(s) not reported", "Lint", "github-actions"}},
		{name: "expected_workflow_missing_one_check_is_refused", runs: rlsCheckRunsBuildMissing, wantExit: 1,
			wantOut: []string{"CI gate: expected check run(s) not reported", "Build (linux/amd64)"}},
		{name: "expected_workflow_green_passes_beside_an_unrelated_run", runs: rlsCheckRunsGreen, wantExit: 0,
			wantOut: []string{"CI gate:     verified"}},
		{name: "override_on_an_unrelated_only_page_is_recorded", runs: rlsCheckRunsUnrelatedGreen, args: []string{"--skip-ci-check"}, wantExit: 0,
			wantOut: []string{"CI gate OVERRIDDEN (--skip-ci-check)"}},
		{name: "unreadable_is_refused", runs: rlsCheckRunsGreen, apiFail: true, wantExit: 1,
			wantOut: []string{"CI gate: could not read the check runs"}},
		{name: "override_on_a_failed_run_is_recorded", runs: rlsCheckRunsFailed, args: []string{"--skip-ci-check"}, wantExit: 0,
			wantOut: []string{"CI gate OVERRIDDEN (--skip-ci-check)", "CI gate:     OVERRIDDEN (--skip-ci-check)"}},
	}
	for _, row := range rows {
		t.Run(row.name, func(t *testing.T) {
			r := rlsBuildReleaseRepo(t, rlsRcChangelog)
			s := rlsNewStub(t)
			s.setCheckRuns(t, row.runs)
			if row.apiFail {
				s.env = append(s.env, "GH_STUB_API_FAIL=1")
			}
			sha := r.git("rev-parse", "HEAD")
			res := rlsRelease(t, r, s, append([]string{rlsRcTag, "--dry-run"}, row.args...)...)
			out := rlsNorm(res.out)
			t.Logf("exit=%d\n%s\ngh calls: %q", res.exit, out, s.calls(t))
			if res.exit != row.wantExit {
				t.Errorf("exit code = %d, want %d", res.exit, row.wantExit)
			}
			rlsMustContain(t, out, row.wantOut...)
			if len(row.args) == 0 {
				want := fmt.Sprintf("commits/%s/check-runs", sha)
				found := false
				for _, c := range s.calls(t) {
					if strings.Contains(c, want) {
						found = true
					}
				}
				if !found {
					t.Errorf("the verdict must be read for the commit to be tagged (%s); gh calls: %q", want, s.calls(t))
				}
			}
			if row.wantExit == 1 && strings.Contains(out, "skipping tag creation") {
				t.Errorf("a refused run must stop before the dry-run preview")
			}
		})
	}
}

// TestReleaseScriptCIGateWithoutGh: with no gh on PATH the gate fails closed,
// and only the explicit override lets the run continue.
func TestReleaseScriptCIGateWithoutGh(t *testing.T) {
	rlsRequireTools(t)

	bin := t.TempDir()
	for _, tool := range []string{"bash", "git", "awk", "grep", "wc", "tr", "find", "mktemp", "rm", "head", "date", "cat", "sed", "dirname", "env", "sh", "basename", "sleep"} {
		p := rlsLookPath(tool)
		if p == "" {
			t.Skipf("%s not available; cannot build a gh-free PATH", tool)
		}
		if err := os.Symlink(p, filepath.Join(bin, tool)); err != nil {
			t.Fatalf("link %s: %v", tool, err)
		}
	}
	for _, args := range [][]string{{}, {"--skip-ci-check"}} {
		name := "no_override"
		wantExit, wantOut := 1, "CI gate: the gh CLI is not available"
		if len(args) > 0 {
			name, wantExit, wantOut = "override", 0, "CI gate OVERRIDDEN (--skip-ci-check)"
		}
		t.Run(name, func(t *testing.T) {
			r := rlsBuildReleaseRepo(t, rlsRcChangelog)
			env := rlsEnv("PATH=" + bin)
			script := filepath.Join(findProjectRootForMirrorTest(t), filepath.FromSlash(rlsReleaseScriptRel))
			res := rlsRun(t, r.work, env, bin+"/bash", append([]string{script, rlsRcTag, "--dry-run"}, args...)...)
			out := rlsNorm(res.out)
			t.Logf("exit=%d\n%s", res.exit, out)
			if res.exit != wantExit {
				t.Errorf("exit code = %d, want %d", res.exit, wantExit)
			}
			rlsMustContain(t, out, wantOut)
		})
	}
}

// rlsLookPath is exec.LookPath that reports "" instead of an error.
func rlsLookPath(tool string) string {
	p, err := exec.LookPath(tool)
	if err != nil {
		return ""
	}
	return p
}

// TestReleaseScriptTagsTheVerifiedCommit (fix round 1, F5): the tag names the
// commit the CI gate verified, so a HEAD that moves after the re-check cannot
// carry the tag. A git shim moves HEAD at the moment `git tag -a` is invoked
// and blocks the push, so the test ends at a local tag in a scratch clone.
func TestReleaseScriptTagsTheVerifiedCommit(t *testing.T) {
	rlsRequireTools(t, "jq")

	realGit, err := exec.LookPath("git")
	if err != nil {
		t.Skip("git not available")
	}
	r := rlsBuildReleaseRepo(t, rlsRcChangelog)
	s := rlsNewStub(t)
	verified := r.git("rev-parse", "HEAD")

	shim := t.TempDir()
	rlsWriteScript(t, shim, "git", fmt.Sprintf(`#!/usr/bin/env bash
case "${1:-} ${2:-}" in
  "tag -a") %q commit -q --allow-empty -m "HEAD moved after the re-check" ;;
  "push "*) echo "git shim: push blocked" >&2; exit 1 ;;
esac
exec %q "$@"
`, realGit, realGit))
	// A non-dry run needs the harness marker; the scratch clone and the push
	// block keep this away from the real repository and any remote.
	// The shim goes in front of the PATH the gh stub already set up, so the
	// real gh can never be reached.
	stubPath := ""
	for _, kv := range s.env {
		if v, ok := strings.CutPrefix(kv, "PATH="); ok {
			stubPath = v
		}
	}
	env := append(append([]string{}, s.env...), "MOAI_RELEASE_VIA_HARNESS=1", "PATH="+shim+string(os.PathListSeparator)+stubPath)

	script := filepath.Join(findProjectRootForMirrorTest(t), filepath.FromSlash(rlsReleaseScriptRel))
	ctx, cancel := context.WithTimeout(context.Background(), rlsDeadline)
	defer cancel()
	cmd := exec.CommandContext(ctx, "bash", script, rlsRcTag)
	cmd.Dir, cmd.Env, cmd.Stdin = r.work, env, strings.NewReader("y\n")
	out, runErr := cmd.CombinedOutput()
	if ctx.Err() != nil {
		t.Fatalf("release.sh exceeded %s\n%s", rlsDeadline, out)
	}
	var exitErr *exec.ExitError
	if runErr != nil && !errors.As(runErr, &exitErr) {
		t.Fatalf("start release.sh: %v", runErr)
	}
	t.Logf("release.sh err=%v\n%s", runErr, rlsNorm(string(out)))

	moved := r.git("rev-parse", "HEAD")
	if moved == verified {
		t.Fatalf("the shim did not move HEAD; the test proves nothing")
	}
	tagged := strings.TrimSpace(rlsRun(t, r.work, rlsEnv(), "git", "rev-parse", "--verify", "--quiet", rlsRcTag+"^{commit}").out)
	if tagged == "" {
		return // refusing to tag is also an acceptable outcome
	}
	if tagged != verified {
		t.Errorf("tag %s names %s, but the CI gate verified %s (HEAD had moved to %s)", rlsRcTag, tagged, verified, moved)
	}
}
