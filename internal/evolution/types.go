// Package evolution implements the Reflective Learning "Write Phase" engine
// (SPEC-REFLECT-001).  It analyzes per-session telemetry produced by
// SPEC-TELEMETRY-001 and proposes targeted improvements to evolvable zones
// inside MoAI skill files.
//
// Design constraints:
//   - User approval is ALWAYS required before any change is applied to disk.
//     The engine only writes proposal files; it never auto-applies them.
//   - Safety layers (FrozenGuard, RateLimiter) block changes to constitutional
//     files and prevent proposal floods.
//   - When in doubt, do NOT generate a proposal (false negatives >> false positives).
package evolution

import (
	"errors"
	"time"
)

// Status represents the confidence tier of a LearningEntry.
// Tiers follow the graduation pipeline from Agency Constitution Section 6.
type Status string

const (
	// StatusObservation is the initial tier — logged but no action taken.
	StatusObservation Status = "observation"

	// StatusHeuristic is promoted after 3 observations.
	StatusHeuristic Status = "heuristic"

	// StatusRule is eligible for graduation after 5 observations + confidence >= 0.80.
	StatusRule Status = "rule"

	// StatusHighConfidence triggers an auto-proposal after 10 observations.
	StatusHighConfidence Status = "high-confidence"

	// StatusGraduated means the proposed change has been approved and applied.
	StatusGraduated Status = "graduated"

	// StatusAntiPattern is set on a single critical failure (never removed by engine).
	StatusAntiPattern Status = "anti-pattern"

	// StatusArchived is set when the entry is moved to the archive.
	StatusArchived Status = "archived"
)

// LearningEntry records a pattern observed across multiple sessions for a
// specific skill and evolvable zone.
type LearningEntry struct {
	// ID is the unique learning identifier in format LEARN-YYYYMMDD-NNN.
	ID string

	// Status is the current graduation tier.
	Status Status

	// SkillID is the MoAI skill identifier (e.g. "moai-lang-go").
	SkillID string

	// ZoneID is the evolvable zone inside the skill file.
	ZoneID string

	// Observations is the count of times this pattern has been seen.
	Observations int

	// Confidence is a value in [0, 1] combining outcome consistency,
	// observation frequency, and recency.
	Confidence float64

	// Created is the UTC time the entry was first recorded.
	Created time.Time

	// Updated is the UTC time the entry was last modified.
	Updated time.Time

	// Observation is a human-readable description of the observed pattern.
	Observation string

	// Evidence holds per-session evidence records supporting this entry.
	Evidence []EvidenceEntry

	// ProposedChange is the proposed modification once the entry reaches
	// StatusRule or StatusHighConfidence.  Nil until a change is proposed.
	ProposedChange *ProposedChange
}

// EvidenceEntry records one session's contribution to a LearningEntry.
type EvidenceEntry struct {
	// SessionID is the Claude Code session that provided this evidence.
	SessionID string

	// Date is the ISO-8601 date the evidence was recorded (YYYY-MM-DD).
	Date string

	// Context is a short human-readable description of what happened.
	Context string
}

// ProposedChange describes a concrete modification to an evolvable zone
// pending user approval.
type ProposedChange struct {
	// TargetFile is the project-root-relative path to the skill file.
	TargetFile string

	// ZoneID is the evolvable zone inside TargetFile to modify.
	ZoneID string

	// Addition is the text to insert into the zone (appended to existing
	// zone content).  The change is additive only — deletion requires a
	// human edit.
	Addition string
}

// RateState tracks weekly proposal counts for the rate limiter.
type RateState struct {
	// WeekStart is the Monday of the current tracking week (UTC, YYYY-MM-DD).
	WeekStart string `yaml:"week_start"`

	// ProposalsThisWeek is the count of proposals generated this week.
	ProposalsThisWeek int `yaml:"proposals_this_week"`

	// LastProposalTimes maps target-file paths to the last proposal time (RFC3339).
	// Used to enforce the 24-hour per-file cooldown.
	LastProposalTimes map[string]string `yaml:"last_proposal_times,omitempty"`
}

// LearningFilter specifies criteria for listing learning entries.
type LearningFilter struct {
	// SkillID filters by skill identifier.  Empty means all skills.
	SkillID string

	// Status filters by status tier.  Empty Status means all tiers.
	Status Status

	// ExcludeArchived excludes entries with StatusArchived when true.
	ExcludeArchived bool
}

// Sentinel errors for the evolution package.
var (
	// ErrFrozenPath is returned by CheckFrozenGuard when the target file is
	// protected and may not be modified by the engine.
	ErrFrozenPath = errors.New("evolution: target file is in the frozen zone")

	// ErrRateLimit is returned by CheckRateLimit when the weekly proposal
	// quota or per-file cooldown has been exceeded.
	ErrRateLimit = errors.New("evolution: rate limit exceeded")

	// ErrDuplicateLearning is returned by CreateLearning when a sufficiently
	// similar observation already exists.
	ErrDuplicateLearning = errors.New("evolution: duplicate learning detected")

	// ErrMaxLearningsReached is returned by CreateLearning when the active
	// learning count has hit the configured limit.
	ErrMaxLearningsReached = errors.New("evolution: maximum active learnings reached")

	// ErrZoneNotFound is returned when the target evolvable zone cannot be
	// located in the skill file.
	ErrZoneNotFound = errors.New("evolution: evolvable zone not found in target file")
)

const (
	// MaxActiveLearnings is the hard cap on the number of non-archived
	// learning entries.  When the count hits this value, the oldest entries
	// are archived before new ones are created.
	MaxActiveLearnings = 50

	// MaxProposalsPerWeek is the maximum number of proposals the engine may
	// generate in a single calendar week.
	MaxProposalsPerWeek = 3

	// ProposalCooldownHours is the minimum number of hours between two
	// proposals that target the same file.
	ProposalCooldownHours = 24

	// HeuristicThreshold is the minimum observation count for StatusHeuristic.
	HeuristicThreshold = 3

	// RuleThreshold is the minimum observation count for StatusRule.
	RuleThreshold = 5

	// HighConfidenceThreshold is the minimum observation count for
	// StatusHighConfidence.
	HighConfidenceThreshold = 10

	// RuleConfidenceThreshold is the minimum Confidence score required for
	// a StatusRule entry to be eligible for a proposal.
	RuleConfidenceThreshold = 0.80

	// MaxSessionProposals is the maximum number of proposals generated per
	// AnalyzeSession call.
	MaxSessionProposals = 3

	// MinToolInvocationsForAnalysis is the minimum number of tool calls in a
	// session before analysis is performed.
	MinToolInvocationsForAnalysis = 3
)
