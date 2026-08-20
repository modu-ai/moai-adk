# t44 Review Findings — web shell chrome (`feat/web-shell-chrome` @ 7473df823)

- 리뷰 세션: `review-tjqim9` (review 컬럼), 2026-08-15
- 대상: `git diff 6ff7c3aaf..7473df823` — `internal/web/screens.go` · `shell.templ` · `shell_templ.go`(생성물) · `assets/console.css`
- 렌즈: `--design` (리드 지정; `--security --deep`은 동일 트리에 `.moai/reports/security-deepscan-20260815T100639Z/` 로 완료)
- 특이사항: **작성자 = 칸반 리드 세션** (run 레인 Bash 차단으로 리드가 직접 구현) → 독립성 강화 모드로 리뷰. 병렬 판사 에이전트 2인(design/quality)이 모두 context window 한계로 실패(아래 "판사 실패 기록")하여, **오케스트레이터가 타깃 검증으로 직접 수행**. 모든 지적은 인용 가능한 증거로 뒷받침됨.
- 검증 규율: 풀 스위트 미실행(리드 지시). `go build/vet/test ./internal/web/...` 은 리드가 exit 0 확인(본 리뷰에서 재실행 안 함 — VCI §3.4 Gaps 참조).

## 요약 판정

| 항목 | 판정 |
|---|---|
| Critical | 0건 |
| Warning | 3건 (W1 RenderedAt 고정, W2 CSS 중복 룰, W3 신규 로직 무테스트) |
| Suggestion | 4건 |
| 전체 | **머지 가능하나 W1·W2는 sync 전 처리 권장** — 둘 다 수 라인 수정 |

리뷰는 보고까지가 범위: 본 세션은 수정하지 않았고, PR 개설도 하지 않았다(sync 컬럼 소관).

## Warnings

### W1. [UX] `RenderedAt`가 최초 페이지 로드에 영원히 고정 — "Live" 옆의 시각이 거짓 프레시함을 암시

- 증거: `app.js:625` `refresh(area)` 는 htmx로 현재 URL을 재요청하되 `select: ".body"`, `swap: "outerHTML"` — **topbar는 교체 대상이 아님**. SSE(`EventSource("/events")`) 이벤트마다 `.body`만 새로고침되고, `liveState` 가 살아 있는 `<header>` 는 서버 렌더 시 1회만 그려진다. `data-live-rendered-at` 속성(`shell.templ` liveState)을 참조하는 JS는 **전무** (`grep data-live-rendered-at internal/web/assets/*.js` → 0건 — 그 속성은 그런 의도의 미배선 훅으로 보임).
- 결과: 탭을 오래 열어 두면 데이터는 계속 새로고침되는데 "Live ⋅ 14:22:31" 의 시각은 로드 시각에 멈춰 있다. 시안의 `실시간 14:22:31 갱신` 은 "데이터 마지막 갱신 시각"으로 읽는 것이 자연스럽다.
- 리드의 해석(서버 렌더 시각)에 대한 답: **최초 로드 순간에만 성립**. 그 이후로는 정체 시각이 "Live" 레이블과 나란히 거짓을 말한다. 시각이 없었을 때보다 나쁘다.
- 권장: SSE refresh 경로(`refresh()` 성공 콜백 또는 각 이벤트 리스너)에서 `data-live-rendered-at` 스팬의 텍스트를 클라이언트 시각으로 갱신 — 미배선 속성이 정확히 그 지점으로 보인다.

### W2. [DESIGN/CRAFT] `.nav__en` 룰이 중복 정의 — 신규 스타일의 상당 부분이 캐스케이드에서 사실상 죽음

- 증거: `console.css` 에 인접한 두 룰 —
  - `:110` (신규) `.nav__en{margin-left:auto;font:var(--text-caption);color:var(--fg-3);font-weight:400}`
  - `:111` (신규) `html[lang="en"] .nav__en{display:none}`
  - `:112` (기존) `.nav__en{margin-left:auto;font-family:var(--font-mono);font-size:10.5px;color:var(--fg-3)}`
- 캐스케이드 분석: 동일 specificity(0,1,0)에서 **뒤에 오른 112가 이김** → `font-family:var(--font-mono)`, `font-size:10.5px` 가 신규 `font:` shorthand의 대응 성분을 덮어쓴다. 110에서 실제로 살아남는 것은 `font-weight:400` (과 margin/color — color는 어차피 동일값). 즉 의도한 caption 스타일링은 적용되지 않고 에코는 기존 mono 10.5px 로 렌더된다.
- 권장: 두 룰을 하나로 병합(기존 112를 수정하는 형태). `html[lang="en"]` 은닉 룰은 그대로 두어도 됨(동작함).

### W3. [QUALITY/TRUST-Tested] 신규 로직에 테스트 0건

- 증거: diff의 4파일 중 `_test.go` 변경 없음. `screenCtxChips` (빈값 필터링 + nil-seam 가드 + 오류 삼킴 분기)와 `RenderedAt` 주입은 전부 무테스트. `app.go:39/61` 의 read seam은 주입 가능하라 테스트 작성이 저렴하다.
- 권장: 최소한 `screenCtxChips` 의 3분기(정상 / 오류 삼킴 / 빈값 필터)에 대한 표 테스트 1개.

## Suggestions

### S1. [QUALITY] 칩 오류 삼킴 — 관측 가능성만 보태라 (리드 질문 1에 대한 답)

