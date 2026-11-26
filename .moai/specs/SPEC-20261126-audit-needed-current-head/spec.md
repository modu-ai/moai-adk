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

`card next` 의 `audit_needed` 는 `card.verdict` 가 하나라도 있는 `review` 카드를 모두 숨겨요. 판정 뒤 수리해서 새 head 로 다시 낸 카드가 감사 목록에 없어서 리더가 손으로 감사자를 찾아야 했어요. 지금 제출한 head 를 정확히 가리키는 판정이 있을 때만 숨기고, 확인할 수 없으면 보여요.

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

초안이에요. 구현 뒤 변경한 파일, 실행한 검사, AC 별 met 를 적어요.
