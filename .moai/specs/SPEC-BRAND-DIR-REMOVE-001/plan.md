---
id: SPEC-BRAND-DIR-REMOVE-001
title: "Clean Removal of .moai/project/brand Directory — Implementation Plan"
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

# SPEC-BRAND-DIR-REMOVE-001 — 구현 계획 (plan.md)

## §A Context

`.moai/project/brand/`(사문 브랜드 스캐폴드)의 클린 제거. 마크다운 6개 삭제만으로는 불충분하며, 라이브 Go 배선(frozen-guard, sentinel 라우팅, migrate-agency)을 함께 해체하고 테스트를 정렬(2개 갱신: `safety/frozen_guard_test.go` + `migrate_agency_test.go`; 3개 무변경 확인: `safety_preservation_test.go`(D3) + `sentinel_catalog_test.go`(D4) + `dirs_test.go`)해야 한다. **제거 드라이버**(`dirs.go` DeprecatedPaths 항목)와 **8-sentinel 택소노미 상수**(`SentinelHarnessFrozenConfig`, D4)는 보존한다. 인접 디자인 시스템 config/문서는 범위 밖.

Tier: **M (standard)** — 3개 Go 패키지(`harness/safety`, `hook`, `cli`) + 템플릿 + 테스트에 걸친 조율된 다중 파일 제거. 신규 아키텍처 없음(behavior-preserving 삭제), 중간 blast radius. 산출물 3-file(spec/plan/acceptance) + progress.md.

## §B Known Issues / 재검증 필수 앵커

라인 번호는 drift 가능 — run-phase 진입 시 content-token 앵커로 재확인:

| 파일 | 앵커 토큰 (라인 아님) | 확인 grep |
|------|---------------------|-----------|
| `internal/harness/safety/frozen_guard.go` | `".moai/project/brand/",` (frozenPrefixes 슬라이스 원소) | `grep -n 'project/brand' internal/harness/safety/frozen_guard.go` |
| `internal/hook/pre_tool.go` | `{".moai/project/brand/", SentinelHarnessFrozenConfig}` 라우팅 원소(line 759) + 주석 line 62 — 상수 정의(line 63)는 **보존** | `grep -n 'project/brand' internal/hook/pre_tool.go` |
| `internal/cli/migrate_agency.go` | `Phase 2: copy context/ → .moai/project/brand/`, `brandDir`, `migrateContext`, help text line 589 | `grep -n 'project/brand\|brandDir\|migrateContext' internal/cli/migrate_agency.go` |
| `internal/defs/dirs.go` | `.moai/project/brand` DeprecatedPaths 항목 — **건드리지 말 것(보존)** | `grep -n 'project/brand' internal/defs/dirs.go` |

## §C Pre-flight (run-phase 진입 전 오케스트레이터/agent 확인)

1. `git fetch origin main` + divergence 확인 (병렬 세션 레이스 방지 — agent-common-protocol Pre-Spawn Sync Check).
2. 위 §B grep 4건으로 앵커 실측 재확인 (라인 drift 흡수).
3. `make build`로 fresh 바이너리 확보 후 시작 (stale-binary 함정 방지).
4. `quality.yaml` `development_mode` 확인 (ddd vs tdd) — 본 작업은 제거(behavior-preserving)이므로 DDD(characterization) 성향이 자연스러움.

## §D Constraints

- **Template-First (CLAUDE.local.md §2)**: 템플릿 md 삭제 및 skill 문서 편집은 `internal/template/templates/` 소스에서 먼저 수행 → `make build` → 로컬 동기화. 로컬 `.moai/project/brand/` 3개 md도 함께 삭제.
- **템플릿 중립성 (§15/§25)**: skill 문서(REQ-BDR-008) 편집 시 이 SPEC의 ID/REQ 토큰을 템플릿에 남기지 말 것. CI guard `internal_content_leak_test.go` + `template-neutrality-check.yaml`.
- **제거 드라이버 보존 (REQ-BDR-006)**: `dirs.go` DeprecatedPaths 항목 무변경. `dirs_test.go`도 그에 따라 무변경 예상.
- **데이터 무손실 (REQ-BDR-005)**: migrate Phase 2 제거는 `.agency/context/`가 `.agency.archived/`에 보존됨을 전제로만 허용.
- **하드코딩 금지 (§14)**: 경로 상수 인라인 신규 추가 금지 (제거만 수행).

## §E Self-Verification (plan-phase audit-ready 신호)

- [x] 실측 인벤토리 재검증 완료 (grep 4종 + defaults.go 위치 + DeprecatedPaths 소비 경로 추적)
- [x] AC(a) "zero matches" 정련 근거 확보 (dirs.go 항목 보존 → 예외 명시 필요)
- [x] scope 결정 1(migrate Phase 2 제거) + 2(design.yaml OUT) + 3(sentinel 상수 KEEP, D4) 문서화
- [x] SPEC ID regex 자가검증 PASS (decomposition 출력)
- [x] Out of Scope 섹션 `### Out of Scope —` H3 4개 + `-` bullet 존재
- [x] D3: `safety_preservation_test`는 package-harness 변수 → IN에서 무변경으로 재분류
- [x] D4: `SentinelHarnessFrozenConfig` 8-sentinel 택소노미 보존 (REQ-BDR-011), `sentinel_catalog_test` 인벤토리 추가·무변경
- [x] D7: migrate summary `"Next step: /moai design"` 제거 M2 편입 (REQ-BDR-005b)
- [x] D6: REQ-BDR 번호 순서 정렬 (§B.4 009→010, §B.5 011)

## §F Milestones (priority-based, 시간 추정 없음)

