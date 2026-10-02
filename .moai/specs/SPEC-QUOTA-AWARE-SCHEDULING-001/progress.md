# Progress — SPEC-QUOTA-AWARE-SCHEDULING-001

## §E.1 Plan-phase Audit-Ready Signal

- plan_status: audit-ready
- plan_artifacts: spec.md, plan.md, acceptance.md, design.md, research.md (Tier L), plus decision-index.md
- plan_complete_at: 2026-10-02
- open_decisions: none — DO-1..DO-12 resolved (oracle; leader verdict on DO-3, DO-7, DO-8; DO-12 `write_backend_at_claim` verified then applied at spec 0.4.0)
- plan_audit_verdict: iteration 1 FAIL 0.74 (`.moai/reports/t1347/plan-audit-iter1.md`, local-only); revised at spec 0.5.0 to close D1-D24, awaiting iteration 2
- open_decisions_note: DO-13 (first-exhausted time surfaced in the status block) is PROVISIONAL, escalated to the leader

## §E.2 Run-phase Evidence

_pending run-phase_

## §E.3 Run-phase Audit-Ready Signal

_pending run-phase_

## §E.4 Sync-phase Audit-Ready Signal

_pending sync-phase_

## §J Lane Decision Log (card t1347, lane-9)

Rule (leader dispatch, operator instruction): in-flight decisions are asked of Jev (TypeSafe System One, `jev-1.13.0`) through the lane's dispatch script; question, answer and confidence are recorded here. Confidence < 0.5 or a hard-to-reverse external action goes to the leader. The oracle's answer is advisory evidence; where a fact it was given could be checked in code, the lane checked it first (DO-5: exit status 3 is `factoryNextNoCardExit`, `internal/cli/factory_card.go:181` (comment at `:179-180`), "3, so a supervising launcher can distinguish it from failure").

| ID | Question (short) | Oracle answer | Confidence | Disposition |
|----|------------------|---------------|-----------:|-------------|
| DO-1 | Stamp the hook-observed 429 turn-end into the record? | not_now | 1.00 | applied |
| DO-2 | Distinguish two Claude accounts? | reset_time_only | 0.97 | applied |
| DO-3 | Default thresholds (5h / 7d / margin / max age) | 90 / 95 / 5 / 30m (P 0.54) | 0.32 LOW | leader: accepted as provisional; all four exposed as config keys; "unmeasured defaults"; re-tune after first measurement is a listed follow-up |
| DO-4 | Template default for `workflow.quota_gate.enabled` | false in template, true in local config | 0.99 | applied |
| DO-5 | Exit status for the hold | reuse status 3 (P 0.90) | 0.81 | applied (SPEC default was 4; changed) |
| DO-6 | Gate arm (a), a card already assigned to the lane? | exempt started cards | 0.99 | applied (SPEC default was gate-all; changed) |
| DO-7 | `integration acquire` under pressure | warn_only (P 0.54 vs 0.46) | 0.07 LOW | leader: warn-only, never blocks, FINAL |
| DO-8 | Steer `--auto` / leader beyond a status block? | steer_auto (P 0.66) | 0.32 LOW | leader: REJECTED the lane's status-only provisional; steering IN SCOPE in its minimal form (recommend non-Claude lanes, Claude lanes hold with status 3, warning-only when no non-Claude lane, no forced re-dispatch); SPEC reclassified Tier M to Tier L |
| DO-9 | Stale high reading before reset | treat_unknown (P 0.93) | 0.86 | applied |
| DO-10 | Claude-lane predicate source | launch provider first, kanban backend fallback | 0.96 | applied |
| DO-11 | Record carrier | per-session record | 1.00 | applied |
| DO-12 | Where to read a lane's backend | write_backend_at_claim (P 0.78) | 0.55 | applied after the lane required feasibility verification (U18); verified by manager-spec and re-checked by the lane: `workers.backend TEXT NOT NULL DEFAULT ''` at `internal/homestate/factory.go:27`; claim inserts at `internal/kanban/factory_slots.go:106` and `:329` omit it |

Leader messages: plan commit and the three low-confidence items reported to the leader; verdict received for DO-3 / DO-7 / DO-8 (above). No hard-to-reverse external action has been taken (no push, PR or delete).

Plan artifacts: `400b5f986` (initial), `fdfb2d0c8` (oracle resolutions + leader verdict, Tier L), `ca57dc350` (DO-12). Tier L: REQ 23 of 25, AC 23 of 25.
