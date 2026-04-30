---
id: SPEC-AUTO-DELEG-001
version: "0.1.0"
status: draft
created_at: 2026-04-30
updated_at: 2026-04-30
author: manager-spec
priority: High
labels: [auto-delegation, claude-md, agent-description, trigger-centric, orchestration]
issue_number: null
wave: 1
tier: 0
scope:
  - CLAUDE.md
  - .claude/agents/moai/*.md
  - internal/template/templates/CLAUDE.md
  - internal/template/templates/.claude/agents/moai/*.md
  - .claude/rules/moai/development/agent-authoring.md
blockedBy: []
dependents: []
---

# SPEC-AUTO-DELEG-001: Auto-Delegation Rules Strengthening

## HISTORY

- 2026-04-30: Initial draft. Wave 1 — Tier 0. Adds 5 [HARD] delegation triggers to CLAUDE.md §1 and rewrites 24 agent descriptions in trigger-centric form. Source: Anthropic "Subagents in Claude Code" — trigger-vs-capability principle.

---

## 1. Overview

The MoAI orchestrator currently delegates to subagents only when the user explicitly says "Use the X subagent" — auto-delegation rate is low. Two structural reasons:

1. **Orchestrator rules**: `CLAUDE.md` §1 has a [HARD] "Multi-File Decomposition" rule that says split work into pieces. It does NOT have rules that say *delegate to a subagent*. Anthropic's verbatim prescription ("ten or more files", "three or more independent pieces of work") is not codified.

2. **Agent descriptions**: 18 of 23 agent descriptions use the keyword `Use PROACTIVELY` but framed around **capability** (e.g., "Use PROACTIVELY for API design"). Anthropic's verbatim guidance is to frame around **trigger conditions**, not capability surface area. 5 agent descriptions don't use the keyword at all.

This SPEC fixes both layers in one PR. CLAUDE.md §1 gains 5 [HARD] delegation triggers (10+ files explored, 3+ independent units, fresh-perspective review, pipeline workflow, 5+ writes). All 24 agent descriptions are rewritten in `Use PROACTIVELY when: <trigger conditions>` form with explicit `NOT for:` clauses. After rollout, baseline auto-delegation rate is measured on a fixed sample of 50 historical user turns and recomputed; target is >= 30 percentage points improvement.

---

## 2. Problem Statement

### 2.1 Current State (Quantified)

- **CLAUDE.md §1**: 11 [HARD] rules. Zero are auto-delegation triggers. The closest ("Multi-File Decomposition") instructs decomposition (TodoWrite-style task splitting), not delegation (Agent() spawn).
- **24 agent descriptions**: 18 use `Use PROACTIVELY for <area>` (capability-centric); 5 omit the keyword entirely. Of the 18, only `expert-security` and a small number of others include explicit `NOT for:` clauses.
- **No baseline measurement** of auto-delegation rate exists. Anecdotal: in recent sessions, the orchestrator spawned `Agent()` only when user-prompted explicitly (~5-10% of complex turns).
- **CLAUDE.local.md §16** (private file) already has self-check triggers, but they are not promoted to the project-shared CLAUDE.md, so other contributors and future sessions don't inherit them.

### 2.2 Pain Points

- **P1 — Capability-only descriptions force keyword-search delegation**: Claude has to match user phrasing to a capability description (e.g., user says "design the API" → matches expert-backend "for API design"). When user phrasing is non-canonical ("how should the auth flow work?"), the keyword-search misses, and the orchestrator handles it directly.
- **P2 — No quantitative trigger floor**: A task touching 12 files and involving 4 independent steps may never reach delegation if the orchestrator decides to "just do it." Anthropic's heuristic ("10+ files, 3+ units") would force delegation; without the rule, the heuristic doesn't fire.
- **P3 — NOT-for clauses are sparse**: Without explicit boundaries, ambiguous tasks risk over-delegation (multiple agents triggered for same task). Today's sparseness mostly produces the opposite (under-delegation) but enabling auto-delegation without NOT-for clauses risks a swing to over-delegation.
- **P4 — Inconsistent style across descriptions**: 5 agents lack PROACTIVELY entirely; 18 use it for capability rather than trigger. Trigger discoverability is uneven.

### 2.3 Why now

Wave 1 — Tier 0. Highest immediate value: every orchestrator turn that should have delegated and didn't is a missed quality opportunity (specialized agents have domain rules, isolation, separate quality gates). The cost of fixing — one PR, ~24 description rewrites + 5 CLAUDE.md rules — is one-time. The benefit recurs on every subsequent turn.

### 2.4 Why this is a [HARD] rule SPEC, not a guideline

Anthropic's blog (research.md §1.1 verbatim) says these triggers are "strong signals to direct Claude toward subagents." A guideline would let the orchestrator defer them. A [HARD] rule mandates them. SPEC-CORE-BEHAV-001 already establishes the "scope discipline" pattern; this SPEC extends to delegation discipline.

---

## 3. Requirements (EARS)

### 3.1 Ubiquitous (CLAUDE.md §1 Additions)

- **REQ-DEL-001** [Ubiquitous] THE CLAUDE.md SECTION 1 HARD RULES SHALL include exactly five [HARD] delegation triggers, immediately after the "Multi-File Decomposition" rule, in the order: file-exploration, independence, freshness, pipeline, writes.

- **REQ-DEL-002** [Ubiquitous] THE FIVE [HARD] TRIGGER ENTRIES SHALL be phrased to be cross-referenceable from agent descriptions and from `.claude/rules/moai/development/agent-patterns.md`. Each entry SHALL include a quantitative threshold where applicable.

### 3.2 Event-Driven (Trigger Activation)

- **REQ-DEL-003** [Event-Driven] WHEN the MoAI orchestrator receives a user task, THE ORCHESTRATOR SHALL evaluate the 5 delegation triggers as a precondition to direct execution. If any trigger fires, the orchestrator SHALL delegate to a specialist subagent.

- **REQ-DEL-004** [Event-Driven] WHEN file exploration count exceeds 10 (e.g., a task that would require reading 11+ source files), THE ORCHESTRATOR SHALL delegate to a specialist subagent (typically `Explore` or a domain manager). The estimate is taken at task-start, not after the fact.

- **REQ-DEL-005** [Event-Driven] WHEN a task contains 3 or more independent work units (e.g., "design API", "write frontend components", "set up CI"), THE ORCHESTRATOR SHALL delegate at least one unit to a specialist subagent. Independent units are units whose outputs do not depend on each other.

- **REQ-DEL-006** [Event-Driven] WHEN a task requires fresh perspective (e.g., review of code the orchestrator just produced, audit of plan-phase artifacts), THE ORCHESTRATOR SHALL delegate to an isolated subagent (typically `evaluator-active`, `plan-auditor`, or `manager-quality`).

- **REQ-DEL-007** [Event-Driven] WHEN a task has clear pipeline structure (input → transform → output, e.g., SPEC → implement → test), THE ORCHESTRATOR SHALL delegate to per-stage specialist subagents (the "pipeline" pattern from agent-patterns.md §Pattern 1).

- **REQ-DEL-008** [Event-Driven] WHEN a task creates 5 or more files of the same kind (e.g., 5 new test files, 5 new agent definitions, 5 new SPEC documents), THE ORCHESTRATOR SHALL delegate to the appropriate builder or expert subagent.

### 3.3 Where (Agent Description Format)

- **REQ-DEL-009** [Where] WHERE an agent definition file `.claude/agents/moai/<name>.md` exists, THE DESCRIPTION FRONTMATTER FIELD SHALL begin with one of two patterns:
  - Single-line trigger: `<Capability summary>. Use PROACTIVELY when: <trigger conditions>`
  - Multi-line trigger:
    ```
    <Capability summary>. Use PROACTIVELY when:
    - <trigger condition 1>
    - <trigger condition 2>
    - <trigger condition 3>
    NOT for: <explicit boundaries>
    ```
  Each agent SHALL specify at least 2 trigger conditions and an explicit `NOT for:` clause.

- **REQ-DEL-010** [Where] WHERE an agent description currently uses the legacy `Use PROACTIVELY for <area>` form, THE DESCRIPTION SHALL be rewritten to `Use PROACTIVELY when: <triggers>`. The legacy "for" phrasing SHALL NOT be retained.

- **REQ-DEL-011** [Where] WHERE multiple agents could match an ambiguous user request, EACH AGENT'S `NOT for:` CLAUSE SHALL explicitly delegate to the more-specific agent (e.g., expert-backend NOT for "security audits, use expert-security"). This forms a routing hint chain.

- **REQ-DEL-012** [Where] WHERE the existing description body contains useful capability descriptors (e.g., "Backend architecture and database specialist"), THE CAPABILITY SUMMARY SHALL be retained as the leading sentence; only the trigger portion is rewritten.

### 3.4 Measurement (Baseline + Pilot)

- **REQ-DEL-013** [Ubiquitous] THE PROJECT SHALL conduct a baseline auto-delegation rate measurement using a fixed sample of 50 historical user turns. The sample SHALL be drawn from session transcripts dated within the prior 60 days. The baseline rate is the proportion of these turns that resulted in `Agent()` invocation **without** the user-explicit phrasing pattern `"Use the <X> subagent"` in the request.

- **REQ-DEL-014** [Ubiquitous] AFTER SPEC ROLLOUT, THE PROJECT SHALL re-measure the auto-delegation rate using a comparable fixed sample of 50 user turns from the post-rollout window. The pilot rate SHALL be at least 30 percentage points higher than the baseline rate.

- **REQ-DEL-015** [State-Driven] WHILE the differential measurement is in progress, IF the pilot rate is not at least 30 percentage points above baseline AND the rollout is at least 14 days old, THEN the SPEC SHALL be flagged for trigger-threshold tuning AND a regression report filed in `.moai/reports/auto-deleg-tuning-{DATE}/`.

### 3.5 Cross-Reference Consistency

- **REQ-DEL-016** [Ubiquitous] THE 5 [HARD] DELEGATION TRIGGERS IN CLAUDE.md §1 SHALL be cross-referenced from `.claude/rules/moai/development/agent-authoring.md` (description format guidance) and from `.claude/rules/moai/workflow/spec-workflow.md` (delegation in plan/run/sync phases).

- **REQ-DEL-017** [Ubiquitous] THE CLAUDE.md §14 (Parallel Execution Safeguards) AND §16 (anti-direct-execution self-check, currently CLAUDE.local.md only) SHALL be reconciled with the new §1 triggers — no contradictions, complementary coverage.

### 3.6 Constraints (Unwanted)

- **REQ-DEL-018** [Unwanted] THE NEW [HARD] TRIGGERS SHALL NOT mandate delegation for: typo fixes, single-line changes, single-file reads with explicit path, command invocations with all arguments. These are explicit exceptions (matching CLAUDE.md §7 Rule 5 exception list).

- **REQ-DEL-019** [Unwanted] THE AGENT DESCRIPTIONS SHALL NOT contain marketing language, redundant adjectives, or capability lists exceeding 3 items in the leading sentence. The trigger list IS the long list; the leading sentence is the one-line capability identifier.

- **REQ-DEL-020** [Unwanted] THE DESCRIPTION REWRITE SHALL NOT change agent behavior — only delegation discoverability. Agent body content (Workflow Steps, Scope Boundaries, etc.) is out of scope.

---

## 4. Acceptance Criteria

(Detailed Given-When-Then scenarios live in `acceptance.md`. This is the summary list.)

- **AC-1**: CLAUDE.md §1 contains exactly 5 [HARD] delegation triggers, all phrased per REQ-DEL-002. Diff is reviewable in one chunk.
- **AC-2**: All 24 agent description fields conform to REQ-DEL-009: trigger-centric, with at least 2 triggers and an explicit `NOT for:` clause.
- **AC-3**: Differential measurement: post-rollout auto-delegation rate >= baseline + 30 percentage points (REQ-DEL-014).
- **AC-4**: An audit script (`internal/template/agent_description_audit_test.go`) parses the 24 agent files and asserts each conforms to REQ-DEL-009 (regex check). Test passes.
- **AC-5**: CLAUDE.md §1, §14, §16 (promoted from CLAUDE.local.md) are mutually consistent — no rule says "always delegate" while another says "do not delegate" in the same scenario.
- **AC-6**: `.claude/rules/moai/development/agent-authoring.md` has a new section "Description Trigger Format" referencing REQ-DEL-009.
- **AC-7**: Template mirrors (`internal/template/templates/CLAUDE.md` and `.claude/agents/moai/*.md` mirrors) are in sync with the local files. `make build` succeeds.
- **AC-8**: NOT-for clauses form a directed graph (e.g., expert-backend → expert-security; manager-spec → manager-strategy for architecture decisions). The graph is acyclic. Verified via a graph-extraction test.
- **AC-9**: Sample baseline measurement artifact at `.moai/reports/auto-deleg-baseline-{DATE}/sample-50-turns.md` and pilot artifact `.moai/reports/auto-deleg-pilot-{DATE}/sample-50-turns.md` exist with explicit Agent() invocation counts.
- **AC-10**: A side-by-side diff document `.moai/reports/auto-deleg-pilot-{DATE}/before-after.md` shows description before/after for each of the 24 agents.

---

## 5. REQ-ID Matrix

| REQ-ID | Type | Priority | Verification | Acceptance Criterion |
|--------|------|----------|--------------|----------------------|
| REQ-DEL-001 | Ubiquitous | Critical | CLAUDE.md content audit | AC-1, AC-5 |
| REQ-DEL-002 | Ubiquitous | High | Phrasing audit | AC-1 |
| REQ-DEL-003 | Event-Driven | Critical | Trigger inspection (logs / transcripts) | AC-3 |
| REQ-DEL-004 | Event-Driven | High | Sample-based audit | AC-3, AC-9 |
| REQ-DEL-005 | Event-Driven | High | Sample-based audit | AC-3, AC-9 |
| REQ-DEL-006 | Event-Driven | High | Sample-based audit | AC-3, AC-9 |
| REQ-DEL-007 | Event-Driven | Medium | Pipeline-task fixture audit | AC-3 |
| REQ-DEL-008 | Event-Driven | High | Sample-based audit | AC-3, AC-9 |
| REQ-DEL-009 | Where | Critical | Audit script (regex) | AC-2, AC-4 |
| REQ-DEL-010 | Where | High | Audit script (negative match: legacy form) | AC-2, AC-4 |
| REQ-DEL-011 | Where | High | NOT-for graph extraction | AC-8 |
| REQ-DEL-012 | Where | Medium | Diff review | AC-10 |
| REQ-DEL-013 | Ubiquitous | Critical | Artifact file check | AC-9 |
| REQ-DEL-014 | Ubiquitous | Critical | Computed differential | AC-3 |
| REQ-DEL-015 | State-Driven | High | Tuning artifact (if triggered) | AC-3 |
| REQ-DEL-016 | Ubiquitous | High | Cross-reference link audit | AC-6 |
| REQ-DEL-017 | Ubiquitous | High | Manual rule consistency review | AC-5 |
| REQ-DEL-018 | Unwanted | Critical | Exception list comparison | AC-5 |
| REQ-DEL-019 | Unwanted | Medium | Description style audit | AC-2 |
| REQ-DEL-020 | Unwanted | Critical | Body content diff (must be empty) | AC-2, AC-10 |

**Total**: 20 requirements (5 Ubiquitous, 6 Event-Driven, 4 Where, 1 State-Driven, 4 Unwanted, 0 Optional). Distribution covers all five EARS patterns except Optional.

---

## 6. Out of Scope (Exclusions — What NOT to Build)

- **EX-1**: Agent body content modification is **out of scope**. Workflow Steps, Scope Boundaries, Delegation Protocol, Adaptive Behavior sections of agent files are NOT touched. Only the YAML frontmatter `description` field and surrounding capability summary lines are rewritten.
- **EX-2**: Creating new agents is out of scope. The 24-agent fleet count is fixed; this SPEC operates on existing agents.
- **EX-3**: Removing or merging existing agents is out of scope.
- **EX-4**: A real-time auto-delegation telemetry system (continuous monitoring of delegation rate) is out of scope. AC-3 measurement is one-shot during pilot. Continuous monitoring is a follow-up SPEC.
- **EX-5**: Modifying the `Agent()` tool semantics or adding new spawn modes is out of scope. This SPEC operates entirely on rule documents (CLAUDE.md, agent descriptions, agent-authoring.md).
- **EX-6**: Changing the underlying agent invocation pattern (natural-language delegation via "Use the X subagent") is out of scope. The pattern remains; trigger conditions become explicit.
- **EX-7**: Multi-language localization of trigger conditions (Korean / Japanese / Chinese keyword equivalents in descriptions) is preserved as today. This SPEC does not expand the i18n keyword catalog.
- **EX-8**: Static analysis tooling that auto-rewrites descriptions is out of scope. Rewrites are manual to preserve domain semantics.
- **EX-9**: Changes to `Explore` (system-provided agent) description are out of scope unless `Explore` is in the 24-count user-facing fleet (verify in plan phase).
- **EX-10**: Modifying agent `model` field assignments (Opus/Sonnet/Haiku) is out of scope. SPEC-ADVISOR-001 in this same Wave handles model allocation; this SPEC handles description format.

---

## 7. Open Questions

- **OQ-1**: Should the 5 [HARD] triggers be inserted as a single grouped block (under a new "Delegation Triggers" sub-heading) or interleaved into the existing 11 §1 rules in their natural order? Decision deferred to plan phase. Default: grouped, with a new sub-heading.

- **OQ-2**: For REQ-DEL-014's 30-percentage-point target — is this a hard gate (rollback if missed) or a target (acknowledge and continue)? Default: target with REQ-DEL-015 escalation; not a hard rollback.

- **OQ-3**: Does the 24-count include `Explore` (Anthropic-provided system agent)? Or does it include only `.claude/agents/moai/` files? Verify in plan phase. Recommended: 24 = 23 local files + 1 Explore reference if Explore has a description in CLAUDE.md §4.

- **OQ-4**: For REQ-DEL-005 ("3 or more independent work units"), what counts as "independent"? Two units are independent if outputs don't feed into each other. The plan phase should provide a 1-page heuristic with examples.

- **OQ-5**: For agents with i18n keyword sections (e.g., "EN: ..., KO: ..., JA: ..., ZH: ..."), does the trigger-centric rewrite apply only to the EN section or to all four? Default: all four for consistency. Translations may be approximate.

- **OQ-6**: Should `NOT for:` clauses use specific agent names (e.g., "NOT for: security audits — use expert-security") or generic categorical phrasing ("NOT for: security audits")? Default: specific names, to enable the routing-hint graph (REQ-DEL-011).

- **OQ-7**: For the 5 agents currently lacking the `Use PROACTIVELY` keyword (evaluator-active and 4 others), is rewriting them in scope of this SPEC or do they need a separate review? Default: in scope; they receive the same treatment as the 18 partial-conformance agents.

---

**Total lines**: ~250
**Status**: draft — awaiting plan-auditor review
