# Native fork 실제 관측 — 2026-09-11

## Claim

설치된 Claude Code 2.1.268에서 같은 격리 native config를 유지하고 `--resume <parent> --fork-session --session-id <child>`를 함께 넘겼다. 새 요청의 metadata와 새 transcript는 사전에 지정한 child UUID를 사용했고, parent의 합성 reasoning carrier와 도구 호출·결과 ID를 유지했다. 정상 종료 후 parent transcript는 fork 직전과 바이트 단위로 같았다. 이는 native 인수 조합의 로컬 양성 관측이며 gateway 제품 또는 실제 OpenAI 통합 PASS가 아니다.

## Evidence

실행 명령:

```sh
/opt/homebrew/bin/timeout 310 python3 .moai/state/verify/01a08e7b-6aa0-7361-ab7e-ea8da1f02228/gateway-entry/native_fork_after19.py
```

종료 코드 `0`. 원본 stdout은 증거 디렉터리 `summary.json`과 같으며 주요 필드는 `status: OBSERVED`, `error: null`, `request_count: 4`, `production_reasoning_carrier_validated: false`, `upstream_inference_called_by_mock: false`다.

정확한 두 번째 native 명령:

```text
timeout 125 claude --resume c08b4132-3a5d-422a-a791-63ffa71e559b --fork-session --session-id 76745239-c346-4586-8eea-b7a5548a105f --settings /var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/gateway-picker-local-o59e88yz/project/picker-overlay.json --strict-mcp-config --allowedTools Read
```

`fork-comparison.json` 원문:

```json
{
  "parent_id": "c08b4132-3a5d-422a-a791-63ffa71e559b",
  "child_id": "76745239-c346-4586-8eea-b7a5548a105f",
  "parent_unchanged": true,
  "parent_before_sha256": "66d3a255216ca6b650b54bd882ef0bb0dae55ee17778756055047c3da370b722",
  "parent_after_sha256": "66d3a255216ca6b650b54bd882ef0bb0dae55ee17778756055047c3da370b722",
  "child_marker": true,
  "child_tool_id": true,
  "child_sha256": "68f91caef29b166a4740b772b5a9df87a47f64b581a92fbc49d7bdd9d4f28c08",
  "child_transcript_session_ids": [
    "",
    "76745239-c346-4586-8eea-b7a5548a105f"
  ]
}
```

저장된 raw request JSON의 `metadata.user_id`를 JSON decode한 결과:

```json
[
  {
    "file": "request-001.json",
    "model": "gpt-5.6-sol",
    "session_id": "c08b4132-3a5d-422a-a791-63ffa71e559b"
  },
  {
    "file": "request-002.json",
    "model": "gpt-5.6-sol",
    "session_id": "c08b4132-3a5d-422a-a791-63ffa71e559b"
  },
  {
    "file": "request-003.json",
    "model": "gpt-5.6-sol",
    "session_id": "c08b4132-3a5d-422a-a791-63ffa71e559b"
  },
  {
    "file": "request-004.json",
    "model": "gpt-5.6-sol",
    "session_id": "76745239-c346-4586-8eea-b7a5548a105f"
  }
]
```

`request-004.json`에서 carrier·tool_use·tool_result와 child UUID를 함께 판정한 결과:

```json
[
  {
    "file": "request-004.json",
    "sha256": "dc25cf07a71f4d346fc90ee9cd740bab7e224cc6ee10ff5b244c8e8ad15ecc41",
    "path": "/v1/messages?beta=true",
    "model": "gpt-5.6-sol",
    "stream": true,
    "anthropic_version": "2023-06-01",
    "anthropic_beta_tokens": [
      "claude-code-20250219",
      "interleaved-thinking-2025-05-14",
      "redact-thinking-2026-02-12",
      "thinking-token-count-2026-05-13",
      "context-management-2025-06-27",
      "prompt-caching-scope-2026-01-05",
      "mid-conversation-system-2026-04-07",
      "effort-2025-11-24"
    ],
    "marker_echo": true,
    "tool_result": true,
    "marker_and_tool_result": true,
    "session_id": "76745239-c346-4586-8eea-b7a5548a105f",
    "full_carrier_match": true,
    "tool_use_match": true,
    "tool_result_match": true
  }
]
```

두 native 프로세스 종료 관측 원문:

```json
{
  "exit": 0,
  "stdout": "1|0||97947|env\n",
  "stderr": ""
}
{
  "exit": 0,
  "stdout": "1|0||98415|env\n",
  "stderr": ""
}
```

## Baseline-attribution

이 실행에서 `moai session current` 출력은 `01a08e7b-6aa0-7361-ab7e-ea8da1f02228`이었다. `git -C <WT> rev-parse --short HEAD` 출력은 `81c1d58f9`, `git -C <WT> branch --show-current` 출력은 `WT-unified-gateway`였다.

- WT: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/moai-proxy-unified`
- native binary: `/Users/goos/.local/share/claude/versions/2.1.268`
- 관측 시작: `2026-09-11T11:05:10Z` (19:00 KST gate 이후)
- script SHA-256: `64536ac7323794ab74813979d71a1b7b70e392f41f1d933460cc35b23e07ac47`
- 증거: `.moai/state/verify/01a08e7b-6aa0-7361-ab7e-ea8da1f02228/gateway-entry/native-fork-20260911T110510Z-566c4f0f/`
- 기존 `picker_resume_after19.py`를 별도 파일로 복제하여 fork 인수, child UUID 판정, parent/child transcript 판정을 추가했다. 원본 및 제품·SPEC 파일은 변경하지 않았다.
- 모델은 local mock의 `gpt-5.6-sol`만 요청되었다. private HOME·config·project, `env -i`, synthetic auth token, strict MCP 설정을 사용했다. mock에는 upstream forwarding 코드가 없다.

## Gaps

실제 OpenAI reasoning 암호문, production receipt 서명·hash 상속, launcher의 소유 인덱스·lease·egress gate, parent 삭제 후 child 재개, 여러 fork의 경합, Windows native 동작은 검증하지 않았다. initial new와 fork 두 프로세스만 실행했으며 fork child를 다시 재개하는 세 번째 실행은 하지 않았다. child JSONL의 sessionId 없는 보조 행은 빈 문자열로 표시했다. credential 저장소를 읽거나 복사하지 않았고 사용자 계정 인증을 사용하지 않았다.

## Residual-risk

이 native 인수 동작은 설치 버전에 귀속된다. 이후 버전에서도 같은 UUID 의미가 유지되는지는 별도 호환 시험이 필요하다. 합성 marker의 유지 자체는 gateway receipt의 변조·손실 방지를 입증하지 않는다. parent transcript 불변은 해당 파일에 대한 관측이며 native profile 전체가 불변이라는 주장은 아니다.

## Cleanup

프로세스별 125초 timeout, probe alarm 290초, 외부 timeout 310초 및 ExitStack/finally 정리를 사용했다. 실행 후 ps·tmux·socket·filesystem을 직접 조회한 `cleanup-readback.json` 원문:

```json
{
  "scratch_removed": true,
  "first_pid_absent": true,
  "second_pid_absent": true,
  "tmux_exit": 1,
  "tmux_stderr": "no server running on /private/tmp/tmux-501/gateway-native-fork-97934-0e3554\n",
  "port_connect_errno": 61
}
```
