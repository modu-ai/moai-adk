---
id: SPEC-20261121-fail-to-pass-evidence
title: "A new or changed test is shown failing on the base before it passes on the head, by a tool the train and the audits can run"
status: draft
revision: 1
created: 2026-11-21
---

# A new or changed test is shown failing on the base before it passes on the head, by a tool the train and the audits can run

## Intent

The READY note and the audit say that a test was written first and failed first, but nothing in the repository shows it: the red run is the lane's own word, and an audit proves it only by hand-made mutants. A test that already passes on the base proves nothing about the change it is meant to pin.

The user asked for the root-cause report's urgent fixes to be done (2026-11-21); the report (`test-load-redesign-2026-11-21.md`, piece S12) asks for fail-to-pass evidence in the train and the audits only, because it costs a base build: the tests the head adds or changes are run on the base and are expected to fail, and the head then passes them.

This change adds one maintainer tool, `tools/f2p_evidence.py`, that lists the Rust integration tests (`tests/*.rs`) the head adds or changes against the base, runs them on the base source with the head's version of their test files, and prints one summary `f2p=<n> flagged=<m>`. A new test that passes on the base is flagged unless the caller records a reason (a pure refactor has none to fail). The tool reads the repository and writes nothing in it.

## Non-goals

- Wiring the summary into `moai check affected`: that file is held by another card; the tool's JSON line is what the wiring reads, and the wiring is a pair line to the holder.
- Python tests (`tests/*_test.py`) and unit tests inside `src/`: named as a limit, one layer at a time.
- Running the head after the base: the lane's own check already runs the head; this tool shows only the base side.
- Deciding whether a changed test was weakened: that is a separate signal.

## Requirements

- REQ-1: The tool shall list the Rust integration tests that the head adds to or changes in `tests/*.rs` against the base, by file and name, marking each `new` or `changed`.
- REQ-2: The tool shall run each listed test on the base source with the head's version of its test file and record whether it failed or passed there.
- REQ-3: When a listed test passes on the base and no reason is given, the tool shall flag it in its summary and exit 1. · failure: none — a recorded reason is REQ-4
- REQ-4: Where the caller gives a reason, the tool shall record it in the report and exit 0 although a listed test passes on the base.
- REQ-5: The tool shall print one JSON line on stdout whose `summary` reads `f2p=<n> flagged=<m>`, `n` counting the listed tests that failed on the base and `m` those that passed there without a reason.
- REQ-6: The tool shall report a test whose test file does not build on the base as failed there with the reason `build_error`.
- REQ-7: The tool shall refuse an unknown revision, a missing directory, a missing cargo program, a base that is not an ancestor of the head or a run that gives no verdict for a listed test with exit 2, a stderr line that starts `f2p_evidence: ` and an empty stdout.
- REQ-8: The tool shall write nothing in the repository and build the base source in a temporary folder that it removes.
- REQ-9: Where the caller asks for the list only, the tool shall print the list without running a test.

## Acceptance

- AC-1 (REQ-1, REQ-9): on a fixture repository whose head adds one test and changes another in `tests/a.rs` and leaves a third unchanged, the list holds exactly the added test as `new` and the changed one as `changed`; with the list-only option no program runs · verify: `python3 -m unittest tests.f2p_evidence_test.ListTest`
- AC-2 (REQ-2, REQ-5, REQ-6): the fake `cargo` of the fixture reports one listed test failing and one passing on the base: the summary reads `f2p=1 flagged=1` and exit is 1; when the base does not build, every listed test is failed with `build_error` and the summary reads `f2p=2 flagged=0` · verify: `python3 -m unittest tests.f2p_evidence_test.RunTest`
- AC-3 (REQ-3, REQ-4): with a reason the same fixture exits 0, the report holds the reason and `flagged` is 0; without it exit is 1 · verify: `python3 -m unittest tests.f2p_evidence_test.ReasonTest`
- AC-4 (REQ-7, REQ-8): each refusal exits 2 with the prefix on stderr and empty stdout, and after every run the fixture repository is byte for byte as before and the temporary folder is gone · verify: `python3 -m unittest tests.f2p_evidence_test.RefusalTest tests.f2p_evidence_test.CleanTest`

## Scope

- write: `tools/f2p_evidence.py` (new), `tests/f2p_evidence_test.py` (new), `.moai/specs/SPEC-20261121-fail-to-pass-evidence/**`
- paired (pair line to the holder of the file, after this lands): `src/check_affected.rs` and `src/ops/check.rs` print `f2p=<n>` from this tool's JSON in train mode and in audits
- read-only: `src/check_affected.rs`, the root-cause report
- forbidden: the Go repository, the real HOME, any card number in a file name, a SPEC id, a branch or a commit

## Plan

Tests first (RED). `tests/f2p_evidence_test.py` builds a temporary git repository per test (base commit with `tests/a.rs`, head commit that adds and changes a test) and a fake `cargo` (a script in the temporary folder that prints `test <name> ... ok|FAILED` for each name after `--exact`, as the environment tells it, and `error[E0432]` with exit 101 for the build case); the tool takes it through `--cargo`. On the tip alone the tests fail on assertions (the tool is absent) — the first red run uses the missing file as a failed import, so the red commit holds a stub-free test that fails on the missing tool.

Then the tool: parse `#[test]` functions of `tests/*.rs` at base and head with a brace counter (a function body is its text up to the matching brace), list names new at the head and names whose normalized body differs, extract the base with `git archive` into a temporary folder, copy the head's version of every changed test file over it, touch the extracted files (a stale modification time makes cargo reuse an old build), and run one `cargo test --test <stem> -- --exact <names>` per file; read the `test … ok|FAILED` lines.

## Risks

- Running a base build costs a heavy slot: the tool is for the train and the audits, and its docstring says to call it under the heavy-slot command.
- A test that needs the new code to compile fails with `build_error`; that counts as failing on the base, which is the evidence wanted, and is marked.
- A `#[test]` inside a macro or a generated name is not seen; a stated limit.
- A test in a file the head only renames is listed as new; the base has no such name.

## Result

(not yet written)
