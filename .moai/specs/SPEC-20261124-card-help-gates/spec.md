---
id: SPEC-20261124-card-help-gates
title: "Card help and gate tests say exactly what the gates do, and they can fail"
status: draft
revision: 1
created: 2026-11-24
---

# Card help and gate tests say exactly what the gates do, and they can fail

## Intent

카드 C-437 이에요. 감사 #420, #428, #444 가 PASS 로 착지시키면서 남긴 비차단 지적을 한 번에 닫아요: 도움말이 게이트가 실제로 하는 일을 다 말하지 않고, 몇몇 게이트 테스트는 코드를 지워도 통과해요(뮤턴트 M8, m9, N1 이 살아남았어요). 도움말과 문장이 코드와 다르면 코드가 이기므로(core rule 21) 문장을 코드에 맞추고, 지워도 통과하는 테스트는 실패할 수 있게 고쳐요.

읽어서 확인한 사실(코드, 2026-11-24):

- `card note --help` 는 마커를 `` `LANE-READY-INTEGRATION`, `REPAIRED` or `READY ` `` 로 적어요(`src/ops/card.rs` Note). 지금 테스트는 `READY` 낱말만 보는데 그 낱말은 `LANE-READY-INTEGRATION` 안에도 있어서 마커 이름을 바꿔도 통과해요(M8).
- 첫 줄 앞에 빈칸이 있는 READY 줄은 거절되고(`src/card/ready_receipt.rs` 의 `marker_of_line(first.trim_start())`, 테스트 `a_glued_marker_is_one_rule_for_gate_inbox_and_status_line`), 영수증 요약 줄의 `base` 가 head 의 조상이 아니면 영수증이 아니에요(`ready_receipt.rs` 의 `merge-base --is-ancestor`). 도움말은 둘 다 말하지 않아요.
- `takeover_gates`(`src/card/dispatch.rs`)는 `cur.status == "held"` 이고 시계가 읽히고 임대가 끝났을 때만 느슨한 문이고 나머지는 엄격한 문이에요. 그 첫 조건(held)을 지워도 통과하는 테스트가 없어요(m9). 시계 분기는 `SELECT unixepoch()` 가 실패해야만 닿는데 SQLite 의 이 함수는 실패하지 않아서 CLI 에서 만들 수 없어요(D-1).
- `card withdraw` 도움말 첫 줄은 `draft or ready` 만 말하는데 본문과 코드는 `active` 와 `review` 도 받아요(AUDIT #444 N2). `card take` 가 fold 된 카드를 거절하는 문장은 `…and withdrawn: folded into C-2: …` 이에요(`src/ledger/reservation.rs` 의 `CardNotReady` 표시, N3). `open_cards_of`(`src/card/spec_link.rs`)는 `state <> 'done'` 이라 `hold`(withdraw 와 fold 두 길로만 들어가요, `workflow.rs`·`close.rs`)도 열린 카드로 세요(O6).

## Non-goals

- `src/card/dispatch.rs`: C-445 가 쥐어요. 읽기만 하고 고치지 않아요.
- 시계 분기(`unixepoch()` 읽기 실패)의 테스트: CLI 로는 만들 수 없어요(D-1).
- `card reclaim` 의 동작 바꾸기. 도움말과 테스트만 바꿔요.

## Requirements

- REQ-1: The `card note` help shall name the markers `LANE-READY-INTEGRATION`, `REPAIRED` and `READY ` with its trailing space, the refusal of a first line that has a blank before its marker, and the rule that the `base` of the receipt's summary line is an ancestor of the head.
- REQ-2: The `card reclaim` help shall say that unpushed commits do not stop the takeover of a card whose lease has ended and do stop the takeover of a card whose lease is still live.
- REQ-3: The `card withdraw` help first line shall name the `draft`, `ready`, `active` and `review` states.
- REQ-4: IF `card take` refuses a card that was folded, THEN the refusal shall say `folded into C-<n>` in place of `withdrawn`.
- REQ-5: The `open_cards_of` list shall leave out a card in the `hold` state, so `card add --spec` does not count a withdrawn or folded card as an open card of that SPEC.
- REQ-6: IF a reservation is `expired_candidate` or `held_for_review` and its lease end has passed, THEN `card reclaim` shall judge the recorded worktree with the strict gate that refuses unpushed commits.
- REQ-7: The `card next` answer for a project whose profile cannot be read shall carry the same `warnings` as the answer for the same ledger without a stale draft.

## Acceptance

- AC-1 (REQ-1): 정리한 공백으로 읽은 `card note --help` 가 `` `READY ` ``(역따옴표와 끝 공백 포함), 앞 공백 거절, 요약 줄 `base` 의 조상 규칙을 말해요. · verify: `cargo nextest run --offline --locked --test card_ready_receipt_cli -E 'test(the_note_help_names_the_receipt_gate)'`
- AC-2 (REQ-1): `src/ops/card.rs` 의 `` `READY ` `` 마커 이름을 바꾼 버려지는 복사본에서 AC-1 의 명령이 0 이 아닌 코드로 끝나요(뮤턴트 M8 이 이제 죽어요). · verify: `cp -c` 한 복사본에서 `` `READY ` `` 를 `` `RDY ` `` 로 바꾼 뒤 AC-1 의 명령을 돌려 exit 100
- AC-3 (REQ-2): `card reclaim --help` 가 임대가 끝난 뒤에만 안 푼 커밋이 인수를 막지 않는다고 말하고, 그 말이 빠지면 테스트가 실패해요. · verify: `cargo nextest run --offline --locked --test card_reclaim_early_cli -E 'test(the_reclaim_help_says_when_unpushed_commits_stop_the_takeover)'`
- AC-4 (REQ-6): `held_for_review` 와 `expired_candidate` 예약을 임대 끝 뒤로 만들고, 안 푼 커밋이 있는 워크트리를 기록해 두고 `card reclaim` 하면 거절되고 장부는 그대로예요. `cur.status == "held"` 조건을 지우면 이 테스트가 실패해요. · verify: `cargo nextest run --offline --locked --test card_reclaim_early_cli -E 'test(a_reservation_that_is_not_held_goes_through_the_strict_gate)'`
- AC-5 (REQ-3): `card withdraw --help` 첫 줄이 네 상태를 말하고 `card take --help` 에는 그 줄이 없어요. · verify: `cargo nextest run --offline --locked --test card_withdraw_cli -E 'test(the_take_and_withdraw_help_first_lines_stay_apart)'`
- AC-6 (REQ-4): fold 된 카드를 `card take` 하면 거절 문장에 `folded into C-2` 가 있고 `withdrawn` 낱말이 없어요. withdraw 된 카드의 거절 문장은 그대로 `withdrawn: <사유>` 를 말해요. · verify: `cargo nextest run --offline --locked --test card_fold_cli --test card_withdraw_cli -E 'test(take_of_a_folded_card_does_not_say_withdrawn) | test(take_of_a_withdrawn_card_still_says_withdrawn)'`
- AC-7 (REQ-5): 한 SPEC 에 withdraw 된 카드와 fold 된 카드만 있을 때 `card add --spec` 가 그 SPEC 의 새 카드를 받고, 열린 카드가 하나 있으면 이전처럼 거절해요. · verify: `cargo nextest run --offline --locked --test card_add_spec_cli -E 'test(a_hold_card_is_not_an_open_card_of_its_spec)'`
- AC-8 (REQ-7): `tests/card_stale_cli.rs` 의 "경고 없음" 확인이 stale 낱말 하나가 아니라 `warnings` 배열 전체를 같은 fixture 의 stale 초안 없는 판과 비교해서, 프로필을 못 읽는 경고가 어떤 낱말로 와도 실패해요. 프로필 읽기 경로에 다른 낱말의 경고를 더한 복사본에서 이 테스트가 실패해요. · verify: `cargo nextest run --offline --locked --test card_stale_cli -E 'test(an_unreadable_profile_uses_the_default_days_and_adds_no_warning)'`
- AC-9 (REQ-1..REQ-7): 쓰기 범위 밖의 `src` 파일 diff 가 0이고 `src/card/dispatch.rs` 는 그대로예요. · verify: `git diff --stat <base>..<head> -- src`

## Scope

- write: `.moai/specs/SPEC-20261124-card-help-gates/spec.md`, `src/ops/card.rs`(공유 파일: `card line-lease` 로 줄만 빌려요), `src/ledger/reservation.rs`, `src/card/spec_link.rs`, `tests/card_ready_receipt_cli.rs`, `tests/card_reclaim_early_cli.rs`, `tests/card_withdraw_cli.rs`, `tests/card_fold_cli.rs`, `tests/card_add_spec_cli.rs`, `tests/card_stale_cli.rs`
- read-only: `src/card/dispatch.rs`(C-445), `src/card/ready_receipt.rs`
- forbidden: 실제 HOME 과 실제 장부(임시 HOME 만), 네트워크, `CLAUDE.md`, `src/card/dispatch.rs` 쓰기

## Plan

1. RED: 위 AC 의 테스트를 먼저 써서 실패를 봐요(AC-1, 3, 4, 5, 6, 7, 8). AC-4 의 `held` 조건과 AC-2 의 마커 뮤턴트는 코드를 고치기 전에 한 번 돌려서 테스트가 살아남던 때와 다르게 이제 죽는지 봐요.
2. GREEN: `src/ops/card.rs` 의 도움말(note, reclaim, withdraw 첫 줄)을 `card line-lease` 로 고치고, `reservation.rs` 의 거절 문장과 `spec_link.rs` 의 `open_cards_of` 를 고쳐요.
3. 기존 테스트 중 문장이 바뀐 것(`card_withdraw_cli.rs` 의 고정된 첫 줄, `tests/fixtures/help` 골든이 있으면 그것)을 맞춰요.
4. READY: 같은 head 에서 `moai check affected` 를 한 번, 뮤턴트 줄을 새 가드 테스트마다 하나씩 적어요.

## Risks

- `tests/fixtures/help/**` 골든이 `card withdraw` 첫 줄을 담고 있으면 그 파일이 깨져요(공유 파일이라 line-lease 로 고쳐요). 읽어서는 아직 담고 있는 파일이 없다고 봤으나(rg, 2026-11-24) 확인은 RED 에서 해요.
- `open_cards_of` 를 바꾸면 `card add --spec` 가 `hold` 카드가 있는 SPEC 에 두 번째 카드를 받아요. 이것이 O6 의 뜻이고, 열린 카드는 여전히 막아요.
- 표시 문장을 바꾸면 `withdrawn: <사유>` 를 고정한 다른 테스트가 있을 수 있어요(RED 에서 `rg` 로 확인).

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 시계 분기(`SELECT unixepoch()` 실패 → 엄격한 문)는 이 카드에서 테스트하지 않아요. 카드 기준 2 의 그 반쪽은 `src/card/dispatch.rs` 에 넣을 수 있는 자리가 생긴 뒤(C-445 착지 후)의 후속 기준으로 남겨요 | SQLite 의 `unixepoch()` 는 실패하지 않아서 CLI 로 만들 수 없고, 단위 테스트는 `dispatch.rs` 안에 있어야 하는데 C-445 가 쥐어요 | 후속 기준을 닫기 |
| D-2 | 도움말의 마커·조상 문장은 `src/card/ready_receipt.rs` 의 실제 동작에서 가져와요(고치는 것은 문장이지 동작이 아니에요) | core rule 21: 지침과 코드가 다르면 코드가 이겨요 | 문장을 되돌리기 |

## Result

브랜치 `spec/card-help-gates-20261124`(기준 NeXT `b18c7639a9`)의 커밋 SPEC `52c144a738` → RED `235e44b87d` → GREEN `358f699e43` → stale 단언 보강 `ef16be102b` 순서예요. 승인은 리더 노트 8131(digest `sha256:11662e1c…`, `spec show` 로 같은 값을 읽음)이고 D-1 은 (a)예요. RED(`target/card-help-gates/red.log`)는 새 테스트 8개 중 5개 실패, 3개 통과(통과한 셋은 현재 코드가 이미 맞는 쪽의 가드라 뮤턴트로 확인해요). GREEN 은 관련 테스트 바이너리 10개 216개 중 215개 통과(`green.log`), 남은 하나는 제 테스트 문장 오류(`SPEC-2026…` 안의 `C-2`)였고 고친 뒤 `card_add_spec_cli` 11/11 통과예요.

- AC-1: 충족. `the_note_help_names_the_receipt_gate` 가 공백을 접은 도움말에서 `` `READY ` ``, 앞 공백 거절, 조상 규칙을 읽어요.
- AC-2: 충족. `src/ops/card.rs` 의 `` `READY ` `` 를 `` `RDY ` `` 로 바꾼 버려지는 복사본에서 그 테스트가 실패해요(`mutA.log`, exit 100).
- AC-3: 충족. `the_reclaim_help_says_when_unpushed_commits_stop_the_takeover`; 문장에서 `Once the lease has ended,` 를 지운 복사본에서 실패(exit 100).
- AC-4: 충족. `a_reservation_that_is_not_held_goes_through_the_strict_gate`(`held_for_review` 와 `expired_candidate` 두 상태); `cur.status == "held"` 를 지운 복사본에서 실패(exit 100).
- AC-5: 충족. `the_take_and_withdraw_help_first_lines_stay_apart`; 첫 줄을 되돌린 복사본에서 실패(exit 100).
- AC-6: 충족. `take_of_a_folded_card_does_not_say_withdrawn`(표시 문장을 `withdrawn: {reason}` 로 되돌린 복사본에서 실패)와 `take_of_a_withdrawn_card_still_says_withdrawn`(`withdrawn: ` 접두를 지운 복사본에서 실패, `mutB.log`), 둘 다 exit 100.
- AC-7: 충족. `a_hold_card_is_not_an_open_card_of_its_spec`; `NOT IN ('done', 'hold')` 를 `<> 'done'` 으로 되돌린 복사본에서 실패(exit 100).
- AC-8: 충족. `an_unreadable_profile_uses_the_default_days_and_adds_no_warning` 가 stale 초안 없는 판과 경고 배열 전체를 비교하고 빈 배열도 요구해요. 나이와 상관없이 붙는 경고는 두 판이 똑같이 달기 때문에 비교만으로는 잡히지 않아서 빈 배열 단언을 더했어요(SPEC 문장보다 강한 쪽). `src/card/work.rs` 에 다른 낱말의 경고를 더한 복사본에서 실패(exit 100).
- AC-9: 충족. `git diff --stat b18c7639a9..HEAD -- src` → `src/card/spec_link.rs`, `src/ledger/reservation.rs`, `src/ops/card.rs` 세 파일, `src/card/dispatch.rs` 는 0줄이에요. `src/ops/card.rs` 는 `card line-lease` 로 패치를 적용했어요.
- 남은 것: 카드의 기준 2 는 시계 분기 반쪽이 남아 `met` 로 적지 않아요(D-1 (a), C-445 착지 뒤 `src/card/dispatch.rs` 안 단위 테스트).
