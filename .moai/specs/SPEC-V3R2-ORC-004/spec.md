---
id: SPEC-V3R2-ORC-004
title: "Worktree MUST Rule for write-heavy role profiles"
version: "0.1.0"
status: draft
created: 2026-04-23
updated: 2026-04-23
author: GOOS
priority: P1 High
phase: "v3.0.0 — Phase 3 — Agent Cleanup"
module: ".claude/agents/moai/, .claude/rules/moai/workflow/worktree-integration.md, internal/cli/agent_lint.go, .moai/config/sections/workflow.yaml"
dependencies:
  - SPEC-V3R2-CON-001
  - SPEC-V3R2-ORC-001
  - SPEC-V3R2-ORC-002
related_problem:
  - P-A11
  - P-A22
related_theme: "Layer 4 — Orchestration, Master §4.4, §14 Worktree Isolation Rules"
breaking: false
bc_id: []
lifecycle: spec-anchored
tags: "agent, worktree, isolation, implementer, tester, designer, v3r2"
---

# SPEC-V3R2-ORC-004: Worktree MUST Rule for write-heavy role profiles

## HISTORY

| Version | Date       | Author | Description                          |
|---------|------------|--------|--------------------------------------|
| 0.1.0   | 2026-04-23 | GOOS   | Initial draft (Wave 4 SPEC writer, round 2) |

---

## 1. Goal (목적)

`.claude/rules/moai/workflow/worktree-integration.md` §HARD Rules currently says "Implementation teammates in team mode (role_profiles: implementer, tester, designer) MUST use `isolation: worktree` when spawned via Agent()" and "One-shot sub-agents making cross-file changes SHOULD use `isolation: worktree`". R5 audit identifies 6 cross-file-write agents still missing the flag: manager-ddd, manager-tdd (retired → manager-cycle in ORC-001), expert-backend, expert-frontend, expert-refactoring, researcher (problem P-A11). Additionally P-A22 flags researcher's self-contradiction — its body line 49 says "All experiments in worktree isolation when possible" but its frontmatter omits the flag.

This SPEC upgrades worktree isolation from **SHOULD** to **MUST** for all v3r2 agents classified as write-heavy (any agent that writes files across ≥3 paths per invocation, matching the team-mode write-heavy role profiles). The lint rule LR-05 is promoted from warning to error, and `internal/cli/agent_lint.go` gains an enforcement pass that rejects any write-heavy agent without `isolation: worktree`.

The rule also addresses spawn-time enforcement: when MoAI orchestrator invokes `Agent(subagent_type: "general-purpose")` for team-mode role profiles {implementer, tester, designer}, the spawn call MUST include `isolation: "worktree"`; a sanity check in the orchestrator's spawner wrapper rejects non-conforming invocations (sub-agents cannot be patched post-spawn).

### 1.1 Background

Claude Code 2.1.49+ introduced `isolation: worktree` in agent frontmatter. The mechanics: agent CWD is set to a `.claude/worktrees/<auto-name>/` ephemeral worktree; file writes land in that worktree; on agent completion, changes merge back to the parent branch via Claude Code's native worktree integration. Purpose: prevent cross-agent file-write conflicts during parallel execution.

In moai v2.13.2, team-mode write-heavy agents (implementer, tester, designer) implicitly get worktree isolation via `.moai/config/sections/workflow.yaml` role_profiles. But standalone agents (expert-backend, expert-frontend, expert-refactoring, manager-cycle, researcher) have no isolation flag — they write directly to the main worktree. When spawned in parallel (e.g., `/moai run` with two SPECs via Agent() calls in the same message), this causes silent file-write conflicts that corrupt both SPECs' implementations.

The SHOULD rule is insufficient: contributors omit the flag and the failure surfaces only as subtle git conflicts during merge. Upgrading to MUST with CI + runtime enforcement prevents reintroduction.

*Source: r5-agent-audit.md §Worktree correctness table; problem-catalog.md P-A11, P-A22; worktree-integration.md §HARD Rules; major-v3-master.md §14 Worktree Isolation Rules.*

### 1.2 Non-Goals

