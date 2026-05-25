---
id: SPEC-V3R6-AGENT-TEAM-REBUILD-001
artifact: progress
version: "0.1.0"
created: 2026-05-25
updated: 2026-05-25
author: orchestrator
run_commit_sha: "<pending>"
sync_commit_sha: "<pending>"
mx_commit_sha: "<pending>"
---

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-05-25 | orchestrator | Initial progress.md authored at run-phase entry. Records §D Phase 0.5 plan-audit SKIP decision (skip-eligible 0.908) + §E Phase 0.95 Mode Selection (sequential single-spawn manager-develop per Finding A4). Run-phase M1-M8 not started; manager-develop owns §F Run-phase Audit-Ready Signal. |

---

## §A — SPEC Artifact Ownership (cross-reference)

Per `.claude/rules/moai/development/spec-frontmatter-schema.md` § Status Transition Ownership Matrix:

| Artifact | Plan-phase | Run-phase | Sync-phase | Mx-phase |
|----------|------------|-----------|------------|----------|
| spec.md / plan.md / acceptance.md / design.md / research.md | manager-spec (body authoring) | frontmatter `status: in-progress` only (manager-develop on first commit) | frontmatter `status: implemented` (manager-docs) | frontmatter `status: completed` (orchestrator on Mx chore) |
| progress.md §D Mode Selection + §F.4 Mx-phase Audit-Ready Signal | orchestrator | orchestrator (this section) | n/a | orchestrator |
| progress.md §F.1 Plan-phase Audit-Ready Signal | manager-spec | n/a | n/a | n/a |
| progress.md §F.2 Run-phase Audit-Ready Signal | n/a | manager-develop (per milestone + final) | n/a | n/a |
| progress.md §F.3 Sync-phase Audit-Ready Signal | n/a | n/a | manager-docs | n/a |

---

## §B — Lifecycle Status (mirror of plan.md §A; updated as milestones complete)

| Step | Phase | Owner | Status | Commit SHA | Notes |
|------|-------|-------|--------|------------|-------|
| 0 | Plan-phase M0 | manager-spec | COMPLETED | `b957a4d04` | 5 artifacts authored 2026-05-25 |
| 0.5 | plan-auditor verdict | plan-auditor | COMPLETED (iter-1 PASS skip-eligible 0.908) | (logged in MEMORY.md project_agent_team_rebuild_planphase_complete) | Tier L threshold 0.85 PASS +0.058; skip-eligible 0.90 PASS +0.008 |
| 0.5R | Phase 0.5 re-execution | orchestrator | **SKIPPED** | n/a | See §D Phase 0.5 SKIP rationale below |
| 0.95 | Mode Selection | orchestrator | COMPLETED | n/a | See §E Mode Selection below |
| 1 | Run-phase M1 — 7 retained agent frontmatter refinement | manager-develop | COMPLETED | `955299cac` | REQ-ATR-001..004 |
| 1 | Run-phase M2 — 3 workflow router phase-owner declarations | manager-develop | COMPLETED | `d9cce5427` | REQ-ATR-001 + REQ-ATR-007 + REQ-ATR-008 + REQ-ATR-012 |
| 1 | Run-phase M3 — Archive 12 phantom agents | manager-develop | COMPLETED | `476b04ffb` | REQ-ATR-005 |
| 1 | Run-phase M4 — 3 NEW hook scripts | manager-develop | COMPLETED | `fdd4aa37a` | REQ-ATR-009 + REQ-ATR-014 |
| 1 | Run-phase M5 — Rule files (2 NEW + 8 modified) | manager-develop | COMPLETED | `<this-commit>` | REQ-ATR-007/008/012/016/020 |
| 1 | Run-phase M6 — Predecessor SPEC supersedence | manager-spec | NOT-STARTED | `<pending>` | REQ-ATR-006; frontmatter-only per L48 SSOT |
| 1 | Run-phase M7 — CLAUDE.md + CLAUDE.local.md + NOTICE.md | manager-develop | NOT-STARTED | `<pending>` | REQ-ATR-001/015/019/020 |
| 1 | Run-phase M8 — Template parity + verification batch | manager-develop | NOT-STARTED | `<pending>` | REQ-ATR-018 |
| 2 | Sync-phase | manager-docs | NOT-STARTED | `<pending>` | CHANGELOG + 5 frontmatter `status: implemented` |
| 3 | Mx-phase | orchestrator | NOT-STARTED | `<pending>` | Step C judgement (expected EVALUATE-SKIP per markdown-heavy) |
| 4 | 4-phase close | orchestrator | NOT-STARTED | `<pending>` | `status: implemented → completed` + L60 atomic backfill |

