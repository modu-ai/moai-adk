// release_provenance_test.go: guards for the release provenance gate
// (SPEC-GITHUB-FLOW-DEFAULT-001 M3, AC-GFD-008, design D-6, REQ-GFD-008).
//
// The gate is the seven-check body of the verify-provenance job in
// .github/workflows/release.yml, moved to scripts/verify-release-provenance.sh
// so it can be run against fixtures. The tests judge the shipped gate, not only
// the script: the workflow must call the script with the tag, must no longer
// carry the checks inline, and the call line replayed on every fixture must
// give the same verdict as calling the script directly.
//
// Verdict rows are exact — the verdict lines of a suffix-less tag are the ones
// the inline step printed before the move, observed on scratch fixtures and
// pinned here (TestReleaseProvenanceStepReplaySuffixless is that pin).
package template_test

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

const rlsGateTrailerVia = "Released-via: harness:release"

// rlsProvKind says how a fixture tag is built; each kind violates (or satisfies)
// exactly one of the seven checks.
type rlsProvKind int

const (
	provOK          rlsProvKind = iota // annotated, trailer correct, commit on main
	provLightwseven                    // check 1: not an annotated tag
	provNoTrailer                      // check 2: annotation has no provenance trailer
	provBadVersion                     // check 3: trailer version differs from the tag
	provBadCommit                      // check 4: trailer commit differs from the tagged commit
	provOffMain                        // check 7: tagged commit is not an ancestor of origin/main
)

type rlsProvRow struct {
	name      string
	tag       string
	kind      rlsProvKind
	changelog bool   // CHANGELOG.md at the tagged commit has this version's section
	ssot      string // version in .moai/config/sections/system.yaml at the tagged commit
	wantExit  int
	// wantOut is the exact normalized verdict output (SHAs become <sha>).
	wantOut []string
}

const rlsGateTail = "Releases must be produced by the /harness:release maintainer harness."

func rlsGateFail(msg string) []string {
	return []string{"::error::RELEASE_PROVENANCE_GATE: " + msg, rlsGateTail}
}

func rlsRcSkipNotice(tag string) string {
	return fmt.Sprintf("RELEASE_PROVENANCE_GATE: pre-release tag %s: skipping check 5 (CHANGELOG), check 6 (version SSOT) per the rc rule.", tag)
}

func rlsRcPass(tag string) string {
	return fmt.Sprintf("RELEASE_PROVENANCE_GATE: 5 applicable checks passed for %s (<sha>); checks 5 and 6 skipped (pre-release).", tag)
}

// rlsSuffixlessRows are the suffix-less (formal release) tags: their seven
// checks must be exactly what the inline step did before the move.
func rlsSuffixlessRows() []rlsProvRow {
	return []rlsProvRow{
		{name: "6_all_seven_checks_hold", tag: "v9.9.9", kind: provOK, changelog: true, ssot: "v9.9.9", wantExit: 0,
			wantOut: []string{
				"RELEASE_PROVENANCE_GATE: all 7 checks passed for v9.9.9 (<sha>)."}},
		{name: "4_no_changelog_section", tag: "v9.9.8", kind: provOK, changelog: false, ssot: "v9.9.8", wantExit: 1,
			wantOut: rlsGateFail("check 5 (CHANGELOG): CHANGELOG.md at <sha> has no '## [9.9.8]' or '## [v9.9.8]' section.")},
		{name: "5_ssot_version_differs", tag: "v9.9.7", kind: provOK, changelog: true, ssot: "v9.9.6", wantExit: 1,
			wantOut: rlsGateFail("check 6 (version SSOT): .moai/config/sections/system.yaml at <sha> has version='v9.9.6', expected 'v9.9.7'.")},
		{name: "check1_lightwseven_tag", tag: "v9.9.6", kind: provLightwseven, changelog: true, ssot: "v9.9.6", wantExit: 1,
			wantOut: rlsGateFail("check 1 (annotated tag): 'v9.9.6' is a commit object, not an annotated tag.")},
		{name: "check2_no_trailer", tag: "v9.9.5", kind: provNoTrailer, changelog: true, ssot: "v9.9.5", wantExit: 1,
			wantOut: rlsGateFail("check 2 (provenance trailer): tag annotation has no 'Released-via: harness:release' line.")},
		{name: "check3_trailer_version_differs", tag: "v9.9.4", kind: provBadVersion, changelog: true, ssot: "v9.9.4", wantExit: 1,
			wantOut: rlsGateFail("check 3 (version binding): trailer Release-version='v9.9.99' != pushed tag 'v9.9.4'.")},
		{name: "check4_trailer_commit_differs", tag: "v9.9.3", kind: provBadCommit, changelog: true, ssot: "v9.9.3", wantExit: 1,
			wantOut: rlsGateFail("check 4 (commit binding): trailer Release-commit='<sha>' != tagged commit '<sha>'.")},
		{name: "check7_not_on_main", tag: "v9.9.2", kind: provOffMain, changelog: true, ssot: "v9.9.2", wantExit: 1,
			wantOut: rlsGateFail("check 7 (main ancestry): tagged commit <sha> is not an ancestor of origin/main.")},
	}
}

