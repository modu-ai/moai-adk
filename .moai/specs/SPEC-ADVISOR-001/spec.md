---
id: SPEC-ADVISOR-001
version: "0.1.0"
status: draft
created_at: 2026-04-30
updated_at: 2026-04-30
author: manager-spec
priority: High
labels: [advisor, cost-optimization, multi-model, manager-spec, evaluator-active, opus-4-7]
issue_number: null
wave: 1
tier: 0
scope:
  - .claude/agents/moai/manager-spec.md
  - .claude/agents/moai/evaluator-active.md
  - .claude/rules/moai/development/model-policy.md
  - .moai/config/sections/observability.yaml
  - internal/hook/post_tool.go
blockedBy: []
dependents: []
---

# SPEC-ADVISOR-001: Advisor Strategy Adoption (Sonnet executor + Opus advisor)

## HISTORY

- 2026-04-30: Initial draft. Wave 1 — Tier 0. Pilot scope: manager-spec, evaluator-active. Source: Anthropic "The Advisor Strategy" blog.

---

## 1. Overview

Adopt Anthropic's **Advisor Strategy** in MoAI-ADK to reduce Opus consumption while preserving (or improving) reasoning quality at high-uncertainty decision points. The strategy splits agent execution into two roles:

- **Executor** (Sonnet or Haiku): produces work, calls tools, emits user-facing output, maintains the agent session.
- **Advisor** (Opus): reads the executor's shared context on demand, returns guidance only — no tools, no user-facing output, no side effects.

Anthropic's measured baseline (verbatim, see research.md §1.1):
- Sonnet executor + Opus advisor: **2.7% quality improvement** on SWE-bench Multilingual, **11.9% cost reduction** vs Sonnet solo.
- Haiku executor + Opus advisor: **41.2% on BrowseComp** (vs 19.7% solo), **85% cheaper** than Sonnet solo.

This SPEC pilots the strategy on two MoAI-ADK agents whose decision surface is well-bounded:
1. **manager-spec** — currently runs `model: opus, effort: xhigh`; promoted to Sonnet executor + Opus advisor (max_uses: 3 per SPEC).
2. **evaluator-active** — currently runs `model: sonnet, effort: high`; gains an Opus advisor budget (max_uses: 2 per evaluation) for the four scored dimensions.

---

## 2. Problem Statement

### 2.1 Current State

Four MoAI agents declare `model: opus` directly: manager-spec, manager-strategy, expert-security, researcher. Plan-phase reasoning agents thus consume Opus across their full execution, including mechanical work (frontmatter assembly, schema validation, file scaffolding) where Sonnet is sufficient. evaluator-active, conversely, is Sonnet but scores four dimensions with hard FAIL thresholds — judgment calls where bounded Opus consultation would tighten verdicts without inflating cost.

### 2.2 Pain Points

- **P1 — Cost floor too high**: Plan-phase Opus runs cost 4-5x more than Sonnet runs of equivalent token count. Across an active codebase (50+ SPECs/year), this is the dominant orchestration cost line.
- **P2 — Reasoning is uniformly priced**: A SPEC frontmatter line and an EARS-pattern selection both cost the same Opus tokens, despite asymmetric reasoning need.
- **P3 — Evaluator score drift risk**: Sonnet alone, under uncertainty pressure on dimension boundaries (see research.md §2.2), may rationalize verdicts. Opus consultation at scoring decision points is a known mitigation that is not yet wired in.
- **P4 — Pattern unmeasured locally**: Anthropic's claimed savings (11.9% cost, 2.7% quality) cannot be replicated without per-agent telemetry distinguishing executor cost from advisor cost.

### 2.3 Why now

Wave 1 — Tier 0. Highest immediate value: the cost reduction is realized on every plan invocation; the quality improvement applies to every evaluator pass. No upstream dependency. Hooks-side observability work (`internal/hook/post_tool.go`) is in flight and naturally absorbs the per-call model-attribution extension.

---

## 3. Requirements (EARS)

### 3.1 Ubiquitous

