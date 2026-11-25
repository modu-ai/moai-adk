---
id: SPEC-20261125-review-lease-hold
title: "A card in review keeps its lease until the verdict is recorded, and a reclaimed reservation brings back its scope"
status: approved
revision: 1
created: 2026-11-25
approval:
  by: user
  at: 2026-11-25
  basis: "사용자가 리더 대화의 선택지에서 직접 고른 답(카드 노트 9093): 임대 방식 \"24시간 창 (Recommended)\", 승인 \"승인, (가)(나) 모두 (Recommended)\"; 리더가 기록 직전에 moai spec show 로 digest 를 다시 읽었어요"
  digest: "sha256:8348f5f7940606534b66a96240881687bdecea679f7efb4b5ad7eb066012897f"
---

# A card in review keeps its lease until the verdict is recorded, and a reclaimed reservation brings back its scope

## Intent

제출한 카드는 독립 감사를 기다리는데, 그동안에도 예약의 임대(lease)는 계속 흘러요(`card take` 때 기본 14400초). 임대가 끝나면 작성한 레인은 카드를 닫을 수 없어요: `release` 는 `review` 카드에 거절되고(`CardInReview`), `reopen` 은 살아 있는 임대가 필요하고, 다른 세션은 `card reclaim` 만 쓸 수 있는데 이것은 평생 3회까지만 세어요. 다른 저장소(mo.ai.kr 리더의 개선 요청 부록 4, moai 4.2648.35)의 실측이 결과를 보여 줘요: review 카드 하나가 임대를 두 번 잃어 generation 3, reclaim 2/3 이 됐고, FAIL 판정이 나와 고치려 하면 한도가 이미 없어요. 이 저장소에서도 2026-11-25 05:36Z 까지 한 카드가 묶여 같은 일이 있었어요. 같은 실측에서 reclaim 할 때마다 범위가 비어 돌아와서 레인이 매번 `card scope` 로 손으로 되살렸어요. 원인은 이 저장소의 코드에 있어요: `card review` 는 `--keep` 으로 이름을 대지 않은 경로를 모두 돌려주고, `card reclaim` 은 옛 예약의 범위(`reservation_scope`)를 그대로 복사하니 복사할 게 없고, `yield_overlaps` 가 끝난 임대의 경로를 더 지울 수도 있어요. 이 SPEC 은 review 카드의 임대가 판정까지 이어지게 하고, 판정이 나오면 레인에게 새 임대를 주고, reclaim 과 reopen 이 카드 자신의 범위를 되살리게 하고, FAIL 뒤 첫 reclaim 은 한도에 세지 않아요.

## Interview

- Q1: review 에서 감사를 기다리는 동안 임대가 끝나 reclaim 횟수를 써요. 어떻게 고칠까요? 선택지: (가) review 카드의 임대가 판정이 기록될 때까지 이어지고, 판정 뒤 레인이 새 임대를 받아요 (나) 리더가 다른 세션의 예약을 갱신해요 (다) 임대가 끝난 뒤의 인수는 한도에 세지 않아요. → A: 아직 묻지 않았어요. 리더의 기본안은 (가)이고, 이 SPEC 은 (가)에 24시간 상한을 둔 안을 추천해요(D-1): 하루 넘게 감사가 안 오는 review 는 리더가 봐야 할 멈춘 카드이고, 임대 확인 지점은 그대로 둘 수 있어요. 말 그대로 멈추는 안(상한 없음)을 원하면 원장에 정지 표시가 필요해요(D-1 선택 2).
- Q2: reclaim 하면 새 예약이 경로 없이 와서 레인이 손으로 써요. reclaim(과 FAIL 뒤 reopen)이 `card take` 처럼 카드 자신의 범위를 되살릴까요? → A: 아직 묻지 않았어요. 기본: 예. 다른 카드의 살아 있는 예약과 공유 파일에 걸리는 경로는 빼고 답에 이름을 적어요(D-3).
- Q3: FAIL 판정 뒤 레인이 reclaim 3회를 이미 다 썼을 수 있어요. FAIL 판정마다 첫 reclaim 을 한도에 세지 않을까요? → A: 아직 묻지 않았어요. 기본: 예, FAIL 판정마다 하나(D-2). 한도에 닿은 카드의 일반적인 출구는 `SPEC-20261125-card-take-records-worktree` 의 몫이고 여기서 정하지 않아요.

## Non-goals

