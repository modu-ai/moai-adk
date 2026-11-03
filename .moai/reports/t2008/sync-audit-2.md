auditor-model: claude-opus-5-5
verdict: FAIL
score: 0.74
audited_sha: 0e010a2d3f038b406fb0c81592e60c1fd5175b84

# Sync audit 2: card t2008 (Class B, no SPEC), PR #1961 (draft), branch WT-run-tdd-result

- Tree: /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t2008
- Diff measured: `origin/main...HEAD`, with merge base `59464c5f37` (this equals origin/main, and the absorb `a75589ea90` is in history). 17 files, +792/-48.
- HEAD and working tree, at the start and again just before this file was written: `git rev-parse HEAD` returned `0e010a2d3f038b406fb0c81592e60c1fd5175b84` and `git status --short` printed nothing.
- Judging build (VCI 2.2): `moai version` returned `v3.2.0-rc.39   v3.2.0-1041-gdad3603de1   built 2026-11-02T19:17:51Z`, installed at `/Users/goos/go/bin/moai`, not built from this tree. The moai MCP server that ran `audit_multi` reports `build_commit: b1806498fe` (v3.2.0-rc.38, built 2026-10-21). Tree HEAD is `0e010a2d3f`. I did not check whether either build commit is an ancestor of HEAD (see Gaps).
- Prior context: `sync-audit-1.md` (judged `d0177299cf`, PASS-WITH-DEBT 0.88, debts D1 and D2). I used it only as background and did not re-score from it.
- card-review.md: absent by decision. Reason: "Codex cost hold; independent audit Codex leg covers". `codex_review` was not run.

## Overall verdict

FAIL. The required Codex gate returned `fail`, and the leg is fail-closed. My own reading independently confirms all three Codex findings. Two of them mean the earlier debts are only partly cleared:
- D1 (RUN-06 producer) is not cleared on the normal path.
- D2 (RUN-10 stale labels) is partly cleared.

The mechanical checks the lane claimed all reproduce green: tests, mirrors, the LOC ceiling, gofmt/vet, and lint on internal/template. The failure is about whether the repairs are complete, not about any red check.

## Findings (structured defect list)

- **F1 [High] [blocking] `.claude/skills/moai/workflows/run/context-loading.md:158-169`. The D1 producer cannot be reached on the normal path.**
  - The identity fields (`Repository root`, `Worktree`, `Lease owner`, `HEAD`, `plan_artifact_hash`) are written only in the "If it does not exist: Create the file now" branch.
  - The plan phase always creates `progress.md`: `.claude/agents/moai/manager-spec.md:96` says "`progress.md` is emitted at every Tier", and `:125` says "emit a `progress.md` file carrying the canonical `§E` section skeleton". So a normal `/moai run` reaches the "exists" branch, and the identity fields are never written.
  - Line 169 only "Refresh[es] `HEAD` and `plan_artifact_hash` in the same entry", and that entry does not exist.
  - Effect: the new rule "invalidate a phase ... that carries an unrecorded identity field" invalidates every completed phase on every resume. This fails safe (nothing is wrongly skipped), but resume is never usable, and the debt "Resume Check identity fields have no producer" is still open for the standard flow.
  - Codex reported the same independently (P2, high confidence).
  - Required fix:
    - In the "exists" branch, add a step: when the identity block is missing, write it now, scoped to phases checkpointed from this point on. Never treat earlier checkpoints as verified after the fact.
    - Extend `TestRunResumeIdentity` so it asserts that the identity write is reachable when progress.md already exists.
    - Mirror the change and regenerate the catalog hash.

