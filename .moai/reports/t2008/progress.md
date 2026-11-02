# t2008 progress record (lane-7, Class B, Run contract reconciliation)

card: t2008 | branch: WT-run-tdd-result | tree: .moai/worktrees/t2008 | run: tnsei4 | stop: local commits only (no push, no PR)

decision record: decided_by=lane-7 (claude, lane orchestrator) evidence_refs=leader dispatch + follow-up (RUN-03 -> RUN-01 -> RUN-02),decisions.md R22/R23,adk-workflow-pages-handoff.md appendix 3 (read only, not copied) ladder_path=disk-evidence (auto-semantics.md §6 step 1)

## Baseline (tree HEAD 5c3f0ebf27584d2c45deb2d00f9d69a4eea932c4 = origin/main at intake; `git merge --ff-only origin/main` from cb30e05b3e)

The handoff items are a static survey of 5c3f0ebf27; each is a hypothesis until reproduced here.

### RUN-03 - TDD result classes not connected to the run workflow (reproduced before any edit)

- `bash .claude/hooks/tests/test-tdd-result-contract.sh` -> exit 1, no output (`set -e`; first failing grep ends the script).
- The four class names and the three rule literals pass: `grep -c -F "intended assertion" .claude/rules/moai/workflow/tdd-result-contract.md` -> 3, `"verbatim output"` -> 1, `"never an"` -> 1.
- The three literals the test requires in `.claude/skills/moai/workflows/run/task-decomposition.md` are absent: `grep -c -F "semantic result"` -> 0, `"Compile errors"` -> 0, `"regression set"` -> 0.
- Carriers of the class names across `.claude .agents .codex internal/template/templates`: `grep -rlE "REGRESSION_FAILURE|EXPECTED_RED|TOOL_FAILURE|tdd-result-contract"` -> only `.claude/hooks/tests/test-tdd-result-contract.sh`, `.claude/rules/moai/workflow/tdd-result-contract.md`, and its template mirror. No agent, workflow, or skill names the contract.
- `run.md:158-161` and `run/phase-execution.md:481-482` order "clear the goal and escalate" on any test assertion failure; no clause separates EXPECTED_RED in the RED step.
- Nothing runs the test: `grep -rn "test-tdd-result-contract" .github Makefile internal .claude/hooks` -> no output. No template copy exists (`ls internal/template/templates/.claude/hooks/tests/` -> no such directory).
- Verdict on the hypothesis: confirmed (the red is for the stated reason - the TDD section never carries the contract - not an unrelated file).
- Mirrors before the edit: `cmp` live vs `internal/template/templates/` for run.md, run/phase-execution.md, run/task-decomposition.md -> identical (exit 0 each).

Gaps: the survey's claim that this misleads a real run is not exercised (no workflow run was performed); the test runs only against the live tree.

### RUN-03 repair (commit after 1624cd7e6c)

- Test first: `.claude/hooks/tests/test-tdd-result-contract.sh` gains 9 lines requiring `tdd-result-contract.md` in task-decomposition.md, and `tdd-result-contract.md`, `EXPECTED_RED`, `REGRESSION_FAILURE` in run.md and run/phase-execution.md. The existing literal checks are untouched, so the test is stricter, not weaker.
- Edits (each mirrored byte-identical to `internal/template/templates/`): task-decomposition.md Phase 12 gets one requirement bullet naming the contract with the three literals; run.md ac_converge prompt and run/phase-execution.md semantic-failure bullet each state that `EXPECTED_RED` is not a semantic failure and `REGRESSION_FAILURE` is.
- `bash .claude/hooks/tests/test-tdd-result-contract.sh` -> `PASS: TDD distinguishes expected RED from regression and tool failures` (was exit 1 at baseline).
- Mutation (scratch copy under the session scratchpad, only run.md reverted to `git show 5c3f0ebf27:.claude/skills/moai/workflows/run.md`): the strengthened test -> exit 1. The new run.md check bites.
- Catalog hash: after the skill edit `go test -count=1 -run 'Mirror|Embed|Manifest|DateProvenance|BindingLedger' ./internal/template/` -> `TestManifestHashFormat` FAIL (`CATALOG_HASH_UNSTABLE`, stored 18df0ebd... vs computed f3f8912d...). `go run ./internal/template/scripts/gen-catalog-hashes.go --all` (the Makefile `build` step) changed one line of `internal/template/catalog.yaml`. Same command rerun -> `ok github.com/modu-ai/moai-adk/internal/template 3.290s`.
- Customer templates: only files that already have a mirror changed, in the same commit; no `moai update` forced.

### RUN-01 - plan-gate teaching conflicts (reproduced before any edit; static conflict, no runtime bypass found)

