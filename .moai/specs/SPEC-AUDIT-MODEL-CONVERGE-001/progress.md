# SPEC-AUDIT-MODEL-CONVERGE-001 — progress

## §E.1 Plan-phase Audit-Ready Signal

```yaml
plan_status: audit-ready
plan_complete_at: 2026-10-02T05:56:24Z
amended_at: 2026-10-02T06:10:42Z         # 0.1.1 — operator decisions D4-D6 applied (spec.md HISTORY)
amended_again_at: 2026-10-02T06:54:52Z   # 0.1.2 — plan-audit iteration 1 amendment (PA1-D1..D11), leader rulings D7'-D11
amended_final_at: 2026-10-02T07:29:41Z   # 0.1.3 — plan-audit iteration 2 final revision (PA2-D1..D9), decisions D12-D13: scope reduced
card: t1423
tier: L
artifacts: [spec.md, plan.md, acceptance.md, design.md, research.md, progress.md]
authored_at_head: c50da9c2f
amended_at_head: 739556db5
counts: { requirements: 20, acceptance_criteria: 20, milestones: 7, run_phase_files: ~36 }
spec_lint: "moai spec lint --strict SPEC-AUDIT-MODEL-CONVERGE-001 (after the 0.1.3 revision) -> exit 0, "No findings — all SPEC documents are valid"; judging build v3.2.0-rc.24 gc50da9c2f"
plan_audit:
  iteration_1: { verdict: FAIL, score: 0.81, threshold: 0.85, report: .moai/reports/t1423/plan-audit-iter1.md, audited_sha: 53a42f013a45d376ba3e4289af677bc878d6f9a4 }
  iteration_2: { verdict: FAIL, score: 0.81, threshold: 0.85, report: .moai/reports/t1423/plan-audit-iter2.md, audited_sha: 739556db5684a79feda06cad2b92d18837274fee }
  iteration_3: pending                   # the cap; escalate per the Retry Loop Contract if it does not pass
flagged_assumptions: [OQ-3, OQ-6, OQ-8, OQ-9, OQ-11, OQ-12]
closed_open_questions: [OQ-1 (D7' confirmed by the leader), OQ-2 (D5), OQ-4 (REQ-ACV-012/-015), OQ-5 (D10 override), OQ-7 (D6), OQ-10 (D12 pure checker)]
open_clarifications: []
amendment_0_1_6_note: "leader-approved final extension 10-02: NEW-3 sentence on supplied gates at enforcement (spec 0.1.6); NEW-1 receipt-gate repair is code only"
amendment_0_1_5_note: "leader-approved amendment 10-02 after the audit ceiling; checker input is --result-file; inline --result removed (spec 0.1.5; plan.md gains the M4b row)"
iteration_4_note: "iteration 4 = delta confirmation over the Tier L ceiling of 3, leader-approved 10-02 (same criterion as t1411); D1 and D2 repaired. Carry-over: DB1-DB8 of plan-audit-iter3 to be checked for resolution at the sync stage — DB1 legacy-window-fail-open-by-design; DB2 full-result-form-invites-shell-injection; DB3 reader-fix-reaches-review-gate-and-codex_task; DB4 instruction-text-criteria-are-shallow; DB5 signature-brittleness; DB6 startup-regular-file-effects-unmeasured; DB7 codex-hosted-auditor-cannot-run-the-verb; DB8 req-014-vs-design-d8-sync-legacy-wording"
```

## §E.2 Run-phase Evidence

### M1 — baseline-first characterization and the TestMain scrub

Measured against base HEAD `25e6e9737` (branch `WT-audit-model-convergence`, worktree `.moai/worktrees/t1423`); the M1 commit adds only test files, the golden, the `spec.md` status line and this block. Every Go run is `unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED [CLAUDE_PROJECT_DIR] && go test ...` in one invocation. Judging toolchain: `golangci-lint v2.1.6` (the CI pin); no installed `moai` build was used for any measurement.

Pre-flight (before any change): `go build ./...` exit 0; `GOOS=windows GOARCH=amd64 go build ./...` exit 0; `golangci-lint run --timeout=2m ./internal/cli/...` exit 0, `0 issues.`; `grep -c "CLAUDE_PROJECT_DIR" internal/cli/main_test.go` -> `0` (E31 reproduced).

AC-ACV-019 RED (test added, scrub line NOT yet added):

    $ CLAUDE_PROJECT_DIR=<tree> go test -count=1 -v -run '^TestMain_ScrubsClaudeProjectDir$' ./internal/cli/
    --- FAIL: TestMain_ScrubsClaudeProjectDir (0.00s)
    FAIL	github.com/modu-ai/moai-adk/internal/cli	1.222s
    exit=1

AC-ACV-019 GREEN (one `os.Unsetenv(config.EnvClaudeProjectDir)` line added in `TestMain`):

    $ CLAUDE_PROJECT_DIR=<tree> go test -count=1 -v -run '^TestMain_ScrubsClaudeProjectDir$' ./internal/cli/
    --- PASS: TestMain_ScrubsClaudeProjectDir (0.00s)
    ok  	github.com/modu-ai/moai-adk/internal/cli	0.923s
    exit=0
    $ grep -c "EnvClaudeProjectDir" internal/cli/main_test.go   -> 3

AC-ACV-008 GREEN on arrival (golden generated from the unmodified production handler, then compared without the update toggle; the swept set is 2 tests x 2 subcases, no `[no tests to run]`):

    $ go test -count=1 -v -run '^(TestAuditMulti_NoConfigNoArgs_ByteIdentical|TestAuditMulti_PinsOnlyConfig_ByteIdentical)$' ./internal/cli/
    --- PASS: TestAuditMulti_NoConfigNoArgs_ByteIdentical (0.01s)
    --- PASS: TestAuditMulti_PinsOnlyConfig_ByteIdentical (0.01s)
    ok  	github.com/modu-ai/moai-adk/internal/cli	1.035s
    exit=0

Mutant probe on the golden: changing one summary in the `pins-only/all-pass` entry makes `TestAuditMulti_PinsOnlyConfig_ByteIdentical/all-pass` fail (exit 1); the entry was restored and the pair re-run green.

Post-change: `GOOS=windows GOARCH=amd64 go build ./...` exit 0; `go vet ./internal/cli/` exit 0; `golangci-lint run --timeout=2m ./internal/cli/...` exit 0, `0 issues.`; `gofmt -l` on both test files empty; `grep -rn 'AskUserQuestion\|mcp__askuser__'` on both test files: no output, exit 1.

