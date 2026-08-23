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

### M3 — 교리 전문화 + Template-First (2026-08-23, 리드 승인 후 진행)

문서 전용 마일스톤 — Go 코드 변경 없음. **TDD/RED 비적용 명시**: 문서 마일스톤이라 RED-GREEN 사이클이 성립하지 않는다(패블리케이팅할 실패-선언 테스트가 없음) — RED 증거를 조작하지 않고 이 문장으로 갈음한다. 이항 판정은 AC grep으로 성립.

**AC 이항 검증 매트릭스 (M3 소관)**

| AC | Status | Verification Command | Actual Output |
|----|--------|---------------------|---------------|
| AC-CSM-011 | PASS | `grep -c "Codex" .claude/rules/moai/workflow/cross-session-messaging.md` + 동일 grep 미러 | 본품 `5` / 미러 `5` (≥3 양측; base-0에서 5로) |
| AC-CSM-012 | PASS | `grep -c "session_msg" internal/mcp/catalog.go` = `4` (M2 인용, 무변경) && `grep -n "session_msg" .claude/rules/moai/core/moai-mcp-tools.md` ≥1행 + 미러 && `grep -c "재시작\|restart"` 양측 ≥1 | catalog.go `4` · 본품 118-121행 4개 도구행 + 미러 `4` · restart 단어 본품 `1`/미러 `1` |

**PRESERVE #6 증명 (기존 조항 바이트 동일)**: `git diff` 제거 라인 전량 = 2줄 — ① `moai-mcp-tools.md` 헤더 `21 tools`→`25 tools` (카탈로그 갱신 위임의 본질적 일부), ② `cross-session-messaging.md` 버전 푸터 `1.2.0`→`1.3.0` (섹션 추가에 따른 버전 승격 — 조항이 아닌 메타데이터). 이 외 모든 기존 바이트 무변경 (확장 절 순수 추가).

**Template-First + 중립성**: 본품과 미러 동일 커밋 편집 + `make build` exit 0 (`catalog.yaml updated successfully (12899 bytes)` — 재생성 결과 바이트 동일, `git diff --name-only internal/template/catalog.yaml` = 0행, parity 위험 없음). 미러 중립성 grep `SPEC-CODEX-SESSION-MSG-001\|t187\|REQ-CSM` → 0행. 본품 SPEC 교차참조 1회(Origin 각주) — 미러에는 부재.

**rule-authoring (b)+(c) duty** (커밋 바디에 동일 문구 탑재):
- `cross-session-messaging.md` +1,842B (16,672→18,514; 미러 +1,779B): Codex 피어와 한 번도 메시지하지 않는 세션도 매 턴 + 매 `/clear`마다 이 확장을 다시 지불한다. 정당화: 브로커는 Codex 유일 경로이고 규율 조항은 사용 순간부터 구속되는데, 확장 절은 새 법이 아니라 이미 적재된 조항명의 재해석 표 — 신규 개념 부담 없이 기존 조항에 대한 표 매핑으로 지불을 최소화했다.
- `moai-mcp-tools.md` +1,380B (7,357→8,737; 본품·미러 동일): 브로커를 쓰지 않는 세션도 4개 카탈로그 행 + 재시작 문장을 지불한다. 정당화: 이 카탈로그는 에이전트가 MCP-vs-CLI 우선경로를 판정하는 유일한 지도 — 지도에 없는 도구는 라우팅 불가능하다. 재시작 문장은 "새 도구가 안 보인다" 디버깅 실패 재발(AP-5, t184 스큐)을 차단하는 1문장 보험이다.

Gaps (M3): CLAUDE.md §4 "existing 21 MCP tools" 표기는 본 마일스톤 범위 밖(본품 카탈로그 규칙만 위임됨) — sync-phase 정오후보로 기록. t192 5-tools 텍스트는 본 트리에 부재(리드 실측 76b2c4ece 기준) — 수정 대상 없음.

### M4 — e2e 실주고받기 (session-msg-e2e) — 2026-08-23, 직렬 1회, 실프로세스

**`session-msg-e2e`** — AC-CSM-013 관측 기록. 전 과정 실프로세스: Claude측 액터 = 실제 `bin/moai-dev mcp-server` 프로세스(stdio JSON-RPC 드라이빙, 드라이버: `.moai/reports/t187/e2e/drive_mcp.py`), Codex측 = 실제 `codex exec` 세션(codex-cli 0.147.0)이 스폰한 또 다른 `bin/moai-dev mcp-server`.

**0단계 — 리드 판정 이탈 기록(전역 재설치 금지)**: `~/go/bin/moai`(v3.1.3-rc.0, 릴리스 함대·훅 사용 중)을 재설치하지 않았다. 대신 `go build -o bin/moai-dev ./cmd/moai`(트리 로컬, `bin/`은 .gitignore:12로 비추적). 격리 CODEX_HOME의 `mcp_servers.moai.command`가 이 절대경로를 직접 지목. **전역 재설치는 병합 후 — main 기반 브랜치 바이너리의 전역 설치가 release 전용 수정을 regress시킴(리드 판정 2026-08-23)**.

