---
id: SPEC-20261125-bundle-reads-repaired-note
title: "moai bundle reads the newest hand-back note of a card, a REPAIRED note included, and names the note it read"
status: draft
revision: 1
created: 2026-11-25
parent: null
---

# moai bundle reads the newest hand-back note of a card, a REPAIRED note included, and names the note it read

## Interview

- Q1: 수리 뒤의 노트를 `moai bundle check` 가 읽지 못하는 문제를 어떤 규칙으로 고치나요? -> A: "bundle check 가 카드의 가장 새 제출 노트(READY 또는 REPAIRED)의 head 와 guard-mutant 줄을 읽고 어느 노트를 읽었는지 답에 적어요. 리더가 쓴 노트와 AUDIT 노트는 제출 노트가 아니에요. 게이트를 느슨하게 하지 않아요: 새 노트에 guard-mutant 줄이 없으면 옛 노트의 줄로 채우지 말고 없다고 경고해요."(리더 배차 2026-11-24T20:47Z 와 카드 본문, 사용자 위임)

## Intent

같은 문제가 2026-11-24 밤에 감사 두 건에서 나왔어요: 레인이 감사 FAIL 뒤에 `REPAIRED <card> head=<40hex>` 노트로 고친 head 와 고친 `guard-mutant:` 줄을 냈는데도 `moai bundle check <card>` 는 옛 `LANE-READY-INTEGRATION` 노트의 head 와 그 노트의 줄을 읽고 `mutant_line_unreadable`, `result_missing` 경고를 그대로 냈어요. 원인은 `src/bundle.rs` 의 `newest_marker` 와 `ready_body_naming` 이 첫 줄이 `LANE-READY-INTEGRATION` 으로 시작하는 노트만 제출 노트로 보는 것이에요. 반면 `card note` 의 영수증 게이트(`src/card/ready_receipt.rs` 의 `marker_of_line`)는 `LANE-READY-INTEGRATION`, `REPAIRED`, `READY ` 셋을 제출 노트로 봐서, 게이트가 받아 준 REPAIRED 노트를 번들 도구가 못 읽어요. 감사자는 수리 head 를 이 도구로 확인할 수 없고 옛 head 의 경고를 보게 돼요.

이 SPEC 은 번들 도구가 게이트와 같은 판정(`marker_of_line`)으로 제출 노트를 고르고, 가장 새 제출 노트의 head 와 줄을 읽고, 어느 노트를 읽었는지 답에 적게 해요. 읽는 사람: 레인, 감사자, 리더.

## Non-goals

- 게이트(`card note`)와 `card inbox`·상태 줄의 노트 판정은 바꾸지 않아요. 확인만 했어요: `card inbox` 의 `ready_for_integration` 가족은 `LANE-READY-INTEGRATION` 첫 줄만 세고 `REPAIRED` 는 세지 않아요. 이 구분은 `inbox_family_follows_marker_of_line` 테스트가 고정한 것이라 이 SPEC 이 바꾸지 않고 리더에게 알려요.
- guard-mutant 줄의 문법과 경고 코드(`mutant_line_missing`, `mutant_line_unreadable`, `mutant_survived`)는 바꾸지 않아요. 새 노트에 줄이 없으면 옛 줄로 채우지 않고 `mutant_line_missing` 으로 경고하는 것은 지금 동작이고 테스트로 고정해요.
- 첫 줄이 `LANE-READY-INTEGRATION` 이고 둘째 줄에 `REPAIRED` 를 쓰는 노트는 새 모양이 아니에요: 그 노트는 첫 줄의 head 로 읽혀요(게이트도 첫 줄만 봐요).
- 게이트의 영수증 판정, 감사 판정(`AUDIT` 노트 읽기), 병합 검사, 병합 몸통 만들기의 규칙은 바꾸지 않아요. 이들은 head 만 받아서 새 head 를 자동으로 쓰게 돼요.

## Requirements

<!-- This SPEC adds no blocking check. -->
- REQ-1: When `bundle` reads the hand-back note of a card, the command shall take the newest note whose first line starts with `LANE-READY-INTEGRATION`, `REPAIRED` or `READY `, the three markers that `card note` accepts, and read the head from that line.
- REQ-2: When `bundle check` or `bundle ready` names the head of a card, the command shall also name the note it read, by its number and its marker. · failure: none — the note is a report.
- REQ-3: When `bundle check` reads the `guard-mutant:` lines of a head, the command shall read them from the newest hand-back note that names the head alone, whatever the lines of an older note say. · failure: none — a warning is listed, never a block.
- REQ-4 (REQ-1): IF a note is written by the leader or is a decision note, THEN the command shall not read it as a hand-back note whatever its first line says.

## Acceptance

