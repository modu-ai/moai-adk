# Acceptance — SPEC-V3R3-WEB-001: Cockpit Foundation + Workflow Tracker

7개 Given/When/Then 시나리오로 EARS 요구사항 (REQ-WEB-001 ~ REQ-WEB-005) + read-only invariant + graceful shutdown을 검증한다. 모든 시나리오는 자동화 가능하며 (단위/통합 테스트), `/moai run` Phase F의 golden-path E2E에서 통합 실행된다.

---

## Scenario 1 — CLI Entry Point + Server Start
**Covers:** REQ-WEB-001 (Ubiquitous — Server Lifecycle / Single-Page Route)

**Given:**
- 사용자가 moai 바이너리를 설치한 상태
- 사용자가 moai-adk-go 프로젝트 루트 디렉토리에 위치
- `.moai/specs/SPEC-*/` 디렉토리가 최소 1개 이상 존재

**When:**
- 사용자가 터미널에서 `moai cockpit --port 8080 --no-open` 명령을 실행

**Then:**
- HTTP 서버가 `127.0.0.1:8080` (loopback 전용)에 바인딩되어 시작됨
- stdout 또는 stderr에 "Cockpit listening on :8080" (또는 동등한 메시지)이 출력됨
- 프로세스가 foreground에 유지되며 SIGINT 수신을 기다림
- `--no-open` 플래그로 인해 브라우저 자동 오픈이 발생하지 않음
- 별도 터미널에서 `curl -s http://localhost:8080/`이 HTTP 200 응답을 반환

---

## Scenario 2 — Workflow Tracker Initial Render
**Covers:** REQ-WEB-002 (Event-Driven — Page Load Triggers Workflow Tracker Render), Goal Secondary 1 (Page load p95 < 200ms)

**Given:**
- Cockpit 서버가 `localhost:8080`에서 실행 중
- 최소 1개의 SPEC 디렉토리가 존재하며 해당 디렉토리에 유효한 `progress.md`가 존재
- progress.md는 active phase 정보 (plan/run/sync) + 최근 timestamp를 포함

**When:**
- 브라우저(또는 HTTP 클라이언트)가 `GET http://localhost:8080/` 요청 발송

**Then:**
- 서버 응답 시간 ToFB (Time-to-First-Byte)가 p95 기준 200ms 미만
- 응답 본문이 완전 렌더링된 HTML 페이지 (doctype + Workflow Tracker 패널 포함)
- 패널 내에 다음 3개 데이터 필드가 모두 표시됨:
  - active SPEC ID (예: `SPEC-V3R3-WEB-001`)
  - 현재 Phase 식별자 (`plan` | `run` | `sync` 중 하나)
  - 마지막 progress.md 체크포인트 timestamp (ISO 8601 또는 사람이 읽을 수 있는 포맷)

---

## Scenario 3 — 5-Second Polling Refresh
**Covers:** REQ-WEB-003 (State-Driven — Polling While Running), Goal Secondary 2 (Polling fetch < 300ms)

**Given:**
- Workflow Tracker 패널이 브라우저에 렌더링되어 있음
- 브라우저 탭이 열려 있으며 클라이언트 사이드 폴링 메커니즘이 활성화됨
- 서버가 정상 동작 중

**When:**
- 마지막 fragment 페치로부터 5초가 경과
- 그 사이에 외부 프로세스가 `progress.md`를 수정하여 새 체크포인트를 추가

**Then:**
- 페이지가 자동으로 `GET /api/workflow-tracker` (또는 동등한 fragment 엔드포인트) 요청 발송
- 서버 응답 시간이 300ms 미만 (per request)
- 응답이 partial HTML fragment (전체 페이지가 아닌 패널 마크업만)
- Workflow Tracker 패널 콘텐츠가 in-place로 교체됨 (전체 페이지 새로고침 없음)
- 브라우저는 전체 페이지 새로고침을 수행하지 않음 (네트워크 로그에 `GET /` 부재)
- 갱신된 패널은 `progress.md`의 새 체크포인트를 반영

---

## Scenario 4 — Graceful Empty State
**Covers:** REQ-WEB-004 (Unwanted Behavior — Graceful Empty / Error State)

**Given:**
- `.moai/specs/` 디렉토리가 비어 있거나 존재하지 않음
  (또는 모든 SPEC 디렉토리에 `progress.md`가 부재)
