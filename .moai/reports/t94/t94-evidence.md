# t94 evidence — concurrency-cap rationale rebuild (orchestration-mode-selection.md)

Card t94, worktree t85, branch WT-t85. Base HEAD: 59dd62d02 (card t85's implementation commit — untouched).
Task: rebuild the 3-5 "ceiling" rationale in `orchestration-mode-selection.md` so the three conflated
limits are distinguished, and reconcile the factory N=8 default landed by t85 on this same tree.

> Filename note: the card prescribed `.moai/reports/t94/report.md`; the session's write guard blocks that
> exact filename for subagents, so the same content is persisted as `t94-evidence.md` (this file) in the
> same directory. Sibling artifact: `t85-run-evidence.md`.

## 1. The three limits — verification quotes (what was actually read)

| # | Limit | Binds | Value | Source actually read (file:line, pre-edit tree @ 59dd62d02) |
|---|-------|-------|-------|--------------------------------------------------------------|
| 1 | Subagent fan-out — HARD bound | concurrent subagents per turn, every `Agent()` spawn surface | `CLAUDE_CODE_MAX_CONCURRENT_SUBAGENTS`, default 20 | `CLAUDE.md:145` — "Runtime fan-out caps (distinct from nesting depth, §4): `CLAUDE_CODE_MAX_CONCURRENT_SUBAGENTS` (default 20) per-turn; per-session total cap removed v2.1.224" and `.claude/rules/moai/core/moai-constitution.md:31` — "runtime cap is 20" |
| 2 | Workflow agent concurrency | agents within ONE dynamic workflow (Mode 6 only) | 16 concurrent per workflow + 1,000-total backstop | `.claude/rules/moai/workflow/dynamic-workflows.md:91` — "Up to **16 concurrent agents** (fewer on machines with limited CPU cores); **1,000 agents total per run** as a runaway-loop backstop." The `min(16, available CPUs − 2)` derivation is RUNTIME-DOCUMENTED (Workflow tool description), not repo-measured — labeled as such in the rewritten §C.2 |
| 3 | Team size — ADVISORY | named teammates in ONE Agent Team (Mode 3, experimental) | 3-5 | `.claude/rules/moai/core/moai-constitution.md:36` — "Agent Teams (shared TaskList, start with 3-5 teammates)"; `dynamic-workflows.md:24` — "3-5 teammates (Anthropic recommendation)" |

The conflation (pre-edit): §C.2 quoted the limit-3 team-size recommendation ("Start with 3-5 teammates",
Anthropic Agent Teams guidance) as the grounding for the Mode 4 SUBAGENT fan-out cap — a number that never
bound subagent fan-out. t85's factory default N=8 then "violated" a ceiling that did not exist.

Verified non-hit: `agent-common-protocol.md` § Background Agent Execution does NOT carry a 3-5 statement
(grep for `3-5|3–5|ceiling` — only retry-ceiling/token-ceiling hits). The mission's suspicion of that file
was incorrect; no edit was needed there.

## 2. Edits — before / after

All edits applied identically to the template mirror (`internal/template/templates/...`), 8 byte-identical
pairs + root `CLAUDE.md` ↔ `internal/template/templates/CLAUDE.md`.

### 2.1 Primary: `.claude/rules/moai/workflow/orchestration-mode-selection.md` (frontmatter 1.1.0→1.2.0, footer 1.3.1→1.4.0)

- **§A Mode 4 row** — before: "3-5 concurrent sub-agents (single message, multiple `Agent()` calls)"
  → after: "3-5 concurrent sub-agents ADVISORY (hard bound: runtime subagent cap, default 20 — §C.2); single message, multiple `Agent()` calls"
- **§B tree** — before: "Mode 4: PARALLEL (3-5 concurrent Agent() in single message)"
  → after: "Mode 4: PARALLEL (3-5-advisory concurrent Agent() in single message; hard bound is the runtime subagent cap — §C.2)"
- **§C.2 rebuilt** — before: compound clause said "maximum 3-5 concurrent `Agent()` calls in a single message
  per Anthropic verbatim 'Start with 3-5 teammates'" + "The 3-5 ceiling applies to Mode 4 ... contradicts
  Anthropic's published guidance"
  → after: heading "§C.2 Mode 4 (Parallel) Compound preference — three concurrency limits (SSOT)"; compound
  clause says "within the fan-out bounds below"; new three-limit table (rows 1/2/3 above); closing paragraph
  grounds Mode 4's 3-5 as an advisory band in cache/prefix economics (stagger-spawn, cache-aware-execution
  directive 2) + coordination overhead, names the numeric coincidence with limit 3 as coincidence, states
  limit 1 as the hard bound, and keeps write fan-out separately bounded (never two write-capable agents).
