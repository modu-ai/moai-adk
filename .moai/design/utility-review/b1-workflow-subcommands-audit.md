# B1 — Workflow Subcommands Audit (/moai plan / run / sync / project)

Audit date: 2026-04-23
Auditor: B1 Audit Agent
Scope: 4 workflow-core subcommands

---

## Executive Summary

- **Subcommands audited**: 4 (`/moai plan`, `/moai run`, `/moai sync`, `/moai project`)
- **Skills examined**: 7 (4 main workflow + 3 team variant files)
- **Supporting Go packages**: 2 (`internal/workflow/`, `internal/core/project/`)
- **Total workflow LOC**: 3,487 lines across 4 main skill files

### Overall Health Scores

| Subcommand | Score | Primary Concern |
|---|---|---|
| `/moai plan` | 6/10 | Phase numbering incoherent, /clear not triggered inside plan, team/plan.md uses deleted `team-reader` agent |
| `/moai run` | 7/10 | Phase 2.0 Sprint Contract only `thorough` harness, progress.md purely text-based with no Go schema enforcement |
| `/moai sync` | 7/10 | Token budget (40K) vs actual Phase 0 fan-out is unrealistically low; SPEC lifecycle enforcement relies on text comparison |
| `/moai project` | 6/10 | manager-project not used in any Phase 3+ delegation; `team-reader` orphan reference in triggers; DB detection writes state artifact but has no schema enforcement |

### Top 5 Cross-Cutting Issues

1. **Deleted `team-reader` agent still referenced** — `team/plan.md` spawns `Agent(subagent_type: "team-reader")` but this agent was removed as part of SPEC-TEAM-001 dynamic generation migration. No static team agent definitions should exist.

2. **Token budget documentation vs reality mismatch** — spec-workflow.md states 30K/180K/40K budgets, but there is no enforcement mechanism. The sync workflow alone has 8 pre-phases before Phase 1; this cannot fit in 40K. The budgets are aspirational text, not enforced limits.

3. **Phase numbering is non-monotonic in plan.md** — execution order is `1A → 0.3 → 0.3.1 → 0.4 → 0.5 → 1.25 → 1B`, which means Phase 0.x runs AFTER Phase 1A. This is confusing and makes resume/checkpoint logic ambiguous.

4. **progress.md state is pure text without schema** — `SPEC-V3R2-RT-004` (Typed Session State + Phase Checkpoint) is drafted but not implemented. The current progress.md is free-form append-only text; no Go loader validates it.

5. **`/clear` not triggered inside plan.md** — spec-workflow.md mandates `/clear` after plan completion. plan.md has no `/clear` invocation in its body. Only `team/plan.md` Phase 5 Cleanup executes `/clear`. Sub-agent plan mode has no context reset path.

---

## /moai plan Audit

### File Inventory

| File | Lines | Role |
|---|---|---|
| `.claude/skills/moai/workflows/plan.md` | 764 | Primary orchestration skill |
| `.claude/skills/moai/team/plan.md` | 200 | Team variant |
| `.claude/agents/moai/manager-spec.md` | 140 | Primary agent |
| `.claude/rules/moai/workflow/spec-workflow.md` | ~198 | Normative reference |

### D1 — Workflow Integrity

**Phase sequence (documented):** `1A → 0.3 → 0.3.1 → 0.4 → 0.5 → 1.25 → 1B → Decision Point 1 → 1.5 → 2 → 2.3 → 2.5 → 3 → 3.5 → 3.6 → Decision Point 2 → Decision Point 3 → Decision Point 3.5`

**DEFECT D1-P1 (Critical):** Phase numbering is non-monotonic. plan.md line 79 introduces `Phase 1A`, then line 103 introduces `Phase 0.3`. The execution sequence requires reading Phase 1A first, then 0.3, then 0.4, then 0.5, then 1.25, then 1B. A reader or agent parsing this document top-to-bottom encounters a false ordering. Resume logic referencing "last completed phase" would be ambiguous — is "Phase 1A" before or after "Phase 0.3"?

**DEFECT D1-P2 (Medium):** Token budget constraint of 30K is declared in spec-workflow.md but plan.md has no enforcement mechanism. plan.md loads 8+ files at context loading (config.yaml, git-strategy.yaml, language.yaml, product.md, structure.md, tech.md, codemaps/ listing, specs/ listing), which on large projects can easily consume the entire 30K budget before research begins.

**DEFECT D1-P3 (Medium):** `/clear` is mandated in spec-workflow.md (line 109: "After /moai plan completion (mandatory)") but plan.md has no `/clear` invocation. Only `team/plan.md` Phase 5 Cleanup at line 188 executes `/clear`. Sub-agent plan workflow has no context reset path.

**Success criteria:** Documented at plan.md lines 716-730. Reasonably complete. The `spec-compact.md` auto-generation requirement (line 719, 725) is a new contract not present in spec-workflow.md, creating a drift between the normative rule and the workflow.

### D2 — Agent Delegation Accuracy

**Correct delegations:**
- Phase 1A: Explore subagent (read-only) — correct
- Phase 1B: manager-spec subagent — correct (line 285)
- Phase 2: manager-spec subagent (same agent as 1B) — correct
- Phase 2.3: plan-auditor subagent — correct
- Phase 2.5: manager-git subagent — correct
- Phase 3: manager-git subagent — correct

**DEFECT D2-P1 (Critical):** `team/plan.md` Phase 1 (line 68) spawns `Agent(subagent_type: "team-reader")`. The `team-reader` agent was deleted as part of SPEC-TEAM-001 migration to dynamic general-purpose teams. The current architecture uses `Agent(subagent_type: "general-purpose")` with role profile overrides. team/plan.md has not been updated to reflect this. Referenced deletion evidence: `.claude/worktrees/ci-fork-skip/.moai/specs/SPEC-TEAM-001/spec.md` lines 27, 143 confirm deletion of `team-reader.md`.

