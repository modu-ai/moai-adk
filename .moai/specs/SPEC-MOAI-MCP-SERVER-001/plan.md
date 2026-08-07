# Plan — SPEC-MOAI-MCP-SERVER-001

> Implementation plan for the moai self-hosted MCP server (Tier L foundation). Milestones are ordered by **decision-reversibility** (the decisions most likely to change lead; mechanical/refactoring steps defer to the bottom). Priority labels only — no time estimates (CLAUDE.local.md §4, output-style §10 HARD). Per-milestone AC bindings point at `acceptance.md`.

## §A. Context

This SPEC operationalizes report §3.4 + §3.6: a thin stdio JSON-RPC MCP server (`moai mcp-server`) wrapping the SAME `internal/` core the CLI uses, plus the Phase-1 codex audit backend, the GLM audit backend, the 3-way `audit_model` + per-auditor `audit_gate` contract, mandatory fail-open, and the Template-First reversal that provisions exactly one neutral `.mcp.json` entry. It is the P2 foundation of the autonomy·speed Epic and unblocks a future SPEC-AUDIT-MULTI-MODEL (the `multi` convergence logic).

The architecture invariant is **same-core-two-surfaces**: every MCP tool handler calls an existing `internal/` function; no core logic is forked into the MCP layer. This keeps the SDK swappable (R3) and makes drift between the CLI and MCP surfaces impossible by construction.

## §B. Known Issues / Drift from the Design Report

The report's file:line references are 2026-08-03. The following drifted when re-verified against tree `b57de3ab1` (full evidence in `research.md`):

- `template.ResolveAgentModelEffort` — report cited `handlers.go:85`; the **definition** is at `internal/template/profile_matrix.go:385`. `internal/web/handlers.go:86` is a *caller* site ("G3-1 repoint"), not the definition. SSOT for REQ-MCP-013 = `profile_matrix.go:385`.
- `goal.LoadGoal` — report cited `state.go:43`; actual `internal/goal/state.go:57`. Goal arming has no single `Arm()` function — the CLI `arm` RunE (`internal/cli/goal.go:105`/`:110`) composes `goal.NewGoal` (`schema.go:115`) + `goal.SaveGoal` (`state.go:76`). `goal_arm` wraps that composition.
- wizard package — report cited `wizard/questions.go`; actual package path `internal/cli/wizard/` (`Page3Questions` ref at `internal/cli/wizard/questions.go:15`, `WizardResult` at `internal/cli/wizard/types.go:12`).
- `runtime.AuditCache` — report cited `audit_cache.go:43`; the interface is at `internal/runtime/audit_cache.go:50`, the `InMemoryCache` methods at `:110`/`:146`/`:157`.
- `buildMoaiMCPServerEntry` — does NOT exist (report phrasing implied reuse). It is a NEW function modeled on `buildZAIMCPEntry` (`internal/cli/glm_tools.go:391`); the atomic-config infrastructure (`mutateClaudeJSONAtomic:541`, `resolveConfigPath:362`) DOES exist and is reused.
- The `.mcp.json` reversal sentence is confirmed at `.claude/rules/moai/core/settings-management.md:33` verbatim.
- **INFINITE-GOAL dependency already landed**: `goal.Ceiling.MaxTurns == 0` infinite entry point is already implemented (`internal/goal/schema.go:36` comment + `evaluate.go:320` `> 0` guard). The epic's INFINITE-GOAL prerequisite is satisfied for this SPEC's purposes.

## §C. Pre-flight (run before M1)

