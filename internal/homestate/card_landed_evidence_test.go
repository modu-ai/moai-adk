package homestate

// card_landed_evidence_test.go — card t1962: the squash-landing evidence
// helpers behind `factory rebind-evidence`. ResolveLandedEvidence admits a
// commit that resolves and is reachable from a ref; CardEngagementStart reports
// when a run first recorded an event for a card, so a landing record from an
// earlier issuance can be told apart.

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestResolveLandedEvidence(t *testing.T) {
	repo := frNewRepo(t, false)
	ctx := context.Background()

	t.Run("a_commit_reachable_from_the_ref_resolves_to_its_full_name", func(t *testing.T) {
		got, err := ResolveLandedEvidence(ctx, repo.Dir, repo.Commit[:12], frIntegration)
		if err != nil {
			t.Fatalf("ResolveLandedEvidence: %v", err)
		}
		if got != repo.Commit {
			t.Fatalf("got %q, want the full name %q of the abbreviated input", got, repo.Commit)
		}
	})

	t.Run("a_commit_not_reachable_from_the_ref_is_refused", func(t *testing.T) {
		// main stops at C, so the merge commit M on integration is not on it.
		_, err := ResolveLandedEvidence(ctx, repo.Dir, repo.Merge, "main")
		if err == nil || !strings.Contains(err.Error(), "not reachable from main") {
			t.Fatalf("err = %v, want a not-reachable refusal", err)
		}
	})

	t.Run("an_unusable_ref_is_refused", func(t *testing.T) {
		for _, ref := range []string{"", "  ", "-x"} {
			if _, err := ResolveLandedEvidence(ctx, repo.Dir, repo.Commit, ref); err == nil || !strings.Contains(err.Error(), "not usable") {
				t.Errorf("ref %q: err = %v, want the unusable-ref refusal", ref, err)
			}
		}
	})

	t.Run("something_that_is_not_a_commit_is_refused", func(t *testing.T) {
		for _, sha := range []string{"zzzz", "0123456789abcdef0123456789abcdef01234567"} {
			if _, err := ResolveLandedEvidence(ctx, repo.Dir, sha, frIntegration); err == nil {
				t.Errorf("sha %q: want a refusal, got none", sha)
			}
		}
	})
}

func TestCardEngagementStart(t *testing.T) {
	db := frOpen(t)
	ctx := context.Background()
	insert := func(runID, payload, at string) {
		t.Helper()
		if _, err := db.DB.Exec(`INSERT INTO events(run_id,kind,payload_json,created_at) VALUES(?,?,?,?)`,
			runID, "card.assigned", payload, at); err != nil {
			t.Fatal(err)
		}
	}

	t.Run("no_event_for_the_card_is_unknown_not_early", func(t *testing.T) {
		insert("run-x", `{"card_id":"other"}`, "2026-10-25T00:00:00Z")
		_, found, err := db.CardEngagementStart(ctx, "run-x", "t9")
		if err != nil || found {
			t.Fatalf("found=%v err=%v, want not found and no error", found, err)
		}
	})

	t.Run("the_earliest_event_by_time_wins_not_by_text", func(t *testing.T) {
		// "…05.5Z" sorts before "…05Z" as text ('.' < 'Z') but is the LATER instant.
		insert("run-y", `{"card_id":"t1"}`, "2026-10-26T09:00:05.5Z")
		insert("run-y", `{"card_id":"t1"}`, "2026-10-26T09:00:05Z")
		insert("run-y", `{"card_id":"t1"}`, "not a time")
		insert("run-y", `{"card_id":"t2"}`, "2026-10-01T00:00:00Z")
		insert("run-z", `{"card_id":"t1"}`, "2026-10-02T00:00:00Z")
		got, found, err := db.CardEngagementStart(ctx, "run-y", "t1")
		if err != nil || !found {
			t.Fatalf("found=%v err=%v, want the engagement start", found, err)
		}
		if want := time.Date(2026, 10, 26, 9, 0, 5, 0, time.UTC); !got.Equal(want) {
			t.Fatalf("start = %s, want %s (the other run's and the other card's events must not count)", got, want)
		}
	})
}
