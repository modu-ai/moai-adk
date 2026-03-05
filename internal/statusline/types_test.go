package statusline

import "testing"

// TestNormalizeMode는 하위 호환성 모드 이름 정규화를 검증한다.
// REQ-V3-MODE-001: "minimal" → "compact" 변환
// REQ-V3-MODE-002: "verbose" → "full" 변환
func TestNormalizeMode(t *testing.T) {
	tests := []struct {
		name  string
		input StatuslineMode
		want  StatuslineMode
	}{
		// 하위 호환성: 이전 이름 → 새 이름 변환
		{"minimal은 compact로 변환", "minimal", ModeCompact},
		{"verbose는 full로 변환", "verbose", ModeFull},
		// 현재 이름은 변경 없음
		{"default는 변경 없음", "default", ModeDefault},
		{"compact는 변경 없음", "compact", ModeCompact},
		{"full은 변경 없음", "full", ModeFull},
		// 엣지 케이스
		{"빈 값은 변경 없음", "", ""},
		{"알 수 없는 값은 변경 없음", "custom", "custom"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NormalizeMode(tt.input)
			if got != tt.want {
				t.Errorf("NormalizeMode(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
