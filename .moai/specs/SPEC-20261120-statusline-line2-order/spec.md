---
id: SPEC-20261120-statusline-line2-order
title: "Statusline line 2: git counts, card counts, repository, branch, with no slash inside a group"
status: draft
revision: 1
created: 2026-11-20
---

# Statusline line 2: git counts, card counts, repository, branch, with no slash inside a group

## Intent

사용자 지시(2026-11-19T21:0xZ, 리더 대화, 카드 C-137)로 상태줄 2번 줄의 순서와 묶음 표기를 바꿔요. 지금 줄은 `📡 repo | ⬇️ 0 / 🔀 0 | 🗂️ 50 / ▶️ 5 / 🔃 15 / ✅ 65 | 🅱️ branch` 이고, 사용자가 적은 예시는 `⬇️ 0  🔀 0 | 🗂️ 50 ▶️ 5 🔃 15 ✅ 65 | 📡  modu-ai/moai-adk | 🅱️  v2026.11.0` 이고, 리더가 정한(결정 D-1, 카드 노트 3570) 결과 줄은 `⬇️ 0 🔀 0 | 🗂️ 50 ▶️ 5 🔃 15 ✅ 65 | 📡 modu-ai/moai-adk | 🅱️ v2026.11.0` 이에요. 순서는 git 숫자 묶음 → 카드 숫자 묶음 → 저장소 → 가지, 묶음 사이는 ` | `, 묶음 안 슬래시는 없어요.

바뀌지 않는 것: 무엇이 언제 빠지는지(저장소 없으면 git 숫자 없음 등), C-121 의 대기 규칙과 표지, 숫자 값, 색, 줄 1과 줄 3.

읽어서 확인한 사실(rule 10, 명령으로):
- 카드 본문의 "40칸에서 줄이는 순서"는 코드에 없어요. 렌더러는 아무것도 줄이지 않고 `COLUMNS` 도 읽지 않아요(`src/statusline/render.rs` 의 "never shortens" 주석, `docs/rust-product-path.md` 항목 3 끝, `tests/statusline_tag_order.rs` 가 80칸과 40칸 출력이 바이트로 같음을 증명). 그래서 "줄임 순서 그대로"는 "40칸에서도 80칸과 바이트가 같다"로 읽어요(REQ-6).
- 예시 문자열의 `📡` 와 `🅱️` 뒤는 칸이 둘이고 git 두 항목 사이도 둘이었는데, 리더 설명은 "이모지 뒤 한 칸"이었어요. 리더가 정했어요(D-1, 카드 노트 3570): 모든 아이콘 뒤 한 칸, 묶음 안 항목 사이 한 칸(git 두 항목 사이도), 묶음 사이 ` | `. 예시의 두 칸은 변이 선택자(U+FE0F)가 붙은 아이콘의 터미널 표시로 봐요.

## Non-goals

- 줄 1·줄 3, 색, 숫자의 뜻, `statusline.yaml` 의 옵션 이름을 바꾸기.
- 40칸에서 무언가를 줄이는 새 동작 만들기(지금 없어요).
- `tests/fixtures/statusline/` 의 옛 Go 기록 고치기(읽기 전용 보관본).
- `docs/rust-product-path.md` 항목 2의 글: 그 문서는 아직 제가 쥔 C-131 이 예약하고 있어서 이 조각에 없고, C-131 이 닫힌 뒤 별도 조각이에요(D-4).

## Requirements

- REQ-1: The git counts group `⬇️ <issues> 🔀 <prs>` shall stand first on line 2, its two items separated by one space and no slash.
- REQ-2: The card counts group `🗂️ <remaining> ▶️ <working> 🔃 <waiting> ✅ <done>` shall follow the git counts group, its items separated by one space and no slash.
- REQ-3: While the waiting notes cannot be read within the render budget, the card counts group shall keep its place and order with the three items `🗂️`, `▶️` and `✅`.
- REQ-4: The repository `📡 <owner>/<name>` shall follow the card counts group.
- REQ-5: The branch `🅱️ <branch>` with its behind and dirty counts shall follow the repository.
- REQ-6: The groups of line 2 shall be separated by ` | `.
- REQ-7: The renderer shall separate every icon of line 2 from its text, and the items inside a group from each other, by exactly one space.
- REQ-8: The renderer shall give the same bytes under `COLUMNS` 40 and `COLUMNS` 80 for line 2, dropping nothing at either width.
- REQ-9: Every item that line 2 leaves out today shall stay left out for the same reasons, together with its separator.
- REQ-10: The renderer shall count the cards of a ledger as before, with the C-121 waiting rule and markers unchanged.
- REQ-11: The golden files and the fixtures and every test that reads line 2 shall carry the new shape.
- REQ-12: Every `tests/statusline*.rs` target shall pass.

## Acceptance

