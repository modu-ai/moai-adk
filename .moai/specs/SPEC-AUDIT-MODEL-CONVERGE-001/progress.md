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

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

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
