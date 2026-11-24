---
id: SPEC-20261124-followup-per-spec
title: "A SPEC has one open follow-up card; later non-blocking findings and recurrences add to an existing card"
status: approved
revision: 1
created: 2026-11-24
approval:
  by: leader
  at: 2026-11-24
  basis: "리더 결정 노트 8040 (2026-11-24T00:18:40Z, 사용자 위임 2026-11-23T13:57Z, 노트 8036 에 답해요: 커밋 699181646e 의 판 1 승인, D-1~D-8 기본값 그대로, CLAUDE.md 는 재발 줄과 Flow 7 마지막 문장 두 곳만 tier-2 감사); 리더가 노트에 적었고 레인은 그 노트를 인용만 해요"
  digest: "sha256:04db71600eb592938ebb639ba48644f4f90fee817fa5401d2407b6373cec3db2"
---

# A SPEC has one open follow-up card; later non-blocking findings and recurrences add to an existing card

## Intent

사용자가 2026-11-24 카드 조사 뒤 고른 개선이에요(카드 C-443). 카드를 만드는 규칙이 개수 제한 없이 카드를 늘리고, `card add` 는 같은 SPEC 의 카드가 이미 있는지 확인하지 않아요.

읽어서 확인한 사실(코드와 글, 2026-11-24):

- 카드를 만드는 규칙이 세 곳에 있어요: `CLAUDE.md` 의 "Recurring problems change the source" 줄(같은 종류의 문제가 두 번이면 그 라운드에 카드를 열어요), `docs/leader-rules.md` 의 Refill (3)(감사의 non-blocking 노트를 카드로 바꿔요), `skills/moai/references/flows/next-flow.md` 의 Landing 마지막 줄과 `CLAUDE.md` 의 Flow 7(실패한 원격 실행은 repair card 가 돼요). 셋 다 "한 건당 카드 한 장" 이고 한도가 없어요.
- `card add` 는 `src/card/command.rs` 의 `add` 한 곳이고 같은 제목이나 같은 SPEC 의 카드를 찾지 않아요. 카드와 SPEC 의 연결은 `card.spec` 이벤트로만 남아요(`src/card/spec_link.rs` 머리말; 새 열이 없어요). 지금은 `card add` 가 연결을 못 써서 `card spec C-n <SPEC-ID>` 를 따로 불러야 해요.
- `card criterion-add` 는 카드 상태를 따지지 않고 기준을 덧붙여요(`done` 만 거절, `src/ledger/workflow.rs`). 그래서 한 SPEC 의 후속 일을 한 카드의 기준으로 모을 수 있어요.

그래서 두 가지를 바꿔요. (1) `card add --spec` 이 연결을 한 번에 쓰고, 같은 SPEC 에 열린 카드가 이미 있으면 거절해요. 후속 카드는 SPEC 마다 열린 것 하나만 허용하고(`--follow-up`), 갈라 내는 카드는 `--split-of` 로 이유를 남겨요. (2) 세 규칙 글을 "기준 하나를 후속 카드에 더해요" 와 "같은 원인의 열린 카드에 노트를 더해요" 로 고쳐요.

## Interview

- Q1: 무엇을 막나요? → A: 같은 SPEC 에 열린 카드가 둘 이상 생기는 일을 `card add --spec` 에서 막아요. 사용자가 2026-11-24 카드 조사 뒤 고른 개선이에요(카드 본문, 리더 노트 8040).
- Q2: 후속 일은 어디로 가나요? → A: SPEC 의 열린 후속 카드에 기준 하나(`card criterion-add`)로 모이고, 후속 카드가 없을 때만 `--follow-up` 으로 새로 열어요(D-6).
- Q3: 나중 연결(`card spec`)은요? → A: 건드리지 않아요(D-3, 리더 노트 8040 의 D-1~D-8 기본값 승인).
- Q4: 기존 장부의 카드는요? → A: 합치거나 지우지 않아요(Non-goals).

## Non-goals

- `card spec C-n <SPEC-ID>` 로 나중에 연결하는 길의 거절. 이미 카드가 여럿 붙은 SPEC 이 많고 그 길을 막으면 기존 작업이 깨져요(D-3).
- 이미 있는 카드를 합치거나 지우는 일, 기존 장부 이행.
- 후속 카드가 가진 기준을 SPEC 파일로 옮기는 일, `card show` 가 관계를 보여주는 일(`card add` 의 답과 거절 문장이 관계를 말해요).
- `moai card land`, `moai card next` 의 동작 바꾸기.
- 카드 한 장에 기준이 몇 개까지인지 한도 두기.

## Requirements

