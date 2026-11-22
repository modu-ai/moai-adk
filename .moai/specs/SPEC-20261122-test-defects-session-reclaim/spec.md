---
id: SPEC-20261122-test-defects-session-reclaim
title: "Two defects: the registry tests read the shell's session id, and card reclaim counts the runtime state folder as work"
status: draft
revision: 1
created: 2026-11-22
---

# Two defects: the registry tests read the shell's session id, and card reclaim counts the runtime state folder as work

## Intent

Every lane shell turns `moai check affected` red because the session-register tests pass the shell's `CLAUDE_CODE_SESSION_ID` to the binary they start, and a dead holder cannot be reclaimed while its worktree holds untracked files in the runtime state folder, which are tool state and not work.

## Non-goals

- No change to what `session-register` records: the `session` key from the session id stays; only the tests stop inheriting the shell's value.
- No change to how the reclaim gate reads commits: unpushed work and the process check stay as they are.

## Requirements

- REQ-1: When a test of `session-register` starts the moai binary, the child process shall run without the `CLAUDE_CODE_SESSION_ID` variable of the shell that runs the test.
- REQ-2 (REQ-1): IF a child process of the session-register tests would still inherit the variable, THEN a test shall fail and name that child.
- REQ-3: When `card reclaim` counts the uncommitted work of a recorded worktree, the count shall leave out every untracked file under the runtime state folder of that worktree.
- REQ-4 (REQ-3): IF a tracked file is changed or an untracked file lies outside the runtime state folder, THEN the reclaim gate shall count it as work.

## Acceptance

- AC-1 (REQ-1, REQ-2): every child process that `tests/session_registry_cli.rs` starts has the variable removed from its environment, and the whole target passes with the variable set in the shell · verify: `cargo test --offline --locked --test session_registry_cli session_registry_children_`
- AC-2 (REQ-3): a worktree whose only untracked files are under the runtime state folder (the counts file, another file beside it, a file in a folder below) is reclaimed and every file stays · verify: `cargo test --offline --locked --test card_reclaim_early_cli a_runtime_state_folder_is_not_work`
- AC-3 (REQ-4): an untracked file outside the state folder and a changed tracked file each keep the refusal and name one uncommitted change, with a state folder file beside them · verify: `cargo test --offline --locked --test card_reclaim_early_cli a_listed_tool_cache_is_not_work`

## Scope

- write: `tests/session_registry_cli.rs`, `tests/card_reclaim_early_cli.rs`, `src/card/early_reclaim.rs` (the entry rule and its comment only), this SPEC folder

## Plan

1. REQ-1 and REQ-2: a red test that reads the environment of the command the tests build (`Command::get_envs`), then the variable is removed in the one helper that builds it.
2. REQ-3 and REQ-4: the red test first, then the entry rule of `is_tool_cache_entry` names the state folder instead of one file; the older test that pinned "another name in the same folder is work" is rewritten to the new rule and keeps its other three cases.

## Decisions

| Id | Decision | Source |
|---|---|---|
| D-1 | The whole runtime state folder is tool state; the list of one file name is widened to the folder. Only untracked files count as cache; a tracked file there that changed is still work. | leader note 5653 ("leave the runtime state folder out of the remaining-work judgement, or put it in .gitignore") |
| D-2 | The test clears the variable itself (the second option of note 5660); the expected entry does not change. | leader note 5660 |

### Research

| Id | Fact | Read |
|---|---|---|
| R-1 | With the real session id in the shell, `register_mode_is_stored_and_renewed` fails: the entry gets a `session` key beside `schema_version`, `name`, `instance`, `lease_until`, `registered_at`. With the variable unset or a short test value the target passes (22 passed). | run in the defects worktree, 2026-11-22 |
| R-2 | `is_tool_cache_entry` in `src/card/early_reclaim.rs` accepts only `state/github/counts.json` of the current state folder and `.moai/state/github/counts.json` of the earlier one; the test `a_listed_tool_cache_is_not_work` pins that a second name in the same folder is work. | read in the tree at the integration tip |

## Risks

- A person who keeps an untracked file of their own under the state folder loses the refusal for it; the folder is tool state by definition and the file is untracked, so a changed tracked file stays work (REQ-4).
- The entry rule lives in a file another card reserves for a flow rename; the edit is one rule and its comment and the other card's hunks are not touched.

## Result

- AC-1: met. `session_registry_children_do_not_inherit_the_shell_session_id` was red first (the command left the variable to the shell) and passes; the whole `session_registry_cli` target passes with the real session id set in the shell (23 passed); 2 mutants killed (the removal dropped, another variable name removed).
- AC-2: met. `a_runtime_state_folder_is_not_work` was red first (holds 3 uncommitted change(s)) and passes: the counts file, a file beside it, a file two folders below and the earlier generation's state folder are reclaimed and every file stays.
- AC-3: met. `a_listed_tool_cache_is_not_work` keeps its refusals for an untracked file outside the state folder and a changed tracked file, one uncommitted change each; the unit test `only_an_untracked_state_folder_file_is_a_cache` was rewritten red first; 3 mutants killed (no trailing slash, the earlier folder dropped, the untracked check dropped).
- The rule lives in `src/card/early_reclaim.rs`, a file another card reserves; its holder answered that the hunk does not overlap theirs (note 6346), and the edit is by explicit path.