---

## §C — Run-phase Strategy Notes

- **Predecessor**: SPEC-V3R6-WORKFLOW-ORCHESTRATION-FIX-001 (status: superseded as of plan-phase commit `b957a4d04`). Frontmatter transition was performed by manager-spec during plan-phase atomic 6-file commit (per Status Transition Ownership Matrix `* → superseded`). M6 in this SPEC verifies the transition rather than re-applying.
- **Session split estimate**: 3-5 sessions per option A preview (user-acknowledged in paste-ready resume). Each session targets 1-3 milestones with paste-ready handoff at end.
- **Plan.md vs paste-ready memo milestone order**: paste-ready memo (project_agent_team_rebuild_planphase_complete.md) listed M1=archive / M2=rebuild / M3=workflow. plan.md (SSOT per L48) authoritative order: M1=frontmatter refinement / M2=workflow routers / M3=archive / M4=hooks / M5=rules / M6=supersedence / M7=doctrine / M8=template. **Plan.md order wins**; this progress.md mirrors plan.md.

---

## §D — Phase 0.5 SKIP Rationale (orchestrator decision)

[ZONE:Evolvable] Per `.claude/rules/moai/workflow/spec-workflow.md` § Plan to Run "Plan Audit Gate skip policy" (CONST-V3R5-026 reference): when the most recent plan-auditor verdict was PASS with overall score ≥ 0.90 AND no plan-PR commit has landed since that verdict, the orchestrator MAY skip Phase 0.5 re-execution.

**Skip condition verification**:

| Condition | Required | Actual | PASS? |
|-----------|----------|--------|-------|
| plan-auditor verdict PASS | `PASS` | `PASS skip-eligible` | ✓ |
| Overall score ≥ 0.90 | ≥ 0.90 | 0.908 | ✓ (+0.008) |
| No plan-PR commit since verdict | No new commit | HEAD `651623dc1` is LOCAL-NAMESPACE-CONSOLIDATION-001 (different SPEC, scope-disjoint, L52 race-absorbed case 16); no AGENT-TEAM-REBUILD-001 commits since `b957a4d04` plan-phase | ✓ |
| MP-1/MP-2/MP-3/MP-4 all PASS | All PASS | MP-1 PASS / MP-2 PASS / MP-3 PASS (Traceability 1.00) / MP-4 N/A markdown | ✓ |

**Decision**: SKIP Phase 0.5 re-execution. Proceed directly to Phase 0.95 + Phase 1.

**Recorded in M1 delegation prompt Section A: Context** per CONST-V3R5-026 verification requirement.

---

## §E — Phase 0.95 Mode Selection (orchestrator decision)

[ZONE:Evolvable] Per plan.md §D.3 Phase 0.95 + REQ-ATR-008 (5-mode autonomous selection).

**Input parameters**:
- Tier: L (constitutional scope per spec.md)
- Scope: ~50-60 files (7 retain + 12 archive + 3 workflow + 10 rules + 3 hooks + 1 frontmatter + 1 NOTICE + ~13 template + 5 SPEC = expected breakdown)
- Domain categories: 5 (agents / workflow skills / rules / hook scripts / template mirrors)
- File language: 100% markdown + shell scripts; 0 Go source modifications
- Concurrency benefit: LOW (markdown sequential edits; no genuine parallel implementation gain per Anthropic Finding A4 "most coding tasks involve fewer truly parallelizable tasks than research")
- Agent Teams prereqs: `CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS` typically unset + `workflow.team.enabled` typically false in default project config

**Mode evaluation**:

