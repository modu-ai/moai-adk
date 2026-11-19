---
id: SPEC-20261119-statusline-waiting-cards
title: "The status line tells cards being worked on from cards waiting for integration"
status: draft
revision: 2
created: 2026-11-19
---

# The status line tells cards being worked on from cards waiting for integration

## Intent

상태줄의 카드 수에서 진행 중 숫자(그때는 `🔥`)는 `active` + `review` 를 한 숫자로 세요. 2026-11-19 사용자가 "진행 중 수가 34 에서 안 줄어드는 이유"를 물었을 때, 그중 약 26장은 이미 push 한 head 가 리더의 묶음(bundle)이나 감사를 기다리는 카드였어요 — 숫자만 봐서는 병목이 레인이 아니라 통합이라는 것이 안 보였어요. 이 SPEC 은 같은 읽기 전용 원장 읽기에서 노트의 마커 줄을 한 번 더 읽어, 일하고 있는 카드와 push 하고 통합을 기다리는 카드를 두 숫자로 나눠 보여 줘요.

## Non-goals

- 원장 쓰기, 새 상태 값(`card_transition` 표 변경), 새 명령 — 상태줄은 읽기만 해요.
- 남은(`🗂️`)·끝난(`✅`) 숫자의 뜻 변경(`🗂️` = draft + ready + hold, `✅` = done 그대로; 남은 숫자의 아이콘은 예전 `📥` 에서 바뀌었어요).
- Codex 상태줄(원장을 읽는 항목이 없어 해당 없음 — 기존 계약 그대로).
- 노트 마커의 새 형식 도입: 레인이 이미 쓰는 줄(`LANE-READY-INTEGRATION`, `감사 요청`, `AUDIT … verdict:`, `ACK`, `PLAN`)만 읽어요.
- `docs/rust-product-path.md` 의 항목 2 문구 갱신: 그 파일은 다른 예약이 잡고 있어 이 SPEC 의 범위 밖이고, 예약이 풀린 뒤 후속 조각으로 해요.

## Requirements

- REQ-1: When the status line shows the card counts, it shall show four numbers, `🗂️ <remaining> / ▶️ <working> / 🔃 <waiting> / ✅ <done>`, where `<working>` plus `<waiting>` equals the cards in `active` or `review`.
- REQ-2: While a card is in `review`, the status line shall count it as waiting.
- REQ-3: While a card is in `active`, the status line shall count it as waiting only when the newest of its notes whose first line starts with a marker is a waiting marker, and as working in every other case.
- REQ-4: When a note's first line starts with `LANE-READY-INTEGRATION`, or starts with `LANE-DONE` and carries `audit requested` or `감사 요청`, or starts with `감사 요청`, or starts with `AUDIT` and carries `verdict: PASS`, the status line shall read it as a waiting marker. · failure: none — 첫 줄이 이 모양이 아니면 마커가 아니라서 읽는 쪽에 실패 경로가 없어요.
- REQ-5: When a note's first line starts with `ACK ` or `PLAN ` (the word and one space, as the lane rules write them), or starts with `AUDIT` and carries `verdict: FAIL`, the status line shall read it as a working marker. · failure: none — 첫 줄이 이 모양이 아니면 중립 노트라 읽음이 바뀌지 않아요.
- REQ-6 (REQ-1): IF the notes cannot be read within the render budget, THEN the status line shall show the three-number form `🗂️ <remaining> / ▶️ <active + review> / ✅ <done>` and no `🔃`.
- REQ-7: The status line shall read the notes with the connection and the budget it already uses for the counts, with no folder, file, journal, process or network use.
- REQ-8: The status line shall give each of the four numbers its own icon, so color is never the only signal.
- REQ-9: The status line shall leave the ledger file byte for byte the same.
- REQ-10: While the `cards` option is off, the status line shall show none of the four numbers.

## Acceptance

- AC-1 (REQ-1, REQ-2, REQ-3): a ledger with 3 `active` (one newest-marker waiting, two working), 2 `review`, 20 `draft`, 25 `ready`, 5 `hold`, 12 `done` renders `🗂️ 50 / ▶️ 2 / 🔃 3 / ✅ 12`.
  verify: `cargo test --locked --offline --test statusline_cards`
- AC-2 (REQ-3, REQ-4, REQ-5): the marker reading is pinned at its edges — a waiting marker newest, a working marker newer than it, a neutral note newer than it, no marker at all, a marker on the second line only, a note head cut at exactly 1000 characters, several waiting notes on one card counted once.
  verify: `cargo test --locked --offline --lib statusline::state::card_counts_tests`
