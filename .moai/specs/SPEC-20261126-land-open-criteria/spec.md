---
id: SPEC-20261126-land-open-criteria
title: "A landing keeps open the criteria the verdict holds open"
status: draft
revision: 1
created: 2026-11-26
---

## Interview

- none: 이 카드는 리더가 착지 실수에서 연 카드이고, 사용자에게 간 질문은 없어요. 남은 선택은 Decisions의 D-1과 D-2에 선택지와 기본값을 적었고 승인 때 정해요.

# A landing keeps open the criteria the verdict holds open

## Intent

2026-11-26 여덟째 묶음에서 `card land`를 `--keep-open` 없이 부르자, 열린 기준 전부가 `verdict event <n> PASS` 근거로 met이 되고 두 카드가 닫혔어요. 감사 판정의 근거 글에는 "그 기준은 열려 있어요"라고 적혀 있었지만, 착지는 근거 글을 읽지 않고 판정 기록에는 열어 둘 기준을 적는 자리가 없어요. 완료한 카드를 되살리는 길도 없어요.

이 SPEC은 감사자가 판정에 "열어 둘 기준"을 번호로 적고, 착지가 그 번호를 met으로 바꾸지 않으며 카드를 닫지 않게 해요. 작성자가 착지 전에 그 기준을 met으로 표시했어도 감사자의 보류가 이겨요. 표준 흐름에서는 `card submit`이 열린 기준이 없을 것을 요구하므로, 감사 시점의 기준은 보통 이미 met이에요.

착지 관문을 바꾸는 일이라 tier 2예요. 바뀌는 방향은 "닫지 않는 쪽" 하나뿐이에요.

## Non-goals

- 열린 번호를 적지 않은 판정의 착지 동작은 바꾸지 않아요(카드 기준 2).
- 판정의 근거 글을 해석하지 않아요. 번호로 적은 것만 보류예요.
- `card finish`와 `card criterion`의 동작, 원장 스키마는 바꾸지 않아요.
- 그때 닫힌 두 카드는 고치지 않아요: 남은 일은 후속 카드로 옮겨져 있어요.

## Requirements

- REQ-1: When `card verdict` records a verdict, the verdict event shall hold the `--open <n>` numbers under `open`, sorted and without repeats, and no `open` key when none is given.
- REQ-2 (REQ-1): IF `--open` is given with a `fail` verdict, or names a number that is not a criterion of the card or is a withdrawn criterion, THEN `card verdict` shall refuse and write nothing.
- REQ-3: When `card land` reads the verdicts of a landing head, also inside `card land-batch` and on a rerun for a head the branch already holds, the landing shall take as held the union of the `open` numbers of each reviewer's newest PASS on the verdict head, leaving out withdrawn criteria. · failure: none — a ledger read error refuses the landing before the first write, as a verdict read does today.
- REQ-4: When the held set is not empty, the landing shall land the head as a slice that records `card.landed` with `keep_open` true, marks no criterion met, keeps the card open and answers `held` as a list of `{seq, reviewer, event}` together with `open_criteria`. · failure: none — the one transaction writes the whole landing or nothing, so a refusal leaves the criteria and the card state as they were.
- REQ-5 (REQ-4): When a held criterion is already `met`, the landing shall set it back to `open` in the same transaction as the landing event, with the replaced evidence recorded in that event. · failure: none — the same transaction as REQ-4.
- REQ-6: IF `--met` names a held criterion, THEN `card land` shall refuse before the first write with a message that names the held numbers and their verdict events.
- REQ-7: When no newest PASS on the verdict head holds a criterion, `card land` and `card land-batch` shall write and answer as before this change. · failure: none — the existing landing tests pass without an edit.
- REQ-8: When the stored `open` of a verdict event is present and is not a list of positive integers, the ledger reader shall not read that event as a verdict. · failure: none — a head with no readable PASS is refused with `no_pass` as today.
- REQ-9: The audit guidance `skills/moai/references/audit.md` shall tell the auditor to record every criterion it leaves open with `--open <n>`, and that the evidence text is not read. · failure: none — guidance text.

## Acceptance

- AC-1 (REQ-3, REQ-4): in a temporary ledger and repository, a card with open criteria 1 and 2 and a PASS with `--open 2` lands: the branch moves, the card is not `done`, criteria 1 and 2 stay open, the answer names `held` 2 and the landing event has `keep_open` true · verify: `cargo nextest run --locked --offline --test card card_land_open_criteria::a_held_criterion_keeps_the_card_open_and_marks_nothing`
- AC-2 (REQ-5): with criteria 1 and 2 both marked met by the holder and a PASS with `--open 2`, the landing leaves 1 met, sets 2 back to open, keeps the card open and the landing event carries the replaced evidence · verify: `cargo nextest run --locked --offline --test card card_land_open_criteria::a_held_criterion_the_author_marked_met_is_opened_again`
- AC-3 (REQ-1, REQ-2, REQ-8): `--open` with `fail`, with an unknown number and with a withdrawn number each refuse and add no event; repeated and unsorted numbers are stored once and sorted; a PASS without `--open` has no `open` key and its repeat finds the same event; an event whose `open` is a string is no verdict · verify: `cargo nextest run --locked --offline --test card card_land_open_criteria::card_verdict_stores_only_a_hold_it_can_keep`
- AC-4 (REQ-6): `card land --met 2=ref` on a head whose PASS holds 2 refuses, the branch stays and no event is written · verify: `cargo nextest run --locked --offline --test card card_land_open_criteria::land_refuses_met_for_a_held_criterion`
- AC-5 (REQ-3): with two reviewers the hold is the union of their newest PASS records; a reviewer's newer PASS without `--open` releases that reviewer's hold only · verify: `cargo nextest run --locked --offline --test card card_land_open_criteria::the_hold_is_the_union_of_each_reviewers_newest_pass`
- AC-6 (REQ-3, REQ-4, REQ-7): `card land-batch` with one held head and one clean head lands both, the held card stays open and the clean card is closed · verify: `cargo nextest run --locked --offline --test card card_land_open_criteria::land_batch_lands_a_held_head_as_a_slice`
- AC-7 (REQ-7): the landing tests that exist at the base commit pass without an edit · verify: `cargo nextest run --locked --offline --test card card_land`
- AC-8 (REQ-9): the audit guidance names the option · verify: `grep -F -e '--open <n>' skills/moai/references/audit.md`

