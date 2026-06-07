---
spec_id: SPEC-WEB-CONSOLE-007
status: draft
era: V3R6
tier: M
development_mode: tdd
deferred_to: SPEC-WEB-CONSOLE-008   # workflow/git-strategy/harness/llm nested editing (REQ-WC-012 boundary lift + new validators + sentinel retarget)
---

# SPEC-WEB-CONSOLE-007 — Progress

## §F.1 Plan-phase

- **Authored**: 2026-06-06 by manager-spec.
- **Artifacts**: spec.md (12-field frontmatter + GEARS REQ-WC7-001..014 + §F Exclusions), plan.md (Tier M 정당화 + M1..M6 + test-class per milestone), acceptance.md (AC-WC7-001..020 + traceability + MUST-PASS gate), design.md (nested serialization + curated inventory + per-field validation map + 신규 위젯 + write-seam load-modify-write + nested isolation), research.md (전체 file:line ground-truth).
- **Tier**: M (right-size 정당화 plan.md §F — 6 nested 필드는 작으나 2 신규 Templ 위젯 + write-seam 심화 + 검증 export seam + 4종 TDD 거동으로 Tier S 초과; 단일 패키지 + 신규 validator 0개 + 서버 계약 무변경으로 Tier L 미달).
- **Curated 편집 필드 인벤토리** (정확히 이것만, spec.md §E):
  - quality: `test_coverage_target`(int, 기존 0-100), `enforce_quality`(bool), `tdd_settings.min_coverage_per_commit`(int, 기존 0-100).
  - git_convention: `convention`(enum, 기존 유지) + `auto_detection.confidence_threshold`(float [0,1], 기존), `auto_detection.enabled`(bool), `custom.pattern`(string, 기존 custom-required).
- **CRITICAL SCOPE CONSTRAINT 준수**: 두 기존 검증기(validateQualityConfig/validateGitConventionConfig) 확장/export seam만; 신규 validator 함수 0개.
- **HARD invariants**: 8개 (spec.md §B). HARD-2(006 sentinel 무수정) + HARD-4(nested isolation 증명) 핵심.
- **Deferred → 008** (spec.md §F): workflow/git-strategy/harness/llm nested 편집(boundary lift + 신규 validator + sentinel retarget), partial-swap fragment, 동적 섹션 레지스트리.
- **SPEC ID self-check**: decomposition: SPEC ✓ | WEB ✓ | CONSOLE ✓ | 007 ✓ → PASS (canonical regex `^SPEC(-[A-Z][A-Z0-9]*)+-\d{3}$`).
- **Plan-phase commit**: 926816abe

## §F.2 Plan Audit Gate
- **iter-1 (plan-phase)**: PASS-WITH-DEBT 0.83 (Tier M 임계 0.80) + D1-D4 패치(02441d3db: AC grep idioms + citation off-by-one).
- **iter-2 (run-phase Phase 0.5 재실행, cache MISS)**: PASS-WITH-DEBT, aggregate **0.84** (+0.01 monotonic, no regression). MP-1..4 전부 PASS/N-A.
- **D1 (BLOCKING) RESOLVED**: §F "Exclusions" h2 → `moai spec lint --strict` MissingExclusions ERROR (live 검증). Fix: `### §F.1 Out of Scope` h3 sub-section 추가 + numbered→dash 변환 (orchestrator-direct 기계적 패치, L_orchestrator_direct_plan_patch). 검증: 007 MissingExclusions 0 (repo-wide 19 debt 중 007 제외 확인).
- **D2 (SHOULD-FIX) accepted-debt + mitigation**: 10 AC `go test -run 'PATTERN'` false-GREEN 위험(no-match→exit 0). AC-020 full-suite가 최종 backstop. 추가 완화: manager-develop 위임에 RED-verification discipline 주입(test EXISTS + FAIL 확인, bare exit 0 금지) + orchestrator post-impl 검증도 grep-guarded.
- **D3 (SHOULD-FIX) + D4-D6 (MINOR)**: accepted debt — review-1.md 참조. D4(validator citation 96/163) HOLD 확인.
- **Report**: .moai/reports/plan-audit/SPEC-WEB-CONSOLE-007-review-1.md
- **Verdict**: PASS-WITH-DEBT 0.84 → GATE-2 진입 (0.84 < 0.90 → 비-skip이나 PASS 임계 0.80 초과).

## §E.1 Run-phase

GATE-2 승인 후 manager-develop cycle_type=tdd (Mode 5 sub-agent sequential M1-M6). 베이스 HEAD 076fb44b6 (plan-audit D1 패치 commit). spec.md status draft→in-progress 전환은 M1 커밋에서 수행(소유권 전환).

