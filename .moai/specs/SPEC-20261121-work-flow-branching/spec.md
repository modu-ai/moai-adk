---
id: SPEC-20261121-work-flow-branching
title: "The lane workflow branches on the profile's work flow: m-flow has no pull request and no origin base"
status: draft
revision: 1
created: 2026-11-21
---

# The lane workflow branches on the profile's work flow: m-flow has no pull request and no origin base

## Intent

`skills/moai/workflows/mode-agent.md` describes one path only, the pull request path: the `state` step compares `HEAD` with `origin/<base branch>`, the `claim` step starts the worktree from `origin/<base>` on a branch named with the card number, `pr-gate` opens a pull request with `gh pr create`, `merge-train` waits for the GitHub `ci`, and `hand-off` links the pull request with `card pr` and closes the card after the pull request merged. `skills/moai/workflows/sync.md` step 5 describes the other path, M-Flow (user decisions of 2026-11-20): the lane pushes its own branch, merges the local integration branch tip in its own worktree, posts `MERGE-READY`, the leader fast-forwards, a pull request is opened only for a formal release, and a card that landed locally is closed on the local landing evidence. A lane in an m-flow project that follows `mode-agent.md` can therefore work from `origin`, open a pull request per card, or force-push a branch. Cause, inferred and not verified: one lane of an m-flow project force-pushed a branch and the lane workflow was the text it followed; the SPEC does not claim that this was the cause of that incident. This SPEC makes the five steps (`state`, `claim`, `pr-gate`, `merge-train` and `hand-off`, called the branching steps below) and the introduction line that says pull requests stay on GitHub branch on the project's `work_flow`: an m-flow clause for each, the existing pull request sentences kept word for word under the label for the pull request flows, and one pointer to the canonical M-Flow text in `sync.md` step 5, which is not restated. Two related defects in the same lines are fixed here by the leader's decision: the branch name that carries the card number (`card/<n>-<slug>`) becomes `spec/<slug>-<yyyymmdd>` for every flow, and the m-flow description in `references/internal/interview.md` that says "one pull request" per SPEC is corrected to the M-Flow rule (a pull request only for a formal release). It changes skill text and tests only.

## Non-goals

- `workflows/sync.md` and `workflows/mode-leader.md`: the M-Flow text there stays the canonical one.
- Flows other than `m-flow` and `github-flow` (`trunk-based`, `git-flow`): they follow the pull request path unchanged; naming their own paths is a later SPEC.
- Changing how a card is closed in code (`src/card/work.rs`, `src/ops/card.rs`) or what `work_flow` records.
- The stage order, the report rules and the loop rule, which are other slices of the same card.

## Requirements

- REQ-1: The `state` step shall tell a lane to read the project's work flow once with `moai work-flow-plan --project .` and to take `profile.flow` of its answer as the flow.
- REQ-2: The `state` step shall tell an m-flow lane to compare `HEAD` with the tip of the profile's integration branch in its own worktree and not with `origin/<base branch>`.
- REQ-3: The `claim` step shall tell an m-flow lane to create its worktree from the tip of the profile's integration branch and not from `origin/<base>`.
- REQ-4: The `claim` step shall name the branch of a card `spec/<slug>-<yyyymmdd>` for every flow, with no card number in it.
- REQ-5: The `pr-gate` step shall tell an m-flow lane to open no pull request, to push its branch and to post the `READY` note.
- REQ-6: The `merge-train` step shall give an m-flow lane a pointer to step 5 of the sync workflow instead of a restatement of that step.
- REQ-7: The `hand-off` step shall tell an m-flow lane to close its card on the local landing evidence and to force-push no branch.
- REQ-8: Each branching step shall keep its pull request sentences word for word under the label `github-flow and the other pull request flows`, except for the branch name of REQ-4.
- REQ-9: The introduction line that says pull requests stay on GitHub shall say that a project in m-flow opens a pull request only for a formal release.
- REQ-10: The m-flow entry of `references/internal/interview.md` shall say that a pull request is opened only for a formal release and not for each SPEC.
- REQ-11: The skill's source manifest shall match the files after the change.

## Acceptance

- AC-1 (REQ-1, REQ-2): the `state` step reads the flow with the command and gives the m-flow base without `origin/<base`.
  verify: `cargo test --locked --offline --test work_flow_branching the_state_step_reads_the_flow_and_gives_m_flow_a_local_base`
- AC-2 (REQ-3, REQ-4): the `claim` step starts an m-flow worktree from the local integration tip and no step names a branch with the card number.
  verify: `cargo test --locked --offline --test work_flow_branching the_claim_step_starts_from_the_local_tip_and_names_no_card_number`
