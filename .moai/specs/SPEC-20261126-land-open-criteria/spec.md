---
id: SPEC-20261126-land-open-criteria
title: "A landing keeps open the criteria the verdict holds open"
status: draft
revision: 1
created: 2026-11-26
branch: spec/land-open-criteria-kz-20261126
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

## Result

- 결정: D-1 = A, D-2 = C (승인 노트 9863, 리더의 위임 승인). 승인 블록은 착지 때 리더가 써요.
- 구현: `card verdict --open <n>`, 착지의 보류 집합 계산(`land/check.rs`의 `held`), 보류가 있는 조각 착지와 met 기준 다시 열기(`ledger/workflow.rs`의 한 트랜잭션), `audit.md` 안내 한 문단, 운영 문서의 land·verdict 행.
- RED/GREEN: 새 테스트 8개(`tests/card/card_land_open_criteria.rs`). AC-1~AC-6의 6개에 더해, 보류가 있어도 닫는 착지와 같은 독립 PASS를 요구하는지(`a_hold_asks_for_the_same_pass_as_a_close`), 철회한 기준은 더 보류되지 않는지(`a_withdrawn_criterion_is_held_no_longer`)를 확인해요.
- 기존 테스트: card의 land·verdict·finish·submit·ledger, audit 문구, help, 운영 문서, lib의 card·ledger 모듈에서 777개 통과(0 실패).

SPEC과 달라진 점과 한계예요.

- `tests/card/main.rs`는 프로필의 공유 파일이 아니어서 줄 임대 대신 카드 범위에 넣었어요. `src/help.rs`는 고칠 필요가 없었어요.
- REQ-6보다 넓게, 보류가 있는 head에는 `--met`을 번호와 상관없이 거절해요. 조각 착지는 어떤 기준도 met으로 바꾸지 않아서(REQ-4), 보류 밖 번호의 `--met`을 조용히 버리지 않으려는 거예요.
- 착지 사건이 이미 있는 head의 다시 실행은 착지 사건을 또 쓰지 않아요. 그사이 `met`이 된 보류 기준은 `card.held_reopened` 사건을 쓰며 다시 열어요(감사 수리에서 바뀜).
- 이 옵션을 모르는 이전 설치본 `moai`는 `open`을 넘기고 카드를 닫아요. 새 빌드를 설치한 착지부터 효과가 있어요.
- `card finish`는 바꾸지 않았어요(범위 밖). 보류가 있는 카드도 `card finish`의 기존 조건으로는 닫을 수 있어요.

검증 기록이에요(구현 커밋 `8e335035e8`).

- SPEC 검증 줄: AC-1~AC-6 각 1개 통과, AC-7 `--test card card_land` 177개 통과, AC-8 `grep` exit 0.
- 위험 경로 근거는 Rust 통합 테스트예요(결정 10013, 임시 HOME에서 실제 `moai-cli`): `a_held_criterion_the_author_marked_met_is_opened_again`(AC-2: 보류한 met 기준을 다시 열고 근거를 착지 사건에 남김, 같은 head의 다시 실행은 닫지 않음)과 `land_refuses_met_for_a_held_criterion`(AC-4: 거절, 브랜치 그대로, 사건 없음).
- 첫 영수증(head `ddd048c5b2`)은 `ledger` 단계에서 FAIL이었어요: 새 테스트가 상태 폴더 이름을 직접 적어 `state_folder_literal_scan`에 걸렸어요. 픽스처가 `moai spec new`가 답한 경로를 쓰도록 고쳤고 `allow.tsv`는 바꾸지 않았어요. 상태 폴더 resolver는 이 픽스처(예전 폴더의 프로필)에서 SPEC이 써진 곳과 다른 폴더를 답해 쓰지 않았어요. 아래 뮤턴트는 이 픽스처 수정 전의 `8e335035e8`에서 돌렸고, 수정은 픽스처 준비 부분만 바꿨어요.

guard-mutant 9줄이에요(각 테스트만 실행, 모두 실패, 뒤에 `git checkout`으로 복원).