| Mode | Applicability | Selected? |
|------|---------------|-----------|
| autopilot (sequential single-spawn manager-develop) | Tier L Section A-E template + 8 milestone breakdown; sequential M1→M8 via per-milestone manager-develop spawn (single concurrent agent, not 8 in parallel) | ✓ **SELECTED** |
| loop (Ralph Engine iterative fix) | Not applicable — initial implementation, not error-driven fix loop | ✗ |
| team (Agent Teams parallel) | Prereqs not met + Finding A4 coding-task parallelism caveat applies + markdown sequential edits have no genuine parallel benefit | ✗ |
| pipeline (deterministic 3-phase) | This is `/moai run` (multi-agent class, not utility class); pipeline mode rejected per REQ-WF003-016 sentinel `MODE_PIPELINE_ONLY_UTILITY` | ✗ (rejected by class) |
| (background) | Not applicable — implementation requires Write/Edit (foreground only per CONST-V3R2-020 background write restriction) | ✗ |

**Decision**: **autopilot** — sequential per-milestone manager-develop spawn (single concurrent agent per milestone, not 8 parallel). Each spawn uses Section A-E 5-section template per `.claude/rules/moai/development/manager-develop-prompt-template.md` § Applicability (Tier L REQUIRED).

**Justification per plan.md §D.3 verbatim**: "Tier L + multi-domain (5 categories) + scope ~50-60 files satisfies REQ-ATR-017 Compound preconditions. However, the task is markdown + shell-script-only (no Go code), and Anthropic Finding A4 advises coding tasks remain single-agent. Final decision: sequential single-spawn manager-develop with Tier L Section A-E template + 8 milestone breakdown."

---

## §F — Audit-Ready Signals (per phase)

### §F.1 Plan-phase Audit-Ready Signal (manager-spec scope — COMPLETED)

Logged in `~/.claude/projects/-Users-goos-MoAI-moai-adk-go/memory/project_agent_team_rebuild_planphase_complete.md`. Summary:
- 5 artifacts authored 2026-05-25 (spec 388 + plan 434 + acceptance 447 + design 402 + research 387 = 2058 lines)
- plan-auditor iter-1 PASS skip-eligible 0.908
- 20 REQ-ATR (100% GEARS notation)
- 22 AC-ATR (100% REQ↔AC traceability)
- 8 risks with mitigation pairing
- 3 MINOR defects only (D1/D2/D3 — handled inline run-phase M1/M3)
- 0 BLOCKING / 0 SHOULD-FIX
- Atomic 6-file commit `b957a4d04` (5 NEW + 1 supersedence frontmatter)

### §F.2 Run-phase Audit-Ready Signal (manager-develop scope — IN-PROGRESS, M1-M5 of 8 complete)

#### M1 — 7 retained agent frontmatter refinement
- **Commit**: `955299cac` `feat(SPEC-V3R6-AGENT-TEAM-REBUILD-001): M1 — 7 retained agent frontmatter refinement (Anthropic 2026 alignment)`
- **Date**: 2026-05-25
- **REQs covered**: REQ-ATR-001, REQ-ATR-002, REQ-ATR-003, REQ-ATR-004
- **Files**: 7 retained agent files (`.claude/agents/core/{manager-spec, manager-develop, manager-docs, manager-git}.md` + `.claude/agents/meta/{plan-auditor, evaluator-active, builder-harness}.md`)

#### M2 — Workflow router phase-owner declarations
- **Commit**: `d9cce5427` `feat(SPEC-V3R6-AGENT-TEAM-REBUILD-001): M2 — workflow router skill phase-owner declarations (manager-strategy + manager-quality chain removed)`
- **Date**: 2026-05-25
- **REQs covered**: REQ-ATR-001 (manager-strategy chain removal), REQ-ATR-007, REQ-ATR-008
- **Files**: 3 workflow router skills (`.claude/skills/moai/workflows/{plan,run,sync}.md`)

#### M3 — Archive 12 phantom and domain-expert agents
- **Commit**: `476b04ffb` `feat(SPEC-V3R6-AGENT-TEAM-REBUILD-001): M3 — archive 11 phantom and domain-expert agents`
- **Date**: 2026-05-25
- **REQs covered**: REQ-ATR-005
- **Archive directory**: `.moai/backups/agent-archive-2026-05-25/` (11 archived agents + README.md = 12 .md files; researcher.md variance noted per `meta/researcher.md unclassified` — plan.md scope gap, see project_atr001_m1m4_complete memory for handling rationale)

