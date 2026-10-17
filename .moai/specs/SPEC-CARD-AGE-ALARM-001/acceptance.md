# Acceptance: SPEC-CARD-AGE-ALARM-001 — work-item-age alarms (card t1838, stage G1)

## Two-cell discipline

Every release-blocking criterion below carries a **RED-now cell** (`.claude/rules/moai/development/verification-completeness.md` §2): the command (single invocation, no pipes, no chaining), its verbatim stdout, its exit code as its own field, and the tree SHA the measurement was taken on. Criteria whose red form cannot be observed on the pre-implementation tree — because the code they constrain does not exist yet, making a red observation vacuous — are authored as **regression-guard** with the reason stated, per the same rule's undecidable disposition.

**Document-level pin:** every RED-now cell in this file was measured on tree `7faf75fdfbe27e442ea2c17a1a90caf5792d09e6` (branch `WT-10-17-g1`, worktree of card t1838), 2026-10-17 — except E-004 and E-006, whose criteria the plan-audit iteration 1 repairs touched (D1/D2 green-path additions; D3 retarget) and which were re-measured on tree `1cac5de4fa4c0f45942c8e2e037de1bd7f2fac2d` (the audited commit; criterion-level pins win where present). Commands are repo-root-relative. The probes are absence observations: each is red **for the right stated reason** — the symbol, key, or surface the criterion requires does not exist on this tree, and the milestone named in the green path is what creates it. None is a wrong-reason red (nothing pre-existing blocks it) and none is vacuous (the green path flips exactly this observation — for E-006 the flip is plan-entailed by the M3 marker pin, plan.md M3 (2)). The 0.1.2 rule-C correction (plan-audit iteration 2) touched AC-CAA-004's green path and re-measured its E-004 cell on tree `c14bb152a9e87900cbf9233e2e9029d24573c7a1` (the iter2-audited commit; the correction commit itself touches only SPEC artifacts, so code paths are identical).

**Load-rule note (10-17):** the RED cells are read-only greps precisely because heavy verification (builds, package-wide tests) was load-ruled out at authoring time; no RED cell below depends on a deferred heavy run.

## AC-CAA-001 — Enable flag exists, default OFF (REQ-CAA-001) — release-blocking

**RED-now** (`E-001`):

- Command: `grep -c "card_age_alarm" internal/config/defaults.go`
- Verbatim stdout: `0`
- Exit code: `1`
- Tree SHA: `7faf75fdfbe27e442ea2c17a1a90caf5792d09e6`

Red because: no `card_age_alarm` key exists anywhere in the config defaults on this tree; the config surface the requirement demands is absent.

**Green path:** M1 — `go test -run TestLoadCardAgeAlarm ./internal/config` observes `enabled=false` and the four compiled defaults from an unset workflow file; the same grep then counts ≥ 1. The byte-identical-when-off behavior itself is guarded by AC-CAA-007.

## AC-CAA-002 — Four threshold keys with defaults and template block, no hardcoded thresholds (REQ-CAA-002) — release-blocking

**RED-now** (`E-002a`, `E-002b`):

- Command: `grep -rln "CardAgeAlarm" internal/config/`
- Verbatim stdout: (empty)
- Exit code: `1`
- Tree SHA: `7faf75fdfbe27e442ea2c17a1a90caf5792d09e6`

- Command: `grep -c "card_age_alarm" internal/template/templates/.moai/config/sections/workflow.yaml`
- Verbatim stdout: `0`
- Exit code: `1`
- Tree SHA: `7faf75fdfbe27e442ea2c17a1a90caf5792d09e6`

Red because: neither the config type nor the shipped template block exists.

**Green path:** M1 — `go test -run TestCardAgeAlarmBounds ./internal/config` observes a non-positive duration and a warn-not-strictly-below-alert pair refused with the key path named, positive in-bound values accepted; both greps then count ≥ 1. M3's test (d) observes the degrade-not-refuse behavior on the status surface.

