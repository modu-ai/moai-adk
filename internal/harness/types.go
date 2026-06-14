// Package harness — JSONL log schema type definitions.
// REQ-HL-001: Define event schema v1.
// REQ-HL-002: Define Pattern, Tier, Promotion types.
// REQ-HL-005~008: Define Phase 3 Safety Architecture types.
package harness

import "time"

// LogSchemaVersion is the schema version of JSONL log files.
// Increment this version for future schema changes to track backward compatibility.
//
// Bumped "v1" → "v2" by SPEC-HARNESS-OUTCOME-CAPTURE-001 (REQ-OC-010) to mark the
// additive apply_outcome event type + its omitempty Event fields. The bump is a
// marker only — existing readers tolerate the additive omitempty fields and the
// new event type without change (REQ-OC-009).
const LogSchemaVersion = "v2"

// EventType is an enum representing observable event kinds.
// REQ-HL-001: Records /moai subcommand, agent invocation, SPEC-ID reference, /moai feedback events.
type EventType string

const (
	// EventTypeMoaiSubcommand represents /moai subcommand invocation.
	EventTypeMoaiSubcommand EventType = "moai_subcommand"

	// EventTypeAgentInvocation represents agent (subagent) invocation.
	EventTypeAgentInvocation EventType = "agent_invocation"

	// EventTypeSpecReference represents SPEC-ID reference (mention).
	EventTypeSpecReference EventType = "spec_reference"

	// EventTypeFeedback represents /moai feedback event.
	EventTypeFeedback EventType = "feedback"

	// @MX:NOTE: [AUTO] Three new event kinds for session termination (Stop hook),
	// subagent termination (SubagentStop hook), and user prompt submission (UserPromptSubmit hook).
	// REQ-HRN-OBS-015: uses SEMANTIC values (not the runtime hook names).
	// Round C will add more in the future: REQ-HRN-OBS-016 (SPEC-V3R4-HARNESS-003).

	// EventTypeSessionStop represents Claude Code Stop hook (session termination).
	EventTypeSessionStop EventType = "session_stop"

	// EventTypeSubagentStop represents Claude Code SubagentStop hook (subagent termination).
	EventTypeSubagentStop EventType = "subagent_stop"

	// EventTypeUserPrompt represents Claude Code UserPromptSubmit hook (user prompt submitted).
	EventTypeUserPrompt EventType = "user_prompt"

	// EventTypeApplyOutcome represents a completed harness Apply outcome record
	// (SPEC-HARNESS-OUTCOME-CAPTURE-001 REQ-OC-004). It is additive — it does not
	// redefine, reorder, or remove any existing EventType constant (C8). The event
	// carries the regression-gate verdict + project-health delta + proposal_id
	// correlation key in the additive omitempty Outcome* fields below.
	EventTypeApplyOutcome EventType = "apply_outcome"
)

