# Mutation runner

`tools/mutate.py` measures explicitly specified one-line mutations against narrow Cargo test targets. It emits raw evidence, not an audit verdict. A completed run, a `FAILED` target, or exit 0 does **not** mean audit PASS. The auditor decides whether the mutation and assertion establish the intended behavior.

## Prepare and invoke

Commit the intended source, tests, and runner before mutation. The runner requires a committed HEAD and a clean source Git index and working tree, including no untracked files. Before Cargo runs, the initial snapshot must compare every tracked entry's physical bytes with the raw HEAD blob, validate stage-0 index entries and Git modes, and reject assume-unchanged or skip-worktree flags even when the physical bytes are clean. Git status or a clean diff alone is insufficient. A nonexecutable file with filesystem permissions `0600` is compatible with Git mode `100644`. The runner must preserve the original raw index rather than clearing flags or refreshing it to hide differences. It does not commit for you, copy dirty changes, or repair the source tree. Keep the specification and output in the source worktree's ignored `target/` directory; an untracked specification elsewhere would fail the clean-tree guard.

Run from the **source Git root**, with Python 3.9+, Git, Cargo, and the trusted project's offline dependencies available. For example, save this JSON as `target/mutations.json`:

```json
{
  "schema_version": 1,
  "timeout_seconds": 300,
  "mutations": [
    {
      "id": "git-env-removal",
      "file": "src/identity.rs",
      "old": "        cmd.env_remove(key);",
      "new": "        let _ = key;",
      "targets": [{"kind": "test", "name": "product_cli"}]
    }
  ]
}
```

```sh
CARGO_TARGET_DIR="$PWD/target/mutation-cargo" \
  python3 tools/mutate.py --spec target/mutations.json \
  > target/mutation-result.json
```

The example describes a deliberate removal of Git-environment scrubbing; it does not claim that this target kills that mutation. Choose the target and optional filter that exercise your assertion, and inspect the recorded output. The exact `old` line must still exist uniquely in the committed file at the HEAD you run.

`CARGO_TARGET_DIR` is mandatory for this tool. Unlike ordinary developer build-directory examples in `AGENTS.local.md`, the runner requires a resolved, strict descendant of **this source worktree's `target/`**. The `target/` root itself, external paths, another worktree's target, sibling-prefix matches, and escaping symlinks are rejected. Do not share this root with another run or session. The runner selects a unique build-directory child beneath it, never the shared root itself.

The no-sharing rule applies to `CARGO_TARGET_DIR`, not dependency caches. Explicit `CARGO_HOME`, `RUSTUP_HOME`, `RUSTC` and other toolchain/cache settings are respected; a deliberately prepared dependency cache such as this repository's `/tmp/moai-rc-product-cargo` may be shared. Unset settings use scratch defaults. With actual Rustup proxies, explicitly provide a prepared project/toolchain cache when the scratch defaults do not contain the required offline toolchain and dependencies. Missing preparation produces a Cargo error, not evidence of a killed mutation.

## Specification

The root object has `schema_version: 1`, a nonempty `mutations` array, and optional positive finite `timeout_seconds` (default 300). Each mutation supplies:

- `id`: the mutation identifier.
- `file`: a relative, tracked regular-file path. Absolute paths, traversal and symlinks are rejected.
- `old`, `new`: nonempty, different, complete single lines, including indentation but not line terminators. `old` must occur exactly once across the file, not merely be the first matching substring. The matching source line must be actual code: after indentation, `//`, `///`, `//!`, and `#` starts are rejected. This conservative guard also excludes Rust attribute lines beginning with `#`. This restriction applies to `old`, not `new`: a replacement comment is allowed when deliberately removing a guard.
- `targets`: a nonempty array of single-target selectors. `kind` is `lib`, `bin`, `test`, `example`, or `bench`. `lib` has no `name`; the other kinds require a simple literal target name, without globs, paths or option injection. An optional nonempty `filter` must not begin with `-`.

Each target causes exactly one argv-based invocation in the clone:

```text
cargo test --offline --locked --message-format=json --lib [filter]
cargo test --offline --locked --message-format=json --<kind> <name> [filter]
```

There is no shell-command field, arbitrary Cargo argument list, automatic baseline, retry, or generated mutation. Supply separate target entries rather than combining selectors in one call. In particular, a name filter must not silently select zero tests in a second target. Cargo output and exit status must both be inspected.

Cargo's JSON build protocol is the authority for separating compilation from test execution. Classification requires a successful `build-finished` record, no unsuccessful `build-finished` record, and exactly one selected test `compiler-artifact` with `profile.test: true` and an executable. Named selectors must match the artifact's target kind and name; a library artifact may use the `lib` or `proc-macro` kind. Missing, unsuccessful or ambiguous build/artifact evidence is an `error`, not a killed mutation. There is no legacy plain-text compiler-output fallback.

Only a libtest summary in stdout **after the final successful `build-finished` record** may establish runtime counts. Stderr, pre-build stdout, JSON diagnostic strings, multiline/rendered compiler diagnostics and build-script output are not runtime-summary evidence, even if they contain an exact `test result: FAILED.` string. Retain both raw streams; do not flatten diagnostic JSON into candidate runtime text. With valid build/artifact evidence but no stdout runtime summary, exit 0 is `not_run` and nonzero exit is `error`; a stderr-only summary with exit 101 is therefore `error`, never `FAILED`. One Cargo invocation still handles both build and runtime; no second invocation is used to infer success.

## Isolation and restoration evidence

The source worktree is read-only with respect to tracked files, raw index bytes and refs; the runner does create unique ignored artifacts beneath its `target/`. It removes inherited Git redirection variables and checks the actual source Git root rather than accepting an inherited alternate index or repository. It records the source HEAD and verifies source HEAD, raw index, stage entries, semantic flags, tracked contents and cleanliness again at the end. Initial source, final source (including rejection/error paths), and clone snapshots must share the same no-follow tracked-entry reader: validate every parent component from a trusted directory handle before inspecting or reading the leaf. Neither leaf-only `lstat` nor a subsequent ordinary path read protects against a symlinked parent. Leaf symlinks are recorded by their literal link text without dereferencing; mutation input must still be a regular file. Native Windows reparse parents require equivalent checks. A rejected external-parent path must cause zero external file opens, not merely leave the external bytes unchanged. The runner never runs restore or reset against the source.

For each run it retains an independent local clone beneath the source `target/`, cloned with `--no-local --no-hardlinks --no-checkout` and checked out detached at the exact committed HEAD. This is not a linked Git worktree; it has no shared hardlinks or alternates and does not write source refs or index. Dirty and untracked source files are not copied. Git commands and target execution use synthetic HOME, XDG and temporary directories by default, while respecting explicitly supplied toolchain/cache settings as described above.

Only the clone receives the replacement. The runner records the actual `git diff -U0` and checks that it contains one removed code line and one added replacement line, which may be a comment. On every outcome, `finally` handles restoration from saved HEAD bytes after target processes are stopped. Restoration checks HEAD, index stage entries and semantic flags (including assume-unchanged and skip-worktree), tracked file bytes, and `git diff --quiet HEAD`. Clone index stat-cache refreshes are distinct from semantic index changes and from the source's raw-index preservation requirement. A clean diff alone must not hide changed bytes or index flags; a successful test command alone is not restoration proof.

If a target replaces the mutation file with a symlink, FIFO, directory or other nonregular object, restoration must not follow it into another location. Unsafe or failed restoration is reported as false, exits unsuccessfully and stops subsequent mutations. The original source is never reset to hide a mismatch.

On timeout or interruption, the runner stops child processes before restoration, using a POSIX process group or a native Windows Job Object. On Windows the process is created suspended, assigned to the job immediately, then resumed; the job lifetime is independent of its parent process. If the OS refuses termination, the runner does not bypass that refusal or claim successful restoration: it retains the artifacts with `quiescent: false`, `restored: false`, global `blocked`, and no subsequent mutations. The parent observed POSIX `killpg` returning `EPERM` on the current host and exercises this fail-closed boundary rather than presenting it as successful restoration. Windows runtime execution is `not_run`.