## AC-CAA-003 — Gate-wait age anchor from events, `updated_at` fallback (REQ-CAA-003) — release-blocking

**RED-now** (`E-003`):

- Command: `grep -rln "GateWaitAge" internal/homestate/`
- Verbatim stdout: (empty)
- Exit code: `1`
- Tree SHA: `7faf75fdfbe27e442ea2c17a1a90caf5792d09e6`

Red because: no gate-age derivation exists in the store package on this tree.

**Green path:** M2 — `go test -run TestGateWaitAge ./internal/homestate` observes (a) the T7 entry into `kickoff` anchoring on the entering event's `created_at` while a later `card.fields` write (which stamps `updated_at`, `internal/homestate/card_picked.go:197`) does not reset the age, (b) the mid-merge expiry into `blocked` anchoring on the `lease.expired` event, and (c) a no-events legacy row falling back to `updated_at`.

## AC-CAA-004 — Three signals computed and displayed separately; activity never counts as progress (REQ-CAA-004, REQ-CAA-005) — release-blocking

**RED-now** (`E-004`):

- Command: `grep -rln "CardAgeSignals" internal/homestate/`
- Verbatim stdout: (empty)
- Exit code: `1`
- Tree SHA: `c14bb152a9e87900cbf9233e2e9029d24573c7a1` (re-measured at the 0.1.2 rule-C correction; observation identical to the iter1 pin `1cac5de4f` and the authoring pin `7faf75fdf`)

Red because: the three-signal view does not exist; today no code distinguishes stage-checkpoint, commit, and heartbeat ages anywhere in the store package.

**Green path:** M2 — `go test -run TestCardAgeSignals ./internal/homestate` observes the three signal ages computed from their three distinct sources (events / git head commit of `WorktreePath` / `cards.heartbeat_at`), `unknown` rendered for an empty worktree path and an empty heartbeat, and the conflation mutant failing: a fresh heartbeat renewal does NOT advance the no-progress age. `go test -run TestNoProgressAge ./internal/homestate` observes the both-unknown path anchored on the card's creation event — never on the most recent transition (iter2 B1). `go test -run TestNoProgressAgeCreationAnchor ./internal/homestate` observes the iter2 creation-anchor pins: a pre-stage expiry → re-acquire crash loop preserves the no-progress age from creation, a fresh card whose worktree head is an old inherited base commit computes an age of ~0 from creation and raises no alarm, and a genuinely stuck pre-stage card still alerts, from creation. `go test -run TestAgeAnchorsFutureTimestamps ./internal/homestate` observes the iter2 validity reach: with future-dated sources beyond `CardAgeAlarmClockSkewTolerance`, the gate-wait anchor (REQ-CAA-003), both `updated_at` fallbacks, and the creation-event anchor all render `unknown` — no derived age is negative on any path. `go test -run TestStageCheckpointSameStageReEntry ./internal/homestate` observes the D1 stage-change rule: an expiry → re-acquire → T4 same-stage re-entry does not refresh the stage checkpoint or the no-progress age, while stage changes in both directions (a forward advance; a backward T6 `plan-audit` → `plan` after an audit FAIL) do. `go test -run TestFactoryStatusAgeAlarmFutureCommit ./internal/cli` (M3) observes the D2 validity rule: a future-dated head commit renders the commit signal `unknown`, excluded from the newest-of, with no alarm band suppressed.

## AC-CAA-005 — Additive `age_alarms` JSON block (REQ-CAA-006) — release-blocking

**RED-now** (`E-005`):

- Command: `grep -c "age_alarms" internal/cli/factory_card.go`
- Verbatim stdout: `0`
- Exit code: `1`
- Tree SHA: `7faf75fdfbe27e442ea2c17a1a90caf5792d09e6`

Red because: the JSON report type (`factoryStatusReport`, `internal/cli/factory_card.go:3065-3083`) carries no alarm key on this tree.

