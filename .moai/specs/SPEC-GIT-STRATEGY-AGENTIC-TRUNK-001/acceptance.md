# Acceptance — SPEC-GIT-STRATEGY-AGENTIC-TRUNK-001

Document-level pin: every observation below was taken on tree `f4175587119b5cfeb8d8348fc74cada8764e89ed` (branch `WT-agentic-trunk`, plan-audit iteration 1 repair). That commit differs from `6041c38a4` (the tree of research.md) in the six SPEC files only, so every code, template, config and guidance observation holds on both. A criterion-level pin wins where one is given.

Iteration 2 repair (version 0.3.0): the cells it re-observed carry the criterion-level pin `c6e590507b39e27acc3305a3a1aa3dc488aef0c0` (AC-AGT-016 P4, AC-AGT-020 Judge 6, AC-AGT-021 Judges B and C, AC-AGT-023 Judges 3 to 5, ledger E19 and E22 to E25). `git diff --stat f4175587119b5cfeb8d8348fc74cada8764e89ed..c6e590507b39e27acc3305a3a1aa3dc488aef0c0 -- internal .claude AGENTS.local.md docs-site .moai/config .moai/docs` prints nothing (exit 0), so no code, template, config or guidance path moved between the two pins.

Iteration 3 repair (version 0.3.1, defect D1 only): AC-AGT-020 Judge 6 and its two failing controls were re-observed on tree `45636352d8716b57b0284c24318ad0399d606309` (ledger E26); no other criterion changed.

Rescope (version 0.4.0, 2026-10-21): AC-AGT-005, AC-AGT-006, AC-AGT-011, AC-AGT-016, AC-AGT-017, AC-AGT-018 (Judge 5) and AC-AGT-023 (Judge 2) were re-authored for the target-holder routing; their fresh RED-now cells were observed on tree `a6c69e154b62eb33763212a3034e505bd1046fe1` (branch `WT-agentic-trunk`, M0-M5 implemented, final pre-rescope HEAD) and are carried in ledger E27 to E32. All other criteria keep their original pins; the code, template, config and guidance observations they carry are unaffected by the rescope (the rescope edited SPEC artifacts only). AC-AGT-016 P4a and P4b are leader-time probes; this pass observed the worktree-side state only (E30) and the leader re-measures both from the primary checkout at cutover, as the criterion states.

Amendment (2026-10-24, leader-directed in-place amendment; spec.md `amendment_of` + HISTORY `## Amendments`): AC-AGT-010's judge is scoped with one exception — the home-root resolver reads at `internal/cli/worktree/list.go:103` and `internal/cli/worktree/migrate.go:286` (`LoadGitFlowIntegrationConfig(...).IntegrationTarget`), introduced by absorbed #1858 — which fall under follow-up card t1949's invariant re-scope. No other criterion changed. (This document carries no frontmatter, so the amendment date is recorded here.)

## Conventions

- A judge is a single invocation. A Go-test judge is `go test <pkg> -run '^(TestA|TestB)$' -count=1 -v`; it is green only when stdout carries each named `--- PASS:` line, exit is 0 and `[no tests to run]` is absent. A selector that matches nothing prints `ok ... [no tests to run]` with exit 0 (ledger E01), so exit 0 alone is never a verdict.
- Each Go criterion has a flip judge that selects only NEW test names (red now: `[no tests to run]`; green: the named `--- PASS:` lines) and, where an existing test is in scope, a separate guard judge that selects only EXISTING names (green now, must stay green). A mixed selector is not used: a selector naming one existing and one new test prints a passing line today (observed in plan-audit iteration 1 for AC-AGT-001), so its red would be a missing second line, not a failing run.
- Classes. Flip criteria are release-blocking and carry the four RED-now elements (command, verbatim stdout, exit code, tree SHA) either inline or as a cited ledger entry E01 to E32 (end of this file). Regression guards (AC-AGT-002, AC-AGT-010, the guard judges, AC-AGT-014 judge B, AC-AGT-021 judge A, AC-AGT-024) are green today by design and say so. AC-AGT-022 is judged in the sync phase. AC-AGT-016 is executed by the leader (below).
- Output order of a multi-file `grep -c` or `grep -l` is not stable on this machine; judges read the labeled per-file lines, never their order.
- The doctor exits 0 whether a row is `ok` or `warn` (warnings are advisory, ledger E18), so a doctor-based judge reads the named row and never the exit code.
- Quoted counts and rows are observations of the pinned tree; M0 re-checks the pinned paths against the run-phase start tree (`git diff --stat f4175587119b5cfeb8d8348fc74cada8764e89ed..HEAD -- internal .claude AGENTS.local.md docs-site .moai/config .moai/docs`) and re-observes only a selector whose package changed.

## A. Value set and reader (M1)

### AC-AGT-001 (maps REQ-AGT-001, REQ-AGT-002)
Given the allowed-workflow set and the disposition classifier,
When the tests run,
Then `agentic-trunk` is the fifth and last allowed value, classifies valid and not git-flow, `Agentic-Trunk`, `agentic_trunk`, `agentic-trunk ` and the empty string classify invalid, and `git-flow`, `github-flow`, `gitlab-flow`, `release-flow` classify as before.
- Flip judge: `go test ./internal/config -run '^TestAgenticTrunkValueSet$' -count=1 -v`. Pass stdout: `--- PASS: TestAgenticTrunkValueSet (`; exit 0.
- Guard judge (green now): `go test ./internal/config -run '^TestWorkflowDisposition$' -count=1 -v`. Observed: `--- PASS: TestWorkflowDisposition (0.02s)` with six subtests including `allowed_set_is_exactly_four_flows`, then `ok  github.com/modu-ai/moai-adk/internal/config  0.216s`, exit 0. M1 updates that subtest to five values (an update to the new truth, never a skip); the guard must print `--- PASS: TestWorkflowDisposition (` after M1.
- RED-now (flip): ledger E01, exit 0, `[no tests to run]`. Why red: the test and the value do not exist. Iteration 1 recorded that the earlier mixed selector printed a PASS today; that selector is retired.
- Green path: M1.

### AC-AGT-002 (maps REQ-AGT-001)
Given the existing git-flow and github-flow reader tests,
When the compatibility pack runs after M1 and again after M2,
Then every named test passes with no assertion weakened: `TestWorkflowTargetResolution`, `TestEmptyTargetGuidance`, `TestEmptyTargetGuidanceResolvedTargetIsSilent`, `TestTargetProvenance`.
- Judge: `go test ./internal/config -run '^(TestWorkflowTargetResolution|TestEmptyTargetGuidance|TestEmptyTargetGuidanceResolvedTargetIsSilent|TestTargetProvenance)$' -count=1 -v`
- Pass stdout: four top-level `--- PASS:` lines; exit 0.
- Observed now (regression guard, green by design; subtest lines elided, tree f4175587): `--- PASS: TestEmptyTargetGuidance (0.01s)`, `--- PASS: TestEmptyTargetGuidanceResolvedTargetIsSilent (0.00s)`, `--- PASS: TestTargetProvenance (0.00s)`, `--- PASS: TestWorkflowTargetResolution (0.00s)`, then `ok  github.com/modu-ai/moai-adk/internal/config  0.188s`, exit 0. It is the compatibility witness: it must stay green, and M2 may change one expected guidance string only by adding the `agentic-trunk` alternative.

### AC-AGT-003 (maps REQ-AGT-003, REQ-AGT-004)
Given a manual profile with `workflow: agentic-trunk` and `develop_branch: develop`, and a personal and a team profile with the same workflow,
When the strategy reader loads each,
Then the manual profile reports integration target `main` and develop-branch seam `main`; personal and team report integration target `main` and an empty develop-branch seam; `environment` and `release_branch_prefix` values change nothing.
- Judge: `go test ./internal/config -run '^TestAgenticTrunkTarget$' -count=1 -v`
- Pass stdout: `--- PASS: TestAgenticTrunkTarget (` (subtests for manual, personal, team); exit 0.
- RED-now: ledger E02, exit 0, `[no tests to run]`. Why red: the test and the behavior do not exist.
- Green path: M1.

### AC-AGT-004 (maps REQ-AGT-005)
Given the reader's predicates,
When the table runs (git-flow manual, agentic-trunk manual, agentic-trunk team, github-flow manual, absent file),
Then the local-integration predicate is true for the first two only, and the git-flow-only predicate is true for the first only.
- Judge: `go test ./internal/config -run '^TestAgenticTrunkLocalIntegrationPredicate$' -count=1 -v`
- Pass stdout: `--- PASS: TestAgenticTrunkLocalIntegrationPredicate (`; exit 0.
- RED-now: ledger E03, exit 0, `[no tests to run]`. Why red: the predicate does not exist.
- Green path: M1.

## B. Consumers (M2)

