package homestate

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"
)

// Transition-API refusals. Every refusal leaves the card row and the event
// log exactly as they were; callers match with errors.Is.
var (
	ErrStaleVersion        = errors.New("stale card version")
	ErrIllegalTransition   = errors.New("illegal card transition")
	ErrReservedEdge        = errors.New("reserved card transition")
	ErrLegacyState         = errors.New("legacy card state")
	ErrEvidence            = errors.New("card evidence refused")
	ErrLeaseExpired        = errors.New("card lease expired")
	ErrLeaseHolder         = errors.New("card lease holder refused")
	ErrUnknownPredecessor  = errors.New("unknown predecessor card")
	ErrPredecessorUnmerged = errors.New("predecessor card not merged")
	ErrDecider             = errors.New("card decider refused")
	ErrInvalidCardInput    = errors.New("invalid card transition input")
)

// TransitionRequest asks the transition API to move one card. Callers supply
// identities, paths, and commit SHAs only — never a verdict, a pass flag, a
// tree hash, or an ancestry claim (REQ-FR-010); the API reads the evidence
// itself.
type TransitionRequest struct {
	RunID, CardID, To string
	// ExpectedVersion is the version the caller last read; a mismatch is
	// refused with ErrStaleVersion.
	ExpectedVersion int64
	// Actor names who asks: a worker label for holder edges, an operator or
	// command name otherwise.
	Actor string
	// Decider is required on operator-decision edges; F1 accepts only "human".
	Decider string
	// Owner is the lane label T2 (picked → assigned) records.
	Owner string
	// SHA and ArtifactPath are the audit-entry evidence (T5, T11) and the
	// run → sync commit (T10).
	SHA, ArtifactPath string
	// MergeSHA and RemeasurePath are the merged-local evidence (T16).
	MergeSHA, RemeasurePath string
	// IntegrationBranch names the local integration branch (T16, T17).
	IntegrationBranch string
	// Question is required entering needs-decision; Reason entering failed.
	Question, Reason string
	// Now is the injected clock; zero means time.Now().
	Now time.Time
}

// TransitionEdge is one requested (from, to) pair of the transition table.
type TransitionEdge struct{ ID, From, To string }

type edgeGuard int

const (
	guardAssign edgeGuard = iota + 1
	guardLeaseAcquire
	guardStageResume
	guardEntry
	guardVerdictAny
	guardVerdictPass
	guardKickoffDecision
	guardHuman
	guardCommit
	guardLeaseValid
	guardNone
	guardMerge
	guardPush
	guardNoRemote
	guardQuestion
	guardResumeDecision
	guardUnblock
	guardAbandon
	guardFail
)

type transitionEdge struct {
	id       string
	from, to string
	guard    edgeGuard
}

// @MX:ANCHOR: [AUTO] the F1 card transition table — the complete set of requested edges the record accepts
// @MX:REASON: AC-005 pins its size at 65 accepted pairs; adding, dropping, or re-guarding a row changes what every factory writer may do (REQ-FR-004)
var transitionTable = buildTransitionTable()