Golden facts: the golden is one JSON object keyed `<root>/<case>` (4 entries); the update toggle is `UPDATE_AUDIT_MULTI_DEFAULT_GOLDEN=1` (per-file env toggle, the repo's existing golden convention) and merges the cases a run covers. `build_commit`, `build_lag` and `tree_root` (asserted equal to the fixture root first) are dropped by the test, not stored. The all-pass and codex-inconclusive cases both resolve to `overall_verdict: pass` with no `plan_source` and no `gate_unmet` member.

Not covered at M1 (a Gap, not a pass): the clause of AC-ACV-008 that the plan resolved for each root is the default plan with every entry `explicit: false` cannot be asserted before the resolver exists (M2); that assertion is to be added with M2 and kept by M3-M7. The full `internal/cli` suite was not run (heavy-run rule); CI runs it on push.

### M2 — audit plan resolver (`internal/config/audit_plan.go`)

Measured against base HEAD `dd44df3cf` (the M1 commit), branch `WT-audit-model-convergence`, worktree `.moai/worktrees/t1423`; the M2 commit adds `internal/config/audit_plan.go`, `internal/config/audit_plan_test.go`, one added assertion in `internal/cli/mcp_audit_multi_baseline_test.go`, and this block (no production code in `internal/cli`, golden untouched, `UPDATE_AUDIT_MULTI_DEFAULT_GOLDEN` never set). Every Go run is `unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED CLAUDE_PROJECT_DIR && go test ...` in one invocation; judging toolchain `golangci-lint v2.1.6` (the CI pin); no installed `moai` build was used.

Pre-flight: `go build ./...` exit 0; `GOOS=windows GOARCH=amd64 go build ./...` exit 0; `golangci-lint run --timeout=2m ./internal/config/...` exit 0, `0 issues.`; `ls internal/config/audit_plan.go` -> `No such file or directory`, exit 1; `grep -rn "audit_plan" internal/config` -> no output, exit 1 (RED-now: no resolver).

RED, observed in two forms. (a) Tests only, no resolver: `go test ./internal/config -run '^TestResolveAuditPlan_' -count=1` -> `internal/config/audit_plan_test.go:79:44: undefined: AuditPlan` ... `too many errors`, exit 1 — a compile failure, a TOOL_FAILURE and not a RED. (b) Semantic RED against a throwaway stub (types only, `ResolveAuditPlan` returns a zero plan, no behaviour):

    $ go test ./internal/config -run '^TestResolveAuditPlan_' -count=1 -v
    --- FAIL: TestResolveAuditPlan_TokenTable (0.00s)        # "plan has 0 backend entries, want 3: []" x5 sub-tests
    --- FAIL: TestResolveAuditPlan_Precedence (0.00s)
    --- FAIL: TestResolveAuditPlan_PrecedenceExample (0.00s)
    --- FAIL: TestResolveAuditPlan_ExplicitGates (0.00s)
    --- FAIL: TestResolveAuditPlan_FromConfig (0.00s)
    --- FAIL: TestResolveAuditPlan_RejectsUnknown (0.00s)    # 7 sub-tests: no error returned
    --- FAIL: TestResolveAuditPlan_TrimsWhitespace (0.00s)
    --- PASS: TestResolveAuditPlan_SourceIsPure (0.00s)      # guard, green against the stub by design
    FAIL	github.com/modu-ai/moai-adk/internal/config	0.375s
    exit=1

GREEN (real resolver, same command): 8 `--- PASS:` lines (`TokenTable`, `Precedence`, `PrecedenceExample`, `ExplicitGates`, `FromConfig`, `RejectsUnknown`, `TrimsWhitespace`, `SourceIsPure`), 22 sub-test `--- PASS:` lines, `ok  github.com/modu-ai/moai-adk/internal/config  0.414s`, exit 0. Whole package: `go test ./internal/config -count=1` -> `ok ... 11.947s`, exit 0.

Positive controls inside the sweeps: token table asserts 5 tokens swept; precedence sweep asserts 5 tokens x 3 backends x 4 caller choices x 4 configured choices = 240 cases; the purity test asserts the file holds `func ResolveAuditPlan(` and that the import scan saw at least one import.

Mutant probes (resolver edited, tests re-run, resolver restored byte-identical to the committed copy): `claude` token mapped to Claude alone -> `TestResolveAuditPlan_TokenTable` and `_Precedence` FAIL; argument and config.gates swapped in the ladder -> `_Precedence` FAIL; the empty token marks codex explicit -> `TestAuditMulti_PinsOnlyConfig_ByteIdentical` FAIL with `pins-only: codex = {... Source:config.model Explicit:true}, want ... source "default", explicit false`.

AC-ACV-008 clause left as a Gap at M1 is closed: `internal/cli/mcp_audit_multi_baseline_test.go` gains `assertDefaultAuditPlan`, called once per root by `runAuditMultiBaselineRoot`, so both named tests assert that the plan resolved from the RAW section (`loadWorkflowAuditSection`, then `config.ResolveAuditPlan` with no caller gates) is the default plan with every entry `explicit: false`, source `default`, gates required / required / advisory, and `ExplicitGates()` empty. Run: `-run '^TestAuditMulti_NoConfigNoArgs_ByteIdentical'` -> `--- PASS` plus `all-pass` and `codex-inconclusive`, `ok`, exit 0; `-run '^TestAuditMulti_PinsOnlyConfig_ByteIdentical'` -> same, exit 0. The end anchor `$` was dropped from both `-run` patterns because the worktree guard refused a command containing `$'` (see Gaps); no other test name shares either prefix.

AC-ACV-004 (resolver part): `grep -c "^package config" internal/config/audit_plan.go` -> `1`, exit 0; `grep -nE "\"(os|time|net|io|path/filepath|os/exec)\"" internal/config/audit_plan.go` -> no output, exit 1; `grep -n "NewDefaultWorkflowConfig\|NewDefaultConfig"` over `audit_plan.go`, `audit_pin.go`, `mcp_audit_multi.go`, `mcp_worktree_root.go`, `internal/auditreceipt/store.go` -> no output, exit 1; `grep -n 'AskUserQuestion\|mcp__askuser__\|NewDefaultWorkflowConfig\|NewDefaultConfig\|internal/cli' internal/config/audit_plan.go` -> no output, exit 1. The same two import/constructor clauses are also pinned by `TestResolveAuditPlan_SourceIsPure`.

Coverage of the new file: `go test ./internal/config -run '^TestResolveAuditPlan_' -coverprofile=<scratch> -count=1` then `go tool cover -func=<scratch>` filtered to `audit_plan.go`: `ModelSource`, `FromConfig`, `ExplicitGates`, `auditPlanCells`, `gateAt`, `normalizeAuditGate`, `ResolveAuditPlan` all `100.0%`.

Post-change: `go build ./...` exit 0; `GOOS=windows GOARCH=amd64 go build ./...` exit 0; `go vet ./internal/config/` exit 0; `go vet ./internal/cli/` exit 0; `golangci-lint run --timeout=2m ./internal/config/... ./internal/cli/` exit 0, `0 issues.`; `gofmt -l` on the three touched Go files empty, exit 0.

Interface decisions M3/M4 consume (spec text is silent on them; flagged for the plan-audit trail): `ResolveAuditPlan(audit AuditConfig, supplied AuditGates) (AuditPlan, error)`; a supplied gate outside the closed set is an error too (`gates.<backend> "<v>" unknown (want one of off|advisory|required)`), whitespace-only values read as unset, a token or gate is trimmed before matching and matching is case-sensitive; on error the returned plan is the zero value. `AuditPlan.ModelSource()` is `config` when the token is non-empty and `default` otherwise, `AuditPlan.FromConfig()` is the `plan_source: "config"` predicate. The gate error key text for a configured value is `audit.gates.<backend>`; the REQ-ACV-003 token text is pinned as `audit_model "<v>" unknown (want one of claude|codex|glm|multi)`.

Not covered at M2 (a Gap, not a pass): the full `internal/cli` suite was not run (heavy-run rule; CI runs it on push); no consumer of the resolver exists yet (M3/M4), so AC-ACV-005 stays open and `internal/auditreceipt` / `audit_multi` still read gates as before; `TestResolveAuditPlan_RejectsUnknown` pins the resolver half of AC-ACV-003 only (the `audit_multi` and verb halves are M3/M4).

### M3a — config-plan consumers and the receipt predicate

Measured against base HEAD `5bc7c15c6` (the M2 commit), branch `WT-audit-model-convergence`, worktree `.moai/worktrees/t1423`; the M3a commit changes `internal/cli/{mcp_audit_multi,mcp_convergence,mcp_worktree_root}.go`, `internal/auditreceipt/store.go`, adds `internal/cli/mcp_audit_multi_config_plan_test.go`, extends `internal/auditreceipt/store_test.go`, edits one case of `internal/cli/codex_audit_nonrequired_golden_test.go` (see Gaps) and appends this block. `internal/cli/testdata` is untouched and `UPDATE_AUDIT_MULTI_DEFAULT_GOLDEN` was never set; `mcp_codex.go` is untouched (the codex turn reader and the codex-leg deadline are M3b). Every Go run is `unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED CLAUDE_PROJECT_DIR && go test ...` in one invocation, under the lease `moai slot acquire --resource go-test-internal-cli --max-duration 20m` (acquired, released before this block was written, `moai slot status` -> `free`); judging toolchain `golangci-lint v2.1.6`; no installed `moai` build was used.

Pre-flight on `5bc7c15c6`: `go build ./...` exit 0; `GOOS=windows GOARCH=amd64 go build ./...` exit 0; `golangci-lint run --timeout=2m ./internal/cli/... ./internal/auditreceipt/...` exit 0, `0 issues.` (baseline); M1 pair `-run '^(TestAuditMulti_NoConfigNoArgs_ByteIdentical|TestAuditMulti_PinsOnlyConfig_ByteIdentical)'` -> two `--- PASS:`, `ok`, exit 0.

RED (semantic: every test reached its assertion; no compile failure), the new `internal/cli` tests against the unchanged handler:

    --- FAIL: TestAuditMulti_ConfigPlanFallback        # "codex alone: backends invoked = [claude codex glm], want [codex]"; plan_source = <nil>, want "config" (6 sub-tests)
    --- FAIL: TestAuditMulti_ModelMulti_CodexUnavailable_FailsClosedNamed   # "overall_verdict = pass, want fail"; "gate_unmet = <nil>, want \"codex\"" (6 sub-tests; the unconfigured-sibling sub-test PASSes)
    --- FAIL: TestAuditMulti_ExplicitClaudeToken       # plan_source = <nil>, want "config"
    --- FAIL: TestAuditMulti_UnknownConfiguredTokenIsToolError   # "IsError = false, want a tool error" (3 sub-tests)
    --- FAIL: TestAuditMulti_PaddedConfiguredGateEnforcesLikeTheTrimmedValue   # "overall_verdict = pass, gate_unmet = <nil>, want fail / codex"
    --- FAIL: TestAuditMulti_InvalidSuppliedGateIsNotAToolError  # `codex gate = "requird", want the default "required"`
    --- FAIL: TestAuditMulti_ModelMultiRecordsReceiptOnUnmetGate # `gate_unmet = <nil>, want codex (precondition)`
    --- FAIL: TestCodexAudit_ModelCodexUnmetGateFails  # `verdict = "inconclusive", want fail`
    --- PASS: TestAuditMulti_ArgsWinOverConfig, TestAuditMulti_ArgumentRequiredIsNotExplicit   # guards, green on arrival by design (only arguments were ever read)
    FAIL	github.com/modu-ai/moai-adk/internal/cli	1.466s     exit=1

RED for the receipt predicate (new test run against `git show HEAD:internal/auditreceipt/store.go`, then the edited store restored): `TestCodexGateRequired_FromModelToken` FAIL for `multi`, `codex`, `padded token` (`CodexGateRequired(...) = false, want true`; the other 11 rows PASS), `TestCodexGateRequired_ExactMatchOnly` FAIL on the padded-gate row, exit 1.

What today's handler did with an INVALID supplied gate (measured by the RED above): the string was passed through untouched (`codex gate = "requird"`); the backend was invoked (it is not `off`), and the value was neither `required`, `advisory` nor `off`, so the convergence engine treated the entry as none of the three. After M3a an invalid or non-string supplied value reads as NOT SUPPLIED (`suppliedGate`), so that backend takes its gate from the tree's plan or the default — still no error. A whitespace-padded config gate (` required `) used to reach the explicit-gate check untrimmed and never matched `required` (silently unenforced); the resolver trims, so every surface now reads it as `required`.

GREEN: `go test ./internal/cli -run '^(TestAuditMulti_ConfigPlanFallback|...|TestCodexAudit_PaddedRequiredGateFailsClosed|TestCodexAudit_NonRequiredGateGoldenByteIdentical|TestAuditMulti_NoHardErrorPath_AC_AMM_024|TestConvergence_NoNewAuditModelEnum|TestConvergence_NoDirectFrontmatterRead|TestAuditMulti_NoConfigNoArgs_ByteIdentical|TestAuditMulti_PinsOnlyConfig_ByteIdentical)' -count=1 -v` -> 17 top-level `--- PASS:` lines, `ok  github.com/modu-ai/moai-adk/internal/cli  1.684s`, exit 0 (includes the M1 baseline pair). `go test ./internal/auditreceipt -run '^(TestCodexGateRequired_FromModelToken|TestCodexGateRequired_ExactMatchOnly)' -v` -> two `--- PASS:`, `ok`, exit 0; `go test ./internal/auditreceipt ./internal/config -count=1` -> both `ok`, exit 0. Existing families: `-run '^(TestAuditMulti_.*|TestRunMultiAudit_.*|TestConverge_.*|TestConvergence_.*|TestCodexAudit.*|TestCodexTurnReview_.*|TestMain_.*|TestWSR0.*)' ./internal/cli` -> `ok  ... 140.941s`, exit 0; `go test ./internal/hook -run '^(TestAuditReceiptGuard|TestServedModel|TestWSR00[789]|TestSubagentStop|TestSubagentStart|TestPreToolUse)'` -> `ok ... 25.477s`, exit 0; `go test ./internal/template -run '^TestClaudeAuditTemplateSurfacesAndCatalogHash'` -> `--- PASS`, `ok`, exit 0.

Structural: `grep -rl --exclude="*_test.go" "ResolveAuditPlan(" internal` -> `internal/config/audit_plan.go`, `internal/cli/mcp_worktree_root.go`, `internal/cli/mcp_audit_multi.go`, `internal/auditreceipt/store.go`, exit 0 (`audit_plan_cmd.go` is M4); `grep -c "internal/config" internal/auditreceipt/store.go` -> `1`; `go list -deps ./internal/config | grep -c 'moai-adk/internal/auditreceipt'` -> `0` (no import cycle); `grep -n "NewDefaultWorkflowConfig\|NewDefaultConfig"` over the five consumer files -> no output, exit 1; new `AskUserQuestion` / `mcp__askuser__` occurrences: none (the four matches in `mcp_audit_multi.go` / `mcp_convergence.go` are pre-existing comments).

Post-change: `go build ./...` exit 0; `GOOS=windows GOARCH=amd64 go build ./...` exit 0; `go vet ./internal/cli ./internal/auditreceipt` exit 0; `golangci-lint run --timeout=2m ./internal/cli/... ./internal/auditreceipt/...` exit 0, `0 issues.` (no new finding vs baseline); `gofmt -l` on the seven touched Go files empty. Coverage (`-coverprofile` over `^(TestAuditMulti_.*|TestRunMultiAudit_.*|TestCodexAudit_.*|TestWSR0.*)`): `handleAuditMulti`, `readGatesArgument`, `suppliedGate`, `planGates`, `auditSectionForRoot`, `resolveAuditGates` 100.0%; `runMultiAudit` 95.0%; `go test ./internal/auditreceipt -coverprofile`: `CodexGateRequired` 100.0%, `rawCodexGate` 92.3%, package 88.1%.

Gaps: the one change outside the file list the delegation named is a deviation forced by the SPEC's own rule (EC-1: token and gate values match after trimming): `internal/cli/codex_audit_nonrequired_golden_test.go` pinned `codex: "required "` as a NON-required state (golden `testdata/codex-audit-nonrequired/required-with-trailing-space.golden`), and `TestCodexGateRequired_ExactMatchOnly` pinned the same padded row as false (an earlier SPEC's "a trimmed match would turn a typo into an enforcement" decision — the SPEC text does not mention it). The store test row was flipped in place and the golden-test case was removed (the golden file itself was left untouched and is now an orphan; deleting it is a leader decision because `internal/cli/testdata` was to stay unchanged); the padded reading is pinned by `TestCodexAudit_PaddedRequiredGateFailsClosed` and `TestAuditMulti_PaddedConfiguredGateEnforcesLikeTheTrimmedValue`. `rawCodexGate` now returns "" (not required) for an invalid `audit.model` even when `gates.codex` is `required` (the SPEC's "invalid value reads as not required" on this hook path; the old reader ignored `model`). The wider `internal/cli` and `internal/hook` suites were not run (heavy-run rule); CI runs them on push. M3b (codex turn reader, the codex-leg deadline, `DefaultCodexAuditLegTimeout`) is not started.

### M3b — the codex turn reader rule and the codex-leg deadline

Measured against base HEAD `69d5b65b3`, branch `WT-audit-model-convergence`, worktree `.moai/worktrees/t1423`; the M3b commit changes `internal/cli/mcp_codex.go` (only the two non-completed returns of `awaitCodexTurnReview`), `internal/cli/mcp_convergence.go` (only `performCodexAudit`), `internal/config/defaults.go` (the `DefaultCodexAuditLegTimeout` var), adds `internal/cli/codex_turn_incomplete_test.go` and appends this block. `internal/cli/testdata` and `internal/hook` are untouched; `UPDATE_AUDIT_MULTI_DEFAULT_GOLDEN` was never set. Every Go run is one `unset ... CLAUDE_PROJECT_DIR && go test ...` invocation; judging toolchain `golangci-lint v2.1.6`; no installed `moai` build judged any claim.

Pre-flight on `69d5b65b3`: `go build ./...` exit 0; `GOOS=windows GOARCH=amd64 go build ./...` exit 0; `golangci-lint run --timeout=2m ./internal/cli/... ./internal/config/...` exit 0, `0 issues.`; `grep -rn "DefaultCodexAuditLegTimeout\|codexLegTimeout\|codexAuditTimeout" internal` exit 1 (E17 RED-now); `grep -n "context.WithTimeout" internal/cli/mcp_convergence.go` exit 1 (E18); the three `return bestCodexReviewText(reviewText, agentText), nil` lines at 1264, 1268, 1319 (E30).

RED (semantic, every test reached its assertion; the only production line present was the `DefaultCodexAuditLegTimeout` var, added first so the test file compiles):

    --- FAIL: TestCodexTurnReview_StreamClosedBeforeCompletion_Inconclusive   # finding body: verdict = "fail", want "inconclusive"; findings = 1, want 0; clean body: verdict = "pass", want "inconclusive"; error = nil
    --- FAIL: TestCodexTurnReview_ContextEnded_DiscardsPartial                # caller cancellation: verdict = "fail", want "inconclusive"; error = nil; findings = 1
    --- FAIL: TestAuditMulti_CodexLegDeadline_FailsClosedNamed                # audit_multi took 3.0s (the test's safety valve, no deadline existed); codex summary = "codex review output was blank: ...", want a timeout
    --- FAIL: TestAuditMulti_CodexLegDeadline_UnconfiguredFailsOpen           # audit_multi took 3.0s, want well under 1s
    --- PASS: TestCodexTurnReview_CompletedTurnUnchanged, TestAuditMulti_CodexLegDeadline_CallerCancelIsNotATimeout   # controls, green on arrival by design
    FAIL	github.com/modu-ai/moai-adk/internal/cli	7.230s     exit=1

GREEN: the same `-run` -> eleven `--- PASS:` lines (six top-level, including the cancel-vs-timeout control, plus five sub-tests), `ok  github.com/modu-ai/moai-adk/internal/cli  1.442s`, exit 0. The summary of a deadline-ended leg is `codex leg timed out after 20m0s` at the production value; a leg ended by the caller's cancellation never carries it. The existing codex families (`TestCodexTask.*|TestCodexBlankReview.*|TestCodexAudit.*|TestCodexReviewGate.*|TestCharacterize.*|TestRunCodexReviewGate_.*|TestRunCodexReviewRPC.*|TestMCPEOFReaderPreservesFinalBytes`, under the `go-test-internal-cli` lease, acquired and released) -> 173 top-level `--- PASS:` lines (48/16/49/11/41/5/2/1 per family), 0 `--- FAIL`, `ok ... 106.013s`, exit 0: no existing test pinned the partial-text return (OQ-11 holds for the existing families). The M1 baseline pair, `TestAuditMulti_.*`, `TestRunMultiAudit_.*`, `TestConverge_.*`, `TestMain_.*`, the structural guards and the six-case `TestCodexAudit_NonRequiredGateGoldenByteIdentical` -> 64 top-level `--- PASS:`, 0 `--- FAIL`, `ok`; `TestClaudeAuditTemplateSurfacesAndCatalogHash` (in `internal/template`) `--- PASS`. Scoped hook tests `TestReviewGate.*|TestCodexReviewGate.*` -> six `--- PASS:`, `ok ./internal/hook`. Greps: one `return bestCodexReviewText(reviewText, agentText), nil` line (the completed arm, now 1323); `DefaultCodexAuditLegTimeout = DefaultCodexAuditTimeout` at one line; `codexLegTimeout|codexAuditTimeout` -> no output, exit 1.

Post-change: both builds exit 0; `go vet ./internal/cli ./internal/config` exit 0; `golangci-lint ... ./internal/cli/... ./internal/config/...` `0 issues.`; `gofmt -l` on the four touched Go files empty. Coverage (`-coverprofile` over `^(TestCodexTurnReview_.*|TestAuditMulti_CodexLegDeadline_.*)`): every changed block (`mcp_codex.go` 1267-1275, `mcp_convergence.go` 627-636) has count 1.

Gaps: the hook package does not call the turn reader (the Stop-hook gate runs through `internal/cli/codex_review_gate.go` and `runTurn`), so the scoped hook tests exercise the wrappers only; the behaviour change for the gate is measured by the `TestCodexReviewGate*` and `TestRunCodexReviewGate_*` families in `internal/cli`. The full `internal/cli` and `internal/hook` suites were not run (heavy-run rule). The stale doc comment above `awaitCodexTurnReview` ("Returns "" on EOF / deadline (fail-open)") was left alone (outside the permitted lines); the two arms now return an error as well. The deadline tests run a fake session whose `recv` unblocks when its start context ends, as `exec.CommandContext` does for a real process; a live codex was not exercised, so the real kill-then-stdout-close path is inferred from the existing liveness tests, not observed. Behavioural changes to disclose at sync: a Stop-hook review that times out or crashes after a partial finding now fails open (inconclusive) instead of blocking on the partial finding; a `codex_task` whose stream closes mid-turn now reports `failed` with the cause instead of `completed` with partial output. The three deadline-wording and cancel tests live in `codex_turn_incomplete_test.go`, not in `mcp_audit_multi_config_plan_test.go` as plan.md M3 lists them, because the delegation limited the changed test files to the new one.

### M4 — the read-only verb `moai verify audit-plan` and its pure checker

Measured against base HEAD `7998ea35b`, branch `WT-audit-model-convergence`, worktree `.moai/worktrees/t1423`; the M4 commit adds `internal/cli/audit_plan_cmd.go`, `internal/cli/audit_plan_cmd_test.go` and appends this block (the verb registers itself through `verifyExtraCommands` from its own `init()`, so no edit to `verify.go` was needed). `internal/cli/testdata`, `internal/hook`, `internal/mcp` are untouched; `UPDATE_AUDIT_MULTI_DEFAULT_GOLDEN` was never set. Every Go run is one `unset ... CLAUDE_PROJECT_DIR && go test ...` invocation; judging toolchain `golangci-lint v2.1.6`; the only moai build exercised for the verb is one built from this tree into the scratchpad (`go build -o <scratch>/moai-m4 ./cmd/moai`), the installed `moai` judged only ledger E3.

Pre-flight on `7998ea35b`: `go build ./...` exit 0; `GOOS=windows GOARCH=amd64 go build ./...` exit 0; `golangci-lint run --timeout=2m ./internal/cli/...` exit 0, `0 issues.`; `ls internal/cli/audit_plan_cmd.go` absent; `go test ./internal/cli -run TestVerify` -> `ok 25.240s` (green on arrival). Ledger E3, installed binary: `moai verify audit-plan --help` printed the `verify` GROUP help (`COMMANDS record / check / codex-review`, `FLAGS -h --help, --project-root`), exit 0.

RED (semantic: a stub verb was registered first and printed nothing, so every test compiled and reached its assertion): 12 top-level `--- FAIL` (74 `--- FAIL` lines with sub-tests), for example `TestAuditPlanCmd_PinsOnlyIsDefault: stdout is not a JSON object: EOF`, `TestAuditPlanCmd_ProjectRoot: stdout is not a JSON object: EOF`, `TestAuditPlanCmd_RawReadThroughTheReportingLoader: the verb must read through loadWorkflowAuditSection`, `FAIL github.com/modu-ai/moai-adk/internal/cli 2.245s`, exit 1.

GREEN: `go test ./internal/cli -count=1 -v -run '^TestAuditPlanCmd_'` -> 16 top-level `--- PASS:` (the nine AC names `PrintsPlan`, `PinsOnlyIsDefault`, `InvalidConfigExitsOne`, `WritesNoFiles`, `ProjectRoot`, `OutputContract`, `NoAskUserQuestion`, `UnreadableConfig`, `ResultCheck`, plus `RegisteredUnderVerify`, `HelpNamesTheVerbAndTheRootObligation`, `ResultCheck_Hardening`, `ResultCheck_PredicateMutantsAreKilled`, `ChecksNoStore`, `RawReadThroughTheReportingLoader`, `ConsumesTheResolver`), `ok 2.817s`, exit 0. `TestAuditPlanCmd_ResultCheck` carries the nine sub-tests `a_old_server_shaped_result` ... `i_required_entries_with_a_malformed_verdict`.

Mutants (D4), mechanical: `TestAuditPlanCmd_ResultCheck_PredicateMutantsAreKilled` runs six predicate variants over 27 fixtures, the real predicate agrees with all, and each mutant dies (logged): `literal_verdict_not_inconclusive` (the AC mutant, `verdict != "inconclusive" && gate == "required"`) on `i_verdict_member_absent, i_verdict_empty, i_verdict_pas` and the `n_`/`t_` fixtures and on none of the first five (a)-(e) rows; `pass_only` on `p_fail_is_answered`; `case_and_space_folded` on `n_*`; `non_string_accepted` on `t_verdict_*` and the absent member; `any_nonempty_verdict` on `i_verdict_pas`; `gate_ignored` on `d_codex_gate_advisory`. Real-source mutation as well: replacing the body of `auditEntryAnswered` with the literal predicate failed `TestAuditPlanCmd_ResultCheck/i_required_entries_with_a_malformed_verdict` and 11 `Hardening` sub-tests (exit 1); changing `v != pass && v != fail` to `v != pass` failed `ResultCheck_Hardening/p_fail_is_answered` (exit 1); the file was restored byte-identical (`cmp` exit 0).

Structural: `grep -n 'AskUserQuestion\|mcp__askuser__\|audit-multi\|loadConvergenceResult\|check-session\|NewDefaultWorkflowConfig\|NewDefaultConfig' internal/cli/audit_plan_cmd.go` -> no output, exit 1; `grep -n loadWorkflowAuditSection ...` -> lines 18, 230, 244, exit 0; `grep -n workflowAuditPins ...` -> exit 1; `grep -rl --exclude="*_test.go" "ResolveAuditPlan(" internal` -> `internal/config/audit_plan.go`, `internal/cli/mcp_worktree_root.go`, `internal/cli/audit_plan_cmd.go`, `internal/cli/mcp_audit_multi.go`, `internal/auditreceipt/store.go`, exit 0 (AC-ACV-005 complete: definition + four consumers, three packages). Built from this tree: `moai-m4 verify audit-plan --help` -> a usage line `moai verify audit-plan [--flags]`, exit 0; on an invalid token: stderr exactly `audit-plan: audit_model "grok" unknown (want one of claude|codex|glm|multi)`, stdout empty, exit 1.

Post-change: both builds exit 0; `go vet ./internal/cli` exit 0; `golangci-lint run --timeout=2m ./internal/cli/...` `0 issues.` (one `QF1002` in the new test file fixed before the final run); `gofmt -l` on the two files empty. Coverage (`-coverprofile` over `^TestAuditPlanCmd_`): `audit_plan_cmd.go` 113/117 statements = 96.6%. Regression: `TestVerify` `ok 22.182s`; `TestAuditMulti_NoConfigNoArgs_ByteIdentical` and `TestAuditMulti_PinsOnlyConfig_ByteIdentical` `--- PASS`; the three `TestConvergence_No*` guards `--- PASS`.

Gaps: (1) the worktree-isolation guard REFUSED every Bash command that passes a JSON argument (`--result '{}'`, `'[1]'`, `'"x"'`, `"{}"`, and a full digest) with "construct too complex to verify" — only brace- and quote-free words pass (`--result abc` ran, `--result e30=` ran) — so `--result` was never exercised through a real binary from a guarded session; every checker claim rests on the cobra-level tests above, and the inline-single-quoted JSON input form that design.md §D.5 calls "one plain command" is NOT accepted by this session's guard (see E7 of the report). A heredoc python mutation script was refused as well; the real-source mutations above were made with the Edit tool instead. (2) `--result` that is not a JSON object is an `audit-plan:` error with exit 1 and empty stdout (AC-ACV-015 (f), REQ-ACV-016), not `ok: false`. (3) The broad regex `ByteIdentical` also selected three unrelated `TestTodo*` tests that failed with "lane boundary: a lane session cannot mutate the queue" — session environment, not M4; they were not re-measured in a clean environment. (4) A config-orphaned worktree whose primary cannot be identified reports `config_status: unreadable` with a note naming the primary (the SPEC does not say what the verb prints there; `resolveAuditGates` assumes codex `required`, a plan cannot be assumed). (5) Additions beyond the SPEC text: `config_root` (omitempty, set when the primary's workflow.yaml was read), `model_source: "unknown"` in the unreadable state, `convergence_check` is present and not ok when the plan is unreadable, and a `--project-root` that is not a directory exits 2 (a mistyped worktree path must not read as the default plan).

