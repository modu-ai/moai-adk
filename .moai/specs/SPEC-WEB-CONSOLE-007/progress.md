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

### Phase 0.95 Mode Selection (orchestrator autonomous)
- **Decision: sub-agent** (Mode 5, sequential single-spawn manager-develop cycle_type=tdd, M1-M6).
- **Input**: tier=M, scope≈12 files (internal/web + thin internal/config export seam), domain=1, concurrency-benefit=LOW (coding-heavy, Finding A4), Agent-Teams-prereqs=not-met (harness standard).
- **Evaluation**: Mode 1 trivial=no · Mode 2 background=no(writes) · Mode 3 agent-team=no(prereq unmet + single-domain) · Mode 4 parallel=no(coding-heavy single-domain, Finding A4) · Mode 6 workflow=no(<30 files, semantic TDD non-mechanical).
- **Justification**: coding-heavy single-package TDD → Finding A4 sequential sub-agent default. No Round split (Tier M <30 tasks). GATE-2 user-approved this session.

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
  - commit 01c9a1c5a.

- **M4 — write-seam round-trip + nested isolation 증명 + EC-1/EC-2 + server reject (RED→GREEN)**
  - `writeProjectNestedConfig`(projectconfig.go, M3에서 함께 작성됨): load-modify-write — LoadRaw 전체 섹션 구조체 복사(`q := cfg.Quality` / `gc := cfg.GitConvention`) 후 `*Set` 게이트로 타깃 nested 필드만 변경 → SetSection → Save. nested-of-nested(TDDSettings/AutoDetection) sibling 라이드. scalar write 후 실행(2 LoadRaw cycle, 둘 다 동일 on-disk 섹션에 수렴).
  - 신규 테스트 파일 projectnested_test.go(10 함수, 실 seam + 실 프로젝트 root, seedNestedProject가 편집대상+비-편집 sibling nested 필드 시드):
    - `TestProjectNestedRoundTrip`(AC-007), `TestProjectNestedSiblingPreserved`(AC-008 — coverage_exemptions.max_exempt_percentage=42 / tdd_settings.test_first_required / lsp_quality_gates.enabled / min_coverage_per_commit 4 sibling 보존), `TestProjectNestedGitConventionSiblingPreserved`(AC-009 — formatting.verbose / validation.max_length / auto_detection.enabled 보존).
    - `TestProjectNestedEmptyPreserves`(AC-010 EC-1), `TestProjectNestedToggleEC1`(AC-011a — companion 없음=보존), `TestProjectNestedToggleUnchecked`(AC-011b — companion+미체크=false).
    - `TestProjectNestedAtomicReject`(AC-012 — valid+invalid 동시 → 400, 둘 다 무write), `TestProjectNestedOutOfRangeReject`(AC-013 — 150 → 400 + 기존 메시지 + write 0), `TestProjectNestedCustomPatternRequired`(AC-014 — convention=custom+빈 pattern → 400 + 기존 custom-required + write 0), `TestSaveNestedFullPage`(AC-019 — hx-boost full-page swap, partial fragment 미도입).
  - **Class A 마크업 디테일 2건 test-fix**(코드 아님): Templ `<!doctype html>` 소문자 정규화(coverage_test.go convention `strings.ToLower`) + 메시지 내 apostrophe HTML-escape(`'`→`&#39;`)로 인한 exact-string 단언 조정. 10/10 GREEN.
  - AC-WC7-007/008/009/010/011a/011b/012/013/014/019 PASS. full web+config+models suite GREEN.
  - commit 08561f8c5.

- **M5 — 통합 + 회귀 가드 + 커버리지 강화 (boundary verification + REFACTOR 검토)**
  - **HARD-2 sentinel 무수정 확인**: `git diff 076fb44b6 -- internal/web/integration_test.go` 빈 결과(byte-identical). 006 scope-boundary 테스트(TestScopeBoundary/TestPersistence/TestGolden) 무수정 GREEN. AC-WC7-016/017 PASS.
  - **AC-WC7-015**: web 레이어 직접 YAML write 0 (`yaml.Marshal|os.WriteFile` in projectconfig.go/handlers.go == 0).
  - **AC-WC7-018**: offline zero-network — CDN ref 0 (templ + assets css/js).
  - **AC-WC7-019 (2nd)**: hx-target/hx-swap 0 (partial-swap fragment 미도입, hx-boost full-page 유지).
  - **AC-WC7-004**: new-validator 0 (불변).
  - **AC-WC7-003**: templ generate idempotent drift-free CLEAN.
  - **AC-WC7-020 커버리지 ground-truth 정정**: spec.md "90.9% baseline"은 **재현 불가 stale 값**. 격리 worktree로 base 076fb44b6 실측 → **71.6%**(reproducible). 내 HEAD = **72.5%**(+0.9%). AC-020 의도("하락 없음") 충족. (L_glyph_count_provisional 류 — 문서화 baseline이 provisional, ground-truth는 71.6%.)
  - 커버리지 강화 테스트 projectnested_error_test.go(5 함수): git_convention nested round-trip(write seam git_convention 브랜치) + rejected echo-back(applyNestedForm 전 *Set 브랜치) + nested write-failure 핸들러 에러 경로 + nested read-failure inline 에러 + readProjectNestedConfig EC-5 defaults.
  - **REFACTOR 검토**: 두 write seam(writeProjectConfig/writeProjectNestedConfig)의 load-modify-write 공유 헬퍼 추출은 plan.md §I anti-over-engineering 경고 대상(서로 다른 필드셋, 마진 이익) → scope discipline으로 분리 유지.
  - golangci-lint(web+config) 0 issues. cross-platform build host+windows exit 0.
  - commit bfbafe4d2.

