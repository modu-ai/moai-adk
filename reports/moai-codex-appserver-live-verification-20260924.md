# Claude Code → Codex CLI 자동 전달 실세션 검증

**판정:** 제한된 macOS 실험에서 **가능함(PASS)**. MoAI MCP 사서함을 폴링하는 임시 브리지가 Codex App Server의 `turn/start`로 전달하자, 이미 열려 있던 Codex CLI 원격 TUI 화면과 저장된 턴 기록에 메시지가 나타났다. 작업 중 `turn/steer`도 별도로 통과했다. **제품 구현·운영 지원 판정은 아직 아님.**  
**기준선:** 2026-09-24, 저장소 `main` / `2213871af`, 설치 `moai-adk v3.2.0-rc.13` (`g60017eb83`), `codex-cli 0.156.1`, Claude Code `2.1.281`, macOS. 임시 작업 루트 `/private/tmp/moai-appserver-live-20260924`.  
**연결 관계:** [기능 재정의 보고서](moai-codex-broker-feature-reset-20260923.html)의 “현재 제품에 자동 push 경로 없음” 판정은 유지된다. 이 보고서는 **저장소 밖 임시 브리지**로 그 경로의 기술적 성립을 시험한 후속 증거다.

## 1. Claim: 어느 단계까지 실제로 되었나

```text
실제 Claude Code(GLM 백엔드)
  → mcp__moai__session_msg_register / session_msg_send
  → 공유 .moai/state/session-msg 사서함
  → 임시 브리지의 session_msg_poll
  → Codex App Server turn/start
  → 실행 중인 codex --remote TUI에 표시
  → thread/read에서 완료된 최종 답변 확인
```

App Server와 TUI의 직접 시험에서는 다음 두 상태가 각각 통과했다.

| 수신 Codex 상태 | 브리지 측 API | 실제 관측 |
|---|---|---|
| 대기 중 | `turn/start(threadId, input)` | 다른 WebSocket 클라이언트가 시작한 턴의 `TUI_PUSH_VISIBLE_OK`가 TUI에 표시되고 `thread/read`에서 `completed`로 확인됨 |
| 작업 중 | `turn/steer(threadId, expectedTurnId, input)` | API가 같은 turnId를 반환했고 TUI 최종 답변 `TUI_BASE_DONE TUI_STEER_VISIBLE_OK` 확인 |

이번 시험의 **실제 Claude Code 발신 → TUI 수신**은 대기 중 `turn/start` 경로다. 실제 Claude Code 발신 메시지로 활성 턴의 `turn/steer`까지 연결한 왕복은 별도 검증이 필요하다.

## 2. Evidence: 명령과 축약하지 않은 판정 출력

아래는 이번 실행에서 관측한 판정용 출력이다. 장식적인 TUI 제어 문자열과 모델의 중간 스트리밍 조각은 제외하고, 최종 화면 문구는 그대로 옮겼다. 인증 정보는 기록하지 않았다.

```text
git rev-parse --short HEAD
2213871af

codex --version
codex-cli 0.156.1

claude --version
2.1.281 (Claude Code)

moai version
moai-adk v3.2.0-rc.13
[v3.2.0-rc.13] [moai_cp/20260910_130400-2916-g60017eb83] [built 2026-09-23T23:56:01Z]

codex app-server --listen ws://127.0.0.1:61026
codex app-server (WebSockets)
  listening on: ws://127.0.0.1:61026
  readyz: http://127.0.0.1:61026/readyz
  healthz: http://127.0.0.1:61026/healthz
  note: binds localhost only (use SSH port-forwarding for remote access)
```

격리된 스레드 `01a0d0b9-e89c-7e01-8d3d-355da719af07`의 직접 API 시험:

```text
IDLE_TURN_ID 01a0d0b9-e9de-7a22-99b9-26766a08ff3c
item/completed.agentMessage.text = IDLE_DELIVERY_OK
turn/completed.status = completed

ACTIVE_TURN_ID 01a0d0b9-fc14-7500-8057-b90d1687fcac
STEER_ACCEPTED {"turnId": "01a0d0b9-fc14-7500-8057-b90d1687fcac"}
item/completed.agentMessage.text = "BASE_DONE\nSTEER_DELIVERY_OK"
turn/completed.status = completed
```

실행 중인 원격 CLI는 `codex --remote ws://127.0.0.1:61026 -C /private/tmp/moai-appserver-live-20260924 --no-alt-screen -a never -s read-only`로 연결했다. 별도 WebSocket 클라이언트가 대상 스레드 `01a0d0bc-c7ae-7321-909e-7f01ae9ddca8`에 요청했고, TUI 화면과 `thread/read`가 같은 결과를 보였다.

```text
turn/start → turn.id 01a0d0be-25b6-74a3-887e-c83b3250ebbb
TUI 최종 화면: TUI_PUSH_VISIBLE_OK
thread/read: status=completed, messages=["TUI_PUSH_VISIBLE_OK"]

turn/start → turn.id 01a0d0bf-9b33-7382-91e1-54e893dbe185
turn/steer → {"turnId":"01a0d0bf-9b33-7382-91e1-54e893dbe185"}
TUI 최종 화면: TUI_BASE_DONE TUI_STEER_VISIBLE_OK
thread/read: status=completed, messages=["구스~오뽜, 8초 기다리는 명령 실행할게.","TUI_BASE_DONE TUI_STEER_VISIBLE_OK"]
```

실제 Claude Code 발신 시험은 `moai glm -- --print ... --mcp-config <임시 설정> --strict-mcp-config --allowedTools <등록·송신 도구> --permission-mode dontAsk`로 실행했다. Claude Code가 GLM 백엔드에서 반환한 메시지 ID와 브리지의 수신·확인, Codex 저장 기록이 일치했다.

