# SPEC-V3R3-CLI-TUI-001 Implementation Plan

> **Goal**: 모두의AI 디자인 v2를 `internal/tui/` 패키지로 incarnate 하고 14개 명령어 출력을 마이그레이션
> **Methodology** (committed per milestone — `quality.yaml` `development_mode`와 무관):
> - **TDD**: M1, M2, M7 (greenfield: `internal/tui/` 패키지 신규, golden snapshot suite)
> - **DDD**: M3, M4, M5, M6 (legacy migration: 기존 `banner.go`/`doctor.go`/`init.go` 등 동작 보존 → characterization test → migrate)
> **Branch**: `feat/SPEC-V3R3-CLI-TUI-001` (현재 worktree `cli-tui-v2`)
> **Total Milestones**: 7 (M1~M7) — priority labels only, no time estimates

---

## 1. 마일스톤 개요

| Milestone | Priority | Methodology      | 의존성  | 산출물 요약                                                |
|-----------|----------|------------------|---------|------------------------------------------------------------|
| M1        | High     | TDD              | -       | `internal/tui/` 패키지 골격 + theme.go + box.go + pill.go    |
| M2        | High     | TDD              | M1      | status.go + form.go + table.go + prompt.go + term.go + help.go |
| M3        | High     | DDD characterize | M2      | `banner.go` 마이그레이션 (terra cotta → deep teal)            |
| M4        | High     | DDD              | M3      | doctor / status / version / update 4-command batch          |
| M5        | High     | DDD              | M4      | `init` huh wizard 강화 (Stepper + RadioRow)                  |
| M6        | Medium   | DDD              | M5      | cc / loop / statusline / help / error 5-command batch       |
| M7        | Medium   | TDD              | M6      | dark/light auto-detect + NO_COLOR + golden snapshot suite    |

[HARD-PLAN-1] M1~M7는 sequential 진행. 병렬 worktree 사용 금지 — 각 milestone이 직전 milestone의 API 출력을 검증해야 하기 때문.

---

## 2. M1 (Priority: High, TDD): TUI 패키지 골격 + Core Tokens

### 2.1 산출물

- `internal/tui/theme.go` (~200 LOC) — 라이트/다크 토큰 struct 두 개
- `internal/tui/theme_test.go` — 토큰 값이 디자인 소스와 1:1 매칭 검증
- `internal/tui/box.go` (~120 LOC) — `Box(opts)` / `ThickBox(opts)` 함수 두 개
- `internal/tui/box_test.go` — golden snapshot (라이트/다크 × 4 케이스 = 8 files)
- `internal/tui/pill.go` (~80 LOC) — `Pill(kind, solid, label)` 함수
- `internal/tui/pill_test.go` — 6 variants × 2 themes × 2 (solid/outline) = 24 snapshots
- `internal/tui/doc.go` — 패키지 godoc + 디자인 소스 출처 표기

### 2.2 TDD 사이클

RED phase:
1. `theme_test.go`에 `TestLightTokens` / `TestDarkTokens` 작성 — 디자인 소스 `tui.jsx:9-69`의 28개 키 (chrome/bg/fg/accent/...)별 정확한 hex/rgba 값 비교
2. `box_test.go`에 `TestBoxBasic` / `TestThickBoxAccent` 작성 — `testdata/box-light-basic.golden`, `testdata/box-dark-accent.golden` 등 8개 snapshot
3. `pill_test.go`에 `TestPillVariants` 작성 — 6 kind × 2 theme × 2 solid = 24 snapshots
4. 모두 컴파일 실패 / 함수 부재로 실패 확인

GREEN phase:
5. `theme.go` 작성 — `LightTheme()` / `DarkTheme()` exported 함수, 내부적으로 `Theme` struct return
6. `box.go` 작성 — `lipgloss.NewStyle().Border(...).Padding(...).Foreground(...)`로 Box() 구현
7. `pill.go` 작성 — `lipgloss` Background/Foreground 조합

REFACTOR phase:
8. 토큰 키 중복 정리 (예: `accentSoft` vs `accentSofter`)
9. `internal/tui/internal/runeguard.go` — 한글 폭 보정 헬퍼 (mattn/go-runewidth wrapper)

### 2.3 검증 명령

```bash
cd "$(git rev-parse --show-toplevel)"
go test ./internal/tui/... -run "TestLight|TestDark|TestBox|TestPill" -v
go vet ./internal/tui/...
golangci-lint run ./internal/tui/...
```

### 2.4 종료 조건 (DOD)

- [ ] `go test ./internal/tui/...` GREEN (theme_test + box_test + pill_test)
- [ ] 디자인 토큰 28키 모두 `tui.jsx:9-69`와 1:1 매칭 (REQ-CLI-TUI-002)
- [ ] golden snapshot 8개 (box) + 24개 (pill) commit
- [ ] godoc에 디자인 소스 출처(`source/project/tui.jsx`) 명시

---

## 3. M2 (Priority: High, TDD): Status / Form / Table / Prompt / Term / Help Primitives

### 3.1 산출물

- `internal/tui/status.go` — `StatusIcon(kind)`, `Spinner(label)`, `Progress(value, max, opts)`, `Stepper(current, total)`
- `internal/tui/form.go` — `RadioRow(opts)`, `CheckRow(opts)` (huh 폼 헬퍼)
- `internal/tui/table.go` — `KV(key, value, opts)`, `CheckLine(status, label, value, hint)`, `Section(title, opts)`
- `internal/tui/prompt.go` — `Prompt(host, path, branch, dirty, cmd)`, `Cursor()`
- `internal/tui/term.go` — `Term(title, width, body, opts)` 윈도우 chrome (스크린샷 전용; 실제 출력은 no-op 모드)
- `internal/tui/help.go` — `HelpBar(items []KeyHint)`
- 각각 `_test.go` golden snapshot

### 3.2 TDD 사이클 요약

- RED: 6개 새 컴포넌트별 golden snapshot 테스트 (총 ~36 snapshots: 6 컴포넌트 × 2 테마 × 3 케이스)
- GREEN: lipgloss API 매핑
- REFACTOR: 공통 헬퍼 (`renderCard`, `renderRow`) 추출

### 3.3 종료 조건

- [ ] `tui.Spinner` / `tui.Progress`는 stateless 렌더링 (현재 ANSI string 반환만; live update는 호출자 책임)
- [ ] `tui.Term`은 `MOAI_SCREENSHOT=1` 환경변수에서만 macOS-style traffic-light chrome 출력 (실제 터미널에선 빈 출력)
- [ ] golden snapshot 36+ commit

---

## 4. M3 (Priority: High, DDD): banner.go 마이그레이션

### 4.1 ANALYZE phase

기존 `internal/cli/banner.go` 동작 분석:
- `PrintBanner(version)`: 8줄 ASCII art + terra cotta 색 + 설명 + 버전 출력
- `PrintWelcomeMessage()`: 보라색 환영 메시지 (init wizard 시작점)

호출 지점 grep: `grep -rn "PrintBanner\|PrintWelcomeMessage" internal/`
→ 4개 진입점: init, update, version, doctor

### 4.2 PRESERVE phase (characterization test)

신규 `internal/cli/banner_test.go`에 현재 출력을 golden capture:
1. `TestBanner_Current_Light` (NO_COLOR=0, MOAI_THEME=light)
2. `TestBanner_Current_Dark` (NO_COLOR=0, MOAI_THEME=dark)
3. `TestBanner_NoColor` (NO_COLOR=1)
4. `TestWelcome_Current_Light/Dark`

이 단계의 snapshot은 **마이그레이션 후 폐기 예정** (terra cotta 출력이라 디자인 소스와 다름). 단, 회귀 검증을 위해 M3 시작 시점에 capture 후 git stash.

### 4.3 IMPROVE phase

1. `internal/cli/banner.go` 재작성:
   - terra cotta 색 (`#C45A3C/#DA7756`) 제거 (REQ-CLI-TUI-013 위반 정정)
   - `internal/tui/theme.go`의 `Theme().Accent` 사용
   - 8줄 ASCII art는 그대로 유지 (§2.2 Non-Goal)
   - 배너 하단에 `tui.Pill` 3개 (`v3.2.4`, `go 1.26`, `claude 1.0.18`) 추가 → 디자인 소스 `screens.jsx:ScreenBanner` 매칭
