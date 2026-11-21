---
id: SPEC-20261121-card-hold-gate
title: "A structured HOLD and prerequisite on a card are read by card next and card take before any reservation or worktree, and nothing lifts a HOLD by itself"
status: draft
revision: 1
created: 2026-11-21
---

# 카드의 구조화된 HOLD·선행 조건을 next 와 take 가 워크트리·예약 전에 읽는다

## Interview

- Q1: 무엇이 문제였나요? → A: 사용자가 확인한 현장 점검(2026-11-21)에서 Codex CLI 화면 13개(리더 1 + 레인 12)를 다른 프로젝트에서 보니, 레인 7개가 준비된 후보를 반복해서 조회했지만 HOLD·선행 조건·범위 때문에 여럿을 시작하지 못했어요. 레인 둘은 HOLD 를 읽기 전에 워크트리를 만들었고, 한 레인은 카드를 잡은 뒤 HOLD 를 읽고 돌려줬어요. 그 프로젝트와 살아 있는 레인 설정·보호 설정은 건드리지 않아요.
- Q2: 지금은 어떻게 되어 있나요? → A: 원장에 `hold` 상태는 있지만 `card withdraw` 로만 가고(그 카드는 `card.withdraw` 이벤트가 이유를 남기고 `card take` 가 이유를 말하며 거절해요), HOLD 의 종류(돈·법무·외부 계약·비밀·실제 데이터)를 구조로 남기는 곳이 없어요. `card next` 는 `ready` 가 아닌 카드를 이유 없이 건너뛰고, `blocks` 링크(`card_link`)를 보지 않으며, 그 링크를 만드는 명령도 없어요. 그래서 HOLD 는 자유 글(카드 본문·노트)로만 적혀서 레인이 읽기 전에 워크트리부터 만들 수 있어요.
- Q3: 무엇을 만들까요? → A: (1) HOLD 를 구조로 기록하고 푸는 명령 `card hold`·`card unhold`(이벤트 `card.hold`·`card.unhold`, 새 상태값 없음), (2) 선행 조건을 구조로 기록하는 명령 `card prereq`(기존 `blocks` 링크), (3) `card next` 가 HOLD·선행 조건이 있는 카드를 이유와 함께 건너뛰고, (4) `card take` 가 예약을 쓰기 전에 거절하고, (5) `card show` 가 둘을 보여 주고, (6) 레인 시작 안내가 워크트리 전에 그걸 읽도록 하는 것까지예요.
- Q4: 자유 글 HOLD 는요? → A: 읽지도 풀지도 추정하지도 않아요. 노트에 "HOLD" 라는 말이 있어도 코드는 구조화된 이벤트만 읽고, HOLD 를 푸는 것은 사람이 `card unhold` 를 부르는 일뿐이에요(리더 HOLD 해제 규칙 그대로).
- Q5: 이번에 안 하는 것은요? → A: 같은 전체 조회를 반복하지 않게 하는 상태 지문(fingerprint)과 상태가 풀리면 다시 시작하는 기능은 다음 조각이에요. 이 SPEC 은 거절과 안내까지만 해요.

## Intent

HOLD 와 선행 조건이 자유 글에만 있으면 레인은 카드를 잡고 워크트리를 만든 뒤에야 알아요. 이 SPEC 은 그 둘을 원장의 구조화된 사실로 만들어 `card next`(건너뜀 이유)와 `card take`(예약 전 거절)가 읽게 하고, 워크트리를 만드는 단계의 안내가 먼저 읽도록 해서 빈 워크트리와 잡았다 돌려주기를 없애요. 풀기는 항상 사람이 해요.

## Non-goals

- 상태값 추가(스키마의 `state` CHECK 는 그대로), 스키마 변경. HOLD 는 이벤트, 선행 조건은 기존 `blocks` 링크예요.
- 자유 글 HOLD 를 자동으로 풀거나 추정하는 것, 리더·사용자 HOLD 결정을 코드가 대신하는 것.
- 같은 전체 조회를 줄이는 상태 지문과 재개 알림(Q5, 다음 조각), `card withdraw` 의 동작 변경.
- 그 현장 프로젝트와 살아 있는 레인의 설정을 바꾸는 것.

## Requirements

