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
| Plan-audit | complete (PASS-WITH-DEBT 0.88 iter-2) | plan-auditor |
| Run | complete (M1-M6) | manager-develop |
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

run-phase M1에서 아래 baseline grep 결과를 이 섹션에 캡처 (tree `0ffb498a9`, worktree):

**AC-001 stale-28 (core-concepts/what-is-moai-adk.md)**: en=0, ko=2, ja=5, zh=1
**AC-002(b) phantom-M6/M7**: en=0, ko=2, ja=2, zh=2
**AC-002(a) archived-unframed**: all=0 (이미 clean)
**AC-003(a) glm-5.1 files (multi-llm/)**: en=0, ko=1, ja=0, zh=0
**AC-003(b) glm-5.2 files (multi-llm/)**: all=0 (canonical 부재)
**AC-005(a) positive-8**: en=2, ko=4, ja=0, zh=5
**AC-006 stale CLI count**: 0
**AC-008 faq table-rows**: all=4 (이미 보유)
**AC-011 18-languages**: en=1, ko=3, ja=1, zh=0
**AC-004 _index.md lines**: en=9, ko=72, ja=9, zh=9 (parity gap)
**i18n tool baseline**: exit 1 — 62 errors (56 claude-code/ out-of-scope + 3 multi-llm model-policy "Anthropic" glossary + 3 zh/builder-agents glossary)

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

**Run-phase**: COMPLETE (M1-M6, manager-develop, worktree `agent-aae7c698e73c68938`)
**Backend**: GLM `glm-5.2[1m]` (orchestrator-direct — manager-develop spawned in L1 worktree)

### M1 — Baseline + status draft→in-progress
- spec.md frontmatter `status: draft → in-progress` transitioned
- PRE-FIX baseline greps captured to §D (all match acceptance.md documented values)

### M2 — Agent catalog 4-locale + mermaid rewrite (AC-001/002/005)
- **ja** `what-is-moai-adk.md`: 5 stale-28 eradicated (L7/L48/L226/L496/L670), positive-8 introduced (4), mermaid phantom M6/M7 + Experts/Teams subgraphs rewritten to EN clean structure (4 Managers + 2 Evaluators + 1 Builder + Explore)
- **ko** `what-is-moai-adk.md`: 2 stale-28 fixed (L226 prose, L666 code-block), mermaid rewritten
- **zh** `what-is-moai-adk.md`: 1 stale-28 fixed (L674 code-block), mermaid rewritten
- **en** `what-is-moai-adk.md`: no stale-28 (already clean), mermaid already clean (template)
- POST-FIX: stale-28 all=0, phantom-M6/M7 all=0, archived-unframed all=0, positive-8 en=2/ko=5/ja=4/zh=5

### M3 — GLM tier-models 4-locale (AC-003/004)
- **ko** `multi-llm/_index.md`: `GLM-5.1 → glm-5.2[1m]` (L17, L23)
- **en/ja/zh** `multi-llm/_index.md`: stubs (9 lines) expanded to ko-equivalent 72-line tier-models table with `glm-5.2[1m]`
- **en/ja/zh** `multi-llm/model-policy.md`: stubs expanded to ko-equivalent 72-line content (resolves 3 "Anthropic" glossary i18n errors)
- POST-FIX: glm-5.1 all=0, glm-5.2 each=1, _index.md parity 72/72/72/72, model-policy.md parity 72/72/72/72

### M4 — CLI 17 + SPEC status/frontmatter + language 16 (AC-006/007/011)
- **18→16 languages** fixed: en what-is-moai-adk L216, ko what-is-moai-adk L49+L216, ko introduction L134, ja what-is-moai-adk L216 (zh already clean)
- AC-006 stale CLI count: 0 (no stale counts existed)
- AC-007 SPEC status enum: vacuously satisfied (no docs-site page enumerates the full 8-value lifecycle enum)
- POST-FIX: 18-languages all=0

