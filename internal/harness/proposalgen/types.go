// Package proposalgen consumes the harness learning log
// (.moai/harness/learning-history/tier-promotions.jsonl) and emits draft SPEC
// proposal candidates that close the V3R4 self-evolving harness learning loop.
//
// The package is a pure consumer of the canonical internal/harness.Promotion
// schema (internal/harness/types.go) and produces structured ProposalCandidate
// values plus optional on-disk scaffolding under .moai/proposals/<draft-id>/.
//
// HARD subagent boundary: this package and its CLI surface in
// internal/cli/harness/ MUST NOT invoke AskUserQuestion. The orchestrator owns
// the Approve/Modify/Reject gate (REQ-PGN-010/011) after consuming the JSON
// payload emitted by `moai harness propose`.
//
// SPEC: SPEC-V3R6-HARNESS-PROPOSAL-GEN-001 (REQ-PGN-001..014).
package proposalgen

import (
	"time"
)

// GeneratorVersion is the schema version embedded in proposal.json artifacts
// and CLI JSON output. Bump when the proposal directory layout or stdout JSON
// shape changes in an incompatible way.
const GeneratorVersion = "0.1.0"

// ConfidenceThreshold is the minimum Promotion.Confidence required for a
// promotion to advance to candidate mapping. Per REQ-PGN-004.
const ConfidenceThreshold = 0.70

// ProposalCandidate is a draft SPEC proposal derived from a single Promotion
// record. Each candidate carries the originating pattern_key plus the metadata
// needed for the scaffolder to render spec.md + proposal.json files.
type ProposalCandidate struct {
	// PatternKey is the originating Promotion.PatternKey (e.g.,
	// "code_change:func_extract:auth_module").
	PatternKey string `json:"pattern_key"`

	// ObservationCount is the cumulative observation count at the time the
	// originating Promotion was recorded.
	ObservationCount int `json:"observation_count"`

	// Confidence is the originating Promotion.Confidence (0.0..1.0).
	Confidence float64 `json:"confidence"`

	// Tier is the originating Promotion.ToTier
	// (always "recommendation" or "approval_required" after mapping).
	Tier string `json:"tier"`

	// SourceTs is the originating Promotion.Ts timestamp (UTC).
	SourceTs time.Time `json:"source_ts"`

	// DraftID is the derived draft identifier
	// (PROPOSAL-<YYYYMMDD>-<sha256(pattern_key)[:8]>).
	DraftID string `json:"draft_id"`
}

// GeneratorResult is the structured outcome of a generator run, suitable for
// emission as JSON to CLI stdout. REQ-PGN-009 defines the exact field set.
type GeneratorResult struct {
	// Proposals is the list of candidates that survived mapping + limit.
	// Empty slice on no-op path (not nil — JSON encodes as `[]`).
	Proposals []ProposalCandidate `json:"proposals"`

	// Reason is a short machine-readable diagnostic explaining the result
	// state (e.g., "no-actionable-patterns", "tier-promotions.jsonl absent
	// or empty", "ok").
	Reason string `json:"reason"`

	// MalformedLines is the count of JSONL lines that failed to unmarshal
	// during reader pass. Accumulated per REQ-PGN-002.
	MalformedLines int `json:"malformed_lines"`

	// EvaluatedPatterns is the number of unique pattern_key values observed
	// across the input promotions. Surfaces the mapper's evaluation surface
	// to the orchestrator. Per REQ-PGN-005.
	EvaluatedPatterns int `json:"evaluated_patterns"`

	// AutoDelegate signals the orchestrator AskUserQuestion gate handoff
	// (REQ-PGN-010). True only when --auto flag is set AND Proposals is
	// non-empty; otherwise false.
	AutoDelegate bool `json:"auto_delegate"`
}

// OutputFlags captures the CLI flag set for the `moai harness propose`
// subcommand. Decoupled from cobra to keep the run loop testable.
type OutputFlags struct {
	// InputPath overrides the canonical tier-promotions.jsonl path. Empty
	// string means use the default
	// (.moai/harness/learning-history/tier-promotions.jsonl).
	InputPath string

	// OutputDir is the root proposal directory. Empty string means use the
	// default (.moai/proposals).
	OutputDir string

	// Auto enables the auto_delegate signal to the orchestrator.
	Auto bool

	// DryRun suppresses all filesystem writes; the CLI still emits the
	// structured JSON result to stdout.
	DryRun bool

	// Limit caps the number of proposals returned. Zero or negative is
	// treated as the default (5).
	Limit int
}

// DefaultInputPath is the canonical tier-promotions.jsonl path relative to
// project root. Mirrors `.moai/harness/learning-history/tier-promotions.jsonl`.
const DefaultInputPath = ".moai/harness/learning-history/tier-promotions.jsonl"

// DefaultOutputDir is the canonical proposal directory relative to project
// root. Mirrors `.moai/proposals`.
const DefaultOutputDir = ".moai/proposals"

// DefaultLimit is the default --limit value when not set or set to <=0.
const DefaultLimit = 5
