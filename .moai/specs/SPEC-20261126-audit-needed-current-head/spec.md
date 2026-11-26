---
id: SPEC-20261126-audit-needed-current-head
title: "audit_needed: 지금 head 에 판정이 없는 review 카드는 감사 목록에 보여요"
status: draft
revision: 1
created: 2026-11-26
branch: spec/audit-needed-current-head-20261126
---

# audit_needed: 지금 head 에 판정이 없는 review 카드는 감사 목록에 보여요

## Intent

이 변경 전의 `card next` 는 `card.verdict` 가 하나라도 있는 `review` 카드를 `audit_needed` 에서 모두 숨겼어요. 판정 뒤 수리해서 새 head 로 다시 낸 카드가 감사 목록에 없어서 리더가 손으로 감사자를 찾아야 했어요. 지금 제출한 head 를 정확히 가리키는 판정이 있을 때만 숨기고, 확인할 수 없으면 보여요.

## Interview

- Q1: 어떤 카드를 숨기나요? → A: 지금 head 를 정확히 가리키는 판정이 PASS·FAIL 어느 쪽이든 하나라도 있는 카드예요(카드의 LEADER-DECISION 노트 9971 D-B, 사용자 위임에 따른 리더 판단이고 사용자 직접 선택은 아니에요).
- Q2: FAIL 만 있으면요? → A: 숨겨요. 감사가 아니라 수리를 기다리고, 수리 head 는 새 head 라 다시 보여요(9971).
- Q3: 확인이 안 되면요? → A: 보여요. head 확인 불가, 그 head 의 판정 없음, 읽히지 않는 판정 행이 해당해요(9971).
- Q4: 지금 head 는 어떻게 정하나요? → A: 작업 수 상한과 같은 함수예요(카드 기준 3, SPEC-20261126-wip-limit-pass-review D-1·D-2).

## Non-goals

- 작업 수 상한의 셈, `card next` 의 다른 칸, `src/card/work.rs`, `card orchestrate`, 착지 관문은 바꾸지 않아요.
- `card.verdict` 이벤트 없이 `AUDIT …` 노트만 있는 판정은 전처럼 읽지 않아요.
- 합친 head 의 판정을 git 으로 추적하지 않아요.
- 감사자 배정과 감사 착수 노트는 목록에 반영하지 않아요.

## Requirements

- REQ-1: The `audit_needed` list of `card next` shall name each `review` card that none of REQ-2 and REQ-5 and REQ-7 leaves out, oldest first.
- REQ-2: WHILE the current head of a card is confirmed and every `card.verdict` row of the card reads as a verdict, the list shall leave the card out when at least one verdict of any reviewer, PASS or FAIL, names exactly that head.
- REQ-3 (REQ-2): IF the current head cannot be confirmed as REQ-2 and REQ-3 of SPEC-20261126-wip-limit-pass-review define it, THEN the list shall name the card.
- REQ-4 (REQ-2): IF a `card.verdict` row of the card cannot be read as a verdict, THEN the list shall name the card even when another row names its current head.
- REQ-5: The list shall keep leaving out a card whose reservation of the current generation is held by the asking session.
- REQ-6: The list and the work-in-progress limit shall read the current head and the readable verdicts through one private function.
- REQ-7 (REQ-1): IF the session registry cannot be read, THEN the list shall stay empty.

## Acceptance

- AC-1 (REQ-1, REQ-2): PASS 나 FAIL 뒤 새 head 로 다시 낸 카드(reopen 뒤 재제출, review 중 `REPAIRED` 노트)가 목록에 나오고, 상한에서 `next_is` 가 `audit` 이고 `next` 가 그 카드예요 · verify: `cargo test --locked --offline --test card card_wip_limit::audit_head_ac1_`
- AC-2 (REQ-2): 지금 head 에 PASS, FAIL, 검토자 둘의 엇갈린 판정이 있는 카드는 목록에서 빠져요. 뒤에 다른 head 의 판정, 리더나 `decision` 노트, 같은 head 재제출이 있어도 빠진 채이고, 옆의 판정 없는 카드는 보여요 · verify: `cargo test --locked --offline --test card card_wip_limit::audit_head_ac2_`
- AC-3 (REQ-3): 판정이 있어도 노트 없는 재제출, 읽히지 않는 최신 노트, `card.note` 이벤트 없는 노트, 노트 없음, reclaim 뒤에는 보여요 · verify: `cargo test --locked --offline --test card card_wip_limit::audit_head_ac3_`
- AC-4 (REQ-4): 읽히지 않는 판정 행이 있으면 그 행 하나뿐이든 지금 head 의 PASS 옆이든 보여요 · verify: `cargo test --locked --offline --test card card_wip_limit::audit_head_ac4_`
- AC-5 (REQ-5): 다시 낸 카드는 보유 세션에게는 안 보이고 다른 세션에게는 보여요 · verify: `cargo test --locked --offline --test card card_wip_limit::audit_head_ac5_`
- AC-6 (REQ-6): 두 호출자가 한 함수를 써요. 그 함수의 판정 행 수 비교를 지우면 목록 테스트와 상한 테스트가 모두 실패해요 · verify: `cargo test --locked --offline --test card card_wip_limit::pass_review_ac4_`
- AC-7 (REQ-5, REQ-7): 기존 테스트는 고치지 않고 통과해요(읽을 수 없는 레지스트리의 빈 목록은 기존 `registry_error_next_…` 테스트가 붙잡아요) · verify: `cargo test --locked --offline --test card card_wip_limit::`
- AC-8: 가벼운 검사 · verify: `git diff --check`; `cargo fmt --all -- --check`; `cargo clippy --locked --offline --all-targets -- -D warnings`

