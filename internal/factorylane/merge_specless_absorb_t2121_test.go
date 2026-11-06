package factorylane

// merge_specless_absorb_t2121_test.go — card t2121: a SPEC-less card's audit
// verdict survives the integration branch being absorbed into the card branch
// (a --no-ff merge of main) when the card's own diff is byte-identical, and the
// reader also accepts the convention name sync-audit*.md. Every fixture is a
// real throwaway git repository.

import (
	"strings"
	"testing"
)

// mainAdvance adds one commit to main and returns to the card branch.
func (r specLessRepo) mainAdvance(t *testing.T, rel, body string) {
	t.Helper()
	r.git(t, "checkout", "-q", "main")
	r.write(t, rel, body)
	r.commit(t, "main advance "+rel)
	r.git(t, "checkout", "-q", "WT-card")
}

func (r specLessRepo) absorbMain(t *testing.T) {
	t.Helper()
	r.git(t, "merge", "-q", "--no-ff", "-m", "absorb main", "main")
}

func wantSyncAudit(t *testing.T, run MergeCheckRun, pass bool, wants ...string) {
	t.Helper()
	for _, c := range run.Checks {
		if c.Name != CheckSyncAudit {
			continue
		}
		if c.Passed != pass {
			t.Fatalf("sync-audit passed=%v, want %v; detail: %s", c.Passed, pass, c.Detail)
		}
		for _, w := range wants {
			if !strings.Contains(c.Detail, w) {
				t.Errorf("detail lacks %q: %s", w, c.Detail)
			}
		}
		return
	}
	t.Fatal("no sync-audit check recorded")
}

func TestT2121AbsorbTolerance(t *testing.T) {
	t.Run("clean_main_absorb_keeps_the_verdict", func(t *testing.T) {
		r, head := newSpecLessRepo(t)
		r.write(t, r.verdictRel(), verdictBody("PASS", head))
		r.commit(t, "evidence")
		r.mainAdvance(t, "main_only.txt", "m\n")
		r.absorbMain(t)
		wantSyncAudit(t, evalSpecLess(t, r, ""), true, "verdict: PASS")
	})
	t.Run("absorb_when_main_touched_another_file_and_evidence_added_after", func(t *testing.T) {
		r, head := newSpecLessRepo(t)
		r.mainAdvance(t, "main_only.txt", "m\n")
		r.absorbMain(t)
		r.write(t, r.verdictRel(), verdictBody("PASS", head))
		r.commit(t, "evidence after absorb")
		wantSyncAudit(t, evalSpecLess(t, r, ""), true)
	})
	t.Run("card_code_changed_after_audit_refuses", func(t *testing.T) {
		r, head := newSpecLessRepo(t)
		r.write(t, r.verdictRel(), verdictBody("PASS", head))
		r.commit(t, "evidence")
		r.mainAdvance(t, "main_only.txt", "m\n")
		r.absorbMain(t)
		r.write(t, "card.txt", "card changed\n")
		r.commit(t, "late code change")
		wantSyncAudit(t, evalSpecLess(t, r, ""), false, "stale")
	})
	t.Run("late_code_change_without_absorb_refuses", func(t *testing.T) {
		r, head := newSpecLessRepo(t)
		r.write(t, r.verdictRel(), verdictBody("PASS", head))
		r.write(t, "card.txt", "card changed\n")
		r.commit(t, "code + evidence")
		wantSyncAudit(t, evalSpecLess(t, r, ""), false, "stale")
	})
	t.Run("evil_merge_edits_a_card_owned_file", func(t *testing.T) {
		r, head := newSpecLessRepo(t)
		r.write(t, r.verdictRel(), verdictBody("PASS", head))
		r.commit(t, "evidence")
		r.mainAdvance(t, "main_only.txt", "m\n")
		r.git(t, "merge", "-q", "--no-ff", "--no-commit", "main")
		r.write(t, "card.txt", "card plus conflict-resolution edit\n")
		r.git(t, "add", "-A")
		r.git(t, "commit", "-q", "-m", "absorb main with edit")
		wantSyncAudit(t, evalSpecLess(t, r, ""), false, "stale")
	})
	t.Run("evil_merge_edits_a_path_outside_the_card_diff", func(t *testing.T) {
		r, head := newSpecLessRepo(t)
		r.write(t, r.verdictRel(), verdictBody("PASS", head))
		r.commit(t, "evidence")
		r.mainAdvance(t, "main_only.txt", "m\n")
		r.git(t, "merge", "-q", "--no-ff", "--no-commit", "main")
		r.write(t, "base.txt", "base smuggled\n")
		r.git(t, "add", "-A")
		r.git(t, "commit", "-q", "-m", "absorb main with smuggled edit")
		wantSyncAudit(t, evalSpecLess(t, r, ""), false, "stale")
	})
	t.Run("merge_of_a_non_main_branch_refuses", func(t *testing.T) {
		r, head := newSpecLessRepo(t)
		r.write(t, r.verdictRel(), verdictBody("PASS", head))
		r.commit(t, "evidence")
		r.git(t, "checkout", "-q", "-b", "other", "main")
		r.write(t, "other.txt", "o\n")
		r.commit(t, "other work")
		r.git(t, "checkout", "-q", "WT-card")
		r.git(t, "merge", "-q", "--no-ff", "-m", "merge other", "other")
		wantSyncAudit(t, evalSpecLess(t, r, ""), false, "stale")
	})
	t.Run("unresolvable_integration_ref_fails_closed", func(t *testing.T) {
		r, head := newSpecLessRepo(t)
		r.write(t, r.verdictRel(), verdictBody("PASS", head))
		r.commit(t, "evidence")
		r.mainAdvance(t, "main_only.txt", "m\n")
		r.absorbMain(t)
		run, err := EvaluateMergeTriple(MergeTripleInput{
			Lane: "lane-1", Card: r.card, Branch: "WT-card", Develop: "no-such-ref", RepoDir: r.dir,
		}, ExecGitRunner{Dir: r.dir})
		if err != nil {
			t.Fatal(err)
		}
		wantSyncAudit(t, run, false, "stale")
	})
	t.Run("regression_audited_equals_head", func(t *testing.T) {
		r, head := newSpecLessRepo(t)
		r.write(t, r.verdictRel(), verdictBody("PASS", head))
		wantSyncAudit(t, evalSpecLess(t, r, ""), true)
	})
	t.Run("regression_evidence_dir_only_change", func(t *testing.T) {
		r, head := newSpecLessRepo(t)
		r.write(t, r.verdictRel(), verdictBody("PASS", head))
		r.commit(t, "evidence")
		wantSyncAudit(t, evalSpecLess(t, r, ""), true)
	})
}

