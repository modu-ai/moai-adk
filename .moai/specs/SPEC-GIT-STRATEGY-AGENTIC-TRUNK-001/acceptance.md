# Acceptance — SPEC-GIT-STRATEGY-AGENTIC-TRUNK-001

Document-level pin: every RED-now observation is taken on tree `6041c38a4` unless a criterion says otherwise. Observed values are in research.md R6; criteria marked "M0" take their RED-now at the baseline step (plan.md M0) with the same selector form.

Conventions. Judges are single invocations. A Go-test judge is `go test <pkg> -run '^(TestA|TestB)$' -count=1 -v`; it is green only when stdout carries each named `--- PASS:` line, exit is 0 and `no tests to run` is absent. A selector that matches nothing is red (research R6). Every criterion is release-blocking for its milestone and carries RED-now (command, stdout, exit, tree) and a green path (milestone and passing output); the criteria whose RED-now says "M0" take that observation at the baseline step, because the same vacuous-green form was observed for `internal/config` only. AC-002, AC-010 and the defaults half of AC-014 are regression guards that are green today by design. AC-022 is judged in the sync phase.

## A. Value set and reader (M1)

### AC-AGT-001 (maps REQ-AGT-001, REQ-AGT-002)
Given the allowed-workflow set and the disposition classifier,
When the tests run,
Then `agentic-trunk` is the fifth and last allowed value, classifies valid and not git-flow, `Agentic-Trunk`, `agentic_trunk`, `agentic-trunk ` and the empty string classify invalid, and `git-flow`, `github-flow`, `gitlab-flow`, `release-flow` classify as before.
- Judge: `go test ./internal/config -run '^(TestWorkflowDisposition|TestAgenticTrunkValueSet)$' -count=1 -v`
- Pass stdout: `--- PASS: TestWorkflowDisposition (` and `--- PASS: TestAgenticTrunkValueSet (`; exit 0.
- RED-now: same command, stdout `testing: warning: no tests to run ... [no tests to run]`, exit 0 for the missing test; `TestWorkflowDisposition` asserts `want 4` and turns red once the set has 5 until updated. Why red: the value does not exist.
- Green path: M1.

### AC-AGT-002 (maps REQ-AGT-001)
Given the existing git-flow and github-flow reader tests,
When the compatibility pack runs after M1 and again after M2,
Then every named test passes with no assertion weakened: `TestWorkflowTargetResolution`, `TestEmptyTargetGuidance`, `TestEmptyTargetGuidanceResolvedTargetIsSilent`, `TestTargetProvenance`.
- Judge: `go test ./internal/config -run '^(TestWorkflowTargetResolution|TestEmptyTargetGuidance|TestEmptyTargetGuidanceResolvedTargetIsSilent|TestTargetProvenance)$' -count=1 -v`
- Pass stdout: four `--- PASS:` lines; exit 0.
- RED-now: green today (a regression guard, not a flip criterion); it is released-blocking only as "must stay green". Why it matters: it is the compatibility witness.

### AC-AGT-003 (maps REQ-AGT-003, REQ-AGT-004)
Given a manual profile with `workflow: agentic-trunk` and `develop_branch: develop`, and a personal and a team profile with the same workflow,
When the strategy reader loads each,
Then the manual profile reports integration target `main` and develop-branch seam `main`; personal and team report integration target `main` and an empty develop-branch seam; `environment` and `release_branch_prefix` values change nothing.
- Judge: `go test ./internal/config -run '^TestAgenticTrunkTarget$' -count=1 -v`
- Pass stdout: `--- PASS: TestAgenticTrunkTarget (` (subtests for manual, personal, team); exit 0.
- RED-now: `[no tests to run]`, exit 0. Why red: the test and the behavior do not exist.
- Green path: M1.

### AC-AGT-004 (maps REQ-AGT-005)
Given the reader's predicates,
When the table runs (git-flow manual, agentic-trunk manual, agentic-trunk team, github-flow manual, absent file),
Then the local-integration predicate is true for the first two only, and the git-flow-only predicate is true for the first only.
- Judge: `go test ./internal/config -run '^TestAgenticTrunkLocalIntegrationPredicate$' -count=1 -v`
- Pass stdout: `--- PASS: TestAgenticTrunkLocalIntegrationPredicate (`; exit 0.
- RED-now: `[no tests to run]`, exit 0.
- Green path: M1.

## B. Consumers (M2)

