# progress.md — SPEC-CODEX-SESSION-MSG-001

카드 t187 (운영자 지시 2026-08-23). Codex-Claude 세션 간 양방향 메시징 — moai MCP 브로커 + A2A 정합 엔벨로프.

## §E.1 Plan-phase Audit-Ready Signal

```yaml
plan_status: audit-ready
plan_complete_at: 2026-08-23
plan_commit: 2715f00a5                       # 최초 plan-phase 산출 커밋 (audit-fix 4329f45e6가 v0.2.0 적분)
tier: L
artifacts: 5                               # spec.md + plan.md + acceptance.md + design.md + research.md
requirements: 15                           # REQ-CSM-001..015 (상한 25)
acceptance_criteria: 15                    # AC-CSM-001..015 (상한 25; review-1 D1/D2로 014/015 추가)
spec_lint: exit 0 (2026-08-23, 본 워크트리 WT-codex-session-msg)
design_decision: "axis-(ii) A2A-aligned semantics over MCP broker + file store (research.md §4)"
```

- plan-phase 자가검증: `moai spec lint` exit 0 / SPEC ID 정규식 PASS / frontmatter 12필드 + era: V3R6 + tier: L / 3설계축 전부 research.md §4에 근거와 함께 기록.
- plan-audit review-1 (FAIL 0.840, Traceability 0.70) D1-D7 7건 전부 반영 — iter-2 v0.2.0. 상세는 spec.md HISTORY.
- 3축 비교·A2A 실측 페치 로그는 research.md §3-§4, 채택 구조는 design.md.

## §E.2 Run-phase Evidence

### M1 — 데이터 모델 + 파일 스토어 (2026-08-23)

TDD RED-GREEN-REFACTOR. 모든 명령은 본 워크트리 `WT-codex-session-msg` 에서 실행 (baseline HEAD `854e2c21b` + M1 파일 — 커밋 전 작업 트리; M1 커밋 SHA는 §E.6 리드 보고에 기록).

**§E.8 TDD RED 증거 (GREEN 이전 축어 출력 — 핵심 동작별 첫 테스트)**

| 동작 | 테스트 | RED 축어 출력 (발췌) |
|---|---|---|
| envelope 검증 | `TestEnvelopeA2AAlignment` | `--- FAIL: TestEnvelopeA2AAlignment ... validation_rejects_invalid_messages/empty_messageId: expected validation error containing "messageId", got nil` (+19 subtest FAIL — 정렬 subtest는 타입만으로 PASS, 검증 subtest 전부 FAIL) |
| 등록 멱등 | `TestRegisterIdempotent` | `--- FAIL: TestRegisterIdempotent ... agent_test.go:26: first register returned empty agentId` |
| 하트비트 온/오프라인 | `TestHeartbeatOnlineOffline` | `--- FAIL: TestHeartbeatOnlineOffline ... agent_test.go:129: agent  not listed (got 0 agents)` |
| send/poll/ack 클레임 의미론 | `TestSendPollAck` | `--- FAIL: TestSendPollAck ... store_test.go:39: send returned empty messageId` |
| 클레임 TTL 환원 | `TestClaimExpiryReturn` | `--- FAIL: TestClaimExpiryReturn ... store_test.go:204: first poll delivered 0 messages, want 1` |
| 메시지 TTL 삭제 | `TestMessageExpirySweep` | `--- FAIL: TestMessageExpirySweep ... store_test.go:260: poll A expiredCount = 0, want 1 / claim before expiry failed: []` |
| 동시성 보존 | `TestConcurrentSendPoll` | `--- FAIL: TestConcurrentSendPoll ... received 0 messages, want 100 (loss or duplication)` |

**AC 이항 검증 매트릭스 (M1 소관 7항)**

