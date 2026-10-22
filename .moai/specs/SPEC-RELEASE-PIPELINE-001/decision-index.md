# decision-index.md — SPEC-RELEASE-PIPELINE-001

> Stateless on the status axis (no `status:` field; the SPEC's lifecycle lives in spec.md).
> Authored because `.moai/config/sections/interview.yaml` carries `decision_gate: on`.
> One row per decision surfaced during assembly that the operator did not settle.
> Labels are exactly: DECIDED / POLICY-COVERED / EVIDENCE-NEEDED / FOUNDER.

> Scope note (v0.4.0, 2026-10-22): the hold-and-split decision itself — M3 + D11 out to a follow-up card; M2/M4/M5 (+D10) retained — was settled by the operator (relayed by leader lane-17, recorded in `.moai/reports/t1536/progress.md` § Split 결정). Because it is settled, it is not a row here, and that record is a card-local note, not a committed authority anchor. Rows Q1-Q6 carry over; Q4 is re-opened because its premise changed (card t1858); Q7 and Q8 are new. Q9 is added in v0.5.0 (plan-audit iter-4 D13 offered routing the ceiling value to a decision row).

### Q1: Card item 1 is already landed on main under this card id (PR #1770 + t1478 follow-ups) — does M1 remain in this SPEC as a verify-and-preserve milestone, or is item 1 closed as delivered and the SPEC re-scoped to the open items?

Label: FOUNDER
Authority anchor: none — no register entry covers re-anchoring a card item found landed at plan phase.
Why unresolved: the 10-06 card text predates the landing; only the operator can retire a card item. (The v0.4.0 split narrowed the open items to 2/4/5 but did not retire item 1.)
Class: implementation-level
Default: keep M1 as verify-and-preserve (rule: preserves current behavior — the landed surface stays guarded)
Alternate: close item 1 as delivered and re-scope the SPEC to items 2, 4, 5
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
Why unresolved: cancelling is cheaper for redundant drafter updates but cancels a destructive cleanup mid-deletion.
Class: implementation-level
Default: queue, never cancel — cleanup is never cancelled mid-deletion (rule: smaller user-visible surface — no new cancellation behaviors introduced)
Alternate: shared group with cancel-on-arrival
Operator verdict: DEFAULT-APPLIED 2026-10-16T19:17:57Z manager-spec/plan-phase

> Audit repair note (iter1 D3, 2026-10-17): extended, not reversed — `cancel-in-progress: false` on BOTH workflows (the incoming run's own flag governs). Label/Class/default direction unchanged.

> Audit repair note (iter2 D8, 2026-10-17): extended again — the group key is a ref-independent literal name on BOTH sides. Label/Class/default direction unchanged.

### Q4: When the cost jobs are input-gated, are their required check names (`Lint`, the five `Build (…)`) retained and published by skip-markers on a positively observed non-match, or replaced by one aggregate required name?

Label: FOUNDER
Authority anchor: none — the release-pr-gate aggregate precedent answers a similar question under different conditions (release PRs), so it is not an identical-conditions DECIDED anchor.
Why unresolved: both surfaces can keep every PR mergeable; they differ in which required-name sets must change. Since card t1858, `.github/required-checks.yml` carries a `release_gate` list (`Lint`, `Test (ubuntu-latest)`, `Build (linux/amd64)`) that `scripts/release.sh` demands on the tagged main commit and that `internal/config/required_checks_release_gate_test.go` holds to be a subset of `branches.main.contexts`; the `ci/**` candidate set names the same checks. An aggregate therefore means rewriting three name sets plus a release-script input; retained names mean none.
Class: implementation-level
Default: retain the per-job required names; publish each on a positively observed non-match through a skip-marker job in the `test-skip-marker` pattern (rule: preserves current behavior — every required-name set and the `release_gate` subset invariant stay unchanged, required-checks.yml unedited)
Alternate: single aggregate required name for the gated family, with `branches.main.contexts`, `branches."ci/**".contexts`, and `release_gate` rewritten in the same commit and the aggregate publishing on push to main
Operator verdict: DEFAULT-APPLIED 2026-10-22T12:26:32Z manager-spec/plan-phase

> Re-anchor note (v0.4.0, 2026-10-22): the prior verdict (`DEFAULT-APPLIED 2026-10-16T19:17:57Z`, default = single aggregate under the "smaller user-visible surface" rule, since no option then preserved behavior) is superseded. Its premise — that every option changes current behavior — no longer holds once `release_gate` landed: retaining the names now preserves every required-name set, so the Default rule's first clause selects it. Direction reversed; Label and Class unchanged. Recorded per `verification-completeness.md` §3.

### Q5: Does the browser job's green/mutation split take the form of two jobs, or one job with separately bounded stages?

Label: FOUNDER
Authority anchor: none — SPEC-APPJS-FIRE-GUARD-001's ADDITIVE-ONLY clause bound its own card, not this split's shape.
Why unresolved: both shapes satisfy REQ-RPC-007/008; they differ in runner cost (second Chrome provisioning) and in the check surface ("Test (browser fire guard)" name).
Class: implementation-level
Default: one job with per-stage ceilings and explicit upload headroom (rule: preserves current behavior — job structure and check name stay)
Alternate: separate green and mutation jobs, each with its own budget and Chrome provisioning
Operator verdict: DEFAULT-APPLIED 2026-10-16T19:17:57Z manager-spec/plan-phase

### Q6: Are registry-gated retries extended to the macOS/Windows release race legs?

Label: EVIDENCE-NEEDED
Authority anchor: none — the flaky registry (`scripts/ci/flaky-registry.txt`) carries ubuntu-CI-measured signatures; no macOS/Windows flaky measurements exist.
Why unresolved: extending the retry wrapper to the 3-OS release legs requires per-platform transient-failure signatures that have not been measured. (SPEC-CI-RELEASE-VERIFY-SHARD-001, completed, also kept the release legs wrapper-free; that is consistent with, not a substitute for, the missing measurement.)
Operator verdict:

### Q7: Which fetch source supplies the cleanup filter's target-branch field, now that `gh release list` rejects `targetCommitish` (D10)?

Label: FOUNDER
Authority anchor: none — no register entry covers the cleanup workflow's fetch source.
Why unresolved: three shapes were considered. (a) list with supported fields, then per-candidate `gh release view <tag> --json …,targetCommitish` (EV-12: supported) — camelCase end to end, one extra call per candidate, the same call REQ-RPC-011's re-query needs; (b) `gh api repos/{owner}/{repo}/releases` — one call, snake_case payload (EV-13) requiring an explicit mapping step before the filter; (c) drop `targetCommitish` — removes the `release/*` exemption the workflow header (release-drafter-cleanup.yml:13) states as a safety constraint, so REQ-RPC-013 excludes it and it is not an option.
Class: implementation-level
Default: (a) list + per-candidate `gh release view` (rule: smaller user-visible surface — no change of payload shape and no mapping layer; the first two rule clauses do not discriminate, since the current fetch fails and either repair is a single revert)
Alternate: (b) REST listing with explicit snake_case→camelCase mapping
Operator verdict: DEFAULT-APPLIED 2026-10-22T12:26:32Z manager-spec/plan-phase

### Q8: Does input-class gating apply to pull_request events only, or also to push events (main, `ci/**` candidates)?

Label: FOUNDER
Authority anchor: none — no register entry decides cost-job behavior per event class.
Why unresolved: ci.yml runs on pull_request, push to main, candidate push to `ci/**`, and workflow_dispatch (ci.yml:16-31). The card names PR cost; push events feed the `release_gate` check `scripts/release.sh` reads on the tagged main commit and the candidate verdict read from `branches."ci/**".contexts`. Gating pushes saves cost there too, but replaces real Lint/Build runs on those commits with markers.
Class: implementation-level
Default: gate pull_request events only; push and workflow_dispatch keep the always-run behavior (rule: preserves current behavior — the tagged main commit and every candidate keep real Lint/Build runs)
Alternate: gate push events as well, publishing the required names through markers on a positively observed non-match
Operator verdict: DEFAULT-APPLIED 2026-10-22T12:26:32Z manager-spec/plan-phase

### Q9: Does this SPEC place a step-level ceiling on the release race step, now that no ceiling value satisfies both "never cuts a currently-green leg" and "leaves pre-step plus evidence headroom inside the 40-minute job" (plan-audit iter-4 D13)?

Label: FOUNDER
Authority anchor: none — SPEC-CI-RELEASE-VERIFY-SHARD-001 fixed `-timeout 35m` and job `timeout-minutes: 40` as not-to-be-raised, but no register entry decides a race-step ceiling.
Why unresolved: measured in dispatched run 39729541531 (acceptance.md EV-15, n=1): the windows cli-1 race step ran 37m28s and cli-3 36m42s, while the windows cli-2 leg spent 4m21s before its race step. A ceiling that keeps those green legs green needs ≥ 38 minutes; a ceiling that leaves the measured pre-step (rounded up to 5) plus a 1-minute evidence allowance needs ≤ 34. The interval is empty, so every option trades one property for the other. Run-to-run variance is unmeasured.
Class: implementation-level
Default: place no step ceiling on the release race step in this SPEC; the release half of REQ-RPC-007 is held by the always()-evidence-step and not-raised invariants only (rule: preserves current behavior — no currently-green leg changes verdict)
Alternate: place a 34-minute ceiling on the race step (40 − 5 − 1), accepting that legs measured above 34 minutes (windows cli-1, cli-3) turn red until their shards are rebalanced under SPEC-CI-RELEASE-VERIFY-SHARD-001's ownership
Operator verdict: DEFAULT-APPLIED 2026-10-22T12:54:18Z manager-spec/plan-phase