2. `PrintWelcomeMessage()`도 보라색(`#5B21B6/#7C3AED`) 제거 → `Theme().Accent`로 통일
3. 새 golden snapshot으로 교체

### 4.4 종료 조건

- [ ] `internal/cli/banner.go`에서 hex 색 코드 0건 (`grep -E '#[0-9a-fA-F]{6}' internal/cli/banner.go` empty)
- [ ] `go test ./internal/cli/... -run TestBanner` GREEN
- [ ] 라이트/다크 양쪽에서 시각 회귀 0건 (M2 golden snapshot 활용)

---

## 5. M4 (Priority: High, DDD): doctor / status / version / update 4-command Batch

### 5.1 영향받는 파일 (Multi-File 분해 대상)

[HARD-PLAN-2] CLAUDE.md §1 Rule 2 (Multi-File Decomposition) 적용. 6+ 파일 변경이므로 4 step으로 쪼갠다.

| Step | 파일                                            | 디자인 소스                              |
|------|-------------------------------------------------|------------------------------------------|
| 4a   | `internal/cli/version.go` (or `pkg/version/`)    | `screens.jsx:ScreenVersion`               |
| 4b   | `internal/cli/doctor*.go` (`doctor.go`, `doctor_config.go`, `doctor_brain.go` 등 6 파일) | `screens.jsx:ScreenDoctor`                |
| 4c   | `internal/cli/status.go` 등                       | `screens.jsx:ScreenStatus`                |
| 4d   | `internal/cli/update.go`                          | `screens.jsx:ScreenUpdate`                |

### 5.2 DDD 사이클 (각 step 동일 패턴)

ANALYZE: 현재 출력 capture (golden snapshot, 라이트/다크/no-color × 3)
PRESERVE: 동작 보존 테스트 추가 (예: `--json` 플래그 출력 변하지 않음)
IMPROVE: `internal/tui/` API 호출로 교체

### 5.3 step 4b (doctor) 세부

- `tui.CheckLine(status, label, value, hint)` 호출로 19개 검사 항목 그룹화:
  - "시스템" 그룹: Go / Git / Claude Code / GitHub CLI (4 항목)
  - "MoAI-ADK" 그룹: config.yaml / 훅·슬래시 명령 / MX 앵커 / Glamour 캐시 / 텔레메트리 (5 항목)
  - 추가 10+ 항목은 현재 doctor가 실제 검사하는 내용 기반
- `tui.Box("요약", accent=true)` 하단 박스에 `tui.Pill` 3개 (통과/주의/실패)
- 디자인 소스: `screens.jsx:ScreenDoctor`

### 5.4 종료 조건

- [ ] 4 step 모두 golden snapshot diff 검증 통과 (라이트/다크/no-color)
- [ ] 4 step 모두 hex 색 코드 0건
- [ ] `--json` / `--quiet` 등 기존 플래그 출력은 변경 없음 (PRESERVE)
- [ ] Multi-File Decomposition: TodoList로 4a→4b→4c→4d 순차 진행

---

## 6. M5 (Priority: High, DDD): init huh Wizard 강화

### 6.1 영향받는 파일

- `internal/cli/init.go`
- `internal/cli/wizard/wizard.go` (huh 사용 핵심)
- `internal/cli/profile_setup.go` (init 흐름의 일부)

### 6.2 DDD 사이클

ANALYZE: 현재 init wizard의 단계 수와 huh 폼 구성 확인
- 디자인 소스 `screens.jsx:ScreenInit` 기준: 6단계 wizard, step 3은 "언어 · 런타임" 선택 (Go/TS/Python/Rust/Mono)

PRESERVE: wizard 진행 / 취소 / 검증 동작 그대로

