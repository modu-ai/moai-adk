# MoAI Web Console 재설계 브리프

> 이 문서는 **다른 Claude 세션(또는 Claude Design)에 통째로 전달하기 위한 지침서**입니다.
> 그대로 붙여넣으면 대상 화면, 코드 경로, 지켜야 할 제약, 산출물 요구가 한 번에 전달됩니다.
> 첨부 스크린샷 3장은 같은 디렉터리에 있습니다.

작성 기준 시점: 2026-08-14 / 대상 브랜치: `main` / 로컬 구동 주소: `http://127.0.0.1:3041`

---

## 0. 한 줄 요약

`moai web`이 띄우는 **로컬 루프백 설정 콘솔**을, 모두의AI 서비스 디자인 시스템을 토대로 **백지에서 다시 설계**해 주세요.
토큰·타이포·브랜드 자산은 모두의AI 시스템을 따르되, **레이아웃·정보구조·컴포넌트 어휘는 제약 없이 새로 제안**해도 됩니다.

---

## 1. 이 제품이 무엇인가

- `moai web` (터미널 명령)이 `127.0.0.1:3041`에 뜨는 **1인용 로컬 웹 콘솔**입니다. 외부 배포 웹사이트가 아닙니다.
- 하는 일: `~/.moai` 프로필 환경설정 + 프로젝트 설정(`.moai/config/sections/*.yaml`) 편집, 그리고 읽기 전용 SPEC 보드.
- 사용자: 개발자 본인 한 명. 세션당 수 분 머물며 값을 바꾸고 저장하고 닫습니다.
- 성격: **밀도 높은 설정 편집 도구**에 가깝고, 마케팅 페이지가 아닙니다. 지금 디자인은 문서 사이트의 여백 감각을 그대로 들고 와서 이 성격과 어긋납니다 — 이것이 재설계의 핵심 동기입니다.

### 기술 스택 (변경 불가)

