---
id: SPEC-V3R2-MIG-003
title: Config Loader Completeness (5 unloaded YAML sections)
version: "0.1.0"
status: draft
created: 2026-04-23
updated: 2026-04-23
author: Wave 2 SPEC writer (Layer 6/7/Cleanup)
priority: P1 High
phase: "v3.0.0 — Phase 8 — Migration Tool + Docs"
module: "internal/config/types.go, internal/config/loader.go, .moai/config/sections/"
dependencies:
  - SPEC-V3R2-EXT-004
related_gap:
  - r6-config-audit
  - r6-unused-yaml-sections
related_theme: "Theme 2 — Runtime Hardening"
breaking: true
bc_id: [BC-V3R2-013]
lifecycle: spec-anchored
tags: "config, yaml, loader, harness, constitution, context, interview, design, sunset, v3"
---

# SPEC-V3R2-MIG-003: Config Loader Completeness

## HISTORY

| Version | Date       | Author | Description                                                            |
|---------|------------|--------|------------------------------------------------------------------------|
| 0.1.0   | 2026-04-23 | Wave 2 | Initial SPEC — add loaders for 5 unloaded YAML sections or mark dormant |

---

## 1. Goal (목적)

R6 audit §5는 `.moai/config/sections/` 아래 **5개 YAML 섹션이 Go 런타임에서 로드되지 않음**을 확인했다: `constitution.yaml`, `context.yaml`, `interview.yaml`, `design.yaml` (런타임 미사용; `migrate_agency.go`만 참조), `harness.yaml`. 추가로 `sunset.yaml`은 struct만 있고 실질 hot path 없음. CLAUDE.md와 agent prompts는 이 값들을 사용하지만 Go 측에서 schema validation + runtime enforcement가 없어 drift가 가능하다. 본 SPEC은 각 섹션에 대해 (a) Go 스키마(`internal/config/types.go`) 추가, (b) loader 함수 추가, (c) 기본 사용처 1개 이상 확립 — **또는** (d) 공식적으로 "dormant(사용 안함)" 판정을 내리고 documentation으로 명시한다.

### 1.1 배경

R6 §5.1 per-yaml usage grep: 13개 section은 Go runtime이 활용, 1개(sunset)는 dormant(struct 존재 but no hot path), 5개(constitution/context/interview/design/harness)는 **template-only** (CLAUDE.md + skill 문서에서만 언급). §5.3 schema gaps: "constitution, context, interview, design, harness, mx"의 struct가 `internal/config/types.go`에 부재.

§5 v3 action items: "Add loaders for constitution.yaml, context.yaml, interview.yaml, design.yaml, harness.yaml. Either remove sunset.yaml or activate it. Unify workflow.yaml schema (currently only role_profiles are read)."

### 1.2 비목표 (Non-Goals)

- 새 YAML 섹션 추가
- 기존 13개 loaded section 재작성
- Config migration (format 변경)
- Config hot-reload (SPEC-V3R2-MIG-002의 configChangeHandler와 연동 가능하나 본 SPEC scope는 loader만)
- Config encryption
- Config UI
- `workflow.yaml` schema 전체 통합 (R6 §5 추가 권장, 별도 SPEC)

---

## 2. Scope (범위)

### 2.1 In Scope

- **Owns**: `internal/config/types.go` 확장 (5개 신규 struct), `internal/config/loader.go` 확장 (5개 신규 loader), `sunset.yaml` 활성화 또는 retire 판정.
- 5개 YAML 섹션 판정:
  - `harness.yaml`: **ADD LOADER** — harness level(minimal/standard/thorough) 판정 로직이 SPEC-V3R2-WF-003의 mode auto-select에 필요 (critical).
  - `constitution.yaml`: **ADD LOADER** — 기술 제약 체크 (forbidden libraries, approved frameworks) runtime enforcement 근거.
  - `context.yaml`: **ADD LOADER** — CLAUDE.md Section 16 Context Search Protocol의 token budget, staleness threshold 런타임 활용.
  - `interview.yaml`: **ADD LOADER** — SPEC-V3R2-WF-003의 discovery mode에서 round/question limit 참조.
  - `design.yaml`: **ADD LOADER for runtime** — 현재 `migrate_agency.go`에서만 사용; GAN loop / sprint contract / pipeline adaptation 값들을 runtime 레이어에서 활용.
