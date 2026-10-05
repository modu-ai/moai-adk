# SPEC-CODEX-TEST-GAPS-001 — Progress

## §E.1 Plan-phase Audit-Ready Signal

```yaml
phase: plan
spec: SPEC-CODEX-TEST-GAPS-001
status: draft
tier: M
cycle_type_recommendation: ddd
harness: standard
authored: 2026-09-07
fix_rounds:
  - "v0.2.0 (2026-09-07): plan-audit iter-1 FAIL 0.875 -> D1 tier S->M, D2 non-vacuous test selectors, D3 base-pinned union diff gate, D4 pid un-skipped + REQ-CTG-012 + zero-0.0% retarget, D5 REQ-CTG-011 quality gates, D6 background-run recipe note, D7 census denominators"
worktree: /Users/goos/MoAI/moai-adk-go/.claude/worktrees/t501
branch: WT-codex-uncovered
base: origin/develop @ ace1c5440
id_check: "PASS (regex ^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$, verbatim Bash output)"
uniqueness_check: "PASS (no SPEC-CODEX-TEST-GAPS in catalog; 16 SPEC-CODEX-* siblings)"
premise_verification: "8/8 named functions located at cited files (grep -n, this run, this tree); pid 3-branch body re-read at mcp_codex.go:494-499 in fix round"
refuted_premise_recorded: true
evidence_paths:
  - .moai/reports/t501/namegrep-counts.txt
  - .moai/reports/t501/coverage-perfunc.txt
  - .moai/reports/t501/coverage-run.log
```

## §E.2 Run-phase Evidence

### Pre-flight (Section C)

