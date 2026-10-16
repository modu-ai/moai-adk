# decision-index.md — SPEC-RELEASE-PIPELINE-001

> Stateless on the status axis (no `status:` field; the SPEC's lifecycle lives in spec.md).
> Authored because `.moai/config/sections/interview.yaml` carries `decision_gate: on`.
> One row per decision surfaced during assembly that the operator did not settle.
> Labels are exactly: DECIDED / POLICY-COVERED / EVIDENCE-NEEDED / FOUNDER.

### Q1: Card item 1 is already landed on main under this card id (PR #1770 + t1478 follow-ups) — does M1 remain in this SPEC as a verify-and-preserve milestone, or is item 1 closed as delivered and the SPEC re-scoped to items 2-5?

Label: FOUNDER
Authority anchor: none — no register entry (prior completed SPEC HISTORY, config setting, or constitution clause) covers re-anchoring a card item found landed at plan phase.
Why unresolved: the 10-06 card text predates the landing; the dispatch still names five items, and only the operator can retire a card item.
Class: implementation-level
Default: keep M1 as verify-and-preserve (rule: preserves current behavior — the card's five-milestone structure and the landed surface both stay)
Alternate: close item 1 as delivered and re-scope the SPEC to items 2-5
Operator verdict: DEFAULT-APPLIED 2026-10-16T19:17:57Z manager-spec/plan-phase

### Q2: Does input-class splitting narrow the release 3-OS matrix for harness-only release diffs (e.g. ubuntu-only), or does the 3-OS release coverage stay keyed to the shared go_code filter regardless of class?

Label: FOUNDER
Authority anchor: none — the bias-to-run doctrine lives in landed workflow comments and the filter file, not in a register-admitted artifact.
Why unresolved: the card's phrase "adjust 3-OS harness scope" is ambiguous between narrowing the matrix and keeping coverage while splitting ci.yml cost jobs.
Class: implementation-level
Default: no narrowing — harness roots stay inside go_code and the release matrix keeps 3-OS coverage for every test-input class (rule: preserves current behavior)
Alternate: narrow the release matrix to ubuntu-only legs for harness-class-only release diffs
Operator verdict: DEFAULT-APPLIED 2026-10-16T19:17:57Z manager-spec/plan-phase

### Q3: For the shared drafter/cleanup concurrency group, does an incoming run cancel an in-flight run, or queue?

Label: FOUNDER
Authority anchor: none — no register entry covers cross-workflow concurrency semantics for the draft list.
Why unresolved: cancelling is cheaper for redundant drafter updates but cancels a destructive cleanup mid-deletion; the card names the shared group without choosing the cancellation semantics.
Class: implementation-level
Default: queue, never cancel, on the destructive side — cleanup is never cancelled mid-deletion (rule: smaller user-visible surface — no new cancellation behaviors introduced)
Alternate: shared group with cancel-on-arrival (incoming run cancels whatever is in flight)
Operator verdict: DEFAULT-APPLIED 2026-10-16T19:17:57Z manager-spec/plan-phase

### Q4: When the cost jobs are input-gated, do their per-job required check names get replaced by the single aggregate, or retained alongside an extended intentional-skip marker?

Label: FOUNDER
Authority anchor: none — the release-pr-gate aggregate precedent (SPEC-V3R4-CI-FASTTRACK-001 lineage) answers a similar question under different conditions (release PRs, not ci.yml cost jobs), so it is not an identical-conditions DECIDED anchor.
Why unresolved: both surfaces keep branch protection satisfied; they differ in PR-check visibility and in how much branch-protection config changes.
Class: implementation-level
Default: single aggregate required name for the gated family (rule: smaller user-visible surface — the preserve-behavior and single-revert tests do not discriminate because every option changes the always-run behavior this card exists to change)
Alternate: retain per-job required names and extend the skip-marker family to the newly gated jobs
Operator verdict: DEFAULT-APPLIED 2026-10-16T19:17:57Z manager-spec/plan-phase

### Q5: Does the browser job's green/mutation split take the form of two jobs, or one job with separately bounded stages?

Label: FOUNDER
Authority anchor: none — SPEC-APPJS-FIRE-GUARD-001's ADDITIVE-ONLY clause bound its own card, not this split's shape.
Why unresolved: both shapes satisfy REQ-RPC-007/008; they differ in runner cost (second Chrome provisioning) and in the required-check surface ("Test (browser fire guard)" name).
Class: implementation-level
Default: one job with per-stage ceilings and explicit upload headroom (rule: preserves current behavior — job structure and check name stay; the split is internal)
Alternate: separate green and mutation jobs, each with its own budget and Chrome provisioning
Operator verdict: DEFAULT-APPLIED 2026-10-16T19:17:57Z manager-spec/plan-phase

### Q6: Are registry-gated retries extended to the macOS/Windows release race legs?

Label: EVIDENCE-NEEDED
Authority anchor: none — the flaky registry (`scripts/ci/flaky-registry.txt`) carries ubuntu-CI-measured signatures; no macOS/Windows flaky measurements exist.
Why unresolved: extending the retry wrapper to the 3-OS release legs requires per-platform transient-failure signatures that have not been measured; a registry without platform-matched entries would either retry semantic failures or silently match nothing.
Operator verdict:
