# Acceptance — SPEC-AUDIT-MULTI-MODEL-001

> Verification layer. Each criterion is `AC-AMM-NNN`, written as binary-testable `Given … When … Then …`. GEARS requirements live in `spec.md` (`REQ-AMM-NNN`); this file MUST NOT restate them as requirements — Given-When-Then is the correct format here (audit contract: `plan-auditor.md` M3 § Scope / MP-2). Traceability matrix + severity + Definition of Done at the end.

## Severity legend

- **MUST** — blocks `completed` (a failing MUST AC blocks convergence/merge).
- **SHOULD** — strong quality signal; a fail is a debt item, not a blocker.
- **MAY** — opportunistic / nice-to-have verification.

## §D. AC Matrix

### M0 — Design lock (convergence data model + algorithm)

**AC-AMM-001** (MUST) — convergence data model locked
- **Given** M0 completes,
- **When** the `ConvergenceResult` shape is inspected in `design.md` §1,
- **Then** it carries `per_backend_verdicts[]`, `overall_verdict`, `disagreement_flag`, and `residual_risk_note`, and the shape is referenced from `progress.md` §E.2 as a locked M0 decision. (REQ-AMM-001)

### M1 — Convergence engine core (parallel fan-out + fail-open composition)

**AC-AMM-002** (MUST) — parallel fan-out across active backends
- **Given** `audit_model: multi` and the `audit_multi` tool is invoked with a `claude_verdict`,
- **When** codex and GLM gates are both non-`off` and both backends are present and authenticated,
- **Then** the engine invokes `codex_audit` and the GLM direct z.ai API call in parallel (verifiable by timing: the wall-clock is closer to `max(t_codex, t_glm)` than to `t_codex + t_glm`), and the `per_backend_verdicts[]` array contains all three backends' verdicts. (REQ-AMM-002)

**AC-AMM-003** (MUST) — super-review independence preserved
- **Given** the `audit_multi` tool invocation with `claude_verdict` as input,
- **When** the codex and GLM backend calls are inspected (via test instrumentation or handler source),
- **Then** the `claude_verdict` field is NOT included in the codex/GLM prompt context — it is consumed ONLY by the synthesis step; the secondary backends produce uncorrelated second opinions. (REQ-AMM-003, C4)

**AC-AMM-004** (MUST) — fail-open on missing/unauthenticated backend
- **Given** a selected non-Claude backend (codex or glm) is missing, unauthenticated, errors, or returns a malformed response,
- **When** the convergence runs,
- **Then** that backend's verdict in `per_backend_verdicts[]` is `VerdictInconclusive` (reusing the existing constant from `internal/cli/mcp_codex.go`), the convergence continues over the remaining active backends, and the workflow does NOT hard-block on the missing optional dependency. (REQ-AMM-004, C2)

**AC-AMM-005** (MUST) — model/effort SSOT via `ResolveAgentModelEffort`
- **Given** the codex and GLM backend invocations,
- **When** they resolve model and effort,
- **Then** they call `template.ResolveAgentModelEffort` (`internal/template/profile_matrix.go:385`) and do NOT read agent frontmatter or `llm.agent_overrides` directly (verified by grep: no direct frontmatter/override read in the convergence-engine package). (REQ-AMM-005, C6)

### M2 — Convergence algorithm (disagreement = advisory, NOT block)