Deadline regressions use private Python dependency-injection seams, not CLI options, specification fields or environment-variable backdoors: `run_target(argv, workspace, env, timeout, *, clock=None, started=None, spawn=None, stop=None)`. Production defaults resolve to the monotonic clock, process creation and normal child-tree quiescence inside the function. After owned process/job setup, calculate `deadline = clock() + timeout`, invoke the optional test-only `started(process)` handshake, then calculate `remaining = deadline - clock()`. An expired or exact deadline (`remaining <= 0`) must set `timed_out: true` before any `communicate` call; otherwise pass only the remaining time to `communicate`. A `TimeoutExpired` also sets that field. Do not restart a full timeout after the handshake.

Tests should coordinate a real helper process through a READY/IPC handshake and a private controlled clock, including delayed startup, rather than assume a child starts or creates a log within 0.2 seconds. Controlled cleanup can request the helper's ordinary IPC exit; it must not bypass host termination refusal. Refused cleanup remains fail-closed. These seams establish deterministic boundary observations without changing production timeout or process-tree policy.

This is **runner-write isolation for trusted projects, tests and toolchains**, not an OS sandbox. The trusted operator and project are responsible for the safety of explicitly selected toolchain/cache paths. Cargo build scripts and test executables are executable code; synthetic HOME and a clone do not confine malicious code or guarantee that it cannot access the host or network. The runner itself does not fetch, commit, push, post verdicts or issue network requests; `--offline` constrains Cargo dependency access, not arbitrary project code.

## JSON results and exit status

Stdout contains exactly one JSON object, including on reported errors. Retain it as evidence. The schema-version-1 result contains:

| Field | Meaning |
| --- | --- |
| `schema_version` | `1` |
| `head` | Source commit used for the clone |
| `workspace` | Retained clone workspace |
| `cargo_target_dir` | Actual unique build directory used by this run |
| `source_unchanged` | End-of-run source-preservation result |
| `mutations` | Per-mutation evidence |
| `status` | `completed`, `blocked`, or `error` |
| `errors` | Reported problems |

Each mutation records `id`, `file`, actual `diff`, `targets`, `restored`, and `restore_diff_exit`. Each target records `argv`, `exit`, complete `stdout` and `stderr`, parsed `passed` and `failed` counts, `build_success` (boolean or null), `test_executed` (boolean), `timed_out` (boolean), and `status`. Runtime classification requires the Cargo build/artifact boundary above; neither a compiler summary lookalike nor the command exit alone establishes `test_executed`.

| Target status | Required interpretation |
| --- | --- |
| `FAILED` | Valid selected build/artifact evidence, an actual post-build stdout libtest `test result: FAILED.` summary, more than zero passed and failed tests, and nonzero command exit. Raw test-failure evidence, not an automatic audit verdict. |
| `survived` | Valid selected build/artifact evidence, an actual post-build stdout libtest OK summary, more than zero passed tests, zero failed tests, and exit 0. |
| `not_run` | Literal zero-passed result (including zero passed with failed tests), or missing runtime summary with exit 0 after a valid selected build. Not a survival or kill claim. |
| `error` | Compilation, invalid/missing/ambiguous build or artifact evidence, timeout, process, or unclassified nonzero command exit without the required stdout runtime evidence, including a stderr-only summary. Not a kill. |

The zero-passed policy is deliberately conservative: **`0 passed; 1 failed` remains `not_run`**, even though a test failed. Its original summary and exit remain in the evidence; the tool does not silently promote it to `FAILED`. A compiler error or Cargo nonzero exit alone never proves a killed mutation. No-summary failures are not evidence of a successful selection; consult the raw error output.

Global `completed` exits 0 even when a mutation survives. `not_run`, target errors, blocked input, source mismatches and restoration failures exit 1; usage errors may exit 2 while still producing JSON. Read the JSON, rather than treating the process exit as an audit decision.

