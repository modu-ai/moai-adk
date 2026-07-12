---
name: moai
description: >
  MoAI unified orchestrator for autonomous development. Routes natural
  language or subcommands (plan, run, sync, project, fix, loop, mx,
  feedback, review, clean, codemaps, gate, harness) to specialized
  agents.
allowed-tools: Agent, AskUserQuestion, Skill, TaskCreate, TaskUpdate, TaskList, TaskGet, Bash, Read, Write, Edit, Glob, Grep
argument-hint: "[subcommand] [args] | \"natural language task\""
---

## Pre-execution Context

!`git status --porcelain 2>/dev/null || true`
!`git branch --show-current 2>/dev/null || true`

## Essential Files

.moai/config/config.yaml

---

## Authority References

Rules and constraints governing all workflows are always loaded from these sources. Do NOT duplicate their content here:

- Core identity, orchestration principles, agent catalog: CLAUDE.md
- Quality gates, security boundaries: .claude/rules/moai/core/moai-constitution.md
- SPEC workflow phases, token budgets: .claude/rules/moai/workflow/spec-workflow.md
- Development methodologies (DDD/TDD): .claude/rules/moai/workflow/spec-workflow.md (Run Phase section)
- Agent definitions: See CLAUDE.md Section 4. For agent creation, use builder-harness subagent (artifact_type=agent).
- @MX tag rules and protocol: .claude/rules/moai/workflow/mx-tag-protocol.md

---

## Routing Observation Ledger

When dispatching a subcommand or workflow, the orchestrator records the routing decision to the append-only routing-ledger (`.moai/state/routing-ledger.jsonl`) via `moai harness ledger record` at dispatch time — the request text is piped via stdin and only a privacy-preserving digest is stored, never verbatim user text. As the routed pipeline reaches gate points, machine evidence is appended via `moai harness ledger evidence` (gate exits, audit verdicts, verify-log paths). Outcome is never supplied as an input; it is finalized from machine evidence only. This observation is opt-in and fail-open — it never blocks routing, and it is a silent no-op unless the harness observability opt-in is enabled.

---

## Intent Router

### Raw User Input

$ARGUMENTS

### Routing Instructions

[HARD] Route the Raw User Input above using the strict priority order below. Extract the FIRST WORD of the input for subcommand matching. All text after the subcommand keyword is CONTEXT to be passed to the matched workflow — it is NOT a routing signal and MUST NOT influence which workflow is selected.

## Execution Mode Flags (mutually exclusive)

- `--team`: Force Mode 3 (agent-team) of the Phase 0.95 6-mode catalog (`.claude/rules/moai/workflow/orchestration-mode-selection.md` §A), subject to its capability gate
- `--solo`: Force Mode 5 (sub-agent — single sequential agent per phase)
- No flag: The orchestrator auto-selects from the full 6-mode catalog at Phase 0.95; the complexity auto-select thresholds are stated once in `orchestration-mode-selection.md` §B.1 (machine source: `workflow.yaml` `auto_selection`) and are not restated here

The `--team` / `--solo` flags are forced overrides onto the catalog; the flag-free default resolves through the catalog decision tree (§B) and its capability gates. The `--mode` dispatch axis is a separate axis — see the crosswalk in `orchestration-mode-selection.md` §G.1 (correspondence, not merge).

### Priority 1: Explicit Subcommand Matching

[HARD] Extract the FIRST WORD from the Raw User Input section above. If it matches any subcommand below (or its alias), route to that workflow IMMEDIATELY. Do NOT analyze the remaining text for routing — it is context for the matched workflow:

