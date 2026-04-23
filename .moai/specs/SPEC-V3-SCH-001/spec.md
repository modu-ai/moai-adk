---
id: SPEC-V3-SCH-001
title: Formal Config Schema Framework (validator/v10 + JSON Schema export)
version: 0.1.0
status: draft
created: 2026-04-22
updated: 2026-04-22
author: manager-spec
priority: Critical
phase: "Phase 1 — Foundation"
module: "internal/config/schema/"
dependencies: []
related_gap: [156, 157, 158, 163]
related_theme: "Theme 3 — Schema & Validation Layer"
breaking: true
bc_id: [BC-006]
lifecycle: spec-anchored
tags: "v3, schema, validator, config, validation, foundation, P0"
---

# SPEC-V3-SCH-001: Formal Config Schema Framework

## HISTORY

- 2026-04-22 v0.1.0: 최초 작성. master-v3 §3.3, gap-matrix rows #156/#157/#163, W1.5 §6.1/§6.7 근거. Theme 3 — Schema & Validation Layer의 핵심 SPEC. 모든 v3 SPEC의 의존성 기반.

---

## 1. Goal (목적)

moai-adk-go v2.12 기준 22개 `.moai/config/sections/*.yaml` 파일 모두 `gopkg.in/yaml.v3` 구조체 태그만으로 파싱되어 있어, `development_mode: tddd`와 같은 오타가 silent zero-value로 무시된다. Claude Code의 Zod v4 기반 ~110 스키마 (W1.5 §6.1) 대비 moai는 **0개**의 formal schema를 보유하며, 이는 향후 Tier 1/2 테마(Hook Protocol v2, Plugin System, Migration Framework)의 기반 인프라 부재를 의미한다.

본 SPEC은 `go-playground/validator/v10`를 채택해 22개 config section 각각에 (a) 구조체 + validate 태그, (b) `Validate() error` 메서드, (c) JSON Schema export를 부여하고, `moai doctor config --fix`로 오타를 자동 수정한다. 동시에 strict validation 도입은 v2.x 설정의 silent typo를 error로 승격하므로 dual-parse shim으로 1 minor version 동안 후방 호환을 유지한다.

### 1.1 배경

W1.6 §10.1에서 확인된 moai 현재 상태:
- `internal/config/`: `manager.go`, `loader.go`, `types.go`, `defaults.go`, `envkeys.go`, `validation.go` — 구조체만 존재, 런타임 검증 없음
- `.moai/config/sections/*.yaml`: 22개 섹션 (user, language, quality, workflow, llm, system, project, design, harness, memory 등)
- 오타 발견 경로: **사용자 버그 리포트**가 유일함

gap-matrix #156 (Critical): Zod v4 + lazySchema 패턴 부재 → Go-idiomatic validator/v10로 파리티 확보.
gap-matrix #157 (High): Generated TypeScript types from Zod → Go에서는 `json.RawMessage` 기반 JSON Schema export로 에디터 통합.
gap-matrix #163 (High): `parseSettingsFile` + `updateSettingsForSource` round-trip safety → validator.Struct() + yaml.Marshal() round-trip.

### 1.2 비목표 (Non-Goals)

- **CUE 도입 금지**: master-v3 §9 open question #6에 따라 Go-native validator/v10 채택. CUE는 external language/cross-file dependency로 moai의 9-direct-dep 철학 (W1.6 §14.1)과 충돌.
- **Zod-compatible 런타임 빌드 금지**: TypeScript 생태계 재현 불필요. JSON Schema draft-07 export만 제공.
- **신규 Heavy 의존성 금지**: `invopop/jsonschema` (~2KB binary size 증가)만 허용. OTEL, reflect-heavy frameworks 거부.
- **Agent/Skill/Plugin Manifest 스키마는 본 SPEC 범위 외**: SPEC-V3-AGT-001, SPEC-V3-SKL-001, SPEC-V3-PLG-001에서 별도 정의. 본 SPEC은 `.moai/config/sections/*.yaml` 22개에만 한정.
- **Settings Source Layering 병합은 SPEC-V3-SCH-002로 분리**

---

## 2. Scope (범위)

### 2.1 In Scope