- AC-3 (REQ-5, REQ-6, REQ-7): `pr-gate`, `merge-train` and `hand-off` each have an m-flow clause; the m-flow clauses carry no `gh pr`, `card pr` or `origin/<` instruction, and the `hand-off` clause forbids a force push.
  verify: `cargo test --locked --offline --test work_flow_branching the_m_flow_clauses_carry_no_pull_request_or_origin_instruction`
- AC-4 (REQ-8): every `origin/<base>`, `gh pr create` and `card pr` of the five steps sits after the label for the pull request flows.
  verify: `cargo test --locked --offline --test work_flow_branching the_pull_request_sentences_stay_under_their_label`
- AC-5 (REQ-9, REQ-10): the introduction line and the interview entry say a pull request only for a formal release.
  verify: `cargo test --locked --offline --test work_flow_branching the_pull_request_only_at_a_formal_release_is_said_in_both_places`
- AC-6 (REQ-6): the merge-train pointer names the sync workflow and the M-Flow text is not restated in the lane workflow.
  verify: `cargo test --locked --offline --test work_flow_branching the_m_flow_merge_text_is_pointed_to_and_not_copied`
- AC-7 (REQ-8): the tests that pin these steps still pass (the pull request sentences are unchanged).
  verify: `cargo test --locked --offline --test mode_workflow_steps --test mode_cards_53 --test lane_rules_2879 --test lane_rules_2401 --test lane_rules_2836 --test mode_guidance_scan --test flaky_followups --test agent_defs_package --test close_step_2903 --test claim_renew_doc_3042`
- AC-8 (REQ-11): the manifest matches the package.
  verify: `python3 -I tools/source_manifest.py check`

## Scope

- write: `.moai/specs/SPEC-20261121-work-flow-branching/**`, `tests/work_flow_branching.rs`, `skills/moai/workflows/mode-agent.md`, `skills/moai/references/internal/interview.md`, `skills/moai/references/source-manifest.json` (rows of the two skill files only, by the leader's pair approval)
- read-only: `skills/moai/workflows/sync.md`, `skills/moai/workflows/mode-leader.md`, `src/profile.rs`, `src/work_flow.rs`, every test that pins a sentence of the five steps
- forbidden: a card, Issue or pull request number in the sentences this change adds, a change to any sentence of the pull request path other than the branch name, and an edit of a file another card holds before that card has released it or agreed in a pair line

## Plan

1. Write the test first: `tests/work_flow_branching.rs` reads the lane workflow and the interview reference by path; it fails on the tip because no step names the flow, the claim step names `card/<n>-<slug>` and the interview entry says "one pull request".
2. Add the m-flow clause and the label to each of the five steps and to the introduction line; rename the branch; correct the interview entry. The pull request sentences keep their bytes, which the pinned tests confirm.
3. Run the new test and every test that reads the lane workflow without a name filter; if a pinned sentence names the old branch, change that test in the same commit and say so; regenerate the manifest rows; mutants (at most five); ask for an independent audit.

## Risks

- The lane workflow is read by many tests; they cut a step as one physical line, so the clauses stay on the step's line and the pull request sentences are not split.
- The branch rename touches pinned wording (`card/<n>-<slug>` appears three times in the `claim` step): the tests that pin it are changed together with the text; the AC-7 list is the set to run.
- A failed or empty `work-flow-plan` read is not a flow. The lane reports the flow as unread and does not guess (D-3); a project that never recorded a flow follows the pull request path as before.
- Inference, not verified: that the lane workflow caused the force push of the incident the leader described.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | Step numbers and names stay; each branching step gets a clause, not a new step | `procedure_keys` and the step tests pin the list | Add a flow step in a later SPEC with the tests |
| D-2 | The pull request sentences stay byte for byte under a label; only the branch name changes | The pinned tests and the audit rules read those words | Reword in a later SPEC |
| D-3 | A failed or empty read of the flow is reported unread and not guessed; a profile without `work_flow` follows the pull request path | The interview asks for the flow only when none is recorded; guessing m-flow would drop the CI gate | Make the unread case a blocked card |
| D-4 | The m-flow clauses point to `sync.md` step 5 and do not copy it | One canonical text, put once per file (user decisions of 2026-11-20) | Copy the text and keep both in step by a test |
| D-5 | The branch name is `spec/<slug>-<yyyymmdd>` in both flows | Leader decision: no card number in a branch; the m-flow definition already uses it | Name it with the card number again |
| D-6 | The interview m-flow entry is corrected in this SPEC | Leader decision; it contradicts `sync.md` step 5 | Revert the entry |

## Result

(written after the checks)
