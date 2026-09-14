package receipt

import (
	"errors"
	"sync"
)

var (
	// ErrCompactForeign rejects a PostCompact issued for another scope.
	ErrCompactForeign = errors.New("compaction notice belongs to another scope")
	// ErrCompactDigest rejects a returned summary that is not byte-exact against
	// the authenticated digest; a substring or padded match never passes.
	ErrCompactDigest = errors.New("compaction summary does not match the authenticated digest")
	// ErrCompactDuplicate rejects rebasing an epoch that already rebased.
	ErrCompactDuplicate = errors.New("compaction epoch already rebased")
	// ErrCompactStale rejects a notice older than the applied rebase.
	ErrCompactStale = errors.New("compaction notice predates the applied rebase")
	// ErrCompactAmbiguous rejects an epoch jump: the missing notice is a lost
	// PostCompact and the resulting state is ambiguous, never guessed.
	ErrCompactAmbiguous = errors.New("compaction notice leaves an epoch gap")
)

// CompactBase is the authenticated PostCompact notification: the scope it was
// issued for, the monotonically increasing compaction epoch, and the exact
// digest of the summary the App Server turn returned. It is never derived
// from model text or unauthenticated hooks.
type CompactBase struct {
	Scope   string
	Epoch   uint64
	Summary Digest
}

// RebaseLedger owns the compaction digest ledger of one authenticated scope.
// Public history rebases at most once per epoch, and only when the returned
// summary turn hashes to the exact authenticated digest.
type RebaseLedger struct {
	mu      sync.Mutex
	scope   string
	applied uint64
}

// NewRebaseLedger starts or restores the ledger for one scope. appliedEpoch is
// the last epoch whose rebase completed before a process restart, so a replay
// of a pre-restart notice cannot rebase twice.
func NewRebaseLedger(scope string, appliedEpoch uint64) *RebaseLedger {
	return &RebaseLedger{scope: scope, applied: appliedEpoch}
}

// Applied reports the last epoch whose rebase completed.
func (l *RebaseLedger) Applied() uint64 {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.applied
}

// Rebase contrasts the returned summary turn against the authenticated
// PostCompact and, on an exact digest and epoch match, records the rebase.
// Duplicate, stale, foreign, gapped and non-exact inputs yield explicit
// errors — never a second rebase or a guessed state.
func (l *RebaseLedger) Rebase(base CompactBase, summary []byte) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.scope == "" || base.Scope != l.scope {
		return ErrCompactForeign
	}
	if base.Epoch == 0 {
		return ErrInvalid
	}
	if Hash(summary) != base.Summary {
		return ErrCompactDigest
	}
	switch {
	case base.Epoch == l.applied:
		return ErrCompactDuplicate
	case base.Epoch < l.applied:
		return ErrCompactStale
	case base.Epoch > l.applied+1:
		return ErrCompactAmbiguous
	}
	l.applied = base.Epoch
	return nil
}

// Rebase resets the public history after a verified compaction: the session
// keeps its identity and generation, but no completed prefix validates
// history observations until the compacted base republishes. appliedEpoch is
// the epoch whose rebase this reset completes — recorded in the same manifest
// transaction so a restart restores the exact value (AC-MG-026 (b)).
func (m *Manifest) Rebase(appliedEpoch uint64) error {
	if m.session == (Digest{}) || m.generation == 0 || appliedEpoch == 0 {
		return ErrInvalid
	}
	m.candidates = []Candidate{}
	m.appliedEpoch = appliedEpoch
	if _, e := m.Marshal(); e != nil {
		return e
	}
	return nil
}

// AppliedEpoch reports the last compaction epoch whose rebase this manifest
// durably records. Zero means the scope never rebased.
func (m *Manifest) AppliedEpoch() uint64 { return m.appliedEpoch }
