# t44 Review Findings — 후속 커밋 (`feat/web-shell-chrome` @ 26859cb4b)

- 리뷰 세션: `review-tjqim9` (review 컬럼), 2026-08-15 — 1차 리뷰(`t44-review-findings.md`)의 후속
- 대상: `git diff 7473df823..26859cb4b` — t45 i18n + W1/W2/W3 수정 + `screen_chrome_test.go` 신규
- 검증 방식: 판사 에이전트 미사용(전회 사망 사유 동일 위험 + 시간 압박). 오케스트레이터 타깃 검증 — `i18n.js`는 grep/값 확인만으로 국한.
- 검증 규율: `go test ./internal/web/...` 리드 확인치. 본 세션 실행 없음(VCI §3.4).

## 판정 요약

| 등급 | 개수 | 비고 |
|---|---|---|
| Critical | 0 | — |
| Warning | 1 | W4 (`.finally()` 스탬프 — 경계선, 수정은 1줄) |
| Suggestion | 3 | S5 baseline 부재, S6 panelHelp 키 충돌(리드 인지), S7 잡음 |

**머지 관점 결론: W4를 afterSwap 스탬프로 바꾸면 이의 없음. allowlist 2건은 정당 — 우회가 아니라 계약대로의 사용.**

## 리드 질문 4건에 대한 판정 (증거 포함)

### Q1 🔴 allowlist 예외 2건 — **정당. 승인 권장**

- 메커니즘 확인: `i18n_untranslated_allowlist_test.go` 헤더의 거버넌스 계약(REQ-I18NGOV-001/003/004/005/016/023) — 폐쇄 분류(컴파일 타임), 항목 상한 30, orphan check가 번역된 키의 스테일 항목을 자동 제거, negative-control 테스트가 무허가 미번역을 검증. **allowlist 추가는 검출기 우회가 아니라 계약이 규정한 정상 경로**다.
- `stat.spec`: 4 로케일 전부 값 `"SPEC"` (en/ko/ja/zh 실측). 형제 키는 번역됨(`stat.drift` → 드리프트/ドリフト/漂移) → "산문 몰래 승인"(reviewer assertion b)에 해당하지 않음. SPEC이 디렉터리(`.moai/specs/`)·CLI(`moai spec`)·감사 JSON이 공유하는 식별자라는 논거는 사실과 일치.
- `statNote.must-fix`: 4 로케일 전부 `"MUST-FIX"` — `internal/spec/audit.go:65` `Severity string json:"severity" // MUST-FIX | INFO` 에서 실제 방출되는 토큰(grep 확인). 사용자가 `moai spec audit --json` 출력에서 grep하는 값과 콘솔 배지가 일치해야 한다는 기존 `board.badge.mustfix` 항목의 논리와 동일한 값을 공유 — 일관성 논거 성립.
- 사소한 반론 1개: `stat.spec`의 Reason이 `reasonProperNoun`인데 `statNote.must-fix`처럼 `reasonTechnicalIdentifier`가 더 정확해 보임(분류 네이트, 블로킹 아님).

### Q2 🔴 i18nSlug 고아 키 강등 — **주장 확인 + 1개 갭**

- 확인: `app.js:254-267` — 누락 키 경로는 baseline 복원, baseline 없으면 **기존 텍스트 유지**. 주석 명문: "an element is never blanked". 제목 문구 변경 → 키 고아 → 서버 영어 텍스트가 그대로 남는다. **화면이 빈 적은 없다.**
- 갭(S5): 새 요소들(stat 라벨/노트, panel 제목/help, screen 제목)은 `data-i18n-baseline` 속성이 없다. baseline은 "로케일 재전환 시 이전 로케일 텍스트가 갇히는" 사고를 막는 장치인데 이 요소들은 보호 안 됨 — 키가 한 로케일에만 존재하게 되는 비대칭 추가 시 갇힘 가능. 오늘은 4 로케일 전부 키가 있어(실측) 발생 경로 없음. 권장: 새 5개소에 `data-i18n-baseline={원문}` 추가.

### Q3 🟡 stampRefreshed `.finally()` — **결함 확인 (W4)**

- `app.js` `.finally(function(){ refreshing=false; stampRefreshed(); })` — fetch 거부(네트워크 실패) 시에도 스탬프. 주석은 "when the browser received the swap"라고 말하지만 swap을 받지 못한 실패에서도 시각이 갱신된다. `htmx.ajax`는 4xx/5xx를 promise 거부로 안 할 수 있어 `.then()` 이동도 불완전.
- 권장 수정(1줄): document 레벨 `htmx:afterSwap` 이벤트 리스너에서 스탬프 — 성공 스왑에만 발생. 경계선 경고: 실패 시 실제로는 live 표시가 꺼지는 완화가 있어 사용자 오도 폭은 작다.

### Q4 🟡 panelHelp.sessions 충돌 — **리드 진단 그대로 (S6)**

- 실측: `screens.templ:76` overview `@panel("Sessions", …, "", "roomy")` (help 빈 문자열) vs `:379` monitor `@panel("Sessions", …, "Only PID-confirmed entries are marked active.", "roomy")`. 동일 제목 → `panel.sessions`/`panelHelp.sessions` 공유. overview 쪽은 `widgets.templ`의 `if help != ""` 로 렌더 안 함 → 현재 무해. 미래에 overview에 help 붙으면 monitor 문구 유출 — 리드 기술 그대로. 권장: `panelHelp.<area>.<slug>` 이름공간화 또는 위험 주석을 키 옆에 명시(이미 템플릿 주석에 일부 있음).

## 1차 지적 반영 확인

| 1차 지적 | 반영 확인 |
|---|---|
| W1 RenderedAt 고정 | ✅ 클라이언트 스탬프로 해결 방향 정확 (잔여: Q3/.finally) |
| W2 .nav__en 중복 룰 | ✅ 죽은 신규 룰 삭제, 기존 mono 룰 유지 + 은닉 룰만 추가, 코멘트 갱신 — console.css diff로 확인 |
| W3 무테스트 | ✅ `screen_chrome_test.go` 신규 — `TestScreenCtxChips` + `TestI18nSlug` (107행; 분기 세부는 본 세션에서 행 단위 검증 안 함 — CI가 1차 판정자) |

## Suggestion (잡음)

- **S7**: stat 노트 중 개수 합성 문자열("4 in-progress" 류)은 `statNote.<숫자>-…` 키가 사전에 없어 모든 로케일에서 영어 잔류 — 템플릿 주석에 명시된 의도적 범위 축소. ko UI에서 라벨은 번역·노트는 영어인 혼합 렌더가 오늘 존재함(화장적). 후속 카드에서 message-format 도입 시 해소.

## Gaps (VCI §3.4)

- `go test ./internal/web/...` — 리드 보고 재인용, 본 세션 미실행.
- 생성물(`*_templ.go`) 바이트 동기 — 미검증(존재 수준만).
- 브라우저 4-로케일 실기동 확인 없음(정적 검증 + 키 실측만).

## 종합

- Quality: PASS w/ WARN(경계선 1) — W4
- UX: PASS w/ WARN(경계선 1) — W4
- Design: PASS
- i18n 거버넌스: **PASS — allowlist 2건 승인 권장** (S5 baseline 보강 권장)

리뷰는 보고까지. 수정·머지·PR 편집 없음.
