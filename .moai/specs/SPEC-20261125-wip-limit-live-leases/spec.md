---
id: SPEC-20261125-wip-limit-live-leases
title: "The work-in-progress limit counts only the cards whose lease is alive, and a refusal names the cards left out"
status: draft
revision: 1
created: 2026-11-25
---

# The work-in-progress limit counts only the cards whose lease is alive, and a refusal names the cards left out

## Interview

- Q1: 작업 수 상한은 어떤 카드를 세야 하나요? 임대가 끝난 오래된 카드까지 세어 `card add` 가 막혀요. → A: "임대가 살아 있는 카드만 세기 (Recommended)"(사용자 선택 2026-11-25T05:54:10Z, 카드 노트 9024). 다른 선택지는 (나) HOLD 와 별개의 '보류 중' 표시를 두기, (다) 레인당 배수를 프로필에서 바꾸기였어요.
- Q2: 상한 자체(레인 수의 두 배, 사용자 결정 2026-11-24)를 바꾸나요? → A: 아니요. 세는 대상만 바뀌고 `wip_limit` 배수와 "0 이면 끔"은 그대로예요.
- Q3: 세지 않는 카드를 어떻게 알리나요? → A: 거절 문구와 `card next` 가 임대가 끝난 카드를 따로 적어요(카드 본문의 기준 2).

## Intent

다른 저장소 실측(2026-11-25, 카드 본문): `card add refused: the work-in-progress limit is reached: 11 cards are in work or in review and the limit is 2 (2 per lane, 1 lane(s))`. 11장 중 7장은 임대가 끝난 지 오래된 `active`·`review` 카드(사용자가 "열린 채 둔다"고 한 것 포함)였고, 새 카드도 `--split-of` 분할 카드도 만들 수 없었어요.

읽은 코드(`src/card/wip_limit.rs`, 2026-11-25): `in_work` 는 `state IN ('active', 'review')` 인 카드를 모두 세고, `Reading::at_limit` 은 그 수가 `wip_limit` 곱하기 살아 있는 레인 수 이상이면 참이에요. 임대(`reservation.lease_until`)를 보지 않아요. `reservation` 표는 카드마다 generation 별 행을 갖고, `status` 가 `held`, `held_for_review`, `expired_candidate`, `released`, `reclaimed` … 중 하나예요(`src/ledger/schema.rs:112`).

이 SPEC 이 하는 일: 상한이 임대가 끝나지 않은 현재 generation 의 예약을 가진 카드만 세요. 세지 않은 카드는 거절 문구와 `card next` 에 따로 적어요. 살아 있는 작업이 상한을 넘으면 지금처럼 거절해요.

## Non-goals

- 상한 배수(`work_flow.wip_limit`)와 "0 이면 끔", "살아 있는 레인이 없으면 끔"을 바꾸지 않아요.
- 임대가 끝난 카드를 닫거나 풀지 않아요(`card reclaim` 의 몫).
- `card next` 가 감사 대기 카드를 고르는 방식(`audit_needed`)을 바꾸지 않아요.
- 통합 브랜치 보호 훅이나 착지 게이트를 바꾸지 않아요.

## Requirements

<!-- This SPEC changes which cards the work-in-progress gate counts (a gate that refuses `card take` and `card add`); a card with a live lease is counted exactly as before, so the gate is not weakened for live work. -->
- REQ-1: The work-in-progress limit shall count an `active` or `review` card only when its reservation of the card's current generation has a lease that has not ended. · failure: none — a card with no live lease is no live work, and REQ-3 lists it.
- REQ-2: When the cards counted by REQ-1 reach the limit, `card take` and `card add` shall refuse as they do today. · failure: none — the refusal is the behavior, and REQ-4 names the cards.
- REQ-3: A card in work or in review that REQ-1 leaves out shall be listed apart from the counted cards with the end time of its lease or the words `no reservation`. · failure: none — a card cannot be left out unlisted.
- REQ-4: The refusal text shall name the counted cards and name the cards left out by REQ-1 in a separate part. · failure: none — the text is an error answer, never a reservation change.
- REQ-5: The `wip` object of `card next` shall carry the count of counted cards and the list of cards left out. · failure: none — an old reader that ignores the new keys keeps the old keys.
- REQ-6: The limit shall stay off when `work_flow.wip_limit` is 0 or when no registered lane is live. · failure: none — the same two off-switches as before.

