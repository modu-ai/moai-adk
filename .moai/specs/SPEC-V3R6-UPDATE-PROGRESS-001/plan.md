---
id: SPEC-V3R6-UPDATE-PROGRESS-001
title: "SPEC-V3R6-UPDATE-PROGRESS-001 — Implementation Plan"
version: "0.1.0"
status: draft
created: 2026-05-23
updated: 2026-05-23
author: manager-spec
priority: P3
phase: "v3.6.0"
module: "internal/cli, internal/tui"
lifecycle: spec-anchored
tags: "v3r6, ux, tui, ansi-escape, progress-line, plan"
tier: S
---

# Implementation Plan — SPEC-V3R6-UPDATE-PROGRESS-001

## 1. 개요

`moai update` 출력 corruption (Defect #5) 정정을 위한 `tui.ProgressLine` 추상화 도입 및 ~22+ 호출 사이트 마이그레이션의 단계별 실행 계획.

본 SPEC은 **Tier S (Small)** 분류로, 단일 추상화 + 기계적 호출 사이트 교체로 구성된다. CLAUDE.local.md §23 Hybrid Trunk doctrine에 따라 plan-phase markdown은 main 직진 가능하며, run-phase는 1 PR 또는 main 직진 모두 허용된다.

## 2. 기술 접근 (Technical Approach)

### 2.1 ProgressLine API 시그니처

```go
// in internal/tui/status.go (or new internal/tui/progress_line.go)

// ProgressLineHandle is the handle returned by ProgressLine.
// It represents an in-flight progress line that can be transitioned
// to a Done or Fail terminal state, or updated mid-progress.
type ProgressLineHandle struct {
    out     io.Writer
    isTTY   bool
    theme   Theme
    indent  string  // e.g. "  " (two spaces) to match existing layout
    // ... internal state if needed
}

// ProgressLine starts rendering a progress line to out.
// On TTY: writes "  ○ {message}" without trailing newline.
// On non-TTY: writes "  ○ {message}\n" with newline.
// theme may be nil; LightTheme is used.
func ProgressLine(out io.Writer, message string, theme *Theme) *ProgressLineHandle

// Done transitions the handle to a success terminal state.
// On TTY: emits "\r\033[2K  ✓ {successMsg}\n" to clear and rewrite.
// On non-TTY: emits "  ✓ {successMsg}\n" on a new line.
func (h *ProgressLineHandle) Done(successMsg string)

// Fail transitions the handle to an error terminal state.
// On TTY: emits "\r\033[2K  ✗ {errMsg}\n".
// On non-TTY: emits "  ✗ {errMsg}\n".
func (h *ProgressLineHandle) Fail(errMsg string)

// Update re-renders the progress line with a new message (REQ-UPR-007 optional).
// On TTY: emits "\r\033[2K  ○ {newMsg}" preserving line-clear semantics.
// On non-TTY: emits "  ○ {newMsg}\n" on a new line.
func (h *ProgressLineHandle) Update(message string)
```

### 2.2 ANSI escape 선택 근거

- `\r` 단독: 커서 0 컬럼 이동만, 라인 길이 변화 미보호 → 결함 잔존
- `\r\033[2K`: CSI `2K` = Erase in Line, Mode 2 = entire line clear → 안전
- `\033[K` (Mode 0, erase cursor to end): cursor 위치부터 끝까지만 clear → `\r` 뒤에 사용 시 OK이지만 명시적 의도 표현 약함
- 채택: `\r\033[2K` (full-line clear, 의도 명시적, 모든 modern terminal에서 동작)

### 2.3 TTY 감지

`mattn/go-isatty.IsTerminal(fd uintptr)` 사용. `out` 이 `*os.File`인 경우 `out.Fd()`을, 그 외 (`bytes.Buffer`, `io.Writer` interface 등 — 주로 테스트)는 `false` 로 간주.

```go
func isTerminal(out io.Writer) bool {
    if f, ok := out.(*os.File); ok {
        return isatty.IsTerminal(f.Fd())
    }
    return false
}
```

### 2.4 Theme 통합

기존 `cliSuccess`, `cliError`, `cliMuted` (`internal/cli/update.go:51-57`) 는 cli 패키지 내부에 위치. `tui.ProgressLine`은 같은 색을 위해 `tui.Theme.Success` / `tui.Theme.Danger` / `tui.Theme.Dim` 토큰을 직접 사용. lipgloss `AdaptiveColor` 패턴 유지.

내부 헬퍼:

```go
func (h *ProgressLineHandle) symProgress() string  // theme.Dim → "○"
func (h *ProgressLineHandle) symSuccess() string   // theme.Success → "✓"
func (h *ProgressLineHandle) symError() string     // theme.Danger → "✗"
```

`tui.StatusIcon("run"|"ok"|"err")` 재사용 가능 (`internal/tui/status.go:18`).

## 3. Milestones

본 SPEC은 priority-based milestone 분해를 따른다 (시간 추정 없음, `.claude/rules/moai/core/agent-common-protocol.md` §Time Estimation 준수).

### M0 — Foundation: tui.ProgressLine API 신설 (Priority: Critical, blocking)

**Scope**:
- `internal/tui/progress_line.go` 신규 작성 (또는 `internal/tui/status.go` 확장 — `tui` 패키지 응집성 기준으로 결정)
- `ProgressLineHandle` struct + `ProgressLine` factory + `Done` / `Fail` / `Update` methods
- Internal TTY 감지 헬퍼 (`isTerminal`)
- Symbol 헬퍼 (theme 기반)

**Deliverables**:
- `internal/tui/progress_line.go` (또는 `status.go` 확장분)
- `internal/tui/progress_line_test.go` — golden output tests:
  - `TestProgressLine_TTY` — TTY 분기 `\r\033[2K` prefix 검증
  - `TestProgressLineNonTTY` — non-TTY 분기 `\n` 분리 출력 검증
  - `TestProgressLine_Update` — Update method 동작 (REQ-UPR-007)

**Verification**: `go test ./internal/tui/ -run TestProgressLine` PASS.

**Dependencies**: 없음. 본 milestone 완료 전까지 M1-M3 진행 불가.

### M1 — Migration: internal/cli/update.go 22 사이트 (Priority: High)

**Scope**: `internal/cli/update.go` 의 모든 `\r  %s` 페어를 `tui.ProgressLine` 호출로 교체.

**전형 변환 패턴**:

Before:
```go
_, _ = fmt.Fprintf(out, "  %s Backing up .moai/config...", symProgress())
// ... work ...
if backupErr != nil {
    _, _ = fmt.Fprintf(out, "\r  %s Backup failed: %v\n", symError(), backupErr)
} else {
    _, _ = fmt.Fprintf(out, "\r  %s .moai/config backed up\n", symSuccess())
}
```

After:
```go
pl := tui.ProgressLine(out, "Backing up .moai/config...", nil)
// ... work ...
if backupErr != nil {
    pl.Fail(fmt.Sprintf("Backup failed: %v", backupErr))
} else {
    pl.Done(".moai/config backed up")
}
```

**Site Inventory** (`grep -n '\\r  %s' internal/cli/update.go | wc -l` → 22 페어 후보, 실제 페어 매칭은 progress↔success/error 짝으로 검증):

- Lines 536-544: `.moai/config` backup
- Lines 558-569: template validation
- Lines 590-601: template deployment
- Lines 645-656: user config backup (2차)
- Lines 678-690: user settings restore
- Lines 1505-1516: glob/remove pair
- Lines 1522-1533: stat/remove pair
- Lines 1541-1549: legacy config remove
- Lines 1585-1600: legacy migration

(정확한 페어링은 M1 실행 시 호출 스코프 분석으로 확정)

**Verification**:
- `grep -rn '\\r  %s' internal/cli/update.go | grep -v _test.go | wc -l` → `0`
- `grep -rn 'tui\.ProgressLine\b' internal/cli/update.go | wc -l` → 22 (각 페어당 1 호출 site)
- `go test ./internal/cli/ -run TestUpdate` PASS (회귀 부재)
- 수동 smoke: `moai update --yes 2>&1 | cat | grep -E '[a-z]\.{3}'` → 0 matches

**Dependencies**: M0 완료.

### M2 — Migration: internal/cli/init.go + internal/cli/update_cleanup.go (Priority: Medium)

**Scope**: `internal/cli/init.go` 및 `internal/cli/update_cleanup.go` 잔여 `\r  %s` 페어 마이그레이션. M1과 동일 패턴 적용.

**Verification**:
- `grep -rn '\\r  %s' internal/cli/ | grep -v _test.go | wc -l` → `0` (cli 패키지 전체)
- 회귀 테스트 PASS

**Dependencies**: M1 완료 (M1과 병렬 진행 가능하나 동일 패턴 검증을 위해 순차 진행 권장).

### M3 — Validation and Status Update (Priority: Medium)

**Scope**:
- 통합 smoke test 실행 (`moai update --yes` 실제 호출)
- AC-UPR-001~009 일괄 검증
- `progress.md` 작성 (M0-M3 evidence 기록)
- `spec.md` `status: draft → implemented`, `version: 0.1.0 → 0.2.0`
- `golangci-lint run ./internal/tui/ ./internal/cli/` zero NEW issues
- Cross-platform 검증 (가능하면 darwin + linux + windows; 최소 darwin 필수)

**Verification**:
- 모든 AC PASS
- `go test ./...` PASS (or unchanged baseline)
- spec-lint warning 0 NEW

**Dependencies**: M2 완료.

## 4. Test Strategy

### 4.1 Unit Tests (M0)

`internal/tui/progress_line_test.go`:

- `TestProgressLine_TTY`: `*os.File` (pipe end with `isatty` true 강제)에 ProgressLine 출력 후 byte stream 캡쳐, `"\r\033[2K"` substring 존재 + 가시 메시지 string 일치 확인.
- `TestProgressLineNonTTY`: `*bytes.Buffer` (non-TTY)로 출력 후 progress + result가 별도 라인(`\n` 분리)이며 `"\r"` / `"\033"` 부재 확인.
- `TestProgressLine_Update`: TTY 모드에서 Update(newMsg) 후 `"\r\033[2K"` prefix + 새 메시지 확인.
- `TestProgressLine_DoneAfterFail`: Done과 Fail 중첩 호출 시 idempotent 동작 (또는 panic으로 명시 — design choice; M0 시 결정).

### 4.2 Integration Tests (M1-M2)

`internal/cli/update_test.go` (기존 테스트 확장):
- `*bytes.Buffer` out으로 `moai update` 코드 경로 실행 → non-TTY fallback이 정상 동작하는지 확인 (기존 회귀 차원)

### 4.3 Manual Smoke Test (M3)

```bash
# 실제 TTY 환경에서 moai update 실행
moai update --yes 2>&1 | tee /tmp/moai-update-smoke.log

# 출력 corruption sentinel: "[a-z]\.{3}" 패턴 = "g..." 등 트레일링 문자
grep -E '[a-z]\.{3}' /tmp/moai-update-smoke.log | wc -l
# expected: 0 (단 정상 메시지 중 "..." 자체는 progress 메시지에 포함되므로 결과 라인만 grep)

# 명확화: 결과 라인 ("✓" 또는 "✓"로 시작)만 검사
grep '^.*✓.*[a-z]\.{2,}' /tmp/moai-update-smoke.log | wc -l
# expected: 0
```

### 4.4 Cross-platform

- darwin (필수): M3 검증
- linux: CI workflow에서 자동 검증 (`go test ./...`)
- windows: lipgloss VT 모드가 기본 활성이므로 동작 기대; CI workflow에서 자동 검증

## 5. Risks and Mitigation (재정리)

| Risk | Mitigation |
|---|---|
| R1 (legacy Windows cmd.exe) | lipgloss feature-detect; worst case는 결함 이전 상태로 regression 아님 |
| R2 (non-TTY line count 증가) | REQ-UPR-003 명시 Plain mode + CI 로그 grep parser 영향 없음 검증 |
| R3 (마이그레이션 누락) | AC-UPR-004 sentinel grep 0 enforce; 각 M wave 후 grep 검증 |
| R4 (tui→isatty layering) | 이미 cli→isatty 의존 존재; tui→isatty 단방향 OK |

### 추가 위험 (감지 후 추가)

- **R5 — symProgress 기존 호출과 ProgressLine 신규 호출 혼재**: M1 진행 중 일부 사이트에서 `symProgress()` 단독 호출이 남거나 (페어 없음 단독 진행 메시지), ProgressLine과 중첩되는 경우. 마이그레이션 시 progress↔result 페어 매칭을 ast 검토 또는 line range 분석으로 확정.
- **R6 — golden test 갱신 부담**: TTY golden은 escape sequence를 byte literal로 포함하므로 readability가 낮음. `\\r\\033[2K` literal을 const로 추출하여 가독성 보존.

## 6. Out of Scope

### 6.1 Out of Scope — 본 plan 범위 외 (재확인)

- 애니메이션 스피너 라이브러리 도입 — out
- multi-line progress / progress bar - out (`tui.Progress` 별도)
- 색상 테마 전면 개편 — out
- `--no-progress` flag 추가 — out (REQ-UPR-003 fallback으로 충분)
- 별도 SPEC: UPDATE-ARCHIVE-CONTRACT-001 (`--force` + skip-sync), UPDATE-NOISE-001 (reserved filename + 3-way merge)

## 7. References

- `internal/tui/status.go` — Spinner stateless precedent (theme nil 허용, `LightTheme` fallback, no goroutine)
- `internal/cli/update.go:23` — `mattn/go-isatty` 기존 사용
- `internal/cli/update.go:51-57` — lipgloss AdaptiveColor 패턴
- `.claude/rules/moai/workflow/spec-workflow.md` — Tier S doctrine
- `.claude/rules/moai/languages/go.md` — Go tooling, error handling
- `SPEC-V3R6-GEARS-MIGRATION-001` — GEARS notation 캐논
- VT100 / ANSI X3.64 / ECMA-48 — CSI 2K (Erase in Line) reference

---

## HISTORY

- **2026-05-23 v0.1.0 draft** — 초기 작성. M0-M3 4 milestone 분해 (Foundation / update.go migration / init.go+cleanup migration / Validation). Tier S, R5-R6 추가 위험 식별.
