---
id: SPEC-20261119-statusline-worktree-label
title: "The status line opens line 3 with the worktree as WT-c<card>-<name>"
status: draft
revision: 1
created: 2026-11-19
---

# The status line opens line 3 with the worktree as WT-c<card>-<name>

## Intent

상태줄은 지금 worktree 이름을 2번째 줄 끝에 `🌳 moai-adk-c103` 처럼 보여 줘요. 폴더 이름(프로젝트 이름 + 번호)은 그 worktree 가 무슨 일을 하는지 말해 주지 않고, 줄 끝이라 눈에 잘 안 띄어요. 2026-11-19 사용자 지시(원문): "🌳 moai-adk-c103 | 🔋 CTX: ██░░░░░░░░ 28% | 🔋 Week: █████░░░░░ 55% 이렇게 worktree를 3번째 줄 맨앞으로 변경하고, 워크트리는 WT-{card-no}-worktree-name 이렇게 표기를 하자." 이어서 같은 날 결정(원문): "🌳 WT-103-moai-adk-c103 -> 🌳 WT-c103-{워크트리가 뭔지 알수 있도록 네이밍} 이런식으로 앞으로 사용하도록 하자."

이 SPEC 은 두 가지를 해요. (1) worktree 조각을 3번째 줄(막대 줄)의 맨 앞으로 옮겨요. (2) 조각을 `🌳 WT-c<카드 번호>-<일을 알 수 있는 이름>` 으로 써요. 카드 번호는 그 worktree 를 예약으로 잡고 있는 카드에서, 이름은 그 예약의 가지 이름에서 읽어요 — 둘 다 상태줄이 이미 여는 읽기 전용 원장에서요.

## Non-goals

- worktree 폴더 이름 바꾸기, 새 worktree 를 만들 때의 이름 규칙(`~/.moai/worktrees/<project>-<id>`, 2026-11-16 사용자 결정): 이 SPEC 은 표시만 바꿔요. "앞으로 사용하도록" 이 새 worktree 이름 규칙까지 뜻하는지는 사용자 확인이 필요한 별도 일이에요(Decisions D-7).
- 원장 쓰기, 새 명령, 새 표 — 상태줄은 읽기만 해요.
- 메인 체크아웃(linked worktree 가 아닌 곳)에서의 표시: 그곳에서는 예전처럼 조각이 없어요.
- `docs/rust-product-path.md` 와 `docs/codex-adapter-contract.md` 의 문구 갱신: 그 파일은 다른 예약이 잡고 있어 후속 조각으로 해요.
- Codex 상태줄(원장을 읽는 항목이 없어 해당 없음).

## Requirements

- REQ-1: While the session runs in a linked worktree and the `worktree` option is on, the status line shall write the worktree segment at the front of line 3, before the `CTX:` and `Week:` bars, and not on line 2.
- REQ-2: When a held reservation of the project names the folder of the session's worktree, the status line shall write the segment as `🌳 WT-c<card>-<name>`, where `<card>` is the number of the card that holds it. · failure: none — 이 모양이 안 되는 경우(예약 없음·원장 못 읽음)는 REQ-5 가 정해요.
- REQ-3: The status line shall take `<card>` from the card with the lowest number when more than one held reservation names the folder.
- REQ-4: The status line shall take `<name>` from the branch of that reservation: the part after the eight-digit date and a dash in its last path segment (`spec/20261119-statusline-waiting-cards` gives `statusline-waiting-cards`), else the folder name without its `<project>-` prefix and without the card number, else nothing, and then the segment is `🌳 WT-c<card>`.
- REQ-5 (REQ-2): IF no held reservation names the folder, or the ledger cannot be read within the render budget, THEN the status line shall write the worktree name the host gave, with no `WT-` prefix and no card number, at the same place on line 3.
- REQ-6: The status line shall show at most 32 characters of `<name>`, cut with `…` after the 31st.
- REQ-7: The status line shall drop control characters from `<name>`.
- REQ-8: The status line shall read the reservations with the connection and the budget it already uses for the card counts, with no folder, file, journal, process or network use.
- REQ-9: The status line shall leave the ledger file byte for byte the same.
- REQ-10: While the session is not in a linked worktree, the status line shall write no worktree segment.
- REQ-11: While the `worktree` option is off, the status line shall write the worktree segment on no line.

## Acceptance

- AC-1 (REQ-1, REQ-10, REQ-11): the real binary, in a linked worktree, prints the segment first on line 3 and nothing on line 2; in a main checkout it prints none; with the `worktree` option off none.
  verify: `cargo test --locked --offline --test statusline_worktree_label`
