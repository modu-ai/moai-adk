---
id: SPEC-20261120-leader-round-card-loop-signals
title: "the leader round and lane lines follow the card loop: ready markers wait, long foreground is reported, waits are dispatch candidates, compact is answered"
status: draft
revision: 2
created: 2026-11-20
---

# the leader round and lane lines follow the card loop: ready markers wait, long foreground is reported, waits are dispatch candidates, compact is answered

## Intent

The leader round still speaks of open pull requests, a pull-request cap and `LANE-REPORT`, in fixed Korean, and it does not know the markers of the card loop (card, audit, merge-ready, next card). A lane that has posted its ready line is judged `continue` and gets the restart line typed into it while it waits for its audit or its merge. A command that runs in the foreground for a long time, with the screen unchanged, is read as a turn that runs and is left forever. A lane that waits on a file or an audit is not offered to the leader as a lane that can take another card, and a lane that asks for a compact gets nothing. This change teaches the pure part of the round (`src/leader_round.rs`) those four judgments, and puts the loop lines (the duty line, the restart line, the end-condition text and the resume line of `src/lane_duty.rs` and `src/lane_loop.rs`) in the user's language and in the words of the card loop. The code waits for the card that renames the mode flags, which holds the same files; this SPEC and the recorded screens it will be tested with come first.

## Requirements

- REQ-1: The leader round shall count a line that starts with `LANE-READY-INTEGRATION`, `MERGE-READY` or `MERGED-LOCAL` as a lane marker, found like the other markers (start of a line, one host bullet allowed, not quoted, not inside a fence), accepting a line with no lane word and listing a line that carries `lane=<other>` under `ignored`.
- REQ-2: WHEN the last marker of a lane is `LANE-READY-INTEGRATION` or `MERGE-READY`, the leader round shall leave the lane with the reason `waiting_for_merge` and send it no line. · failure: none
- REQ-3: WHEN the last marker of a lane is `MERGED-LOCAL` and no claim follows it, the leader round shall treat it as a card boundary, like `LANE-DONE`. · failure: none
- REQ-4: WHEN the last marker of a lane is `LANE-READY-INTEGRATION`, `MERGE-READY` or a `LANE-WAIT` with a word `on=<text>`, the leader round shall name the lane in the round's `waiting_lanes` and give the action a `waiting` object with `on` (`merge` for the ready lines) and, when the line has `until <time>`, `until`.
- REQ-5: The observation of a lane shall take an optional `foreground` object with `same_tail_since` (a time) and `shell_running` (a boolean), written by the observer who keeps the earlier reads.
- REQ-6: WHEN the turn of a lane runs, `foreground.shell_running` is true and `same_tail_since` is at least 600 seconds before `read_at`, the leader round shall answer the action `report` with the reason `long_foreground`, send nothing, and name the lane in `long_foreground_lanes`.
- REQ-7: WHEN the last marker of a lane is `LANE-COMPACT` and the host of the lane is `claude`, the leader round shall answer the action `compact` with the lines `/compact` and the duty line of the lane, in that order.
- REQ-8 (REQ-7): IF the host of the lane that printed `LANE-COMPACT` is not `claude`, THEN the leader round shall leave the lane with the reason `compact_not_supported`.
- REQ-9: The loop lines shall follow the card loop of taking a card then audit then merge-ready then the next card, naming no open pull requests, no pull-request cap and no `LANE-REPORT`.
- REQ-10: The loop lines shall be written in the user's language, English or Korean, chosen the way the other human screens choose it.
- REQ-11 (REQ-6): IF a `foreground` object has a time that does not parse or a key the observation file does not know, THEN the leader round shall refuse the observation file with the parse error.
- REQ-12 (REQ-4): IF a `LANE-WAIT` line has no `on=<text>` word, THEN the leader round shall leave the lane with the reason `lane_wait` as before and leave it out of `waiting_lanes`.

## Acceptance

