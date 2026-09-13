# Hybrid: 초기 native 동적 도구 + 늦게 발견한 도구 dispatcher

## 판정

| 여정 | 결과 | 시간 | 산출물 |
|---|---|---:|---:|
| native ToolSearch → late echo dispatcher → 정확히 두 번 호출 후 최종 응답 | FAIL (추가 호출 발생) | 26.4초 | 3 |

첫 두 호출의 경로와 인수는 정확했고 최종 표식도 일치했다. 그러나 echo 중복 호출 한 번과 ToolSearch 재호출 두 번이 추가되어 전체 인수조건은 FAIL이다. 도구 경로의 실행 가능성은 관측했지만 호출 횟수와 안정성을 합격으로 표시하지 않는다.

## 실행 근거

```sh
/opt/homebrew/bin/python3.13 .moai/reports/t649/appserver-redesign/probe/hybrid_probe.py
```

Codex 0.154.0, 관리형 ChatGPT 인증, Sol, turn/start 한 번. 단일 등록 dispatcher 외에 ToolSearch를 자체 정확한 스키마의 native dynamic function으로 등록했다.

```text
1 ToolSearch(query=select:fixture_echo)                 valid
2 moai_claude_tool(name=fixture_echo, marker=...)       valid
3 moai_claude_tool(name=fixture_echo, marker=...)       rejected by exact-two-call fixture
4 ToolSearch(query=select:fixture_echo)                 rejected extra
5 ToolSearch(query=select:fixture_echo)                 rejected extra
final=MOAI_HYBRID_OK
turn_status=completed
dispatch_validations=[true,true,false,false,false]
success=false
```

시험 클라이언트는 dispatcher를 통한 ToolSearch와 알 수 없는 이름·잘못된 인수를 허용하지 않았다. 또한 정확히 두 번이라는 시험 순서 제한에 따라 세 번째 이후 호출을 거절했다. 이는 제품 registry의 일반 정책이 아니라 이 실증의 품질 판정 조건이다. 새 호출 ID로 같은 합법적인 도구를 다시 호출하는 경우까지 제품에서 막아서는 안 된다. 최초 검색 결과에는 새 echo 스키마와 dispatcher 사용 지시를 반환했다. App Server 도구 등록 목록은 변경하지 않았다.

## 격리와 MCP 이벤트

기존 격리 설정을 유지했다. hook 이벤트는 없었다. MCP startup 이벤트의 안전한 구조를 추가로 캡처했다:

```json
{"threadId":"str","name":"str","status":"ready","error":"NoneType","failureReason":"NoneType"}
```

단순 빈 전체 시작 완료 이벤트가 아니라 이름을 가진 서버의 ready 상태임을 확인했다. 이름 값 자체는 저장하지 않아 어떤 서버인지 식별하지 못한다. 따라서 모든 MCP가 비활성화되었다는 주장은 할 수 없다. 네이티브 실행 배타성 검증은 계속 미해결이다.

## 잔여 위험

샘플 1회다. 왜 모델이 올바른 결과 뒤에 반복했는지는 확인하지 않았다. 호출 ID 기반 중복 처리 정책과 도구 결과 완료 의미를 구현과 회귀 검증에 포함해야 한다. 이 캡처는 request/call ID를 보존하지 않았으므로 반복이 새 ID였는지 동일 ID 재전송이었는지 구분할 수 없다. 따라서 프로토콜 replay 결함으로 분류하지 않는다. 오류가 없는 완료나 최종 표식만으로 이 여정을 PASS로 삼으면 안 된다. 이전 실증은 덮어쓰지 않았다. 추가 유료 호출은 하지 않았다.

산출물: hybrid_probe.py, hybrid-result.json, hybrid-verdict.md.