- 비교 증거: settings 쪽 `buildIndexView`(`handlers.go:253-263`)는 **동일한 read 실패로 500** 을 낸다. 4개 화면은 칩 없이 정상 렌더. 같은 실패가 화면마다 다르게 보이는 비대칭은 존재한다.
- 판정: 리드의 방어 **대체로 타당하며 Q5 부류가 아니다**. Q5는 "감사 실패를 정상으로 렌더" — 안전신호가 거짓 OK를 말하는 사건이었다. 칩 부재는 아무 주장도 하지 않는 부재다(빈값 필터가 "미설정"도 칩 없음으로 만들므로 해석 여지가 동일함). settings가 500인 것도 합리적 — 거기서 prefs는 칩이 아니라 **페이지 본문**이다.
- 단, 지금은 **완전 무로그** → 실패가 발생했는지조차 알 수 없다. `slog.Debug` 한 줄씩만 보태면 비대칭이 "조용한" 상태는 해소된다.

### S2. [DESIGN] `rail__word` "모두의AI" — 4개 로케일 전부 한국어 하드코딩

- 증거: `shell.templ` 에 `<span class="rail__word">모두의AI</span>` 정적. `i18n.js` 에 rail 관련 키는 `rail.profile`/`rail.project`/`rail.shutdown` ×4 로케일이 존재하나 **워드마크 키는 없음** — 레일에서 유일한 비번역 문자열.
- 판정: 브랜드 워드마크의 원어 고정은 정당한 선택일 수 있으나(제품 한국어 명칭), 지금은 **결정이 기록되지 않은 침묵**이다. 의도적이라면 코멘트 한 줄로, 아니라면 키 추가로.

### S3. [DESIGN] `.live__ts` tabular-nums vs `.save__msg`

- `console.css:148` `.save__msg` 도 시각 문자열(`vm.SavedAt`, shell.templ:265)을 렌더하나 `font-variant-numeric:tabular-nums` 없음. 새 `.live__ts`(:153)만 있다. 시각 표기엔 양쪽 다 적용이 일관적.

### S4. [QUALITY] 소소한 것들

- `filepath.Base("")` → `"."` — ProjectRoot 공백 시 crumb이 "." (실서버에선 발생 비현실적, 화장적).
- `screenCtxChips`/`settingsCtxChips` 의 빈값 필터 루프(~8행) 중복 — 지금 크기는 허용 가능, 제3 호출자 생기면 헬퍼로.

## Verified OK (근거와 함께)

| 항목 | 근거 |
|---|---|
| `html[lang="en"]` 전제 (리드 질문 2) | 서버는 항상 `lang={vm.Lang}` = "en" 렌더(`shell.templ:92`, `shellVM`/`settingsShellVM` 모두 `Lang:"en"` 하드코딩) → 초기 로드 시 에코는 CSS로 은닉. `app.js:235` `document.documentElement.setAttribute("lang", locale)` 가 전환 시 갱신. `nav.*` 키 4종 ×4 로케일 존재 (grep `uniq -c` 확인). en에서 "Overview Overview" 는 렌더되지 않으며, JS 무효 시에도 정상 강등. 4 로케일 전부 성립. |
| 반응형 접힘 (리드 질문 4) | `console.css:127` `@media (max-width:1200px)` 가 `.nav__en` 포함 `span:not(.count)` 은닉 — `:111` 은닉 룰과 조합 이상 없음. 단 ≤1200px 에서는 라벨까지 접히는 아이콘 전용 레일(기존 동작). |
| 생성물 동기 (존재 수준) | `shell_templ.go` 에 `rail__word`/`nav__en`/`live__ts` + `liveState(vm.Live, vm.RenderedAt)` 호출 확인. **바이트 동기는 미검증** — `templ generate` 를 스크래치에서 돌리지 않았음(읽기 전용 리뷰). CI가 판정자. |
| MX 태그 | 신규 exported 심볼 없음(구조체 필드 1 + unexported 메서드 1) — 의무 트리거 없음. |
| aria | 에코 스팬 `aria-hidden="true"` — 스크린리더가 영어 에코를 이중 낭독하지 않게 하는 올바른 처리. |

## 판사 실패 기록 (재발 방지용)

design/quality 판사 에이전트 2인 모두 `API Error: The model has reached its context window limit` 로 사망. 추정 원인: `internal/web/assets/i18n.js` 가 **155KB** — 전체 파일 리드 시 일반 서브에이전트 윈도우 초과. 리드가 겪은 선행 "--design 렌즈 에이전트 무응답"과 동일 원인으로 추정. 이 표면의 향후 에이전트에는 "i18n.js 는 grep/행 범위로만 읽을 것"을 지시에 명시할 것.

## Gaps (VCI §3.4)

- `go build/vet/test` exit 0 — 리드 보고 재인용, 본 세션 미실행.
- `shell_templ.go` 바이트 동기(`templ generate` 멱등성) — 미검증.
- 실브라우저 4-로케일 전환·≤1200px 렌더 — 코드 정적 검증만, 화면 확인은 리드의 한국어 캡처(`e2e/screenshots/t44-overview-ko.png`, 워크트리 `.moai/` 상태 기인 표시는 무해함을 확인)뿐.

## 종합

- Security: N/A (deep scan 아티팩트 별도 존재; 본 렌즈 범위 외)
- Quality: **PASS w/ WARN** (W3 무테스트, S1 무로그)
- UX: **PASS w/ WARN** (W1 거짓 프레시함)
- Design: **PASS w/ WARN** (W2 캐스케이드 결함, S2 워드마크 결정 미기록)
- TRUST 5: Tested ✗(W3) · Readable ✓ · Unified ✓ · Secured N/A · Trackable ✓

W1·W2는 각각 수~십 수 라인 수정이므로 sync 전 run 패스에서 처리 가능한 크기. 본 리뷰는 보고까지.
