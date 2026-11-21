---
id: SPEC-20261121-generic-flow-questions
title: "The project questions about the git flow and autonomy use plain words and recommend GitHub flow"
status: draft
revision: 1
created: 2026-11-21
---

# The project questions about the git flow and autonomy use plain words and recommend GitHub flow

## Intent

`/moai project` asks every user which git work flow the project follows and what may run without asking. The question bank (`skills/moai/references/internal/interview.md`) recommends m-flow, MoAI's own flow, and explains it with MoAI operations (a worktree and a branch per SPEC, an audit by another agent, a leader who merges). The `autonomy` row shows the user the internal words "SKILL.md sorts as case (2)". The install-flow report finds both (section 7, G7): a question for any user must not carry the maintainers' tooling terms or defaults.

The user decided (2026-11-20, "모두 다 카드 만들어서 진행하고, 각 어떻게 할 것인지 다시 물어보면 내가 결정"; the pick is on the card as a leader note): the `work_flow` recommendation is GitHub flow; m-flow stays as an option that is described in plain words, "when several AI agents work at the same time"; "case (2)" is replaced by plain wording, "hard to undo, or touches money, secrets or deletion".

The agent-facing detail stays: the flow section still states the four flows, the branch name rule and the Claude Code worktree base for an m-flow pick. Only what the user is shown changes: the question, the options and the order.

## Non-goals

- Removing m-flow, changing its rules, its branch names or the worktree base write that follows an m-flow pick.
- The other rows of the question bank (versioning is already generic).
- The rules of SKILL.md that use "case (2)" for the agent: they are agent text, not a question to the user.
- A `work_flow` default written into a profile without a question: the row is still asked when the profile has none.
- The stale doc comment on `Profile.work_flow` in `src/profile.rs` that says m-flow is the recommended default: a comment in a file this change does not hold; it is reported in the READY note.

## Requirements

- REQ-1: The options of the `work_flow` row shall list GitHub flow first and mark it recommended.
- REQ-2: The options of the `work_flow` row shall describe m-flow as the choice for when several AI agents work at the same time.
- REQ-3: The options of the `work_flow` row shall not use the words SPEC, worktree, audit, leader or card.
- REQ-4: The flow section of the question bank shall name GitHub flow the recommended default.
- REQ-5: The flow section of the question bank shall tell the model to describe m-flow to the user in plain words before the user picks it.
- REQ-6: The options of the `autonomy` row shall name what is asked first as what is hard to undo or touches money, secrets or deletion.
- REQ-7: The question and the options of the `autonomy` row shall not contain `case (2)` or `SKILL.md`.
- REQ-8: The PROJECT workflow shall ask the `work_flow` row with GitHub flow first and recommended.

## Acceptance

- AC-1 (REQ-1, REQ-2, REQ-3): the options cell of the `work_flow` row starts with `GitHub flow (recommended)`, holds the phrase `when several AI agents work at the same time` in the m-flow option, still holds trunk-based and git-flow, and holds none of the five words · verify: `cargo test --test generic_flow_questions work_flow_row`
- AC-2 (REQ-4, REQ-5): the section `## Git work flow` names GitHub flow as the recommended default, no longer calls m-flow the recommended default, and holds one sentence that tells the model to describe m-flow in plain words and to keep SPEC, worktree, audit and leader words for after the pick · verify: `cargo test --test generic_flow_questions flow_section`
- AC-3 (REQ-6, REQ-7): the question cell and the options cell of the `autonomy` row hold `hard to undo, or touches money, secrets or deletion` and neither `case (2)` nor `SKILL.md`; the sorting rule for the model is a sentence outside the table · verify: `cargo test --test generic_flow_questions autonomy_row`
- AC-4 (REQ-8): `skills/moai/workflows/project.md` names GitHub flow first and recommended where it asks the `work_flow` row and no longer says m-flow first and recommended · verify: `cargo test --test generic_flow_questions project_workflow`
- AC-5 (REQ-1, REQ-6): no regression, the existing pin test of the row reads the new order, and the tests that read the bank for branch names and the decision rows stay green · verify: `cargo test --test agent_defs_package --test spec_branch_name_first --test project_decide_model`

## Scope

- write: `skills/moai/references/internal/interview.md`, `tests/generic_flow_questions.rs` (new), `tests/agent_defs_package.rs` (the one assertion on the `work_flow` options), `.moai/specs/SPEC-20261121-generic-flow-questions/**`
- paired (written by the cards that hold these files, from a pair line, merged in one chain with this change): `skills/moai/workflows/project.md` (one sentence in the work-flow step), `skills/moai/references/source-manifest.json` (the sha256 entries of the two skill files)
- read-only: `skills/moai/SKILL.md`, `src/profile.rs`
- forbidden: the Go repository, the real HOME, any card number in the shipped text

## Plan

Tests first (RED). `tests/generic_flow_questions.rs` reads the question bank and the PROJECT workflow as text, like `tests/agent_defs_package.rs`: no command, host or HOME runs. The one assertion in `tests/agent_defs_package.rs` that pins "m-flow (recommended)" moves to the new order in the same RED commit. On the tip alone the new groups fail on assertions.

Then the text of the question bank: the `work_flow` options read `GitHub flow (recommended): short branches from the default branch, a pull request per change, merge after review and CI; m-flow, for when several AI agents work at the same time on one project; trunk-based; git-flow`; the `autonomy` options read `Reads, tests and edits inside the task's scope run without asking, and anything hard to undo, or touches money, secrets or deletion, is asked first (recommended); also ask before every commit; ask before every file edit` and the question loses its "Sort the answer the way SKILL.md …" sentence, which becomes one sentence under the table for the model. The flow section calls GitHub flow the recommended default and adds the plain-words sentence.

The PROJECT workflow sentence comes from its holder as a pair line.

## Risks

- The text tests that read the bank for other reasons (branch names, decision rows) must stay green; AC-5 runs them.
- The shipped text must not gain a maintainer token (a card number, an issue number, a repository link or the local-rules file name): no new text names one, and the ratchet count of the shipped paths must not rise; the check runs after the edit.
- `skills/moai` edits need their sha256 entries in the source manifest in the same change; the manifest is held by other cards, so the entries come with the merge chain from its holders (a pair line), and the READY names it.
- A person who already picked m-flow keeps it: a profile with a `work_flow` is not asked again.
- The recommendation moves for new answers only; nothing rewrites an existing profile.

## Result

(not yet written)