## Scope

- write: `src/card/wip_limit.rs`, `tests/card/card_wip_limit.rs`, `.moai/specs/SPEC-20261126-audit-needed-current-head/**`
- read-only: `src/card/work.rs`, `src/bundle.rs`, `src/card/ready_receipt.rs`, `src/ledger/workflow.rs`
- forbidden: `src/ledger/`, `src/card/land/`, `tests/common/`, 실제 프로젝트의 원장과 레지스트리

## Plan

1. RED: `tests/card/card_wip_limit.rs` 에 `audit_head_ac1_` 부터 `audit_head_ac5_` 까지의 테스트를 덧붙여요. `ac2` 는 바뀌지 않는 동작을 붙잡는 테스트라 지금도 통과하고, 나머지는 실패해야 해요.
2. GREEN: `src/card/wip_limit.rs` 에서 `passed_head` 의 현재 head·판정 행 읽기를 비공개 함수 하나로 옮기고(동작 불변), `audit_needed` 가 SQL 의 판정 존재 조건 대신 그 함수를 쓰게 해요.
3. 가드 뮤턴트를 하나씩: 판정의 head 일치, PASS 만 숨김, 최신 판정만 비교, 함수의 행 수 비교, 확인 불가를 판정됨으로 취급, 보유자 조건.
4. Result 에 명령과 종료 코드를 적고 `moai check affected --card` 영수증을 한 번 받아요.

## Risks

- R1: 착지자의 합친 head 에만 판정이 있는 카드는 목록에 남아 두 번 감사될 수 있어요(D-2).
- R2: FAIL 받은 head 를 그대로 다시 내면 숨겨진 채예요(예: 환경 때문에 난 FAIL). 리더가 손으로 배정해요.
- R3: 지금 head 에 FAIL 과 PASS 가 함께 있으면 감사 목록에는 없고 상한에서는 세어져요.
- R4: reclaim 된 카드, 새 노트 없이 다시 보낸 카드, 가장 새 표시 노트가 읽히지 않는 카드(`READY (slice 1) …` 같은 첫 줄)는 PASS 가 있어도 다시 보여요.
- R5: `review` 카드마다 현재 head 를 읽어요. 임대가 살아 있는 카드는 두 번 읽어요. `card_note` 에는 `(card_id, ref)` 부분 색인만 있어 읽기 비용을 재지 않았어요. git 프로세스는 늘지 않아요.
- R6: 카드 하나의 읽기가 SQL 오류를 내면 `card next` 전체가 실패해요. `in_work` 와 같은 종류예요.
- R7: `card orchestrate` 는 `AUDIT` 노트도 판정으로 읽어요. 노트만 있는 판정은 `card next` 목록에 전처럼 남아요.
- R8: 감사 중인지는 보지 않아서, 새로 보이는 카드에 이미 감사자가 있을 수 있어요.
- 기준선(읽기 전용 sqlite 근사, 2026-11-26): 새 규칙이면 판정이 있는 `review` 카드 22장 가운데 2장이 새로 보일 것으로 어림했어요. 노트의 `parse_ready` 와 읽히지 않는 행 검사는 건너뛴 근사라 코드의 증거는 아니에요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 지금 head 의 판정이 하나라도 있으면 숨겨요 | 노트 9971 D-B | PASS 만 숨겨요 |
| D-2 | 정확히 그 head 의 판정만 세요 | 9971 의 "정확히", 상한과 같아요 | git 으로 합친 head 를 따라가요 |
| D-3 | `card.verdict` 이벤트만 판정이에요 | 9971 의 "판정 행", 지금과 같아요 | `AUDIT` 노트도 읽어요 |
| D-4 | 두 호출자가 비공개 함수 하나를 써요 | 카드 기준 3 | 복사본을 둬요 |
| D-5 | tier 1 이라 카드 본문과 기준이 설계이고 이 SPEC 은 승인을 기다리지 않는 초안이에요 | 삭제·권한·관문·HOME 변경이 아니고, 감사 목록을 보여 주는 쪽만 바꿔요. 바꾸는 파일은 `audit.md` 의 tier 2 접두 목록에 없어요. 제출 head 의 `moai bundle check` 가 floor 2 를 내면 리더 승인을 먼저 받아요 | 리더의 위임 승인을 받아요 |

