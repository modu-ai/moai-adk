# SPEC-WORKFLOW-LIFECYCLE-001 — Progress

> 본 파일은 lifecycle-tracking artifact. manager-spec/develop/docs가 phase별로 소유. §E.* namespace는 `internal/spec/era.go`가 parser-load-bearing으로 사용 — 리터럴 heading 절대 변경 금지 (spec-frontmatter-schema.md § progress.md Section Map).

## §E.1 Plan-phase Audit-Ready Signal

```yaml
plan_status: audit-ready
plan_complete_at: 2026-07-09T17:40:00Z
plan_tier: L
plan_artifact_count: 6
plan_artifacts:
  - spec.md
  - plan.md
  - acceptance.md
  - design.md
  - research.md
  - progress.md
plan_req_count: 12
plan_ac_count: 30
plan_mp_results:
  MP-1: "REQ-WFL-001..012 연속, 중복/공백 없음 (spec.md §B)"
  MP-2: "12 REQ 전부 GEARS compound-clause (Ubiquitous/When/Where/While) — plan-auditor replay 시 확인"
  MP-3: "12 canonical fields + tier: L + tags + era 부재(H-4 auto-detect 예상) — plan-auditor replay 시 확인"
  MP-4: "언어 중립 — Tier-differentiated 입력 계약이 어떤 언어도 PRIMARY로 승격하지 않음"
plan_frontmatter_check:
  12_canonical_fields: present
  snake_case_alias_check: "no created_at/updated_at/labels/spec_id"
  tier_field: "L (explicit)"
  tags_format: "comma-separated string"
plan_self_check:
  spec_id_regex: "PASS (decomposition: SPEC ✓ | WORKFLOW ✓ | LIFECYCLE ✓ | 001 ✓ → PASS)"
  spec_id_command_evidence: "[[ \"$ID\" =~ ^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$ ]] && echo PASS → PASS"
plan_key_design_decisions:
  R1_delta_spec: "completed → in-progress 재전이 + amendment_of: Optional 필드 + HISTORY ## Amendments sub-section (새 amended 상태 추가 안함)"
  R2_depends_on: "Phase 0.5 sub-step 'Depends_on Pre-flight Check' 확장 (0.6 신설 아님) + status: completed 단일 충족 조건 + AskUserQuestion 3-option blocker"
  R3_tier_l: "3면 SSOT 동기화 — spec-frontmatter-schema.md tier: 필드에 5-file 명시 + plan-auditor.md Input Contract를 Tier-differentiated로 재작성 + spec-workflow.md hash subject list를 Go 정합(4-file)으로 명문화"
plan_out_of_scope_h3_count: 5
plan_out_of_scope_h3_topics:
  - "Out of Scope — Go 코드 동작 변경"
  - "Out of Scope — P2 위생 백로그"
  - "Out of Scope — depends_on cycle detection / topological sort"
  - "Out of Scope — amendment의 자동 감지 훅"
  - "Out of Scope — Tier S/M에 대한 design/research Optional 허용"
plan_constraints_honored:
  doc_only: true
  no_go_code_changes: true
  template_mirror_required: true
  credentials_yml_untouched: true
```

## §E.2 Run-phase Evidence

Path abbreviations: `L-SFS` = `.claude/rules/moai/development/spec-frontmatter-schema.md`, `L-WF` = `.claude/rules/moai/workflow/spec-workflow.md`, `L-PA` = `.claude/agents/moai/plan-auditor.md`, `L-MS` = `.claude/agents/moai/manager-spec.md`, `T-*` = template mirror.

### R1 — delta-spec lifecycle

| AC | REQ | Command | Output | Status |
|----|-----|---------|--------|--------|
| AC-WFL-001 | WFL-001 | `grep -c 'amendment_of' L-SFS` | `2` | PASS |
| AC-WFL-002 | WFL-001 | `grep -c 'completed → in-progress' L-SFS` + `grep -c 'amendment' L-SFS` | `3`, `3` | PASS |
| AC-WFL-003 | WFL-002 | `grep -c '## Amendments' L-MS` | `1` | PASS |
| AC-WFL-004 | WFL-002 | `grep -c 'prior_completed' L-SFS` | `3` | PASS |
| AC-WFL-005 | WFL-003 | `grep -c 'in-place amendment' L-SFS` | `3` | PASS |
| AC-WFL-006 | WFL-004 | `grep -c 'amendment transition' L-WF` | `1` | PASS |

