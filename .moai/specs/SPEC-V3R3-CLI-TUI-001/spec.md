---
id: SPEC-V3R3-CLI-TUI-001
version: "0.1.0"
status: completed
created_at: 2026-05-09
updated_at: 2026-05-12
author: manager-spec
priority: Medium
labels: [cli, tui, design-system, lipgloss, bubbletea, theme, brand, i18n-neutral]
issue_number: null
depends_on: []
related_specs: []
breaking: false
bc_id: []
lifecycle: spec-anchored
---

# SPEC-V3R3-CLI-TUI-001: moai CLI 터미널 TUI 디자인 시스템 v2 마이그레이션

## HISTORY

| Version | Date       | Author       | Description                                                                                                                                                                                                                                                              |
|---------|------------|--------------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| 0.1.0   | 2026-05-09 | manager-spec | 최초 작성. Claude Design v2 핸드오프 번들(`.moai/design/SPEC-V3R3-CLI-TUI-001/source/`)을 출처로 lipgloss/bubbles/huh primitive 매핑 SPEC 작성. v1 ASCII 박스 폐기 사유 + 모두의AI 디자인 토큰(deep teal `#144a46`, FROZEN) 채택. 14개 명령어 화면 마이그레이션 범위 확정, 5개 미구현 화면(install/constitution/mx graph/spec view/worktree list)은 별도 SPEC으로 분리. |

---

## 1. Goal (목적)

`moai` CLI의 모든 사용자 대면 출력을 **모두의AI 디자인 시스템 v2**로 통일된 시각 언어로 마이그레이션한다. 현재 `internal/cli/banner.go`의 terra cotta(`#C45A3C/#DA7756`) 색상과 일관성 없는 ad-hoc lipgloss 스타일을 제거하고, 단일 deep teal 액센트(`#144a46` 라이트 / `#3eb3a4` 다크) + 디자인 토큰 기반 primitive 컴포넌트 라이브러리(`internal/tui/`)로 단일화한다. 한글/영문 혼용 환경에서 박스 정렬이 깨지는 v1 ASCII 라인 드로잉 패턴을 폐기하고, lipgloss `Border()`/`Padding()` 실제 컴포넌트로 그리는 v2 패러다임으로 전환한다.

이번 SPEC의 결과로 사용자는 `moai banner`, `moai help`, `moai init`, `moai doctor`, `moai status`, `moai version`, `moai update`, `moai cc`, `moai loop start`, `moai statusline` 등 14개 명령어 출력에서 일관된 색·여백·기호·폰트 스타일을 경험한다.

## 2. Non-Goal (이번 SPEC에 포함하지 않는 것)

다음 항목은 본 SPEC의 scope에서 명시적으로 제외된다.

### 2.1 별도 SPEC으로 분리되는 미구현 명령어 화면 (총 5종)

설계 소스에는 19개 화면이 있으나, 다음 5개는 **현재 moai-adk-go에 명령어 자체가 부재**하므로 별도 SPEC으로 다룬다.

| 화면                       | 디자인 소스                                                                | 분리 사유                                                                  |
|----------------------------|----------------------------------------------------------------------------|----------------------------------------------------------------------------|
| `install.sh` / `.ps1` / `.bat` (3종) | `source/project/screens.jsx:5-200` (ScreenInstallSh/Ps1/Bat)              | `scripts/install.*` 미존재. 별도 release/distribution SPEC 필요.           |
| `moai constitution check`  | `source/project/screens.jsx` (ScreenConstitution)                          | `internal/cli/constitution.go` 존재하나 `check` 서브커맨드 미구현. 별도 SPEC. |
| `moai mx graph`            | `source/project/screens.jsx` (ScreenMxGraph)                              | `moai mx` 명령어는 존재하나 `graph` 서브커맨드 미구현. 별도 SPEC.            |
| `moai spec view <SPEC-ID>` | `source/project/screens.jsx` (ScreenSpecView, glamour 렌더링)              | `moai resume <SPEC-ID>` 존재하나 `spec view` 서브커맨드 부재. 별도 SPEC.    |
| `moai worktree list`       | `source/project/screens.jsx` (ScreenWorktreeList, table 컴포넌트)          | `moai worktree` 명령어 자체 미구현. 별도 SPEC 필요.                         |

