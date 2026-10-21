# Progress — SPEC-GIT-STRATEGY-AGENTIC-TRUNK-001

## §E.1 Plan-phase Audit-Ready Signal

plan_status: audit-ready
plan_complete_at: 2026-10-18
plan_tree: c6e590507b39e27acc3305a3a1aa3dc488aef0c0 (branch WT-agentic-trunk; the tree the iteration 2 audit judged and the iteration 2 repair observed); tier L; 25 requirements, 25 acceptance criteria (version 0.3.1); open questions OQ-1 to OQ-7 carry recorded defaults (spec.md section 5).
plan_audit: iteration 1 returned FAIL 0.75 on tree f4175587119b5cfeb8d8348fc74cada8764e89ed (report .moai/reports/t1831/plan-audit.md, defects D1-D11); version 0.2.0 repaired D1-D7 and folded the optional D8-D11. Iteration 2 returned FAIL 0.875 on tree c6e590507b39e27acc3305a3a1aa3dc488aef0c0 (report .moai/reports/t1831/plan-audit-iter2.md, blocking D1 and D2, optional D3-D6, codex receipt rcpt-d75d4a819c98c8c571ab00c4); version 0.3.0 repairs D1 and D2 and the optional D3-D6. Iteration 3 returned FAIL on the required codex gate (report .moai/reports/t1831/plan-audit-iter3.md, aggregate 0.875, blocking D1, optional D2-D6); version 0.3.1 repairs D1 only (Judge 6 block end and two failing controls, ledger E26). The iteration 4 re-audit has not run; this record does not claim it passes.

Known debt (plan-audit iteration 3, not repaired):
- D2: AC-AGT-021 Judge B (iv) compares the 322-byte size difference only; a same-length edit to one copy of the factory-dispatch line 222 passes it, and Judge B (iii) greps the clause prefix only.
- D3: AC-AGT-023 Judge 3 is vacuous at N0 = 0 for the alternation blind spot; Judges 4 and 5 carry the non-vacuity, Judge 3 alone does not.
- D4: REQ-AGT-017 P4 wording uses `pgrep -o` (oldest process only) while the P4b exception reads over all running servers; wording mismatch, resolved the same way by a reasonable engineer.
- D5: the `ps eww` foreign-server judge (servers whose CLAUDE_PROJECT_DIR is another project) is not recorded as its own ledger entry.
- D6: REQ-AGT-021 literal coverage of the `workflow.yaml` template comment is stated by the requirement but not judged line by line.
- N1 (plan-audit iteration 4): a 4-backtick nested fence and a `~~~` fence push are not caught by AC-AGT-020 Judge 6.
- N2 (plan-audit iteration 4): the AC-AGT-020 pattern is not case-insensitive although the AC text says it is.

hold 확정 — 2986: leader ruling on board d-20261018T024858Z-2986, 2026-10-18: card t1831 is held at the plan stage after the iter4 delta audit (verdict FAIL value only via the required codex gate, own score 0.90, blocking_count 0, receipt rcpt-28a602dcf1e13dfd35d106ff, report .moai/reports/t1831/plan-audit-iter4.md at audited_sha 0dd8dcfbfc60e2438106dae71691a48979d9bf2f); the required codex gate is not lowered; whether to accept N1 as debt and go to kickoff or to narrow the SPEC is the operator's call on return; no further audit rounds, no SPEC edits, no run entry; the tree is preserved and the lease is no longer renewed.

kickoff — operator form, hold resolved 2026-10-21: the hold's open question is resolved. The factory leader re-dispatched card t1831 to lane-4 (run tn8rpj) with the card body unchanged ("카드 본문 그대로") — the SPEC is NOT narrowed, the recorded debt is accepted, and the card proceeds to run. Mechanical evidence the hold was lifted before the lease: the §9.3 keep-set refuses a card in the hold state, and this lease succeeded. The iter4 verdict label stays FAIL on the label axis (via the required codex gate only), so the §9.1 default-autonomous admission predicate does not apply; this transition is the operator form the hold reserved.

decision record: decided_by=operator (via factory leader re-dispatch, run tn8rpj, lane-4, 2026-10-21) evidence_refs=board d-20261018T024858Z-2986; .moai/reports/t1831/plan-audit-iter4.md (verdict FAIL via required codex gate, own score 0.90, blocking_count 0, rcpt-28a602dcf1e13dfd35d106ff, audited_sha 0dd8dcfbfc60e2438106dae71691a48979d9bf2f); card t1831 body unchanged; lease success (hold-state refusal not fired) ladder_path=plan->run Kickoff (operator form; auto-semantics.md §9.1)

