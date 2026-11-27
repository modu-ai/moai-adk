---
id: SPEC-20261124-card-help-gates
title: "Card help and gate tests say exactly what the gates do, and they can fail"
status: approved
revision: 2
created: 2026-11-24
approval:
  by: user
  at: 2026-11-27
  basis: "리더 노트 10807 (USER-APPROVAL, 2026-11-27, 리더 세션의 질문 도구): 사용자가 직접 골랐어요 — \"둘 다 거절 (Recommended)\"(D-3 A), D-4 a(`card spec` 에 관계 옵션을 더하지 않아요), \"함수로 빼서 테스트 (Recommended)\"(D-5 a). 후속·분리 관계가 있는 카드, SPEC 없는 카드, 같은 SPEC 으로의 다시 연결, `--unlink`, 이미 한 SPEC 에 이어진 열린 카드 둘 이상은 거절하지 않아요."
  digest: "sha256:0080e7f151e802b96cd076da8d7d0c48faae8684a784c37a214f42e07028ec09"
---

# Card help and gate tests say exactly what the gates do, and they can fail

## Intent

카드 C-437 이에요. 감사 #420, #428, #444 가 PASS 로 착지시키면서 남긴 비차단 지적을 한 번에 닫아요: 도움말이 게이트가 실제로 하는 일을 다 말하지 않고, 몇몇 게이트 테스트는 코드를 지워도 통과해요(뮤턴트 M8, m9, N1 이 살아남았어요). 도움말과 문장이 코드와 다르면 코드가 이기므로(core rule 21) 문장을 코드에 맞추고, 지워도 통과하는 테스트는 실패할 수 있게 고쳐요.

읽어서 확인한 사실(코드, 2026-11-24):