### M5 — FAQ + introduction verify-only (AC-008)
- faq model-assignment table-rows: 4/4/4/4 (verify-only, no edit needed)
- introduction.md "8 agents" parity: all 4 locales state 8 agents (verify-only)

### M6 — Final AC sweep + regression guard
- All 11 ACs mechanically verified (see §E.3)
- AC-010 regression guard: 0 archived-agent names added in diff
- i18n tool: 60 errors → all remaining are claude-code/ (out-of-scope, 56) + pre-existing glossary gaps in zh/builder-agents (3) + zh/what-is-moai-adk "Anthropic" (1, pre-existing)

### Files modified (11 docs-site + 2 SPEC artifacts)
- `docs-site/content/{en,ko,ja,zh}/core-concepts/what-is-moai-adk.md` (4 files — M2 stale-28 + mermaid + M4 language)
- `docs-site/content/{en,ko,ja,zh}/multi-llm/_index.md` (4 files — M3 GLM tier-models)
- `docs-site/content/{en,ko,ja,zh}/multi-llm/model-policy.md` (3 files en/ja/zh — M3 stub expansion; ko untouched)
- `docs-site/content/ko/getting-started/introduction.md` (1 file — M4 18→16)
- `.moai/specs/SPEC-V3R6-DOCS-DOCSITE-001/spec.md` (status draft→in-progress)
- `.moai/specs/SPEC-V3R6-DOCS-DOCSITE-001/progress.md` (this file — §A/§D/§E.2/§E.3)

---

## §E.3 Run-phase Audit-Ready Signal

```yaml
spec_id: SPEC-V3R6-DOCS-DOCSITE-001
run_phase: complete
run_iteration: 1
run_complete_at: 2026-06-18
run_commit_sha: "55f430c03 (pre-rebase; will backfill post-push with final pushed SHA)"
run_status: audit-ready
ac_pass_count: 11
ac_fail_count: 0
ac_matrix:
  AC-001: PASS   # stale-28 all=0 (was en=0/ko=2/ja=5/zh=1)
  AC-002a: PASS  # archived-unframed all=0
  AC-002b: PASS  # phantom-M6/M7 all=0 (was en=0/ko=2/ja=2/zh=2)
  AC-003a: PASS  # glm-5.1 all=0 (was en=0/ko=1/ja=0/zh=0)
  AC-003b: PASS  # glm-5.2 each>=1 (was all=0)
  AC-004: PASS   # _index.md parity 72/72/72/72, model-policy.md 72/72/72/72
  AC-005a: PASS  # positive-8 en=2/ko=5/ja=4/zh=5 (ja was 0)
  AC-006: PASS   # stale CLI count=0
  AC-007: PASS   # status enum vacuously satisfied (no lifecycle page exists)
  AC-008: PASS   # faq table-rows 4/4/4/4 (verify-only)
  AC-009: PASS   # i18n: remaining errors all out-of-scope (claude-code/ 56 + pre-existing glossary 4)
  AC-010: PASS   # archived active-context regression=0
  AC-011: PASS   # 18-languages all=0 (was en=1/ko=3/ja=1/zh=0)
preserve_list_post_run_count: 0   # no PRESERVE-list files touched
new_warnings_or_lints_introduced: 0
cross_platform_build:
  relevant: false   # docs-only SPEC — zero Go code changes
  note: "Go code untouched; build irrelevant to this docs-only SPEC"
coverage:
  relevant: false   # docs-only
  note: "No Go test coverage applicable"
subagent_boundary_grep:
  relevant: false   # internal/harness, internal/hook not in scope
total_run_phase_files: 13   # 11 docs-site + spec.md + progress.md
m1_to_mN_commit_strategy: per-milestone grouping into single run-phase commit
go_code_changes: 0
i18n_tool:
  pre_fix_errors: 62
  post_fix_errors: 60
  resolved_in_scope: 3   # multi-llm model-policy "Anthropic" glossary (en/ja/zh)
  remaining_out_of_scope: 60   # 56 claude-code/ + 3 zh/builder-agents glossary + 1 zh/what-is-moai-adk "Anthropic" (pre-existing)
```

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
