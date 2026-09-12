# Picker 후속 관측 — Astra·Terra·Default 로컬 요청 확인

## 최신 판정

19:20 KST의 세 번째 관측에서 Astra·Terra의 s 선택 이후 실제 로컬 요청과 Default의 실제 Sol 요청을 확인했다. Default 선택 직후 snapshot은 `{}`이며 명시 model 값이 없다. 아래 두 실패 기록을 보존하고, 성공한 후속 실행의 근거를 문서 마지막에 추가했다. 전체 제품·실제 제공자 PASS는 아니다.

## 첫 실행 Claim

2026-09-11 19:00 KST 이후 새 `picker_followup_after19.py`를 격리 HOME·프로젝트·전용 tmux·로컬 mock으로 한 번 실행했다. 시작 확인 화면 이후 prompt 관측이 timeout되어 exit 1, INCONCLUSIVE로 끝났다. 요청은 0개이며 Astra·Terra·Default 검증은 충족하지 못했다. 실제 제공자·계정은 사용하지 않았다.

## Evidence

실행 전 clock:

```text
2026-09-11 10:12:27 UTC
```

원본 picker_reasoning_after19.py와 기존 관측 보고서·Astra/Terra 확인창·초기 picker 화면을 읽고 새 ignored 파일을 작성했다. `ast.parse` 출력은 `SYNTAX_OK`였으며 gate는 2026-09-11 10:00:00 UTC를 유지했다. 실행 명령:

```sh
/opt/homebrew/bin/timeout 280 python3 .moai/state/verify/01a08e7b-6aa0-7361-ab7e-ea8da1f02228/gateway-entry/picker_followup_after19.py
```

단일 handle 22438의 최종 출력:

```json
{
  "status": "INCONCLUSIVE",
  "error": "TimeoutError: unobserved condition: prompt",
  "picker_actions": [],
  "request_count": 0,
  "synthetic_reasoning_emitted": false,
  "marker_echo_observed": false,
  "tool_result_observed": false,
  "marker_and_tool_result_same_request": false,
  "production_reasoning_carrier_validated": false,
  "upstream_inference_called_by_mock": false,
  "evidence": "/Users/goos/MoAI/moai-adk-go/.claude/worktrees/moai-proxy-unified/.moai/state/verify/01a08e7b-6aa0-7361-ab7e-ea8da1f02228/gateway-entry/picker-followup-20260911T101336Z-09c09c4d"
}
```

첫 화면의 실제 선택 부분:

```text
 ❯ No, exit
   Yes, I trust this folder

 Enter to confirm · Esc to cancel
```

Down 이후 `screen-trust-selected.txt`에는 `❯ Yes, I trust this folder`가 선택된 상태가 실제 기록되어 있다. 그 뒤 Enter를 보냈으나 마지막 `screen-timeout-prompt.txt`는 빈 문자열이었다. 스크립트 대기는 `No, exit` 선택 문자열의 부재만 조건으로 삼아 그 조건 자체는 약하지만, 이번 실행의 별도 capture에서는 Yes 선택이 확인된다. Yes 선택 뒤 prompt가 나오지 않은 원인은 이번 기록으로 확정하지 않는다. pane 종료 전 stderr/exit code를 별도로 보존하지 않아 제품 종료 오류와 관측 문제를 구별할 수 없다. 이를 제품 라우팅 결함으로 해석하지 않는다.

정리 후 `owned-runtime.json`의 전용 socket·pane PID·port·scratch를 이용해 `tmux -L <owned socket> list-sessions`, `os.kill(pid, 0)`, `socket.connect_ex`, `Path.exists()`로 읽기 전용 확인했다. 출력:

```json
{
  "scratch_exists": false,
  "pane_pid_exists": false,
  "tmux_list_exit": 1,
  "tmux_list_stderr": "no server running on /private/tmp/tmux-501/gateway-picker-followup-63116-ccbd56\n",
  "port_connect_ex": 61
}
```

## Baseline-attribution

작업 트리 `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/moai-proxy-unified`, 부모가 지정하고 직전 작업 종료 시 직접 읽은 HEAD `81c1d58f9`. 실행 script SHA-256:

```text
303d5d9cd2346cf08f2d2c943714feccda4e6d9f1677d46d1bd9ef668aecdb04
```