// Event is the single JSONL line schema for usage-log.jsonl file.
// REQ-HL-001: Includes timestamp, event_type, subject, context_hash, tier_increment fields.
// REQ-HRN-FND-010: preserves the existing 4-field schema. New fields are additive omitempty only.
//
// @MX:NOTE: [AUTO] 12 new fields added as omitempty. Existing entries remain compatible.
// REQ-HRN-OBS-009: preserves the existing 4-field JSONL schema — new fields are additive omitempty only.
// @MX:TODO: [AUTO] Plan to add gate with learning.enabled setting key in Phase 4.
// @MX:SPEC: SPEC-V3R3-HARNESS-LEARNING-001 REQ-HL-001
type Event struct {
	// Timestamp is the event occurrence time (UTC).
	Timestamp time.Time `json:"timestamp"`

	// EventType is the event kind (EventType enum).
	EventType EventType `json:"event_type"`

	// Subject is the event subject (e.g., "/moai plan", "expert-backend", "SPEC-001").
	Subject string `json:"subject"`

	// ContextHash is the session context identifier (hash for collision prevention).
	ContextHash string `json:"context_hash"`

	// TierIncrement is the tier count increment caused by this event (default 0).
	TierIncrement int `json:"tier_increment"`

	// SchemaVersion is the log schema version (LogSchemaVersion).
	SchemaVersion string `json:"schema_version"`

	// ── Stop event optional fields (REQ-HRN-OBS-003) ─────────────────────

	// SessionID is the Claude Code session identifier (Stop event only, omitempty).
	SessionID string `json:"session_id,omitempty"`

	// LastAssistantMessageHash is the SHA-256 hash of the last assistant message
	// (Stop event only, omitempty).
	LastAssistantMessageHash string `json:"last_assistant_message_hash,omitempty"`

	// LastAssistantMessageLen is the character count of the last assistant message
	// (Stop event only, omitempty).
	LastAssistantMessageLen int `json:"last_assistant_message_len,omitempty"`

	// ── SubagentStop event optional fields (REQ-HRN-OBS-005) ─────────────

	// AgentName is the subagent name (SubagentStop event only, omitempty).
	AgentName string `json:"agent_name,omitempty"`

	// AgentType is the subagent type (SubagentStop event only, omitempty).
	AgentType string `json:"agent_type,omitempty"`

	// AgentID is the subagent instance identifier (SubagentStop event only, omitempty).
	AgentID string `json:"agent_id,omitempty"`

	// ParentSessionID is the parent session identifier (SubagentStop event only, omitempty).
	ParentSessionID string `json:"parent_session_id,omitempty"`

	// ── UserPromptSubmit event optional fields (REQ-HRN-OBS-007) ─────────

	// PromptHash is the SHA-256 hash of the user prompt (UserPromptSubmit only, omitempty).
	PromptHash string `json:"prompt_hash,omitempty"`

	// PromptLen is the character count of the user prompt (UserPromptSubmit only, omitempty).
	PromptLen int `json:"prompt_len,omitempty"`

	// PromptLang is the detected language code of the user prompt
	// (UserPromptSubmit only, omitempty). Examples: "ko", "en", "ja", "zh", "".
	PromptLang string `json:"prompt_lang,omitempty"`

	// PromptPreview is the first 64 bytes of the user prompt (UTF-8 boundary-safe truncation).
	// (UserPromptSubmit only, opt-in Strategy B, omitempty).
	// REQ-HRN-OBS-013 / AC-HRN-OBS-008.a.
	PromptPreview string `json:"prompt_preview,omitempty"`

	// PromptContent is the full user prompt text
	// (UserPromptSubmit only, opt-in Strategy C/Full, omitempty).
	// AC-HRN-OBS-008.b: JSON field name MUST be `prompt_content` (not `prompt_full`).
	PromptContent string `json:"prompt_content,omitempty"`

	// ── ApplyOutcome event optional fields (SPEC-HARNESS-OUTCOME-CAPTURE-001) ──
	// These fields are additive omitempty only and are set ONLY on apply_outcome
	// events (REQ-OC-004, C9). They carry the regression-gate verdict, the
	// project-health delta (baseline + candidate triples + regressed dimensions),
	// and the proposal_id correlation key reused from the lineage record (DD-3).
	// The verdict (OutcomeVerdict) is the load-bearing signal — the int/float64
	// triple fields omitempty-drop genuine zeros (a kept Δ=0 outcome is still
	// identifiable by OutcomeVerdict == "kept" with the triple fields absent).

	// OutcomeVerdict is the Apply verdict: "kept" | "rolled-back" (apply_outcome only).
	OutcomeVerdict string `json:"outcome_verdict,omitempty"`

	// OutcomeDecision is the transition decision: "approved" | "regression-blocked"
	// (apply_outcome only).
	OutcomeDecision string `json:"outcome_decision,omitempty"`

	// OutcomeProposalID is the lineage correlation key reused from the lineage
	// record (Proposal.ID — DD-3) (apply_outcome only).
	OutcomeProposalID string `json:"outcome_proposal_id,omitempty"`

	// OutcomeBaseTests / OutcomeBaseCoverage / OutcomeBaseLint are the baseline
	// MetricTriple measured BEFORE the apply (apply_outcome only, omitempty).
	OutcomeBaseTests    int     `json:"outcome_baseline_tests,omitempty"`
	OutcomeBaseCoverage float64 `json:"outcome_baseline_coverage,omitempty"`
	OutcomeBaseLint     int     `json:"outcome_baseline_lint,omitempty"`

	// OutcomeCandTests / OutcomeCandCoverage / OutcomeCandLint are the candidate
	// MetricTriple measured AFTER the apply (apply_outcome only, omitempty).
	OutcomeCandTests    int     `json:"outcome_candidate_tests,omitempty"`
	OutcomeCandCoverage float64 `json:"outcome_candidate_coverage,omitempty"`
	OutcomeCandLint     int     `json:"outcome_candidate_lint,omitempty"`

	// OutcomeRegressed is the list of regressed dimensions (e.g. ["coverage"])
	// (apply_outcome rolled-back case only, omitempty).
	OutcomeRegressed []string `json:"outcome_regressed,omitempty"`
}

