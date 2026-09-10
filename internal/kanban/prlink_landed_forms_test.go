// prlink_landed_forms_test.go — SPEC-TODO-LANDING-ATTRIBUTION-001 §B criteria
// (M1: the attribution predicate, REQ-TLA-001..006, REQ-TLA-013).
//
// Every test drives the querier's PUBLIC surface (GitLandedQuerier.Landed)
// over fixture repositories whose history is exactly the falsifying subjects
// acceptance.md §A names, so the assertion is about the predicate's verdict
// rather than about a transcription of it. Each criterion is asserted in BOTH
// directions: a title-attributed commit reads landed, and a body-mention-only
// or other-card-attributed commit reads not-landed — a one-directional suite
// lets "everything is landed" pass.
package kanban

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// formCommit is one fixture commit: a subject and an optional body.
type formCommit struct {
	subject string
	body    string
}

// formRepo builds a throwaway repository whose history is exactly entries
// (one commit per entry, oldest first) and whose refs/remotes/origin/<branch>
// names the final commit — the ref the landed question is asked about.
func formRepo(t *testing.T, branch string, entries ...formCommit) string {
	t.Helper()
	dir := t.TempDir()
	gitCfg := filepath.Join(dir, "gitconfig")
	run := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		cmd.Env = append(cmd.Environ(),
			"GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@example.com",
			"GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@example.com",
			"GIT_CONFIG_GLOBAL="+gitCfg, "GIT_CONFIG_NOSYSTEM=1",
		)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, out)
		}
	}
	run("init", "-q", "-b", "main")
	for i, e := range entries {
		file := "f" + string(rune('a'+i)) + ".txt"
		args := []string{"commit", "-q", "-m", e.subject}
		if e.body != "" {
			args = append(args, "-m", e.body)
		}
		if err := os.WriteFile(filepath.Join(dir, file), []byte(e.subject+"\n"), 0o600); err != nil {
			t.Fatalf("write %s: %v", file, err)
		}
		run("add", file)
		run(args...)
	}
	run("update-ref", "refs/remotes/origin/"+branch, "HEAD")
	return dir
}

// landed asks the querier built over dir's git and asserts the answer equals
// want, naming the subject whose verdict is under test in the failure message.
func landed(t *testing.T, dir, ref, cardID string) LandingAnswer {
	t.Helper()
	answer, err := (GitLandedQuerier{Run: gitIn(dir), Ref: ref}).Landed(cardID)
	if err != nil {
		t.Fatalf("Landed(%s) against %s: %v", cardID, ref, err)
	}
	return answer
}

// AC-TLA-001 — conventional-commit scope, both directions. The t237 fixture
// is MUT-WHOLE-MESSAGE's falsifier: today's whole-message grep reads two t230
// commits' BODY mention of t237 as t237 having landed.
func TestLandedPredicate_Form1_ConventionalScope(t *testing.T) {
	dir := formRepo(t, "develop",
		formCommit{subject: "chore: seed the tree"},
		formCommit{subject: "docs(t440): record develop-absorb re-measure evidence"},
	)
	if got := landed(t, dir, "origin/develop", "t440"); got != LandingLanded {
		t.Errorf("t440 = %q, want %q — scope position attributes", got, LandingLanded)
	}

	// Red direction: the body mention attributes nothing. The body text is
	// transcribed from the measured t237 false positive (spec.md §A.1).
	dir = formRepo(t, "main",
		formCommit{
			subject: "docs(t230): t230 sync-audit evidence",
			body:    "the card about to change the hook body is t237. Re-measured: the issue is OPEN",
		},
	)
	if got := landed(t, dir, "origin/main", "t237"); got != LandingNotLanded {
		t.Errorf("t237 = %q, want %q — a body mention is not an attribution", got, LandingNotLanded)
	}
	if got := landed(t, dir, "origin/main", "t230"); got != LandingLanded {
		t.Errorf("t230 = %q, want %q — the commit's own scope still attributes", got, LandingLanded)
	}
}

