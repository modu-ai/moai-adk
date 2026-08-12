# docs-site 디자인 컴포넌트 + 아이콘 사용 규약 (Claude Warm Editorial)

> CLAUDE.local.md §17.1에서 이관 (2026-08-12 다이어트). docs-site 작업 시 참조.

docs-site는 Claude Warm Editorial 디자인 시스템(코랄 `#cc785c` · Pretendard · Goorm Sans Code · **라이트 단일 테마**)을 따른다. 토큰은 `static/moai-brand.css`(FROZEN) + `static/moai-design.css`, 폰트는 `layouts/partials/head/custom.html`(Pretendard Variable jsdelivr + Goorm Sans Code goorm CDN)에서 로드.

- [HARD] **다크 테마 미사용**: 라이트 단일. `[data-theme="dark"]` 분기는 dead code(신규 추가 금지). 코드 카드의 다크 배경(`#181715`)은 테마가 아닌 컴포넌트 스타일.
- [HARD] **이모지 대신 아이콘**: 본문 콘텐츠의 장식용 이모지(`📖 💡 🚀 ✨ 🎉 🔥 📌` 등) 대신 `{{</* icon <name> [variant] */>}}` shortcode(인라인 SVG)를 사용. 정의: `layouts/shortcodes/icon.html`. variant: `ok|warn|danger|primary|muted`. 아이콘: check, check-circle, x, x-circle, warning, info, bulb, rocket, star, flash, sparkles, target, package, book, search, wrench, database, rotate, clock, arrow-right.
  - **유지(이모지 아님 — 치환 금지)**: 타이포그래피 기호 `→ ← ↓ ✓ ✗`(흐름 서술); MoAI 오케스트레이터 배너 예시 코드블록 내 브랜딩 이모지(`🤖 🗿 📋 🎯` 등 — 출력 스타일 재현 목적). 시맨틱 표 마커(`✅`)는 `{{</* icon check ok */>}}`로 전환 가능하나 의미 보존 필수.
  - 신규/수정 콘텐츠는 본 규약 준수. **기존 콘텐츠 일괄 전환은 의미·4-locale 파리티·브랜딩 보존을 위해 파일별 판단으로 진행(blind sed 금지)**.
- [HARD] **사이드바 아이콘**: `data/menu/main.yaml`의 `icon:` 값은 `layouts/partials/menu.html` SVG switch에 대응 case 필수. 신규 icon 값 추가 시 menu.html에 path case도 함께 추가(미매칭 시 빈 svg 렌더 → 아이콘 누락).
- [HARD] **코드블록**: 모든 fenced code는 `layouts/_default/_markup/render-codeblock.html` render hook이 macOS 다크 카드로 렌더(트래픽라이트·언어 pill·복사 버튼·멀티라인 줄번호). Chroma `noClasses=false` + `.code-card .chroma` 스코프 코랄 신택스. 줄번호는 `lineNumbersInTable=true`. pre 여백은 geekdoc 기본 `padding:0`을 이기기 위해 `!important` 필요.
- [HARD] **Mermaid**: `foot.html`이 CDN UMD(`mermaid@10`, `window.mermaid` 노출)를 로드해 코랄 themeVariables(`theme:'base'`)로 렌더. geekdoc 내장 번들은 `window.mermaid` 미노출이라 사용 금지(기본 라벤더 자체 렌더됨).
- [HARD] **푸터**: geekdoc 기본 `.gdoc-footer` 다크틸 배경을 `.cw-footer` 라이트 캔버스로 오버라이드(`!important`).
- [HARD] **CSS 캐시 버스팅**: custom.html이 `hash.FNV32a (readFile ...)`로 CSS URL에 `?h=` 해시 부여(프로덕션 full build에서 정확). dev `hugo server`는 template 변경 시에만 해시 갱신되므로 CSS만 수정 후 미반영 시 서버 재시작 또는 하드 리로드.