```
guard-mutant: card_land_open_criteria::card_verdict_stores_only_a_hold_it_can_keep | src/ledger/workflow.rs record_verdict hold_refusal | exit 100
guard-mutant: card_land_open_criteria::card_verdict_stores_only_a_hold_it_can_keep | src/ledger/workflow.rs verdict_from_row held_of unreadable-open rejection | exit 100
guard-mutant: card_land_open_criteria::land_refuses_met_for_a_held_criterion | src/card/land.rs --met refusal while a criterion is held | exit 100
guard-mutant: card_land_open_criteria::a_held_criterion_the_author_marked_met_is_opened_again | src/ledger/workflow.rs reopen_held call | exit 100
guard-mutant: card_land_open_criteria::a_held_criterion_keeps_the_card_open_and_marks_nothing | src/card/land.rs held.is_empty condition around the met fill | exit 100
guard-mutant: card_land_open_criteria::the_hold_is_the_union_of_each_reviewers_newest_pass | src/card/land/check.rs newest-PASS-per-reviewer dedupe | exit 100
guard-mutant: card_land_open_criteria::land_batch_lands_a_held_head_as_a_slice | src/card/land.rs holds read on the verdict head (replaced by the landing head) | exit 100
guard-mutant: card_land_open_criteria::a_hold_asks_for_the_same_pass_as_a_close | src/card/land.rs PASS requirement keyed on the requested keep_open | exit 100
guard-mutant: card_land_open_criteria::a_withdrawn_criterion_is_held_no_longer | src/card/land/check.rs withdrawn-criterion filter | exit 100
```

영수증(`check affected --card`)의 결과와 head는 READY 노트에 적어요.

### 감사 FAIL 수리 (2026-11-26, 판정 노트 10176, 이벤트 20919)

원인: 보류를 걸고 착지한 머지에 판정이 하나 더 붙으면, 같은 `card land`의 다시 실행이 보류를 그 머지에서만 읽어(비어 있음) 시작 행의 PASS에 기대 보류 기준을 `met`으로 바꾸고 카드를 닫았어요. 기준 1(REQ-3)을 어겼어요.

수리 커밋 `6e55426880`:

- 브랜치에 이미 있는 head의 다시 실행은 판정 head, 그 head 자신, 머지면 둘째 부모의 보류를 모두 읽어요. 시작 행이 기댄 판정의 head나 `card.landed.held`만 읽는 방법은 쓰지 않았어요: 앞의 것은 observe 모드에서 기댄 판정 없이 착지한 조각을 놓치고, 뒤의 것은 번호만 있어 해제할 길이 없어요. git의 부모로 읽으면 원장에 새 값을 안 쓰고 두 경우를 모두 덮어요.
- 보류는 보류한 head에 같은 검토자가 새 PASS를 기록해야 풀려요(머지에 적은 PASS는 부모의 보류를 풀지 않아요). `audit.md`와 도움말, 운영 문서에 적었어요.
- 방향은 "닫지 않는 쪽"뿐이에요. 첫 착지가 머지의 판정만 읽은 조각도, 다시 실행에서는 부모의 보류를 읽어요. 첫 착지가 읽는 head는 바꾸지 않았어요(REQ-3 그대로, 감사의 비차단 줄).

함께 고친 비차단 항목:

- 해제한 뒤 같은 보류를 다시 걸면 새 사건을 써요(다시 해제도 같아요). 그 검토자의 그 head 최신 판정이 PASS이고 보류 번호가 달라질 때만이라, FAIL 뒤의 반복은 전처럼 아무것도 안 써요.
- 착지 사건이 이미 있어도 `met`이 된 보류 기준을 다시 열어요(REQ-5).
- `land-batch`의 `landed[]`가 보류된 head에 `held`·`open_criteria`를 실어요.
- 테스트 픽스처의 `take`가 임시 경로를 워크트리로 적어요. 보류 PASS 뒤의 FAIL, `--open 0`, 정수가 아닌 저장값의 테스트를 더했어요.
- `--keep-open`의 "열린 기준 없음" 거절이 보류가 있으면 빠지는 것은 의도예요: 모든 기준이 `met`인 보류 head도 착지하면서 보류 기준을 다시 열어요(D-2 = C).
- 고치지 않은 것: `card finish`·`card done`은 보류를 읽지 않아요(범위 밖). 보류가 있는 머지의 채움 근거 줄은 전처럼 판정 head로 적혀요. 승인된 SPEC의 영어 본문은 바꾸지 않았어요.

검증:

- RED: 수리 전 코드에서 `card_land_open_criteria` 11개 중 5개 실패(`target/keziah/red-558.log`). 감사의 재현과 같은 모양인 `a_rerun_after_a_later_verdict_keeps_the_hold`가 `closed: true`로 실패했어요. `--open 0` 한 건은 명령 해석기가 먼저 거절하는 것이어서 테스트 기대값을 고쳤어요.
- GREEN: 11개 통과. 관련 대상 593개 통과(0 실패), `cargo clippy --all-targets -D warnings` 깨끗함.

