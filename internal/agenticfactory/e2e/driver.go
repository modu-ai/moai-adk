// Package e2e holds the Issue #1948 end-to-end scenario harness.
//
// Scenarios assert only on observable outcomes through Driver. The ModelDriver
// is the reference behavior from SPEC-AGENTIC-FACTORY-001 (design.md §2-§3);
// the real claim/runner/reconcile packages bind to the same interface once
// they land, so each scenario runs unchanged against both.
package e2e

import (
	"context"
	"fmt"
	"sync"
)

// ClaimResult is what a claimant observes after a claim attempt.
type ClaimResult struct {
	Won       bool
	Owner     string
	AttemptID string
}

// Driver is the observable surface the scenarios exercise.
type Driver interface {
	Claim(ctx context.Context, issue, session string) (ClaimResult, error)
	LiveAttempts(issue string) int
}

type claimRecord struct {
	owner    string
	attempts int
	live     int
}

// ModelDriver is an in-memory reference implementation of the contract.
type ModelDriver struct {
	mu     sync.Mutex
	claims map[string]*claimRecord
}

// NewModelDriver returns an empty reference driver.
func NewModelDriver() *ModelDriver {
	return &ModelDriver{claims: map[string]*claimRecord{}}
}

// Claim grants the first caller ownership and a single attempt; later callers
// observe the owner and receive no attempt.
func (m *ModelDriver) Claim(_ context.Context, issue, session string) (ClaimResult, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if rec, ok := m.claims[issue]; ok {
		return ClaimResult{Owner: rec.owner}, nil
	}
	rec := &claimRecord{owner: session, attempts: 1, live: 1}
	m.claims[issue] = rec
	return ClaimResult{Won: true, Owner: session, AttemptID: fmt.Sprintf("%s/a%d", issue, rec.attempts)}, nil
}

// LiveAttempts reports how many attempts are live for the Issue.
func (m *ModelDriver) LiveAttempts(issue string) int {
	m.mu.Lock()
	defer m.mu.Unlock()
	if rec, ok := m.claims[issue]; ok {
		return rec.live
	}
	return 0
}

var _ Driver = (*ModelDriver)(nil)
