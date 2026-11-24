---
id: SPEC-20261122-test-defects-session-reclaim
title: "Two defects: the registry tests read the shell's session id, and card reclaim counts the runtime state folder as work"
status: draft
revision: 7
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
- REQ-15: When moai starts a git child, the child shall run with the trace2 targets `GIT_TRACE2`, `GIT_TRACE2_EVENT` and `GIT_TRACE2_PERF` set to `0`.
- REQ-16 (REQ-15): IF the git configuration of the user sets a trace2 target, THEN the child shall still write no trace.
- REQ-17 (REQ-13): IF `PROP_ONLY` is set and is not a number, THEN the Rust property harness shall fail and name the value.
- REQ-18: When the file-list read or the checkout-top read of a recorded worktree fails or writes a line on its standard error that is not a noise line, the refusal shall carry the first such line git wrote.
- REQ-19 (REQ-18): IF the only lines git writes on its standard error say that a name in an attributes file is not a valid attribute name, THEN the read shall count as complete with its list as the answer.
- REQ-20: When moai starts a git read of a recorded worktree, the child shall run with `core.fsmonitor` switched off.
- REQ-21 (REQ-20): IF the configuration of the worktree names an fsmonitor hook that does not exist, THEN the read shall list the worktree without a line on its standard error.
- REQ-22: When a recorded worktree exists, the gate shall judge it only when git places the top of a checkout at that folder.
- REQ-23 (REQ-22): IF the folder lost its own `.git` and lies inside another checkout, THEN the gate shall refuse it as unreadable and name where git places it.
- REQ-24 (REQ-9): IF the ignored-entry read or the ignored-file list read of the early reclaim writes a warning, THEN the early reclaim shall refuse as unreadable, each read on its own.

## Acceptance

