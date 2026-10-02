# SPEC-TODO-AUTO-PICK-001 — Progress

## §E.1 Plan-phase Audit-Ready Signal

```yaml
plan_status: audit-ready
plan_complete_at: 2026-10-02T15:12:45Z   # iteration-3 correction (UTC; spec.md dates it 2026-10-03 local); iteration-2 repair 2026-10-02T14:51:20Z; iteration-1 repair 2026-10-02T12:22:07Z; first plan completion 2026-10-02T11:55:40Z
card: t1448
tier: M
plan_head: 625f01718   # iteration-3 correction base HEAD (the correction commit follows it); iteration-2 repair base 63daaf6a7; iteration-1 repair base b3646de10; first plan authored at 4bf547bca
plan_audit: FAIL iteration 3 (0.80, MP-1..9 PASS, one blocker F1); F1 and notes corrected in this commit WITHOUT a re-audit — leader decision pending   # iteration 1: FAIL 0.73 (b3646de10); iteration 2: FAIL 0.77 (63daaf6a7); iteration 3: FAIL 0.80 (625f01718; .moai/reports/t1448/plan-audit-iter3.md, not committed); Tier M threshold 0.80; the plan-artifact hashes the iteration-3 verdict recorded are invalidated by this correction
artifact_set: spec.md, plan.md, acceptance.md, research.md, spec-compact.md, progress.md
requirements: 16       # Tier M ceiling 16
acceptance_criteria: 14  # Tier M ceiling 16
needs_clarification: 0
```

Plan-phase notes (what a reader of this record needs, nothing populated for later phases):

- Baselines the criteria measure against were observed in this tree before any run commit and land
  in the plan commits (`verification-claim-integrity.md` §2.3): `acceptance.md` ledger rows
  L1-L47, C1-C5, S1-S2, context rows G1-G8.
- **Iteration-3 correction (F1, F2, S1-S8) — not audited.** The audit (`625f01718`) returned FAIL
  0.80 on one blocker: `internal/template/catalog.yaml` stores hashes for the `moai` and
  `moai-kanban-foreman` skill directories and the template `manager-todo.md`, and the plan did not
  regenerate them. The plan now does (plan §5, M5, M6, AC-TAU-010/-012/-013, DoD 7), and the red was
  **observed** on a reversible perturbation (ledger G7: exit 1, `CATALOG_HASH_UNSTABLE` ×3,
  `CATALOG_HASH_SKINNY` ×2; restored, `git status --short` empty, `cmp` against backups empty). The
  correction changes the plan artifacts, so the iteration-3 hashes no longer describe them and a
  verdict on the corrected plan does not exist; this record does not claim a PASS. REQ 16, AC 14.
- Iteration-2 repair: N1..N7 and the leader's additional requirements are resolved in
  `plan.md` §10 (Audit resolution map). The two deliberate deltas the leader may veto without ripple
  remain isolated: D-DEF (second clause of REQ-TAU-007) and D-LANE (REQ-TAU-008).
- **AC baseline snapshot guard — observed first, as the leader required.**
  `unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test ./internal/spec -run TestACCounterFullCorpusMatchesBaseline -count=1`
  on `63daaf6a7`, **before** this repair edited `acceptance.md`: exit 0, `ok  …/internal/spec  6.230s`
  (**already green** — the two earlier SPEC commits had not moved a recorded count). After the
  repair, with the anchored selector and `-v`: exit 0,
  `--- PASS: TestACCounterFullCorpusMatchesBaseline (6.84s)`; the test lists this SPEC's
  `acceptance.md` as `absent-from-snapshot … COUNT 14` (reported, not failed — a SPEC not in the
  snapshot is not a baseline movement). **No regeneration was needed**, so
  `.moai/reports/t338/ac-count-baseline.txt` is untouched and not staged.
- The compensating-control claim for the decision record stays retracted: no party executes a re-read
  of the line today (spec §B.5, §G); auto-semantics §9.1/§10 wording stands for the other gates.
- Unverified items are listed in `research.md` R10 and R12.
- The throwaway probe file `internal/cli/zz_t1448_probe_test.go` was created and deleted during
  research; it never entered a commit. Its four observations are `research.md` R2 (O1-O4). The
  scratch binary and queue of rows S1/S2 live outside the tree; nothing of them is committed.

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
