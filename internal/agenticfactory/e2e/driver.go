// Package e2e holds the Issue #1948 end-to-end scenario harness.
//
// Scenarios assert only on observable outcomes. ModelDriver is the reference
// behavior from SPEC-AGENTIC-FACTORY-001 (design.md §2-§3, provisional); the
// real claim/runner/reconcile packages bind to the same surface once they
// land, so each scenario runs unchanged against both.
package e2e

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

// ClaimResult is what a claimant observes after a claim attempt.
type ClaimResult struct {
	Won       bool
	Owner     string
	AttemptID string
}

// Delivery is the four-stage delivery state (design.md §3).
type Delivery string

const (
	DeliverySent      Delivery = "sent"
	DeliveryReceived  Delivery = "received"
	DeliveryExecuted  Delivery = "executed"
	DeliveryCompleted Delivery = "completed"
)

// Probe asks the receiving side what state a command reached.
type Probe func(cmdID string) (Delivery, error)

// AckOutcome is the result of handling a missing ACK.
type AckOutcome struct {
	Resent      bool
	State       Delivery
	BlockReason string
}

// Liveness is the owner liveness classification (dead / live / indeterminate).
type Liveness string

const (
	OwnerLive          Liveness = "live"
	OwnerDead          Liveness = "dead"
	OwnerIndeterminate Liveness = "indeterminate"
)

// ResumeResult reports whether a resume produced a new attempt.
type ResumeResult struct {
	NewAttempt bool
	AttemptID  string
	Epoch      int
}

// Remote is an observation of the remote merge state.
type Remote string

const (
	RemoteUnknown      Remote = "unknown"
	RemoteLookupFailed Remote = "lookup-failed"
)

// Completion is the evidence required to reach done (REQ-AF-002).
type Completion struct {
	RemoteMain  bool
	LocalSynced bool
	Deployed    bool
	DeployNA    string // non-empty reason stands in for a deploy record
}

type record struct {
	owner        string
	epoch        int
	attempts     []string
	attemptState map[string]string
	state        string
	reason       string
	auditRounds  int
	closed       bool
}

// ModelDriver is an in-memory reference implementation of the contract.
type ModelDriver struct {
	mu     sync.Mutex
	issues map[string]*record
	sends  map[string]int
	deliv  map[string]Delivery

	failClose bool
}

// NewModelDriver returns an empty reference driver.
func NewModelDriver() *ModelDriver {
	return &ModelDriver{
		issues: map[string]*record{},
		sends:  map[string]int{},
		deliv:  map[string]Delivery{},
	}
}

func attemptID(issue string, n int) string { return fmt.Sprintf("%s/a%d", issue, n) }

func key(issue, cmd string) string { return issue + "|" + cmd }

// Claim grants the first caller ownership and a single attempt; later callers
// observe the owner and receive no attempt.
func (m *ModelDriver) Claim(_ context.Context, issue, session string) (ClaimResult, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if rec, ok := m.issues[issue]; ok {
		return ClaimResult{Owner: rec.owner}, nil
	}
	id := attemptID(issue, 1)
	m.issues[issue] = &record{
		owner: session, epoch: 1, attempts: []string{id},
		attemptState: map[string]string{id: "live"}, state: "claimed",
	}
	return ClaimResult{Won: true, Owner: session, AttemptID: id}, nil
}

// NextDispatch claims the next Issue for a worker.
func (m *ModelDriver) NextDispatch(ctx context.Context, issue, session string) (ClaimResult, error) {
	return m.Claim(ctx, issue, session)
}

// LiveAttempts reports how many attempts are live for the Issue.
func (m *ModelDriver) LiveAttempts(issue string) int {
	m.mu.Lock()
	defer m.mu.Unlock()
	rec, ok := m.issues[issue]
	if !ok {
		return 0
	}
	n := 0
	for _, s := range rec.attemptState {
		if s == "live" {
			n++
		}
	}
	return n
}

// AttemptState returns the state of one attempt.
func (m *ModelDriver) AttemptState(issue, attempt string) string {
	m.mu.Lock()
	defer m.mu.Unlock()
	if rec, ok := m.issues[issue]; ok {
		return rec.attemptState[attempt]
	}
	return ""
}

// Dispatch records a send once per command id; repeats are no-ops.
func (m *ModelDriver) Dispatch(_ context.Context, issue, cmd string) (Delivery, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.issues[issue]; !ok {
		return "", errors.New("no claim")
	}
	k := key(issue, cmd)
	if _, dup := m.deliv[k]; !dup {
		m.sends[k]++
		m.deliv[k] = DeliverySent
	}
	return m.deliv[k], nil
}

