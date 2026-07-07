# SPEC-HOOK-FACTFORCE-ADVISORY-001 — Progress

> Living document tracking plan/run/sync phase evidence. Section structure per `.claude/rules/moai/development/spec-frontmatter-schema.md` § progress.md Section Map. The `§E.*` namespace is parser-load-bearing for era.go classification — see the Section Map before renaming any `§E.N` heading.

## §A. Plan-Phase Entry

| Field | Value |
|-------|-------|
| SPEC ID | SPEC-HOOK-FACTFORCE-ADVISORY-001 |
| Tier | S |
| Lifecycle | spec-anchored |
| Created | 2026-07-08 |
| Author (plan) | manager-spec |
| Supersedes | SPEC-HOOK-PREEDIT-INVESTIGATE-001 (status: completed → superseded) |
| Plan-phase commit | (combined into M1 run-phase commit — plan artifacts were untracked at run-phase entry) |
| plan-auditor iter-1 verdict | PASS 0.99 (skip-eligible; Phase 0.5 re-execution skipped per the 4-condition compound predicate) |
| Implementation Kickoff Approval | OBTAINED — user approved run-phase entry |

## §B. Run-Phase Entry

| Field | Value |
|-------|-------|
| Mode (Phase 0.95) | Mode 5 sub-agent (sequential, single-milestone Tier S) — see §F |
| cycle_type | tdd (quality.yaml development_mode default; Tier S single-M1) |
| Run-phase commit (M1) | (pending — populated post-commit) |
| Run-phase start timestamp | 2026-07-08T21:09:53Z (make build timestamp) |

## §C. Sync-Phase Entry

| Field | Value |
|-------|-------|
| Sync-phase commit | (pending) |
| Sync-phase start timestamp | (pending) |
| CHANGELOG entry added | (pending) |

## §D. Open Blockers

(none — plan-phase artifact creation is the current step; no blockers surfaced during authoring)

## §E.1 Plan-phase Audit-Ready Signal

```yaml
plan_status: audit-ready
plan_complete_at: 2026-07-08T00:00:00Z
plan_artifacts:
  - spec.md (12 canonical frontmatter fields, 12 REQ-FA requirements in GEARS notation, 7 Out-of-Scope entries, 2 new NFRs: jq-prohibition + systemMessage-validity)
  - plan.md (single M1 milestone, file-by-file change list with Template-First ordering, 3 change-sites documented in §F.2, hook-independence.md Mode G row wording change in §F.3)
  - acceptance.md (15 AC-FA ACs in Given-When-Then format, 13 MUST + 2 SHOULD, 10 edge cases, quality gate criteria, Definition of Done)
  - progress.md (this file)
plan_tier: S
plan_loc_estimate: ~40 LOC net (3 change-sites in hook script ~20 LOC + Mode G row 1-word substitution + mirrors)
plan_files_affected: 5 (template hook + template hook-independence.md + local hook + local hook-independence.md + make build)
plan_runtime_changes:
  go_changes: false
  settings_json_changes: false
  template_first_ordering: true
  additive_only: true
  fail_open_default: true
  jq_free: true
plan_audit_preconditions:
  spec_lint_clean: pending
  ears_gears_compliant: true
  exclusions_section_present: true
  frontmatter_12_fields: true
plan_auditor_iter: 0
plan_supersedes: SPEC-HOOK-PREEDIT-INVESTIGATE-001
```

## §E.2 Run-phase Evidence

### Files changed (M1, Template-First ordering)

| # | Path | Change |
|---|------|--------|
| 1 | `internal/template/templates/.claude/hooks/moai/gateguard-fact-force.sh` | 3 change-sites: header comment (gate→advisory), fail-open comment (removed "ONLY exit 2"), §11 emit-and-block→emit-advisory-exit-0 (awk JSON escape + `printf` systemMessage + `exit 0`) |
| 2 | `make build` | catalog.yaml regenerated (hashes shifted) + binary recompiled; exit 0 |
| 3 | `.claude/hooks/moai/gateguard-fact-force.sh` | local mirror of #1 (byte-identical) |
| 4 | `internal/template/templates/.claude/rules/moai/development/hook-independence.md` | §3 Mode G row rationale: `"no gate"` → `"no advisory notice"` |
| 5 | `.claude/rules/moai/development/hook-independence.md` | local mirror of #4 (byte-identical) |
| — | `acceptance.md` §A AC-FA-015 row | D1 fix: Requirement `§C.2` → `REQ-FA-005 + §C.2` |
| — | `plan.md` §A.3 items #4/#5 | D2 fix: LOC-delta `"1-word change"` → `"rationale-column wording change"` |

### AC PASS/FAIL matrix (15/15)

