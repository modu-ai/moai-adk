// leader_approval.go — the leader approval receipt (SPEC-FACTORY-COMPLETION-RECOVERY-001
// M1, REQ-FCR-001/002/005): the record that the factory LEADER read a card's
// verification evidence and approved its completion, bound to the four values
// that make the approval non-transferable — the card UUID, the factory run
// the approval names, the card's factory version at approval time, and the
// evidence hash that was reviewed.
//
// The gate only VERIFIES. Issuance is a leader-path act (IssueLeaderApproval
// is reachable from no lane-callable CLI surface — REQ-FCR-014), and a
// receipt whose issuer marker names the performing owner is refused wherever
// it is presented (REQ-FCR-005 — the independent axis is performer ≠
// approver: the same leader issuing the receipt and executing the done is
// the normal single-leader flow, never a refusal reason).
//
// The type lives in homestate because both consumers already import it: the
// backlog done surfaces (internal/cli) and the done transitions in this
// package. factory → homestate is the existing dependency direction, so a
// factory placement would close a cycle (TestHomestateDoesNotImportFactory).
//
// The name deliberately avoids "receipt": internal/cli/todo.go's store
// issuance comments already use that word for an unrelated concept.
package homestate

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

// Issuer role markers carried on every receipt. ApprovalIssuerLeader marks
// the leader path; ApprovalIssuerLane marks a performing owner's self-issued
// receipt, which no gate accepts and no issuance path mints.
const (
	ApprovalIssuerLeader = "leader"
	ApprovalIssuerLane   = "lane"
)

// Approval refusals. A refusal leaves the card row, the event log, and the
// receipt store exactly as they were; callers match with errors.Is.
var (
	ErrApprovalMissing      = errors.New("leader approval receipt absent")
	ErrApprovalIssuer       = errors.New("leader approval issuer refused")
	ErrApprovalCardMismatch = errors.New("leader approval card binding refused")
	ErrApprovalRunMismatch  = errors.New("leader approval run binding refused")
	ErrApprovalStale        = errors.New("leader approval stale")
	ErrApprovalHashMismatch = errors.New("leader approval evidence hash refused")
)

// LeaderApproval is one leader evidence-review receipt.
type LeaderApproval struct {
	// CardUUID is the backlog card's identity — the value that survives a
	// card id's reuse across runs, so a receipt can never be silently
	// re-pointed at a reissued card.
	CardUUID string `json:"card_uuid"`
	// RunID is the factory run the approval names. A fresh run starts the
	// same card's version at 1 in a new row, so (uuid, version, hash) alone
	// reproduces across runs — the run binding is what stops the previous
	// run's approval from closing the new run's work (REQ-FCR-001).
	RunID string `json:"run_id"`
	// CardID is the queue id the receipt was issued for (diagnostic and the
	// transition-side lookup key; the binding authority is CardUUID).
	CardID string `json:"card_id"`
	// FactoryVersion is the card's factory version at approval time.
	FactoryVersion int64 `json:"factory_version"`
	// EvidenceHash is the evidence hash the leader reviewed — the card row's
	// EvidenceSHA as it stood at issuance.
	EvidenceHash string `json:"evidence_hash"`
	// Issuer names who issued the receipt; IssuerRole carries the
	// leader/lane marker the gates refuse on.
	Issuer     string `json:"issuer"`
	IssuerRole string `json:"issuer_role"`
	IssuedAt   string `json:"issued_at"`
}

