# SPEC-TODO-AUTO-PICK-001 — Progress

## §E.1 Plan-phase Audit-Ready Signal

```yaml
plan_status: audit-ready
plan_complete_at: 2026-10-02T15:12:45Z   # iteration-3 correction (UTC; spec.md dates it 2026-10-03 local); iteration-2 repair 2026-10-02T14:51:20Z; iteration-1 repair 2026-10-02T12:22:07Z; first plan completion 2026-10-02T11:55:40Z
card: t1448
tier: M
plan_head: 625f01718   # iteration-3 correction base HEAD (the correction commit follows it); iteration-2 repair base 63daaf6a7; iteration-1 repair base b3646de10; first plan authored at 4bf547bca
plan_audit: FAIL iteration 3 (0.80, MP-1..9 PASS, one blocker F1); F1 and notes corrected in this commit WITHOUT a re-audit — leader decision pending   # iteration 1: FAIL 0.73 (b3646de10); iteration 2: FAIL 0.77 (63daaf6a7); iteration 3: FAIL 0.80 (625f01718; .moai/reports/t1448/plan-audit-iter3.md, not committed); Tier M threshold 0.80; the plan-artifact hashes the iteration-3 verdict recorded are invalidated by this correction
plan_audit_iter4: PASS 0.85 (delta, operator-commissioned; audited hash 724841520)   # Clarity 0.80, Completeness 0.88, Testability 0.82, Traceability 0.88; Tier M threshold 0.80, margin +0.045; auditor-model claude-sonnet-5-5[1m]; verdict file .moai/reports/t1448/plan-audit-iter4-delta.md (card tree only, not committed); the iteration-3 line above is history and stays as written
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

### Iteration-4 delta plan-audit result (appended; iterations 1-3 above are unchanged)

- Iteration 4 is an operator-commissioned delta audit, outside the Tier M ceiling of 2 audits,
  commissioned after the iteration-3 FAIL (0.80). Verdict **PASS**, overall score **0.85**
  (Clarity 0.80, Completeness 0.88, Testability 0.82, Traceability 0.88); margin +0.045 over the
  Tier M threshold 0.80 (iteration 3 was 0.80). auditor-model `claude-sonnet-5-5[1m]`. Audited
  commit `724841520` (full `72484152041c8e894947fb356835678b4d9b0023`). Verdict file:
  `.moai/reports/t1448/plan-audit-iter4-delta.md` (local, card tree only).
- Scope: closure of F1 (catalog.yaml row, generator step, two guard tests, AC-TAU-010/-012/-013)
  plus new contradictions from the F2 and S1..S8 edits. F1 RESOLVED; no MUST-FIX open; MP-1..MP-9
  PASS or N/A. No audit receipts (no cross-model tool invoked).
- Optional notes D1..D5: D1 serial-slot re-adoption is order-conditional; D2 two more English
  user-facing doc hits (`docs-site/content/en/advanced/factory-mode.md:73`,
  `docs-site/content/en/cli-reference/launchers.md:35`) fall in sync-phase scope; D3 the M6
  selector literal is not printed in M6 itself but is in M5 step 5 and AC-TAU-010; D4 G7
  abbreviates sha256 values; D5 this progress record was stale until this append.
- Per-file sha256 at `724841520` (the audited hashes; re-computed at the same HEAD and equal):

  ```text
  spec.md        3d68cb79b5f2e09f8e98223f35d2d2a8e283670b263ac6c2715d272bb744d9ce
  plan.md        d251637af3f4b29f76b873dbc1abccb621888975c98e5ea1ab04b58d6d5931f9
  acceptance.md  e7242951b8b0fe7bdd835194aca41bed31400f702f0796d6beaaf2ea6d8377db
  research.md    6dc5361d386e9318a72e7402358b3013213f6164fabe69afd57334085356a26a
  spec-compact.md b0a1b4f78eda903e3320b6e587b303fe5947ca8993e516b28c50aff1d9182e76
  progress.md (at 724841520, before this append)
                 76d81e719d7c6ed79c97752f3dcb43d98f00395d429516ff925f4dbb6dd92a80
  ```

  `progress.md` is not a member of the plan-artifact hash set
  `{acceptance.md, design.md, plan.md, research.md, spec.md, tasks.md}`
  (`.claude/rules/moai/workflow/spec-workflow.md` § Report Persistence), so this append does not
  invalidate the verdict.

### Plan->run Kickoff record (autonomous form, `auto-semantics.md` §9.1)

The Kickoff conditions are met: independent plan-audit verdict PASS (iteration 4, by plan-auditor);
`plan_status: audit-ready` recorded above; plan-artifact hashes unchanged since that verdict
(sha256 values above re-computed at `724841520`); no open blocker (no MUST-FIX, no MP violation).
Run Phase 1 skip contract (`spec-workflow.md` § Phase Transitions) holds on all three conditions:
verdict PASS; score 0.85 >= Tier M threshold 0.80; artifact hash unchanged.

```text
decision record: decided_by=operator-decision(2026-10-03, AskUserQuestion answer relayed by the factory leader: 'correct, then re-audit only F1') + lane-4 orchestrator evidence_refs=.moai/reports/t1448/plan-audit-iter4-delta.md;.moai/reports/t1448/plan-audit-iter3.md;audited-hash=724841520 ladder_path=gate-row plan->run Kickoff (AUTONOMOUS, auto-semantics §9.1; operator decision 'correct then re-audit F1 only', audited hash 724841520)
```

Forward notes for the run phase:

- develop (`7109e0900` at the time of writing) already carries t1451's edits to `kanban-dispatch.md`
  (card-review stage + standing recheck cron paragraphs), `kanban-dispatch-detail.md`,
  `auto-semantics.md` and `catalog.yaml`. The SPEC's pinned baselines (RED-now cells at `4bf547bca`,
  `kanban-dispatch.md` byte baselines 26959 live / 26637 mirror) must be RE-MEASURED after the develop
  absorption, and the new values recorded in §E.2 by the run phase.
- The five old-authority literals L5/L6/L7/L8/L25 were re-checked on develop and are all still present
  in live and mirror, so the SPEC's replacement targets exist.
- develop `kanban-dispatch.md` is 28308 B live / 27986 B mirror; their 322 B difference (the
  pre-existing line-177 drift) is unchanged.
- A later card (t1399, lane-3) renames `kanban-dispatch*.md` -> `factory-dispatch*.md`, the foreman
  skill and `internal/kanban` -> `internal/factory`. If it lands first, the run/sync phase re-greps the
  old paths and runs that card's rename script (`.moai/reports/t1399/rename/`) before re-measuring.

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