| AC | Status | Verification Command | Actual Output |
|----|--------|---------------------|---------------|
| AC-CSM-001 | PASS | `go test ./internal/sessionmsg/ -run TestEnvelopeA2AAlignment -v` | `PASS — ok github.com/modu-ai/moai-adk/internal/sessionmsg 0.182s` (전 subtest PASS: camelCase 코어 키 6개 + Part kind 판별자 + raw/url 기각 + 크기/부분수 상한) |
| AC-CSM-002 | PASS | `go test ./internal/sessionmsg/ -run TestRegisterIdempotent -v` | `PASS — ok ... 0.192s` (kind+name 재등록 → 동일 agentId + 하트비트 갱신 + `agents/<agentId>.json` 존재 단언) |
| AC-CSM-003 | PASS | `go test ./internal/sessionmsg/ -run TestSendPollAck -v` | `PASS — ok ... 0.196s` (send→pending 파일, poll→claimed 이동·반환, ack_ids→claimed 삭제, 미등록 수신자→`UnknownAgentError`+known 목록) |
| AC-CSM-004 | PASS | `go test ./internal/sessionmsg/ -run TestClaimExpiryReturn -v` | `PASS — ok ... 0.202s` (ClaimedAt=현재-11분 → 다음 poll 동일 messageId 재클레임, ClaimedAt 갱신 단언) |
| AC-CSM-005 | PASS | `go test ./internal/sessionmsg/ -run TestMessageExpirySweep -v` | `PASS — ok ... 0.215s` (pending·claimed 양측 TTL 초과 삭제 + `ExpiredCount` 보고) |
| AC-CSM-006 | PASS | `go test -race ./internal/sessionmsg/ -run TestConcurrentSendPoll -v` | `--- PASS: TestConcurrentSendPoll (0.12s) — ok ... 1.542s` (10 sender×10 msg + 2 poller → 100 유일 배달, 경쟁 보고 0건) |
| AC-CSM-014 | PASS | `go test ./internal/sessionmsg/ -run TestHeartbeatOnlineOffline -v` | `PASS — ok ... 0.197s` (가짜 시계 31분 경과→`online:false`; register·poll·send(sender) 각각 `online:true` 복귀) |

**품질 게이트 (M1)**

| 항목 | 명령 | 축어 출력 |
|---|---|---|
| E2 빌드 | `go build ./... && GOOS=windows GOARCH=amd64 go build ./...` | `E2_FINAL_OK` (exit 0 양측) |
| E3 커버리지 | `go test -coverprofile=/tmp/smsg_cover2.out ./internal/sessionmsg/` | `ok github.com/modu-ai/moai-adk/internal/sessionmsg 0.484s coverage: 86.9% of statements` (≥85% 목표 충족; 잔여 미커버는 fault-injection I/O 오류 분기) |
| E4 경계 grep | `grep -rn 'AskUserQuestion\|mcp__askuser' internal/sessionmsg/ \| grep -v _test` | 0행 (exit 1 — 매치 없음) |
| E5 린트 | `golangci-lint run --timeout=2m ./internal/sessionmsg/... ./internal/config/...` | `0 issues.` |
| AC-CSM-010 (M1 소관 부분) | `grep -c "DefaultSessionMsg" internal/config/defaults.go` | `12` (≥5; 선언 6 + 주석 인용 6 — `var` 형태 전부) |
| AC-CSM-008 (M1 소관 부분) | `grep -rn "exec.Command\|codex-jobs\|app-server\|net.Listen\|http.Listen" internal/sessionmsg/ \| grep -v _test` | 0행 (위임 토큰 부재; `internal/cli/mcp_session_msg.go`는 M2에서 동일 grep 적용) |
| config 패키지 회귀 | `go test ./internal/config/` | `ok github.com/modu-ai/moai-adk/internal/config 1.455s` |

**M1 산출물**: `internal/sessionmsg/{envelope,agent,store,lock,lock_unix,lock_windows}.go` + 테스트 4파일 (envelope/agent/store/edge) + `internal/config/defaults.go` 임계값 6 `var` (지정 5종 + REQ-CSM-005 부분 수 상한 `DefaultSessionMsgMaxParts` 8). 잠금 패턴은 `internal/session`에서 복제(원본 무변경 — §D PRESERVE 3), 쓰기는 전부 `internal/atomicfile` 경유. 테스트 전부 `t.TempDir()` 격리 (§D PRESERVE 7 — `.moai/state/**` 무훼손).