- Implementing the worktree creation/teardown mechanism (Claude Code native feature, managed by CC runtime).
- Changing role_profiles in `.moai/config/sections/workflow.yaml` (this SPEC reads them; SPEC-V3R2-ORC-005 owns the workflow.yaml schema).
- `isolation: remote` mode (rejected in v2 and v3r2 per Master §13 Non-Goals).
- Extending worktree to read-only agents (researcher, analyst, reviewer role profiles MUST NOT use isolation per worktree-integration.md).
- Worktree isolation for commands or skills.
- Recovery protocol for worktree creation failures (CC handles at runtime).
- Performance benchmarks of worktree vs no-worktree execution.
- Cross-worktree communication patterns (not applicable at this layer).

---

## 2. Scope (범위)

### 2.1 In Scope

- Update `.claude/rules/moai/workflow/worktree-integration.md` §HARD Rules to upgrade "SHOULD" → "MUST" for agents whose role classification is write-heavy (manager-cycle, expert-backend, expert-frontend, expert-refactoring, researcher, plus team role profiles implementer/tester/designer).
- Add `isolation: worktree` frontmatter field to the following v3r2 agents (post-ORC-001):
  - `manager-cycle.md` (replaces retired manager-ddd + manager-tdd)
  - `expert-backend.md` (modified)
  - `expert-frontend.md` (modified)
  - `expert-refactoring.md` (modified)
  - `researcher.md` (modified — aligns body claim with frontmatter per P-A22)
- Promote SPEC-V3R2-ORC-002 lint rule LR-05 from warning to error for agents matching the write-heavy classifier.
- Implement a write-heavy classifier in `internal/cli/agent_lint.go`:
  - Heuristic: any agent whose `tools:` list contains `Write` or `Edit` AND whose role hint matches `implementer|tester|designer|cycle|refactoring|backend|frontend|researcher` is classified write-heavy.
- Reject team-mode spawn calls for `role_profiles: implementer|tester|designer` that do not include `isolation: "worktree"` (CI lint over workflow.yaml configuration).
- Document prompt path rules for worktree-isolated agents (relative paths, no `cd` prefix) per existing worktree-integration.md §Prompt Path Rules; this SPEC cross-references but does not restate.
- Template-first: all agent frontmatter updates land under `internal/template/templates/.claude/agents/moai/` first; `make build` regenerates; local tree byte-identical.

### 2.2 Out of Scope (Exclusions — What NOT to Build)

- Adding `isolation: worktree` to read-only agents (manager-strategy, manager-quality, evaluator-active, plan-auditor) — worktree-integration.md HARD rule forbids this.
- Adding `isolation: worktree` to scope-limited agents (manager-docs writes to `docs-site/`, manager-git operates on the main worktree by design, manager-project writes to `.moai/project/` only).
- Changing Claude Code's worktree mechanism.
- Adding new isolation modes beyond `none | worktree`.
- Modifying tests on existing worktree CI (covered by existing SPEC-WORKTREE-001 regression tests).
- Retroactively adding isolation flags to retired v2 agents.
- Configuring git worktree pruning schedules.
- Changing `--tmux` flag behavior for `claude --worktree`.

---

## 3. Environment (환경)

- Runtime: moai-adk-go v3.0.0-alpha.3+ (Phase 3, after ORC-001 lands)
- Claude Code v2.1.97+ required (CWD isolation fix and Stop/SubagentStop hook stability); v2.1.111+ recommended
- Affected frontmatter field: `isolation: none | worktree` (enum, defaults to `none`)
- Affected v3r2 agents receiving `isolation: worktree`: manager-cycle, expert-backend, expert-frontend, expert-refactoring, researcher (5 agents)
- Affected team-mode role profiles: implementer, tester, designer (existing workflow.yaml values preserved)
- Affected rule file: `.claude/rules/moai/workflow/worktree-integration.md` §HARD Rules section
- Affected lint rule: LR-05 (promoted from warning to error for write-heavy classifier)
- Lint runtime target: <500ms per full agent tree scan (unchanged from ORC-002 budget)

---

## 4. Assumptions (가정)

- SPEC-V3R2-ORC-001 has merged; manager-cycle and builder-platform exist.
- SPEC-V3R2-ORC-002 has merged; LR-05 exists as a warning rule.
- Claude Code v2.1.97+ worktree isolation is stable (prior versions leaked agent CWD to parent session per coding-standards.md compatibility table).
- The write-heavy classifier heuristic (tools contain Write/Edit AND role hint matches) covers all agents flagged in R5 without false positives; edge cases are reviewed manually.
- manager-cycle, builder-platform are created by ORC-001 with their frontmatter templates already including the `isolation:` key (default "none"); this SPEC flips to "worktree" for manager-cycle only (builder-platform is single-file writes per R5 and stays at `none`).
- team-mode role_profiles in `workflow.yaml` already assign `isolation: worktree` for implementer, tester, designer; this SPEC does not modify those values but lints them.
- worktree-integration.md §Prompt Path Rules already covers relative-path discipline; no new rule text.
- No current consumer passes absolute paths to worktree-isolated agents (enforced at orchestrator prompt level via existing checks).

