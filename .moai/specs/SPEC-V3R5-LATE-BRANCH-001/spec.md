---
id: SPEC-V3R5-LATE-BRANCH-001
title: "Late-Branch Workflow — main 작업 + late branch + PR squash + local main reset"
version: "0.3.0"
status: completed
created: 2026-05-20
updated: 2026-05-20
author: GOOS Kim
priority: P1
phase: "v3.5.0"
module: ".moai/config/sections/git-strategy + .claude/skills/moai/workflows/plan/spec-assembly + .claude/skills/moai/SKILL.md + .claude/agents/moai/manager-git + .claude/rules/moai/workflow/spec-workflow + internal/template/templates (5 mirrors)"
lifecycle: spec-anchored
tags: "workflow, late-branch, git-strategy, dogfooding, mega-sprint, v3r5"
---

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.3.0 | 2026-05-20 | GOOS Kim (via MoAI) | Sync-phase complete — lifecycle close. status `implemented → completed`. Single-integrated PR pattern (user decision C): plan + run + sync 3-phase 모두 main에 누적 후 1 PR로 통합 머지. docs-site 영향 없음 (workflow/config 변경만 — product/structure/tech.md 모두 영향 없음). plan-PR + run-PR + sync-PR 3-way split 대신 1 통합 PR로 simplification. Cherry-pick 패턴: 본 SPEC 9 (plan/run/chore/status) + sync 1 = 10 commits만 origin/main 기준 새 feat/SPEC-V3R5-LATE-BRANCH-001 branch로 분리. STATUSLINE-V2145-001 SPEC 4 commits은 main에 그대로 보존 (별도 PR 처리). 본 세션과 평행으로 사용자가 sync/SPEC-V3R5-STATUSLINE-V2145-001 branch에서 STATUSLINE sync 작업 진행 — main 복귀 후 본 SPEC 우선 완료 결정. Late-branch Phase A(plan)+B(run)+sync 모두 main commit, Phase C(branch+push+PR) 진입 직전. |
| 0.2.0 | 2026-05-20 | GOOS Kim (via manager-develop ddd) | Run-phase complete — status `draft → in-progress → implemented`. 7 ACs binary PASS (AC-LB-001 yq config switch / AC-LB-002 spec-assembly Phase 3 conditional / AC-LB-003 manager-git Personal Mode + Invocation Pattern / AC-LB-004 spec-workflow Step 1 precondition + Step 4 closure / AC-LB-005 template mirror parity via TestRuleTemplateMirrorDrift + TestLateBranchTemplateMirror / AC-LB-006 scripted E2E in /tmp / AC-LB-007 gh issue create gating). 11 files changed across 6 milestones + 1 catalog hash sync commit (M1 81188cad1, M2 aff3a0829, M3 826533f8b, M4 e2c8e2582, M5 8dca0384c, M6 44b8194ae, chore c9b857b7b). Self-application dogfooding succeeded — plan/run commits landed directly on main with branch deferred to Phase C. |
| 0.1.2 | 2026-05-20 | GOOS Kim (via MoAI) | REQ-LB-008 promoted from Optional → Mandatory per plan-auditor iter1 Q2 CRITICAL recommendation. Existing `internal/template/rule_template_mirror_test.go` `workflowOptMirroredPaths` covers only 9 paths, NONE matching the 4 LATE-BRANCH markdown files; without allowlist extension AC-LB-005 was vacuous. M6 delivers `lateBranchMirroredPaths` + `TestLateBranchTemplateMirror` parallel test with shared `RULE_TEMPLATE_MIRROR_DRIFT` sentinel. |
| 0.1.1 | 2026-05-20 | GOOS Kim (via MoAI) | Mid-draft policy extension — REQ-LB-009 + AC-LB-007 + EXCL-LB-008 + R-LB-005 added for "no auto GitHub Issue" policy. `issue_number` frontmatter field removed (D2 decision). 1 new affected file (`.claude/skills/moai/SKILL.md` + template mirror). Source: `feedback_no_github_issue_for_specs.md` user directive 2026-05-20. |
| 0.1.0 | 2026-05-20 | GOOS Kim (via MoAI orchestrator) | Initial draft — T2 Standard scope (4 source files + 4 template mirrors). Formalizes Late-branch workflow pattern decided 2026-05-20 (`feedback_late_branch_workflow.md`, `project_v3r5_late_branch_decision.md`). Self-applying: this SPEC's plan-phase commits land directly on `main` as dogfooding demonstration. 7 EARS REQs + 6 binary ACs + 4 Edge Cases + 4 Risks + 3 Exclusions. Backward-compatibility D-decision deferred to plan.md §D-Decisions (Option a/b/c assessment). |

