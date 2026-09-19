# SPEC-ROSTER-NUMERAL-AXIS-001 — Progress

card t930 · branch `WT-numeral-roster-guard` · base `690dfe369`

## §E.1 Plan-phase Audit-Ready Signal

- Artifacts authored: `spec.md`, `plan.md`, `acceptance.md`, `progress.md` (Tier M, Class C).
- Requirements: REQ-RNA-001 … REQ-RNA-013 (GEARS). Acceptance: AC-RNA-001 … AC-RNA-015
  (15 criteria, under the Tier M ceiling of 16).
- Decisions D1 (noun class) and D2 (discharge rule) are ADOPTED and recorded in `spec.md` §D
  with their rejected alternatives and measured cost. Both are **operator-unanswered** — put to
  the operator, no answer inside the window — so they are the lead's / this SPEC's judgment and
  stay reviewable at the Implementation Kickoff Approval gate.
- **D3 (mirror derivation) — WITHDRAWN by operator decision, 2026-09-18.** Kept in `spec.md` §D
  as a recorded rejected alternative with its four findings (wrong premise measured at 13 rows /
  33 result rather than 20 / 26; derivation blinds the guard to sibling divergence;
  `readmeSite()` precedent overstated; withdrawal shrinks the SPEC). M5 stands at 46 authored
  rows with no folding.
- Plan-audit iteration 1 FAIL (0.775) → revision v0.2.0 cleared D1-D12. Iteration 2 FAIL (0.825,
  score clears; verdict rested on an introduced Must-vs-Must contradiction plus the partially
  resolved D3) → revision v0.3.0, authorised by the operator as a third iteration past the Tier
  M cap of two. Verdicts at `.moai/reports/t930/plan-audit.md` and
  `.moai/reports/t930/plan-audit-iter2.md` — both gitignored by operator directive, never
  committed.
- Cost arithmetic (measured by the dispatching lead in this tree at HEAD `6abcc85fa`): 63 hit
  paths, 20 registry paths carrying `ClaimCount`, 17 hits discharged, **46 residual**, 1 extra
  path the rejected any-row rule would free. Mirror-fold measurement at HEAD `d8f140b25`: 17 of
  the 46 under `internal/template/templates/`, 14 with a local twin on disk, **13 foldable
  pairs** → 33 rows, which is what withdrew D3.
- The 16-vs-17 `ClaimCount` disagreement between the auditor's pass-1 parse and the lead's
  figures is SETTLED in favour of the lead: a text-level `Path:`/`Claims:` parse cannot see the
  four rows `readmeSite()` builds from a `path` parameter. AC-RNA-013 requires the run-phase
  re-derivation to read `Registry()` rather than the file for this reason.

### Implementation Kickoff Approval

- **Given 2026-09-18 by the operator, CONDITIONAL on this SPEC reaching plan-audit PASS.**
- Recorded here so run-phase entry rests on a written approval rather than a remembered one.
- **Condition met 2026-09-18.** The confirmation read over revision v0.3.0 returned four of the
  lead's five axes clean and one minor finding (`plan.md` citing a stale `AC-RNA-014` after the
  16 → 15 renumbering), which was verified and repaired in `89a4831e5` together with the
  unlabelled 17-vs-20 `ClaimCount` pair in §A. The SPEC is PASS; run-phase entry rests on this
  written approval and that recorded condition.
- One criterion needed repair AFTER the audit, for a reason the audit could not have seen:
  `AC-RNA-010` anchored `$IMPL` to the `internal/harness/rosterguard/` directory, which resolves
  to card t922's own commit `bdaafe6fe` once develop is absorbed. Re-anchored to this card's own
  new file in `fd17f4bb0`. The audit was correct at `c9a1e3e1e`; the tree moved underneath it.
- Plan-phase measurements attributed in `spec.md` §A: population figures supplied by the
  dispatching lead (this tree, base `690dfe369`); independently re-measured here —
  `go test -count=1 ./internal/harness/rosterguard/...` → `ok … 1.096s`,
  `profileMatrixAgentOrder` = 13 names, `registry.go` = 30 `Site` rows.
- Status: `draft`.

## §E.2 Run-phase Evidence

### Run-phase baseline — measured BEFORE any implementation edit

Tree: `29a4266f3` on `WT-numeral-roster-guard`, after absorbing develop (`0 8` against it,
working tree clean). The absorption was done first ON PURPOSE: a baseline measured on a
pre-absorption tree would not be a baseline for the tree the implementation actually lands on.

Green-before-edit, scoped to the affected package:

```
go build ./internal/harness/rosterguard/...            exit 0
go vet   ./internal/harness/rosterguard/...            exit 0
go test -count=1 ./internal/harness/rosterguard/...    ok  4.287s
```

`internal/harness/rosterguard/` currently holds `axis.go`, `check.go`, `registry.go`,
`root_test.go`, `rosterguard_test.go`. **`numeral.go` does not exist**, so `AC-RNA-010`'s
`$IMPL` is currently empty — which is the criterion failing, as designed, until this card
creates that file.

### Population re-derived in-run (REQ-RNA-012, AC-RNA-013)

```
hit paths                       : 64
registry paths carrying Count   : 20
hits discharged by a Count row  : 17
hits needing a NEW row/exempt   : 47
extra paths the any-row rule would free : 1   (internal/web/agentfm.go)
```

**The residual is 47, not the 46 carried in `spec.md` §A.** The plan-phase figure was measured
at `6abcc85fa`; absorbing develop brought in one further hit path,
`.claude/skills/moai-foundation-quality/references/reference.md` — the local twin of a template
file that was already residual, having gained the same claim during those commits.

This is the obligation working rather than an error in either number: a plan-phase figure is not
a run baseline, and had the run reused 46 it would have under-registered by one row and the
layer would have reported an undeclared hit. `spec.md` §A stays at 46 as the attributed
plan-phase figure; 47 is the run baseline, measured here, on this tree.

The residual's shape is unchanged: 47 paths, of which 17 sit under
`internal/template/templates/` as mirrors, and the rest split between genuinely stale claims
(the `11 retained agents` / `11-agent catalog` family), legitimate historical citations (the
`then-8-agent catalog` family, `17->8` test comments, dated research), measured false positives
(`172); MoAI retained agents`, `05-25), the agent catalog`, `001): agent catalog`), and
rosterguard's own files describing themselves.

### Scope adjudications

Recorded at `.moai/reports/t930/adjudication.md` (disk-only per the operator directive in
`.gitignore:225-229`): (a) roster drift living as prose in agent definition files is already
reached by the adopted noun class — measured, no widening made; (b) the tier axis in
`advanced/no-haiku-3tier.md` is out of scope on two independent grounds — `docs-site/` is
outside the swept tree, and all four locales carry zero roster-noun hits because the claims are
about tiers, which is cellguard's subject matter.

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

🗿 MoAI