- **REQ-ADV-001** [Ubiquitous] THE SYSTEM SHALL document the advisor protocol contract in `.claude/rules/moai/development/model-policy.md` so that advisor invocations are auditable.

- **REQ-ADV-002** [Ubiquitous] WHILE an advisor session is active, THE ADVISOR SHALL NOT call tools, SHALL NOT emit user-facing output, and SHALL NOT modify files. The advisor's only output is structured guidance returned to the executor.

### 3.2 Event-Driven

- **REQ-ADV-003** [Event-Driven] WHEN an executor agent (Sonnet or Haiku) encounters a decision flagged as high-uncertainty per §3.4 trigger conditions, THE EXECUTOR SHALL invoke the configured Opus advisor with the executor's current shared session context.

- **REQ-ADV-004** [Event-Driven] WHEN the advisor returns guidance, THE EXECUTOR SHALL incorporate the guidance into its reasoning AND SHALL record (advisor_invocations += 1, advisor_tokens += response_tokens, advisor_latency_ms) into the post-tool observability log.

- **REQ-ADV-005** [Event-Driven] WHEN the executor reaches the configured `max_uses` cap, THE EXECUTOR SHALL proceed using its own primary reasoning AND SHALL emit a structured note `advisor_cap_reached: true` in the agent's completion summary.

### 3.3 State-Driven / Where

- **REQ-ADV-006** [Where] WHERE an agent's frontmatter declares an `advisor` directive (model, max_uses, trigger_class), THE ORCHESTRATOR SHALL load that configuration into the agent's runtime context at invocation time.

- **REQ-ADV-007** [Where] WHERE Claude Code's sub-agent runtime supports `advisor_20260301` (verified by §3.5 verification step), THE PILOT AGENTS SHALL declare advisor configuration via `tools` field per Anthropic's API pattern. WHERE the runtime does NOT support this tool type in sub-agent context, THE PILOT AGENTS SHALL use the fallback Agent() spawn pattern (§3.6).

- **REQ-ADV-008** [Where] WHERE the advisor protocol is unavailable AND the fallback Agent() pattern also fails (network, rate limit, schema mismatch), THE EXECUTOR SHALL proceed without consultation, log `advisor_fallback: failed`, and complete the task with primary reasoning only — never blocking on advisor availability.

### 3.4 Trigger Conditions (Decision Uncertainty)

The following conditions MUST trigger advisor consultation (subject to `max_uses` cap):

| Pilot agent | Trigger class | Source of uncertainty | Expected calls per invocation |
|-------------|---------------|------------------------|------------------------------|
| manager-spec | EARS pattern selection ambiguity (Event-Driven vs State-Driven vs Complex) | research.md §2.3 | 1-2 |
| manager-spec | SPEC vs Report classification borderline | research.md §2.3 | 0-1 |
| manager-spec | Exclusions section: feature-vs-exclusion judgment call | research.md §2.3 | 0-1 |
| evaluator-active | Functionality dimension boundary (0.75 vs 0.50) | research.md §2.2 | 0-1 |
| evaluator-active | Security severity classification (Critical vs High borderline) | research.md §2.2 | 0-1 |
| evaluator-active | Consistency violation severity | research.md §2.2 | 0-1 |

- **REQ-ADV-009** [Event-Driven] WHEN a decision falls outside the trigger classes listed above, THE EXECUTOR SHALL NOT invoke the advisor (preserves cost discipline).

### 3.5 Compatibility Verification (Path Selection)

- **REQ-ADV-010** [Ubiquitous] THE PROJECT SHALL include a verification probe (Bash command in plan.md) that confirms whether `advisor_20260301` is exposed inside Claude Code sub-agent contexts BEFORE migrating either pilot agent.

- **REQ-ADV-011** [Event-Driven] WHEN verification confirms native support, THE PILOT AGENTS SHALL use the Anthropic-native pattern (research.md §1.1 verbatim API).

- **REQ-ADV-012** [Event-Driven] WHEN verification fails or returns ambiguous, THE PILOT AGENTS SHALL use the Agent() fallback pattern (§3.6) AND THE PROJECT SHALL file an upstream feedback note for Anthropic Claude Code team.

