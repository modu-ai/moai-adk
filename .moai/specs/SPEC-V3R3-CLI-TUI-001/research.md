# SPEC-V3R3-CLI-TUI-001 Research & Design Analysis

> 본 문서는 plan phase에서 수행한 분석 결과를 종합한다. 디자인 소스 인용, 스택 검증, 명령어 갭 분석, 의존성 영향도, 16-language neutrality 영향, 핵심 결정 근거를 포함한다.

---

## 1. 디자인 소스 출처 + 핵심 의도 (verbatim 인용)

### 1.1 핸드오프 번들 위치

```
.moai/design/SPEC-V3R3-CLI-TUI-001/source/
├── README.md                                  # 코딩 에이전트 핸드오프 가이드
├── chats/
│   └── chat1.md                                # 사용자/디자이너 대화 (의도 + 두 차례 이터레이션)
└── project/
    ├── moai CLI 터미널 디자인 v2.html          # 메인 디자인 캔버스
    ├── moai CLI 터미널 디자인.html             # v1 (사용자가 거부함)
    ├── colors_and_type.css                     # 모두의AI 디자인 토큰 (Notion v1.0, FROZEN)
    ├── tui.jsx                                  # TUI primitive 컴포넌트 라이브러리 (430+ LOC)
    ├── screens.jsx                              # 19개 명령어 화면 렌더러 (52KB)
    ├── scenes.jsx                               # v1 폐기 버전 (참조용)
    ├── terminals.jsx                            # 캔버스 인프라
    ├── design-canvas.jsx                        # 캔버스 인프라
    ├── assets/moai-logo-*.png                   # 로고
    └── uploads/*.png                            # 사용자 첨부 스크린샷
```

### 1.2 v1 폐기 사유 (chat1.md verbatim)

사용자의 거부 메시지 (`chat1.md` line 111):

> CLI 에서 터미널 라인이 깨진다. 이런 라인 말고 moai-adk 에서 go 에서 사용중인 라이브러리를 이용해서 선이 깨지지 않게 미적으로 디자인이 우수한 형태로 다시 전면 재수정하도록하자. tui가 현대적으로 다시 업데이트를 하도록 하자. 전체 다시 수정해라.

디자이너의 진단 (`chat1.md` line 117-118):

> 박스 그리기가 깨진 이유는 한국어 글자(2칸 너비)와 영문 글자(1칸 너비)를 같은 박스 안에 섞으면서 모든 라인을 같은 길이로 만들 수 없었기 때문입니다.
> moai-adk가 실제로 사용하는 라이브러리(`charmbracelet/lipgloss` + `bubbles`)의 스타일을 따라가는 게 정답이에요 — 이 라이브러리들은 ASCII 라인 드로잉 대신 **CSS의 border + padding처럼 작동하는 스타일 박스**를 사용합니다. 박스 글자를 텍스트로 그리지 않고, 실제 박스로 그리면 정렬이 깨지지 않습니다.

### 1.3 v2 해결 전략 (chat1.md verbatim line 123-126)

> 1. **lipgloss 스타일 시스템 매핑** — `Border()`/`Padding()`/`Foreground()` 를 실제 CSS border/padding으로 1:1 변환
> 2. **bubbles 컴포넌트 차용** — `huh` 폼, `spinner`, `progress`, `table`, `list` 같은 실제 사용 컴포넌트의 시각 언어
> 3. **모노스페이스 정렬에 의존하지 않는 레이아웃** — flex/grid 기반, 한글 본문은 Pretendard, 토큰/경로/숫자만 JetBrains Mono
> 4. **현대적 TUI 트렌드** — gum/glamour/mods의 라운드 박스, 그라디언트 헤더, 인라인 칩, 미니 spark, 키보드 힌트 바

### 1.4 디자인 시스템 적용 포인트 (chat1.md verbatim line 95-102)

> - 어두운 청록(`#144a46`) 단일 코어 — 마젠타·오렌지 그라디언트 금지 규칙 준수
> - ASCII 박스 드로잉(`╭ ─ ╮`)으로 카드 구조, `│` 좌측 컬러 액센트 패턴은 회피
> - 라이트는 ivory `#fafaf7` (순백 회피), 다크는 ink `#09110f` (순흑 회피)
> - 한국어 본문은 Pretendard, 터미널 출력은 JetBrains Mono
> - "모두의" / "베타 테스터" / "CX 7원칙" 어휘, 친근체 + 존대 혼용
> - 이모지 미사용 — 대신 색 점/체크/박스 기호로 상태 표현
> - `prefers-reduced-motion` 친화 (커서 깜빡임 외 큰 모션 없음)