**DEFECT D2-P2 (Low):** plan.md triggers list (line 25) still references `manager-git` as an agent, but plan.md only conditionally creates a branch/worktree. For "no flag" path (current branch), manager-git is never invoked. The trigger reference is harmless but creates confusing expectations.

**AskUserQuestion usage:** plan.md uses AskUserQuestion at orchestrator level only (Decision Points 1, 2, 3, 3.5 and annotation cycle, Phase 0.3/0.3.1 interview). manager-spec agent body does NOT use AskUserQuestion — correct per agent-common-protocol.md.

### D3 — State Management

**State artifacts declared in plan.md:**
- `.moai/specs/SPEC-{ID}/interview.md` (Phase 0.3.1, line 169)
- `.moai/specs/SPEC-{ID}/design-direction.md` (Phase 1.25, line 273)
- `.moai/specs/SPEC-{ID}/spec.md`, `plan.md`, `acceptance.md` (Phase 2, line 376)
- `.moai/specs/SPEC-{ID}/spec-compact.md` (Phase 2, line 409)
- `.moai/reports/plan-audit/SPEC-{ID}-review-N.md` (Phase 2.3)

**DEFECT D3-P1 (Medium):** `progress.md` for the plan phase is never mentioned in plan.md. The run phase creates and updates `progress.md` (run.md lines 119-124), but there is no plan-phase equivalent. If plan is interrupted mid-annotation-cycle, there is no structured way to resume from the exact iteration count.

**DEFECT D3-P2 (Medium):** Interview round counter state (Phase 0.3.1, annotation cycle at line 348) is tracked only in the LLM context — there is no persistent file artifact. After session interruption, round count is lost. The annotation cycle counter "Track iteration count and display: 'Annotation cycle {N}/6'" (line 346) relies on context memory only.

**Re-planning Gate:** Not applicable to plan phase.

### D4 — Integration with Hooks

**Manager-spec hooks (agent frontmatter line 23-26):**
- `SubagentStop`: fires `spec-completion` handler via `handle-agent-hook.sh`
- No `PreToolUse` or `PostToolUse` hooks defined

**DEFECT D4-P1 (Low):** Phase 2.3 invokes `plan-auditor` subagent but `plan-auditor.md` has no hooks defined (observed from agent file listing). If plan-auditor crashes silently, the failure may not be surfaced to the orchestrator until the subsequent report read fails.

**MX tag handling:** Phase 3.5 (MX Tag Planning) is declared as MANDATORY (line 590). It scans for fan_in >= 3 functions and outputs an `mx_plan` section in plan.md. The plan phase correctly generates an MX context map for the run phase.

### D5 — User Interaction Patterns

**Annotation cycle:** Documented at lines 334-350. 1-6 iterations, AskUserQuestion between each. The guard rule (line 349: "DO NOT implement any code — only update the plan document") is explicitly enforced in every agent prompt. This is a strength.

**DEFECT D5-P1 (Medium):** Phase 0.3 Clarity Scoring (lines 113-133) uses a subjective scoring rubric evaluated by the LLM without any calibration baseline. Score-to-rounds mapping (line 126) produces 0 interview rounds for clarity 1-3 — but the Phase 0.3 section says score 1-3 triggers "ask one broad clarification question instead" (line 133), while the table at line 126 says "0 (request too vague)". These two statements are contradictory: one says 0 rounds, the other says one question.

**Error recovery:** Missing SPEC ID input is handled (line 757: "AskUserQuestion prompts user for feature description"). But there is no documented error path for manager-spec agent failure mid-SPEC-creation.

### D6 — Output Artifacts

| Artifact | Location | Schema Defined? | Version Stable? |
|---|---|---|---|
| spec.md | `.moai/specs/SPEC-{ID}/` | 8 required frontmatter fields (line 378) | Yes |
| plan.md | `.moai/specs/SPEC-{ID}/` | Prose, no schema | No |
| acceptance.md | `.moai/specs/SPEC-{ID}/` | Given/When/Then format | Yes |
| spec-compact.md | `.moai/specs/SPEC-{ID}/` | Subset of spec.md | Yes |
| interview.md | `.moai/specs/SPEC-{ID}/` | Template at line 170 | Yes |
| research.md | `.moai/specs/SPEC-{ID}/` | Prose, no schema | No |
| design-direction.md | `.moai/specs/SPEC-{ID}/` | Template at line 272 | Yes |

**DEFECT D6-P1 (Low):** `issue_number: 0` placeholder in spec.md frontmatter (line 378) is semantically ambiguous. Issue number 0 could mean "not yet created" or "no GitHub integration". The convention should be `issue_number: null` or `-1`, not 0, to avoid confusion with real GitHub issue #0.

### D7 — Team Mode Variant Consistency

**DEFECT D7-P1 (Critical):** As noted in D2-P1, `team/plan.md` uses `subagent_type: "team-reader"` which is a deleted agent. The team/plan.md has not been migrated to `Agent(subagent_type: "general-purpose")` with role profile overrides.

**DEFECT D7-P2 (Low):** `team/plan.md` Phase 5 Cleanup (line 179) runs `moai cc` to remove GLM env vars. This is CG mode-specific cleanup. In pure Agent Teams mode (no GLM), `moai cc` is a no-op but harmless.

### /moai plan Verdict