<!-- This SPEC adds a check that blocks. It does not start in observe: `--spec` is a new optional flag, so no existing call can be blocked, and the user chose the refusal in the card text (D-8). -->
- REQ-1: WHEN `card add` is given `--spec <SPEC-ID>`, the command shall write the card, its criteria, the `card.add` event and one `card.spec` event naming the SPEC in one transaction.
- REQ-2 (REQ-1, REQ-6): IF the SPEC id is not of the form `SPEC-<yyyymmdd>-<slug>`, the SPEC is not found under `.moai/specs`, the SPEC has an error finding, or `--split-of` names a card that does not exist or names another SPEC or none, THEN the `card add` command shall refuse and write nothing.
- REQ-3 (REQ-1): IF `card add --spec` is given neither `--follow-up` nor `--split-of` and a card that is not `done` already names the SPEC, THEN the command shall refuse, write nothing and name each such card with the three ways out (`--follow-up`, `--split-of <card>`, `card criterion-add <card> --ac`).
- REQ-4 (REQ-1): WHEN `card add --spec --follow-up` is given and no card that is not `done` holds the follow-up relation on that SPEC, the command shall add the card and write `"relation":"follow-up"` in the payload of its `card.spec` event.
- REQ-5 (REQ-4): IF a card that is not `done` already holds the follow-up relation on the SPEC, THEN the `card add` command shall refuse `--follow-up`, write nothing and name that card and `moai card criterion-add <card> --ac <text>`.
- REQ-6 (REQ-1): WHEN `card add --spec --split-of <card>` is given and `<card>` names that SPEC, the command shall add the card and write `"relation":"split-of"` and `"of":"<card>"` in the payload of its `card.spec` event.
- REQ-7: IF `--follow-up` or `--split-of` is given without `--spec`, or both are given together, THEN the `card add` command shall refuse and write nothing.
- REQ-8: WHEN `card add` is given none of `--spec`, `--follow-up` and `--split-of`, the command shall add the card exactly as it did before this SPEC. · failure: none — such a call takes the path it took before
- REQ-9: The Refill step 3 of `docs/leader-rules.md` shall say that an audit's non-blocking finding becomes a criterion of the SPEC's open follow-up card through `moai card criterion-add` and that a follow-up card is opened with `card add --spec <SPEC-ID> --follow-up` only when the SPEC has none open.
- REQ-10: The recurrence rule in the Standing principles of `CLAUDE.md` shall say that a recurrence whose cause already has an open card adds a note to that card and opens no new one.
- REQ-11: The failed-remote-run sentence in `skills/moai/references/flows/next-flow.md` (Landing) and in Flow 7 of `CLAUDE.md` shall say that the failure becomes a criterion of the open follow-up card of the SPEC the failed head names beside the unchanged words `Remote checks never block a landing`.
- REQ-12: The lane rules in `docs/lane-rules.md` shall say that a lane writes a non-blocking finding in its note and never adds a card for it.

## Acceptance

