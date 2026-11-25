---
id: SPEC-20261125-card-no-file-locks
title: "카드가 쥔 파일 경로의 쓰기 잠금을 없애요: 한 카드 한 작성자는 그대로, 겹침은 제출과 착지가 찾아요"
status: draft
revision: 1
created: 2026-11-25
---

# 카드가 쥔 파일 경로의 쓰기 잠금을 없애요: 한 카드 한 작성자는 그대로, 겹침은 제출과 착지가 찾아요

## Interview

- Q1: 무엇이 레인을 멈추나요? -> A: 다른 카드가 쥔 경로가 겹치는 카드는 `card take` 도 `card scope` 넓히기도 `reservation conflict in write scope` 로 거절되고, `card next` 는 그 카드를 `scope_overlap` 으로 건너뛰어요. 카드 노트 8963(테스트 파일 한 고정), 8967(제출에 필요한 네 경로를 이미 커밋), 8970(`src/update.rs` 세 줄), 8984(삭제 목록 다섯 경로)가 이 거절에 막힌 실제 사례예요. 그때마다 레인은 `LANE-BLOCKED kind=decision` 을 쓰고 다른 카드의 착지를 기다려요.
- Q2: 사용자가 정한 것은요? -> A: 2026-11-25T05:00:41Z, 선택 그대로 "지금 1순위 카드로 (Recommended)": 파일 단위 쓰기 잠금을 없애고, 한 카드 한 작성자는 그대로 두며, 파일 겹침은 제출(READY)과 착지 머지가 찾아요(앞선 결정: 잠금 제거, 제출 때 검사). 카드 소유(예약, generation)와 HOLD, 공유 파일의 line-lease 를 어떻게 둘지는 이 SPEC 이 선택지로 내요.
- Q3: 지금 잠금은 어디에 있나요? -> A: 한 함수예요. `src/ledger/reservation.rs` 의 `find_overlap` 이 `reserve`(take), `widen_in_tx`(card scope), `src/ledger/workflow.rs` 의 `card next` 건너뛰기에서 같이 불려요. 만료된 임대는 `yield_overlaps` 가 새 take·넓히기에게 겹친 경로를 내줘요(`card.yield` 이벤트). `card check` 의 `not under what C-n reserved` 는 겹침이 아니라 카드 자신이 예약한 범위의 검사라 잠금이 아니에요.
- Q4: 겹침은 이미 어디서 잡히나요? -> A: 착지예요. tip 이 움직이면 `card land` 가 `stale_base` 로 거절하고 착지 담당이 머지를 한 번 해서 충돌을 풀어요(NeXT-Flow). 제출 쪽에는 같은 파일이 tip 에서 바뀌었다는 안내가 아직 없어요.
- Q5: tier 2 보호 경로의 거절은 지금 무엇인가요? -> A: take·넓히기에서 막는 규칙은 없어요. tier 2 바닥은 착지 때(`src/card/land/check.rs` 의 `floor_tier`, `src/gate.rs` 의 `tier_of`) 판정을 올려서 이 변경을 쓰지 않은 세션의 전체 감사를 요구해요. 이 SPEC 은 그 바닥을 건드리지 않아요.

## Intent

두 카드가 같은 파일을 만지는 일은 흐름에서 보통이에요. 지금은 먼저 잡은 카드가 끝날 때까지 뒤의 카드가 멈추는 방식으로 막지만, 실제로는 끝나기를 기다리는 시간이 병렬의 이익을 다 먹어요. NeXT-Flow 는 이미 "변경은 착지 때 머지 한 번"으로 겹침을 풀어요. 이 SPEC 은 take·넓히기·건너뛰기의 겹침 거절을 없애고, 겹침을 막는 대신 알려요: take·넓히기·`card check` 는 겹친 경로와 카드를 답에 보여 주고, 제출 때는 base 이후 tip 에서 같은 파일이 바뀌었는지 목록을 보여 줘요. 한 카드에는 한 작성자만 있다는 예약과 generation 울타리, HOLD, 선행 조건, tier 2 착지 바닥은 그대로예요. 읽는 사람: 사용자(승인), 레인, 감사자, 착지 담당.

## Non-goals