func buildTransitionTable() []transitionEdge {
	t := []transitionEdge{
		{"T2", CardPicked, CardAssigned, guardAssign},
		{"T3", CardAssigned, CardLeased, guardLeaseAcquire},
		{"T4a", CardLeased, CardPlan, guardStageResume},
		{"T4b", CardLeased, CardRun, guardStageResume},
		{"T4c", CardLeased, CardPlanAudit, guardStageResume},
		{"T4d", CardLeased, CardSync, guardStageResume},
		{"T4e", CardLeased, CardSyncAudit, guardStageResume},
		{"T4f", CardLeased, CardMergeReady, guardStageResume},
		{"T5", CardPlan, CardPlanAudit, guardEntry},
		{"T6", CardPlanAudit, CardPlan, guardVerdictAny},
		{"T7", CardPlanAudit, CardKickoff, guardVerdictPass},
		{"T8", CardKickoff, CardAssigned, guardKickoffDecision},
		{"T9", CardKickoff, CardBlocked, guardKickoffDecision},
		{"T10", CardRun, CardSync, guardCommit},
		{"T11", CardSync, CardSyncAudit, guardEntry},
		{"T12", CardSyncAudit, CardSync, guardVerdictAny},
		{"T13", CardSyncAudit, CardMergeReady, guardVerdictPass},
		{"T14", CardMergeReady, CardMerging, guardLeaseValid},
		{"T15", CardMerging, CardMergeReady, guardNone},
		{"T16", CardMerging, CardMergedLocal, guardMerge},
		{"T17", CardMergedLocal, CardPushed, guardPush},
		{"T18", CardMergedLocal, CardDone, guardNoRemote},
	}
	t21 := append(append([]string{CardPicked, CardAssigned}, leaseHoldingStates...), CardMergedLocal, CardPushed, CardCIGreen)
	for _, from := range t21 {
		t = append(t, transitionEdge{"T21", from, CardNeedsDecision, guardQuestion})
	}
	for i, to := range []string{CardAssigned, CardPicked, CardMergedLocal, CardPushed, CardCIGreen} {
		t = append(t, transitionEdge{"T22" + string(rune('a'+i)), CardNeedsDecision, to, guardResumeDecision})
	}
	t = append(t,
		transitionEdge{"T23", CardNeedsDecision, CardBlocked, guardHuman},
		transitionEdge{"T24", CardBlocked, CardAssigned, guardUnblock},
	)
	for _, from := range cardStates {
		if !IsTerminalCardState(from) {
			t = append(t, transitionEdge{"T25", from, CardAbandoned, guardAbandon})
		}
	}
	for _, from := range leaseHoldingStates {
		t = append(t, transitionEdge{"T26", from, CardFailed, guardFail})
	}
	return t
}

// TransitionEdges returns the requested edges of the transition table
// (T2-T26 without the reserved T19/T20; T1 creates a row and T27/T28 are
// automatic, so none of them is a requested pair).
func TransitionEdges() []TransitionEdge {
	out := make([]TransitionEdge, 0, len(transitionTable))
	for _, e := range transitionTable {
		out = append(out, TransitionEdge{ID: e.id, From: e.from, To: e.to})
	}
	return out
}

func findEdge(from, to string) (transitionEdge, bool) {
	for _, e := range transitionTable {
		if e.from == from && e.to == to {
			return e, true
		}
	}
	return transitionEdge{}, false
}

func isReservedEdge(from, to string) bool {
	return (from == CardPushed && to == CardCIGreen) || (from == CardCIGreen && to == CardDone)
}

// resumeTarget maps a needs-decision card's recorded resume state onto the
// state a `resume` decision returns it to (T22a-e): a lease-holding stage
// returns through `assigned`, so the owner re-leases before resuming.
func resumeTarget(resume string) string {
	switch {
	case resume == CardAssigned || IsLeaseHoldingState(resume):
		return CardAssigned
	case slices.Contains([]string{CardPicked, CardMergedLocal, CardPushed, CardCIGreen}, resume):
		return resume
	default:
		return ""
	}
}

// ResumeTarget is resumeTarget for callers outside the package (the decide
// command computes the requested target from the card it read).
func ResumeTarget(resume string) string { return resumeTarget(resume) }

// cardTransitionFault is a test seam: when set, it is called at named points
// inside the transition transaction and a non-nil return aborts it.
var cardTransitionFault func(stage string) error

func injectFault(stage string) error {
	if cardTransitionFault == nil {
		return nil
	}
	return cardTransitionFault(stage)
}

// transitionPlan is the computed effect of one accepted transition.
type transitionPlan struct {
	next     Card
	kind     string
	evidence map[string]string
	note     string
	// keepColumns leaves every column but state and version as they were
	// (the operator abandon, which records nothing else about the card).
	keepColumns bool
}

