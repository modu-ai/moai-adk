---
id: SPEC-E2E-REVIVAL-001
title: "Revive /moai e2e as a multi-platform, token-minimized E2E testing subsystem (web + mobile + desktop)"
version: "0.1.0"
status: draft
created: 2026-07-13
updated: 2026-07-13
author: manager-spec
priority: P1
phase: "v3.1.0 target"
module: "internal/template/templates/.claude (commands/moai, skills/moai/workflows, agents/moai) + internal/template (catalog.yaml, CI-guard tests)"
lifecycle: spec-anchored
tags: "e2e, testing, playwright, maestro, tauri, electron, template, agent, workflow, token-minimization"
era: V3R6
tier: L
related_specs: [SPEC-SUBCOMMAND-RETIRE-001, SPEC-HARNESS-EXECUTE-E2E-001, SPEC-THIN-CMDS-001]
---

# SPEC-E2E-REVIVAL-001 — Revive /moai e2e as a multi-platform, token-minimized E2E testing subsystem

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-07-13 | manager-spec | Plan-phase artifact set authored (Tier L, 6 artifacts: spec/plan/acceptance/research/design/progress). Baseline: retired workflow recovered via `git show c6b04d39c~1:.claude/skills/moai/workflows/e2e.md` (452 lines, web-only). Tool stack verified live on 2026-07-13 (see research.md §Sources). |

---

## §A Context & Problem

SPEC-SUBCOMMAND-RETIRE-001 (commit c6b04d39c, 2026-07-01) permanently removed the `/moai e2e` capability: the thin command `.claude/commands/moai/e2e.md`, the workflow skill `.claude/skills/moai/workflows/e2e.md` (452 lines), and the corresponding intent-router row — from BOTH the local tree and the template source. The retired version was **web-only** (Playwright CLI / Agent Browser / Chrome DevTools MCP / Claude in Chrome) but already encoded the load-bearing token-cost principle: **CLI output parsing = low cost, MCP round-trips = high cost**.

This SPEC revives the capability as a modern subsystem with three deltas over the retired baseline:

1. **Platform expansion**: web + **mobile** (Maestro / Appium / Detox) + **desktop** (Playwright-Electron / WebdriverIO+tauri-service / OS-level fallback), selected via project-type auto-detection.
2. **Dedicated specialist agent**: a new `e2e-specialist` retained agent (the retired version delegated to manager-develop; no dedicated agent ever existed).
3. **Token-minimization as a first-class requirement**: CLI-output-first execution, bounded output tails, file-redirect for verbose logs, MCP only when CLI cannot do the job.

Distribution is via the TEMPLATE SOURCE (`internal/template/templates/`) per the Template-First Rule, with catalog.yaml and CI-guard count-constant reconciliation.

### Measured baseline (verified 2026-07-13, this tree)

| Surface | Current state | Post-SPEC state |
|---------|---------------|-----------------|
| `.claude/skills/moai/SKILL.md` Priority 1 router (both trees) | 14 subcommand rows, **no** `e2e` row (`grep -c` for `e2e`: 0) | 15 rows, `e2e` row restored |
| `CLAUDE.md` §3 `Subcommands:` line (both trees) | 13 tokens, no `e2e` | 14 tokens incl. `e2e` |
| `CLAUDE.md` §4 agent catalog (both trees) | "exactly **10 retained agents** (9 MoAI-custom + 1 `Explore`)" | 11 retained (10 MoAI-custom + 1 `Explore`) |
| `.claude/agents/moai/*.md` (both trees) | 9 files | 10 files (+ `e2e-specialist.md`) |
| `.claude/commands/moai/` | 13 files per tree (local `.md`, template `.md.tmpl`) | 14 per tree (+ `e2e`) |
| `internal/template/catalog_tier_audit_test.go` | `expectedSkillCount = 28` (L158), `expectedAgentCount = 9` (L231) | 28 (UNCHANGED — see note), 10 |
| `internal/template/catalog_loader_test.go` | `expectedTotal = 37` (L55) | 38 |
| `internal/template/catalog.yaml` `core.agents` | 9 entries | 10 entries |