- 다른 이유로 reclaim 3회 한도에 닿은 카드의 일반적인 출구(`SPEC-20261125-card-take-records-worktree` 의 몫).
- 어떤 세션이 다른 세션의 예약을 갱신하거나 인수할 수 있는지(소유 확인은 그대로: 살아 있는 세션의 예약은 가져가지 않아요).
- 작업 중 한도와 그것이 세는 대상(`SPEC-20261125-wip-limit-live-leases`).
- 카드 사이의 경로 겹침 거절 제거(`SPEC-20261125-card-no-file-locks`): 이 SPEC 은 겹침 규칙이 있는 곳에서만 그것을 읽어요.
- `active` 카드의 임대, 끝난 임대의 sweep, 사라진 보유자의 이른 reclaim, `card adopt`.

## Requirements

- REQ-1: When a card moves to `review`, the review command shall extend the current reservation's lease to end no earlier than the ledger clock plus 86400 seconds.
- REQ-2 (REQ-1): When a verdict is recorded on a card in `review`, the verdict command shall extend the current reservation's lease to end no earlier than the ledger clock plus 14400 seconds. · failure: none — a lease that already ends later is left as it is.
- REQ-3 (REQ-1): IF a card stays in `review` for 86400 seconds without a verdict, THEN its lease shall end as every other lease does. · failure: none — the reclaim and the sweep rules apply from then on.
- REQ-4: While a card is in `review` and its lease has not ended, the reclaim command shall refuse with `lease_live` and name that the lease is held for the audit and when it ends. · failure: none — the message only names the reason.
- REQ-5: When a reclaim or a reopen leaves the new or the continued reservation with no paths, the command shall fill the scope from the card's own scope, as a take does. · failure: none — a reservation that has paths is not changed.
- REQ-6 (REQ-5): IF a path of the card's own scope overlaps a path of another card's reservation or a shared file of the profile, THEN the command shall leave that path out, name it in the answer and still succeed.
- REQ-7: When a reclaim follows a FAIL verdict of the card and no `card.reclaim` event of the card is newer than that verdict, the reclaim command shall leave that reclaim out of the three-reclaim allowance. · failure: none — every other reclaim counts as before.
- REQ-8 (REQ-7): The reclaim command shall write `counted` in its event and name in its answer whether the reclaim counted, the reclaims counted so far and the allowance.
- REQ-9: The lease rules shall apply as before this SPEC to a card that is not in `review`, to the sweep, to the scope yield and to the early reclaim. · failure: none — a lane in the active state keeps its four hours.

## Acceptance

- AC-1 (REQ-1, REQ-2): in a temporary ledger, `card review` on a card with a short lease leaves a lease that ends at least 86400 seconds after the review, and the first `card verdict` on it leaves a lease that ends at least 14400 seconds after the verdict; a lease that already ends later stays as it is · verify: `cargo test --locked --offline --test card card_review_lease_hold_cli::ac1_`
- AC-2 (REQ-3, REQ-4): a card in review whose lease is live is refused by `card reclaim` with `lease_live` and a message that names the audit and the end time; after the 86400 seconds (test clock) the reclaim succeeds and counts · verify: `cargo test --locked --offline --test card card_review_lease_hold_cli::ac2_`
- AC-3 (REQ-5, REQ-6): a card reviewed without `--keep`, whose lease then ended, is reclaimed with the card's own scope again; a path that another card's live reservation overlaps is left out and named, and the reclaim still succeeds; `card reopen` on a review card with an empty reservation fills it the same way · verify: `cargo test --locked --offline --test card card_review_lease_hold_cli::ac3_`
- AC-4 (REQ-7, REQ-8): a card with three counted reclaims and a FAIL verdict is reclaimed once more with `counted` false, a second reclaim after the same verdict is refused with `reclaim_limit`, and a card without a FAIL verdict is refused at the fourth reclaim as before; the answer names the reclaims counted and the allowance · verify: `cargo test --locked --offline --test card card_review_lease_hold_cli::ac4_`
- AC-5 (REQ-9): the lease of an `active` card, the sweep of an ended lease and `card adopt` give the same results as before · verify: `cargo test --locked --offline --test card card_reclaim_cli`
  verify: `cargo test --locked --offline --test card card_adopt_cli`
  verify: `cargo test --locked --offline --test card ledger_lease_yield`
  verify: `cargo test --locked --offline --test card ledger_sweep`
  verify: `cargo test --locked --offline --test card ledger_workflow`
  verify: `cargo test --locked --offline --test card ledger_scope_return`

## Scope

- write: `.moai/specs/SPEC-20261125-review-lease-hold/spec.md`
- write: `src/ledger/reservation.rs`, `src/ledger/workflow.rs`, `src/card/work.rs`, `src/card/dispatch.rs`
- write: `tests/card/card_review_lease_hold_cli.rs`, `tests/card/main.rs`, `tests/card/card_reclaim_cli.rs`
- read-only: `src/ledger/sweep.rs`, `src/ledger/adopt.rs`, `src/card/wip_limit.rs`, `src/card/early_reclaim.rs`
- forbidden: the real HOME, the installed moai binary, another project, the ledger schema (no new column or table), a card number inside any file other than the card's own notes

