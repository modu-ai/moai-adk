---
id: SPEC-20261121-maintainer-eval-cases
title: "Four maintainer evaluation cases from a day of lessons replay deterministically, on frozen evidence and on the current tree"
status: draft
revision: 3
created: 2026-11-21
---

# Four maintainer evaluation cases from a day of lessons replay deterministically, on frozen evidence and on the current tree

## Intent

The self-improvement report (section 4.4) names four lessons from one day of work in this repository: a commit message that claimed a repair it did not make, a test that bounded work with a wall clock, guidance whose step list drifted from the documented order, and a test deleted outside the change's scope. A rule that came out of such a lesson is only worth keeping if it can be checked again. This slice turns the four lessons into the first maintainer evaluation set: a case file, frozen copies of the evidence, and a replay tool whose four checks are mechanical (no model, no network). The replay is the part of the later evaluation gate that can run today; wiring it to the `moai improve` engine waits for that engine.

The evidence is frozen because three of the incident commits are not on the local integration branch and a branch that is cleaned up takes its commits with it. Two cases also run on the current tree: the wall-clock test must stay fixed (`clean`), and the guidance drift is still open (`open`: the replay flags it now, and reports `now_clean` without failing once someone fixes it, so a correct fix is never punished).

Repository-only: the files live under `tools/` and `tests/`, outside the shipped text and the package manifest (the previous slice's ratchet holds them out).

## Non-goals

- Wiring the cases to the `moai improve` engine, signals or the apply step (later slices of the same card).
- Judging claims in prose: the claim list of the first case is what an audit agent extracts; the check only compares that list with the changed files.
- A fifth case, or generating cases from the card ledger.
- Any change of the shipped skills, agents or help text, and any fix of the open guidance drift.

## Requirements

- REQ-1: The replay shall run every target of every case through the check the case names and compare the verdict with the expected one.
- REQ-2: WHEN every target ends as expected, the replay shall exit 0. · failure: none
- REQ-3 (REQ-2): IF a target ends with another verdict than the strict one it expects, THEN the replay shall exit 1 and name the case and the target with both verdicts.
- REQ-4: The replay shall report a target that expects `open` and ends clean as `now_clean` while it exits 0 for it.
- REQ-5: IF the case file or a frozen input is missing or malformed or names an unknown check, THEN the replay shall exit 2 with the reason on stderr and nothing on stdout.
- REQ-6: The check `claims_in_diff` shall flag a case when a claimed path is absent from the changed-file list.
- REQ-7: The check `no_clock_in_tests` shall flag a test function whose body reads a clock.
- REQ-8: The check `stage_order` shall flag a guidance line that lacks a stage of the chain line or lists the stages in another order.
- REQ-9: The check `unscoped_test_deletion` shall flag a removed test function that the allowed list does not name.
- REQ-10: The replay shall write no file.
- REQ-11: The tool and the case file and the frozen inputs and the test shall lie outside the shipped paths and outside the package source manifest.
- REQ-12: The table `tools/repo_scans.json` shall list the test of the replay with the files the replay reads as its scope.

## Acceptance

- AC-1 (REQ-6): a changed-file list that holds every claimed path is clean; one claimed path missing flags and names it; an empty claim list is clean; the frozen first case (the changed files of the audited commit and a claim list that also names the two paths the audit found not changed) flags and its repaired twin, whose claims are only the changed paths, is clean; a missing or non-text claim file is refused · verify: `python3 -m unittest tests.maintainer_cases_test.ClaimsTest`
- AC-2 (REQ-7): a test function that calls `Instant::now`, `.elapsed()` or `SystemTime::now` flags; the same call in a function that is not a test does not; a test without a clock is clean; a nested brace and a string holding a brace do not end the function body early; the frozen pre-fix test flags and the current `src/card/input.rs` is clean · verify: `python3 -m unittest tests.maintainer_cases_test.ClockTest`
- AC-3 (REQ-8): a round line that holds every stage of the chain line in order is clean; a missing stage flags and names it; the stages in another order flag; extra words between the stages do not matter; the current `mode-agent.md` round line against the `chains/index.md` line is the open case · verify: `python3 -m unittest tests.maintainer_cases_test.StageOrderTest`
- AC-4 (REQ-9): a diff that removes a test function flags it; naming it in the allowed list makes it clean; a removed non-test function does not flag; a function removed and added back in the same diff is clean; the frozen test diff of the audited deletion flags exactly the two tests the audit named, and the same diff followed by the diff that put them back is clean · verify: `python3 -m unittest tests.maintainer_cases_test.DeletionTest`
- AC-5 (REQ-1, REQ-2, REQ-3, REQ-4): on the repository the replay exits 0 and reports the four cases, each target `ok` except the open one; with a case file whose expectation is flipped it exits 1 naming the case and target; an open target whose input turns clean is `now_clean` and the exit is 0 · verify: `python3 -m unittest tests.maintainer_cases_test.ReplayTest`
- AC-6 (REQ-5, REQ-10): each of a missing case file, a missing frozen file, a case with an unknown check, a case with a missing key, a target expecting a value outside `flag`/`clean`/`open`, and a case file with a duplicate id exits 2 with stderr text and empty stdout; a hash of the repository paths the replay reads is the same before and after a run · verify: `python3 -m unittest tests.maintainer_cases_test.RefusalTest`
- AC-7 (REQ-11, REQ-12): none of the new files is under `skills/moai/` or `agents/` or named in `skills/moai/references/source-manifest.json`; the repo-scans table lists the test and its completeness test passes · verify: `python3 -m unittest tests.maintainer_cases_test.RepoTreeTest` and `python3 -m unittest tests.repo_scans_test`

## Scope

- write: `tools/maintainer_cases.py` (new), `tools/improve/maintainer-cases.json` (new), `tools/improve/cases/**` (new frozen inputs), `tests/maintainer_cases_test.py` (new), `tools/repo_scans.json` (one entry, last step), `.moai/specs/SPEC-20261121-maintainer-eval-cases/**`
- read-only: `src/card/input.rs`, `skills/moai/workflows/mode-agent.md`, `skills/moai/references/chains/index.md`, `skills/moai/references/source-manifest.json`, `tests/repo_scans_test.py`
- forbidden: any edit of the shipped paths, `AGENTS.local.md`, the Go repository, the real HOME

## Plan

Tests first (RED), then the tool and the cases, then the frozen inputs captured from the incident commits with the commands recorded in the Result, then mutants, then the table entry.

1. Case file shape: `{"version": 1, "cases": [{"id", "lesson", "report", "check", "targets": {name: {"input": <path or folder>, "expect": "flag"|"clean"|"open"}}}]}`; `input` paths are relative to the root; the frozen ones live under `tools/improve/cases/<id>/`.
2. Checks take one input each: `claims_in_diff` a folder with `claims.txt` (one claimed path per line, no quotes: the docs-drift scan of `tests/ci_scope_test.py` reads a quoted docs path in `tools/` as a path a test reads) and `changed-files.txt`; `no_clock_in_tests` a Rust file; `stage_order` a JSON `{"chain": {"file", "contains"}, "round": {"file", "contains"}}` naming the two lines to compare by a literal phrase; `unscoped_test_deletion` a folder with `diff.patch` and `allowed.json`.
3. Replay output: `{"ok", "cases": [{"id", "targets": [{"target", "check", "expect", "verdict", "status", "detail"}]}]}`, `status` one of `ok`, `regressed`, `unexpected`, `now_clean`.
4. RED: the test file builds temporary inputs for each check, the case-file refusals and the replay on the repository. GREEN: the tool, the case file, the frozen inputs. Written for Python 3.9, standard library only.
5. Mutants (at most 8): the clock check also flags non-test functions, a closing brace inside a string ends the body, the stage order compared as a set, an empty allowed list treated as allow-all, `now_clean` counted as a failure, `open` treated like `flag`, the claim comparison reversed, the replay exits 0 on a regressed target.

Evidence captured for the frozen inputs (read 2026-11-21 with read-only `git show`): the message and the changed-file list of the commit that wrote "the heavy-slot default-3 leftovers say five"; the pre-fix wall-clock test of `src/card/input.rs` (the version before the commit that replaced three clock tests by a work count); the diff of the commit that removed `claude_settings_op_neither_writes_nor_drops_hooks` and the diff of the commit that put it back. The incident commits are not ancestors of the local integration branch at the time of writing. The copies have their trailing whitespace trimmed so that `git diff --check` passes on the branch; the checks read only the first character of a diff line and function names, so the verdicts do not change (the replay output was compared byte for byte before and after).

## Risks

- The first case needs a claim list that a model or a person extracts; the check is only the comparison. A wrong list gives a wrong verdict, so the frozen list names the two paths the audit named. The list of allowed removals of the deletion case holds the five tests whose names say `omp`, the retired host the audited change removed on purpose; the two tests the audit named as out of scope are the ones left off it. `message.txt` of the first case is evidence for a reader and no check reads it.
- A clock check by text cannot see a clock hidden behind a helper function; it names the three direct markers and the Result says what it cannot see.
- The stage check is a subsequence match: every stage of the chain line must appear in order, and an extra out-of-order mention of one stage elsewhere in the line does not fail it. A parenthesis inside the chain line ends the stage list there. Both are limits of a text comparison, not of the guidance.
- The `open` guidance case makes the replay depend on guidance text that other branches will edit. A fix turns it `now_clean` (exit 0); only a clean target that becomes flagged fails.
- The frozen inputs copy text of earlier commits, which may hold card numbers from the incident; the previous slice's ratchet does not scan `tools/`, and the frozen copies are evidence, not artifacts a user sees.
- `tools/repo_scans.json` is held by another card at the time of writing; the entry (REQ-12) waits for it.
- Quoted docs paths in `tools/` count as paths the tests read (the first draft kept the claims as JSON and `tests.ci_scope_test` failed on `docs/ci-rc-ci-spec.md`); the claim lists are plain text for that reason, and a test pins that they hold no quote.

## Result

Built tests first. The first RED run of `python3 -m unittest tests.maintainer_cases_test` (tool absent) gave `Ran 46 tests`, `FAILED (failures=46)`, all assertion failures; the tool and the data turned it green. A second RED round (12 failures of 48) moved the claim lists from JSON to plain text lines, see Risks. First green state: `python3 -m unittest tests.maintainer_cases_test` → `Ran 48 tests ... OK` on Python 3.14.2 and on the system `/usr/bin/python3` 3.9.6; the repair round below raised it to 61.

- AC-1: met — `ClaimsTest` (4 tests) and the refusal tests for a missing or non-UTF-8 claim file pass; on the repository the frozen first case flags `claimed but not changed: src/main.rs, tools/heavy_check_slot.py` and its repaired twin is clean.
- AC-2: met — `ClockTest` (14 tests) passes; on the repository the pre-fix test flags `secret_screen_time_on_one_long_word_has_a_bound` and `src/card/input.rs` is clean (`no clock in 26 test function(s)`).
- AC-3: met — `StageOrderTest` (7 tests) passes; on the repository the round line of `mode-agent.md` is reported `open` with `missing or out of order: SYNC (chain: PLAN → RUN → VERIFY → SYNC → AUDIT → merge)`.
- AC-4: met — `DeletionTest` (12 tests) passes; the frozen audited diff flags exactly `apply_refuses_a_stale_replace_digest_after_the_file_changed` and `claude_settings_op_neither_writes_nor_drops_hooks`, and the same diff followed by the restoring diff is clean (`7 removed test(s), all allowed or put back`).
- AC-5: met — `ReplayTest` (7 tests) passes; `python3 -I tools/maintainer_cases.py replay` on the repository exits 0 with seven targets, six `ok` and one `open`.
- AC-6: met — `RefusalTest` (15 tests) passes: each refusal exits 2 with the `maintainer_cases: ` prefix on stderr and empty stdout; the repository run changes no byte of `tools/improve`, `src/card/input.rs` or `skills/moai`.
- AC-7: met — `RepoTreeTest` (2 tests) passes: no new file is under `skills/moai/` or `agents/` or named in the source manifest. The table entry `maintainer-eval-cases` is in `tools/repo_scans.json`; without the two entries `tests.repo_scans_test` fails on `['tests/maintainer_cases_test.py', 'tests/shipped_text_ratchet_test.py']`, with them `Ran 25 tests ... OK` (the build path set in `CARGO_TARGET_DIR`, which the other scans need).

Mutants (8, one rule each, run on the committed tool, the tree restored after each): the clock check flagging non-test functions, the clock check reading comments and strings as code, the stage order compared as a set, an empty allowed list allowing everything, `now_clean` counted as a failure, an `open` target that ends clean called `unexpected`, the claim comparison reversed, and the replay exiting 0 on a regressed target. All 8 were caught (failing tests per mutant: 2, 2, 1, 3, 1, 1, 6, 2); the unmutated run is green and the tree is clean after.

Other checks run: `git diff --check` over the branch exits 0 (first it flagged trailing whitespace in the frozen copies, which were trimmed, the replay output byte-identical); `python3 -m unittest tests.ci_scope_test tests.skill_references_test tests.shipped_text_ratchet_test` → 88 tests OK; every new file is free of the card-number shape (scan with the ratchet's own pattern: 0 lines).

What the replay cannot see: a clock read behind a helper function; whether a claim list is complete (a person or an audit agent extracts it); test functions other than `#[test]`-marked ones (with any path and arguments) and Python `def test*`; guidance that restates the stages in other words than the stage names; an out-of-order extra mention of a stage (see Risks). Not run: the Rust test suites (no Rust file changed), Windows and Linux, the later engine wiring, and a replay inside the `moai improve` gate (the engine does not exist yet).

Repair round (an independent audit of the first head gave one blocking finding, in the other SPEC, and non-blocking ones about this tool): the audit showed that the first version of the deletion check missed a renamed test in a real `git diff` (the unchanged `#[test]` line is context there, not a `-` line), a comment or a one-line `#[test] fn` between marker and name, and treated any added function of the same name as the test put back; so the Result's earlier sentence that a renamed test counts as removed was wrong at that head. It is true now: a context marker counts for the removed and the added side, and an added function counts as put back only when it is a test. The clock check now reads `#[tokio::test(...)]`, a return type with a semicolon (`[u8; 4]`) and raw byte strings (`br#"..."#`), and a hyphen is no longer a stage-name boundary (`RUN-time` is not `RUN`). Thirteen tests were added (RED: 8 failures, 5 pins of behaviour that was right and had no test: nested block comments, raw strings, char literals, an attribute with arguments in a removal, an escaping input path) and one test was rewritten because a mutant survived it (the escape test passed for another reason: its input did not exist; it now uses an existing folder outside the root). Final run: `Ran 61 tests ... OK` on Python 3.14.2 and 3.9.6 (`ClaimsTest` 4, `ClockTest` 14, `DeletionTest` 12, `RefusalTest` 15, `ReplayTest` 7, `RepoTreeTest` 2, `StageOrderTest` 7). Mutants of the repair round: twelve, each caught after the escape test was fixed (the clock check on non-test functions, nested comments ending early, raw strings not recognised, char literals not blanked, an escaping input path allowed, the byte prefix not skipped, attribute arguments not read, a semicolon anywhere ending the signature search, a context marker counted for one side only, an added non-test counted as put back, a hyphen as a stage boundary, a comment resetting the marker); the first mutant set was run again, one of its eight patterns (the clock check on non-test functions) was re-made for the changed line and is the first of the twelve. The replay on the repository gives the same seven verdicts as before the repair.
