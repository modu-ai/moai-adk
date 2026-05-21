---
id: SPEC-V3R5-STATUSLINE-PROFILE-WIZARD-001
title: "Implementation Plan — profile setup wizard에 statusline preset/segments 입력 UI 추가"
version: "0.1.0"
created: 2026-05-22
updated: 2026-05-22
---

# Implementation Plan — SPEC-V3R5-STATUSLINE-PROFILE-WIZARD-001

Tier S — 3-artifact 채택 근거: spec/plan/acceptance 3 분리. AC가 binary verification command 중심이며 plan-auditor + run-phase Self-Verification matrix가 동일 명령 참조해야 하므로 acceptance.md를 별도 분리. spec.md inline AC는 high-level summary로 유지. Tier S 표준 2-artifact는 단순 SPEC에 적합하나, 본 SPEC은 7 AC + 3 Edge Case + 4 locale × 23 keys = 92 translation 데이터 검증 → acceptance.md 분리가 plan-auditor 평가 효율을 높인다.

## 1. Implementation Strategy

### 1.1 핵심 원칙

- **Frontend-Only Change**: backend(`internal/profile/preferences.go`, `internal/profile/sync.go`)는 무변경. AC-SPW-007이 byte-identical 검증.
- **Sequential, Single-File Atomic Edit**: 각 파일은 한 번에 모든 관련 변경을 atomic MultiEdit으로 적용 (Tier S small-scope 권장 패턴).
- **Test-First for New Functions**: `normalizeStatuslinePreset`은 helper로 단순 — 함수 정의 + unit test를 같은 commit에 포함.
- **Locale Parity**: en이 canonical reference. ko/ja/zh는 en과 1:1 키 매칭 검증 (test가 자동 보장).

### 1.2 Brownfield Strategy (PRESERVE)

다음 파일은 wizard에서 호출되지만 본 SPEC scope에서 무변경:

| 파일 | 이유 |
|------|------|
| `internal/profile/preferences.go` | StatuslinePreset/StatuslineSegments 필드 이미 존재 (line 46-47) |
| `internal/profile/sync.go` | syncStatusline + defaultStatuslineSegments 이미 작동 |
| `internal/statusline/renderer.go` | renderer 영역 (별도 SPEC `STATUSLINE-PRESET-FIX-001`) |
| `internal/statusline/builder.go` | builder data 수집 영역 |
| `.moai/config/sections/statusline.yaml` | scheme 무변경 |
| `internal/template/templates/.moai/config/sections/statusline.yaml` | template parity |

### 1.3 위험 (Risks)

| ID | Risk | Likelihood | Impact | Mitigation |
|----|------|-----------|--------|------------|
| R-SPW-001 | huh `WithHideFunc` conditional rendering이 form lifecycle 시점에 statuslinePreset 값을 평가하지 못함 (huh 버전 의존성) | M | M | `huh.NewGroup`을 2개로 분리: 첫 group(preset Select)이 끝난 후 두 번째 group(segments MultiSelect with `WithHideFunc`) 진입. 또는 사용자가 custom 선택 시만 prefs.StatuslineSegments 저장. **fallback**: `WithHideFunc` 작동 안 하면 form 분리(`form1.Run()` 후 if preset == "custom" then form2.Run()) 방식으로 전환. M1 자체 검증 필수. |
| R-SPW-002 | Translation parity 누락 (en 추가 후 ko/ja/zh 중 일부 누락) | L | M | TestProfileSetupTranslations_PresetSegments가 4 locale 모두 검증 (loop으로 빈 문자열 확인). CI에서 강제. |
| R-SPW-003 | huh v0.x → v0.7+ API 차이 (MultiSelect 옵션 API 변경) | L | L | 현재 wizard 코드에서 `huh.NewSelect` + `huh.NewOption` 사용 패턴 확인 후 동일 스타일로 작성. go.mod의 huh 버전 확인 후 implementation. |
| R-SPW-004 | 11-segment defaultStatuslineSegments vs 15-segment statusline.yaml mismatch 노출 | L | L | 본 SPEC에서 다루지 않음 (out-of-scope §3.2 명시). MultiSelect는 15개 모두 표시하되 default 선택은 기존 prefs.StatuslineSegments 값을 따른다 (sync.go 보수적 default 보존). |

## 2. Milestones (Priority-Based)

phase 순서: M1 → M2 → M3 → M4. 시간 추정 없음 (`time-estimation` HARD 준수).

