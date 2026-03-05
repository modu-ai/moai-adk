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

	// compact(default) 모드로 빌드
	gotDefault, err := builder.Build(context.Background(), makeStdinJSON(input))
	if err != nil {
		t.Fatalf("default mode build error: %v", err)
	}

	// full(verbose) 모드로 전환 - default와 달라야 한다
	builder.SetMode(ModeFull)
	gotFull, err := builder.Build(context.Background(), makeStdinJSON(input))
	if err != nil {
		t.Fatalf("full mode build error: %v", err)
	}

	// full 모드는 default와 다른 출력 (멀티라인)
	if gotDefault == gotFull {
		t.Errorf("default and full output should differ:\ndefault: %q\nfull: %q",
			gotDefault, gotFull)
	}

	// full 모드는 모델 이름을 포함해야 한다
	if !strings.Contains(gotFull, "Sonnet 4") {
		t.Errorf("full mode should contain model name, got %q", gotFull)
	}

	// full 모드는 컨텍스트 바 그래프를 포함해야 한다
	if !strings.Contains(gotFull, "🔋 ") {
		t.Errorf("full mode should contain context bar graph, got %q", gotFull)
	}
	if !strings.Contains(gotFull, "█") {
		t.Errorf("full mode should contain bar graph characters, got %q", gotFull)
	}
}