// AC-TLA-002 clauses 1-2 — trailing parenthetical, and MUT-SUBJECT-ONLY's
// t443 falsifier: the token sits mid-subject while the subject's own trailing
// attribution is another card's.
func TestLandedPredicate_Form2_TrailingParenthetical(t *testing.T) {
	dir := formRepo(t, "develop",
		formCommit{subject: "chore: seed the tree"},
		formCommit{subject: "docs: refresh catalog moai whole-tree hash (t401)"},
	)
	if got := landed(t, dir, "origin/develop", "t401"); got != LandingLanded {
		t.Errorf("t401 = %q, want %q — the trailing group attributes", got, LandingLanded)
	}

	dir = formRepo(t, "develop",
		formCommit{subject: "chore: seed the tree"},
		formCommit{subject: "chore(catalog): revert sync-auditor hash to develop value — t443 jurisdiction (t461)"},
	)
	if got := landed(t, dir, "origin/develop", "t443"); got != LandingNotLanded {
		t.Errorf("t443 = %q, want %q — a mid-subject mention attributes nothing (MUT-SUBJECT-ONLY)", got, LandingNotLanded)
	}
	// §D tiebreak, both directions: the trailing group wins over the scope.
	if got := landed(t, dir, "origin/develop", "t461"); got != LandingLanded {
		t.Errorf("t461 = %q, want %q — the trailing group outranks the scope", got, LandingLanded)
	}
}

// AC-TLA-002 clause 3 — form 2b, MUT-PAREN-END-ANCHOR's t210 falsifier: the
// card-bearing group is followed by a pull-request reference group, so form
// 2's end-of-subject anchor misses the repository's dominant PR landing path.
func TestLandedPredicate_Form2b_ReferenceGroup(t *testing.T) {
	dir := formRepo(t, "develop",
		formCommit{subject: "chore: seed the tree"},
		formCommit{subject: "feat(kanban): moai todo pr — read-only card-to-PR and landed link (t210) (#1628)"},
	)
	if got := landed(t, dir, "origin/develop", "t210"); got != LandingLanded {
		t.Errorf("t210 = %q, want %q — form 2b: a card group before one reference group (MUT-NO-FORM-2B)", got, LandingLanded)
	}
}