---

## 5. Requirements (EARS 요구사항)

### 5.1 Ubiquitous (항시)

**REQ-ORC-004-001 (Ubiquitous) — 규칙 업그레이드**
The file `.claude/rules/moai/workflow/worktree-integration.md` §HARD Rules **shall** declare: "Implementation agents that write files across 3 or more paths per invocation MUST use `isolation: worktree` in their frontmatter. This includes v3r2 agents manager-cycle, expert-backend, expert-frontend, expert-refactoring, researcher, and team-mode role profiles implementer, tester, designer."

**REQ-ORC-004-002 (Ubiquitous) — 대상 에이전트 플래그**
Each of the following v3r2 agents **shall** declare `isolation: worktree` in its frontmatter: manager-cycle, expert-backend, expert-frontend, expert-refactoring, researcher.

**REQ-ORC-004-003 (Ubiquitous) — 읽기전용 에이전트 제외**
The read-only agents manager-strategy, manager-quality, evaluator-active, plan-auditor **shall** NOT declare `isolation: worktree`; their `permissionMode: plan` is the authoritative safety mechanism.

**REQ-ORC-004-004 (Ubiquitous) — workflow.yaml 읽기-쓰기 역할**
The file `.moai/config/sections/workflow.yaml` role_profiles **shall** continue to declare `isolation: worktree` for implementer, tester, and designer; `isolation: none` for researcher (read-only role profile), analyst, architect, reviewer.

**REQ-ORC-004-005 (Ubiquitous) — 템플릿 동기화**
All frontmatter edits and rule edits **shall** first land in `internal/template/templates/` and then be mirrored byte-identically to the local tree via `make build` (CLAUDE.local.md §2 Template-First Rule).

### 5.2 Event-Driven (이벤트 기반)

**REQ-ORC-004-006 (Event-Driven) — 린트 승격 발동**
**When** this SPEC merges, SPEC-V3R2-ORC-002 lint rule LR-05 ("missing isolation: worktree for write-heavy role") **shall** be promoted from warning severity to error severity for agents matching the write-heavy classifier in REQ-007.

**REQ-ORC-004-007 (Event-Driven) — 쓰기 집중 분류기**
**When** the lint engine scans an agent definition, it **shall** classify the agent as write-heavy if both conditions are true: (a) the `tools:` field contains `Write` OR `Edit` OR `MultiEdit`, AND (b) either the agent name matches the regex `manager-cycle|expert-backend|expert-frontend|expert-refactoring|researcher` OR the agent's role_profile (if team-mode) is `implementer|tester|designer`.

**REQ-ORC-004-008 (Event-Driven) — 스폰-시점 검증**
**When** the MoAI orchestrator invokes `Agent(subagent_type: "general-purpose")` with a role_profile parameter, the spawn wrapper **shall** verify that the override parameter set contains `isolation: "worktree"` if the role_profile is implementer, tester, or designer; a mismatch **shall** result in a structured blocker report to the orchestrator (`ORC_WORKTREE_REQUIRED`).

### 5.3 State-Driven (상태 기반)

**REQ-ORC-004-009 (State-Driven) — 린트 검증 활성**
**While** the lint promotion (REQ-006) is active, `moai agent lint` exit code 1 **shall** occur for any write-heavy agent (per REQ-007) lacking `isolation: worktree`.

**REQ-ORC-004-010 (State-Driven) — 경로 규칙 유지**
**While** a worktree-isolated agent is invoked, the orchestrator's spawn prompt **shall** use project-root-relative paths for write-target files and **shall** NOT include `cd /absolute/path &&` prefixes in Bash commands (cross-reference to worktree-integration.md §Prompt Path Rules).

### 5.4 Optional (선택)

**REQ-ORC-004-011 (Optional) — researcher 일관성**
**Where** researcher is retained in the v3r2 roster (per Master §7.2 KEEP verdict), the agent body line claiming "All experiments in worktree isolation when possible" **shall** be updated to match the now-mandatory frontmatter declaration; body text shifts from "when possible" to a definitive statement.

