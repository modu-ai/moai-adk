# t2008 verdict (lane-7, Class B, Run contract reconciliation) - lane's own verdict, local commits only

Policy rules applied: `.claude/rules/moai/core/verification-claim-integrity.md` §2 (attribution), §2.3 (the baseline lands in its own commit before the change), §3.1; `.claude/rules/moai/development/verification-completeness.md` §1.1 (observed failure first), §2 (mutant probe). Decisions: `.moai/reports/pr-repair-20261101/decisions.md` R22, R23 (read through the dispatch). Source of the hypotheses: `/Users/goos/MoAI/moai-adk-go/.moai/reports/adk-workflow-pages-handoff.md` appendix 3 (read only, not copied): a static survey of 5c3f0ebf27, so every item was reproduced here before any edit.

Tree measured: branch WT-run-tdd-result, base `5c3f0ebf27584d2c45deb2d00f9d69a4eea932c4` (= origin/main at intake, `git merge --ff-only` from cb30e05b3e). Judging build: `go test` ran against this tree's source and the hook test is `bash` over this tree; no installed moai build judged anything. Nothing is pushed; no PR; `.github/workflows` untouched.

Verdict: PASS for RUN-01, RUN-02, RUN-03 and RUN-06 (reproduced, repaired, regression-tested). RUN-05 reproduced and NOT repaired (a policy ruling is the leader's). RUN-07, RUN-08, RUN-10, RUN-12, RUN-13 not repaired in this pass (below).

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

Per-item reproduction output, RED output and design notes: `.moai/reports/t2008/progress.md` (committed; baseline records precede each fix commit, so the commit graph witnesses the order).

## Not repaired - disposition and reason

- RUN-05 (final verdict owner): contradiction reproduced inside the factory-dispatch rule set (stub says the final PASS/FAIL is the leader's, detail and mechanics say the per-card verdict is the lane's). Edit withheld: the sentence is in the leader's own always-loaded/role-injected stub, so the ruling is the leader's; the progress record carries the proposed decision-kind split.
- RUN-08 (autofix budget unit): the wording "per PR push" (manager-develop, prompt template) versus a PR-scoped persisted counter (`ci-autofix-protocol.md`) was read in this tree; not repaired because `.claude/agents/moai/manager-develop.md` is emitted from `internal/template/agentemit/source/` and the Codex twin must move with it - a separate emit change.
- RUN-12 (worktree path guidance): two worktree-integration.md statements read (home projects root for `moai worktree new` versus "WT worktrees live only under `.moai/worktrees/`"); not repaired - a rule-file edit with a path-policy owner; the SKILL.md claim was not read.
- RUN-07, RUN-10, RUN-13: not examined beyond the survey text; RUN-13 needs a read of the Go consumers, RUN-07 and RUN-10 are design/generation changes. Nothing here claims them verified or refuted.

## Gaps

- No end-to-end `/moai run` entry or resume was executed; the three skill-text repairs are checked by literal and mutant tests, not by observing an orchestrator follow them.
- `make build` was not run; only its catalog-hash step. Only `./internal/template/` and `./internal/runtime/` (comment change) were tested; no full `./internal/cli/...` run (forbidden by R22).
- `spec-workflow.md` Grace Window paragraph (rule file) still describes FAIL warn-only; correcting it needs its own byte statement and ledger check.
- Refusals: the worktree guard refused no command in this pass. Output of `gen-catalog-hashes.go` was redirected to the scratchpad after the first run printed 71 lines (the first run's output was not bounded).

## Residual risk

- The new tests are word-level; a repair can satisfy them with text that is semantically wrong (the survey's behavioral fixtures are not built).
- RUN-01 wording depends on `auto-semantics.md` §9.1; if that ruling changes, the precedence paragraph in `run/phase-execution.md` must change with it.