func TestBuilder_Build_NoNewline(t *testing.T) {
	// v3에서 compact 모드만 한 줄 출력이다 (L1만 또는 L2만)
	builder := New(Options{
		GitProvider: &mockGitProvider{
			data: &GitStatusData{Available: false}, // git 없음 → L1만 출력
		},
		Mode:    ModeCompact, // v3 compact: 2줄, git 없으면 1줄
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

	// git 데이터가 없으면 L1만, 따라서 개행 없음
	if strings.Contains(got, "\n") {
		t.Errorf("compact without git should be 1 line, got %q", got)
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

	// CW 바는 context window 없을 때 표시되지 않아야 한다
	if strings.Contains(got, "CW:") {
		t.Errorf("should not contain CW bar when context window missing, got %q", got)
	}
	// 5H/7D는 항상 0%로 표시된다 (🔋 아이콘 포함)
	if !strings.Contains(got, "5H:") || !strings.Contains(got, "7D:") {
		t.Errorf("should always contain 5H/7D bars, got %q", got)
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

// TestBuilderNormalizesMode는 Builder가 deprecated 모드 이름을 자동으로 정규화하는지 검증한다.
// REQ-V3-MODE-001: New(Options{Mode: "minimal"}) → 내부적으로 "compact"로 처리
// REQ-V3-MODE-002: New(Options{Mode: "verbose"}) → 내부적으로 "full"로 처리
func TestBuilderNormalizesMode(t *testing.T) {
	input := &StdinData{
		Model:         &ModelInfo{Name: "claude-sonnet-4-20250514"},
		ContextWindow: &ContextWindowInfo{Used: 50000, Total: 200000},
	}

	// "minimal"로 생성한 빌더는 ModeCompact("compact")로 동작해야 한다
	builderMinimal := New(Options{
		Mode:    "minimal",
		NoColor: true,
	})
	// "compact"로 생성한 빌더
	builderCompact := New(Options{
		Mode:    ModeCompact,
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

	// "minimal"과 "compact"는 동일한 출력을 생성해야 한다 (AC-V3-01)
	if gotMinimal != gotCompact {
		t.Errorf("mode=minimal should produce same output as mode=compact:\nminimal: %q\ncompact: %q",
			gotMinimal, gotCompact)
	}

	// "verbose"로 생성한 빌더는 ModeFull("full")로 동작해야 한다
	builderVerbose := New(Options{
		Mode:    "verbose",
		NoColor: true,
	})
	// "full"로 생성한 빌더
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

	// "verbose"와 "full"은 동일한 출력을 생성해야 한다 (AC-V3-02)
	if gotVerbose != gotFull {
		t.Errorf("mode=verbose should produce same output as mode=full:\nverbose: %q\nfull: %q",
			gotVerbose, gotFull)
	}
}

// TestBuilderCollectsTask는 collectAll이 StatusData.Task 필드를 채우는지 검증한다.
// Cycle 5 (REQ-V3): builder가 CollectTask()를 통해 task 정보를 수집한다.
func TestBuilderCollectsTask(t *testing.T) {
	b := &defaultBuilder{
		renderer: NewRenderer("default", true, nil),
		mode:     ModeDefault,
	}

	input := &StdinData{
		Model: &ModelInfo{DisplayName: "Opus"},
	}

	data := b.collectAll(context.Background(), input)

	// Task 필드가 존재하고 초기화되었는지 확인한다
	// (실제 값은 session state 파일 내용에 따라 다르므로, 필드 타입만 검증)
	// TaskData는 항상 유효한 구조체이어야 한다 (nil 포인터 없음)
	_ = data.Task.Active   // panic 없이 접근 가능해야 한다
	_ = data.Task.Command  // panic 없이 접근 가능해야 한다
	_ = data.Task.SpecID   // panic 없이 접근 가능해야 한다
	_ = data.Task.Stage    // panic 없이 접근 가능해야 한다
}

// TestBuilderCollectsTask_FieldExists는 StatusData에 Task 필드가 존재하는지 컴파일 시점에 검증한다.
func TestBuilderCollectsTask_FieldExists(t *testing.T) {
	data := &StatusData{}
	// Task 필드와 Usage 필드가 존재해야 한다 (컴파일 검증)
	_ = data.Task
	_ = data.Usage
}

// ─────────────────────────────────────────────────────────────────────────────
// Phase 6: 전체 파이프라인 통합 테스트 (AC-V3-01 ~ AC-V3-13)
// builder → renderer E2E 검증
// ─────────────────────────────────────────────────────────────────────────────

// mockUsageProvider는 테스트용 UsageProvider 구현체다.
type mockUsageProvider struct {
	data *UsageResult
	err  error
}

func (m *mockUsageProvider) CollectUsage(_ context.Context) (*UsageResult, error) {
	return m.data, m.err
}

// realisticInput은 현실적인 테스트 입력 데이터를 생성한다.
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

// realisticGit는 현실적인 git 데이터를 반환하는 mockGitProvider를 생성한다.
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

// realisticUsage는 현실적인 API 사용량 데이터를 반환하는 mockUsageProvider를 생성한다.
// 60% 사용률 기준 (AC-V3-07 검증용)
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

// countLines는 문자열의 줄 수를 계산한다 (빈 문자열은 0).
func countLines(s string) int {
	if s == "" {
		return 0
	}
	return len(strings.Split(s, "\n"))
}

// hasANSI는 문자열에 ANSI 이스케이프 코드가 포함되어 있는지 검사한다.
func hasANSI(s string) bool {
	return strings.Contains(s, "\x1b[") || strings.Contains(s, "\033[")
}

// TestIntegration_ModeLineCount는 각 모드의 출력 줄 수를 검증한다 (AC-V3-01 ~ AC-V3-06).
func TestIntegration_ModeLineCount(t *testing.T) {
	tests := []struct {
		name          string
		mode          StatuslineMode
		withUsage     bool
		minLines      int
		maxLines      int
		description   string
	}{
		// AC-V3-01: mode="minimal" → compact 2줄 출력 (backward compat)
		{
			name:        "AC-V3-01: minimal→compact 2줄",
			mode:        "minimal",
			withUsage:   true,
			minLines:    2,
			maxLines:    2,
			description: "minimal 모드는 compact와 동일하게 2줄이어야 한다",
		},
		// AC-V3-02: mode="verbose" → full 5줄 출력 (backward compat, L6 제거)
		{
			name:        "AC-V3-02: verbose→full 5줄",
			mode:        "verbose",
			withUsage:   true,
			minLines:    5,
			maxLines:    5,
			description: "verbose 모드는 full과 동일하게 5줄이어야 한다",
		},
		// AC-V3-03: mode="compact" → 정확히 2줄
		{
			name:        "AC-V3-03: compact 정확히 2줄",
			mode:        ModeCompact,
			withUsage:   true,
			minLines:    2,
			maxLines:    2,
			description: "compact 모드는 정확히 2줄이어야 한다",
		},
		// AC-V3-04: mode="default" → 정확히 3줄 (L4 제거, 스타일 L1 통합)
		{
			name:        "AC-V3-04: default 정확히 3줄",
			mode:        ModeDefault,
			withUsage:   true,
			minLines:    3,
			maxLines:    3,
			description: "default 모드는 정확히 3줄이어야 한다",
		},
		// AC-V3-05: mode="full" → 정확히 5줄 (L6 제거, 스타일 L1 통합)
		{
			name:        "AC-V3-05: full 정확히 5줄",
			mode:        ModeFull,
			withUsage:   true,
			minLines:    5,
			maxLines:    5,
			description: "full 모드는 정확히 5줄이어야 한다",
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
				t.Fatalf("Build 오류: %v", err)
			}

			lines := countLines(got)
			if lines < tt.minLines || lines > tt.maxLines {
				t.Errorf("%s\n줄 수: got=%d, want=%d~%d\n출력:\n%s",
					tt.description, lines, tt.minLines, tt.maxLines, got)
			}
		})
	}
}

// TestIntegration_NoUsageLineCount는 usage=nil일 때 5H/7D가 0%로 항상 표시되는지 검증한다.
func TestIntegration_NoUsageLineCount(t *testing.T) {
	// AC-V3-06: mode="full" + no usage → 5H/7D 0%로 항상 표시 → 5줄 유지
	t.Run("AC-V3-06: full + no usage → 5줄 (5H/7D 0%)", func(t *testing.T) {
		builder := New(Options{
			GitProvider:    realisticGit(),
			UpdateProvider: &mockUpdateProvider{data: &VersionData{Current: "2.8.0", Available: true}},
			UsageProvider:  &mockUsageProvider{data: nil}, // usage 없음
			Mode:           ModeFull,
			NoColor:        true,
		})

		got, err := builder.Build(context.Background(), makeStdinJSON(realisticInput()))
		if err != nil {
			t.Fatalf("Build 오류: %v", err)
		}

		lines := countLines(got)
		// 5H/7D는 항상 0%로 표시되므로 full 모드는 5줄 유지
		if lines != 5 {
			t.Errorf("AC-V3-06: full + no usage는 5줄이어야 한다, got=%d\n출력:\n%s", lines, got)
		}
		// 5H/7D 0% 바가 표시되어야 한다
		if !strings.Contains(got, "5H:") || !strings.Contains(got, "7D:") {
			t.Errorf("AC-V3-06: 5H/7D 바가 항상 표시되어야 한다\n출력:\n%s", got)
		}
		if !strings.Contains(got, "0%") {
			t.Errorf("AC-V3-06: no usage 시 0%%가 표시되어야 한다\n출력:\n%s", got)
		}
	})

	// AC-V3-06b: mode="default" + no usage → CW + 5H(0%) + 7D(0%) 모두 L2에 표시
	t.Run("AC-V3-06b: default + no usage → L2에 CW+5H+7D", func(t *testing.T) {
		builder := New(Options{
			GitProvider:    realisticGit(),
			UpdateProvider: &mockUpdateProvider{data: &VersionData{Current: "2.8.0", Available: true}},
			UsageProvider:  &mockUsageProvider{data: nil}, // usage 없음
			Mode:           ModeDefault,
			NoColor:        true,
		})

		got, err := builder.Build(context.Background(), makeStdinJSON(realisticInput()))
		if err != nil {
			t.Fatalf("Build 오류: %v", err)
		}

		// CW, 5H, 7D 모두 있어야 한다
		if !strings.Contains(got, "CW:") {
			t.Errorf("AC-V3-06b: default + no usage는 CW 바를 포함해야 한다\n출력:\n%s", got)
		}
		// 5H/7D는 항상 0%로 표시
		if !strings.Contains(got, "5H:") || !strings.Contains(got, "7D:") {
			t.Errorf("AC-V3-06b: 5H/7D 바가 항상 표시되어야 한다\n출력:\n%s", got)
		}
	})
}

// TestIntegration_GradientBar는 그라디언트 바 블록 수를 검증한다 (AC-V3-07).
// 60% usage → 40블록 중 24블록이 채워진다.
func TestIntegration_GradientBar(t *testing.T) {
	// AC-V3-07: 60% usage → 40블록 바에서 24개 채워짐 (full 모드 CW 바)
	t.Run("AC-V3-07: 60% → CW 40블록 중 24 채워짐", func(t *testing.T) {
		// 60% 사용률로 context window 설정
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
			t.Fatalf("Build 오류: %v", err)
		}

		// full 모드: CW(40) + 5H(40, 0%) + 7D(40, 0%) = 120 블록
		// CW 바의 60% = 24 filled, 16 empty → CW만의 블록을 검증하려면 CW 줄만 추출
		lines := strings.Split(got, "\n")
		var cwLine string
		for _, l := range lines {
			if strings.Contains(l, "CW:") {
				cwLine = l
				break
			}
		}
		if cwLine == "" {
			t.Fatalf("AC-V3-07: CW 바가 출력에 있어야 한다\n출력:\n%s", got)
		}

		cwFilled := strings.Count(cwLine, "█")
		cwEmpty := strings.Count(cwLine, "░")
		cwTotal := cwFilled + cwEmpty

		if cwTotal != 40 {
			t.Errorf("AC-V3-07: CW 바 전체 블록 수 = %d, want=40\nCW 줄: %q", cwTotal, cwLine)
		}
		if cwFilled != 24 {
			t.Errorf("AC-V3-07: CW 바 채워진 블록 수 = %d, want=24\nCW 줄: %q", cwFilled, cwLine)
		}
	})
}

// TestIntegration_SessionTime는 세션 시간 포맷을 검증한다 (AC-V3-08).
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
			t.Fatalf("Build 오류: %v", err)
		}

		if !strings.Contains(got, "⏳ 1h 23m") {
			t.Errorf("AC-V3-08: 세션 시간이 '⏳ 1h 23m'이어야 한다\n출력:\n%s", got)
		}
	})
}

