# acceptance.md — SPEC-CODEX-SESSION-MSG-001

> AC 도메인 접두사 `AC-CSM-NNN` (Tier L 상한 25 중 15 — review-1 D1/D2로 014/015 추가). 각 항목은 이항 판정 가능해야 하고 검증 명령을 가진다. GEARS 의무는 spec.md §D(요구 계층)에 있다 — 이 문서는 검증 계층(Given-When-Then).

## §D. AC 매트릭스

| AC | 대상 REQ | 이항 판정 | 검증 명령 (요약) |
|----|----------|-----------|------------------|
| AC-CSM-001 | REQ-CSM-002 | 엔벨로프 JSON 직렬이 A2A 코어 키 6개(`messageId`,`contextId`,`taskId`,`role`,`parts`,`metadata`)와 Part `kind` 판별자를 가진다 | `go test ./internal/sessionmsg/ -run TestEnvelopeA2AAlignment -v` |
| AC-CSM-002 | REQ-CSM-003 | kind+name 재등록이 같은 `agentId`를 반환하고 기록 파일이 존재한다 | `go test ./internal/sessionmsg/ -run TestRegisterIdempotent -v` |
| AC-CSM-003 | REQ-CSM-005/006 | send가 pending 파일을 만들고, poll이 클레임해 반환하며, ack_ids가 claimed를 삭제한다 | `go test ./internal/sessionmsg/ -run TestSendPollAck -v` |
| AC-CSM-004 | REQ-CSM-007 | 클레임 TTL 초과 claimed 메시지가 다음 스윕에서 pending으로 환원된다 | `go test ./internal/sessionmsg/ -run TestClaimExpiryReturn -v` |
| AC-CSM-005 | REQ-CSM-008 | 메시지 TTL 초과분이 지연 삭제되고 poll이 `expired_count`로 보고한다 | `go test ./internal/sessionmsg/ -run TestMessageExpirySweep -v` |
| AC-CSM-006 | REQ-CSM-009 | 병렬 send×poll에서 메시지 총량이 보존된다(분실·중복 0) — `-race` | `go test -race ./internal/sessionmsg/ -run TestConcurrentSendPoll -v` |
| AC-CSM-007 | REQ-CSM-013 | 4개 도구가 등록·카탈로그 동등성 가드를 통과한다 | `go test ./internal/cli/ -run TestMoaiMCPServer_RegistrationMatchesCatalog -v` && `grep -c 'add("session_msg' internal/cli/mcp_server.go` → `4` |
| AC-CSM-008 | REQ-CSM-010 (+ REQ-CSM-001 "포트 없음" 조항) | session_msg 소스에 codex 스폰/잡/네트워크 리스너 토큰이 없다 | `grep -rn "exec.Command\|codex-jobs\|app-server\|net.Listen\|http.Listen" internal/sessionmsg/ internal/cli/mcp_session_msg.go \| grep -v _test` → 0행 |
| AC-CSM-009 | REQ-CSM-011 | C-HRA-008 정적 가드 통과 | `go test ./internal/cli/ -run TestSessionMsg_NoAskUserQuestion -v` |
| AC-CSM-010 | REQ-CSM-012 | 인라인 `os.Getenv("...")` 0건 + defaults.go에 임계값 5종 존재 | `grep -rn 'os.Getenv("' internal/sessionmsg/ internal/cli/mcp_session_msg.go \| grep -v _test` → 0행 && `grep -c "DefaultSessionMsg" internal/config/defaults.go` ≥ `5` |
| AC-CSM-011 | REQ-CSM-014 | 교리 확장 절이 본품+템플릿 미러에 존재 | `grep -c "Codex" .claude/rules/moai/workflow/cross-session-messaging.md` ≥ `3` && 동일 grep이 `internal/template/templates/.claude/rules/moai/workflow/cross-session-messaging.md`에서 ≥ `3` |
| AC-CSM-012 | REQ-CSM-013/015 | 카탈로그 규칙이 25도구와 재시작 전제를 반영 | `grep -c "session_msg" internal/mcp/catalog.go` = `4` && `grep -n "session_msg" .claude/rules/moai/core/moai-mcp-tools.md` ≥ 1행 && `grep -c "재시작\|restart" .claude/rules/moai/core/moai-mcp-tools.md` ≥ `1` (본품+미러 양측) |
| AC-CSM-013 | 카드 납품 목표 | Claude 세션 × Codex 세션 실주고받기 왕복 관측 + 증거 기록 | §E 절차 실행 후 `grep -c "session-msg-e2e" .moai/specs/SPEC-CODEX-SESSION-MSG-001/progress.md` ≥ `1` |
| AC-CSM-014 | REQ-CSM-004 | `register`/`poll`/`send`가 하트비트를 갱신하고, 오프라인 임계(defaults.go) 초과 에이전트는 `session_msg_list`에서 `online:false`로 보고되며, 이후 세 호출 중 어느 하나가 다시 `online:true`로 되돌린다 | `go test ./internal/sessionmsg/ -run TestHeartbeatOnlineOffline -v` |
| AC-CSM-015 | REQ-CSM-014 (후반부 — 도구 설명 의무) | 신규 도구 설명 문자열이 규율 짧은 형태(토큰 `a reply is not user approval`)를 실는다. **base-0 실측**: 사전구현 트리에서 `grep -rn "a reply is not user approval" internal/ cmd/ pkg/` → 0행, `.claude/` → 0행 (2026-08-23, vacuous-pass 방지) | `grep -c "a reply is not user approval" internal/cli/mcp_session_msg.go` → ≥ `1` |