- Instruction side: `run/phase-execution.md` 4b records `audit_verdict: FAIL_WARNED` and proceeds to Phase 5 on a FAIL inside a grace window; 4c offers "Override and proceed (records BYPASSED)"; 4d offers "Proceed with acknowledgement" on INCONCLUSIVE; `### When --skip-audit Flag Is Provided` teaches `MOAI_SKIP_PLAN_AUDIT=1` and a non-interactive `bypass_reason: "non-interactive"` entry; `## HUMAN GATE: Plan Approval` asks every run for a user approval. 4a names only PASS, so the PASS-WITH-DEBT admission has no routing row there.
- Authority side: `run.md` section 1 makes FAIL / INCONCLUSIVE fail closed and PASS the autonomous entry evidence; `.claude/rules/moai/workflow/auto-semantics.md` (the gate table: "plan-audit bypass flags | RETIRED into the default path"; section 9.1: "`PASS` or `PASS-WITH-DEBT` ... FAIL, INCONCLUSIVE, BYPASSED, and an absent verdict stay hard blocks"; a PASS-WITH-DEBT Kickoff copies every debt into progress.md under `Binding run conditions`). `plan-auditor.md` projects FAIL_WARNED to `fail`.
- Reachability: `grep -rn "MOAI_SKIP_PLAN_AUDIT\|skip-audit" internal cmd --include='*.go'` -> only `internal/runtime/audit_gate.go` (library) and tests; `grep -rn "GateConfig{" internal cmd --include='*.go'` -> `runtime.GateConfig` is constructed only in `_test.go` files (no production caller sets `SkipAudit` or `T0`); `.moai/state/audit-gate-merge-at.txt` does not exist in this tree and no Go code reads it (grace is only reachable through `MOAI_AUDIT_GATE_T0` or a test). The production Kickoff admission (`internal/runtime/audit_gates.go`) routes through `auditverdict.Admit` and mentions no bypass. So the hypothesis "FAIL_WARNED lets a FAIL pass" is refuted as a runtime defect and confirmed only as instruction text that an orchestrator following the skill body could act on.
- RED test (new, `internal/template/run_plan_gate_precedence_test.go`): `go test -count=1 -run TestRunPlanGatePrecedence ./internal/template/` -> exit 1, `TestRunPlanGatePrecedence_Deployed` FAIL with 11 missing statements (Step 4 precedence: `Precedence`, `auto-semantics.md`, `§9.1`, `hard block`; Step 4a: `PASS-WITH-DEBT`, `Binding run conditions`; skip-audit section: `Legacy`, `retired`, `§9.1`; Plan Approval gate: `keep-set`, `§9.1`). `TestRunPlanGatePrecedence_Mutants` passes (checker can fail: a bare body fails all groups, the positive control passes, and dropping any one of five required words is detected).
- Scope kept small: no section is deleted (the existing `--skip-audit` marker test still requires the keyword); the repair adds precedence statements and the PASS-WITH-DEBT row, it does not change Go code or the rule files.

### RUN-01 repair (commit after 2669aae81f)

- Edits to `run/phase-execution.md` (+ byte-identical template mirror, 6+/1- each): a `[HARD] Precedence` paragraph under Step 4 (Kickoff gate decides entry; FAIL, INCONCLUSIVE, BYPASSED, absent verdict are a hard block; 4b/4c/4d are legacy interactive forms that never turn FAIL or INCONCLUSIVE into entry; `FAIL_WARNED` is a historical label, never a pass); a PASS-WITH-DEBT row in 4a (enumerated debt copied to progress.md under `Binding run conditions`, matching auto-semantics §9.1, so the admitted-debt flow is preserved, not tightened to PASS-only); a Legacy/retired note above the `--skip-audit` section (the section and its keyword stay; `skills_audit_test.go` still requires it); an operator-form/keep-set note at the Plan Approval gate.
- `go test -count=1 -run 'TestRunPlanGatePrecedence|TestSkillsContainPlanAuditGateMarkers|Mirror|Embed|Manifest|DateProvenance|BindingLedger' ./internal/template/` -> `ok github.com/modu-ai/moai-adk/internal/template 4.880s` (the RED test now passes; the existing `--skip-audit` marker test still passes). Catalog hash regenerated with the same script as RUN-03 (1 line).
- Regression: `bash .claude/hooks/tests/test-tdd-result-contract.sh` -> PASS.
- Not changed on purpose: Go code, `spec-workflow.md` Grace Window paragraph and the other rule files (always-loaded budget and binding-ledger surfaces; recorded as residual risk), and `run.md` section 1 (already fail-closed).

