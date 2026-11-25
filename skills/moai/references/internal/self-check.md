# Internal: self-check

When to read: after the change is made and before you ask for an audit or open a pull request, as `moai-impl` or any session that wrote the change. One line per check, with the command to run or the thing to look for. A line that does not fit the change or the project: skip it with the reason, under `Not run` or in the `## Self-Audit` section, never silently.

Run the project's own equivalents of each line that fits. Every build uses the worktree's build folder and `CARGO_INCREMENTAL=0` (`workflows/run.md`, step 1). An item that did not run, whatever the reason, is reported as not run, never as passed.

## Tests

1. Limits: every limit the change enforces has tests at N (passes), N+1 (fails) and empty (`internal/test-design`).
2. Guards: at audit tier 2 only, one mutation per new guard, the riskiest first, at most 12; below tier 2 none is asked for. Each must turn a test red; a surviving mutant is a missing test. Restore the file after each mutant with `git checkout -- <path>` or a plain `cp`, never `cp -p`: it brings back the old file time, cargo then judges the file unchanged and the next build still holds the mutant. At tier 2 record each as one `guard-mutant: <test> | <code removed or flipped> | exit <n>` line in the READY note (`flows/next-flow`); a mutant that survives is `exit 0` there and is repaired before READY, not reported. A tier 2 change with no new guard test writes `guard-mutant: none | <reason>`. A fake executable that records its run does so with the shell's own redirection, never with a command from PATH.
3. Receipt: run the `verify:` commands of the SPEC's criteria, then the receipt of `workflows/run.md`, step 7. The set before READY is that receipt and no other. Tests are Rust only. Write each command with its exit code in the READY note, and read every failing CI leg before you repair it.
4. Lint: `cargo clippy --offline --locked --all-targets -- -D warnings` and `cargo fmt --check`.
5. Files tests read: for each document or skill file you changed, `grep -rn '<path>' tests/` and run the tests that read it; `moai repo-scans --refs <tip> HEAD` lists the repo-wide scans the change set needs; run only their `cargo` checks, as `moai card land` asks for cargo checks only.
6. HOME: every test that runs the binary uses a temporary HOME.
7. Leftovers: a mutant copy of the tree, a temporary HOME for a dangerous-path run and any `mktemp` folder go only under `<worktree>/target/<name>/` (`TMPDIR=<worktree>/target/<name>` for a run that makes its own), and are removed by a relative path before READY; what could not be removed is named in the READY note's `leftovers:` line (`moai card note --help` shows it).
8. Paths: a new file read or write has tests for a symlink, a FIFO, a missing parent and a folder without permission; it refuses rather than follows.
9. Platform: a change to pty, process or signal code runs its macOS-only tests locally. An ordinary pull request's CI runs the light Linux matrix; the release pull request runs the full CI on Linux, macOS and Windows before it merges.

## Package and documents

10. Size: `moai package-inspect skills/moai` exits 0 and no package file is over 65,536 bytes. Check `wc -c` of each file you grew and leave a margin.
11. Manifest: after a change under `skills/moai/`, refresh the sha256 of each changed file in the source manifest with `shasum -a 256`; a new file needs its `origin` by hand.
12. Help: every example in a help text you changed runs as typed.
13. Docs: every statement you changed about a platform, version, default, limit, count or command matches the code or a command's output you can name; fix the text or the code before the audit.

## Pull request

14. Scope: `git diff --name-only <base>...HEAD` lists only paths inside the write scope, and nothing touches a held item.
15. Whitespace: `git diff --check <base>...HEAD` prints nothing, as in CI.
16. Commits: `git interpret-trailers --parse` on each new commit shows the `Moai-Spec` and `Moai-Session` trailers with the attribution last; no commit message, branch name or pull request title carries a card number.
17. Tier and body: the lane's ready note names the tier in a line `audit-tier: 0|1|2 <reason>`, with the changed paths as the reason (`workflows/audit.md`, step 9). A release pull request's body has `## Self-Audit`, which lists every command with its exit code and every item not run with its reason; check it with `moai pr-self-audit-check --file BODY.md`.
18. Acceptance criteria: each row is ticked with its evidence (command and exit code, or a comment link).
