---
id: SPEC-BRAND-DIR-REMOVE-001
title: "Clean Removal of .moai/project/brand Directory"
version: "0.1.0"
status: completed
created: 2026-07-08
updated: 2026-07-08
author: GOOS
priority: P2
phase: "v3.0.0"
module: "internal/harness/safety,internal/hook,internal/cli"
lifecycle: spec-anchored
tier: M
tags: "cleanup, brand, design-retirement, removal, deprecation"
depends_on: [SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001]
---

# SPEC-BRAND-DIR-REMOVE-001 — `.moai/project/brand/` 디렉터리 클린 제거

## HISTORY

| Date | Version | Author | Change |
|------|---------|--------|--------|
| 2026-07-08 | 0.1.0 | GOOS | 최초 작성 (plan-phase). 실측 의존성 인벤토리 재검증 완료 — dirs.go DeprecatedPaths 항목은 제거 드라이버로서 **보존** 결정, AC(a) 정련. |

---

## §A Context / 배경

`.moai/project/brand/`는 3개의 브랜드 컨텍스트 마크다운 파일(`brand-voice.md`, `target-audience.md`, `visual-identity.md`)을 담는 디렉터리다. 이 3개 파일은 현재 모두 미채워진 `_TBD_` 스캐폴드 상태이며, 원래 `/moai design` 브랜드 인터뷰 절차가 채우도록 설계되었다.

그러나 MoAI 디자인 생산 시스템 전체가 **은퇴(RETIRED)**했다 (`.claude/rules/moai/design/constitution.md` 은퇴 배너 참조 — `/moai design`는 더 이상 라우팅되는 서브커맨드가 아니며, 디자인 언어 요청은 `/moai plan`으로, 품질·보안 리뷰는 `/moai review`로 라우팅된다). 따라서 브랜드 디렉터리는 **소비자가 없는 사문(dead) 경로**다.

이 경로는 이미 제거 예약되어 있다: `internal/defs/dirs.go`의 `DeprecatedPaths` 레지스트리가 `.moai/project/brand`를 `DeprecatedSince/DeprecatedBy: SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001`, `RemovalSchedule: "v3.0.0"`로 표시하고 있다. 본 SPEC은 그 예약된 제거를 **실제로 운영화(operationalize)**한다.

### §A.1 실측 의존성 인벤토리 (ground-truth, plan-phase 재검증 완료)

경로 `.moai/project/brand/`는 라이브 Go 코드 + 테스트 + 템플릿 문서에 배선되어 있다. 마크다운 6개(로컬 3 + 템플릿 3)만 삭제하면 `moai init`이 경로를 재스캐폴드하고, frozen-guard/sentinel이 존재하지 않는 경로를 보호하며, `migrate agency`가 그 경로에 쓰기를 시도하는 잔재가 남는다.