- 카드 소유(예약, generation, 임대)와 HOLD, 선행 조건, 기준 없는 카드와 빈 범위의 거절. 모두 그대로예요(REQ-6).
- tier 2 착지 바닥과 독립 감사 규칙, `card land` 의 `stale_base` 와 머지 절차. 그대로예요(REQ-7).
- 공유 파일 규칙(프로필의 `shared_files` 를 통째로 예약하지 않는 규칙과 `card line-lease`)의 제거. 겹침 거절이 없으면 필요가 줄지만 이번 조각은 그대로 둬요(D-3). 후속에서 지워요.
- 문서와 스킬 글(`CLAUDE.md`, `docs/lane-rules.md`, `skills/moai/**`)의 고침. 잠금이 사라졌다는 설명은 그 파일을 쥔 카드가 해요.
- 겹침을 자동으로 풀거나 머지하는 일. 알릴 뿐 풀지 않아요.
- 카드 쓰기 범위 자체의 폐지. 범위는 그 카드가 쓸 곳을 적는 선언으로 남고 `card check` 가 범위 밖 경로를 거절하는 것도 그대로예요(D-2).

## Requirements

<!-- This SPEC removes a check that blocks (the overlap refusal) and adds none: it adds information to answers. Every refusal that is kept is listed in REQ-6 and REQ-7. -->
- REQ-1: When `card take` reserves a ready card whose write scope can be written under a pattern of another card's active reservation, the card ledger shall make the reservation and answer each overlap as `overlaps` with the other card, its holder, its pattern and the pattern of this card. · failure: none — an empty scope, an invalid pattern, a card that is not ready or already held still refuse as before (REQ-6).
- REQ-2: When `card scope` widens a held card by a pattern that another card's active reservation can also write, the card ledger shall widen the reservation and answer each overlap as `overlaps`. · failure: none — a narrowing, a pattern added twice, a case-only difference or a holder that is not the caller still refuse as before.
- REQ-3: When `card next` lists the cards that can be taken, it shall list a card whose write scope overlaps another card's reservation as ready instead of skipping it with the reason `scope_overlap`. · failure: none — the other skip reasons stay (held, empty scope, review with no holder, no criterion, HOLD, prerequisite).
- REQ-4: When `card check` answers `allowed: true` for a path that another card's active reservation can also write, the card ledger shall answer `overlaps` with that card and its pattern for the path. · failure: none — a path outside the reservation of the checked card is still refused.
- REQ-5: When `card note` keeps a hand-back note, the card ledger shall answer `tip_overlap` as the paths that both the range from the note's base to its head and the range from that base to the integration tip changed, and `[]` when none. · failure: none — when git cannot read a range the answer says `tip_overlap: "not_checked"` with the reason and the note is kept.
- REQ-6: IF a card is not ready, is held, is on HOLD, has an unfinished prerequisite, has no criterion that counts or has an empty write scope, THEN `card take` and `card next` shall refuse or skip it as before. · failure: none — the refusal is the safe behavior.
- REQ-7: When a head is landed, `card land` shall judge it exactly as before, the tier 2 floor, the independent verdict, the receipt and the `stale_base` refusal included. · failure: none — this SPEC changes no landing check.
- REQ-8: When a take or a widening overlaps another card's reservation, the card ledger shall leave every other card's reservation, its scope and its lease unchanged without writing a `card.yield` event. · failure: none — nothing is taken from the other card.
- REQ-9: IF a call names a generation that is not the current one, a holder that is not the reservation's holder or a lease that has ended, THEN the card ledger shall refuse it as before. · failure: none — the fence is the safe behavior.

## Acceptance

