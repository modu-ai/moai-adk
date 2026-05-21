---
id: SPEC-V3R5-STATUSLINE-PROFILE-WIZARD-001
title: "profile setup wizard에 statusline preset/segments 입력 UI 추가"
version: "0.2.0"
status: implemented
created: 2026-05-22
updated: 2026-05-22
author: manager-spec
priority: P2
phase: "v2.20.0-rc1"
module: "internal/cli"
lifecycle: spec-anchored
tags: "statusline, profile, wizard, ui, huh, tier-s, quick-win"
tier: S
---

# SPEC-V3R5-STATUSLINE-PROFILE-WIZARD-001 — profile setup wizard에 statusline preset/segments 입력 UI 추가

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-05-22 | manager-spec | 초안 작성. `ProfilePreferences.StatuslinePreset` + `StatuslineSegments` 필드(`internal/profile/preferences.go:46-47`)는 이미 존재하며 `syncStatusline`(`internal/profile/sync.go:95-145`)이 `.moai/config/sections/statusline.yaml`로 sync 하는 로직도 구현되어 있으나, `profile_setup.go:282-300` 의 huh.NewForm 4번째 Display group은 Mode + Theme만 입력받아 사용자가 preset/segments를 wizard에서 입력할 수 없음. Quick Win: backend 인프라 완비, frontend wizard input UI만 누락. Tier S 분류 근거: 3 source 파일 (`profile_setup.go` + `profile_setup_translations.go` + 신규/기존 test 파일) + LOC delta 추정 ~200-280 < 300 (plan.md §3.1 정밀 산정 반영, plan-auditor S2 fix). |
| 0.2.0 | 2026-05-22 | orchestrator-direct | run-phase 완료 (manager-develop 위임 시 `WorktreeCreate hook returned a path that is not a directory: {}` 반복 실패 — Claude Code runtime autonomous L1 isolation regression, cleanupBogusRootDir lesson 일치, CLAUDE.md §16 "위임 대상 부재" 조항으로 orchestrator-direct 전환). M1 translations 23 키 × 4 locale + helper 추가 / M2 Display group에 Preset Select + 별도 Segments MultiSelect group with WithHideFunc("custom") / M3 2 신규 테스트 추가 (TestNormalizeStatuslinePreset 7 cases + TestStatuslineAllSegments_CardinalityAndOrder + TestProfileSetupTranslations_PresetSegments 4 locale × 23 cells = 92 cells) / M4 7/7 ACs PASS (AC-SPW-001 line 352 / AC-SPW-002 line 376 NewMultiSelect + 50 NewOption / AC-SPW-003 PASS / AC-SPW-004 PASS / AC-SPW-005 internal/cli/... PASS / AC-SPW-006 darwin+windows+linux exit 0 / AC-SPW-007 git diff 02a0a8304 0 lines preserved). plan-auditor S4 absorbed (nil-map default = 15-segment all-true). status `draft → implemented`. Late-Branch policy main 직진 (push deferred to sync-phase per REQ-LB-005). |

## 1. Background

### 1.1 출처

본 SPEC은 statusline preset/segments 설정의 **frontend gap**을 해소한다. 사용자가 `moai profile setup` wizard 실행 시 다음 두 옵션이 노출되지 않아 CLI 흐름만으로는 preset 변경이 불가능하다:

1. `statusline.preset` (full / compact / minimal / custom) — 미노출
2. `statusline.segments` (15개 segment 토글 맵, custom preset 시만 의미) — 미노출

### 1.2 다루는 결함

#### 1.2.1 Backend는 완비, Frontend만 누락 (verifiable evidence)

`ProfilePreferences` struct에 두 필드가 이미 존재하며 sync 로직도 작동한다:

| 검증 항목 | 위치 | 상태 |
|-----------|------|------|
| `StatuslinePreset string` 필드 정의 | `internal/profile/preferences.go:46` | EXISTS |
| `StatuslineSegments map[string]bool` 필드 정의 | `internal/profile/preferences.go:47` | EXISTS |
| `SyncToProjectConfig` 두 필드 sync 분기 | `internal/profile/sync.go:69-77` | EXISTS |
| `syncStatusline` preset/segments merge 로직 | `internal/profile/sync.go:95-145` | EXISTS |
| `defaultStatuslineSegments` 11-segment map | `internal/profile/sync.go:148-162` | EXISTS |
| `.moai/config/sections/statusline.yaml` 15-segment 정의 | (statusline.yaml) | EXISTS |
| wizard huh.NewForm Display group preset/segments input | `internal/cli/profile_setup.go:282-300` | **MISSING** |
| profileSetupText 번역 키 (Preset*, Segment*) | `internal/cli/profile_setup_translations.go` | **MISSING** |
| `normalizeStatuslinePreset()` 정규화 함수 | `internal/cli/profile_setup.go` (상단 helper) | **MISSING** |

#### 1.2.2 사용자 영향

