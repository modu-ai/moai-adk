package factorylane

// merge_tag_shadow_repair_t2144_test.go — card t2144 repair round 1 (audit F1):
// integrationRevSHA must resolve by EXACT ref path only (refs/heads/<name>, then
// refs/remotes/<name>; a refs/... name verbatim). The earlier tag-first
// rev-parse fallback let a local tag named origin/main shadow the
// remote-tracking ref on the PR delivery path. Real throwaway git repositories.

import (
	"strings"
	"testing"
)

func resolveIntegration(r specLessRepo, name string) (string, string, bool) {
	return integrationRevSHA(ExecGitRunner{Dir: r.dir}, name)
}

func TestT2144TagShadowRepair(t *testing.T) {
	t.Run("P3_tag_origin_main_does_not_shadow_remote_tracking_tree_identity", func(t *testing.T) {
		r, _ := newSpecLessRepo(t)
		base := r.git(t, "rev-parse", "main")
		r.mainAdvance(t, "main_only.txt", "m\n")
		r.git(t, "update-ref", "refs/remotes/origin/main", "refs/heads/main")
		r.git(t, "tag", "origin/main", base) // tag -> OLD commit
		run := evalWithDevelop(t, r, "origin/main")
		if c := checkOf(t, run, CheckTreeIdentity); c.Passed {
			t.Fatalf("tree-identity measured the tag origin/main, not refs/remotes/origin/main: %s", c.Detail)
		}
	})

	t.Run("P3_tag_origin_main_does_not_shadow_remote_tracking_conflict_free", func(t *testing.T) {
		r, _ := newSpecLessRepo(t)
		base := r.git(t, "rev-parse", "main")
		r.git(t, "checkout", "-q", "main")
		r.write(t, "card.txt", "main side\n")
		r.commit(t, "main touches card.txt")
		r.git(t, "checkout", "-q", "WT-card")
		r.git(t, "update-ref", "refs/remotes/origin/main", "refs/heads/main")
		r.git(t, "tag", "origin/main", base)
		run := evalWithDevelop(t, r, "origin/main")
		if c := checkOf(t, run, CheckConflictFree); c.Passed {
			t.Fatalf("conflict-free measured the tag, not the remote-tracking ref: %s", c.Detail)
		}
	})

	t.Run("P3_tag_origin_main_does_not_forge_the_absorb_check", func(t *testing.T) {
		r, head := newSpecLessRepo(t)
		r.write(t, r.verdictRel(), verdictBody("PASS", head))
		r.commit(t, "evidence")
		r.git(t, "checkout", "-q", "-b", "other", "refs/heads/main")
		r.write(t, "evil.txt", "unaudited\n")
		evil := r.commit(t, "unaudited work")
		r.git(t, "checkout", "-q", "WT-card")
		r.git(t, "merge", "-q", "--no-ff", "-m", "merge other", "other")
		r.git(t, "update-ref", "refs/remotes/origin/main", "refs/heads/main")
		r.git(t, "tag", "origin/main", evil)
		wantSyncAudit(t, evalWithDevelop(t, r, "origin/main"), false, "stale")
	})

	t.Run("a_tag_named_like_remote_ref_still_resolves_the_remote_ref", func(t *testing.T) {
		r, _ := newSpecLessRepo(t)
		base := r.git(t, "rev-parse", "main")
		r.mainAdvance(t, "main_only.txt", "m\n")
		r.git(t, "update-ref", "refs/remotes/origin/main", "refs/heads/main")
		want := r.git(t, "rev-parse", "refs/remotes/origin/main")
		r.git(t, "tag", "origin/main", base)
		got, why, ok := resolveIntegration(r, "origin/main")
		if !ok || got != want {
			t.Fatalf("got ok=%v sha=%s why=%q; want %s", ok, got, why, want)
		}
	})

	t.Run("b_only_a_tag_exists_is_refused_naming_both_forms", func(t *testing.T) {
		r, _ := newSpecLessRepo(t)
		r.git(t, "tag", "only-tag", r.git(t, "rev-parse", "main"))
		_, why, ok := resolveIntegration(r, "only-tag")
		if ok {
			t.Fatal("a tag alone must not resolve as the integration ref")
		}
		for _, want := range []string{"refs/heads/only-tag", "refs/remotes/only-tag"} {
			if !strings.Contains(why, want) {
				t.Errorf("refusal lacks %q: %s", want, why)
			}
		}
	})

	t.Run("c_suffixed_names_fail_closed", func(t *testing.T) {
		r, _ := newSpecLessRepo(t)
		for _, name := range []string{"main~0", "main^", "main@{0}", "--all"} {
			if sha, _, ok := resolveIntegration(r, name); ok {
				t.Errorf("%q resolved to %s; want fail-closed refusal", name, sha)
			}
		}
	})

	t.Run("d_exact_forms_resolve", func(t *testing.T) {
		r, _ := newSpecLessRepo(t)
		r.git(t, "update-ref", "refs/remotes/origin/main", "refs/heads/main")
		want := r.git(t, "rev-parse", "refs/heads/main")
		r.git(t, "tag", "main", r.git(t, "rev-parse", "HEAD")) // shadowing tag
		for _, name := range []string{"main", "origin/main", "refs/heads/main", "refs/remotes/origin/main"} {
			got, why, ok := resolveIntegration(r, name)
			if !ok || got != want {
				t.Errorf("%q: ok=%v sha=%s why=%q; want %s", name, ok, got, why, want)
			}
		}
	})

	// F2: a local branch literally named origin/main wins over the
	// remote-tracking ref, as in git itself; pinned so a change is deliberate.
	t.Run("f2_local_branch_named_origin_main_wins_like_git", func(t *testing.T) {
		r, _ := newSpecLessRepo(t)
		r.git(t, "update-ref", "refs/remotes/origin/main", "refs/heads/main")
		r.git(t, "branch", "origin/main", "WT-card")
		want := r.git(t, "rev-parse", "refs/heads/origin/main")
		got, _, ok := resolveIntegration(r, "origin/main")
		if !ok || got != want {
			t.Fatalf("got ok=%v sha=%s; want local branch %s", ok, got, want)
		}
	})
}