- AC-1 (REQ-1, REQ-2): READY 노트 뒤에 head 가 다른 `REPAIRED` 노트가 있는 카드에서 `bundle check <card>` 는 REPAIRED 노트의 head 를 읽고 `heads[0].read_note` 가 `{"note_id": <REPAIRED 노트 번호>, "marker": "REPAIRED"}` 예요. 첫 READY 노트만 있는 카드는 `{"note_id": <번호>, "marker": "LANE-READY-INTEGRATION"}` 이고, `bundle ready` 의 각 head 도 `marker` 를 말해요 · verify: `cargo test --locked --offline --test bundle_check_cli a_repaired_note_after_a_ready_note_is_the_note_that_is_read`
- AC-2 (REQ-3): 위 카드의 REPAIRED 노트에 `guard-mutant:` 줄이 있으면 경고가 없고, 없으면 옛 READY 노트에 좋은 줄이 있어도 `mutant_line_missing` 경고가 나와요. 옛 head 를 `<card>=<40hex>` 로 직접 지정하면 옛 노트의 줄이 읽혀요 · verify: `cargo test --locked --offline --test bundle_check_cli a_repaired_note_without_mutant_lines_is_not_filled_from_the_old_note`
- AC-3 (REQ-4): 리더가 쓴 결정 노트의 첫 줄이 `REPAIRED <card> head=<40hex>` 이거나 `LANE-READY-INTEGRATION ...` 이어도 제출 노트로 읽히지 않아요: 레인의 READY 노트가 더 오래돼도 그 head 가 읽혀요 · verify: `cargo test --locked --offline --test bundle_check_cli a_note_written_by_the_leader_is_not_a_hand_back_note`
- AC-4 (REQ-1): `REPAIRED` 노트의 모양이 틀리면(카드 이름이 다르거나 head 가 40자리 소문자 16진이 아님) 가장 새 제출 노트가 읽히지 않는 것이고 옛 노트로 되돌아가지 않아요: `no_ready_line` 으로 알려요 · verify: `cargo test --locked --offline --test bundle_check_cli a_newest_hand_back_note_that_does_not_read_is_not_replaced_by_an_older_one`
- AC-5: 이웃 테스트가 그대로 통과해요 · verify: `cargo test --locked --offline --test bundle_check_cli --test bundle_ready_cli --test bundle_body_cli --test bundle_build_cli --test bundle_verify_cli` and `cargo test --locked --offline --lib bundle`

## Scope

- write: `.moai/specs/SPEC-20261125-bundle-reads-repaired-note/**`, `src/bundle.rs`(제출 노트 고르기·읽기, `check`·`ready` 의 답), `src/ops/bundle.rs`(`Ready` 도움말 글), `tests/bundle_check_cli.rs`(테스트 더하기)
- read-only: `src/card/ready_receipt.rs`(`marker_of_line` 를 그대로 써요), `src/card/inbox.rs`, `src/bundle/build.rs`, `src/bundle/verify.rs`, `src/bundle/mutant_line.rs`, `tests/bundle_ready_cli.rs`, `tests/bundle_body_cli.rs`, `tests/bundle_build_cli.rs`, `tests/bundle_verify_cli.rs`
- forbidden: 게이트와 `card inbox`·상태 줄의 판정 코드, `skills/**`, `CLAUDE.md`, 이 저장소의 `.moai/project.yaml`, 실제 HOME

## Plan

1. SPEC 초안을 쓰고 커밋해요. 승인은 감사 때 리더가 노트로 해요; 레인은 승인을 기록하지 않고 초안 위에서 RED 부터 구현해요.
2. RED: `tests/bundle_check_cli.rs` 에 AC-1 ~ AC-4 의 테스트를 써서 현재 팁에서 옳은 이유(옛 READY 노트의 head 를 읽음, `read_note` 키 없음, 리더 노트 판정 없음)로 실패하는 것을 봐요. 이미 맞는 동작을 지키는 단정(AC-2 의 직접 지정, AC-4 의 옛 노트 되돌아감 없음)은 통과하는 가드로 두어요.
3. GREEN: `Note` 가 `kind`·`author` 를 가지고, `newest_marker`·`ready_body_naming`·`parse_ready` 가 게이트의 `marker_of_line` 로 세 표지를 읽고 리더·결정 노트를 거르게 하고, `Ready` 가 읽은 표지를 들고, `check`·`ready` 의 답이 `read_note`·`marker` 를 내요. 빌드·검증·몸통은 `newest_marker` 를 그대로 불러서 고쳐져요.
4. 확인: 새 테스트, AC-5 의 이웃 테스트, `cargo fmt --check`, `cargo clippy --locked --offline --all-targets -- -D warnings`(대상 한정), 새 가드 테스트마다 가드 뮤턴트 한 줄(`guard-mutant: <테스트> | <지운 코드> | exit <n>`), 위험 경로는 읽기 전용 도구라 해당 없음을 READY 에 적어요, `moai check affected` 를 READY 에서 한 번.
5. SYNC: Result 에 AC 별 met·not_run 을 적어요. 감사는 이 변경을 쓰지 않은 세션이 맡아요.

## Risks