// Transition applies one requested card transition as a single transaction:
// it compares the expected version, verifies the edge and its guard (reading
// evidence itself), writes the new row with version+1, and appends one
// `card.transition` event — all of these or none.
//
// @MX:ANCHOR: [AUTO] the only writer that moves a factory card between states
// @MX:REASON: every factory writer (assign, decide, both dispatch mirrors, and the F2/F3 lane and controller verbs) funnels through this version-checked transaction (REQ-FR-005)
func (f *FactoryDB) Transition(ctx context.Context, req TransitionRequest) (Card, error) {
	if strings.TrimSpace(req.RunID) == "" || strings.TrimSpace(req.CardID) == "" {
		return Card{}, fmt.Errorf("%w: run id and card id are required", ErrInvalidCardInput)
	}
	if !IsCardState(req.To) {
		return Card{}, fmt.Errorf("%w: unknown target state %q", ErrInvalidCardInput, req.To)
	}
	now := req.Now
	if now.IsZero() {
		now = time.Now()
	}
	now = now.UTC()
	var result Card
	err := f.withCardTx(ctx, req.RunID, func(tx *sql.Tx) (func(), error) {
		cur, err := loadCard(ctx, tx, req.RunID, req.CardID)
		if err != nil {
			return nil, err
		}
		// An expired lease is returned before any other transition (T27/T28),
		// whatever the request asked for and whatever version it carried.
		if cur.LeaseExpired(now) {
			if err := applyLeaseExpiry(ctx, tx, cur, now); err != nil {
				return nil, err
			}
			return nil, commitThen(fmt.Errorf("%w: card %s lease held by %q expired at %s", ErrLeaseExpired, cur.CardID, cur.LeaseHolder, cur.LeaseExpiresAt))
		}
		// The version compare comes first: a caller whose read is stale has no
		// standing to be told anything about the current state's edges.
		if cur.Version != req.ExpectedVersion {
			return nil, staleErr(cur, req.ExpectedVersion)
		}
		if cur.Legacy() {
			if req.To != CardAbandoned {
				return nil, fmt.Errorf("%w: card %s holds pre-F1 state %q; only an operator abandon is accepted", ErrLegacyState, cur.CardID, cur.State)
			}
			if req.Decider != DeciderHuman {
				return nil, fmt.Errorf("%w: abandon requires decider %q", ErrDecider, DeciderHuman)
			}
			next := cur
			next.State = CardAbandoned
			result, err = commitTransition(ctx, tx, cur, transitionPlan{next: next, kind: "card.transition", keepColumns: true}, req, now)
			return nil, err
		}
		if isReservedEdge(cur.State, req.To) {
			return nil, fmt.Errorf("%w: %s → %s is reserved; the CI verdict reader that admits it is owned by F3", ErrReservedEdge, cur.State, req.To)
		}
		edge, ok := findEdge(cur.State, req.To)
		if !ok {
			return nil, fmt.Errorf("%w: %s → %s", ErrIllegalTransition, cur.State, req.To)
		}
		plan, err := f.planTransition(ctx, tx, cur, edge, req, now)
		if err != nil {
			return nil, err
		}
		result, err = commitTransition(ctx, tx, cur, plan, req, now)
		return nil, err
	})
	if err != nil {
		return Card{}, err
	}
	return result, nil
}

func staleErr(cur Card, expected int64) error {
	return fmt.Errorf("%w: card %s is at version %d, request expected %d", ErrStaleVersion, cur.CardID, cur.Version, expected)
}

// withCardTx runs fn inside one write transaction (the database's DSN takes
// the write lock at BEGIN). fn may return a post-commit callback.
func (f *FactoryDB) withCardTx(ctx context.Context, _ string, fn func(tx *sql.Tx) (func(), error)) error {
	var tx *sql.Tx
	if err := retryFactoryBusy(ctx, func() error {
		var err error
		tx, err = f.DB.BeginTx(ctx, nil)
		return err
	}); err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	after, err := fn(tx)
	var keep *committedRefusal
	if err != nil && !errors.As(err, &keep) {
		return err
	}
	if cerr := tx.Commit(); cerr != nil {
		return cerr
	}
	if after != nil {
		after()
	}
	if keep != nil {
		return keep.err
	}
	return nil
}