IMPROVE:
1. wizard 상단에 `tui.Stepper(current, total=6)` 추가
2. huh `Select`/`Form`의 라디오 prefix가 `◆`(선택) / `◇`(미선택) 강제 — huh 라이브러리 옵션 또는 custom theme 적용
3. step 헤더는 `tui.Box(title, accent=true)`로 감싸기
4. 각 step 하단에 keyboard hint footer (`tui.HelpBar`)
5. 정보성 박스 (`info` 카드)는 `tui.Box(border=dashed)` + `tui.StatusIcon("info")`

### 6.3 종료 조건

- [ ] init wizard 6단계 모두 `tui.Stepper` 표시
- [ ] 라디오 prefix `◆/◇` 적용 (이모지 0건 검증)
- [ ] 기존 wizard 검증 로직 (project name 유효성 등) 회귀 0건
- [ ] golden snapshot: 각 step의 라이트/다크 × 6 step = 12 snapshots

---

## 7. M6 (Priority: Medium, DDD): cc / loop / statusline / help / error 5-command Batch

### 7.1 영향받는 파일

| 명령어            | 파일                              | 디자인 소스                                  |
|-------------------|-----------------------------------|----------------------------------------------|
| `moai cc`         | `internal/cli/cc.go`               | `screens.jsx:ScreenCc`                        |
| `moai cg/glm`     | `internal/cli/cg.go`, `glm.go`     | (cc와 같은 패밀리)                            |
| `moai loop start` | `internal/cli/loop_start.go` 등    | `screens.jsx:ScreenLoopStart` (4-phase strip) |
| `moai statusline` | `internal/statusline/*.go` 다수    | `screens.jsx:ScreenStatusline`                |
| `moai help`       | cobra root.go의 `Long` template     | `screens.jsx:ScreenHelp`                      |
| 에러 출력 (전역)  | `internal/cli/render.go` + 분산     | `screens.jsx:ScreenError`                     |

### 7.2 핵심 변경

- `loop start`: 4-phase Stepper (`Spec → Plan → Impl → Sync`) + 페이즈별 `tui.Box`
- `statusline`: 기존 `internal/statusline/theme.go`를 `internal/tui/theme.go`의 thin wrapper로 재구성 (R-07 위험 대응)
- `error`: 전역 `RenderError(err)` 함수 신설, `tui.ThickBox(color=danger) + tui.StatusIcon("err")`로 통일
- `help`: cobra `SetUsageTemplate`을 `tui` 컴포넌트 호출로 교체 — `tui.Section` (4 그룹) + `tui.HelpBar`

### 7.3 종료 조건

- [ ] 5 명령어 + 1 전역 에러 출력 모두 `internal/tui/` API 호출만
- [ ] `internal/statusline/`는 `internal/tui/theme.go`만 import (자체 색 상수 0건)
- [ ] golden snapshot 라이트/다크 × 6 surface = 12+ snapshots

---

## 8. M7 (Priority: Medium, TDD): Auto-Detect + NO_COLOR + Golden Snapshot Suite

### 8.1 산출물

- `internal/tui/detect.go` — 테마 감지 우선순위 결정 함수 `Resolve(env)` (NO_COLOR > MOAI_THEME > HasDarkBackground > default-dark)
- `internal/tui/detect_test.go` — 4 우선순위 케이스별 단위 테스트
- `internal/tui/profile.go` — `colorprofile.Detect()` 활용한 truecolor/256/16/no-color fallback
- `internal/tui/profile_test.go` — 4 ColorProfile × 2 theme = 8 케이스
- `internal/tui/golden/` 디렉토리 — M1-M6 누적 snapshot 통합 인덱스
- `internal/tui/i18n.go` (골격만) — `messages/ko.yaml`, `messages/en.yaml` 두 카탈로그 + lookup 함수
- `Makefile` 타겟 추가: `make tui-snapshot` — 모든 golden snapshot 재생성 + diff

### 8.2 TDD 사이클

RED:
1. NO_COLOR=1 시 `Theme()` 호출 결과가 모든 색 키 = `lipgloss.NoColor{}` 검증
2. `MOAI_THEME=invalid` 시 default-dark fallback
3. 한글 + 영문 mixed 한 줄 18종 정렬 검증 (`runewidth.StringWidth` 적용 확인)
4. ko/en 메시지 카탈로그 동일 키 셋 검증

