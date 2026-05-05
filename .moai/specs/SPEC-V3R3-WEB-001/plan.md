# Plan — SPEC-V3R3-WEB-001: Cockpit Foundation + Workflow Tracker

## Overview

본 plan은 SPEC-V3R3-WEB-001의 구현 단계를 6개 Phase (A→F)로 분해하고, 후속 WEB-002~008 SPEC들이 의존할 수 있는 안정적인 foundation을 확립한다. 구현 기술은 `proposal.md` Technical Stack Note에 검증된 후보 (Go + a-h/templ + HTMX)를 출발점으로 두되, **최종 기술 결정은 `/moai run` 진입 시 manager-strategy 위임으로 확정**한다. plan 단계에서는 capability·요구사항·검증 가능한 acceptance 기준만 고정한다.

본 SPEC은 First Principles Layer 4 root 욕구를 단독으로 충족하는 최소 가치 슬라이스이며, lessons #9 wave-split 적용 시 **Wave 1 단독 SPEC**으로 분류된다. WEB-002와 WEB-003은 Wave 2로 병렬 처리 예정 (둘 다 read-only 데이터 페치 패턴 공유).

## Implementation Phases

### Phase A — CLI Subcommand Registration + Flag Parsing
- `cmd/moai/main.go`에 `cockpit` 서브커맨드 등록
- 플래그 파싱: `--port <int>` (default 후보 8080), `--no-open` (boolean)
- 사용 도움말 (`moai cockpit --help`) 출력 확인
- 단위 테스트: 플래그 파싱 결과 검증 (테이블 드리븐, `t.TempDir()` 격리)
- 산출물: 서브커맨드 등록 + 플래그 구조체 + 테스트
- 의존성: 없음 (기존 main.go 확장만)

### Phase B — HTTP Server Scaffold + Graceful Shutdown
- HTTP 서버 부트스트랩 (loopback 바인딩 강제: `127.0.0.1:<port>`)
- `GET /` 라우트 등록 (Phase E에서 Workflow Tracker 페이지 핸들러로 교체)
- 시그널 핸들링: SIGINT / SIGTERM 수신 시 graceful shutdown (in-flight 요청 5초 드레인 → listener close → exit 0)
- 옵션 `--no-open` 처리: false면 `localhost:<port>` 브라우저 자동 오픈 시도 (실패해도 서버는 계속)
- 단위 테스트: 서버 시작/중단 라이프사이클 (포트 충돌 시나리오 포함)
- 산출물: 서버 부트스트랩 함수 + 시그널 핸들러 + 테스트
- 의존성: Phase A

### Phase C — SPEC Discovery via Filesystem Walk
- `.moai/specs/SPEC-*/` 디렉토리 발견 (Glob 패턴)
- 각 SPEC 디렉토리의 `progress.md` 존재 여부 확인
- 발견 결과를 in-memory 캐시에 저장 (TTL = polling interval / 2 = 2.5초; R2 mitigation)
- 빈 디렉토리 / 권한 오류 / 누락된 progress.md를 graceful하게 처리 (REQ-WEB-004 충족)
- 단위 테스트: 다양한 디렉토리 상태 (빈/단일/다수/오류) 시나리오
- 산출물: SPEC 발견 walker 함수 + 캐시 + 테스트
- 의존성: 없음 (Phase A/B와 병렬 가능)

### Phase D — progress.md Parser
- `progress.md`의 마지막 체크포인트 추출:
  - 마지막 `created_at` / `updated_at` / 체크포인트 timestamp 추출
  - 현재 phase 식별자 추출 (plan/run/sync 중 하나)
- 파싱 실패 시 graceful fallback ("Unknown phase" 등 표시 가능한 중립값)
- mtime 폴백: 파일 mtime을 last-checkpoint timestamp로 사용 (가장 안정적인 hint)
- REQ-WEB-005를 위해 모든 SPEC의 progress.md mtime 비교 → 가장 최근 SPEC 선택
- 단위 테스트: 다양한 progress.md 포맷 (정상/누락 필드/깨진 YAML)
- 산출물: progress 파서 + multi-SPEC 정렬 함수 + 테스트
- 의존성: Phase C (디렉토리 walker 결과 소비)

### Phase E — Workflow Tracker templ Component + HTMX Fragment Endpoint
- 페이지 셸 (HTML doctype + HTMX 스크립트 로드 + 빈 컨테이너)
- Workflow Tracker 패널 컴포넌트:
  - SPEC ID 표시
  - 현재 Phase 표시 (plan/run/sync 배지)
  - 마지막 progress 체크포인트 timestamp
  - 멀티 SPEC 시 "(N other active SPECs)" 비-인터랙티브 표시 (REQ-WEB-005)
