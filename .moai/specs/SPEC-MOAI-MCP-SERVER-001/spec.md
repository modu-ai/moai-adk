---
id: SPEC-MOAI-MCP-SERVER-001
title: "moai self-hosted MCP server — thin stdio JSON-RPC wrapper over the internal/ core (3-way audit backends + status/trend tools)"
version: "0.1.0"
status: in-progress
created: 2026-08-05
updated: 2026-08-05
author: manager-spec
priority: P2
phase: "v3.2 target"
module: internal/cli/mcp_server
lifecycle: spec-anchored
tier: L
tags: "mcp, json-rpc, audit, codex, glm, template-first, fail-open, ssot"
---

# SPEC-MOAI-MCP-SERVER-001 — moai self-hosted MCP server (autonomy·speed Epic, P2 foundation)

## HISTORY

- 2026-08-05 (plan-phase, iter-1) — Initial Tier L authoring. Operationalizes §3.4 (Codex 감사 위임 — moai-mcp 내장) + §3.6 (moai 자체 MCP 서버) of `.moai/reports/moai-autonomy-workflow-redesign-20260803.html`. This SPEC is the **foundation**: a thin stdio JSON-RPC MCP server (`moai mcp-server`) wrapping the SAME `internal/` core functions the CLI Cobra handlers call. It **unblocks** the future SPEC-AUDIT-MULTI-MODEL (multi-model convergence orchestration); convergence logic itself is explicitly out of scope. Design source: report §3.4 + §3.6 (IN scope); §3.5 (audit speed optimizations A1-A11), §3.7 (trend MCP tools) are OUT of scope.

## §A. User Story

**As a** MoAI-ADK user (or an external MCP-capable client — Cursor, Cline, Zed, another LLM host) who wants an AI orchestrator to inspect moai state and run audits **declaratively and repeatedly**,

**I want** a single, long-running `moai mcp-server` process that exposes moai's existing capabilities as typed JSON-RPC tools (with JSON Schema, so the AI never guesses flags), and that exposes the 3-way audit backends (Claude / Codex / GLM) behind a uniform `audit_model` + per-auditor `audit_gate` contract with mandatory fail-open,

**so that** (a) an AI orchestrator can call `session_list` / `goal_status` / `spec_audit` / `verify_snapshot` etc. without paying a Go-binary cold-start per invocation or guessing `--help` output; (b) cross-model audit (Claude + Codex + GLM) is reachable from BOTH Claude and GLM host sessions through the identical local stdio server; and (c) the audit attack surface narrows to a fixed, declared tool set — never arbitrary `moai <subcommand> --flag` combinations — which is the safety net the fully-autonomous tier needs.

## §B. Scope Summary

This SPEC delivers the **foundation** only — the server scaffold, the core read/status/trend tools, the 3-way audit backend plumbing (single-backend selection + per-auditor gate), the auth/secret/fail-open contract, and the Template-First reversal that provisions exactly one neutral `.mcp.json` entry. The multi-model **convergence orchestration** (running Claude + Codex + GLM in parallel and synthesizing disagreements) is a separate future SPEC (SPEC-AUDIT-MULTI-MODEL) that this foundation unblocks.

The server is a **thin wrapper**: every tool handler calls an existing `internal/` function that the CLI Cobra handlers already call. No new core logic is authored behind the MCP surface — the two surfaces (CLI + MCP) share one source of truth.

### Out of Scope — multi-model convergence orchestration

- The `audit_model: multi` / `cross-model` **convergence LOGIC** (parallel fan-out across Claude + Codex + GLM, disagreement synthesis, super-review pattern) → **future SPEC-AUDIT-MULTI-MODEL** (not yet authored; this SPEC unblocks it). This SPEC delivers only the `audit_model` config field, the single-backend selection, and the per-auditor `audit_gate`; the `multi` value is a declared-but-not-orchestrated enum token here.

### Out of Scope — audit speed optimizations (A1-A11)

- Report §3.5 audit optimizations (sticky hash cache, skip-threshold alignment, 4dim binding promotion, shared diagnostic snapshot, docs ∥ audit parallelization, etc.) → **separate SPEC**. This SPEC may reuse `runtime.AuditCache` read-only where it already exists, but does not change the audit-caching or skip policy.

### Out of Scope — Codex Phase 2 tools

- `codex_task` (rescue), `codex_job_status/result/cancel`, `codex_transfer`, and a native Go JSON-RPC client for `codex app-server` → **follow-up SPEC**. This SPEC delivers only the Phase 1 codex surface: `codex_audit` (native + adversarial) + `codex_setup` + the Stop-hook review gate, all via `codex` binary shellout.

