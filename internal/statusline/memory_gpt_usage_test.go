package statusline

import "testing"

// The gauge faithfully exposes the gateway's supplied usage and window; it
// cannot infer native occupancy from multiple generations' summed input usage.
func TestCollectMemoryGPTObservedSegmentUsageDropIsNotCompaction(t *testing.T) {
	isolateModelEnv(t)
	t.Setenv("CLAUDE_AUTOCOMPACT_PCT_OVERRIDE", "85")
	for _, tt := range []struct{ input, window, want int }{
		{831857, 872000, 100}, {421548, 872000, 56}, {211050, 272000, 91},
	} {
		data := CollectMemory(&StdinData{ContextWindow: &ContextWindowInfo{ContextWindowSize: tt.window, CurrentUsage: &CurrentUsageInfo{InputTokens: tt.input}}})
		if got := usagePercent(data.TokensUsed, data.TokenBudget); got != tt.want {
			t.Fatalf("input=%d window=%d gauge=%d want=%d", tt.input, tt.window, got, tt.want)
		}
	}
}