**Score: 6/10.** Functional for the primary happy path (sub-agent, no team). Two critical defects: (1) non-monotonic phase numbering creates resume ambiguity, (2) team/plan.md references a deleted agent making team mode broken. Three medium defects reduce confidence in edge cases.

---

## /moai run Audit

### File Inventory

| File | Lines | Role |
|---|---|---|
| `.claude/skills/moai/workflows/run.md` | 860 | Primary orchestration skill |
| `.claude/skills/moai/team/run.md` | 360 | Team variant (CG + Agent Teams modes) |
| `.claude/agents/moai/manager-ddd.md` | ~160 | DDD implementation agent |
| `.claude/agents/moai/manager-tdd.md` | ~160 | TDD implementation agent |
| `.claude/rules/moai/workflow/workflow-modes.md` | ~80 | DDD/TDD methodology reference |
| `internal/workflow/` | 6 files | Go worktree orchestration |

### D1 — Workflow Integrity

**Phase sequence (documented):** `0.5 (conditional) → 0.9 → 0.95 → 1 → Decision Point 1 → 1.5 → 1.6 → 1.7 → 1.8 → 2.0 (thorough only) → Delta Detection → 2A/2B → 2.5 → 2.7 → 2.75 → 2.8a (standard/thorough) → 2.8b → 2.9 → 3 → 4`

**STRENGTH R1-S1:** Phase numbering in run.md is monotonic and well-structured. The phase table at the top (lines 67-82) is clear about which phases are skipped for which harness levels.

**DEFECT R1-P1 (Medium):** Phase 2.0 Sprint Contract Negotiation is gated to `harness level = thorough` only (run.md line 402). The design constitution (section 11) requires Sprint Contracts for `thorough` and recommends them for `standard`. run.md only implements the thorough path; the standard path has no Sprint Contract.

**DEFECT R1-P2 (Medium):** Token budget (180K) is documented in spec-workflow.md but not enforced by run.md. The context loading section (lines 85-98) loads 10+ files including SPEC documents, progress.md, tasks.md, structure.md, tech.md, and codemaps/. On large projects this pre-load could easily consume 40-60K tokens before Phase 1 begins.

**DEFECT R1-P3 (Low):** The Drift Guard Check thresholds (lines 487-497) at 20% warning and 30% trigger are hardcoded in the skill text. If these need tuning, they require a skill edit rather than config change. They should be driven from `quality.yaml` or `harness.yaml`.

### D2 — Agent Delegation Accuracy

**Phase routing:**
- Phase 0.95 Scale-Based Mode Selection: Inline logic (no agent) — correct
- Phase 1: manager-strategy — correct
- Phase 2A: manager-ddd — correct
- Phase 2B: manager-tdd — correct
- Phase 2.5: manager-quality — correct
- Phase 2.8a: evaluator-active — correct
- Phase 2.8b: manager-quality — correct
- Phase 3: manager-git — correct

**DEFECT D2-P1 (Medium):** Phase 0.95 Fix Mode detection (line 226: "SPEC scope ≤ 3 files, single domain → Fix Mode, agents: expert-debug + expert-testing") bypasses manager-strategy entirely. If the SPEC has acceptance criteria that require coordination across files, going directly to expert-debug risks missing strategy analysis. The scale-based bypass has no AskUserQuestion confirmation.

**DEFECT D2-P2 (Low):** run.md line 731 references `Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>` hardcoded in commit message format. This should be dynamic (based on model-policy.md). With Opus 4.7 deployments, this attribution becomes stale.

**Delegation pattern chain:** MoAI → manager-strategy → (manager-ddd or manager-tdd) → manager-quality → evaluator-active → manager-git. This is correct and follows the documented agent chain.

**AskUserQuestion compliance:** run.md uses AskUserQuestion at orchestrator level for Decision Point 1 (line 259), quality gate escalation (line 580), and Phase 4 options (line 750). manager-ddd and manager-tdd agents do not use AskUserQuestion — consistent with agent-common-protocol.md.

### D3 — State Management

**State artifacts declared in run.md:**
- `.moai/specs/SPEC-{ID}/progress.md` (lines 119-124, 181-185, 344-347, 363-367)
- `.moai/specs/SPEC-{ID}/tasks.md` (Phase 1.5, line 314)
- `.moai/specs/SPEC-{ID}/contract.md` (Phase 2.0, line 424)

**STRENGTH R3-S1:** progress.md resume logic is well-designed (lines 116-124). If progress.md exists, load it and skip completed phases. The file persists across sessions.

**DEFECT R3-P1 (Critical):** progress.md format is pure free-form text (see template at line 120). There is no schema, no Go loader, and no validation. SPEC-V3R2-RT-004 (Typed Session State + Phase Checkpoint) is drafted to address this gap but is not yet implemented. The current text-only progress.md cannot be reliably parsed by any tool.

**DEFECT R3-P2 (Medium):** Re-planning Gate (Phase 2.7, line 592) says "Check `.moai/specs/SPEC-{ID}/progress.md` for stagnation signals" but the detection method requires counting acceptance criteria completion entries. With free-form text progress.md, this counting is done by the LLM reading the file — not by a deterministic parser. False positives/negatives are likely.

**DEFECT R3-P3 (Medium):** tasks.md "Drift Guard" (lines 487-497) compares `planned_files` from tasks.md against `actual_files` from divergence tracking. But `actual_files` is defined as output from the manager-ddd/tdd agent (lines 472-481), which is passed via context — not written to disk. If the agent is re-invoked in a new session, the drift tracking resets.

### D4 — Integration with Hooks