### M4b — the checker input becomes a file (`--result-file <path>`)

Measured against base HEAD `8ed87e9e7` (SPEC amendment 0.1.5), branch `WT-audit-model-convergence`, worktree `.moai/worktrees/t1423`; the M4b commit changes `internal/cli/audit_plan_cmd.go` and `internal/cli/audit_plan_cmd_test.go`, adds the two build-tagged fixture files `internal/cli/audit_plan_cmd_fifo_unix_test.go` (`//go:build unix`, `syscall.Mkfifo`) and `internal/cli/audit_plan_cmd_fifo_other_test.go` (`//go:build !unix`, reports unsupported so the FIFO cases skip), and appends this block. `internal/cli/testdata`, `internal/hook`, `internal/mcp` untouched; `UPDATE_AUDIT_MULTI_DEFAULT_GOLDEN` never set. Every Go run is one `unset ... CLAUDE_PROJECT_DIR && go test ...` invocation; judging toolchain `golangci-lint v2.1.6`; the only moai build exercised is one built from this tree into the scratchpad.

Pre-flight on `8ed87e9e7`: `go build ./...` exit 0; `GOOS=windows GOARCH=amd64 go build ./...` exit 0; `golangci-lint run --timeout=2m ./internal/cli/...` `0 issues.` (baseline); `go test ./internal/cli -run '^TestAuditPlanCmd_' -count=1` -> `ok 2.544s` (M4 green on arrival); RED-now `grep -c "result-file" internal/cli/audit_plan_cmd.go` -> `0`.

