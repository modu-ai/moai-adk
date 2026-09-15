package codexbridge

import (
	"encoding/json"
	"github.com/modu-ai/moai-adk/internal/codextools"
	"testing"
)

func TestUsageContextTracksLastGenerationWithoutLosingSegmentBilling(t *testing.T) {
	c := &conversation{owner: codextools.Binding{ThreadID: "thread"}, turn: "turn", usageTotal: &Usage{}}
	var segment Segment
	total := int64(0)
	for _, input := range []int64{206169, 208381, 208577, 208730} {
		total += input
		raw, _ := json.Marshal(map[string]any{"threadId": "thread", "turnId": "turn", "tokenUsage": map[string]any{"last": map[string]int64{"inputTokens": input, "cachedInputTokens": 0, "outputTokens": 10}, "total": map[string]int64{"inputTokens": total, "cachedInputTokens": 0, "outputTokens": total / 1000}}})
		c.observeUsage(raw, &segment)
	}
	if segment.Usage == nil || segment.Usage.InputTokens != 831857 {
		t.Fatalf("billing lost: %+v", segment.Usage)
	}
	if segment.ContextUsage == nil || segment.ContextUsage.InputTokens != 208730 {
		t.Fatalf("context inflated by generation count: %+v", segment.ContextUsage)
	}
	before := *segment.ContextUsage
	c.observeUsage(usageEvent("foreign", 999, 0, 1).Params, &segment)
	if *segment.ContextUsage != before {
		t.Fatal("foreign usage changed context")
	}
}
