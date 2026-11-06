# Plan: SPEC-CARD-AGE-ALARM-001 — work-item-age alarms (card t1838, stage G1)

Tier M, Class C, display-only. This plan carries the run-phase milestones. Every milestone lands its rejection test before its implementation (design §6 stage discipline; `.claude/rules/moai/development/verification-completeness.md` §1.1 — a check whose red has never been seen has proven nothing).

## A. Context

- Work tree: this card's worktree, branch `WT-10-17-g1`, plan-phase base `7faf75fdfbe27e442ea2c17a1a90caf5792d09e6`.
- SPEC artifacts: `.moai/specs/SPEC-CARD-AGE-ALARM-001/{spec,plan,acceptance,decision-index}.md` (this file set).
- Normative design input: `.moai/reports/card-policy-redesign-20261017/design.md` (primary checkout; not committed in this branch's tree — see decision-index Q2 note on anchors).
- Sibling overlap: card t1839 (lane-17) lands eligibility-filtered selection in `internal/cli/factory_card.go` (the `factory next` region, `:973-1017` per the design audit). This plan's edits to that file are three additive touches in the status region (`:3065-3215`), far from the selection region, and all new logic lives in new files — see §F.

## B. Measured substrate (what exists at `7faf75fdf`)

| Concern | Source (file:line) |
|---|---|
| Card states, gate states, terminal states | `internal/homestate/card_record.go:17-37`, `:59-64`, `:74` |
| Lease-holding states (heartbeat written only here) | `internal/homestate/card_record.go:66-68` |
| `updated_at` / `heartbeat_at` columns | `internal/homestate/factory.go:52`, `:56`; struct `card_record.go:107`, `:111` |
| Events table (kind, payload, created_at) | `internal/homestate/factory.go:79-85` |
| Event append (every transition; creation `from:""`) | `internal/homestate/card_transition.go:901-908`; `card_picked.go:173` |
| Field-only writes stamp `updated_at`, append `card.fields`, no transition | `internal/homestate/card_picked.go:197` |
| Lazy expiry → blocked carries payload `to` on `lease.expired` | `internal/homestate/card_transition.go:422-439` (`:426-430`, `:434-438`) |
| Gate-entry writes (T21/T7/T9/T23) | `internal/homestate/card_transition.go:789-797`, `:636-640` |
| Status command / text renderer / JSON report | `internal/cli/factory_card.go:3106-3178` / `:3180-3215` / `:3065-3083` |
| In-command degrade precedent (quota cells) | `internal/cli/factory_card.go:3131-3136` |
| Config key pattern (struct/loader/defaults/template/bounds) | `internal/config/types.go:435-450`; `internal/config/loader_factory_lease.go:24`; `internal/config/defaults.go:61-70`, `:1454-1456`; `internal/homestate/factory_lease.go:36-60`; template `workflow.yaml:21-29` |
| Delivery states | `internal/homestate/card_pr_states.go:36-38` |
| T4 stage-resume edges (leased → the kept stage only) | `internal/homestate/card_transition.go:137-142`, `:811-818`; expiry keeps stage `:418-421` |
| Clock-skew tolerance precedent (future timestamp → unknown) | `internal/config/defaults.go:82-87` (`QuotaClockSkewTolerance`) |

No commit timestamp exists in the store (`evidence_sha`/`merge_sha` are single stage-boundary SHAs); no hook-activity record exists beyond `heartbeat_at`; no `created_at` column exists. spec.md §Background names all three gaps and what each requirement reads instead.

## C. New symbols (named here so tests and review share one vocabulary)

- `internal/config`: `CardAgeAlarmConfig` (struct + `yaml:"card_age_alarm"` field on the workflow config), `LoadCardAgeAlarm(projectRoot)`, constants `DefaultCardAgeAlarmEnabled` (false), `DefaultCardAgeAlarmGateWarnAfter`, `DefaultCardAgeAlarmGateAlertAfter`, `DefaultCardAgeAlarmStallWarnAfter`, `DefaultCardAgeAlarmStallAlertAfter`, and the compiled skew tolerance `CardAgeAlarmClockSkewTolerance` (iter1 D2 — mirrors `QuotaClockSkewTolerance`, `internal/config/defaults.go:82-87`; a compiled value, not a workflow key).
- `internal/homestate/card_age.go` (new file): `gateWaitingStates` (the class-a set), `PipelineCheckpoint(state) bool` (pipeline-stage membership for the REQ-CAA-004 anchor), `StageChangeCheckpoint(...)` (the iter1 D1 anchor: the most recent stage-change event — `to` a pipeline stage, differing from the prior such `to`), `GateWaitAge(...)`, `CardAgeSignals{...}`, `NoProgressAge(...)` — all read-only against an open `FactoryDB`.
- `internal/cli/factory_age_alarm.go` (new file): alarm computation driver (config resolve, per-card classification, banding, sort), the text section renderer, the JSON block type, and the bounded git head-commit reader.

## D. Milestones (RED first, per design §6)

### M1 — Configuration keys

1. RED: `internal/config/loader_card_age_alarm_test.go` — unset config resolves `enabled=false` + the four compiled defaults; a non-positive duration and a warn-not-below-alert pair are refused with the key path named; the template carries the block. Test fails to compile against the absent symbols (rejection test).
2. Implement: `CardAgeAlarmConfig` + workflow field (`types.go`), `LoadCardAgeAlarm` (new loader file, mirroring `LoadFactoryLease`'s fallback-and-warn read shape), the five constants in `defaults.go` (each comment declaring the value UNMEASURED-chosen, decision-index Q5), plus the compiled `CardAgeAlarmClockSkewTolerance` beside the quota precedent (`QuotaClockSkewTolerance`, `internal/config/defaults.go:82-87`; same shape — an unmeasured compiled value with a comment, not a workflow key; iter1 D2), and the template `workflow.yaml` block mirroring the defaults with the OFF flag visible.

### M2 — Age measurement (read-only, in `internal/homestate`)

1. RED: `internal/homestate/card_age_test.go` — seeded via the real transition API on a temp-dir factory DB (`t.TempDir`; test-isolation house rule): (a) a card entering `kickoff` via T7 measures gate age from the entering event's `created_at`, not from a later `card.fields` write; (b) a mid-merge expiry into `blocked` anchors on the `lease.expired` event; (c) a legacy row (no events) falls back to `updated_at`; (d) the three signals compute separately and a fresh heartbeat does NOT advance the no-progress age (the conflation mutant); (e) class-a/class-b partition and terminal/legacy exclusion; (f) the iter1 D1 stage-change regression, named `TestStageCheckpointSameStageReEntry`: a lease expiry → re-acquire → T4 same-stage re-entry (the expiry keeps the stage — `internal/homestate/card_transition.go:418-421` — and `stageAdmits` admits only that stage, `:811-818`) does NOT refresh the stage checkpoint or the no-progress age (the crash-loop mutant, which counts re-entry as progress, must fail this test), while stage changes in both directions DO refresh it — a forward advance and a backward T6 `plan-audit` → `plan` after an audit FAIL; (g) the iter2 rule-C creation-anchor cases, named `TestNoProgressAgeCreationAnchor`: (i) a pre-stage crash loop — ≥ 10 expiry → re-acquire cycles through the real transition API, no stage entry, no commit — preserves the no-progress age from the card's creation event, so a loop past the warn band still alarms (the B1 mutant, anchoring the both-unknown path on the most recent transition, must fail this test); (ii) a fresh card whose seeded worktree head carries a committer timestamp far older than the card (the inherited-base case) computes a no-progress age of ~0 measured from creation and raises no alarm (the B2 mutant, measuring from the inherited commit, must fail); (iii) a genuinely stuck pre-stage card — created past the alert band, no stage entry, no commit — still alerts, from creation (the true positive the clamp must keep); (h) the iter2 validity reach, named `TestAgeAnchorsFutureTimestamps`: with store sources stamped in the future beyond `CardAgeAlarmClockSkewTolerance` (a rolled-back reading clock, or a DB read from another machine), every derived age renders `unknown` and none is negative on any path — the REQ-CAA-003 gate-wait anchor event, its `updated_at` fallback, the creation-event anchor of REQ-CAA-005, and the legacy `updated_at` fallback (the B3 mutant, returning a negative age through any path, must fail).
2. Implement: `internal/homestate/card_age.go` — one bounded events query per render (`kind IN ('card.transition','lease.expired')`, payload `card_id`/`to` parsed), the anchors of REQ-CAA-003/004/005. No writes anywhere in the file.

### M3 — Status wiring and display

1. RED: `internal/cli/factory_age_alarm_test.go` — (a) disabled flag renders byte-identical status output (golden against the pre-change output over a seeded DB); (b) enabled flag renders the alarm section: banded `warn`/`alert`, oldest-first, one line per card with the three signal ages; (c) `--json` carries the additive `age_alarms` array and no pre-existing key moves; (d) a refused key degrades to exactly one warning line and the cards still render; (e) the no-write invariant: events count and card versions unchanged across an alarm computation over the seeded DB (the mutant — an alarm path that appends an event — must fail this test); (f) the iter1 D2 timestamp-validity case, named `TestFactoryStatusAgeAlarmFutureCommit`: over a temp git worktree whose head commit carries a future committer timestamp (a skewed committing machine), the commit signal renders `unknown` — excluded from the newest-of, never a negative age — and no alarm band is suppressed (the stage checkpoint still drives the class age, and the both-unknown fallback of REQ-CAA-005 still applies when it is the only signal).
2. Implement: `internal/cli/factory_age_alarm.go` + the three additive touches in `internal/cli/factory_card.go` — the `AgeAlarms` field on `factoryStatusReport` (placed before `Quota`, which stays last), one compute call in the status `RunE`, one render call in `writeFactoryStatusText`. The text alarm section renders one line per alarmed card under the literal token `age-alarm` (a tag-prefixed line, the `factoryRecordUnavailableTag` precedent in `writeFactoryStatusText`), and the renderer emitting it lives in `internal/cli/factory_age_alarm.go` — so the token is guaranteed to appear under `internal/cli/` from M3 on, which is exactly the flip acceptance criterion E-006 observes (iter1 D3). The git head-commit reader: one bounded invocation per candidate card with a non-empty `WorktreePath`; missing/unreadable worktree renders `unknown`; the read never fails the command.

Milestone order is fixed (M1 → M2 → M3); each milestone's RED lands in its first commit, implementation in its second (`fix(SPEC-CARD-AGE-ALARM-001): M<N> ...` / `feat(SPEC-CARD-AGE-ALARM-001): M<N> ...`).

## E. Verification plan

- Scoped single tests (`go test -run <Name> ./internal/config`, `./internal/homestate`, `./internal/cli`) are the per-milestone verification commands.
- Package-wide suites, cross-package builds, and any `-race` run are heavy under the 10-17 machine-load rule (`moai slot acquire --resource go-test-heavy` first); where the slot cannot be acquired the run is recorded and deferred, never skipped silently and never run anyway.
- The no-write invariant (M3 test (e)) is the display-only guard for REQ-CAA-007 and stays in the tree as a regression guard after first green.

## F. Sibling overlap with t1839 (merge-conflict discipline)

t1839 edits `internal/cli/factory_card.go` in the `factory next` selection region; this SPEC touches the same file only in the status region with three additive changes. Rules this plan follows: no reformatting or re-wrapping of existing status-region lines beyond the three touches; all new logic in `internal/cli/factory_age_alarm.go` and `internal/homestate/card_age.go`; no shared helper extracted from `factory_card.go` into a location t1839 might also move. A merge conflict between the two cards should be resolvable by taking both sides verbatim.

## G. Risks

- **Band noise from batch-waiting delivery cards** — intended visibility; calibration is G1's purpose (spec.md §Residual risk).
- **Events-table scan growth** — one bounded query per render; `updated_at` fallback keeps the alarm answering if events are ever pruned.
- **Git read on a locked worktree** — renders `unknown`; the status command never fails on it.
- **t1839 conflict** — §F discipline keeps both sides additive.