> [참고] v1 카피의 ivory(`#fafaf7`)와 ink(`#09110f`)는 v2(tui.jsx)에서 ivory(`#fbfaf6`) + ink(`#0a110f`)로 미세조정됨. spec.md §5.1 토큰값이 최종 진실이다.

---

## 2. tui.jsx → lipgloss API 매핑 표

설계 소스 `source/project/tui.jsx`의 React/CSS primitive를 Go lipgloss API로 매핑하는 1:1 변환표다. M1-M2 milestone 작성 시 참조한다.

### 2.1 컴포넌트 매핑

| jsx 컴포넌트       | source 위치          | lipgloss 매핑                                                                                                       | Go 시그니처 (제안)                                                            |
|--------------------|----------------------|---------------------------------------------------------------------------------------------------------------------|-------------------------------------------------------------------------------|
| `Box`              | tui.jsx:88-133       | `lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(P, P).BorderForeground(c)`                            | `Box(opts BoxOpts) string`                                                     |
| `ThickBox`         | tui.jsx:136-154      | `Border(thickBorder)` (custom 2px 시각 효과는 Bold + ThickBorder 조합)                                                | `ThickBox(opts BoxOpts) string`                                                |
| `Pill`             | tui.jsx:157-177      | `lipgloss.NewStyle().Background(bg).Foreground(fg).Padding(0, 1).Margin(0, 1)` + 라운드 corner 시뮬레이션              | `Pill(kind PillKind, solid bool, label string) string`                         |
| `StatusIcon`       | tui.jsx:180-197      | `lipgloss.NewStyle().Foreground(c).Bold(true).Width(2)` + char map (`✓ ! ✗ · ○ ● →`)                                 | `StatusIcon(kind IconKind) string`                                             |
| `Spinner`          | tui.jsx:200-210      | (정적 출력) `⠋⠙⠹⠸⠼⠴⠦⠧⠇⠏` 또는 단일 `●` / 호출자가 frame 관리                                                          | `Spinner(label string, frame int) string`                                      |
| `Progress`         | tui.jsx:213-233      | `lipgloss.NewStyle().Width(W)` + filled char (`█`) + unfilled char (`░`) gradient                                    | `Progress(value, max int, opts ProgressOpts) string`                           |
| `Stepper`          | tui.jsx:236-250      | filled dot (`●`) + unfilled dot (`○`) + ` n/N ` suffix                                                                | `Stepper(current, total int) string`                                           |
| `RadioRow`         | tui.jsx:253-276      | `◆` (selected) / `◇` (unselected) prefix + colored label                                                              | `RadioRow(opts RadioOpts) string`                                              |
| `CheckRow`         | tui.jsx:279-293      | `[✓]` / `[ ]` checkbox + label                                                                                        | `CheckRow(checked bool, label, hint string) string`                            |
| `KV`               | tui.jsx:296-305      | label (dim) + spaces + value (mono fg)                                                                                | `KV(key, value string, opts KVOpts) string`                                    |
| `CheckLine`        | tui.jsx:308-319      | `<StatusIcon> <label> <value> · <hint>` row                                                                            | `CheckLine(status IconKind, label, value, hint string) string`                 |
| `Section`          | tui.jsx:322-340      | accent bar (`▍` 또는 `█`) + bold title + dim sub                                                                       | `Section(title string, opts SectionOpts) string`                               |
| `Prompt`           | tui.jsx:343-355      | `❯ ` + path (cyan) + ` on ` + branch (warn/success) + cmd                                                              | `Prompt(host, path, branch string, dirty bool, cmd string) string`             |
| `Cursor`           | tui.jsx:357-364      | (애니메이션) `█` blinking; static fallback `█`                                                                          | `Cursor() string`                                                              |
| `Term`             | tui.jsx:367-414      | (스크린샷 chrome) traffic-light + title bar — **실제 출력에선 `MOAI_SCREENSHOT=1`에서만 활성, default no-op**            | `Term(opts TermOpts, body string) string`                                      |
| `HelpBar`          | tui.jsx:417-434      | key chip (`<kbd>`) + description, ` · ` separator로 join                                                              | `HelpBar(items []KeyHint) string`                                              |

