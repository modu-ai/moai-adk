# Mutation runner

`tools/mutate.py` measures explicitly specified one-line mutations against narrow Cargo test targets. It emits raw evidence, not an audit verdict. A completed run, a `FAILED` target, or exit 0 does **not** mean audit PASS. The auditor decides whether the mutation and assertion establish the intended behavior.

## Prepare and invoke

Commit the intended source, tests, and runner before mutation. The runner requires a committed HEAD and a clean source Git index and working tree, including no untracked files. It does not commit for you, copy dirty changes, or repair the source tree. Keep the specification and output in the source worktree's ignored `target/` directory; an untracked specification elsewhere would fail the clean-tree guard.

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
cargo test --offline --locked --lib [filter]
cargo test --offline --locked --<kind> <name> [filter]
```

There is no shell-command field, arbitrary Cargo argument list, automatic baseline, retry, or generated mutation. Supply separate target entries rather than combining selectors in one call. In particular, a name filter must not silently select zero tests in a second target. Cargo output and exit status must both be inspected.

## Isolation and restoration evidence

The source worktree is read-only with respect to tracked files, raw index bytes and refs; the runner does create unique ignored artifacts beneath its `target/`. It removes inherited Git redirection variables and checks the actual source Git root rather than accepting an inherited alternate index or repository. It records the source HEAD and verifies source HEAD, raw index, tracked contents and cleanliness again at the end. It never runs restore or reset against the source.

For each run it retains an independent local clone beneath the source `target/`, cloned with `--no-local --no-hardlinks --no-checkout` and checked out detached at the exact committed HEAD. This is not a linked Git worktree; it has no shared hardlinks or alternates and does not write source refs or index. Dirty and untracked source files are not copied. Git commands and target execution use synthetic HOME, XDG and temporary directories by default, while respecting explicitly supplied toolchain/cache settings as described above.

Only the clone receives the replacement. The runner records the actual `git diff -U0` and checks that it contains one removed code line and one added replacement line, which may be a comment. On every outcome, `finally` handles restoration from saved HEAD bytes after target processes are stopped. Restoration checks HEAD, index stage entries and semantic flags (including assume-unchanged and skip-worktree), tracked file bytes, and `git diff --quiet HEAD`. Clone index stat-cache refreshes are distinct from semantic index changes and from the source's raw-index preservation requirement. A clean diff alone must not hide changed bytes or index flags; a successful test command alone is not restoration proof.

If a target replaces the mutation file with a symlink, FIFO, directory or other nonregular object, restoration must not follow it into another location. Unsafe or failed restoration is reported as false, exits unsuccessfully and stops subsequent mutations. The original source is never reset to hide a mismatch.

On timeout or interruption, the runner stops child processes before restoration, using a POSIX process group or a native Windows Job Object. On Windows the process is created suspended, assigned to the job immediately, then resumed; the job lifetime is independent of its parent process. If the OS refuses termination, the runner does not bypass that refusal or claim successful restoration: it retains the artifacts with `quiescent: false`, `restored: false`, global `blocked`, and no subsequent mutations. The parent observed POSIX `killpg` returning `EPERM` on the current host and exercises this fail-closed boundary rather than presenting it as successful restoration. Windows runtime execution is `not_run`.

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

Each mutation records `id`, `file`, actual `diff`, `targets`, `restored`, and `restore_diff_exit`. Each target records `argv`, `exit`, complete `stdout` and `stderr`, parsed `passed` and `failed` counts, and `status`:

| Target status | Required interpretation |
| --- | --- |
| `FAILED` | Actual libtest `test result: FAILED.` summary, more than zero passed and failed tests, and nonzero command exit. Raw test-failure evidence, not an automatic audit verdict. |
| `survived` | Actual libtest OK summary, more than zero passed tests, zero failed tests, and exit 0. |
| `not_run` | Literal zero-passed result, missing/zero-selected summary, or a target that did not run. Not a survival or kill claim. |
| `error` | Compilation, timeout, process, or unclassified command failure without the required test-failure evidence. Not a kill. |

The zero-passed policy is deliberately conservative: **`0 passed; 1 failed` remains `not_run`**, even though a test failed. Its original summary and exit remain in the evidence; the tool does not silently promote it to `FAILED`. A compiler error or Cargo nonzero exit alone never proves a killed mutation. No-summary failures are not evidence of a successful selection; consult the raw error output.

Global `completed` exits 0 even when a mutation survives. `not_run`, target errors, blocked input, source mismatches and restoration failures exit 1; usage errors may exit 2 while still producing JSON. Read the JSON, rather than treating the process exit as an audit decision.

## Retained artifacts and manual cleanup

The clone, synthetic home, temporary files, logs and build output remain under the source worktree's owned `target/` paths; explicitly selected external caches remain at their configured locations. There is **no automatic deletion of artifacts, caches or homes**, including on failure. Inspect `workspace`, `cargo_target_dir`, diffs, raw target output and restoration evidence before any cleanup. Cleanup is a separate manual action: confirm the exact retained paths belong to this run and that no run still uses them. Do not remove source worktrees, another session's target, shared dependency caches, real HOME, credentials or preserved project state.

## Design sources

Official documentation checked on **2026-11-11** is recorded with read timestamps and reader-response SHA-256 values in [Issue #2403 design comment 6795796148](https://github.com/modu-ai/moai-adk/issues/2403#issuecomment-6795796148). Those records establish the source reads, not runner validation or test results.

- [Python subprocess](https://docs.python.org/3.14/library/subprocess.html): adopt argv execution with `shell=False`, explicit environment and timeout handling; reject arbitrary shell-command specifications. The cited documentation version is 3.14; the runner's syntax targets Python 3.9+.
- [Cargo test](https://doc.rust-lang.org/cargo/commands/cargo-test.html): adopt explicit single-target selection and raw test-summary evidence; reject target globs, combined target/filter calls, and interpreting any nonzero Cargo exit as a kill.
- [Git clone](https://git-scm.com/docs/git-clone): adopt an independent detached clone without local hardlink optimization; reject shared clones/build directories and direct source mutation followed by restore. `copytree` could mix in dirty/untracked files, while `git archive` can apply `export-ignore` or `export-subst` rather than preserving exact HEAD bytes, so neither is used.
- [Windows Job Objects](https://learn.microsoft.com/en-us/windows/win32/procthread/job-objects), checked **2026-11-11**: adopt native job-based child-process lifetime management rather than assuming POSIX process groups on Windows. Read time and SHA-256 are retained in the parent's `evidence/windows-official-sources.json`; this source evidence does not establish Windows runtime validation.

Automatic mutation generation, GitHub/CI publication and audit judgment remain out of scope. This document specifies operation and evidence interpretation; it does not claim an observed test count, validation run or platform result.