### M1 — 라이브 가드 배선 제거 (Priority High)
- `internal/harness/safety/frozen_guard.go` (package `safety`) `frozenPrefixes`에서 `.moai/project/brand/` 원소 제거 (REQ-BDR-003)
- `pre_tool.go` `frozenZonePrefixes`에서 brand 라우팅 원소(`{".moai/project/brand/", SentinelHarnessFrozenConfig}`, line 759) + line 62 주석 정리 (REQ-BDR-004). `SentinelHarnessFrozenConfig` **상수(line 63)는 8-sentinel 택소노미 멤버로 보존** (REQ-BDR-011, D4) — 라우팅 원소만 제거.
- 대응 테스트 정렬: `internal/harness/safety/frozen_guard_test.go`(package-safety, brand 단언 제거)만 갱신.
- **무변경 확인(수정 금지)**: `safety_preservation_test.go`(package-**harness** `frozenPrefixes`, brand 미포함 — D3, 다른 변수), `sentinel_catalog_test.go`(상수 보존 — D4).
- 검증: `go test ./internal/harness/... ./internal/hook/...`

### M2 — migrate-agency Phase 2 제거 (Priority High)
- `migrate_agency.go` Phase 2(`context/ → brand/`) 제거. 권고: phase 재번호 `[1/6]`..`[6/6]` → `[1/5]`..`[5/5]`, `migrateContext` 호출부 + `brandDir` 변수 + 도움말 텍스트(line 589) 제거. `migrateContext` 함수가 다른 곳에서 미사용이면 함께 제거(dead code), 사용되면 유지.
- **D7**: `/moai design`(은퇴 서브커맨드) 스테일 참조 **2건** 제거 — line ~297 `result.summary` append `"Next step: /moai design"` + line ~301 `r.logf("Run '/moai design' to continue.")` (REQ-BDR-005b). 같은 파일 편집이므로 M2 scope 내 처리. 검증: `grep -n "moai design" internal/cli/migrate_agency.go` → 0 매치
- 대응 테스트 정렬: `migrate_agency_test.go` — phase count/로그 문자열/전송 파일 수/summary 단언 갱신
- 검증: `.agency/context/` 파일이 `.agency.archived/`에 보존되는지 테스트로 확인 (REQ-BDR-005 무손실)

### M3 — 템플릿 + 로컬 md 삭제 + 재빌드 (Priority High)
- `internal/template/templates/.moai/project/brand/` 3개 md 삭제 (REQ-BDR-007)
- 로컬 `.moai/project/brand/` 3개 md 삭제
- `make build`로 embedded FS 재생성
- 검증: `moai init /tmp/<tmpdir>` → `.moai/project/brand/` 미생성 확인 (AC-BDR-004)

### M4 — guard 문서 스테일 참조 정리 (Priority Medium, SHOULD)
- `internal/template/templates/.claude/skills/moai-harness-learner/SKILL.md:~142` brand guard 예시 제거/교체 (REQ-BDR-008)
- `internal/template/templates/.claude/skills/moai/workflows/harness.md:~181` 동일
- `make build` + 로컬 동기화
- 검증: 템플릿 중립성 CI guard 통과, 편집에 SPEC-ID 미포함 (REQ-BDR-010)

### M5 — 전수 검증 (Priority High)
- **in-scope grep**: `grep -rn "project/brand" internal/harness/ internal/hook/ internal/cli/ | grep -v "_test.go"` → **0 매치** (AC-BDR-001a)
- **full-tree grep**: `grep -rn "project/brand" internal/ cmd/ pkg/ | grep -v "_test.go"` → **정확히 11건** survivor (dirs.go 보존 1 + 선언된 OUT 10; AC-BDR-001c). 예상 밖 매치 0
- D7 검증: `grep -n "moai design" internal/cli/migrate_agency.go` → summary 스테일 라인 없음 (AC-BDR-008)
- D4 검증: `SentinelHarnessFrozenConfig` 상수 + `sentinel_catalog_test.go:74` 잔존 (AC-BDR-009)
- `go test ./...` 전체 통과 (REQ-BDR-009 / AC-BDR-002)
- `go vet ./...`, `golangci-lint run` 기준선 유지

## §G Anti-Patterns (하지 말 것)

- **AP-1**: `dirs.go` DeprecatedPaths 항목까지 제거 → 기존 사용자 브랜드 디렉터리 고아화 (REQ-BDR-006 위반). "grep zero matches"를 문자 그대로 추구하다 발생하는 함정.
- **AP-2**: 마크다운 6개만 삭제하고 라이브 배선(frozen-guard/sentinel/migrate) 방치 → `moai init` 재스캐폴드 + 존재하지 않는 경로 가드.
- **AP-3**: scope를 디자인 시스템 config(`design.yaml`/`DesignConfig`)까지 확장 → blast radius 폭발, tighter scope 원칙 위반.
- **AP-4**: migrate Phase 2를 리다이렉트(새 목적지 발명)로 처리 → 은퇴 시스템에 새 배선 추가.
- **AP-5**: 초기 실패 테스트만 고치고 `go test ./...` 전체 미실행 → cascading 실패 누락 (CLAUDE.local.md §6).
- **AP-6**: skill 문서 편집에 SPEC-ID 리크 → 템플릿 중립성 CI 실패 (§25).

## §H Cross-References

- spec.md §A.1 (실측 인벤토리), §C (scope 결정), §D (out of scope)
- `CLAUDE.local.md` §2 (Template-First), §14 (하드코딩), §15/§25 (중립성), §6 (테스트 격리)
- `internal/cli/update_cleanup.go` `scanDeprecatedPaths` (제거 드라이버 소비 경로)