// rlsRcRows are the pre-release rows: AC-GFD-008 fixtures (1)-(3) plus the
// plan-audit carry-over D13 (checks 1, 2 and 3 still hold for an rc tag), and
// two near-miss tags that the rc rule must NOT cover.
func rlsRcRows() []rlsProvRow {
	return []rlsProvRow{
		{name: "1_rc_trailer_ok_main_ancestor_no_changelog_no_ssot", tag: "v9.9.9-rc.1", kind: provOK, changelog: false, ssot: "v9.0.0", wantExit: 0,
			wantOut: []string{rlsRcSkipNotice("v9.9.9-rc.1"), rlsRcPass("v9.9.9-rc.1")}},
		{name: "2_rc_trailer_commit_wrong", tag: "v9.9.9-rc.2", kind: provBadCommit, changelog: false, ssot: "v9.0.0", wantExit: 1,
			wantOut: rlsGateFail("check 4 (commit binding): trailer Release-commit='<sha>' != tagged commit '<sha>'.")},
		{name: "3_rc_commit_not_on_main", tag: "v9.9.9-rc.3", kind: provOffMain, changelog: false, ssot: "v9.0.0", wantExit: 1,
			wantOut: []string{rlsRcSkipNotice("v9.9.9-rc.3"),
				"::error::RELEASE_PROVENANCE_GATE: check 7 (main ancestry): tagged commit <sha> is not an ancestor of origin/main.", rlsGateTail}},
		{name: "d13_rc_check1_lightwseven_tag", tag: "v9.9.9-rc.4", kind: provLightwseven, changelog: false, ssot: "v9.0.0", wantExit: 1,
			wantOut: rlsGateFail("check 1 (annotated tag): 'v9.9.9-rc.4' is a commit object, not an annotated tag.")},
		{name: "d13_rc_check2_no_trailer", tag: "v9.9.9-rc.5", kind: provNoTrailer, changelog: false, ssot: "v9.0.0", wantExit: 1,
			wantOut: rlsGateFail("check 2 (provenance trailer): tag annotation has no 'Released-via: harness:release' line.")},
		{name: "d13_rc_check3_trailer_version_differs", tag: "v9.9.9-rc.6", kind: provBadVersion, changelog: false, ssot: "v9.0.0", wantExit: 1,
			wantOut: rlsGateFail("check 3 (version binding): trailer Release-version='v9.9.99' != pushed tag 'v9.9.9-rc.6'.")},
		// The rc rule covers the project's `-rc.N` form only. The legacy
		// undotted `-rcN` and other pre-release identifiers keep all seven checks.
		{name: "near_miss_legacy_undotted_rc_keeps_check5", tag: "v9.9.9-rc12", kind: provOK, changelog: false, ssot: "v9.9.9-rc12", wantExit: 1,
			wantOut: rlsGateFail("check 5 (CHANGELOG): CHANGELOG.md at <sha> has no '## [9.9.9-rc12]' or '## [v9.9.9-rc12]' section.")},
		{name: "near_miss_other_prerelease_keeps_check5", tag: "v9.9.9-beta.1", kind: provOK, changelog: false, ssot: "v9.9.9-beta.1", wantExit: 1,
			wantOut: rlsGateFail("check 5 (CHANGELOG): CHANGELOG.md at <sha> has no '## [9.9.9-beta.1]' or '## [v9.9.9-beta.1]' section.")},
	}
}

