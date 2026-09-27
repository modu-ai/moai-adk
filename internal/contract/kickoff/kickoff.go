// Package kickoff owns the contract-mode Kickoff: the plan→run gate check
// (`moai contract kickoff-check`) and the kickoff decision
// (`moai contract decide`) with its preconditions and outcome rules R1-R5.
package kickoff

import (
	"context"
	"errors"
	"time"

	"github.com/modu-ai/moai-adk/internal/contract"
	"github.com/modu-ai/moai-adk/internal/contract/receipt"
	"github.com/modu-ai/moai-adk/internal/jev"
)

// autonomousKickoffEnabled turns the non-human signature paths on. It stays
// false until every activation condition holds; while false, kickoff-check
// refuses llm and llm+jev signatures as autonomous-kickoff-inactive.
const autonomousKickoffEnabled = false

// JevDoctrineAmended is true once the Jev display-only principle is amended
// for the contract-mode Kickoff cross-check. While false, rule R3 routes a
// Jev-answered llm+jev decision to a human and the A1 signer refuses an
// effective llm+jev receipt. Both read this one constant.
const JevDoctrineAmended = false

// AutonomousKickoffEnabled reports the compiled activation state.
func AutonomousKickoffEnabled() bool { return autonomousKickoffEnabled }

// kickoff-check reason codes.
const (
	ReasonCardMismatch         = "card-mismatch"
	ReasonDeciderNotPermitted  = "decider-not-permitted"
	ReasonNotSignedValid       = "not-signed-valid"
	ReasonSignatureNotRecorded = "signature-not-recorded"
	ReasonReceiptNotIssued     = "receipt-not-issued"
	ReasonReceiptNotApproved   = "receipt-not-approved"
	ReasonDeciderMismatch      = "decider-mismatch"
	ReasonRevoked              = "revoked"
	ReasonInactive             = "autonomous-kickoff-inactive"
)

// Config is the effective configuration the gate and the decision read.
type Config struct {
	Mode             string
	Decider          string
	DeciderJevSole   bool
	JevEnabled       bool
	JevMinConfidence float64
}

// CheckInput is one kickoff-check.
type CheckInput struct {
	Root, SpecID, Card  string
	Config              Config
	Policy              contract.Policy
	RegistryRuleIDs     []string
	RegistryFrozenFiles []string
	SpecStatus          string
	Enabled             bool
	DoctrineAmended     bool
	Store               *receipt.Store
}

// CheckResult is the kickoff-check verdict.
type CheckResult struct {
	Pass                     bool     `json:"pass"`
	Reason                   string   `json:"reason"`
	Reasons                  []string `json:"reasons"`
	Mode                     string   `json:"mode"`
	Decider                  string   `json:"decider"`
	AutonomousKickoffEnabled bool     `json:"autonomous_kickoff_enabled"`
	JevDoctrineAmended       bool     `json:"jev_doctrine_amended"`
	VerifyState              string   `json:"verify_state"`
	VerifyReasons            []string `json:"verify_reasons"`
	SignerKind               string   `json:"signer_kind"`
}

var (
	ErrUsage             = errors.New("contract decide: usage")
	ErrCardMismatch      = errors.New("card-mismatch")
	ErrDeciderJevRefused = errors.New("decider-jev-refused")
	ErrIntegrity         = receipt.ErrIntegrity
)

// JevAsker is the Jev call seam.
type JevAsker func(ctx context.Context, req jev.Request) jev.Result

// DecideInput is one kickoff decision.
type DecideInput struct {
	Root, SpecID, Card  string
	Judgement           []byte
	Config              Config
	Policy              contract.Policy
	RegistryRuleIDs     []string
	RegistryFrozenFiles []string
	DoctrineAmended     bool
	Store               *receipt.Store
	Now                 func() time.Time
	Jev                 JevAsker
	CommitBodies        func(root, specRel string) ([]string, error)
}

// DecideResult is what Decide recorded.
type DecideResult struct {
	Outcome          string                 `json:"outcome"`
	Reason           string                 `json:"reason,omitempty"`
	Rule             string                 `json:"rule"`
	RequestedDecider string                 `json:"requested_decider"`
	EffectiveDecider string                 `json:"effective_decider,omitempty"`
	FallbackReason   string                 `json:"fallback_reason,omitempty"`
	ReceiptPath      string                 `json:"receipt_path,omitempty"`
	ReceiptSHA256    string                 `json:"receipt_sha256,omitempty"`
	Preconditions    []receipt.Precondition `json:"preconditions"`
}