## Repair boundary matrix

The parent observed `python3 -B -m unittest discover -v -s tests -p 'mutate*_test.py'` exit 0, **59 tests OK**, on 2026-11-11 with Python 3.14.2 and mirrored native Cargo/Rustc 1.94.1. That discovery run included `ActualCompilerTests` (6), `SyntheticCompilerProtocolTests` (14), `SourceSafety` (13), `DeadlineHandshakeTests` (10) and `MutationCLI` (16). The parent subsequently observed **1 test OK** in a separate execution of `DeadlineHandshakeTests.test_actual_cli_original_point_two_timeout_preserves_source_and_restores_clone`; the deadline class now has 11 cases, giving 60 distinct observed cases across the two runs, not a single 60-test discovery run. `passed` below records those regression-test observations, not an audit PASS or a claim that rejected inputs executed Rust tests. Test-ID prefixes map respectively to `tests/mutate_compiler_test.py` (first two classes), `tests/mutate_source_safety_test.py`, `tests/mutate_deadline_test.py` and `tests/mutate_test.py`; grouped method names share the stated class prefix. Native Windows and live SIGINT observations remain explicitly `not_run`.

### F1: build versus runtime evidence

Real offline Cargo/Rustc fixtures are required alongside fake-Cargo protocol tests; fake output alone is not compiler proof. The exact auditor proc-macro prints `test result: FAILED. 1 passed; 1 failed; 0 ignored; 0 filtered out;` and expands to `compile_error!("AUDIT COMPILE FAILURE: NO TEST EXECUTED");`.

| Planned case | Required boundary | Execution evidence |
| --- | --- | --- |
| F1-stderr: exact proc-macro spoof | Actual compile failure is target `error`, global nonzero, never `FAILED`; no tests executed. | `passed`: `ActualCompilerTests.test_actual_proc_macro_stderr_summary_is_compile_error_not_killed` and `SyntheticCompilerProtocolTests.test_synthetic_compiler_stderr_summary_rejected`. |
| F1-stdout: compile-time stdout spoof | Summary before successful build completion cannot establish runtime counts. | `passed`: `SyntheticCompilerProtocolTests.test_synthetic_compiler_stdout_before_build_finished_rejected` and `test_synthetic_stdout_forged_build_json_cannot_hide_failed_build`. |
| F1-multiline: multiline diagnostic | Embedded/newline-separated summary text cannot escape the build boundary. | `passed`: `ActualCompilerTests.test_actual_multiline_compile_error_summary_is_not_runtime` and `SyntheticCompilerProtocolTests.test_synthetic_multiline_diagnostic_summary_rejected`. |
| F1-JSON: diagnostic JSON string | Parse protocol records, not summary substrings inside strings. | `passed`: `SyntheticCompilerProtocolTests.test_synthetic_json_rendered_diagnostic_summary_rejected`. |
| F1-rendered: rendered compiler diagnostic | Rendered diagnostic text is not runtime stdout evidence. | `passed`: `SyntheticCompilerProtocolTests.test_synthetic_json_rendered_diagnostic_summary_rejected` and `ActualCompilerTests.test_actual_multiline_compile_error_summary_is_not_runtime`. |
| F1-build-script: build-script spoof | Build-script output cannot become a selected test's runtime summary. | `passed`: `ActualCompilerTests.test_actual_build_script_summary_and_panic_is_build_error` and `SyntheticCompilerProtocolTests.test_synthetic_build_script_summary_not_runtime`. |
| F1-build-markers: missing/failed completion | Missing build completion or any failed build cannot establish execution, even with later successful/spoofed records. | `passed`: `SyntheticCompilerProtocolTests.test_synthetic_missing_build_finished_rejected`, `test_synthetic_failed_build_cannot_be_overridden_by_success` and `test_synthetic_stdout_forged_build_json_cannot_hide_failed_build`. |
| F1-artifact: missing/wrong/non-test/nonexecutable/ambiguous artifact | Require exactly one matching executable test artifact. | `passed`: `SyntheticCompilerProtocolTests.test_synthetic_missing_selected_artifact_rejected`, `test_synthetic_wrong_selected_artifact_rejected`, `test_synthetic_non_test_artifact_rejected`, `test_synthetic_artifact_without_executable_rejected` and `test_synthetic_ambiguous_selected_artifacts_rejected`. |
| F1-runtime-stderr: valid build, stderr-only summary, exit 101 | No stdout runtime summary: nonzero exit is `error`, never `FAILED`. | `passed`: `SyntheticCompilerProtocolTests.test_synthetic_runtime_stderr_is_not_libtest_stdout`. |
| F1-runtime-failed: actual test failure | Real post-build stdout with one passed, one failed and nonzero exit is `FAILED`. | `passed`: `ActualCompilerTests.test_actual_two_rust_tests_one_pass_one_fail_kills_mutation`. |
| F1-runtime-ok: positive passed count | Valid build, positive passed count, zero failed and exit 0 is `survived`. | `passed`: `ActualCompilerTests.test_actual_passing_rust_test_survives_mutation` and `MutationCLI.test_status_is_based_on_real_libtest_counts_and_exit_not_any_failure`. |
| F1-zero: zero passed, including one failed | Preserve conservative `not_run`; never promote zero passed to a kill. | `passed`: `ActualCompilerTests.test_actual_zero_rust_tests_is_not_run` and `MutationCLI.test_status_is_based_on_real_libtest_counts_and_exit_not_any_failure`. |

