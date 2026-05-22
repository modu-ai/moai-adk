---
id: SPEC-V3R5-STATUSLINE-PROFILE-WIZARD-001
title: "Acceptance Criteria — profile setup wizard에 statusline preset/segments 입력 UI 추가"
version: "0.1.0"
created: 2026-05-22
updated: 2026-05-22
---

# Acceptance Criteria — SPEC-V3R5-STATUSLINE-PROFILE-WIZARD-001

7개 binary AC. 각 AC는 단일 verifiable command + 명시적 PASS 조건을 가진다. 본 acceptance.md는 plan-auditor와 run-phase Section E self-verification matrix가 동일 명령으로 PASS/FAIL 판정하도록 작성되었다.

REQ ↔ AC 매핑:

| REQ | AC | 검증 표면 |
|-----|----|----------|
| REQ-SPW-001 | AC-SPW-001 | profile_setup.go Display group에 Preset Select |
| REQ-SPW-002 | AC-SPW-002 | profile_setup.go Display group에 Segments MultiSelect (conditional) |
| REQ-SPW-003 | AC-SPW-003 | profile_setup_translations.go 4-locale × (Preset* + Segment*) keys |
| REQ-SPW-004 | AC-SPW-004 | profile_setup.go normalizeStatuslinePreset helper |
| REQ-SPW-005 | AC-SPW-005, AC-SPW-006 | 회귀 테스트 + cross-platform |
| (out-of-scope guard) | AC-SPW-007 | sync.go / preferences.go 무변경 |

---

## AC-SPW-001: StatuslinePreset Select widget 존재

**Given** `profile_setup.go` Display group (Section 4, line 282-300 region)
**When** 사용자가 wizard 4번째 group에 진입
**Then** Preset Select widget이 표시되고 4개 옵션 (`full`, `compact`, `minimal`, `custom`)을 노출한다.

### Binary Verification Commands

```bash
# 1. StatuslinePresetTitle 사용 확인 (huh.NewSelect Title 인자)
grep -n 'StatuslinePresetTitle' /Users/goos/MoAI/moai-adk-go/internal/cli/profile_setup.go
# Expected: ≥1 match
```

```bash
# 2. 4개 preset 옵션 모두 huh.NewOption 으로 정의되어 있는지 확인
grep -E 'huh\.NewOption\([^,]*(PresetFull|PresetCompact|PresetMinimal|PresetCustom)' /Users/goos/MoAI/moai-adk-go/internal/cli/profile_setup.go | wc -l
# Expected: 4 (exactly)
```

```bash
# 3. statuslinePreset 변수가 prefs 저장 시 사용되는지 확인
grep -E 'StatuslinePreset\s*:\s*statuslinePreset|StatuslinePreset:\s*normalizeStatuslinePreset' /Users/goos/MoAI/moai-adk-go/internal/cli/profile_setup.go
# Expected: ≥1 match
```

**PASS condition**: 세 명령 모두 만족.

---

## AC-SPW-002: Segments MultiSelect widget 존재 (custom 조건부)

**Given** AC-SPW-001 PASS 상태 (Preset Select 추가됨)
**When** 사용자가 `custom` preset을 선택
**Then** Segments MultiSelect widget이 표시되고 15개 segment 키를 노출한다. `custom` 외 preset 선택 시 widget은 숨겨지거나 무시된다.

### Binary Verification Commands

```bash
# 1. StatuslineSegmentsTitle 사용 확인
grep -n 'StatuslineSegmentsTitle' /Users/goos/MoAI/moai-adk-go/internal/cli/profile_setup.go
# Expected: ≥1 match
```

```bash
# 2. huh.NewMultiSelect widget 사용 확인 (이전에 wizard에 없던 신규 API)
grep -n 'huh\.NewMultiSelect' /Users/goos/MoAI/moai-adk-go/internal/cli/profile_setup.go
# Expected: ≥1 match
```

```bash
# 3. 15개 segment 키 모두 옵션으로 정의되어 있는지 확인
grep -cE 'huh\.NewOption\([^,]*(SegmentModel|SegmentContext|SegmentOutputStyle|SegmentDirectory|SegmentGitStatus|SegmentGitBranch|SegmentClaudeVersion|SegmentMoaiVersion|SegmentSessionTime|SegmentUsage5h|SegmentUsage7d|SegmentEffortThinking|SegmentPR|SegmentTask|SegmentWorktree)' /Users/goos/MoAI/moai-adk-go/internal/cli/profile_setup.go
# Expected: 15 (exactly)
```

```bash
# 4. custom 조건부 표시: WithHideFunc 또는 statuslinePreset 변수 참조 확인
grep -E 'WithHideFunc|statuslinePreset\s*!=\s*"custom"|statuslinePreset\s*==\s*"custom"' /Users/goos/MoAI/moai-adk-go/internal/cli/profile_setup.go
# Expected: ≥1 match (conditional rendering 구현 증거)
```

