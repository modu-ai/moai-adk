# t649 도구 실패 뒤 GPT 400 오류 분석 및 수정

## Claim
Claude의 정상적인 `tool_result.is_error=true`를 MoAI Responses 변환기가 지원하지 않아 로컬에서 HTTP 400으로 거절했다. 실패를 명시하는 텍스트와 기존 오류 본문을 같은 function_call_output에 보존하도록 수정하고 로컬 rc.8을 교체했다.

## Evidence
사용자 지정 로그 `/Users/goos/.moai/state/gateway-conversations/families/25c2e533-f510-49ba-b639-24c4c94efd36/native/debug/25c2e533-f510-49ba-b639-24c4c94efd36.txt`:
- 773~779행: Bash 종료 코드 5 및 PostToolUseFailure.
- 791행: 다음 /v1/messages 요청.
- 793행: `API error (attempt 1/11): 400 400`.
- 동일 family의 native/projects JSONL 110행: tool_result의 is_error=true.
- 앞선 ListAgents/SendMessage와 후속 도구 실행은 성공했으므로 이 로그의 실패 시점은 통신 이후 Bash 오류 반환이다.

수정 전 최소 회귀 테스트 실행:
`go test -p 1 ./internal/gateway/translate -run TestRequestErrorToolResultPreservesFailureAndContinuation -count=1`

```text
--- FAIL: TestRequestErrorToolResultPreservesFailureAndContinuation (0.00s)
    tool_error_result_test.go:24: valid failed tool result rejected: error tool results require an explicit mapping
FAIL
FAIL github.com/modu-ai/moai-adk/internal/gateway/translate 0.348s
```

수정 후 실행:
`go test -p 1 ./internal/gateway/translate ./internal/gateway -count=1 -timeout=120s`

```text
ok  github.com/modu-ai/moai-adk/internal/gateway/translate 1.412s
ok  github.com/modu-ai/moai-adk/internal/gateway 6.531s
```

실제 GPT 구독 + Claude Code + Bash 호출 대조 실험(live_probe.py; 정확한 argv는 before.json/after.json):

| 관측 | 기존 설치본 | 수정 빌드 |
|---|---|---|
| Bash 명령 | exit 5 | exit 5 |
| 도구 결과 | is_error=true, Exit code 5 | is_error=true, Exit code 5 |
| 최종 모델 결과 | API Error: 400 Bad Request | MOAI_TOOL_FAILURE_HANDLED |
| 최종 is_error | true | false |
| 종료 코드 | 1 | 0 |

## Baseline-attribution
워크트리: /Users/goos/MoAI/moai-adk-go/.claude/worktrees/moai-proxy-unified
HEAD: 81c1d58f9cf7045594ee61d5e4ff380948ce9eba (기존 미커밋 변경 포함)
설치 경로: /Users/goos/go/bin/moai
BuildID: v3.2.0-rc.8-t649-tool-error-d123ddaa90fb
SHA256: 578beb982c489dc47a6ad11c1396cea3c07ef035677c53eb9aef8f2d2b885991
백업: /Users/goos/.moai/releases/moai-before-tool-error-06dd7f8fdc30
빌드 명령, 수정 파일 해시 및 버전 확인 원문: build.json.

## 구현
- request.go: boolean 형식 검증은 유지하고 true이면 `Tool execution failed (is_error=true).`를 도구 출력의 첫 블록에 추가한다. 원래 오류 본문과 call_id를 보존한다.
- 기존 true 무조건 거부 테스트를 정상 실패 결과의 번역 테스트로 교체했다. 잘못된 is_error 타입, 호출/결과 짝 검증은 유지한다.
- 회귀 테스트는 문자열·블록 배열·빈 오류 본문과 뒤따르는 사용자 메시지를 검증한다.

## 공식 근거
- [Claude 도구 오류 처리](https://platform.claude.com/docs/en/agents-and-tools/tool-use/handle-tool-calls): 실행 오류를 content와 is_error=true로 반환한다.
- [OpenAI 함수 결과 형식](https://developers.openai.com/api/docs/guides/function-calling#formatting-results): 함수 결과에 오류 코드·본문·성공/실패를 전달할 수 있다. 이번 실패 표시 문구는 MoAI가 선택한 번역 규칙이다.

## Gaps
기존 사용자 세션 전체 요청을 패킷으로 캡처하거나 재전송하지 않았다. 대신 로그의 실패 형식으로 단위 테스트와 실제 GPT 대조 실험을 수행했다. 원래 t1080 배차 작업을 재실행하거나 변경하지 않았다. 실험은 hooks/MCP를 끄고 무해한 Bash exit 5만 허용한 print 모드이다. Windows 실행 및 전체 저장소 테스트는 수행하지 않았다.

## Residual-risk
현재 실행 중인 GPT 프로세스에는 새 바이너리가 적용되지 않는다. 레인을 종료 후 `moai gpt -f lane-6`로 다시 실행한다. 이 변경은 잘못된 jq 명령 자체를 수정하지 않으며, 실패를 모델에 전달해 후속 처리를 계속하게 한다. 노력 수준 환경변수 override 현상은 별도 문제로 남는다.
