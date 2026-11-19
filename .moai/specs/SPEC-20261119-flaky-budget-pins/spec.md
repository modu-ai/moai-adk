---
id: SPEC-20261119-flaky-budget-pins
title: "Flaky-policy slice 2 pins: the rule's placement and the 700/701 edge"
status: draft
revision: 1
created: 2026-11-19
---

# Flaky-policy slice 2 pins: the rule's placement and the 700/701 edge

## Intent

묶음 2 감사(#3225, 판정 댓글 6998272188·6998566146)의 후속 중 이 카드 범위에서 끝낼 것: flaky 규칙의
자리(merge-train 단계 안)를 핀해 규칙을 파일 끝으로 옮겨도 시험이 통과하던 구멍(N1)을 막고, 700바이트
예산의 경계를 커밋된 시험으로 읽혀요(700 통과·701 실패 — N10) — 지금의 컴파일타임 상수 단언은 "N과 N+1"이라는
주석 말과 달리 경계를 읽히지 않았어요(N11). 원 정책 SPEC(SPEC-20261119-flaky-policy)의 Plan 정정(N5)은 이미
그 개정 2에서 끝났고, 이 SPEC 은 핀만 다뤄요.

## Non-goals

- `mode-leader.md` 의 문구 수리(N2: card add 에 --ac 없음) — C-103 이 그 파일을 잡고 있어요.
- `verify.md` 의 Issue 표기 정합(N3)과 G5 의 재실행 통과 판별(N4) — 카드 범위 밖, 후속 카드.
- 정책 SPEC 본문의 요구 재작성 — 원 SPEC 은 그대로, 이 SPEC 이 슬라이스 2의 핀을 말해요.

## Requirements

- REQ-1: When the pinning suite runs, it shall fail when the flaky rule is moved out of the merge-train step to the file's end, so the rule's placement is pinned (audit N1). · failure: none
- REQ-2: The budget test shall drive the bound at N and N+1 through the one helper the check uses, so 700 passes and 701 fails in place of the compile-time-only guard (audit N10, N11).

## Acceptance

- AC-1 (REQ-1): `the_flaky_rule_sits_inside_the_merge_train_step` 가 현재 문단에서 통과하고, 규칙 블록을 절차 뒤로 옮긴 변형에서 실패해요.
  verify: `cargo test --locked --offline --test workflow_size_budget the_flaky_rule_sits_inside_the_merge_train_step`
- AC-2 (REQ-2): `flaky_rule_stays_within_its_byte_budget` 가 700 통과·701 실패를 헬퍼 호출로 읽히고, 규칙·파일 경계 단언은 그대로 유지해요.
  verify: `cargo test --locked --offline --test workflow_size_budget flaky_rule_stays_within_its_byte_budget`

## Scope

- write: `tests/workflow_size_budget.rs`, `.moai/specs/SPEC-20261119-flaky-budget-pins/spec.md`
- read-only: `skills/moai/workflows/mode-leader.md`, `.moai/specs/SPEC-20261119-flaky-policy/spec.md`
- forbidden: `mode-leader.md`·`verify.md`·`src/**` 편집(다른 카드 몫), 실제 HOME·원장

## Plan

1. 시험 먼저: 새 핀 두 개와 헬퍼를 `tests/workflow_size_budget.rs` 에 더해요. 위치 핀은 현재 문단에서 바로
   통과하는 핀(RED 해당 없음 — 기존 올은 동작의 고정), 경계 헬퍼는 기존 단언을 교체해요(700/701 은 교체 전
   단언이 없던 경계라 RED 대신 "교체 후 최초 실행"이 첫 판정이에요).
2. 원 SPEC 에서 이 슬라이스가 잠시 실린 REQ-5/6·AC-5/6 을 새 SPEC 으로 옮겨요(개정 2 커밋의 뒤정리).
3. fmt·clippy --all-targets·lib 전체·home_isolation_scan·trailer-check 배터리.

## Risks

- 위치 핀의 단계 경계 문자열(`\n10. **`)이 워크플로 편집에 흔들릴 수 있어요 — 실패 메시지가 바이트 위치를
  알려 쉽게 고쳐요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 슬라이스 2의 핀을 원 SPEC 에서 새 SPEC 으로 분리 | 리더 결정(a): 원 SPEC 을 되도록 건드리지 않음; 가지-slug 대응 훅이 같은 id 폴더를 요구 | 두 SPEC 을 다시 합치는 커밋 |
| D-2 | 컴파일타임 const 단언은 지우고 헬퍼 N/N+1 로 대체 | 감사 N11: const 단언은 경계를 읽히지 않음; clippy assertions_on_constants 와도 부딪혔던 자리 | const 가드 복원 |
