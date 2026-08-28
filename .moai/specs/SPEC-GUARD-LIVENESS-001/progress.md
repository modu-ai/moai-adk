# SPEC-GUARD-LIVENESS-001 — Progress (card t333, surfacing model)

## §E.1 Plan-phase Audit-Ready Signal

Plan-phase artifacts authored on `091966c55` @ `WT-guard-liveness` (worktree `.claude/worktrees/t333`).

- Artifacts: `spec.md`, `plan.md`, `acceptance.md` (Tier M set) + this file.
- Requirements: 13 (Tier M ceiling 16). Acceptance criteria: 13 (ceiling 16). Both under, which is the outcome of the scope reduction.
- **Two baselines.** RED-now cells pin `091966c55`; card t326 citations pin `origin/develop` at `ec15ec2cd`, a diverged tree (diverged, `merge-base --is-ancestor` false). Each t326 citation names its tree inline — reading the baseline for a t326 surface reports a landed feature as absent (`spec.md` §A.10).
- Primary evidence artifact: `.moai/reports/t333/trigger-axis-observation.md` (tracked at `c30f761dd`).
- Every RED-now cell is pinned to `091966c55` and its command was re-run during the split; no cell was carried across without re-measurement.

### Audit history

| Iteration | Verdict | Score | Outcome |
|---|---|---|---|
| 1 | PASS-WITH-DEBT | 0.800 | 7 blocking defects (D1-D7) closed at v0.5.0 |
| 2 | PASS-WITH-DEBT | 0.800 (flat) | N1, N2, N5 closed at v0.6.0; Traceability 0.75→1.00, Completeness 1.00→0.75 |
| 3 | **FAIL + STOP** | 0.667 | Regression clause fired. Operator chose **scope reduction**, the audit's own recommendation. No fourth repair round. |
| 5 (new cycle, iter 2) | **PASS-WITH-DEBT** | 0.8625 | Monotonic (0.75 → 0.8625), no STOP. **Fit to implement after three localized text repairs**, none touching the design or the milestone map; Traceability and Completeness verified not to have degraded under the prior repair. D9, D10, D11 closed at v1.4.0, plus two further stale twins found by sweeping rather than by audit. |
| 4 (new cycle, iter 1) | **FAIL** | 0.75 | No STOP (0.75 > 0.667). Traceability and Completeness graded excellent; the failure was concentrated in two criticals. All six blocking (D1-D6) plus both optional (D7, D8) closed at v1.3.0. T9's second layer — the "invoked unconditionally, evaluates conditionally" mutant — closed by moving AC-GDL-003's counts to the query layer. |

**The split.** The state model moved to `SPEC-GUARD-STATE-MODEL-001`, **card t347**. This SPEC keeps the surfacing model, which converged: the auditor re-ran both prior N1 mutants and could not revive either. The seam is a consumed contract, not a `depends_on` (`spec.md` §B.1), so **this artifact set is finishable on its own** — nothing in it requires a t347 artifact to exist, and it goes to the Implementation Kickoff Approval gate independently of t347's dispatch.

**§A.9 (instance 7) added at v1.1.0** — a verdict produced, correct, red, and never collected, by the orchestrator running this card. It is the live case for §A.8 and the evidence that discipline without a mechanism is insufficient.

**Iter-3 findings disposition:** T9 closed here (AC-GDL-003, the criterion the audit named most consequential). T3 dissolved structurally by the contract restatement rather than corrected as two sentences. T7, T10, T11 folded into `spec.md` §D.3. N4 taken in AC-GDL-006. T2, T4, T1, T5, T8 travel with the state model as its starting material.

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