- `internal/config/schema/` 신규 패키지 생성 (section당 1개 파일)
- 22개 YAML section 모두 Go 구조체 + `validate:"..."` 태그 + `Validate() error` 메서드
- `internal/config/schema/registry.go`: section name → validator 매핑 레지스트리
- `moai doctor config` 서브커맨드: 22개 YAML 로드 + 검증 + 위반 리포트
- `moai doctor config --fix`: 결정적(deterministic) 오류 auto-repair (e.g., `tddd` → `tdd`)
- JSON Schema export: `docs-site/static/schemas/*.json` 22개 파일 생성 (`make schemas` 타겟)
- YAML 파일 헤드에 `# yaml-language-server: $schema=https://adk.mo.ai.kr/schemas/<section>.json` 주입 (template-first)
- Dual-parse shim: strict 실패 시 v2 lenient fallback + systemMessage 경고
- `MOAI_CONFIG_STRICT=0` 환경변수 escape (master-v3 BC-006)
- Round-trip safety 보증: Validate 후 re-marshal 시 시맨틱 보존

### 2.2 Out of Scope (Exclusions — What NOT to Build)

- Agent frontmatter 스키마 (SPEC-V3-AGT-001 관할)
- Skill frontmatter 스키마 (SPEC-V3-SKL-001 관할)
- Hook declaration 스키마 확장 (SPEC-V3-HOOKS-001 관할)
- Plugin manifest 스키마 (SPEC-V3-PLG-001 관할)
- Settings source layering (user/project/local 병합) — SPEC-V3-SCH-002 관할
- Enterprise managed/policy settings tier — v3.2 이후 정의
- Binary-size 최적화 (jsonschema 라이브러리 fork/custom reflect) — 추후 최적화 단계
- `moai config edit` interactive editor — v3.2 이후

---

## 3. Environment (환경)

- 런타임: Go 1.26 (go.mod:3), moai-adk-go v3.0.0-alpha.1+
- 외부 의존성 (신규 추가):
  - `github.com/go-playground/validator/v10` (Go-native struct tag validator, ~200KB)
  - `github.com/invopop/jsonschema` (Go struct → JSON Schema draft-07 reflect, optional)
- YAML 파서: `gopkg.in/yaml.v3` (기존, 변경 없음)
- 영향 디렉터리:
  - `internal/config/schema/` (신규 22 파일 + registry)
  - `internal/cli/doctor_config.go` (신규 또는 기존 doctor.go 확장)
  - `internal/template/templates/.moai/config/sections/*.yaml` (스키마 주석 추가)
  - `docs-site/static/schemas/` (JSON Schema export 대상)
  - `Makefile` (`schemas` 타겟)
- 대상 OS: macOS / Linux / Windows 동등

---

## 4. Assumptions (가정)

- A-001 (High): v2.12 기준 22개 YAML section은 `internal/config/` 구조체에 1:1 매핑되어 있다. 추가/제거 없이 태그 증설만으로 스키마화 가능.
- A-002 (High): `go-playground/validator/v10`의 built-in validators (oneof, min/max, required, dive, semver 등)로 22개 섹션의 제약 90% 이상을 커버한다. 나머지 10%는 custom validator로 대응.
- A-003 (Medium): 사용자 기존 config에 silent typo가 존재할 수 있으며, `--fix` 모드에서 결정적으로 복구 가능한 비율은 ~70%이다. 나머지는 사용자 개입 필요.
- A-004 (High): JSON Schema draft-07은 VS Code / JetBrains IDE의 YAML language server (redhat.vscode-yaml)와 호환된다.
- A-005 (Medium): `invopop/jsonschema` 라이브러리가 Go 구조체 → JSON Schema 변환 시 validator 태그를 충실히 반영한다. 미반영 시 hand-rolled schema 함수로 대체.
- A-006 (High): Round-trip (YAML → struct → validate → YAML) 시 주석 소실은 허용되나 값은 보존된다. moai가 사용자 YAML을 직접 write-back 하는 경로는 migration 한정.

---

## 5. Requirements (EARS 요구사항)

### 5.1 Ubiquitous Requirements

**REQ-SCH-001-001 (Ubiquitous)**
시스템은 **항상** 22개 `.moai/config/sections/*.yaml` section 각각에 대해 `internal/config/schema/<section>_schema.go`에 Go 구조체 + `validate:"..."` 태그를 제공한다.

**REQ-SCH-001-002 (Ubiquitous)**
시스템은 **항상** 각 스키마 구조체에 `Validate() error` 메서드를 노출해 validator/v10 기반 검증을 수행한다.

**REQ-SCH-001-003 (Ubiquitous)**
시스템은 **항상** 각 스키마 구조체에 `JSONSchema() ([]byte, error)` 메서드를 노출해 draft-07 호환 JSON Schema를 반환한다.

