# 공식 Codex snapshot의 GPT 문맥 metadata

## Claim

기존 공식 Codex 소스 snapshot의 내장 catalog에서 대상 GPT 네 exact ID가 모두 context_window=272000, max_context_window=872000으로 선언되어 있다. 이는 upstream 소스가 가진 기본 metadata이며 현재 계정 endpoint에서 측정한 최대 입력량이 아니다. 제품 capability를 이번 읽기로 활성화하지 않았다.

## Evidence

명령 `git -C /tmp/openai-codex-audit.zbvNXB rev-parse HEAD`의 출력:
```text
5a9eb145c4c05fcfc7158d7c25b80e1322eccae1
```
`git -C /tmp/openai-codex-audit.zbvNXB status --short -- codex-rs/models-manager/models.json codex-rs/app-server-protocol/src/protocol/v2/model.rs`는 exit0, 출력 없음이었다.

Python으로 models-manager/models.json을 json.loads하고 네 exact slug를 선택한 원본 필드값:
```json
[
  {
    "slug": "gpt-6-astra",
    "context_window": 272000,
    "max_context_window": 872000,
    "auto_compact_token_limit": null,
    "default_reasoning_level": "low",
    "supported_reasoning_levels": [
      {
        "effort": "low",
        "description": "Fast responses with lighter reasoning"
      },
      {
        "effort": "medium",
        "description": "Balances speed and reasoning depth for everyday tasks"
      },
      {
        "effort": "high",
        "description": "Greater reasoning depth for complex problems"
      },
      {
        "effort": "xhigh",
        "description": "Extra high reasoning depth for complex problems"
      },
      {
        "effort": "max",
        "description": "Maximum reasoning depth for the hardest problems"
      },
      {
        "effort": "ultra",
        "description": "Maximum reasoning with automatic task delegation"
      }
    ],
    "input_modalities": [
      "text",
      "image"
    ],
    "supports_parallel_tool_calls": true
  },
  {
    "slug": "gpt-5.6-sol",
    "context_window": 272000,
    "max_context_window": 872000,
    "auto_compact_token_limit": null,
    "default_reasoning_level": "low",
    "supported_reasoning_levels": [
      {
        "effort": "low",
        "description": "Fast responses with lighter reasoning"
      },
      {
        "effort": "medium",
        "description": "Balances speed and reasoning depth for everyday tasks"
      },
      {
        "effort": "high",
        "description": "Greater reasoning depth for complex problems"
      },
      {
        "effort": "xhigh",
        "description": "Extra high reasoning depth for complex problems"
      },
      {
        "effort": "max",
        "description": "Maximum reasoning depth for the hardest problems"
      },
      {
        "effort": "ultra",
        "description": "Maximum reasoning with automatic task delegation"
      }
    ],
    "input_modalities": [
      "text",
      "image"
    ],
    "supports_parallel_tool_calls": true
  },
  {
    "slug": "gpt-5.6-terra",
    "context_window": 272000,
    "max_context_window": 872000,
    "auto_compact_token_limit": null,
    "default_reasoning_level": "medium",
    "supported_reasoning_levels": [
      {
        "effort": "low",
        "description": "Fast responses with lighter reasoning"
      },
      {
        "effort": "medium",
        "description": "Balances speed and reasoning depth for everyday tasks"
      },
      {
        "effort": "high",
        "description": "Greater reasoning depth for complex problems"
      },
      {
        "effort": "xhigh",
        "description": "Extra high reasoning depth for complex problems"
      },
      {
        "effort": "max",
        "description": "Maximum reasoning depth for the hardest problems"
      },
      {
        "effort": "ultra",
        "description": "Maximum reasoning with automatic task delegation"
      }
    ],
    "input_modalities": [
      "text",
      "image"
    ],
    "supports_parallel_tool_calls": true
  },
  {
    "slug": "gpt-5.6-luna",
    "context_window": 272000,
    "max_context_window": 872000,
    "auto_compact_token_limit": null,
    "default_reasoning_level": "medium",
    "supported_reasoning_levels": [
      {
        "effort": "low",
        "description": "Fast responses with lighter reasoning"
      },
      {
        "effort": "medium",
        "description": "Balances speed and reasoning depth for everyday tasks"
      },
      {
        "effort": "high",
        "description": "Greater reasoning depth for complex problems"
      },
      {
        "effort": "xhigh",
        "description": "Extra high reasoning depth for complex problems"
      },
      {
        "effort": "max",
        "description": "Maximum reasoning depth for the hardest problems"
      }
    ],
    "input_modalities": [
      "text",
      "image"
    ],
    "supports_parallel_tool_calls": true
  }
]
```
파일 SHA-256: 897614e513a591b362f76c34e4b4d31af5809a2b52f8c0ab57f2990c5bf4ab3d

직접 읽은 protocol/src/openai_models.rs:385의 기본 effective percent는95, :449 max_context_window 주석은 config override의 최대값, :451 auto_compact 부재는 기본context의90%, :509 usable_context_window는 context*effectivepercent/100이다. 272000×95%=258400은 그 코드의 기본 headroom 계산 결과이며 별도 서버 상한 관측이 아니다.

app-server-protocol/src/protocol/v2/model.rs Model 응답 구조에는 reasoning efforts/modalities 등이 있고 context 수치 필드는 읽은 Model 구조에 없다. 모델 목록 표시만으로 context 값을 실계정에서 조회했다고 주장하지 않는다.

## Baseline-attribution

2026-09-11 부모 세션01a08e7b-6aa0-7361-ab7e-ea8da1f02228, 제품 WT-unified-gateway/81c1d58f9와 별개의 위 official source snapshot. 제품 파일 수정/실제 네트워크 호출/사용자 credential 접근 없음.

## Gaps

현재 서버 metadata 갱신값, 계정별 context 제한, 실제 tokenizer 정확성, 큰 입력의 runtime 수용, Claude/GLM cap는 검증하지 않았다. canonical source의 선언 수치와 실제 계정 관측을 구별한 설계·AC가 필요하다.

## Residual-risk

872000을 기본 context로 표시하거나, 95% headroom을 출력 token 상한으로 설명하지 않는다. catalog의 truncation_policy.limit=10000을 총 context 상한으로 혼동하지 않는다. 현재 제품의 byte·취소 제한과 입력 추정의 오차는 별도다.
