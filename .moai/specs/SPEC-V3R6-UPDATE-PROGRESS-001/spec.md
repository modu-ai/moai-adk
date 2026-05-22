---
id: SPEC-V3R6-UPDATE-PROGRESS-001
title: "moai update 진행 라인 출력 깨짐 정정 — tui.ProgressLine 추상화"
version: "0.1.0"
status: draft
created: 2026-05-23
updated: 2026-05-23
author: manager-spec
priority: P3
phase: "v3.6.0"
module: "internal/cli, internal/tui"
lifecycle: spec-anchored
tags: "v3r6, ux, tui, ansi-escape, progress-line"
tier: S
---

# SPEC-V3R6-UPDATE-PROGRESS-001 — moai update 진행 라인 출력 깨짐 정정 (tui.ProgressLine 추상화)

## Section A — Context and Motivation

### A.1 배경

`moai update` 플로우는 사용자에게 진행 상태를 보여주기 위해 `internal/cli/update.go`, `internal/cli/init.go`, `internal/cli/update_cleanup.go` 전반에 걸쳐 **"`\r` 덮어쓰기"** 패턴을 사용한다. 이 패턴은 동일 라인에 진행 메시지를 먼저 출력하고, 작업 완료 후 캐리지 리턴(`\r`)으로 커서를 0 컬럼으로 되돌린 뒤 결과 메시지를 다시 쓰는 방식이다.

전형적 호출 페어 (`internal/cli/update.go:536`, `:543` 등):

- 진행 메시지: `fmt.Fprintf(out, "  %s Backing up .moai/config...", symProgress())`
- 결과 메시지: `fmt.Fprintf(out, "\r  %s .moai/config backed up\n", symSuccess())`

### A.2 결함 — Defect #5 (Low Cosmetic)

2026-05-23 세션 `moai update` 감사에서 출력 corruption이 사용자 로그로 표면화되었다 (`/Users/goos/MoAI/moai-adk-go/CLAUDE.local.md` §11 frequent issues 패밀리, audit 별도 보고).

실제 재현 로그 (사용자 환경 2026-05-23):

```
✓ .moai/config backed upg...
✓ Removed .claude/settings.jsonn...
✓ Removed .claude/commands/moaii...
```

각 라인 끝의 "g...", "n...", "i..." 꼬리는 직전 progress 메시지(`Backing up .moai/config...` 등)의 잔여 문자다. 결과 메시지가 진행 메시지보다 짧을 때 `\r`만으로는 라인 전체를 덮어쓰지 못한다.

### A.3 근본 원인

- 진행 라인이 N 글자로 출력됨
- `\r` 가 커서를 0 컬럼으로 되돌림
- 결과 라인이 M 글자로 출력 (`M < N`)
- (N - M) 만큼의 트레일링 문자가 이전 라인의 끝부분으로 남음

이 패턴은 코드 전반에 약 15-20곳의 페어로 흩어져 있으며 (`internal/cli/update.go`에서만 22 hit 확인 `grep -rn '\\r  %s' internal/cli/ | grep -v _test.go | wc -l → 22`), 호출 시점마다 메시지 길이를 일일이 계산하거나 패딩하지 않으면 동일한 결함이 재발한다.

### A.4 영향 (Cosmetic이지만 UX 인상 손상)

- **기능적 영향 없음** — 데이터 손상, 백업 실패, 마이그레이션 오류 없음
- **UX 인상 손상** — `moai update`는 신규 사용자가 가장 먼저 마주하는 동기화 명령이며, 깨진 출력은 도구 신뢰도(품질 인상)에 직접 영향
- **회귀 위험** — 새 진행/결과 페어가 추가될 때마다 동일 결함이 재발할 구조적 결함

### A.5 의도

`tui.ProgressLine` 추상화를 도입하여 진행 라인 전이를 **재사용 가능한 단일 책임 헬퍼**로 캡슐화하고, 모든 기존 `\r` 페어를 마이그레이션함으로써:

1. 결함의 즉시 해소 (line clear escape sequence 적용)
2. 향후 회귀 차단 (호출자가 길이 산술을 알 필요 없음)
3. Non-TTY 환경 fallback 명문화 (CI 로그, 파이프 출력)

---

## Section B — Goals and Non-goals

### B.1 Goals (목표)