RED: first a COMPILE failure (`undefined: auditPlanResultMaxBytes`, `undefined: readBoundedResult`, `undefined: errAuditResultTooLarge`, `too many errors`, `FAIL ... [build failed]`) — a tool-class failure, not counted. Then stubs for the three symbols (no behaviour, no flag change) gave the semantic RED: 7 top-level `--- FAIL` (61 `--- FAIL` lines with sub-tests), exit 1, for example `TestAuditPlanCmd_HelpNamesTheVerbAndTheRootObligation: help must mention "--result-file"`, `TestAuditPlanCmd_ResultCheck/a_old_server_shaped_result: exit=-1 stderr=""` (the verb answers the unknown flag), `TestAuditPlanCmd_ReadIsBoundedWhileReading/an_endless_source_is_cut_at_cap+1: data len=0 err=<nil>, want errAuditResultTooLarge and no data`, `TestAuditPlanCmd_InlineResultFlagIsGone: err = <nil>, want an unknown-flag error for --result`, `FAIL github.com/modu-ai/moai-adk/internal/cli 2.149s`.

GREEN: `go test ./internal/cli -count=1 -v -run '^TestAuditPlanCmd_'` -> 18 top-level `--- PASS:`, 0 `--- FAIL`, `ok 4.122s`, exit 0. `TestAuditPlanCmd_ResultCheck` carries (a)-(i) as `a_old_server_shaped_result` ... `i_required_entries_with_a_malformed_verdict` and `j_unusable_result_file` with sub-tests `j1_missing_path`, `j2_directory`, `j3_empty_file`, `j4_json_that_is_not_an_object`, `j5_fifo`, `j5b_symlink_to_fifo`, `j5c_symlink_to_a_regular_file_is_followed`, `j6_oversized_file` (cap+1 bytes of a valid padded JSON object), `j6b_a_file_of_exactly_the_cap_is_accepted`; `TestAuditPlanCmd_ReadIsBoundedWhileReading` (counting reader that never ends: stops at exactly cap+1 bytes served, plus exact-cap and cap+1 inputs and a read-error control); `TestAuditPlanCmd_InlineResultFlagIsGone`; the D4 `_Hardening` fixtures and `_PredicateMutantsAreKilled` now feed a file.