**AC-AMM-006** (MUST) — convergence verdict derivation — all required PASS
- **Given** all backends whose `audit_gate == required` return `PASS` (and no required backend returns `FAIL`),
- **When** the convergence algorithm derives `overall_verdict`,
- **Then** `overall_verdict == PASS` and `disagreement_flag == false`. (REQ-AMM-006 #1)

**AC-AMM-007** (MUST) — convergence verdict derivation — any required FAIL, no disagreement
- **Given** exactly one required backend returns `FAIL` and the other required backend(s) also return `FAIL` or `VerdictInconclusive`,
- **When** the convergence algorithm derives `overall_verdict`,
- **Then** `overall_verdict == FAIL` and the `residual_risk_note` records which backend(s) failed. (REQ-AMM-006 #2)

**AC-AMM-008** (MUST) — convergence verdict derivation — required-backends disagreement → FAIL + disagreement_flag (advisory, NOT a new block category)
- **Given** at least one required backend returns `PASS` and at least one required backend returns `FAIL` (a split),
- **When** the convergence algorithm derives `overall_verdict`,
- **Then** `overall_verdict == FAIL` (conservative — the required-gate contract holds per backend), `disagreement_flag == true`, and the `residual_risk_note` describes the split; the result does NOT introduce a new `VerdictDisagreement` enum value (the per-backend verdicts each remain one of the existing `pass`/`fail`/`inconclusive`/etc. values). (REQ-AMM-006 #3, REQ-AMM-008)

**AC-AMM-009** (MUST) — advisory-only backends never flip overall_verdict to FAIL
- **Given** a backend whose `audit_gate == advisory` returns `FAIL` while all required backends return `PASS`,
- **When** the convergence algorithm derives `overall_verdict`,
- **Then** `overall_verdict == PASS`, the advisory backend's `FAIL` is recorded in `per_backend_verdicts[]`, and `disagreement_flag == true` (the advisory-vs-required conflict is surfaced) but the result is NOT a BLOCK on the autonomous flow. (REQ-AMM-006 #4, REQ-AMM-007, C3)

**AC-AMM-010** (MUST) — disagreement surfaced as Verification Matrix residual-risk + advisory
- **Given** a `ConvergenceResult` with `disagreement_flag == true`,
- **When** the orchestrator renders its Completion Report / Verification Matrix,
- **Then** the disagreement is present as residual-risk + advisory (a row in the matrix + a `residual_risk_note`), and the autonomous flow is NOT interrupted by the disagreement signal alone (the flow proceeds unless a required backend's `FAIL` separately blocks it per REQ-AMM-006 #2). (REQ-AMM-007, C3)

**AC-AMM-011** (MUST) — no new VerdictDisagreement enum
- **Given** the `review-output.schema.json` `verdict` field,
- **When** the convergence engine and `ConvergenceResult` are inspected,
- **Then** no new `VerdictDisagreement` enum value exists; disagreement is captured by the `disagreement_flag` boolean on `ConvergenceResult` (additive — the per-backend verdicts each remain one of the existing values). (REQ-AMM-008)

### M3 — MCP tool surface (`audit_multi`)

**AC-AMM-012** (MUST) — `tools/list` declares `audit_multi` with JSON Schema
- **Given** the running `moai mcp-server` with the `audit_multi` tool registered,
- **When** the host requests `tools/list`,
- **Then** the response declares an `audit_multi` tool with a `name`, an `inputSchema` declaring `claude_verdict` (object), `target` (enum), optional `focus` (string), and a structured `ConvergenceResult` output shape. (REQ-AMM-009)

**AC-AMM-013** (MUST) — `audit_multi` is a thin wrapper (no backend re-implementation)
- **Given** the `audit_multi` handler source,
- **When** it is inspected,
- **Then** it reads the active `audit_model` + per-auditor `audit_gate` from config, fans out by calling the existing `codex_audit` and GLM-audit handlers (no re-implementation of the binary-shellout / API-call internals), passes the result to the convergence engine, and returns the `ConvergenceResult`. (REQ-AMM-010, C1)

**AC-AMM-014** (MUST) — `audit_multi` respects per-auditor `audit_gate`
- **Given** `audit_model: multi` with `audit_gate.codex == off`,
- **When** the `audit_multi` tool runs,
- **Then** the codex backend is NOT invoked, `per_backend_verdicts[]` carries no codex entry, and the convergence proceeds over claude + glm only. (REQ-AMM-010)

### M4 — plan-audit + sync-audit cross-model wiring (Path A)

**AC-AMM-015** (MUST) — plan-auditor cross-model path via Skill
- **Given** the plan-auditor agent is spawned and the project has `audit_model: multi`,
- **When** the agent reaches the audit step,
- **Then** the agent loads the `moai-ref-cross-model-audit` Skill (per the existing skill-routing protocol — no new routing mechanism), the skill body instructs the agent to call `mcp__moai__audit_multi` with its own in-session claude analysis as `claude_verdict`, and the agent folds the returned `ConvergenceResult` into its audit verdict + Verification Matrix residual-risk surface. (REQ-AMM-011)

**AC-AMM-016** (MUST) — sync-auditor cross-model path via the same Skill
- **Given** the sync-auditor agent is spawned and the project has `audit_model: multi`,
- **When** the agent reaches its 4-dimension quality read,
- **Then** the same `moai-ref-cross-model-audit` Skill is loaded (single skill, both audit entry points — no duplication) and the convergence result is folded into the 4-dimension verdict. (REQ-AMM-011)

**AC-AMM-017** (MUST) — skill body preserves independence verbatim
- **Given** the `moai-ref-cross-model-audit` Skill body,
- **When** it is inspected,
- **Then** it passes only the synthesized `claude_verdict` object to the MCP tool (NOT the full Claude analysis text as prompt context for secondary backends), and the skill body states the independence-preservation rule in plain prose so it is auditable from the skill body. (REQ-AMM-012, C4)

### M5 — Fully-autonomous goal convergence gate (Path C)

**AC-AMM-018** (MUST) — multi-review-gate opt-in + BranchGuard pattern + 900 s timeout
- **Given** the `moai hook multi-review-gate` Stop hook configuration,
- **When** its manifest is inspected,
- **Then** it is opt-in (`workflow.multi_review_gate.enabled`, BranchGuard pattern — sibling to `workflow.codex.review_gate`), the moai-default 5 s hook timeout is overridden to 900 s for this hook only, and it emits the standard ALLOW/BLOCK contract. (REQ-AMM-013)

**AC-AMM-019** (MUST) — review-gate self-gate prevents false blocks
- **Given** `workflow.multi_review_gate.enabled` is set and the previous turn produced no code edit (status report / review-result / no-op),
- **When** the `moai hook multi-review-gate` Stop hook fires,
- **Then** it ALLOWs immediately without invoking any backend (the mandatory self-gate — same heuristic as codex-review-gate), so a non-editing turn is never falsely blocked. (REQ-AMM-013, R4)

**AC-AMM-020** (MUST) — gate verdict follows the convergence policy (ALLOW on all-required-PASS, BLOCK on any required FAIL)
- **Given** the multi-review-gate fires on a code-edit turn,
- **When** it reads the most recent `ConvergenceResult`,
- **Then** it emits ALLOW when all required backends PASS, BLOCK when any required backend FAILs (the disagreement-among-required case produces `overall_verdict = FAIL` per AC-AMM-008 and BLOCKs conservatively), and disagreement among advisory-only backends NEVER BLOCKs (AC-AMM-009). (REQ-AMM-014)

**AC-AMM-021** (MUST) — gate fail-open to claude when all non-Claude backends missing
- **Given** all non-Claude backends are missing or unauthenticated at gate-fire time,
- **When** the gate evaluates,
- **Then** it falls back to claude-only evaluation (ALLOW if the in-session claude verdict is PASS; BLOCK only if it is FAIL) — the autonomous flow is NEVER hard-blocked on a missing optional dependency. (REQ-AMM-015, C2)

### M6 — Skill authoring + Template-First (Path A skill body)

**AC-AMM-022** (MUST) — Skill authored under template source + `make build` + §25 neutrality
- **Given** the `moai-ref-cross-model-audit` Skill,
- **When** it is staged,
- **Then** it exists under `internal/template/templates/.claude/skills/moai-ref-cross-model-audit/`, is mirrored to the local `.claude/skills/` tree, is regenerated via `make build`, and passes §25 template-neutrality (no SPEC-ID, no commit SHA, no internal date, no macOS-bias path) verified by the CI guard (`template-neutrality-check.yaml` / `internal_content_leak_test.go`). (REQ-AMM-016, C9)

**AC-AMM-023** (MUST) — Skill uses canonical MCP tool reference + independence rule verbatim
- **Given** the `moai-ref-cross-model-audit` Skill body,
- **When** it is inspected,
- **Then** it references the MCP tool via the canonical `mcp__moai__audit_multi` name (so a future tool-rename propagates via grep) and states the independence-preservation rule (pass only the synthesized `claude_verdict`, not the full analysis) verbatim. (REQ-AMM-017)

### Cross-cutting

**AC-AMM-024** (MUST) — subagent boundary (structured result, no `AskUserQuestion`)
- **Given** any convergence-engine or `audit_multi` handler facing a missing-input or inconclusive condition,
- **When** it cannot proceed,
- **Then** it returns a structured `ConvergenceResult` (including `VerdictInconclusive` per backend) and NEVER calls `AskUserQuestion` or emits a free-form user question (verified by grep: no `AskUserQuestion` / `mcp__askuser` reference in the convergence-engine package). (REQ-AMM-018, C5)

**AC-AMM-025** (MUST) — hardcoding prevention (env/threshold constants + config-block reuse)
- **Given** the new env-var names and thresholds introduced by this SPEC,
- **When** the code is inspected,
- **Then** env-var names are constants in `internal/config/envkeys.go`, thresholds/defaults live in `internal/config/defaults.go`, and the `multi_review_gate` config block reuses the existing `workflow.codex.review_gate` structural pattern (a sibling `multi_review_gate` key under `workflow:` — no new schema shape). (REQ-AMM-019)

## Edge cases (negative tests, MUST)

- **EC-1**: codex binary absent AND GLM key absent → `ConvergenceResult.overall_verdict` falls back to the in-session claude verdict; `per_backend_verdicts[]` carries `VerdictInconclusive` for codex and glm; no hard block (AC-AMM-004, AC-AMM-021).
- **EC-2**: codex returns `FAIL`, GLM returns `PASS`, claude returns `PASS` (all three required) → `overall_verdict = FAIL`, `disagreement_flag = true`, `residual_risk_note` describes the codex-vs-claude+glm split; the gate BLOCKs conservatively (AC-AMM-008, AC-AMM-020).
- **EC-3**: codex returns `FAIL` (advisory gate), claude returns `PASS` (required), GLM `off` → `overall_verdict = PASS`, `disagreement_flag = true`, the advisory `FAIL` is recorded, the flow does NOT block (AC-AMM-009, AC-AMM-010).
- **EC-4**: all required backends `VerdictInconclusive` (both missing) → `overall_verdict` follows the in-session claude verdict (fail-open to claude — AC-AMM-021).
- **EC-5**: multi-review-gate fires on a no-edit turn → ALLOW immediately, no backend invoked (self-gate — AC-AMM-019).
- **EC-6**: a future edit passes `claude_verdict` into the codex prompt context → a run-phase test fails (independence regression — AC-AMM-003, R1).
- **EC-7**: a future edit invents a `VerdictDisagreement` enum value → a run-phase test fails (REQ-AMM-008 — disagreement is a boolean flag, not an enum value — AC-AMM-011).

## §D.7 Closure gates (Definition of Done)

- All MUST ACs (AC-AMM-001..025) PASS with attributed evidence (command + verbatim output) in `progress.md` §E.2/§E.3.
- **Full suite green**: `go test ./...` + `go vet ./...` + §25 CI guard green (full suite, not affected-package-only — per the trust-but-verify lesson that affected-package-only self-report misses cross-cutting failures).
- **Sentinel-flip same-commit invariant**: the `multiConvergenceImplemented` sentinel flip (`internal/cli/mcp_audit.go:31`, `false → true`) and the convergence engine land in the SAME commit (R3) — no window where the sentinel lies. (This is a process gate on M1, not a behavioral AC; the traceability annotation below records the orphan resolution.)
- No `[NEEDS CLARIFICATION]` markers remain in `plan.md` / `research.md`.
- The `moai-ref-cross-model-audit` Skill is mirrored to template source + CI guard + `make build` in one change (AC-AMM-022).
- The disagreement policy (advisory, NOT block — REQ-AMM-006 #3, REQ-AMM-007, C3) is enforced by a regression test (EC-3, AC-AMM-009).

## Traceability matrix

| REQ | AC | Milestone |
|-----|----|-----------|
| REQ-AMM-001 | AC-AMM-001 | M0 |
| REQ-AMM-002 | AC-AMM-002 | M1 |
| REQ-AMM-003 | AC-AMM-003 | M1 |
| REQ-AMM-004 | AC-AMM-004 | M1 |
| REQ-AMM-005 | AC-AMM-005 | M1 |
| REQ-AMM-006 | AC-AMM-006, AC-AMM-007, AC-AMM-008, AC-AMM-009 | M2 |
| REQ-AMM-007 | AC-AMM-010 | M2 |
| REQ-AMM-008 | AC-AMM-011 | M2 |
| REQ-AMM-009 | AC-AMM-012 | M3 |
| REQ-AMM-010 | AC-AMM-013, AC-AMM-014 | M3 |
| REQ-AMM-011 | AC-AMM-015, AC-AMM-016 | M4 |
| REQ-AMM-012 | AC-AMM-017 | M4 |
| REQ-AMM-013 | AC-AMM-018, AC-AMM-019 | M5 |
| REQ-AMM-014 | AC-AMM-020 | M5 |
| REQ-AMM-015 | AC-AMM-021 | M5 |
| REQ-AMM-016 | AC-AMM-022 | M6 |
| REQ-AMM-017 | AC-AMM-023 | M6 |
| REQ-AMM-018 | AC-AMM-024 | cross |
| REQ-AMM-019 | AC-AMM-025 | cross |

> **AC-AMM-001 — process/milestone gate (traceability annotation).** AC-AMM-001 ("convergence data model locked at M0") is intentionally outside the REQ→AC behavioral traceability set: it verifies that the design commitments (ConvergenceResult shape, convergence algorithm, independence contract, disagreement policy, sentinel-flip rule) land at M0 — it is not itself a behavioral requirement. The sentinel-flip same-commit invariant (formerly a numbered AC) lives in §D.7 Closure gates as a process gate on M1. Recorded here to resolve the traceability orphan rather than masquerading as a behavioral AC.