### F2: tracked bytes and index semantics

Rejections require zero Cargo calls and unchanged literal bytes, raw index, HEAD and refs. Hidden dirty bytes use the auditor's exact `b'UNCOMMITTED SYNTHETIC BYTES\n'`; clean Git status is not the assertion.

| Planned case | Required boundary | Execution evidence |
| --- | --- | --- |
| F2-assume: assume-unchanged only | Reject hidden dirty bytes before Cargo without clearing the flag. | `passed`: `SourceSafety.test_hidden_assume_unchanged_neighbor_is_blocked_without_cargo`. |
| F2-skip: skip-worktree only | Reject hidden dirty bytes before Cargo without clearing the flag. | `passed`: `SourceSafety.test_hidden_skip_worktree_neighbor_is_blocked_without_cargo`. |
| F2-both: both flags | Reject exact auditor fixture with hidden bytes and both flags intact. | `passed`: `SourceSafety.test_hidden_combined_flags_neighbor_is_blocked_without_cargo`. |
| F2-clean-flags: HEAD-equal bytes with flags | Reject semantic flags even without a physical content difference. | `passed`: `SourceSafety.test_clean_assume_unchanged_flag_is_rejected_without_index_rewrite`, `test_clean_skip_worktree_flag_is_rejected_without_index_rewrite` and `test_clean_combined_flags_are_rejected_without_index_rewrite`. |
| F2-other: dirt on unselected tracked file | Validate every tracked entry, not only the mutation path. | `passed`: `SourceSafety.test_hidden_assume_unchanged_neighbor_is_blocked_without_cargo`, `test_hidden_skip_worktree_neighbor_is_blocked_without_cargo` and `test_hidden_combined_flags_neighbor_is_blocked_without_cargo`. |
| F2-selected: dirt on selected source | Compare physical source bytes against raw HEAD before execution. | `passed`: `SourceSafety.test_hidden_assume_unchanged_selected_file_is_blocked_without_cargo` and `test_hidden_skip_worktree_selected_file_is_blocked_without_cargo`. |
| F2-staged: staged divergence | Reject index content/modes differing from HEAD without resetting the index. | `passed`: `SourceSafety.test_staged_ordinary_neighbor_is_blocked_without_index_rewrite` and `MutationCLI.test_dirty_source_is_rejected_without_resetting_user_changes`. |
| F2-normal-mode: nonexecutable `0600` source | Do not confuse filesystem permissions with Git executable mode. | `passed`: `MutationCLI.test_clone_restoration_uses_its_own_checkout_mode`. |