- **plan** (aliases: spec): SPEC document creation workflow
- **run** (aliases: impl): DDD/TDD implementation workflow (per quality.yaml development_mode)
- **sync** (aliases: docs, pr): Documentation synchronization and PR creation
- **project** (aliases: init): Project documentation generation
- **feedback** (aliases: fb, bug, issue): GitHub issue creation
- **fix**: Auto-fix errors in a single pass
- **loop**: Iterative auto-fix until completion conditions are satisfied
- **mx**: MX tag scan and annotation for codebase
- **review** (aliases: code-review): Code review with security and MX tag compliance
- **clean** (aliases: dead-code): Identify and safely remove dead code
- **codemaps**: Generate architecture documentation in `.moai/project/codemaps/`
- **gate** (aliases: check, pre-commit): Lightweight pre-commit quality gate (lint+format+type-check+test)
- **harness** (aliases: hrn, learn): harness lifecycle management — learning-lifecycle verbs (status / apply / rollback &lt;date&gt; / disable) + v4-lifecycle verbs (list / edit / remove / doctor), all dispatching through the unified `moai harness` Go-binary Cobra subcommand tree; the slash command is the documented user-facing entry point
- **goal**: Condition-declared universal agentic loop — arm a completion condition (`/moai goal "<condition>"`), check status, clear, or resume; evaluated each turn-end by the `stop-goal` Stop hook

### Priority 2: SPEC-ID Detection

Only if Priority 1 did not match: Check if the Raw User Input contains a pattern matching SPEC-XXX (such as SPEC-AUTH-001). If found, route to the **run** workflow automatically. The SPEC-ID becomes the target for DDD/TDD implementation.

### Priority 3: Natural Language Classification

Only if BOTH Priority 1 AND Priority 2 did not match: Classify the intent of the ENTIRE Raw User Input as natural language. This priority is NEVER reached when the first word matches a known subcommand.

[HARD] The cue words listed below are **English exemplars**, NOT literal-match requirements. Classify intent semantically for any `conversation_language` — a Korean, Japanese, Chinese, or other-language request expressing the same intent routes identically. Do not require the literal English tokens to appear.

- Planning and design language (design, architect, plan, spec, requirements, feature request) routes to **plan**
- Quality gate language (format, check, pre-commit, quality gate) routes to **gate**
- Security language (security, audit, owasp, vulnerability, injection, xss, csrf) routes to **review** (with `--security` scope)
- Error and fix language (fix, error, bug, broken, failing, lint) routes to **fix**
- Iterative and repeat language (keep fixing, until done, repeat, iterate, all errors) routes to **loop**
- Documentation language (document, sync, docs, readme, changelog, PR) routes to **sync** or **project**
- Feedback and bug report language (report, feedback, suggestion, issue) routes to **feedback**
- MX tag language (mx tag, annotation, code context, legacy annotate) routes to **mx**
- Implementation language (implement, build, create, add, develop) with clear scope routes to **moai** (default autonomous)

### Priority 4: Default Behavior

If the intent remains ambiguous after all priority checks, use AskUserQuestion to present the top 2-3 matching workflows and let the user choose.

If the intent is clearly a development task with no specific routing signal, default to the **moai** workflow (plan -> run -> sync pipeline) for full autonomous execution.

---

## Workflow Quick Reference

### plan - SPEC Document Creation

Purpose: Create comprehensive specification documents using GEARS format with Research-Plan-Annotate cycle.
Phases: Deep Research (research.md) -> SPEC Planning -> Annotation Cycle (1-6 iterations) -> SPEC Creation -> Independent Review (plan-auditor)
Agents: manager-spec (primary), Explore (research), plan-auditor (quality gate), manager-git (conditional)
Flags: --worktree, --branch, --resume SPEC-XXX, --team, --issue (opt-in; default skips GitHub Issue creation per the late-branch opt-in policy)
For detailed orchestration: Read ${CLAUDE_SKILL_DIR}/workflows/plan.md

### run - DDD/TDD Implementation

Purpose: Implement SPEC requirements through configured development methodology.
Agents: manager-develop (cycle_type=ddd|tdd per quality.yaml, primary), manager-git
Flags: --resume SPEC-XXX, --team
For detailed orchestration: Read ${CLAUDE_SKILL_DIR}/workflows/run.md

### sync - Documentation Sync and PR