// committedRefusal marks a refusal whose transaction still commits — the
// lease-expiry return is written, and the request that found it is refused.
type committedRefusal struct{ err error }

func (c *committedRefusal) Error() string { return c.err.Error() }
func (c *committedRefusal) Unwrap() error { return c.err }

func commitThen(err error) error { return &committedRefusal{err: err} }

// applyLeaseExpiry is T27/T28: an expired lease-holding card returns to
// `assigned` (holder cleared; stage, worktree, and evidence kept), or to
// `blocked` when it expired mid-merge. It reads and writes the record only —
// never the card's worktree.
func applyLeaseExpiry(ctx context.Context, tx *sql.Tx, cur Card, now time.Time) error {
	next := cur
	next.State = CardAssigned
	payload := map[string]any{"card_id": cur.CardID, "from": cur.State, "holder": cur.LeaseHolder, "expired_at": cur.LeaseExpiresAt}
	if cur.State == CardMerging {
		next.State = CardBlocked
		payload["interrupted"] = "merge"
		payload["merge_sha"] = cur.MergeSHA
	}
	next.LeaseHolder, next.LeaseExpiresAt = "", ""
	next.Version = cur.Version + 1
	next.UpdatedAt = now.Format(time.RFC3339Nano)
	payload["to"], payload["version"] = next.State, next.Version
	if err := updateCardRow(ctx, tx, next, cur.Version); err != nil {
		return err
	}
	return appendEvent(ctx, tx, cur.RunID, "lease.expired", payload, now)
}

// holderGuards are the edges only the current lease holder may request.
var holderGuards = map[edgeGuard]bool{
	guardStageResume: true, guardEntry: true, guardVerdictAny: true, guardVerdictPass: true,
	guardCommit: true, guardLeaseValid: true, guardNone: true, guardMerge: true, guardFail: true,
}

// RenewLease extends the holder's lease on a lease-holding card and refreshes
// the roster heartbeat. A renewal by any other label is refused with no
// change; an already-expired lease is returned first (T27/T28). Renewal does
// not change the card's state, so it does not advance the version.
func (f *FactoryDB) RenewLease(ctx context.Context, runID, cardID, label string, now time.Time) (Card, error) {
	if now.IsZero() {
		now = time.Now()
	}
	now = now.UTC()
	var result Card
	err := f.withCardTx(ctx, runID, func(tx *sql.Tx) (func(), error) {
		cur, err := loadCard(ctx, tx, runID, cardID)
		if err != nil {
			return nil, err
		}
		if cur.LeaseExpired(now) {
			if err := applyLeaseExpiry(ctx, tx, cur, now); err != nil {
				return nil, err
			}
			return nil, commitThen(fmt.Errorf("%w: card %s lease expired at %s", ErrLeaseExpired, cur.CardID, cur.LeaseExpiresAt))
		}
		if !IsLeaseHoldingState(cur.State) || cur.LeaseHolder == "" || strings.TrimSpace(label) != cur.LeaseHolder {
			return nil, fmt.Errorf("%w: %q does not hold the lease on card %s", ErrLeaseHolder, label, cur.CardID)
		}
		next := cur
		next.HeartbeatAt = now.Format(time.RFC3339Nano)
		next.LeaseExpiresAt = now.Add(FactoryLeaseDuration).Format(time.RFC3339Nano)
		next.UpdatedAt = next.HeartbeatAt
		res, err := tx.ExecContext(ctx, `UPDATE cards SET heartbeat_at=?,lease_expires_at=?,updated_at=? WHERE run_id=? AND card_id=? AND version=?`,
			next.HeartbeatAt, next.LeaseExpiresAt, next.UpdatedAt, cur.RunID, cur.CardID, cur.Version)
		if err != nil {
			return nil, err
		}
		if n, err := res.RowsAffected(); err != nil || n == 0 {
			return nil, fmt.Errorf("%w: card %s changed under the renewal", ErrStaleVersion, cur.CardID)
		}
		if _, err := tx.ExecContext(ctx, `UPDATE workers SET heartbeat_at=? WHERE label=?`, next.HeartbeatAt, label); err != nil {
			return nil, err
		}
		result = next
		return nil, nil
	})
	if err != nil {
		return Card{}, err
	}
	return result, nil
}

