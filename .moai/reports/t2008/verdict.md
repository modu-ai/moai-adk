# t2008 verdict (lane-7, Class B, Run contract reconciliation) - lane's own verdict, local commits only

Policy rules applied: `.claude/rules/moai/core/verification-claim-integrity.md` §2 (attribution), §2.3 (the baseline lands in its own commit before the change), §3.1; `.claude/rules/moai/development/verification-completeness.md` §1.1 (observed failure first), §2 (mutant probe). Decisions: `.moai/reports/pr-repair-20261101/decisions.md` R22, R23 (read through the dispatch). Source of the hypotheses: `/Users/goos/MoAI/moai-adk-go/.moai/reports/adk-workflow-pages-handoff.md` appendix 3 (read only, not copied): a static survey of 5c3f0ebf27, so every item was reproduced here before any edit.

Tree measured: branch WT-run-tdd-result, base `5c3f0ebf27584d2c45deb2d00f9d69a4eea932c4` (= origin/main at intake, `git merge --ff-only` from cb30e05b3e). Judging build: `go test` ran against this tree's source and the hook test is `bash` over this tree; no installed moai build judged anything. Nothing is pushed; no PR; `.github/workflows` untouched.

Verdict: PASS for RUN-01, RUN-02, RUN-03, RUN-06 and RUN-10 (reproduced, repaired, regression-tested). RUN-05 reproduced and NOT repaired by the leader's ruling (working basis: lane verdict = self-verification recommendation, progress and landing = the leader's evidence read; unification goes to the 3.2 approval-ledger design). RUN-13 refuted against current code and config. RUN-07, RUN-08, RUN-12 and the spec-workflow Grace Window paragraph not repaired (below).

## Claims

| # | Claim | Evidence (command -> observed output) | Baseline |
|---|---|---|---|
| C1 | RUN-03 reproduced | `bash .claude/hooks/tests/test-tdd-result-contract.sh` -> exit 1; `grep -c -F` for `semantic result`, `Compile errors`, `regression set` in `run/task-decomposition.md` -> 0, 0, 0; no workflow, skill or agent carries `EXPECTED_RED` | @5c3f0ebf27, commit `1624cd7e6c` (baseline record, before the fix) |
| C2 | RUN-03 repaired | same command -> `PASS: TDD distinguishes expected RED from regression and tool failures`; mutant (run.md reverted to base in a scratch tree) -> exit 1; commit `8c4b4bfa51` | this run |
| C3 | RUN-01: conflict is instruction text only | `runtime.GateConfig{` appears only in `_test.go` (grep over internal, cmd); `grep -rn audit-gate-merge-at internal --include='*.go'` -> one comment line, no reader; the production Kickoff admission goes through `auditverdict.Admit` and names no bypass | this run |
| C4 | RUN-01 RED then GREEN | `go test -count=1 -run TestRunPlanGatePrecedence ./internal/template/` -> exit 1 with 11 missing statements at `2669aae81f`; after `930374eaf5` -> ok | this run |
| C5 | RUN-02 RED then GREEN | same shape with `TestRunPlanCacheDescription`: 7 findings at `5868b4eb40`; after `8b35d3899c` -> ok; `go test -count=1 -run 'Cache|Hash|Review' ./internal/runtime/` -> `ok ... 0.381s`; `go vet ./internal/runtime/` -> no output | this run |
| C6 | RUN-06 RED then GREEN | `go test -count=1 -run TestRunResumeIdentity ./internal/template/` -> exit 1, 5 missing words at `0c498dc3b3`; after the fix commit -> ok | this run |
| C7 | Combined scoped run | `go test -count=1 -run 'TestRunPlan\|TestRunResumeIdentity\|TestSkillsContainPlanAuditGateMarkers\|Mirror\|Embed\|Manifest\|DateProvenance\|BindingLedger' ./internal/template/` -> `ok github.com/modu-ai/moai-adk/internal/template 3.324s` | this run, HEAD of the RUN-06 fix |
| C8 | Mirrors and catalog | each edited skill file `cmp`-identical to its `internal/template/templates/` twin (copied in the same commit); `TestManifestHashFormat` failed with `CATALOG_HASH_UNSTABLE` after the first skill edit and passes after `go run ./internal/template/scripts/gen-catalog-hashes.go --all` (one line of `catalog.yaml` per commit) | this run |
| C9 | RUN-10 RED then GREEN | `go test -count=1 -run TestRunPhaseReferences ./internal/template/` -> exit 1 with 3 findings (`'proceeding to Phase 3' names no phase heading`, router `Phase 11~4`, router `2.8`) at `5bd50fabf0`; after the fix commit -> ok; combined scoped run (now also `TestRunPhaseReferences`) -> `ok github.com/modu-ai/moai-adk/internal/template 6.576s`; the RUN-03 hook test -> PASS | this run |
| C10 | RUN-13 not reproduced | `candidate_ci.enabled: false` and `base_freshness_gate.enabled: false` in the live and template `workflow.yaml`; `candidateCIRequired` documents unreadable-config = off | this run |

Per-item reproduction output, RED output and design notes: `.moai/reports/t2008/progress.md` (committed; baseline records precede each fix commit, so the commit graph witnesses the order).

## Not repaired - disposition and reason

- RUN-05 (final verdict owner): contradiction reproduced inside the factory-dispatch rule set (core:44 and dispatch:148 say the final PASS/FAIL is the leader's; detail:140/169/176 and mechanics:145 say the per-card verdict is the lane's). The leader confirmed the conflict and ruled: no rule sentence changes now; operating basis as in the Verdict line; unification goes to the 3.2 approval-ledger (R18) design. The decision-kind split is preserved in `progress.md` as input.
- RUN-08 (autofix budget unit), RUN-12 (worktree path guidance) and the `spec-workflow.md` Grace Window paragraph: deferred by the leader's ruling. Reasons recorded in `progress.md` (agentemit source and Codex twin for RUN-08; a rule-file edit with a path-policy owner for RUN-12 and the Grace paragraph).
- RUN-07 (ac_converge evidence admission): a design change (loop stop versus completion gate, evidence keys); no current-code defect was shown, so nothing is claimed either way; left for the 3.2 design.
- RUN-13: see C10 - refuted for the shipped and live configuration; the residual (unreadable config reads as off if a project later turns the gate on) needs a policy decision.

## Gaps

- RUN-03 pins wording only ("문구 존재만 고정"): the hook test and the added greps prove the sentences exist, not that an orchestrator or `manager-develop` classifies a RED correctly; no behavioral fixture (normal RED->GREEN->regression, compile failure, timeout, other-test failure, re-failure after GREEN, zero tests) exists.
- No end-to-end `/moai run` entry or resume was executed; the three skill-text repairs are checked by literal and mutant tests, not by observing an orchestrator follow them.
- `make build` was not run; only its catalog-hash step. Only `./internal/template/` and `./internal/runtime/` (comment change) were tested; no full `./internal/cli/...` run (forbidden by R22).
- `spec-workflow.md` Grace Window paragraph (rule file) still describes FAIL warn-only; correcting it needs its own byte statement and ledger check.
- Refusals: the worktree guard refused no command in this pass. Output of `gen-catalog-hashes.go` was redirected to the scratchpad after the first run printed 71 lines (the first run's output was not bounded).

## Residual risk

- The new tests are word-level; a repair can satisfy them with text that is semantically wrong (the survey's behavioral fixtures are not built).
- RUN-01 wording depends on `auto-semantics.md` §9.1; if that ruling changes, the precedence paragraph in `run/phase-execution.md` must change with it.