```text
Claude Code 결과: 메시지 전송 완료. messageId: msg-1e9dd6da8f7a7ed8
BRIDGE_AGENT_ID codex-ae035d78
BROKER_RECEIVED {"messageId":"msg-1e9dd6da8f7a7ed8","senderKind":"claude","text":"GLM_CLAUDE_TO_CODEX_LIVE_20260924"}
PUSH_RESPONSE turn.id = 01a0d104-2540-7681-9266-43c623aecc99
ACKED_COUNT 1
TUI 최종 화면: CROSS_HOST_DELIVERED_OK GLM_CLAUDE_TO_CODEX_LIVE_20260924
thread/read: status=completed, messages=["CROSS_HOST_DELIVERED_OK GLM_CLAUDE_TO_CODEX_LIVE_20260924"]
```

## 3. Baseline-attribution: 무엇을 시험한 것인가

- 저장소 코어는 현재 HEAD `2213871af`였지만, 호출된 설치 `moai`는 더 뒤의 빌드 `g60017eb83`이었다. 이 시험은 **설치 바이너리의 MCP 동작** 근거다. 현재 HEAD 바이너리와의 동일성을 주장하지 않는다.
- 브리지 파일 `/private/tmp/moai-appserver-bridge-20260924.py`, Claude MCP 설정과 수신 프로젝트는 `/private/tmp`에 격리했다. 저장소 제품 코드에는 브리지를 추가하지 않았다.
- 두 WebSocket 클라이언트와 Codex TUI가 **같은 로컬 App Server**를 사용했다. 이것이 기존 독립 `codex` 프로세스의 임의 스레드로 push할 수 있다는 뜻은 아니다.
- [Codex App Server 공식 문서](https://learn.chatgpt.com/docs/app-server)는 `turn/start`, 활성 턴의 `turn/steer`, `codex --remote` 연결을 설명한다. 원격 WebSocket/App Server 표면은 문서에서 실험적이고 운영용으로 미지원이라고 표시한다.

## 4. Gaps: 통과로 확장할 수 없는 범위

1. **Anthropic 본계정 미실행:** `claude -p`는 첫 API 요청에서 `429` 주간 한도에 걸렸고 도구 호출 없이 끝났다. 실제 발신 성공은 **Claude Code 클라이언트 + GLM 백엔드** 조합이다.
2. **동시 상시 브리지 지연 미측정:** 첫 브리지는 메시지 송신 전에 제한된 폴링 창이 끝났다(`BROKER_TIMEOUT`). Claude가 보낸 메시지는 사서함에 남았고, 브리지를 다시 실행하자 자동으로 poll·App Server 전달이 이뤄졌다. 따라서 보관·재시작 전달은 보였지만 상시 실행 중 도착 지연은 측정하지 않았다.
3. **단방향:** Codex 결과를 Claude Code의 살아 있는 턴으로 다시 자동 전달하는 역방향은 시험하지 않았다.
4. **플랫폼:** macOS 로컬 WebSocket·공유 파일 시스템만 시험했다. Linux·Windows·WSL2와 다른 머신 간 전달은 미검증이다.
5. **운영:** 재연결, 다중 발신·중복, 충돌, 실제 승인 요청, 네트워크 단절, 서버 재시작, 버전 변경, 장시간 운용은 미검증이다.

## 5. Residual-risk와 진행 판정

**진행 권고: 제한된 구현 실험은 진행해도 된다. 기본 제품 기능·Factory·운영 지원으로 승격하는 것은 보류한다.** 지금의 실증은 “Codex App Server가 관리하는 원격 TUI에 메시지를 자동 표시하는 경로가 성립한다”는 기술적 가능성을 충분히 뒷받침한다.

제품 코드로 옮길 때는 다음 순서로 검증해야 한다.

| 순서 | 필수 보강 | 완료 게이트 |
|---|---|---|
| 1 | MoAI가 소유하는 App Server·TUI 연결 수명과 정확한 thread/turn 식별 | 잘못된 사용자 스레드·프로젝트에 전달 0건 |
| 2 | Claude/GLM 메시지를 **동료 출처의 데이터**로 라벨링하고 사용자 지시·승인으로 승격하지 않는 입력 경계 | 악의적 메시지로 파일 변경·권한 상승·승인 통과 0건 |
| 3 | `turn/start` 수락, `turn/completed`, TUI 표시, 브로커 ack를 별도 상태로 저장 | 실패·재시작·중복에서 메시지 손실과 중복 작업 없음 |
| 4 | `turn/steer`의 활성 턴 ID 경합·대기 턴 전환·승인 중단 처리 | 실제 Claude Code 발신으로 두 상태 모두 종단 PASS |
| 5 | 현재 대상 Codex 버전과 지원 OS별 회귀 시험 | macOS·Linux·Windows 지원 범위별 판정, 실험적 App Server 정책 명시 |

**이번 임시 브리지의 한계:** `turn/start`가 수락되자마자 사서함을 ack했다. 그 뒤 Codex 턴이 실패하면 메시지가 유실될 수 있으므로 제품 구현에서는 이 순서를 그대로 쓰면 안 된다. 또한 브리지는 동료 메시지를 일반 턴의 사용자 입력으로 보냈다. 시험 문구만 사용했으므로 결과는 유효하지만, 제품에서는 출처·권한 분리 후 입력 경로를 정해야 한다.

**정리:** MoAI MCP 파일 브로커만으로는 자동 push가 되지 않는다. **MoAI 브리지 + Codex App Server + 같은 서버에 연결된 Codex CLI** 조합은 이 실험에서 실제 화면 표시까지 성공했다. 구현 여부 결정은 이 두 문장을 분리한 상태에서 내려야 한다.
