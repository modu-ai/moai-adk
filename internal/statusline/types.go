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
	// ModeMinimal shows only model name and context percentage.
	ModeMinimal StatuslineMode = "minimal"

	// ModeDefault shows git status, context percentage, and cost.
	ModeDefault StatuslineMode = "default"

	// ModeVerbose shows all collected data with full detail.
	ModeVerbose StatuslineMode = "verbose"
)

// StdinData represents the JSON input from Claude Code's statusline hook.
type StdinData struct {
	HookEventName string         `json:"hook_event_name"`
	SessionID     string         `json:"session_id"`
	CWD           string         `json:"cwd"`
	Model         string         `json:"model"`
	Workspace     string         `json:"workspace"`
	Cost          *CostData      `json:"cost"`
	ContextWindow *ContextWindow `json:"context_window"`
}

// CostData represents the session cost information from Claude Code.
type CostData struct {
	TotalUSD     float64 `json:"total_usd"`
	InputTokens  int     `json:"input_tokens"`
	OutputTokens int     `json:"output_tokens"`
}

// ContextWindow represents the context window usage from Claude Code.
type ContextWindow struct {
	Used  int `json:"used"`
	Total int `json:"total"`
}

// StatusData aggregates all collected data for rendering.
type StatusData struct {
	Git     GitStatusData
	Memory  MemoryData
	Metrics MetricsData
	Version VersionData
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
	Model     string
	CostUSD   float64
	Available bool
}

// VersionData holds version and update information.
type VersionData struct {
	Current         string
	Latest          string
	UpdateAvailable bool
	Available       bool
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