- `card note --help` 는 마커를 `` `LANE-READY-INTEGRATION`, `REPAIRED` or `READY ` `` 로 적어요(`src/ops/card.rs` Note). 지금 테스트는 `READY` 낱말만 보는데 그 낱말은 `LANE-READY-INTEGRATION` 안에도 있어서 마커 이름을 바꿔도 통과해요(M8).
- 첫 줄 앞에 빈칸이 있는 READY 줄은 거절되고(`src/card/ready_receipt.rs` 의 `marker_of_line(first.trim_start())`, 테스트 `a_glued_marker_is_one_rule_for_gate_inbox_and_status_line`), 영수증 요약 줄의 `base` 가 head 의 조상이 아니면 영수증이 아니에요(`ready_receipt.rs` 의 `merge-base --is-ancestor`). 도움말은 둘 다 말하지 않아요.
- `takeover_gates`(`src/card/dispatch.rs`)는 `cur.status == "held"` 이고 시계가 읽히고 임대가 끝났을 때만 느슨한 문이고 나머지는 엄격한 문이에요. 그 첫 조건(held)을 지워도 통과하는 테스트가 없어요(m9). 시계 분기는 `SELECT unixepoch()` 가 실패해야만 닿는데 SQLite 의 이 함수는 실패하지 않아서 CLI 에서 만들 수 없어요(D-1).
- `card withdraw` 도움말 첫 줄은 `draft or ready` 만 말하는데 본문과 코드는 `active` 와 `review` 도 받아요(AUDIT #444 N2). `card take` 가 fold 된 카드를 거절하는 문장은 `…and withdrawn: folded into C-2: …` 이에요(`src/ledger/reservation.rs` 의 `CardNotReady` 표시, N3). `open_cards_of`(`src/card/spec_link.rs`)는 `state <> 'done'` 이라 `hold`(withdraw 와 fold 두 길로만 들어가요, `workflow.rs`·`close.rs`)도 열린 카드로 세요(O6).

조각 2(개정 2, 2026-11-27)가 남은 카드 기준 셋을 맡아요. 조각 1(REQ-1..REQ-7)은 착지했고 그 결과는 아래 Result 에 그대로 있어요. 읽어서 확인한 사실(코드, 2026-11-27, NeXT `df710fbe20`):

- 기준 8(감사 #437 N2): `card ready`(`src/card/dispatch.rs` 의 `ready`)는 `draft` 와 `hold` 를 받고, `hold` 카드를 되살릴 때 그 카드가 이름 부른 SPEC 에 다른 열린 카드가 있는지 보지 않아요. `card spec`(`src/card/spec_link.rs` 의 `link`)도 열린 카드를 SPEC 에 연결할 때 보지 않아요. 둘 다 `card add --spec` 가 쓰는 `spec_link::add_conflict` 를 안 불러요. 그래서 한 SPEC 에 열린 카드가 둘이 될 수 있어요(`card add` 의 "한 SPEC 에 열린 카드는 하나" 규칙에 틈이 있어요).
- 기준 2 의 시계 반쪽: `takeover_gates`(`src/card/dispatch.rs` 1006행 근처)는 `cur.status == "held"` 이고 `SELECT unixepoch()` 가 읽히고 임대가 끝났을 때만 느슨한 문이에요. 시계가 읽히지 않는 가지는 CLI 로 만들 수 없어서 지금 테스트가 없어요. C-445 는 `done` 이라 `dispatch.rs` 는 이제 쥐는 카드가 없어요.
- 기준 9(감사 #437 N1): 조각 1 의 READY 뮤턴트 여덟 줄 가운데 일곱은 한 복사본에 같이 적용해 한 번 돌린 것이라 줄마다 따로 증거가 아니었어요.
- 지금 장부(2026-11-27, 읽기 전용 `sqlite3`)에서 열린 카드 중 SPEC 에 연결된 것은 47개, 열린 카드가 둘 이상인 SPEC 은 0개, 되살리면 거절될 `hold` 카드는 0개(`hold` 121개)예요. 곧 쓸 규칙은 지금 데이터를 하나도 거절하지 않아요.

## Interview

- Q1: 무엇이 필요한가요? → A: 카드 C-437 의 남은 기준(리더 지시, 2026-11-27): 기준 2 의 시계 반쪽, 기준 8, 기준 9. 리더는 기준 8 이 게이트 동작을 바꾸므로 새로 거절되는 것과 대가를 선택지로 적어 달라고 했어요.
- Q2: 아래 Decisions 의 "승인 때 골라요" 항목(D-3, D-4, D-5)을 어떻게 고를까요? → A: 사용자가 직접 골랐어요(2026-11-27, 리더 세션의 질문 도구, 원장 노트 10807): D-3 A "둘 다 거절 (Recommended)", D-4 a (`card spec` 에 관계 옵션을 더하지 않아요), D-5 a "함수로 빼서 테스트 (Recommended)". 세 가지 모두 초안의 기본값과 같아요.

## Non-goals

- 조각 1 의 `src/card/dispatch.rs`: 조각 1 은 이 파일을 안 고쳤어요(AC-9). 조각 2 는 `ready` 한 함수와 `takeover_gates` 의 판정 한 곳만 고쳐요(REQ-8, REQ-11).
- 시계 읽기 실패를 CLI 로 만들기: 만들 수 없어요(D-1). 조각 2 는 판정 함수를 따로 빼서 단위 테스트로 세 가지를 모두 고정해요(D-5).
- `card reclaim` 의 동작 바꾸기. 도움말과 테스트만 바꿔요(REQ-11 도 판정을 함수로 옮길 뿐 결과는 같아요).
- 이미 한 SPEC 에 열린 카드가 둘인 기존 데이터를 고치기: 규칙은 앞으로의 연결과 되살림만 막아요.

## Requirements

- REQ-1: The `card note` help shall name the markers `LANE-READY-INTEGRATION`, `REPAIRED` and `READY ` with its trailing space, the refusal of a first line that has a blank before its marker, and the rule that the `base` of the receipt's summary line is an ancestor of the head.
- REQ-2: The `card reclaim` help shall say that unpushed commits do not stop the takeover of a card whose lease has ended and do stop the takeover of a card whose lease is still live.
- REQ-3: The `card withdraw` help first line shall name the `draft`, `ready`, `active` and `review` states.
- REQ-4: IF `card take` refuses a card that was folded, THEN the refusal shall say `folded into C-<n>` in place of `withdrawn`.
- REQ-5: The `open_cards_of` list shall leave out a card in the `hold` state, so `card add --spec` does not count a withdrawn or folded card as an open card of that SPEC.
- REQ-6: IF a reservation is `expired_candidate` or `held_for_review` and its lease end has passed, THEN `card reclaim` shall judge the recorded worktree with the strict gate that refuses unpushed commits.
- REQ-7: The `card next` answer for a project whose profile cannot be read shall carry the same `warnings` as the answer for the same ledger without a stale draft.
- REQ-8: IF `card ready` is asked for a card in `hold` whose newest SPEC link names a SPEC and carries no relation, and that SPEC already has another open card (a card that is neither `done` nor `hold`), THEN `card ready` shall refuse with a `conflict in spec` text that names the open cards and the ways out and write nothing. A `hold` card whose link relation is `follow-up`, `follow-up-of` or `split-of`, a `hold` card that names no SPEC, and a `draft` card are never refused by this rule.
- REQ-9: IF `card spec` links a card that is neither `done` nor `hold` to a SPEC that already has another open card, THEN `card spec` shall refuse with the same `conflict in spec` text and write nothing. `card spec --unlink`, the link of a `done` or `hold` card, and a link to the SPEC the card already names (`changed: false`) are never refused.
- REQ-10: The `card ready` help and the `card spec` help shall each say the refusal of REQ-8 or REQ-9 and the way out, in words that match the refusal text.
- REQ-11: The decision of whether a reclaim is the takeover of an ended lease shall be one function that takes the reservation status, the lease end and the clock reading (`Result<i64>`), so that a test can pass an unreadable clock.
- REQ-12: Each guard-mutant line of a READY note of this card shall come from a run of exactly one mutant, with its own log file that names the mutant and ends with the exit code.

## Acceptance

- AC-1 (REQ-1): 정리한 공백으로 읽은 `card note --help` 가 `` `READY ` ``(역따옴표와 끝 공백 포함), 앞 공백 거절, 요약 줄 `base` 의 조상 규칙을 말해요. · verify: `cargo nextest run --offline --locked --test card_ready_receipt_cli -E 'test(the_note_help_names_the_receipt_gate)'`
- AC-2 (REQ-1): `src/ops/card.rs` 의 `` `READY ` `` 마커 이름을 바꾼 버려지는 복사본에서 AC-1 의 명령이 0 이 아닌 코드로 끝나요(뮤턴트 M8 이 이제 죽어요). · verify: `cp -c` 한 복사본에서 `` `READY ` `` 를 `` `RDY ` `` 로 바꾼 뒤 AC-1 의 명령을 돌려 exit 100
- AC-3 (REQ-2): `card reclaim --help` 가 임대가 끝난 뒤에만 안 푼 커밋이 인수를 막지 않는다고 말하고, 그 말이 빠지면 테스트가 실패해요. · verify: `cargo nextest run --offline --locked --test card_reclaim_early_cli -E 'test(the_reclaim_help_says_when_unpushed_commits_stop_the_takeover)'`
- AC-4 (REQ-6): `held_for_review` 와 `expired_candidate` 예약을 임대 끝 뒤로 만들고, 안 푼 커밋이 있는 워크트리를 기록해 두고 `card reclaim` 하면 거절되고 장부는 그대로예요. `cur.status == "held"` 조건을 지우면 이 테스트가 실패해요. · verify: `cargo nextest run --offline --locked --test card_reclaim_early_cli -E 'test(a_reservation_that_is_not_held_goes_through_the_strict_gate)'`
- AC-5 (REQ-3): `card withdraw --help` 첫 줄이 네 상태를 말하고 `card take --help` 에는 그 줄이 없어요. · verify: `cargo nextest run --offline --locked --test card_withdraw_cli -E 'test(the_take_and_withdraw_help_first_lines_stay_apart)'`
- AC-6 (REQ-4): fold 된 카드를 `card take` 하면 거절 문장에 `folded into C-2` 가 있고 `withdrawn` 낱말이 없어요. withdraw 된 카드의 거절 문장은 그대로 `withdrawn: <사유>` 를 말해요. · verify: `cargo nextest run --offline --locked --test card_fold_cli --test card_withdraw_cli -E 'test(take_of_a_folded_card_does_not_say_withdrawn) | test(take_of_a_withdrawn_card_still_says_withdrawn)'`
- AC-7 (REQ-5): 한 SPEC 에 withdraw 된 카드와 fold 된 카드만 있을 때 `card add --spec` 가 그 SPEC 의 새 카드를 받고, 열린 카드가 하나 있으면 이전처럼 거절해요. · verify: `cargo nextest run --offline --locked --test card_add_spec_cli -E 'test(a_hold_card_is_not_an_open_card_of_its_spec)'`
- AC-8 (REQ-7): `tests/card_stale_cli.rs` 의 "경고 없음" 확인이 stale 낱말 하나가 아니라 `warnings` 배열 전체를 같은 fixture 의 stale 초안 없는 판과 비교해서, 프로필을 못 읽는 경고가 어떤 낱말로 와도 실패해요. 프로필 읽기 경로에 다른 낱말의 경고를 더한 복사본에서 이 테스트가 실패해요. · verify: `cargo nextest run --offline --locked --test card_stale_cli -E 'test(an_unreadable_profile_uses_the_default_days_and_adds_no_warning)'`
- AC-9 (REQ-1..REQ-7): (조각 1) 쓰기 범위 밖의 `src` 파일 diff 가 0이고 `src/card/dispatch.rs` 는 그대로예요. 조각 1 의 base `b18c7639a9` 에서 `0d4106ae6a` 까지 이미 확인했어요. · verify: `git diff --stat b18c7639a9..0d4106ae6a -- src`
- AC-10 (REQ-8): 임시 장부에서 SPEC 의 열린 카드 하나가 있을 때, 관계 없이 그 SPEC 에 연결된 `hold` 카드를 `card ready` 하면 `conflict in spec` 으로 거절되고(열린 카드 이름이 문장에 있어요) 카드는 `hold` 그대로이며 이벤트가 안 늘어요. 열린 카드가 `done` 이 되거나 `hold` 가 되면 같은 `card ready` 가 받아들여져요. `follow-up` 관계의 `hold` 카드와 SPEC 이 없는 `hold` 카드는 열린 카드가 있어도 받아들여져요 · verify: `cargo nextest run --offline --locked --test card -E 'test(a_hold_card_does_not_become_ready_beside_an_open_card_of_its_spec)'`
- AC-11 (REQ-9): 임시 장부에서 열린 카드가 있는 SPEC 에 다른 열린 카드를 `card spec` 하면 `conflict in spec` 으로 거절되고 이벤트가 안 늘어요. 같은 SPEC 로 다시 연결(`changed: false`), `--unlink`, `done` 카드의 연결, 열린 카드가 없는 SPEC 로의 연결은 받아들여져요 · verify: `cargo nextest run --offline --locked --test card -E 'test(card_spec_does_not_add_a_second_open_card_to_a_spec)'`
- AC-12 (REQ-10): 공백을 접어 읽은 `card ready --help` 와 `card spec --help` 가 거절과 빠져나갈 길을 말하고, 그 문장을 지우면 테스트가 실패해요 · verify: `cargo nextest run --offline --locked --test card -E 'test(the_ready_and_spec_help_name_the_open_card_refusal)'`
- AC-13 (REQ-6, REQ-11): `dispatch.rs` 안 단위 테스트가 `held` + 읽히는 시계 + 끝난 임대는 느슨한 문, `held_for_review`·`expired_candidate` 상태, 읽히지 않는 시계(`Err`), 끝나지 않은 임대는 각각 엄격한 문을 낸다는 것을 고정하고, 기존 `card_reclaim_early_cli` 테스트가 고치지 않고 통과해요 · verify: `cargo nextest run --offline --locked --lib -E 'test(the_takeover_gate_choice_follows_status_clock_and_lease)'` 와 `cargo nextest run --offline --locked --test card_reclaim_early_cli`
- AC-14 (REQ-12): READY 노트의 `guard-mutant:` 줄마다 `target/card-help-gates-2/mut-<이름>.log` 가 하나씩 있고 모두 `exit=100` 으로 끝나요. 뮤턴트 하나를 되돌리기 전에 다음 뮤턴트를 적용하지 않아요 · verify: `grep -L 'exit=100' target/card-help-gates-2/mut-*.log` 가 아무것도 내지 않고, `ls target/card-help-gates-2/mut-*.log | wc -l` 가 READY 의 `guard-mutant:` 줄 수와 같아요
- AC-15 (REQ-8..REQ-12): 조각 2 의 `src` diff 는 Scope 의 조각 2 쓰기 목록 안에만 있어요 · verify: `git diff --stat <base>..<head> -- src`

## Scope

조각 1(착지함, 경로는 당시 이름이에요):
- write: `.moai/specs/SPEC-20261124-card-help-gates/spec.md`, `src/ops/card.rs`(공유 파일: `card line-lease` 로 줄만 빌려요), `src/ledger/reservation.rs`, `src/card/spec_link.rs`, `tests/card_ready_receipt_cli.rs`, `tests/card_reclaim_early_cli.rs`, `tests/card_withdraw_cli.rs`, `tests/card_fold_cli.rs`, `tests/card_add_spec_cli.rs`, `tests/card_stale_cli.rs`

조각 2(이 개정이 승인되면 `card scope` 로 더해요):
- write: `.moai/specs/SPEC-20261124-card-help-gates/spec.md`, `src/card/dispatch.rs`(`ready` 한 함수와 `takeover_gates` 의 판정 한 곳, 그리고 그 파일 안 `mod tests`), `src/card/spec_link.rs`(`link` 와 거절 문장 도우미), `src/ops/card.rs`(공유 파일: `card line-lease` 로 `ready`·`spec` 도움말 줄만 빌려요), `tests/fixtures/help/*.txt`(공유 파일: 도움말 골든이 바뀌면 `card line-lease` 로만), `docs/design-review/appendix/operations/card.md`(공유 파일: 거절 한 줄, `card line-lease` 로만), `tests/card/card_withdraw_cli.rs`, `tests/card/card_spec_cli.rs`, `tests/card/card_add_spec_cli.rs`, `tests/card/card_reclaim_early_cli.rs`(읽기와 새 테스트 추가만)
- read-only: `src/card/ready_receipt.rs`, `src/ledger/reservation.rs`
- forbidden: 실제 HOME 과 실제 장부(임시 HOME 만), 네트워크, `CLAUDE.md`, `card reclaim` 의 동작 변경, 이미 열린 카드가 둘인 기존 장부 데이터의 수정, 새 테스트 파일, 셸 스크립트와 Python

## Plan

1. RED: 위 AC 의 테스트를 먼저 써서 실패를 봐요(AC-1, 3, 4, 5, 6, 7, 8). AC-4 의 `held` 조건과 AC-2 의 마커 뮤턴트는 코드를 고치기 전에 한 번 돌려서 테스트가 살아남던 때와 다르게 이제 죽는지 봐요.
2. GREEN: `src/ops/card.rs` 의 도움말(note, reclaim, withdraw 첫 줄)을 `card line-lease` 로 고치고, `reservation.rs` 의 거절 문장과 `spec_link.rs` 의 `open_cards_of` 를 고쳐요.
3. 기존 테스트 중 문장이 바뀐 것(`card_withdraw_cli.rs` 의 고정된 첫 줄, `tests/fixtures/help` 골든이 있으면 그것)을 맞춰요.
4. READY: 같은 head 에서 `moai check affected` 를 한 번, 뮤턴트 줄을 새 가드 테스트마다 하나씩 적어요.

조각 2 계획(승인된 뒤에만 시작해요; 지금은 SPEC 만 써요):

5. (지금, SPEC 폴더만) 이 개정을 채우고 `moai spec check` 를 돌려 SPEC-READY 노트로 올리고 멈춰요. 승인은 사용자의 것이라 이 레인이 기록하지 않아요.
6. 승인 뒤 `moai card scope` 로 조각 2 쓰기 범위를 더하고, 쓰기 전마다 `moai card check` 를 해요. `src/ops/card.rs`·골든·`operations/card.md` 는 `card line-lease` 로만 쓰고 커밋 직후 돌려줘요.
7. RED(`moai tdd red`, 슬롯 `--max 2`): AC-10..AC-13 의 테스트를 먼저 써요. 변경 전 트리에서 AC-10·AC-11·AC-12 는 단언 실패여야 해요(지금은 거절이 없고 도움말에 문장이 없어요). AC-13 은 판정 함수가 없어 컴파일되지 않아서 RED 가 단언 실패가 되도록, 함수만 먼저 현재 식 그대로 빼는 작은 커밋(`Moai-Phase: refactor`, 동작 그대로)을 RED 앞에 두고 AC-13 의 단위 테스트가 현재 식으로 통과하는지 먼저 봐요(통과하는 가드이므로 뮤턴트로 확인해요). 기존 `card_reclaim_early_cli` 는 이 커밋 앞뒤로 같은 결과여야 해요.
8. GREEN: `spec_link::open_cards_of` 를 부르는 `conflict in spec` 도우미(verb 를 `add`/`ready`/`spec` 로 받는 한 곳, `add_conflict` 의 문장 모양은 그대로)를 만들고 `dispatch::ready`(hold 일 때만)와 `spec_link::link`(연결할 때만, 카드가 열려 있을 때만)가 같은 트랜잭션 안에서 부르게 해요. 카드가 SPEC 에 건 관계는 `card.spec` 이벤트의 `relation` 으로 읽어요(`link_json` 과 같은 읽기). 도움말 줄(REQ-10)은 `line-lease` 로 고쳐요.
9. 뮤턴트는 가드 테스트마다 하나씩, 한 번에 하나만 적용해 `target/card-help-gates-2/mut-<이름>.log` 로 남겨요(AC-14). 뮤턴트 후보: 도우미 호출 지움, `hold` 조건 지움, 관계 검사 지움(`follow-up` 도 거절됨), `done`·`hold` 제외 지움, `--unlink` 가 거절됨, 판정 함수의 `held` 조건 지움, 시계 `Err` 가 느슨한 문이 됨, 도움말 문장 지움.
10. READY: 같은 head 에서 `check affected --card` 맨몸으로, `repo_scans_table`·`state_folder`·`source_manifest`·`home_isolation_scan` 은 새 테스트 파일을 만들지 않으니 `check affected` 가 고르는 대로 따라가요.

## Risks

- `tests/fixtures/help/**` 골든이 `card withdraw` 첫 줄을 담고 있으면 그 파일이 깨져요(공유 파일이라 line-lease 로 고쳐요). 읽어서는 아직 담고 있는 파일이 없다고 봤으나(rg, 2026-11-24) 확인은 RED 에서 해요.
- `open_cards_of` 를 바꾸면 `card add --spec` 가 `hold` 카드가 있는 SPEC 에 두 번째 카드를 받아요. 이것이 O6 의 뜻이고, 열린 카드는 여전히 막아요.
- 표시 문장을 바꾸면 `withdrawn: <사유>` 를 고정한 다른 테스트가 있을 수 있어요(RED 에서 `rg` 로 확인).
- (조각 2) `card spec` 은 관계를 적지 않아요(`card spec` 으로 만든 연결은 관계가 없어요). REQ-9 를 켜면, 열린 카드가 있는 SPEC 에 이미 있는 다른 카드를 나중에 `card spec` 으로 붙이려던 길이 막혀요. 빠져나갈 길은 새 카드를 `card add --spec <SPEC> --follow-up` 으로 만들거나, 일을 기존 열린 카드의 기준으로 `card criterion-add` 하는 것이에요. 곧 쓸 규칙이 지금 장부에서 거절하는 것은 없어요(위 사실). D-4 에서 `card spec` 에 관계 옵션을 더하는 길도 있어요.
- (조각 2) `ready` 는 되살리는 `hold` 카드만 보고 `draft` 는 안 봐요: `draft` 는 이미 열린 카드로 세어져서 `card ready` 가 열린 카드를 늘리지 않아요.
- (조각 2) 판정 함수로 빼면서 `SELECT unixepoch()` 를 부르는 자리가 달라지지 않게 `takeover_gates` 가 읽은 시계 값을 그대로 넘겨요. 느슨한 문이 열리는 조건이 하나라도 달라지면 `card_reclaim_early_cli` 가 깨지도록 REQ-11 이 네 가지를 고정해요.
- (조각 2) `src/card/dispatch.rs` 는 카드 명령의 큰 파일이에요. 쥐고 있던 C-445 는 끝났지만, 착지 전 `card scope` 의 `overlaps` 로 겹치는 열린 카드를 다시 봐요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 시계 분기(`SELECT unixepoch()` 실패 → 엄격한 문)는 이 카드에서 테스트하지 않아요. 카드 기준 2 의 그 반쪽은 `src/card/dispatch.rs` 에 넣을 수 있는 자리가 생긴 뒤(C-445 착지 후)의 후속 기준으로 남겨요 | SQLite 의 `unixepoch()` 는 실패하지 않아서 CLI 로 만들 수 없고, 단위 테스트는 `dispatch.rs` 안에 있어야 하는데 C-445 가 쥐어요 | 후속 기준을 닫기 |
| D-2 | 도움말의 마커·조상 문장은 `src/card/ready_receipt.rs` 의 실제 동작에서 가져와요(고치는 것은 문장이지 동작이 아니에요) | core rule 21: 지침과 코드가 다르면 코드가 이겨요 | 문장을 되돌리기 |
| D-3 | (승인 때 골라요, 조각 2) 새로 거절하는 길: **A** `card ready`(hold 되살림)와 `card spec` 둘 다 `card add` 와 같은 규칙으로 거절(기본안, 카드 기준 8 의 글) · **B** `card ready` 만 거절하고 `card spec` 은 그대로(되살림 틈만 막아요) · **C** 둘 다 안 막고 틈을 문서로만 남겨요 | A: 열린 카드가 둘인 SPEC 이 `card add` 로는 못 생기는데 두 길로는 생겨서 규칙이 반쪽이에요. 대가는 위 Risks: `card spec` 으로 두 번째 열린 카드를 붙이는 길과 관계 없는 `hold` 카드를 되살리는 길이 막혀요(지금 장부에서 거절되는 것은 0개). B: `card spec` 은 카드를 나중에 SPEC 에 붙이는 길이라 막으면 흔한 흐름이 걸릴 수 있어요(장부의 `card.spec` 이벤트 354개 중 351개가 관계 없는 연결인데, `card add --spec` 의 plain add 와 `card spec` 을 이 숫자로 가를 수 없어서 얼마나 흔한지는 확인하지 못했어요). B 는 대가가 작고 틈이 한 길 남아요. C: 동작 변경이 없고 틈이 둘 다 남아요. 사용자 선택 필요 | A→B 는 `link` 의 호출 한 줄을 지움, 둘 다 되돌리기는 도우미 호출 지움 |
| D-4 | (승인 때 골라요, D-3 이 A 일 때만) `card spec` 에 관계 옵션: **a** 안 더해요(기본안, 빠져나갈 길은 `card add --spec --follow-up` 과 `card criterion-add`) · **b** `--follow-up`·`--follow-up-of <card>`·`--split-of <card>` 를 `card spec` 에 더해 관계를 적고 그 관계면 거절하지 않아요 | a: `card spec` 의 모양이 그대로고 변경이 작아요. b: 이미 만든 카드를 SPEC 의 후속으로 붙일 수 있지만 `card spec` 의 옵션·도움말·골든·운영 문서가 늘고 `origin_refusal` 을 다시 불러야 해서 범위가 커져요 | b→a 는 옵션 지움 |
| D-5 | (승인 때 골라요, 조각 2) 기준 2 의 시계 반쪽: **a** 판정을 함수로 빼고 `dispatch.rs` 안 단위 테스트로 네 가지를 고정(기본안) · **b** 시계 반쪽은 테스트하지 않고 기준 2 의 글을 리더가 줄이도록 남겨요(`unixepoch()` 는 실패하지 않아요) | a: `takeover_gates` 의 식 하나를 함수로 옮길 뿐 동작은 그대로이고(REQ-11), 지우면 통과하던 `held` 조건과 시계 `Err` 가지가 둘 다 죽어요. 대가는 `dispatch.rs` 열 줄 안팎의 변경과 `card reclaim` 이 게이트 하나를 지나는 길이 바뀌었는지 `card_reclaim_early_cli` 로 한 번 더 확인하는 일이에요. b: 코드 변경이 없지만 시계 가지를 지워도 통과하는 상태가 남고 카드 기준 2 를 `met` 로 못 적어요 | a→b 는 함수를 식으로 되돌리기 |

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

조각 2(개정 2): 아직 시작하지 않았어요. 위 조각 1 결과는 개정 1 의 것이고 AC-10..AC-15 의 결과는 승인과 구현 뒤에 적어요.