- AC-1 (REQ-1, REQ-8): 한 카드가 `src/a.rs` 를 예약한 원장에서 다른 카드를 `src/**` 로 take 하면 예약이 만들어지고 답의 `overlaps` 가 첫 카드·그 holder·두 패턴을 담으며, 첫 카드의 범위·임대와 이벤트에 `card.yield` 가 늘지 않아요; 임대가 끝난 첫 예약이어도 같아요 · verify: `cargo nextest run --test card -E 'test(/card_no_file_locks_cli::ac1_/)'`
- AC-2 (REQ-2, REQ-8): 보유 카드를 다른 카드가 쥔 경로로 `card scope` 넓히면 넓혀지고 `overlaps` 가 나오며 다른 카드의 범위는 그대로예요 · verify: `cargo nextest run --test card -E 'test(/card_no_file_locks_cli::ac2_/)'`
- AC-3 (REQ-3): 두 카드의 범위가 겹쳐도 `card next` 가 둘 다 `ready` 로 올리고 `skipped` 에 `scope_overlap` 이 없어요; HOLD 카드와 선행 조건 카드는 지금처럼 건너뛰어요 · verify: `cargo nextest run --test card -E 'test(/card_no_file_locks_cli::ac3_/)'`
- AC-4 (REQ-4): 다른 카드도 쥔 경로의 `card check` 가 `allowed: true` 이고 `overlaps` 에 그 카드와 패턴이 있어요; 겹치지 않는 경로는 `overlaps` 가 비어 있어요; 예약 밖 경로는 지금처럼 거절돼요 · verify: `cargo nextest run --test card -E 'test(/card_no_file_locks_cli::ac4_/)'`
- AC-5 (REQ-5): 임시 저장소에서 base 이후 tip 이 `src/a.rs` 를 바꾸고 lane head 도 `src/a.rs` 를 바꾸면 READY 노트의 답이 `tip_overlap: ["src/a.rs"]` 이고, 겹침이 없으면 `[]` 이며, base 를 읽을 수 없으면 `not_checked` 와 이유를 내고 노트는 남아요 · verify: `cargo nextest run --test card -E 'test(/card_no_file_locks_cli::ac5_/)'`
- AC-6 (REQ-6, REQ-9): HOLD 카드, 선행 조건 카드, 이미 보유된 카드, 기준 없는 카드, 빈 범위 카드의 take 가 지금과 같은 코드로 거절되고, 낡은 generation 의 `card check`·`card scope` 가 거절돼요; 보호 절의 가드 뮤턴트(HOLD 검사 한 줄 삭제)가 시험 실패로 잡혀요 · verify: `cargo nextest run --test card -E 'test(/card_no_file_locks_cli::ac6_/)'` and `cargo nextest run --test card -E 'test(/ledger_reservation::/)'`
- AC-7 (REQ-7): `card land` 가 tier 2 바닥과 `stale_base` 를 지금처럼 판정해요 · verify: `cargo nextest run --test card -E 'test(/card_land_cli::/)'` and `cargo nextest run --test card -E 'test(/card_land_batch_cli::/)'`
- AC-8 (REQ-1, REQ-2, REQ-3): 겹침 거절을 고정하던 옛 테스트(`ledger_reservation`, `ledger_workflow`, `ledger_lease_yield`, `card_take_scope_cli`, `card_work_cli`, `card_generated_scope`, `plan_admission`, `boundary_repro`)가 새 동작(겹침 허용과 `overlaps` 보고)을 고정하도록 바뀌고 모두 통과해요 · verify: `cargo nextest run --test card -E 'test(/ledger_reservation::|ledger_workflow::|ledger_lease_yield::|card_take_scope_cli::|card_work_cli::|card_generated_scope::/)'`

## Scope

- write: `.moai/specs/SPEC-20261125-card-no-file-locks/**`, `src/ledger/reservation.rs`, `src/ledger/workflow.rs`, `src/ledger/scope_return.rs`, `src/card/work.rs`, `src/card/landed.rs`, `src/card/ready_receipt.rs`, `tests/card/card_no_file_locks_cli.rs`(새 파일), `tests/card/ledger_reservation.rs`, `tests/card/ledger_workflow.rs`, `tests/card/ledger_lease_yield.rs`, `tests/card/card_take_scope_cli.rs`, `tests/card/card_work_cli.rs`, `tests/card/card_generated_scope.rs`, `tests/plan_admission.rs`, `tests/boundary_repro.rs`
- shared (line lease, 한두 줄): `src/ops/card.rs`(도움말 문구에서 `scope_overlap` 설명), `tests/card/main.rs`(`mod` 한 줄)
- read-only: `src/scope.rs`, `src/card/land/check.rs`, `src/card/take_scope.rs`
- forbidden: 착지 판정(`src/card/land/**`)의 변경, tier 2 바닥 목록, `CLAUDE.md`, `skills/**`, `docs/lane-rules.md`, 이 저장소의 `.moai/project.yaml`, 실제 HOME

## Plan