### Out of Scope — third-party / trend MCP bundling

- Report §3.7 trend MCP tools (Playwright, ast-grep, Semgrep, GitHub, Postgres/Sentry bundling and opt-in recipes) → **separate future SPEC (SPEC-TREND-MCP)**. This SPEC provisions exactly ONE local server — `moai mcp-server` — and touches no third-party MCP entries.

### Out of Scope — web live WebSocket dashboard

- The real-time `moai web` board (WebSocket + fsnotify over `.moai/state/`) → **v3.1 deployment target, separate future SPEC (SPEC-WEB-LIVE-BOARD)**. The MCP server and the web board share the same `internal/` read functions; the web board does NOT require the MCP server and vice versa.

## §C. Requirements (GEARS)

> Domain prefix `REQ-MCP-NNN` maps to SPEC domain `MOAI-MCP-SERVER`. GEARS compound clauses are used throughout. Every "wraps existing `internal/` function" requirement is backed by a verified file:line citation in `research.md`; the report's 2026-08-03 references were re-verified against tree `b57de3ab1` and the drifts are recorded there.

### M1 — Server scaffold + core read/status tools + `.mcp.json` provisioning

**REQ-MCP-001** (Ubiquitous) The `moai mcp-server` subcommand shall expose a long-running stdio JSON-RPC MCP server built on the `mark3labs/mcp-go` SDK, attached via `root.go` `AddCommand` and modeled on the `goal.go` blocking-`RunE` stdio pattern.

**REQ-MCP-002** (Capability gate) **Where** the user has not opted in, the `moai mcp-server` server and its `.mcp.json` provisioning shall remain OFF by default (opt-in), provisioning exactly one neutral local entry `{"mcpServers":{"moai":{"command":"moai","args":["mcp-server"]}}}` and no third-party entries.

**REQ-MCP-003** (Ubiquitous) Each MCP tool handler shall be a thin wrapper that calls the SAME `internal/` function the corresponding CLI Cobra handler calls, so the CLI and MCP surfaces share one source of truth and can never diverge.

**REQ-MCP-004** (Event-driven) **When** the MCP host requests `tools/list`, the server shall declare each tool's name and JSON Schema (input parameters + structured result shape), so an AI caller invokes tools type-safely with zero flag-guessing.

**REQ-MCP-005** (Ubiquitous) The core read/status/trend tools shall wrap the verified integration points: `session_list`→`session.QueryActiveWork`, `goal_status`→`goal.LoadGoal`, `goal_arm`→goal arming (`NewGoal`+`SaveGoal`, the `cli/goal.go` `arm` RunE logic), `spec_progress`→the SPEC scanner, `verify_snapshot`/`verify_trend`→`verify.Load`+`verify.RecordCheck` (first CLI/MCP surface for verify), `spec_audit`/`spec_drift`→`spec.Audit`, and `audit_cache`→`runtime.AuditCache` (`ComputeHash`/`Lookup`/`Store`).

### M2 — Codex audit backend (Phase 1 tools) + Stop-hook gate

**REQ-MCP-006** (Event-driven) **When** the `codex_audit` tool is invoked, the server shall shell out to the `codex` binary (audit intelligence = the codex model), unifying the native and adversarial review modes via a `mode` enum (native → `review/start`, adversarial → `turn/start` + adversarial-review prompt), and shall return a result shaped by `review-output.schema.json` (`verdict` / `summary` / `findings[severity,title,body,file,line,confidence,recommendation]` / `next_steps`).

**REQ-MCP-007** (Capability gate) **Where** the user runs `codex_setup`, the server shall perform a Go reimplementation of the setup probe (`exec.LookPath("codex")` + `codex --version`), classify the auth provider (ChatGPT / apiKey / provider), and expose the `enable_review_gate` toggle — without relying on any Node.js / `.mjs` bridge.

**REQ-MCP-008** (State-driven) **While** `workflow.codex.review_gate.enabled` is set (opt-in, BranchGuard pattern), the `moai hook codex-review-gate` Stop hook shall enforce an ALLOW/BLOCK contract with a mandatory self-gate — "the previous turn produced no code edit / is a status report / is a review-result ⇒ ALLOW immediately" — to prevent false blocks, with a 900 s timeout override (the moai-default 5 s does not apply to this hook).

### M3 — GLM audit backend + 3-way config + auth + fail-open

