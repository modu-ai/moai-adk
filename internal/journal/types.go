// Package journal provides crash-safe session logging for MoAI-ADK workflows.
// Journal entries are stored as append-only JSONL files in SPEC directories,
// enabling context recovery after token exhaustion, terminal crashes, or context compaction.
package journal

import "time"

// Entry represents a single journal event.
type Entry struct {
	ID          string            `json:"id"`
	SessionID   string            `json:"session_id"`
	Timestamp   time.Time         `json:"ts"`
	Type        string            `json:"type"`              // session_start, phase_begin, phase_end, checkpoint, session_end, crash_detected
	SpecID      string            `json:"spec_id,omitempty"`
	Phase       string            `json:"phase,omitempty"`
	Status      string            `json:"status"`            // in_progress, completed, interrupted, failed
	Context     map[string]string `json:"context,omitempty"` // free-form KV: files, next_step, etc.
	TokensUsed  int               `json:"tokens_used,omitempty"`
	IssueNumber int               `json:"issue_number,omitempty"`
}

// ActionEntry represents a single agent action for replay logging.
// Actions are buffered and flushed periodically or at session boundaries.
type ActionEntry struct {
	Timestamp  time.Time `json:"ts"`
	SessionID  string    `json:"session_id"`
	AgentName  string    `json:"agent_name"`
	AgentID    string    `json:"agent_id"`
	Action     string    `json:"action"`                // "start", "tool_use", "complete", "error"
	ToolName   string    `json:"tool_name,omitempty"`
	DurationMS int64     `json:"duration_ms,omitempty"`
	Details    string    `json:"details,omitempty"`
}

// ResumeContext is synthesized from journal entries for session recovery.
type ResumeContext struct {
	SpecID         string   `json:"spec_id"`
	LastSessionID  string   `json:"last_session_id"`
	LastPhase      string   `json:"last_phase"`
	LastStatus     string   `json:"last_status"`
	EndReason      string   `json:"end_reason"`
	TokensUsed     int      `json:"tokens_used"`
	SessionCount   int      `json:"session_count"`
	FilesModified  []string `json:"files_modified,omitempty"`
	NextAction     string   `json:"next_action,omitempty"`
	Resumable      bool     `json:"resumable"`
}
