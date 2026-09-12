# Native Claude 재개 — 합성 reasoning·도구 ID·세션 ID 운반 관측

## Claim

설치 Claude Code 2.1.268의 두 프로세스를 동일한 격리 HOME·프로젝트에서 정상 실행하고, 첫 프로세스를 종료한 뒤 `--resume <정확한 세션 UUID>`로 재개했다. 첫 프로세스의 Read 도구 왕복 요청과 재개한 프로세스의 새 사용자 요청에서 **68바이트 합성 redacted_thinking.data, 172바이트 도구 호출·결과 ID, metadata.user_id 내부 session_id가 함께 유지되는 것을 확인했다.**

두 프로세스 모두 exit 0이었다. mock은 제공자로 전달하지 않으며 실제 OpenAI reasoning을 사용하지 않았다. 결과는 native 클라이언트의 합성 운반·저장·재개 관측이고 제품 또는 upstream PASS가 아니다.

## Evidence

실행 전 실제 설치 CLI 도움말을 읽었다.

```sh
env -u ANTHROPIC_BASE_URL -u ANTHROPIC_API_KEY -u ANTHROPIC_AUTH_TOKEN -u Z_AI_API_KEY /Users/goos/.local/bin/claude --help | rg -n -C 2 'resume|continue|session-id|model|persistence'
```

관련 출력:

```text
-c, --continue                        Continue the most recent conversation in
                                      the current directory
--fork-session                        When resuming, create a new session ID
                                      instead of reusing the original (use
                                      with --resume or --continue)
-r, --resume [value]                  Resume a conversation by session ID, or
                                      open interactive picker with optional
                                      search term
--session-id <uuid>                   Use a specific session ID for the
                                      conversation (must be a valid UUID)
```

첫 프로세스는 합성 UUID를 `--session-id`에 주고 `--model gpt-5.6-sol`로 시작했다. TUI의 Default를 선택한 직후 설정 snapshot을 남기고, mock이 합성 reasoning 및 hash-bound tool ID와 함께 Read 요청을 한 번 반환했다. 도구 결과가 돌아오고 native session JSONL에 carrier가 저장된 것을 확인한 뒤 `/exit`로 정상 종료했다. 둘째 프로세스는 같은 config/project에서 아래 인수로 시작했다. CLI model override는 없다.

```text
claude --resume 7f3010a7-5285-4ef8-a8a4-7237a2d0b6b1 --settings <격리 overlay> --strict-mcp-config --allowedTools Read
```

실행 명령:

```sh
/opt/homebrew/bin/timeout 310 python3 .moai/state/verify/01a08e7b-6aa0-7361-ab7e-ea8da1f02228/gateway-entry/picker_resume_after19.py
```

한 실행 handle 89056의 최종 stdout 중 판정 필드:

```json
{
  "status": "OBSERVED",
  "error": null,
  "request_count": 4,
  "synthetic_reasoning_emitted": true,
  "marker_echo_observed": true,
  "tool_result_observed": true,
  "marker_and_tool_result_same_request": true,
  "production_reasoning_carrier_validated": false,
  "upstream_inference_called_by_mock": false
}
```

exit 0이었다. 전체 stdout은 summary.json에 기록되어 있다.

원문 네 요청을 독립적으로 json.loads로 읽고, full marker bytes·tool_use.id·tool_result.tool_use_id를 대조했다. 모든 metadata.user_id JSON의 session_id도 같은 UUID인지 assert로 확인했다.

| 요청 | 프로세스/역할 | model | session_id | 정확한 carrier / tool_use / tool_result 개수 |
|---|---|---|---|---|
| 001 | 최초, title 요청 | gpt-5.6-sol | 7f3010a7-5285-4ef8-a8a4-7237a2d0b6b1 | 0 / 0 / 0 |
| 002 | 최초, 도구 호출 직전 | gpt-5.6-sol | 같은 UUID | 0 / 0 / 0 |
| 003 | 최초, Read 결과 왕복 | gpt-5.6-sol | 같은 UUID | 1 / 1 / 1 |
| 004 | 재개, 새 사용자 턴 | gpt-5.6-sol | 같은 UUID | 1 / 1 / 1 |

첫 왕복·재개 원문 SHA-256:

```text
request-003.json e5ef6a87b49ba353d58b2414f40163c6f4694b609e14957fd6a35045c2f76908
request-004.json 92b8896c0ef8493deff8c3560e5af69f8c97a3baf08613d0bd0b258aa3b7d2c6
```

tool ID 내부 base64url JSON을 decode하여 `opaque_sha256`이 합성 marker의 SHA-256과 일치함을 assert로 확인했다. 이 ID는 probe가 만든 비서명 형식이며 인증 receipt가 아니다.

저장된 native JSONL 세 시점 모두 각 줄의 JSON parse가 성공했고 marker와 tool ID가 존재했다.

```text
session-before-exit.jsonl marker_present True tool_id_present True d0e89f9603e20430c2440262a03aacaaa0517f2a145a33b2231055af07e3b415
session-after-exit.jsonl marker_present True tool_id_present True 0c93b9a5e09c5d86c2fe87b8a4d30caebc182c407eb845efc9b35434ecf90ee2
session-after-resume.jsonl marker_present True tool_id_present True ad1f8a7d3c9172248c4e02703087fd8a5c02eda8654da36f764a3816c5a88248
```

