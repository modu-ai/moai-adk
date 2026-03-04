package statusline

import (
	"strings"
	"testing"
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

func TestRender_CompactMode(t *testing.T) {
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
		t.Errorf("compact mode should contain model with emoji, got %q", got)
	}
	if !strings.Contains(got, "🔋") {
		t.Errorf("compact mode should contain battery emoji, got %q", got)
	}
	if !strings.Contains(got, "💬 Mr.Alfred") {
		t.Errorf("compact mode should contain output style with emoji, got %q", got)
	}
	if !strings.Contains(got, "📁 moai-adk-go") {
		t.Errorf("compact mode should contain directory with emoji, got %q", got)
	}
	if !strings.Contains(got, "📊") {
		t.Errorf("compact mode should contain git status with emoji, got %q", got)
	}
	if !strings.Contains(got, "+3") {
		t.Errorf("compact mode should contain staged count, got %q", got)
	}
	if !strings.Contains(got, "M2") {
		t.Errorf("compact mode should contain modified count with 'M', got %q", got)
	}
	if !strings.Contains(got, "?1") {
		t.Errorf("compact mode should contain untracked count, got %q", got)
	}
	if !strings.Contains(got, "🗿 v1.14.5") {
		t.Errorf("compact mode should contain MoAI version with 🗿 emoji, got %q", got)
	}
	if !strings.Contains(got, "🔀 main") {
		t.Errorf("compact mode should contain branch with emoji, got %q", got)
	}
}

func TestRender_VerboseMode_MultiLine(t *testing.T) {
	// REQ-SLE-033: verbose mode renders up to 3 newline-separated lines
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

	// Verbose output must contain newlines
	if !strings.Contains(got, "\n") {
		t.Errorf("verbose mode should produce multi-line output, got %q", got)
	}

	lines := strings.Split(got, "\n")

	// Line 1: Model | Context Graph | Output Style
	if !strings.Contains(lines[0], "🤖 Sonnet 4") {
		t.Errorf("verbose line 1 should contain model, got %q", lines[0])
	}
	if !strings.Contains(lines[0], "🔋") {
		t.Errorf("verbose line 1 should contain context, got %q", lines[0])
	}
	if !strings.Contains(lines[0], "💬 Yoda") {
		t.Errorf("verbose line 1 should contain output style, got %q", lines[0])
	}
	// Line 1 should NOT contain git branch or directory
	if strings.Contains(lines[0], "📁") {
		t.Errorf("verbose line 1 should NOT contain directory, got %q", lines[0])
	}

	// Line 2: Directory | Branch | Git Changes
	if len(lines) < 2 {
		t.Fatal("verbose mode should have at least 2 lines")
	}
	if !strings.Contains(lines[1], "📁 my-project") {
		t.Errorf("verbose line 2 should contain directory, got %q", lines[1])
	}
	if !strings.Contains(lines[1], "🔀 main") {
		t.Errorf("verbose line 2 should contain branch, got %q", lines[1])
	}
	if !strings.Contains(lines[1], "📊") {
		t.Errorf("verbose line 2 should contain git status, got %q", lines[1])
	}

	// Line 3: Claude Version | MoAI Version | Cost
	if len(lines) < 3 {
		t.Fatal("verbose mode should have 3 lines when all data available")
	}
	if !strings.Contains(lines[2], "🔅 v1.0.80") {
		t.Errorf("verbose line 3 should contain Claude version, got %q", lines[2])
	}
	if !strings.Contains(lines[2], "🗿 v1.2.0") {
		t.Errorf("verbose line 3 should contain MoAI version, got %q", lines[2])
	}
	if !strings.Contains(lines[2], "$0.42") {
		t.Errorf("verbose line 3 should contain cost, got %q", lines[2])
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
	// REQ-SLE-035: CostUSD rendered on line 3 as "$X.XX"
	r := newTestRenderer()
	data := &StatusData{
		Metrics:           MetricsData{Model: "Opus 4.5", CostUSD: 1.23, Available: true},
		Memory:            MemoryData{TokensUsed: 50000, TokenBudget: 200000, Available: true},
		Version:           VersionData{Current: "2.0.0", Available: true},
		ClaudeCodeVersion: "1.0.80",
	}

	got := r.Render(data, ModeVerbose)
	if !strings.Contains(got, "$1.23") {
		t.Errorf("verbose mode should render cost as $1.23, got %q", got)
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
	// REQ-SLE-031: ModeMinimal = 1 line: Model | Context Graph | Output Style
	r := newTestRenderer()
	data := &StatusData{
		Metrics:     MetricsData{Model: "Sonnet 4", Available: true},
		Memory:      MemoryData{TokensUsed: 50000, TokenBudget: 200000, Available: true},
		OutputStyle: "MoAI",
	}

	got := r.Render(data, ModeMinimal)

	// Minimal mode should not contain newlines
	if strings.Contains(got, "\n") {
		t.Errorf("minimal mode should be single line, got %q", got)
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
	if strings.Contains(got, "🔋") {
		t.Errorf("should not contain battery emoji when memory unavailable, got %q", got)
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
	if got != "MoAI" {
		t.Errorf("empty data should return fallback, got %q", got)
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

	got := r.Render(data, ModeDefault)

	if !strings.Contains(got, " | ") {
		t.Errorf("sections should be separated by ' | ', got %q", got)
	}
}

func TestRender_NoNewline(t *testing.T) {
	r := newTestRenderer()
	data := &StatusData{
		Git:     GitStatusData{Branch: "main", Available: true},
		Memory:  MemoryData{TokensUsed: 50000, TokenBudget: 200000, Available: true},
		Metrics: MetricsData{Model: "Sonnet 4", Available: true},
	}

	got := r.Render(data, ModeDefault)

	if strings.Contains(got, "\n") {
		t.Errorf("output should not contain newline, got %q", got)
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
			name: "compact preset config",
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
			},
			wantContain:    []string{"🤖 Opus 4.5", "💬 MoAI", "📁 moai-adk-go", "🔀 main"},
			wantNotContain: []string{"🔋"},
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

func TestRender_SegmentFiltering_MinimalModeIgnoresConfig(t *testing.T) {
	// Minimal mode should ignore segment config (REQ-SL-042)
	segmentConfig := map[string]bool{
		SegmentModel: false, SegmentContext: false,
	}
	r := NewRenderer("default", true, segmentConfig)
	data := &StatusData{
		Metrics: MetricsData{Model: "Opus 4.5", Available: true},
		Memory:  MemoryData{TokensUsed: 50000, TokenBudget: 200000, Available: true},
	}

	got := r.Render(data, ModeMinimal)

	// Minimal mode uses hard-coded rendering, should still show model and context
	if !strings.Contains(got, "🤖 Opus 4.5") {
		t.Errorf("minimal mode should show model regardless of config, got %q", got)
	}
	if !strings.Contains(got, "🔋") {
		t.Errorf("minimal mode should show context regardless of config, got %q", got)
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
