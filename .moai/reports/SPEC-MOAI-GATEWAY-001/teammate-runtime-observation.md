# 설치 Claude named Agent → teammate override 관측

## Claim

**관측 상태: OBSERVED_ONLY. 후보 callback 실행 확인. 전체 TEAMMATE 구현 PASS 아님.**

2026-09-11 UTC 10:00 이후 기존 `teammate_after19.py`를 한 번 실행했다. 실제 Claude Code 2.1.268의 대화형 화면과 로컬 mock 요청에서 named Agent schema를 확인했고, 합성 Agent 호출 뒤 `CLAUDE_CODE_TEAMMATE_COMMAND`에 지정한 관측 helper가 실행됐다. helper는 인수·허용한 환경 이름의 존재와 hash를 저장하고 종료했으며 실제 자식을 exec하지 않았다.

확인한 결과:

- `request-002.json`의 실제 Agent schema는 `name`을 포함한다. required는 `description`, `prompt`이며 `additionalProperties:false`다. model enum은 `sonnet`, `opus`, `haiku`, `fable`이다. `run_in_background` 필드는 없어서 mock payload에 넣지 않았다. `team_name`과 `mode`는 schema 설명에서 deprecated/ignored다.
- named 호출은 `name=gateway_probe`, `subagent_type=general-purpose`, `model=sonnet`으로 들어갔다. `request-003.json`에는 같은 tool call ID의 tool_result가 있고, 실행 성공을 나타내는 내부 metadata가 들어 있다. 최종 화면에도 Agent 호출과 named agent 항목이 나타난다.
- helper argv에는 `--agent-id`, `--agent-name`, `--team-name`, `--agent-color`, `--parent-session-id`, `--agent-type general-purpose`, `--model claude-sonnet-5`가 전달됐다. 정확한 값은 `helper-observed.json`에 보존했다.
- helper 환경에는 `ANTHROPIC_BASE_URL`, `CLAUDE_CONFIG_DIR`, `TMUX`, `TMUX_PANE`가 있었다. **lead에 넣은 합성 ANTHROPIC_AUTH_TOKEN은 helper에서 없었다.** `ANTHROPIC_CUSTOM_HEADERS`와 `CLAUDE_CODE_SUBAGENT_MODEL`도 없었지만 이 두 변수는 lead에 넣지 않았으므로 전달 중 제거됐다고 해석하지 않는다.
- helper의 TMUX_PANE hash는 `%1`의 SHA-256과 일치한다. 종료 직후 pane 목록에는 `%0` 하나만 남아 있었으므로 helper가 실행 중인 `%1` pane 목록을 직접 채집한 것은 아니다.
- runner 종료 후 helper PID·관측된 PPID·lead pane의 gtimeout PID 모두 존재하지 않았고 임시 HOME/project의 상위 디렉터리도 제거됐다.

## Evidence

시각 확인은 clock 도구로 했다. 실행 전 마지막 출력:

```text
2026-09-11 10:00:08 UTC
```

그 전에 UTC 09:54:07, 09:54:28, 09:56:33, 09:58:44, 09:59:46을 확인하고 대기했다. client 실행은 하지 않았다. runner 자체에도 UTC 10:00 이전 DEFERRED gate가 유지돼 있다.

단일 실행 명령:

```text
$ timeout 200 python3 .moai/state/verify/01a08e7b-6aa0-7361-ab7e-ea8da1f02228/gateway-entry/teammate_after19.py
{"status": "OBSERVED_ONLY", "error": null, "agent_call_emitted": true, "helper_observed": true, "requests": [{"file": "request-001.json", "model": "claude-sonnet-5", "path": "/v1/messages?beta=true", "sha256": "e55e1dfdfb22dfb8e4ef8af94f694816c9646c9b2a10b504fb7986dade1406fe"}, {"file": "request-002.json", "model": "claude-sonnet-5", "path": "/v1/messages?beta=true", "sha256": "dbdfdef605a0c57f58acde84439279ad9c68b1eba1fdc7494e00bd35dee1cd8d"}, {"file": "request-003.json", "model": "claude-sonnet-5", "path": "/v1/messages?beta=true", "sha256": "84024cbd3ac0da8e46a58ebaf683cc92a8b2dd4354f867b33e60e0b846243feb"}], "provider_forwarding": false, "production_bootstrap_validated": false, "evidence": "/Users/goos/MoAI/moai-adk-go/.claude/worktrees/moai-proxy-unified/.moai/state/verify/01a08e7b-6aa0-7361-ab7e-ea8da1f02228/gateway-entry/teammate-20260911T100025Z-923eb9a9"}
```

