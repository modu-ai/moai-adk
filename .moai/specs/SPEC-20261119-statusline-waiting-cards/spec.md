---
id: SPEC-20261119-statusline-waiting-cards
title: "The status line tells cards being worked on from cards waiting for integration"
status: draft
revision: 1
created: 2026-11-19
---

# The status line tells cards being worked on from cards waiting for integration

## Intent

상태줄의 카드 수 `🔥 진행 중` 은 `active` + `review` 를 한 숫자로 세요. 2026-11-19 사용자가 "진행 중 수가 34 에서 안 줄어드는 이유"를 물었을 때, 그중 약 26장은 이미 push 한 head 가 리더의 묶음(bundle)이나 감사를 기다리는 카드였어요 — 숫자만 봐서는 병목이 레인이 아니라 통합이라는 것이 안 보였어요. 이 SPEC 은 같은 읽기 전용 원장 읽기에서 노트의 마커 줄을 한 번 더 읽어, 일하고 있는 카드와 push 하고 통합을 기다리는 카드를 두 숫자로 나눠 보여 줘요.

## Non-goals

- 원장 쓰기, 새 상태 값(`card_transition` 표 변경), 새 명령 — 상태줄은 읽기만 해요.
- `📥`·`✅` 의 기존 뜻 변경(`📥` = draft + ready + hold, `✅` = done 그대로).
- Codex 상태줄(원장을 읽는 항목이 없어 해당 없음 — 기존 계약 그대로).
- 노트 마커의 새 형식 도입: 레인이 이미 쓰는 줄(`LANE-READY-INTEGRATION`, `감사 요청`, `AUDIT … verdict:`, `ACK`, `PLAN`)만 읽어요.
- `docs/rust-product-path.md` 의 항목 2 문구 갱신: 그 파일은 다른 예약이 잡고 있어 이 SPEC 의 범위 밖이고, 예약이 풀린 뒤 후속 조각으로 해요.

## Requirements

- REQ-1: When the status line shows the card counts, it shall show four numbers, `🔥 <working> / 📮 <waiting> / 📥 <remaining> / ✅ <done>`, where `<working>` plus `<waiting>` equals the cards in `active` or `review`.
- REQ-2: While a card is in `review`, the status line shall count it as waiting.
- REQ-3: While a card is in `active`, the status line shall count it as waiting only when the newest of its notes whose first line starts with a marker is a waiting marker, and as working in every other case.
- REQ-4: When a note's first line starts with `LANE-READY-INTEGRATION`, or starts with `LANE-DONE` and carries `audit requested` or `감사 요청`, or starts with `감사 요청`, or starts with `AUDIT` and carries `verdict: PASS`, the status line shall read it as a waiting marker. · failure: none — 첫 줄이 이 모양이 아니면 마커가 아니라서 읽는 쪽에 실패 경로가 없어요.
- REQ-5: When a note's first line starts with `ACK` or `PLAN`, or starts with `AUDIT` and carries `verdict: FAIL`, the status line shall read it as a working marker. · failure: none — 첫 줄이 이 모양이 아니면 중립 노트라 읽음이 바뀌지 않아요.
- REQ-6 (REQ-1): IF the notes cannot be read within the render budget, THEN the status line shall show the three-number form `🔥 <active + review> / 📥 <remaining> / ✅ <done>` and no `📮`.
- REQ-7: The status line shall read the notes with the connection and the budget it already uses for the counts, with no folder, file, journal, process or network use.
- REQ-8: The status line shall give each of the four numbers its own icon, so color is never the only signal.
- REQ-9: The status line shall leave the ledger file byte for byte the same.
- REQ-10: While the `cards` option is off, the status line shall show none of the four numbers.

## Acceptance

- AC-1 (REQ-1, REQ-2, REQ-3): a ledger with 3 `active` (one newest-marker waiting, two working), 2 `review`, 20 `draft`, 25 `ready`, 5 `hold`, 12 `done` renders `🔥 2 / 📮 3 / 📥 50 / ✅ 12`.
  verify: `cargo test --locked --offline --test statusline_cards`