### AC-AGT-005 (maps REQ-AGT-006)
Given a project whose manual workflow is `agentic-trunk`, with a fake `gh` that records any call,
When `factory complete` and `factory merge ready` run on a card that cleared the readiness triple,
Then the integration window is taken and recorded, no `gh pr` call is made, and no pull request is opened.
- Judge: `go test ./internal/cli -run '^(TestFactoryCompleteWindowGitFlowUnchanged|TestAgenticTrunkFactoryCompleteTakesWindow|TestAgenticTrunkMergeReadyTakesWindow)$' -count=1 -v`
- Pass stdout: three `--- PASS:` lines; exit 0.
- RED-now: M0 (selector form of research R6; the two new names match nothing).
- Green path: M2 (the first name is the git-flow sibling, kept green).

### AC-AGT-006 (maps REQ-AGT-007)
Given session-exit auto-merge enabled and manual `agentic-trunk`,
When a clean session exit has commits ahead of `main`,
Then the branch merges into `main` under the same guards as manual git-flow; for a non-manual agentic-trunk project the skip notice names both workflows.
- Judge: `go test ./internal/cli -run '^(TestAutoMergeHappyPath|TestAutoMergeUnconfiguredTarget|TestAgenticTrunkAutoMerge)$' -count=1 -v`
- Pass stdout: three `--- PASS:` lines; exit 0.
- RED-now: M0.
- Green path: M2. `TestAutoMergeUnconfiguredTarget` is updated if it pins the old notice string.

### AC-AGT-007 (maps REQ-AGT-008)
Given manual `agentic-trunk` and a worktree holding `main`,
When `moai integration acquire --name lane --card c1` runs from a card tree without `--branch`,
Then the recorded window names the worktree holding `main`, stderr carries no caller-fallback warning, and the git-flow and github-flow acquire behaviors are unchanged.
- Judge: `go test ./internal/cli -run '^(TestIntegrationAcquire_GitFlowEmptyDevelopWarnsOnCallerFallback|TestIntegrationAcquire_GitHubFlowEmptyDevelopDoesNotWarn|TestAgenticTrunkIntegrationAcquire)$' -count=1 -v`
- Pass stdout: three `--- PASS:` lines; exit 0.
- RED-now: M0.
- Green path: M2.

### AC-AGT-008 (maps REQ-AGT-009)
Given an invalid workflow value and a non-manual git-flow profile,
When the diagnostics render,
Then the allowed-set listing contains all five values and the git-flow fix strings name `agentic-trunk (target main)` beside `github-flow (target main)`.
- Judge: `go test ./internal/config -run '^TestAgenticTrunkGuidance$' -count=1 -v`
- Pass stdout: `--- PASS: TestAgenticTrunkGuidance (`; exit 0.
- RED-now: `[no tests to run]`, exit 0.
- Green path: M2.

### AC-AGT-009 (maps REQ-AGT-010)
Given a manual `agentic-trunk` profile that still carries `develop_branch`, `release_branch_prefix` and `rc_version_format`, plus `lead_push_threshold: 20` and `merge_method: merge`,
When the reader and the doctor run,
Then no diagnostic mentions the three keys, the target is `main`, and `lead_push_threshold` and `merge_method` read through unchanged.
- Judge: `go test ./internal/cli -run '^TestAgenticTrunkIgnoredKeys$' -count=1 -v`
- Pass stdout: `--- PASS: TestAgenticTrunkIgnoredKeys (`; exit 0.
- RED-now: M0.
- Green path: M2.

### AC-AGT-010 (maps REQ-AGT-011)
Given the landing-check code,
When searched for a workflow-value read,
Then it contains none.
- Judge: `git grep -c -E 'LoadGitFlowIntegrationConfig|WorkflowGitFlow|WorkflowGitHubFlow|WorkflowAgenticTrunk|"agentic-trunk"' -- internal/cli/worktree`
- Pass stdout: empty; exit 1 (no match).
- RED-now (observed, tree `6041c38a4`): stdout empty, exit 1. Why already green: it is a regression guard that fixes the invariant before M2 can break it; it is not a flip criterion, and it is a positive-control-free grep, so the run also records `git grep -c LoadGitFlowIntegrationConfig -- internal/cli/factory_card.go` (stdout a positive count) to show the pattern fires where the identifier exists.
- Green path: unchanged by M2; failing means M2 added a workflow read to the landing checks.

## C. Display (M3)

