---
id: SPEC-V3R6-PLAN-AUDITOR-GEARS-ALIGN-001
title: "plan-auditor GEARS-aware rubric 정렬 — Plan"
version: "0.1.0"
status: in-progress
created: 2026-05-25
updated: 2026-05-25
author: manager-spec
priority: P1
phase: "v3.0.0"
module: ".claude/agents/meta, internal/template/templates"
lifecycle: spec-anchored
tags: "gears, ears, plan-auditor, rubric, alignment, sprint-10, v3.0.0, plan"
tier: S
---

# Plan — SPEC-V3R6-PLAN-AUDITOR-GEARS-ALIGN-001

## 1. Goal

`.claude/agents/meta/plan-auditor.md` + 템플릿 미러를 GEARS-aware rubric 으로 정렬하여, plan-auditor 가 GEARS 신규 SPEC 과 EARS legacy SPEC 양쪽 모두에 일관된 판정을 내리도록 한다. 본 plan 은 Tier S minimal 형식 (`spec-workflow.md` § SPEC Complexity Tier) 을 따른다 — 단일 plan-phase commit 으로 2 artifact (spec.md + plan.md) 작성 후, run-phase 에서 2 file (local + mirror) 본문 편집.

### 1.1 Out of Scope

본 plan 의 실행 범위에 **포함되지 않는** 항목 (spec.md §1.2.1 Out of Scope 와 일관):

- `internal/spec/lint.go` 의 lint rule 변경 — 선행 SPEC `SPEC-V3R6-GEARS-MIGRATION-001` M2 에서 완료됨
- 4-locale docs-site `moai-plan.md` 본문 수정 — 선행 SPEC M3 에서 완료됨
- 5개 authoring guide skill 본문 수정 — `SPEC-V3R6-SKILL-GEARS-ALIGN-001` M1-M5 에서 완료됨
- 88개 기존 EARS SPEC 본문의 GEARS retroactive 재작성 — provisional `SPEC-V3R6-GEARS-SWEEP-001` 으로 deferred
- 다른 meta agent (`evaluator-active.md`, `claude-code-guide.md`) 의 EARS/GEARS 참조 정렬 — 별도 Sprint 10 cohort SPEC 으로 처리
- `.claude/agents/core/manager-spec.md` 본문 수정 — `SPEC-V3R6-SKILL-GEARS-ALIGN-001` v0.2.0 M5 에서 완료됨

## 2. Strategy

### 2.1 작업 분해 원칙

- **단일 책임**: 본 SPEC 은 plan-auditor.md (+ 미러) 만 수정한다. 다른 agent / skill / docs-site / lint engine 은 손대지 않는다 (`spec.md` §1.2.1 Out of Scope 엄격 준수).
- **Template-First Rule**: 모든 본문 편집은 local edit → 단일 방향 `cp` mirror copy 순서로 수행 (R4 risk 회피). 마지막에 `diff -q` 로 byte-identical 검증.
- **Self-dogfood as canary**: spec.md §2 REQ 자체가 GEARS 표기로 작성되어, lint engine 의 `LegacyEARSKeyword` rule 이 본 SPEC 통과 여부로 정책 일관성을 검증 (REQ-PAGA-009 / AC-PAGA-007).
- **Transitional asymmetry acknowledgment** (M0 에서 명시): 본 SPEC 의 plan-auditor 실행은 *current* (EARS-only rubric) 본문을 사용하나, 선행 SPEC `SPEC-V3R6-SKILL-GEARS-ALIGN-001` plan-auditor iter-1 PASS 0.892 가 동일 패턴의 PASS-equivalence 를 이미 증명함.

### 2.2 일관성 보호 장치 (PRESERVE invariant)

- 현재 working tree 10 entries (4 modified config + 6 untracked) PRESERVE — 본 작업이 추가하는 4 entries (spec.md + plan.md + plan-auditor.md edit + mirror edit) 외 변동 0 (AC-PAGA-008).
- 본 SPEC 의 plan-phase commit 은 spec.md + plan.md 2 artifact 만 stage; run-phase 가 plan-auditor.md + 미러 2 file 을 별도 stage.
- 88개 기존 EARS SPEC 본문 절대 수정 금지 (deferred to provisional `SPEC-V3R6-GEARS-SWEEP-001`).

### 2.3 H3 Out of Scope 패턴 준수

