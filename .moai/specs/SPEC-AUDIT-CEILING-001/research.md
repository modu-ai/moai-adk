# research.md — SPEC-AUDIT-CEILING-001

Source verification for card t1500. Every claim below was measured in this
worktree (`.moai/worktrees/t1500`, branch `WT-audit-ceiling-counter`). The
original measurement set was taken at HEAD `2f492df19` on 2026-10-04; the
v0.4.0 re-plan (operator decision D9, 2026-10-06) re-executed every
load-bearing measurement in this tree at HEAD `69a085b2d` and reproduced
every value (plan.md §C carries the re-measurement table). The leader's
research note (`/private/tmp/claude-501/.../scratchpad/autonomy/autonomy.md`)
is a research artifact used only as a claim source — every load-bearing
claim was re-verified against the repository; corrections are recorded per
row.

Scope note (D9): the run-gate doc reconciliation (former work item 3) was
removed from this SPEC by operator decision. Rows 4-5 below record the
measured conflict text because §A still cites it as motivation; the files
are no longer edit targets, and their baseline greps were retired from
plan.md §C and acceptance.md with the scope cut.

## 1. Source-verification table

| # | Note claim | Verdict | Measured reality (file:line, this tree) |
|---|---|---|---|
| 1 | harness.yaml :75-78 = tier ceilings S1/M2/L3 | CONFIRMED | `plan_audit_tier_ceilings` at `.moai/config/sections/harness.yaml:75-78` — `S: 1`, `M: 2`, `L: 3`; re-verified at 69a085b2d |
| 2 | harness.yaml :80-84 = `plan_audit_ceiling_policy` (auto_delta_rounds 1, on_final_hit hold-and-split) | CONFIRMED, line-range corrected | The block sits at :82-84 (`plan_audit_ceiling_policy:` :82, `auto_delta_rounds: 1` :83, `on_final_hit: hold-and-split` :84); :80-81 are comment lines; re-verified at 69a085b2d |
| 3 | plan-auditor.md :690-714 Retry Loop Contract (ceiling + STOP on score regression) | CONFIRMED | `## Retry Loop Contract` heading at `.claude/agents/moai/plan-auditor.md:688`; ceiling prose :690-692; STOP escalation :708; ceiling policy + `auto_delta_rounds` + final hit :714 |
| 4 | phase-execution.md :99-113 still offers "Override and proceed" on FAIL | CONFIRMED (motivation only — no longer an edit target) | 4c (grace expired) at `.claude/skills/moai/workflows/run/phase-execution.md:99-106`. Baseline grep: `Override and proceed` = 1 hit @ `2f492df19`. The reconciliation of this text is follow-up card material (spec.md §E) |
| 5 | auto-semantics.md :200 says bypass flags retired | CONFIRMED (motivation only — no longer an edit target) | §9 row at `.claude/rules/moai/workflow/auto-semantics.md:200`. Also 4d INCONCLUSIVE offers "Proceed with acknowledgement" (:111-113) and "max 3 retries total" (:112). Follow-up card material |
| 6 | auto-semantics.md :154-160 fail-closed authority-gate invariant | CONFIRMED | §7: "a NEGATIVE or INCONCLUSIVE audit verdict is FAIL-CLOSED — the gate does not open on an unresolved audit" (:154-160) — the doctrine posture REQ-ACE-010 implements |
| 7 | §9 gate inventory lists ~10 of the note's 30 gates | CONFIRMED (observation only) | §9 (:190-201) carries 10 disposition rows. The v0.3.0 plan to add 11 rows was removed by the D9 scope cut; the inventory expansion is follow-up card material |
| 8 | spec-workflow.md already describes the ceiling policy in prose | CONFIRMED | `.claude/rules/moai/workflow/spec-workflow.md:158`: "The tier ceiling in `harness.plan_audit_tier_ceilings` is the only iteration cap; a ceiling hit follows `harness.plan_audit_ceiling_policy` … in every session" — SPEC constraint C1 binds to this text instead of duplicating it |
| 9 | (note B4) "the ceiling is prose in an agent body, not a counter that refuses" | CONFIRMED (mechanically) | `grep -rn "audit.round\|AuditRound\|iteration.count" internal/runtime/*.go` (non-test) = 0 hits (re-executed at 69a085b2d); `AuditResult` (`internal/runtime/audit_gate.go:54-95`) carries no round/counter field |
| 10 | (note B4) t1469/t1482: convergence fail + auditor PASS admitted | UNVERIFIABLE here | `.moai/reports/t1469/` and `.moai/reports/t1482/` do not exist in this tree. Recorded as operator-reported motivating instances only; the mechanical half is confirmed instead: the verdict-file format (`.moai/docs/audit-artifact-convention.md` § What) carries no convergence fields and `internal/auditverdict/verdict.go` parses none (grep "convergence\|receipt" = 0 hits, re-executed at 69a085b2d), so a required-backend fail cannot influence admission today |