### AC-AGT-005 (maps REQ-AGT-006; rescope 0.4.0: target-holder routing)
Given a project whose manual workflow is `agentic-trunk`,
When `factory complete` and `factory merge ready` run on a card that cleared the readiness triple,
Then the delivery routes by the target holder: where a worktree other than the primary checkout holds `main`, the integration window is taken and recorded, no `gh pr` call is made, and no pull request is opened; where only the primary checkout holds `main`, the card's pull request is opened, `pr-open` is recorded, and NO pull-request merge request of any kind is issued (no `--auto`; the leader merges, and a later completion run records `merged-pr`).
- Pin judge A (dedicated-worktree route, implemented M2, green): `go test ./internal/cli -run '^(TestAgenticTrunkFactoryCompleteTakesWindow|TestAgenticTrunkMergeReadyTakesWindow)$' -count=1 -v`. Pass stdout: two `--- PASS:` lines; exit 0. The fixtures build a dedicated integration worktree holding `main`, so the routing keeps them on the local-merge path.
- Guard judge (green): `go test ./internal/cli -run '^TestFactoryCompleteWindowGitFlowUnchanged$' -count=1 -v`. Observed green at M0 (progress record §E.2: `--- PASS: TestFactoryCompleteWindowGitFlowUnchanged (1.69s)`, exit 0); must stay green.
- Flip judge (primary route, NEW tests, red now): `go test ./internal/cli -run '^(TestAgenticTrunkCompletePrimaryPrEdge|TestAgenticTrunkCompleteNoAutoMerge)$' -count=1 -v`. Pass stdout: two `--- PASS:` lines; exit 0.
- RED-now (flip, tree a6c69e154b, ledger E31): exit 0, `[no tests to run]`. Why red: the two tests do not exist, and today the completed M2 code takes the local-merge route unconditionally for `agentic-trunk` — a primary-only project would refuse at the no-worktree-holds-main check instead of opening the card's pull request, and the pr-open call still carries `--auto` (observed verbatim at `internal/cli/factory_card_pr.go:444`: `factoryGHCall(wt, "pr", "merge", strconv.Itoa(pr.Number), "--auto", mergeFlag, "--match-head-commit", checkedTip)`).
- Green path: M5b.

### AC-AGT-006 (maps REQ-AGT-007; rescope 0.4.0: primary route skips)
Given session-exit auto-merge enabled and manual `agentic-trunk`,
When a clean session exit has commits ahead of `main`,
Then the branch merges into `main` under the same guards as manual git-flow where a worktree other than the primary checkout holds `main`; where only the primary checkout holds `main`, the auto-merge skips with a notice naming the missing dedicated landing surface; for a non-manual agentic-trunk project the skip notice names both workflows.
- Flip judge A (dedicated-worktree route, implemented M2, green): `go test ./internal/cli -run '^TestAgenticTrunkAutoMerge$' -count=1 -v`. Pass stdout: `--- PASS: TestAgenticTrunkAutoMerge (`; exit 0.
- Flip judge B (primary route, NEW test, red now): `go test ./internal/cli -run '^TestAgenticTrunkAutoMergeSkipsPrimaryTarget$' -count=1 -v`. Pass stdout: `--- PASS: TestAgenticTrunkAutoMergeSkipsPrimaryTarget (`; exit 0.
- Guard judge (green): `go test ./internal/cli -run '^(TestAutoMergeHappyPath|TestAutoMergeUnconfiguredTarget)$' -count=1 -v`. Observed green at M0 (progress record §E.2); `TestAutoMergeUnconfiguredTarget` was updated in M2 if it pinned the old notice string.
- RED-now (flip B, tree a6c69e154b, ledger E31): exit 0, `[no tests to run]`. Why red: the test does not exist and the M2 gate (`!UsesLocalIntegration() || DevelopBranch == ""`, `session_worktree_automerge.go:176`) merges into whatever worktree holds `main` — on the primary route it would write into the shared primary checkout's working tree.
- Green path: M5b (judge B), M2 (judge A, already green).

