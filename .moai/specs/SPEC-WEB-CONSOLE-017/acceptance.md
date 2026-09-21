# SPEC-WEB-CONSOLE-017 — Acceptance

카드 t1051 (issue #1709). REQ-A = 사용자 표면(REQ-WC-017-001/002), REQ-B = 유지보수자 표면(REQ-WC-017-003..006). 두 축은 절대 병합되지 않는다.

---

## §D.1 AC Matrix

### 축 1 — REQ-A (사용자 표면: 인라인 실패 사유)

**AC-WC17-001** — (release-blocking, RED) 인라인 슬롯 도달 + 본문-문구 판별식.
- **Given** 설정 페이지가 렌더돼 있고 저장이 제출되지 않았으며, **When** `recordingSeams` 확장 하니스로 어느 한 persistence seam 을 실패시켜 저장을 제출하면, **Then** (a) 제출 **전** DOM 에는 §5.1 표의 어느 seam 실패 문구도 존재하지 않고, (b) 실패 응답 처리 **후** 사용자 표면의 `save__msg save__msg--error` alert 요소(`shell.templ:274`) 안에 그 seam 의 실패 문구가 존재하며, (c) 판정은 **문구 본문 텍스트**로 하고 배너 CSS 클래스와 오류값(`err.Error()` 파생 문자열)으로 하지 않는다.
- **RED 근거(현재 코드)**: 실패 응답은 `handlers.go:648` 이 그리는 500 풀페이지고 폼은 `root.templ:56` `hx-boost="true"` 라 비-2xx 본문이 스왑되지 않는다 — 문구가 슬롯에 도달하는 경로가 없다. (a)는 오늘 이미 참이지만 (b)가 오늘 거짓이므로 RED 다; (a)는 가드의 반대 방향 변이다(§D.2).
- **반증 명령 형태**: `go test -run TestSaveFailureReasonReachesInlineSlot ./internal/web/...`

**AC-WC17-002** — (release-blocking, RED) 9-seam 문구 커버 — 두 표면 모두.
- **Given** §5.1 표의 9개 seam 각각에 대해, **When** 그 seam 을 강제로 실패시키면, **Then** 인라인 슬롯(REQ-A)과 stderr 로그(REQ-B) **양쪽**에 그 seam 의 층-식별 문구/토큰이 나타난다 — 9개 전부, 누락 없이. 표의 문구가 016 개정으로 바뀌면 가드 리터럴을 같은 커밋에서 갱신한다(spec.md §6-3).
- **RED 근거**: stderr 쪽은 오늘 0행(§D.1 AC-WC17-003 RED 근거)이고 inline 쪽은 AC-WC17-001의 RED 근거로 도달 불가 — 9 seam 중 어느 것도 양쪽을 충족하지 못한다.
- **반증 명령 형태**: `go test -run TestSaveFailureSeamCoverage ./internal/web/...` (표 기반 서브테스트 9건)

### 축 2 — REQ-B (유지보수자 표면: stderr 층 로그)

**AC-WC17-003** — (release-blocking, RED) stderr 존재·접두어 일원화·성공 침묵.
- **Given** 설정 저장이 진행되고, **When** 저장이 어느 seam 에서든 실패하면, **Then** (a) stderr 에 저장-실패 행이 **1행 이상** 기록되고 — 0행은 통과가 아니다 — (b) 그 **모든** 행이 `^moai web: ` 접두어로 시작하며 실패한 층을 식별하고, (c) **When** 저장이 성공하면 **Then** 저장-실패 stderr 행이 **정확히 0행**이다. (a)+(b)를 한 AC 로 묶는다 — 0-카운트 단독 검사는 `0+0=0` 동어반복이라 통과를 주장할 수 없다.
- **RED 근거**: 현재 `internal/web` 의 `os.Stderr` 기록은 2곳 전부 비-저장 경로(`server.go:252` 파일 감시, `:290` 브라우저 열기) — 저장 실패 시 stderr 행이 0이므로 (a)가 오늘 거짓이다.
- **반증 명령 형태**: `go test -run TestSaveFailureStderrLog ./internal/web/...` (stderr 캡처 하니스)

**AC-WC17-004** — (regression-guard) 자격증명 비-유출 (sentinel).
- **Given** GLM/Jev 키 필드에 sentinel 값을 넣어 저장을 제출하고, **When** 어느 seam 에서든 실패하면, **Then** stderr 행과 인라인 실패 메시지 **어느 쪽에도** sentinel 값이 나타나지 않는다(HARD-3, REQ-WC-017-005; SPEC-GLM-KEY-INPUT-001 REQ-GKI-004-003 연속).
- **분류 근거**: stderr 표면이 존재하기 전(AC-WC17-003 착지 전)에는 이 AC 를 RED 로 만들 입력이 없다 — `verification-completeness.md` §2.1의 undecidable disposition 에 따라 **regression-guard** 로 분류하며 release-blocking 으로 기록하지 않는다. AC-WC17-003 착지 후 RED-가능 검증으로 전환된다.

### 비-회귀

**AC-WC17-005** — (regression-guard) 성공 경로 무변경.
- **Given** 모든 seam 이 성공하는 저장을 제출하면, **When** 응답이 처리되면, **Then** 배너는 `Settings saved.` + `BannerKind` ok(`handlers.go:597-601`), 인라인 슬롯은 saved 상태, seam 실패 문구와 저장-실패 stderr 행은 0 이다(REQ-WC-017-006).
- **분류 근거**: 오늘 이미 참인 성질의 고정 가드 — 본 SPEC 이 그것을 깨지 않음을 M1-M3 각 착지에서 재단정한다.

## §D.2 반증력 규칙 — "지금 통과하는 AC 는 AC 의 결함이다"

- 각 release-blocking AC 는 위 RED 근거를 갖는다(채용 2-셀: RED-now + green path — `verification-completeness.md` §2). RED 근거가 없는 AC 는 채용하지 않는다.
- **가드 판별식은 본문 문구(HARD-4)**. 클래스 단정 금지의 근거: `fieldsets.templ:607` 이 `banner banner--warn` 을 제출 전부터 렌더하므로 클래스 존재 단정은 공허하게 참이고, `banner--error` 는 `root.templ:29-34` 기준 존재하지 않는다.
- **양방향 원칙**: 사전-부재(제출 전 문구 없음)와 사후-존재(실패 후 문구 있음)를 모두 단정한다 — 한 방향만 돌리면 공허 참(부재만 검사하면 아무것도 안 잡으면서 통과) 또는 눈먼 패턴(존재만 검사하면 대조 없는 초록)이 된다.
- AC-WC17-004/005 는 분류 근거와 함께 regression-guard 로 명시한다 — release-blocking 목록에 세지 않는다.

## §D.3 추적성 (REQ ↔ AC)

| REQ | 축 | AC |
|-----|----|----|
| REQ-WC-017-001 | REQ-A | AC-WC17-001, AC-WC17-002 |
| REQ-WC-017-002 | REQ-A | AC-WC17-001, AC-WC17-002 |
| REQ-WC-017-003 | REQ-B | AC-WC17-002, AC-WC17-003 |
| REQ-WC-017-004 | REQ-B | AC-WC17-003 |
| REQ-WC-017-005 | REQ-B | AC-WC17-004 |
| REQ-WC-017-006 | REQ-B | AC-WC17-003 (c), AC-WC17-005 |

## §D.4 Definition of Done

1. AC-WC17-001..003 (release-blocking) 전부 PASS — 명령·출력·HEAD SHA 귀속 동반(E1 형식).
2. AC-WC17-004/005 (regression-guard) PASS — 분류 근거 유지.
3. HARD-1..7 위반 0건 — E4 경계 grep 포함.
4. `recordingSeams` seam 목록 보존(HARD-2) — 형제 SPEC-016의 하니스 의존 회귀 0.
5. `go build ./...` + `GOOS=windows` 빌드 exit 0; `golangci-lint run` NEW 0; `go test ./internal/web/...` PASS.
6. spec lint 가 본 SPEC 의 REQ 6건을 수집함을 확인(수집-가능 형태 `- REQ-WC-017-NNN: ... SHALL ...`) — 「✓ No findings」가 공허이지 않음을 표본으로 닫는다.