**REQ-SCH-001-004 (Ubiquitous)**
시스템은 **항상** `internal/config/schema/registry.go`의 글로벌 레지스트리를 통해 section name(문자열) → validator 매핑을 제공한다.

**REQ-SCH-001-005 (Ubiquitous)**
시스템은 **항상** 스키마 정의 시점에 Zero-trust 원칙을 적용한다: 모든 문자열 열거(oneof)는 화이트리스트 방식, 모든 수치 범위는 min/max 경계 명시.

### 5.2 Event-Driven Requirements

**REQ-SCH-001-010 (Event-Driven)**
**When** 사용자가 `moai doctor config`를 실행하면, the 시스템 **shall** 22개 YAML 파일 모두를 로드하고 검증해 위반 사항을 `{severity, path, line, message, suggestion}` 구조로 표시한다.

**REQ-SCH-001-011 (Event-Driven)**
**When** `moai doctor config --fix`가 실행되면, the 시스템 **shall** 결정적으로 복구 가능한 오류(typo on oneof, out-of-range with obvious clamp)를 auto-repair하고, 복구 전 백업을 `.moai/backups/{ISO-timestamp}/`에 저장한다.

**REQ-SCH-001-012 (Event-Driven)**
**When** `make schemas` 타겟이 실행되면, the 시스템 **shall** 22개 JSON Schema 파일을 `docs-site/static/schemas/<section>.json`에 minified 형태로 export한다.

**REQ-SCH-001-013 (Event-Driven)**
**When** config 로더가 YAML 파일을 파싱하고 validator.Struct() 검증이 실패하면, the 시스템 **shall** dual-parse shim 모드에서 v2 lenient fallback을 수행하고 `systemMessage`로 경고를 표시한다.

**REQ-SCH-001-014 (Event-Driven)**
**When** 스키마 변경 사항이 registry.go에 등록되면, the 시스템 **shall** 빌드 타임에 `make schemas` CI 체크를 강제해 export된 JSON Schema가 최신 상태임을 보증한다.

### 5.3 State-Driven Requirements

**REQ-SCH-001-020 (State-Driven)**
**While** `MOAI_CONFIG_STRICT=0` 환경변수가 설정된 동안, the 시스템 **shall** validator 실패를 error 대신 warning으로 강등하고 lenient parse를 유지한다.

**REQ-SCH-001-021 (State-Driven)**
**While** `moai doctor config --fix`가 수행 중인 동안, the 시스템 **shall** 원본 파일을 수정하기 전에 반드시 `.moai/backups/{timestamp}/<section>.yaml.bak` 스냅샷을 생성한다.

**REQ-SCH-001-022 (State-Driven)**
**While** round-trip 검증 모드가 활성화된 동안 (테스트 전용), the 시스템 **shall** YAML → struct → Validate() → yaml.Marshal() 사이클에서 값 변경이 발생하지 않음을 확인한다.

### 5.4 Optional Requirements

**REQ-SCH-001-030 (Optional)**
**Where** 사용자의 에디터가 YAML language server(redhat.vscode-yaml)를 지원할 때, the 시스템 **shall** 각 YAML 파일 헤드 주석으로 `$schema` URL을 주입해 inline validation과 autocomplete를 활성화한다.

**REQ-SCH-001-031 (Optional)**
**Where** 스키마가 custom validator를 필요로 하는 경우, the 시스템 **shall** `internal/config/schema/validators.go`에 재사용 가능한 validator (semver, permission-rule-syntax, glob-pattern)를 등록한다.

**REQ-SCH-001-032 (Optional)**
**Where** 사용자가 `moai doctor config --json` 플래그를 사용하면, the 시스템 **shall** 기계 판독 가능한 JSON 형식으로 검증 결과를 출력해 CI 파이프라인 통합을 지원한다.

### 5.5 Unwanted Behavior (Must Not)

**REQ-SCH-001-040 (Unwanted)**
시스템은 validator 태그가 없는 필드를 silent accept **하지 않아야 한다**. 태그 누락은 `moai doctor schema --validate` 시점에 error로 보고된다.

**REQ-SCH-001-041 (Unwanted)**
시스템은 사용자 YAML에 대한 lossy write-back을 **수행하지 않아야 한다**. `--fix` 모드에서도 주석은 보존하거나 별도 파일로 분리한다.

