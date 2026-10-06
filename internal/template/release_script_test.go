// release_script_test.go: guards for scripts/release.sh
// (SPEC-GITHUB-FLOW-DEFAULT-001 M3: AC-GFD-007 head admission, AC-GFD-010
// matrix gate, design D-6 rc CHANGELOG skip, design D-24 the switch).
//
// Every test runs the real script with `--dry-run` inside a scratch repository
// whose origin is a scratch bare repository, with a stub `gh` first on PATH
// (release_fixture_test.go). `--dry-run` returns before the tag is created, so
// nothing here can tag or push.
//
// TestReleaseScriptCharacterization pins the behaviour of every non-detached,
// non-rc, no-option invocation as it was before this milestone; it is green
// before and after, which is the proof that those invocations did not change.
package template_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const (
	rlsRcTag       = "v9.9.9-rc.1"
	rlsMatrixName  = "Release PR Multi-OS Verification"
	rlsMatrixFile  = "release-pr-multi-os.yml"
	rlsOtherSHA    = "1111111111111111111111111111111111111111"
	rlsRcChangelog = "# Changelog\n\n## [Unreleased]\n\n## [v9.9.9-rc.1] - 2026-01-01\n\n- fixture\n"
)

// rlsBuildReleaseRepo builds origin + clone with two commits on main, both
// pushed: the older one gives the "behind origin/main" fixture its target.
func rlsBuildReleaseRepo(t *testing.T, changelog string) *rlsRepo {
	t.Helper()
	r := rlsNewRepo(t)
	r.commit(map[string]string{"CHANGELOG.md": changelog, "README.md": "first\n"}, "first")
	r.git("push", "-q", "origin", "main")
	r.commit(map[string]string{"README.md": "second\n"}, "second")
	r.git("push", "-q", "origin", "main")
	return r
}

// rlsRelease runs scripts/release.sh from the real repository inside the
// scratch clone. The script locates its repository from the working directory.
func rlsRelease(t *testing.T, r *rlsRepo, s *rlsStub, args ...string) rlsResult {
	t.Helper()
	script := filepath.Join(findProjectRootForMirrorTest(t), filepath.FromSlash(rlsReleaseScriptRel))
	return rlsRun(t, r.work, s.env, "bash", append([]string{script}, args...)...)
}

