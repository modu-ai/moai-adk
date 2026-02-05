package statusline

import (
	"strings"
	"testing"
)

// newTestRenderer creates a Renderer with NoColor=true for predictable test output.
func newTestRenderer() *Renderer {
	return NewRenderer("default", true)
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

func TestRender_VerboseMode(t *testing.T) {
	r := newTestRenderer()
	data := &StatusData{
		Git:         GitStatusData{Branch: "main", Staged: 3, Modified: 2, Untracked: 1, Available: true},
		Memory:      MemoryData{TokensUsed: 50000, TokenBudget: 200000, Available: true},
		Metrics:     MetricsData{Model: "Sonnet 4", Available: true},
		Directory:   "my-project",
		OutputStyle: "Yoda",
		Version:     VersionData{Current: "1.2.0", Available: true},
	}

	got := r.Render(data, ModeVerbose)

	// Verbose mode has same format as compact
	if !strings.Contains(got, "🤖 Sonnet 4") {
		t.Errorf("verbose mode should contain model, got %q", got)
	}
	if !strings.Contains(got, "📊 +3 M2 ?1") {
		t.Errorf("verbose mode should contain git status, got %q", got)
	}
	if !strings.Contains(got, "🔀 main") {
		t.Errorf("verbose mode should contain branch, got %q", got)
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
		{"minimal", false},
		{"nerd", false},
		{"unknown", false},
	}

	for _, tt := range tests {
		t.Run(tt.theme, func(t *testing.T) {
			r := NewRenderer(tt.theme, tt.noColor)
			if r == nil {
				t.Fatal("NewRenderer returned nil")
			}
			if r.noColor != tt.noColor {
				t.Errorf("noColor = %v, want %v", r.noColor, tt.noColor)
			}
		})
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