1. SPEC 초안을 커밋하고 digest 를 `LANE-BLOCKED kind=decision` 으로 올려요. 레인은 승인을 기록하지 않아요. 승인 전에는 구현 코드를 쓰지 않아요(RED 테스트까지만 써 둬요).
2. RED: `tests/card/card_no_file_locks_cli.rs` 에 AC-1~AC-6 을 써요(임시 HOME 의 임시 원장, 진짜 `moai card …` 호출). 지금 코드는 겹침을 거절하므로 AC-1~AC-5 가 올바른 이유로 실패해요.
3. GREEN: `reservation.rs` 의 `reserve_in_tx`·`widen_in_tx` 에서 `find_overlap` 거절과 `yield_overlaps` 호출을 걷어 내고 겹침 목록을 만들어 답에 실어요(`find_overlap` 은 목록을 내는 함수로 바꿔 `card check`·`card next` 가 같이 써요). `workflow.rs` 의 `card next` 에서 건너뛰기 갈래를 지우고 `SkipReason::ScopeOverlap` 와 그 사용처(`landed.rs`, `work.rs`, `scope_return.rs`)를 정리해요. `ready_receipt.rs` 의 READY 관문이 base·head·통합 tip 의 `git diff --name-only` 교집합을 `tip_overlap` 으로 답해요.
4. 옛 테스트를 새 동작에 맞춰 고쳐요(AC-8). 도움말 한두 줄은 line-lease 로 써요.
5. 확인: 새 테스트와 이웃 테스트, `cargo fmt --all`, 새 가드 테스트마다 가드 뮤턴트 한 줄(셸 `sed` 로 정확히 한 번 치환하고 `git checkout --` 로 되돌려요, Python 없음), `moai check affected --card` 를 READY 에서 한 번. 위험 경로 실행은 임시 HOME 의 원장에서 겹치는 두 카드를 take 하고 착지까지 보는 것이 해당해요.
6. SYNC: Result 에 AC 별 met·not_run 을 적어요. tier 2(게이트가 바뀌어요)이므로 이 변경을 쓰지 않은 세션이 전체 감사로 맡아요.

## Risks

- 겹침이 착지 때 충돌로 늦게 드러나요: 지금은 take 때 멈추고, 이후에는 제출·착지 때 알게 돼요. 완화: 제출 답의 `tip_overlap` 과 take·넓히기·`card check` 의 `overlaps` 가 일찍 알려 주고, 착지의 `stale_base` 머지는 그대로예요. 그래도 같은 줄을 둘이 고치면 착지 담당이 풀어야 해요.
- 같은 파일을 두 레인이 동시에 고쳐 의미가 어긋나는 일(충돌 없이 합쳐진 뒤의 회귀)은 이 SPEC 이 막지 않아요. 지금도 잠금이 막은 것은 같은 경로의 동시 쓰기이지 의미 충돌이 아니에요. 착지의 넓은 검사와 감사가 보완해요.
- `card.yield` 이벤트를 쓰던 읽는 쪽(`card show` 의 타임라인, 정리 도구)이 더는 새 이벤트를 받지 못해요: 읽는 쪽은 과거 이벤트를 계속 읽으니 깨지지 않지만 AC-8 의 옛 테스트가 이 의존을 드러내요.
- 공유 파일 규칙(`shared_files`, `line-lease`)이 필요 없어졌는데 남아요(D-3): 레인은 계속 `line-lease` 를 쓸 수 있고 쓰지 않아도 돼요. 후속 카드가 지워요.

## Decisions

승인 전 제안이에요. 사용자나 위임받은 리더가 고르면 revision 을 올려 `approval` 을 기록해요(레인은 승인을 기록하지 않아요). 기본값은 권장안이에요.

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | take·넓히기·건너뛰기의 겹침 거절을 모두 없애고 `overlaps` 로 알려요(기본). 대안: tier 2 바닥 경로만 한 카드가 독점하게 남겨 두고 나머지만 푼다 | 사용자 결정이 "파일 단위 쓰기 잠금을 없앤다"예요. tier 2 는 이미 착지 때 이 변경을 쓰지 않은 세션의 전체 감사가 막아요. 독점 규칙을 남기면 바닥 목록을 take 쪽이 읽어야 하고(새 의존), 사례 넷이 모두 일반 경로였어요 | 겹침 확인 함수를 거절하는 갈래에 다시 연결해요 |
| D-2 | `card check` 는 카드 자신이 예약한 범위 밖 경로를 계속 거절해요(기본). 대안: 범위 개념을 없애고 아무 경로나 허용 | 범위는 그 카드가 쓸 곳을 적은 선언이고 감사자가 범위 이탈을 읽는 근거예요. 막혔던 사례도 넓히기가 풀려서 해결돼요 | 검사 갈래를 지워요(범위 폐지는 별도 SPEC) |
| D-3 | 공유 파일 규칙(`shared_files`, `card line-lease`)을 이번에는 그대로 둬요(기본). 대안: 같은 변경에서 함께 지워요 | 겹침 거절이 없으면 line-lease 의 필요가 사라지지만 지우려면 프로필 스키마, 도움말, 테스트, 문서가 같이 바뀌어요. 한 카드 한 관심사로 나눠요 | 후속 카드 |
| D-4 | 제출 때 겹침 목록은 hand-back 노트가 남겨질 때 답에 `tip_overlap` 으로 내요(기본). 대안: 같은 목록을 원장 이벤트로도 남겨 `card show` 에 보인다 | READY 관문이 이미 base·head 를 git 으로 읽어요(추가 의존 없음). 이벤트는 새 종류와 읽는 쪽이 늘어요. 알림이 목적이라 답으로 충분해요 | 이벤트 쓰기 한 줄을 더해요 |
| D-5 | 만료된 임대의 경로를 새 take 에게 내주던 `yield_overlaps` 는 호출을 없애요(기본). 대안: 만료 예약의 범위를 계속 정리한다 | 겹침이 허용되면 내줄 이유가 없고, 그 함수는 다른 카드의 예약 범위를 줄이는 쓰기라 REQ-8(다른 카드는 그대로)과 맞지 않아요. 만료 예약은 `card reclaim` 이 정리해요 | 호출을 되살려요 |