> **Count note**: the workflow skill lives INSIDE the `moai` skill directory (`.claude/skills/moai/workflows/e2e.md`), so `expectedSkillCount` (top-level skill directories) does NOT change. Only the agent-count constants move.

---

## §B Requirements (GEARS)

### Group A — Project-type detection & toolchain selection

- **REQ-E2E-001** (Ubiquitous): The e2e workflow shall auto-detect the project platform type (`web` / `mobile` / `desktop` / `mixed`) from project markers before any toolchain selection (marker matrix: design.md §B).
- **REQ-E2E-002** (Event-driven): **When** project-type detection completes, the workflow shall map the detected type to the default toolchain per the platform-toolchain matrix (web → Playwright CLI; mobile → Maestro; desktop → Playwright-Electron for Electron apps, WebdriverIO + `@wdio/tauri-service` for Tauri apps).
- **REQ-E2E-003** (Event-driven): **When** more than one platform type is detected (`mixed`), the workflow shall enumerate per-platform toolchain candidates and surface the selection through the ORCHESTRATOR's AskUserQuestion channel (one question per platform surface, recommended option first).
- **REQ-E2E-004** (Ubiquitous): The workflow shall route ALL user-facing toolchain/journey selection through the orchestrator's AskUserQuestion channel; the e2e-specialist agent shall obtain selections exclusively via its spawn prompt.
- **REQ-E2E-005** (Capability gate): **Where** a `--tool <name>` flag is provided, the workflow shall bypass the selection question and use the named toolchain directly.
- **REQ-E2E-006** (Event-detected): **When** the selected toolchain is detected as not installed (version probe fails), the workflow shall surface the exact install command(s) and, upon user approval via the orchestrator, perform the installation and re-verify with a version probe before proceeding.
- **REQ-E2E-007** (Event-detected): **When** no e2e-able surface is detected (e.g., a pure library with no web/mobile/desktop entry point), the workflow shall report "no e2e target detected" with the marker evidence and exit gracefully without creating test artifacts.

### Group B — Token minimization (first-class)