## Result

구현했어요(브랜치 `spec/audit-needed-current-head-20261126`, 기반 `b3048a821a`). 이 SPEC 은 D-5 대로 승인 기록이 없는 초안이에요. 커밋은 초안 `6920afb801`, red 테스트 `5084ece0ea`(`tests/card/card_wip_limit.rs`, 318줄 추가, 삭제 0), 코드 `38fa79810b`(`src/card/wip_limit.rs` 한 파일)예요. 독립 감사는 이 레인이 하지 않아요.

- 바뀐 것: `audit_needed` 의 SQL 에서 "`card.verdict` 이벤트가 하나라도 있으면 빼는" 조건을 없앴어요. 이제 지금 head 를 확인할 수 있고, 카드의 `card.verdict` 행이 모두 읽히고, 어느 검토자의 판정이든(PASS·FAIL) 하나가 그 head 를 정확히 가리킬 때만 감사 목록에서 빼요. 현재 generation 의 예약을 묻는 세션이 가진 카드를 빼는 조건은 그대로예요. head 와 판정 행 읽기는 비공개 함수 `head_and_verdicts` 로 옮겨 `passed_head`(작업 수 상한)와 `audit_needed` 가 같이 써요. `passed_head` 의 동작은 바꾸지 않았고, `src/card/work.rs` 는 고치지 않았어요.
- 실행 방법: 아래 테스트 실행은 `moai heavy-slot-acquire --name lamech --max 2 -- cargo nextest run --locked --offline --no-fail-fast --test card …` 이고 `CARGO_INCREMENTAL=0` 으로 돌렸어요. SPEC 의 `verify:` 줄은 적힌 `cargo test` 형식 그대로는 돌리지 않았고, 같은 테스트를 nextest 로 돌렸어요. 로그 파일은 레인 워크트리의 `target/` 아래 로그 폴더에 있어서 파일 이름만 적어요.
- red, `5084ece0ea` 의 테스트를 담은 트리: 필터 `card_wip_limit::` 35개 실행, 30 통과, 5 실패, exit 100(`red.log`). 새 테스트 6개 가운데 `audit_head_ac1_` 두 개와 `audit_head_ac3_`, `ac4_`, `ac5_` 가 실패했고, 바뀌지 않는 동작을 붙잡는 `audit_head_ac2_` 는 통과했어요.
- green: 필터 `card_wip_limit::` 또는 `card_next`, 40개 실행, 40 통과, exit 0(`green.log`).
- 호출자: 필터 `orchestrate_cli`, `card_submit_cli`, `card_work_cli`, 68개 실행, 68 통과, exit 0(`callers.log`).
- 뮤턴트에서 되돌린 뒤: 필터 `card_wip_limit::` 또는 `card_read_speed::`, 39개 실행, 39 통과(`restored.log`). `mutants.status` 에 `restored exit=0`, 원본 사본과의 비교 `cmp exit=0` 이 있어요.
- clippy: `cargo clippy --locked --offline --all-targets -- -D warnings` exit 0(`clippy.log`).
- 위의 테스트, 뮤턴트, clippy 실행 뒤에 `audit_needed` 의 문서 주석 한 줄에서 비공개 항목 링크를 없앴어요. 실행에 쓴 원본 사본과 커밋한 `src/card/wip_limit.rs` 의 차이는 그 한 줄이에요. 커밋한 바이트에서는 `cargo fmt --all -- --check` 와 `git diff --check b3048a821a 38fa79810b` 가 exit 0 이었고, nextest 와 clippy 를 그 바이트에서 검사하는 것은 READY 영수증 몫이에요.
- 가드 뮤턴트(하나씩 적용, 각 `--test card` 필터 `card_wip_limit::`, `mutant-1.log` 부터 `mutant-6.log`, 종료 코드는 `mutants.status`):
  - guard-mutant: audit_head_ac1_at_the_limit_next_offers_the_resubmitted_card, audit_head_ac1_a_new_head_after_a_verdict_is_listed, audit_head_ac5_the_holder_is_not_asked_to_audit_its_own_card | 판정의 head 일치(판정이 비어 있지 않은지 검사로 바꿈) | exit 100
  - guard-mutant: audit_head_ac1_a_new_head_after_a_verdict_is_listed, audit_head_ac2_a_verdict_on_the_current_head_hides_the_card | FAIL 판정도 숨기는 동작(PASS 만 숨김) | exit 100
  - guard-mutant: audit_head_ac2_a_verdict_on_the_current_head_hides_the_card | 모든 판정의 head 비교(최신 판정만 비교) | exit 100
  - guard-mutant: audit_head_ac4_a_card_with_an_unreadable_verdict_row_is_listed, pass_review_ac4_a_verdict_row_that_cannot_be_read_makes_the_card_count | 판정 행 수 비교 | exit 100
  - guard-mutant: ac4_next_names_the_cards_left_out_and_keeps_the_audit_list, audit_head_ac3_a_card_whose_current_head_cannot_be_confirmed_is_listed, next_at_the_limit_names_the_cards_that_wait_for_an_audit, audit_head_ac4_a_card_with_an_unreadable_verdict_row_is_listed | 확인 불가를 목록에 두는 처리(판정됨으로 취급) | exit 100
  - guard-mutant: audit_head_ac5_the_holder_is_not_asked_to_audit_its_own_card, next_at_the_limit_names_the_cards_that_wait_for_an_audit | 보유자 조건 | exit 100