// rlsBuildProvenanceRepo builds one scratch repository carrying every fixture
// tag. Each tag points at its own commit on main (linear history), except the
// off-main kind, whose commit lives on a side branch that is never pushed.
func rlsBuildProvenanceRepo(t *testing.T, rows []rlsProvRow) *rlsRepo {
	t.Helper()
	r := rlsNewRepo(t)
	r.commit(map[string]string{
		"CHANGELOG.md":                      "# Changelog\n\n## [Unreleased]\n",
		".moai/config/sections/system.yaml": "system:\n  version: v0.0.0\n",
	}, "base")
	r.git("push", "-q", "origin", "main")

	for _, row := range rows {
		changelog := "# Changelog\n\n## [Unreleased]\n"
		if row.changelog {
			changelog += "\n## [" + strings.TrimPrefix(row.tag, "v") + "]\n\n- fixture\n"
		}
		files := map[string]string{
			"fixture.txt":                       row.tag + "\n", // keeps every fixture commit distinct
			"CHANGELOG.md":                      changelog,
			".moai/config/sections/system.yaml": "system:\n  version: " + row.ssot + "\n",
		}
		if row.kind == provOffMain {
			r.git("checkout", "-q", "-b", "side-"+row.tag, "main")
		}
		sha := r.commit(files, "fixture commit for "+row.tag)

		annotation := func(version, commit string, withTrailer bool) string {
			msg := "Release notes for " + row.tag + "\n"
			if withTrailer {
				msg += "\n" + rlsGateTrailerVia + "\nRelease-version: " + version + "\nRelease-commit: " + commit + "\n"
			}
			return msg
		}
		switch row.kind {
		case provLightwseven:
			r.git("tag", row.tag, sha)
		case provNoTrailer:
			r.git("tag", "-a", row.tag, "-m", annotation(row.tag, sha, false), sha)
		case provBadVersion:
			r.git("tag", "-a", row.tag, "-m", annotation("v9.9.99", sha, true), sha)
		case provBadCommit:
			r.git("tag", "-a", row.tag, "-m", annotation(row.tag, strings.Repeat("0", 40), true), sha)
		default: // provOK, provOffMain
			r.git("tag", "-a", row.tag, "-m", annotation(row.tag, sha, true), sha)
		}
		if row.kind == provOffMain {
			r.git("checkout", "-q", "main")
			r.git("push", "-q", "origin", "refs/tags/"+row.tag)
			continue
		}
		r.git("push", "-q", "origin", "main")
		r.git("push", "-q", "origin", "refs/tags/"+row.tag)
	}
	r.git("fetch", "-q", "origin", "+refs/heads/main:refs/remotes/origin/main")
	return r
}

func rlsRowOut(res rlsResult) []string {
	n := rlsNorm(res.out)
	if n == "" {
		return nil
	}
	return strings.Split(n, "\n")
}

