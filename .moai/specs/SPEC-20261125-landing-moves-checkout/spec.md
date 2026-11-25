---
id: SPEC-20261125-landing-moves-checkout
title: "작업 수에 상한이 생기고 card next 가 감사할 카드를 말해요"
status: draft
revision: 2
created: 2026-11-25
branch: spec/landing-moves-checkout-20261125
---

# 작업 수에 상한이 생기고 card next 가 감사할 카드를 말해요

## Intent

카드 완료 흐름 재설계의 두 번째 조각이에요(SPEC-20261124-leader-brief). 승인 근거는 사용자 결정 2026-11-24T21:52:39Z(리더 대화, 선택지로 받은 답) 4번이라 이 SPEC 은 승인을 기다리지 않아요. 이 SPEC 도 승인을 기록하지 않고 READY 에 digest 를 적어요.

revision 1 은 체크아웃된 통합 브랜치를 `card land` 가 함께 옮기는 일과 착지하면 카드가 닫히는 일도 담았어요. 앞의 것은 다른 SPEC(SPEC-20261125-land-checked-out-branch)이 착지했고, 뒤의 것은 `SPEC-20261124-land-closes-card` 가 이미 하고 있어요(`--keep-open` 이 없는 착지는 열린 기준을 PASS 판정의 증거로 `met` 으로 채우고 카드를 닫아요; 배차 노트의 "남은 기준을 withdrawn" 은 따르지 않아요, 감사가 덮은 사실을 지우기 때문이에요). 그래서 이 revision 은 남은 일 하나만 담아요.

- 작업 수 상한: 작업 중(`active`)과 검사 대기(`review`) 카드가 레인 수의 두 배에 닿으면 새 카드를 잡는 대신 검사와 착지를 도와야 해요. `card next` 는 감사가 필요한 카드를 이름으로 말해요(감사 레인용).

## Non-goals

- `card land` 와 `card land-batch` 는 바꾸지 않아요.
- 열린 채로 착지된 기존 카드를 닫기: 리더가 `card finish` 로 처리해요.
- 레인 수를 세는 새 레지스트리: 기존 세션 레지스트리를 읽기만 해요.
- 감사 대상 판단을 새로 만들지 않아요: `card.verdict` 이벤트가 없는 `review` 카드를 "감사가 필요하다"고 읽을 뿐이에요.

## Requirements

- REQ-1: IF the cards in state `active` or `review` number at least the limit, THEN `card take` shall refuse and name the cards that wait for a check or a landing. · failure: none — 원장이 그대로예요. 한도는 살아 있는 등록된 레인 수에 `work_flow.wip_limit`(기본 2)를 곱한 수이고, 살아 있는 레인이 하나도 없으면 한도를 적용하지 않아요.
- REQ-2: IF the cards in state `active` or `review` number at least the limit, THEN `card add` shall refuse unless it is given `--by user`. · failure: none — 사용자의 add 는 막지 않아요.
- REQ-3: When `work_flow.wip_limit` is `0`, the limit shall be off. · failure: none — 같은 키의 다른 숫자는 곱이에요.
- REQ-4: When `card next` finds the limit reached, it shall name the cards in `review` that carry no audit verdict and that this session did not write, in place of a card to take. · failure: none — 감사할 카드가 없으면 그 사실을 말해요(`next_is` 가 `limit`).

## Acceptance

- AC-1 (REQ-1, REQ-2, REQ-3): 한도에서 `card take` 와 `card add` 가 거절하고 대기 카드를 말하며, `--by user` 의 add 는 통과하고, `wip_limit` 0 은 한도를 꺼요 · verify: `cargo test --locked --offline --test card card_wip_limit`
- AC-2 (REQ-4): 한도에서 `card next` 가 감사가 필요한 카드를 말해요 · verify: `cargo test --locked --offline --test card card_wip_limit::next_`
- AC-3 (REQ-1, REQ-2, REQ-3, REQ-4): 이웃 테스트가 그대로 통과해요 · verify: `cargo test --locked --offline --test card card_dispatch_cli`
  verify: `cargo test --locked --offline --test card card_cli`
  verify: `cargo test --locked --offline --lib profile`

## Scope

- write: `.moai/specs/SPEC-20261125-landing-moves-checkout/spec.md`, `src/card/dispatch.rs`, `src/card/command.rs`, `src/card/work.rs`, `src/card/wip_limit.rs`, `src/card/mod.rs`, `tests/card/card_wip_limit.rs`, `tests/card/main.rs`
- shared (줄 임대): `src/profile.rs` (`wip_limit` 키)
- read-only: `src/lane_registry.rs`
- forbidden: `src/card/land.rs`, `src/card/landed.rs`, 진짜 HOME, 설치된 moai 바이너리, 다른 프로젝트, 파일 안의 카드 번호

## Plan

1. RED 먼저: AC-1·AC-2 의 한도 테스트가 맞는 이유로 실패하는 것을 봤어요(`tests/card/card_wip_limit.rs`, 임시 저장소와 임시 HOME).
2. `src/card/wip_limit.rs`: 한도와 대기 카드 목록, `take`·`add` 가 부르고 `next` 가 감사 후보를 말해요. 프로필 키 `work_flow.wip_limit`.
3. 지킴이마다 뮤턴트(최대 3줄), 이웃 테스트, clippy, READY 에서 `check affected --card`.

## Risks

- 한도는 등록된 레인 수에 기대요. 등록이 없으면 레인 수는 1 이라 한도는 2 예요.
- "넘으면"과 "닿으면": 한도는 상한이라 카드 수가 한도와 같을 때부터 새 카드를 막아요(`take`·`add`·`next`). 한도를 하나 넘은 뒤에야 막는 해석은 쓰지 않았어요.
- `card next` 의 `audit_needed` 는 `card.verdict` 이벤트만 읽어요. 이벤트 없이 노트로만 판정한 옛 카드는 계속 감사가 필요한 카드로 보여요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 한도 = `wip_limit`(기본 2) × 살아 있는 레인 수, 레인이 0 이면 한도 없음, `wip_limit` 0 은 끔 | 사용자 결정 4번(작업 중과 검사 대기를 합쳐 레인 수의 두 배)이에요 | 키를 0 으로 해요 |
| D-2 | 한도에 닿으면 막아요(카드 수 >= 한도) | 상한의 뜻이에요: 한도를 넘는 카드를 만들지 않아요 | 비교를 `>` 로 바꿔요 |
| D-3 | 레인이 하나도 등록되지 않았으면 한도를 적용하지 않아요(최소 1 레인으로 세지 않아요) | 레인을 등록하지 않고 카드를 3장 이상 잡는 기존 테스트 28개가 기본 한도에 걸렸어요(영수증 실행 4402개 중 29개 실패 중 28개). 실제 운영의 레인은 `/moai work` 가 먼저 등록하므로 한도는 그대로 걸려요 | `live_lanes` 에 최소 1 을 되돌려요 |

## Result

(구현 뒤에 적어요.)