**REQ-MCP-009** (Event-driven) **When** the GLM audit backend is selected, the server shall call the z.ai GLM API directly (the Anthropic-compatible `https://api.z.ai/api/anthropic` endpoint, reusing the existing z.ai client plumbing) to submit the audit prompt — NOT the z.ai MCP server, and NOT through any z.ai gateway.

**REQ-MCP-010** (Capability gate) **Where** the user configures audit, the `audit_model` config field shall accept `{claude, codex, glm, multi}` and a per-auditor `audit_gate` shall accept `{off, advisory, required}` (default `required`), with the default profile = claude + codex required, glm advisory (user-enabled). The `multi` convergence orchestration is a declared token only here; its logic is owned by the future SPEC-AUDIT-MULTI-MODEL.

**REQ-MCP-011** (Unwanted) The server shall NOT serialize any codex or GLM secret into a git-tracked `.mcp.json`; codex auth uses OAuth material in `~/.moai/.env.codex` and GLM auth uses the API key in `~/.moai/.env.glm` (both via the `loadGLMKey` pattern), and any env value that must reach the host runtime is expressed as a `${CODEX_API_KEY}` / `${GLM_API_KEY}` literal expanded by the Claude Code runtime.

**REQ-MCP-012** (Event-driven) **When** a selected non-Claude auditor is missing or unauthenticated, the server shall return `VerdictInconclusive` for that auditor and fall back to the active auditor (claude); the workflow MUST NOT hard-block on a missing optional dependency (fail-open mandatory).

**REQ-MCP-013** (State-driven) **While** resolving audit model/effort, the `codex_audit` and GLM-audit tools SHALL resolve model and effort through `template.ResolveAgentModelEffort` (verified definition at `internal/template/profile_matrix.go:385`) — the identical interpreter used by the web console and the launcher — and SHALL NOT read agent frontmatter or `llm.agent_overrides` directly (fork risk).

**REQ-MCP-014** (Unwanted) MCP tool handlers shall NOT invoke `AskUserQuestion` or emit free-form user-facing questions (subagent boundary); on a missing-input or inconclusive condition the tool returns a structured result (including `VerdictInconclusive`) and the orchestrator translates it through its own `AskUserQuestion` channel.

### M4 — init/web selection UI + Template-First reversal

**REQ-MCP-015** (Event-driven) **When** the user runs `moai init` or uses `moai web`, the wizard (`internal/cli/wizard/questions.go` `Page3Questions`) and the web console shall surface `audit_model` selection, per-auditor `audit_gate` selection, a `codex_audit_enabled` flag, and an `mcp_tools_opt_in` flag, applied through `applyWizardPage3ToOpts` (`internal/cli/init.go`) and the web handler's identical interpreter.

**REQ-MCP-016** (Ubiquitous) The Template-First reversal shall reverse the `.claude/rules/moai/core/settings-management.md` (line 33) statement "MoAI-ADK no longer ships or provisions MCP servers via `.mcp.json`" to provision exactly ONE local server (`moai mcp-server`), and the reversal SHALL be made generically/neutral (no SPEC-ID, no commit SHA, no internal date, no macOS-bias path) so it passes §25 template-neutrality, updating the template source `internal/template/templates/` AND the CI guard (`template-neutrality-check.yaml` / `internal_content_leak_test.go`) in the SAME change with `make build`.

### Cross-cutting

**REQ-MCP-017** (Capability gate) **Where** any new env-var name, threshold, or default is introduced by this SPEC, it SHALL be defined as a constant in `internal/config/envkeys.go` (env-var names) or `internal/config/defaults.go` (thresholds/defaults), and the `.mcp.json` entry + `mcp-server` command SHALL stay generic — no hardcoded absolute paths, model names, or org identifiers in distributed surfaces (CLAUDE.local.md §14 hardcoding prevention).

**REQ-MCP-018** (Capability gate) **Where** any `.claude/` template change is made, it SHALL be mirrored to `internal/template/templates/` and regenerated via `make build` (CLAUDE.local.md §2 Template-First), with §25 neutrality enforced by the CI guard.

## §D. Acceptance Criteria (summary — full Given-When-Then in acceptance.md)

Each requirement maps to one or more binary-testable acceptance criteria (AC-MCP-NNN) enumerated in `acceptance.md`. The acceptance matrix binds: M1 ↔ AC-MCP-001..006, M2 ↔ AC-MCP-007..010, M3 ↔ AC-MCP-011..017, M4 ↔ AC-MCP-018..021, cross-cutting ↔ AC-MCP-022..024. Severity classification, traceability, and Definition of Done live in `acceptance.md`.