## Result

- AC-1 met: `card_no_file_locks_cli::ac1_a_take_over_another_cards_paths_is_made_and_reports_the_overlap` 와 `ac1_an_ended_lease_keeps_its_paths_when_another_card_takes_over_them` (take 가 만들어지고 `overlaps` 가 첫 카드·holder·두 패턴을 담으며, 첫 카드의 범위·임대와 `card.yield` 이벤트 수가 그대로예요; 임대가 끝난 첫 예약도 범위를 지켜요).
- AC-2 met: `ac2_a_widening_over_another_cards_paths_is_made_and_reports_the_overlap` (넓히기가 되고 `overlaps` 가 나오며 다른 카드 범위는 그대로예요).
- AC-3 met: `ac3_card_next_no_longer_skips_a_card_for_an_overlap` (겹치는 두 카드가 모두 `ready`, `scope_overlap` 없음, HOLD 카드는 `hold` 로 건너뛰어요).
- AC-4 met: `ac4_card_check_names_the_other_card_that_holds_the_path` (겹치는 경로는 `allowed: true` 와 `overlaps`, 겹치지 않는 경로는 빈 목록, 예약 밖 경로는 그대로 거절).
- AC-5 met: `ac5_a_ready_note_lists_the_files_the_tip_changed_under_the_lane`, `ac5_a_ready_note_with_no_shared_file_answers_an_empty_list`, `ac5_a_ready_note_keeps_the_note_when_the_integration_branch_is_unknown` (임시 git 저장소: `["src/a.rs"]`, `[]`, 통합 브랜치를 몰라서 `not_checked` 와 이유를 내고 노트는 남아요). 겹침 목록은 영수증 base 가 아니라 통합 tip 과 head 의 merge-base 에서 구해요(`landed::tip_overlap`).
- AC-6 met: `ac6_the_card_owner_hold_scope_and_generation_checks_stay` (이미 보유된 카드, HOLD, 빈 범위, 낡은 generation 의 `card check`·`card scope` 가 그대로 거절돼요)와 이웃 `ledger_reservation::` 테스트. 가드 뮤턴트: 이미 보유 검사 끔 → ac6 실패, HOLD 관문 끔 → ac6 실패. 빈 범위·낡은 generation 거절은 이 변경이 건드리지 않아 뮤턴트를 두지 않았어요.
- AC-7 met: `card_land_cli::`, `card_land_batch_cli::` 를 포함한 `--test card` 전체 1400개가 통과했어요(착지 코드는 바꾸지 않았어요).
- AC-8 met: 옛 테스트 `ledger_reservation`, `ledger_workflow`, `ledger_lease_yield`(통째로 다시 씀), `card_take_scope_cli`, `card_work_cli`, `card_generated_scope`, `card_dispatch_cli`, `card_sweep_cli`, `card_scope_widen_cli`, `card_line_lease_cli`, `ledger_scope_return` 이 새 동작을 고정해요. `plan_admission` 과 `boundary_repro` 는 겹침 판정 함수만 시험해서 바꾸지 않았어요.
- 범위 목록에 없던 파일: `src/card/dispatch.rs`(take·scope 답을 만들어요), `src/card/landed.rs`(`tip_overlap` 을 둬요; `src/card/ready_receipt.rs` 는 다른 카드가 쥐고 있어서 쓰지 않았어요), 위 테스트 파일 넷(`card_dispatch_cli`, `card_sweep_cli`, `card_scope_widen_cli`, `card_line_lease_cli`, `ledger_scope_return`).
- 알려진 틈: `card line-lease` 는 기다리지 않고 바로 실행해서(옵션 `--wait-secs` 는 받지만 쓰지 않아요) 두 카드가 같은 공유 파일에 동시에 줄을 붙일 수 있어요(D-3: 후속 카드가 공유 파일 규칙을 지워요). 문서(`docs/design-review/appendix/ledger_reservation.md`, `operations/card.md`)와 스킬 글의 "겹침 거절" 설명은 고치지 않았어요.
