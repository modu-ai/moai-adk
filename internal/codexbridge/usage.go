package codexbridge

import (
	"encoding/json"
	"math"
)

// Usage contains provider-reported tokens. InputTokens includes cache reads
// and writes; OutputTokens already includes reasoning tokens.
type Usage struct {
	InputTokens           int64 `json:"inputTokens"`
	CachedInputTokens     int64 `json:"cachedInputTokens"`
	CacheWriteInputTokens int64 `json:"cacheWriteInputTokens"`
	OutputTokens          int64 `json:"outputTokens"`
}

func (u Usage) valid() bool {
	return u.InputTokens >= 0 && u.CachedInputTokens >= 0 && u.CacheWriteInputTokens >= 0 && u.OutputTokens >= 0 && u.CachedInputTokens <= u.InputTokens && u.CacheWriteInputTokens <= u.InputTokens-u.CachedInputTokens
}

// observeUsage is called with c.mu held. Totals are thread-cumulative, not
// HTTP-segment counts. A resumed thread's first report establishes a baseline
// only: it may describe an earlier generation. Fresh threads begin at zero.
// Repeated snapshots contribute nothing.
func (c *conversation) observeUsage(raw json.RawMessage, segment *Segment) {
	var event struct {
		ThreadID   string `json:"threadId"`
		TurnID     string `json:"turnId"`
		TokenUsage *struct {
			Last  json.RawMessage `json:"last"`
			Total json.RawMessage `json:"total"`
		} `json:"tokenUsage"`
	}
	if json.Unmarshal(raw, &event) != nil || event.ThreadID != c.owner.ThreadID || event.TurnID != c.turn || event.TokenUsage == nil {
		return
	}
	last, lastOK := decodeUsage(event.TokenUsage.Last)
	total, totalOK := decodeUsage(event.TokenUsage.Total)
	if !lastOK || !totalOK {
		return
	}
	var delta Usage
	if previous := c.usageTotal; previous != nil {
		if *previous == total {
			return
		}
		diff := Usage{InputTokens: total.InputTokens - previous.InputTokens, CachedInputTokens: total.CachedInputTokens - previous.CachedInputTokens, CacheWriteInputTokens: total.CacheWriteInputTokens - previous.CacheWriteInputTokens, OutputTokens: total.OutputTokens - previous.OutputTokens}
		if diff.valid() {
			delta = diff
		} else if c.usageTurn == c.turn {
			return
		} else {
			// A reset/compaction is a new baseline, not a charge for last.
			c.usageTotal = &total
			c.usageTurn = c.turn
			return
		}
	} else {
		c.usageTotal = &total
		c.usageTurn = c.turn
		return
	}
	if segment.Usage != nil && (delta.InputTokens > math.MaxInt64-segment.Usage.InputTokens || delta.CachedInputTokens > math.MaxInt64-segment.Usage.CachedInputTokens || delta.CacheWriteInputTokens > math.MaxInt64-segment.Usage.CacheWriteInputTokens || delta.OutputTokens > math.MaxInt64-segment.Usage.OutputTokens) {
		return
	}
	c.usageTotal = &total
	c.usageTurn = c.turn
	segment.ContextUsage = &last
	if segment.Usage == nil {
		segment.Usage = &Usage{}
	}
	segment.Usage.InputTokens += delta.InputTokens
	segment.Usage.CachedInputTokens += delta.CachedInputTokens
	segment.Usage.CacheWriteInputTokens += delta.CacheWriteInputTokens
	segment.Usage.OutputTokens += delta.OutputTokens
}

func decodeUsage(raw json.RawMessage) (Usage, bool) {
	var required map[string]json.RawMessage
	var usage Usage
	if json.Unmarshal(raw, &required) != nil || json.Unmarshal(raw, &usage) != nil {
		return usage, false
	}
	for _, name := range []string{"inputTokens", "cachedInputTokens", "outputTokens"} {
		value, ok := required[name]
		if !ok || string(value) == "null" {
			return usage, false
		}
	}
	return usage, usage.valid()
}
