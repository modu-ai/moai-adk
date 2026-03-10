// Package statusline implements the MoAI-ADK statusline rendering system
// for Claude Code integration. It collects data from multiple sources
// (git, context window, session cost, version) and renders a single-line
// status display with color coding and multiple display modes.
package statusline

import (
	"context"
	"io"
)

// StatuslineMode defines the display verbosity of the statusline.
type StatuslineMode string

const (
	// ModeCompact displays all info in a compact 2-line layout (v3 name).
	ModeCompact StatuslineMode = "compact"

	// ModeDefault displays git status, context ratio in a 3-line layout.
	ModeDefault StatuslineMode = "default"

	// ModeFull displays all collected data in detailed 5-line layout (v3 name).
	ModeFull StatuslineMode = "full"

	// ModeMinimal is a deprecated alias for backward compatibility.
	// Use ModeCompact in new code.
	//
	// Deprecated: Use ModeCompact.
	ModeMinimal StatuslineMode = "minimal"

	// ModeVerbose is a deprecated alias for backward compatibility.
	// Use ModeFull in new code.
	//
	// Deprecated: Use ModeFull.
	ModeVerbose StatuslineMode = "verbose"
)

// NormalizeMode converts deprecated mode names to current names for backward compatibility.
// REQ-V3-MODE-001: "minimal" → "compact"
// REQ-V3-MODE-002: "verbose" → "full"
// Other values are returned unchanged.
func NormalizeMode(mode StatuslineMode) StatuslineMode {
	switch mode {
	case "minimal":
		return ModeCompact
	case "verbose":
		return ModeFull
	default:
		return mode
	}
}

// StdinData represents the JSON input from Claude Code's statusline hook.
// Matches the official JSON structure from https://code.claude.com/docs/en/statusline
type StdinData struct {
	HookEventName  string             `json:"hook_event_name"`
	SessionID      string             `json:"session_id"`
	TranscriptPath string             `json:"transcript_path"`
	CWD            string             `json:"cwd"` // Legacy field, prefer Workspace.CurrentDir
	Model          *ModelInfo         `json:"model"`
	Workspace      *WorkspaceInfo     `json:"workspace"`
	Cost           *CostData          `json:"cost"`
	ContextWindow  *ContextWindowInfo `json:"context_window"`
	OutputStyle    *OutputStyleInfo   `json:"output_style"`
	Version        string             `json:"version"` // Claude Code version (e.g., "1.0.80")
}

// ModelInfo represents the model information from Claude Code.
type ModelInfo struct {
	ID          string `json:"id"`           // e.g., "claude-opus-4-1"
	DisplayName string `json:"display_name"` // e.g., "Opus" - use this directly
	Name        string `json:"name"`         // Legacy field, same as ID
}

// WorkspaceInfo represents the workspace directory information from Claude Code.
type WorkspaceInfo struct {
	CurrentDir string `json:"current_dir"` // Current working directory
	ProjectDir string `json:"project_dir"` // Original project directory (use this for display)
}

// OutputStyleInfo represents the output style from Claude Code.
type OutputStyleInfo struct {
	Name string `json:"name"` // e.g., "MoAI", "R2-D2", "Yoda"
}

// CostData represents the session cost information from Claude Code.
type CostData struct {
	TotalUSD          float64 `json:"total_usd"`
	TotalCostUSD      float64 `json:"total_cost_usd"`
	InputTokens       int     `json:"input_tokens"`
	OutputTokens      int     `json:"output_tokens"`
	TotalDurationMS   int     `json:"total_duration_ms"`
	TotalLinesAdded   int     `json:"total_lines_added"`
	TotalLinesRemoved int     `json:"total_lines_removed"`
}

// ContextWindowInfo represents the context window usage from Claude Code.
// Matches the official JSON structure from https://code.claude.com/docs/en/statusline
type ContextWindowInfo struct {
	// Pre-calculated percentages (most accurate - use these directly)
	UsedPercentage      *float64 `json:"used_percentage"`      // 0-100, pre-calculated by Claude Code
	RemainingPercentage *float64 `json:"remaining_percentage"` // 0-100, pre-calculated by Claude Code
	ContextWindowSize   int      `json:"context_window_size"`  // e.g., 200000

	// Cumulative session totals
	TotalInputTokens  int `json:"total_input_tokens"`  // Cumulative input tokens across session
	TotalOutputTokens int `json:"total_output_tokens"` // Cumulative output tokens across session

	// Legacy/fallback: raw token counts (for backward compatibility)
	Used  int `json:"used"`
	Total int `json:"total"`

	// Current usage breakdown (may be null if no messages yet)
	CurrentUsage *CurrentUsageInfo `json:"current_usage"`
}