## 2. Go surfaces measured (the counter and the receipt belong here)

- `internal/runtime/audit_gate.go` — `Verdict` enum :17-36 (PASS / FAIL /
  FAIL_WARNED / BYPASSED / INCONCLUSIVE; no PASS-WITH-DEBT value — the
  PASS-WITH-DEBT label lives only in `internal/auditverdict` and the
  convention doc); `GateConfig.Invoke` :199-297 (5-step protocol; Step 3 is
  the auditor spawn point a ceiling check must precede; the default branch
  :288-292 folds unknown verdicts into INCONCLUSIVE); `EnvSkipAudit` :48 and
  the bypass path :206-219 (the CLI-side bypass the §9 RETIRED row governs).
  Liveness (measured at f2f815008, re-measured at 69a085b2d, correcting the
  original presentation of `Invoke` as live infrastructure): `grep -rn
  "runtime\.GateConfig" internal/ cmd/ --include="*.go" | grep -v _test` → 0
  non-test matches (17 test-only) — `GateConfig` is test-only
  infrastructure, which is why the ceiling enforcement of design.md §2 lands
  at the LIVE seams (`decide.go`, `homestate`) and `Invoke` carries the same
  Step 0 only as a library-level consumer.
- `internal/runtime/audit_cache.go` — `ComputeHash` :121-141 (hash-only cache
  validity, SPEC-AUDIT-SNAPSHOT-001 A1); `planArtifactNames` :92-100 (the
  artifact subject set, includes `decision-index.md`).
- `internal/runtime/audit_review.go` — `ResolveLatestPlanAudit` :25-67
  enumerates the `<SPEC-ID>-review-<N>.md` iteration stream and "deliberately
  comes from the iteration stream … never from the date-stamped run history
  file" — the existing precedent for deriving state from iteration evidence.
- `internal/auditverdict/verdict.go` — THE one admission predicate ("the
  contract rule, the kickoff evaluator, and the card-transition guard all
  decide admission here", :1-4); `Parse` :69-129 (case-insensitive
  `key: value` lines, duplicate-key inadmissibility); `Admit` :195-233 (label,
  score threshold, must_pass_failed, blocking_count, hash binding,
  PASS-WITH-DEBT debt enumeration). No receipt/convergence fields exist.
- Enforcement call sites of the predicate:
  `internal/contract/kickoff/decide.go:373-376` (contract-mode kickoff),
  `internal/contract/rules.go:160` (label-only),
  `internal/homestate/card_evidence_readers.go:171-191` +
  `internal/homestate/card_audit_kickoff.go:28` (factory card transition).