- AC-2 (REQ-3, REQ-4, REQ-5): the marker reading is pinned at its edges — a waiting marker newest, a working marker newer than it, a neutral note newer than it, no marker at all, a marker on the second line only.
  verify: `cargo test --locked --offline --lib statusline::state::card_counts_tests`
- AC-3 (REQ-6, REQ-7, REQ-9): a failing note read gives the three-number form, and the HOME tree is the same before and after a render.
  verify: `cargo test --locked --offline --test statusline_cards`
- AC-4 (REQ-1, REQ-8, REQ-10): the golden for the four-number segment is checked in (plain, ansi, ansi256) and the existing card golden follows the new form.
  verify: `cargo test --locked --offline --test statusline_golden`

## Scope

- write: `src/statusline/state.rs`, `src/statusline/render.rs`, `tests/statusline_cards.rs`, `tests/statusline_golden.rs`, `tests/statusline_fixtures_test.py`, `tools/statusline_golden.py`, `tests/fixtures/statusline-claude/cases.json`, `tests/fixtures/statusline-claude/golden/b13-*`, `tests/fixtures/statusline-claude/golden/b14-*`, `tests/fixtures/statusline-claude/ledger/**`, `.moai/specs/SPEC-20261119-statusline-waiting-cards/**`
- read-only: `src/ledger/**` (schema, `state_counts`), `docs/rust-product-path.md`, `docs/codex-adapter-contract.md`
- forbidden: the real HOME and ledger, the network, every write to a ledger from the status line

## Plan

1. RED first: `card_counts_tests` gets the marker cases and `tests/statusline_cards.rs` the four-number case; both fail because `CardCounts` has no `waiting`.
2. `state.rs`: `CardCounts` gains `waiting`; `in_progress` keeps its name and means working. One more `SELECT` over `card_note` for the cards in `active`, reading only the first line of each note body (bounded length), newest first; a pure function maps a first line to waiting / working / neutral.
3. `render.rs`: the segment prints `📮` between `🔥` and `📥` with its own theme color; when the notes could not be read, the segment is the old three numbers.
4. Goldens: regenerate `b13` and add `b14` with a checked-in ledger holding waiting cards. The prose in `docs/rust-product-path.md` follows in a later slice once its reservation is free.
5. Whole `--lib` once (statusline callers), fmt, fmt-ops, `clippy --all-targets`, `spec check`, then push the spec branch only (M-Flow).

## Risks

- A lane that writes a marker in another wording is read as working until it writes one of the known lines; the count then errs toward "working", which is the old behavior, never toward hiding a card.
- One extra `SELECT` costs render time; the notes read stops at the render budget and falls back to the three numbers (REQ-6).
- Every consumer that parses the card segment by position breaks on the new number; the grep of the repository for `📥` readers is part of RUN.
- Until the documentation slice lands, `docs/rust-product-path.md` still describes the three-number form.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | the waiting icon is `📮`, placed between `🔥` and `📥` | the request asks for a split of the in-progress number; the mailbox reads as "posted, waiting to be picked up" and does not collide with the used `🔥 📥 ✅ 🅱️ 🌳 🔗` | change the icon in one theme field and the goldens |
| D-2 | a card in `review` is always waiting; an `active` card is waiting only by its newest marker | `review` is the hand-off state; `active` carries the pushed-and-waiting cards of the local bundle flow, which never enter `review` | count only markers |
| D-3 | notes without a known marker are neutral | the marker lines are the only ones lanes already write by rule; free text must not flip a card | add a marker later |
| D-4 | on a note-read failure the old three numbers show | the file's rule: a count that cannot be read is hidden, never shown as current | show `📮 0` |
| D-5 | the documentation prose is a later slice | its file is reserved by another holder; a narrowed take is the only way to start now | widen the scope after the reservation is released |

## Result

(RUN fills this.)
