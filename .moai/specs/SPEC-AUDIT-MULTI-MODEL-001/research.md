# Research — SPEC-AUDIT-MULTI-MODEL-001

> Codebase integration verification + literature grounding. Every integration-point file:line reference the design report (`.moai/reports/moai-autonomy-workflow-redesign-20260803.html`, dated 2026-08-03) and the predecessor SPEC-MOAI-MCP-SERVER-001 `research.md` (verified at `b57de3ab1`, 2026-08-05) rely on was re-verified against the CURRENT worktree at `f85ff4c3e` (branch `feat/spec-audit-multi-model`, 2026-08-06). Drifts are recorded explicitly so no requirement rests on a stale location. This is the `research.md` sibling required when a SPEC touches existing code (skill § Red Flags).

## §1. Verification method

Each integration point below was verified by an executed `grep` against the worktree at `f85ff4c3e`. The commands were batched (agent-common-protocol § Parallel Execution) and their verbatim output drove this table. Lines cited as `file:line` reflect the CURRENT tree; where the predecessor SPEC's citation differs, the drift is named. Run-phase MUST re-confirm these before forking code.

## §2. Literature grounding — super-review + cross-model adversarial review

The design rationale rests on three convergent research findings cited in the autonomy report (§2.4 + §2.5):

### 2.4 Super-review pattern (Drew Hyde [R3])

> **Citation**: Drew Hyde, "Multi-Model Code Review via MCP" — `https://drewhyde.io/blog/codex-mcp-claude-code` (report reference [R3]).

**Core finding**: Claude 8-dimension primary review → Codex independent second opinion (Claude analysis NOT shared with Codex, to preserve independence) → orchestrator synthesis. The pattern works because different model families are trained on different corpora / objectives / by different teams → their failure modes are uncorrelated → they are genuinely complementary, not redundant. Report's framing: "Codex is not smarter, it is differently smart."

**MoAI application (this SPEC)**: the `audit_multi` MCP tool operationalizes this pattern. The `claude_verdict` input is the primary review; codex + glm are the independent second opinions; the convergence engine (design.md §3) is the orchestrator synthesis. Independence is preserved mechanically (design.md §5) — `claude_verdict` is consumed ONLY by the synthesis step, never as prompt context for codex/glm.

### 2.5 AgentOrchestra peer verification (arXiv:2506.12508 [R5])

> **Citation**: AgentOrchestra, arXiv:2506.12508 — `https://arxiv.org/html/2506.12508v1` (report reference [R5]).

**Core finding**: hierarchical planner + peer workers cross-verifying each other's output + memory-based dynamic re-planning. GAIA 82.42% SOTA. The peer cross-verification is structurally stronger than author self-report — a second worker inspecting the same artifact catches what the author missed.

**MoAI application (this SPEC)**: the convergence engine is the peer-verification step. The in-session claude verdict is the author's self-report; codex/glm are the peer workers. The `disagreement_flag` is the peer-disagreement signal; the `residual_risk_note` is what a human (or the orchestrator) reads to adjudicate.

### Cross-model adversarial review (Frontiers 10.3389/fcomp.2025.1655469)

> **Citation**: "Cross-model adversarial code review" — Frontiers in Computer Science 10.3389/fcomp.2025.1655469 (cited in report §2.5).

**Core finding**: GPT-4 self-code-repair 85.5% vs other-LLM code 77.4%. Hybrid LLM + static-analysis achieves +30% over LLM-only. The mechanical tools (lint, test, LSP) are first-class inputs — LLM alone is insufficient.

**MoAI application (this SPEC)**: the convergence engine treats the per-backend verdicts as LLM signals layered on top of the existing mechanical quality gates (`go test`, `golangci-lint`, `go vet`). The MoAI quality-gate pipeline is NOT replaced by the convergence — the convergence is an additional signal, and the mechanical gates remain authoritative for the Trackable / Tested dimensions of TRUST 5.

## §3. Integration-point inventory (verified at `f85ff4c3e`)

