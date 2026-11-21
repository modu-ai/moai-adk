---
id: SPEC-20261121-leader-run-lease
title: "one supervising leader run per project: a lease that a second tick, a restart and a manual run respect, for a Claude Code and a Codex CLI leader"
status: draft
revision: 1
created: 2026-11-21
parent: SPEC-20261120-leader-round-card-loop-signals
---

# one supervising leader run per project: a lease that a second tick, a restart and a manual run respect, for a Claude Code and a Codex CLI leader

## Intent

The leader supervises a project by rounds that repeat about every five minutes. A Claude Code leader repeats them with its own timer; a Codex CLI leader, which has none, types the round instruction into its own terminal (`leader-pulse`). Both commands that serve a round, `leader-round` and `leader-pulse`, only read or only print, so nothing stops two ticks from overlapping, or a manual run and a restart from starting a second leader. The duplicate dispatch and the duplicate merge are already refused (the card ledger's generation fence; the merge lock with fast-forward only), so the missing piece is the guard on the run itself.

Orca's automations page (https://orca.dev/docs/cli/automations, read 2026-11-21T01:31:19Z) does not supply it: `--reuse-session` only continues in the previous live automation terminal, and the page gives no lock, skip-if-running or dedupe flag and says nothing about a manual run that overlaps a running one, a restart or a sleeping machine. This change adds one lease per project, taken at the start of a tick and released at its end; a second tick skips, an ended lease is taken over (a restart), a project with nothing ready or under a HOLD takes no lease and says what wakes it. It touches no automation setting, approval or permission, does not widen supervision to every lane of the product's own repository, and keeps the local integration branch as the base and the independent audit.

## Requirements

- REQ-1: The `leader-run begin` command shall take the supervision lease of the project for the named session and print `taken` when no unexpired lease is held and at least one card is ready.
- REQ-2 (REQ-1): IF an unexpired lease of another session is held, THEN the `leader-run begin` command shall print `skipped` with the holder, the age and the lease end, and change nothing.
- REQ-3 (REQ-1): IF the lease has ended, THEN the `leader-run begin` command shall take it over and say that it replaced the ended lease of the earlier holder.
- REQ-4: The `leader-run end` command shall release the lease only for the session that holds it and refuse the release of another session's lease.
- REQ-5 (REQ-1): IF the caller reports no ready card or a HOLD, THEN the `leader-run begin` command shall take no lease and print `idle` with the state change that wakes the run.
- REQ-6: The MODE-LEADER text shall tell a Claude Code or Codex CLI leader to call `leader-run begin` first in each tick, to run the round only on `taken`, to skip it on `skipped` and `idle`, and to call `leader-run end` last.

## Acceptance

- AC-1 (REQ-1, REQ-2, REQ-3, REQ-4, REQ-5): in a temporary project, two overlapping begins give one `taken` and one `skipped`, an ended lease is taken over, a begin that reports no ready card gives `idle` and the next one that reports a ready card gives `taken`, and an end by another session is refused; the times and results of the run are written in the Result · verify: `cargo test --locked --offline --test leader_run_cli`
- AC-2 (REQ-6): the MODE-LEADER text carries the lease rule once and the scan and reader tests of the file pass · verify: `cargo test --locked --offline --test leader_run_cli mode_text`

## Scope

- write: `.moai/specs/SPEC-20261121-leader-run-lease/**`
- write: `src/leader_run.rs`, `src/ops/leader_run.rs`, `tests/leader_run_cli.rs` (new files; the operation registry file when it is free)
- write (when the file is free): `skills/moai/workflows/mode-leader.md`, `skills/moai/references/source-manifest.json`

## Decisions

- D-1 (open, for the leader to confirm; note on the card): the lease is a file in the project state folder written by a new `leader-run` command (begin, end, status), the same for a Claude Code and a Codex CLI leader. Other places were offered: a ledger row (a schema change, one place for all state) and folding the lease into `leader-pulse` (which today never reads state). Default: the lease file.
- D-2: Duplicate dispatch stays refused by the card ledger's generation fence and duplicate merge by the merge lock with fast-forward only; this change adds no second guard for them and tests only the run lease.
- D-3: A lease ends by time, not by a probe of the holder: a leader that stopped without `end` is taken over after the lease end. The length is a setting with a default longer than one tick; the tests give it explicitly.
- D-4: The checks run in a temporary project with a temporary HOME; no real HOME, Orca setting, approval or permission is read or written, and times come from a clock read in the same step.

## Plan

1. This card, now: this SPEC (default D-1) and the leader's answer.
2. Tests first (`tests/leader_run_cli.rs`, red), then `src/leader_run.rs` and the operation, then the MODE-LEADER text when its file is free, then the manifest.
3. Whole-area run, clippy, rustfmt, spec check, audit, merge-ready.

## Result

- AC-1: not yet — nothing written.
- AC-2: not yet — nothing written.
