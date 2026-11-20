---
id: SPEC-20261120-port-ac-command-lint
title: "Port the recorded-AC-command lint to the moai ac-commands command (tools slice 2)"
status: draft
revision: 1
created: 2026-11-20
---

## Intent

The tools slice 1 survey (the card's table, order 1) picks the recorded-AC-command lint as the
first port: `tools/check_ac_commands.py` (174 lines) scans recorded verification commands in
docs and reports, and every `cargo test --test <name> [filters…]` line must name an existing
Rust test target and select at least one test — cargo exits 0 with "running 0 tests" for a
filter that matches nothing, which has passed as verification more than once (#2352, survey
2.11). The check is registered in the scans table as `recorded-ac-commands`, but its Rust test
`tests/ac_command_lint.rs` is only a wrapper: `run_lint` spawns `python3 -I
tools/check_ac_commands.py` — the exact Python-only shape this card removes, because the
product's behaviour is tested in Rust only (user decision 2026-11-20).

Survey (read-only): the tool's contract is — `--tree <root>` or `--file <path>`, anything
else is usage (exit 2); the tree scan set is `docs/**/*.md`, `.moai-cli/reports/*.md`, and
`.moai-cli/reports/issue-bodies/*.md` whose leading issue number is 2153 or later (frozen
RC1 records stay out); a candidate is a `cargo test` line in a fenced code block or a
backtick code span, env prefixes (`K=V ` repetitions) are stripped, a backtick is stripped at
the ends; the target must exist as `tests/<name>.rs` and its positional filters (after
`--test <name>`, up to the next flag, `<`-placeholders excluded) must select at least one
test in `cargo test --offline --locked --test <name> -- -- --list`'s `path: test` lines — one
listing per target, cached; cargo's own filter rule decides (no filter selects everything, a
filter matches as a substring of the full test path). Output: one `checked N recorded files,
M --test targets listed` line, one `VIOLATION: <file>: …` line per violation, exit 0/1.
The HANGUL regex in the tool has no caller — dead, not ported.

## Non-goals

- The scans-table entry's identity: `recorded-ac-commands` keeps its id, tests and checks —
  the Rust test keeps running `cargo test --test ac_command_lint …`; only the wrapper inside
  it changes and the rules list stops naming the Python file.
- The git-fixture lint and the nextest-alone guard in the same test file: untouched.
- The other 25 tools of the slice 1 table and their order: later slices.
- The `moai_cli::` library module: the logic stays in the one ops file (the registry rule
  adds one file; `src/lib.rs` is reserved by another card this round — the ops file is the
  whole port, with no library export, recorded as a decision).
- Changing what the lint flags: a violation before the port is a violation after.

## Requirements

- REQ-1: While `moai ac-commands-check` runs with `--tree <root>` or `--file <path>`, the command shall scan exactly the same recorded files the Python tool scans — the tree's docs markdown, root reports markdown, and issue-body records numbered 2153 or later, or the one file named. · failure: none — pinned by AC-2.
- REQ-2: While a candidate line carries a `cargo test --test <name>` command, the command shall flag the exact violations the Python tool flags — a missing `tests/<name>.rs`, a target whose listing has no tests, and filters that select none — carrying the same one-line wording and the same summary line (printed with the failure lines, as an envelope operation reports its detail), and keeping the same exit decisions: usage is the envelope's exit 2, any violation makes the command fail with exit 1, a clean tree succeeds with exit 0 with the summary in its result. · failure: none — pinned by AC-3.
- REQ-3: While the lint runs, the command shall spend at most one cargo listing per test target for the whole invocation (the cache), reading the listing from `cargo test --offline --locked --test <name> -- --list`. · failure: none — pinned by AC-4.
- REQ-4: While the scans table names the check, its rules shall point at the Rust port and not at the Python tool, with the Python tool and its wrapper state in the test file gone in the same change. · failure: none — pinned by AC-5.

## Acceptance

- AC-1 (REQ-1, REQ-2): the test file's wrapper runs `moai ac-commands-check` instead of `python3 -I tools/check_ac_commands.py`, and the existing contract tests keep pinning the tree scan (success on the tree), the zero-selection flag (failure, the file named) and the nextest-alone guard — the wrapper change is the RED step: with no command yet, they fail first.
  verify: `cargo test --locked --offline --test ac_command_lint`
- AC-2 (REQ-1): new tests pin the scan-set edges the wrapper never pinned — an issue-body record numbered 2152 stays unjudged while 2153 is judged, an env-prefixed command is judged after the prefix is stripped, a placeholder filter is skipped, and a fenced block line and a backtick span are both judged.
  verify: `cargo test --locked --offline --test ac_command_lint`
