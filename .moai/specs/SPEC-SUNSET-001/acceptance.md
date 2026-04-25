---
spec_id: SPEC-SUNSET-001
version: 1.0.0
status: backfilled
created_at: 2026-04-24
author: manager-spec (backfill)
backfill_reason: |
  REVERSE-DOC (PARTIAL): 본 acceptance.md 는 SPEC 작성 이후 부분 구현이 완료된
  상태에서 SDD 아티팩트 공백을 메우기 위해 역공학 방식으로 생성되었다.
  plan-auditor 2026-04-24 감사에서 `acceptance.md` 미존재가 지적되어 backfill됨.

  PARTIAL 상태 (spec.md 명시: "advisory only for v1"):
  - [완료] internal/template/templates/.moai/config/sections/sunset.yaml — 설정 파일
  - [완료] internal/config/types.go:244-256 — SunsetConfig / SunsetCondition Go 구조체
  - [완료] internal/config/types.go:314 — "sunset" section name 등록
  - [미완료] Per-gate pass/fail 추적 (REQ-SUN-003, SPEC-OBSERVE-001 의존)
  - [미완료] `.moai/reports/sunset-recommendations.md` 로그 작성 (REQ-SUN-004)
  - [미완료] Advisor 모드 전환 런타임 로직 (REQ-SUN-005)

  따라서 AC-001, AC-002 는 구현 근거로 자동 검증 가능,
  AC-003, AC-004, AC-005 는 PARTIAL — pending implementation 으로 표기.
---

# Acceptance Criteria — SPEC-SUNSET-001

> Build to Delete Framework 의 실제 구현 상태로부터 역도출된 검증 기준이다.
> spec.md 에 "Non-Goals: Automatic gate removal (advisory only for v1)" 이 명시되어
> 있어, v1 범위는 (a) 설정 스키마, (b) Go 구조체 파싱에 한정된다. 추적/권고 로그는
> SPEC-OBSERVE-001 에 의존하는 미래 기능이다.

## Traceability Matrix

| REQ ID | AC ID | 검증 수단 (파일:라인) |
|---|---|---|
| REQ-SUN-001 (sunset.yaml 정의) | AC-001 | `internal/template/templates/.moai/config/sections/sunset.yaml` |
| REQ-SUN-002 (각 조건의 5개 필드) | AC-002 | `internal/config/types.go:249-256` SunsetCondition struct |
| REQ-SUN-003 (per-gate pass/fail 추적) | AC-003 | **PARTIAL — pending** (SPEC-OBSERVE-001 의존) |
| REQ-SUN-004 (recommendations.md 로그) | AC-004 | **PARTIAL — pending** (관련 핸들러 미구현) |
| REQ-SUN-005 (advisory only) | AC-005 | **PARTIAL — design 의도, 자동 전환 로직 미구현** |

---

## AC-001 — `sunset.yaml` 설정 파일 정의

**Given** `moai init` 또는 `moai update` 가 템플릿을 배포한 상태
**When** `.moai/config/sections/sunset.yaml` 파일을 검사할 때
**Then** 파일은 존재해야 하며, 최상위 키 `sunset` 아래 `enabled: true` 와 `conditions:` 배열을 포함해야 한다. `conditions` 배열의 각 항목은 `gate`, `metric`, `threshold`, `action`, `description` 5개 필드를 가져야 한다.

**Verification**:
- `ls internal/template/templates/.moai/config/sections/sunset.yaml` → 파일 존재 (820B)
- `yq '.sunset.enabled' sunset.yaml` → `true`
- `yq '.sunset.conditions | length' sunset.yaml` → 3 (vet, lint, test)
- 실제 내용 (sunset.yaml L6-L23):
  - `enabled: true`
  - 3 conditions: gate=vet (threshold 50), gate=lint (threshold 30), gate=test (threshold 20)
  - 모두 `action: "advisor"`, `metric: "consecutive_passes"`

---

## AC-002 — SunsetCondition Go 구조체 정의

**Given** `moai` 바이너리가 설정 파일을 로드하는 상태
**When** `internal/config/types.go` 의 `SunsetCondition` 구조체가 YAML을 unmarshal 할 때
**Then** 구조체는 `Gate`, `Metric`, `Threshold`, `Action`, `Description` 5개 필드를 가져야 하며, 각 필드는 YAML 태그로 소문자 snake_case 와 매핑되어야 한다. `SunsetConfig` 는 `Enabled` (bool) 와 `Conditions` (`[]SunsetCondition`) 2개 필드를 가진다.

