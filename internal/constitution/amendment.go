// Package constitution implements the FROZEN/EVOLVABLE zone model of the MoAI-ADK rule tree.
// amendment implements the constitutional amendment protocol through the 5-layer safety gate of SPEC-V3R2-CON-002.
package constitution

import (
	"fmt"
	"time"
)

// ErrFrozenAmendment represents an error for amendment attempts to Frozen zone rules.
// Returned when there is no evidence for Frozen→Evolvable demotion.
type ErrFrozenAmendment struct {
	RuleID string
	Reason string
}

func (e *ErrFrozenAmendment) Error() string {
	return fmt.Sprintf("Frozen zone amendment rejected: %s - %s", e.RuleID, e.Reason)
}

// ErrCanaryUnavailable is returned when there are insufficient SPECs to run Canary evaluation.
type ErrCanaryUnavailable struct {
	RequiredCount int
	ActualCount   int
}

func (e *ErrCanaryUnavailable) Error() string {
	return fmt.Sprintf("Canary unavailable: minimum %d SPECs required (current: %d)", e.RequiredCount, e.ActualCount)
}

// ErrCanaryRejected is returned when a score drop is detected during Canary evaluation.
type ErrCanaryRejected struct {
	RuleID         string
	ScoreDrop      float64
	Threshold      float64
	AffectedSpecs  []string
}

func (e *ErrCanaryRejected) Error() string {
	return fmt.Sprintf("Canary rejected: %s score drop %.2f > threshold %.2f (affected SPECs: %v)",
		e.RuleID, e.ScoreDrop, e.Threshold, e.AffectedSpecs)
}

// ErrContradictionDetected is returned when an amendment contradicts existing rules.
type ErrContradictionDetected struct {
	NewRuleID      string
	ConflictingIDs []string
	Conflicts      []string // Contradiction descriptions
}

func (e *ErrContradictionDetected) Error() string {
	return fmt.Sprintf("Contradiction detected: %s contradicts the following rules: %v\nDetails: %v",
		e.NewRuleID, e.ConflictingIDs, e.Conflicts)
}

// ErrRateLimitExceeded is returned when the amendment frequency limit is exceeded.
type ErrRateLimitExceeded struct {
	MaxPerWeek    int
	CooldownHours int
	NextAllowedAt time.Time
}

func (e *ErrRateLimitExceeded) Error() string {
	return fmt.Sprintf("Rate limit exceeded: maximum %d times/week, %d hour cooldown (next allowed: %s)",
		e.MaxPerWeek, e.CooldownHours, e.NextAllowedAt.Format(time.RFC3339))
}

// ErrAmendmentInProgress is returned when another amendment is already in progress.
// Violates single-writer lock.
type ErrAmendmentInProgress struct {
	LockFilePath string
}

func (e *ErrAmendmentInProgress) Error() string {
	return fmt.Sprintf("Amendment already in progress: lock file %s exists", e.LockFilePath)
}

// ErrRolledBack is returned when attempting to re-amend a rule that has already been rolled back.
type ErrRolledBack struct {
	RuleID       string
	RolledBackAt time.Time
	CooldownDays int
}

func (e *ErrRolledBack) Error() string {
	return fmt.Sprintf("Rolled back rule: %s was rolled back at %s. %d day cooldown in progress",
		e.RuleID, e.RolledBackAt.Format(time.RFC3339), e.CooldownDays)
}

// AmendmentProposal represents a constitutional amendment proposal.
// Direct implementation of SPEC-V3R2-CON-002 REQ-CON-002-001.
type AmendmentProposal struct {
	// RuleID is the ID of the rule to be modified (CONST-V3R2-NNN).
	RuleID string
	// Before is the current clause text.
	Before string
	// After is the new clause text.
	After string
	// Evidence is the amendment justification (required for Frozen zone).
	Evidence string
	// CanaryResult is the canary evaluation result.
	CanaryResult *CanaryResult
	// Contradicts is the contradiction detection result.
	Contradicts *ContradictionResult
	// Approved indicates whether it has been approved (HumanOversight layer).
	Approved bool
	// ApprovedBy is the approver identifier ("human" or system ID).
	ApprovedBy string
	// ApprovedAt is the approval timestamp.
	ApprovedAt time.Time
}

// CanaryResult represents the canary evaluation result.
type CanaryResult struct {
	// Available indicates whether canary execution is possible.
	Available bool
	// EvaluatedSpecs is the list of evaluated SPEC IDs (maximum 3).
	EvaluatedSpecs []string
	// ScoreBefore is the average score before amendment.
	ScoreBefore float64
	// ScoreAfter is the average score after amendment (shadow evaluation).
	ScoreAfter float64
	// MaxDrop is the maximum score drop.
	MaxDrop float64
	// Passed indicates whether canary passed (ScoreDrop <= 0.10).
	Passed bool
	// Reason is the cause for canary unavailability or failure.
	Reason string
}

