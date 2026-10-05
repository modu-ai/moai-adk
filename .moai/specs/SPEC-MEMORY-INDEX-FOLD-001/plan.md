# SPEC-MEMORY-INDEX-FOLD-001 — Implementation Plan

## §A Context

- 카드 t1065 (kanban queue) — 원 처방 "MEMORY.md 인덱스를 읽기 한도 이하로 접기"는 리드 세션의 09-22 병행 정비로 이미 과잉 이행됨. 본 SPEC은 그 과정에서 발생한 색인 유실 3건의 수복 + 판정 기록 + 증거 보고로 범위가 재정의됨 (Dispatch premise FIXED — 재도출 금지).
- 대상 트리: `/Users/goos/.moai/claude-profiles/moai-adk/projects/-Users-goos-MoAI-moai-adk-go/memory/` — **저장소 외부**, 절대 경로 필수.
- 워크트리: `.claude/worktrees/t1065` (branch `WT-index-fold`). 모든 저장소 내 쓰기는 이 워크트리 안.

## §B Known Issues

- **`wc -m` 로케일 함정** (plan-phase 실측): `LC_ALL=en_US.UTF-8 wc -m`이 이 머신에서 바이트 수 18456을 반환. char 권위는 `python3 len()`, 바이트는 `wc -c`로 이원화.
- **동시 편집**: 리드 세션이 `MEMORY.md`를 편집 중 (13:00:33 KST +44 bytes 실측). 베이스라인 혼용 금지.
- **측정치 만료**: dispatch의 13,408/18,412 (11:28) 및 plan-phase의 13,443/18,456 (13:1x) 모두 참고값 — REQ-MIF-001이 실행 시점에 재계측한 값이 유일한 유효 베이스라인. 급감 분해 수치(−11,490 / −614 / +2,115)는 근사치(APPROXIMATE) — `measurements-plan.md` 고지 참조.

## §C Pre-flight

- [ ] 워크트리 확인: `git rev-parse --show-toplevel` → `.claude/worktrees/t1065`
- [ ] 증거 파일 존재: `evidence/t1065_before.md`, `evidence/t1065_diff.txt` (plan-phase에서 /tmp에서 복사 완료 — 39,429 / 4,425 bytes 실측)
- [ ] 유실 3건 재확인: before-image vs current 링크 타깃 집합 diff = 정확히 3건 (plan-phase 13:1x 실측, 추가 유실 없음)
- [ ] 2차 색인 존재: `feedback_index_lessons_202609.md`, `project_card_archive_2026_09.md`

## §D Constraints

- 절대 경로 (기억 디렉터리), 순수 추가 (additive doctrine), verbatim zero-rewording, `MEMORY.md` 무편집, 카드 내 접기 금지 (드리프트 시 보고-에스컬레이션).

## §E Self-Verification

- 수용기준 AC-001..AC-005 (spec.md §3) — 전부 기계 검증 가능 (grep -cF, sha256, diff, test -f, 파일 존재).
- 검증 배치는 단일 턴 병렬 read-only Bash로 집행 (verification-claim-integrity.md §3 5섹션 형식으로 보고).

## §F Phase 4 Mode Selection

**Mode: serial**

| Input parameter | Value |
|---|---|
| Task parallelizability | 불가 — 단일 파일 쌍(2차 색인 2건)에 대한 순차 쓰기 + 쓰기 전후 해시 재획득이 순서 의존 |
| Shared-state mutation | 있음 — 기억 디렉터리 2개 파일에 쓰기, `MEMORY.md` 해시 창 보호 |
| Verification dependency | REQ-MIF-004는 REQ-MIF-003의 완료에 의존 (사후 해시는 사후 상태를 측정) |
| Failure isolation requirement | 불필요 — 마일스톤 3개, 단일 실행자 |
| Approval gate | Implementation Kickoff Approval (plan→run HUMAN GATE) 통과 후 진행 |

