---
id: SPEC-20261123-self-check-rust-only
title: "self-check.md 의 테스트 문장이 Rust 전용 규칙을 말하고 repo-scans 는 cargo 검사만 돈다고 말함"
status: draft
revision: 1
created: 2026-11-23
---

# self-check.md 의 테스트 문장이 Rust 전용 규칙을 말하고 repo-scans 는 cargo 검사만 돈다고 말함

## Intent

`skills/moai/references/internal/self-check.md` 의 `## Tests` 목록에서 두 문장이 저장소 규칙과 어긋나요. 3번 항목은 READY 전에 돌릴 묶음에 "the Python tests of the area" 를 넣어서 `CLAUDE.md` 의 "Tests are Rust only" 와 맞지 않아요. 5번 항목은 `moai repo-scans --refs <tip> HEAD` 가 "repo-wide scans" 를 나열한다고만 말해서, 나열된 표에는 `python3 -m unittest …` 줄이 7개 있는데 그 줄을 손으로 돌려야 하는지 알 수 없어요. 착지 명령은 표에서 cargo 검사만 요구해요(`src/card/land/scans.rs` 의 `cargo_only`).

이 SPEC 은 두 문장을 고치고, 문장의 뜻이 뒤집히면 실패하는 고정 테스트를 `tests/next_flow_skill_scan.rs` 에 새 함수로 더해요. C-400 이 같은 파일을 고치고 있어서 테스트는 기존 함수를 건드리지 않고 파일 끝에 새 함수로만 둬요.

## Non-goals

- `tools/repo_scans.json` 표의 Python 줄을 지우거나 `moai repo-scans` 의 출력을 바꾸는 일(표는 그대로, 그 줄의 처리는 다른 카드의 일).
- `self-check.md` 의 6번 항목(Python)과 머리말의 `--full-python`, `py-*` 항목 설명을 고치는 일(D-1).
- `next-flow.md` 의 repo-scans 문장(C-400 의 일)과 `lane-rules.md`.
- `tools/self_check.sh` 의 동작.

## Requirements

- REQ-1: The third item of the `## Tests` list in `skills/moai/references/internal/self-check.md` shall name only cargo commands and `cfg(target_os)` tests as the set run before READY, with no Python test in it. · failure: none — AC-1 pins it.
- REQ-2: The fifth item of the same list shall say that `moai repo-scans --refs <tip> HEAD` lists the repo-wide scans and that only the cargo checks of that list are run, the Python lines of the table never by hand. · failure: none — AC-2 pins it.
- REQ-3: The pin tests shall fail for a text that keeps the words of the two items and reverses their meaning. · failure: a surviving reversed text is a missing test, repaired before READY.
- REQ-4: `skills/moai/references/source-manifest.json` shall be rewritten by `tools/source_manifest.py write`, pass `tools/source_manifest.py check` and keep every package file at most 65,536 bytes. · failure: none — AC-4 runs both.

## Acceptance

- AC-1 (REQ-1): item 3 of the `## Tests` list names the cargo commands (`--bin moai`, `--bin moai-cli`, `--lib`, `--test <name>`), the `tools/repo_scans.json` checks that are cargo commands and the macOS-only tests, and holds no `Python` or `unittest`.
  verify: `cargo nextest run --locked --offline --test next_flow_skill_scan -E 'test(self_check_tests_item_names_no_python_test)'`
- AC-2 (REQ-2): item 5 keeps its `moai repo-scans --refs <tip> HEAD` command and says that only the cargo checks of the list are run and that its Python lines are not run by hand.
  verify: `cargo nextest run --locked --offline --test next_flow_skill_scan -E 'test(self_check_repo_scans_item_runs_cargo_checks_only)'`
- AC-3 (REQ-3): a copy of each item with the same words and the reversed meaning (item 3 asks for the Python tests of the area again, item 5 says to run every check of the list or the Python lines too) is rejected by the same checking function the pin tests use.
  verify: `cargo nextest run --locked --offline --test next_flow_skill_scan -E 'test(self_check_pins_reject_reversed_sentences)'`
- AC-4 (REQ-4): the manifest tool writes and then checks without a difference, and the skill package passes its inspection.
  verify: `python3 -I tools/source_manifest.py check`
- AC-5 (REQ-1, REQ-2, REQ-3, REQ-4): the SPEC passes the checker.
  verify: `moai spec check SPEC-20261123-self-check-rust-only`

## Scope

- write: `.moai/specs/SPEC-20261123-self-check-rust-only/**`, `skills/moai/references/internal/self-check.md` (items 3 and 5 only), `skills/moai/references/source-manifest.json` (rewritten by the tool), `tests/next_flow_skill_scan.rs` (new test functions at the end of the file; no existing function changes)
- read-only: every other file; `tools/repo_scans.json`, `src/card/land/scans.rs` and `tools/self_check.sh` are read for the facts above and not edited
- tier: 1 (a skill reference text and a text pin test); the floor list of `references/audit.md` is checked with `moai bundle check`
- history: found by the audit of C-400 (N3)

## Plan

1. SPEC first (this file); approval comes from the user through the leader (`LANE-BLOCKED kind=decision`).
2. RED: the three new test functions at the end of `tests/next_flow_skill_scan.rs`; they fail on the present text for the stated reason.
3. Edit items 3 and 5 of `self-check.md`.
4. `python3 -I tools/source_manifest.py write`, then `check`; `moai package-inspect skills/moai`.
5. Guard mutants: reverse each item's meaning keeping its words; each must turn a pin red.
6. `moai check affected` once at READY, unwrapped.

## Risks

- C-400 edits the same test file and `source-manifest.json`: new functions only, and the manifest is rewritten by the tool at the merge, so a conflict is resolved by running the tool again, not by hand.
- The pin reads the item text by its list number; a renumbered list breaks it. The test finds an item by its first words, not by its number.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | only items 3 and 5 change; item 6 (Python unittest files and `tests/skill_references_test.py`) and the `--full-python`, `py-source-manifest`, `py-repo-scans` lines of the header stay and become a follow-up card | the card names two sentences, and those lines describe `tools/self_check.sh` items that exist today; changing them changes what that tool is said to do | widen the scope to item 6 and the header in this SPEC |
| D-2 | the sentence says only cargo checks of the `repo-scans` list are run, and does not say the table has no Python lines | the table still has 7 `python3` lines (read 2026-11-23) while the landing keeps cargo checks only (`cargo_only`) | name the Python lines in the sentence |
| D-3 | the pin tests are new functions at the end of the file | C-400 edits the same file; separate functions merge without a conflict | add to existing functions |

## Result

- (draft: filled in when the work is done)
