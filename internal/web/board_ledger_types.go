// board_ledger_types.go — provisional stand-ins for the local ledger's read API
// (Issue #1948, local-only redesign; contract .moai/reports/issue-1948/local-adk-design-20261030.md §2-§3).
//
// ledger/api.go is lane 18's and does not exist yet. These types carry the
// contract's §2 table columns under the names the board reads, and BoardReader
// has the §3 Reader method set. When api.go lands they become aliases
// (type LedgerCard = ledger.Card, ...) and nothing else here changes.
//
// Contract gap: §3's Reader has no lease_keys method, so LedgerLease carries its
// conflict keys in Keys; the real Lease type must do the same or Reader needs a
// Keys method.
package web

import (
	"context"
	"time"
)

type LedgerCard struct {
	ID, Title, Stage, SubStage string
	Parked                     bool
	ParkReason                 string
	Priority, ReclaimCount     int
}

type LedgerLease struct {
	CardID, SessionID, Worktree        string
	Epoch                              int64
	PID                                int
	HeartbeatAt, NextWakeAt, ExpiresAt time.Time
	Keys                               []string
}

type LedgerQueueItem struct {
	CardID, CandidateSHA, State, FailReason string
	EnqueuedAt                              time.Time
}

type LedgerNote struct {
	Seq                int64
	Kind, CardID, Text string
	At                 time.Time
}

// BoardReader is the read-only ledger source the board renders.
type BoardReader interface {
	Cards(ctx context.Context) ([]LedgerCard, error)
	Leases(ctx context.Context) ([]LedgerLease, error)
	Queue(ctx context.Context) ([]LedgerQueueItem, error)
	Notes(ctx context.Context, sinceSeq int64) ([]LedgerNote, error)
}