### M1 (Priority: High): Translation 데이터 + helper 함수

**Goal**: profile_setup_translations.go에 신규 23개 키(8 핵심 + 15 segment)를 4 locale로 추가하고 normalizeStatuslinePreset helper를 정의한다.

**Atomic Edit Targets**:
- `internal/cli/profile_setup_translations.go`: profileSetupText struct에 23개 필드 추가 + 4 locale × 23 entries 채우기 (atomic MultiEdit 1회)
- `internal/cli/profile_setup.go`: file 상단(~line 50-77 region)에 `statuslinePresetCanonical` slice + `isCanonicalStatuslinePreset` + `normalizeStatuslinePreset` helper 추가 (atomic MultiEdit 1회)

**Verification**: AC-SPW-003 + AC-SPW-004의 grep + unit test 명령으로 PASS 확인. cross-platform build pre-check.

**Dependencies**: 없음 (entry point).

### M2 (Priority: High): wizard Display group widget 추가

**Goal**: profile_setup.go의 huh.NewForm 4번째 Display group(line 282-300)에 Preset Select + Segments MultiSelect widget을 추가한다.

**Atomic Edit Targets**:
- `internal/cli/profile_setup.go`: Display group 본문 + statuslinePreset/statuslineSegmentsSelection 변수 선언 + prefs 저장 시점 변환 로직 (atomic MultiEdit 1회)

**huh API 사용 패턴 (검증 후 확정, R-SPW-001 mitigation)**:
- 1단계: 기존 Display group에 Preset Select 추가
- 2단계: huh.NewGroup 분리 + `WithHideFunc(func() bool { return statuslinePreset != "custom" })` 시도
- 3단계: 작동 안 하면 form 2개로 분리 (form1.Run() → if custom then form2.Run())

**Conversion logic** (선택된 string slice → `map[string]bool`):
- 모든 15 segment 키에 대해 선택 list 포함 여부로 true/false 결정
- 결과를 `prefs.StatuslineSegments` 에 저장

**Verification**: AC-SPW-001 + AC-SPW-002의 grep 명령으로 PASS 확인.

**Dependencies**: M1 (translation 키 + normalizer 필요).

### M3 (Priority: High): 회귀 테스트 추가

**Goal**: AC-SPW-003 + AC-SPW-004 + AC-SPW-005 검증을 위한 test 함수 추가.

**Atomic Edit Targets**:
- `internal/cli/profile_setup_translations_test.go`: `TestProfileSetupTranslations_PresetSegments` 추가 — 4 locale × 23 키 빈 문자열 검증 (table-driven)
- `internal/cli/profile_setup_test.go` 또는 신규 `internal/cli/normalize_statusline_preset_test.go`: `TestNormalizeStatuslinePreset` 추가 — 유효 입력 4건 + 무효 입력 2건 + 빈 문자열 1건 = 7 sub-test cases (table-driven, EC-SPW-001/002/003 커버)

**Verification**: `go test ./internal/cli/...` 전체 PASS. NEW failure 0. coverage 측정.

**Dependencies**: M1 + M2 (검증 대상 코드 존재 필요).

### M4 (Priority: Medium): 최종 검증 + commit 준비

**Goal**: 7개 AC + 3개 Edge Case + cross-platform 모두 PASS 상태 확인 후 commit.

**Tasks**:
- AC-SPW-001 ~ AC-SPW-007 binary verification 명령 7-item parallel batch 실행 (verification-batch-pattern 적용)
- `golangci-lint run --timeout=2m` NEW issues 0 확인
- `git diff --stat` 으로 변경 파일 범위 확인 (out-of-scope 파일 변경 없는지)
- HISTORY 행 추가 + status 갱신은 manager-docs sync-phase에서 처리 (run-phase에서는 status `draft` 유지)
- Conventional Commit 메시지 작성: `feat(SPEC-V3R5-STATUSLINE-PROFILE-WIZARD-001): wizard에 statusline preset/segments 입력 UI 추가`

**Dependencies**: M1 + M2 + M3 모두 완료.

## 3. Technical Approach

### 3.1 파일 변경 매트릭스 (verified line ranges)