**manager-ddd agent hooks (frontmatter):**
- `PreToolUse` (Write|Edit|MultiEdit): `ddd-pre-transformation` — correct
- `PostToolUse` (Write|Edit|MultiEdit): `ddd-post-transformation` — correct
- `SubagentStop`: `ddd-completion` — correct

**manager-tdd agent hooks (frontmatter):**
- `PreToolUse` (Write|Edit|MultiEdit): `tdd-pre-implementation` — correct
- `PostToolUse` (Write|Edit|MultiEdit): `tdd-post-implementation` — correct
- `SubagentStop`: `tdd-completion` — correct

**LSP Quality Gates:** run.md lines 703-708 enforce zero-error LSP policy during run. This is well-specified but depends on LSP servers being installed. Phase 1.7 (File Structure Scaffolding, line 349) captures the LSP baseline, which is the correct approach.

**DEFECT D4-P1 (Low):** Phase 2.75 Pre-Review Quality Gate (line 598) references `workflows/gate.md` for execution but `gate.md` is a separate skill. This indirect invocation ("Execute gate workflow equivalent") is ambiguous — the run.md author may intend Read(`${CLAUDE_SKILL_DIR}/gate.md`) + execution, but this is not explicitly coded.

### D5 — User Interaction Patterns

**STRENGTH R5-S1:** Decision Point 1 Plan Approval (lines 258-285) has a robust pre-approval checklist (proportionality, code reuse, reference implementations, simplicity) that pushes back against over-engineering. This is well-designed.

**DEFECT R5-P1 (Medium):** Phase 0.95 Scale-Based Mode Selection (lines 215-235) auto-selects execution mode without user confirmation. For "Large cross-cutting change → Team Mode" (line 227), the agent spawns a team without asking the user. This violates the Context-First Discovery principle for potentially expensive operations.

**Error recovery:** "SPEC not found" error path is documented (line 849). Phase 2.8a FAIL path has a 3-cycle retry limit before user escalation (lines 644-648). These are adequate.

### D6 — Output Artifacts

| Artifact | Location | Schema Defined? | Version Stable? |
|---|---|---|---|
| Implementation code | Project files (per SPEC) | N/A | N/A |
| progress.md | `.moai/specs/SPEC-{ID}/` | Template only (no schema) | No |
| tasks.md | `.moai/specs/SPEC-{ID}/` | Table format defined (line 316) | Yes |
| contract.md | `.moai/specs/SPEC-{ID}/` | Not defined in run.md | No |
| MX_TAG_REPORT | In-context only | No persistent artifact | No |

**DEFECT D6-P1 (Low):** `MX_TAG_REPORT` (Phase 2.9, line 699) is described as output but there is no persistent file path. The report exists only as agent output, not as an auditable `.moai/reports/` artifact.

### D7 — Team Mode Variant Consistency

**team/run.md** is well-structured, distinguishing CG Mode and Agent Teams Mode correctly. The `Agent(subagent_type: "general-purpose")` spawn pattern is correct per current architecture.

**DEFECT D7-P1 (Low):** `team/run.md` Phase 2.2 (line 147) states "All teammates spawn in parallel." The description prompt for `tester` teammate (line 208) says "Read implementation files but do NOT modify them." However, the `tester` has `mode: "acceptEdits"` and `isolation: "worktree"`, which means it could write implementation files if it misunderstands its role. The restriction is enforced only via the prompt instruction, not via mode/tool restrictions.

### /moai run Verdict

**Score: 7/10.** The main sub-agent flow is well-designed. The harness routing system, MX context scanning, and LSP quality gates are strengths. Key gaps: progress.md has no schema enforcement, Sprint Contract is thorough-only, and auto-mode-selection bypasses user approval for potentially expensive team spawns.

---

## /moai sync Audit

### File Inventory

| File | Lines | Role |
|---|---|---|
| `.claude/skills/moai/workflows/sync.md` | 1,157 | Primary orchestration skill (largest workflow file) |
| `.claude/skills/moai/team/sync.md` | 60 | Team variant (rationale doc only) |
| `.claude/agents/moai/manager-docs.md` | ~120 | Primary documentation agent |
| `.claude/agents/moai/manager-git.md` | (exists) | Git operations agent |

### D1 — Workflow Integrity

**Phase sequence:** `0 → 0.08 → 0.1 → 0.5 → 0.55 → 0.6 → 0.7 → 1 → 2 → 3 → 4`

**DEFECT S1-P1 (Critical):** The declared token budget of 40K (spec-workflow.md) is structurally incompatible with sync.md's actual phase fan-out. Before Phase 1 (the first "planning" phase), sync.md runs:
- Phase 0: Gate workflow (lint + format + type-check + test in parallel)
- Phase 0.08: DB schema refresh (conditional)
- Phase 0.1: Full test suite execution
- Phase 0.5: Language detection + parallel diagnostics + deep code review (manager-quality subagent)
- Phase 0.55: Security scan (expert-security subagent)
- Phase 0.6: MX tag validation across 16 languages
- Phase 0.7: Coverage measurement + gap analysis + test generation (expert-testing subagent)

Each subagent invocation consumes its own context window. The orchestrator's 40K token budget is consumed by coordination overhead alone. The documented budget is aspirational fiction.

**DEFECT S1-P2 (Medium):** Phase 3.1.5 Local CI Mirror Validation (lines 798-907) contains hardcoded Go-specific cross-compile commands (line 840-845) targeting `./cmd/moai/`. This is moai-adk-go project-specific logic inside a general-purpose sync workflow that should work for all 16 supported languages. Other languages (Python, TypeScript) have only brief 3-command equivalents while Go has a detailed multi-target cross-compile matrix.