// TestReleaseScriptCharacterization pins what the script does for invocations
// this milestone must not change: a formal tag from a synced main, the CHANGELOG
// heading forms, every refusal on a branch that is not main, a dirty tree, an
// unknown flag. The expectations were observed on the pre-change script.
func TestReleaseScriptCharacterization(t *testing.T) {
	rlsRequireTools(t)

	formal := "# Changelog\n\n## [Unreleased]\n\n## [9.9.8] - 2026-01-01\n\n- fixture\n"
	rows := []struct {
		name      string
		changelog string
		prepare   func(*rlsRepo)
		args      []string
		wantExit  int
		wantOut   []string
	}{
		{name: "formal_tag_synced_main_bare_heading", changelog: formal, args: []string{"v9.9.8", "--dry-run"}, wantExit: 0,
			wantOut: []string{"On expected branch: main", "Local main synced with origin", "Tag v9.9.8 does not exist yet",
				"CHANGELOG.md contains ## [9.9.8] section", "DRY RUN - skipping tag creation"}},
		{name: "formal_tag_v_prefixed_heading", changelog: "# Changelog\n\n## [v9.9.8] - 2026-01-01\n\n- fixture\n", args: []string{"v9.9.8", "--dry-run"}, wantExit: 0,
			wantOut: []string{"CHANGELOG.md contains ## [v9.9.8] section"}},
		{name: "rc_tag_with_a_changelog_section_still_uses_it", changelog: rlsRcChangelog, args: []string{rlsRcTag, "--dry-run"}, wantExit: 0,
			wantOut: []string{"CHANGELOG.md contains ## [v9.9.9-rc.1] section", "## [v9.9.9-rc.1] - 2026-01-01"}},
		{name: "formal_tag_without_a_changelog_section_is_refused", changelog: "# Changelog\n\n## [Unreleased]\n", args: []string{"v9.9.8", "--dry-run"}, wantExit: 1,
			wantOut: []string{"CHANGELOG.md missing section '## [9.9.8]' (or '## [v9.9.8]'). Add release notes first."}},
		{name: "feature_branch_is_refused", changelog: formal, args: []string{"v9.9.8", "--dry-run"}, wantExit: 1,
			prepare: func(r *rlsRepo) { r.git("checkout", "-q", "-b", "feature") },
			wantOut: []string{"Must be on 'main' branch (current: feature). Use --hotfix for hotfix branches."}},
		{name: "hotfix_on_a_pushed_feature_branch_is_allowed", changelog: formal, args: []string{"v9.9.8", "--dry-run", "--hotfix"}, wantExit: 0,
			prepare: func(r *rlsRepo) {
				r.git("checkout", "-q", "-b", "hotfix-x")
				r.git("push", "-q", "origin", "hotfix-x")
			},
			wantOut: []string{"On branch 'hotfix-x' (hotfix mode - allowed)", "On expected branch: hotfix-x", "Local hotfix-x synced with origin"}},
		{name: "local_main_ahead_of_origin_is_refused", changelog: formal, args: []string{"v9.9.8", "--dry-run"}, wantExit: 1,
			prepare: func(r *rlsRepo) { r.commit(map[string]string{"README.md": "third\n"}, "local only") },
			wantOut: []string{"Local 'main' diverged from origin (ahead: 1, behind: 0). Pull/push first."}},
		{name: "dirty_tree_is_refused", changelog: formal, args: []string{"v9.9.8", "--dry-run"}, wantExit: 1,
			prepare: func(r *rlsRepo) {
				if err := os.WriteFile(filepath.Join(r.work, "stray.txt"), []byte("x\n"), 0o644); err != nil {
					t.Fatalf("write stray file: %v", err)
				}
			},
			wantOut: []string{"Working tree is dirty. Commit or stash changes first."}},
		{name: "unknown_flag_is_refused", changelog: formal, args: []string{"v9.9.8", "--no-such-flag"}, wantExit: 1,
			wantOut: []string{"Unknown flag: --no-such-flag (try --help)"}},
	}

	for _, row := range rows {
		t.Run(row.name, func(t *testing.T) {
			r := rlsBuildReleaseRepo(t, row.changelog)
			s := rlsNewStub(t)
			if row.prepare != nil {
				row.prepare(r)
			}
			res := rlsRelease(t, r, s, row.args...)
			out := rlsNorm(res.out)
			t.Logf("exit=%d\n%s", res.exit, out)
			if res.exit != row.wantExit {
				t.Errorf("exit code = %d, want %d", res.exit, row.wantExit)
			}
			for _, w := range row.wantOut {
				// The pre-change script prints an em dash in two of these lines; the
				// pin is on the words, so compare with the dash normalized.
				if !strings.Contains(strings.ReplaceAll(out, "—", "-"), w) {
					t.Errorf("output lacks %q", w)
				}
			}
			for _, c := range s.calls(t) {
				if strings.HasPrefix(c, "run ") {
					t.Errorf("a no-option invocation must not touch gh run: %q", c)
				}
			}
		})
	}
}

// TestReleaseScriptHeadAdmission is AC-GFD-007: the release harness runs from a
// detached worktree, so the script must admit a detached HEAD that IS
// origin/main's commit and nothing else that is not a synced main branch.
func TestReleaseScriptHeadAdmission(t *testing.T) {
	rlsRequireTools(t)

	rows := []struct {
		name     string
		prepare  func(*rlsRepo)
		wantExit int
		wantOut  string
	}{
		{name: "a_detached_at_origin_main_tip", wantExit: 0, wantOut: "Detached HEAD equals origin/main",
			prepare: func(r *rlsRepo) { r.git("checkout", "-q", "--detach", "origin/main") }},
		{name: "b_detached_at_an_older_commit", wantExit: 1, wantOut: "Detached HEAD is not origin/main (ahead: 0, behind: 1)",
			prepare: func(r *rlsRepo) { r.git("checkout", "-q", "--detach", "HEAD~1") }},
		{name: "c_main_branch_equal_to_origin_main", wantExit: 0, wantOut: "Local main synced with origin",
			prepare: func(r *rlsRepo) {}},
		{name: "d_non_main_branch_at_the_same_commit", wantExit: 1, wantOut: "Must be on 'main' branch (current: feature)",
			prepare: func(r *rlsRepo) { r.git("checkout", "-q", "-b", "feature") }},
		{name: "e_main_ahead_of_origin_main", wantExit: 1, wantOut: "Local 'main' diverged from origin (ahead: 1, behind: 0)",
			prepare: func(r *rlsRepo) { r.commit(map[string]string{"README.md": "third\n"}, "local only") }},
	}

	for _, row := range rows {
		t.Run(row.name, func(t *testing.T) {
			r := rlsBuildReleaseRepo(t, rlsRcChangelog)
			s := rlsNewStub(t)
			row.prepare(r)
			res := rlsRelease(t, r, s, rlsRcTag, "--dry-run")
			t.Logf("exit=%d\n%s", res.exit, rlsNorm(res.out))
			if res.exit != row.wantExit {
				t.Errorf("exit code = %d, want %d", res.exit, row.wantExit)
			}
			rlsMustContain(t, rlsNorm(res.out), row.wantOut)
		})
	}
}

