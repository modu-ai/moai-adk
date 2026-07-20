---
id: SPEC-E2E-REVIVAL-001
title: "Revive /moai e2e as a multi-platform, token-minimized E2E testing subsystem (web + mobile + desktop)"
version: "0.1.2"
status: completed
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
related_specs: [SPEC-SUBCOMMAND-RETIRE-001, SPEC-HARNESS-EXECUTE-E2E-001]
---

# SPEC-E2E-REVIVAL-001 — Revive /moai e2e as a multi-platform, token-minimized E2E testing subsystem

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-07-13 | manager-spec | Plan-phase artifact set authored (Tier L, 6 artifacts: spec/plan/acceptance/research/design/progress). Baseline: retired workflow recovered via `git show c6b04d39c~1:.claude/skills/moai/workflows/e2e.md` (452 lines, web-only). Tool stack verified live on 2026-07-13 (see research.md §Sources). |
| 0.1.1 | 2026-07-13 | manager-spec | Audit-fix pass (plan-audit iter-1 FAIL 0.80, D1-D12). User-resolved (orchestrator AskUserQuestion round): D-7 desktop-native fallback DEFERRED to follow-up (REQ-E2E-502 removed; AC-E2E-006 first narrowed to REQ-E2E-007 per D6); D-8 Maestro mobile default CONFIRMED; H-3 docs-site deferred. D2: REQ-E2E-302 extended to the full re-measured 12-file/24-site count-literal surface + post-change invariance grep. D3: pipe-broken table-cell commands relocated to acceptance.md § Executable Command Block (unescaped, non-vacuous). D4: AC-E2E-027 added for REQ-E2E-405. D5: §A baseline corrected — catalog.yaml core.agents is 8, not 9 (builder-harness lives under harness_generated). D7: lint gates target spec.md FILE (directory arg fails with ParseFailure). D8: SPEC-THIN-CMDS-001 dropped from related_specs (survives only as a commands_audit_test.go provenance comment). D9/D10/D12: AC-E2E-012 count prose, GEARS Event-driven labels, TestAgentFrontmatterAudit coverage claims corrected. |
| 0.1.2 | 2026-07-13 | manager-spec | Residual-fix pass (plan-audit iter-2 PASS-WITH-DEBT 0.88, D13-D15). D13: REQ-E2E-302 surface re-derived TRULY repo-wide, restructured into rings — ring 2 template-distributed skill modules (4 template files / 8 sites; the 4 LOCAL siblings observed TWO generations stale at "8 retained agents / 7 MoAI-custom" [7 sites] — run-phase normalizes both trees), ring 3 README.md (3 sites; ko/ja/zh carry no ERE-family matches → locale-language review), ring 4 Go tier-profile display pin (NEW REQ-E2E-305 + C-7 carve-out — hard-coded display list means no test fails on omission), docs-site 10-file inventory EXPLICITLY deferred (§E). CMD-019-INV widened to 19 files (re-measured baseline 38) + CMD-019-INV-B added for the local 8/7-era family (baseline 7). AC-E2E-028 added (28 ACs). D14: G3/E7 recalibrated to "0 errors" (pre-existing StatusGitConsistency WARNING expected until sync close). D15: AC-E2E-008 in-cell command moved to CMD-008. |

---

## §A Context & Problem

SPEC-SUBCOMMAND-RETIRE-001 (commit c6b04d39c, 2026-07-01) permanently removed the `/moai e2e` capability: the thin command `.claude/commands/moai/e2e.md`, the workflow skill `.claude/skills/moai/workflows/e2e.md` (452 lines), and the corresponding intent-router row — from BOTH the local tree and the template source. The retired version was **web-only** (Playwright CLI / Agent Browser / Chrome DevTools MCP / Claude in Chrome) but already encoded the load-bearing token-cost principle: **CLI output parsing = low cost, MCP round-trips = high cost**.

This SPEC revives the capability as a modern subsystem with three deltas over the retired baseline:

1. **Platform expansion**: web + **mobile** (Maestro / Appium / Detox) + **desktop** (Playwright-Electron for Electron apps, WebdriverIO + tauri-service for Tauri apps), selected via project-type auto-detection. Native-desktop (non-Electron/non-Tauri) automation is DEFERRED to a follow-up SPEC (user decision 2026-07-13; see §E Exclusions).
2. **Dedicated specialist agent**: a new `e2e-tester` retained agent (the retired version delegated to manager-develop; no dedicated agent ever existed).
3. **Token-minimization as a first-class requirement**: CLI-output-first execution, bounded output tails, file-redirect for verbose logs, MCP only when CLI cannot do the job.

