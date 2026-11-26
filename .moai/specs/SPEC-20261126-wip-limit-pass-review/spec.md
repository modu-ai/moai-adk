---
id: SPEC-20261126-wip-limit-pass-review
title: "wip 한도: 지금 head 에 PASS 판정이 있는 review 카드는 작업 수 상한에서 빼요"
status: draft
revision: 1
created: 2026-11-26
---

# wip 한도: 지금 head 에 PASS 판정이 있는 review 카드는 작업 수 상한에서 빼요

## Intent

작업 수 상한(`work_flow.wip_limit` × 살아 있는 레인 수)은 lease 가 살아 있는 `active` 와 `review` 카드를 세요. 독립 감사의 PASS 가 기록되고 착지만 기다리는 `review` 카드도 한 칸을 차지해서, 레인이 할 일이 남지 않은 카드 때문에 다른 `card take` 가 막혔어요. 풀려면 보유자가 카드를 `reopen` 하고 `release` 해야 했어요. 착지 대기는 리더의 몫이므로, 지금 제출한 head 에 PASS 가 기록된 `review` 카드는 세지 않아요. 상한은 관문이라서, 지금 head 를 확인할 수 없는 카드는 전처럼 세요.

## Interview

- Q1: 어떤 `review` 카드를 빼나요? → A: 지금 제출한 head 에 PASS 판정이 기록된 카드만이에요. 판정 없는 카드를 모두 빼는 안과 리더의 "감사 대기" 표시는 하지 않아요(카드의 LEADER-DECISION 노트, 사용자 위임에 따른 리더 판단).
- Q2: "지금 head" 는 어디서 읽나요? → A: 카드의 가장 새 hand-back 노트의 첫 줄이에요. 원장의 `card.review` 이벤트에는 head 가 없어요(D-1).
- Q3: 확인이 안 되면 어느 쪽으로 틀리나요? → A: 세는 쪽이에요. 상한을 약하게 만들지 않아요(REQ-3, REQ-5).
- Q4: 상한 숫자나 새 쓰기 명령이 있나요? → A: 없어요(REQ-8, Non-goals).

## Non-goals

- 상한 숫자(`per_lane`, `lanes`, `limit`)와 상한을 끄는 조건은 바꾸지 않아요.
- `card next` 의 `audit_needed`, 착지 관문, `reclaim` 과 `release` 의 판정 규칙은 바꾸지 않아요.
- 새 쓰기 명령("감사 대기" 표시)은 만들지 않아요.
- 합친 head 에만 PASS 가 있는 경우를 git 으로 추적하지 않아요.

## Requirements

- REQ-1: The work-in-progress limit shall leave a `review` card whose lease is alive out of its count only when the card has a current head and that head is passed.
- REQ-2: The current head of a card shall be the head that `bundle::parse_ready` reads for that card from the first line of its newest hand-back note, where a note by `leader` or of kind `decision` is no hand-back note.
- REQ-3 (REQ-1): IF the card has no hand-back note or its newest one does not parse or the `card.note` event of that note is missing or is not newer than every `card.reopen` and `reserve` and `card.reclaim` event of the card, THEN the work-in-progress limit shall count the card and read no older note.
- REQ-4: A current head shall count as passed only while a `card.verdict` event names exactly that head with the newest verdict of every reviewer on it being PASS and the newest verdict of the card on any head being PASS.
- REQ-5 (REQ-4): IF a `card.verdict` row of the card cannot be read as a verdict, THEN the work-in-progress limit shall count the card.
- REQ-6: The refusal text and the `wip` object of `card next` shall list each card left out by REQ-1 apart from the counted cards with the reason `pass verdict` and its current head.
- REQ-7: The commands `card take` and `card add` and `card next` shall use one count.
- REQ-8: The work-in-progress limit shall keep `per_lane` and `lanes` and `limit` and the off switches and the count of `active` cards and the reasons `lease ended` and `no reservation` unchanged.

## Acceptance

- AC-1 (REQ-1, REQ-6, REQ-7): with one lane, a card in work and a passed review card, `card next` shows a count of 1 with `at_limit` false and a `not_counted` entry with the reason, the head and a null `lease_until`; `card add` and `card take` pass; at the limit both refuse and name the passed card as not counted · verify: `cargo test --locked --offline --test card card_wip_limit::pass_review_ac1_`
- AC-2 (REQ-1, REQ-4): these cards count: no verdict; a FAIL after a PASS on the head; one reviewer's FAIL then another's PASS; a PASS on the head then a FAIL on another head; a PASS on another head only. A PASS on the head followed by a PASS on another head stays left out · verify: `cargo test --locked --offline --test card card_wip_limit::pass_review_ac2_`
- AC-3 (REQ-2, REQ-3): these cards count: reopened and handed back with another head; reopened and sent to review without a note; released, taken again and sent to review without a note; no hand-back note; a newest hand-back note that does not parse over an older passed one. A later note by the leader or of kind `decision` changes nothing; the same head handed back again with a new note stays left out · verify: `cargo test --locked --offline --test card card_wip_limit::pass_review_ac3_`
- AC-4 (REQ-5): a verdict row that cannot be read beside a PASS makes the card count · verify: `cargo test --locked --offline --test card card_wip_limit::pass_review_ac4_`
- AC-5 (REQ-8): an `active` card with a passed head still counts; `per_lane`, `lanes` and `limit` are the same before and after a PASS; `wip_limit` 0 stays off; the two lease reasons are unchanged · verify: `cargo test --locked --offline --test card card_wip_limit::pass_review_ac5_`
- AC-6 (REQ-8): the module's earlier tests pass without an edit · verify: `cargo test --locked --offline --test card card_wip_limit::`
- AC-7: light checks · verify: `git diff --check`; `cargo fmt --check`; `cargo clippy --locked --offline --all-targets -- -D warnings`; `moai spec check`; `moai trailer-check`