Purpose: Synchronize documentation with code changes and prepare pull requests.
Agents: manager-docs (primary), sync-auditor (quality gate), manager-git
Modes: auto, force, status, project. Flags: --merge, --skip-mx
For detailed orchestration: Read ${CLAUDE_SKILL_DIR}/workflows/sync.md

### gate - Pre-Commit Quality Gate

Purpose: Lightweight pre-commit quality check running lint, format, type-check, and tests in parallel. Also integrated into run (Phase 2.75) and sync (Phase 0) workflows as automatic pre-checks.
Agents: Direct execution (no agent delegation)
Flags: --fix, --staged, --file PATH
Integration: Automatically invoked by run workflow (Phase 2.75) and sync workflow (Phase 0.0.1) with --fix behavior.
For detailed orchestration: Read ${CLAUDE_SKILL_DIR}/workflows/gate.md

### goal - Condition-Declared Agentic Loop

Purpose: Arm a completion condition (mechanical commands + model claims); the `stop-goal` Stop-hook evaluator blocks each turn-end until the conditions hold or a turn ceiling (default 30) is reached.
Verbs: `/moai goal "<condition>"` (register + arm), `status [--all]`, `clear`, `resume`.
Progression mode: autonomous (default) vs. semi-autonomous — chosen at Implementation Kickoff Approval; the gate stays mandatory in both modes.
For detailed orchestration: Read ${CLAUDE_SKILL_DIR}/workflows/goal.md

### fix - Auto-Fix Errors

Purpose: Autonomously detect and fix LSP errors, linting issues, and type errors.
Agents: manager-develop (cycle_type=autofix), Agent(general-purpose) with domain whitelist (fixes)
Flags: --dry, --sequential, --level N, --resume, --team
For detailed orchestration: Read ${CLAUDE_SKILL_DIR}/workflows/fix.md

### loop - Iterative Auto-Fix

Purpose: Repeatedly fix issues until completion conditions are satisfied or max iterations reached.
Agents: manager-develop (cycle_type=autofix), Agent(general-purpose) with domain whitelist
Flags: --max N, --auto-fix, --seq
For detailed orchestration: Read ${CLAUDE_SKILL_DIR}/workflows/loop.md

### mx - MX Tag Scan and Annotation

Purpose: Scan codebase and add @MX code-level annotations for AI agent context.
Agents: Explore (scan), Agent(general-purpose) with backend scope (annotation)
Flags: --all, --dry, --priority P1-P4, --force, --team
For detailed orchestration: Read ${CLAUDE_SKILL_DIR}/workflows/mx.md

### review - Code Review

Purpose: Multi-perspective code review with security, performance, quality, and UX analysis.
Agents: sync-auditor (review), Agent(general-purpose) with security scope
Flags: --staged, --branch, --security, --team
For detailed orchestration: Read ${CLAUDE_SKILL_DIR}/workflows/review.md

### clean - Dead Code Removal

Purpose: Identify and safely remove unused code with test verification.
Agents: manager-develop, Agent(general-purpose) with refactoring scope
Flags: --dry, --safe-only, --file PATH
For detailed orchestration: Read ${CLAUDE_SKILL_DIR}/workflows/clean.md

### codemaps - Architecture Documentation

Purpose: Scan codebase and generate architecture documentation.
Agents: Explore, manager-docs
Flags: --force, --area AREA
For detailed orchestration: Read ${CLAUDE_SKILL_DIR}/workflows/codemaps.md

### (default) - MoAI Autonomous Workflow

Purpose: Full autonomous research -> plan -> annotate -> run -> sync pipeline.
Phases: Parallel Exploration (research.md) -> SPEC Generation -> Annotation Cycle -> Implementation -> Sync
Agents: Explore, manager-spec, plan-auditor (quality gate), manager-develop, manager-docs, manager-git, sync-auditor (quality gate)
Flags: --loop, --max N, --branch, --pr, --resume SPEC-XXX, --team, --solo, --issue (opt-in; default skips GitHub Issue creation per the late-branch opt-in policy)
For detailed orchestration: Read ${CLAUDE_SKILL_DIR}/workflows/moai.md

