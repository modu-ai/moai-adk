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
	clearGLMEnv(t)
	t.Setenv("CLAUDE_AUTOCOMPACT_PCT_OVERRIDE", "100") // Disable CW scaling for this test
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
	clearGLMEnv(t)

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
	clearGLMEnv(t)

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

	// Build in default mode
	gotDefault, err := builder.Build(context.Background(), makeStdinJSON(input))
	if err != nil {
		t.Fatalf("default mode build error: %v", err)
	}

	// Switch to full(verbose) mode - should differ from default
	builder.SetMode(ModeFull)
	gotFull, err := builder.Build(context.Background(), makeStdinJSON(input))
	if err != nil {
		t.Fatalf("full mode build error: %v", err)
	}

	// Full mode should differ from default output (multiline)
	if gotDefault == gotFull {
		t.Errorf("default and full output should differ:\ndefault: %q\nfull: %q",
			gotDefault, gotFull)
	}

	// Full mode should contain model name
	if !strings.Contains(gotFull, "Sonnet 4") {
		t.Errorf("full mode should contain model name, got %q", gotFull)
	}

	// Full mode should contain context bar graph
	if !strings.Contains(gotFull, "🔋 ") {
		t.Errorf("full mode should contain context bar graph, got %q", gotFull)
	}
	if !strings.Contains(gotFull, "█") {
		t.Errorf("full mode should contain bar graph characters, got %q", gotFull)
	}
}

