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
	"strconv"
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
	// ErrApprovalRunUnresolvable — the card's factory rows all live in runs
	// this store no longer describes, so the current dispatch cannot be
	// determined. A completion gate refuses; a scan reads it as an
	// unanswerable question (ReceiptGateUnknown).
	ErrApprovalRunUnresolvable = errors.New("leader approval run unresolvable")
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

// canonicalApprovalOwnerLabel mirrors internal/factory's canonical
// owner-label mapper (REQ-TSP-051): legacy leader spellings map to
// `leader[-suffix]`, and `worker`/`agent`[/`-<n>`] map to
// `lane[|-<n>]`. homestate cannot import factory (the layering guard,
// TestHomestateDoesNotImportFactory), so the mapping is duplicated here and
// pinned equal by TestOwnerLabelNormalizationMatchesFactory in the external
// test package. Idempotent: a canonical label returns unchanged.
func CanonicalApprovalOwnerLabel(label string) string {
	switch {
	case label == "lead" || strings.HasPrefix(label, "lead-"):
		return "leader" + strings.TrimPrefix(label, "lead")
	case label == "worker" || label == "agent":
		return "lane"
	case strings.HasPrefix(label, "worker-"), strings.HasPrefix(label, "agent-"):
		role, suffix, found := strings.Cut(label, "-")
		if n, err := strconv.Atoi(suffix); found && err == nil && n >= 1 {
			_ = role
			return "lane-" + suffix
		}
	}
	return label
}