- **F2 [Medium] [blocking] `context-loading.md:169` (with `:157`). A single shared identity record that is overwritten at each checkpoint cannot support per-phase invalidation.**
  - Line 157 compares "the recorded identity" per completed phase and invalidates "every later phase that consumed its output".
  - Line 169 overwrites one shared `HEAD` / `plan_artifact_hash` at every checkpoint.
  - Example: the plan hash changes after Phase 5's checkpoint, then Phase 9's checkpoint refreshes the shared value. Phase 5 is now compared against the new value and is not invalidated, even though it ran on the old plan.
  - Codex reported the same independently (P2, high confidence).
  - Required fix: record `HEAD` and `plan_artifact_hash` on each phase checkpoint line, as the values that phase ran on. Compare per phase. Pin this with a literal test.

- **F3 [Medium] [blocking] The D2 / RUN-10 stale phase labels are only partly cleared, and the test cannot see the remaining ones.** Retired labels still present at the audited SHA:
  - `run/task-decomposition.md:2`: `description: "Run Phase 11~4 — ..."`. This is the same retired label RUN-10 removed from the router.
  - `run/task-decomposition.md:6`: `phase: "Phase 11-4: ..."`.
  - `run/phase-execution.md:6`: `phase: "Phase 1-1.8: ..."`. D2 renumbered the `description:` line one key above but left this one.
  - `run.md:71`: `Phase Sequence (0.5~1.8)`, in the router itself.
  - `run/phase-execution.md:16` ("entering Phase 1"), `:70` ("proceed to Phase 1"), `:96` ("continue to Phase 1"), `:199` ("skip to Phase 1"). These are hand-offs inside Phase 1/2 that point back to Phase 1. Line 96 sits right under the line this card added (`:94`) and contradicts `:95` ("proceeding to Phase 5").
  - Why the test misses them: `internal/template/run_phase_reference_test.go:39` checks `Phase 11~4` only against `router`, `:47` checks a fixed literal list, and `:23` matches only the `proceeding to Phase N` hand-off form in task-decomposition.md.
  - Codex reported the same independently (P2, high confidence).
  - Required fix:
    - Correct the labels and the four hand-offs, mirror them, and regenerate the catalog hash.
    - Extend the checker to cover frontmatter `description:` / `phase:` in all run bodies, the router's `Phase Sequence` range, and the `proceed to` / `continue to` / `skip to` hand-off forms in phase-execution.md, checked against existing headings.
    - Keep the mutants.

- **F4 [Low] [optional] `context-loading.md:163`.** "`Repository root: {git rev-parse --show-toplevel of the main checkout}`". Inside a worktree, `--show-toplevel` returns the worktree path, so the placeholder contradicts itself. Recording and comparison use the same command, so no false match follows. Fix: name the command that resolves the main checkout (the parent of `git rev-parse --path-format=absolute --git-common-dir`), or drop "of the main checkout".

- **F5 [Low] [optional] `phase-execution.md:94-95`.** A PASS-WITH-DEBT verdict is routed under "4a. PASS", which then logs `verdict=PASS`. The log line should carry the verdict that was actually admitted, so the debt admission can be seen in the log.

- **F6 [Low] [optional] `phase-execution.md` `--skip-audit` section together with `context-loading.md:167`.** The skip path skips Step 1, so `plan_artifact_hash` is never computed, and the Resume Check then always invalidates. This fails safe and does not need a change. Note it beside F1.

- **F7 [Info] [optional] `.moai/workflows/run*.md`.** These are tracked snapshot copies, last touched in `82677fd271` (2026-10-05). They differ from the live files (`cmp` exit 1 for all four) and already differed from base (`base-vs-.moai=1`, 25 diff lines for run.md). The drift predates this card and is out of scope. I note it because a reader could take it for an unmirrored copy.

## Claim / Evidence / Baseline-attribution / Gaps / Residual-risk

**Claim 1. The scoped template tests pass at the audited SHA.**
- Evidence: `go test -count=1 -run 'TestRunPlan|TestRunResumeIdentity|TestRunPhaseReferences|TestSkillsContainPlanAuditGateMarkers|Mirror|Embed|Manifest|DateProvenance|BindingLedger' ./internal/template/` printed `exit=0` and `ok  	github.com/modu-ai/moai-adk/internal/template	4.064s`.
- Baseline-attribution: this run, tree `0e010a2d3f`, toolchain from PATH.

