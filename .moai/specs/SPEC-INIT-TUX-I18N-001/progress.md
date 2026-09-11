# SPEC-INIT-TUX-I18N-001 — progress

## §E.1 Plan-phase Audit-Ready Signal

plan_status: audit-ready (iteration 3)
plan_complete_at: 2026-09-11

Plan 단계 산출물(manager-spec, 카드 t586, Tier L): spec.md v0.2.2(REQ 18개), plan.md, acceptance.md(AC 22개), design.md, research.md, progress.md. 워크트리 `.claude/worktrees/t586`, 브랜치 `WT-init-tux-i18n`, 착수 HEAD `18144b7aca714ea8924363b1eab4640cf101c6d0`, v0.2.1 개정 착수 HEAD `d8ebb39298b206ec8cc4183e728f27f48994df86`, v0.2.2 개정 착수 HEAD `538b56f1923c7b72e8dcb8379d55d05e4fadb1c5`.

1회차 plan 감사(`.moai/reports/t586/plan-audit.md`, FAIL 0.67) 결함 D1~D17 을 반영했다. 리드 판정 Q1~Q4 로 확인 필요 표식 4건을 닫았고, 남은 표식은 없다.

v0.2.1 개정(리드 판정 Q5, 2회차 감사 전): REQ-ITI-017 제안(`agent_wiring`·`autonomy_tier` 를 `Agents & Autonomy` 한 그룹으로, init 위저드 3페이지 → 2페이지)을 확정했다. 페이지 수(AC-ITI-018)와 스테퍼 분모(AC-ITI-021)를 서로 독립인 AC 로 나눴고, 그룹 라벨 비렌더와 번역 키 없음을 AC-ITI-022 로 고정했다(측정 `research.md` §6.1). `research.md` §13 의 실행 확인 4건은 `plan.md` §F M1 착수 검증 V-a~V-d 로 옮겨 추적한다 — run 단계 증거 절(§E.2)은 run 단계 소유라 이 개정에서 자리를 만들지 않았다. 2회차 plan 감사는 FAIL 0.84 였다(`.moai/reports/t586/plan-audit-iter2.md`, 결함 N1~N10). 요구 18개가 Tier M 상한 16 을 넘어 Tier L 로 올렸다(리드 판정: 분할하지 않음). Tier L plan 감사 통과 기준은 0.85 다.

v0.2.2 개정(2회차 감사 결함 N1~N10 반영, 3회차 감사 전): 실제 HOME 무기록 판정을 트리 전체 매니페스트에서 코드로 도출한 감시 목록(W1~W6)으로 바꾸고 가짜 HOME 양성 대조군을 더했다(N1, `acceptance.md` §B P8·AC-ITI-020 (4), 도출은 `research.md` §14). AC-ITI-003 의 pty 판정에서 단계 표시 줄 조건을 뺐다(N2). AC-ITI-010 (4) 에 S2 양성 절·S3·S4·S6 성질 제거 뮤턴트를 더했다(N3). 부재 단정 세 곳에 같은 형태의 대조군을 붙였다(N4, `research.md` §16). tmux `-e` 로 넘기는 자식 환경 정리 목록과 실효 환경 관측을 넣었다(N5, `research.md` §15). N6~N10 은 서술을 고쳤다. REQ 18·AC 22, Tier L 유지. 3회차 plan 감사 결과는 아직 없다.

t583 선행 게이트: t583(SPEC-INIT-QUIET-WIZARD-001)은 plan 단계이고 미커밋이다. `questions.go`·`wizard.go`·`types.go`·`translations.go`·`init.go` 를 건드리는 마일스톤(M4~M8)은 t583 병합·흡수와 흡수 트리 재측정 뒤에만 시작한다(`spec.md` §A.7).

SPEC ID 자기 검사:

```
$ ID="SPEC-INIT-TUX-I18N-001"; [[ "$ID" =~ ^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$ ]] && echo PASS || echo FAIL
PASS
```

제품 코드 기준선 동일성(v0.2.1 착수 HEAD 에서 측정. 3회차 트리의 제품 코드 차이는 `research.md` §0):

```
$ git diff --stat e7a7d4bb3 HEAD -- internal cmd pkg go.mod go.sum; echo "diff-exit=$?"
diff-exit=0
```

(diff 출력 없음.)

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
