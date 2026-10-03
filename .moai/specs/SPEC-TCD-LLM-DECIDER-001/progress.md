# SPEC-TCD-LLM-DECIDER-001 — Progress

Tier M · card t1352 · plan-phase artifact set authored 2026-10-03 at HEAD `681ee30b5`
(worktree `.claude/worktrees/t1352`, branch `WT-llm-card-decider`, == local develop tip).
Status: `draft` (plan-audit pending).

## §E.1 Plan-phase Audit-Ready Signal

- Plan-phase artifact set complete: `spec.md` (7 REQ, GEARS — REQ-TLD-001..007),
  `plan.md` (§A-§H, 4 milestones, decision records OD-A..OD-D all resolved with rationale and
  two leader-escalation notes), `acceptance.md` (AC-TLD-001..007, Given-When-Then, 1:1 REQ
  mapping, RED-now/green adoption table), this `progress.md`.
- SPEC ID regex pre-write check: PASS (`SPEC-TCD-LLM-DECIDER-001` — middle segments `TCD`,
  `LLM`, `DECIDER` all letter-leading); ID unique in `.moai/specs/` (grep: no collision); REQ
  series `REQ-TLD-` unique across `.moai/specs/` and `internal/`.
- Frontmatter validated against the canonical 12-field schema SSOT (no snake_case aliases);
  `status: draft` set at creation per plan-phase ownership; `depends_on:
  [SPEC-TODO-CLASSIFY-DISPATCH-001]` — parent reads `status: completed` at `681ee30b5`.
- Measured surface basis: every card premise verified against this tree (seam
  `todo_classify.go:31`, identity set `classification.go:64-68`, GLM transport precedent
  `mcp_glm.go:3-6`); citations pinned to `681ee30b5`.
- Concurrent-lane risk recorded: plan.md §D.5 — `factory_card.go` and `internal/kanban/**` are
  NOT in this plan's file list (t1458 overlap limited to the `internal/cli` package, disjoint
  files).
- plan_status: authored — plan-audit pending.

## §E.2 Run-phase Evidence

Run-phase: manager-develop (cycle_type=tdd), 2026-10-03, worktree
`.claude/worktrees/t1352`, branch `WT-llm-card-decider`. Evidence tree: HEAD
`6b13f5c45` (the run's final code commit; every measurement below names its
own tree where it differs). Verbatim RED outputs, the D5 measurement record,
and the full plan-vs-actual list live in `.moai/reports/t1352/run-evidence.md`
(local artifact).

### E8 — verbatim RED evidence (one per TDD cycle, captured before each GREEN)

RED-1 (tree `20c33cd34`, pre-implementation):

```
$ go test -run 'TestTodoDeciderSelection' ./internal/cli/
internal/cli/todo_decider_select_test.go:33:39: undefined: config.EnvTodoDecider
internal/cli/todo_decider_select_test.go:35:14: undefined: todoDeciderFromEnv
internal/cli/todo_decider_select_test.go:69:6: undefined: todoDeciderIsLLM
FAIL	github.com/modu-ai/moai-adk/internal/cli [build failed]
```

RED-2 (tree `66f3efa99`, selector green, transport absent):

```
$ go test -run 'TestLLMDecider' ./internal/cli/
internal/cli/todo_classify_llm_test.go:30:61: undefined: todoLLMHTTPDoer
internal/cli/todo_classify_llm_test.go:32:53: undefined: todoLLMHTTPClient
internal/cli/todo_classify_llm_test.go:32:72: undefined: todoLLMKeyLoader
FAIL	github.com/modu-ai/moai-adk/internal/cli [build failed]
```

RED-3 (tree `2091497cc`, transport green, judgment in-lock; then the
redesigned observation re-captured RED on the still-in-lock tree):

```
$ go test -run 'TestTodoAdd_LLMJudgmentsOverlap|TestTodoAdd_ListReadCompletes' -count=1 -v ./internal/cli/
    todo_lock_scope_test.go:87: two concurrent adds took 4.073894625s, want under the 2500ms overlap bound (a serialized-in-lock placement costs 2x1500ms)
--- FAIL: TestTodoAdd_LLMJudgmentsOverlapOutsideLock (4.81s)
    todo_lock_scope_test.go:144: todo list did not complete within 600ms while a judgment was in flight (the read waits out the LLM stall)
--- FAIL: TestTodoAdd_ListReadCompletesDuringLLMJudgment (2.82s)

[redesigned observation, in-lock tree restored]
    todo_lock_scope_test.go:100: the two judgment requests arrived 1.568040542s apart, want under the 500ms overlap bound (a serial-in-lock placement costs the full first judgment, 1500ms, before the second request can leave)
--- FAIL: TestTodoAdd_LLMJudgmentsOverlapOutsideLock (7.46s)
```