| 파일 | 종류 | 변경 종류 | 예상 LOC delta |
|------|------|----------|----------------|
| `internal/cli/profile_setup.go` (412 lines) | source | helper 추가 + Display group 확장 | +50 ~ +70 |
| `internal/cli/profile_setup_translations.go` (418 lines) | source | struct 확장 + 4 locale × 23 entries | +100 ~ +130 |
| `internal/cli/profile_setup_translations_test.go` (existing) | test | TestProfileSetupTranslations_PresetSegments 추가 | +30 ~ +50 |
| `internal/cli/profile_setup_test.go` (1930 bytes, existing) | test | TestNormalizeStatuslinePreset 추가 | +20 ~ +30 |

총 LOC delta: ~200 ~ ~280 < Tier S 임계 300 LOC. 영향 source 파일 2개 + test 파일 2개 = 4 < 5 Tier S 임계.

### 3.2 huh API 사용 패턴 (verified from profile_setup.go:282-300)

기존 패턴 (Display group):

```go
huh.NewGroup(
    huh.NewSelect[string]().
        Title(t.StatuslineModeTitle).
        Description(t.StatuslineModeDesc).
        Options(
            huh.NewOption(t.ModeDefault, "default"),
            huh.NewOption(t.ModeFull, "full"),
        ).
        Value(&statuslineMode),
    // ... StatuslineTheme Select
).Title(t.DisplayTitle),
```

본 SPEC이 추가하는 패턴 (예시, R-SPW-001 mitigation에 따라 최종 형태는 implementation 시 확정):

```go
// In Display group 또는 별도 group:
huh.NewSelect[string]().
    Title(t.StatuslinePresetTitle).
    Description(t.StatuslinePresetDesc).
    Options(
        huh.NewOption(t.PresetFull, "full"),
        huh.NewOption(t.PresetCompact, "compact"),
        huh.NewOption(t.PresetMinimal, "minimal"),
        huh.NewOption(t.PresetCustom, "custom"),
    ).
    Value(&statuslinePreset),

// Conditional MultiSelect (별도 group with WithHideFunc):
huh.NewGroup(
    huh.NewMultiSelect[string]().
        Title(t.StatuslineSegmentsTitle).
        Description(t.StatuslineSegmentsDesc).
        Options(
            huh.NewOption(t.SegmentModel, "model"),
            // ... 14 more
        ).
        Value(&statuslineSegmentsSelection),
).WithHideFunc(func() bool { return statuslinePreset != "custom" })
```

### 3.3 Helper 함수 패턴 (verified from profile_setup.go:18-77)

기존 패턴 (`statuslineModeCanonical`, `isCanonicalStatuslineMode`, `normalizeStatuslineMode`):

```go
var statuslinePresetCanonical = []string{"full", "compact", "minimal", "custom"}

func isCanonicalStatuslinePreset(s string) bool {
    for _, v := range statuslinePresetCanonical {
        if v == s {
            return true
        }
    }
    return false
}

func normalizeStatuslinePreset(p string) string {
    if p == "" {
        return ""  // EC-SPW-003: 빈 문자열 그대로 보존 → syncStatusline 기존 yaml 값 유지
    }
    if isCanonicalStatuslinePreset(p) {
        return p  // EC-SPW-001: 유효 값 보존
    }
    return ""  // EC-SPW-002: 무효 값 reset
}
```

### 3.4 Segments map ↔ string slice 변환 (M2 보조 로직)

prefs read (default 결정):

```go
// existing prefs.StatuslineSegments map[string]bool 에서 true 값 키만 slice로 추출
var statuslineSegmentsSelection []string
if existingPrefs.StatuslineSegments != nil {
    for key, enabled := range existingPrefs.StatuslineSegments {
        if enabled {
            statuslineSegmentsSelection = append(statuslineSegmentsSelection, key)
        }
    }
}
```

prefs write (저장 시점 변환):

```go
// MultiSelect 결과 (선택된 string slice) → map[string]bool
// 15개 segment 모두에 대해 explicit true/false 설정
allSegmentKeys := []string{
    "claude_version", "context", "directory", "effort_thinking",
    "git_branch", "git_status", "moai_version", "model",
    "output_style", "pr", "session_time", "task",
    "usage_5h", "usage_7d", "worktree",
}
selectedSet := make(map[string]bool, len(statuslineSegmentsSelection))
for _, key := range statuslineSegmentsSelection {
    selectedSet[key] = true
}
segmentsMap := make(map[string]bool, len(allSegmentKeys))
for _, key := range allSegmentKeys {
    segmentsMap[key] = selectedSet[key]
}
// 단, preset != "custom" 인 경우 segmentsMap을 nil로 두어 syncStatusline 기존 yaml 보존
if statuslinePreset == "custom" {
    prefs.StatuslineSegments = segmentsMap
}
```

