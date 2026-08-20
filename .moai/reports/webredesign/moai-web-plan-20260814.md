# MoAI Web Console 확장 기획안 — 요약

기준: 2026-08-14 / moai-adk-go `main`. 정본은 `moai-web-menu-spec.md`(467줄), 시각 지침은 `moai-web-redesign-brief.md`.

## 메뉴 5영역

| 영역 | 라우트 | 상태 |
|---|---|---|
| 개요 | `/` | 신규 |
| 칸반 | `/kanban` | 신규 (뷰 A 체인 보드 + 뷰 B SPEC 파이프라인) |
| SPEC | `/specs`, `/specs/{id}` | 기존 확장 |
| 모니터 | `/monitor` | 신규 |
| 설정 | `/settings` | 기존 (라우트 이동만) |

## MoAI-Kanban 역할별 모델

| # | 역할 | 모델 | effort | 과금 |
|---|---|---|---|---|
| 0 | Lead | Opus 5 | medium | Claude Max |
| 1 | Plan | Opus 5 | high | Claude Max |
| 2 | Run | GLM-5.2 | xhigh | API 종량 |
| 3 | Review | Opus 5 | medium | Claude Max |
| 4 | Sync | GLM-5.2 | high | API 종량 |

## 확정 결정

- D1 칸반 카드는 SPEC frontmatter에서 유도 (전용 보드 저장소 없음)
- D2 실시간은 SSE (Go 표준 라이브러리, 외부 의존성 0)
- D3 모니터링 화면 전부 읽기 전용 (SPEC 상태 전이 소유권 규율 보존)

## 핵심 사실

- MCP 17도구 = 내부 Go 패키지 래퍼(`session.QueryActiveWork`, `goal.LoadGoal`, `spec.ListDocs`, `verify.Load`). 웹은 같은 패키지를 직접 호출 — `board.go:89`가 이미 그렇게 함. **동기화 신규 작업 0.**
- 신규 작업은 변경 알림 채널 하나뿐. `fsnotify`는 `go.mod:64`에 간접 의존으로 이미 존재.
- 번들 `htmx.min.js`는 2.0.4 코어이고 SSE 확장 없음(`EventSource` 0건) → `app.js`에 `EventSource` 직접 배선 필요.

## 데이터 공백 (검증됨)

| # | 항목 | 사실 |
|---|---|---|
| 1 | 역할 | `kanban.Record`(record.go:55)·`active-sessions.json` 모두 role 필드 없음. 라벨은 `cc.go:115`/`glm.go:184`에서 파싱만 하고 미저장 |
| 2 | 컨텍스트 사용률 | `context-usage.json`은 프로젝트당 단일 슬롯, last-writer-wins. 관측: `368a2bd9…/260000` → `e463a3c9…/0`으로 덮어써짐 |
| 3 | 모델·effort | 런타임 실시간 값, 미저장. `Backend`(claude/glm)만 저장됨 |
| 4 | 단계 상태 | 디스크에 없음. 하트비트 추정으로 시작 권장 |

## 3단계 로드맵

| 단계 | 내용 | 선행 |
|---|---|---|
| 1 | 개요·SPEC·모니터·설정·칸반 뷰 B·SSE 채널 | 없음 |
| 2 | 칸반 뷰 A 기본형 (기동·역할·백엔드·하트비트) | `Record.role` 추가 |
| 3 | 칸반 뷰 A 완성형 (모델·effort·컨텍스트·비용) | context-usage 세션별 분리 + 모델 스냅샷 |

## 디자인 비의존 작업 (즉시 착수 가능)

`internal/web/events.go`(SSE+fsnotify) · `internal/kanban/record.go`(role) · `internal/cli/cc.go`·`glm.go`(역할 기록) · `overview.go`·`kanban.go`·`monitor.go`(뷰모델) · `app.go`(라우트 골격).
디자인 의존: `*.templ` 템플릿만.

## 남은 결정

1. 뷰 B 4컬럼 축소 — 권장 축소
2. 단계 상태 판정 방식 — 권장 하트비트 추정 → 리드 기록
3. 생산자 작업 별도 SPEC 분리 — 권장 분리
4. 비용 금액 표시 — 권장 1단계는 정액/종량 배지만
5. 개요를 `/`로 — 권장 이동
6. SPEC 상세 페이지 vs 패널 — 미정