### R2 — depends_on run-phase enforcement

| AC | REQ | Command | Output | Status |
|----|-----|---------|--------|--------|
| AC-WFL-007 | WFL-005 | `grep -c 'Depends_on Pre-flight' L-WF` | `2` | PASS |
| AC-WFL-008 | WFL-005 | `grep -c 'depends_on' L-WF` + `grep -c 'Phase 0.5' L-WF` | `4`, `9` | PASS |
| AC-WFL-009 | WFL-006 | `grep -c 'fulfillment' L-WF` | `1` | PASS |
| AC-WFL-010 | WFL-006 | `grep -c 'status: completed' L-WF` | `1` | PASS |
| AC-WFL-011 | WFL-007 | `grep -c 'ignore-deps' L-WF` | `2` | PASS |
| AC-WFL-012 | WFL-007 | `grep -c 'depends-on-override.log' L-WF` | `2` | PASS |

### R3 — Tier L artifacts + plan-auditor input contract + hash alignment

| AC | REQ | Command | Output | Status |
|----|-----|---------|--------|--------|
| AC-WFL-013 | WFL-008 | `grep -c 'design\.md' L-SFS` + `grep -c 'research\.md' L-SFS` | `1`, `1` | PASS |
| AC-WFL-014 | WFL-008 | `grep -c 'Tier L' L-SFS` | `2` | PASS |
| AC-WFL-015 | WFL-009 | `grep -c 'Tier-differentiated input contract' L-PA` | `2` | PASS |
| AC-WFL-016 | WFL-009 | `grep -c 'Tier L: design.md + research.md are required inputs' L-PA` | `1` | PASS |
| AC-WFL-017 | WFL-009 | `grep -c 'design\.md' L-PA` + `grep -c 'research\.md' L-PA` | `1`, `1` | PASS |
| AC-WFL-018 | WFL-010 | `grep -c 'planArtifactNames' L-WF` | `1` | PASS |
| AC-WFL-019 | WFL-010 | `grep -c 'tasks\.md' L-WF` | `3` | PASS |
| AC-WFL-020 | WFL-010 | `grep -c 'manual-skip judgment inputs' L-WF` | `1` | PASS |
| AC-WFL-021 | WFL-010 | `grep -c 'V3R4' L-WF` | `1` | PASS |
| AC-WFL-022 | WFL-011 | `grep -c 'cache-invalidating event' L-WF` | `1` | PASS |

### Cross-cutting — Template Mirror synchronization

| AC | REQ | Command | Output | Status |
|----|-----|---------|--------|--------|
| AC-WFL-023 | WFL-012 | `grep -c 'Tier-differentiated input contract' T-PA` + `grep -c 'design\.md' T-PA` + `grep -c 'research\.md' T-PA` | `2`, `1`, `1` | PASS |
| AC-WFL-024 | WFL-012 | `grep -c 'amendment_of' T-SFS` + `grep -c 'in-place amendment' T-SFS` | `2`, `3` | PASS |
| AC-WFL-025 | WFL-012 | `grep -c 'Depends_on Pre-flight' T-WF` + `grep -c 'ignore-deps' T-WF` + `grep -c 'manual-skip judgment inputs' T-WF` + `grep -c 'cache-invalidating event' T-WF` | `2`, `2`, `1`, `1` | PASS |
| AC-WFL-026 | WFL-012 | `grep -c 'amendment_of' T-MS` + `grep -c '## Amendments' T-MS` | `1`, `1` | PASS |
| AC-WFL-027 | WFL-012 | `make build; echo "exit=$?"` | `exit=0` | PASS |
| AC-WFL-028 | WFL-012 | `grep -rc 'SPEC-WORKFLOW-LIFECYCLE\|REQ-WFL' T-{agents,rules}/... \| awk sum` | `0` | PASS |
| AC-WFL-029 | ALL | `command -v moai; moai spec lint; grep -c 'ERROR' (self)` | `tool=0`, `ERROR_count=0` | PASS |
| AC-WFL-030 | WFL-012 | `git diff --name-only HEAD -- credentials.yml` + `git diff --cached --name-only -- credentials.yml` | `0`, `0` | PASS |

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-07-09T12:47:00Z
run_commit_sha: 477e700de
run_status: audit-ready
ac_pass_count: 30
ac_fail_count: 0
preserve_list_post_run_count: 0
l44_pre_commit_fetch: "0 0 (synced with origin/main at run start)"
l44_post_push_fetch: "not pushed (worktree branch — orchestrator handles push)"
new_warnings_or_lints_introduced: 0
cross_platform_build:
  go_build_exit: 0
  make_build_exit: 0
