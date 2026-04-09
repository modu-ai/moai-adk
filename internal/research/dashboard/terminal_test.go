package dashboard

import (
	"regexp"
	"strings"
	"testing"
)

// ansiRe는 ANSI 이스케이프 시퀀스를 매칭하는 정규식이다.
var ansiRe = regexp.MustCompile(`\x1b\[[0-9;]*m`)

// stripANSI는 문자열에서 ANSI 이스케이프 코드를 제거한다.
func stripANSI(s string) string { return ansiRe.ReplaceAllString(s, "") }

func TestRenderDashboard(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		data     *DashboardData
		contains []string // 출력에 포함되어야 하는 문자열
		absent   []string // 출력에 포함되지 않아야 하는 문자열
	}{
		{
			name: "전체 통과 데이터는 100%를 포함",
			data: &DashboardData{
				Target:         "eval-quality",
				Baseline:       0.80,
				CurrentScore:   1.0,
				TargetScore:    0.95,
				Experiments:    10,
				MaxExperiments: 20,
				KeepCount:      8,
				DiscardCount:   2,
				PerCriterion: []CriterionStatus{
					{Name: "accuracy", PassRate: 1.0, Weight: "MUST"},
					{Name: "speed", PassRate: 1.0},
				},
			},
			contains: []string{"100%", "eval-quality", "10/20"},
		},
		{
			name: "혼합 데이터는 올바른 백분율 포함",
			data: &DashboardData{
				Target:         "mixed-test",
				Baseline:       0.50,
				CurrentScore:   0.75,
				TargetScore:    0.90,
				Experiments:    5,
				MaxExperiments: 20,
				KeepCount:      3,
				DiscardCount:   2,
				PerCriterion: []CriterionStatus{
					{Name: "accuracy", PassRate: 0.80, Weight: "MUST"},
					{Name: "speed", PassRate: 0.60},
				},
			},
			contains: []string{"75%", "80%", "60%", "5/20"},
		},
		{
			name: "실험 0건은 0/20 표시",
			data: &DashboardData{
				Target:         "zero-exp",
				Baseline:       0.50,
				CurrentScore:   0.50,
				TargetScore:    0.90,
				Experiments:    0,
				MaxExperiments: 20,
				KeepCount:      0,
				DiscardCount:   0,
			},
			contains: []string{"0/20"},
		},
		{
			name: "점수 개선 시 + 접두사 포함",
			data: &DashboardData{
				Target:         "improving",
				Baseline:       0.60,
				CurrentScore:   0.85,
				TargetScore:    0.90,
				Experiments:    3,
				MaxExperiments: 10,
				KeepCount:      2,
				DiscardCount:   1,
			},
			contains: []string{"+"},
		},
		{
			name: "점수 퇴보 시 - 접두사 포함",
			data: &DashboardData{
				Target:         "regressing",
				Baseline:       0.80,
				CurrentScore:   0.65,
				TargetScore:    0.90,
				Experiments:    3,
				MaxExperiments: 10,
				KeepCount:      1,
				DiscardCount:   2,
			},
			contains: []string{"-"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := RenderDashboard(tt.data)
			stripped := stripANSI(result)

			for _, want := range tt.contains {
				if !strings.Contains(stripped, want) {
					t.Errorf("출력에 %q가 포함되어야 하지만 없음.\n출력:\n%s", want, stripped)
				}
			}
			for _, absent := range tt.absent {
				if strings.Contains(stripped, absent) {
					t.Errorf("출력에 %q가 없어야 하지만 있음.\n출력:\n%s", absent, stripped)
				}
			}
		})
	}
}

func TestRenderCompact(t *testing.T) {
	t.Parallel()

	data := &DashboardData{
		Target:         "compact-test",
		Baseline:       0.50,
		CurrentScore:   0.75,
		TargetScore:    0.90,
		Experiments:    5,
		MaxExperiments: 20,
		KeepCount:      3,
		DiscardCount:   2,
	}

	result := RenderCompact(data)
	stripped := stripANSI(result)

	// 단일 라인이어야 함 (마지막 개행 제외)
	trimmed := strings.TrimRight(stripped, "\n")
	if strings.Contains(trimmed, "\n") {
		t.Errorf("RenderCompact는 단일 라인이어야 하지만 개행 포함:\n%s", trimmed)
	}

	// 필수 요소 확인
	for _, want := range []string{"Research:", "compact-test", "75%", "5/20", "3K", "2D"} {
		if !strings.Contains(stripped, want) {
			t.Errorf("컴팩트 출력에 %q가 포함되어야 함. 출력: %s", want, stripped)
		}
	}
}