## §E. Constraints (non-functional)

- **C1 — Same-core-two-surfaces**: the CLI and the MCP server call identical `internal/` functions; no core logic is forked into the MCP layer.
- **C2 — fail-open mandatory**: a missing or unauthenticated optional auditor (codex, glm) NEVER hard-blocks; it returns `VerdictInconclusive` and falls back to claude. Evidence of absence is not evidence of failure in either direction.
- **C3 — secret hygiene**: codex/glm credentials are NEVER serialized into git-tracked `.mcp.json`; `${VAR}` literals are expanded by the host runtime.
- **C4 — SSOT for model/effort**: `template.ResolveAgentModelEffort` (`profile_matrix.go:385`) is the sole interpreter; agent frontmatter and `llm.agent_overrides` are never read directly by MCP handlers.
- **C5 — subagent boundary**: MCP tools return structured results (incl. `VerdictInconclusive`); they never call `AskUserQuestion`.
- **C6 — opt-in (default off)**: the server, its `.mcp.json` provisioning, and the codex review gate all ship inert to distributed users.
- **C7 — Template-First + §25 neutrality**: any template change is mirrored to `internal/template/templates/` + `make build` + CI guard, and is generic/neutral.
- **C8 — hardcoding prevention**: env names → `envkeys.go` constants; thresholds → `defaults.go`; no org/model/absolute-path in distributed surfaces.

## §F. Dependencies

- **go.mod**: adds `github.com/mark3labs/mcp-go` (approved single new dependency; no MCP SDK currently present). Go 1.26.4, module `github.com/modu-ai/moai-adk`.
- **Existing internal packages (read-mostly reuse)**: `internal/spec`, `internal/session`, `internal/goal`, `internal/verify`, `internal/runtime` (AuditCache), `internal/template` (ResolveAgentModelEffort, IsGLMBackend), `internal/config` (envkeys, defaults), `internal/cli` (glm_tools.go atomic-config helpers, goal.go pattern, wizard package).
- **External binaries (runtime, optional)**: `codex` CLI (for `codex_audit`; fail-open if absent).
- **Epic positioning (not hard technical deps)**: the autonomy·speed Epic sequences this SPEC after SPEC-AUTONOMY-TIERS-001 and SPEC-INFINITE-GOAL-001, and it unblocks a future SPEC-AUDIT-MULTI-MODEL. The foundation is self-contained — the MCP server functions without the autonomy-tier wiring; only the M4 init/web selection surface shares the autonomy-tier config namespace.

## §G. Risks

- **R1 — `codex mcp-server` is experimental** (upstream): the binary shellout surface and the `review/start`·`turn/start` RPC names may change. Mitigation: install-time tool-surface probe + fail-open (REQ-MCP-012) + the Phase-1-only scope.
- **R2 — Template-First reversal widens the distributed `.mcp.json` attack surface**: provisioning a local binary as an MCP server is a new template surface. Mitigation: exactly ONE neutral entry, opt-in default-off, no third-party entries, §25 neutrality CI guard.
- **R3 — SDK choice drift**: `mark3labs/mcp-go` is community-maintained; the official `modelcontextprotocol/go-sdk` is the maintained alternative. Mitigation: design.md records the rationale; the thin-wrapper architecture means the SDK is swappable without touching the `internal/` core. (User decision: mark3labs/mcp-go — not re-litigated.)
- **R4 — verify-before-claim on integration points**: the report's file:line references are from 2026-08-03. Several drifted (see research.md). Any requirement resting on a drifted reference is re-grounded at the verified current location before run-phase.

## §H. Cross-References

- Design source: `.moai/reports/moai-autonomy-workflow-redesign-20260803.html` §3.4 (Codex 감사 위임 — moai-mcp 내장) + §3.6 (moai 자체 MCP 서버).
- Epic memory: `project_autonomy_mcp_server_next.md` (autonomy·speed Epic, this SPEC = next P2).
- Sibling (unblocked by this): future SPEC-AUDIT-MULTI-MODEL — the `multi` convergence logic.
- Sibling (out of scope): future SPEC-TREND-MCP (§3.7), future SPEC-WEB-LIVE-BOARD (v3.1), the §3.5 audit-optimization SPEC.
- Schema SSOT: `.claude/rules/moai/development/spec-frontmatter-schema.md`.