### 2.2 토큰 매핑

| jsx 토큰 카테고리       | tui.jsx 위치     | Go 매핑                                                                  |
|-------------------------|------------------|--------------------------------------------------------------------------|
| 라이트 28키             | tui.jsx:9-38     | `Theme` struct field별 `lipgloss.Color("#xxxxxx")`                       |
| 다크 28키               | tui.jsx:39-68    | (동일 struct, dark 값 채움)                                              |
| `accentSoft` (rgba)     | tui.jsx:23, 53   | rgba는 ANSI에서 native 미지원 → opacity blend는 dim faint 패턴으로 대체  |
| `shadow` (CSS shadow)   | tui.jsx:37, 67   | 터미널에서 무시 (메타데이터로만 보존)                                     |
| `titleBar` (gradient)   | tui.jsx:13, 42   | `Term` chrome에서만 사용 (스크린샷 전용); ANSI에선 단색 fallback           |

### 2.3 폰트 매핑

| jsx 폰트       | 용도                  | Go 매핑                                                              |
|----------------|-----------------------|----------------------------------------------------------------------|
| Pretendard     | 한글 본문             | (터미널 위임; lipgloss는 Bold/Italic 메타만 가짐)                     |
| Inter          | 라틴 본문 (2차)       | (터미널 위임)                                                        |
| JetBrains Mono | 코드/경로/숫자/식별자 | (모든 ANSI 출력은 모노스페이스로 가정)                                |

---

## 3. screens.jsx → 명령어 surface 매핑 + 갭 분석

### 3.1 19개 화면 vs 현재 moai-adk-go 명령어 surface

`internal/cli/` 디렉토리의 cobra 명령어를 직접 확인한 결과:

#### 존재하는 명령어 (현재 implementing 가능)

| 디자인 화면              | screens.jsx 위치                       | moai-adk-go 명령어               | 영향받는 파일                                       |
|--------------------------|----------------------------------------|----------------------------------|-----------------------------------------------------|
| ScreenBanner             | screens.jsx:ScreenBanner (line ~155)    | `moai` (root no-arg)              | `internal/cli/banner.go`                             |
| ScreenHelp               | screens.jsx:ScreenHelp                  | `moai help`, `moai --help`        | cobra root.go의 `Long` template + `internal/cli/render.go` |
| ScreenInit               | screens.jsx:ScreenInit                  | `moai init`                       | `internal/cli/init.go`, `internal/cli/wizard/`       |
| ScreenDoctor             | screens.jsx:ScreenDoctor                | `moai doctor`                     | `internal/cli/doctor*.go` (6 파일)                   |
| ScreenStatus             | screens.jsx:ScreenStatus                | `moai status`                     | `internal/cli/status.go`                             |
| ScreenVersion            | screens.jsx:ScreenVersion               | `moai version`                    | `pkg/version/version.go` 호출 출력                    |
| ScreenUpdate             | screens.jsx:ScreenUpdate                | `moai update`                     | `internal/cli/update.go`                             |
| ScreenCc                 | screens.jsx:ScreenCc                    | `moai cc`                         | `internal/cli/cc.go`                                 |
| ScreenLoopStart          | screens.jsx:ScreenLoopStart             | `moai loop start`                 | `internal/cli/loop_*.go`                             |
| ScreenStatusline         | screens.jsx:ScreenStatusline            | `moai statusline`                 | `internal/statusline/*.go`                           |
| ScreenError              | screens.jsx:ScreenError                 | (전역 에러 출력)                  | `internal/cli/render.go` + 분산                       |

→ **합계 11개 화면 (15개 surface 포함; cg/glm/pause/resume 등 family 포함하면 14)**

#### 부재 명령어 (별도 SPEC)

| 디자인 화면              | screens.jsx 위치          | 분리 사유                                                          |
|--------------------------|---------------------------|--------------------------------------------------------------------|
| ScreenInstallSh          | screens.jsx:5-200          | `scripts/install.sh` 미존재                                          |
| ScreenInstallPs1         | screens.jsx:60-...         | `scripts/install.ps1` 미존재                                         |
| ScreenInstallBat         | screens.jsx:120-...        | `scripts/install.bat` 미존재                                         |
| ScreenConstitution       | screens.jsx:Constitution    | `moai constitution` 존재; `check` 서브커맨드 미구현                  |
| ScreenMxGraph            | screens.jsx:MxGraph         | `moai mx` 존재; `graph` 서브커맨드 미구현                             |
| ScreenSpecView           | screens.jsx:SpecView        | `moai resume <ID>` 존재; `spec view` 서브커맨드 부재 + glamour 의존성 |
| ScreenWorktreeList       | screens.jsx:WorktreeList    | `moai worktree` 자체 미구현                                           |
| ScreenTelemetry          | screens.jsx:Telemetry       | `moai telemetry` 존재 — 이번 SPEC scope에 포함 검토 (currently in §3.2 last row) |

