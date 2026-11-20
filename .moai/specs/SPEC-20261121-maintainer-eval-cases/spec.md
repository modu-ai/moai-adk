---
id: SPEC-20261121-maintainer-eval-cases
title: "Four maintainer evaluation cases from a day of lessons replay deterministically, on frozen evidence and on the current tree"
status: draft
revision: 1
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

- AC-1 (REQ-6): a changed-file list that holds every claimed path is clean; one claimed path missing flags and names it; an empty claim list is clean; the frozen first case (the commit message and changed files of the audited commit, with the two paths the audit named as not changed) flags and its repaired twin is clean · verify: `python3 -m unittest tests.maintainer_cases_test.ClaimsTest`
- AC-2 (REQ-7): a test function that calls `Instant::now`, `.elapsed()` or `SystemTime::now` flags; the same call in a function that is not a test does not; a test without a clock is clean; a nested brace and a string holding a brace do not end the function body early; the frozen pre-fix test flags and the current `src/card/input.rs` is clean · verify: `python3 -m unittest tests.maintainer_cases_test.ClockTest`
- AC-3 (REQ-8): a round line that holds every stage of the chain line in order is clean; a missing stage flags and names it; the stages in another order flag; extra words between the stages do not matter; the current `mode-agent.md` round line against the `chains/index.md` line is the open case · verify: `python3 -m unittest tests.maintainer_cases_test.StageOrderTest`
- AC-4 (REQ-9): a diff that removes a test function flags it; naming it in the allowed list makes it clean; a removed non-test function does not flag; a function removed and added back in the same diff is clean; the frozen diff of the audited deletion flags and the diff that restored the test is clean · verify: `python3 -m unittest tests.maintainer_cases_test.DeletionTest`
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
2. Checks take one input each: `claims_in_diff` a folder with `claims.json` (a list of paths) and `changed-files.txt`; `no_clock_in_tests` a Rust file; `stage_order` a JSON `{"chain": {"file", "contains"}, "round": {"file", "contains"}}` naming the two lines to compare by a literal phrase; `unscoped_test_deletion` a folder with `diff.patch` and `allowed.json`.
3. Replay output: `{"ok", "cases": [{"id", "targets": [{"target", "check", "expect", "verdict", "status", "detail"}]}]}`, `status` one of `ok`, `regressed`, `unexpected`, `now_clean`.
4. RED: the test file builds temporary inputs for each check, the case-file refusals and the replay on the repository. GREEN: the tool, the case file, the frozen inputs. Written for Python 3.9, standard library only.
5. Mutants (at most 8): the clock check also flags non-test functions, a closing brace inside a string ends the body, the stage order compared as a set, an empty allowed list treated as allow-all, `now_clean` counted as a failure, `open` treated like `flag`, the claim comparison reversed, the replay exits 0 on a regressed target.

Evidence captured for the frozen inputs (read 2026-11-21 with read-only `git show`): the message and the changed-file list of the commit that wrote "the heavy-slot default-3 leftovers say five"; the pre-fix wall-clock test of `src/card/input.rs` (the version before the commit that replaced three clock tests by a work count); the diff of the commit that removed `claude_settings_op_neither_writes_nor_drops_hooks` and the diff of the commit that put it back. The incident commits are not ancestors of the local integration branch at the time of writing.

## Risks

- The first case needs a claim list that a model or a person extracts; the check is only the comparison. A wrong list gives a wrong verdict, so the frozen list names the two paths the audit named.
- A clock check by text cannot see a clock hidden behind a helper function; it names the three direct markers and the Result says what it cannot see.
- The `open` guidance case makes the replay depend on guidance text that other branches will edit. A fix turns it `now_clean` (exit 0); only a clean target that becomes flagged fails.
- The frozen inputs copy text of earlier commits, which may hold card numbers from the incident; the previous slice's ratchet does not scan `tools/`, and the frozen copies are evidence, not artifacts a user sees.
- `tools/repo_scans.json` is held by another card at the time of writing; the entry (REQ-12) waits for it.

## Result

(not yet written)