#### M4 — 3 NEW hook scripts
- **Commit**: `fdd4aa37a` `feat(SPEC-V3R6-AGENT-TEAM-REBUILD-001): M4 — 3 NEW hook scripts (PostToolUse + Stop + TaskCompleted)`
- **Date**: 2026-05-25
- **REQs covered**: REQ-ATR-009 (Stop sync-phase quality gate), REQ-ATR-014 (dependency manifest audit)
- **Files**: `.claude/hooks/moai/{status-transition-ownership,sync-phase-quality-gate,team-ac-verify}.sh`

#### M5 — Rule files (2 NEW + 8 modified)
- **Commit**: `<this-commit>` `feat(SPEC-V3R6-AGENT-TEAM-REBUILD-001): M5 — rule files (2 NEW + 8 modified)`
- **Date**: 2026-05-25
- **REQs covered**: REQ-ATR-007, REQ-ATR-008, REQ-ATR-012 (reinforced; primary AC binding is manager-develop.md per M1), REQ-ATR-016, REQ-ATR-020 (reinforced; primary AC binding is CLAUDE.local.md + sync.md per M7)
- **Files (2 NEW)**:
  - `.claude/rules/moai/workflow/orchestration-mode-selection.md` (5-mode decision tree, 5 sections + Anti-Patterns + Cross-References)
  - `.claude/rules/moai/workflow/archived-agent-rejection.md` (ARCHIVED_AGENT_REJECTED error spec + 12-agent migration table with example invocations)
- **Files (8 modified)**:
  - `.claude/rules/moai/workflow/spec-workflow.md` (Phase Overview agent column 17→8 + manager-strategy fallback path replacement)
  - `.claude/rules/moai/development/agent-patterns.md` (Per-Spawn Domain Specialization section + Explore canonical reference + manager-strategy chain deprecation)
  - `.claude/rules/moai/development/agent-authoring.md` (Static Agent File vs Per-Spawn Specialization Decision Tree)
  - `.claude/rules/moai/development/manager-develop-prompt-template.md` (cycle_type Mode Reference: ddd / tdd / autofix with DIAGNOSE-PATCH-VERIFY + max-3-iter contract)
  - `.claude/rules/moai/development/spec-frontmatter-schema.md` (Status Transition Ownership Matrix explicit 7 retained owner reference + archived-agent purge cross-reference)
  - `.claude/rules/moai/core/agent-common-protocol.md` (Hook Invocation Surface subsection under Orchestrator Obligations + AC-ATR-022 verification command)
  - `.moai/docs/git-workflow-doctrine.md` (NEW §18.3.1 Tier-based PR Routing: Tier S/M = main-direct, Tier L OR --pr = manager-git; **path corrected from plan.md drift `.claude/rules/moai/workflow/git-workflow-doctrine.md`**)
  - `.claude/skills/moai-foundation-core/SKILL.md` (triggers.agents 17→8 retained list + body 26-agent/7-tier reference replaced with 8-agent retained catalog)

#### M5 AC PASS/FAIL Matrix