**Rationale**: 쓰기가 순서 의존적(해시 사전/사후 창)이고 병렬 분해 이득이 없으며, 팬아웃 대상 독립 항목이 0개 — orchestration-mode-selection.md §D 기준 serial이 유일하게 정합. fanout/agent-team/sweep 배제.

## §G Milestones (priority-based, 가역성 역순 — 변경 가능성 높은 결정 우선)

### M1 — 판정 기록 + 재계측 (Priority High)

0. `moai memory doctor` 실행 — 대상 기억 경로가 세션이 실제로 적재하는 스토어인지 확인(감사 창에서 two-stores 아티팩트 관측됨). 출력을 progress.md에 기록하고, 경로 불일치 시 즉시 리드 에스컬레이션(복원 대상이 읽히지 않는 스토어를 고치는 것이 되므로).

REQ-MIF-001 + REQ-MIF-002. 실행 시점 재계측(python3 len / wc -c 병기, ISO-8601 KST 타임스탬프), NO-ADDITIONAL-FOLD 판정과 전제-변천 근거 사슬을 progress.md와 판정서 초안에 기록. **가장 변경 가능성 높은 결정(판정)을 선두에** — 이 값이 이후 모든 마일스톤의 전제.

### M2 — 유실 3건 수복 (Priority High)

REQ-MIF-003. 사전 해시/사본 확보(`MEMORY.md` + 2개 대상 인덱스 모두) → **분기 (run 시점 유실 집합 ≠ 3건)**: 델타를 기록하고 diff 증거(`evidence/t1065_diff.txt`) 기준으로 수복을 계속한다 — 내용 보존이 지배 논리 — 중복·불일치 항목은 보고에 명시 → `Edit` 툴로 각 2차 색인 EOF에 빈 줄 분리 블록 추가 (2건 → `feedback_index_lessons_202609.md`, 1건 → `project_card_archive_2026_09.md`) → 사후 해시 재획득(3개 파일 모두). 순수 추가만, 기존 라인 무변경.

### M3 — 무손상 검증 + 보고 (Priority Medium)

REQ-MIF-004 + REQ-MIF-005. 단일 턴 병렬 read-only 배치: (a) python3 containment ×3 (백틱 위해 — 인용 셸 템플릿 금지), (b) MEMORY.md sha256 쓰기-창 전/후 비교, (c) test -f ×3, (d) 인덱스 diff = 추가 블록만. → `.moai/reports/t1065/verdict.md` 5섹션 작성 (고아 20건 + 전제-변천 서술 포함).

**동시성 프로토콜 (전 마일스톤)**: `MEMORY.md`와 2개 대상 인덱스 파일의 해시/사본을 쓰기 직전/직후 재획득(3개 파일 모두 재베이스라인 대상). 실행 중 변경 감지 → REQ-MIF-001 재베이스라인 + 델타 보고, 측정 혼용 금지.

## §H Anti-Patterns

- `wc -m`을 char 수로 인용 (이 머신에서 바이트 반환 실측됨)
- dispatch의 수치를 실행 베이스라인으로 재사용 (baseline-integrity 위반)
- grep 적중 줄을 선언 줄로 취급 (feedback_a_grep_hit_line_is_not_the_declaration_line)
- `MEMORY.md` "청소" 명분의 drive-by 편집 (scope discipline)
- 쓰기 성공 보고만으로 무손상 판정 (해시/diff 관측 없이)

## §I Cross-References

- 증거: `evidence/t1065_before.md` (정비 전 전체 이미지), `evidence/t1065_diff.txt` (§ DROPPED ENTRIES 원문)
- 판정서: `.moai/reports/t1065/verdict.md` (run-phase 산출)
- 형제 SPEC: SPEC-MEMORY-DIET-001 (MEMORY.md 다이어트 선행 작업), SPEC-AGENT-MEMORY-DRAIN-001
- 규율: verification-claim-integrity.md (5섹션 보고), orchestration-mode-selection.md §D (모드 선택)
