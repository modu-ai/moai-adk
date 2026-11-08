# 세션 간 양방향 통신·이종 코딩 에이전트 상호운용 조사 (2026-11-08)

이 문서는 읽기 전용 조사 결과이며 구현 승인이 아니다. 주 모델은 "MoAI Rust 런처가 띄운 세션을 위한, 필요할 때만 켜는 선택적 경량 허브"다. GitHub Issues/PR이 소유·진행·완료의 영구 원장이고 허브는 실시간 전달층이다. 모든 미결 선택은 HOLD(§15)로 사용자 브레인스토밍에서 결정한다. 관련 Issue 반영 위치는 §14.

조사 세션 메모: 하위 조사 5갈래의 보고가 조정자 주소로 전송되어 세션 기록에서 회수·통합했고, 그중 Claude 문서 갈래는 경량 모델로 실행되어 1차 문서 재확인 조건이 붙어 있다. 관측된 로컬 토큰 1건은 보고서에서 제외했다.

All five research lanes have in fact completed their work (their reports went to `team-lead` via SendMessage; I recovered the full bodies from the subagent transcripts in this session's own directory, read-only). Nothing further is pending. Below is the combined report.

---

# moai `-l`(leader) / `-a`(agent) 세션별 양방향 통신 — 조사 보고 (설계 입력용)

- 조사일: 2026-11-08. 표기: **검증**(공식 문서·스펙·소스에서 확인), **직접확인**(이 환경에서 읽기 전용으로 관측), **미확인**(공식 근거 없음), **미지원**(문서가 명시적으로 불가).
- 범위 원칙: Go 세대(`/Users/goos/MoAI/moai-adk-go`)는 참고 전용. 설계 근거는 전부 현행 공식 문서.
- 전제(사용자 정정 반영): 1차 모델은 **MoAI Rust 런처가 자신이 띄운 세션들을 위해 필요할 때만 켜는 경량 로컬 허브**. Orca/Hide는 비교 참고일 뿐 전제가 아니며, 상시 데몬·원격 호스팅은 가정하지 않는다. GitHub Issues/PR이 소유권·진행·완료의 내구 원장이고, 허브는 실시간 전달층에 그친다.
- 조사 방식 메모: 하위 조사 5개 레인은 보고를 `team-lead` 주소로 보냈고 저에게는 도착하지 않았습니다. 각 레인의 전송 본문을 이 세션의 subagent 기록(`~/.claude/projects/.../subagents/*.jsonl`)에서 읽기 전용으로 회수해 통합했습니다. 하위 레인 중 하나(Claude 문서 레인)는 `haiku` 모델로 실행됐으므로 해당 레인 인용은 1차 문서 재확인 조건을 붙였습니다.

## 1. 핵심 결론

1. **"한 호스트가 유휴/작업 중에 비동기 메시지를 받아 다음 턴에 반영한다"는 능력은 호스트마다 전혀 다르고, 공식 문서로 확인되는 것은 Claude Code(cross-session messaging, `-p` stream-json 큐), Codex app-server(`turn/steer`, `thread/inject_items`), OpenCode 서버(`POST /session/:id/message` 합류, dev 브랜치 inbox) 셋뿐이다.** ZCode는 미확인, ACP는 스펙상 미지원(취소·모드 변경만 허용).
2. **"MCP가 연결돼 있으면 비동기 메시지를 읽는다"는 가정은 틀렸다.** Claude Code에서 모델 턴에 주입되는 서버→클라이언트 푸시는 preview 기능 `channels`(`--channels`, `notifications/claude/channel`)만 문서화돼 있고, Codex/OpenCode는 범용 서버 알림의 모델 주입이 미확인이다. 최신 MCP 스펙(2026-07-28)은 서버 발신 요청을 MRTR로 대체하고 Roots/Sampling/Logging을 폐기했다.
3. **허브는 "도착 → 호스트 수신 → 모델 턴 반영 → 승인된 작업 실행"의 4단 영수증 사다리를 구분해야 하며, 마지막 단은 어떤 호스트에서도 메시지로 증명되지 않는다.** 증명은 GitHub PR/커밋 SHA/CI/검증 파일에서만 읽는다.
4. **GitHub Issues는 CAS가 없다**(문서가 POST/PUT/PATCH/DELETE 조건부 요청 미지원을 명시). assignee·label은 잠금이 될 수 없고, 서버가 부여하는 단조 증가 comment `id`를 fencing token으로 쓰는 "최소 id 생존 claim 승리" 규약 + 리더의 읽기 측 검증이 유일한 현실적 경로다(설계 제안으로 표기). 이 저장소에서 `dependencies/blocked_by`·`sub_issues` 엔드포인트는 HTTP 200으로 **직접확인**됐다.
5. 세션 간 공통 계약 후보로 **ACP(Agent Client Protocol)** 가 가장 넓게 걸친다(OpenCode·Gemini CLI 네이티브, Claude/Codex는 ACP 재단 어댑터, Rust crate `agent-client-protocol` 5.0.0). 단 벤더 공식 소유는 미확인이고, 스펙이 세션당 1턴·중간 메시지 없음이라 "허브→세션 비동기 알림"에는 부적합하다. 결론은 **원장(GitHub) + 호스트 네이티브 어댑터 + 선택적 허브** 조합이며, 단일 프로토콜 강제는 HOLD.

## 2. 호스트별 사실 기반 지원표

버전: Claude Code 2.1.319(직접확인, `claude --version`), Codex CLI 0.163.0(직접확인; GitHub 최신 안정 rust-v0.167.0 2026-11-06 검증), ZCode v3.17.2(README 2026-11-03 검증; 로컬 PATH에 없음), OpenCode v1.19.5(2026-10-28 검증; 로컬 미설치), ACP protocolVersion 1(검증).

| 항목 | Claude Code | OpenAI Codex | Z.ai ZCode | OpenCode | ACP(공통 계약) |
|---|---|---|---|---|---|
| 비대화 실행 | 검증 `claude -p`, `--output-format stream-json`, `--bare`(훅·MCP·CLAUDE.md 생략, API 키 전용), `--bg` + `claude attach/logs/stop/rm/agents` — code.claude.com/docs/en/cli-reference, headless | 검증 `codex exec [--json] [-C dir] [--worktree] [--ephemeral] [-o file]`; `--full-auto`는 "no longer supported"; 승인은 `-c approval_policy=never` — developers.openai.com/codex/noninteractive, cli/reference | 검증(저장소 문서) `zcode --prompt "…" --cwd … --mode plan|yolo --json --output-format stream-json --disallowed-tools …`; `--prompt` 기본이 yolo("not read-only, not a sandbox") — github.com/zai-org/ZCode apps/zcode-cli/docs/usage/13-automation.md | 검증 `opencode run --format json --session <id> --continue --auto --standalone --server <url>` — docs.opencode.ai/docs/cli, dev commands.mdx | 해당 없음(클라이언트가 에이전트 프로세스를 spawn) |
| SDK/서버/프로토콜 | 검증 Agent SDK TS/Python(`query()`, streaming input, `canUseTool`, `interrupt()`, `resume/forkSession`); stdin stream-json; `--permission-prompt-tool mcp__…`, `--permission-prompts host|none`(v2.1.259+). stdio control 요청 와이어 스키마는 **미확인**(미공개) | 검증 `codex app-server` JSON-RPC: `--listen stdio://|ws://IP:PORT|unix://[PATH]|off`, `--ws-auth capability-token|signed-bearer-token`; `generate-ts/generate-json-schema`; 상태 "experimental… may change without notice". `codex mcp-server`는 **제거됨** | 미확인(CLI 가이드에 hooks/plugins 장은 있으나 본문 미열람) | 검증 `opencode serve`(기본 127.0.0.1:4096, `GET /doc` OpenAPI 3.1, `OPENCODE_SERVER_PASSWORD`), SDK `@opencode-ai/sdk`; `opencode acp` | 검증 JSON-RPC 2.0, 개행 구분, stdio만 표준; 원격 전송은 RFD 단계 — agentclientprotocol.com/protocol/transports |
| 이벤트 스트림 | 검증 stream-json(`system/init`, `--include-hook-events`, `--replay-user-messages`); 훅 `Stop/SessionEnd/Notification` | 검증 exec `--json` JSONL(`thread.started`, `turn.started/completed/failed`, `item.*`); app-server `turn/*`, `item/*`, `thread/status/changed{waitingOnApproval}`; `notify`(agent-turn-complete만); hooks(`Stop`, `PostToolUse`, `PermissionRequest`, …, `features.hooks`) | 검증(저장소) `--output-format stream-json`, JSON에 sessionId/traceId/turnId | 검증 SSE `GET /event`(디렉터리별), `GET /global/event`; v1.19.5 이벤트명 `session.status`, `session.idle`, `permission.updated`, `message.part.updated` 등(SDK 생성 타입) | 검증 `session/update` 알림 14종(`agent_message_chunk`, `tool_call`, `plan`, …) |
| 세션 재개 | 검증 `--session-id <uuid>`, `-r/--resume <id|name|path>`, `-n/--name`, `--fork-session`, 기록 `~/.claude/projects/<slug>/<id>.jsonl`(스키마는 미확인) | 검증 `codex exec resume <id>`, `codex fork`, `thread/resume{threadId}`, `codex queue --thread <id>`; rollout 경로 `~/.codex/sessions/YYYY/MM/DD/rollout-*.jsonl`은 **직접확인**이나 문서 미확인, 형식 "not stable" | 검증(저장소) `--resume <sessionId>`, `--continue` | 검증 `--session`, `--continue`, `session export/import`, DB `~/.local/share/opencode/opencode.db` | 검증 `session/load`(선택 capability `loadSession`), `session/resume` 게이트 |
| 중단/취소 | 검증 SIGINT=턴 종료, SIGTERM=exit 143(턴 미완); SDK `interrupt()`; stream-json `interrupt_cancel_queued_v1` capability. CLI stdin 인터럽트 메시지 스키마 **미확인** | 검증 `turn/interrupt{threadId,turnId}` → `turn.status=interrupted`, 대기 중 승인요청 `serverRequest/resolved`로 정리; exec는 SIGINT→내부 interrupt(소스 관측, 문서 미확인) | 미확인 | 검증 `POST /session/:id/abort`; 승인 대기 중 abort도 run 중단 | 검증 `session/cancel` 알림 → `stopReason: cancelled`(오류 아님) |
| 로컬/클라우드 | 검증 `--cloud`(생성·후속 `-p … --cloud <id>`), `--teleport`(단방향), Remote Control(`claude remote-control --spawn worktree --capacity 32`; 스크립트 구동 API **미확인**; 구독 필요, `ANTHROPIC_BASE_URL` 비Anthropic 시 불가) | 검증 `codex cloud exec|status|list|apply|diff`(EXPERIMENTAL, 폴링만), `codex remote-control`, `codex --remote ws://…` | 검증 데스크톱·웹·터미널 UI, "mobile remote control"(docs.z.ai/devpack/tool/zcode); 세부 미확인 | 검증 로컬만; OpenCode Zen은 모델 게이트웨이(실행 아님); share 중단 | 로컬 stdio만 |
| 동시성 | 검증 worktree가 공식 병렬 수단(`-w` → `.claude/worktrees/<name>`, 브랜치 `worktree-<name>`, `EnterWorktree/ExitWorktree`, `.worktreeinclude`); Agent Teams는 세션 1팀, 세션 간 공유 불가 | 검증 app-server 1프로세스 다중 thread·다중 클라이언트 구독; `thread/list{cwd}` 필터; Desktop은 chat당 worktree(`$CODEX_HOME/worktrees`); exec 다중 프로세스 동일 repo 명시 **미확인** | 검증(저장소) `--continue`는 동시 실행에 비권장, `--cwd`·`ZCODE_DATA_BASE_DIR` 격리 | 검증 사용자당 공유 서버 1개·세션 다수, `--standalone` 전용 서버 | 검증 프로세스당 다중 세션, 세션당 1턴 |
| 권한/승인 | 검증 `--permission-mode acceptEdits|auto|bypassPermissions|manual|dontAsk|plan`, `--allowedTools/--disallowedTools`, `PermissionRequest` 훅, `canUseTool`; `-p` 대화상자 불가 | 검증 `approval_policy on-request|never|granular`(`untrusted` 미지원), `sandbox_mode read-only|workspace-write|danger-full-access`; **workspace-write에서 `.git`는 보호 경로라 `git commit`에 승인/에스컬레이션 필요**; app-server `item/*/requestApproval` → `accept|acceptForSession|decline|cancel`; 승인 타임아웃 미확인 | 검증(저장소) `--mode plan|yolo`, `--disallowed-tools` | 검증 `permissions[]{action,resource,effect allow|deny|ask}`; 런타임 `POST /session/:id/permissions/:permissionID {response: once|always|reject}`; `run --auto` | 검증 `session/request_permission` `allow_once|allow_always|reject_once|reject_always`; 정책은 클라이언트 측 |
| 비동기 수신 메시지 | 검증(§3) | 검증(§3) | **미확인** | 검증(소스)+docs 미확인(§3) | **미지원**(§3) |
| 이 환경 직접확인 | `claude --help` 플래그(`--name`, `--bare`, `--worktree`, `--fork-session`, `--input-format`, `--remote-control`, `--teleport`, `--cloud`) 존재; 소켓·레지스트리(§4.5) | `codex --version`, `--help`, `codex features list`(`hooks stable true`, `worktrees stable true`, `steer removed true` 등) | `which zcode` 없음; `~/.zcode/` 디렉터리 존재(이름만) | 미설치 | 미설치(crate만 docs.rs 확인) |

## 3. 호스트별 비동기 메시지 수신 동작 (유휴 / 작업 중 / 승인 대기 / 다음 턴 주입)

| 호스트·경로 | 유휴 시 | 작업 중 | 승인 프롬프트 대기 중 | 다음 턴 주입 | 근거 |
|---|---|---|---|---|---|
| Claude Code cross-session `SendMessage`(v2.1.224+, Windows 2.1.234+) | 검증: 새 턴 시작 | 검증: 툴 호출 사이에 읽음, 실행 중 툴은 중단 안 함; 미전달분은 transcript "Waiting to be delivered"(2.1.309+) | **미확인**(문서 침묵) | 검증: `<cross-session-message from=…>`로 래핑, "사용자가 아님"으로 전달, 승인 프롬프트는 그대로 발생, 슬래시 명령은 평문 | code.claude.com/docs/en/cross-session-messaging |
| 위 경로의 보류(hold) | 검증: 수신 기본값은 두 세션의 권한 클래스로 결정(수신자가 bypass면 송신자도 bypass여야 즉시 전달); 대화형은 승인 대화상자, `-p`는 대화상자 없이 `dialogExpiry`(기본 5m, `60s|5m|10m|never`) 뒤 폐기·송신자에 만료 통지; `crossSessionInbound accept|hold|refuse`; 보류 최대 100건, 큐 최대 50건, 송신 폭주 거부 | | | | cross-session-messaging, settings-reference `#crosssessioninbound` `#dialogexpiry` `#isolatepeermachines` |
| Claude Code `-p --input-format stream-json` | 검증: 큐 메시지가 턴 종료 후 새 턴 시작 | 검증: 큐잉(거부 아님); SDK `SDKUserMessage.priority` `next`(툴 후 같은 턴)/`later`(턴 종료 후)/`now`(인터럽트, 2.1.286+ `origin.kind=human`이면 양보) ; `shouldQuery:false`로 턴 미시작 첨부 | 미확인 | 검증 | agent-sdk/streaming-vs-single-mode, agent-sdk/typescript |
| Claude Code MCP 클라이언트 | 검증: `tools/prompts/resources list_changed`는 목록 갱신만; elicitation form/url 지원(비대화형에서는 서버에 취소 요청); sampling 미확인; **모델 턴 주입은 `channels`만**(preview, `--channels`, 송신자 allowlist, 수신 ACK 없음, 바쁠 때 다음 턴에 묶음 전달) | | | | docs/en/mcp, channels-reference |
| Claude Code 훅 | 검증: `UserPromptSubmit`/`SessionStart` `hookSpecificOutput.additionalContext`(10,000자 초과 시 파일), `SessionStart.initialUserMessage`, `Stop`에서 continue; `Notification` 훅 출력의 모델 전달은 미확인 | | | | docs/en/hooks |
| Codex app-server | 검증: `turn/start` | 검증: `turn/steer{threadId,input,expectedTurnId}`(활성 턴에 추가, 새 `turn/started` 없음), `thread/inject_items`(턴 미시작 기록 추가); 바쁜 thread에 일반 `turn/start`는 활성 턴에 접힘(소스 주석), 큐/즉시 여부 **미확인** → 명시적 `turn/steer` 권장 | 검증: 서버 발신 JSON-RPC 요청에 클라이언트가 응답할 때까지 item `inProgress`; 서버측 타임아웃 없음(미확인); interrupt 시 `serverRequest/resolved`; 재접속 클라이언트에 pending 재전송(소스), humanCommand는 지정 연결만 | 검증 | developers.openai.com/codex/app-server |
| Codex `codex exec` | — | **미지원**: stdin은 시작 시 1회 소비; 후속은 `exec resume <id>` 또는 `codex queue --thread`(데몬 필요) | — | — | noninteractive, cli/reference |
| Codex MCP 클라이언트 | 검증: tools, resources(`mcpServer/resource/read`), elicitation(`mcpServer/elicitation/request`, `approval_policy.granular.mcp_elicitations`); prompts/sampling/roots/범용 서버 알림 **미확인** | | | | developers.openai.com/codex/mcp |
| OpenCode 서버 | 검증 | 검증(소스 v1.19.5): `POST /session/:id/message`가 진행 중 run에 합류(step 경계에서 반영), 거부 없음; dev: `delivery: queue|steer`, `session.inbox.*` | 검증(소스): `permission.updated` 후 Deferred 대기, 턴 차단, `session.status=busy`; `reject`는 다른 pending도 거부, `always`는 패턴 저장 | 검증 | docs.opencode.ai/docs/server + 소스 |
| ZCode | 미확인 | 미확인 | 미확인 | 미확인 | — |
| ACP | 검증: `session/prompt` | **미지원**: 중간 메시지 없음(`session/cancel`, `session/set_mode`만) | 검증: `session/request_permission` 응답까지 턴 차단; cancel 시 클라이언트가 `cancelled`로 응답 | 해당 없음 | agentclientprotocol.com/protocol/prompt-turn |

직접확인 1건: 이 세션(Claude Code 2.1.319, in-process teammate 경로)에서 상위 조정자의 메시지가 제가 툴 결과를 받는 경계에서 "The coordinator sent a message while you were working" 형태로 **작업 중 주입**됐습니다. 위 "툴 호출 사이에 읽음" 문서 서술과 일치합니다.

## 4. 전송 계층 비교

### 4.1 비교표

| 방식 | 사실 | 장점 | 한계 | 적합도(허브용) |
|---|---|---|---|---|
| IRC형 서버/프로토콜(RFC 1459·2812, IRCv3 `msgid`·`echo-message`·`server-time`·`chathistory`) | 검증: 기본 IRC에 수신 영수증·영속성 없음, 오프라인 수신자는 메시지 유실; IRCv3가 메시지 id·송신 확인·타임스탬프·선택적 재생 추가 | 다대다·채널·브로드캐스트 모델이 성숙 | 전달 영수증 부재, 바이너리/구조화 페이로드 부적합, 별도 서버 프로세스 필요 | 참고 모델로만(채널·msgid 개념). 상시 서버 전제와 충돌 |
| MCP tools + notifications(2025-06-18 / 2025-11-25 / 2026-07-28) | 검증: stdio 개행 구분, Streamable HTTP; 2026-07-28은 stateless(initialize 제거, `server/discover`, `subscriptions/listen`, MRTR, tasks 확장, Sampling/Roots/Logging 폐기) | 모든 호스트가 "클라이언트"로 붙을 수 있음, 도구 호출로 폴링형 수신(inbox read) 가능 | **서버→모델 푸시는 호스트 의존**(Claude channels preview만, Codex/OpenCode 미확인); 호스트가 MCP 서버를 spawn하므로 허브 수명이 세션에 묶임; 스펙 대개편 진행 중 | "읽기 도구"로는 적합, "푸시"로는 부적합 |
| CLI send/read/wait(허브 CLI를 세션이 Bash로 호출) | 검증: Codex `notify`/hooks, Claude hooks, OpenCode SSE가 "밖으로" 내보내는 경로는 모두 있음; "안으로"는 세션이 `moai inbox read`를 호출할 때만 | 구현 단순, 권한 모델이 각 호스트의 Bash 승인에 종속돼 안전, 설치 부담 0 | 세션이 읽지 않으면 모름(폴링), 바쁠 때 주입 불가 | 기본선(모든 호스트 공통 최소) |
| 로컬 UDS / Windows named pipe / loopback | 검증: SOCK_STREAM은 순서·신뢰 보장, 경계 없음→프레이밍 필요(LSP `Content-Length` 또는 JSON Lines); 발신자 검증 Linux `SO_PEERCRED`, macOS `getpeereid`/`LOCAL_PEERCRED`, Windows named pipe SD(기본 DACL이 Everyone 읽기), Windows AF_UNIX는 보조 데이터 없음(SCM_CREDENTIALS 불가)·socketpair 없음 | 저지연, 로컬 신원 확인 가능, Rust `tokio`/`interprocess` 2.4.7로 양 OS 추상화 | 호스트가 소켓을 "읽어 모델에 넣는" 경로는 호스트 네이티브(Claude 소켓, Codex ws app-server, OpenCode HTTP)만 | 허브↔어댑터 내부 전송으로 적합 |
| 파일 메일박스(atomic rename, Maildir `tmp/→new/→cur/`) | 검증: POSIX `rename()` 원자성, Maildir 무잠금 전달 | 데몬 없음, 크래시 내성, 모든 호스트가 파일 읽기 가능, worktree 어디서든 가시 | 알림은 폴링/`notify` 크레이트 필요, Windows 교차 프로세스 rename 의미 재확인 필요(미확인) | 큐·영수증의 **내구층**으로 적합(허브 없는 평시 모드) |

### 4.2 Claude Code 자체 소켓 버스 — 읽기 전용 관측

- 문서(검증): 세션별 inbox 소켓, Unix 소켓이 기본이며 `/tmp/cc-socks-<uid>`로 폴백, Windows는 named pipe; 환경변수 `CLAUDE_CODE_MESSAGING_SOCKET`, `CLAUDE_CODE_MESSAGING_TOKEN`; **자기 세션의 자식 프로세스**가 `{"type":"auth","token":"…"}` 뒤 `{"type":"user","message":{"role":"user","content":"…"}}`를 소켓에 써 넣는 것은 문서화(바이너리 도움말에도 `socat - UNIX-CONNECT:` 예시). 외부 비자식 프로세스가 **타 세션** 소켓에 쓰는 것은 미확인(미문서).
- 직접확인: `/tmp/cc-socks/` 아래 `<pid>.sock` 90개, 소유 프로세스 전부 `claude` 2.1.319(대화형 `moai cc` 레인들과 Hide의 `--output-format stream-json --input-format stream-json` SDK 프로세스 포함 → `-p`도 소켓 바인드한다는 문서와 일치). `~/.claude/sessions/<pid>.json` 레지스트리 필드: `pid, sessionId, cwd, startedAt, procStart, version, peerProtocol, peerFeatures, kind, entrypoint, pidDomain, messagingSocketPath, name, nameSource, nameSince, status, updatedAt, statusUpdatedAt` + `<pid>.<hash>.key`. **이 레이아웃·경로는 미문서 내부 구현이므로 설계가 의존하면 안 되는 관측치**로만 기록.
- `ListAgents`/`SendMessage`는 모델 도구이며 CLI API가 아님(검증) → **외부 Rust 프로세스는 이 버스에 직접 참여할 수 없다**. 허브가 Claude 세션에 "밀어 넣는" 유일한 문서 경로는 `-p` stdin(큐) 또는 자식 훅의 자기 소켓 투입.

### 4.3 Go 세대 교훈(참고 전용, 경로만)

`cross-session-messaging-detail.md`가 기록한 운영 사실: 전송 결과 3형(`queued` / `routing` 객체=유실 / `[Cross-session delivery notice]`=보류·거부), 폭주 거부, `tengu_harbor_kite` 공유 플래그 슬롯이 서드파티 백엔드(GLM) 세션의 채널을 끊는 현상, `CLAUDE_CODE_HARBOR_KITE=1` 비공개 탈출구. 이 중 플래그 슬롯·환경변수는 **미문서 내부**라 현행 설계 근거로 쓰지 않음. 재사용 가치가 있는 개념: "메시지는 nudge, 위임은 디스크 큐", "idle notice는 스케줄 힌트", "stopped-teammate 재가동 금지".

## 5. ACK/영수증 사다리 (설계 제안)

| 단계 | 의미 | 증명 신호(호스트별) | 결코 아닌 것 |
|---|---|---|---|
| L0 hub-accepted | 허브(또는 파일 큐)가 내구 기록에 append, `msg_id`(ULID)와 `idempotency_key`(issue#·stage·attempt) 부여 | 허브 로컬 — fsync된 append 또는 `new/` rename 완료 | 수신 아님 |
| L1 received-by-host | 호스트 런타임이 받았음 | Claude: `-p` stdin write + `--replay-user-messages` 에코, cross-session 결과 "queued"(단 "읽음" 아님, 문서 명시); Codex: `turn/start`·`turn/steer` JSON-RPC 응답; OpenCode: HTTP 200/204, dev `session.inbox.enqueued`; ACP: 해당 없음(턴 종료 응답만); ZCode: 미확인 | 모델이 봤다는 뜻 아님 |
| L2 reflected-in-model-turn | 모델 입력에 실제 포함 | Claude: stream-json `user` 이벤트/transcript의 `<cross-session-message>`; Codex: `item/started{userMessage}`·`turn/started`; OpenCode: `message.updated`(user); ACP: `session/update user_message_chunk` | 작업이 됐다는 뜻 아님 |
| L3 approved-work-executed | 승인된 작업이 실제 수행·검증됨 | **원장만**: PR의 head SHA, 필수 체크 통과, `.moai/reports/<issue>/verdict.md`류 증거 파일, 리더의 재측정 | 어떤 호스트의 메시지·idle 통지·`Stop` 훅도 L3가 아님 |

주의: Claude `notify_when_idle`(검증: main conversation 전용, 로컬 세션만, 12시간 만료)은 "언제 읽을지"의 힌트. Codex `notify`는 `agent-turn-complete`만(승인 요청은 TUI 알림에만 있고 `notify`에는 안 옴). 사다리 상 L1~L2 신호를 Issue 댓글로 미러링하면 리더가 허브 없이도 동일 상태를 복원할 수 있다(원장 우선 원칙).

## 6. 운영 모델 3종 비교

| | A. 기존 오케스트레이터 참고(Orca 1.4.251 / Hide 1.2.2) | **B. 선택적 로컬 허브(1차 모델)** | C. 원격/클라우드(옵션) |
|---|---|---|---|
| 사실 | 직접확인(`--help`만): Orca는 `worktree create`, `agent launch`, `terminal send/read/wait`, `orchestration dispatch/ask/check/reply/gate-create|resolve`, `serve`(헤드리스 런타임), `account select`(OpenCode/Devin). Hide는 `agent spawn --kind claude|codex|opencode… --repo --branch --brief`, 편지(`request send --intent <key>` 재시도 시 중복 전송 없음, kind `request|block|report|progress`), mission 노드에 `--pr --commit --verdict --checks`. 둘 다 공개 프로토콜 문서 **미확인**; Hide 카탈로그는 이 머신에 claude/codex 설치·opencode 미설치로 표시 | 허브는 `moai -l`(또는 선택 UX) 세션이 켤 때만 뜨고, 평시 단일 세션은 켜지 않음; 허브가 각 호스트 네이티브 경로(§3)로만 전달 | Claude `--cloud`/Remote Control(구독 필요, 스크립트 API 미확인), Codex `cloud exec`(실험, 폴링만), 둘 다 비Anthropic/OpenAI 백엔드(GLM) 불가 |
| 설치·운영 부담 | 별도 앱 설치·상시 실행, 계정 연동 | 바이너리 1개, 소켓/파일 1디렉터리, 세션 종료 시 정리 | 계정·요금·네트워크 권한 |
| 비용 | 외부 도구 종속 | 낮음(허브 유휴 비용 0) | 토큰 외 과금·대역 |
| 보안 경계 | 외부 도구의 권한 모델에 위탁(검증 불가) | 사용자 UID 내부, UDS 권한·peer cred로 발신 검증 가능 | 토큰·원격 승인 경계 신설 필요 |
| 결합도 | Orca/Hide 명령 스키마에 결합(`orca agent-context --json` 312 커맨드, 버전 종속) | 호스트 공식 플래그/프로토콜에만 결합 | 벤더 클라우드 API 변동 |
| 결론 | 비교 참고·상호운용 후보(MoAI가 계약/어댑터만 제공) — HOLD | 채택 방향(사용자 결정) | 필요 입증·추가 권한 전제 — HOLD |

## 7. 허브 전송 설계 옵션(모델 B 내부)

| 옵션 | 구성 | 영향 | 선결조건 | 위험 | 미확인 |
|---|---|---|---|---|---|
| H-A 로컬 브로커(UDS/named pipe, JSON Lines) + 호스트 어댑터 | `moai -l`이 `$XDG_RUNTIME_DIR` 또는 `~/.moai-cli/run/<project-key>/hub.sock`(0700) 바인드; 어댑터: Claude=`-p --input-format stream-json` 자식 소유 또는 훅→자기 소켓, Codex=app-server `--listen unix://`/stdio 자식 소유 + `turn/steer`, OpenCode=`serve` HTTP/SSE, ZCode=`--output-format stream-json`(수신 미확인) | 실시간성 최고, 승인 중계 가능(Codex requestApproval, OpenCode permissions, Claude `--permission-prompt-tool`) | Rust `tokio`+`interprocess`; 발신 검증(`SO_PEERCRED`/`getpeereid`/pipe SD); Windows AF_UNIX 보조데이터 없음 → 토큰 파일 인증 | 허브가 세션의 부모가 되면 허브 종료=세션 고아(§10 정리), 대화형 세션은 이 경로로 수신 불가(Claude 대화형엔 stdin 큐 없음) | 외부 프로세스→Claude 타 세션 소켓 투입(미문서) |
| H-B 파일 큐/아웃박스(`~/.moai-cli/...` 또는 `<repo>/.moai-cli/mailbox/`) + 호스트 네이티브 메시징 보조 | Maildir식 `tmp/new/cur`, 세션은 `moai inbox read`(CLI)로 읽고 `moai report`로 씀; 가능하면 Claude cross-session nudge 추가 | 데몬 0, 모든 호스트 공통, 원장(GitHub)과 자연 정합 | `notify` 크레이트로 변경 감지; 세션 측 폴링 규약(훅 `UserPromptSubmit`/`Stop`에서 inbox 확인) | 바쁜 세션 즉시 주입 불가, 폴링 토큰 비용 | Windows rename 의미; ZCode 훅 |
| H-C MCP 서버를 허브로 노출 | 각 호스트가 `moai mcp`를 stdio 서버로 붙임; 도구 `inbox.read`, `report.write`, `claim.*` | 구조화 결과·승인 경계 상속(각 호스트 MCP 도구 승인), 01-C "CLI와 MCP 동일 계약"과 정합 | MCP 스펙 2026-07-28 전환 추적(stateless, `server/discover`) | **푸시 불가**(Claude channels preview 제외), 호스트가 서버를 spawn하므로 세션당 프로세스 1개=허브가 아님 | Codex/OpenCode의 서버 알림 모델 주입 |

세 옵션은 배타적이 아니다. 사실 조합은 "H-B를 내구층, H-A를 선택적 실시간층, H-C를 읽기/보고 도구층"이며, 어떤 조합을 RC1에 넣을지는 HOLD(§15).

## 8. 라이프사이클 옵션

| 옵션 | 동작 | 장점 | 단점 |
|---|---|---|---|
| L1 세션 결합형 | `moai -l`이 허브를 자식으로 띄우고 리더 세션 종료 시 함께 종료(`setsid` 그룹, Windows Job Object `KILL_ON_JOB_CLOSE`) | 고아 없음, 평시 비용 0 | 리더 재시작 시 레인 재접속 필요(재접속 재생은 Codex app-server 문서상 pending 재전송·OpenCode SSE 재구독·Claude는 디스크 큐로 복구) |
| L2 온디맨드 소켓 활성화 | 첫 `moai -a` 또는 `moai -l`이 소켓 없으면 생성, 마지막 참가자 퇴장 후 유휴 N분에 자살 | 단일 세션은 절대 켜지지 않음 | 활성 레이스(동시 2개 기동) → `fd-lock`으로 단일화 |
| L3 상시 데몬 | 사용자 로그인 시 기동 | 가장 단순한 주소 체계 | 사용자 결정과 충돌(비채택) |
| 원격 | 필요 입증 후 Codex `--listen ws://` + `--ws-auth`(비루프백은 "rollout 중 기본 비인증" 문서 경고) | — | 추가 권한·토큰 관리 — HOLD |

macOS에는 `PR_SET_PDEATHSIG` 대응이 없어(미지원) 부모 감시는 `kqueue EVFILT_PROC NOTE_EXIT`로 대체(검증). Claude `--bg` 세션은 터미널 없이 보류 대화상자가 데드라인 넘겨 열려 있음(검증) → 무인 레인은 `crossSessionInbound: accept`를 `--settings`로 주입해야 메시지가 들어온다(검증; Go 세대 `crosssession_settings.go`가 같은 키를 주입했음 — 참고).

## 9. GitHub Issues 원장 경계와 claim/fencing

사실(모두 검증, 일부 직접확인):
- 조건부 쓰기 미지원: "Conditional requests for unsafe methods, such as POST, PUT, PATCH, and DELETE are not supported" → assignee(≤10, PATCH는 집합 교체, push 권한 없으면 조용히 무시)·label·Projects Status는 last-writer-wins.
- 조건부 GET: ETag/`If-None-Match` → 304, 1차 한도 미차감 — **직접확인**(`X-Ratelimit-Used` 443→443).
- 댓글: 서버 부여 int64 `id`, "ordered by ascending ID"; 반응은 사용자별 멱등(재요청 200).
- 의존성: `GET/POST/DELETE /issues/{n}/dependencies/blocked_by`, 관계당 50개; 서브이슈 100개/8단계; 이 저장소에서 `#1987` `blocked_by` → `#1985`, parent `#1976`, timeline에 `blocked_by_added`·`blocking_added` — **직접확인**. "blocked 상태가 close를 막는가"는 미확인.
- PR 자동 close 키워드는 **기본 브랜치 대상 PR만** → `v2026.11.0`로 가는 PR은 `Refs #n`, close는 별도 절차.
- 한도: 사용자 토큰 5,000/h, 콘텐츠 생성 80/min·500/h, 동시 100, 변경 요청 간 1초; `GITHUB_TOKEN`으로 만든 이벤트는 Actions를 재트리거하지 않음.
- `gh auth status`: GoosLab, `gho_` OAuth, scopes `admin:org, delete_repo, gist, repo, workflow`(값 미기록). fine-grained PAT 최소 권한 행은 요약본 출처라 미확인(재확인 조건: docs.github.com/en/rest/authentication/permissions-required-for-fine-grained-personal-access-tokens).

설계 제안(문서 아님, Kleppmann 2016 fencing token 적용):
1. claim = 이슈에 `claim:` 댓글(`session_id`, `host`, `epoch`, `lease_until`, `worktree`, `hub_msg_id`). 동일 이슈의 **살아 있는 claim 중 최소 comment id가 승자**; 패자는 즉시 철회 댓글.
2. 모든 후속 산출물(브랜치명, 커밋 트레일러, PR 본문 `Refs #n claim:<comment_id>`, 보고 댓글)이 토큰을 운반; 리더는 토큰 불일치 산출물을 무시(GitHub은 거부해 주지 않으므로 **읽기 측 검증이 집행점**).
3. 리스 갱신은 claim 댓글 **편집이 아니라 추가 댓글**(편집 이력은 timeline에 남지 않음) 또는 `reactions`로 heartbeat; 만료 뒤 takeover는 자동이 아니라 운영자 승인(06-B HOLD와 정합).
4. 의존 게이트는 native `blocked_by`를 읽고, 해제는 선행 이슈 close + PR merge SHA를 함께 확인.
5. 결과 영수증: `report:` 댓글 1건/단계, idempotency_key = `issue#:stage:attempt`, 재전송 시 동일 키 댓글이 있으면 생략(80/min 한도 보호).
6. 중복·재정렬: 허브 `msg_id`(ULID)와 댓글 id 둘을 함께 기록, 소비자는 둘 중 하나로 중복 제거; 이벤트 순서는 timeline id 순을 진실로.
7. 연결 끊김/재시작: 세션은 재시작 후 자기 claim 토큰이 여전히 최소 생존인지 재조회(06-B "중단 재조회"), 아니면 handoff 댓글 후 종료; 미커밋 변경은 worktree에 남기고 브랜치 push 여부를 보고.
8. 감사 추적: 댓글·timeline·PR 이벤트가 전부이며 허브 로그는 보조. 허브 메시지만으로는 소유·승인·완료를 **절대** 확정하지 않음.

## 10. 레인 루프 계약 — 호스트별 매핑

루프: claim → worktree 진입 → 구현 → 로컬 build/test → PR(base `v2026.11.0`) → worktree 이탈 → 다음 이슈.

| 단계 | Claude Code | Codex | OpenCode | ZCode | ACP 경로 |
|---|---|---|---|---|---|
| 기동·worktree | 검증 `claude -w <name>`(`.claude/worktrees/<name>`, `worktree-<name>`) 또는 런처가 `git worktree add` 후 `cwd`; `--add-dir`; `.worktreeinclude` | 검증 `codex exec --worktree`(위치 미문서) 또는 런처 worktree + `-C`; app-server `thread/start{cwd}`; **workspace-write에서 `.git` 보호 → commit에 승인** | 검증 `run`은 `$PWD`; dev `--worktree`; 세션 `directory` | 검증 `--cwd`(저장소 문서) | 검증 `session/new{cwd 절대경로}` |
| 이슈 본문 전달 | `-p` 첫 메시지/`SessionStart.initialUserMessage` 훅; `--append-system-prompt`로 계약 삽입 | `turn/start.input` 또는 exec PROMPT(stdin) | `POST /session/:id/message` | `--prompt` | `session/prompt` |
| 진행 보고 | 훅 `Stop`/`PostToolUse`가 `moai report` 호출; cross-session nudge(선택) | hooks `Stop`/`PostToolUse`, `notify`, app-server 이벤트 | SSE 구독(허브) 또는 CLI 보고 | stream-json 파싱(허브) | `session/update` |
| 승인 경계 | `--permission-mode` + `--permission-prompt-tool`(허브 MCP) 또는 `PermissionRequest` 훅; `bypassPermissions` 금지 권고 | app-server `requestApproval` 응답은 허브가 아니라 **운영자 결정**을 중계; exec는 `approval_policy=never`+sandbox로 사전 결정, `declined` 수용 | `permissions[]` 정책 + `permissions/:id` 응답 | `--mode plan` 우선, yolo는 운영자 승인 | `request_permission` 클라이언트 정책 |
| PR | `gh pr create -B v2026.11.0 -H <branch> --draft`, `Refs #n claim:<id>`, `gh pr ready`, `gh pr merge --auto`(저장소 설정 필요) — 동일 | 동일 | 동일 | 동일 | 동일 |
| 이탈·다음 | `ExitWorktree`(검증) → 다음 claim; 컨텍스트 미승계 원칙 | `thread/archive` 또는 새 thread | 새 session | 새 세션 | `session/close`(게이트) |
| 중단 | SIGINT(턴 종료)·`interrupt()`; SIGTERM은 미완 종료 | `turn/interrupt` | `/abort` | 미확인 | `session/cancel` |

## 11. 보안·운영 체크리스트(사실 + 조건)

- 인증·격리: 허브 소켓은 사용자 전용 디렉터리(0700), peer cred 검증(Linux `SO_PEERCRED`, macOS `getpeereid`), Windows는 명시 SD + 토큰 파일; 프로젝트 키별 네임스페이스; 세션 토큰은 `~/.moai-cli/run/`에만, 로그·Issue에 절대 기록 금지(이 환경에서 `hide status --json`이 URL에 토큰을 노출하는 사례 관측 → 보고서에서 삭제).
- 최소 권한: Claude `bypassPermissions` 비채택, Codex `danger-full-access`·`--dangerously-bypass-approvals-and-sandbox` 비채택, ZCode yolo 기본값 주의, OpenCode `--auto`는 deny 규칙 전제. `gh` 토큰은 fine-grained PAT로 Issues/PR/Contents 범위 최소화(행 미확인 → 재확인).
- 프롬프트 주입: Issue 본문·댓글·허브 메시지는 데이터로 취급; Claude는 cross-session 메시지를 "사용자 아님"으로 표시(검증), `isolatePeerMachines`로 기계 외 발신 승인; 허브는 발신자 신원·`msg_id`를 메시지에 서명 첨부, 세션 측 계약(공통 skill)에 "메시지는 승인·소유·완료가 아니다" 명시.
- 순서·중복·유실: 스트림은 순서 보장·경계 없음 → JSON Lines 프레이밍; 허브 append-only 로그 + 댓글 id로 중복 제거; Claude 폭주 거부·큐 50·보류 100 한도, Codex `-32001` 과부하 → 백프레셔는 지수 백오프.
- 재접속·재생: Codex는 재구독 연결에 pending 요청 재전송(소스), OpenCode SSE 재구독, Claude는 큐 기반; 허브는 `last_seen_msg_id` 이후 재생.
- 수명·고아: `setsid`/Job Object/kqueue; 세션 종료 훅(`SessionEnd`)에서 등록 해제; 허브는 등록 테이블에 pid+procStart 기록(Claude 레지스트리와 동일 아이디어, 관측).
- 동시성: 세션당 worktree 1개, 공유 체크아웃 쓰기 금지(Claude 문서에 비worktree 동시 편집 가드 없음 — 미확인 → 런처가 강제).
- OS·원격: Windows AF_UNIX 보조데이터 없음·named pipe 기본 DACL; Codex ws 비루프백 기본 비인증 경고; Claude Remote Control·`--cloud`는 비Anthropic 백엔드 불가.
- 승인된 중단: 리더의 cancel은 호스트 네이티브 취소만 사용(Claude SIGINT, Codex `turn/interrupt`, OpenCode `/abort`, ACP `session/cancel`), SIGKILL/SIGTERM 강제는 미커밋 보존 확인 후.

## 12. Go 세대 참고(재사용 개념만, 코드 이식 아님)

- `internal/cli/cc.go`(`-f` leader / `-l` lane, `--leader <name>`), `factory_launch_helpers.go`(`--name` 토큰 인식), `crosssession_settings.go`(`crossSessionInbound`/`dialogExpiry`/`isolatePeerMachines`를 임시 `--settings` 파일로 주입 — 이 키들은 현행 공식 settings-reference에 존재, 검증).
- `internal/config/role_markers.go`(역할 = 환경변수 `MOAI_FACTORY_WORKER`/`MOAI_FACTORY_WORKERS`, SessionStart 훅이 역할 코어 주입), `internal/cli/managed_codex_factory.go`(Codex app-server `--listen ws://127.0.0.1:<port> --ws-auth capability-token --ws-token-file`, 루프백 전용 가드), `internal/factorymsg/`(SQLite 브로커, claim 토큰, handoff rebind), `internal/cli/factory_messaging.go`(`probe/request/ack/fallback` — 채널 가용성 판정을 명령 실패가 아닌 진단으로).
- 규칙: `.claude/rules/moai/workflow/factory-dispatch*.md`(디스패치 블록 `card/spec/cmd/wt/evidence/lens`, 큐가 위임 채널, 완료는 읽고 신뢰하지 않음, 리더 잔여 역할 표), `cross-session-messaging*.md`(전송 결과 3형, idle notice 힌트, stopped-teammate 금지, 공유 플래그 슬롯 — 미문서), `auto-semantics.md §6`(결정 사다리).

## 13. 논문·베스트 프랙티스(채택/비채택)

| 출처 | 채택 | 이유·검증 조건 |
|---|---|---|
| Kleppmann, "How to do distributed locking"(2016) martin.kleppmann.com | 채택 | fencing token → claim 댓글 id; 조건: 리더 읽기 측 검증 구현 |
| Gray & Cheriton, "Leases"(SOSP 1989) dl.acm.org/doi/10.1145/74850.74870 | 참고 | 리스 만료 모델; ACM은 봇 403, 미러 PDF는 스캔본 → 본문 미확인 |
| Stripe, "Designing robust and predictable APIs with idempotency"(2017) stripe.com/blog/idempotency | 채택 | 보고 댓글·허브 메시지 idempotency key |
| Richardson, Transactional outbox microservices.io | 채택 | 진행 기록과 발신 메시지를 한 로컬 커밋으로 → H-B 파일 큐 |
| GitHub REST best practices docs.github.com/en/rest/using-the-rest-api/best-practices-for-using-the-rest-api | 채택 | 조건부 GET·직렬 변경·`retry-after` |
| MCP 2026-07-28 changelog modelcontextprotocol.io/specification/latest/changelog | 참고 | H-C 선택 시 stateless 모델 추적 |
| ACP spec agentclientprotocol.com | 참고 | 공통 계약 후보; 세션당 1턴 한계 |
| Anthropic, "Building effective agents"(2024), "How we built our multi-agent research system"(2025) anthropic.com/engineering | 채택 | 오케스트레이터-워커, 명시 브리프, 체크포인트 |
| AutoGen(arXiv 2308.08155), MetaGPT(2308.00352), ChatDev(2307.07924), Guo et al. 조사(2402.01680), Masterman et al.(2404.11584) | 참고 | 구조화 산출물을 전달 매체로(MetaGPT), 미검증 주장 비용(ChatDev); 직접 설계 입력은 아님 |
| IRC RFC 1459/2812, IRCv3 msgid/echo-message/chathistory | 비채택 | 영수증·영속성 부재, 상시 서버 전제 |
| JSON Lines jsonlines.org, LSP base protocol, POSIX rename, Maildir cr.yp.to/proto/maildir.html | 채택 | 프레이밍·무잠금 메일박스 |
| "Agents orchestrating agents" | 제외 | 정규 논문 미발견(미확인) |

## 14. RC Issue 매핑

| Issue | 본 조사 결과 반영 | HOLD(Issue 기재) |
|---|---|---|
| 00-D #1987 protocol·상태 의미 | 메시지 봉투(`msg_id`, `idempotency_key`, `kind: request|progress|blocked|report`, 발신자 신원), 상태 어휘 `unclaimed/claimed/in_progress/blocked/pr_open/merged/abandoned`, 영수증 사다리 L0~L3를 JSON Schema fixture 후보로 | P06 typed schema/version·MCP parity HOLD(유지) |
| 01-C #1990 CLI·MCP 동일 계약 | H-C는 푸시 불가·spawn 수명 제약이므로 "읽기/보고 도구" 패리티만 가능; no-server 기본과 정합 | P05 bridge, P06 HOLD |
| 02-B #1992 Claude adapter | `-p` stream-json 큐·`--permission-prompt-tool`·훅 `additionalContext`·cross-session 설정 키 3종·`--bare` 비호환(소켓 없음) | — |
| 02-C #1993 Codex Desktop adapter | app-server가 유일 양방향 면; Desktop 외부 주소 미확인; `.git` 보호 경로로 commit 승인 필요; `codex mcp-server` 제거 | — |
| 02-D #1994 ZCode 지원표 | ZCode는 실존 공식 제품(docs.z.ai/devpack/tool/zcode, zai-org/ZCode v3.17.2), 헤드리스 플래그 검증, 바이너리 배포 채널·수신 메시지·훅 미확인; Claude Code via `api.z.ai/api/anthropic`, Codex via `api.z.ai/api/v1`(responses) 검증 | P15 adapter 순서 HOLD |
| 06-A #2005 단일 리더·claim | 댓글 id fencing(설계 제안), CAS 부재 사실, 중복 claim 거부는 읽기 측 | P08 claim primitive·lease HOLD |
| 06-B #2006 중단 재조회·handoff | 호스트별 중단 수단, Codex pending 재전송, 자동 takeover 미승인 유지 | P08 HOLD |
| 06-C #2007 worktree·PR | 호스트별 worktree 수단, `Refs #n`(비기본 브랜치), `--auto` merge 조건 | — |
| 07-A #2008 GitHub polling adapter | 조건부 GET 304 직접확인, 한도, `GITHUB_TOKEN` 재트리거 불가, Actions `issues/issue_comment` 반응자 | — |
| 07-C #2010 간결 보고·artifact | 보고 댓글 멱등 키, 로그는 필요 시만, 토큰 노출 금지 | P09 artifact retention HOLD |
| 07-D #2011 native 판단·context | 허브 메시지는 판단 입력이 아님; 완료 판정은 원장 증거만 | P10 HOLD |

## 15. HOLD(사용자 결정 필요)

1. 허브 UX/문법(`moai cc` / `moai -l` / 단독 `moai`의 호스트 선택) — 미결.
2. 허브 전송 조합(H-A/H-B/H-C) 및 RC1 포함 범위.
3. 허브 수명 옵션(L1 세션 결합 vs L2 온디맨드) 및 소켓/큐 경로 표준.
4. claim 규약 채택 여부(댓글 id fencing)와 리스 길이·갱신 수단(댓글 vs reaction).
5. 승인 중계 범위: 허브가 Codex/OpenCode 승인 요청을 운영자에게 중계만 할지, 정책 사전 결정(`approval_policy=never`+sandbox)으로 갈지.
6. ACP를 선택 어댑터로 추가할지(벤더 공식성 미확인, 1턴 제약).
7. ZCode adapter 범위(헤드리스만 vs 수신 경로 조사 후).
8. 원격(Codex `--listen ws://`, Claude Remote Control/`--cloud`) 도입 — 필요 입증 전 보류.
9. Orca/Hide 상호운용(계약/어댑터 제공) 여부.
10. 신규 구현 전부(이 보고서는 조사 결과이며 어떤 구현도 승인하지 않음).

## 16. 미확인 목록과 검증 조건

- Claude stdin 제어 메시지 와이어 스키마; 외부 프로세스의 타 세션 소켓 투입; 승인 프롬프트 대기 중 cross-session 메시지 처리; `CLAUDE_CODE_SESSION_NAME` 효과; transcript JSONL 스키마 → 조건: 공식 문서 추가 또는 격리 환경 실측(쓰기 발생, 승인 필요).
- Codex: 바쁜 thread의 일반 `turn/start` 큐/즉시 여부; exec 다중 프로세스 동일 repo; `--worktree` 위치; 승인 타임아웃; 범용 MCP 알림 주입 → 조건: 격리 repo에서 app-server 실측.
- ZCode: 배포 채널, 훅/MCP, 수신 메시지, 승인 모델 → 조건: `apps/zcode-cli/docs/usage/09-hooks.md`·`10-plugins.md` 열람 + 설치 실측.
- OpenCode: 릴리스 v1.19.5와 dev 문서 간 플래그 차이, `remember` 필드, Rust SDK 부재 → 조건: `opencode --help`·`/doc` 실측.
- ACP: v2 안정성, Claude/Codex 어댑터의 벤더 소유, 원격 전송.
- GitHub: fine-grained PAT 행, 의존성 요금제 제한, blocked 이슈 close 제한, 웹훅 순서.
- Orca/Hide 내부 프로토콜(도움말 외 미확인).

## 17. 주요 출처(2026-11-08 확인)

Claude Code: code.claude.com/docs/en/{cli-reference, headless, agent-sdk/typescript, agent-sdk/python, agent-sdk/streaming-vs-single-mode, cross-session-messaging, agent-teams, remote-control, hooks, worktrees, mcp, channels-reference, settings-reference, claude-code-on-the-web, env-vars}.
Codex: developers.openai.com/codex/{noninteractive, app-server, cli/reference, config-reference, config-advanced, sandbox, hooks, mcp, skills, guides/agents-md, environments/git-worktrees}, learn.chatgpt.com/docs/mcp-server, github.com/openai/codex(codex-rs/app-server, app-server-protocol, exec/src/exec_events.rs).
Z.ai: docs.z.ai/devpack/tool/{zcode, claude, codex, others}, docs.z.ai/devpack/overview, zcode.z.ai/en/docs/configuration, github.com/zai-org/ZCode.
OpenCode: docs.opencode.ai/docs/{server, sdk, acp, config, permissions, agents, mcp-servers, rules}, github.com/anomalyco/opencode(v1.19.5, dev).
ACP: agentclientprotocol.com/protocol/{overview, initialization, session-setup, prompt-turn, schema, extensibility, transports}, docs.rs/agent-client-protocol(5.0.0), github.com/agentclientprotocol/{claude-agent-acp, codex-acp, rust-sdk}.
MCP/전송: modelcontextprotocol.io/specification/{2025-06-18, 2025-11-25, latest}/changelog, jsonrpc.org/specification, microsoft.github.io/language-server-protocol/specifications/lsp/3.17/specification/#baseProtocol, jsonlines.org, man7.org/linux/man-pages/man7/unix.7.html, developer.apple.com getpeereid(3)/kqueue(2), learn.microsoft.com named-pipe-security-and-access-rights, devblogs.microsoft.com/commandline/af_unix-comes-to-windows, pubs.opengroup.org rename()/setsid()/socket(), cr.yp.to/proto/maildir.html, rfc-editor.org/rfc/rfc1459·rfc2812, ircv3.net/specs/extensions/{message-tags, message-ids, echo-message, server-time, chathistory}.
GitHub: docs.github.com/en/rest/issues/{issue-dependencies, sub-issues, assignees, issues, comments, timeline}, docs.github.com/en/rest/using-the-rest-api/{best-practices-for-using-the-rest-api, rate-limits-for-the-rest-api, issue-event-types}, docs.github.com/en/graphql/overview/rate-limits-and-node-limits-for-the-graphql-api, docs.github.com/en/issues/tracking-your-work-with-issues/using-issues/{creating-issue-dependencies, adding-sub-issues, linking-a-pull-request-to-an-issue}, docs.github.com/en/actions/reference/workflows-and-actions/events-that-trigger-workflows, github.blog/changelog/{2025-08-21-dependencies-on-issues, 2025-04-09-evolving-github-issues-and-projects, 2025-09-11-…}, cli.github.com/manual/{gh_api, gh_issue_edit, gh_pr_create, gh_pr_merge, gh_pr_ready, gh_issue_lock}.
논문·글: §13 표 참조.

---

Process notes for the coordinator: no files were created or modified in either repository (scratchpad-only harvest files under `/private/tmp/claude-501/-Users-goos-MoAI-moai-adk/be5223b1-1507-455d-b8ef-0cd5d37f4e62/scratchpad/`); no git/GitHub state changes; one observed secret (Hide local URL token) was deliberately omitted. The five research agents may still emit follow-up messages to `team-lead`; they contain nothing beyond what is integrated above.