Distribution is via the TEMPLATE SOURCE (`internal/template/templates/`) per the Template-First Rule, with catalog.yaml and CI-guard count-constant reconciliation.

### Measured baseline (verified 2026-07-13, this tree)

| Surface | Current state | Post-SPEC state |
|---------|---------------|-----------------|
| `.claude/skills/moai/SKILL.md` Priority 1 router (both trees) | 14 subcommand rows, **no** `e2e` row (`grep -c` for `e2e`: 0) | 15 rows, `e2e` row restored |
| `CLAUDE.md` §3 `Subcommands:` line (both trees) | 13 tokens, no `e2e` | 14 tokens incl. `e2e` |
| `CLAUDE.md` §4 agent catalog (both trees) | "exactly **10 retained agents** (9 MoAI-custom + 1 `Explore`)" | 11 retained (10 MoAI-custom + 1 `Explore`) |
| `.claude/agents/moai/*.md` (both trees) | 9 files | 10 files (+ `e2e-tester.md`) |
| `.claude/commands/moai/` | 13 files per tree (local `.md`, template `.md.tmpl`) | 14 per tree (+ `e2e`) |
| `internal/template/catalog_tier_audit_test.go` | `expectedSkillCount = 28` (L158), `expectedAgentCount = 9` (L231) | 28 (UNCHANGED — see note), 10 |
| `internal/template/catalog_loader_test.go` | `expectedTotal = 37` (L55) | 38 |
| `internal/template/catalog.yaml` `core.agents` | 8 entries (builder-harness lives under `harness_generated.agents`) | 9 entries (catalog-wide agent total 9 → 10; `expectedTotal` 37 → 38 math unchanged) |

> **Count note**: the workflow skill lives INSIDE the `moai` skill directory (`.claude/skills/moai/workflows/e2e.md`), so `expectedSkillCount` (top-level skill directories) does NOT change. Only the agent-count constants move.

---

## §B Requirements (GEARS)

### Group A — Project-type detection & toolchain selection

- **REQ-E2E-001** (Ubiquitous): The e2e workflow shall auto-detect the project platform type (`web` / `mobile` / `desktop` / `mixed`) from project markers before any toolchain selection (marker matrix: design.md §B).
- **REQ-E2E-002** (Event-driven): **When** project-type detection completes, the workflow shall map the detected type to the default toolchain per the platform-toolchain matrix (web → Playwright CLI; mobile → Maestro; desktop → Playwright-Electron for Electron apps, WebdriverIO + `@wdio/tauri-service` for Tauri apps).
- **REQ-E2E-003** (Event-driven): **When** more than one platform type is detected (`mixed`), the workflow shall enumerate per-platform toolchain candidates and surface the selection through the ORCHESTRATOR's AskUserQuestion channel (one question per platform surface, recommended option first).
- **REQ-E2E-004** (Ubiquitous): The workflow shall route ALL user-facing toolchain/journey selection through the orchestrator's AskUserQuestion channel; the e2e-tester agent shall obtain selections exclusively via its spawn prompt.
- **REQ-E2E-005** (Capability gate): **Where** a `--tool <name>` flag is provided, the workflow shall bypass the selection question and use the named toolchain directly.
- **REQ-E2E-006** (Event-driven): **When** the selected toolchain is detected as not installed (version probe fails), the workflow shall surface the exact install command(s) and, upon user approval via the orchestrator, perform the installation and re-verify with a version probe before proceeding.
- **REQ-E2E-007** (Event-driven): **When** no e2e-able surface is detected (e.g., a pure library with no web/mobile/desktop entry point), the workflow shall report "no e2e target detected" with the marker evidence and exit gracefully without creating test artifacts. A detected `desktop-native` surface (non-Electron/non-Tauri) routes to this same graceful branch with a deferral notice (automation deferred to a follow-up SPEC — see §E Exclusions).

### Group B — Token minimization (first-class)

