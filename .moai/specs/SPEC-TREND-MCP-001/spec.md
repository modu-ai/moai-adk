---
id: SPEC-TREND-MCP-001
title: "Trend MCP tooling — Playwright + ast-grep bundle, opt-in recipes, generic atomic-RMW entry management"
version: "0.1.1"
status: draft
created: 2026-08-07
updated: 2026-08-07
author: manager-spec
priority: P2
phase: "v3.2 target"
module: "internal/cli (mcp_server.go, glm_tools.go), internal/template/templates/.mcp.json (NEW), .claude/rules/moai/core/settings-management.md"
lifecycle: spec-anchored
era: V3R6
tier: M
tags: "mcp, trend, playwright, ast-grep, semgrep, template-first, neutrality, atomic-rmw, opt-in, fail-open"
depends_on: [SPEC-MOAI-MCP-SERVER-001]
related_specs: [SPEC-HARNESS-MCP-PROVISION-001, SPEC-GLM-MCP-001]
---

# SPEC-TREND-MCP-001 — Trend MCP tooling (autonomy·speed Epic, §3.7, P2)

## HISTORY

- 2026-08-07 (plan-phase, iter-1) — Initial Tier M authoring. Operationalizes §3.7 "트렌드 MCP 도구 — 번들·opt-in·스킵 (R2)" of `.moai/reports/moai-autonomy-workflow-redesign-20260803.html`. Depends on the completed foundation SPEC-MOAI-MCP-SERVER-001 (the `moai mcp-server` thin stdio wrapper is live at `internal/cli/mcp_server.go`, registered via `newMCPServerCmd()` at `internal/cli/root.go:214`). This SPEC reverses the MOAI-MCP-SERVER-001 REQ-MCP-002 "no third-party entries" clause in the narrow sense (no third-party LOCAL-STDIO-WITH-SECRETS entries) while preserving its load-bearing intent (no secret-bearing, non-neutral, or credential-requiring servers in the template `.mcp.json`). All bundled third-party entries (context7, chrome-devtools, Playwright, ast-grep) are secret-free neutral HTTP/npx surfaces that pass §25 template-neutrality.

## §A. User Story

**As a** MoAI-ADK user (or an external MCP-capable host) who wants the harness to ship a curated, neutral set of high-leverage MCP servers — and a safe CLI to manage them at runtime,

**I want** (a) the template to provision a small, neutral, secret-free bundle (Context7, chrome-devtools, Playwright; ast-grep shipped as a default-disabled entry), (b) a generic `moai mcp add|remove|list` CLI that manages third-party entries via the existing atomic-RMW seam (`mutateClaudeJSONAtomic`) instead of hand-editing `.mcp.json`, (c) a recipe catalogue covering opt-in tools (Semgrep, GitHub MCP, Postgres/neon, Sentry, Codecov) and a skip-rationale record for retired/duplicate tools (Sequential Thinking, Filesystem, Git, Memory/KG, Brave/Exa),

**so that** (a) the e2e-tester agent and docs-site 4-locale functional verification gain a Playwright surface complementary to chrome-devtools' perf/Lighthouse role; (b) `/moai codemaps`, `/moai mx`, and migrations gain an opt-in ast-grep structural-search surface for the 16 supported languages; (c) the TRUST 5 Secured dimension gains a mechanical Semgrep path that turns the prose `moai-ref-owasp` / `moai-ref-secops` references into registry-rule evidence; and (d) users NEVER hand-edit `.mcp.json` (the atomic-RMW guard prevents mid-Claude-Code-session corruption, matching the SPEC-CLIFIX-CONCURRENCY-001 family).

## §B. Scope Summary

This SPEC delivers the §3.7 trend-MCP layer: a template-managed `.mcp.json` (the first one — `internal/template/templates/` currently has none), a generic atomic-RMW entry-management CLI, recipe documentation, skip-rationale documentation, and the doctrine update in `settings-management.md` that reconciles the new multi-entry provisioning contract with MOAI-MCP-SERVER-001 REQ-MCP-002. The MCP layer remains a SUPPLY surface — it feeds evidence to the existing gates; it does not replace `/moai gate` LSP, the verification-claim-integrity attributed-baseline, or the sync-auditor 4-dimension scoring.