## Plan

1. 이 SPEC 의 승인과 `SPEC-20261125-card-no-file-locks` 의 착지를 기다려요: 그 SPEC 이 카드 사이 경로 겹침 거절을 없애므로 REQ-6 이 빼야 할 것이 바뀌어요. 먼저 착지하면 REQ-6 의 겹침 절은 공유 파일 경우로 줄어요.
2. 시험 먼저: `tests/card/card_review_lease_hold_cli.rs` 에 AC-1 ~ AC-4 의 실패하는 시험을 시험 시계가 있는 임시 원장에 써요.
3. `src/ledger/workflow.rs`: `review` 의 상태 변경(REQ-1)과 판정 기록(REQ-2)과 같은 트랜잭션에서 임대를 늘려요. "임대가 시계 + n초 이상" 도우미는 하나만 둬요.
4. `src/ledger/reservation.rs`: review 카드의 `lease_live` 답 문구(REQ-4)와, FAIL 판정 뒤 reclaim 을 한도 계산에서 빼는 일(REQ-7)을 넣어요. 센 횟수는 계속 이벤트에서 구하고 `counted: false` 페이로드가 유일한 예외예요.
5. `src/card/dispatch.rs`, `src/card/work.rs`: reclaim 과 reopen 뒤 빈 범위를 카드 자신의 범위로 채우고(`take` 의 겹침·공유 파일 규칙, REQ-5·6) reclaim 답에 REQ-8 의 세 필드를 더해요.
6. AC-5 의 이웃 시험을 돌리고 Result 에 기준마다 한 줄을 적고 안 돌린 실행을 모두 밝혀요.

## Risks

- review 카드의 임대가 길어지므로 `card review --keep` 으로 남긴 경로가 최대 하루 예약된 채 있어요. 리더는 판정 없는 review 카드로 보고 감사를 요청할 수 있고, 감사가 끝내 안 오면 REQ-3 의 상한이 끝내요.
- `SPEC-20261125-card-take-records-worktree` 와 `SPEC-20261125-card-no-file-locks` 도 `src/ledger/reservation.rs` 와 `src/card/dispatch.rs` 를 바꿔요. 이 SPEC 은 둘 위로 rebase 해야 하고 reclaim 횟수를 세는 질의는 한 곳에 있어야 해요.
- 작업 중 한도(`SPEC-20261125-wip-limit-live-leases`)는 임대가 살아 있는 카드를 세므로 review 카드가 최대 하루 세어져요. 바꾼 뒤 한도가 몇 개의 review 카드를 쥐는지 측정해야 해요.
- review 중에 레인이 사라지면 갱신도 reopen 도 못 해요. 상한 뒤에는 reclaim 이 유일한 길이고 REQ-7 로는 FAIL 뒤 첫 번째만 공짜예요. 두 번 FAIL 한 카드는 흐름대로 리더에게 가요.
- reclaim 이 범위를 채울 때 레인이 더는 필요 없는 경로를 예약할 수 있어요. 답에 채운 경로를 적고 `card scope` 로 고칠 수 있어요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | review 카드의 임대는 `review` 때 24시간으로, 판정 때 기본 4시간으로 늘려요(REQ-1 ~ REQ-3). 선택지: 1 이 창 방식(추천: 스키마 변경 없고 모든 임대 확인이 한 규칙을 유지) · 2 판정까지 말 그대로 멈춤(원장에 정지 표시와 모든 임대 확인의 변경이 필요) · 3 변경 없음, 리더가 다른 세션의 예약을 갱신. | 리더의 기본안은 판정까지 멈춤이고, 이 SPEC 은 하루 안에 오는 감사에는 같은 효과인 창 방식을 제안해요. 사용자는 아직 고르지 않았어요. | 두 연장과 reclaim 문구(REQ-1, REQ-2, REQ-4)를 지워요. |
| D-2 | FAIL 판정 뒤 첫 reclaim 은 세지 않아요(REQ-7). | FAIL 뒤 고치는 일이 한도에 막히면 안 돼요. 두 번째 FAIL 은 리더에게 가므로 FAIL 하나에 하나만 면제하면 한도는 한정돼요. | 모든 reclaim 을 다시 세요(이벤트 예외를 지워요). |
| D-3 | reclaim 과 reopen 이 빈 범위를 카드 자신의 범위로 채우고 겹치거나 공유 파일인 경로는 빼요(REQ-5, REQ-6). | `card take` 가 이미 빈 범위를 이렇게 채우고, `card review` 는 `--keep` 이 없으면 경로를 모두 돌려줘서 일부러 비워요. | 채움을 지워요. 레인이 예전처럼 손으로 넓혀요. |
| D-4 | 86400초 상한은 기존 최대 임대(`MAX_LEASE_SECS`)와 같아요. | 새 숫자가 없어요: 임대는 이미 하루까지 가능해요. | 상수 하나를 바꿔요. |

