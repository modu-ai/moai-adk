# progress.md — SPEC-FACTORY-LANE-AUTONOMY-001 (card t1338)

## §E.1 Plan-phase Audit-Ready Signal

```yaml
plan_status: audit-ready
plan_complete_at: 2026-09-29
tier: L
artifacts: 5 (spec.md, plan.md, acceptance.md, design.md, research.md) + progress.md
req_count: 16
ac_count: 17
needs_clarification_markers: 0
baseline: worktree .moai/worktrees/t1338, branch WT-lane-autonomy-umbrella, base develop 145c3d98c
run_entry_gate: M0 — t1240 (SPEC-FACTORY-SELF-DISPATCH-001) develop merge confirmed mechanically before M1
```

**Plan-audit trajectory** (reports under `.moai/reports/t1338/`, first lines name the serving
auditor model `glm-5.3-flash` per the GLM-lane attribution rule): iter1 `plan-audit-iter1.md`
**PASS 0.97** (Tier L threshold 0.85; 0 blocking / 0 major / 3 minor / 2 advisory; iteration
1/3 — loop closed at first PASS). The 3 minors were polished post-audit (plan.md H1 id typo
AUTOMY→AUTONOMY; research.md autoEvidencePath :36→:32, autoLiveness :47→:51 — each re-verified
by the author against `internal/cli/todo_auto.go`). Carried advisories: D3's "landed-enough"
predicate remains documented discipline until t1241's SPEC text is pinnable at M3 entry (F4);
plan commit `dade0e534` carries no `Authored-By-Agent:` trailer (session attribution reminder
takes precedence; lint INFO OwnershipTransitionUnmeasured, non-strict).

**Phase 2/6 research skip rationale (FO-PLAN-1 note)**: the plan-research fan-out script
(`plan-research-fanout`) exists, but a SINGLE-Explorer pass was chosen for this card. Recorded
decision of the dispatching orchestrator, not a silent deviation: the domain is one coherent
factory-operations subsystem, the 4 boundary SPECs/documents already document the layered surfaces
(t1240/t1241 branch-resident, F1 + t1306 landed in develop), and the carried reconnaissance
(research.md R1-R5) covers every surface the 4 fragments touch. A multi-lens fan-out would have
re-derived the same boundary evidence at ~4x the read cost with no new decision input.

## §E.2 Run-phase Evidence

### M1 — Messaging-fallback detection + self-service switch (2026-09-29, owner: manager-develop)

Surfaces: new package `internal/factorylane` (probe + directed-request observations +
fallback-transition log), CLI verbs `moai factory messaging probe|request|ack`,
`moai factory fallback [declare|restore]`, `moai factory handoff adopt`; config constants
`DefaultFactoryNoResponseMinutes` / `DefaultFactoryFallbackBoundMinutes`
(`internal/config/defaults.go`; heartbeat-age bound reuses `DefaultSessionMsgAgentOfflineMinutes`
— single source). Milestones M2 done (below); M3-M5 pending.