func TestRenderProgressBar(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		ratio       float64
		width       int
		wantFilled  int
		wantEmpty   int
	}{
		{
			name:       "100% 채움",
			ratio:      1.0,
			width:      25,
			wantFilled: 25,
			wantEmpty:  0,
		},
		{
			name:       "0% 비움",
			ratio:      0.0,
			width:      25,
			wantFilled: 0,
			wantEmpty:  25,
		},
		{
			name:       "50% 절반",
			ratio:      0.5,
			width:      25,
			wantFilled: 12, // int(0.5 * 25) = 12
			wantEmpty:  13,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := renderProgressBar(tt.ratio, tt.width)

			filledCount := strings.Count(result, "█")
			emptyCount := strings.Count(result, "░")

			if filledCount != tt.wantFilled {
				t.Errorf("채운 블록 수: got %d, want %d (bar: %s)", filledCount, tt.wantFilled, result)
			}
			if emptyCount != tt.wantEmpty {
				t.Errorf("빈 블록 수: got %d, want %d (bar: %s)", emptyCount, tt.wantEmpty, result)
			}
			// 총 길이 확인
			totalBlocks := filledCount + emptyCount
			if totalBlocks != tt.width {
				t.Errorf("총 블록 수: got %d, want %d", totalBlocks, tt.width)
			}
		})
	}
}

func TestRenderCriterionLine(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		criterion  CriterionStatus
		maxNameLen int
		contains   []string
		absent     []string
	}{
		{
			name:       "MUST 가중치 포함",
			criterion:  CriterionStatus{Name: "accuracy", PassRate: 0.90, Weight: "MUST"},
			maxNameLen: 10,
			contains:   []string{"MUST", "accuracy", "90%"},
		},
		{
			name:       "가중치 없음",
			criterion:  CriterionStatus{Name: "speed", PassRate: 0.70, Weight: ""},
			maxNameLen: 10,
			contains:   []string{"speed", "70%"},
			absent:     []string{"MUST"},
		},
		{
			name:       "100% 통과율",
			criterion:  CriterionStatus{Name: "quality", PassRate: 1.0, Weight: ""},
			maxNameLen: 10,
			contains:   []string{"quality", "100%"},
		},
		{
			name:       "0% 통과율",
			criterion:  CriterionStatus{Name: "coverage", PassRate: 0.0, Weight: "MUST"},
			maxNameLen: 10,
			contains:   []string{"coverage", "0%", "MUST"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := renderCriterionLine(tt.criterion, tt.maxNameLen)
			stripped := stripANSI(result)

			for _, want := range tt.contains {
				if !strings.Contains(stripped, want) {
					t.Errorf("출력에 %q가 포함되어야 함. 출력: %s", want, stripped)
				}
			}
			for _, absent := range tt.absent {
				if strings.Contains(stripped, absent) {
					t.Errorf("출력에 %q가 없어야 함. 출력: %s", absent, stripped)
				}
			}
		})
	}
}

func TestRenderDashboardNilData(t *testing.T) {
	t.Parallel()

	result := RenderDashboard(nil)
	if result != "" {
		t.Errorf("nil 데이터는 빈 문자열을 반환해야 함, got: %s", result)
	}
}

func TestRenderCompactNilData(t *testing.T) {
	t.Parallel()

	result := RenderCompact(nil)
	if result != "" {
		t.Errorf("nil 데이터는 빈 문자열을 반환해야 함, got: %s", result)
	}
}

func TestRenderProgressBarEdgeCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		ratio float64
		width int
	}{
		{"음수 비율은 0으로 클램프", -0.5, 25},
		{"1 초과 비율은 1로 클램프", 1.5, 25},
		{"너비 0", 0.5, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := renderProgressBar(tt.ratio, tt.width)

			filledCount := strings.Count(result, "█")
			emptyCount := strings.Count(result, "░")
			totalBlocks := filledCount + emptyCount

			if totalBlocks != tt.width {
				t.Errorf("총 블록 수: got %d, want %d", totalBlocks, tt.width)
			}
		})
	}
}
