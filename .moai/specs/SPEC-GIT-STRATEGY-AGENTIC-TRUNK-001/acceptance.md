# Acceptance — SPEC-GIT-STRATEGY-AGENTIC-TRUNK-001

Document-level pin: every observation below was taken on tree `f4175587119b5cfeb8d8348fc74cada8764e89ed` (branch `WT-agentic-trunk`, plan-audit iteration 1 repair). That commit differs from `6041c38a4` (the tree of research.md) in the six SPEC files only, so every code, template, config and guidance observation holds on both. A criterion-level pin wins where one is given.

## Conventions

- A judge is a single invocation. A Go-test judge is `go test <pkg> -run '^(TestA|TestB)$' -count=1 -v`; it is green only when stdout carries each named `--- PASS:` line, exit is 0 and `[no tests to run]` is absent. A selector that matches nothing prints `ok ... [no tests to run]` with exit 0 (ledger E01), so exit 0 alone is never a verdict.
- Each Go criterion has a flip judge that selects only NEW test names (red now: `[no tests to run]`; green: the named `--- PASS:` lines) and, where an existing test is in scope, a separate guard judge that selects only EXISTING names (green now, must stay green). A mixed selector is not used: a selector naming one existing and one new test prints a passing line today (observed in plan-audit iteration 1 for AC-AGT-001), so its red would be a missing second line, not a failing run.
- Classes. Flip criteria are release-blocking and carry the four RED-now elements (command, verbatim stdout, exit code, tree SHA) either inline or as a cited ledger entry E01 to E21 (end of this file). Regression guards (AC-AGT-002, AC-AGT-010, the guard judges, AC-AGT-014 judge B, AC-AGT-021 judge A, AC-AGT-024) are green today by design and say so. AC-AGT-022 is judged in the sync phase. AC-AGT-016 is executed by the leader (below).
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

### AC-AGT-005 (maps REQ-AGT-006)
Given a project whose manual workflow is `agentic-trunk`, with a fake `gh` that records any call,
When `factory complete` and `factory merge ready` run on a card that cleared the readiness triple,
Then the integration window is taken and recorded, no `gh pr` call is made, and no pull request is opened.
- Flip judge: `go test ./internal/cli -run '^(TestAgenticTrunkFactoryCompleteTakesWindow|TestAgenticTrunkMergeReadyTakesWindow)$' -count=1 -v`. Pass stdout: two `--- PASS:` lines; exit 0.
- Guard judge (green now): `go test ./internal/cli -run '^TestFactoryCompleteWindowGitFlowUnchanged$' -count=1 -v`. Observed: `--- PASS: TestFactoryCompleteWindowGitFlowUnchanged (5.82s)`, `ok  github.com/modu-ai/moai-adk/internal/cli  7.257s`, exit 0.
- RED-now (flip): ledger E09, exit 0, `[no tests to run]`. Why red: the two tests do not exist; `factoryGitHubFlow` is already false for the new value, so the behavior needs a pin, not a code change.
- Green path: M2.

### AC-AGT-006 (maps REQ-AGT-007)
Given session-exit auto-merge enabled and manual `agentic-trunk`,
When a clean session exit has commits ahead of `main`,
Then the branch merges into `main` under the same guards as manual git-flow; for a non-manual agentic-trunk project the skip notice names both workflows.
- Flip judge: `go test ./internal/cli -run '^TestAgenticTrunkAutoMerge$' -count=1 -v`. Pass stdout: `--- PASS: TestAgenticTrunkAutoMerge (`; exit 0.
- Guard judge (green now): `go test ./internal/cli -run '^(TestAutoMergeHappyPath|TestAutoMergeUnconfiguredTarget)$' -count=1 -v`. Observed: `--- PASS: TestAutoMergeHappyPath (1.21s)`, `--- PASS: TestAutoMergeUnconfiguredTarget (0.00s)`, `ok  github.com/modu-ai/moai-adk/internal/cli  2.472s`, exit 0. `TestAutoMergeUnconfiguredTarget` is updated in M2 if it pins the old notice string.
- RED-now (flip): ledger E10, exit 0, `[no tests to run]`. Why red: the test does not exist and the gate `!IsGitFlow()` skips the new value.
- Green path: M2.

