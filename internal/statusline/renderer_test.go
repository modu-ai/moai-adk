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
		Metrics: MetricsData{Model: "sonnet-4", Available: true},
		Memory:  MemoryData{TokensUsed: 50000, TokenBudget: 200000, Available: true},
	}

	got := r.Render(data, ModeMinimal)

	if !strings.Contains(got, "sonnet-4") {
		t.Errorf("minimal mode should contain model name, got %q", got)
	}
	if !strings.Contains(got, "Ctx: 25%") {
		t.Errorf("minimal mode should contain context percentage, got %q", got)
	}
	// Minimal should NOT contain git or cost info
	if strings.Contains(got, "main") {
		t.Errorf("minimal mode should not contain git info, got %q", got)
	}
	if strings.Contains(got, "$") {
		t.Errorf("minimal mode should not contain cost info, got %q", got)
	}
}

func TestRender_DefaultMode(t *testing.T) {
	r := newTestRenderer()
	data := &StatusData{
		Git:     GitStatusData{Branch: "main", Modified: 2, Staged: 3, Available: true},
		Memory:  MemoryData{TokensUsed: 50000, TokenBudget: 200000, Available: true},
		Metrics: MetricsData{Model: "sonnet-4", CostUSD: 0.05, Available: true},
	}

	got := r.Render(data, ModeDefault)

	if !strings.Contains(got, "main") {
		t.Errorf("default mode should contain branch name, got %q", got)
	}
	if !strings.Contains(got, "+3") {
		t.Errorf("default mode should contain staged count, got %q", got)
	}
	if !strings.Contains(got, "~2") {
		t.Errorf("default mode should contain modified count, got %q", got)
	}
	if !strings.Contains(got, "Ctx: 25%") {
		t.Errorf("default mode should contain context percentage, got %q", got)
	}
	if !strings.Contains(got, "$0.05") {
		t.Errorf("default mode should contain cost, got %q", got)
	}
}

func TestRender_VerboseMode(t *testing.T) {
	r := newTestRenderer()
	data := &StatusData{
		Git: GitStatusData{
			Branch: "main", Staged: 3, Modified: 2,
			Ahead: 1, Behind: 0, Available: true,
		},
		Memory:  MemoryData{TokensUsed: 50000, TokenBudget: 200000, Available: true},
		Metrics: MetricsData{Model: "sonnet-4", CostUSD: 0.05, Available: true},
		Version: VersionData{Current: "1.2.0", Latest: "1.3.0", UpdateAvailable: true, Available: true},
	}

	got := r.Render(data, ModeVerbose)

	if !strings.Contains(got, "main") {
		t.Errorf("verbose mode should contain branch, got %q", got)
	}
	if !strings.Contains(got, "+3") {
		t.Errorf("verbose mode should contain staged count, got %q", got)
	}
	if !strings.Contains(got, "~2") {
		t.Errorf("verbose mode should contain modified count, got %q", got)
	}
	if !strings.Contains(got, "^1") {
		t.Errorf("verbose mode should contain ahead count, got %q", got)
	}
	if !strings.Contains(got, "v0") {
		t.Errorf("verbose mode should contain behind count, got %q", got)
	}
	if !strings.Contains(got, "50K/200K") {
		t.Errorf("verbose mode should contain token counts, got %q", got)
	}
	if !strings.Contains(got, "(25%)") {
		t.Errorf("verbose mode should contain percentage, got %q", got)
	}
	if !strings.Contains(got, "$0.05") {
		t.Errorf("verbose mode should contain cost, got %q", got)
	}
	if !strings.Contains(got, "v1.2.0") {
		t.Errorf("verbose mode should contain version, got %q", got)
	}
	if !strings.Contains(got, "(update!)") {
		t.Errorf("verbose mode should contain update notification, got %q", got)
	}
}

func TestRender_NoUpdateNotification(t *testing.T) {
	r := newTestRenderer()
	data := &StatusData{
		Version: VersionData{Current: "1.3.0", Latest: "1.3.0", UpdateAvailable: false, Available: true},
		Memory:  MemoryData{TokensUsed: 50000, TokenBudget: 200000, Available: true},
		Metrics: MetricsData{CostUSD: 0.05, Available: true},
	}

	got := r.Render(data, ModeVerbose)

	if strings.Contains(got, "(update!)") {
		t.Errorf("should not show update notification when up to date, got %q", got)
	}
}

func TestRender_EmptyGit(t *testing.T) {
	r := newTestRenderer()
	data := &StatusData{
		Git:     GitStatusData{Available: false},
		Memory:  MemoryData{TokensUsed: 50000, TokenBudget: 200000, Available: true},
		Metrics: MetricsData{CostUSD: 0.05, Available: true},
	}

	got := r.Render(data, ModeDefault)

	// Should not contain any git info
	if strings.Contains(got, "main") {
		t.Errorf("should not contain git branch when unavailable, got %q", got)
	}
	// But should still have context and cost
	if !strings.Contains(got, "Ctx:") {
		t.Errorf("should still contain context info, got %q", got)
	}
}

