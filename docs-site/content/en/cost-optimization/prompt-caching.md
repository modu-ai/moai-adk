---
title: Prompt Caching — Break-Even Analysis and Implementation Guide
weight: 30
draft: false
---

Prompt caching reuses an identical prompt prefix across multiple requests to
cut inference costs at a 90% discount (0.1x the base cost). If the "context
diet" axis of MoAI-ADK Tokenomics is about shrinking the always-loaded
context, prompt caching is about reusing the remaining context cheaply. This
guide explains the break-even rule, the cache mechanics, and when to enable
caching in a MoAI project.

## The break-even rule

**Only enable the 1-hour cache when a session makes 2 or more consecutive API requests.**

Single-request sessions (one-shot queries, single-turn commands) incur the 2x
write premium with no cache-read benefit, so the per-request cost ends up
higher than the uncached baseline. Conversely, with multi-turn interactions or
repeated SPEC analysis within an hour, caching offsets its cost on the second
request and saves 67% or more on subsequent ones.

### Cost comparison

Based on Claude Opus 4.5:

| Scenario | No caching | With 1-hour cache | Savings |
|---------|----------|-----------------|-------|
| 1 request, 10K tokens | $0.05 | $0.0625 | -25% (premium) |
| 2 requests, 10K + 10K | $0.10 | $0.0625 + $0.005 = $0.0675 | 32% saved |
| 3 requests, 10K + 10K + 10K | $0.15 | $0.0625 + 2×$0.005 = $0.0725 | 52% saved |
| 5 requests, 10K + 10K + 10K + 10K + 10K | $0.25 | $0.0625 + 4×$0.005 = $0.0825 | 67% saved |

The break-even point is **2 requests**: the first request's 2x write premium
is recovered by the 90%-discounted cache reads of subsequent requests inside
the 1-hour TTL window.

## How cache control works

When cache control is enabled on a prompt prefix, the cache lifecycle follows
this pattern:

1. **First request (cache write)**: the prefix is written to the cache after the API response completes. Cost: `prefix_tokens × 2.0 (1-hour cache) or 1.25 (5-minute cache)`.
2. **Subsequent requests (cache read)**: if the prefix is identical and within TTL, it is retrieved from the cache. Cost: `prefix_tokens × 0.1`.
3. **Automatic backtracking**: the system checks the last 20 messages for a cached prefix match. If found, the read cost applies.

### Cache placement best practice

Place the cache-control breakpoint at **the last stable block before
per-request data**:

```python
# Correct: stable system prompt (cacheable)
response = client.messages.create(
    model="claude-opus-4-8",
    max_tokens=1024,
    system=[
        {
            "type": "text",
            "text": "You are a code reviewer...",
            "cache_control": {"type": "ephemeral", "ttl": "1h"}
        }
    ],
    # Mutable per-request data below (not cached)
    messages=[{"role": "user", "content": user_query}]
)

# Wrong: cache breakpoint on mutable data
# Current time: {timestamp}
cache_control={"type": "ephemeral"}
# ^ The timestamp changes every request, so it never matches
```

## Configuration: session_ttl and spec_ttl

MoAI caching is configured in `.moai/config/sections/cache.yaml`:

```yaml
cache:
  enabled: false  # set to true to enable caching
  session_ttl: "1h"  # session-level cache TTL
  spec_ttl: "5m"     # SPEC body cache TTL
  min_cacheable_tokens: 2048  # minimum tokens for a cache write
```

### Opting out via session_ttl: "off"

To disable caching for a given session (e.g., when one-shot requests
dominate):

```yaml
cache:
  enabled: true
  session_ttl: "off"  # disable the cache for this session
  spec_ttl: "5m"      # only the SPEC body cache is used
```

With `session_ttl: "off"`:
- Session-level cache writes are skipped
- SPEC body caching still applies if `spec_ttl` is configured
- Useful for interrupt-driven workflows where single requests are the norm

## Monitoring cache performance

Use `moai doctor` to view the cache hit rate and decide whether to enable
caching:

```bash
moai doctor --cache-metrics
```

Sample output:

```
Cache performance (last 7 days):
  Cache hit rate: 67%
  Total cache reads: 450K tokens
  Total cache writes: 150K tokens
  Savings: $2.15 (68% cost reduction versus no caching)

Single-turn request ratio: 12%  ⚠️ Warning: 12% of requests are single-turn
                            (no cache benefit for these).
```

The MoAI statusline also shows a cache hit rate (cache_hit) segment, so you
can see the effect of the context diet and your caching configuration right
in the session.

### Interpreting the metrics

- **Hit rate > 60%**: the cache is effective. Keep it enabled.
- **Hit rate 30-60%**: moderate benefit. Consider enabling if your sessions are multi-turn heavy.
- **Single-turn ratio > 30%**: limited caching benefit. Verify the 2+ request assumption still holds.
- **Minimum-token threshold warnings**: configure `min_cacheable_tokens` to avoid caching small prompts (overhead > savings).

## When cache misses happen

A cache hit requires:

- ✓ An identical prompt prefix up to the breakpoint
- ✓ Within the TTL window (1 hour or 5 minutes)
- ✓ The same workspace/organization context
- ✓ No changes to any block before the breakpoint (tools, system, top-level parameters)

Cache misses (common causes):

- ✗ Tool definition change (different tool parameters)
- ✗ Web search toggled on/off
- ✗ Image added to or removed from the prefix
- ✗ Extended thinking settings changed
- ✗ Content before the breakpoint differs (including whitespace)

## Minimum token thresholds

A cache write is issued only when the prefix exceeds the per-model minimum:

- **Claude Opus 4.5, 4.7, 4.8, Haiku 4.5**: 2,048 tokens minimum
- **Claude Opus 4.1, Sonnet models, other Haiku versions**: 1,024 tokens minimum

Requests below the minimum are processed without caching (no error —
automatic fallback).

## Pre-warming (advanced)

To eliminate cache-miss latency on the first user interaction:

```python
# Pre-warm the cache (before the user arrives)
client.messages.create(
    model="claude-opus-4-8",
    max_tokens=0,  # no output token billing
    system="Long system prompt (5000 tokens)...",
    cache_control={"type": "ephemeral", "ttl": "1h"},
    messages=[{"role": "user", "content": "warmup"}]
)
# Cost: system_tokens × $2.0/MTok (cache write)

# Later: the user's request hits the warmed cache
client.messages.create(
    model="claude-opus-4-8",
    max_tokens=1024,
    system="Long system prompt (identical)...",
    cache_control={"type": "ephemeral", "ttl": "1h"},
    messages=[{"role": "user", "content": user_input}]
)
# Cost: system_tokens × $0.1/MTok (cache read from the warm-up)
```

## Summary

- **Enable caching**: sessions with 2+ consecutive API requests within an hour
- **Disable caching**: one-shot queries or interrupt-driven workflows
- **Monitor**: measure real hit rates with `moai doctor --cache-metrics` and the statusline cache_hit segment
- **Optimize**: place cache breakpoints on stable content (system prompts, instructions), never on mutable data (queries, timestamps)

For more details, see the [Anthropic prompt caching documentation](https://platform.claude.com/docs/en/docs/build-with-claude/prompt-caching).