### AC-AGT-007 (maps REQ-AGT-008)
Given manual `agentic-trunk` and a worktree holding `main`,
When `moai integration acquire --name lane --card c1` runs from a card tree without `--branch`,
Then the recorded window names the worktree holding `main`, stderr carries no caller-fallback warning, and the git-flow and github-flow acquire behaviors are unchanged.
- Flip judge: `go test ./internal/cli -run '^TestAgenticTrunkIntegrationAcquire$' -count=1 -v`. Pass stdout: `--- PASS: TestAgenticTrunkIntegrationAcquire (`; exit 0.
- Guard judge (green now): `go test ./internal/cli -run '^(TestIntegrationAcquire_GitFlowEmptyDevelopWarnsOnCallerFallback|TestIntegrationAcquire_GitHubFlowEmptyDevelopDoesNotWarn)$' -count=1 -v`. Observed: both `--- PASS:` lines (0.18s, 0.21s), `ok  github.com/modu-ai/moai-adk/internal/cli  1.189s`, exit 0.
- RED-now (flip): ledger E11, exit 0, `[no tests to run]`. Why red: the test does not exist.
- Green path: M2.

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
Then it contains none.
- Judge: `git grep -c -E 'LoadGitFlowIntegrationConfig|WorkflowGitFlow|WorkflowGitHubFlow|WorkflowAgenticTrunk|"agentic-trunk"' -- internal/cli/worktree`
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

### AC-AGT-016 (maps REQ-AGT-017)
Executed by the leader, with the operator's confirmation, in a window where no lane is mid-landing: replacing the installed binary affects every session on the machine. The precondition set below is the single list; REQ-AGT-017, plan.md M6 and design.md section 4 carry the same P1 to P5.

Given the flip commit,
When the commit graph, the installed binary and the running servers are inspected,
Then P1 to P5 all hold.
- P1, ancestry. Judges (five invocations, each exit 0, stdout empty): `git merge-base --is-ancestor <M1-sha> <flip-sha>` and the same for M2, M3, M4 and M5. The flip commit touches one path: `git diff-tree --no-commit-id --name-only -r <flip-sha>` prints `.moai/config/sections/git-strategy.yaml` only. Judge-form control, observed on this tree: `git merge-base --is-ancestor f4175587119b5cfeb8d8348fc74cada8764e89ed 6041c38a4` exits 1 and the reverse pair exits 0, so the form separates a wrong order from a right one.
- P2, the installed binary is the tree build of the flip commit's parent. Judge: `sh scripts/verify-local-install.sh bin/moai /Users/goos/go/bin/moai <flip-parent-short-sha>`; pass: exit 0 and a final line `local-install-check: OK — /Users/goos/go/bin/moai is byte-identical to bin/moai and reports commit <flip-parent-short-sha>`.
- P3, the installed binary understands the value. Judge: `moai doctor` (the installed binary by name, on purpose) with the working directory set to a fixture project whose only file is `.moai/config/sections/git-strategy.yaml` holding `git_strategy:` / `mode: manual` / `manual:` / `workflow: agentic-trunk` / `worktree_base_branch: main` (two-space indents). Pass: the `Git Strategy Workflow` row has status `ok` and its message contains `agentic-trunk`, `serial` and `batch`.
- P4, no stale server. Judge: `moai doctor --check "MCP Server Version"` with the working directory at the project root (the primary checkout). Pass: an `ok` row reading `N running MCP server(s) match the installed binary (<flip-parent-short-sha>)`; the stale state is a `warn` row `running MCP server is stale (...)`. The leader reconnects every server (Claude Code `/mcp` reconnect or session restart) and records the reconnect in the progress record.
- P5, execution condition. Judge: `moai integration status` prints `release-integration window: free`; the operator's confirmation is recorded in the progress record (a state witness, not a flip judge).
- RED-now (tree f4175587, this machine, installed build `v3.2.0-rc.36-t1822c`, commit `ee981a932`): P1 no flip commit exists, `git log --format=%h -n 1 -S'workflow: agentic-trunk' -- .moai/config/sections/git-strategy.yaml` prints nothing, exit 0; P2 `sh scripts/verify-local-install.sh` prints `local-install-check: built binary is missing or not executable: bin/moai`, exit 1 (no tree build is installed); P3 ledger E18, the installed build prints a `warn` row for the fixture (the value is unknown to it) and exits 0; P4 ledger E19, the row reads `(ee981a932)`, which is not the flip parent's short SHA, so it is red for the right reason; P5 ledger E20 (`free`, already satisfied, a witness).
- Why this gate can show its hazard: P3 turns red on a build that lacks M1 to M4, and P4 turns red on a reinstall that was not followed by a reconnect. A flip before M2, M3 or M5, or with the old binary still installed, fails P1, P2 or P3.
- Green path: M6. The commit graph is the only ordering evidence (verification-claim-integrity section 2.3).