// TestReleaseScriptRcChangelogSkip is design D-6's release.sh limb: validation
// 8 (CHANGELOG section) is skipped for an `-rc.N` tag and only for it.
func TestReleaseScriptRcChangelogSkip(t *testing.T) {
	rlsRequireTools(t)
	noSection := "# Changelog\n\n## [Unreleased]\n"

	rows := []struct {
		name     string
		tag      string
		wantExit int
		wantOut  string
	}{
		{name: "rc_tag_without_a_section_passes", tag: rlsRcTag, wantExit: 0, wantOut: "no CHANGELOG.md section required"},
		{name: "legacy_undotted_rc_keeps_the_requirement", tag: "v9.9.9-rc12", wantExit: 1, wantOut: "CHANGELOG.md missing section '## [9.9.9-rc12]'"},
		{name: "other_prerelease_keeps_the_requirement", tag: "v9.9.9-beta.1", wantExit: 1, wantOut: "CHANGELOG.md missing section '## [9.9.9-beta.1]'"},
		{name: "formal_tag_keeps_the_requirement", tag: "v9.9.9", wantExit: 1, wantOut: "CHANGELOG.md missing section '## [9.9.9]'"},
	}
	for _, row := range rows {
		t.Run(row.name, func(t *testing.T) {
			r := rlsBuildReleaseRepo(t, noSection)
			s := rlsNewStub(t)
			res := rlsRelease(t, r, s, row.tag, "--dry-run")
			t.Logf("exit=%d\n%s", res.exit, rlsNorm(res.out))
			if res.exit != row.wantExit {
				t.Errorf("exit code = %d, want %d", res.exit, row.wantExit)
			}
			rlsMustContain(t, rlsNorm(res.out), row.wantOut)
		})
	}
}

// matrixScenario is one `gh` double of AC-GFD-010: the canned runs of the
// multi-OS workflow and the jobs of the run the script is expected to read.
type matrixScenario struct {
	name string
	// build returns the runs JSON for the target SHA and registers jobs files.
	build func(t *testing.T, s *rlsStub, sha string) string
	// wantReason is a substring the refusal must carry; empty means pass.
	wantReason string
}

func rlsRunJSON(id int, sha, event, status, conclusion string) string {
	concl := "null"
	if conclusion != "" {
		concl = `"` + conclusion + `"`
	}
	return fmt.Sprintf(`{"databaseId":%d,"headSha":%q,"event":%q,"status":%q,"conclusion":%s,"workflowName":%q}`,
		id, sha, event, status, concl, rlsMatrixName)
}

func rlsLegsJSON(win string) string {
	job := func(name, conclusion string) string {
		return fmt.Sprintf(`{"name":%q,"status":"completed","conclusion":%q}`, name, conclusion)
	}
	return `{"jobs":[` + strings.Join([]string{
		job("Detect Release PR", "success"),
		job("Release Verify (ubuntu-latest)", "success"),
		job("Release Verify (macos-latest)", "success"),
		job("Release Verify (windows-latest)", win),
	}, ",") + `]}`
}

