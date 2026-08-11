# Acceptance — SPEC-MOAI-MCP-SERVER-001

> Verification layer. Each criterion is `AC-MCP-NNN`, written as binary-testable `Given … When … Then …`. GEARS requirements live in `spec.md` (`REQ-MCP-NNN`); this file MUST NOT restate them as requirements — Given-When-Then is the correct format here (audit contract: `plan-auditor.md` M3 § Scope / MP-2). Traceability matrix + severity + Definition of Done at the end.

## Severity legend

- **MUST** — blocks `completed` (a failing MUST AC blocks convergence/merge).
- **SHOULD** — strong quality signal; a fail is a debt item, not a blocker.
- **MAY** — opportunistic / nice-to-have verification.

## §D. AC Matrix

### M1 — Server scaffold + core read/status tools + `.mcp.json` provisioning

**AC-MCP-001** (MUST) — `moai mcp-server` stdio round-trip
- **Given** the `moai` binary is built with the new `mcp-server` subcommand registered via `root.go` `AddCommand`,
- **When** a JSON-RPC `initialize` request is sent over stdio, followed by `tools/list`,
- **Then** the server returns a valid `initialize` response and a non-empty `tools/list` array, and the process stays alive (long-running, blocking `RunE`) until the stdio stream closes.

> **Amendment notice (2026-08-12, v0.1.0 → v0.2.0).** AC-MCP-002 and AC-MCP-006 below are amended in place. Both traced to REQ-MCP-002, whose opt-in precondition the project owner inverted to default-on so the self-hosted `moai` MCP server becomes a first-class default of MoAI-ADK. An auditor asking "why is a completed MUST AC being inverted?" should read this: the original opt-in clause encoded a **risk posture**, not a technical constraint, and the risk it guarded (an unsolicited process started by the runtime at session start) is now accepted deliberately — the process is the user's own already-installed `moai` binary running a fixed subcommand, with no credentials and no third-party code. Every other clause of both ACs — exactly one neutral entry, written through `mutateClaudeJSONAtomic` + `resolveConfigPath` + `buildMoaiMCPServerEntry`, with no secret values serialized — is preserved verbatim, because those clauses are what actually made the original AC load-bearing. The implementing SPEC is `SPEC-MCP-DEFAULT-ON-001`; full rationale and scope live in `spec.md` HISTORY § Amendments.

**AC-MCP-002** (MUST) — **[AMENDED]** default-on provisioning, explicit opt-out honored
- **Given** a fresh project and a user who did NOT decline MCP provisioning,
- **When** `moai init` runs,
- **Then** exactly one `moai` entry IS written to the project's `.mcp.json`. **And given** a user who explicitly declined, **when** `moai init` runs, **then** no `moai` entry is written and the decline is silent. (REQ-MCP-002 as amended, C6)
- Original (v0.1.0, superseded): "**Given** a fresh project with no user opt-in, **When** `moai init` / `moai update` runs without the MCP opt-in flag, **Then** no `moai` entry is written to `.mcp.json`, and the `mcp-server` subcommand is not provisioned into any distributed template by default."

**AC-MCP-003** (MUST) — thin-wrapper parity with the CLI
- **Given** a core tool (e.g. `session_list`, `spec_audit`) and the equivalent CLI invocation,
- **When** both are run against the same project state,
- **Then** the MCP tool result and the CLI `--json` result are produced by the SAME `internal/` function (verified by code inspection: the handler calls that function, no reimplementation). (REQ-MCP-003, C1, AP-1)

**AC-MCP-004** (MUST) — `tools/list` declares JSON Schema
- **Given** the running `moai mcp-server`,
- **When** the host requests `tools/list`,
- **Then** every declared tool carries a `name` and an `inputSchema` (JSON Schema), so a caller can invoke it type-safely without `--help` guessing. (REQ-MCP-004)

**AC-MCP-005** (MUST) — core tools wrap the verified integration points
- **Given** the core tools are registered,
- **When** each is invoked against representative project state,
- **Then** `session_list` returns `session.QueryActiveWork` output, `goal_status` returns `goal.LoadGoal` output, `goal_arm` arms via `goal.NewGoal`+`goal.SaveGoal`, `spec_progress` returns the SPEC scanner output, `verify_snapshot`/`verify_trend` consume `verify.Load`+`verify.RecordCheck`, `spec_audit`/`spec_drift` return `spec.Audit` output, and `audit_cache` reads `runtime.AuditCache`. Each handler's source cites the verified file:line from `research.md`. (REQ-MCP-005)

