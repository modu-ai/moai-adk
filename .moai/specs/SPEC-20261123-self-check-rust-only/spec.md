---
id: SPEC-20261123-self-check-rust-only
title: "self-check.md 의 테스트 문장이 Rust 전용 규칙을 말하고 repo-scans 는 cargo 검사만 돈다고 말하며, next-flow.md 의 감사 브리프 문장이 스캔 테스트 이름과 PASS 영수증 단계를 다시 말함"
status: draft
revision: 2
created: 2026-11-23
---

# self-check.md 의 테스트 문장이 Rust 전용 규칙을 말하고 repo-scans 는 cargo 검사만 돈다고 말하며, next-flow.md 의 감사 브리프 문장이 스캔 테스트 이름과 PASS 영수증 단계를 다시 말함

## Intent

`skills/moai/references/internal/self-check.md` 의 `## Tests` 목록에서 두 문장이 저장소 규칙과 어긋나요. 3번 항목은 READY 전에 돌릴 묶음에 "the Python tests of the area" 를 넣어서 `CLAUDE.md` 의 "Tests are Rust only" 와 맞지 않아요. 5번 항목은 `moai repo-scans --refs <tip> HEAD` 가 "repo-wide scans" 를 나열한다고만 말해서, 나열된 표에는 `python3 -m unittest …` 줄이 7개 있는데 그 줄을 손으로 돌려야 하는지 알 수 없어요. 착지 명령은 표에서 cargo 검사만 요구해요(`src/card/land/scans.rs` 의 `cargo_only`).

이 SPEC 은 두 문장을 고치고, 문장의 뜻이 뒤집히면 실패하는 고정 테스트를 `tests/next_flow_skill_scan.rs` 에 새 함수로 더해요.

개정 2(사용자 승인 때 리더가 덧붙인 범위): 같은 표를 말하는 `skills/moai/references/flows/next-flow.md` 의 감사 브리프 문장이 `SPEC-20261123-doc-sentence-pins` 로 cargo 검사만 돈다는 뜻을 얻는 대신 두 뜻을 잃었어요: 감사 브리프가 스캔 테스트의 이름을 대야 한다는 것, 그리고 각 스캔을 PASS 영수증의 `scan:<id>` 단계로 적는다는 것. 이 SPEC 이 두 뜻을 되살리고, 파일이 `7974` 바이트(그 SPEC 이 착지한 크기) 아래에 머물게 문장을 줄여 써요. 되살린 문장에도 뜻이 뒤집히면 실패하는 고정 테스트를 더해요.

## Non-goals

- `tools/repo_scans.json` 표의 Python 줄을 지우거나 `moai repo-scans` 의 출력을 바꾸는 일(표는 그대로, 그 줄의 처리는 다른 카드의 일).
- `self-check.md` 의 6번 항목(Python)과 머리말의 `--full-python`, `py-*` 항목 설명을 고치는 일(D-1, 사용자가 3·5번만으로 승인).
- `lane-rules.md` 와 `next-flow.md` 의 repo-scans 문장 밖의 줄을 바꾸는 일(바이트를 맞추려고 같은 파일의 두 곳을 줄이는 일만 D-5 로 해요).
- `tools/self_check.sh` 의 동작.

## Requirements

- REQ-1: The third item of the `## Tests` list in `skills/moai/references/internal/self-check.md` shall name only cargo commands and `cfg(target_os)` tests as the set run before READY, with no Python test in it. · failure: none — AC-1 pins it.
- REQ-2: The fifth item of the same list shall say that `moai repo-scans --refs <tip> HEAD` lists the repo-wide scans and that only the cargo checks of that list are run, the Python lines of the table never by hand. · failure: none — AC-2 pins it.
- REQ-3: The pin tests shall fail for a text that keeps the words of the two items and reverses their meaning. · failure: a surviving reversed text is a missing test, repaired before READY.
- REQ-4: `skills/moai/references/source-manifest.json` shall be rewritten by `tools/source_manifest.py write`, pass `tools/source_manifest.py check` and keep every package file at most 65,536 bytes. · failure: none — AC-4 runs both.
- REQ-5: The `Audit brief` bullet of `skills/moai/references/flows/next-flow.md` shall tell the session that briefs an audit to name the scan tests that `moai repo-scans --refs <tip> <head>` picks and to record each as a receipt step `scan:<id>` of the PASS, beside the cargo-only reading it already has. · failure: none — AC-6 pins it.
- REQ-6: The flow reference shall stay at most 7974 bytes. · failure: none — AC-6 measures it.
- REQ-7: The pin tests of the bullet shall fail for a text that keeps the words of the bullet and reverses or drops one of its meanings. · failure: a surviving mutant is a missing test, repaired before READY.