**REQ-SCH-001-042 (Unwanted)**
시스템은 v3.2 이후 `MOAI_CONFIG_STRICT=0` fallback을 제공**하지 않아야 한다**. v3.2에서는 strict 기본, v4.0에서 환경변수 자체 제거.

### 5.6 Complex Requirements

**REQ-SCH-001-050 (Complex)**
**While** dual-parse shim 기간 (v3.0~v3.2) 동안, **when** config 로더가 strict 검증 실패를 감지하면, the 시스템 **shall** (a) lenient 파서로 fallback, (b) 위반 내용을 `.moai/reports/config-validation-{timestamp}.md`에 기록, (c) 사용자에게 `systemMessage`로 "Config section {name} failed v3 schema; using v2 best-effort parse. See docs-site/migration/v3-schemas.md" 경고를 표시한다.

---

## 6. Acceptance Criteria (수용 기준 요약)

**AC-SCH-001-01**: 22개 `.moai/config/sections/*.yaml` 모두에 대응하는 `internal/config/schema/<section>_schema.go` 파일이 존재하고, 각각 `Validate() error` 메서드를 노출한다. `registry.go`의 맵에 22개 엔트리 모두 등록된다.

**AC-SCH-001-02**: 고의 오타 config (e.g., `development_mode: tddd`, `test_coverage_target: 150`)에 대해 `moai doctor config`가 2개 error를 보고하고, line 번호와 suggestion을 포함한다. `--fix` 모드에서 복구 후 재실행 시 0 error.

**AC-SCH-001-03**: `make schemas` 실행 후 `docs-site/static/schemas/` 디렉터리에 22개 `.json` 파일이 생성되고, 각각 draft-07 준수. `jsonschema -c` CLI 검증 통과.

**AC-SCH-001-04**: v2.12 기준 사용자 corpus 10개 프로젝트에 대해 `moai doctor config`를 실행했을 때, dual-parse fallback 발생률 ≤ 30%이고, `--fix` auto-repair 후 전체 통과.

**AC-SCH-001-05**: `MOAI_CONFIG_STRICT=0` 환경변수 설정 시 v2 lenient parse가 활성화되고 error 대신 warning이 표시된다. 환경변수 unset 시 strict 모드 복귀 확인.

**AC-SCH-001-06**: Round-trip 테스트(`internal/config/schema/roundtrip_test.go`)가 22개 섹션 모두에 대해 값 불변성을 검증하며 통과한다.

**AC-SCH-001-07**: YAML language server integration: VS Code에서 템플릿으로 배포된 quality.yaml 열면 `$schema` 주석으로 autocomplete가 활성화된다.

**AC-SCH-001-08**: Binary size 증가 ≤ 600 KB (validator/v10 + invopop/jsonschema 합산 측정치 근거).

---

## 7. Constraints (제약)

- **[HARD] 9-direct-dep 철학 (W1.6 §14.1)**: 신규 의존성은 validator/v10, invopop/jsonschema 2개로 제한. 추가 요청 시 open question 필요.
- **[HARD] Template-first 원칙 (CLAUDE.local.md §2)**: 스키마 주석이 포함된 YAML은 `internal/template/templates/.moai/config/sections/*.yaml` 먼저 수정 후 `make build`.
- **[HARD] Language-neutral 원칙 (CLAUDE.local.md §15)**: 16개 언어 동등 취급. 스키마에 특정 언어 bias 금지 (e.g., default `primary_language: go` 하드코드 금지).
- **[HARD] Round-trip safety**: validator에 `.transform()` 등 변형 로직 금지 (W1.5 §6.2 comment 참조). validator는 검증만, 변환은 loader가 담당.
- Binary size 증가 상한 4MB (master-v3 success metric). 본 SPEC 기여분 ≤ 1 MB.

---

## 8. Risks & Mitigations (리스크 및 완화)