Binding run conditions (copied from the plan-audit debt, accepted at this kickoff; both sync verdict owners re-read these; an undisposed condition is a failed must-pass criterion that caps the sync verdict at FAIL — auto-semantics.md §9.1):
- D2 (iter3): AC-AGT-021 Judge B (iv) compares the 322-byte size difference only; a same-length edit to one copy of the factory-dispatch line 222 passes it, and Judge B (iii) greps the clause prefix only.
- D3 (iter3): AC-AGT-023 Judge 3 is vacuous at N0 = 0 for the alternation blind spot; Judges 4 and 5 carry the non-vacuity, Judge 3 alone does not.
- D4 (iter3): REQ-AGT-017 P4 wording uses `pgrep -o` (oldest process only) while the P4b exception reads over all running servers; wording mismatch, resolved the same way by a reasonable engineer.
- D5 (iter3): the `ps eww` foreign-server judge (servers whose CLAUDE_PROJECT_DIR is another project) is not recorded as its own ledger entry.
- D6 (iter3): REQ-AGT-021 literal coverage of the `workflow.yaml` template comment is stated by the requirement but not judged line by line.
- N1 (iter4): a 4-backtick nested fence and a `~~~` fence push are not caught by AC-AGT-020 Judge 6.
- N2 (iter4): the AC-AGT-020 pattern is not case-insensitive although the AC text says it is.

Scope boundary recorded at kickoff: this card's run covers M0-M5 and M7 (repo-local guidance, authored AFTER the leader's M6 so the ordering rule of plan.md M7 holds); M6 (cutover: installed-binary replacement with operator confirmation, P1-P5) is leader-executed per plan.md M6 and design.md section 4 — the lane does NOT flip `.moai/config/sections/git-strategy.yaml`; the live-value flip commit and the M7 commits follow it on this branch; M8 (docs-site) is sync-phase (manager-docs). The card's completion report names M6 as the remaining leader step with its P1-P5 list.

## §E.2 Run-phase Evidence

### M0 — Baseline re-check (2026-10-21, run tree 95e777a36e, branch WT-agentic-trunk)

M0 is a RE-CHECK, not a first measure: the RED-now cells in acceptance.md were authored on tree
`f4175587119b5cfeb8d8348fc74cada8764e89ed` before the absorb merge `89996a6de5` (the branch was 341
commits behind main and is now current with it). Commands, verbatim outputs and exit codes below are
this run, this tree, HEAD `95e777a36e` (VCI §2 attribution triple).

Pin-path diff (the acceptance.md §Conventions re-check command):

```
$ git diff --stat f4175587119b5cfeb8d8348fc74cada8764e89ed..HEAD -- internal .claude AGENTS.local.md docs-site .moai/config .moai/docs | tail -1
 267 files changed, 19076 insertions(+), 2904 deletions(-)
```

The absorb merge moved 155 files across internal/config, internal/cli, internal/statusline,
internal/hook, internal/escalation and internal/cli/update, so every guard selector whose package
changed was re-observed; the flip selectors (all naming only new tests) were re-observed as a
population. Divergences found during M0 against the pinned observations are listed at the end.

Pre-flight (Section C of the delegation):

```
$ git rev-parse --short HEAD && git branch --show-current      → 95e777a36e / WT-agentic-trunk
$ go build ./...                                                → exit 0 (no output)
$ GOOS=windows GOARCH=amd64 go build ./...                      → exit 0 (no output)
$ golangci-lint run --timeout=2m 2>&1 | tail -5                 → exit 0
  internal/discovery/leader_readers_darwin.go:93:31: SA1019: unix.SYS_PROC_INFO is deprecated:
  Use libSystem wrappers instead of direct syscalls. (staticcheck)
  1 issues:
  * staticcheck: 1
```

Lint baseline: exactly 1 pre-existing finding (staticcheck SA1019 in internal/discovery, a package
this SPEC does not touch). NEW findings during M1-M5 are reported against this baseline (E5).

Cross-SPEC policy conflict grep (delegation Section C.4): all matches in
`internal/config internal/cli internal/statusline internal/hook internal/escalation` are unrelated
retired-CG-launcher comments and other SPECs' superseded markers; none touches the git-strategy
workflow domain. No conflict.

Guard judges re-observed (all previously green on the pin tree; must stay green):