- G1: `internal/tui/` 패키지에 `ProgressLine` API 신설 — 진행 라인을 깨끗하게 전이하는 추상화
- G2: ANSI escape `\r\033[2K` (caret return + erase entire line)을 내부 구현으로 사용해 라인 전체를 명시적으로 clear
- G3: Non-TTY 출력 환경에서는 fallback (개행 분리 출력)으로 escape sequence를 발산하지 않음
- G4: 기존 `\r  %s` 페어 22+ 호출 사이트를 `tui.ProgressLine`으로 일괄 마이그레이션
- G5: TTY 환경에서 가시 메시지 텍스트 byte-identical 유지 (보이지 않는 escape sequence만 추가)
- G6: 마이그레이션 누락 회귀 차단 — grep 기반 sentinel AC로 영구 0 enforce

### B.2 Non-goals (범위 외)

- NG1: 애니메이션 스피너 라이브러리 도입 (`bubbles/spinner` 등) — 본 SPEC은 무상태(stateless) ProgressLine만 다룬다. 애니메이션은 별도 SPEC 후보
- NG2: Curses/termbox 의존성 추가 — lipgloss + ANSI escape only
- NG3: 색상 테마 전면 개편 (`cliPrimary`, `cliSuccess` 이미 존재; 본 SPEC은 재사용)
- NG4: 진행률 % 표시, multi-line progress bar — 기존 `tui.Progress`/`tui.Stepper`와 별개 추상화

### B.3 Out of Scope (h3)

본 SPEC과 관련은 있으나 별도 SPEC으로 분리:

- **SPEC-V3R6-UPDATE-ARCHIVE-CONTRACT-001** (예정) — `--force` + skip-sync archive 계약
- **SPEC-V3R6-UPDATE-NOISE-001** (예정) — reserved filename + 3-way merge noise 감축
- **Animated spinners** — `bubbles/spinner` 또는 lipgloss 기반 애니메이션 (out of scope; 본 SPEC은 stateless ProgressLine만)
- **Color theme overhaul** — `cliPrimary` / `cliSuccess` 이미 존재하므로 본 SPEC은 재사용만

---

## Section C — Functional Requirements (GEARS notation)

본 섹션의 모든 요구사항은 GEARS notation (Generalized Expression for AI-Ready Specs, SPEC-V3R6-GEARS-MIGRATION-001 캐논)을 사용한다. 레거시 EARS conditional pattern은 회피하고, `Where` (상태) / `When` (이벤트) / `While` (지속 조건) 복합 modality를 사용한다.

### REQ-UPR-001 — ProgressLine API 도입 (Ubiquitous)

The `internal/tui` package **shall** expose a function `ProgressLine(out io.Writer, message string, theme *Theme) *ProgressLineHandle` that initiates a progress line render and returns a handle exposing `Done(successMsg string)` and `Fail(errMsg string)` methods for terminal transition.

**Rationale**: 호출자가 라인 길이 산술이나 escape sequence 처리를 알 필요 없도록 단일 책임 헬퍼를 제공한다. `internal/tui/status.go`의 stateless `Spinner` / `Progress` precedent와 동일 패턴 (theme nil 허용, LightTheme fallback).

### REQ-UPR-002 — ANSI escape 기반 line clear (Ubiquitous)

While `out` is a TTY-backed writer, the `ProgressLine` implementation **shall** prefix every result/failure message with the ANSI escape sequence `\r\033[2K` (carriage return + CSI `2K` — erase entire line) so that the full line buffer is cleared before the new message is rendered.

**Rationale**: `\r` 단독은 커서만 0 컬럼으로 이동시키므로 짧은 메시지가 긴 메시지를 덮어쓰지 못한다. `\033[2K` (CSI Erase in Line, Mode 2)는 라인 전체를 clear하여 트레일링 문자 잔여를 차단한다.

### REQ-UPR-003 — Non-TTY fallback (State-driven)

Where `out` is **not** a terminal (detected via `isatty.IsTerminal` on the underlying file descriptor), When `Done`/`Fail` is invoked, the `ProgressLine` implementation **shall** emit the progress message followed by `\n`, then the result message followed by `\n` — without any `\r` or ANSI escape sequence — so that piped/redirected output (CI logs, `2>&1 | cat`) remains readable as plain text.

