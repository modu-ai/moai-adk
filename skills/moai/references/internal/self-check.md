# Internal: self-check

When to read: after the change is made and before you ask for an audit or open a pull request, as `moai-impl` or any session that wrote the change. One line per check, with the command to run or the thing to look for. A line that does not fit the change or the project: skip it with the reason, under `Not run` or in the `## Self-Audit` section, never silently.

In the moai-adk repository, `tools/self_check.sh` runs the machine items in one go (`bin-moai`, `bin-moai-cli`, `clippy`, `fmt`, `fmt-ops`, `git-env-scrub`, `home-isolation`, `ac-command-lint`, `package-inspect`, `package-size`, `manifest`, `diff-check`, `python`) and prints one `self-check: <item> exit=<code>` line each; copy those lines into the `## Self-Audit` section. `--full-python` runs CI's whole Python suite instead of the touched files, in the python slot: its own slot folder (`MOAI_PYTHON_SLOT_DIR`, default `${XDG_CACHE_HOME:-$HOME/.cache}/moai-cli/python-slots`) with at most `MOAI_PYTHON_SLOT_MAX` runs at a time (default 2), so a whole Python run never takes a cargo slot. A heavy or Python item the machine's slot refused prints `exit=75 slot-refused (not run)`: report it as not run, never as passed. `--pre-ready` is the whole-tree list a head passes before the lane reports it ready: `diff-check`, `fmt`, `fmt-ops`, `manifest`, `spec-check`, `py-source-manifest`, `py-repo-scans`, `skill-package`, fast ones first; each item prints one `self-check:` line (`exit=<code>`, `skip (<reason>)` or `not_run (<reason>)`), and `--pre-ready` and `--full-python` cannot be given together. `py-source-manifest` and `py-repo-scans` run in the python slot, `skill-package` in the heavy slot, the other items without a slot. `spec-check` and `py-repo-scans` need a built `moai-cli` (`MOAI_CLI_BIN`, or `${CARGO_TARGET_DIR:-target}/debug/moai-cli`): when it is missing or older than `src/`, `build.rs`, `Cargo.toml` or `Cargo.lock`, the item prints `not_run (no built moai-cli: <path>)` or `not_run (stale moai-cli: <path> is older than <file>)`; build it first (`internal/build-folder`). An item that printed `not_run`, whatever the reason, is reported as not run, never as passed. The Windows type check is not an item: run it by hand when the change touches platform-specific code (`workflows/verify.md`). Elsewhere, run the project's own equivalents. Every build here uses the worktree's build folder and `CARGO_INCREMENTAL=0` (`internal/build-folder`).

## Tests

1. Limits: every limit the change enforces has tests at N (passes), N+1 (fails) and empty (`internal/test-design`).
2. Guards: one mutation per new guard, the riskiest first; at most 5 for audit tier 1 and 12 for tier 2. Each must turn a test red; a surviving mutant is a missing test.
3. Whole binaries: `cargo test --offline --locked --bin moai` and `cargo test --offline --locked --bin moai-cli`, whole: never with a test-name filter. The help rules test every command there, and a filtered run misses them.
4. Lint: `cargo clippy --offline --locked --all-targets -- -D warnings`, `cargo fmt --check`, and `rustfmt --check --edition 2021 src/ops/*.rs`: `build.rs` generates the module list of the operation files, so `cargo fmt` never sees them.
5. Files tests read: for each document or skill file you changed, `grep -rn '<path>' tests/` and run the tests that read it; `moai repo-scans --refs <tip> HEAD` lists the repo-wide scans the change set needs.
6. Python: the unittest files of the tests and tools you changed, plus `tests/skill_references_test.py`, which reads every `skills/...` path in a tracked file; CI runs `python3 -m unittest discover -s tests -p '*_test.py'` (`--full-python`).
7. Git: a new git spawn in `src/` goes through `scrub_git_env` (`--test git_env_scrub`); a test that runs git uses `tests/common/git.rs` (`--test ac_command_lint`).
8. HOME: every test that runs the binary uses a temporary HOME (`--test home_isolation_scan`).
9. Paths: a new file read or write has tests for a symlink, a FIFO, a missing parent and a folder without permission; it refuses rather than follows.
10. Platform: a change to pty, process or signal code runs its macOS-only tests locally; pull request CI runs Linux x86_64 only.

## Package and documents

11. Size: `moai package-inspect skills/moai` exits 0 and no package file is over 65,536 bytes (`profile::LIMIT`). Check `wc -c` of each file you grew and leave a margin.
12. Manifest: after a change under `skills/moai/`, in the moai-adk repository run `tools/source_manifest.py write`, then `tools/source_manifest.py check` (both with `python3 -I`); a new file needs its `origin` by hand.
13. Help: every example in a help text you changed runs as typed.
14. Docs: every statement you changed about a platform, version, default, limit, count or command matches the code or a command's output you can name; fix the text or the code before the audit.

## Pull request

15. Scope: `git diff --name-only <base>...HEAD` lists only paths inside the write scope, and nothing touches a held item.
16. Whitespace: `git diff --check <base>...HEAD` prints nothing, as in CI.
17. Commits: `git interpret-trailers --parse` on each new commit shows the attribution last, after the claim token where the branch has a claim; card branches carry no claim token until C-53 S3.
18. Body: the first line is `audit-tier: 0|1|2 <reason>`; `## Self-Audit` lists every command with its exit code and every item not run with its reason; check it with `moai pr-self-audit-check --file BODY.md`.
19. Acceptance criteria: each row is ticked with its evidence (command and exit code, or a comment link).
