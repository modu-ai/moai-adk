---
id: SPEC-FACTORY-MANAGED-SESSION-001
title: "design.md — Factory 관리 세션 계층 설계 결정"
version: "0.1.0"
created: 2026-10-01
updated: 2026-10-01
author: GOOS (manager-spec)
tier: L
---

# design.md — 설계 결정 D-1..D-7

> 상태축 없음(spec-frontmatter-schema.md § Artifact Statelessness). 생명주기는 `spec.md`만 운반한다.

## D-1 — F3 경계 (첫 번째 설계 포크, 카드 사전조건)

**결정**: 관리 세션 계층은 **배달 계층**만 소유한다 — 메시지가 모델 턴에 들어가고 receipt가 기록되는 것까지. 카드 처분·병합·재알림은 전부 SPEC-FACTORY-CONTROLLER-001(t1241 F3)의 표면이다.

F3의 현 상태(정찰 R7): M1a만 착지 — SPEC-FACTORY-RECORD-001 amendment 0.3.0(커밋 c6fab1eab)이 T29b/T29c lane↔controller 핸드오버 에지를 추가. 컨트롤러 코드는 develop에 없고 SPEC-FACTORY-CONTROLLER-001은 브랜치 거주(WT-factory-controller)다. t1365 Residual-risk가 지적한 "관리 세션 계층이 F3 진화 방향과 겹칠 수 있다"는 위험을 이 테이블이 소멸시킨다.

**인터페이스 경계 테이블**:

| 표면 | 관리 세션 계층 (본 SPEC) | F3 (SPEC-FACTORY-CONTROLLER-001) |
|---|---|---|
| 메시지 배달 | 소유 — claim → 메타데이터 주입 → body 조회 → receipt | 소비 — 배달된 status_report를 판정 입력으로 사용 |
| receipt | 배달 증거로만 기록 (REQ-MS-015) | 완료 판정의 근거 중 하나로 소비 |
| 병합 창 | 창을 잡지 않고, 병합하지 않는다 | `moai integration acquire/release` 창 소유자 |
| 카드 상태 | 읽지도 쓰지 않는다 | T29b/T29c 핸드오버 이벤트 발행 |
| 임대(lease) | 브로커 Claim lease(가변, store 소유)만 | lane lease 만료 reaping |
| Decider | 없음 | `human \| llm \| jev` |
| 재알림 | 없음 — claim은 일회성, 재큐는 store 정책 | stalled-card re-alert |

**기각된 대안**:
- (a) PR 본문 그대로 병합 — 30파일 충돌 + `agent-1..N` 어휘 회귀(t1365 Evidence 3·4). 기각.
- (b) 관리 계층이 F3 기계까지 흡수 — 범위 폭발, t1365 Residual-risk 경고 그대로 재현. 기각.
- (c) 관리 계층을 F3 착지 후로 연기 — F3도 배달된 보고가 입력이므로 계층 분리가 순서 의존을 만들지 않는다. 기각.

**비침범 검증 가능성**: AC-MS-010 — 신규 관리 파일에서 병합-자동화 상징(`merge-window`, `Decider`, `T29b`, `T29c`, `handover` 이벤트 발행) grep 0적중. SPEC-FACTORY-LANE-AUTONOMY-001 §F의 umbrella exclusion과 같은 규율.

## D-2 — 전송: gorilla/websocket 채택 (App Server 클라이언트 한정)

**결정**: `github.com/gorilla/websocket v1.5.3`을 신규 의존으로 채택한다. 사용은 Codex App Server의 루프백 WS 클라이언트(다이얼 + ReadJSON/WriteJSON) 한정이다.

**의존성 회피 사다리 적용**(`moai-constitution.md` § Agent Core Behaviors 4):
1. stdlib — Go 표준라이브러리에 WebSocket 클라이언트가 없다(`net/http`는 업그레이드 서버용 힌트만). 사다리 3단 실패.
2. 기존 의존 — `go.mod`에 WS 클라이언트 없음. 사다리 5단 실패.
3. `golang.org/x/net/websocket` — 하위호환 모드로 유지보수 우선순위가 낮고, 헤더 주입(Authorization Bearer)과 JSON 편의에서 열위. 기각.
4. gorilla/websocket — 2023 유지보수 재개 이후 사실상 표준, PR에서 이미 macOS 실측됨. **채택.**

**cross-host 범위 구체화**: WS 엔드포인트는 **같은 호스트 루프백 전용**(`127.0.0.1` 에페멀 포트)이다 — 런처가 그 호스트의 codex 프로세스를 소유하기 때문. "cross-host"는 leader↔lane이 서로 다른 런처·백엔드 조합인 경우(GLM lead ↔ Codex lane 등)를 뜻하며, 머신 경계는 아니다.

**인증/비밀 처리**: 런처당 32바이트 랜덤 capability token(`crypto/rand` → hex), `0600` 임시 디렉터리의 토큰 파일로 codex에 전달, WS 접속은 `Authorization: Bearer <token>`. 토큰은 프로세스 수명과 같이 소멸하고 디스크에 상주하지 않는다(임시 디렉터리는 defer 제거).

**기각된 대안**: (a) x/net/websocket — 위 3; (b) 자체 RFC6455 구현 — 프레이밍 자체구현은 과잉 공학; (c) WS 없이 파일 폴링만 — Codex 무인 턴 주입(App Server `turn/start`)이 불가능해 PR의 핵심 가치 소멸.