// CurrentUsageInfo contains detailed token usage breakdown.
type CurrentUsageInfo struct {
	InputTokens         int `json:"input_tokens"`
	CacheCreationTokens int `json:"cache_creation_input_tokens"`
	CacheReadTokens     int `json:"cache_read_input_tokens"`
	OutputTokens        int `json:"output_tokens"`
}

// StatusData aggregates all collected data for rendering.
type StatusData struct {
	Git               GitStatusData
	Memory            MemoryData
	Metrics           MetricsData
	Version           VersionData  // MoAI-ADK version from config
	ClaudeCodeVersion string       // Claude Code version from JSON input (e.g., "1.0.80")
	Directory         string       // Project directory name (e.g., "modu-saju")
	OutputStyle       string       // Output style name (e.g., "Mr.Alfred", "R2-D2")
	Task              TaskData     // Current active task (rendering enabled in Phase 4)
	Usage             *UsageResult // API usage (nil when unavailable)
}

// GitStatusData holds git repository status information.
type GitStatusData struct {
	Branch    string
	Modified  int
	Staged    int
	Untracked int
	Ahead     int
	Behind    int
	Available bool
}

// MemoryData holds context window token usage information.
type MemoryData struct {
	TokensUsed  int
	TokenBudget int
	Available   bool
}

// MetricsData holds session cost and model information.
type MetricsData struct {
	Model             string
	CostUSD           float64
	SessionDurationMS int // Total session duration in milliseconds (REQ-V3-TIME-001)
	Available         bool
}

// VersionData holds version and update information.
type VersionData struct {
	Current         string
	Latest          string
	UpdateAvailable bool
	Available       bool
}

// Segment key constants identify individual statusline segments.
const (
	SegmentModel         = "model"
	SegmentContext       = "context"
	SegmentOutputStyle   = "output_style"
	SegmentDirectory     = "directory"
	SegmentGitStatus     = "git_status"
	SegmentClaudeVersion = "claude_version"
	SegmentMoaiVersion   = "moai_version"
	SegmentGitBranch     = "git_branch"

	// v3 new segment constants (REQ-V3-TIME-003, enabled in Phase 4)
	SegmentSessionTime = "session_time" // Session duration
	SegmentUsage5H     = "usage_5h"     // 5-hour API usage
	SegmentUsage7D     = "usage_7d"     // 7-day API usage
	SegmentTask        = "task"         // Current active task
)

// UsageData represents API usage information.
type UsageData struct {
	UsedTokens  int64   `json:"used_tokens"`
	LimitTokens int64   `json:"limit_tokens"`
	Percentage  float64 `json:"percentage"`          // 0-100
	ResetsAt    string  `json:"resets_at,omitempty"` // ISO 8601 reset timestamp
}

// UsageResult represents 5-hour/7-day usage results.
type UsageResult struct {
	Usage5H *UsageData // 5-hour usage
	Usage7D *UsageData // 7-day usage
}

// contextLevel represents the severity level for context window usage coloring.
type contextLevel int

const (
	levelOk    contextLevel = iota // < 50% usage
	levelWarn                      // 50-80% usage
	levelError                     // >= 80% usage
)

// GitDataProvider abstracts git data collection for testability.
type GitDataProvider interface {
	// CollectGitStatus retrieves current git repository status.
	// Returns empty GitStatusData (not error) when no git repo exists.
	CollectGitStatus(ctx context.Context) (*GitStatusData, error)
}

// UpdateProvider abstracts version update checking for testability.
type UpdateProvider interface {
	// CheckUpdate checks if a newer version is available.
	// Uses caching to avoid repeated network calls.
	CheckUpdate(ctx context.Context) (*VersionData, error)
}

// Builder composes the statusline output from collected data.
type Builder interface {
	// Build generates the formatted statusline string from the given input.
	// Reads JSON from r, collects data from all sources, and returns
	// a single-line formatted string. Never returns an error that prevents
	// output; always produces at least a minimal fallback string.
	Build(ctx context.Context, r io.Reader) (string, error)

	// SetMode switches between statusline display modes.
	SetMode(mode StatuslineMode)
}