→ **합계 8개 화면 (5개는 별도 SPEC)**. ScreenTelemetry는 본 SPEC §3.2 14번째 surface로 포함 (현재 `internal/cli/telemetry.go` 존재).

### 3.2 19개 화면별 핵심 visual element 추출

| 화면 ID     | 핵심 컴포넌트                                                     | 디자인 의도                                          |
|-------------|-------------------------------------------------------------------|------------------------------------------------------|
| Banner      | ASCII art + ThickBox accent + Pill triplet + Section grid          | 진입점, 브랜드 첫인상                                |
| Help        | Section x4 (그룹별) + KV row + HelpBar                              | 명령 카탈로그                                        |
| Init        | Stepper + Box accent + RadioRow x5 + 정보 Box (dashed)              | 6단계 wizard, 라디오 선택                            |
| Doctor      | Section x2 (시스템 + MoAI-ADK) + CheckLine x9 + Box 요약             | 환경 진단, 19+ 검사 항목                              |
| Status      | KV grid + Pill 카드 (4 색) + Progress row x3                        | 프로젝트 현황 + SPEC 카드                            |
| Version     | ThickBox + ASCII mini + Pill triplet                               | 버전 카드 (단순)                                     |
| Update      | Box accent + huh confirm + Progress 다운로드                         | 업데이트 흐름                                        |
| Cc          | Box 환경 점검 + Spinner 부트스트랩                                  | Claude Code 런처                                     |
| LoopStart   | 4-phase Stepper (Spec/Plan/Impl/Sync) + 페이즈별 Box                | 자율 개발 페이즈                                     |
| Statusline  | 단일 라인 출력 (tmux/vim 임베드)                                    | 색 토큰 통일                                         |
| Error       | ThickBox danger + StatusIcon err + KV (root cause)                  | 거버넌스 위반 / 에러                                  |
| Telemetry   | Section + KV + 미니 스파크라인 (선택)                                | 통계 출력                                            |
| (별도 SPEC) Install Sh/Ps1/Bat | ThickBox + CheckLine + Progress + Box 완료      | 설치 흐름                                            |
| (별도 SPEC) Constitution | Section 7 (CX 7원칙) + Pill 점수                       | 거버넌스 검사                                        |
| (별도 SPEC) MxGraph     | Tree/Graph render (custom, charm 외부)                       | 앵커 그래프                                          |
| (별도 SPEC) SpecView    | glamour markdown render + KV metadata                        | SPEC 카드 보기                                       |
| (별도 SPEC) WorktreeList| Table 컴포넌트 + Pill (status)                               | 워크트리 목록                                        |

---

## 4. 의존성 영향도 (go.mod 직접 검증)

### 4.1 현재 charmbracelet 의존성 (go.mod 검증)

```
github.com/charmbracelet/bubbletea v1.3.10              # direct
github.com/charmbracelet/huh v0.8.0                      # direct
github.com/charmbracelet/lipgloss v1.1.0                 # direct
github.com/charmbracelet/x/powernap v0.1.4               # direct (huh 의존)
github.com/charmbracelet/bubbles v1.0.0                  # indirect ← 본 SPEC에서 direct 승격 가능
github.com/charmbracelet/colorprofile v0.4.1             # indirect
github.com/charmbracelet/x/ansi v0.11.6                  # indirect
github.com/charmbracelet/x/cellbuf v0.0.15               # indirect
github.com/charmbracelet/x/exp/strings ...               # indirect
github.com/charmbracelet/x/term v0.2.2                   # indirect
```

go.sum에는 위 외에도 transitive 의존성 (`x/conpty`, `x/errors`, `x/exp/golden`, `x/termios`, `x/xpty`)이 존재하지만, 본 SPEC에서 직접 import하는 것은 `lipgloss`/`huh`/`bubbles` 한정.

### 4.2 본 SPEC에서 추가로 필요한 의존성: 0건