// planTransition checks the edge's guard and computes the new row.
func (f *FactoryDB) planTransition(ctx context.Context, tx *sql.Tx, cur Card, edge transitionEdge, req TransitionRequest, now time.Time) (transitionPlan, error) {
	next := cur
	next.State = req.To
	plan := transitionPlan{next: next, kind: "card.transition", evidence: map[string]string{}}
	nowText := now.Format(time.RFC3339Nano)
	if holderGuards[edge.guard] && (cur.LeaseHolder == "" || strings.TrimSpace(req.Actor) != cur.LeaseHolder) {
		return plan, fmt.Errorf("%w: %s → %s is the lease holder's edge; %q does not hold the lease (holder %q)", ErrLeaseHolder, cur.State, req.To, req.Actor, cur.LeaseHolder)
	}
	switch edge.guard {
	case guardAssign:
		owner := strings.TrimSpace(req.Owner)
		if owner == "" {
			return plan, fmt.Errorf("%w: assigning a card requires an owner label", ErrInvalidCardInput)
		}
		plan.next.OwnerLabel = owner
	case guardLeaseAcquire:
		label := strings.TrimSpace(req.Actor)
		var registered int
		if err := tx.QueryRowContext(ctx, `SELECT count(*) FROM workers WHERE label=?`, label).Scan(&registered); err != nil {
			return plan, err
		}
		if label == "" || registered == 0 || label != cur.OwnerLabel {
			return plan, fmt.Errorf("%w: %q is not the registered owner %q of card %s", ErrLeaseHolder, label, cur.OwnerLabel, cur.CardID)
		}
		plan.next.LeaseHolder = label
		plan.next.HeartbeatAt = nowText
		plan.next.LeaseExpiresAt = now.Add(FactoryLeaseDuration).Format(time.RFC3339Nano)
		if _, err := tx.ExecContext(ctx, `UPDATE workers SET heartbeat_at=? WHERE label=?`, nowText, label); err != nil {
			return plan, err
		}
	case guardStageResume:
		if !stageAdmits(cur.Stage, req.To) {
			return plan, fmt.Errorf("%w: leased card stage %q does not resume into %s", ErrIllegalTransition, cur.Stage, req.To)
		}
	case guardEntry:
		path := strings.TrimSpace(req.ArtifactPath)
		full, err := verifyAuditEntry(ctx, cur.WorktreePath, req.SHA, path)
		if err != nil {
			return plan, err
		}
		plan.next.EvidenceSHA = full
		plan.next.EvidencePath = path
		plan.evidence["sha"], plan.evidence["artifact_path"] = full, path
	case guardVerdictAny, guardVerdictPass:
		phase := "plan-audit"
		if cur.State == CardSyncAudit {
			phase = "sync-audit"
		}
		v, err := readAuditVerdict(cur.WorktreePath, cur.CardID, phase, cur.EvidenceSHA)
		if err != nil {
			return plan, err
		}
		if edge.guard == guardVerdictPass && v.Verdict != "PASS" && v.Verdict != "PASS-WITH-DEBT" {
			return plan, fmt.Errorf("%w: verdict file %s reads %s", ErrEvidence, v.Path, v.Verdict)
		}
		plan.evidence["verdict_file"], plan.evidence["verdict"], plan.evidence["audited_sha"] = v.Path, v.Verdict, v.AuditedSHA
		if edge.guard == guardVerdictPass && req.To == CardKickoff {
			plan.next.DecisionGate = DecisionGateKickoff
			plan.next.DecisionResume = CardRun
			plan.next.DecisionQuestion = ""
		}
	case guardKickoffDecision:
		if req.Decider != DeciderHuman {
			return plan, fmt.Errorf("%w: F1 accepts only decider %q, got %q", ErrDecider, DeciderHuman, req.Decider)
		}
		if req.To == CardAssigned {
			plan.next.Stage = CardRun
		}
		clearDecision(&plan.next)
		plan.next.Decider, plan.next.DecidedAt = DeciderHuman, nowText
	case guardHuman, guardUnblock, guardResumeDecision:
		if req.Decider != DeciderHuman {
			return plan, fmt.Errorf("%w: F1 accepts only decider %q, got %q", ErrDecider, DeciderHuman, req.Decider)
		}
		if edge.guard == guardResumeDecision {
			if resumeTarget(cur.DecisionResume) != req.To {
				return plan, fmt.Errorf("%w: needs-decision card resumes to %q, not %s", ErrIllegalTransition, resumeTarget(cur.DecisionResume), req.To)
			}
			if isResumableStage(cur.DecisionResume) {
				plan.next.Stage = cur.DecisionResume
			}
		}
		clearDecision(&plan.next)
		plan.next.Decider, plan.next.DecidedAt = DeciderHuman, nowText
	case guardAbandon:
		if req.Decider != DeciderHuman {
			return plan, fmt.Errorf("%w: F1 accepts only decider %q, got %q", ErrDecider, DeciderHuman, req.Decider)
		}
		plan.keepColumns = true
	case guardCommit:
		full, err := verifyCommitAtHead(ctx, cur.WorktreePath, req.SHA)
		if err != nil {
			return plan, err
		}
		plan.next.EvidenceSHA = full
		plan.evidence["sha"] = full
	case guardLeaseValid, guardNone:
	case guardMerge:
		ev, err := verifyMerge(ctx, cur.WorktreePath, req.MergeSHA, req.RemeasurePath, req.IntegrationBranch)
		if err != nil {
			return plan, err
		}
		plan.next.MergeSHA, plan.next.MergeTree, plan.next.RemeasurePath = ev.SHA, ev.Tree, ev.RemeasurePath
		plan.evidence["merge_sha"], plan.evidence["merge_tree"], plan.evidence["remeasure_path"] = ev.SHA, ev.Tree, ev.RemeasurePath
	case guardPush, guardNoRemote:
		if req.Decider != DeciderHuman {
			return plan, fmt.Errorf("%w: F1 accepts only decider %q, got %q", ErrDecider, DeciderHuman, req.Decider)
		}
		remote, err := hasRemote(ctx, cur.WorktreePath)
		if err != nil {
			return plan, fmt.Errorf("%w: %v", ErrEvidence, err)
		}
		if edge.guard == guardNoRemote {
			if remote {
				return plan, fmt.Errorf("%w: merged-local → done is refused while a remote is configured", ErrIllegalTransition)
			}
			plan.note = "no remote — no CI verdict"
		} else {
			if !remote {
				return plan, fmt.Errorf("%w: merged-local → pushed requires a configured remote", ErrIllegalTransition)
			}
			ref, err := verifyPushed(ctx, cur.WorktreePath, cur.MergeSHA, req.IntegrationBranch)
			if err != nil {
				return plan, err
			}
			plan.evidence["merge_sha"], plan.evidence["remote_ref"] = cur.MergeSHA, ref
		}
		plan.next.Decider, plan.next.DecidedAt = DeciderHuman, nowText
	case guardQuestion:
		q := strings.TrimSpace(req.Question)
		if q == "" {
			return plan, fmt.Errorf("%w: entering needs-decision requires a question", ErrInvalidCardInput)
		}
		plan.next.DecisionGate = DecisionGateQuestion
		plan.next.DecisionQuestion = q
		plan.next.DecisionResume = cur.State
	case guardFail:
		reason := strings.TrimSpace(req.Reason)
		if reason == "" {
			return plan, fmt.Errorf("%w: moving a card to failed requires a reason", ErrInvalidCardInput)
		}
		plan.next.FailureReason = reason
	default:
		return plan, fmt.Errorf("%w: edge %s has no guard", ErrIllegalTransition, edge.id)
	}
	return plan, nil
}

