# progress.md — SPEC-V3R6-ORCH-IGGDA-001

> Plan-phase skeleton. §E.2–§E.4 are placeholder headings only (per the canonical progress.md §E skeleton generation protocol). This agent populates only §E.1 (plan-phase audit-ready signal); §E.2/§E.3 belong to manager-develop (run-phase) and §E.4 belongs to manager-docs (sync-phase).

---

## §E.1 Plan-phase Audit-Ready Signal

**SPEC-ID**: SPEC-V3R6-ORCH-IGGDA-001
**Tier**: L (5-artifact: spec.md + plan.md + acceptance.md + design.md + research.md)
**Status**: draft (plan-phase authored, awaits plan-auditor independent audit)
**Authored**: 2026-06-19 by manager-spec

**Artifacts**:
- `spec.md` — 28 GEARS REQs (REQ-IGGDA-001 through REQ-IGGDA-028) across 5 deliverables (D1–D5)
- `plan.md` — 6 milestones (M1–M6), Template-First, FROZEN-amend first
- `acceptance.md` — 42 ACs (41 MUST-PASS + 1 SHOULD-PASS; AC-005 counted as 5 sub-branches 005a–005e, AC-037 MUST-PASS-conditional-on-M5-Go-touch), both Path B branches covered (AC-IGGDA-004 auto-proceed + AC-IGGDA-005a–005e explicit-gate × 5)
- `design.md` — 4-phase architecture + Path B kickoff handling + Stop hook driver + bounded recursive loop + **§F FROZEN invariant analysis (load-bearing)**
- `research.md` — existing-component mapping (file:line for run.md, orchestration-mode-selection.md, runtime-recovery-doctrine.md, CLAUDE.local.md §19.1) + FROZEN lineage + Anthropic plan-editor mandate relationship + related SPEC survey

**Frontmatter schema**: 12 canonical fields present + `era: V3R6` + `tier: L` + `depends_on: [SPEC-AUTONOMY-RUN-GOAL-001, SPEC-V3R6-WORKFLOW-EFFORT-MAP-001, SPEC-V3R6-ORCH-INTERRUPT-LEDGER-001]`.

**SPEC ID self-check**: `decomposition: SPEC ✓ | V3R6 ✓ | ORCH ✓ | IGGDA ✓ | 001 ✓ → PASS` (regex `^SPEC(-[A-Z][A-Z0-9]*)+-\d{3}$`).

**Out of Scope**: 5 `### Out of Scope — <topic>` H3 sub-headings present (Go runtime-layer hook evolution; Agent Teams integration; Version preflight closure; `[ZONE:Frozen]` marker removal; Cross-SPEC IGGDA rollout).

**Plan-auditor scrutiny focus** (per FROZEN-amend risk): design.md §F (FROZEN invariant analysis) + acceptance.md AC-IGGDA-004/005a–005e (both Path B branches).

**Q4 resolved (pre-flight finding)**: `moai spec audit --filter-spec=<SPEC-ID>` does NOT exist today. Available flags: `--filter-era`, `--include-grandfathered`, `--json`, `--strict`. plan.md M5 will add `--filter-spec` (additive Go flag in `internal/spec/audit.go` + `internal/cli/spec_audit.go`). Until M5 lands, the D4 Stop hook driver falls back to JSON-parsing the full `moai spec audit --json` output client-side (filtering by `spec_id` in the `drift_findings[]` array).

**Canonical lint status**: `moai spec lint .moai/specs/SPEC-V3R6-ORCH-IGGDA-001/spec.md` → `✓ No findings — all SPEC documents are valid`. (Multi-file lint invocation is non-canonical — both catalog precedents SPEC-AUTONOMY-RUN-GOAL-001 and SPEC-V3R6-AGENT-TEAM-REBUILD-001 produce findings under multi-file invocation; the linter is designed for single-file spec.md = SSOT invocation.)

**Defect-fix note (2026-06-19, plan-auditor iter-1 PASS-WITH-DEBT 0.82 follow-up)**: 7 defects applied — D1 BLOCKING (acceptance.md §E MUST-PASS count 37→41), D2 SHOULD-FIX (AC-037 removed from SHOULD-PASS line), D3 BLOCKING (plan.md M→AC mapping re-authored; M5 AC-025→AC-037; 9 orphaned ACs assigned), D4 SHOULD-FIX (M1 explicit 005a-e enumeration), D5 SHOULD-FIX (AC-013 Evidence grep target `iggda-phase-driver.sh`), D7 SHOULD-FIX (AC-005d title EXPLICIT-GATE→RETURN-TO-PHASE-0), D10 SHOULD-FIX (design.md §F.6 multi-round-Socratic preventive complement documented). 3 MINOR debt tolerated per user decision: D6 (spec.md:189 bounded timeout 30s), D8 (design.md:227 pgp keyword overlap), D9 (research.md:11 run.md line-count claim).

