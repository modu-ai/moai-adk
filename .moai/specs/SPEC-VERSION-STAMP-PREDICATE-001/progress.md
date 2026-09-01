# SPEC-VERSION-STAMP-PREDICATE-001 — 진행 기록

카드 t392 · Tier M · 워크트리 `.claude/worktrees/t392`(브랜치 `WT-version-stamp-predicate`)

## §E.1 Plan-phase Audit-Ready Signal

```yaml
plan_status: audit-ready
plan_complete_at: 2026-09-01
tier: M
plan_audit_iteration: 3
plan_audit_iteration_note: "iteration 3 exists under an operator override of the Tier M two-iteration ceiling"
requirements: 15
acceptance_criteria: 15
artifacts:
  - .moai/specs/SPEC-VERSION-STAMP-PREDICATE-001/spec.md
  - .moai/specs/SPEC-VERSION-STAMP-PREDICATE-001/plan.md
  - .moai/specs/SPEC-VERSION-STAMP-PREDICATE-001/acceptance.md
  - .moai/specs/SPEC-VERSION-STAMP-PREDICATE-001/progress.md
measurement_tree: 9a3e2dabe9ab12a4a7313db2d8ab5c0247b24bb4
spec_id_regex_check: PASS
```

plan-phase에서 반증된 전제 1건: 운영자 지시가 지목한 `origin/release/v3.1.4` 는 이 검사의
RED이 **아니다**(그 트리의 권위 토큰은 `v3.1.4` 이고 낡은 golden 은 `v3.1.3` 을 담으므로
현재-토큰 스윕에 걸리지 않는다). 완결성 단언의 RED은 pin 이전의 이 트리에 있다.
근거: `spec.md` §1.1, `plan.md` §D.

### iter-2 (plan-audit FAIL 0.79 → 수리)

`.moai/reports/t392/plan-audit.md` 의 D1-D8에 답했다. 설계가 바뀐 것 둘:

1. **스윕 모집단 = 추적 파일 집합**(`git ls-files`). 파일시스템 walk를 택하지 않은 이유와
   그 결과를 실측으로 기록했다(`spec.md` §2.0). REQ-VSP-010이 「git 호출 없음」에서
   「이력·타 ref·네트워크 없음」으로 좁혀졌다.
2. **「스윕 개수 = 28」 상수 제거.** 범프 numstat 실측(`eba919e44` 6파일 전부 스탬프)으로
   그 상수가 범프마다 무너질 뿐 아니라 **애초에 틀렸음**을 확인했다 — 서술이 옛 릴리스를
   인용한 상태는 이 SPEC이 정상이라고 판정한 상태다. 상수는 등록부의 성질 둘(28 · 7)로
   옮겼고, 감사가 D5로 지목한 M4→M5 파손은 그 변경으로 **소멸**했다(`plan.md` §D.0).

plan-phase에서 반증된 감사 주장 1건: 미추적 토큰 보유 파일은 6이 아니라 **7**이며 전부 이
카드 자신의 산출물이다(`comm -23` 실측, `acceptance.md` AC-VSP-012). 형제 워크트리 수도
181이 아니라 **183**이다. 둘 다 감사 보고 자신이 예고한 드리프트다.

### iter-3 (plan-audit PASS-WITH-DEBT 0.84 → 수리, **운영자가 2회 상한 해제**)

[HARD] Tier M은 plan-audit 2회가 상한이고 iter-2가 그 두 번째였다. 이 라운드는 **운영자가
이번 한 번에 한해 해제해** 성립한 것이며 통상 경로가 아니다. `plan-audit-iter2.md` 의
N1-N6 · D6에 답했다.

설계가 바뀐 것 하나: **REQ/AC-VSP-015 — 스윕 도달 범위.** iter-2의 뺄셈(「스윕 개수 = 28」
제거)이 옳았지만, 그 상수가 스윕이 얼마나 넓게 닿았는가에 대한 **유일한 하한**이기도 했다.
크기 리터럴(감사가 제시한 10,048)은 범프에 불변이나 **개발에 불변이 아니어서** 채택하지
않고, 양변을 실행 시점에 얻는 단언 둘로 대체했다: 등록부 ⊆ 모집단 · 판정 수 = 넘긴 수
(`spec.md` §6.2). 이 둘이 못 덮는 한 칸은 §8 R-9로 이름 붙여 열어 뒀다.

나머지는 산출물 편집이다. 뿌리가 하나인 결함 둘(N2 · N3)이 있었다 — **영문 문서에 한국어
문안과 한국어 판정 grep을 못박고 있었다**(`grep -cP '[가-힣]' .moai/docs/version-management.md`
→ 0). `plan.md` §E를 영문으로 재작성하고, 판정 구절을 §E 문안에만 존재하는 영문 구절로 다시
못박았으며(현재 전부 0건 = 미리 못박은 RED), 닫힌-수 정규식을 구조형으로 바꿔 뮤턴트 여덟에
전부 걸리고 열린 문안에 0건임을 이 세션에서 실측했다.

iter-3 판정: **PASS-WITH-DEBT 0.90**(보고서 `.moai/reports/t392/plan-audit-iter3.md`, opus에서 작성 후 weekly-limit 중단으로 디스크에서 복구됨) — D1(015 카운트 키 이중 고정 `examined_of`/`handed` + §D.2 행 부재; 정본 키 `handed`로 통일)·D2(`현 L83·L90` → `현 L82·L88` 좌표 정정)는 이번 라운드에서 수리했고, D3-D5는 감사자 권고에 따라 선택적 부채로 남긴다.

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