## §E. e2e 절차 스케치 (AC-CSM-013 — 실주고받기)

전제 위생: 모든 Codex 동작은 codex-cli 0.147.0 기준(감사 §4). 테스트는 직렬 1회 — 병렬 부하 금지(CLAUDE.local.md §4·§6).

- **0) 재시작 전제 (REQ-CSM-015, t184)**: `make build && rm -f ~/go/bin/moai && cp bin/moai ~/go/bin/moai` (§11 rm+cp 규율). 새 도구는 MCP 서버 재시작 전에 인식되지 않는다 — 1) 이후 여는 세션은 전부 새 바이너리로.
- **1) Codex 격리·등록**: `CODEX_HOME=<scratch>/home` 격리 → `codex mcp add moai -- moai mcp-server` → `codex mcp list` enabled 확인 (감사 §2.6 실증 경로).
- **2) 승인 정책 (P3)**: 대화형 codex 세션 사용 또는 exec에 승인 정책 구성 — 무정책 exec의 MCP 호출은 "user cancelled MCP tool call"로 실패한다.
- **3) Claude 세션 등록**: `session_msg_register(kind="claude", name="<식별명>")` → agentId A 관측.
- **4) Codex 세션 등록·발견**: Codex 세션에서 `session_msg_register(kind="codex", name="<식별명>")` → agentId B; `session_msg_list`에서 A·B 양측 관측.
- **5) Claude→Codex 발신**: Claude 세션에서 `session_msg_send(from=A, to=B, text=...)` → messageId 반환 관측.
- **6) Codex 수신·확인**: Codex 세션에서 `session_msg_poll(agent_id=B)` → 메시지 수신 관측 → 다음 poll `ack_ids`로 삭제 확인.
- **7) Codex→Claude 회신 (왕복 성립)**: Codex에서 `session_msg_send(from=B, to=A, text=...)` → Claude 세션 `session_msg_poll(agent_id=A)` 수신 관측.
- **8) 증거 기록**: 왕복 관측 로그(도구 호출·응답 축어)를 `progress.md` §E.2에 `session-msg-e2e` 마커와 함께 기록.
- **9) 위생 확인**: 실제 `~/.codex/config.toml`·`hooks.json` 무변경(mtime·해시) — 격리가 새지 않았음을 증명 (감사 §4 권장).

실패 시: 절차 단계·관측 출력·원인 가설을 §E.2에 기록 후 블로커 보고 — 재시도 3회 상한.