---

## § Phase 0.95 Mode Selection

**Decision**: `sub-agent` (Mode 5, sequential per milestone)

**Input parameters**:
- tier: L
- scope: ~10-15 files (orchestration-mode-selection.md, run.md, iggda-phase-driver.sh NEW, regression test NEW, internal/spec/audit.go, internal/cli/spec_audit.go, template mirrors, docs)
- domain count: ~5 (rules/markdown, shell hook, Go source, Go test, docs/drift)
- file language mix: markdown + shell + Go (mixed)
- concurrency benefit: LOW (coding-heavy, dependent milestone chain M1→M6)
- Agent Teams prereqs: workflow.team.enabled=true + CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS=1 (moot — coding-heavy overrides)

**Mode evaluation**:

| Mode | Selected? | Rationale |
|------|-----------|-----------|
| trivial (1) | NO | semantic changes, multiple files, FROZEN-amend |
| background (2) | NO | Write tasks, not read-only |
| agent-team (3) | NO | coding-heavy; Anthropic coding-task parallelism caveat |
| parallel (4) | NO | coding-heavy, NOT research-heavy; dependent milestones |
| workflow (6) | NO | semantic/new-code/multi-rule, NOT mechanical-uniform; coding-heavy |
| sub-agent (5) | YES | default fallback; coding-heavy + dependent milestone chain |

**Justification**: IGGDA M1-M6 are coding-heavy (Go source M5 + Go tests M4) + semantic rule edits (M1 FROZEN-amend in orchestration-mode-selection.md, M2 recursive loop in run.md) + new mechanisms (M3 Stop hook driver). Per Anthropic's coding-task parallelism caveat ("most coding tasks involve fewer truly parallelizable tasks than research"), sequential sub-agent is the correct default. Mode 6 workflow rejected — M1-M6 are semantic/new-code/multi-rule, not a single uniform mechanical transform. Milestones are dependent (M1 FROZEN-amend must land before M2 builds on the amended predicate; M5 `--filter-spec` flag is an M3 Stop-hook dependency). Implementation Kickoff Approval PASSED (Tier L explicit gate, user-approved 2026-06-19).

---

## §E.2 Run-phase Evidence

**Run-phase**: M1–M6 (manager-develop, cycle_type=ddd for rule amendments + cycle_type=tdd for M5 Go). 2026-06-19.

**Commits** (on branch `worktree-agent-a6a00e00dfdb5f0ee`, to be merged to main by orchestrator):
- `11f213d3a` — M1 IGGDA safe-condition predicate + 4-phase pipeline (FROZEN-amend)
- `d4fb1e861` — M2 bounded recursive self-diagnosis loop (D3)
- `2ad42d830` — M3 moai-aware Stop hook driver (D4) + M2 run.md leak-fix
- `e5df49495` — M5 moai spec audit --filter-spec flag (Go, TDD)
- (this commit) — M4 independent-audit preservation regression guard + M6 close-prep

### E.2.1 AC PASS/FAIL matrix