- AC-3 (REQ-6, REQ-7, REQ-9): a failing note read — a deadline spent before or in the middle of the read — gives the three-number form, the working number never goes below zero, and the HOME tree is the same before and after a render.
  verify: `cargo test --locked --offline --lib statusline::`
- AC-4 (REQ-1, REQ-8, REQ-10): the golden for the four-number segment is checked in (plain, ansi, ansi256) and the existing card golden follows the new form.
  verify: `cargo test --locked --offline --test statusline_golden`

## Scope

- write: `src/statusline/state.rs`, `src/statusline/render.rs`, `tests/statusline_cards.rs`, `tests/statusline_golden.rs`, `tests/statusline_fixtures_test.py`, `tools/statusline_golden.py`, `tests/fixtures/statusline-claude/cases.json`, `tests/fixtures/statusline-claude/golden/b13-*`, `tests/fixtures/statusline-claude/golden/b14-*`, `tests/fixtures/statusline-claude/ledger/**`, `.moai/specs/SPEC-20261119-statusline-waiting-cards/**`
- read-only: `src/ledger/**` (schema, `state_counts`), `docs/rust-product-path.md`, `docs/codex-adapter-contract.md`
- forbidden: the real HOME and ledger, the network, every write to a ledger from the status line

## Plan

1. RED first: `card_counts_tests` gets the marker cases and `tests/statusline_cards.rs` the four-number case; both fail because `CardCounts` has no `waiting`.
2. `state.rs`: `CardCounts` gains `waiting`; `in_progress` keeps its name and means working. One more `SELECT` over `card_note` for the cards in `active`, reading only the first line of each note body (bounded length), newest first; a pure function maps a first line to waiting / working / neutral.
3. `render.rs`: the segment prints `🗂️ / ▶️ / 🔃 / ✅` in that order, `🔃` with its own theme color; when the notes could not be read, the segment is the same three numbers without `🔃`.
4. Goldens: regenerate `b13` and add `b14` with a checked-in ledger holding waiting cards. The prose in `docs/rust-product-path.md` follows in a later slice once its reservation is free.
5. Whole `--lib` once (statusline callers), fmt, fmt-ops, `clippy --all-targets`, `spec check`, then push the spec branch only (M-Flow).

## Risks

