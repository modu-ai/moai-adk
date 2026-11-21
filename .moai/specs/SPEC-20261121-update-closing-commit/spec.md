---
id: SPEC-20261121-update-closing-commit
title: "The update route ends with the closing commit offer and a reading of what a Go-generation update left"
status: draft
revision: 1
created: 2026-11-21
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
  verify: `python3 -I tools/source_manifest.py check`

## Scope

- write: `.moai/specs/SPEC-20261121-update-closing-commit/**`, `tests/update_closing_commit.rs`, `skills/moai/workflows/update.md`, `skills/moai/references/source-manifest.json` (regenerated after the edit)
- read-only: `skills/moai/workflows/project.md`, `src/project_update.rs`, `tests/update_route.rs`
- forbidden: a card, Issue or pull request number in the sentences this change adds, and any edit of a file another card holds before that card has released it or agreed in a pair line

## Plan

1. Write the pinned test `tests/update_closing_commit.rs` first; it fails on the tip because `update.md` has none of these sentences.
2. Take `workflows/update.md` and the source manifest when their holders release them; add step 7 and the completion-check sentence.
3. Regenerate the manifest row with `python3 -I tools/source_manifest.py write`; run the new test, `update_route`, `agent_defs_package`, `product_cli skill_package` and every other test that names `workflows/update.md`.
4. Ask for an independent audit; the implementer does not audit.

## Risks

- A neighbouring card edits `workflows/update.md` too (the harness label, the always-ask paragraph): the pair line names the exact sentence each adds, and whoever lands second merges by sentence.
- The text names the `go-residue` step before its code lands: it is written as "when the plan has the step", so it stays true either way; a pinned test checks the condition words.
- A scan that pins sentence text reads the file by path and checks named words only; a reworded sentence has to keep them.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | The closing step is step 7 of the update procedure and is run after step 5 | Appending keeps the numbers of steps 1 to 6 that other tests and notes cite; the text says to run it before the final report | Move the sentence and the pinned order check |
| D-2 | On m-flow the commit goes on `spec/<slug>-<yyyymmdd>` made with `git switch -c`, not with `moai spec take` | `spec take` refuses while the work tree has a change outside the SPEC's folder, which is the state after an update | Change one sentence and its pin |
| D-3 | No push, no pull request | The project route's closing step does not push either | Add a separate offer |
| D-4 | A `go-residue` or `claude-md` step is read only when the plan has it | The detection code is a separate slice | Drop the condition when the step exists everywhere |

## Result

Filled in by the implementation lane on 2026-11-21; the audit checks it again. Base `53cce6c580` (the tip when the SPEC was written); the green text is committed as `b5c0de7d9c`, and the local tip `174d8fcecf9a` is merged in as `d9190111e8` (tree equal to the `git merge-tree` result, no conflict). Evidence files are in this worktree's `target/notes/`.

Red (before the text): the previous lane pinned the four tests and showed each one red on the tip text by a script over the file (no cargo run then); the first cargo run is the green one below.

- AC-1: met — `cargo test --locked --offline --test update_closing_commit` (`green1.out`): 4 passed, 0 failed, exit 0; `closing_commit_is_the_last_step` is one of the four. Step 7 sits after steps 5 and 6 and before `## Exceptions and Permissions`; the completion check ends with the closing-commit sentence.
- AC-2: met — same run, `closing_commit_lists_marks_and_warns` (listing under the apply, `git ls-files` marking, kept-local folders, the "comes back" warning, the two-choice question, index read, per-path `git add`, `git commit -- <path>`, the way back, no push, secret-looking names).
- AC-3: met — same run, `m_flow_commits_on_a_branch_of_its_own` (also asserts the text does not use `moai spec take`).
- AC-4: met — same run, `go_generation_leftovers_are_read_out` (the four residue classes, `tracked`/`resurrected`, the `claude-md` choices, the leftovers).
- AC-5: met — `python3 -I tools/source_manifest.py check` prints `{"conflicted": false, "changed": [], "dropped": [], "ok": true}` on the merged tree; `cargo test --locked --offline --test update_route --test agent_defs_package --test product_cli` (same `green1.out`): 9, 19 and 79 passed.

### Checks

`moai check affected 174d8fcecf9a..HEAD` (the branch's own build of the binary, `--max 2`, head `d9190111e8`), summary line as printed: `check-affected base=174d8fcecf9a head=d9190111e862 targets=42 py=8 scans=listed:41 diff_check=0 fmt=0 clippy=0 test_lib=0 test_bins=0 test_targets=0 build_cli=0 python=1 result=FAIL`. The one failing step is `tests.repo_scans_test … test_every_walking_test_file_is_listed_or_exempt` (log `target/notes/affected/08-python.log`): `tests/update_closing_commit.rs` is read as a test that walks the repository (it names `CARGO_MANIFEST_DIR` and the pinned words `git ls-files`), while it reads only the one named `skills/moai/workflows/update.md`. The way out is one line in the EXEMPT table of `tests/repo_scans_test.py`, a file another change holds; the lane asked the leader how it lands (card note on the decision). Until that line is in, this SPEC's own checks are green and the full check stays red on that one step.

Mutants (five, on the committed `skills/moai/workflows/update.md`, each restored, `restored: True`; `target/notes/mutate.out`): the "comes back" warning dropped, `git commit -- <path>` turned into a plain `git commit`, the m-flow branch rule dropped, the `go_marker` class dropped, the completion sentence dropped — each killed by the named test of the matching AC.

### Not run

- The route on a real Go-generation project (`/moai update` in a host): not run; the text is pinned by words only, and the host check stays `not_run` as the workflow says.
- The code slices this text names (`go-residue`, `claude-md` steps): they are other changes; the text reads them only "when the plan has the step".
- A second run of the same tests on the same tree: not repeated (each result above is the one run's output).