func matrixScenarios() []matrixScenario {
	return []matrixScenario{
		{name: "a_three_legs_green_on_the_target_sha", build: func(t *testing.T, s *rlsStub, sha string) string {
			s.setJobs(t, 101, rlsLegsJSON("success"))
			return "[" + rlsRunJSON(101, sha, "workflow_dispatch", "completed", "success") + "]"
		}},
		{name: "b_no_run", wantReason: "found for", build: func(t *testing.T, s *rlsStub, sha string) string { return "[]" }},
		{name: "c_green_run_on_another_sha", wantReason: "found for", build: func(t *testing.T, s *rlsStub, sha string) string {
			s.setJobs(t, 102, rlsLegsJSON("success"))
			return "[" + rlsRunJSON(102, rlsOtherSHA, "workflow_dispatch", "completed", "success") + "]"
		}},
		{name: "d_failed_run_on_the_target_sha", wantReason: "concluded failure", build: func(t *testing.T, s *rlsStub, sha string) string {
			s.setJobs(t, 103, rlsLegsJSON("failure"))
			return "[" + rlsRunJSON(103, sha, "workflow_dispatch", "completed", "failure") + "]"
		}},
		{name: "e_run_still_in_progress", wantReason: "not completed", build: func(t *testing.T, s *rlsStub, sha string) string {
			s.setJobs(t, 104, rlsLegsJSON("success"))
			return "[" + rlsRunJSON(104, sha, "workflow_dispatch", "in_progress", "") + "]"
		}},
		{name: "f_run_success_but_one_leg_skipped", wantReason: "leg 'Release Verify (windows-latest)' is skipped", build: func(t *testing.T, s *rlsStub, sha string) string {
			s.setJobs(t, 105, rlsLegsJSON("skipped"))
			return "[" + rlsRunJSON(105, sha, "workflow_dispatch", "completed", "success") + "]"
		}},
		{name: "g_green_run_from_a_pull_request_event", wantReason: "none came from workflow_dispatch", build: func(t *testing.T, s *rlsStub, sha string) string {
			s.setJobs(t, 106, rlsLegsJSON("success"))
			return "[" + rlsRunJSON(106, sha, "pull_request", "completed", "success") + "]"
		}},
	}
}

// TestReleaseScriptMatrixGate is AC-GFD-010: two switch states times seven gh
// doubles. With the switch off the script must not look at the matrix at all
// (no `gh run` call is logged); with `--require-matrix-run` only the run that
// is complete, green on all three legs, on the target SHA and dispatched by
// hand passes, and every other double is refused before the tag with the
// condition named.
func TestReleaseScriptMatrixGate(t *testing.T) {
	rlsRequireTools(t, "jq")

	for _, sw := range []struct {
		name string
		args []string
	}{
		{name: "switch_off", args: []string{rlsRcTag, "--dry-run"}},
		{name: "switch_on", args: []string{rlsRcTag, "--dry-run", "--require-matrix-run"}},
	} {
		for _, sc := range matrixScenarios() {
			t.Run(sw.name+"/"+sc.name, func(t *testing.T) {
				r := rlsBuildReleaseRepo(t, rlsRcChangelog)
				s := rlsNewStub(t)
				sha := r.git("rev-parse", "HEAD")
				s.setRuns(t, sc.build(t, s, sha))

				res := rlsRelease(t, r, s, sw.args...)
				out := rlsNorm(res.out)
				t.Logf("exit=%d\n%s\ngh calls: %q", res.exit, out, s.calls(t))

				if sw.name == "switch_off" {
					if res.exit != 0 {
						t.Errorf("switch off: exit code = %d, want 0 (the matrix is not consulted)", res.exit)
					}
					for _, c := range s.calls(t) {
						if strings.HasPrefix(c, "run ") {
							t.Errorf("switch off: the script called %q; it must not read the matrix at all", c)
						}
					}
					return
				}

				if sc.wantReason == "" {
					if res.exit != 0 {
						t.Errorf("switch on, green run: exit code = %d, want 0", res.exit)
					}
					rlsMustContain(t, out, "Matrix gate: run 101 of "+rlsMatrixFile+" is green on all three OS legs")
					return
				}
				if res.exit != 1 {
					t.Errorf("switch on: exit code = %d, want 1 (refused before the tag)", res.exit)
				}
				rlsMustContain(t, out, "Matrix gate:", sc.wantReason)
			})
		}
	}
}
