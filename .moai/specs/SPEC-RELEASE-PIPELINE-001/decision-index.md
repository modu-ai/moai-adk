# decision-index.md — SPEC-RELEASE-PIPELINE-001

> Stateless on the status axis (no `status:` field; the SPEC's lifecycle lives in spec.md).
> Authored because `.moai/config/sections/interview.yaml` carries `decision_gate: on`.
> One row per decision surfaced during assembly that the operator did not settle.
> Labels are exactly: DECIDED / POLICY-COVERED / EVIDENCE-NEEDED / FOUNDER.

> Scope note (v0.4.0, 2026-10-22): the hold-and-split decision itself — M3 + D11 out to a follow-up card; M2/M4/M5 (+D10) retained — was settled by the operator (relayed by leader lane-17, recorded in `.moai/reports/t1536/progress.md` § Split 결정). Because it is settled, it is not a row here, and that record is a card-local note, not a committed authority anchor. Rows Q1-Q6 carry over; Q4 is re-opened because its premise changed (card t1858); Q7 and Q8 are new. Q9 is added in v0.5.0 (plan-audit iter-4 D13 offered routing the ceiling value to a decision row). Q10 is added in v0.8.0 to record the operator's relayed verdict on sync-audit iter-3 F3 (in-place amendment).

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

> Supersession note (v0.8.0, 2026-10-23): this row's premise — one shared drafter/cleanup group — is superseded by Q10 (separate groups). Its applied answer "never cancel" survives: both sides keep `cancel-in-progress: false`. The row is kept as recorded.

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

### Q9: Does this SPEC place a step-level ceiling on the release race step, now that the "never pre-empts `-timeout 35m`" claim is withdrawn (plan-audit iter-4 D13) and the green-leg race-step duration a ceiling would have to admit is unmeasured (plan-audit iter-5 D20)?

Label: FOUNDER
Authority anchor: none — SPEC-CI-RELEASE-VERIFY-SHARD-001 fixed `-timeout 35m` and job `timeout-minutes: 40` as not-to-be-raised, but no register entry decides a race-step ceiling.
Why unresolved: the only measurement is dispatched run 39729541531 (acceptance.md EV-15) — a FAILING run (`conclusion: failure`, `event: workflow_dispatch`, branch `WT-cloud-windows-portability-repair`, not main), n=1, windows legs only, in which cli-1, cli-2, cli-3, cli-4 and remainder all concluded `failure`, race step included. Its race-step durations (cli-1 37m28s, cli-3 36m42s) are durations of failing legs, so they bound nothing about how long a green leg's race step runs, and no lower bound for a ceiling that would leave green legs green is established. Its pre-step duration (cli-2 4m21s, rounded up to 5) precedes the tests and is therefore outcome-independent, so a ceiling that leaves that pre-step allowance plus a 1-minute evidence allowance inside the 40-minute job needs ≤ 34 — itself an n=1, windows, non-main-branch figure. Whether a ceiling ≤ 34 would cut any currently-green leg is unknown. Assumption (unmeasured, not a requirement of this SPEC): a green-run re-measurement — `gh run list --workflow release-pr-multi-os.yml --status success`, then the EV-15 jq on a listed run — is the follow-up that would establish the green-leg bound. Run-to-run variance is unmeasured.
Class: implementation-level
Default: place no step ceiling on the release race step in this SPEC; the release half of REQ-RPC-007 is held by the always()-evidence-step and not-raised invariants only (rule: preserves current behavior — no step ceiling is added; the unmeasured green-leg duration is not claimed)
Alternate: place a 34-minute ceiling on the race step (40 − 5 − 1), accepting that any leg whose race step exceeds 34 minutes is cut and turns red until its shard is rebalanced under SPEC-CI-RELEASE-VERIFY-SHARD-001's ownership; in the one measured run (failing run 39729541531) the windows cli-1 and cli-3 race steps exceeded 34 minutes, and whether green legs do is unmeasured
Operator verdict: DEFAULT-APPLIED 2026-10-22T12:54:18Z manager-spec/plan-phase

### Q10: Do the Release Drafter workflow and the drafter-cleanup workflow share one concurrency group?

Label: FOUNDER
Authority anchor: none — no register entry decides cross-workflow concurrency for the draft list; the verdict below is an operator verdict relayed by the leader session, recorded in this SPEC's progress.md §G, and it is not cited as a committed register anchor.
Why unresolved (at the time of the sync-audit iter-3 finding F3): GitHub keeps at most one running and one pending run per concurrency group, and a newer pending run replaces the older pending one. Under the v0.7.x shared group, a newer drafter run replaces a pending weekly cleanup, so that week's cleanup is lost (`.moai/reports/t1536/sync-audit-iter3.md` F3). Keeping the shared group preserves current behavior and accepts the loss; separating the groups lets the two workflows run at the same time, so the cleanup's safety must come from somewhere other than the group.
Class: implementation-level
Default: (b) keep the shared group and accept the pending-run replacement as debt (rule: preserves current behavior)
Alternate: (a) separate the drafter and cleanup groups; correctness via the M5 re-query before each delete plus an idempotent delete
Operator verdict: OPERATOR-DECIDED 2026-10-23 option (a) — separate groups (relayed by the leader session to lane-9 and recorded as relayed; the lane did not observe the operator directly; not DEFAULT-APPLIED). Rationale as relayed: pending-run replacement in a shared group loses the weekly cleanup; correctness is carried by the M5 re-query before delete plus idempotent delete, so no concurrency group can drop a pending cleanup. Applied in spec.md v0.8.0 (REQ-RPC-012, § Amendments).