**PASS condition**: 네 명령 모두 만족.

---

## AC-SPW-003: 4-locale 번역 키 모두 존재

**Given** `profile_setup_translations.go`의 `profileSetupText` struct 및 `profileSetupTexts` map
**When** `go test -run TestProfileSetupTranslations_PresetSegments` 실행
**Then** 4-locale (en/ko/ja/zh) 각각에서 모든 신규 키가 빈 문자열이 아니다.

### Binary Verification Commands

```bash
# 1. profileSetupText struct에 신규 필드 모두 존재 확인 (struct 정의)
grep -E 'StatuslinePresetTitle\s+string|StatuslinePresetDesc\s+string|StatuslineSegmentsTitle\s+string|StatuslineSegmentsDesc\s+string|PresetFull\s+string|PresetCompact\s+string|PresetMinimal\s+string|PresetCustom\s+string' /Users/goos/MoAI/moai-adk-go/internal/cli/profile_setup_translations.go | wc -l
# Expected: ≥8 (8개 핵심 필드)
```

```bash
# 2. 15 segment 레이블 필드 정의 확인
grep -E 'Segment(Model|Context|OutputStyle|Directory|GitStatus|GitBranch|ClaudeVersion|MoaiVersion|SessionTime|Usage5h|Usage7d|EffortThinking|PR|Task|Worktree)\s+string' /Users/goos/MoAI/moai-adk-go/internal/cli/profile_setup_translations.go | wc -l
# Expected: 15 (exactly)
```

```bash
# 3. 4 locale × (8 핵심 + 15 segment) = 92 항목 모두 비어있지 않음 검증
go test -count=1 -run TestProfileSetupTranslations_PresetSegments ./internal/cli/...
# Expected: PASS (test 자체가 빈 문자열 검증 포함)
```

**PASS condition**: 세 명령 모두 만족.

---

## AC-SPW-004: normalizeStatuslinePreset 함수 정의 + 유효/무효 처리

**Given** `profile_setup.go` 상단 helper 영역
**When** `go test -run TestNormalizeStatuslinePreset` 실행
**Then** 유효 입력 (`full`, `compact`, `minimal`, `custom`)이 통과되고 무효 입력 (`fullbar`, `invalid`, etc)이 빈 문자열로 fall back된다.

### Binary Verification Commands

```bash
# 1. 함수 정의 존재 확인
grep -n 'func normalizeStatuslinePreset' /Users/goos/MoAI/moai-adk-go/internal/cli/profile_setup.go
# Expected: ≥1 match
```

```bash
# 2. isCanonicalStatuslinePreset 패턴 따름 확인 (statuslinePresetCanonical slice 권장)
grep -E 'statuslinePresetCanonical|isCanonicalStatuslinePreset' /Users/goos/MoAI/moai-adk-go/internal/cli/profile_setup.go
# Expected: ≥1 match
```

```bash
# 3. 유효/무효 처리 unit test PASS
go test -count=1 -run TestNormalizeStatuslinePreset ./internal/cli/...
# Expected: PASS
```

**PASS condition**: 세 명령 모두 만족.

---

## AC-SPW-005: 회귀 테스트 PASS + 전체 cli 패키지 회귀 없음

**Given** 변경 적용 후
**When** `go test ./internal/cli/...` 실행
**Then** 전체 cli 패키지 테스트가 PASS이며 NEW failure가 0이다.

### Binary Verification Commands

```bash
# 1. 전체 cli 패키지 회귀 테스트
go test -count=1 ./internal/cli/...
# Expected: PASS (exit 0)
```

```bash
# 2. 신규 추가 테스트 함수 모두 PASS
go test -count=1 -v -run 'TestProfileSetupTranslations_PresetSegments|TestNormalizeStatuslinePreset' ./internal/cli/...
# Expected: PASS, 각 테스트가 출력에 PASS로 표시됨
```

```bash
# 3. Coverage 측정 (cli 패키지 baseline 유지 — drop 허용 ≤ 2pp)
go test -count=1 -cover ./internal/cli/...
# Expected: coverage 측정 가능, 새 헬퍼/widget 코드도 적어도 한 번 실행됨
```

**PASS condition**: 세 명령 모두 만족.

---

## AC-SPW-006: Cross-platform build PASS

**Given** 변경 적용 후
**When** darwin/amd64 + windows/amd64 build
**Then** 두 GOOS 모두 build exit 0.

### Binary Verification Commands

```bash
# 1. native (darwin) build
go build ./...
# Expected: exit 0
```

```bash
# 2. windows cross-compile
GOOS=windows GOARCH=amd64 go build ./...
# Expected: exit 0
```

```bash
# 3. linux cross-compile (extra safety)
GOOS=linux GOARCH=amd64 go build ./...
# Expected: exit 0
```