**Claim 2. The `internal/skills` package passes, including the router LOC ceiling.**
- Evidence: `go test -count=1 ./internal/skills/` printed `ok  	github.com/modu-ai/moai-adk/internal/skills	0.270s`.
- Evidence: `wc -l` printed `199 .claude/skills/moai/workflows/run.md` and `199 internal/template/templates/.claude/skills/moai/workflows/run.md`. The ceiling is 200.

**Claim 3. The TDD result contract hook test passes.**
- Evidence: `bash .claude/hooks/tests/test-tdd-result-contract.sh` printed `PASS: TDD distinguishes expected RED from regression and tool failures` and `hook_exit=0`.

**Claim 4. The live skill files are byte-identical to their template mirrors.**
- Evidence: `cmp` returned `run.md=0`, `context-loading=0`, `phase-execution=0`, `task-decomposition=0`.

**Claim 5. The RUN-02 wording matches the Go code.**
- Evidence for the hash: `internal/runtime/audit_cache.go:151-155` reads `// Include the path and exact bytes as separators ...` / `fmt.Fprintf(h, "%s:%d\x00", name, len(data))` / `h.Write(data)` / `h.Write([]byte{0})`. This matches "exact bytes ... framed by file name, length, and NUL".
- Evidence for the cache source: `internal/runtime/audit_review.go:38-55` reads `prefix := specID + "-review-"`, then `strconv.Atoi(...)`, then `if parseErr != nil || n < 1 || n <= bestNumber { continue }`. This matches "largest integer N wins; a name that does not parse is ignored".

**Claim 6. The RUN-01 claim "no Go code reads audit-gate-merge-at.txt" holds.**
- Evidence: `grep -rn "audit-gate-merge-at" --include='*.go' internal cmd` matched only `internal/runtime/audit_gate.go:159:	// T0 is the grace window start time. Read from .moai/state/audit-gate-merge-at.txt.`, which is a comment.
- Evidence that the precedence terms exist in the rule: `.claude/rules/moai/workflow/auto-semantics.md:231,234,240-241` carry `PASS-WITH-DEBT`, `(FAIL, INCONCLUSIVE, BYPASSED,` and `Binding run conditions`.

**Claim 7. Static checks are clean.**
- Evidence: `gofmt -l` over the four new test files and `audit_cache.go` printed nothing (`gofmt_exit=0`).
- Evidence: `go vet ./internal/template/ ./internal/runtime/` returned `vet_exit=0`.
- Evidence: `golangci-lint run --timeout=4m ./internal/template/` printed `0 issues.`.

**Claim 8. Security probes found nothing in the diff.**
- Evidence: a secret-pattern grep over the `origin/main...HEAD` diff matched only the prose word "token" in progress.md (diff lines 449 and 456), which are false positives.
- Evidence: an exec/http/WriteFile grep over added lines printed nothing. The only Go production change is a comment (`audit_cache.go` +3/-1).

**Claim 9. Findings F1 to F3 are real.**
- Evidence for F1: `manager-spec.md:96` and `:125` (quoted above), together with `context-loading.md:158-169`.
- Evidence for F3: `grep -nE ...` printed `task-decomposition.md:2:description: "Run Phase 11~4 ..."`, `phase-execution.md:6:  phase: "Phase 1-1.8: ..."`, `phase-execution.md:16/70/96/199` ("Phase 1"), and `run.md:71: ... → Phase Sequence (0.5~1.8) + mode routing`. Separately, `head -12 task-decomposition.md` showed `phase: "Phase 11-4: ..."`.
- The `TestRunPhaseReferences` code was read at lines 22-23, 39 and 47.

**Baseline-attribution.** Every command above ran in this session against tree `0e010a2d3f038b406fb0c81592e60c1fd5175b84`, with a clean working tree, from `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t2008`.