// ─────────────────────────────────────────────
// Phase 2: Pattern Classification Types (REQ-HL-002)
// ─────────────────────────────────────────────

// Tier is an enum representing the maturity stage of a pattern.
// REQ-HL-002: Classified based on cumulative observation counts {1,3,5,10}.
//
// @MX:ANCHOR: [AUTO] Used by multiple callers as ClassifyTier return value.
// @MX:REASON: [AUTO] fan_in >= 3: learner.go, learner_test.go, applier.go
type Tier int

const (
	// TierObservation is a pattern observed 1+ times (lowest tier).
	TierObservation Tier = iota + 1

	// TierHeuristic is a pattern observed 3+ times; subject to description enrichment.
	TierHeuristic

	// TierRule is a pattern observed 5+ times; subject to trigger injection.
	TierRule

	// TierAutoUpdate is a pattern observed 10+ times; auto-update candidate.
	TierAutoUpdate
)

// String returns Tier as a human-readable string.
func (t Tier) String() string {
	switch t {
	case TierObservation:
		return "observation"
	case TierHeuristic:
		return "heuristic"
	case TierRule:
		return "rule"
	case TierAutoUpdate:
		return "auto_update"
	default:
		return "unknown"
	}
}

// Pattern is a single pattern identified by (event_type, subject, context_hash) combination.
// REQ-HL-002: Learner reads JSONL logs and aggregates into patterns.
type Pattern struct {
	// Key is the unique identifier in "event_type:subject:context_hash" format.
	Key string

	// EventType is the event kind of the pattern.
	EventType EventType

	// Subject is the pattern subject (e.g., "/moai plan", "expert-backend").
	Subject string

	// ContextHash is the context identifier.
	ContextHash string

	// Count is the total number of times this pattern was observed.
	Count int

	// Confidence is the pattern confidence score (0.0 ~ 1.0).
	// If below 0.70, classified as TierObservation regardless of Count.
	Confidence float64

	// Tier is the currently classified tier (set after ClassifyTier call).
	Tier Tier
}

// Promotion is the tier promotion event recorded in tier-promotions.jsonl.
// Based on plan.md §4.2 schema.
type Promotion struct {
	// Ts is the promotion timestamp (UTC RFC3339).
	Ts time.Time `json:"ts"`

	// PatternKey is in "event_type:subject" format (per plan.md §4.2).
	PatternKey string `json:"pattern_key"`

	// FromTier is the tier string before promotion.
	FromTier string `json:"from_tier"`

	// ToTier is the promoted tier string.
	ToTier string `json:"to_tier"`

	// ObservationCount is the observation count at promotion time.
	ObservationCount int `json:"observation_count"`

	// Confidence is the pattern confidence score.
	Confidence float64 `json:"confidence"`
}

// ─────────────────────────────────────────────
// Phase 3: 5-Layer Safety Architecture Types (REQ-HL-005~008)
// ─────────────────────────────────────────────

