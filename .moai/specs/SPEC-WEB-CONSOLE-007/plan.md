# SPEC-WEB-CONSOLE-007 — Implementation Plan

## §A. Context

web-console-v4 코호트 S2b. SPEC-WEB-CONSOLE-006(HTMX+Templ 마이그레이션, completed `5714bae97`)의 Templ 섹션 트리 + HTMX foundation을 enabler로, `moai web` 콘솔의 편집 깊이를 quality + git_convention 두 섹션의 curated nested 필드(6개 + 기존 스칼라 2개 유지)로 확장한다. development_mode: tdd → run-phase cycle_type=tdd (RED→GREEN→REFACTOR).

## §B. Known Issues / Constraints from Ground-Truth
- `validateQualityConfig`/`validateGitConventionConfig`는 현재 unexported(internal/config) → 재사용 위해 thin export seam 필요(신규 규칙 아님).
- bool checkbox EC-1: 미체크=폼 미전송 → "보존"과 "false 변경" 구분 위해 hidden companion 패턴(design.md §D.2).
- SetSection은 섹션 전체 교체 → load-modify-write 필수(design.md §D.1).
- 006 sentinel(integration_test.go:197-205)은 quality/git-convention을 **포함하지 않음** → nested 확장이 자연히 무수정 GREEN.

## §C. Pre-flight (run-phase 진입 전 검증)
```bash
git rev-parse --abbrev-ref HEAD          # feat/SPEC-WEB-CONSOLE-007
go test ./internal/web/... ./internal/config/... ./pkg/models/...   # baseline GREEN
which templ || go install github.com/a-h/templ/cmd/templ@latest      # templ codegen 가용
```

## §D. Constraints (HARD)
spec.md §B HARD-1..8 전부. 특히:
- CRITICAL SCOPE CONSTRAINT: 신규 validator 함수 0개. 두 기존 검증기 확장/export seam만.
- HARD-2: integration_test.go:197-205 무수정. 설계가 변경 강제 시 → STOP + blocker(008 침범).
- HARD-4: nested isolation 증명 테스트 필수.

## §E. Self-Verification (각 milestone GREEN 기준)
- `go test ./internal/web/... -count=1` exit 0
- `templ generate` drift-free (`git diff --exit-code internal/web/*_templ.go` 재생성 후 clean)
- `grep -c "func validateWorkflowConfig\|func validateGitStrategyConfig" internal/config/validation.go` == 0 (신규 validator 0개 불변)
- 006 sentinel 테스트(`TestScopeBoundary` 류 in integration_test.go) PASS 무수정

## §F. Tier 결정 + 정당화

**선택: Tier M (standard).**

**정당화 (right-size 판단)**:
- 편집 필드 표면은 6개(+스칼라 2 유지)로 작지만, **단순 Tier S로 축소 불가**한 다음 이유로 Tier M:
  1. **2개 신규 재사용 Templ 위젯**(toggle + numberField) 생성 — 각각 Class A markup-parity 테스트 + `templ generate` codegen 동반. (위젯 생성은 단일 파일 편집이 아님.)
  2. **write-seam 심화** — load-modify-write nested isolation은 단순 스칼라 추가가 아니라 nested-of-nested(TDDSettings/AutoDetection) 보존 메커니즘 + bool EC-1 companion 패턴.
  3. **검증 export seam** — internal/config의 2개 검증기 export + web validateProjectConfig 확장(다중 파일: validation.go + validate.go + handlers.go + projectconfig.go + page.templ + fieldsets.templ).
  4. **TDD 사이클 4종 신규 거동**(round-trip / nested 보존 / EC-1·EC-2 / server reject) — 각각 RED→GREEN.
- 반대로 **Tier L은 과대**: 단일 패키지(internal/web + 얇은 config export), 신규 섹션/validator 0개, 서버 계약 무변경(hx-boost full-page 유지), AC 수 ~20 이내. Round 분할(SSE stall 임계 30+ task) 불요.

## §G. Milestones (priority-ordered, no time estimates)

| Mn | Owner | Objective | Test-class | Key files |
|----|-------|-----------|------------|-----------|
| **M1** | manager-develop (tdd) | 신규 Templ 위젯 `toggle` + `numberField` + Class A markup-parity 테스트 (RED→GREEN). `templ generate` codegen. | Class A (신규 parity) | page.templ, page_templ.go, templ_helpers_test.go |
| **M2** | manager-develop (tdd) | config 검증 export seam: `validateQualityConfig`/`validateGitConventionConfig`를 재사용 가능한 export wrapper로 노출(신규 규칙 0개). 단위 테스트로 기존 0-100/[0,1]/custom-required 규칙 재사용 확인. | Class B (HTTP-무관 단위) | internal/config/validation.go, validation_test.go |
| **M3** | manager-develop (tdd) | 폼 파싱 + view-model 확장: `projectNestedForm` 파싱(dot-path PostFormValue, *Set 플래그, bool companion), pageView 신규 현재값 필드, readProjectConfig 확장(GET 현재값). RED→GREEN. | Class B + Class A(fieldsetProject 확장 parity) | handlers.go, validate.go, projectconfig.go, fieldsets.templ |
| **M4** | manager-develop (tdd) | write-seam load-modify-write 확장 + **nested isolation 증명 테스트**(sibling nested 필드 보존, HARD-4). EC-1(empty/companion 보존) + EC-2(atomic reject) 신규 필드 확장. server reject(out-of-range/custom-required) 테스트. RED→GREEN. | Class B (round-trip, isolation, EC) | projectconfig.go, handlers.go, projectconfig_test.go, projectconfig_handler_test.go |
| **M5** | manager-develop (tdd) | 통합 + 회귀 가드: 006 scope-boundary sentinel(integration_test.go:197-205) 무수정 GREEN 확인, `grep` 신규 validator 0개 불변, full `go test ./internal/web/...` + `templ generate` drift-free. REFACTOR(중복 위젯 chrome 정리). | Class B (integration, 무수정 sentinel) | integration_test.go(READ-only 단언), 전체 |
| **M6** | manager-develop (tdd) | i18n 키 + Project fieldset 카운트 갱신("2 fields" → 신규), offline 0-network 재확인(htmx/font self-host 무변경), 최종 검증 배치. | Class A (markup) + Class B | i18n.js(static), fieldsets.templ, assets.go(무변경 확인) |

> Round 분할 없음(Tier M, 30+ task 미만 — SSE stall 임계 미달, `sprint-round-naming.md` 기준 Milestone만 사용).

## §H. Cross-References
- spec.md §B(HARD) / §E(인벤토리) / §F(Exclusions)
- design.md §B(validation map) / §C(위젯) / §D(write-seam isolation)
- research.md(전체 file:line)
- acceptance.md(AC-WC7-NNN)

## §I. Anti-Patterns (회피 목록)
- 동적 reflection path-walker(over-engineering) → 명시적 PostFormValue 6개.
- 신규 validator 함수(CRITICAL SCOPE 위반) → export seam만.
- 섹션 전체를 폼으로 교체(HARD-4 위반) → load-modify-write.
- partial-swap fragment 도입(008 스코프) → hx-boost full-page 유지.
- integration_test.go:197-205 수정(008 침범) → 무수정, 변경 강제 시 blocker.
