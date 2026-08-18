# Cache-Aware Execution

Prompt-caching-aware ordering rules for orchestrator execution. Anthropic prompt caching is a **prefix match** over the rendered request (`tools` → `system` → `messages`): cache reads cost ~0.1× the base input price, writes cost 1.25× (5-minute TTL). The 5-minute TTL is **idle-based** — any gap longer than 5 minutes without a request (typically a blocking `AskUserQuestion` wait) expires the cache, and the next turn re-writes the full accumulated prefix at 1.25×. These rules govern WHEN and IN WHAT ORDER the orchestrator acts; they change no gate semantics and never bypass any approval gate.

> **Loading scope**: Intentionally always-loaded — the directives bind ordering decisions the orchestrator makes on any non-trivial turn (gate placement, agent spawns, rule edits, `/clear` timing).

## Directives

1. **Front-load user gates** [ZONE:Evolvable] While intent-drain gates can be asked early (Clarify stage, small context), ask them there rather than late in a large context. A blocking user wait late in a session risks expiring the cache over the entire accumulated prefix — the larger the context, the more expensive each gate-wait becomes. Unavoidable late gates (sync approval, completion decisions) SHOULD be batched into consecutive rounds so the expiry window is paid at most once, not per question.

2. **Stagger-spawn parallel same-type agents** [ZONE:Evolvable] When fanning out N parallel subagents that share the same agent definition (identical system prompt + rules prefix), spawn ONE first, and spawn the remaining N−1 after the first has started producing output. Concurrent requests cannot read a cache entry that is still being written — simultaneous fan-out makes all N pay the cold cache write for the shared prefix, while a staggered fan-out lets N−1 spawns read the first spawn's cache. This composes with (does not replace) the fanout bounds in `orchestration-mode-selection.md` §C.2 (the 3-5 advisory band this directive grounds; the hard bound is the runtime subagent cap).

3. **Defer session-loaded file edits to task end** [ZONE:Evolvable] Files loaded into the session prefix at start (`.claude/rules/`, `CLAUDE.md`, output styles, always-loaded skills) invalidate the entire cache prefix when edited mid-session — every subsequent turn re-writes from the edit point. Batch such edits at the END of a task, or immediately before a `/clear` boundary. This aligns naturally with the Template-First cycle (edit → `make build` → commit → session boundary).

4. **Consider `/clear` before large batches** [ZONE:Evolvable] `/clear` discards the warm cache but shrinks the prefix. When a large multi-spawn batch is about to start and the current context is bloated with completed unrelated work, a `/clear` + paste-ready resume BEFORE the batch is cheaper than carrying the bloated prefix through N spawns — even below the model-specific handoff threshold in `context-window-management.md`. When only short follow-up work remains, keep the warm cache instead.

5. **Inherit the session model on spawns** [ZONE:Evolvable] Caches are model-scoped: a per-spawn model override splits the spawn off from every cache the session has built. Omit model overrides unless the task genuinely requires a different tier (already the default per agent-authoring guidance); this directive records the caching cost of violating it.

6. **Pass files by `@`-mention, not by name** [ZONE:Evolvable] [HARD] When a prompt needs a file's content, pass it with an `@`-mention or a Read call rather than citing the filename for the model to fetch — one deterministic load beats a fetch-retry cycle. Use `/context` only as a one-shot audit of what is loaded, not a routine check.

7. **Keep command output bounded** [ZONE:Evolvable] [HARD] Every command must bound what it returns: quiet flags, targeted queries, or redirect-to-file with the exit code and a bounded tail. `BASH_MAX_OUTPUT_LENGTH` is the runtime backstop, not the target; the file-redirect contract in `agent-common-protocol.md` § Parallel Execution applies to all commands.

8. **Prefer the quiet form of routine commands** [ZONE:Evolvable] [HARD] Call everyday commands in their quiet form — `--no-progress`, `-q`, machine-readable output with a targeted filter — not forms that emit spinners, banners, tables, or color noise: the same decision bytes, a fraction of the context cost.

9. **Weigh session length as a cost axis** [ZONE:Evolvable] [HARD] One long session is cheaper than several short ones for the same work — every fresh session re-pays the always-loaded prefix at write price, a continuing one reads it from cache — if it stays warm: a >5-min idle gap or prefix edit reverts it to write price. Treat session splitting as directive 4 treats `/clear`: a cost to justify, not a default.

10. **A mid-session model or effort switch busts the cache** [ZONE:Evolvable] [HARD] Changing model or effort mid-session (thinking budget included — `MAX_THINKING_TOKENS`) discards the prompt cache; prefer a natural boundary for the switch. Directive 5 and this one govern the main session's cache; `agent-common-protocol.md` § Per-Spawn Model Injection governs which model a subagent runs on — different axes, not a contradiction.

## Non-goals

- These directives NEVER justify skipping, weakening, or reordering an approval gate's *semantics* — Implementation Kickoff Approval and all HUMAN GATEs remain mandatory where defined. Only the *placement and batching* of questions is governed here.
- Claude Code manages cache breakpoints internally; the orchestrator does not (and cannot) place `cache_control` markers. These rules optimize the variables the orchestrator does control: ordering, spawn timing, edit timing.

## Cross-references

- `.claude/rules/moai/workflow/orchestration-mode-selection.md` — fanout parallel fan-out (stagger-spawn composes with its concurrency ceiling)
- `.claude/rules/moai/workflow/context-window-management.md` — model-specific `/clear` thresholds (directive 4 is an additional, earlier trigger)
- `.claude/rules/moai/core/agent-common-protocol.md` § Parallel Execution — single-turn verification batching (already cache-optimal: incremental append)
- `.claude/rules/moai/core/askuser-protocol.md` — gate mechanics (unchanged by this rule)
- `.claude/rules/moai/workflow/cache-aware-execution-reference.md` — cited cache numbers for directives 6-10

---

Version: 1.1.0
Classification: Evolvable operational rule — execution ordering only; gate semantics unchanged.