GREEN:
5. 위 4 테스트가 통과하도록 detect.go / profile.go / i18n.go 구현

### 8.3 종료 조건 (전체 SPEC)

- [ ] 7 milestone 모두 GREEN
- [ ] `internal/tui/` 패키지 외부에서 hex 색 코드 0건 (REQ-CLI-TUI-013)
- [ ] `internal/tui/` 패키지 외부에서 이모지 0건 (REQ-CLI-TUI-014)
- [ ] 14개 명령어 surface 모두 `internal/tui/` API 호출만
- [ ] go.mod 신규 require 0건 (REQ-CLI-TUI-008)
- [ ] golden snapshot 100+ files (라이트/다크/no-color × 19개 화면 + variant)
- [ ] 16개 언어 i18n 메시지 카탈로그 골격 + ko/en 두 카탈로그 채움
- [ ] `acceptance.md` AC-CLI-TUI-001 ~ 010 모두 검증

---

## 9. 기술적 접근 (Technical Approach)

### 9.1 lipgloss API 매핑 전략

설계 소스 `tui.jsx`의 React/CSS 컴포넌트 → lipgloss Style 1:1 변환표:

| jsx CSS 속성                              | lipgloss API                                              |
|-------------------------------------------|-----------------------------------------------------------|
| `border: 1px solid`                        | `lipgloss.NewStyle().Border(lipgloss.RoundedBorder())`     |
| `border-radius: 8`                         | (lipgloss 기본 RoundedBorder 사용)                          |
| `padding: 12px 16px`                       | `.Padding(1, 2)` (cell 단위 변환: 1 line + 2 col)           |
| `color: t.accent`                          | `.Foreground(lipgloss.Color("#144a46"))` (또는 AdaptiveColor)|
| `background: t.accentSofter`               | `.Background(lipgloss.Color("..."))` 또는 dim faint 패턴    |
| `font-weight: 700`                         | `.Bold(true)`                                              |
| `letter-spacing: -0.025em`                 | (터미널 미지원, 무시)                                       |
| `font-family: JetBrains Mono`              | (터미널 위임, lipgloss 메타 없음)                            |
| display: flex; gap: 8                      | `lipgloss.JoinHorizontal(lipgloss.Top, parts...)` + 패딩    |
| `display: grid; grid-template: 1fr 1fr`    | 좌우 width 계산 후 `lipgloss.JoinHorizontal`               |

### 9.2 한글 폭 정렬 전략

- 모든 박스 width 계산은 `runewidth.StringWidth(s)` 기반
- `lipgloss.Width(rendered)` 검증 후 `Padding`/`Width`로 수동 보정
- golden test에서 한글 + 영문 mixed 18 케이스 강제 검증 (예: "환경 감지 Layer 1 · 환경 변수 우선")

### 9.3 NO_COLOR 통합

```go
// internal/tui/detect.go (의사코드)
func Resolve(env Env) Theme {
    if env.NoColor() { return MonochromeTheme() }  // 모든 색 = lipgloss.NoColor{}
    if t := env.MoaiTheme(); t == "light" { return LightTheme() }
    if t := env.MoaiTheme(); t == "dark"  { return DarkTheme() }
    if lipgloss.HasDarkBackground() { return DarkTheme() }
    return DarkTheme() // safe default
}
```

### 9.4 16-Language Neutrality

`internal/tui/messages/` 디렉토리 골격:
```
messages/
  ko.yaml   # 한국어 (M7에서 채움 — 디자인 소스 카피 기반)
  en.yaml   # 영어 (M7에서 채움 — fallback)
  zh.yaml   # 중국어 (TBD by 별도 SPEC)
  ja.yaml   # 일본어 (TBD by 별도 SPEC)
  ...       # 16개 언어 (12개는 TBD)
```

`tui.Translate(key, lang)` lookup만 본 SPEC scope. BCP47 fallback chain은 별도 SPEC.

---

## 10. Multi-File Decomposition Strategy

