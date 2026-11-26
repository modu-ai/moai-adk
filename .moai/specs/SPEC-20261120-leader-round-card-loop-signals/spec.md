---
id: SPEC-20261120-leader-round-card-loop-signals
title: "the leader round and lane lines follow the card loop: ready markers wait, long foreground is reported, waits are dispatch candidates, compact is answered"
status: draft
revision: 3
created: 2026-11-20
---

# the leader round and lane lines follow the card loop: ready markers wait, long foreground is reported, waits are dispatch candidates, compact is answered

## Intent

The leader round still speaks of open pull requests, a pull-request cap and `LANE-REPORT`, in fixed Korean, and it does not know the markers of the card loop (card, audit, merge-ready, next card). A lane that has posted its ready line is judged `continue` and gets the restart line typed into it while it waits for its audit or its merge. A command that runs in the foreground for a long time, with the screen unchanged, is read as a turn that runs and is left forever. A lane that waits on a file or an audit is not offered to the leader as a lane that can take another card, and a lane that asks for a compact gets nothing. This change teaches the pure part of the round (`src/leader_round.rs`) those four judgments, and puts the loop lines (the duty line, the restart line, the end-condition text and the resume line of `src/lane_duty.rs` and `src/lane_loop.rs`) in the user's language and in the words of the card loop. The code waits for the card that renames the mode flags, which holds the same files; this SPEC and the recorded screens it will be tested with come first.

## Requirements

- REQ-1: The leader round shall count a line that starts with `LANE-READY-INTEGRATION`, `MERGE-READY` or `MERGED-LOCAL` as a lane marker, found like the other markers (start of a line, one host bullet allowed, not quoted, not inside a fence), accepting a line with no lane word and listing a line that carries `lane=<other>` under `ignored`.
- REQ-2: WHEN the last marker of a lane is `LANE-READY-INTEGRATION` or `MERGE-READY`, the leader round shall leave the lane with the reason `waiting_for_merge` and send it no line. · failure: none
- REQ-3: WHEN the last marker of a lane is `MERGED-LOCAL` or `LANE-DONE` and no claim follows it, the leader round shall treat it as a card boundary and answer the action `clear` with the lines `/clear` and the duty line of the lane, whatever the lane's host and context percentage. · failure: none
- REQ-4: WHEN the last marker of a lane is `LANE-READY-INTEGRATION`, `MERGE-READY` or a `LANE-WAIT` with a word `on=<text>`, the leader round shall name the lane in the round's `waiting_lanes` and give the action a `waiting` object with `on` (`merge` for the ready lines) and, when the line has `until <time>`, `until`.
- REQ-5: The observation of a lane shall take an optional `foreground` object with `same_tail_since` (a time) and `shell_running` (a boolean), written by the observer who keeps the earlier reads, besides the `context_used_pct` it already takes.
- REQ-6: WHEN the turn of a lane runs, `foreground.shell_running` is true and `same_tail_since` is at least 600 seconds before `read_at`, the leader round shall answer the action `report` with the reason `long_foreground`, send nothing, and name the lane in `long_foreground_lanes`.
- REQ-7: WHEN the turn of a lane is idle, no card boundary is its last marker, and the last marker is `LANE-COMPACT` or `context_used_pct` is at least 80, the leader round shall answer the action `compact` with, for a `claude` lane, the one line `/compact` followed by the duty line of the lane, and, for a `codex` lane, the line `/compact` and then the duty line as a second line.
- REQ-8 (REQ-7): IF the host of the lane that needs a compact is neither `claude` nor `codex`, THEN the leader round shall leave the lane with the reason `compact_not_supported`.
- REQ-9: The loop lines shall follow the card loop of taking a card then audit then merge-ready then the next card, naming no open pull requests, no pull-request cap and no `LANE-REPORT`, and written in the user's language, English or Korean, chosen the way the other human screens choose it.
- REQ-10 (REQ-6): IF a `foreground` object has a time that does not parse or a key the observation file does not know, THEN the leader round shall refuse the observation file with the parse error.
- REQ-11 (REQ-4): IF a `LANE-WAIT` line has no `on=<text>` word, THEN the leader round shall leave the lane with the reason `lane_wait` as before and leave it out of `waiting_lanes`.
- REQ-12: The workflow texts of the two modes shall tell the lane to print `LANE-DONE` after its `MERGE-READY` or `READY` note and expect `/clear` before the next card, to print `LANE-COMPACT` mid-card when its host shows the context at 80 percent or more, and tell the leader round to send `/clear` with the next card to a lane that printed `LANE-DONE` and `/compact` to a lane at 80 percent or more.

## Acceptance

- AC-1 (REQ-1, REQ-2, REQ-4): the screens `ready-integration`, `merge-ready`, `quoted-and-fenced-merge-ready` and `merge-ready-other-lane` get the answers of `tests/fixtures/leader_round/screens/expected.json` · verify: `cargo test --locked --offline --test leader_round_signals ready`
- AC-2 (REQ-3): the screens `merged-local` and the existing `lane_done` screen clear and redispatch with the duty line, for a Claude lane and a Codex lane · verify: `cargo test --locked --offline --test leader_round_signals merged_local`
- AC-3 (REQ-4, REQ-11): the screens `lane-wait-on` and `lane-wait-expired` get the answers of `expected.json`, the first with its `on` and `until`, and a wait line without `on=` is not listed · verify: `cargo test --locked --offline --test leader_round_signals wait`
- AC-4 (REQ-5, REQ-6, REQ-10): the screens `long-foreground-599`, `long-foreground-600` and `long-foreground-no-shell` get the answers of `expected.json` (599 seconds leaves, 600 seconds reports), and a broken `foreground` is refused · verify: `cargo test --locked --offline --test leader_round_signals foreground`
- AC-5 (REQ-7, REQ-8): the screens `lane-compact`, `lane-compact-codex`, `lane-compact-omp`, `context-80-claude`, `context-79-claude` and `context-80-running` get the answers of `expected.json` (a Claude lane gets `/compact` with its duty line, a Codex lane the two lines, an OMP lane the reason, 79 percent and a running turn get nothing) · verify: `cargo test --locked --offline --test leader_round_signals compact`
- AC-6 (REQ-9): the loop lines name no pull-request cap and no `LANE-REPORT`, and come in both languages · verify: `cargo test --locked --offline --test leader_round_signals wording --test lane_duty_line --test lane_loop_cli`
- AC-7 (REQ-1…REQ-11): every screen of the folder is read by one test that checks `expected.json` against the answer, so a screen without an expectation or an expectation without a screen fails · verify: `cargo test --locked --offline --test leader_round_signals every_screen`
- AC-8 (REQ-12): the texts of the two modes carry the three lane lines and the two leader lines, once each, and the scan and reader tests of those files pass · verify: `cargo test --locked --offline --test leader_round_signals mode_text`

