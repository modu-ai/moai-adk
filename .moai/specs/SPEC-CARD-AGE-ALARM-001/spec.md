---
id: SPEC-CARD-AGE-ALARM-001
title: "Work-item-age alarms on gate-waiting and stalled factory cards (card t1838, card-policy stage G1)"
version: "0.1.0"
status: draft
created: 2026-10-17
updated: 2026-10-17
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

For a card whose state is `kickoff`, `needs-decision`, or `blocked`, the system shall compute the gate-wait age as now minus the `created_at` of the most recent `events` row of that card whose kind is `card.transition` or `lease.expired` and whose payload `to` equals the card's current state. When no such event exists (a legacy or predating row), the system shall fall back to `cards.updated_at`. Both reads are read-only against the existing schema.

### REQ-CAA-004 — Three signals, displayed separately

For every alarmed card the system shall compute and display three signal ages separately:

- **stage checkpoint** — now minus the `created_at` of the most recent `card.transition` event of that card whose payload `to` is a pipeline state (the working and delivery states; the set is fixed by REQ-CAA-008);
- **commit** — now minus the committer timestamp of the head commit of the card's worktree branch, read at render time through a read-only git invocation against `WorktreePath`; when the worktree path is empty, missing, or unreadable, the signal shall render `unknown`, never a fabricated value;
- **activity** — now minus `cards.heartbeat_at`; when the column is empty, the signal shall render `unknown`.

The alarm shall not conflate activity with progress (design principle 2): the heartbeat signal shall never enter the no-progress age computation of REQ-CAA-005.

### REQ-CAA-005 — No-progress age

For a card covered by class (b) of REQ-CAA-008, the system shall compute the no-progress age as now minus the newest of the stage-checkpoint and commit timestamps. When both are unknown, it shall fall back to now minus the most recent `card.transition` event timestamp of that card (the `updated_at` fallback of REQ-CAA-003 when no event exists).

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