// VerifyBinding checks the receipt against the named binding values — the
// verifier's own identity knowledge, not a caller-supplied verdict. cardUUID
// is the backlog identity; a verifier that cannot know it (the factory row
// carries no uuid) passes "" and the uuid axis is skipped rather than failed
// — the CLI surfaces, which do know it, always pass it. ownerLabel is the
// factory row's performing owner: a receipt issued by that owner is
// refused wherever it is presented (REQ-FCR-005 — performer ≠ approver; the
// role marker is data, and this comparison is the gate's own), skipped only
// when the row names no owner.
func (a LeaderApproval) VerifyBinding(cardUUID, runID string, version int64, evidenceHash, ownerLabel string) error {
	if a.IssuerRole != ApprovalIssuerLeader {
		return fmt.Errorf("%w: issuer %q holds role %q, not %q", ErrApprovalIssuer, a.Issuer, a.IssuerRole, ApprovalIssuerLeader)
	}
	if ownerLabel != "" && CanonicalApprovalOwnerLabel(a.Issuer) == CanonicalApprovalOwnerLabel(ownerLabel) {
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
	// An empty hash is NOT a valid evidence binding (review round-10 P2-1):
	// a card with no recorded evidence cannot satisfy the evidence axis, on
	// the card side or on the receipt side.
	if evidenceHash == "" || a.EvidenceHash == "" {
		return fmt.Errorf("%w: empty evidence hash — the receipt binds %q, the card carries %q", ErrApprovalHashMismatch, a.EvidenceHash, evidenceHash)
	}
	if a.EvidenceHash != evidenceHash {
		return fmt.Errorf("%w: receipt binds evidence %q, the card carries %q", ErrApprovalHashMismatch, a.EvidenceHash, evidenceHash)
	}
	return nil
}

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
	// An empty evidence hash is not an evidence binding (review round-10
	// P2-1): the leader reviewed SOMETHING, and the hash is what the gate
	// re-checks at close time.
	if strings.TrimSpace(a.EvidenceHash) == "" {
		return LeaderApproval{}, fmt.Errorf("%w: issuance requires the reviewed evidence hash", ErrApprovalHashMismatch)
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
// across every run and with no run preference. Read-only; the verification
// paths use findLeaderApprovalPreferringRun instead.
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

// findLeaderApprovalPreferringRun selects the receipt a verification judges:
// the one bound to the named current run wins over a newer receipt minted
// against an older run, and among same-run receipts the latest issued_at
// wins. Shared by the gate and the read-only scan path, so the two can never
// disagree about which receipt they are looking at.
func findLeaderApprovalPreferringRun(ctx context.Context, q queryRower, cardUUID, runID string) (LeaderApproval, error) {
	a, err := scanLeaderApproval(q.QueryRowContext(ctx,
		`SELECT card_uuid,run_id,card_id,factory_version,evidence_hash,issuer,issuer_role,issued_at FROM leader_approvals WHERE card_uuid=? ORDER BY (run_id=?) DESC, issued_at DESC LIMIT 1`, cardUUID, runID))
	if errors.Is(err, sql.ErrNoRows) {
		return LeaderApproval{}, ErrApprovalMissing
	}
	return a, err
}

// recordedCardRow resolves the factory row a backlog card's verification
// binds against: the row the card's DISPATCH BINDING names. The binding is
// recorded at assignment time — RecordPicked writes it in the same
// transaction that creates or re-records the card row — so it cannot drift
// from the assign path the way a modification time (round-2), a runtime
// assignment record (round-3), or run creation order (round-4) all
// measured dead. A card with no binding, or a binding whose row is gone,
// is unresolvable: the resolution refuses (ErrApprovalRunUnresolvable) and
// the recovery is one idempotent command — re-record the card with
// `factory assign <card> --run <run>`.

// queryRowerEx is the read surface recordedCardRow needs: single-row and
// multi-row queries inside the caller's transaction or on the bare handle.
// *sql.DB and *sql.Tx both satisfy it.
type queryRowerEx interface {
	queryRower
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
}

func recordedCardRow(ctx context.Context, q queryRowerEx, cardID string) (Card, bool, error) {
	var runID string
	err := q.QueryRowContext(ctx, `SELECT run_id FROM card_dispatch WHERE card_id=?`, cardID).Scan(&runID)
	if errors.Is(err, sql.ErrNoRows) {
		// Distinguish a card the factory has never engaged (not linked —
		// the gate does not apply) from rows orphaned from their dispatch
		// record (unresolvable — fail closed).
		var rows int
		if err := q.QueryRowContext(ctx, `SELECT count(*) FROM cards WHERE card_id=?`, cardID).Scan(&rows); err != nil {
			return Card{}, false, err
		}
		if rows == 0 {
			return Card{}, false, nil
		}
		return Card{}, false, fmt.Errorf("%w: card %s has factory rows but no recorded dispatch binding — recover it with `factory assign %s --run <run> --to <owner>` (review round-4, SPEC-FACTORY-COMPLETION-RECOVERY-001)", ErrApprovalRunUnresolvable, cardID, cardID)
	}
	if err != nil {
		return Card{}, false, err
	}
	row, err := loadCard(ctx, q, runID, cardID)
	if errors.Is(err, ErrCardNotFound) {
		return Card{}, false, fmt.Errorf("%w: card %s dispatch binding names run %q, which holds no row for the card", ErrApprovalRunUnresolvable, cardID, runID)
	}
	if err != nil {
		return Card{}, false, err
	}
	return row, true, nil
}

// ensureRunRowTx inserts the run row when missing so a bound card is never
// stranded without the metadata its dispatch names.
func ensureRunRowTx(ctx context.Context, tx *sql.Tx, runID, nowText string) error {
	// SQL: 'active' is the runs table's status literal
	// (factory_run_retire.go classifies on the same literal).
	_, err := tx.ExecContext(ctx, `INSERT INTO runs(run_id,status,manifest_json,created_at,updated_at) VALUES(?,'active','{}',?,?) ON CONFLICT(run_id) DO NOTHING`,
		runID, nowText, nowText)
	return err
}

// upsertDispatchBindingTx records cardID -> runID as the card's current
// factory engagement inside the caller's transaction.
func upsertDispatchBindingTx(ctx context.Context, tx *sql.Tx, cardID, runID, nowText string) error {
	_, err := tx.ExecContext(ctx, `INSERT INTO card_dispatch(card_id,run_id,recorded_at) VALUES(?,?,?) ON CONFLICT(card_id) DO UPDATE SET run_id=excluded.run_id,recorded_at=excluded.recorded_at`,
		cardID, runID, nowText)
	return err
}

// RecordDispatchBinding records the card's dispatch binding — THIS run is
// the card's current factory engagement — and ensures the named run row in
// one transaction. Assignment callers invoke it when an assignment succeeds
// without the T2 transition (the idempotent same-lane re-assign), so the
// binding follows every successful assignment path and provides the
// documented recovery for pre-binding rows at assigned-or-later states
// (review round-4 P1-1/P1-2).
func (f *FactoryDB) RecordDispatchBinding(ctx context.Context, cardID, runID string, now time.Time) error {
	cardID, runID = strings.TrimSpace(cardID), strings.TrimSpace(runID)
	if cardID == "" || runID == "" {
		return fmt.Errorf("%w: card id and run id are required", ErrInvalidCardInput)
	}
	if now.IsZero() {
		now = time.Now()
	}
	nowText := now.UTC().Format(time.RFC3339Nano)
	return retryFactoryBusy(ctx, func() error {
		tx, err := f.DB.BeginTx(ctx, nil)
		if err != nil {
			return err
		}
		defer func() { _ = tx.Rollback() }()
		if err := ensureRunRowTx(ctx, tx, runID, nowText); err != nil {
			return err
		}
		if err := upsertDispatchBindingTx(ctx, tx, cardID, runID, nowText); err != nil {
			return err
		}
		return tx.Commit()
	})
}

// VerifyApprovalReadonly is the scan-time counterpart of the gate: the same
// row resolution and the same receipt selection as the archive-moment gate,
// outside any transaction. Its verdict is advisory — the gate is
// authoritative. ErrCardNotFound when the card has no factory row;
// ErrApprovalRunUnresolvable when the dispatch cannot be determined.
func (f *FactoryDB) VerifyApprovalReadonly(ctx context.Context, cardID, cardUUID string) error {
	row, linked, err := recordedCardRow(ctx, f.DB, cardID)
	if err != nil {
		return err
	}
	if !linked {
		return ErrCardNotFound
	}
	a, err := findLeaderApprovalPreferringRun(ctx, f.DB, cardUUID, row.RunID)
	if err != nil {
		return err
	}
	return a.VerifyBinding(cardUUID, row.RunID, row.Version, row.EvidenceSHA, row.OwnerLabel)
}

// CardRowExistsReadonly reports whether the card has any factory row,
// without touching the dispatch or receipt tables (an older-schema store
// may lack both). Read-only safe; the REQ-FCR-002 scoping decision for a
// pre-binding store uses this.
func (f *FactoryDB) CardRowExistsReadonly(ctx context.Context, cardID string) (bool, error) {
	var n int
	err := f.DB.QueryRowContext(ctx, `SELECT count(*) FROM cards WHERE card_id=?`, cardID).Scan(&n)
	return n > 0, err
}

// RecordedCardRowReadonly resolves the card's dispatch factory row for
// read-only consumers that must not touch the receipt table (an
// older-schema store may not have one). linked=false when the card has no
// factory row; ErrApprovalRunUnresolvable when the dispatch cannot be
// determined.
func (f *FactoryDB) RecordedCardRowReadonly(ctx context.Context, cardID string) (Card, bool, error) {
	return recordedCardRow(ctx, f.DB, cardID)
}

// ApprovalGate is the serialization point a backlog close verifies its
// receipt at: a factory write transaction (the DSN opens transactions
// IMMEDIATE) held across the close's PERSISTENCE, so a concurrent factory
// transition cannot interleave between the archive-moment verification and
// the queue write — it waits on the write lock, and the close's
// verification reads rows only after the transition has settled. The
// transaction never writes, so the gate settles with Rollback on every path
// (after a refusal, after a failure, and after the guarded mutation has
// persisted — for a read-only transaction rollback and commit are
// equivalent, and rollback cannot fail the close it guarded).
type ApprovalGate struct {
	tx   *sql.Tx
	done bool
}

// HoldApprovalGate opens the gate. The caller MUST settle it (Commit or
// Rollback) on every path.
func (f *FactoryDB) HoldApprovalGate(ctx context.Context) (*ApprovalGate, error) {
	var tx *sql.Tx
	if err := retryFactoryBusy(ctx, func() error {
		var err error
		tx, err = f.DB.BeginTx(ctx, nil)
		return err
	}); err != nil {
		return nil, err
	}
	return &ApprovalGate{tx: tx}, nil
}

// Row reads the archive-moment factory row for cardID inside the gate's
// transaction, resolved through the runs table's dispatch binding (see
// recordedCardRow). linked=false when the card has no factory row — the card
// is not factory-linked and the receipt gate does not apply (REQ-FCR-002's
// scope sentence).
func (g *ApprovalGate) Row(ctx context.Context, cardID string) (card Card, linked bool, err error) {
	c, linked, err := recordedCardRow(ctx, g.tx, cardID)
	if err != nil || !linked {
		return Card{}, false, err
	}
	return c, true, nil
}

// Verify checks the receipt bound to cardUUID against the archive-moment
// row. The receipt bound to the row's CURRENT run is preferred — a newer
// receipt minted against an older run never shadows it — and the full
// quadruple binding plus the performer check run here (REQ-FCR-002a,
// REQ-FCR-004, REQ-FCR-005).
func (g *ApprovalGate) Verify(ctx context.Context, card Card, cardUUID string) error {
	a, err := findLeaderApprovalPreferringRun(ctx, g.tx, cardUUID, card.RunID)
	if errors.Is(err, ErrApprovalMissing) {
		return fmt.Errorf("%w: card %s needs a leader approval receipt to complete", ErrApprovalMissing, card.CardID)
	}
	if err != nil {
		return err
	}
	return a.VerifyBinding(cardUUID, card.RunID, card.Version, card.EvidenceSHA, card.OwnerLabel)
}

// Commit settles the gate by committing its (read-only) transaction.
func (g *ApprovalGate) Commit() error {
	if g.done {
		return nil
	}
	g.done = true
	return g.tx.Commit()
}

// Rollback settles the gate without committing. Because the transaction
// never writes, this is the settle path on EVERY exit: refusal, failure,
// and after the guarded mutation has persisted alike.
func (g *ApprovalGate) Rollback() error {
	if g.done {
		return nil
	}
	g.done = true
	return g.tx.Rollback()
}

// verifyTransitionApproval is the done transitions' receipt gate
// (REQ-FCR-002b): it runs inside FactoryDB.Transition's own transaction, so
// the version and evidence it binds against are the row as the transition
// will commit it. approvalUUID is the backlog identity the caller resolved
// for this card; an empty value is refused — a done edge must never skip
// the card-identity check (review round-6 P1-2).
func verifyTransitionApproval(ctx context.Context, tx *sql.Tx, cur Card, approvalUUID string) error {
	approvalUUID = strings.TrimSpace(approvalUUID)
	if approvalUUID == "" {
		return fmt.Errorf("%w: card %s completion requires its backlog card uuid for the approval check", ErrApprovalCardMismatch, cur.CardID)
	}
	a, err := findLeaderApprovalForCardTx(ctx, tx, cur.RunID, cur.CardID)
	if errors.Is(err, ErrApprovalMissing) {
		return fmt.Errorf("%w: card %s cannot complete without a leader approval receipt", ErrApprovalMissing, cur.CardID)
	}
	if err != nil {
		return err
	}
	return a.VerifyBinding(approvalUUID, cur.RunID, cur.Version, cur.EvidenceSHA, cur.OwnerLabel)
}
