package factorylane

// merge_tag_shadow_t2144_test.go — card t2144 (follow-up of the t2121 audit
// debt G2): the merge-readiness checks consumed the integration name as a bare
// git revision, and git resolves a bare name tag-first, so a local tag named
// like the integration branch redirected every check to the tag's commit (the
// t2121 auditor's probe N5). The integration name now resolves explicitly:
// refs/heads/<name> when that branch exists, otherwise the name as given
// (origin/main, refs/...) — and refuses when neither resolves. Every fixture is
// a real throwaway git repository.

import (
	"strings"
	"testing"
)

func evalWithDevelop(t *testing.T, r specLessRepo, develop string) MergeCheckRun {
	t.Helper()
	run, err := EvaluateMergeTriple(MergeTripleInput{
		Lane: "lane-1", Card: r.card, Branch: "WT-card", Develop: develop, RepoDir: r.dir,
	}, ExecGitRunner{Dir: r.dir})
	if err != nil {
		t.Fatalf("EvaluateMergeTriple: %v", err)
	}
	return run
}

func checkOf(t *testing.T, run MergeCheckRun, name string) MergeCheck {
	t.Helper()
	for _, c := range run.Checks {
		if c.Name == name {
			return c
		}
	}
	t.Fatalf("no %s check recorded", name)
	return MergeCheck{}
}

func TestT2144TagShadow(t *testing.T) {
	t.Run("tree_identity_measures_the_branch_not_the_tag", func(t *testing.T) {
		r, _ := newSpecLessRepo(t)
		base := r.git(t, "rev-parse", "main")
		r.mainAdvance(t, "main_only.txt", "m\n") // main is ahead; the card did not absorb it
		r.git(t, "tag", "main", base)            // tag main -> OLD commit
		run := evalWithDevelop(t, r, "main")
		if c := checkOf(t, run, CheckTreeIdentity); c.Passed {
			t.Fatalf("tree-identity passed against the tag's old commit; want refusal against the branch: %s", c.Detail)
		}
	})

	t.Run("conflict_free_measures_the_branch_not_the_tag", func(t *testing.T) {
		r, _ := newSpecLessRepo(t)
		base := r.git(t, "rev-parse", "main")
		r.git(t, "checkout", "-q", "main")
		r.write(t, "card.txt", "main side\n")
		r.commit(t, "main touches card.txt")
		r.git(t, "checkout", "-q", "WT-card")
		r.git(t, "tag", "main", base)
		run := evalWithDevelop(t, r, "main")
		if c := checkOf(t, run, CheckConflictFree); c.Passed {
			t.Fatalf("conflict-free passed against the tag's old commit; want the branch's conflict: %s", c.Detail)
		}
	})

	t.Run("verdict_valid_only_against_the_branch_is_accepted", func(t *testing.T) {
		r, head := newSpecLessRepo(t)
		r.write(t, r.verdictRel(), verdictBody("PASS", head))
		r.commit(t, "evidence")
		base := r.git(t, "rev-parse", "main")
		r.mainAdvance(t, "main_only.txt", "m\n")
		r.git(t, "merge", "-q", "--no-ff", "-m", "absorb main", "refs/heads/main")
		r.git(t, "tag", "main", base)
		wantSyncAudit(t, evalWithDevelop(t, r, "main"), true, "verdict: PASS")
	})

	t.Run("forged_by_tag_absorb_of_an_unaudited_commit_is_refused", func(t *testing.T) {
		r, head := newSpecLessRepo(t)
		r.write(t, r.verdictRel(), verdictBody("PASS", head))
		r.commit(t, "evidence")
		r.git(t, "checkout", "-q", "-b", "other", "refs/heads/main")
		r.write(t, "evil.txt", "unaudited\n")
		evil := r.commit(t, "unaudited work")
		r.git(t, "checkout", "-q", "WT-card")
		r.git(t, "merge", "-q", "--no-ff", "-m", "merge other", "other")
		r.git(t, "tag", "main", evil) // the tag claims the unaudited commit is "main"
		wantSyncAudit(t, evalWithDevelop(t, r, "main"), false, "stale")
	})

	t.Run("no_tag_behaviour_unchanged", func(t *testing.T) {
		r, head := newSpecLessRepo(t)
		r.write(t, r.verdictRel(), verdictBody("PASS", head))
		r.commit(t, "evidence")
		r.mainAdvance(t, "main_only.txt", "m\n")
		r.absorbMain(t)
		run := evalWithDevelop(t, r, "main")
		wantSyncAudit(t, run, true, "verdict: PASS")
		if c := checkOf(t, run, CheckConflictFree); !c.Passed {
			t.Fatalf("conflict-free: %s", c.Detail)
		}
		if c := checkOf(t, run, CheckTreeIdentity); !c.Passed {
			t.Fatalf("tree-identity: %s", c.Detail)
		}
	})

	for _, name := range []string{"origin/main", "refs/heads/main"} {
		t.Run("qualified_name_still_works_"+strings.ReplaceAll(name, "/", "_"), func(t *testing.T) {
			r, head := newSpecLessRepo(t)
			r.write(t, r.verdictRel(), verdictBody("PASS", head))
			r.commit(t, "evidence")
			r.mainAdvance(t, "main_only.txt", "m\n")
			r.absorbMain(t)
			r.git(t, "update-ref", "refs/remotes/origin/main", "refs/heads/main")
			run := evalWithDevelop(t, r, name)
			wantSyncAudit(t, run, true, "verdict: PASS")
			if c := checkOf(t, run, CheckConflictFree); !c.Passed {
				t.Fatalf("conflict-free: %s", c.Detail)
			}
			if c := checkOf(t, run, CheckTreeIdentity); !c.Passed {
				t.Fatalf("tree-identity: %s", c.Detail)
			}
		})
	}

	t.Run("nonexistent_name_is_refused_naming_both_candidates", func(t *testing.T) {
		r, _ := newSpecLessRepo(t)
		run := evalWithDevelop(t, r, "no-such-branch")
		for _, name := range []string{CheckConflictFree, CheckTreeIdentity} {
			c := checkOf(t, run, name)
			if c.Passed {
				t.Fatalf("%s passed for a nonexistent integration name", name)
			}
			for _, want := range []string{"no-such-branch", "refs/heads/no-such-branch"} {
				if !strings.Contains(c.Detail, want) {
					t.Errorf("%s detail lacks %q: %s", name, want, c.Detail)
				}
			}
		}
	})
}