본 SPEC 의 §1.2.1 Out of Scope 는 h3 sub-heading 으로 작성됨 — `internal/spec/lint.go` `MissingExclusionsRule` 의 h3 pattern detection 을 만족하여 sibling baseline failure 발생 없음 (이전 `SPEC-V3R6-AGENT-RESPONSIBILITY-REALIGN-001` 부터 확립된 패턴).

## 3. Milestones

본 SPEC 의 milestone 은 Tier S minimal 형식 (4 milestone, M0 pre-flight + M1 local edit + M2 mirror parity + M3 status transition placeholder). manager-develop run-phase delegation 시 M1+M2 가 단일 spawn 으로 수행될 가능성이 높다 (~150 LOC 마크다운, 단일 파일 + 미러 copy).

### M0 — Pre-flight verification (manager-develop 진입 직전)

**Purpose**: run-phase 진입 baseline 확정 + transitional asymmetry 인지.

**Actions**:
- `git rev-parse HEAD` → run-phase entry commit SHA 기록 (run-phase 자체 commit 의 parent 가 됨)
- `git fetch origin && git rev-list --count --left-right origin/main...HEAD` → expect `0 0` (pre-spawn fetch obligation per `agent-common-protocol.md` § Pre-Spawn Sync Check)
- `grep -c "EARS" .claude/agents/meta/plan-auditor.md` → expect 14 (baseline)
- `grep -c "GEARS" .claude/agents/meta/plan-auditor.md` → expect 0 (baseline)
- `diff -q .claude/agents/meta/plan-auditor.md internal/template/templates/.claude/agents/meta/plan-auditor.md` → expect empty (mirror parity baseline)
- `wc -l .claude/agents/meta/plan-auditor.md` → expect 463 (baseline LOC)
- **Transitional asymmetry note**: 본 SPEC plan-auditor 실행은 unchanged EARS-only rubric 사용. Opus 4.7 natural-language mapping (GEARS `When`→Event-driven, `While`→State-driven, `Where`→Optional, `shall not`→Unwanted) 로 PASS-equivalence 확보. 선례: `SPEC-V3R6-SKILL-GEARS-ALIGN-001` plan-auditor iter-1 PASS 0.892.

**Exit criterion**: 6개 pre-flight check 모두 PASS, run-phase 진입 준비 완료.

### M1 — plan-auditor.md local body edits (run-phase 주요 작업)

**Purpose**: REQ-PAGA-001..007 충족 — 본문 4개 surface 편집 (MP-2 label + M3 5-pattern rubric + M2 failure-modes + cross-reference + generalized subject acknowledgment).

**Actions** (편집 순서, 의존성 없음 — 어떤 순서로든 적용 가능):

1. **MP-2 label 갱신** (REQ-PAGA-001, line 126 area):
   - Before: `**(MP-2) EARS Format Compliance**: Every acceptance criterion must match one of the five EARS patterns listed in M3.`
   - After: `**(MP-2) EARS/GEARS Format Compliance**: Every acceptance criterion must match one of the five EARS/GEARS patterns listed in M3 (legacy EARS notation valid during the 6-month backward-compatibility window).`
   - AC impact: AC-PAGA-001 (GEARS count++), AC-PAGA-002 (EARS/GEARS combined form +1)

2. **M3 Score 1.0 anchor 5-pattern 갱신** (REQ-PAGA-002 + REQ-PAGA-003 + REQ-PAGA-004, lines 58-64 area):
   - Before (5 patterns):
     ```
     - Ubiquitous: "The [system] shall [response]"
     - Event-driven: "When [trigger], the [system] shall [response]"
     - State-driven: "While [condition], the [system] shall [response]"
     - Optional: "Where [feature exists], the [system] shall [response]"
     - Unwanted: "If [undesired condition], then the [system] shall [response]"
     ```
   - After (GEARS Five Patterns canonical):
     ```
     - Ubiquitous: "The <subject> shall <behavior>" (GEARS generalizes EARS "the system" to any noun: system, component, service, agent, function, artifact)
     - Event-driven: "**When** <event-detected>, the <subject> shall <behavior>"
     - State-driven: "**While** <state>, the <subject> shall <behavior>"
     - Capability gate (Where): "**Where** <capability / feature flag / static config>, the <subject> shall <behavior>" (GEARS reframes the EARS "Optional" pattern as capability gate)
     - Unwanted: "<subject> **shall not** <action>" (GEARS canonical negative form; legacy `If [undesired condition], then [action]` form **[DEPRECATED — use shall not]** retained only within the 6-month backward-compatibility window)
     ```
   - Plus inline note: `Unified compound clause: '**Where** ... **While** ... **When** ... the <subject> shall <behavior>' — any subset of the three modifiers may chain.`
   - AC impact: AC-PAGA-001 (GEARS count++), AC-PAGA-003 (`shall not` present)