- **M6 — i18n 4-locale parity + Project fieldset 카운트 + offline 재확인 + 최종 검증 배치**
  - **i18n 작업은 M3에서 흡수됨**(TestDataI18nKeysSubsetOfDictionary gate per-milestone GREEN 요구). M6는 parity 확인 + 최종 검증 전담.
  - i18n 4-locale parity 확인: 신규 12키 전부 4 occurrences(en/ko/ja/zh 1회씩). count.project "8 fields"/"8개 항목"/"8 項目"/"8 项" 4-locale 갱신.
  - offline 0-network 재확인: htmx.min.js + self-host font(go:embed) 무변경, 신규 CDN 0.
  - 최종 검증 배치: AC-WC7-001..020 전부 PASS(MUST-PASS 13개 + SHOULD-PASS 7개), full go test ./... (SPEC-007 scope packages web/config/models 전부 GREEN), golangci-lint full 0 issues, host+windows build exit 0, templ idempotent drift-free.
  - **무관 pre-existing flaky**: internal/hook TestHookWrapper_{MoaiBinaryFallback,ValidJSON}가 full-suite 5s-timeout 병렬 부하에서 간헐 FAIL → 격리 재실행 PASS(0.39s/0.54s). internal/hook은 SPEC-007 무수정(git diff 076fb44b6..HEAD --stat 무관). prior-cohort CI-FLAKY-STABILIZE 패턴.

## §E.2 Run-phase Evidence (AC PASS/FAIL matrix)

| AC | MUST-PASS | Status | Verification | Actual |
|----|-----------|--------|--------------|--------|
| AC-WC7-001 | | PASS | TestToggleHelperMarkupParity | --- PASS |
| AC-WC7-002 | | PASS | TestNumberFieldHelperMarkupParity | --- PASS |
| AC-WC7-003 | ✓ | PASS | templ generate && git diff --exit-code | CLEAN (idempotent) |
| AC-WC7-004 | ✓ | PASS | grep -cE 'func validate(Workflow\|GitStrategy\|Harness\|Llm)Config' | 0 |
| AC-WC7-005 | ✓ | PASS | TestValidateQualitySection | --- PASS (기존 메시지 재사용) |
| AC-WC7-006 | ✓ | PASS | TestValidateGitConventionSection | --- PASS (기존 메시지 재사용) |
| AC-WC7-007 | ✓ | PASS | TestProjectNestedRoundTrip | --- PASS (target=85 영속) |
| AC-WC7-008 | ✓ | PASS | TestProjectNestedSiblingPreserved | --- PASS (4 sibling 보존) |
| AC-WC7-009 | | PASS | TestProjectNestedGitConventionSiblingPreserved | --- PASS (3 sibling 보존) |
| AC-WC7-010 | | PASS | TestProjectNestedEmptyPreserves | --- PASS (empty=70 보존) |
| AC-WC7-011a | | PASS | TestProjectNestedToggleEC1 | --- PASS (companion 없음=보존) |
| AC-WC7-011b | | PASS | TestProjectNestedToggleUnchecked | --- PASS (companion+미체크=false) |
| AC-WC7-012 | ✓ | PASS | TestProjectNestedAtomicReject | --- PASS (400, 둘 다 무write) |
| AC-WC7-013 | ✓ | PASS | TestProjectNestedOutOfRangeReject | --- PASS (150→400 + write 0) |
| AC-WC7-014 | ✓ | PASS | TestProjectNestedCustomPatternRequired | --- PASS (custom 빈 pattern→400) |
| AC-WC7-015 | ✓ | PASS | grep yaml.Marshal\|os.WriteFile in web | 0 |
| AC-WC7-016 | ✓ | PASS | TestSaveScopeBoundary (006 sentinel 무수정) | --- PASS |
| AC-WC7-017 | ✓ | PASS | git diff 076fb44b6 -- integration_test.go (workflow/harness/git-strategy 추가 라인) | 0 |
| AC-WC7-018 | | PASS | grep CDN refs in templ+assets | 0 |
| AC-WC7-019 | | PASS | TestSaveNestedFullPage + grep hx-target/hx-swap | --- PASS + 0 |
| AC-WC7-020 | ✓ | PASS | full suite GREEN + coverage | web/config/models ok; cov 72.5% (base 71.6% +0.9%) |

