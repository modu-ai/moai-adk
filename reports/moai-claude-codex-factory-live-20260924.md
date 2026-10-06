# Claude Code ↔ Codex Factory 실세션 조사

2026-09-24 · 격리 시험 프로젝트: `/private/tmp/moai-cross-host-factory-20260924`, `/private/tmp/moai-factory-cross-test-20260924`

## 판정

**MoAI MCP의 `session_msg_*`로 양쪽 호스트가 명시적으로 송신·폴링·확인하는 왕복은 실측 성공했다.** Claude Code가 Codex에 시험 작업을 보냈고, Codex가 파일을 생성한 뒤 완료 메시지를 돌려보냈으며 Claude Code가 이를 확인했다. 이 시험은 두 개의 독립적인 비대화형 호스트 턴을 사람이 순서대로 실행한 결과다. 자동 도착, 기존 턴 깨우기, 카드 배차 규칙은 검증하지 않았다.

**설치된 `moai-adk v3.2.0-rc.13`에는 Codex `-f` 실행 경로가 있다.** `moai codex -f` lead와 `moai codex -f agent` worker 실행을 관찰했다. `agent`는 폐기 예정 별칭이며 `worker`로 전환하라는 메시지가 나온다. 현재 `moai codex --help`는 이 경로를 안내하지 않는다. 저장소 HEAD `2213871af`와 설치 바이너리 빌드 `g60017eb83`가 달라 소스 HEAD만으로 설치본 기능을 판정할 수 없다.

**두 방향의 Factory 카드 완료·병합 요청 흐름은 이번 실측에서 성립하지 않았다.** 각 방향에서 lead/worker의 run 등록까지는 관찰했지만 endpoint가 `launch_pending`이었고, Codex worker의 `factory_msg_send`는 `factory endpoint is launch-pending`으로 거절됐다. 이 상태가 대화형 터미널 신뢰 화면과 세션 시작 훅 미완료 때문인지, 제품 로직 문제인지는 아직 확정하지 못했다. 실제 카드 발행, 배차, 작업 완료, PR·병합 요청은 수행되지 않았다.

## Claim · Evidence · Baseline-attribution

측정 기준은 이 실행의 로컬 명령 출력이다. `moai --version` → `moai-adk v3.2.0-rc.13`, `codex --version` → `codex-cli 0.156.1`, `claude --version` → `2.1.281 (Claude Code)`. 공유 저장소를 수정하지 않도록 시험 데이터와 MoAI 홈을 임시 디렉터리에 격리했다.

| 관찰 | 실행 명령 | 실제 출력의 핵심 부분 |
|---|---|---|
| Claude 계정 | `claude auth status`; `CLAUDE_CONFIG_DIR=.../moai-adk claude -p 'Reply exactly: CLAUDE_ADK_LIVE_0924' --output-format json --model haiku` | 기본 프로필은 `api_error_status:429`; `moai-adk` 프로필은 `exit_code:0`, `result:"CLAUDE_ADK_LIVE_0924"` |
| Codex 수신자 등록 | `codex exec ... 'Call ... session_msg_register ...'` | `tool:"session_msg_register"`, `agentId:"codex-6c3bae73"` |
| Claude → Codex | `claude -p ... --mcp-config ... --strict-mcp-config ...` | Claude 결과: `claude-66852fa5`, `msg-72d33c2222dcc653`, 수신자 `codex-6c3bae73` |
| Codex 수신·작업·회신 | `codex exec ... 'Call session_msg_poll ... write receipt.txt ... session_msg_send ...'` | 수신 `msg-72d33c2222dcc653`, 파일 명령 `exit_code:0`, 회신 `msg-005746ae7b0cfc4f`, `ackedCount:1` |
| Claude 역방향 확인 | `claude -p ... 'call session_msg_poll for claude-66852fa5 ... ack_ids'` | `msg-005746ae7b0cfc4f`, 확인 수 `1`, 남은 수 `0` |
| 결과물 | `cat receipt.txt` | `CARD-DISPATCH-0924 completed` |

처음 Codex MCP 등록은 승인 정책 `never`에서 `MCP tool call requires approval`로 실패했다. 같은 격리 프로젝트에서 `--dangerously-bypass-approvals-and-sandbox`를 사용한 재시도는 성공했다. 이는 기본 승인 정책에서 무인 운행할 수 있다는 증거가 아니다.

## Factory 양방향 행렬

| 방향 | 실행한 진입점 | 확인된 단계 | 막힌 단계 |
|---|---|---|---|
| Codex lead → Claude worker | `moai codex -f`; `moai cc -p moai-adk -f worker` | Factory run `tluq0e` 등록. broker `peers`: `lead/codex`, `worker-1/claude`, 둘 다 `launch-pending:*` | 실제 대화형 세션 바인딩, 메시지 전송, 카드 처리 미검증 |
| Claude lead → Codex worker | `moai cc -p moai-adk -f`; `moai codex -f agent`; 별도 `moai codex -f worker-2 -- exec ...` | Factory run `tluq3d` 등록. `factory_msg_status`에 `lead/claude`, `worker-1/codex`, `worker-2/codex`가 `endpoint_state:live`, `binding_state:launch_pending` | `factory_msg_send` 결과 `factory endpoint is launch-pending`; 카드 배차 불가 |