- preset이 `compact`/`minimal`/`custom`인 사용자가 wizard 재실행 시 자동으로 `default` 또는 `full` 로 fall back될 위험 (현재 wizard가 preset 값을 read/write 하지 않으므로 sync에서 누락)
- 새 사용자가 wizard를 통해 preset 변경 불가 — `.moai/config/sections/statusline.yaml`을 수동 편집해야 함
- custom preset 사용자가 segments 토글 불가

### 1.3 Quick Win 정의

본 SPEC은 다음을 변경하지 **않는다**:
- `ProfilePreferences` struct (필드 이미 존재)
- `SyncToProjectConfig` / `syncStatusline` (sync 이미 작동)
- `internal/statusline/` renderer (출력 영역 무관)
- `.moai/config/sections/statusline.yaml` 스키마

본 SPEC은 오직 다음만 추가한다:
1. wizard Display group에 Preset Select widget
2. preset이 `custom`일 때 Segments MultiSelect widget
3. 4-locale 번역 키 + 정규화 함수 + 회귀 테스트

## 2. Requirements (EARS)

### REQ-SPW-001 (Event-Driven): Preset 입력 widget 추가

**WHEN** 사용자가 `moai profile setup [name]` 명령으로 wizard를 실행하고 Display group(`profile_setup.go:282-300`)에 도달하면 **THEN** wizard는 Statusline mode/theme Select 사이 또는 직후에 `huh.NewSelect[string]()` for StatuslinePreset을 표시해야 한다.

- 옵션: 4개 (`full`, `compact`, `minimal`, `custom`)
- Default: `normalizeStatuslinePreset(existingPrefs.StatuslinePreset)` 결과
- 결과: 선택값을 `prefs.StatuslinePreset` 필드에 저장
- 빈 문자열(`""`) 허용 — `syncStatusline`이 기존 yaml 값 보존

### REQ-SPW-002 (State-Driven): Segments 입력 widget — custom preset 조건부

**IF** 사용자가 REQ-SPW-001에서 `custom` preset을 선택했다면 **THEN** wizard는 후속 `huh.NewMultiSelect[string]()` for StatuslineSegments를 표시해야 한다.

- 옵션: 15개 segment 키 (`claude_version`, `context`, `directory`, `effort_thinking`, `git_branch`, `git_status`, `moai_version`, `model`, `output_style`, `pr`, `session_time`, `task`, `usage_5h`, `usage_7d`, `worktree`)
- Default: 기존 `existingPrefs.StatuslineSegments` map에서 `true` 값인 키 list
- 결과: 선택된 list를 `map[string]bool`로 변환 (선택=true, 미선택=false)
- huh의 conditional rendering은 `WithHideFunc` 또는 별도 group으로 구현 (실행 시점에 preset 값 평가)

### REQ-SPW-003 (Ubiquitous): 4-locale 번역 키 추가

The system **shall** add the following keys to `profileSetupText` struct and provide translations for all 4 locales (en / ko / ja / zh):

- `StatuslinePresetTitle`, `StatuslinePresetDesc`
- `PresetFull`, `PresetCompact`, `PresetMinimal`, `PresetCustom`
- `StatuslineSegmentsTitle`, `StatuslineSegmentsDesc`
- 15개 segment 레이블 키 (`SegmentModel`, `SegmentContext`, `SegmentOutputStyle`, `SegmentDirectory`, `SegmentGitStatus`, `SegmentGitBranch`, `SegmentClaudeVersion`, `SegmentMoaiVersion`, `SegmentSessionTime`, `SegmentUsage5h`, `SegmentUsage7d`, `SegmentEffortThinking`, `SegmentPR`, `SegmentTask`, `SegmentWorktree`)

기존 번역 패턴 (예: `ModelOpus`, `EffortLevelXHigh`)과 일관성을 유지한다. en이 canonical reference이며 ko/ja/zh는 번역 품질을 유지하되 segment 식별성을 위해 영어 키 이름을 괄호로 병기할 수 있다 (예: ko `"Model (모델)"`).

### REQ-SPW-004 (Ubiquitous): 정규화 함수 추가

The system **shall** add a `normalizeStatuslinePreset(p string) string` helper function in `profile_setup.go` near the existing `normalizeStatuslineMode` / `normalizeStatuslineTheme` helpers (top of file, around line 50-77 region), following the `isCanonical*` check pattern.

- 4개 canonical 값 (`full`, `compact`, `minimal`, `custom`)을 통과시키고
- 빈 문자열은 빈 문자열로 반환 (`syncStatusline`이 기존 yaml 값 보존하도록)
- 알 수 없는 값 (예: 이전 버전 prefs의 `fullbar`)은 빈 문자열로 reset

### REQ-SPW-005 (Unwanted + Event-Driven): 회귀 테스트

The system **shall not** allow regressions in the wizard's preset/segments input behavior. **WHEN** `go test ./internal/cli/...` is executed **THEN** the test suite must verify:

- 4 locale 각각에서 `StatuslinePresetTitle`, `StatuslinePresetDesc`, `StatuslineSegmentsTitle`, `StatuslineSegmentsDesc` 가 빈 문자열이 아님
- 4 locale 각각에서 모든 PresetFull/Compact/Minimal/Custom 레이블이 빈 문자열이 아님
- 4 locale 각각에서 15개 Segment 레이블이 모두 빈 문자열이 아님
- `normalizeStatuslinePreset` 유효 입력 4개 통과 + 무효 입력 (`"fullbar"`, `"invalid"`) → `""` 반환
- 정규화 + write + read round-trip이 등가 (existing prefs와 default 값 일치)

## 3. Acceptance Criteria (Binary)

전체 AC는 `acceptance.md`에 binary verification command 형식으로 정의된다. spec.md §3는 high-level summary만 제공한다.

| AC | 요약 | Binary Check |
|----|------|--------------|
| AC-SPW-001 | StatuslinePreset Select widget 존재 | `grep -n 'StatuslinePresetTitle' internal/cli/profile_setup.go` → ≥1 match |
| AC-SPW-002 | Segments MultiSelect widget 존재 (custom 조건부) | `grep -n 'StatuslineSegmentsTitle\|NewMultiSelect' internal/cli/profile_setup.go` → ≥2 matches |
| AC-SPW-003 | 4-locale 번역 키 모두 존재 (en/ko/ja/zh × Preset* + Segment* keys) | `go test -run TestProfileSetupTranslations_PresetSegments ./internal/cli/...` → PASS |
| AC-SPW-004 | `normalizeStatuslinePreset` 정규화 함수 정의 + 유효/무효 처리 | `go test -run TestNormalizeStatuslinePreset ./internal/cli/...` → PASS |
| AC-SPW-005 | 회귀 테스트 PASS + 전체 cli 패키지 회귀 없음 | `go test ./internal/cli/...` → PASS (NEW failures = 0) |
| AC-SPW-006 | Cross-platform build PASS | `go build ./...` + `GOOS=windows GOARCH=amd64 go build ./...` → exit 0 둘 다 |
| AC-SPW-007 | sync round-trip 정상 (`SyncToProjectConfig` 무변경 확인) | `git diff internal/profile/sync.go internal/profile/preferences.go` → no output |

각 AC의 단위 검증 명령은 `acceptance.md` Given-When-Then 시나리오에서 정확히 정의된다.

## Exclusions (What NOT to Build)

### 4.1 Out of Scope

본 SPEC에서 다루지 **않는** 항목:

- statusline renderer 변경 (`internal/statusline/renderer.go`, `builder.go`, `types.go`) — 출력 layout 무관
- statusline.yaml 스키마 변경 (15개 segment fixed list 유지)
- segment 추가/제거 — 별도 SPEC로 처리 (`SPEC-V3R5-STATUSLINE-SEGMENTS-EXTEND-NNN` 가칭, 향후)
- `ProfilePreferences` struct 필드 추가/변경
- `syncStatusline` / `SyncToProjectConfig` 로직 변경
- `defaultStatuslineSegments` 11→15 확장 (현 11이 의도된 보수적 default — 별도 결정 필요, `SPEC-V3R5-STATUSLINE-DEFAULT-FULL-NNN` 가칭)
- Path B/C/D 등 다른 dead key/wizard gap 처리 (이번 Quick Win은 preset/segments에 한정)
- `.moai/config/sections/statusline.yaml` 의 `preset: full` short-circuit 동작 (별도 SPEC `SPEC-V3R5-STATUSLINE-PRESET-FIX-001` proposed, not yet filed — project memory `feedback_workflow_inflation_root_cause` 등에서 언급, plan-auditor S5 fix)
- 4-locale 번역 품질 audit (en이 canonical, ko/ja/zh는 일관성 수준만 검증)

### 4.2 Edge Cases (must handle)

다음 edge case는 본 SPEC scope에 포함되며 acceptance.md에서 정확히 정의된다:

- **EC-SPW-001 (existing user preserve)**: 기존 prefs에 `StatuslinePreset: "compact"` 값이 있는 사용자가 wizard 재실행 시, Preset Select widget의 default 값이 `compact`로 표시되어야 한다 (값 손실 방지)
- **EC-SPW-002 (invalid legacy value)**: 이전 버전 또는 수동 편집으로 `StatuslinePreset: "fullbar"` (canonical 4개 외 값)이 prefs에 저장된 경우 wizard는 panic 없이 `normalizeStatuslinePreset` 통해 `""` 으로 fall back하고 사용자가 다시 선택할 수 있어야 한다 (forward-compat 보장)
- **EC-SPW-003 (empty preset preserved)**: 사용자가 빈 문자열을 선택 (또는 default 유지)한 경우 `prefs.StatuslinePreset = ""` 저장되며 `syncStatusline` (`sync.go:123-125`)이 기존 yaml의 `preset` 값을 보존해야 한다 (override 회피)
