# MoAI for Codex 기능 재정의: 파일 브로커와 실제 push의 경계

**기준선:** 2026-09-23 · 저장소 `main` / `2213871af` · 설치 Codex CLI `0.156.1`  
**판정 대상:** `session_msg_register/list/send/poll`, Claude Code→Codex CLI 전달, Codex판 MoAI의 필수·선택 기능.  
**이전 보고서 정정:** [Codex 플랫폼 재설계 보고서](moai-codex-platform-redesign-20260923.html)의 “이번 조사에서 실제 양쪽 호스트 세션 간 전달은 시험하지 않았다”는 이번 조사 자체에는 맞지만, 저장소에는 2026-08-23의 관련 실프로세스 시험 기록이 이미 있었다. 그 시험은 **Claude Code 호스트가 아니라 `kind=claude`로 등록한 MCP 테스트 드라이버**와 `codex exec`의 왕복이다. 따라서 “실제 Claude Code→살아 있는 Codex CLI로 자동 push 성공”의 근거는 아니다.

> **2026-09-24 후속 검증:** [실세션 검증 보고서](moai-codex-appserver-live-verification-20260924.html)에서 저장소 밖 임시 브리지와 Codex App Server를 연결해 **Claude Code(GLM 백엔드) → MoAI MCP 사서함 → Codex 원격 TUI 자동 표시**를 관측했다. 아래의 “현재 브로커만으로는 push 불가” 판정과 양립한다. 자동 연결은 아직 MoAI 제품 코드에 없다.

## 1. 결론

**구스님의 의심은 제품 의미에서 맞다.** 현재 MCP 브로커는 메시지 파일을 상대 사서함에 **적재**할 수 있다. Codex 세션이 `session_msg_poll`을 **스스로 호출하면** 읽을 수도 있다. 그러나 `session_msg_send`만으로 이미 실행 중인 Codex CLI의 턴이 깨어나거나, 화면에 메시지가 자동 표시되거나, 모델 입력에 삽입되는 경로는 현재 코드에서 확인되지 않았다. 이를 “Claude Code→Codex CLI push”라고 부르면 오해를 만든다.

```text
Claude측 MCP 호출 → session_msg_send → 공유 디스크 pending/*.json
                                                │
                                     Codex가 명시적으로 poll 호출
                                                ↓
                                      Codex 도구 응답으로 수신

send 성공 ≠ Codex에 자동 알림 ≠ Codex 턴 재개 ≠ 업무 완료
```

**MoAI for Codex의 기본 기능은 이 브로커 없이 완주해야 한다.** 기본 실행은 Codex의 지침·스킬·서브에이전트·훅·MCP를 통한 단일 소유 세션의 plan→run→sync와 증거 판정으로 정의한다. 브로커는 같은 프로젝트 디스크를 공유하는 두 실행자가 직접 등록·폴링하기로 합의한 경우의 **선택적 사서함**으로만 남긴다. 자동 도착이 필요한 제품 기능은 별도 호스트 제어 어댑터와 종단 검증이 생기기 전까지 지원하지 않는다.

## 2. 이번 조사에서 직접 관측한 근거

| 주장 | 현재 트리의 기계적 근거 | 판정 범위 |
|---|---|---|
| `send`는 파일 적재 | `internal/sessionmsg/store.go:205-286`의 `Send`가 수신자 `pending/<messageId>.json`에 `writeJSONAtomic` 수행 후 ID를 반환 | 파일 쓰기 성공. 수신 호스트의 실행·표시는 확인하지 않음 |
| 수신은 `poll` 요청 필요 | `internal/sessionmsg/store.go:288-390`, `internal/cli/mcp_session_msg.go:126-144`가 명시 호출에서 pending→claimed 및 반환 수행 | 폴링하지 않는 세션의 자동 수신 경로가 아님 |
| MCP 등록은 호출 가능성 | `internal/cli/mcp_server.go:439-481`, `.codex/config.toml:1-4` | 도구 노출은 지속 수신 루프나 알림 구독을 만들지 않음 |
| 자동 폴링 연결 없음 | `rg`로 `.claude/skills`, `.agents/skills`, `.codex`, `.claude/hooks`, `.codex/hooks`, `internal/codexwiring`의 `session_msg_*` 호출 검색: **0건**. `internal/cli`에서도 핸들러·등록·테스트 외 호출 **0건** | 현재 조사 범위의 호출자가 없다. 외부 사용자 스크립트까지 부재하다는 뜻은 아님 |
| 교차 실프로세스 기록의 정체 | `.moai/specs/SPEC-CODEX-SESSION-MSG-001/progress.md:126-168`: Claude측은 stdio JSON-RPC `drive_mcp.py` 테스트 드라이버, Codex측은 `codex exec` | 공유 사서함의 명시적 poll 왕복을 과거에 보였음. 실제 Claude Code 호스트나 비동기 push 시험은 아님 |
| 프로젝트 루트 일치 필요 | `internal/cli/mcp_session_msg.go:29-56`는 `CLAUDE_PROJECT_DIR` 또는 CWD로 루트 선택. 위 progress의 첫 시도는 루트 불일치로 `LISTCOUNT: 1`, 환경을 고정한 뒤 `2` | 같은 저장소처럼 보여도 MCP 자식 프로세스의 루트가 다르면 전달되지 않음 |