**Rationale**: CI runner, `2>&1 | cat`, log aggregator 등 non-terminal 환경에서 ANSI escape는 그대로 텍스트로 남아 가독성을 해친다. Progress + result를 별도 라인으로 분리 출력하여 plain text 환경에서 의도된 의미가 보존되도록 한다. `internal/cli/update.go:23`이 이미 `mattn/go-isatty` 의존성을 사용 중이므로 dependency surface 증가 없음.

### REQ-UPR-004 — 기존 호출 사이트 마이그레이션 (Event-driven)

When the SPEC is implemented, all existing `"\r  %s "` literal patterns in `internal/cli/update.go`, `internal/cli/init.go`, and `internal/cli/update_cleanup.go` **shall** be migrated to use the `tui.ProgressLine` API, such that a grep over `internal/cli/` (excluding `_test.go`) for the literal pattern `'\r  %s'` returns zero matches.

**Rationale**: 부분 마이그레이션은 결함을 부분적으로만 해소하고 일관성을 깨뜨린다. 22+ 사이트 모두 단일 헬퍼로 통합하여 향후 새 사이트 추가 시에도 동일 API를 사용하도록 강제한다.

### REQ-UPR-005 — 가시 메시지 byte-identical 보장 (Ubiquitous)

While running under a TTY, the `ProgressLine.Done`/`Fail` rendered output **shall** preserve the visible message text (after stripping ANSI escape sequences) byte-identical to the pre-migration output, such that user-facing message wording is unchanged and only invisible control sequences are added.

**Rationale**: 본 SPEC은 cosmetic defect 정정으로 사용자 가시 텍스트(메시지 wording)는 보존되어야 한다. `cliSuccess.Render`, `symSuccess()`, message string 모두 그대로 유지하며 escape sequence prefix만 추가한다. 회귀 검증의 baseline.

### REQ-UPR-006 — Test coverage golden 출력 (Ubiquitous)

The `internal/tui` package **shall** include golden-output tests covering both the TTY path (with `\r\033[2K` prefix in the result message) and the non-TTY path (with `\n`-separated progress and result messages), such that `go test -run TestProgressLine` exercises both branches and any future regression is caught by snapshot comparison.

**Rationale**: TTY/non-TTY 분기는 환경 의존이므로 단위 테스트가 양쪽 분기를 모두 강제로 exercise해야 한다. 향후 escape sequence 변경, fallback 정책 수정 시 회귀 차단.

### REQ-UPR-007 — Handle 재사용성 (Optional / Where exists)

Where a long-running operation requires mid-progress message updates, the `ProgressLineHandle` **shall** expose an `Update(message string)` method that re-renders the progress line with a new message while preserving the line-clear semantics defined in REQ-UPR-002 and REQ-UPR-003.

**Rationale**: 일부 호출 사이트 (예: 대용량 백업 진행 중 단계 갱신)는 단일 progress→result 페어를 넘어 중간 상태 갱신이 필요할 수 있다. 본 요구사항은 미래 호출 사이트 확장성을 위한 optional API이며, M1 마이그레이션 단계에서는 사용되지 않을 수 있다.

---

## Section D — Acceptance Criteria

각 AC는 binary (PASS / FAIL) 형태로 검증 명령과 기대 결과를 함께 제공한다. 상세는 [acceptance.md](./acceptance.md) 참조.

| AC ID | 요구사항 | 검증 방식 |
|---|---|---|
| AC-UPR-001 | `tui.ProgressLine` API 존재 (REQ-UPR-001) | godoc + grep `func ProgressLine\b` |
| AC-UPR-002 | TTY path에서 `\r\033[2K` prefix 적용 (REQ-UPR-002) | golden test `TestProgressLine_TTY` |
| AC-UPR-003 | Non-TTY path에서 `\n` 분리 출력 (REQ-UPR-003) | golden test `TestProgressLineNonTTY` |
| AC-UPR-004 | `'\r  %s'` 패턴 zero 마이그레이션 (REQ-UPR-004) | grep count |
| AC-UPR-005 | ProgressLine 호출 사이트 수 일치 (REQ-UPR-004) | grep count |
| AC-UPR-006 | 가시 메시지 byte-identical (REQ-UPR-005) | strip-ansi + diff |
| AC-UPR-007 | golden tests 통과 (REQ-UPR-006) | `go test -run TestProgressLine` |
| AC-UPR-008 | Smoke test (실제 `moai update`) (REQ-UPR-002 통합) | `moai update --yes 2>&1 \| cat \| grep -E '[a-z]\.{3}'` |
| AC-UPR-009 | `Update` method 동작 (REQ-UPR-007) | unit test `TestProgressLine_Update` (optional path) |

