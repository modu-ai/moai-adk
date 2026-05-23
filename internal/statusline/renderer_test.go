package statusline

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

// newTestRenderer creates a Renderer with NoColor=true for predictable test output.
func newTestRenderer() *Renderer {
	return NewRenderer("default", true, nil)
}

func TestRender_MinimalMode(t *testing.T) {
	r := newTestRenderer()
	data := &StatusData{
		Metrics: MetricsData{Model: "Opus 4.5", Available: true},
		Memory:  MemoryData{TokensUsed: 50000, TokenBudget: 200000, Available: true},
	}

	got := r.Render(data, ModeMinimal)

	if !strings.Contains(got, "🤖 Opus 4.5") {
		t.Errorf("minimal mode should contain model name with emoji, got %q", got)
	}
	if !strings.Contains(got, "🔋") {
		t.Errorf("minimal mode should contain battery emoji, got %q", got)
	}
	if !strings.Contains(got, "25%") {
		t.Errorf("minimal mode should contain context percentage, got %q", got)
	}
}

func TestRender_DefaultMode(t *testing.T) {
	r := newTestRenderer()
	data := &StatusData{
		Git:         GitStatusData{Branch: "main", Modified: 2, Staged: 3, Untracked: 1, Available: true},
		Memory:      MemoryData{TokensUsed: 50000, TokenBudget: 200000, Available: true},
		Metrics:     MetricsData{Model: "Opus 4.5", Available: true},
		Directory:   "moai-adk-go",
		OutputStyle: "Mr.Alfred",
		Version:     VersionData{Current: "1.14.5", Available: true},
		Workspace: WorkspaceData{
			Repo: &RepoInfo{Host: "github.com", Owner: "modu-ai", Name: "moai-adk"},
		},
	}

	got := r.Render(data, ModeDefault)

	// Check all sections are present with emojis
	if !strings.Contains(got, "🤖 Opus 4.5") {
		t.Errorf("default mode should contain model with emoji, got %q", got)
	}
	if !strings.Contains(got, "🔋") {
		t.Errorf("default mode should contain battery emoji, got %q", got)
	}
	if !strings.Contains(got, "💬 Mr.Alfred") {
		t.Errorf("default mode should contain output style with emoji, got %q", got)
	}
	if !strings.Contains(got, "📁 moai-adk-go") {
		t.Errorf("default mode should contain directory with emoji, got %q", got)
	}
	if !strings.Contains(got, "🅱️") && !strings.Contains(got, "📬") && !strings.Contains(got, "📫") {
		t.Errorf("default mode should contain git status with emoji (🅱️ or mailbox), got %q", got)
	}
	if !strings.Contains(got, "+3") {
		t.Errorf("default mode should contain staged count, got %q", got)
	}
	if !strings.Contains(got, "M2") {
		t.Errorf("default mode should contain modified count with 'M', got %q", got)
	}
	if !strings.Contains(got, "?1") {
		t.Errorf("default mode should contain untracked count, got %q", got)
	}
	if !strings.Contains(got, "🗿 v1.14.5") {
		t.Errorf("default mode should contain MoAI version with moai emoji, got %q", got)
	}
	if !strings.Contains(got, "main") {
		t.Errorf("default mode should contain branch with emoji, got %q", got)
	}
}

func TestRender_VerboseMode_MultiLine(t *testing.T) {
	// v3 maps ModeVerbose → ModeFull for 6-line layout
	r := newTestRenderer()
	data := &StatusData{
		Git:               GitStatusData{Branch: "main", Staged: 3, Modified: 2, Untracked: 1, Available: true},
		Memory:            MemoryData{TokensUsed: 50000, TokenBudget: 200000, Available: true},
		Metrics:           MetricsData{Model: "Sonnet 4", CostUSD: 0.42, Available: true},
		Directory:         "my-project",
		OutputStyle:       "Yoda",
		Version:           VersionData{Current: "1.2.0", Available: true},
		ClaudeCodeVersion: "1.0.80",
	}

	got := r.Render(data, ModeVerbose)

	// full mode outputs multiple lines
	if !strings.Contains(got, "\n") {
		t.Errorf("full mode should produce multi-line output, got %q", got)
	}

	lines := strings.Split(got, "\n")

	// v3 full L1: Model, Claude version, MoAI version (no prefix)
	if !strings.Contains(lines[0], "🤖 Sonnet 4") {
		t.Errorf("full line 1 should contain model, got %q", lines[0])
	}
	if !strings.Contains(lines[0], "v1.0.80") {
		t.Errorf("full line 1 should contain Claude version, got %q", lines[0])
	}
	if !strings.Contains(lines[0], "🗿 v1.2.0") {
		t.Errorf("full line 1 should contain MoAI version, got %q", lines[0])
	}
	// v3 does not render cost (REQ-V3-TIME-005)
	if strings.Contains(got, "$") {
		t.Errorf("full mode should NOT contain cost in v3, got %q", got)
	}
}

func TestRender_VerboseMode_OmitsEmptyLines(t *testing.T) {
	// REQ-SLE-034: omit lines when all segments unavailable
	r := newTestRenderer()
	// Only model + context available (no git, no directory, no version)
	data := &StatusData{
		Metrics: MetricsData{Model: "Sonnet 4", Available: true},
		Memory:  MemoryData{TokensUsed: 50000, TokenBudget: 200000, Available: true},
	}

	got := r.Render(data, ModeVerbose)

	lines := strings.Split(strings.TrimRight(got, "\n"), "\n")
	// Line 2 (directory/branch/git) should be omitted since all are unavailable
	for _, line := range lines {
		if strings.Contains(line, "📁") || strings.Contains(line, "🅱️") {
			t.Errorf("verbose mode should omit empty line 2, but found git/dir segment in %q", line)
		}
	}
}

func TestRender_VerboseMode_CostRendering(t *testing.T) {
	// REQ-V3-TIME-005: v3 does not render cost ($) segment
	r := newTestRenderer()
	data := &StatusData{
		Metrics:           MetricsData{Model: "Opus 4.5", CostUSD: 1.23, Available: true},
		Memory:            MemoryData{TokensUsed: 50000, TokenBudget: 200000, Available: true},
		Version:           VersionData{Current: "2.0.0", Available: true},
		ClaudeCodeVersion: "1.0.80",
	}

	got := r.Render(data, ModeVerbose)
	// v3 does not render cost
	if strings.Contains(got, "$1.23") || strings.Contains(got, "$") {
		t.Errorf("full mode should NOT render cost in v3, got %q", got)
	}
	// Version info should still be displayed (no prefix in full mode)
	if !strings.Contains(got, "v1.0.80") {
		t.Errorf("full mode should still contain Claude version, got %q", got)
	}
}

func TestRender_VerboseMode_ZeroCostOmitted(t *testing.T) {
	// Zero cost (no session cost yet) should not show $0.00
	r := newTestRenderer()
	data := &StatusData{
		Metrics:           MetricsData{Model: "Opus 4.5", CostUSD: 0, Available: true},
		Memory:            MemoryData{TokensUsed: 50000, TokenBudget: 200000, Available: true},
		Version:           VersionData{Current: "2.0.0", Available: true},
		ClaudeCodeVersion: "1.0.80",
	}

	got := r.Render(data, ModeVerbose)
	if strings.Contains(got, "$0.00") {
		t.Errorf("verbose mode should not render zero cost, got %q", got)
	}
}

func TestRender_MinimalMode_OutputStyle(t *testing.T) {
	// ModeMinimal maps to ModeDefault
	r := newTestRenderer()
	data := &StatusData{
		Metrics:     MetricsData{Model: "Sonnet 4", Available: true},
		Memory:      MemoryData{TokensUsed: 50000, TokenBudget: 200000, Available: true},
		OutputStyle: "MoAI",
	}

	got := r.Render(data, ModeMinimal)

	// Default mode: L1 (info) + L2 (bars) = 2 lines without git
	lines := strings.Split(got, "\n")
	if len(lines) != 2 {
		t.Errorf("minimal/default mode without git should be 2 lines, got %d lines: %q", len(lines), got)
	}
}

