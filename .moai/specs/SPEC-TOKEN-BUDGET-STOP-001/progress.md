# SPEC-TOKEN-BUDGET-STOP-001 — Progress

> Plan-phase artifact. Run-phase evidence (§E.2, §E.3) and sync-phase evidence (§E.4) are populated by manager-develop and manager-docs respectively. This file is parser-load-bearing: `internal/spec/era.go` `hasAnyProgressMarker` greps for the literal `§E.2` / `§E.3` / `§E.4` substrings — the headings below MUST be preserved verbatim.

---

## §E.1 Plan-phase Audit-Ready Signal

```
plan_status: audit-ready
plan_complete_at: 2026-07-09
tier: M
artifacts: 4 (spec.md + plan.md + acceptance.md + progress.md)
gears_requirements: 7
acceptance_criteria: 9
out_of_scope_topics: 6
spec_id_self_check: PASS (SPEC-TOKEN-BUDGET-STOP-001 → ^SPEC(-[A-Z][A-Z0-9]*)+-\d{3}$)
```

**Plan-phase self-verification** (per plan.md §E):
- [x] All GEARS requirements use valid patterns (Ubiquitous / When / While / Where + generalized subject; no IF/THEN).
- [x] All gap claims cite measured file + measured line range (vci §2 attribution).
- [x] §Out of Scope has ≥1 `### Out of Scope — <topic>` H3 sub-heading with `-` bullets (6 topics).
- [x] 12 canonical frontmatter fields, no snake_case aliases.
- [x] SPEC ID matches `^SPEC(-[A-Z][A-Z0-9]*)+-\d{3}$` (Pre-Write Self-Check → PASS).
- [x] /clear prohibition + warning-first preservation stated as HARD constraints.
- [x] Template-First boundary stated (Go code NOT templated; doctrine edits templated).
- [x] Cross-referenced devices verified to exist (plan.md §C — 11 devices verified).

---

## §E.2 Run-phase Evidence

### AC PASS/FAIL Matrix (all 9 MUST-PASS — observed 2026-07-09)

| AC ID | Status | Verification Command | Actual Output |
|-------|--------|---------------------|---------------|
| AC-TBS-001 | PASS | `go test -run TestGracefulAbortAtHardLimit ./internal/runtime/ -v` | PASS — CheckGracefulAbort returns non-empty handoff + shouldAbort=true at 90% |
| AC-TBS-002 | PASS | `go test -run TestGracefulAbortHandoffFormat ./internal/runtime/ -v` | PASS — handoff contains all 6-block tokens (✂──── markers, ultrathink., applied lessons:, 전제 검증:, 실행:) |
| AC-TBS-003 | PASS | `grep -rn '"os/exec"' internal/runtime/ \| grep -v _test.go` + `go test -run TestNoAutoClearInvocation ./internal/runtime/ -v` | 0 os/exec imports in budget.go/persist.go/config.go; TestNoAutoClearInvocation PASS |
| AC-TBS-004 | PASS | `go test -run TestRecordCallNoErrorOnBudgetExhaustion ./internal/runtime/ -v` | PASS — RecordCall compiles with no error return; completes at 140% without panic |
| AC-TBS-005 | PASS | `grep -c '\.moai/state/verify' .claude/rules/moai/core/agent-common-protocol.md` | 1 match (persistence location named); reachability obligation stated (3 matches) |
| AC-TBS-006 | PASS | `grep -c 'verification-claim-integrity\|§1.1\|surface' agent-common-protocol.md` + `grep 'drop the evidence'` + `grep 'evidence:' moai.md` | vci §1.1 surface 1+2 named (3 matches); "drop the evidence" rejected (1 match); evidence path in banner (L413 + L615) |
| AC-TBS-007 | PASS | `grep '\.moai/state/verify' template/.../agent-common-protocol.md` + `ls template/.../internal/runtime/` | template mirror carries same edit; internal/template/templates/internal/runtime/ not present (No such file) |
| AC-TBS-008 | PASS | `git diff --name-only HEAD -- .claude/hooks/moai/` | no new hook scripts authored by D (empty diff — no .claude/hooks/moai/ changes) |
| AC-TBS-009 | PASS | `grep 'func (t \*Tracker) RecordCall' internal/runtime/budget.go` + `go test -run TestHardLimitWarning ./internal/runtime/ -v` | signature unchanged `func (t *Tracker) RecordCall(agentName string, tokensIn, tokensOut int)` (no error return); TestHardLimitWarning PASS |