### 3.6 Fallback Pattern

- **REQ-ADV-013** [Where] WHERE fallback is active, THE EXECUTOR SHALL emit guidance request via `Agent(subagent_type: "researcher", model: "opus", mode: "plan", description: "advisor consultation: <decision class>")`. The researcher agent receives a focused decision question, returns structured guidance, and exits without writing any files.

### 3.7 Telemetry / Observability

- **REQ-ADV-014** [Ubiquitous] THE OBSERVABILITY LOG (`.moai/observability/hook-metrics.jsonl` per `.claude/rules/moai/core/settings-management.md`) SHALL record per agent invocation: `executor_model`, `executor_tokens`, `advisor_model`, `advisor_invocations`, `advisor_tokens`, `advisor_latency_ms_total`, `advisor_cap_reached`.

- **REQ-ADV-015** [Unwanted] THE TELEMETRY LOG SHALL NOT record raw advisor responses, prompts, or any API key material — only token counts and model identifiers.

### 3.8 Quality Guardrails

- **REQ-ADV-016** [Unwanted] THE EXECUTOR SHALL NOT treat advisor output as authoritative override of acceptance criteria, SPEC schema, or evaluator must-pass thresholds. The advisor advises; the executor's own primary reasoning remains accountable.

- **REQ-ADV-017** [State-Driven] WHILE the pilot is active, IF the dimension-score variance for evaluator-active exceeds 0.10 vs the pre-pilot baseline over a window of 10 evaluations, THEN the pilot SHALL roll back to the previous configuration AND a regression report SHALL be filed in `.moai/reports/advisor-regression-{DATE}/`.

---

## 4. Acceptance Criteria

(Detailed Given-When-Then scenarios live in `acceptance.md`. This is the summary list.)