- AC-1 (REQ-1, REQ-2, REQ-4): the screens `ready-integration`, `merge-ready`, `quoted-and-fenced-merge-ready` and `merge-ready-other-lane` get the answers of `tests/fixtures/leader_round/screens/expected.json` · verify: `cargo test --locked --offline --test leader_round_signals ready`
- AC-2 (REQ-3): the screen `merged-local` clears and redispatches with the duty line · verify: `cargo test --locked --offline --test leader_round_signals merged_local`
- AC-3 (REQ-4, REQ-12): the screens `lane-wait-on` and `lane-wait-expired` get the answers of `expected.json`, the first with its `on` and `until`, and a wait line without `on=` is not listed · verify: `cargo test --locked --offline --test leader_round_signals wait`
- AC-4 (REQ-5, REQ-6, REQ-11): the screens `long-foreground-599`, `long-foreground-600` and `long-foreground-no-shell` get the answers of `expected.json` (599 seconds leaves, 600 seconds reports), and a broken `foreground` is refused · verify: `cargo test --locked --offline --test leader_round_signals foreground`
- AC-5 (REQ-7, REQ-8): the screens `lane-compact` and `lane-compact-omp` get the answers of `expected.json` · verify: `cargo test --locked --offline --test leader_round_signals compact`
- AC-6 (REQ-9, REQ-10): the loop lines name no pull-request cap and no `LANE-REPORT`, and come in both languages · verify: `cargo test --locked --offline --test leader_round_signals wording --test lane_duty_line --test lane_loop_cli`
- AC-7 (REQ-1…REQ-12): every screen of the folder is read by one test that checks `expected.json` against the answer, so a screen without an expectation or an expectation without a screen fails · verify: `cargo test --locked --offline --test leader_round_signals every_screen`

## Scope

- write: `.moai/specs/SPEC-20261120-leader-round-card-loop-signals/spec.md`
- write: `tests/fixtures/leader_round/screens/**`
- write: `src/leader_round.rs`
- write: `src/lane_duty.rs`
- write: `src/lane_loop.rs`
- write: `tests/leader_round_signals.rs`
- write: `tests/lane_duty_line.rs`
- write: `tests/leader_round_cli.rs`

## Decisions

- D-1: The markers of REQ-1 carry the card, not the lane, as their second word (`MERGE-READY <card> …`); the screen the line stands on belongs to one lane, so the lane check accepts a missing lane word and refuses only `lane=<other>`. A line of another lane that is quoted or fenced is already out. Reversible: a later change may add `lane=<lane>` to the lines the lane prints.
- D-2: A long foreground command is read from an observation key written by the observer, not guessed from one screen: the round has one read per lane and the principle of the module is that nothing is guessed from the screen. The limit is 600 seconds measured to `read_at`; the observer starts `same_tail_since` when the tail first equals the earlier read.
- D-3: `/compact` is sent only to a Claude lane. Whether Codex and OMP lanes accept the same line is not measured here; they are left with a reason until a card measures it (an unverified line typed into a lane is worse than no line).
- D-4: The language of the lines is chosen with the `Lang` of the human screens (`ui_text`), the same way every other human text of the tool is; the lines the lane types into itself stay one language per run.
- D-5 (open, for the leader to confirm): The end condition is computed from a forge snapshot (open pull requests, startable Issues). Its wording can follow the card loop now, but a card-ledger end condition (no card ready, no lane waiting for audit or merge) is a computation, not a wording. Default: this SPEC changes the wording only where it stays true to what is computed, and the computation moves in a card of its own.
- D-6: The code files of the Scope are held by the card that renames the mode flags until it has landed; this card holds only the SPEC and the screens until then, and widens its scope when the files are returned.

## Plan

1. This card, now: the SPEC and the twelve screens with `expected.json` and the README (in this branch; no code).
2. After the rename card has landed and returned the files: `tests/leader_round_signals.rs` reading every screen against `expected.json` (red first), then the markers, the `waiting_for_merge` and boundary judgments, `waiting_lanes`, `foreground`, `compact`, then the wording and the two languages, then the tests that pin the old texts (`lane_duty_line`, `lane_loop_cli`, `leader_round_cli`).
3. Run the whole area, clippy, rustfmt, spec check, audit, merge-ready.

## Result

Not yet written: the code is blocked on the rename card (same files). Recorded here: the SPEC and the twelve screens are in this branch.
