# SPEC-V3R3-WEB-001 (Compact) — Cockpit Foundation + Workflow Tracker

> ⚠️ **ARCHIVED 2026-05-04** — Superseded by SPEC-V3R3-CONSOLE-NNN (Console paradigm, CRUD-capable, Bun+Hono+React+TS Claude Agent SDK). Do NOT load this file in `/moai run`. See spec.md HISTORY 0.2.0 entry.

> Compact extract of `spec.md` for `/moai run` token efficiency. Excludes Overview, Why, HISTORY, Source References. ~30% token savings vs full spec.md.

## Goals

| Goal ID | Type | Target |
|---------|------|--------|
| Goal Primary | Productivity | User CLI roundtrip count reduces by 30%+ (baseline vs post-Cockpit, shell history measurement) |
| Goal Secondary 1 | UX (latency) | Page load p95 < 200ms |
| Goal Secondary 2 | Performance (polling cost) | Polling fetch time per request < 300ms |
| Goal Secondary 3 | Stability (forward concern) | gh API rate-limit hit count = 0 (forward concern for WEB-003) |
| Goal Anti | Read-only invariant | Write-action click count = 0 |

## EARS Requirements

### REQ-WEB-001 (Ubiquitous — Server Lifecycle / Single-Page Route)

The Cockpit server **shall** expose a single HTTP route `GET /` on `localhost:<port>` that renders the Cockpit page containing the Workflow Tracker panel, where `<port>` defaults to a sensible value (candidate: 8080) and is configurable via `--port` flag. The server **shall** bind exclusively to the loopback interface (localhost-only invariant).

### REQ-WEB-002 (Event-Driven — Page Load Triggers Workflow Tracker Render)

**When** a browser issues `GET /` to the Cockpit server, the server **shall** respond with a fully rendered HTML page within 200ms (p95) containing the Workflow Tracker panel populated with: (a) the active SPEC ID, (b) the current Phase identifier (one of: `plan`, `run`, `sync`), and (c) the timestamp of the most recent `progress.md` checkpoint.

### REQ-WEB-003 (State-Driven — Polling While Running)

**While** the Cockpit server process is running and a browser tab holds the Cockpit page open, the page **shall** issue a `GET` request to a fragment endpoint (candidate: `/api/workflow-tracker`) every 5 seconds via a client-side polling mechanism, and the server **shall** respond with the partial HTML fragment representing the current Workflow Tracker state within 300ms per request, and the page **shall** replace the Workflow Tracker panel content in-place without triggering a full-page navigation.

### REQ-WEB-004 (Unwanted Behavior — Graceful Empty / Error State)

**If** the SPEC discovery walker encounters an empty `.moai/specs/` directory, no readable `progress.md` files, or a filesystem read error, **then** the Workflow Tracker panel **shall** render a neutral "No active SPEC" empty state with no error toast, modal, or spinner, and the page **shall not** display any state-mutating UI control (no buttons, no forms, no `POST`/`PUT`/`DELETE`/`PATCH` triggers).

### REQ-WEB-005 (Complex — Multi-Active-SPEC Resolution)

**While** more than one SPEC directory under `.moai/specs/SPEC-*/` contains a non-empty `progress.md` file, **when** the Workflow Tracker renders, the server **shall** select the SPEC whose `progress.md` has the most recent modification time (mtime) as the primary surface, and **shall** display a non-interactive indicator showing the count of other active SPECs (e.g., "(N other active SPECs)") with no navigation affordance.

## Acceptance Scenarios (Given/When/Then)

### Scenario 1 — CLI Entry Point + Server Start (REQ-WEB-001)

**Given:** moai 바이너리 설치 + moai-adk-go 프로젝트 루트 + `.moai/specs/SPEC-*/` 1개 이상 존재
**When:** `moai cockpit --port 8080 --no-open` 실행
**Then:** `127.0.0.1:8080` loopback 바인딩, "Cockpit listening on :8080" 로그, foreground 유지, `--no-open`로 브라우저 미오픈, `curl http://localhost:8080/` 200 응답

### Scenario 2 — Workflow Tracker Initial Render (REQ-WEB-002, Goal Secondary 1)

