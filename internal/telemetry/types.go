package telemetry

import "time"

// UsageRecord represents a single skill invocation event.
// All fields are serialized to JSON for JSONL storage.
type UsageRecord struct {
	Timestamp   time.Time `json:"ts"`
	SessionID   string    `json:"session_id"`
	SkillID     string    `json:"skill_id"`
	Trigger     string    `json:"trigger"`       // "explicit" | "auto"
	ContextHash string    `json:"context_hash"`  // SHA-256 first 8 chars (no PII)
	AgentType   string    `json:"agent_type"`
	Phase       string    `json:"phase"` // plan|run|sync|none
	DurationMs  int64     `json:"duration_ms"`
	Outcome     string    `json:"outcome"` // success|partial|error|unknown
}

// Constants for trigger types.
const (
	TriggerExplicit = "explicit"
	TriggerAuto     = "auto"
)

// Constants for outcome types.
const (
	OutcomeSuccess = "success"
	OutcomePartial = "partial"
	OutcomeError   = "error"
	OutcomeUnknown = "unknown"
)

// Event represents a session event for outcome determination.
// Heuristics use these events to infer whether a skill invocation was successful.
type Event struct {
	// ToolName is the name of the tool that was invoked (e.g., "Bash", "Write", "Read").
	ToolName string
	// IsError indicates whether this event represents an error condition.
	IsError bool
	// IsTestPass indicates whether this event represents all tests passing.
	IsTestPass bool
	// IsTestFail indicates whether this event represents a test failure.
	IsTestFail bool
}

// Report is the aggregated telemetry report for a time window.
type Report struct {
	// Days is the time window covered by the report.
	Days int
	// Skills contains per-skill aggregated stats.
	Skills []SkillStats
	// CoOccurrences lists pairs of skills that appear in the same session.
	CoOccurrences []CoOccurrence
	// Underutilized lists skills with fewer than 3 uses in the window.
	Underutilized []UnderutilizedSkill
}

// SkillStats holds aggregated usage statistics for a single skill.
type SkillStats struct {
	SkillID string
	Uses    int
	Success int
	Partial int
	Error   int
	Unknown int
}

// CoOccurrence describes two skills appearing together in sessions.
type CoOccurrence struct {
	SkillA string
	SkillB string
	Count  int
}

// UnderutilizedSkill describes a skill with low usage.
type UnderutilizedSkill struct {
	SkillID string
	Uses    int
}