### Out of Scope — multi-model convergence orchestration

- The `audit_model: multi` parallel fan-out + disagreement synthesis → owned by the future SPEC-AUDIT-MULTI-MODEL (this SPEC unblocks nothing there; the audit backends are untouched).

### Out of Scope — per-project-type MCP provisioning via /moai project

- The `/moai project` Phase 3.6 MCP matrix + harness-builder MCP fragment → owned by SPEC-HARNESS-MCP-PROVISION-001 (completed). This SPEC does NOT touch the project-type detection surface; the new generic CLI (`moai mcp add`) is a USER-operated entry-management tool, not a project-type provisioning step.

### Out of Scope — replacing /moai gate or sync-auditor

- `/moai gate` LSP, verification-claim-integrity attributed-baseline, and sync-auditor 4-dim scoring remain the SSOT gates. Semgrep, Codecov, and the audit backends SUPPLY evidence to those gates; this SPEC does NOT redefine any gate semantics, threshold, or PASS/FAIL contract.

### Out of Scope — retiring the GLM tools enable/disable CLI

- `moai glm tools enable|disable` (SPEC-GLM-MCP-001) remains the AUTH-bearing registrar for Z.AI servers (it carries the `${Z_AI_API_KEY}` token and the bearer-header template). This SPEC's generic `moai mcp add` does NOT handle authenticated HTTP servers; the GLM CLI continues to own that surface. The two CLIs share the underlying `mutateClaudeJSONAtomic` seam but diverge at the entry-construction layer.

### Out of Scope — docs-site + README 4-locale user-facing description

- The docs-site page + README 4-locale description of the new MCP catalogue surface → sync-phase (manager-docs). This SPEC's plan.md identifies the surface; the prose authoring is deferred to `/moai sync`.

## §C. Requirements (GEARS)