## Acceptance

- AC-1 (REQ-1): in a temporary ledger with several `active` and `review` cards whose leases ended and a live lane, a new `card add` by a session succeeds · verify: `cargo test --offline --locked --test card_wip_live_lease ac1_`
- AC-2 (REQ-1, REQ-2): live cards that reach the limit make `card add` and `card take` refuse, and one card more below the limit succeeds (guard against counting too little) · verify: `cargo test --offline --locked --test card_wip_live_lease ac2_`
- AC-3 (REQ-3, REQ-4): the refusal text names the counted cards and, apart from them, each card left out with its lease end time or `no reservation` · verify: `cargo test --offline --locked --test card_wip_live_lease ac3_`
- AC-4 (REQ-5): `card next` answers `wip` with the counted number and the left-out list, and a card left out stays out of the limit while an `audit_needed` entry stays as before · verify: `cargo test --offline --locked --test card_wip_live_lease ac4_`
- AC-5 (REQ-6): `wip_limit` 0 and a ledger with no live lane stay off · verify: `cargo test --offline --locked --test card_wip_live_lease ac5_`

## Decisions

각 결정은 되돌릴 수 있고 분류는 none 이에요. 추천이 맨 앞이에요. D-1 은 사용자가 정했어요.

- D-1 (세는 대상, 사용자 결정 노트 9024): 임대가 살아 있는 카드만 세요(가). 나·다는 고르지 않았어요.
- D-2 (예약이 아예 없는 `active`·`review` 카드): (A, 기본·권장) 세지 않고 `no reservation` 으로 따로 적어요(살아 있는 작업이 아니에요). (B) 세요(원장이 어긋난 카드를 눈에 띄게 하지만 상한을 막아요).
- D-3 (임대가 살아 있는 `review` 카드): (A, 기본·권장) 센다 — 감사를 기다리는 동안 작성자의 임대가 살아 있으면 그 카드는 살아 있는 작업이에요. (B) `review` 는 임대와 상관없이 세지 않아요.
- D-4 (이 SPEC 의 코드 범위): `src/card/wip_limit.rs` 한 파일이에요. 카드 코드를 고치는 다른 카드가 있으므로 구현은 `card check` 가 허락하는 때에 시작해요.

## Scope

- write: `.moai/specs/SPEC-20261125-wip-limit-live-leases/**`, `src/card/wip_limit.rs`, `tests/card_wip_live_lease.rs`
- write (한 줄, line-lease): `tools/repo_scans.json`(새 시험 파일 등록)
- read-only: `src/card/work.rs`(`card next` 가 읽기만 해요), `src/card/dispatch.rs`, `src/card/command.rs`, `src/ledger/schema.rs`
- forbidden: 상한 배수·끄는 규칙을 바꾸는 변경, 게이트를 느슨하게 하는 변경, 전역 설치본과 실제 HOME, 네트워크, Python

## Plan

1. RED: `tests/card_wip_live_lease.rs`(AC-1..AC-5)가 지금 코드에서 실패해요. 임시 원장에 임대 끝난 카드 여럿과 살아 있는 카드를 만들어요(실제 원장과 HOME 을 안 써요).
2. `in_work` 를 현재 generation 예약의 임대 시각까지 읽게 하고 `Reading` 이 세는 카드와 뺀 카드를 따로 담아요(REQ-1, REQ-3).
3. `refusal` 문구와 `Reading::json` 에 뺀 카드를 적어요(REQ-4, REQ-5). GREEN.
4. 가드 뮤턴트(살아 있는 카드를 세지 않게 하는 변경이 AC-2 에서 죽어요), `moai check affected --card` 영수증, READY.

## Risks

- R-1: 임대 시각을 잘못 읽으면 살아 있는 카드를 안 세서 상한이 사실상 꺼져요. AC-2 가 상한 직전과 직후를 모두 시험해요(가드 뮤턴트).
- R-2: 시계가 어긋나면 임대 끝이 달라 보여요. 다른 임대 판정(`card reclaim`)과 같은 DB 시계를 써요.
- R-3: 임대가 끝난 `review` 카드가 감사를 기다릴 때 상한에서 빠지면 레인이 새 카드를 더 받아요. `audit_needed` 는 그대로라서 그 카드는 `card next` 의 감사 목록에 계속 보여요.
