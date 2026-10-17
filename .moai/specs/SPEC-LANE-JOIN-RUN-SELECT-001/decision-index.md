# SPEC-LANE-JOIN-RUN-SELECT-001 — Decision Index

Decision gate: `interview.decision_gate: on` (`.moai/config/sections/interview.yaml`, line 6, read at tree `fb749d372e61`). Each row below was surfaced while assembling the SPEC. No row carries a recommendation for a judgment call; a `Default:` line is the published Default rule applied (preserves current behavior; else single-revert undo; else smaller user-visible surface). The artifact is stateless on the status axis; the SPEC lifecycle is in `spec.md`.

Authority-register search performed: `.moai/specs/*/` HISTORY and Amendments for `AmbiguousRunsError`, `ResolveActiveRun`, `ReconcileActiveRuns` and `chooser` returned no completed SPEC deciding the row below; `.moai/config/sections/*.yaml` carries no setting covering it. The decisions already closed by the predecessor's audits (cancel promise, re-read source, regression-guard labeling) live in a sibling card tree that is not part of the committed authority register, so they are carried in `spec.md` §F as closed scope and are not rows here. The card itself fixes the A-part boundary, the interactive-only chooser and the N2 shape (re-query through a read path); those are dispatch-fixed scope (`spec.md` §E), not rows.

### Q1: Does a re-presentation of the chooser re-run the automatic reconciler, or only read the store?

Label: FOUNDER
Class: implementation-level
Authority anchor: none (no committed artifact decides this)
Why unresolved: the gate already runs the reconciler once, inside `ResolveActiveRun`, before the first presentation. A re-read that goes back through `ResolveActiveRun` with an empty selector would run it again (a write that can retire further records-free dead runs mid-prompt) and, for exactly one active run, would also run the environment-setting join (the N2 defect). A re-read that only reads shows such a run, if one appears, as a dead zero-card row with `cleanup then join` offered, leaving the retirement to the operator's choice.
Default: read only; the re-presentation runs no reconciler (rule: the smaller user-visible surface — the chooser performs no write the operator did not choose)
Alternate: run the reconciler through a separate call that does not set the environment variable, before each re-read
Operator verdict: DEFAULT-APPLIED 2026-10-17T14:53:06Z manager-spec+plan-author
