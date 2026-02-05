package statusline

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"
)

// mockGitProvider implements GitDataProvider for testing.
type mockGitProvider struct {
	data *GitStatusData
	err  error
}

func (m *mockGitProvider) CollectGitStatus(_ context.Context) (*GitStatusData, error) {
	return m.data, m.err
}

// mockUpdateProvider implements UpdateProvider for testing.
type mockUpdateProvider struct {
	data *VersionData
	err  error
}

func (m *mockUpdateProvider) CheckUpdate(_ context.Context) (*VersionData, error) {
	return m.data, m.err
}

// slowGitProvider simulates a slow git collection.
type slowGitProvider struct {
	delay time.Duration
}

func (s *slowGitProvider) CollectGitStatus(ctx context.Context) (*GitStatusData, error) {
	select {
	case <-time.After(s.delay):
		return &GitStatusData{Branch: "main", Available: true}, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func makeStdinJSON(data *StdinData) *bytes.Buffer {
	b, err := json.Marshal(data)
	if err != nil {
		// Should never happen with well-formed test data
		panic("failed to marshal test stdin data: " + err.Error())
	}
	return bytes.NewBuffer(b)
}

func TestBuilder_Build_FullData(t *testing.T) {
	builder := New(Options{
		GitProvider: &mockGitProvider{
			data: &GitStatusData{
				Branch: "main", Modified: 2, Staged: 3, Available: true,
			},
		},
		UpdateProvider: &mockUpdateProvider{
			data: &VersionData{
				Current: "1.2.0", Latest: "1.3.0",
				UpdateAvailable: true, Available: true,
			},
		},
		ThemeName: "default",
		Mode:      ModeDefault,
		NoColor:   true,
	})

	input := &StdinData{
		Model:         &ModelInfo{Name: "claude-sonnet-4-20250514"},
		Cost:          &CostData{TotalUSD: 0.05},
		ContextWindow: &ContextWindowInfo{Used: 50000, Total: 200000},
		CWD:           "/Users/test/my-project",
		OutputStyle:   &OutputStyleInfo{Name: "Mr.Alfred"},
	}

	got, err := builder.Build(context.Background(), makeStdinJSON(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Default mode: model + context graph + output style + git status + version + branch
	if !strings.Contains(got, "🤖 Sonnet 4") {
		t.Errorf("should contain model name with emoji, got %q", got)
	}
	if !strings.Contains(got, "🔋 ") {
		t.Errorf("should contain context bar graph, got %q", got)
	}
	if !strings.Contains(got, "█") {
		t.Errorf("should contain bar graph characters, got %q", got)
	}
	if !strings.Contains(got, "25%") {
		t.Errorf("should contain context percentage, got %q", got)
	}
	if !strings.Contains(got, "💬 Mr.Alfred") {
		t.Errorf("should contain output style, got %q", got)
	}
	if !strings.Contains(got, "📁 my-project") {
		t.Errorf("should contain directory, got %q", got)
	}
	if !strings.Contains(got, "📊 +3 M2") {
		t.Errorf("should contain git status, got %q", got)
	}
	if !strings.Contains(got, "🗿 v1.2.0") {
		t.Errorf("should contain MoAI version with 🗿 emoji, got %q", got)
	}
	if !strings.Contains(got, "🔀 main") {
		t.Errorf("should contain branch, got %q", got)
	}
}

func TestBuilder_Build_NilReader(t *testing.T) {
	builder := New(Options{
		Mode:    ModeDefault,
		NoColor: true,
	})

	got, err := builder.Build(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should produce fallback output, not panic
	if got == "" {
		t.Error("nil reader should still produce output")
	}
}

func TestBuilder_Build_InvalidJSON(t *testing.T) {
	builder := New(Options{
		Mode:    ModeDefault,
		NoColor: true,
	})

	invalidJSON := bytes.NewBufferString("{ invalid json content")

	got, err := builder.Build(context.Background(), invalidJSON)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should produce fallback output, not panic
	if got == "" {
		t.Error("invalid JSON should still produce output")
	}
}

func TestBuilder_Build_EmptyReader(t *testing.T) {
	builder := New(Options{
		Mode:    ModeDefault,
		NoColor: true,
	})

	got, err := builder.Build(context.Background(), bytes.NewBuffer(nil))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got == "" {
		t.Error("empty reader should still produce output")
	}
}

func TestBuilder_Build_GitProviderFailure(t *testing.T) {
	builder := New(Options{
		GitProvider: &mockGitProvider{
			err: errors.New("git not available"),
		},
		Mode:    ModeDefault,
		NoColor: true,
	})

	input := &StdinData{
		Model:         &ModelInfo{Name: "claude-opus-4-5-20250514"},
		ContextWindow: &ContextWindowInfo{Used: 50000, Total: 200000},
		Cost:          &CostData{TotalUSD: 0.05},
	}

	got, err := builder.Build(context.Background(), makeStdinJSON(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should still have model and context, without git
	if !strings.Contains(got, "🤖 Opus 4.5") {
		t.Errorf("should contain model despite git failure, got %q", got)
	}
	if !strings.Contains(got, "🔋 ") {
		t.Errorf("should contain context despite git failure, got %q", got)
	}
	if !strings.Contains(got, "█") {
		t.Errorf("should contain bar graph characters, got %q", got)
	}
}

func TestBuilder_Build_AllProvidersFail(t *testing.T) {
	builder := New(Options{
		GitProvider: &mockGitProvider{
			err: errors.New("git failed"),
		},
		UpdateProvider: &mockUpdateProvider{
			err: errors.New("update failed"),
		},
		Mode:    ModeDefault,
		NoColor: true,
	})

	got, err := builder.Build(context.Background(), bytes.NewBufferString("{}"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should produce at least fallback output
	if got == "" {
		t.Error("all-failure case should still produce output")
	}
}

func TestBuilder_SetMode(t *testing.T) {
	builder := New(Options{
		GitProvider: &mockGitProvider{
			data: &GitStatusData{Branch: "main", Modified: 2, Available: true},
		},
		Mode:    ModeDefault,
		NoColor: true,
	})

	input := &StdinData{
		Model:         &ModelInfo{Name: "claude-sonnet-4-20250514"},
		ContextWindow: &ContextWindowInfo{Used: 50000, Total: 200000},
	}

	// Test in default mode
	gotDefault, err := builder.Build(context.Background(), makeStdinJSON(input))
	if err != nil {
		t.Fatalf("default mode build error: %v", err)
	}

	// Switch to minimal
	builder.SetMode(ModeMinimal)
	gotMinimal, err := builder.Build(context.Background(), makeStdinJSON(input))
	if err != nil {
		t.Fatalf("minimal mode build error: %v", err)
	}

	// Minimal should be different (no git info, no directory)
	if gotDefault == gotMinimal {
		t.Errorf("default and minimal output should differ:\ndefault: %q\nminimal: %q",
			gotDefault, gotMinimal)
	}

	// Minimal should have model name
	if !strings.Contains(gotMinimal, "Sonnet 4") {
		t.Errorf("minimal mode should contain model name, got %q", gotMinimal)
	}

	// Minimal should have context bar graph
	if !strings.Contains(gotMinimal, "🔋 ") {
		t.Errorf("minimal mode should contain context bar graph, got %q", gotMinimal)
	}
	if !strings.Contains(gotMinimal, "█") {
		t.Errorf("minimal mode should contain bar graph characters, got %q", gotMinimal)
	}
}

func TestBuilder_Build_NoNewline(t *testing.T) {
	builder := New(Options{
		GitProvider: &mockGitProvider{
			data: &GitStatusData{Branch: "main", Available: true},
		},
		Mode:    ModeDefault,
		NoColor: true,
	})

	input := &StdinData{
		Model:         &ModelInfo{Name: "claude-haiku-3-5-20241022"},
		ContextWindow: &ContextWindowInfo{Used: 50000, Total: 200000},
	}

	got, err := builder.Build(context.Background(), makeStdinJSON(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if strings.Contains(got, "\n") {
		t.Errorf("output should not contain newline, got %q", got)
	}
}

func TestBuilder_Build_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	builder := New(Options{
		GitProvider: &slowGitProvider{delay: 5 * time.Second},
		Mode:        ModeDefault,
		NoColor:     true,
	})

	input := &StdinData{
		Model:         &ModelInfo{Name: "claude-sonnet-4"},
		ContextWindow: &ContextWindowInfo{Used: 50000, Total: 200000},
	}

	got, err := builder.Build(ctx, makeStdinJSON(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should return partial data (context from stdin, no git)
	if got == "" {
		t.Error("cancelled context should still produce output")
	}
}

func TestBuilder_Build_MissingContextWindow(t *testing.T) {
	builder := New(Options{
		Mode:    ModeDefault,
		NoColor: true,
	})

	input := &StdinData{
		Model: &ModelInfo{Name: "claude-sonnet-4-20250514"},
		Cost:  &CostData{TotalUSD: 0.05},
	}

	got, err := builder.Build(context.Background(), makeStdinJSON(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should not panic and should not contain context info
	if strings.Contains(got, "🔋") {
		t.Errorf("should not contain context when missing, got %q", got)
	}
}

func TestBuilder_Build_MissingCost(t *testing.T) {
	builder := New(Options{
		Mode:    ModeDefault,
		NoColor: true,
	})

	input := &StdinData{
		Model:         &ModelInfo{Name: "claude-sonnet-4-20250514"},
		ContextWindow: &ContextWindowInfo{Used: 50000, Total: 200000},
	}

	got, err := builder.Build(context.Background(), makeStdinJSON(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should still have model and context
	if !strings.Contains(got, "Sonnet 4") {
		t.Errorf("should contain model, got %q", got)
	}
	if !strings.Contains(got, "🔋 ") {
		t.Errorf("should contain context, got %q", got)
	}
	if !strings.Contains(got, "█") {
		t.Errorf("should contain bar graph characters, got %q", got)
	}
}

func TestBuilder_DefaultMode(t *testing.T) {
	builder := New(Options{
		NoColor: true,
	})

	// Mode should default to ModeDefault
	got, err := builder.Build(context.Background(), bytes.NewBufferString("{}"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Just verify it doesn't panic
	if got == "" {
		t.Error("should produce output with empty mode")
	}
}