Mutants (D5/D6), real-source, each restored byte-identical (`cmp` exit 0): removing the regular-file check (`if false && ...`) failed `j2_directory`, `j5_fifo` and `j5b_symlink_to_fifo` (the latter two by the 5 s deadline, `the verb did not return within 5s`); `io.LimitReader(r, limit+1)` -> `limit` failed `j6_oversized_file` and two `ReadIsBoundedWhileReading` sub-tests; `io.ReadAll(r)` (unbounded) failed `an_endless_source_is_cut_at_cap+1` through the reader's 64 MiB safety stop.

Structural and regression: `grep -n "audit-multi\|loadConvergenceResult\|check-session" internal/cli/audit_plan_cmd.go` -> no output; `grep -c "result-file"` -> 16; `grep -n '"result"'` -> no output (old flag gone); `TestAuditPlanCmd_NoAskUserQuestion` and `_ChecksNoStore` pass; M1 pair (`^TestAuditMulti_NoConfigNoArgs_ByteIdentical`, `^TestAuditMulti_PinsOnlyConfig_ByteIdentical`) `--- PASS`; `TestConvergence_NoNewAuditModelEnum` and `TestAuditMulti_NoHardErrorPath_AC_AMM_024` `--- PASS`. Post-change: both builds exit 0; `go vet ./internal/cli/` exit 0 and `GOOS=windows GOARCH=amd64 go vet ./internal/cli/` exit 0; `golangci-lint run --timeout=2m ./internal/cli/...` `0 issues.`; `gofmt -l` on the four touched Go files empty. Coverage (`-coverprofile` over `^TestAuditPlanCmd_`): per function `readBoundedResult` 100.0%, `readAuditResultFile` 85.0%, `checkAuditResult` 97.1%, `newVerifyAuditPlanCmd` 98.1%; file 139/146 statements = 95.2% (summed by hand from the profile blocks); the uncovered statements are I/O-error arms a test cannot reach without a fault injector (open failing after a successful stat, a read error from a regular file, a stat error other than not-exist).

Built-binary end-to-end (`go build -o <scratch>/moai-m4b ./cmd/moai`, run from the guarded worktree session, plain commands, the path the only argument): digest file on this worktree -> `config_status: absent`, `convergence_check {"ok": true, "unmet": []}`, exit 0; on a scratch tree with `model: "multi"` the old-server-shaped digest file -> `ok: false`, `unmet: ["codex"]`, reason naming the codex entry and `plan_source`, exit 0, and the passing digest -> `ok: true`, exit 0; missing path -> `audit-plan: --result-file <path>: does not exist`, exit 1; a directory -> `not a regular file (mode drwxr-xr-x)`, exit 1; a FIFO under `timeout 5` -> `audit-plan: --result-file <path>: not a regular file (mode prw-r--r--)`, exit 1, returned at once (no hang); the old spelling `--result abc` -> `Unknown flag: --result.`, exit 1. This is the first run of the checker through a real binary under the worktree guard: the guard accepted every file-path invocation.

Gaps: (1) the Windows path of the verb is compiled (`GOOS=windows go build` and `go vet` exit 0) but not run; the FIFO and symlink cases skip there by design. (2) `TestAuditPlanCmd_WritesNoFiles` passed under RED as well, because its checks do not read the verb's exit status; its new file-based runs prove "no regular file under `.moai` changed" only now that the flag is real. (3) The size bound is checked by the bounded reader and by exact-cap / cap+1 fixtures, not by a multi-hundred-megabyte file. (4) Reading "exactly that path once" is by construction (one `os.Stat`, one `os.Open`, one bounded read); no test counts opens. (5) The `-run` patterns have no end anchor (worktree guard); the selected set was counted with `go test -list '^TestAuditPlanCmd_'` -> 18 names, equal to the 18 `--- PASS` lines.

### M5 — stale-deferral cleanup (AC-ACV-017)

Measured against base HEAD `50103e618` (M4b), branch `WT-audit-model-convergence`, worktree `.moai/worktrees/t1423`; the M5 commit changes `internal/cli/mcp_audit.go`, `internal/cli/mcp_audit_test.go`, `internal/config/audit_models.go`, `internal/config/closed_sets.go`, `internal/settings/schema_sections.go`, `internal/web/mcp_audit_surface_test.go` and appends this block. `internal/cli/testdata`, `internal/hook`, `internal/mcp` untouched; `UPDATE_AUDIT_MULTI_DEFAULT_GOLDEN` never set. Every Go run is one `unset ... CLAUDE_PROJECT_DIR && go test ...` invocation; judging toolchain `golangci-lint v2.1.6`.

RED-now on `50103e618`: `grep -rn "multiConvergenceImplemented" --include="*.go" internal` -> 5 hits (`mcp_audit_test.go:40`, `:41`, `mcp_audit.go:27`, `:31`, `:49`), exit 0; `grep -rn "deferred to SPEC-AUDIT-MULTI-MODEL" internal` -> `mcp_audit.go:51`, exit 0; `grep -rn --exclude="*_test.go" "AP-8" internal` -> `audit_models.go:28`, `mcp_audit.go:8`, `:27`, exit 0; `grep -n "const sentinel" internal/web/mcp_audit_surface_test.go` -> `58: const sentinel = "activeAuditBackend"`, exit 0. Pre-flight: `go build ./...` exit 0; `GOOS=windows GOARCH=amd64 go build ./...` exit 0; `golangci-lint run --timeout=2m ./internal/cli/... ./internal/config/... ./internal/web/... ./internal/settings/...` `0 issues.` (baseline).

Vacuity observation (verification-completeness §1.1): with `multiConvergenceImplemented` and `activeAuditBackend` deleted and the web guard still unedited, `go test ./internal/web -run TestWebConsole_AuditNoForkedInterpreter -count=1 -v` -> `--- PASS`, `ok`, exit 0 — the original guard passes vacuously once its sentinel names no symbol. After the retarget, a mutant sentinel (`ResolveAuditPlanMUTANT`) gave `--- FAIL: TestWebConsole_AuditSentinelExists` (`func ResolveAuditPlanMUTANT( is not declared in internal/config non-test source`) while `TestWebConsole_AuditNoForkedInterpreter` stayed `--- PASS`, exit 1; the file was restored to `const sentinel = "ResolveAuditPlan"` and re-run green.

GREEN: `grep -rn "multiConvergenceImplemented" --include="*.go" internal` -> no output, exit 1; `grep -rn "multiConvergenceImplemented\|activeAuditBackend\|deferred to SPEC-AUDIT-MULTI-MODEL" internal` -> no output, exit 1; `grep -rn --exclude="*_test.go" "AP-8" internal` -> no output, exit 1; `grep -n "const sentinel" internal/web/mcp_audit_surface_test.go` -> `80:const sentinel = "ResolveAuditPlan"`, exit 0; `go test ./internal/web -run 'TestWebConsole_AuditNoForkedInterpreter|TestWebConsole_AuditSentinelExists' -count=1 -v` -> both `--- PASS` (the control carries two sub-tests, the live-symbol check and a negative control), `ok`, exit 0. `go build ./...` after the deletions: exit 0 (the compiler found no other reference to the removed symbols).

Regression: `./internal/cli -run '^(TestBuildAuditEnvBlock|TestAuditMulti_NoConfigNoArgs_ByteIdentical|TestAuditMulti_PinsOnlyConfig_ByteIdentical|TestConvergence_NoNewAuditModelEnum|TestConvergence_NoDirectFrontmatterRead|TestMCPAudit_)'` -> 8 top-level `--- PASS`, `ok`, exit 0; `./internal/settings -run TestAuditPinFields` -> 2 `--- PASS`; `./internal/config` -> `ok`; `./internal/web` four audit tests `--- PASS`. Post-change: both builds exit 0; `go vet` of the four packages exit 0; `golangci-lint` `0 issues.`; `gofmt -l` on the six Go files empty. The full `internal/cli` suite was not run (heavy-run rule).

Gaps: (1) `internal/config/mcp_audit_config_test.go:114` still carries the words "owned by a future SPEC (AP-8)" in a test comment — a test file outside the M5 file list; AC-ACV-017 forbids `AP-8` only in non-test Go files, so it is left and disclosed. (2) The deleted `TestActiveAuditBackend_*` tests had no replacement: unknown-model rejection is covered by `config.ResolveAuditPlan` tests (`internal/config/audit_plan_test.go`) and the `audit-plan` verb's `grok` case, which this block did not re-run in isolation. (3) The names `activeAuditBackend` / `multiConvergenceImplemented` remain in other SPECs' markdown (`SPEC-MOAI-MCP-SERVER-001`, `SPEC-AUDIT-MULTI-MODEL-001`, `SPEC-MCP-CONSOLE-001`, `SPEC-WEB-CONSOLE-REDESIGN-001`) as history; none edited.

### M6 — committed config flip (AC-ACV-018, AC-ACV-019 re-run)

Measured against base HEAD `251e5a51e` (M5), branch `WT-audit-model-convergence`, worktree `.moai/worktrees/t1423`; the M6 commit changes `.moai/config/sections/workflow.yaml`, adds `internal/config/workflow_audit_committed_test.go` and appends this block. `internal/template`, `internal/cli/testdata`, `internal/hook`, `internal/mcp` untouched; `UPDATE_AUDIT_MULTI_DEFAULT_GOLDEN` never set. Every Go run is one `unset ... CLAUDE_PROJECT_DIR && go test -count=1 -v -run ...` invocation (the worktree guard refuses a `-run` pattern containing `$` or the word "commit", so patterns are unanchored at the end and the selected set was counted from `--- PASS` lines); judging toolchain `golangci-lint v2.1.6`. The `moai slot` lease `go-test-internal-cli` was held for the runs and released.