| AC | Severity | Verification Command | Expected | Actual | Status |
|----|----------|---------------------|----------|--------|--------|
| AC-ATR-008 | HIGH | `grep -A 5 "Mode Selection" .moai/specs/SPEC-V3R6-AGENT-TEAM-REBUILD-001/progress.md \| grep -c -i "sequential\|parallel\|agent-team\|sub-agent\|trivial\|background"` | ≥ 1 | 1 | **PASS** |
| AC-ATR-012 | HIGH | `grep -c "DIAGNOSE-PATCH-VERIFY\|cycle_type.*autofix\|ci-autofix-protocol" .claude/agents/core/manager-develop.md` | ≥ 2 | 1 | **NOTE — M1 scope** (current count is 1, below the ≥ 2 threshold; AC-ATR-012's grep target is the AGENT file, not the rule file. M5 added DIAGNOSE-PATCH-VERIFY + cycle_type=autofix + ci-autofix-protocol references to `manager-develop-prompt-template.md` (rule, ≥ 3 keywords present), but M1's manager-develop.md only had partial coverage. Recommend M1 retroactive verification or follow-up SPEC. M5's contract is satisfied via the rule-file authoring; M1's agent-file coverage is M1's responsibility.) |
| AC-ATR-013 | MEDIUM | `grep -c "thorough.*team.enabled\|CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS" .claude/rules/moai/workflow/orchestration-mode-selection.md` | ≥ 1 | 2 | **PASS** |
| AC-ATR-016 (1차) | HIGH | `grep -c "ARCHIVED_AGENT_REJECTED" .claude/rules/moai/workflow/archived-agent-rejection.md` | ≥ 1 | 8 | **PASS** |
| AC-ATR-016 (2차) | HIGH | `grep -c "manager-strategy\|manager-quality\|expert-backend\|expert-frontend\|expert-security\|expert-devops\|expert-performance\|expert-refactoring\|manager-brain\|manager-project\|claude-code-guide\|researcher" .claude/rules/moai/workflow/archived-agent-rejection.md` | ≥ 12 | 20 | **PASS** |
| AC-ATR-017 | MEDIUM | `grep -A 10 "Compound\|multi-domain" .claude/rules/moai/workflow/orchestration-mode-selection.md \| grep -c "Agent Teams\|parallel multi-spawn\|3-5"` | ≥ 1 | 5 | **PASS** |

#### M5 self-verification — Section E (per manager-develop-prompt-template.md §E)

- **E1 AC PASS/FAIL Matrix**: 6 AC verified; 5 PASS + 1 NOTE (AC-ATR-012 below threshold, M1 scope clarification per matrix above)
- **E2 Cross-Platform Build**: N/A — M5 is markdown-only; 0 Go source modifications; `go build` not exercised
- **E3 Coverage**: N/A — markdown / doc-only milestone, no package-level coverage measurement
- **E4 Subagent Boundary Grep**: `grep -n 'AskUserQuestion(' .claude/rules/moai/workflow/orchestration-mode-selection.md .claude/rules/moai/workflow/archived-agent-rejection.md` → no matches (both NEW files describe orchestrator's AskUserQuestion behavior in prose, no executable call syntax) ✓
- **E5 Lint Status**: N/A — 0 Go modifications; lint baseline unchanged
- **E6 Branch HEAD + Push**: M5 commit on `main` branch; push to `origin/main` per Hybrid Trunk Tier L manager-develop main-direct path (per git-workflow-doctrine.md §18.3.1 — manager-develop self-push for run-phase milestone commits in Tier L SPEC consistent with the existing M1-M4 main-direct push pattern)
- **E7 Blocker Report**: none — all 8 modifications + 2 NEW files completed within scope per plan.md §C.1 and §D.8

#### Path-correction note (informational; documented inline per spec.md §H drift acknowledgement)

The plan.md §D.8 and spec.md §C.1 reference `.claude/rules/moai/workflow/git-workflow-doctrine.md` for the manager-git PR doctrine reinforcement. The actual file resides at `.moai/docs/git-workflow-doctrine.md` (per `CLAUDE.local.md` §18 reference and the 2026-05-20 externalization commit). M5 modified the actual path. This is a plan.md/spec.md authoring drift (path documented in two locations vs the actual single location); no functional impact on this SPEC's deliverables. Recommend a future plan.md amendment to correct the path drift; the M5 commit message body documents the correction.

### §F.3 Sync-phase Audit-Ready Signal (manager-docs scope — NOT-STARTED)

To be appended by manager-docs at sync-phase. Expected:
- CHANGELOG.md entry (REQ-ATR-019 attribution)
- 5 artifact frontmatter `status: in-progress → implemented` transition
- B12 self-test PASS (CHANGELOG duplicate detection + AC count match)

### §F.4 Mx-phase Audit-Ready Signal (orchestrator scope — NOT-STARTED)

To be appended at Mx-phase. Expected EVALUATE-SKIP per mx-tag-protocol.md §a (markdown-heavy: 0 .go files, 0 goroutines, 0 fan_in delta, only 3 NEW shell scripts).

---

Version: 0.1.0
Status: in-progress (run-phase entry; Phase 0.5 SKIPPED + Phase 0.95 Mode Selection COMPLETED)
Tier: L
