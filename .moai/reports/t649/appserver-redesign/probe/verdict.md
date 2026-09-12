# Codex App Server 동적 도구 실증

## 판정

| 여정 | 상태 | 시간 | 근거 |
|---|---|---:|---|
| 공식 stdio → 관리형 ChatGPT 인증 → Sol → 동적 moai_echo → 클라이언트 결과 → 최종 답변 | PASS | 11.25초 | dynamic-result.json |
| Claude만 도구를 실행하도록 네이티브 실행 완전 차단 | NOT-RUN | — | hooks/MCP 이벤트가 관측되어 별도 검증 필요 |

## 실행 기준과 근거

`codex --version` → `codex-cli 0.154.0`

```sh
codex app-server generate-json-schema --experimental --out .moai/reports/t649/appserver-redesign/probe/schema
codex app-server generate-json-schema --out .moai/reports/t649/appserver-redesign/probe/schema-stable
python3 .moai/reports/t649/appserver-redesign/probe/dynamic_probe.py
```

실행 결과:

```json
{
  "account_type": "chatgpt",
  "reported_model": "gpt-5.6-sol",
  "tool_calls": [
    {
      "tool": "moai_echo",
      "arguments": {}
    }
  ],
  "final_texts": [
    "MOAI_APPSERVER_ECHO_OK"
  ],
  "turn_status": "completed",
  "success": true,
  "process_exit": -15,
  "duration_seconds": 11.25
}
```

기존 Codex HOME와 공식 실행기에서 인증을 처리했다. account/read에서는 계정 유형만 보관했고 이메일·토큰을 읽거나 출력하지 않았다. 합성 ephemeral thread이며 작은 요청 한 번을 보냈다. 도구 결과는 클라이언트가 `contentItems:[{type:inputText,text:MOAI_APPSERVER_ECHO_OK}]`로 반환했다. 요청은 item/tool/call, 완료는 turn/completed로 확인했다. stdio 서버 프로세스 그룹에 SIGTERM을 보내 종료를 확인했다.

## 스키마 확인

- initialize.capabilities.experimentalApi=true를 사용했다.
- dynamicTools는 thread/start에 있으며 안정 스키마에는 없다. function/namespace, deferLoading을 지원한다.
- turn/start와 thread/resume에는 dynamicTools 필드가 없다. 생성된 ClientRequest 목록에서 도구 집합 교체 메서드는 확인되지 않았다.
- turn/start.model은 해당 턴 및 후속 턴의 모델 재정의다.
- turn/interrupt의 필수 인자는 threadId와 turnId다. 이번 정상 완료 실증에서 취소는 호출하지 않았다.
- thread/resume은 threadId를 받는다. history 인자는 명시적으로 UNSTABLE, FOR CODEX CLOUD, DO NOT USE다. 임의 Claude 이력 주입에 쓰면 안 된다.
- model/list에는 Sol/Astra가 반환되었으나 이 응답에서 context 최대 필드를 확인하지 못했다. 따라서 context 한도는 이 실증의 증명 범위가 아니다.

## 네이티브 실행 격리의 미검증 부분

실행 시 features.shell_tool=false, web_search=disabled, read-only sandbox, approval_policy=never를 적용했다. 그러나 hooks와 MCP startup 이벤트가 있었다. 프롬프트로 다른 도구를 금지한 것은 강제 격리의 증거가 아니다.

로컬 공식 소스 `/tmp/openai-codex-audit.zbvNXB`의 `codex-rs/core/src/tools/spec_plan.rs` 확인:

- 1090 부근: shell_tool=false는 셸 도구 등록을 차단한다.
- 1143: update_plan은 tools.update_plan.enabled에 따라 등록된다.
- 1255: apply_patch는 environment가 있고 모델이 지원하면 등록된다. shell_tool=false만으로 제거되지 않는다.
- 1271: view_image는 environment와 feature에 따라 등록된다.
- thread/start와 turn/start의 environments=[]는 스키마상 환경 접근을 비활성화한다. 위 환경 의존 도구의 제거 후보이지만 실행 검증은 하지 않았다.
- tools.experimental_request_user_input.enabled=false, features.codex_hooks=false 및 MCP 서버별 비활성화 등 추가 격리 구성이 필요하다. 전체 네이티브 도구를 일괄 allowlist로 제한하는 공개 설정은 이번 조사에서 찾지 못했다.

## 잔여 위험

동적 도구 왕복이 된다는 사실은 Claude Code의 승인·Agent·ToolSearch·resume 전체를 대체할 수 있다는 증거가 아니다. 고정 dispatcher 설계에서는 도구명/인자 스키마, caller thread/turn 결속, 단 한 번 실행, 취소 후 재호출 방지, Claude 승인 전달을 별도 검증해야 한다. 실증 중 기존 사용자 hooks가 실행되었으므로 향후 실증은 hooks/MCP/환경 접근을 먼저 격리해야 한다.

## 산출물

- dynamic_probe.py / dynamic-result.json
- schema-findings.json
- schema/ 및 schema-stable/: 설치본에서 생성한 프로토콜
- schema-generation.log / schema-stable-generation.log


## 추가 격리 실증

`/opt/homebrew/bin/python3.13 .moai/reports/t649/appserver-redesign/probe/isolation_probe.py` → exit 0.

6.79초, 관리형 ChatGPT 인증, Sol, moai_echo 1회, 최종 `MOAI_APPSERVER_ECHO_OK`, turn completed. thread/start와 turn/start 양쪽에 environments=[]를 주입했다. 셸·웹 검색에 더하여 hooks/plugin hooks/plugins/apps/view_image/multi-agent 기능을 false로 하고, plan 및 request_user_input을 비활성화했다. 사용자 config의 MCP 서버 이름만 열거하여 서버별 enabled=false override를 적용했다. 전체 인자는 isolation-result.json의 isolation_overrides에 저장했다.

hook 이벤트는 0개였다. MCP startup 이벤트는 1개 남았으며 payload는 캡처하지 않았다. 이 이벤트 하나로 서버가 실제 연결되었다거나 모두 비활성화되었다고 판정할 수 없다. 네이티브 도구 호출은 관측하지 않았지만, 사용 가능한 전체 도구 목록과 의도적 네이티브 호출 거절은 검증하지 않았다. 따라서 동적 도구 왕복 PASS, Claude만 실행한다는 배타성은 미검증으로 유지한다.

초기 설정 검사 3회는 MCP override 키의 따옴표 처리 때문에 서버 초기화 단계에서 실패했다. 모델 요청은 발생하지 않았다. `mcp_servers.node_repl.enabled=false` 형태로 수정한 뒤 위 추가 실증 1회가 성공했다. 설정 파일 자체는 수정하지 않았다.

추가 산출물: isolation_probe.py, isolation-result.json.