증거 디렉터리는 위 summary의 evidence 경로다. 신규 script·실행 산출물·이 보고서만 작성했고 제품 코드와 원본 script는 변경하지 않았다. 새 script의 mock handler는 원본을 재사용했으며 외부 forwarding 코드가 없다. 조건 argv 초기 모델은 `claude-sonnet-5`, picker에는 요청된 GPT 네 ID와 Opus 5·Sonnet 5가 포함된다. 이번 실행은 버전 banner까지 도달하지 못하여 실제 실행 버전을 화면으로 재확인하지 못했다.

## Gaps

Astra·Terra의 s 선택 이후 첫 사용자 요청, Default 표시·요청·직후 설정 snapshot은 전부 미관측이다. 요청 0이므로 max_tokens·thinking·output_config·context_management 등 raw 정책 metadata도 없다. 실제 제공자 수용·통합 경로·CI도 관측하지 않았다. OS 수준 전체 네트워크 capture는 수행하지 않았다.

## Residual-risk

후속 작업에서는 설정 확인창의 부재가 아니라 `❯ Yes, I trust this folder`의 존재를 확인해야 한다. picker 이동도 새로 선택된 행이 비어 있지 않은 조건이 필요하다. 이번에는 명시된 단일 실행 범위를 지켜 재시작하지 않았다. 종료 상태와 정리 결과를 기록했으며, 이 실패를 picker 제품 PASS로 합산해서는 안 된다.


## 승인된 진단 후속 실행 — 10:18 UTC

### Claim

부모의 추가 지시에 따라 이전 성공 script와 이번 차이를 비교하고, 종료 상태 보존과 양의 선택 확인을 추가한 뒤 새 격리 실행을 한 번 수행했다. 이번에는 trust 선택 timeout이 발생했으며 요청은 0개였다. 모델 시험 단계에 도달하지 않았고 판정은 INCONCLUSIVE다. 첫 실패 산출물과 당시 실행 script 사본을 보존했다.

### Evidence

원본 성공 probe는 시작 후 5초가 지난 뒤 첫 입력을 보냈고, 후속 probe는 첫 화면 조건을 만족하는 즉시 입력했다. 이 차이를 diff로 확인했다. 기존 부재 조건의 약점은 다음 표현식 실행으로 확인했다. 이것이 첫 실패의 원인임을 증명하는 것은 아니다.

```text
SYNTAX_OK
Old absence predicate on blank screen: True
New positive predicate on blank screen: False
```

수정은 ignored probe에 한정했다. 전용 tmux config에 remain-on-exit을 켜고 pane_dead/status/signal/pid/command 및 terminal log를 보존했다. 선택 대기는 비어 있지 않은 선택 행과 Yes 문자열을 요구하도록 바꿨다. 실제 클라이언트 상태를 더 볼 수 있어 원인 구분에 도움이 되는 새 실행으로 판단했다. 동일한 timeout 280 실행 명령을 사용했으며 단일 handle 83243의 결과는 다음과 같다.

```text
exit 1
status: INCONCLUSIVE
error: TimeoutError: unobserved condition: trust-selection
request_count: 0
```

새 증거 디렉터리:

`.moai/state/verify/01a08e7b-6aa0-7361-ab7e-ea8da1f02228/gateway-entry/picker-followup-20260911T101833Z-257986c2/`

pane-timeout-trust-selection.json 원문:

```json
{
  "exit": 0,
  "stdout": "0|||97162|gtimeout\n",
  "stderr": ""
}
```

표시 순서는 pane_dead|pane_dead_status|pane_dead_signal|pane_pid|pane_current_command다. 즉 timeout 시 pane_dead=0으로 살아 있었다. terminal log와 timeout 화면에는 여전히 `❯ No, exit`이 보였고, 새 양의 확인 조건이 만족되지 않아 Enter는 보내지 않았다. 첫 Down 입력 이후 선택 전환이 관측되지 않은 사실까지 확인된다. 클라이언트 입력 준비 시점·단말 입력 전달·화면 갱신 가운데 정확한 원인은 여전히 확정하지 못했다.

### Baseline-attribution

실행 script 사본과 SHA-256은 새 증거 디렉터리에 보존했다. client-binary.json은 실행 PATH의 claude가 `/Users/goos/.local/share/claude/versions/2.1.268`로 해석됨을 기록한다. 모델 배너는 이번에도 관측하지 못했다. gate·격리 HOME·로컬 mock·전용 socket·전체 timeout을 유지했으며 원본 사용자 설정을 사용하거나 바꾸지 않았다.