### AC-AGT-017 (maps REQ-AGT-018)
Given the live yaml after the flip,
When searched,
Then the manual profile carries `workflow: agentic-trunk`, personal and team keep `github-flow`, `lead_push_threshold: 20` remains, and the three git-flow-only keys and the stale comment are gone.
- Judge 1: `grep -c 'workflow: agentic-trunk' .moai/config/sections/git-strategy.yaml` (expect `1`).
- Judge 2: `grep -c 'workflow: github-flow' .moai/config/sections/git-strategy.yaml` (expect `2`).
- Judge 3: `grep -c -E 'lead_push_threshold: 20' .moai/config/sections/git-strategy.yaml` (expect `1`).
- Judge 4: `grep -c -E 'develop_branch:|release_branch_prefix:|rc_version_format:|stays git-flow' .moai/config/sections/git-strategy.yaml` (expect `0`, exit 1).
- RED-now (tree f4175587, observed): Judge 1 stdout `0`, exit 1; Judge 2 stdout `2`, exit 0; Judge 3 stdout `1`, exit 0; Judge 4 stdout `4`, exit 0 (the three keys plus the line-28 comment `stays git-flow`). Why red: the value has not been flipped. Judges 2 and 3 are green now by design; they guard personal, team and the threshold against a careless edit.
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
- Judge 5 (the four-step statement, shared with AC-AGT-023): `grep -c -F 'agentic-trunk: lane worktree, serial local landing, batch pull request, tagged release' .claude/rules/local/gitflow-lane-protocol.md AGENTS.local.md` (expect at least `1` per file).
- Judge 6: `grep -c 'Card workflow (git-flow transition' .claude/rules/local/repo-local-pr-policy.md` (expect `0`) and `grep -c agentic-trunk .claude/rules/local/repo-local-pr-policy.md .moai/docs/git-workflow-doctrine.md .moai/docs/git-local-workflow-doctrine.md .moai/docs/gitflow-integration-chain.md` (expect at least `1` per file).
- RED-now (tree f4175587, observed): Judge 1 stdout `0`, exit 1; Judge 2 stdout `.claude/rules/local/gitflow-lane-protocol.md:2` and `AGENTS.local.md:0`, exit 0 (the H1 `# git-flow Lane Protocol (moai-adk-go, local-only)` and the blockquote line match); Judge 3 prints the path, exit 0 (green now, must stay); Judge 4 stdout `0`, exit 1; Judge 5 stdout `...gitflow-lane-protocol.md:0` and `AGENTS.local.md:0`, exit 1; Judge 6 first command stdout `1`, exit 0, second command `0` for all four files, exit 1. Why red: the name has not been applied.
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
Then the value-domain line lists three values, the stop-on-unmatched-value sentence names all three values, a `Strategy: agentic-trunk` heading exists once after the git-flow block, and its route block contains no push step.
- Judge 1: `grep -c -E '^The canonical value domain is exactly .github-flow., .git-flow. and .agentic-trunk.' internal/template/templates/.claude/skills/moai/workflows/sync/delivery.md` (expect `1`).
- Judge 2: `grep -c -E '^##### Strategy: agentic-trunk' internal/template/templates/.claude/skills/moai/workflows/sync/delivery.md` (expect `1`).
- Judge 3: `grep -n -E '^##### Strategy:|Push the integration branch' internal/template/templates/.claude/skills/moai/workflows/sync/delivery.md` (expect exactly four lines in this order: the `github-flow` heading, the `git-flow` heading, the one `Push the integration branch` line, the `agentic-trunk` heading).
- Judge 4 (the stop sentence, so `/moai sync` cannot stop on the new value): `grep -c -E 'neither .github-flow. nor .git-flow.' internal/template/templates/.claude/skills/moai/workflows/sync/delivery.md` and `grep -c -F '{github-flow, git-flow}' internal/template/templates/.claude/skills/moai/workflows/sync/delivery.md` (each expect `0`, exit 1).
- Judge 5 (the positive form): `grep -c -E 'neither .github-flow., .git-flow. nor .agentic-trunk.' internal/template/templates/.claude/skills/moai/workflows/sync/delivery.md` and `grep -c -F '{github-flow, git-flow, agentic-trunk}' internal/template/templates/.claude/skills/moai/workflows/sync/delivery.md` (each expect `1`).
- RED-now (tree f4175587, observed): Judge 1 stdout `0`, exit 1 (line 22 still names the two-value domain); Judge 2 `0`, exit 1; Judge 3 prints three lines, `238:##### Strategy: github-flow`, `268:##### Strategy: git-flow`, `278:6. Push the integration branch: ...`, exit 0 (no `agentic-trunk` heading); Judge 4 first command stdout `1`, exit 0, second command stdout `1`, exit 0 (line 29 still stops on the two-value form); Judge 5 both stdout `0`, exit 1. Why red: the edit is not made. Mutant check: an edit that fixes line 22 and the heading but keeps line 29 fails Judge 4; deleting line 29 fails Judge 5.
- Green path: M5.

