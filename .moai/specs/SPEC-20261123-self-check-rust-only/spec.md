---
id: SPEC-20261123-self-check-rust-only
title: "self-check.md 의 테스트 문장이 Rust 전용 규칙을 말하고 repo-scans 는 cargo 검사만 돈다고 말하며, next-flow.md 의 감사 브리프 문장이 스캔 테스트 이름과 PASS 영수증 단계를 다시 말함"
status: draft
revision: 3
created: 2026-11-23
---

# self-check.md 의 테스트 문장이 Rust 전용 규칙을 말하고 repo-scans 는 cargo 검사만 돈다고 말하며, next-flow.md 의 감사 브리프 문장이 스캔 테스트 이름과 PASS 영수증 단계를 다시 말함

## Intent

`skills/moai/references/internal/self-check.md` 의 `## Tests` 목록에서 두 문장이 저장소 규칙과 어긋나요. 3번 항목은 READY 전에 돌릴 묶음에 "the Python tests of the area" 를 넣어서 `CLAUDE.md` 의 "Tests are Rust only" 와 맞지 않아요. 5번 항목은 `moai repo-scans --refs <tip> HEAD` 가 "repo-wide scans" 를 나열한다고만 말해서, 나열된 표에는 `python3 -m unittest …` 줄이 7개 있는데 그 줄을 손으로 돌려야 하는지 알 수 없어요. 착지 명령은 표에서 cargo 검사만 요구해요(`src/card/land/scans.rs` 의 `cargo_only`).

이 SPEC 은 두 문장을 고치고, 문장의 뜻이 뒤집히면 실패하는 고정 테스트를 `tests/next_flow_skill_scan.rs` 에 새 함수로 더해요.

개정 2(사용자 승인 때 리더가 덧붙인 범위): 같은 표를 말하는 `skills/moai/references/flows/next-flow.md` 의 감사 브리프 문장이 `SPEC-20261123-doc-sentence-pins` 로 cargo 검사만 돈다는 뜻을 얻는 대신 두 뜻을 잃었어요: 감사 브리프가 스캔 테스트의 이름을 대야 한다는 것, 그리고 각 스캔을 PASS 영수증의 `scan:<id>` 단계로 적는다는 것. 이 SPEC 이 두 뜻을 되살리고, 파일이 `7974` 바이트(그 SPEC 이 착지한 크기) 아래에 머물게 문장을 줄여 써요. 되살린 문장에도 뜻이 뒤집히면 실패하는 고정 테스트를 더해요.

