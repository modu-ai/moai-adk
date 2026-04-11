package dashboard

import (
	"regexp"
	"strings"
	"testing"
)

// ansiRe matches ANSI escape sequences.
var ansiRe = regexp.MustCompile(`\x1b\[[0-9;]*m`)

// stripANSI removes ANSI escape codes from a string.
func stripANSI(s string) string { return ansiRe.ReplaceAllString(s, "") }

func TestRenderDashboard(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		data     *DashboardData
		contains []string // strings that must appear in output
		absent   []string // strings that must not appear in output
	}{
		{
			name: "all-pass data contains 100%",
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
			name: "mixed data contains correct percentages",
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
			name: "zero experiments shows 0/20",
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
			name: "score improvement shows + prefix",
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
			name: "score regression shows - prefix",
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
					t.Errorf("output should contain %q but does not.\noutput:\n%s", want, stripped)
				}
			}
			for _, absent := range tt.absent {
				if strings.Contains(stripped, absent) {
					t.Errorf("output should not contain %q but does.\noutput:\n%s", absent, stripped)
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

	// Must be a single line (excluding trailing newline)
	trimmed := strings.TrimRight(stripped, "\n")
	if strings.Contains(trimmed, "\n") {
		t.Errorf("RenderCompact must be a single line but contains newline:\n%s", trimmed)
	}

	// Required elements
	for _, want := range []string{"Research:", "compact-test", "75%", "5/20", "3K", "2D"} {
		if !strings.Contains(stripped, want) {
			t.Errorf("compact output should contain %q. output: %s", want, stripped)
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
			name:       "100% filled",
			ratio:      1.0,
			width:      25,
			wantFilled: 25,
			wantEmpty:  0,
		},
		{
			name:       "0% empty",
			ratio:      0.0,
			width:      25,
			wantFilled: 0,
			wantEmpty:  25,
		},
		{
			name:       "50% half filled",
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
				t.Errorf("filled block count: got %d, want %d (bar: %s)", filledCount, tt.wantFilled, result)
			}
			if emptyCount != tt.wantEmpty {
				t.Errorf("empty block count: got %d, want %d (bar: %s)", emptyCount, tt.wantEmpty, result)
			}
			// Verify total length
			totalBlocks := filledCount + emptyCount
			if totalBlocks != tt.width {
				t.Errorf("total block count: got %d, want %d", totalBlocks, tt.width)
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
			name:       "MUST weight included",
			criterion:  CriterionStatus{Name: "accuracy", PassRate: 0.90, Weight: "MUST"},
			maxNameLen: 10,
			contains:   []string{"MUST", "accuracy", "90%"},
		},
		{
			name:       "no weight",
			criterion:  CriterionStatus{Name: "speed", PassRate: 0.70, Weight: ""},
			maxNameLen: 10,
			contains:   []string{"speed", "70%"},
			absent:     []string{"MUST"},
		},
		{
			name:       "100% pass rate",
			criterion:  CriterionStatus{Name: "quality", PassRate: 1.0, Weight: ""},
			maxNameLen: 10,
			contains:   []string{"quality", "100%"},
		},
		{
			name:       "0% pass rate",
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
					t.Errorf("output should contain %q. output: %s", want, stripped)
				}
			}
			for _, absent := range tt.absent {
				if strings.Contains(stripped, absent) {
					t.Errorf("output should not contain %q. output: %s", absent, stripped)
				}
			}
		})
	}
}

func TestRenderDashboardNilData(t *testing.T) {
	t.Parallel()

	result := RenderDashboard(nil)
	if result != "" {
		t.Errorf("nil data should return empty string, got: %s", result)
	}
}

func TestRenderCompactNilData(t *testing.T) {
	t.Parallel()

	result := RenderCompact(nil)
	if result != "" {
		t.Errorf("nil data should return empty string, got: %s", result)
	}
}

func TestRenderProgressBarEdgeCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		ratio float64
		width int
	}{
		{"negative ratio is clamped to 0", -0.5, 25},
		{"ratio above 1 is clamped to 1", 1.5, 25},
		{"width 0", 0.5, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := renderProgressBar(tt.ratio, tt.width)

			filledCount := strings.Count(result, "█")
			emptyCount := strings.Count(result, "░")
			totalBlocks := filledCount + emptyCount

			if totalBlocks != tt.width {
				t.Errorf("total block count: got %d, want %d", totalBlocks, tt.width)
			}
		})
	}
}