| 카테고리 | 파일 | 참조 성격 | 본 SPEC 처리 |
|---------|------|----------|--------------|
| Go 소스 (라이브 사용) | `internal/harness/safety/frozen_guard.go:22` | frozen-prefix 목록 | **IN — 제거** |
| Go 소스 (라이브 사용) | `internal/hook/pre_tool.go:759` (+ line 62 주석) | `frozenZonePrefixes`의 brand 라우팅 항목 | **IN — 라우팅 항목만 제거** |
| Go 소스 (택소노미 상수) | `internal/hook/pre_tool.go:63` | `SentinelHarnessFrozenConfig` 상수 (8-sentinel `HARNESS_FROZEN_*` 택소노미 멤버) | **RETAIN — 보존 (REQ-BDR-011, D4)** |
| Go 소스 (라이브 사용) | `internal/cli/migrate_agency.go:214,220,411,589` | `context/ → brand/` 마이그레이션 대상 | **IN — 마이그레이션 제거** |
| Go 소스 (제거 드라이버) | `internal/defs/dirs.go:292` | `DeprecatedPaths` 제거 레지스트리 항목 | **RETAIN — 보존 (핵심 결정)** |
| Go 소스 (디자인 잔재) | `internal/config/defaults.go:679` | `DesignConfig.BrandContext.Dir` 기본값 | **OUT — 별도 SPEC** |
| 테스트 | `internal/harness/safety/frozen_guard_test.go` (package `safety`) | package-safety `frozenPrefixes`의 brand 단언 | **IN — 갱신** |
| 테스트 | `internal/cli/migrate_agency_test.go` | 마이그레이션 phase count/로그/summary 단언 | **IN — 갱신** |
| 테스트 | `internal/harness/safety_preservation_test.go` (package `harness`) | package-**harness** `frozenPrefixes`(4개 `.claude/` 항목, brand 없음)를 단언하는 **다른 변수**; brand 언급은 line 87 주석뿐 | **무변경 (D3 — REQ-BDR-003이 편집하는 package-safety 변수와 무관)** |
| 테스트 | `internal/harness/sentinel_catalog_test.go:74` | 8-sentinel 택소노미 상수 존재 단언 | **무변경 (D4 — REQ-BDR-011로 상수 보존)** |
| 테스트 | `internal/defs/dirs_test.go` | `DeprecatedPaths` 항목 단언 | **무변경 (항목 보존)** |
| 마크다운 (삭제 대상) | `internal/template/templates/.moai/project/brand/*.md` (3) | 스캐폴드 소스 | **IN — 삭제** |
| 마크다운 (삭제 대상) | `.moai/project/brand/*.md` (3, 로컬) | 스캐폴드 사본 | **IN — 삭제** |
| 템플릿 문서 (guard 예시) | `.claude/skills/moai-harness-learner/SKILL.md:142` | guard default-deny 예시 | **IN(SHOULD) — 스테일 참조 정리** |
| 템플릿 문서 (guard 예시) | `.claude/skills/moai/workflows/harness.md:181` | guard default-deny 예시 | **IN(SHOULD) — 스테일 참조 정리** |
| 템플릿 문서 (디자인 잔재) | `.../design.yaml:38` | 디자인 config `dir:` | **OUT — 별도 SPEC** |
| 템플릿 문서 (디자인 잔재) | `.../design/constitution.md:3,36,64,81,100` | RETIRED 디자인 헌법 | **OUT — 별도 SPEC** |
| 템플릿 문서 (디자인 잔재) | `.../zone-registry.md:625` | CONST-V3R2-072 브랜드 절 미러 | **OUT — 별도 SPEC** |
| 템플릿 문서 (디자인 잔재) | `.../CLAUDE.md:197` | 디자인 시스템 config 언급 | **OUT — 별도 SPEC** |
| config (디자인 잔재) | `.moai/config/sections/design.yaml`, `internal/config/testdata/design-valid/design.yaml` | 디자인 config `dir:` | **OUT — 별도 SPEC** |

> **핵심 발견 (제거 드라이버 보존):** `internal/defs/dirs.go`의 `.moai/project/brand` `DeprecatedPaths` 항목은 사문 잔재가 **아니다**. 이 항목은 `internal/cli/update_cleanup.go`의 `scanDeprecatedPaths()` → `update_clean_install.go`가 소비하여, `moai update` 시 **기존 사용자 프로젝트에서 브랜드 디렉터리를 삭제**하는 메커니즘이다. 이 항목을 제거하면 이미 `.moai/project/brand/`를 가진 기존 사용자는 정리를 받지 못한다(고아화). 따라서 항목은 v3.0.0 이후 사용자 업그레이드가 충분히 진행될 때까지 **의도적으로 보존**한다. 이 결정이 아래 REQ-BDR-006 및 acceptance.md AC-BDR-001a 정련의 근거다.

---

## §B Requirements (GEARS notation)

### §B.1 라이브 경로 배선 제거 (핵심)