### 현재 트리에서 실행한 좁은 테스트

```text
git rev-parse --short HEAD
2213871af

go test ./internal/sessionmsg -count=1 -timeout 90s
ok  github.com/modu-ai/moai-adk/internal/sessionmsg  0.356s

go test ./internal/cli -run 'Test(SessionMsg|MoaiMCPServer_RegistrationMatchesCatalog)' -count=1 -timeout 90s
ok  github.com/modu-ai/moai-adk/internal/cli  0.548s
```

이 테스트는 저장·클레임·ack와 MCP 핸들러를 확인한다. **현재 설치된 Claude Code와 Codex CLI의 두 실세션에서 push를 재현한 시험이 아니다.** 과거 progress의 승인 우회 조건도 현 제품의 정상 승인 경로 성공으로 옮겨 적을 수 없다.

## 3. 왜 브로커가 push가 아닌가

| 필요한 단계 | 현재 브로커 | 제품상 필요한 보장 |
|---|---|---|
| 송신 | 등록된 ID를 검사하고 파일을 적재 | 전달 시도와 저장 성공 구분 |
| 수신자 실행 상태 | `lastHeartbeat`의 최근성으로 `online` 표시 | 실제 Codex 프로세스·스레드 연결 및 수신 가능 상태 |
| 자동 알림 | 구현 확인 안 됨 | 새 메시지 발생 시 호스트에 이벤트 전달 |
| 턴 입력 | `poll`의 도구 응답으로만 들어옴 | 수신자 턴에 입력을 안전하게 추가하거나 새 턴 시작 |
| 확인 | `poll(ack_ids)`로 파일 삭제 | 수신·처리·사용자에게 표시·업무 반영을 각각 구분 |
| 권한 | 공유 디스크에 접근 가능한 등록 ID 기반 | 발신자 인증, 수신자 소유권, 사용자 승인과 작업 권한 경계 |

`kind=claude`는 등록자가 제공한 문자열이다(`internal/sessionmsg/agent.go:76-115`). 이것만으로 호출 주체가 Claude Code인지 증명하지 않는다. `session_msg_list`의 `online`도 하트비트 나이 기준(`internal/config/defaults.go:403-420`)이므로 사용자 앞에서 Codex 세션이 읽고 있다는 보장이 아니다. 현재 브로커를 작업 지시·사용자 승인·완료 판정에 쓰지 않는 기존 규칙은 유지해야 한다.

## 4. 공식 Codex 표면과 설계 선택