두 번째 수리 커밋 `a7e0eddec7`(READY 전에 찾음): 보류를 푼 뒤의 다시 실행이 시작 행의 판정, 곧 보류를 적었던 PASS에 기대 기준을 `met`으로 채우며 닫았어요. 기준 1의 글자에 어긋나요. 닫는 호출에서 보류가 없고 시작 행의 PASS가 보류를 적었으면 지금의 판정을 다시 읽어요(`land/check.rs`의 `read_for_close`). 그래서 머지에 새 FAIL이 있으면 그 머지를 다시 통과시킬 때까지 거절해요. 보류를 적지 않은 시작 행은 전과 같아요. 반복한 `card verdict`는 같은 값의 가장 새 사건을 답해요. RED: 새 단언 3개 테스트가 실패(`target/keziah/red-558c.log`), GREEN: 11개 통과.

head `a7e0eddec7`에서: SPEC 검증 줄 AC-1~AC-8 모두 exit 0(AC-7 180개 통과), `cargo clippy --all-targets -D warnings` exit 0, 관련 대상 593개 통과.

guard-mutant 16줄이에요(head `a7e0eddec7`, 위 9줄을 대신해요. 각 테스트만 실행, 모두 그 테스트에서 실패, 뒤에 `git checkout`으로 복원, 워크트리 깨끗함).

```
guard-mutant: card_land_open_criteria::card_verdict_stores_only_a_hold_it_can_keep | src/ledger/workflow.rs record_verdict hold_refusal | exit 100
guard-mutant: card_land_open_criteria::card_verdict_stores_only_a_hold_it_can_keep | src/ledger/workflow.rs verdict_from_row held_of unreadable-open rejection | exit 100
guard-mutant: card_land_open_criteria::land_refuses_met_for_a_held_criterion | src/card/land.rs --met refusal while a criterion is held | exit 100
guard-mutant: card_land_open_criteria::a_held_criterion_the_author_marked_met_is_opened_again | src/ledger/workflow.rs reopen_held on the landing event | exit 100
guard-mutant: card_land_open_criteria::a_held_criterion_keeps_the_card_open_and_marks_nothing | src/card/land.rs held.is_empty condition around the met fill | exit 100
guard-mutant: card_land_open_criteria::the_hold_is_the_union_of_each_reviewers_newest_pass | src/card/land/check.rs newest-PASS-per-reviewer dedupe | exit 100
guard-mutant: card_land_open_criteria::land_batch_lands_a_held_head_as_a_slice | src/card/land.rs holds read on the verdict head (replaced by the landing head) | exit 100
guard-mutant: card_land_open_criteria::a_hold_asks_for_the_same_pass_as_a_close | src/card/land.rs PASS requirement keyed on the requested keep_open | exit 100
guard-mutant: card_land_open_criteria::a_withdrawn_criterion_is_held_no_longer | src/card/land/check.rs withdrawn-criterion filter | exit 100
guard-mutant: card_land_open_criteria::a_rerun_after_a_later_verdict_keeps_the_hold | src/card/land.rs extra hold heads of a rerun | exit 100
guard-mutant: card_land_open_criteria::a_rerun_after_a_later_verdict_keeps_the_hold | src/card/land/check.rs per-head key of the newest-PASS dedupe | exit 100
guard-mutant: card_land_open_criteria::a_rerun_opens_a_held_criterion_marked_met_since | src/ledger/workflow.rs reopen on a rerun with a recorded landing | exit 100
guard-mutant: card_land_open_criteria::a_hold_set_again_after_a_release_applies | src/ledger/workflow.rs changes_a_hold exception to the repeat | exit 100
guard-mutant: card_land_open_criteria::a_hold_set_again_after_a_release_applies | src/ledger/workflow.rs changes_a_hold newest-verdict-is-PASS condition | exit 100
guard-mutant: card_land_open_criteria::land_batch_lands_a_held_head_as_a_slice | src/card/land/batch.rs held copied into the batch answer | exit 100
guard-mutant: card_land_open_criteria::a_rerun_after_a_later_verdict_keeps_the_hold | src/card/land/check.rs start row not trusted when its PASS held (read_for_close) | exit 100
```