- SPEC 검사: `moai spec check SPEC-20261126-audit-needed-current-head` exit 0, 오류 0, 경고 0(설치된 `moai 4.2648.55-dev`).

- AC-1: met — `audit_head_ac1_a_new_head_after_a_verdict_is_listed` 와 `audit_head_ac1_at_the_limit_next_offers_the_resubmitted_card` 가 red 에서 실패하고 green 에서 통과해요.
- AC-2: met — `audit_head_ac2_a_verdict_on_the_current_head_hides_the_card` 는 Plan 대로 red 에서도 통과했고 green 에서 통과해요. "PASS 만 숨김" 과 "최신 판정만 비교" 뮤턴트에서는 실패해요.
- AC-3: met — `audit_head_ac3_a_card_whose_current_head_cannot_be_confirmed_is_listed` 가 red 에서 실패하고 green 에서 통과해요.
- AC-4: met — `audit_head_ac4_a_card_with_an_unreadable_verdict_row_is_listed`, 같은 red/green.
- AC-5: met — `audit_head_ac5_the_holder_is_not_asked_to_audit_its_own_card`, 같은 red/green.
- AC-6: met — `passed_head` 와 `audit_needed` 가 `head_and_verdicts` 를 불러요. 그 함수의 행 수 비교를 지운 뮤턴트에서 `audit_head_ac4_…` 와 `pass_review_ac4_…` 가 모두 실패했고, 두 테스트는 green 에서 통과해요.
- AC-7: met — 기존 테스트는 고치지 않았고(추가만 318줄, 삭제 0) `registry_error_next_names_no_card_not_even_one_that_waits_for_an_audit` 를 포함해 green 실행에서 통과해요.
- AC-8: met — 위의 `git diff --check`, `cargo fmt --all -- --check`(커밋한 바이트), clippy(주석 한 줄을 고치기 전 바이트)가 exit 0 이에요.

알려 둘 것:

- `docs/later.md` 의 작업 상한 셈 줄은 아직 `audit_needed` 가 다시 제출한 카드를 숨긴다고 적어요.
- SPEC-20261126-wip-limit-pass-review 의 Risks 줄과 Result 의 "알려 둘 것" 도 판정이 하나라도 있는 카드가 감사 목록에서 빠진다는 이전 동작을 적고 있어요.
- 위의 두 문서는 이 작업의 쓰기 범위 밖이라 고치지 않았어요.

not_run:

- `check affected --card` 영수증은 이 Result 가 든 head 를 이름 짓는 READY 노트가 담아요.
- 전체 `cargo test`, 실제 프로젝트 원장에 대한 이 빌드의 실행, Windows 와 Linux, 독립 감사.