검증 결과:
- ✅ `lipgloss v1.1.0` — 모든 Border/Padding/Foreground/AdaptiveColor API 활용 가능
- ✅ `huh v0.8.0` — Form/Select/Input 모두 활용 가능; 라디오 prefix `◆/◇` 커스터마이징은 huh `Theme` API로 가능 (M5에서 검증)
- ✅ `bubbles v1.0.0` — Spinner/Progress 시각 언어 차용 (직접 import 시 indirect → direct 승격, 본 SPEC에서 허용된 유일 변경)
- ✅ `colorprofile v0.4.1` — `colorprofile.Detect()`로 truecolor/256/16/no-color fallback 자동
- ✅ `x/ansi v0.11.6` — ANSI escape 분석 (golden test에서 사용)
- ✅ `x/cellbuf v0.0.15` — 한글 cell width 계산 보조
- ✅ `mattn/go-runewidth` (transitive) — `runewidth.StringWidth(s)` 한글 폭 계산 진실 공급원

### 4.3 의존성에 추가하지 않는 것 (Non-Goal)

- ❌ `charmbracelet/glamour` — `spec view` 명령어 SPEC에서 별도 추가
- ❌ `bubbletea` 직접 의존 — 본 SPEC은 stateless rendering만; bubbletea는 init wizard에서 huh가 내부적으로 사용 중 (transitive)
- ❌ 외부 ANSI 라이브러리 (예: `fatih/color`, `mgutz/ansi`) — lipgloss 만으로 충분

### 4.4 go.mod 변경 시나리오

본 SPEC PR의 go.mod 변경은 **0줄 또는 1줄(bubbles indirect 주석 제거)** 만 허용.

```diff
# Before (현재)
github.com/charmbracelet/bubbles v1.0.0 // indirect

# After (M2에서 bubbles 직접 import 시)
github.com/charmbracelet/bubbles v1.0.0
```

이 외 모든 변경은 REQ-CLI-TUI-008 위반.

---

## 5. 16-Language Neutrality 영향 분석

### 5.1 CLAUDE.local.md §15 정책 적용

본 프로젝트의 16-language 정책 (CLAUDE.local.md §15):
> `internal/template/templates/` 하위는 16개 언어(go/python/typescript/javascript/rust/java/kotlin/csharp/ruby/php/elixir/cpp/scala/r/flutter/swift) 동등 취급

**중요**: 본 SPEC은 `internal/tui/` (= **non-template** Go 코드)를 변경하므로, CLAUDE.local.md §15의 직접적 적용 대상은 아니다. 그러나:

- `internal/tui/` primitive는 모든 16-language 사용자의 CLI 출력에 영향
- 한국어 카피("자주 쓰는 명령" 등)가 함수 시그니처에 hardcoded 되면 다국어 사용자에게 부적합
- 따라서 i18n 카탈로그 골격(`internal/tui/messages/`)을 도입하여 **컴포넌트 자체는 언어 중립**, **카피만 catalog에서 lookup** 하는 분리

### 5.2 16-Language ↔ messages catalog 매핑

본 프로젝트의 16-language list는 사용자 프로젝트의 **구현 언어**(Go/Python/...)이지, CLI **사용자 언어**(ko/en/zh/ja)와 다르다. 두 개념을 구분:

| 개념                           | 예시                                | 본 SPEC 적용                                    |
|--------------------------------|-------------------------------------|-------------------------------------------------|
| 16 implementation languages    | go, python, typescript, ...         | `internal/tui/` primitive는 모두 동일하게 동작   |
| Conversation languages         | ko (Korean), en (English), zh, ja   | `messages/{lang}.yaml` catalog                  |

본 SPEC `messages/`는 conversation language 단위로 작성 (M7에서 ko/en 두 개 채움; zh/ja 등 12개 lang은 별도 SPEC).

### 5.3 검증 방법

- AC-CLI-TUI-009: `messages/` 디렉토리에 16개 yaml 파일 존재 (실제로는 conversation language 16개 — moai-adk-go가 16-language를 conversation language list와 등치하면; 정책 명확화 필요)
- 영어 fallback: 누락된 lang 키는 영어로 fallback (REQ-CLI-TUI-019)

> **Open Question (run phase)**: `internal/tui/messages/`의 언어 갯수를 16개로 할지, 4개(ko/en/zh/ja, docs-site §17과 일치)로 할지 결정 필요. plan-auditor 1차 audit에서 정책 합의.