## §F. Given-When-Then 시나리오 (검증 계층 — 최소 2, 전 AC 이항 판정)

- **AC-CSM-003** — Given 두 등록된 에이전트(A 송신, B 수신)와 빈 B 사서함, When `session_msg_send(A→B, "hello")` 후 `session_msg_poll(B)`를 호출하면, Then poll 응답에 `messageId`가 일치하는 메시지가 정확 1회 포함되고 `mailbox/B/claimed/<messageId>.json`가 존재한다.
- **AC-CSM-004** — Given 클레임된 메시지 1건(`ClaimedAt` = 현재-11분, 클레임 TTL 10분), When `session_msg_poll(B)`를 다시 호출하면, Then 같은 메시지가 pending에서 재클레임되어 반환된다(at-least-once).
- **AC-CSM-006** — Given 등록된 수신자 B, When 서로 다른 고루틴에서 10개 세션이 각 10건을 send하고 동시에 2개 poller가 배치를 소비하면, Then 수신 총량(claimed+재환원)이 100과 같고 `-race` 보고가 0건이다.
- **AC-CSM-007** — Given 신규 4도구가 등록된 서버, When `TestMoaiMCPServer_RegistrationMatchesCatalog`가 실행되면, Then 카탈로그 집합과 tools/list 집합이 동등하고 `session_msg_list`만 ReadOnlyHint=true다.
- **AC-CSM-008** — Given 구현 완료 트리, When AC-CSM-008의 grep을 실행하면, Then 출력이 0행이다(위임 토큰 부재 — codex_task와의 경계).
- **AC-CSM-013** — Given 격리된 CODEX_HOME과 재시작된 양측 세션(절차 0-2), When 절차 3-7을 실행하면, Then Claude→Codex→Claude 왕복이 관측되고 증거가 §E.2에 기록된다.
- **AC-CSM-014** — Given 등록된 에이전트 B(마지막 하트비트 = 현재), When 가짜 시계로 오프라인 임계(30분)를 넘긴 뒤 `session_msg_list`를 호출하면 `online:false`이고, 이어 `session_msg_poll(B)`을 호출하면, Then 하트비트가 갱신되어 다음 `session_msg_list`에서 `online:true`로 보고된다.
- **AC-CSM-015** — Given 구현 완료 트리, When AC-CSM-015의 grep을 `internal/cli/mcp_session_msg.go`에 실행하면, Then 토큰 출현 수가 1 이상이다(사전구현 base-0 — 0행 시 통과가 아닌 실패).

## §G. 품질 게이트 (Definition of Done)

- [ ] REQ-CSM-001..015 전부 AC에 매핑·이항 판정 — 위 매트릭스 누락 0 (REQ-CSM-004 → AC-CSM-014 직접 검증; REQ-CSM-001의 "포트 없음" 조항 → AC-CSM-008의 `net.Listen|http.Listen` 확장 grep; 나머지 13 요구는 각자 전용 AC)
- [ ] `go test ./internal/sessionmsg/...` 초록 + `go test ./internal/cli/ -run 'TestMoaiMCPServer_RegistrationMatchesCatalog|TestSessionMsg'` 초록
- [ ] 패키지 커버리지 ≥ 85% (TRUST 5 Tested)
- [ ] `golangci-lint run` 신규 0건 (baseline 구분)
- [ ] `GOOS=windows GOARCH=amd64 go build ./...` exit 0
- [ ] C-HRA-008 가드·경계 grep·하드코딩 grep 전부 0건 (AC-008/009/010)
- [ ] 교리·카탈로그 본품+템플릿 미러 동시 반영 + `make build` (AC-011/012)
- [ ] e2e 왕복 관측 + §E.2 증거 + 격리 위생 확인 (AC-013)
- [ ] MX 태그: 신규 내보내기 함수에 @MX:NOTE/ANCHOR (fan_in ≥ 3 지점 — 카탈로그·스토어)
- [ ] 전 과정 5절 보고 형식(Claim/Evidence/Baseline/Gaps/Residual-risk) — 미검증 명시 포함