## Acceptance

- AC-1 (REQ-1): item 3 of the `## Tests` list (the line that holds `Whole binaries:`) names the cargo commands (`--bin moai`, `--bin moai-cli`, `--lib`, `--test <name>`), the cargo checks of `tools/repo_scans.json` and the macOS-only tests, says `Tests are Rust only`, and holds no `Python` or `unittest`.
  verify: `cargo nextest run --locked --offline --test next_flow_skill_scan -E 'test(self_check_tests_item_names_no_python_test)'`
- AC-2 (REQ-2): item 5 (the line that holds `Files tests read:`) keeps its `moai repo-scans --refs <tip> HEAD` command and says to run only their `cargo` checks and to leave out a Python line.
  verify: `cargo nextest run --locked --offline --test next_flow_skill_scan -E 'test(self_check_repo_scans_item_runs_cargo_checks_only)'`
- AC-3 (REQ-3): a copy of each item with the same words and the reversed meaning (item 3 asks for the Python tests of the area again, item 5 says to run all their checks and to include a Python line) is rejected by the checking function the pin tests use.
  verify: `cargo nextest run --locked --offline --test next_flow_skill_scan -E 'test(self_check_pins_reject_reversed_sentences)'`
- AC-4 (REQ-4): the manifest tool checks without a difference after its write, and the skill package passes its inspection.
  verify: `python3 -I tools/source_manifest.py check`
- AC-5 (REQ-1, REQ-2, REQ-3, REQ-4, REQ-5, REQ-6, REQ-7): the SPEC passes the checker.
  verify: `moai spec check SPEC-20261123-self-check-rust-only`
- AC-6 (REQ-5, REQ-6): the `Audit brief` bullet starts `Audit brief: name the scan tests` with the `moai repo-scans --refs <tip> <head>` command, says `, and record each as a receipt step` with `scan:<id>` and `of the PASS`, still holds the cargo-only words the landed pin reads, and the file is 7974 bytes or less.
  verify: `cargo nextest run --locked --offline --test next_flow_skill_scan -E 'test(land_scans_the_flow_names_the_scan_tests_and_the_pass_receipt_step)'`
- AC-7 (REQ-7): the bullet without `name the scan tests`, without `of the PASS`, with `never name` or `never record` in their place and with the words of the cargo-only reading reversed is rejected by the checking function.
  verify: `cargo nextest run --locked --offline --test next_flow_skill_scan -E 'test(land_scans_the_flow_scan_receipt_mutants_are_caught)'`
- AC-8 (REQ-5, REQ-6): the tests of the landed pin, the stage and size tests and the source manifest still pass on the changed text.
  verify: `cargo nextest run --locked --offline --test mode_workflow_steps --test next_flow_skill_scan --test source_manifest_facts --test skill_structure --test workflow_size_budget --test lane_next_rules --test repo_claude_md`

## Scope

- write: `.moai/specs/SPEC-20261123-self-check-rust-only/**`, `skills/moai/references/internal/self-check.md` (items 3 and 5 only), `skills/moai/references/flows/next-flow.md` (the `Audit brief` bullet, and two short trims in the guard-mutant bullet to stay at 7974 bytes), `skills/moai/references/source-manifest.json` (rewritten by the tool), `tests/next_flow_skill_scan.rs` (new test functions at the end of the file; no existing function changes)
- read-only: every other file; `tools/repo_scans.json`, `src/card/land/scans.rs` and `tools/self_check.sh` are read for the facts above and not edited
- tier: 1 (skill reference texts and text pin tests); the floor list of `references/audit.md` is checked with `moai bundle check`
- history: found by the audits of `SPEC-20261123-doc-sentence-pins` (N3: the two sentences of `self-check.md`; N5: the two meanings the `next-flow.md` bullet lost)

## Plan