- **§C.4 added (new)** — "Factory workers (default 8) are not Mode 4 fan-out": `moai glm -f` fleet is
  tmux/session workers under the workers-registry / free-slot discipline (registry prune, live-claim probe,
  staggered activation — as implemented by t85 in `internal/kanban/factory_slots.go`), not `Agent()` calls
  in one orchestrator turn; where a lead DOES invoke `glm_task`/`Agent()` fan-out, limit 1 + stagger-spawn
  govern exactly as for Mode 4.
- **§E anti-pattern replaced** — before: "Spawning > 5 concurrent agents in Mode 4 — exceeds
  Anthropic-recommended 3-5 ceiling"
  → after: "Quoting the team-size advisory as a Mode 4 fan-out cap, or treating Mode 4's 3-5 band as a hard
  cap" (with the stagger-spawn forfeiture clause).
- **§F citations** — "Mode 3 ceiling" → "Mode 3 team-size advisory"; the "Start with 3-5 teammates" quote
  now annotated "(team-size advisory binding Mode 3 only — never a subagent fan-out cap; see §C.2)".
- **t92 frame preserved**: §C.1 (experimental re-allow, genealogy, sentinel history), Mode 3 row, §G/§G.1/§G.2
  untouched; the 3-5 in the Mode 3 row was already correctly scoped as team size and stays.

### 2.2 Pointing texts (minimal updates)

- `moai-constitution.md:31` — before: "MoAI ceiling: 3-5 concurrent Agent() calls (Mode 4); runtime cap is
  20." → after: "Mode 4 fan-out bounds: hard bound is the runtime subagent cap (`CLAUDE_CODE_MAX_CONCURRENT_SUBAGENTS`, default 20 per turn); MoAI's 3-5 is a cache/coordination ADVISORY for fan-out, distinct
  from the 3-5 TEAM-SIZE advisory that binds Mode 3 teammates only." — §C.2 SSOT anchor retained.
- `CLAUDE.md:145` (root + template) — "MoAI's own ceiling is 3-5 (Mode 4)" → "MoAI's own Mode 4 band is 3-5
  as an advisory (cache/coordination economics — the runtime cap is the hard bound, and the team-size 3-5
  advisory binds Mode 3 teammates only)."
- `orchestrator-templates.md:15` — table row now "(advisory — hard bound is the runtime subagent cap, default 20)".
- `orchestrator-templates.md:92` — "**Ceiling**: 3-5 ..." → "**Advisory band**: 3-5 ... not a hard cap and
  not the Agent-Teams team-size number; the hard bound is the runtime subagent cap (default 20)."
- `spec-workflow.md:448` — "3-5 concurrent read-only `Agent()` in one turn" → "+ advisory band; hard bound is
  the runtime subagent cap, per orchestration-mode-selection.md §C.2".
- `session-handoff.md:88` — "stays within the 3-5 concurrent `Agent()` ceiling" → "stays within the Mode 4
  fan-out bounds (the 3-5 advisory band under the runtime subagent cap — orchestration-mode-selection.md §C.2)".
- `session-handoff-examples.md:285` — table row "3-5 concurrent" → "3-5-advisory concurrent".
- `session-handoff-examples.md:296(b)` — the explicit conflation "respects the 3-5 concurrent `Agent()` ceiling
  (§C.2, applied equally to Mode 3 and Mode 4)" → "respects the Mode 4 fan-out bounds (§C.2: the 3-5 advisory
  band under the runtime subagent cap as hard bound; the 3-5 team-size advisory binds Mode 3 teammates only)".