## Scope

- write: `src/card/wip_limit.rs`, `tests/card/card_wip_limit.rs`, `.moai/specs/SPEC-20261126-wip-limit-pass-review/**`
- read-only: `src/bundle.rs`, `src/card/ready_receipt.rs`, `src/card/inbox.rs`, `src/card/work.rs`, `src/card/command.rs`, `src/card/dispatch.rs`, `src/ledger/workflow.rs`, `src/ledger/reservation.rs`
- forbidden: `src/bundle.rs`, `src/ledger/`, `src/card/land/`, `src/lane_registry.rs`, `tests/common/`, the ledger and the registry of a real project

## Plan

1. RED: in `tests/card/card_wip_limit.rs` add the tests `pass_review_ac1_…` to `pass_review_ac5_…` with their helpers, before the live-lock section. Each "counts" case first shows the card left out and then counted, so that every new test fails on this tree and none passes without the change.
2. GREEN: in `src/card/wip_limit.rs` add the passed head to a card in work, one function that decides it (REQ-2 to REQ-5), the partition in `read`, and the reason and head in `json` and in the refusal text. The refusal's heading for the left-out cards no longer says "no live lease".
3. Guard mutants, one at a time: the head match; the per-reviewer FAIL check; the newest-on-any-head check; each of the three boundary event kinds; the fall back to an older note; the leader and decision filter; the raw row count; the `review` state condition. Each must fail a named test.
4. A run of the real binary on a throwaway HOME and project under the worktree's `target/`: a limit that is reached, then a PASS that frees one place, then a reopen with a new head that takes it back.
5. Record the Result with the commands and exit codes.

## Risks

- This loosens a gate. The guard mutants and the "counts" cases hold the strict side.
- A lane that commits during review without a new hand-back note stays left out. A landing still needs the exact head.
- A PASS that the landing would refuse (no receipt, the reviewer wrote the head) still leaves the count.
- A PASS recorded only on a batch merge head keeps the card counted.
- `audit_needed` hides a card that has any verdict, so a card handed back with a new head after a verdict counts here but is not offered for audit. This is not changed by this SPEC.
- The reads are not one transaction, as the limit's earlier reads are not.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | The current head is the one the newest hand-back note names, tied to the card's work cycle by the note's event id. | `card.review` carries no head; `card submit` writes the note before the review, so the order is read against the events that put a card back into work. | Record the head on the review event and read it there. |
| D-2 | A verdict counts only on exactly that head; a PASS on a merge head that contains it does not replace one on the head. | It needs no git call and fails on the strict side. | Add a clean-merge coverage check through git. |
| D-3 | Any recorded PASS on the head is enough; receipt and reviewer identity are the landing's checks. | The card asks for a recorded PASS. | Require the landing's conditions. |
| D-4 | The rule that a note by `leader` or of kind `decision` is no hand-back note is repeated in this module. | `bundle` keeps its method private and is not in this card's scope. | Make the method public and call it. |
| D-5 | The same head handed back again after a reopen with a new note keeps its PASS. | The card counts a card again when the head changes. | Require a verdict newer than the reopen. |
| D-6 | The card is a tier 2 gate change, so the SPEC needs the leader's delegated approval before code is written. | The change loosens a gate. | none |

## Result

구현했어요(브랜치 `spec/wip-limit-pass-review-20261126`, 기반 `91dd561a5a`). 위임 승인은 카드의 DELEGATED-APPROVAL 노트(개정 1, approval_digest `sha256:8a89079a…39f4`)이고, 승인 front matter 는 그 노트의 지시대로 착지 때 리더가 써요. 커밋은 초안 `293d7a5c8f`, red 테스트 `29c933c7fe`, 코드 `f2126ea02c`(`src/card/wip_limit.rs` 한 파일)예요. 독립 감사는 이 레인이 하지 않아요.

