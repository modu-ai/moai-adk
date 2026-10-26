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

	t.Run("the_event_the_factory_itself_writes_when_it_picks_a_card_is_found", func(t *testing.T) {
		// Hand-inserted payloads would pass even if the factory's own events
		// stopped carrying card_id, which would turn the engagement check into
		// a silent no-op; this one is written by the pick path itself.
		rebindPick(t, db, "run-real", "t7", "")
		got, found, err := db.CardEngagementStart(ctx, "run-real", "t7")
		if err != nil || !found {
			t.Fatalf("found=%v err=%v, want the pick event to count as the engagement start", found, err)
		}
		if !got.Equal(rebindNow) {
			t.Fatalf("start = %s, want the pick instant %s", got, rebindNow)
		}
	})

	t.Run("the_earliest_event_by_time_wins_not_by_text", func(t *testing.T) {
		// "…05.5Z" sorts before "…05Z" as text ('.' < 'Z') but is the LATER instant.
		insert("run-y", `{"card_id":"t1"}`, "2026-10-26T09:00:05.5Z")
		insert("run-y", `{"card_id":"t1"}`, "2026-10-26T09:00:05Z")
		insert("run-y", `{"card_id":"t1"}`, "not a time")
		insert("run-y", `{"card_id":"t2"}`, "2026-10-01T00:00:00Z")
		got, found, err := db.CardEngagementStart(ctx, "run-y", "t1")
		if err != nil || !found {
			t.Fatalf("found=%v err=%v, want the engagement start", found, err)
		}
		if want := time.Date(2026, 10, 26, 9, 0, 5, 0, time.UTC); !got.Equal(want) {
			t.Fatalf("start = %s, want %s (the other card's events must not count)", got, want)
		}
	})

	t.Run("an_unrelated_runs_events_do_not_count", func(t *testing.T) {
		// run-v holds an earlier event for the same card id, but no consolidation
		// record ties run-v to run-u: it is another issuance of the card, not
		// history that was moved into run-u (contrast the moved-card subtest below).
		insert("run-v", `{"card_id":"t3"}`, "2026-10-02T00:00:00Z")
		insert("run-u", `{"card_id":"t3"}`, "2026-10-05T00:00:00Z")
		got, found, err := db.CardEngagementStart(ctx, "run-u", "t3")
		if err != nil || !found {
			t.Fatalf("found=%v err=%v, want the engagement start", found, err)
		}
		if want := time.Date(2026, 10, 5, 0, 0, 0, 0, time.UTC); !got.Equal(want) {
			t.Fatalf("start = %s, want %s (an unrelated run's events must not count)", got, want)
		}
	})

	// REQ-FCR-025 (card t1963, audit finding M1): a consolidation re-keys a card
	// into the target run but leaves its earlier events under the SOURCE run, so a
	// card moved in by consolidation has its history read from the run(s) it was
	// moved from. Before this, only the target run's own events counted and the
	// engagement start of a moved card came out too late.
	t.Run("a_card_moved_in_by_consolidation_keeps_its_source_run_history", func(t *testing.T) {
		mdb := openSandboxFactory(t)
		rbRun(t, mdb, "run-old", rbPIDDead)
		rbRun(t, mdb, "run-new", rbPIDLive)
		rbCard(t, mdb, "run-old", "c1", CardRun)
		lceEvent(t, mdb, "run-old", "card.transition", map[string]any{"card_id": "c1", "from": "", "to": CardPicked}, "2026-01-01T00:00:00Z")
		lceRebind(t, mdb, rbOpts(), "run-old", "run-new")
		lceEvent(t, mdb, "run-new", "card.transition", map[string]any{"card_id": "c1", "from": CardPicked, "to": CardRun}, "2026-06-01T00:00:00Z")

		got, found, err := mdb.CardEngagementStart(ctx, "run-new", "c1")
		if err != nil || !found {
			t.Fatalf("found=%v err=%v, want the engagement start", found, err)
		}
		if want := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC); !got.Equal(want) {
			t.Fatalf("start = %s, want %s (the history the card carries from its source run, REQ-FCR-025)", got, want)
		}
	})

	t.Run("an_abandoned_or_left_behind_card_contributes_nothing_from_the_source_run", func(t *testing.T) {
		// lceSeedConsolidated: c1 moved, c4 abandoned (run-new already held it),
		// c5 terminal and left behind. Source events: c4 at +6m, target c4 at +7m.
		mdb, _, _, _, _, _ := lceSeedConsolidated(t)
		got, found, err := mdb.CardEngagementStart(ctx, "run-new", "c4")
		if err != nil || !found {
			t.Fatalf("c4: found=%v err=%v, want the target's own engagement start", found, err)
		}
		if want := lceBase.Add(7 * time.Minute); !got.Equal(want) {
			t.Fatalf("c4 start = %s, want %s (the abandoned source copy at +6m must not count)", got, want)
		}
		if _, found, err := mdb.CardEngagementStart(ctx, "run-new", "c5"); err != nil || found {
			t.Fatalf("c5: found=%v err=%v, want not found (the left-behind card has no history on run-new)", found, err)
		}
	})

	t.Run("a_chain_of_consolidations_returns_the_earliest_over_the_whole_chain", func(t *testing.T) {
		// lceSeedChain: c1 moves run-a -> run-b -> run-c (its first event is in
		// run-a at +0m); c9 moves run-b -> run-c (+31m); c2 is abandoned on the
		// second hop (run-c's own c2 event at +32m).
		mdb, _, _, _, _, _, _ := lceSeedChain(t)
		for _, tc := range []struct {
			run, card string
			want      time.Time
		}{
			{"run-c", "c1", lceBase},
			{"run-b", "c1", lceBase},
			{"run-c", "c9", lceBase.Add(31 * time.Minute)},
			{"run-c", "c2", lceBase.Add(32 * time.Minute)},
		} {
			got, found, err := mdb.CardEngagementStart(ctx, tc.run, tc.card)
			if err != nil || !found {
				t.Fatalf("%s/%s: found=%v err=%v", tc.run, tc.card, found, err)
			}
			if !got.Equal(tc.want) {
				t.Errorf("%s/%s start = %s, want %s", tc.run, tc.card, got, tc.want)
			}
		}
	})
}