**AC-MCP-006** (MUST) — **[AMENDED]** `.mcp.json` single neutral entry via atomic-config helpers
- **Given** MCP provisioning is enabled — which, per the amended REQ-MCP-002, is the default state absent an explicit opt-out (original precondition: "the user opts in to MCP provisioning"),
- **When** provisioning runs,
- **Then** exactly one `moai` entry `{"command":"moai","args":["mcp-server"]}` is written through `mutateClaudeJSONAtomic` + `resolveConfigPath` + a NEW `buildMoaiMCPServerEntry` (modeled on `buildZAIMCPEntry`), no third-party entries are added, and no secret values are serialized. (REQ-MCP-002)

### M2 — Codex audit backend (Phase 1 tools) + Stop-hook gate

**AC-MCP-007** (MUST) — `codex_audit` unified modes + schema output
- **Given** the `codex` binary is present and authenticated,
- **When** `codex_audit` is invoked with `mode=native` and separately with `mode=adversarial`,
- **Then** native shells out to `review/start`, adversarial shells out to `turn/start` + the adversarial-review prompt, and BOTH return a result shaped by `review-output.schema.json` (`verdict`, `summary`, `findings[]` with `severity/title/body/file/line/confidence/recommendation`, `next_steps`). (REQ-MCP-006)

**AC-MCP-008** (MUST) — `codex_setup` Go probe, no Node bridge
- **Given** the `codex_setup` tool,
- **When** it runs,
- **Then** it performs `exec.LookPath("codex")` + `codex --version` (Go reimplementation), classifies the auth provider (ChatGPT / apiKey / provider), and exposes `enable_review_gate` — with NO dependency on any Node.js / `.mjs` bridge. (REQ-MCP-007)

**AC-MCP-009** (MUST) — review-gate self-gate prevents false blocks
- **Given** `workflow.codex.review_gate.enabled` is set and the previous turn produced no code edit (status report / review-result / no-op),
- **When** the `moai hook codex-review-gate` Stop hook fires,
- **Then** it ALLOWs immediately without invoking codex (the mandatory self-gate), so a non-editing turn is never falsely blocked. (REQ-MCP-008)

**AC-MCP-010** (MUST) — review gate opt-in + 900 s timeout
- **Given** the review-gate hook configuration,
- **When** its manifest is inspected,
- **Then** it is opt-in (`workflow.codex.review_gate.enabled`, BranchGuard pattern), the moai-default 5 s hook timeout is overridden to 900 s for this hook only, and it emits the standard ALLOW/BLOCK contract. (REQ-MCP-008)

### M3 — GLM audit backend + 3-way config + auth + fail-open

**AC-MCP-011** (MUST) — GLM direct z.ai API call
- **Given** the GLM audit backend is selected and `~/.moai/.env.glm` holds a valid key,
- **When** a GLM audit is invoked,
- **Then** the server calls `https://api.z.ai/api/anthropic` (Anthropic-compatible) directly with the audit prompt, using the existing z.ai client plumbing — NOT the z.ai MCP server and NOT any z.ai gateway. (REQ-MCP-009)

**AC-MCP-012** (MUST) — `audit_model` + `audit_gate` enums + default profile
- **Given** the audit config schema,
- **When** `audit_model` and per-auditor `audit_gate` are validated,
- **Then** `audit_model` accepts exactly `{claude, codex, glm, multi}`, `audit_gate` accepts exactly `{off, advisory, required}`, the default `audit_gate` is `required`, and the default profile is claude + codex required, glm advisory. `multi` is accepted as a token but its convergence logic is NOT implemented here. (REQ-MCP-010)

**AC-MCP-013** (MUST) — secret hygiene (`${VAR}` literal, no serialization)
- **Given** a project whose `.mcp.json` is git-tracked,
- **When** codex/glm provisioning writes any env reference,
- **Then** the value is a `${CODEX_API_KEY}` / `${GLM_API_KEY}` literal (expanded by the host runtime), NEVER the resolved secret value; and a grep of the committed `.mcp.json` finds no literal API key. (REQ-MCP-011, C3, AP-2)

