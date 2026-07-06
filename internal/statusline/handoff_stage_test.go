// Package statusline tests for SPEC-HANDOFF-THRESHOLD-001 (Handoff-v2 M4) D1:
// two-stage handoff-guide gate (handoffGuideStage enum + wrapper), the
// auto-compact-aware hard ceiling formula (+ clamp), and the M1 no-regression
// invariant (suffix is a pure function of usage, never gated by HandoffConfig).
package statusline

import (
	"strings"
	"testing"
)

// mem builds a MemoryData with a valid (renderable) budget for the given
// raw context window size + tokens used. TokenBudget is set >0 so
// renderBarsInline renders the CW bar; the stage is computed from the raw
// ContextWindowSize/TokensUsed independent of TokenBudget.
func mem(cwSize, tokensUsed int) MemoryData {
	return MemoryData{
		ContextWindowSize: cwSize,
		TokensUsed:        tokensUsed,
		TokenBudget:       cwSize * 85 / 100, // any positive budget; not used for stage
		Available:         true,
	}
}

// AC-THRESHOLD-006 static assertion: neither the stage classifier nor the
// suffix path takes a HandoffConfig — the statusline suffix is a pure function
// of *StatusData usage. A compile error here means the M1 no-regression
// invariant (suffix must not be config-gated) was violated.
var _ func(*StatusData) handoffStage = handoffGuideStage

// TestHandoffGuideStage_WrapperEquivalence — AC-THRESHOLD-001.
// shouldShowHandoffGuide MUST equal (handoffGuideStage != none) for every input,
// and nil / cwSize<=0 stay hidden.
func TestHandoffGuideStage_WrapperEquivalence(t *testing.T) {
	t.Parallel()

	cases := []*StatusData{
		nil,
		{Memory: MemoryData{ContextWindowSize: 0, TokensUsed: 100_000}},
		{Memory: mem(1_000_000, 500_000)}, // 1M @ 50% → soft
		{Memory: mem(1_000_000, 490_000)}, // 1M @ 49% → none
		{Memory: mem(256_000, 230_400)},   // 256K @ 90% → soft
		{Memory: mem(256_000, 227_840)},   // 256K @ 89% → none
		{Memory: mem(256_000, 245_760)},   // 256K @ 96% → hard
		{Memory: mem(200_000, 180_000)},   // 200K @ 90% → soft
		{Memory: mem(500_000, 250_000)},   // 500K @ 50% → soft (large band)
		{Memory: mem(499_999, 250_000)},   // just below cutoff @ ~50% → none
	}
	for i, data := range cases {
		want := handoffGuideStage(data) != handoffStageNone
		got := shouldShowHandoffGuide(data)
		if got != want {
			t.Errorf("case %d: shouldShowHandoffGuide=%v, (stage!=none)=%v", i, got, want)
		}
	}
}

// TestRenderBarsInline_TwoStageSuffix — AC-THRESHOLD-002.
// none → no suffix; soft → "(⚠️/clear)" (M1 verbatim); hard → "(🛑/clear!)".
func TestRenderBarsInline_TwoStageSuffix(t *testing.T) {
	t.Parallel()

	r := NewRenderer("default", true, nil)

	tests := []struct {
		name       string
		data       *StatusData
		wantSoft   bool
		wantHard   bool
		wantNoClrs bool
	}{
		{"none", &StatusData{Memory: mem(256_000, 100_000)}, false, false, true},   // 39% → none
		{"soft", &StatusData{Memory: mem(256_000, 230_400)}, true, false, false},   // 90% → soft
		{"hard", &StatusData{Memory: mem(256_000, 245_760)}, false, true, false},   // 96% → hard
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			out := r.renderBarsInline(tt.data, 10)
			hasSoft := strings.Contains(out, "(⚠️/clear)")
			hasHard := strings.Contains(out, "(🛑/clear!)")
			if hasSoft != tt.wantSoft {
				t.Errorf("soft marker=%v want %v in %q", hasSoft, tt.wantSoft, out)
			}
			if hasHard != tt.wantHard {
				t.Errorf("hard marker=%v want %v in %q", hasHard, tt.wantHard, out)
			}
			if tt.wantNoClrs && strings.Contains(out, "/clear") {
				t.Errorf("expected no /clear suffix, got %q", out)
			}
		})
	}
}

