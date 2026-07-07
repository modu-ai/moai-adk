---
id: SPEC-BRAND-DIR-REMOVE-001
title: "Clean Removal of .moai/project/brand Directory — Acceptance Criteria"
version: "0.1.0"
status: draft
created: 2026-07-08
updated: 2026-07-08
author: GOOS
priority: P2
phase: "v3.0.0"
module: "internal/harness/safety,internal/hook,internal/cli"
lifecycle: spec-anchored
tier: M
tags: "cleanup, brand, design-retirement, removal, deprecation"
---

# SPEC-BRAND-DIR-REMOVE-001 — Acceptance Criteria (acceptance.md)

## §D AC Matrix

| AC ID | REQ | 판정 방법 | 통과 기준 |
|-------|-----|----------|-----------|
| AC-BDR-001a | REQ-BDR-001,003,004,005 | `grep -rn "project/brand" internal/harness/ internal/hook/ internal/cli/ \| grep -v "_test.go"` | **0 매치** — in-scope 라이브 코드 3사이트(frozen_guard.go / pre_tool.go 라우팅 / migrate_agency.go) 전부 제거 확인 |
| AC-BDR-001b | REQ-BDR-006 | `grep -n "project/brand" internal/defs/dirs.go` | DeprecatedPaths `.moai/project/brand` 항목이 `RemovalSchedule: "v3.0.0"`와 함께 **존재** (보존 드라이버) |
| AC-BDR-001c | REQ-BDR-006 + §D OUT | `grep -rn "project/brand" internal/ cmd/ pkg/ \| grep -v "_test.go"` | 정확히 **11건** 잔존 (M1~M4 완료 기준): dirs.go:292(보존) + 선언된 OUT 10건 = defaults.go:679, config/testdata/design-valid/design.yaml:11, 템플릿 CLAUDE.md:197, 템플릿 design.yaml:38, 템플릿 constitution.md ×5, zone-registry.md:625. **예상 밖 매치 0** |
| AC-BDR-002 | REQ-BDR-009 | `go test ./...` | 전체 통과 (0 fail) |
| AC-BDR-003 | REQ-BDR-003,004 | frozen-guard/sentinel 목록 grep | package-safety `frozenPrefixes`(`internal/harness/safety/frozen_guard.go`)와 `frozenZonePrefixes`(pre_tool.go) 어디에도 `.moai/project/brand/` 라우팅 없음 (상수 `SentinelHarnessFrozenConfig`는 보존) |
| AC-BDR-004 | REQ-BDR-002,007 | temp 디렉터리 `moai init` | `moai init`이 `.moai/project/brand/`를 스캐폴드하지 **않음** |
| AC-BDR-005 | REQ-BDR-005 | migrate-agency 테스트 | `moai migrate agency`가 `.moai/project/brand/`에 쓰지 않고, `.agency/context/` 파일이 `.agency.archived/`에 보존 |
| AC-BDR-006 | REQ-BDR-007 | 템플릿 FS 확인 | `internal/template/templates/.moai/project/brand/` 미존재 + `make build` 성공 |
| AC-BDR-007 | REQ-BDR-008,010 | skill 문서 + 중립성 grep | guard skill 문서에 스테일 brand 예시 없음 **AND** 편집 파일에 SPEC-ID/REQ 토큰 리크 없음 |
| AC-BDR-008 | REQ-BDR-005b (D7) | `grep -n "moai design" internal/cli/migrate_agency.go` | **0 매치** — 스테일 `/moai design` 참조 2건(L297 `result.summary` + L301 `r.logf`) 모두 제거 |
| AC-BDR-009 | REQ-BDR-011 (D4) | `grep -c "SentinelHarnessFrozenConfig" internal/hook/pre_tool.go internal/harness/sentinel_catalog_test.go` | 상수 정의(pre_tool.go:63) + catalog test(sentinel_catalog_test.go:74) **잔존**; `frozenZonePrefixes` 라우팅 사용처(line 759)만 제거 |

## §D.1 Given-When-Then 시나리오

