---
id: SPEC-20261125-update-closing-commit
title: "The update route ends with the closing commit offer and a reading of what a Go-generation update left"
status: draft
revision: 1
created: 2026-11-25
---

# The update route ends with the closing commit offer and a reading of what a Go-generation update left

## Intent

`/moai update` moves the earlier Go generation's files on disk and then stops. The project route (`workflows/project.md`, step 8) ends with an offer to commit what the route changed; the update route has no such step. A file that git tracks and that a step moved comes back at the next `git checkout`, and in every worktree made from that branch, so a project whose update was never committed finds the Go state again after the first branch switch (the report `go-generation-update-2026-11-21.md`, G4, section 8 row S6). This SPEC gives `workflows/update.md` the same closing offer, a way to read the three things a Go-generation update leaves (the `go-residue` step, the `claude-md` step and the leftover folders), and the sentence that says what happens without the commit. It changes skill text and its pinned test only.

## Non-goals

- The `go-residue` detection, its apply and revert commands, and the `claude-md` step as its own step: those are code slices that come before or after this one; the text names a step only as "when the plan has it".
- Any change of the project route's own closing step.
- Pushing, opening a pull request, or editing a managed file by hand.

## Requirements

- REQ-1: The UPDATE workflow shall end its procedure with a step that offers the closing commit after the final re-run and before the final report, using the read of `git status --porcelain=v1 -z --untracked-files=all` and the rules of the project route's closing step.
- REQ-2: The UPDATE workflow shall tell the session to list each changed path under the apply that moved or wrote it, to mark each moved path that git tracks as tracked, and to keep the backup, receipt and leftover folders out of the commit.
- REQ-3: The UPDATE workflow shall tell the session that an uncommitted move of a tracked file comes back at the next checkout and in every worktree made from that branch.
- REQ-4: The UPDATE workflow shall tell the session that, when the project's work flow is m-flow and the current branch is its integration branch, the closing commit goes on a branch of its own named `spec/<slug>-<yyyymmdd>` and never on the integration branch.
- REQ-5: The UPDATE workflow shall tell the session how to read a `go-residue` step, a `claude-md` step and the leftover folders of the plan when the plan carries them.
- REQ-6: The UPDATE workflow shall keep the closing commit from staging a secret-looking path, from using `git add -A`, and from pushing.
- REQ-7: The UPDATE workflow's completion check shall name the closing commit as applied, left uncommitted by the user's pick, or not offered with its reason.

## Acceptance

- AC-1 (REQ-1, REQ-7): the closing step and its place in the procedure and the completion check are in `workflows/update.md`.
  verify: `cargo test --locked --offline --test update_closing_commit closing_commit_is_the_last_step`
- AC-2 (REQ-2, REQ-3, REQ-6): the listing, tracking, kept-local and resurrection sentences and the commit safety rules are in the text.
  verify: `cargo test --locked --offline --test update_closing_commit closing_commit_lists_marks_and_warns`
- AC-3 (REQ-4): the m-flow branch rule is in the text.
  verify: `cargo test --locked --offline --test update_closing_commit m_flow_commits_on_a_branch_of_its_own`
- AC-4 (REQ-5): the three readings are in the text, each saying what to show and what to ask.
  verify: `cargo test --locked --offline --test update_closing_commit go_generation_leftovers_are_read_out`
- AC-5: the source manifest row of `workflows/update.md` matches the file and the package tests pass.
  verify: `cargo test --locked --offline --test source_manifest_facts`

## Scope

- write: `.moai/specs/SPEC-20261125-update-closing-commit/**`, `tests/update_closing_commit.rs`, `skills/moai/workflows/update.md`, and, through a short line lease because the project profile lists them as shared files, one row of `skills/moai/references/source-manifest.json` and the skill-package scan entry of `tools/repo_scans.json`
- read-only: `skills/moai/workflows/project.md`, `src/project_update.rs`, `tests/update_route.rs`
- forbidden: a card, Issue or pull request number in the sentences this change adds, and any edit of a file another card holds before that card has released it or agreed in a pair line

## Plan

1. Write the pinned test `tests/update_closing_commit.rs` first; it fails on the tip because `update.md` has none of these sentences.
2. Add step 8 (the tip already has a step 7, the leftover-space preview), the reading paragraph and the completion-check sentence to `workflows/update.md`.
3. Set the manifest row of `workflows/update.md` to the file's sha256 (`shasum -a 256`, one edit, no Python); run the new test, `source_manifest_facts`, `update_route`, `agent_defs_package`, `product_cli skill_package` and every other test that names `workflows/update.md`.
4. Ask for an independent audit; the implementer does not audit.

## Risks