- `cache-aware-execution.md:11` — "composes with ... the Mode 4 concurrency ceiling" → "composes with ... the
  Mode 4 fan-out bounds in `orchestration-mode-selection.md` §C.2 (the 3-5 advisory band this directive
  grounds; the hard bound is the runtime subagent cap)".

Template neutrality: added text contains NO SPEC IDs, no card IDs (t85/t92/t94), no dates, no commit SHAs;
factory + teams frame described by feature name only. Pre-existing `SPEC-XXX`/`SPEC-MYPROJ-001` placeholders
in touched template files are unchanged antecedent content.

## 3. Scope boundary (deliberate)

8 skill-file sites under `.claude/skills/moai/workflows/` (`review.md:71`, `plan/spec-assembly.md:174`,
`run/phase-execution.md:412`, `run/task-decomposition.md:74,260`, `sync/doc-execution.md:126`,
`sync/quality-gates-quality.md:209,307`) say "3-5 concurrent per the Mode 4 ceiling (§C.2)". Left unchanged:
they defer to §C.2 as SSOT (the reference still resolves) and do not conflate team-size with fan-out; the
mission's verification grep scope is `.claude/rules/` only. Residual: the word "ceiling" there now resolves
to the §C.2 advisory band — cosmetic rename candidate for a follow-up card.

## 4. Verification outputs (this run, tree @ working state post-edit)

- `diff` of all 8 local↔template pairs + root↔template CLAUDE.md → `ALL 8 PAIRS IDENTICAL`
- `make build` → exit 0 ("catalog.yaml updated successfully (12407 bytes)"; catalog.yaml NOT modified in git —
  it does not hash rule files: `grep -c "orchestration-mode-selection|moai-constitution" internal/template/catalog.yaml` → 0)
- `grep -rn "3-5" .claude/rules/ internal/template/templates/.claude/rules/` → every remaining hit is either
  a team-size advisory (moai-constitution:36, dynamic-workflows:24/83, §A Mode 3 row, §F quote, §C.2 table
  row 3) or an explicitly-marked Mode 4 advisory with hard-bound pointer; NONE presents 3-5 as a subagent
  fan-out hard cap (irrelevant antecedent hits: askuser-protocol-reference:144 "3-5 edit", skill-ab-testing:100 "3-5 samples")
- `grep -n "§C.2" moai-constitution.md` → :31 "See orchestration-mode-selection.md §C.2 (SSOT)" resolves to
  `orchestration-mode-selection.md:127` "### §C.2 Mode 4 (Parallel) Compound preference — three concurrency
  limits (SSOT)"; `:169` "### §C.4 Factory workers (default 8) are not Mode 4 fan-out"
- `go test ./internal/template/ -run 'TestRuleTemplateMirrorDrift|TestSanitizedPairParity|TestTemplateNoInternalContentLeak' -count=1`
  → `ok github.com/modu-ai/moai-adk/internal/template 2.566s`
- No *_test.go file references "orchestration-mode-selection" (grep → empty)

## 5. Gaps

- `min(16, available CPUs − 2)` is runtime-documented (Workflow tool description), not repo-measured — no
  Workflow tool is exposed in this session to observe directly; labeled as runtime-documented in §C.2.
- The 8 skill-file "Mode 4 ceiling" phrasings (§3 above) left as-is.
- Pre-existing real SPEC IDs in template `spec-workflow.md:331,347` (SPEC-AUDIT-SNAPSHOT-001) are antecedent
  content outside this card's scope; flagged for a neutrality follow-up if the CI guard ever tightens.

## 6. Residual risks

- CI `template-neutrality-check.yaml` triggers on the template path change; added lines are clean, but the
  guard scans whole files containing antecedent placeholder SPEC tokens — prior commits touching the same
  files passed, so expected green; CI is the final judge.
- Semantic risk: relaxing "3-5 ceiling" to "advisory under hard bound 20" could be read as licensing larger
  write fan-out — closed by the retained "never two write-capable agents concurrently" clause in §C.2 and
  the write-fan-out stays-foreground doctrine in orchestrator-templates.md.

Diff: 16 files, 62 insertions(+), 38 deletions(-).