**PASS condition**: 세 명령 모두 exit 0.

---

## AC-SPW-007: sync 영역 무변경 (out-of-scope guard)

**Given** SPEC scope는 wizard UI에 한정 (sync logic은 이미 작동)
**When** SPEC 변경 적용 후 git diff
**Then** `internal/profile/sync.go`, `internal/profile/preferences.go`, `internal/statusline/` 의 source 파일이 byte-identical.

### Binary Verification Commands

```bash
# 1. sync.go + preferences.go 변경 없음 검증 (main 기준)
git diff main -- internal/profile/sync.go internal/profile/preferences.go
# Expected: empty output
```

```bash
# 2. statusline renderer 변경 없음 검증
git diff main -- internal/statusline/renderer.go internal/statusline/builder.go internal/statusline/types.go
# Expected: empty output
```

```bash
# 3. .moai/config/sections/statusline.yaml 변경 없음 (template + local 둘 다)
git diff main -- internal/template/templates/.moai/config/sections/statusline.yaml .moai/config/sections/statusline.yaml
# Expected: empty output
```

**PASS condition**: 세 명령 모두 empty output (no diff).

---

## Edge Case Scenarios (Given-When-Then)

### EC-SPW-001: 기존 preset 값 보존 (preserve existing value)

**Given** 사용자 `~/.moai/claude-profiles/work/preferences.yaml` 에 다음이 저장됨:
```yaml
statusline_preset: compact
statusline_segments:
  context: true
  directory: true
  git_branch: true
```

**When** `moai profile setup work` 실행 후 Display group에 도달

**Then** Preset Select widget의 default(highlighted) 선택지가 `compact`이며, 사용자가 Enter만 눌러도 `prefs.StatuslinePreset == "compact"`로 유지되어 저장된다.

Verification: wizard interactive test는 자동화 어려우므로 unit test가 default 값 결정 로직(`normalizeStatuslinePreset(existingPrefs.StatuslinePreset)`)을 검증한다.

```bash
go test -count=1 -run 'TestNormalizeStatuslinePreset/preserves_compact' ./internal/cli/...
# Expected: PASS
```

### EC-SPW-002: 잘못된 legacy 값 fall back (forward-compat)

**Given** 사용자 prefs.yaml에 `statusline_preset: fullbar` (canonical 4개 외 값) 저장됨 (이전 버전 또는 수동 편집)

**When** `moai profile setup` 실행

**Then** wizard는 panic 없이 시작하고, Preset Select의 default 값이 `""` (또는 첫 옵션)으로 reset되며, 사용자는 4개 유효 옵션 중에서 다시 선택할 수 있다.

Verification:

```bash
go test -count=1 -run 'TestNormalizeStatuslinePreset/rejects_invalid' ./internal/cli/...
# Expected: PASS — input "fullbar" → output ""
```

### EC-SPW-003: 빈 preset 보존 (no override)

**Given** 사용자가 wizard에서 Preset Select의 default(빈 문자열 또는 미선택) 상태로 진행

**When** wizard 완료 후 prefs 저장 + `SyncToProjectConfig` 호출

**Then** `syncStatusline`(`sync.go:123-125`)의 `if prefs.StatuslinePreset != ""` 분기가 false로 평가되어 기존 `.moai/config/sections/statusline.yaml`의 `preset` 값이 보존된다 (override 회피).

Verification: 본 동작은 기존 sync 로직(AC-SPW-007이 byte-identical 보장) 이므로 별도 unit test 추가 없이 sync 로직 회귀 테스트에 의해 보장된다.

```bash
# 기존 sync 회귀 테스트 PASS 유지
go test -count=1 ./internal/profile/...
# Expected: PASS (no regression)
```

---

## Definition of Done

다음 조건이 모두 충족되면 SPEC 구현 완료로 간주한다:

- [ ] AC-SPW-001 ~ AC-SPW-007 모두 PASS
- [ ] EC-SPW-001 ~ EC-SPW-003 모두 verification command PASS
- [ ] `go test -count=1 ./...` 전체 PASS (또는 NEW failure 0)
- [ ] `golangci-lint run --timeout=2m` NEW issues = 0
- [ ] `go build ./...` + `GOOS=windows GOARCH=amd64 go build ./...` 둘 다 exit 0
- [ ] spec.md status `draft → implemented`, version `0.1.0 → 0.2.0` 갱신
- [ ] HISTORY 행 추가 (run-phase 작업 요약 + manager-develop verdict)
- [ ] Conventional Commit `plan(SPEC-V3R5-STATUSLINE-PROFILE-WIZARD-001): ...` (plan-phase) / `feat(...)` (run-phase) 메시지 + `🗿 MoAI <email@mo.ai.kr>` trailer
- [ ] PRESERVE 대상 파일 (pre-existing dirty + 무관 untracked) 무변경 — `git status --porcelain` 으로 변경 범위 확인