---

## 1. 개요 (Overview)

### 1.1 사용자 directive (verbatim, 2026-05-20)

> "사용자는 SPEC마다 자동으로 feat/SPEC-XXX branch가 생성되는 현재 mode: team + auto_branch: true 패턴이 무겁다고 판단. main 위에서 작업 누적 후 PR 시점에만 분기하는 Late-branch 패턴을 선호."

### 1.2 Mission statement

`/moai plan`이 자동으로 `feat/SPEC-*` branch를 생성하는 현재 동작을 **opt-out** 가능하게 만들고, 사용자가 `main` 위에서 SPEC commits을 누적한 뒤 PR 시점에만 `git switch -c feat/SPEC-*`로 분기하여 squash merge 후 local main을 `origin/main`으로 재정렬하는 4-phase Late-branch 워크플로를 1급(first-class)으로 지원한다. **`mode: team`은 유지**되며, branch protection (4 required checks) + PR/CI 게이트는 변경 없이 그대로 활용된다.

### 1.3 본 SPEC의 dogfooding 자기 적용 (self-application)

본 SPEC의 **plan-phase 산출물 (spec.md, plan.md, acceptance.md, progress.md, spec-compact.md)이 직접 `main` 위에서 commit된다**. Late-branch 패턴을 SPEC 작성 시점부터 시연 (demonstration) 하기 위함이다. 이는 본 SPEC이 정의하는 패턴의 첫 실제 적용 사례이며, `plan/SPEC-V3R5-LATE-BRANCH-001` branch는 `git switch -c` 시점까지 생성되지 않는다. 사용자가 PR 생성 시점에 명시적으로 분기한다.

### 1.4 현재 상태 (empirical, main HEAD post-W3 sync)

- `.moai/config/sections/git-strategy.yaml` `team` section: `auto_branch: true`, `auto_pr: true`, `branch_creation.auto_enabled: true`, `branch_creation.prompt_always: false` (current defaults, line 50-74)
- `.claude/skills/moai/workflows/plan/spec-assembly.md`: Phase 3 (Git Environment Setup) 분기 로직 line 253-340. 현재 `branch_creation.auto_enabled` 분기 미존재 — 항상 branch creation 수행 가정
- `.claude/agents/moai/manager-git.md`: Personal Mode SPEC Git Workflow 표 존재. `main_late_branch` 옵션 부재
- `.claude/rules/moai/workflow/spec-workflow.md`: Step 1 (plan) entry condition은 "main checkout"만 명시 (line ~80). Step 4 (cleanup) cleanup procedure은 worktree disposal 중심, `git reset --hard origin/main` 미문서화
- Template mirrors (4 files under `internal/template/templates/`): byte-equivalent counterparts 존재 (`git-strategy.yaml.tmpl`, `manager-git.md`, `spec-assembly.md`, `spec-workflow.md`) — 본 SPEC은 local + template 양쪽 모두 수정 필요

### 1.5 본 SPEC의 목표 (T2 Standard scope)