| AC | Milestone | Status | Verification |
|----|-----------|--------|--------------|
| AC-IGGDA-001 (gate still issued) | M1 | PASS | `grep -c "AskUserQuestion" orchestration-mode-selection.md` ≥ 1; gate present in §H |
| AC-IGGDA-002 (preferences drained) | M1 | PASS | §H.2 + §I.1 Phase 0 drain documented |
| AC-IGGDA-003 (non-approval blocks) | M1 | PASS | §H.2 "User veto is never overridden" |
| AC-IGGDA-004 (auto-proceed branch) | M1 | PASS | §H.1 condition (a)-(d) + §H.2 auto-proceed branch |
| AC-IGGDA-005a (explicit-gate Tier L) | M1 | PASS | §H.1 condition (c) FAIL → explicit-gate |
| AC-IGGDA-005b (security/payment/critical keywords) | M1 | PASS | §H.3 keyword list (security/payment/critical) |
| AC-IGGDA-005c (--pr destructive scope) | M1 | PASS | §H.3 destructive scope markers |
| AC-IGGDA-005d (return-to-Phase-0) | M1 | PASS | §H.1 condition (a) FAIL → return-to-Phase-0 |
| AC-IGGDA-005e (plan-auditor FAIL/INCONCLUSIVE) | M1 | PASS | §H.1 condition (b) FAIL → surface-to-user |
| AC-IGGDA-006 (decision logged) | M1 | PASS | §H.4 IGGDA Kickoff Predicate logging contract |
| AC-IGGDA-007 (Phase 0.5 not skipped) | M1 | PASS | §I.2 phase ordering invariant |
| AC-IGGDA-008 ([ZONE:Frozen] marker preserved) | M1 | PASS | `grep -c "[ZONE:Frozen]" orchestration-mode-selection.md` = 5 |
| AC-IGGDA-009 (4-phase pipeline in order) | M1 | PASS | §I.1 Phase 0→1→2→3 table + §I.2 ordering invariant |
| AC-IGGDA-010 (Phase 1 safe-transition) | M1/M3 | PASS | §I.1 Phase 1 + §I.3 Stop hook driver contract |
| AC-IGGDA-011 (Phase 2 safe-transition) | M1/M3 | PASS | §I.1 Phase 2 + §I.3 driver |
| AC-IGGDA-012 (Phase 3 safe-transition) | M1/M3 | PASS | §I.1 Phase 3 + §I.3 driver |
| AC-IGGDA-013 (reads progress.md + moai spec audit) | M3/M5 | PASS | grep confirms 9 matches; moai spec audit --filter-spec wired (M5) |
| AC-IGGDA-014 (Recovery-Signal Carve-Out exit 0) | M3 | PASS | grep confirms prompt_too_long/max_output_tokens/media_size/compact in driver |
| AC-IGGDA-015 (graceful degradation /goal unavailable) | M3 | PASS | driver self-gates (exit 0 on missing SPEC_ID/progress.md/moai) |
| AC-IGGDA-016 (exit codes + JSON, NEVER AskUserQuestion) | M3 | PASS | ledger_note field present; subagent-boundary grep 0 matches |
| AC-IGGDA-017 (subagent boundary grep 0) | M3 | PASS | grep -rn AskUserQuestion hooks → 0 (excluding comments) |
| AC-IGGDA-018 (timeout ≤5s) | M3 | PASS | settings.json.tmpl `"timeout": 5` |
| AC-IGGDA-019 (Windows stub) | M3 | PASS (SHOULD) | iggda-phase-driver_windows.cmd exists, exit 0 |
| AC-IGGDA-020 (no runtime-path writes) | M3 | PASS | driver writes only to MOAI_HOOK_STDERR_LOG |
| AC-IGGDA-021 (max-3-iteration bound) | M2 | PASS | run.md §3 HARD bound + escalation |
| AC-IGGDA-022 (mechanical → DIAGNOSE-PATCH-VERIFY) | M2 | PASS | run.md §1 classification + §2 pattern |
| AC-IGGDA-023 (semantic escalates immediately) | M2 | PASS | run.md §1 + §4 HARD immediate escalate |
| AC-IGGDA-024 (5 circuit-breaker invariants) | M2 | PASS | run.md §5 all 5 enumerated |
| AC-IGGDA-025 (iteration log to progress.md) | M2 | PASS | run.md §6 logging contract |
| AC-IGGDA-026 (forbidden paths) | M2 | PASS | run.md §7 .env/credentials/ci-watch/scope |
| AC-IGGDA-027 (plan-auditor fresh context) | M1/M4 | PASS | §J.1 + regression guard PASS |
| AC-IGGDA-028 (sync-auditor fresh context) | M1/M4 | PASS | §J.1 + regression guard PASS |
| AC-IGGDA-029 (FAIL/INCONCLUSIVE halts) | M1/M4 | PASS | §J.2 + regression guard PASS |
| AC-IGGDA-030 (self-audit vs independent-audit) | M1/M4 | PASS | §J.3 + regression guard PASS |
| AC-IGGDA-031 (IGGDA-completeness terminal gate) | M1/M3 | PASS | §I.1 Phase 3 terminal predicate (0 MUST-FIX + sync-auditor ≥ threshold + git clean) |
| AC-IGGDA-032 (preferences before autonomy) | M1 | PASS | §I.1 Phase 0 + §H.2 |
| AC-IGGDA-033 (flat hierarchy) | M2 | PASS | run.md §8 flat-hierarchy (orchestrator sole spawner) |
| AC-IGGDA-034 (background read-only) | M2 | PASS | run.md §2 sub-agent foreground |
| AC-IGGDA-035 (no destructive auto) | M1 | PASS | §H.2 + run.md §7 forbidden paths |
| AC-IGGDA-036 (spec-lint clean) | M6 | PASS | moai spec lint → 0 errors (1 StatusGitConsistency warning — expected) |
| AC-IGGDA-037 (Go test suite + coverage ≥85%) | M5 | PASS | go test ./internal/spec/... PASS, 87.6% coverage |
| AC-IGGDA-038 (template neutrality CI) | M6 | PASS | go test ./internal/template/... -run TestTemplateNeutralityAudit PASS |

**Total**: 41 MUST-PASS PASS + 1 SHOULD-PASS PASS (AC-019 Windows stub).

### E.2.2 Cross-platform build

```
$ go build ./...                          → exit 0
$ GOOS=windows GOARCH=amd64 go build ./... → exit 0
```