- HTMX 폴링 속성 부착: `hx-get="/api/workflow-tracker" hx-trigger="every 5s" hx-swap="outerHTML"`
- Fragment 엔드포인트 `GET /api/workflow-tracker` 구현 (partial HTML 응답, < 300ms 목표)
- 빈 상태 / 에러 상태는 동일 fragment 내에서 처리 (REQ-WEB-004 — 토스트/모달/스피너 없음)
- 단위 테스트: 다양한 상태 (정상 단일/멀티/빈/에러) 렌더링 검증
- 산출물: 페이지 셸 + 패널 컴포넌트 + fragment 핸들러 + 테스트
- 의존성: Phase B (서버), Phase C+D (데이터 소스)

### Phase F — 5s Polling Integration Test (Golden-Path E2E)
- 테스트 환경 부트스트랩: `t.TempDir()` 내에 가짜 `.moai/specs/SPEC-TEST-001/progress.md` 생성
- Cockpit 서버 시작 (랜덤 포트로 충돌 회피)
- HTTP 클라이언트로 `GET /` 요청 → Workflow Tracker 패널 HTML 검증
- 5초 후 `GET /api/workflow-tracker` 요청 → fragment 응답 검증
- progress.md 외부 수정 후 다음 폴링 응답에 변경 반영 확인
- 응답 시간 측정: page load < 200ms (p95), fragment fetch < 300ms (per request)
- read-only invariant 검증: 응답 HTML에 `<form>`, `<button type="submit">`, `hx-post`, `hx-put`, `hx-delete`, `hx-patch` 부재 확인
- 산출물: golden-path E2E 테스트 + 성능 메트릭 어설션
- 의존성: Phase A~E 전체

## Stack Decision Deferred

[HARD] **본 plan은 구현 기술의 최종 결정을 포함하지 않는다.** 구현 기술 선택은 `/moai run` 진입 시 manager-strategy 위임으로 확정되며, 이는 다음을 의미한다:

- `proposal.md` Technical Stack Note의 후보 조합 (Go + a-h/templ + HTMX)은 **출발점 가설**일 뿐, 본 plan은 이를 binding 결정으로 채택하지 않는다.
- manager-strategy는 First Principles에 따라 후보 스택을 재평가하며, 다음을 산출한다:
  - 백엔드 라우터 선택 (chi vs net/http vs 다른 옵션)
  - 템플릿 렌더링 선택 (a-h/templ vs html/template vs 다른 옵션)
  - 프론트 폴링 메커니즘 선택 (HTMX vs vanilla JS fetch vs 다른 옵션)
  - 의존성 추가 영향 분석 (`go.mod` 변경 범위 / 라이선스 호환 / 빌드 사이즈)
- 본 plan의 EARS 요구사항(REQ-WEB-001~005)과 acceptance 시나리오는 기술 선택과 무관하게 유효하다.

## Technology Stack (CANDIDATES — not binding)

다음은 `proposal.md` Technical Stack Note + `research.md` §2.3에서 검증된 후보다. `/moai run` 단계에서 채택 또는 대체될 수 있다.

| 영역 | 후보 | 근거 |
|------|------|------|
| 백엔드 라우터 | Go `chi` 또는 `net/http` | `research.md` §2.1, Go stdlib만으로도 충분 |
| 템플릿 | `a-h/templ` (Apache 2.0) | templ Fragments + HTMX partial swap 패턴 (`research.md` §2.2) |
| 프론트 폴링 | HTMX (BSD) `hx-trigger="every 5s"` | 코드량 1/10 vs WebSocket (`ideation.md` Critical Decisions) |
| UI 컴포넌트 | 선택적 Tailwind 또는 templUI (MIT) | 학습 가속 옵션, MVP 단계 필수 아님 |
| 데이터 소스 | filesystem walk (`.moai/specs/SPEC-*/`) + `progress.md` 읽기 | 외부 DB 없음, 의존성 0 유지 |
| CLI 통합 | `cmd/moai/main.go` 확장 → `moai cockpit [--port N] [--no-open]` | 기존 binary에 통합, 별도 install 불필요 |

## Risk Analysis & Mitigations

`proposal.md` R1-R5 중 본 SPEC Foundation에 직접 연관된 항목:

| Risk | 본 SPEC 영향 | 완화 |
|------|-------------|------|
| R2: 폴링 누적 IO 부담 (filesystem) | 5초 주기 walker가 `.moai/specs/` 반복 읽기 | Phase C에 in-memory cache (TTL = polling interval / 2 = 2.5초) 의무화. 캐시 무효화는 mtime 비교로 트리거. |
| R3: gh CLI 출력 포맷 변경 | N/A (본 SPEC은 gh CLI 미사용) | Forward concern: WEB-003에서 모든 gh 호출은 `--json` 플래그 사용 의무. 본 SPEC에서는 직접 영향 없음. |
| R5: Solo-dev panel-incremental 추정 부정확 | WEB-001은 가장 작은 viable slice — 본 SPEC IS the smallest slice | 본 SPEC만 단독으로도 First Principles Layer 4 가치 발생. Wave 1 완료 후 dogfooding으로 만족도 검증 가능. |

본 SPEC에서 신규 식별된 리스크:

| Risk | 영향 | 완화 |
|------|------|------|
| R6: 포트 충돌 (default 8080 사용 중) | 서버 시작 실패 | Phase B에서 포트 바인딩 실패 시 명확한 에러 메시지 + 종료 코드 출력. `--port` 플래그로 사용자 우회 가능. |
| R7: 시그널 핸들링 누락 시 좀비 프로세스 | Ctrl+C 후 프로세스 미종료 | Phase B에 SIGINT/SIGTERM 핸들러 + 5초 드레인 + listener close + exit 0 의무화 (acceptance Scenario 7로 검증). |
| R8: progress.md 포맷 변동 | 파서 깨짐 → empty state로 폴백 | Phase D에 graceful fallback (mtime 사용) + REQ-WEB-004 적용. 파싱 실패는 에러가 아닌 "Unknown phase"로 표시. |

## MX Tag Plan

[MANDATORY — Phase 3.5] 본 SPEC 구현 시 다음 MX 태그를 반드시 배치한다:

### @MX:ANCHOR — Cockpit Entry Function
- **위치**: Cockpit 서브커맨드 진입 함수 (Phase A 산출물)
- **이유**: 본 함수는 high fan_in이 예상됨 — 단위 테스트 + 통합 테스트 + 후속 WEB-002~008 panel SPEC들이 이 진입점을 확장 또는 참조한다.
- **목적**: invariant contract 명시 — 진입 함수 시그니처와 책임 경계를 명확히 표시하여 향후 panel 추가 시 회귀 방지.

### @MX:NOTE — SPEC Discovery Walker
- **위치**: `.moai/specs/SPEC-*/` 발견 walker 함수 (Phase C 산출물)
- **이유**: filesystem 레이아웃 가정 (예: `.moai/specs/` 위치, `SPEC-*` 명명 규칙, `progress.md` 존재 가정)을 코드에 투명하게 문서화 필요.
- **목적**: context와 intent 전달 — 향후 SPEC 디렉토리 구조 변경 시 영향 범위를 즉시 식별 가능.

### @MX:WARN — HTTP Server Graceful Shutdown
- **위치**: 시그널 핸들러 + listener close 코드 (Phase B 산출물)
- **이유**: 다중 goroutine + signal handling은 race condition·좀비 프로세스 위험 영역.
- **목적**: danger zone 표시 + @MX:REASON 동반 ("5초 드레인 후 강제 종료, in-flight 요청은 best-effort 완료")

### @MX:TODO — Phase B-F Implementation Completion Gates
- **위치**: 각 Phase 산출물의 미완성 hook 포인트 (Run 진입 시점)
- **이유**: incremental 구현 중 미완성 영역을 명시적으로 추적.
- **해소 시점**: GREEN phase 완료 (각 Phase의 acceptance 시나리오 통과) 시 즉시 제거 — 미해소 @MX:TODO는 SYNC phase에서 quality gate 차단.

## Reference Implementations

`research.md` §2.3에서 검증된 시작점 템플릿 (구현 시 참조 — 본 SPEC에는 binding 결정 아님):

- `Piszmog/go-htmx-template` — Go 1.25+ native CSRF, sqlc, Tailwind via go tools, `air` hot reload
- `josephspurrier/gohtmxapp` — `localhost:8080/dashboard` 즉시 실행 가능 패턴, `make watch` 한 줄

## Dependencies

본 plan은 신규 외부 의존성을 binding하지 않는다. `/moai run` 단계에서 결정될 후보:

- Go stdlib (`net/http`, `os/signal`, `path/filepath`) — 의존성 추가 없음
- `a-h/templ` (Apache 2.0) — 후보, 채택 시 `go.mod` 추가
- `chi` (MIT) — 후보, `net/http`로 충분 시 미채택 가능
- HTMX (BSD) — 클라이언트 사이드, `static/htmx.min.js` 정적 파일로 임베드 가능

[HARD] 의존성 추가 결정은 `/moai run` 단계의 manager-strategy 책임이며, 본 plan은 의존성 0 유지를 우선순위 1로 권고한다 (Cost Structure: localhost only, 외부 SaaS 없음 — `ideation.md` §7).
