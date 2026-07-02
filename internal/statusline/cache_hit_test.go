package statusline

import (
	"strings"
	"testing"
)

// TestCacheHitPercent covers the cache-hit-ratio derivation incl. graceful
// degradation (AC-TEF-005 compute, AC-TEF-006 zero/degrade, overflow edge).
//
// # REQ-TEF-005, REQ-TEF-006
func TestCacheHitPercent(t *testing.T) {
	tests := []struct {
		name        string
		cacheRead   int
		cacheCreate int
		wantPct     int
		wantOK      bool
	}{
		{name: "read2000 create5000 → 28%", cacheRead: 2000, cacheCreate: 5000, wantPct: 28, wantOK: true},
		{name: "read equals create → 50%", cacheRead: 5000, cacheCreate: 5000, wantPct: 50, wantOK: true},
		{name: "all read, some create → high", cacheRead: 9000, cacheCreate: 1000, wantPct: 90, wantOK: true},
		{name: "zero creation → degrade (no ratio)", cacheRead: 2000, cacheCreate: 0, wantPct: 0, wantOK: false},
		{name: "both zero → degrade (no 0/0)", cacheRead: 0, cacheCreate: 0, wantPct: 0, wantOK: false},
		{name: "negative creation → degrade", cacheRead: 100, cacheCreate: -5, wantPct: 0, wantOK: false},
		{name: "negative read canceling creation → denom==0 divide-by-zero defense", cacheRead: -5, cacheCreate: 5, wantPct: 0, wantOK: false},
		{name: "very large counts → no overflow", cacheRead: 1_500_000_000, cacheCreate: 500_000_000, wantPct: 75, wantOK: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pct, ok := cacheHitPercent(tt.cacheRead, tt.cacheCreate)
			if ok != tt.wantOK {
				t.Errorf("cacheHitPercent(%d,%d) ok = %v, want %v", tt.cacheRead, tt.cacheCreate, ok, tt.wantOK)
			}
			if pct != tt.wantPct {
				t.Errorf("cacheHitPercent(%d,%d) pct = %d, want %d", tt.cacheRead, tt.cacheCreate, pct, tt.wantPct)
			}
		})
	}
}

// TestRenderCacheHit covers the segment rendering incl. graceful degradation:
// nil usage (null current_usage) and zero cache-creation both omit the segment
// with no fabricated value and no panic (AC-TEF-005, AC-TEF-006).
//
// # REQ-TEF-005, REQ-TEF-006
func TestRenderCacheHit(t *testing.T) {
	tests := []struct {
		name  string
		usage *CurrentUsageInfo
		want  string
	}{
		{name: "present with cache → renders %", usage: &CurrentUsageInfo{CacheReadTokens: 2000, CacheCreationTokens: 5000}, want: "💾 28%"},
		{name: "null current_usage → omit", usage: nil, want: ""},
		{name: "zero cache_creation → omit", usage: &CurrentUsageInfo{CacheReadTokens: 2000, CacheCreationTokens: 0}, want: ""},
		{name: "both zero → omit (no 0/0)", usage: &CurrentUsageInfo{CacheReadTokens: 0, CacheCreationTokens: 0}, want: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := &StatusData{CacheUsage: tt.usage}
			if got := renderCacheHit(data); got != tt.want {
				t.Errorf("renderCacheHit() = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestCacheHitSegmentToggle verifies the cache segment is config-toggleable
// consistent with existing segment conventions (parallel to SegmentEffortThinking):
// disabled via config → absent; default/enabled → present (AC-TEF-007).
//
// # REQ-TEF-007
func TestCacheHitSegmentToggle(t *testing.T) {
	data := &StatusData{
		Metrics:   MetricsData{Available: true, Model: "Opus"},
		CacheUsage: &CurrentUsageInfo{CacheReadTokens: 2000, CacheCreationTokens: 5000},
	}

	// Default (nil segmentConfig → all enabled): segment present.
	enabledLine := NewRenderer("", true, nil).renderInfoLine(data, false)
	if !strings.Contains(enabledLine, "💾") {
		t.Errorf("default renderInfoLine = %q, want cache-hit segment present", enabledLine)
	}

	// Explicitly disabled: segment absent.
	disabledLine := NewRenderer("", true, map[string]bool{SegmentCacheHit: false}).renderInfoLine(data, false)
	if strings.Contains(disabledLine, "💾") {
		t.Errorf("disabled renderInfoLine = %q, want cache-hit segment absent", disabledLine)
	}

	// Explicitly enabled: segment present.
	explicitLine := NewRenderer("", true, map[string]bool{SegmentCacheHit: true}).renderInfoLine(data, false)
	if !strings.Contains(explicitLine, "💾") {
		t.Errorf("enabled renderInfoLine = %q, want cache-hit segment present", explicitLine)
	}
}
