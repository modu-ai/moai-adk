package router_test

import (
	"testing"

	"github.com/modu-ai/moai-adk/internal/harness/router"
)

// TestEffortForLevel — AC-HRN-001-10, REQ-HRN-001-005.
// 3-row 테이블: minimal→medium, standard→high, thorough→xhigh.
func TestEffortForLevel(t *testing.T) {
	t.Parallel()

	cfg := minimalHarnessConfig()

	tests := []struct {
		level    router.Level
		wantEffort string
	}{
		{router.LevelMinimal, "medium"},
		{router.LevelStandard, "high"},
		{router.LevelThorough, "xhigh"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(string(tt.level), func(t *testing.T) {
			t.Parallel()
			got := router.EffortForLevel(tt.level, cfg)
			if got != tt.wantEffort {
				t.Errorf("EffortForLevel(%q): got %q, want %q", tt.level, got, tt.wantEffort)
			}
		})
	}
}

// TestEffortForLevel_Fallback — EffortMapping이 없으면 기본값 반환.
func TestEffortForLevel_Fallback(t *testing.T) {
	t.Parallel()

	// EffortMapping이 비어있는 config
	emptyCfg := &router.ConfigProxy{EffortMapping: map[string]string{}}

	tests := []struct {
		level    router.Level
		wantEffort string
	}{
		{router.LevelMinimal, "medium"},
		{router.LevelStandard, "high"},
		{router.LevelThorough, "xhigh"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(string(tt.level)+"_fallback", func(t *testing.T) {
			t.Parallel()
			got := router.EffortForLevelFromProxy(tt.level, emptyCfg)
			if got != tt.wantEffort {
				t.Errorf("EffortForLevel fallback(%q): got %q, want %q", tt.level, got, tt.wantEffort)
			}
		})
	}
}