1. Re-confirm every integration-point file:line in `research.md` against the tree at run-phase start (some may drift again before run).
2. Confirm `mark3labs/mcp-go` resolves and its server-creation API surface (`server.NewStdioServer` / tool registration / JSON Schema) matches the thin-handler assumption; pin the version in `go.mod`. If the API diverges materially, surface as a blocker before forking code.
3. Confirm the §25 CI guard (`template-neutrality-check.yaml`, `internal_content_leak_test.go`) current rule set so the M4 reversal is authored to pass it on the first try.
4. Run the SPEC ID regex check (`[[ "$ID" =~ ^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$ ]]`) — the canonical ID `SPEC-MOAI-MCP-SERVER-001` PASSES; the task's codename `SPEC-MOAI-MCP-SERVER` FAILS (no numeric tail) and MUST NOT be used as the frontmatter `id`.

## §D. Constraints (binding — see spec.md §E for the prose)

C1 same-core · C2 fail-open · C3 secret hygiene · C4 model/effort SSOT (`profile_matrix.go:385`) · C5 subagent boundary · C6 opt-in default-off · C7 Template-First + §25 neutrality · C8 hardcoding prevention.

## §E. Self-Verification (per-milestone, run-phase §E.2 evidence)

Each milestone's evidence package cites verbatim command + output for: the build (`go build ./...`), the affected-package tests, the MCP-server smoke (stdio `initialize` → `tools/list` → `tools/call` round-trip), and — where the milestone touches distributed surfaces — the §25 neutrality CI guard run. Fail-open and secret-hygiene ACs get a dedicated negative test (missing `codex` binary, `${VAR}` literal not resolved into git-tracked `.mcp.json`).

## §F. Milestones (ordered by decision-reversibility — highest-change-likelihood first)

### M0 — Design-decision stabilization (lead — most likely to change)

The decisions most likely to be re-opened are isolated here so human review focuses on them first:

- **SDK choice + version pin** (`mark3labs/mcp-go` vs official `modelcontextprotocol/go-sdk`). User decision recorded (mark3labs/mcp-go); M0 confirms the API surface and pins the version. Reversal cost = re-touching only the thin handler layer (the `internal/` core is untouched by design).
- **`audit_model` + `audit_gate` config schema** (field names, enum values, default profile = claude+codex required / glm advisory). This is a user-facing config surface shared with the autonomy tiers; lock the names here.
- **`review-output.schema.json` adoption** for `codex_audit` output (verdict/findings/next_steps). Confirm the schema shape is the contract both the MCP result and the orchestrator-translation layer expect.
- **The Template-First reversal wording** (the exact neutral sentence replacing settings-management.md:33 + the exact neutral `.mcp.json` entry shape).

AC binding: AC-MCP-022 (schema decisions documented + locked), AC-MCP-018 (reversal wording locked, neutral).

### M1 — Server scaffold + core read/status tools + `.mcp.json` provisioning

- New `internal/cli/mcp_server.go` + `root.go` `AddCommand` (`goal.go` blocking-`RunE` stdio pattern).
- `mark3labs/mcp-go` server registering thin tool handlers over the verified integration points (REQ-MCP-005).
- `.mcp.json` single-entry provisioning: reuse `glm_tools.go` `mutateClaudeJSONAtomic` + `resolveConfigPath`; NEW `buildMoaiMCPServerEntry` (modeled on `buildZAIMCPEntry` at `glm_tools.go:391`).
- `tools/list` JSON Schema declaration for every core tool (REQ-MCP-004).

AC binding: AC-MCP-001 (subcommand + stdio round-trip), AC-MCP-002 (opt-in default-off), AC-MCP-003 (thin-wrapper — same `internal/` function as CLI), AC-MCP-004 (tools/list JSON Schema), AC-MCP-005 (core tools wrap verified integration points), AC-MCP-006 (.mcp.json single neutral entry via atomic-config helpers).

### M2 — Codex audit backend (Phase 1 tools) + Stop-hook gate