total_run_phase_files: 8
m1_to_mN_commit_strategy:
  M1: "b00cc6380 — spec-frontmatter-schema.md + manager-spec.md + SPEC dir (draft→in-progress)"
  M2: "bcb35f7d1 — spec-workflow.md (Depends_on Pre-flight Check)"
  M3: "e558f1e0f — spec-frontmatter-schema.md + plan-auditor.md (Tier L + Tier-differentiated input contract)"
  M4: "b3afb98bd — spec-workflow.md (hash subject list + cache-invalidating event)"
  M5: "477e700de — mirror 4 files + make build + progress.md §E.2/§E.3"
run_phase_scope:
  live_files_edited: 4
  mirror_files_edited: 4
  go_code_changes: 0
  doc_only: true
  credentials_yml_touched: false
  neutrality_violations: 0
```

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: 2026-07-09
sync_commit_sha: 5f2e5738c   # backfilled from pending-backfill-sync placeholder in follow-up commit (amend FORBIDDEN — self-referential hazard avoided per repo convention); sync commit subject: docs(SPEC-WORKFLOW-LIFECYCLE-001): sync-phase artifacts + 3-phase close
sync_status: audit-ready
b12_self_test_a: "grep -c 'SPEC-WORKFLOW-LIFECYCLE-001' CHANGELOG.md → 0 (pre-emission, no duplicate — PASS)"
b12_self_test_b: "acceptance.md SSOT AC matrix rows = 30 (grep -cE '^\\| AC-WFL-[0-9]+ \\|') — CHANGELOG entry claims 30/30, match — PASS"
b12_self_test_c: "ls 8/8 claimed paths exist (live 4 + mirror 4) — PASS"
changelog_entry_position: "[Unreleased] ### Added — 1st bullet (before SPEC-CADENCE-BRIDGE-001)"
frontmatter_status_transitions:
  spec_md: "in-progress → completed (merged 3-phase close on this single sync commit; updated: 2026-07-09 unchanged-correct)"
  plan_md: "n/a — no YAML frontmatter (Tier L header-only artifact)"
  acceptance_md: "n/a — no YAML frontmatter (Tier L header-only artifact)"
3_phase_close: "plan(v0.1.0, plan-auditor iter-1 PASS-WITH-DEBT 0.96 skip-eligible per .moai/reports/plan-audit/SPEC-WORKFLOW-LIFECYCLE-001-review-1.md) → run(b00cc6380 M1 / bcb35f7d1 M2 / e558f1e0f M3 / b3afb98bd M4 / 477e700de M5, merge 3c410a2db, NOT pushed — orchestrator handles push after sync-auditor review) → sync(this commit)"
ac_total: "30/30 PASS (§E.2 matrix; independently re-verified by orchestrator in main tree post-merge)"
neutrality: "AC-WFL-028 grep sum 0 — no internal-token (SPEC-WORKFLOW-LIFECYCLE / REQ-WFL) leak into template tree"
catalog_yaml_debt: "pending — internal/template/catalog.yaml regenerated by make build with parallel-session file hashes (manager-docs.md/manager-git.md/goal-directive.md mirrors + llm.yaml/workflow.yaml); NOT committed in this sync (pathspec isolation to avoid sweeping parallel in-progress hashes) — orchestrator commits catalog.yaml separately after all parallel sessions close"
run_commit_sha_backfill: "477e700de (M5 terminal run-phase commit; backfilled from pending-backfill-M5 placeholder in this sync commit — SHA known only post-merge per self-referential hazard convention)"
```
