# Progress: SPEC-INTEGRATION-STALE-BASE-001

## §E.1 Plan-phase Audit-Ready Signal

- plan_complete_at: 2026-10-18T00:00:00Z
- plan_status: audit-ready; passed plan audit iter4 (PASS 0.83, audited_sha 5e4fd1197b34be71378a4d86c709b2c5d082bae6, plan_artifact_hash 10fb8cdbcf80a5d3433050fb9a70427608fad855dd0a623f02721e1c465f62dd)
- plan_pin: fb749d372e61a6b10c53091362c76b4412827427
- open_decisions: none blocks Kickoff; Q3 (deferred scope) and Q4 (scheduling) are non-blocking; Q1 and Q2 carry DEFAULT-APPLIED
- gaps: plan.md §G (G1-G14)

audit_ready: true

## §E.2 Run-phase Evidence

Evidence is attributed per verification-claim-integrity §2 and §2.2: every cell names its command and the verbatim output of this run, against this tree. Code state under test: HEAD `b38974aea245f1f3c3b5972360b0507bd7268f00` (the commits after the last code commit touch `progress.md` only). Tests were built from this tree by `go test` (go1.26.8, darwin/arm64). The repository-wide test verdict is owned by the CI run on the project's integration branch and is PENDING at report time.

### E.2.1 Commit graph (ordering witnessed by the commit graph, §2.3)

```
8897f196d feat(SPEC-INTEGRATION-STALE-BASE-001): M1 start - status in-progress and phase 4 mode selection (card t1834)
9272392f6 test(SPEC-INTEGRATION-STALE-BASE-001): M1 RED config surface tests for the base-freshness gate (card t1834)
8532dc06c feat(SPEC-INTEGRATION-STALE-BASE-001): M1 GREEN config surface for the base-freshness gate (card t1834)
cd0b7b01f test(SPEC-INTEGRATION-STALE-BASE-001): M1 RED gate function tests for the base-freshness gate (card t1834)
b6bc98567 test(SPEC-INTEGRATION-STALE-BASE-001): M1 RED refine the threshold-literal scanner (card t1834)
9f407c528 feat(SPEC-INTEGRATION-STALE-BASE-001): M2 GREEN the pure base-freshness gate function (card t1834)
c59b56898 test(SPEC-INTEGRATION-STALE-BASE-001): M3 RED CLI call-site tests for the base-freshness gate (card t1834)
daa853b62 test(SPEC-INTEGRATION-STALE-BASE-001): M3 RED tighten the adoption fixture so the exemption is load-bearing (card t1834)
2f4cc07a1 test(SPEC-INTEGRATION-STALE-BASE-001): M2 cover the unreadable-git-answer branches of the gate (card t1834)
037b40dbe feat(SPEC-INTEGRATION-STALE-BASE-001): M3 GREEN wire the base-freshness gate into acquire and complete (card t1834)
ace76cb07 docs(SPEC-INTEGRATION-STALE-BASE-001): M4 guidance section for the pre-integration base freshness gate (card t1834)
b38974aea test(SPEC-INTEGRATION-STALE-BASE-001): M3 use os.DevNull for the fixture git configuration (card t1834)
```