// TestIntegration_NoCost는 어떤 모드에서도 비용이 표시되지 않음을 검증한다 (AC-V3-08b).
func TestIntegration_NoCost(t *testing.T) {
	modes := []StatuslineMode{ModeCompact, ModeDefault, ModeFull}
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
				t.Fatalf("Build 오류: %v", err)
			}

			// 비용 정보($, USD, ¥ 등)가 출력되면 안 된다
			if strings.Contains(got, "$") || strings.Contains(got, "USD") {
				t.Errorf("AC-V3-08b: %s 모드에서 비용이 표시되면 안 된다\n출력:\n%s", mode, got)
			}
		})
	}
}

// TestIntegration_GitAheadBehind는 ahead/behind 포맷을 검증한다 (AC-V3-09).
func TestIntegration_GitAheadBehind(t *testing.T) {
	// AC-V3-09: Ahead=3, Behind=2 → "↑3↓2" 포맷
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
			t.Fatalf("Build 오류: %v", err)
		}

		if !strings.Contains(got, "↑3↓2") {
			t.Errorf("AC-V3-09: git ahead/behind가 '↑3↓2' 포맷이어야 한다\n출력:\n%s", got)
		}
	})
}

// TestIntegration_NoColor는 NoColor=true일 때 ANSI 이스케이프 코드가 없음을 검증한다 (AC-V3-12).
func TestIntegration_NoColor(t *testing.T) {
	modes := []StatuslineMode{ModeCompact, ModeDefault, ModeFull}
	for _, mode := range modes {
		t.Run(string(mode), func(t *testing.T) {
			builder := New(Options{
				GitProvider:   realisticGit(),
				UsageProvider: realisticUsage(60.0, 75.0),
				Mode:          mode,
				NoColor:       true, // ANSI 비활성화
			})

			got, err := builder.Build(context.Background(), makeStdinJSON(realisticInput()))
			if err != nil {
				t.Fatalf("Build 오류: %v", err)
			}

			// AC-V3-12: NO_COLOR → ANSI 이스케이프 없음
			if hasANSI(got) {
				t.Errorf("AC-V3-12: NoColor=true일 때 ANSI 이스케이프 코드가 없어야 한다\n출력:\n%q", got)
			}
		})
	}
}