func TestRender_EmptyGit(t *testing.T) {
	r := newTestRenderer()
	data := &StatusData{
		Git:     GitStatusData{Available: false},
		Memory:  MemoryData{TokensUsed: 50000, TokenBudget: 200000, Available: true},
		Metrics: MetricsData{Model: "Haiku 3.5", Available: true},
	}

	got := r.Render(data, ModeDefault)

	// Should not contain any git info
	if strings.Contains(got, "🅱️") || strings.Contains(got, "📬") || strings.Contains(got, "📫") {
		t.Errorf("should not contain git status emoji when unavailable, got %q", got)
	}
	if strings.Contains(got, "🅱️") {
		t.Errorf("should not contain branch emoji when unavailable, got %q", got)
	}
	// But should still have model and context
	if !strings.Contains(got, "🤖") {
		t.Errorf("should still contain model emoji, got %q", got)
	}
	if !strings.Contains(got, "🔋") {
		t.Errorf("should still contain context, got %q", got)
	}
}

func TestRender_EmptyMemory(t *testing.T) {
	r := newTestRenderer()
	data := &StatusData{
		Git:     GitStatusData{Branch: "main", Modified: 2, Available: true},
		Memory:  MemoryData{Available: false},
		Metrics: MetricsData{Model: "Sonnet 3.5", Available: true},
		Workspace: WorkspaceData{
			Repo: &RepoInfo{Host: "github.com", Owner: "modu-ai", Name: "moai-adk"},
		},
	}

	got := r.Render(data, ModeDefault)

	if !strings.Contains(got, "main") {
		t.Errorf("should contain git info, got %q", got)
	}
	// CW bar should not display when memory is unavailable
	if strings.Contains(got, "CW:") {
		t.Errorf("should not contain CW bar when memory unavailable, got %q", got)
	}
	// 5H/7D always display at 0%
	if !strings.Contains(got, "5H:") || !strings.Contains(got, "7D:") {
		t.Errorf("should always contain 5H/7D bars, got %q", got)
	}
}

func TestRender_NilData(t *testing.T) {
	r := newTestRenderer()
	got := r.Render(nil, ModeDefault)
	if got != "MoAI" {
		t.Errorf("nil data should return fallback, got %q", got)
	}
}

func TestRender_AllEmpty(t *testing.T) {
	r := newTestRenderer()
	data := &StatusData{}
	got := r.Render(data, ModeDefault)
	// 5H/7D bars always display at 0%, so output exists even with empty data
	if got == "" {
		t.Errorf("empty data should still produce output (5H/7D bars), got empty")
	}
	if !strings.Contains(got, "5H:") || !strings.Contains(got, "7D:") {
		t.Errorf("empty data should contain 5H/7D 0%% bars, got %q", got)
	}
}

func TestRender_GitOnlyBranch(t *testing.T) {
	r := newTestRenderer()
	data := &StatusData{
		Git:    GitStatusData{Branch: "main", Staged: 0, Modified: 0, Untracked: 0, Available: true},
		Memory: MemoryData{TokensUsed: 50000, TokenBudget: 200000, Available: true},
		Workspace: WorkspaceData{
			Repo: &RepoInfo{Host: "github.com", Owner: "modu-ai", Name: "moai-adk"},
		},
	}

	got := r.Render(data, ModeDefault)

	// Layout v3 CH3 (2026-05-22 fix): repo info present -> "🔀 owner/name | 🅱️ branch" form.
	// clean -> no dirty suffix.
	if !strings.Contains(got, "🔀 modu-ai/moai-adk | 🅱️ main") && !strings.Contains(got, "📭 main +0") {
		t.Errorf("should show clean branch as '🔀 modu-ai/moai-adk | 🅱️ main' or '📭 main +0', got %q", got)
	}
}

// TestRender_GitOnlyBranch_NoRepo verifies the 🔀 segment is hidden entirely
// when Workspace.Repo is nil (git not initialized or no remote configured).
// User policy 2026-05-22: hide segment to avoid "🔀 (🅱️ main)" partial display.
func TestRender_GitOnlyBranch_NoRepo(t *testing.T) {
	r := newTestRenderer()
	data := &StatusData{
		Git:    GitStatusData{Branch: "main", Staged: 0, Modified: 0, Untracked: 0, Available: true},
		Memory: MemoryData{TokensUsed: 50000, TokenBudget: 200000, Available: true},
	}

	got := r.Render(data, ModeDefault)

	if strings.Contains(got, "🔀") {
		t.Errorf("repo info absent → 🔀 segment must be hidden, got %q", got)
	}
}

func TestRender_Separator(t *testing.T) {
	r := newTestRenderer()
	data := &StatusData{
		Git:     GitStatusData{Branch: "main", Modified: 2, Available: true},
		Memory:  MemoryData{TokensUsed: 50000, TokenBudget: 200000, Available: true},
		Metrics: MetricsData{Model: "Opus 4.5", Available: true},
	}

	got := r.Render(data, ModeDefault) // default mode: multiple segments on one line

	// v3 separator is " │ " (U+2502 box-drawing character)
	if !strings.Contains(got, " │ ") {
		t.Errorf("sections should be separated by ' │ ', got %q", got)
	}
}

func TestRender_NoNewline(t *testing.T) {
	// default mode with no git yields 2 lines (L1 + L2 bars)
	r := newTestRenderer()
	data := &StatusData{
		Metrics: MetricsData{Model: "Sonnet 4", Available: true},
		Memory:  MemoryData{TokensUsed: 50000, TokenBudget: 200000, Available: true},
		// no git
	}

	got := r.Render(data, ModeDefault) // default: 2 lines without git (info + bars)
	lines := strings.Split(got, "\n")

	if len(lines) != 2 {
		t.Errorf("default without git should be 2 lines, got %d lines: %q", len(lines), got)
	}
}

func TestNewRenderer_ThemeVariants(t *testing.T) {
	tests := []struct {
		theme   string
		noColor bool
	}{
		{"default", false},
		{"default", true},
		{"catppuccin-mocha", false},
		{"catppuccin-latte", false},
		{"unknown", false},
	}

	for _, tt := range tests {
		t.Run(tt.theme, func(t *testing.T) {
			r := NewRenderer(tt.theme, tt.noColor, nil)
			if r == nil {
				t.Fatal("NewRenderer returned nil")
			}
			if r.noColor != tt.noColor {
				t.Errorf("noColor = %v, want %v", r.noColor, tt.noColor)
			}
		})
	}
}

func TestNewRenderer_ThemeIsStored(t *testing.T) {
	// When a theme is supplied, the renderer should store it.
	r := NewRenderer("catppuccin-mocha", false, nil)
	if r == nil {
		t.Fatal("NewRenderer returned nil")
		return // staticcheck SA5011
	}
	if r.theme == nil {
		t.Error("Renderer.theme should be set (not nil) after NewRenderer")
	}
}

func TestNewRenderer_NoColorIgnoresTheme(t *testing.T) {
	// When NoColor=true, theme colors must NOT be applied (REQ-SLE-016).
	r := NewRenderer("catppuccin-mocha", true, nil)
	data := &StatusData{
		Metrics: MetricsData{Model: "Opus 4.5", Available: true},
		Memory:  MemoryData{TokensUsed: 50000, TokenBudget: 200000, Available: true},
	}
	got := r.Render(data, ModeDefault)
	// No ANSI escape codes should be present in no-color mode
	if strings.Contains(got, "\x1b[") {
		t.Errorf("NoColor=true output should not contain ANSI escapes, got %q", got)
	}
}

func TestRender_ContextBarGraph(t *testing.T) {
	r := newTestRenderer()
	data := &StatusData{
		Memory:  MemoryData{TokensUsed: 82000, TokenBudget: 200000, Available: true},
		Metrics: MetricsData{Model: "Opus 4.5", Available: true},
	}

	got := r.Render(data, ModeDefault)

	// Should contain bar graph characters
	if !strings.Contains(got, "█") {
		t.Errorf("should contain bar graph characters, got %q", got)
	}
	// Should contain percentage
	if !strings.Contains(got, "41%") {
		t.Errorf("should contain percentage, got %q", got)
	}
	// Should use 🔋 emoji (<=70% used)
	if !strings.Contains(got, "🔋") {
		t.Errorf("should use battery emoji for <=70%% usage, got %q", got)
	}
}