개정 3(카드 C-412, 감사 #404 의 비차단 N1·N2·N4): 수리 뒤의 핀은 문구를 앞뒤 낱말까지 읽지만 문구 뒤에 붙는 말은 읽지 않아요. 그래서 3번 항목의 `` `tools/repo_scans.json`, and the macOS `` 를 `` `tools/repo_scans.json` left out, and the macOS `` 로 바꿔도 테스트가 통과해요(N1). 핀이 읽지 않는 말, 곧 `the changed` → `the unchanged`, `also` → `never`, `of the PASS for` → `of the PASS, never for` 같은 바꾸기도 통과해요(N2). 이 개정은 핀 테스트가 문장을 항목(줄)의 끝까지 통째로 읽게 하고, 문장의 낱말 하나를 지우거나 그 앞에 `not` 을 붙인 글도 실패하는지 낱말마다 훑어 보이게 해요. 문서 문장은 바꾸지 않아요(핀만 엄격해져요). D-7 행의 중첩 백틱도 정리해요(N4).

## Non-goals

- `tools/repo_scans.json` 표의 Python 줄을 지우거나 `moai repo-scans` 의 출력을 바꾸는 일(표는 그대로, 그 줄의 처리는 다른 카드의 일).
- `self-check.md` 의 6번 항목(Python)과 머리말의 `--full-python`, `py-*` 항목 설명을 고치는 일(D-1, 사용자가 3·5번만으로 승인).
- `lane-rules.md` 와 `next-flow.md` 의 repo-scans 문장 밖의 줄을 바꾸는 일(바이트를 맞추려고 같은 파일의 세 곳을 줄이는 일만 D-5 로 해요).
- `tools/self_check.sh` 의 동작.

## Requirements

- REQ-1: The third item of the `## Tests` list in `skills/moai/references/internal/self-check.md` shall name only cargo commands and `cfg(target_os)` tests as the set run before READY, with no Python test in it. · failure: none — AC-1 pins it.
- REQ-2: The fifth item of the same list shall say that `moai repo-scans --refs <tip> HEAD` lists the repo-wide scans and that only the cargo checks of that list are run, the Python lines of the table never by hand. · failure: none — AC-2 pins it.
- REQ-3: The pin tests shall fail for a text that keeps the words of the two items and reverses their meaning. · failure: a surviving reversed text is a missing test, repaired before READY.
- REQ-4: `skills/moai/references/source-manifest.json` shall be rewritten by `tools/source_manifest.py write`, pass `tools/source_manifest.py check` and keep every package file at most 65,536 bytes. · failure: none — AC-4 runs both.
- REQ-5: The `Audit brief` bullet of `skills/moai/references/flows/next-flow.md` shall tell the session that briefs an audit to name the scan tests that `moai repo-scans --refs <tip> <head>` picks and to record each as a receipt step `scan:<id>` of the PASS, beside the cargo-only reading it already has. · failure: none — AC-6 pins it.
- REQ-6: The flow reference shall stay at most 7974 bytes. · failure: none — AC-6 measures it.
- REQ-7: The pin tests of the bullet shall fail for a text that keeps the words of the bullet and reverses or drops one of its meanings. · failure: a surviving mutant is a missing test, repaired before READY.
- REQ-8: The pin tests of items 3 and 5 of `self-check.md` and of the `Audit brief` bullet of `next-flow.md` shall read each pinned sentence whole to the end of its line, so that a word changed, added or put after the pinned phrase fails them.
- REQ-9: The pin tests of REQ-8 shall fail for a text in which any one word of a pinned sentence is deleted or has `not` put before it.
- REQ-10: The `## Decisions` table of this SPEC shall write a code span that holds backticks with double backticks around it and no single-backtick span that holds a backtick.

## Acceptance

- AC-1 (REQ-1): item 3 of the `## Tests` list (the line that holds `Whole binaries:`) names the cargo commands (`--bin moai`, `--bin moai-cli`, `--lib`, `--test <name>`), the cargo checks of `tools/repo_scans.json` and the macOS-only tests, says `Tests are Rust only`, and holds no `Python` or `unittest`.
  verify: `cargo nextest run --locked --offline --test next_flow_skill_scan -E 'test(self_check_tests_item_names_no_python_test)'`
- AC-2 (REQ-2): item 5 (the line that holds `Files tests read:`) keeps its `moai repo-scans --refs <tip> HEAD` command and says to run only their `cargo` checks and to leave out a Python line.
  verify: `cargo nextest run --locked --offline --test next_flow_skill_scan -E 'test(self_check_repo_scans_item_runs_cargo_checks_only)'`
- AC-3 (REQ-3): a copy of each item with the same words and the reversed meaning (item 3 asks for the Python tests of the area again, item 5 says to run all their checks and to include a Python line) is rejected by the checking function the pin tests use.
  verify: `cargo nextest run --locked --offline --test next_flow_skill_scan -E 'test(self_check_pins_reject_reversed_sentences)'`
- AC-4 (REQ-4): the manifest tool checks without a difference after its write, and the skill package passes its inspection.
  verify: `python3 -I tools/source_manifest.py check`
- AC-5 (REQ-1, REQ-2, REQ-3, REQ-4, REQ-5, REQ-6, REQ-7, REQ-8, REQ-9, REQ-10): the SPEC passes the checker.
  verify: `moai spec check SPEC-20261123-self-check-rust-only`
- AC-6 (REQ-5, REQ-6): the `Audit brief` bullet starts `Audit brief: name the scan tests` with the `moai repo-scans --refs <tip> <head>` command, says `, and record each as a receipt step` with `scan:<id>` and `of the PASS`, still holds the cargo-only words the landed pin reads, and the file is 7974 bytes or less.
  verify: `cargo nextest run --locked --offline --test next_flow_skill_scan -E 'test(land_scans_the_flow_names_the_scan_tests_and_the_pass_receipt_step)'`
- AC-7 (REQ-7): the bullet without `name the scan tests`, without `of the PASS`, with `never name` or `never record` in their place and with the words of the cargo-only reading reversed is rejected by the checking function.
  verify: `cargo nextest run --locked --offline --test next_flow_skill_scan -E 'test(land_scans_the_flow_scan_receipt_mutants_are_caught)'`
- AC-8 (REQ-5, REQ-6): the tests of the landed pin, the stage and size tests and the source manifest still pass on the changed text.
  verify: `cargo nextest run --locked --offline --test mode_workflow_steps --test next_flow_skill_scan --test source_manifest_facts --test skill_structure --test workflow_size_budget --test lane_next_rules --test repo_claude_md`

- AC-9 (REQ-8): 3번 항목의 `tools/repo_scans.json` 뒤에 `left out` 을 붙인 글, `the changed` 를 `the unchanged` 로, `also` 를 `never` 로 바꾼 글, 5번 항목 끝에 말을 붙인 글, 감사 브리프 글머리표의 `of the PASS for` 를 `of the PASS, never for` 로 바꾼 글이 각각 원문과 다르고 핀 함수에서 문제가 돼요.
  verify: `cargo nextest run --locked --offline --test next_flow_skill_scan -E 'test(pinned_sentences_are_read_to_their_end)'`
- AC-10 (REQ-9): 3번 항목의 끝 문장, 5번 항목의 끝 문장, 감사 브리프 글머리표 전체에서 낱말을 하나씩 지운 글과 낱말 앞에 `not` 을 하나씩 붙인 글이 모두 원문과 다르고 모두 핀 함수에서 문제가 돼요.
  verify: `cargo nextest run --locked --offline --test next_flow_skill_scan -E 'test(every_word_of_a_pinned_sentence_is_read)'`
- AC-11 (REQ-10): `## Decisions` 표에 백틱이 든 코드 조각은 이중 백틱으로 감싸 있고, 백틱이 든 단일 백틱 조각이 없어요.
  verify: manual: D-7 행을 열어 코드 조각의 여는 백틱과 닫는 백틱 개수를 읽어요(감사자가 렌더된 표로도 확인해요).

## Scope

- write: `.moai/specs/SPEC-20261123-self-check-rust-only/**`, `skills/moai/references/internal/self-check.md` (items 3 and 5 only), `skills/moai/references/flows/next-flow.md` (the `Audit brief` bullet, and three short trims in the guard-mutant bullet to stay at 7974 bytes), `skills/moai/references/source-manifest.json` (rewritten by the tool), `tests/next_flow_skill_scan.rs` (new test functions at the end of the file; the audit repair D-7 also changes the checking function `repo_scans_cargo_only_problems`, the item-3 phrase list of `self_check_tests_item_problems` and the mutant lists of two tests)
- read-only: every other file; `tools/repo_scans.json`, `src/card/land/scans.rs` and `tools/self_check.sh` are read for the facts above and not edited
- tier: 1 (skill reference texts and text pin tests); the floor list of `references/audit.md` is checked with `moai bundle check`
- history: found by the audits of `SPEC-20261123-doc-sentence-pins` (N3: the two sentences of `self-check.md`; N5: the two meanings the `next-flow.md` bullet lost)

## Plan

1. SPEC first (this file); approval comes from the user through the leader. Approved 2026-11-23T13:20:42Z as `3·5번만` with the added `next-flow.md` scope, recorded in the card note 7530 (author `leader`, the user's choice relayed); this file is not changed to approved here.
2. Base: the three files `next-flow.md`, `source-manifest.json` and `tests/next_flow_skill_scan.rs` are taken from the integration tip `e8315e45da` in their own commit (D-4), so the landed pin can run against the rewritten bullet.
3. RED: the five new test functions at the end of `tests/next_flow_skill_scan.rs`; they fail on the present text for the stated reason.
4. Edit items 3 and 5 of `self-check.md`, and the `Audit brief` bullet and three trims of `next-flow.md`.
5. `python3 -I tools/source_manifest.py write`, then `check`; `moai package-inspect skills/moai`.
6. Guard mutants: reverse each sentence keeping its words; each must turn a pin red.
7. `moai check affected` once at READY, unwrapped.
8. 개정 3: RED 먼저 — 새 테스트 둘(`pinned_sentences_are_read_to_their_end`, `every_word_of_a_pinned_sentence_is_read`)을 파일 끝에 넣고 지금 핀에서 실패하는 걸 봐요. GREEN — 세 핀 함수(3번 항목, 5번 항목, 감사 브리프 글머리표)가 문장을 줄 끝까지 통째로 읽게 고쳐요. 문서 문장은 바꾸지 않아서 `source-manifest.json` 도 그대로예요(`tools/source_manifest.py check` 로 확인). 개정 3 의 쓰기 범위는 이 SPEC 폴더와 `tests/next_flow_skill_scan.rs` 둘이에요.

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
| D-5 | three short trims in the guard-mutant bullet of the flow reference (`in its body`, the date remark `(four times on 2026-11-23)` and `it reports`) make room for the restored meanings | the budget is 7974 bytes (leader, card note 7530) and the bullet needs about 40 bytes more; the trimmed words are in no test | shorten the new bullet instead |
| D-6 | the pin of the restored meanings reads whole phrases anchored at their start (`Audit brief: name the scan tests`, `, and record each as a receipt step`) | a mutant with `never name` or `never record` keeps every word and must still fail | read single words |
| D-7 | audit repair of #404: each pinned sentence is read whole, from the word before it to the word after it (``picks; run only their `cargo` checks on the merge tree, leaving out a Python line, and record each`` in the flow bullet; `` area, every `cargo` check of `tools/repo_scans.json` `` and `can run. Tests are Rust only.` in item 3), and the four opposite sentences the auditor found (`picks; never run only their`, `not leaving out`, ``not every `cargo` check of``, `Never assume Tests are Rust only.`) are mutants of the pin tests | the loose substring pins passed a text that kept every word and reversed the meaning, which REQ-3 and REQ-7 say must fail (audit #404 FAIL B1, B2); the checking function of the landed pin is read in the same whole-phrase form | go back to the single phrases |
| D-8 | 개정 3(카드 C-412): 핀은 문장을 항목(줄)의 끝까지 통째로 읽고(3번 항목은 ``The set before READY`` 부터 줄 끝까지, 5번 항목은 ``lists the repo-wide scans`` 의 앞 명령부터 줄 끝까지, 감사 브리프 글머리표는 줄 전체), 새 테스트 둘은 새 함수로 파일 끝에 더하고 낱말 하나를 지우거나 앞에 ``not`` 을 붙이는 글을 낱말마다 훑어요 | 수리 뒤 핀이 문구만 읽어서 문구 뒤에 붙는 말(N1)과 핀이 읽지 않는 낱말(N2)을 바꿔도 통과해요(감사 #404 비차단). 줄 끝까지 읽으면 같은 줄에 다른 카드가 문장을 더할 때 핀도 고쳐야 하는 값이 있지만, 그 줄은 이 SPEC 이 문장으로 고정한 줄이에요 | 핀을 문구 읽기로 되돌리고 새 테스트 둘을 지우기 |

## Result

- AC-1: met — `self_check_tests_item_names_no_python_test` (red before the edit: item 3 named `python` and lacked `Tests are Rust only`; pass after; a mutant that puts `the Python tests of the area` back into item 3 fails it)
- AC-2: met — `self_check_repo_scans_item_runs_cargo_checks_only` (red before, pass after; a mutant that turns item 5 into `run all their checks and include a Python line` fails it)
- AC-3: met — `self_check_pins_reject_reversed_sentences` (red before, pass after; six mutants of item 3 and four of item 5 each change the text and each turn the pin red; the repair D-7 added `not every cargo check of` and `Never assume Tests are Rust only.`, which the loose pin had let through)
- AC-4: met — `python3 -I tools/source_manifest.py write` then `check` exit 0; `moai package-inspect skills/moai` exit 0; `self-check.md` 6948 bytes, `next-flow.md` 7964 bytes
- AC-5: met — `moai spec check SPEC-20261123-self-check-rust-only` 0 errors 0 warnings
- AC-6: met — `land_scans_the_flow_names_the_scan_tests_and_the_pass_receipt_step` (red before: the bullet lacked both phrases; pass after, with the file at 7964 of 7974 bytes; a mutant `never name the scan tests` and a mutant `and never record each` each fail it)
- AC-7: met — `land_scans_the_flow_scan_receipt_mutants_are_caught` (red before: the first mutant did not change the text; pass after, five mutants; the repair D-7 added `picks; never run only their` and `not leaving out` to `land_scans_the_flow_cargo_only_mutants_are_caught`, which reads the same checking function)
- AC-8: met — `mode_workflow_steps`, `next_flow_skill_scan`, `source_manifest_facts`, `skill_structure`, `workflow_size_budget`, `lane_next_rules`, `repo_claude_md`: 119 passed, 0 failed (nextest, one heavy slot)
- Gaps: item 6 and the header of `self-check.md` keep their Python lines (D-1); the Python lines of `tools/repo_scans.json` stay (non-goal). The `check affected` receipt of the head is in the READY note (a number in this file would name an earlier head).
