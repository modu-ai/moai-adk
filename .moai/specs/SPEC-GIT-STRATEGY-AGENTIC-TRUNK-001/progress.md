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

run 중단·재개 — 2026-10-21 (lane-4, ledger closure per agent-common-protocol § Ledger Closure (a)): the M0-M5 delegation to manager-develop terminated early on API 429 usage limit (5-hour window; reset 2026-10-21 17:59:59 KST; request id 20261021153537976f8d9f38b34c2c, model glm-5.3-flash). It had committed M0 (8b00a111f9, baseline re-check) and M1 (27ebdd9aac, value set/disposition/reader), captured M2 RED verbatim, and its last stated step was the two M2 code changes (guidance fix strings, auto-merge gate/notice). Partial state left UNCOMMITTED in the tree and preserved untouched: M internal/config/agentic_trunk_test.go, M internal/config/integration_target_guidance.go, M internal/config/integration_target_guidance_test.go, ?? internal/cli/agentic_trunk_test.go, ?? internal/escalation/agentic_trunk_test.go, ?? internal/hook/agentic_trunk_test.go. The lane session was itself quota-blocked ~15:35-19:02 KST (queued cron firings drained at recovery). The card lease expired during the block and was restored by the recorded 8859 sequence: factory_stage refusal (expired) reset the lease state → `moai factory next --run tn8rpj --card t1831` from the parent checkout re-leased it with stage=run intact. The delegation was resumed via SendMessage to the same agent after the quota reset (19:05 KST); the delegate continues from its own transcript mid-M2 with the tree exactly as it left it. A second transient request-rate 429 hit mid-M2-GREEN (~19:12 KST, request id 20261021181226b3100bd5773c49e5) and the delegate was resumed again at 19:16 KST; M2 then landed (cba38dd548), followed by M3 (679e474fee) and M4 (9967c92b90).

rescope — leader relay of the operator 10-19 github-flow transition, received 2026-10-21 ~19:4x KST mid-M5 (lane judgment recorded per auto-semantics.md §6; premises VERIFIED before acting per verification-claim-integrity.md §1.1 surface 4): the leader reported (i) the dedicated integration worktree is abolished and the PRIMARY checkout now holds `main` (shared surface, not a clean landing surface), (ii) the leader executed the operator's 10-19 github-flow flip today, and (iii) `factory_card_pr.go:444` requests `gh pr merge --auto`, in conflict with the operator's leader-only-merge principle. Verified in this tree, 2026-10-21: `git worktree list` row 1 reads `/Users/goos/MoAI/moai-adk-go 28b727065b [main]` and no dedicated main-integration worktree exists (premise i CONFIRMED); main's committed `.moai/config/sections/git-strategy.yaml` manual profile still reads `workflow: git-flow` at 28b727065b (the flip is leader-side, not yet in a commit this tree can see); `internal/cli/factory_card_pr.go:444` reads `factoryGHCall(wt, "pr", "merge", strconv.Itoa(pr.Number), "--auto", mergeFlag, "--match-head-commit", checkedTip)` (premise iii CONFIRMED); the operator 10-19 decision is recorded in lane memory (project_flow_transition_per_card_pr_1019: per-card PR flow, leader merges qualified PRs, config flip after sync PR then unblocked same day). Consequence accepted: the SPEC's agentic-trunk delivery model (UsesLocalIntegration → integration-window local merge into `main`, no PR) formalizes a structure this repository no longer has, and adopting it here would block all landings — the same wall the leader names.

LANE DECISION (binding for this card's remaining work, subject to leader/operator override): ship the agentic-trunk VALUE, do not adopt it here. Concretely: (1) M1-M5 output SURVIVES — the value set, classifier, reader/target, local-integration predicate, display, guards, and the M5 rule/delivery sentences describe GENERIC product behavior for user projects that do carry a dedicated landing surface; they stay true and stay. (2) M6 as authored (this repository's live flip to agentic-trunk) is DROPPED — this repository's final live value is `github-flow` per the operator's 10-19 decision already executed by the leader; the M6 replacement is a small alignment commit only (the yaml manual profile records github-flow with an accurate comment; github-flow is a valid value in any running binary, so the installed-binary P1-P5 dependency does not apply). (3) M7 (repo-local guidance) is re-scoped to describe the CURRENT per-card PR flow; the t1830(F) shared-file sentinels are re-judged against the abolished integration worktree (the Integration-Worktree Rule label references a surface that no longer exists — reconcile with the structure, not the name). (4) Leader item ② (the `--auto` merge request vs leader-only merge) is recorded here as a SEPARATE finding, not folded into this card: it is pre-existing github-flow delivery behavior outside the card's value scope; the lane attaches the finding (factory_card_pr.go:444 verbatim above) and the queue's owner decides its routing. (5) The SPEC body rescope (REQ-AGT-017/-018 cutover requirements, AC-016/AC-017, goal wording) is manager-spec work, sequenced AFTER the current delegate returns (one writer per tree); the delegate is left to finish M5 + §E evidence as authored, since that work is delivery-agnostic.