Gaps: no end-to-end run entry was executed (the survey's FAIL / INCONCLUSIVE / env-bypass / future-T0 matrix is not run; only the instruction text and the production-caller absence were checked); the precedence test is literal-based plus mutants.
Residual risk: `spec-workflow.md` still says a FAIL emits warnings only during the grace window, so a reader of that rule alone can still take the legacy reading; correcting it needs a rule edit with its own byte statement and ledger check.

### RUN-02 - plan hash and cache source described differently from the Go helper (reproduced before any edit)

- Instruction side: `run/phase-execution.md` Step 1 says `ComputeHash` hashes "the whitespace-normalized UTF-8 content"; Step 2 says to read `.moai/reports/plan-audit/<SPEC-ID>-<YYYY-MM-DD>.md` "(today's date)" for the cache lookup.
- Code side (read in this tree): `internal/runtime/audit_cache.go` ComputeHash writes `name:len\x00` + the exact bytes + `\x00` per artifact over `acceptance.md, decision-index.md, design.md, plan.md, research.md, spec.md, tasks.md` (its own comment: "formatting normalization is intentionally not part of cache identity"); `internal/runtime/audit_review.go` FileAuditCache.Lookup resolves the latest plan-phase review through `ResolveLatestPlanAudit` and states the date-stamped run-gate report "is intentionally not consulted; it is an append-only history surface, not a cache key or verdict source".
- RED test (new, `internal/template/run_plan_cache_description_test.go`): `go test -count=1 -run TestRunPlanCacheDescription ./internal/template/` -> exit 1, `TestRunPlanCacheDescription_Deployed` FAIL with 7 findings (Step 1: missing `exact bytes`, still says `whitespace-normalized`; Step 2: missing `ResolveLatestPlanAudit`, `review-`, `history`, `gate receipt`, still says `(today's date)`). `ComputeHash` is already named, so the red is for the stated reason. `TestRunPlanCacheDescription_Mutants` passes (stale body fails every group, positive control passes, keeping the stale phrase beside the new words is rejected, each required word is load-bearing).
- Plan: fix the two steps in the skill body (+ mirror) with one table of the three surfaces (daily report = history, latest review-N = verdict and cache source, gate receipt = separate); correct the stale "daily report file serves as the durable cache source" comment on `InMemoryCache` (comment only; no function body changes).

### RUN-02 repair (commit after 5868b4eb40)

- `run/phase-execution.md` (+ byte-identical mirror): Step 1 now says `ComputeHash` hashes the exact bytes framed by file name, length and NUL, with no whitespace normalization, and tells the reader not to hand-compute the hash; Step 2 now reads the highest-numbered `.moai/reports/plan-audit/<SPEC-ID>-review-<N>.md` as `ResolveLatestPlanAudit` selects it, with a three-row table (date-stamped report = history only; highest review-N = verdict and cache source; gate receipt = separate admission evidence) and the rule that a later FAIL review ends the hit.
- Self-correction during the repair (not a published claim): my first draft of Step 2 named `plan-audit.md` under a card directory; reading `ResolveLatestPlanAudit` (`prefix := specID + "-review-"`, largest integer N, in `reportDir`) showed the file is `<SPEC-ID>-review-<N>.md` in `.moai/reports/plan-audit/`, and the text was corrected before the test ran. The card-directory `plan-audit.md` stream in `spec-workflow-detail.md` is a different surface and is not touched.
- `internal/runtime/audit_cache.go`: one stale comment on `InMemoryCache` (it said the daily report file serves as the durable cache source) now says `FileAuditCache` reads the highest-numbered review-N and the daily report is history. Comment only; no function body changed.
- `go test -count=1 -run 'TestRunPlan|TestSkillsContainPlanAuditGateMarkers|Mirror|Embed|Manifest|DateProvenance|BindingLedger' ./internal/template/` -> `ok github.com/modu-ai/moai-adk/internal/template 4.195s` (both new tests pass); `go test -count=1 -run 'Cache|Hash|Review' ./internal/runtime/` -> `ok github.com/modu-ai/moai-adk/internal/runtime 0.381s`; `go vet ./internal/runtime/` -> no output, exit 0. Catalog hash regenerated (same script).

Gaps: the survey's acceptance matrix (whitespace-only edit, decision-index-only edit, next-day reuse, higher-numbered FAIL, duplicate or corrupt hash, same-name other repo) is not run against an instruction consumer; the instruction and the Go helper are only aligned in wording. `audit_gate.go` still carries legacy 24h/daily wording in comments (not touched). Pre-existing lint notes (`persist.go`, `audit_gate_test.go` unusedparams) are outside this change.

### RUN-05 - final verdict owner written for both leader and lane (reproduced; NOT repaired - needs a policy ruling)

- Same-rule contradiction, read in this tree: `factory-dispatch.md` (full body) and `factory-dispatch-core.md` say "**The final PASS/FAIL verdict is the leader's**, read from the evidence on disk and never delegated to the lane that produced the work" and the deputy clause keeps "final PASS/FAIL verdicts" with the leader. `factory-dispatch-detail.md` § The verdict's home says the lane-autonomy revision "moved the per-card completion verdict to the lane ... the leader's retained duties ... no longer include per-card PASS/FAIL"; `factory-dispatch-mechanics.md` says "The lane's verdict is its own" and its leader-retained table has no per-card completion verdict. `sync.md` makes the four-dimension verdict binding for the sync-phase quality decision on the happy path.
- Why no edit: the sentence sits in the leader's own always-loaded stub and role-injected core (byte budget, binding ledger, leader-held authority). Which reading wins is the leader's ruling; the survey also asks for decision kinds with separate names, which the dispatch put out of scope (full RACI). A one-sentence reconciliation is proposed to the leader instead: name the kinds - worker implementation report, independent auditor verdict, runtime admission, lane per-card delivery verdict (evidence: independent receipt, CI, readiness triple), and the leader's cross-lane / keep-set / batch merge approval.

### RUN-06 - Resume Check skips completed phases without re-verifying identity (reproduced before any edit)

- `grep -n -i "identity\|HEAD\|plan_artifact_hash\|invalidate\|lease" .claude/skills/moai/workflows/run/context-loading.md` -> no output. The `### Resume Check` section says to identify the last completed phase and "skip all completed phases". `manager-develop.md` "Checkpoint and resume" names only `--resume latest`. Existing runtime data that could anchor a check: `internal/factory/factory_runtime.go` `factoryProvenance` (SpecID, SpecPath, SpecSHA256, GitCommit, CapturedAt).
- The survey's wording is kept: this is a missing link, not a proven un-verified resume (other gates may cover it).
- RED test (new, `internal/template/run_resume_identity_test.go`): `go test -count=1 -run TestRunResumeIdentity ./internal/template/` -> exit 1, `TestRunResumeIdentity_Deployed` FAIL, five findings (missing `repository root`, `HEAD`, `plan_artifact_hash`, `card`, `invalidate`); `TestRunResumeIdentity_Mutants` passes.

### RUN-06 repair (commit after 0c498dc3b3)

- `run/context-loading.md` Resume Check (+ byte-identical mirror, catalog hash regenerated): before skipping a completed phase the orchestrator re-verifies repository root and worktree path, card id and lease owner, HEAD (equal or descended with only expected commits), a clean working tree beyond recorded changes, and the recomputed `plan_artifact_hash`; it skips only phases whose identity matches, invalidates a differing phase and every later phase that consumed it, never resumes a cancelled or revoked run from its checkpoint, and logs the differing field. The no-progress.md branch is unchanged.
- `go test -count=1 -run 'TestRunPlan|TestRunResumeIdentity|TestSkillsContainPlanAuditGateMarkers|Mirror|Embed|Manifest|DateProvenance|BindingLedger' ./internal/template/` -> `ok github.com/modu-ai/moai-adk/internal/template 3.324s`; `bash .claude/hooks/tests/test-tdd-result-contract.sh` -> PASS.
- Gaps: no resume was executed; the survey's matrix (other repo/card/head, expired reassignment, cancel then retry, crash between write and checkpoint, duplicate complete, stale owner return) is not run.

### Leader rulings received (cross-session message after the first report; recorded as input, not as an approval of any push)

decision record: decided_by=lane-7 (claude, lane orchestrator) evidence_refs=leader message (RUN-05 ruling, deferrals, RUN-07/10/13 rule, RUN-03 gap wording) ladder_path=gate boundary (leader ruling, auto-semantics.md §6)

- RUN-05: no rule sentence is changed in this card. Working basis until the 3.2 approval ledger (R18) design: a lane's verdict is a self-verification recommendation; progress and landing are the leader's reading of the evidence. The decision-kind split proposed above is preserved here as input to that design.
- RUN-08, RUN-12 and the `spec-workflow.md` Grace Window paragraph: deferred, not repaired.
- RUN-07, RUN-10, RUN-13: a small RED-first repair only when current code reproduces the defect; otherwise a refutation is recorded.
- RUN-03 Gaps (leader's wording): the test pins the presence of wording only ("문구 존재만 고정"): `test-tdd-result-contract.sh` plus the added run.md / phase-execution.md greps prove that sentences exist, not that an orchestrator or `manager-develop` classifies a RED correctly; no behavioral fixture (normal RED->GREEN->regression, compile failure, timeout, other-test failure, re-failure after GREEN, zero tests) was built.

### RUN-13 - optional gate read as required (checked against current code and config: not reproduced as a defect here)

- Code read: `internal/factory/candidate_landing_check.go` `candidateCIRequired` returns false for an absent key AND for an unreadable config, and says so in its comment ("the step never invents a gate the project did not ask for"); `CandidateGrantGate` returns nil unless that flag is true.
- Config measured: `grep -n -A3 candidate_ci:` on `.moai/config/sections/workflow.yaml` and on `internal/template/templates/.moai/config/sections/workflow.yaml` -> `enabled: false` in both; `base_freshness_gate` in the live config -> `enabled: false`, `max_behind_landings: 0` (the file's own comment: "Ships OFF"; any config/branch/git uncertainty admits with one notice line).
- Verdict: the survey's condition ("if AF policy requires the check as required") is not met by the shipped or live configuration, so a corrupt config cannot downgrade a required gate here; the unreadable-equals-off behavior is documented design. Not repaired; no RED test could be made red on the current tree. Residual: if a project later turns `candidate_ci.enabled` on and relies on it, an unreadable config would silently read as off at the merge step; that needs a policy decision (fail-closed when the flag was once on), not a wording fix.
- Gap: `base_freshness.go` and the `integration_window_ops.go` grant path were read only at the cited lines (the grant gate delegates to `candidateCIRequired`); the full complete->merge seam was not exercised.

### RUN-07 - ac_converge evidence admission (not repaired; no current-code reproduction attempted)

- The survey item is a design change (separate the loop stop from the completion/merge gate and admit evidence by key). `run.md` already states the predicates are transcript-judged, and `manager-develop.md` already leaves the repository-wide verdict PENDING for CI. No behavior of current code was shown wrong, so nothing is claimed either way; left for the 3.2 approval-ledger design.

### RUN-10 - stale phase labels (reproduced before any edit)

- `run.md:51` router row reads "Phase 11~4: Implementation ... quality validation (Phase 13/2.8) ... git operations (Phase 19) ... (Phase 20)"; `run/task-decomposition.md:252` (Phase 18) ends with "before proceeding to Phase 3", while that file's headings are `## Phase 11:` to `## Phase 20:` (Phase 19 follows Phase 18). `run.md` metadata also still reads `version: "2.6.0"`, `updated: "2026-02-23"` (not touched; a metadata bump is not a behavior fix).
- RED test (new, `internal/template/run_phase_reference_test.go`): `go test -count=1 -run TestRunPhaseReferences ./internal/template/` -> exit 1, `TestRunPhaseReferences_Deployed` FAIL with three findings (`'proceeding to Phase 3' names no phase heading`; router `Phase 11~4`; router `2.8`); `TestRunPhaseReferences_Mutants` passes.
- Plan: change only those three strings (+ template mirrors, catalog hash).
- Repair (commit after 5bd50fabf0): `run.md` router row now reads "Phase 11-20: Implementation ... DDD/TDD cycles (Phase 11/12), quality validation (Phase 13), git operations (Phase 19), completion guidance (Phase 20)"; `task-decomposition.md` Phase 18 hand-off now says "before proceeding to Phase 19". Mirrors copied, catalog hash regenerated. `go test -count=1 -run 'TestRunPlan|TestRunResumeIdentity|TestRunPhaseReferences|TestSkillsContainPlanAuditGateMarkers|Mirror|Embed|Manifest|DateProvenance|BindingLedger' ./internal/template/` -> `ok github.com/modu-ai/moai-adk/internal/template 6.576s`; hook test PASS.
- Not changed: `run.md` metadata `version: "2.6.0"` / `updated: "2026-02-23"` and the `mode-orchestration.md` "Version: 2.11.0" tail (metadata bumps are not a behavior fix; Gap).

### Pre-push record (leader instruction: impact tests, lint record, one push, draft PR; lease recovered)

decision record: decided_by=lane-7 (claude, lane orchestrator) evidence_refs=leader message (push once without force, draft PR, record tests and lint),this section ladder_path=gate boundary (leader ruling, auto-semantics.md §6)

- Lease: `moai factory status --run tnsei4` showed t2008 `leased` with the lease expired (11:31:21Z). Leader-approved recovery, official verbs, one call each: stage in the tree -> "card lease expired" (expected); status -> `t2008 state=assigned owner=lane-7 lease=none version=5`; `moai factory next --card t2008 --run tnsei4` from the parent checkout -> `t2008 stage=- worktree=t2008`; stage again in the tree -> `t2008 run v7 lease renewed`. Same owner, no takeover. (The `linked #1957` shown by `next` is t2007's PR `WT-plan-plan-md`, not this card; the branch `WT-run-tdd-result` did not exist on origin before this push.)
- Base: origin/main still `5c3f0ebf27` (the tree base), so no absorb was needed.
- Impact tests at this head (before this record commit): `go test -count=1 -run 'TestRunPlan|TestRunResumeIdentity|TestRunPhaseReferences|TestSkillsContainPlanAuditGateMarkers|Mirror|Embed|Manifest|DateProvenance|BindingLedger' ./internal/template/` -> `ok ... 7.466s`; `go test -count=1 -run 'Cache|Hash|Review' ./internal/runtime/` -> `ok ... 0.449s`; `bash .claude/hooks/tests/test-tdd-result-contract.sh` -> PASS.
- Static checks: `gofmt -l` over the five changed Go files -> no output; `go vet ./internal/template/ ./internal/runtime/` -> no output, exit 0.
- Lint: `golangci-lint run --timeout=5m ./internal/template/ ./internal/runtime/` -> exit 3, "parallel golangci-lint is running" (another session holds the lint slot; not retried, not forced). Gap: golangci-lint was NOT run on this head; CI's Lint job is the lint measurement. Earlier editor diagnostics for the new test files (a `stringscut` hint) were fixed before the RED commit; other diagnostics named untouched files (`always_loaded_budget_test.go`, `navigator_audit_test.go`, `persist.go`, `audit_gate_test.go`).

### CI finding on PR #1961 (head a8d9535c96, run 41333015840): own defect found by CI, reproduced before the fix

- Terminal CI on the draft PR: three failing jobs - Test (ubuntu-latest) 124927038831, Race Test (internal/cli 1 of 4) 124927038850, Race Test (remainder packages) 124927038949. Log `--- FAIL` lines: `TestCleanReportsArchive_MovesOnlyEligibleCandidates` (jobs 831, 850; the #1956 calendar defect, not this PR) and `TestEntryRouterLOCCeiling` (jobs 831, 949): `workflow_split_test.go:154: ENTRY_ROUTER_LOC_VIOLATION: run.md has 202 LOC (ceiling: 200)`.
- Cause: the RUN-03 edit to `.claude/skills/moai/workflows/run.md` (ac_converge prompt, +3/-1 lines) and its template mirror took the router from 200 to 202 lines. My impact-test selection covered only `./internal/template/` and `./internal/runtime/`; `./internal/skills/` holds the router LOC ceiling and was not run. That is a gap in my scoping, not a flaky test.
- Local reproduction before the fix: `git show 5c3f0ebf27:.claude/skills/moai/workflows/run.md | wc -l` -> 200; `wc -l` of the live and the template run.md -> 202 each; `go test -count=1 -run TestEntryRouterLOCCeiling ./internal/skills/` -> exit 1, `ENTRY_ROUTER_LOC_VIOLATION: run.md has 202 LOC (ceiling: 200)`.
- Plan: keep the EXPECTED_RED / REGRESSION_FAILURE qualifier (the RUN-03 hook test requires `tdd-result-contract.md`, `EXPECTED_RED`, `REGRESSION_FAILURE` in run.md) and reflow the ac_converge prompt so run.md is at most 200 lines; mirror, regenerate the catalog hash, rerun the ceiling test and the whole `./internal/skills/` package.

### Repair of the run.md LOC ceiling failure (commit after fb98b58453)

- `run.md` ac_converge prompt reflowed from 5 lines to 2 (the EXPECTED_RED / REGRESSION_FAILURE qualifier and `tdd-result-contract.md` kept verbatim), template mirror copied, catalog hash regenerated. `wc -l .claude/skills/moai/workflows/run.md` -> 199 (base 200, CI-failing 202).
- `go test -count=1 ./internal/skills/` -> `ok github.com/modu-ai/moai-adk/internal/skills 0.210s` (the ceiling test passes; the whole package ran); `go test -count=1 -run 'TestRunPlan|TestRunResumeIdentity|TestRunPhaseReferences|TestSkillsContainPlanAuditGateMarkers|Mirror|Embed|Manifest|DateProvenance|BindingLedger' ./internal/template/` -> `ok ... 4.756s`; `bash .claude/hooks/tests/test-tdd-result-contract.sh` -> PASS.
- Scoping rule learned: a skill or router file under `.claude/skills/` is also measured by `./internal/skills/` (LOC ceilings, split rules); include that package whenever a skill file changes.
- Still red on the PR head by design: `TestCleanReportsArchive_MovesOnlyEligibleCandidates` (the #1956 calendar defect); no change here.

Gaps: no whole-file semantic test of the new clauses (literal greps only, plus the one mutation); `make build` was not run (only its hash step); no Go package other than internal/template was run.
Residual risk: the three docs still describe `go test ./...` exit 0 as an ac_converge predicate (RUN-07, out of the stated order for this pass).

### Handoff record (PR #1961; leader instruction R97: freeze the head, then one independent sync audit)

decision record: decided_by=lane-7 (claude, lane orchestrator) evidence_refs=leader message R97 (operator approval),this section ladder_path=gate boundary (operator-approved audit, auto-semantics.md §6)

- Code head before this record: `c9828cb40f4a22c9c5e91a4a12afd44c404f8f45`; this record is the only change in its commit (no code change), pushed once without force.
- State: PR #1961 draft, no hold. CI at c9828cb40f (run 41335379169): the only failing jobs are Test (ubuntu-latest) and Race Test (internal/cli 1 of 4), both on `TestCleanReportsArchive_MovesOnlyEligibleCandidates` (the #1956 calendar defect, non-required per the leader). Required-check rows (Lint, Build, Analyze) were not individually re-read in this record.
- origin/main was not absorbed into this branch (leader ordered the independent audit first); the audit judges the branch as pushed.
- Open items for the audit: the Gaps and Residual risk lists above (RUN-03 wording-only pin, RUN-07 left, no end-to-end `/moai run`, `make build` not run, lint measured by CI only).
- Verdict file expected at `.moai/reports/t2008/sync-audit-1.md` (audited_sha = the head after this commit). Ready-for-review and merge are the leader's.

### Post-audit record (R156/R158 absorb, R314 resume: audit debts D1 and D2 repaired)

decision record: decided_by=lane-7 (claude, lane orchestrator) evidence_refs=leader messages R158,R314; `.moai/reports/t2008/sync-audit-1.md` (audited_sha d0177299cf, uncommitted, adoption on hold) ladder_path=gate boundary (leader ruling, auto-semantics.md §6)

- R158 absorb: `git merge --no-ff origin/main` (no conflicts), `go test ./internal/cli -run TestCleanReportsArchive` and `go test ./internal/skills` passed, pushed once without force `d0177299cf` -> `a75589ea90`; CI at that head: no failing job, required Analyze/Build/Lint pass (`gh pr checks 1961 --required`). `sync-audit-1` judged `d0177299cf`, not the later heads; its adoption is on hold (R97 batch not approved).
- D2 (stale phase labels), RED `6b1e3a5802`: `go test -count=1 -run TestRunPhaseReferences ./internal/template/` failed with five retired labels (`Phase 1~1.8`, `Phase 1.8`, `11/2B`, `or 2B`, the `Phase 5 / Phase 19` auditor row). GREEN `4c1f69d10b`: run.md rows 50/59/87/195 and context-loading.md 43/44 and phase-execution.md frontmatter/374/473/476 renumbered (`Phase 1-10`, `Phase 9`, `Phase 11/12`, `plan-auditor Phase 1, sync-auditor Phase 16/17`); mirrors copied, catalog hash regenerated; run.md stays 199 lines. The `2.8a`/decimal labels in context-loading.md lines 109/112 (harness-level skip lists) were not touched: they are display aliases under the stable phase-ID contract and need a mapping decision (Gap).
- D1 (RUN-06 producer), RED `baec92566f`: `go test -count=1 -run TestRunResumeIdentity ./internal/template/` failed on six missing items. GREEN `f871445166`: the progress.md initial entry in context-loading.md now records `Repository root`, `Worktree`, `Lease owner`, `HEAD` and `plan_artifact_hash` (refreshed at each phase checkpoint), and the Resume Check invalidates a phase with an unrecorded identity field.
- Verification after GREEN: `go test -count=1 -run 'TestRunPlan|TestRunResumeIdentity|TestRunPhaseReferences|TestSkillsContainPlanAuditGateMarkers|Mirror|Embed|Manifest|DateProvenance|BindingLedger' ./internal/template/` ok; `go test -count=1 ./internal/skills/` ok; `bash .claude/hooks/tests/test-tdd-result-contract.sh` PASS.
- Not repaired: RUN-07 (design change, left for the approval-ledger design), RUN-05/08/12 (deferred by ruling), the RUN-03 behavioral-fixture gap. Gaps: no whole-file semantic test of the new clauses (literal pins only); `make build` and golangci-lint not run locally (CI reads lint); CI at the new head not yet read.

### Sync-phase record

decision record: decided_by=lane-7 (claude, lane orchestrator; manager-docs sync pass) evidence_refs=leader dispatch for the t2008 sync pass, this section ladder_path=disk-evidence (auto-semantics.md §6 step 1)

Scope: Class B card, no SPEC directory (`git ls-files .moai/specs/t2008 .moai/specs/SPEC-t2008` -> no output), so no spec.md status transition, no §E.4 block in a SPEC progress.md, and no B12 acceptance-criterion count apply. Pre-flight: HEAD `c32801f0a7`, branch `WT-run-tdd-result`, `git status --short` empty; `git merge-base --is-ancestor a75589ea90 HEAD` exit 0 (the origin/main absorb is in this head's history); the five commits after it are the D1/D2 RED/GREEN pairs and the post-audit record.

#### Docs checked

| Path | Result |
|---|---|
| `README.md`, `README.ko.md`, `README.ja.md`, `README.zh.md` | no mention of the changed identifiers; unchanged |
| `docs-site/content/{ko,en,ja,zh}/` (all pages) | no mention of the changed identifiers; unchanged |
| `docs-site/content/en/workflow-commands/moai-run.md` lines 59-61 (Resume feature) | "On re-run, work continues from the last successful stage checkpoint." Less specific than the new Resume Check, but not contradicted by it; unchanged |
| `docs-site/content/en/workflow-commands/moai-run.md` lines 210-230 (Plan Audit Gate, Kickoff Approval) | cache-skip wording (hash unchanged, score >= 0.90) is consistent with the Go helper; the Kickoff text is a pre-existing divergence, see Residual-risk R1; unchanged |
| `CHANGELOG.md` `[Unreleased]` | `grep -c t2008 CHANGELOG.md` -> 0 before any edit; convention check below; no entry added |

Identifier scan, verbatim: `grep -rlE "tdd-result-contract|EXPECTED_RED|REGRESSION_FAILURE|plan_artifact_hash|FAIL_WARNED|MOAI_SKIP_PLAN_AUDIT|Resume Check|whitespace-normalized|Phase 2\.8|Phase 11~4|ResolveLatestPlanAudit" README.md README.ko.md README.ja.md README.zh.md docs-site/content` -> no output. The other three locale copies of `moai-run.md` were located by the scan above but not individually read (the en page is the structural twin); see Gaps.

#### CHANGELOG convention check (why no entry)

- Card-attributed entries in `[Unreleased]` (`grep -n -o -E "\(card t(200[0-9]|199[0-9]|19[0-9][0-9])[^)]*\)" CHANGELOG.md`): t1968 and t1963 (SPEC cards, Tier M/L), t1921 (a Go behavior fix in `internal/cli`). No skill-wording, rule-wording or test-only card has an entry, and none of t2005 to t2007 (the sibling cards of this batch) has one.
- This card changes instruction wording in deployed skill bodies, their template mirrors, the catalog hash, three new template tests and one Go comment. It changes no CLI behavior, config key, or API. By the observed convention an entry is expected for a SPEC close or a user-visible Go behavior change; neither applies, so adding one would create a precedent the file does not show.
- Reversible: if the leader wants an entry because `moai update` redelivers the skill bodies to user projects, it is one line under `### Changed`.

#### Docs changed

None. Reason: no user-facing document names the changed contract; the one divergence found (R1) predates this card and belongs to an autonomy-semantics docs sweep, not to this reconciliation.

#### Audit-ready signal

```yaml
sync_phase:
  card: t2008
  class: B
  spec: none
  sync_complete_at: 2026-11-03
  sync_commit_sha: pending-backfill
  sync_status: complete-pending-audit
  gap: "no independent audit was run in this pass; sync-audit-1.md judged d0177299cf (audited_sha), not a later head"
  changelog_entry: none (convention check above)
  docs_changed: none
```

`sync_status` is `complete-pending-audit`, not `complete`: every docs check below was performed, but the gate token `complete` is reserved for a state with a verdict that covers the head, and the only verdict on disk covers an earlier head.

#### Five-section evidence

**Claim 1.** No README or docs-site page names the identifiers this card changed.
- Evidence: the `grep -rlE` command in "Docs checked" -> no output (exit code not captured).
- Baseline-attribution: run in this pass against tree HEAD `c32801f0a7`.
- Gaps: a grep matches tokens, not paraphrases; a page describing the same behavior in other words would not be found (the two `moai-run.md` passages above were found by reading, not by this grep).

**Claim 2.** No CHANGELOG entry for t2008 exists and none is expected by convention.
- Evidence: `grep -c t2008 CHANGELOG.md` -> `0`; the card-id listing command above -> three entries (t1968, t1963, t1921), no t2005..t2007.
- Baseline-attribution: this pass, tree HEAD `c32801f0a7`.
- Gaps: "expected by convention" is a judgment from three data points in `[Unreleased]`, not a documented rule; I did not read a CHANGELOG policy file.

**Claim 3.** The branch contains the absorb and the D1/D2 repairs, and the tree was clean at the start of this pass.
- Evidence: `git rev-parse --short HEAD` -> `c32801f0a7`; `git branch --show-current` -> `WT-run-tdd-result`; `git status --short` -> no output; `git merge-base --is-ancestor a75589ea90 HEAD` -> exit 0; `git log --oneline a75589ea90..HEAD` -> the five commits `6b1e3a5802`, `4c1f69d10b`, `baec92566f`, `f871445166`, `c32801f0a7`.
- Baseline-attribution: this pass.
- Gaps: `.moai/reports/t2008/sync-audit-1.md` is gitignored (`git check-ignore -v` -> `.gitignore:235:.moai/reports/*`), so it does not show in status; it was neither read, staged, edited nor deleted in this pass.

**Gaps (not observed in this pass).**
1. No independent audit was run in this pass. `sync-audit-1.md` judged `d0177299cf`; heads `a75589ea90` (absorb) and the five later commits are not covered by any audit verdict.
2. No test, build, lint or CI read was performed in this pass (docs-only scope, per the dispatch); the test results quoted in earlier sections are the lane's earlier runs, not re-observed here. CI at the current head was not read.
3. The ko, ja and zh copies of `moai-run.md` were not read; the en copy was.
4. No CHANGELOG policy document was searched for.

**Residual-risk.**
- R1. `docs-site/content/en/workflow-commands/moai-run.md` (and presumably the ko/ja/zh twins) says the Implementation Kickoff Approval is a user-approval gate asked through `AskUserQuestion` that "cannot be bypassed under any circumstances" and is not skipped even for PASS-with-debt. The run skill body after this card, and `auto-semantics.md` §9.1, make the Kickoff gate autonomous by default with the operator form kept for keep-set cases. The page therefore reads stricter than the enforced contract. Not introduced by t2008 (its last edits are `5bcc778571` and `cb69888824`, earlier cards), and fixing it needs four-locale edits plus a humanize pass; recommended as a separate docs card.
- R2. The skill-body wording is pinned by literal-presence tests only (see the earlier Gaps), so a docs reader or orchestrator behavior is not shown to change.
- R3. Customer projects receive the new wording only on `moai update`; until then deployed copies keep the old Resume Check and phase labels.
