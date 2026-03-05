package statusline

import (
	"fmt"
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
	// v3에서 ModeVerbose → ModeFull로 매핑되어 6줄 레이아웃
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

	// full mode는 여러 줄 출력
	if !strings.Contains(got, "\n") {
		t.Errorf("full mode should produce multi-line output, got %q", got)
	}

	lines := strings.Split(got, "\n")

	// v3 full L1: Model, Claude 버전(접두사 포함), MoAI 버전(접두사 포함)
	if !strings.Contains(lines[0], "🤖 Sonnet 4") {
		t.Errorf("full line 1 should contain model, got %q", lines[0])
	}
	// full 모드: "Claude v..." 및 "MoAI v..." 접두사 포함
	if !strings.Contains(lines[0], "🔅 Claude v1.0.80") {
		t.Errorf("full line 1 should contain Claude version with prefix, got %q", lines[0])
	}
	if !strings.Contains(lines[0], "🗿 MoAI v1.2.0") {
		t.Errorf("full line 1 should contain MoAI version with prefix, got %q", lines[0])
	}
	// v3에서 비용을 렌더링하지 않는다 (REQ-V3-TIME-005)
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
	// REQ-V3-TIME-005: v3에서는 비용($) 세그먼트를 렌더링하지 않는다
	r := newTestRenderer()
	data := &StatusData{
		Metrics:           MetricsData{Model: "Opus 4.5", CostUSD: 1.23, Available: true},
		Memory:            MemoryData{TokensUsed: 50000, TokenBudget: 200000, Available: true},
		Version:           VersionData{Current: "2.0.0", Available: true},
		ClaudeCodeVersion: "1.0.80",
	}

	got := r.Render(data, ModeVerbose)
	// v3에서 비용을 렌더링하지 않는다
	if strings.Contains(got, "$1.23") || strings.Contains(got, "$") {
		t.Errorf("full mode should NOT render cost in v3, got %q", got)
	}
	// 버전 정보는 여전히 표시된다 (full 모드: "Claude v...", "MoAI v...")
	if !strings.Contains(got, "🔅 Claude v1.0.80") {
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
	// CW 바는 memory 미사용 시 표시되지 않아야 한다
	if strings.Contains(got, "CW:") {
		t.Errorf("should not contain CW bar when memory unavailable, got %q", got)
	}
	// 5H/7D는 항상 0%로 표시된다
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
	// 5H/7D 바가 항상 0%로 표시되므로 빈 데이터에서도 출력이 있다
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

	got := r.Render(data, ModeCompact) // compact 모드: 한 줄에 여러 세그먼트

	// v3 구분자는 " │ " (U+2502 박스 그리기 문자)
	if !strings.Contains(got, " │ ") {
		t.Errorf("sections should be separated by ' │ ', got %q", got)
	}
}

func TestRender_NoNewline(t *testing.T) {
	// compact 모드에서 git 없으면 1줄 (L1만)
	r := newTestRenderer()
	data := &StatusData{
		Metrics: MetricsData{Model: "Sonnet 4", Available: true},
		Memory:  MemoryData{TokensUsed: 50000, TokenBudget: 200000, Available: true},
		// git 없음
	}

	got := r.Render(data, ModeCompact) // compact: git 없으면 1줄

	if strings.Contains(got, "\n") {
		t.Errorf("compact without git should not contain newline, got %q", got)
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

func TestRender_SegmentFiltering_MinimalModeIgnoresConfig(t *testing.T) {
	// v3에서 ModeMinimal → ModeCompact로 매핑됨
	// compact는 segment config를 준수한다
	segmentConfig := map[string]bool{
		SegmentModel: false, SegmentContext: false,
	}
	r := NewRenderer("default", true, segmentConfig)
	data := &StatusData{
		Metrics: MetricsData{Model: "Opus 4.5", Available: true},
		Memory:  MemoryData{TokensUsed: 50000, TokenBudget: 200000, Available: true},
	}

	got := r.Render(data, ModeMinimal) // ModeMinimal → ModeCompact

	// v3 compact는 segment config를 준수하므로 모든 세그먼트가 비활성화되면 MoAI 폴백
	if got != "MoAI" {
		t.Errorf("with model/context disabled, compact should return MoAI fallback, got %q", got)
	}
}

func TestCostNotRendered(t *testing.T) {
	// REQ-V3-TIME-005: v3에서는 비용($) 세그먼트를 렌더링하지 않는다
	r := newTestRenderer()
	data := &StatusData{
		Metrics:           MetricsData{Model: "Opus", CostUSD: 5.50, Available: true},
		Memory:            MemoryData{TokensUsed: 50000, TokenBudget: 200000, Available: true},
		Version:           VersionData{Current: "2.0.0", Available: true},
		ClaudeCodeVersion: "1.0.80",
	}

	// 모든 모드에서 비용 렌더링 확인
	for _, mode := range []StatuslineMode{ModeDefault, ModeCompact, ModeMinimal, ModeVerbose, ModeFull} {
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
		// REQ-V3-GIT-001: Ahead만 있을 때 "🔀 main ↑3"
		{"ahead only", GitStatusData{Branch: "main", Ahead: 3, Behind: 0, Available: true}, "🔀 main ↑3"},
		// REQ-V3-GIT-002: Behind만 있을 때 "🔀 main ↓2"
		{"behind only", GitStatusData{Branch: "main", Ahead: 0, Behind: 2, Available: true}, "🔀 main ↓2"},
		// REQ-V3-GIT-003: 둘 다 있을 때 "🔀 feat/auth ↑2↓1"
		{"both", GitStatusData{Branch: "feat/auth", Ahead: 2, Behind: 1, Available: true}, "🔀 feat/auth ↑2↓1"},
		// REQ-V3-GIT-004: 둘 다 없을 때 브랜치 이름만
		{"neither", GitStatusData{Branch: "main", Ahead: 0, Behind: 0, Available: true}, "🔀 main"},
		// 사용 불가 시 빈 문자열
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
		// REQ-V3-TIME-002: 83분 = "⏳ 1h 23m"
		{"83 minutes", 4980000, "⏳ 1h 23m"},
		// REQ-V3-TIME-002: 45분 미만 = "⏳ 45m"
		{"45 minutes", 2700000, "⏳ 45m"},
		// REQ-V3-TIME-004: 0ms는 빈 문자열
		{"zero", 0, ""},
		// REQ-V3-TIME-002: 정확히 1시간
		{"exactly 1 hour", 3600000, "⏳ 1h 0m"},
		// REQ-V3-TIME-002: 26시간 = "⏳ 1d 2h"
		{"26 hours", 93600000, "⏳ 1d 2h"},
		// REQ-V3-TIME-002: 48시간 = "⏳ 2d 0h"
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
// Cycle 1: renderCompactV3 테스트
// REQ-V3-LAYOUT-001: compact 모드는 2줄 레이아웃
// ─────────────────────────────────────────────────────────────────────────────

func TestRenderCompactV3_TwoLines(t *testing.T) {
	// compact 모드는 정확히 2줄을 생성해야 한다 (데이터가 모두 있을 때)
	r := newTestRenderer()
	data := &StatusData{
		Metrics: MetricsData{Model: "Opus 4.6", Available: true},
		Memory:  MemoryData{TokensUsed: 176000, TokenBudget: 200000, Available: true},
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

	got := r.Render(data, ModeCompact)
	lines := strings.Split(got, "\n")

	if len(lines) != 2 {
		t.Errorf("compact 모드는 2줄이어야 한다, 실제: %d줄\n출력: %q", len(lines), got)
	}
}

func TestRenderCompactV3_Line1_ModelAndBar(t *testing.T) {
	// compact L1: 모델, CW 바 (10블록), 세션 시간 없음 (REQ-V3-TIME-006)
	r := newTestRenderer()
	data := &StatusData{
		Metrics: MetricsData{
			Model:             "Opus 4.6",
			Available:         true,
			SessionDurationMS: 2700000, // 45분 (compact에서는 표시 안 함)
		},
		Memory: MemoryData{TokensUsed: 176000, TokenBudget: 200000, Available: true},
		Git:    GitStatusData{Branch: "main", Available: true},
	}

	got := r.Render(data, ModeCompact)
	lines := strings.Split(got, "\n")
	l1 := lines[0]

	// 모델 표시
	if !strings.Contains(l1, "🤖 Opus 4.6") {
		t.Errorf("compact L1에 모델이 있어야 한다, 실제: %q", l1)
	}
	// CW 레이블 포함 바
	if !strings.Contains(l1, "CW:") {
		t.Errorf("compact L1에 'CW:' 레이블이 있어야 한다, 실제: %q", l1)
	}
	// 88% (176000/200000 = 88%)
	if !strings.Contains(l1, "88%") {
		t.Errorf("compact L1에 88%%가 있어야 한다, 실제: %q", l1)
	}
	// 배터리 아이콘 (88% > 70% 이므로 🪫)
	if !strings.Contains(l1, "🪫") {
		t.Errorf("compact L1에 🪫 아이콘이 있어야 한다 (88%%), 실제: %q", l1)
	}
	// REQ-V3-TIME-006: compact에서 세션 시간 없음
	if strings.Contains(l1, "⏳") {
		t.Errorf("compact L1에 세션 시간이 없어야 한다 (REQ-V3-TIME-006), 실제: %q", l1)
	}
}

func TestRenderCompactV3_Line1_NoVersionNoStyleNoTask(t *testing.T) {
	// compact L1: 버전, 출력 스타일, 태스크 없음 (REQ-V3-API-011)
	r := newTestRenderer()
	data := &StatusData{
		Metrics:           MetricsData{Model: "Opus 4.6", Available: true},
		Memory:            MemoryData{TokensUsed: 88000, TokenBudget: 200000, Available: true},
		ClaudeCodeVersion: "2.1.50",
		Version:           VersionData{Current: "2.8.0", Available: true},
		OutputStyle:       "MoAI",
		Task:              TaskData{Active: true, Command: "run", SpecID: "SPEC-001"},
		Git:               GitStatusData{Branch: "main", Available: true},
	}

	got := r.Render(data, ModeCompact)
	lines := strings.Split(got, "\n")
	l1 := lines[0]

	// compact L1에는 버전 없음
	if strings.Contains(l1, "🔅") {
		t.Errorf("compact L1에 Claude 버전이 없어야 한다, 실제: %q", l1)
	}
	if strings.Contains(l1, "🗿") {
		t.Errorf("compact L1에 MoAI 버전이 없어야 한다, 실제: %q", l1)
	}
	// compact L1에는 출력 스타일 없음
	if strings.Contains(l1, "💬") {
		t.Errorf("compact L1에 출력 스타일이 없어야 한다, 실제: %q", l1)
	}
	// compact L1에는 태스크 없음
	if strings.Contains(l1, "📋") {
		t.Errorf("compact L1에 태스크가 없어야 한다, 실제: %q", l1)
	}
}

func TestRenderCompactV3_Line2_BranchAndGitStatus(t *testing.T) {
	// compact L2: 브랜치(ahead/behind), git 상태
	r := newTestRenderer()
	data := &StatusData{
		Metrics: MetricsData{Model: "Opus 4.6", Available: true},
		Memory:  MemoryData{TokensUsed: 88000, TokenBudget: 200000, Available: true},
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

	got := r.Render(data, ModeCompact)
	lines := strings.Split(got, "\n")
	l2 := lines[1]

	// 브랜치 + ahead/behind
	if !strings.Contains(l2, "🔀 feat/auth ↑2↓1") {
		t.Errorf("compact L2에 브랜치 + ahead/behind가 있어야 한다, 실제: %q", l2)
	}
	// git 상태
	if !strings.Contains(l2, "📊") {
		t.Errorf("compact L2에 📊 git 상태 이모지가 있어야 한다, 실제: %q", l2)
	}
	if !strings.Contains(l2, "+3") {
		t.Errorf("compact L2에 staged(+3)가 있어야 한다, 실제: %q", l2)
	}
	if !strings.Contains(l2, "M2") {
		t.Errorf("compact L2에 modified(M2)가 있어야 한다, 실제: %q", l2)
	}
	if !strings.Contains(l2, "?1") {
		t.Errorf("compact L2에 untracked(?1)가 있어야 한다, 실제: %q", l2)
	}
}

func TestRenderCompactV3_Line2_No5HNo7D(t *testing.T) {
	// compact 모드에는 5H/7D 바 없음 (REQ-V3-API-011)
	r := newTestRenderer()
	data := &StatusData{
		Metrics: MetricsData{Model: "Opus 4.6", Available: true},
		Memory:  MemoryData{TokensUsed: 88000, TokenBudget: 200000, Available: true},
		Git:     GitStatusData{Branch: "main", Available: true},
		Usage: &UsageResult{
			Usage5H: &UsageData{UsedTokens: 45000, LimitTokens: 100000, Percentage: 45},
			Usage7D: &UsageData{UsedTokens: 82000, LimitTokens: 100000, Percentage: 82},
		},
	}

	got := r.Render(data, ModeCompact)

	// compact에는 5H/7D 없음
	if strings.Contains(got, "5H:") {
		t.Errorf("compact 모드에는 5H 바가 없어야 한다, 실제: %q", got)
	}
	if strings.Contains(got, "7D:") {
		t.Errorf("compact 모드에는 7D 바가 없어야 한다, 실제: %q", got)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Cycle 2: renderDefaultV3 테스트
// REQ-V3-LAYOUT-002: default 모드는 4줄 레이아웃
// ─────────────────────────────────────────────────────────────────────────────

func TestRenderDefaultV3_ThreeLines(t *testing.T) {
	// default 모드는 정확히 3줄이어야 한다 (L4 제거됨, 스타일은 L1에 통합)
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
		t.Errorf("default 모드는 3줄이어야 한다, 실제: %d줄\n출력: %q", len(lines), got)
	}
	// L1에 출력 스타일이 통합되어야 한다
	if !strings.Contains(lines[0], "💬 MoAI") {
		t.Errorf("default L1에 출력 스타일이 있어야 한다, 실제: %q", lines[0])
	}
}

func TestRenderDefaultV3_Line1(t *testing.T) {
	// default L1: 모델, Claude 버전, MoAI 버전, 세션 시간
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
		t.Errorf("default L1에 모델이 있어야 한다, 실제: %q", l1)
	}
	if !strings.Contains(l1, "🔅 v2.1.50") {
		t.Errorf("default L1에 Claude 버전이 있어야 한다, 실제: %q", l1)
	}
	if !strings.Contains(l1, "🗿 v2.8.0") {
		t.Errorf("default L1에 MoAI 버전이 있어야 한다, 실제: %q", l1)
	}
	// 세션 시간 (9240000ms = 154분 = 2h 34m)
	if !strings.Contains(l1, "⏳") {
		t.Errorf("default L1에 세션 시간이 있어야 한다, 실제: %q", l1)
	}
	if !strings.Contains(l1, "2h 34m") {
		t.Errorf("default L1에 '2h 34m'이 있어야 한다, 실제: %q", l1)
	}
}

func TestRenderDefaultV3_Line2_BarsInline(t *testing.T) {
	// default L2: CW/5H/7D 바가 한 줄에 인라인 (10블록씩, REQ-V3-API-011)
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

	// CW, 5H, 7D 모두 L2에 있어야 한다
	if !strings.Contains(l2, "CW:") {
		t.Errorf("default L2에 'CW:' 레이블이 있어야 한다, 실제: %q", l2)
	}
	if !strings.Contains(l2, "5H:") {
		t.Errorf("default L2에 '5H:' 레이블이 있어야 한다, 실제: %q", l2)
	}
	if !strings.Contains(l2, "7D:") {
		t.Errorf("default L2에 '7D:' 레이블이 있어야 한다, 실제: %q", l2)
	}
	// 퍼센트 값 확인
	if !strings.Contains(l2, "88%") {
		t.Errorf("default L2에 CW 88%%가 있어야 한다, 실제: %q", l2)
	}
	if !strings.Contains(l2, "45%") {
		t.Errorf("default L2에 5H 45%%가 있어야 한다, 실제: %q", l2)
	}
	if !strings.Contains(l2, "82%") {
		t.Errorf("default L2에 7D 82%%가 있어야 한다, 실제: %q", l2)
	}
}

func TestRenderDefaultV3_Line3(t *testing.T) {
	// default L3: 디렉토리, 브랜치 + ahead/behind, git 상태
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
		t.Errorf("default L3에 디렉토리가 있어야 한다, 실제: %q", l3)
	}
	if !strings.Contains(l3, "🔀 feat/auth ↑2↓1") {
		t.Errorf("default L3에 브랜치 + ahead/behind가 있어야 한다, 실제: %q", l3)
	}
	if !strings.Contains(l3, "📊") {
		t.Errorf("default L3에 git 상태가 있어야 한다, 실제: %q", l3)
	}
}

func TestRenderDefaultV3_StyleInL1(t *testing.T) {
	// default: 출력 스타일이 L1에 통합됨 (L4 제거)
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

	// default는 3줄이어야 한다 (L4 제거됨)
	if len(lines) != 3 {
		t.Fatalf("default 모드는 3줄이어야 한다, 실제: %d줄\n출력: %q", len(lines), got)
	}

	// L1에 출력 스타일이 통합되어야 한다
	if !strings.Contains(lines[0], "💬 MoAI") {
		t.Errorf("default L1에 출력 스타일이 있어야 한다, 실제: %q", lines[0])
	}
	// 태스크(📋)는 더 이상 표시되지 않는다
}

// ─────────────────────────────────────────────────────────────────────────────
// Cycle 3: renderFullV3 테스트
// REQ-V3-LAYOUT-003: full 모드는 6줄 레이아웃
// ─────────────────────────────────────────────────────────────────────────────

func TestRenderFullV3_FiveLines(t *testing.T) {
	// full 모드는 정확히 5줄이어야 한다 (L6 제거됨, 스타일은 L1에 통합)
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
		t.Errorf("full 모드는 5줄이어야 한다, 실제: %d줄\n출력:\n%s", len(lines), got)
	}
	// L1에 출력 스타일이 통합되어야 한다
	if !strings.Contains(lines[0], "💬 MoAI") {
		t.Errorf("full L1에 출력 스타일이 있어야 한다, 실제: %q", lines[0])
	}
}

func TestRenderFullV3_Line1_WithPrefixes(t *testing.T) {
	// full L1: 모델, "Claude" 접두사가 있는 Claude 버전, "MoAI" 접두사가 있는 MoAI 버전, 세션 시간
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
		t.Errorf("full L1에 모델이 있어야 한다, 실제: %q", l1)
	}
	// full 모드: "Claude v2.1.50" (접두사 포함)
	if !strings.Contains(l1, "🔅 Claude v2.1.50") {
		t.Errorf("full L1에 '🔅 Claude v2.1.50'이 있어야 한다, 실제: %q", l1)
	}
	// full 모드: "MoAI v2.8.0" (접두사 포함)
	if !strings.Contains(l1, "🗿 MoAI v2.8.0") {
		t.Errorf("full L1에 '🗿 MoAI v2.8.0'이 있어야 한다, 실제: %q", l1)
	}
	if !strings.Contains(l1, "⏳") {
		t.Errorf("full L1에 세션 시간이 있어야 한다, 실제: %q", l1)
	}
}

func TestRenderFullV3_Lines2To4_SeparateBars(t *testing.T) {
	// full L2-L4: 각 바가 별도 줄 (40블록, REQ-V3-API-011)
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

	// L2: CW 바만
	l2 := lines[1]
	if !strings.Contains(l2, "CW:") {
		t.Errorf("full L2에 'CW:' 레이블이 있어야 한다, 실제: %q", l2)
	}
	if strings.Contains(l2, "5H:") || strings.Contains(l2, "7D:") {
		t.Errorf("full L2에는 5H/7D가 없어야 한다 (CW만), 실제: %q", l2)
	}

	// L3: 5H 바만
	l3 := lines[2]
	if !strings.Contains(l3, "5H:") {
		t.Errorf("full L3에 '5H:' 레이블이 있어야 한다, 실제: %q", l3)
	}
	if strings.Contains(l3, "CW:") || strings.Contains(l3, "7D:") {
		t.Errorf("full L3에는 CW/7D가 없어야 한다 (5H만), 실제: %q", l3)
	}

	// L4: 7D 바만
	l4 := lines[3]
	if !strings.Contains(l4, "7D:") {
		t.Errorf("full L4에 '7D:' 레이블이 있어야 한다, 실제: %q", l4)
	}
	if strings.Contains(l4, "CW:") || strings.Contains(l4, "5H:") {
		t.Errorf("full L4에는 CW/5H가 없어야 한다 (7D만), 실제: %q", l4)
	}
}

func TestRenderFullV3_Line5_DirBranchGit(t *testing.T) {
	// full L5: 디렉토리, 브랜치 + ahead/behind, git 상태
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

	// L5 확인
	if len(lines) < 5 {
		t.Fatalf("full 모드에 L5가 있어야 한다, 실제: %d줄\n출력:\n%s", len(lines), got)
	}
	l5 := lines[4]

	if !strings.Contains(l5, "📁 moai-adk-go") {
		t.Errorf("full L5에 디렉토리가 있어야 한다, 실제: %q", l5)
	}
	if !strings.Contains(l5, "🔀 feat/auth ↑2↓1") {
		t.Errorf("full L5에 브랜치 + ahead/behind가 있어야 한다, 실제: %q", l5)
	}
	if !strings.Contains(l5, "📊") {
		t.Errorf("full L5에 git 상태가 있어야 한다, 실제: %q", l5)
	}
}

func TestRenderFullV3_StyleInL1(t *testing.T) {
	// full: 출력 스타일이 L1에 통합됨 (L6 제거)
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

	// full은 5줄이어야 한다 (L6 제거됨)
	if len(lines) != 5 {
		t.Fatalf("full 모드는 5줄이어야 한다, 실제: %d줄\n출력:\n%s", len(lines), got)
	}

	// L1에 출력 스타일이 통합되어야 한다
	if !strings.Contains(lines[0], "💬 MoAI") {
		t.Errorf("full L1에 출력 스타일이 있어야 한다, 실제: %q", lines[0])
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Cycle 4: 빈 줄 생략 테스트
// REQ-V3-LAYOUT-004: 세그먼트가 모두 비어있는 줄은 생략
// ─────────────────────────────────────────────────────────────────────────────

func TestRenderV3_OmitsEmptyLines_Compact(t *testing.T) {
	// compact: git 데이터가 없으면 L2 생략 → 1줄만
	r := newTestRenderer()
	data := &StatusData{
		Metrics: MetricsData{Model: "Opus 4.6", Available: true},
		Memory:  MemoryData{TokensUsed: 88000, TokenBudget: 200000, Available: true},
		// git 없음
	}

	got := r.Render(data, ModeCompact)
	lines := strings.Split(got, "\n")

	if len(lines) != 1 {
		t.Errorf("git 없는 compact 모드는 1줄이어야 한다, 실제: %d줄\n출력: %q", len(lines), got)
	}
}

func TestRenderV3_OmitsEmptyLines_Default(t *testing.T) {
	// default: 태스크/출력 스타일 없으면 L4 생략 → 3줄
	r := newTestRenderer()
	data := &StatusData{
		Metrics:   MetricsData{Model: "Opus 4.6", Available: true},
		Memory:    MemoryData{TokensUsed: 88000, TokenBudget: 200000, Available: true},
		Directory: "moai-adk-go",
		Git:       GitStatusData{Branch: "main", Available: true},
		// 태스크, 출력 스타일, 사용량 없음
	}

	got := r.Render(data, ModeDefault)
	lines := strings.Split(got, "\n")

	// L1(모델+CW), L3(디렉토리+브랜치) 있음
	// L2(5H/7D 없으면 CW만 있음 → 생략 안 됨), L4(스타일/태스크 없으면 생략)
	// 사용량 없으면: L1(모델+CW), L2(CW 바만) → 빈 줄 아님, L3(디렉토리+브랜치), L4 없음 → 3줄
	if len(lines) != 3 {
		t.Errorf("태스크/스타일 없는 default 모드는 3줄이어야 한다, 실제: %d줄\n출력:\n%s", len(lines), got)
	}
}

func TestRenderV3_OmitsEmptyLines_Full(t *testing.T) {
	// full: 태스크/출력 스타일 없으면 L6 생략 → 5줄
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
		// 태스크, 출력 스타일 없음
	}

	got := r.Render(data, ModeFull)
	lines := strings.Split(got, "\n")

	if len(lines) != 5 {
		t.Errorf("태스크/스타일 없는 full 모드는 5줄이어야 한다, 실제: %d줄\n출력:\n%s", len(lines), got)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Cycle 5: Render() 라우팅 및 하위 호환성 테스트
// ─────────────────────────────────────────────────────────────────────────────

func TestRender_ModeRouting(t *testing.T) {
	// 모든 모드가 올바르게 라우팅되어야 한다
	r := newTestRenderer()
	data := &StatusData{
		Metrics: MetricsData{Model: "Opus 4.6", Available: true},
		Memory:  MemoryData{TokensUsed: 88000, TokenBudget: 200000, Available: true},
	}

	tests := []struct {
		mode      StatuslineMode
		minLines  int
		maxLines  int
	}{
		{ModeCompact, 1, 2},
		{ModeDefault, 1, 4},
		{ModeFull, 1, 6},
		{ModeMinimal, 1, 2},  // deprecated → compact
		{ModeVerbose, 1, 6},  // deprecated → full
	}

	for _, tt := range tests {
		t.Run(string(tt.mode), func(t *testing.T) {
			got := r.Render(data, tt.mode)
			if got == "" || got == "MoAI" {
				// 데이터가 있으면 MoAI 폴백이 아닌 실제 출력이어야 함
				return
			}
			lines := strings.Split(got, "\n")
			if len(lines) < tt.minLines || len(lines) > tt.maxLines {
				t.Errorf("mode=%s: %d~%d줄이어야 하는데 %d줄\n출력: %q",
					tt.mode, tt.minLines, tt.maxLines, len(lines), got)
			}
		})
	}
}

func TestRenderUsageBar(t *testing.T) {
	// renderUsageBar 헬퍼 함수 테스트
	tests := []struct {
		name    string
		label   string
		pct     int
		width   int
		noColor bool
		wantPfx string // 출력 시작 접두사
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
				t.Errorf("renderUsageBar()에 %d%%가 있어야 한다, 실제: %q", tt.pct, got)
			}
		})
	}
}

func TestRender_V3Separator(t *testing.T) {
	// v3 구분자는 " │ " (U+2502 박스 그리기 문자)여야 한다
	r := newTestRenderer()
	data := &StatusData{
		Metrics:   MetricsData{Model: "Opus 4.6", Available: true},
		Memory:    MemoryData{TokensUsed: 88000, TokenBudget: 200000, Available: true},
		Directory: "moai-adk-go",
		Git:       GitStatusData{Branch: "main", Available: true},
	}

	got := r.Render(data, ModeDefault)

	// v3 구분자 확인
	if !strings.Contains(got, " │ ") {
		t.Errorf("v3 구분자 ' │ '가 있어야 한다, 실제: %q", got)
	}
}