**Verification**:
- `internal/config/types.go:240-256`:
  ```go
  type SunsetConfig struct {
      Enabled    bool              `yaml:"enabled"`
      Conditions []SunsetCondition `yaml:"conditions"`
  }

  type SunsetCondition struct {
      Gate        string `yaml:"gate"`
      Metric      string `yaml:"metric"`
      Threshold   int    `yaml:"threshold"`
      Action      string `yaml:"action"`
      Description string `yaml:"description"`
  }
  ```
- `internal/config/types.go:28` = `Sunset SunsetConfig \`yaml:"sunset"\`` — 상위 config에 등록됨
- `internal/config/types.go:314` — "sunset" section 이름이 유효 섹션 목록에 포함 (`"pricing", "ralph", "workflow", "state", "statusline", "gate", "sunset",`)

---

## AC-003 — Per-gate pass/fail 추적 (PARTIAL)

**Given** 품질 게이트 (go vet, golangci-lint, go test) 실행이 세션 요약에 기록되는 상태
**When** 여러 세션에 걸쳐 게이트가 반복 실행될 때
**Then** 시스템은 각 게이트별 연속 통과 횟수(consecutive_passes)를 세션 요약 또는 영구 상태 파일에 누적 추적해야 한다.

**Status**: **PARTIAL — pending implementation**
**Depends On**: SPEC-OBSERVE-001 (trace system)
**Verification Gap**:
- `grep -rn "consecutive_passes\|sunset.*track\|gate.*pass.*count" internal/ --include="*.go"` → 관련 추적 로직 부재
- 현재 구현은 설정만 파싱 가능하며, pass/fail 카운터와 세션 요약 통합이 미구현
- spec.md L27-L28: "The trace system (SPEC-OBSERVE-001) SHALL track per-gate pass/fail counts in session summaries" — 트레이스 시스템에 명시적 의존

**Remediation**: SPEC-OBSERVE-001 구현 완료 후 per-gate 카운터 통합.

---

## AC-004 — Sunset 권고 로그 작성 (PARTIAL)

**Given** 한 게이트의 `consecutive_passes` 카운터가 `threshold` 에 도달한 상태
**When** 다음 게이트 평가가 수행될 때
**Then** 시스템은 `.moai/reports/sunset-recommendations.md` 파일에 권고 엔트리를 기록해야 한다. 엔트리는 gate 이름, 달성된 threshold, 권고 action (advisor), 타임스탬프를 포함한다.

**Status**: **PARTIAL — pending implementation**
**Verification Gap**:
- `grep -rn "sunset-recommendations" internal/ pkg/` → 매치 없음
- `.moai/reports/` 디렉터리에 sunset 관련 로그 파일 부재
- 권고 생성 트리거와 파일 쓰기 로직 미구현

**Remediation**: AC-003 구현 후 임계치 도달 시 `.moai/reports/sunset-recommendations.md` append 로직 추가.

---

## AC-005 — Advisory-only 보장 (자동 삭제 금지)

**Given** 어떤 sunset 조건이 threshold 에 도달한 상태
**When** 시스템이 해당 조건을 평가할 때
**Then** 게이트는 **자동으로 제거되거나 우회되지 않아야** 한다. 권고는 로그 파일에만 기록되며, 실제 gate 제거나 advisor 모드 전환은 사람의 승인(PR 머지 등) 이 필요하다.

**Status**: **PARTIAL — design intent 충족, 자동 전환 로직 자체가 미구현이므로 간접적으로 충족**
**Verification**:
- sunset.yaml L12, L17, L22 모든 조건이 `action: "advisor"` 로 설정됨 (`remove` action 없음)
- spec.md L31: "Sunset actions SHALL be advisory only — actual gate removal requires human approval"
- spec.md L64: "Non-Goals: Automatic gate removal (advisory only for v1)"
- 현재 자동 게이트 제거 코드 경로가 **존재하지 않으므로** advisory-only 제약이 자동으로 충족됨
- 단, 이는 "구현 부재에 의한 충족" 이지 "적극적 방어 로직에 의한 충족" 이 아님 → 후속 구현 시 `action: "remove"` 분기 추가 금지 정책 필요

**Verification Gap**:
- 자동 전환 로직이 추가될 때 "human approval" 게이트 (AskUserQuestion 또는 PR 승인 체크) 가 함께 구현되어야 함
- 현재는 방어적 검사 없이 advisory-only 가 "구현 부재" 로 자연 보장됨

**Remediation**: 후속 구현 시 SunsetConfig 에 `require_human_approval: true` 필드 추가 및 자동 전환 전 체크 삽입 권장.