- REQ-1: When `card hold C-n --category <category> --reason <text> --by <name>` runs on a card in the draft or ready state that nobody holds, the command shall append one `card.hold` event that records the category, the reason and the by · failure: none
- REQ-2 (REQ-1): IF the category is not one of money, legal, external_contract, secret, real_data or other, THEN `card hold` shall refuse and change nothing.
- REQ-3: When `card unhold C-n --by <name> --reason <text>` runs on a card whose newest hold event is a `card.hold`, the command shall append one `card.unhold` event · failure: none
- REQ-4: Where no `card unhold` command is run, a `card.hold` event shall stay in force whatever the notes of the card say.
- REQ-5: When `card prereq C-n --needs C-m` runs, the command shall add one active `blocks` link from C-m to C-n and refuse a card that needs itself or a card that already needs C-n · failure: none
- REQ-6: When `card next` lists a ready card whose newest hold event is a `card.hold`, the card shall appear in `skipped` with the reason `hold`, the category and the reason, and not in `ready` and not as the next card · failure: none
- REQ-7: When `card next` lists a ready card with an active `blocks` link from a card that is not done, the card shall appear in `skipped` with the reason `prerequisite` that names those cards · failure: none
- REQ-8: IF the card has a hold or an unfinished prerequisite, THEN `card take` shall refuse before it writes any reservation, name the category and reason or the unfinished cards, and say that nothing was taken.
- REQ-9: When `card show` prints a card, the output shall carry its current hold (category, reason, by, time, or null) and its unfinished prerequisites · failure: none
- REQ-10: Where the lane start guidance tells a lane to create a worktree for a card, it shall first tell the lane to read the hold and the prerequisites of the card from `card next` or `card show` and to create no worktree for a card that is skipped for them · failure: none

## Acceptance

- AC-1 (REQ-1…REQ-4): temp-repo and ledger tests: a ready card gets a hold with each of the six categories; a bad category and a held (reserved) or done card are refused with nothing written; `card unhold` appends the unhold event and the card is offered again; only the `card unhold` command ever appends an unhold event (no other command writes one); a note that says HOLD in free text changes nothing in `card next` or `card take`; the event rows hold the category, reason and by.
  verify: `cargo test --offline --locked --test card_hold_gate_cli`
- AC-2 (REQ-5, REQ-7): tests: `card prereq` adds the link (a second call changes nothing), refuses a self need and a two-card cycle; a card whose prerequisite is not done is skipped with `prerequisite` and the card names; when the prerequisite card is done the card is in `ready`.
  verify: `cargo test --offline --locked --test card_hold_gate_cli`
- AC-3 (REQ-6, REQ-8): tests: `card next` skips a held card with the reason `hold`, the category and the reason, not in `ready` and not as `next`; `card take` of a held card or of a card with an unfinished prerequisite exits 1 naming them, says nothing was taken, and leaves no reservation row, no state change and no event (ledger snapshot equal).
  verify: `cargo test --offline --locked --test card_hold_gate_cli`
- AC-4 (REQ-9): test: `card show` carries `hold` (null without one, the fields with one) and `prerequisites`.
  verify: `cargo test --offline --locked --test card_hold_gate_cli`
- AC-5 (REQ-10): the named guidance sentences are in the shipped agent-mode text and the skill source manifest matches it; the card commands have operations documents, help rows, examples and goldens.
  verify: `cargo test --offline --locked --test card_hold_gate_cli guidance`; `cargo test --offline --locked --bin moai-cli help_table_names_every_command_once`; `cargo test --offline --locked --test help_layout`
- AC-6 (REQ-1…REQ-9): at most five mutants of the new code, each turning a test red, are restored.
  verify: manual: the mutant table in the Result.

## Scope

