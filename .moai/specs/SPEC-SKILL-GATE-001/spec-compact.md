---
id: SPEC-SKILL-GATE-001
document: spec-compact
version: 1.0.0
created_at: 2026-04-21
updated_at: 2026-04-21
source: spec.md (Requirements + Acceptance Criteria + Exclusions sections)
---

# SPEC-SKILL-GATE-001: Compact View

본 문서는 spec.md의 Requirements (EARS), Acceptance Criteria, Exclusions 세 섹션만 발췌한 컴팩트 뷰이다. 자동 생성된 참고용이며, 계약의 단일 원천은 spec.md이다.

## Requirements (EARS)

### R1 — `/batch` 참조 제거

- **REQ-BATCH-001** (Ubiquitous): `.claude/skills/moai/workflows/run.md` `clean.md` `mx.md` `coverage.md` 4개 파일에 `Skill("batch")` 리터럴 0건.
- **REQ-BATCH-002** (Ubiquitous): 동일 4파일에 `### Batch Mode Decision [MANDATORY EVALUATION]` 섹션 헤더 0건.
- **REQ-BATCH-003** (Event-driven): WHEN `run`/`clean`/`mx`/`coverage` 워크플로우가 시작될 때, THEN orchestrator는 batch-mode decision step 없이 메인 sequential phase로 직접 진입.

### R2 — `/simplify` 참조 제거

- **REQ-SIMPLIFY-001** (Ubiquitous): `run.md` `sync.md` `review.md` 3개 파일에 `Skill("simplify")` 리터럴 0건.
- **REQ-SIMPLIFY-002** (Ubiquitous): `run.md`의 `### Phase 2.10: Simplify Pass [MANDATORY]` 섹션 제거. `sync.md`의 `### Phase 0.05: Code Simplification Review` 섹션 제거. `review.md`의 `### Simplify Pass [MANDATORY EVALUATION]` 섹션 제거.
- **REQ-SIMPLIFY-003** (Ubiquitous): `.claude/rules/moai/workflow/workflow-modes.md`에 `Skill("simplify")` 리터럴 0건 + `Phase 2.10` 교차 참조 0건.
- **REQ-SIMPLIFY-004** (Ubiquitous): `sync.md`의 Completion Criteria 목록에서 `Phase 0.05: Code simplification review completed` 항목 제거.

### R3 — 문서 일관성 보존

- **REQ-CONSISTENCY-001** (Ubiquitous): 편집 후 run.md의 Phase 번호 흐름(Phase 2.9 → LSP Quality Gates → Phase 3)과 sync.md의 흐름(Phase 0 → Phase 0.1 → Phase 0.5)은 시각적 연속성 유지.
- **REQ-CONSISTENCY-002** (Ubiquitous): "Pre-submission Self-Review runs after Skill(\"simplify\")" 같은 고아 교차 참조 0건.

---

## Acceptance Criteria

- **AC-1 (R1 grep)**: `grep -rn 'Skill("batch")' .claude/` 결과 0 라인.
- **AC-2 (R1 structural)**: `grep -rn '### Batch Mode Decision' .claude/` 결과 0 라인.
- **AC-3 (R2 grep)**: `grep -rn 'Skill("simplify")' .claude/` 결과 0 라인.
- **AC-4 (R2 structural)**: `grep -n '### Phase 2.10' run.md`, `grep -n '### Phase 0.05' sync.md`, `grep -n '### Simplify Pass' review.md` 전부 0 라인.
- **AC-5 (R2 completion criteria)**: `grep -n 'Phase 0.05' sync.md` 결과 0 라인.
- **AC-6 (R3 cross-reference)**: `grep -rn 'Phase 2.10' .claude/` 결과 0 라인.
- **AC-7 (Regression shield)**: run.md에서 Phase 2.9 → LSP Quality Gates → Phase 3 시각적 연속성(수동 spot check).
- **AC-8 (CLAUDE docs)**: `grep -n '/batch\|/simplify\|Skill("batch")\|Skill("simplify")' CLAUDE.md CLAUDE.local.md` 결과 0 라인 (감사 시점 이미 0건).
- **AC-9 (Orphan)**: `grep -rn 'batch mode\|simplify pass' .claude/ --include="*.md" -i` 결과 본 SPEC 제거 개념 기준 0건.
- **AC-10 (CI/build)**: `make build` 성공, `go test -race ./...` 전체 green, `golangci-lint run` 0 issues.

---

## Exclusions (What NOT to Build)

- `/simplify` 자동 발동 강제 인프라(Stop hook, JSONL 파서, 게이트 이벤트) — 향후 별도 SPEC 필요.
- Claude Code 바이너리 수정 — 불가능, Anthropic 책임.
- 다른 MoAI 스킬(`moai-workflow-loop`, `moai-workflow-gan-loop` 등) 호출 경로 변경.
- `/moai fix`, `/moai loop` 등 다른 워크플로우 커맨드의 내부 로직 변경.
- `.moai/config/sections/*.yaml` 설정 키 추가/제거.
- Go 소스 코드 변경(`internal/hook/`, `internal/cli/` 등) — 본 SPEC은 문서만 건드림.

---

## REQ ↔ AC Traceability (9 REQs × 10 ACs)

| REQ | Primary AC | 검증 초점 |
|-----|-----------|----------|
| REQ-BATCH-001 | AC-1 | 리터럴 0건 |
| REQ-BATCH-002 | AC-2 | 섹션 헤더 0건 |
| REQ-BATCH-003 | AC-7, AC-10 | 흐름 연속성 + build |
| REQ-SIMPLIFY-001 | AC-3 | 리터럴 0건 |
| REQ-SIMPLIFY-002 | AC-4 | 3개 phase 헤더 동시 부재 |
| REQ-SIMPLIFY-003 | AC-3, AC-6 | workflow-modes.md 참조 0건 |
| REQ-SIMPLIFY-004 | AC-5 | sync.md Completion Criteria |
| REQ-CONSISTENCY-001 | AC-7 | 시각 연속성 |
| REQ-CONSISTENCY-002 | AC-6, AC-9 | 교차 참조 고아 없음 |

9/9 REQ covered. 회귀 방어는 AC-10.

### 감사 → REQ 매핑

| Warning | REQ Family | Primary AC | 상태 |
|---------|------------|-----------|------|
| `/batch` disableModelInvocation: true (바이너리 제약) | REQ-BATCH-001..003 | AC-1, AC-2, AC-7 | 제거 완료 |
| `/simplify` 91 세션 0% 발동 (behavioral) | REQ-SIMPLIFY-001..004 | AC-3, AC-4, AC-5 | 제거 완료 |
| Phase 2.10 교차 참조 고아 방지 | REQ-CONSISTENCY-001..002 | AC-6, AC-7, AC-9 | 정리 완료 |
| 빌드/테스트 회귀 방어 | — | AC-10 | 방어됨 |

2개 Dead-Instruction 영역이 REQ 패밀리로 매핑되고 AC로 검증. 회귀 방어는 AC-10으로 별도 보장.
