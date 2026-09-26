package homestate

import (
	"context"
	"database/sql"
	"errors"
	"slices"
	"strings"
	"time"

	"github.com/modu-ai/moai-adk/internal/config"
)

// Card states of the F1 factory state machine. The set is closed: a row whose
// state lies outside it was written before the schema-4 migration and is a
// legacy row (see Card.Legacy).
const (
	CardPicked        = "picked"
	CardAssigned      = "assigned"
	CardLeased        = "leased"
	CardPlan          = "plan"
	CardPlanAudit     = "plan-audit"
	CardKickoff       = "kickoff"
	CardRun           = "run"
	CardSync          = "sync"
	CardSyncAudit     = "sync-audit"
	CardMergeReady    = "merge-ready"
	CardMerging       = "merging"
	CardMergedLocal   = "merged-local"
	CardPushed        = "pushed"
	CardCIGreen       = "ci-green"
	CardDone          = "done"
	CardNeedsDecision = "needs-decision"
	CardBlocked       = "blocked"
	CardFailed        = "failed"
	CardAbandoned     = "abandoned"
)

// Decision gates recorded while a card waits on an operator.
const (
	DecisionGateKickoff  = "kickoff"
	DecisionGateQuestion = "question"
)

// DeciderHuman is the only decider F1 accepts.
const DeciderHuman = "human"

var cardStates = []string{
	CardPicked, CardAssigned, CardLeased, CardPlan, CardPlanAudit, CardKickoff,
	CardRun, CardSync, CardSyncAudit, CardMergeReady, CardMerging, CardMergedLocal,
	CardPushed, CardCIGreen, CardDone, CardNeedsDecision, CardBlocked, CardFailed,
	CardAbandoned,
}

var leaseHoldingStates = []string{
	CardLeased, CardPlan, CardPlanAudit, CardRun, CardSync, CardSyncAudit, CardMergeReady, CardMerging,
}

// resumableStages are the working states a card can re-enter from `leased`
// through the stage-guarded T4 edges; `stage` records the last one reached.
var resumableStages = []string{CardPlan, CardPlanAudit, CardRun, CardSync, CardSyncAudit, CardMergeReady}

var terminalStates = []string{CardDone, CardFailed, CardAbandoned}

// CardStates returns the 19 F1 states in pipeline order.
func CardStates() []string { return slices.Clone(cardStates) }

// IsCardState reports whether s is one of the 19 F1 states.
func IsCardState(s string) bool { return slices.Contains(cardStates, s) }

// IsLeaseHoldingState reports whether a card in state s holds a worker lease.
func IsLeaseHoldingState(s string) bool { return slices.Contains(leaseHoldingStates, s) }

// IsTerminalCardState reports whether s admits no further transition.
func IsTerminalCardState(s string) bool { return slices.Contains(terminalStates, s) }

func isResumableStage(s string) bool { return slices.Contains(resumableStages, s) }

// FactoryLeaseDuration is how long a card lease lasts past its last
// heartbeat (config.DefaultFactoryLeaseDuration; no YAML key in F1).
var FactoryLeaseDuration = config.DefaultFactoryLeaseDuration

// ErrCardNotFound reports that no card record exists for the run and card id.
var ErrCardNotFound = errors.New("factory card not found")

// Card is one row of the factory `cards` table (schema version 4).
type Card struct {
	RunID            string `json:"run_id"`
	CardID           string `json:"card_id"`
	OwnerLabel       string `json:"owner"`
	State            string `json:"state"`
	Version          int64  `json:"version"`
	EvidencePath     string `json:"evidence_path"`
	UpdatedAt        string `json:"updated_at"`
	Stage            string `json:"stage"`
	LeaseHolder      string `json:"lease_holder"`
	LeaseExpiresAt   string `json:"lease_expires_at"`
	HeartbeatAt      string `json:"heartbeat_at"`
	DecisionGate     string `json:"decision_gate"`
	DecisionQuestion string `json:"decision_question"`
	DecisionResume   string `json:"decision_resume"`
	Decider          string `json:"decider"`
	DecidedAt        string `json:"decided_at"`
	FailureReason    string `json:"failure_reason"`
	HintPrefer       string `json:"prefer"`
	HintAfter        string `json:"after"`
	SpecID           string `json:"spec_id"`
	WorktreePath     string `json:"worktree_path"`
	EvidenceSHA      string `json:"evidence_sha"`
	MergeSHA         string `json:"merge_sha"`
	MergeTree        string `json:"merge_tree"`
	RemeasurePath    string `json:"remeasure_path"`
	ContractSpecID   string `json:"contract_spec_id"`
	ContractSHA256   string `json:"contract_sha256"`
	ContractSignedAt string `json:"contract_signed_at"`
	ContractEvent    string `json:"contract_event"`
}