- `sunset.yaml`: struct 존재만 해도 4개 참조뿐이므로 **dormant 공식화** — docstring "dormant (not actively enforced)" + 차기 SPEC에서 활성화 대기.
- 각 loader에 단위 테스트 추가.
- schema docs: 각 struct의 godoc에 사용 hot path 1개 이상 명시.

### 2.2 Out of Scope (Exclusions — What NOT to Build)

- 신규 YAML 섹션 추가
- 13개 loaded section의 loader 재작성
- Config hot-reload 구현 (MIG-002가 trigger 담당)
- Config encryption
- Config validation UI
- `workflow.yaml`의 role_profiles 외 스키마 통합
- `mx.yaml` struct 신설 (R6 §5.3 언급; 현재 ad-hoc parsing 유지; 별도 SPEC)

---

## 3. Environment (환경)

- 런타임: moai-adk-go (Go 1.26+), `internal/config/`, `.moai/config/sections/`
- 영향 디렉터리:
  - 수정: `internal/config/types.go`, `loader.go`, `loader_test.go`
  - 신설: `internal/config/harness.go`, `constitution.go`, `context.go`, `interview.go` (각 파일 분리 권장)
  - 참조: `.moai/config/sections/{harness,constitution,context,interview,design,sunset}.yaml`
- 외부 레퍼런스: R6 §5 config audit

---

## 4. Assumptions (가정)

- 5개 YAML 섹션의 현재 파일 스키마는 stable하고 변경 없이 Go struct로 reverse-engineer 가능하다.
- `gopkg.in/yaml.v3`은 이미 의존성이며 신규 패키지 도입 없다.
- 각 섹션의 default 값은 기존 template에서 추출 가능하다.
- `sunset.yaml`의 SunsetConfig struct는 손대지 않고 유지한다 (backward compat).
- harness.yaml의 3개 level(minimal/standard/thorough)은 이미 복수 agent prompts에서 참조되고 있으므로 runtime loader 도입은 기능 확장이 아닌 gap 보완이다.

---

## 5. Requirements (EARS 요구사항)

### 5.1 Ubiquitous Requirements

**REQ-MIG003-001**
The system **shall** define Go structs in `internal/config/types.go` for 5 sections: `HarnessConfig`, `ConstitutionConfig`, `ContextConfig`, `InterviewConfig`, `DesignConfig`.

**REQ-MIG003-002**
The system **shall** provide loader functions `LoadHarnessConfig`, `LoadConstitutionConfig`, `LoadContextConfig`, `LoadInterviewConfig`, `LoadDesignConfig` that read the corresponding YAML file.

**REQ-MIG003-003**
Each loader **shall** return a typed struct on success and an aggregated error on validation failure.

**REQ-MIG003-004**
Each loader **shall** provide sensible defaults when the YAML file is absent (graceful degradation).

**REQ-MIG003-005**
Each new struct **shall** have a godoc block documenting at least one runtime hot path that consumes it.

**REQ-MIG003-006**
The `SunsetConfig` struct **shall** receive an official `// DORMANT` godoc marker noting that no hot path currently enforces it.

**REQ-MIG003-007**
Each new loader **shall** have unit tests under `internal/config/loader_test.go` covering: valid file load, missing file default, malformed file error.

### 5.2 Event-Driven Requirements

**REQ-MIG003-008**
**When** `LoadHarnessConfig` is called and the file is malformed, it **shall** return `HARNESS_CONFIG_PARSE_ERROR` with the YAML line number.

**REQ-MIG003-009**
**When** `LoadConstitutionConfig` is called and `forbidden_libraries` list is non-empty, it **shall** expose the list for runtime policy enforcement (consumed by SPEC-V3R2-EXT-004 framework optional hook).

**REQ-MIG003-010**
**When** `LoadContextConfig` runs, it **shall** parse token_budget and staleness_window_days from CLAUDE.md §16 spec for context search.

**REQ-MIG003-011**
**When** `LoadInterviewConfig` runs, it **shall** parse clarity_threshold and max_rounds fields consumed by SPEC-V3R2-WF-003 discovery mode.

### 5.3 State-Driven Requirements

**REQ-MIG003-012**
**While** the user's `.moai/config/sections/harness.yaml` is absent, the loader **shall** return a default HarnessConfig with level: `standard`.

**REQ-MIG003-013**
**While** a new YAML section is added to `.moai/config/sections/` without a Go loader, CI **shall** fail with `YAML_SECTION_NO_LOADER` (after v3.0 cutover).

### 5.4 Optional Requirements

**REQ-MIG003-014**
**Where** `design.yaml` is consumed by the GAN loop runtime (not just migrate_agency), the loader **shall** expose the sprint_contract, adaptation.phase_weights fields.

**REQ-MIG003-015**
**Where** a future SPEC activates `sunset.yaml` hot path, the struct **shall** be upgraded and the `// DORMANT` marker removed.

### 5.5 Complex Requirements (Unwanted Behavior / Composite)

**REQ-MIG003-016 (Unwanted Behavior)**
**If** any of the 5 newly-loaded sections have a struct field defined in Go but missing in the YAML template default, **then** `make build` **shall** fail with `CONFIG_STRUCT_YAML_MISMATCH`.

**REQ-MIG003-017 (Unwanted Behavior)**
**If** any loader hard-codes values (e.g., `const maxRounds = 4` overriding YAML), **then** code review **shall** reject with `LOADER_HARDCODE_VIOLATION`.

**REQ-MIG003-018 (Complex: State + Event)**
**While** v3.0.0 is being released, **when** `sunset.yaml` is evaluated, the system **shall** log `SUNSET_CONFIG_DORMANT_NOTICE` once per session to remind future maintainers.

---

## 6. Acceptance Criteria (수용 기준 요약)

- **AC-MIG003-01**: Given `internal/config/types.go` When inspected Then 5 new structs (`HarnessConfig`, `ConstitutionConfig`, `ContextConfig`, `InterviewConfig`, `DesignConfig`) exist (maps REQ-MIG003-001).
- **AC-MIG003-02**: Given valid `harness.yaml` When `LoadHarnessConfig` is called Then typed HarnessConfig is returned (maps REQ-MIG003-002).
- **AC-MIG003-03**: Given absent `harness.yaml` When `LoadHarnessConfig` is called Then default config with level=`standard` is returned (maps REQ-MIG003-004, REQ-MIG003-012).
- **AC-MIG003-04**: Given malformed `harness.yaml` When loader runs Then `HARNESS_CONFIG_PARSE_ERROR` with line number (maps REQ-MIG003-008).
- **AC-MIG003-05**: Given each new struct When godoc is inspected Then at least one runtime hot path is documented (maps REQ-MIG003-005).
- **AC-MIG003-06**: Given `SunsetConfig` struct When godoc inspected Then `// DORMANT` marker is present (maps REQ-MIG003-006).
- **AC-MIG003-07**: Given `go test ./internal/config/...` When executed Then valid/missing/malformed cases pass for all 5 loaders (maps REQ-MIG003-007).
- **AC-MIG003-08**: Given `ConstitutionConfig.ForbiddenLibraries` non-empty When runtime policy enforcement runs Then the list is exposed per REQ-MIG003-009.
- **AC-MIG003-09**: Given `ContextConfig` loaded When context search runs Then token_budget and staleness_window_days are honored (maps REQ-MIG003-010).
- **AC-MIG003-10**: Given `InterviewConfig` loaded When SPEC-V3R2-WF-003 discovery runs Then clarity_threshold and max_rounds are consumed (maps REQ-MIG003-011).
- **AC-MIG003-11**: Given `DesignConfig` loaded When GAN loop runs Then sprint_contract and adaptation.phase_weights are exposed (maps REQ-MIG003-014).
- **AC-MIG003-12**: Given a new `.moai/config/sections/foo.yaml` without loader When CI runs Then `YAML_SECTION_NO_LOADER` failure (maps REQ-MIG003-013).
- **AC-MIG003-13**: Given a hardcoded value overriding YAML When code review runs Then `LOADER_HARDCODE_VIOLATION` (maps REQ-MIG003-017).
- **AC-MIG003-14**: Given a Go struct field absent in YAML template default When `make build` runs Then `CONFIG_STRUCT_YAML_MISMATCH` failure (maps REQ-MIG003-016).
- **AC-MIG003-15**: Given v3.0.0 session When sunset.yaml is evaluated Then `SUNSET_CONFIG_DORMANT_NOTICE` logged once (maps REQ-MIG003-018).