3. **M2 failure-modes list 추가** (REQ-PAGA-005, lines 41-47 area):
   - Append to existing bulleted list:
     ```
     - ACs use IF/THEN syntax without an explicit `[DEPRECATED — use shall not]` annotation, post 6-month backward-compatibility window cutoff (cohort closure date)
     ```
   - AC impact: AC-PAGA-004 (IF/THEN deprecation marker present)

4. **Cross-reference 추가** (REQ-PAGA-006):
   - Add inline near M3 rubric anchor block:
     ```
     **GEARS migration guide**: See [GEARS notation reference](docs-site/content/en/workflow-commands/moai-plan.md#gears-notation) — 4-locale (en / ko / ja / zh).
     ```
   - AC impact: AC-PAGA-005 (`#gears-notation` link present)

5. **Generalized subject + backward-compat policy paragraph** (REQ-PAGA-007 + REQ-PAGA-003, append near M3 rubric):
   - Add policy paragraph:
     ```
     **Generalized subject substitution**: GEARS replaces the EARS hardcoded "the system" subject with `<subject>`, which may be any noun (system, component, service, agent, function, artifact). The 88 existing EARS SPECs retain "The system" as the default subject for readability; the GEARS-aware rubric judges both forms as PASS-equivalent at Score 1.0.
     ```
   - AC impact: AC-PAGA-001 (GEARS count++)

**Verification** (M1 exit):
- `grep -c "GEARS" .claude/agents/meta/plan-auditor.md` ≥ 5 (AC-PAGA-001)
- `grep -c "EARS/GEARS" .claude/agents/meta/plan-auditor.md` ≥ 1 (AC-PAGA-002)
- `grep -cE "shall not|negative form" .claude/agents/meta/plan-auditor.md` ≥ 1 (AC-PAGA-003)
- `grep -cE "IF/THEN.*deprecation|backward-compatibility window" .claude/agents/meta/plan-auditor.md` ≥ 1 (AC-PAGA-004)
- `grep -cE "#gears-notation|moai-plan.*gears" .claude/agents/meta/plan-auditor.md` ≥ 1 (AC-PAGA-005)

**Estimated LOC delta**: ~70 lines added (5 surface edits, average 14 lines each net).

### M2 — Template mirror parity + verification

**Purpose**: REQ-PAGA-008 충족 — local + mirror byte-identical.

**Actions**:
- `cp -f .claude/agents/meta/plan-auditor.md internal/template/templates/.claude/agents/meta/plan-auditor.md` (single-direction overwrite)
- `diff -q .claude/agents/meta/plan-auditor.md internal/template/templates/.claude/agents/meta/plan-auditor.md` → expect empty output (AC-PAGA-006)
- `wc -c internal/template/templates/.claude/agents/meta/plan-auditor.md` → baseline 26469 + delta from M1 (~70 lines × ~50B/line ≈ +3.5KB)

**Verification** (M2 exit):
- `diff -q` empty (AC-PAGA-006 PASS)
- `git diff --cached --name-only | sort -u` shows exactly 2 paths: `.claude/agents/meta/plan-auditor.md` + `internal/template/templates/.claude/agents/meta/plan-auditor.md`

### M3 — Status transition (chore, manager-develop owned)

**Purpose**: spec.md frontmatter `status: draft → in-progress` (Status Transition Ownership Matrix per `spec-frontmatter-schema.md` § Status Transition Ownership Matrix — manager-develop owns `draft → in-progress`).

**Actions**:
- Edit `.moai/specs/SPEC-V3R6-PLAN-AUDITOR-GEARS-ALIGN-001/spec.md` frontmatter line 5: `status: draft` → `status: in-progress`
- Edit same file's frontmatter line 7: `updated: 2026-05-25` (already current; no change unless run-phase spans a day boundary)
- Edit plan.md frontmatter similarly
- This edit happens **after** M1 + M2 verification PASS within the same run-phase commit

**Verification** (M3 exit):
- `grep "^status:" .moai/specs/SPEC-V3R6-PLAN-AUDITOR-GEARS-ALIGN-001/spec.md` returns `status: in-progress`
- `grep "^status:" .moai/specs/SPEC-V3R6-PLAN-AUDITOR-GEARS-ALIGN-001/plan.md` returns `status: in-progress`