### project - Project Documentation

Purpose: Generate project documentation by analyzing the existing codebase.
Agents: Explore, manager-docs, Agent(general-purpose) with devops scope (optional)
Output: product.md, structure.md, tech.md in .moai/project/
For detailed orchestration: Read ${CLAUDE_SKILL_DIR}/workflows/project.md

### feedback - GitHub Issue Creation

Purpose: Collect user feedback and create GitHub issues.
Agents: orchestrator-direct (records feedback via gh CLI)
For detailed orchestration: Read ${CLAUDE_SKILL_DIR}/workflows/feedback.md

### harness - Harness Lifecycle + Natural-Language Build (argument-branching)

This single `harness` subcommand dispatches to ONE of two workflows based on the FIRST token of `$ARGUMENTS` (argument-based routing — no second command is introduced). Apply the routing rule before any workflow-specific logic:

- **Reserved verb** (`status` / `apply` / `rollback` / `disable`) → route to the existing **harness learning lifecycle** workflow (Branch A below). This path is unchanged.
- **Reserved verb** (`list` / `edit` / `remove` / `doctor`) → route to the **harness-v4 lifecycle** handler (Branch A.1 below). These enumerate / edit / atomically-remove harness-v4 entries and run the reference-integrity smoke gate (`doctor`) via the `moai harness <verb>` Go binary subcommand.
- **Anything else** (a natural-language harness-creation request, e.g. "build a harness for CLI template development") → route to the **harness build entry** workflow (Branch B below).

#### Branch A — harness learning lifecycle (reserved verbs: status / apply / rollback / disable)

Purpose: Surface the harness learning subsystem (observer, 4-tier proposal ladder, 5-layer safety pipeline) to the user via the slash command path. The lifecycle verbs (status / apply / rollback / disable) dispatch through the unified `moai harness` Go-binary Cobra subcommand tree, which performs the file-system operations. Tier-4 application is gated by orchestrator-issued AskUserQuestion.
Skills: moai-harness-learner (Tier-4 surfacing companion), moai-meta-harness (project-specific harness generation, indirect)
Verbs: status (tier distribution + telemetry) | apply (next Tier-4 proposal → AskUserQuestion → 5-layer pipeline → snapshot + write) | rollback &lt;YYYY-MM-DD&gt; (restore snapshot) | disable (set learning.enabled: false)
Artifacts: `.moai/harness/usage-log.jsonl`, `.moai/harness/proposals/`, `.moai/harness/learning-history/snapshots/`, `.moai/harness/learning-history/applied/`, `.moai/harness/learning-history/frozen-guard-violations.jsonl`
Authoritative SPEC: the harness foundation policy (supersedes V3R3-HARNESS-001, V3R3-HARNESS-LEARNING-001, V3R3-PROJECT-HARNESS-001)
For detailed orchestration: Read ${CLAUDE_SKILL_DIR}/workflows/harness.md

#### Branch A.1 — harness-v4 lifecycle (reserved verbs: list / edit / remove / doctor)