- AC-1 (REQ-1, REQ-2): 임시 HOME 과 SPEC 이 든 임시 저장소에서 `card add --spec <id>` 가 카드를 만들고 `card.spec` 이벤트가 그 SPEC 을 가리켜요(`card show` 의 `spec` 이 같아요). 없는 SPEC, 형식이 틀린 id, error 가 있는 SPEC 은 거절되고 장부의 카드 수가 그대로예요. · verify: `cargo nextest run --offline --locked --test card_add_spec_cli -E 'test(add_with_spec_links_in_one_write) | test(a_missing_or_broken_spec_is_refused)'`
- AC-2 (REQ-3): 같은 SPEC 에 `done` 이 아닌 카드가 있을 때 `--spec` 만 준 두 번째 `card add` 가 거절되고, 문장이 그 카드 이름과 `--follow-up`, `--split-of`, `card criterion-add` 를 적고 카드 수가 그대로예요. 그 카드가 `done` 이면 받아요. · verify: `cargo nextest run --offline --locked --test card_add_spec_cli -E 'test(a_second_open_card_on_one_spec_is_refused) | test(a_done_card_does_not_count)'`
- AC-3 (REQ-4, REQ-5): `--follow-up` 첫 번째는 받고 이벤트 payload 에 `relation` 이 `follow-up` 이에요. 열린 후속 카드가 있는 SPEC 의 두 번째 `--follow-up` 은 거절되고 그 카드 이름과 `card criterion-add` 를 적어요. 후속 카드가 `done` 이 되면 새 `--follow-up` 을 받아요. · verify: `cargo nextest run --offline --locked --test card_add_spec_cli -E 'test(one_open_follow_up_per_spec) | test(a_closed_follow_up_makes_room)'`
- AC-4 (REQ-6, REQ-2): `--split-of C-1` 은 C-1 이 그 SPEC 을 가리킬 때 받고 payload 가 `split-of` 와 `C-1` 을 적어요. C-1 이 없거나 다른 SPEC 이거나 SPEC 이 없으면 거절돼요. · verify: `cargo nextest run --offline --locked --test card_add_spec_cli -E 'test(split_of_names_its_origin) | test(split_of_needs_the_same_spec)'`
- AC-5 (REQ-7): `--follow-up` 이나 `--split-of` 를 `--spec` 없이 주거나 둘을 함께 주면 거절되고 카드 수가 그대로예요. · verify: `cargo nextest run --offline --locked --test card_add_spec_cli -E 'test(relation_flags_need_a_spec_and_exclude_each_other)'`
- AC-6 (REQ-8): `--spec` 없이 부른 `card add` 는 이전과 같이 받아요(기존 카드 테스트가 그대로 통과해요). · verify: `cargo nextest run --offline --locked --test card_cli --test card_dispatch_cli`
- AC-7 (REQ-9, REQ-12): `docs/leader-rules.md` 의 Refill 3번이 후속 카드 기준 더하기와 `--follow-up` 규칙을 적고, `docs/lane-rules.md` 가 레인이 카드를 더하지 않는다고 적어요. · verify: `cargo nextest run --offline --locked --test followup_per_spec_rules -E 'test(leader_and_lane_rules_name_the_follow_up_card)'`
- AC-8 (REQ-10, REQ-11): `CLAUDE.md` 재발 줄과 출처 행, 원격 실패 문장 두 곳이 새 글을 담고 기존 글 고정 테스트가 그대로 통과해요. · verify: `cargo nextest run --offline --locked --test followup_per_spec_rules --test recurring_problems_rule --test repo_claude_md --test next_flow_skill_scan`

## Scope

- write: `src/card/command.rs`, `src/card/spec_link.rs`, `src/ops/card.rs`, `tests/card_add_spec_cli.rs`(새 파일), `tests/followup_per_spec_rules.rs`(새 파일), `tools/repo_scans.json`(줄 임대로, 새 테스트 두 개를 영역 표에 올리는 한 곳), `CLAUDE.md`(재발 줄과 Flow 7 의 마지막 문장만; tier-2 감사), `docs/leader-rules.md`, `docs/lane-rules.md`, `docs/claude-md-rule-provenance.md`, `skills/moai/references/flows/next-flow.md`, `docs/design-review/appendix/operations/card.md`, `.moai/specs/SPEC-20261124-followup-per-spec/spec.md`
- 공유 파일 `src/ops/card.rs`: 2026-11-24 에 다른 카드의 예약이 쥐고 있어서 이 카드의 예약에는 없어요. 구현은 이 파일이 필요 없는 부분(`src/card/command.rs`, `src/card/spec_link.rs`, 테스트, 규칙 글)부터 하고, `Add` 인자 몇 줄은 그 예약이 풀린 뒤 범위 요청이나 줄 임대로 받아요.
- 범위 확장 요청(RED 에서 실제로 깨진 파일만, 그때 리더 노트로 받아요): `tests/recurring_problems_rule.rs`, `tests/repo_claude_md.rs`, `tests/next_flow_skill_scan.rs`, `tests/workflow_size_budget.rs`, `tests/fixtures/help/**`, `src/help.rs` (어느 것이 깨지는지는 unverified)
- read-only: `src/ledger/workflow.rs`(`add_criteria_as`), `src/card/input.rs`, `src/spec_doc.rs`, `src/worktree_gc.rs`(`AddRequest` 를 부르는 곳; `add` 의 모양을 바꾸지 않아요, D-2)
- forbidden: 실제 HOME 과 실제 장부(임시 HOME 만), 네트워크, 장부 스키마 바꾸기, `card spec` 의 동작, 기존 카드와 장부 행의 이행

## Plan

1. RED: 위 AC 의 테스트 이름으로 `tests/card_add_spec_cli.rs` 와 `tests/followup_per_spec_rules.rs` 를 먼저 써서 실패를 봐요.
2. `src/ops/card.rs` 의 `Add` 에 `--spec`, `--follow-up`, `--split-of`(둘은 서로 배타, 셋 다 `--spec` 필요)를 더해요.
3. `src/card/spec_link.rs` 에 열린 카드와 후속 관계를 찾는 읽기 함수를, `src/card/command.rs` 에 `add_with_spec`(검사 뒤 `add` 와 같은 한 트랜잭션에 `card.spec` 이벤트를 함께 써요)를 더해요. `add` 와 `AddRequest` 의 모양은 그대로 둬요.
4. GREEN 뒤 규칙 글 다섯 곳을 고치고(`CLAUDE.md` 는 두 문장, 출처 행은 `tests/recurring_problems_rule.rs` 가 고정한 문구를 지키며 이 SPEC 을 더해요), 고정 테스트가 깨지면 범위 확장을 요청해요.
5. `check affected` 를 READY head 에서 한 번, 위험 경로(거절하는 명령)는 임시 장부에서 직접 한 번 돌려요.

