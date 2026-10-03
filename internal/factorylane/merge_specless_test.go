package factorylane

// merge_specless_test.go — SPEC-GITHUB-FLOW-DEFAULT-001 M2-B follow-up (card
// t1453, C5): the sync-audit condition of the merge-readiness triple for a card
// that has no SPEC. A SPEC-less card (Class A/B) closes on its audit verdict
// file, `.moai/reports/<card-id>/verdict.md`; with an empty SpecDir the
// condition reads that file instead of a SPEC progress.md, and a card with a
// SPEC keeps the §E.4 reader byte for byte. Every fixture is a real throwaway
// git repository, because the verdict's audited commit is judged against the
// card's real HEAD.

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

type specLessRepo struct {
	dir  string
	card string
}

func (r specLessRepo) git(t *testing.T, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = r.dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
	return strings.TrimSpace(string(out))
}

func (r specLessRepo) write(t *testing.T, rel, body string) {
	t.Helper()
	path := filepath.Join(r.dir, rel)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func (r specLessRepo) commit(t *testing.T, msg string) string {
	t.Helper()
	r.git(t, "add", "-A")
	r.git(t, "commit", "-q", "-m", msg)
	return r.git(t, "rev-parse", "HEAD")
}

func (r specLessRepo) verdictRel() string {
	return filepath.Join(".moai", "reports", r.card, "verdict.md")
}

// newSpecLessRepo builds main with one base commit and a card branch carrying
// one card commit, checked out; the returned sha is the card's HEAD.
func newSpecLessRepo(t *testing.T) (specLessRepo, string) {
	t.Helper()
	r := specLessRepo{dir: t.TempDir(), card: "t9"}
	r.git(t, "init", "-q", "-b", "main")
	r.git(t, "config", "user.email", "test@example.com")
	r.git(t, "config", "user.name", "Test")
	r.write(t, "base.txt", "base\n")
	r.commit(t, "base")
	r.git(t, "checkout", "-q", "-b", "WT-card")
	r.write(t, "card.txt", "card\n")
	return r, r.commit(t, "card work")
}

func evalSpecLess(t *testing.T, r specLessRepo, specDir string) MergeCheckRun {
	t.Helper()
	run, err := EvaluateMergeTriple(MergeTripleInput{
		Lane: "lane-1", Card: r.card, SpecDir: specDir,
		Branch: "WT-card", Develop: "main", RepoDir: r.dir,
	}, ExecGitRunner{Dir: r.dir})
	if err != nil {
		t.Fatalf("EvaluateMergeTriple: %v", err)
	}
	return run
}

func verdictBody(verdict, sha string) string {
	return "# Verdict\n\nverdict: " + verdict + "\naudited_sha: " + sha + "\n"
}

func TestMergeTripleSpecLessVerdict(t *testing.T) {
	const wantPath = ".moai/reports/t9/verdict.md"
	cases := []struct {
		name  string
		setup func(t *testing.T, r specLessRepo, head string) (specDir string)
		pass  bool
		// wants are substrings the sync-audit detail must carry (pass or fail).
		wants []string
	}{
		{
			name: "spec_card_unchanged_ignores_a_failing_verdict_file",
			setup: func(t *testing.T, r specLessRepo, head string) string {
				specDir := filepath.Join(r.dir, ".moai", "specs", "SPEC-X-001")
				r.write(t, filepath.Join(".moai", "specs", "SPEC-X-001", "progress.md"),
					"## §E.4 Sync-phase Audit-Ready Signal\n\nsync_status: complete\n")
				r.write(t, r.verdictRel(), verdictBody("FAIL", head))
				return specDir
			},
			pass:  true,
			wants: []string{"§E.4 sync_status: complete"},
		},
		{
			name: "specless_pass_bound_to_head",
			setup: func(t *testing.T, r specLessRepo, head string) string {
				r.write(t, r.verdictRel(), verdictBody("PASS", head))
				return ""
			},
			pass:  true,
			wants: []string{wantPath, "PASS"},
		},
		{
			name: "specless_pass_with_debt_bound_to_head",
			setup: func(t *testing.T, r specLessRepo, head string) string {
				r.write(t, r.verdictRel(), verdictBody("PASS-WITH-DEBT", head))
				return ""
			},
			pass:  true,
			wants: []string{wantPath, "PASS-WITH-DEBT"},
		},
		{
			name: "specless_pass_audited_commit_precedes_the_evidence_only_commit",
			setup: func(t *testing.T, r specLessRepo, head string) string {
				r.write(t, r.verdictRel(), verdictBody("PASS", head))
				r.commit(t, "evidence: verdict for the audited commit")
				return ""
			},
			pass:  true,
			wants: []string{wantPath},
		},
		{
			name: "specless_absent_file",
			setup: func(t *testing.T, r specLessRepo, head string) string {
				return ""
			},
			pass:  false,
			wants: []string{wantPath, "does not exist", "verdict: PASS", "audited_sha:"},
		},
		{
			name: "specless_unreadable_file",
			setup: func(t *testing.T, r specLessRepo, head string) string {
				if runtime.GOOS == "windows" || os.Geteuid() == 0 {
					t.Skip("file permissions cannot make a file unreadable here")
				}
				r.write(t, r.verdictRel(), verdictBody("PASS", head))
				if err := os.Chmod(filepath.Join(r.dir, r.verdictRel()), 0o000); err != nil {
					t.Fatal(err)
				}
				t.Cleanup(func() { _ = os.Chmod(filepath.Join(r.dir, r.verdictRel()), 0o644) })
				return ""
			},
			pass:  false,
			wants: []string{wantPath, "unreadable"},
		},
		{
			name: "specless_verdict_path_is_a_directory",
			setup: func(t *testing.T, r specLessRepo, head string) string {
				if err := os.MkdirAll(filepath.Join(r.dir, r.verdictRel()), 0o755); err != nil {
					t.Fatal(err)
				}
				return ""
			},
			pass:  false,
			wants: []string{wantPath, "not a regular file"},
		},
		{
			name: "specless_no_verdict_line",
			setup: func(t *testing.T, r specLessRepo, head string) string {
				r.write(t, r.verdictRel(), "# Verdict\n\naudited_sha: "+head+"\n")
				return ""
			},
			pass:  false,
			wants: []string{wantPath, "no `verdict:` line"},
		},
		{
			name: "specless_verdict_fail",
			setup: func(t *testing.T, r specLessRepo, head string) string {
				r.write(t, r.verdictRel(), verdictBody("FAIL", head))
				return ""
			},
			pass:  false,
			wants: []string{wantPath, "verdict: FAIL"},
		},
		{
			name: "specless_conflicting_verdict_lines",
			setup: func(t *testing.T, r specLessRepo, head string) string {
				r.write(t, r.verdictRel(), verdictBody("PASS", head)+"verdict: FAIL\n")
				return ""
			},
			pass:  false,
			wants: []string{wantPath, "conflicting `verdict:` lines"},
		},
		{
			name: "specless_no_audited_sha_line",
			setup: func(t *testing.T, r specLessRepo, head string) string {
				r.write(t, r.verdictRel(), "verdict: PASS\n")
				return ""
			},
			pass:  false,
			wants: []string{wantPath, "no `audited_sha:` line"},
		},
		{
			name: "specless_pass_with_a_stale_audited_sha",
			setup: func(t *testing.T, r specLessRepo, head string) string {
				r.write(t, r.verdictRel(), verdictBody("PASS", head))
				r.write(t, "card.txt", "card, changed after the audit\n")
				r.commit(t, "code changed after the audited commit")
				return ""
			},
			pass:  false,
			wants: []string{wantPath, "audited_sha", "stale", "card.txt"},
		},
		{
			name: "specless_audited_sha_names_no_commit",
			setup: func(t *testing.T, r specLessRepo, head string) string {
				r.write(t, r.verdictRel(), verdictBody("PASS", strings.Repeat("a", 40)))
				return ""
			},
			pass:  false,
			wants: []string{wantPath, "audited_sha", "does not name a commit"},
		},
		{
			name: "specless_audited_sha_is_not_an_ancestor_of_head",
			setup: func(t *testing.T, r specLessRepo, head string) string {
				r.git(t, "checkout", "-q", "-b", "other", "main")
				r.write(t, "other.txt", "other\n")
				other := r.commit(t, "unrelated commit")
				r.git(t, "checkout", "-q", "WT-card")
				r.write(t, r.verdictRel(), verdictBody("PASS", other))
				return ""
			},
			pass:  false,
			wants: []string{wantPath, "audited_sha", "not an ancestor"},
		},
	}
	visited := 0
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			visited++
			r, head := newSpecLessRepo(t)
			specDir := tc.setup(t, r, head)
			run := evalSpecLess(t, r, specDir)
			sync := run.Checks[0]
			if sync.Name != CheckSyncAudit {
				t.Fatalf("first check is %q, want %q", sync.Name, CheckSyncAudit)
			}
			if sync.Passed != tc.pass {
				t.Fatalf("sync-audit passed=%v, want %v\ndetail: %s", sync.Passed, tc.pass, sync.Detail)
			}
			for _, want := range tc.wants {
				if !strings.Contains(sync.Detail, want) {
					t.Errorf("detail lacks %q:\n%s", want, sync.Detail)
				}
			}
			if tc.pass && !run.AllPassed {
				t.Errorf("a passing sync-audit left the triple failing %s: %+v", run.FailedCondition, run.Checks)
			}
			if !tc.pass && run.FailedCondition != CheckSyncAudit {
				t.Errorf("failed_condition=%q, want %q", run.FailedCondition, CheckSyncAudit)
			}
		})
	}
	if visited != len(cases) {
		t.Fatalf("visited %d rows, want %d", visited, len(cases))
	}
}

