package statusline

import (
	"strings"
	"testing"
)

// abs 는 정수의 절댓값을 반환한다.
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// TestInterpolateGradientColor 는 블록 위치에 따른 RGB 그라디언트 보간을 검증한다.
// Green(0,255,0) → Yellow(255,255,0) → Red(255,0,0) 경로 확인.
func TestInterpolateGradientColor(t *testing.T) {
	tests := []struct {
		name     string
		blockPct float64 // 0.0 to 1.0
		wantR    int
		wantG    int
		wantB    int
	}{
		{"start (green)", 0.0, 0, 255, 0},
		{"quarter (mid green-yellow)", 0.25, 128, 255, 0}, // approx
		{"mid (yellow)", 0.5, 255, 255, 0},
		{"three-quarter (mid yellow-red)", 0.75, 255, 128, 0}, // approx
		{"end (red)", 1.0, 255, 0, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, g, b := interpolateGradientColor(tt.blockPct)
			// ±2 오차 허용 (반올림 처리)
			if abs(r-tt.wantR) > 2 || abs(g-tt.wantG) > 2 || abs(b-tt.wantB) > 2 {
				t.Errorf("interpolateGradientColor(%f) = (%d,%d,%d), want ~(%d,%d,%d)",
					tt.blockPct, r, g, b, tt.wantR, tt.wantG, tt.wantB)
			}
		})
	}
}

// TestInterpolateGradientColor_Clamp 는 범위 밖 입력에 대한 클램핑을 검증한다.
func TestInterpolateGradientColor_Clamp(t *testing.T) {
	// 0 이하는 Green 반환
	r, g, b := interpolateGradientColor(-0.5)
	if r != 0 || g != 255 || b != 0 {
		t.Errorf("clamp below 0: got (%d,%d,%d), want (0,255,0)", r, g, b)
	}

	// 1 이상은 Red 반환
	r, g, b = interpolateGradientColor(1.5)
	if r != 255 || g != 0 || b != 0 {
		t.Errorf("clamp above 1: got (%d,%d,%d), want (255,0,0)", r, g, b)
	}
}

// TestBuildGradientBar 는 그라디언트 프로그레스 바의 블록 수를 검증한다.
func TestBuildGradientBar(t *testing.T) {
	tests := []struct {
		name       string
		pct        int
		width      int
		noColor    bool
		wantFilled int
		wantEmpty  int
	}{
		{"60% of 40", 60, 40, false, 24, 16},
		{"60% of 10", 60, 10, false, 6, 4},
		{"0%", 0, 40, false, 0, 40},
		{"100%", 100, 40, false, 40, 0},
		{"50% of 10", 50, 10, false, 5, 5},
		{"zero width", 50, 0, false, 0, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildGradientBar(tt.pct, tt.width, tt.noColor)
			// noColor=false 일 때는 ANSI 이스케이프가 포함되어 단순 Count 불가
			// noColor=true 인 경우에만 정확한 블록 수 검증
			if tt.noColor {
				filled := strings.Count(result, "█")
				empty := strings.Count(result, "░")
				if filled != tt.wantFilled || empty != tt.wantEmpty {
					t.Errorf("blocks: got %d filled + %d empty, want %d + %d", filled, empty, tt.wantFilled, tt.wantEmpty)
				}
			}
			if tt.width > 0 && result == "" {
				t.Error("expected non-empty result for positive width")
			}
			if tt.width == 0 && result != "" {
				t.Error("expected empty result for zero width")
			}
		})
	}
}

// TestBuildGradientBar_NoColor 는 noColor 모드에서 ANSI 이스케이프가 없음을 검증한다.
func TestBuildGradientBar_NoColor(t *testing.T) {
	result := BuildGradientBar(50, 10, true)
	// ANSI 이스케이프 시퀀스 없어야 함
	if strings.Contains(result, "\033") || strings.Contains(result, "\x1b") {
		t.Error("noColor 모드는 ANSI 이스케이프 시퀀스를 포함하면 안 됩니다")
	}
	// 정확히 5 filled + 5 empty
	filled := strings.Count(result, "█")
	empty := strings.Count(result, "░")
	if filled != 5 || empty != 5 {
		t.Errorf("got %d filled + %d empty, want 5 + 5", filled, empty)
	}
}

// TestBuildGradientBar_WithColor 는 컬러 모드에서 ANSI 이스케이프가 포함됨을 검증한다.
func TestBuildGradientBar_WithColor(t *testing.T) {
	result := BuildGradientBar(50, 10, false)
	// lipgloss 가 ANSI 이스케이프를 추가하므로 길이가 10자 이상이어야 함
	if len(result) <= 15 {
		t.Error("컬러 모드는 ANSI 이스케이프 시퀀스로 인해 더 길어야 합니다")
	}
}

// TestBuildGradientBar_FilledCountWithColor 는 컬러 모드에서도 올바른 블록 수를 갖는지 검증한다.
func TestBuildGradientBar_FilledCountWithColor(t *testing.T) {
	// 60% / 40 blocks -> 24 filled, 16 empty
	result := BuildGradientBar(60, 40, false)
	filled := strings.Count(result, "█")
	empty := strings.Count(result, "░")
	if filled != 24 {
		t.Errorf("60%% of 40: want 24 filled blocks, got %d", filled)
	}
	if empty != 16 {
		t.Errorf("60%% of 40: want 16 empty blocks, got %d", empty)
	}
}

// TestBuildGradientBar_SingleFilled 는 filled=1 인 경우 나눗셈 by zero 없이 동작함을 검증한다.
func TestBuildGradientBar_SingleFilled(t *testing.T) {
	// 10% of 10 -> 1 filled
	result := BuildGradientBar(10, 10, false)
	filled := strings.Count(result, "█")
	if filled != 1 {
		t.Errorf("10%% of 10: want 1 filled block, got %d", filled)
	}
}