**MUST-PASS 13/13 GREEN. SHOULD-PASS 7/7 GREEN. 총 20/20 PASS.**

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-06-07
run_commit_sha: 5b05a74cf   # M1 8e555211f / M2 924438a94 / M3 01c9a1c5a / M4 08561f8c5 / M5 bfbafe4d2 / M6 5b05a74cf (final)
run_status: implemented
ac_pass_count: 20
ac_fail_count: 0
preserve_list_post_run_count: 0   # PRESERVE 외 변경 0 (006 sentinel byte-identical, integration_test.go 무수정)
l44_pre_commit_fetch: n/a          # 격리 worktree(wt-spec007-run), 단일 세션, orchestrator가 push 전 fetch 수행
l44_post_push_fetch: n/a           # DO NOT PUSH (orchestrator 검증 게이트 후 push)
new_warnings_or_lints_introduced: 0   # golangci-lint full 0 issues
cross_platform_build:
  host: pass        # go build ./... exit 0
  windows: pass     # GOOS=windows GOARCH=amd64 go build ./... exit 0
total_run_phase_files: 12   # page.templ page_templ.go templ_helpers_test.go validation.go validation_test.go projectconfig.go handlers.go validate.go app.go fieldsets.templ fieldsets_templ.go i18n.js + 2 신규 테스트(projectnested_parse_test.go projectnested_test.go projectnested_error_test.go) + spec.md/progress.md
m1_to_mN_commit_strategy: "M1-M6 milestone별 분리 commit, Authored-By-Agent: manager-develop trailer, draft→in-progress (M1) 소유권 전환, NOT pushed (orchestrator push gate)"
new_validator_functions: 0   # AC-WC7-004 CRITICAL SCOPE CONSTRAINT
006_sentinel_byte_identical: true   # integration_test.go:197-205 무수정
coverage_baseline_correction: "spec.md 90.9%는 재현불가 stale; ground-truth base 076fb44b6 = 71.6%(격리 worktree 실측), HEAD = 72.5%(+0.9%)"
```

## §F.3 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: 2026-06-07
sync_method: orchestrator-direct   # bounded Tier M internal console feature; web-console cohort 004/005/006 pattern (L_orchestrator_direct_sync_tier_m); Authored-By-Agent trailer omitted → OwnershipTransitionRule silent SKIP
sync_status: implemented
deliverables:
  - CHANGELOG.md   # [Unreleased] ### Added — SPEC-WEB-CONSOLE-007 entry (B12 dedup verified 0)
  - spec.md        # frontmatter status in-progress→implemented + version 0.1.0→0.2.0
  - progress.md    # this §F.3
docs_site: not-required   # internal moai-web console feature, no dedicated docs-site page (006 precedent)
readme: not-required
orchestrator_independent_verification:   # Trust-but-verify batch in main checkout, post FF-merge
  full_suite: pass          # internal/web + internal/config + pkg/models all ok
  head_coverage: 72.5%      # = base 71.6% +0.9% (independently re-measured at 076fb44b6)
  ac_pass: 20/20            # MUST-PASS 13/13 spot-checked --- PASS via D2-guard (=== RUN confirmed)
  new_validator: 0
  006_sentinel_diff: 0      # git diff 076fb44b6 HEAD -- integration_test.go empty (byte-identical)
  subagent_boundary: 0
  yaml_marshal_web: 0
  cdn_refs: 0
  templ_drift: clean        # templ generate (from internal/web) + git diff --exit-code clean
  cross_platform_build: "host+windows exit 0"
  golangci_lint: "0 issues"
integration: "L1 worktree (agent-ae1fb1c0fe675fe86, branch wt-spec007-run, runtime-autonomous) → FF-merge into feat/SPEC-WEB-CONSOLE-007 (base 076fb44b6 == wt base → SHA-preserving FF). L_l1_worktree_cherrypick/FF pattern."
push: HEAD:main FF (Hybrid Trunk Tier M main-direct, user-approved post-sync)
next: SPEC-WEB-CONSOLE-008 (workflow/git-strategy/harness/llm nested editing — boundary lift + new validators + sentinel retarget)
```
