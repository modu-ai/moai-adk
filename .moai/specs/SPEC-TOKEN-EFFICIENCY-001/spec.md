---
id: SPEC-TOKEN-EFFICIENCY-001
title: "Token-Efficiency P0 bundle — always-loaded budget guard + cache-hit-ratio statusline"
version: "0.2.0"
status: in-progress
created: 2026-07-02
updated: 2026-07-02
author: manager-spec
priority: P0
phase: "v3.0.0"
module: "internal/config, internal/statusline"
lifecycle: spec-anchored
tags: "token-efficiency, prompt-cache, statusline, budget-guard, observability, tier-m"
issue_number: null
tier: M
---

# SPEC-TOKEN-EFFICIENCY-001 — Token-Efficiency P0 bundle

## §A Background, Goal, Scope

### §A.1 Background

moai-adk is a **harness ON TOP OF Claude Code (CC)**. CC owns native graduated
compaction (`Budget Reduction → Snip → Microcompact → Context Collapse →
Auto-Compact`) and leader-session prompt caching; both are CC-internal and MUST
NOT be reimplemented by the harness (over-engineering guard, cf.
`.claude/rules/moai/workflow/runtime-recovery-doctrine.md` AP-RR-001).

The token lever moai-adk **does** control is the **project-context layer it
injects into every CC turn**: `CLAUDE.md` + the always-loaded
`.claude/rules/moai/**` rule files (those WITHOUT a `paths:` frontmatter
restriction) + the output-style `.claude/output-styles/moai/moai.md` + the
loaded head of `MEMORY.md`. CC re-reads that whole layer at the cache-READ rate
(0.1× base input) on **every** turn, and degrades recall via n²-attention
context-rot as it grows. So **minimizing (size)** and **stabilizing
(byte-stability)** that layer is the controllable target.