Purpose: Manage harness-v4 entries — enumerate built harnesses, locate their manifest + specialist files for editing, atomically remove a harness with all its artifacts, or run the reference-integrity smoke gate. The four verbs dispatch to the `moai harness <verb>` Go binary subcommand which performs the filesystem work (scan `.claude/commands/harness/*.md` joined with `manifest.json`; atomic remove with fail-closed orphan prevention; doctor cross-references manifest/specialist/skill file existence).
Verbs: list (enumerate all harnesses: name + domain + entry command) | edit &lt;name&gt; (show manifest + specialist + skill paths for editing — manifest is the SSOT) | remove &lt;name&gt; (atomic removal of command + workflow + specialists + skills + manifest; fail-closed if any artifact is missing) | doctor (reference-integrity smoke gate: verifies every built harness's manifest/specialist/skill files exist and cross-reference correctly)
CLI: `moai harness list [--json]`, `moai harness edit <name> [--json]`, `moai harness remove <name>`, `moai harness doctor` (all support `--project-root`)
Artifacts: `.claude/commands/harness/<name>.md` (thin-wrapper command), `.claude/commands/harness/<name>/manifest.json` (SSOT), `.claude/workflows/harness-<name>-run.js` (Runner), `.claude/agents/harness/harness-<name>*-specialist.md` (specialists), `.claude/skills/harness-<name>*/` (companion skills)
Namespace: `.claude/commands/harness/`, `.claude/workflows/harness-*.js`, `.claude/agents/harness/`, and `.claude/skills/harness-*/` are USER-OWNED — `moai update` preserves them (backup if needed, never overwrites).

#### Branch B — harness build entry (natural-language request)

Purpose: Turn a natural-language harness-creation request into a concrete harness via Context-First Discovery (extract domain / goal / constraints / scope), harness `<name>` derivation (the name is derived from the request — NOT statically supplied by the user), explicit orchestrator-issued approval, then transition into the orchestrator-direct Builder (4 signal-driven phases: ANALYZE / PLAN / GENERATE / ACTIVATE). The orchestrator MUST conduct AskUserQuestion Socratic rounds (max 4 questions per round) when intent clarity is below 100%.
Skills: moai-meta-harness (project-specific harness generation, indirect)
Builder: orchestrator-direct processing (NOT a dynamic-workflow script) — the entry's Phases 0-3 hand off to `${CLAUDE_SKILL_DIR}/workflows/harness-builder.md` for the 4-phase creation logic. The orchestrator holds the PLAN→GENERATE AskUserQuestion approval gate directly.
For detailed orchestration: Read ${CLAUDE_SKILL_DIR}/workflows/harness-build-entry.md

---

## Execution Directive

When this skill is activated, execute the following steps in order:

Step 1 - Parse Arguments:
Extract subcommand keywords and flags from the Raw User Input. Recognized global flags: --resume [ID], --seq, --team, --solo. Also detect `ultrathink` keyword in the input text.

**CRITICAL: Deep analysis mode:**
- `ultrathink` keyword detected → Activate Claude's native extended reasoning (xhigh effort mode). This is native Claude behavior with no MCP dependency.

Step 1.5 - Flag-Subcommand Compatibility Validation:
[HARD] After parsing the subcommand and flags (Step 1), validate flag-subcommand compatibility BEFORE routing. If a forbidden combination is detected, STOP all further processing and output an error in the user's conversation_language. Do NOT proceed to Step 2.

Forbidden flag-subcommand combinations:

| Flag | Allowed subcommands | Forbidden subcommands |
|------|---------------------|------------------------|
| `--worktree` | `plan` | `run`, `sync`, default (autonomous) |
| `--branch` | `plan`, default (autonomous) | `run`, `sync` |

Rationale: `--worktree` provisions an isolated workspace at SPEC initialization; only `/moai plan --worktree` creates one, so `/moai run` and `/moai sync` MUST operate within the worktree already established during `plan` — re-creating during run/sync corrupts the SPEC lifecycle and is rejected at the router level. `--branch` (feature-branch creation) is parsed at both `plan` and the default autonomous pipeline, but remains forbidden for `run`/`sync` for the same re-creation-corruption reason.

Error message template (Korean conversation_language; substitute the actual flag and subcommand):
```
에러: --worktree 플래그는 /moai plan 전용입니다.
/moai run 과 /moai sync 는 plan 단계에서 생성된 기존 worktree/branch를 재사용합니다.

올바른 사용법:
  /moai plan SPEC-XXX --worktree    (worktree 생성)
  /moai run SPEC-XXX                (기존 worktree/branch 재사용)
  /moai sync SPEC-XXX               (기존 worktree/branch 재사용)

다시 실행하려면 --worktree 플래그를 제거한 형태로 호출하세요.
```

For English (`en` conversation_language), translate the message; the structure remains identical.

Step 2 - Route to Workflow:
Apply the Intent Router (Priority 1 through Priority 4) to determine the target workflow. If ambiguous, use AskUserQuestion to clarify with the user.

Step 2.5 - Project Documentation Check:
Before executing plan, run, sync, fix, loop, or default workflows, verify project documentation exists by checking for `.moai/project/product.md`. If product.md does NOT exist, use AskUserQuestion to ask the user (in their conversation_language):

Question: Project documentation not found. Would you like to create it first?
Options:
- Create project documentation (Recommended): Generates product.md, structure.md, tech.md through a guided interview. This helps MoAI understand your project context for better results in all subsequent workflows.
- Skip and continue: Proceed without project documentation. MoAI will have less context about your project.

This check does NOT apply to: project, feedback subcommands.

[HARD] Beginner-Friendly Option Design:
All AskUserQuestion calls throughout MoAI workflows MUST follow these rules:
- The first option MUST always be the recommended choice, clearly marked with "(Recommended)" suffix
- Every option MUST include a detailed description explaining what it does and its implications

Step 2.8 - Requirement Analysis & Completion Condition:
Before loading the workflow body (Step 3), produce a requirement-analysis record for the routed request:

1. **Requirement summary** (1-3 sentences): what the user asked for, restated in the orchestrator's own words.
2. **Completion condition**: the end state that means "done". Where the condition is machine-verifiable (test exit code, lint-clean state, grep count, bounded turn count), express it in `/goal`-compatible transcript-measurable form per `.claude/rules/moai/workflow/goal-directive.md` (one measurable end state + a stated check + a bound clause). Do NOT invent a parallel evaluator: set the condition via `/goal` when the runtime supports it; otherwise the orchestrator evaluates the identical condition text per-turn (graceful degradation — no new machinery).
3. **Pipeline contract**: `full-pipeline` (default natural-language route — run-phase completion auto-chains into sync) or `single-phase` (explicit `run`/`sync` subcommand — chaining is offered as the "(Recommended)" next-step option, never fired silently).
4. **Orchestration-shape pre-signal**: an early input to the Phase 0.95 6-mode selection (`orchestration-mode-selection.md` §A) — noted here, decided at Phase 0.95.

Trivial-scope exemption: skip this step entirely for `feedback`, `gate`, `codemaps`, `sync` status mode, and any Stage-1-Clarify exception per `askuser-protocol.md` § Ambiguity Triggers and Exceptions.
Socratic-first ordering: while intent clarity is below 100%, run the Socratic interview (per `askuser-protocol.md`) BEFORE deriving the completion condition — the condition encodes drained intent, never a guess.
A derived completion condition NEVER authorizes autonomous run-phase entry — Implementation Kickoff Approval remains mandatory at the plan→run boundary.

Step 3 - Load Workflow Details:
Read `workflows/<name>.md` for the target subcommand. (The Agent Teams static layer is retired; a `--team` flag falls back to sub-agent mode per `.claude/rules/moai/workflow/orchestration-mode-selection.md` — there is no separate `team/<name>.md` workflow file.)

Step 4 - Read Configuration:
Load relevant configuration from .moai/config/config.yaml and section files as needed.

Step 5 - Initialize Task Tracking:
Use TaskCreate to register discovered work items with pending status.

Step 6 - Execute Workflow Phases:
Follow the workflow-specific phase instructions. Delegate all implementation to appropriate agents via Agent(). Collect user approvals at designated checkpoints via AskUserQuestion.

Step 7 - Track Progress:
Update task status using TaskUpdate as work progresses (pending to in_progress to completed).

Step 8 - Present Results:
Display results to the user in their conversation_language using Markdown format.

Step 9 - Declare Completion:
When all workflow phases complete successfully, state that the workflow is complete in the Completion Report (banner / prose) so the result is unambiguous.

Step 10 - Guide Next Steps:
Use AskUserQuestion to present the user with logical next actions based on the completed workflow.

---

Version: 2.8.0
Last Updated: 2026-07-07