func TestRender_EmptyMemory(t *testing.T) {
	r := newTestRenderer()
	data := &StatusData{
		Git:     GitStatusData{Branch: "main", Available: true},
		Memory:  MemoryData{Available: false},
		Metrics: MetricsData{CostUSD: 0.05, Available: true},
	}

	got := r.Render(data, ModeDefault)

	if !strings.Contains(got, "main") {
		t.Errorf("should contain git info, got %q", got)
	}
	if strings.Contains(got, "Ctx:") {
		t.Errorf("should not contain context when unavailable, got %q", got)
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

func TestRender_ZeroCostNotShownInDefault(t *testing.T) {
	r := newTestRenderer()
	data := &StatusData{
		Git:     GitStatusData{Branch: "main", Available: true},
		Memory:  MemoryData{TokensUsed: 50000, TokenBudget: 200000, Available: true},
		Metrics: MetricsData{CostUSD: 0, Available: true},
	}

	got := r.Render(data, ModeDefault)

	if strings.Contains(got, "$0.00") {
		t.Errorf("default mode should not show zero cost, got %q", got)
	}
}

func TestRender_GitOnlyBranch(t *testing.T) {
	r := newTestRenderer()
	data := &StatusData{
		Git:    GitStatusData{Branch: "main", Available: true},
		Memory: MemoryData{TokensUsed: 50000, TokenBudget: 200000, Available: true},
	}

	got := r.Render(data, ModeDefault)

	if !strings.Contains(got, "main") {
		t.Errorf("should show branch name, got %q", got)
	}
	// Should not have staged/modified indicators with zero counts
	if strings.Contains(got, "+0") || strings.Contains(got, "~0") {
		t.Errorf("should not show zero counts, got %q", got)
	}
}

func TestRender_Separator(t *testing.T) {
	r := newTestRenderer()
	data := &StatusData{
		Git:     GitStatusData{Branch: "main", Modified: 2, Available: true},
		Memory:  MemoryData{TokensUsed: 50000, TokenBudget: 200000, Available: true},
		Metrics: MetricsData{CostUSD: 0.05, Available: true},
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
		Metrics: MetricsData{CostUSD: 0.05, Available: true},
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
			if tt.theme == "nerd" && !r.useNerd {
				t.Error("nerd theme should set useNerd=true")
			}
		})
	}
}

func TestRender_ContextColorLogic(t *testing.T) {
	// Test that the correct style level is chosen based on usage
	tests := []struct {
		name  string
		used  int
		total int
		want  contextLevel
	}{
		{"green - 25%", 50000, 200000, levelOk},
		{"yellow - 65%", 130000, 200000, levelWarn},
		{"red - 90%", 180000, 200000, levelError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := contextUsageLevel(tt.used, tt.total)
			if got != tt.want {
				t.Errorf("contextUsageLevel(%d, %d) = %d, want %d",
					tt.used, tt.total, got, tt.want)
			}
		})
	}
}

func TestRender_ApplyContextStyle_WithColor(t *testing.T) {
	// Test that color styles are applied when NoColor=false.
	// In non-TTY test environments lipgloss may strip ANSI codes,
	// so we verify the text content is preserved regardless.
	r := NewRenderer("default", false)

	tests := []struct {
		name  string
		used  int
		total int
	}{
		{"green style", 50000, 200000},
		{"yellow style", 130000, 200000},
		{"red style", 180000, 200000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := r.applyContextStyle("Ctx: 25%", tt.used, tt.total)
			if !strings.Contains(got, "Ctx: 25%") {
				t.Errorf("styled output should contain original text, got %q", got)
			}
		})
	}
}

func TestRender_ContextBriefUnavailable(t *testing.T) {
	r := newTestRenderer()
	data := &StatusData{
		Memory: MemoryData{Available: false},
	}
	got := r.renderContextBrief(data)
	if got != "" {
		t.Errorf("renderContextBrief with unavailable memory should return empty, got %q", got)
	}
}

func TestRender_ContextVerboseUnavailable(t *testing.T) {
	r := newTestRenderer()
	data := &StatusData{
		Memory: MemoryData{Available: false},
	}
	got := r.renderContextVerbose(data)
	if got != "" {
		t.Errorf("renderContextVerbose with unavailable memory should return empty, got %q", got)
	}
}

func TestRender_VersionEmptyCurrent(t *testing.T) {
	r := newTestRenderer()
	data := &StatusData{
		Version: VersionData{Available: true, Current: ""},
	}
	got := r.renderVersion(data)
	if got != "" {
		t.Errorf("renderVersion with empty current should return empty, got %q", got)
	}
}