| AC | Status | Verification command | Actual output |
|----|--------|---------------------|---------------|
| AC-FLA-001 | PASS (M1 Go surface) | `go test ./internal/factorylane/ -run TestProbe -count=1` + `go test ./internal/cli/ -run TestFactoryMessagingProbe -count=1` | `ok github.com/modu-ai/moai-adk/internal/factorylane` + `ok ... internal/cli` (probe verbs report `channel: available` / `channel: channel-unavailable` + `--json` verdict; lane switch act = `/moai:todo --auto` doctrine wiring) |
| AC-FLA-002 | PASS (M1 Go surface) | `go test ./internal/factorylane/ -run 'TestSweep|TestAck|TestProbe' -count=1` | `ok ... internal/factorylane` (timer expiry records `no_response_at`; channel unavailable until `requested_at + bound`; ack before expiry stays available) |
| AC-FLA-003 | PASS (M1 Go surface) | `go test ./internal/factorylane/ -run TestDeclareFallback -count=1` + `go test ./internal/cli/ -run TestFactoryFallback -count=1` | `ok` both (exactly one event per switch — second declare refused, count stays 1; restore enables a counted second switch; count-by-lane query prints events) |
| AC-FLA-004 | PASS (M1 Go surface) | `go test ./internal/cli/ -run TestAdopt -count=1` | `ok ... internal/cli` (adopt prints progress.md content + evidence BEFORE work, derives `recorded phase: run`/`sync` from markers, refuses non-picked and nothing-recorded) |
| AC-FLA-005 | PASS (M1 Go surface) | `go test ./internal/cli/ -run TestAdoptAppendsResumption -count=1` | `ok ... internal/cli` (previous owner's progress.md + evidence SHA-256 unchanged; `resumption.jsonl` appended alongside) |

M1 E2-E6 outputs (builds, coverage, boundary grep, lint, commit list) are carried verbatim in the
run-phase completion report returned to the orchestrator and pinned to the M1 commit SHA in git
history; M0 gate evidence: `d43e50bb3` ancestor + SPEC-FACTORY-SELF-DISPATCH-001 present (§F
record). RED-before-GREEN evidence per TDD cycle (5 cycles) is likewise in the completion report
(verbatim failing-test output captured before each GREEN).

### M2 — Classified pickup consumption (2026-09-30, owner: manager-develop)

Surfaces: consumer decision core `internal/factorylane/pickup.go` — the DECLARED MINIMAL
CONSUMPTION INTERFACE per design.md D2 (`Classifier` seam, `NormalizeClassification` tolerance
point, `PlanPickup` rules, `Hold` read from the F1 lease model) — and the CLI verb
`moai factory pickup plan` (`internal/cli/factory_pickup.go`, registered beside
messaging/fallback). t1332 producer GATED: absent as a SPEC at M2 entry (plan.md §F M2 gate), so
every card evaluates through the REQ-FLA-007 fallback default; `StaticClassifier` and the CLI's
flag adapter are the documented wiring points t1332's future reader replaces — no producer schema
defined here (AC-FLA-008 grep 0, below). Exclusivity rides the existing lease records
(`homestate.IsLeaseHoldingState` holders) — no new lock, integration window untouched.

| AC | Status | Verification command | Actual output |
|----|--------|---------------------|---------------|
| AC-FLA-006 | PASS (M2 Go surface) | `go test ./internal/factorylane/ -run TestPickup -count=1` + same with `-race` | `ok github.com/modu-ai/moai-adk/internal/factorylane` (sequential group held by another lane → denied, `WaitOn` names the holder; free group → allowed exclusive `MultiPick=false`; parallel → allowed `MultiPick=true`; two-lane concurrent contention over one sequential group ends with exactly 1 winner, parallel pair → 2 winners; `-race` clean) + CLI edge `ok ... internal/cli` (`--axis sequential --group alpha` prints the classified decision) |
| AC-FLA-007 | PASS (M2 Go surface) | `go test ./internal/factorylane/ -run TestPickupUnclassified -count=1` + `go test ./internal/cli/ -run TestFactoryPickupPlanFallsBackWithoutMetadata -count=1` | `ok` both (unclassified card → `Allowed=true` with `Fallback=true` behavior marker and `MultiPick=false` — the operator-picked single-dispatch behavior, no autonomous multi-pick; reason names REQ-FLA-007) |
| AC-FLA-008 | PASS (M2 Go surface) | `go test ./internal/factorylane/ -run 'TestNormalize|TestPickupUnknownAxis|TestStaticClassifier' -count=1` + producer-schema symbol grep | `ok` (unknown axis token → `Known=false`, decision exits nil-error with `ToleratedUnknown=true` + tolerated log; absent metadata → fallback classification) + `grep -rnE 'MetadataSchema\|SchemaDefinition\|type [A-Za-z]*Metadata struct\|ClassificationSchema\|ProducerSchema\|BacklogItemMetadata' internal/factorylane/ internal/cli/factory_pickup.go` → **0 matches** (no producer-schema code in the consumer) |

M2 E2-E8 outputs (both builds, coverage, boundary grep, lint 0 issues, commit list, verbatim RED
per cycle, the AC-FLA-008 grep, and the `TestSD_AC021_LegacySpellingsRefused` guard result) are
carried in the run-phase completion report returned to the orchestrator; suite evidence persisted
at `.moai/state/verify/t1338-m2/` under the M2 commit SHA in git history. E3 GAP (recorded like
M1's): the `internal/cli` FULL-suite aggregate coverage line was NOT obtained — the full-suite run
under the slot lease hit go test's default 10m timeout and was killed mid-package (`panic: test
timed out after 10m0s`, 601.8s, **0** `--- FAIL` lines — no test failure, an elapsed-bound kill;
the mid-dump `coverage: 24.8%` line is the partial-run value and is not an aggregate). The
substantive measure is the per-package aggregate `go test ./internal/factorylane/ -cover` →
**87.4%** (includes pickup.go) plus the lane-local targeted run `go test ./internal/cli/ -run
'TestFactory|TestSD_AC021' -count=1` → `ok` (116.5s). The full-suite aggregate remains CI's
(observed on `origin/develop` at batch push).

### M3 — Lane-direct merge conditions (2026-09-30, owner: manager-develop)

Surfaces: the condition-triple check core `internal/factorylane/merge.go` — `EvaluateMergeTriple`
(conditions a/b/c over a `GitRunner` seam; `GitExitError` carries merge-tree's exit-1-as-verdict),
`MergeCheckRun` records on the shared factory-fallback store (`merge-checks/<lane>/`, one file per
run), `VerifyRunBeforeAcquire`, and the AC-FLA-011 predicate `WindowCoversMerge` — plus the CLI
verbs `moai factory merge ready` (triple → record → window through the existing
`kanban.AcquireIntegrationLock`; every refusal is a verdict and exits 0) and
`moai factory merge gate` (REQ-FLA-011 as a checked property; the negative case is the tested
property). No new serialization mechanism: the integration window is the only one (design.md D3);
this surface never performs a merge.

Re-pin (plan.md M3 pre-flight, Known Issue B2): the t1241 card text re-read from the live queue
this run — 「병합 창 자동화(자율 모드: sync-audit PASS·충돌 없음·HEAD^{tree}=HEAD^2^{tree} 면 로컬
develop 병합)」 — matches the dispatch's documented discipline verbatim. Tree identity carries
BOTH forms: the literal `HEAD^{tree} == HEAD^2^{tree}` rev-parse form on a prepared merge commit
(`--merge-commit`), and the equivalent pre-merge merge-tree form — a clean merge commit's tree is
exactly the merge-tree result and HEAD^2 is the merged branch — with the equivalence verified on
a real throwaway repository (TestMergeTriple_PreMergeFormMatchesLiteralPostMergeForm). The
window's acquire stamp carries RFC3339 second precision, so the pre-acquire proof is evaluated at
the stamp's own resolution; the record-only verifier stays strict, and a failed proof releases
the window again.

| AC | Status | Verification command | Actual output |
|----|--------|---------------------|---------------|
| AC-FLA-009 | PASS (M3 Go + CLI surface) | `go test ./internal/factorylane/ -run 'TestEvaluateMergeTriple\|TestVerifyRunBeforeAcquire\|TestStoreMergeCheckRun\|TestMergeTriple_\|TestExecGitRunner\|TestCheckSyncAudit\|TestRecordMergeCheckRun\|TestLatestMergeCheckRun' -count=1` + `go test ./internal/cli/ -run TestFactoryMergeReady -count=1` | `ok` both (evidence `.moai/state/verify/t1338-m3/e1-ac009.txt` + `e1-ac010.txt`) — every failing condition NAMED: sync-audit on `audit-ready`/empty/missing §E.4, conflict-free on merge-tree exit 1 carrying the conflicted paths, tree-identity on both forms' mismatch; merge-tree tool failure fails closed; the run records all three checks + `failed_condition`; the cleared run predates the acquire stamp |
| AC-FLA-010 | PASS (M3 CLI surface) | `go test ./internal/cli/ -run TestFactoryMergeReady_HeldWindowRefusedWithHolderNamed -count=1` | `ok ... internal/cli` — window pre-held by lane-7 → `verdict: waiting` with the holder NAMED, the lock unchanged (still sess-other/lane-7), the lane's checks recorded, exit 0 |
| AC-FLA-011 | PASS (M3 predicate + CLI) | `go test ./internal/factorylane/ -run TestWindowCoversMerge -count=1` + `go test ./internal/cli/ -run TestFactoryMergeGate -count=1` | `ok` both (evidence `.moai/state/verify/t1338-m3/e1-ac011-negative.txt`) — NEGATIVE CASE: no record → `merge gate: REFUSED — no integration acquire record exists`; stale holder / foreign lane / unparseable timestamp / future acquire each refuse naming the missing property; a live own-lane hold → PROCEED |

M3 E2-E8: builds darwin + `GOOS=windows` OK (pre-commit batch, tree content == `51ad23ae6`);
factorylane package aggregate `go test ./internal/factorylane/ -cover` → **89.0%**; internal/cli
full-suite aggregate GAP (recorded like M1/M2 — lane-local slices only: `go test ./internal/cli/
-run 'TestFactory\|TestSD_AC021' -count=1` → `ok`, 63.4s); boundary greps 0 (AskUserQuestion,
`.Send(`); lint `golangci-lint run --timeout=8m` on both affected packages → **0 issues, exit 0**
(`e5-lint.txt`; one QF1001 finding on the tie clause repaired by variable extraction before this
measurement); guard `TestSD_AC021_LegacySpellingsRefused` PASS (legacy "lead" spellings 0 in the
new surface). RED verbatim: `.moai/state/verify/t1338-m3/red-cycle1.txt` (8 factorylane tests
failing on stubs) + `red-cycle2-3.txt` (5 CLI tests failing on the stub verbs); AC-FLA-011
negative-case verbatim in `e1-ac011-negative.txt`. Commits: `0931c6a8b` (triple), `51ad23ae6`
(verbs) — no push (leader batch). Fixtures build throwaway git repositories under t.TempDir; no
test merges into any real develop branch, and the window is exercised against a throwaway lock
root, never the developer's state.

### M4 — Post-push disposal machine check + `--auto` (2026-09-30, owner: manager-develop)

Surfaces: the origin-landing machine check on BOTH `worktree done` paths (`internal/cli/worktree/
done.go`) — `originLandingRefusal` (git fetch origin develop + `git rev-list --count --left-right
origin/develop...<branch>` over the `landingGitCmd` target-anchored seam, M3 ExecGitRunner
pattern), gated on card branches (`isCardBranch`, the gitflow `WT-` prefix) and ordered AFTER the
L1 tier guard and the anchored-session guard so both fire exactly as before (REQ-FLA-013); no
flag bypasses it (REQ-FLA-015); no CI status is consulted (design D4). Non-card (`feature/*`)
disposal keeps its pre-SPEC behavior — the sync workflow's --auto cleanup runs after a squash PR
merge, where a branch tip is by construction not on the base branch, so an unconditional check
would wrongly refuse that flow. Landing predicate: under the gitflow --no-ff merge discipline the
card branch tip is a parent of the card's merge commit, so the tip's reachability from
origin/develop IS the merge commit's landing; a fetch failure refuses fail-closed.

| AC | Status | Verification command | Actual output |
|----|--------|---------------------|---------------|
| AC-FLA-012 | PASS (M4 Go surface) | `go test ./internal/cli/worktree/ -run 'TestDoneLandingCheck_RefusesUnlandedCardWorktree/manual_unpushed_branch\|TestDoneLandingCheck_RefusesUnlandedCardWorktree/manual_local_only_merge' -count=1` | `ok` (evidence `.moai/state/verify/t1338-m4/green-verbatim.txt`) — unlanded card tree refused with `MERGE_NOT_ON_ORIGIN` + the machine-check output shown (`git rev-list --count --left-right origin/develop...WT-lane-card => "0\t1" (0 left-only / 1 right-only commits)`); local-only --no-ff merge (unpushed) refused the same way; tree survives both |
| AC-FLA-013 | PASS (indirect per acceptance §D.3) | full `go test ./internal/cli/worktree/ -count=1` BEFORE vs AFTER the change + `TestDoneLandingCheck_GuardsOutrankLandingCheck` | BEFORE (tree == `57dcbe857`, pre-change): `ok ... 30.632s`; AFTER: `ok ... 96.155s` (`.moai/state/verify/t1338-m4/e8-full-suite-after.txt`) — existing suite green UNMODIFIED; ordering pin: an unlanded card tree that is ALSO L1 gets `L1_SESSION_WORKTREE` (not the landing refusal), an unlanded card tree with a live anchored session gets `ANCHORED_SESSIONS_PRESENT` — both guards outrank the new check |
| AC-FLA-014 | PASS (M4 Go surface) | `go test ./internal/cli/worktree/ -run TestDoneLandingCheck_LandedCardDisposalCompletes -count=1` | `ok` — after `landCard` (--no-ff merge + push to the local bare origin), `--auto` AND manual disposal complete exit 0 with removal observable (tree gone on both paths); CI-status absence grep-proven: `grep -nE 'exec\.Command\("gh"\|"gh",|gh pr \|gh api \|ciStatus\|ci-status\|CIStatus' internal/cli/worktree/done.go internal/cli/worktree/done_landing_test.go` → **0 matches**; the check runs exactly two git subcommands (`landingGitCmd(targetPath, "fetch", ...)` + `landingGitCmd(targetPath, "rev-list", ...)`), both read-only |
| AC-FLA-015 | PASS (M4 Go surface) | `go test ./internal/cli/worktree/ -run 'TestDoneLandingCheck_RefusesUnlandedCardWorktree/auto_unpushed_branch_refused\|TestDoneLandingCheck_RefusesUnlandedCardWorktree/fetch_failure_refused_fail_closed' -count=1` | `ok` — `--auto` on an unlanded card tree refused (`MERGE_NOT_ON_ORIGIN`, tree survives); fetch failure (origin pointed at a missing remote, work fully merged) → `ORIGIN_LANDING_UNCONFIRMED` fail-closed refusal, tree survives — acceptance §D.2 edge covered |

M4 E2-E8: builds darwin + `GOOS=windows` OK; coverage `go test -cover ./internal/cli/worktree/`
→ **86.4%**, `go test -cover ./internal/factorylane/` → **89.0%** (both ≥ 85; internal/cli
full-suite aggregate GAP recorded like M1-M3 — CI's job at batch push); lint
`golangci-lint run --timeout=8m` on factorylane + cli + worktree → **0 issues, exit 0**
(`.moai/state/verify/t1338-m4/e5-lint.txt`); guard `TestSD_AC021_LegacySpellingsRefused` PASS;
boundary greps 0 (AskUserQuestion, `.Send(`). RED verbatim (pre-GREEN, 4 refusal cells failing
with `nil error = removal proceeded` — done silently removed unlanded card trees on both paths):
`.moai/state/verify/t1338-m4/red-verbatim.txt`. Fixtures build throwaway bare remotes +
repositories under t.TempDir (local "origin", no network); the machine check is never aimed at a
real remote in tests. No push (leader batch).

### M5 — Cross-fragment integration + observability polish (2026-09-30, owner: manager-develop)

Surfaces: the AC-FLA-017 joint three-lane scenario test
(`internal/factorylane/scenario_joint_test.go` — ONE test process simulating
the mixed workload over the existing fragment APIs: lane-fb's fallback switch
logged, lane-a/lane-b contending over a sequential group and a parallel card,
and lane-a's merge proceeding only inside an acquired window; t.TempDir +
FakeClock + the scripted GitRunner + the forged WindowSnapshot test seam — no
real cross-session state, no real window, no real merge) and the count-by-lane
query surface (`Store.TransitionCountsByLane` +
`moai factory fallback --all`, additive: the bare single-lane query keeps its
lane-identity contract, pinned by test). Scope pin held: no doctrine file
touched.

| AC | Status | Verification command | Actual output |
|----|--------|---------------------|---------------|
| AC-FLA-016 | PASS (negative, per acceptance §D.3) | full M1-M5 diff greps: `git diff BASE...HEAD` with BASE = `git merge-base develop HEAD` = `8ea2febe2` (recomputed, not pinned), non-test added lines | 2,362 added non-test diff lines; messaging-delegation send-patterns (`.Send(` / `SendMessage` / `session_msg_send` / `MsgSend` / `SendMessageTo`) on added lines → **0** (positive control: the same pattern matches pre-existing `internal/cli/mcp_factory_msg.go:62`, so the zero is a meaningful absence); `sessionmsg.` usage in added lines = `NewStore`/`DefaultStateRoot`/`AgentInfo` only — the probe's read-only registry read (fragment 1's declared design), no send; card-admission surfaces (`internal/cli/todo*`, `internal/cli/project*`) in the diff → **0 files**. Evidence: `.moai/state/verify/t1338-m5/e4-diff-files-final.txt` + `e4-sessionmsg-usages.txt` |
| AC-FLA-017 | PASS (joint scenario) | `go test ./internal/factorylane/ -run TestJointThreeLaneScenario -count=1 -v` | `ok` (evidence `.moai/state/verify/t1338-m5/joint-scenario.txt`) — one joint log observable: fallback transition logged exactly once (`lane-fb=1`); sequential group gamma ends with exactly 1 holder under concurrency (the loser's re-check names the holder); the parallel pair ends with 2 holders; the merge triple all-passed and recorded, the merge refused with no window ("no integration acquire record exists — the window was never taken"), proceeding inside lane-a's live window (acquired AFTER the checks — the AC-FLA-009 order), and refused for lane-b (the window is per-lane) |

M5 E2-E8: builds darwin + `GOOS=windows` OK; factorylane aggregate coverage
`go test ./internal/factorylane/ -cover` → **88.7%** (includes both new
surfaces); `-race` clean after one repair — the race detector caught a defect
in the NEW scenario helper itself (the holders-now log line read `winners`
outside the mutex; fixed by copying under the lock), verbatim pre-fix warning
+ post-fix `ok` at `.moai/state/verify/t1338-m5/red-race-verbatim.txt` +
`green-factorylane-race.txt`; production code needed NO change (the M1-M4
surfaces came through the full-package `-race` clean — the joint scenario
exposed no M1-M4 defect). internal/cli full-suite aggregate GAP recorded like
M1-M4 (lane-local slice: `go test ./internal/cli/ -run
'TestFactoryFallback|TestFactoryMessaging|TestFactoryPickup|TestFactoryMerge|TestAdopt'
-count=1` → `ok`, 16.1s); lint `golangci-lint run --timeout=8m` on factorylane
+ cli + worktree → **0 issues** (`e5-lint.txt`); vocabulary guard
`TestSD_AC021_LegacySpellingsRefused` PASS; boundary greps 0
(`e4-boundary-greps.txt`). TDD RED verbatim for the query surface:
`red-counts-package.txt` + `red-counts-cli.txt` (stub returning empty → both
new assertions failing; the existing-contract pin passed already), GREEN
`green-counts-*.txt`. Commit: `ad2d752a3` — no push (leader batch).



## §E.3 Run-phase Audit-Ready Signal

```yaml
run_status: audit-ready
run_complete_at: 2026-09-30
run_commit_sha: ad2d752a3
ac_pass_count: 17
ac_fail_count: 0
ac_coverage: >-
  AC-FLA-001..015 evidence complete in §E.2 (M1-M4 sections);
  AC-FLA-016 negative full-diff verification and AC-FLA-017 joint
  three-lane scenario in §E.2 M5 (AC-FLA-017 severity Should per the
  acceptance matrix — delivered, not dropped)
preserve_list_post_run_count: 0
l44_pre_commit_fetch: >-
  HEAD re-read immediately before each commit in this lane
  (4d60234ab for M5 entry; ad2d752a3 landed on WT-lane-autonomy-umbrella);
  no push from the lane
l44_post_push_fetch: >-
  n/a in-lane — the batch push and its post-push fetch are the leader's
  (card discipline: no push from the lane)
new_warnings_or_lints_introduced: 0
cross_platform_build:
  darwin: pass
  windows_amd64: pass
total_run_phase_files: 30
m1_to_mN_commit_strategy: >-
  one commit per milestone unit (M1 ef4c444da, M2 ace746123,
  M3 0931c6a8b+51ad23ae6+57dcbe857, M4 4d60234ab, M5 ad2d752a3) on
  WT-lane-autonomy-umbrella; develop baseline absorbed at fce80341f;
  batch push + CI aggregate owned by the leader
coverage:
  factorylane_aggregate: 88.7%
  internal_cli_full_suite: >-
    GAP recorded (M1-M4 precedent: the full-suite run exceeds go test's
    default 10m timeout — the kill is not a failure and the partial
    coverage line is not an aggregate); lane-local slices green in §E.2;
    the aggregate remains CI's at the batch push
gaps:
  - >-
    internal/cli full-suite aggregate coverage NOT obtained from the lane
    (recorded GAP above; no slot-lease full-suite attempt in M5, following
    the M3/M4 lane-local precedent)
  - >-
    AC-FLA-016 is negative verification by design (acceptance §D.3): its
    evidence is the full-diff grep zero WITH a positive control, not a
    positive test
  - >-
    AC-FLA-013 remains indirect verification (acceptance §D.3): existing
    worktree suite green unmodified, measured at M4
```

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase — owner: manager-docs>_

## §F Phase 4 Mode Selection

**Kickoff gate record**: Implementation Kickoff Approval PASSED in-lane (2026-09-29, operator
answered the lane session's AskUserQuestion directly): run entry approved + progression axis =
autonomous (goal armed after this log). M0 evidence at entry: `d43e50bb3` ancestor of
`origin/develop` AND `SPEC-FACTORY-SELF-DISPATCH-001` present in `origin/develop` tree (both
literal plan.md tests, observed this run). Pre-run baseline absorb: `origin/develop` `8ea2febe2`
merged into the branch as `fce80341f` (incl. t1240 F2 + t1306) so M1-M5 build on landed
self-lease surfaces. Phase 1 plan-audit iter2 scheduled at entry (post-audit polish changed
plan.md/research.md bytes → skip-eligibility condition 3 fails → re-execute per the single
authoritative skip contract).

**Input parameters**: tier L; scope ~15-25 files across internal/cli/factory,
internal/cli/worktree, internal/sessionmsg, internal/kanban; domain count 1 (one coherent
factory-operations subsystem, Go-dominant + doctrine doc touches); concurrency benefit LOW
(coding-heavy); agent-teams prereqs not requested.

**Mode evaluation**:

| Mode | Selected | Rationale |
|---|---|---|
| direct | no | multi-file feature work, no trivial single-line shape |
| fanout | no | coding-heavy — Anthropic coding-task parallelism caveat |
| sweep | no | semantic new code, multi-rule, inter-file dependency |
| serial | **YES** | per-milestone manager-develop spawns, Section A-E template |

**Decision: serial**

**Justification**: coding-heavy implementation across a cohesive subsystem; milestones M1-M5
share one package graph and one writer tree (single card worktree), so sequential per-milestone
delegation is the simpler correct envelope. Tier L auto-routing to manager-lead was considered
(the ≥3-milestone AND ≥10-file predicate is met on paper) and declined per the §B.2 boundary
default toward the simpler mode: no cross-domain fan-out is warranted — the fragments consume
one boundary surface each and are implemented against one tree. Boundary case note: file count
and milestone count sit above their thresholds while domain count sits below; the simpler-mode
tie-breaker resolves to serial.