- **REQ-BDR-001** (Ubiquitous): The `moai` binary **shall not** treat `.moai/project/brand/` as a live scaffolded, frozen, or writable location after this removal.
- **REQ-BDR-002** (Event-driven): **When** a user runs `moai init` in a fresh project, the scaffolder **shall not** create the `.moai/project/brand/` directory.
- **REQ-BDR-003** (Ubiquitous): The harness frozen-guard prefix list (`internal/harness/safety/frozen_guard.go` `frozenPrefixes`) **shall not** contain the `.moai/project/brand/` prefix.
- **REQ-BDR-004** (Ubiquitous): The pre-tool frozen-zone sentinel list (`internal/hook/pre_tool.go` `frozenZonePrefixes`) **shall not** contain the `.moai/project/brand/` entry.
- **REQ-BDR-005** (Event-driven): **When** a user runs `moai migrate agency`, the migration runner **shall not** write brand context to `.moai/project/brand/`, and the source `.agency/context/` files **shall** remain preserved in the `.agency.archived/` backup produced by the earlier backup phase (no data loss).
- **REQ-BDR-005b** (Event-driven, D7): **When** the `moai migrate agency` run completes, the runner **shall not** emit any reference to the retired `/moai design` subcommand — this covers **BOTH** stale references: the `result.summary` append `"Next step: /moai design"` (`internal/cli/migrate_agency.go:297`) AND the log line `r.logf("Run '/moai design' to continue.")` (`internal/cli/migrate_agency.go:301`). Both stale references **shall** be removed (or replaced with a neutral message) within M2 scope. Verification: `grep -n "moai design" internal/cli/migrate_agency.go` returns **0 matches** (AC-BDR-008).

### §B.2 제거 드라이버 보존 (역방향 요구)

- **REQ-BDR-006** (Ubiquitous): The `internal/defs/dirs.go` `DeprecatedPaths` registry **shall** retain the `.moai/project/brand` entry (with its `RemovalSchedule: "v3.0.0"`) so that `moai update` continues to remove the directory from existing user projects; this entry is the removal driver and **shall not** be deleted by this SPEC.

### §B.3 템플릿 소스 정리

- **REQ-BDR-007** (Ubiquitous): The template source `internal/template/templates/.moai/project/brand/` directory (3 markdown files) **shall not** exist after this removal, and the embedded template FS **shall** be regenerated (`make build`) so `moai init` reflects the removal.
- **REQ-BDR-008** (State-driven): **While** the harness-guard skill documents (`moai-harness-learner/SKILL.md`, `moai/workflows/harness.md`) reference `.moai/project/brand/` as a guard-mechanism example, the documents **should** be updated to remove the stale reference so guard docs stay consistent with the removed guard code (REQ-BDR-003/004).

### §B.4 검증 무결성 및 중립성

- **REQ-BDR-009** (Ubiquitous): The full Go test suite (`go test ./...`) **shall** pass after all in-scope edits and the affected test files are aligned.
- **REQ-BDR-010** (Ubiquitous): The template source edits introduced by this removal **shall** remain content-neutral — no internal SPEC-ID, REQ token, or audit citation **shall** leak into any `internal/template/templates/**` file (CLAUDE.local.md §15/§25).

### §B.5 dead sentinel 상수 처리 (D4)

- **REQ-BDR-011** (Ubiquitous): The `SentinelHarnessFrozenConfig` constant (`internal/hook/pre_tool.go:63`, string value `HARNESS_FROZEN_CONFIG_VIOLATION`) **shall** be RETAINED as a member of the stable 8-sentinel `HARNESS_FROZEN_*` taxonomy (defined by REQ-HRA-008 of `SPEC-V3R5-HARNESS-AUTONOMY-001`, enforced by `internal/harness/sentinel_catalog_test.go`). Only the `frozenZonePrefixes` routing entry that maps `.moai/project/brand/` to this sentinel (`pre_tool.go:759`) **shall** be removed. Rationale: the constant is exported (Go `unused` linters do not flag it when its sole routing use is removed), and semantically represents the "frozen config" zone category (the vision doc maps system-level `.moai/config/sections/*.yaml` to it), not brand exclusively — retiring it would break the 8-member catalog test and contradict REQ-HRA-008. Consequently `sentinel_catalog_test.go` requires **no change**.

---

## §C 범위 결정 근거 (scope decisions 1-3)

### Scope Decision 1 — `migrate_agency.go` Phase 2 처리

현재 `moai migrate agency`는 6-phase 구조(`[1/6]`..`[6/6]`)이며 Phase 2가 `.agency/context/ → .moai/project/brand/`를 복사한다. 브랜드 디렉터리가 제거되므로 이 phase는 존재하지 않는 목적지에 쓰기를 시도하게 된다.

