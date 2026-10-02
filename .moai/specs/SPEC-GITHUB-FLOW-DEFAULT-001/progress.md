# SPEC-GITHUB-FLOW-DEFAULT-001 — 진행 기록

## §E.1 Plan-phase Audit-Ready Signal

plan_complete_at: 2026-10-02T15:10:29Z
plan_status: audit-ready
plan_revision: 3
plan_revision_1_complete_at: 2026-10-02T12:02:24Z
plan_revision_2_complete_at: 2026-10-02T12:45:25Z

개정 3(spec 0.1.2, 마지막 감사 회차용)은 plan-audit 2회차(0.84, FAIL, 필수 기준 아홉 개 모두 PASS)의 결함 D1~D6(N-04·N-01·N-03·N-05·N-07·N-02)을 반영했고 선택 D7~D12 도 모두 반영했다. 신호를 적는 조건 — `moai spec lint SPEC-GITHUB-FLOW-DEFAULT-001`(이 트리 HEAD 에서 빌드한 `moai`) 종료 코드 0·`0 error(s), 0 warning(s)`, SPEC 디렉터리 안 표지 문자열 0, REQ 22개·AC 23개 — 은 이 개정 끝에 관측했다. REQ·AC 개수는 늘지 않았다(상한 25·25). 3회차 plan-audit 는 델타 감사(D1~D12 + 회귀 확인 + 순서 동사)로 진행된다.

개정 2(spec 0.1.1)는 plan-audit 1회차(0.74, FAIL, MP-7·MP-9 실패)의 필수 결함 F-01~F-15·F-21 과 권고 F-16~F-19·F-22·F-23 을 반영했고 F-20 은 M3(f) 삭제와 워크플로 이름 정정으로 처분했다. 신호를 적는 조건 — `moai spec lint SPEC-GITHUB-FLOW-DEFAULT-001` 종료 코드 0, SPEC 디렉터리 안 표지 문자열 0 — 은 이 개정 끝에 관측했다. 2회차 plan-audit 는 델타 감사(F 항목 + 회귀 확인)로 진행된다.

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

## §G 리더에게 하는 권고 (이 카드는 카드를 발행하지 않는다 — 카드 발행은 리더의 일)

각 항목은 권고이고 카드 id 는 **리더가 카드를 발행한 뒤 이 자리에 적는다**. 이 카드는 id 를 만들지 않는다.

| # | 권고 | 근거 | 리더가 발행한 카드 id |
|---|---|---|---|
| 1 | 병합 큐 후속 측정: 호스티드 러너 청구 분, CodeRabbit 한도 이력, 녹색 PR 뒤 적색 main 의 빈도를 재현 가능한 방식(창 고정)으로 측정하고 채택 임계를 정해 병합 큐를 다시 판정 | design D-3·D-20, `research.md` §9 | (발행 전) |
| 2 | 와이어 식별자 개명(`push_develop`·`push-develop`·`local-merge-develop`·`--develop-worktree`)과 `contract/testdata/mission_surface_baseline.txt` 의 호환 이주 | design D-12·D-21 | (발행 전) |
| 3 | 계약 모드 에스컬레이션 분류기(`internal/escalation` 의 `pushWhy` 류)가 github-flow 의 카드 브랜치 push 를 허용하도록 하는 변경. 새 행동 어휘와 git-flow 비영향 증명이 필요하며, 그 전까지 계약 모드 레인은 PR 전달의 카드 브랜치 push 에서 에스컬레이션을 만난다 | design D-23 (1회차 감사 F-12) | (발행 전) |
| 4 | develop 삭제 뒤 정리 카드: 워크플로 push 트리거의 `develop` 제거와 잔존 점검(`git grep -n -w develop -- .github/workflows` 종료 코드 1) | design D-13 단계 4, AC-GFD-023 | (발행 전) |
| 5 | develop 적색 CI 수리 카드들(절체의 선행 조건, D-17). 측정된 적색 집합은 `.moai/reports/t1453/m0-develop-ci-red.md` — 수리 여부와 분할은 리더의 판단 | design D-17 | (발행 전) |
| 6 | t1452 카드 본문을 (c) 동일 테스트 명령 재실행 억제와 PR 전 병합 준비 점검으로 좁히는 편집(리더 결정 D-11, 본문 편집은 리더 몫) | design D-11 | 해당 없음(기존 카드) |
| 7 | t810 의 닫는 처분: t1453 이 t810 을 흡수했으므로(D-19) 이 카드가 닫힐 때 운영자에게 처분을 올린다. t810 의 카드·워크트리·`SPEC-LATE-BRANCH-REDESIGN-001` 은 이 카드가 건드리지 않았다 | design D-19 | 해당 없음(기존 카드) |

## §G.1 운영자가 직접 수행하는 단계 (런북, 이 카드의 레인은 실행하지 않는다)

- `moai constitution amend` 두 번(`CONST-V3R5-027`, `CONST-V3R5-028`) — 5단째 인간 승인이 대화형 Y/N 이다(D-26).
- develop→main PR 병합(운영자·리더), develop 보호·삭제.
- main 필수 체크 목록에서 `Release PR Multi-OS Gate` 제거(D-22).
- 리더 몫: develop push 세 번(런북 0·2·4단계), 레인 정지와 재기동.