- 바뀐 것: lease 가 살아 있는 `review` 카드마다 현재 head 를 읽어요. 가장 새 hand-back 노트(리더와 `decision` 노트는 빼요)의 head 이고, 그 노트의 `card.note` 이벤트가 `card.reopen`·`reserve`·`card.reclaim` 보다 뒤일 때만이에요. 그 head 를 정확히 가리키는 판정이 있고, 검토자마다 가장 새 판정이 PASS 이고, 카드의 가장 새 판정도 PASS 이고, 읽히지 않는 판정 행이 없으면 상한에서 빼요. 빠진 카드는 `card next` 의 `wip.not_counted` 에 `reason` "pass verdict" 와 `head` 로 나오고, 거절 문구는 "; not counted: … PASS on head <12자>" 로 알려요. 거절 문구의 제목에서 "(no live lease)" 는 없앴어요. 상한 숫자, 끄는 조건, `active` 카드 세기, `audit_needed` 는 그대로예요.
- red, `29c933c7fe`: `cargo nextest run --locked --offline --no-fail-fast --test card -E 'test(/card_wip_limit::/)'` 29개 실행, 24 통과, 5 실패, exit 100. 새 테스트 5개가 모두 "PASS 카드가 아직 세어져요" 로 실패했어요(`target/c535/red.log`).
- green(`f2126ea02c` 의 작업 트리): `cargo nextest run --locked --offline --no-fail-fast --test card -E 'test(/card_wip_limit::/) | test(/card_next/)'` 34개 실행, 34 통과, exit 0(`target/c535/green.log`).
- clippy: `cargo clippy --locked --offline --all-targets -- -D warnings` exit 0(`target/c535/clippy.log`).
- 임시 HOME·프로젝트에서 빌드한 실제 바이너리(`target/c535/dry.sh`, exit 0, 출력 `target/c535/dry.out`): 레인 1개(상한 2)에서 review 카드와 active 카드로 `card take` 가 거절됐어요. 그 head 에 PASS 를 기록하자 `count` 1, `not_counted` 에 그 카드와 head 가 나오고 take 가 통과했어요. 다시 상한에 닿으면 거절 문구가 "not counted: <카드> (review, t1, PASS on head 9134e6ec4872)" 를 보였어요. `reopen`, 노트 없는 재제출, 새 head 의 노트(판정 없음), 그 head 의 FAIL 에서는 다시 세어졌고(`count` 3, 거절), 새 head 의 PASS 에서 다시 빠졌어요. `wip_limit` 0 은 꺼진 채였어요.
- 가드 뮤턴트(`target/c535/mutants.out`, 각 `--test card -E 'test(/card_wip_limit::/)'`, 되돌린 뒤 `git status --short` 는 비어 있었어요):
  - guard-mutant: pass_review_ac2 | 판정의 head 일치 | exit 100
  - guard-mutant: pass_review_ac2 | 검토자별 FAIL 검사 | exit 100
  - guard-mutant: pass_review_ac2 | 카드의 가장 새 판정 검사 | exit 100
  - guard-mutant: pass_review_ac4 | 판정 행 수 비교 | exit 100
  - guard-mutant: pass_review_ac3 | 경계에서 `card.reopen` 제거 | exit 100
  - guard-mutant: pass_review_ac3 | 경계에서 `reserve` 제거 | exit 100
  - guard-mutant: pass_review_ac3 | 경계에서 `card.reclaim` 제거 | exit 100
  - guard-mutant: pass_review_ac3 | 읽히지 않는 최신 노트 뒤의 옛 노트로 넘어감 | exit 100
  - guard-mutant: pass_review_ac3 | 리더·decision 노트 거름 | exit 100
  - guard-mutant: pass_review_ac5 | `review` 상태 조건 | exit 100
- 가벼운 검사: `git diff --check` exit 0, `cargo fmt --all -- --check` exit 0, `moai trailer-check 91dd561a5a..HEAD` failing 0, `moai spec check` 오류 0.

- AC-1: met — `pass_review_ac1_a_passed_review_card_leaves_room_and_is_listed_apart` 가 red 에서 실패하고 green 에서 통과해요.
- AC-2: met — `pass_review_ac2_only_a_head_that_is_still_passed_leaves_the_count`, 같은 red/green.
- AC-3: met — `pass_review_ac3_only_the_head_of_the_current_hand_back_counts_as_passed`, 같은 red/green.
- AC-4: met — `pass_review_ac4_a_verdict_row_that_cannot_be_read_makes_the_card_count`, 같은 red/green.
- AC-5: met — `pass_review_ac5_only_a_review_card_changes_and_the_rest_stays`, 같은 red/green.
- AC-6: met — 모듈의 기존 테스트는 고치지 않았고(추가만 468줄, 삭제 0) 같은 green 실행에서 통과해요.
- AC-7: met — 위의 가벼운 검사와 clippy.

알려 둘 것:

- SQL 읽기가 실패하면 오류가 올라가서 `take`·`add` 가 거절돼요. 그 실패를 억지로 만드는 테스트는 없어요.
- `audit_needed` 는 판정이 하나라도 있는 카드를 감사 목록에서 빼요. 새 head 로 다시 낸 카드는 여기서는 세지만 감사 목록에는 없어요(리더가 따로 카드로 열었어요).

not_run:

- `check affected --card` 영수증은 이 Result 가 든 head 를 이름 짓는 READY 노트가 담아요.
- 실제 프로젝트 원장에 대한 이 빌드의 실행, 다른 운영체제, 독립 감사.
