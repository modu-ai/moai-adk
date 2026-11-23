---
id: SPEC-20261122-test-defects-session-reclaim
title: "Two defects: the registry tests read the shell's session id, and card reclaim counts the runtime state folder as work"
status: draft
revision: 4
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
- REQ-5: When the early reclaim counts the ignored entries of a recorded worktree, the count shall leave out every ignored file under the runtime state folder of that worktree.
- REQ-6 (REQ-5): IF an ignored file lies outside the runtime state folder and off the regenerable list, THEN the early reclaim shall refuse and name that file.
- REQ-7: When the Rust property harness checks its time budget, the budget shall grow with the one-minute load average above the number of logical CPUs of the machine.
- REQ-8 (REQ-7): IF a property run exceeds its budget, THEN the failure shall name the property, the seed, the elapsed time, the budget and the load it was scaled by.
- REQ-9: When the reclaim gate reads a recorded worktree through git, the gate shall take any warning git writes as an unreadable worktree.
- REQ-10 (REQ-9): IF git warns that it could not open a folder of the worktree, THEN the reclaim shall be refused as unreadable.
- REQ-11: When moai starts a git child, the child shall run without every variable whose name starts with `GIT_TRACE`.
- REQ-12 (REQ-11): IF the shell that runs `card reclaim` sets a trace variable, THEN the reclaim shall judge the worktree by its work and not by the trace.
- REQ-13: When the Rust property harness replays one case with `PROP_ONLY`, the index shall name a case of the run.
- REQ-14 (REQ-13): IF the index is at or beyond the case count, THEN the run shall fail and name the property, the index, the case count and the seed.

## Acceptance

- AC-1 (REQ-1, REQ-2): every child process that `tests/session_registry_cli.rs` starts has the variable removed from its environment, and the whole target passes with the variable set in the shell · verify: `cargo test --offline --locked --test session_registry_cli session_registry_children_`
- AC-2 (REQ-3): a worktree whose only untracked files are under the runtime state folder (the counts file, another file beside it, a file in a folder below) is reclaimed and every file stays · verify: `cargo test --offline --locked --test card_reclaim_early_cli a_runtime_state_folder_is_not_work`
- AC-3 (REQ-4): an untracked file outside the state folder and a changed tracked file each keep the refusal and name one uncommitted change, with a state folder file beside them · verify: `cargo test --offline --locked --test card_reclaim_early_cli a_listed_tool_cache_is_not_work`
- AC-4 (REQ-3): in a worktree on the earlier project layout, an untracked file under that layout's state folder is a cache and a file beside it is work, so the current-folder rule and the earlier-generation rule each decide on their own · verify: `cargo test --offline --locked --lib an_old_layout_runtime_state_folder_is_not_work`
- AC-5 (REQ-2): `tests/session_registry_cli.rs` builds every command for the binary under test in its one helper, and a test fails when the file builds one anywhere else · verify: `cargo test --offline --locked --test session_registry_cli every_child_`
- AC-6 (REQ-5, REQ-6): a worktree that ignores its runtime state folder and holds an ignored file there is reclaimed early and the file stays, while an ignored file outside that folder still refuses the reclaim · verify: `cargo test --offline --locked --test card_reclaim_early_cli an_ignored_state_folder_file_is_not_work`
- AC-7 (REQ-7, REQ-8): the budget of `tests/common/prop.rs` is 30 s when the load is at or below the CPU count or cannot be read, grows in proportion above it, and is capped; the over-budget message carries the seed, the elapsed time, the budget and the load · verify: `cargo test --offline --locked --test prop_reservation the_property_budget_`
- AC-8 (REQ-9, REQ-10): a worktree with an untracked folder that cannot be opened (mode 000) is refused as unreadable, where git exits 0 and only warns · verify: `cargo test --offline --locked --test card_reclaim_early_cli an_unreadable_folder_is_not_a_clean_worktree`
- AC-9 (REQ-3, REQ-4): an untracked file whose name holds a space is compared as written: under the runtime state folder it is tool state, outside it is one uncommitted change · verify: `cargo test --offline --locked --test card_reclaim_early_cli a_name_with_a_space_is_read_as_written`- AC-10 (REQ-11, REQ-12): the name test removes every documented trace variable and keeps an unrelated one, and a `card reclaim` run with `GIT_TRACE`, `GIT_TRACE2` and `GIT_TRACE_SETUP` set refuses a dirty worktree for its one change and reclaims a clean one · verify: `cargo nextest run --lib -E 'test(git_trace_variables_are_removed_by_prefix)'` and `cargo nextest run --test card_reclaim_early_cli -E 'test(a_git_trace_variable_does_not_make_a_worktree_unreadable)'`
- AC-11 (REQ-13, REQ-14): an index at or beyond the case count is refused with the property, the index, the case count and the seed; an index inside the run or none is accepted · verify: `cargo nextest run --test prop_reservation -E 'test(a_replay_index_beyond_the_cases_is_refused)'`