- **REQ-E2E-100** (Ubiquitous): The subsystem shall execute CLI-output-first: every capability achievable via CLI invocation shall use the CLI path; MCP tools shall be used ONLY for capabilities the selected CLI cannot provide (the per-capability classification lives in the workflow's tool matrix).
- **REQ-E2E-101** (Ubiquitous): The e2e-tester shall bound in-context command output to exit code + bounded tail (≤50 lines OR ≤2KB, whichever is smaller), redirecting verbose output to files under the run-artifacts directory with citable paths (file-redirect contract, `agent-common-protocol.md` §Parallel Execution).
- **REQ-E2E-102** (State-driven): **While** an MCP-backed tool is active, the e2e-tester shall prefer snapshot/batch reads (accessibility tree, DOM snapshot, aggregated trace insights) over per-element round-trips.
- **REQ-E2E-103** (Event-driven): **When** a test run produces reports (HTML report, traces, screenshots, recordings), the workflow shall persist them under project-local `e2e/` artifact directories and cite paths in the report — never inline the artifact content.
- **REQ-E2E-104** (Capability gate): **Where** recording is requested (`--record`), the workflow shall use the selected toolchain's native trace/recording facility (Playwright trace, Maestro recording, agent-browser trace) rather than MCP screenshot loops.
- **REQ-E2E-105** (Unwanted): The subsystem shall not require any MCP server as a hard dependency; every default platform path shall be fully executable with CLI-only tools.

### Group C — Deliverable artifacts

- **REQ-E2E-200** (Ubiquitous): The template source shall gain a thin command wrapper at `internal/template/templates/.claude/commands/moai/e2e.md.tmpl` conforming to the Thin Command Pattern: frontmatter with `description` + `allowed-tools` (CSV string), body <20 non-empty lines containing a `Skill(` invocation or `subagent` delegation directive (enforced by `TestCommandsThinPattern`).
- **REQ-E2E-201** (Ubiquitous): The template source shall gain a workflow skill at `internal/template/templates/.claude/skills/moai/workflows/e2e.md` with `user-invocable: false` frontmatter and the phase structure: Detection → Selection → Journey Mapping → Script Creation → Execution → Recording (optional) → Report.
- **REQ-E2E-202** (Ubiquitous): The template source shall gain a dedicated agent at `internal/template/templates/.claude/agents/moai/e2e-tester.md` conforming to retained-agent frontmatter conventions (name, description with "NOT for:" clause, `tools:` CSV, `model: inherit`, `effort`, `color`, `permissionMode`, `memory`, `skills:` YAML array ≤2 entries).
- **REQ-E2E-203** (Unwanted): The e2e-tester agent shall not list `Agent` in its `tools:` CSV and shall not list or invoke `AskUserQuestion`; when required inputs are missing, it shall return the structured `## Missing Inputs` blocker report (`agent-common-protocol.md` §Blocker Report Format).
- **REQ-E2E-204** (Ubiquitous): The workflow skill body shall delegate detection, script creation, and execution phases to the `e2e-tester` subagent BY NAME, and the agent file shall exist at the path that delegation resolves to, in BOTH trees (cross-file reachability, not token presence).

### Group D — Router & catalog reachability

- **REQ-E2E-300** (Ubiquitous): The `/moai` SKILL.md Intent Router Priority 1 subcommand list shall regain an `**e2e**` row routing to the e2e workflow, in BOTH trees.
- **REQ-E2E-301** (Ubiquitous): The `moai` SKILL.md frontmatter `description` subcommand enumeration AND the `CLAUDE.md` §3 `Subcommands:` line shall include `e2e`, in BOTH trees.
- **REQ-E2E-302** (Ubiquitous): The retained-agent count-literal surface shall be updated consistently in BOTH trees. A TRULY repo-wide grep (`'10 retained agents|9 MoAI-custom|10-agent'` over CLAUDE.md, README*, .claude, internal/template, internal/web, docs-site; re-executed 2026-07-13) structures the surface into rings. **Ring 1 — doctrine tree (12 files / 24 sites**; content-token anchored — line numbers drift between trees):
  - `CLAUDE.md` §4 (3 sites each tree): the count text ("exactly 10 retained agents (9 MoAI-custom + 1 Explore)" → "exactly 11 retained agents (10 MoAI-custom + 1 Explore)"), the "flat-hierarchy 10-agent consolidation rationale" sentence (→ 11-agent), the "one of the 10 retained agents above" sentence (→ 11); PLUS a catalog table row and a Selection Decision Tree entry for `e2e-tester` (appended as entry 12; manager-design remains entry 11)
  - `.claude/rules/moai/development/agent-authoring.md` (3 sites each tree): "exactly 10 retained agents" / "10-agent retention ceiling" family → 11
  - `.claude/rules/moai/development/agent-patterns.md` (3 sites each tree): incl. the MoAI-custom name enumeration (gains `e2e-tester`) and "all 9 MoAI-custom agents" → 10
  - `.claude/rules/moai/development/model-policy.md` (1 site each tree): "All 9 MoAI-custom retained agents … 10-agent catalog" → 10 / 11-agent
  - `.claude/rules/moai/workflow/spec-workflow.md` (1 site each tree): "exactly 10 retained agents (named list)" → 11 + named list gains `e2e-tester`
  - `.claude/agents/moai/manager-design.md` (1 site each tree): "10 retained agents" → 11

  **Ring 2 — template-distributed skill modules (4 template files / 8 sites)**: `internal/template/templates/.claude/skills/moai-foundation-core/SKILL.md` (1 site), `…/moai-foundation-core/modules/agents-reference.md` (3 sites — this file IS the extended agent-catalog reference and gains an `e2e-tester` table row, not just count edits), `…/moai-foundation-core/modules/INDEX.md` (3 sites), `…/moai-foundation-quality/SKILL.md` (1 site). OBSERVED PRE-EXISTING DRIFT (recorded baseline, audit iter-2 D13): the 4 LOCAL siblings are TWO generations stale — they carry the "8 retained agents / 7 MoAI-custom / 8-agent" family (7 sites measured) and ZERO 10/9-family matches. Run-phase normalizes BOTH trees to the 11/10 target content.

  **Ring 3 — repo-root user docs**: `README.md` (3 sites measured). `README.ko.md` / `README.ja.md` / `README.zh.md` carry NO ERE-family matches (measured 2026-07-13); run-phase reviews their locale-language agent-count claims manually and updates any found (4-locale parity obligation).

  **Ring 4 — Go display surface**: owned by REQ-E2E-305 below.

  **Deferred ring — docs-site**: count-literal corrections across the 10-file docs-site inventory are EXPLICITLY deferred with the docs-site follow-up (§E Exclusions).

  After the update, invariance greps over the touched files shall return 0 matches: **CMD-019-INV** (19-file in-scope surface, 10/9-family; measured baseline 38) AND **CMD-019-INV-B** (local skill siblings, 8/7-era family; measured baseline 7) — acceptance.md § Executable Command Block.
- **REQ-E2E-303** (Capability gate): **Where** natural-language input expresses e2e-testing intent in any `conversation_language`, the Priority 3 semantic classification shall route to the e2e workflow (cue exemplars added to the P3 list; semantic, not literal-match).
- **REQ-E2E-304** (Ubiquitous): `internal/template/catalog.yaml` `core.agents` shall gain an `e2e-tester` entry (name/tier/path/version authored manually; hash computed by `make build` → `gen-catalog-hashes --all`). NOTE: the hash generator only UPDATES existing entries — it does not create them (verified in `gen-catalog-hashes.go`).
- **REQ-E2E-305** (Ubiquitous): The Go tier-profile display surface shall gain an `e2e-tester` row so the agent renders in the `moai web` model-policy preview: `tierProfileAgentOrder` + its per-agent tier-profile entries in `internal/template/model_policy.go`, with the display pins reconciled in the SAME commit — `model_policy_test.go` length pin 10 → 11, the 60-cell assertion → 66 (2 plans × 11 agents × 3 tiers), the "10 retained agents" test comments → 11; `internal/web/modelpolicy_test.go` per-plan row expectation 10 → 11 (and its "10 retained agents × 3 tiers" comment). This is a DATA-ROW + test-pin change under the C-7 carve-out (no new runtime logic). Rationale (audit iter-2 D13): the display list is hard-coded, so NO test fails on omission — without this REQ, e2e-tester would be silently absent from the preview (doctrine-vs-display drift).

### Group E — Distribution & CI reconciliation

- **REQ-E2E-400** (Ubiquitous): All new artifacts shall be authored in the TEMPLATE tree FIRST (`internal/template/templates/`), then synced to the local tree, per the Template-First Rule (CLAUDE.local.md §2 [HARD]).
- **REQ-E2E-401** (Ubiquitous): CI-guard count constants shall be reconciled with provenance comments in the same commit as the artifact additions: `expectedAgentCount` 9 → 10 (`catalog_tier_audit_test.go`), `expectedTotal` 37 → 38 (`catalog_loader_test.go`). `expectedSkillCount` stays 28.
- **REQ-E2E-402** (Ubiquitous): The new command file shall pass `TestCommandsThinPattern` and `TestCommandsFrontmatterConsistency`; the new agent shall pass `TestAgentFrontmatterAudit`, `TestAllAgentsInCatalog`, and `TestCatalogReferencesValid`; `go test ./internal/template/...` shall exit 0.
- **REQ-E2E-403** (Unwanted): Template-tree content shall not contain internal SPEC IDs, REQ/AC tokens, audit citations, internal work dates, commit SHAs, or CLAUDE.local references (template neutrality, CLAUDE.local.md §25) — the existing neutrality CI guard shall pass unchanged.
- **REQ-E2E-404** (Event-driven): **When** template authoring completes, `make build` shall be run so the embedded FS (`//go:embed all:templates`) and catalog hashes regenerate; the built binary shall carry the three new artifacts.
- **REQ-E2E-405** (Ubiquitous): The project-marker detection matrix shall treat all supported project ecosystems equally — no language/framework is presented as "primary" while others are "planned" (16-language neutrality, CLAUDE.local.md §15).

### Group F — Boundaries

- **REQ-E2E-500** (Unwanted): The revival shall not resurrect any of the 7 retired design-pack skills (`moai-domain-ideation`, `moai-domain-research`, `moai-domain-brand-design`, `moai-domain-copywriting`, `moai-domain-design-handoff`, `moai-workflow-design`, `moai-workflow-gan-loop`) nor the other 4 retired subcommands (design/brain/coverage/security).
- **REQ-E2E-501** (Ubiquitous): The e2e-tester shall operate within the subagent boundary: blocker reports instead of prompts, results returned to the orchestrator, no nested agent spawning.

> REQ-E2E-502 (desktop-native opt-in fallback) was REMOVED at v0.1.1 — the user deferred native-desktop automation to a follow-up SPEC. The `desktop-native` detection classification remains and routes to the REQ-E2E-007 graceful branch with a deferral notice.

---

## §C Non-functional Constraints

- **C-1** [HARD] Template-First: no file lands in local `.claude/` without a template-tree sibling authored first.
- **C-2** [HARD] Thin Command Pattern: command body <20 non-empty lines (`TestCommandsThinPattern` R3).
- **C-3** [HARD] Subagent boundary: e2e-tester never prompts the user (Frozen-zone rule).
- **C-4** [HARD] Template neutrality: zero internal-development traces in template content (existing CI guard is the gate).
- **C-5** [HARD] All CI-guard tests, `go test ./...`, and `golangci-lint run` green at close.
- **C-6** Token economy: the workflow skill body itself stays lean (~≤450 lines, at parity with the retired baseline despite 3× platform scope — achieved by matrix-driven structure over per-tool prose duplication).
- **C-7** No new Go runtime LOGIC: this SPEC ships markdown/YAML template artifacts + test-constant updates + ONE Go data-row addition (the REQ-E2E-305 tier-profile display row in `model_policy.go` — carve-out added at v0.1.2 so the display surface tracks the catalog). The `moai` Go binary gains no new subcommand and no new control flow (the `/moai e2e` surface is slash-command-routed, exactly like the retired version).

---

## §D Acceptance Criteria Map

The full AC matrix (28 ACs), Given-When-Then scenarios, edge cases, and Definition of Done live in `acceptance.md`. Mapping summary:

| REQ group | ACs |
|-----------|-----|
| A (detection/selection) | AC-E2E-001 … AC-E2E-006 |
| B (token minimization) | AC-E2E-007 … AC-E2E-011 |
| C (artifacts) | AC-E2E-012 … AC-E2E-016 |
| D (router/catalog reachability) | AC-E2E-017 … AC-E2E-021, AC-E2E-028 |
| E (distribution/CI) | AC-E2E-022 … AC-E2E-025, AC-E2E-027 |
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

### Out of Scope — docs-site 4-locale documentation and count-literal corrections

- adk.mo.ai.kr documentation for the revived subcommand is deferred to a follow-up (user decision 2026-07-13).
- The deferral EXPLICITLY covers the docs-site stale agent-count literals as well (audit iter-2 D13(d)): repo-wide grep (2026-07-13) inventories 10 docs-site files carrying the "10 retained agents|9 MoAI-custom|10-agent" family — content/en: advanced/builder-agents.md, advanced/claude-md-guide.md, claude-code/agentic/sub-agents.md, core-concepts/harness-engineering.md, core-concepts/what-is-moai-adk.md, getting-started/faq.md, multi-llm/model-policy.md, workflow-commands/moai-harness.md; content/ko: claude-code/agentic/sub-agents.md; layouts/index.html. The follow-up owns updating these to the 11/10 catalog under the 4-locale parity obligation.

### Out of Scope — Native-desktop (non-Electron/non-Tauri) automation fallback

- OS-level accessibility / computer-use automation for native desktop toolkits (pure macOS apps, WinUI, Qt/GTK) is DEFERRED to a follow-up SPEC (user decision 2026-07-13; former REQ-E2E-502 removed at v0.1.1). Detection still classifies `desktop-native`; the workflow reports the deferral and exits via the REQ-E2E-007 graceful branch.

### Out of Scope — Device-farm / cloud-device execution

- Maestro Cloud, BrowserStack, Sauce Labs, and other remote device farms are not integrated; local simulator/emulator/device execution only.