> Domain prefix `REQ-TMC-NNN` maps to SPEC domain `TREND-MCP`. GEARS compound clauses are used throughout. Every "wraps existing `internal/` function" requirement is backed by a verified file:line citation in `research.md` (the report's 2026-08-03 references were re-verified against the worktree HEAD `1010e9c43`).

### M1 — Template-managed .mcp.json + bundled neutral entries

**REQ-TMC-001** (Ubiquitous) The template source SHALL establish `internal/template/templates/.mcp.json` as the template-managed MCP provisioning surface, carrying exactly five structural neutral entries: three active at the distributed default (`context7`, `chrome-devtools`, `playwright`) plus two documented-but-disabled (`ast-grep` opt-in per REQ-TMC-004, `moai` opt-in per REQ-TMC-003). The active-vs-structural distinction is explicit: exactly 3 entries in the active `mcpServers` map at the distributed default, and 5 entries documented in the template/recipe catalogue (3 active + 2 disabled).

**REQ-TMC-002** (Unwanted) The template `.mcp.json` SHALL NOT carry any secret, any resolved API key, any SPEC ID, any commit SHA, any macOS-bias path, any `CLAUDE.local.md` reference, or any `PR #N` reference — every entry is either secret-free npx/uvx stdio OR HTTP-with-`${VAR}`-literal expanded by the Claude Code runtime at load, and every entry passes the §25 CI guard (`template-neutrality-check.yaml` + `internal_content_leak_test.go`).

**REQ-TMC-003** (Capability gate) **Where** the user has not opted into the `moai mcp-server` local server, the template provisioning SHALL leave the `moai` entry OFF in the distributed default (the entry is opt-in per SPEC-MOAI-MCP-SERVER-001 REQ-MCP-002; this SPEC does not change that single-server opt-in property — only widens the third-party-neutral-entry surface around it).

**REQ-TMC-004** (State-driven) **While** the `ast-grep` tool is documented in the recipe catalogue, the template `.mcp.json` SHALL ship it default-DISABLED via omission from the active `mcpServers` map — the entry is documented in the recipe catalogue with a one-line `moai mcp add ast-grep ...` activation command, and NO JSONC/`$comment`-anchored disable form is used in the template (the active distributed map carries only the three default-on entries: `context7`, `chrome-devtools`, `playwright`).

### M2 — Generic atomic-RMW entry-management CLI

**REQ-TMC-005** (Event-driven) **When** the user runs `moai mcp add <name> --command <cmd> [--args ...] [--scope user|project] [--env KEY=VAL ...] [--type stdio|http] [--url URL] [--headers ...]`, the CLI SHALL register the entry via the existing `mutateClaudeJSONAtomic` seam (`internal/cli/glm_tools.go:541`), reusing the same flock + compare-retry guard, the same backup-before-publish discipline, and the same idempotent-skip behavior — NEVER hand-editing `.mcp.json` or `~/.claude.json`.

**REQ-TMC-006** (Event-driven) **When** the user runs `moai mcp remove <name> [--scope user|project]`, the CLI SHALL remove only the named entry, preserving every unrelated `mcpServers` entry (context7, chrome-devtools, moai, zai-mcp-server, web_search_prime, web_reader, etc.) via the same partial-delete safety contract `disableMCPServerForTool` already enforces (`internal/cli/glm_tools.go:852`).

**REQ-TMC-007** (Event-driven) **When** the user runs `moai mcp list [--scope user|project]`, the CLI SHALL emit a deterministic, machine-readable listing (JSON when `--json`, plain text otherwise) of the current `mcpServers` entries, distinguishing local-stdio from HTTP and flagging entries that carry `${VAR}` literal env references (so a user can see at a glance which entries need runtime secret expansion).

**REQ-TMC-008** (Capability gate) **Where** the new generic CLI shares the atomic-RMW seam with `glm_tools.go`, the shared helper (`mutateClaudeJSONAtomic`, `writeClaudeJSONBytes`, `backupClaudeJSON`, `readClaudeJSONRaw`, `withClaudeJSONLock`) SHALL be reused UNCHANGED — no fork, no parallel lock convention, no second marshal location. If a refactor extracts a narrower helper, the refactor is additive (the existing `glm_tools.go` callers continue to compile and pass their existing tests).

**REQ-TMC-009** (Unwanted) The generic CLI SHALL NOT accept a positional secret value on the command line; any `--env KEY=VAL` whose VALUE resolves to a non-`${VAR}`-literal form SHALL be rejected with a structured error pointing the user to the `${VAR}`-literal form (secret hygiene matching the MOAI-MCP-SERVER-001 C3 invariant — resolved secrets are NEVER serialized into a git-tracked `.mcp.json`).

**REQ-TMC-010** (State-driven) **While** the subagent boundary (C-HRA-008) binds, the new CLI subcommands SHALL NOT call `AskUserQuestion` or emit free-form user-facing questions — every interaction is positional args + `--flag` defaults + structured stderr errors, matching the canonical `TestWeb_NoAskUserQuestion` static-guard pattern at `internal/cli/web_test.go`.

### M3 — Recipe catalogue + skip rationale + doctrine reconciliation

**REQ-TMC-011** (Ubiquitous) The recipe catalogue SHALL cover every tool classified in §3.7 of the autonomy report: bundled (Playwright default-on, ast-grep default-disabled), opt-in recipes (Semgrep, GitHub MCP, Postgres/neon, Sentry, Codecov), and skipped-with-rationale (Sequential Thinking — retired by `ultrathink`; Filesystem — native Read/Write/Edit duplication; Git — branch-guard conflict; Memory/KG — auto-memory duplication; Brave/Exa — built-in `WebSearch` duplication).

**REQ-TMC-012** (Capability gate) **Where** the catalogue documents an opt-in recipe, the recipe SHALL be a copy-pasteable `.mcp.json` snippet AND a one-line `moai mcp add ...` equivalent — so a user can choose either path (manual JSON edit or atomic-RMW CLI), and the two forms are byte-equivalent after normalization.

**REQ-TMC-013** (Ubiquitous) The doctrine update SHALL revise `.claude/rules/moai/core/settings-management.md` § MCP Configuration to describe the new multi-entry provisioning contract: exactly one LOCAL-STDIO server (`moai`, opt-in per MOAI-MCP-SERVER-001) PLUS a small set of neutral third-party entries (context7, chrome-devtools, Playwright default-on; ast-grep default-disabled) that carry no secrets and pass §25 — the prior "no third-party entries" wording is reconciled to "no third-party entries THAT CARRY SECRETS, require credentials, or fail §25 neutrality".

**REQ-TMC-014** (Unwanted) The catalogue SHALL NOT present Semgrep, Codecov, or any audit tool as a gate replacement — every recipe entry carries an explicit note that `/moai gate` LSP, verification-claim-integrity attributed-baseline, and sync-auditor 4-dim scoring remain the SSOT gates, and the tool only SUPPLIES evidence (per the autonomy report §3.7 footer and CLAUDE.md §6).

### Cross-cutting

**REQ-TMC-015** (Capability gate) **Where** any new env-var name, threshold, or default is introduced by this SPEC, it SHALL be defined as a constant in `internal/config/envkeys.go` (env-var names) or `internal/config/defaults.go` (thresholds/defaults), and the template `.mcp.json` + new CLI SHALL stay generic — no hardcoded absolute paths, model names, or org identifiers in distributed surfaces (CLAUDE.local.md §14 hardcoding prevention).

**REQ-TMC-016** (Capability gate) **Where** any `.claude/` template change or new template file is made, it SHALL be mirrored to `internal/template/templates/` and regenerated via `make build` (CLAUDE.local.md §2 Template-First), with §25 neutrality enforced by the CI guard — the new `.mcp.json` is the load-bearing new surface and MUST pass `TestTemplateNeutralityAudit` + `internal_content_leak_test.go` on first commit.

## §D. Acceptance Criteria (summary — full Given-When-Then in acceptance.md)

Each requirement maps to one or more binary-testable acceptance criteria (AC-TMC-NNN) enumerated in `acceptance.md`. The acceptance matrix binds 16 ACs across 16 REQs: M1 ↔ AC-TMC-001..004, M2 ↔ AC-TMC-005..010, M3 ↔ AC-TMC-011..014, cross-cutting ↔ AC-TMC-015..016. Severity classification, traceability, and Definition of Done live in `acceptance.md`.

## §E. Constraints (non-functional)

- **C1 — Template-First**: the new `.mcp.json` is authored at `internal/template/templates/.mcp.json` FIRST, then `make build` regenerates the embedded catalog, then it lands at the repo-root `.mcp.json` (the current repo-root `.mcp.json` is NOT template-managed today; this SPEC establishes that surface and aligns the repo-root file with the template source).
- **C2 — §25 neutrality**: the template `.mcp.json` passes both CI guards (`template-neutrality-check.yaml` for C1/C2/C4/C5/C6/C8 + `internal_content_leak_test.go` for SPEC-ID C1 / commit-SHA C7 / date C3).
- **C3 — secret hygiene**: resolved secrets are NEVER serialized into the template `.mcp.json`; only `${VAR}` literals (expanded by the Claude Code runtime) appear, and the generic CLI rejects positional secret values.
- **C4 — atomic-RMW reuse**: the new generic CLI reuses `mutateClaudeJSONAtomic` unchanged; no fork, no second lock convention.
- **C5 — subagent boundary**: the new CLI subcommands never call `AskUserQuestion`; the canonical static guard extends to the new files.
- **C6 — opt-in preserved**: the `moai` local server stays opt-in (MOAI-MCP-SERVER-001 REQ-MCP-002 single-server opt-in property preserved).
- **C7 — MCP-not-gate-replacement**: every recipe entry carries the "supply, do not redefine" note.
- **C8 — hardcoding prevention**: env names → `envkeys.go`; thresholds → `defaults.go`; no org/model/absolute-path in distributed surfaces.

## §F. Dependencies

- **go.mod**: NO new dependency. The generic CLI reuses the existing `internal/cli` + `internal/atomicfile` + `internal/lockfile` packages already on the tree. The `mark3labs/mcp-go` SDK added by SPEC-MOAI-MCP-SERVER-001 is consumed only by the `moai mcp-server` subcommand; this SPEC does not extend its use.
- **Existing internal packages (read-mostly reuse)**: `internal/cli` (`glm_tools.go` atomic-RMW seam, `mcp_server.go` entry builder pattern, `mcp_audit.go` secret-hygiene helpers), `internal/config` (envkeys, defaults), `internal/atomicfile`, `internal/lockfile`.
- **External runtimes (runtime, optional)**: `npx` (Node.js ≥ 22 — already required by chrome-devtools/context7) for Playwright; `uvx` (Python ≥ 3.11) for ast-grep-mcp; neither is required at template-distribution time (the `.mcp.json` entries are inert until the user's Claude Code session loads them).
- **Epic positioning (not hard technical deps)**: the autonomy·speed Epic sequences this SPEC after SPEC-MOAI-MCP-SERVER-001 (completed foundation). The TREND-MCP layer is self-contained — it functions without the future AUDIT-MULTI-MODEL SPEC, and without the autonomy-tier wiring.

## §G. Risks

- **R1 — Doctrine-reversal perception**: widening MOAI-MCP-SERVER-001 REQ-MCP-002's "no third-party entries" clause could be read as reversing a just-completed contract. Mitigation: the reversal is narrowly scoped to secret-free neutral entries that pass §25; the load-bearing intent (no secrets / no credentials / no non-neutrality in the template `.mcp.json`) is preserved and stated explicitly in REQ-TMC-002 + REQ-TMC-013 + the doctrine update.
- **R2 — ast-grep default-disabled UX**: shipping a commented-out entry is unusual; users may miss it. Mitigation: the recipe catalogue surfaces it prominently; the `moai mcp list` output flags default-disabled entries.
- **R3 — Shared seam fork hazard**: extracting a narrower helper from `glm_tools.go` risks forking the atomic-RMW behavior. Mitigation: REQ-TMC-008 mandates reuse-unchanged; any refactor is additive and the existing `glm_tools.go` tests must continue to pass.
- **R4 — uvx runtime availability**: ast-grep-mcp requires Python ≥ 3.11 + uv installed; not every user has it. Mitigation: the entry is default-disabled; users who want it install `uv` explicitly.
- **R5 — Semgrep/Codecov gate-conflation**: users may treat Semgrep findings or Codecov coverage as MoAI gate verdicts. Mitigation: REQ-TMC-014 + every recipe entry carries the "supply, do not redefine" note; the catalogue explicitly names the SSOT gates.

## §H. Cross-References

- Design source: `.moai/reports/moai-autonomy-workflow-redesign-20260803.html` §3.7 (트렌드 MCP 도구 — 번들·opt-in·스킵).
- Dependency (completed): `.moai/specs/SPEC-MOAI-MCP-SERVER-001/` — owns the `moai mcp-server` thin wrapper, the `buildMoaiMCPServerEntry` builder, the audit-backends secret hygiene, and the REQ-MCP-002 "no third-party entries" clause this SPEC reconciles.
- Sibling (completed): `.moai/specs/SPEC-HARNESS-MCP-PROVISION-001/` — owns the per-project-type `/moai project` Phase 3.6 MCP matrix surface (distinct from this SPEC's user-operated generic CLI).
- Sibling (completed): `.moai/specs/SPEC-GLM-MCP-001/` — owns the AUTH-bearing Z.AI registrar (`moai glm tools enable|disable`), whose `${Z_AI_API_KEY}`-literal pattern this SPEC's generic CLI generalizes (without inheriting the auth surface).
- Sibling (completed): `.moai/specs/SPEC-CLIFIX-CONCURRENCY-001/` — owns the `mutateClaudeJSONAtomic` flock + compare-retry guard this SPEC reuses unchanged.
- Schema SSOT: `.claude/rules/moai/development/spec-frontmatter-schema.md`.
- §25 SSOT: `.moai/docs/template-internal-isolation-doctrine.md` + `.github/workflows/template-neutrality-check.yaml` + `internal/template/internal_content_leak_test.go`.
