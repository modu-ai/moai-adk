---
id: SPEC-TCD-LLM-DECIDER-001
title: "LLM-based card-classification decider at the todo add seam — a GLM-backed Decider(llm) implementation fulfilling the deferred half of REQ-TCD-012"
version: "0.1.0"
status: draft
created: 2026-10-03
updated: 2026-10-03
author: manager-spec (card t1352)
priority: P3
phase: "v3.2.0 target"
module: "internal/cli, internal/config"
lifecycle: spec-anchored
tags: "todo, classification, decider, llm, glm, fail-safe, seam, default-off, card-t1352"
tier: M
card: t1352
depends_on: [SPEC-TODO-CLASSIFY-DISPATCH-001]
related_specs: [SPEC-AUTONOMY-CONTRACT-001, SPEC-FACTORY-SELF-DISPATCH-001]
---

# SPEC: LLM-based card-classification decider (Decider llm)

## HISTORY

| Version | Date | Change |
|---------|------|--------|
| 0.1.0 | 2026-10-03 | Initial plan-phase authoring (card t1352, operator directive 2026-10-03), measured in worktree `.claude/worktrees/t1352`, branch `WT-llm-card-decider`, at HEAD `681ee30b5` (== local develop tip). Delivers the implementation the completed SPEC-TODO-CLASSIFY-DISPATCH-001 explicitly deferred (its amendment 0.3.1, leader ruling 2026-09-29 OD-2): the card text names REQ-TCD-012's "product provides Decider(llm)" clause. All card premises verified against the tree — the seam (`todoCardDecider`, `todo_classify.go:31`), the identity `llm` already in the accepted set (`classification.go:64-68`), and the existing GLM transport precedent (`mcp_glm.go:3-6`) — none falsified. |

> **Provenance discipline.** Every `file:line` citation was measured at HEAD `681ee30b5` in this
> worktree. The parent SPEC's artifacts are read-only for this SPEC — referenced, never edited.

## §A Context

### A.1 The deferred half

SPEC-TODO-CLASSIFY-DISPATCH-001 shipped the classification seam: the package var
`todoCardDecider` (`internal/cli/todo_classify.go:31`) resolves every admitted card's judgment,
the `--classification-file` flag swaps in a static carrier for one invocation, and
`DefaultCardDecider` (identity `default`) is the shipped judgment backend. Its leader ruling
2026-09-29 (OD-2, amendment 0.3.1) deferred the `Decider(llm)` implementation to a follow-up
card. The identity `llm` is already a first-class constant in the accepted decider set
(`internal/kanban/classification.go:43,64-68`) — it has vocabulary but no shipped
implementation. Card t1352 closes that gap: an LLM-based `CardDecider` implementation added at
the seam behind `DefaultCardDecider`.

The card is framed Low priority: "자동 배차 볼륨이 커져 display-only 분류로 부족해질 때 상향".
The SPEC respects that framing — a minimal, correct, **default-off** surface. No speculative
configuration, no new config-file key, no provider abstraction.

### A.2 What this SPEC adds — exactly two pieces

1. **An LLM decider implementation** implementing `kanban.CardDecider`
   (`internal/kanban/classification.go:245`), recording decider identity `llm`, whose judgment
   comes from one in-process HTTPS call to the project's configured GLM endpoint using the
   existing GLM credential reader — the same transport pattern `internal/cli/mcp_glm.go:3-6`
   already ships for the `glm_audit` MCP tool (z.ai Anthropic-compatible endpoint,
   `config.DefaultGLMBaseURL`, `~/.moai/.env.glm` credential).
2. **A default-off activation surface**: an environment variable selects the LLM decider as the
   standing judgment backend for the add paths. Unset, the add path is byte-identical to today.

The six-field classification contract (`CardClassification`,
`classification.go:85-92`: priority, blocked, mode, decider identity, classified-at, reason) is
untouched — the LLM judgment lands in the same shape through the same validator.

## §B Requirements (GEARS)