| AC | Status | Evidence |
|----|--------|----------|
| AC-FA-001 | PASS | first Edit → exit 0 + stdout JSON containing `First-edit advisory`, `IMPORTERS`, `DATA SCHEMAS`, `USER INSTRUCTION`; state file created |
| AC-FA-002 | PASS | second Edit same (session,path) → exit 0, empty stdout |
| AC-FA-003 | PASS | different session_id same path → advisory emitted; 2 state files |
| AC-FA-004 | PASS | `MOAI_FACT_FORCE=off` → exit 0, empty stdout, skip-log line appended, 0 state files |
| AC-FA-005 | PASS | `grep 'exit 2'` rc=1 (zero matches, both copies); `grep AskUserQuestion` rc=1 |
| AC-FA-006 | PASS | 260–620ms << 5000ms timeout (5s MUST met); <100ms soft-target not met (process-spawn overhead inherited from predecessor shell architecture) |
| AC-FA-007 | PASS | settings.json NOT touched (git diff empty) |
| AC-FA-008 | PASS | no PostToolUse registration added (settings.json unchanged) |
| AC-FA-009 | PASS | subagent payload (agent_id present) → same per-session logic; state keyed by session_id (functional parity with AC-FA-001/002) |
| AC-FA-010 | PASS | `tool_name=Bash` → exit 0, empty stdout, 0 state files |
| AC-FA-011 | PASS | `diff` template↔local empty for BOTH hook + hook-independence.md |
| AC-FA-012 | PASS | state file mode `100600` (0o600), 1 line, JSON parses (keys: first_seen/path/session_id/via) |
| AC-FA-013 | PASS | Read → state created, no advisory; subsequent Write same path → no advisory (Read-as-investigation) |
| AC-FA-014 | PASS | `jq -e '.systemMessage \| type'` → `"string"` (awk escape produces valid single-line JSON) |
| AC-FA-015 | PASS | payload without session_id → exit 0, empty stdout, 0 state files (fail-open) |

### Build + parity verification

- `make build` → exit 0 (catalog.yaml regenerated, binary compiled)
- `bash -n` on both hook copies → clean
- template↔local `diff` → empty (hook + hook-independence.md, both byte-identical)
- `grep -c 'jq ' <hook>` → 0 (jq-free per §C.5 NFR)
- Go source untouched (`git diff --name-only | grep '\.go$'` → empty)

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_status: audit-ready
run_complete_at: 2026-07-08T21:20:00Z
run_commit_sha: (pending backfill — commit lands post-evidence)
ac_pass_count: 15
ac_fail_count: 0
ac_should_pass_with_debt:
  - AC-FA-006 (< 100ms soft-target; 5s hard timeout met)
preserve_list_post_run_count: 0  # no PRESERVE-list files touched
l44_pre_commit_fetch: n/a (L1 worktree isolation — see Notes)
l44_post_push_fetch: (pending — push state TBD)
new_warnings_or_lints_introduced: 0  # shell-only, no Go changes
cross_platform_build:
  go_changes: false
  settings_json_changes: false
total_run_phase_files: 7  # 2 template + 2 local mirror + acceptance.md + plan.md + progress.md (+ catalog.yaml build artifact)
m1_to_mN_commit_strategy: single-M1 commit (Tier S)
notes:
  - "L1 worktree isolation: implementation executed in .claude/worktrees/agent-a1a523d767c33dbb2 (branch worktree-agent-a1a523d767c33dbb2); commit/push logistics surfaced to orchestrator"
  - "Plan-phase artifacts were untracked at run-phase entry; M1 commit carries both plan + run artifacts (combined)"
```

## §F Phase 0.95 Mode Selection

**Decision: Mode 5 (sub-agent)** — sequential single-milestone Tier S delegation.

| Mode | Selected? | Rationale |
|------|-----------|-----------|
| 1 trivial | no | multi-file change (hook + hook-independence.md + SPEC artifacts), not a typo |
| 2 background | no | write-heavy implementation, not read-only |
| 3 agent-team | no | Tier S scope (≤3 source files); team prereqs not met |
| 4 parallel | no | single-domain (hook scripts), coding-heavy — Anthropic coding-task parallelism caveat |
| 5 sub-agent | **yes** | coding-heavy single-milestone Tier S; the simpler sequential mode suffices |
| 6 workflow | no | < 30 files, non-mechanical (3 distinct change-sites), coding-heavy |

Input parameters: tier=S, scope≈7 files, domains=1 (hook scripts), language-mix=shell+markdown, concurrency-benefit=LOW (sequential dependency: template→build→mirror).

## §E.4 Sync-phase Audit-Ready Signal

_(pending sync-phase — manager-docs populates)_
