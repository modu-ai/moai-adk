# Progress — SPEC-V3R6-DOCS-DOCSITE-001

> Lifecycle progress tracker. Plan-phase artifact.
> §E.1은 manager-spec이 plan-phase에서 채운다. §E.2-§E.5는 placeholder heading만 (run/sync/Mx phase에서 채움).

**SPEC ID**: SPEC-V3R6-DOCS-DOCSITE-001
**Tier**: L
**Status**: draft (plan-phase complete, awaiting plan-auditor + Implementation Kickoff Approval)

---

## §A. Phase Status

| Phase | Status | Owner |
|-------|--------|-------|
| Plan | complete | manager-spec |
| Plan-audit | pending | plan-auditor |
| Run | pending (awaiting §19.1 Implementation Kickoff Approval) | manager-develop |
| Sync | pending | manager-docs |
| Mx | pending | orchestrator-direct / manager-docs |

---

## §B. Artifact Set

- [x] `spec.md` — GEARS requirements (REQ-DOCSITE-001..008, iter-2: +008 language count)
- [x] `plan.md` — Tier L 6-milestone plan (M1..M6), iter-2: M2 mermaid structural + M4 language count
- [x] `acceptance.md` — 11 ACs (AC-DOCSITE-001..011), iter-2: digit-boundary protected + live-grep evidence inline
- [x] `research.md` — 4-locale drift inventory (D-001..D-009 + D-001b/D-004b + ADJ-1), iter-2 D2 re-survey
- [x] `progress.md` — this file (§E.1 plan-phase signal)

---

## §C. Plan-phase Decisions

1. **Tier L** — ≈10-11 실제 편집 파일 + 4-locale 동시 조정 + 6 facts-axis 교차 + no-IA-redesign + primary-source 추적 요건 결합 (D9: "24 pages / 4× exceed" 과장 제거).
2. **6 docs-truth axes scope** — "31 skills"는 OUT OF SCOPE (uncoupled). "18 languages"는 **IN SCOPE** (REQ-008, FROZEN rule CONST-V3R2-004 위반). D6.
3. **M2 우선순위** — `core-concepts/what-is-moai-adk.md` **ja "28" 5곳 (primary, positive "8" = 0)** + ko/ja/zh mermaid M6/M7 phantom structural drift. D2 재조사 결과 ja가 최악 (이전 "zh heavily drifted" 주장은 오류).
4. **AC 기계적 검증** — iter-2: 모든 AC regex를 live tree(`8108d4311`)에서 사전 검증. digit-boundary 보정(`(^|[^0-9])`)으로 "8 inside 28" substring false-positive 차단. no-op check(archived-name-in-mermaid) 제거 → 실제 구조적 drift(M6/M7 phantom) 검출로 교체 (D3).

---

## §D. Baseline Capture (run-phase M1 시작 전)

run-phase M1에서 아래 baseline grep 결과를 이 섹션에 캡처:

```bash
# (M1에서 채움 — 수정 전 상태)
grep -rln 'glm-5\.1\|GLM-5\.1' docs-site/content/{en,ko,ja,zh}/multi-llm/
grep -rEln '28\s*(specialized\s*)?agents?|28個.*エージェント|28.*代理' docs-site/content/{en,ko,ja,zh}/
bash scripts/docs-i18n-check.sh
```

---

## §E.1 Plan-phase Audit-Ready Signal

**Plan-phase**: COMPLETE (iter-2 — 10 defects D1..D10 resolved per plan-auditor iter-1 FAIL 0.62)
**Artifacts**: 5/5 (spec.md + plan.md + acceptance.md + research.md + progress.md)
**Tier**: L (justified — ≈10-11 edit files + 4-locale simultaneous + 6 facts-axis × page-family + no-IA-redesign + primary-source traceability; D9 dropped "24 pages / 4× exceed" inflation)
**docs-truth re-verification**: 6/6 axes PASS against primary sources (tree `8108d4311`, 2026-06-17)
**Drift inventory**: 11 issues (D-001, D-001b, D-002, D-003, D-004, D-004b, D-005, D-006, D-007, D-008, D-009) + 1 adjacent (ADJ-1 "31 skills", out of scope)
**AC count**: 11 (4 Critical, 3 High, 4 Medium) — iter-2 added AC-011 (language count)
**REQ count**: 8 (iter-2 added REQ-DOCSITE-008 language count)
**Plan-auditor**: pending iter-2 (after D1..D10 fixes)
**Implementation Kickoff Approval (§19.1)**: pending (after plan-auditor PASS)

**Audit-ready YAML**:
```yaml
spec_id: SPEC-V3R6-DOCS-DOCSITE-001
plan_phase: complete
plan_iteration: 2
tier: L
artifacts:
  spec_md: present
  plan_md: present
  acceptance_md: present
  research_md: present
  progress_md: present
docs_truth_reverification:
  agent_catalog: PASS
  spec_status_enum: PASS
  frontmatter_12_fields: PASS
  cli_surface: PASS
  glm_tier_models: PASS
  language_count_16: PASS  # CONST-V3R2-004 FROZEN rule
drift_inventory:
  total_issues: 11
  critical: 2    # D-001 (stale 28), D-004 (glm-5.1 stale)
  high: 3        # D-001b (mermaid structural), D-003 (agent-guide verify), D-004b (glm-5.2 absent)
  medium: 3      # D-005, D-006, D-009 (18-languages)
  low: 2         # D-007 (faq verify-only downgrade), D-008
  ok_verify_only: 1  # D-002
  adjacent_out_of_scope: 1  # ADJ-1 "31 skills"
ac_count: 11
req_count: 8
iter2_defects_resolved: [D1, D2, D3, D4, D5, D6, D7, D8, D9, D10]
plan_auditor: pending_iter2
implementation_kickoff_approval: pending
```

---

## §E.2 Run-phase Evidence

_<pending run-phase — manager-develop가 M1..M6 실행 후 채움>_

---

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

---

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase — manager-docs가 채움>_

---

## §E.5 Mx-phase Audit-Ready Signal

_<pending Mx-phase>_

---

## §F. Next Actions

1. plan-auditor independent audit (bias prevention, GEARS compliance, AC mechanical-verifiability check)
2. plan-auditor PASS (≥0.85 skip-eligible OR ≥0.75 with debt) 후 orchestrator §19.1 Implementation Kickoff Approval을 사용자에게 AskUserQuestion으로 요청
3. 사용자 승인 후 `/moai run SPEC-V3R6-DOCS-DOCSITE-001` → manager-develop M1 진입
4. M1: baseline grep 캡처 (§D) + drift inventory 확정
5. M2: `core-concepts/what-is-moai-adk.md` ja/zh "8 retained" 조정 (가장 가시적 drift)