**결정: Phase 2(context → brand 마이그레이션)를 제거한다.** 근거:
1. 디자인 시스템이 은퇴하여 브랜드 컨텍스트의 라이브 소비자가 없다 — 마이그레이션 목적지 자체가 무의미하다.
2. Phase 1이 `.agency/` 전체를 `.agency.archived/`로 백업하므로, `.agency/context/`의 브랜드 파일은 **아카이브에 보존**된다(데이터 손실 없음).
3. 리다이렉트(다른 위치로 복사)는 은퇴한 시스템을 위해 새 목적지를 발명하는 것이므로 클린 제거 원칙에 반한다.

구현 세부(phase 번호 재조정 `[1/6]`..`[6/6]` → `[1/5]`..`[5/5]` vs no-op 유지)는 run-phase 소관이며 plan.md §F에 권고를 남긴다.

**D7 추가 결정 — migrate 스테일 `/moai design` 참조 (2건):** `migrate_agency.go`가 은퇴 서브커맨드 `/moai design`을 **두 곳**에서 안내한다 — (1) line 297 `result.summary` append `"Next step: /moai design"`, (2) line 301 `r.logf("Run '/moai design' to continue.")`. **두 참조 모두** 제거(또는 중립 메시지로 교체)한다 (REQ-BDR-005b). 이 라인들은 M2에서 편집하는 **같은 파일** 내부이므로 M2 scope에 편입한다. 별도 SPEC로 미루지 않는 이유: (a) 이미 열려 있는 파일, (b) 은퇴 서브커맨드를 사용자에게 안내하는 명백한 오류 메시지, (c) migrate_agency_test.go summary/로그 단언이 이미 M2 갱신 대상. AC-BDR-008의 grep `grep -n "moai design" internal/cli/migrate_agency.go`는 두 라인 모두 매치하므로, 두 라인 제거 후 0 매치로 일관 검증된다.

### Scope Decision 2 — `design.yaml` + RETIRED `design/constitution.md` (tighter scope 채택)

이 SPEC은 **브랜드 디렉터리 + 그것을 라이브 경로로 취급하는 코드**만 제거한다. 인접 디자인 시스템 잔재(`design.yaml`의 `dir:`, `defaults.go`의 `DesignConfig.BrandContext.Dir`, RETIRED `design/constitution.md`, `zone-registry.md` CONST-V3R2-072 브랜드 절, `CLAUDE.md` 디자인 config 언급, testdata fixture)는 **범위 밖(별도 후속 SPEC)**이다. 근거:
1. **더 좁은 범위** — 태스크 지침이 명시적으로 tighter scope를 권고.
2. **결합도** — `BrandContext.Dir`는 `defaultDesignConfig()` 내부이며, 제거하면 `DesignConfig` 구조체 스키마·`Validate`·다수 testdata fixture로 blast radius가 확대된다(디자인 시스템 config 은퇴 영역).
3. **무해한 불활성 문자열** — `design.yaml`의 `dir:` 값과 `defaults.go` 기본값은 은퇴한 서브시스템의 config 문자열일 뿐, 디렉터리를 생성하지 않는다. 남겨도 AC(c) `moai init` 무스캐폴드를 깨지 않는다(스캐폴드는 템플릿 md 존재 여부로 결정되며 config `dir:` 값과 무관).
4. **테스트 안전** — `internal/config/testdata/design-valid/design.yaml`을 건드리지 않으면 config 검증 테스트가 무변경으로 통과 유지.
5. `constitution.md`는 FROZEN 절 원천(zone-registry가 미러)이자 역사적 기록으로 의도적 보존 중 — 삭제는 헌법 레지스트리와 결합.

**경계 명시:** 디자인 시스템 config/문서 잔재 은퇴는 `SPEC-DESIGN-CONFIG-RETIRE-001`(가칭, 미저작) 후속 SPEC 소관.

### Scope Decision 3 — `SentinelHarnessFrozenConfig` 상수 처리 (D4, KEEP)