### Run-phase milestones (manager-develop appends below)

- **M1 — 신규 Templ 위젯 (toggle + numberField) + Class A markup-parity (RED→GREEN)**
  - 신규 위젯 2종 page.templ에 추가: `toggle`(bool checkbox + hidden companion `__present` for EC-1 bool semantics) + `numberField`(int/float `<input type="number">`, min/max/step client-hint). 각각 optSelect field chrome + segmentCheckbox/fieldsetIdentity input 패턴 재사용.
  - Class A markup-parity 테스트 2종 templ_helpers_test.go에 추가: `TestToggleHelperMarkupParity`(clean/checked/errored 3-render, hidden companion 단언, checked 음성 단언) + `TestNumberFieldHelperMarkupParity`(clean/errored, min/max/step/aria-invalid 단언). RED→GREEN 확인(`--- PASS` 명시 실행, B-D2 완화).
  - `templ generate` (internal/web CWD에서 실행 — 커밋된 bare-filename FileName 형식 일치) → page_templ.go 재생성. fieldsets_templ.go/root_templ.go 무변경.
  - AC-WC7-001 PASS, AC-WC7-002 PASS, AC-WC7-003 drift-free(idempotent regen). 전체 web suite GREEN, host build exit 0.
  - commit 8e555211f.

- **M2 — config 검증 export seam (신규 규칙 0개)**
  - `ValidateQualitySection(*models.QualityConfig)` + `ValidateGitConventionSection(*models.GitConventionConfig)` thin exported wrapper를 validation.go에 추가 — 기존 unexported `validateQualityConfig`/`validateGitConventionConfig`로 verbatim forward(`return validateQualityConfig(q)`). IsValidConvention/ValidConventions가 convention SSOT를 export하는 패턴과 동형. **신규 validator 함수 0개**(AC-WC7-004 invariant 유지: `grep -cE 'func validate(Workflow|GitStrategy|Harness|Llm)Config'` == 0).
  - 단위 테스트 2종 validation_test.go에 추가: `TestValidateQualitySection`(in-range pass / test_coverage_target=150 → 기존 "must be between 0 and 100" 메시지 재사용 / min_coverage_per_commit=-5 동일) + `TestValidateGitConventionSection`(custom+빈 pattern → 기존 "pattern is required..." / confidence_threshold=1.5 → 기존 "must be between 0.0 and 1.0"). RED→GREEN 명시 실행 확인(B-D2).
  - AC-WC7-004 PASS(0), AC-WC7-005 PASS, AC-WC7-006 PASS. 전체 config suite GREEN.
  - commit 924438a94.

- **M3 — 폼 파싱 + view-model 확장 + fieldsetProject 렌더 (RED→GREEN)**
  - `projectNestedForm` 구조체 + `parseProjectNestedForm(r)` 추가(projectconfig.go): dot-path PostFormValue 6필드 명시 파싱(reflection path-walker 금지), `*Set` 플래그(EC-1 empty=미제출), bool hidden companion(`__present`)으로 "false 변경" vs "보존" 구분, `ParseErrs` 타입변환 가드(EC-4류).
  - `readProjectNestedConfig`(신규 read seam, GET echo-back, 기존 readProjectConfig 무변경) + `projectNestedCurrent` 추가. pageView에 Cur* nested 필드 6개 추가. app.go에 readProjectNestedConfig/writeProjectNestedConfig 주입 seam 추가 + newApp 배선.
  - handleSave 배선: parseProjectNestedForm + validateProjectNestedConfig(3번째 validator, merge) + writeProjectNestedConfig(scalar write 후). applyNestedCurrent/applyNestedForm(rejected echo-back) + successProjectView/rejectedProjectView 헬퍼.
  - fieldsetProject 확장(fieldsets.templ): 6 nested 위젯(numberField×3, toggle×2 with companion, projectTextField×1) 추가, `count.project` 2→8. root.templ 합성 라인 무변경(fieldsetProject 내부만).
  - **i18n 12키(6필드×title/desc) 4-locale(en/ko/ja/zh) + count.project 8 갱신** — TestDataI18nKeysSubsetOfDictionary gate가 per-milestone GREEN 요구하여 M6의 i18n 작업을 M3로 당겨 흡수.
  - 테스트: `TestParseProjectNestedForm`(5 sub: all-set/empty EC-1/companion-false/int-guard/float-guard) + `TestProjectFieldsetRendersNestedWidgets`(6위젯 렌더 + checked/unchecked 단언). RED→GREEN 명시 실행. 전체 web suite GREEN(i18n parity 포함).