## Scope

- write: `tests/session_registry_cli.rs`, `tests/card_reclaim_early_cli.rs`, `src/card/early_reclaim.rs` (the entry rules, their comments and unit tests), `src/card/dispatch.rs` (the worktree gate and its git reads), `src/identity.rs` (the git environment scrub), `tests/common/prop.rs` and `tests/prop_reservation.rs` (the time budget, the replay index and the comments naming the property SPEC), this SPEC folder

## Plan

1. REQ-1 and REQ-2: a red test that reads the environment of the command the tests build (`Command::get_envs`), then the variable is removed in the one helper that builds it.
2. REQ-3 and REQ-4: the red test first, then the entry rule of `is_tool_cache_entry` names the state folder instead of one file; the older test that pinned "another name in the same folder is work" is rewritten to the new rule and keeps its other three cases.
3. The audit follow-ups (revision 2): the earlier-layout unit test (AC-4) and the one-helper test (AC-5) first, then REQ-5 red first in `card_reclaim_early_cli` and the ignored-entry rule of `is_regenerable_ignored` widened to the runtime state folder; the two comments that still name one listed cache file are brought up to the folder rule.
4. Slice 3 (revision 3, the leader's note on the flaky property budget and the audit of slice 2): the budget function and its unit tests in `tests/common/prop.rs` first (red), then `run_property` checks the elapsed time against it and names the load in its message.
5. Slice 3, the audit of slice 2: AC-8 and AC-9 red first, then the gate's git reads go through one helper that refuses on a warning and read paths NUL-separated (`-z`), and the state folder is resolved once per check instead of once per entry.
6. Slice 4 (revision 4, the audit of slice 3), the part that needs no write in `src/card/dispatch.rs`: AC-10 and AC-11 red first (the name test and the CLI test with the trace variables set; the replay-index test), then the scrub removes the `GIT_TRACE` prefix and `run_property` refuses an index beyond the cases; the comments that named the property SPEC by its card number name its id.

## Decisions

| Id | Decision | Source |
|---|---|---|
| D-1 | The whole runtime state folder is tool state; the list of one file name is widened to the folder. Only untracked files count as cache; a tracked file there that changed is still work. | the leader's note on the reclaim defect ("leave the runtime state folder out of the remaining-work judgement, or put it in .gitignore") |
| D-2 | The test clears the variable itself (the second option the leader offered); the expected entry does not change. | the leader's note on the session id defect |
| D-3 | An ignored file under the runtime state folder is not work either; the ignored-entry rule takes the same folder as the untracked rule. | the leader's decision in the audit of slice 1 (N2): "yes, an ignored state-folder file is not work either" |
| D-4 | The earlier-layout test builds that layout through the resolver constant for the earlier project folder, not by spelling it. | leader rule: tests never spell the state folder |
| D-5 | The 30 s budget of SPEC-20261119-property-tests AC-3 is kept for an idle machine and multiplied by load / CPUs (at least 1, at most 8) instead of dropping the wall-time check; a failure still prints the seed. | the leader's note on the flaky budget offered "count cases, not wall time; or scale the budget"; scaling keeps the other SPEC's acceptance on an idle machine, which dropping it would remove |
| D-6 | The scrub removes by the prefix `GIT_TRACE`, which covers `GIT_TRACE`, `GIT_TRACE_*`, `GIT_TRACE2` and `GIT_TRACE2_*`; moai's own git children never trace. | the audit of slice 3 (N2: with one variable set, every reclaim of an existing worktree is refused) |
| D-7 | A replay index beyond the cases fails the run instead of running zero cases. | the audit of slice 3 (N7) |

### Research

| Id | Fact | Read |
|---|---|---|
| R-1 | With the real session id in the shell, `register_mode_is_stored_and_renewed` fails: the entry gets a `session` key beside `schema_version`, `name`, `instance`, `lease_until`, `registered_at`. With the variable unset or a short test value the target passes (22 passed). | run in the defects worktree, 2026-11-22 |
| R-2 | `is_tool_cache_entry` in `src/card/early_reclaim.rs` accepts only `state/github/counts.json` of the current state folder and `.moai/state/github/counts.json` of the earlier one; the test `a_listed_tool_cache_is_not_work` pins that a second name in the same folder is work. | read in the tree at the integration tip |
| R-3 | The git manual documents `GIT_TRACE`, `GIT_TRACE_FSMONITOR`, `GIT_TRACE_PACK_ACCESS`, `GIT_TRACE_PACKET`, `GIT_TRACE_PACKFILE`, `GIT_TRACE_PERFORMANCE`, `GIT_TRACE_REFS`, `GIT_TRACE_SETUP`, `GIT_TRACE_SHALLOW`, `GIT_TRACE_CURL`, `GIT_TRACE2`, `GIT_TRACE2_EVENT`, `GIT_TRACE2_PERF` and `GIT_TRACE_REDACT`; every one starts with `GIT_TRACE`. | `man git` of git 2.54.0 (Apple Git-157), read 2026-11-23 |

Slice 4 leaves the follow-ups that write `src/card/dispatch.rs` (carry git's first standard-error line in the refusal and decide which lines are noise; a worktree that lost its `.git` inside a checkout) for the card that reserves that file to release it.

## Risks

- A person who keeps an untracked or an ignored file of their own under the state folder loses the refusal for it; the folder is tool state by definition and the file is untracked, so a changed tracked file stays work (REQ-4).
- The entry rule lives in a file another card reserves for a flow rename; the edit is one rule and its comment and the other card's hunks are not touched.

## Result

Slice 1 (revision 1):

- AC-1: met. `session_registry_children_do_not_inherit_the_shell_session_id` was red first (the command left the variable to the shell) and passes; the whole `session_registry_cli` target passes with the real session id set in the shell (23 passed); 2 mutants killed (the removal dropped, another variable name removed).
- AC-2: met. `a_runtime_state_folder_is_not_work` was red first (holds 3 uncommitted change(s)) and passes: the counts file, a file beside it, a file two folders below and the earlier generation's state folder are reclaimed and every file stays.
- AC-3: met. `a_listed_tool_cache_is_not_work` keeps its refusals for an untracked file outside the state folder and a changed tracked file, one uncommitted change each; the unit test `only_an_untracked_state_folder_file_is_a_cache` was rewritten red first; 3 mutants killed (no trailing slash, the earlier folder dropped, the untracked check dropped).
- The rule lives in `src/card/early_reclaim.rs`, a file another card reserves; its holder answered that the hunk does not overlap theirs, and the edit is by explicit path.

Slice 2 (revision 2, the audit follow-ups):

- AC-4: met. `an_old_layout_runtime_state_folder_is_not_work` (unit test in `src/card/early_reclaim.rs`) builds the earlier project layout through `crate::PROJECT_STATE_DIR`; the resolver picks that layout's state folder, which differs from the earlier generation's, and the untracked rule and the ignored-entry rule each accept both folders on their own. Two mutants killed (the current-folder arm of the ignored-entry rule dropped, the earlier-generation arm dropped).
- AC-5: met. `every_child_is_built_by_the_one_helper` passes: the only line of `tests/session_registry_cli.rs` that builds a command is the one in `registrar`.
- AC-6: met. `an_ignored_state_folder_file_is_not_work` was red first (a2a5a0407d: the refusal named the ignored `state/github/counts.json` as a draft) and passes (11b4621f3a): with the state folder ignored, an ignored file there is reclaimed and stays, and an ignored file outside it still refuses and is the one named.
- The two comments that named one listed cache file now name the folder rule (`tests/card_reclaim_early_cli.rs`, `src/card/dispatch.rs`).

Slice 3 (revision 3):

- AC-7: met. `budget` and `over_budget` in `tests/common/prop.rs`: 30 s at or below one runnable thread per CPU, with no reading or with no CPU count; 30 s times load / CPUs above it, at most 240 s; the over-budget message names the property, the cases, the seed, the elapsed time, the budget and the load on the CPU count. Red at 587365b5ac (both budget tests failed on the 30 s stub), green at 2e904cfe01 (`prop_reservation`, `prop_headver`, `prop_scope_paths` passed).
- AC-8: met. A mode-000 untracked folder let the reclaim through at 587365b5ac (`ok:true`, generation 2); at 2e904cfe01 the gate refuses it as unreadable and the ledger is unchanged. Removing the warning check makes the test fail again (mutant killed).
- AC-9: met. `my notes.md` outside the state folder is one uncommitted change, and `my counts.json` under it is tool state: red at 587365b5ac (the quoted name counted as work), green at 2e904cfe01. Reading without `-z` makes the test fail again (mutant killed); `a_z_list_keeps_names_and_folds_a_rename_origin` pins the rename origin field.
- The state folder is resolved once per check (`state_prefix`), and the note numbers left this SPEC.
- Not claimed: the Python twin's 30 s budget in `tests/prop_manifest_test.py` (outside the card's write scope) and the text pin of `every_child_is_built_by_the_one_helper`, which stays as written.

Slice 4 (revision 4):

- AC-10: met. `git_trace_variables_are_removed_by_prefix` (unit test, `src/identity.rs`) and `a_git_trace_variable_does_not_make_a_worktree_unreadable` were red first (the reclaim of a dirty worktree with `GIT_TRACE`, `GIT_TRACE2` and `GIT_TRACE_SETUP` set answered "could not be read" instead of naming its one change) and pass; removing the `GIT_TRACE` prefix from the scrub makes both fail again (mutant killed).
- AC-11: met. `a_replay_index_beyond_the_cases_is_refused` and `the_run_refuses_a_replay_index_beyond_the_cases` (`tests/common/prop.rs`): an index at or beyond the cases fails with the property, the index, the cases and the seed, and a replay inside the run runs exactly one case; red first, and dropping the check from the run makes the second fail again (mutant killed).
- The comments of `tests/common/prop.rs` and `tests/prop_reservation.rs` that named the property SPEC by a card number now name its id; one earlier commit body still carries a card number and stays as pushed.
- `spec check`: 0 errors, 1 warning (more than 12 requirements).
- Not claimed, waiting for `src/card/dispatch.rs` to be free: the standard-error line carried in the refusal and the noise lines of git (a missing fsmonitor hook, an invalid attribute line), and a recorded worktree that lost its `.git` inside a checkout. Not claimed, not run: the stderr rule pinned on each of its three reads; the probe of which situation makes the ignored-files read and the file-list read warn on their own was refused by the host and not repeated.