// TestHardCeiling_Formula — AC-THRESHOLD-003(a).
// Default auto-compact (85) → hard ceiling = min(95, 85+10) = 95, above soft.
// 256K band: rawPct≥95 → hard, 90≤rawPct<95 → soft, <90 → none.
func TestHardCeiling_Formula(t *testing.T) {
	t.Parallel()

	if got := hardCeilingPct(256_000); got != 95 {
		t.Fatalf("hardCeilingPct(256K) = %v, want 95", got)
	}
	if got := hardCeilingPct(1_000_000); got != 95 {
		t.Fatalf("hardCeilingPct(1M) = %v, want 95", got)
	}

	stageOf := func(tokens int) handoffStage {
		return handoffGuideStage(&StatusData{Memory: mem(256_000, tokens)})
	}
	if s := stageOf(245_760); s != handoffStageHard { // 96%
		t.Errorf("256K @ 96%% stage=%d, want hard", s)
	}
	if s := stageOf(243_200); s != handoffStageHard { // 95.0%
		t.Errorf("256K @ 95%% stage=%d, want hard", s)
	}
	if s := stageOf(240_640); s != handoffStageSoft { // 94.0%
		t.Errorf("256K @ 94%% stage=%d, want soft", s)
	}
	if s := stageOf(230_400); s != handoffStageSoft { // 90.0%
		t.Errorf("256K @ 90%% stage=%d, want soft", s)
	}
	if s := stageOf(227_840); s != handoffStageNone { // 89.0%
		t.Errorf("256K @ 89%% stage=%d, want none", s)
	}
}

// TestHardCeiling_ClampBelowSoft — AC-THRESHOLD-003(b).
// A degenerate auto-compact override (60) computes 60+10=70 < soft(90) for the
// standard band, so the ceiling clamps up to soft (90). Stage-2 then collapses
// onto the soft threshold: rawPct≥90 → hard, <90 → none (soft window absorbed).
// Non-parallel: t.Setenv is incompatible with t.Parallel.
func TestHardCeiling_ClampBelowSoft(t *testing.T) {
	t.Setenv("CLAUDE_AUTOCOMPACT_PCT_OVERRIDE", "60")

	if got := hardCeilingPct(256_000); got != 90 {
		t.Fatalf("clamped hardCeilingPct(256K) = %v, want 90 (== soft)", got)
	}

	stageOf := func(tokens int) handoffStage {
		return handoffGuideStage(&StatusData{Memory: mem(256_000, tokens)})
	}
	if s := stageOf(230_400); s != handoffStageHard { // 90.0% → hard (clamped)
		t.Errorf("clamped 256K @ 90%% stage=%d, want hard", s)
	}
	if s := stageOf(227_840); s != handoffStageNone { // 89.0% → none
		t.Errorf("clamped 256K @ 89%% stage=%d, want none", s)
	}
}

// TestNoM1Regression_SuffixUnconditional — AC-THRESHOLD-006 (INVARIANT).
// With the shipped defaults (HandoffConfig{Mode:"manual", Guide:false}) the soft
// suffix MUST still render — the statusline never sees HandoffConfig, so the
// suffix cannot be gated on Guide==true (which would make it vanish by default,
// an M1 regression). We prove it by rendering WITHOUT any config injection.
func TestNoM1Regression_SuffixUnconditional(t *testing.T) {
	t.Parallel()

	r := NewRenderer("default", true, nil)
	// 256K @ 90% raw → soft stage. No HandoffConfig is ever passed in.
	out := r.renderBarsInline(&StatusData{Memory: mem(256_000, 230_400)}, 10)
	if !strings.Contains(out, "(⚠️/clear)") {
		t.Fatalf("soft suffix vanished under default config — M1 regression: %q", out)
	}
}
