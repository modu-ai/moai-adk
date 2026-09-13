# Claude picker·합성 reasoning 운반 — 19시 이후 단일 관측

## Claim

설치된 **Claude Code v2.1.268**을 허용 시각 이후 격리 HOME·프로젝트·전용 tmux·로컬 mock으로 한 번 실행했다. 생성된 요청은 9개다. 외부 제공자로 전달하는 mock 코드는 없으며 실제 OpenAI encrypted reasoning을 사용하지 않았다. 관측 범위는 부분 충족이고 전체 picker·reasoning 제품 PASS가 아니다.

| 항목 | 직접 관측 | 한계 |
|---|---|---|
| GPT 네 ID의 picker 표시 | gpt-6-astra, gpt-5.6-sol, gpt-5.6-terra, gpt-5.6-luna 행을 모두 확인 | 실제 제공자 지원·entitlement 판정 아님 |
| 네 모델의 세션 선택 | 각 ID에 대해 `Set model to ... for this session only`와 다음 picker의 선택 표시 확인 | Astra·Terra의 해당 사용자 턴 요청은 없음 |
| Sol·Luna 요청 model | Sol request-003/004, Luna request-005/006에 정확한 model | Astra·Terra 요청 라우팅은 Gap |
| s 동작 | 실제 안내는 `s to use this session only`; 네 s 이후 snapshot에 model 저장 없음 | s가 기본값 저장이라는 준비 단계 해석은 틀림 |
| Enter·Default 동작 | `Enter to set as default`; Default 행은 ANTHROPIC_DEFAULT_MODEL의 gpt-5.6-sol 표시, 선택 후 저장 안내 | LOCAL_DEFAULT_ROW를 포함한 요청은 0개라 실제 Default 라우팅 Gap |
| 최종 설정 | config/settings.json에 model=gpt-5.6-sol | Default 선택과 후속 직접 `/model gpt-5.6-sol` 뒤 snapshot이라 저장 시점 분리 불가; 새 프로세스 재시작 미측정 |
| 합성 reasoning·tool ID 왕복 | request-009에서 동일 redacted_thinking.data와 172바이트 tool ID가 같은 tool_use/tool_result 쌍으로 돌아옴 | 실제 OpenAI opaque·resume·유실 탐지·다른 provider 전환 미검증 |

Astra와 Terra의 s 직후 화면에는 대화 cache를 다시 읽는 모델 전환 확인창이 있었다. 이후 화면에는 세션 선택 완료가 나타났지만 해당 marker 요청은 없었다. 준비 스크립트의 고정 키 순서가 이 확인창을 별도 단계로 처리하지 않았으므로 이 관측을 모델 거절이나 라우팅 결함으로 단정하지 않는다. 재실행은 하지 않았다.

## Evidence

작업 디렉터리:
`/Users/goos/MoAI/moai-adk-go/.claude/worktrees/moai-proxy-unified`.

실행 전 `picker_reasoning_after19.md`, 실제 스크립트 전체, `reasoning-binding-candidate.md`를 읽었다. 시계를 반복 확인하고 허용 시각 이전에는 설치 client를 실행하지 않았다.

```text
$ clock.curr_time
2026-09-11 09:59:16 UTC
$ clock.curr_time
2026-09-11 10:00:09 UTC

$ timeout 280 python3 .moai/state/verify/01a08e7b-6aa0-7361-ab7e-ea8da1f02228/gateway-entry/picker_reasoning_after19.py
[exit 0; 하나의 실행 handle 55409를 끝까지 poll]
```

기존 스크립트의 내부 time gate는 유지했다. 실행 디렉터리는 아래와 같다.

`.moai/state/verify/01a08e7b-6aa0-7361-ab7e-ea8da1f02228/gateway-entry/picker-reasoning-20260911T100024Z-48d416f8/`

실제 summary의 판정 필드:

```json
{
  "status": "OBSERVED",
  "error": null,
  "request_count": 9,
  "synthetic_reasoning_emitted": true,
  "marker_echo_observed": true,
  "tool_result_observed": true,
  "marker_and_tool_result_same_request": true,
  "production_reasoning_carrier_validated": false,
  "upstream_inference_called_by_mock": false
}
```

화면 파일을 실제 읽은 문자열:

```text
screen-picker-initial.txt:
Claude Code v2.1.268
Default (recommended)  Use the default model (currently gpt-5.6-sol) · Set by ANTHROPIC_DEFAULT_MODEL
Enter to set as default · s to use this session only · Esc to cancel

screen-after-s-turn-gpt-6-astra.txt:
Set model to gpt-6-astra for this session only
screen-after-s-turn-gpt-5.6-terra.txt:
Set model to gpt-5.6-terra for this session only

screen-default-result.txt:
Set model to gpt-5.6-sol (default) and saved as your default for new sessions
screen-before-reasoning.txt:
Set model to gpt-5.6-sol and saved as your default for new sessions

screen-final.txt:
Read 1 file
LOCAL_MOCK_OK:gpt-5.6-sol
```

9개 raw JSON을 json.loads로 읽고 marker membership·model·SHA-256을 검산했다. `LOCAL_DEFAULT_ROW`는 전부 false였다. 조건 파일의 합성 marker·ID와 request-009의 실제 블록을 비교한 결과는 `carrier-readback.json`으로 저장했다.

