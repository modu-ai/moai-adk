---
id: SPEC-20261123-ac-lint-work-count
title: "The AC command lint test bounds work, not seconds"
status: draft
revision: 1
created: 2026-11-23
---

# The AC command lint test bounds work, not seconds

## Intent

`the_repository_lint_starts_no_cargo_and_ends_within_five_seconds` in `tests/ac_command_lint.rs` decided by the clock: the lint over the repository had to end in under 5 s. It passed in 1.4–2.3 s on a quiet machine and failed at 5.18 s under machine load 200 on 2026-11-22. A time limit cannot tell a slow machine from a slow lint, and `skills/moai/workflows/verify.md` says a test must not depend on wall-clock time. This is the second test of the same class after the card secret screen (`SPEC-20261120-card-screen-work-count`), so it gets the same cure: count the work. What the old limit protected is that the lint stays a pure reader: it starts no process (the 68–112 s cargo build of `SPEC-20261122-ac-lint-no-cargo` was a spawned process) and it reads each recorded file and each test source once (`list_tests` caches one reading per target). Both facts are countable, and no load changes a count. The readers are the people who run `check affected` and the auditors who read a red run.

## Non-goals

- Changing `tools/check_ac_commands.py`; the test observes the tool, it does not change it.
- Replacing the other elapsed-time assertions of the repository; they are listed in the Result, and only one that has failed under load is a reason to change it.
- Keeping any second-based limit, not even a backstop that never decides.

## Requirements

- REQ-1: While the lint runs over the repository, the lint test shall fail when the interpreter's audit events show that the lint started any process.
- REQ-2: While the lint runs over the repository, the lint test shall fail when the lint opens a recorded document or a target root more than once, or a shared test module more often than the number of targets it listed.
- REQ-3: While the lint runs over the repository, the lint test shall fail when the lint opens a repository file outside `docs/`, `.moai-cli/reports/` and `tests/`.
- REQ-4: While the lint runs over the repository, the lint test shall fail when the counter saw fewer file opens than the recorded files the lint reports having checked, so a counter that stopped counting cannot pass.
- REQ-5: The lint test shall not read the clock to decide its result.

## Acceptance

- AC-1 (REQ-1, REQ-2, REQ-3, REQ-4): the lint over the repository exits 0 with 0 processes started, every file within its open bound and no open outside the three folders · verify: `cargo nextest run --locked --offline --test ac_command_lint`
- AC-2 (REQ-1): a lint that starts one process of any name makes the test fail · verify: mutant in the Result, `cargo nextest run --locked --offline --test ac_command_lint`
- AC-3 (REQ-2): a lint that reads each target's sources again for every command (the cache removed) makes the test fail, under the same load as a quiet run · verify: mutant in the Result, `cargo nextest run --locked --offline --test ac_command_lint`
- AC-4 (REQ-4): a counter that records nothing makes the test fail · verify: mutant in the Result, `cargo nextest run --locked --offline --test ac_command_lint`
- AC-5 (REQ-5): the test file has no clock read, and the test passes 30 of 30 times with a machine load of 100 or more · verify: `rg -n "Instant|elapsed|Duration|SystemTime" tests/ac_command_lint.rs` prints nothing, and the load run in the Result

## Scope

- write: `.moai/specs/SPEC-20261123-ac-lint-work-count/spec.md`, `tests/ac_command_lint.rs`
- read-only: `tools/check_ac_commands.py`, `tests/*.rs` (the lint's input), `docs/`
- forbidden: `tools/check_ac_commands.py`, `src/`, `.config/nextest.toml`

## Plan

1. RED: run the two mutants (a process started; the cache removed) against the unchanged test: the old test stays green on both, because neither adds seconds on a quiet machine.
2. GREEN: replace the time assertion with a small interpreter harness that sets an audit hook, runs the lint script unchanged, and writes the opened paths and the process-start count to a file; the test asserts REQ-1 to REQ-4 from that file.
3. Mutants, the 30-run load check, one `moai check affected`, HANDOFF READY.

## Risks

- The harness counts what the interpreter reports through audit events (`open`, `subprocess.Popen`, `os.system`, `os.exec*`, `os.posix_spawn`, `os.spawn*`, `os.fork*`). A child started by a C extension that bypasses those events is not seen; the lint imports `re`, `sys` and `pathlib` only, and the fake `cargo` on `PATH` stays as a second witness for the one process this test was written for.
- A read the lint does through a path outside the three folders (a new scan root) fails REQ-3 until the folder list is widened on purpose.
- The count is exact only for the lint as it is; a legitimate second reading of one file would fail REQ-2 and needs the cache or this bound changed on purpose, never a larger number.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | Observe the lint from outside with an interpreter audit hook instead of adding counters to `tools/check_ac_commands.py` | the leader's dispatch names only the test file; a hook leaves the tool and its scope untouched | add a counter line to the tool's summary and assert on it |
| D-2 | Keep the fake `cargo` on `PATH` besides the process count | the SPEC `SPEC-20261122-ac-lint-no-cargo` named it; it states the protected fact in its own words | delete the marker once the process count is trusted |
