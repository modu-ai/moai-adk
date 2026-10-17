// release_ci_gate_test.go: guards for the CI gate in scripts/release.sh (card
// t1833). A pushed v* tag cannot be deleted or moved, so the script refuses to
// place one on a commit whose CI has not passed on origin. Every test runs the
// real script with `--dry-run` in a scratch repository with a stub `gh`
// (release_fixture_test.go); nothing here tags or pushes.
package template_test

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

const (
	rlsCheckRunsGreen = `{"total_count":3,"check_runs":[` +
		`{"name":"build","status":"completed","conclusion":"success"},` +
		`{"name":"lint","status":"completed","conclusion":"success"},` +
		`{"name":"docs","status":"completed","conclusion":"skipped"}]}`
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
			wantOut: []string{"CI status on HEAD: success (3 check runs", "CI gate:     verified"}},
		{name: "pending_is_refused", runs: rlsCheckRunsPending, wantExit: 1,
			wantOut: []string{"CI gate: 1 of 2 check run(s)", "not completed"}},
		{name: "failed_is_refused", runs: rlsCheckRunsFailed, wantExit: 1,
			wantOut: []string{"CI gate: 1 of 2 check run(s)", "did not succeed"}},
		{name: "failed_beats_pending", runs: rlsCheckRunsFailedAndPending, wantExit: 1,
			wantOut: []string{"did not succeed"}},
		{name: "absent_is_refused", runs: rlsCheckRunsAbsent, wantExit: 1,
			wantOut: []string{"CI gate: no check runs"}},
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
