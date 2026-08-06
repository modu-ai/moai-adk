---
id: SPEC-AUDIT-MULTI-MODEL-001
title: "Multi-model audit convergence — parallel cross-backend fan-out + disagreement synthesis for audit_model: multi (super-review pattern)"
version: "0.1.0"
status: in-progress
created: 2026-08-06
updated: 2026-08-07
author: manager-spec
priority: P1
phase: "v3.2 target"
module: internal/cli
lifecycle: spec-anchored
tier: L
tags: "audit, multi-model, convergence, super-review, cross-model, fail-open, ssot"
---

# SPEC-AUDIT-MULTI-MODEL-001 — Multi-model audit convergence (autonomy·speed Epic, P1 residual)

## HISTORY

- 2026-08-06 (plan-phase, iter-1) — Initial Tier L authoring. Operationalizes the deferred `audit_model: multi` convergence orchestration that SPEC-MOAI-MCP-SERVER-001 (PR #1378, status: completed) explicitly declared as a stored-but-not-orchestrated token. Design source: `.moai/reports/moai-autonomy-workflow-redesign-20260803.html` §3.4 (Codex 감사 위임), §3.6 v3 extension (~lines 421-439 — 3-way gate, default required, fail-open), Q1 (~lines 490-498 — three routing paths A/B/C). This SPEC layers ONLY the `multi` parallel orchestration + convergence on top of the already-shipped single-backend infrastructure; it does NOT re-implement what MOAI-MCP-SERVER delivered.

## §A. User Story

**As a** MoAI-ADK user who has opted into cross-model audit (`audit_model: multi`) because correlated model blind spots (§2.4 of the design report — "Codex is not smarter, it is differently smart") are a known failure mode of single-model review,

**I want** the moai MCP server to run two-or-more audit backends in parallel (claude-in-session + codex + glm, per their `audit_gate`) and converge their verdicts into a single synthesized result with disagreement surfaced as Verification Matrix residual-risk + advisory (NOT a hard block),

**so that** (a) the super-review pattern (Drew Hyde [R3]: Claude primary → independent secondary → orchestrator synthesis) is reachable from both plan-audit and sync-audit through the identical MCP surface; (b) the fully-autonomous goal-convergence loop has a multi-model safety net (Path C — Stop-hook gate extending the codex-review-gate pattern); and (c) cross-model disagreement NEVER interrupts the autonomous flow — it is surfaced as advisory signal, preserving the fail-open identity (a missing/unauthenticated optional backend falls back to active backends, and a split verdict falls back to the required-gate contract).

## §B. Scope Summary

This SPEC delivers the **convergence orchestration** only — the parallel fan-out engine across the three audit backends (claude always present as the in-session verdict + codex + glm per their gates), the convergence algorithm (all-required-PASS → PASS; any required FAIL → FAIL; disagreement among required backends → advisory residual-risk, NOT block), the `audit_multi` MCP tool, the cross-model Skill that wires plan-audit/sync-audit to it, and the multi-review-gate Stop hook for fully-autonomous goal convergence.

The single-backend infrastructure (stdio server, codex_audit, glm_audit, audit_model/audit_gate enums, `VerdictInconclusive → claude` fail-open, `ResolveAgentModelEffort` SSOT, codex-review-gate Stop hook) is **already shipped by SPEC-MOAI-MCP-SERVER-001** and is consumed here, NOT re-implemented. The boundary is made explicit in §C (quoted verbatim from the deferral ACs).

### Out of Scope — single-backend audit infrastructure (SPEC-MOAI-MCP-SERVER-001, completed)

- The stdio JSON-RPC server, the `mark3labs/mcp-go` SDK integration, the per-backend `audit_gate ∈ {off, advisory, required}`, the `audit_model ∈ {claude, codex, glm, multi}` enum (the `multi` token was declared there), the codex binary shellout (`codex_audit`), the GLM direct z.ai API call (`glm_audit`), the `VerdictInconclusive → claude` fail-open, and the codex-review-gate Stop hook → **all consumed from SPEC-MOAI-MCP-SERVER-001, NOT rebuilt here**. The only change to that surface is flipping the `multiConvergenceImplemented` sentinel (`internal/cli/mcp_audit.go:31`) from `false` to `true` — a one-line, grep-visible declaration that the convergence logic now exists.

### Out of Scope — audit speed optimizations (A1-A11)

- Report §3.5 audit optimizations (sticky hash cache, skip-threshold alignment, 4dim binding promotion, shared diagnostic snapshot, docs ∥ audit parallelization, etc.) → **separate SPEC**. This SPEC reuses `runtime.AuditCache` and the per-tier PASS thresholds read-only where they already exist; it does not change caching or skip policy.

### Out of Scope — trend MCP / third-party bundling

- Report §3.7 trend MCP tools (Playwright, ast-grep, Semgrep, GitHub, Postgres/Sentry bundling) → **SPEC-TREND-MCP**. This SPEC touches only the moai MCP server's audit surface; no third-party entries.

### Out of Scope — web live WebSocket dashboard

- The real-time `moai web` board (WebSocket + fsnotify over `.moai/state/`) → **v3.1 deployment target, SPEC-WEB-LIVE-BOARD**. The convergence result is readable from the web board via the same `internal/` functions, but the board is NOT required for convergence and vice versa.

### Out of Scope — cross-model adversarial-review prompt engineering

- The design of the adversarial-review prompt body (what questions the secondary model is asked) is owned by the `moai-ref-cross-model-audit` Skill content (M3), NOT by this SPEC's requirements. This SPEC owns the MECHANISM (the skill loads, the skill body calls the MCP tool, independence is preserved); the skill's prompt content is an evolvable skill-body asset.

## §C. Boundary — verbatim deferral quote from SPEC-MOAI-MCP-SERVER-001

The two acceptance criteria that demarcate the boundary are quoted verbatim:

> **AC-MCP-012** (MUST) — `audit_model` + `audit_gate` enums + default profile
> - Given the audit config schema,
> - When `audit_model` and per-auditor `audit_gate` are validated,
> - Then `audit_model` accepts exactly `{claude, codex, glm, multi}`, `audit_gate` accepts exactly `{off, advisory, required}`, the default `audit_gate` is `required`, and the default profile is claude + codex required, glm advisory. **`multi` is accepted as a token but its convergence logic is NOT implemented here.**

> **AC-MCP-017** (SHOULD) — 3-way selection resolves a single active backend
> - Given `audit_model` is one of `claude` / `codex` / `glm`,
> - When an audit runs,
> - Then exactly one backend executes (per its `audit_gate`); **`audit_model: multi` is accepted as a stored value but its parallel-orchestration behavior is NOT implemented in this SPEC (deferred to SPEC-AUDIT-MULTI-MODEL).**

This SPEC authors exactly that deferred parallel-orchestration + convergence logic. Every requirement below is additive to the MOAI-MCP-SERVER surface; none re-implements a single-backend capability.

## §D. Requirements (GEARS)

> Domain prefix `REQ-AMM-NNN` maps to SPEC domain `AUDIT-MULTI-MODEL`. GEARS compound clauses are used throughout. Every "wraps existing `internal/` function" requirement is backed by a verified file:line citation in `research.md`; integration points were re-verified against the worktree at `f85ff4c3e` (branch `feat/spec-audit-multi-model`, 2026-08-06).

### M0 — Design lock (convergence data model + algorithm)

**REQ-AMM-001** (Ubiquitous) The convergence engine SHALL be a synthesis function over N backend verdicts each shaped per `review-output.schema.json` (`verdict` / `summary` / `findings[severity,title,body,file,line,confidence,recommendation]` / `next_steps`), producing a single `ConvergenceResult` that carries `per_backend_verdicts[]`, an `overall_verdict`, a `disagreement_flag`, and a `residual_risk_note`.

### M1 — Convergence engine core (parallel fan-out + fail-open composition)

**REQ-AMM-002** (Event-driven) **When** `audit_model: multi` is selected and the `audit_multi` MCP tool is invoked, the engine SHALL fan out across the active backends in parallel — codex via `codex_audit` (per its `audit_gate`), GLM via the GLM direct z.ai API call (per its `audit_gate`) — and SHALL accept the in-session claude verdict as a synthesis input (the claude verdict is produced by the auditor agent itself, NOT by a separate MCP call).

**REQ-AMM-003** (Unwanted) The convergence engine SHALL NOT pass the claude verdict (the in-session analysis) to the codex or GLM backends as input context — super-review independence (Drew Hyde [R3]: "Claude analysis undisclosed to preserve independence") MUST be preserved so the secondary backends produce uncorrelated second opinions, not contaminated re-samples.

**REQ-AMM-004** (Event-detected) **When** a selected non-Claude backend (codex or glm) is missing, unauthenticated, errors, or returns a malformed response, the engine SHALL treat that backend's verdict as `VerdictInconclusive` (reusing the existing constant from `internal/cli/mcp_codex.go`) and SHALL continue convergence over the remaining active backends (fail-open mandatory — a missing optional backend NEVER hard-blocks the workflow).

**REQ-AMM-005** (State-driven) **While** resolving model and effort for any backend invocation, the engine SHALL resolve through `template.ResolveAgentModelEffort` (`internal/template/profile_matrix.go:385`) — the identical interpreter used by the web console, the launcher, and the single-backend MCP handlers — and SHALL NOT read agent frontmatter or `llm.agent_overrides` directly (fork risk — the same constraint as REQ-MCP-013).

### M2 — Convergence algorithm (disagreement = advisory, NOT block)

**REQ-AMM-006** (Capability gate) **Where** the per-backend verdicts are collected, the convergence algorithm SHALL derive `overall_verdict` by the following ordered policy:
1. If every backend whose `audit_gate == required` returned `PASS` (and no required backend returned `FAIL`) → `overall_verdict = PASS`.
2. If any backend whose `audit_gate == required` returned `FAIL` → `overall_verdict = FAIL`.
3. If the required backends disagree (at least one required `PASS` and at least one required `FAIL`) → `overall_verdict = FAIL` (the required-gate contract holds per backend), **AND** the `disagreement_flag` SHALL be set, **AND** the `residual_risk_note` SHALL describe the split so the orchestrator can surface it as advisory.
4. Backends whose `audit_gate == advisory` or `off` NEVER flip `overall_verdict` to `FAIL`; their verdicts are recorded in `per_backend_verdicts[]` for transparency and contribute to the `disagreement_flag` only when they conflict with a required-backend verdict.

**REQ-AMM-007** (Ubiquitous) The `disagreement_flag` and `residual_risk_note` SHALL be surfaced as Verification Matrix residual-risk + advisory signal in the orchestrator's Completion Report — they SHALL NOT produce a hard BLOCK on the autonomous flow (cross-model disagreement is information, not a gate; the per-backend `audit_gate: required` already governs block behavior per backend, and a split among required backends falls back to the conservative `FAIL` per REQ-AMM-006 #2 rather than inventing a new "disagreement-block" category).

**REQ-AMM-008** (Unwanted) The engine SHALL NOT invent a new `VerdictDisagreement` enum value; the `review-output.schema.json` `verdict` field is the SSOT, and disagreement is captured by the `disagreement_flag` boolean on `ConvergenceResult` (additive — the per-backend verdicts each remain one of the existing `pass`/`fail`/`inconclusive`/etc. values).

### M3 — MCP tool surface (`audit_multi`)

**REQ-AMM-009** (Event-driven) **When** the MCP host requests `tools/list`, the server SHALL declare an `audit_multi` tool with a JSON Schema declaring inputs (`claude_verdict` object, `target` enum, optional `focus` string) and a structured `ConvergenceResult` output shape, so an AI caller invokes it type-safely with zero flag-guessing.

**REQ-AMM-010** (State-driven) **While** the `audit_multi` tool is a thin wrapper, its handler SHALL (a) read the active `audit_model` + per-auditor `audit_gate` from config, (b) fan out to codex/glm backends only when their gate is not `off`, (c) pass the result to the convergence engine, and (d) return the `ConvergenceResult` — reusing the single-backend handlers (`codex_audit`, GLM audit) verbatim, NOT re-implementing their binary-shellout / API-call internals.

### M4 — plan-audit + sync-audit cross-model wiring (Path A)

**REQ-AMM-011** (Capability gate) **Where** the plan-auditor or sync-auditor agent is spawned and the project has opted into `audit_model: multi`, the agent SHALL load the `moai-ref-cross-model-audit` Skill (new — M5) whose body instructs the agent to call `mcp__moai__audit_multi` with its own in-session claude analysis as `claude_verdict`, and to fold the returned `ConvergenceResult` into its audit verdict + the Verification Matrix residual-risk surface — preserving the existing skill-routing protocol 100% (no new routing mechanism).

**REQ-AMM-012** (Unwanted) The Skill body SHALL NOT instruct the agent to share its full Claude analysis text with the secondary backends; it SHALL pass only the `claude_verdict` object (synthesized verdict + summary) as an input to the convergence engine, and the convergence engine uses `claude_verdict` ONLY for the synthesis step — never as prompt context for codex/glm (super-review independence, REQ-AMM-003).

### M5 — Fully-autonomous goal convergence gate (Path C)

**REQ-AMM-013** (State-driven) **While** `workflow.multi_review_gate.enabled` is set (opt-in, BranchGuard pattern — the same pattern as `workflow.codex.review_gate.enabled`), the `moai hook multi-review-gate` Stop hook SHALL enforce an ALLOW/BLOCK contract with the same mandatory self-gate as the codex-review-gate ("the previous turn produced no code edit / is a status report / is a review-result ⇒ ALLOW immediately") to prevent false blocks, with a 900 s timeout override (the moai-default 5 s does not apply to this hook).

**REQ-AMM-014** (Event-driven) **When** the multi-review-gate fires on a code-edit turn, it SHALL read the most recent `ConvergenceResult`, apply the convergence policy (REQ-AMM-006), and emit ALLOW (all required backends PASS) or BLOCK (any required backend FAIL); a disagreement among required backends (split verdict) produces `overall_verdict = FAIL` per REQ-AMM-006 #2 and the gate BLOCKs conservatively — disagreement among advisory-only backends NEVER BLOCKs (it surfaces as advisory).

**REQ-AMM-015** (Event-detected) **When** all non-Claude backends are missing or unauthenticated at gate-fire time, the multi-review-gate SHALL fail open to claude-only evaluation (ALLOW if the in-session claude verdict is PASS; BLOCK only if it is FAIL) — the autonomous flow is NEVER hard-blocked on a missing optional dependency.

### M6 — Skill authoring + Template-First (Path A skill body)

**REQ-AMM-016** (Ubiquitous) The `moai-ref-cross-model-audit` Skill SHALL be authored under template source (`internal/template/templates/.claude/skills/moai-ref-cross-model-audit/`), mirrored to the local `.claude/skills/` tree, and regenerated via `make build` — per CLAUDE.local.md §2 Template-First; the skill body SHALL pass §25 template-neutrality (no SPEC-ID, no commit SHA, no internal date, no macOS-bias path) enforced by the CI guard.

**REQ-AMM-017** (Unwanted) The Skill body SHALL NOT hardcode the MCP tool name as a string literal embedded in prose-only fashion; it SHALL use the canonical `mcp__moai__audit_multi` tool reference so a future tool-rename propagates mechanically via grep, and SHALL state the independence-preservation rule (pass only the synthesized `claude_verdict`, not the full analysis) verbatim so the rule is auditable from the skill body.

### Cross-cutting

**REQ-AMM-018** (Unwanted) MCP tool handlers and the convergence engine SHALL NOT invoke `AskUserQuestion` or emit free-form user-facing questions (subagent boundary — REQ-MCP-014 carried forward); on a missing-input or inconclusive condition the tool returns a structured `ConvergenceResult` (including `VerdictInconclusive` per backend) and the orchestrator translates it through its own `AskUserQuestion` channel.

**REQ-AMM-019** (Capability gate) **Where** any new env-var name, threshold, or default is introduced by this SPEC, it SHALL be defined as a constant in `internal/config/envkeys.go` (env-var names) or `internal/config/defaults.go` (thresholds/defaults) per CLAUDE.local.md §14 hardcoding prevention, and the `multi_review_gate` config block SHALL reuse the existing `workflow.codex.review_gate` structural pattern (no new schema shape — only a sibling `multi_review_gate` key under `workflow:`).

## §E. Constraints (non-functional)

- **C1 — Additive to MOAI-MCP-SERVER**: the stdio server, single-backend execution, `audit_model`/`audit_gate` enums, `VerdictInconclusive → claude` fail-open, and `ResolveAgentModelEffort` SSOT are consumed, NOT re-implemented. The only edit to `internal/cli/mcp_audit.go` is flipping the `multiConvergenceImplemented` sentinel to `true`.
- **C2 — fail-open identity**: a missing/unauthenticated optional backend (codex, glm) NEVER hard-blocks the convergence; it returns `VerdictInconclusive` for that backend and converges over the rest. Evidence of absence is not evidence of failure in either direction (the invariant binds both ways).
- **C3 — disagreement = advisory, NOT block**: per the fixed user decision, cross-model disagreement is surfaced as Verification Matrix residual-risk + advisory; the autonomous flow is NOT interrupted by disagreement. The per-backend `audit_gate: required` governs block behavior per backend; a split among required backends falls back to the conservative `overall_verdict = FAIL` (REQ-AMM-006 #2), not a new disagreement-block. The "advisory, not block" framing governs (a) the `disagreement_flag` signal itself and (b) conflicts where at least one disagreeing backend is advisory; it does NOT weaken the per-backend `audit_gate: required` contract — a required backend's FAIL still BLOCKs regardless of what other backends say, and a required-split is resolved conservatively (`overall_verdict = FAIL`) per REQ-AMM-006 #2.
- **C4 — super-review independence**: the claude verdict is NEVER passed to the codex/glm backends as prompt context; it is used ONLY for the synthesis step (Drew Hyde [R3]).
- **C5 — subagent boundary**: MCP tools return structured `ConvergenceResult`; they never call `AskUserQuestion`.
- **C6 — SSOT for model/effort**: `template.ResolveAgentModelEffort` (`profile_matrix.go:385`) is the sole interpreter; agent frontmatter and `llm.agent_overrides` are never read directly by convergence handlers (fork risk).
- **C7 — secret hygiene**: `${CODEX_API_KEY}` / `${GLM_API_KEY}` literals only (expanded by the Claude Code runtime); never serialized secrets. The `audit_multi` tool adds NO new env surface — it reuses the existing `buildAuditEnvBlock(config.AuditModelMulti)` two-literal block already shipped by MOAI-MCP-SERVER.
- **C8 — opt-in default-off**: the `audit_multi` tool is inert unless `audit_model: multi` is selected; the multi-review-gate ships inert to distributed users (lives in `settings.local.json`, never template-default).
- **C9 — Template-First + §25 neutrality**: any new Skill is mirrored to `internal/template/templates/` + `make build` + CI guard, and is generic/neutral.

## §F. Dependencies

- **SPEC-MOAI-MCP-SERVER-001 (status: completed, PR #1378)** — hard dependency. This SPEC consumes its stdio server, `codex_audit`, GLM audit, `audit_model`/`audit_gate` enums, `VerdictInconclusive`, `ResolveAgentModelEffort` SSOT, and the codex-review-gate Stop-hook pattern. The `multiConvergenceImplemented` sentinel at `internal/cli/mcp_audit.go:31` is flipped from `false` to `true` here.
- **Existing internal packages (read-mostly reuse)**: `internal/cli` (mcp_server.go, mcp_codex.go, mcp_glm.go, mcp_audit.go, codex_review_gate.go), `internal/config` (AuditModel/AuditGate enums at `internal/config/defaults.go:637-641` + `mcp_audit_config.go`), `internal/template` (ResolveAgentModelEffort, IsGLMBackend).
- **No NEW go.mod dependencies**: `mark3labs/mcp-go` is already in go.mod (landed by MOAI-MCP-SERVER). Parallel fan-out uses `errgroup` (already in go.mod — see `internal/cli/` and `internal/web/` callers).
- **External binaries (runtime, optional)**: `codex` CLI (for the codex backend; fail-open if absent), GLM API key (for the GLM backend; fail-open if absent).

## §G. Risks

- **R1 — Independence-preservation regression**: if a future edit accidentally passes `claude_verdict` into the codex/glm prompt context, the super-review independence is silently destroyed and the secondary verdicts become correlated re-samples. Mitigation: REQ-AMM-003 + REQ-AMM-012 (verbatim skill rule) + a run-phase test that asserts the codex/glm call payloads do NOT include the `claude_verdict` field.
- **R2 — Disagreement policy drift**: the fixed user decision (disagreement = advisory, NOT block) could be re-litigated by a future edit that promotes disagreement to a BLOCK. Mitigation: REQ-AMM-006 #3 + REQ-AMM-007 + C3 codify the policy in the SPEC; design.md §3 records the rationale so the reasoning survives; a run-phase test asserts `disagreement_flag == true` does NOT produce a Stop-hook BLOCK when the required-gate contract is otherwise satisfied.
- **R3 — Sentinel-flip blast radius**: flipping `multiConvergenceImplemented` from `false` to `true` is the one-line change that activates the entire surface. If the convergence engine is absent or broken, `audit_model: multi` now silently does nothing (the sentinel would be true but the engine missing). Mitigation: the sentinel flip lands in the SAME commit as the engine (M1) so there is no window where the sentinel lies.
- **R4 — Stop-hook self-gate false-block regression**: the multi-review-gate inherits the codex-review-gate self-gate; a regression there blocks non-edit turns. Mitigation: reuse the codex-review-gate self-gate logic verbatim (no new heuristic); the run-phase test suite includes the no-edit-turn ALLOW case (AC-AMM-013).
- **R5 — Verify-before-claim on integration points**: the report's file:line references are from 2026-08-03; the MCP-SERVER SPEC re-verified them at `b57de3ab1` (2026-08-05). This plan-phase re-verifies them again at `f85ff4c3e` (2026-08-06) in `research.md`; run-phase MUST re-confirm before forking code.

## §H. Cross-References

- Design source: `.moai/reports/moai-autonomy-workflow-redesign-20260803.html` §3.4 (Codex 감사 위임), §3.6 v3 extension, Q1 (three routing paths A/B/C).
- Research literature: §2.4 super-review pattern (Drew Hyde [R3]), §2.5 AgentOrchestra peer verification (arXiv:2506.12508 [R5]), cross-model adversarial review (Frontiers 10.3389/fcomp.2025.1655469).
- Hard dependency: SPEC-MOAI-MCP-SERVER-001 (completed) — `spec.md` REQ-MCP-010 (audit_model + audit_gate), `acceptance.md` AC-MCP-012 + AC-MCP-017 (the deferral boundary quoted in §C).
- Schema SSOT: `.claude/rules/moai/development/spec-frontmatter-schema.md`.
- Epic memory: `project_autonomy_mcp_server_next.md` (autonomy·speed Epic — this SPEC = P1 residual after MOAI-MCP-SERVER P2 closed).