---

## 7. Constraints (제약)

- 9-direct-dep 정책 준수 (기존 `yaml.v3` + stdlib만 사용).
- Sensible defaults 필수 — 파일 부재가 runtime panic을 일으키지 않도록.
- Hardcoded 금지 (CLAUDE.local.md §14 `internal/config/envkeys.go` 패턴 존중).
- `sunset.yaml`은 활성화 하지 않고 dormant 유지 (별도 SPEC에서 activation 권한).
- 각 loader는 thread-safe 읽기 보장 (Go `sync.RWMutex` 필요 시).

---

## 8. Risks & Mitigations (리스크 및 완화)

| 리스크 | 영향 | 완화 |
|---|---|---|
| YAML schema가 파일에 implicit 의존 (default missing) | 런타임 panic | sensible defaults 구현 필수 + 단위 테스트 |
| `harness.yaml` loader 도입이 기존 agent behavior 변화 | 회귀 | default = `standard`로 기존 상태 유지; 옵트인 변경 |
| `constitution.yaml` forbidden library 체크가 false positive | 사용자 빌드 차단 | 본 SPEC은 loader만; enforcement는 opt-in hook으로 별도 SPEC |
| 사용자가 `sunset.yaml`을 삭제 | loader panic | REQ-MIG003-006의 dormant marker + graceful path |
| 5개 loader 모두 async I/O 시도 | 디스크 대기 | 동기 loader + SessionStart에서 한 번만 호출 |

---

## 9. Dependencies (의존성)

### 9.1 Blocked by

- SPEC-V3R2-EXT-004: migration framework이 config 변경 migration step 실행.

### 9.2 Blocks

- SPEC-V3R2-MIG-001: v2→v3 migrator가 config loader completeness step 실행.
- SPEC-V3R2-WF-003: multi-mode router의 harness.yaml 의존성.

### 9.3 Related

- SPEC-V3R2-EXT-001 (memory): InterviewConfig와 memory staleness settings 간 관계.
- SPEC-V3R2-EXT-002 (Go loader): output-style loader가 본 SPEC loader pattern을 참조 가능.

---

## 10. Traceability (추적성)

- REQ 총 18개: Ubiquitous 7, Event-Driven 4, State-Driven 2, Optional 2, Complex 3.
- AC 총 15개, 모든 REQ에 최소 1개 AC 매핑 (100% 커버리지).
- Wave 2 소스 앵커: R6 §5.1/§5.2/§5.3 config audit.
- BC 영향: 없음 (loader 추가, 기존 behavior 보존; defaults 보장).
- 구현 경로 예상:
  - `internal/config/types.go` (5개 struct 추가)
  - `internal/config/harness.go`, `constitution.go`, `context.go`, `interview.go`, `design_runtime.go` (파일 분리)
  - `internal/config/loader.go` (5개 loader 함수 추가)
  - `internal/config/loader_test.go` (확장)
  - 문서 업데이트: `.claude/rules/moai/core/settings-management.md`
- **File:line anchors** (per D5 traceability requirement):
  - `docs/design/major-v3-master.md:L1084` (§11.8 MIG-003 definition)
  - `docs/design/major-v3-master.md:L972` (§8 BC-V3R2-013 — config loaders)
  - `docs/design/major-v3-master.md:L995` (§9 Phase 8 Migration Tool + Docs)
  - `.moai/design/v3-redesign/synthesis/problem-catalog.md` (P-H06, P-H07, P-H20)

---

End of SPEC.
