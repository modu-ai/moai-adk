package cli

// factory_af_flow_hints_t2091_test.go — card t2091: the audit follow-ups of card
// t2071. The two landing hints must appear on exactly the refusals they were
// written for, and say what is true of the record.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/factory"
)

// A landing record that exists but names no delivering commit must not be
// described as "no landing is recorded" (F2).
func TestAFFlowRebindHintWhenTheLandingRecordNamesNoCommit(t *testing.T) {
	root, _, _, _, store := squashLandedFixture(t, "")
	emptyStageRow(t, root)
	if err := store.Mutate(func(r *factory.BacklogRecord) error {
		for i := range r.Items {
			if r.Items[i].ID == "t1" {
				r.Items[i].Landing = &factory.LandingEvidence{
					Ref: "origin/main", RefHead: "0000000000000000000000000000000000000000",
					ObservedAt: "2026-10-26T00:00:00Z",
				}
			}
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	_, err := rebindRun(t, "a landing record without a delivering commit")
	if err == nil || !strings.Contains(err.Error(), "cannot infer the audit phase") {
		t.Fatalf("err = %v, want the audit-phase refusal kept", err)
	}
	if !strings.Contains(err.Error(), "names no delivering commit") {
		t.Errorf("err = %q, want it to say the record names no delivering commit", err)
	}
	if strings.Contains(err.Error(), "no landing is recorded") {
		t.Errorf("err = %q, a landing record exists and must not be called absent", err)
	}
}

// The rebind hint belongs to the stage-empty refusal only: a card with a stage
// whose verdict is missing keeps the verdict refusal untouched (F3, mutant a).
func TestAFFlowRebindHintIsNotOnTheVerdictPath(t *testing.T) {
	root, _, _, _, _ := squashLandedFixture(t, "")
	card := fcCard(t, root, "t1")
	if err := os.Remove(filepath.Join(card.WorktreePath, ".moai", "reports", "t1", "sync-audit-iter1.md")); err != nil {
		t.Fatalf("remove the verdict file: %v", err)
	}
	_, err := rebindRun(t, "verdict missing, no landing record")
	if err == nil || !strings.Contains(err.Error(), "verdict to re-bind onto") {
		t.Fatalf("err = %v, want the verdict refusal", err)
	}
	for _, banned := range []string{"no landing is recorded", "moai todo landed"} {
		if strings.Contains(err.Error(), banned) {
			t.Errorf("err = %q, the landing hint must not ride the verdict path (found %q)", err, banned)
		}
	}
}

// A landing record observed before this run engaged the card belongs to an
// earlier issuance and is refused; recording the landing again replaces it with
// a fresh observation, which the chain then accepts. The procedure doc says so
// (card t2091), so the behavior is pinned here.
func TestAFFlowEarlierIssuanceRecordIsRefusedAndRecordingAgainHeals(t *testing.T) {
	root, _, _, _, store := squashLandedFixture(t, "")
	emptyStageRow(t, root)
	merge := afFlowLocalMainRoot(t, root)
	engageCard(t, root, "2020-01-01T00:00:00Z")

	if _, _, err := runTodo(t, "landed", "t1", "--sha", merge, "--ref", "main"); err != nil {
		t.Fatalf("todo landed: %v", err)
	}
	setLandingObservedAt(t, store, "2019-01-01T00:00:00Z")
	_, err := rebindRun(t, "record observed before this run engaged the card")
	if err == nil || !strings.Contains(err.Error(), "earlier issuance") {
		t.Fatalf("err = %v, want the earlier-issuance refusal", err)
	}

	if _, _, err := runTodo(t, "landed", "t1", "--sha", merge, "--ref", "main"); err != nil {
		t.Fatalf("recording the landing again: %v", err)
	}
	if out, err := rebindRun(t, "landing recorded again after the earlier-issuance refusal"); err != nil {
		t.Fatalf("rebind-evidence after recording again: %v\nout: %s", err, out)
	}
	if got := fcCard(t, root, "t1").EvidenceSHA; got != merge {
		t.Fatalf("evidence_sha = %q, want %s", got, merge)
	}
}

// The `--ref <branch>` suggestion is for a default origin/<branch> ref only, and
// only when the commit exists but is unreachable (F3, mutants b and c).
func TestAFFlowLandedHintOnlyForAnUnreachableCommitOnAnOriginRef(t *testing.T) {
	root, _, _, _, _ := squashLandedFixture(t, "")
	merge := afFlowLocalMainRoot(t, root)
	orphan := fcGit(t, root, "commit-tree", merge+"^{tree}", "-p", merge, "-m", "never landed anywhere")

	t.Run("a_local_ref_gets_no_suggestion", func(t *testing.T) {
		_, _, err := runTodo(t, "landed", "t1", "--sha", orphan, "--ref", "main")
		if err == nil || !strings.Contains(err.Error(), "is not reachable from main") {
			t.Fatalf("err = %v, want the reachability refusal against main", err)
		}
		if strings.Contains(err.Error(), "--ref") {
			t.Errorf("err = %q, a non-origin ref must not be told to pass --ref", err)
		}
	})
	t.Run("a_missing_commit_gets_no_suggestion", func(t *testing.T) {
		_, _, err := runTodo(t, "landed", "t1", "--sha", "0123456789abcdef0123456789abcdef01234567")
		if err == nil || !strings.Contains(err.Error(), "names no commit") {
			t.Fatalf("err = %v, want the existence refusal", err)
		}
		if strings.Contains(err.Error(), "--ref") {
			t.Errorf("err = %q, the existence refusal must not carry the --ref suggestion", err)
		}
	})
}