- write: `.moai/specs/SPEC-20261121-card-hold-gate/**`, `src/ledger/workflow.rs` (`next` and the hold, unhold and prerequisite functions), `src/ledger/reservation.rs` (the refusal before the reserve write), `src/card/work.rs`, `src/card/dispatch.rs`, `src/card/command.rs`, `src/card/input.rs`, `src/ops/card.rs`, `src/help.rs`, `tests/fixtures/help/all.txt`, `tests/fixtures/help/all-60.txt`, `tests/fixtures/help/all-80.txt`, `tests/fixtures/help/all-dumb.txt`, `docs/design-review/appendix/operations/card.md`, `skills/moai/workflows/mode-agent.md`, `skills/moai/references/source-manifest.json`, `tests/card_hold_gate_cli.rs`
- read-only: `src/ledger/schema.rs`, `src/card/spec_link.rs`, `tests/common/home.rs`
- forbidden: schema changes, the real HOME and ledger, the project the field check came from
- 읽은 사실(2026-11-21, tip `64faee764c`): `src/ledger/workflow.rs` 의 `next` 는 `state <> 'done'` 카드 중 `ready` 만 후보로 올리고 건너뛴 이유는 held, scope_overlap, empty_scope, review_no_holder 넷이에요(`SkipReason`). `card_link` 의 `blocks` 는 스키마에 있지만(`src/ledger/schema.rs`) 읽거나 쓰는 코드와 명령이 없어요. `withdraw`(`src/ledger/workflow.rs:729`)는 `hold` 상태와 `card.withdraw` 이벤트 선례예요. `take` 는 `src/card/dispatch.rs:366` 에서 곧바로 `reserve` 를 불러요. 쓰기 파일 대부분(`work.rs`, `dispatch.rs`, `help.rs`, `mode-agent.md`)은 다른 카드가 예약 중일 수 있어, 필요한 파일을 쓰기 전에 보유 카드가 반환하는 순서를 리더에게 물어요.

## Plan

1. **S1 원장**: RED — `tests/card_hold_gate_cli.rs` 를 임시 HOME 과 원장으로 먼저 써요(AC-1~AC-4). GREEN — `src/ledger/workflow.rs` 에 hold·unhold·prerequisite 함수와 `next` 의 건너뜀 이유, `card take` 의 예약 전 거절.
2. **S2 명령·표시·도움말**: `card hold`·`card unhold`·`card prereq` 를 `src/ops/card.rs`·`src/card/*` 에 잇고 `card show` 를 넓혀요. 운영 문서, 도움말 행·예시·골든.
3. **S3 안내**: `mode-agent.md` 의 시작 단계에 "워크트리 전에 HOLD·선행 조건을 읽는다" 문장을 넣고 매니페스트를 다시 생성해요.
4. 변형 5개 이하, 이웃 검사를 돌려 Result 에 적어요.

## Decisions

각 결정은 (무엇 · 선택지 · 기본값 · 되돌릴 수 있음 · 범주 · 기한 · 결정자) 순서예요. 모두 제안이고 승인은 SPEC 이 병합될 때 사용자·리더가 골라요.

- D1 (제안, 되돌릴 수 있음, 범주 none): HOLD 는 새 상태값이 아니라 `card.hold`·`card.unhold` 이벤트예요. 대안: `hold` 상태 재사용. 기본값: 이벤트. 이유: `hold` 상태는 이미 철회(withdraw)를 뜻하고 스키마 CHECK 는 값을 고정해요.
- D2 (제안, 되돌릴 수 있음, 범주 none): HOLD 종류는 money, legal, external_contract, secret, real_data, other 여섯이에요(MoAI 의 HOLD 다섯 범주 + 그 밖). 대안: 자유 단어. 기본값: 여섯 고정.
- D3 (제안, 되돌릴 수 있음, 범주 none): 선행 조건은 기존 `blocks` 링크(`from_card` 가 `to_card` 를 막음)를 쓰고 새 링크 종류를 만들지 않아요. 대안: 새 종류. 기본값: 기존.
- D4 (제안, 되돌릴 수 있음, 범주 none): 사람이 HOLD 를 거는 권한은 제한하지 않아요(카드를 만들고 고치는 쪽과 같음). 풀기도 같지만 사람이 부를 때만 일어나요. 대안: 리더만. 기본값: 제한 없음, 이벤트가 by 를 남겨요.

## Risks

- 쓰기 파일 대부분을 다른 카드가 예약 중일 수 있어 순서가 어긋나면 S2·S3 가 밀려요. S1(`workflow.rs`, 새 시험 파일)만으로도 `next`·`take` 가 읽기 시작해요.
- 기존 시험이 `card next` 의 건너뜀 목록이나 `card show` 의 키를 정확히 맞춘다면 새 키(`hold`·`prerequisites`) 때문에 빨개질 수 있어요. 이웃 검사에서 드러나요.
- `blocks` 링크가 이미 다른 도구로 만들어진 원장이 있으면(가져오기) 그 의미(누가 누구를 막는지)가 이 SPEC 과 달라질 수 있어요. 방향은 스키마 주석과 같게 `from_card` 가 `to_card` 를 막는 것으로 읽고 시험으로 고정해요.
- 이 SPEC 은 같은 전체 조회를 줄이는 기능을 하지 않아 현장의 반복 조회는 남아요(Q5, 다음 조각).

## Result

(구현 뒤에 적어요.)
