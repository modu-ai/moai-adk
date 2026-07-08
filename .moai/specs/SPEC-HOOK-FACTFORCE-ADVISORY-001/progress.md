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
run_commit_sha: 3455b0ee8
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

- sync_status: complete (orchestrator-direct sync-phase — manager-docs 위임이 API "Prompt is too long" 실패로 orchestrator-direct fallback; shared-checkout 다중 세션 race 환경에서 pathspec commit으로 정밀 제어)
- sync_complete_at: 2026-07-08
- sync_commit_sha: (pending backfill — sync commit lands post; 확립된 convention per SPEC-BRAND-DIR-REMOVE-001 `2c637ece2` / SPEC-V3R6-AGENTIC-LOOP-CONFIG-001)
- changelog_entry_position: CHANGELOG.md `## [Unreleased]` > `### Changed` — SPEC-HOOK-FACTFORCE-ADVISORY-001 entry (exit-2 → exit-0 advisory rewrite summary + 15/15 AC PASS citation per §E.2 headline)
- frontmatter_status_transitions:
  - spec.md: in-progress → completed (이 sync commit에서 atomic 전환)
  - plan.md / acceptance.md / progress.md: YAML frontmatter 미보유 (`# Title` 직시작) → status field 없음, spec.md가 canonical frontmatter 소유자 per spec-frontmatter-schema.md (12 required fields는 spec.md 전용)
- updated_field_refresh: 2026-07-08 (이미 run-phase에서 2026-07-08 — 변경 불필요)
- sync_phase_deliverables:
  1. progress.md §E.3 `run_commit_sha` backfill: `3455b0ee8` (run commit `feat(SPEC-HOOK-FACTFORCE-ADVISORY-001): M1 exit-2 block → exit-0 advisory systemMessage`, origin/main confirmed via `git log --oneline origin/main | grep 3455b0ee8`)
  2. progress.md §E.4 (this section) filled with sync-phase audit-ready signal
  3. spec.md frontmatter status: in-progress → completed (이 sync commit, atomic)
  4. CHANGELOG.md `## [Unreleased]` > `### Changed` entry added
- observed_evidence_cross_refs:
  - §E.2 run-phase evidence: commit `3455b0ee8` — AC-FA-001..015 (15/15 AC PASS, `bash -n` clean both hook copies, template↔local byte-identical parity, `grep -c 'jq ' <hook>` = 0 jq-free)
  - §E.3 run-phase signal: `ac_pass_count: 15`, `ac_fail_count: 0`, shell-only no Go changes (`git diff -- '*.go'` empty), cross-platform build N/A (shell + markdown only)
  - orchestrator independent verification (7/7 PASS): V1 exit-2 제거 · V2 template==local parity · V3 jq-free comment only · V4 systemMessage + 4 substrings · V5 hook-independence Mode G row · V6 commit == origin/main · V7 spec lint clean
- b12_self_test_a (pre-emission grep): `grep -c 'SPEC-HOOK-FACTFORCE-ADVISORY-001' CHANGELOG.md` = 0 (pre-emission) → no duplicate entry, emission safe
- b12_self_test_b (AC count match): acceptance.md SSOT = 15 AC rows (AC-FA-001..015, §A AC Index); §E.2 headline = 15/15 AC PASS; CHANGELOG entry cites "15/15 AC PASS" matching acceptance.md authoritative count (NOT progress.md)
- b12_self_test_c (file path verification): 4 implementation file paths verified — `internal/template/templates/.claude/hooks/moai/gateguard-fact-force.sh` (template hook) + `.claude/hooks/moai/gateguard-fact-force.sh` (local mirror, byte-identical) + `internal/template/templates/.claude/rules/moai/development/hook-independence.md` Mode G row + `.claude/rules/moai/development/hook-independence.md` (local mirror); 실제 hook 파일 Read로 exit-2 제거 + §11 advisory emit + jq-free(awk escape) 확인 (plan.md 의존 않음)
- canary_compliance_check: N/A (advisory-only PreToolUse hook, no canary deployment surface)
- l44_pre_commit_fetch: `origin/main` = `e62d52e2b` (병렬 세션 TOKEN-VERIFY-DIET-001 plan-phase + INTERNAL-SECURITY-001 sync 완료 후 적재; divergence `0 0`, ff-pushable, working tree FACTFORCE dir clean)
- l44_post_push_fetch: (pending — orchestrator pushes post-sync)
- race_absorbed: shared-checkout 병렬 세션 race — manager-docs 위임 prompt-too-long 실패 중 병렬 세션 A(INTERNAL-SECURITY-001 sync f3193bac8~c0ac267b5, 별개 세션 f5a621d0) + 병렬 세션 B(TOKEN-VERIFY-DIET-001 plan-phase e62d52e2b) commit 적재; FACTFORCE 범위(SPEC dir + CHANGELOG.md FACTFORCE entry)와 불겹침 → pathspec commit으로 clean 흡수 (conflict 없음, feedback_shared_checkout_concurrent_commit_race 교훈 적용)
- residual_debt: 없음 (run-phase 15/15 AC 전부 PASS; SHOULD AC-FA-006 <100ms soft-target만 5s hard timeout met으로 PASS-WITH-DEBT — §E.2에 이미 투명 기록)