### 3.5 Known Issues 적용 (Section B filtered for Tier S)

- **B4 frontmatter**: spec/plan/acceptance 3 파일 모두 canonical 12-field schema 준수 (`created:` not `created_at:`, `tags:` not `labels:`).
- **B5 CI 3-tier**: spec-lint + golangci-lint + Test (per OS) 각각 PASS 의무. AC-SPW-005/006이 검증.
- **B6 spec-lint heading**: `### 4.1 Out of Scope` h3 사용 (spec.md). `## Exclusions (What NOT to Build)` h2 wrapper. MissingExclusions ERROR 회피.
- **B8 working tree hygiene**: pre-existing 11개 dirty(M) + 11개 untracked(??) PRESERVE — SPEC artifacts 3개만 touch.

## 4. Pre-flight Check (Section C)

Plan-auditor + run-phase 진입 전 확인:

```bash
# 1. 현재 branch + baseline
git rev-parse HEAD  # → 02a0a8304be7330dd9ba62cbf24be3976f8f0ee6 (refreshed 2026-05-22, plan-auditor S1 fix)
git branch --show-current  # → main

# 2. cross-platform build 사전 확인 (NEW baseline)
go build ./...
GOOS=windows GOARCH=amd64 go build ./...

# 3. profile + cli + statusline 패키지 baseline
go test -count=1 ./internal/cli/... ./internal/profile/... ./internal/statusline/...

# 4. lint baseline
golangci-lint run --timeout=2m ./internal/cli/... ./internal/profile/... 2>&1 | tail -10

# 5. PRESERVE 대상 enumeration
git status --porcelain | grep -v "^??" | head -20  # M files (preserved)
git status --porcelain | grep "^??" | head -20     # untracked (preserved)
```

## 5. Constraints (Section D)

### 5.1 DO NOT VIOLATE

- DO NOT modify `internal/profile/preferences.go` (StatuslinePreset/StatuslineSegments 필드 이미 존재)
- DO NOT modify `internal/profile/sync.go` (syncStatusline 작동)
- DO NOT modify `internal/statusline/` (renderer 영역 무관)
- DO NOT modify `.moai/config/sections/statusline.yaml` (스키마 무변경)
- DO NOT modify `internal/template/templates/.moai/config/sections/statusline.yaml` (template parity)
- DO NOT touch pre-existing 11 dirty(M) files + 11 untracked(??) files (PRESERVE)
- DO NOT amend commits, DO NOT force-push, DO NOT use `--no-verify`
- DO NOT add `time-estimation` (Priority labels only)
- DO NOT use AskUserQuestion (subagent boundary HARD — blocker report only)

### 5.2 사용 의무

- Conventional Commits: `plan(SPEC-V3R5-STATUSLINE-PROFILE-WIZARD-001): ...` (plan-phase) / `feat(SPEC-V3R5-STATUSLINE-PROFILE-WIZARD-001): ...` (run-phase)
- Commit trailer: `🗿 MoAI <email@mo.ai.kr>`
- Late-Branch policy (main 직진, push deferred to sync-phase per SPEC-V3R5-LATE-BRANCH-001 REQ-LB-005)

## 6. Self-Verification Deliverables (Section E)

run-phase manager-develop이 보고서에 포함할 사항:

### E1. AC Binary PASS/FAIL Matrix

| AC | Status | Verification Command | Actual Output |
|----|--------|---------------------|---------------|
| AC-SPW-001 | TBD | `grep -n 'StatuslinePresetTitle' internal/cli/profile_setup.go` | TBD |
| AC-SPW-002 | TBD | `grep -n 'huh\.NewMultiSelect' internal/cli/profile_setup.go` | TBD |
| AC-SPW-003 | TBD | `go test -run TestProfileSetupTranslations_PresetSegments ./internal/cli/...` | TBD |
| AC-SPW-004 | TBD | `go test -run TestNormalizeStatuslinePreset ./internal/cli/...` | TBD |
| AC-SPW-005 | TBD | `go test -count=1 ./internal/cli/...` | TBD |
| AC-SPW-006 | TBD | `go build ./...` + `GOOS=windows ... go build ./...` | TBD |
| AC-SPW-007 | TBD | `git diff main -- internal/profile/sync.go internal/profile/preferences.go` | TBD |