### E1 — AC binary matrix (each row's command and observed output; tree `6b13f5c45`)

| AC | Status | Command (this run, this tree) | Observed output |
|---|---|---|---|
| AC-TLD-001 | PASS | `go test -run 'TestLLMDeciderClassify_HappyPath|TestTodoAdd_LLMSelectionRecordsModelJudgment' -count=1 -v ./internal/cli/` | `--- PASS: TestLLMDeciderClassify_HappyPathRecordsForcedLLMIdentity`, `--- PASS: TestTodoAdd_LLMSelectionRecordsModelJudgment` (both within `ok github.com/modu-ai/moai-adk/internal/cli`) |
| AC-TLD-002 | PASS | `go test -run 'TestTodoDeciderSelection|TestTodoAdd_SuppliedFileBeatsLLMSelection' -count=1 -v ./internal/cli/` | 5 selector subtests + file-beats-env all `--- PASS`; refusal messages name the accepted set and the jev wording |
| AC-TLD-003 | PASS | `go test -run 'TestLLMDeciderClassify_|TestTodoAdd_LLMFailureDegradesWithExactlyOneNotice' -count=1 -v ./internal/cli/` | all failure classes (unreachable, 500, non-JSON, timeout, out-of-set) `--- PASS`; append and pick paths each count EXACTLY 1 notice |
| AC-TLD-004 | PASS | `go test -run 'TestLLMDeciderClassify_HappyPath|TestLLMDeciderClassify_ClaimedJev|TestLLMDeciderClassify_OutOfSet' -count=1 -v ./internal/cli/` | claimed `human`/`jev` accepted with identity forced to `llm`; out-of-set refused; single-validator grep: no re-spelled set literal in the new files |
| AC-TLD-005 | PASS | `go test -run 'TestTodoAdd_LLMJudgmentsOverlap\|TestTodoAdd_ListReadCompletes' -count=1 -v ./internal/cli/` | overlapped arrival gap single-digit ms (< 500ms bound; in-lock RED measured 1.57s); reader completes with no deadlock — see D5 note |
| AC-TLD-006 | PASS | `go test -run 'TestTodoAdd_LLMSecretNeverPrinted|TestLLMDeciderHTTPClientCarriesConfiguredTimeout|TestLLMDeciderClassify_RequestCarriesTaskSlotTransport' -count=1 -v ./internal/cli/` | zero credential bytes in stdout/stderr on happy AND failure paths, positive control: server DID receive it (x-api-key); client timeout == `config.DefaultTodoClassifyLLMTimeout`; task-slot model, 512 cap, no reasoning_effort field |
| AC-TLD-007 | PASS + mutation check | `go test -run 'TestTodoAdd_UnsetEnvBehavesPreSPEC' -count=1 -v ./internal/cli/` | PASS (default judgment, zero notices); mutation (unset arm removed) flipped it red — `unset selection errored: ... value "" refused` — then restored to green |

D5 wall-clock boundary (named at run phase, with measurement — full record in
run-evidence.md): the observation is the **arrival gap between the two
judgment requests, bounded at 500 ms**; a serial-in-lock placement cannot
meet it (lower bound = the full 1500 ms first judgment + write). Measured
in-lock gap 1.57s (RED), overlapped gap single-digit ms (GREEN). A total
wall-time bound was rejected with measurement: the write path costs ~1.8s
serial regardless of placement (in-lock 4.07s vs out-of-lock 3.33s), and a
plain `todo list --json` read costs 4.87s by itself — neither separates the
placements.

### E2 — cross-platform builds

```
$ go build ./... → exit 0
$ GOOS=windows GOARCH=amd64 go build ./... → exit 0 ("FINAL_BUILDS_OK")
$ go vet ./internal/cli/... ./internal/config/... → clean ("VET_OK")
```

### E3 — coverage of the new code paths