DECISION SUPERSEDED — operator clarification relayed by the leader, 2026-10-21 ~19:5x KST, ~10 minutes after the entry above: the operator confirmed verbatim "지금 사용하는 흐름은 agentic-trunk — 최근 추가된 워크플로우". The leader's earlier github-flow-readjustment framing is WITHDRAWN by its own sender. Corrected directive: (α) the SPEC's final value IS agentic-trunk — this repository ADOPTS it after the card lands; (β) the leader's current github-flow flip is a TEMPORARY BRIDGE only (the installed build 8249bff8c3 lacks the enum this card's M1 is adding — verifiable: the installed binary predates this branch); (γ) the rescope task REMAINS, re-aimed: agentic-trunk's authored delivery ("dedicated integration worktree serial local landing") conflicts with today's structure (no dedicated main tree; the primary checkout holds `main` as a shared, not-clean surface — the lane's verified premises above stand unchanged and are the CONSTRAINTS the redesign must satisfy); the manager-spec pass designs the serial landing surface FOR THE CURRENT STRUCTURE (or a PR-edge hybrid) and reflects the batch PR cycle (lead_push_threshold ≈ 20 landings) in the value definition; (δ) the `factory_card_pr.go:444 --auto` vs leader-only-merge review stays inside this card's rescope review. Operator intent per the leader: agentic-trunk is the main flow; the other values are transition support. The (1) point of the superseded entry still holds — M1-M5 value mechanics survive; the delivery-semantics layer (REQ-AGT-006..008 wiring, M6 cutover, M7 guidance) is what the rescope redesigns. Sequencing unchanged: delegate finishes M5 + §E → manager-spec rescope pass (design choice between serial-surface redesign and PR-edge hybrid belongs to that pass, under the verified structural constraints) → manager-develop applies the deltas → card-review → sync.

run 2차 중단·재개 — 2026-10-22 00:0x KST (lane-4, ledger closure per agent-common-protocol § Ledger Closure (a)): the M5b delegation hit the 5-hour usage limit again (429, request id 202610212007482952882528094f8d, reset 2026-10-21 23:00:25 KST) during its §E verification stretch — its last state: AC-AGT-020 Judges 1-5 PASS, Judge 6 refused by the isolation guard (the awk program carried the `|` form) and it was re-running the equivalent program in the guard-allowed `-v` form. Partial state left UNCOMMITTED in the tree (14 files: M5b items 1-5 edits incl. factory_target_holder.go, the factory_card_pr.go --auto removal, the automerge primary guard, and the item-5 template+deployed pair) and preserved untouched; HEAD still fba207f214. The lane session was quota-blocked ~20:07-24:00 KST (queued cron firings drained at recovery); the lease expired and was restored by the recorded 8859 sequence (stage refusal reset → parent-checkout next, stage=run intact). The delegation was resumed via SendMessage to the same agent at 00:02 KST after the reset; the delegate continues from its own transcript with the tree exactly as it left it. M5b then completed (34a792afe5 + b1806498fe) and was spot-verified by the lane (three flip judges re-run PASS; --auto grep empty on factory_card_pr.go).

M6 round — P4 wait record (2026-10-22 00:3x KST, leader interim report relayed to the lane): P1 PASS (M1-M5b ancestry), P2 PASS (rc build v3.2.0-rc.38 @ b1806498fe installed, byte-verified via scripts/verify-local-install.sh), P3 PASS (fixture doctor reports the agentic-trunk row ok). P4 is BLOCKED: 8 old-build `moai mcp-server` processes are still running (6 lanes' servers + other-project servers), started before the install; flipping the yaml now would make those servers reject the `agentic-trunk` value and break in-flight cards. The leader's respawn experiment (killing its own server) confirmed the host does not reconnect a killed server until that session is relaunched. Completion condition recorded by the leader: as lane sessions cycle to card completion, the server population refreshes; the leader then re-probes P4, makes the one-commit flip, and signals the lane. Until that signal: M7 stays un-started (plan.md M7 ordering — the flip commit must precede the M7 commits on this branch), the lane holds at HEAD b1806498fe.

M7 pre-edit counts — taken early at the leader's request, 2026-10-22 00:3x KST, on tree b1806498fe (`grep -c` verbatim): N1 (`performs landing and resync only`) = 1 in `.claude/rules/local/gitflow-lane-protocol.md` and 1 in `AGENTS.local.md`; N0 (`Integration-Worktree Rule`) = 0 in both, and 0 in `.claude/rules/local/repo-local-pr-policy.md` (not in the counting pair); F2 not applicable (N0 = 0 everywhere — the t1830 label does not exist in the repo-local copies, so there is nothing to reconcile; the template copies keeping the label is expected). N1 + N0 ≥ 1 for both counted files — the plan's zero-against-zero blocker condition is not triggered.

긴급 착지 절차 추가 — leader directive 2026-10-22 ~05:4x KST (auto-merge incident, operator merge hold in effect): EVERY pull request this lane opens is converted to draft immediately after creation (`gh pr ready <n> --undo`), neutralizing any auto-merge request; merges are the leader's alone. This applies to this card's landing (factory complete → PR) because the currently installed binary still carries the pre-M5b `--auto` request at factory_card_pr.go:444 — the M5b removal takes effect only after M6's rc install. The step is consistent with this SPEC's own design (the --auto removal) and stays in force until the leader lifts the hold.

[방침 갱신, 이전 지시 대체] — leader directive 2026-10-22 ~05:5x KST: the draft-undo step cannot guarantee blocking before the auto-merge request is issued, so the COMPLETE/LANDING STEP ITSELF is SUSPENDED — the old-build `complete` hardcodes `--auto` (factory_card_pr.go:444) and cannot be split. Permitted during the hold: repairs, tests, push, and creating pull requests as draft FROM THE START. FORBIDDEN: any call of the old-build `moai factory complete` until the hold lifts. Effect on this card: the lane's landing step is deferred — consistent with the existing M6-signal wait (M7/card-review/sync precede landing anyway); when the hold lifts, the landing runs either under the M6-installed rc.38+ binary (where the M5b code path issues NO merge request) or under an explicit leader authorization.

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

### M1 — Value set, disposition, reader (commit 27ebdd9aac, card t1831)

RED-first. The RED state was captured twice: the acceptance.md ledger E01-E06 cells (`[no tests to
run]`, re-confirmed at M0) and, on this tree, the compile failure of the new test file before the
implementation existed (verbatim, exit reported by go as a build error):

```
internal/config/agentic_trunk_test.go:24:25: undefined: WorkflowAgenticTrunk
internal/config/agentic_trunk_test.go:106:11: got.UsesLocalIntegration undefined (type GitFlowIntegrationConfig has no field or method UsesLocalIntegration)
```

GREEN (final run, HEAD ab639f9898):

```
$ go test ./internal/config -run '^(TestAgenticTrunkValueSet|TestAgenticTrunkTarget|TestAgenticTrunkLocalIntegrationPredicate|TestAgenticTrunkDevelopBranchProjection)$' -count=1 -v
--- PASS: TestAgenticTrunkValueSet (0.00s)
--- PASS: TestAgenticTrunkTarget (0.00s)
--- PASS: TestAgenticTrunkLocalIntegrationPredicate (0.00s)
--- PASS: TestAgenticTrunkDevelopBranchProjection (0.00s)
ok  	github.com/modu-ai/moai-adk/internal/config	0.326s
exit 0
```

Guard updates: `TestWorkflowDisposition` updated 4→5 (subtest now `allowed_set_is_exactly_five_flows`;
`--- PASS: TestWorkflowDisposition (0.01s)`, exit 0 — an update to the new truth, never a skip).
AC-002 compatibility pack stays green (final run: four top-level `--- PASS:` lines, exit 0).
Implementation: `WorkflowAgenticTrunk = "agentic-trunk"` appended as the fifth value,
`agenticTrunkFixedTarget = "main"` beside `githubFlowFixedTarget`, classifier takes the value as
`DispositionValidNonGitFlow` (exact case-sensitive), `WorkflowIntegrationTarget` gains the case,
`AgenticTrunkWorkflow` field + `DevelopBranch` projection to main under manual agentic-trunk
(develop_branch ignored), new `UsesLocalIntegration()`; the second `git-flow` literal replaced by
the `WorkflowGitFlow` constant. Every changed function 100% covered (per-function profile:
`AllowedWorkflows`, `ClassifyWorkflowDisposition`, `WorkflowIntegrationTarget`,
`LoadGitFlowIntegrationConfig`, `UsesLocalIntegration`, `IsGitFlow`, `LoadGitFlowDevelopBranch`
all 100.0%).

### M2 — Consumers (commit cba38dd548)

RED captured verbatim before the two code changes (final RED run of the flip tests, exit 1):

```
$ go test ./internal/cli -run '^TestAgenticTrunkAutoMerge$' -count=1 -v
    agentic_trunk_test.go:294: main HEAD is not a two-parent merge commit: [519d46e...] (notice: "moai: auto-merge: skipped (no integration target): the project is not manual git-flow or git_strategy develop_branch is unset; set git_strategy.<mode>.develop_branch to enable session-exit auto-merge\n")
--- FAIL: TestAgenticTrunkAutoMerge (0.56s)
$ go test ./internal/config -run '^TestAgenticTrunkGuidance$' -count=1 -v
--- FAIL: TestAgenticTrunkGuidance (0.00s)
```

Code changes: the session-exit auto-merge gate `!IsGitFlow()` → `!UsesLocalIntegration()` with the
skip notice naming both workflows (`skipped (no integration target): the project is not manual
git-flow or manual agentic-trunk, or git_strategy develop_branch is unset; set
git_strategy.<mode>.develop_branch (git-flow) or git_strategy.<mode>.workflow: agentic-trunk in
manual mode to enable session-exit auto-merge`); the two git-flow guidance fix strings now end
"`github-flow (target main) or agentic-trunk (target main)`". Everything else is pinned by test —
the factory delivery path, the acquire window-on-main projection, and the three develop-branch seam
consumers (hook `integrationBranchFor`, escalation `cardBase`, contract push check) needed no code
change because M1's projection already carries them. Final GREEN (final run): AC-005 both tests,
`TestAgenticTrunkAutoMerge` (2 subtests), `TestAgenticTrunkIntegrationAcquire` (2 subtests),
`TestAgenticTrunkIgnoredKeys`, `TestAgenticTrunkPushCheckIntegration` (2 subtests),
`TestAgenticTrunkIntegrationBranchFor` (3 subtests), `TestAgenticTrunkCardBase` — all
`--- PASS:`, exit 0; guards (`TestFactoryCompleteWindowGitFlowUnchanged`, auto-merge pair, acquire
pair, doctor) stay green; AC-010 grep `exit 1` (no match) at final HEAD.

### M3 — Display (commit 679e474fee)

RED observed: `vet: internal/statusline/agentic_trunk_test.go:59:21: undefined: resolveWorkflowLabel`.
GREEN (final run):

```
$ go test ./internal/statusline -run '^(TestAgenticTrunkStatuslineLabel|TestAgenticTrunkStatuslineAbsent|TestAgenticTrunkStatuslineNoExtraGit)$' -count=1 -v
--- PASS: TestAgenticTrunkStatuslineLabel (0.00s)      (3 subtests)
--- PASS: TestAgenticTrunkStatuslineAbsent (0.01s)     (6 subtests: github-flow, absent file, unreadable, invalid value, empty value, golden)
--- PASS: TestAgenticTrunkStatuslineNoExtraGit (0.03s) (git shim logs zero invocations)
ok  	github.com/modu-ai/moai-adk/internal/statusline	0.374s    exit 0
```

Implementation: `workflow_provider.go` (`resolveWorkflowLabel` — valid non-github-flow only, one
small yaml read, no git process), `StatusData.WorkflowLabel`, the branch segment appends
` (<workflow>)` after the dirty count, the builder populates the label from the state-anchored root;
doctor `workflowStandingBranches` gains the agentic-trunk case ("single trunk main; cards land
serially in the integration worktree; one batch pull request per lead_push_threshold landings;
tagged release"). `TestAgenticTrunkDoctor` `--- PASS:`, exit 0; guard `TestDoctorGitStrategyWorkflow`
stays green; the full statusline package passes (`ok ... 17.564s`, exit 0) — REQ-PLO-009's
byte-identity cases are pinned inside `TestAgenticTrunkStatuslineAbsent`.

### M4 — Pins and guards (commit 9967c92b90)

```
$ go test ./internal/cli/update -run '^(TestAgenticTrunkValueSurvivesUpdate|TestUpdate_GitStrategyValuesSurvive)$' -count=1 -v
--- PASS: TestAgenticTrunkValueSurvivesUpdate (0.01s)
--- PASS: TestUpdate_GitStrategyValuesSurvive (0.01s)
ok  	github.com/modu-ai/moai-adk/internal/cli/update	0.355s    exit 0
$ go test ./internal/config -run '^TestAgenticTrunkDistributedDefaultsUnchanged$' -count=1 -v   → --- PASS: (3 subtests), exit 0
$ go test ./internal/template -run '^TestAgenticTrunkSweepGuardNotArmed$' -count=1 -v            → --- PASS: (4 subtests), exit 0
$ grep -c 'workflow: agentic-trunk' internal/template/templates/.moai/config/sections/git-strategy.yaml.tmpl
0
exit 1
```

The OQ-2 armed finding count was pre-measured at M0 (266 live develop-base lines) and is recorded
above; it is not zero, so no zero-count follow-up card opens from OQ-2.

### M5 — Template guidance, Template-First (commit ab639f9898)

Template sources edited first (`internal/template/templates/**`), then `make build` (exit 0, catalog
regenerated), then the deployed projections of the eight pairs. Two sentence kinds per the iteration-2
D2 repair: (a) the shared `WT-*` integration-route sentences in `worktree-integration.md`,
`spec-workflow.md`, `delivery-policy.md` and `delivery.md`'s landing-surface paragraph now read
"(the `WT-*` integration route shared by git-flow and agentic-trunk)"; (b) the git-flow
per-version release-branch push sentences (`factory-dispatch.md` lines 219/222,
`factory-dispatch-core.md` line 64, `factory-dispatch-mechanics.md` lines 142-149/243) stay
git-flow ONLY and each file gained one sentence: "Under agentic-trunk the lane takes no
`release/vX.Y.Z` branch and runs no push: `moai factory complete` lands the card through the
integration-window local merge into `main`; the only push is the leader's batch pull request from
`release/main-batch-YYYYMMDD`." — inserted into `factory-dispatch.md` line 222 after "leaves the
batch pull request with the leader." and before "The full window procedure:". `delivery.md` gained
the three-value domain line, the extended stop sentence (`neither `github-flow`, `git-flow` nor
`agentic-trunk`` / `{github-flow, git-flow, agentic-trunk}`), a table row, and the
`##### Strategy: agentic-trunk` block (git-flow route steps 1-5, exit/report, precondition; no
push step; a non-WT branch reports the trunk state; block measured below). SKILL.md gained the
agentic-trunk section; the template `workflow.yaml` comment now names both local-integration
workflows.

AC-020 judges (template delivery.md, final state):

```
Judge 1: 1 (exit 0)   Judge 2: 1 (exit 0)
Judge 3: 239:##### Strategy: github-flow / 269:##### Strategy: git-flow / 279:6. Push the integration branch: ... / 314:##### Strategy: agentic-trunk  (four lines, in order)
Judge 4: 0 and 0 (exit 1 each)   Judge 5: 1 and 1 (exit 0)
Judge 6 (pipe-free re-statement of the acceptance program): block_lines=16 push_lines=0 (exit 0)
  control on the git-flow block: block_lines=34 push_lines=8 (matches ledger E26)
```

AC-021 judges (final state):

```
Judge A: seven `cmp -s` pairs (factory-dispatch-mechanics, factory-dispatch-core,
  worktree-integration, delivery-policy, spec-workflow, delivery.md, SKILL.md) all exit 0.
Judge B(i):  git diff --no-index --numstat → "1	1	{internal/template/templates/.claude => .claude}/rules/moai/workflow/factory-dispatch.md", exit 1
Judge B(ii): 'disposes remote-landing-confirmed trees' → template :0, deployed :1
Judge B(iii): 'Under agentic-trunk the lane takes no' → :1/:1; 'git strategy names git-flow;' → :1/:1;
              'names git-flow or agentic-trunk' → :0/:0 (exit 1)
Judge B(iv): 26217 (template) / 26539 (deployed) — difference 322 bytes (unchanged)
Judge C: coupling pattern over the three release-branch files (template + deployed, six lines) → all :0, exit 1
```

AC-019 judges: the nine template files each carry agentic-trunk (exit 0); github-flow counts
(factory-dispatch.md:1, factory-dispatch-core.md:1, factory-dispatch-mechanics.md:2, delivery.md:12,
others 0-or-more per the judge's own floor). Whole `internal/template` package run after the final
build: only `TestBindingLedgerIntegrity` fails — 2 pre-existing failures (below).

M5 divergences recorded:

- **D-M5-1 (reconciliation, no rule weakened):** `factory-dispatch-core.md` (the always-loaded
  stub) measured 9,999 UTF-16 units at HEAD — one unit under the REQ-RIB-005 10,000 budget — so the
  required agentic-trunk sentence could not fit. The stub's non-[HARD] navigation prose was
  compressed (the detail-companion paragraph's parentheticals and the duplicated `---` separator;
  every [HARD] rule, every companion pointer, and the deployed/template byte-parity kept), landing
  at 9,966 UTF-16 units. `TestRoleCoreStubGate` green both trees (verbatim: "9966 UTF-16 (budget
  10000), role-core markers 0").
- **D-M5-2 (pre-existing, outside this SPEC's scope):** `TestBindingLedgerIntegrity` fails with 2
  rows (ALB-0139, ALB-0148): "after_text not at companion:.claude/rules/moai/core/
  moai-mcp-tools-catalogue.md and before_text no longer in source .claude/rules/moai/core/
  moai-mcp-tools.md (obligation lost)". The moai-mcp-tools files and the binding ledger are not in
  this SPEC's file set and were changed by the absorb merge, not by this card. Same class as the
  known t1858 red; needs its own repair card.
- Line numbers quoted in the plan/acceptance for delivery.md (22/29) and factory-dispatch.md (219/222)
  matched this tree exactly; no AC text was invalidated by the absorb merge.

### M5b — Delivery routing delta, rescope 0.4.0 (commit 34a792afe5)

The target-holder routing (design.md section 1, candidate B): a worktree OTHER than the primary
checkout holding `main` keeps the integration-window local merge (M0-M5 behavior); where ONLY the
primary checkout holds `main`, `factory complete` and `factory merge ready` take the existing
pull-request edge and issue NO merge request — the leader merges, and a later completion run
records `merged-pr` through `factoryObservePRMerge`. New file `internal/cli/factory_target_holder.go`
(`factoryPrimaryHoldsTarget`, `factoryAgenticTrunkPrimaryRoute`); the probe sits beside the
no-worktree-holds-main refusal, and an empty probe (no tree holds the branch) is NOT primary-only —
the existing refusals keep answering that state, which is what keeps the M2 pin judges green.
Session-exit auto-merge gains the second guard clause on the `UsesLocalIntegration` gate, keyed on
`AgenticTrunkWorkflow` (git-flow keeps its present answers — see the D-M5b-2 divergence below); the
skip notice names the missing dedicated landing surface. The doctor's agentic-trunk
standing-branch phrase is reworded holder-neutral with the pinned tokens (`agentic-trunk` via the
message prefix, `main`, `serial`, `batch`) kept. Template-First item 5: the `factory-dispatch.md`
line-222 sentence gained the primary-route clause in BOTH copies; the `delivery.md` agentic-trunk
block gained the same clause as a second `main branch` bullet carrying no `push`/`gh pr` token;
`make build` (exit 0, catalog regenerated), deployed copies brought in line, both pairs committed
together.

RED evidence (TDD, each flip observed red before green). Pre-edit tree fba207f214 — ledger E31
re-observed, both selectors verbatim:

```
go test ./internal/cli -run '^(TestAgenticTrunkCompletePrimaryPrEdge|TestAgenticTrunkCompleteNoAutoMerge)$' -count=1 -v   exit 0
testing: warning: no tests to run
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	0.929s [no tests to run]
go test ./internal/cli -run '^TestAgenticTrunkAutoMergeSkipsPrimaryTarget$' -count=1 -v   exit 0
testing: warning: no tests to run
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	0.757s [no tests to run]
```

Behavioral RED after the tests were written, before the implementation (exit 1, verbatim deciding
lines): `agentic_trunk_test.go:385: complete: factory complete: cannot resolve this session's id;
the integration window needs a holder address` (the completion still took the local-merge path);
`agentic_trunk_test.go:425: merge ready took the window on the primary route`; and the automerge
subtests merged through to the dirty-check instead of skipping (`agentic_trunk_test.go:485/492`,
`...:547: skipped (integration worktree /Users/goos/MoAI/moai-adk-go has uncommitted changes)` —
the pre-guard path hunting the integration worktree against the process cwd, exactly the hazard
the guard exists to close).

Gate matrix at HEAD 34a792afe5 (all commands `go test ./internal/cli -run '<selector>'
-count=1 -v`, exit 0, deciding lines verbatim):

```
AC-AGT-005 flip:  --- PASS: TestAgenticTrunkCompletePrimaryPrEdge (2.98s)
                  --- PASS: TestAgenticTrunkCompleteNoAutoMerge (3.20s)
AC-AGT-006 flip B: --- PASS: TestAgenticTrunkAutoMergeSkipsPrimaryTarget (0.59s)
AC-AGT-005 pin A + guard:
                  --- PASS: TestAgenticTrunkFactoryCompleteTakesWindow (3.30s)
                  --- PASS: TestAgenticTrunkMergeReadyTakesWindow (0.74s)
                  --- PASS: TestFactoryCompleteWindowGitFlowUnchanged (2.95s)
AC-AGT-006 judge A + guards:
                  --- PASS: TestAgenticTrunkAutoMerge (1.32s)
                  --- PASS: TestAutoMergeHappyPath (0.92s)
                  --- PASS: TestAutoMergeUnconfiguredTarget (0.02s)
AC-AGT-007 group: --- PASS: TestAgenticTrunkIntegrationAcquire (0.46s) + both
                  TestIntegrationAcquire_Git*Flow judges PASS
AC-AGT-011:       --- PASS: TestAgenticTrunkDoctor (0.00s)
                  --- PASS: TestDoctorGitStrategyWorkflow (0.01s)
```

Guards (same tree, measured pre-commit on content byte-identical to 34a792afe5): the full
`-run '^TestAutoMerge'` family — 14 tests, all `--- PASS` (OffBaseline through RealImplErrorPaths);
the github-flow delivery suite — 11 tests, all `--- PASS` (TestFactoryCompleteGitHubFlowPR
35.28s through TestFactoryCompleteGitHubFlowClosesTheQueueCard 2.08s, including the updated
MergeMethod matrix and TestMergeReadinessBeforePR); the two specless guards PASS. The github-flow
assertions that pinned the old merge request were UPDATED to the new truth, never skipped: the
delivery subtest now asserts no `gh pr merge` call exists and renames to
`pushes_the_branch_opens_a_pr_against_main_and_requests_no_merge`; the re-run guard asserts
`pr merge` count 0; TestFactoryCompleteGitHubFlowMergeMethod now pins that every merge_method
value (and none) leaves the same wire. TestFactoryCompleteWindowGitFlowUnchanged and
TestAgenticTrunkFactoryCompleteTakesWindow stay green — the routing keeps dedicated-worktree
fixtures on the local-merge path.

Item-5 judges (AC-019/020/021, post-edit):

```
AC-019: nine template files each carry agentic-trunk (exit 0), counts 1..6.
AC-020 J1-J5: 1 / 1 / four heading lines in order (239/269/279/314) / 0,0 / 1,1 — all as pinned.
AC-020 J6: block_lines=17 push_lines=0 (template AND deployed, exit 0);
  control on the git-flow block: block_lines=34 push_lines=8 — exactly ledger E26's value.
AC-021 Judge A: seven cmp -s pairs exit 0.
Judge B(i): numstat "1	1	{...=>...}/factory-dispatch.md", exit 1 (exactly one differing line).
Judge B(ii): 'disposes remote-landing-confirmed trees' → template :0, deployed :1 (preserved).
Judge B(iii): 'Under agentic-trunk the lane takes no' :1/:1; 'git strategy names git-flow;' :1/:1;
  'names git-flow or agentic-trunk' :0/:0; the M5b clause '; where only the primary checkout
  holds `main`, `moai factory complete` opens the card' :1/:1.
Judge B(iv): 26343 (template) / 26665 (deployed) — difference 322 bytes, unchanged (both copies
  grew by the same 126-byte clause from 26217/26539).
Judge C: coupling pattern over the six labeled lines → all :0, exit 1.
```

E-batch results at HEAD 34a792afe5: `go build ./...` exit 0; `GOOS=windows GOARCH=amd64 go build
./...` exit 0; `golangci-lint run --timeout=2m ./internal/cli/...` → `0 issues.` (exit 0);
boundary grep `AskUserQuestion|mcp__askuser` over the nine touched files → no matches (exit 1);
anchored coverage (`-coverprofile` over the ten-test selector): `factoryPrimaryHoldsTarget
85.7%`, `factoryAgenticTrunkPrimaryRoute 75.0%`, the uncovered arms being the defensive
`identifyPrimaryCheckout`-error and empty-target paths, `sessionExitAutoMerge 47.2%`,
`factoryDeliverByPR 58.2%`, `factoryCompleteCard 54.7%` (the deep pre-existing branches outside
the anchored selector; every NEW branch is exercised by the flip judges' assertions).

M5b divergences recorded:

- **D-M5b-1 (guard refusal, judge preserved):** AC-AGT-020 Judge 6's verbatim awk program carries
  a `|` form the worktree isolation guard refuses (refused twice: the plain program and the
  guard's own `-v re='a|b'` suggestion — Gaps entry per verification-claim-integrity §3.1). A
  pipe-free equivalent was used instead: same block-boundary rules (`/^#+ /` with
  `length($1)<=4` replacing `^(#|##|###|####) `, index-based once-per-line push counting with
  arithmetic OR), validated by the git-flow control reproducing ledger E26's exact 34/8 — the
  counter is a measurement, not an absence. On the agentic block it reads block_lines=17
  push_lines=0 on both copies.
- **D-M5b-2 (guard scoping, judge-mandated):** the session-exit skip clause is keyed on
  `AgenticTrunkWorkflow`. A first unconditional version broke AC-AGT-006's own guard judge
  `TestAutoMergeHappyPath` (manual git-flow, primary-only develop — the flow's fixture merges into
  the develop holder as today, with the REQ-WKW-008 dirty-target guard protecting the primary).
  AC-AGT-006 scopes the skip to the agentic-trunk project ("where only the primary checkout holds
  `main`" under `Given ... manual agentic-trunk`), so the scoped clause is the contract-correct
  reading, and the AC-006 judge group is green with it.
- The M2 fixture `autoMergeAgenticTrunkFixture` was re-aimed to the dedicated-worktree shape (the
  primary vacates main, a `.claude/worktrees/main` integration worktree takes it) — the update to
  the new truth REQ-AGT-024 admits, keeping AC-AGT-006 judge A's merge assertions meaningful under
  the routing; its merge assertions now read the integration worktree and assert the primary's own
  branch did not move.
- **D-M5b-3 (fixup during implementation):** an Edit intended for the merge-request block in
  `factoryDeliverByPR` initially landed on the readiness prologue and corrupted the function; the
  region was immediately rewritten whole and verified (`go vet` clean, no `gh pr merge`/
  `mergeFlag`/`factoryPRMergeFlag` references remain, the function flow re-read end to end). No
  corrupted state was committed.

## §E.3 Run-phase Audit-Ready Signal

run_complete_at: 2026-10-22 (M0-M5 + M5b delegation complete; the run PHASE is not complete — see run_status)
run_commit_sha: 34a792afe5 (the M5b code commit; the progress-record commit carrying this signal follows it, so its own SHA is unknowable until it lands — backfilled by the next progress-record commit if needed)
run_status: partial — M0 (baseline re-check), M1 (value set/disposition/reader), M2 (consumers), M3 (display), M4 (pins/guards), M5 (template guidance), M5b (delivery routing delta, rescope 0.4.0) complete and green; M7 (repo-local guidance) is a later run-phase delegation per the §E.1 scope boundary; M6 (cutover, P1-P5) is the leader's step; M8 (docs-site) is sync-phase; M9 (close, AC-024) pending.
ac_pass_count: 19 (AC-001..015, AC-019, AC-020, AC-021, AC-025 — every gate in this delegation's scope; the AC-005/006 flip judges flipped green in M5b)
ac_fail_count: 0 (two pre-existing template-package failures are recorded as divergences, not this SPEC's ACs: the armed-mode ratchet of the diagnostic sweep branch, and TestBindingLedgerIntegrity rows ALB-0139/0148 broken by the absorb merge — see D-M5-2)
preserve_list_post_run_count: 0 files outside the delegation's scope touched; `.moai/config/sections/git-strategy.yaml`, `AGENTS.local.md`, `.claude/rules/local/gitflow-lane-protocol.md`, `git-strategy.yaml.tmpl`, `manager-git.md/.toml`, `auto-semantics.md`, `spec-workflow-detail.md`, `tab_schema.json`, docs-site, `.moai/harness/*`, `.moai/state/*` all untouched (verified by git status at each commit and by `git status --short .moai/config/sections/git-strategy.yaml` printing nothing after M5 and M5b)
l44_pre_commit_fetch: n/a (no push in this delegation; the lane handles landing — no fetch performed, none owed)
l44_post_push_fetch: n/a (no push performed)
new_warnings_or_lints_introduced: 0 (lint over all touched packages: "0 issues"; baseline was 1 pre-existing staticcheck finding in internal/discovery — unchanged, package untouched)
cross_platform_build.darwin: go build ./... exit 0 at HEAD 34a792afe5
cross_platform_build.windows: GOOS=windows GOARCH=amd64 go build ./... exit 0 at HEAD 34a792afe5
total_run_phase_files: 39 (10 Go source incl. the new factory_target_holder.go, 14 Go test files incl. 6 new agentic_trunk test files, spec.md frontmatter, catalog.yaml, 9 template sources, 8 deployed projections — some files counted once across categories)
m1_to_mN_commit_strategy: one commit per milestone (M0 8b00a111f9 docs, M1 27ebdd9aac feat, M2 cba38dd548 feat, M3 679e474fee feat, M4 9967c92b90 test, M5 ab639f9898 feat, M5b 34a792afe5 feat); the M0 baseline record precedes M1 per verification-claim-integrity §2.3; the M1-M5b commits precede the M6 flip per P1 lineage; no push, no PR, no merge — the lane handles landing.

Residual risks for the sync audit:

- The red `internal/config` test `TestNoHardcodedContexts` (card t1858) and the red
  `TestBindingLedgerIntegrity` rows (ALB-0139/0148, absorb-merge damage) remain red on this branch;
  neither is caused by this SPEC's diff (both subject files untouched by it) and both have named
  repair owners outside this card.
- The factory-dispatch-core stub reconciliation (D-M5-1) compressed non-[HARD] navigation prose to
  fit REQ-RIB-005; an audit wanting the fuller companion descriptions finds them on the companion
  files themselves (the pointers were kept).
- The OQ-2 armed finding count (266) is a snapshot of a moving repository; the residue-sweep card
  should re-measure at its own start.
- M5b's AC-020 Judge 6 ran through a pipe-free equivalent of the pinned awk program (the verbatim
  program is refused by the worktree isolation guard, D-M5b-1); the equivalence rests on the
  git-flow control reproducing ledger E26's 34/8 exactly, and the M5 diff review of the block
  remains the second layer over reworded-push phrasings the counter cannot see.

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