// Proposal is the automatic update proposal payload.
// REQ-HL-005: System generates when Tier 4 is reached and passes through safety pipeline.
//
// @MX:ANCHOR: [AUTO] Proposal is the input type for the entire safety pipeline.
// @MX:REASON: [AUTO] fan_in >= 3: safety/pipeline.go, safety/frozen_guard.go, safety/canary.go
type Proposal struct {
	// ID is the proposal unique identifier.
	ID string `json:"id"`

	// TargetPath is the target file path to modify (e.g., ".claude/skills/my-harness-x/SKILL.md").
	TargetPath string `json:"target_path"`

	// FieldKey is the target frontmatter field to modify (e.g., "description", "triggers").
	FieldKey string `json:"field_key"`

	// NewValue is the new value to apply.
	NewValue string `json:"new_value"`

	// PatternKey is the pattern identifier that generated this proposal.
	PatternKey string `json:"pattern_key"`

	// Tier is the tier of the pattern that generated this proposal.
	Tier Tier `json:"tier"`

	// ObservationCount is the pattern observation count.
	ObservationCount int `json:"observation_count"`

	// CreatedAt is the proposal creation timestamp.
	CreatedAt time.Time `json:"created_at"`
}

// DecisionKind is the final decision kind of the safety pipeline.
type DecisionKind string

const (
	// DecisionApproved is the case when all safety layers passed.
	DecisionApproved DecisionKind = "approved"

	// DecisionRejected is the case when rejected by one or more safety layers.
	DecisionRejected DecisionKind = "rejected"

	// DecisionPendingApproval is the case when pending L5 Human Oversight approval.
	DecisionPendingApproval DecisionKind = "pending_approval"
)

// Decision is the safety pipeline execution result.
// REQ-HL-005~008: Summarizes L1→L2→L3→L4→L5 evaluation results.
type Decision struct {
	// Kind is the decision result kind.
	Kind DecisionKind `json:"kind"`

	// RejectedBy is the layer number that caused rejection (1~5, 0 if passed).
	RejectedBy int `json:"rejected_by,omitempty"`

	// Reason is the rejection or pending reason.
	Reason string `json:"reason,omitempty"`

	// OversightProposal is the payload to return to orchestrator when pending L5 approval.
	OversightProposal *OversightProposal `json:"oversight_proposal,omitempty"`
}

// Session is a single session metric for effectiveness calculation.
// REQ-HL-007: Used in canary check for baseline vs proposal comparison.
type Session struct {
	// ID is the session identifier.
	ID string `json:"id"`

	// SubcommandSuccessRate is the /moai subcommand success rate (0.0~1.0).
	SubcommandSuccessRate float64 `json:"subcommand_success_rate"`

	// AgentInvocationSuccess is the agent invocation success rate (0.0~1.0).
	AgentInvocationSuccess float64 `json:"agent_invocation_success"`

	// CompletionRate is the SPEC completion rate (0.0~1.0).
	CompletionRate float64 `json:"completion_rate"`

	// Timestamp is the session start timestamp.
	Timestamp time.Time `json:"timestamp"`
}

// CanaryResult is the canary check execution result.
// REQ-HL-007: rejected=true if effectiveness drop exceeds 0.10 compared to baseline.
type CanaryResult struct {
	// BaselineScore is the baseline effectiveness score before proposal application.
	BaselineScore float64 `json:"baseline_score"`

	// ProjectedScore is the predicted effectiveness score after proposal application.
	ProjectedScore float64 `json:"projected_score"`

	// Delta is ProjectedScore - BaselineScore (negative means drop).
	Delta float64 `json:"delta"`

	// Rejected is true if delta exceeded threshold (-0.10) and was rejected.
	Rejected bool `json:"rejected"`

	// Reason is the rejection reason (set only when Rejected=true).
	Reason string `json:"reason,omitempty"`
}

// ContradictionType is the kind of detected contradiction.
type ContradictionType string

const (
	// ContradictionOverlappingTriggers occurs when multiple skills have overlapping trigger keywords.
	ContradictionOverlappingTriggers ContradictionType = "overlapping_triggers"

	// ContradictionChainRules is a contradiction in chaining-rules.yaml (same phase, different rules).
	ContradictionChainRules ContradictionType = "contradictory_chain_rules"
)