---

## Edge Cases and Defensive Behaviors

- **EC-01**: `sunset.enabled: false` 일 때 → `conditions` 배열이 무시되어야 함. 현재는 파싱 후 사용되는 코드 경로 부재로 자연 충족. 후속 구현 시 런타임 분기 필요.
- **EC-02**: `conditions` 배열이 빈 `[]` 일 때 → 아무 게이트도 추적하지 않음. Go 구조체는 빈 slice 를 허용하므로 파싱 에러 없음.
- **EC-03**: `threshold` 가 0 또는 음수일 때 → spec.md 에 validation 규칙 없음. 후속 구현 시 `threshold > 0` 검증 권장.
- **EC-04**: 동일 `gate` 에 대해 여러 조건이 정의될 때 → spec.md 에 충돌 해결 규칙 없음. 후속 구현 시 duplicate detection 권장.
- **EC-05**: `action` 필드 값이 예상되지 않은 문자열 (예: "remove") 일 때 → 현재는 validation 없음. spec.md 는 "advisor/remove" 를 Non-Goals 로 명시하나 `remove` 는 v1에서 지원하지 않음. 후속 구현 시 enum 검증 필요.

---

## Definition of Done

| 항목 | 상태 | 근거 |
|---|---|---|
| sunset.yaml 설정 파일 정의 | [x] DONE | `internal/template/templates/.moai/config/sections/sunset.yaml` 23 lines |
| SunsetConfig / SunsetCondition Go 구조체 | [x] DONE | `internal/config/types.go:240-256` |
| Section name 등록 | [x] DONE | `internal/config/types.go:314` valid sections list |
| 3개 기본 조건 (vet/lint/test) | [x] DONE | sunset.yaml L9-L23 |
| Per-gate pass/fail 추적 | [ ] PARTIAL | SPEC-OBSERVE-001 의존, 미구현 |
| Sunset recommendations 로그 작성 | [ ] PARTIAL | 트리거 + 파일 쓰기 미구현 |
| Advisory-only 방어 로직 | [ ] PARTIAL | 현재는 구현 부재로 자연 충족 |
| Gate 이름 언어 중립 (go_vet → vet) | [x] DONE | sunset.yaml L2-L3 주석: "Gate names are language-agnostic" |

---

## Quality Gate Alignment

| Gate | Criterion | Evidence |
|---|---|---|
| Tested | sunset.yaml YAML 파싱 가능 + Go struct unmarshal | `yq eval '.' sunset.yaml`, Go tests (구조체 기반) |
| Readable | 각 조건의 `description` 필드로 의도 명시 | sunset.yaml L13, L18, L23 |
| Unified | 3개 조건 모두 동일 5-필드 스키마 | AC-001, AC-002 |
| Secured | Advisory-only 기본값 (자동 제거 불가) | AC-005 (design intent) |
| Trackable | spec.md 명시적 Non-Goals + SPEC-ID 레퍼런스 | sunset.yaml L1-L4 주석 헤더 |

---

## Reverse-Doc Divergence Notes

1. **스펙과 구현이 범위 일치**: spec.md Non-Goals 섹션이 v1 을 "advisory only" 로 명시적으로 한정. 현재 구현은 **설정 스키마 정의** 단계까지 완료되어 v1 스코프와 일치.
2. **런타임 추적/권고 기능은 의도적 후속 작업**: AC-003/004/005 는 spec.md 본문에 명시되지만 v1 범위에 포함되지 않음. 명시적 후속 과제.
3. **의존성 체인**: REQ-SUN-003 이 "The trace system (SPEC-OBSERVE-001) SHALL track..." 로 SPEC-OBSERVE-001 의 구현에 강하게 의존. SPEC-OBSERVE-001 없이는 REQ-SUN-003~005 가 구현 불가.
4. **언어 중립성 보강**: sunset.yaml 헤더 주석 (L2-L3) 이 "Gate names are language-agnostic; the runtime maps them to project-specific tools." 로 명시 — spec.md 의 `go_vet`, `golangci_lint` 예시는 구현 시 `vet`, `lint` 로 일반화되었음. SPEC 과 구현 사이의 선량한 개선 (template language neutrality 원칙 준수).
5. **권고**: 후속 SPEC 에서 (a) SPEC-OBSERVE-001 완료 대기, (b) per-gate 카운터 통합, (c) advisor mode 런타임 효과 정의 (어떻게 "advisor 모드" 가 동작하는가?) 를 명확히 할 것.