| 층 | 기술 |
|---|---|
| 서버 | Go `net/http` + `ServeMux` |
| 템플릿 | [a-h/templ](https://templ.guide) — `.templ` → `*_templ.go` 코드 생성 |
| 상호작용 | htmx (`hx-boost`) + 순수 `app.js` (프레임워크 없음, 빌드 스텝 없음) |
| 스타일 | 단일 손으로 쓴 CSS 파일 (Tailwind·SCSS·PostCSS 없음) |
| 자산 | 전부 `go:embed`로 바이너리에 내장 |

---

> **⚠️ 범위 확장 안내 (2026-08-14 추가)**
> 이 브리프는 **현행 2화면** 기준으로 작성됐습니다. 이후 메뉴 자체를 5개 영역(개요·칸반·SPEC·모니터·설정)으로
> 확장하기로 결정했습니다. 정보구조·화면 목록·데이터 원천은 짝 문서 **`moai-web-menu-spec.md`**가 정본입니다.
> 시각 재설계를 할 때는 **두 문서를 함께** 읽고, 아래 인벤토리는 "현행 상태"로만 참고하세요.
> 새로 들어오는 화면 유형이 만드는 시각적 요구는 §5.5에 정리했습니다.

## 2. 화면 인벤토리

| # | 화면 | 라우트 | 렌더 소스 | 비고 |
|---|---|---|---|---|
| 1 | 콘솔 본체 (설정 편집) | `GET /` | `internal/web/root.templ` | 프로필 바 + 10개 탭 + 저장 바 |
| 2 | SPEC 보드 (읽기 전용) | `GET /specs` | `internal/web/board.templ` | 요약 배지 + 리스트 |
| — | 저장 | `POST /save` | — | 폼 제출 대상 |
| — | 프로필 생성/삭제/이름변경 | `POST /profile/create`, `/profile/delete`, `/profile/rename` | `root.templ` 내 프로필 바 | |
| — | GLM 키 노출 | `POST /glm-key/reveal` | `internal/web/glmkey.go` | |
| — | 서버 종료 | `POST /__shutdown__` | 앱바 전원 버튼 | |
| — | 정적 자산 | `GET /static/*` | `internal/web/assets.go` | 임베드 FS |

### 콘솔 탭 10개 (순서 고정 — `internal/web/schemaform.go:34` `consoleTabs()`)

`Identity` · `Language` · `LLM` · `3rd Party LLM` · `Workflow` · `Git & Worktree` · `Audit` · `Agents` · `Report` · `MCP`

탭별 필드 수는 크게 다릅니다. Identity는 필드 1개, Agents는 11개 에이전트 × 2개 셀렉트, Statusline 계열은 17개까지 갑니다. **한 화면 안에서 밀도 편차가 10배 이상**이라는 점이 레이아웃 설계의 가장 큰 변수입니다.

---

## 3. 코드베이스 경로 지도

모든 경로는 리포지토리 루트(`/Users/goos/MoAI/moai-adk-go`) 기준입니다.

### 3.1 화면을 그리는 곳 — 여기를 고치면 화면이 바뀝니다

```
internal/web/root.templ        (330줄)  콘솔 본체: <head>, 앱바, 프로필 바, 탭 nav, 저장 바
internal/web/board.templ       (139줄)  SPEC 보드 페이지
internal/web/page.templ        (203줄)  기본 위젯: select / toggle / number / 에러 표시
internal/web/fieldsets.templ   (814줄)  탭 안쪽 패널·필드셋 — 분량이 가장 큼
internal/web/icons.templ        (54줄)  인라인 SVG 아이콘 세트 (CDN 없음)
```

> `.templ` 파일을 고치면 반드시 `make templ-generate` 또는 `make build`를 돌려 `*_templ.go`를 재생성해야 합니다.
> 생성물(`root_templ.go` 등)은 **직접 편집하지 마세요** — 다음 생성 때 덮어씁니다.
> `internal/web/root.templ.backup`은 과거 잔재이며 무시하세요.

### 3.2 스타일·스크립트·자산

```
internal/web/assets/console.css   (1041줄)  ★ 디자인 토큰 + 전체 컴포넌트 CSS (단일 파일)
internal/web/assets/app.js         (520줄)  탭 전환, 폼 상호작용, 종료 버튼
internal/web/assets/i18n.js       (2028줄)  ★ 4개 언어 사전 (en/ko/ja/zh) — data-i18n 키로 치환
internal/web/assets/htmx.min.js             벤더 번들 (수정 금지)
internal/web/assets/fonts/*.woff2           Pretendard 5종 + Noto Sans CJK(JP/SC) 6종 + Goorm Sans Code
internal/web/assets/mascots/*.png           마스코트 6포즈 (thinking / explaining / searching / pointing / teaching / coffee)
```

### 3.3 데이터·서버 (디자인 작업에서 건드리지 않음)

```
internal/web/app.go            라우트 테이블 + 루프백/동일출처 미들웨어
internal/web/server.go         서버 기동·종료
internal/web/handlers.go       핸들러
internal/web/schemaform.go     ★ 탭 정의 · 패널 매핑 · 아이콘 지정
internal/web/viewmodel*.go     뷰모델
internal/settings/schema*.go   설정 스키마(섹션·필드 정의)의 원천
internal/cli/web.go            `moai web` 명령
```

### 3.4 디자인을 잠그고 있는 테스트 (반드시 읽고 시작)

```
internal/web/restyle_test.go            폰트 self-host, 다크테마 부재, 인라인 SVG, 접근성 단서, 상태 토큰
internal/web/mascots_test.go            마스코트 포즈 임베드 + 앱바 배지는 thinking 포즈
internal/web/tab_layout_test.go         탭 순서, 모든 탭에 패널 존재
internal/web/i18n_governance_test.go    data-i18n 키 4개 언어 전량 커버(정방향/역방향)
internal/web/widget_contract_test.go    위젯 마크업 계약
internal/web/console_ux_fix_test.go / webux_followup_test.go / appbar_context_test.go 등
```

이 테스트들은 **재설계를 막으려는 것이 아니라, 깨지면 안 되는 불변식을 지키는 장치**입니다.
새 디자인이 불변식 자체를 바꿔야 한다면, 테스트를 함께 고치고 **왜 바꿨는지 근거를 남기세요** (예: 마스코트 포즈 교체).

---

## 4. 디자인 시스템의 원천

### 4.1 SSOT 파일

```
docs-site/static/moai-brand.css   (2181줄)  ★ 모두의AI 디자인 시스템 정본 (FROZEN)
docs-site/static/moai-docs-theme.css (898줄)  문서 사이트 적용층 (참고용)
```

`moai-brand.css`의 `:root` 블록이 **토큰 정본**입니다. 새 디자인은 여기서 출발하세요.

### 4.2 정본 토큰 (v2-renewal, 2026-08-11 재동결)

```
브랜드    --color-primary #3d7d5f  (hover #316750 / active #265240)
          --color-ink     #060606
          --color-bg      #f4f4f4   ← 순백 #fff 대체 금지
          --color-surface #ffffff
중립 램프  --neutral-50 #f4f4f4 → 950 #060606 (마스코트 그레이 파생, --neutral-400 #9fa0a0가 기준점)
상태      --color-success #2e8a63 / --color-warning #c47b2a / --color-danger #c44a3a / --color-info #2a8a8c
전경      --fg-1 ink / --fg-2 #3d3d3d / --fg-3 #6c6c6c / --fg-on-primary #ffffff
경계      --border-1 #e6e6e6 / --border-2 #ebebeb / --border-strong #d3d3d3
타이포     --font-sans Pretendard / --font-mono "Goorm Sans Code"
          자간은 음수 기본 (--tracking-body -0.025em, --tracking-heading -0.05em)
스케일     --text-xs 0.75rem … --text-6xl 3.75rem, --space-1 0.25rem … --space-32 8rem
반경      --radius-sm 4 / md 8 / lg 16 / xl 24 / pill 32 / full, --radius-card 12px
```

### 4.3 ⚠️ 확인된 토큰 드리프트 — 재설계 때 정리해 주세요

`internal/web/assets/console.css`는 **문서 사이트가 이미 갱신한 상태 색을 따라가지 못했습니다.** 두 파일을 직접 대조해 확인한 사실입니다.

| 토큰 | 정본 (`moai-brand.css`) | 현행 콘솔 (`console.css`) |
|---|---|---|
| `--color-success` | `#2e8a63` | `#5db872` (구버전) |
| `--color-warning` | `#c47b2a` | `#d4a017` (구버전) |
| `--color-danger` | `#c44a3a` | `#c64545` (구버전) |
| `--color-info` | `#2a8a8c` | `#5db8a6` (구버전) |
| `--fg-2` | `#3d3d3d` | `#565656` |
| `--border-1` | `#e6e6e6` | `#d1d1d1` |

더 문제는 `internal/web/restyle_test.go`의 `TestStatusTokensDocsSiteParity`가 **구버전 값을 "docs-site parity"라는 이름으로 고정**하고 있다는 점입니다(`restyle_test.go:550`). 정본을 따라가려면 이 테스트의 기대값도 함께 갱신해야 하며, 대비비(AA) 카브아웃(`--status-text-*` color-mix)이 새 색에서도 성립하는지 다시 확인해야 합니다.

---

## 5. 지금 화면의 문제 (재설계 출발점)

첨부 스크린샷을 함께 보세요.

- `current-01-console.png` — 콘솔 첫 화면 (Identity 탭, 필드 1개)
- `current-02-specboard.png` — SPEC 보드
- `current-03-agents-tab.png` — Agents 탭 (밀도가 가장 높은 화면)

관찰된 문제:

1. **문서 사이트 여백을 설정 도구에 그대로 이식했습니다.** `.page { max-width: 880px }`(console.css:412)로 묶여 있어 1440px 화면에서 좌우가 크게 비고, Identity 탭은 화면 하나를 필드 하나에 씁니다.
2. **프로필 바가 상시 최상단을 점유합니다.** 생성·이름변경·삭제 폼 3개가 항상 펼쳐져 있어, 세션당 몇 번 쓰지 않는 기능이 매번 쓰는 설정 영역을 밀어냅니다.
3. **탭 10개가 두 줄로 접힙니다.** MCP 하나만 아래로 떨어져 정렬이 깨집니다.
4. **밀도가 화면마다 널뜁니다.** 같은 카드·같은 간격 규칙을 필드 1개짜리 탭과 22개짜리 탭에 똑같이 적용합니다.
5. **앱바 정보가 비활성 텍스트 덩어리입니다.** `lang / model / effort / dev` 문맥이 한 줄 모노스페이스로 붙어 있어 읽기도 조작하기도 어렵습니다.
6. **저장 바가 페이지 맨 아래에 있습니다.** 긴 탭에서는 스크롤을 끝까지 내려야 저장할 수 있습니다.
7. **위계가 평평합니다.** 섹션 제목·필드 라벨·설명문·키 배지가 비슷한 크기로 붙어 있어 훑어보기가 어렵습니다.

---

### 5.5 새 화면 유형이 만드는 요구 (MoAI-Kanban 반영)

메뉴 확장으로 **성격이 전혀 다른 화면**이 들어옵니다. 지금 CSS에는 이를 표현할 어휘가 없으므로, 디자인 시스템에 새로 정의해야 합니다.

**MoAI-Kanban**은 하나의 작업을 5개 터미널 세션(Lead·Plan·Run·Review·Sync)에 나눠 태우는 개발 방법론입니다. 역할마다 모델과 백엔드를 다르게 걸어(판단이 걸린 자리는 Opus, 분량이 많은 자리는 GLM) 토큰 비용을 줄입니다. 칸반 화면은 그 **5개 세션이 지금 어떤 상태인지**를 비춥니다.

새로 필요한 시각 어휘:

| 어휘 | 쓰이는 곳 | 지금 있나 |
|---|---|---|
| **생존 표시** — 활성 / 미기동 / 응답없음 | 세션 열, 세션 목록 | 앱바 루프백 점 하나뿐 |
| **역할 열(column)** — 5개 역할을 나란히 | 칸반 뷰 A | 없음 |
| **백엔드 배지** — claude(정액) / glm(종량) | 세션 열, 개요 | 없음 |
| **진행 게이지** — 컨텍스트 사용률 % | 세션 열 | 없음 |
| **단계 표시** — 완료 / 작업중 / 대기 / 막힘 | 칸반, 개요 | 배너 4종만 있음 |
| **경고 상태** — 미기동 역할, 정체된 목표 | 개요, 칸반 | 배너로만 가능 |
| **추정값 표기** — 단정이 아니라 추정임을 드러냄 | 단계 상태, 세션 활성 | 없음 |

특히 마지막 항목을 강조합니다. 세션 레지스트리에는 죽은 프로세스의 항목이 남을 수 있고 단계 상태도 추정으로 시작합니다. **확인된 사실과 추정을 시각적으로 구분**하지 않으면 화면이 조용히 거짓말을 합니다. 확정값과 추정값의 표현을 다르게 설계해 주세요.

밀도 요구도 달라집니다. 설정 화면은 세로로 긴 폼이지만, 칸반 뷰 A는 **가로 5열을 한 화면에 담아야** 합니다. 지금의 `max-width: 880px`로는 불가능합니다.

---

## 6. 재설계 지침

### 6.1 범위 — 백지 재해석

다음을 **자유롭게 다시 제안**해 주세요.

- 전체 레이아웃 (사이드바 / 2단 / 마스터-디테일 / 전폭 등 어떤 구조든)
- 정보구조 (탭 10개를 유지할지, 그룹으로 묶을지, 검색·필터를 둘지)
- 컴포넌트 어휘 (카드·필드 행·토글·셀렉트·배지·배너의 형태를 새로 정의)
- 밀도 체계 (compact / comfortable 같은 단계를 도입해도 됩니다)
- 저장·프로필·상태 표시의 배치와 상호작용 모델
- 마스코트를 어디에, 얼마나 쓸지 (지금은 앱바 1 + 페이지 헤드 1)

### 6.2 계속 지켜야 할 것

- **브랜드 정체성**: `moai-brand.css`의 토큰 어휘(§4.2)를 기준으로 삼습니다. 필요하면 토큰을 **추가**하되, 브랜드 3색(primary / ink / bg)은 유지합니다.
- **타이포그래피**: Pretendard(본문) + Goorm Sans Code(모노). 새 폰트를 도입하면 self-host woff2 서브셋을 직접 만들어야 하므로, 도입하려면 그 비용을 명시하고 제안하세요.
- **라이트 단일 테마**: 다크 테마는 정책적으로 폐지되었고 `TestDarkThemeAbsence`가 이를 잠급니다.

### 6.3 설계 목표

1. 자주 쓰는 것을 앞에 — 세션 대부분은 값 몇 개를 바꾸고 저장합니다.
2. 밀도를 내용에 맞게 — 필드 1개 탭과 22개 탭이 같은 규칙을 쓰지 않도록.
3. 저장 상태를 항상 보이게 — 변경됨/저장됨/실패를 스크롤 위치와 무관하게 알 수 있게.
4. 훑어보기 가능한 위계 — 섹션 → 필드 → 설명 → 키의 4단계가 크기·색·간격으로 구분되게.
5. 넓은 화면을 쓰되 줄 길이는 제어 — 전폭을 쓰더라도 설명문 한 줄이 지나치게 길어지지 않게.

---

## 7. 절대 제약 (불변식)

깨면 테스트가 실패하거나 제품이 망가집니다.

| # | 제약 | 근거 |
|---|---|---|
| C1 | **네트워크 요청 0.** CDN 폰트·CSS·JS·이미지 금지. 모든 자산은 `internal/web/assets/`에 두고 `/static/*` 상대 경로로 참조. | 오프라인 불변식, `restyle_test.go` |
| C2 | **라이트 단일 테마.** `[data-theme="dark"]`·`prefers-color-scheme` 다크 블록 금지. | `TestDarkThemeAbsence` |
| C3 | **아이콘은 인라인 SVG.** 아이콘 폰트·외부 스프라이트 금지. | `TestInlineSVGIconsNoCDN`, `icons.templ` |
| C4 | **i18n 4개 언어 유지.** 새로 넣는 사용자 문구는 `data-i18n="키"`를 달고 `i18n.js`의 en/ko/ja/zh **네 곳 모두**에 항목을 추가. 정방향·역방향 커버리지 테스트가 있습니다. | `i18n_governance_test.go` |
| C5 | **폼 계약 불변.** `<input>/<select>`의 `name` 속성, 폼 `action`, POST 라우트를 바꾸지 마세요. 서버는 `name` 기준으로 저장합니다. | `TestNameAttributesPreserved` |
| C6 | **탭 순서·패널 대응 유지** (구조 자체를 재설계한다면 `schemaform.go`와 테스트를 함께 갱신하고 근거를 남길 것). | `tab_layout_test.go` |
| C7 | **접근성 유지 이상.** `role="tablist"`/`aria-selected`/`aria-invalid`/`aria-describedby`/라벨 연결을 유지하고, 대비는 AA를 만족시킬 것. | `TestAccessibilityCues` |
| C8 | **빌드 스텝 추가 금지.** Node·번들러·Tailwind 도입 불가. CSS는 손으로 쓴 단일 파일, JS는 순수 스크립트. | 프로젝트 정책 |
| C9 | **`internal/web`은 배포 템플릿이 아닙니다.** `internal/template/templates/` 아래가 아니므로 Template-First 미러 규칙은 적용되지 않습니다. 단 `.templ` 수정 후 코드 생성은 필수. | `CLAUDE.local.md` §2 |

---

## 8. 산출물 요구

다음 순서로 내주세요.

1. **진단** — 현행 화면의 문제를 스크린샷 기준으로 짚고, 무엇을 바꿀지 한 문단.
2. **방향 제안 2~3안** — 각 안의 레이아웃 골격을 ASCII 또는 간단한 도식으로. 장단점과 위험을 함께.
3. **선택안의 디자인 시스템 정의**
   - 토큰 델타: `moai-brand.css` 정본 대비 추가·변경한 토큰과 그 이유
   - 컴포넌트 스펙: 필드 행 / 섹션 / 탭 / 배지 / 배너 / 버튼 / 저장 바의 상태별(기본·호버·포커스·비활성·오류) 정의
   - 밀도·간격 규칙
4. **화면 설계** — 최소 5화면: 개요 · 칸반(5역할 체인 보드) · SPEC 목록 · 설정(밀도 낮은 탭 / 높은 탭 각 1개).
   칸반은 §5.5의 새 어휘가 실제로 어떻게 보이는지 확인하는 자리이므로 빠뜨리지 마세요.
5. **구현 계획** — 어떤 파일을 어떤 순서로 고칠지, 함께 갱신해야 할 테스트 목록 포함.
6. **§4.3 토큰 드리프트 처리 방안** — 정본 값으로 맞출지, 새 팔레트로 대체할지와 그 근거.

한 번에 코드를 다 쓰지 말고, **3번(디자인 시스템 정의)까지 먼저 내고 확인을 받은 뒤** 구현으로 넘어가 주세요.

---

## 9. 검증 방법

```bash
# 1) templ 코드 생성 (.templ 수정 후 필수)
make templ-generate

# 2) 웹 패키지 테스트 — 디자인 불변식이 여기서 걸립니다
go test ./internal/web/... ./internal/settings/...

# 3) 전체 스위트 (교차 회귀 확인)
go test ./...

# 4) 실제 화면 확인
make build && ./bin/moai web     # http://127.0.0.1:3041
```

시각 확인은 `http://127.0.0.1:3041/` 와 `/specs` 두 라우트를 1440×1000 및 좁은 폭에서 각각 보세요.

---

## 10. 첨부

| 파일 | 내용 |
|---|---|
| `current-01-console.png` | 콘솔 첫 화면 (Identity 탭) — 여백 과다, 프로필 바 점유 |
| `current-02-specboard.png` | SPEC 보드 |
| `current-03-agents-tab.png` | Agents 탭 — 실제 밀도, 탭 두 줄 접힘 |

---

## 부록 A. 참고 SPEC

이 콘솔의 디자인 이력은 다음 SPEC들에 남아 있습니다. 배경이 필요할 때만 읽으세요.

```
.moai/specs/SPEC-WEB-CONSOLE-004/    시각 리스타일 + 오프라인 폰트 불변식
.moai/specs/SPEC-WEB-CONSOLE-005/    ja/zh CJK 폰트 서브셋
.moai/specs/SPEC-WEB-CONSOLE-006/    (design.md 포함)
.moai/specs/SPEC-WEB-CONSOLE-011/    프로필 CRUD + SPEC 보드
.moai/specs/SPEC-DESIGN-DOCS-V31-001/  모두의AI 토큰 v2-renewal 재동결 (정본 근거)
```