- A neighbouring card edits `workflows/update.md` too (the harness label, the always-ask paragraph): the pair line names the exact sentence each adds, and whoever lands second merges by sentence.
- The text names the `go-residue` step before its code lands: it is written as "when the plan has the step", so it stays true either way; a pinned test checks the condition words.
- A scan that pins sentence text reads the file by path and checks named words only; a reworded sentence has to keep them.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | The closing step is step 8 of the update procedure and is run after step 5 | Appending keeps the numbers of steps 1 to 7 that other tests and notes cite (the tip already has step 7, the leftover-space preview); the text says to run it before the final report | Move the sentence and the pinned order check |
| D-2 | On m-flow the commit goes on `spec/<slug>-<yyyymmdd>` made with `git switch -c`, not with `moai spec take` | `spec take` refuses while the work tree has a change outside the SPEC's folder, which is the state after an update | Change one sentence and its pin |
| D-3 | No push, no pull request | The project route's closing step does not push either | Add a separate offer |
| D-4 | A `go-residue` or `claude-md` step is read only when the plan has it | The detection code is a separate slice (read on 2026-11-25: `src/` has no `go-residue` step yet) | Drop the condition when the step exists everywhere |
| D-5 | The `claude-md` reading names two choices, the managed `CLAUDE.md` (recommended) and leave, and never "import only" | The product removed the import pick from the picks step (`agents_picks_step` offers `move`, `leave` and `remove-originals`) | Add the choice back with its pin if the product offers it again |
| D-6 | The closing offer shows the proposed commit as a dry-run diff first, and never amends or sweeps in the user's other changes | The project route's closing step carries the same rules; "the same closing commit offer" is the requirement. The `effect-intend` markers of that step are not repeated: no operation marker exists for the update route | Drop the sentence and its two pins |
| D-7 | The verify line of AC-5 is the Rust test `source_manifest_facts`, and the manifest row is set from `shasum -a 256` | The project allows no Python of its own; the Python tool and its merge driver stay outside this card | Restore the earlier line only when the tool is a `moai` command |

## Result

Filled in the documents step from the work tree on top of the integration tip `71d819ce8a` (HEAD `71d819ce8a4e7f5c3fc1805b8befb485f3712f81`); the change is not committed yet, so no commit id is named here. An earlier attempt of 2026-11-21 on the branch `spec/update-closing-commit-20261121` reached a ready note without an audit verdict; its base moved and this SPEC replaces it. Evidence files are in this worktree's `target/notes/`: `red.out` and `green.out` are cargo output, quoted only as they print.

### Changed

- `skills/moai/workflows/update.md`: step 8 "Offer the closing commit" after the leftover-space step 7, the paragraph "Reading what a Go-generation update left", and one sentence at the end of the Completion Check that names the closing commit as committed, left uncommitted by the user's pick, or not offered with its reason.
- `tests/update_closing_commit.rs`: four tests (`closing_commit_is_the_last_step`, `closing_commit_lists_marks_and_warns`, `m_flow_commits_on_a_branch_of_its_own`, `go_generation_leftovers_are_read_out`). The order check names step 8 after steps 5, 6 and 7; new pins are the dry-run diff sentence, "an existing commit is never amended", and the `claude-md` choices without "import only" (a check that the Reading paragraph does not contain those words).
- `skills/moai/references/source-manifest.json`: the one sha256 row of `workflows/update.md`, set to `c55a1e45486a5fe872f72b9b80a4dfeda20867fe84c42b0663cdf288bdce83c0`. `shasum -a 256 skills/moai/workflows/update.md` prints that value in this step, and `target/notes/update.new.md` is byte for byte the same file (`cmp`).
- `tools/repo_scans.json`: three lines of the `skill-package-agents` entry (the test in the covered list, its check command, and its rule path).

### Red and green

- Red: `cargo test --locked --offline --test update_closing_commit` on the tip text of `update.md` (the new text was set aside first) exited 101; `red.out` prints `test result: FAILED. 0 passed; 4 failed; 0 ignored; 0 measured; 0 filtered out`. Each test names the missing sentence; the first reads "update.md is missing" followed by the step 8 heading.
- Green: the same command plus `--test source_manifest_facts` on the new text exited 0; `green.out` prints `test result: ok. 8 passed; 0 failed` for `source_manifest_facts` and `test result: ok. 4 passed; 0 failed` for `update_closing_commit`. Both runs went through the heavy slot, and its last line prints the exit code.

### Acceptance

- AC-1: met — `closing_commit_is_the_last_step` ok in `green.out`.
- AC-2: met — `closing_commit_lists_marks_and_warns` ok in `green.out`.
- AC-3: met — `m_flow_commits_on_a_branch_of_its_own` ok in `green.out`.
- AC-4: met — `go_generation_leftovers_are_read_out` ok in `green.out`.
- AC-5: met — `source_manifest_facts`, 8 passed in `green.out` (`this_repository_manifest_is_what_the_tool_writes` is one of them).

The four AC-1 to AC-4 tests ran as the whole `update_closing_commit` test target, not as four separate runs with a name filter.

### Not run

- `moai check affected --card <base>..<head>`: not run here; it is run at submit and quoted in the ready note.
- Mutants of the new text: none run. The five mutants of the earlier attempt were on a different base and are not reused.
- The route on a real Go-generation project in a host: not run, so how the steps read in a live `/moai update` is not measured.
- The `go-residue` and `claude-md` code slices: other changes, not part of this one; the text reads them only when the plan has the step, and `src/` carried no `go-residue` step when D-4 was written.
- Other tests that name `workflows/update.md` (`update_route`, `agent_defs_package`, `product_cli`) and `repo_scans_table`: not run here; the affected-set receipt covers them.