### PRESERVE invariants verified post-run

- RecordCall signature: `func (t *Tracker) RecordCall(agentName string, tokensIn, tokensOut int)` — NO error return added (AC-TBS-004/009).
- IsAtHardLimit signature: `func (t *Tracker) IsAtHardLimit(agentName string) bool` — unchanged.
- budget.go:18 `/clear` prohibition comment — preserved.
- budget.go:15-16 BC-V3R3-006 warning-first comment — preserved.
- @MX:ANCHOR (budget.go:20) + @MX:SPEC (budget.go:22) — preserved; new @MX:ANCHOR added on CheckGracefulAbort.
- TestNoAutoClearInvocation (budget_test.go:328) + TestHardLimitWarning (budget_test.go:126) + TestHardLimitAt90Pct (budget_test.go:106) + TestPersistProgressAt75Pct (budget_test.go:199) — all PASS (no regression).
- session-handoff.md body — NOT modified (reused, not changed).
- 7 verification keywords (go test, coverprofile, grep, sentinel, cmd/moai, bench, lint) in agent-common-protocol.md — preserved.
- Parallel-execution HARD obligation in agent-common-protocol.md — preserved.

---

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-07-09
run_commit_sha: 032eb9733 (M1) + 841843a9b (M2)
run_status: audit-ready
ac_pass_count: 9
ac_fail_count: 0
preserve_list_post_run_count: 0 regressions
coverage_internal_runtime: 90.5% of statements (target >=85%)
lint_issues_new: 0 (baseline 0; golangci-lint clean)
cross_platform_build: darwin=0 windows=0
l44_pre_commit_fetch: origin/main == HEAD (4a41138de) at pre-flight; 0 0 divergence
l44_post_push_fetch: <pending push>
new_warnings_or_lints_introduced: 0
total_run_phase_files: 7 (3 Go: budget.go + persist.go + budget_test.go; 2 LIVE doctrine: agent-common-protocol.md + moai.md; 2 template doctrine: agent-common-protocol.md + moai.md)
m1_to_mN_commit_strategy: 2 commits (M1 Go + frontmatter draft->in-progress; M2 doctrine + template mirror + progress.md §E.2/§E.3)
```

---

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: 2026-07-09
sync_commit_sha: 1b116e939f
sync_status: audit-ready
epic: Token-Economy (D 4/4 → 100% complete; A=ACCOUNTING + B=ROUTING + C=VERIFY-DIET + D=this)
ac_total: 9/9 MUST-PASS PASS
coverage_internal_runtime: 90.5% of statements (>=85% target)
lint_internal_runtime: 0 issues
cross_platform_build: darwin=0 windows=0
spec_lint_gate3: 0 errors (StatusGitConsistency warning resolved by in-progress→completed transition)
3_phase_close: plan(93c38003b) → run(032eb9733 M1 + 841843a9b M2) → sync(this commit)
race_note: shared-checkout parallel session completed M1+M2 first; this orchestrator-direct session independently verified AC 9/9 on HEAD (go test/cover/lint/vet + AC grep), discarded its duplicate impl, performed stash-drop-accident recovery (git stash apply 131aaa339 exit 0, 0 file loss — 11 parallel-session files + moai-easy.md restored)
```

---

## §F Phase 0.95 Mode Selection

Orchestrator mode selection logged (run-phase executed by parallel session as Mode 5; this orchestrator-direct sync session records the classification retroactively).

### Decision: sub-agent (Mode 5)

Tier M + coding-heavy → Mode 5 (sub-agent sequential) per Anthropic's coding-task parallelism caveat. M1 (Go graceful-abort via TDD RED-GREEN-REFACTOR) and M2 (doctrine persistence extension) carry a sequential dependency — M2's file-redirect contract states the persistence obligation M1's Go layer fulfills. A single `manager-develop` delegation (cycle_type=tdd) handled both milestones. Implementation Kickoff Approval: PASS (user-approved run-phase entry).

---

## History

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-07-09 | manager-spec | Initial draft — plan-phase §E.1 audit-ready signal emitted. §E.2-§E.4 + §F placeholder headings for run/sync-phase. |