### AC-AGT-011 (maps REQ-AGT-012)
Given manual `agentic-trunk`,
When the doctor check "Git Strategy Workflow" runs,
Then status is OK and the message contains `agentic-trunk`, `main`, `serial`, and `batch`.
- Judge: `go test ./internal/cli -run '^(TestDoctorGitStrategyWorkflow|TestAgenticTrunkDoctor)$' -count=1 -v`
- Pass stdout: `--- PASS: TestDoctorGitStrategyWorkflow (` and `--- PASS: TestAgenticTrunkDoctor (`; exit 0.
- RED-now: M0.
- Green path: M3. A second, cutover-time judgment by the tree-built binary is in AC-016.

### AC-AGT-012 (maps REQ-AGT-013)
Given a statusline render with a branch,
When the workflow is `agentic-trunk` and `git-flow`, the label appears after the branch segment; when it is `github-flow`, absent file, unreadable file or an invalid value, the output is byte-equal to the pre-change golden, and the fake git repository records the same number of calls as before.
- Judge: `go test ./internal/statusline -run '^(TestAgenticTrunkStatuslineLabel|TestAgenticTrunkStatuslineAbsent|TestAgenticTrunkStatuslineNoExtraGit)$' -count=1 -v`
- Pass stdout: `--- PASS: TestAgenticTrunkStatuslineLabel (`, `--- PASS: TestAgenticTrunkStatuslineAbsent (` and `--- PASS: TestAgenticTrunkStatuslineNoExtraGit (`; exit 0.
- RED-now: `[no tests to run]`, exit 0 (observed at M0 on `internal/statusline`).
- Green path: M3.

## D. Pins and guards (M4)

### AC-AGT-013 (maps REQ-AGT-014)
Given a project whose manual workflow is `agentic-trunk` and a template whose manual workflow is `github-flow`,
When `moai update` reconciles `git-strategy.yaml`,
Then the manual workflow is still `agentic-trunk` afterward.
- Judge: `go test ./internal/cli/update -run '^(TestUpdate_GitStrategyValuesSurvive|TestAgenticTrunkValueSurvivesUpdate)$' -count=1 -v`
- Pass stdout: both `--- PASS:` lines; exit 0.
- RED-now: M0.
- Green path: M4.

### AC-AGT-014 (maps REQ-AGT-015)
Given the shipped defaults,
When read,
Then the built-in default configuration and the template yaml carry `github-flow` for all three profiles and no `workflow: agentic-trunk` line.
- Judge A: `go test ./internal/config -run '^TestAgenticTrunkDistributedDefaultsUnchanged$' -count=1 -v` (stdout `--- PASS: TestAgenticTrunkDistributedDefaultsUnchanged (`, exit 0).
- Judge B: `grep -c 'workflow: agentic-trunk' internal/template/templates/.moai/config/sections/git-strategy.yaml.tmpl` (stdout `0`, exit 1).
- RED-now: A `[no tests to run]`; B `0`, exit 1 (already satisfied: a regression guard). Why: pins REQ-AGT-015 before any template edit.
- Green path: A flips at M4; B must stay `0` at every milestone.

### AC-AGT-015 (maps REQ-AGT-016)
Given the sweep guard's arming predicate,
When fed an `agentic-trunk` root, a `git-flow` root and a `github-flow` root,
Then only the github-flow root arms it.
- Judge: `go test ./internal/template -run '^TestGitHubFlowSweepGuard$' -count=1 -v`
- Pass stdout: `--- PASS: TestGitHubFlowSweepGuard (` including the arming subtest with the added agentic-trunk case; exit 0.
- RED-now: M0 (the arming subtest has no agentic-trunk case: `grep -c agentic-trunk internal/template/github_flow_sweep_guard_test.go` prints `0`, exit 1).
- Green path: M4.

## E. Cutover (M6)

### AC-AGT-016 (maps REQ-AGT-017)
Given the flip commit,
When the commit graph and the binary are inspected,
Then (a) the M1 commit is an ancestor of the flip commit, (b) the flip commit touches `.moai/config/sections/git-strategy.yaml` only, (c) a binary built from a tree containing M1 to M3 and invoked by path reports the doctor check OK for `agentic-trunk`, with the judging build's commit named next to the tree HEAD.
- Judge (a): `git merge-base --is-ancestor <M1-sha> <flip-sha>` (stdout empty, exit 0).
- Judge (b): `git show --stat --format=%h <flip-sha>` (stdout lists one path).
- Judge (c): `./bin/moai doctor` run from the tree after `make build` (stdout contains `git_strategy workflow = agentic-trunk` on an OK line); the report states `./bin/moai version` output and `git rev-parse --short HEAD`.
- RED-now: `git log --format=%h -n 1 -S'workflow: agentic-trunk' -- .moai/config/sections/git-strategy.yaml` prints nothing, exit 0 (no flip commit exists, tree `6041c38a4`).
- Green path: M6. The graph witness is the only ordering evidence (verification-claim-integrity section 2.3).