### M4 — Self-dogfood lint + final commit (run-phase exit)

**Purpose**: REQ-PAGA-009 충족 — self-lint zero LegacyEARSKeyword + zero FrontmatterInvalid.

**Actions**:
- `go run ./cmd/moai spec lint .moai/specs/SPEC-V3R6-PLAN-AUDITOR-GEARS-ALIGN-001/spec.md` → expect 0 errors
- `go run ./cmd/moai spec lint .moai/specs/SPEC-V3R6-PLAN-AUDITOR-GEARS-ALIGN-001/plan.md` → expect 0 errors
- `go run ./cmd/moai spec lint .moai/specs/SPEC-V3R6-PLAN-AUDITOR-GEARS-ALIGN-001/spec.md 2>&1 | grep -c "LegacyEARSKeyword"` → expect 0 (AC-PAGA-007 part 1)
- `go run ./cmd/moai spec lint .moai/specs/SPEC-V3R6-PLAN-AUDITOR-GEARS-ALIGN-001/spec.md 2>&1 | grep -c "FrontmatterInvalid"` → expect 0 (AC-PAGA-007 part 2)
- Pre-commit verification batch (read-only, parallel per `agent-common-protocol.md` § Parallel Execution):
  - `git status` → expect clean working tree except staged plan-auditor.md + mirror + spec.md frontmatter + plan.md frontmatter
  - `git diff --cached --name-only | sort -u | wc -l` → expect 4 (REQ-PAGA-008 + M3)
  - `git fetch origin && git rev-list --count --left-right origin/main...HEAD` → expect `0 N` where N = local ahead by plan-phase commit + run-phase commit (no parallel origin advance)
- Single run-phase commit: `feat(SPEC-V3R6-PLAN-AUDITOR-GEARS-ALIGN-001): GEARS-aware rubric body + mirror parity + status in-progress` with 🗿 MoAI trailer

**Verification** (M4 exit):
- AC-PAGA-007 PASS (0 LegacyEARSKeyword + 0 FrontmatterInvalid)
- Working tree count post-commit ≤ baseline + 4 (AC-PAGA-008)

## 4. Pre-commit Self-Verification Batch (plan-phase commit)

본 plan-phase commit (이 plan.md + spec.md 2 artifact) 직전 self-verification (orchestrator 가 manager-spec return 후 trust-but-verify 로 수행):

1. `go run ./cmd/moai spec lint .moai/specs/SPEC-V3R6-PLAN-AUDITOR-GEARS-ALIGN-001/spec.md` → 0 errors
2. `go run ./cmd/moai spec lint .moai/specs/SPEC-V3R6-PLAN-AUDITOR-GEARS-ALIGN-001/plan.md` → 0 errors
3. `grep -c "^id:\|^title:\|^version:\|^status:\|^created:\|^updated:\|^author:\|^priority:\|^phase:\|^module:\|^lifecycle:\|^tags:" spec.md` → 12 (canonical frontmatter)
4. `grep -c "^id:\|^title:\|^version:\|^status:\|^created:\|^updated:\|^author:\|^priority:\|^phase:\|^module:\|^lifecycle:\|^tags:" plan.md` → 12
5. `grep -c "#### 1.2.1 Out of Scope" spec.md` ≥ 1 (h3 pattern, MissingExclusions avoidance)
6. `grep -cE "\*\*shall\*\*|\*\*shall not\*\*" spec.md` ≥ 9 (every REQ-PAGA-001..009 has bold shall)
7. REQ count: `grep -c "^### REQ-PAGA-" spec.md` = 9
8. AC count: `grep -c "^### AC-PAGA-" spec.md` = 8 (binary AC inline per Tier S)
9. `grep -c "IF " spec.md | head` — only inside quoted deprecation discussion (R1 mitigation check)
10. `git diff --cached --name-only` shows exactly `.moai/specs/SPEC-V3R6-PLAN-AUDITOR-GEARS-ALIGN-001/spec.md` + `.../plan.md` (2 paths)

## 5. Constraints

- **DO NOT** modify any other agent body or skill body (scope strict per spec.md §1.2.1)
- **DO NOT** touch `internal/spec/lint.go` or any `_test.go` file
- **DO NOT** modify the 4 modified config files + 6 untracked items in PRESERVE list
- **DO NOT** use `## Out of Scope` h2 standalone — MUST use `#### 1.2.1 Out of Scope` h3
- **DO NOT** use snake_case frontmatter alias (`created_at` / `updated_at` / `labels` / `spec_id` all REJECTED by schema)
- **DO NOT** invoke `AskUserQuestion` from subagent (boundary HARD per `askuser-protocol.md` § Orchestrator–Subagent Boundary) — return blocker report instead
- **DO NOT** delete or modify the legacy `If [undesired condition], then [action]` form mention; convert to `[DEPRECATED — use shall not]` annotation

