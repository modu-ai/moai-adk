# Internal: self-check

When to read: after the change is made and before you ask for an audit or open a pull request, as `moai-impl` or any session that wrote the change. One line per check, with the command to run or the thing to look for. A line that does not fit the change or the project: skip it with the reason, under `Not run` or in the `## Self-Audit` section, never silently.

In the moai-adk repository, `tools/self_check.sh` runs the machine items in one go (`bin-moai`, `bin-moai-cli`, `clippy`, `fmt`, `git-env-scrub`, `home-isolation`, `ac-command-lint`, `package-inspect`, `package-size`, `manifest`) and prints one `self-check: <item> exit=<code>` line each; copy those lines into the `## Self-Audit` section. Elsewhere, run the project's own equivalents.

## Tests

1. Limits: every limit the change enforces has tests at N (passes), N+1 (fails) and empty (`internal/test-design`).
2. Guards: one mutation per new guard, at most five per change; each must turn a test red. A surviving mutant is a missing test.
3. Whole binaries: `cargo test --offline --locked --bin moai` and `cargo test --offline --locked --bin moai-cli`, whole: never with a test-name filter. The help rules test every command there, and a filtered run misses them.
4. Lint: `cargo clippy --offline --locked --all-targets -- -D warnings` and `cargo fmt --check`.
5. Files tests read: for each document or skill file you changed, `grep -rn '<path>' tests/` and run the tests that read it; `moai repo-scans --refs <tip> HEAD` lists the repo-wide scans the change set needs.
6. Git: a new git spawn in `src/` goes through `scrub_git_env` (`--test git_env_scrub`); a test that runs git uses `tests/common/git.rs` (`--test ac_command_lint`).
7. HOME: every test that runs the binary uses a temporary HOME (`--test home_isolation_scan`).
8. Paths: a new file read or write has tests for a symlink, a FIFO, a missing parent and a folder without permission; it refuses rather than follows.
9. Platform: a change to pty, process or signal code runs its macOS-only tests locally; pull request CI runs Linux x86_64 only.

## Package and documents

10. Size: `moai package-inspect skills/moai` exits 0 and no package file is over 65,536 bytes (`profile::LIMIT`). Check `wc -c` of each file you grew and leave a margin.
11. Manifest: after a change under `skills/moai/`, rewrite `references/source-manifest.json` and check it (`write`, then `check`); a new file needs its `origin` by hand.
12. Help: every example in a help text you changed runs as typed.
13. Docs: each row or sentence you changed says what the code does (`internal/doc-claims`).

## Pull request

14. Scope: `git diff --name-only <base>...HEAD` lists only paths inside the write scope, and nothing touches a held item.
15. Commits: `git interpret-trailers --parse` on each new commit shows the claim token and the attribution last.
16. Body: the first line is `audit-tier: 0|1|2 <reason>`; `## Self-Audit` lists every command with its exit code and every item not run with its reason; check it with `moai pr-self-audit-check --file BODY.md`.
17. Acceptance criteria: each row is ticked with its evidence (command and exit code, or a comment link).