---

## Section E — Risks and Plan

### E.1 Risks

| Risk ID | Description | Likelihood | Impact | Mitigation |
|---|---|---|---|---|
| R1 | ANSI escape `\033[2K`가 legacy Windows cmd.exe / 비-VT 터미널에서 미지원 | Low | Low (cosmetic) | lipgloss는 이미 Windows ANSI 지원을 feature-detect; `runtime.GOOS == "windows"`에서도 modern Windows Terminal / PowerShell 7+는 VT 모드 default. Worst case는 결함 이전 상태 (잔여 문자) — regression 아님 |
| R2 | Non-TTY 출력 line count가 기존 대비 증가 (라인 수 증가 → 로그 verbosity 증가) | Medium | Low | REQ-UPR-003에서 명시적 Plain mode 정의 — progress + result 각각 1 line × N 호출. 기존 TTY 모드는 1 line 유지. CI 로그 grep parser는 line-oriented가 아니라 keyword-oriented이므로 영향 없음 |
| R3 | 22+ 호출 사이트 마이그레이션 중 일부 누락 → 결함이 잔존 사이트에서 재발 | Medium | Medium | AC-UPR-004 (grep count = 0) sentinel로 CI/local에서 영구 enforce; M1 단계별 wave 마이그레이션 + 각 wave 후 grep 검증 |
| R4 | `tui` 패키지가 `mattn/go-isatty` 의존성을 신규 import → cyclic / layering 위험 | Low | Low | `internal/cli/update.go:23`이 이미 `mattn/go-isatty`를 사용 중이며, `tui` → `isatty`는 단방향 (cli → tui → isatty). `go-isatty`는 stdlib 가까운 저수준 라이브러리이므로 layering 위반 없음 |

### E.2 Plan (요약)

전체 milestone과 task 분해는 [plan.md](./plan.md) 참조. 요약:

- **M0**: `tui.ProgressLine` API 설계 (signature, handle, theme integration) + golden test 작성
- **M1**: `internal/cli/update.go` 22 사이트 마이그레이션 + AC-UPR-004 grep zero 검증
- **M2**: `internal/cli/init.go` / `internal/cli/update_cleanup.go` 잔여 사이트 마이그레이션
- **M3**: smoke test (실제 `moai update` 실행으로 AC-UPR-008 검증) + status 갱신 (`status: draft → implemented`)

Tier S 작업으로 단일 PR/main 직진 가능 (Hybrid Trunk doctrine, CLAUDE.local.md §23).

---

## Section F — References

- `internal/tui/status.go` — stateless `Spinner` / `Progress` / `Stepper` precedent. 본 SPEC `ProgressLine`도 동일 stateless pattern을 따른다.
- `internal/cli/update.go:51-57` — lipgloss `AdaptiveColor` 사용 패턴 (`cliPrimary`, `cliSuccess` 등). `ProgressLine` 내부에서도 동일 스타일 재사용.
- `internal/cli/update.go:23` — `mattn/go-isatty` 의존성 기존 사용처. `tui.ProgressLine`이 isatty를 직접 사용하더라도 dependency surface 증가 없음.
- `.claude/rules/moai/workflow/spec-workflow.md` — Tier S SPEC 작성 doctrine.
- `.claude/rules/moai/languages/go.md` — Go 1.23+ tooling, error wrapping, golangci-lint enforcement.
- `SPEC-V3R6-GEARS-MIGRATION-001` (merged 2026-05-22 `134a43fac`) — GEARS notation 캐논; 본 SPEC은 IF/THEN 패턴 회피 + Where/When/While 사용.

---

## HISTORY

- **2026-05-23 v0.1.0 draft** — 초기 작성. 2026-05-23 `moai update` 감사 세션 Defect #5 cosmetic 정정. `internal/cli/` 22+ `\r  %s` 페어 출력 corruption 보고서 기반. GEARS notation 적용 (IF/THEN 0건). 7 REQ + 9 AC + 4 Risk. Tier S 분류.