**공유 스토어 루트(하중 재불변량)**: 양측 루트 = `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t187` (스토어: `.moai/state/session-msg/`). 증명 3중: (a) Claude측 기록의 `cwd` 필드 = 워크트리 절대경로, (b) codex 재등록 후 `session_msg_list`가 양측 에이전트 동시 관측(LISTCOUNT: 2, CLAUDE_AGENTS에 claude- 반환), (c) `find`로 양측 기록 파일이 같은 `agents/` 디렉터리에 존재. **1차 시도 실패 기록(근거 보존)**: `codex mcp add`를 env 없이 하면 codex가 MCP 자식 프로세스에 `CLAUDE_PROJECT_DIR`을 전달하지 않아(환경 새니타이즈) codex측 서버가 cwd(`/tmp/m4-e2e/work`)로 루트를 오해석 — `LISTCOUNT: 1`로 적발, codex 기록이 `/tmp/m4-e2e/work/.moai/state/...`에 잘못 착지한 것을 find로 확인. 수정: `codex mcp add moai --env CLAUDE_PROJECT_DIR=<worktree> -- <abs>/bin/moai-dev mcp-server`(per-server env 고정) 후 재실행.

**승인 정책(P3 실측 재확인)**: 시도 1 `-c approval_policy="never" -s read-only` → MCP 호출 2회 모두 `user cancelled MCP tool call`로 취소(감사 §2.6 그대로 재현). 시도 2 `--dangerously-bypass-approvals-and-sandbox` → 완주. 채택: 후자(프롬프트는 MCP 도구 호출만 지시, 셸 명령 없음, `timeout` + 격리 CODEX_HOME + 스크래치 cwd로 외부 구속).

**3-7단계 축어 증거** (도구 응답의 structuredContent/최종 보고 축어):

```
# 3) Claude측 등록 (드라이버 → bin/moai-dev mcp-server, tools/call session_msg_register)
{"content":[{"type":"text","text":"session_msg_register: ok"}],"structuredContent":{"agentId":"claude-ed1c3afc","kind":"claude","name":"claude-e2e","description":"M4 e2e Claude-side actor","version":"1","capabilities":{"messaging":true},"cwd":"/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t187","pid":27305,"host":"goos.local","registeredAt":"2026-08-23T12:03:18.377971Z","lastHeartbeat":"2026-08-23T12:03:18.377971Z"}}
→ agentId A = claude-ed1c3afc

# 4) Codex측 등록 + 발견 (codex exec, 신규 세션)
mcp: moai/session_msg_register (completed)
mcp: moai/session_msg_list (completed)
AGENTID: codex-79279481
LISTCOUNT: 2
CLAUDE_AGENTS: claude-ed1c3afc
→ agentId B = codex-79279481; 목록에 A·B 양측 관측 (공유 루트 증명)

# 5) Claude→Codex 발신 (드라이버 tools/call session_msg_send)
{"content":[{"type":"text","text":"session_msg_send: ok"}],"structuredContent":{"from":"claude-ed1c3afc","messageId":"msg-57d40535d4a2f107","to":"codex-79279481"}}
→ messageId = msg-57d40535d4a2f107

# 6) Codex 수신·확인·회신 (codex exec, 자기 messageId 추출 후 ack)
mcp: moai/session_msg_send (completed)
RECEIVED_TEXT: hello from claude over the session-msg broker (M4 e2e)
ACKED_COUNT: 1
SENT_MESSAGE_ID: msg-e1b0259fedda5c0b

# 7) Codex→Claude 회신 수신 (드라이버 tools/call session_msg_poll, Claude측)
{"content":[{"type":"text","text":"session_msg_poll: ok"}],"structuredContent":{"ackedCount":0,"expiredCount":0,"messages":[{"message":{"messageId":"msg-e1b0259fedda5c0b","role":"agent","parts":[{"kind":"text","text":"reply from codex: round trip complete"}]},"delivery":{"senderId":"codex-79279481","senderKind":"codex","sentAt":"2026-08-23T12:08:48.114333Z","expiresAt":"2026-08-24T12:08:48.114333Z","claimedAt":"2026-08-23T12:08:58.895267Z"}}],"remaining":0}}
→ 왕복 성립: 송신문 축어 수신(RECEIVED_TEXT 일치), ack 삭제 확인(ACKED_COUNT 1), 회신 수신(senderKind codex, A2A camelCase 엔벨로프 관측)
```

**t184 재시작 전제 실측**: 드라이버 매 호출이 신규 서버 프로세스 — `tools/list`가 25도구 + session_msg 4종을 즉시 관측: `{"init_server":{"name":"moai","version":"v3.1.2"},"total_tools":25,"session_msg_tools":["session_msg_list","session_msg_poll","session_msg_register","session_msg_send"]}`.

**9단계 위생 증명**: 실제 `~/.codex/config.toml`·`hooks.json`의 SHA-256 + mtime before/after 완전 일치(`diff` 무출력, `HYGIENE_UNCHANGED`). 격리 누수 없음. 잔여 프로세스 `pgrep -lx moai-dev` → 0 (codex 종료 시 MCP 서버 회수 확인). 격리 스크래치: `/tmp/m4-e2e/`(home·work), 드라이버·로그: `.moai/reports/t187/e2e/`(비추적).

**스토어 착지 증명**: `agents/{claude-ed1c3afc,codex-79279481}.json` 양측 공유 루트에 존재; `mailbox/claude-ed1c3afc/claimed/msg-e1b0259fedda5c0b.json`(회신 클레임 상태 — at-least-once 시맨틱대로 ack 전 보존), codex 사서함은 ack 완료로 empty.

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