- `codex_audit` tool: `codex` binary shellout, `mode` enum (native `review/start` / adversarial `turn/start` + adversarial-review prompt), `review-output.schema.json` output (REQ-MCP-006).
- `codex_setup` tool: Go reimplementation — `exec.LookPath("codex")` + `codex --version`, ChatGPT/apiKey/provider classification, `enable_review_gate` toggle (REQ-MCP-007). No `.mjs` bridge.
- Stop-hook gate: `moai hook codex-review-gate` (Go reimplementation of the `.sh` concept), opt-in `workflow.codex.review_gate.enabled` (BranchGuard pattern), ALLOW/BLOCK contract, mandatory self-gate (no-edit / status-report / review-result ⇒ ALLOW), 900 s timeout override (REQ-MCP-008).

AC binding: AC-MCP-007 (codex_audit unified modes + schema output), AC-MCP-008 (codex_setup Go probe, no Node bridge), AC-MCP-009 (review gate self-gate prevents false blocks), AC-MCP-010 (review gate opt-in + 900 s timeout).

### M3 — GLM audit backend + 3-way config + auth + fail-open

- GLM audit: z.ai GLM API direct call (`https://api.z.ai/api/anthropic`, Anthropic-compatible; reuse the existing z.ai client plumbing) — NOT the z.ai MCP server, NOT any z.ai gateway (REQ-MCP-009).
- 3-way `audit_model` ∈ {claude, codex, glm, multi} + per-auditor `audit_gate` ∈ {off, advisory, required} (default required; default profile claude+codex required, glm advisory). `multi` is a declared token only — convergence logic is SPEC-AUDIT-MULTI-MODEL (REQ-MCP-010).
- Auth branching: codex OAuth (`~/.moai/.env.codex`, `loadGLMKey` pattern), glm API key (`~/.moai/.env.glm`); `${CODEX_API_KEY}` / `${GLM_API_KEY}` literals expanded by the host runtime, NEVER serialized into git-tracked `.mcp.json` (REQ-MCP-011).
- Fail-open mandatory: missing/unauthenticated codex or glm ⇒ `VerdictInconclusive` ⇒ fall back to claude (REQ-MCP-012).
- Model/effort SSOT: `codex_audit` + GLM audit resolve via `template.ResolveAgentModelEffort` (`profile_matrix.go:385`) only (REQ-MCP-013).
- Subagent boundary: MCP tools return structured results (incl. `VerdictInconclusive`); never `AskUserQuestion` (REQ-MCP-014).

AC binding: AC-MCP-011 (GLM direct z.ai API call), AC-MCP-012 (audit_model + audit_gate enums + default profile), AC-MCP-013 (secret hygiene — `${VAR}` literal, no serialization), AC-MCP-014 (fail-open on missing/unauth), AC-MCP-015 (model/effort SSOT via ResolveAgentModelEffort), AC-MCP-016 (subagent boundary — structured result, no AskUserQuestion), AC-MCP-017 (3-way selection resolves a single active backend; `multi` token accepted but not orchestrated).

### M4 — init/web selection UI + Template-First reversal + docs

- `moai init` wizard: `audit_model` + per-auditor `audit_gate` + `codex_audit_enabled` + `mcp_tools_opt_in` in `Page3Questions` (`internal/cli/wizard/questions.go`); applied via `applyWizardPage3ToOpts` (`internal/cli/init.go`) (REQ-MCP-015).
- `moai web` console: same selection surface (identical interpreter).
- **Template-First reversal**: reverse settings-management.md:33 to provision exactly ONE local server; generic/neutral (no SPEC-ID/SHA/internal-date/macOS-path) → §25 neutrality; update template source `internal/template/templates/` AND the CI guard (`template-neutrality-check.yaml` / `internal_content_leak_test.go`) in the SAME change + `make build` (REQ-MCP-016, REQ-MCP-018).
- Hardcoding prevention: new env names → `internal/config/envkeys.go` constants; new thresholds → `internal/config/defaults.go`; `.mcp.json` entry + command stay generic (REQ-MCP-017).