1. SPEC first (this file); approval comes from the user through the leader. Approved 2026-11-23T13:20:42Z as `3·5번만` with the added `next-flow.md` scope, recorded in the card note 7530 (author `leader`, the user's choice relayed); this file is not changed to approved here.
2. Base: the three files `next-flow.md`, `source-manifest.json` and `tests/next_flow_skill_scan.rs` are taken from the integration tip `e8315e45da` in their own commit (D-4), so the landed pin can run against the rewritten bullet.
3. RED: the five new test functions at the end of `tests/next_flow_skill_scan.rs`; they fail on the present text for the stated reason.
4. Edit items 3 and 5 of `self-check.md`, and the `Audit brief` bullet and two trims of `next-flow.md`.
5. `python3 -I tools/source_manifest.py write`, then `check`; `moai package-inspect skills/moai`.
6. Guard mutants: reverse each sentence keeping its words; each must turn a pin red.
7. `moai check affected` once at READY, unwrapped.

## Risks

- The landing merges this head onto a tip that may have moved; the bullet line and the manifest row are then the only places of a conflict, and the lander resolves the bullet to this head's text and the manifest by running the tool.
- A pin that reads an item by its list number breaks on a renumbered list: the tests find an item by its first words (`Whole binaries:`, `Files tests read:`).
- The byte budget is a tight 7974: a later sentence added to the flow reference needs its own trim.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | only items 3 and 5 of `self-check.md` change; item 6 (Python unittest files and `tests/skill_references_test.py`) and the `--full-python`, `py-source-manifest`, `py-repo-scans` lines of the header stay and become a follow-up with `tools/self_check.sh` | user decision 2026-11-23 `3·5번만`, card note 7530; those lines describe `tools/self_check.sh` items that exist today | widen the scope to item 6 and the header |
| D-2 | the sentence says only the cargo checks of the `repo-scans` list are run, and does not say the table has no Python lines | the table still has 7 `python3` lines (read 2026-11-23) while the landing keeps cargo checks only (`cargo_only`) | name the Python lines in the sentence |
| D-3 | the pin tests are new functions at the end of the file | the file is edited by other SPECs; separate functions merge without a conflict | add to existing functions |
| D-4 | the three files of the landed `SPEC-20261123-doc-sentence-pins` are copied from the tip into this branch as their own commit before any change | the bullet this SPEC rewrites exists only on the tip, and its landed pin tests must run against the new text; the lane did not merge the tip, only read its files | start from the old base and let the lander resolve the same line |
| D-5 | two short trims in the guard-mutant bullet of the flow reference (the date remark and two words) make room for the restored meanings | the budget is 7974 bytes (leader, card note 7530) and the bullet needs about 40 bytes more; the trimmed words are in no test | shorten the new bullet instead |
| D-6 | the pin of the restored meanings reads whole phrases anchored at their start (`Audit brief: name the scan tests`, `, and record each as a receipt step`) | a mutant with `never name` or `never record` keeps every word and must still fail | read single words |

## Result

- AC-1: met — `self_check_tests_item_names_no_python_test` (red before the edit: item 3 named `python` and lacked `Tests are Rust only`; pass after; a mutant that puts `the Python tests of the area` back into item 3 fails it)
- AC-2: met — `self_check_repo_scans_item_runs_cargo_checks_only` (red before, pass after; a mutant that turns item 5 into `run all their checks and include a Python line` fails it)
- AC-3: met — `self_check_pins_reject_reversed_sentences` (red before, pass after; four mutants of item 3 and four of item 5 each change the text and each turn the pin red)
- AC-4: met — `python3 -I tools/source_manifest.py write` then `check` exit 0; `moai package-inspect skills/moai` exit 0; `self-check.md` 6948 bytes, `next-flow.md` 7964 bytes
- AC-5: met — `moai spec check SPEC-20261123-self-check-rust-only` 0 errors 0 warnings
- AC-6: met — `land_scans_the_flow_names_the_scan_tests_and_the_pass_receipt_step` (red before: the bullet lacked both phrases; pass after, with the file at 7964 of 7974 bytes; a mutant `never name the scan tests` and a mutant `and never record each` each fail it)
- AC-7: met — `land_scans_the_flow_scan_receipt_mutants_are_caught` (red before: the first mutant did not change the text; pass after, five mutants)
- AC-8: met — `mode_workflow_steps`, `next_flow_skill_scan`, `source_manifest_facts`, `skill_structure`, `workflow_size_budget`, `lane_next_rules`, `repo_claude_md`: 119 passed, 0 failed (nextest, one heavy slot)
- Gaps: item 6 and the header of `self-check.md` keep their Python lines (D-1); the Python lines of `tools/repo_scans.json` stay (non-goal). The `check affected` receipt of the head is in the READY note (a number in this file would name an earlier head).