### F3: no-follow snapshots

Use a synthetic private directory, never actual HOME or credentials. The exact auditor sentinel is `b'PRIVATE SYNTHETIC SENTINEL: MUST NOT READ\n'`. Rejection must prove zero external opens, zero Cargo calls and preserved source/index/HEAD/refs, not merely unchanged sentinel bytes.

| Planned case | Required boundary | Execution evidence |
| --- | --- | --- |
| F3-single-parent: external `src` directory symlink | Validate parent before leaf metadata/read in initial snapshot. | `passed`: `SourceSafety.test_directory_symlink_parent_blocks_with_zero_external_opens`. |
| F3-multilevel: symlink deeper in parent chain | Validate every parent, not only the immediate directory. | `passed`: `SourceSafety.test_multilevel_symlink_parent_blocks_with_zero_external_opens`. |
| F3-leaf: tracked leaf symlink | Snapshot literal link text without target reads; reject as mutation file. | `passed`: `SourceSafety.test_tracked_leaf_symlink_is_literal_with_zero_external_opens` and `MutationCLI.test_tracked_symlink_and_symlink_parent_are_rejected`. |
| F3-normal: ordinary tracked file | Permit regular paths with byte-exact snapshots. | `passed`: `MutationCLI.test_mutant_is_observed_once_per_explicit_target_in_independent_clone`. |
| F3-rejection-final: source rechecked on rejection | Final/error snapshot uses the same no-follow reader; still zero external opens. | `passed`: `SourceSafety.test_ordinary_blocked_input_final_snapshot_preserves_source`, `test_directory_symlink_parent_blocks_with_zero_external_opens` and `test_multilevel_symlink_parent_blocks_with_zero_external_opens`. |
| F3-clone: unsafe replacement in clone | Restoration snapshots must not dereference unsafe replacements. | `passed`: `MutationCLI.test_restoration_failure_stops_before_next_mutation_and_never_follows_symlink`. |
| F3-Windows-reparse: native reparse parent | Equivalent parent-by-parent rejection remains required. | `not_run`: native Windows host unavailable; not a platform scope exemption. |

### F4: deterministic deadlines and process lifecycle

Start the controlled deadline after owned spawn/job setup. READY/IPC synchronization establishes helper startup independently of wall-clock scheduling; neither sleep-based log assumptions nor an enlarged timeout substitute for this boundary.

| Planned case | Required boundary | Execution evidence |
| --- | --- | --- |
| F4-expired: negative remaining time | Set `timed_out: true` before `communicate`; do not pass a nonpositive timeout. | `passed`: `DeadlineHandshakeTests.test_expired_deadline_stops_before_communicate`. |
| F4-exact: zero remaining time | Equality is timeout, not a fresh full timeout. | `passed`: `DeadlineHandshakeTests.test_exact_deadline_is_expired_before_communicate`. |
| F4-not-expired: positive remaining time | Pass only remaining duration; normal completion leaves `timed_out: false`. | `passed`: `DeadlineHandshakeTests.test_not_expired_uses_remaining_budget`. |
| F4-startup-delay: READY after deadline | Real helper handshake plus controlled elapsed time proves timeout without subsecond startup assumptions. | `passed`: `DeadlineHandshakeTests.test_slow_ready_startup_expires_without_wall_clock_assumption`. |
| F4-original-timeout-integration: unchanged `timeout_seconds: 0.2` | Exercise `main` → `execute` → `run_target`, preserve source and restore the clone using READY plus a controlled clock, without a real 0.2-second wait. | `passed`: `DeadlineHandshakeTests.test_actual_cli_original_point_two_timeout_preserves_source_and_restores_clone`, separately executed after the 59-test discovery run (1 test OK). |
| F4-handshake-budget: handshake consumes deadline | Never reset timeout after the READY handshake. | `passed`: `DeadlineHandshakeTests.test_handshake_does_not_reset_original_deadline`. |
| F4-owned-start: delayed process creation | Begin deadline only after owned process/job setup. | `passed`: `DeadlineHandshakeTests.test_deadline_begins_after_owned_process_creation`. |
| F4-communicate-timeout: timeout during communicate | Catch `TimeoutExpired`, mark timeout and quiesce before restoration. | `passed`: `DeadlineHandshakeTests.test_communicate_timeout_sets_observable_timed_out`. |
| F4-cleanup-refused: stop cannot establish quiescence | Retain `quiescent: false`, `restored: false`, global `blocked`; stop subsequent mutations without bypassing refusal. | `passed`: `DeadlineHandshakeTests.test_expired_deadline_cleanup_refusal_is_fail_closed`; earlier host `EPERM` is not new successful-cleanup evidence. |
| F4-creation-error: spawn fails | Report process error without fabricated runtime or timeout; preserve source/restoration invariants. | `passed`: `DeadlineHandshakeTests.test_process_creation_error_is_not_timeout_and_has_no_child`. |
| F4-other-error: process exception | Non-timeout exceptions are not test kills; cleanup remains fail-closed. | `passed`: `DeadlineHandshakeTests.test_communicate_exception_is_error_not_timeout`. |
| F4-Windows-timeout: native Job Object lifecycle | Suspended spawn, job assignment/resume and child-tree timeout cleanup require native execution. | `not_run`: native Windows host unavailable. |
| F4-live-SIGINT: actual interruption | Quiesce owned children before restoration; termination refusal remains fail-closed. | `not_run`: no live SIGINT execution evidence supplied; injected interruption does not establish a live signal result. |