OpenAI는 Codex [MCP 클라이언트 설정](https://learn.chatgpt.com/docs/extend/mcp?surface=cli)을 문서화한다. 이는 **Codex가 도구를 호출하는 통로**이지, 임의 MCP 서버가 실행 중인 Codex에게 자발적으로 새 모델 턴을 밀어 넣는 계약이 아니다. [App Server](https://learn.chatgpt.com/docs/app-server)는 관리 중인 스레드에 `turn/start`, 활성 턴에 `turn/steer`를 제공하며, `turn/steer`는 활성 턴 ID가 맞아야 한다. 이는 MoAI가 **스레드 수명과 권한을 소유한 클라이언트**를 별도로 구현할 경우의 후보이지, 기존 사용자 소유 CLI 세션에 브로커가 곧장 쓸 수 있다는 뜻이 아니다. [Codex 훅](https://learn.chatgpt.com/docs/hooks)은 정의된 이벤트 시점에 실행되고, `UserPromptSubmit`은 이미 제출되는 사용자 프롬프트에 추가 컨텍스트를 줄 수 있다. 대기 중인 빈 세션을 즉시 깨우는 구독 API로 취급하지 않는다.

설치 `codex queue --help`는 기존 스레드에 메시지를 대기시키는 명령을 표시한다. **이번 조사에서는 실제 전달·수신과 대상 세션 소유권을 실행 검증하지 않았다.** 공식 문서 계약과 실증 없이 이 명령을 MoAI의 핵심 push 구현으로 채택하지 않는다.

## 5. MoAI for Codex 기능 계약 재정의

| 기능 | 결정 | 사용자에게 약속할 동작 | 구현·검증 조건 |
|---|---|---|---|
| 초기화·진단 | **필수** | 현재 Codex CLI용 `AGENTS.md`, 스킬, agent, 훅, MCP를 설치·검사 | 같은 빌드의 생성물과 실제 Codex 로드·호출 확인 |
| plan→run→sync·goal | **필수** | Codex 단일 소유 세션에서 SPEC 작성, 구현, 독립 감사, 기록을 완주 | 실제 Codex 종단 fixture; 실패·중단·재개와 오판 0건 |
| 병렬 하위 작업 | **필수 범위 내 선택** | 같은 운영 세션의 Codex 네이티브 서브에이전트로 독립 작업을 나눔 | 명시적 파일 소유권·worktree·감사 결과 확인 |
| Claude→Codex 작업 위임 | **명시 호출 시 선택** | 기존 `codex_task`처럼 별도 Codex 작업을 시작하고 결과를 회수 | 사용자 승인·쓰기 opt-in·잡 상태·실패 복구를 판정. 살아 있는 타인 세션의 push로 설명하지 않음 |
| `session_msg_*` 사서함 | **기본 흐름에서 제외, 선택 기능으로 유지** | 같은 루트의 두 실행자가 등록·송신·수동 poll·ack할 수 있음 | `mailbox`로 명명, opt-in 노출, 루트/권한/TTL 진단, 실제 두 호스트 시험 |
| 살아 있는 Codex CLI 자동 push | **미지원** | 약속하지 않음 | 호스트 제어 통로·세션 소유권·동의·종단 지연/실패 시험 전에는 미지원 유지 |
| Codex 여러 카드 Factory | **미지원** | `-f`와 Claude 네이티브 세션 통신을 사용하지 않음 | 별도 기능으로 설계·검증하기 전 지원표에서 제외 |

### 기능에서 빼야 할 것

1. 기본 Codex 설명과 도움말에서 “Claude·Codex가 직접 메시지한다”, “push”, “online peer” 표현을 제거한다. 현재 확인된 동작은 **공유 디스크 사서함의 수동 조회**다. `CHANGELOG.md:103`의 “directly” 주장과 웹 UI의 “deliver” 문구도 지원 범위에 맞게 고친다.
2. 기본 plan/run/sync, goal, Factory, 완료 판정에서 `session_msg_*` 의존을 없앤다. 현재 검색 범위에서 호출 연결이 없어 동작 변경보다 **계약 정정**이 먼저다.
3. 네 도구를 즉시 삭제하지는 않는다. 단위 테스트와 과거 명시적 poll 왕복은 실제 효용을 보인다. 기본 MCP 도구 카탈로그에서 숨기는 것은 사용처 조사·호환성 이행·별도 opt-in 프로필을 마련한 다음 결정한다.
4. A2A 모양의 엔벨로프를 상호 운용 보증처럼 광고하지 않는다. 인증·구독·턴 주입은 엔벨로프 형식만으로 생기지 않는다.

### 빠진 기능을 채울 순서

| 단계 | 우선순위 | 작업 | 완료 게이트 |
|---|---|---|---|
| R0 | 높음 | `send`/`poll`/실제 push 용어를 분리하고 Codex 지원표·문서·진단을 고침 | 도움말·CHANGELOG·보고서에 자동 수신 주장 0건 |
| R1 | 높음 | Codex 자체 plan→run→sync·goal 경로를 브로커 없이 완주 | 실제 Codex CLI의 SPEC fixture와 검증 로그·독립 감사 일치 |
| R2 | 중간 | 사서함을 opt-in으로 정리하고 루트 고정·수신자 확인·권한·중복/TTL 처리 진단 추가 | 같은 루트/다른 루트, 오프라인, 중복, 만료, 승인 거부 사례 PASS |
| R3 | 중간 | 실제 Claude Code 호스트와 실제 Codex CLI에서 양방향 **수동 poll** 재시험 | 양쪽 도구 호출·응답·messageId·ack를 같은 실행 로그로 기록 |
| R4 | 조건부 | push가 제품 요구가 되면 MoAI 소유 Codex App Server 세션 또는 검증된 CLI 메시지 표면을 별도 어댑터로 구현 | 구독/턴 입력/사용자 표시/승인/중단/재시도/지연을 실세션에서 각각 입증 |
| R5 | 조건부 | R4가 성립한 경우에만 자동 알림을 지원표에 추가 | 세 OS와 대상 Codex 버전의 종단 회귀 검증 |

## 6. 검증 공백과 잔여 위험

**명시적 공백:** 현재 설치 버전에서 실제 Claude Code→기존 Codex CLI 세션의 자동 도착, `codex queue` 실전 전달, App Server를 통한 기존 사용자 소유 CLI 스레드 제어, 세 운영체제의 사서함 동작은 이번 조사에서 실행하지 않았다. 과거 e2e 기록은 별도 checkout·Codex 0.147.0·테스트 드라이버 조건이다.

**잔여 위험:** 공유 루트 오인, 이름 재등록에 따른 대상 착오, 하트비트의 온라인 오판, 수신자가 poll하지 않아 TTL이 지난 메시지, ack 전 재전달, 동일 프로젝트에 접근하는 다른 프로세스의 등록 ID 사용, MCP 쓰기 승인 거부가 남는다. `send` 성공을 “상대가 봤다”로 승격하면 이 위험이 제품 결함이 된다.

**최종 판정:** 브로커의 **파일 전송과 요청형 수신은 구현·단위 검증됐고**, 과거에는 MCP 드라이버↔`codex exec` 명시적 poll 왕복이 기록됐다. **실제 Claude Code→실행 중인 Codex CLI 자동 push는 현재 기능으로 입증되지 않았으며, 코드상 수신 트리거도 없다.** MoAI for Codex의 필수 기능에서 브로커를 빼고 선택적 사서함으로 정확히 이름 붙이는 것이 지금의 설계 결정이다.