**Gaps.**
1. The worktree-isolation guard refused one compound Bash batch (the skills test, hook test, cmp and wc combined). Fallback: I ran each command separately and recorded its output above. No result rests on the refused form. A second refused command (`git show` inside a loop) was replaced by a single `git show` into a scratchpad file.
2. I did not check whether the judging builds (`dad3603de1` installed, `b1806498fe` MCP server) are ancestors of HEAD. `moai verify audit-plan` and `audit_multi` ran on those builds.
3. I did not run golangci-lint on `./internal/runtime/`, or read its exit code (the zsh `PIPESTATUS` came back empty). The "0 issues." output covers `./internal/template/` only.
4. I did not run `make build`, `go test ./...` or `./internal/cli` (out of scope by instruction). I did not read CI at this head.
5. I did not run `/moai run` or a resume end to end. F1 and F2 come from reading the instructions, not from a resume that was run.
6. Only the Codex backend ran. The Claude and GLM gates are `off` by config (`moai verify audit-plan`: claude off, codex required, glm off). No leg was skipped or failed open (`fail_open_backends: []`).

**Residual-risk.**
- R1. The new clauses are pinned by literal-presence tests only (the RUN-03 hook greps, and the RUN-01/02/06 Go literal tests plus mutants). An orchestrator's real behavior is not exercised.
- R2. Customer projects get the wording only on `moai update`.
- R3. `spec-workflow.md` still carries the legacy Grace Window paragraph (deferred by ruling). A reader of that rule alone can still take the FAIL_WARNED reading.

## Dimension scores (flat weighted mode; built-in default profile, since no SPEC means no `evaluator_profile`)

| Dimension | Score | Verdict | Evidence |
|---|---|---|---|
| Functionality (40%) | 62/100 | FAIL | D1 producer unreachable on the normal path (F1); per-phase identity overwritten (F2); D2 partial (F3). Claims 1-6 reproduce. |
| Security (25%) | 95/100 | PASS | Claim 8: no secrets or injection surfaces; Go change is comment-only. |
| Craft (20%) | 74/100 | PASS | Claim 7 clean; the RUN-10 test misses the remaining labels (F3); tests check wording presence only (R1). |
| Consistency (15%) | 88/100 | PASS | Claim 4 mirrors byte-identical; catalog hash test passes (Claim 1); F4/F5 are small wording inconsistencies. |

Must-pass firewall: Functionality fails, so the overall result is FAIL regardless of the other scores. The required Codex gate is also FAIL.

## Binding run conditions

None apply. There is no SPEC progress.md and no `Binding run conditions` heading in `.moai/reports/t2008/progress.md`.

## Prior debt disposition (from sync-audit-1)

| Debt | Status | Evidence |
|---|---|---|
| D1 Resume Check producer | partly cleared, still open on the normal path | F1 |
| D2 stale phase labels | partly cleared | F3 |

## Per-backend verdicts (audit_multi digest)

- overall_verdict: `fail`
- gate_unmet: not present in the result. `residual_risk_note`: "required-backend FAIL: codex".
- plan_source: `config`
- per_backend_verdicts:
  - backend `codex`, gate `required`, verdict `fail`. Findings: P2 at context-loading.md:158, P2 at context-loading.md:169, P2 at run_phase_reference_test.go:34.
- participant_count: 1; fail_open_backends: []; disagreement_flag: null
- tree_root: `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t2008`; build_commit: `b1806498fe`
- audit_receipt: `rcpt-91596827d55973873e76a231`

This verdict is not final until the orchestrator writes the digest to `.moai/state/audit-plan-result.json` and `moai verify audit-plan --project-root <toplevel> --result-file <path>` passes its convergence check. Here, the required-backend fail already makes the convergence result `fail`.

## Iteration history

- sync-audit-1: judged `d0177299cf`, PASS-WITH-DEBT 0.88 (D1, D2). Not adopted (hold).
- sync-audit-2 (this file): judged `0e010a2d3f`, FAIL 0.74 (F1-F3 blocking). Scope for the re-audit: the F1-F3 defect delta.
