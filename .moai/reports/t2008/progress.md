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

Gaps: no whole-file semantic test of the new clauses (literal greps only, plus the one mutation); `make build` was not run (only its hash step); no Go package other than internal/template was run.
Residual risk: the three docs still describe `go test ./...` exit 0 as an ac_converge predicate (RUN-07, out of the stated order for this pass).