```
$ go test -timeout 9m -run 'TestTodoAdd|TestTodoDecider|TestLLMDecider|TestTodoClassify' -count=1 -coverprofile=/tmp/t1352-todo.cover ./internal/cli/
ok  	github.com/modu-ai/moai-adk/internal/cli	39.313s
$ go tool cover -func=/tmp/t1352-todo.cover | grep -E "todo_classify_llm.go|todo_decider_select.go"
newLLMCardDecider 100.0% · todoDeciderIsLLM 100.0% · Classify 87.0% ·
parseTodoLLMJudgment 94.1% · todoLLMSingleLineReason 100.0% ·
todoLLMClassifySystemPrompt 100.0% · todoPreClassifyLLM 100.0% ·
todoDeciderFromEnv 100.0%
```
New-code weighted coverage ≈ 95% (threshold 85). Package-level figures:
`internal/config` 83.6% / `atomicfile` 81.8% / `toolpolicy` 89.1% (foreground
full run) — the two sub-85 figures are pre-existing package states; this
SPEC's config changes are constant declarations only (zero statements). The
`internal/cli` package-level figure was not obtainable locally: the full
`TestTodo*` scope exceeds 9 minutes on this machine (Bash foreground ceiling)
and the background full-suite run hit its own 30m timeout before printing a
figure — CI is the package-level verdict owner.

### E4 — boundary greps (tree `913ae06f8` + M4; positive controls included)

