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
			result, err = commitTransition(ctx, tx, cur, transitionPlan{next: next, kind: "card.transition"}, req, now)
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
	if err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	if after != nil {
		after()
	}
	return nil
}

// planTransition checks the edge's guard and computes the new row.
func (f *FactoryDB) planTransition(ctx context.Context, tx *sql.Tx, cur Card, edge transitionEdge, req TransitionRequest, now time.Time) (transitionPlan, error) {
	next := cur
	next.State = req.To
	plan := transitionPlan{next: next, kind: "card.transition", evidence: map[string]string{}}
	nowText := now.Format(time.RFC3339Nano)
	switch edge.guard {
	case guardAssign:
		owner := strings.TrimSpace(req.Owner)
		if owner == "" {
			return plan, fmt.Errorf("%w: assigning a card requires an owner label", ErrInvalidCardInput)
		}
		plan.next.OwnerLabel = owner
	case guardLeaseAcquire:
	case guardStageResume:
		if !stageAdmits(cur.Stage, req.To) {
			return plan, fmt.Errorf("%w: leased card stage %q does not resume into %s", ErrIllegalTransition, cur.Stage, req.To)
		}
	case guardEntry:
		sha, path := strings.TrimSpace(req.SHA), strings.TrimSpace(req.ArtifactPath)
		if sha == "" || path == "" {
			return plan, fmt.Errorf("%w: audit entry requires a commit SHA and an artifact path", ErrEvidence)
		}
		plan.next.EvidenceSHA = sha
		plan.next.EvidencePath = path
		plan.evidence["sha"], plan.evidence["artifact_path"] = sha, path
	case guardVerdictAny, guardVerdictPass:
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
	case guardCommit:
		sha := strings.TrimSpace(req.SHA)
		if sha == "" {
			return plan, fmt.Errorf("%w: run → sync requires a commit SHA", ErrEvidence)
		}
		plan.next.EvidenceSHA = sha
		plan.evidence["sha"] = sha
	case guardLeaseValid, guardNone:
	case guardMerge:
		plan.next.MergeSHA = strings.TrimSpace(req.MergeSHA)
		plan.next.RemeasurePath = strings.TrimSpace(req.RemeasurePath)
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
		} else if !remote {
			return plan, fmt.Errorf("%w: merged-local → pushed requires a configured remote", ErrIllegalTransition)
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
		plan.next.FailureReason = strings.TrimSpace(req.Reason)
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