1. **D1 Config switch**: `git-strategy.yaml` `team` section 4 keys 변경 (`auto_branch`/`auto_pr`/`auto_enabled`/`prompt_always`). 변경 정책 = **plan.md §D-Decisions 결정 (Option a/b/c)**.
2. **D2 Skill body**: `spec-assembly.md` Phase 3에 `auto_enabled == false` 분기 추가 (branch creation skip + cwd preserved on `main`). **추가 (v0.1.1)**: Phase 2.5 (GitHub Issue Creation) MUST default to skip; gated only by explicit `--issue` flag per REQ-LB-009.
3. **D3 Agent body**: `manager-git.md` Personal Mode 표 + Branch Management 섹션에 `main_late_branch` workflow option 추가. Late-Branch Invocation Pattern 4-phase 절차 (A→D) 문서화.
4. **D4 Rule update**: `spec-workflow.md` Step 1 entry precondition + Step 4 cleanup procedure에 Late-branch 패턴 명시. `git reset --hard origin/main` post-squash 표준 종결 단계로 등록.
5. **D5 Template mirror parity**: D1-D4의 변경을 `internal/template/templates/` 하위 4 mirror 파일에 byte-equivalent (modulo `.tmpl` placeholders) 복사. **추가 (v0.1.1)**: `.claude/skills/moai/SKILL.md`와 template mirror 5번째 페어 추가 — `--issue` flag 시맨틱 flip (opt-out via `--no-issue` → opt-in via `--issue`).
6. **D6 No migration tool**: 기존 사용자 `git-strategy.yaml` 파일의 자동 재작성은 본 SPEC scope 외. `moai update` 동작은 plan.md §D-Decisions에서 backward-compat 옵션과 함께 결정. **추가 (v0.1.1)**: `issue_number` frontmatter 필드의 retroactive 제거는 EXCL-LB-008로 명시 — 새 SPEC만 prospective 적용.

### 1.6 Brownfield State Inventory

본 SPEC은 brownfield 작업이다. 4 source 파일 + 4 template mirror 파일 모두 기존 존재하며 [EXTEND] 대상이다. [NEW] 파일은 없다.

| Marker | 파일 | 분류 | 근거 |
|--------|------|------|------|
| [EXTEND] | `.moai/config/sections/git-strategy.yaml` | EXTEND | `team` section 4 키 값 변경. 구조 추가 없음. |
| [EXTEND] | `.claude/skills/moai/workflows/plan/spec-assembly.md` | EXTEND | Phase 3에 조건 분기 추가. 기존 Phase 3.0 BODP Gate 보존. **추가 (v0.1.1)**: Phase 2.5 (GitHub Issue Creation) default skip per REQ-LB-009. |
| [EXTEND] | `.claude/skills/moai/SKILL.md` | EXTEND | **신규 (v0.1.1)**: `--issue` flag 시맨틱 flip. opt-out via `--no-issue` → opt-in via `--issue` (default = skip issue creation). |
| [EXTEND] | `.claude/agents/moai/manager-git.md` | EXTEND | Personal Mode 표에 행 추가 + Late-Branch Invocation Pattern 절 신설. |
| [EXTEND] | `.claude/rules/moai/workflow/spec-workflow.md` | EXTEND | Step 1 entry precondition + Step 4 cleanup procedure 보강. |
| [EXTEND] | `internal/template/templates/.moai/config/sections/git-strategy.yaml.tmpl` | EXTEND | local mirror — D1 변경 byte-equivalent (modulo template variables). |
| [EXTEND] | `internal/template/templates/.claude/skills/moai/workflows/plan/spec-assembly.md` | EXTEND | local mirror — D2 변경 (Phase 3 branch + Phase 2.5 default skip). |
| [EXTEND] | `internal/template/templates/.claude/skills/moai/SKILL.md` | EXTEND | **신규 (v0.1.1)**: local mirror — `--issue` flag 시맨틱 flip. |
| [EXTEND] | `internal/template/templates/.claude/agents/moai/manager-git.md` | EXTEND | local mirror — D3 변경. |
| [EXTEND] | `internal/template/templates/.claude/rules/moai/workflow/spec-workflow.md` | EXTEND | local mirror — D4 변경. |

