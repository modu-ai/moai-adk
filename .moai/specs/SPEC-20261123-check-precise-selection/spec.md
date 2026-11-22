---
id: SPEC-20261123-check-precise-selection
title: "check affected: a precise lane selection, the full set once per train candidate, and the tests that pin skill and doc text"
status: draft
revision: 1
created: 2026-11-23
---

# check affected: a precise lane selection, the full set once per train candidate, and the tests that pin skill and doc text

## Intent

`check affected` picks one set for every caller: every target the scan list names, every consumer
of `tests/common`, every OS-tagged test. A lane's READY therefore carries the weight of a merge
train candidate run. The user decided on 2026-11-21 (test-load redesign
`~/.moai/reports/moai-adk/test-load-redesign-2026-11-21.md` §S3): the lane runs a precise
selection, the train runs the full set once per candidate, the answer says which one ran, and a
MERGE-READY accepts only the full run's PASS. Two follow-ups were added to this card later: a
change to skill or doc text selects the tests that read that text (leader note 5115, the C-218 and
C-246 misses), and a range whose base lies behind the merge base with the integration tip is
reported, because it re-tests the tip's own changes (leader note 5395, the 343-target run).

What is kept from the earlier, unlanded branch `spec/check-two-modes-20261121` (12376225e5,
`SPEC-20261121-check-two-modes`): its three lane rules (scan scope, `tests/common` sub-modules, OS
area), the full set as the default with `--lane` choosing the precise set (its D-5, so every
existing caller and test keeps today's set), the `selection` answer field, and its fixture cases.
What changes: its tests called the library (`select_mode`, `Mode`), which the one-slot card is
reshaping now; these fixtures drive the command (`--list`) in temporary repositories instead, so
they hold whatever the library looks like after that card lands.

## Non-goals

- The per-item map of `tests/common/mod.rs` and a sub-command map of `src/main.rs` (report §S3b);
  a lane change to `tests/common/mod.rs` keeps selecting every consumer.
- Reusing a receipt by tree hash across a tip move (leader D-b; another card).
- The workflow guidance of the lane and leader modes: the cards that hold those files update them;
  this SPEC writes the command's operations page only.
- The ff gate: it keeps requiring `result=PASS` of a full run.

## Requirements

- REQ-1: The bare command shall select today's full set, so the full set contains every set `--lane` selects for the same range.
- REQ-2: When `--lane` is given, a scan of `tools/repo_scans.json` shall add its targets only when a changed path matches that scan's `scope` or is one of its `rules` files. · failure: none — a scan without a readable scope is left out of the lane and kept in the full set
- REQ-3: When `--lane` is given and only `tests/common/<sub>.rs` changed in `tests/common`, the selection shall hold the consumers whose source names `common::<sub>` and no other consumer. · failure: none — a change to `tests/common/mod.rs` keeps every consumer
- REQ-4: When `--lane` is given, an OS-tagged test shall be selected only when its name's first word names the area of a changed path. · failure: none — the full set keeps every OS-tagged test of the host
- REQ-5: The answers of `--list` and of a run shall name the selection in a `selection` field that reads `lane` or `full`.
- REQ-6: When a changed path under `skills/` or `docs/` appears literally in a Rust test file, the selection shall include that test in both selections. · failure: none — a test file that cannot be read adds nothing by this rule
- REQ-7: When the base of the range is a strict ancestor of the merge base of the head and the ref `--integration-ref` names, the answer shall carry a `range_note` that names that merge base and the count of extra changed paths. · failure: none — an unreadable ref gives no note and the selection is unchanged
- REQ-8: The operations page of the command shall describe the two selections, the `selection` field, the text-pin rule and the range note.

## Acceptance

- AC-1 (REQ-1, REQ-2): a scan whose scope misses the diff is left out with `--lane` and kept without it, a scan without a scope is left out with `--lane`, a scan whose rule file changed is kept with `--lane`, and the full set contains the lane set · verify: `cargo test --locked --offline --test check_affected_modes_cli scan`
- AC-2 (REQ-3): a change to `tests/common/home.rs` selects the `home` consumer and not the `cli` consumer with `--lane`, and both without it · verify: `cargo test --locked --offline --test check_affected_modes_cli common`
- AC-3 (REQ-4): with `--lane` a host-OS test whose first word misses the diff areas is left out and one whose first word matches is kept · verify: `cargo test --locked --offline --test check_affected_modes_cli os`
- AC-4 (REQ-5): `selection` reads `lane` with `--lane` and `full` without it, in `--list` · verify: `cargo test --locked --offline --test check_affected_modes_cli selection`
- AC-5 (REQ-6): a change to `skills/moai/workflows/x.md` or `docs/a.md` selects the test whose source names that path, with and without `--lane` · verify: `cargo test --locked --offline --test check_affected_modes_cli text`
- AC-6 (REQ-7): a range from an old fork point after the tip was merged in carries `range_note` with the merge base and the extra count, and a range from the merge base carries none · verify: `cargo test --locked --offline --test check_affected_modes_cli range`
- AC-7 (REQ-8): the operations page names `--lane`, `selection`, the text-pin rule and `range_note` · verify: `cargo test --locked --offline --test check_affected_modes_cli page`
- AC-8: the existing tests of the command pass unchanged · verify: `cargo test --locked --offline --test check_affected_cli`

## Scope

- write now: this SPEC folder, the new `tests/check_affected_modes_cli.rs`
- write after the one-slot card lands and gives them back: `src/check_affected.rs`, `src/ops/check.rs`, `docs/design-review/appendix/operations/check.md`
- read-only: `tools/repo_scans.json`, `tests/check_affected_cli.rs` (held by the one-slot card), `tests/common/**`
- forbidden: card numbers in the SPEC id, the branch, commits and file text; the workflow guidance files (held by other cards)

## Plan

1. S1 fixtures (now, leader decision 6896): `tests/check_affected_modes_cli.rs` builds a temporary repository per case and runs `check affected --list` with and without `--lane`; every case fails on the tip by assertion or by the refused `--lane` flag.
2. S2 code (after the one-slot card lands): the selection takes a lane flag; the scan scope, common sub-module and OS area rules apply to the lane; the text-pin rule applies to both; `--lane` on the command; `selection` and `range_note` in the answers.
3. S3: the operations page; the measured lane and full target counts on this repository's last ten READY ranges go into the Result.

## Risks

- A scan scoped to all of `src/` still joins most lane runs; the Result records the measured drop.
- A test whose name does not say its area is missed by a lane run on another OS; the full train run catches it, and the MERGE-READY rule keeps that run the gate.
- The text-pin rule reads test sources for literal paths; a test that builds the path at run time is not seen (the doc-reference scan does the same today).

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | The bare command keeps the full set; `--lane` picks the precise set | earlier branch D-5: every existing caller and test keeps its set, and the MERGE-READY rule reads the bare run | flip the default and add `--train` |
| D-2 | The fixtures drive the command, not the library | the one-slot card is reshaping `src/check_affected.rs`; the command's `--list` answer is the stable surface (leader decision 6896: fixtures first, no src write) | call the library once it settles |
| D-3 | The field is `selection: lane|full` beside the existing `mode: list|run` | `mode` already names list or run in the answer | rename both together |
| D-4 | The range note warns and never refuses | a refusal would stop the merge-candidate runs that need the wide range on purpose; the note makes the cost visible (planning default) | refuse unless a flag allows it |