### AC-AGT-017 (maps REQ-AGT-018)
Given the live yaml after the flip,
When searched,
Then the manual profile carries `workflow: agentic-trunk`, personal and team keep `github-flow`, `lead_push_threshold: 20` remains, and the three git-flow-only keys and the stale comment are gone.
- Judge 1: `grep -c 'workflow: agentic-trunk' .moai/config/sections/git-strategy.yaml` (expect `1`).
- Judge 2: `grep -c 'workflow: github-flow' .moai/config/sections/git-strategy.yaml` (expect `2`).
- Judge 3: `grep -c -E 'lead_push_threshold: 20' .moai/config/sections/git-strategy.yaml` (expect `1`).
- Judge 4: `grep -c -E 'develop_branch:|release_branch_prefix:|rc_version_format:|stays git-flow' .moai/config/sections/git-strategy.yaml` (expect `0`, exit 1).
- RED-now (tree `6041c38a4`): Judge 1 prints `0`, exit 1; Judge 4 prints `4`, exit 0 (the three keys plus the line-28 comment `stays git-flow`). Why red: the value has not been flipped.
- Green path: M6.

## F. Guidance (M5, M7, M8)

### AC-AGT-018 (maps REQ-AGT-019)
Given the repo-local guidance,
When searched,
Then the protocol H1 and the AGENTS.local.md section 4 heading name `agentic-trunk`, the protocol path is unchanged, and no live line calls this repository's model "git-flow Lane Protocol" or "switched from GitHub Flow to git-flow".
- Judge 1: `grep -c -E '^# agentic-trunk Lane Protocol' .claude/rules/local/gitflow-lane-protocol.md` (expect `1`).
- Judge 2: `grep -c -E 'git-flow Lane Protocol|switched from GitHub Flow to git-flow' .claude/rules/local/gitflow-lane-protocol.md AGENTS.local.md` (expect `0` per file, exit 1).
- Judge 3: `git ls-files .claude/rules/local/gitflow-lane-protocol.md` (prints the path, exit 0).
- Judge 4: `grep -c agentic-trunk AGENTS.local.md` (expect at least `2`).
- RED-now: Judge 1 prints `0`, exit 1; Judge 2 prints `.claude/rules/local/gitflow-lane-protocol.md:2` and `AGENTS.local.md:0`, exit 0 (the H1 `# git-flow Lane Protocol (moai-adk-go, local-only)` and the blockquote line match); Judge 4 on AGENTS.local.md is `0` today. Why red: the name has not been applied.
- Green path: M7.

### AC-AGT-019 (maps REQ-AGT-020)
Given the seven template rule sources and the skill,
When searched,
Then each names `agentic-trunk`, and each still names `github-flow` as the distributed default.
- Judges (one invocation each, expect at least `1`): `grep -c agentic-trunk <path>` for `internal/template/templates/.claude/rules/moai/workflow/{factory-dispatch-mechanics,factory-dispatch,factory-dispatch-core,worktree-integration,delivery-policy,spec-workflow}.md` and `internal/template/templates/.claude/skills/moai-ref-git-workflow/SKILL.md`.
- Default-preserved judge: `git grep -c 'github-flow' -- internal/template/templates/.claude/rules/moai/workflow/factory-dispatch-mechanics.md` equals its pre-change count recorded at M0 (the count may rise, never fall).
- RED-now: every `grep -c agentic-trunk` prints `0`, exit 1 (observed for `delivery.md`; the rest at M0).
- Green path: M5.