### E.2.3 Lint status

```
$ golangci-lint run --timeout=2m → 0 issues (NEW = 0; pre-existing baseline = 0)
```

### E.2.4 Subagent-boundary grep (C-HRA-008 family)

```
$ grep -rn 'AskUserQuestion\|mcp__askuser' internal/template/templates/.claude/hooks/moai/iggda-phase-driver.sh internal/template/templates/.claude/hooks/moai/handle-iggda-phase-driver.sh.tmpl internal/cli/spec_audit.go internal/spec/audit.go | grep -v "_test.go" | grep -v "^[^:]*:[0-9]*:[ \t]*//" | grep -vE "MUST NOT invoke|asymmetric|translates|NEVER done here"
→ 0 matches (PASS)
```

### E.2.5 Template neutrality

```
$ go test ./internal/template/... -run TestTemplateNeutralityAudit → PASS
$ go test ./internal/template/... (full suite incl. leak test) → PASS
```

### E.2.6 [ZONE:Frozen] marker preservation proof

```
$ grep -c "\[ZONE:Frozen\]" internal/template/templates/.claude/rules/moai/workflow/orchestration-mode-selection.md → 5
$ grep -c "Implementation Kickoff Approval" internal/template/templates/.claude/rules/moai/workflow/orchestration-mode-selection.md → 18
$ grep -n "Before any run-phase autonomy\|score-independent" internal/template/templates/.claude/skills/moai/workflows/run.md → lines 122, 124 (preserved verbatim)
```

### E.2.7 Pre-existing baseline failures (NOT caused by this SPEC)

- `internal/cli TestSpecAudit_JSONSchema_DriftFindings` — legacy `Y_Y_Y_Y_StatusDrift` 5-marker fixture; broke during SPEC-V3R6-LIFECYCLE-REDESIGN-001 lifecycle migration. Verified pre-existing via `git stash` against `2ad42d830` (fails identically before M5). Out of scope.
- `internal/cli TestSpecAudit_StrictMode_ExitsNonZeroOnDrift` — same legacy fixture. Out of scope.

### E.2.8 Gaps / Residual-risk

- **Gaps**: sync-phase (§E.4) not populated — owned by manager-docs (out of run-phase scope). 3-phase close not executed (manager-docs owns sync-phase). README/CHANGELOG not updated (manager-docs).
- **Residual-risk**: the IGGDA safe-condition predicate (§H) is a policy-layer rule amendment; its runtime enforcement depends on the orchestrator reading the rule at gate-evaluation time. The M3 Stop hook driver is the mechanical layer; full orchestrator-side predicate evaluation wiring is the IGGDA-preflight follow-up (FL-3 candidate). The FROZEN-amend is reversible — restoring Path A (deleting §H/§I/§J) reverts the amendment without affecting the [ZONE:Frozen] marker.

---

## §E.3 Run-phase Audit-Ready Signal

**Run-phase status**: audit-ready (all MUST-PASS ACs PASS, 41/41; 1 SHOULD-PASS PASS).
**Run-phase complete at**: 2026-06-19
**Run commit sha range**: `11f213d3a` (M1) → `e5df49495` (M5) + this M4/M6 commit
**Run commit sha (HEAD)**: _(this commit)_
**run_status**: audit-ready
**ac_pass_count**: 41 (MUST-PASS) + 1 (SHOULD-PASS) = 42
**ac_fail_count**: 0
**preserve_list_post_run_count**: 0 (no PRESERVE-list items regressed; [ZONE:Frozen] marker preserved at 5 occurrences; run.md:122,124 preserved verbatim)
**l44_pre_commit_fetch**: not applicable (1-person OSS Hybrid Trunk, worktree-isolated agent, no parallel session race detected via `git fetch origin main` — local-ahead-only)
**l44_post_push_fetch**: deferred (orchestrator manages push to main)
**new_warnings_or_lints_introduced**: 0 (golangci-lint 0 issues; spec-lint 0 errors; 1 StatusGitConsistency WARNING is expected — fires because git history has no `in-progress` commit yet; clears once commits land)
**cross_platform_build.linux**: exit 0
**cross_platform_build.windows**: exit 0
**total_run_phase_files**: 11 (orchestration-mode-selection.md template+local, run.md template+local, iggda-phase-driver.sh + wrapper .tmpl + Windows .cmd, settings.json.tmpl, iggda-audit-preservation-guard.sh, audit.go, audit_test.go, spec_audit.go, spec.md frontmatter, progress.md §E.2/§E.3)
**m1_to_mN_commit_strategy**: per-milestone Conventional Commits (`feat(SPEC-V3R6-ORCH-IGGDA-001): M{N} <subject>`), 5 commits (M1, M2, M3+leakfix, M5, M4+M6)

---

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase — manager-docs populates>_

---