// AC-TLA-003 — merge subjects, six directions.
func TestLandedPredicate_Form3_Merges(t *testing.T) {
	t.Run("clause 1 and 2", func(t *testing.T) {
		dir := formRepo(t, "develop",
			formCommit{subject: "chore: seed the tree"},
			formCommit{subject: "Merge branch 'WT-kanban-fix' into develop (card t263)"},
			formCommit{subject: "docs(t263): die-at-exit reproduced (0/5) — remedy sequenced behind t216"},
		)
		if got := landed(t, dir, "origin/develop", "t263"); got != LandingLanded {
			t.Errorf("t263 = %q, want %q — integration-targeted merge with a single-card group", got, LandingLanded)
		}
		if got := landed(t, dir, "origin/develop", "t216"); got != LandingNotLanded {
			t.Errorf("t216 = %q, want %q — the other card's subject attributes t263, not t216", got, LandingNotLanded)
		}
	})

	t.Run("clause 3 — absorb-direction merges attribute nothing", func(t *testing.T) {
		dir := formRepo(t, "develop",
			formCommit{subject: "chore: seed the tree"},
			formCommit{subject: "Merge branch 'WT-audit-evidence-store' into WT-audit-advice-integrity (t387 depends on t386 convention doc)"},
			formCommit{subject: "Merge origin/develop into WT-audit-participant-count — absorb upstream before integration (card t284)"},
		)
		for _, card := range []string{"t386", "t387", "t284"} {
			if got := landed(t, dir, "origin/develop", card); got != LandingNotLanded {
				t.Errorf("%s = %q, want %q — a merge into a non-landed target is an absorb record (MUT-MERGE-ANY-TOKEN)", card, got, LandingNotLanded)
			}
		}
	})

	t.Run("clause 4 — form 3b's green direction", func(t *testing.T) {
		dir := formRepo(t, "develop",
			formCommit{subject: "chore: seed the tree"},
			formCommit{subject: "Merge branch 'WT-mx-tag-edges' into develop (card t412 — SPEC-MX-TAG-EDGES-001)"},
		)
		if got := landed(t, dir, "origin/develop", "t412"); got != LandingLanded {
			t.Errorf("t412 = %q, want %q — the trailing group carries text after the id, so form 2's anchor misses it (MUT-NO-FORM-3B)", got, LandingLanded)
		}
	})

	t.Run("clause 5 — absorb siblings of a landed target", func(t *testing.T) {
		dir := formRepo(t, "develop",
			formCommit{subject: "chore: seed the tree"},
			formCommit{subject: "Merge branch 'origin/develop' into WT-mx-tag-edges (window absorption, card t412)"},
			formCommit{subject: "Merge branch 'origin/develop' into WT-mx-tag-edges (pre-run absorption, card t412)"},
			formCommit{subject: "Merge branch 'WT-edge-confidence' into WT-mx-tag-edges (card t412 dependency absorption)"},
		)
		if got := landed(t, dir, "origin/develop", "t412"); got != LandingNotLanded {
			t.Errorf("t412 = %q, want %q — same card, wrong target: only the integration-targeted merge lands (MUT-NO-TARGET-TEST)", got, LandingNotLanded)
		}
	})

	t.Run("clause 6 — the target is derived from the resolved ref, never spelled", func(t *testing.T) {
		// The resolved landed ref names release/v9, not develop. The
		// attributing target must move with it (REQ-TLA-013,
		// MUT-HARDCODED-DEVELOP's only falsifier).
		dir := formRepo(t, "release/v9",
			formCommit{subject: "chore: seed the tree"},
			formCommit{subject: "Merge branch 'WT-x' into release/v9 (card t900 — note)"},
			formCommit{subject: "Merge branch 'WT-y' into develop (card t901 — note)"},
		)
		if got := landed(t, dir, "origin/release/v9", "t900"); got != LandingLanded {
			t.Errorf("t900 = %q, want %q — the merge target matches the branch the resolved ref names", got, LandingLanded)
		}
		if got := landed(t, dir, "origin/release/v9", "t901"); got != LandingNotLanded {
			t.Errorf("t901 = %q, want %q — a develop merge is not a landing when the resolved ref names release/v9", got, LandingNotLanded)
		}
	})
}

// AC-TLA-003b — the card-led merge shapes (forms 3a and 3c).
func TestLandedPredicate_Form3a_3c_CardLedMerges(t *testing.T) {
	dir := formRepo(t, "develop",
		formCommit{subject: "chore: seed the tree"},
		formCommit{subject: "Merge card t244 (WT-team-ac-verify-wiring) into develop: keep-dormant verdict"},
		formCommit{subject: "docs(hooks): t244 verdict — keep team-ac-verify dormant"},
	)
	if got := landed(t, dir, "origin/develop", "t244"); got != LandingLanded {
		t.Errorf("t244 = %q, want %q — the card is the first token after the merge verb (MUT-NO-FORM-3A)", got, LandingLanded)
	}
	// The same card's other subject: the scope names a package and t244 sits
	// mid-sentence — not an attribution.
	if got := landed(t, dir, "origin/develop", "t999"); got != LandingNotLanded {
		t.Errorf("t999 = %q, want %q — control: an absent card is not-landed", got, LandingNotLanded)
	}

	dir = formRepo(t, "develop",
		formCommit{subject: "chore: seed the tree"},
		formCommit{subject: "merge: t79 — glm_task delegation family (branch WT-t80)"},
	)
	if got := landed(t, dir, "origin/develop", "t79"); got != LandingLanded {
		t.Errorf("t79 = %q, want %q — the merge: spelling of the card-led shape (MUT-NO-FORM-3C)", got, LandingLanded)
	}
	// MUT-FIRST-TOKEN-OF-GROUP's falsifier in the same fixture: the ONLY card
	// token in the trailing group is a branch name, and a branch name is a
	// place, not an attribution.
	if got := landed(t, dir, "origin/develop", "t80"); got != LandingNotLanded {
		t.Errorf("t80 = %q, want %q — a branch name inside a group attributes nothing", got, LandingNotLanded)
	}
}

