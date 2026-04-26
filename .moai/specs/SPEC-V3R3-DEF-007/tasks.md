---
spec_id: SPEC-V3R3-DEF-007
title: Task Decomposition — Convention Compliance Sweep
version: "1.0.0"
status: draft
created: 2026-04-25
related_plan: .moai/specs/SPEC-V3R3-DEF-007/plan.md
related_spec: .moai/specs/SPEC-V3R3-DEF-007/spec.md
---

# 작업 분해 — SPEC-V3R3-DEF-007

> **범례**:
> - **File owner**: 단독 소유 파일 경로
> - **Depends on**: 선행 task ID
> - **Wave**: A.1 / A.2 / A.3
> - **Parallel OK**: 동일 Wave 내 병렬 가능 여부
> - **Template + local pair**: template과 local 동시 수정 task

---

## 전체 Task 개요

| Wave | Task 수 | Parallel 가능 | Sequential 필수 |
|------|---------|---------------|-----------------|
| A.1 Skill frontmatter | 11 | 11개 모두 독립 → parallel | T-A1-1 ~ 11 |
| A.2 Agent body | 1 | — | T-A2-1 |
| A.3 Verification | 3 | T-A3-1만 독립; A3-2/3 sequential | T-A3-1 ~ 3 |

**총 task 수: 15**

---

## Wave A.1 — Skill Frontmatter Sweep (11 tasks, parallel OK)

### T-A1-1: moai-domain-backend frontmatter 추가
- **File owner**: `.claude/skills/moai-domain-backend/SKILL.md`, `internal/template/templates/.claude/skills/moai-domain-backend/SKILL.md`
- **Depends on**: 없음
- **Parallel OK**: Yes
- **Action**: 두 파일의 metadata 블록 직후에 progressive_disclosure 블록 삽입
- **Verification**: `grep -c "progressive_disclosure:" <path>` = 1
- **Rollback**: git diff 부분 revert

### T-A1-2: moai-domain-frontend frontmatter 추가
- **File owner**: `.claude/skills/moai-domain-frontend/SKILL.md`, template pair
- **Depends on**: 없음
- **Parallel OK**: Yes
- 이하 동일 패턴

### T-A1-3: moai-domain-db-docs frontmatter 추가
- **File owner**: `.claude/skills/moai-domain-db-docs/SKILL.md`, template pair
- **Parallel OK**: Yes

### T-A1-4: moai-formats-data frontmatter 추가
- **File owner**: `.claude/skills/moai-formats-data/SKILL.md`, template pair
- **Parallel OK**: Yes

### T-A1-5: moai-framework-electron frontmatter 추가
- **File owner**: `.claude/skills/moai-framework-electron/SKILL.md`, template pair
- **Parallel OK**: Yes

### T-A1-6: moai-library-mermaid frontmatter 추가
- **File owner**: `.claude/skills/moai-library-mermaid/SKILL.md`, template pair
- **Parallel OK**: Yes

### T-A1-7: moai-library-nextra frontmatter 추가
- **File owner**: `.claude/skills/moai-library-nextra/SKILL.md`, template pair
- **Parallel OK**: Yes

### T-A1-8: moai-library-shadcn frontmatter 추가
- **File owner**: `.claude/skills/moai-library-shadcn/SKILL.md`, template pair
- **Parallel OK**: Yes

### T-A1-9: moai-tool-ast-grep frontmatter 추가
- **File owner**: `.claude/skills/moai-tool-ast-grep/SKILL.md`, template pair
- **Parallel OK**: Yes

### T-A1-10: moai-workflow-ddd frontmatter 추가
- **File owner**: `.claude/skills/moai-workflow-ddd/SKILL.md`, template pair
- **Parallel OK**: Yes

### T-A1-11: moai-workflow-loop frontmatter 추가
- **File owner**: `.claude/skills/moai-workflow-loop/SKILL.md`, template pair
- **Parallel OK**: Yes

### Wave A.1 Checkpoint
- 22개 파일 모두 progressive_disclosure 블록 포함 확인 (grep loop)
- AC-DEF007-01 verification

---

## Wave A.2 — Agent Body Update (1 task)

### T-A2-1: manager-git body에 Scope Boundaries + Delegation Protocol 섹션 추가
- **File owner**: `.claude/agents/moai/manager-git.md`, `internal/template/templates/.claude/agents/moai/manager-git.md`
- **Depends on**: 없음
- **Parallel OK**: — (단일 task)
- **Action**:
  1. `manager-git.md` body의 Primary Mission 또는 Core Capabilities 섹션 위치 확인
  2. 직후에 `## Scope Boundaries` 섹션 삽입 (IN SCOPE: git workflow / OUT OF SCOPE: SPEC writing, code, doc sync)
  3. 이어서 `## Delegation Protocol` 섹션 삽입 (manager-spec, manager-ddd/tdd, manager-docs, manager-quality 위임 규칙)
  4. Template + local 양쪽 동일 적용
- **Verification**: 두 파일 모두 두 섹션 grep 통과
- **Rollback**: git diff revert

### Wave A.2 Checkpoint
- AC-DEF007-02 verification

---

## Wave A.3 — Verification (3 tasks)

### T-A3-1: Diff 검증 (additive only)
- **File owner**: 검증 보고서 (콘솔 출력)
- **Depends on**: T-A1-* 모두, T-A2-1 완료
- **Parallel OK**: — (선행 의존)
- **Action**:
  - `git diff -U0 .claude/skills/ internal/template/templates/.claude/skills/ .claude/agents/moai/manager-git.md internal/template/templates/.claude/agents/moai/manager-git.md`
  - 추가 라인(`+`)만 존재, 삭제 라인(`-`) 없음 확인 (whitespace 제외)
- **Verification**: AC-DEF007-03 통과

### T-A3-2: make build 실행
- **File owner**: `internal/template/embedded.go` (자동 갱신)
- **Depends on**: T-A3-1
- **Parallel OK**: No
- **Action**: `make build` 실행, exit code 0 확인
- **Verification**: AC-DEF007-04 통과

### T-A3-3: go test 회귀 검증
- **Depends on**: T-A3-2
- **Parallel OK**: No
- **Action**: `go test -count=1 ./internal/template/...` 실행
- **Verification**: AC-DEF007-05 통과 (모든 테스트 PASS)

### Wave A.3 Checkpoint
- AC-DEF007-01 ~ 05 모두 통과
- DoD 모두 충족

---

## Edge Case Tasks (조건부)

### T-EC-1: 기존 progressive_disclosure 블록 부분 존재 처리
- **Trigger**: T-A1-* 실행 중 grep으로 기존 블록 일부 발견 시
- **Action**: 누락된 필드만 추가, 블록 전체 중복 삽입 금지
- **Verification**: 블록이 1개만 존재 (`grep -c` = 1)

### T-EC-2: manager-git에 이미 Scope Boundaries 존재 처리
- **Trigger**: T-A2-1 실행 전 grep으로 이미 존재 확인
- **Action**: 해당 파일 skip, 콘솔 보고 "Already compliant"
- **Verification**: 중복 섹션 없음

### T-EC-3: Template / local pre-existing drift 처리
- **Trigger**: T-A1-* / T-A2-1 실행 전 두 파일 diff 확인
- **Action**: drift 발견 시 sweep 중단, operator에게 surface
- **Verification**: 사용자 명시적 승인 후 진행
