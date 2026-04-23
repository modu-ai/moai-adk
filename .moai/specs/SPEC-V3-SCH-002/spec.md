---
id: SPEC-V3-SCH-002
title: Settings 3-tier Source Layering (user / project / local)
version: 0.1.0
status: draft
created: 2026-04-22
updated: 2026-04-22
author: manager-spec
priority: High
phase: "Phase 1 — Foundation"
module: "internal/config/sources/"
dependencies:
  - SPEC-V3-SCH-001
related_gap: [138, 164]
related_theme: "Theme 3 — Schema & Validation Layer"
breaking: true
bc_id: [BC-002, BC-003]
lifecycle: spec-anchored
tags: "v3, schema, settings, source-layering, merge, foundation, P1"
---

# SPEC-V3-SCH-002: Settings 3-tier Source Layering

## HISTORY

- 2026-04-22 v0.1.0: 최초 작성. master-v3 §3.3 Design approach step 5-6, gap-matrix #138, #164 (scope-reduced to 3-tier per §9 open question #4), W1.5 §2.2/§9.2 근거.

---

## 1. Goal (목적)

Claude Code는 6-tier settings precedence를 사용한다 (W1.5 §9.2): `userSettings → projectSettings → localSettings → flagSettings → policySettings → managedSettings`. moai-adk-go v2.12는 단일 `.moai/config/sections/*.yaml` 티어만 보유해 (gap-matrix #164), 사용자가 개인 환경(user) vs 팀 공유 설정(project) vs 머신별 재정의(local)를 분리할 수 없다.

본 SPEC은 master-v3 §9 open question #4의 권장 default에 따라 **3-tier scope** (user/project/local)를 도입한다. Enterprise policy/managed tier는 v3.2로 지연. 각 tier는 SPEC-V3-SCH-001에서 정의된 스키마로 검증되며, deep-merge 의미론 (map: key-wise / array: replace / scalar: override)으로 병합된다. 마지막 write-wins 정책은 `local > project > user`.

### 1.1 배경

W1.5 §9.2 gap analysis:
- CC의 `--setting-sources user,project,local` 필터링 플래그 (gap-matrix #138)
- moai는 `.moai/config/sections/` 단일 경로만 인식
- Multi-machine 개발자는 `.moai/config/sections/` 전체를 dotfiles로 symlink해야 하는 고통 (feedback 근거)

gap-matrix #164 (High): 6-source layering → 3-tier 축소 도입. v3.2에서 flag/policy/managed 추가 여지 보존.

### 1.2 비목표 (Non-Goals)

- **6-tier 풀 파리티 금지** (§9 open question #4): v3.0은 3-tier. policy/flag/managed는 v3.2 이후.
- **Plugin tier 금지**: plugin scope은 SPEC-V3-PLG-001의 install scope와 분리. 본 SPEC은 config 파일만 다룸.
- **Hook settings의 별도 layering 금지**: 본 SPEC의 결과물로 hook declaration 역시 동일 3-tier로 병합 (SPEC-V3-HOOKS-003에서 실제 dedup 구현).
- **원격 remote-managed settings 금지**: Anthropic-only 기능으로 적용 불가 (master-v3 §10 T4-POLICY-01).
- **Interactive edit UI 금지**: v3.2 이후.

---

## 2. Scope (범위)

### 2.1 In Scope

- `internal/config/sources/` 신규 패키지: userSettings, projectSettings, localSettings 로더
- 3-tier 정의:
  - **user**: `~/.moai/config/sections/*.yaml` (개인 global)
  - **project**: `.moai/config/sections/*.yaml` (팀 공유, checked-in)
  - **local**: `.moai/config/sections/*.local.yaml` (머신별, gitignored)
- Deep-merge semantics:
  - Map type: 키 단위 병합
  - Array type: replace (upper tier가 완전히 교체)
  - Scalar: override (last-wins = local)
- Precedence: `local > project > user` (last-wins에서 last가 local)
- `.gitignore`에 `*.local.yaml` 자동 추가 (template-first)
- `moai doctor config --sources`: 각 설정 키의 effective value + source 표시
- `--setting-sources <user,project,local>` CLI 플래그: 특정 tier만 로드 (디버깅 용)
- Migration: v2 `.moai/config/sections/*.yaml` 값은 project tier로 자동 배치 (무결성 유지)

### 2.2 Out of Scope (Exclusions — What NOT to Build)

- 원격 managed/policy tier (v3.2 이후)
- Flag settings tier (`--settings <file-or-json>` inline) — v3.2 이후, gap-matrix #139
- Enterprise 상위 tier priority inversion — 불필요
- Hook-specific layering 구현 — SPEC-V3-HOOKS-003으로 분리
- CLI `moai config set/get` 쓰기 도구 — v3.2 이후
- Array의 partial merge (concat) — replace 시맨틱 유지로 일관성
- Windows registry / Linux freedesktop config tier — OS-portable YAML만 사용

---

## 3. Environment (환경)

- 런타임: Go 1.26, moai-adk-go v3.0.0-alpha.1+
- 의존성: SPEC-V3-SCH-001 (schema package + registry)
- YAML 파서: gopkg.in/yaml.v3 (기존)
- 영향 디렉터리:
  - `~/.moai/config/sections/` (신규 user tier 디렉터리, 사용자 수동 또는 `moai init --user-config` 생성)
  - `.moai/config/sections/` (기존 project tier)
  - `.moai/config/sections/*.local.yaml` (신규 local tier, gitignored)
  - `internal/config/sources/` (신규 로더 + merger)
  - `internal/template/templates/.gitignore` (local.yaml 패턴 추가)
- 대상 OS: macOS / Linux / Windows 동등. `~` resolves via `os.UserHomeDir()`.

---

## 4. Assumptions (가정)

- A-001 (High): 사용자의 `~/.moai/config/sections/` 디렉터리는 존재하지 않아도 에러가 아니다. 미존재 시 user tier는 빈 맵.
- A-002 (High): Project tier는 v2 설정의 정식 후계자다. v2 사용자의 기존 파일은 수정 없이 project tier로 인식된다.
- A-003 (Medium): Local tier 파일은 머신별 재정의(예: user.name, 로컬 포트 번호) 용도이며, Git commit 대상이 아니다. `.gitignore` 자동 추가로 보호.
- A-004 (High): Array replace 시맨틱은 Claude Code의 `policySettings` merge 패턴 (W1.5 §2.2)과 일치한다.
- A-005 (Medium): `moai update`가 project tier template을 갱신해도 local tier는 건드리지 않는다 (CLAUDE.local.md §2.1 Protected Directories).
- A-006 (Low): 사용자가 동일 키를 project와 local 모두에 정의할 경우, local wins. 이 혼동을 `moai doctor config --sources`로 해소.

---

## 5. Requirements (EARS 요구사항)

### 5.1 Ubiquitous Requirements

**REQ-SCH-002-001 (Ubiquitous)**
시스템은 **항상** 3-tier (user/project/local)로 `.moai/config/sections/*.yaml`을 로드하고, 각 tier는 SPEC-V3-SCH-001의 스키마로 검증된다.

**REQ-SCH-002-002 (Ubiquitous)**
시스템은 **항상** precedence `local > project > user`의 last-wins 정책으로 최종 effective config를 계산한다.

**REQ-SCH-002-003 (Ubiquitous)**
시스템은 **항상** deep-merge 시 다음 규칙을 따른다: Map(키 단위 병합), Array(replace), Scalar(override).

**REQ-SCH-002-004 (Ubiquitous)**
시스템은 **항상** `.gitignore`에 `.moai/config/sections/*.local.yaml` 패턴이 포함되도록 보증한다 (template-first).

**REQ-SCH-002-005 (Ubiquitous)**
시스템은 **항상** user tier 디렉터리(`~/.moai/config/sections/`)가 존재하지 않을 경우 빈 맵으로 간주하고 에러를 반환하지 않는다.

### 5.2 Event-Driven Requirements

**REQ-SCH-002-010 (Event-Driven)**
**When** `ConfigManager.Load()`가 호출되면, the 시스템 **shall** user → project → local 순으로 YAML 파일을 로드하고, 각 section별로 deep-merge를 수행해 effective config를 반환한다.

**REQ-SCH-002-011 (Event-Driven)**
**When** 사용자가 `moai doctor config --sources`를 실행하면, the 시스템 **shall** 각 설정 키에 대해 effective value와 해당 값이 어느 tier에서 왔는지를 트리 구조로 표시한다.

**REQ-SCH-002-012 (Event-Driven)**
**When** 사용자가 `--setting-sources <list>` 플래그(e.g., `user,project`)로 CLI를 실행하면, the 시스템 **shall** 지정된 tier만 로드해 디버깅 목적의 effective config를 계산한다.

**REQ-SCH-002-013 (Event-Driven)**
**When** `moai init`이 실행되면, the 시스템 **shall** user tier 디렉터리 존재 여부를 확인하고, 부재 시 사용자에게 (AskUserQuestion 위임) `~/.moai/config/sections/` 생성 여부를 묻는다.

**REQ-SCH-002-014 (Event-Driven)**
**When** `moai update`가 실행되면, the 시스템 **shall** project tier template만 갱신하고 user/local tier 파일은 변경하지 않는다 (Protected Directories 원칙).

### 5.3 State-Driven Requirements

**REQ-SCH-002-020 (State-Driven)**
**While** local tier 파일이 존재하지만 parse 실패 상태인 동안, the 시스템 **shall** dual-parse fallback을 적용하고 해당 tier를 "degraded"로 마킹해 `moai doctor config`에서 가시화한다.

**REQ-SCH-002-021 (State-Driven)**
**While** `--setting-sources` 플래그가 활성화된 동안, the 시스템 **shall** 해당 tier 외의 설정은 무시하며, 경고 메시지로 "Only sources [X] loaded; some config may be missing"을 표시한다.

**REQ-SCH-002-022 (State-Driven)**
**While** user와 project tier가 동일 키에 대해 충돌하는 경우, the 시스템 **shall** project tier 값을 사용하되 `moai doctor config --sources` 출력에서 해당 충돌을 명시한다.

### 5.4 Optional Requirements

**REQ-SCH-002-030 (Optional)**
**Where** 사용자가 dotfiles 관리 시스템을 사용하는 경우, the 시스템 **shall** user tier 디렉터리를 symlink로 대체해도 동작한다 (filepath.EvalSymlinks 기반).

**REQ-SCH-002-031 (Optional)**
**Where** 사용자가 `moai doctor config --sources --json`을 사용하면, the 시스템 **shall** 기계 판독 가능한 JSON으로 effective config + source 매핑을 출력한다.

**REQ-SCH-002-032 (Optional)**
**Where** 사용자가 `MOAI_CONFIG_DISABLE_LOCAL=1`을 설정하면, the 시스템 **shall** local tier를 완전히 비활성화한다 (CI 환경 테스트 용).

### 5.5 Unwanted Behavior (Must Not)

**REQ-SCH-002-040 (Unwanted)**
시스템은 array 타입에 대해 element-wise merge를 **수행하지 않아야 한다**. 상위 tier가 array 키를 정의하면 하위 tier의 array는 완전 교체된다.

**REQ-SCH-002-041 (Unwanted)**
시스템은 local tier YAML을 `git commit` 대상에 포함하도록 **허용하지 않아야 한다**. `.gitignore` 보장 실패 시 `moai doctor config`가 warning을 표시한다.

**REQ-SCH-002-042 (Unwanted)**
시스템은 user tier 값을 project tier보다 **우선 적용하지 않아야 한다**. Precedence 반전 금지.

**REQ-SCH-002-043 (Unwanted)**
시스템은 remote-managed/policy tier를 v3.0에서 **구현하지 않아야 한다**. 해당 tier 사용 시도는 "Not implemented in v3.0; planned for v3.2" 메시지와 함께 거부.

### 5.6 Complex Requirements

**REQ-SCH-002-050 (Complex)**
**While** `moai init` 또는 `moai update` 수행 중이며, **when** 사용자가 기존 v2 설정을 가진 상태이면, the 시스템 **shall** (a) 기존 `.moai/config/sections/*.yaml`을 project tier로 유지, (b) user tier 디렉터리 생성 여부 확인(AskUserQuestion), (c) `*.local.yaml` 패턴을 `.gitignore`에 추가, (d) `moai doctor config --sources`를 실행해 3-tier 로딩이 정상 동작함을 사용자에게 표시한다.

---

## 6. Acceptance Criteria (수용 기준 요약)

**AC-SCH-002-01**: user (`~/.moai/config/sections/quality.yaml`), project (`.moai/config/sections/quality.yaml`), local (`.moai/config/sections/quality.local.yaml`) 3 파일에서 서로 다른 `test_coverage_target` 값 (80, 85, 90)을 정의하면, `moai doctor config --sources`가 effective=90, source=local을 출력한다.

**AC-SCH-002-02**: Array key `lsp_quality_gates.plan.forbidden_imports: [a, b]` in project, `[c]` in local → effective = `[c]` (replace semantic 검증).

**AC-SCH-002-03**: `.gitignore`에 `.moai/config/sections/*.local.yaml` 패턴이 존재함을 `git check-ignore .moai/config/sections/quality.local.yaml`로 확인.

**AC-SCH-002-04**: v2.12 사용자 corpus 10개 프로젝트 `moai update` 후 기존 `.moai/config/sections/*.yaml`이 project tier로 유지되며, effective config가 v2와 byte-equivalent.

**AC-SCH-002-05**: `--setting-sources user,project` 플래그 시 local tier 파일이 존재해도 무시되고 effective=project value. Warning 메시지 1회 표시.

**AC-SCH-002-06**: Local tier YAML이 schema 위반 (typo)이면 dual-parse fallback으로 경고 표시, project tier 기반 effective config로 서비스 지속.

**AC-SCH-002-07**: `moai doctor config --sources --json` 출력이 각 키에 대해 `{key, effective_value, source, tier_values: {user, project, local}}` 구조로 검증 가능.

**AC-SCH-002-08**: Symlink된 `~/.moai` (dotfiles 관리)에서 user tier 로드 성공.

---

## 7. Constraints (제약)

- **[HARD] SPEC-V3-SCH-001 완성 선행**: 스키마 정의 없이는 tier 검증 불가.
- **[HARD] Protected Directories 원칙 (CLAUDE.local.md §2)**: `moai update` 중 user/local tier는 불변.
- **[HARD] Array replace 시맨틱**: partial merge 금지. CC와 일관 (W1.5 §2.2).
- **[HARD] v3.2 이후 scope 확장 금지**: v3.0은 3-tier로 고정. 4+ tier는 master-v3 open question 재검토 필요.
- User tier 생성은 `moai init`에서 opt-in (AskUserQuestion). 자동 생성 금지 (사용자 surprise 방지).

---

## 8. Risks & Mitigations (리스크 및 완화)

| Risk ID | 설명 | 확률 | 영향 | 완화 |
|---------|------|------|------|------|
| R-SCH-002-01 | 사용자가 local tier를 commit해 민감 정보 유출 | Low | High | `.gitignore` template-first + `moai doctor config`에서 gitignore 누락 경고 |
| R-SCH-002-02 | Array replace vs concat 혼동 | Medium | Medium | docs-site 4-locale 문서화 + `moai doctor config --sources`에 시맨틱 힌트 |
| R-SCH-002-03 | User tier 파일 대량 생성 후 GC 없음 | Low | Low | `moai doctor config` 에서 user tier orphan 감지 (scheme에 없는 키 경고) |
| R-SCH-002-04 | v2 단일 tier → v3 3-tier 전환 시 기존 설정 값 손실 | Low | Critical | Migration M01 (SPEC-V3-MIG-002)에서 project tier 그대로 유지, backup 선행 |
| R-SCH-002-05 | Precedence 혼동으로 "왜 설정이 무시되나" 디버깅 어려움 | Medium | Medium | `--sources` CLI + `--json` 출력으로 effective path 시각화 |
| R-SCH-002-06 | 3-tier → 6-tier 향후 확장 시 breaking change | Medium | Low | v3.0 SPEC에 4+ tier 확장 여지 명시; 추가 tier는 higher precedence로만 삽입 (기존 local 이하 불변) |

---

## 9. Dependencies (의존성)

### 9.1 Blocked by

- **SPEC-V3-SCH-001** (Formal Config Schema Framework) — tier-aware validation 기반

### 9.2 Blocks

- **SPEC-V3-HOOKS-003** (Hook source precedence 3-tier) — 본 SPEC의 merge 의미론을 hook declaration에 재사용
- **SPEC-V3-MIG-002** (v2→v3 migration content) — M01은 project tier, M02~M05는 다양 tier
- **SPEC-V3-PLG-001** (Plugin manifest) — plugin install scope은 본 SPEC tier와 유사 구조 (user/project/local)
- **SPEC-V3-CLI-001** (CLI subcommand restructure) — `--setting-sources` 플래그 + `moai doctor config --sources` 서브커맨드 정의
- **SPEC-V3-MEM-001** (MEMORY.md 4-type) — memory.yaml 설정 tier 인식

### 9.3 Related

- gap-matrix #138, #164
- W1.5 §2.2 (`--setting-sources` 플래그), §9.2 (gap analysis)
- CLAUDE.local.md §2 (Protected Directories)
- master-v3 §9 open question #4 (3-tier scope 결정)

---

## 10. Traceability (추적성)

| Requirement | Implementation | Verification |
|-------------|----------------|--------------|
| REQ-SCH-002-001, 002, 003 | `internal/config/sources/merger.go` | `TestThreeTierPrecedence`, `TestDeepMergeMapArray` |
| REQ-SCH-002-004 | `internal/template/templates/.gitignore` | `TestGitignoreContainsLocalYAML` |
| REQ-SCH-002-005 | `internal/config/sources/user_loader.go` | `TestUserTierMissingDir` |
| REQ-SCH-002-010 | `internal/config/manager.go Load()` | `TestLoadThreeTiers` |
| REQ-SCH-002-011, 031 | `internal/cli/doctor_config.go --sources` | `TestDoctorConfigSources`, `TestDoctorConfigSourcesJSON` |
| REQ-SCH-002-012, 021 | `internal/cli/root.go --setting-sources` | `TestSettingSourcesFilter` |
| REQ-SCH-002-013 | `internal/cli/init.go` AskUserQuestion delegation | `TestInitUserTierPrompt` |
| REQ-SCH-002-014 | `internal/cli/update.go` + manifest.go ProtectedPaths | `TestUpdatePreservesUserLocalTier` |
| REQ-SCH-002-020 | `internal/config/sources/local_loader.go` dual-parse | `TestLocalTierDegraded` |
| REQ-SCH-002-022 | `internal/cli/doctor_config.go` conflict detection | `TestUserProjectConflictVisible` |
| REQ-SCH-002-030 | `internal/config/sources/user_loader.go` EvalSymlinks | `TestUserTierSymlink` |
| REQ-SCH-002-032 | Env var check in `internal/config/sources/local_loader.go` | `TestDisableLocalEnvVar` |
| REQ-SCH-002-040, 041, 042, 043 | Code review + unit test | `TestArrayReplaceNotConcat`, `TestUnknownTierRejected` |
| REQ-SCH-002-050 | `internal/cli/init.go` + `internal/cli/update.go` flow | `TestV2ToV3UpgradeFlow` |
| AC-SCH-002-01 ~ 08 | 해당 REQ 커버리지 전체 | `go test ./internal/config/sources/...` + 10-project corpus |
| BC-002 (Breaking) | v2 단일 tier → v3 3-tier | master-v3 Breaking Changes Catalog §4 |

---

End of SPEC-V3-SCH-002.