### 시나리오 1 — 라이브 배선 제거 후 grep (핵심, 2단계)
- **Given** 모든 in-scope 편집(M1~M4)이 완료된 트리
- **When (1단계, in-scope)** `grep -rn "project/brand" internal/harness/ internal/hook/ internal/cli/ | grep -v "_test.go"`를 실행하면
- **Then (1단계)** 결과는 **0 매치**여야 한다 — in-scope 라이브 코드 3사이트(`safety/frozen_guard.go` frozen-prefix, `pre_tool.go` frozenZonePrefixes 라우팅, `migrate_agency.go`)가 전부 제거됨.
- **When (2단계, full-tree)** `grep -rn "project/brand" internal/ cmd/ pkg/ | grep -v "_test.go"`를 실행하면
- **Then (2단계)** 결과는 정확히 **11건** — 모두 §C/§D에서 선언한 survivor다: `dirs.go:292`(보존 드라이버, REQ-BDR-006) + declared-OUT 10건(`config/defaults.go`, `config/testdata/design-valid/design.yaml`, 템플릿 `CLAUDE.md`·`design.yaml`·`constitution.md`×5·`zone-registry.md`). 예상 밖 매치 0.
  - **주의 (D2)**: `config/defaults.go`의 `BrandContext.Dir`는 **survivor(OUT, 보존)**다 — Scope Decision 2로 이 SPEC 범위 밖. "0 매치" 대상에서 제외한다.
  - **주의 (D1)**: 순진한 full-tree "1건" 또는 "0건" 기대는 오답이다 — 정확한 기대는 in-scope 0 + full-tree 11(보존 1 + OUT 10)이다.

### 시나리오 2 — 신규 프로젝트 무스캐폴드
- **Given** 템플릿 md 3개 삭제 + `make build` 완료된 바이너리
- **When** `moai init /tmp/brand-test-$$`을 실행하면
- **Then** `/tmp/brand-test-$$/.moai/project/brand/`는 생성되지 않는다 (`ls`로 부재 확인).

### 시나리오 3 — migrate-agency 무손실
- **Given** `.agency/context/`에 브랜드 파일이 있는 fixture 프로젝트
- **When** `moai migrate agency`를 실행하면
- **Then** `.moai/project/brand/`에는 아무것도 쓰이지 않으며, 원본 파일은 `.agency.archived/context/`에 보존되어 접근 가능하다.

### 시나리오 4 — 기존 사용자 정리(제거 드라이버) 유지
- **Given** `.moai/project/brand/`를 이미 가진 기존 사용자 프로젝트
- **When** `moai update`를 실행하면
- **Then** `scanDeprecatedPaths()`가 `.moai/project/brand`를 감지하여(dirs.go 항목 보존 덕분) 백업 후 제거한다 — 기존 사용자 고아화 없음.

## §D.2 Edge Cases

- **EC-1**: `migrateContext` 함수가 Phase 2 외에서 참조되지 않으면 dead code로 함께 제거; 참조되면 유지 (run-phase 확인).
- **EC-2**: `pre_tool.go` line 62 주석이 brand를 예시로 언급 — 주석도 스테일이 되지 않게 정리(코드 원소 제거와 동반).
- **EC-3**: guard skill 문서(REQ-BDR-008)는 brand를 "default-deny 예시"로 기술하나 코드는 frozen으로 등록 — 기존 doc/code 불일치가 있음. 정리 시 잘못된 예시를 존재하는 다른 경로로 교체하거나 예시를 제거하되, 새 오류를 도입하지 말 것.
- **EC-4**: 로컬 `.moai/project/brand/`와 템플릿 `internal/template/templates/.moai/project/brand/` **둘 다** 삭제해야 함 — 한쪽만 삭제 시 Template-First 위반(§2 검증: 신규/삭제 파일은 양쪽 동기).

## §D.3 Quality Gate

- `go test ./...` 0 fail (AC-BDR-002)
- `go vet ./...` clean, `golangci-lint run` 기준선 유지(신규 위반 0)
- `make build` 성공 (embedded FS 재생성)
- 템플릿 중립성 CI (`template-neutrality-check.yaml`) 통과

## §D.4 Definition of Done

- [ ] AC-BDR-001a ~ AC-BDR-009 (001c 포함) 전부 PASS
- [ ] M1~M5 마일스톤 완료
- [ ] in-scope grep (harness/hook/cli, 비테스트) = **0 매치** (AC-BDR-001a)
- [ ] full-tree grep (internal/cmd/pkg, 비테스트) = **정확히 11건** survivor (AC-BDR-001c, 예상 밖 0)
- [ ] `go test ./...` 전체 통과 (실제 출력 인용, verification-claim-integrity)
- [ ] `moai init` temp 디렉터리 무스캐폴드 실측 확인
- [ ] 템플릿 + 로컬 md 6개 삭제 + `make build` 성공
- [ ] 제거 드라이버(dirs.go 항목) 보존 확인 (AC-BDR-001b)
- [ ] `SentinelHarnessFrozenConfig` 상수 보존 + `sentinel_catalog_test.go` 무변경 확인 (AC-BDR-009, D4)
- [ ] migrate `"Next step: /moai design"` 스테일 라인 제거 (AC-BDR-008, D7)
- [ ] 템플릿 편집 중립성 확인 (SPEC-ID 리크 0)