- Cockpit 서버가 정상 시작됨

**When:**
- 브라우저가 `GET http://localhost:8080/` 요청 발송

**Then:**
- 서버 응답 코드가 HTTP 200 (에러가 아님)
- Workflow Tracker 패널이 "No active SPEC" (또는 동등한 중립적 빈 상태 메시지) 표시
- 응답 HTML에 다음 요소들이 모두 부재:
  - 에러 토스트 (`role="alert"` 등)
  - 모달 다이얼로그 (`role="dialog"`)
  - 스피너 / 로딩 인디케이터
  - 모든 state-mutating UI 컨트롤 (`<form>`, `<button type="submit">`, `<input type="submit">`)
  - `hx-post`, `hx-put`, `hx-delete`, `hx-patch` 속성
- 빈 상태 메시지는 중립적 톤으로 표시 (alarm/warning 컬러 사용 안 함)

---

## Scenario 5 — Multi-Active-SPEC Resolution
**Covers:** REQ-WEB-005 (Complex — Multi-Active-SPEC Resolution)

**Given:**
- `.moai/specs/` 아래에 3개의 SPEC 디렉토리가 존재 (예: SPEC-A-001, SPEC-B-001, SPEC-C-001)
- 각 SPEC 디렉토리가 비어있지 않은 `progress.md`를 보유
- 3개 progress.md의 mtime이 서로 다름:
  - SPEC-A-001/progress.md mtime = T-30분
  - SPEC-B-001/progress.md mtime = T-5분 (가장 최근)
  - SPEC-C-001/progress.md mtime = T-1시간

**When:**
- 브라우저가 `GET http://localhost:8080/` 요청 발송

**Then:**
- Workflow Tracker 패널이 SPEC-B-001 (가장 최근 mtime)을 primary로 표시
- 패널 내에 비-인터랙티브 indicator "(2 other active SPECs)" (또는 동등한 표현)이 표시됨
- 해당 indicator는 어떠한 링크/버튼/네비게이션 affordance도 가지지 않음 (read-only)
- 응답 HTML에 SPEC-A-001 / SPEC-C-001로의 anchor (`<a href>`) 또는 HTMX 트리거가 부재

---

## Scenario 6 — Read-Only Invariant Verification
**Covers:** Goal Anti (Write-action click count = 0)

**Given:**
- Cockpit 페이지가 브라우저에 완전 로드됨
- 사용자가 페이지의 모든 시각적 인터랙티브 요소를 식별 가능 (chip, link, badge 등)

**When:**
- 사용자가 페이지 내의 모든 visible 인터랙티브 요소를 순차적으로 클릭/탭
- 브라우저 개발자 도구의 Network 탭이 열려 있어 모든 HTTP 요청을 캡처

**Then:**
- Network 로그에 캡처된 모든 요청이 HTTP `GET` 메서드만 포함
- HTTP `POST`, `PUT`, `DELETE`, `PATCH` 요청이 0건
- 페이지 HTML 정적 검사 결과:
  - `<form method="post">`, `<form method="put">` 등 0건
  - `hx-post`, `hx-put`, `hx-delete`, `hx-patch` 속성 0건
  - `onclick="..."` 핸들러로 fetch/XHR write 요청을 발생시키는 코드 0건

---

## Scenario 7 — Graceful Shutdown
**Covers:** Risk R7 mitigation, server lifecycle integrity

**Given:**
- Cockpit 서버가 `localhost:8080`에서 실행 중
- 1개 이상의 활성 브라우저 연결 (HTMX 폴링 진행 중)

**When:**
- 사용자가 서버 프로세스에 SIGINT (Ctrl+C) 또는 SIGTERM 발송

**Then:**
- 서버가 즉시 새 연결 수락을 중단
- 현재 in-flight HTTP 요청을 최대 5초간 드레인 (graceful)
- 5초 내에 모든 in-flight 요청이 완료되면 즉시 listener close
- 5초 초과 시 강제 종료 (남은 연결은 RST)
- 프로세스 exit code가 0
- 좀비 프로세스 / 미해제 포트 / FD 누수 없음
- stderr 또는 stdout에 "Cockpit shutting down..." (또는 동등한) 메시지 출력