**AC-MCP-014** (MUST) — fail-open on missing/unauthenticated auditor
- **Given** a selected non-Claude auditor (codex or glm) is missing or unauthenticated,
- **When** an audit selecting that auditor runs,
- **Then** the tool returns `VerdictInconclusive` for that auditor and the workflow falls back to the active auditor (claude) WITHOUT hard-blocking. (REQ-MCP-012, C2, AP-3)

**AC-MCP-015** (MUST) — model/effort SSOT via `ResolveAgentModelEffort`
- **Given** the `codex_audit` and GLM-audit code paths,
- **When** they resolve model and effort,
- **Then** they call `template.ResolveAgentModelEffort` (`internal/template/profile_matrix.go:385`) and do NOT read agent frontmatter or `llm.agent_overrides` directly (verified by grep: no direct frontmatter/override read in the MCP handler package). (REQ-MCP-013, C4, AP-4)

**AC-MCP-016** (MUST) — subagent boundary (structured result, no `AskUserQuestion`)
- **Given** any MCP tool handler facing a missing-input or inconclusive condition,
- **When** it cannot proceed,
- **Then** it returns a structured result (including `VerdictInconclusive`) and NEVER calls `AskUserQuestion` or emits a free-form user question (verified by grep: no `AskUserQuestion` / `mcp__askuser` reference in the MCP handler package). (REQ-MCP-014, C5, AP-5)

**AC-MCP-017** (SHOULD) — 3-way selection resolves a single active backend
- **Given** `audit_model` is one of `claude` / `codex` / `glm`,
- **When** an audit runs,
- **Then** exactly one backend executes (per its `audit_gate`); `audit_model: multi` is accepted as a stored value but its parallel-orchestration behavior is NOT implemented in this SPEC (deferred to SPEC-AUDIT-MULTI-MODEL). (REQ-MCP-010, AP-8)

### M4 — init/web selection UI + Template-First reversal + docs

**AC-MCP-018** (MUST) — Template-First reversal applied, neutral
- **Given** `.claude/rules/moai/core/settings-management.md` line 33 currently says "MoAI-ADK no longer ships or provisions MCP servers via `.mcp.json`",
- **When** M4 lands,
- **Then** that statement is reversed to provision exactly ONE local server (`moai mcp-server`), and the replacement prose + `.mcp.json` entry contain NO SPEC-ID, NO commit SHA, NO internal date, NO macOS-bias path (§25 neutrality). (REQ-MCP-016, C7)

**AC-MCP-019** (MUST) — template source + CI guard + `make build` in one change
- **Given** the reversal touches a distributed template surface,
- **When** the M4 change is staged,
- **Then** `internal/template/templates/` mirror, the CI guard (`template-neutrality-check.yaml` / `internal_content_leak_test.go`), and a `make build` regeneration are all in the SAME change, and the §25 neutrality CI check is green. (REQ-MCP-016, REQ-MCP-018, AP-6)

**AC-MCP-020** (MUST) — `moai init` wizard selection surface
- **Given** the `moai init` wizard (`internal/cli/wizard/questions.go` `Page3Questions`),
- **When** the user reaches page 3,
- **Then** `audit_model`, per-auditor `audit_gate`, `codex_audit_enabled`, and `mcp_tools_opt_in` are presented, and the answers flow through `applyWizardPage3ToOpts` (`internal/cli/init.go`) into the project options. (REQ-MCP-015)

**AC-MCP-021** (MUST) — `moai web` console selection surface
- **Given** the `moai web` console,
- **When** the audit/console settings are rendered,
- **Then** the same `audit_model` + per-auditor `audit_gate` selection is surfaced through the identical interpreter used by the wizard (`ResolveAgentModelEffort` SSOT respected). (REQ-MCP-015)

### Cross-cutting

**AC-MCP-022** (MUST) — M0 design decisions documented + locked
- **Given** M0 completes,
- **Then** the SDK version pin, the `audit_model`/`audit_gate` schema names, the adopted `review-output.schema.json` shape, and the reversal wording are recorded in `design.md` and referenced from `progress.md` §E.2.