**DEFECT S1-P3 (Low):** `--merge` flag is documented as deprecated at sync.md line 1017 ("Logs warning: 'The --merge flag is deprecated'") but the flag is still listed in the supported flags section at line 103 without a deprecation notice.

**Success criteria:** sync.md lines 1112-1118 define comprehensive completion criteria. These are well-specified.

### D2 — Agent Delegation Accuracy

**Phase routing:**
- Phase 0.5.4: manager-quality subagent — correct
- Phase 0.55: expert-security subagent — correct
- Phase 0.7: expert-testing subagent — correct
- Phase 1.4: manager-docs subagent — correct
- Phase 2.2: manager-docs subagent — correct
- Phase 3.1: manager-git subagent — correct

**DEFECT D2-P1 (Medium):** Phase 0.5.4 "Deep Code Review" (line 284) invokes manager-quality subagent. But manager-quality is listed as a `haiku` model agent (from manager-docs.md frontmatter, which also uses haiku). Deep code review including OWASP Top 10 analysis requires higher reasoning capability. The `haiku` model assignment for manager-quality (observed via model: haiku in manager-docs.md) may produce shallow security reviews.

**Note:** manager-quality.md was not read in detail but is listed in agents/moai/. Its model field needs verification.

**DEFECT D2-P2 (Low):** Phase 3.3 (PR Ready Transition, line 981) — "If Team mode enabled and PR is draft: Transition to ready via `gh pr ready`". The sync phase is explicitly documented to always use sub-agent mode (team/sync.md). The Team mode check here is dead code.

### D3 — State Management

**State artifacts:**
- `.moai/backups/sync-{timestamp}/` (Phase 2.1, line 605) — correct safety backup
- `.moai/reports/sync-report-{timestamp}.md` (Phase 2.2, line 629) — correct
- `.moai/state/db-detection.json` (referenced by project.md Phase 4.1a, not sync.md directly)
- SPEC status updates to spec.md frontmatter (Phase 2.4, line 697)

**STRENGTH S3-S1:** Safety backup creation before any modifications (Phase 2.1) is a strong safety pattern. The backup integrity check (line 608: "non-empty directory check") is well-designed.

**DEFECT S3-P1 (Medium):** Phase 1.5 SPEC-Implementation Divergence Analysis (lines 555-579) detects divergences between planned and actual implementation. The detection method relies on `git diff` and `git log` — it cannot detect divergences in files that were created and then deleted during run phase, or implementation decisions that did not manifest in file changes.

**DEFECT S3-P2 (Low):** Context Memory Generation in Git Commits (Phase 3.1.1, lines 743-796) embeds AI decision context in commit messages. This is a useful pattern but the commit body schema (with Decision/Pattern/Constraint/Gotcha lines) is defined only in sync.md — there is no validation that manager-git produces commits in this format.

### D4 — Integration with Hooks

**manager-docs agent hooks (frontmatter):**
- `PostToolUse` (Write|Edit): `docs-verification` — correct
- `SubagentStop`: `docs-completion` — correct