| Guard selector (command form `go test <pkg> -run '<sel>' -count=1 -v`) | exit | deciding lines (verbatim) |
|---|---|---|
| `./internal/config` `^TestWorkflowDisposition$` | 0 | `--- PASS: TestWorkflowDisposition (0.01s)` / `ok  github.com/modu-ai/moai-adk/internal/config	0.304s` |
| `./internal/config` `^(TestWorkflowTargetResolution\|TestEmptyTargetGuidance\|TestEmptyTargetGuidanceResolvedTargetIsSilent\|TestTargetProvenance)$` (AC-002) | 0 | `--- PASS: TestEmptyTargetGuidance (0.01s)`, `--- PASS: TestEmptyTargetGuidanceResolvedTargetIsSilent (0.00s)`, `--- PASS: TestTargetProvenance (0.00s)`, `--- PASS: TestWorkflowTargetResolution (0.00s)` / `ok ... 0.137s` |
| `./internal/cli` `^TestFactoryCompleteWindowGitFlowUnchanged$` (AC-005 guard) | 0 | `--- PASS: TestFactoryCompleteWindowGitFlowUnchanged (1.69s)` / `ok ... 2.561s` |
| `./internal/cli` `^(TestAutoMergeHappyPath\|TestAutoMergeUnconfiguredTarget)$` (AC-006 guard) | 0 | `--- PASS: TestAutoMergeHappyPath (0.43s)`, `--- PASS: TestAutoMergeUnconfiguredTarget (0.00s)` / `ok ... 1.094s` |
| `./internal/cli` `^(TestIntegrationAcquire_GitFlowEmptyDevelopWarnsOnCallerFallback\|TestIntegrationAcquire_GitHubFlowEmptyDevelopDoesNotWarn)$` (AC-007 guard) | 0 | `--- PASS: TestIntegrationAcquire_GitFlowEmptyDevelopWarnsOnCallerFallback (0.13s)`, `--- PASS: TestIntegrationAcquire_GitHubFlowEmptyDevelopDoesNotWarn (0.14s)` / `ok ... 0.931s` |
| `./internal/cli` `^TestDoctorGitStrategyWorkflow$` (AC-011 guard) | 0 | `--- PASS: TestDoctorGitStrategyWorkflow (0.01s)` / `ok ... 0.688s` |
| `./internal/cli/update` `^TestUpdate_GitStrategyValuesSurvive$` (AC-013 guard) | 0 | `--- PASS: TestUpdate_GitStrategyValuesSurvive (0.00s)` / `ok ... 0.357s` |
| `./internal/template` `^TestGitHubFlowSweepGuard$` (AC-015 guard) | 0 | `--- PASS: TestGitHubFlowSweepGuard (2.53s)` with subtests `--- PASS: TestGitHubFlowSweepGuard/arming-switch (0.00s)`, `--- PASS: TestGitHubFlowSweepGuard/F-empty-walk (0.00s)` / `ok ... 2.874s` |

AC-010 grep guard re-observed:

```
$ git grep -c -E 'LoadGitFlowIntegrationConfig|WorkflowGitFlow|WorkflowGitHubFlow|WorkflowAgenticTrunk|"agentic-trunk"' -- internal/cli/worktree
(no output)
$ echo $?
1
```

Green by design (no workflow read in the landing checks); unchanged by the absorb merge.

Flip-selector RED re-confirmation (every selector below printed `[no tests to run]` with exit 0 —
the RED-now state of the acceptance.md ledger E01-E17 holds on this tree; the tail line is quoted
verbatim, selector prefix elided):

- `./internal/config` `^TestAgenticTrunkValueSet$` → `ok  github.com/modu-ai/moai-adk/internal/config	0.127s [no tests to run]`
- `./internal/config` `^TestAgenticTrunkTarget$` → `ok ... 0.130s [no tests to run]`
- `./internal/config` `^TestAgenticTrunkLocalIntegrationPredicate$` → `ok ... 0.126s [no tests to run]`
- `./internal/config` `^TestAgenticTrunkGuidance$` → `ok ... 0.130s [no tests to run]`
- `./internal/config` `^TestAgenticTrunkDistributedDefaultsUnchanged$` → `ok ... 0.130s [no tests to run]`
- `./internal/config` `^TestAgenticTrunkDevelopBranchProjection$` → `ok ... 0.139s [no tests to run]`
- `./internal/cli` `^(TestAgenticTrunkFactoryCompleteTakesWindow|TestAgenticTrunkMergeReadyTakesWindow)$` → `ok ... 0.692s [no tests to run]`
- `./internal/cli` `^TestAgenticTrunkAutoMerge$` → `ok ... 0.683s [no tests to run]`
- `./internal/cli` `^TestAgenticTrunkIntegrationAcquire$` → `ok ... 0.681s [no tests to run]`
- `./internal/cli` `^TestAgenticTrunkIgnoredKeys$` → `ok ... 0.670s [no tests to run]`
- `./internal/cli` `^TestAgenticTrunkDoctor$` → `ok ... 0.664s [no tests to run]`
- `./internal/cli` `^TestAgenticTrunkPushCheckIntegration$` → `ok ... 0.803s [no tests to run]`
- `./internal/statusline` `^(TestAgenticTrunkStatuslineLabel|TestAgenticTrunkStatuslineAbsent|TestAgenticTrunkStatuslineNoExtraGit)$` → `ok ... 0.322s [no tests to run]`
- `./internal/cli/update` `^TestAgenticTrunkValueSurvivesUpdate$` → `ok ... 0.165s [no tests to run]`
- `./internal/template` `^TestAgenticTrunkSweepGuardNotArmed$` → `ok ... 0.196s [no tests to run]`
- `./internal/hook` `^TestAgenticTrunkIntegrationBranchFor$` → `ok ... 0.567s [no tests to run]`
- `./internal/escalation` `^TestAgenticTrunkCardBase$` → `ok ... 0.451s [no tests to run]`

