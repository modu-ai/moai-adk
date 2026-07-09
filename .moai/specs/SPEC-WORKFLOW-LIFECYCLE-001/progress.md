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

_<pending run-phase — manager-develop 소관>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase — manager-develop 소관>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase — manager-docs 소관>_