**REQ-ORC-004-012 (Optional) — manager-cycle 테스트 권한**
**Where** manager-cycle executes the test-running phase (RED in TDD cycle; PRESERVE characterization tests in DDD cycle), the agent **may** use its worktree isolation to run tests without affecting the parent worktree's build artifacts.

### 5.5 Unwanted Behavior

**REQ-ORC-004-013 (Unwanted Behavior) — 쓰기집중 에이전트 플래그 누락 금지**
**If** a PR introduces or modifies a write-heavy agent (per REQ-007) without declaring `isolation: worktree`, **then** CI **shall** fail via LR-05 strict-mode error `ORC_WORKTREE_MISSING` identifying the file.

**REQ-ORC-004-014 (Unwanted Behavior) — 읽기전용 에이전트 플래그 추가 금지**
**If** a PR adds `isolation: worktree` to a read-only agent (`permissionMode: plan`), **then** CI **shall** fail via a new lint rule `ORC_WORKTREE_ON_READONLY` (added to agent-lint as part of this SPEC) because worktree overhead is unnecessary when plan mode already prevents writes (worktree-integration.md HARD rule).

**REQ-ORC-004-015 (Unwanted Behavior) — 절대 경로 프롬프트 금지**
**If** an orchestrator prompt for a worktree-isolated agent contains an absolute path to the main project directory for a write-target file, **then** the orchestrator prompt validator **shall** reject the prompt and request the path be made project-root-relative (this rule already exists in worktree-integration.md §HARD Rules; this SPEC cross-references only).

---

## 6. Acceptance Criteria (수용 기준 요약)

Detailed Given-When-Then scenarios are in `acceptance.md`.

Core criteria:

- **AC-ORC-004-01**: `.claude/rules/moai/workflow/worktree-integration.md` §HARD Rules contains the updated MUST clause listing the 5 v3r2 agents + 3 role profiles.
- **AC-ORC-004-02**: Each of manager-cycle, expert-backend, expert-frontend, expert-refactoring, researcher frontmatter contains `isolation: worktree`.
- **AC-ORC-004-03**: The read-only agents (manager-strategy, manager-quality, evaluator-active, plan-auditor) do NOT contain `isolation: worktree`.
- **AC-ORC-004-04**: `.moai/config/sections/workflow.yaml` role_profiles retains `isolation: worktree` for implementer, tester, designer.
- **AC-ORC-004-05**: `moai agent lint` on the v3r2 roster exits 0 with no LR-05 errors.
- **AC-ORC-004-06**: Removing `isolation: worktree` from expert-backend causes CI to fail with `ORC_WORKTREE_MISSING`.
- **AC-ORC-004-07**: Adding `isolation: worktree` to evaluator-active causes CI to fail with `ORC_WORKTREE_ON_READONLY`.
- **AC-ORC-004-08**: Researcher body line updated from "when possible" to definitive worktree-always statement.
- **AC-ORC-004-09**: Attempting to spawn an implementer teammate via Agent() without `isolation: "worktree"` override fails with blocker report `ORC_WORKTREE_REQUIRED`.
- **AC-ORC-004-10**: Prompt containing absolute path to main project for a worktree-isolated agent is rejected by the orchestrator prompt validator.

---

## 7. Constraints (제약)

- [HARD] FROZEN constitution preservation: `worktree-integration.md` §HARD Rules section is EVOLVABLE but the overall rule "read-only teammates MUST NOT use worktree" is FROZEN (Master §1.3 + worktree-integration.md HARD rule 2).
- [HARD] Template-First (CLAUDE.local.md §2).
- [HARD] No absolute paths in worktree-isolated agent prompts (worktree-integration.md §Prompt Path Rules, cross-reference).
- [HARD] Minimum Claude Code version 2.1.97 for worktree isolation stability (coding-standards.md compatibility table).
- [HARD] Lint runtime budget preserved (<500ms).
- [HARD] No new isolation modes (REQ-002 enum remains `none | worktree`).
- [HARD] No bypass via agent body `--isolation=none` override (spawn wrapper rejects).

---

## 8. Risks & Mitigations (리스크 및 완화)