> Value sets reuse the parent SPEC's closed sets: `priority ∈ {high, normal, low}`;
> `blocked ∈ {true, false}`; `mode ∈ {serial, parallelizable}`; the recorded decider identity of
> this implementation is exactly `llm`. The fail-safe default is unchanged
> (`DefaultCardClassification`, `classification.go:98-105`).

- **REQ-TLD-001 (Ubiquitous)** — The product shall ship an LLM-based card-classification decider
  implementing the `kanban.CardDecider` seam, recording the decider identity `llm` on every
  judgment it supplies, so the `llm` identity carries a shipped implementation (fulfilling the
  deferred half of the parent SPEC's REQ-TCD-012); the judgment source is the project's
  configured GLM endpoint with the existing GLM credential reader (`internal/glmcred`), and no
  product path shall route the judgment through `scripts/jev/`, TypeSafe/Jev tooling, or any
  local-only tooling path.

- **REQ-TLD-002 (Where)** — Where the environment variable `MOAI_TODO_DECIDER` carries `llm`,
  every add surface that resolves the standing decider (the CLI `todo add` path and the MCP
  `todo_add` tool, which share `runTodoAddAppendRoot`, `todo.go:787`) shall select the LLM
  decider; where the variable is unset, empty, or `default`, the standing decider shall remain
  `DefaultCardDecider` unchanged; where it carries any other value, the add shall be refused as
  a usage error (exit 2) before anything is written, naming the accepted set, with `jev` and
  `llm+jev` refused by the parent SPEC's named refusal wording; and where `--classification-file`
  is supplied together with the `llm` selection, the supplied judgment file wins and the LLM
  decider shall not be invoked.

- **REQ-TLD-003 (When)** — When the LLM decider fails to return a judgment — absent
  credentials, connection failure, non-2xx response, timeout, non-JSON body, or a response
  outside the closed value sets — the add path shall admit the card with the fail-safe default
  (priority `normal`, blocked `false`, mode `serial`, decider identity `default`), print exactly
  one stderr fallback notice (the existing `todoClassificationFallbackNotice` line,
  `todo_classify.go:36`), and never block admission; a card add shall not gain a hard dependency
  on network or LLM availability.

- **REQ-TLD-004 (Ubiquitous)** — An LLM response shall be accepted only when its priority,
  blocked, and mode fields are inside the closed value sets, validated through the single
  existing validator (`kanban.ValidateCardClassification`, `classification.go:109-123` — no
  second spelling of the value sets); the recorded decider identity shall be exactly `llm`
  regardless of any decider identity the response itself claims, and the add path shall not
  record a model-claimed identity (`human`, `jev`, or otherwise) in place of `llm`.

- **REQ-TLD-005 (Ubiquitous)** — The LLM judgment shall be computed BEFORE the queue lock is
  acquired and attached to the card INSIDE the same locked write (`Mutate`) that appends the
  card, so no machine reader ever observes an admitted card without a recorded classification —
  the parent SPEC's visibility invariant (REQ-TCD-001) is preserved; only the computation's
  placement moves out of the lock, never the attachment.

- **REQ-TLD-006 (Ubiquitous)** — The LLM call shall be bounded by an HTTP timeout whose default
  lives in `internal/config/defaults.go`, shall authenticate through the existing GLM
  credential reader, shall resolve its endpoint and model from config constants, and shall name
  the activation environment variable through its `internal/config/envkeys.go` constant — no
  inline `os.Getenv` literals, no inline endpoint literals, and no secret shall be embedded in
  any log line, error message, or stderr output.

- **REQ-TLD-007 (Ubiquitous)** — With `MOAI_TODO_DECIDER` unset, the add path's behavior shall
  be identical to the pre-SPEC behavior: the deterministic default judgment, no HTTP request
  attempted, no additional stderr output, and no change to the recorded classification of any
  card — the dormant-default framing the card directs.

## §C Constraints

- C1. **Jev boundary (persists, C4 of the parent SPEC).** `scripts/jev/` is never referenced by
  product code; `internal/jevcred`/TypeSafe credentials are not consumed by this decider; the
  `internal/jev` package keeps its `workflow.jev.enabled` default-off gate and gains no new
  consumer; `jev` and `llm+jev` remain refused identities
  (`classification.go:117-119`). The GLM endpoint is the judgment source — Jev is never named
  as one.
- C2. **Invariant amendment, recorded.** The parent SPEC's header invariant "the product
  computes no LLM call and knows no external tooling" (`todo_classify.go:12-15`) is amended by
  this SPEC for exactly one gated surface: the LLM decider of REQ-TLD-001. Rationale: the
  card's mission IS a shipped LLM decider; the invariant's protective intent — no local-only
  tooling dependency, no Jev routing, default-off, fail-safe — is preserved in full, and
  `--classification-file` remains the sole operator-supplied judgment injection path (REQ-TCD-004
  of the parent SPEC, untouched).
- C3. **Single validator, single spelling.** Closed value sets are validated only through
  `kanban.ValidateCardClassification`; the new code introduces no re-spelled set.
- C4. **Constants discipline.** The env var name lands in `internal/config/envkeys.go`; the
  timeout default in `internal/config/defaults.go`; endpoint/base URL/model from existing config
  constants. No inline literals at call sites.
- C5. **No new config-file key.** Activation is environment-only — no YAML section, no loader
  struct, no template change. This sharpens (does not conflict with) the parent SPEC's C5, which
  scoped its "no new config key" ruling to the auto-dispatch default.
- C6. **Scope discipline.** This SPEC touches `internal/cli` (new decider + selector + entry
  point wiring), `internal/config` (constants), and their tests. It does NOT touch
  `internal/kanban/factory_card.go`, any lease/claim path, or the parent SPEC's artifacts.

## §D Dependencies

- **depends_on: SPEC-TODO-CLASSIFY-DISPATCH-001** — the seam this SPEC extends
  (`todoCardDecider`, the fallback notice, the classification data model, the closed sets).
  Its frontmatter reads `status: completed` (measured at this tree, `681ee30b5`).
- **related: SPEC-AUTONOMY-CONTRACT-001** — the decider identity vocabulary (`llm`) this
  implementation gives a shipped backend. **SPEC-FACTORY-SELF-DISPATCH-001** — the MCP
  `todo_add` surface (REQ-SD-024) that shares the add path and inherits the same selection.

## §E Out of Scope

### Out of Scope — Jev and TypeSafe credential wiring

- `scripts/jev/` stays a local-only dev tool (AGENTS.local.md §29). `internal/jevcred` and the
  TypeSafe key file are not consumed by this decider; `internal/jev` gains no consumer; no
  template mirror of any of it.

### Out of Scope — re-classification and scheduling intelligence

- Classification happens at creation time only. No edit-time re-classification, no priority
  score, no deadline system, no scheduler reading the judgment (the parent SPEC's Out of Scope
  carries forward unchanged).

### Out of Scope — multi-provider decider support

- One transport: the configured GLM endpoint with `internal/glmcred`. No provider abstraction,
  no OpenAI/Anthropic-direct endpoints, no per-decider model configuration surface, no
  endpoint override flag.

### Out of Scope — prompt-tuning and accuracy benchmarking

- No classification-accuracy eval harness, no prompt A/B machinery. The recorded `reason` is
  the model's one-line rationale, bounded to a single line — nothing more.

### Out of Scope — asynchronous or retried classification

- The judgment is synchronous with the add. No background job, no retry queue, no
  re-classification of already-admitted cards, no decider result cache.

## §F Cross-references

- Acceptance criteria: `acceptance.md` (AC-TLD-001..007, Given-When-Then, binary-testable,
  RED-now/green adoption table).
- Implementation plan, OD resolutions, and concurrent-lane risk record: `plan.md`.
- Parent SPEC (read-only): `.moai/specs/SPEC-TODO-CLASSIFY-DISPATCH-001/` — seam, closed sets,
  fail-safe semantics, REQ-TCD-012 amendment provenance.