**Green path:** M3 — `go test -run TestFactoryStatusAgeAlarmJSON ./internal/cli` observes the `age_alarms` array present when enabled with an over-threshold seeded card, absent when disabled, and every pre-existing key of `factoryStatusReport` still serialized in its prior position (quota last).

## AC-CAA-006 — Banded, oldest-first text alarm section (REQ-CAA-006) — release-blocking

**RED-now** (`E-006`):

- Command: `grep -rln "age-alarm" internal/cli/`
- Verbatim stdout: (empty)
- Exit code: `1`
- Tree SHA: `1cac5de4fa4c0f45942c8e2e037de1bd7f2fac2d` (re-measured at the iter1 repair)

Red because: no file under `internal/cli/` carries the `age-alarm` token on this tree — the alarm renderer (`internal/cli/factory_age_alarm.go`, plan §C) does not exist yet, and the plan's three additive touches to `factory_card.go` (the `AgeAlarms` field, one compute call, one render call) carry no alarm string. The green flip is plan-entailed: plan.md M3 (2) pins the text alarm section's tag-prefixed lines to the literal token `age-alarm`, rendered by the new file (iter1 D3 — the prior grep pointed at `factory_card.go` for a token the plan never placed there, so a compliant M3 could have left it at zero).

**Green path:** M3 — `go test -run TestFactoryStatusAgeAlarmText ./internal/cli` observes, over a seeded DB with one warn-band and one alert-band card, one alarm line per card carrying card id, run, state, class, band, class age, and the three signal ages, sorted oldest-first, with no other output line changed.

## AC-CAA-007 — Byte-identical output while disabled (REQ-CAA-001) — regression-guard

Reason for guard classification: green-at-arrival — with the feature absent, today's output IS the pre-SPEC output, so no red observation exists to make on this tree (the vacuous direction of verification-completeness §2). Adopted on the green-path-preservation ground: the moment this SPEC's work accidentally changes the disabled-path output, the guard fails.

**Green path (binding from M3):** `go test -run TestFactoryStatusDisabledUnchanged ./internal/cli` compares the disabled-flag output against the pre-change golden over the same seeded card set and must remain byte-identical.

## AC-CAA-008 — Display-only: the alarm path writes nothing (REQ-CAA-007) — regression-guard

Reason for guard classification: the alarm code path does not exist at authoring time; a no-write assertion against absent code is vacuous and cannot be red today. It lands as an M3 rejection test whose mutant is explicit.

**Green path (binding from M3):** `go test -run TestAgeAlarmNoWrites ./internal/cli` observes, across one full alarm computation over a seeded DB, an unchanged events count and unchanged card rows/versions; a mutant that makes the alarm path append an event or bump a row must fail this test (verification-completeness §1.1 — the red must be observable, and here it is, from M3 onward).

## AC-CAA-009 — Coverage set: class a/b partition, terminal and legacy excluded (REQ-CAA-008) — release-blocking

**RED-now** (`E-009`):

- Command: `grep -rln "gateWaitingStates" internal/homestate/`
- Verbatim stdout: (empty)
- Exit code: `1`
- Tree SHA: `7faf75fdfbe27e442ea2c17a1a90caf5792d09e6`

Red because: no gate-waiting state set exists as a named symbol on this tree (the three gate states appear only as individual constants, `internal/homestate/card_record.go:23-37`).

**Green path:** M2 — `go test -run TestAgeAlarmCoverage ./internal/homestate` observes the partition: `kickoff`/`needs-decision`/`blocked` classify as gate-wait; `picked`, `assigned`, `leased`, the pipeline states, and the delivery states (`pr-open`, `merged-pr` included) classify as no-progress; `done`/`failed`/`abandoned` and `Card.Legacy()` rows classify as neither.

## Definition of Done

All release-blocking criteria (001-006, 009) flip green through their named milestone tests with verbatim output recorded in progress.md §E.2; both regression guards (007, 008) land green in M3 and stay in the tree; the alarm path demonstrably writes nothing on every observed run; and no automatic action of any kind exists behind the alarm (G2 boundary intact).
