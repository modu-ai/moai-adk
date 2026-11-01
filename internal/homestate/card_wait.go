// card_wait.go — card t2013: a lane's wait recorded mechanically in the
// factory record. The wait names who must wake, and what condition releases
// it, so the chain end -> result pickup -> next action no longer depends on
// leader memory. The table is ADDITIVE (CREATE TABLE IF NOT EXISTS in
// factoryDDL, the precedent card_metadata_events set): no existing row is
// rewritten and the schema version is unchanged.
package homestate

import (
	"context"
	"time"
)

// The wait condition kinds and the two non-lane wakers.
const (
	WaitCondPredecessorLanded = "predecessor-landed"
	WaitCondOperatorRuling    = "operator-ruling"
	WaitWakerLeader           = "leader"
	WaitWakerOperator         = "operator"
)

// CardWait is one recorded wait: Waker must be woken when the condition
// (CondKind, CondRef) is met. WokenAt is empty until the wake was handed out.
type CardWait struct {
	Seq        int64
	CardID     string
	CondKind   string
	CondRef    string
	Waker      string
	Reason     string
	RecordedBy string
	RecordedAt string
	WokenAt    string
}

// RecordCardWait is a stub in the RED commit.
func (f *FactoryDB) RecordCardWait(ctx context.Context, w CardWait, now time.Time) error {
	return nil
}

// LoadCardWaits is a stub in the RED commit.
func (f *FactoryDB) LoadCardWaits(ctx context.Context, cardID string) ([]CardWait, error) {
	return nil, nil
}

// SettleLandedWaits is a stub in the RED commit.
func (f *FactoryDB) SettleLandedWaits(ctx context.Context, landedCardID string, now time.Time) ([]CardWait, error) {
	return nil, nil
}