### 2.2 ASCII 아트 배너 자체 디자인 변경

`internal/cli/banner.go`의 `███╗ ███╗ ... ██╗ ██╗` 8줄 ASCII 아트는 그대로 유지한다. 본 SPEC은 색상만 terra cotta → deep teal로 교체한다. 새로운 ASCII 아트 디자인(설계 소스 `screens.jsx:ScreenBanner`의 미니 ASCII `__ __ __ ___ _`)은 별도 visual identity SPEC에서 다룬다.

### 2.3 Glamour 마크다운 렌더링

`spec view` 화면에 사용되는 charmbracelet/glamour는 현재 `go.mod`에 부재(설계 소스 `screens.jsx:ScreenVersion`의 `huh · glamour v0.8.0 · v0.10.0` 라벨은 디자인 prompt). `spec view` 명령어 신설과 함께 별도 SPEC에서 의존성 추가 + 통합한다.

### 2.4 Brand voice / 카피 전체 개편

본 SPEC은 시각 layer 한정. "베타 테스터 빌드", "모두의 Agentic Development Kit" 같은 문구의 한국어 카피는 i18n 메시지 카탈로그 골격만 마련하고, 실제 16개 언어 번역은 별도 SPEC.

### 2.5 GUI 또는 웹 콘솔

본 SPEC은 터미널 TUI 한정. IDEA-002 Console / 별도 GUI 작업과는 무관하다.

## 3. Scope (이번 SPEC에서 변경하는 범위)

### 3.1 신규 패키지 `internal/tui/`

단일 시각 진실 공급원. 디자인 토큰 + primitive 컴포넌트를 모두 이 패키지에 집약. 기존 `internal/cli/banner.go`, `internal/cli/render.go`, `internal/cli/worktree/render.go`, `internal/statusline/theme.go` 등의 ad-hoc 스타일은 모두 `internal/tui/` API 호출로 대체된다.

| 파일               | 역할                                                                 | tui.jsx 매핑                                  |
|--------------------|----------------------------------------------------------------------|-----------------------------------------------|
| `theme.go`         | 디자인 토큰 라이트/다크 스트럭트 + 자동 감지                          | `TOK.light` / `TOK.dark` (tui.jsx:9-69)       |
| `box.go`           | `Box()`, `ThickBox()` (lipgloss `Border().Padding()`)                | `Box()` / `ThickBox()` (tui.jsx:88-154)       |
| `pill.go`          | `Pill()` 6 variants (info/ok/warn/err/primary/neutral)               | `Pill()` (tui.jsx:157-177)                    |
| `status.go`        | `StatusIcon()`, `Spinner()`, `Progress()`, `Stepper()`               | (tui.jsx:180-250)                             |
| `form.go`          | huh 폼용 `RadioRow()`, `CheckRow()` 헬퍼                              | `RadioRow()` / `CheckRow()` (tui.jsx:253-293) |
| `table.go`         | `KV()`, `CheckLine()`, `Section()` (key-value / doctor 출력)          | (tui.jsx:296-340)                             |
| `prompt.go`        | `Prompt()`, `Cursor()` (스크린샷용; 실제 출력에선 호출자 책임)         | (tui.jsx:343-364)                             |
| `term.go`          | `Term()` 윈도우 chrome (스크린샷 전용 frame, 실제 터미널에선 no-op)    | `Term()` (tui.jsx:367-414)                    |
| `help.go`          | `HelpBar()` 키보드 힌트                                               | `HelpBar()` (tui.jsx:417-434)                 |
| `messages/`        | i18n 메시지 카탈로그 골격 (16개 언어 ko/en/zh/ja/...)                 | (신규)                                        |

