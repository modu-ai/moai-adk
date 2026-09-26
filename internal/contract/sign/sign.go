package sign

import (
	"errors"
	"io"
	"os"
	"time"

	"github.com/modu-ai/moai-adk/internal/contract"
)

// KickoffNotice is the REQ-CONTRACT-019 notice printed after every successful
// signature, on both signing paths and in both modes.
const KickoffNotice = "Notice: this signature does not yet replace Implementation Kickoff Approval; the Kickoff gate stays in force in every mode."

// Signer values accepted by Options.Signer.
const (
	SignerHuman   = contract.SignerHuman
	SignerLLM     = contract.DeciderLLM
	SignerLLMJev  = contract.DeciderLLMJev
	ModeContract  = "contract"
	deciderJevRaw = "jev"
)

var (
	// ErrUsage wraps every invocation error (the CLI maps it to exit 2).
	ErrUsage = errors.New("contract sign: usage")
	// ErrInternal wraps a produced contract that failed its own pre-write
	// verification; nothing is written.
	ErrInternal = errors.New("contract sign: internal error")
)

// Options is one signing invocation. The policy fields are the caller's
// effective workflow.autonomy values; this package never reads configuration.
type Options struct {
	// ProjectRoot is the project directory containing .moai/specs/.
	ProjectRoot string
	// SpecIDs are the SPECs to sign; repeats are ignored.
	SpecIDs []string
	// Signer is "" or "human" (human path), or "llm" / "llm+jev" (receipt
	// path; it names the receipt's effective decider).
	Signer string
	// ReceiptPath is the --receipt value (receipt path only). An absolute
	// path is taken relative to ProjectRoot.
	ReceiptPath string
	// Resign re-signs a signed contract (REQ-CONTRACT-022).
	Resign bool

	Mode         string // workflow.autonomy.mode (effective)
	BatchSign    bool   // workflow.autonomy.contract.batch_sign
	SecondReview string // workflow.autonomy.contract.second_review
	PushDevelop  bool   // workflow.autonomy.contract.push_develop
	// Decider is the effective workflow.autonomy.kickoff.decider.
	Decider string
	// DeciderJevSole is true when the configured decider is the rejected
	// "jev" (a configuration error). A Decider of "jev" is treated the same.
	DeciderJevSole bool
	// JevEnabled and JevMinConfidence are carried for the receipt issuer; the
	// A1 receipt validator does not evaluate them.
	JevEnabled       bool
	JevMinConfidence float64
	// BudgetDefault fills an absent contract budget.
	BudgetDefault contract.Budget

	// RegistryRuleIDs and RegistryFrozenFiles are the constitution registry
	// values verify needs.
	RegistryRuleIDs     []string
	RegistryFrozenFiles []string
	// AgentMarkers is the closed agent-environment marker set (variable
	// names) checked on the human path. It must be non-empty there.
	AgentMarkers []string
}

// Seams are the side-effecting operations Sign performs. A nil field takes
// its default.
type Seams struct {
	// IsTTY reports whether standard input is a terminal. Default: stdin is
	// a character device.
	IsTTY func() bool
	// Getenv reads an environment variable. Default: os.Getenv.
	Getenv func(string) string
	// ReadLine reads the one confirmation line. Default: one line of stdin.
	ReadLine func() (string, error)
	// Out receives the summary, prompt, refusals, and the notice. Default:
	// io.Discard.
	Out io.Writer
	// Now is the signing clock. Default: time.Now.
	Now func() time.Time
	// GitIdentity returns git user.name and user.email in root; an unset
	// value is "" (not an error). Default: `git config` subprocesses.
	GitIdentity func(root string) (name, email string, err error)
	// GitHead returns HEAD in root. Default: `git rev-parse HEAD`.
	GitHead func(root string) (string, error)
	// NewBatchID returns a fresh non-empty batch identifier. Default:
	// 16 random bytes, hex.
	NewBatchID func() string
	// WriteFile replaces path atomically. Default: temp file in the same
	// directory, then internal/atomicfile.Replace.
	WriteFile func(path string, data []byte, perm os.FileMode) error
}

// Result reports what Sign did.
type Result struct {
	// Refusal is a sign refusal code (contract.SignRefusalCodes), or "" when
	// every contract was signed.
	Refusal string
	// SpecID is the SPEC a refusal concerns ("" when it concerns the whole
	// invocation).
	SpecID string
	// Cause is the one-line cause printed with the refusal.
	Cause string
	// Reasons are the verify reason codes when Refusal is verify_failed.
	Reasons []string
	// Written and Unwritten are the contract.yaml paths (slash,
	// project-relative) written and not written. On success Written holds
	// every contract; on a write failure they are the exact split.
	Written   []string
	Unwritten []string
}

// Sign signs the contracts named in opts (design.md § Signing Flow).
func Sign(opts Options, seams Seams) (Result, error) {
	return Result{}, errors.New("contract sign: not implemented")
}