- **REQ-E2E-100** (Ubiquitous): The subsystem shall execute CLI-output-first: every capability achievable via CLI invocation shall use the CLI path; MCP tools shall be used ONLY for capabilities the selected CLI cannot provide (the per-capability classification lives in the workflow's tool matrix).
- **REQ-E2E-101** (Ubiquitous): The e2e-specialist shall bound in-context command output to exit code + bounded tail (≤50 lines OR ≤2KB, whichever is smaller), redirecting verbose output to files under the run-artifacts directory with citable paths (file-redirect contract, `agent-common-protocol.md` §Parallel Execution).
- **REQ-E2E-102** (State-driven): **While** an MCP-backed tool is active, the e2e-specialist shall prefer snapshot/batch reads (accessibility tree, DOM snapshot, aggregated trace insights) over per-element round-trips.
- **REQ-E2E-103** (Event-driven): **When** a test run produces reports (HTML report, traces, screenshots, recordings), the workflow shall persist them under project-local `e2e/` artifact directories and cite paths in the report — never inline the artifact content.
- **REQ-E2E-104** (Capability gate): **Where** recording is requested (`--record`), the workflow shall use the selected toolchain's native trace/recording facility (Playwright trace, Maestro recording, agent-browser trace) rather than MCP screenshot loops.
- **REQ-E2E-105** (Unwanted): The subsystem shall not require any MCP server as a hard dependency; every default platform path shall be fully executable with CLI-only tools.

### Group C — Deliverable artifacts

- **REQ-E2E-200** (Ubiquitous): The template source shall gain a thin command wrapper at `internal/template/templates/.claude/commands/moai/e2e.md.tmpl` conforming to the Thin Command Pattern: frontmatter with `description` + `allowed-tools` (CSV string), body <20 non-empty lines containing a `Skill(` invocation or `subagent` delegation directive (enforced by `TestCommandsThinPattern`).
- **REQ-E2E-201** (Ubiquitous): The template source shall gain a workflow skill at `internal/template/templates/.claude/skills/moai/workflows/e2e.md` with `user-invocable: false` frontmatter and the phase structure: Detection → Selection → Journey Mapping → Script Creation → Execution → Recording (optional) → Report.
- **REQ-E2E-202** (Ubiquitous): The template source shall gain a dedicated agent at `internal/template/templates/.claude/agents/moai/e2e-specialist.md` conforming to retained-agent frontmatter conventions (name, description with "NOT for:" clause, `tools:` CSV, `model: inherit`, `effort`, `color`, `permissionMode`, `memory`, `skills:` YAML array ≤2 entries).
- **REQ-E2E-203** (Unwanted): The e2e-specialist agent shall not list `Agent` in its `tools:` CSV and shall not list or invoke `AskUserQuestion`; when required inputs are missing, it shall return the structured `## Missing Inputs` blocker report (`agent-common-protocol.md` §Blocker Report Format).
- **REQ-E2E-204** (Ubiquitous): The workflow skill body shall delegate detection, script creation, and execution phases to the `e2e-specialist` subagent BY NAME, and the agent file shall exist at the path that delegation resolves to, in BOTH trees (cross-file reachability, not token presence).

### Group D — Router & catalog reachability

- **REQ-E2E-300** (Ubiquitous): The `/moai` SKILL.md Intent Router Priority 1 subcommand list shall regain an `**e2e**` row routing to the e2e workflow, in BOTH trees.
- **REQ-E2E-301** (Ubiquitous): The `moai` SKILL.md frontmatter `description` subcommand enumeration AND the `CLAUDE.md` §3 `Subcommands:` line shall include `e2e`, in BOTH trees.
- **REQ-E2E-302** (Ubiquitous): The `CLAUDE.md` §4 agent catalog shall register `e2e-specialist` in BOTH trees: the retained-agent count text updated (10 → 11 total; 9 → 10 MoAI-custom), a catalog table row added, and a Selection Decision Tree entry added.
- **REQ-E2E-303** (Capability gate): **Where** natural-language input expresses e2e-testing intent in any `conversation_language`, the Priority 3 semantic classification shall route to the e2e workflow (cue exemplars added to the P3 list; semantic, not literal-match).
- **REQ-E2E-304** (Ubiquitous): `internal/template/catalog.yaml` `core.agents` shall gain an `e2e-specialist` entry (name/tier/path/version authored manually; hash computed by `make build` → `gen-catalog-hashes --all`). NOTE: the hash generator only UPDATES existing entries — it does not create them (verified in `gen-catalog-hashes.go`).

### Group E — Distribution & CI reconciliation

- **REQ-E2E-400** (Ubiquitous): All new artifacts shall be authored in the TEMPLATE tree FIRST (`internal/template/templates/`), then synced to the local tree, per the Template-First Rule (CLAUDE.local.md §2 [HARD]).
- **REQ-E2E-401** (Ubiquitous): CI-guard count constants shall be reconciled with provenance comments in the same commit as the artifact additions: `expectedAgentCount` 9 → 10 (`catalog_tier_audit_test.go`), `expectedTotal` 37 → 38 (`catalog_loader_test.go`). `expectedSkillCount` stays 28.
- **REQ-E2E-402** (Ubiquitous): The new command file shall pass `TestCommandsThinPattern` and `TestCommandsFrontmatterConsistency`; the new agent shall pass `TestAgentFrontmatterAudit`, `TestAllAgentsInCatalog`, and `TestCatalogReferencesValid`; `go test ./internal/template/...` shall exit 0.
- **REQ-E2E-403** (Unwanted): Template-tree content shall not contain internal SPEC IDs, REQ/AC tokens, audit citations, internal work dates, commit SHAs, or CLAUDE.local references (template neutrality, CLAUDE.local.md §25) — the existing neutrality CI guard shall pass unchanged.
- **REQ-E2E-404** (Event-driven): **When** template authoring completes, `make build` shall be run so the embedded FS (`//go:embed all:templates`) and catalog hashes regenerate; the built binary shall carry the three new artifacts.
- **REQ-E2E-405** (Ubiquitous): The project-marker detection matrix shall treat all supported project ecosystems equally — no language/framework is presented as "primary" while others are "planned" (16-language neutrality, CLAUDE.local.md §15).

### Group F — Boundaries

- **REQ-E2E-500** (Unwanted): The revival shall not resurrect any of the 7 retired design-pack skills (`moai-domain-ideation`, `moai-domain-research`, `moai-domain-brand-design`, `moai-domain-copywriting`, `moai-domain-design-handoff`, `moai-workflow-design`, `moai-workflow-gan-loop`) nor the other 4 retired subcommands (design/brain/coverage/security).
- **REQ-E2E-501** (Ubiquitous): The e2e-specialist shall operate within the subagent boundary: blocker reports instead of prompts, results returned to the orchestrator, no nested agent spawning.
- **REQ-E2E-502** (Capability gate): **Where** the target desktop app is neither Electron nor Tauri (native toolkit apps), the workflow shall classify it as `desktop-native` and surface the OS-level accessibility/computer-use fallback as an EXPLICIT OPT-IN with a token-cost warning — never as a silent default. (Scope boundary pending confirmation — see plan.md §B D-7.)

---

## §C Non-functional Constraints

- **C-1** [HARD] Template-First: no file lands in local `.claude/` without a template-tree sibling authored first.
- **C-2** [HARD] Thin Command Pattern: command body <20 non-empty lines (`TestCommandsThinPattern` R3).
- **C-3** [HARD] Subagent boundary: e2e-specialist never prompts the user (Frozen-zone rule).
- **C-4** [HARD] Template neutrality: zero internal-development traces in template content (existing CI guard is the gate).
- **C-5** [HARD] All CI-guard tests, `go test ./...`, and `golangci-lint run` green at close.
- **C-6** Token economy: the workflow skill body itself stays lean (~≤450 lines, at parity with the retired baseline despite 3× platform scope — achieved by matrix-driven structure over per-tool prose duplication).
- **C-7** No new Go runtime code: this SPEC ships markdown/YAML template artifacts + test-constant updates only. The `moai` Go binary gains no new subcommand (the `/moai e2e` surface is slash-command-routed, exactly like the retired version).

---

## §D Acceptance Criteria Map

The full AC matrix (26 ACs), Given-When-Then scenarios, edge cases, and Definition of Done live in `acceptance.md`. Mapping summary:

| REQ group | ACs |
|-----------|-----|
| A (detection/selection) | AC-E2E-001 … AC-E2E-006 |
| B (token minimization) | AC-E2E-007 … AC-E2E-011 |
| C (artifacts) | AC-E2E-012 … AC-E2E-016 |
| D (router/catalog reachability) | AC-E2E-017 … AC-E2E-021 |
| E (distribution/CI) | AC-E2E-022 … AC-E2E-025 |
| F (boundaries) | AC-E2E-026 |

---

## §E Exclusions

The exclusions below are out of scope for this SPEC.

### Out of Scope — Retired design-pack resurrection

- The 7 design-pack skills and 4 sibling subcommands retired by the subcommand-retirement SPEC stay retired. Only `e2e` returns.

### Out of Scope — CI pipeline integration

- No GitHub Actions / CI job templates for running user-project e2e suites. The workflow produces locally runnable suites; wiring them into a user's CI is the user's (or a follow-up SPEC's) concern.

### Out of Scope — Dogfood e2e suites for moai-adk-go itself

- This SPEC ships the CAPABILITY (command + skill + agent). Authoring an actual e2e suite for the moai-adk-go repo or docs-site is not a deliverable.

### Out of Scope — MCP server provisioning automation

- The workflow may INSTRUCT the user how to add `chrome-devtools-mcp` to `.mcp.json`, but automated MCP provisioning stays with the existing `/moai project` MCP-provisioning flow.

### Out of Scope — Visual regression / screenshot-diff subsystem

- Pixel-diff baselines, perceptual diffing, and visual-regression storage are a separate capability; only plain screenshot/trace capture ships here.

### Out of Scope — docs-site 4-locale documentation

- adk.mo.ai.kr documentation for the revived subcommand is deferred (sync-phase decision or follow-up; see plan.md open question H-3).

### Out of Scope — Device-farm / cloud-device execution

- Maestro Cloud, BrowserStack, Sauce Labs, and other remote device farms are not integrated; local simulator/emulator/device execution only.