// Legacy reports whether the row holds a state outside the F1 set (a row
// written before the schema-4 migration).
func (c Card) Legacy() bool { return !IsCardState(c.State) }

// LeaseExpired reports whether the card holds a lease whose expiry is at or
// before now. A card without a lease, or in a state that holds none, never has
// an expired lease.
func (c Card) LeaseExpired(now time.Time) bool {
	if !IsLeaseHoldingState(c.State) || c.LeaseExpiresAt == "" {
		return false
	}
	at, err := time.Parse(time.RFC3339Nano, c.LeaseExpiresAt)
	if err != nil {
		// An unparseable expiry cannot be trusted as a live lease.
		return true
	}
	return !now.Before(at)
}

const cardSelectColumns = `run_id,card_id,owner_label,state,version,evidence_path,updated_at,` +
	`stage,lease_holder,lease_expires_at,heartbeat_at,decision_gate,decision_question,decision_resume,` +
	`decider,decided_at,failure_reason,hint_prefer,hint_after,spec_id,worktree_path,evidence_sha,` +
	`merge_sha,merge_tree,remeasure_path,contract_spec_id,contract_sha256,contract_signed_at,contract_event`

type rowScanner interface{ Scan(dest ...any) error }

func scanCard(row rowScanner) (Card, error) {
	var c Card
	err := row.Scan(&c.RunID, &c.CardID, &c.OwnerLabel, &c.State, &c.Version, &c.EvidencePath, &c.UpdatedAt,
		&c.Stage, &c.LeaseHolder, &c.LeaseExpiresAt, &c.HeartbeatAt, &c.DecisionGate, &c.DecisionQuestion, &c.DecisionResume,
		&c.Decider, &c.DecidedAt, &c.FailureReason, &c.HintPrefer, &c.HintAfter, &c.SpecID, &c.WorktreePath, &c.EvidenceSHA,
		&c.MergeSHA, &c.MergeTree, &c.RemeasurePath, &c.ContractSpecID, &c.ContractSHA256, &c.ContractSignedAt, &c.ContractEvent)
	return c, err
}

type queryRower interface {
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

func loadCard(ctx context.Context, q queryRower, runID, cardID string) (Card, error) {
	// SQL: the concatenated fragment is a compile-time constant; every value goes through a ? placeholder.
	c, err := scanCard(q.QueryRowContext(ctx, `SELECT `+cardSelectColumns+` FROM cards WHERE run_id=? AND card_id=?`, runID, cardID))
	if errors.Is(err, sql.ErrNoRows) {
		return Card{}, ErrCardNotFound
	}
	return c, err
}

// LoadCard reads one card record. It never writes.
func (f *FactoryDB) LoadCard(ctx context.Context, runID, cardID string) (Card, error) {
	return loadCard(ctx, f.DB, runID, cardID)
}

// ListCards reads every card record of runID, or of every run when runID is
// empty, ordered by run and card id. It never writes.
func (f *FactoryDB) ListCards(ctx context.Context, runID string) (_ []Card, err error) {
	// SQL: the concatenated fragment is a compile-time constant; every value goes through a ? placeholder.
	query := `SELECT ` + cardSelectColumns + ` FROM cards`
	var args []any
	if strings.TrimSpace(runID) != "" {
		query += ` WHERE run_id=?`
		args = append(args, runID)
	}
	rows, err := f.DB.QueryContext(ctx, query+` ORDER BY run_id, card_id`, args...)
	if err != nil {
		return nil, err
	}
	defer func() {
		if cerr := rows.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()
	var out []Card
	for rows.Next() {
		c, err := scanCard(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}