- `internal/cli/mcp_convergence.go` — `ConvergenceResult` :115-194:
  `overall_verdict` ∈ {pass, fail}, `PerBackendVerdicts` with
  `Gate ∈ {off, advisory, required}`, `DisagreementFlag`, `AuditReceipt`;
  4-case policy :198-209 (any required FAIL ⇒ overall FAIL). The receipt
  fields of REQ-ACE-008 project this existing result into the verdict file —
  no new convergence semantics are invented. Single-model audits produce no
  `ConvergenceResult` (the deployed auditor contract runs its own review in
  single modes and calls no convergence tool) — the gap iter3 D32 named;
  REQ-ACE-008's single-model producer arm covers it (design.md §3).
- `internal/config/loader.go:344-346` — `plan_audit_global`,
  `plan_audit_tier_ceilings`, `plan_audit_ceiling_policy` acknowledged
  orphans; the latter two carry the comment "prose-consumed by the
  plan-auditor agent body; no Go reader" (retired by REQ-ACE-002; count = 2
  re-verified at 69a085b2d).
  Struct anchors: `internal/config/types.go:1350` (`PlanAuditGlobal`),
  `:1402` (per-level `PlanAudit`), `:1451` (`PlanAuditGlobalConfig`).
- `internal/cli/mcp_worktree_root.go:123-131` — `resolveAuditGates` folds a
  `ResolveAuditPlan` error into an empty gate set (`if err != nil { return
  config.AuditGates{}, "" }`); re-read and confirmed at 69a085b2d — the
  fail-open behavior REQ-ACE-010's third trigger arm corrects.
- `internal/runtime/audit_report.go:15` — the daily run-history writer still
  targets `.moai/reports/plan-audit/`, the directory the convention (§ Where)
  marks FORBIDDEN for verdicts (known divergence, card t1344 lesson). The
  counter of REQ-ACE-001 counts iteration evidence, never this file.

## 3. Template mirrors verified

This SPEC's remaining edit targets (v0.4.0) and their mirrors under
`internal/template/templates/`:

- `.claude/agents/moai/plan-auditor.md` (deployed 67532 bytes at
  measurement; mirror DIFFERS from deployed — pre-existing drift, Out of
  Scope; measured DIFF at f2f815008 and re-measured DIFF at 69a085b2d)
- `.moai/docs/audit-artifact-convention.md` (mirror identical; re-measured
  SAME at 69a085b2d)

Removed from the edit-target set by the D9 scope cut (their mirrors and
pre-existing drift are follow-up card material, no longer this SPEC's
concern): `.claude/skills/moai/workflows/run/phase-execution.md` (DIFF
re-measured at 69a085b2d), `.claude/rules/moai/workflow/auto-semantics.md`,
`.moai/config/sections/harness.yaml` (content unchanged by this SPEC — no
mirror step in any milestone).

## 4. Config decision (simplicity ladder)

No new config keys are required. The policy outcomes (pass-through /
debt-admit / scope-split / hold-record) are behavior defined by this SPEC's
REQ-ACE-004 through REQ-ACE-006 and REQ-ACE-013; `plan_audit_ceiling_policy`
keeps its two keys (`auto_delta_rounds`, `on_final_hit`) and gains only a Go
reader. Extending the YAML would add a second place the ladder could
disagree with the CLI.

## 5. Gaps

- t1469/t1482 incident details were not verifiable from artifacts in this
  tree (see table row 10); the SPEC cites them as operator-reported instances
  and enforces the mechanically-verified gap instead.
- The note's B2 30-gate table was not adopted; the v0.3.0 11-row selection
  was itself removed by the D9 scope cut. All inventory rows are follow-up
  card material.
- Exact `git rev-list` counts and usage-log figures quoted in the note (B3)
  were not re-measured — none are load-bearing for this SPEC.
- The v0.4.0 re-plan did not enumerate the live `.moai/reports/` tree to
  measure today's per-SPEC round counts (the same gap iter3 recorded for
  D31's population claim); §A's 35%-at-ceiling figure stays
  operator-measured from card t1500 intake.