```json
{
  "request": "request-009.json",
  "marker_bytes": 68,
  "tool_id_bytes": 172,
  "binding_call_id": "call_local_probe_read",
  "binding_hash_matches": true,
  "block_observations": [
    {"message_index": 11, "role": "assistant", "blocks": [
      {"type": "redacted_thinking", "marker_equal": true, "call_id_equal": false, "is_error": null, "tool_name": null},
      {"type": "tool_use", "marker_equal": false, "call_id_equal": true, "is_error": null, "tool_name": "Read"}
    ]},
    {"message_index": 12, "role": "user", "blocks": [
      {"type": "tool_result", "marker_equal": false, "call_id_equal": true, "is_error": null, "tool_name": null}
    ]}
  ]
}
```

null은 readback에서 원문에 is_error 필드가 없음을 뜻한다. tool ID의 원래 call_id 및 marker SHA-256은 조건 파일과 일치한다. 서명이나 제공자가 보증한 ID로 해석하지 않는다.

직접 확인한 요청 형태:

| 요청 | model | 형태 |
|---|---|---|
| 001 | claude-sonnet-5 | stream true, max_tokens 64000, thinking disabled, effort high + title JSON schema |
| 002 | claude-sonnet-5 | stream true, max_tokens 64000, thinking adaptive, effort high, clear_thinking_20251015 keep all |
| 003·005·007 | Sol·Luna·Sol | stream true, max_tokens 32000, thinking 필드 없음, effort high + title JSON schema |
| 004·006·008·009 | Sol·Luna·Sol·Sol | stream true, max_tokens 32000, thinking adaptive, effort high, clear_thinking_20251015 keep all |

GPT auxiliary 요청의 thinking 필드 부재는 disabled와 동일하다고 단정하지 않는다. 공통 request 키에는 messages, metadata, model, output_config, stream, system, tools, max_tokens가 있다. 사용자 턴에는 context_management와 thinking이 추가된다. 아직 일반 제품 번역기가 이 raw body를 수용한다는 뜻은 아니다.

정리 관측:

```text
$ [conditions.json의 overlay 경로에서 임시 HOME/project 상위 경로를 계산하여 Path.exists()]
isolated_scratch_exists= False
```

실행 종료 후 승인된 `ps -axo pid,ppid,args | rg 'gateway-picker-|picker_reasoning_after19.py'`에는 조회 명령 자체와 rg만 남았다. 첫 sandbox ps는 권한 오류로 실패하여 읽기 전용 재조회 승인을 받아 수행했다. 스크립트의 정리는 `tmux -L <고유 socket> kill-server`이며 전역 tmux kill을 실행하지 않았다.

## Baseline-attribution

```text
$ git rev-parse --short HEAD
81c1d58f9
$ shasum -a 256 .moai/state/verify/01a08e7b-6aa0-7361-ab7e-ea8da1f02228/gateway-entry/picker_reasoning_after19.py
975e437b914184a3aca1b19357655c42c67776b4329f35206245db9717060724  .moai/state/verify/01a08e7b-6aa0-7361-ab7e-ea8da1f02228/gateway-entry/picker_reasoning_after19.py
```

원시 요청의 관련 SHA-256:

```text
request-003.json 463224360574ca56de98422e4c913879cf9fa513071e1d3a3959792a3346aaf1
request-004.json 1c5040334c697f55a109bc0056995cd0f1fd46ef471c014d729047f7d9f95d65
request-005.json c82f923b9f6a30f3d0398ea792c430492df4d65ca625d065130af4937d186f17
request-006.json 96b6bb4c7f288c53f1c1783b98d426820d458d80d5f46d8ad8ca4219eed3f0d2
request-007.json db443a8165c69510f321e8c8d643848fef810b4aeeeafab136d9b1563a7b39a2
request-008.json cf8c99b6e0203b96510a007c098694c54b151eb35c9f00479fe633e53bf03eb3
request-009.json f3dff54fb5957cad3c6f14be5d65d031dbc3c6e4dfea604b6e4f167c4d63bd69
```

스크립트·제품 코드는 변경하지 않았다. 이 실행의 출력과 carrier-readback.json 및 본 보고서만 작성했다. 조건의 합성 토큰 외 실제 계정 자격은 사용하지 않았다.

## Gaps

네 모델 전부의 실제 요청·개별 기본값 영속성·새 프로세스 재시작은 아직 모두 검증되지 않았다. Astra/Terra와 Default의 사용자 요청 부재는 명시 Gap이다. UI 선택과 실제 라우팅, 설정 파일 쓰기와 새 세션 적용을 구분한다.

실제 OpenAI reasoning 형식·서버 수용·resume·foreign-provider strip·carrier 삭제 mutant의 0 egress는 시험하지 않았다. AC-MG-009의 계약 확정이나 제품 활성화로 확대하지 않는다. 직접 관측한 것은 합성 native redacted_thinking 및 172바이트 ID의 한 번의 Read 왕복뿐이다.

telemetry 억제 플래그는 설정했지만 OS 수준의 외부 네트워크 차단·전체 packet capture를 수행하지 않았다. mock의 외부 forwarding 없음과 실제 client의 모든 비본질 트래픽 부재는 다른 주장이다. 종료 전 client PID와 listener port를 별도 저장하지 않아 개별 PID reaping·port 닫힘을 독립 판정하지 못했다. 임시 경로 삭제와 종료 후 해당 probe 이름의 프로세스 부재만 추가 관측했다.

## Residual-risk

모델 전환 확인창을 처리하지 못한 probe 흐름으로 일부 턴이 전송되지 않았다. 후속 probe를 설계한다면 확인창을 별도 단계로 읽고 사용자 턴의 실제 요청을 확인한 뒤 다음 모델로 이동해야 한다. s/Enter 의미는 실제 화면을 따라야 하며, 이번 관측에서는 s가 세션 전용이고 Enter가 기본값 저장이다. 이번에는 요청받은 단 한 번만 실행했고 결과가 없는 항목을 채우려고 재시작하지 않았다.