// stageAdmits is the T4 guard: a leased card resumes into the stage it last
// reached; an empty stage admits plan and run (Class A/B cards skip plan).
func stageAdmits(stage, to string) bool {
	if stage == "" {
		return to == CardPlan || to == CardRun
	}
	return stage == to
}

func clearDecision(c *Card) {
	c.DecisionGate, c.DecisionQuestion, c.DecisionResume = "", "", ""
}

// commitTransition writes the planned row with a version compare and appends
// the transition event inside the caller's transaction.
func commitTransition(ctx context.Context, tx *sql.Tx, cur Card, plan transitionPlan, req TransitionRequest, now time.Time) (Card, error) {
	next := plan.next
	if isResumableStage(next.State) {
		next.Stage = next.State
	}
	// Leaving the lease-holding states releases the lease (REQ-FR-015,
	// REQ-FR-018): decision-pending, post-merge, paused, and failed cards hold
	// none. The operator abandon records nothing but the state.
	if !IsLeaseHoldingState(next.State) && !plan.keepColumns {
		next.LeaseHolder, next.LeaseExpiresAt = "", ""
	}
	next.Version = cur.Version + 1
	next.UpdatedAt = now.Format(time.RFC3339Nano)
	if err := updateCardRow(ctx, tx, next, cur.Version); err != nil {
		return Card{}, err
	}
	if err := injectFault("before-event"); err != nil {
		return Card{}, err
	}
	payload := map[string]any{
		"card_id": cur.CardID, "from": cur.State, "to": next.State, "version": next.Version, "actor": req.Actor,
	}
	if req.Decider != "" {
		payload["decider"] = req.Decider
	}
	if len(plan.evidence) != 0 {
		payload["evidence"] = plan.evidence
	}
	if plan.note != "" {
		payload["note"] = plan.note
	}
	if err := appendEvent(ctx, tx, cur.RunID, plan.kind, payload, now); err != nil {
		return Card{}, err
	}
	return next, nil
}