- AC-1 (REQ-1, REQ-2): every child process that `tests/session_registry_cli.rs` starts has the variable removed from its environment, and the whole target passes with the variable set in the shell · verify: `cargo test --offline --locked --test session_registry_cli session_registry_children_`
- AC-2 (REQ-3): a worktree whose only untracked files are under the runtime state folder (the counts file, another file beside it, a file in a folder below) is reclaimed and every file stays · verify: `cargo test --offline --locked --test card_reclaim_early_cli a_runtime_state_folder_is_not_work`
- AC-3 (REQ-4): an untracked file outside the state folder and a changed tracked file each keep the refusal and name one uncommitted change, with a state folder file beside them · verify: `cargo test --offline --locked --test card_reclaim_early_cli a_listed_tool_cache_is_not_work`
- AC-4 (REQ-3): in a worktree on the earlier project layout, an untracked file under that layout's state folder is a cache and a file beside it is work, so the current-folder rule and the earlier-generation rule each decide on their own · verify: `cargo test --offline --locked --lib an_old_layout_runtime_state_folder_is_not_work`
- AC-5 (REQ-2): `tests/session_registry_cli.rs` builds every command for the binary under test in its one helper, and a test fails when the file builds one anywhere else · verify: `cargo test --offline --locked --test session_registry_cli every_child_`
- AC-6 (REQ-5, REQ-6): a worktree that ignores its runtime state folder and holds an ignored file there is reclaimed early and the file stays, while an ignored file outside that folder still refuses the reclaim · verify: `cargo test --offline --locked --test card_reclaim_early_cli an_ignored_state_folder_file_is_not_work`
- AC-7 (REQ-7, REQ-8): the budget of `tests/common/prop.rs` is 30 s when the load is at or below the CPU count or cannot be read, grows in proportion above it, and is capped; the over-budget message carries the seed, the elapsed time, the budget and the load · verify: `cargo test --offline --locked --test prop_reservation the_property_budget_`
- AC-8 (REQ-9, REQ-10): a worktree with an untracked folder that cannot be opened (mode 000) is refused as unreadable, where git exits 0 and only warns · verify: `cargo test --offline --locked --test card_reclaim_early_cli an_unreadable_folder_is_not_a_clean_worktree`
- AC-9 (REQ-3, REQ-4): an untracked file whose name holds a space is compared as written: under the runtime state folder it is tool state, outside it is one uncommitted change · verify: `cargo test --offline --locked --test card_reclaim_early_cli a_name_with_a_space_is_read_as_written`
- AC-10 (REQ-11, REQ-12): the name test removes every documented trace variable and keeps an unrelated one, and a `card reclaim` run with `GIT_TRACE`, `GIT_TRACE2` and `GIT_TRACE_SETUP` set refuses a dirty worktree for its one change and reclaims a clean one · verify: `cargo nextest run --lib -E 'test(git_trace_variables_are_removed_by_prefix)'` and `cargo nextest run --test card_reclaim_early_cli -E 'test(a_git_trace_variable_does_not_make_a_worktree_unreadable)'`
- AC-11 (REQ-13, REQ-14): an index at or beyond the case count is refused with the property, the index, the case count and the seed; an index inside the run or none is accepted · verify: `cargo nextest run --test prop_reservation -E 'test(a_replay_index_beyond_the_cases_is_refused)'`
- AC-12 (REQ-15, REQ-16): the scrub puts the three trace2 targets into the child's environment as `0`, and a `card reclaim` of a clean worktree whose user-level git configuration sets `trace2.normalTarget = 2` is reclaimed · verify: `cargo nextest run --lib -E 'test(the_trace2_targets_are_switched_off_for_the_child)'` and `cargo nextest run --test card_reclaim_early_cli -E 'test(a_trace2_config_does_not_make_a_worktree_unreadable)'`
- AC-13 (REQ-17): a `PROP_ONLY` that is not a number is refused and the message names the value; a number or no value is accepted · verify: `cargo nextest run --test prop_reservation -E 'test(a_replay_index_that_is_not_a_number_is_refused)'`
- AC-14 (REQ-18): the refusal of a worktree with an untracked folder that cannot be opened carries git's own line (`could not open directory`), and the first line that is not a noise line is the one carried · verify: `cargo nextest run --test card_reclaim_early_cli -E 'test(an_unreadable_folder_is_not_a_clean_worktree)'` and `cargo nextest run --lib -E 'test(the_refusal_names_the_first_line_git_wrote)'`
- AC-15 (REQ-19): a clean worktree whose attributes file holds an invalid attribute name is reclaimed, and a line that only looks like the noise line (a line number that is empty included) still refuses · verify: `cargo nextest run --test card_reclaim_early_cli -E 'test(an_invalid_attribute_line_does_not_make_a_worktree_unreadable)'` and `cargo nextest run --lib -E 'test(a_noise_line_is_not_a_warning)'`
- AC-16 (REQ-20, REQ-21): a clean worktree whose configuration names a `core.fsmonitor` hook that does not exist is reclaimed · verify: `cargo nextest run --test card_reclaim_early_cli -E 'test(a_missing_fsmonitor_hook_does_not_make_a_worktree_unreadable)'`
- AC-17 (REQ-22, REQ-23): a folder that lost its `.git` inside a checkout that ignores it is refused as unreadable and the refusal names the checkout git finds, where the checkout around it alone would call it clean, and a checkout of its own, a linked worktree beside the checkout or inside it, one whose folder name ends in a space, and a link to a checkout are each judged by themselves · verify: `cargo nextest run --lib -E 'test(a_folder_that_lost_its_git_is_not_judged_by_the_checkout_around_it) | test(a_linked_worktree_and_a_link_to_a_checkout_are_judged_by_themselves) | test(a_worktree_whose_name_ends_in_a_space_is_judged_by_itself)'`
- AC-18 (REQ-9, REQ-24): an ignored folder git cannot open refuses the early reclaim as unreadable through the ignored-entry read, and a warning of the ignored-file list read refuses it through that read · verify: `cargo nextest run --test card_reclaim_early_cli -E 'test(an_unreadable_ignored_folder_is_not_a_clean_worktree) | test(a_warning_of_the_ignored_file_list_refuses)'`
- AC-19 (REQ-15, REQ-16): a clean worktree whose user-level git configuration sets `trace2.eventTarget = 2`, and one that sets `trace2.perfTarget = 2`, is reclaimed · verify: `cargo nextest run --test card_reclaim_early_cli -E 'test(a_trace2_event_target_config_does_not_make_a_worktree_unreadable) | test(a_trace2_perf_target_config_does_not_make_a_worktree_unreadable)'`

## Scope

- write: `tests/session_registry_cli.rs`, `tests/card_reclaim_early_cli.rs`, `src/card/early_reclaim.rs` (the entry rules, their comments and unit tests), `src/card/dispatch.rs` (the worktree gate and its git reads), `src/identity.rs` (the git environment scrub), `tests/common/prop.rs` and `tests/prop_reservation.rs` (the time budget, the replay index and the comments naming the property SPEC), this SPEC folder