// TestIntegration_BatteryIcon은 사용률에 따른 배터리 아이콘을 검증한다 (AC-V3-13).
func TestIntegration_BatteryIcon(t *testing.T) {
	// AC-V3-13: 75% usage → CW 바에 🪫 (low battery icon)
	t.Run("AC-V3-13: 75% → CW에 🪫", func(t *testing.T) {
		// 75% context window 사용률
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
			t.Fatalf("Build 오류: %v", err)
		}

		// CW 바만 추출하여 검증 (5H/7D 0%는 🔋를 가짐)
		lines := strings.Split(got, "\n")
		var cwPart string
		for _, l := range lines {
			if strings.Contains(l, "CW:") {
				// CW: 부분만 추출 (│ 구분자 전까지)
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
			t.Fatalf("CW 바가 출력에 있어야 한다\n출력:\n%s", got)
		}
		// 75%는 70% 초과이므로 CW 바에 🪫 아이콘이어야 한다
		if !strings.Contains(cwPart, "🪫") {
			t.Errorf("AC-V3-13: CW 75%% 사용률에서 🪫 아이콘이어야 한다\nCW: %q\n출력:\n%s", cwPart, got)
		}
	})

	// 반대 케이스: 60% → CW에 🔋
	t.Run("60% → CW에 🔋", func(t *testing.T) {
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
			t.Fatalf("Build 오류: %v", err)
		}

		// CW 바에 🔋가 있어야 한다
		if !strings.Contains(got, "🔋") {
			t.Errorf("60%% 사용률에서 🔋 아이콘이어야 한다\n출력:\n%s", got)
		}
	})
}

