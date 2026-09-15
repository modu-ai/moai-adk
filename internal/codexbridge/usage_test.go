package codexbridge

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/codexapp"
	"github.com/modu-ai/moai-adk/internal/codextools"
)

func usageEvent(turn string, input, cached, output int64) codexapp.Message {
	counts := map[string]int64{"inputTokens": input, "cachedInputTokens": cached, "cacheWriteInputTokens": 0, "outputTokens": output, "reasoningOutputTokens": 0, "totalTokens": input + output}
	raw, _ := json.Marshal(map[string]any{"threadId": "thread-1", "turnId": turn, "tokenUsage": map[string]any{"last": counts, "total": counts}})
	return codexapp.Message{Method: "thread/tokenUsage/updated", Params: raw}
}

func TestMeasuredUsageDoesNotDoubleCountAcrossToolSegments(t *testing.T) {
	e, rpc, _ := fixture(t)
	rpc.transform = func(m codexapp.Message) codexapp.Message {
		rpc.events <- usageEvent("unrelated-turn", 999, 0, 999)
		rpc.events <- usageEvent("turn-thread-1", 100, 60, 15)
		rpc.events <- usageEvent("turn-thread-1", 100, 60, 15)
		return m
	}
	rpc.respondEvents = []codexapp.Message{usageEvent("turn-thread-1", 100, 60, 15), usageEvent("turn-thread-1", 250, 140, 35)}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	q := request("usage")
	first, err := e.Step(ctx, q)
	if err != nil {
		t.Fatal(err)
	}
	if first.Usage == nil || first.Usage.InputTokens != 100 || first.Usage.CachedInputTokens != 60 || first.Usage.OutputTokens != 15 {
		t.Fatal(first.Usage)
	}
	q.Input = nil
	q.ExpectedPrefix = q.PrefixDigest
	q.PrefixDigest = "result"
	q.Results = []ToolResult{{ID: first.Tool.ID, Success: true}}
	last, err := e.Step(ctx, q)
	if err != nil {
		t.Fatal(err)
	}
	if last.Usage == nil || last.Usage.InputTokens != 150 || last.Usage.CachedInputTokens != 80 || last.Usage.OutputTokens != 20 {
		t.Fatal(last.Usage)
	}
}

func TestAbsentUsageRemainsUnknown(t *testing.T) {
	e, _, _ := fixture(t)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	seg, err := e.Step(ctx, request("unknown-usage"))
	if err != nil {
		t.Fatal(err)
	}
	if seg.Usage != nil {
		t.Fatal("fabricated usage", seg.Usage)
	}
}

func TestResumedUsageEstablishesBaselineWithoutBillingOldSnapshot(t *testing.T) {
	c := &conversation{owner: codextools.Binding{ThreadID: "thread-1"}, turn: "turn-thread-1"}
	var segment Segment
	c.observeUsage(json.RawMessage(`{"threadId":"thread-1","turnId":"turn-thread-1","tokenUsage":{"last":{"inputTokens":20,"cachedInputTokens":5,"outputTokens":3},"total":{"inputTokens":1000,"cachedInputTokens":500,"outputTokens":200}}}`), &segment)
	if segment.Usage != nil {
		t.Fatal("historical snapshot reattributed", segment.Usage)
	}
	c.observeUsage(usageEvent("turn-thread-1", 900, 400, 150).Params, &segment)
	if segment.Usage != nil {
		t.Fatal("older snapshot counted", segment.Usage)
	}
	c.observeUsage(usageEvent("turn-thread-1", 1020, 505, 203).Params, &segment)
	if segment.Usage == nil || segment.Usage.InputTokens != 20 || segment.Usage.OutputTokens != 3 {
		t.Fatal("measured difference missing", segment.Usage)
	}
}

func TestMalformedUsageDoesNotBecomeMeasuredZero(t *testing.T) {
	for _, raw := range []string{
		`{"threadId":"thread-1","turnId":"turn-thread-1","tokenUsage":{"last":{},"total":{}}}`,
		`{"threadId":"thread-1","turnId":"turn-thread-1","tokenUsage":{"last":{"inputTokens":10,"cachedInputTokens":20,"outputTokens":2},"total":{"inputTokens":10,"cachedInputTokens":20,"outputTokens":2}}}`,
		`{"threadId":"thread-1","turnId":"turn-thread-1","tokenUsage":{"last":{"inputTokens":null,"cachedInputTokens":0,"outputTokens":0},"total":{"inputTokens":0,"cachedInputTokens":0,"outputTokens":0}}}`,
	} {
		c := &conversation{owner: codextools.Binding{ThreadID: "thread-1"}, turn: "turn-thread-1"}
		var segment Segment
		c.observeUsage(json.RawMessage(raw), &segment)
		if segment.Usage != nil {
			t.Fatal("invalid usage counted", segment.Usage)
		}
	}
}
