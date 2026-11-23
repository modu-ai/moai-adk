---
id: SPEC-20261123-followup-adopt-help-matrix
title: "Follow-ups of two landing audits: the adopt SPEC's plan and AC-1 verify, and the support matrix's Codex status"
status: draft
revision: 1
created: 2026-11-23
---

# Follow-ups of two landing audits: the adopt SPEC's plan and AC-1 verify, and the support matrix's Codex status

## Intent

Two landing audits left findings that block nothing. The plan of `SPEC-20261123-lane-continues-card` names a file the code slice does not change, and its AC-1 verify command skips the test that covers REQ-1 in the adopt test file. The support matrix still lists Codex as a current host and names `codex` launcher forms after Codex support was withdrawn (user decision 2026-11-21). This SPEC changes documents only. The third audit finding, how the launcher help words `-l|-a`, is not decided here (see Open decisions).

## Requirements

- REQ-1: The Plan of `SPEC-20261123-lane-continues-card` shall name `src/ledger/adopt.rs` as the file of the adopt transaction.
- REQ-2: The AC-1 verify command of `SPEC-20261123-lane-continues-card` shall select every `ac1_` test of `session_registry_cli` and of `card_adopt_cli`.
- REQ-3: The Codex section of `docs/host-support-matrix.md` shall say before its table that Codex support is withdrawn and that the table is a record of 2026-11-08.
- REQ-4: The launcher section of `docs/host-support-matrix.md` shall name `cc` and `glm` as the only hosts the launcher starts.

## Acceptance

- AC-1 (REQ-1): the Plan step that names the adopt transaction names `src/ledger/adopt.rs` and no longer names `src/ledger/reservation.rs` as its place · verify: `rg -n "src/ledger/adopt.rs" .moai/specs/SPEC-20261123-lane-continues-card/spec.md`
- AC-2 (REQ-2): the AC-1 verify command of the adopt SPEC runs the `ac1_` tests of both files, including `ac1_a_session_that_comes_back_is_not_its_own_predecessor` · verify: `cargo test --locked --offline --test session_registry_cli --test card_adopt_cli ac1_`
- AC-3 (REQ-3): the matrix names the withdrawal before the Codex table · verify: `rg -n "철회됨\(사용자 결정 2026-11-21\)" docs/host-support-matrix.md`
- AC-4 (REQ-4): the matrix lists no `codex` launcher form · verify: `rg -n "cc\|glm\|codex" docs/host-support-matrix.md`

## Scope

- write: `.moai/specs/SPEC-20261123-lane-continues-card/spec.md`
- write: `.moai/specs/SPEC-20261123-followup-adopt-help-matrix/spec.md`
- write: `docs/host-support-matrix.md`
- not written: the launcher code and its help, the registry, the card commands, and the Codex table rows (a Python test pins their structure).

## Plan

1. `SPEC-20261123-lane-continues-card`: fix Plan step 4, the AC-1 verify command and the AC-1 Result line.
2. `docs/host-support-matrix.md`: the withdrawal paragraph before the Codex table; the launcher paragraph without `codex`.
3. Targeted tests, then one `moai check affected`, then READY with head, base and the receipt.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | The launcher help is not changed: it does not call `-l|-a` deprecated. | The recorded user decision of 2026-11-10 (Issue #2359 comment 6760976883) keeps `-l|-a` and replaces decision D-1 of 2026-11-09 (deprecate); the code prints no deprecation warning. The leader answered the lane's question with the same reading on 2026-11-23 (card note 7258). | Mark `-l|-a` deprecated in the help, after the user reverses the 2026-11-10 decision. |

Observed, not decided here: the help still says `-l|-a` "set its access" (`src/main.rs` `HostLaunch`, `Cc` and `Glm` doc comments, `src/session.rs` `role_entry_note`) while every launch now starts with the same access (SPEC-20261123-launcher-auto-mode). The leader left the help as it is.

## Result

Status: documents changed; the SPEC front matter is not changed here (approval is recorded by the user's pick, not by a lane).

- AC-1: met — `rg -n "src/ledger/adopt.rs" .moai/specs/SPEC-20261123-lane-continues-card/spec.md` printed the Scope line, the Plan step 4 and the Result line (exit 0); Plan step 4 no longer puts the adopt transaction in `src/ledger/reservation.rs`.
- AC-2: met — `cargo test --locked --offline --test session_registry_cli --test card_adopt_cli ac1_` exited 0: `card_adopt_cli` 1 passed (`ac1_a_session_that_comes_back_is_not_its_own_predecessor`) and `session_registry_cli` 3 passed.
- AC-3: met — `rg -n "철회됨\(사용자 결정 2026-11-21\)" docs/host-support-matrix.md` printed line 83, before the Codex table (exit 0).
- AC-4: met — `rg -n "cc\|glm\|codex" docs/host-support-matrix.md` printed nothing (exit 1, no match). `moai codex --dry-run -a` of the installed build prints `moai codex was removed: MoAI supports Claude Code and GLM only (moai cc | moai glm)`.