Exit 0. 최초 실행은 ongoing handle 51064를 반환했고 같은 handle을 poll하여 완료를 받았다. 재실행하지 않았다. run directory 시각은 UTC 10:00:25다. initial 화면의 신뢰 확인은 이 스크립트가 만든 임시 project에 한정해 응답했다.

실제 최종 화면에서 읽은 일부:

```text
Claude Code v2.1.268
Sonnet 5 · API Usage Billing
❯ NATIVE_TEAMMATE_PROBE: create the requested named local teammate once.
⏺ Agent(Observe local teammate launcher)
⏺ LOCAL_MOCK_ONLY
◯ gateway_probe  Reply LOCAL_ONLY. Do not call any tools or network...
```

위 billing 문구는 합성 토큰을 쓴 로컬 화면 표시이며 실제 청구 또는 공급자 호출의 근거가 아니다. pane 목록 원문:

```text
%0 77541 0 gtimeout
```

Python으로 관측 JSON의 pid/ppid 및 pane 목록 PID에 `os.kill(pid,0)`를 호출하고 임시 경로의 존재를 검사했다. 출력과 동일한 `cleanup-observed.json`을 보존했다.

```json
{
  "pid_checks": {
    "77929": "absent",
    "77540": "absent",
    "77541": "absent"
  },
  "temporary_project_parent_exists": false,
  "helper_pane_hash_matches": [
    "%1"
  ],
  "provider_forwarding": false
}
```

`provider_forwarding:false`는 runner의 로컬 mock가 upstream 전달을 하지 않는 구조와 summary에서 온 값이다. 전역 네트워크 패킷을 감시해 얻은 측정값으로 해석하지 않는다.

## Baseline-attribution

- WT: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/moai-proxy-unified`.
- 이번 실행 전 `git rev-parse HEAD`: `81c1d58f9cf7045594ee61d5e4ff380948ce9eba`.
- `shutil.which`와 `Path.resolve`로 확인한 설치 파일: Claude `/Users/goos/.local/share/claude/versions/2.1.268`, tmux `/opt/homebrew/Cellar/tmux/3.6a/bin/tmux`, timeout `/opt/homebrew/Cellar/coreutils/9.9/bin/gtimeout`.
- runner 실행 전후 SHA-256 동일: `0a0534cb25e8d9bc3a4eea55926be48a9edd5c9cd67a58f17f0ea67be3cfbbf5`.
- 모든 원본 관측은 `.moai/state/verify/01a08e7b-6aa0-7361-ab7e-ea8da1f02228/gateway-entry/teammate-20260911T100025Z-923eb9a9/`에 있다. `artifact-manifest.json`에 summary, helper, 세 요청, 두 화면, panes, cleanup의 크기와 SHA-256을 기록했다.
- helper 원본: 1,407 B, SHA-256 `72efe70719f9a0d862c188150b8671f01f341db1bcbb9172a108538fbad0dc2b`.
- 제품·runner source를 수정하지 않았다. 별도 tmux socket을 가진 서버만 스크립트에서 종료했으며 다른 세션의 서버에 종료 명령을 보내지 않았다.

## Gaps

- helper가 실제 Claude 자식을 실행하지 않았으므로 teammate의 실제 provider 요청, 응답 성공, 역할 분담 완료는 관측하지 않았다.
- production single-use bootstrap, 세션 접근 토큰 전달, profile/MCP/permission 보존, 부모 종료 처리, 여러 lead의 분리, 재개·재연결은 검증하지 않았다.
- 원본 TMUX 값과 socket 이름은 helper가 hash만 기록한다. cleanup은 알려진 세 PID 및 임시 경로에 대한 직접 관측이며, 모든 가능한 후손 process의 전체 목록을 전후 비교한 증거는 아니다.
- 코드에서 비필수 트래픽 비활성화 env를 설정하고 inference는 로컬 mock로 받았지만, 프로세스 전체 egress를 OS 수준으로 감시하지 않았다. 실제 공급자 credential을 입력으로 제공하지 않았다.
- callback 변수는 설치 바이너리에서 관측한 연결점이며 공식 안정 API 또는 다른 버전·플랫폼의 지원 보장으로 확대하지 않는다.

## Residual-risk

override helper가 필요한 agent 식별 인수를 받는다는 사실은 production bootstrap 구현의 근거가 된다. 그러나 lead의 합성 auth token이 helper에 자동 전달되지 않은 관측 때문에 별도의 명시적 bootstrap 전달 설계를 검증해야 한다. 현재 hybrid 활성화 제한을 이 후보 관측만으로 해제해서는 안 된다.
