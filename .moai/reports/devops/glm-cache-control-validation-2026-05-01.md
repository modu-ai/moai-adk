# GLM cache_control 호환성 검증 보고서

**작성일**: 2026-05-01  
**작성자**: expert-devops (조사 전용, 코드 변경 없음)  
**조사 방법**: WebFetch + WebSearch (라이브 API 호출 없음)

---

## 1. 요약 (Executive Summary)

- Z.AI 공식 문서(`docs.z.ai`) 어디에도 `cache_control`, `prompt-caching-2024-07-31` 베타 헤더, `anthropic-beta` 지원 여부에 대한 명시적 설명이 없다. 지원도 미지원도 공식 문서화되지 않은 상태다.
- LiteLLM Issue #19923(2026-01-28)에서 Z.AI / GLM 모델에 `cache_control` 필드가 전송 시 **자동으로 제거(strip)** 되는 버그가 확인됐다. 이는 오픈소스 게이트웨이 레이어의 문제이나, Z.AI 프록시 자체의 동작 방식을 간접적으로 암시한다.
- Claude Code `CLAUDE_CODE_DISABLE_EXPERIMENTAL_BETAS=1`은 신규 베타 플래그 추가 시 **반복적으로 우회되는** 구조적 취약점을 가진다. 2026년 2월(Issue #22893), 3월(Issue #30926), 4월(Issue #49648)에 연속 회귀 발생.
- `cache_read_input_tokens` / `cache_creation_input_tokens`을 Z.AI 엔드포인트에서 실측한 공개 사례가 **존재하지 않는다**. 캐싱이 실제로 동작하는지 확인된 바 없다.
- **권장 정책**: 현행 유지(옵션 A). 단, `CLAUDE_CODE_DISABLE_EXPERIMENTAL_BETAS` 우회 회귀에 대응하는 모니터링을 추가할 것.

---

## 2. 현재 코드 상태

### 2.1 `DISABLE_PROMPT_CACHING=1` 주입 위치

| 파일 | 줄 번호(근사) | 함수/컨텍스트 | 역할 |
|------|------------|-------------|------|
| `internal/cli/glm.go` | 176 | `setGLMEnv()` | `moai glm` 실행 시 프로세스 env 주입 |
| `internal/cli/glm.go` | 372 | `injectTmuxSessionEnv()` | tmux 세션 레벨 env 주입 (GLM 팀 모드) |
| `internal/cli/glm.go` | 548 | `injectGLMEnvForTeam()` | settings.local.json 에 팀 모드 주입 |
| `internal/cli/glm.go` | 827 | `buildGLMEnvVars()` | env 맵 생성 헬퍼 (공통 참조 소스) |
| `internal/cli/glm.go` | 872 | `injectGLMEnv()` | settings.local.json 에 직접 주입 |
| `internal/hook/session_start.go` | 245–246 | `ensureGLMCompatibilityFlags()` | SessionStart 훅에서 누락 시 보완 주입 |

> **참고**: `glm.go:142–147` 코드 주석에 주입 이유 명시:  
> `DISABLE_PROMPT_CACHING=1` → 프롬프트 캐싱 미지원으로 인한 전체 시스템 프롬프트 재전송(~30-40K 토큰/요청) → GLM 컨텍스트 소진 가속

### 2.2 `CLAUDE_CODE_DISABLE_EXPERIMENTAL_BETAS=1` 주입 위치

동일 6개 위치에서 `DISABLE_PROMPT_CACHING`과 항상 쌍으로 주입된다. `internal/github/auth/glm.go:35–36`에서는 GitHub Actions CI용으로 `DISABLE_BETAS=true`라는 별도 변수명으로도 등록된다.

### 2.3 회귀 가드 위치

| 파일 | 줄 번호(근사) | 테스트 내용 |
|------|------------|-----------|
| `internal/cli/launcher_test.go` | 745–746 | `moai cc` 전환 후 settings.local.json에 `DISABLE_PROMPT_CACHING` 잔류 금지 (Issue #676) |
| `internal/cli/glm_test.go` | 153–154 | `moai glm` 종료 후 settings.local.json에 `DISABLE_PROMPT_CACHING` 잔류 금지 (Issue #676) |

두 테스트 모두 **잔류(leak) 방지** 목적이며, 주입 자체의 정확성을 검증하지는 않는다.

### 2.4 `moai cc` 전환 시 정리 로직

`internal/cli/launcher.go:235–243`: `moai cc` 실행 시 settings.local.json에서 다음 키를 `delete()` 처리한다.

```
ANTHROPIC_BASE_URL
ANTHROPIC_DEFAULT_{HAIKU,SONNET,OPUS}_MODEL
CLAUDE_CODE_DISABLE_EXPERIMENTAL_BETAS
DISABLE_PROMPT_CACHING
API_TIMEOUT_MS
CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC
```

tmux 세션 환경도 `clearTmuxSessionEnv()`(`glm.go:392–419`)에서 동일하게 정리된다.

---

## 3. Z.AI 공식 문서 분석

### 3.1 FAQ (`https://docs.z.ai/devpack/faq`)

`cache_control`, `anthropic-beta`, `prompt-caching`, `experimental_betas`, `Error 1214` 등의 키워드에 대한 **언급이 전혀 없다**. 문서 내용은 GLM Coding Plan 구독 정보, 모델 설정, 쿼터 관리, MCP 도구에 국한된다.

지원 모델 목록: GLM-5.1, GLM-5-Turbo, GLM-4.7, GLM-4.5-Air.

### 3.2 Quick Start (`https://docs.z.ai/devpack/quick-start`)

**엔드포인트 URL**:
- Coding API: `https://api.z.ai/api/coding/paas/v4`
- General API: `https://api.z.ai/api/paas/v4`
- **Anthropic 호환**: `https://api.z.ai/api/anthropic`

**환경변수 계약**:

```
ANTHROPIC_AUTH_TOKEN = <Z.AI API Key>
ANTHROPIC_BASE_URL   = https://api.z.ai/api/anthropic
API_TIMEOUT_MS       = 3000000 (선택, 기본 제공 값 예시)
```

**캐싱 관련 언급**: 없음.  
**베타 헤더 관련 언급**: 없음.  
**제한사항**: GLM Coding Plan은 Claude Code, Roo Code, Cline 등 공식 지원 도구 내 사용으로 엄격 제한.

### 3.3 Claude Code 통합 가이드 (`https://docs.z.ai/devpack/tool/claude`)

**지원 환경변수**:

| 변수명 | 용도 |
|--------|------|
| `ANTHROPIC_AUTH_TOKEN` | Z.AI API 키 |
| `ANTHROPIC_BASE_URL` | `https://api.z.ai/api/anthropic` |
| `API_TIMEOUT_MS` | `3000000` (5000분 타임아웃) |
| `ANTHROPIC_DEFAULT_OPUS_MODEL` | `GLM-4.7` 또는 `glm-5.1` (프리미엄) |
| `ANTHROPIC_DEFAULT_SONNET_MODEL` | `GLM-4.7` 또는 `glm-5-turbo` (프리미엄) |
| `ANTHROPIC_DEFAULT_HAIKU_MODEL` | `GLM-4.5-Air` |

**캐싱 관련 언급**: 없음.  
**베타 헤더 관련 언급**: 없음.  
**알려진 이슈**: usage 데이터 포맷 불일치로 비용/토큰 카운트 표시 오류 가능. macOS 권한 이슈. Windows 미지원(부분적).

> **결론**: Z.AI 공식 문서 3개 페이지 모두 `cache_control`, `anthropic-beta`, 또는 프롬프트 캐싱에 대해 **어떠한 언급도 없다**. 지원 여부에 대한 공식 입장이 존재하지 않는다.

---

## 4. Community Evidence

### 4.1 `cache_control` 스트리핑 버그 (LiteLLM, 2026-01-28)

**출처**: [BerriAI/litellm Issue #19923](https://github.com/BerriAI/litellm/issues/19923)  
**날짜**: 2026-01-28  
**내용**: GLM/ZAI 모델에 대한 `cache_control` 및 `thinking/reasoning` 파라미터가 정상 작동하지 않는다. 근본 원인은 `ZAIChatConfig`가 `OpenAIGPTConfig`를 상속하며, `transform_request()`에서 `remove_cache_control_flag_from_messages_and_tools()`를 호출하기 때문이다. 즉, **`cache_control` 필드가 요청에서 제거된 상태로 Z.AI로 전송**된다. 제안된 수정사항: `ZAIChatConfig`에서 해당 메서드를 오버라이드하여 GLM 모델은 `cache_control`을 제거하지 않도록.

**의미**: LiteLLM 게이트웨이 레이어의 버그이나, Z.AI 엔드포인트가 `cache_control`을 어떻게 처리하는지(또는 처리하지 않는지)에 대한 커뮤니티 내 혼란이 존재함을 확인.

### 4.2 GLM-5 캐싱 지원 요청 (openclaw, 2026-02-23)

**출처**: [openclaw/openclaw Issue #24497](https://github.com/openclaw/openclaw/issues/24497)  
**날짜**: 2026-02-23  
**내용**: `contextPruning` with `mode: "cache-ttl"` 기능이 ZAI/GLM-5를 지원하지 않음을 버그로 보고. GLM-5 네이티브 캐싱은 `$0.20/M` (표준 대비 1/5 가격)으로 지원된다고 주장. `isCacheTtlEligibleProvider()` 함수의 하드코딩된 검사가 `"zai"`를 포함하지 않아 캐싱이 무음으로 비활성화됨.

**중요 구분**: 이 이슈에서 말하는 "GLM-5 네이티브 캐싱"은 Z.AI의 자체 캐싱 메커니즘을 가리키는 것으로 보이며, Anthropic의 `cache_control: { type: "ephemeral" }` 방식과 동일한지는 **미확인**.

### 4.3 Cerebras `zai-glm-4.7` 프롬프트 캐싱 활성화 요청 (Roo-Code, 2026-01-10)

**출처**: [RooCodeInc/Roo-Code Issue #10601](https://github.com/RooCodeInc/Roo-Code/issues/10601)  
**날짜**: 2026-01-10  
**내용**: Cerebras에서 제공하는 `zai-glm-4.7` 모델(Z.AI의 GLM-4.7과 동일 모델 가중치, 다른 인프라)에 대해 프롬프트 캐싱이 비활성화 상태. Cerebras 공식 문서는 지원 명시. `supportsPromptCaching = true`로 수정 요청.

**중요 구분**: 이 이슈는 Cerebras 인프라에서 실행되는 GLM-4.7에 관한 것이며, `api.z.ai/api/anthropic` 엔드포인트와 **직접 관련 없음**. 단, 동일 모델 가중치를 기반으로 캐싱 지원 가능성을 시사한다.

### 4.4 `CLAUDE_CODE_DISABLE_EXPERIMENTAL_BETAS` 구조적 취약점

**출처**: 복수의 anthropics/claude-code GitHub 이슈

| 이슈 | 날짜 | 내용 |
|------|------|------|
| [Issue #22893](https://github.com/anthropics/claude-code/issues/22893) | 2026-02-xx | `advanced-tool-use-2025-11-20` 베타 플래그가 `DISABLE_EXPERIMENTAL_BETAS=1`에도 불구 전송됨 |
| [Issue #30926](https://github.com/anthropics/claude-code/issues/30926) | 2026-03-xx | v2.1.69에서 `advanced-tool-use-2025-11-20` 도입. 프록시 설정 시 `firstParty`로 감지되어 비호환 플래그 전송 |
| [Issue #20031](https://github.com/anthropics/claude-code/issues/20031) | 2026-01-xx | `DISABLE_EXPERIMENTAL_BETAS=1` 설정에도 `?beta=true` 쿼리 파라미터 잔류 |
| [Issue #49648](https://github.com/anthropics/claude-code/issues/49648) | 2026-04-xx | v2.1.112에서 `claude-code-20250219`, `interleaved-thinking-2025-05-14` 등 베타 플래그가 Bedrock에 전송되어 400 오류 |

**패턴**: Claude Code 신규 베타 플래그 추가 시 `DISABLE_EXPERIMENTAL_BETAS` 게이팅이 누락되는 패턴이 반복된다. v2.1.101에서 한 번 수정됐으나 이후에도 회귀가 지속됐다.

### 4.5 Error 1214 관련

공개 GitHub 이슈나 블로그에서 Z.AI + Error 1214 조합에 대한 문서화된 최신 사례를 발견하지 못했다. memory 파일(`glm_compatibility.md`, 30일 이상 경과)의 기록은 최신 상태를 반영하지 않을 수 있다 — **현재 재현 가능 여부 불명확**.

### 4.6 `cache_read_input_tokens` 실측 사례

**결론**: `api.z.ai/api/anthropic` 엔드포인트에서 `cache_read_input_tokens` 또는 `cache_creation_input_tokens`이 0이 아닌 값으로 반환됐다는 공개 보고가 **없다**. 캐싱이 실제로 동작함을 경험적으로 확인한 사례가 존재하지 않는다.

---

## 5. 모델별 호환성 매트릭스

| 모델 | beta headers 허용 | cache_control 처리 | Anthropic 캐싱 실측 | 비고 |
|------|-----------------|-------------------|-------------------|------|
| GLM-5.1 | 미확인 (문서 없음) | 미확인 | 없음 | Z.AI 최고 성능 모델 |
| GLM-5-Turbo | 미확인 | 미확인 | 없음 | — |
| GLM-4.7 | 30일 전 기록: 베타 헤더 허용 가능 (stale) | 미확인 | 없음 | Cerebras 동명 모델은 캐싱 지원 명시 (다른 인프라) |
| GLM-4.6 | 미확인 | 미확인 | 없음 | 200K ctx, 에이전트/도구 특화 |
| GLM-4.5-Air | 미확인 | 미확인 | 없음 | 저가 Haiku 슬롯 |

**출처 주기**:
- Z.AI 공식 문서 (2026-05-01 조회): 캐싱/베타 헤더 언급 없음
- [openclaw #24497](https://github.com/openclaw/openclaw/issues/24497): GLM-5 네이티브 캐싱 가격 언급 (Z.AI 자체 캐싱, 2026-02-23)
- [LiteLLM #19923](https://github.com/BerriAI/litellm/issues/19923): cache_control 스트리핑 확인 (2026-01-28)
- Memory `glm_compatibility.md` (stale, ~30일 전): GLM-5.1 베타 헤더 거부, GLM-4.7 무관

> **GLM-5.1 vs GLM-4.7 차이**: 30일 전 메모리에 GLM-5.1은 `anthropic-beta` 헤더를 거부하고 GLM-4.7은 영향 없었다는 기록이 있으나, Z.AI 엔드포인트의 펌웨어/프록시 업데이트로 현재 상태는 다를 수 있다. **stale 데이터로 취급**.

---

## 6. 검증 절차 (실 API 키 사용 시)

아래는 실제 Z.AI API 키 보유자가 수행할 수 있는 검증 절차 의사코드다.

### 6.1 단계 1: 베타 헤더 수용 여부 확인

```bash
# anthropic-beta 헤더를 포함한 최소 요청
curl -X POST https://api.z.ai/api/anthropic/v1/messages \
  -H "x-api-key: $ZAI_API_KEY" \
  -H "anthropic-version: 2023-06-01" \
  -H "anthropic-beta: prompt-caching-2024-07-31" \
  -H "content-type: application/json" \
  -d '{"model":"glm-4.7","max_tokens":10,"messages":[{"role":"user","content":"hi"}]}'
# 확인 사항: HTTP 400 (헤더 거부) vs 200 (수용 또는 무시)
# 오류 메시지에 "invalid beta" 또는 "unexpected beta" 포함 여부 확인
```

### 6.2 단계 2: cache_control ephemeral 전송 후 usage 확인

```bash
# system 프롬프트에 cache_control 마킹
curl -X POST https://api.z.ai/api/anthropic/v1/messages \
  -H "x-api-key: $ZAI_API_KEY" \
  -H "anthropic-version: 2023-06-01" \
  -H "content-type: application/json" \
  -d '{
    "model": "glm-4.7",
    "max_tokens": 10,
    "system": [
      {
        "type": "text",
        "text": "You are a helpful assistant. '"$(python3 -c 'print("x"*2048)')"'",
        "cache_control": {"type": "ephemeral"}
      }
    ],
    "messages": [{"role": "user", "content": "hi"}]
  }'
# 응답 usage 필드 확인:
# "usage": {
#   "input_tokens": NNN,
#   "cache_creation_input_tokens": 0,  <-- 0이면 캐싱 안 됨
#   "cache_read_input_tokens": 0       <-- 0이면 캐시 히트 없음
# }
```

### 6.3 단계 3: 두 번째 요청으로 cache_read 확인

동일 요청을 반복하여 `cache_read_input_tokens > 0`인지 확인한다. 동일 시스템 프롬프트가 5분 TTL 내에 재사용되면 캐시 히트가 발생해야 한다(Anthropic 네이티브 기준).

### 6.4 응답 usage 객체 확인 필드

```json
{
  "usage": {
    "input_tokens":                   <-- 비캐시 입력 토큰
    "cache_creation_input_tokens":    <-- 0이면 캐싱 미지원 또는 최솟값 미달
    "cache_read_input_tokens":        <-- 0이면 캐시 히트 없음
    "output_tokens":                  <-- 정상
    "cache_write_input_tokens":       <-- 일부 SDK에서 사용하는 별칭
  }
}
```

`cache_creation_input_tokens`과 `cache_read_input_tokens`이 모두 `0`이거나 응답 JSON에 **존재하지 않으면** Z.AI 엔드포인트가 캐싱을 처리하지 않는다고 판단한다.

---

## 7. 권장 정책

### 옵션 A: 현행 유지 (권장)

두 플래그(`DISABLE_PROMPT_CACHING=1`, `CLAUDE_CODE_DISABLE_EXPERIMENTAL_BETAS=1`)를 모든 `moai glm` 경로에서 유지한다.

**근거**:
1. Z.AI 공식 문서에 캐싱 지원이 명시되지 않았다. 지원 의사를 확인할 수 없다.
2. `cache_read_input_tokens` 실측 사례가 공개적으로 없다. 캐싱이 동작한다는 증거가 없다.
3. LiteLLM 오픈소스 생태계에서도 Z.AI/GLM에 `cache_control`을 전송하면 스트리핑하거나 무시하는 동작이 관측됐다(Issue #19923).
4. `DISABLE_EXPERIMENTAL_BETAS`가 효과적으로 동작하지 않는 회귀 패턴이 2026년 내 3회 관측됐다. GLM 사용자 입장에서 베타 헤더 전송은 Z.AI의 미구현 기능을 호출하는 것으로, 오류 또는 정의되지 않은 동작으로 이어진다.
5. `DISABLE_PROMPT_CACHING=1`의 주된 이유(전체 시스템 프롬프트 재전송 → 컨텍스트 소진)는 캐싱 지원 여부와 무관하게 유효하다. 캐싱이 지원되지 않는 한 이 플래그는 올바른 보호 역할을 한다.

**한계**: `CLAUDE_CODE_DISABLE_EXPERIMENTAL_BETAS` 우회 회귀가 반복되므로, Claude Code 신규 버전 업데이트 시 베타 헤더 전송 여부를 주기적으로 확인해야 한다.

---

### 옵션 B: opt-in 환경변수 (`MOAI_GLM_ENABLE_CACHING=1`)

사용자가 실 API로 §6 절차를 직접 확인한 후 선택적으로 캐싱을 활성화하는 경로를 제공한다.

**구현 예시**:
```go
// setGLMEnv 내에서
if os.Getenv("MOAI_GLM_ENABLE_CACHING") != "1" {
    _ = os.Setenv("DISABLE_PROMPT_CACHING", "1")
}
```

**장점**: 실제로 Z.AI가 캐싱을 지원하는 것이 확인되면 전환 비용 없이 활성화 가능.  
**단점**: 검증 절차 없이 활성화하면 오류 또는 비용 증가 위험. 현재 공개 증거 없음.

현재 시점에서는 **대기** 권고. Z.AI 측의 공식 문서화 또는 커뮤니티 실측 보고가 나왔을 때 검토.

---

### 옵션 C: 모델별 분기

GLM-4.7은 베타 헤더에 덜 민감하다는 30일 전 기록을 근거로 모델별로 분기한다.

**장점**: 이론적으로 일부 모델에서 캐싱 시도 가능.  
**단점**:
1. 30일 전 메모리는 stale이다. Z.AI 프록시 업데이트로 현재 동작이 달라졌을 수 있다.
2. 캐싱이 실제로 동작한다는 확인이 없으므로 모델별 분기는 위험 대비 이득이 불명확하다.
3. `CLAUDE_CODE_DISABLE_EXPERIMENTAL_BETAS` 우회 회귀 패턴으로 인해 특정 모델에서 베타 헤더를 허용해도 Claude Code 자체에서 예상치 않은 베타 플래그를 전송할 수 있다.

현재 시점에서 **권장하지 않음**.

---

### 권장 결론

**옵션 A 유지**. 근거가 불충분한 상태에서 현행 보호 플래그를 변경하는 것은 정의되지 않은 동작으로 이어질 가능성이 높다.

**SPEC-GLM-001 amendment 필요 여부**: 이 보고서의 조사 결과를 `SPEC-GLM-001` 코드 주석(`glm.go:142–147`) 에 보강하는 것은 고려할 수 있다. 그러나 플래그 제거나 opt-in 로직 추가는 §8의 open question이 해소된 후에 별도 SPEC으로 진행할 것을 권장한다.

---

## 8. Open Questions

이 보고서에서 검증되지 않은 항목 및 사용자가 실 API로 확인해야 할 항목:

1. **[검증 필요] Z.AI가 `cache_control: { type: "ephemeral" }` 필드를 수신하면 어떻게 처리하는가?**  
   - 무시(silently ignore)하는가, 오류를 반환하는가, 아니면 실제로 캐싱하는가?  
   - 확인 방법: §6.2 단계 수행 후 응답 usage 필드 확인.

2. **[검증 필요] Z.AI가 `anthropic-beta: prompt-caching-2024-07-31` 헤더를 수신하면 400 오류를 반환하는가?**  
   - 30일 전 GLM-5.1에서는 Error 1214/400이 발생했다는 기록이 있으나 stale.  
   - 확인 방법: §6.1 단계 수행.

3. **[검증 필요] Z.AI의 "GLM-5 네이티브 캐싱"(openclaw #24497에서 언급)이 Anthropic `cache_control` 방식과 동일한 메커니즘인가, 아니면 자체 컨텍스트 캐싱인가?**  
   - Z.AI 공식 문서나 Zhipu AI 개발자 문서에서 확인 필요.

4. **[검증 필요] Claude Code v2.1.119–121 이후 `CLAUDE_CODE_DISABLE_EXPERIMENTAL_BETAS=1` 설정 시 `advanced-tool-use-2025-11-20` 베타 플래그가 실제로 억제되는가?**  
   - 최신 CC 버전(현재 v2.1.119–122 기준) + Z.AI 엔드포인트 조합에서 서버 측 거부 오류 발생 여부 확인.

5. **[검증 필요] `DISABLE_PROMPT_CACHING=1` 없이 `moai glm`을 실행할 경우 컨텍스트 소진이 얼마나 빨리 발생하는가?**  
   - glm.go 주석의 "30-40K 토큰/요청" 추정치가 현재 모델(GLM-5.1, GLM-5-Turbo)에서도 동일하게 적용되는지 확인.

6. **[검증 필요] Z.AI의 Anthropic 호환 엔드포인트가 Anthropic의 공식 메시지 API 변경(베타 헤더 불필요 → `cache_control` 자동 처리)을 반영하는가?**  
   - Anthropic은 `prompt-caching-2024-07-31` 베타 헤더 없이도 `cache_control` 필드만으로 캐싱 동작함을 공식화했다. Z.AI 프록시가 이를 패스스루하는지 확인 필요.

---

## 9. References

| 출처 | URL | 날짜 |
|------|-----|------|
| Z.AI 개발자 문서 FAQ | https://docs.z.ai/devpack/faq | 2026-05-01 조회 |
| Z.AI 빠른 시작 | https://docs.z.ai/devpack/quick-start | 2026-05-01 조회 |
| Z.AI Claude Code 통합 가이드 | https://docs.z.ai/devpack/tool/claude | 2026-05-01 조회 |
| LiteLLM Issue #19923: GLM cache_control 스트리핑 버그 | https://github.com/BerriAI/litellm/issues/19923 | 2026-01-28 |
| openclaw Issue #24497: ZAI GLM-5 cache-ttl 지원 요청 | https://github.com/openclaw/openclaw/issues/24497 | 2026-02-23 |
| Roo-Code Issue #10601: zai-glm-4.7 캐싱 활성화 요청 | https://github.com/RooCodeInc/Roo-Code/issues/10601 | 2026-01-10 |
| anthropics/claude-code Issue #22893: DISABLE_BETAS 미동작 | https://github.com/anthropics/claude-code/issues/22893 | 2026-02 |
| anthropics/claude-code Issue #30926: v2.1.69 beta 플래그 회귀 | https://github.com/anthropics/claude-code/issues/30926 | 2026-03 |
| anthropics/claude-code Issue #20031: beta=true 쿼리 파라미터 잔류 | https://github.com/anthropics/claude-code/issues/20031 | 2026-01 |
| anthropics/claude-code Issue #49648: Bedrock Opus 4.7 beta 거부 | https://github.com/anthropics/claude-code/issues/49648 | 2026-04 |
| APIyi: Claude Code 환경변수 가이드 | https://help.apiyi.com/en/claude-code-environment-variables-complete-guide-en.html | — |
| Anthropic 공식 프롬프트 캐싱 문서 | https://platform.claude.com/docs/en/build-with-claude/prompt-caching | — |
| moai-adk-go 소스: glm.go | internal/cli/glm.go:142–177, 370–384, 546–548, 826–829, 870–873 | — |
| moai-adk-go 소스: session_start.go | internal/hook/session_start.go:238–246 | — |
| moai-adk-go 소스: launcher.go | internal/cli/launcher.go:235–243 | — |
| moai-adk-go 소스: auth/glm.go | internal/github/auth/glm.go:34–39 | — |
| moai-adk-go 테스트: launcher_test.go | internal/cli/launcher_test.go:745–746 | — |
| moai-adk-go 테스트: glm_test.go | internal/cli/glm_test.go:153–154 | — |

---

*이 보고서는 코드 변경을 포함하지 않으며, 조사 전용입니다.*  
*공개 출처 검색 기준일: 2026-05-01*