## D-3 — 어휘 + 브로커 적합성

**결정**: `lane-<n>` · `-l/--leader` · 현 `internal/factorymsg` API(RegisterPeer / RegisterLaunchPending / BindLaunchPending / RollbackLaunchPending / PeerByOwner / Claim+lease / ReadBody(claim token) / Receipt / SettleReceiptControls) 위에서만 재작성한다. `store.go`는 수정하지 않는다(REQ-MS-008).

**근거**: develop `factory.go`는 `lane-<n>`을 표준으로 하고 legacy `worker`/`agent` 토큰을 오류 처리한다(factory.go:103, :241-244, :312-319). 브로커는 PR 베이스보다 진행했다(t1365 Evidence 6 — RegisterPeer·LaunchPending bind/rollback·Claim/lease·ReadBody·receipt). PR의 `agent-1..N` 표준화는 그 정반대 방향이라 전부 기각.

**주의(정찰 R13)**: PR은 `store.go`를 13행 수정했다 — 재작성은 무수정을 원칙으로 하되, run-phase에서 현 API로 불가능한 지점이 확인되면 blocker로 보고한다(침묵 패치 금지).

## D-4 — 크로스플랫폼: 자식 프로세스 소유 모델

**결정**: 관리 세션은 **자식 프로세스 소유** 모델(`exec.Command` + Wait)로 통일한다 — `syscall.Exec` 교체(execve)를 쓰지 않는다. 따라서 `launch_exec_posix.go` / `launch_exec_windows.go` 분할이 추가로 필요 없고, Windows에서도 동일 코드 경로가 동작한다.

**근거**: POSIX exec 경로(`launch_exec_posix.go`)는 프로세스를 교체하므로 소유 폴링 루프를 유지할 수 없다 — 관리 세션의 본질(소유자가 폴링·주입)과 상충. `codex_launcher_guards_test.go:173`의 zero-syscall 가드를 신규 파일도 준수한다.

**Windows 자세**: `GOOS=windows GOARCH=amd64 go build` 통과(AC-MS-014) + 신규 파일 zero-syscall. Windows 네이티브 실세션 인증은 §F로 제외(정직한 Gap 유지 — PR도 macOS 실측만).

**기각**: (a) Windows에서 exec 흉내 — 불가능(EWINDOWS, `launch_exec_windows.go` 주석이 문서화); (b) Windows 지원을 아예 끊고 build tag로 제외 — 기존 크로스빌드 게이트 회귀.

## D-5 — MCP 승인 스코핑: PR e33cf0e0d 채택

**결정**: 관리 Codex 런처가 **자신이 소유한 App Server 프로세스에만(TUI는 미배달 — 후속 카드)** `-c mcp_servers.moai.default_tools_approval_mode="approve"` + `factory_msg_send`/`factory_msg_receipt` 도구별 approve를 인수로 전달한다. 프로젝트 설정의 capability 기반 `mcpApprovalMode = "writes"`(`internal/codexwiring/configtoml.go:22-26, :99`)는 무변경. doctor는 이전 세대의 프로젝트 전역 승인 잔재를 경고하고 사용자 파일을 건드리지 않는다(REQ-MS-009·010).

**근거**: PR 리포트 실측 — 승인 스코핑 전에는 MoAI MCP 승인창에서 무인 수신이 정지했다(`approval_mode="auto"` 시도는 통과 증거에서 제외됐다고 리포트 자체가 기록). 스코핑 후 양방향 왕복에서 승인창 0관측. develop의 전역 완화는 보안 경계 위반이라 기각.

## D-6 — 검증 전략: loopback 3층

**결정**: (1) **단위** — 인수 파서·직렬 큐·프롬프트 조립. (2) **loopback 통합** — 가짜 세션 바이너리를 `go test` 자기 재실행(re-exec) 패턴으로 기동(`codex_launcher_exec_posix_test.go` 선례), 임시 브로커 저장소 상대로 claim→주입→body→receipt 왕복을 실제 제2호스트 없이 검증. (3) **live 게이트** — `MOAI_FACTORY_LIVE_ROOT`/`MOAI_FACTORY_LIVE_RUN` env가 있을 때만 실세션 왕복 테스트 실행(PR의 live_test 계승), CI에서는 skip.

**품질**: acceptance.md의 AC는 전부 기계 판정형. RED-전 관측은 run-phase TDD 사이클이 낸다(manager-develop §E8).

## D-7 — 런처 배선 위치

**결정**: 배선은 세 런처의 factory 진입 분류 이후에 둔다 — `codex_launcher.go`의 `codexFactoryEntryClassify`(:791)와 `runCodex`(:658) 계열, cc/glm의 factory 세션 기동 경로. 관리 경로는 factory 리더·레인 세션에서만 활성화되고 일반(비-factory) 실행은 기존 exec/spawn 경로를 그대로 유지한다. Claude-only run에서 `factory_msg_send` 거부 정책(PR의 `TestFactoryMsgSendRejectsClaudeOnlyRun`)을 계승해 네이티브 SendMessage 정책과 충돌하지 않게 한다(REQ-MS-011).

**기각**: (a) 새 최상위 서브커맨드로 관리 세션 노출 — 표면 이중화; (b) hook에서 세션을 소유 — 훅은 수명이 짧아 소유 루프를 못 유지.
