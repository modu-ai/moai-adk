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
- REQ-8: The PROJECT workflow shall ask the `work_flow` row with GitHub flow first and recommended and with m-flow described to the user as the choice for when several AI agents work at the same time.

## Acceptance

- AC-1 (REQ-1, REQ-2, REQ-3): the options cell of the `work_flow` row starts with `GitHub flow (recommended)`, holds the phrase `when several AI agents work at the same time` in the m-flow option, still holds trunk-based and git-flow, and holds none of the five words · verify: `cargo test --test generic_flow_questions work_flow_row`
- AC-2 (REQ-4, REQ-5): the section `## Git work flow` names GitHub flow as the recommended default, no longer calls m-flow the recommended default, and holds one sentence that tells the model to describe m-flow in plain words and to keep SPEC, worktree, audit and leader words for after the pick · verify: `cargo test --test generic_flow_questions flow_section`
- AC-3 (REQ-6, REQ-7): the question cell and the options cell of the `autonomy` row hold `hard to undo, or touches money, secrets or deletion` and neither `case (2)` nor `SKILL.md`; the sorting rule for the model is a sentence outside the table · verify: `cargo test --test generic_flow_questions autonomy_row`
- AC-4 (REQ-8): `skills/moai/workflows/project.md` names GitHub flow first and recommended where it asks the `work_flow` row, describes m-flow there as `when several AI agents work at the same time`, and no longer says m-flow first and recommended · verify: `cargo test --test generic_flow_questions project_workflow`
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

Built tests first. The RED commit held nine tests and its text was the tip's: `0 passed; 9 failed`, all assertion failures (the audit reproduced it); the tenth test came later with a mutant that survived. The ten tests of the head's file against the tip's text give `0 passed; 10 failed`, and against the changed question bank and profile step `10 passed`. The holders of `skills/moai/workflows/project.md` and the source manifest were released, so the sentence and the manifest entries are written in this branch (`tools/source_manifest.py write`, then `check` exit 0) and no pair line is left open.

- AC-1: met — `work_flow_row_lists_github_flow_first_and_recommended`, `work_flow_row_describes_m_flow_in_plain_words` and `work_flow_row_uses_no_maintainer_words` pass: the options start with `GitHub flow (recommended)`, the m-flow option says `when several AI agents work at the same time`, trunk-based and git-flow remain, and none of SPEC, worktree, audit, leader or card appears.
- AC-2: met — `flow_section_names_github_flow_the_recommended_default`, `only_the_github_flow_bullet_carries_the_recommended_default` and `flow_section_tells_the_model_to_describe_m_flow_in_plain_words_first` pass.
- AC-3: met — `autonomy_row_says_what_is_asked_first_in_plain_words`, `autonomy_row_shows_the_user_no_internal_terms` and `autonomy_sorting_rule_for_the_model_stands_outside_the_table` pass.
- AC-4: met — `project_workflow_asks_the_work_flow_row_with_github_flow_first` passes: the profile step names GitHub flow first and recommended, describes m-flow as `when several AI agents work at the same time`, and no longer says m-flow first and recommended.
- AC-5: met — `agent_defs_package` 19, `spec_branch_name_first` 13 and `project_decide_model` 4 pass with the moved assertion; `project_preview_order` 3, `update_route` 9 and `product_cli skill_package` 18 pass; the shipped-text ratchet (`tools/shipped_text_ratchet.py check`, run from the worktree of the change that adds it; it is not in this branch's tree) exits 0.

Mutants (16, one rule each, on the committed files with every file restored after each run): the recommended mark dropped, m-flow first again, the plain description of m-flow dropped, trunk-based dropped, git-flow dropped, the word audit in the options, the flow section's default dropped, m-flow called the default, the describe-in-plain-words sentence dropped, `case (2)` in the autonomy options, `SKILL.md` in the autonomy question, the sorting sentence dropped, and four on the profile step (m-flow first again, the description dropped, the description in maintainer words, GitHub flow out of the order). The first run left two alive (the flow section's default dropped, m-flow called the default), because nothing read which bullet carries `the recommended default`; `only_the_github_flow_bullet_carries_the_recommended_default` closes both (a pass-only pin), and the second run caught 16 of 16 with the files identical afterwards and 0 changed files. The description in the profile step was not pinned either, so REQ-8 now names it and the profile-step test reads it.

First repair round (the independent audit of head 6e423aa933: no blocking finding; 45 mutants, 33 caught, 12 alive). Four survivors were cheap to close and are pinned now: the autonomy options lost `is asked first` with the phrase kept (the autonomy test reads `deletion, is asked first`), and a second option marked `(recommended)` on trunk-based, m-flow or git-flow (the first work-flow test counts exactly one mark). Both pins are pass-only; the four mutants are caught on the second run with the files identical afterwards. The audit's other survivors stay open as limits of a text test: a negated or inverted instruction that keeps its pinned phrase, a qualified default, an anchor line that is not the real bullet, trunk-based and git-flow dropped from the profile step (outside REQ-8), and the word SPEC in the question cell (REQ-3 names the options cell). Named, not changed: the test `work_flow_row_offers_m_flow_first_and_the_bank_states_each_flow` in `tests/agent_defs_package.rs` now asserts GitHub flow first and keeps its old name, because an earlier SPEC's verify command selects it by that name; the stale comments in `src/profile.rs` and in that test file are not touched.

Limits: the tests prove the text, not that a live session asks the row in this order; host use stays not run until the installed skill carries the text. The stale comment in `src/profile.rs` that still calls m-flow the default is not touched (read-only here).
