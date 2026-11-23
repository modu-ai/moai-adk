---
id: SPEC-20261124-pid-scan-followup
title: "The pid-name scan reads raw byte strings as raw and no longer skips its own file"
status: approved
revision: 1
created: 2026-11-24
approval: {by: leader, at: 2026-11-23, basis: "leader decision note 7783 of card C-421 (2026-11-23T18:37:42Z, delegation of the user 2026-11-23T13:57Z, the leader's judgement and not a selection by the user) answering request note 7782 with option (a)", digest: "sha256:be82130c0d1436c5e295d4b110f6f1cc607258837fdfde7d39c9f5258b34db2b"}
---

# The pid-name scan reads raw byte strings as raw and no longer skips its own file

## Interview

- none: 리더 배차가 요구사항을 전달했어요 — 카드 C-421 본문(AUDIT #417 의 비차단 N3·N4)이 범위를 정했고, 별도 사용자 인터뷰는 없었어요.

## Intent

SPEC-20261124-test-temp-names added `tests/test_temp_names_scan.rs`, a scan that refuses a statement under `tests/` holding both `temp_dir()` and `process::id()`. The independent audit of that card (AUDIT #417) passed it and recorded two non-blocking findings, which this SPEC repairs:

- N3, the lexer. `code_view` blanks comments and the inside of string literals so that text in them is not scanned. It opens a raw string only when the `r` is not preceded by an identifier character, so `br"…"` (raw byte string) and `cr"…"` (raw C string) are read as the identifier `b` or `c` followed by an ordinary string. An ordinary string treats a backslash as an escape, so `br"C:\"` swallows its closing quote and the scan reads the following code as string text: a violation on the next line is missed.
- N4, the self skip. `temp_names_tree_has_no_pid_names_under_the_temp_folder` skips `tests/test_temp_names_scan.rs`. The file needs no skip (its two calls are inside string literals and the tree test passes with the skip off), and the skip would hide a real violation written into the scan file itself.

Grounded in the code at the integration tip `ec6233e7b4`: the raw branch of `code_view` is `tests/test_temp_names_scan.rs:59-76`, its `SELF_FILE` skip is `:19-20` and `:251-253`.

## Non-goals

- The other misses the audit listed under N3, which stay a follow-up and are not designed here: a `temp_dir()` bound to a variable in an earlier statement with `process::id()` in a later statement of the same function, `use std::process::id` followed by a bare `id()`, a `{}` block inside the `join` argument, and `TMPDIR` or `CARGO_TARGET_TMPDIR` joined with a pid (outside the rule of the first SPEC). Each is a rare form; a card opens when one produces a red run.
- Any change of `tools/repo_scans.json`: the scan's id, test name, check command and scope stay as they are.
- The pid names of `tests/hook_dispatch_cli.rs` and `tests/host_wake_cli.rs`, and the unit tests in `src/`: listed in SPEC-20261124-test-temp-names.
- Byte character literals (`b'x'`) and other literal prefixes: the character branch already reads `'x'`, and no other prefix swallows a quote.

## Requirements

- REQ-1: When a raw byte string or raw C string (`br"…"`, `br#"…"#`, `cr"…"`) holds a backslash or a quote, the scan test shall read it as one raw literal and scan the code that follows it · failure: none — a violation on a later line is still reported with its line.
- REQ-2: When the scan walks `tests/`, the scan test shall read every file, the scan file itself included, with no skip by file name · failure: none — the existing tree test passes unchanged, and a violation planted in a file named like the scan file is reported.

## Acceptance

- AC-1 (REQ-1): a statement joining a pid name onto `temp_dir()` on the line after a raw byte string that ends in a backslash, after one with hashes that holds a quote, and after a raw C string that ends in a backslash is reported with its line, and `temp_dir()` and `process::id()` written inside such a literal are not reported · verify: `cargo test --locked --offline --test test_temp_names_scan temp_names_scan_reads_raw_byte_strings_as_raw`
- AC-2 (REQ-2): a scratch project whose `tests/test_temp_names_scan.rs` holds a pid name under the temp folder is reported by the tree walk, and the checked-in `tests/` folder holds no such statement with the skip gone · verify: `cargo test --locked --offline --test test_temp_names_scan temp_names_tree_`

## Scope

- write: `.moai/specs/SPEC-20261124-pid-scan-followup/spec.md`, `tests/test_temp_names_scan.rs`
- read-only: every other file under `tests/`, `tools/repo_scans.json`
- forbidden: `src/`, `Cargo.toml`, `tools/repo_scans.json` (no entry changes: the same scan)

## Plan

1. This SPEC, then the two new test cases and the tree walk split out of the tree test; both red on the current file (AC-1: the raw byte cases; AC-2: the planted file).
2. The raw branch of `code_view` takes the `b` and `c` prefixes; the file-name skip and `SELF_FILE` go; the tree test calls the walk.
3. Mutants for each new guard, one affected run at the READY head, the READY note.

## Risks

- A lexer is a text match. The prefixes covered are those that can open a raw string today (`r`, `br`, `cr`); a later Rust prefix is a follow-up.
- The walk is split into a function that takes the root so a test can plant a file; the tree test keeps its own floor of 100 files, so a walk that finds nothing still fails.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | The optional tracking of a `temp_dir()` variable and of `use std::process::id` stays out | The card marks it optional; both are rare forms and have produced no red run (core rule 20) | A follow-up card with its own SPEC |
| D-2 | The self skip is removed, not replaced by an allow table | The file needs no skip; a table would keep the hiding possible | Re-add a skip with a reason, if a fixture ever needs it |
| D-3 | No change to `tools/repo_scans.json` | Same scan, same test name, check command and files | none needed |

## Result

Not yet implemented (draft, waiting for the leader's approval).