Gaps: `lock_windows.go`는 darwin 개발 머신에서 실행 검증 불가 — `GOOS=windows GOARCH=amd64 go build` 컴파일 게이트만 통과 (교차 컴파일이 테스트를 컴파일하지 않는 한계는 `GOOS=windows go vet ./...` 로 CI에서 보완 필요 — 메모리 feedback_cross_platform_test_compile 선례). 인/아웃 프로세스 혼합 경쟁(실제 다중 프로세스 flock 경합)은 M1 테스트가 단일 프로세스라 교차 프로세스 경로는 NB-flock 로직 검증만 (TestWithAgentLockTimeout·TestAgentLockAcquireRelease).

### M2 — MCP 도구 표면 4종 (2026-08-23, 리드 승인 후 진행)

TDD RED-GREEN-REFACTOR. baseline HEAD `7cd610c0f` (M1 커밋) + M2 작업 트리. 핸들러 테스트는 `t.Setenv("CLAUDE_PROJECT_DIR", t.TempDir())` 로 스토어 루트를 격리 — 생산 해상 경로(resolveProjectDir)와 동일한 경로로 주입, 실제 `.moai/state/**` 무훼손.

**§E.8 TDD RED 증거 (GREEN 이전 축어)**

| 동작 | 테스트 | RED 축어 출력 |
|---|---|---|
| C-HRA-008 정적 가드 | `TestSessionMsg_NoAskUserQuestion` / `TestSessionMsg_NoInlineGetenv` | `read mcp_session_msg.go: open mcp_session_msg.go: no such file or directory` — 파일 부재 시점 FAIL (구현 파일 생성 전 캡처) |
| 카탈로그 동등성 | `TestMoaiMCPServer_RegistrationMatchesCatalog` | `registered tool count = 21, catalog count = 25 (AP-C-4: a registered tool has no catalog entry)` — 카탈로그 4항목 추가 직후, 등록 전 |
| 핸들러 배선 (register) | `TestSessionMsgRegisterHandlerReturnsStableAgentID` (외 4건) | `--- FAIL: TestSessionMsgRegisterHandlerReturnsStableAgentID ... result has no structured content: &{...}` — 빈 스텁 핸들러 대비 전 5 wiring 테스트 FAIL (무결 RED: nil-panic 없이 assertion 실패) |

**AC 이항 검증 매트릭스 (M2 소관)**

| AC | Status | Verification Command | Actual Output |
|----|--------|---------------------|---------------|
| AC-CSM-007 | PASS | `go test ./internal/cli/ -run TestMoaiMCPServer_RegistrationMatchesCatalog -v` && `grep -c 'add("session_msg' internal/cli/mcp_server.go` | `--- PASS: TestMoaiMCPServer_RegistrationMatchesCatalog` · grep = `4` |
| AC-CSM-008 (M2 부분) | PASS | `grep -rn "exec.Command\|codex-jobs\|app-server\|net.Listen\|http.Listen" internal/sessionmsg/ internal/cli/mcp_session_msg.go \| grep -v _test` | 0행 (exit 1) |
| AC-CSM-009 | PASS | `go test ./internal/cli/ -run TestSessionMsg_NoAskUserQuestion -v` | `--- PASS: TestSessionMsg_NoAskUserQuestion` (NoInlineGetenv 동일 PASS) |
| AC-CSM-010 (M2 부분) | PASS | `grep -rn 'os.Getenv("' internal/sessionmsg/ internal/cli/mcp_session_msg.go \| grep -v _test` | 0행 (exit 1) |
| AC-CSM-015 | PASS | `grep -c "a reply is not user approval" internal/cli/mcp_session_msg.go` | `1` (규율 상수 선언 — 등록 블록이 상수를 연결해 4개 도구 설명 전부에 도달; `TestSessionMsgToolsRegisteredWithHintsAndDiscipline`가 ListTools 실측으로 4개 설명 전부 토큰 포함 단언) |

**품질 게이트 (M2)**