// SendCount reports how many times the command was actually sent.
func (m *ModelDriver) SendCount(issue, cmd string) int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.sends[key(issue, cmd)]
}

// OnAckTimeout queries the receiver before any resend. At received or beyond
// nothing is resent; an unqueryable receiver blocks.
func (m *ModelDriver) OnAckTimeout(issue, cmd string, probe Probe) AckOutcome {
	m.mu.Lock()
	defer m.mu.Unlock()
	k := key(issue, cmd)
	got, err := probe(cmd)
	if err != nil {
		m.setBlocked(issue, "connection-failed")
		return AckOutcome{State: m.deliv[k], BlockReason: "connection-failed"}
	}
	m.deliv[k] = got
	if got == DeliverySent {
		m.sends[k]++
		return AckOutcome{Resent: true, State: got}
	}
	return AckOutcome{State: got}
}

// Resume replaces a dead owner's attempt under the same claim; a live or
// indeterminate owner is left alone.
func (m *ModelDriver) Resume(_ context.Context, issue string, l Liveness) (ResumeResult, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	rec, ok := m.issues[issue]
	if !ok {
		return ResumeResult{}, errors.New("no claim")
	}
	if l != OwnerDead {
		return ResumeResult{}, nil
	}
	old := rec.attempts[len(rec.attempts)-1]
	rec.attemptState[old] = "superseded"
	rec.epoch++
	id := attemptID(issue, len(rec.attempts)+1)
	rec.attempts = append(rec.attempts, id)
	rec.attemptState[id] = "live"
	return ResumeResult{NewAttempt: true, AttemptID: id, Epoch: rec.epoch}, nil
}

// RecordAuditRound counts a round; reaching the cap without PASS blocks.
func (m *ModelDriver) RecordAuditRound(issue string, limit int, verdict string) string {
	m.mu.Lock()
	defer m.mu.Unlock()
	rec := m.issues[issue]
	rec.auditRounds++
	if verdict != "PASS" && rec.auditRounds >= limit {
		m.setBlocked(issue, "budget-exhausted")
	}
	return rec.state
}

// MarkReady refuses while blocked.
func (m *ModelDriver) MarkReady(issue string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	rec := m.issues[issue]
	if rec.state == "blocked" {
		return fmt.Errorf("blocked: %s", rec.reason)
	}
	rec.state = "ready"
	return nil
}

// ForceState sets a state directly (scenario setup only).
func (m *ModelDriver) ForceState(issue, state string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.issues[issue].state = state
}

// State returns the state and block reason.
func (m *ModelDriver) State(issue string) (string, string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	rec := m.issues[issue]
	return rec.state, rec.reason
}

// ObserveRemote handles a remote merge observation. Unknown re-queries from
// landing; a failed lookup blocks.
func (m *ModelDriver) ObserveRemote(issue string, r Remote) string {
	m.mu.Lock()
	defer m.mu.Unlock()
	if r == RemoteLookupFailed {
		m.setBlocked(issue, "connection-failed")
	}
	return m.issues[issue].state
}

// FailNextClose makes the next CloseIssue call fail (scenario setup only).
func (m *ModelDriver) FailNextClose() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.failClose = true
}

// CloseIssue closes the Issue. Per S46 (E13/E14, M34) the close comes BEFORE
// done; a failed close leaves the Issue not closed and not done.
func (m *ModelDriver) CloseIssue(issue string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.failClose {
		m.failClose = false
		return errors.New("close failed")
	}
	m.issues[issue].closed = true
	return nil
}

// Complete moves landing to done only with all three records AND a closed
// Issue; a close that failed keeps done at zero.
func (m *ModelDriver) Complete(issue string, c Completion) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	rec := m.issues[issue]
	if !c.RemoteMain || !c.LocalSynced || (!c.Deployed && c.DeployNA == "") {
		return errors.New("done needs remote main, local sync and deploy record or N/A reason")
	}
	if !rec.closed {
		return errors.New("done needs the Issue closed first")
	}
	rec.state = "done"
	return nil
}

// IssueClosed reports whether the Issue was closed.
func (m *ModelDriver) IssueClosed(issue string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.issues[issue].closed
}

func (m *ModelDriver) setBlocked(issue, reason string) {
	if rec, ok := m.issues[issue]; ok {
		rec.state, rec.reason = "blocked", reason
	}
}
