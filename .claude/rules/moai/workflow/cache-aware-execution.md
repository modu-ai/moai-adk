# Cache-Aware Execution

Prompt-caching-aware ordering rules for orchestrator execution. Caching is a **prefix match** over the rendered request: reads are far cheaper than writes, the TTL is **idle-based**, and a blocking wait longer than the window makes the next turn re-write the whole accumulated prefix. The directives below assume the 5-minute default — MoAI sets no TTL override, because cache spend is a user decision. Cited numbers, the TTL-setting caveat, and per-directive rationale: `cache-aware-execution-reference.md`. These rules govern WHEN and IN WHAT ORDER the orchestrator acts; they change no gate semantics and never bypass any approval gate.

> **Loading scope**: Intentionally always-loaded — the directives bind ordering decisions the orchestrator makes on any non-trivial turn (gate placement, agent spawns, rule edits, `/clear` timing).

## Directives

1. **Front-load user gates** [ZONE:Evolvable] Ask intent-drain gates early (Clarify stage, small context) rather than late in a large one: a blocking wait late in a session risks expiring the cache over the whole accumulated prefix. Unavoidable late gates (sync approval, completion decisions) SHOULD be batched into consecutive rounds so the expiry window is paid at most once, not per question.

2. **Stagger-spawn parallel same-type agents** [ZONE:Evolvable] When fanning out N parallel subagents sharing one agent definition, spawn ONE first and the remaining N−1 after it starts producing output: concurrent requests cannot read a cache entry still being written, so simultaneous fan-out makes all N pay the cold write. This composes with (does not replace) the fanout bounds in `orchestration-mode-selection.md` §C.2.

3. **Defer session-loaded file edits to task end** [ZONE:Evolvable] Files loaded into the session prefix at start (`.claude/rules/`, `CLAUDE.md`, output styles, always-loaded skills) invalidate the entire cache prefix when edited mid-session — every subsequent turn re-writes from the edit point. Batch such edits at the END of a task, or immediately before a `/clear` boundary. This aligns naturally with the Template-First cycle (edit → `make build` → commit → session boundary).

4. **Consider `/clear` before large batches** [ZONE:Evolvable] `/clear` discards the warm cache but shrinks the prefix. Apply `.claude/rules/moai/workflow/context-clear-policy.md`: persist a paste-ready handoff and clear before a large multi-spawn batch when unrelated context is bloated, an audit needs independence, or a prefix/model/effort change invalidates the warm state. When only short follow-up work remains and the plan/tree identity is unchanged, keep the warm cache instead.

5. **Inherit the session model on spawns** [ZONE:Evolvable] Caches are model-scoped: a per-spawn model override splits the spawn off from every cache the session has built. MoAI passes no model on a spawn and its agent definitions declare none, so every subagent shares the session's model; this directive records the caching cost of overriding it.

6. **Pass files by `@`-mention, not by name** [ZONE:Evolvable] [HARD] When a prompt needs a file's content, pass it with an `@`-mention or a Read call rather than citing the filename for the model to fetch — one deterministic load beats a fetch-retry cycle. Use `/context` only as a one-shot audit of what is loaded, not a routine check.

7. **Keep command output bounded** [ZONE:Evolvable] [HARD] Every command must bound what it returns: quiet flags, targeted queries, or redirect-to-file with the exit code and a bounded tail. `BASH_MAX_OUTPUT_LENGTH` is the runtime backstop, not the target; the file-redirect contract in `agent-common-protocol.md` § Parallel Execution applies to all commands.

8. **Prefer the quiet form of routine commands** [ZONE:Evolvable] [HARD] Call everyday commands in their quiet form — `--no-progress`, `-q`, machine-readable output with a targeted filter — not forms that emit spinners, banners, tables, or color noise: the same decision bytes, a fraction of the context cost.

9. **Weigh session length as a cost axis** [ZONE:Evolvable] [HARD] One long session is cheaper than several short ones for the same work — every fresh session re-pays the always-loaded prefix at write price, a continuing one reads it from cache — if it stays warm: a >5-min idle gap or prefix edit reverts it to write price. Treat session splitting as directive 4 treats `/clear`: a cost to justify, not a default.

10. **A mid-session model or effort switch busts the cache** [ZONE:Evolvable] [HARD] Changing model or effort mid-session (thinking budget included — `MAX_THINKING_TOKENS`) discards the prompt cache; prefer a natural boundary for the switch.

## Non-goals

- These directives NEVER justify skipping, weakening, or reordering an approval gate's *semantics* — Implementation Kickoff Approval and all HUMAN GATEs remain mandatory where defined. Only the *placement and batching* of questions is governed here.
- Cache breakpoints are managed internally by the runtime; the orchestrator can only order its own actions, so these rules govern ordering, spawn timing, and edit timing.

## Cross-references

- `.claude/rules/moai/workflow/orchestration-mode-selection.md` — fanout parallel fan-out (stagger-spawn composes with its concurrency ceiling)
- `.claude/rules/moai/workflow/context-window-management.md` — model-specific `/clear` thresholds (directive 4 is an additional, earlier trigger)
- `cache-aware-execution-reference.md` — the lazy companion. Load it for § Cited cache numbers · § Directive rationale (directives 6-10).

---

Version: 1.1.0
Classification: Evolvable operational rule — execution ordering only; gate semantics unchanged.