// AC-TLA-004 — body mentions never attribute, and neither add to nor subtract
// from a verdict the subject carries.
func TestLandedPredicate_BodyMentionNeverAttributes(t *testing.T) {
	dir := formRepo(t, "develop",
		formCommit{subject: "chore: seed the tree"},
		formCommit{subject: "chore: groundwork for the delivery", body: "discusses t555 at length"},
	)
	if got := landed(t, dir, "origin/develop", "t555"); got != LandingNotLanded {
		t.Errorf("t555 = %q, want %q — the subject names no card and the body mention attributes nothing", got, LandingNotLanded)
	}

	// The body's LAST line is shaped exactly like a form-2 attribution. A
	// query built on the whole message (%B) would feed that line to the
	// matcher as if it were a subject and answer landed; a subject-stream
	// query (%s) cannot see it.
	dir = formRepo(t, "develop",
		formCommit{subject: "chore: seed the tree"},
		formCommit{subject: "chore: groundwork for the delivery", body: "notes\nfixed the thing (t777)"},
	)
	if got := landed(t, dir, "origin/develop", "t777"); got != LandingNotLanded {
		t.Errorf("t777 = %q, want %q — a body line shaped like an attribution is not a subject", got, LandingNotLanded)
	}

	// The same fixture family with the scope present on a second commit: the
	// body occurrence neither adds to nor subtracts from the verdict.
	dir = formRepo(t, "develop",
		formCommit{subject: "chore: seed the tree"},
		formCommit{subject: "chore: groundwork for the delivery", body: "discusses t555 at length"},
		formCommit{subject: "docs(t555): the real delivery"},
	)
	if got := landed(t, dir, "origin/develop", "t555"); got != LandingLanded {
		t.Errorf("t555 = %q, want %q — the scope commit lands regardless of the body mention", got, LandingLanded)
	}
}

// §D edge case — a token that is a SUBSTRING of a longer token must not
// attribute: t44 inside t443. Word-boundary matching is retained.
func TestLandedPredicate_SubstringIsNotAttribution(t *testing.T) {
	dir := formRepo(t, "develop",
		formCommit{subject: "chore: seed the tree"},
		formCommit{subject: "docs(t443): context for the boundary case"},
	)
	if got := landed(t, dir, "origin/develop", "t44"); got != LandingNotLanded {
		t.Errorf("t44 = %q, want %q — a substring of t443 is not a match", got, LandingNotLanded)
	}
	if got := landed(t, dir, "origin/develop", "t443"); got != LandingLanded {
		t.Errorf("t443 = %q, want %q — the full token attributes", got, LandingLanded)
	}
}

// AC-TLA-007 — the unanswerable path stays three-valued; the two facts stay
// distinguishable.
func TestLandedPredicate_EmptyHistoryIsNotLandedNotUnknown(t *testing.T) {
	dir := formRepo(t, "develop", formCommit{subject: "chore: seed the tree"})
	answer, err := (GitLandedQuerier{Run: gitIn(dir), Ref: "origin/develop"}).Landed("t999")
	if err != nil {
		t.Fatalf("Landed on an empty history: %v", err)
	}
	if answer != LandingNotLanded {
		t.Errorf("empty history = %q, want %q — the query ran and found nothing", answer, LandingNotLanded)
	}
}
