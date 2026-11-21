---
id: SPEC-20261121-skill-text-branch-pr-base
title: "the shipped skill text names no card branch, no card line in a pull request and no moai-adk integration branch"
status: draft
revision: 1
created: 2026-11-21
---

# the shipped skill text names no card branch, no card line in a pull request and no moai-adk integration branch

## Intent

The skill that ships to every user project still carries three wordings of moai-adk's own development: the claim step of `skills/moai/workflows/mode-agent.md` tells a lane to work on a branch `card/<n>-<slug>` in a worktree `.moai/worktrees/<project>-<n>`, its pull-request step asks for a second body line `card: C-<n>`, and the pull-request step of `skills/moai/workflows/sync.md` opens every pull request with `--base v2026.11.0`. A user project has no such branch, its rules say a card number never goes into a SPEC, a branch, a commit or a pull request, and its profile names its own integration branch. The text now names the branch the SPEC tool gives, lets the project's own rule come first, puts the SPEC id in the `Moai-Spec` trailer instead of a card line, and reads the base from the profile. A repository scan keeps the three wordings out of `skills/`.

## Requirements

- REQ-1: The shipped text under `skills/` shall carry no branch name of the form `card/<n>`, no pull request line `card: C-<n>` and no `--base v2026.11.0`. · failure: none — it is a text change and the scan of REQ-6 and REQ-7 refuses a return of any of them.
- REQ-2: The claim step of `mode-agent.md` shall tell a lane to create its worktree under `.moai/worktrees/<project>/<name>/` on the branch that `moai spec take <SPEC-ID>` names and to pass that branch to `moai card take --branch`. · failure: none — it is a text change and its test reads the sentence.
- REQ-3: The claim step of `mode-agent.md` shall say that the project's own rule comes first when the project's guidance or profile names the worktree place or the branch differently. · failure: none — it is a text change and its test reads the sentence.
- REQ-4: The pull-request step of `mode-agent.md` shall put the SPEC id in the `Moai-Spec` trailer and carry no card line. · failure: none — it is a text change and its test reads the sentence.
- REQ-5: The pull-request step of `sync.md` shall name the base as the integration branch of the profile (`work_flow.integration_branch`). · failure: none — it is a text change and its test reads the sentence.
- REQ-6: The scan test shall read every file under `skills/` and fail when a file carries a pattern of REQ-1.
- REQ-7 (REQ-6): IF a file carries a pattern, THEN the scan test shall name the file, the line and the pattern.

## Acceptance

- AC-1 (REQ-1, REQ-6, REQ-7): the scan finds each of the three patterns in a sample line, leaves the allowed forms alone (the word card in prose, `spec/<slug>-<yyyymmdd>`, `--base <integration branch>`), names the file, the line and the pattern, and passes on the shipped `skills/` · verify: `cargo test --locked --offline --test skill_branch_pr_text scan_`
- AC-2 (REQ-2, REQ-3, REQ-4): step 5 of `mode-agent.md` names `moai spec take`, the worktree place `.moai/worktrees/<project>/<name>/` and the sentence that the project's own rule comes first, and step 8 names the `Moai-Spec` trailer and no card line · verify: `cargo test --locked --offline --test skill_branch_pr_text claim_`
- AC-3 (REQ-5): the pull-request step of `sync.md` names `work_flow.integration_branch` as the base and no fixed branch · verify: `cargo test --locked --offline --test skill_branch_pr_text sync_`
- AC-4 (REQ-1, REQ-2, REQ-3, REQ-4, REQ-5): the tests that pinned the old wording moved with the text and pass · verify: `cargo test --locked --offline --no-fail-fast --test lane_rules_2879 --test mode_cards_53 --test mode_workflow_steps --test mode_guidance_scan --test skill_text_citations`

## Scope

- write: `.moai/specs/SPEC-20261121-skill-text-branch-pr-base/**`
- write: `tests/skill_branch_pr_text.rs` (new), `tests/lane_rules_2879.rs`
- write (held by other cards until their leases end or they land; asked by a note on each): `skills/moai/workflows/mode-agent.md`, `tests/mode_cards_53.rs`, `tests/mode_workflow_steps.rs`, `tests/mode_guidance_scan.rs` (held by one card), `skills/moai/workflows/sync.md` (another card)
- generated: `skills/moai/references/source-manifest.json` is regenerated with `tools/source_manifest.py write` at the merge; it is not edited by hand

## Decisions

- D-1: The acceptance of the card says the step points at `moai spec take` and `moai worktree new`. `moai worktree new` does not exist at the tip: the commands are `worktree-plan` and `worktree-apply` (they need an issue-claim record and a branch prefix the user decides), and `moai spec take` creates or switches to the branch `spec/<slug>-<yyyymmdd>` in the current worktree and creates no worktree. The text says what is true: the branch is the name `moai spec take <SPEC-ID>` gives, the worktree is made with `git worktree add` under `.moai/worktrees/<project>/<name>/`. A command that creates the worktree is a separate card. Recorded as a decision note on the card; reversible.
- D-2: The words "the project's own rule comes first" are one marker sentence, so a test and a reader find it in one place; the project's guidance file is `AGENTS.local.md` and its profile is read by the existing profile commands, not by this change.
- D-3: The scan covers `skills/` only. Documents of moai-adk's own development (`docs/`, `.moai/`) and the installed copies outside the repository are not shipped from here and are outside it.

## Plan

1. Tests first (RED): `tests/skill_branch_pr_text.rs` with the pattern function, its sample lines and the scan of `skills/`; see it fail on the three wordings, and the claim and sync tests fail on the unchanged text.
2. When the files are free: change step 5 and step 8 of `mode-agent.md` and the pull-request step of `sync.md`, move the pins in `lane_rules_2879.rs`, `mode_cards_53.rs`, `mode_workflow_steps.rs` and `mode_guidance_scan.rs`, then regenerate the manifest.
3. Run the new file, the tests that read the changed text (a search of the old wording in `tests/` and `tools/`), the repository scans, `--lib`, `--bins`, clippy, rustfmt; then audit and write the Result.

## Result

- AC-1: not yet — the scan test is written and red on the unchanged text.
- AC-2: not yet — the claim and pull-request steps wait for `skills/moai/workflows/mode-agent.md`, held by another card.
- AC-3: not yet — the `sync.md` step waits for the file, held by another card.
- AC-4: not yet — the pins move with the text.
- closure_status: the text change, the audit and the whole-area run are still to come.