- **AC-1**: Verification probe (REQ-ADV-010) executes and produces a recorded yes/no verdict on `advisor_20260301` availability in Claude Code sub-agent context. Probe artifact stored under `.moai/reports/advisor-probe-{DATE}/`.
- **AC-2**: After pilot rollout, manager-spec produces 5 consecutive valid SPECs (passing plan-auditor schema validation) using Sonnet executor + Opus advisor configuration.
- **AC-3**: After pilot rollout, evaluator-active produces 5 consecutive evaluations whose dimension scores differ by no more than 0.10 from a parallel Sonnet-only baseline run on the same artifacts.
- **AC-4**: Per-invocation cost telemetry is recorded. Aggregated over a 10-invocation window, manager-spec pilot achieves **>= 30% Opus token reduction** vs the pre-pilot baseline (lower bar than Anthropic's 11.9% cost-reduction figure because manager-spec was 100% Opus, not Sonnet executor; we expect outsized savings).
- **AC-5**: `max_uses` cap is honored: the orchestrator rejects an executor's 4th advisor call within a single manager-spec invocation, and the agent completes successfully without it.
- **AC-6**: Fallback path (REQ-ADV-013) is exercised at least once via fault injection (e.g., temporary `advisor_20260301` mock failure) and the executor still produces valid output.
- **AC-7**: Observability schema (REQ-ADV-014) extension is documented in `.claude/rules/moai/core/settings-management.md` and validated via a unit test in `internal/hook/post_tool_test.go` (or equivalent).

---

## 5. REQ-ID Matrix

| REQ-ID | Type | Priority | Verification | Acceptance Criterion |
|--------|------|----------|--------------|----------------------|
| REQ-ADV-001 | Ubiquitous | High | Document inspection | AC-7 |
| REQ-ADV-002 | Ubiquitous | Critical | Protocol audit | AC-2, AC-3 |
| REQ-ADV-003 | Event-Driven | High | Trace inspection (telemetry log) | AC-2, AC-3 |
| REQ-ADV-004 | Event-Driven | High | Telemetry validation | AC-4, AC-7 |
| REQ-ADV-005 | Event-Driven | High | Cap-injection test | AC-5 |
| REQ-ADV-006 | Where | High | Frontmatter schema test | AC-1 |
| REQ-ADV-007 | Where | High | Probe + branching | AC-1 |
| REQ-ADV-008 | Where | Critical | Fault-injection test | AC-6 |
| REQ-ADV-009 | Event-Driven | Medium | Trace inspection | AC-4 |
| REQ-ADV-010 | Ubiquitous | Critical | Probe artifact | AC-1 |
| REQ-ADV-011 | Event-Driven | High | Configuration audit | AC-2, AC-3 |
| REQ-ADV-012 | Event-Driven | High | Fallback artifact | AC-6 |
| REQ-ADV-013 | Where | High | Trace inspection | AC-6 |
| REQ-ADV-014 | Ubiquitous | High | Schema test, log inspection | AC-4, AC-7 |
| REQ-ADV-015 | Unwanted | Critical | Log scan for forbidden fields | AC-7 |
| REQ-ADV-016 | Unwanted | Critical | Audit trail review | AC-3 |
| REQ-ADV-017 | State-Driven | High | Variance computation, rollback artifact | AC-3 |

**Total**: 17 requirements (4 Ubiquitous, 6 Event-Driven, 5 Where, 1 State-Driven, 1 Unwanted) — wait, recount: 1 State-Driven, 3 Unwanted (REQ-002 is structural prohibition reframed as Ubiquitous). Net distribution covers all five EARS patterns except Optional (none warranted).

---

## 6. Out of Scope (Exclusions — What NOT to Build)

- **EX-1**: expert-security pilot is **explicitly out of scope** for this SPEC. Security FAIL has hard threshold; until pilot data exists, expert-security retains direct Opus. Deferred to Wave 2.
- **EX-2**: manager-strategy and researcher are out of scope. Their decision surface is broader; advisor caps are harder to calibrate without manager-spec/evaluator-active pilot data.
- **EX-3**: Haiku-executor configurations are out of scope. The pilot starts with Sonnet executors. Haiku migration is a follow-up SPEC contingent on pilot success.
- **EX-4**: A user-facing CLI command (`moai advisor stats`) is out of scope. Telemetry is logged to JSONL only; aggregation tooling deferred.
- **EX-5**: Modifying `model: opus` declarations of other agents (manager-strategy, expert-security, researcher) is out of scope. Only manager-spec changes from Opus → Sonnet+Advisor.
- **EX-6**: KV-cache optimization for advisor sessions is out of scope. Anthropic's reported cost savings depend on cache reuse; we measure what we get and revisit if savings fall below AC-4 threshold.
- **EX-7**: Building a custom advisor protocol (Path D in research.md §3.4 — skill-level wrapper invoking direct Anthropic API) is out of scope. Path C is preferred; Path B (Agent() fallback) is the only contingency.

---

## 7. Open Questions

- **OQ-1**: Does Claude Code expose `advisor_20260301` to sub-agent contexts? **Resolution path**: REQ-ADV-010 verification probe in plan.md.
- **OQ-2**: Should `max_uses` be a per-agent default in frontmatter or a global default with per-agent override in `.moai/config/sections/model-policy.yaml`? **Decision deferred to plan.md design step.**
- **OQ-3**: How should the executor signal advisor consultation in its user-visible output? Suggestions: silent (advisor is internal), footnote ("(advisor consulted on EARS pattern selection)"), or structured field in completion summary. **User feedback solicited via AskUserQuestion in plan phase.**
- **OQ-4**: When the advisor's guidance contradicts the executor's reasoning, what is the resolution policy? Options: (a) defer to advisor (overrides executor), (b) defer to executor (advisor is advisory only), (c) escalate to user via AskUserQuestion. REQ-ADV-016 currently selects (b); confirm in plan phase.
- **OQ-5**: For evaluator-active, is the advisor consulted **before** dimension scoring (input-side) or **after** (review-side)? Anthropic's pattern is advisor-during-execution. Pilot will adopt input-side (advisor consulted at boundary cases during scoring) but may switch to review-side if input-side biases scores.

---

**Total lines**: ~225
**Status**: draft — awaiting plan-auditor review