CLAUDE.md §1 Rule 2 (3+ 파일 수정 시 분해 강제) 준수:

| Milestone | 파일 수 | 분해 단위                                                     |
|-----------|---------|--------------------------------------------------------------|
| M1        | 6       | TodoList: theme → box → pill → tests (4 step)                |
| M2        | 12      | TodoList: status → form → table → prompt → term → help (6 step) |
| M3        | 2       | TodoList: characterization → migration (2 step)              |
| M4        | 9       | TodoList: 4a (version) → 4b (doctor 6 파일) → 4c (status) → 4d (update) (4 step) |
| M5        | 3       | TodoList: stepper → wizard → profile_setup (3 step)          |
| M6        | 8+      | TodoList: cc → cg/glm → loop → statusline → help → error (6 step) |
| M7        | 6       | TodoList: detect → profile → i18n golden suite (3 step)      |

각 step 완료 시 git commit + golden snapshot diff 검증 필수.

---

## 11. 검증 체크리스트 (모든 milestone 공통)

각 milestone PR 머지 전:
- [ ] `go test ./internal/tui/... ./internal/cli/...` GREEN
- [ ] `go vet ./...` clean
- [ ] `golangci-lint run ./...` clean
- [ ] 새/수정된 hex 색 코드는 모두 `internal/tui/theme.go`에서만 정의
- [ ] golden snapshot diff PR description에 attach
- [ ] `make build` 성공 (template 변경이 없음을 verify; 본 SPEC은 internal/ 만 수정)
- [ ] CLAUDE.local.md §15 (16-language neutrality) 위반 0건
- [ ] CLAUDE.local.md §14 (하드코딩 방지) 위반 0건

---

## 12. Open Questions (run phase에서 결정)

1. **`internal/statusline/theme.go` 통합 vs 유지**: M6에서 thin wrapper로 재구성하기로 했으나, `statusline`이 zsh/tmux에 출력되는 형식(no-trailing-newline 등) 차이로 별도 ANSI 처리 필요할 가능성. 실제 작업 시 결정.
2. **`huh` 라이브러리의 라디오 prefix 커스터마이징 범위**: huh `v0.8.0`은 기본 prefix가 `> `/`  `이므로 `◆/◇` 강제는 custom Theme 작성 필요. huh 0.8 API 한계로 불가능 시 wrapper로 직접 라디오 그리기.
3. **Glamour 캐시 표시**: 디자인 소스 `screens.jsx:ScreenDoctor`에서 "Glamour 캐시 갱신 필요" warn 항목 존재. glamour를 본 SPEC에서 도입하지 않으므로 이 항목은 doctor 화면에서 conditional skip 또는 placeholder 처리.
4. **bubbles indirect → direct 승격 시점**: M2 status.go에서 Spinner/Progress 시각 언어 차용 시 `bubbles/spinner`, `bubbles/progress` 직접 import 필요한지, 아니면 lipgloss만으로 충분한지 검증 필요.
5. **Windows cmd.exe legacy 환경의 fallback 검증**: REQ-CLI-TUI-011 충족을 위해 실제 Windows runner가 필요. CI windows-latest matrix가 cmd.exe도 cover 하는지 확인.

위 5개는 plan-auditor 1차 audit 후 acceptance.md 업데이트 또는 별도 follow-up SPEC으로 분리 결정.

---

## 13. Risk Mitigation Cross-Reference

spec.md §8 Risks와 plan.md milestone 대응:

| Risk ID | Mitigation Milestone           |
|---------|--------------------------------|
| R-01    | M1 (golden test 18 mixed cases) + M7 (suite 통합)             |
| R-02    | M7 (`profile.go` colorprofile.Detect)                         |
| R-03    | M3-M6 (각 ddd cycle의 PRESERVE phase)                          |
| R-04    | M1-M2 (19 화면 mock-up 검증 후 API 동결)                        |
| R-05    | M7 (`messages/` 골격 + ko/en)                                  |
| R-06    | M7 (`detect.go` MOAI_THEME override)                          |
| R-07    | M6 (`internal/statusline/theme.go` thin wrapper로 재구성)       |