---

## 6. 핵심 결정 근거 + Trade-off 분석

### 6.1 결정: ASCII 박스 그리기 → lipgloss `Border()` 강제 (REQ-CLI-TUI-004)

**Pros**:
- 한글/영문 폭 차이 발생 시 lipgloss 내부 width 계산이 자동 보정 (최소한 `lipgloss.Width()` 활용 가능)
- 설계 소스 chat1.md verbatim line 117-119 의 사용자 의도와 일치
- 다크/라이트 자동 색 적용 (`AdaptiveColor`)

**Cons**:
- lipgloss도 결국 unicode 박스 문자(`╭─╮│└┘`)를 ANSI에 출력하므로 fragile windows cmd.exe legacy 시 깨질 수 있음 (Risk R-02)
- → mitigation: NO_COLOR fallback + colorprofile.Detect() (REQ-CLI-TUI-011)

**채택**: 채택. ASCII 텍스트 박스 신규 작성은 PR reject.

### 6.2 결정: glamour 의존성 본 SPEC scope 제외

**Pros**:
- `spec view` 명령어 자체가 부재 → 함께 신설하면 SPEC scope 폭증 (>40 EARS REQ 가능성)
- glamour는 markdown rendering에 한정된 의존성 → spec view SPEC에서 정밀 도입 가능

**Cons**:
- 디자인 소스 `screens.jsx:ScreenSpecView`에 glamour 사용 명시 → 일관성 저해
- doctor 화면의 "Glamour 캐시 갱신 필요" warn 항목이 placeholder 처리 필요

**채택**: 제외. `moai spec view` SPEC을 별도로 작성하여 glamour 도입.

### 6.3 결정: 8줄 ASCII 아트 배너 디자인 그대로 유지

**Pros**:
- `internal/cli/banner.go`의 `███╗ ███╗ ...` 8줄은 brand recognition 자산
- 디자인 소스 `screens.jsx:ScreenBanner`의 미니 ASCII (`__ __ __ ___ _`)는 좁은 width 캔버스용 mock-up
- 본 SPEC은 색상만 교체(terra cotta → deep teal)로 risk 최소화

**Cons**:
- 디자인 소스의 `__ __ __ ___ _` 미니 ASCII가 더 modern; 사용자가 향후 변경 요청 가능성

**채택**: 8줄 ASCII 유지. 향후 별도 visual identity SPEC에서 ASCII 자체 재디자인 검토.

### 6.4 결정: `internal/statusline/theme.go` 통합 vs 유지

**Pros (통합)**:
- 단일 진실 공급원 (REQ-CLI-TUI-001)
- 색 정의 중복 제거

**Cons (통합)**:
- statusline은 stdout 출력 형식이 다름 (no trailing newline, tmux escape 등)
- 통합 시 `internal/tui/`가 statusline 특수 케이스를 알게 됨 → 추상화 leak

**채택**: M6에서 thin wrapper (`internal/statusline/theme.go`는 `internal/tui/theme.go`만 import)로 재구성. 색 정의는 `tui` 단일 출처, 출력 포맷팅 로직은 statusline 유지.

### 6.5 결정: `MOAI_THEME` override 환경변수 도입

**Pros**:
- `lipgloss.HasDarkBackground()` 자동 감지가 SSH/tmux nested 환경에서 실패 가능
- 사용자가 명시적 강제 가능
- `MOAI_THEME=light moai doctor` 같은 ad-hoc 검증 용이

**Cons**:
- 추가 환경변수 = 추가 documentation 부담

**채택**: 채택. 사용자 control + golden test 활용도 양쪽 이점.

---

## 7. moai-adk-go 코드베이스 영향도

### 7.1 변경되는 파일 (Migration scope)

