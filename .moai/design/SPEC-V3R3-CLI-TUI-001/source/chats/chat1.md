# Moai CLI Design

_Started 2026-05-06 01:53 UTC_

---

## User

터미널 cli 인 moai 의 모든 명령어와 설치 스크립트등 cli 관련된 모든 내용을 모두의AI 디자인 시스템에 맞게 모든 내용을 ascii 텍스트 디자인과 색상을 디자인해줘. 다크와 라이트 디자인을 부탁해.

## Assistant

_[tool: local_ls]_

_[tool: local_ls]_

_[tool: local_ls]_

_[tool: local_read]_

moai CLI 명령어 전반과 설치 스크립트를 살펴보고 모두의AI 디자인 시스템에 맞춰 ASCII 터미널 디자인을 제작하겠습니다. 먼저 명령어 구조부터 파악합니다.

_[tool: local_ls]_

_[tool: local_ls]_

_[tool: local_read]_

_[tool: local_read]_

I'll continue exploring the CLI commands then build the design.

_[tool: local_grep]_

I have enough info to design. Let me check a few key command outputs and the design system colors quickly.

_[tool: snip]_

_[tool: local_read]_

_[tool: local_read]_

_[tool: local_read]_

I have enough context now. Let me build the design directly. I'll create a comprehensive HTML showing all moai CLI commands styled with the 모두의AI design system in both light and dark themes, with tweaks for switching.

_[tool: copy_files]_

_[tool: read_file]_

I have enough context. Let me build the design now.

_[tool: snip]_

Now I'll build the comprehensive ASCII terminal design. The plan:

1. **Single HTML page** with a `<DCSection>`-based design canvas showing all CLI commands as terminal frames
2. **Light + Dark variants** side by side
3. Coverage: install scripts (sh/ps1/bat), banner, init wizard, doctor, status, version, update, cc/cg/glm launchers, worktree, spec, telemetry, constitution, errors, help
4. Use 모두의AI deep teal (#144a46) as accent, Pretendard for chrome, JetBrains Mono for terminal content
5. ASCII art that respects the brand (no emoji, no rainbow gradients — just teal accent + neutrals)

_[tool: copy_starter_component]_

Now let me create the main file. I'll build terminal frames as React components, with light/dark variants:

_[tool: write_file]_

Now the scenes file with all the CLI command renderings:

_[tool: write_file]_

Now the main HTML file with the design canvas to display all scenes side-by-side, in light and dark variants:

_[tool: write_file]_

_[tool: snip]_

_[tool: done]_

_[tool: fork_verifier_agent]_

moai-adk CLI 전체를 모두의AI 디자인 시스템으로 재해석한 터미널 디자인을 만들었습니다. 라이트와 다크 두 변형으로 나란히 배치된 디자인 캔버스이며, 각 아트보드를 클릭해 확대해서 볼 수 있어요.

**커버한 화면 (각 라이트 / 다크 1쌍씩, 총 38개):**

- **설치 스크립트 3종** — `install.sh` (macOS / Linux), `install.ps1` (Windows PowerShell), `install.bat` (cmd 폴백)
- **메인 인터페이스** — 배너 + 도움말, `moai help` 레퍼런스 카드
- **프로젝트 명령어** — `init` 위저드, `doctor`, `status`, `version`, `update`
- **런처** — `moai cc` (Claude Code 부트스트랩)
- **자율 개발** — `loop start`, `spec view`, `worktree list`
- **품질 / 거버넌스** — `constitution check` (CX 7원칙), `mx graph` (MX 앵커), `telemetry report`
- **시스템** — `statusline` (tmux/vim 임베드), 에러 상태

**디자인 시스템 적용 포인트:**
- 어두운 청록(`#144a46`) 단일 코어 — 마젠타·오렌지 그라디언트 금지 규칙 준수
- ASCII 박스 드로잉(`╭ ─ ╮`)으로 카드 구조, `│` 좌측 컬러 액센트 패턴은 회피
- 라이트는 ivory `#fafaf7` (순백 회피), 다크는 ink `#09110f` (순흑 회피)
- 한국어 본문은 Pretendard, 터미널 출력은 JetBrains Mono
- "모두의" / "베타 테스터" / "CX 7원칙" 어휘, 친근체 + 존대 혼용
- 이모지 미사용 — 대신 색 점/체크/박스 기호로 상태 표현
- `prefers-reduced-motion` 친화 (커서 깜빡임 외 큰 모션 없음)

**다음 단계 제안 (필요하면 말씀해주세요):**
- 특정 명령어를 실제 터미널에서 캡처할 수 있는 ANSI 색상 정의 (`lipgloss`/`charmbracelet` 스타일 매핑)
- `statusline` tmux/zsh 실제 적용 예시
- 한국어 / 영어 두 가지 로케일 변형

## User

CLI 에서 터미널 라인이 깨진다. 이런 라인 말고 moai-adk 에서 go 에서 사용중인 라이브러리를 이용해서 선이 깨지지 않게 미적으로 디자인이 우수한 형태로 다시 전면 재수정하도록하자. tui가    현대적으로 다시 업데이트를 하도록 하자.  전체 다시 수정해라.

## Assistant

_[tool: snip]_

좋은 지적입니다. 박스 그리기가 깨진 이유는 한국어 글자(2칸 너비)와 영문 글자(1칸 너비)를 같은 박스 안에 섞으면서 모든 라인을 같은 길이로 만들 수 없었기 때문입니다. 

moai-adk가 실제로 사용하는 라이브러리(`charmbracelet/lipgloss` + `bubbles`)의 스타일을 따라가는 게 정답이에요 — 이 라이브러리들은 ASCII 라인 드로잉 대신 **CSS의 border + padding처럼 작동하는 스타일 박스**를 사용합니다. 박스 글자를 텍스트로 그리지 않고, 실제 박스로 그리면 정렬이 깨지지 않습니다.

전면 재수정 계획:

1. **lipgloss 스타일 시스템 매핑** — `Border()`/`Padding()`/`Foreground()` 를 실제 CSS border/padding으로 1:1 변환
2. **bubbles 컴포넌트 차용** — `huh` 폼, `spinner`, `progress`, `table`, `list` 같은 실제 사용 컴포넌트의 시각 언어
3. **모노스페이스 정렬에 의존하지 않는 레이아웃** — flex/grid 기반, 한글 본문은 Pretendard, 토큰/경로/숫자만 JetBrains Mono
4. **현대적 TUI 트렌드** — gum/glamour/mods의 라운드 박스, 그라디언트 헤더, 인라인 칩, 미니 spark, 키보드 힌트 바

먼저 moai-adk가 실제 어떤 lipgloss/bubbles 스타일을 쓰는지 짧게 확인:

_[tool: local_grep]_

_[tool: local_grep]_

Confirmed — moai-adk uses **bubbletea + lipgloss + bubbles + huh + glamour**. Now I'll completely rewrite the design using these libraries' real visual language: lipgloss `Border()` (rendered as actual CSS borders), `huh` form selectors with `◆/◇` prefixes, `bubbles/spinner` dots, `bubbles/progress` bars, `glamour` markdown style — no more hand-drawn `╭─╮` boxes.

_[tool: local_read]_

_[tool: local_read]_

I have everything I need. Now I'll completely rewrite the design using lipgloss/bubbles/huh visual primitives — actual CSS borders instead of ASCII characters, proper text alignment that handles Korean character widths.

_[tool: snip]_

_[tool: write_file]_

Now the screens. Each artboard renders a real moai-adk command with proper bubbletea/lipgloss visual output.

_[tool: write_file]_

_[tool: copy_files]_

_[tool: copy_starter_component]_

_[tool: snip]_

_[tool: write_file]_

_[tool: done]_

_[tool: fork_verifier_agent]_

ASCII 라인이 깨지던 문제를 해결하기 위해 디자인 시스템을 전면 재설계했습니다.

**핵심 변경**
- ASCII 박스 그리기 (`╭─╮│└┘`) 완전 제거 → lipgloss `Border()`를 실제 CSS 1px 라운드 보더로 변환
- 한국어/영문 글자 폭 차이로 인한 정렬 깨짐 → flex/grid 기반 레이아웃으로 해소
- bubbletea/bubbles/huh 컴포넌트의 시각 언어 차용 (huh radio `◆/◇`, bubbles spinner 회전 점, progress 그라디언트 바, table 헤더, 키보드 힌트 바)
- glamour 마크다운 렌더링 스타일로 SPEC 카드 표현

**커버한 19개 화면 (라이트/다크 38개 아트보드)**
- **설치**: install.sh / install.ps1 / install.bat
- **메인**: 배너, help
- **프로젝트**: init (huh 위저드 3/6), doctor, status, version, update
- **런처/자율**: cc, loop start (4단계 페이즈 스트립)
- **지식**: spec view (glamour), worktree list (table)
- **거버넌스/관측**: constitution check (CX 7원칙 49점 만점), mx graph, telemetry report (스파크라인)
- **시스템**: statusline (tmux/vim 임베드), error (거버넌스 위반)

`design_canvas`에 7개 섹션으로 묶어 두었으니 클릭해서 풀스크린으로 확인 가능합니다.

## User

Verifier agent check completed

## Assistant

_[tool: snip]_

_[tool: str_replace_edit]_

_[tool: done]_

_[tool: fork_verifier_agent]_

수정 완료. 아트보드가 정상 렌더링됩니다.

## User

cli에서 구현이 되는 TUI가 맞나? 현재 사용중인 라이브러리로 가능한가?

## Assistant

_[tool: snip]_

