---
id: SPEC-SKILL-GATE-001
document: acceptance
version: 1.0.0
created_at: 2026-04-21
updated_at: 2026-04-21
---

# SPEC-SKILL-GATE-001: Acceptance Criteria

본 문서는 spec.md AC-1 ~ AC-10을 Given-When-Then 시나리오로 전개한다. 모든 AC는 grep 명령 하나로 pass/fail 판정 가능하도록 설계되었다.

## Definition of Done

구현 PR은 다음 모든 조건을 만족해야 머지 가능하다:

1. spec.md REQ-BATCH-001 ~ REQ-CONSISTENCY-002 (9 REQ) 모두 구현
2. 본 문서 AC-1 ~ AC-10 전수 통과
3. `go test -race ./...` 전체 green
4. `make build` 성공
5. Target Files 외 수정 없음(scope discipline)

---

## R1 — `/batch` 참조 제거

### Scenario AC-1: `Skill("batch")` 리터럴 전수 0건

**Given**: 리포지토리 최신 HEAD.

**When**: `grep -rn 'Skill("batch")' .claude/` 실행.

**Then**: 출력 0 라인 (exit code 1).

**AC Coverage**: REQ-BATCH-001

### Scenario AC-2: `### Batch Mode Decision` 섹션 헤더 0건

**Given**: 리포지토리 최신 HEAD.

**When**: `grep -rn '### Batch Mode Decision' .claude/` 실행.

**Then**: 출력 0 라인.

**AC Coverage**: REQ-BATCH-002

---

## R2 — `/simplify` 참조 제거

### Scenario AC-3: `Skill("simplify")` 리터럴 전수 0건

**Given**: 리포지토리 최신 HEAD.

**When**: `grep -rn 'Skill("simplify")' .claude/` 실행.

**Then**: 출력 0 라인.

**AC Coverage**: REQ-SIMPLIFY-001, REQ-SIMPLIFY-003

### Scenario AC-4: 3개 phase 헤더 동시 부재

**Given**: 리포지토리 최신 HEAD.

**When**: 다음 3개 grep 명령 순차 실행:
```
grep -n '### Phase 2.10' .claude/skills/moai/workflows/run.md
grep -n '### Phase 0.05' .claude/skills/moai/workflows/sync.md
grep -n '### Simplify Pass' .claude/skills/moai/workflows/review.md
```

**Then**: 3개 명령 모두 출력 0 라인.

**AC Coverage**: REQ-SIMPLIFY-002

### Scenario AC-5: sync.md Completion Criteria 정합성

**Given**: 리포지토리 최신 HEAD.

**When**: `grep -n 'Phase 0.05' .claude/skills/moai/workflows/sync.md` 실행.

**Then**: 출력 0 라인 (Phase 0.05 header + Completion Criteria bullet 양쪽 제거).

**AC Coverage**: REQ-SIMPLIFY-004

---

## R3 — 문서 일관성 보존

### Scenario AC-6: Phase 2.10 교차 참조 전수 제거

**Given**: 리포지토리 최신 HEAD.

**When**: `grep -rn 'Phase 2.10' .claude/` 실행.

**Then**: 출력 0 라인.

**AC Coverage**: REQ-CONSISTENCY-002

### Scenario AC-7: run.md 흐름 시각 연속성

**Given**: run.md 최신 HEAD.

**When**: Phase 2.9 섹션 다음에 나타나는 섹션을 수동 확인.

**Then**: `### LSP Quality Gates` (중간) → `### Phase 3: Git Operations (Conditional)` 순서로 연결됨. Phase 2.10 자리의 공백 또는 손상된 섹션 없음.

**AC Coverage**: REQ-CONSISTENCY-001, REQ-BATCH-003, REQ-SIMPLIFY-002

---

## Regression Defenses

### Scenario AC-8: CLAUDE.md / CLAUDE.local.md 변경 없음

**Given**: 리포지토리 최신 HEAD.

**When**: `grep -n '/batch\|/simplify\|Skill("batch")\|Skill("simplify")' CLAUDE.md CLAUDE.local.md` 실행.

**Then**: 출력 0 라인. (사전 감사에서도 0건이었으므로 이 AC는 "변경 없음" 증명).

**AC Coverage**: Target Files 경계 보존

### Scenario AC-9: 개념 수준 고아 참조 0건

**Given**: 리포지토리 최신 HEAD.

**When**: `grep -rn 'batch mode\|simplify pass' .claude/ --include="*.md" -i` 실행 (case-insensitive).

**Then**: 본 SPEC에서 제거된 개념에 해당하는 라인 0건. (다른 문맥의 "batch" 단어는 제외 — 예: "batch processing" 일반 언급은 제거되지 않음. 본 AC는 "batch mode" "simplify pass" 문자열 조합에 한정.)

**AC Coverage**: 교차 참조 정밀 검증

### Scenario AC-10: 빌드 및 테스트 회귀 없음

**Given**: 구현 완료 후 commit 상태.

**When**:
```
make build
go test -race -count=1 ./...
golangci-lint run --timeout=5m ./...
```

**Then**:
- `make build`: 성공 (embedded.go 재생성)
- `go test -race`: 전체 green
- `golangci-lint`: 0 issues

**AC Coverage**: REQ-BATCH-003, 전체 회귀 방어

---

## Quality Gate Criteria

### TRUST 5 적합성

| TRUST 항목 | 본 SPEC에서의 충족 방식 |
|-----------|----------------------|
| Tested | grep 기반 AC 10개 전수 명시적 검증, `go test -race ./...` 전체 green |
| Readable | 단순 섹션 제거로 워크플로우 가독성 향상(dead instruction 소음 제거) |
| Unified | 7개 파일의 유사 패턴(Batch Mode Decision, Simplify Pass)을 일관 제거 |
| Secured | N/A (문서 전용, 보안 표면 영향 없음) |
| Trackable | SPEC HISTORY + 단일 docs(workflow) 커밋 |

### MX Tag 요구사항

- 본 SPEC은 Go 소스를 수정하지 않으므로 @MX 태그 추가/변경 없음.

---

## Sign-off Checklist

- [x] AC-1 (`Skill("batch")` 0건) 통과
- [x] AC-2 (`### Batch Mode Decision` 0건) 통과
- [x] AC-3 (`Skill("simplify")` 0건) 통과
- [x] AC-4 (3 phase 헤더 동시 부재) 통과
- [x] AC-5 (sync.md Completion Criteria 정합성) 통과
- [x] AC-6 (`Phase 2.10` 0건) 통과
- [x] AC-7 (run.md 흐름 시각 연속성) 통과
- [x] AC-8 (CLAUDE.md 변경 없음) 통과
- [x] AC-9 (개념 수준 고아 0건) 통과
- [ ] AC-10 (빌드/테스트 회귀 없음) 통과 ← 커밋 전 최종 검증
- [ ] 단일 docs(workflow) 커밋 생성 ← 다음 단계
- [x] SPEC status=completed 반영