// A card id that is not a safe path segment is refused before any path is built.
func TestMergeTripleSpecLessRefusesUnsafeCardID(t *testing.T) {
	for _, card := range []string{"", "../t9", "a/b", "-x"} {
		r, head := newSpecLessRepo(t)
		r.write(t, r.verdictRel(), verdictBody("PASS", head))
		r.card = card
		run := evalSpecLess(t, r, "")
		if run.Checks[0].Passed {
			t.Errorf("card %q: sync-audit passed on an unsafe card id", card)
		}
	}
}

// The refusal text names the file to produce and both machine lines, whatever
// the cause: pinned verbatim for the absent-file cause.
func TestMergeTripleSpecLessRefusalText(t *testing.T) {
	r, _ := newSpecLessRepo(t)
	run := evalSpecLess(t, r, "")
	want := "card t9 has no SPEC, so its merge evidence is .moai/reports/t9/verdict.md — " + filepath.Join(r.dir, r.verdictRel()) + " does not exist. " +
		"Produce that file on the card branch with the machine lines `verdict: PASS` (or `verdict: PASS-WITH-DEBT`) and `audited_sha: <full commit SHA the audit read>` " +
		"(the audited commit must be HEAD, or differ from HEAD only under .moai/reports/t9/)"
	if got := run.Checks[0].Detail; got != want {
		t.Errorf("refusal text:\n got: %s\nwant: %s", got, want)
	}
}
