// Package effect records the four delivery states of a command (SPEC-AGENTIC-FACTORY-001
// REQ-AF-003, design §3): sent, received, executed, completed. Each state is set only from
// its own evidence source; a send result or an idle notice never advances past sent.
package effect

import (
	"errors"
	"fmt"
	"regexp"
	"sync"
)

// Stage is one of the four delivery states, in order.
type Stage string

const (
	Sent      Stage = "sent"
	Received  Stage = "received"
	Executed  Stage = "executed"
	Completed Stage = "completed"
)

var order = map[Stage]int{Sent: 1, Received: 2, Executed: 3, Completed: 4}

// EvidenceKind names where an observation came from.
type EvidenceKind string

const (
	SendRecord   EvidenceKind = "send-record"   // sender's queue row / send response
	ReceiptRow   EvidenceKind = "receipt-row"   // recipient's acknowledgement row carrying the command id
	AttemptStart EvidenceKind = "attempt-start" // attempt start row plus worktree HEAD observation
	ResultFile   EvidenceKind = "result-file"   // result file on disk read by the control service

	// Non-evidence: they never advance a stage past sent (design §3, last column).
	SendSuccess  EvidenceKind = "send-success"
	IdleNotice   EvidenceKind = "idle-notice"
	ReplyMessage EvidenceKind = "reply-message"
)

var wantKind = map[Stage]EvidenceKind{
	Sent: SendRecord, Received: ReceiptRow, Executed: AttemptStart, Completed: ResultFile,
}

// Evidence is the observation backing a transition.
type Evidence struct {
	Kind    EvidenceKind
	Ref     string // path or row id
	HeadSHA string // required (full 40 hex) for Completed
}

// Record is the stored state of one command.
type Record struct {
	CommandID string
	Issue     int
	Stage     Stage
	Evidence  Evidence
}

var (
	ErrBackward = errors.New("effect: backward transition refused")
	ErrSkip     = errors.New("effect: stage skip refused")
	ErrEvidence = errors.New("effect: evidence does not support this stage")
)

var fullSHA = regexp.MustCompile(`^[0-9a-f]{40}$`)

// Store persists records. The production store is the local claim store (provided by lane 18);
// MemStore is the in-package stand-in.
type Store interface {
	Get(commandID string) (Record, bool)
	Put(Record) error
}

// Tracker applies the transition rules over a Store.
type Tracker struct{ store Store }

func NewTracker(s Store) *Tracker { return &Tracker{store: s} }

// Advance moves a command to stage `to`. Repeating the current stage with the same command id is
// an idempotent no-op; going back or skipping a stage is refused.
func (t *Tracker) Advance(commandID string, issue int, to Stage, ev Evidence) (Record, error) {
	if commandID == "" {
		return Record{}, errors.New("effect: empty command id")
	}
	toN, ok := order[to]
	if !ok {
		return Record{}, fmt.Errorf("effect: unknown stage %q", to)
	}
	cur, exists := t.store.Get(commandID)
	if exists && cur.Issue != issue {
		return Record{}, fmt.Errorf("effect: command %s belongs to issue %d, not %d", commandID, cur.Issue, issue)
	}
	curN := 0
	if exists {
		curN = order[cur.Stage]
	}
	switch {
	case toN == curN:
		return cur, nil
	case toN < curN:
		return Record{}, fmt.Errorf("%w: %s -> %s", ErrBackward, cur.Stage, to)
	case toN > curN+1:
		return Record{}, fmt.Errorf("%w: %s -> %s", ErrSkip, orNone(cur.Stage), to)
	}
	if ev.Kind != wantKind[to] {
		return Record{}, fmt.Errorf("%w: %s needs %s, got %q", ErrEvidence, to, wantKind[to], ev.Kind)
	}
	if ev.Ref == "" {
		return Record{}, fmt.Errorf("%w: empty evidence ref", ErrEvidence)
	}
	if to == Completed && !fullSHA.MatchString(ev.HeadSHA) {
		return Record{}, fmt.Errorf("%w: completed needs the full 40-hex HEAD SHA", ErrEvidence)
	}
	rec := Record{CommandID: commandID, Issue: issue, Stage: to, Evidence: ev}
	if err := t.store.Put(rec); err != nil {
		return Record{}, err
	}
	return rec, nil
}

func orNone(s Stage) string {
	if s == "" {
		return "(none)"
	}
	return string(s)
}

// MemStore is an in-memory Store with a write counter.
type MemStore struct {
	mu  sync.Mutex
	m   map[string]Record
	seq int
}

func NewMemStore() *MemStore { return &MemStore{m: map[string]Record{}} }

func (s *MemStore) Get(id string) (Record, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	r, ok := s.m[id]
	return r, ok
}

func (s *MemStore) Put(r Record) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.m[r.CommandID] = r
	s.seq++
	return nil
}

// Seq is the number of writes so far.
func (s *MemStore) Seq() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.seq
}

// Rank orders the stages (0 for an unknown or empty stage).
func (s Stage) Rank() int { return order[s] }