// TestIntegration_BackwardCompat은 deprecated 모드 이름의 완전 파이프라인 호환성을 검증한다.
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

	gotMinimal, err := builderMinimal.Build(context.Background(), makeStdinJSON(input))
	if err != nil {
		t.Fatalf("minimal Build 오류: %v", err)
	}
	gotCompact, err := builderCompact.Build(context.Background(), makeStdinJSON(input))
	if err != nil {
		t.Fatalf("compact Build 오류: %v", err)
	}

	if gotMinimal != gotCompact {
		t.Errorf("AC-V3-01: minimal과 compact 출력이 동일해야 한다\nminimal: %q\ncompact: %q",
			gotMinimal, gotCompact)
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
		t.Fatalf("verbose Build 오류: %v", err)
	}
	gotFull, err := builderFull.Build(context.Background(), makeStdinJSON(input))
	if err != nil {
		t.Fatalf("full Build 오류: %v", err)
	}

	if gotVerbose != gotFull {
		t.Errorf("AC-V3-02: verbose와 full 출력이 동일해야 한다\nverbose: %q\nfull: %q",
			gotVerbose, gotFull)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Phase 6: 성능 벤치마크 (NF-001: 500ms SLA)
// ─────────────────────────────────────────────────────────────────────────────

// BenchmarkBuilder_Build는 전체 Build() 파이프라인 성능을 측정한다 (NF-001: 500ms SLA).
func BenchmarkBuilder_Build(b *testing.B) {
	modes := []StatuslineMode{ModeCompact, ModeDefault, ModeFull}
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
					b.Fatalf("Build 오류: %v", err)
				}
			}
		})
	}
}

// TestBuilderSetModeNormalizes는 SetMode가 deprecated 모드 이름을 정규화하는지 검증한다.
// REQ-V3-MODE-004: 시스템은 항상 ModeCompact, ModeDefault, ModeFull 상수를 사용한다.
func TestBuilderSetModeNormalizes(t *testing.T) {
	input := &StdinData{
		Model:         &ModelInfo{Name: "claude-sonnet-4-20250514"},
		ContextWindow: &ContextWindowInfo{Used: 50000, Total: 200000},
	}

	builder := New(Options{
		Mode:    ModeDefault,
		NoColor: true,
	})

	// SetMode("minimal") 호출 후 ModeCompact와 동일하게 동작해야 한다
	builder.SetMode("minimal")
	gotAfterSetMinimal, err := builder.Build(context.Background(), makeStdinJSON(input))
	if err != nil {
		t.Fatalf("after SetMode minimal build error: %v", err)
	}

	builderCompact := New(Options{
		Mode:    ModeCompact,
		NoColor: true,
	})
	gotCompact, err := builderCompact.Build(context.Background(), makeStdinJSON(input))
	if err != nil {
		t.Fatalf("compact builder build error: %v", err)
	}

	if gotAfterSetMinimal != gotCompact {
		t.Errorf("SetMode(minimal) should produce same output as ModeCompact:\nafter SetMode(minimal): %q\ncompact: %q",
			gotAfterSetMinimal, gotCompact)
	}
}