`pre_tool.go:759`의 `frozenZonePrefixes` brand 라우팅 항목을 제거하면(REQ-BDR-004), 이 sentinel을 사용하는 라우팅 원소가 사라진다(brand가 유일 사용처). 상수 `SentinelHarnessFrozenConfig`(line 63)를 함께 retire할지 결정이 필요하다.

**결정: 상수를 보존한다(KEEP).** 실측 근거:
1. **8-sentinel 택소노미 멤버** — `HARNESS_FROZEN_CONFIG_VIOLATION`은 `SPEC-V3R5-HARNESS-AUTONOMY-001` REQ-HRA-008이 정의하는 8-member `HARNESS_FROZEN_*` 카탈로그의 일원이며, `internal/harness/sentinel_catalog_test.go:74`가 이 8개 상수의 존재를 단언한다. 상수 retire 시 카탈로그 테스트가 깨지고 REQ-HRA-008과 모순된다.
2. **카테고리 의미** — 이 sentinel은 "frozen config" 존 카테고리를 표현하며, vision 문서(`harness-autonomy-vision-2026-05-18.md`)는 시스템 레벨 `.moai/config/sections/*.yaml`도 이 sentinel에 매핑한다. brand는 하나의 referent일 뿐 유일 의미가 아니다.
3. **lint-safe** — 상수는 exported(`Sentinel...`)이므로 유일 사용처가 제거되어도 Go `unused` 계열 린터가 flag하지 않는다.

**귀결:** 라우팅 원소(line 759)만 제거하고 상수(line 63)는 보존 → `sentinel_catalog_test.go`는 **무변경**. (REQ-BDR-011)

---

## §D Exclusions — 범위 밖 (out of scope)

본 SPEC이 명시적으로 **하지 않는** 작업을 열거한다.

### Out of Scope — 디자인 시스템 config/스키마 은퇴

- `internal/config/defaults.go`의 `DesignConfig` / `DesignBrandContext.Dir` 기본값 변경 또는 제거
- `internal/template/templates/.moai/config/sections/design.yaml`의 `dir:` 필드 변경
- 로컬 `.moai/config/sections/design.yaml` 및 `internal/config/testdata/design-valid/design.yaml` fixture 변경
- 후속 `SPEC-DESIGN-CONFIG-RETIRE-001`(가칭) 소관

### Out of Scope — RETIRED 디자인 헌법 및 zone-registry 미러

- `internal/template/templates/.claude/rules/moai/design/constitution.md` 삭제 또는 브랜드 참조 제거 (FROZEN 절 원천·역사적 기록으로 보존)
- `internal/template/templates/.claude/rules/moai/core/zone-registry.md:625` CONST-V3R2-072 브랜드 절 제거 (헌법 레지스트리 결합)
- `internal/template/templates/CLAUDE.md:197` 디자인 시스템 config 언급 정리

### Out of Scope — 제거 드라이버 삭제 금지

- `internal/defs/dirs.go`의 `.moai/project/brand` `DeprecatedPaths` 항목 삭제 (REQ-BDR-006에 의해 의도적 보존 — 기존 사용자 `moai update` 정리 드라이버)

### Out of Scope — 구현·게이트·git 작업

- 실제 코드 구현 / 테스트 실행 (run-phase, manager-develop 소관)
- Git 브랜치·커밋·PR (manager-git / 오케스트레이터 소관)
- CHANGELOG·README 갱신 (sync-phase, manager-docs 소관)

---

## §E Cross-References

- **제거 예약 원천**: `internal/defs/dirs.go:292` `DeprecatedPaths[.moai/project/brand]` (`RemovalSchedule: "v3.0.0"`, `DeprecatedBy: SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001`)
- **제거 드라이버 소비 경로**: `internal/cli/update_cleanup.go` `scanDeprecatedPaths()` → `internal/cli/update_clean_install.go`
- **디자인 은퇴 배너**: `.claude/rules/moai/design/constitution.md` (RETIRED)
- **선행 SPEC**: `SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001` (경로를 제거 레지스트리에 등록)
- **Template-First 규율**: `CLAUDE.local.md` §2 (`make build` + local sync)
- **템플릿 중립성**: `CLAUDE.local.md` §15/§25