- AC-3 (REQ-2): new tests pin the three violations' wording classes (missing target, target with no tests, zero-selection) and the summary line's shape on a fixture tree.
  verify: `cargo test --locked --offline --test ac_command_lint`
- AC-4 (REQ-3): the cache — a fixture tree with two recorded files naming the same target runs one listing, observed through a stub `cargo` that counts its calls, so the count is asserted, not inferred from time.
  verify: `cargo test --locked --offline --test ac_command_lint`
- AC-5 (REQ-4): the scans table's rules name `src/ops/ac_commands_check.rs` and no Python file; `tools/check_ac_commands.py` is deleted in the same change; the scans tests keep passing over the table.
  verify: `cargo test --locked --offline --test ac_command_lint`; `python3 -m unittest tests.repo_scans_test`
- AC-6 (REQ-1, REQ-2, REQ-3, REQ-4): neighbours — the whole lint test module, the scans registration, and the ops registry keep today's behaviour; the deleted file is referenced by nothing (repo-wide grep, run as the check the audits use).
  verify: `cargo test --locked --offline --test ac_command_lint`; `python3 -m unittest tests.repo_scans_test`; `moai spec check SPEC-20261120-port-ac-command-lint`

## Scope

- write: `src/ops/ac_commands_check.rs`, `docs/design-review/appendix/operations/ac_commands_check.md`, `tests/ac_command_lint.rs`, `tools/repo_scans.json`, `tools/check_ac_commands.py` (deleted), this SPEC folder
- read-only: `tools/protected_checks.json`, `tests/common/` (the bin-path helpers), `src/ops/mod.rs` (untouched — `build.rs` picks the file up), `src/main.rs` (C-140's reservation — the raw-text precedent stays a precedent)
- reserved by other cards: `src/lib.rs` (C-183), `src/main.rs` (C-140) — the port is the one ops file and carries its logic there; the take's `src/ops/ac_commands.rs` and `docs/…/ac_commands.md` patterns stay unused and go back with the release
- forbidden: changing the scans entry's id, tests or checks, the git-fixture lint, the nextest-alone override, the real HOME

## Plan

1. Commit this SPEC.
2. RED: the test wrapper runs `moai ac-commands-check` (the bin from `tests/common`, as `MOAI_CLI_BIN`) — the two contract tests fail with no command yet.
3. GREEN: `src/ops/ac_commands_check.rs` — clap Args (`--tree <root>` | `--file <path>`, an exclusive required group, so usage is clap's exit 2), the scan set, the fenced/backtick candidate extraction, token cleanup, the three violations with the tool's wording, the summary line, the cargo listing with a per-target cache; the document fragment beside it.
4. Behaviour parity: before the deletion, run the Python tool and the new command on the same tree and compare their violation sets line by line — the comparison recorded in the card.
5. Switch and delete in one commit: repo_scans.json's rules name `src/ops/ac_commands_check.rs`; the Python file is gone; the test file keeps no wrapper to it.
6. Neighbours and the light checks, then push and the audit request note (this slice is `.moai` tools + tests — no protected path; tier per the scans table's own rule).

## Risks

- The listing step runs cargo per new target name (cached per invocation) — the nextest-alone override for this test file (#2661) stays exactly as it is, since the port still spawns `cargo test -- --list`.
- `--file` resolves the root as the caller's cwd in the Python tool; the command keeps that (a single file is judged against `./tests`), and a test pins it so the shape cannot drift.
- The Python tool's violations join the Python list display (`['a', 'b']`) into the one line — the port reproduces the same display shape so recorded outputs stay comparable; a test pins it.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D1 | The port is one ops-envelope operation with the tool's logic inside the file, and its exit decisions ride the envelope: usage 2 (clap), violations 1 (`Err`), clean 0 (`Ok`) | `src/main.rs` is C-140's reservation this round (the raw-text precedent stays out of reach) and `src/lib.rs` is C-183's; the envelope's measured exit contract (Ok→0, Err→1) preserves the tool's 0/1 and clap preserves 2 | A raw-text variant beside `pr-self-audit-check` when `src/main.rs` returns |
| D2 | The subcommand is `ac-commands-check` (`src/ops/ac_commands_check.rs`) with the tool's exact scan, judgement and wording | Same-behaviour port; the `-check` suffix is the checker family's naming (gate-check, labels-check, parity-check) | Rename with the callers |
| D3 | The scans entry keeps its id, tests and checks; only `rules` loses the Python file and names the ops file | The check is the Rust test — the rules list is its reference set | A new entry id |

## Result

구현 에이전트가 채워요 — 이 커밋은 SPEC 과 계획까지예요.