### AC-AGT-007 (maps REQ-AGT-008)
Given manual `agentic-trunk` and a worktree holding `main`,
When `moai integration acquire --name lane --card c1` runs from a card tree without `--branch`,
Then the recorded window names the worktree holding `main`, stderr carries no caller-fallback warning, and the git-flow and github-flow acquire behaviors are unchanged.
- Flip judge: `go test ./internal/cli -run '^TestAgenticTrunkIntegrationAcquire$' -count=1 -v`. Pass stdout: `--- PASS: TestAgenticTrunkIntegrationAcquire (`; exit 0.
- Guard judge (green now): `go test ./internal/cli -run '^(TestIntegrationAcquire_GitFlowEmptyDevelopWarnsOnCallerFallback|TestIntegrationAcquire_GitHubFlowEmptyDevelopDoesNotWarn)$' -count=1 -v`. Observed: both `--- PASS:` lines (0.18s, 0.21s), `ok  github.com/modu-ai/moai-adk/internal/cli  1.189s`, exit 0.
- RED-now (flip): ledger E11, exit 0, `[no tests to run]`. Why red: the test does not exist.
- Green path: M2.
- Rescope 0.4.0 note: UNCHANGED. REQ-AGT-008's acquire behavior is the dedicated-worktree route; the re-aimed requirement's primary-route clause (no window in the primary's working tree) is exercised by the primary-route completion tests of AC-005 (a primary-route `factory complete` takes no window), so no separate judge is added.

### AC-AGT-008 (maps REQ-AGT-009)
Given an invalid workflow value and a non-manual git-flow profile,
When the diagnostics render,
Then the allowed-set listing contains all five values and the git-flow fix strings name `agentic-trunk (target main)` beside `github-flow (target main)`.
- Judge: `go test ./internal/config -run '^TestAgenticTrunkGuidance$' -count=1 -v`
- Pass stdout: `--- PASS: TestAgenticTrunkGuidance (`; exit 0.
- RED-now: ledger E04, exit 0, `[no tests to run]`. Why red: the test does not exist and `AllowedWorkflows()` lists four values.
- Green path: M2.

### AC-AGT-009 (maps REQ-AGT-010)
Given a manual `agentic-trunk` profile that still carries `develop_branch`, `release_branch_prefix` and `rc_version_format`, plus `lead_push_threshold: 20` and `merge_method: merge`,
When the reader and the doctor run,
Then no diagnostic mentions the three keys, the target is `main`, and `lead_push_threshold` and `merge_method` read through unchanged.
- Judge: `go test ./internal/cli -run '^TestAgenticTrunkIgnoredKeys$' -count=1 -v`
- Pass stdout: `--- PASS: TestAgenticTrunkIgnoredKeys (`; exit 0.
- RED-now: ledger E12, exit 0, `[no tests to run]`. Why red: the test does not exist.
- Green path: M2.

### AC-AGT-010 (maps REQ-AGT-011)
Given the landing-check code,
When searched for a workflow-value read,
Then it contains none, EXCEPT the home-root resolver reads introduced by absorbed #1858 — `internal/cli/worktree/list.go:103` and `internal/cli/worktree/migrate.go:286` (the `config.LoadGitFlowIntegrationConfig(...).IntegrationTarget` assignments) — which are carved out of this judge's scope and fall under follow-up card t1949's invariant re-scope.
- Judge: `git grep -c -E 'LoadGitFlowIntegrationConfig|WorkflowGitFlow|WorkflowGitHubFlow|WorkflowAgenticTrunk|"agentic-trunk"' -- internal/cli/worktree` (interpreted under the exception above: a match confined to `list.go:103` / `migrate.go:286` is not a judge failure; any other match in `internal/cli/worktree` is)
- Pass stdout: empty; exit 1 (no match).
- Observed (tree f4175587, regression guard): stdout empty, exit 1. Why already green: the invariant is fixed before M2 can break it; it is not a flip criterion. Positive control, observed: `git grep -c LoadGitFlowIntegrationConfig -- internal/cli/factory_card.go` printed `internal/cli/factory_card.go:3`, exit 0, so the pattern fires where the identifier exists.
- Green path: unchanged by M2; failing means M2 added a workflow read to the landing checks.

## C. Display (M3)

### AC-AGT-011 (maps REQ-AGT-012)
Given manual `agentic-trunk`,
When the doctor check "Git Strategy Workflow" runs,
Then status is OK and the message contains `agentic-trunk`, `main`, `serial`, and `batch`.
- Flip judge: `go test ./internal/cli -run '^TestAgenticTrunkDoctor$' -count=1 -v`. Pass stdout: `--- PASS: TestAgenticTrunkDoctor (`; exit 0.
- Guard judge (green now): `go test ./internal/cli -run '^TestDoctorGitStrategyWorkflow$' -count=1 -v`. Observed: `--- PASS: TestDoctorGitStrategyWorkflow (0.08s)`, `ok  github.com/modu-ai/moai-adk/internal/cli  1.470s`, exit 0.
- RED-now (flip): ledger E13, exit 0, `[no tests to run]`. Why red: the test does not exist; the installed build prints a `warn` row for the value (E18).
- Green path: M3. A cutover-time judgment by the installed binary is in AC-AGT-016 (P3).
- Rescope 0.4.0 note (M5b): the standing-branch phrase is reworded holder-neutral (design.md section 2, `workflowStandingBranches` row) — cards land serially through the integration-window merge OR the card's pull request; a leader release batch per `lead_push_threshold` landings. The pinned tokens `agentic-trunk`, `main`, `serial`, `batch` stay in the message, so this judge's pass condition is unchanged; `TestAgenticTrunkDoctor` is updated to the new truth in M5b if it pins the full phrase, never skipped (REQ-AGT-024).

### AC-AGT-012 (maps REQ-AGT-013)
Given a statusline render with a branch,
When the workflow is `agentic-trunk` and `git-flow`, the label appears after the branch segment; when it is `github-flow`, absent file, unreadable file or an invalid value, the output is byte-equal to the pre-change golden, and the fake git repository records the same number of calls as before.
- Judge: `go test ./internal/statusline -run '^(TestAgenticTrunkStatuslineLabel|TestAgenticTrunkStatuslineAbsent|TestAgenticTrunkStatuslineNoExtraGit)$' -count=1 -v`
- Pass stdout: `--- PASS: TestAgenticTrunkStatuslineLabel (`, `--- PASS: TestAgenticTrunkStatuslineAbsent (` and `--- PASS: TestAgenticTrunkStatuslineNoExtraGit (`; exit 0.
- RED-now: ledger E15, exit 0, `[no tests to run]`. Why red: none of the three tests exists (all three names are new, so a mixed-selector trap cannot occur).
- Scope note for SPEC-PRIMARY-LANDING-ONLY-001 REQ-PLO-009 (byte-identical render whenever the checkout has a branch): `TestAgenticTrunkStatuslineAbsent` pins exactly the github-flow and unset cases that REQ-PLO-009 measures, so the two changes compose; only a valid non-github-flow value, the one this repository will carry after M6, shows the label.
- Green path: M3.

## D. Pins and guards (M4)

### AC-AGT-013 (maps REQ-AGT-014)
Given a project whose manual workflow is `agentic-trunk` and a template whose manual workflow is `github-flow`,
When `moai update` reconciles `git-strategy.yaml`,
Then the manual workflow is still `agentic-trunk` afterward.
- Flip judge: `go test ./internal/cli/update -run '^TestAgenticTrunkValueSurvivesUpdate$' -count=1 -v`. Pass stdout: `--- PASS: TestAgenticTrunkValueSurvivesUpdate (`; exit 0.
- Guard judge (green now): `go test ./internal/cli/update -run '^TestUpdate_GitStrategyValuesSurvive$' -count=1 -v`. Observed: `--- PASS: TestUpdate_GitStrategyValuesSurvive (0.01s)`, `ok  github.com/modu-ai/moai-adk/internal/cli/update  0.281s`, exit 0.
- RED-now (flip): ledger E16, exit 0, `[no tests to run]`. Why red: the test does not exist.
- Green path: M4.

### AC-AGT-014 (maps REQ-AGT-015)
Given the shipped defaults,
When read,
Then the built-in default configuration and the template yaml carry `github-flow` for all three profiles and no `workflow: agentic-trunk` line.
- Judge A (flip): `go test ./internal/config -run '^TestAgenticTrunkDistributedDefaultsUnchanged$' -count=1 -v`. Pass stdout: `--- PASS: TestAgenticTrunkDistributedDefaultsUnchanged (`; exit 0.
- Judge B (regression guard): `grep -c 'workflow: agentic-trunk' internal/template/templates/.moai/config/sections/git-strategy.yaml.tmpl`. Pass stdout `0`, exit 1.
- RED-now: A ledger E05, exit 0, `[no tests to run]`; B observed on tree f4175587: stdout `0`, exit 1 (already satisfied by design; it must stay `0` at every milestone). Why A is red: the pin test does not exist.
- Green path: A flips at M4; B must stay `0` at every milestone.

### AC-AGT-015 (maps REQ-AGT-016)
Given the sweep guard's arming predicate `sweepGuardArmed`,
When fed an `agentic-trunk` root, a `git-flow` root and a `github-flow` root, with `MOAI_SWEEP_GUARD_ARMED` cleared in the test,
Then only the github-flow root arms it.
- Flip judge: `go test ./internal/template -run '^TestAgenticTrunkSweepGuardNotArmed$' -count=1 -v`. Pass stdout: `--- PASS: TestAgenticTrunkSweepGuardNotArmed (` with one subtest per root (`agentic-trunk`, `git-flow`, `github-flow`); exit 0.
- Guard judge (green now, NOT the flip judge): `go test ./internal/template -run '^TestGitHubFlowSweepGuard$' -count=1 -v`. Observed: `--- PASS: TestGitHubFlowSweepGuard (7.42s)` with subtests `arming-switch` and `F-empty-walk`, the log line `sweep guard disarmed: git_strategy workflow="git-flow" (armed only when "github-flow")`, `ok  github.com/modu-ai/moai-adk/internal/template  7.741s`, exit 0. It prints a PASS today, which is why iteration 1 rejected it as the judge. It must still print `--- PASS: TestGitHubFlowSweepGuard (` after M6 flips the live value, and stays disarmed there.
- RED-now (flip): ledger E17, exit 0, `[no tests to run]`. Why red: the dedicated test does not exist.
- Green path: M4. Mutant check: making `sweepGuardArmed` return true for `agentic-trunk` turns the `agentic-trunk` subtest red.

## E. Cutover (M6, executed by the leader)

### AC-AGT-016 (maps REQ-AGT-017; rescope 0.4.0: the bridge is the starting state)
Executed by the leader, with the operator's confirmation, after t1831 lands and the rc build is installed: replacing the installed binary affects every session on the machine. The starting state of the flip is the bridge value `github-flow` (leader flip, 2026-10-21; the installed build `8249bff8c3` predates this branch and lacks the value). The precondition set below is the single list; REQ-AGT-017, plan.md M6 and design.md section 4 carry the same P1 to P5.

Given the flip commit,
When the commit graph, the installed binary and the running servers are inspected,
Then P1 to P5 all hold, with the flip changing the manual profile from the bridge `github-flow` to `agentic-trunk`.
- P1, ancestry. Judges (one invocation per ancestor commit, each exit 0, stdout empty): `git merge-base --is-ancestor <Mx-sha> <flip-sha>` for each of M1, M2, M3, M4, M5 and M5b. The flip commit touches one path: `git diff-tree --no-commit-id --name-only -r <flip-sha>` prints `.moai/config/sections/git-strategy.yaml` only. Judge-form control, observed on this tree: `git merge-base --is-ancestor f4175587119b5cfeb8d8348fc74cada8764e89ed 6041c38a4` exits 1 and the reverse pair exits 0, so the form separates a wrong order from a right one.
- P2, the installed binary is the tree build of the flip commit's parent. Judge: `sh scripts/verify-local-install.sh bin/moai /Users/goos/go/bin/moai <flip-parent-short-sha>`; pass: exit 0 and a final line `local-install-check: OK — /Users/goos/go/bin/moai is byte-identical to bin/moai and reports commit <flip-parent-short-sha>`.
- P3, the installed binary understands the value. Judge: `moai doctor` (the installed binary by name, on purpose) with the working directory set to a fixture project whose only file is `.moai/config/sections/git-strategy.yaml` holding `git_strategy:` / `mode: manual` / `manual:` / `workflow: agentic-trunk` / `worktree_base_branch: main` (two-space indents). Pass: the `Git Strategy Workflow` row has status `ok` and its message contains `agentic-trunk`, `serial` and `batch`.
- P4, no stale server anywhere on the machine. Two probes, because the doctor row reads only the records under one project directory (`<project>/.moai/state/mcp-server`, `internal/cli/mcp_server_runtime.go`) and a server recorded elsewhere, or none, is invisible to it.
  - P4a, the recorded servers. Judge (from the primary checkout): `moai doctor --check "MCP Server Version"`. Pass: an `ok` row reading `N running MCP server(s) match the installed binary (<flip-parent-short-sha>)`; the stale state is a `warn` row `running MCP server is stale (...)`.
  - P4b, every running server (recorded or not, any project directory). Three plain invocations, read in order: `pgrep -o -f '^moai mcp-server$'` prints the pid of the OLDEST running server (the anchors keep the invoking shell, whose command line also contains the words, out of the match); `ps -o pid=,lstart=,command= -p <that-pid>` prints its start time; `stat -f '%Sm' -t '%Y-%m-%dT%H:%M:%S' /Users/goos/go/bin/moai` prints the install time. Pass: the oldest server's start time is later than the install time, so every server started on the new binary. When it is not, each older server is either reconnected, or, only for a server whose project directory differs from this repository (`ps eww -p <pid>` shows `CLAUDE_PROJECT_DIR=`), left running and named in the progress record with its pid and project directory; the accepted residual risk is in this criterion's Residual-risk line. The leader reconnects servers by the Claude Code `/mcp` reconnect or a session restart and records it in the progress record.
  - Known-failing input for P4b (so the probe is observed red, ledger E22): the same three reads with a reference time later than the oldest server's start (a file touched at observation time stands in for a reinstall) order the other way and fail the pass condition.
  - Residual-risk: a left-running server of another project that is later asked to act on this repository by an explicit `project_root` argument still runs the old reader and refuses `agentic-trunk`; the probe lists such servers, it does not reconnect them.
- P5, execution condition (rescope 0.4.0: the integration window no longer exists on the pull-request route). The leader records in the progress record: the operator's confirmation; an open-pull-request read of this repository showing no card pull request in `pr-open` awaiting the leader's merge; and that no `factory complete` run is mid-transition. A state witness, not a flip judge.
- RED-now (tree a6c69e154b, observed 2026-10-21, this machine, installed build `v3.2.0-rc.37`, commit `8249bff8c3`, installed 2026-10-19T21:47:22; ledger E27 to E30): P1 no flip commit exists, `git log --format=%h -n 1 -S'workflow: agentic-trunk' -- .moai/config/sections/git-strategy.yaml` prints nothing, exit 0 (E27); P2 `sh scripts/verify-local-install.sh` prints `local-install-check: FAIL — installed binary differs from bin/moai`, exit 1 (E28; the tree build exists from the M0-M5 run and the installed bridge build differs — red for the right reason: the installed binary is not the flip parent's build, which does not exist yet); P3 (E29) the installed build's doctor, run on the fixture project, prints `warn    Git Strategy Workflow  git_strategy workflow = "agentic-trunk" is not one of the allowed flows (github-flow, git-flow, gitlab-flow, release-flow) — repair by editing .moai/config/sections/git-strategy.yaml` and exits 0 — red because the installed build predates the branch; P4a (E30) from this worktree the row reads `ok  MCP Server Version  no running moai MCP server recorded` (records resolve per project directory; the primary-side row is leader-time at cutover — a Gap of this pass, named in the rescope header); P4b leader-time (E22 holds the method and the constructed known-failing input); P5 leader-time (a witness).
- Why this gate can show its hazard: P3 turns red on a build that lacks M1 to M5b (observed today, E29), P4a turns red on a recorded server that was not reconnected after the reinstall, and P4b turns red on any running server, recorded or not, that predates the reinstall. A flip before M5b, or with the bridge build still installed, fails P1, P2 or P3.
- Green path: M6. The commit graph is the only ordering evidence (verification-claim-integrity section 2.3).

### AC-AGT-017 (maps REQ-AGT-018; rescope 0.4.0: the bridge is the starting state)
Given the live yaml after the flip,
When searched,
Then the manual profile carries `workflow: agentic-trunk`, personal and team keep `github-flow`, `lead_push_threshold: 20` remains, and the three git-flow-only keys and the standing-flow comment are gone.
- Judge 1: `grep -c 'workflow: agentic-trunk' .moai/config/sections/git-strategy.yaml` (expect `1`).
- Judge 2: `grep -c 'workflow: github-flow' .moai/config/sections/git-strategy.yaml` (expect `2` — personal and team; the manual profile has left the bridge value).
- Judge 3: `grep -c -E 'lead_push_threshold: 20' .moai/config/sections/git-strategy.yaml` (expect `1`).
- Judge 4: `grep -c -E 'develop_branch:|release_branch_prefix:|rc_version_format:|stays git-flow|stays github-flow' .moai/config/sections/git-strategy.yaml` (expect `0`, exit 1 — the three keys, the pre-bridge stale comment, and any bridge-worded comment the leader's 2026-10-21 commit left).
- RED-now (tree a6c69e154b, observed 2026-10-21, ledger E27): Judge 1 stdout `0`, exit 1; Judge 2 stdout `2`, exit 0 (personal and team; this tree's committed manual profile still reads `git-flow` — the leader's bridge flip is not yet a commit this tree can see, so the count will read 3 between the bridge commit landing and M6, then 2 after the flip); Judge 3 stdout `1`, exit 0; Judge 4 stdout `4`, exit 0 (the three keys plus the line-28 comment `stays git-flow`). Why red: the value has not been flipped. Judges 2 and 3 are green now by design; they guard personal, team and the threshold against a careless edit. Residual: an unpredicted bridge-comment wording is not in Judge 4's alternation; the M6 diff review reads the one-file commit, which is the second layer.
- Green path: M6.

## F. Guidance (M5, M7, M8)

### AC-AGT-018 (maps REQ-AGT-019)
Given the repo-local guidance,
When searched,
Then the protocol H1 and the AGENTS.local.md section 4.0 heading name `agentic-trunk`, both files state the four steps, the protocol path is unchanged, no live line calls this repository's model "git-flow Lane Protocol" or "switched from GitHub Flow to git-flow", `repo-local-pr-policy.md` names the model, and the three `.moai/docs` notices name it.
- Judge 1: `grep -c -E '^# agentic-trunk Lane Protocol' .claude/rules/local/gitflow-lane-protocol.md` (expect `1`).
- Judge 2: `grep -c -E 'git-flow Lane Protocol|switched from GitHub Flow to git-flow' .claude/rules/local/gitflow-lane-protocol.md AGENTS.local.md` (expect `0` per file, exit 1).
- Judge 3: `git ls-files .claude/rules/local/gitflow-lane-protocol.md` (prints the path, exit 0).
- Judge 4: `grep -c -E '^### §4\.0 .*agentic-trunk' AGENTS.local.md` (expect `1`).
- Judge 5 (the four-step statement, rescope 0.4.0 sentinel, shared with AC-AGT-023): `grep -c -F 'agentic-trunk: lane worktree, serial landing into main, leader release batch per lead_push_threshold, tagged release' .claude/rules/local/gitflow-lane-protocol.md AGENTS.local.md` (expect at least `1` per file). The pre-rescope sentinel (`serial local landing, batch pull request`) is retired with the dedicated-worktree-only model; the routing sentence that follows the sentinel in each file names the target-holder split.
- Judge 6: `grep -c 'Card workflow (git-flow transition' .claude/rules/local/repo-local-pr-policy.md` (expect `0`) and `grep -c agentic-trunk .claude/rules/local/repo-local-pr-policy.md .moai/docs/git-workflow-doctrine.md .moai/docs/git-local-workflow-doctrine.md .moai/docs/gitflow-integration-chain.md` (expect at least `1` per file).
- RED-now (Judge 5 re-observed on tree a6c69e154b for the 0.4.0 sentinel, ledger E32: both files `:0`, exit 1. Judges 1 to 4 and 6 keep their original pin-tree observations (tree f4175587, observed): Judge 1 stdout `0`, exit 1; Judge 2 stdout `.claude/rules/local/gitflow-lane-protocol.md:2` and `AGENTS.local.md:0`, exit 0 (the H1 `# git-flow Lane Protocol (moai-adk-go, local-only)` and the blockquote line match); Judge 3 prints the path, exit 0 (green now, must stay); Judge 4 stdout `0`, exit 1; Judge 6 first command stdout `1`, exit 0, second command `0` for all four files, exit 1. Why red: the name has not been applied.
- Green path: M7.

### AC-AGT-019 (maps REQ-AGT-020)
Given the six template rule files, the git-workflow skill, `delivery.md` and the template `workflow.yaml` comment,
When searched,
Then each names `agentic-trunk`, and each still names `github-flow` as the distributed default.
- Judge (one invocation, expect at least `1` per labeled line): `grep -c agentic-trunk internal/template/templates/.claude/rules/moai/workflow/factory-dispatch-mechanics.md internal/template/templates/.claude/rules/moai/workflow/factory-dispatch.md internal/template/templates/.claude/rules/moai/workflow/factory-dispatch-core.md internal/template/templates/.claude/rules/moai/workflow/worktree-integration.md internal/template/templates/.claude/rules/moai/workflow/delivery-policy.md internal/template/templates/.claude/rules/moai/workflow/spec-workflow.md internal/template/templates/.claude/skills/moai-ref-git-workflow/SKILL.md internal/template/templates/.claude/skills/moai/workflows/sync/delivery.md internal/template/templates/.moai/config/sections/workflow.yaml`
- Default-preserved judge (counts may rise, never fall): `grep -c github-flow` over the same first eight files must print at least `factory-dispatch-mechanics.md:2`, `factory-dispatch.md:1`, `factory-dispatch-core.md:1`, `delivery.md:12`, and `0` or more for `worktree-integration.md`, `delivery-policy.md`, `spec-workflow.md` and `SKILL.md` (counts observed on tree f4175587, exit 0).
- Named exclusions (template files that match the flow-name pattern and are deliberately not edited): `auto-semantics.md` line 376 and `spec-workflow-detail.md` line 31 describe github-flow and `merge_method` only; `manager-git.md` and `manager-git.toml` are agent definitions, outside REQ-AGT-020, which scopes rules and skills; `tab_schema.json` is the wizard schema (OQ-5); `git-strategy.yaml.tmpl` keeps its defaults (REQ-AGT-015).
- RED-now (tree f4175587, observed): the single `grep -c agentic-trunk` invocation prints `:0` for each of the nine labeled files, exit 1.
- Green path: M5.

### AC-AGT-020 (maps REQ-AGT-020)
Given `sync/delivery.md`,
When searched,
Then the value-domain line lists three values, the stop-on-unmatched-value sentence names all three values, a `Strategy: agentic-trunk` heading exists once after the git-flow block, and its route block contains no push instruction of any form (Judge 3 pins the one existing push step, Judge 6 scans the whole block).
- Judge 1: `grep -c -E '^The canonical value domain is exactly .github-flow., .git-flow. and .agentic-trunk.' internal/template/templates/.claude/skills/moai/workflows/sync/delivery.md` (expect `1`).
- Judge 2: `grep -c -E '^##### Strategy: agentic-trunk' internal/template/templates/.claude/skills/moai/workflows/sync/delivery.md` (expect `1`).
- Judge 3: `grep -n -E '^##### Strategy:|Push the integration branch' internal/template/templates/.claude/skills/moai/workflows/sync/delivery.md` (expect exactly four lines in this order: the `github-flow` heading, the `git-flow` heading, the one `Push the integration branch` line, the `agentic-trunk` heading).
- Judge 4 (the stop sentence, so `/moai sync` cannot stop on the new value): `grep -c -E 'neither .github-flow. nor .git-flow.' internal/template/templates/.claude/skills/moai/workflows/sync/delivery.md` and `grep -c -F '{github-flow, git-flow}' internal/template/templates/.claude/skills/moai/workflows/sync/delivery.md` (each expect `0`, exit 1).
- Judge 5 (the positive form): `grep -c -E 'neither .github-flow., .git-flow. nor .agentic-trunk.' internal/template/templates/.claude/skills/moai/workflows/sync/delivery.md` and `grep -c -F '{github-flow, git-flow, agentic-trunk}' internal/template/templates/.claude/skills/moai/workflows/sync/delivery.md` (each expect `1`).
- Judge 6 (no push instruction anywhere in the agentic-trunk block, not only the one existing phrase): `awk 'BEGIN{q=sprintf("%c",96); fence="^[ \t]*" q q q} /^##### Strategy: agentic-trunk/{f=1;next} f&&$0~fence{g=!g} f&&!g&&(/^(#|##|###|####) /||/^##### Strategy:/){f=0} f&&NF{n++} f&&/[Pp]ush|gh pr/{p++} END{print "block_lines=" n+0 " push_lines=" p+0}' internal/template/templates/.claude/skills/moai/workflows/sync/delivery.md` (one invocation). The block runs from its heading to the next sibling: a heading of level 1 to 4, or the next `##### Strategy:` heading, read only outside a fenced snippet (`g` tracks fence state; the fence marker is built with `sprintf("%c",96)` so the program carries no backtick). In `delivery.md` the strategy blocks are level-5 headings (`##### Strategy: github-flow`, `##### Strategy: git-flow`) and the block after them is the level-4 `#### Step 3.3`, so a sub-heading of the block (level 5 not starting `Strategy:`, or level 6) and a `#` comment line inside a fenced snippet stay inside the block and are scanned. Pass stdout: `block_lines=N push_lines=0` with `N` at least `8` (the route's steps 1 to 5, 7 and 8, the non-`WT-*` stop and the landing-surface sentence), exit 0. The pattern rejects `push` in any case form and any `gh pr` call, so a literal `git push origin main`, a `git -C <path> push` and a `Push the integration branch` line all count; the block's own text states "no push" in other words (design section 5).
- RED-now (tree f4175587, observed): Judge 1 stdout `0`, exit 1 (line 22 still names the two-value domain); Judge 2 `0`, exit 1; Judge 3 prints three lines, `238:##### Strategy: github-flow`, `268:##### Strategy: git-flow`, `278:6. Push the integration branch: ...`, exit 0 (no `agentic-trunk` heading); Judge 4 first command stdout `1`, exit 0, second command stdout `1`, exit 0 (line 29 still stops on the two-value form); Judge 5 both stdout `0`, exit 1. Judge 6 (tree c6e590507, ledger E23): `block_lines=0 push_lines=0`, exit 0 (red because of the line floor: the block does not exist). Why red: the edit is not made. Mutant check: an edit that fixes line 22 and the heading but keeps line 29 fails Judge 4; deleting line 29 fails Judge 5; an `agentic-trunk` block that carries `git push origin main` prints `push_lines=1` and fails Judge 6; an empty or missing block prints `block_lines` below the floor and fails Judge 6.
- Judge 6 failing controls (iteration 3 repair, version 0.3.1; tree 45636352d8716b57b0284c24318ad0399d606309, ledger E26): two candidate blocks that carry a push the previous program (`f&&/^#+ /{f=0}`) did not count, each of which MUST print `push_lines` of at least `1` under the program above. (a) a clean block plus a level-6 sub-heading `###### Delivery commands` and the line `git push origin main`: previous program `block_lines=11 push_lines=0`, this program `block_lines=13 push_lines=1`. (b) a clean block plus a fenced snippet whose first line is `# deliver` and then `git push origin main`: previous program `block_lines=12 push_lines=0`, this program `block_lines=15 push_lines=1`. The same program on the template `delivery.md` (no block yet) still prints `block_lines=0 push_lines=0` (red through the floor of 8), and pointed at the git-flow block still prints `block_lines=34 push_lines=8`. Residual: a fence opened with tildes is not modelled (a `~~~` fence with a `# comment` line would end the block early); the block is authored with backtick fences only, and the M5 diff review reads the block. The reworded-push residual recorded below is unchanged.
- Judge 6 controls, observed in ledger E23 (program of version 0.3.0, superseded by the one above; its numbers are history): the same program pointed at the git-flow block prints `block_lines=34 push_lines=8` (the counter fires on a block that has pushes, so a zero is a measurement); on scratch copies of `delivery.md` carrying a candidate block, a clean block prints `block_lines=9 push_lines=0`, the same block with a `git push origin main` step prints `block_lines=10 push_lines=1`. A reworded push such as "Publish the branch to the remote" prints `push_lines=0` and is NOT caught by the program; that residual is closed by the M5 diff review of the block and by Judge 3, and is recorded in the ledger entry.
- Green path: M5.

### AC-AGT-021 (maps REQ-AGT-021)
Given the edited template sources and their deployed counterparts,
When compared,
Then seven pairs are byte-identical and the eighth, `factory-dispatch.md`, differs by the live-only tail sentence alone (the exception REQ-AGT-021 records), and no sentence of the three release-branch files couples `agentic-trunk` to the per-version release-branch push.
- Judge A (regression guard, each `cmp -s <template> <deployed>`, expect stdout empty, exit 0): `factory-dispatch-mechanics.md`, `factory-dispatch-core.md`, `worktree-integration.md`, `delivery-policy.md` and `spec-workflow.md` under `internal/template/templates/.claude/rules/moai/workflow/` against `.claude/rules/moai/workflow/`; `internal/template/templates/.claude/skills/moai/workflows/sync/delivery.md` against `.claude/skills/moai/workflows/sync/delivery.md`; `internal/template/templates/.claude/skills/moai-ref-git-workflow/SKILL.md` against `.claude/skills/moai-ref-git-workflow/SKILL.md`.
- Judge B (the exception pair): (i) `git diff --no-index --numstat internal/template/templates/.claude/rules/moai/workflow/factory-dispatch.md .claude/rules/moai/workflow/factory-dispatch.md` prints `1`, tab, `1`, tab, `{internal/template/templates/.claude => .claude}/rules/moai/workflow/factory-dispatch.md`, exit 1 (exactly one differing line); (ii) `grep -c -F 'disposes remote-landing-confirmed trees' <template> <deployed>` prints `:0` for the template and `:1` for the deployed copy; (iii) the `agentic-trunk` clause that the M5 edit puts into line 222 of both copies, `grep -c -F 'Under agentic-trunk the lane takes no' <template> <deployed>` (prints `:1` for both), the git-flow-only sentence kept, `grep -c -F 'git strategy names git-flow;' <template> <deployed>` (prints `:1` for both), and the retired coupling absent, `grep -c -F 'names git-flow or agentic-trunk' <template> <deployed>` (prints `:0` for both, exit 1); (iv) the byte difference between the two copies stays what it is today, `wc -c <template> <deployed>` prints `25924` and `26246` before M5, a difference of `322` bytes (the tail sentence and its leading space), and after M5 the two sizes differ by the same `322`.
- Judge C (the coupling guard, regression guard): `grep -c -E 'git-flow or agentic-trunk|git-flow and agentic-trunk|agentic-trunk or git-flow|agentic-trunk and git-flow' <factory-dispatch.md, factory-dispatch-core.md and factory-dispatch-mechanics.md, each as template and as deployed copy>` prints `:0` for all six labeled lines (exit 1). These three files describe the git-flow release-branch push (`release/vX.Y.Z`), which spec.md section 4 says `agentic-trunk` does NOT follow; the four files that describe only the shared `WT-*` integration route (`worktree-integration.md`, `spec-workflow.md`, `delivery-policy.md`, `delivery.md`) may couple the two names and are outside this guard.
- Why the exception exists: the deployed `factory-dispatch.md` line 222 carries a tail sentence ("Once the card's work is confirmed landed on the remote integration branch, disposing the card's worktree is the sweep's step: `moai worktree sweep` ...") that the template lacks on purpose. Commit `7ff326d99` (card t1815, "drop live-only worktree-sweep sentence leaked into factory-dispatch mirror") removed it from the template; after a template sentence is carried into a mirror it would ship to user projects. M5 therefore keeps the template without the tail, keeps the deployed copy with it, and applies the same clause to both.
- What the line-222 edit is (iteration 2 restatement): the line is the git-flow variant, "a lane merges its own branch into `release/vX.Y.Z` itself ... pushes `release/vX.Y.Z`", and its sentence "it applies only where the project's git strategy names git-flow" stays exactly as it is. M5 inserts one sentence after "leaves the batch pull request with the leader." and before "The full window procedure:": "Under agentic-trunk the lane takes no `release/vX.Y.Z` branch and runs no push: `moai factory complete` lands the card through the integration-window local merge into `main`, and the only push is the leader's batch pull request from `release/main-batch-YYYYMMDD`." The earlier sentinel `names git-flow or agentic-trunk` would have made `agentic-trunk` a variant that pushes a per-version release branch, contradicting spec.md section 4 and REQ-AGT-010; it is retired and its absence is Judge B (iii).
- RED-now (tree c6e590507, observed this repair): Judge A seven `cmp -s` pairs exit 0 (`factory-dispatch-mechanics`, `factory-dispatch-core`, `worktree-integration`, `delivery-policy`, `spec-workflow`, `delivery.md`, `moai-ref-git-workflow/SKILL.md`); the eighth pair `cmp -s` on `factory-dispatch.md` exits 1; Judge B (i) prints `1	1	{internal/template/templates/.claude => .claude}/rules/moai/workflow/factory-dispatch.md`, exit 1; (ii) prints template `:0`, deployed `:1`, exit 0; (iii) clause `:0` for both, exit 1 (the flip, red because the edit is not made), git-flow-only sentence `:1` for both, exit 0 (guard, must stay), retired coupling `:0` for both, exit 1 (guard, green now); (iv) sizes `25924` and `26246`, difference `322`; Judge C six labeled `:0`, exit 1 (ledger E24). The red this criterion exists to catch: editing the template alone leaves the clause at `:0` for the deployed copy.
- Judge B (iii) and C controls, observed in ledger E24 on scratch copies of both files: the intended edit applied to both copies prints the clause `:1`/`:1`, the git-flow-only sentence `:1`/`:1`, the retired coupling `:0`/`:0`, `git diff --no-index --numstat` `1 1`, and a size difference of `322`; the RETIRED edit (`names git-flow` changed to `names git-flow or agentic-trunk`) prints the clause `:0`/`:0`, the git-flow-only sentence `:0`/`:0` and the retired coupling `:1`/`:1`, so the criterion goes red on exactly the contradiction D2 named.
- Residual risk: a mutant that edits another part of line 222 in one copy only keeps (i) at `1 1`; (iv) catches it unless the edit is length-neutral, and M5's diff review reads the line.
- Green path: M5 after `make build`.

### AC-AGT-022 (maps REQ-AGT-022)
Given the docs-site,
When searched,
Then all four locale `moai-sync.md` pages name `agentic-trunk`.
- Judge: `grep -l agentic-trunk docs-site/content/ko/workflow-commands/moai-sync.md docs-site/content/en/workflow-commands/moai-sync.md docs-site/content/ja/workflow-commands/moai-sync.md docs-site/content/zh/workflow-commands/moai-sync.md` (expect four paths, exit 0).
- RED-now (tree f4175587, observed): stdout empty, exit 1. Why red: no docs-site page names the value.
- Green path: M8 (sync phase).

### AC-AGT-023 (maps REQ-AGT-023)
Given SPEC-PRIMARY-LANDING-ONLY-001 and this SPEC both edit `AGENTS.local.md` and `gitflow-lane-protocol.md`,
When the second one lands,
Then both files carry this SPEC's four-step statement and the landing-surface rule, whichever SPEC landed first.
- The criterion has no "not applicable" case. Two sentinel sets are judged in both files:
  - S-A (this SPEC): Judge 1 `grep -c -E '^# agentic-trunk Lane Protocol' .claude/rules/local/gitflow-lane-protocol.md` (expect `1`); Judge 2 the four-step sentinel of AC-AGT-018 Judge 5 (rescope 0.4.0 string), at least `1` per file.
  - S-B (the landing-surface rule, three judges, each counting a FIXED STRING and compared with a baseline recorded at M7 start, before any M7 edit, in the progress record). The old alternation `performs landing and resync only|Integration-Worktree Rule` is retired: it passed on the bare label alone, and its "at least N0" companion passed vacuously at `N0 = 0`.
    - Judge 3, the t1846 sentence: `grep -c -F 'performs landing and resync only' .claude/rules/local/gitflow-lane-protocol.md AGENTS.local.md`. Record `N1` per file at M7 start; after M7 the count is at least `N1` per file.
    - Judge 4, the t1830 label and its body: `grep -c -F 'Integration-Worktree Rule' .claude/rules/local/gitflow-lane-protocol.md AGENTS.local.md`. Record `N0` per file at M7 start; after M7 the count is at least `N0` per file. When `N0` is at least `1` for a file, the executor also copies ONE verbatim body fragment of at least 40 characters from that rule's statement into the progress record (`F2`), and after M7 `grep -c -F '<F2>' <file>` is at least `1`: this is the body sentinel, fixed from t1830's landed wording rather than guessed, and it fails when the second lander keeps the heading and drops the rule's text.
    - Judge 5, the non-vacuity precondition: at M7 start `N1 + N0` is at least `1` for each file. A file at `0 + 0` carries no landing-surface rule at all, and M7 stops with a blocker report instead of passing a comparison of zero against zero. After M7 each file's `N1' + N0'` is at least that file's start sum and each term is at least its own baseline.
- RED-now (tree c6e590507, observed this repair, ledger E25; Judge 2 re-observed for the 0.4.0 sentinel on tree a6c69e154b, ledger E32 — same shape, both files `:0`, exit 1): Judge 1 stdout `0`, exit 1; Judge 2 `...gitflow-lane-protocol.md:0` and `AGENTS.local.md:0`, exit 1; Judge 3 `...gitflow-lane-protocol.md:1` and `AGENTS.local.md:1`, exit 0 (green now: the t1846 sentence is landed, so `N1` is `1` for both files and this half is a guard); Judge 4 `...gitflow-lane-protocol.md:0` and `AGENTS.local.md:0`, exit 1 (`N0` is `0` for both: the label is absent, so SPEC-PRIMARY-LANDING-ONLY-001 has not landed on this tree; no `F2` exists to record); Judge 5 start sums `1 + 0 = 1` for both files (the precondition holds today because Judge 3 carries the rule, which is what makes the pair non-vacuous at `N0 = 0`). Why red: S-A is absent.
- Mutant checks, observed on scratch copies of `AGENTS.local.md` (ledger E25): dropping the t1846 sentence prints Judge 3 `0` against `N1 = 1` and fails; the state after t1830 landed (t1846 sentence replaced, a labeled rule with a body paragraph added) prints Judge 4 `1` and the `F2` fragment `1`; the SAME state with the heading kept and the body paragraph deleted prints Judge 4 `1` (the label is still there) but the `F2` fragment `0`, so it fails, while the retired alternation prints `1` for that mutant and would have passed; a file with only the title passes Judge 1 and fails Judge 2.
- Green path: M7, judged at the later of the two landings.

## G. Hygiene (M9)

### AC-AGT-024 (maps REQ-AGT-024)
Given this SPEC's branch,
When the diff against the M0 baseline commit is read,
Then no test skip is added and no timeout is raised.
- Judge 1: `git diff <M0-sha>..HEAD -G't\.Skip' --stat -- '*_test.go'` (expect empty, exit 0).
- Judge 2: `git diff <M0-sha>..HEAD -G'timeout' --stat -- Makefile AGENTS.local.md .github` (expect empty, exit 0).
- RED-now: not applicable before M0 exists; the judges compare against the M0 commit, which is why M0 lands in its own commit (verification-claim-integrity section 2.3). Regression guard.
- Green path: M9.

## H. Develop-branch seam consumers (M1, M2)

### AC-AGT-025 (maps REQ-AGT-025)
Given a manual `agentic-trunk` profile that carries no `develop_branch` key (the state REQ-AGT-018 produces),
When the loader projection, the closure push guard, the contract push check and the card-base resolution run,
Then each resolves `main` as the integration branch, with the outcome a git-flow profile whose `develop_branch` is `main` has today.
- Judge 1 (M1): `go test ./internal/config -run '^TestAgenticTrunkDevelopBranchProjection$' -count=1 -v`. Pass: `--- PASS: TestAgenticTrunkDevelopBranchProjection (` with subtests for the key absent, the key present as `develop` (ignored), and personal mode (empty seam); exit 0.
- Judge 2 (M2): `go test ./internal/hook -run '^TestAgenticTrunkIntegrationBranchFor$' -count=1 -v`. Pass: `--- PASS: TestAgenticTrunkIntegrationBranchFor (`; `integrationBranchFor` returns `main` for the fixture tree and for a config-orphaned worktree that falls back to the primary checkout's yaml.
- Judge 3 (M2): `go test ./internal/escalation -run '^TestAgenticTrunkCardBase$' -count=1 -v`. Pass: `--- PASS: TestAgenticTrunkCardBase (`; the fixture repository points `refs/remotes/origin/HEAD` at a branch other than `main`, so the test separates the projection (merge-base with `main`) from the fallback (origin/HEAD).
- Judge 4 (M2): `go test ./internal/cli -run '^TestAgenticTrunkPushCheckIntegration$' -count=1 -v`. Pass: `--- PASS: TestAgenticTrunkPushCheckIntegration (`; under contract autonomy the check does not print `inactive (no git-flow integration branch configured)` and evaluates against `main`.
- RED-now: ledger E06 (Judge 1), E07 (Judge 2), E08 (Judge 3), E14 (Judge 4), each exit 0 with `[no tests to run]`. Why red: none of the four tests exists. Expected once written (not observable before): Judges 2 to 4 fail on a build that does not project the seam, because an unknown value resolves no develop branch and each consumer falls to its inert or fallback path; that is the consumer outcome the key removal of REQ-AGT-018 depends on.
- Not tested, by decision: `internal/config/loader_slot_lease.go` only cites `LoadGitFlowDevelopBranch` in a comment as a modelling reference and calls nothing (design.md row), so it has no behavior to pin.
- Green path: Judge 1 at M1, Judges 2 to 4 at M2.

## Edge cases

- A project that carries `agentic-trunk` in personal or team mode: the target is `main`, the develop seam is empty, `acquire` writes the same unwired-target warning a github-flow project writes (AC-003, AC-007).
- An operator who types `Agentic-Trunk`: classified invalid, doctor warns with the five-value list (AC-001, AC-008).
- Rollback after M6: reverting the flip commit returns a state every shipped binary reads (design.md section 4).

## Definition of Done

All 25 criteria green with the pass stdout named; AC-AGT-016 judged only after the leader's cutover step; the progress record carries, for each cited run, the command, the verbatim stdout, the tree HEAD and the judging binary's commit; Gaps and Residual-risk sections present in the completion report. Quality gates: TRUST 5 on the changed Go packages through the anchored selectors above and `golangci-lint run` scoped to the changed packages; coverage of the new branches by the new tests (every new branch has a test named in sections A to D and H).

## Evidence ledger

Tree `f4175587119b5cfeb8d8348fc74cada8764e89ed` unless an entry names another (E19 and E22 to E25 were observed on `c6e590507b39e27acc3305a3a1aa3dc488aef0c0`); each entry is one command run in a repair pass, its verbatim stdout, and its exit code. Elapsed times vary per run and are not part of any judge. Go runs used `go -C <tree> test`, equal to running from the tree root.

```
E01  exit 0
go test ./internal/config -run '^TestAgenticTrunkValueSet$' -count=1 -v
testing: warning: no tests to run
PASS
ok  	github.com/modu-ai/moai-adk/internal/config	0.369s [no tests to run]

E02  exit 0
go test ./internal/config -run '^TestAgenticTrunkTarget$' -count=1 -v
testing: warning: no tests to run
PASS
ok  	github.com/modu-ai/moai-adk/internal/config	0.165s [no tests to run]

E03  exit 0
go test ./internal/config -run '^TestAgenticTrunkLocalIntegrationPredicate$' -count=1 -v
testing: warning: no tests to run
PASS
ok  	github.com/modu-ai/moai-adk/internal/config	0.187s [no tests to run]

E04  exit 0
go test ./internal/config -run '^TestAgenticTrunkGuidance$' -count=1 -v
testing: warning: no tests to run
PASS
ok  	github.com/modu-ai/moai-adk/internal/config	0.167s [no tests to run]

E05  exit 0
go test ./internal/config -run '^TestAgenticTrunkDistributedDefaultsUnchanged$' -count=1 -v
testing: warning: no tests to run
PASS
ok  	github.com/modu-ai/moai-adk/internal/config	0.180s [no tests to run]

E06  exit 0
go test ./internal/config -run '^TestAgenticTrunkDevelopBranchProjection$' -count=1 -v
testing: warning: no tests to run
PASS
ok  	github.com/modu-ai/moai-adk/internal/config	0.238s [no tests to run]

E07  exit 0
go test ./internal/hook -run '^TestAgenticTrunkIntegrationBranchFor$' -count=1 -v
testing: warning: no tests to run
PASS
ok  	github.com/modu-ai/moai-adk/internal/hook	0.761s [no tests to run]

E08  exit 0
go test ./internal/escalation -run '^TestAgenticTrunkCardBase$' -count=1 -v
testing: warning: no tests to run
PASS
ok  	github.com/modu-ai/moai-adk/internal/escalation	0.640s [no tests to run]

E09  exit 0
go test ./internal/cli -run '^(TestAgenticTrunkFactoryCompleteTakesWindow|TestAgenticTrunkMergeReadyTakesWindow)$' -count=1 -v
testing: warning: no tests to run
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	1.422s [no tests to run]

E10  exit 0
go test ./internal/cli -run '^TestAgenticTrunkAutoMerge$' -count=1 -v
testing: warning: no tests to run
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	1.287s [no tests to run]

E11  exit 0
go test ./internal/cli -run '^TestAgenticTrunkIntegrationAcquire$' -count=1 -v
testing: warning: no tests to run
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	1.246s [no tests to run]

E12  exit 0
go test ./internal/cli -run '^TestAgenticTrunkIgnoredKeys$' -count=1 -v
testing: warning: no tests to run
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	1.124s [no tests to run]

E13  exit 0
go test ./internal/cli -run '^TestAgenticTrunkDoctor$' -count=1 -v
testing: warning: no tests to run
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	1.297s [no tests to run]

E14  exit 0
go test ./internal/cli -run '^TestAgenticTrunkPushCheckIntegration$' -count=1 -v
testing: warning: no tests to run
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	1.094s [no tests to run]

E15  exit 0
go test ./internal/statusline -run '^(TestAgenticTrunkStatuslineLabel|TestAgenticTrunkStatuslineAbsent|TestAgenticTrunkStatuslineNoExtraGit)$' -count=1 -v
testing: warning: no tests to run
PASS
ok  	github.com/modu-ai/moai-adk/internal/statusline	0.376s [no tests to run]

E16  exit 0
go test ./internal/cli/update -run '^TestAgenticTrunkValueSurvivesUpdate$' -count=1 -v
testing: warning: no tests to run
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli/update	0.438s [no tests to run]

E17  exit 0
go test ./internal/template -run '^TestAgenticTrunkSweepGuardNotArmed$' -count=1 -v
testing: warning: no tests to run
PASS
ok  	github.com/modu-ai/moai-adk/internal/template	0.457s [no tests to run]

E18  exit 0   (installed build v3.2.0-rc.36-t1822c, commit ee981a932; working directory = the fixture of AC-AGT-016 P3)
moai doctor
... (the Git Strategy Workflow row, verbatim message text; the box-drawing borders and padding elided)
warn    Git Strategy Workflow  git_strategy workflow = "agentic-trunk" is not one of the allowed flows (github-flow, git-flow, gitlab-flow, release-flow) — repair by editing .moai/config/sections/git-strategy.yaml

E19  exit 0   (re-observed 2026-10-18 on tree c6e590507; installed build as above; working directory = the primary checkout /Users/goos/MoAI/moai-adk-go; the count is volatile, it was 25 at f4175587 and no judge depends on N)
moai doctor --check "MCP Server Version"
ok      MCP Server Version  28 running MCP server(s) match the installed binary (ee981a932)

E20  exit 0   (working directory = the primary checkout)
moai integration status
release-integration window: free
  policy:   open

E21  exit 1   (working directory = the tree root)
sh scripts/verify-local-install.sh
local-install-check: built binary is missing or not executable: bin/moai

E22  exit 0 each   (P4b, observed 2026-10-18T11:1x on tree c6e590507; installed build as above; the process count churns, 32 at the first read and 31 at the last)
pgrep -o -f '^moai mcp-server$'
93989
ps -o pid=,lstart=,command= -p 93989
93989 Sat Oct 17 14:52:06 2026     moai mcp-server
stat -f '%Sm' -t '%Y-%m-%dT%H:%M:%S' /Users/goos/go/bin/moai
2026-10-16T06:56:39
-- population (observation, not a judge): `pgrep -f '^moai mcp-server$'` listed 32 processes; `ps eww -p <pid>` showed CLAUDE_PROJECT_DIR=/Users/goos/MoAI/moai-adk-go for 28 of them (all 28 have a record under the primary .moai/state/mcp-server) and CLAUDE_PROJECT_DIR=/Users/goos/MoAI/mo.ai.kr for 4 (no record under this repository). The oldest process, pid 93989, belongs to the second group. No live pid has a record under any worktree's own .moai/state/mcp-server (a find over .moai/worktrees printed nothing).
-- known-failing input: `stat -f '%Sm' -t '%Y-%m-%dT%H:%M:%S' <scratchpad>/p4-reference-file` (a file touched at observation time, standing in for a reinstall) printed 2026-10-18T11:18:34, which is later than the oldest server's start (2026-10-17T14:52:06): ordered that way the pass condition "the oldest server started after the install" does not hold, so the probe reads red.

E23  exit 0 each   (AC-AGT-020 Judge 6, observed on tree c6e590507, file internal/template/templates/.claude/skills/moai/workflows/sync/delivery.md)
awk '/^##### Strategy: agentic-trunk/{f=1;next} f&&/^#+ /{f=0} f&&NF{n++} f&&/[Pp]ush|gh pr/{p++} END{print "block_lines=" n+0 " push_lines=" p+0}' internal/template/templates/.claude/skills/moai/workflows/sync/delivery.md
block_lines=0 push_lines=0
-- control, same program with the heading pattern changed to the git-flow heading:
block_lines=34 push_lines=8
-- scratch controls (copies of delivery.md under the session scratchpad with a candidate block inserted before "#### Step 3.3"): clean block (9 non-blank lines, steps 1 to 5, 7, 8, no push) -> block_lines=9 push_lines=0; the same block plus "Deliver: `git push origin main`" -> block_lines=10 push_lines=1; the same block plus "Publish the branch to the remote" -> block_lines=10 push_lines=0 (NOT caught; residual, closed by Judge 3 and the M5 diff review).

E24  (AC-AGT-021 Judges B (iii), (iv) and C, observed on tree c6e590507)
grep -c -F 'Under agentic-trunk the lane takes no' <template factory-dispatch.md> <deployed factory-dispatch.md>      exit 1
<template>:0   <deployed>:0
grep -c -F 'names git-flow or agentic-trunk' <both>      exit 1
<template>:0   <deployed>:0
grep -c -F 'git strategy names git-flow;' <both>      exit 0
<template>:1   <deployed>:1
wc -c <template> <deployed>      exit 0
25924 <template>   26246 <deployed>      (difference 322)
grep -c -E 'git-flow or agentic-trunk|git-flow and agentic-trunk|agentic-trunk or git-flow|agentic-trunk and git-flow' <factory-dispatch.md, factory-dispatch-core.md, factory-dispatch-mechanics.md as template and as deployed>      exit 1
six labeled lines, each :0
-- scratch controls (copies of both factory-dispatch.md files under the session scratchpad): intended edit (the one added sentence of AC-AGT-021, inserted after "leaves the batch pull request with the leader.") -> clause :1/:1, git-flow-only sentence :1/:1, retired coupling :0/:0, git diff --no-index --numstat 1 1 (exit 1), sizes 26188 and 26510 (difference 322); retired edit ("names git-flow;" -> "names git-flow or agentic-trunk;") -> clause :0/:0, git-flow-only sentence :0/:0, retired coupling :1/:1.

E25  (AC-AGT-023 Judges 1 to 5, observed on tree c6e590507)
grep -c -E '^# agentic-trunk Lane Protocol' .claude/rules/local/gitflow-lane-protocol.md      exit 1
0
grep -c -F 'agentic-trunk: lane worktree, serial local landing, batch pull request, tagged release' .claude/rules/local/gitflow-lane-protocol.md AGENTS.local.md      exit 1
.claude/rules/local/gitflow-lane-protocol.md:0   AGENTS.local.md:0
grep -c -F 'performs landing and resync only' .claude/rules/local/gitflow-lane-protocol.md AGENTS.local.md      exit 0
.claude/rules/local/gitflow-lane-protocol.md:1   AGENTS.local.md:1      (N1 = 1 and 1)
grep -c -F 'Integration-Worktree Rule' .claude/rules/local/gitflow-lane-protocol.md AGENTS.local.md      exit 1
.claude/rules/local/gitflow-lane-protocol.md:0   AGENTS.local.md:0      (N0 = 0 and 0; Judge 4 alone is vacuous here, Judge 5 start sum 1 + 0 = 1 per file is what keeps the pair non-vacuous)
-- scratch controls (copies of AGENTS.local.md under the session scratchpad): t1846 sentence replaced -> Judge 3 :0; state after t1830 (sentence replaced by "lands cards", a heading "Integration-Worktree Rule" plus a body paragraph added) -> retired alternation :1, Judge 4 label :1, F2 fragment "is the only tree that lands cards onto main and it stays clean" :1; the same state with the body paragraph deleted and the heading kept -> retired alternation :1 (passes, wrongly), Judge 4 label :1, F2 fragment :0 (fails, rightly).
E26  exit 0 each   (AC-AGT-020 Judge 6, iteration 3 repair D1, observed 2026-10-18 on tree 45636352d8716b57b0284c24318ad0399d606309; judging build = the installed moai is not involved, the program is system awk; the scratch files are copies of internal/template/templates/.claude/skills/moai/workflows/sync/delivery.md with a candidate block inserted before "#### Step 3.3: PR Ready Transition", built by a script under the session scratchpad)
awk 'BEGIN{q=sprintf("%c",96); fence="^[ \t]*" q q q} /^##### Strategy: agentic-trunk/{f=1;next} f&&$0~fence{g=!g} f&&!g&&(/^(#|##|###|####) /||/^##### Strategy:/){f=0} f&&NF{n++} f&&/[Pp]ush|gh pr/{p++} END{print "block_lines=" n+0 " push_lines=" p+0}' <file>
-- <file> = the template delivery.md itself (no agentic-trunk block yet): block_lines=0 push_lines=0   (previous program: the same; red through the floor of 8)
-- <file> = the template delivery.md with the heading pattern changed to the git-flow heading: block_lines=34 push_lines=8   (previous program: the same)
-- <file> = scratch clean block (11 non-blank lines, no push word): block_lines=11 push_lines=0   (previous program: the same)
-- <file> = scratch clean block plus "Deliver: `git push origin main`": block_lines=12 push_lines=1   (previous program: the same)
-- FAILING CONTROL (a), clean block plus "###### Delivery commands", a blank line and "git push origin main": block_lines=13 push_lines=1   (previous program f&&/^#+ /{f=0}: block_lines=11 push_lines=0, the leak)
-- FAILING CONTROL (b), clean block plus a fenced snippet of "# deliver" and "git push origin main": block_lines=15 push_lines=1   (previous program: block_lines=12 push_lines=0, the leak)
-- why a bare level cut is not enough: the audit's suggested fix (end the block at heading levels 1 to 5) repairs control (a) and leaves control (b), because "# deliver" is a level-1 heading by pattern; the fence state `g` is what keeps control (b) inside the block. The sibling blocks of the agentic-trunk heading are the level-5 "##### Strategy:" headings (delivery.md lines 238 and 268) and the next section is the level-4 "#### Step 3.3" (line 313), so levels 1 to 4 plus "##### Strategy:" end the block where the next sibling begins.
```

Rescope 0.4.0 entries (E27 to E32), observed 2026-10-21 on tree `a6c69e154b62eb33763212a3034e505bd1046fe1` unless the entry names otherwise:

```
E27  (AC-AGT-017 Judges 1 to 4, and AC-AGT-016 P1, this tree)
grep -c 'workflow: agentic-trunk' .moai/config/sections/git-strategy.yaml      exit 1
0
grep -c 'workflow: github-flow' .moai/config/sections/git-strategy.yaml      exit 0
2
grep -c -E 'lead_push_threshold: 20' .moai/config/sections/git-strategy.yaml      exit 0
1
grep -c -E 'develop_branch:|release_branch_prefix:|rc_version_format:|stays git-flow' .moai/config/sections/git-strategy.yaml      exit 0
4
git log --format=%h -n 1 -S'workflow: agentic-trunk' -- .moai/config/sections/git-strategy.yaml      exit 0
(no output — no flip commit exists)
-- note: this tree's committed manual profile still reads `workflow: git-flow` (line 9); the leader's
-- bridge flip (github-flow) is not yet a commit this tree can see. Judge 4's 0.4.0 alternation adds
-- `stays github-flow`; the four-count above was observed with the pre-rescope alternation, and the
-- added alternative cannot lower the count (it is a pure widening), so the observation holds for
-- the rescoped judge.

E28  (AC-AGT-016 P2, this tree)
sh scripts/verify-local-install.sh      exit 1
local-install-check: FAIL — installed binary differs from bin/moai
-- red for the right reason: bin/moai exists (the M0-M5 run built it) and the installed bridge
-- build (2026-10-19T21:47:22, v3.2.0-rc.37, commit 8249bff8c3) differs from it; the flip parent
-- whose build P2 demands does not exist yet.

E29  (AC-AGT-016 P3, installed build v3.2.0-rc.37, commit 8249bff8c3; working directory = the fixture
project /tmp/t1831-fixture whose only file is .moai/config/sections/git-strategy.yaml with
git_strategy / mode: manual / manual / workflow: agentic-trunk / worktree_base_branch: main)
moai doctor      exit 0
warn    Git Strategy Workflow  git_strategy workflow = "agentic-trunk" is not one of the allowed flows (github-flow, git-flow, gitlab-flow, release-flow) — repair by editing .moai/config/sections/git-strategy.yaml
-- (the row verbatim; box borders and unrelated rows elided). Red because the installed build
-- predates this branch and lacks the enum — the bridge's reason for existing.

E30  (AC-AGT-016 P4a and P4b, worktree-side observation; the primary-side rows are leader-time)
moai doctor --check "MCP Server Version"      exit 0   (working directory = this worktree)
ok      MCP Server Version  no running moai MCP server recorded
stat -f '%Sm' -t '%Y-%m-%dT%H:%M:%S' /Users/goos/go/bin/moai
2026-10-19T21:47:22
-- the records the P4a row reads live under <project>/.moai/state/mcp-server of the project
-- directory the doctor resolves; from this worktree it resolves none, so the row reports "no
-- running server recorded". The leader re-measures P4a from the primary checkout and P4b over all
-- processes at cutover, exactly as E19 and E22 recorded the method.

E31  (AC-AGT-005 flip judge and AC-AGT-006 flip judge B, this tree)
go test ./internal/cli -run '^(TestAgenticTrunkCompletePrimaryPrEdge|TestAgenticTrunkCompleteNoAutoMerge)$' -count=1 -v      exit 0
testing: warning: no tests to run
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	1.023s [no tests to run]
go test ./internal/cli -run '^TestAgenticTrunkAutoMergeSkipsPrimaryTarget$' -count=1 -v      exit 0
testing: warning: no tests to run
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	0.960s [no tests to run]
-- the observed code state behind the red: internal/cli/factory_card_pr.go:444 reads
-- factoryGHCall(wt, "pr", "merge", strconv.Itoa(pr.Number), "--auto", mergeFlag, "--match-head-commit", checkedTip)

E32  (AC-AGT-018 Judge 5 and AC-AGT-023 Judge 2, the 0.4.0 sentinel, this tree)
grep -c -F 'agentic-trunk: lane worktree, serial landing into main, leader release batch per lead_push_threshold, tagged release' .claude/rules/local/gitflow-lane-protocol.md AGENTS.local.md      exit 1
.claude/rules/local/gitflow-lane-protocol.md:0   AGENTS.local.md:0
```
