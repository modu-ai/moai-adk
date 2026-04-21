---
id: SPEC-SKILL-GATE-001
document: plan
version: 1.0.0
created_at: 2026-04-21
updated_at: 2026-04-21
---

# SPEC-SKILL-GATE-001: Implementation Plan

본 문서는 spec.md의 REQ를 구현하기 위한 작업 분해이다. 본 SPEC은 **문서 전용 cleanup**이며 Go 소스 코드 변경이 없다.

## Technical Approach

### 설계 원칙

- **순수 제거**: 섹션 블록을 통째로 삭제. 재작성/대체 텍스트 없음(강제 인프라는 별도 SPEC에서 다룸).
- **scope discipline**: Target Files 7개만 편집. docs-site, CLAUDE.md, Go 소스, 설정 파일은 건드리지 않음.
- **경로 보존**: phase 번호 gap(2.10, 0.05 삭제로 인한)은 의도적이며 git blame + SPEC HISTORY로 추적.
- **검증은 grep 기반**: 모든 AC가 `grep` 명령 하나로 pass/fail 판정 가능하도록 설계.

### 제거 패턴

각 파일에서 다음 패턴을 따른다:

```
### <Target Section Header>   ← 이 줄부터
...
<section body>
...
                               ← 다음 ### 헤더 바로 전까지 제거
### <Next Preserved Section>  ← 이 줄은 유지
```

workflow-modes.md는 예외: 섹션 삭제가 아니라 개별 문장(bullet 또는 문단 내 교차 참조) 제거.

## Milestone Decomposition

### M1 (Priority: High) — `/batch` 참조 4파일 제거

- **Task M1.1**: `run.md` lines 378-396 Batch Mode Decision 블록 제거 (Evaluation/Decision/Batch execution instructions 전체)
- **Task M1.2**: `clean.md` lines 131-146 Batch Mode Decision 블록 제거
- **Task M1.3**: `mx.md` lines 111-125 Batch Mode Decision 블록 제거
- **Task M1.4**: `coverage.md` lines 112-126 Batch Mode Decision 블록 제거
- **Task M1.5**: `grep -rn 'Skill("batch")' .claude/` 결과 0건 확인 (AC-1)
- **Task M1.6**: `grep -rn '### Batch Mode Decision' .claude/` 결과 0건 확인 (AC-2)

### M2 (Priority: High) — `/simplify` 참조 3파일 제거

- **Task M2.1**: `run.md` lines 698-715 Phase 2.10 Simplify Pass 블록 제거
- **Task M2.2**: `sync.md` lines 140-155 Phase 0.05 Code Simplification Review 블록 + Completion Criteria 항목 제거
- **Task M2.3**: `review.md` lines 148-160 Simplify Pass 블록 제거
- **Task M2.4**: `grep -rn 'Skill("simplify")' .claude/skills/` 결과 0건 확인 (AC-3 부분)

### M3 (Priority: Medium) — workflow-modes.md 교차 참조 정리

- **Task M3.1**: DDD IMPROVE 섹션의 "After IMPROVE: Skill(\"simplify\") executes automatically" bullet 제거
- **Task M3.2**: TDD REFACTOR 섹션의 "After REFACTOR: Skill(\"simplify\") executes automatically" bullet 제거
- **Task M3.3**: Pre-submission Self-Review 정의 문단의 "This gate runs after Skill(\"simplify\") and before completion markers" → "This gate runs before completion markers" 로 축약
- **Task M3.4**: Scope 섹션의 "Does not re-run tests (Skill(\"simplify\") already validated test passing)" bullet 제거
- **Task M3.5**: `grep -rn 'Skill("simplify")\|Phase 2.10' .claude/` 결과 0건 확인 (AC-3, AC-6)

### M4 (Priority: High) — 회귀 검증 및 커밋

- **Task M4.1**: `go vet ./...`, `go test -race ./...`, `golangci-lint run ./...` 전부 green (AC-10)
- **Task M4.2**: `make build` 성공 (embedded.go 재생성 확인)
- **Task M4.3**: SPEC v1.0.0 status=completed로 최종화 (spec.md, spec-compact.md, plan.md, acceptance.md 모두 일관)
- **Task M4.4**: 단일 docs 커밋: "docs(workflow): SPEC-SKILL-GATE-001 — /batch and /simplify references removed"

## Milestone Dependency Graph

```
M1 (batch removal, 4 files) ─┐
                             ├→ M4 (verification + commit)
M2 (simplify removal, 3 files) ─┤
                             │
M3 (workflow-modes cleanup) ─┘
```

M1/M2/M3은 서로 독립 병렬 가능(다른 파일 편집). M4는 모두 완료 후.

## Risks and Mitigations

spec.md R-1 ~ R-4에 더해 구현 관점의 추가 리스크:

- **R-5 (Edit 툴 old_string 비유일성)**: 여러 파일에 동일한 "Batch Mode Decision" 헤더가 있어 Edit이 모호해질 수 있음. **완화**: 각 파일을 독립 Edit 호출로 처리하고, old_string에 섹션 전후의 고유한 문맥(이전 섹션 헤더 + 이후 섹션 헤더)을 함께 포함.
- **R-6 (매뉴얼 Edit 중 실수)**: 7개 파일 수동 편집으로 오타/과도 삭제 가능성. **완화**: 각 Edit 후 즉시 grep으로 target 문자열 0건 + 인접 섹션(Phase 2.9, Phase 0.1 등) 존재를 확인하는 빠른 피드백 루프.

## Validation Strategy

- **Grep-first**: 모든 AC가 grep 명령으로 판정 가능 → 수동 검증 불필요.
- **Build smoke test**: `make build` 성공 → embedded.go 재생성, template deployment 정상.
- **Go test -race**: 문서 편집이지만 `.claude/skills/moai/workflows/` 경로는 internal/template/embedded 하위로도 복제되므로 템플릿 테스트(internal/template/commands_audit_test.go 등)에서 검증됨.

## Commit Strategy

본 SPEC은 단일 commit으로 충분하다. 메시지 포맷은 Conventional Commits + MoAI Context Memory 구조:

```
docs(workflow): SPEC-SKILL-GATE-001 — /batch and /simplify 참조 제거

## SPEC Reference
SPEC: SPEC-SKILL-GATE-001
Phase: SYNC
Lifecycle: Level 1 (spec-first)

/batch 는 Claude Code 바이너리에서 disableModelInvocation: true로 모델 호출이
구조적으로 불가능하고, /simplify 는 기술적으로 호출 가능하지만 91개 세션 전수
감사에서 0회 호출 — dead instruction이므로 일괄 제거.

변경 파일 7개:
- run.md: Batch Mode Decision + Phase 2.10 Simplify Pass
- sync.md: Phase 0.05 Code Simplification Review + Completion Criteria
- clean.md/mx.md/coverage.md: Batch Mode Decision 3곳
- review.md: Simplify Pass
- workflow-modes.md: DDD/TDD cycle cross-references + Pre-submission Self-Review 문구

AC-1 ~ AC-10 전수 통과. go test -race ./... green. make build 성공.
```

## Out of Scope Reminder

- `/simplify` 강제 인프라(Stop hook 등) — 별도 SPEC 필요 시점에 도입
- Claude Code 바이너리 수정 — 불가능
- Go 소스 수정 — 본 SPEC은 문서 전용
