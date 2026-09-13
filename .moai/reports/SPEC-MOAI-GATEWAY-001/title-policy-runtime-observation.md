# GPT 제목 JSON 형식 실제 관측

## Claim

사용자가 구독 서버 출력 정책을 승인한 뒤, GPT 네 모델에 각각 한 번씩 synthetic title JSON schema 요청을 보냈다. 모두 HTTP 200, SSE completed, title:string 한 필드만 가진 유효 JSON이었다. 이는 구독 endpoint의 해당 형식 수용 관측이며 Claude→gateway 변환이나 전체 제품 PASS가 아니다.

요청은 reasoning.effort=high, text.format.type=json_schema/name=moai_title_preflight/strict=true, schema={type:object,properties:{title:{type:string}},required:[title],additionalProperties:false}, stream=true/store=false였다. max_output_tokens는 생략했고 실제 생성 token 상한을 보장하지 않는다. 외부 도구 실행 없이 합성 읽기 목록 대화만 사용했다.

## Evidence

실행 명령:
```text
unset ANTHROPIC_BASE_URL ANTHROPIC_API_KEY ANTHROPIC_AUTH_TOKEN Z_AI_API_KEY && MOAI_TITLE_PREFLIGHT_LIVE=1 GOCACHE=/tmp/gateway-policy-cache go test -overlay /var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/gateway-title-probe-54r5x493/overlay.json ./.moai/state/verify/01a08e7b-6aa0-7361-ab7e-ea8da1f02228/auth-live -run '^TestTitlePreflightLive$' -v -count=1 -timeout 230s
```
원문 끝부분:
```text
--- PASS: TestTitlePreflightLive (9.82s)
PASS
ok  github.com/modu-ai/moai-adk/.moai/state/verify/01a08e7b-6aa0-7361-ab7e-ea8da1f02228/auth-live 10.131s
```
각 모델 출력에서 http_status=200, diagnostic.sse_parse_completed=true, title_schema_valid=true를 관측했다. 기존 관측기 result.completed=false는 Content-Type 헤더 부재 때문에 일반 media gate의 PASS를 주지 않는다는 뜻이다. 이번에는 별도 SSE terminal과 실제 JSON schema 판정을 검사했다. 네 응답 모두 Content-Type 값 0개, UTF-8=true였다. 요청은 정확히 네 개이고 재시도하지 않았다.

별도 offline 시험은 payload에 상한 필드가 없고 model/stream/store가 맞는지, 유효 title 및 숫자/추가 필드 음성 판정을 확인했다. live opt-in 없이 TestTitlePreflightLive는 explicit live opt-in required로 SKIP했으며 패키지 exit0/0.388s였다.

## Baseline-attribution

WT-unified-gateway / 81c1d58f9 위 현재 미커밋 AUTH 코드와 전용 auth-live Store. 기존에 새 MoAI scratch 로그인으로 만든 Store만 사용했고 기존 사용자 자격을 복사하지 않았다. 실제 요청 종료 뒤 Store.Close 및 각 HTTP body/transport 정리는 하네스 defer로 수행했다. 새로운 로그인/refresh/logout은 실행하지 않았다.

보존된 하네스/overlay/readback은 .moai/state/verify/01a08e7b-6aa0-7361-ab7e-ea8da1f02228/gateway-entry/title-policy-preflight/ 아래다. 원본 응답은 private responses에 두고 내용을 공개 보고서에 복사하지 않았다. 다음은 독립 파일 readback 결과다.

```json
[
  {
    "model": "gpt-5.6-luna",
    "file": "/Users/goos/MoAI/moai-adk-go/.claude/worktrees/moai-proxy-unified/.moai/state/verify/01a08e7b-6aa0-7361-ab7e-ea8da1f02228/auth-live/responses/gpt-5.6-luna-title-schema-high-548007997.raw",
    "bytes": 9143,
    "sha256": "a55741c12764a02a4a7af366487bc53ee5742285aebf5652780e8d370cacd898",
    "mode": "0o600",
    "completed": true,
    "title_schema_valid": true,
    "capture_utc": "2026-09-11T11:16:38.120564+00:00"
  },
  {
    "model": "gpt-5.6-sol",
    "file": "/Users/goos/MoAI/moai-adk-go/.claude/worktrees/moai-proxy-unified/.moai/state/verify/01a08e7b-6aa0-7361-ab7e-ea8da1f02228/auth-live/responses/gpt-5.6-sol-title-schema-high-4053487312.raw",
    "bytes": 9140,
    "sha256": "a01cce83ad966f736c4c9bfdb60ed8b496f6c4df10bb53000d52a6a280efa1b8",
    "mode": "0o600",
    "completed": true,
    "title_schema_valid": true,
    "capture_utc": "2026-09-11T11:16:33.709733+00:00"
  },
  {
    "model": "gpt-5.6-terra",
    "file": "/Users/goos/MoAI/moai-adk-go/.claude/worktrees/moai-proxy-unified/.moai/state/verify/01a08e7b-6aa0-7361-ab7e-ea8da1f02228/auth-live/responses/gpt-5.6-terra-title-schema-high-3152722342.raw",
    "bytes": 9146,
    "sha256": "e6207c92badc77e6469cf3b0c866afa8a9fe3745f8b03e78b93f5f38fcaec4f3",
    "mode": "0o600",
    "completed": true,
    "title_schema_valid": true,
    "capture_utc": "2026-09-11T11:16:35.898939+00:00"
  },
  {
    "model": "gpt-6-astra",
    "file": "/Users/goos/MoAI/moai-adk-go/.claude/worktrees/moai-proxy-unified/.moai/state/verify/01a08e7b-6aa0-7361-ab7e-ea8da1f02228/auth-live/responses/gpt-6-astra-title-schema-high-1477685935.raw",
    "bytes": 9140,
    "sha256": "6f252f8847765c7967e708293bb6e12602d0fd82edc2379b595937c40944e054",
    "mode": "0o600",
    "completed": true,
    "title_schema_valid": true,
    "capture_utc": "2026-09-11T11:16:31.260313+00:00"
  }
]
```

## Gaps

임의 JSON schema 전체, 다른 effort 값, native adaptive와 동일한 계산량, context_management 의미, 전체 변환/receipt/launcher 연결은 검증하지 않았다. 기존 관측기의 max_output_tokens_support=UNVERIFIED 필드는 이번 요청에서 그 필드를 시험하지 않았기 때문에 그대로 두었다. 앞선 실제 400 관측을 번복하지 않는다.

## Residual-risk

고정한 작은 schema의 수용을 모든 schema 지원으로 확대하지 않는다. 생산 변환기는 명시된 지원 subset과 잘못된 형식의 무송신 대조군을 별도로 구현·검증해야 한다. 200 header와 성공 terminal을 구분하고 byte·취소 상한을 유지한다.