func TestT2121EvidenceFileNames(t *testing.T) {
	syncRel := func(r specLessRepo, name string) string { return ".moai/reports/" + r.card + "/" + name }
	t.Run("only_sync_audit_md_passes", func(t *testing.T) {
		r, head := newSpecLessRepo(t)
		r.write(t, syncRel(r, "sync-audit.md"), verdictBody("PASS", head))
		wantSyncAudit(t, evalSpecLess(t, r, ""), true, "sync-audit.md")
	})
	t.Run("highest_iter_is_read", func(t *testing.T) {
		r, head := newSpecLessRepo(t)
		r.write(t, syncRel(r, "sync-audit.md"), verdictBody("FAIL", head))
		r.write(t, syncRel(r, "sync-audit-iter2.md"), verdictBody("PASS", head))
		wantSyncAudit(t, evalSpecLess(t, r, ""), true, "sync-audit-iter2.md")
	})
	t.Run("iter10_beats_iter2", func(t *testing.T) {
		r, head := newSpecLessRepo(t)
		r.write(t, syncRel(r, "sync-audit-iter2.md"), verdictBody("FAIL", head))
		r.write(t, syncRel(r, "sync-audit-iter10.md"), verdictBody("PASS", head))
		wantSyncAudit(t, evalSpecLess(t, r, ""), true, "sync-audit-iter10.md")
	})
	t.Run("verdict_md_wins_over_sync_audit", func(t *testing.T) {
		r, head := newSpecLessRepo(t)
		r.write(t, r.verdictRel(), verdictBody("FAIL", head))
		r.write(t, syncRel(r, "sync-audit-iter2.md"), verdictBody("PASS", head))
		wantSyncAudit(t, evalSpecLess(t, r, ""), false, "verdict.md")
	})
	t.Run("absent_both_names_both", func(t *testing.T) {
		r, _ := newSpecLessRepo(t)
		wantSyncAudit(t, evalSpecLess(t, r, ""), false, "verdict.md", "sync-audit")
	})
}
