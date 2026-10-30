package e2e

import (
	"context"
	"sync"
	"testing"
)

// Issue #1948 — scenario 1: two sessions race to claim the same Issue.
// Exactly one claim must win; the loser must observe the winner's owner and
// must not receive an attempt.
func TestScenario_CompetingClaim_ExactlyOneWinner(t *testing.T) {
	d := NewModelDriver()
	ctx := context.Background()

	const issue = "repo#1948"
	sessions := []string{"s-a", "s-b", "s-c", "s-d"}

	results := make([]ClaimResult, len(sessions))
	var wg sync.WaitGroup
	for i, s := range sessions {
		wg.Go(func() {
			r, err := d.Claim(ctx, issue, s)
			if err != nil {
				t.Errorf("claim by %s: %v", s, err)
				return
			}
			results[i] = r
		})
	}
	wg.Wait()

	winners := 0
	var winner string
	for i, r := range results {
		if r.Won {
			winners++
			winner = sessions[i]
			if r.AttemptID == "" {
				t.Errorf("winner %s has no attempt id", sessions[i])
			}
		}
	}
	if winners != 1 {
		t.Fatalf("want exactly 1 winner, got %d", winners)
	}
	for i, r := range results {
		if r.Won {
			continue
		}
		if r.AttemptID != "" {
			t.Errorf("loser %s received attempt %q", sessions[i], r.AttemptID)
		}
		if r.Owner != winner {
			t.Errorf("loser %s saw owner %q, want %q", sessions[i], r.Owner, winner)
		}
	}
	if n := d.LiveAttempts(issue); n != 1 {
		t.Errorf("live attempts for %s = %d, want 1", issue, n)
	}
}
