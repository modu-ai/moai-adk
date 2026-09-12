# 병렬 자식 둘의 요청 식별 안정성

## 판정과 실행

| 여정 | 결과 | 시간 | 산출물 |
|---|---|---:|---:|
| 부모 → Agent 둘 → 각 ToolSearch → 각 후속 응답 → 부모 | PASS | 0.85초 | 4 |

```sh
python3 .moai/reports/t649/appserver-redesign/probe/agent-identity-parallel.py
```

실제 Claude Code 2.1.269, 한 CLI 실행, exit 0, HTTP 요청 8개, 유료 요청 없음. 최초 모의 모델 응답에서 Agent tool_use 두 개를 같은 메시지에 반환했다. 각 실제 자식은 ToolSearch로 고정 로컬 MCP fixture를 발견하고 후속 응답을 요청했다. 실제 MCP echo는 실행하지 않았다.

## 구조 증거

```text
request 1 root
request 2 agent abdf056abf05980bf
request 3 agent a4c9673df8e2241c2
request 4 agent a4c9673df8e2241c2 + fixture tool_reference
request 5 agent abdf056abf05980bf + fixture tool_reference
request 6 root
request 7 root
request 8 root
```

두 자식의 x-claude-code-agent-id는 서로 달랐고 각자의 두 요청 사이에서 유지됐다. 자식 요청이 서로 끼어들어도 식별이 유지됐다. 부모 요청에는 agent-id가 없었다. session-id는 전 요청에서 같았다.

관측한 모든 x-claude-code-* 헤더 이름은 session-id와 agent-id 두 종류였다. parent-agent-id, parent-tool-use-id 헤더는 없었다. CLI 제어 출력은 자식 이벤트의 parent_tool_use_id=toolu_fixture_agent_1 또는 toolu_fixture_agent_2를 제공했다. 하지만 저장한 HTTP 필드만으로 어느 agent-id가 어느 parent_tool_use_id인지 결속하는 직접 필드는 없으므로 그 대응을 추측하지 않았다.

## 제한

한 번의 실행에서 자식 둘·각 두 요청만 확인했다. resume, 중첩 자식, 장기 실행·재시작, 자동 압축은 미검증이다. session/agent 헤더는 인증이 아니며 gateway의 인증된 family 범위와 결속해야 한다. 정확한 부모 도구 호출과의 매핑은 별도 제어 채널 근거가 필요하다.

임시 프로필/cwd·합성 인증·로컬 모의 서버, hooks 비활성화, 45초 실행 제한을 사용했다. 원시 프롬프트·계정/device 식별값·인증 헤더는 저장하지 않았다. 프로세스 정리 오류가 없었다. 이전 단일 자식 증거는 보존했다.

산출물: agent-identity-parallel.py, agent-identity-parallel-result.json, agent-identity-parallel-summary.json, agent-identity-parallel-verdict.md.