RED-now on `251e5a51e`: `grep -n "effort: medium" .moai/config/sections/workflow.yaml` -> `25:            effort: medium` (claude pin) plus unrelated tier lines (the first three hits are 25, 64, 65; the output was cut at three); `grep -c "model: multi" .moai/config/sections/workflow.yaml` -> `0`. Pre-flight: `go build ./...` rc 0; `GOOS=windows GOARCH=amd64 go build ./...` rc 0; `golangci-lint run --timeout=2m ./internal/config/...` -> `0 issues.`. New test `TestCommittedWorkflowYamlAuditAlignment`, semantic RED before the edit (verbatim): `committed workflow.audit.model = "", want "multi"`; `committed claude pin effort = "medium", want "high" (DefaultClaudeAuditEffort)`; `committed workflow.yaml header comment still names claude-opus-5-5/medium`; `loaded Audit.Model = "claude", want "multi"`; `loaded Audit.Claude.Effort = "medium", want "high"`; `--- FAIL: TestCommittedWorkflowYamlAuditAlignment`, `FAIL github.com/modu-ai/moai-adk/internal/config`, rc 1.

GREEN: after the edit `TestCommittedWorkflowYamlAuditAlignment` and `TestAuditConfig_DefaultProfile` each `--- PASS`, `ok`; `grep -n -A1 "model: claude-opus-5-5"` -> `28:` model / `29:            effort: high`; `grep -nE "^        model: multi$"` -> `26:        model: multi`; `grep -n "claude-opus-5-5/medium"` -> no output; `grep -c "model: multi"` on the template yaml -> `0`. The test also asserts the template claude pin is `high` with no `model` and no `gates` key under `audit`, the committed file has no `gates` key, codex and GLM pins are unchanged and the Go default `Audit.Model` is `claude`.

Leak-sensitive sets, BEFORE (new test absent, yaml unedited) versus AFTER, same commands, `CLAUDE_PROJECT_DIR` unset: the sorted set of `--- PASS/FAIL/SKIP` lines (subtests included) is identical for every set, no FAIL anywhere. `./internal/cli` prefixes `TestAuditMulti_` 28, `TestRunMultiAudit_` 15, `TestConvergence_` 3, `TestCodexAudit_` 27, `TestAuditPlanCmd_` 18, `TestCodexTurnReview_` 3, `TestMain_` 1 top-level PASS; `./internal/auditreceipt` 26; `./internal/settings` 90 PASS and 1 SKIP; `./internal/web` `TestWebConsole_Audit` 2; `./internal/template` `TestClaudeAuditTemplateSurfacesAndCatalogHash` 1; `./internal/hook` `TestAuditReceiptGuard` 2, `TestCodexReviewGate` 1, `TestReviewGate` 5, `TestSubagentStop_` 4, `TestSubagentStart_` 1, `TestPreToolUse_DeniesPhaseEntry` 1; `./internal/config` whole package 495 top-level PASS before and 496 after, the only difference the new test. With `CLAUDE_PROJECT_DIR` set to this tree: `TestMain_ScrubsClaudeProjectDir` PASS, `TestAuditMulti_` 28, `TestCodexAudit_` 27 and hook `TestAuditReceiptGuard` 2 all PASS; `grep -c "EnvClaudeProjectDir" internal/cli/main_test.go` -> `3`.

Template-drift and `moai update` guards run after the edit (all PASS, no FAIL): `./internal/template` `TestContractMode` 11 PASS + 4 SKIP, `TestJevDoctrineAmendment`, `TestRuleTemplateMirrorDrift`, `TestReviewGate` (4), `TestAC_CONTRACT_020`; `./internal/contract/kickoff` `TestJevAmendmentLinkage`; `./internal/cli` `TestMergeUserFiles_` (5), `TestStripRetiredModelConfig` (4), `TestCollectMergeableFiles_` (2).

Files that read the repository's `workflow.yaml` by path (non-test): `internal/cli/audit_plan_cmd.go:265` (reads `readRoot`, the verb's explicit or primary root), `internal/cli/audit_pin.go:31`, `internal/cli/multi_review_gate.go:218`, `internal/cli/mcp_codex.go:2393` and `:2437`, `internal/cli/doctor_harness.go:55`, `internal/auditreceipt/store.go:225` and `:254` (all take a project or tree root argument, not the cwd); `internal/config/loader.go` and `loader_slot_lease.go` (take a config dir argument); `internal/cli/agentlint/workflow_lint.go:108` (cwd-relative, a lint verb the audit tests do not call). No existing test flipped.

Consequence of the flip (reported, not fixed): in this repository `CodexGateRequired` is now true for the receipt guard, so audits here need a receipt when the codex leg runs, and `audit_multi` without a gates argument fails closed on a missing codex leg. This is the intended behaviour of the card.

Gaps: (1) the four `TestContractMode*` base-ref guards (`InheritedDivergence`, `ConstitutionDriftNotIncreased`, `AlwaysLoadedBudget`, `EmitterSites`) SKIP because `MOAI_GR_BASE` is not set; the change-set allowlist test that does run lists `.moai/config/sections/workflow.yaml` as allowed. (2) The harness shows output but not exit codes for the single-invocation greps; the empty output and `0` count were read from stdout, not from a displayed rc. (3) The full `internal/cli` and `internal/hook` suites were not run (heavy-run rule); only the named sets above. (4) Walk-up-from-cwd readers were judged by reading source and by the AFTER sets staying green, not by an exhaustive sweep of every test in the repository. (5) `gofmt -l internal/config` lists `internal/config/slice.go`, a pre-existing file this card does not touch; `gofmt -l` on the new test file is empty.

### M7 — activation text (AC-ACV-012 document part, AC-ACV-013, AC-ACV-014, AC-ACV-016)

Measured against base HEAD `d1e4e68b4` (M6), branch `WT-audit-model-convergence`, worktree `.moai/worktrees/t1423`. Template copies edited first, then the local mirrors, then `make agents-emit`, then `make build VERSION=dev` (both exit 0). The commit changes the four documents in both trees (`plan-auditor.md`, `sync-auditor.md`, the cross-model audit skill, `workflows/sync.md`), the two template `.codex` role TOMLs (generated), `internal/template/catalog.yaml` (generated hashes for `moai`, `sync-auditor`, `plan-auditor` and the skill), the new test `internal/template/audit_plan_doc_surface_test.go`, and this block. No Go production file, `workflow.yaml`, `sync-audit-4dim.js`, `review.md`, hook, mcp or golden file touched; the `sync-auditor` `tools:` list untouched.

RED-now on `d1e4e68b4`: `grep -c "moai verify audit-plan"` over the eight document copies -> `0` each; `grep -c "plan surface unreachable"` over the four local documents -> `0` each; `grep -c "moai:closure-second-review" .claude/agents/moai/sync-auditor.md` -> `2` (the pre-edit value). `TestClaudeAuditTemplateSurfacesAndCatalogHash`, `TestTemplateNoInternalContentLeak` and `TestCrossModelAuditConsumerLink` PASS on arrival. New test `TestAuditPlanDocSurface` semantic RED before any text (verbatim excerpt): `--- FAIL: TestAuditPlanDocSurface` with eight failing sub-tests (`plan-auditor.md`, `sync-auditor.md`, `SKILL.md`, `sync.md`, each in `source` and `template`), e.g. `does not carry the literal "moai verify audit-plan"`, `"--project-root"`, `"--result-file"`, `"config_status"`, `"config_status: unreadable"`, `"plan surface unreachable, legacy path used"`; `FAIL github.com/modu-ai/moai-adk/internal/template`, rc 1. The test comments state that literal checks are tests-of-record for the TEXT only; the mechanical backstops are the receipt guard and the checker.

GREEN: `TestClaudeAuditTemplateSurfacesAndCatalogHash`, `TestTemplateNoInternalContentLeak`, `TestAuditPlanDocSurface` (8 sub-tests) and `TestAuditPlanDocSurfaceKeepsClosureMarkers` each `--- PASS`, `ok`. `grep -c "moai verify audit-plan"` over the four local documents -> 2, 2, 3, 2; the same over both generated TOMLs -> 2 each; `grep -cF "Single-backend audit mode (per the project"` over the four agent copies -> `0` each; `grep -c "moai:closure-second-review" .claude/agents/moai/sync-auditor.md` -> `2` (unchanged); `config_status: unreadable`, `plan surface unreachable, legacy path used` and `Shared diagnostic snapshot contract` -> 1 each in the four local documents; `cross_model_required`, `audit_multi unreachable`, `audit-plan --result-file` -> at least 1 in both `sync.md` copies; `grep -c "audit_multi"` over both `sync-audit-4dim.js` copies -> `0`. The full `internal/template` package ran under the `moai slot` lease `go-test-internal-template` (released): `ok ... 168.331s`.