## 6. Commit Strategy

### 6.1 Plan-phase commit (본 SPEC 작성 직후, manager-spec return + orchestrator trust-but-verify PASS)

- **Scope**: spec.md + plan.md (2 artifacts)
- **Commit subject**: `feat(SPEC-V3R6-PLAN-AUDITOR-GEARS-ALIGN-001): plan-phase artifacts (Tier S minimal, GEARS-aware rubric scope)`
- **Strategy**: Hybrid Trunk main 직진 (per `CLAUDE.local.md` §23.7 — 1-person OSS Tier S allowed)
- **Trailer**: `🗿 MoAI <email@mo.ai.kr>`

### 6.2 Run-phase commit (manager-develop spawn, M1+M2+M3+M4)

- **Scope**: 4 paths (plan-auditor.md local + mirror + spec.md frontmatter + plan.md frontmatter)
- **Commit subject**: `feat(SPEC-V3R6-PLAN-AUDITOR-GEARS-ALIGN-001): GEARS-aware rubric body + mirror parity + status in-progress`
- **Strategy**: single run-phase spawn (M1+M2+M3 simultaneous edits feasible at ~150 LOC scale)
- **Trailer**: `🗿 MoAI <email@mo.ai.kr>`

### 6.3 Sync-phase commit (manager-docs, M5+M6 placeholder for sync stage)

- **Scope**: CHANGELOG entry + spec.md frontmatter `status: in-progress → implemented` + plan.md mirror + progress.md (if generated)
- **Commit subject**: `chore(SPEC-V3R6-PLAN-AUDITOR-GEARS-ALIGN-001): sync-phase 4-phase close (status implemented + CHANGELOG)`
- **Trailer**: `🗿 MoAI <email@mo.ai.kr>`

## 7. Anti-Patterns to Avoid

- **AP-1 — Interleaved local/mirror edits**: editing local and mirror in parallel branches of the same milestone risks divergence (R4). Always: local edits complete → single-direction `cp` → `diff -q` verify.
- **AP-2 — Deleting the legacy IF/THEN line**: the line must be **retained with `[DEPRECATED — use shall not]` annotation**, not deleted. Deletion would lose the backward-compatibility documentation contract.
- **AP-3 — Modifying meta agent siblings**: `evaluator-active.md` and `claude-code-guide.md` are out of scope (spec.md §1.2.1). Tempting to "fix while we're here" — resist; this triggers cross-attribution leakage (L46 violation).
- **AP-4 — Skipping mirror diff verification**: `cp -f` without subsequent `diff -q` assertion allows silent drift if shell expansion misfires; always assert empty diff before commit.
- **AP-5 — Adding new REQs during run-phase**: scope is locked at plan-phase. New requirements surface → return blocker report to orchestrator → iter-2 plan-phase amendment, not unilateral run-phase REQ addition.

## 8. Cross-References

- spec.md §1.2.1 Out of Scope (h3 pattern)
- spec.md §2 REQ-PAGA-001..009 (GEARS notation REQs)
- spec.md §3 AC-PAGA-001..008 (inline binary AC, Tier S)
- spec.md §4 Risks R1..R4
- `.claude/rules/moai/workflow/spec-workflow.md` § SPEC Complexity Tier (Tier S inline AC justification)
- `.claude/rules/moai/development/spec-frontmatter-schema.md` § Status Transition Ownership Matrix (manager-develop owns `draft → in-progress`)
- `.claude/rules/moai/core/agent-common-protocol.md` § Pre-Spawn Sync Check (M0 pre-flight obligation)
- `.claude/rules/moai/core/agent-common-protocol.md` § Parallel Execution (M4 verification batch)
- `.claude/rules/moai/core/askuser-protocol.md` § Orchestrator–Subagent Boundary (M1-M4 subagent constraint)
- Predecessor `SPEC-V3R6-GEARS-MIGRATION-001` spec.md (lint engine + 4-locale docs-site)
- Predecessor `SPEC-V3R6-SKILL-GEARS-ALIGN-001` spec.md (5 guide files + 5 template mirrors; transitional asymmetry empirical proof PASS 0.892)
