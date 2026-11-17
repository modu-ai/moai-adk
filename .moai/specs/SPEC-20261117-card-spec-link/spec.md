---
id: SPEC-20261117-card-spec-link
title: 카드 ↔ SPEC 연결 — moai card spec, next·take 경고, spec show 의 로컬 카드 (SDD-4)
status: draft
revision: 1
created: 2026-11-17
parent: SPEC-20261117-sdd-foundation
---

## Intent

저장소를 바꾸는 일은 모두 SPEC으로 하고(결정 U-A), 카드는 이 컴퓨터의 원장에 있는 개인 할 일이에요.
지금은 카드가 어느 SPEC의 일인지 적을 곳이 없어서, 레인이 승인되지 않은 SPEC으로(또는 SPEC 없이) 일을
시작해도 아무도 알려 주지 않아요. 이 SPEC은 카드 쪽에만 SPEC 연결을 두고(`moai card spec`), `card next`·
`card take`가 그 SPEC의 상태를 보여 주며 승인 전이면 경고하게 해요. SPEC 파일에는 카드 번호가 들어가지
않아요(결정: 카드 번호는 공유 파일에 쓰지 않음).

## Non-goals

- 병합 관문의 강제 검사(승인된 SPEC 없이는 병합 거절)는 뒤 조각(SDD-6, 관문 G5)이 해요. 여기서는 경고와
  `take --require-approved-spec` 거절까지만 해요.
- `card add --spec`, 기준 n = AC-n 대응, 승인 기록을 쓰는 명령은 이 조각에 없어요(후속).
- 원장 스키마 변경. 연결은 기존 `event` 표의 `card.spec` 사건으로 남겨요.
- 스킬 글(`skills/moai/**`) 변경. `card spec` 안내는 SDD-5의 `references/internal/sdd.md`가 맡아요.

## Requirements

- REQ-1: When `moai card spec <card> <SPEC-ID>` runs, the tool shall record the link as one `card.spec` event in this machine's ledger and leave the SPEC file unchanged.
- REQ-2 (REQ-1): IF the SPEC id has no SPEC id form, the SPEC folder is not in `.moai/specs`, or the SPEC has an error finding, THEN the tool shall refuse and write nothing.
- REQ-3 (REQ-1, REQ-4): IF another connection holds the ledger's write lock past the busy timeout, THEN the tool shall answer `ledger busy` and write nothing.
- REQ-4: When `--unlink` is given, the tool shall record the removal as one `card.spec` event.
- REQ-5: The `card show` and `card list` answers shall print the SPEC id each card names.
- REQ-6: When `card next` or `card take` lists a card, the tool shall print the status of the SPEC it names and warn when a card that changes the repository has no approved SPEC.
- REQ-7 (REQ-6): IF the named SPEC is gone or cannot be read, THEN the tool shall report it as not approved.
- REQ-8: Where `--require-approved-spec` is given, `card take` shall refuse a card that changes the repository without an approved SPEC and reserve nothing.
- REQ-9: When `moai spec show` runs, the tool shall list the cards of this machine's ledger that name the SPEC, marked local.
- REQ-10 (REQ-9): IF the project has no profile or its ledger is absent or unusable, THEN the tool shall show the SPEC with no local card and the reason.
- REQ-11: The ledger shall keep the link without a schema change.
- REQ-12: The repository SPECs shall pass `moai spec check --all`.

## Acceptance

- AC-1 (REQ-1, REQ-4, REQ-5): 연결·다시 연결·같은 연결(사건 없음)·끊기가 되고, `show`·`list`가 SPEC id를 보여 주며 SPEC 파일 바이트는 그대로예요.
  verify: `cargo test --locked --offline --test card_spec_cli a_card_links_to_a_spec_unlinks_and_show_and_list_print_it`