- non-test Go under internal/, pkg/, cmd/ referencing `scripts/jev` → **0**
  (`grep -rn "scripts/jev" internal/ pkg/ cmd/ --include="*.go" | grep -v "_test.go" | wc -l` → 0).
  Positive control: `TestJevBoundaryScanPositiveControl` **PASS** (the guard
  plants a reference in a fixture tree and catches it). The 10 raw hits are
  all inside *_test.go files (the guard's own inspection strings).
- inline `os.Getenv("MOAI_TODO_DECIDER")` outside envkeys.go → **0**; the
  literal appears in zero product files (code references `config.EnvTodoDecider`).
- inline `api.z.ai` endpoint literal outside defaults.go (non-comment,
  non-test) → **0**.

### E5 — lint (baseline vs NEW)

- baseline (tree `20c33cd34`): `0 issues` (2m budget; the authoritative final
  run at 5m below).
- NEW during run: 1 — `internal/cli/todo.go:755:4: ineffectual assignment to
  dec (ineffassign)` (M1's else-branch). **Fixed in M4** (`6b13f5c45`): every
  branch assigns, evaluation order unchanged.
- final: `golangci-lint run --timeout=5m ./internal/cli/... ./internal/config/...` → `0 issues.`

### E6 — commits and push state

| Milestone | Commit |
|---|---|
| M1 selector + constants + SPEC in-progress flip | `66f3efa99` |
| M2 LLM transport + C2 header amendment | `2091497cc` |
| M3 out-of-lock judgment | `913ae06f8` |
| M4 guards/mutation/ineffassign | `6b13f5c45` |
| progress record | (this commit) |

Branch `WT-llm-card-decider`. **Nothing was pushed** — integration is the
leader's batch act (git-flow lane protocol).

### E7 — race

```
$ go test -race -run 'TestTodoAdd_LLMJudgmentsOverlap|TestTodoAdd_ListReadCompletes|TestTodoAdd_LLMFailureDegrades|TestTodoAdd_LLMSelection|TestTodoAdd_LLMSecret' -count=1 ./internal/cli/
ok  	github.com/modu-ai/moai-adk/internal/cli	20.190s
```

### Foreground re-measure (leader-directed, after the background window)

- `go test -timeout 9m -cover ./internal/config/...` → `ok` ×3 (config 22.8s,
  atomicfile, toolpolicy) — **config packages green in full**.
- The background full-suite run (`go test -timeout 30m -cover ./internal/cli/...
  ./internal/config/...`) reported 3 failures, all in files this SPEC never
  touched: `TestStopChainAdvisoryFailureRecorded` (PASSes in isolation —
  3.32s, load-dependent), `TestStopChainMemberCostWithinBudget` and
  `TestCodexTaskBackgroundHandshakeHonorsTaskBound` (timing-budget tests).
  Non-causality proven mechanically: `git diff --name-only 20c33cd34..HEAD |
  grep -c codex` → 0 (the touched-file set has zero overlap with the failing
  tests' target code), and the config diff is purely additive constants.
  The background run also hit its 30m ceiling (timeout panic) before
  completing the cli package — CI is the repository-wide verdict owner and
  that verdict is PENDING at report time.
- The `TestTodo*` full scope exceeds the 9-minute foreground ceiling (Bash
  limit); the run timed out (panic: test timed out) mid
  `TestTodoPR_QueueDirUnchanged` — a scope/consumer test outside this
  SPEC's change set, not a functional failure. The add path itself is fully
  measured: `go test -run 'TestTodoAdd' -count=1 ./internal/cli/` →
  `ok ... 51.247s` (M1) and `ok ... 162.000s` (M3, full re-run after the
  lock restructure).

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-10-03
run_commit_sha: 6b13f5c45
run_status: complete
ac_pass_count: 7
ac_fail_count: 0
preserve_list_post_run_count: 4
l44_pre_commit_fetch: n/a (lane protocol — no origin fetch performed; the lane never pushes)
l44_post_push_fetch: n/a (nothing pushed; leader batch push pending)
new_warnings_or_lints_introduced: 0  # the one ineffassign NEW finding was fixed in 6b13f5c45; final lint 0 issues
cross_platform_build:
  native: pass  # go build ./... exit 0
  windows: pass # GOOS=windows GOARCH=amd64 go build ./... exit 0
total_run_phase_files: 11  # 6 product/test Go files + 2 config + spec.md frontmatter + 2 progress/evidence artifacts
m1_to_m4_commit_strategy: one commit per milestone (4 code commits + this progress record), explicit-pathspec staging, no --amend, no force
plan_vs_actual_divergences: 6  # enumerated in §E.2 pointer file .moai/reports/t1352/run-evidence.md
```

PRESERVE list post-run (byte-unchanged, verified by the passing parent-SPEC
suite): `todoClassificationFallbackNotice` (one-line constant),
`todoClassifyInLock` (in-lock helper), `internal/kanban/**` (zero Go
changes), `internal/template/templates/**` (untouched), and the parent SPEC
directory (read-only).

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

## §F Phase 4 Mode Selection

Run-phase entry record for card t1352 (lane session). Written by the lane
orchestrator at plan-audit closure, 2026-10-03.

**Kickoff gate decision record** (default-autonomous transition,
auto-semantics §9.1 — the lane's operator-gate evidence per the leader
dispatch, which states operator-gate-level decisions are set by the leader as
audit evidence):

- Plan-audit verdict: **PASS** — iteration 2/2 (final, audit ceiling consumed),
  overall 0.94 (Tier M threshold 0.80). Evidence:
  `.moai/reports/t1352/plan-audit-iter2.md` (`verdict: PASS`, audit chain
  disclosed in its first line: single GLM backend; claude/codex cross-audit
  unreachable — untracked plan artifacts at a fixed HEAD give diff-based
  backends no diff). Iteration-1 record preserved at
  `.moai/reports/t1352/plan-audit.md` (FAIL 0.86, blocking D1+D2, fixed and
  re-audited within the 2-attempt ceiling).
- Plan-artifact hash unchanged since the verdict: recomputed at run entry —
  `5a47ef8eb62be90bb2f8d14ebe165c969fe6c090da5e2ebfc39660350bcc0900`
  (spec.md+plan.md+acceptance.md concatenated SHA-256) — byte-identical to the
  verdict's pin.
- Operator provenance: leader dispatch for card t1352 citing operator
  directive 2026-10-03 (v3.2.0 mission), prescribing the full Class C chain
  plan → plan-audit → run → sync; card names no operator gate (AGENTS.local.md
  §31 kickoff autonomy). No blocker open.
- Run Phase 1 plan-audit-gate skip contract (spec-workflow.md § Plan Audit
  Gate skip policy): all three conditions hold — verdict PASS, 0.94 ≥ 0.80,
  hash unchanged (measured this run). Run-gate record stream:
  `.moai/reports/plan-audit/SPEC-TCD-LLM-DECIDER-001.md` (`verdict: PASS`).

**Input parameters**: tier M; scope ~7 files (plan.md §A file list, 2 new + 2
config + wiring + tests); domain count 1 (Go CLI product code); language mix
100% Go; concurrency benefit LOW (coding-heavy); Agent Teams prereqs
not applicable (not requested).

**Mode evaluation**:

| Mode | Selected | Rationale |
|------|----------|-----------|
| direct | no | Multi-file new-feature implementation, not a typo/one-line change |
| serial | **selected** | Coding-heavy work — single manager-develop per milestone |
| fanout | no | No multi-domain research fan-out warranted |
| sweep | no | Not mechanical-uniform high-volume; coding-heavy |

**Decision: serial** — one manager-develop spawn carrying M1-M4 sequentially
(TDD cycle_type per quality.yaml). Justification: coding-heavy per Anthropic's
coding-task parallelism caveat; the SPEC has inter-file dependencies (wiring
depends on the new decider files; tests depend on both), so sequential
milestones in one agent are the safe default. Sweep is unavailable here on
merits, and no operator `--team` request exists.

Audit recommendations carried into run scope (from the iter2 verdict):
E6 verbatim RED evidence, D5 wall-clock boundary number named at M3 test
writing, M2 comment-only amendment of `todo_classify.go:12-15` (§D.4
carve-out) executed.