### Gaps

Astra·Terra·Default 요청, 즉시 설정 저장, raw 정책 metadata는 여전히 0개/미관측이다. 이번 timeout의 원인을 클라이언트 결함으로 판정하지 않는다. 저장소 전체 CI 판정은 이번 작업 범위 밖이며 PENDING이다.

### Residual-risk

첫 렌더링만으로 입력 처리 준비가 끝났다고 판단하는 probe가 너무 이르게 키를 보냈을 가능성이 있다. 다음 관측을 하려면 선택 화면 안정성을 먼저 확인하고, 동일한 No 선택이 유지될 때만 입력을 다시 보내는 제한된 상태 기반 절차가 필요하다. 이는 진단에서 나온 후속 가설이며 이번 실행으로 입증한 원인은 아니다.

후속 실행 script SHA-256:

```text
560efe356b146662eac8a356839a7922512c12fe3df681b0552c62639cf168a6
```

종료 후 독립 정리 readback:

```json
{
  "scratch_exists": false,
  "pane_pid_exists": false,
  "tmux_list_exit": 1,
  "tmux_list_stderr": "no server running on /private/tmp/tmux-501/gateway-picker-followup-97132-c5d0eb\n",
  "port_connect_ex": 61
}
```


## 시작 안정화 후 관측 — 10:20 UTC

### Claim

부모의 계속 진행 지시에 따라 원본에서 성공한 시작 대기 순서를 회복하되, 고정 시간만 보내는 대신 예상 화면이 5초간 계속 보이는지를 확인했다. Yes 선택도 안정된 표시를 확인하고, prompt 입력은 화면에 전체 문자열이 나타난 뒤 Enter를 보냈다. 이 수정 후 단일 실행은 exit 0, OBSERVED로 끝났으며 8개 요청을 기록했다.

| 대상 | 관측한 조작 | 실제 요청 | 판정 범위 |
|---|---|---|---|
| Astra | 정확한 picker 행에서 s, cache Yes 확인 후 Enter | request-003/004 model=gpt-6-astra | 첫 사용자 턴 marker가 포함된 요청 모델 일치 |
| Terra | 정확한 picker 행에서 s, cache Yes 확인 후 Enter | request-005/006 model=gpt-5.6-terra | 첫 사용자 턴 marker가 포함된 요청 모델 일치 |
| Default | 현재 Sol 표시 행에서 Enter, cache Yes 확인 후 Enter | request-007/008 model=gpt-5.6-sol | 화면 표시와 실제 요청 모델 일치 |
| Default 직후 설정 | 직접 /model 명령·다음 사용자 턴 전에 snapshot | config/settings.json은 `{}` + newline | 명시 model 키 저장 없음 |

Astra·Terra s 직후 설정은 모두 `{}`였으며, Default 이후에도 snapshot에 있는 config/.claude.json, config/settings.json, home/.claude.json, project/.claude/settings.json, project/.claude/settings.local.json 어디에도 model 값이 없었다. 뒤의 세 경로는 존재하지 않았다. Default 선택 뒤 화면의 저장 안내만 보고 Sol ID가 설정 파일에 저장되었다고 주장하면 안 된다. 이번에는 직접 `/model gpt-5.6-sol`을 실행하지 않았다.

### Evidence

실행 전 AST와 양의 선택 조건 합성 확인:

```text
SYNTAX_OK
POSITIVE_SELECTION_SYNTHETIC_OK
```

동일 명령을 단일 handle 69730으로 실행했다.

```sh
/opt/homebrew/bin/timeout 280 python3 .moai/state/verify/01a08e7b-6aa0-7361-ab7e-ea8da1f02228/gateway-entry/picker_followup_after19.py
```

최종 출력의 판정 부분은 다음과 같고 전체 stdout 내용은 summary.json에 보존되어 있다.

```text
status: OBSERVED
error: null
request_count: 8
upstream_inference_called_by_mock: false
exit 0
```

실제 화면에서 읽은 행:

```text
Claude Code v2.1.268
❯ 1. Yes, switch to gpt-6-astra
Set model to gpt-6-astra for this session only
LOCAL_MOCK_OK:gpt-6-astra
❯ 1. Yes, switch to gpt-5.6-terra
Set model to gpt-5.6-terra for this session only
LOCAL_MOCK_OK:gpt-5.6-terra
❯ 1. Default (recommended)  Use the default model (currently gpt-5.6-sol) · Set by ANTHROPIC_DEFAULT_MODEL
Enter to set as default · s to use this session only · Esc to cancel
Set model to gpt-5.6-sol (default) and saved as your default for new sessions
LOCAL_MOCK_OK:gpt-5.6-sol
```

모든 raw JSON을 읽어 모델, 해당 단계의 고유 marker membership, SHA-256 일치를 assert로 확인했다.

```text
RAW_ROUTE_MARKER_HASH_READBACK_OK
```

Default 직후 settings.json의 SHA-256은 `ca3d163bab055381827226140568f3bef7eaac187cebd76878e0b63e9e442356`이며 내용은 `{}` + newline이다. Default 이후 사용자 턴 후에도 같은 hash였다. 시작 전과 s 직후 `{}`에는 newline이 없으므로 파일 bytes/hash는 다르지만 모델 값 저장은 없다.

| raw 요청 | max_tokens | thinking | output_config | context_management |
|---|---:|---|---|---|
| 001 Sonnet title | 64000 | disabled | high + title JSON schema | 필드 없음 |
| 002 Sonnet user | 64000 | adaptive | high | clear_thinking_20251015, keep all |
| 003·005·007 Astra·Terra·Sol title | 32000 | 필드 없음 | high + title JSON schema | 필드 없음 |
| 004·006·008 Astra·Terra·Sol user | 32000 | adaptive | high | clear_thinking_20251015, keep all |

모든 요청은 stream=true이고 `/v1/messages?beta=true`에 도착했다. title/user 분류는 output_config의 title schema 유무와 실제 messages 내용에 따른 관측이다. thinking 필드 부재를 disabled로 바꾸어 해석하지 않는다. 세부 원문 metadata와 각 raw SHA는 policy-readback.json에 보존했다.

### Baseline-attribution

증거 디렉터리:

`.moai/state/verify/01a08e7b-6aa0-7361-ab7e-ea8da1f02228/gateway-entry/picker-followup-20260911T102039Z-26558abd/`

실행 script SHA-256:

```text
fc7d8a992c2ad3d444ce9141bae4fa913e74a4a1e332ee8f5ea526730badcbc3
```

원본 script와 두 번의 실패 산출물·각 실행 script 사본을 보존했다. 제품 코드 변경은 없다. 조건은 임시 HOME/CLAUDE_CONFIG_DIR/project, env -i, 합성 토큰, strict MCP, 전용 tmux, 로컬 HTTP mock이며 실제 Claude 배너 버전은 2.1.268이었다. 원본처럼 초기 모델은 Sonnet 5이고 Opus 5를 별도 재실행하지 않았다.

종료 후 전용 자원 독립 확인:

```json
{
  "scratch_exists": false,
  "pane_pid_exists": false,
  "tmux_list_exit": 1,
  "tmux_list_stderr": "no server running on /private/tmp/tmux-501/gateway-picker-followup-11952-b2fda0\n",
  "port_connect_ex": 61
}
```

### Gaps

이번 결과는 설치 Claude와 로컬 mock의 TUI/HTTP 동작이다. 제품 gateway를 통한 제공자 실제 inference, GPT entitlement, 모델 출력 한도 수용, upstream reasoning, resume, 재시작 후 Default 적용은 검증하지 않았다. Sol/Luna의 s 검증은 기존 관측에 있으며 이번에는 Default에 필요한 Sol 요청만 관측했다. OS 수준 전체 네트워크 차단·packet capture는 하지 않았다. 통합 브랜치 전체 CI 결과는 PENDING이다.

### Residual-risk

시작 안정화와 입력 echo 확인을 함께 고친 뒤 관측이 성공했다. 이전 두 실패의 정확한 단일 원인을 인과적으로 분리한 것은 아니므로 "클라이언트 결함 수정"으로 부르지 않는다. Default 표시·현재 요청·즉시 저장 상태는 분리하여 확인했지만 새 프로세스의 동작이나 다른 설정 출처가 우선하는 경우까지 일반화할 수 없다.
