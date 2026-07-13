# Cache-Aware Execution

Prompt-caching-aware ordering rules for orchestrator execution. Anthropic prompt caching is a **prefix match** over the rendered request (`tools` → `system` → `messages`): cache reads cost ~0.1× the base input price, writes cost 1.25× (5-minute TTL). The 5-minute TTL is **idle-based** — any gap longer than 5 minutes without a request (typically a blocking `AskUserQuestion` wait) expires the cache, and the next turn re-writes the full accumulated prefix at 1.25×. These rules govern WHEN and IN WHAT ORDER the orchestrator acts; they change no gate semantics and never bypass any approval gate.

> **Loading scope**: Intentionally always-loaded — the directives bind ordering decisions the orchestrator makes on any non-trivial turn (gate placement, agent spawns, rule edits, `/clear` timing).

## Directives

1. **Front-load user gates** [ZONE:Evolvable] While intent-drain gates can be asked early (Clarify stage, small context), ask them there rather than late in a large context. A blocking user wait late in a session risks expiring the cache over the entire accumulated prefix — the larger the context, the more expensive each gate-wait becomes. Unavoidable late gates (sync approval, completion decisions) SHOULD be batched into consecutive rounds so the expiry window is paid at most once, not per question.

2. **Stagger-spawn parallel same-type agents** [ZONE:Evolvable] When fanning out N parallel subagents that share the same agent definition (identical system prompt + rules prefix), spawn ONE first, and spawn the remaining N−1 after the first has started producing output. Concurrent requests cannot read a cache entry that is still being written — simultaneous fan-out makes all N pay the cold cache write for the shared prefix, while a staggered fan-out lets N−1 spawns read the first spawn's cache. This composes with (does not replace) the Mode 4 concurrency ceiling in `orchestration-mode-selection.md`.

3. **Defer session-loaded file edits to task end** [ZONE:Evolvable] Files loaded into the session prefix at start (`.claude/rules/`, `CLAUDE.md`, output styles, always-loaded skills) invalidate the entire cache prefix when edited mid-session — every subsequent turn re-writes from the edit point. Batch such edits at the END of a task, or immediately before a `/clear` boundary. This aligns naturally with the Template-First cycle (edit → `make build` → commit → session boundary).

4. **Consider `/clear` before large batches** [ZONE:Evolvable] `/clear` discards the warm cache but shrinks the prefix. When a large multi-spawn batch is about to start and the current context is bloated with completed unrelated work, a `/clear` + paste-ready resume BEFORE the batch is cheaper than carrying the bloated prefix through N spawns — even below the model-specific handoff threshold in `context-window-management.md`. When only short follow-up work remains, keep the warm cache instead.

5. **Inherit the session model on spawns** [ZONE:Evolvable] Caches are model-scoped: a per-spawn model override splits the spawn off from every cache the session has built. Omit model overrides unless the task genuinely requires a different tier (already the default per agent-authoring guidance); this directive records the caching cost of violating it.

## Non-goals

- These directives NEVER justify skipping, weakening, or reordering an approval gate's *semantics* — Implementation Kickoff Approval and all HUMAN GATEs remain mandatory where defined. Only the *placement and batching* of questions is governed here.
- Claude Code manages cache breakpoints internally; the orchestrator does not (and cannot) place `cache_control` markers. These rules optimize the variables the orchestrator does control: ordering, spawn timing, edit timing.

## Cross-references

- `.claude/rules/moai/workflow/orchestration-mode-selection.md` — Mode 4 parallel fan-out (stagger-spawn composes with its concurrency ceiling)
- `.claude/rules/moai/workflow/context-window-management.md` — model-specific `/clear` thresholds (directive 4 is an additional, earlier trigger)
- `.claude/rules/moai/core/agent-common-protocol.md` § Parallel Execution — single-turn verification batching (already cache-optimal: incremental append)
- `.claude/rules/moai/core/askuser-protocol.md` — gate mechanics (unchanged by this rule)

---

Version: 1.0.0
Classification: Evolvable operational rule — execution ordering only; gate semantics unchanged.