## Scope

- write: `src/card/land.rs`
- write: `src/card/land/batch.rs`
- write: `src/card/land/check.rs`
- write: `src/card/verdict.rs`
- write: `src/ledger/workflow.rs`
- write: `tests/card/card_land_open_criteria.rs`
- write: `skills/moai/references/audit.md`
- write: `skills/moai/references/source-manifest.json`
- write: `.moai/specs/SPEC-20261126-land-open-criteria/**`
- write (shared files, `card line-lease` only): `src/ops/card.rs`, `tests/card/main.rs`, `docs/design-review/appendix/operations/card.md`, and `src/help.rs` only if a help rule asks for a row for the new option
- read-only: `src/card/finish.rs`, `src/ledger/schema.rs`
- forbidden: `src/ledger/schema.rs`, the behaviour of `card finish`, every other path

## Plan

1. 이 초안을 커밋하고 SPEC-READY를 보내요. 승인 노트가 카드에 오기 전에는 코드를 쓰지 않아요.
2. RED: `tests/card/card_land_open_criteria.rs`에 AC-1부터 AC-6까지의 테스트를 먼저 써서 실패를 확인해요.
3. GREEN: 판정 기록(`--open`, 검증, 읽기) → 보류 집합 계산(`land/check.rs`) → 착지(`land.rs`, `workflow.rs`의 한 트랜잭션에서 다시 열기와 조각 착지) → 배치 답 확인 순서예요. 쓰기 전 검사(`finish_preflight`)도 보류를 알아야 해요.
4. 안내 글과 manifest 해시를 고쳐요(Python 없이).
5. 가드 테스트마다 뮤턴트 하나, 임시 원장에서 위험 경로 실제 실행, `cargo fmt --all`, `moai check affected --card` 한 번, `card submit`.

## Risks

- 설치된 이전 `moai`는 `open` 키를 모르고 넘기므로 그대로 닫아요. 보류는 새 빌드를 설치한 착지부터 효과가 있어요. 리더에게 설치 시점을 알려야 해요.
- 감사자가 `--open`을 안 쓰고 근거 글에만 적으면 지금과 같이 닫혀요. 안내 글(REQ-9)이 유일한 예방이에요.
- met을 open으로 되돌리는 것은 착지의 새 쓰기예요. 바꾼 근거를 사건에 남기고 AC-2가 값을 확인해요.
- 손으로 쓴 망가진 `open` 값은 판정으로 읽지 않아서, 같은 검토자의 더 오래된 PASS가 있으면 그것이 읽혀요. `moai`는 그런 값을 쓰지 않아요.
- `src/ledger/workflow.rs`와 `source-manifest.json`은 다른 카드와 겹칠 수 있어요. 착지 때의 합치기는 리더가 해요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 승인 때 고름. 기본값 A: 감사자가 판정에 `--open <n>`을 적은 기준만 보류. B(리더의 첫 지시 문구): 열린 기준이 하나라도 있으면 착지가 멈추고, 닫으려면 명시 플래그를 요구. | A는 카드 기준 1과 2 그대로예요. B는 기준 2(열린 번호를 안 적은 판정은 지금과 같음)와 맞지 않아 카드 정정 노트가 필요하고, 작성자가 감사 전에 met으로 표시한 기준은 막지 못해요. 대신 B는 감사자가 옵션을 잊어도 열린 기준을 지켜요. | `--open` 옵션과 보류 계산을 지움 |
| D-2 | 승인 때 고름. 기본값 C: 보류된 기준이 이미 met이면 착지가 다시 open으로 돌리고 조각으로 착지. A: 쓰기 전 거절(감사자가 보류 없는 새 PASS를 기록해야 착지). B: open이 아니면 보류를 무시. | C는 감사자의 판단이 이기고 head도 착지해요. A는 모든 기준이 met인 카드에서 보통 착지도 `--keep-open`(열린 기준이 없으면 거절)도 안 되어, 감사자가 보류를 풀기 전에는 조각으로 착지할 길이 없어요. B는 작성자가 met 표시로 보류를 풀 수 있어 기준 1과 어긋나요. | REQ-5를 고른 안으로 바꿈 |
| D-3 | `--open`은 open 또는 met인 기준을 받고 withdrawn은 거절. | `card submit`이 열린 기준 0을 요구해서, open만 받으면 표준 흐름에서 옵션을 쓸 수 없어요. | REQ-2의 조건을 바꿈 |
| D-4 | 보류는 검토자별 최신 PASS의 합집합. 착지가 기대지 않는 PASS의 보류도 셈. | 닫지 않는 쪽이 안전해요. 검토자는 새 PASS로 자기 보류를 풀어요. | 기대는 PASS 하나의 보류만 읽음 |
| D-5 | observe 단계 없이 바로 적용. | 기존 판정에는 `open` 키가 없어서 지금의 흐름에서 새로 막히거나 달라지는 착지가 없어요. | 모드 설정 뒤로 옮김 |