| # | Integration point | Predecessor SPEC citation | Verified current location | Status |
|---|---|---|---|---|
| 1 | deferral sentinel (the one-line flip) | `internal/cli/mcp_audit.go:31` `multiConvergenceImplemented = false` | `internal/cli/mcp_audit.go:31` `const multiConvergenceImplemented = false` | MATCH — this is the sentinel this SPEC flips to `true` at M1 (Definition of Done sentinel-flip same-commit invariant) |
| 2 | `activeAuditBackend` (validates multi token) | `mcp_audit.go:52` | `internal/cli/mcp_audit.go:52` `func activeAuditBackend(model string) (string, error)` — accepts `config.AuditModelMulti` (line 54) | MATCH — the engine consumes this validator |
| 3 | `buildAuditEnvBlock` (the `${VAR}` literal producer — already handles `multi`) | `mcp_audit.go:73` | `internal/cli/mcp_audit.go:73` — returns BOTH `glmAPIKeyEnvLiteral` and `codexAPIKeyEnvLiteral` for `AuditModelMulti` (lines 79-83) | MATCH — NO new env surface needed; C7 satisfied |
| 4 | `VerdictInconclusive` constant | `mcp_codex.go` | `internal/cli/mcp_codex.go` — const declaration at `:65` (`const VerdictInconclusive = "inconclusive"`) and assigned to a `Verdict` field at `:93`; constant value `"inconclusive"` (verified by test at `mcp_codex_test.go:113-114`) | MATCH — reused by the convergence engine for fail-open (REQ-AMM-004) |
| 5 | `codex_audit` handler | `mcp_codex.go` | `internal/cli/mcp_codex.go` (13.5KB) — the codex binary shellout handler | MATCH — the convergence engine CALLS it, does not fork it (C1) |
| 6 | GLM audit handler | `mcp_glm.go` | `internal/cli/mcp_glm.go` (12.2KB) — the z.ai direct API call handler | MATCH — same |
| 7 | `ResolveAgentModelEffort` SSOT | `internal/template/profile_matrix.go:385` | `internal/template/profile_matrix.go` `func ResolveAgentModelEffort(cfg config.LLMConfig, agent string) (me config.ModelEffort, mapped bool)` at `:385` | MATCH — the engine resolves model/effort via this sole interpreter (REQ-AMM-005, C6) |
| 8 | `AuditModel*` / `AuditGate*` enum constants | `internal/config/defaults.go:637-641` | `internal/config/defaults.go:637` default `AuditConfig{Model: AuditModelClaude, Gates: {Claude: Required, Codex: Required, GLM: Advisory}}`; enum values verified by `internal/config/mcp_audit_config_test.go:15-37` | MATCH — the engine reads this config; NO new enum (REQ-AMM-008) |
| 9 | codex-review-gate Stop hook (the pattern to extend) | `internal/cli/codex_review_gate.go` | `internal/cli/codex_review_gate.go` (the PURE gate logic) + `codex_review_gate_test.go` (the `withChangeDetector` seam at `:26`) | MATCH — the multi-review-gate REUSES the `withChangeDetector` seam (M5) |
| 10 | review-gate subcommand registration | `internal/cli/hook_e2e_test.go:374` | `internal/cli/hook_e2e_test.go:374` `"codex-review-gate": true` — the hook-command-tree registration pattern | MATCH — the `multi-review-gate` subcommand registers as a sibling |
| 11 | plan-auditor `tools:` carries `Skill` | `.claude/agents/moai/plan-auditor.md:7` | `.claude/agents/moai/plan-auditor.md:7` `tools: Read, Grep, Glob, Bash, Write, Edit, TaskCreate, TaskUpdate, TaskList, TaskGet, Skill` | MATCH — Path A wiring is already possible; the skill body instructs the call (M4) |
| 12 | sync-auditor `tools:` carries `Skill` | `.claude/agents/moai/sync-auditor.md:9` | `.claude/agents/moai/sync-auditor.md:9` `tools: Read, Grep, Glob, Bash, TaskCreate, TaskUpdate, TaskList, TaskGet, Skill` | MATCH — same |
| 13 | `errgroup` in go.mod (for parallel fan-out) | (implied stdlib + golang.org/x/sync) | go.mod carries `golang.org/x/sync` (errgroup's package) — already used by `internal/cli/` and `internal/web/` callers | MATCH — NO new dependency (M1 uses errgroup) |
| 14 | `mark3labs/mcp-go` SDK | `go.mod` (landed by predecessor) | go.mod carries `github.com/mark3labs/mcp-go` (landed by SPEC-MOAI-MCP-SERVER-001) | MATCH — `audit_multi` tool registration reuses the existing SDK pattern |
| 15 | MCP server tool-registration block | `internal/cli/mcp_server.go` | `internal/cli/mcp_server.go` (23KB) — the tool-registration block is where `audit_multi` is added (additive, M3) | MATCH — the only edit to `mcp_server.go` is REGISTERING the new tool |
| 16 | `IsGLMBackend` (GLM-sensitivity detection) | `internal/template/glm_effort_overlay.go:189` | `internal/template/glm_effort_overlay.go:189` `func IsGLMBackend(cfg config.LLMConfig) bool` | MATCH — relevant for GLM-aware convergence behavior (M1 may consult it) |
| 17 | env-var constants target | `internal/config/envkeys.go` | `internal/config/envkeys.go` exists (47 constant lines per predecessor research) | MATCH — target for any new env names (M7, REQ-AMM-019) |
| 18 | thresholds/defaults target | `internal/config/defaults.go` | `internal/config/defaults.go` exists | MATCH — target for any new thresholds (M7, REQ-AMM-019) |

## §4. Environment facts

- **Module**: `github.com/modu-ai/moai-adk`; **Go**: `1.26.4` (`go.mod`).
- **Predecessor SPEC state**: SPEC-MOAI-MCP-SERVER-001 is `status: completed` (PR #1378 merged); its 24/24 AC PASS; its `multiConvergenceImplemented = false` sentinel at `mcp_audit.go:31` is the deferral marker this SPEC flips.
- **No NEW go.mod dependencies**: `mark3labs/mcp-go` (MCP SDK) and `golang.org/x/sync` (errgroup) are both already in go.mod. This SPEC adds zero new dependencies.
- **Existing review-gate infrastructure**: the codex-review-gate Stop hook (opt-in, BranchGuard pattern, 900 s timeout, self-gate via `withChangeDetector`, ALLOW/BLOCK contract) is the proven pattern the multi-review-gate extends. The `hook_e2e_test.go:374` registry already accounts for the codex-review-gate; the multi-review-gate adds a sibling entry (M5).
- **Skill-routing infrastructure**: both auditor agents (`plan-auditor`, `sync-auditor`) carry `Skill` in their `tools:` — the existing `skill-routing.md` protocol routes the new `moai-ref-cross-model-audit` skill into them with NO change to the routing mechanism (M4).

## §5. Boundary against predecessor — what is NOT re-implemented

This SPEC consumes the following surfaces from SPEC-MOAI-MCP-SERVER-001 verbatim (the only edit to `mcp_audit.go` is the sentinel flip):

- The stdio JSON-RPC server (`mcp_server.go`).
- The `codex_audit` binary shellout handler (`mcp_codex.go`).
- The GLM direct z.ai API call handler (`mcp_glm.go`).
- The `audit_model ∈ {claude, codex, glm, multi}` and `audit_gate ∈ {off, advisory, required}` enums (`internal/config/`).
- The `VerdictInconclusive → claude` fail-open (C2 of predecessor — this SPEC's C2 inherits it).
- The `ResolveAgentModelEffort` SSOT (`profile_matrix.go:385`).
- The codex-review-gate Stop-hook pattern (`codex_review_gate.go`).
- The `buildAuditEnvBlock(config.AuditModelMulti)` two-literal block (already produces both `${GLM_API_KEY}` and `${CODEX_API_KEY}` — NO new env surface).

The boundary is explicit (spec.md §C quotes AC-MCP-012 and AC-MCP-017 verbatim). A run-phase audit verifies that no single-backend handler is forked into the convergence engine — the engine CALLS the handlers, it does not re-implement them (AC-AMM-013, AP-AMM-1).

## §6. Fail-open identity verification

The fail-open identity (C2) is inherited from SPEC-MOAI-MCP-SERVER-001 and extends naturally to the multi-backend case:

- **Single-backend fail-open (predecessor)**: a missing codex returns `VerdictInconclusive`, falls back to claude.
- **Multi-backend fail-open (this SPEC)**: a missing codex OR glm returns `VerdictInconclusive` for that backend, convergence continues over the remaining active backends. If ALL non-Claude backends are missing, convergence falls back to the in-session claude verdict (REQ-AMM-015, AC-AMM-021).

The convergence algorithm (design.md §3) handles this explicitly: the "all required inconclusive" branch falls back to `claude_v.verdict`. The fail-open identity is preserved across the multi-backend composition — a missing optional backend NEVER hard-blocks the workflow.

## §7. Cross-references

- Design report: `.moai/reports/moai-autonomy-workflow-redesign-20260803.html` §2.4 (super-review pattern [R3]), §2.5 (AgentOrchestra [R5] + cross-model adversarial review Frontiers citation), §3.4 (Codex 감사 위임), §3.6 v3 extension, Q1 routing paths.
- spec.md §C (verbatim deferral quote), §E (constraints C1-C9), §G (risks R1-R5).
- plan.md §B (drift catalogue — the sentinel location), §F (milestones), §G (anti-patterns AP-AMM-1..10).
- design.md §1 (ConvergenceResult data model), §2 (parallel-execution model), §3 (convergence algorithm), §5 (independence mechanism), §6 (Stop-hook extension).
- acceptance.md — AC-AMM-001..027 + EC-1..EC-7 + Definition of Done.
- Predecessor research: `.moai/specs/SPEC-MOAI-MCP-SERVER-001/research.md` (verified at `b57de3ab1`, 2026-08-05 — the 20-point integration inventory this SPEC consumes).
- Schema SSOT: `.claude/rules/moai/development/spec-frontmatter-schema.md` (frontmatter), `internal/spec/lint.go` `specIDPattern` (SPEC ID regex).

> No open clarifications — the two fixed user decisions (Tier L single SPEC; disagreement = advisory NOT block) settle the scope and policy questions. The two DQs in `design.md` §7 (state-file write vs re-invoke; missing claude_verdict handling) are M0 design-lock decisions, not user-input blockers.