func TestRender_HighContextUsage(t *testing.T) {
	r := newTestRenderer()
	data := &StatusData{
		Memory:  MemoryData{TokensUsed: 180000, TokenBudget: 200000, Available: true},
		Metrics: MetricsData{Model: "Sonnet 4", Available: true},
	}

	got := r.Render(data, ModeDefault)

	// Should use 🪫 emoji (>70% used)
	if !strings.Contains(got, "🪫") {
		t.Errorf("should use empty battery emoji for >70%% usage, got %q", got)
	}
	if !strings.Contains(got, "90%") {
		t.Errorf("should show 90%%, got %q", got)
	}
}

func TestRender_ModelShortening(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"Opus 4.5", "claude-opus-4-5-20250514", "Opus 4.5"},
		{"Sonnet 4", "claude-sonnet-4-20250514", "Sonnet 4"},
		{"Sonnet 3.5", "claude-3-5-sonnet-20241022", "Sonnet 3.5"},
		{"Haiku 3.5", "claude-3-5-haiku-20241022", "Haiku 3.5"},
		{"Non-Claude", "gpt-4", "gpt-4"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ShortenModelName(tt.input)
			if got != tt.expected {
				t.Errorf("ShortenModelName(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestRender_WithOutputStyle(t *testing.T) {
	r := newTestRenderer()
	data := &StatusData{
		Metrics:     MetricsData{Model: "Opus 4.5", Available: true},
		Memory:      MemoryData{TokensUsed: 50000, TokenBudget: 200000, Available: true},
		OutputStyle: "R2-D2",
	}

	got := r.Render(data, ModeDefault)

	if !strings.Contains(got, "💬 R2-D2") {
		t.Errorf("should contain output style with emoji, got %q", got)
	}
}

func TestRender_VersionUpdateNotification(t *testing.T) {
	r := newTestRenderer()
	data := &StatusData{
		Metrics: MetricsData{Model: "Opus 4.5", Available: true},
		Memory:  MemoryData{TokensUsed: 50000, TokenBudget: 200000, Available: true},
		Version: VersionData{
			Current:         "2.0.0",
			Latest:          "2.0.1",
			UpdateAvailable: true,
			Available:       true,
		},
	}

	got := r.Render(data, ModeDefault)

	if !strings.Contains(got, "🗿 v2.0.0") {
		t.Errorf("should contain current version, got %q", got)
	}
	if !strings.Contains(got, "🗿 v2.0.1") {
		t.Errorf("should contain update notification, got %q", got)
	}
}

func TestRender_VersionNoUpdate(t *testing.T) {
	r := newTestRenderer()
	data := &StatusData{
		Metrics: MetricsData{Model: "Opus 4.5", Available: true},
		Memory:  MemoryData{TokensUsed: 50000, TokenBudget: 200000, Available: true},
		Version: VersionData{
			Current:         "2.0.0",
			UpdateAvailable: false,
			Available:       true,
		},
	}

	got := r.Render(data, ModeDefault)

	if !strings.Contains(got, "🗿 v2.0.0") {
		t.Errorf("should contain current version, got %q", got)
	}
	if strings.Contains(got, "⬆️") {
		t.Errorf("should NOT contain update notification when no update, got %q", got)
	}
}

func TestRender_WithDirectory(t *testing.T) {
	r := newTestRenderer()
	data := &StatusData{
		Metrics:   MetricsData{Model: "Sonnet 4", Available: true},
		Memory:    MemoryData{TokensUsed: 50000, TokenBudget: 200000, Available: true},
		Directory: "my-awesome-project",
	}

	got := r.Render(data, ModeDefault)

	if !strings.Contains(got, "📁 my-awesome-project") {
		t.Errorf("should contain directory with emoji, got %q", got)
	}
}

// --- TDD RED: Tests for segment filtering ---

func TestRender_SegmentFiltering(t *testing.T) {
	fullData := &StatusData{
		Git:               GitStatusData{Branch: "main", Modified: 2, Staged: 3, Untracked: 1, Available: true},
		Memory:            MemoryData{TokensUsed: 50000, TokenBudget: 200000, Available: true},
		Metrics:           MetricsData{Model: "Opus 4.5", Available: true},
		Directory:         "moai-adk-go",
		OutputStyle:       "MoAI",
		ClaudeCodeVersion: "1.0.80",
		Version:           VersionData{Current: "2.3.1", Available: true},
		Workspace: WorkspaceData{
			Repo: &RepoInfo{Host: "github.com", Owner: "modu-ai", Name: "moai-adk"},
		},
	}

	tests := []struct {
		name           string
		segmentConfig  map[string]bool
		wantContain    []string
		wantNotContain []string
	}{
		{
			name:          "nil config shows all segments",
			segmentConfig: nil,
			wantContain:   []string{"🤖 Opus 4.5", "🔋", "💬 MoAI", "📁 moai-adk-go", "🔀", "v1.0.80", "🗿 v2.3.1", "main"},
		},
		{
			name:          "empty config shows all segments",
			segmentConfig: map[string]bool{},
			wantContain:   []string{"🤖 Opus 4.5", "🔋", "💬 MoAI", "📁 moai-adk-go", "🔀", "v1.0.80", "🗿 v2.3.1", "main"},
		},
		{
			name: "model disabled hides model",
			segmentConfig: map[string]bool{
				SegmentModel: false, SegmentContext: true, SegmentOutputStyle: true,
				SegmentDirectory: true, SegmentGitStatus: true, SegmentClaudeVersion: true,
				SegmentMoaiVersion: true, SegmentGitBranch: true,
			},
			wantContain:    []string{"🔋", "💬 MoAI", "📁 moai-adk-go"},
			wantNotContain: []string{"🤖"},
		},
		{
			name: "minimal preset config",
			segmentConfig: map[string]bool{
				SegmentModel: true, SegmentContext: true, SegmentOutputStyle: false,
				SegmentDirectory: false, SegmentGitStatus: true, SegmentClaudeVersion: false,
				SegmentMoaiVersion: false, SegmentGitBranch: true,
			},
			wantContain:    []string{"🤖 Opus 4.5", "🔋", "🔀", "main"},
			wantNotContain: []string{"💬", "📁", "🔅", "🗿"},
		},
		{
			name: "all segments disabled returns MoAI fallback",
			segmentConfig: map[string]bool{
				SegmentModel: false, SegmentContext: false, SegmentOutputStyle: false,
				SegmentDirectory: false, SegmentGitStatus: false, SegmentClaudeVersion: false,
				SegmentMoaiVersion: false, SegmentGitBranch: false,
				SegmentUsage5H: false, SegmentUsage7D: false, SegmentSessionTime: false,
			},
			wantContain: []string{"MoAI"},
		},
		{
			name: "unknown segment key defaults to enabled",
			segmentConfig: map[string]bool{
				"unknown_segment": false,
				SegmentModel:      true,
			},
			wantContain: []string{"🤖 Opus 4.5", "🔋", "💬 MoAI", "📁 moai-adk-go"},
		},
		{
			name: "only context disabled",
			segmentConfig: map[string]bool{
				SegmentModel: true, SegmentContext: false, SegmentOutputStyle: true,
				SegmentDirectory: true, SegmentGitStatus: true, SegmentClaudeVersion: true,
				SegmentMoaiVersion: true, SegmentGitBranch: true,
				SegmentUsage5H: false, SegmentUsage7D: false,
			},
			wantContain:    []string{"🤖 Opus 4.5", "💬 MoAI", "📁 moai-adk-go", "main"},
			wantNotContain: []string{"CW:"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewRenderer("default", true, tt.segmentConfig)
			got := r.Render(fullData, ModeDefault)

			for _, want := range tt.wantContain {
				if !strings.Contains(got, want) {
					t.Errorf("should contain %q, got %q", want, got)
				}
			}
			for _, notWant := range tt.wantNotContain {
				if strings.Contains(got, notWant) {
					t.Errorf("should NOT contain %q, got %q", notWant, got)
				}
			}
		})
	}
}

func TestRender_SegmentFiltering_MinimalModeRespectsConfig(t *testing.T) {
	// v3 maps ModeMinimal → ModeDefault
	// default respects segment config
	segmentConfig := map[string]bool{
		SegmentModel: false, SegmentContext: false,
	}
	r := NewRenderer("default", true, segmentConfig)
	data := &StatusData{
		Metrics: MetricsData{Model: "Opus 4.5", Available: true},
		Memory:  MemoryData{TokensUsed: 50000, TokenBudget: 200000, Available: true},
	}

	got := r.Render(data, ModeMinimal) // ModeMinimal → ModeDefault

	// default: with model/context disabled, bars line still renders (5H/7D default to 0%)
	// Should contain usage bars even when model/context disabled
	if !strings.Contains(got, "5H:") {
		t.Errorf("with model/context disabled, default should still show 5H bar, got %q", got)
	}
}

func TestCostNotRendered(t *testing.T) {
	// REQ-V3-TIME-005: v3 does not render cost ($) segment
	r := newTestRenderer()
	data := &StatusData{
		Metrics:           MetricsData{Model: "Opus", CostUSD: 5.50, Available: true},
		Memory:            MemoryData{TokensUsed: 50000, TokenBudget: 200000, Available: true},
		Version:           VersionData{Current: "2.0.0", Available: true},
		ClaudeCodeVersion: "1.0.80",
	}

	// Verify cost rendering across all modes
	for _, mode := range []StatuslineMode{ModeDefault, ModeMinimal, ModeVerbose, ModeFull} {
		output := r.Render(data, mode)
		if strings.Contains(output, "$5.50") || strings.Contains(output, "$") {
			t.Errorf("cost should not be rendered in v3 (mode=%s), got: %s", mode, output)
		}
	}
}

func TestRenderGitBranchV3(t *testing.T) {
	tests := []struct {
		name     string
		git      GitStatusData
		worktree string
		want     string
	}{
		// Clean state → "🅱️ main +0"
		{"clean", GitStatusData{Branch: "main", Available: true}, "", "🅱️ main +0"},
		// Modified files (no longer prefixed with 🔨 — mailbox/M-count carry the signal)
		{"modified", GitStatusData{Branch: "feat", Modified: 2, Available: true}, "", "🅱️ feat +2"},
		// Staged files (no longer prefixed with 📦)
		{"staged", GitStatusData{Branch: "main", Staged: 1, Available: true}, "", "🅱️ main +1"},
		// Modified + Staged: dirty count is sum, no prefix emoji
		{"modified+staged", GitStatusData{Branch: "main", Modified: 2, Staged: 1, Available: true}, "", "🅱️ main +3"},
		// Ahead only → "🅱️ feat ↑2 +0"
		{"ahead", GitStatusData{Branch: "feat", Ahead: 2, Available: true}, "", "🅱️ feat ↑2 +0"},
		// Behind only → "🅱️ feat ↓1 +0"
		{"behind", GitStatusData{Branch: "feat", Behind: 1, Available: true}, "", "🅱️ feat ↓1 +0"},
		// Ahead + Behind → "🅱️ feat ↑2 ↓1 +0"
		{"ahead+behind", GitStatusData{Branch: "feat", Ahead: 2, Behind: 1, Available: true}, "", "🅱️ feat ↑2 ↓1 +0"},
		// Ahead + dirty → "🅱️ feat ↑2 +3"
		{"ahead+dirty", GitStatusData{Branch: "feat", Ahead: 2, Modified: 3, Available: true}, "", "🅱️ feat ↑2 +3"},
		// Zero ahead/behind explicitly omitted
		{"zero ahead/behind", GitStatusData{Branch: "main", Ahead: 0, Behind: 0, Available: true}, "", "🅱️ main +0"},
		// Unavailable → empty string
		{"unavailable", GitStatusData{Available: false}, "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := &StatusData{Git: tt.git, Worktree: tt.worktree}
			got := renderGitBranch(data)
			if got != tt.want {
				t.Errorf("renderGitBranch() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestRenderSessionTime(t *testing.T) {
	tests := []struct {
		name string
		ms   int
		want string
	}{
		// REQ-V3-TIME-002: 83min = "⏳ 1h 23m"
		{"83 minutes", 4980000, "⏳ 1h 23m"},
		// REQ-V3-TIME-002: under 45min = "⏳ 45m"
		{"45 minutes", 2700000, "⏳ 45m"},
		// REQ-V3-TIME-004: 0ms → empty string
		{"zero", 0, ""},
		// REQ-V3-TIME-002: exactly 1 hour
		{"exactly 1 hour", 3600000, "⏳ 1h 0m"},
		// REQ-V3-TIME-002: 26 hours = "⏳ 1d 2h"
		{"26 hours", 93600000, "⏳ 1d 2h"},
		// REQ-V3-TIME-002: 48 hours = "⏳ 2d 0h"
		{"48 hours", 172800000, "⏳ 2d 0h"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := renderSessionTime(tt.ms)
			if got != tt.want {
				t.Errorf("renderSessionTime(%d) = %q, want %q", tt.ms, got, tt.want)
			}
		})
	}
}

func TestIsSegmentEnabled(t *testing.T) {
	tests := []struct {
		name          string
		segmentConfig map[string]bool
		key           string
		want          bool
	}{
		{"nil config always enabled", nil, SegmentModel, true},
		{"empty config always enabled", map[string]bool{}, SegmentModel, true},
		{"enabled segment", map[string]bool{SegmentModel: true}, SegmentModel, true},
		{"disabled segment", map[string]bool{SegmentModel: false}, SegmentModel, false},
		{"unknown key defaults to enabled", map[string]bool{SegmentModel: true}, "unknown", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewRenderer("default", true, tt.segmentConfig)
			got := r.isSegmentEnabled(tt.key)
			if got != tt.want {
				t.Errorf("isSegmentEnabled(%q) = %v, want %v", tt.key, got, tt.want)
			}
		})
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Cycle 2: renderDefaultV3 tests
// REQ-V3-LAYOUT-002: default mode uses 4-line layout
// ─────────────────────────────────────────────────────────────────────────────

func TestRenderDefaultV3_ThreeLines(t *testing.T) {
	// default mode must produce exactly 3 lines (L4 removed, style merged into L1)
	r := newTestRenderer()
	data := &StatusData{
		Metrics:           MetricsData{Model: "Opus 4.6", Available: true, SessionDurationMS: 9240000},
		Memory:            MemoryData{TokensUsed: 176000, TokenBudget: 200000, Available: true},
		ClaudeCodeVersion: "2.1.50",
		Version:           VersionData{Current: "2.8.0", Available: true},
		Directory:         "moai-adk-go",
		Git: GitStatusData{
			Branch:    "feat/auth",
			Ahead:     2,
			Behind:    1,
			Staged:    3,
			Modified:  2,
			Untracked: 1,
			Available: true,
		},
		OutputStyle: "MoAI",
		Task:        TaskData{Active: true, Command: "run", SpecID: "SPEC-SLV3-001", Stage: "improve"},
		Usage: &UsageResult{
			Usage5H: &UsageData{UsedTokens: 45000, LimitTokens: 100000, Percentage: 45},
			Usage7D: &UsageData{UsedTokens: 82000, LimitTokens: 100000, Percentage: 82},
		},
	}

	got := r.Render(data, ModeDefault)
	lines := strings.Split(got, "\n")

	if len(lines) != 3 {
		t.Errorf("default mode must be 3 lines, got: %d lines\noutput: %q", len(lines), got)
	}
	// Output style must be merged into L1
	if !strings.Contains(lines[0], "💬 MoAI") {
		t.Errorf("default L1 must contain output style, got: %q", lines[0])
	}
}

func TestRenderDefaultV3_Line1(t *testing.T) {
	// default L1: model, Claude version, MoAI version, session time
	r := newTestRenderer()
	data := &StatusData{
		Metrics:           MetricsData{Model: "Opus 4.6", Available: true, SessionDurationMS: 9240000}, // 2h 34m
		Memory:            MemoryData{TokensUsed: 88000, TokenBudget: 200000, Available: true},
		ClaudeCodeVersion: "2.1.50",
		Version:           VersionData{Current: "2.8.0", Available: true},
		Directory:         "moai-adk-go",
		Git:               GitStatusData{Branch: "main", Available: true},
	}

	got := r.Render(data, ModeDefault)
	lines := strings.Split(got, "\n")
	l1 := lines[0]

	if !strings.Contains(l1, "🤖 Opus 4.6") {
		t.Errorf("default L1 must contain model, got: %q", l1)
	}
	if !strings.Contains(l1, "v2.1.50") {
		t.Errorf("default L1 must contain Claude version, got: %q", l1)
	}
	if !strings.Contains(l1, "🗿 v2.8.0") {
		t.Errorf("default L1 must contain MoAI version, got: %q", l1)
	}
	// Session time (9240000ms = 154min = 2h 34m)
	if !strings.Contains(l1, "⏳") {
		t.Errorf("default L1 must contain session time, got: %q", l1)
	}
	if !strings.Contains(l1, "2h 34m") {
		t.Errorf("default L1 must contain '2h 34m', got: %q", l1)
	}
}

func TestRenderDefaultV3_Line2_BarsInline(t *testing.T) {
	// default L2: CW/5H/7D bars inline on one line (10 blocks each, REQ-V3-API-011)
	r := newTestRenderer()
	data := &StatusData{
		Metrics:   MetricsData{Model: "Opus 4.6", Available: true},
		Memory:    MemoryData{TokensUsed: 176000, TokenBudget: 200000, Available: true},
		Directory: "moai-adk-go",
		Git:       GitStatusData{Branch: "main", Available: true},
		Usage: &UsageResult{
			Usage5H: &UsageData{UsedTokens: 45000, LimitTokens: 100000, Percentage: 45},
			Usage7D: &UsageData{UsedTokens: 82000, LimitTokens: 100000, Percentage: 82},
		},
	}

	got := r.Render(data, ModeDefault)
	lines := strings.Split(got, "\n")
	l2 := lines[1]

	// CW, 5H, 7D must all be on L2
	if !strings.Contains(l2, "CW:") {
		t.Errorf("default L2 must contain 'CW:' label, got: %q", l2)
	}
	if !strings.Contains(l2, "5H:") {
		t.Errorf("default L2 must contain '5H:' label, got: %q", l2)
	}
	if !strings.Contains(l2, "7D:") {
		t.Errorf("default L2 must contain '7D:' label, got: %q", l2)
	}
	// Verify percentage values
	if !strings.Contains(l2, "88%") {
		t.Errorf("default L2 must contain CW 88%%, got: %q", l2)
	}
	if !strings.Contains(l2, "45%") {
		t.Errorf("default L2 must contain 5H 45%%, got: %q", l2)
	}
	if !strings.Contains(l2, "82%") {
		t.Errorf("default L2 must contain 7D 82%%, got: %q", l2)
	}
}

func TestRenderDefaultV3_Line3(t *testing.T) {
	// default L3: directory, branch + ahead/behind, git status
	r := newTestRenderer()
	data := &StatusData{
		Metrics:   MetricsData{Model: "Opus 4.6", Available: true},
		Memory:    MemoryData{TokensUsed: 88000, TokenBudget: 200000, Available: true},
		Directory: "moai-adk-go",
		Git: GitStatusData{
			Branch:    "feat/auth",
			Ahead:     2,
			Behind:    1,
			Staged:    3,
			Modified:  2,
			Untracked: 1,
			Available: true,
		},
		Workspace: WorkspaceData{
			Repo: &RepoInfo{Host: "github.com", Owner: "modu-ai", Name: "moai-adk"},
		},
	}

	got := r.Render(data, ModeDefault)
	lines := strings.Split(got, "\n")
	l3 := lines[2]

	// Layout v3 amend: directory back on L3 head (before repo_branch).
	if !strings.Contains(l3, "📁 moai-adk-go") {
		t.Errorf("default L3 must contain directory at head, got: %q", l3)
	}
	// Layout v3 CH3 (2026-05-22 fix): combined repo+branch segment.
	// "🔀 owner/name | 🅱️ branch ↑N ↓N +N" (pipe separator, repo prefix required).
	// dirty = Staged(3) + Modified(2) + Untracked(1) = 6
	if !strings.Contains(l3, "🔀 modu-ai/moai-adk | 🅱️ feat/auth ↑2 ↓1 +6") {
		t.Errorf("default L3 must contain combined repo_branch segment with pipe separator, got: %q", l3)
	}
	if strings.Contains(l3, "📦") || strings.Contains(l3, "🔨") {
		t.Errorf("default L3 must not contain legacy 📦/🔨 prefix, got: %q", l3)
	}
}

func TestRenderDefaultV3_StyleInL1(t *testing.T) {
	// default: output style merged into L1 (L4 removed)
	r := newTestRenderer()
	data := &StatusData{
		Metrics:     MetricsData{Model: "Opus 4.6", Available: true},
		Memory:      MemoryData{TokensUsed: 88000, TokenBudget: 200000, Available: true},
		OutputStyle: "MoAI",
		Task:        TaskData{Active: true, Command: "run", SpecID: "SPEC-SLV3-001", Stage: "improve"},
		Directory:   "moai-adk-go",
		Git:         GitStatusData{Branch: "main", Available: true},
	}

	got := r.Render(data, ModeDefault)
	lines := strings.Split(got, "\n")

	// default must be 3 lines (L4 removed)
	if len(lines) != 3 {
		t.Fatalf("default mode must be 3 lines, got: %d lines\noutput: %q", len(lines), got)
	}

	// Output style must be merged into L1
	if !strings.Contains(lines[0], "💬 MoAI") {
		t.Errorf("default L1 must contain output style, got: %q", lines[0])
	}
	// Task (📋) is no longer displayed
}

// ─────────────────────────────────────────────────────────────────────────────
// Cycle 3: renderFullV3 — single layout-independent test (L1 prefix verification)
// Note: REQ-V3-LAYOUT-003 (full mode = 6-line layout) was retired per
//       renderer.go:50 — `mode` parameter accepted but collapses to default via
//       NormalizeMode. SPEC-V3R5-STATUSLINE-FULL-MODE-CLEANUP-001 deleted 5
//       retired-layout assertions (FiveLines, Lines2To4_SeparateBars,
//       Line5_DirBranchGit, StyleInL1, WithResetTimes). The remaining test
//       below (Line1_WithPrefixes) verifies L1 prefix content which is
//       layout-independent (same in both default 3-line and the retired
//       5-line layout).
// ─────────────────────────────────────────────────────────────────────────────

func TestRenderFullV3_Line1_WithPrefixes(t *testing.T) {
	// full L1: model, Claude version with "Claude" prefix, MoAI version with "MoAI" prefix, session time
	r := newTestRenderer()
	data := &StatusData{
		Metrics:           MetricsData{Model: "Opus 4.6", Available: true, SessionDurationMS: 9240000},
		Memory:            MemoryData{TokensUsed: 88000, TokenBudget: 200000, Available: true},
		ClaudeCodeVersion: "2.1.50",
		Version:           VersionData{Current: "2.8.0", Available: true},
		Directory:         "moai-adk-go",
		Git:               GitStatusData{Branch: "main", Available: true},
		Usage: &UsageResult{
			Usage5H: &UsageData{UsedTokens: 45000, LimitTokens: 100000, Percentage: 45},
			Usage7D: &UsageData{UsedTokens: 82000, LimitTokens: 100000, Percentage: 82},
		},
	}

	got := r.Render(data, ModeFull)
	lines := strings.Split(got, "\n")
	l1 := lines[0]

	if !strings.Contains(l1, "🤖 Opus 4.6") {
		t.Errorf("full L1 must contain model, got: %q", l1)
	}
	// full mode: no prefix, same as default
	if !strings.Contains(l1, "v2.1.50") {
		t.Errorf("full L1 should contain 'v2.1.50', got: %q", l1)
	}
	if !strings.Contains(l1, "🗿 v2.8.0") {
		t.Errorf("full L1 should contain 'moai v2.8.0', got: %q", l1)
	}
	if !strings.Contains(l1, "⏳") {
		t.Errorf("full L1 must contain session time, got: %q", l1)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Cycle 4: empty line omission tests
// REQ-V3-LAYOUT-004: omit lines where all segments are empty
// ─────────────────────────────────────────────────────────────────────────────

func TestRenderV3_OmitsEmptyLines_Default(t *testing.T) {
	// default: omit L4 when no task/output style → 3 lines
	r := newTestRenderer()
	data := &StatusData{
		Metrics:   MetricsData{Model: "Opus 4.6", Available: true},
		Memory:    MemoryData{TokensUsed: 88000, TokenBudget: 200000, Available: true},
		Directory: "moai-adk-go",
		Git:       GitStatusData{Branch: "main", Available: true},
		// No task, output style, or usage
	}

	got := r.Render(data, ModeDefault)
	lines := strings.Split(got, "\n")

	// L1(model+CW), L3(directory+branch) present
	// L2(CW only without 5H/7D → not omitted), L4(omitted without style/task)
	// Without usage: L1(model+CW), L2(CW bar only → not empty), L3(directory+branch), no L4 → 3 lines
	if len(lines) != 3 {
		t.Errorf("default mode without task/style must be 3 lines, got: %d lines\noutput:\n%s", len(lines), got)
	}
}

// TestRenderV3_OmitsEmptyLines_Full deleted per SPEC-V3R5-STATUSLINE-FULL-MODE-CLEANUP-001
// — asserted retired 5-line full layout (`len(lines) != 5` for ModeFull). Discovered
// in run-phase (iter 2 plan-auditor missed: grep pattern `TestRenderFullV3` did not
// match `TestRenderV3_*`). TestRenderV3_OmitsEmptyLines_Default preserved above as it
// verifies the current default 3-line behavior.

// ─────────────────────────────────────────────────────────────────────────────
// Cycle 5: Render() routing and backward compatibility tests
// ─────────────────────────────────────────────────────────────────────────────

func TestRender_ModeRouting(t *testing.T) {
	// All modes must route correctly
	r := newTestRenderer()
	data := &StatusData{
		Metrics: MetricsData{Model: "Opus 4.6", Available: true},
		Memory:  MemoryData{TokensUsed: 88000, TokenBudget: 200000, Available: true},
	}

	tests := []struct {
		mode     StatuslineMode
		minLines int
		maxLines int
	}{
		{ModeDefault, 1, 4},
		{ModeFull, 1, 6},
		{ModeCompact, 1, 4},  // deprecated → default
		{ModeMinimal, 1, 4},  // deprecated → default
		{ModeVerbose, 1, 6},  // deprecated → full
	}

	for _, tt := range tests {
		t.Run(string(tt.mode), func(t *testing.T) {
			got := r.Render(data, tt.mode)
			if got == "" || got == "MoAI" {
				// With data present, output should be actual render not MoAI fallback
				return
			}
			lines := strings.Split(got, "\n")
			if len(lines) < tt.minLines || len(lines) > tt.maxLines {
				t.Errorf("mode=%s: expected %d~%d lines but got %d lines\noutput: %q",
					tt.mode, tt.minLines, tt.maxLines, len(lines), got)
			}
		})
	}
}

func TestRenderUsageBar(t *testing.T) {
	// renderUsageBar helper function tests
	tests := []struct {
		name    string
		label   string
		pct     int
		width   int
		noColor bool
		wantPfx string // expected output prefix
	}{
		{
			name:    "CW 88% noColor",
			label:   "CW:",
			pct:     88,
			width:   10,
			noColor: true,
			wantPfx: "🪫 CW:",
		},
		{
			name:    "5H 45% noColor",
			label:   "5H:",
			pct:     45,
			width:   10,
			noColor: true,
			wantPfx: "🔋 5H:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := renderUsageBar(tt.label, tt.pct, tt.width, tt.noColor)
			if !strings.HasPrefix(got, tt.wantPfx) {
				t.Errorf("renderUsageBar() = %q, wantPfx %q", got, tt.wantPfx)
			}
			if !strings.Contains(got, fmt.Sprintf("%d%%", tt.pct)) {
				t.Errorf("renderUsageBar() must contain %d%%, got: %q", tt.pct, got)
			}
		})
	}
}

func TestRender_V3Separator(t *testing.T) {
	// v3 separator must be " │ " (U+2502 box-drawing character)
	r := newTestRenderer()
	data := &StatusData{
		Metrics:   MetricsData{Model: "Opus 4.6", Available: true},
		Memory:    MemoryData{TokensUsed: 88000, TokenBudget: 200000, Available: true},
		Directory: "moai-adk-go",
		Git:       GitStatusData{Branch: "main", Available: true},
	}

	got := r.Render(data, ModeDefault)

	// Verify v3 separator
	if !strings.Contains(got, " │ ") {
		t.Errorf("v3 separator ' │ ' must be present, got: %q", got)
	}
}

func TestFormatResetTimeRelative(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		input     string
		wantSub   string // expected substring
		wantEmpty bool
	}{
		{"empty string", "", "rolling", false},
		{"invalid format", "not-a-date", "rolling", false},
		{"future 2h30m", time.Now().Add(2*time.Hour + 30*time.Minute).UTC().Format(time.RFC3339), "2h", false},
		{"future 45m", time.Now().Add(45*time.Minute + 5*time.Second).UTC().Format(time.RFC3339), "m", false},
		{"past time", time.Now().Add(-1 * time.Hour).UTC().Format(time.RFC3339), "rolling", false},
		{"7D: future 1d3h30m", time.Now().Add(27*time.Hour + 30*time.Minute + 5*time.Second).UTC().Format(time.RFC3339), "1D", false},
		{"7D: future 2d only", time.Now().Add(48*time.Hour + 5*time.Minute).UTC().Format(time.RFC3339), "2D", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatResetTimeRelative(tt.input)
			if tt.wantEmpty && got != "" {
				t.Errorf("expected empty, got %q", got)
			}
			if !tt.wantEmpty && !strings.Contains(got, tt.wantSub) {
				t.Errorf("expected substring %q, got %q", tt.wantSub, got)
			}
		})
	}
}

func TestFormatResetTimeAbsolute(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		input     string
		wantEmpty bool
		wantSub   string // substring to check
	}{
		{"empty string", "", false, ""},
		{"invalid format", "not-a-date", false, ""},
		{"valid RFC3339", "2026-01-21T14:00:00Z", false, "Jan"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatResetTimeAbsolute(tt.input)
			if tt.wantEmpty && got != "" {
				t.Errorf("expected empty, got %q", got)
			}
			if !tt.wantEmpty && !strings.Contains(got, tt.wantSub) {
				t.Errorf("expected substring %q in %q", tt.wantSub, got)
			}
		})
	}
}

func TestRenderUsageBarWithReset(t *testing.T) {
	t.Parallel()

	// Without reset
	got := renderUsageBarWithReset("5H:", 25, 10, true, "")
	if strings.Contains(got, "(") {
		t.Errorf("should not contain parentheses when empty, got %q", got)
	}

	// With reset
	got = renderUsageBarWithReset("5H:", 25, 10, true, "2h 15m")
	if !strings.Contains(got, "(2h 15m)") {
		t.Errorf("should contain '(2h 15m)', got %q", got)
	}
}

// TestRenderDirGitLine_WorktreeIndicator verifies that the "[WT] " prefix appears in
// the branch segment when a worktree is active and SegmentWorktree is enabled.
// REQ-CC297-003: worktree prefix shown in branch segment when worktree is active.
// The legacy 🌿 emoji was replaced with the textual "[WT] " marker for clarity.
func TestRenderDirGitLine_WorktreeIndicator(t *testing.T) {
	tests := []struct {
		name           string
		worktree       string
		segmentEnabled bool
		wantWT         bool
	}{
		{
			name:           "worktree present and segment enabled: shows [WT]",
			worktree:       "/repo/.claude/worktrees/abc123",
			segmentEnabled: true,
			wantWT:         true,
		},
		{
			name:           "worktree present but segment disabled: no [WT]",
			worktree:       "/repo/.claude/worktrees/abc123",
			segmentEnabled: false,
			wantWT:         false,
		},
		{
			name:           "no worktree with segment enabled: no [WT]",
			worktree:       "",
			segmentEnabled: true,
			wantWT:         false,
		},
		{
			name:           "no worktree and segment disabled: no [WT]",
			worktree:       "",
			segmentEnabled: false,
			wantWT:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			segCfg := map[string]bool{
				SegmentGitBranch: true,
				SegmentDirectory: true,
				SegmentGitStatus: false,
				SegmentWorktree:  tt.segmentEnabled,
			}
			r := NewRenderer("default", true, segCfg)
			data := &StatusData{
				Git:       GitStatusData{Branch: "feat/test", Available: true},
				Directory: "myproject",
				Worktree:  tt.worktree,
				Workspace: WorkspaceData{
					Repo: &RepoInfo{Host: "github.com", Owner: "modu-ai", Name: "moai-adk"},
				},
			}
			got := r.renderDirGitLine(data)
			// Legacy emoji must never appear regardless of state.
			if strings.Contains(got, "🌿") {
				t.Errorf("legacy 🌿 emoji must not appear, got %q", got)
			}
			if tt.wantWT && !strings.Contains(got, "[WT] ") {
				t.Errorf("expected [WT] prefix when worktree is active, got %q", got)
			}
			if !tt.wantWT && strings.Contains(got, "[WT]") {
				t.Errorf("unexpected [WT] prefix when worktree is inactive, got %q", got)
			}
		})
	}
}

// TestRenderer_EffortThinking_InfoLine verifies renderInfoLine integrates the
// effort/thinking indicator according to GWT-1 through GWT-6 acceptance criteria.
//
// GWT-1: effort="high", thinking=false  → "🧠 high" present, "·t" absent
// GWT-2: effort="max",  thinking=true   → "🧠 max·t" present
// GWT-3: effort absent, thinking=true   → "·t" present without "🧠" prefix
// GWT-4: both absent                    → neither "🧠" nor "·t" present
// GWT-5: effort="", thinking=false      → silent omit (REQ-CC2122-003)
// GWT-6: effort="ultra", thinking=false → "🧠 ultra" (unknown level passthrough, REQ-CC2122-004)
func TestRenderer_EffortThinking_InfoLine(t *testing.T) {
	tests := []struct {
		name         string
		effort       *EffortInfo
		thinking     *ThinkingInfo
		wantContains []string
		wantAbsent   []string
	}{
		{
			// GWT-1
			name:         "GWT-1: effort=high, thinking=false → 🧠 high present, ·t absent",
			effort:       &EffortInfo{Level: "high"},
			thinking:     &ThinkingInfo{Enabled: false},
			wantContains: []string{"🧠 high"},
			wantAbsent:   []string{"·t"},
		},
		{
			// GWT-2
			name:         "GWT-2: effort=max, thinking=true → 🧠 max 🧠 max·t present",
			effort:       &EffortInfo{Level: "max"},
			thinking:     &ThinkingInfo{Enabled: true},
			wantContains: []string{"🧠 max·t"},
			wantAbsent:   []string{},
		},
		{
			// GWT-3: thinking-only produces ·t without 🧠 prefix
			name:         "GWT-3: effort absent, thinking=true → ·t only",
			effort:       nil,
			thinking:     &ThinkingInfo{Enabled: true},
			wantContains: []string{"·t"},
			wantAbsent:   []string{"🧠"},
		},
		{
			// GWT-4
			name:         "GWT-4: both absent → neither 🧠 nor ·t",
			effort:       nil,
			thinking:     nil,
			wantContains: nil,
			wantAbsent:   []string{"🧠", "·t"},
		},
		{
			// GWT-5: empty string treated same as absent
			name:         "GWT-5: effort empty string → silent omit",
			effort:       &EffortInfo{Level: ""},
			thinking:     nil,
			wantContains: nil,
			wantAbsent:   []string{"🧠", "·t"},
		},
		{
			// GWT-6: unknown level passed through without filtering
			name:         "GWT-6: unknown level ultra → 🧠 ultra (raw passthrough)",
			effort:       &EffortInfo{Level: "ultra"},
			thinking:     &ThinkingInfo{Enabled: false},
			wantContains: []string{"🧠 ultra"},
			wantAbsent:   []string{"·t"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := newTestRenderer()
			data := &StatusData{
				Metrics:  MetricsData{Model: "Opus 4.5", Available: true},
				Memory:   MemoryData{TokensUsed: 50000, TokenBudget: 200000, Available: true},
				Effort:   tt.effort,
				Thinking: tt.thinking,
			}
			got := r.renderInfoLine(data, false)

			for _, want := range tt.wantContains {
				if !strings.Contains(got, want) {
					t.Errorf("want %q in renderInfoLine output, got %q", want, got)
				}
			}
			for _, absent := range tt.wantAbsent {
				if strings.Contains(got, absent) {
					t.Errorf("want %q absent from renderInfoLine output, got %q", absent, got)
				}
			}
		})
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// SPEC-V3R5-STATUSLINE-V2145-001 — M2 PR segment renderer tests
// ─────────────────────────────────────────────────────────────────────────────

// TestRenderPRSegment_Format verifies the rendered shape is "#<number> ⌥<state>"
// per REQ-SLV-013. The exact unicode glyph U+2325 (⌥) is the review-state marker.
// AC-SLV-013 verification target.
func TestRenderPRSegment_Format(t *testing.T) {
	tests := []struct {
		name string
		pr   *PRInfo
		want string
	}{
		{
			name: "approved pr 1023",
			pr:   &PRInfo{Number: 1023, URL: "https://x/pull/1023", ReviewState: "approved"},
			want: "💌 PR #1023 (⌥approved)",
		},
		{
			name: "pending pr 42",
			pr:   &PRInfo{Number: 42, URL: "https://x/pull/42", ReviewState: "pending"},
			want: "💌 PR #42 (⌥pending)",
		},
		{
			name: "changes_requested pr 7",
			pr:   &PRInfo{Number: 7, URL: "https://x/pull/7", ReviewState: "changes_requested"},
			want: "💌 PR #7 (⌥changes_requested)",
		},
		{
			name: "draft pr 99",
			pr:   &PRInfo{Number: 99, URL: "https://x/pull/99", ReviewState: "draft"},
			want: "💌 PR #99 (⌥draft)",
		},
		{
			name: "empty url tolerated",
			pr:   &PRInfo{Number: 50, URL: "", ReviewState: "approved"},
			want: "💌 PR #50 (⌥approved)",
		},
		{
			name: "empty review_state: no marker",
			pr:   &PRInfo{Number: 100, URL: "https://x/pull/100", ReviewState: ""},
			want: "💌 PR #100",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// noColor renderer for predictable byte-level comparison
			r := NewRenderer("default", true, map[string]bool{SegmentPR: true})
			data := &StatusData{PR: tt.pr}
			got := r.renderPRSegment(data)
			if got != tt.want {
				t.Errorf("renderPRSegment() = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestRenderPRSegment_Absence verifies that no segment is emitted when PR is
// nil or Number == 0, OR when SegmentPR is disabled.
// REQ-SLV-015 + REQ-SLV-012.
// AC-SLV-015 verification target.
func TestRenderPRSegment_Absence(t *testing.T) {
	tests := []struct {
		name          string
		pr            *PRInfo
		segmentConfig map[string]bool
	}{
		{
			name:          "pr nil + enabled",
			pr:            nil,
			segmentConfig: map[string]bool{SegmentPR: true},
		},
		{
			name:          "pr Number zero + enabled",
			pr:            &PRInfo{Number: 0, URL: "https://x", ReviewState: "approved"},
			segmentConfig: map[string]bool{SegmentPR: true},
		},
		{
			name:          "pr present + disabled",
			pr:            &PRInfo{Number: 1023, URL: "https://x", ReviewState: "approved"},
			segmentConfig: map[string]bool{SegmentPR: false},
		},
		{
			name:          "pr present + unset (legacy backward compat)",
			pr:            &PRInfo{Number: 1023, URL: "https://x", ReviewState: "approved"},
			segmentConfig: map[string]bool{SegmentModel: true}, // pr key missing
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewRenderer("default", true, tt.segmentConfig)
			data := &StatusData{PR: tt.pr}
			got := r.renderPRSegment(data)
			if got != "" {
				t.Errorf("renderPRSegment() = %q, want empty string", got)
			}
		})
	}
}

// TestRenderPRSegment_ZeroNumber verifies that pr.Number == 0 is treated as
// "PR absent" per REQ-SLV-015 ("no placeholder text such as #N/A or #0").
// AC-SLV-015 specifically requires (a) pr == nil AND (b) pr.Number == 0 cases.
func TestRenderPRSegment_ZeroNumber(t *testing.T) {
	r := NewRenderer("default", true, map[string]bool{SegmentPR: true})
	data := &StatusData{PR: &PRInfo{Number: 0, URL: "https://x", ReviewState: "approved"}}
	got := r.renderPRSegment(data)
	if got != "" {
		t.Errorf("renderPRSegment() with Number=0 = %q, want empty", got)
	}
}

// TestRenderPRSegment_Nil verifies that a nil PR pointer is safely handled.
// AC-SLV-015 specifically requires this case.
func TestRenderPRSegment_Nil(t *testing.T) {
	r := NewRenderer("default", true, map[string]bool{SegmentPR: true})
	data := &StatusData{PR: nil}
	got := r.renderPRSegment(data)
	if got != "" {
		t.Errorf("renderPRSegment() with nil PR = %q, want empty", got)
	}
}

// TestPRReviewStateColor verifies the review_state → ANSI color mapping
// per plan.md D3 / REQ-SLV-014 table.
// AC-SLV-014 verification target.
//
// Forces lipgloss to TrueColor profile so the test runs deterministically
// regardless of TTY detection in the test environment.
func TestPRReviewStateColor(t *testing.T) {
	// Force ANSI emission in non-TTY test environment
	prevProfile := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(termenv.TrueColor)
	t.Cleanup(func() { lipgloss.SetColorProfile(prevProfile) })

	tests := []struct {
		name        string
		reviewState string
		wantStyled  bool // expect ANSI escape codes in output
	}{
		// All known review_state values should produce styled (ANSI) output
		{name: "approved → green styled", reviewState: "approved", wantStyled: true},
		{name: "pending → yellow styled", reviewState: "pending", wantStyled: true},
		{name: "changes_requested → red styled", reviewState: "changes_requested", wantStyled: true},
		{name: "draft → gray/muted styled", reviewState: "draft", wantStyled: true},
		// Unknown / empty: raw passthrough — segment renders but no color escape
		{name: "unknown state: raw passthrough no color", reviewState: "merged", wantStyled: false},
		{name: "empty state: no marker no color", reviewState: "", wantStyled: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Color-enabled renderer
			r := NewRenderer("default", false, map[string]bool{SegmentPR: true})
			data := &StatusData{
				PR: &PRInfo{Number: 1, URL: "https://x", ReviewState: tt.reviewState},
			}
			got := r.renderPRSegment(data)

			// Always contains "#1"
			if !strings.Contains(got, "#1") {
				t.Fatalf("output should contain #1, got %q", got)
			}

			// ANSI escape code presence check (lipgloss emits \x1b[ escape sequences)
			hasANSI := strings.Contains(got, "\x1b[")
			if tt.wantStyled && !hasANSI {
				t.Errorf("expected ANSI styling for %q, got plain output: %q", tt.reviewState, got)
			}
			if !tt.wantStyled && hasANSI {
				t.Errorf("expected no ANSI styling for %q, got styled output: %q", tt.reviewState, got)
			}
		})
	}
}

// TestPRReviewStateColor_NoColor verifies that NoColor=true suppresses all
// ANSI codes even when the global lipgloss profile is set to TrueColor.
// REQ-SLV-014 + Options.NoColor invariant.
func TestPRReviewStateColor_NoColor(t *testing.T) {
	// Force TrueColor profile globally to prove that renderer-level NoColor=true
	// suppresses ANSI emission independently from lipgloss profile detection.
	prevProfile := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(termenv.TrueColor)
	t.Cleanup(func() { lipgloss.SetColorProfile(prevProfile) })

	states := []string{"approved", "pending", "changes_requested", "draft", "merged", ""}
	for _, state := range states {
		t.Run(state, func(t *testing.T) {
			r := NewRenderer("default", true, map[string]bool{SegmentPR: true})
			data := &StatusData{PR: &PRInfo{Number: 1, URL: "https://x", ReviewState: state}}
			got := r.renderPRSegment(data)
			if strings.Contains(got, "\x1b[") {
				t.Errorf("NoColor=true must not emit ANSI escapes, got %q", got)
			}
		})
	}
}

// TestRenderDirGitLine_PRSegment verifies that the PR segment integrates with
// the L3 directory/git/PR line composition when SegmentPR is enabled.
// REQ-SLV-016: PR segment is composed alongside directory/branch/git_status.
// AC-SLV-013 + REQ-SLV-013: "#<number> ⌥<state>" appears in renderDirGitLine output.
func TestRenderDirGitLine_PRSegment(t *testing.T) {
	tests := []struct {
		name          string
		pr            *PRInfo
		segmentConfig map[string]bool
		wantContains  []string
		wantAbsent    []string
	}{
		{
			name: "pr enabled + approved: segment appears on L3",
			pr:   &PRInfo{Number: 1023, URL: "https://x", ReviewState: "approved"},
			segmentConfig: map[string]bool{
				SegmentDirectory: true,
				SegmentGitBranch: true,
				SegmentGitStatus: true,
				SegmentPR:        true,
			},
			wantContains: []string{"#1023", "⌥approved"},
		},
		{
			name: "pr disabled: segment absent from L3",
			pr:   &PRInfo{Number: 1023, URL: "https://x", ReviewState: "approved"},
			segmentConfig: map[string]bool{
				SegmentDirectory: true,
				SegmentGitBranch: true,
				SegmentPR:        false,
			},
			wantContains: []string{"🔀"},
			wantAbsent:   []string{"#1023", "⌥"},
		},
		{
			name: "pr nil: segment absent from L3",
			pr:   nil,
			segmentConfig: map[string]bool{
				SegmentDirectory: true,
				SegmentGitBranch: true,
				SegmentPR:        true,
			},
			wantContains: []string{"🔀"},
			wantAbsent:   []string{"#", "⌥"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewRenderer("default", true, tt.segmentConfig)
			data := &StatusData{
				Git:       GitStatusData{Branch: "feat/x", Available: true},
				Directory: "myproject",
				PR:        tt.pr,
				Workspace: WorkspaceData{
					Repo: &RepoInfo{Host: "github.com", Owner: "modu-ai", Name: "moai-adk"},
				},
			}
			got := r.renderDirGitLine(data)
			for _, want := range tt.wantContains {
				if !strings.Contains(got, want) {
					t.Errorf("expected %q in renderDirGitLine output, got %q", want, got)
				}
			}
			for _, absent := range tt.wantAbsent {
				if strings.Contains(got, absent) {
					t.Errorf("expected %q absent from renderDirGitLine output, got %q", absent, got)
				}
			}
		})
	}
}