// VerifyBinding checks the receipt against the named binding values — the
// verifier's own identity knowledge, not a caller-supplied verdict. cardUUID
// is the backlog identity; a verifier that cannot know it (the factory row
// carries no uuid) passes "" and the uuid axis is skipped rather than failed
// — the CLI surfaces, which do know it, always pass it. ownerLabel is the
// factory row's performing owner: a receipt IssuerED by that owner is
// refused wherever it is presented (REQ-FCR-005 — performer ≠ approver; the
// role marker is data, and this comparison is the gate's own), skipped only
// when the row names no owner.
func (a LeaderApproval) VerifyBinding(cardUUID, runID string, version int64, evidenceHash, ownerLabel string) error {
	if a.IssuerRole != ApprovalIssuerLeader {
		return fmt.Errorf("%w: issuer %q holds role %q, not %q", ErrApprovalIssuer, a.Issuer, a.IssuerRole, ApprovalIssuerLeader)
	}
	if ownerLabel != "" && a.Issuer == ownerLabel {
		return fmt.Errorf("%w: issuer %q is the card's performing owner — a performer does not approve its own work", ErrApprovalIssuer, a.Issuer)
	}
	if cardUUID != "" && a.CardUUID != cardUUID {
		return fmt.Errorf("%w: receipt names card %s, the closing card is %s", ErrApprovalCardMismatch, a.CardUUID, cardUUID)
	}
	if a.RunID != runID {
		return fmt.Errorf("%w: receipt binds run %q, the card's run is %q", ErrApprovalRunMismatch, a.RunID, runID)
	}
	if a.FactoryVersion != version {
		return fmt.Errorf("%w: receipt binds version %d, the card is at version %d", ErrApprovalStale, a.FactoryVersion, version)
	}
	if a.EvidenceHash != evidenceHash {
		return fmt.Errorf("%w: receipt binds evidence %q, the card carries %q", ErrApprovalHashMismatch, a.EvidenceHash, evidenceHash)
	}
	return nil
}

const leaderApprovalsDDL = `
CREATE TABLE IF NOT EXISTS leader_approvals (
  card_uuid TEXT NOT NULL,
  run_id TEXT NOT NULL,
  card_id TEXT NOT NULL,
  factory_version INTEGER NOT NULL,
  evidence_hash TEXT NOT NULL,
  issuer TEXT NOT NULL,
  issuer_role TEXT NOT NULL,
  issued_at TEXT NOT NULL,
  PRIMARY KEY(card_uuid, run_id)
);
`

// IssueLeaderApproval records the leader's evidence-review receipt. The
// leader path is the only caller (REQ-FCR-014): a receipt carrying the
// performing lane's role marker is refused here before it can ever reach a
// gate. Re-issuance for the same (card, run) replaces — the leader
// re-approves at the new version after a transition bumped the card.
func (f *FactoryDB) IssueLeaderApproval(ctx context.Context, a LeaderApproval) (LeaderApproval, error) {
	a.CardUUID = strings.TrimSpace(a.CardUUID)
	a.RunID = strings.TrimSpace(a.RunID)
	a.CardID = strings.TrimSpace(a.CardID)
	a.Issuer = strings.TrimSpace(a.Issuer)
	if a.CardUUID == "" || a.RunID == "" || a.CardID == "" {
		return LeaderApproval{}, fmt.Errorf("%w: leader approval requires card uuid, run id, and card id", ErrInvalidCardInput)
	}
	if a.IssuerRole != ApprovalIssuerLeader {
		return LeaderApproval{}, fmt.Errorf("%w: issuance is the leader path's act, got role %q", ErrApprovalIssuer, a.IssuerRole)
	}
	if a.Issuer == "" {
		return LeaderApproval{}, fmt.Errorf("%w: issuance requires an issuer name", ErrInvalidCardInput)
	}
	if a.IssuedAt == "" {
		a.IssuedAt = time.Now().UTC().Format(time.RFC3339Nano)
	}
	err := retryFactoryBusy(ctx, func() error {
		_, err := f.DB.ExecContext(ctx, `INSERT INTO leader_approvals(card_uuid,run_id,card_id,factory_version,evidence_hash,issuer,issuer_role,issued_at) VALUES(?,?,?,?,?,?,?,?)
			ON CONFLICT(card_uuid,run_id) DO UPDATE SET card_id=excluded.card_id,factory_version=excluded.factory_version,evidence_hash=excluded.evidence_hash,issuer=excluded.issuer,issuer_role=excluded.issuer_role,issued_at=excluded.issued_at`,
			a.CardUUID, a.RunID, a.CardID, a.FactoryVersion, a.EvidenceHash, a.Issuer, a.IssuerRole, a.IssuedAt)
		return err
	})
	if err != nil {
		return LeaderApproval{}, err
	}
	return a, nil
}