## Scope

- write: `.moai/specs/SPEC-20261120-leader-round-card-loop-signals/spec.md`
- write: `tests/fixtures/leader_round/screens/**`
- write: `src/leader_round.rs`
- write: `src/lane_duty.rs`
- write: `src/lane_loop.rs`
- write: `tests/leader_round_signals.rs`
- write: `tests/lane_duty_line.rs`
- write: `tests/leader_round_cli.rs`
- write (when the files are free, after the cards that hold them have merged): `skills/moai/workflows/mode-agent.md`, `skills/moai/workflows/mode-leader.md`, `skills/moai/references/source-manifest.json`
- write: `.moai/specs/SPEC-20261120-leader-round-card-loop-signals/research-context.md`, `tests/fixtures/leader_round/screens/**`

## Decisions

- D-1: The markers of REQ-1 carry the card, not the lane, as their second word (`MERGE-READY <card> …`); the screen the line stands on belongs to one lane, so the lane check accepts a missing lane word and refuses only `lane=<other>`. A line of another lane that is quoted or fenced is already out. Reversible: a later change may add `lane=<lane>` to the lines the lane prints.
- D-2: A long foreground command is read from an observation key written by the observer, not guessed from one screen: the round has one read per lane and the principle of the module is that nothing is guessed from the screen. The limit is 600 seconds measured to `read_at`; the observer starts `same_tail_since` when the tail first equals the earlier read.
- D-3 (changed after reading both hosts' documents, `research-context.md`): a compact is answered for a Claude lane and a Codex lane. Claude Code takes the next step as the argument (`/compact [instructions]`, documented); Codex `/compact` takes no argument and does not run during a running turn (source), so its next step is a second line and only an idle lane gets it. OMP is not measured and keeps `compact_not_supported`.
- D-4: The language of the lines is chosen with the `Lang` of the human screens (`ui_text`), the same way every other human text of the tool is; the lines the lane types into itself stay one language per run.
- D-5 (open, for the leader to confirm): The end condition is computed from a forge snapshot (open pull requests, startable Issues). Its wording can follow the card loop now, but a card-ledger end condition (no card ready, no lane waiting for audit or merge) is a computation, not a wording. Default: this SPEC changes the wording only where it stays true to what is computed, and the computation moves in a card of its own.
- D-6: The code files of the Scope are held by the card that renames the mode flags until it has landed; this card holds only the SPEC and the screens until then, and widens its scope when the files are returned.
- D-7: `/clear` is documented by both hosts under that name (Codex also has `/new`), so one line serves both and the round already types `/clear` and the duty line at a card boundary. This change fixes that for every lane host and for every context percentage: a card boundary is a clear, a compact belongs to the middle of a card. The old rule that a `LANE-REPORT` at 85 percent clears is replaced by the compact of REQ-7.
- D-8: The percentage is read from the host. Claude Code documents one program-readable value, `context_window.used_percentage` of the status line input (input tokens only, `null` early in a session); the observer writes it into `context_used_pct` and a `null` stays below the limit, as today. Codex documents no value a script can read (the footer and `/status` are for people), so a Codex lane prints `LANE-COMPACT` itself when its footer shows 80 percent or more, and the observer may fill `context_used_pct` from the footer when it can read it. The limit is a constant of 80 beside the existing `CLEAR_CONTEXT_PCT`.
- D-9: After `/clear` the session id of a lane may change (Codex starts a new thread in the source; the Claude Code documents are silent). A lane's reservation therefore must not depend on the session id; that continuity is the work of the card that carries reservations over a clear, not of this one. Until it has landed, a `clear` answer is only sent to a lane whose reservation the leader has already released or handed on.

## Plan

1. This card, now: the SPEC, `research-context.md` and sixteen screens (in this branch; no code).
2. After the rename card has landed and returned the files: `tests/leader_round_signals.rs` reading every screen against `expected.json` (red first), then the markers, the `waiting_for_merge` and boundary judgments, `waiting_lanes`, `foreground`, `compact`, then the wording and the two languages, then the tests that pin the old texts (`lane_duty_line`, `lane_loop_cli`, `leader_round_cli`).
3. Run the whole area, clippy, rustfmt, spec check, audit, merge-ready.

## Result

Not yet written: the code is blocked on the rename card (same files). Recorded here: the SPEC and the sixteen screens (`tests/fixtures/leader_round/screens/synthetic-*.json`) are in this branch. `tests/fixtures/leader_round/screens/expected.json`, which the criteria name, is not in this branch yet: it is written together with `tests/leader_round_signals.rs` in step 2 of the plan, and an earlier commit body that said "expected.json for every screen" was wrong about this head.