Evidence:
- arXiv:2601.06007 "Don't Break the Cache" (https://arxiv.org/abs/2601.06007)
  measured a **41–80% cost cut** from keeping stable system context lean and
  cache-friendly.
- Anthropic "Effective context engineering for AI agents"
  (https://www.anthropic.com/engineering/effective-context-engineering-for-ai-agents)
  — n² attention means every always-loaded token degrades recall; prefer JIT
  loading.
- Claude Code "How Claude Code uses prompt caching"
  (https://code.claude.com/docs/en/prompt-caching) — 0.1× cache-read rate;
  model + effort participate in the cache key; 3-layer prefix.
- CC Engineering "Prompt caching is everything"
  (https://claude.com/blog/lessons-from-building-claude-code-prompt-caching-is-everything)
  — CC's own team SEV-alerts on cache-hit rate; static-first / dynamic-last
  prefix discipline.

This bundle is deliberately **low-risk**: it adds guards + one doc correction +
one observability signal. It introduces **no new runtime subsystem**. It
mechanizes and protects the diet already achieved by the Epic Steering-Align
work (CLAUDE.md 650→409 lines) against silent regression.

### §A.2 Goal

Ship two independently-shippable P0 improvements to the harness's own
token-efficiency posture:

- **P0-1** — a CI/`go test` guard that measures the always-loaded context
  surface's token count and FAILS when it exceeds a configured budget (prevents
  silent regression of the achieved diet).
- **P0-2** — surface a `cache_read : cache_creation` ratio in the moai
  statusline as an early-warning signal of prompt-prefix churn (the same signal
  CC's own team alerts on), with graceful degradation when the underlying data
  is absent.

> **Scope carve-out (v0.2.0).** This SPEC originally bundled three P0 items;
> the middle item — the output-style §4 "~7×" attribution reword — was carved
> into a separate follow-up SPEC, **SPEC-DIVECC-ATTRIBUTION-FIX-001**, after a
> plan-audit FAIL (two BLOCKING defects, both confined to that item: an
> unattributed "VERIFIED" self-claim and a grep gate structurally blind to the
> `.moai/` subtree). This revised SPEC ships the two surviving items,
> renumbered contiguously: **P0-1** (budget guard) and **P0-2** (cache-hit-ratio
> statusline — the item formerly labeled P0-3). The two survivors are pure-Go
> and doc-independent; the carved-out reword is doc-only, so nothing in this
> SPEC depends on it.

### §A.3 Scope decisions and verification verdicts

One precondition the delegating brief flagged was resolved during plan-phase
investigation (recorded here so the run-phase does not re-derive it):

1. **P0-2 statusline data availability — VERIFIED (was flagged UNVERIFIED).**
   The cache fields ARE exposed by the CC statusline stdin JSON at
   `context_window.current_usage.cache_read_input_tokens` and
   `context_window.current_usage.cache_creation_input_tokens` (official schema:
   https://code.claude.com/docs/en/statusline). The moai Go statusline already
   models them — `internal/statusline/types.go` `CurrentUsageInfo` fields
   `CacheReadTokens` / `CacheCreationTokens` — and already consumes them
   (`internal/statusline/memory.go` sums them into `tokensUsed`). P0-2 is
   therefore a **new derivation on already-parsed data**, not a new ingestion.
   The official docs also state `current_usage` is `null` before the first API
   call AND again after `/compact` until the next call repopulates it — this is
   the exact trigger for the required graceful-degradation path.

The former second precondition (the "~7×" source-wording reword) is carved out
to SPEC-DIVECC-ATTRIBUTION-FIX-001 and is no longer in this SPEC's scope.

### §A.4 Tier M rationale

- Scope: 2 independent items across 2 Go domains (Go config guard, Go statusline
  derivation). Estimated 6–8 files, ~250–500 LOC.
- Per `.claude/rules/moai/workflow/spec-workflow.md` § SPEC Complexity Tier:
  Tier M = 300–1000 LOC, 5–15 files, 3-file artifact set (spec + plan +
  acceptance), plan-auditor PASS threshold 0.80.
- Not Tier L: P0-2's data-availability uncertainty (the only genuine unknown) is
  resolved in §A.3; there is no deep-architecture decision requiring research.md
  / design.md. The one open decision (P0-1 token-estimation method + budget
  value) is a plan.md decision, not a research question.
- Not Tier S: 2 new Go behaviors (guard enumeration + statusline derivation)
  across 2 packages + their tests sit at the upper Tier S boundary; retained as
  Tier M (threshold 0.80) per the delegating decision so the re-audit applies
  the same bar as the original bundle. The carved-out doc-only item is removed,
  but the two surviving Go behaviors keep the SPEC above the Tier S single-file
  envelope.

### §A.5 Out of Scope

This section enumerates what this SPEC deliberately does NOT build.

### Out of Scope — CC-native mechanisms (over-engineering guard)

- Reimplementing CC's native graduated compaction (`Budget Reduction → Snip →
  Microcompact → Context Collapse → Auto-Compact`) — CC owns it; the harness
  sits on top and cannot modify the native loop.
- Reimplementing or wrapping CC leader-session prompt caching — CC owns the
  cache; moai only observes its effect (P0-2) and keeps the injected prefix lean
  (P0-1).
- Dynamic-TTL cache auto-tuning, context-editing, or prompt-prefix rewriting for
  the CC leader path — CC already handles cache lifetime and prefix assembly.

### Out of Scope — P0-1 budget guard boundaries

- Enforcing the budget on conditionally-loaded (`paths:`-scoped) rule files —
  only the always-loaded surface (no `paths:` restriction) is in scope; scoped
  rules load JIT and do not inflate every turn.
- Automatically trimming or rewriting over-budget files — the guard FAILS and
  reports; it does NOT auto-edit content (scope discipline).
- Counting agent/skill bodies, `.claude/rules` files that carry a `paths:`
  restriction, or on-demand MEMORY.md topic files — these are not part of the
  always-loaded per-turn surface.

### Out of Scope — the "~7×" attribution reword (carved out)

- The entire "~7×" attribution correction (output-style §4 reword, template
  mirror parity, archive reference correction) is carved out to
  **SPEC-DIVECC-ATTRIBUTION-FIX-001** and is NOT built here. This SPEC does not
  touch `.claude/output-styles/moai/moai.md`, its template mirror, or
  `.moai/research/dive-into-claude-code-archive.md`.
- Note: the P0-1 budget guard still *reads* `.claude/output-styles/moai/moai.md`
  as part of the always-loaded surface it measures — but it never *modifies* it.
  The read is a measurement input, not an edit target.

### Out of Scope — P0-2 statusline boundaries

- Modifying the `.moai/status_line.sh.tmpl` wrapper — it only forwards stdin to
  `moai statusline`; the derivation lives in the Go package, so the wrapper
  needs no change.
- Ingesting new stdin fields — the cache fields are already parsed
  (`CurrentUsageInfo`); no schema change is required.
- Historical cache-ratio trend, logging, or alerting — P0-2 surfaces a
  point-in-time ratio in the statusline only; no persistence layer.

## §B Requirements (GEARS Format)

### P0-1 — always-loaded token-budget guard

#### REQ-TEF-001 (Ubiquitous, mandatory)

The token-budget guard **shall** measure the aggregate estimated token count of
the always-loaded context surface and **shall** fail (non-zero test result) when
that aggregate exceeds a configured budget constant.

#### REQ-TEF-002 (Event-driven, mandatory)

**When** the guard runs under `go test`, the system **shall** compute the
always-loaded surface as the union of: `CLAUDE.md`; every
`.claude/rules/moai/**/*.md` file whose frontmatter carries NO `paths:`
restriction; the output-style `.claude/output-styles/moai/moai.md`; and the
loaded head of `MEMORY.md` (first 200 lines OR 25KB, whichever is reached
first, matching the CC auto-memory loader cap).

#### REQ-TEF-003 (Ubiquitous, mandatory)

The system **shall** derive the token estimate via a single documented
deterministic method exposed as a named constant/function, and **shall** express
the budget as a named constant (or config key). The exact estimation method and
budget value are a plan.md decision (see plan.md §D) — the requirement is that
both are named, deterministic, and documented, not hidden magic numbers.

#### REQ-TEF-004 (Unwanted-behavior / state, mandatory)

**While** the measured surface is at or under budget, the guard **shall** pass
without output. The guard **shall not** fail on, or count, any
`.claude/rules/moai/**` file that carries a `paths:` frontmatter restriction
(conditionally-loaded rules are out of scope per §A.5).

### P0-2 — cache-hit-ratio statusline signal

> (Formerly P0-3 in the 3-item bundle; renumbered P0-2 after the "~7×" reword
> was carved out. REQ IDs are renumbered contiguously 005→007.)

#### REQ-TEF-005 (Event-driven, mandatory)

**When** the statusline receives stdin JSON whose
`context_window.current_usage` is non-null with a positive
`cache_creation_input_tokens`, the system **shall** compute and surface a
cache-read-vs-cache-creation signal (a ratio or a cache-hit percentage derived
from `cache_read_input_tokens` and `cache_creation_input_tokens`).

#### REQ-TEF-006 (Unwanted-behavior / graceful degradation, mandatory)

**While** `context_window.current_usage` is null (before the first API call, or
after `/compact` until repopulation) OR `cache_creation_input_tokens` is zero,
the system **shall** degrade gracefully — omit the cache-ratio signal (or show
the existing context-% only) — and **shall not** fabricate a value or divide by
zero.

#### REQ-TEF-007 (Capability gate / Where, mandatory)

**Where** the cache-ratio segment is enabled through the existing statusline
segment configuration, the system **shall** render it consistently with the
existing segment toggle conventions (`internal/statusline` segment config), so
users who do not want the signal can disable it without code changes.

## §C Acceptance Criteria

See [acceptance.md](./acceptance.md) for the canonical AC enumeration
(AC-TEF-001 … AC-TEF-007) with independently verifiable commands and expected
outcomes. acceptance.md is the SSOT for AC bodies; the §B REQ rationale
references AC IDs but does not duplicate AC text. REQ↔AC bijection: 7 REQ
(REQ-TEF-001…007) ↔ 7 AC (AC-TEF-001…007).

## §D Constraints

- **D-1**: No reimplementation of CC-native compaction or caching (§A.5).
- **D-2**: P0-1 and P0-2 are independently shippable — a blocker on one MUST
  NOT block the other (they are separate Go packages: `internal/config` and
  `internal/statusline`).
- **D-3**: P0-1 guard MUST be deterministic and hermetic (reads repo files under
  the test working tree; no network, no clock, no env dependence beyond the repo
  path).
- **D-4**: P0-2 MUST NOT change the stdin schema or the `.moai/status_line.sh.tmpl`
  wrapper; the derivation lives in `internal/statusline` Go code.
- **D-5**: Every research/benchmark claim cited in this SPEC is attributed to a
  verified source (§A.1); P0-2 data availability is verified against the official
  statusline docs + the existing Go struct (§A.3), satisfying
  `.claude/rules/moai/core/verification-claim-integrity.md`.
- **D-6**: Tier M — plan-auditor PASS threshold 0.80; Section A-E delegation
  template applies at run-phase.

## §E Risks

- **Risk-E1 (Medium) — P0-1 budget value calibration.** A budget set too tight
  fails on the current always-loaded surface (measured at plan-time); too loose
  never fires. Mitigation: plan.md §D sets the budget from the measured current
  baseline plus a headroom margin, recorded as a named constant with a comment
  explaining the derivation. The AC (AC-TEF-002) asserts PASS-on-current-tree
  relatively (no hardcoded file count), so the guard is not pinned to a snapshot.
- **Risk-E2 (Low) — P0-1 always-loaded enumeration drift.** The set of
  no-`paths:` rule files changes as rules are added/scoped. Mitigation: the
  guard enumerates dynamically (globs `.claude/rules/moai/**` and inspects
  frontmatter for `paths:`), so it self-updates rather than hardcoding the list.
- **Risk-E3 (Low) — P0-2 ratio semantics ambiguity.** "cache_read : cache_creation"
  can read as a raw ratio or a hit-percentage. Mitigation: plan.md §D fixes the
  exact display form and the zero/null degradation before implementation.
