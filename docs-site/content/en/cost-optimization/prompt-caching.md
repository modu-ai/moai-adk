# Prompt Caching — Break-Even Analysis

Prompt caching reduces inference costs by reusing identical prompt prefixes across requests at 90% discount (0.1x base cost). This guide explains the break-even rule, cache mechanisms, and implementation best practices.

## Break-Even Rule

**Enable 1-hour cache only when your session generates 2 or more consecutive API requests.**

Single-request sessions incur a 2x write premium with no cache reuse benefit — cost per request is higher than uncached baseline. For multi-turn interactions within 1 hour, caching pays for itself on the second request and saves 67%+ on subsequent requests.

### Cost Comparison

Using Claude Opus 4.5 as reference:

| Scenario | No Cache | With 1h Cache | Savings |
|----------|----------|---------------|---------|
| 1 request, 10K tokens | $0.05 | $0.0625 | -25% (premium) |
| 2 requests, 10K + 10K | $0.10 | $0.0675 | 32% savings |
| 3 requests, 10K + 10K + 10K | $0.15 | $0.0725 | 52% savings |
| 5 requests, 5× 10K | $0.25 | $0.0825 | 67% savings |

Break-even is **2 requests**: the 2x write premium on the first request is recouped by 90%-discount cache reads on subsequent requests within the 1-hour TTL window.

## How Cache Control Works

When you enable cache control on a prompt prefix, the lifecycle follows this pattern:

1. **First Request (Cache Write)**: Prefix written to cache after API response. Cost: `prefix_tokens × 2.0 (1h) or 1.25 (5m)`.
2. **Subsequent Requests (Cache Read)**: Identical prefix within TTL retrieved from cache. Cost: `prefix_tokens × 0.1`.
3. **Automatic Lookback**: System checks last 20 messages for cached prefix match.

### Best Practice: Stable Prefix Placement

Place the cache control breakpoint at **the last stable block before per-request data**:

```python
# Correct: stable system prompt (cacheable)
response = client.messages.create(
    model="claude-opus-4-8",
    system=[{
        "type": "text",
        "text": "You are a code reviewer...",
        "cache_control": {"type": "ephemeral", "ttl": "1h"}
    }],
    # Changeable per-request data below (not cached)
    messages=[{"role": "user", "content": user_query}]
)

# Wrong: cache breakpoint on changing data
# Current time: {timestamp}
cache_control={"type": "ephemeral"}
# ^ Will NEVER match—timestamp changes every request
```

## Configuration: session_ttl and spec_ttl

MoAI caching is configured in `.moai/config/sections/cache.yaml`:

```yaml
cache:
  enabled: false
  session_ttl: "1h"  # Session-level cache TTL
  spec_ttl: "5m"     # SPEC body cache TTL
  min_cacheable_tokens: 2048
```

### Opt-out via session_ttl: "off"

To disable caching for a specific session:

```yaml
cache:
  session_ttl: "off"  # Disables session-level cache
  spec_ttl: "5m"      # SPEC body cache still applies
```

## Monitoring Cache Performance

Use `moai doctor` to view cache hit rate:

```bash
moai doctor --cache-metrics
```

### Interpreting Metrics

- **Hit rate > 60%**: Cache is effective. Keep enabled.
- **Hit rate 30-60%**: Moderate benefit. Consider enabling for multi-turn sessions.
- **Single-turn ratio > 30%**: Limited benefit. Verify the 2+ request assumption.
- **Min token threshold warning**: Configure `min_cacheable_tokens` to avoid caching tiny prompts.

## When Cache Misses Occur

Cache hits require:
- ✓ Identical prompt prefix up to breakpoint
- ✓ Within TTL window (1 hour or 5 minutes)
- ✓ Same workspace/organization context
- ✓ All blocks before breakpoint unchanged (tools, system)

Cache misses (common causes):
- ✗ Tool definitions changed
- ✗ Web search toggled
- ✗ Images added/removed
- ✗ Extended thinking settings changed
- ✗ Content before breakpoint differs

## Minimum Token Thresholds

Cache writes only issue when prefix exceeds model-specific minimum:

- **Claude Opus 4.5, 4.7, 4.8, Haiku 4.5**: 2,048 tokens minimum
- **Claude Opus 4.1, Sonnet, other Haiku**: 1,024 tokens minimum

Requests below minimum silently fall back to uncached processing.

## Pre-warming (Advanced)

Eliminate cache-miss latency for first user interaction:

```python
# Pre-warm cache (before users arrive)
client.messages.create(
    model="claude-opus-4-8",
    max_tokens=0,  # No output tokens billed
    system="Long system prompt (5000 tokens)...",
    cache_control={"type": "ephemeral", "ttl": "1h"},
    messages=[{"role": "user", "content": "warmup"}]
)
# Cost: system_tokens × $2.0/MTok (cache write)

# Later: User request hits warm cache
client.messages.create(
    model="claude-opus-4-8",
    max_tokens=1024,
    system="Long system prompt (identical)...",
    cache_control={"type": "ephemeral", "ttl": "1h"},
    messages=[{"role": "user", "content": user_input}]
)
# Cost: system_tokens × $0.1/MTok (cache read from warmup)
```

## Summary

- **Enable cache**: for sessions with 2+ consecutive API requests within 1 hour
- **Disable cache**: for one-shot queries or interrupt-driven workflows
- **Monitor**: use `moai doctor --cache-metrics` to measure actual hit rates
- **Optimize**: place cache breakpoints on stable content (system prompt, instructions), not changing data (queries, timestamps)

For more details, see [Anthropic prompt caching documentation](https://platform.claude.com/docs/en/docs/build-with-claude/prompt-caching).