func updateCardRow(ctx context.Context, tx *sql.Tx, c Card, expected int64) error {
	res, err := tx.ExecContext(ctx, `UPDATE cards SET owner_label=?,state=?,version=?,evidence_path=?,updated_at=?,`+
		`stage=?,lease_holder=?,lease_expires_at=?,heartbeat_at=?,decision_gate=?,decision_question=?,decision_resume=?,`+
		`decider=?,decided_at=?,failure_reason=?,hint_prefer=?,hint_after=?,spec_id=?,worktree_path=?,evidence_sha=?,`+
		`merge_sha=?,merge_tree=?,remeasure_path=?,contract_spec_id=?,contract_sha256=?,contract_signed_at=?,contract_event=? `+
		`WHERE run_id=? AND card_id=? AND version=?`,
		c.OwnerLabel, c.State, c.Version, c.EvidencePath, c.UpdatedAt,
		c.Stage, c.LeaseHolder, c.LeaseExpiresAt, c.HeartbeatAt, c.DecisionGate, c.DecisionQuestion, c.DecisionResume,
		c.Decider, c.DecidedAt, c.FailureReason, c.HintPrefer, c.HintAfter, c.SpecID, c.WorktreePath, c.EvidenceSHA,
		c.MergeSHA, c.MergeTree, c.RemeasurePath, c.ContractSpecID, c.ContractSHA256, c.ContractSignedAt, c.ContractEvent,
		c.RunID, c.CardID, expected)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return fmt.Errorf("%w: card %s changed under the transition", ErrStaleVersion, c.CardID)
	}
	return nil
}

func appendEvent(ctx context.Context, tx *sql.Tx, runID, kind string, payload any, now time.Time) error {
	raw, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO events(run_id,kind,payload_json,created_at) VALUES(?,?,?,?)`, runID, kind, string(raw), now.Format(time.RFC3339Nano))
	return err
}