func scanLeaderApproval(row rowScanner) (LeaderApproval, error) {
	var a LeaderApproval
	err := row.Scan(&a.CardUUID, &a.RunID, &a.CardID, &a.FactoryVersion, &a.EvidenceHash, &a.Issuer, &a.IssuerRole, &a.IssuedAt)
	return a, err
}

// FindLeaderApproval reads the most recent receipt issued for a card uuid,
// across every run. Read-only.
func (f *FactoryDB) FindLeaderApproval(ctx context.Context, cardUUID string) (LeaderApproval, error) {
	a, err := scanLeaderApproval(f.DB.QueryRowContext(ctx,
		`SELECT card_uuid,run_id,card_id,factory_version,evidence_hash,issuer,issuer_role,issued_at FROM leader_approvals WHERE card_uuid=? ORDER BY issued_at DESC LIMIT 1`, cardUUID))
	if errors.Is(err, sql.ErrNoRows) {
		return LeaderApproval{}, ErrApprovalMissing
	}
	return a, err
}

// findLeaderApprovalForCardTx reads the most recent receipt stored for the
// (run, card) pair inside the caller's transaction — the done transitions'
// lookup key. A receipt stored under another run is simply absent here, and
// VerifyBinding's run check refuses a row whose run_id was pointed elsewhere
// anyway.
func findLeaderApprovalForCardTx(ctx context.Context, q queryRower, runID, cardID string) (LeaderApproval, error) {
	a, err := scanLeaderApproval(q.QueryRowContext(ctx,
		`SELECT card_uuid,run_id,card_id,factory_version,evidence_hash,issuer,issuer_role,issued_at FROM leader_approvals WHERE run_id=? AND card_id=? ORDER BY issued_at DESC LIMIT 1`, runID, cardID))
	if errors.Is(err, sql.ErrNoRows) {
		return LeaderApproval{}, ErrApprovalMissing
	}
	return a, err
}

// LatestCardByID reads a card's most recently updated factory row across all
// runs — the card's current factory engagement. Read-only. ErrCardNotFound
// when the card has no factory row (not factory-linked).
func (f *FactoryDB) LatestCardByID(ctx context.Context, cardID string) (Card, error) {
	c, err := scanCard(f.DB.QueryRowContext(ctx,
		`SELECT run_id,card_id,owner_label,state,version,evidence_path,updated_at,stage,lease_holder,lease_expires_at,heartbeat_at,decision_gate,decision_question,decision_resume,decider,decided_at,failure_reason,hint_prefer,hint_after,spec_id,worktree_path,evidence_sha,merge_sha,merge_tree,remeasure_path,contract_spec_id,contract_sha256,contract_signed_at,contract_event FROM cards WHERE card_id=? ORDER BY updated_at DESC LIMIT 1`, cardID))
	if errors.Is(err, sql.ErrNoRows) {
		return Card{}, ErrCardNotFound
	}
	return c, err
}

// ApprovalGate is the serialization point a backlog close verifies its
// receipt at: a factory write transaction (the DSN opens transactions
// IMMEDIATE) held across the close, so a concurrent factory transition
// cannot interleave between the archive-moment verification and the close —
// it waits on the write lock, and the close's verification reads the row
// only after the transition has settled. Hold → Verify → settle (Commit
// after the guarded close lands, Rollback on any refusal or close failure).
type ApprovalGate struct {
	tx     *sql.Tx
	card   Card
	linked bool
	done   bool
}

// HoldApprovalGate opens the gate for a queue card. A card with no factory
// row is not factory-linked; the gate returns unlinked and Verify is a
// no-op — the completion surface behaves exactly as before (REQ-FCR-002's
// scope sentence). The caller MUST settle the gate (Commit or Rollback) in
// every path.
func (f *FactoryDB) HoldApprovalGate(ctx context.Context, cardID string) (*ApprovalGate, error) {
	var tx *sql.Tx
	if err := retryFactoryBusy(ctx, func() error {
		var err error
		tx, err = f.DB.BeginTx(ctx, nil)
		return err
	}); err != nil {
		return nil, err
	}
	gate := &ApprovalGate{tx: tx}
	c, err := loadLatestCardByIDTx(ctx, tx, cardID)
	if errors.Is(err, ErrCardNotFound) {
		return gate, nil
	}
	if err != nil {
		_ = tx.Rollback()
		return nil, err
	}
	gate.card, gate.linked = c, true
	return gate, nil
}