### 1.7 Out of Scope (요약)

- 기존 사용자 프로젝트의 `git-strategy.yaml` 자동 재작성 (migration tooling)
- `main` direct push 차단을 위한 pre-push hook 구현 (별도 SPEC 후보)
- Worktree pattern 확장 (이미 SPEC-V3R4-WORKTREE에서 cover)
- Personal mode `main_direct`/`develop_direct` 옵션 활성화 (PR-less workflow는 별도 결정)

상세 enumeration은 §3.5 Exclusions 참조.

---

## 2. Context — Background and Decision History

### 2.1 결정 시점과 사용자 의도

2026-05-20 세션에서 사용자(GOOS행님, 1인 메인테이너)는 현재 `mode: team` + `auto_branch: true` 패턴이 SPEC마다 branch 전환 오버헤드를 발생시켜 "무겁다"고 판단했다. 단일 메인테이너 환경에서 SPEC 생성 → 구현 → PR 사이클이 단일 흐름이므로 branch 전환 없이 `main` 위에서 누적 작업을 수행하되, PR/CI gate는 그대로 활용하고자 한다.

### 2.2 왜 branch protection은 유지하는가

`mode: team` 유지의 핵심 이유:

- **CI signature**: branch protection 4 required checks (lint/test/race/build)가 PR마다 강제 실행됨 — 1인 메인테이너 환경에서도 self-review 보강 효과
- **Squash history**: PR squash merge로 main 히스토리가 1-SPEC = 1-commit으로 정리됨
- **PR review surface**: CI 결과, label, comment trail이 SPEC 단위로 보존됨
- **Admin override**: chicken-and-egg 상황(예: lint baseline cleanup SPEC)에서 admin-squash-merge 우회 경로 보존

Late-branch 패턴은 위 4가지 가치를 **유지**하면서 branch 전환 비용만 제거한다. branch는 PR 생성 시점까지 미생성 상태로 미뤄지며 (late = 뒤로 미룬다는 의미), 이는 SPEC 작성 동안 working tree가 main을 자연스럽게 따라간다는 의미다.

### 2.3 4-Phase 워크플로 절차

`feedback_late_branch_workflow.md`에 기록된 실행 가능 (executable) 절차:

**Phase A — SPEC 생성 (main 위에서)**
```bash
git checkout main && git pull origin main
/moai plan SPEC-XXX "description"   # → .moai/specs/SPEC-XXX/{5 files}, NO branch
git add .moai/specs/SPEC-XXX/
git commit -m "spec(SPEC-XXX): initial plan"   # main 직접 commit
```

**Phase B — Implementation (main 위에서 누적, push 안 함)**
```bash
git commit -m "🔴 RED: ..."
git commit -m "🟢 GREEN: ..."
git commit -m "♻ REFACTOR: ..."
```

**Phase C — PR 시점 분기 + push + merge**
```bash
git switch -c feat/SPEC-XXX
git push -u origin feat/SPEC-XXX
gh pr create --base main --title "..." --body "..."
# CI 통과 → admin-squash-merge OR 정상 squash merge
gh pr merge <PR-NUM> --squash --delete-branch
```

**Phase D — Local main 동기화**
```bash
git checkout main
git reset --hard origin/main   # local main → squashed remote main
git pull origin main           # 안전 확인
```

### 2.4 자기 적용 (dogfooding) 제약 (constraint of self-application)

본 SPEC의 plan-phase 자체가 이 패턴을 실행한다. `manager-spec` subagent가 5개 산출물을 `.moai/specs/SPEC-V3R5-LATE-BRANCH-001/`에 작성한 후, 사용자가 main에 직접 commit한다. 이후 plan-PR을 위해 `git switch -c plan/SPEC-V3R5-LATE-BRANCH-001`을 실행하여 분기하고 push한다. 이 절차는 본 SPEC이 구현하려는 패턴의 첫 시연이며, 동시에 본 SPEC을 작성하는 사용자가 패턴의 모든 단계를 직접 경험하는 도구이기도 하다.