func rlsSameLines(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// rlsCheckRow asserts exit code and the exact verdict lines of one row.
func rlsCheckRow(t *testing.T, row rlsProvRow, res rlsResult) {
	t.Helper()
	t.Logf("%s: exit=%d output=%q", row.tag, res.exit, rlsNorm(res.out))
	if res.exit != row.wantExit {
		t.Errorf("exit code = %d, want %d\n--- output ---\n%s", res.exit, row.wantExit, res.out)
	}
	if got := rlsRowOut(res); !rlsSameLines(got, row.wantOut) {
		t.Errorf("verdict lines differ\n got:  %q\n want: %q", got, row.wantOut)
	}
}

// rlsReplayStepBody runs the workflow step's own body (whatever it currently
// is) with TAG set, from the scratch clone, the way the hosted runner does:
// bash -e -o pipefail <file>.
func rlsReplayStepBody(t *testing.T, r *rlsRepo, tag string) rlsResult {
	t.Helper()
	step := rlsProvenanceStep(t)
	dir := t.TempDir()
	script := rlsWriteScript(t, dir, "step.sh", step.Run)
	return rlsRun(t, r.work, rlsEnv("TAG="+tag), "bash", "-e", "-o", "pipefail", script)
}

// rlsDirectProvenance calls the provenance script with the tag as its one
// argument. A missing script is the expected RED before the move, reported as
// a failure of the row, not a crash of the test.
func rlsDirectProvenance(t *testing.T, r *rlsRepo, tag string) rlsResult {
	t.Helper()
	return rlsRun(t, r.work, rlsEnv(), "bash", rlsProvenanceRel, tag)
}

// TestReleaseProvenanceStepReplaySuffixless is the characterization pin of the
// suffix-less gate: it replays the workflow step body on every formal-tag
// fixture and asserts the verdict the INLINE step printed before the move. It
// is green before and after the move; that is the equivalence proof.
func TestReleaseProvenanceStepReplaySuffixless(t *testing.T) {
	rlsRequireTools(t)
	rows := rlsSuffixlessRows()
	r := rlsBuildProvenanceRepo(t, rows)
	r.installRepoFile(rlsProvenanceRel)
	for _, row := range rows {
		t.Run(row.name, func(t *testing.T) {
			rlsCheckRow(t, row, rlsReplayStepBody(t, r, row.tag))
		})
	}
}

// TestReleaseProvenanceRcRule is AC-GFD-008's script limb: the six fixtures of
// the acceptance criterion plus the D13 carry-over rows and the near-miss tags,
// each judged by calling scripts/verify-release-provenance.sh <tag> directly.
func TestReleaseProvenanceRcRule(t *testing.T) {
	rlsRequireTools(t)
	if _, err := os.Stat(filepath.Join(findProjectRootForMirrorTest(t), filepath.FromSlash(rlsProvenanceRel))); err != nil {
		t.Fatalf("%s does not exist (the provenance logic is still inline in %s): %v", rlsProvenanceRel, releaseWorkflowRelPath, err)
	}
	rows := append(rlsSuffixlessRows()[:3:3], rlsRcRows()...)
	r := rlsBuildProvenanceRepo(t, rows)
	r.installRepoFile(rlsProvenanceRel)
	for _, row := range rows {
		t.Run(row.name, func(t *testing.T) {
			rlsCheckRow(t, row, rlsDirectProvenance(t, r, row.tag))
		})
	}
}

// provCallLine matches the one call line the workflow must carry.
var provCallLine = regexp.MustCompile(`(?m)^[ \t]*bash scripts/verify-release-provenance\.sh([^\n]*)$`)

// TestReleaseWorkflowInvokesProvenanceScript is AC-GFD-008's wiring limb
// (W1-W3): the shipped gate, not only the script, must change.
func TestReleaseWorkflowInvokesProvenanceScript(t *testing.T) {
	rlsRequireTools(t)
	root := findProjectRootForMirrorTest(t)
	raw, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(releaseWorkflowRelPath)))
	if err != nil {
		t.Fatalf("read %s: %v", releaseWorkflowRelPath, err)
	}
	whole := string(raw)
	step := rlsProvenanceStep(t)

	t.Run("W1_inline_checks_5_and_6_are_gone_from_the_workflow", func(t *testing.T) {
		if m := regexp.MustCompile(`check (5|6) \(`).FindString(whole); m != "" {
			t.Errorf("%s still carries an inline %q; checks 5 and 6 belong to %s", releaseWorkflowRelPath, m, rlsProvenanceRel)
		}
	})

	t.Run("W2_workflow_calls_the_script", func(t *testing.T) {
		if !strings.Contains(whole, "verify-release-provenance") {
			t.Errorf("%s never names verify-release-provenance", releaseWorkflowRelPath)
		}
	})

	calls := provCallLine.FindAllStringSubmatchIndex(step.Run, -1)
	t.Run("W3_call_line_passes_exactly_the_tag", func(t *testing.T) {
		if len(calls) != 1 {
			t.Fatalf("the step body must carry exactly one call line, found %d\n--- run body ---\n%s", len(calls), step.Run)
		}
		args := strings.TrimSpace(step.Run[calls[0][2]:calls[0][3]])
		if args != `"${TAG}"` {
			t.Errorf("call passes %q, want exactly %q (the same single <tag> argument the direct call uses)", args, `"${TAG}"`)
		}
		if got := step.Env["TAG"]; got != "${{ github.ref_name }}" {
			t.Errorf("step env TAG = %q, want %q", got, "${{ github.ref_name }}")
		}
	})

	t.Run("W3_network_fetches_stay_in_the_workflow_before_the_call", func(t *testing.T) {
		if len(calls) != 1 {
			t.Fatalf("the step body must carry exactly one call line (asserted above), found %d", len(calls))
		}
		callAt := calls[0][0]
		for _, fetch := range []string{
			`git fetch --force origin "refs/tags/${TAG}:refs/tags/${TAG}"`,
			`git fetch origin +refs/heads/main:refs/remotes/origin/main`,
		} {
			at := strings.Index(step.Run, fetch)
			if at < 0 {
				t.Errorf("the step body lost the fetch line %q", fetch)
			} else if at > callAt {
				t.Errorf("fetch line %q comes after the script call; the script reads the refs the fetch installs", fetch)
			}
		}
	})

	t.Run("D14_step_body_no_longer_references_what_the_checks_read", func(t *testing.T) {
		low := strings.ToLower(step.Run)
		for _, banned := range []string{"changelog", "system.yaml", "check 5", "check 6", "annotated tag", "git show"} {
			if strings.Contains(low, banned) {
				t.Errorf("the step body still references %q: an inline check survived next to the script call\n--- run body ---\n%s", banned, step.Run)
			}
		}
	})

	t.Run("W3_replayed_call_line_matches_the_direct_call_on_every_fixture", func(t *testing.T) {
		rows := append(rlsSuffixlessRows(), rlsRcRows()...)
		r := rlsBuildProvenanceRepo(t, rows)
		hasScript := r.installRepoFile(rlsProvenanceRel)
		for _, row := range rows {
			t.Run(row.name, func(t *testing.T) {
				replay := rlsReplayStepBody(t, r, row.tag)
				rlsCheckRow(t, row, replay)
				if !hasScript {
					t.Errorf("%s does not exist; there is no direct call to compare the replay with", rlsProvenanceRel)
					return
				}
				direct := rlsDirectProvenance(t, r, row.tag)
				if replay.exit != direct.exit || rlsNorm(replay.out) != rlsNorm(direct.out) {
					t.Errorf("replay and direct call disagree\n replay: exit=%d %q\n direct: exit=%d %q",
						replay.exit, rlsNorm(replay.out), direct.exit, rlsNorm(direct.out))
				}
			})
		}
	})
}

// TestProvenanceScriptNoQuietGrepUnderPipefail carries the SIGPIPE guard of
// release_workflow_pipefail_test.go over to the script the checks moved into:
// moving the code out of the guarded file must not move it out of the guard.
func TestProvenanceScriptNoQuietGrepUnderPipefail(t *testing.T) {
	t.Parallel()
	root := findProjectRootForMirrorTest(t)
	data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rlsProvenanceRel)))
	if err != nil {
		t.Fatalf("read %s: %v", rlsProvenanceRel, err)
	}
	var offenders []string
	for i, line := range strings.Split(string(data), "\n") {
		if m := pipeIntoGrep.FindStringSubmatch(line); m != nil && isQuietGrep(strings.Fields(m[1])) {
			offenders = append(offenders, fmt.Sprintf("%s  (%s:%d)", strings.TrimSpace(line), rlsProvenanceRel, i+1))
		}
	}
	if len(offenders) > 0 {
		t.Errorf("PIPEFAIL_SIGPIPE_HAZARD: %s pipes into `grep -q` under pipefail; use `| grep ... >/dev/null`.\n  %s",
			rlsProvenanceRel, strings.Join(offenders, "\n  "))
	}
}