## Plan

1. REQ-1 and REQ-2: a red test that reads the environment of the command the tests build (`Command::get_envs`), then the variable is removed in the one helper that builds it.
2. REQ-3 and REQ-4: the red test first, then the entry rule of `is_tool_cache_entry` names the state folder instead of one file; the older test that pinned "another name in the same folder is work" is rewritten to the new rule and keeps its other three cases.
3. The audit follow-ups (revision 2): the earlier-layout unit test (AC-4) and the one-helper test (AC-5) first, then REQ-5 red first in `card_reclaim_early_cli` and the ignored-entry rule of `is_regenerable_ignored` widened to the runtime state folder; the two comments that still name one listed cache file are brought up to the folder rule.
4. Slice 3 (revision 3, the leader's note on the flaky property budget and the audit of slice 2): the budget function and its unit tests in `tests/common/prop.rs` first (red), then `run_property` checks the elapsed time against it and names the load in its message.
5. Slice 3, the audit of slice 2: AC-8 and AC-9 red first, then the gate's git reads go through one helper that refuses on a warning and read paths NUL-separated (`-z`), and the state folder is resolved once per check instead of once per entry.
6. Slice 4 (revision 4, the audit of slice 3), the part that needs no write in `src/card/dispatch.rs`: AC-10 and AC-11 red first (the name test and the CLI test with the trace variables set; the replay-index test), then the scrub removes the `GIT_TRACE` prefix and `run_property` refuses an index beyond the cases; the comments that named the property SPEC by its card number name its id.
7. Slice 5 (revision 5, the audit of slice 4): the line break lost before AC-10 is restored; the doc comment of `scrub_git_env` names the trace prefix; AC-12 and AC-13 red first, then the scrub sets the three trace2 targets to `0` and `PROP_ONLY` is parsed by one function that refuses a value that is not a number.

8. Slice 6 (revision 6, the leader's slice 5 of the card: the follow-ups that wait for `src/card/dispatch.rs`, now free): the red tests first (the refusal's git line, the invalid attribute line, the missing fsmonitor hook, the folder that lost its `.git`, the ignored-entry and ignored-file list reads, the event and perf trace2 targets), then the git helpers return the reason with the failure, a noise line is skipped, `core.fsmonitor` is switched off on every read, and the gate checks that git places the checkout top at the recorded folder.
9. Slice 7 (revision 7, the audit of slice 6): the red tests first (a linked worktree whose folder name ends in a space, and the noise look-alike with an empty line number), then the checkout top is cut only at its line break and the dead name guard of the noise rule goes; REQ-18 is narrowed to the two reads that carry git's line (D-14).

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
| D-8 | The three trace2 targets are set to `0`, not only removed, because the user's git configuration (`trace2.normalTarget`) turns the trace on when the variable is absent; the other trace variables stay removed. | the audit of slice 4 (N1) |
| D-9 | A `PROP_ONLY` that is not a number fails the run; `PROP_SEED` keeps its reading. | the audit of slice 4 (N5) |
| D-10 | A standard-error line is noise only when git says it carried on and the list is complete; today that is one shape, `<name> is not a valid attribute name: <file>:<line>`. Every other line, an unknown one included, keeps the refusal. | the audit of slice 3 (N1: decide which lines are noise); R-4 |
| D-11 | The missing fsmonitor hook is prevented, not filtered: every read runs with `-c core.fsmonitor=false`. A read of a worktree gains nothing from a watcher, and git would otherwise exec the hook or start the built-in daemon. | R-4 |
| D-12 | The stderr rule is pinned on the ignored-entry read with a real folder git cannot open, and on the ignored-file list read with a stand-in git named by `MOAI_GIT_BIN`, because no real situation made the list read warn while the entry read stayed silent. | R-4 |
| D-13 | A recorded folder that is not the top of a checkout of its own is unreadable, never clean: a leftover folder of a removed worktree keeps refusing until someone removes it. | the audit of slice 3 (N6) |
| D-14 | REQ-18 is narrowed to the two reads that carry git's line, the file-list read and the checkout-top read. The reads that count unpushed commits, find the integration branch or read the stash (`git_output`) keep refusing as unreadable without a reason from git, which is the safe direction and the same refusal as before. The other choice, making `git_output` carry the reason, is not taken: the change is larger and the refusal does not change. | the audit of slice 6 (N1), the leader's decision on that audit (delegation of the user, 2026-11-23T13:57Z), reversible, category none |

### Research

| Id | Fact | Read |
|---|---|---|
| R-1 | With the real session id in the shell, `register_mode_is_stored_and_renewed` fails: the entry gets a `session` key beside `schema_version`, `name`, `instance`, `lease_until`, `registered_at`. With the variable unset or a short test value the target passes (22 passed). | run in the defects worktree, 2026-11-22 |
| R-2 | `is_tool_cache_entry` in `src/card/early_reclaim.rs` accepts only `state/github/counts.json` of the current state folder and `.moai/state/github/counts.json` of the earlier one; the test `a_listed_tool_cache_is_not_work` pins that a second name in the same folder is work. | read in the tree at the integration tip |
| R-3 | The git manual documents `GIT_TRACE`, `GIT_TRACE_FSMONITOR`, `GIT_TRACE_PACK_ACCESS`, `GIT_TRACE_PACKET`, `GIT_TRACE_PACKFILE`, `GIT_TRACE_PERFORMANCE`, `GIT_TRACE_REFS`, `GIT_TRACE_SETUP`, `GIT_TRACE_SHALLOW`, `GIT_TRACE_CURL`, `GIT_TRACE2`, `GIT_TRACE2_EVENT`, `GIT_TRACE2_PERF` and `GIT_TRACE_REDACT`; every one starts with `GIT_TRACE`. | `man git` of git 2.54.0 (Apple Git-157), read 2026-11-23 |
| R-4 | git 2.54.0, one repository per case, the three reads of the gate and the early reclaim. An untracked folder of mode 000: the status read warns (`warning: could not open directory 'locked/': Permission denied`), exit 0, nothing listed for it. An ignored folder of mode 000: the plain status read is silent, the `--ignored=traditional` read and the ignored-file list read both warn. A mode-000 folder below an ignored or untracked folder: every read that descends warns and lists the rest. A missing `core.fsmonitor` hook: two `fatal: cannot exec '<hook>': No such file or directory` lines on every read, exit 0, the lists complete. An attribute file with `we!rd` in it and a file whose content git compares: one line `we!rd is not a valid attribute name: .gitattributes:1` on the two status reads, exit 0, the lists complete. | probe test run in this worktree, 2026-11-23 |

Slices 4 and 5 left the follow-ups that write `src/card/dispatch.rs` for the card that reserves that file to release it; slice 6 is those follow-ups.

## Risks

- A person who keeps an untracked or an ignored file of their own under the state folder loses the refusal for it; the folder is tool state by definition and the file is untracked, so a changed tracked file stays work (REQ-4).
- The entry rule lives in a file another card reserves for a flow rename; the edit is one rule and its comment and the other card's hunks are not touched.
- A folder that is left behind by a removed worktree and is no checkout of its own now refuses a reclaim until it is removed; before, the checkout around it decided (REQ-22, D-13). The refusal names where git places the folder, so the person knows what to remove.
- A noise line is a guess about git's wording; another git version may word it differently, and then the read refuses (the safe direction) and the refusal carries that line (REQ-18).
- `-c core.fsmonitor=false` (D-11) was confirmed on git 2.54.0 only (R-4). The audit of slice 6 named git before 2.36 as taking another path for a missing hook (not measured here); a line git still writes then keeps the refusal, the safe direction.

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

Slice 5 (revision 5, the repair of the slice 4 audit):

- Line break: AC-10 stood at the end of the AC-9 line, so `moai spec show` listed no AC-10; it is its own line again and the show lists AC-10, AC-11, AC-12 and AC-13.
- AC-12: met. With `trace2.normalTarget = 2` in the user-level git configuration of the temporary home, the reclaim of a clean worktree answered "could not be read" at the red commit and passes now; `the_trace2_targets_are_switched_off_for_the_child` pins the three targets as `0`. Removing the loop that sets them makes both tests fail again (mutant killed). The configuration of the worktree itself (`git config` in the repository) does not turn the trace on in git 2.54.0, so the test uses the user-level file.
- AC-13: met. `a_replay_index_that_is_not_a_number_is_refused`: a `PROP_ONLY` that is empty, a word, negative, fractional or followed by a letter fails and names the value; a number or no value passes. Reading the value leniently again makes the test fail (mutant killed).
- The doc comment of `scrub_git_env` names the trace prefix, the trace2 targets and the numbered config pairs.
- Not claimed, unchanged from slice 4: the follow-ups that write `src/card/dispatch.rs`, the stderr rule pinned on each of its three reads, the text pin and the Python twin. The checker `spec check` does not catch two list items on one line; that is for a checker card.

Slice 6 (revision 6, the follow-ups of the audits of slices 3 and 4 that write `src/card/dispatch.rs`):

- AC-14: met. The git helpers return the reason with a failure (how git ended and its first line) and `git_entries` names the first line that is no noise line; the refusal of a mode-000 untracked folder now ends with `git said: warning: could not open directory 'locked/': Permission denied`. Red at bf70f93465 (the refusal named no line, the first line of a list led by a noise line was the noise line), green at 79a87aeaf2. Mutant: the line cut to nothing makes three CLI tests fail.
- AC-15: met. One noise shape (`<name> is not a valid attribute name: <file>:<line>`, D-10): `a_noise_line_is_not_a_warning` keeps five look-alikes as warnings; a clean worktree whose attributes file holds `we!rd` is reclaimed (red: "could not be read"). Mutant: the shape never matches, both tests fail.
- AC-16: met. Every read runs with `-c core.fsmonitor=false`; a clean worktree whose configuration names a missing hook is reclaimed (red: "could not be read"). Mutant: the argument removed, the test fails.
- AC-17: met. `a_folder_that_lost_its_git_is_not_judged_by_the_checkout_around_it`: a folder inside a checkout that ignores it, with the checkout's commits all pushed, was `Ok` at bf70f93465 and is refused as "not the top of a git checkout of its own" naming the checkout at 79a87aeaf2; a clone of its own is still `Ok`. `a_linked_worktree_and_a_link_to_a_checkout_are_judged_by_themselves` pins what a lane really records (a `git worktree add` worktree beside the checkout and inside it, a link to a checkout), green at birth. Mutants: the check removed kills the first test; comparing the two paths as written, without resolving them, kills the second.
- AC-18: met. The ignored-entry read with a real mode-000 ignored folder (the plain status read stays silent there, R-4) and the ignored-file list read with a stand-in git (`MOAI_GIT_BIN`, D-12) each refuse as unreadable and carry the line. Mutants: ignoring warnings on the `ls-files` read kills only the second test, on the `--ignored=traditional` read only the first.
- AC-19: met, pinned at birth (green at bf70f93465, behaviour already in the scrub). `trace2.eventTarget = 2` and `trace2.perfTarget = 2` in the user-level configuration; removing `GIT_TRACE2_EVENT`, or `GIT_TRACE2_PERF`, from the child's environment after the scrub (a mutant in `dispatch.rs`, the scrub itself is outside the card's scope) kills the matching test only.
- Real runs of the dangerous path (a read that starts git): the CLI tests run the real binary on a mode-000 untracked folder, a mode-000 ignored folder, a missing fsmonitor hook, an attributes file with an invalid name and a stand-in git; the unit test runs the real git on a folder that lost its `.git`.
- Not claimed, unchanged: the text pin of `every_child_is_built_by_the_one_helper`, the Python twin, `GIT_TRACEBACK` removed by the prefix (harmless), the card numbers of `tests/prop_headver.rs` and `tests/prop_scope_paths.rs` (outside the scope), and the checker `spec check` that misses two list items on one line (a checker card). A noise line other than the one shape keeps the refusal.

Slice 7 (revision 7, the repair of the audit of slice 6):

- AC-15: met. The look-alike `we!rd is not a valid attribute name: .gitattributes:` (the line number empty) joins `a_noise_line_is_not_a_warning`; it is green at birth because the guard stood already. Mutant: the empty-line-number guard removed makes the test fail (exit 100).
- AC-17: met. `a_worktree_whose_name_ends_in_a_space_is_judged_by_itself`: a real linked worktree whose folder name ends in a space was refused as "not the top of a git checkout of its own" at d4270312e0 (red, exit 100) because the checkout top was trimmed of every white space; `checkout_top_is` now cuts only the closing line break and the worktree is `Ok` at 325dbb1e35.
- The dead name guard of `is_noise_line` is gone: the line is trimmed before it is read, so the space that opens the separator already means a name stands before it.
- REQ-18 is narrowed to the file-list read and the checkout-top read (D-14), the git version of the fsmonitor switch is a risk (Risks).