**Given:** Cockpit 서버 `localhost:8080` 실행 + 최소 1 SPEC 디렉토리 + 유효 progress.md
**When:** `GET http://localhost:8080/` 요청
**Then:** ToFB p95 < 200ms, 완전 렌더링 HTML, 패널에 SPEC ID + Phase (plan/run/sync) + 마지막 체크포인트 timestamp 모두 표시

### Scenario 3 — 5-Second Polling Refresh (REQ-WEB-003, Goal Secondary 2)

**Given:** 패널 렌더링 + 브라우저 탭 열림 + 클라이언트 사이드 폴링 활성
**When:** 5초 경과 + 외부에서 progress.md 수정
**Then:** 페이지가 자동으로 `GET /api/workflow-tracker` (또는 동등한 fragment 엔드포인트) 발사, 응답 < 300ms, partial HTML fragment, 패널 콘텐츠가 in-place로 교체, 전체 페이지 새로고침 부재, 새 체크포인트 반영

### Scenario 4 — Graceful Empty State (REQ-WEB-004)

**Given:** `.moai/specs/` 비어있거나 progress.md 부재
**When:** `GET http://localhost:8080/` 요청
**Then:** HTTP 200, "No active SPEC" 중립 메시지, 토스트/모달/스피너 부재, `<form>`·`<button type="submit">`·`hx-post`·`hx-put`·`hx-delete`·`hx-patch` 모두 부재

### Scenario 5 — Multi-Active-SPEC Resolution (REQ-WEB-005)

**Given:** 3개 SPEC 디렉토리 + 각각 비어있지 않은 progress.md + 서로 다른 mtime (B-001 가장 최근)
**When:** `GET http://localhost:8080/` 요청
**Then:** 패널이 SPEC-B-001을 primary 표시, "(2 other active SPECs)" 비-인터랙티브 indicator 표시, 다른 SPEC으로의 anchor/HTMX 트리거 부재

### Scenario 6 — Read-Only Invariant Verification (Goal Anti)

**Given:** Cockpit 페이지 완전 로드 + 모든 인터랙티브 요소 식별 + Network 탭 열림
**When:** 사용자가 모든 visible 인터랙티브 요소 클릭
**Then:** Network 로그 모두 GET only, POST/PUT/DELETE/PATCH 0건, HTML에 `<form method="post">`·`hx-post|put|delete|patch`·write `onclick` fetch 모두 0건

### Scenario 7 — Graceful Shutdown

**Given:** 서버 `localhost:8080` 실행 + 활성 브라우저 연결 (HTMX 폴링 중)
**When:** SIGINT 또는 SIGTERM 발송
**Then:** 새 연결 수락 즉시 중단, in-flight 요청 최대 5초 드레인, 완료 시 listener close, 5초 초과 시 강제 종료, exit code 0, 좀비 프로세스/미해제 포트/FD 누수 부재, "Cockpit shutting down..." 메시지 출력

## Files to Modify

- `cmd/moai/main.go` — Register the new `cockpit` subcommand with flags `--port <int>` (default candidate 8080) and `--no-open` (boolean, suppress browser auto-open).

## What NOT to Build

- **C2 Worktree Switchboard panel** → SPEC-V3R3-WEB-002
- **C3 CI/PR Glance Wall panel** → SPEC-V3R3-WEB-003
- **C4 Memory Surfboard panel** → SPEC-V3R3-WEB-004
- **C5 Drift Sentinel panel** → SPEC-V3R3-WEB-005
- **C6 Quick Action Launcher** (optional) → SPEC-V3R3-WEB-006
- **C7 User Configuration system** (optional) → SPEC-V3R3-WEB-007
- **C8 Brand Tokenization integration** (optional) → SPEC-V3R3-WEB-008
- **All write actions** — read-only invariant; no POST/PUT/DELETE/PATCH triggered by user interaction
- **Authentication / CSRF / multi-user** — localhost-only invariant; loopback binding is sole access boundary
- **External database** — filesystem walk + git/gh CLI shell-out only; no SQLite/Postgres/external store
- **Real-time streaming via SSE/WebSocket** — polling sufficient per research.md §2.2
- **Multiple-page navigation** — single-page invariant; no `/specs`, `/dashboard`, `/settings` siblings
- **Brand visual integration beyond reading** — deferred to WEB-008