- Tree: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t501` — `git rev-parse --show-toplevel` resolved to the t501 worktree; `git branch --show-current` = `WT-codex-uncovered`; `git rev-parse --short HEAD` = `e652acf40` at entry (this run, this tree).
- Baseline lint: `golangci-lint run internal/cli/...` → `0 issues.` (rc=0).
- Evidence dir `.moai/state/verify/t501/` created; the three plan-phase evidence files re-read from `.moai/reports/t501/`.
- Plan-audit gate skip taken per the authoritative skip contract: iter-2 verdict PASS 1.00 ≥ Tier M 0.80; artifact hash unchanged (only commits after the audited tree `24df2ae45` are cd855f296 audit report + e652acf40 §F mode log — neither is a hash subject). Kickoff approval granted 2026-09-07.

### M1 — TestTerminateCodexProcess (REQ-CTG-001)

- File: `internal/cli/codex_job_control_test.go` (+78 lines). Real body driven directly (seam NOT swapped). Helper-process re-exec pattern per plan.md §F M1 locked decision; sentinel env `T501_TERMINATE_HELPER_CHILD`; child kill registered via `t.Cleanup` BEFORE the first assertion.
- One iteration during authoring: the success-arm assertion first used `ProcessState.Exited()`, which is FALSE for signal-killed children on darwin/linux; fixed to the cross-platform `errors.As(err, &exec.ExitError)` observation (child gone = contract). RED evidence of that authoring step: `codex_job_control_test.go:736: child did not exit from the termination: err=signal: killed` — assertion defect, not a production defect.

### M2/M3/M5/M6 — mcp_codex_test.go (+212 lines across 3 files)

- `TestCodexIDMatches` 5-arm table; `TestAwaitCodexResponse` canceled-context arm; `TestCodexSessionError` via the production fail-open shape `codexHandshakeFailure`; `TestRealCodexConnPid` 3 direct constructions, no subprocess.
- M4 `TestCodexCountExecutingImports` in `codex_contract_test.go` (HTML-comment shapes). M7 `TestCodexGatePrintf` + `TestDefaultCodexInitGenerator` in `codex_init_test.go` (whitelist-refusal fixture: existing `.codex/hooks.json` carrying top-level `"version"` trips REQ-CW-003).

### Mutant-RED evidence (E8 — operator HARD requirement)

Each mutant was applied to the production body, the scoped selector run captured RED, and the mutant was FULLY reverted (`git diff --stat` / `git status --short` showing only test files) before GREEN was re-observed. All mutants are reverted in the final tree; AC-CTG-009 union is empty.

| Test | Mutant (temporary, reverted) | Verbatim RED output (abridged to the FAIL lines) |
|---|---|---|
| TestTerminateCodexProcess | `codex_job_control.go` refusal arm → `return nil` | `codex_job_control_test.go:705: terminateCodexProcess(-1) = nil error, want the ownership refusal` · `:705: terminateCodexProcess(0) = nil error, want the ownership refusal` · `--- FAIL: TestTerminateCodexProcess (0.00s)` |
| TestCodexIDMatches | `mcp_codex.go` `return n == want` → `n != want` | `mcp_codex_test.go:576: integer id match: codexIDMatches(7, 7) = false, want true` · `:576: integer id mismatch: codexIDMatches(8, 7) = true, want false` · `--- FAIL: TestCodexIDMatches (0.00s)` |
| TestAwaitCodexResponse | `mcp_codex.go` ctx.Err() check dropped | `mcp_codex_test.go:601: awaitCodexResponse with a canceled context = (msg {ID:[57 57] ...}, err <nil>), want the context.Canceled error` · `--- FAIL: TestAwaitCodexResponse (0.00s)` |
| TestCodexSessionError | `Error()` → `return e.summary` | `mcp_codex_test.go:620: Error() = "codex initialize rejected", want the cause verbatim "codex initialize write failed: broken pipe"` · `--- FAIL: TestCodexSessionError (0.00s)` |
| TestRealCodexConnPid | nil-arm disjunction order swapped | `--- FAIL: TestRealCodexConnPid (0.00s)` + `panic: runtime error: invalid memory address or nil pointer dereference [recovered, repanicked]` |
| TestCodexCountExecutingImports | comment-close arm dropped | `codex_contract_test.go:429: multi-line comment — directive after the --> close: codexCountExecutingImports = 0, want 1` · `--- FAIL: TestCodexCountExecutingImports (0.00s)` |
| TestCodexGatePrintf | silent-drop error arm → `panic(err)` | `--- FAIL: TestCodexGatePrintf (0.00s)` + `panic: disk full (t501 fixture)` |
| TestDefaultCodexInitGenerator | wrap → bare pass-through | `codex_init_test.go:1074: the generator must WRAP the cause, not pass it through bare; got "codex wiring refused: rendered hooks.json fails the whitelist: unaccepted top-level key \"version\""` · `--- FAIL: TestDefaultCodexInitGenerator (0.00s)` |

### E1 — AC matrix (verbatim selector runs, this run, tree `1269e6d8d`)

| AC | Status | Command | Observed output |
|----|--------|---------|-----------------|
| AC-CTG-001 | PASS | `go test -count=1 -run 'TestTerminateCodexProcess' -v ./internal/cli/` | `--- PASS: TestTerminateCodexProcessHelper (0.00s)` · `--- PASS: TestTerminateCodexProcess (0.00s)` · `ok github.com/modu-ai/moai-adk/internal/cli 0.856s` |
| AC-CTG-002 | PASS | `go test -count=1 -run 'TestCodexIDMatches' -v ./internal/cli/` | `--- PASS: TestCodexIDMatches (0.00s)` · `ok ... 0.709s` |
| AC-CTG-003 | PASS | `go test -count=1 -run 'TestAwaitCodexResponse' -v ./internal/cli/` | `--- PASS: TestAwaitCodexResponse (0.00s)` · `ok ... 0.682s` |
| AC-CTG-004 | PASS | `go test -count=1 -run 'TestCodexCountExecutingImports' -v ./internal/cli/` | `--- PASS: TestCodexCountExecutingImports (0.00s)` · `ok ... 0.683s` |
| AC-CTG-005 | PASS | `go test -count=1 -run 'TestCodexGatePrintf' -v ./internal/cli/` | `--- PASS: TestCodexGatePrintf (0.00s)` · `ok ... 0.769s` |
| AC-CTG-006 | PASS | `go test -count=1 -run 'TestDefaultCodexInitGenerator' -v ./internal/cli/` | `--- PASS: TestDefaultCodexInitGenerator (0.00s)` · `ok ... 0.693s` |
| AC-CTG-007 | PASS | `go test -count=1 -run 'TestCodexSessionError' -v ./internal/cli/` | `--- PASS: TestCodexSessionError (0.00s)` · `ok ... 0.998s` |
| AC-CTG-008 | PASS | `sed -n '/^## D. Documented-Skip Record/,/^## E/p' spec.md` | Exactly the two rows (`writeCodexRequest`/`writeCodexEnvelope` marshal arms; `terminateCodexProcess` FindProcess arm) + the v0.2.0 `(realCodexConn).pid` removal note; spec.md untouched by this run |
| AC-CTG-009 | PASS | `git diff --name-only 24df2ae45..HEAD` UNION `git status --short`, filtered to non-test `.go` | EMPTY (base SHA cited verbatim: `24df2ae45`); union contains only the plan-audit report, this progress.md, and the 4 `_test.go` files |
| AC-CTG-010 | PASS | `go tool cover -func=.moai/state/verify/t501/cover-after.out` (extract persisted at `.moai/state/verify/t501/cover-after-perfunc.txt`) | `terminateCodexProcess 75.0%` (FindProcess arm = documented skip) · `(realCodexConn).pid 100.0%` · `codexSessionError Error/Unwrap 100.0%` · `codexIDMatches 100.0%` (>60%) · `awaitCodexResponse 100.0%` · `codexCountExecutingImports 100.0%` · `codexGatePrintf 100.0%` · `defaultCodexInitGenerator 100.0%` — ZERO 0.0% rows among the 118 baseline function rows |
| AC-CTG-011 | PASS | `gofmt -l internal/cli/` (empty) · `go vet ./internal/cli/` (rc=0) · `golangci-lint run internal/cli/...` → `0 issues.` | see §E.3 |
| AC-CTG-012 | PASS | `go test -count=1 -run 'TestRealCodexConnPid' -v ./internal/cli/` | `--- PASS: TestRealCodexConnPid (0.00s)` · `ok ... 0.684s` |

### E2 — cross-platform test compile

`GOOS=windows go test -c -o /dev/null ./internal/cli/` → exit 0 (`windows-test-compile-ok`). Note observed and honored: plain `go build` does not compile `_test.go`, so the test-binary compile is the E2 gate.

### E4 — subagent-boundary grep

`grep -rn 'AskUserQuestion\|mcp__askuser' <4 touched test files>` → zero matches (rc=1). Tests-only scope; no non-test file touched.

### E5 — lint

Baseline (pre-work): `golangci-lint run internal/cli/...` → `0 issues.` Post-work: `0 issues.` NEW issues introduced: 0.

### E6 — git state

- Commits this run: `ffa06117f` (M1), `1269e6d8d` (M2-M7), plus this progress commit. HEAD at measurement: `1269e6d8d`.
- `git fetch origin develop` + `git rev-list --count origin/develop..HEAD` → 6 (2 this SPEC + 4 pre-existing lane commits). NOT pushed — lane push prohibition (lead batch-pushes develop).

### M8 — full scoped run (E1 recipe, background task)

```
unset MOAI_KANBAN ... && go test -count=1 -timeout 700s -coverprofile=.moai/state/verify/t501/cover-after.out ./internal/cli/ > .moai/state/verify/t501/cover-after.log 2>&1; rc appended
```
Log (verbatim): `ok  	github.com/modu-ai/moai-adk/internal/cli	337.374s	coverage: 80.8% of statements` / `rc=0`. (Baseline 515.6s / 80.7% — machine load differs between runs; the coverage delta +0.1pt is the SPEC's own work.)

### Gaps and residual notes (VCI §3)

- GAP: none among the 12 ACs — every row above carries command + observed output from this run.
- Residual: the fresh extract contains ONE 0.0% function OUTSIDE the AC-CTG-010 denominator — `codex_review_gate.go:183 runCodexReviewGate 0.0%`. `codex_review_gate.go` predates this SPEC (last touched #1430) and its 9 rows are absent from the baseline 118-row extract the AC pins (`grep -c codex_review_gate .moai/reports/t501/coverage-perfunc.txt` = 0), so AC-CTG-010's "zero among the 118" holds as written; the omission of that file from the baseline denominator is a plan-phase scoping observation for the lead, not a defect introduced here.
- B8 hygiene note: `git status --short` after all test runs showed NO unexpected fixture rewrites (perf-benchmark fixtures untouched).

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-09-07
run_commit_sha: "1269e6d8d"   # HEAD the evidence was measured at; the progress commit itself backfills after
run_status: complete
ac_pass_count: 12
ac_fail_count: 0
preserve_list_post_run_count: 0   # zero non-test .go files modified (AC-CTG-009 union empty at base 24df2ae45)
l44_pre_commit_fetch: "git fetch origin develop run before final count (FETCH_HEAD updated 2026-09-07)"
l44_post_push_fetch: "N/A — lane never pushes (lead batch-pushes develop)"
new_warnings_or_lints_introduced: 0
cross_platform_build:
  windows_test_compile: "GOOS=windows go test -c -o /dev/null ./internal/cli/ → exit 0"
  note: "plain go build excludes _test.go; the test-binary compile is the gate"
total_run_phase_files: 4
m1_to_mn_commit_strategy: "M1 solo commit (ffa06117f) + M2-M7 grouped commit (1269e6d8d) + progress.md commit; mutants all reverted before every commit"
evidence_dir: .moai/state/verify/t501/
```


## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: 2026-09-07
sync_commit_sha: "82674b94c"   # backfilled from pending-backfill-sync (D3 exemption); sync commit 82674b94c
sync_status: complete
b12_self_test_a: "grep -c 'SPEC-CODEX-TEST-GAPS-001' CHANGELOG.md → 0 (pre-emission); entry appended once, then 1"
b12_self_test_b: "distinct AC ids in acceptance.md = 12 (AC-CTG-001..012); CHANGELOG entry references the same 12"
b12_self_test_c: "every path in the CHANGELOG entry verified to exist (spec.md, .moai/state/verify/t501/, .moai/reports/t501/, 4 _test.go files)"
changelog_entry_position: "[Unreleased] → Added, first entry"
frontmatter_status_transitions:
  spec_md: "in-progress → implemented → completed (merged close on the sync commit)"
  updated_field: "refreshed to 2026-09-07 (sync commit date)"
canary_compliance_check:
  readme_sync: "N/A — tests-only, no user-facing behavior change"
  docs_site_sync: "N/A — same rationale"
  mx_tags: "no obligations — zero production code added or modified (AC-CTG-009 union empty of non-test .go)"
union_gate_post_sync: "git diff --name-only 24df2ae45..HEAD filtered to non-test .go — EMPTY (verified after this commit)"
sync_close_last_write: "git status --short empty after the sync commit"
```

## §F Phase 4 Mode Selection

```yaml
phase: run
logged_at: 2026-09-07
logged_after: "Implementation Kickoff Approval obtained 2026-09-07 (operator, via lead-session AskUserQuestion; lead relayed with HEAD cd855f296 measured)"
input_parameters:
  tier: M
  scope_files: "<5 test files extended (codex_job_control_test.go, mcp_codex_test.go, codex_contract_test.go, codex_init_test.go)"
  domain_count: 1
  file_language_mix: "100% Go test code, single package internal/cli"
  concurrency_benefit: "LOW — coding-heavy (Anthropic coding-task parallelism caveat); milestones M1+M2 both edit-adjacent test files, fan-out would create write contention on shared test files"
  agent_teams_prereqs: "not requested (no operator --team)"
mode_evaluation:
  direct: "not selected — 8 test items across 4+ files exceeds a single-response edit"
  serial: "selected — one manager-develop carries M1-M8 in audited order with the mutant-RED discipline interleaved"
  fanout: "not selected — coding-heavy work, 1 domain, <10 files; shared test files would collide across concurrent writers"
  sweep: "not selected — semantic test authoring, not mechanical-uniform; <30 files"
  agent-team: "not selected — explicit-request-only; no request"
decision: serial
justification: >
  Tests-only work confined to one package with per-milestone edits landing in
  shared test files makes concurrency a write-conflict risk rather than a speed
  win. The audited milestone order (M1=U3 terminateCodexProcess, M2=U2
  codexIDMatches, M3=U2 ctx-cancel arm) already fronts the card's priority
  clusters, so a single sequential spawn preserves both the priority directive
  and the mutant-RED evidence chain (each test shown RED under a named mutant
  before GREEN on the pristine tree). Serial is the default fallback for
  coding-heavy work per the Anthropic parallelism caveat.
boundary_case: "none — no threshold-adjacent inputs"
sweep_confirmation: "N/A (sweep not selected)"
```

