---
id: SPEC-20261121-check-receipts
title: "check affected leaves a receipt for a passing tree and reuses it for the same tree"
status: draft
revision: 1
created: 2026-11-21
---

# check affected leaves a receipt for a passing tree and reuses it for the same tree

## Intent

같은 내용을 레인의 READY, 감사자, 리더의 병합 뒤 검사가 각각 다시 돌려요. tip 이 움직여서 같은 결과의 병합을 다시 만들 때도 마찬가지예요. 2026-11-21 조사에서 한 카드가 전체 검사를 세 번 넘게 돌렸고(`test-load-redesign-2026-11-21.md` P4·S6), 그때마다 cargo 단계가 수 분씩 걸렸어요.

이 SPEC 은 `moai check affected` 가 모든 단계가 통과한 검사의 영수증을 남기고, 같은 트리(tree)를 같은 단계 목록과 같은 도구 버전으로 다시 검사하면 단계를 하나도 돌리지 않고 영수증을 보여 주게 해요. 영수증은 기계가 만든 사실만 담아요. 감사가 어디까지 다시 돌릴지는 이 SPEC 이 정하지 않고, 감사가 새로 돌리고 싶을 때를 위해 `--fresh` 만 줘요.

## Non-goals

- 감사자의 독립 규칙을 바꾸는 일(감사가 영수증을 받아들일지는 리더·사용자 결정이에요).
- 실패한 검사의 영수증, 단계 하나만 재사용하는 일.
- 열차 모드나 선택 방식의 변경(다른 카드의 일).
- 영수증을 다른 기계나 다른 프로젝트와 나누는 일: 영수증은 이 프로젝트의 로컬 파일이에요.

## Requirements

- REQ-1: When every step of a `check affected` run exits 0, the run shall write one receipt named by its key under the check-receipts folder of the project state root, the key being the digest of the head tree id, the planned steps and the versions of cargo and rustc. · failure: none — REQ-2 covers a run with a failing step.
- REQ-2: If a step of a `check affected` run exits non-zero, then the run shall write no receipt.
- REQ-3: When a receipt for the key exists and `--fresh` is not given, `check affected` shall run no step and answer the stored step results with `receipt` state `reused` and the word `receipt=reused` before `result=PASS` in the summary line. · failure: none — REQ-6 covers a receipt that cannot be used.
- REQ-4: If the head tree id, the planned steps or a tool version differs from a receipt's, then `check affected` shall run every step.
- REQ-5: When `--fresh` is given, `check affected` shall run every step and, when all pass, replace the receipt of that key. · failure: none — REQ-2 covers a failing run.
- REQ-6: If a receipt file cannot be read, is not valid JSON or names another key, then `check affected` shall ignore it and run every step.
- REQ-7: The receipt write shall go through a temporary file in the same folder and a rename, refusing a symlinked receipt folder.
- REQ-8: If the tool versions cannot be read, then `check affected` shall run every step and write no receipt.
- REQ-9: The answer of every run shall carry `receipt` with its `state` (`written`, `reused`, `fresh` or `none`) and its `key`.

## Acceptance

- AC-1 (REQ-1, REQ-9): a run whose steps all pass leaves one receipt file under the check-receipts folder of the state root and the answer reads `receipt` state `written`.
  verify: `cargo test --locked --offline --test check_receipt_cli a_passing_run_leaves_one_receipt`
- AC-2 (REQ-2, REQ-9): a run with a failing step leaves no receipt and the answer reads state `none`.
  verify: `cargo test --locked --offline --test check_receipt_cli a_failing_run_leaves_no_receipt`
- AC-3 (REQ-3, REQ-9): the same tree checked again runs no step, keeps `result=PASS`, and the summary line carries `receipt=reused` before the result.
  verify: `cargo test --locked --offline --test check_receipt_cli the_same_tree_reuses_the_receipt`
- AC-4 (REQ-4): a changed tree and a changed tool version each run every step.
  verify: `cargo test --locked --offline --test check_receipt_cli any_change_runs_every_step`
- AC-5 (REQ-5, REQ-9): `--fresh` runs every step, answers state `fresh` and the receipt is replaced.
  verify: `cargo test --locked --offline --test check_receipt_cli fresh_runs_and_replaces`
- AC-6 (REQ-6): an unreadable, malformed or wrong-key receipt is ignored and every step runs.
  verify: `cargo test --locked --offline --test check_receipt_cli a_broken_receipt_is_ignored`
- AC-7 (REQ-7): a receipt folder that is a symlink gets no receipt and the run still gives its answer.
  verify: `cargo test --locked --offline --test check_receipt_cli a_symlinked_folder_gets_no_receipt`
- AC-8 (REQ-8): when `cargo --version` cannot be read, every step runs and no receipt is written.
  verify: `cargo test --locked --offline --test check_receipt_cli unreadable_versions_leave_no_receipt`
- AC-9 (REQ-1, REQ-2, REQ-3, REQ-4, REQ-5, REQ-6, REQ-7, REQ-8, REQ-9): the existing `check affected` tests still pass.
  verify: `cargo test --locked --offline --test check_affected_cli`
- AC-10 (REQ-1, REQ-2, REQ-3, REQ-4, REQ-5, REQ-6, REQ-7, REQ-8, REQ-9): the SPEC passes the checker.
  verify: `moai spec check SPEC-20261121-check-receipts`

## Scope

- write: `.moai/specs/SPEC-20261121-check-receipts/**`, `src/check_affected.rs`, `src/ops/check_tree.rs` (the `--fresh` flag), `tests/check_receipt_cli.rs` (new), `tests/check_affected_cli.rs` (only if an existing test repeats a passing run on one tree)
- read-only: every other file, the real HOME
- tier 1: an independent audit of the exact head

## Plan

1. SPEC first (this file).
2. RED: `tests/check_receipt_cli.rs`, one test per acceptance line.
3. `check_affected.rs`: the key (tree id, planned steps, cargo and rustc versions), the receipt read and the atomic write, the reuse before the steps run.
4. `ops/check_tree.rs`: the `--fresh` flag and its help line.
5. The existing `check affected` tests, then the unnarrowed `check affected` set, then an independent audit (tier 1).

## Risks

- A stale receipt hides a failure: the key holds the tree id, so any change of a tracked file gives another key; untracked files and environment differences are outside the key (the run already ignores untracked files).
- Audit independence: a reused receipt is a machine fact about the same tree, the auditor decides whether to ask for `--fresh`; this SPEC does not change that rule.
- Receipt files accumulate in the state root: they are small; removal of old ones is a later change.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | the key holds the tree id, the planned steps and the cargo and rustc versions | the report names tree id and tool versions; the planned steps make a changed selection another key | add more words to the key |
| D-2 | the summary line changes only when a receipt is reused | the key order and the end `result=` of a normal run stay as consumers expect | always print the receipt word |
| D-3 | only a passing run leaves a receipt | a reused failure would hide a fix | none |

## Result

(written at SYNC)
