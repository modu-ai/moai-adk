# Role-core budget fix — TestRoleInjectionAssemblyBudget (lane-17)

Tree: `af-contract-spec` (branch `WT-agentic-factory-contract`), integration `10ae2453c5` merged as `eab13bfd30`,
plus the uncommitted edits named below. Judging build: `go test` compiled from this tree (no installed `moai` involved).

## Claim

1. `TestRoleInjectionAssemblyBudget` passes for all four subtests (template/deployed x factory-lane/factory-leader):
   role core 4664 -> 4060 UTF-16 units (ceiling 4092; design target 3800 NOT reached).
2. No obligation was deleted: every role-core sentence was shortened in place; the long wording stays in the
   non-core body paragraphs of `factory-dispatch.md` (e.g. the full disk-wait paragraph) and in `auto-semantics.md`.
3. Template and deployed copies are byte-identical (`cp` from the template, Template-First).
4. `TestCompositeBudgetRealAssembly/lane_en_glm` is STILL RED (11056 > 10000) and is not caused by the role core
   (see Gaps / Residual-risk).

## Evidence (verbatim, this run)

Before (at `eab13bfd30`, core 4664):
`total=9716/9718 ... core=4664 (ceiling 4092, target 3800)` -> FAIL, all four subtests.

After:
```
$ go test ./internal/hook -run '^TestRoleInjectionAssemblyBudget$' -count=1 -v
    role_injection_budget_test.go:163: template tree: role factory-lane composite=9112 (budget 9146) core=4060 (ceiling 4092, target 3800)
    role_injection_budget_test.go:163: template tree: role factory-leader composite=9114 (budget 9146) core=4060 (ceiling 4092, target 3800)
    role_injection_budget_test.go:163: deployed tree: role factory-lane composite=9112 (budget 9146) core=4060 (ceiling 4092, target 3800)
    role_injection_budget_test.go:163: deployed tree: role factory-leader composite=9114 (budget 9146) core=4060 (ceiling 4092, target 3800)
--- PASS: TestRoleInjectionAssemblyBudget (0.00s)
ok  	github.com/modu-ai/moai-adk/internal/hook	0.786s
```
Anchored run, also PASS in the same invocation: `TestInjectedCoreCarriesObligationElements`,
`TestRoleCoreCarriesNoQueueStart`, `TestLaneRuleOmitsCronOnCodexLane`.

Template package, anchored: `TestRoleCoreStubGate`, `TestFactoryDispatchBodyCarriesRenewIntervalLinkage`,
`TestAutoSemanticsArmFollowsRenewInterval`, `TestBindingLedgerIntegrity`, `TestLaneRecheckDoctrine` -> `ok  ... 1.449s`.

## Changes made

- `factory-dispatch.md` (template, then deployed): 20 of 24 role-core blocks shortened in two passes (wake rule, hold rule,
  relations, cross-check x3, card-review, evidence read, git-flow merge, effort, delegation channel, language,
  completion reports, worktree transition, launcher, branches, path form, verification x3). Strings pinned by
  `role_core_elements_test.go` kept verbatim.
- `factory-dispatch-core.md` (template + deployed): the wake sentence shortened (stub was 10052 > 10000).
- `internal/template/testdata/binding_ledger.json`: `after_text` of 15 rows (ALB-0260, 0262, 0263, 0271, 0275,
  0276, 0278, 0280, 0287, 0290, 0297, 0298, 0299, 0300, 0303) re-pointed at the shortened role-core sentences.
- Three template tests re-pinned from the retired cron wording to the self-paced wake the user chose
  (`renew_interval_body_test.go`, `auto_semantics_renew_test.go`, `lane_recheck_doctrine_test.go`).

## Baseline-attribution

Every output above was produced in this run against this tree. Before-state numbers are from the first anchored run
after the merge. The `make build` (exit 0) ran from this tree; `catalog.yaml` showed no tracked change.

## Gaps

- Design target 3800 not reached (4060); only the ceiling 4092 is satisfied, margin 32 units.
- The full `internal/template` package ran once (545s) BEFORE the last four fixes; it was not re-run in full after them.
  The five failing tests were re-run anchored (PASS); other packages (`internal/cli` etc.) were not run.
- `TestCompositeBudgetRealAssembly/lane_en_glm` fails: composite 11056 vs cap 10000 (margin -1056);
  lane_ko_glm 9759 (+241), leader_en 9936 (+64), leader_ko 8541 (+1459) pass.
- Not measured: the same composite at main `cb30e05b3e` (no second checkout was made).

## Residual-risk

- Cause of the lane_en_glm overshoot (read from `git diff b793b2356a HEAD`): the integration merged the Go rewrite of
  `laneRecheckRule` (self-paced wake, `session_start_factory_i18n.go`), whose English text is about 1,650 units against
  about 830 for Korean. That text is outside the role core, so no further role-core cut fits (a cut of ~1,050 would
  delete obligations). Fix needs a decision: shorten the English `laneRecheckRule` (it carries the `/moai goal` and
  decision-order sentences) or raise the cap. Owner of that Go text is not this lane's scope.
- The `maxFireGap` subtests of `TestAutoSemanticsArmFollowsRenewInterval` still test cron-arithmetic for a design the
  user retired; they pass vacuously now and should be removed by whoever owns that SPEC.