### AC-AGT-021 (maps REQ-AGT-021)
Given the edited template sources and their deployed counterparts,
When compared,
Then seven pairs are byte-identical and the eighth, `factory-dispatch.md`, differs by the live-only tail sentence alone.
- Judge A (regression guard, each `cmp -s <template> <deployed>`, expect stdout empty, exit 0): `factory-dispatch-mechanics.md`, `factory-dispatch-core.md`, `worktree-integration.md`, `delivery-policy.md` and `spec-workflow.md` under `internal/template/templates/.claude/rules/moai/workflow/` against `.claude/rules/moai/workflow/`; `internal/template/templates/.claude/skills/moai/workflows/sync/delivery.md` against `.claude/skills/moai/workflows/sync/delivery.md`; `internal/template/templates/.claude/skills/moai-ref-git-workflow/SKILL.md` against `.claude/skills/moai-ref-git-workflow/SKILL.md`.
- Judge B (the exception pair): (i) `git diff --no-index --numstat internal/template/templates/.claude/rules/moai/workflow/factory-dispatch.md .claude/rules/moai/workflow/factory-dispatch.md` prints `1`, tab, `1`, tab, `{internal/template/templates/.claude => .claude}/rules/moai/workflow/factory-dispatch.md`, exit 1 (exactly one differing line); (ii) `grep -c -F 'disposes remote-landing-confirmed trees' <template> <deployed>` prints `:0` for the template and `:1` for the deployed copy; (iii) `grep -c -F 'names git-flow or agentic-trunk' <template> <deployed>` prints `:1` for both (the sentinel of the M5 edit to the line, applied to both copies).
- Why the exception exists: the deployed `factory-dispatch.md` line 222 carries a tail sentence ("Once the card's work is confirmed landed on the remote integration branch, disposing the card's worktree is the sweep's step: `moai worktree sweep` ...") that the template lacks on purpose. Commit `7ff326d99` (card t1815, "drop live-only worktree-sweep sentence leaked into factory-dispatch mirror") removed it from the template; after a template sentence is carried into a mirror it would ship to user projects. M5 therefore keeps the template without the tail, keeps the deployed copy with it, and applies the `agentic-trunk` edit to both.
- RED-now (tree f4175587, observed): Judge A seven `cmp -s` pairs exit 0 (`factory-dispatch-mechanics`, `factory-dispatch-core`, `worktree-integration`, `delivery-policy`, `spec-workflow`, `delivery.md`, `moai-ref-git-workflow/SKILL.md`); the eighth pair `cmp -s` on `factory-dispatch.md` exits 1; Judge B (i) prints `1	1	{internal/template/templates/.claude => .claude}/rules/moai/workflow/factory-dispatch.md`, exit 1; (ii) prints template `:0`, deployed `:1`, exit 0; (iii) prints `:0` for both, exit 1. Why red: (iii) is the flip, the sentinel edit is not made; (i) and (ii) are the exception recorded green. The red this criterion exists to catch: editing the template alone leaves (iii) at `:0` for the deployed copy.
- Residual risk: a mutant that edits another part of line 222 in one copy only keeps (i) at `1 1`; M5's diff review reads the line.
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
  - S-A (this SPEC): Judge 1 `grep -c -E '^# agentic-trunk Lane Protocol' .claude/rules/local/gitflow-lane-protocol.md` (expect `1`); Judge 2 the four-step sentinel of AC-AGT-018 Judge 5, at least `1` per file.
  - S-B (the landing-surface rule): Judge 3 `grep -c -E 'performs landing and resync only|Integration-Worktree Rule' .claude/rules/local/gitflow-lane-protocol.md AGENTS.local.md` (expect at least `1` per file). Before SPEC-PRIMARY-LANDING-ONLY-001 lands the first alternative is the landed t1846 sentence; after it lands the second alternative is its label; either form passes, and a file with neither fails.
  - Judge 4, nothing of the other SPEC removed: at M7 start the executor records `N0` per file with `grep -c 'Integration-Worktree Rule' .claude/rules/local/gitflow-lane-protocol.md AGENTS.local.md` in the progress record (before any M7 edit); after M7 the same command prints at least `N0` per file.