## Result

구현 head 는 카드의 READY 노트에 있어요(영수증도 거기에 있어요).

- AC-1 met: `ac1_review_holds_the_lease_for_a_day_and_a_verdict_for_four_hours` 통과(짧은 임대 60초 → review 뒤 86400초 이상, 판정 뒤 14400초 안팎, 답의 `lease_until` 이 장부와 같아요). `ac1_a_lease_that_already_ends_later_is_left_as_it_is` 통과(리뷰와 판정 모두 더 긴 임대를 줄이지 않아요).
- AC-2 met: `ac2_a_review_card_with_a_live_lease_is_refused_and_reclaimed_after_the_day` 통과(`lease_live` 코드, 메시지에 "audit" 와 끝나는 시각, 임대를 끝낸 뒤 reclaim 은 세요).
- AC-3 met: `ac3_a_reclaim_after_a_review_without_keep_brings_back_the_cards_own_scope`, `ac3_a_path_another_card_holds_is_left_out_and_named`, `ac3_a_shared_file_is_left_out_of_the_fill`, `ac3_reopen_fills_an_empty_reservation_and_leaves_a_full_one` 통과. 채운 경로는 `scope_filled`, 뺀 경로와 이유는 `scope_left_out` 으로 답과 `card.reclaim`·`card.reopen` 이벤트에 있어요.
- AC-4 met: `ac4_a_reclaim_after_a_fail_verdict_is_free_once` 통과(세 번 센 카드는 판정 없이 네 번째가 `reclaim_limit`, FAIL 뒤 첫 reclaim 은 `counted` false 이고 `reclaims_counted` 3 / `reclaim_allowance` 3, 같은 판정으로 두 번째는 거절). `ac4_a_pass_verdict_frees_nothing_and_an_ordinary_reclaim_counts` 통과.
- AC-5 met: `card_reclaim_cli`, `card_adopt_cli`, `ledger_lease_yield`, `ledger_sweep`(그리고 `card_reclaim_early_cli`, `ledger_reservation`, `ledger_workflow`, `ledger_scope_return`, `card_verdict_cli`) 가 바꾸기 전과 같은 결과로 통과했어요. 활성 카드의 임대와 sweep 은 코드가 달라지지 않았어요(`hold_in_tx` 는 `review` 상태에서만 움직여요).
- 설계 메모: `card review` 의 트랜잭션은 이 SPEC 의 쓰기 범위 밖인 `src/ledger/scope_return.rs` 에 있어서, 임대 연장은 같은 트랜잭션이 아니라 review 직후 명령 계층(`src/card/work.rs`)의 두 번째 단계로 해요. 실패하면 review 는 그대로 서고 답의 `warnings` 에 적혀요(`ac1_a_failed_hold_after_review_is_a_warning_and_the_review_stands` 가 고정해요). 리더 승인.
- 범위에 더한 파일: `tests/card/card_line_lease_cli.rs` 하나. `review_without_keep_gives_back_the_whole_scope` 가 reopen 뒤 범위가 비어 있음을 박아 두었는데 REQ-5(D-3)가 그것을 바꿔서, 새 규칙(채워진 범위, 한 줄 더 넓히기)에 맞게 고쳤어요. 첫 `check affected --card` 가 이 한 곳에서만 실패했어요(1545/1546).
- 긴장: `card take` 는 겹침 방침을 바꾼 앞선 변경 뒤로 겹치는 경로를 더 거절하지 않지만 REQ-6/AC-3 은 다른 카드의 살아 있는 예약과 겹치는 경로를 채움에서 빼라고 해서, 적힌 대로 구현했어요. 필요하면 REQ-6 을 줄이는 일은 리더 몫이에요.
- not_run: 가드 뮤턴트(tier 1 이라 요구되지 않아요: `bundle check` audit_floor 1). 실제 HOME 의 설치본 장부로는 돌리지 않았어요(임시 HOME 만 썼어요). 구현 전 RED 는 새 시험 10개 중 9개가 설계한 이유로 실패하는 것을 확인했어요(1개는 이미 맞는 회귀 방지).
