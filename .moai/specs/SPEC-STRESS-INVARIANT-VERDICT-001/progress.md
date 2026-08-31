# SPEC-STRESS-INVARIANT-VERDICT-001 — Progress

Card t372 · worktree `.claude/worktrees/t372` · branch `WT-stress-invariant-guard` ·
base `origin/develop` = `b9149857c`.

## §E.1 Plan-phase Audit-Ready Signal

### Iteration 1 — audited FAIL 0.69 (`.moai/reports/t372/plan-audit.md`)

- Artifacts written: `spec.md`, `plan.md`, `acceptance.md`, `progress.md`. Declared `tier: S`.
- Requirements REQ-SIV-001 .. REQ-SIV-017 (17); acceptance AC-SIV-001 .. AC-SIV-013 (13).
- All 7 must-pass checks passed; the FAIL was score-driven (Clarity 0.65, Completeness 0.75,
  Testability 0.55, Traceability 0.90).

### Iteration 2 — fix round (this revision, `version: 0.2.0`)

Tier reclassified **S → M**. Justification (`spec.md` § Tier classification): the artifact carries
16 REQ / 14 AC, which fits Tier M's 16/16 ceiling and not Tier S's 8/8; tiering up rather than
splitting is the rule's stated response to a budget breach, and the higher threshold (0.80) plus the
2-iteration ceiling are accepted deliberately on a card whose hazard is that concealment resembles
repair. The Tier M artifact set is exactly what exists, so `acceptance.md` is no longer a deviation.

Defects answered, all against source read on tree `9d4f79281`:

| Defect | Change |
|---|---|
| D1 | AC-SIV-009's duplicate-id branch deleted, with its exclusion reason (`backlog_sqlite.go` `id … UNIQUE`) recorded. RED must now originate at a named assertion **inside** the invariant block, citing message + source line. Four non-discharging RED sources enumerated. |
| D2 | REQ-SIV-008 added: `successes + starved + hardFailures == stressWriters * stressAddsPerWriter`. AC-SIV-014 added. Machine-independent by construction; fractional floors explicitly forbidden. |
| D3 | AC-SIV-009 branch (c) restated as a `last_seq` advance **above** the item count, with the downward direction excluded and `normalizeBacklogRecord`'s erasure recorded. |
| D4 | Tier S → M (above). REQ layer consolidated 17 → 16 by four merges; AC-SIV-008 / AC-SIV-009 survive intact as separate binding criteria. |
| D5 | REQ-SIV-009 reworded *covers* → *coherent with … at the declared per-mutation cost*; the guard's own messages must state the same and claim no sufficiency. |
| D6 | REQ-SIV-014 clause 4 added: the Unix sentinel admits every `unix.Flock` failure, `errors.Is` traverses `errors.Join`. Sentinel narrowing recorded as an out-of-scope follow-up candidate. |
| D7 | AC-SIV-001 / AC-SIV-005 given a deterministic seeded-holder construction (`acquireBoardLockImpl` + `t.Cleanup`, 1-2 adds) and a named verification verb. M2 steps 4-5 carry the implementation shape. |
| D8 | AC-SIV-013 rationale + REQ-SIV-014 clause 3: a green window evidences only that no new failure mode was introduced, because the invariant criterion was already red in 0 of 14 runs. |
| D9 | REQ-SIV-014 clause 2 corrected: one green `Race Test` job (`51daada00`) and one run green inside a job reddened by another test (`c6aa61346`). |
| D10 | REQ-SIV-004 relabelled `(Where)`; the old REQ-SIV-002's second obligation split into its own `(Unwanted)` requirement. |
| D11 | REQ-SIV-016 added for scope discipline; AC-SIV-012 now traces to it rather than to a section. |
| D12 | `plan.md` §B states the 4.2% margin and the catch-set table, plus the cost-independent reduction `supportedWriters * headroom >= stressWriters * stressAddsPerWriter`. |
| D13 | AC-SIV-008 requires the old guard's GREEN under the same mutant, `-v` output evidencing a non-zero selector match, and the RED naming which test failed. |

Operator observation folded in: REQ-SIV-010 forbids rebuilding the guard's floor from the budget's
own inputs, citing the pre-existing unreachable `budget < floor` branch in
`TestBoardLockWaitBudgetDerivedFromNamedInputs` as the shape not to reproduce. Repairing that branch
is recorded as out of scope.

- SPEC ID regex check executed as Bash: `PASS`.
- Requirements: REQ-SIV-001 .. REQ-SIV-016 (16), GEARS notation, no residual `IF/THEN`.
- Acceptance: AC-SIV-001 .. AC-SIV-014 (14); AC-SIV-008 / AC-SIV-009 remain the binding
  mutant-evidence pair (both directions required).
- Ground truth consumed, not re-measured: `.moai/reports/t370/verdict.md`,
  `.moai/reports/t370/measurements.md`.
- No implementation code written at plan-phase. No push, no PR, no CI run, no load generated.

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