- AC-2 (REQ-2, REQ-3, REQ-4, REQ-6, REQ-7): the label is pinned at its edges — a spec branch, a branch of another shape with a folder name that keeps a rest, one that keeps nothing, two cards holding one folder (the lowest), a name of exactly 32 and of 33 characters, a name with control characters.
  verify: `cargo test --locked --offline --lib statusline::state::worktree_label_tests`
- AC-3 (REQ-5, REQ-8, REQ-9): with no reservation for the folder, with a released one, and with a read that runs out of budget, the segment is the host's name alone; the ledger file digest and the HOME tree are the same before and after.
  verify: `cargo test --locked --offline --lib statusline::`
- AC-4 (REQ-1, REQ-5, REQ-11): the golden files that carried the worktree on line 2 carry it first on line 3 (plain, ansi, ansi256), and a case with a held reservation carries the label.
  verify: `cargo test --locked --offline --test statusline_golden`

## Scope

- write: `src/statusline/state.rs`, `src/statusline/render.rs`, `tests/statusline_worktree_label.rs`, `tests/statusline_cards.rs`, `tests/statusline_golden.rs`, `tests/statusline_fixtures_test.py`, `tools/statusline_golden.py`, `tests/fixtures/statusline-claude/cases.json`, `tests/fixtures/statusline-claude/golden/**`, `tests/fixtures/statusline-claude/ledger/**`, `.moai/specs/SPEC-20261119-statusline-worktree-label/**`
- read-only: `src/ledger/**` (schema, `reservation` table), `docs/rust-product-path.md`, `docs/codex-adapter-contract.md`
- forbidden: the real HOME and ledger, the network, every write to a ledger from the status line

## Plan

1. RED first: `tests/statusline_worktree_label.rs` runs the real binary against a temporary HOME whose ledger holds a reservation for a temporary linked worktree, and fails because the segment is still at the end of line 2 and has the host's name.
2. `state.rs`: `Status` gains the label parts (`card`, `name`); one more read of the `reservation` table for the folder, by the same read-only connection (`SELECT card_id, branch FROM reservation WHERE worktree = ? AND status IN ('held','held_for_review','expired_candidate') ORDER BY card_id LIMIT 1`); a pure function builds the name from the branch, the folder name and the project id. The folder is the nearest ancestor of `workspace.current_dir` (else `cwd`) that holds a `.git` entry, compared as a canonical path.
3. `render.rs`: the worktree segment leaves `git_line` and opens `bars_line`; the label text is built there; `worktree` still switches it.
4. Goldens: every case that showed `🌳` on line 2 is recorded again; a new case carries a ledger fixture with a held reservation. The documents follow in a later slice.
5. Whole `--lib` once, fmt, fmt-ops, `clippy --all-targets`, `spec check`, then push the spec branch only (M-Flow).

## Risks

- The folder recorded in a reservation is the path the holder typed (`card take --worktree`); a symlinked or differently spelled path would not match the session's folder and the segment would fall back to the host's name (REQ-5), never to a guess. Both sides are compared as canonical paths.
- Line 3 gets longer by up to about 40 characters; the cap of 32 characters on `<name>` bounds it. The renderer does not read `COLUMNS`, so a very narrow terminal may still wrap it.
- Roughly a dozen golden cases move one segment from line 2 to line 3; the check of the golden files is part of RUN, one case at a time.
- A card whose reservation names a worktree another session also uses shows that card's number in both sessions; this is the ledger's truth, not the session's.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | the card number is written `c<number>` in lower case (`WT-c103-…`) | user decision 2026-11-19: "WT-c103-{워크트리가 뭔지 알수 있도록 네이밍}" | change the format string in one place and the goldens |
| D-2 | the card comes from a held reservation that names the worktree folder, not from digits in the folder name | the ledger is the source of truth for who works where; folder names such as `moai-adk-107-hooks` carry a number only by habit | parse the folder name |
| D-3 | the name comes from the reservation's branch slug, then from the folder name minus project prefix and card number | the user wants a name that shows the work; a spec branch is named after the work (`spec/<date>-<slug>`) | take the folder name only |
| D-4 | two cards holding one folder show the lowest number | deterministic, and the lowest is the older work | show the newest, or both as `c103+c104` |
| D-5 | no card means the host's worktree name alone, with no `WT-` | the leader's reading of the instruction (card criterion 1): a session without a card shows the worktree name alone; nothing is invented | show `WT-<name>` |
| D-6 | `<name>` is cut to 32 characters | line 3 also holds two bars; a label grows with the branch slug | raise or drop the cap |
| D-7 | the new-worktree naming rule is not changed here | "앞으로 사용하도록" may also mean new worktree names; that changes the 2026-11-16 location and name rule and needs the user's answer | a later SPEC once the user chooses |
| D-8 | the documents are a later slice | their file is reserved by another holder | write them when it is free |

## Result

(draft: filled in with evidence at RUN.)