Binding clarifications carried by the text: the digest file is written fresh and overwritten immediately before each check and an earlier audit's file is never reused (D7); in the sync phase the orchestrator writes the digest file and runs the check, and the cold read-only `sync-auditor` returns the result members and states its verdict is not final until the orchestrator's check passes (D4/D8); the auditors and the orchestrator pass only the file path, never inline JSON, and pass `--project-root` as their own toplevel (the worktree's `CLAUDE_PROJECT_DIR` names the primary checkout).

Gaps: (1) literal checks prove the text, not that a model obeys it; (2) the `sync.md` text was not exercised in a live sync run; (3) `review.md` still carries prose that interprets the `audit_model` token (out of scope, decision D11); (4) whether a subagent's `Write` is allowed at the `.moai/state/` path was not observed (plan-auditor); (5) the harness shows exit codes only where echoed.

### M8 — repair of sync-audit findings F1, F2, F3, F6 (leader-approved, option A)

Measured against base HEAD `6393b49fe` (M7), branch `WT-audit-model-convergence`, worktree `.moai/worktrees/t1423`. Pre-flight on arrival: `go build ./...` rc 0; `GOOS=windows GOARCH=amd64 go build ./...` rc 0; `golangci-lint run --timeout=2m ./internal/cli/...` (v2.1.6) -> `0 issues.` rc 0; the set `^(TestAuditMulti_|TestRunMultiAudit_|TestConvergence_|TestCodexAudit_|TestAuditPlanCmd_|TestCodexTurnReview_|TestMain_)` -> `ok internal/cli 14.535s`, 95 top-level tests; `./internal/auditreceipt` and `./internal/config` `ok`. Every Go run was one `unset ... CLAUDE_PROJECT_DIR && go test ...` invocation; no lease was needed (the set runs in about 14 s). `UPDATE_AUDIT_MULTI_DEFAULT_GOLDEN` never set; the golden file untouched.

Change: `handleAuditMulti` reads the audit section once at call start and fixes the enforcement from it (`callStartEnforcement`): the operator-written gates (`audit.gates`, `audit.model` cells), a supplied `off`/`advisory` replacing them, a supplied `required` adding nothing (design.md §D.2). The result is carried in the new `MultiAuditConfig.EnforcementGates` (+ `EnforcementNote` for the orphaned-root assumption); `runMultiAudit` enforces from it and re-reads `workflowAuditGates` only when the field is nil (callers that carry no plan). F6: the `awaitCodexTurnReview` doc comment only.

RED (semantic: each reaches its assertion and fails because behaviour is missing), `TestAuditMulti_SuppliedGateOverridesConfiguredRequiredAtEnforcement` and `TestAuditMulti_EnforcementFollowsTheCallStartPlan`, rc 1: `gate_unmet = codex, want absent: the entry says advisory, so enforcement must agree`; `overall_verdict = fail, want pass (an advisory gate with no verdict is fail-open)`; `overall_verdict = pass, gate_unmet = <nil>, want fail / codex (the call-start plan required codex)`; `overall_verdict = fail, want pass (nothing was configured at call start)`. F3 RED by source overlay (the survivor mutant applied to `audit_plan_cmd.go`, reverted afterwards): `TestAuditPlanCmd_` -> only the two new tests fail (`convergence_check = &{OK:true ...}, want ok=false`; `mutant plan_source_only_when_model is killed by: g_old_server_digest_without_plan_source`), the rest of the suite green.

GREEN: the same set `ok internal/cli 13.164s`, 100 top-level tests (95 + 5 new: the F1 test, the F2 test, the gates-only checker fixture test, the plan_source mutant test, the carry-less fallback test; `TestCallStartEnforcement` is a sixth, outside the prefix list). Coverage (`-coverprofile`, same names): `callStartEnforcement` 100.0%, `handleAuditMulti` 100.0%, `runMultiAudit` 92.1%, `checkAuditResult` 97.1%. `go vet ./internal/cli/` rc 0; `golangci-lint run --timeout=2m ./internal/cli/...` -> `0 issues.` rc 0; `gofmt -l` on the touched files empty; both builds rc 0; `./internal/auditreceipt` and `./internal/config` `ok`.

Gaps: (1) the full `internal/cli` suite was not run (heavy-run rule); (2) the narrow case of a config-orphaned root whose primary checkout cannot be identified is unit-tested (`TestCallStartEnforcement`) but not driven through the handler end to end; (3) the tree's receipt predicate (`receiptCodexGateRequired`) still reads configuration after the fan-out, so under a mid-call edit the `audit_receipt` member can differ from the enforced verdict; that reader is outside the enforcement path and was not changed; (4) the mutant in `TestAuditPlanCmd_ResultCheck_PlanSourceMutantIsKilled` is run over the real checker with a transformed plan, which is exactly the survivor's semantics, but it is not an edit of the checker source (the source-level overlay RED above is).

### M9 — repair of delta sync-audit finding NEW-1 (leader-approved, the last extension)

Measured against base HEAD `6074d7f48` (SPEC 0.1.6), branch `WT-audit-model-convergence`, worktree `.moai/worktrees/t1423`. Pre-flight on arrival: `go build ./...` rc 0; `GOOS=windows GOARCH=amd64 go build ./...` rc 0; `golangci-lint run --timeout=2m ./internal/cli/...` (v2.1.6) -> `0 issues.` rc 0; the set `^(TestAuditMulti_|TestRunMultiAudit_|TestConvergence_|TestCodexAudit_|TestAuditPlanCmd_|TestCallStartEnforcement|TestCodexTurnReview_|TestMain_)` -> `ok internal/cli 13.330s`, 101 top-level tests (`go test -list` count 101); `./internal/auditreceipt`, `./internal/config` and `./internal/hook -run ^TestAuditReceiptGuard` `ok`. Every Go run was one `unset ... CLAUDE_PROJECT_DIR && go test ...` invocation; no lease was needed (the set runs in about 13 s). `UPDATE_AUDIT_MULTI_DEFAULT_GOLDEN` never set; the golden file untouched.

Change: `recordAuditReceiptAt` (new, `internal/cli/mcp_audit_receipt.go`) is `recordAuditReceipt` with the codex gate the exposure follows passed in; a non-nil gate is the call-start one, nil keeps the `receiptCodexGateRequired(root)` re-read. `recordAuditReceipt` delegates with nil, so the single-backend `codex_audit` call sites are byte-identical. `runMultiAudit` passes `EnforcementGates.Codex == required` when the handler carried a plan (the same value the enforcement keys on: configured gates, a supplied `off`/`advisory` replacing them, a supplied `required` adding nothing, the orphaned-root assumed-required case), and nil otherwise. The receipt is still recorded whenever codex participated; only its exposure (`audit_receipt`) follows the gate.

RED (semantic: each reaches its assertion and fails because behaviour is missing), rc 1: `audit_receipt = "", want an id: the call-start plan required codex, so the failure's receipt is exposed (REQ-ACV-009)`; `audit_receipt = "rcpt-7ec49e52e21080617259840a", want absent: nothing was configured at call start, so the codex gate was not required`; `audit_receipt = "rcpt-88e348c04afe25a5be5f04fe" (exposed = true), want exposed = false (gate_unmet = <nil>)` (advisory over a configured multi); `audit_receipt = "rcpt-3b1764f043e138af737a24a1" (exposed = true), want exposed = false (gate_unmet = <nil>)` (advisory over an audit.gates required). The supplied-`required` neighbours and the carry-less test passed before and after.

GREEN: the same set `ok internal/cli 12.707s`, 104 top-level tests (101 + `TestAuditMulti_ReceiptExposureFollowsTheCallStartGate`, `TestAuditMulti_ReceiptExposureAgreesWithASuppliedGate`, `TestRunMultiAudit_CarrylessCallerKeepsTheReceiptReRead`); no existing test edited. Coverage (`-coverprofile`, same names): `recordAuditReceipt` 100.0%, `recordAuditReceiptAt` 88.9% (the uncovered lines are the unchanged empty-root and store-root-error returns), `runMultiAudit` 92.3%, `handleAuditMulti` 100.0%. `go vet ./internal/cli/` rc 0; `golangci-lint run --timeout=2m ./internal/cli/...` -> `0 issues.` rc 0; `gofmt -l` on the touched files empty; both builds rc 0; `./internal/auditreceipt`, `./internal/config`, `./internal/hook -run ^TestAuditReceiptGuard` `ok`.

Behaviour note: a supplied `advisory` over a configured `required` used to expose a receipt (the config re-read said required) while the entry and the enforcement said advisory; it no longer does. The SubagentStop guard (`internal/hook/audit_receipt_guard.go`) keeps its own read of the session tree's configuration, unchanged by this milestone.

Gaps: (1) the full `internal/cli` suite was not run (heavy-run rule); (2) the orphaned-root-with-unidentifiable-primary path is carried through `callStartEnforcement` (unit-tested) but not driven through the handler end to end for the receipt; (3) the hook-side `CodexGateRequired` read at the guard is a separate, later check and was not changed.

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_status: audit-ready
sync_complete_at: 2026-10-02T14:50:22Z
sync_commit_sha: 6d2e3d4fce6be9d960091277a9c6d6b779d57fdf   # the sync commit; backfilled by the lane orchestrator in a following commit (mechanical placeholder completion, no trailer claim)
card: t1423
tier: L
run_commit_range: 25e6e9737..4ae90d62f   # M1 dd44df3cf .. M9 4ae90d62f; 25e6e9737 is the parent of M1 (git log --oneline)
spec_amendments: { "0.1.5": 8ed87e9e7, "0.1.6": 6074d7f48 }
frontmatter_status_transitions: { spec_md: "in-progress -> implemented -> completed (one sync commit)" }
changelog_entry_position: "[Unreleased] / Added, first entry"
b12_self_test_a: "grep -c SPEC-AUDIT-MODEL-CONVERGE-001 CHANGELOG.md -> 0 before the entry (no duplicate)"
b12_self_test_b: "20 live AC-ACV identifiers in acceptance.md (unique-id count, no reserved tokens present); the entry says 20"
b12_self_test_c: "every implementation file the entry names was Read, and its path exists (ls)"
plan_audit:
  iteration_1: { verdict: FAIL, score: 0.81, report: .moai/reports/t1423/plan-audit-iter1.md }
  iteration_2: { verdict: FAIL, score: 0.81, report: .moai/reports/t1423/plan-audit-iter2.md }
  iteration_3: { verdict: FAIL, score: 0.84, report: .moai/reports/t1423/plan-audit-iter3.md }
  iteration_4: { verdict: PASS-WITH-DEBT, score: 0.91, report: .moai/reports/t1423/plan-audit-iter4.md, audited_sha: 259fbd697 }
  iteration_5_delta_0_1_5: { verdict: PASS-WITH-DEBT, score: 0.89, report: .moai/reports/t1423/plan-audit-iter5.md, audited_sha: 8ed87e9e7 }
sync_audit:
  full: { verdict: PASS-WITH-DEBT, score: 85, report: .moai/reports/t1423/sync-audit.md, audited_sha: 6393b49fe }
  delta_m8: { verdict: PASS-WITH-DEBT, score: 88, report: .moai/reports/t1423/sync-audit-delta.md, audited_sha: 513ba5660 }
  delta_m9: { verdict: PASS-WITH-DEBT, score: 91, report: .moai/reports/t1423/sync-audit-delta2.md, audited_sha: 4ae90d62f }
cross_model_note: "GLM answered HTTP 401 in every sync-audit round (advisory); the codex leg answered; the MCP server of the auditing session predates the card, so the plan_source check reads ok:false pre-rollout — the enforcement is proved by unit tests, not by a live post-rollout run"
carried_debt:          # all carried, none edited in this sync commit; owner = the leader's follow-up
  - { id: NEW-2, owner: leader-follow-up, title: "on a config-orphaned root with an unidentifiable primary, a supplied advisory relaxes the assumed-required codex gate (internal/cli/mcp_audit_multi.go callStartEnforcement)" }
  - { id: F5, owner: leader-follow-up, title: "the real binary writes .moai/state/config-cache.json on a cache miss, against the SPEC's start-up 'directories only' claim (design.md D.5, spec.md R-33)" }
  - { id: DB1, owner: leader-follow-up, title: "legacy-window-fail-open-by-design (sync-audit: CARRIED, owner leader rollout, plan.md section J)" }
  - { id: DB2, owner: leader-follow-up, title: "full-result-form-invites-shell-injection (sync-audit: RESOLVED, the input is a file path via --result-file, never inline JSON)" }
  - { id: DB3, owner: leader-follow-up, title: "reader-fix-reaches-review-gate-and-codex_task (sync-audit: CARRIED, disclosed in the CHANGELOG entry)" }
  - { id: DB4, owner: leader-follow-up, title: "instruction-text-criteria-are-shallow (sync-audit: CARRIED as tests-of-record, mechanical backstops verified)" }
  - { id: DB5, owner: leader-follow-up, title: "signature-brittleness (sync-audit: CARRIED with a pin, TestAuditPlanDocSurface asserts the literal in all eight copies)" }
  - { id: DB6, owner: leader-follow-up, title: "startup-regular-file-effects-unmeasured (measured and refined into F5)" }
  - { id: DB7, owner: leader-follow-up, title: "codex-hosted-auditor-cannot-run-the-verb (sync-audit: UNOBSERVED, CARRIED, owner leader smoke)" }
  - { id: DB8, owner: leader-follow-up, title: "req-014-vs-design-d8-sync-legacy-wording (sync-audit: RESOLVED in the instruction text, the SPEC clause is still unedited)" }
  - { id: iter5-D1, owner: leader-follow-up, title: "design.md:216 says the verb reads nothing under .moai/state beside an instructed file under .moai/state (doc wording)" }
  - { id: iter5-D2, owner: leader-follow-up, title: "acceptance.md:572 green path is labelled (M4) and should read (M4b) (doc wording)" }
  - { id: iter5-D3, owner: leader-follow-up, title: "stale 'seven milestones' prose beside a plan table with an M4b row (doc wording)" }
  - { id: delta2-N1, owner: leader-follow-up, title: "with a supplied advisory over a configured required, the receipt is stored but its id is not exposed" }
  - { id: delta2-N2, owner: leader-follow-up, title: "design.md section D.3 receipts bullet still words the exposure as the re-read of the tree configuration" }
  - { id: M5-stale-comment, owner: leader-follow-up, title: "internal/config/mcp_audit_config_test.go:114 still carries the AP-8 deferral wording" }
```

## §F Phase 4 Mode Selection

Written by the orchestrator (lane-8, factory run tm9i7y) before the first run-phase `Agent()` spawn. The `plan_audit` block of §E.1 stops at iteration 2 (owned by manager-spec); the plan-audit record below is the later one.

### F.1 Plan-audit record (final)

| Iteration | Subject (`audited_sha`) | Verdict | Score | Report (local, gitignored) |
|---|---|---|---|---|
| 1 | 53a42f013a45d376ba3e4289af677bc878d6f9a4 | FAIL | 0.81 | `.moai/reports/t1423/plan-audit-iter1.md` |
| 2 | 739556db5684a79feda06cad2b92d18837274fee | FAIL | 0.81 | `.moai/reports/t1423/plan-audit-iter2.md` |
| 3 | 31127d938cbef1dc9d5fec796f2645e1ea2ce2fa | FAIL | 0.84 | `.moai/reports/t1423/plan-audit-iter3.md` |
| 4 (delta confirmation, over the Tier L ceiling of 3, leader-approved 10-02, same criterion as t1411) | 259fbd6978ab59c5b44c8b94b0d337c88bce4ef6 | PASS-WITH-DEBT | 0.91 | `.moai/reports/t1423/plan-audit-iter4.md` |

Auditor model on every iteration: `claude-sonnet-5-5[1m]` (the session model, inherited by the subagent; the `claude-opus-5-5` audit pin was not applied to the in-session leg). Cross-model leg on every iteration: Claude anchor + codex (required) answered, GLM (advisory) HTTP 401, no receipt. Iteration 4: codex `fail` with two P2 findings that the auditor adjudicated as debt D3/D4 below; the auditor named the alternative reading (codex `fail` on a required gate binding regardless => FAIL) in its report.

Open debts carried into the run phase, owner and check point as classified by the auditor:
- DB1-DB8 of iteration 3 (see §E.1 `iteration_4_note`): to be checked for resolution at the sync stage (leader ruling).
- D3 (iteration 4): `plan.md` M4 text lists checker fixtures `(a)-(h)` while AC-ACV-015 has nine; carry "fixtures (a)-(i); AC-ACV-015 governs" in the M4 delegation prompt (editing plan.md would change its audited hash).
- D4 (iteration 4): fixture (i) leaves case/whitespace normalisation (`"PASS"`, `" pass "`), a `pass`-only checker and a wrong-JSON-type `verdict` unkilled, and no fixture holds a required `fail` entry; M4 adds these cases.

### F.2 Kickoff gate (plan -> run), autonomous form

Conditions of `.claude/rules/moai/workflow/auto-semantics.md` §9.1, each checked at 2026-10-02 against HEAD `259fbd697`:
1. Independent plan-audit verdict: PASS-WITH-DEBT 0.91 at `audited_sha` 259fbd697 (not a plain PASS; the entry rests on the leader's ruling A, which accepted the loop's outcome and set DB1-DB8 for the sync stage — disclosed here, not hidden). FAIL/INCONCLUSIVE would be a hard block; none stands.
2. Plan phase audit-ready: §E.1 `plan_status: audit-ready`.
3. Plan-artifact hashes unchanged since that verdict: `git status --short` empty and `git rev-parse HEAD` equals `audited_sha` at the time of this record.
4. No blocker open: none; the two below-gate Jev picks of round 3 were ruled by the leader.

decision record: decided_by=lane-8+orchestrator evidence_refs=.moai/reports/t1423/plan-audit-iter4.md#PASS-WITH-DEBT(0.91;audited_sha=259fbd6978ab59c5b44c8b94b0d337c88bce4ef6),.moai/reports/t1423/plan-audit-iter3.md#DB1-DB8,leader-ruling-A(10-02),progress.md#E.1(plan_status=audit-ready) ladder_path=gate-row:plan-to-run-Kickoff-AUTONOMOUS(auto-semantics-9.1)+leader-approved-ceiling-extension

No `/moai goal` is armed (the lane's own stage list is the termination judge; an armed goal's Stop-hook evaluator would block the lane's waits on background agents).

### F.3 Phase 4 mode selection

Input parameters: tier L; scope about 36 run-phase files (Go source and tests, local+template document mirrors, two `.codex` TOMLs, the committed `workflow.yaml`, a skill catalog hash); domains: Go (`internal/cli`, `internal/config`, `internal/auditreceipt`, `internal/web`), tests, agent/skill/workflow documents, config; file language mix: Go + markdown + yaml/toml; concurrency benefit: LOW (coding-heavy, shared files across milestones, one writer per tree); Agent Teams: not requested.

| Mode | Selected | Rationale |
|---|---|---|
| direct | no | not trivial |
| serial | **yes** | coding-heavy, ordered milestones M1..M7 with a baseline-first commit that must precede every behaviour commit |
| fanout | no | not research-heavy; write-capable work in one tree |
| sweep | no | not one uniform mechanical transform |
| agent-team | no | not requested |

Decision: serial

Justification: milestones M1..M7 are ordered by the commit graph (M1's golden precedes every behaviour change; M7 activation text is the card's last commit), the work is coding-heavy with shared files, and one writer per tree applies. `manager-lead` is not used: its entry predicate holds on paper (7 milestones, ~36 files, cross-domain), but a lane's standing spawn authority is depth-1 only, so a `manager-lead` spawned here could not spawn its write-capable leaf workers. Each milestone is one delegation, sequentially, with the Section A-E template.

Boundary case: none (scope and milestone count are far from the thresholds).

### F.4 Run-phase decisions

**M3a — padded gate values (EC-1 consequence).** The audited SPEC (EC-1, AC-ACV-003) makes the resolver match token and gate values exactly and case-sensitively AFTER trimming surrounding whitespace. Because every consumer now reads the resolver, a configured `gates.codex: "required "` (trailing space) is enforced fail-closed on every surface (`audit_multi`, `codex_audit`, the receipt predicate). Before this card it was deliberately NOT required; two pre-existing pins from an earlier completed SPEC (codex audit gate axes) encoded that: the golden case `required-with-trailing-space` of `TestCodexAudit_NonRequiredGateGoldenByteIdentical` and a row of `TestCodexGateRequired_ExactMatchOnly`. Uppercase `REQUIRED` stays non-required. The plan-audit did not raise this conflict; the M3a worker flipped the row, removed the golden case (commit f96cdb7da, files outside the milestone list, disclosed in its report), and the orchestrator deleted the orphaned golden file. Direction: a padded `required` is almost certainly a typo of `required`, and the old behaviour left such a gate silently unenforced, so the change is in the safe (fail-closed) direction; the card's purpose (audit.model consumed) does not otherwise touch single-backend `codex_audit`.

decision record: decided_by=lane-8+orchestrator evidence_refs=Jev(jev-1.13.0;padded_gate=accept_spec_trim_and_disclose@0.99;orphan_golden=delete_orphan_file@0.99),acceptance.md#EC-1,commit-f96cdb7da,.moai/reports/t1423/progress.md ladder_path=④-jev(operator-directed-script-path-same-TypeSafe-API;MCP-jev_ask-blocked-by-primary-config)

To re-measure at the sync stage: the behaviour change is recorded here so the sync audit can confirm it (a) is true on the final tree, (b) is acceptable, and (c) is mentioned in the CHANGELOG entry as a behaviour note.
