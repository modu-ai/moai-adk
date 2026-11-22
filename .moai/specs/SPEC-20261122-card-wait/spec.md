---
id: SPEC-20261122-card-wait
title: "moai card wait returns the next matching note or a timeout"
status: draft
revision: 1
created: 2026-11-22
---

# moai card wait returns the next matching note or a timeout

## Intent

사용자가 승인한 재설계(2026-11-22, `~/.moai/reports/moai-adk/agentic-redesign-2026-11-22.md` 카드 ⑧, 레인 백그라운드 조사 `lane-background-survey-2026-11-22.md` F8 고칠 방법 3·F10)의 일이에요. 레인은 자기 카드에 감사 판정이나 리더 답이 오기를 기다릴 때 모니터, 예비 깨우기, `sleep` 반복문을 겹쳐 걸어요. 레인마다 손으로 만든 감시 스크립트가 일곱 가지이고, 끝나는 시각이 없는 것도 있어요(F1·F8). 그래서 한 레인이 소식 없이 한 시간에 2–4번 깨요. 같은 노트를 두 번 받기도 하고(F9·F10), 끝나지 않는 셸도 남아요.

이 SPEC 은 `moai card wait` 하나를 만들어요. 이름 붙인 카드들에 새 노트가 오거나, 어느 카드든 첫 줄에 `all lanes` 가 든 공지가 오면 그 노트들을 한 번 출력하고 끝나요. 오지 않으면 정한 시각에 `timeout` 으로 끝나요. 시간 제한은 반드시 줘야 해서 끝없이 기다리는 일이 없어요. 원장은 읽기만 해요.

## Non-goals

- 레인 규칙 글(언제 기다리고 언제 다시 거는지)은 리더 지시와 `moai card wait` 를 부르는 스킬·워크플로 쪽 일이에요. 이 SPEC 은 명령만 만들어요.
- 원장 변경을 알림으로 받는 길(파일 감시, 데이터베이스 알림)은 만들지 않아요. 짧은 읽기 전용 조회를 정한 간격으로 되풀이해요.
- 레인 이름이 본문에 나오는 노트를 고르는 일은 하지 않아요. 노트 본문에는 다른 레인 이름이 흔히 나와서 잡음이 커요.
- `card inbox` 와 `card show` 는 바꾸지 않아요.

## Requirements

- REQ-1: When `moai card wait` starts, it shall require `--timeout` between 1 second and 12 hours in the form `<n>`, `<n>s`, `<n>m` or `<n>h` · failure: none — REQ-2 refuses anything else.
- REQ-2 (REQ-1): IF `--timeout` or `--every` is missing, malformed or out of range, or no `--card` is given, THEN card wait shall refuse as a usage error with exit 2 before it reads the ledger.
- REQ-3: When card wait starts, it shall fix a start mark that is `--since` when given and otherwise the largest note id in the ledger at that moment · failure: none — REQ-6 handles a ledger it cannot read.
- REQ-4: When a poll finds local notes above the start mark that are on a named card or whose first line holds `all lanes` in any case on a card that is not done, and whose author is not a `--not-author` name, card wait shall print all of them in note-id order with `outcome: "note"` and exit 0 · failure: none — a poll that finds none waits for the next one.
- REQ-5: When the timeout passes before a poll finds a matching note, card wait shall print `outcome: "timeout"` with the seconds it waited and exit 0 · failure: none — the wait never runs longer than the timeout plus one poll.
- REQ-6: IF a named card does not exist or the ledger is absent or refused, THEN card wait shall refuse with exit 1 and its error kind before it waits, and nothing is created.
- REQ-7: When card wait polls, it shall open the ledger read-only every `--every` seconds (default 15, between 1 and 300) and write nothing · failure: none — a poll that hits a busy ledger counts as a poll that found nothing.

## Acceptance

- AC-1 (REQ-1, REQ-4): 기다리는 동안 다른 프로세스가 이름 붙인 카드에 노트를 쓰면 그 노트 하나가 `outcome: "note"` 로 나오고 종료 코드는 0이에요. 기다리기 전에 있던 노트는 나오지 않아요 · verify: `cargo test --locked --offline --test card_wait_cli a_note_written_during_the_wait_ends_it_and_older_notes_are_not_returned`
- AC-2 (REQ-5): 맞는 노트가 없으면 `--timeout 2s --every 1s` 가 `outcome: "timeout"` 으로 2초 이상, 정한 상한 안에 끝나요 · verify: `cargo test --locked --offline --test card_wait_cli without_a_matching_note_the_wait_ends_at_its_timeout`
- AC-3 (REQ-2): `--timeout` 이 없거나, `0`·`13h`·`5x` 이거나, `--every 0`·`301` 이거나, `--card` 가 없으면 종료 코드 2이고 원장을 열지 않아요 · verify: `cargo test --locked --offline --test card_wait_cli a_wait_without_a_valid_timeout_or_card_is_a_usage_error`
- AC-4 (REQ-4): `--not-author` 이름의 노트와 이름 붙이지 않은 카드의 보통 노트는 나오지 않고, 다른 카드의 첫 줄 `ALL LANES` 공지는 나와요 · verify: `cargo test --locked --offline --test card_wait_cli own_notes_and_other_cards_are_skipped_and_an_all_lanes_line_is_returned`
- AC-5 (REQ-3): `--since N` 을 주면 이미 있는 N 보다 큰 노트가 첫 조회에서 바로 나와요 · verify: `cargo test --locked --offline --test card_wait_cli since_returns_notes_already_above_the_mark_at_once`
- AC-6 (REQ-6, REQ-7): 없는 카드는 기다리기 전에 거절되고 종료 코드 1이에요. 원장이 없는 HOME 에서는 거절되고 아무것도 만들지 않아요 · verify: `cargo test --locked --offline --test card_wait_cli an_unknown_card_or_an_absent_ledger_is_refused_before_the_wait`