- `bundle ready` 가 REPAIRED 노트의 head 도 목록에 올려요: 감사 FAIL 뒤 수리된 카드가 `ready` 목록에서 수리 head 로 보이게 되는 것이 의도이고, 옛 head 는 더 보이지 않아요. 같은 head 가 두 노트에 있으면 가장 새 노트가 이겨요.
- `READY ` 표지(게이트가 받는 짧은 말)도 번들이 읽게 돼요. 게이트가 받아 준 노트를 도구가 못 읽는 일이 다시 없게 하려는 것이고, 모양이 틀린 줄은 지금처럼 읽히지 않는 노트로 알려져요.
- 리더·결정 노트를 거르는 기준이 노트의 종류(`decision`)와 저자(`leader`)예요. 레인이 READY 노트를 `--kind decision` 으로 쓰면 읽히지 않고 `no_ready_line` 이 나와요: 안전한 쪽으로 닫히고 이유가 답에 있어요.
- 답에 키가 더해져요(`heads[].read_note`, `ready` 의 `heads[].marker`). 키를 더하기만 하므로 키를 읽는 쪽은 그대로 동작하고, 객체 통째 비교하는 테스트가 있으면 이웃 테스트가 알려요(Plan 4).
- `card inbox` 가 `REPAIRED` 를 요청으로 세지 않는 것은 이 SPEC 밖이에요(Non-goals); 리더가 그 구분을 바꿀지 정해요.

## Decisions

승인 전 제안이에요. 사용자나 위임받은 리더가 고르면 revision 을 올려 `approval` 을 기록해요(레인은 승인을 기록하지 않아요).

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 번들 도구가 제출 노트를 게이트의 `marker_of_line` 로 판정한다(세 표지). 대안: 번들 안에서 `LANE-READY-INTEGRATION` 과 `REPAIRED` 둘만 직접 비교 | 게이트가 받아 준 노트를 도구가 못 읽는 일은 판정이 둘이라서 생겼어요; 판정을 하나로 두면 새 표지가 한 곳에서 늘어요. 기본: 하나 | 번들 안의 비교로 되돌려요 |
| D-2 | 가장 새 제출 노트가 읽히지 않는 모양이면 옛 노트로 되돌아가지 않고 `no_ready_line` 으로 알린다(지금 동작 그대로) | 카드 본문: 게이트를 느슨하게 하지 않아요; 코드 주석도 "옛 head 는 낡았다"고 해요. 기본: 이대로 | — |
| D-3 | 리더·결정 노트는 첫 줄과 상관없이 제출 노트가 아니다(저자 `leader` 또는 종류 `decision`) | 카드 본문 "리더가 쓴 노트와 AUDIT 노트는 제출 노트가 아니에요"; AUDIT 노트는 첫 줄이 표지가 아니라서 이미 걸러지고, 리더가 표지로 시작하는 줄을 인용해 쓸 때를 막아요. 기본: 이대로 | 거르는 줄을 지워요 |
| D-4 | 어느 노트를 읽었는지는 답의 새 키(`read_note`, `marker`)로 말한다. 대안: 기존 키의 뜻을 넓힘 | 키를 더하기만 하면 읽는 쪽이 깨지지 않아요(`note_id` 는 `ready` 에 이미 있어요). 기본: 새 키 | 키를 지워요 |

## Result

- RED(커밋 beb00f9ed4, `target/c486/red1.log`): 새 테스트 4개가 팁에서 옳은 이유로 실패했어요 — REPAIRED 를 건너뛰고 옛 head 를 읽음, 리더가 인용한 READY 줄을 제출 노트로 읽음, 읽히지 않는 새 노트 대신 옛 노트로 되돌아감.
- GREEN(수리 커밋 65b4f8db13, `target/c486/green1.log`): 새 테스트 4개와 `bundle_check_cli`·`bundle_ready_cli`·`bundle_body_cli`·`bundle_build_cli`·`bundle_verify_cli`·lib `bundle` 115개가 모두 통과했어요(nextest, 슬롯 `--max 3`). 단위 테스트 `the_grammar_reads_every_marker_the_gate_takes` 가 세 표지의 읽기를 고정해요.
- 가드 뮤턴트 10개 중 10개가 컴파일되고 잡혔어요(한 슬롯 호출, 출력 `target/c486/mutants.out`, 개별 로그 `target/c486/mutant-*.log`, 스크립트 `target/c486/mutants.py`). 편집은 정확히 한 번씩, 파일은 `finally` 에서 되돌렸고 `git status --short` 가 비었어요.
- 관찰: `card inbox` 는 `REPAIRED` 를 `ready_for_integration` 으로 세지 않아요(`inbox_family_follows_marker_of_line` 테스트가 고정). 이 SPEC 은 바꾸지 않았어요; 바꿀지는 리더가 정해요.
- AC-1 met · AC-2 met · AC-3 met · AC-4 met · AC-5 met (위 GREEN 실행). `moai check affected` 영수증은 READY 노트에 있어요.
- Gaps: 전체 `cargo test` 는 not_run(리더 몫). 이미 열린 READY 노트의 head 가 `REPAIRED` 로 대체된 카드의 `bundle ready` 목록은 실제 원장에서 보지 않았어요(임시 원장 테스트만).