## Risks

- `card spec C-n <SPEC-ID>` 로 나중에 연결하는 길은 막지 않아요(D-3). 그 길로 한 SPEC 에 카드가 여럿 붙을 수 있어요. 이 SPEC 은 새 길(`card add --spec`)에만 한도를 두고, 그 길이 쓰인 뒤에도 늘면 다음 카드가 `card spec` 에 같은 검사를 더해요.
- 기존 장부에는 이 SPEC 이전의 카드가 한 SPEC 에 여럿 있어요. 새 `card add --spec` 은 그 SPEC 에 대해 거절돼요(`--split-of` 또는 `criterion-add` 로 돌아가요). 이것이 의도예요.
- "열린 카드" 는 `done` 이 아닌 상태 전부(`draft`, `ready`, `active`, `review`, `hold`)예요. `hold` 카드가 후속 카드를 오래 막을 수 있어요. 그때는 `criterion-add` 로 기준을 더해요.
- 규칙 글이 바뀌어도 리더가 새 카드를 열 수 있어요. 글은 한도를 알리는 것이고, 막는 것은 `card add --spec` 의 거절이에요.
- `CLAUDE.md` 는 tier-2 감사 대상이고 `tests/repo_claude_md.rs`, `tests/recurring_problems_rule.rs` 가 문장 일부를 고정해요. 고정된 문장은 지우지 않고 뒤에 더해요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 후속·분할 관계는 `card.spec` 이벤트 payload 의 `relation`(`follow-up`, `split-of`)과 `of` 로 남기고 새 열은 만들지 않아요 | 카드 연결이 이벤트로만 남는 현재 구조(`spec_link.rs` 머리말)를 따라요. `card export` 가 이벤트를 그대로 실어 나르고 장부 스키마 버전을 올리지 않아요 | 새 이벤트 payload 키를 읽지 않으면 이전과 같아요 |
| D-2 | `add` 와 `AddRequest` 는 그대로 두고 `add_with_spec` 을 따로 만들어요 | `src/worktree_gc.rs` 가 `AddRequest` 를 직접 짓고 있어 모양을 바꾸면 그 파일도 범위가 돼요 | `add_with_spec` 을 지우면 돼요 |
| D-3 | 한도는 `card add --spec` 에만 두고 `card spec` 은 건드리지 않아요 | 이미 한 SPEC 에 여러 카드가 붙은 장부가 많아 `card spec` 을 막으면 기존 작업이 깨져요 | 나중 카드에서 `card spec` 에 같은 검사를 더해요 |
| D-4 | "열린 카드" 는 `done` 이 아닌 모든 상태예요 | `card criterion-add` 가 `done` 만 거절하는 것과 같은 선이에요 | 상태 목록 한 곳을 바꿔요 |
| D-5 | 거절은 항상 이름 붙은 길(`--follow-up`, `--split-of`, `criterion-add`)을 함께 적어요 | 거절만 하면 리더가 우회로를 찾아요 | 문장만 바뀌어요 |
| D-6 | 후속 카드는 SPEC 마다 열린 것 하나이고 그 SPEC 의 본 카드와 별개예요. 본 카드가 열려 있어도 첫 `--follow-up` 은 받아요 | 카드 본문이 "한 SPEC 에 열린 후속 카드는 하나" 라고 했어요 | 규칙 하나를 뺀 REQ-4 만 바뀌어요 |
| D-7 | 규칙 글이 말하는 대상은 리더예요(Refill 3 의 사람)이고 레인은 노트에 적기만 해요(REQ-12) | 카드 만들기는 리더의 일이에요(`CLAUDE.md` Leader rules) | 레인 규칙 한 줄을 지워요 |
| D-8 | 거절은 observe 없이 처음부터 막아요 | `--spec` 은 새 선택 플래그라 기존 호출이 막힐 수 없고, 사용자가 카드 본문에서 거절을 골랐어요 | 거절을 경고로 바꾸는 한 곳이에요 |

## Result