- A lane that writes a marker in another wording is read as working until it writes one of the known lines; the count then errs toward "working", which is the old behavior, never toward hiding a card.
- One extra `SELECT` costs render time; the notes read stops at the render budget and falls back to the three numbers (REQ-6).
- Every consumer that parses the card segment by position breaks on the new number; the grep of the repository for `🗂️`, `📥` and `🔥` readers is part of RUN.
- Until the documentation slice lands, `docs/rust-product-path.md` still describes the three-number form.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | the order is remaining, working, waiting, done with the icons `🗂️ ▶️ 🔃 ✅`; the three-number form keeps the same order and icons without `🔃` | the user's decision of 2026-11-19, quoted by the leader: "🔥 2 / 📮 3 / 📥 50 / ✅ 12 -> 🗂️ 50 / ▶️ 2 / 🔃 3 / ✅ 12"; it replaces the first revision's `🔥 📮 📥 ✅` (the lane's own choice) | change the icons or the order in `Renderer::cards` and the theme comments, then re-record the goldens |
| D-2 | a card in `review` is always waiting; an `active` card is waiting only by its newest marker | `review` is the hand-off state; `active` carries the pushed-and-waiting cards of the local bundle flow, which never enter `review` | count only markers |
| D-3 | notes without a known marker are neutral | the marker lines are the only ones lanes already write by rule; free text must not flip a card | add a marker later |
| D-4 | on a note-read failure the old three numbers show | the file's rule: a count that cannot be read is hidden, never shown as current | show `🔃 0` |
| D-5 | the documentation prose is a later slice | its file is reserved by another holder; a narrowed take is the only way to start now | widen the scope after the reservation is released |

## Result

- AC-1: met — 2026-11-19 이 worktree(팁 d25ecb1ba1)에서 `cargo test --locked --offline --test statusline_cards` 8 passed 0 failed(실제 바이너리가 3 active 중 대기 마커 노트가 있는 1장·2 review·20 draft·25 ready·5 hold·12 done 원장에서 `🗂️ 50 / ▶️ 2 / 🔃 3 / ✅ 12` 를 그려요; 원장 없음·빈 원장 `🗂️ 0 / ▶️ 0 / 🔃 0 / ✅ 0`·done 만의 사례·`cards` 옵션 끔·입력의 프로젝트 사례도 같은 형태).
- AC-2: met — `cargo test --locked --offline --lib statusline::` 116 passed 0 failed(`statusline::state::card_counts_tests` 의 마커 시험: 대기 마커 다섯 모양 각각, 더 새 일하는 마커·더 새 대기 마커의 번갈아 이김, 더 새 중립 노트는 건너뜀, 마커 없음, 둘째 줄·줄 중간의 마커와 둘째 줄에만 있는 `audit requested`·`verdict: PASS`·`verdict: FAIL` 는 마커 아님, draft·ready·hold·done 카드의 노트는 나누지 않음). RED 먼저: 필드만 있는 뼈대에서 새·바뀐 시험 15개가 단언에서 실패했어요. 변이 5(가장 오래된 노트가 결정·review 를 대기로 안 셈·일하는 마커 무시·본문 전체를 첫 줄로 읽음·읽기 실패에 대기 숫자 0 표시): M1·M2·M3·M5 는 첫 실행에서 exit 101, M4 는 첫 실행에서 생존(시험 공백 — 둘째 줄에만 있는 문구 사례 없음) 후 사례를 더해 exit 101. 복원 후 변이 문구 0건.
- AC-3: met — `statusline::state::card_counts_tests::a_failed_note_read_gives_the_three_number_form`: 지난 렌더 마감에 `waiting_cards` 가 `None`, 시간이 남으면 `Some(2)`; `statusline::render::tests::waiting_cards_get_their_own_number_and_none_gives_the_three_number_form`: `None` 은 `🗂️ 50 / ▶️ 28 / ✅ 12` 세 숫자(같은 순서·아이콘, `🔃` 없음). HOME 불변은 `statusline_cards` 의 `render_from` 이 렌더 앞뒤 HOME 트리(항목·바이트·수정 시각)를 비교하고, 단위 시험은 원장 파일 다이제스트와 HOME 트리를 비교해요.
- AC-4: met — `python3 -I tools/statusline_golden.py --bin target/debug/moai --check` checked 111 golden outputs, 실패 0(`b13-card-counts` 를 `🗂️ 50 / ▶️ 3 / 🔃 2 / ✅ 12` 로 다시 기록, `b14-card-counts-waiting` 3 변형과 원장 fixture `cards-5-50-12-waiting.sqlite` 신규: `🗂️ 50 / ▶️ 2 / 🔃 3 / ✅ 12`); `python3 -m unittest tests.statusline_fixtures_test tests.ci_scope_test tests.repo_scans_test` 76 OK. `statusline_golden` 시험 바이너리는 이 하니스를 그대로 돌려요.
- 이웃: `cargo fmt --check` 0, `rustfmt --check --edition 2021 src/ops/*.rs` 0, `cargo clippy --offline --locked --all-targets -- -D warnings` 0, `spec check` 0/0. `CardCounts` 를 쓰는 곳은 `src/statusline/` 뿐이에요(저장소 전체 grep).
- revision 2(사용자 결정 2026-11-19, 리더가 원문 인용 — D-1): 순서·아이콘을 `🗂️ 남은 / ▶️ 일하는 중 / 🔃 통합 대기 / ✅ 끝난` 으로 바꿨고 세 숫자 꼴도 같은 순서·아이콘(`🗂️ / ▶️ / ✅`)이에요. 시험을 새 문구로 먼저 바꿔 구 `Renderer::cards` 에서 10개가 단언 실패한 것을 확인한 뒤(RED) 구현을 바꿨어요. 이 라운드 변이 4(읽기 실패 폴백에 대기 숫자 0 표시·네 숫자 꼴의 남은/일하는 순서 뒤바꿈·세 숫자 꼴의 순서 뒤바꿈·대기 숫자에 일하는 중 색 사용) 모두 `cargo test --locked --offline --lib statusline::render` 에서 exit 101 로 죽고 복원 후 `src` 변경 0건. 이 라운드의 실행: lib `statusline::` 116 passed, `statusline_cards` 8 passed, golden `--check` 111 outputs ok, 파이썬 3 모듈 76 OK, fmt 0, fmt-ops 0, clippy `--all-targets` 0, `spec check` 0/0.
- 돌리지 않은 것: lib 전체(통합 전 리더 점검 세트), Linux·Windows. 후속 조각: `docs/rust-product-path.md` 항목 2 의 문구(다른 예약이 잡고 있어 이 SPEC 범위 밖, D-5).