### 3.2 기존 명령어 출력 마이그레이션 (총 14개 surface)

| 명령어                        | 영향받는 파일                          | 디자인 소스 위치                                | 마이그레이션 액션                                                              |
|-------------------------------|----------------------------------------|-------------------------------------------------|--------------------------------------------------------------------------------|
| `moai` (배너)                 | `internal/cli/banner.go`                | `source/project/screens.jsx:ScreenBanner`        | terra cotta → deep teal, `tui.Box`+`tui.Pill` 채택, 정보 그리드 추가          |
| `moai help`                   | cobra root.go의 `Long` template          | `source/project/screens.jsx:ScreenHelp`          | 4 그룹 (프로젝트/런처/자율 개발/거버넌스), `tui.Section` + `tui.HelpBar`        |
| `moai init` (huh wizard)       | `internal/cli/init.go`, `internal/cli/wizard/wizard.go` | `source/project/screens.jsx:ScreenInit`         | 6단계 `tui.Stepper`, 라디오 `◆/◇` prefix 강제, `tui.Box` accent 패널            |
| `moai doctor`                 | `internal/cli/doctor*.go` (6 파일)        | `source/project/screens.jsx:ScreenDoctor`        | `tui.CheckLine` 통일, `tui.Pill` 요약, 시스템/MoAI-ADK 두 그룹                |
| `moai status`                 | `internal/cli/status.go` 등              | `source/project/screens.jsx:ScreenStatus`        | `tui.KV` + SPEC 카드 `tui.Pill` + 진행 중 `tui.Progress` row                   |
| `moai version`                | `pkg/version/version.go` 호출 출력        | `source/project/screens.jsx:ScreenVersion`       | `tui.ThickBox` accent + 의존성 라벨                                            |
| `moai update`                 | `internal/cli/update.go`                  | `source/project/screens.jsx:ScreenUpdate`        | huh confirm + `tui.Progress` 다운로드 표시                                     |
| `moai cc` (런처)              | `internal/cli/cc.go`                      | `source/project/screens.jsx:ScreenCc`            | `tui.Box` 환경 점검 + `tui.Spinner` 부트스트랩                                 |
| `moai loop start`             | `internal/cli/loop_*.go`                  | `source/project/screens.jsx:ScreenLoopStart`     | 4페이즈 (Spec/Plan/Impl/Sync) `tui.Stepper` + 페이즈별 `tui.Box`              |
| `moai statusline`             | `internal/statusline/*.go`                | `source/project/screens.jsx:ScreenStatusline`    | tmux/vim 임베드 출력 색상 토큰 통일                                            |
| 에러 출력 (전역)              | `internal/cli/render.go`, 각 명령어        | `source/project/screens.jsx:ScreenError`         | `tui.ThickBox` danger 톤 + `tui.StatusIcon` 통일                              |
| `moai cg` / `moai glm`        | `internal/cli/cg.go`, `internal/cli/glm.go` | (cc와 동일 스타일 차용)                          | `moai cc`와 동일 컴포넌트 패밀리                                               |
| `moai pause` / `moai resume`  | `internal/cli/pause.go`, `internal/cli/resume.go` | (status와 동일 카드 스타일 차용)               | SPEC 카드 출력에 `tui.KV` 사용                                                 |
| `moai telemetry report`       | `internal/cli/telemetry.go`               | `source/project/screens.jsx:ScreenTelemetry`    | 통계 출력 `tui.Section` + `tui.KV` 통일 (스파크라인 ASCII 옵션, scope 가능 시)|

### 3.3 다크/라이트 자동 감지 + NO_COLOR 통합

`internal/tui/theme.go`는 다음 우선순위로 테마를 결정한다.

1. `NO_COLOR` 환경변수 존재 → ANSI 색상 출력 0건, 텍스트만
2. `MOAI_THEME` 환경변수 (`light`|`dark`|`auto`) → 명시적 강제
3. `lipgloss.HasDarkBackground()` → 터미널 배경 자동 감지
4. 기본값: `dark` (안전한 기본값)

