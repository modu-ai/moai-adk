# SPEC-INIT-TUX-I18N-001 — progress

## §E.1 Plan-phase Audit-Ready Signal

plan_status: audit-ready (iteration 2)
plan_complete_at: 2026-09-11

Plan 단계 산출물(manager-spec, 카드 t586, Tier L): spec.md v0.2.1(REQ 18개), plan.md, acceptance.md(AC 22개), design.md, research.md, progress.md. 워크트리 `.claude/worktrees/t586`, 브랜치 `WT-init-tux-i18n`, 착수 HEAD `18144b7aca714ea8924363b1eab4640cf101c6d0`, v0.2.1 개정 착수 HEAD `d8ebb39298b206ec8cc4183e728f27f48994df86`.

1회차 plan 감사(`.moai/reports/t586/plan-audit.md`, FAIL 0.67) 결함 D1~D17 을 반영했다. 리드 판정 Q1~Q4 로 확인 필요 표식 4건을 닫았고, 남은 표식은 없다.

v0.2.1 개정(리드 판정 Q5, 2회차 감사 전): REQ-ITI-017 제안(`agent_wiring`·`autonomy_tier` 를 `Agents & Autonomy` 한 그룹으로, init 위저드 3페이지 → 2페이지)을 확정했다. 페이지 수(AC-ITI-018)와 스테퍼 분모(AC-ITI-021)를 서로 독립인 AC 로 나눴고, 그룹 라벨 비렌더와 번역 키 없음을 AC-ITI-022 로 고정했다(측정 `research.md` §6.1). `research.md` §13 의 실행 확인 4건은 `plan.md` §F M1 착수 검증 V-a~V-d 로 옮겨 추적한다 — run 단계 증거 절(§E.2)은 run 단계 소유라 이 개정에서 자리를 만들지 않았다. 2회차 plan 감사 결과는 아직 없다. 요구 18개가 Tier M 상한 16 을 넘어 Tier L 로 올렸다(리드 판정: 분할하지 않음). Tier L plan 감사 통과 기준은 0.85 다.

t583 선행 게이트: t583(SPEC-INIT-QUIET-WIZARD-001)은 plan 단계이고 미커밋이다. `questions.go`·`wizard.go`·`types.go`·`translations.go`·`init.go` 를 건드리는 마일스톤(M4~M8)은 t583 병합·흡수와 흡수 트리 재측정 뒤에만 시작한다(`spec.md` §A.7).

SPEC ID 자기 검사:

```
$ ID="SPEC-INIT-TUX-I18N-001"; [[ "$ID" =~ ^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$ ]] && echo PASS || echo FAIL
PASS
```

제품 코드 기준선 동일성:

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