| 항목 | 명령 | 축어 출력 |
|---|---|---|
| 전 M2 배터리 | `go test ./internal/cli/ -run 'TestSessionMsg\|TestMoaiMCPServer_RegistrationMatchesCatalog' -v` | 8 test 전부 `--- PASS` · `ok github.com/modu-ai/moai-adk/internal/cli` |
| 카탈로그 파생 소비자 | `go test ./internal/web/ -run TestMCPConsoleToolCountMatchesCatalog -v` + `go test ./internal/settings/` | `--- PASS: TestMCPConsoleToolCountMatchesCatalog` (25 도구 파생 무계수 고정) · `ok internal/settings` |
| E2 빌드 | `go build ./... && GOOS=windows GOARCH=amd64 go build ./...` | `E2_M2_BUILDS_OK` (exit 0 양측) |
| E5 린트 | `golangci-lint run --timeout=2m ./internal/sessionmsg/... ./internal/mcp/...` + `--timeout=3m ./internal/cli/` | `0 issues.` / `0 issues.` (baseline 구분 불필요 — 0) |
| 카탈로그 AC 선제 검사 | `grep -c "session_msg" internal/mcp/catalog.go` | `4` (M3 AC-CSM-012 첫 조항과 정합 — 주석에 도구명 열거를 넣지 않아 항목 4행만 카운트) |

**M2 산출물**: `internal/cli/mcp_session_msg.go`(핸들러 4 + 규율 상수 + 구조화 오류 `sessionMsgToolErr` — UnknownAgentError의 StructuredContent 보존) + `internal/cli/mcp_server.go`(등록 4건, add() 패턴, `session_msg_list`만 ReadOnlyHint=true — 나머지 3개는 명시적 false) + `internal/mcp/catalog.go`(4항목, WriteCapable 3 true/1 false) + 테스트 2파일(boundary 가드 2 + wiring 5). 도구 인자는 design.md §6 표와 동일(선택 인자 `?` 규칙 포함 — text는 Required).

Gaps (M2): 실주행 MCP 서버(t184 재시작 스큐 포함) 검증은 M4 e2e 소관 — 본 단계는 단위+가드+in-process ListTools 실측까지. `data` 인자의 MCP 와이어 전달(map→RawMessage 재인코딩)은 wiring 테스트가 직접 인자 주입으로만 검증 — 실제 호스트(Codex/Claude) 직렬화 경로는 M4.

## §F Phase 4 Mode Selection

Implementation Kickoff Approval: **통과 (운영자 승인 2026-08-23, 리드 경유 — 진행 모드: 반자율, 각 마일스톤 경계 리드 보고·승인, goal 엔진 무장 없음)**.

입력 변수: tier=L · scope≈6파일(M1: internal/sessionmsg 4-5파일 + defaults.go + 테스트) · 도메인 2(Go 코어 패키지·config) · 언어 혼합 Go 100% · 동시성 이득 LOW(코딩 집약 — Anthropic 코딩 과제 병렬성 주의) · Agent Teams 전제 미충족(명시 요청 없음).

| 모드 | 선택 | 한줄 근거 |
|---|---|---|
| direct | 아니오 | 신규 패키지 + TDD 사이클 — 오케스트레이터 직접 수행 금지 영역 |
| serial | **선택** | 코딩 집약 Tier L의 기본 경로(Anthropic 코딩 과제 병렬성 주의); 마일스톤별 순차 위임이 반자율 진행 모드의 경계 보고와 정합 |
| fanout | 아니오 | 도메인 2·연구 아님 — 3-5 밴드 근거 미충족 |
| sweep | 아니오 | ~30파일 기계 변환 아님 — 신규 코드 |

Decision: serial
Justification: M1은 단일 Go 패키지 + config 임계값의 코딩 집약 작업으로 진짜 병렬 가능 분해가 없다(Anthropic 코딩 과제 병렬성 주의). 반자율 진행 모드가 마일스톤 경계마다 리드 보고를 요구하므로 순차 위임이 보고 경계와 일치한다. M2-M4도 같은 형태라 재평가 없이 serial 유지(스코프 변형 시 재평가).

Phase 1 (Plan Audit Gate) 재실행 스킵: 최종 판정 PASS 0.987 ≥ Tier L 임계 0.85 · 판정 후 plan 산출물 해시 무변경(iter-2 감사 HEAD = 현 HEAD 4329f45e6) · 스킵 자격은 판정 재실행에만 적용 — Implementation Kickoff Approval은 별도 통과(위).

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