- AC-2 (REQ-2): 없는 SPEC은 `not_found`, SPEC id 모양이 아닌 글·오류 있는 SPEC·`#번호`는 `invalid_input`, 없는 카드는 `not_found`이고 원장 폴더는 그대로예요.
  verify: `cargo test --locked --offline --test card_spec_cli a_missing_or_unparsable_spec_is_refused_and_nothing_is_written`
- AC-3 (REQ-3): 다른 연결이 쓰기 잠금을 쥐면 `ledger busy`(`io`)이고 `card.spec` 사건은 0개, 다른 연결의 쓰기는 남아요.
  verify: `cargo test --locked --offline --test card_spec_cli a_busy_ledger_refuses_the_link_and_keeps_the_other_write`
- AC-4 (REQ-6, REQ-7): `next`·`take`가 `spec`(id·found·status·errors·approved)과 `warnings`를 보여 주고, `research`·`audit` 카드는 경고하지 않으며, 승인 뒤 경고가 사라져요.
  verify: `cargo test --locked --offline --test card_spec_cli next_and_take_print_the_spec_status_and_warn_until_it_is_approved`;
  `cargo test --locked --offline --lib spec_link`
- AC-5 (REQ-7, REQ-8): `--require-approved-spec`은 초안 SPEC·SPEC 없음·사라진 SPEC을 거절하고 예약을 만들지 않으며, 승인된 SPEC과 `research` 카드는 받아요.
  verify: `cargo test --locked --offline --test card_spec_cli take_with_require_approved_spec_refuses_until_the_spec_is_approved`
- AC-6 (REQ-9, REQ-10): `spec show`의 `local_cards`가 지금 그 SPEC을 가리키는 카드만 번호 순으로 보여 주고, 원장 없음·프로필 없음·쓸 수 없는 원장은 빈 목록과 이유예요.
  verify: `cargo test --locked --offline --test card_spec_cli spec_show_`
- AC-7 (REQ-11): 연결 뒤에도 `user_version`이 그대로이고, 사건 payload는 `{"from": null, "to": <id>}`, `card export`의 `events.jsonl`에 남으며, 기존 카드 시험이 그대로 통과해요.
  verify: `cargo test --locked --offline --test card_spec_cli the_link_is_an_event_of_the_current_schema_and_the_export_keeps_it`;
  `cargo test --locked --offline --test card_cli --test card_dispatch_cli --test card_work_cli --test card_export_cli`
- AC-8 (REQ-12): 저장소의 SPEC이 모두 통과해요.
  verify: `cargo test --locked --offline --test spec_cli the_repository_specs_pass_check`
- AC-9 (REQ-1…REQ-12): 이웃 검사가 모두 통과해요.
  verify: `cargo test --locked --offline --bin moai`; `cargo test --locked --offline --test help_layout`;
  `cargo test --locked --offline --test home_isolation_scan`; `cargo test --locked --offline --test product_cli`;
  `cargo test --locked --offline --test spec_cli --test spec_doc --test mode_cards_53`; `cargo fmt --check`;
  `cargo clippy --locked --offline --all-targets -- -D warnings`;
  `python3 -m unittest tests/repo_scans_test.py tests/ci_scope_test.py tests.skill_references_test`

## Scope

- write: `src/card/spec_link.rs`, `src/card/mod.rs`, `src/card/command.rs`, `src/card/dispatch.rs`,
  `src/card/work.rs`, `src/ops/card.rs`, `src/ops/spec.rs`, `src/spec_doc.rs`, `tests/card_spec_cli.rs`,
  `tests/card_cli.rs`, `docs/design-review/appendix/operations/card.md`,
  `docs/design-review/appendix/operations/spec.md`, `.moai/specs/SPEC-20261117-card-spec-link/spec.md`
- read-only: `src/ledger/schema.rs`, `src/ledger/workflow.rs`, `src/ledger/reservation.rs`, `src/contract.rs`,
  `skills/moai/**`
- forbidden: `.gitignore`, 원장 스키마 변경, 실제 HOME, 네트워크 쓰기(push, PR, Issue)

## Plan