### AC-AGT-020 (maps REQ-AGT-020)
Given `sync/delivery.md`,
When searched,
Then the value-domain line lists three values, a `Strategy: agentic-trunk` heading exists once, and its route block contains no push step.
- Judge 1: `grep -c -E '^The canonical value domain is exactly .github-flow., .git-flow. and .agentic-trunk.' internal/template/templates/.claude/skills/moai/workflows/sync/delivery.md` (expect `1`).
- Judge 2: `grep -c -E '^##### Strategy: agentic-trunk' internal/template/templates/.claude/skills/moai/workflows/sync/delivery.md` (expect `1`).
- Judge 3: `grep -n -E 'Push the integration branch' internal/template/templates/.claude/skills/moai/workflows/sync/delivery.md` (expect exactly one hit, in the git-flow block; the agentic-trunk block has none — read the line numbers against the heading line numbers from Judge 2).
- RED-now: Judge 1 `0` exit 1 (line 22 reads `...exactly github-flow and git-flow.`); Judge 2 `0` exit 1.
- Green path: M5.

### AC-AGT-021 (maps REQ-AGT-021)
Given the edited template sources and their deployed counterparts,
When compared,
Then each pair is byte-identical.
- Judges (each `cmp -s <template> <deployed>`, expect stdout empty, exit 0): the seven rule sources against `.claude/rules/moai/workflow/<same>.md`, `delivery.md` against `.claude/skills/moai/workflows/sync/delivery.md`, and `moai-ref-git-workflow/SKILL.md` against `.claude/skills/moai-ref-git-workflow/SKILL.md`.
- RED-now: M0 records `cmp -s` exit 0 for each unedited pair (the baseline pairs agree); after editing the template alone, the same command exits 1, which is the red this criterion exists to catch.
- Green path: M5 after `make build`.

### AC-AGT-022 (maps REQ-AGT-022)
Given the docs-site,
When searched,
Then all four locale `moai-sync.md` pages name `agentic-trunk`.
- Judge: `grep -l agentic-trunk docs-site/content/ko/workflow-commands/moai-sync.md docs-site/content/en/workflow-commands/moai-sync.md docs-site/content/ja/workflow-commands/moai-sync.md docs-site/content/zh/workflow-commands/moai-sync.md` (expect four paths, exit 0).
- RED-now: `grep -lr agentic-trunk docs-site/content` prints nothing, exit 1 (observed).
- Green path: M8 (sync phase).

### AC-AGT-023 (maps REQ-AGT-023)
Given SPEC-PRIMARY-LANDING-ONLY-001 and this SPEC both edit the shared files,
When the second one lands,
Then `gitflow-lane-protocol.md` contains both this SPEC's title and t1830's Integration-Worktree Rule sentences.
- Judge 1: `grep -c -E '^# agentic-trunk Lane Protocol' .claude/rules/local/gitflow-lane-protocol.md` (expect `1`).
- Judge 2: `grep -c 'Integration-Worktree Rule' .claude/rules/local/gitflow-lane-protocol.md` (expect at least `1` once t1830 has landed; before t1830 lands the count is the current `0` and this judge is not yet applicable — record which case applied).
- RED-now: Judge 1 `0`, exit 1 (observed above); Judge 2 prints `0` today because t1830 has not landed (tree `6041c38a4`).
- Green path: M7, judged at the later of the two landings.

## G. Hygiene (M9)

### AC-AGT-024 (maps REQ-AGT-024)
Given this SPEC's branch,
When the diff against the M0 baseline commit is read,
Then no test skip is added and no timeout is raised.
- Judge 1: `git diff <M0-sha>..HEAD -G't\.Skip' --stat -- '*_test.go'` (expect empty, exit 0).
- Judge 2: `git diff <M0-sha>..HEAD -G'timeout' --stat -- Makefile AGENTS.local.md .github` (expect empty, exit 0).
- RED-now: not applicable before M0 exists; the judges compare against the M0 commit, which is why M0 lands in its own commit (verification-claim-integrity section 2.3).
- Green path: M9.

## Edge cases

- A project that carries `agentic-trunk` in personal or team mode: the target is `main`, the develop seam is empty, `acquire` writes the same unwired-target warning a github-flow project writes (AC-003, AC-007).
- An operator who types `Agentic-Trunk`: classified invalid, doctor warns with the five-value list (AC-001, AC-008).
- Rollback after M6: reverting the flip commit returns a state every shipped binary reads (design.md section 4).

## Definition of Done

All 24 criteria green with the pass stdout named; M6 judged only after the cutover step; the progress record carries, for each cited run, the command, the verbatim stdout, the tree HEAD and the judging binary's commit; Gaps and Residual-risk sections present in the completion report. Quality gates: TRUST 5 on the changed Go packages through the anchored selectors above and `golangci-lint run` scoped to the changed packages; coverage of the new branches by the new tests (every new branch has a test named in section A to D).