**LSP Quality Gates:** sync.md lines 305-310 enforce max 10 warnings (more lenient than run phase's zero errors). This progressive relaxation is well-designed.

**MX Tag Handling:** Phase 0.6 (lines 349-451) is the most detailed multi-language MX tag implementation in the system. 16 languages with language-specific warn patterns. This is a strength. The P1/P2 blocking behavior (line 353: "[HARD] P1/P2 violations BLOCK sync") creates a hard gate.

**DEFECT D4-P1 (Low):** Phase 0.6 blocking rule (line 353) says to "Run /moai run to add missing tags." But invoking `/moai run` from inside `/moai sync` would create a nested workflow invocation. The correct instruction should be "Run `/moai mx` to add missing tags" — the dedicated MX workflow.

### D5 — User Interaction Patterns

**STRENGTH S5-S1:** Human gates are well-distributed: Phase 0 offers Abort after gate failure, Phase 0.1 offers Fix/Continue/Abort after test failure, Phase 0.55 offers Fix/Continue/Abort after security findings, Phase 1.6 requires documentation scope approval.

**DEFECT S5-P1 (Medium):** Phase 0.4 Backward Compatibility Assessment (line 228: "Require explicit user acknowledgment via AskUserQuestion") is triggered for breaking changes. But the assessment is done inline (no subagent delegation), meaning it competes with other Phase 0 work for context bandwidth.

**DEFECT S5-P2 (Low):** Phase 3.4 Auto-Merge trigger conditions (line 1006: "`is_worktree_context == true` AND `--no-merge` flag NOT set") change the merge behavior automatically for worktree contexts. Users who run sync from a worktree context expecting the same behavior as non-worktree contexts will be surprised by auto-merge. This behavioral difference is documented but easy to miss.

### D6 — Output Artifacts

| Artifact | Location | Schema Defined? |
|---|---|---|
| sync-report-{timestamp}.md | `.moai/reports/` | Prose, no schema |
| Safety backup | `.moai/backups/sync-{timestamp}/` | Directory copy |
| PR (GitHub) | GitHub | Via `gh` CLI |
| Updated documentation | Various | Language-dependent |

**DEFECT D6-P1 (Low):** `sync-report-{timestamp}.md` format is not defined in sync.md. The report is generated by manager-docs (Phase 2.2 line 629) without a schema. Consumer tools cannot reliably parse the report.

### D7 — Team Mode Variant Consistency

sync.md line 1089: "The sync phase always uses sub-agent mode (manager-docs), even when --team is active."  
team/sync.md confirms this and provides rationale.

This is consistent. The rationale (sequential consistency, single voice, small output set) is sound.

### /moai sync Verdict

**Score: 7/10.** The phase structure is comprehensive and the multi-language MX validation is excellent. The token budget claim of 40K is structurally impossible given the actual phase depth. The Go-specific CI mirror hardcoding is a language-neutrality violation.

---

## /moai project Audit

### File Inventory

| File | Lines | Role |
|---|---|---|
| `.claude/skills/moai/workflows/project.md` | 706 | Primary orchestration skill |
| `.claude/agents/moai/manager-project.md` | ~120 | Setup specialist agent |
| `internal/core/project/` | 11 Go files | Language/methodology detection, validation |

### D1 — Workflow Integrity

**Phase sequence:** `0 → 0.3 or 1 → 1.5 → 2 → 3 → 3.1 → 3.3 → 3.5 → 3.7 → 4.1a → 4`

**STRENGTH P1-S1:** Project Type Detection (Phase 0, line 40) with explicit Existing/New routing is correct. The `[HARD] Auto-detect project type` rule prevents skipping this gate.

**DEFECT P1-P1 (Critical):** manager-project agent is listed in the `triggers.agents` list (project.md line 26: `"manager-project"`) but is NEVER delegated to in the workflow. project.md Phase 3 delegates directly to `manager-docs` (line 244: "[HARD] Delegate documentation generation to the manager-docs subagent"). manager-project is entirely bypassed. This is a fundamental mismatch: the agent catalog lists manager-project as "Project setup specialist" but the workflow that invokes it never calls it.

**DEFECT P1-P2 (Medium):** Phase 3.7 (Development Methodology Auto-Configuration, line 376) sets `development_mode` in `quality.yaml` based on test file ratio. The Go implementation in `internal/core/project/methodology_detector.go` presumably does the same detection. But the skill text (line 396: "Use the Bash tool with a targeted YAML update (read, modify, write back)") describes inline YAML modification via Bash — not delegation to the Go backend. The two implementations may have diverging detection logic.

**DEFECT P1-P3 (Low):** Phase 4 ordering inconsistency: the phase header is `## Phase 4: Completion` (line 459) but Phase 4.1a (DB Detection, line 413) is described BEFORE Phase 4. The annotation `## Phase 4.1a: DB Detection` reads as if it precedes Phase 4 but is actually a sub-phase of Phase 4.

### D2 — Agent Delegation Accuracy

**Phase routing:**
- Phase 0: MoAI orchestrator (AskUserQuestion) — correct
- Phase 0.3: MoAI orchestrator (interview loop) — correct but unusual (most workflows delegate interviews)
- Phase 1: Explore subagent — correct
- Phase 1.5: MoAI orchestrator (interview loop) — correct
- Phase 2: MoAI orchestrator (AskUserQuestion) — correct
- Phase 3: manager-docs subagent — correct
- Phase 3.1: plan-auditor subagent — correct
- Phase 3.3: Explore + manager-docs subagents — correct
- Phase 3.5: expert-devops subagent (optional LSP install) — correct
- Phase 3.7: MoAI orchestrator (Bash YAML update) — unusual (not delegated)

**DEFECT D2-P1 (Critical):** manager-project agent exists and has a detailed workflow (6 routing modes, complexity analysis, etc.) but is never invoked by project.md. The agent's `[HARD]` rule at line 31 states "This agent runs as a SUBAGENT via Agent() in isolated, stateless context — CANNOT use AskUserQuestion." This suggests it was designed to receive pre-collected user choices. But project.md collects choices directly via AskUserQuestion and then delegates to manager-docs — bypassing manager-project entirely.

**DEFECT D2-P2 (Low):** Phase 3.3 Codemaps Generation delegates to "Explore + manager-docs subagents" and says "delegate to codemaps workflow (workflows/codemaps.md)" (line 335) but does not specify the delegation pattern (Read codemaps.md + execute? Direct Agent invocation?). The boundary between Phase 3.3 and the codemaps workflow is vague.

### D3 — State Management

**State artifacts:**
- `.moai/project/interview.md` (Phase 0.3 and 1.5)
- `.moai/project/product.md`, `structure.md`, `tech.md` (Phase 3)
- `.moai/project/codemaps/*.md` (Phase 3.3)
- `.moai/state/db-detection.json` (Phase 4.1a, line 443)

**STRENGTH P3-S1:** The `db-detection.json` state artifact (Phase 4.1a, lines 443-455) has a well-defined JSON schema with `detected`, `matched_keywords`, `source_files`, `scanned_at`, and `tech_md_hash` fields. The `tech_md_hash` stale-detection mechanism is a good engineering pattern.

**DEFECT P3-P1 (Medium):** The DB detection state artifact is written to `.moai/state/db-detection.json` but there is no Go struct defined for this schema in `internal/core/project/`. The schema exists only as a comment in the skill text. If Go code needs to read this state, it would require ad-hoc JSON parsing.

**DEFECT P3-P2 (Low):** Phase 3.7 directly writes `quality.yaml` via Bash (line 396: "Use the Bash tool with a targeted YAML update"). This bypasses any configuration validation that `internal/config/` might provide. A malformed YAML write could corrupt the quality configuration.

### D4 — Integration with Hooks

**manager-project agent** has no hooks defined in its frontmatter (file lines 1-20 show no hooks section). Given that manager-project writes to `.moai/project/`, PostToolUse hooks for docs-verification would be appropriate.

**DEFECT D4-P1 (Low):** Neither manager-project nor the project workflow skill has hooks for `docs-verification`. This means project document creation is not subject to the same post-write verification that manager-docs applies via its hooks.

### D5 — User Interaction Patterns

**Interview structure:** project.md has TWO interview paths:
- Phase 0.3: New project interview (3 rounds, lines 68-127)
- Phase 1.5: Existing project interview (3 rounds, lines 158-215)

Each uses AskUserQuestion with 4 options. This is consistent with the AskUserQuestion constraints (max 4 questions per call, max 4 options).

**DEFECT D5-P1 (Medium):** Phase 2 User Confirmation (line 219) presents analysis summary and asks user to "Proceed with documentation generation." This is a third human gate after two interview rounds. For simple existing projects, this creates three consecutive AskUserQuestion stops before any files are written. The UX may feel over-orchestrated.

**DEFECT D5-P2 (Low):** Phase 3.5 LSP check (line 339) offers "Auto-install now: Use expert-devops subagent to install (requires confirmation)". This is an appropriate guard. However, the "Show installation instructions" option implies printing instructions to the chat — there is no documented artifact location for these instructions.

### D6 — Output Artifacts

| Artifact | Location | Schema Defined? |
|---|---|---|
| product.md | `.moai/project/` | Fields defined (line 254) |
| structure.md | `.moai/project/` | Fields defined (line 255) |
| tech.md | `.moai/project/` | Fields defined (line 256) |
| codemaps/overview.md | `.moai/project/codemaps/` | Structure defined (line 327) |
| codemaps/modules.md | `.moai/project/codemaps/` | Structure defined (line 328) |
| interview.md | `.moai/project/` | Template defined (line 108) |
| db-detection.json | `.moai/state/` | JSON schema defined (line 443) |

**DEFECT D6-P1 (Low):** There is no documented maximum token size for product.md, structure.md, or tech.md. These files are loaded as context in plan.md, run.md, and sync.md. Unbounded file growth degrades downstream workflow performance progressively.

### D7 — Team Mode Variant Consistency

project.md has no team mode variant file (only `team/plan.md`, `team/run.md`, `team/sync.md`, `team/debug.md`, `team/review.md` exist). The `triggers.agents` list includes `"manager-project"` but this agent is never used.

**DEFECT D7-P1 (Low):** The absence of a team/project.md is not a defect — project initialization is a one-time sequential task that does not benefit from parallelism. However, the lack of documentation explicitly stating this (similar to team/sync.md's rationale) leaves the omission unexplained.

### /moai project Verdict

**Score: 6/10.** The workflow has good interview structure and the db-detection state artifact is well-designed. However, manager-project is listed as the responsible agent but is entirely bypassed. The methodology auto-configuration (Phase 3.7) duplicates logic already in `internal/core/project/methodology_detector.go` without coordination.

---

## Cross-Cutting Findings

### CF-1: Deleted `team-reader` Agent Still Referenced (Critical)

**Appears in:** team/plan.md lines 60-108

team/plan.md uses `subagent_type: "team-reader"` for spawning research teammates. SPEC-TEAM-001 deleted team-reader.md as part of the migration to dynamic `general-purpose` agent spawning. This makes team mode for the plan phase broken.

**Fix:** Replace `subagent_type: "team-reader"` with `subagent_type: "general-purpose"` plus the appropriate role profile parameters (`model: "haiku"`, `mode: "plan"`, etc.) in team/plan.md.

### CF-2: Token Budget Documentation vs Reality (Medium)

**Appears in:** plan.md, run.md, sync.md, spec-workflow.md

The stated budgets (30K/180K/40K) are not enforced by any mechanism. Sync.md's actual phase structure requires multiple subagent invocations before Phase 1 — far exceeding 40K in orchestrator context alone. The budgets in spec-workflow.md should either be updated to reflect actual consumption, or enforcement mechanisms should be added.

### CF-3: progress.md Without Schema (Medium)

**Appears in:** run.md, spec-workflow.md

Phase progress is tracked in free-form Markdown text. SPEC-V3R2-RT-004 (Typed Session State + Phase Checkpoint) addresses this but is not yet implemented. Until then, all resume and stagnation detection logic relies on LLM text parsing of progress.md.

### CF-4: /clear Trigger Gap (Medium)

**Appears in:** plan.md, spec-workflow.md

spec-workflow.md mandates `/clear` after plan completion. plan.md does not invoke `/clear`. Only team/plan.md Phase 5 Cleanup executes `/clear`. The /clear requirement is unenforced for the 90% of users on sub-agent mode.

### CF-5: Language-Neutrality Violation in sync.md (Low)

**Appears in:** sync.md lines 823-845

The Local CI Mirror Validation (Phase 3.1.5) contains detailed Go-specific cross-compile commands targeting `./cmd/moai/`. This is moai-adk-go project-specific logic embedded in a general-purpose workflow template that is deployed to all user projects. Per CLAUDE.local.md Section 15: "templates are neutral; no language bias."

---

## Integration with v3R2 SPECs

| Finding | Relevant SPEC(s) | Coverage |
|---|---|---|
| CF-3 (progress.md schema) | SPEC-V3R2-RT-004 (Typed Session State + Phase Checkpoint) | Directly addresses |
| CF-2 (token budget) | SPEC-V3R2-WF-003 (Multi-Mode Router) — partial | Partial (routing, not budgets) |
| CF-1 (team-reader deletion) | SPEC-TEAM-001 (completed) | Gap: team/plan.md was not updated |
| D1-P2 run.md harness routing | SPEC-V3R2-HRN-001 (Harness Routing) | Directly addresses harness.yaml loader |
| D2-P1 project.md manager-project bypass | None identified | No SPEC covers this gap |
| S1-P1 sync.md token budget | None identified | No SPEC covers sync budget reality |
| D1-P1 plan.md phase numbering | SPEC-V3R2-WF-002 (Thin-Wrapper Enforcement) | Indirect — workflow structure reform |
| D6-P1 run.md MX_TAG_REPORT no artifact | SPEC-V3R2-SPC-002 (@MX TAG v2 with hook JSON) | Directly addresses MX output format |

---

## Recommendations (Prioritized)

### Priority Critical

**REC-1: Fix team/plan.md to use `general-purpose` spawning pattern.**

Replace `Agent(subagent_type: "team-reader")` with `Agent(subagent_type: "general-purpose", model: "haiku", mode: "plan")` pattern in team/plan.md lines 60-108. This unblocks team mode for the plan phase.

**REC-2: Clarify or remove manager-project from project.md delegation chain.**

Either: (a) invoke manager-project in Phase 3 instead of going directly to manager-docs, or (b) document that manager-project is an alternative entry point for the CLI (`moai init` → manager-project, `/moai project` → project.md skill directly). The current state has a catalog mismatch.

### Priority High

**REC-3: Fix plan.md phase numbering to be monotonic.**

Renumber plan.md phases so that execution order matches label order. Proposed: `Phase 1: Initial Setup → Phase 2: Clarity Evaluation → Phase 3: Interview → Phase 4: UltraThink → Phase 5: Research → Phase 6: Design Direction → Phase 7: SPEC Planning → Phase 8: Validation`. The current 0.x-before-1.x numbering creates resume ambiguity.

**REC-4: Implement typed phase checkpoint (advance SPEC-V3R2-RT-004).**

Until progress.md has a typed schema, all stagnation detection and resume logic is fragile. Implement the `PhaseState` and `Checkpoint` types from SPEC-V3R2-RT-004 as a priority. This unblocks reliable session resume for both run and plan phases.

**REC-5: Add /clear invocation to plan.md Completion section.**

plan.md line 716 (Completion Criteria) should include a final step: "Execute /clear to reset context for the Run phase." This enforces the spec-workflow.md mandate and aligns sub-agent and team mode behavior.

### Priority Medium

**REC-6: Move Go-specific CI mirror commands out of sync.md.**

Phase 3.1.5 should either detect the project language and route to language-specific command blocks (all 16 languages equally treated), or the Go-specific cross-compile matrix should be in a separate `sync-go.md` module loaded via JIT detection. This restores language neutrality per CLAUDE.local.md §15.

**REC-7: Make Drift Guard thresholds configurable.**

run.md drift thresholds (20% warning, 30% trigger) should be moved to `harness.yaml` or `quality.yaml` to allow per-project tuning without editing skill files.

**REC-8: Define Sprint Contract for standard harness level.**

design constitution §11 recommends Sprint Contracts for standard harness. run.md Phase 2.0 implements only thorough. Add a lightweight Sprint Contract for standard harness (fewer negotiation rounds, no cross-validation).

### Priority Low

**REC-9: Write team/project.md rationale document.**

Similar to team/sync.md, document why /moai project has no team variant. This prevents future confusion about whether team mode was accidentally omitted.

**REC-10: Replace `issue_number: 0` with `issue_number: null`.**

plan.md line 378 specifies `issue_number: 0` as the default for no-issue creation. Use `null` or `-1` to avoid ambiguity with a real GitHub issue #0.

---

## Sources

Files audited:
- `/Users/goos/MoAI/moai-adk-go/.claude/skills/moai/workflows/plan.md` (764 lines, v2.8.0)
- `/Users/goos/MoAI/moai-adk-go/.claude/skills/moai/workflows/run.md` (860 lines, v2.11.0)
- `/Users/goos/MoAI/moai-adk-go/.claude/skills/moai/workflows/sync.md` (1157 lines, v3.7.0)
- `/Users/goos/MoAI/moai-adk-go/.claude/skills/moai/workflows/project.md` (706 lines, v2.2.0)
- `/Users/goos/MoAI/moai-adk-go/.claude/skills/moai/team/plan.md` (200 lines, v2.7.0 / v3.0.0)
- `/Users/goos/MoAI/moai-adk-go/.claude/skills/moai/team/run.md` (360 lines, v4.1.0)
- `/Users/goos/MoAI/moai-adk-go/.claude/skills/moai/team/sync.md` (60 lines, v2.5.0)
- `/Users/goos/MoAI/moai-adk-go/.claude/agents/moai/manager-spec.md`
- `/Users/goos/MoAI/moai-adk-go/.claude/agents/moai/manager-ddd.md`
- `/Users/goos/MoAI/moai-adk-go/.claude/agents/moai/manager-tdd.md`
- `/Users/goos/MoAI/moai-adk-go/.claude/agents/moai/manager-docs.md`
- `/Users/goos/MoAI/moai-adk-go/.claude/agents/moai/manager-project.md`
- `/Users/goos/MoAI/moai-adk-go/.claude/rules/moai/workflow/spec-workflow.md`
- `/Users/goos/MoAI/moai-adk-go/.claude/rules/moai/workflow/workflow-modes.md`
- `/Users/goos/MoAI/moai-adk-go/internal/workflow/` (specid.go, worktree_orchestrator.go, errors.go)
- `/Users/goos/MoAI/moai-adk-go/internal/core/project/` (11 Go files)
- SPEC-V3R2-RT-004, SPEC-V3R2-HRN-001, SPEC-V3R2-WF-001, SPEC-V3R2-WF-002, SPEC-V3R2-WF-003, SPEC-V3R2-ORC-004, SPEC-V3R2-ORC-005, SPEC-V3R2-SPC-002

---

Audit complete.
Written to: `/Users/goos/MoAI/moai-adk-go/.moai/design/utility-review/b1-workflow-subcommands-audit.md`