### E2. Cross-Platform Build 결과

```
$ go build ./...                          → exit 0
$ GOOS=windows GOARCH=amd64 go build ./... → exit 0
$ GOOS=linux GOARCH=amd64 go build ./...   → exit 0
```

### E3. Coverage 측정

```
$ go test -count=1 -cover ./internal/cli/...
```

(cli 패키지 coverage baseline 유지 — 신규 helper + widget 코드 실행 경로 포함 확인)

### E4. Subagent Boundary Grep (C-HRA-008 류 — Tier S filter)

본 SPEC은 `internal/cli/` 영역 변경이므로 C-HRA-008 (subagent 도메인 `internal/harness/`, `internal/hook/` 한정) 직접 영향 없음. 그러나 일관성 확인:

```
$ grep -rn 'AskUserQuestion\|mcp__askuser' internal/cli/profile_setup.go internal/cli/profile_setup_translations.go | grep -v "_test.go" | grep -v "// "
(no output expected)
```

### E5. Lint Status

```
$ golangci-lint run --timeout=2m
```

NEW issues 발견 시 explicit report (line + diagnosis). pre-existing baseline은 별도 mark.

### E6. Branch HEAD + Push 상태

- 새 commits SHA 리스트
- 본 SPEC plan-phase: main 직진 (Late-Branch). push는 sync-phase에 통합.

### E7. Blocker Report (있을 시)

위임 prompt에서 명시 안 된 사용자 결정 필요 항목 (예: huh `WithHideFunc` 미작동 시 form 분리 결정, segment 정렬 순서, ja/zh 번역 품질 검토 요청 등)이 발견되면 structured 보고. **AskUserQuestion 절대 호출 금지** (subagent boundary HARD per agent-common-protocol).

## 7. Risks & Mitigations Summary

| Phase | Risk | Mitigation |
|-------|------|------------|
| M1 | translation 키 누락 | 23개 키 + 4 locale을 atomic MultiEdit으로 묶음. 즉시 TestProfileSetupTranslations_PresetSegments로 검증 |
| M2 | huh `WithHideFunc` 미작동 | form 분리 fallback 준비 (3.2 참조). M2 자체 검증으로 작동 여부 확인 |
| M2 | segments map 변환 오류 | M3 unit test (M2와 함께 작성)가 round-trip 보장. 15개 키 enumeration 명시 |
| M3 | test가 cli 패키지 외 segment 영향 | _test.go 파일은 `package cli` 내에 위치. profile/statusline 패키지 호출 없음 (translation slice + normalizer만 검증) |
| M4 | lint NEW issues | M1-M3 commit 전 incremental lint 확인. gofmt + go vet 사전 통과 |

## 8. Out of Scope (h3 — spec-lint MissingExclusions 회피)

### 8.1 본 SPEC 미포함

- statusline renderer 변경
- statusline.yaml 스키마 변경
- segment list 추가/제거
- Path B/C/D 다른 wizard gap 처리
- defaultStatuslineSegments 11→15 확장
- 4-locale 번역 audit (quality grade)
- preset `full` short-circuit 동작 변경 (`STATUSLINE-PRESET-FIX-001` 가칭)

## 9. Tier S 정당화

본 SPEC이 Tier S 분류 기준을 만족함을 명시:

| 기준 | 임계 | 본 SPEC | 결과 |
|------|------|---------|------|
| LOC delta | < 300 | ~200-280 | PASS |
| 영향 source 파일 | < 5 | 2 (profile_setup.go + profile_setup_translations.go) | PASS |
| 영향 test 파일 | (test는 별도 카운트) | 2 (기존 _test.go 확장 + 신규 함수) | OK |
| 영향 패키지 | 1 | `internal/cli` 만 | PASS |
| Backend 변경 | 없음 | 없음 | PASS |
| Cross-platform 회귀 위험 | 낮음 | huh API는 cross-platform — windows/linux/darwin 모두 동일 동작 | PASS |
| plan-auditor 임계 | 0.75 | target | TBD |

Tier S 채택 결정 — plan-auditor PASS 임계 0.75 적용. Section A-E delegation template 적용은 OPTIONAL이나, 본 plan.md는 Section C/D/E 명시적으로 포함하여 run-phase Section A-E 완성도를 높였다 (5-section 권장 패턴).