AC binding: AC-MCP-018 (reversal applied, neutral), AC-MCP-019 (template source + CI guard + make build in one change), AC-MCP-020 (init wizard selection surface), AC-MCP-021 (web console selection surface), AC-MCP-023 (env/threshold constants — hardcoding prevention), AC-MCP-024 (full `go test ./...` + §25 CI guard green).

### Critical path

`M0 (decisions) → M1 (scaffold + core + .mcp.json) → M2 (codex backend + gate) → M3 (glm backend + 3-way + fail-open) → M4 (init/web UI + reversal + docs)`. M2 and M3 are largely independent backend work but both depend on M1's server scaffold and the `review-output.schema.json`/`audit_model` decisions locked in M0. M4 depends on M3 (the selection UI surfaces the 3-way config). Within the autonomy·speed Epic, this SPEC follows SPEC-AUTONOMY-TIERS-001 / SPEC-INFINITE-GOAL-001 and precedes SPEC-AUDIT-MULTI-MODEL.

### Gates

- **G-M0**: SDK API surface confirmed + version pinned + config schema names locked + reversal wording locked. (Blocks M1-M4.)
- **G-M1**: stdio `initialize` → `tools/list` → `tools/call` round-trip green; opt-in default-off verified; thin-wrapper parity with CLI verified for ≥1 core tool.
- **G-M2**: `codex_audit` returns schema-shaped output in both modes; review gate self-gate proven (no false block on a status-report turn); fail-open on missing `codex`.
- **G-M3**: GLM direct z.ai call returns a verdict; `${VAR}` literal NOT serialized into git-tracked `.mcp.json`; fail-open on missing glm key; model/effort resolved via `ResolveAgentModelEffort` (not frontmatter).
- **G-M4**: settings-management.md:33 reversed + neutral; template source + CI guard + `make build` in one change; init + web selection surfaces functional; full `go test ./...` + §25 CI guard green.

## §G. Anti-Patterns (do NOT)

- **AP-1 — Forking core logic into the MCP layer**: any MCP handler that re-implements (instead of calls) an `internal/` function violates C1. The handler's job is argument-mapping + result-shaping only.
- **AP-2 — Serializing secrets into `.mcp.json`**: writing `CODEX_API_KEY`/`GLM_API_KEY` values (not `${VAR}` literals) into the git-tracked config violates C3/REQ-MCP-011.
- **AP-3 — Hard-block on missing optional auditor**: a `codex`/`glm` miss that returns a hard error instead of `VerdictInconclusive` + fallback violates C2/REQ-MCP-012.
- **AP-4 — Reading agent frontmatter / `llm.agent_overrides` directly**: bypassing `ResolveAgentModelEffort` violates C4/REQ-MCP-013 (fork risk).
- **AP-5 — `AskUserQuestion` from an MCP tool**: violating the subagent boundary (C5/REQ-MCP-014).
- **AP-6 — Template change without mirror + `make build` + CI guard**: violating Template-First / §25 (C7/REQ-MCP-016/018).
- **AP-7 — Using the unnumbered codename `SPEC-MOAI-MCP-SERVER` as the frontmatter `id`**: fails the lint regex; the canonical ID is `SPEC-MOAI-MCP-SERVER-001`.
- **AP-8 — Hardcoding the `multi` convergence logic here**: that belongs to SPEC-AUDIT-MULTI-MODEL; this SPEC accepts `multi` as a token only.

## §H. Cross-References

- spec.md (this SPEC) — requirements, scope, non-goals.
- acceptance.md — AC-MCP-001..024 Given-When-Then + traceability + Definition of Done.
- research.md — integration-point file:line verification + drift notes + SDK rationale evidence.
- design.md — same-core-two-surfaces diagram, full tool table, auth/gate/fail-open state machine, dependency-choice rationale, Template-First reversal plan.
- progress.md — §E.1-§E.4 lifecycle skeleton (parser-load-bearing `§E.N` tokens).
- Design source: `.moai/reports/moai-autonomy-workflow-redesign-20260803.html` §3.4 + §3.6.
