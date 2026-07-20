---
title: Prompt Caching — The Concept and How Claude Code Uses It
weight: 30
draft: false
---

Prompt caching is an API feature that **reuses the front portion (the prefix)
of a request without reprocessing it when that portion is identical to the
previous request**. Tokens read from the cache are billed at **0.1x** the base
input rate, so the larger the repeated context (system prompt, project
instructions, conversation history), the greater the savings. If the "context
diet" axis of MoAI-ADK Tokenomics is about shrinking the always-loaded
context, prompt caching is about reusing the remaining context cheaply.

{{< callout type="info" >}}
**In plain terms** — on every turn the model rereads the whole conversation
from the start. Caching is the bookmark that says "the front portion is exactly
as I read it before" and skips ahead. If even a single character of the front
portion changes, the bookmark is invalidated and rereading resumes from that
point.
{{< /callout >}}

## Core Concepts (Common to the API)

- **Prefix matching**: a cache hit requires the content up to the breakpoint
  (including tool definitions, system prompt, and message history) to be
  **100% identical**. A single differing whitespace character forces
  everything after that point to be recomputed.
- **Price multipliers** (relative to the base input rate): cache write with a
  5-minute TTL **1.25x** · 1-hour TTL **2x** · cache read **0.1x**.
- **TTL (lifetime)**: 5 minutes by default, 1 hour optional. Every reuse within
  the TTL extends the lifetime for free.
- **Per-model minimum cacheable tokens**: a prefix shorter than this is not
  cached (processed normally, no error). For example, Fable 5 = 512,
  Opus 4.8 · Sonnet 5 = 1,024, Opus 4.7 = 2,048, Haiku 4.5 = 4,096 tokens.

## Claude Code Users — Caching Is Automatic

Claude Code **manages prompt caching automatically**. There is no need — and no
way — to set `cache_control` yourself. Per the official documentation, the
behavior is as follows:

- **Automatic TTL selection**: on subscription plans (Pro/Max/Team/Enterprise),
  Claude Code automatically requests a 1-hour TTL (included in the plan fee, so
  no extra cost). Via an API key or a cloud provider, the default is 5 minutes,
  and you can opt into 1 hour with `ENABLE_PROMPT_CACHING_1H=1`.
- **Environment-variable control**: `FORCE_PROMPT_CACHING_5M=1` (force
  5 minutes), `DISABLE_PROMPT_CACHING=1` (fully disable — not recommended
  outside debugging), and per-model variants such as
  `DISABLE_PROMPT_CACHING_OPUS`.
- **Request-layout optimization**: Claude Code arranges requests so that
  rarely-changing content (system prompt → project context → conversation)
  comes first, raising the prefix hit rate.

### Actions That Invalidate the Cache (one turn slower and costlier)

- Switching models (`/model`) · changing effort (`/effort`) — the cache is
  partitioned per model and per effort
- Connecting/disconnecting an MCP server (when tool definitions are loaded into
  the prefix)
- Adding/removing a full tool-deny rule
- `/compact` (conversation history is replaced by a summary) · the first turn
  after a Claude Code upgrade

### Actions That Preserve the Cache

- Editing repository files (added to the conversation only when read) · invoking
  a skill/command · switching permission mode · `/rewind` (returns to an
  already-cached prefix) · editing CLAUDE.md mid-session (the cache is preserved,
  but **the change is not applied either** — it takes effect on the next
  `/clear` or restart)

### Checking the Hit Rate

- The ratio of `cache_creation_input_tokens` (write) to
  `cache_read_input_tokens` (read) in the response is the metric — the higher
  the read share, the better it is working.
- The cache_hit segment of the MoAI statusline lets you check in real time
  during a session.
- If writes stay high every turn, that is a sign that one of the "invalidating
  actions" above is changing the prefix.

## Calling the API Directly (reference)

The usage example below applies **only to developers calling the Anthropic API
directly**. It does not apply to Claude Code users.

```python
# Direct API call: place the cache breakpoint on a stable system prompt
response = client.messages.create(
    model="claude-opus-4-8",
    max_tokens=1024,
    system=[{
        "type": "text",
        "text": "Stable system prompt...",
        "cache_control": {"type": "ephemeral", "ttl": "1h"}
    }],
    messages=[{"role": "user", "content": user_query}]
)
```

The principle is a single one: **place the breakpoint on the last stable block
before the data that changes on every request** (the query, a timestamp). The
break-even point is 2 requests — the first request's write premium is recovered
by the 0.1x read of the second request within the TTL.

## The Scope of MoAI cache.yaml

`.moai/config/sections/cache.yaml` (`enabled`, `session_ttl`) applies **only to
the `cache_control` injection when MoAI calls the Anthropic API directly through
its own SDK-wrapper path**. **It is unrelated to caching in Claude Code
sessions** — Claude Code's caching is managed automatically by the runtime as
described above, and MoAI cannot intervene in it.

> **GLM backend**: z.ai (GLM) uses content-similarity-based **implicit
> caching**, so MoAI does not inject `cache_control` on the GLM path.

## Summary

- **Claude Code users**: nothing to configure. Keep the hit rate high by doing
  model/effort switches and `/compact` only at natural boundaries between tasks.
- **Direct API callers**: breakpoint on a stable block; 1-hour TTL only when
  there are 2 or more requests.
- **Monitoring**: the statusline cache_hit segment + the
  `cache_read/creation` token ratio.

**Sources (official documentation):**

- [How Claude Code uses prompt caching](https://code.claude.com/docs/en/prompt-caching) — automatic management, automatic TTL selection, invalidating/preserving actions, environment variables
- [Prompt caching (API)](https://platform.claude.com/docs/en/build-with-claude/prompt-caching) — cache_control, price multipliers, per-model minimum tokens
- [Manage costs effectively](https://code.claude.com/docs/en/costs) — Claude Code's automatic cost optimization
