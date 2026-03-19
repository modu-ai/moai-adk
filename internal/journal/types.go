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