1. 시험 먼저: `tests/card_spec_cli.rs` 8개를 쓰고 RED(`unrecognized subcommand 'spec'`, `local_cards` 없음)를 기록해 따로 커밋.
2. `spec_doc::specs_dir`·`find`로 SPEC 찾기를 라이브러리로 옮기고, `card::spec_link`(연결 쓰기·읽기·상태·경고)를 만듦.
3. `card show`·`list`·`next`·`take`와 `ops/card.rs`의 `spec` 동작·`--require-approved-spec`, `ops/spec.rs`의 `show --home`·`local_cards`.
4. GREEN, 이웃 시험, 형식·clippy·Python, 새 규칙 mutation 5개 이하.
5. 동작 문서(`card.md`·`spec.md`)와 이 SPEC의 Result.

## Risks

- 경고 관문은 예약 트랜잭션 밖에서 연결을 읽어요: 읽은 직후 다른 세션이 연결을 바꾸면 그 `take`는 옛 상태로
  판정해요 → 이 조각은 경고 관문이고, 강제는 병합 관문(뒤 조각)이 커밋 트레일러로 해요.
- 카드 종류로 "저장소를 바꾸는 카드"를 정해요(`research`·`audit` 말고 모두): 종류를 잘못 고른 카드는 잘못 경고되거나
  빠져요 → 경고일 뿐이고, 종류는 `card add --kind`로 정해요.
- 연결이 사건으로만 있어서 `card export`의 카드 파일(`C-n.md`)에는 보이지 않아요(`events.jsonl`에는 있음) → 되살리기
  형식(K3)이 생길 때 다뤄요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D1 | 연결은 카드 쪽에만: `card spec <card> <id>`·`--unlink`; SPEC에는 `cards:` 키가 없음 | 리더 결정(되돌릴 수 있음), 결정 U-A…U-D | SPEC 앞머리에 `cards:` 추가 |
| D2 | 연결은 `card.spec` 사건(payload `from`·`to`)이고 지금 연결 = 가장 새 사건의 `to`; 스키마 버전 1 그대로 | 가장 싼 길(core-rules 3), 내보내기·되살리기가 사건 종류를 가리지 않음 | 버전 2 스키마에 `card.spec_id` 열과 이전 시험 |
| D3 | "해석된다" = `spec check`가 오류로 보는 발견이 없음(경고는 허용) | 구현 판단 | 읽기만 되면 연결 |
| D4 | 저장소를 바꾸는 카드 = `research`·`audit`을 뺀 모든 종류; SPEC이 없어도 경고 | 결정 U-A(조사·질문은 카드만) | `NO_REPO_CHANGE` 목록 바꾸기 |
| D5 | 승인 = 상태 `approved`(옛 Go 상태는 새 낱말로 읽음)이고 오류 0 | 구현 판단 | `completed`도 승인으로 |
| D6 | 같은 연결을 다시 주면 사건 없이 `changed: false`; 끊긴 카드를 다시 끊어도 같음 | 구현 판단 | 매번 사건 |
| D7 | `--require-approved-spec` 거절은 `card take refused: --require-approved-spec: …`이고 예약 전 | 리더 설계 | 거절 대신 경고만 |
| D8 | `spec show`는 원장 문제(없음·프로필 없음·쓸 수 없음)로 실패하지 않고 `local_cards.ledger`에 이유를 적음 | 구현 판단 | 원장 오류를 그대로 실패로 |
| D9 | 스킬 글은 이번에 바꾸지 않음(`mode-agent.md`는 한도 근처, 안내는 SDD-5의 `sdd.md`) | 리더 지시(키우지 않기) | 내부 참조에 한 줄 추가 |

## Result

pending — 구현 에이전트가 검사 뒤에 채워요.

- AC-1: pending
- AC-2: pending
- AC-3: pending
- AC-4: pending
- AC-5: pending
- AC-6: pending
- AC-7: pending
- AC-8: pending
- AC-9: pending
