---
id: SPEC-CARD-AGE-ALARM-001
title: "Work-item-age alarms on gate-waiting and stalled factory cards (card t1838, card-policy stage G1)"
version: "0.1.2"
status: completed
created: 2026-10-17
updated: 2026-10-18
author: manager-spec (card t1838)
priority: P1
phase: "v3.2.0 target"
module: "internal/homestate"
lifecycle: spec-anchored
tags: "factory, card-age, alarm, gate-wait, stall, status, display-only, card-t1838"
tier: M
related_specs: [SPEC-FACTORY-LEASE-RENEW-001, SPEC-QUOTA-AWARE-SCHEDULING-001]
---

# SPEC: Work-item-age alarms on gate-waiting and stalled factory cards (stage G1)

## HISTORY

- 0.1.0 — 2026-10-17 — Plan-phase artifacts authored for card t1838 (Class C, Tier M) on branch `WT-10-17-g1` from HEAD `7faf75fdfbe27e442ea2c17a1a90caf5792d09e6`. The card body sets the scope: measure the work-item age of cards waiting on gates (kickoff, needs-decision, blocked) and of cards with no progress signal, and display the result as alarms on the leader's situational panel and the status output; no automatic action (quarantine is G2, a separate future card gated on this SPEC's calibration data); progress signals — stage checkpoints, commits, hook activity — are displayed as distinct signals. The normative design input is the leader-authored card-policy redesign doc (`.moai/reports/card-policy-redesign-20261017/design.md`, primary checkout; measured: not present in this branch's tree at `7faf75fdf`), principles 2 and 9, the "Stuck work" row of §3, the G1/G2 rows of §6, and §8's first residual risk. Every age anchor and signal source cited below was measured against the store schema at `7faf75fdf`; where the store lacks a timestamp, the gap is named and the requirement states what is read instead. Requirement count 8, criterion count 9 (Tier M ceilings 16 / 16).
- 0.1.1 — 2026-10-17 — Plan-audit iteration 1 repairs (verdict FAIL, score 0.88, three blocking findings; `.moai/reports/t1838/plan-audit-iter1.md`): D1 — the REQ-CAA-004 stage-checkpoint anchor now counts stage changes only; a same-stage re-entry (the T4 expiry → re-acquire cycle) no longer refreshes it (M2 regression test added to plan.md). D2 — a timestamp-validity rule added to REQ-CAA-004/005 (a future-dated signal renders `unknown`, mirroring `QuotaClockSkewTolerance`, `internal/config/defaults.go:82-87`; named M3 test in plan.md). D3 — AC-CAA-006's RED grep retargeted to a plan-entailed observation (plan.md M3 pins the `age-alarm` marker; cell re-measured on `1cac5de4f`). Substrate citations added by this revision were measured on tree `1cac5de4fa4c0f45942c8e2e037de1bd7f2fac2d`, whose code paths are identical to the authoring pin `7faf75fdf` (the intervening plan-phase commit touched only SPEC artifacts). REQ-CAA-001..008 numbering unchanged. The optional finding D4's dual reading of "pipeline state" is resolved in passing by naming the stage set explicitly.
- 0.1.2 — 2026-10-18 — Plan-audit iteration 2 rule-C correction (verdict FAIL, score 0.86; the score regression 0.88 → 0.86 fired the plan-auditor escalation STOP signal, so this correction is exactly the verdict's own scope-reduction proposal — `.moai/reports/t1838/plan-audit-iter2.md`, board d-20261015T094733Z-9e76): (a) REQ-CAA-005 admits the card's creation-event timestamp into the no-progress-age basis — the age is measured from the newest of the stage-checkpoint, commit, and creation-event timestamps — so the age can neither predate the card's existence (finding B2: a fresh card whose worktree head is an inherited base commit older than the band gets age-from-creation, not an instant stall) nor be reset by pool-state cycling (finding B1: with stage and commit unknown the creation event anchors the age, so a pre-pipeline expiry → re-acquire crash loop keeps its stuck time visible; the prior both-unknown fallback — the most recent `card.transition` — was the reset vector and is withdrawn); the verified iter1 D1 post-entry predicate is untouched, creation being by construction the card's earliest event. (b) The timestamp-validity rule now binds every timestamp the displayed ages derive from (finding B3): the REQ-CAA-005 basis members including the creation event, the legacy `updated_at` fallback, and the REQ-CAA-003 gate-wait anchor with its fallback — a future-dated source beyond `CardAgeAlarmClockSkewTolerance` renders `unknown`, and no derived age is ever negative. (c) plan.md M2 pins the four regression cases (pre-stage crash loop preserves age; fresh card with old base does not alert; stuck pre-stage card still alerts from creation; rolled-back clock yields no negative age on any path), and acceptance.md's AC-CAA-004 green path mirrors the corrected basis (E-004 re-measured on `c14bb152a`). decision-index gains Q9/Q10. REQ-CAA-001..008 numbering unchanged; no new requirement added (sub-clauses within REQ-CAA-003/004/005 only). Substrate citations reused by this revision (creation event `internal/homestate/card_picked.go:173`; T3 re-acquire `internal/homestate/card_transition.go:136`; `QuotaClockSkewTolerance` `internal/config/defaults.go:82-87`) were re-verified on tree `c14bb152a9e87900cbf9233e2e9029d24573c7a1`, whose code paths are identical to the authoring pin `7faf75fdf` (the intervening commits touched only SPEC artifacts).

## Scope

This SPEC carries stage G1 of the card-policy redesign: display-only work-item-age alarms on the two stuck-work classes the design names — cards waiting on a human gate, and cards showing no progress signal. The alarm renders on the factory status output the leader reads. It takes no action of any kind: no quarantine, no takeover, no notification, no write.

1. **Age measurement.** Gate-wait age for cards in `kickoff`, `needs-decision`, and `blocked`; no-progress age for cards with no stage-checkpoint or commit signal. Both anchors derive from timestamps the store already records (requirements REQ-CAA-003, REQ-CAA-005); the one signal the store does not record — the last commit — is read from the card worktree's git history, which is itself a recorded timestamp (REQ-CAA-004).
2. **Three-signal distinction.** Stage checkpoints, commits, and hook (lease-heartbeat) activity are three different signals (design principle 2: liveness ≠ activity ≠ progress). The alarm output shows all three ages separately and never lets activity masquerade as progress (REQ-CAA-004).
3. **Display.** Alarm rows on the status text output and the `--json` report, banded and sorted by age. Display-only (REQ-CAA-006, REQ-CAA-007).
4. **Configuration.** An enable flag (default OFF, design §6 stage discipline) and four age-band thresholds as `workflow.card_age_alarm.*` keys with compiled defaults in `internal/config/defaults.go`; no hardcoded thresholds (REQ-CAA-001, REQ-CAA-002).

The SPEC touches `internal/config` (types, loader, defaults), `internal/homestate` (a new read-only age-measurement file), `internal/cli` (the factory status wiring and a new alarm file), and the workflow configuration template. It does not modify the transition table, the lease semantics, the selection path, or any writing verb.

## Background (measured substrate)

All file:line citations below were measured on tree `7faf75fdfbe27e442ea2c17a1a90caf5792d09e6`.

### The store

- The `cards` table (`internal/homestate/factory.go:45-78`) carries `updated_at` (`:52`), `heartbeat_at` (`:56`), `decision_gate` (`:57`), and `stage` (`:53`), with no `created_at` column. The `Card` struct mirrors them (`internal/homestate/card_record.go:100-138`; `UpdatedAt` `:107`, `HeartbeatAt` `:111`, decision fields `:112-114`, `EvidenceSHA` `:122`, `MergeSHA` `:123`).
- The `events` table (`internal/homestate/factory.go:79-85`) appends one row per record act: `appendEvent` (`internal/homestate/card_transition.go:901-908`) writes `kind`, `payload_json` (carrying `card_id`, `from`, `to`, `version`), and `created_at`. Card creation is itself a `card.transition` event with `from: ""`, `to: picked` (`internal/homestate/card_picked.go:173`). Field-only writes append a `card.fields` event and stamp `updated_at` without any transition (`internal/homestate/card_picked.go:197`) — this is why `updated_at` alone is not a safe age anchor (decision-index Q1).
- Every card-row write stamps `updated_at`: transitions through `updateCardRow` (`internal/homestate/card_transition.go:875-899`), lease renewal (`:523-528`), and lazy lease expiry (`:422-439`; expiry mid-merge lands `blocked` at `:426-430` and appends a `lease.expired` event whose payload carries `to` at `:434-438`). Renewal and expiry touch lease-holding states only; a card sitting in `kickoff`, `needs-decision`, or `blocked` has no lease to renew (`leaseHoldingStates`, `internal/homestate/card_record.go:66-68`).
- Entering the gate states: T21 enters `needs-decision` from any active state (`guardQuestion`, `internal/homestate/card_transition.go:789-797`); T7 lands `kickoff` on a PASS plan verdict (`:636-640`); T9/T23 land `blocked`; lazy expiry lands `blocked` for mid-merge cards. Terminal states are `done`, `failed`, `abandoned` (`internal/homestate/card_record.go:74`); the delivery states `pr-open` and `merged-pr` exist beside them (`internal/homestate/card_pr_states.go:36-38`).

### The display surfaces

- `moai factory status` is the leader's situational read. The command is `newFactoryStatusCommand` (`internal/cli/factory_card.go:3106-3178`); the text renderer is `writeFactoryStatusText` (`:3180-3215`); the `--json` report is `factoryStatusReport` (`:3065-3083`) with per-card views from `factoryCardViewOf` (`:3085-3097`). No other card-level leader-panel surface exists: the factory subcommand registry (`internal/cli/factory_handoff_recover.go:56-61`) and the MCP factory tools (`internal/cli/mcp_server.go:619-628`, `factory_msg_*` only) were measured — there is no factory-status MCP tool and no separate board renderer.
- The command already degrades a failed advisory read rather than failing the whole surface: the queue-classification read degrades the mode/priority cells when it faults (`internal/cli/factory_card.go:3131-3136`). The alarm follows that precedent (decision-index Q3).

### The configuration precedent

- `workflow.factory_lease.*` is the house pattern for factory tuning keys: struct and workflow field (`internal/config/types.go:435-450`), dedicated loader (`internal/config/loader_factory_lease.go:24`), compiled defaults (`internal/config/defaults.go:61-70`, wired at `:1454-1456`), bounds checked in the reader (`internal/homestate/factory_lease.go:36-60`), and a shipped template block (`internal/template/templates/.moai/config/sections/workflow.yaml:21-29`). The alarm keys mirror this shape (REQ-CAA-002), with UNMEASURED chosen default values in the quota-gate tradition (`internal/config/defaults.go:89-98`).

### Timestamp gaps in the store (named, not invented)

1. **No commit timestamp exists in the factory store.** `evidence_sha` records the single run→sync commit (T10) and `merge_sha` the merge (T16) — no per-commit record, no last-commit timestamp. REQ-CAA-004 therefore reads the last-commit age from the head commit of the card's recorded worktree (`WorktreePath`, `internal/homestate/card_record.go:121`) at render time — git's own recorded timestamps, read-only. Recording a commit timestamp in the store is explicitly out of scope (Boundaries).
2. **No hook-activity record beyond the lease heartbeat.** `cards.heartbeat_at` is written only on lease acquire, audit-kickoff lease, and renewal (`internal/homestate/card_transition.go:601-605`, `:661-665`, `:523-528`) — only in lease-holding states. The activity signal therefore renders `unknown` (never a fabricated zero) for gate-state and unleased cards. A finer hook-level activity record does not exist and is not added here (Boundaries).
3. **No `created_at` column on `cards`.** A picked card's age anchors on its creation event (`internal/homestate/card_picked.go:173`) through the events path; no column is added.

## Requirements (GEARS)

### REQ-CAA-001 — Enable flag, default OFF

When `workflow.card_age_alarm.enabled` is false or unset, `moai factory status` (both the text output and the `--json` report) shall produce output byte-identical to the pre-SPEC output over the same card set, and the alarm code path shall perform no work beyond resolving the flag.

### REQ-CAA-002 — Threshold keys, defaults, and degradation

The system shall read four age-band thresholds — `workflow.card_age_alarm.gate_warn_after`, `gate_alert_after`, `stall_warn_after`, `stall_alert_after` — through a dedicated config loader with compiled defaults in `internal/config/defaults.go` (UNMEASURED chosen values; see decision-index Q5) and a shipped template block mirroring them. The measurement and rendering code shall contain no age-threshold literal. When a key fails its bound — not a positive duration, or a warn band not strictly below its alert band — the status command shall omit the alarm section and print exactly one warning line naming the refused key, and shall not refuse the rest of the status output.

### REQ-CAA-003 — Gate-wait age anchor

For a card whose state is `kickoff`, `needs-decision`, or `blocked`, the system shall compute the gate-wait age as now minus the `created_at` of the most recent `events` row of that card whose kind is `card.transition` or `lease.expired` and whose payload `to` equals the card's current state. When no such event exists (a legacy or predating row), the system shall fall back to `cards.updated_at`. Both the anchor event's `created_at` and the `updated_at` fallback are subject to the timestamp-validity rule of REQ-CAA-004 (plan-audit iteration 2 finding B3): a source dated in the future beyond `CardAgeAlarmClockSkewTolerance` renders that source `unknown`, and the gate-wait age renders `unknown` when no valid basis remains — never a negative age. Both reads are read-only against the existing schema.

### REQ-CAA-004 — Three signals, displayed separately

For every alarmed card the system shall compute and display three signal ages separately:

- **stage checkpoint** — now minus the `created_at` of the most recent `card.transition` event of that card that records a **stage change**: its payload `to` is a pipeline stage, and that `to` differs from the `to` of the card's most recent prior `card.transition` event whose `to` is a pipeline stage (a first entry into the pipeline, with no such prior event, is a stage change). The pipeline stages are the working states (`plan`, `plan-audit`, `run`, `sync`, `sync-audit`, `merge-ready`, `merging`) and the delivery states (`merged-local`, `pushed`, `ci-green`, `pr-open`, `merged-pr`) — the class (b) set of REQ-CAA-008 minus the pre-pipeline pool states (`picked`, `assigned`, `leased`), which a card passes through but never works in. An event whose `to` equals the pipeline stage the card already holds is a **same-stage re-entry** and shall NOT refresh the checkpoint. The T4 stage-resume edges are exactly this shape by construction — an expired card returns to `assigned` with its stage kept (`applyLeaseExpiry`, `internal/homestate/card_transition.go:418-421`) and may then resume only into that kept stage (`stageAdmits`, `:811-818`; the T4 edges at `:137-142`) — so an expiry → re-acquire → same-stage cycle (`lease.expired` → T3 → T4, with no commit and no stage advance) leaves the checkpoint, and with it the no-progress age of REQ-CAA-005, anchored at the last genuine stage change instead of being reset on every cycle: that cycle is the crash-loop pattern this SPEC exists to surface, and activity is not progress (design principle 2). Stage changes in either direction qualify — a backward move (T6 `plan-audit` → `plan` or T12 `sync-audit` → `sync` after an audit FAIL; T15 `merging` → `merge-ready` after a refused merge) refreshes the checkpoint exactly as a forward advance does; only re-entry into the stage already held is excluded (plan-audit iteration 1 finding D1; regression-pinned by plan.md M2 (f));
- **commit** — now minus the committer timestamp of the head commit of the card's worktree branch, read at render time through a read-only git invocation against `WorktreePath`; when the worktree path is empty, missing, or unreadable, or the git invocation itself fails, the signal shall render `unknown`, never a fabricated value. The commit timestamp is also subject to the validity rule below, because it is the one signal that crosses machine boundaries (worktree commits arrive from arbitrary machines);
- **activity** — now minus `cards.heartbeat_at`; when the column is empty, the signal shall render `unknown`.

Timestamp validity: a signal timestamp that lies in the future beyond a compiled clock-skew tolerance renders that signal `unknown` — never an age of zero, never a negative age (plan-audit iteration 1 finding D2). The tolerance is the compiled constant `CardAgeAlarmClockSkewTolerance` in `internal/config/defaults.go`, mirroring the house precedent `QuotaClockSkewTolerance` (`internal/config/defaults.go:82-87`, REQ-QAS-005: a record from another machine or a skewed clock must not win by being "newest"). It is a compiled value, not a `workflow.card_age_alarm.*` key, and it is the one sanctioned home of the number under REQ-CAA-002's no-literal rule. The rule binds every timestamp the displayed ages derive from, not the three signal timestamps alone (plan-audit iteration 2 finding B3): the REQ-CAA-005 basis members — the stage-checkpoint, commit, and creation-event timestamps — the legacy `updated_at` fallback, and the REQ-CAA-003 gate-wait anchor with its own `updated_at` fallback. A future-dated source renders that source `unknown`; an age whose basis members are all invalid renders `unknown`; no derived age — gate-wait or no-progress — is ever negative. The commit read remains the realistic future-dated case, as worktree commits arrive from arbitrary machines; the store-written paths bind for the rolled-back-clock and DB-read-from-another-machine cases, where the reading machine's own clock did not stamp the source.

The alarm shall not conflate activity with progress (design principle 2): the heartbeat signal shall never enter the no-progress age computation of REQ-CAA-005.

### REQ-CAA-005 — No-progress age

For a card covered by class (b) of REQ-CAA-008, the system shall compute the no-progress age as now minus the newest of three basis timestamps: the stage-checkpoint timestamp (REQ-CAA-004), the commit timestamp, and the card's creation-event timestamp — the `created_at` of the card's creation event, the `card.transition` event whose payload `from` is empty (`internal/homestate/card_picked.go:173`; exactly one exists per card that has events, and it is by construction that card's earliest event). A basis member rendered `unknown` — absent, unreadable, or invalid under the timestamp-validity rule of REQ-CAA-004 — contributes no timestamp to the newest-of: the newest-of is taken over the known members only, so a future-dated member can never yield a negative age and never suppresses either band. The creation member bounds the age in both directions (plan-audit iteration 2 findings B1/B2): the age can never predate the card's existence — a fresh card whose worktree head is an inherited base commit older than the band measures from creation, not from the base (B2) — and pool-state cycling can never reset it: when the stage-checkpoint and commit members are both unknown, the creation event anchors the age, so a pre-pipeline card cycling through expiry → re-acquire (each cycle a fresh T3 `assigned` → `leased` transition, `internal/homestate/card_transition.go:136` — activity, not progress, design principle 2) keeps its stuck time visible (B1). The prior both-unknown fallback — now minus the most recent `card.transition` event of that card — was that reset vector and is withdrawn. Because creation precedes every other event of the card, a card holding a known stage checkpoint or a post-creation commit computes an unchanged newest-of: the iter1 D1 post-entry predicate is untouched. When no event exists at all (a legacy row), the system shall fall back to `cards.updated_at` (the fallback of REQ-CAA-003), subject to the same validity rule.

### REQ-CAA-006 — Alarm rows, banding, sorting, surfaces

The system shall render one alarm row per card whose gate-wait age or no-progress age crosses its class's warn threshold, on both status surfaces: a text section in `writeFactoryStatusText` (`internal/cli/factory_card.go:3180-3215`) and an additive `age_alarms` array in the `--json` report placed so every pre-existing key keeps its place (the quota block stays last). Each row shall carry the card id, run id, state, class (`gate-wait` or `no-progress`), band (`warn` or `alert`), the class age, and the three signal ages of REQ-CAA-004. Rows shall sort oldest-first. A card below every threshold shall render no alarm row.

### REQ-CAA-007 — Display-only invariant

The alarm path shall be strictly read-only: no card-row write, no event append, no lease change, no transition, no queue write, no notification, and no process signal. Nothing in this SPEC shall take, prepare, or schedule an automatic action on an alarmed card — quarantine and takeover belong to stage G2, which is gated on the calibration data this SPEC's alarms produce.

### REQ-CAA-008 — Coverage set

Class (a) — gate wait — is exactly the states `kickoff`, `needs-decision`, and `blocked`. Class (b) — no progress — is every card whose state is a non-terminal F1 state (`internal/homestate/card_record.go:59-64`, with the delivery states of `internal/homestate/card_pr_states.go`) and not in class (a); the delivery states (`merged-local`, `pushed`, `ci-green`, `pr-open`, `merged-pr`) are included, because a landing that waits past its batch is stuck work the leader must see. Terminal states (`done`, `failed`, `abandoned`) and legacy rows (`Card.Legacy()`, `internal/homestate/card_record.go:142`) shall be excluded from both classes.

## Boundaries (explicit out-of-scope)

- **Automatic takeover or quarantine (G2).** Not here. G2 is a separate future card gated on this SPEC's calibration data (design §6).
- **Process-liveness probing (stage A).** G1 reads recorded timestamps only. It does not probe pids, sessions, or process identity; the single git invocation of REQ-CAA-004 reads commit timestamps git already recorded and touches no process state.
- **Lease or ownership semantics (stage A).** The lease duration, renewal, expiry, and holder rules are unchanged.
- **Metadata verbs (stage E).** `set-spec`, reclassify, and evidence carry are not added or altered.
- **Eligibility-filtered selection (stage C, sibling card t1839).** The `factory next` selection path is not touched. t1839 lands in the same file (`internal/cli/factory_card.go`); this SPEC therefore keeps its edits there additive and minimal (a struct field, a compute call, a render call), with the alarm logic in new files, so the two cards review separately (plan §Sibling overlap).
- **The serial slot (stage D / t1835) and the integration window.** Untouched.
- **Recording new timestamps in the store.** No commit-timestamp column, no hook-activity table, no `created_at` column. The gaps named in Background are read around, not filled.
- **MCP tools and skill bodies.** No factory-status MCP tool exists (measured, `internal/cli/mcp_server.go:619-628`) and none is added; the foreman and watchdog skills are mirrored templates and are not edited — the alarm section is self-describing in the status output.

## Residual risk

- **Band noise on batch-waiting cards.** Delivery-state cards legitimately wait for the leader's batch; with the proposed stall bands they will alarm once past the threshold. That is intended visibility, not a defect, and the calibration data G1 exists to produce will tune it; a separate batch-wait class is a possible follow-up once data exists.
- **Git-read cost at render.** One bounded `git log -1` per alarmed-card candidate with a non-empty worktree path. On a run with many leased cards this is a bounded per-card subprocess cost inside an on-demand command, not a lifecycle hook. A missing or locked worktree renders `unknown`; it never fails the status output.
- **Events-table dependence.** The anchors of REQ-CAA-003/004/005 read the `events` table, which has no pruning today; the read is one bounded query per status render. If a future change prunes or partitions events, the `updated_at` fallback keeps the alarm answering, less precisely.

## Cross-references

- Design input: `.moai/reports/card-policy-redesign-20261017/design.md` (primary checkout; not committed in this branch's tree at `7faf75fdf`) — §2 principles 2 and 9, §3 "Stuck work" row, §6 G1/G2, §8 residual risks.
- Config-key precedent: SPEC-FACTORY-LEASE-RENEW-001 (`workflow.factory_lease.*`); UNMEASURED-defaults precedent: SPEC-QUOTA-AWARE-SCHEDULING-001 (`workflow.quota_gate.*`).
- Two-cell acceptance discipline: `.claude/rules/moai/development/verification-completeness.md` §2.
