package statusline

import (
	"fmt"
	"strings"
	"testing"
	"time"
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
	if !strings.Contains(got, "📊") {
		t.Errorf("default mode should contain git status with emoji, got %q", got)
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
		t.Errorf("default mode should contain MoAI version with 🗿 emoji, got %q", got)
	}
	if !strings.Contains(got, "🔀 main") {
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
	if !strings.Contains(lines[0], "🔅 v1.0.80") {
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
		if strings.Contains(line, "📁") || strings.Contains(line, "🔀") || strings.Contains(line, "📊") {
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
	if !strings.Contains(got, "🔅 v1.0.80") {
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
	if strings.Contains(got, "📊") {
		t.Errorf("should not contain git status emoji when unavailable, got %q", got)
	}
	if strings.Contains(got, "🔀") {
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
	}

	got := r.Render(data, ModeDefault)

	if !strings.Contains(got, "🔀 main") {
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
	}

	got := r.Render(data, ModeDefault)

	if !strings.Contains(got, "🔀 main") {
		t.Errorf("should show branch name, got %q", got)
	}
	// Should not have git status emoji when all counts are zero
	if strings.Contains(got, "📊") {
		t.Errorf("should not show git status when all counts are zero, got %q", got)
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
	if !strings.Contains(got, "⬆️ v2.0.1") {
		t.Errorf("should contain update notification with emoji, got %q", got)
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
			wantContain:   []string{"🤖 Opus 4.5", "🔋", "💬 MoAI", "📁 moai-adk-go", "📊", "🔅 v1.0.80", "🗿 v2.3.1", "🔀 main"},
		},
		{
			name:          "empty config shows all segments",
			segmentConfig: map[string]bool{},
			wantContain:   []string{"🤖 Opus 4.5", "🔋", "💬 MoAI", "📁 moai-adk-go", "📊", "🔅 v1.0.80", "🗿 v2.3.1", "🔀 main"},
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
			wantContain:    []string{"🤖 Opus 4.5", "🔋", "📊", "🔀 main"},
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
			wantContain:    []string{"🤖 Opus 4.5", "💬 MoAI", "📁 moai-adk-go", "🔀 main"},
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
		name string
		git  GitStatusData
		want string
	}{
		// REQ-V3-GIT-001: Ahead only → "🔀 main ↑3"
		{"ahead only", GitStatusData{Branch: "main", Ahead: 3, Behind: 0, Available: true}, "🔀 main ↑3"},
		// REQ-V3-GIT-002: Behind only → "🔀 main ↓2"
		{"behind only", GitStatusData{Branch: "main", Ahead: 0, Behind: 2, Available: true}, "🔀 main ↓2"},
		// REQ-V3-GIT-003: Both present → "🔀 feat/auth ↑2↓1"
		{"both", GitStatusData{Branch: "feat/auth", Ahead: 2, Behind: 1, Available: true}, "🔀 feat/auth ↑2↓1"},
		// REQ-V3-GIT-004: Neither present → branch name only
		{"neither", GitStatusData{Branch: "main", Ahead: 0, Behind: 0, Available: true}, "🔀 main"},
		// Unavailable → empty string
		{"unavailable", GitStatusData{Available: false}, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := &StatusData{Git: tt.git}
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
	if !strings.Contains(l1, "🔅 v2.1.50") {
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
	}

	got := r.Render(data, ModeDefault)
	lines := strings.Split(got, "\n")
	l3 := lines[2]

	if !strings.Contains(l3, "📁 moai-adk-go") {
		t.Errorf("default L3 must contain directory, got: %q", l3)
	}
	if !strings.Contains(l3, "🔀 feat/auth ↑2↓1") {
		t.Errorf("default L3 must contain branch + ahead/behind, got: %q", l3)
	}
	if !strings.Contains(l3, "📊") {
		t.Errorf("default L3 must contain git status, got: %q", l3)
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
// Cycle 3: renderFullV3 tests
// REQ-V3-LAYOUT-003: full mode uses 6-line layout
// ─────────────────────────────────────────────────────────────────────────────

func TestRenderFullV3_FiveLines(t *testing.T) {
	// full mode must produce exactly 5 lines (L6 removed, style merged into L1)
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

	got := r.Render(data, ModeFull)
	lines := strings.Split(got, "\n")

	if len(lines) != 5 {
		t.Errorf("full mode must be 5 lines, got: %d lines\noutput:\n%s", len(lines), got)
	}
	// Output style must be merged into L1
	if !strings.Contains(lines[0], "💬 MoAI") {
		t.Errorf("full L1 must contain output style, got: %q", lines[0])
	}
}

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
	if !strings.Contains(l1, "🔅 v2.1.50") {
		t.Errorf("full L1 should contain '🔅 v2.1.50', got: %q", l1)
	}
	if !strings.Contains(l1, "🗿 v2.8.0") {
		t.Errorf("full L1 should contain '🗿 v2.8.0', got: %q", l1)
	}
	if !strings.Contains(l1, "⏳") {
		t.Errorf("full L1 must contain session time, got: %q", l1)
	}
}

func TestRenderFullV3_Lines2To4_SeparateBars(t *testing.T) {
	// full L2-L4: each bar on separate line (40 blocks, REQ-V3-API-011)
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

	got := r.Render(data, ModeFull)
	lines := strings.Split(got, "\n")

	// L2: CW bar only
	l2 := lines[1]
	if !strings.Contains(l2, "CW:") {
		t.Errorf("full L2 must contain 'CW:' label, got: %q", l2)
	}
	if strings.Contains(l2, "5H:") || strings.Contains(l2, "7D:") {
		t.Errorf("full L2 must not contain 5H/7D (CW only), got: %q", l2)
	}

	// L3: 5H bar only
	l3 := lines[2]
	if !strings.Contains(l3, "5H:") {
		t.Errorf("full L3 must contain '5H:' label, got: %q", l3)
	}
	if strings.Contains(l3, "CW:") || strings.Contains(l3, "7D:") {
		t.Errorf("full L3 must not contain CW/7D (5H only), got: %q", l3)
	}

	// L4: 7D bar only
	l4 := lines[3]
	if !strings.Contains(l4, "7D:") {
		t.Errorf("full L4 must contain '7D:' label, got: %q", l4)
	}
	if strings.Contains(l4, "CW:") || strings.Contains(l4, "5H:") {
		t.Errorf("full L4 must not contain CW/5H (7D only), got: %q", l4)
	}
}

func TestRenderFullV3_Line5_DirBranchGit(t *testing.T) {
	// full L5: directory, branch + ahead/behind, git status
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
		Usage: &UsageResult{
			Usage5H: &UsageData{UsedTokens: 45000, LimitTokens: 100000, Percentage: 45},
			Usage7D: &UsageData{UsedTokens: 82000, LimitTokens: 100000, Percentage: 82},
		},
	}

	got := r.Render(data, ModeFull)
	lines := strings.Split(got, "\n")

	// Verify L5
	if len(lines) < 5 {
		t.Fatalf("full mode must have L5, got: %d lines\noutput:\n%s", len(lines), got)
	}
	l5 := lines[4]

	if !strings.Contains(l5, "📁 moai-adk-go") {
		t.Errorf("full L5 must contain directory, got: %q", l5)
	}
	if !strings.Contains(l5, "🔀 feat/auth ↑2↓1") {
		t.Errorf("full L5 must contain branch + ahead/behind, got: %q", l5)
	}
	if !strings.Contains(l5, "📊") {
		t.Errorf("full L5 must contain git status, got: %q", l5)
	}
}

func TestRenderFullV3_StyleInL1(t *testing.T) {
	// full: output style merged into L1 (L6 removed)
	r := newTestRenderer()
	data := &StatusData{
		Metrics:     MetricsData{Model: "Opus 4.6", Available: true},
		Memory:      MemoryData{TokensUsed: 88000, TokenBudget: 200000, Available: true},
		OutputStyle: "MoAI",
		Task:        TaskData{Active: true, Command: "run", SpecID: "SPEC-SLV3-001", Stage: "improve"},
		Directory:   "moai-adk-go",
		Git:         GitStatusData{Branch: "main", Available: true},
		Usage: &UsageResult{
			Usage5H: &UsageData{UsedTokens: 45000, LimitTokens: 100000, Percentage: 45},
			Usage7D: &UsageData{UsedTokens: 82000, LimitTokens: 100000, Percentage: 82},
		},
	}

	got := r.Render(data, ModeFull)
	lines := strings.Split(got, "\n")

	// full must be 5 lines (L6 removed)
	if len(lines) != 5 {
		t.Fatalf("full mode must be 5 lines, got: %d lines\noutput:\n%s", len(lines), got)
	}

	// Output style must be merged into L1
	if !strings.Contains(lines[0], "💬 MoAI") {
		t.Errorf("full L1 must contain output style, got: %q", lines[0])
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

func TestRenderV3_OmitsEmptyLines_Full(t *testing.T) {
	// full: omit L6 when no task/output style → 5 lines
	r := newTestRenderer()
	data := &StatusData{
		Metrics:   MetricsData{Model: "Opus 4.6", Available: true},
		Memory:    MemoryData{TokensUsed: 88000, TokenBudget: 200000, Available: true},
		Directory: "moai-adk-go",
		Git:       GitStatusData{Branch: "main", Available: true},
		Usage: &UsageResult{
			Usage5H: &UsageData{UsedTokens: 45000, LimitTokens: 100000, Percentage: 45},
			Usage7D: &UsageData{UsedTokens: 82000, LimitTokens: 100000, Percentage: 82},
		},
		// No task or output style
	}

	got := r.Render(data, ModeFull)
	lines := strings.Split(got, "\n")

	if len(lines) != 5 {
		t.Errorf("full mode without task/style must be 5 lines, got: %d lines\noutput:\n%s", len(lines), got)
	}
}

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
			wantPfx: "CW: 🪫",
		},
		{
			name:    "5H 45% noColor",
			label:   "5H:",
			pct:     45,
			width:   10,
			noColor: true,
			wantPfx: "5H: 🔋",
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
		wantPfx   string // expected prefix (e.g. "in ")
		wantEmpty bool
	}{
		{"empty string", "", "", true},
		{"invalid format", "not-a-date", "", true},
		{"future 2h30m", time.Now().Add(2*time.Hour + 30*time.Minute).UTC().Format(time.RFC3339), "in 2h", false},
		{"future 45m", time.Now().Add(45 * time.Minute).UTC().Format(time.RFC3339), "in 4", false},
		{"past time", time.Now().Add(-1 * time.Hour).UTC().Format(time.RFC3339), "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatResetTimeRelative(tt.input)
			if tt.wantEmpty && got != "" {
				t.Errorf("expected empty, got %q", got)
			}
			if !tt.wantEmpty && !strings.HasPrefix(got, tt.wantPfx) {
				t.Errorf("expected prefix %q, got %q", tt.wantPfx, got)
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
		{"empty string", "", true, ""},
		{"invalid format", "not-a-date", true, ""},
		{"valid RFC3339", "2026-01-21T14:00:00Z", false, "at "},
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

func TestRenderFullV3_WithResetTimes(t *testing.T) {
	t.Parallel()

	r := newTestRenderer()
	resetTime := time.Now().Add(2*time.Hour + 15*time.Minute).UTC().Format(time.RFC3339)

	data := &StatusData{
		Metrics:   MetricsData{Model: "Opus 4.6", Available: true},
		Memory:    MemoryData{TokensUsed: 100000, TokenBudget: 200000, Available: true},
		Directory: "test-project",
		Git:       GitStatusData{Branch: "main", Available: true},
		Usage: &UsageResult{
			Usage5H: &UsageData{Percentage: 25.0, ResetsAt: resetTime},
			Usage7D: &UsageData{Percentage: 60.0, ResetsAt: "2026-01-21T14:00:00Z"},
		},
	}

	got := r.Render(data, ModeFull)
	lines := strings.Split(got, "\n")

	if len(lines) < 5 {
		t.Fatalf("full mode should have 5 lines, got %d: %q", len(lines), got)
	}

	// L3 (5H) should contain "Resets in"
	if !strings.Contains(lines[2], "Resets in") {
		t.Errorf("5H line should contain 'Resets in', got: %q", lines[2])
	}

	// L4 (7D) should contain "Resets" with a date
	if !strings.Contains(lines[3], "Resets") {
		t.Errorf("7D line should contain 'Resets', got: %q", lines[3])
	}
}

func TestRenderUsageBarWithReset(t *testing.T) {
	t.Parallel()

	// Without reset
	got := renderUsageBarWithReset("5H:", 25, 10, true, "")
	if strings.Contains(got, "Resets") {
		t.Errorf("should not contain Resets when empty, got %q", got)
	}

	// With reset
	got = renderUsageBarWithReset("5H:", 25, 10, true, "in 2h15m")
	if !strings.Contains(got, "(Resets in 2h15m)") {
		t.Errorf("should contain '(Resets in 2h15m)', got %q", got)
	}
}