### 2.5 메모리 참조 (Memory References)

- `~/.claude/projects/-Users-goos-MoAI-moai-adk-go/memory/project_v3r5_late_branch_decision.md` — 결정 요약 + paste-ready resume + EARS REQ/AC drafts (본 SPEC의 종자)
- `~/.claude/projects/-Users-goos-MoAI-moai-adk-go/memory/feedback_late_branch_workflow.md` — 워크플로 패턴 lessons + 4-phase 절차 + caveats (`main` direct push 금지, 병렬 SPEC 제약, fetch+rebase 누락 시 conflict)

---

## 3. EARS Requirements

### 3.1 Ubiquitous

**REQ-LB-001**: The system shall, when `team.branch_creation.auto_enabled == false`, ensure `/moai plan` does NOT create a `feat/SPEC-*` branch and leaves cwd on `main` (or the user's current branch if not main, with a warning).

**REQ-LB-002**: The system shall, when `auto_enabled == false`, allow `manager-spec` to commit SPEC files directly to `main` via the orchestrator's commit pipeline, without attempting to push.

### 3.2 Event-Driven

**REQ-LB-003**: When the user manually invokes `git switch -c feat/SPEC-*` after main commits, `manager-git` shall detect the Late-branch pattern via `git rev-list main..HEAD --count > 0 && git branch --show-current matches feat/SPEC-*` and emit the appropriate Phase C procedure (push + PR create + squash merge + Phase D reset).

**REQ-LB-004**: When `/moai plan` Phase 3 (Git Environment Setup) executes, the skill shall read `team.branch_creation.auto_enabled` from `git-strategy.yaml` and skip branch creation entirely when the value is `false`, surfacing the resulting cwd-on-main state as "Late-branch (main commit + late switch)" mode in the display.

### 3.3 State-Driven

**REQ-LB-005**: `manager-git` shall support `main_late_branch` as a Personal Mode SPEC Git Workflow option in its documented options table, with the 4-phase procedure (A→D) documented in a "Late-Branch Invocation Pattern" subsection under Branch Management.

**REQ-LB-006**: `spec-workflow.md` Step 4 (Cleanup) shall document `git reset --hard origin/main` as the canonical Late-branch closure step after squash merge, including the prerequisite `git checkout main && git fetch origin` and the post-condition verification `git status --porcelain` returning empty.

### 3.4 Unwanted

**REQ-LB-007**: The system shall NOT `git push origin main` from any automated path during Phase A/B/C (Late-branch flow), even if `auto_push: true` is set in `team.automation`. Push is permitted ONLY after the user manually executes `git switch -c feat/SPEC-*` in Phase C.

### 3.5 Mandatory — Template Mirror Parity (promoted from Optional at v0.1.2)

**REQ-LB-008 (Mandatory)**: The `internal/template/rule_template_mirror_test.go` test suite MUST extend to verify byte-equivalent (modulo `.tmpl` template variables) parity for the new D1-D5 field changes between `.moai/`/`.claude/` and `internal/template/templates/`. M6 delivers `lateBranchMirroredPaths` array + `TestLateBranchTemplateMirror` parallel test covering the 3 markdown mirror pairs (spec-assembly, manager-git, SKILL.md); spec-workflow.md mirror is covered by the pre-existing `workflowOptMirroredPaths` entry; the `.yaml.tmpl` mirror is verified by AC-LB-005 yq+grep check rather than byte equality due to Go template variable expansion.

**Source**: plan-auditor iter1 Q2 CRITICAL recommendation. Without promotion, AC-LB-005 verification was vacuous because the existing allowlist covered only SPEC-V3R5-WORKFLOW-OPT-001 files.

### 3.6 Mandatory — No-Auto-Issue Policy (added v0.1.1)

**REQ-LB-009 (Mandatory)**: When `/moai plan` is invoked WITHOUT the explicit `--issue` flag, the workflow MUST NOT auto-create a GitHub Issue. Issue creation is opt-in via `/moai plan --issue` only. The `manager-spec` agent prompt MUST NOT contain instructions to create GitHub issues. `spec-assembly.md` Phase 2.5 (GitHub Issue Creation) MUST be a silent skip by default.

**Maps to**: AC-LB-007 (binary verification).

**Source**: `feedback_no_github_issue_for_specs.md` user directive 2026-05-20. Sole-maintainer optimization aligned with Late-branch theme — same workflow simplification motivation as D1 (auto-branch opt-out).

---

## 4. Exclusions (What NOT to Build)

- **EXCL-LB-001**: Migration tooling that auto-rewrites existing user `git-strategy.yaml` files. `moai update` behavior is governed by plan.md §D-Decisions backward-compat option selection.
- **EXCL-LB-002**: pre-push hook implementation to block `git push origin main` direct push. This is a follow-on safety SPEC (`SPEC-V3R5-MAIN-PUSH-GUARD-*` candidate) and explicitly out of this SPEC's scope.
- **EXCL-LB-003**: Personal mode `main_direct`/`develop_direct` workflow options (PR-less workflow activation). Late-branch preserves the PR gate — these options bypass it and are a separate decision.
- **EXCL-LB-004**: Worktree pattern extension. L2/L3 worktree behavior is governed by SPEC-V3R4-WORKTREE and `worktree-integration.md`. Late-branch is orthogonal: it operates within a single main checkout, not multiple worktrees.
- **EXCL-LB-005**: Parallel multi-SPEC support on a single checkout. Late-branch explicitly restricts one-SPEC-at-a-time (per C-LB-002). Parallel SPECs require the worktree pattern.
- **EXCL-LB-006**: `mode: manual` and `mode: personal` Late-branch activation. Only `mode: team` Late-branch behavior is in scope. `manual`/`personal` `auto_enabled: false` semantics are pre-existing and not modified.
- **EXCL-LB-007**: GUI/IDE plugin integration. CLI + Claude Code orchestration only.
- **EXCL-LB-008**: Migration tool that retroactively removes `issue_number` from existing SPEC frontmatter is OUT OF SCOPE. Historical SPECs retain their `issue_number` as immutable history records. Per D2 (plan.md §2): the field-removal policy applies prospectively to new SPECs only; existing SPECs preserve their issue_number values verbatim.

---

## 5. References

- **Memory files (read in full)**: `project_v3r5_late_branch_decision.md`, `feedback_late_branch_workflow.md`
- **Affected source files**: `.moai/config/sections/git-strategy.yaml` (lines 50-74 team section), `.claude/skills/moai/workflows/plan/spec-assembly.md` (Phase 3 line 253-340), `.claude/agents/moai/manager-git.md` (Personal Mode table), `.claude/rules/moai/workflow/spec-workflow.md` (Step 1 + Step 4)
- **Template mirrors**: `internal/template/templates/.moai/config/sections/git-strategy.yaml.tmpl`, `internal/template/templates/.claude/skills/moai/workflows/plan/spec-assembly.md`, `internal/template/templates/.claude/agents/moai/manager-git.md`, `internal/template/templates/.claude/rules/moai/workflow/spec-workflow.md`
- **Related SPECs**: SPEC-V3R5-HARNESS-AUTONOMY-001 (W3, merged, structure reference), SPEC-V3R4-WORKTREE (worktree pattern boundary), SPEC-V3R2-WF-001 (workflow phase ordering FROZEN)
- **Related rules**: `.claude/rules/moai/workflow/spec-workflow.md` § SPEC Phase Discipline, `.claude/rules/moai/development/branch-origin-protocol.md` § BODP signals
- **Self-application**: This SPEC's plan-phase commits land directly on `main` as the first concrete demonstration of the pattern defined herein. `plan/SPEC-V3R5-LATE-BRANCH-001` branch will be created at PR-creation time only.