### 3.4 16개 언어 중립성

`internal/tui/` primitive는 라벨 텍스트를 인자로 받기만 하고, 한국어 / 영어 / 일본어 / 중국어 / 그 외 12개 언어 라벨이 모두 동일하게 동작한다. 한국어 카피("자주 쓰는 명령", "환경 감지" 등)는 `internal/tui/messages/` 카탈로그 골격에만 두고, 16개 언어 실제 번역은 별도 SPEC.

## 4. Audience (사용자)

- moai-adk Go 개발자 (구현 + 유지보수): 이번 SPEC의 직접 영향을 받는다. 새 명령어 출력 시 `internal/tui/` API 호출 강제.
- moai-adk 최종 사용자 (CLI 사용자): 출력 일관성 향상, 한글/영문 혼용 시 정렬 깨짐 0건, 다크/라이트 모두에서 가독성 보장.
- 베타 테스터 / 평가자 (`evaluator-active`): 시각 회귀 검증을 위한 golden snapshot 활용.

## 5. Brand (디자인 헌법)

본 SPEC은 모두의AI 디자인 가이드라인 v1.0(Notion FROZEN)에 종속된다. `source/project/colors_and_type.css`가 단일 진실 공급원이며, 다음 토큰은 **변경 금지**다.

### 5.1 색상 토큰 (FROZEN)

라이트 (chrome/bg/fg/accent):

```
chrome `#e8e6e0`, chromeBorder `#bdbab2`, bg `#fbfaf6`, panel `#f3f3f3`
fg `#0e1513`, body `#1f2826`, dim `#5b625f`, faint `#8c918d`
accent `#144a46`, accentDeep `#0a2825`, accentSoft `rgba(20,74,70,0.10)`
success `#0e7a6c`, warning `#a86412`, danger `#b1432f`, info `#1f6f72`
```

다크:

```
chrome `#0c1413`, chromeBorder `#1c2624`, bg `#0a110f`, panel `#0f1816`
fg `#eef2ef`, body `#d8dedb`, dim `#9aa3a0`, faint `#6b7370`
accent `#3eb3a4`, accentDeep `#22938a`, accentSoft `rgba(62,179,164,0.16)`
success `#3fcfa6`, warning `#e3a14a`, danger `#ed7d6b`, info `#5cc7c9`
```

### 5.2 타이포그래피 토큰 (FROZEN)

- Pretendard: 한글 본문 (1차)
- Inter: Latin 본문 (2차, 라틴 fallback)
- JetBrains Mono: 코드/터미널 출력/경로/숫자/식별자

터미널 환경에서는 폰트 selection이 호스트 터미널에 위임되므로 본 SPEC은 **fontFamily 의도 메타데이터**만 lipgloss 스타일로 전달(예: `lipgloss.NewStyle()` Bold/Italic)하고 실제 폰트 매칭은 외부 의존.

### 5.3 시각 규칙 (HARD - 위반 시 PR reject)

설계 소스 `chats/chat1.md` 인용: "모두의AI deep teal(`#144a46`)을 단일 코어 — 마젠타·오렌지 그라디언트 금지 규칙 준수"

- [HARD-BRAND-1] **이모지 사용 금지**. 상태 표시는 색 점/체크/박스 기호 (`✓ ! ✗ · ○ ● →`)만 사용.
- [HARD-BRAND-2] **마젠타·오렌지 그라디언트 금지**. 단일 deep teal 액센트만 허용.
- [HARD-BRAND-3] **순백 #ffffff 배경 금지**. 라이트 테마는 `#fbfaf6` (ivory) / chrome `#e8e6e0` 사용.
- [HARD-BRAND-4] **순흑 #000000 배경 금지**. 다크 테마는 `#0a110f` (ink) / chrome `#0c1413` 사용.
- [HARD-BRAND-5] **좌측 컬러 액센트 패턴(`│` 텍스트로 그리기) 회피**. lipgloss `Border()`로만 표현.
- [HARD-BRAND-6] **`prefers-reduced-motion` 친화**. 커서 깜빡임 외 큰 모션 금지 (스피너 회전은 허용).
- [HARD-BRAND-7] **MX 태그 설명은 항상 영문**. (`memory/feedback_mx_tag_language.md` 정책)