## Retained artifacts and manual cleanup

The clone, synthetic home, temporary files, logs and build output remain under the source worktree's owned `target/` paths; explicitly selected external caches remain at their configured locations. There is **no automatic deletion of artifacts, caches or homes**, including on failure. Inspect `workspace`, `cargo_target_dir`, diffs, raw target output and restoration evidence before any cleanup. Cleanup is a separate manual action: confirm the exact retained paths belong to this run and that no run still uses them. Do not remove source worktrees, another session's target, shared dependency caches, real HOME, credentials or preserved project state.

## Design sources

Official documentation checked on **2026-11-11** is recorded with read timestamps and reader-response SHA-256 values in [Issue #2403 design comment 6795796148](https://github.com/modu-ai/moai-adk/issues/2403#issuecomment-6795796148). Those records establish the source reads, not runner validation or test results.

- [Python subprocess](https://docs.python.org/3.14/library/subprocess.html): adopt argv execution with `shell=False`, explicit environment and timeout handling; reject arbitrary shell-command specifications. The cited documentation version is 3.14; the runner's syntax targets Python 3.9+.
- [Cargo test](https://doc.rust-lang.org/cargo/commands/cargo-test.html): adopt explicit single-target selection and raw test-summary evidence; reject target globs, combined target/filter calls, and interpreting any nonzero Cargo exit as a kill.
- [Cargo external tools / JSON messages](https://doc.rust-lang.org/cargo/reference/external-tools.html): use `compiler-artifact` and `build-finished` to separate build evidence from subsequent test stdout; compiler diagnostics and build-script output cannot establish a libtest runtime failure.
- [Git clone](https://git-scm.com/docs/git-clone): adopt an independent detached clone without local hardlink optimization; reject shared clones/build directories and direct source mutation followed by restore. `copytree` could mix in dirty/untracked files, while `git archive` can apply `export-ignore` or `export-subst` rather than preserving exact HEAD bytes, so neither is used.
- [Windows Job Objects](https://learn.microsoft.com/en-us/windows/win32/procthread/job-objects), checked **2026-11-11**: adopt native job-based child-process lifetime management rather than assuming POSIX process groups on Windows. Read time and SHA-256 are retained in the parent's `evidence/windows-official-sources.json`; this source evidence does not establish Windows runtime validation.

Automatic mutation generation, GitHub/CI publication and audit judgment remain out of scope. The parent-observed regression results above validate the exercised boundaries, not an audit verdict; native Windows and live SIGINT remain unexecuted.
