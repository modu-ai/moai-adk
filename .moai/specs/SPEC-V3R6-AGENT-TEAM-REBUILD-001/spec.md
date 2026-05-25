---
id: SPEC-V3R6-AGENT-TEAM-REBUILD-001
title: "Agent catalog Anthropic 2026 alignment: 17→5-8 consolidation + hook enforcement"
version: "0.1.0"
status: draft
created: 2026-05-25
updated: 2026-05-25
author: manager-spec
priority: P0
phase: "v3.0.0"
module: ".claude/agents + .claude/skills/moai/workflows + .claude/rules/moai/* + internal/template/templates/*"
lifecycle: spec-anchored
tags: "agent-team-rebuild, consolidation, anthropic-2026-alignment, hook-enforcement, gears, sprint-10-cohort-X, supersedes"
depends_on: [SPEC-V3R6-WORKFLOW-ORCHESTRATION-FIX-001, SPEC-V3R6-GEARS-MIGRATION-001, SPEC-V3R6-FOUNDATION-CORE-GEARS-ALIGN-001]
supersedes: [SPEC-V3R6-WORKFLOW-ORCHESTRATION-FIX-001]
tier: L
---

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-05-25 | manager-spec | Initial Tier L SPEC authored from 3 deep audit reports (17-agent SRP audit, workflow phase ownership audit, Anthropic 2026 verbatim audit). Supersedes SPEC-V3R6-WORKFLOW-ORCHESTRATION-FIX-001 (in plan-phase iter-2 PASS-WITH-DEBT 0.8625) — Audit 3 findings invalidate the original 17-phase-fix scope because catalog inflation (17 agents vs Anthropic 3-5 ceiling) + hierarchical fiction (`manager-strategy → manager-develop` violates "subagents cannot spawn subagents") + 12 phantom agents (0 invocations across 4 SPEC sessions) require fundamental architecture pivot to 17→5-8 consolidation + hook-based enforcement. 20 REQ-ATR with ≥80% GEARS notation self-dogfood. |

---

## §A — Goals

Realign the MoAI agent catalog with Anthropic 2026 best-practice published guidance by:
(a) **consolidating** 17 agents → 8 retained agents (12 archived) per Anthropic's "define when keep spawning the same kind of worker" criterion and the official 3-built-in-subagent ceiling (Explore, Plan, general-purpose);
(b) **eliminating hierarchical fiction** — the `manager-strategy → manager-develop` chain pattern is architecturally impossible per Anthropic verbatim "Subagents cannot spawn other subagents. If your workflow requires nested delegation, use Skills or chain subagents from the main conversation"; all multi-agent orchestration moves to the main session (the MoAI orchestrator);
(c) **replacing 6 domain-expert agents** with the Anthropic-recommended per-spawn-prompt specialization pattern (`Agent(general-purpose, model: opus, tools: <domain whitelist>)` injected at delegation time, not embedded in agent definition files);
(d) **converting orchestrator-discipline into mechanically-unbypassable hook enforcement** — PostToolUse hook for Status Transition Ownership Matrix, Stop hook for sync-phase quality gates (lint + test + coverage delta), TaskCompleted hook for Agent Teams per-AC PASS evidence;
(e) **superseding** SPEC-V3R6-WORKFLOW-ORCHESTRATION-FIX-001 (in plan-phase iter-2) — the predecessor SPEC's 17-phase-restoration scope is invalidated by Audit 3's discovery that the underlying catalog itself violates Anthropic's published architecture; restoring phantom agents would compound the architectural fiction rather than resolve it.

The pivot is constitutional in scope (affects every `Agent()` invocation in the project and every `/moai plan|run|sync` workflow) and is therefore classified Tier L per `.claude/rules/moai/workflow/spec-workflow.md` § SPEC Complexity Tier.

---

## §B — Background

### §B.1 Audit 3 verbatim findings (Anthropic 2026 alignment)

Three deep audit reports were conducted during the immediate prior session turn to assess the MoAI agent catalog against the recent (2026-mid) Anthropic published guidance. The findings are research-grade verbatim citations from official Anthropic documentation, the Claude Agent SDK reference, and the Agent Teams documentation:

**Finding A1 (CRITICAL — Catalog inflation 2-5× over Anthropic ceiling)**: Anthropic ships exactly **3 built-in subagents** (Explore, Plan, general-purpose). The Agent Teams documentation states verbatim: *"Start with 3-5 teammates for most workflows. This balances parallel work with manageable coordination."* MoAI's current catalog ships 17 custom agents (8 manager-* + 6 expert-* + 1 builder-* + 2 evaluator/audit), placing the project 2-5× above the recommended ceiling. The cost is observable: 12 of 17 agents recorded 0 invocations across the recent 4-SPEC cohort.

**Finding A2 (CRITICAL — Hierarchical fiction architecturally impossible)**: Anthropic Sub-agents documentation states verbatim: *"Subagents cannot spawn other subagents. If your workflow requires nested delegation, use Skills or chain subagents from the main conversation."* MoAI's `manager-strategy → manager-develop` hierarchical chain pattern — as documented in `SPEC-V3R6-WORKFLOW-ORCHESTRATION-FIX-001` §B.2 R4 — is therefore not merely under-utilized; it is architecturally impossible to execute in the current Claude Code runtime. The predecessor SPEC's "restore the chain" remediation strategy would fail at runtime regardless of declarative rule changes.

**Finding A3 (HIGH — Phantom agent criterion failure)**: Anthropic best-practices guidance states verbatim: *"Define a custom subagent when you keep spawning the same kind of worker with the same instructions."* 12 of MoAI's 17 agents fail this criterion outright: `manager-strategy`, `manager-quality`, `manager-brain`, `manager-project`, `claude-code-guide`, `researcher`, and 6 `expert-*` agents all recorded 0 invocations across the recent 4 SPEC sessions. The retained 5 (`manager-spec`, `manager-develop`, `manager-docs`, `plan-auditor`, `manager-git`) plus 2 dormant-but-justified (`evaluator-active`, `builder-harness`) and 1 Anthropic built-in (`Explore`) form the consolidation target of 8.