**AC-MCP-023** (MUST) — hardcoding prevention (env/threshold constants)
- **Given** the new env-var names and thresholds introduced by this SPEC,
- **When** the code is inspected,
- **Then** env-var names are constants in `internal/config/envkeys.go`, thresholds/defaults live in `internal/config/defaults.go`, and the `.mcp.json` entry + `mcp-server` command are generic (no hardcoded absolute paths / model names / org identifiers in distributed surfaces). (REQ-MCP-017, C8)

**AC-MCP-024** (MUST) — full suite green
- **Given** all milestones land,
- **When** `go test ./...` + `go vet ./...` + the §25 CI guard run,
- **Then** all pass (full suite, not affected-package-only — per the trust-but-verify lesson that affected-package-only self-report misses cross-cutting failures).

## Edge cases (negative tests, MUST)

- **EC-1**: `codex` binary absent ⇒ `codex_audit` returns `VerdictInconclusive`, no panic, workflow falls back to claude (AC-MCP-014).
- **EC-2**: `~/.moai/.env.glm` absent/unreadable ⇒ GLM audit returns `VerdictInconclusive`, falls back to claude (AC-MCP-014).
- **EC-3**: `.mcp.json` write races with another session ⇒ `mutateClaudeJSONAtomic` (flock + RMW) preserves atomicity; no partial write.
- **EC-4**: `audit_model: multi` selected ⇒ stored and acknowledged, but NO parallel orchestration runs (convergence is SPEC-AUDIT-MULTI-MODEL; AC-MCP-017).
- **EC-5**: review gate fires on a no-edit turn ⇒ ALLOW immediately (self-gate; AC-MCP-009).
- **EC-6**: a git-tracked `.mcp.json` after provisioning ⇒ grep finds NO literal `CODEX_API_KEY`/`GLM_API_KEY` value, only `${...}` literals (AC-MCP-013).

## §D.7 Closure gates (Definition of Done)

- All MUST ACs PASS with attributed evidence (command + verbatim output) in `progress.md` §E.2/§E.3.
- `go test ./...` + `go vet ./...` + §25 CI guard green (AC-MCP-024).
- The `codename → canonical ID` normalization (`SPEC-MOAI-MCP-SERVER` → `SPEC-MOAI-MCP-SERVER-001`) is reflected consistently in all artifacts.
- No `[NEEDS CLARIFICATION]` markers remain in `plan.md` / `research.md`.
- The Template-First reversal is mirrored to template source + CI guard + `make build` in one change (AC-MCP-019).

## Traceability matrix

| REQ | AC | Milestone |
|-----|----|-----------|
| REQ-MCP-001 | AC-MCP-001 | M1 |
| REQ-MCP-002 | AC-MCP-002, AC-MCP-006 | M1 |
| REQ-MCP-003 | AC-MCP-003 | M1 |
| REQ-MCP-004 | AC-MCP-004 | M1 |
| REQ-MCP-005 | AC-MCP-005 | M1 |
| REQ-MCP-006 | AC-MCP-007 | M2 |
| REQ-MCP-007 | AC-MCP-008 | M2 |
| REQ-MCP-008 | AC-MCP-009, AC-MCP-010 | M2 |
| REQ-MCP-009 | AC-MCP-011 | M3 |
| REQ-MCP-010 | AC-MCP-012, AC-MCP-017 | M3 |
| REQ-MCP-011 | AC-MCP-013 | M3 |
| REQ-MCP-012 | AC-MCP-014 | M3 |
| REQ-MCP-013 | AC-MCP-015 | M3 |
| REQ-MCP-014 | AC-MCP-016 | M3 |
| REQ-MCP-015 | AC-MCP-020, AC-MCP-021 | M4 |
| REQ-MCP-016 | AC-MCP-018, AC-MCP-019 | M4 |
| REQ-MCP-017 | AC-MCP-023 | cross |
| REQ-MCP-018 | AC-MCP-019, AC-MCP-024 | cross |

> **AC-MCP-022 — process/milestone gate (traceability annotation).** AC-MCP-022 ("M0 design decisions documented + locked") is intentionally outside the REQ→AC behavioral traceability set: it verifies that the design decisions REQ-MCP-001 (SDK/server scaffold), REQ-MCP-010 (`audit_model`/`audit_gate` schema), and REQ-MCP-016 (Template-First reversal wording) rest on are locked at M0 — it is not itself a behavioral requirement. Recorded here to resolve the traceability orphan rather than masquerading as a behavioral AC.