// ContradictionResult represents the contradiction detection result.
type ContradictionResult struct {
	// Conflicts are the conflicting rule IDs and descriptions.
	Conflicts []ConflictDetail
	// HasBlockingContradiction indicates whether a blocking contradiction exists.
	HasBlockingContradiction bool
}

// ConflictDetail represents a single contradiction detail.
type ConflictDetail struct {
	ConflictingRuleID string
	Description       string
	IsBlocking        bool // If true, cannot proceed without resolution
}

// FrozenGuard is Layer 1 safety gate.
// Blocks modifications to Frozen zone rules.
type FrozenGuard interface {
	// Check checks if the proposal modifies a Frozen zone rule.
	// Frozen→Evolvable demotion is allowed with Evidence.
	Check(proposal *AmendmentProposal, currentZone Zone) error
}

// Canary is Layer 2 safety gate.
// Detects score drops through shadow evaluation.
type Canary interface {
	// Evaluate evaluates the proposal's impact on the last 3 completed SPECs.
	// < 3 SPECs → CanaryUnavailable error.
	// ScoreDrop > 0.10 → CanaryRejected error.
	// canary_gate: false rule → CanaryUnavailable (skip).
	Evaluate(proposal *AmendmentProposal, projectDir string) (*CanaryResult, error)
}

// ContradictionDetector is Layer 3 safety gate.
// Detects contradictions with existing rules.
type ContradictionDetector interface {
	// Scan checks if the proposal contradicts other rules in the registry.
	// Returns a list of ConflictDetail if contradictions are found.
	Scan(proposal *AmendmentProposal, registry *Registry) (*ContradictionResult, error)
}

// RateLimiter is Layer 4 safety gate.
// Limits amendment frequency.
type RateLimiter interface {
	// Admit checks if the proposal is within rate limits.
	// Max 3 amendments/7-day window, 24h cooldown, 50 active learnings cap.
	// Rolled-back rules have 30-day cooldown.
	Admit(proposal *AmendmentProposal, evolutionLogPath string) error
}

// HumanOversight is Layer 5 safety gate.
// Obtains approval through terminal Y/N prompt in CLI.
// (AskUserQuestion is orchestrator-only, so implemented directly in CLI).
type HumanOversight interface {
	// Approve shows the proposal diff to the user and requests approval.
	// In dry-run mode, always returns true (no actual approval).
	Approve(proposal *AmendmentProposal, dryRun bool) (bool, error)
}

// AmendmentLog represents an evolution-log.md entry.
// Direct implementation of SPEC-V3R2-CON-002 REQ-CON-002-003.
type AmendmentLog struct {
	// ID is the unique identifier in LEARN-YYYYMMDD-NNN format.
	ID string
	// RuleID is the modified rule ID (CONST-V3R2-NNN).
	RuleID string
	// ZoneBefore is the zone before modification.
	ZoneBefore Zone
	// ZoneAfter is the zone after modification.
	ZoneAfter Zone
	// ClauseBefore is the clause text before modification.
	ClauseBefore string
	// ClauseAfter is the clause text after modification.
	ClauseAfter string
	// CanaryVerdict is the canary evaluation result.
	CanaryVerdict string // "passed", "skipped", "rejected", "unavailable"
	// Contradictions are the contradiction detection results.
	Contradictions []string
	// ApprovedBy is the approver ("human" or system ID).
	ApprovedBy string
	// ApprovedAt is the approval timestamp.
	ApprovedAt time.Time
	// RolledBack indicates whether it has been rolled back.
	RolledBack bool
	// RollbackReason is the rollback reason (optional).
	RollbackReason string
	// RollbackAt is the rollback timestamp (optional).
	RollbackAt *time.Time
}

// Validate validates the required fields of AmendmentLog.
func (l *AmendmentLog) Validate() error {
	if l.ID == "" {
		return fmt.Errorf("amendment log ID is empty")
	}
	if l.RuleID == "" {
		return fmt.Errorf("rule ID is empty")
	}
	if l.ClauseBefore == "" || l.ClauseAfter == "" {
		return fmt.Errorf("clause is empty")
	}
	if l.ApprovedBy == "" {
		return fmt.Errorf("approved by is empty")
	}
	if l.ApprovedAt.IsZero() {
		return fmt.Errorf("approved at is empty")
	}
	return nil
}