**Finding A4 (HIGH — Coding-task parallelism caveat)**: Anthropic verbatim: *"most coding tasks involve fewer truly parallelizable tasks than research, and LLM agents are not yet great at coordinating and delegating to other agents in real time."* The MoAI workflow's run-phase coding tasks SHOULD therefore remain single-agent (`manager-develop` standalone with cycle_type), not decomposed into manager-strategy chain. Parallel multi-spawn applies primarily to research phase (Explore + analyst + architect role profiles), not to implementation.

**Finding A5 (MEDIUM — Domain expertise via spawn-prompt, not via agent file)**: Anthropic's recommended pattern for domain expertise is per-spawn parameter injection: `Agent(subagent_type: "general-purpose", model: "opus", tools: ["Read", "Write", "Edit", "Bash"], prompt: "<domain-specific instructions>")`. The 6 `expert-*` agents (backend, frontend, security, devops, performance, refactoring) duplicate this pattern as static agent files, which (a) violates the "define when keep spawning" criterion (0 invocations confirms infrequent use) and (b) traps domain knowledge inside individual agent definitions rather than the active conversation context.

**Finding A6 (MEDIUM — Hook-based enforcement available)**: Anthropic hook documentation (Stop, PostToolUse, SubagentStop, TaskCompleted) provides the canonical mechanism for converting orchestrator-discipline into mechanically-unbypassable enforcement. Quality gates (lint + test + coverage delta), Status Transition Ownership, and per-AC PASS verification are all hook-enforceable today.

### §B.2 Predecessor SPEC supersedence rationale

`SPEC-V3R6-WORKFLOW-ORCHESTRATION-FIX-001` (currently at plan-phase iter-2 PASS-WITH-DEBT 0.8625, commits `97e42eaf7` + `22b4dc691`) sought to **restore** 17 silently-skipped phases. Audit 3 reveals that the restoration target itself is architecturally invalid:

- Restoring `manager-strategy → manager-develop` chain (predecessor REQ-WOF-003 + REQ-WOF-006) violates Anthropic verbatim "subagents cannot spawn subagents". The chain cannot execute at runtime.
- Restoring sync-phase 3 quality specialists (predecessor REQ-WOF-002 + REQ-WOF-007) reinforces the phantom-agent pattern. The specialists' work is more reliably enforced via Stop hooks (lint + test + coverage delta) than via orchestrator-discipline manager-quality/expert-security spawn calls that historically went unfulfilled.
- Restoring Producer-Reviewer cycle (predecessor REQ-WOF-010) with `evaluator-active` is sound, but is one of only 3 phases worth restoring out of the original 17. The remaining 14 represent architectural debt to be retired, not restored.

Per Status Transition Ownership Matrix in `.claude/rules/moai/development/spec-frontmatter-schema.md`, the `* → superseded` transition is owned by manager-spec (when authoring the new superseding SPEC). This SPEC's run-phase M6 transitions the predecessor SPEC's frontmatter `status` to `superseded` with HISTORY entry citing the architectural-pivot rationale above.

### §B.3 Retain-vs-archive matrix (8 retain + 12 archive)

**Retain (8 agents)** — each justified by recent invocation frequency AND Anthropic alignment:

| # | Agent | Recent invocation freq | Anthropic alignment | Rationale |
|---|-------|------------------------|---------------------|-----------|
| 1 | `manager-spec` | 100% of recent SPEC sessions | Aligned (plan-phase canonical owner) | Absorbs manager-strategy planning role; planning IS strategy under Anthropic 4-phase Step 2 |
| 2 | `manager-develop` | 100% of recent SPEC sessions | Aligned (single-agent coding task per Finding A4) | Absorbs cycle_type=ddd / cycle_type=tdd / NEW cycle_type=autofix (R-12 Audit 1 proposal) |
| 3 | `manager-docs` | 75% of recent SPEC sessions | Aligned (sync-phase + frontmatter ownership per Matrix) | Absorbs manager-project (project docs ARE docs); manager-brain workflow becomes Explore + manager-spec sequence |
| 4 | `plan-auditor` | Active for every plan-phase iter | Aligned (Phase 0.5 independent audit per harness Producer-Reviewer pattern) | Bias-prevention auditor; remains independent |
| 5 | `evaluator-active` | Currently dormant (0 invoc) but justified by harness `thorough` mode | Aligned (Producer-Reviewer pattern, harness pattern #4) | Activate per harness level — not a phantom; legitimate dormancy until thorough mode invoked |
| 6 | `builder-harness` | Used in skill/agent authoring (low freq but recurring) | Aligned (meta-builder for catalog evolution); Audit 1 SRP score 0.85 | Retained for skill / agent / plugin authoring; powers `/moai project` Phase 5+ interview |
| 7 | `manager-git` | Active in Tier L OR `--pr` flag PR sessions | Aligned (manager-git is delegation surface for Hybrid Trunk PR ops) | Clarified usage per REQ-ATR-020 — Tier L OR `--pr` flag triggers manager-git; default Tier S/M is main-direct push |
| 8 | `Explore` (Anthropic built-in) | NEW retention | Aligned (Anthropic-shipped, no MoAI customization needed) | Replaces custom `claude-code-guide` + `researcher` for read-only investigations |

**Archive (12 agents)** — each archived with rationale:

| # | Archive | Archive type | Replacement pattern | Anthropic citation |
|---|---------|--------------|---------------------|---------------------|
| 1 | `manager-strategy` | **Critical archive** (architecturally impossible) | Absorb into `manager-spec` (planning IS strategy per Anthropic 4-phase Step 2) | "Subagents cannot spawn other subagents" |
| 2 | `manager-quality` | **Phantom archive** (0 invoc) | Stop hook enforcement (lint + test + coverage delta) OR `/moai gate` skill | "Define when keep spawning" criterion failure |
| 3 | `manager-brain` | **Phantom archive** (0 invoc) | Explore + manager-spec sequence in `/moai brain` workflow | "Define when keep spawning" criterion failure |
| 4 | `manager-project` | **Phantom archive** (0 invoc) | Absorb into `manager-docs` (project docs ARE docs) | "Define when keep spawning" criterion failure |
| 5 | `claude-code-guide` | **Phantom archive** (0 invoc) | Use Anthropic built-in `Explore` for upstream investigation | "3 built-in subagents" ceiling |
| 6 | `researcher` | **Phantom archive** (0 invoc) | Use Anthropic built-in `Explore` for auto-research | "3 built-in subagents" ceiling |
| 7 | `expert-backend` | **Domain-expert archive** (per-spawn-prompt replacement) | `Agent(general-purpose, model: opus, tools: <backend whitelist>, prompt: <domain instructions>)` | Finding A5 spawn-prompt specialization pattern |
| 8 | `expert-frontend` | **Domain-expert archive** | Same pattern as #7 with frontend whitelist | Finding A5 |
| 9 | `expert-security` | **Domain-expert archive** | Stop hook (dependency manifest audit on `go.mod` changes) + per-spawn for code review | Finding A5 + Finding A6 |
| 10 | `expert-devops` | **Domain-expert archive** | Same pattern as #7 with devops whitelist | Finding A5 |
| 11 | `expert-performance` | **Domain-expert archive** | Same pattern as #7 with performance whitelist | Finding A5 |
| 12 | `expert-refactoring` | **Domain-expert archive** | Same pattern as #7 with refactoring whitelist | Finding A5 |

### §B.4 Hook-based enforcement (replaces specialist coverage)

Per Finding A6, the following Anthropic-canonical hook patterns convert orchestrator-discipline into mechanically-unbypassable enforcement:

- **PostToolUse hook** (`status-transition-ownership.sh`): on Write/Edit to any `.moai/specs/SPEC-*/spec.md` or `plan.md` or `acceptance.md` body content (frontmatter excluded), verify that the invoking agent name matches the canonical owner per the Status Transition Ownership Matrix. Reject with exit 2 if mismatch (e.g., `manager-docs` attempts Write on `spec.md` body — only manager-spec can author SPEC body content). Replaces archived `manager-quality` declarative audit.
- **Stop hook** (`sync-phase-quality-gate.sh`): on sync-phase commit completion, verify (a) lint exit 0 (`golangci-lint run --timeout=2m`), (b) test suite exit 0 (`go test ./...`), (c) coverage delta non-negative for changed files. Block commit via hook exit 2 on failure. Replaces archived `manager-quality` + `expert-security` declarative coverage.
- **TaskCompleted hook** (team mode only, `team-ac-verify.sh`): on team-mode Agent Teams TaskCompleted event, verify per-AC PASS evidence file exists (`.moai/specs/SPEC-*/ac-evidence/AC-<ID>.txt`). Exit 2 = reject completion. Activates only under `harness: thorough` + team mode prerequisites.

---

## §C — Scope

### §C.1 In-scope file inventory

This SPEC modifies the following 5 file categories:

**Tier 1 — Agent files (17 total: 8 retain frontmatter-only refinement + 12 archive to backups + 1 NEW per-spawn pattern documentation)**:

Retain (8 files, frontmatter + minor refinement only):
- `.claude/agents/core/manager-spec.md` — absorbs manager-strategy planning role; description + tools + NOT-for refinement
- `.claude/agents/core/manager-develop.md` — frontmatter + cycle_type=autofix addition + NOT-for refinement
- `.claude/agents/core/manager-docs.md` — absorbs manager-project; description + NOT-for refinement
- `.claude/agents/core/manager-git.md` — usage clarification (Tier L OR `--pr` flag only); NOT-for refinement
- `.claude/agents/meta/plan-auditor.md` — frontmatter + NOT-for refinement (no body changes)
- `.claude/agents/meta/evaluator-active.md` — frontmatter + harness-level invocation surface (no body changes)
- `.claude/agents/meta/builder-harness.md` — frontmatter + NOT-for refinement (no body changes)
- (Anthropic built-in `Explore` — no MoAI file; documented in `.claude/rules/moai/development/agent-patterns.md` as canonical read-only-investigation agent)

Archive (12 files moved to `.moai/backups/agent-archive-2026-05-25/`):
- `.claude/agents/core/manager-strategy.md` → backup
- `.claude/agents/core/manager-quality.md` → backup
- `.claude/agents/core/manager-brain.md` → backup
- `.claude/agents/core/manager-project.md` → backup
- `.claude/agents/meta/claude-code-guide.md` → backup
- `.claude/agents/agency/researcher.md` (if exists) → backup
- `.claude/agents/expert/expert-backend.md` → backup
- `.claude/agents/expert/expert-frontend.md` → backup
- `.claude/agents/expert/expert-security.md` → backup
- `.claude/agents/expert/expert-devops.md` → backup
- `.claude/agents/expert/expert-performance.md` → backup
- `.claude/agents/expert/expert-refactoring.md` → backup

**Tier 2 — Workflow router skills (3 files, owner declarations)**:
- `.claude/skills/moai/workflows/plan.md` — phase owner declarations (Explore + manager-spec); remove manager-strategy chain references
- `.claude/skills/moai/workflows/run.md` — phase owner declarations (manager-develop single-agent per Finding A4); remove manager-strategy chain; add Phase 0.95 Mode Selection per REQ-ATR-008; add cycle_type=autofix per REQ-ATR-012
- `.claude/skills/moai/workflows/sync.md` — phase owner declarations (manager-docs single-agent); replace manager-quality/expert-security/manager-develop-coverage spawn references with Stop hook reference

**Tier 3 — Rule files (~10 files, owner-declaration + archived-agent cross-references)**:
- `.claude/rules/moai/workflow/spec-workflow.md` — Phase Overview table agent column updates; archive table reference
- `.claude/rules/moai/development/agent-patterns.md` — domain-expert spawn-prompt pattern documentation (per Finding A5); `Explore` canonical read-only-investigation agent documentation; deprecate manager-strategy chain pattern
- `.claude/rules/moai/development/agent-authoring.md` — per-spawn specialization vs static agent file decision tree
- `.claude/rules/moai/development/manager-develop-prompt-template.md` — cycle_type=autofix mode reference
- `.claude/rules/moai/development/spec-frontmatter-schema.md` — Status Transition Ownership Matrix archived-agent references purge
- `.claude/rules/moai/core/agent-common-protocol.md` — orchestrator obligations (PostToolUse + Stop + TaskCompleted hook invocation surface) reference
- `.claude/rules/moai/workflow/orchestration-mode-selection.md` (NEW) — 5-mode autonomous selection rule per REQ-ATR-008
- `.claude/rules/moai/workflow/archived-agent-rejection.md` (NEW) — ARCHIVED_AGENT_REJECTED error specification + migration table per REQ-ATR-016
- `CLAUDE.md` Section 4 (Agent Catalog) — 17→8 retention list + archive table reference
- `CLAUDE.local.md` §23 (Hybrid Trunk) — manager-git PR doctrine reconciliation per REQ-ATR-020

**Tier 4 — Hook scripts (3 NEW files)**:
- `.claude/hooks/moai/status-transition-ownership.sh` (PostToolUse) — REQ-ATR-009 cross-reference
- `.claude/hooks/moai/sync-phase-quality-gate.sh` (Stop) — REQ-ATR-009 sync-phase
- `.claude/hooks/moai/team-ac-verify.sh` (TaskCompleted, team mode) — optional, dormant by default

**Tier 5 — SPEC artifacts (this SPEC, 5 files)**:
- `.moai/specs/SPEC-V3R6-AGENT-TEAM-REBUILD-001/spec.md` (this file)
- `.moai/specs/SPEC-V3R6-AGENT-TEAM-REBUILD-001/plan.md`
- `.moai/specs/SPEC-V3R6-AGENT-TEAM-REBUILD-001/acceptance.md`
- `.moai/specs/SPEC-V3R6-AGENT-TEAM-REBUILD-001/design.md`
- `.moai/specs/SPEC-V3R6-AGENT-TEAM-REBUILD-001/research.md`

**Tier 6 — Template parity (mirror to `internal/template/templates/`, ~13 files)**:
- `internal/template/templates/.claude/agents/{core,meta,expert}/*.md` — byte-for-byte sync with local retained 8 agents (archived 12 removed from template)
- `internal/template/templates/.claude/skills/moai/workflows/{plan,run,sync}.md` — byte-for-byte sync
- `internal/template/templates/.claude/rules/moai/**/*.md` — byte-for-byte sync of modified rule files
- `internal/template/templates/.claude/hooks/moai/*.sh` — byte-for-byte sync of new hook scripts

**Tier 7 — Supersedence of predecessor SPEC (1 file modification)**:
- `.moai/specs/SPEC-V3R6-WORKFLOW-ORCHESTRATION-FIX-001/spec.md` — frontmatter `status → superseded` + HISTORY v0.1.2 entry citing architectural-pivot rationale (REQ-ATR-006)

**Tier 8 — Documentation (1 file modification)**:
- `NOTICE.md` — retention rationale + archive dates + Anthropic 2026 alignment citation per Apache 2.0 attribution discipline (REQ-ATR-019)

**Estimated scope total**: ~50-60 files (this SPEC's 5 artifacts + 17 agent files + 3 workflow skills + ~10 rule files + 3 hook scripts + ~13 template mirrors + predecessor SPEC supersedence + NOTICE.md). LOC equivalent: ~1500-2000 lines markdown edits + ~300-500 lines new hook script shell code.

### §C.2 Agent retention matrix

See §B.3 above for the 8-retain / 12-archive matrix with per-agent rationale and Anthropic citations.

### §C.3 Out of Scope

#### §C.3.1 Out of Scope — Retained agent body re-authoring

The 8 retained agents' body content (Markdown beyond the YAML frontmatter and the immediate description-refinement scope) is NOT re-authored in this SPEC. Only frontmatter (description, tools, NOT-for clauses) and minor scoped refinement (e.g., cycle_type=autofix mode addition for manager-develop, manager-project absorption notes for manager-docs) are within scope. Full body re-authoring is deferred to a follow-up SPEC if/when the LOC ≤ 500 anti-bloat constraint (REQ-ATR-002) requires compression.

#### §C.3.2 Out of Scope — Workflow skill body re-authoring

Workflow skill bodies (`.claude/skills/moai/workflows/{plan,run,sync}.md` beyond the phase-owner-declaration scope) are NOT re-authored in this SPEC. Sub-skill modules under `.claude/skills/moai/workflows/{plan,run}/*.md` are NOT modified in this SPEC (deferred to a follow-up SPEC). The current SPEC's scope is restricted to owner declarations + hook-reference replacement of phantom-agent spawn references; the 23-workflow + 10-sub-skill 13K LOC body re-authoring (Audit 2 finding) is a separate SPEC.

#### §C.3.3 Out of Scope — 88 pre-v3 SPEC bodies

Strict L48 SSOT discipline — the 88 pre-v3 EARS SPECs under `.moai/specs/SPEC-*` (excluding the V3R6 family) are read-only inputs. No body content modification is in scope. The 6-month GEARS backward-compatibility window (expires 2026-11-22) remains in effect.

#### §C.3.4 Out of Scope — CHANGELOG.md history modification

Pre-v3.0.0 CHANGELOG history (entries dated before this SPEC's creation) is NOT modified. Only the current SPEC's sync-phase Unreleased entry is added.

#### §C.3.5 Out of Scope — Predecessor SPEC bodies in `SPEC-V3R6-*` directories

The 5 predecessor SPECs (`SPEC-V3R6-WORKFLOW-ORCHESTRATION-FIX-001` is supersedence target — frontmatter-only modification scoped to REQ-ATR-006; all others read-only) are NOT body-modified.

---

## §D — Requirements (GEARS Notation Self-Dogfood ≥80%)

This SPEC dogfoods GEARS notation across the 20 REQ-ATR-XXX entries below. Pattern distribution: 4 Ubiquitous + 3 Event-driven `When` + 3 State-driven `While` + 2 Capability-gate `Where` + 2 Event-detected (replaces Unwanted IF/THEN) + 1 Compound + 5 additional GEARS-shaped (Event-driven mixed). The deprecated `IF/THEN` modality is forbidden per SPEC-V3R6-GEARS-MIGRATION-001.

### REQ-ATR-001 (Ubiquitous)

The MoAI agent catalog shall consist of exactly 8 retained agents (`manager-spec`, `manager-develop`, `manager-docs`, `plan-auditor`, `evaluator-active`, `builder-harness`, `manager-git`, plus the Anthropic built-in `Explore`) plus zero phantom agents (zero MoAI-custom files for `manager-strategy`, `manager-quality`, `manager-brain`, `manager-project`, `claude-code-guide`, `researcher`, and the 6 `expert-*` agents).

### REQ-ATR-002 (Ubiquitous)

Each retained agent file (the 7 MoAI-custom retained agents — `Explore` is Anthropic-shipped and not a MoAI file) shall have body line count ≤ 500 lines per Anthropic CLAUDE.md anti-bloat guidance.

### REQ-ATR-003 (Ubiquitous)

Each retained agent file shall declare an explicit `tools:` CSV whitelist in YAML frontmatter (no inherited-from-default tools; the explicit declaration enables hook-based tool-scope verification).

### REQ-ATR-004 (Ubiquitous)

Each retained agent file shall declare explicit `NOT-for:` clauses in its description field per Anthropic action-oriented description guidance, enumerating the use cases where the agent is **not** the correct delegation target (e.g., manager-develop NOT-for: SPEC body authoring).

### REQ-ATR-005 (Event-driven `When`)

**When** this SPEC's run-phase milestone M3 (Agent Archive) begins, the orchestrator shall archive the 12 phantom and domain-expert agent files (`manager-strategy`, `manager-quality`, `manager-brain`, `manager-project`, `claude-code-guide`, `researcher`, and the 6 `expert-*` files) to `.moai/backups/agent-archive-2026-05-25/` (preserving the original directory structure) **before** deleting the originals from `.claude/agents/` and the template mirror at `internal/template/templates/.claude/agents/`.

### REQ-ATR-006 (Event-driven `When`)

**When** this SPEC's plan-phase commits land on `main`, the predecessor SPEC `.moai/specs/SPEC-V3R6-WORKFLOW-ORCHESTRATION-FIX-001/spec.md` frontmatter shall transition `status: <current> → superseded` (with `updated: 2026-05-25`) per the Status Transition Ownership Matrix `* → superseded` ownership clause assigning manager-spec as the responsible agent.

### REQ-ATR-007 (Event-driven `When`)

**When** the MoAI orchestrator processes a `/moai <subcommand>` instruction from a paste-ready resume Block 5, the orchestrator shall invoke `Skill("moai", arguments: "<subcommand> $ARGUMENTS")` (the canonical Skill router invocation), not manual `Read` of the `SKILL.md` body file — re-affirming WORKFLOW-ORCHESTRATION-FIX-001 REQ-WOF-005 as the single carried-forward router requirement from the superseded SPEC.

### REQ-ATR-008 (Event-driven `When`)

**When** the run-phase enters Phase 2 implementation mode selection (per design.md §B.4 5-mode decision tree), the orchestrator shall log the chosen mode (trivial / background / agent-team / parallel / sub-agent) and the selection rationale in `.moai/specs/SPEC-{ID}/progress.md` § Mode Selection — re-affirming WORKFLOW-ORCHESTRATION-FIX-001 REQ-WOF-004 with explicit progress.md logging.

### REQ-ATR-009 (Event-driven `When`)

**When** a sync-phase commit completion event fires, the Stop hook (`.claude/hooks/moai/sync-phase-quality-gate.sh`) shall verify (a) `golangci-lint run --timeout=2m` exits 0, (b) `go test ./...` exits 0, (c) coverage delta for changed files is ≥ 0; the hook shall block the sync commit via exit code 2 on any verification failure.

### REQ-ATR-010 (State-driven `While`)

**While** the `manager-develop` agent operates with `cycle_type=ddd`, the agent body shall follow ANALYZE-PRESERVE-IMPROVE per `.claude/rules/moai/workflow/spec-workflow.md` § Run Phase DDD Mode.

### REQ-ATR-011 (State-driven `While`)

**While** the `manager-develop` agent operates with `cycle_type=tdd`, the agent body shall follow RED-GREEN-REFACTOR per `.claude/rules/moai/workflow/spec-workflow.md` § Run Phase TDD Mode.

### REQ-ATR-012 (State-driven `While`)

**While** the `manager-develop` agent operates with `cycle_type=autofix` (NEW mode added per Audit 1 Proposal 1), the agent body shall follow DIAGNOSE-PATCH-VERIFY with a maximum of 3 iterations and CI-autofix-protocol-compliant escalation per `.claude/rules/moai/workflow/ci-autofix-protocol.md`.

### REQ-ATR-013 (Capability-gate `Where`)

**Where** the harness level is `thorough` AND `workflow.team.enabled: true` is set in `.moai/config/sections/workflow.yaml` AND environment variable `CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS=1` is set, the orchestrator may consider Agent Teams mode (3-5 teammates per Anthropic Agent Teams documentation verbatim "Start with 3-5 teammates for most workflows") as a candidate mode in the Phase 0.95 Mode Selection decision tree.

### REQ-ATR-014 (Capability-gate `Where`)

**Where** a sync-phase commit modifies `go.mod`, `go.sum`, `package-lock.json`, `Pipfile.lock`, or `Cargo.lock` files, the Stop hook (`sync-phase-quality-gate.sh`) shall additionally execute a dependency manifest audit (`govulncheck ./...` for Go projects; `npm audit --omit=dev` for Node projects; equivalents for other languages) and shall block the sync commit on detected high-severity vulnerabilities — mechanically replacing the archived `expert-security` declarative dependency manifest audit pattern.

### REQ-ATR-015 (Event-detected — replaces Unwanted IF/THEN)

**When** the MoAI orchestrator detects an autonomous-flow attempt to skip GATE-2 (the plan-to-implement HUMAN GATE corresponding to Anthropic's Ctrl+G plan editor mandate) — for example, a paste-ready resume Block 5 containing `/moai run SPEC-XXX` without a preceding GATE-2 user confirmation — the orchestrator shall halt the autonomous flow, preload `AskUserQuestion` via `ToolSearch(query: "select:AskUserQuestion")`, and trigger an `AskUserQuestion` round to elicit the gate decision before proceeding; the `score ≥ 0.90 skip-eligible` autonomous bypass policy for Phase 0.5 plan-auditor verdicts shall be retained, but only Phase 0.5 (not GATE-2) is skip-eligible.

### REQ-ATR-016 (Event-detected — replaces Unwanted IF/THEN)

**When** the MoAI orchestrator detects an attempt to spawn `Agent(subagent_type="manager-strategy")` or any of the other 11 archived agents (`manager-quality`, `manager-brain`, `manager-project`, `claude-code-guide`, `researcher`, `expert-backend`, `expert-frontend`, `expert-security`, `expert-devops`, `expert-performance`, `expert-refactoring`), the spawn shall be rejected with `ARCHIVED_AGENT_REJECTED` error pointing the user to the retention matrix in §B.3 above and to the migration table in `.claude/rules/moai/workflow/archived-agent-rejection.md`.

### REQ-ATR-017 (Compound — combines `Where` + `While` + `When`)

`[Where the harness level is standard or thorough] [While the task scope is multi-domain (≥3 domains OR ≥10 files)] [When the orchestrator selects an execution mode in Phase 0.95]`, the orchestrator shall prefer Agent Teams mode if all REQ-ATR-013 prerequisites are met; otherwise the orchestrator shall fall back to parallel multi-spawn of retained agents (maximum 3-5 concurrent `Agent()` calls in a single message per Anthropic verbatim "Start with 3-5 teammates").

### REQ-ATR-018 (Ubiquitous)

The template directories (`internal/template/templates/.claude/agents/`, `internal/template/templates/.claude/skills/moai/workflows/`, `internal/template/templates/.claude/rules/moai/`, `internal/template/templates/.claude/hooks/moai/`) shall mirror the corresponding local-project changes (8 retained agents, 3 workflow router skills, ~10 rule files, 3 hook scripts) byte-for-byte per the Template-First discipline in `CLAUDE.local.md` §2.

### REQ-ATR-019 (Ubiquitous)

`NOTICE.md` shall be updated with the agent retention rationale, the 2026-05-25 archive date, and the Anthropic 2026 alignment citation (verbatim quotations of Findings A1-A6 with source URLs) per the Apache 2.0 attribution discipline already established for the revfactory/harness and Karpathy imports.

### REQ-ATR-020 (Compound — combines `While` + `Where`)

`[While SPEC-V3R6-WORKFLOW-ORCHESTRATION-FIX-001 is in superseded state] [Where this SPEC's run-phase milestone M7 (manager-git PR doctrine reconciliation) executes]`, the orchestrator shall consolidate `CLAUDE.local.md` §23.7 Hybrid Trunk policy with `.claude/skills/moai/workflows/sync.md` PR creation logic such that: Tier S/M SPECs default to main-direct push (Hybrid Trunk default per `.claude/rules/moai/workflow/git-workflow-doctrine.md`); Tier L SPECs OR explicit `--pr` flag invocations route through `manager-git` for PR creation (`feat/SPEC-XXX` branch + `gh pr create`).

---

## §E — Acceptance Criteria Reference

See `.moai/specs/SPEC-V3R6-AGENT-TEAM-REBUILD-001/acceptance.md` for the full mandatory acceptance criteria matrix (≥20 AC-ATR-XXX entries with Given/When/Then format, severity, evidence command, pass criterion, and REQ-ATR-XXX traceability mapping at 100% coverage).

---

## §F — Non-Goals

The following items are explicitly out of scope for this SPEC:

1. **Re-authoring the body content of the 8 retained agents** beyond the frontmatter (description, tools, NOT-for) + minor scoped refinement (cycle_type=autofix, manager-project absorption notes). Full body re-authoring deferred to a follow-up SPEC.
2. **Re-authoring the body content of the 3 workflow router skills** beyond the phase-owner-declaration + hook-reference-replacement scope. The 23-workflow + 10-sub-skill 13K LOC body re-authoring (Audit 2 finding) is deferred to a follow-up SPEC.
3. **Modifying the 88 pre-v3 SPEC bodies** under `.moai/specs/SPEC-*` (excluding V3R6 family). Strict L48 SSOT discipline. The 6-month GEARS backward-compatibility window remains in effect.
4. **Modifying pre-v3.0.0 CHANGELOG.md history.** Only the current SPEC's sync-phase Unreleased entry is added.
5. **Modifying the predecessor SPEC bodies under `.moai/specs/SPEC-V3R6-*`** (all 5 V3R6 predecessor SPECs are read-only inputs; only `SPEC-V3R6-WORKFLOW-ORCHESTRATION-FIX-001` frontmatter is modified per REQ-ATR-006).
6. **Implementing hook-based enforcement for all 12 archived agent coverage areas.** Only the 3 highest-priority hooks are in scope: PostToolUse Status Transition Ownership, Stop sync-phase quality gate, TaskCompleted team-mode AC verify. Other replacements (e.g., expert-devops deployment validation, expert-performance benchmark enforcement) remain as per-spawn-prompt patterns without hook automation.
7. **Modifying the 16-language rule files** under `.claude/rules/moai/languages/` — out of orchestration scope per the language-neutrality contract.
8. **Modifying the canonical SSOT files** (`.claude/rules/moai/development/spec-frontmatter-schema.md`, `.claude/rules/moai/core/askuser-protocol.md`, `.claude/rules/moai/core/zone-registry.md`) beyond the cross-reference / agent-name-purge scope.
9. **`/moai feedback`, `/moai review`, `/moai e2e`, `/moai design` workflow re-classification** — deferred per spec-workflow.md "Out of scope of this matrix" callout (existing classification preserved).

---

## §G — Risks

| # | Risk | Severity | Likelihood | Mitigation |
|---|------|----------|------------|------------|
| G1 | Invested-work loss — the 12 agent bodies represent significant authoring time (∼3,500-5,000 LOC across the archived 12 files) and rationale embedded in body content may be discarded prematurely. | MEDIUM | HIGH | Per REQ-ATR-005, archive (NOT delete) the 12 files to `.moai/backups/agent-archive-2026-05-25/` preserving the original directory structure. `git log --follow` history remains intact. A future "agent revival" SPEC may restore individual agents if Audit 3 conclusions are revised. The archive is reversible. |
| G2 | User confusion during transition — paste-ready resume messages, prior commit-message references, and external documentation may invoke archived agents (e.g., `Agent(manager-strategy, ...)`), producing runtime errors. | HIGH | MEDIUM | Per REQ-ATR-016, the orchestrator emits `ARCHIVED_AGENT_REJECTED` errors with a migration table pointing to the replacement pattern (e.g., `manager-strategy → manager-spec` planning role absorption). A dedicated rule file `.claude/rules/moai/workflow/archived-agent-rejection.md` documents the migration table. The error message includes the correct replacement-pattern example. |
| G3 | Hook-based enforcement causes false positives — the Stop hook quality gate may fail legitimate commits (e.g., commits that fix lint issues but have non-zero `golangci-lint` exit due to pre-existing baseline issues outside scope). | HIGH | MEDIUM | Per design.md §D Hook Architecture, each hook script supports `--baseline-mode` invocation (read baseline from `.moai/state/lint-baseline.json`) and `--skip-hook=<name>` opt-out flag (logged to `.moai/logs/hook-skip.log` for audit). A per-hook test suite under `.moai/tests/hooks/` validates expected behavior with golden-output fixtures. |
| G4 | Tier L plan-auditor verdict below 0.85 PASS threshold — the SPEC's 5-artifact set must achieve plan-auditor PASS ≥ 0.85 to advance to run-phase. | MEDIUM | LOW | Self-audit estimate: ~0.88-0.91 PASS (Tier L MARGINAL to skip-eligible). The 20 REQ-ATR + ≥20 AC-ATR + 100% traceability + risk-mitigation pairing in this §G achieves the four MP-1..MP-4 dimensions. Skip-eligibility 0.90 is preferred but not required; max-3-iter retry contract provides recovery path. |
| G5 | Multi-session execution span — Tier L SPEC run-phase with ~50-60 files may exceed single-session context budget. | MEDIUM | MEDIUM | Per `.claude/rules/moai/workflow/session-handoff.md`, paste-ready resume protocol + auto-memory project entry persists state across `/clear` cycles. The 8 milestones (M1-M8) decompose naturally into ≤4 milestones per session at standard Opus 4.7 1M context. |
| G6 | Template-local drift — `internal/template/templates/` mirror diverges from `.claude/agents/` local content if mirror sync is incomplete. | MEDIUM | LOW | Per REQ-ATR-018, byte-for-byte parity is enforced. Run-phase milestone M8 (Template-First parity check) runs `diff -r .claude/agents/ internal/template/templates/.claude/agents/` and `diff -r .claude/skills/moai/workflows/ internal/template/templates/.claude/skills/moai/workflows/` as verification gate. Acceptance.md AC-ATR-018 binds this check. |
| G7 | Manager-git Hybrid Trunk policy regression — clarifying Tier L OR `--pr` flag triggers manager-git PR ops may regress current Tier S/M/L = main-direct default in CLAUDE.local.md §23.7. | LOW | LOW | Per REQ-ATR-020, the consolidation is **explicit clarification not policy change**: current behavior (Tier S/M = main-direct, Tier L OR `--pr` = PR) is documented as the canonical rule. No existing SPEC sessions are retroactively affected. |
| G8 | Anthropic 2026 guidance evolution — the audit citations (Findings A1-A6) are 2026-mid Anthropic published guidance; if Anthropic ships new guidance reversing these positions, this SPEC's architectural conclusions may need revision. | LOW | LOW | The NOTICE.md attribution per REQ-ATR-019 documents the source URLs and quote dates, enabling future auditors to identify the guidance epoch. A future "agent revival" SPEC can re-evaluate against new Anthropic guidance if released. |

---

## §H — Cross-References

### §H.1 Audit reports (input rationale for this SPEC)

- **Audit 1** (17 agent SRP audit, 2026-05-25 prior session turn): 19 agents reviewed, 13/19 ≥0.85 SRP score, 12/17 phantom (0 invocations across recent 4 SPEC sessions). Per-agent SRP healthy, system-level utilization broken.
- **Audit 2** (workflow phase ownership audit, 2026-05-25 prior session turn): 23 workflows + 10 sub-skills (13K LOC total). 0/45 phases have explicit owner declarations. 0/6 HUMAN GATEs mechanically enforced. Team mode dead variant. AMBER assessment.
- **Audit 3** (Anthropic 2026 verbatim audit, 2026-05-25 prior session turn): 6 critical findings A1-A6 as cited in §B.1 above. Sources: claude.com/docs/en/{sub-agents, agent-teams, best-practices, hooks}, anthropic.com/engineering/built-multi-agent-research-system, platform.claude.com/docs/en/about-claude/models/whats-new-claude-4-7.

### §H.2 Predecessor / superseded SPEC

- **`SPEC-V3R6-WORKFLOW-ORCHESTRATION-FIX-001`** — in plan-phase iter-2 PASS-WITH-DEBT 0.8625 (commits `97e42eaf7` + `22b4dc691`). Per REQ-ATR-006, this SPEC's run-phase M6 transitions the predecessor frontmatter `status: superseded` with HISTORY v0.1.2 entry citing the architectural-pivot rationale.

### §H.3 Other predecessor SPECs (read-only inputs per L48 SSOT)

- **`SPEC-V3R6-GEARS-MIGRATION-001`** — GEARS notation v0.2.0 canonical migration (depends_on)
- **`SPEC-V3R6-FOUNDATION-CORE-GEARS-ALIGN-001`** — foundation core GEARS alignment (depends_on, closed at `0156c7003`)

### §H.4 Canonical rule SSOTs (read-only inputs)

- `.claude/rules/moai/workflow/spec-workflow.md` — Plan/Run/Sync phase canonical SSOT + SPEC Complexity Tier (S/M/L) + Phase 0.5 Plan Audit Gate
- `.claude/rules/moai/core/agent-common-protocol.md` — User Interaction Boundary (AskUserQuestion HARD) + Parallel Execution batching
- `.claude/rules/moai/core/askuser-protocol.md` — Channel Monopoly + Socratic Interview Structure + Free-form Circumvention Prohibition
- `.claude/rules/moai/development/manager-develop-prompt-template.md` — Section A-E 5-section delegation template + B1-B12 Known Issues
- `.claude/rules/moai/development/spec-frontmatter-schema.md` — Canonical 12 frontmatter fields + Status Transition Ownership Matrix
- `.claude/rules/moai/workflow/session-handoff.md` — Paste-ready resume 6-block + Block 5 Skill router invocation
- `.claude/rules/moai/workflow/context-window-management.md` — Model-specific threshold
- `.claude/rules/moai/workflow/ci-autofix-protocol.md` — CI auto-fix loop max-3-iter + AskUserQuestion escalation (REQ-ATR-012 cross-reference)
- `.claude/rules/moai/development/branch-origin-protocol.md` — BODP audit trail
- `.claude/rules/moai/workflow/git-workflow-doctrine.md` — Hybrid Trunk policy
- `CLAUDE.md` § Agent Catalog (target of REQ-ATR-001 17→8 retention list update)
- `CLAUDE.local.md` §23 (target of REQ-ATR-020 manager-git PR doctrine reconciliation)

### §H.5 External authoritative sources (Audit 3 citations)

- claude.com/docs/en/sub-agents — *"Subagents cannot spawn other subagents."* (Finding A2)
- claude.com/docs/en/best-practices — *"Define a custom subagent when you keep spawning the same kind of worker with the same instructions."* (Finding A3)
- claude.com/docs/en/agent-teams — *"Start with 3-5 teammates for most workflows. This balances parallel work with manageable coordination."* (Finding A1)
- anthropic.com/engineering/built-multi-agent-research-system — *"most coding tasks involve fewer truly parallelizable tasks than research, and LLM agents are not yet great at coordinating and delegating to other agents in real time."* (Finding A4)
- claude.com/docs/en/hooks — Stop, PostToolUse, SubagentStop, TaskCompleted hook documentation (Finding A6)
- platform.claude.com/docs/en/about-claude/models/whats-new-claude-4-7 — Opus 4.7 Adaptive Thinking

### §H.6 Out-of-scope follow-up SPECs (deferred)

- (future) `SPEC-V3R6-AGENT-BODY-COMPRESSION-001` — full body re-authoring of the 8 retained agents to enforce REQ-ATR-002 LOC ≤ 500 across all agent files
- (future) `SPEC-V3R6-WORKFLOW-SKILL-REBUILD-001` — re-author the 23 workflow + 10 sub-skill bodies (13K LOC Audit 2 finding) to align with the 8-agent catalog
- (future) `SPEC-V3R6-HOOK-EXTENSION-001` — additional hook scripts for archived-agent coverage areas not addressed by the 3 in-scope hooks (G6)
- (future) `SPEC-V3R6-AGENT-REVIVAL-001` (conditional) — restore individual archived agents if Anthropic 2026 guidance is revised or per-agent invocation evidence appears (G8)

---

Version: 0.1.0
Status: draft (plan-phase initial authoring)
Tier: L (constitutional scope, 5-artifact set including this spec.md + plan.md + acceptance.md + design.md + research.md; supersedes SPEC-V3R6-WORKFLOW-ORCHESTRATION-FIX-001)