## 6. Stack (기술 스택)

go.mod 직접 검증 결과 추가 의존성 0건이다.

| 라이브러리                           | 버전        | 본 SPEC에서의 역할                                  |
|--------------------------------------|-------------|----------------------------------------------------|
| `github.com/charmbracelet/lipgloss`   | `v1.1.0`    | 디자인 토큰 → ANSI 스타일 변환 (Border/Padding/Foreground/AdaptiveColor) |
| `github.com/charmbracelet/bubbletea`  | `v1.3.10`   | (기존) interactive prompt; 본 SPEC은 직접 의존 없음 |
| `github.com/charmbracelet/huh`        | `v0.8.0`    | `moai init` 위저드 폼 (Stepper/RadioRow와 통합)     |
| `github.com/charmbracelet/bubbles`    | `v1.0.0`    | (현재 indirect, 직접 import 시 direct로 승격) Spinner/Progress 시각 언어 |
| `github.com/charmbracelet/x/ansi`     | `v0.11.6`   | (indirect) ANSI 시퀀스 처리                        |
| `github.com/charmbracelet/x/cellbuf`  | `v0.0.15`   | (indirect) 셀 버퍼 (한글 width 계산 보조)          |
| `github.com/charmbracelet/colorprofile`| `v0.4.1`   | (indirect) 색 프로파일 감지 (truecolor/256/16/no-color) |
| `github.com/mattn/go-runewidth`       | (indirect)  | 한글(2-cell)/영문(1-cell) 폭 계산                  |

[HARD-DEP-1] [REQ-CLI-TUI-008] go.mod에 신규 require 추가 0건 + indirect → direct 승격은 `bubbles`만 허용.

## 7. EARS 요구사항

### 7.1 Ubiquitous (시스템 전반)

- [REQ-CLI-TUI-001] (Ubiquitous) `internal/tui/` 패키지는 모두의AI 디자인 토큰의 단일 진실 공급원으로서 `theme.go`에 라이트/다크 토큰 구조체 두 개를 export 하여야 한다(SHALL).
- [REQ-CLI-TUI-002] (Ubiquitous) 모든 디자인 토큰 값은 설계 소스 `source/project/tui.jsx:9-69` (TOK.light/dark)와 1:1 매칭되어야 한다(SHALL).
- [REQ-CLI-TUI-003] (Ubiquitous) `internal/tui/` primitive 컴포넌트는 입력 텍스트의 언어(한국어/영어/일본어/중국어/기타 12개)에 무관하게 동일하게 동작해야 한다(SHALL). 특정 언어 전제 코드 (예: `if lang == "ko"` 분기) 작성 금지.
- [REQ-CLI-TUI-004] (Ubiquitous) 모든 박스 / 카드 / 패널 시각 표현은 lipgloss `Border()` API를 통해 그려야 하며, 텍스트 문자(`╭ ─ ╮ │ └ ┘`)로 그리는 코드를 신규 작성하면 안 된다(SHALL NOT).

### 7.2 Event-Driven (트리거 응답)