| Risk                                                            | Impact | Mitigation                                                                                                  |
|-----------------------------------------------------------------|--------|-------------------------------------------------------------------------------------------------------------|
| Worktree creation overhead slows single-agent invocations       | MEDIUM | Only write-heavy agents get the flag; single-file writers (manager-docs, manager-git) remain at `none`       |
| Write-heavy classifier false-positive on scope-limited agents   | MEDIUM | Role-hint regex (REQ-007) is explicit; non-matching agents stay at `none`                                   |
| Worktree merge-back conflicts on overlapping file writes        | HIGH   | CC runtime handles merge; team-mode file ownership from workflow.yaml prevents overlapping writes pre-spawn |
| Researcher flag change contradicts v2 body claim                | LOW    | REQ-011 updates body text to match frontmatter                                                              |
| Legacy SPECs hardcoding `isolation: none` for what is now write-heavy | LOW    | SPEC-V3R2-MIG-001 migrator flips the flag; log the rewrite                                                  |
| Read-only agents gain flag via contributor misunderstanding     | MEDIUM | REQ-014 explicit lint rule rejects                                                                          |
| Spawn-time override path change breaks existing orchestrator    | LOW    | REQ-008 spawn-wrapper check is additive; existing correct spawns unaffected                                 |
| Claude Code worktree version mismatch on older CLI              | LOW    | `moai doctor` version check per CLAUDE.local.md §14                                                         |

---

## 9. Dependencies (의존성)

### 9.1 Blocked by

- **SPEC-V3R2-CON-001** (FROZEN/EVOLVABLE codification)
- **SPEC-V3R2-ORC-001** (17-agent roster baseline, creates manager-cycle and builder-platform)
- **SPEC-V3R2-ORC-002** (LR-05 lint rule exists as warning)

### 9.2 Blocks

- **SPEC-V3R2-ORC-005** (Dynamic team generation formalization) — references role_profiles isolation contract.
- **SPEC-V3R2-MIG-001** (migrator) — rewrites v2 agents to add `isolation: worktree` where required.

### 9.3 Related

- **SPEC-V3R2-RT-003** (Sandbox execution layer) — worktree isolation and sandbox execution are complementary (worktree = git isolation, sandbox = process isolation). Not mutually exclusive.
- **SPEC-WORKTREE-001** (legacy v2 SPEC) — continues to provide baseline worktree mechanics tests.

---

## 10. Traceability (추적성)

- REQ-to-AC mapping: REQ-001 → AC-01; REQ-002 → AC-02; REQ-003 → AC-03; REQ-004 → AC-04; REQ-005 → `diff -r` CI gate; REQ-006 → AC-05, AC-06; REQ-007 → AC-06 classifier test; REQ-008 → AC-09; REQ-009 → AC-05, AC-06; REQ-010 → AC-10; REQ-011 → AC-08; REQ-012 → manager-cycle test-run regression; REQ-013 → AC-06; REQ-014 → AC-07; REQ-015 → AC-10.
- Total REQ count: 15 (Ubiquitous 5, Event-Driven 3, State-Driven 2, Optional 2, Unwanted 3)
- Expected AC count: 10
- Wave 1/2 sources:
  - `r5-agent-audit.md` §Worktree correctness table (6 agents flagged)
  - `problem-catalog.md` P-A11 (HIGH), P-A22 (HIGH)
  - `worktree-integration.md` §HARD Rules (existing SHOULD rule upgraded here)
  - `major-v3-master.md` §14 Worktree Isolation Rules
  - `pattern-library.md` O-4 (Multi-mode router) — team mode is the primary consumer
- Code-side paths:
  - `.claude/rules/moai/workflow/worktree-integration.md` (modified, REQ-001)
  - `.claude/agents/moai/manager-cycle.md` (modified, REQ-002)
  - `.claude/agents/moai/expert-backend.md` (modified, REQ-002)
  - `.claude/agents/moai/expert-frontend.md` (modified, REQ-002)
  - `.claude/agents/moai/expert-refactoring.md` (modified, REQ-002)
  - `.claude/agents/moai/researcher.md` (modified, REQ-002, REQ-011)
  - `.moai/config/sections/workflow.yaml` (verified, REQ-004)
  - `internal/cli/agent_lint.go` (modified, REQ-007, REQ-013, REQ-014)
  - `internal/template/templates/.claude/...` (mirrors, REQ-005)
- **File:line anchors** (per D5 traceability requirement):
  - `docs/design/major-v3-master.md:L1055` (§11.4 ORC-004 definition)
  - `.moai/design/v3-redesign/synthesis/problem-catalog.md` (P-A11, P-A22)
  - `.claude/rules/moai/workflow/worktree-integration.md` (HARD Rules — existing SHOULD upgraded here)

---

End of SPEC.