M0 conclusion: every guard judge the acceptance.md pins as green-now is green on the run tree; every
flip selector is still RED (`no tests to run`); the builds pass on darwin and windows; the lint
baseline carries exactly one pre-existing finding. The absorb merge invalidated no RED-now cell —
the RED cells name tests that still do not exist — and moved no guard to red. Line numbers quoted in
acceptance.md/plan.md (e.g. delivery.md line 22/29, factory-dispatch.md line 222) are re-verified at
the milestone that edits them; a moved line number is a recorded divergence, not a blocked AC.

OQ-2 pre-measurement (plan M4 obligation, recorded here at M0 for the follow-up card):

```
$ MOAI_SWEEP_GUARD_ARMED=1 go test ./internal/template -run '^TestGitHubFlowSweepGuard$' -count=1 -v
exit 1; deciding lines (verbatim):
    github_flow_sweep_guard_test.go:225: sweep guard ARMED: forced by MOAI_SWEEP_GUARD_ARMED=1
    github_flow_sweep_guard_test.go:233: weak hits: allowed=0 remaining=25; strong remaining=241; allow entries used=0/0 (cap 40)
    github_flow_sweep_guard_test.go:337: 266 live develop-base line(s) remain (241 strong, 25 weak); first 40: ...
    github_flow_sweep_guard_test.go:341: ratchet: marker-exempted lines = 18, ceiling 11 — a marker silenced more live text than the recorded baseline
--- FAIL: TestGitHubFlowSweepGuard (2.59s)
    --- PASS: TestGitHubFlowSweepGuard/arming-switch (0.00s)
    --- PASS: TestGitHubFlowSweepGuard/F-empty-walk (0.00s)
```

Armed finding count for OQ-2: **266 live develop-base line(s) remain (241 strong, 25 weak)** —
NOT zero, so per OQ-2's condition no follow-up card opens on a zero count; the count is recorded for
the card that owns the residue sweep. Two observations about the armed mode itself: (1) the armed
run fails on this tree through the marker-exempted ratchet (18 > ceiling 11) — a pre-existing state
of the absorbed tree, independent of this SPEC's files (the un-armed run passes; the armed branch is
diagnostic-only, `github_flow_sweep_guard_test.go:323-342`); (2) the count is a repository residue
measurement, not a judgment of this SPEC's changes — none of the 266 lines is in a file this SPEC
edits except `.claude/rules/local/gitflow-lane-protocol.md:196` (weak, already `[SUPERSEDED]`-marked)
and template rule files whose develop-era text this SPEC does not touch (M5 adds sentences only).

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

## §F Phase 4 Mode Selection

Input parameters: tier=L; scope ~25-40 files (Go config/cli/statusline + template rules/skill + repo-local rules); domains = Go source, templates, repo-local guidance, config (4); file language mix = Go + Markdown; concurrency benefit = LOW (coding-heavy, Anthropic coding-task parallelism caveat); Agent Teams prereqs = not requested.

| Mode | Selected | Rationale |
|------|----------|-----------|
| direct | no | multi-file, multi-domain change; not a typo/single-line fix |
| serial | **yes** | coding-heavy implementation; one manager-develop per milestone sequence is the safe default |
| fanout | no | write-capable fan-out on one card is prohibited (one writer per tree); research fan-out is Class-C plan-phase only and plan is done |
| sweep | no | not a single uniform mechanical transform; semantic multi-rule work |

Decision: serial (orchestration-mode-selection.md §B: coding-heavy → serial; §A default fallback).

Justification: the milestones are sequential (M1 data model → M2 consumers → M3 display → M4 guards → M5 template guidance, each gated on the prior), the change is coding-heavy across Go and Markdown, and a single manager-develop spawn per stage keeps one writer on the lane tree. The §C.3 sweep preconditions fail (no uniform transform, not ≥30 mechanical files).

Boundary case: none — the decision tree resolved on the first coding-heavy branch.