- AC-1 (REQ-1, REQ-2, REQ-4…REQ-7): the unit tests of the renderer pin the whole line for the full case, `⬇️ 7 🔀 3 | 🗂️ 50 ▶️ 5 🔃 15 ✅ 65 | 📡 owner/repo | 🅱️ main ↓2 +3 | 🔗 #42 pending`, and for each pair of groups the order git counts, cards, repository, branch.
  verify: `cargo test --locked --offline --lib statusline::render`
- AC-2 (REQ-1, REQ-2, REQ-4…REQ-7, REQ-11): the golden `b14-card-counts-waiting` carries, on line 2, the shape of the leader's result line with its own numbers (`⬇️ 46 🔀 0 | 🗂️ 50 ▶️ 2 🔃 3 ✅ 12 | 📡 anthropics/claude-code | 🅱️ main | 🔗 #1234 pending`), plain, ansi and ansi256; the checked-in file and the binary agree.
  verify: `cargo test --locked --offline --test statusline_golden`
- AC-3 (REQ-3, REQ-9, REQ-10): a registered ledger without cards shows `🗂️ 0 ▶️ 0 🔃 0 ✅ 0`; the cases without a repository, without a branch, with `git_branch` off, with `cards` off and without forge counts keep their items and drop what they dropped before; the counts are the same numbers as before for the same ledgers.
  verify: `cargo test --locked --offline --test statusline_cards`
- AC-4 (REQ-8): line 2 is byte-equal under `COLUMNS` 80 and 40, ko and en, plain and truecolor.
  verify: `cargo test --locked --offline --test statusline_tag_order`
- AC-5 (REQ-11, REQ-12): every statusline rendering target passes: `statusline_cards`, `statusline_golden`, `statusline_tag_order`, `statusline_worktree_label`, `statusline_git_env`, `statusline_counts_refresh`, `statusline_home_state`, `statusline_guard_false_positive`, and `python3 -m unittest tests.statusline_fixtures_test`.
  verify: `cargo test --locked --offline --test statusline_cards --test statusline_golden --test statusline_tag_order --test statusline_worktree_label --test statusline_git_env --test statusline_counts_refresh --test statusline_home_state --test statusline_guard_false_positive`

## Scope

- write: `src/statusline/render.rs`, `src/statusline/state.rs` and `src/statusline/counts.rs` (doc comments only), `tests/statusline_*.rs`, `tests/statusline_fixtures_test.py`, `tests/fixtures/statusline-claude/cases.json` and `golden/**`, `.moai/specs/SPEC-20261120-statusline-line2-order/**`
- read-only: `tests/fixtures/statusline/**` (historical Go archive), `docs/rust-product-path.md` (held by C-131; its statusline item 2 follows in a later piece), `docs/design-review/**`
- forbidden: the real HOME and ledger, the network, a change of the counts or their markers

## Plan

1. SPEC first (this file); the one open question (D-1) went to the leader and is answered.
2. RED: change the whole-line assertions of the renderer unit tests and `tests/statusline_cards.rs` to the new shape; they fail on the current code.
3. GREEN: `git_line`, `repo_branch` and `cards` in `src/statusline/render.rs` write the groups in the new order; the two spacings live in one place each.
4. Regenerate the goldens with `tools/statusline_golden.py --write`; read the diff of every file: only line 2 changes.
5. Run every `tests/statusline*.rs` target and the unit tests of `src/statusline`; mutants after committing (at most five): swap the card group and the repository, put the slash back in the git group, drop the second space of the git group, move the forge counts after the cards, change the spacing after `📡`.
6. fmt, fmt-ops, `clippy --all-targets`, `spec check`, the scan tests, the registered checks, `--bins`, trailer-check, tip `merge-tree`, an independent audit, then the READY line.

## Risks

- A golden is three files per case (plain, ansi, ansi256); a regenerated file that changes something besides line 2 is a bug of this change, so the diff of every golden is read.
- Keeping `🔃` out when the notes cannot be read changes the group length; the three-item shape follows the same rule as before.
- The goldens of cases that show no forge counts and no cards (branch only, `b01`, `b02`, `d01`–`d04`, `b10`) stay byte for byte what they were: with one space after every icon the order change does not reach them.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | one space after every icon and between the items of a group (git counts too), ` | ` between groups | leader decision, card note 3570: the two spaces of the user's example are the terminal's drawing of the variation-selector icons; the user may reassign | the four gap constants at the top of `src/statusline/render.rs`, then the goldens are rewritten |
| D-2 | "40 columns" means byte equality with 80 columns | the renderer never shortens (read in code, docs and a test) | add a width reading later as its own card |
| D-3 | the forge counts stay shown only with a repository name, the repository and branch only with a git branch, the cards on their own without both | these are the rules today; this change reorders, it does not add or drop items | none |
| D-4 | the documentation of item 2 is a later piece | `docs/rust-product-path.md` is reserved by C-131, which is READY and waits for the release PR | a `card scope` widening once C-131 is done |

## Result

(not yet)