재개 후 화면은 새 `LOCAL_FOLLOWUP_resumed_...` 사용자 턴 아래 `LOCAL_MOCK_OK:gpt-5.6-sol` 응답을 보여 주었다. `Read 1 file` 및 이전 응답도 복원되었다. Default 선택 직후와 재개 후 `config/settings.json` 내용은 모두 `{}` + newline이었다. 이는 **재개한 세션에서 Sol이 요청된 관측**이며, 완전히 새로운 세션의 Default 적용을 증명하지 않는다.

### 요청 정책과 제한된 헤더 관측

raw 요청은 모두 stream=true, max_tokens=32000이다. 002/003/004에는 thinking=adaptive, output_config.effort=high, context_management.edits의 clear_thinking_20251015/keep=all이 있었다. 001에는 title JSON schema+effort high가 있었고 thinking/context_management 필드는 없었다.

헤더는 `Anthropic-Version`과 허용 문자 `[A-Za-z0-9._-]+`에 맞는 `Anthropic-Beta` token 이름만 기록했다. 인증·임의 custom 헤더는 기록하지 않았다. 재개 요청 004의 값:

```text
Anthropic-Version: 2023-06-01
Anthropic-Beta:
  claude-code-20250219
  interleaved-thinking-2025-05-14
  redact-thinking-2026-02-12
  thinking-token-count-2026-05-13
  context-management-2025-06-27
  prompt-caching-scope-2026-01-05
  mid-conversation-system-2026-04-07
  effort-2025-11-24
```

합성 토큰을 쓰는 mock 프로필의 헤더이며 실제 OAuth 세션의 beta 집합으로 일반화하지 않는다.

### 첫 진단 실패와 수정

처음 실행 handle 6415는 request_count=0, input-echo timeout으로 exit 1이었다. 증거 `picker-resume-20260911T103254Z-1b2393e2`와 실행 script 사본을 보존했다. 캡처된 화면에서 긴 Read fixture 경로가 두 줄에 걸쳐 보였고, 실제 화면 문자열에 원래 prompt를 검사하여 다음을 확인했다.

```text
EXACT_INPUT_ECHO False
WHITESPACE_NORMALIZED_INPUT_ECHO True
SYNTAX_OK
```

따라서 단말 줄바꿈을 포함한 literal membership 검사가 잘못된 실패를 만든다는 원인이 확인되었다. 화면 입력 관측에만 공백 정규화를 적용하고 정상+재개 두 프로세스로 구성된 수정 실행을 한 번 수행했다. 실패 실행은 Enter 이전에 멈췄으며 제품 또는 제공자 요청 실패가 아니었다.

## Baseline-attribution

작업 트리 `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/moai-proxy-unified`의 기존 지정 HEAD는 `81c1d58f9`다. 이번에는 새 ignored script·그 산출물·보고서만 작성했다. 실제 banner는 Claude Code 2.1.268이었다. 성공 실행 증거:

`.moai/state/verify/01a08e7b-6aa0-7361-ab7e-ea8da1f02228/gateway-entry/picker-resume-20260911T103342Z-9e003c97/`

실행 script SHA-256:

```text
d0bcfd831dbe4cc444a41aa785640534670c1452aacd7abdfbdb922425535a96
```

time gate는 2026-09-11 10:00 UTC 이후만 허용한다. 프로세스별 timeout 125초, probe alarm 290초, outer timeout 310초로 묶었다. 전용 tmux에는 remain-on-exit을 켜 정상 종료 상태를 기록했다. 동일 임시 config와 project를 두 프로세스에만 유지한 뒤 정리했다.

정상 종료 pane 상태의 형식은 pane_dead|pane_dead_status|pane_dead_signal|pane_pid|pane_current_command다.

```text
first:  1|0||32426|env
second: 1|0||33727|env
```

종료 후 PID 존재, 전용 tmux server, 임시 경로, port 연결을 별도로 확인한 원문:

```json
{
  "scratch_exists": false,
  "panes": [
    {"pid": "32426", "exists": false},
    {"pid": "33727", "exists": false}
  ],
  "tmux_exit": 1,
  "tmux_stderr": "no server running on /private/tmp/tmux-501/gateway-picker-resume-32375-e5b62e\n",
  "port_connect_ex": 61
}
```

## Gaps

`--continue`는 도움말로만 확인했고 실행하지 않았다. 새 프로세스에서 완전히 새로운 세션을 시작했을 때 Default가 적용되는지는 이번 범위가 아니다. 같은 provider의 하나의 합성 carrier만 시험했으며 실제 OpenAI encrypted reasoning, upstream 수용, provider 전환 제거, 일부 삭제·변조·재정렬·다른 세션 replay, 캐시 압축, native 설정 격리 전체 요구는 검증하지 않았다.

metadata.user_id.session_id의 유지 관측은 이 문자열을 보안상 신뢰할 수 있다는 증거가 아니다. hash-bound ID 역시 인증·위조 방지 수단으로 판정하지 않는다. 실제 Gateway/adapter 제품 코드가 이 요청을 받는 시험은 하지 않았다. 저장소 전체 CI 판정은 PENDING이다.

## Residual-risk

관측은 native Claude의 특정 버전, 동일 계정 없는 mock 프로필, 특정 세션에서 성립한다. 불투명 데이터나 tool ID 크기·형식이 달라지는 경우 및 다른 모델/설정/세션 재개 조건까지 보장하지 않는다. OS 수준 전체 packet capture는 하지 않았으므로 mock의 외부 forwarding 부재와 클라이언트 전체 부가 통신 부재를 구분한다.