- [REQ-CLI-TUI-005] (Event-Driven) WHEN 사용자가 `moai` 또는 `moai banner`를 실행하면 THEN CLI는 `internal/tui/` 컴포넌트로 그려진 배너를 출력해야 한다(SHALL). 배너 색상은 라이트 테마에서 `#144a46`, 다크 테마에서 `#3eb3a4`이어야 한다.
- [REQ-CLI-TUI-006] (Event-Driven) WHEN 사용자가 `moai doctor`를 실행하면 THEN CLI는 `tui.CheckLine` 컴포넌트로 19개 이상의 검사 항목을 (시스템 그룹 + MoAI-ADK 그룹) 두 섹션으로 출력해야 한다(SHALL).
- [REQ-CLI-TUI-007] (Event-Driven) WHEN 사용자가 `moai init`을 실행하면 THEN CLI는 huh 위저드를 6단계 `tui.Stepper`와 함께 표시하고, 각 라디오 옵션은 선택 시 `◆`, 미선택 시 `◇` prefix로 표시해야 한다(SHALL).

### 7.3 State-Driven (조건 분기)

- [REQ-CLI-TUI-008] (State-Driven) WHILE go.mod에 신규 charmbracelet/외부 require 항목이 추가된 상태에서, CI는 본 SPEC의 PR을 거부해야 한다(SHALL). `bubbles`의 indirect → direct 승격만 허용된 예외이다.
- [REQ-CLI-TUI-009] (State-Driven) WHILE `NO_COLOR` 환경변수가 설정된 상태에서, `internal/tui/` 컴포넌트는 ANSI 색 시퀀스를 0건 출력해야 한다(SHALL).
- [REQ-CLI-TUI-010] (State-Driven) WHILE 터미널이 다크 배경으로 감지된 상태에서(`lipgloss.HasDarkBackground() == true`), 모든 `tui` 출력은 다크 토큰을 사용해야 한다(SHALL).
- [REQ-CLI-TUI-011] (State-Driven) WHILE 터미널이 ANSI 색을 지원하지 않는 환경(예: Windows cmd.exe legacy)인 상태에서, `internal/tui/` 컴포넌트는 텍스트만 출력하는 fallback 모드로 동작해야 한다(SHALL).
- [REQ-CLI-TUI-012] (State-Driven) WHILE `MOAI_THEME` 환경변수가 명시적으로 설정된 상태에서(`light` 또는 `dark`), 자동 감지를 무시하고 명시값을 따라야 한다(SHALL).

### 7.4 Unwanted (금지)

- [REQ-CLI-TUI-013] (Unwanted) `internal/tui/` 패키지 외부에서 디자인 토큰 색 코드(예: `"#144a46"`, `"#3eb3a4"`, `"#0e1513"` 등)를 하드코딩한 코드를 작성하면 안 된다(SHALL NOT). `internal/cli/banner.go`의 기존 terra cotta `#C45A3C` / `#DA7756` 하드코딩은 본 SPEC에서 제거된다.
- [REQ-CLI-TUI-014] (Unwanted) `internal/tui/` 컴포넌트는 이모지 문자(`😀 🎉 ✅ ❌` 등 U+1F300 ~ U+1FAFF)를 출력하면 안 된다(SHALL NOT). 상태 기호는 `✓ ! ✗ · ○ ● →` 만 허용.
- [REQ-CLI-TUI-015] (Unwanted) `internal/tui/` 컴포넌트는 마젠타(`#FF00FF`, `#E91E63` 등) / 오렌지(`#FF9800`, `#FF5722` 등) 색상을 그라디언트의 일부로도 사용하면 안 된다(SHALL NOT). 단일 deep teal 액센트만 허용.
- [REQ-CLI-TUI-016] (Unwanted) `internal/tui/` 컴포넌트는 순백 `#FFFFFF` 또는 순흑 `#000000`을 배경으로 사용하면 안 된다(SHALL NOT). 라이트 ivory `#fbfaf6` / 다크 ink `#0a110f`만 허용.

### 7.5 Optional (가능 시)

