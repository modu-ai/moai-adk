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
	// ModeCompact는 한 줄에 모든 정보를 표시하는 기본 모드다 (v3 명칭).
	ModeCompact StatuslineMode = "compact"

	// ModeDefault는 git 상태, 컨텍스트 비율, 비용을 표시하는 기본 모드다.
	// 내부적으로 "compact"와 동일하게 동작한다.
	ModeDefault StatuslineMode = "default"

	// ModeFull은 모든 수집 데이터를 상세하게 표시하는 모드다 (v3 명칭).
	ModeFull StatuslineMode = "full"

	// ModeMinimal은 하위 호환성을 위한 deprecated 별칭이다.
	// 새 코드에서는 ModeCompact를 사용할 것.
	//
	// Deprecated: ModeCompact를 사용하라.
	ModeMinimal StatuslineMode = "minimal"

	// ModeVerbose는 하위 호환성을 위한 deprecated 별칭이다.
	// 새 코드에서는 ModeFull을 사용할 것.
	//
	// Deprecated: ModeFull을 사용하라.
	ModeVerbose StatuslineMode = "verbose"
)

// NormalizeMode는 하위 호환성을 위해 deprecated 모드 이름을 현재 이름으로 변환한다.
// REQ-V3-MODE-001: "minimal" → "compact"
// REQ-V3-MODE-002: "verbose" → "full"
// 그 외 값은 변경 없이 반환한다.
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
	Version           VersionData // MoAI-ADK version from config
	ClaudeCodeVersion string      // Claude Code version from JSON input (e.g., "1.0.80")
	Directory         string      // Project directory name (e.g., "modu-saju")
	OutputStyle       string      // Output style name (e.g., "Mr.Alfred", "R2-D2")
	Task              TaskData    // 현재 활성 태스크 (Phase 4에서 렌더링 활성화)
	Usage             *UsageResult // API 사용량 (nil이면 사용 불가)
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
	SessionDurationMS int  // 세션 총 실행 시간 (밀리초, REQ-V3-TIME-001)
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

	// v3 신규 세그먼트 상수 (REQ-V3-TIME-003, Phase 4에서 활성화됨)
	SegmentSessionTime = "session_time" // 세션 실행 시간
	SegmentUsage5H     = "usage_5h"     // 5시간 API 사용량
	SegmentUsage7D     = "usage_7d"     // 7일 API 사용량
	SegmentTask        = "task"         // 현재 활성 태스크
)

// UsageData 는 API 사용량 정보를 나타낸다.
type UsageData struct {
	UsedTokens  int64   `json:"used_tokens"`
	LimitTokens int64   `json:"limit_tokens"`
	Percentage  float64 `json:"percentage"` // 0-100
}

// UsageResult 는 5시간/7일 사용량 결과를 나타낸다.
type UsageResult struct {
	Usage5H *UsageData // 5시간 사용량
	Usage7D *UsageData // 7일 사용량
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