| 카테고리                | 파일 수 | 주요 파일                                                                   |
|-------------------------|---------|-----------------------------------------------------------------------------|
| 신규 패키지 `internal/tui/` | ~15 | theme/box/pill/status/form/table/prompt/term/help/detect/profile/i18n + tests |
| `internal/cli/banner.go` | 1 | terra cotta 제거                                                            |
| `internal/cli/doctor*.go` | 6 | doctor.go, doctor_config.go, doctor_brain.go, ...                            |
| `internal/cli/status.go` 등 | 2-3 | status.go, status_render.go                                                  |
| `internal/cli/version.go` | 1 | (또는 pkg/version 호출 출력)                                                 |
| `internal/cli/update.go` | 1 |                                                                              |
| `internal/cli/init.go` + `internal/cli/wizard/` + `internal/cli/profile_setup.go` | 3 | huh wizard 강화                                                              |
| `internal/cli/cc.go` + `cg.go` + `glm.go` | 3 |                                                                              |
| `internal/cli/loop_*.go` | 2-3 | loop_start.go 등                                                             |
| `internal/cli/render.go` | 1 | RenderError 신설                                                             |
| `internal/statusline/*.go` | ~5 | theme.go thin wrapper로 재구성                                               |
| `internal/cli/telemetry.go` | 1 |                                                                              |
| `internal/cli/pause.go` + `resume.go` | 2 | (status family 차용)                                                         |
| 새 testdata/golden snapshot | 100+ | 라이트/다크/no-color × 19개 화면 + variant                                     |

### 7.2 변경되지 않는 파일 (Out-of-scope)

- `internal/template/templates/**` 전체 (본 SPEC은 internal/ 만 변경, embedded 산출물 변경 없음)
- `internal/hook/**`, `internal/lsp/**`, `internal/mcp/**` 등 비-CLI 모듈
- `cmd/moai/**` (entry point는 변경 없음)
- `pkg/version/version.go` 자체 (호출하는 cli 출력만 변경)

### 7.3 Test 영향

- 기존 `*_test.go` 테스트는 일부 golden snapshot 비교 시 ANSI 출력 변경으로 실패할 수 있음
- M3-M6 각 milestone에서 PRESERVE phase로 회귀 검증 + golden 재생성

---

## 8. 디자인 소스 인용 횟수 종합

본 SPEC 4개 artifact 전체에서 디자인 소스 인용 횟수:

| Artifact         | source/project/* 인용 횟수 | source/chats/chat1.md 인용 횟수 |
|------------------|----------------------------|--------------------------------|
| spec.md          | 8 (tui.jsx + screens.jsx)   | 1 (line 117-118)               |
| plan.md          | 12 (screens.jsx 모든 화면)  | 0                              |
| acceptance.md    | 1 (tui.jsx:9-69 토큰)        | 0                              |
| research.md      | 25+ (모든 인용 verbatim)    | 4 verbatim block               |
| **Total**        | **46+**                     | **5**                          |

---

## 9. Open Questions (run phase에서 결정)

plan.md §12와 동일하나 정책 결정 필요한 항목 추가:

1. **`internal/statusline/theme.go` 통합 vs 유지**: M6에서 thin wrapper 결정 (research §6.4)
2. **`huh` 라이브러리의 라디오 prefix 커스터마이징 범위**: huh 0.8 API 제한 시 wrapper로 직접 그리기
3. **Glamour 캐시 표시**: doctor 화면에서 conditional skip 또는 placeholder
4. **bubbles indirect → direct 승격 시점**: M2 status.go에서 결정
5. **Windows cmd.exe legacy fallback**: CI windows-latest matrix에 cmd.exe runner 추가 vs unit test mock
6. **`messages/` 언어 갯수**: 16-language 또는 4-language(ko/en/zh/ja, docs-site §17 일관)
7. **golden snapshot 도구 선택**: `charmbracelet/x/exp/golden` (이미 indirect 의존) vs 표준 lib `os.ReadFile + bytes.Equal`
8. **`MOAI_REDUCED_MOTION=1` 환경변수 표준화**: 본 SPEC scope 내 또는 별도 SPEC

위 8개는 plan-auditor audit 후 acceptance.md update 또는 follow-up SPEC 분리.

---

## 10. Self-Audit (작성자 검증)

- [x] 핸드오프 번들 모든 파일(README/chats/project) 내용 확인 후 작성
- [x] chat1.md verbatim 인용 4 block (line 95-102, 111, 117-118, 123-126)
- [x] tui.jsx primitive 매핑 표 16 컴포넌트 + 토큰/폰트 매핑
- [x] screens.jsx 19 화면 vs 현재 명령어 surface 갭 분석 (11 in-scope, 8 별도 SPEC, 14 surface 추출)
- [x] go.mod 직접 검증 → 추가 의존성 0건
- [x] 16-language vs conversation language 구분 명확화
- [x] core 결정 5건 trade-off 분석
- [x] open question 8건 명시
- [x] 디자인 소스 인용 51회 + verbatim 5 block