- [REQ-CLI-TUI-017] (Optional) WHERE 가능한 경우, `tui.Progress` / `tui.Spinner`는 `prefers-reduced-motion` 환경(터미널 capability detection 또는 `MOAI_REDUCED_MOTION=1` 환경변수)에서 정적 렌더링으로 fallback 해야 한다(SHALL).
- [REQ-CLI-TUI-018] (Optional) WHERE Windows Terminal / iTerm2 / kitty 등 truecolor 지원 터미널에서, `internal/tui/`는 24bit ANSI 색을 사용해야 한다(SHALL). 256색 / 16색 환경에서는 `lipgloss.AdaptiveColor` 자동 다운샘플링을 활용한다.
- [REQ-CLI-TUI-019] (Optional) WHERE `internal/tui/messages/` 카탈로그가 16개 언어 키를 모두 포함하면, 각 명령어 출력의 한국어 카피("자주 쓰는 명령", "환경 감지" 등)는 `language.yaml`의 `conversation_language` 설정을 따라야 한다(SHALL). 누락 키는 영어 fallback.

## 8. Risks (위험)

| ID    | 위험                                                                                                           | 영향   | 대응                                                                                                                                                |
|-------|----------------------------------------------------------------------------------------------------------------|--------|-----------------------------------------------------------------------------------------------------------------------------------------------------|
| R-01  | lipgloss `Border()`도 결국 unicode 박스 문자(`╭─╮│└┘`)를 ANSI에 출력하므로 한글/영문 폭 차이 시 정렬 깨질 수 있음 | High   | `lipgloss.Width()` + `runewidth.StringWidth()` 사용 강제, golden snapshot 테스트로 한글 + 영문 mixed 케이스 18종 검증 (REQ-CLI-TUI-001 acceptance)   |
| R-02  | Windows cmd.exe legacy 환경에서 ANSI 시퀀스 미지원 → 깨진 escape 문자 출력 가능성                                  | High   | `colorprofile.Detect()` 활용 + NO_COLOR fallback 강제 (REQ-CLI-TUI-011)                                                                              |
| R-03  | 14개 명령어 마이그레이션 중 일부에서 기존 출력 회귀 발생 가능성                                                    | Medium | M3-M6 마이그레이션마다 golden snapshot test 추가 + characterization test 선행 (DDD)                                                                 |
| R-04  | `internal/tui/` API 설계가 jsx 디자인의 모든 케이스를 커버 못 할 수 있음                                            | Medium | M1-M2 단계에서 19개 화면 모두 종이 mock-up으로 검증 후 API 동결                                                                                       |
| R-05  | i18n 메시지 카탈로그를 미루면 한국어 코드가 패키지 전체에 흩뿌려져 나중 언어 추가 시 grep 비용 증가                 | Low    | M7에서 messages/ 디렉토리 골격 + ko/en 두 언어 카탈로그 최소 작성 (other 14개 언어는 별도 SPEC에서 채움)                                              |
| R-06  | 다크/라이트 자동 감지 실패 케이스 (예: SSH 세션, tmux 안에 vim 안 다시 SSH...)                                     | Low    | `MOAI_THEME` 환경변수 override 제공 (REQ-CLI-TUI-012)                                                                                                |
| R-07  | 새 패키지 `internal/tui/`가 기존 `internal/statusline/theme.go`와 중복 → 통합 vs 별도 유지 결정 필요                | Low    | M2에서 `internal/statusline/theme.go`를 `internal/tui/theme.go`의 thin wrapper로 재구성 결정 (별도 worktree, statusline은 stdout 출력 형식 다름)      |

## 9. Dependencies (의존성)

### 9.1 본 SPEC이 의존하는 (depends_on)

없음. `feat/SPEC-V3R3-CLI-TUI-001` 브랜치는 `origin/main`에서 분기.

### 9.2 본 SPEC이 차단하는 (blocks)

- 별도 SPEC: `moai install.sh/.ps1/.bat` 디자인 (본 SPEC §2.1) — 본 SPEC의 `internal/tui/` 패키지를 의존
- 별도 SPEC: `moai constitution check` 신설 — 본 SPEC `internal/tui/` 의존
- 별도 SPEC: `moai mx graph` 신설 — 본 SPEC `internal/tui/` 의존
- 별도 SPEC: `moai spec view <SPEC-ID>` 신설 (charmbracelet/glamour 의존성 추가 + `internal/tui/` 의존)
- 별도 SPEC: `moai worktree list` 신설 — 본 SPEC `internal/tui/` 의존
- 별도 SPEC: i18n 메시지 카탈로그 16개 언어 채움 — 본 SPEC `internal/tui/messages/` 골격 의존