| Risk ID | 설명 | 확률 | 영향 | 완화 |
|---------|------|------|------|------|
| R-SCH-001-01 | v2 사용자 설정에서 silent typo가 대량 발견되어 `--fix` 복구율이 낮음 | Medium | High | Dual-parse shim 1 minor version 유지; `--fix` 우선, 수동 교정 가이드 문서(docs-site/migration/v3-schemas.md) 제공 |
| R-SCH-001-02 | validator/v10 태그 문법의 표현력 부족으로 복잡 제약 누락 | Medium | Medium | `internal/config/schema/validators.go`에 custom validator 허용; 태그 불가 시 `Validate()` 메서드 내 if-block 허용 (단 문서화 필수) |
| R-SCH-001-03 | invopop/jsonschema 리플렉션 오류로 JSON Schema export 부정확 | Low | Medium | `make schemas-check` CI 테스트에서 각 스키마에 대해 positive/negative 샘플 교차 검증; 실패 시 hand-rolled fallback |
| R-SCH-001-04 | Round-trip 시 yaml.v3 파서의 anchor/alias 소실 | Low | Low | moai config는 anchor 미사용; 테스트 커버리지로 감지 |
| R-SCH-001-05 | 에디터 integration 실패 (CDN 404, offline 환경) | Low | Low | `$schema` URL은 선택적 코멘트; 오프라인에서도 validator 동작 무관 |
| R-SCH-001-06 | Binary size 목표 초과 | Low | Medium | invopop/jsonschema optional로 build tag 분리 가능; `//go:build jsonschema` 태그로 분리 빌드 |

---

## 9. Dependencies (의존성)

### 9.1 Blocked by

- 없음 (Phase 1 Foundation의 첫 SPEC)

### 9.2 Blocks

- **SPEC-V3-SCH-002** (Settings source layering) — 본 SPEC의 스키마 정의가 선행되어야 merge 대상 타입이 확정됨.
- **SPEC-V3-MIG-001** (Migration framework) — 마이그레이션이 config 수정 후 Validate() 호출로 idempotency 확인.
- **SPEC-V3-MIG-002** (v2→v3 migration content) — M01 `template_version` 마이그레이션은 project.yaml 스키마에 의존.
- **SPEC-V3-HOOKS-001 ~ 006** — Hook 선언 검증이 settings.json 스키마 확장에 의존 (discriminated union).
- **SPEC-V3-AGT-001** (Agent frontmatter v2) — agent_schema.go 구조를 본 SPEC의 registry 패턴에서 재사용.
- **SPEC-V3-PLG-001** (Plugin manifest) — plugin.json 스키마가 validator/v10 패턴 준수.
- **SPEC-V3-CLI-001** (CLI subcommand restructure) — `moai doctor config` subcommand 구조가 본 SPEC에서 정의됨.

### 9.3 Related

- gap-matrix #156, #157, #158, #160, #161, #163
- W1.5 §6.1–§6.7 (Zod v4 + lazySchema + parseSettingsFile round-trip)
- W1.6 §10.1 (현재 moai config 상태)
- CLAUDE.local.md §14 (하드코딩 방지)

---

## 10. Traceability (추적성)

| Requirement | Implementation | Verification |
|-------------|----------------|--------------|
| REQ-SCH-001-001, 002, 003 | `internal/config/schema/<section>_schema.go` (22 files) | `schema_registry_test.go` — 22 entries present |
| REQ-SCH-001-004 | `internal/config/schema/registry.go` | `TestRegistryCompleteness` |
| REQ-SCH-001-010, 011 | `internal/cli/doctor_config.go` | `TestDoctorConfigReportsErrors`, `TestDoctorConfigFixReplacesTypos` |
| REQ-SCH-001-012 | `Makefile` `schemas` target + `cmd/moai-schemas/main.go` | `TestSchemaExportRoundtrip` |
| REQ-SCH-001-013, 020, 050 | `internal/config/loader.go` dual-parse shim | `TestDualParseFallback`, `TestStrictEnvVarOverride` |
| REQ-SCH-001-014 | CI job `schemas-check` in `.github/workflows/ci.yml` | CI green |
| REQ-SCH-001-021 | `internal/config/schema/fix.go` backup step | `TestFixBackupSnapshot` |
| REQ-SCH-001-022 | `internal/config/schema/roundtrip_test.go` | `TestRoundTripAllSections` |
| REQ-SCH-001-030 | Template YAML 주석 | `TestTemplateSchemaCommentPresent` |
| REQ-SCH-001-031 | `internal/config/schema/validators.go` | `TestCustomValidators` |
| REQ-SCH-001-032 | `internal/cli/doctor_config.go --json` | `TestDoctorConfigJSONOutput` |
| REQ-SCH-001-040, 041, 042 | Code review + lint | Manual QA |
| AC-SCH-001-01 ~ 08 | 해당 REQ 커버리지 전체 | `go test ./internal/config/schema/...` + `make schemas` + 10-project corpus run |
| BC-006 (Breaking) | Dual-parse shim, `MOAI_CONFIG_STRICT=0` escape | master-v3 Breaking Changes Catalog §4 |

---

End of SPEC-V3-SCH-001.