// ContradictionItem is a single contradiction item.
type ContradictionItem struct {
	// Type is the contradiction kind.
	Type ContradictionType `json:"type"`

	// Description is the detailed contradiction description.
	Description string `json:"description"`

	// ConflictingPaths is the list of file paths where contradiction occurred.
	ConflictingPaths []string `json:"conflicting_paths"`

	// ConflictingValues is the value pair that caused the contradiction.
	ConflictingValues []string `json:"conflicting_values"`
}

// ContradictionReport is the detection result of contradiction detector.
// REQ-HL-008: Empty Items means no contradiction, return to orchestrator if Items exist.
type ContradictionReport struct {
	// Items is the list of detected contradictions.
	Items []ContradictionItem `json:"items"`
}

// HasContradiction returns true if any contradiction exists.
func (r ContradictionReport) HasContradiction() bool {
	return len(r.Items) > 0
}

// OversightOption is a single AskUserQuestion option.
// REQ-HL-005: max 4 options, first option is recommended.
type OversightOption struct {
	// Label is the option label (e.g., "Approve (Recommended)", "Reject").
	Label string `json:"label"`

	// Description is the option detailed description.
	Description string `json:"description"`

	// Value is the value for programming processing (e.g., "approve", "reject").
	Value string `json:"value"`

	// Recommended is true if this is the recommended option (set only on first option).
	Recommended bool `json:"recommended"`
}

// OversightProposal is the payload that L5 Human Oversight returns to orchestrator.
// REQ-HL-005: Subagents do not call AskUserQuestion directly, but return this struct.
// Orchestrator (MoAI) receives this payload and displays to users via AskUserQuestion.
//
// @MX:ANCHOR: [AUTO] OversightProposal is the subagent→orchestrator interface boundary.
// @MX:REASON: [AUTO] fan_in >= 3: safety/oversight.go, safety/pipeline.go, Phase 4 coordinator
type OversightProposal struct {
	// Question is the question text to display to the user.
	Question string `json:"question"`

	// Options is the list of options (max 4, first is recommended).
	Options []OversightOption `json:"options"`

	// ProposalID is the ID of the related Proposal.
	ProposalID string `json:"proposal_id"`

	// Context is additional context for the user to make a decision.
	Context string `json:"context"`
}

// ─────────────────────────────────────────────
// M6 Auditable Lineage Logging (SPEC-HARNESS-LOOP-CLOSURE-001)
// ─────────────────────────────────────────────

// LineageEntry는 manifest.jsonl에 append되는 단일 apply 전환 기록이다 (M6 auditable lineage).
// 매 apply 전환마다 정확히 하나의 entry가 기록된다 (accept 시 1개, reject 시 1개).
// rejected candidate도 기록되지만 active harness는 변경하지 않는다 (REQ-HLC-004).
// optional 필드는 omitempty를 사용해 기존 reader와 향후 schema 추가에 backward-compatible하다.
//
// @MX:ANCHOR: [AUTO] LineageEntry는 lineage writer/loader와 Apply 통합의 공유 record 타입.
// @MX:REASON: [AUTO] fan_in >= 3: lineage.go(WriteLineageEntry/LoadManifest), applier.go(Apply), lineage_test.go
type LineageEntry struct {
	// ProposalID는 이 전환을 발생시킨 proposal의 ID.
	ProposalID string `json:"proposal_id"`

	// TargetPath는 전환 대상 파일 경로.
	TargetPath string `json:"target_path"`

	// AppliedSurface는 accept 시 변경된 frontmatter 필드 키 (예: "description").
	// reject 시 비어 있다 (omitempty).
	AppliedSurface string `json:"applied_surface,omitempty"`

	// Decision은 전환 결정이다: "approved" | "rejected".
	Decision string `json:"decision"`

	// Timestamp는 전환 기록 시각 (UTC).
	Timestamp time.Time `json:"timestamp"`

	// Reason은 전환 사유 (reject 시 거부 layer의 사유, approve 시 승인 설명).
	Reason string `json:"reason,omitempty"`
}