다음은 관찰된 출력의 발췌다.

```text
$ moai codex -f agent
factory: `-f agent` (and the agent-<n> label) is a deprecated spelling — use `-f worker`; launching as worker-1
```

```text
$ moai factory runs
RUN ID  STATUS  OWNER  PID    CREATED                      UPDATED
tluq0e  active  live   20017  2026-09-24T04:33:50.583813Z  2026-09-24T04:33:50.583813Z
```

```text
"Capability":"hook-boundary","NextDelivery":"pending-until-next-turn"
"endpoint_state":"live","binding_state":"launch_pending"
factory_msg_send: factory endpoint is launch-pending
```

대화형 테스트가 끝난 뒤 시험 프로세스를 종료했고 `tluq0e`, `tluq3d`는 `moai factory runs --retire <run-id>`로 정리했다. `session_msg_*` 파일 브로커와 `factory_msg_*` Factory 브로커는 별개의 경로다. 전자의 성공을 후자의 배차 성공으로 확대해 해석할 수 없다.

## Gaps

- 대화형 TUI의 신뢰 화면 이후 세션 시작 훅이 실제로 실행돼 `launch_pending`이 해제되는지 확인하지 못했다.
- Codex lead ↔ Claude worker Factory 메시지의 송신·본문 수신·receipt를 확인하지 못했다.
- Claude lead ↔ Codex worker Factory 메시지는 송신 단계에서 거절됐다. 수신·receipt를 확인하지 못했다.
- `moai todo` 카드 발행, 작업용 worktree, plan→run→sync, 검사, PR 생성 및 병합 요청은 실행하지 않았다.
- 로컬 임시 프로젝트에서의 성공은 macOS/Linux/Windows 전체 플랫폼 보증이 아니다.
- 공유 저장소의 변경 파일이나 실제 운영 카드는 시험에 사용하지 않았다.

## Residual-risk와 다음 검증 순서

**High — 세션 바인딩을 먼저 재현한다.** 대화형 TUI 신뢰 단계를 완료한 뒤 Codex와 Claude 각각의 SessionStart 훅 출력, Factory `peers.session_uuid`, `factory_msg_status.binding_state`를 같은 run에서 수집한다. `launch_pending`이 지속되면 진입점의 세션 ID·프로세스 ID 전달과 훅 신뢰 경로를 조사한다. 훅을 우회해 브로커 DB에 직접 쓰는 시험으로 대체하지 않는다.

**High — 양방향 Factory 메시지 계약을 통과시킨다.** 바인딩이 확인된 뒤 `factory_msg_send → list → body → receipt → status`를 두 방향에서 각각 실행한다. 정상 전송, 중복 전송, 수신자 종료, 재시작 후 재연결을 별도 사례로 기록한다. `NextDelivery: pending-until-next-turn`이라는 현재 응답에 맞춰 자동 깨우기 주장을 검증한다.

**High — 실제 카드 한 건을 격리된 Git 프로젝트에서 끝까지 돌린다.** `moai todo add`로 시험 카드 발행, 명시적 선택과 worker 배차, worker의 별도 worktree 작업·테스트, 완료 보고와 lead의 근거 확인, 시험용 원격에 PR·병합 요청까지 순서대로 관찰한다. 두 lead 방향을 각각 통과시켜야 동등 지원이라고 부른다.

**Medium — 제품 표면을 맞춘다.** 현재 설치본의 `moai codex --help`에 Factory 진입점과 `worker` 표기를 드러내고, `agent` 별칭의 경고·제거 정책을 문서와 일치시킨다. 승인 정책 `never`에서 쓰기 MCP 호출이 막힌 사례를 무인 Factory 운영 조건에 반영한다.

**Medium — 성능을 측정한다.** 이 시험의 `codex exec` Factory 턴에는 `Exceeded skills context budget` 이벤트와 입력 토큰 `131052`가 나왔다. 별도 최소 템플릿·도구 표면 시험을 만들어 바인딩 성공률, 턴 입력 토큰, 메시지 지연, 재시도 횟수를 측정한 뒤 불필요한 스킬·MCP 설명을 줄인다. 단일 출력만으로 구조적 토큰 낭비를 확정하지 않는다.

공식 문서는 [OpenAI의 MCP 연결 안내](https://developers.openai.com/learn/docs-mcp)와 [Codex 플러그인 훅·신뢰 안내](https://developers.openai.com/plugins/guides/submit-claude-plugin)를 참조했다. 제품 지원 범위의 최종 판정은 위 실세션 테스트를 통과한 뒤 내려야 한다.