RED precedes GREEN in the graph for every milestone: M1 config 9272392f6 -> 8532dc06c; M2 gate cd0b7b01f + b6bc98567 -> 9f407c528; M3 call sites c59b56898 + daa853b62 -> 037b40dbe. Two test commits postdate their GREEN and are labelled so: 2f4cc07a1 (post-hoc coverage of three unreadable-git-answer branches, each shown red against a mutant first) and b38974aea (os.DevNull portability). Commit 8897f196d ends with the `Authored-By-Agent` trailer rather than the `🗿 MoAI` marker (the status-transition commit; the trailer is intact for the ownership rule's own predicate, and every later commit ends with the marker).

### E.2.2 RED evidence (E8) - verbatim, captured before each GREEN

The committed RED of each milestone is a compile failure on the symbol the test names (closure rule). M1 config, `go test ./internal/config/ -run '^(TestWorkflowBaseFreshnessGateDefaults|TestWorkflowBaseFreshnessGateLoaderRoundTrip|TestWorkflowBaseFreshnessGateTemplateParity|TestBaseFreshnessTemplateBlockSelfCheck)$' -v -count=1`, exit 1:

```
# github.com/modu-ai/moai-adk/internal/config [github.com/modu-ai/moai-adk/internal/config.test]
internal/config/workflow_base_freshness_gate_test.go:25:18: cfg.Workflow.BaseFreshnessGate undefined (type WorkflowConfig has no field or method BaseFreshnessGate)
internal/config/workflow_base_freshness_gate_test.go:28:25: cfg.Workflow.BaseFreshnessGate undefined (type WorkflowConfig has no field or method BaseFreshnessGate)
internal/config/workflow_base_freshness_gate_test.go:53:19: cfg.Workflow.BaseFreshnessGate undefined (type WorkflowConfig has no field or method BaseFreshnessGate)
internal/config/workflow_base_freshness_gate_test.go:56:25: cfg.Workflow.BaseFreshnessGate undefined (type WorkflowConfig has no field or method BaseFreshnessGate)
internal/config/workflow_base_freshness_gate_test.go:95:37: NewDefaultConfig().Workflow.BaseFreshnessGate undefined (type WorkflowConfig has no field or method BaseFreshnessGate)
FAIL	github.com/modu-ai/moai-adk/internal/config [build failed]
FAIL
```

M1/M2 gate function, `go test ./internal/factory/ -run '^(TestLandingsBehindCountsFirstParent|TestBaseFreshnessKBoundary|TestBaseFreshnessRequiresValidRecordForCurrentTree|TestBaseFreshnessFailOpenCauses|TestBaseFreshnessGateOffAdvisesOnlyPure|TestBaseFreshnessK0MatchesMergeCause2|TestBaseFreshnessGateWritesNothing|TestBaseFreshnessExitCodeIsDistinct|TestBaseFreshnessDecisionCodeHasNoThresholdLiteral)$' -v -count=1`, exit 1:

```
# github.com/modu-ai/moai-adk/internal/factory [github.com/modu-ai/moai-adk/internal/factory.test]
internal/factory/base_freshness_test.go:139:48: undefined: BaseFreshnessInput
internal/factory/base_freshness_test.go:140:9: undefined: BaseFreshnessInput
internal/factory/base_freshness_test.go:165:9: undefined: EvaluateBaseFreshness
internal/factory/base_freshness_test.go:165:51: undefined: BaseFreshnessSeams
internal/factory/base_freshness_test.go:175:8: undefined: EvaluateBaseFreshness
internal/factory/base_freshness_test.go:175:48: undefined: BaseFreshnessSeams
internal/factory/base_freshness_test.go:187:8: undefined: EvaluateBaseFreshness
internal/factory/base_freshness_test.go:187:48: undefined: BaseFreshnessSeams
internal/factory/base_freshness_test.go:191:36: undefined: BaseFreshnessSentinel
internal/factory/base_freshness_test.go:205:10: undefined: EvaluateBaseFreshness
internal/factory/base_freshness_test.go:205:10: too many errors
FAIL	github.com/modu-ai/moai-adk/internal/factory [build failed]
FAIL
```

M3 CLI call sites, `go test ./internal/cli/ -run '^(TestIntegrationAcquireRefusesStaleBase|TestBaseFreshnessGateOffAdvisesOnly|TestBaseFreshnessCLIUncertaintyAdmitsWithCauseLine|TestBaseFreshnessRefusalSentinelAndExitCode|TestFactoryCompleteRefusesStaleBaseBeforeAcquire|TestBaseFreshnessKDoesNotRelaxMergeStep|TestFactoryCompleteAdoptionNotRefusedByBaseFreshness)$' -v -count=1`, exit 1:

```
# github.com/modu-ai/moai-adk/internal/cli [github.com/modu-ai/moai-adk/internal/cli.test]
internal/cli/integration_base_freshness_test.go:265:12: undefined: baseFreshnessSeams
internal/cli/integration_base_freshness_test.go:266:4: undefined: baseFreshnessSeams
internal/cli/integration_base_freshness_test.go:272:25: undefined: baseFreshnessSeams
FAIL	github.com/modu-ai/moai-adk/internal/cli [build failed]
FAIL
```

A compile failure proves little about behaviour, so the CLI tests were also run once against the pre-implementation tree with ONLY the seam variable `baseFreshnessSeams` supplied by an uncommitted scratch file (removed afterwards). Behavioural RED, verbatim failing lines (fixture version of commit c59b56898; the adoption fixture was tightened afterwards, see E.2.5):

```
    integration_base_freshness_test.go:210: the refusal must start with BASE_FRESHNESS_VIOLATION:, got "integration window: your ticket timed out after 2s at queue position 1 behind lane-9 — re-acquire with --wait re-enters at the ...
--- FAIL: TestIntegrationAcquireRefusesStaleBase (2.68s)
    integration_base_freshness_test.go:233: want exactly one advisory line on standard error, got []
--- FAIL: TestBaseFreshnessGateOffAdvisesOnly (0.35s)
    integration_base_freshness_test.go:309: want exactly one line naming "configuration unreadable", got []
    integration_base_freshness_test.go:309: want exactly one line naming "negative", got []
    integration_base_freshness_test.go:309: want exactly one line naming "no-such-branch", got []
    integration_base_freshness_test.go:309: want exactly one line naming "injected rev-list failure", got []
--- FAIL: TestBaseFreshnessCLIUncertaintyAdmitsWithCauseLine (4.42s)
    --- FAIL: TestBaseFreshnessCLIUncertaintyAdmitsWithCauseLine/unreadable_configuration (0.35s)
    --- FAIL: TestBaseFreshnessCLIUncertaintyAdmitsWithCauseLine/negative_K (0.56s)
    --- FAIL: TestBaseFreshnessCLIUncertaintyAdmitsWithCauseLine/integration_branch_unresolvable (0.67s)
    --- FAIL: TestBaseFreshnessCLIUncertaintyAdmitsWithCauseLine/git_failure_injected (0.57s)
    integration_base_freshness_test.go:325: expected a base-freshness refusal, got success
    integration_base_freshness_test.go:333: expected a base-freshness refusal, got success
    integration_base_freshness_test.go:342: expected a base-freshness refusal, got success
--- FAIL: TestBaseFreshnessRefusalSentinelAndExitCode (2.73s)
    --- FAIL: TestBaseFreshnessRefusalSentinelAndExitCode/landings_behind_exceeds_K=0 (0.62s)
    --- FAIL: TestBaseFreshnessRefusalSentinelAndExitCode/landings_behind_exceeds_K=2 (0.78s)
    --- FAIL: TestBaseFreshnessRefusalSentinelAndExitCode/no_valid_record (0.44s)
    integration_base_freshness_test.go:370: the refusal must start with BASE_FRESHNESS_VIOLATION:, got "factory complete: the merge step refused: integration merge: the integration branch moved past the record's base 99cd178ada4e  ...
--- FAIL: TestFactoryCompleteRefusesStaleBaseBeforeAcquire (4.04s)
    integration_base_freshness_test.go:444: the refusal must start with BASE_FRESHNESS_VIOLATION:, got "factory complete: refused — no valid re-measure record for the candidate tree ea4e55980979 (no re-measure record for tree ea4e ...
--- FAIL: TestFactoryCompleteAdoptionNotRefusedByBaseFreshness (5.57s)
    --- FAIL: TestFactoryCompleteAdoptionNotRefusedByBaseFreshness/a_commit_after_the_merge_is_not_an_adoption;_the_gate_evaluates (2.54s)
```

### E.2.3 AC matrix (E1) - green path at the code state above

Swept counts are the top-level `--- PASS` lines; every count is non-zero and every FAIL count is 0.

**AC-ISB-001** PASS

- `go test ./internal/cli/ -run ^TestIntegrationAcquireRefusesStaleBase$ -v -count=1`
  -> exit=0 | --- PASS: TestIntegrationAcquireRefusesStaleBase (1.72s) | PASS | ok  	github.com/modu-ai/moai-adk/internal/cli	2.725s | top-level PASS count: 1  FAIL count: 0

**AC-ISB-002** PASS

- `go test ./internal/cli/ -run ^TestFactoryCompleteRefusesStaleBaseBeforeAcquire$ -v -count=1`
  -> exit=0 | --- PASS: TestFactoryCompleteRefusesStaleBaseBeforeAcquire (1.82s) | PASS | ok  	github.com/modu-ai/moai-adk/internal/cli	2.856s | top-level PASS count: 1  FAIL count: 0

**AC-ISB-003** PASS

- `go test ./internal/factory/ -run ^TestLandingsBehindCountsFirstParent$ -v -count=1`
  -> exit=0 | --- PASS: TestLandingsBehindCountsFirstParent (1.79s) | PASS | ok  	github.com/modu-ai/moai-adk/internal/factory	2.012s | top-level PASS count: 1  FAIL count: 0

**AC-ISB-004** PASS

- `go test ./internal/factory/ -run ^TestBaseFreshnessKBoundary$ -v -count=1`
  -> exit=0 | --- PASS: TestBaseFreshnessKBoundary (1.37s) | PASS | ok  	github.com/modu-ai/moai-adk/internal/factory	1.578s | top-level PASS count: 1  FAIL count: 0

**AC-ISB-005** PASS

- `go test ./internal/factory/ -run ^TestBaseFreshnessRequiresValidRecordForCurrentTree$ -v -count=1`
  -> exit=0 | --- PASS: TestBaseFreshnessRequiresValidRecordForCurrentTree (3.15s) | PASS | ok  	github.com/modu-ai/moai-adk/internal/factory	3.352s | top-level PASS count: 1  FAIL count: 0

**AC-ISB-007** PASS

- `go test ./internal/config/ -run ^(TestWorkflowBaseFreshnessGateDefaults|TestWorkflowBaseFreshnessGateLoaderRoundTrip|TestWorkflowBaseFreshnessGateTemplateParity)$ -v -count=1`
  -> exit=0 | --- PASS: TestWorkflowBaseFreshnessGateDefaults (0.00s) | --- PASS: TestWorkflowBaseFreshnessGateTemplateParity (0.00s) | --- PASS: TestWorkflowBaseFreshnessGateLoaderRoundTrip (0.00s) | PASS | ok  	github.com/modu-ai/moai-adk/internal/config	0.187s | top-level PASS count: 3  FAIL count: 0

- supplementary (the source-scan clause, the exit-code distinctness of AC-ISB-011, the pure advisory of AC-ISB-008 and the unreadable-git branches of AC-ISB-009): `go test ./internal/factory/ -run '^(TestBaseFreshnessDecisionCodeHasNoThresholdLiteral|TestBaseFreshnessExitCodeIsDistinct|TestBaseFreshnessUncertainGitAnswers|TestBaseFreshnessGateOffAdvisesOnlyPure)$' -v -count=1`
  -> exit=0 | --- PASS: TestBaseFreshnessUncertainGitAnswers | --- PASS: TestBaseFreshnessGateOffAdvisesOnlyPure | --- PASS: TestBaseFreshnessExitCodeIsDistinct | --- PASS: TestBaseFreshnessDecisionCodeHasNoThresholdLiteral | ok internal/factory 2.335s | top-level PASS count: 4  FAIL count: 0

**AC-ISB-008** PASS

- `go test ./internal/cli/ -run ^TestBaseFreshnessGateOffAdvisesOnly$ -v -count=1`
  -> exit=0 | --- PASS: TestBaseFreshnessGateOffAdvisesOnly (0.66s) | PASS | ok  	github.com/modu-ai/moai-adk/internal/cli	1.637s | top-level PASS count: 1  FAIL count: 0

**AC-ISB-009** PASS

- `go test ./internal/factory/ -run ^TestBaseFreshnessFailOpenCauses$ -v -count=1`
  -> exit=0 | --- PASS: TestBaseFreshnessFailOpenCauses (3.10s) | PASS | ok  	github.com/modu-ai/moai-adk/internal/factory	3.294s | top-level PASS count: 1  FAIL count: 0

- `go test ./internal/cli/ -run ^TestBaseFreshnessCLIUncertaintyAdmitsWithCauseLine$ -v -count=1`
  -> exit=0 | --- PASS: TestBaseFreshnessCLIUncertaintyAdmitsWithCauseLine (5.75s) | PASS | ok  	github.com/modu-ai/moai-adk/internal/cli	7.017s | top-level PASS count: 1  FAIL count: 0

**AC-ISB-010** PASS

- `go test ./internal/factory/ -run ^TestBaseFreshnessK0MatchesMergeCause2$ -v -count=1`
  -> exit=0 | --- PASS: TestBaseFreshnessK0MatchesMergeCause2 (4.27s) | PASS | ok  	github.com/modu-ai/moai-adk/internal/factory	4.453s | top-level PASS count: 1  FAIL count: 0

- `go test ./internal/factory/ -run ^(TestMergeStepPreMergeCausesReleaseWithDistinctCodes|TestMergeStepMidstepBranchTipMoveRefusesMerge)$ -v -count=1`
  -> exit=0 | --- PASS: TestMergeStepPreMergeCausesReleaseWithDistinctCodes (7.04s) | --- PASS: TestMergeStepMidstepBranchTipMoveRefusesMerge (0.57s) | PASS | ok  	github.com/modu-ai/moai-adk/internal/factory	7.861s | top-level PASS count: 2  FAIL count: 0

- `go test ./internal/cli/ -run ^TestBaseFreshnessKDoesNotRelaxMergeStep$ -v -count=1`
  -> exit=0 | --- PASS: TestBaseFreshnessKDoesNotRelaxMergeStep (1.97s) | PASS | ok  	github.com/modu-ai/moai-adk/internal/cli	2.895s | top-level PASS count: 1  FAIL count: 0

**AC-ISB-011** PASS

- `go test ./internal/cli/ -run ^TestBaseFreshnessRefusalSentinelAndExitCode$ -v -count=1`
  -> exit=0 | --- PASS: TestBaseFreshnessRefusalSentinelAndExitCode (4.74s) | PASS | ok  	github.com/modu-ai/moai-adk/internal/cli	5.924s | top-level PASS count: 1  FAIL count: 0

**AC-ISB-013** PASS

- `go test ./internal/config/ -run ^TestShippedConfigKeysHaveReaders$ -v -count=1`
  -> exit=0 | --- PASS: TestShippedConfigKeysHaveReaders (5.67s) | PASS | ok  	github.com/modu-ai/moai-adk/internal/config	5.833s | top-level PASS count: 1  FAIL count: 0

**AC-ISB-014** PASS

- `go test ./internal/factory/ -run ^TestBaseFreshnessGateWritesNothing$ -v -count=1`
  -> exit=0 | --- PASS: TestBaseFreshnessGateWritesNothing (0.48s) | PASS | ok  	github.com/modu-ai/moai-adk/internal/factory	0.732s | top-level PASS count: 1  FAIL count: 0

**AC-ISB-016** PASS

- `go test ./internal/cli/ -run ^TestFactoryCompleteAdoptionNotRefusedByBaseFreshness$ -v -count=1`
  -> exit=0 | --- PASS: TestFactoryCompleteAdoptionNotRefusedByBaseFreshness (8.65s) | PASS | ok  	github.com/modu-ai/moai-adk/internal/cli	9.973s | top-level PASS count: 1  FAIL count: 0

**AC-ISB-006** PASS

```
$ grep -rn 'type RemeasureRecord struct' internal --include='*.go'
internal/factory/integration_remeasure.go:45:type RemeasureRecord struct {
exit=0
$ grep -c 'ValidateRemeasureRecord' internal/factory/base_freshness.go
2
exit=0
$ grep -c 'ReadRemeasureRecord' internal/factory/base_freshness.go
2
exit=0
$ grep -n '"merge"' internal/factory/base_freshness.go
exit=1 (1 = no git merge invocation)
```

**AC-ISB-012** PASS (the template-block neutrality half is `TestWorkflowBaseFreshnessGateTemplateParity` in G-07)

```
$ grep -c 'base freshness' .claude/rules/moai/workflow/factory-dispatch-gates.md
2
exit=0
$ diff -q .claude/rules/moai/workflow/factory-dispatch-gates.md internal/template/templates/.claude/rules/moai/workflow/factory-dispatch-gates.md
exit=0 (silent)
```

**AC-ISB-013** PASS (G-13b above, plus)

```
$ grep -c 'workflow.base_freshness_gate' internal/config/testdata/shipped_key_inventory.yaml
2
exit=0
$ grep -n 'configCacheSchemaVersion = ' internal/config/cache.go
54:const configCacheSchemaVersion = 13
exit=0
```

**AC-ISB-014** PASS (G-14b above, plus)

```
$ grep -n -E 'homestate|card_evidence' internal/factory/base_freshness.go internal/cli/integration_base_freshness.go
exit=1 (1 = no match, both files present; positive control: G-06b printed 2)
```

**AC-ISB-015** PASS (plan pin fb749d372e61a6b10c53091362c76b4412827427)

```
$ git diff --name-only fb749d372e61a6b10c53091362c76b4412827427 HEAD
.claude/rules/moai/workflow/factory-dispatch-gates.md
.moai/config/sections/workflow.yaml
.moai/reports/t1834/plan-audit-iter1.md
.moai/reports/t1834/plan-audit-iter2.md
.moai/reports/t1834/plan-audit-iter3.md
.moai/reports/t1834/plan-audit-iter4.md
.moai/specs/SPEC-INTEGRATION-STALE-BASE-001/acceptance.md
.moai/specs/SPEC-INTEGRATION-STALE-BASE-001/decision-index.md
.moai/specs/SPEC-INTEGRATION-STALE-BASE-001/plan.md
.moai/specs/SPEC-INTEGRATION-STALE-BASE-001/progress.md
.moai/specs/SPEC-INTEGRATION-STALE-BASE-001/spec.md
internal/cli/factory_card.go
internal/cli/integration.go
internal/cli/integration_base_freshness.go
internal/cli/integration_base_freshness_test.go
internal/config/cache.go
internal/config/defaults.go
internal/config/testdata/shipped_key_inventory.yaml
internal/config/types.go
internal/config/workflow_base_freshness_gate_test.go
internal/factory/base_freshness.go
internal/factory/base_freshness_test.go
internal/template/templates/.claude/rules/moai/workflow/factory-dispatch-gates.md
internal/template/templates/.moai/config/sections/workflow.yaml
exit=0
$ git diff --name-only fb749d372e61a6b10c53091362c76b4412827427 HEAD -- AGENTS.local.md .claude/rules/local internal/cli/session_worktree.go internal/cli/local_main_resync.go internal/cli/integration_merge.go internal/factory/integration_merge_step.go internal/factory/integration_remeasure.go
exit=0 (empty)
$ git diff --name-only fb749d372e61a6b10c53091362c76b4412827427 HEAD -- internal/factory/base_freshness.go
internal/factory/base_freshness.go
exit=0 (control prints the gate path)
```


### E.2.4 Build, boundary, lint, coverage (E2-E5)

**E2 cross-platform build.** Scoped, observed: `go build ./internal/factory/ ./internal/config/ ./internal/cli/` -> exit 0. NOT observed (Gap 1): `go build ./...`, `GOOS=windows GOARCH=amd64 go build ./...`, `make build` (lease `go-test-heavy` held by another session: `moai slot acquire --resource go-test-heavy --max-duration 15m` -> exit 3, not forced). The new non-test files import only `errors`, `fmt`, `strconv`, `strings`, `io`, `os`, `path/filepath`, cobra and in-repo packages; no `syscall`, no build-tagged code.

**E3 coverage of the new non-test files** (profile built from the gate's own tests only, `-coverprofile`, this run): `internal/factory/base_freshness.go` 61/61 statements = 100.0% (`TestLandingsBehindCountsFirstParent`, `TestBaseFreshnessKBoundary`, `TestBaseFreshnessRequiresValidRecordForCurrentTree`, `TestBaseFreshnessFailOpenCauses`, `TestBaseFreshnessUncertainGitAnswers`, `TestBaseFreshnessGateOffAdvisesOnlyPure`, `TestBaseFreshnessK0MatchesMergeCause2`, `TestBaseFreshnessGateWritesNothing`, `TestBaseFreshnessExitCodeIsDistinct`, `TestBaseFreshnessDecisionCodeHasNoThresholdLiteral`); `internal/cli/integration_base_freshness.go` 21/24 = 87.5% (uncovered: the `cfg == nil` guard after a nil-error load, and the `os.Getwd` failure branch); `internal/cli/factory_card.go` `factoryAdoptedTip` 100.0%. The config change adds a struct, a default and a cache constant (no executable statements). Package-level coverage of `internal/factory` / `internal/cli` was not measured (Gap 2).

**E4 subagent-boundary grep.** `grep -n 'AskUserQuestion' internal/factory/base_freshness.go internal/cli/integration_base_freshness.go` -> no output, exit 1. The template command `grep -rn 'AskUserQuestion' internal/factory internal/config internal/cli | grep -v '_test.go' | grep -v '// '` prints only documentation strings in files this card did not touch (`internal/cli/harness.go`, `pr_watch_cmd.go`, `harness_mute.go`, `harness_clusters.go`, and `testdata/` rollouts); `git diff f7133cfdc..HEAD -- internal | grep '^+' | grep -c 'AskUserQuestion'` -> `0`.

**E5 lint.** `gofmt -l internal/factory internal/cli internal/config` -> no output. `go vet ./internal/factory/ ./internal/config/ ./internal/cli/` -> exit 0. `golangci-lint run --timeout=4m ./internal/factory/ ./internal/config/ ./internal/cli/` (golangci-lint v2.1.6, built with go1.26.8; a tool the repository's Makefile does not pin, so its provenance is the installed binary) -> exit 1 with `3 issues: errcheck: 3`, all in `internal/cli/factory_run_adopt.go:79,91,103` (`Error return value of fmt.Fprintf is not checked`), a file `git diff --stat f7133cfdc..HEAD` shows unchanged: baseline, none new. Template-package audits that guard rule and template content (`go test ./internal/template/ -run '^(TestCardID|TestWorkflowRulePathsPinned|TestDeclaredRuleMirrorForks|TestRuleTemplateMirrorDrift|TestTemplateNoInternalContentLeak|TestLeakClass|TestDeployedRuleSectionRefsResolve|TestDeployedAlwaysLoaded|TestSanitizedPairParity|TestRoleCoreStubGate)' -count=1 -v`) -> 18 top-level `--- PASS`, 0 FAIL, `ok ... 4.784s`. `moai spec lint .moai/specs/SPEC-INTEGRATION-STALE-BASE-001/spec.md --json --strict` -> `[]`, exit 0 (installed build `ee981a932`, an ancestor of HEAD, not a build of this tree: Gap 3).

**E6 push state.** Nothing pushed, no PR, no tag (the delegation forbids them). Commits listed in E.2.1; `git status --short` is read at the end of the run and reported in the completion message.

**Regression families (E1 supplement).** CLI regression family on the M3 code state, `go test ./internal/cli/ -run '^(TestMWQ|TestSD_AC0|TestAcquire|TestIntegration|TestFactoryComplete|TestDrift|TestSettingsDrift|TestCandidate)' -count=1 -v` -> 129 top-level `--- PASS`, 0 FAIL, `ok ... 311.314s`. `go test ./internal/config/ -count=1` (M1 state) -> `ok ... 17.864s`. `go test ./internal/template/embedemit/ -run 'Test(GoldenCommittedArtifactsMatchEmission|TemplatesDiskMatchesManifest)$' -count=1` -> `ok`.


### E.2.5 Mutant probes (verification-completeness §2)

Each mutant was applied to the working tree, the named test run, and the file restored (`diff -q` against the saved copy printed nothing). A test that stays green under its mutant is vacuous; the first adoption fixture did exactly that and was replaced (row 9).

| # | Mutant | Test run | Observed |
|---|--------|----------|----------|
| 1 | drop `--first-parent` from the `rev-list` count | `TestLandingsBehindCountsFirstParent` | `--- FAIL` |
| 2 | `behind >= in.K` instead of `>` | `TestBaseFreshnessKBoundary` | `--- FAIL` |
| 3 | constant tolerance (`behind > 0`) | `TestBaseFreshnessKBoundary`, `TestBaseFreshnessDecisionCodeHasNoThresholdLiteral` | both `--- FAIL` |
| 4 | a missing record read as a valid one | `TestBaseFreshnessRequiresValidRecordForCurrentTree` | `--- FAIL` |
| 5 | treat any ancestry-probe error as staleness | `TestBaseFreshnessUncertainGitAnswers` | `--- FAIL` (ancestry subtest) |
| 6 | ignore a failed tree read | `TestBaseFreshnessUncertainGitAnswers` | `--- FAIL` (tree subtest) |
| 7 | ignore a non-numeric count | `TestBaseFreshnessUncertainGitAnswers` | `--- FAIL` (count subtest) |
| 8 | drop the notice write; refusal exit code 1 instead of 16 | `TestBaseFreshnessCLIUncertaintyAdmitsWithCauseLine`, `TestBaseFreshnessGateOffAdvisesOnly`, `TestBaseFreshnessRefusalSentinelAndExitCode` | all `--- FAIL` |
| 9 | no adoption exemption (`if false && factoryAdoptedTip(...)`) | `TestFactoryCompleteAdoptionNotRefusedByBaseFreshness` | first fixture: PASS (vacuous: re-measuring the merged tree rewrote the candidate's record with the tip as base). Tightened fixture (commit daa853b62): `--- FAIL` with `card t1 is 1 landings behind local develop`, and green with the real code |


### E.2.6 Gaps and residual risk

**Gaps (explicitly NOT observed):**

1. **Heavy repository-wide gates not run.** `go build ./...`, `GOOS=windows GOARCH=amd64 go build ./...` and `make build` need the single `go-test-heavy` lease. `moai slot acquire --resource go-test-heavy --max-duration 15m` exited 3 twice (held by `lane-2`, then `lane-22`, both live and within their bound); it was not forced and not waited on. Commands pending, in order: `moai slot acquire --resource go-test-heavy --max-duration 15m`, `go build ./...`, `GOOS=windows GOARCH=amd64 go build ./...`, `make build`, `git status --short` (expect the generated catalog and embed files unchanged), `moai slot release --resource go-test-heavy`. What was observed instead is scoped: `go build ./internal/cli/`, `go vet` over the three changed packages, and `make build`'s read-only `embed-manifest-check` sub-check (`go test ./internal/template/embedemit/ -run 'Test(GoldenCommittedArtifactsMatchEmission|TemplatesDiskMatchesManifest)$' -count=1` -> `ok`). The other `make build` prerequisites (`agents-emit-check`, `commands-emit-check`, `tool-policy-drift-check`, `templ-generate`, `gen-catalog-hashes.go --all`) were not run; no agent, command, skill, settings or templ file is among the changed paths (E.2.3 AC-ISB-015 listing).
2. **No full-package suite for `internal/factory` or `internal/cli`.** Run: the nine gate tests and the merge-step regression pair in `internal/factory`; the seven CLI gate tests plus a 129-test family (`^(TestMWQ|TestSD_AC0|TestAcquire|TestIntegration|TestFactoryComplete|TestDrift|TestSettingsDrift|TestCandidate)`, all `--- PASS`, `ok ... 311.314s`, run on the M3 code state before the docs commit and the two test-only follow-ups) and the whole `internal/config` package (`ok ... 17.864s`, M1 state). Package-level coverage (the 85% target) is therefore unmeasured: the profiles below cover the gate's own tests only (`internal/factory` 8.3%, `internal/cli` 5.7% of their packages by statements). Per-file coverage of the new non-test files is reported in E.2.4.
3. **`moai spec lint` was judged by an installed build.** `moai spec lint .moai/specs/SPEC-INTEGRATION-STALE-BASE-001/spec.md --json --strict` printed `[]`, exit 0, with the installed `moai-adk v3.2.0-rc.36-t1822c` (commit `ee981a932`), which `git merge-base --is-ancestor ee981a932 HEAD` shows is an ancestor of this tree's HEAD (exit 0): not a build of this tree (verification-claim-integrity §2.2). Building one is a `go build` of `./cmd/moai`, which is a heavy build.
4. **The template-package audits that scan rule and template files** (`internal/template` card-id leak, rule-mirror, content-leak, section-ref, always-loaded-budget tests) are reported in E.2.4 if they ran; if that row is absent they did not.
5. **K above 0 as queue-wait slack was not exercised with a live queue.** The tests pin the entry side (K=3 admits two landings behind) and that the merge step still exits 2 with the window released; a lane re-absorbing while queued and then merging was not run.
6. **`first-parent` counting was exercised on merge-commit landings and direct commits only** (plan Gap G2): not on fast-forwarded resyncs, squash-merged batches, or a main that received both in one range.
7. **Commit 8897f196d (the status-transition commit) ends with the `Authored-By-Agent` trailer, not the `🗿 MoAI` marker** the delegation asked every commit to end with. No commit was amended (the delegation forbids it); the ownership rule's own predicate reads the trailer, and every later commit ends with the marker.

**Residual risk (what could still be wrong despite what was observed):**

- Any failure of `ReadRemeasureRecord` is judged "no valid record", because that verifier loses `os.ErrNotExist` before returning (it is not in this card's allow-list). With the gate enabled, an unreadable record file therefore refuses rather than admits; the refusal text names the underlying error.
- `moai factory complete` writes its one notice line to its output stream (stdout) because the shared body has no error stream; `moai integration acquire` writes to standard error. A caller that parses complete's stdout sees the extra line only when the gate is uncertain or off-and-stale.
- Both window-entry verbs are also edited by the neighbouring landing-surface card (t1847). Each gained exactly one added call and nothing was reordered, but the order the two cards land in is a leader decision (decision-index Q4).
- The CLI tests run the verbs in-process with a changed working directory and a pinned `CLAUDE_PROJECT_DIR`; a lane whose real session environment names a different project root than its worktree was not reproduced.
- Repository-wide test and build verdicts belong to the CI run on the project's integration branch and are PENDING at report time.

### E.2.8 iter12 F1/F2 repair round (lane-6, 2026-10-21)

Context: the leader re-dispatched the card to lane-6 (run tn8rpj) on 10-21 with completion instructions. Lane re-measured 103 first-parent landings behind local main, absorbed (merge 2439da8c90; re-measured 0 behind, `git merge-base --is-ancestor main HEAD` rc=0), then repaired the two P2 findings of sync-audit iter12. Fix lines verified on disk by the lane: factory_card.go:2889 (`record.Tree == tipTree` in `factoryAdoptedTip`) and factory_card.go:3047 (`git symbolic-ref --quiet HEAD`, `refs/heads/` stripped, in `factoryBranchOfWorktree`).

| Commit | Content |
|---|---|
| 26dad7f9ef | RED F1 record-identity test (TestFactoryCompleteAdoptionRequiresRecordTreeIdentity) |
| e9e95f4d40 | F1 fix: adoption exemption requires the record to name the tip tree |
| 7891703291 | RED F2 same-name-tag test v1 (TestFactoryBranchOfWorktreeWithSameNameTag) |
| 1eaf9bb649 | F2 end-to-end rescope to the adoption path (the merge-step path carries an out-of-scope defect, below) |
| 1a16da10f6 | F2 fix: branch read is symbolic-ref with refs/heads/ stripped |

Confirming scoped runs at 1a16da10f6 (manager-develop Section E, env-scrubbed): `go test ./internal/factory/ -run 'Freshness|Remeasure'` ok 12.654s; `go test ./internal/config/ -run 'BaseFreshness|ShippedConfigKeys'` ok 3.655s; `go test ./internal/cli/ -run 'Freshness|TestIntegrationAcquireRefusesStaleBase|TestFactoryComplete|TestAcquire|TestIntegration'` ok 109.710s. `gofmt -l internal/cli/` clean, `go vet ./internal/cli/` exit 0. RED-verbatim evidence for both fixes recorded in the manager-develop report (F1: adoption exemption fired on a forged record pre-fix; F2: `heads/WT-card-branch` read + complete failure on a same-named tag). Transparency disclosure carried: 26dad7f9ef was re-created via `git reset --soft HEAD~1` seconds after its first creation repaired a mangled trailer; no `--amend`, nothing pushed, nothing cited the broken form; commit-graph order (test commit before fix commit) witnesses RED-before-GREEN.

Scope delta: the non-.moai diff against main is now 19 paths — iter12's 18 plus `internal/cli/factory_card_pr_gitenv_test.go` (the F2 regression test's home). All five commit bodies carry the `🗿 MoAI` marker and `card t1834`.

New out-of-scope finding (observed live, NOT repaired — pre-existing on a non-card path): `internal/factory/integration_merge_step.go:950` reads the branch with `rev-parse --abbrev-ref HEAD`, the same class as F2; a tag sharing the card branch's name makes the merge step refuse ("is on heads/WT-alpha, not a WT- branch") on the non-adoption complete path. iter12's F3 named only `internal/cli/integration.go:126`. Left to separate routing (scope discipline; the card's plan-pin preserve check showed integration_merge_step.go untouched by this card).

Decision record (§G continuation): the 10-21 leader re-dispatch with completion instructions is read as the renewed operator-side authorization after the d-20261018T222141Z-398b ceiling exception ended iter12's rounds; exactly ONE confirming audit (iter13, codex required) runs on this authorization. decided_by=lane-6 orchestrator; evidence_refs=.moai/reports/t1834/sync-audit-iter12.md, the dispatch message, iter12 Recommendations; ladder_path=① disk evidence.

(Placement note: §E.2.8 above sits before §E.2.7 — the newest addendum was inserted at the section head; the pre-existing §E.2.7 body is preserved verbatim below.)

### E.2.7 Audit-fix round (sync-audit codex findings 1-3)

Appended; no earlier evidence rewritten. RED `187e58c1b` + `da0e3e840`, GREEN `619402f00` (commit graph orders RED before GREEN).

| Finding | RED reproduction (observed) | Fix | Test |
|---|---|---|---|
| 1 [P1] refusal silent | real process, pre-fix build: `exit=16 stdout=[] stderr=[]`; fixed build: exit 16, sentinel line on stderr | `checkBaseFreshness` writes the refusal to the verb's writer before returning the exit-16 error | `TestBaseFreshnessRefusalReachesTheUserThroughTheRootErrorPath`, `TestFactoryCompleteRefusalIsVisible` (both through `runFang`) |
| 2 [P2] foreign tree field accepted | `Admit:true` for a record file named for the current tree whose `tree` names another | `EvaluateBaseFreshness` requires `rec.Tree == res.Tree` after validation | `TestBaseFreshnessRefusesRecordNamingAnotherTree` (K=0, K=2) |
| 3 [P2] config failure read as gate-off | stderr `base-freshness (advisory, gate off): would refuse ...` for malformed YAML, wrong-typed key, unreadable file | `workflowConfigFailure` routes a present-but-unloaded `workflow.yaml` to `ConfigErr` with the read/parse cause | `TestBaseFreshnessConfigLoadFailureIsUncertainty` (3 cases) |

Mutant probes: each fix reverted alone, the matching test went red, the fix was restored, `cmp` against the saved copy identical. Complete's refusal line goes to stdout (the verb's existing writer); acquire's goes to stderr.

Delta re-audit P2 (non-disclosure): the `parse workflow.yaml` cause quoted the YAML parser error, and the loader's tier-validation error was passed through, so a secret-looking value reached the notice (reproduced for malformed YAML and a rejected tier token; wrong-type and read-failure cases did not leak). RED `d2f33e61d`, GREEN `8025a3253`: causes are now fixed text (`invalid YAML syntax`, `a value has the wrong type`, `workflow.yaml was rejected by validation`; the read cause names only the path). Test: `TestBaseFreshnessConfigFailureNoticeDoesNotEchoConfigContent`; both fixes mutant-probed red, restored byte-identical.

Third sync audit (2 x P2, same class: untrusted text in gate output) - swept as a class. RED `b3e5df8ee`, GREEN `81054c29c`. Sites now bounded: git failures (4 operations + the unresolvable-branch cause) render as `git <op> failed (exit N)` / `could not run`; record validation renders from a fixed phrase set (`no re-measure record is stored for it`, `the stored re-measure record is unreadable`, `the record is incomplete or malformed`, `its command exited N`, `its structured test report is missing or empty`); the non-numeric rev-list count and the Getwd failure carry no text; the config causes (E.2.7 above) are fixed phrases; SHAs printed are validated full SHAs, branch and paths are operator- or tool-constructed. Test: `TestBaseFreshnessOutputNeverEchoesUntrustedInput` (8 sources x gate off/on). Outside this card, observed: the settings-drift UNDETERMINED line printed by acquire forwards git stderr verbatim; not changed here.

## §E.3 Run-phase Audit-Ready Signal

- run_complete_at: 2026-10-17T16:01:36Z
- run_commit_sha: pending-backfill (the evidence commit that carries this block cannot cite its own hash; code state under test is `b38974aea245f1f3c3b5972360b0507bd7268f00`)
- run_status: all 16 acceptance criteria PASS with observed output at the code state above; heavy repository-wide gates pending (see E.2.6)
- ac_pass_count: 16
- ac_fail_count: 0
- preserve_list_post_run_count: 0 (no neighbour-owned path changed: G-15b printed nothing)
- l44_pre_commit_fetch: not performed (no fetch was run in this card tree; no push was requested)
- l44_post_push_fetch: not applicable (no push)
- new_warnings_or_lints_introduced: 0 (golangci-lint on internal/factory, internal/config, internal/cli printed 3 errcheck findings, all in internal/cli/factory_run_adopt.go, a file this card did not change)
- cross_platform_build.linux_darwin_scoped: go build ./internal/cli/ and go vet on the three changed packages exit 0
- cross_platform_build.windows: not run (heavy lease go-test-heavy held by another session, `moai slot acquire` exit 3)
- total_run_phase_files: 15 (non-SPEC, non-report paths against the plan pin)
- m1_to_mN_commit_strategy: per-milestone RED commit(s) then GREEN commit, committed directly on the card branch, not pushed

audit_ready: true

## §E.4 Sync-phase Audit-Ready Signal

- sync_complete_at: 2026-10-18T00:18:54Z
- sync_commit_sha: 1dfeadd805adf207b06d0f6b681f8a0a8557c25d
- sync_commit_sha_note: backfilled in the commit after the sync commit (a commit cannot cite its own hash); the sync commit carried the pending-backfill placeholder; run HEAD at sync start was `d45a13131`
- sync_status: audit-ready; documentation synced, SPEC status moved in-progress -> implemented -> completed on the single sync commit; independent sync audit and repository-wide build/test verdicts pending
- b12_self_test_a: `grep -c 'SPEC-INTEGRATION-STALE-BASE-001' CHANGELOG.md` printed `0` before emission (exit 1, no prior entry), and `1` after
- b12_self_test_b: ac_source=`.moai/specs/SPEC-INTEGRATION-STALE-BASE-001/acceptance.md`, tier=M (resolved, non-empty); the live-identifier counter printed `16` (stderr `live=16 excluded=0 ambiguous=0`, exit 0); the CHANGELOG entry states `AC-ISB-001..016` (16 criteria). The 16 requirement lines `- REQ-ISB-NNN` in spec.md match the entry's `REQ-ISB-001..016`
- b12_self_test_c: every path named in the CHANGELOG entry was checked with `ls` (all 7 implementation files, the two rule/config files and the SPEC progress record exist); the 9 test names it cites were each found by `grep -rl '^func <name>('` in the file the entry names
- changelog_entry_position: first bullet under `## [Unreleased]` / `### Added` in CHANGELOG.md
- frontmatter_status_transitions.spec_md: `status: in-progress` -> `implemented` -> `completed` applied atomically on this sync commit (the schema's merged close); `updated:` already `2026-10-18` (the sync date), unchanged. A first pass that stopped at `implemented` was measured by `moai spec audit` (MCP `spec_audit`, filter SPEC-INTEGRATION-STALE-BASE-001) as `SyncStatusDrift` MUST-FIX ("§E.2 + §E.4 + sync_commit_sha present (sync complete) but status != completed"), which is why the terminal value is written here
- frontmatter_status_transitions.plan_md_acceptance_md: not applicable (neither file carries a frontmatter block, so there is no `status:` or `updated:` to refresh)
- frontmatter_status_transitions.progress_md: not applicable (no frontmatter; body sections only)
- docs_scope: CHANGELOG.md only. The rule guidance (`factory-dispatch-gates.md` plus its template mirror) and the `workflow.base_freshness_gate` config block landed in run milestones M1 and M4; plan.md §G9 records that no docs-site or README page for `workflow.*` keys was planned, and none was added
- canary_compliance_check: not applicable (this SPEC defines no forward-looking policy that its own sync tests)

audit_ready: true

## §F Phase 4 Mode Selection

Decision: serial. Tier M, 15 files, single-domain Go + template + rule guidance, coding-heavy; fanout/sweep not applicable. Kickoff gate met in autonomous form (plan-audit iter4 PASS 0.83, hash bound, no open blocker).

## §G Override and Refusal Record

- 2026-10-17T15:22:13Z SPEC-INTEGRATION-STALE-BASE-001 ceiling-refusal outcome=hold reasons="plan-audit ceiling reached (round count 4 >= tier ceiling 2); the verdict matches no admitting arm and holds, entry blocked (REQ-ACE-006) — release path: the split/new-SPEC route of REQ-ACE-005 or an operator decision recorded in progress.md §G" evidence=/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1834/.moai/reports/t1834/plan-audit-iter1.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1834/.moai/reports/t1834/plan-audit-iter2.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1834/.moai/reports/t1834/plan-audit-iter3.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1834/.moai/reports/t1834/plan-audit-iter4.md
- 2026-10-17T15:24:06Z SPEC-INTEGRATION-STALE-BASE-001 ceiling-outcome outcome=pass-through reasons="plan-audit ceiling reached (round count 4 >= tier ceiling 2 + 1 delta rounds); the verdict is admission-clean and admits without a question (REQ-ACE-013)" evidence=/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1834/.moai/reports/t1834/plan-audit-iter1.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1834/.moai/reports/t1834/plan-audit-iter2.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1834/.moai/reports/t1834/plan-audit-iter3.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1834/.moai/reports/t1834/plan-audit-iter4.md
- 2026-10-17T15:24:12Z SPEC-INTEGRATION-STALE-BASE-001 ceiling-outcome outcome=pass-through reasons="plan-audit ceiling reached (round count 4 >= tier ceiling 2 + 1 delta rounds); the verdict is admission-clean and admits without a question (REQ-ACE-013)" evidence=/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1834/.moai/reports/t1834/plan-audit-iter1.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1834/.moai/reports/t1834/plan-audit-iter2.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1834/.moai/reports/t1834/plan-audit-iter3.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1834/.moai/reports/t1834/plan-audit-iter4.md