// Linked reports whether the gated card has a factory row.
func (g *ApprovalGate) Linked() bool { return g.linked }

// Card is the factory row the gate read inside its transaction — the
// archive-moment state the verification binds against.
func (g *ApprovalGate) Card() Card { return g.card }

// Verify checks the receipt for a factory-linked card against the row the
// gate's transaction read. The full quadruple binding is checked: the card
// uuid the backlog close knows, the run, the factory version, and the
// evidence hash — all read from the archive-moment row (REQ-FCR-002a,
// REQ-FCR-004).
func (g *ApprovalGate) Verify(ctx context.Context, cardUUID string) error {
	if !g.linked {
		return nil
	}
	// A card can carry receipts for several runs (it was re-dispatched after
	// an earlier run). The receipt bound to the card's CURRENT run is the
	// one that matters: it wins over a newer receipt minted against an old
	// run, and among same-run receipts the latest issued_at wins.
	a, err := scanLeaderApproval(g.tx.QueryRowContext(ctx,
		`SELECT card_uuid,run_id,card_id,factory_version,evidence_hash,issuer,issuer_role,issued_at FROM leader_approvals WHERE card_uuid=? ORDER BY (run_id=?) DESC, issued_at DESC LIMIT 1`, cardUUID, g.card.RunID))
	if errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("%w: card %s needs a leader approval receipt to complete", ErrApprovalMissing, g.card.CardID)
	}
	if err != nil {
		return err
	}
	return a.VerifyBinding(cardUUID, g.card.RunID, g.card.Version, g.card.EvidenceSHA, g.card.OwnerLabel)
}

// Commit settles the gate after the close it guards has landed.
func (g *ApprovalGate) Commit() error {
	if g.done {
		return nil
	}
	g.done = true
	return g.tx.Commit()
}

// Rollback settles the gate without committing — every refusal and every
// close failure takes this path, and the transaction wrote nothing either
// way.
func (g *ApprovalGate) Rollback() error {
	if g.done {
		return nil
	}
	g.done = true
	return g.tx.Rollback()
}

func loadLatestCardByIDTx(ctx context.Context, q queryRower, cardID string) (Card, error) {
	c, err := scanCard(q.QueryRowContext(ctx,
		`SELECT run_id,card_id,owner_label,state,version,evidence_path,updated_at,stage,lease_holder,lease_expires_at,heartbeat_at,decision_gate,decision_question,decision_resume,decider,decided_at,failure_reason,hint_prefer,hint_after,spec_id,worktree_path,evidence_sha,merge_sha,merge_tree,remeasure_path,contract_spec_id,contract_sha256,contract_signed_at,contract_event FROM cards WHERE card_id=? ORDER BY updated_at DESC LIMIT 1`, cardID))
	if errors.Is(err, sql.ErrNoRows) {
		return Card{}, ErrCardNotFound
	}
	return c, err
}

// verifyTransitionApproval is the done transitions' receipt gate
// (REQ-FCR-002b): it runs inside FactoryDB.Transition's own transaction, so
// the version and evidence it binds against are the row as the transition
// will commit it. The uuid axis is skipped — the factory row carries no
// backlog identity — and is enforced by the CLI surfaces, which do know it.
func verifyTransitionApproval(ctx context.Context, tx *sql.Tx, cur Card) error {
	a, err := findLeaderApprovalForCardTx(ctx, tx, cur.RunID, cur.CardID)
	if errors.Is(err, ErrApprovalMissing) {
		return fmt.Errorf("%w: card %s cannot complete without a leader approval receipt", ErrApprovalMissing, cur.CardID)
	}
	if err != nil {
		return err
	}
	return a.VerifyBinding("", cur.RunID, cur.Version, cur.EvidenceSHA, cur.OwnerLabel)
}
