# B: 고정 dispatcher 동적 도구 발견 실증

## 판정

| 여정 | 결과 | 시간 | 산출물 |
|---|---|---:|---:|
| 공식 App Server → ToolSearch dispatch → 새 도구 스키마 전달 → echo dispatch → 최종 답변 | PASS | 12.38초 | 3 |

한 개의 고정 moai_claude_tool만 thread/start.dynamicTools에 등록했다. 실제 모델은 GPT-5.6 Sol이고 기존 관리형 ChatGPT 인증을 사용했다. turn/start 1회에서 도구 요청 2회가 발생했다.

## 실행과 관측

```sh
/opt/homebrew/bin/python3.13 .moai/reports/t649/appserver-redesign/probe/dispatcher_probe.py
```

```json
[
 {"tool":"moai_claude_tool","arguments":{"name":"ToolSearch","arguments":{"query":"select:fixture_echo"}}},
 {"tool":"moai_claude_tool","arguments":{"name":"fixture_echo","arguments":{"marker":"MOAI_DISPATCHER_OK"}}}
]
```

```text
dispatch_validations=[true,true]
final_texts=[MOAI_DISPATCHER_OK]
turn_status=completed
success=true
errors=[]
```

고정 스키마는 name:string, arguments:object 두 필드가 필수다. 처음에는 ToolSearch 실제 스키마만 합성 작업 지시로 제공했다. 첫 호출 결과에 fixture_echo의 전체 input_schema와 tool_reference 구조를 텍스트 데이터로 전달했다. 모델이 이를 읽고 두 번째 호출에서 요구한 인수를 정확히 생성했다. App Server 동적 도구 목록을 교체하지 않았다.

클라이언트는 요청 순서·이름·객체 키·정확한 인수를 검사하여 일치할 때만 결과를 반환했다. 알 수 없는 이름과 잘못된 인수는 코드상 거절하지만, 이번 모델은 올바른 인수를 보냈으므로 거절 분기 자체의 runtime 증거는 아니다.

## 환경 및 미검증

Codex CLI 0.154.0. environments=[], 네이티브 셸·웹 검색·hooks·plugins·apps·view_image·multi-agent 비활성화와 사용자 MCP 서버별 비활성화를 앞선 격리 실증과 동일하게 사용했다. hooks 0개, MCP startup 이벤트 1개. 전체 네이티브 도구의 부재와 호출 거절은 여전히 미검증이다. 정상 완료 후 프로세스 그룹 SIGTERM, exit -15를 확인했다.

이 실증은 합성 도구 탐색·스키마 추종의 실행 가능성만 입증한다. 실제 Claude ToolSearch 호출이나 승인, Agent, 복잡한 중첩 인수, 도구 오류, 병렬 호출, 취소, 이미지 결과, 모델별 정확도와 비용 비교는 하지 않았다. 제품 채택 완료로 해석하면 안 된다.

## 산출물

- dispatcher_probe.py
- dispatcher-result.json
- dispatcher-verdict.md