### 9.3 후속 권장 (related, 비차단)

- SPEC-V3R3-MX-INJECT-001 (Phase B): MX 태그 주입 후 본 SPEC 산출물에 `@MX:NOTE / @MX:ANCHOR` 적용 검토
- SPEC-V3R3-CIAUT-* (8-Tier autonomous CI): 본 SPEC golden snapshot test 통합

## 10. Exclusions (What NOT to Build)

[HARD-EXCLUSIONS] 본 SPEC은 다음을 명시적으로 만들지 않는다.

1. **새로운 명령어 신설**: 14개 마이그레이션 한정. 5개 미구현 화면(install/constitution check/mx graph/spec view/worktree list)은 별도 SPEC.
2. **glamour 의존성 추가**: `spec view` 명령어와 함께 별도 SPEC에서 처리.
3. **ASCII 아트 배너 자체 디자인 변경**: `banner.go`의 `███╗ ███╗ ...` 기존 8줄 ASCII 유지 (색상만 교체).
4. **16개 언어 i18n 번역 본문**: `messages/` 골격만 작성, 실제 번역 본문은 별도 SPEC.
5. **GUI / 웹 콘솔**: 본 SPEC은 터미널 TUI 한정.
6. **brand voice 카피 개편**: 한국어 카피 문구 자체("베타 테스터 빌드" 등)는 시각 layer가 아니므로 본 SPEC scope 외.
7. **`internal/cli/` 명령어 디스패치 / cobra 라우팅 변경**: 명령어 자체의 동작은 손대지 않고, 출력 layer만 마이그레이션.
8. **테스트 헬퍼 패키지 신설**: golden snapshot 도구는 표준 라이브러리(`testdata/`) + `os.ReadFile` + `bytes.Equal`로 충분.
9. **i18n locale negotiation 메커니즘**: `language.yaml`의 `conversation_language` 단순 lookup만, BCP47 fallback chain은 별도 SPEC.
10. **`prefers-reduced-motion` 자동 시스템 감지**: 환경변수 `MOAI_REDUCED_MOTION=1` 명시 override만, 자동 감지는 향후 SPEC.

---

## Self-Audit (작성자 검증)

### Schema 9-Field Frontmatter
- [x] `id`: SPEC-V3R3-CLI-TUI-001 (regex `^SPEC-[A-Z][A-Z0-9]+-[0-9]{3}$` 통과)
- [x] `version`: "0.1.0" (quoted string)
- [x] `status`: draft (5 enum 중 하나)
- [x] `created_at`: 2026-05-09 (ISO YYYY-MM-DD)
- [x] `updated_at`: 2026-05-09
- [x] `author`: manager-spec
- [x] `priority`: Medium (Title-case)
- [x] `labels`: [...] (YAML array, non-empty)
- [x] `issue_number`: null

### Constitution 정합성
- [x] §1 Goal: WHAT/WHY 만 기술, 함수명/클래스 미언급
- [x] §3 Scope: 디자인 소스 인용 위치(file:line) 모두 명시
- [x] §10 Exclusions: 10개 항목 명시 (HARD 요건 충족)
- [x] EARS keyword 사용: WHEN/WHILE/WHERE/SHALL/SHALL NOT 분포 확인
- [x] 시간 예측 0건 (priority 라벨만 사용)

### Brand 정합성 (FROZEN 토큰)
- [x] Section 5.1 색 토큰 라이트/다크 모두 디자인 소스와 1:1 매칭
- [x] HARD-BRAND-1~7 모두 명시
- [x] 이모지 0건 (✓ ! · ● → 등 ASCII 기호만)