func TestBuilder_Build_NoNewline(t *testing.T) {
	// Default mode without git produces 2 lines (L1 info + L2 bars)
	builder := New(Options{
		GitProvider: &mockGitProvider{
			data: &GitStatusData{Available: false}, // no git → no L3
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

	// Without git data, default renders L1 (info) + L2 (bars) = 2 lines
	lines := strings.Split(got, "\n")
	if len(lines) != 2 {
		t.Errorf("default without git should be 2 lines, got %d lines: %q", len(lines), got)
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

	// CW bar should not be shown when context window is missing
	if strings.Contains(got, "CW:") {
		t.Errorf("should not contain CW bar when context window missing, got %q", got)
	}
	// 5H/7D always shown at 0% (with 🔋 icon)
	if !strings.Contains(got, "5H:") || !strings.Contains(got, "7D:") {
		t.Errorf("should always contain 5H/7D bars, got %q", got)
	}
}

func TestBuilder_Build_MissingCost(t *testing.T) {
	clearGLMEnv(t)

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

// TestBuilderNormalizesMode verifies that Builder automatically normalizes deprecated mode names.
// REQ-V3-MODE-001: New(Options{Mode: "minimal"}) → internally treated as "default"
// REQ-V3-MODE-002: New(Options{Mode: "verbose"}) → internally treated as "full"
// REQ-V3-MODE-003: New(Options{Mode: "compact"}) → internally treated as "default"
func TestBuilderNormalizesMode(t *testing.T) {
	input := &StdinData{
		Model:         &ModelInfo{Name: "claude-sonnet-4-20250514"},
		ContextWindow: &ContextWindowInfo{Used: 50000, Total: 200000},
	}

	// Builder created with "minimal" should behave as ModeDefault
	builderMinimal := New(Options{
		Mode:    "minimal",
		NoColor: true,
	})
	// Builder created with "compact" should also behave as ModeDefault
	builderCompact := New(Options{
		Mode:    ModeCompact,
		NoColor: true,
	})
	// Builder created with "default"
	builderDefault := New(Options{
		Mode:    ModeDefault,
		NoColor: true,
	})

	gotMinimal, err := builderMinimal.Build(context.Background(), makeStdinJSON(input))
	if err != nil {
		t.Fatalf("minimal builder build error: %v", err)
	}
	gotCompact, err := builderCompact.Build(context.Background(), makeStdinJSON(input))
	if err != nil {
		t.Fatalf("compact builder build error: %v", err)
	}
	gotDefault, err := builderDefault.Build(context.Background(), makeStdinJSON(input))
	if err != nil {
		t.Fatalf("default builder build error: %v", err)
	}

	// "minimal", "compact", and "default" should all produce identical output
	if gotMinimal != gotDefault {
		t.Errorf("mode=minimal should produce same output as mode=default:\nminimal: %q\ndefault: %q",
			gotMinimal, gotDefault)
	}
	if gotCompact != gotDefault {
		t.Errorf("mode=compact should produce same output as mode=default:\ncompact: %q\ndefault: %q",
			gotCompact, gotDefault)
	}

	// Builder created with "verbose" should behave as ModeFull("full")
	builderVerbose := New(Options{
		Mode:    "verbose",
		NoColor: true,
	})
	// Builder created with "full"
	builderFull := New(Options{
		Mode:    ModeFull,
		NoColor: true,
	})

	gotVerbose, err := builderVerbose.Build(context.Background(), makeStdinJSON(input))
	if err != nil {
		t.Fatalf("verbose builder build error: %v", err)
	}
	gotFull, err := builderFull.Build(context.Background(), makeStdinJSON(input))
	if err != nil {
		t.Fatalf("full builder build error: %v", err)
	}

	// "verbose" and "full" should produce identical output (AC-V3-02)
	if gotVerbose != gotFull {
		t.Errorf("mode=verbose should produce same output as mode=full:\nverbose: %q\nfull: %q",
			gotVerbose, gotFull)
	}
}

// TestBuilderCollectsTask verifies that collectAll populates the StatusData.Task field.
// Cycle 5 (REQ-V3): builder collects task info via CollectTask().
func TestBuilderCollectsTask(t *testing.T) {
	b := &defaultBuilder{
		renderer: NewRenderer("default", true, nil),
		mode:     ModeDefault,
	}

	input := &StdinData{
		Model: &ModelInfo{DisplayName: "Opus"},
	}

	data := b.collectAll(context.Background(), input)

	// Verify Task field exists and is initialized
	// (actual values depend on session state file, so only verify field types)
	// TaskData must always be a valid struct (no nil pointers)
	_ = data.Task.Active  // must be accessible without panic
	_ = data.Task.Command // must be accessible without panic
	_ = data.Task.SpecID  // must be accessible without panic
	_ = data.Task.Stage   // must be accessible without panic
}

// TestBuilderCollectsTask_FieldExists verifies Task field exists on StatusData at compile time.
func TestBuilderCollectsTask_FieldExists(t *testing.T) {
	data := &StatusData{}
	// Task and Usage fields must exist (compile-time verification)
	_ = data.Task
	_ = data.Usage
}

// ─────────────────────────────────────────────────────────────────────────────
// Phase 6: Full pipeline integration tests (AC-V3-01 ~ AC-V3-13)
// builder → renderer E2E verification
// ─────────────────────────────────────────────────────────────────────────────

// mockUsageProvider implements UsageProvider for testing.
type mockUsageProvider struct {
	data  *UsageResult
	err   error
	calls int
}

func (m *mockUsageProvider) CollectUsage(_ context.Context) (*UsageResult, error) {
	m.calls++
	return m.data, m.err
}

// realisticInput creates realistic test input data.
func realisticInput() *StdinData {
	return &StdinData{
		Model:         &ModelInfo{Name: "claude-opus-4-6-20250514"},
		Cost:          &CostData{TotalUSD: 0.15, TotalDurationMS: 4980000},
		ContextWindow: &ContextWindowInfo{Used: 120000, Total: 200000},
		CWD:           "/Users/test/moai-adk-go",
		OutputStyle:   &OutputStyleInfo{Name: "MoAI"},
		Version:       "2.1.69",
	}
}

// realisticGit creates a mockGitProvider with realistic git data.
func realisticGit() *mockGitProvider {
	return &mockGitProvider{
		data: &GitStatusData{
			Branch:    "feat/statusline-v3",
			Modified:  5,
			Staged:    3,
			Untracked: 1,
			Ahead:     3,
			Behind:    2,
			Available: true,
		},
	}
}

// realisticUsage creates a mockUsageProvider with realistic API usage data.
// Based on 60% usage (for AC-V3-07 verification)
func realisticUsage(pct5H, pct7D float64) *mockUsageProvider {
	return &mockUsageProvider{
		data: &UsageResult{
			Usage5H: &UsageData{
				UsedTokens:  int64(pct5H * 1000),
				LimitTokens: 100000,
				Percentage:  pct5H,
			},
			Usage7D: &UsageData{
				UsedTokens:  int64(pct7D * 1000),
				LimitTokens: 100000,
				Percentage:  pct7D,
			},
		},
	}
}

// countLines counts the number of lines in a string (empty string returns 0).
func countLines(s string) int {
	if s == "" {
		return 0
	}
	return len(strings.Split(s, "\n"))
}

// hasANSI checks if a string contains ANSI escape codes.
func hasANSI(s string) bool {
	return strings.Contains(s, "\x1b[") || strings.Contains(s, "\033[")
}

// TestIntegration_ModeLineCount verifies output line count for each mode (AC-V3-01 ~ AC-V3-06).
func TestIntegration_ModeLineCount(t *testing.T) {
	tests := []struct {
		name        string
		mode        StatuslineMode
		withUsage   bool
		minLines    int
		maxLines    int
		description string
	}{
		// AC-V3-01: mode="minimal" → default 3-line output (backward compat)
		{
			name:        "AC-V3-01: minimal→default 3 lines",
			mode:        "minimal",
			withUsage:   true,
			minLines:    3,
			maxLines:    3,
			description: "minimal mode should produce 3 lines like default",
		},
		// AC-V3-02: mode="verbose" → full 5-line output (backward compat)
		{
			name:        "AC-V3-02: verbose→full 5 lines",
			mode:        "verbose",
			withUsage:   true,
			minLines:    5,
			maxLines:    5,
			description: "verbose mode should produce 5 lines like full",
		},
		// AC-V3-03: mode="compact" → default 3-line output (backward compat)
		{
			name:        "AC-V3-03: compact→default 3 lines",
			mode:        ModeCompact,
			withUsage:   true,
			minLines:    3,
			maxLines:    3,
			description: "compact mode should produce 3 lines like default",
		},
		// AC-V3-04: mode="default" → exactly 3 lines (style integrated into L1)
		{
			name:        "AC-V3-04: default exactly 3 lines",
			mode:        ModeDefault,
			withUsage:   true,
			minLines:    3,
			maxLines:    3,
			description: "default mode should produce exactly 3 lines",
		},
		// AC-V3-05: mode="full" → exactly 5 lines (style integrated into L1)
		{
			name:        "AC-V3-05: full exactly 5 lines",
			mode:        ModeFull,
			withUsage:   true,
			minLines:    5,
			maxLines:    5,
			description: "full mode should produce exactly 5 lines",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var usageProv *mockUsageProvider
			if tt.withUsage {
				usageProv = realisticUsage(45.0, 82.0)
			}

			builder := New(Options{
				GitProvider:    realisticGit(),
				UpdateProvider: &mockUpdateProvider{data: &VersionData{Current: "2.8.0", Available: true}},
				UsageProvider:  usageProv,
				Mode:           tt.mode,
				NoColor:        true,
			})

			got, err := builder.Build(context.Background(), makeStdinJSON(realisticInput()))
			if err != nil {
				t.Fatalf("Build error: %v", err)
			}

			lines := countLines(got)
			if lines < tt.minLines || lines > tt.maxLines {
				t.Errorf("%s\nline count: got=%d, want=%d~%d\noutput:\n%s",
					tt.description, lines, tt.minLines, tt.maxLines, got)
			}
		})
	}
}

// TestIntegration_NoUsageLineCount verifies 5H/7D always shown at 0% when usage=nil.
func TestIntegration_NoUsageLineCount(t *testing.T) {
	// AC-V3-06: mode="full" + no usage → 5H/7D always shown at 0% → 5 lines
	t.Run("AC-V3-06: full + no usage → 5 lines (5H/7D 0%)", func(t *testing.T) {
		builder := New(Options{
			GitProvider:    realisticGit(),
			UpdateProvider: &mockUpdateProvider{data: &VersionData{Current: "2.8.0", Available: true}},
			UsageProvider:  &mockUsageProvider{data: nil}, // no usage
			Mode:           ModeFull,
			NoColor:        true,
		})

		got, err := builder.Build(context.Background(), makeStdinJSON(realisticInput()))
		if err != nil {
			t.Fatalf("Build error: %v", err)
		}

		lines := countLines(got)
		// 5H/7D always shown at 0%, so full mode keeps 5 lines
		if lines != 5 {
			t.Errorf("AC-V3-06: full + no usage should be 5 lines, got=%d\noutput:\n%s", lines, got)
		}
		// 5H/7D 0% bars must be shown
		if !strings.Contains(got, "5H:") || !strings.Contains(got, "7D:") {
			t.Errorf("AC-V3-06: 5H/7D bars must always be shown\noutput:\n%s", got)
		}
		if !strings.Contains(got, "0%") {
			t.Errorf("AC-V3-06: should show 0%% when no usage\noutput:\n%s", got)
		}
	})

	// AC-V3-06b: mode="default" + no usage → CW + 5H(0%) + 7D(0%) all shown in L2
	t.Run("AC-V3-06b: default + no usage → L2 CW+5H+7D", func(t *testing.T) {
		builder := New(Options{
			GitProvider:    realisticGit(),
			UpdateProvider: &mockUpdateProvider{data: &VersionData{Current: "2.8.0", Available: true}},
			UsageProvider:  &mockUsageProvider{data: nil}, // no usage
			Mode:           ModeDefault,
			NoColor:        true,
		})

		got, err := builder.Build(context.Background(), makeStdinJSON(realisticInput()))
		if err != nil {
			t.Fatalf("Build error: %v", err)
		}

		// CW, 5H, 7D must all be present
		if !strings.Contains(got, "CW:") {
			t.Errorf("AC-V3-06b: default + no usage should contain CW bar\noutput:\n%s", got)
		}
		// 5H/7D always shown at 0%
		if !strings.Contains(got, "5H:") || !strings.Contains(got, "7D:") {
			t.Errorf("AC-V3-06b: 5H/7D bars must always be shown\noutput:\n%s", got)
		}
	})
}

// TestIntegration_GradientBar verifies gradient bar block counts (AC-V3-07).
// 60% usage → 24 of 40 blocks filled.
func TestIntegration_GradientBar(t *testing.T) {
	t.Setenv("CLAUDE_AUTOCOMPACT_PCT_OVERRIDE", "100") // Disable CW scaling for exact block count tests
	// AC-V3-07: 60% usage → 24 filled in 40-block bar (full mode CW bar)
	t.Run("AC-V3-07: 60% → 24 of 40 CW blocks filled", func(t *testing.T) {
		// Set context window to 60% usage
		input := &StdinData{
			Model:         &ModelInfo{Name: "claude-opus-4-6-20250514"},
			ContextWindow: &ContextWindowInfo{Used: 60000, Total: 100000},
			CWD:           "/Users/test/project",
		}

		builder := New(Options{
			UsageProvider: &mockUsageProvider{data: nil},
			Mode:          ModeFull,
			NoColor:       true,
		})

		got, err := builder.Build(context.Background(), makeStdinJSON(input))
		if err != nil {
			t.Fatalf("Build error: %v", err)
		}

		// full mode: CW(40) + 5H(40, 0%) + 7D(40, 0%) = 120 blocks
		// CW bar 60% = 24 filled, 16 empty → extract CW line only for verification
		lines := strings.Split(got, "\n")
		var cwLine string
		for _, l := range lines {
			if strings.Contains(l, "CW:") {
				cwLine = l
				break
			}
		}
		if cwLine == "" {
			t.Fatalf("AC-V3-07: CW bar must be in output\noutput:\n%s", got)
		}

		cwFilled := strings.Count(cwLine, "█")
		cwEmpty := strings.Count(cwLine, "░")
		cwTotal := cwFilled + cwEmpty

		if cwTotal != 40 {
			t.Errorf("AC-V3-07: CW bar total blocks = %d, want=40\nCW line: %q", cwTotal, cwLine)
		}
		if cwFilled != 24 {
			t.Errorf("AC-V3-07: CW bar filled blocks = %d, want=24\nCW line: %q", cwFilled, cwLine)
		}
	})
}

// TestIntegration_SessionTime verifies session time format (AC-V3-08).
func TestIntegration_SessionTime(t *testing.T) {
	// AC-V3-08: TotalDurationMS=4980000 → "⏳ 1h 23m"
	// 4980000 ms = 4980 seconds = 83 minutes = 1h 23m
	t.Run("AC-V3-08: 4980000ms → ⏳ 1h 23m", func(t *testing.T) {
		input := &StdinData{
			Model: &ModelInfo{Name: "claude-opus-4-6-20250514"},
			Cost:  &CostData{TotalDurationMS: 4980000},
		}

		builder := New(Options{
			UsageProvider: &mockUsageProvider{data: nil},
			Mode:          ModeDefault,
			NoColor:       true,
		})

		got, err := builder.Build(context.Background(), makeStdinJSON(input))
		if err != nil {
			t.Fatalf("Build error: %v", err)
		}

		if !strings.Contains(got, "⏳ 1h 23m") {
			t.Errorf("AC-V3-08: session time should be '⏳ 1h 23m'\noutput:\n%s", got)
		}
	})
}

// TestIntegration_NoCost verifies no cost is shown in any mode (AC-V3-08b).
func TestIntegration_NoCost(t *testing.T) {
	modes := []StatuslineMode{ModeDefault, ModeFull}
	for _, mode := range modes {
		t.Run(string(mode), func(t *testing.T) {
			input := &StdinData{
				Model: &ModelInfo{Name: "claude-opus-4-6-20250514"},
				Cost:  &CostData{TotalUSD: 1.5},
			}

			builder := New(Options{
				UsageProvider: &mockUsageProvider{data: nil},
				Mode:          mode,
				NoColor:       true,
			})

			got, err := builder.Build(context.Background(), makeStdinJSON(input))
			if err != nil {
				t.Fatalf("Build error: %v", err)
			}

			// Cost info ($, USD, etc.) must not be in output
			if strings.Contains(got, "$") || strings.Contains(got, "USD") {
				t.Errorf("AC-V3-08b: cost must not be shown in %s mode\noutput:\n%s", mode, got)
			}
		})
	}
}

// TestIntegration_GitAheadBehind verifies ahead/behind format (AC-V3-09).
func TestIntegration_GitAheadBehind(t *testing.T) {
	// AC-V3-09: Ahead=3, Behind=2 → "↑3↓2" format
	t.Run("AC-V3-09: Ahead=3, Behind=2 → ↑3↓2", func(t *testing.T) {
		builder := New(Options{
			GitProvider: &mockGitProvider{
				data: &GitStatusData{
					Branch:    "feat/test",
					Ahead:     3,
					Behind:    2,
					Available: true,
				},
			},
			UsageProvider: &mockUsageProvider{data: nil},
			Mode:          ModeDefault,
			NoColor:       true,
		})

		got, err := builder.Build(context.Background(), makeStdinJSON(&StdinData{
			Model: &ModelInfo{Name: "claude-sonnet-4-20250514"},
		}))
		if err != nil {
			t.Fatalf("Build error: %v", err)
		}

		if !strings.Contains(got, "↑3↓2") {
			t.Errorf("AC-V3-09: git ahead/behind should be in '↑3↓2' format\noutput:\n%s", got)
		}
	})
}

// TestIntegration_NoColor verifies no ANSI escape codes when NoColor=true (AC-V3-12).
func TestIntegration_NoColor(t *testing.T) {
	modes := []StatuslineMode{ModeDefault, ModeFull}
	for _, mode := range modes {
		t.Run(string(mode), func(t *testing.T) {
			builder := New(Options{
				GitProvider:   realisticGit(),
				UsageProvider: realisticUsage(60.0, 75.0),
				Mode:          mode,
				NoColor:       true, // ANSI disabled
			})

			got, err := builder.Build(context.Background(), makeStdinJSON(realisticInput()))
			if err != nil {
				t.Fatalf("Build error: %v", err)
			}

			// AC-V3-12: NO_COLOR → no ANSI escapes
			if hasANSI(got) {
				t.Errorf("AC-V3-12: NoColor=true should have no ANSI escape codes\noutput:\n%q", got)
			}
		})
	}
}

// TestIntegration_BatteryIcon verifies battery icon based on usage percentage (AC-V3-13).
func TestIntegration_BatteryIcon(t *testing.T) {
	// AC-V3-13: 75% usage → CW bar shows 🪫 (low battery icon)
	t.Run("AC-V3-13: 75% → CW 🪫", func(t *testing.T) {
		// 75% context window usage
		input := &StdinData{
			Model:         &ModelInfo{Name: "claude-opus-4-6-20250514"},
			ContextWindow: &ContextWindowInfo{Used: 75000, Total: 100000},
			CWD:           "/Users/test/project",
		}

		builder := New(Options{
			UsageProvider: &mockUsageProvider{data: nil},
			Mode:          ModeDefault,
			NoColor:       true,
		})

		got, err := builder.Build(context.Background(), makeStdinJSON(input))
		if err != nil {
			t.Fatalf("Build error: %v", err)
		}

		// Extract CW bar only for verification (5H/7D 0% have 🔋)
		lines := strings.Split(got, "\n")
		var cwPart string
		for _, l := range lines {
			if strings.Contains(l, "CW:") {
				// Extract CW: part only (before │ separator)
				parts := strings.Split(l, "│")
				for _, p := range parts {
					if strings.Contains(p, "CW:") {
						cwPart = p
						break
					}
				}
				break
			}
		}

		if cwPart == "" {
			t.Fatalf("CW bar must be in output\noutput:\n%s", got)
		}
		// 75% > 70% threshold, so CW bar should show 🪫 icon
		if !strings.Contains(cwPart, "🪫") {
			t.Errorf("AC-V3-13: CW 75%% usage should show 🪫 icon\nCW: %q\noutput:\n%s", cwPart, got)
		}
	})

	// Opposite case: 60% → CW shows 🔋
	t.Run("60% → CW 🔋", func(t *testing.T) {
		input := &StdinData{
			Model:         &ModelInfo{Name: "claude-opus-4-6-20250514"},
			ContextWindow: &ContextWindowInfo{Used: 60000, Total: 100000},
		}

		builder := New(Options{
			UsageProvider: &mockUsageProvider{data: nil},
			Mode:          ModeDefault,
			NoColor:       true,
		})

		got, err := builder.Build(context.Background(), makeStdinJSON(input))
		if err != nil {
			t.Fatalf("Build error: %v", err)
		}

		// CW bar should have 🔋
		if !strings.Contains(got, "🔋") {
			t.Errorf("60%% usage should show 🔋 icon\noutput:\n%s", got)
		}
	})
}

// TestIntegration_BackwardCompat verifies full pipeline compatibility for deprecated mode names.
func TestIntegration_BackwardCompat(t *testing.T) {
	input := realisticInput()

	builderMinimal := New(Options{
		GitProvider:   realisticGit(),
		UsageProvider: realisticUsage(45.0, 60.0),
		Mode:          "minimal",
		NoColor:       true,
	})
	builderCompact := New(Options{
		GitProvider:   realisticGit(),
		UsageProvider: realisticUsage(45.0, 60.0),
		Mode:          ModeCompact,
		NoColor:       true,
	})
	builderDefault := New(Options{
		GitProvider:   realisticGit(),
		UsageProvider: realisticUsage(45.0, 60.0),
		Mode:          ModeDefault,
		NoColor:       true,
	})

	gotMinimal, err := builderMinimal.Build(context.Background(), makeStdinJSON(input))
	if err != nil {
		t.Fatalf("minimal Build error: %v", err)
	}
	gotCompact, err := builderCompact.Build(context.Background(), makeStdinJSON(input))
	if err != nil {
		t.Fatalf("compact Build error: %v", err)
	}
	gotDefault, err := builderDefault.Build(context.Background(), makeStdinJSON(input))
	if err != nil {
		t.Fatalf("default Build error: %v", err)
	}

	// minimal, compact, and default should all produce identical output
	if gotMinimal != gotDefault {
		t.Errorf("minimal and default output must be identical\nminimal: %q\ndefault: %q",
			gotMinimal, gotDefault)
	}
	if gotCompact != gotDefault {
		t.Errorf("compact and default output must be identical\ncompact: %q\ndefault: %q",
			gotCompact, gotDefault)
	}

	builderVerbose := New(Options{
		GitProvider:   realisticGit(),
		UsageProvider: realisticUsage(45.0, 60.0),
		Mode:          "verbose",
		NoColor:       true,
	})
	builderFull := New(Options{
		GitProvider:   realisticGit(),
		UsageProvider: realisticUsage(45.0, 60.0),
		Mode:          ModeFull,
		NoColor:       true,
	})

	gotVerbose, err := builderVerbose.Build(context.Background(), makeStdinJSON(input))
	if err != nil {
		t.Fatalf("verbose Build error: %v", err)
	}
	gotFull, err := builderFull.Build(context.Background(), makeStdinJSON(input))
	if err != nil {
		t.Fatalf("full Build error: %v", err)
	}

	if gotVerbose != gotFull {
		t.Errorf("AC-V3-02: verbose and full output must be identical\nverbose: %q\nfull: %q",
			gotVerbose, gotFull)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Phase 6: Performance benchmark (NF-001: 500ms SLA)
// ─────────────────────────────────────────────────────────────────────────────

// BenchmarkBuilder_Build measures full Build() pipeline performance (NF-001: 500ms SLA).
func BenchmarkBuilder_Build(b *testing.B) {
	modes := []StatuslineMode{ModeDefault, ModeFull}
	input := realisticInput()

	for _, mode := range modes {
		b.Run(string(mode), func(b *testing.B) {
			builder := New(Options{
				GitProvider:    realisticGit(),
				UpdateProvider: &mockUpdateProvider{data: &VersionData{Current: "2.8.0", Available: true}},
				UsageProvider:  realisticUsage(60.0, 75.0),
				Mode:           mode,
				NoColor:        true,
			})

			b.ResetTimer()
			for range b.N {
				_, err := builder.Build(context.Background(), makeStdinJSON(input))
				if err != nil {
					b.Fatalf("Build error: %v", err)
				}
			}
		})
	}
}

// TestBuild_OfficialSchema verifies that the exact official Claude Code statusline
// JSON schema (with integer resets_at) is parsed and rendered correctly.
//
// See Issue #549.
func TestBuild_OfficialSchema(t *testing.T) {
	clearGLMEnv(t)
	t.Setenv("CLAUDE_AUTOCOMPACT_PCT_OVERRIDE", "100")

	builder := New(Options{
		Mode:    ModeDefault,
		NoColor: true,
	})

	// Use the EXACT JSON from the Claude Code official docs
	officialJSON := `{
		"model": {"id": "claude-opus-4-6", "display_name": "Opus"},
		"context_window": {"used_percentage": 25, "context_window_size": 200000},
		"rate_limits": {
			"five_hour": {"used_percentage": 23.5, "resets_at": 1738425600},
			"seven_day": {"used_percentage": 41.2, "resets_at": 1738857600}
		},
		"workspace": {"current_dir": "/home/user/project", "project_dir": "/home/user/project"},
		"version": "1.0.80",
		"cost": {"total_cost_usd": 0.05}
	}`

	got, err := builder.Build(context.Background(), strings.NewReader(officialJSON))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Model should appear in the output
	if !strings.Contains(got, "Opus") {
		t.Errorf("output should contain model display name 'Opus', got:\n%s", got)
	}

	// Claude Code version should appear
	if !strings.Contains(got, "1.0.80") {
		t.Errorf("output should contain claude version '1.0.80', got:\n%s", got)
	}

	// Directory should appear
	if !strings.Contains(got, "project") {
		t.Errorf("output should contain directory 'project', got:\n%s", got)
	}

	// Usage bars should appear - 5H and 7D
	if !strings.Contains(got, "5H:") {
		t.Errorf("output should contain '5H:' bar, got:\n%s", got)
	}
	if !strings.Contains(got, "7D:") {
		t.Errorf("output should contain '7D:' bar, got:\n%s", got)
	}

	// 5H percentage should appear (23%)
	if !strings.Contains(got, "23%") {
		t.Errorf("output should contain 5H usage '23%%', got:\n%s", got)
	}
}

// TestBuild_RateLimitsPreferredOverUsage verifies that in default mode,
// data.RateLimits (from Claude Code stdin JSON) is preferred over data.Usage
// (from MoAI API call) for the 5H/7D bars.
//
// See Issue #549 Bug 3.
func TestBuild_RateLimitsPreferredOverUsage(t *testing.T) {
	t.Setenv("CLAUDE_AUTOCOMPACT_PCT_OVERRIDE", "100")

	// Mock usage provider returns 99% (should be ignored when RateLimits present)
	mockUsage := &mockUsageProvider{
		data: &UsageResult{
			Usage5H: &UsageData{Percentage: 99},
			Usage7D: &UsageData{Percentage: 88},
		},
	}

	builder := New(Options{
		Mode:          ModeDefault,
		NoColor:       true,
		UsageProvider: mockUsage,
	})

	// JSON with rate_limits at 23% - should take priority over the 99% from UsageProvider
	inputJSON := `{
		"rate_limits": {
			"five_hour": {"used_percentage": 23.5, "resets_at": 1738425600},
			"seven_day": {"used_percentage": 41.2, "resets_at": 1738857600}
		}
	}`

	got, err := builder.Build(context.Background(), strings.NewReader(inputJSON))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 23% should appear (from RateLimits), not 99% (from Usage)
	if !strings.Contains(got, "23%") {
		t.Errorf("output should contain 5H rate limit '23%%' from RateLimits (not 99%% from Usage), got:\n%s", got)
	}
	if strings.Contains(got, "99%") {
		t.Errorf("output should NOT contain '99%%' from Usage when RateLimits is present, got:\n%s", got)
	}
}

// TestBuild_SkipsUsageProviderWhenRateLimitsPresent verifies that when stdin
// provides rate_limits (Claude Code 2.1.80+), the usage provider (which would
// perform blocking Anthropic OAuth API calls) is not invoked at all.
//
// See Issue #646 — statusline intermittent disappearance caused by 5s HTTP
// timeout on every usage collector cache miss.
func TestBuild_SkipsUsageProviderWhenRateLimitsPresent(t *testing.T) {
	mockUsage := &mockUsageProvider{
		data: &UsageResult{
			Usage5H: &UsageData{Percentage: 50},
			Usage7D: &UsageData{Percentage: 50},
		},
	}

	builder := New(Options{
		Mode:          ModeDefault,
		NoColor:       true,
		UsageProvider: mockUsage,
	})

	inputJSON := `{
		"rate_limits": {
			"five_hour": {"used_percentage": 23.5, "resets_at": 1738425600},
			"seven_day": {"used_percentage": 41.2, "resets_at": 1738857600}
		}
	}`

	if _, err := builder.Build(context.Background(), strings.NewReader(inputJSON)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if mockUsage.calls != 0 {
		t.Errorf("usage provider should not be called when stdin rate_limits is present, got %d calls", mockUsage.calls)
	}
}

// TestBuild_CallsUsageProviderWhenRateLimitsAbsent verifies the usage provider
// is still invoked when stdin does not contain rate_limits (legacy Claude Code
// or non-Claude-Code harness).
func TestBuild_CallsUsageProviderWhenRateLimitsAbsent(t *testing.T) {
	mockUsage := &mockUsageProvider{
		data: &UsageResult{
			Usage5H: &UsageData{Percentage: 50},
			Usage7D: &UsageData{Percentage: 50},
		},
	}

	builder := New(Options{
		Mode:          ModeDefault,
		NoColor:       true,
		UsageProvider: mockUsage,
	})

	// Stdin without rate_limits
	inputJSON := `{"model": {"id": "claude-opus-4-6"}}`

	if _, err := builder.Build(context.Background(), strings.NewReader(inputJSON)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if mockUsage.calls != 1 {
		t.Errorf("usage provider should be called exactly once when stdin lacks rate_limits, got %d calls", mockUsage.calls)
	}
}

// TestBuilderSetModeNormalizes verifies SetMode normalizes deprecated mode names.
// REQ-V3-MODE-004: system always uses ModeDefault, ModeFull constants.
func TestBuilderSetModeNormalizes(t *testing.T) {
	input := &StdinData{
		Model:         &ModelInfo{Name: "claude-sonnet-4-20250514"},
		ContextWindow: &ContextWindowInfo{Used: 50000, Total: 200000},
	}

	builder := New(Options{
		Mode:    ModeFull,
		NoColor: true,
	})

	// After SetMode("minimal"), should behave same as ModeDefault
	builder.SetMode("minimal")
	gotAfterSetMinimal, err := builder.Build(context.Background(), makeStdinJSON(input))
	if err != nil {
		t.Fatalf("after SetMode minimal build error: %v", err)
	}

	builderDefault := New(Options{
		Mode:    ModeDefault,
		NoColor: true,
	})
	gotDefault, err := builderDefault.Build(context.Background(), makeStdinJSON(input))
	if err != nil {
		t.Fatalf("default builder build error: %v", err)
	}

	if gotAfterSetMinimal != gotDefault {
		t.Errorf("SetMode(minimal) should produce same output as ModeDefault:\nafter SetMode(minimal): %q\ndefault: %q",
			gotAfterSetMinimal, gotDefault)
	}
}

// TestCollectAll_ExtractsEffortThinking verifies that collectAll correctly maps
// effort and thinking from stdin to StatusData fields.
// REQ-CC2122-001: collectAll passes StdinData.Effort → StatusData.Effort
// REQ-CC2122-002: collectAll passes StdinData.Thinking → StatusData.Thinking
// REQ-CC2122-003: nil input → StatusData.Effort/Thinking remain nil
func TestCollectAll_ExtractsEffortThinking(t *testing.T) {
	tests := []struct {
		name         string
		input        *StdinData
		wantEffort   *EffortInfo
		wantThinking *ThinkingInfo
	}{
		{
			name: "effort level present: propagated to StatusData",
			input: &StdinData{
				Effort: &EffortInfo{Level: "high"},
			},
			wantEffort:   &EffortInfo{Level: "high"},
			wantThinking: nil,
		},
		{
			name: "thinking enabled: propagated to StatusData",
			input: &StdinData{
				Thinking: &ThinkingInfo{Enabled: true},
			},
			wantEffort:   nil,
			wantThinking: &ThinkingInfo{Enabled: true},
		},
		{
			name: "both present: both propagated",
			input: &StdinData{
				Effort:   &EffortInfo{Level: "max"},
				Thinking: &ThinkingInfo{Enabled: true},
			},
			wantEffort:   &EffortInfo{Level: "max"},
			wantThinking: &ThinkingInfo{Enabled: true},
		},
		{
			name:         "both absent: StatusData fields remain nil",
			input:        &StdinData{},
			wantEffort:   nil,
			wantThinking: nil,
		},
		{
			name:         "nil input: StatusData fields remain nil",
			input:        nil,
			wantEffort:   nil,
			wantThinking: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &defaultBuilder{
				renderer: NewRenderer("default", true, nil),
				mode:     ModeDefault,
			}
			data := b.collectAll(context.Background(), tt.input)

			// Verify Effort field
			if tt.wantEffort == nil {
				if data.Effort != nil {
					t.Errorf("Effort = %+v, want nil", data.Effort)
				}
			} else {
				if data.Effort == nil {
					t.Fatalf("Effort is nil, want %+v", tt.wantEffort)
				}
				if data.Effort.Level != tt.wantEffort.Level {
					t.Errorf("Effort.Level = %q, want %q", data.Effort.Level, tt.wantEffort.Level)
				}
			}

			// Verify Thinking field
			if tt.wantThinking == nil {
				if data.Thinking != nil {
					t.Errorf("Thinking = %+v, want nil", data.Thinking)
				}
			} else {
				if data.Thinking == nil {
					t.Fatalf("Thinking is nil, want %+v", tt.wantThinking)
				}
				if data.Thinking.Enabled != tt.wantThinking.Enabled {
					t.Errorf("Thinking.Enabled = %v, want %v", data.Thinking.Enabled, tt.wantThinking.Enabled)
				}
			}
		})
	}
}

// TestCollectAll_ExtractsWorktree verifies that collectAll correctly extracts
// git_worktree from workspace input data.
// REQ-CC297-003: collectAll passes WorkspaceInfo.GitWorktree to StatusData.Worktree
func TestCollectAll_ExtractsWorktree(t *testing.T) {
	tests := []struct {
		name      string
		input     *StdinData
		wantWT    string
	}{
		{
			name: "worktree path present: stored in StatusData.Worktree",
			input: &StdinData{
				Workspace: &WorkspaceInfo{
					CurrentDir: "/repo/.claude/worktrees/abc123",
					ProjectDir: "/repo",
					GitWorktree: "/repo/.claude/worktrees/abc123",
				},
			},
			wantWT: "/repo/.claude/worktrees/abc123",
		},
		{
			name: "no worktree: empty string",
			input: &StdinData{
				Workspace: &WorkspaceInfo{
					CurrentDir: "/repo",
					ProjectDir: "/repo",
				},
			},
			wantWT: "",
		},
		{
			name:   "workspace nil: empty string",
			input:  &StdinData{},
			wantWT: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &defaultBuilder{
				renderer: NewRenderer("default", true, nil),
				mode:     ModeDefault,
			}
			data := b.collectAll(context.Background(), tt.input)
			if data.Worktree != tt.wantWT {
				t.Errorf("Worktree = %q, want %q", data.Worktree, tt.wantWT)
			}
		})
	}
}