- RED-now (tree f4175587, observed): Judge 1 stdout `0`, exit 1; Judge 2 `...gitflow-lane-protocol.md:0` and `AGENTS.local.md:0`, exit 1; Judge 3 `...gitflow-lane-protocol.md:1` and `AGENTS.local.md:1`, exit 0 (green now: the t1846 sentence is landed; this half is a guard); Judge 4 `N0` is `0` for both files (the `Integration-Worktree Rule` label is absent, so SPEC-PRIMARY-LANDING-ONLY-001 has not landed on this tree), exit 1. Why red: S-A is absent.
- Mutant check: a file with only the title passes Judge 1 and fails Judge 2; deleting the landing-surface sentence fails Judge 3; rewriting a landed `Integration-Worktree Rule` away while keeping the t1846 phrase fails Judge 4.
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

Tree `f4175587119b5cfeb8d8348fc74cada8764e89ed`; each entry is one command run in this repair, its verbatim stdout, and its exit code. Elapsed times vary per run and are not part of any judge. Go runs used `go -C <tree> test`, equal to running from the tree root.

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

E19  exit 0   (installed build as above; working directory = the primary checkout /Users/goos/MoAI/moai-adk-go)
moai doctor --check "MCP Server Version"
ok      MCP Server Version  25 running MCP server(s) match the installed binary (ee981a932)

E20  exit 0   (working directory = the primary checkout)
moai integration status
release-integration window: free
  policy:   open

E21  exit 1   (working directory = the tree root)
sh scripts/verify-local-install.sh
local-install-check: built binary is missing or not executable: bin/moai
```