## Scope

- write: `.moai/specs/SPEC-20261122-card-wait/**`, `src/card/wait.rs`(새), `src/card/mod.rs`, `src/ops/card.rs`, `tests/card_wait_cli.rs`(새), `docs/design-review/appendix/operations/card.md`
- read-only: `src/card/inbox.rs`, `src/card/input.rs`, `src/ledger/**`, `src/main.rs`
- forbidden: 실제 HOME, 상태 폴더 이름을 글자로 쓴 시험 기대값, 새 Python 시험, 원장에 쓰기

## Plan

1. RED 시험 `tests/card_wait_cli.rs`(AC-1~AC-6): 임시 HOME 과 임시 프로젝트, `moai card add`·`card note` 로 만든 원장. AC-1 은 기다림을 자식 프로세스로 띄운 뒤 노트를 써요. 명령이 없으니 모두 실패해요.
2. `src/card/wait.rs`: 시간 값 읽기(`<n>`, `s`, `m`, `h`), 시작 표시, 한 번의 조회(첫 줄만 데이터베이스 밖으로, `card inbox` 처럼), 고르기 규칙은 행만 받는 순수 함수로 두고 단위 시험, 마감까지의 반복.
3. `src/ops/card.rs` 에 `Wait` 동작과 도움말, `src/card/mod.rs` 에 모듈, `card.md` 에 표 한 줄.

## Risks

- 원장이 다른 레인의 쓰기로 잠깐 바쁘면 그 조회는 아무것도 못 찾은 것으로 쳐요. 다음 조회에서 찾으니 늦어질 뿐 놓치지는 않아요. 노트 id 는 늘기만 해서 시작 표시 위의 노트는 언젠가 보여요.
- `all lanes` 를 첫 줄의 평범한 글로 고르니, 공지가 아닌 노트가 그 말을 첫 줄에 쓰면 레인이 한 번 더 깨요. 2026-11-22 원장 C-319 의 118개 노트 중 첫 줄에 그 말이 든 노트는 3개이고 셋 다 공지였어요.
- 조회 간격이 짧으면 레인이 많을 때 읽기가 늘어요. 기본 15초, 최소 1초예요. 한 번의 조회는 첫 줄만 읽어요.
- 기다리는 동안 출력이 없어서 사람이 화면으로 진행을 보지 못해요. 끝날 때 한 번만 출력하는 것이 목적이에요(F1).

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 시간이 다 돼도 종료 코드는 0이고 `outcome: "timeout"` 으로 알려요. | 종료 코드는 0·1·2 만 쓰는 계약이에요(`docs/rust-product-path.md` 152행). 시간이 다 된 것은 실패가 아니에요. | 카드 열림 |
| D-2 | 공지는 "첫 줄에 `all lanes` 가 대소문자 무관으로 들어 있음"으로 골라요. 따로 정한 표시는 없어요. | 2026-11-22 원장 C-319 의 공지 3개(6001, 6345, 6587)가 모두 첫 줄에 그 말을 썼어요. 조사 F10 이 이 표시를 남기라고 했어요. | 카드 열림 |
| D-3 | 시작 표시의 기본값은 기다리기 시작할 때 원장 전체의 가장 큰 노트 id 예요. | 노트 id 는 원장 전체에서 늘기만 해요. 이름 붙인 카드와 공지 둘 다에 한 표시로 맞아요. | 카드 열림 |
| D-4 | 맞는 노트가 여럿이면 한 번에 모두 내요. 각 노트는 첫 줄만 200자까지 내요(`card inbox` 와 같아요). | 레인은 한 번 깨서 모두 보고, 본문은 `card show` 로 읽어요. 노트 하나가 1 MiB 일 수 있어요. | 카드 열림 |
| D-5 | `--card` 는 여러 번 주거나 쉼표로 이어 줄 수 있어요. 카드 이름은 `card show` 와 같은 꼴(`C-12`, `#3066`)이에요. | 조사 F8 의 `--cards C-1,C-2` 와 카드 완료 기준의 `--card C-n` 둘 다 받아요. | 카드 열림 |

## Result

(status draft — written when the change lands)