구현은 브랜치 `spec/followup-per-spec-20261124`(기준 NeXT `34d87870dc`)의 커밋 SPEC `699181646e` → 승인 기록 `7516ff12b9` → 규칙 글과 규칙 테스트 `f5b6d5b2f4` → 코드 안쪽과 RED 테스트 `2d18bf5d38` → 인자 받기와 GREEN `7d50e9268b` 순서예요. RED 는 커밋하지 않고 작업 트리에서 먼저 돌렸어요: `card_add_spec_cli` 10개 중 9개 실패(`target/c443/red1.log`, 실패 이유는 `unexpected argument '--spec'`, 통과한 하나는 바뀌지 않은 `card add` 테스트), 규칙 테스트 2개 모두 실패(`target/c443/red2.log`).

- AC-1: met — `add_with_spec_links_in_one_write`, `a_missing_or_broken_spec_is_refused` passed(`target/c443/green6.log`, 79개 중 79개 통과). 없는 SPEC 은 `not_found`(`spec not found` 문장), 형식이 틀린 id 와 error 있는 SPEC 은 `invalid_input` 이고 카드와 사건이 0 이에요(REQ-2 는 거절만 정하고 kind 는 정하지 않아요).
- AC-2: met — `a_second_open_card_on_one_spec_is_refused`, `a_done_card_does_not_count` passed.
- AC-3: met — `one_open_follow_up_per_spec`, `a_closed_follow_up_makes_room` passed.
- AC-4: met — `split_of_names_its_origin`, `split_of_needs_the_same_spec` passed.
- AC-5: met — `relation_flags_need_a_spec_and_exclude_each_other` passed.
- AC-6: met — `card_cli`, `card_dispatch_cli`(이웃으로 `card_spec_cli`, `card_criterion_add_cli` 도) 가 `green6.log` 에서 그대로 통과했어요. `an_add_without_spec_is_unchanged` 도 passed.
- AC-7: met — `leader_and_lane_rules_name_the_follow_up_card` passed.
- AC-8: met — `recurrence_and_remote_failure_rules_use_existing_cards` 와 `recurring_problems_rule`, `repo_claude_md`, `next_flow_skill_scan` 이 함께 통과했어요(`target/c443/green4.log`, 154개 중 154개; 이 뒤에 바뀐 규칙 글은 없어요).
- 뮤턴트(적용을 스크립트가 낱말 한 번 일치로 확인하고 `git restore` 로 되돌린 뒤 `git status` 가 깨끗함을 확인): ① 열린 카드 검사의 `Plain if open.is_empty()` 를 `Plain` 으로 → exit 100(`a_second_open_card_on_one_spec_is_refused`) ② 후속 카드 찾기를 `find(|_| false)` 로 → exit 100(`one_open_follow_up_per_spec`) ③ `--split-of` 의 같은 SPEC 검사를 항상 거짓으로 → exit 100(`split_of_needs_the_same_spec`) ④ `doc.errors() > 0` 를 `> 1_000_000` 으로 → exit 100(`a_missing_or_broken_spec_is_refused`) ⑤ `--follow-up` 의 `requires = "spec"` 삭제 → exit 100(`relation_flags_need_a_spec_and_exclude_each_other`) ⑥ `state <> 'done'` 를 다른 값으로 → exit 100(`a_done_card_does_not_count`) ⑦ `CLAUDE.md` 재발 문장 삭제 → exit 100 ⑧ `leader-rules.md` 의 `only when the SPEC has none open` 삭제 → exit 100. 로그는 `target/c443/m1.log`~`m8.log`.
- 줄 임대로 받은 공유 파일: `tools/repo_scans.json`(새 규칙 테스트를 `reads-docs`, `reads-root-files`, `reads-skills` 에 올림), `tests/areas.toml`(`followup` 단어), `tests/next_flow_skill_scan.rs`(`FLOW_REF_TIP_BYTES` 7974 → 8051, 원격 실패 문장이 77바이트 늘었어요), `src/ops/card.rs`(`Add` 인자 셋과 갈래). 모두 받은 즉시 돌려줬어요.
- 위험한 길: 이 변경은 지우거나 되돌리거나 프로세스를 시작하거나 권한 규칙을 만들지 않아요(`none`). 거절 길(두 번째 열린 카드, 두 번째 후속 카드, 틀린 `--split-of`)은 임시 HOME 의 임시 장부에서 실제 `moai` 명령으로 돌렸고 카드와 사건 수가 그대로임을 확인했어요.
- 범위 밖·미검증: `card spec C-n <SPEC-ID>`(나중 연결)는 한도를 두지 않아요(D-3). 기존 장부의 SPEC 당 여러 카드는 그대로예요. 실제 장부와 실제 HOME 에는 쓰지 않았어요.
