# SPEC 0.11.0 적용 근거

## Claim

spec.md·plan.md·design.md·acceptance.md에 App Server 전환을 반영했다. 버전0.11.0,
REQ24/묘비2·AC24/묘비1 유지. t650~t654 16개 하위 시나리오를 기존 AC에 연결했다.
progress.md·제품 코드·git 통합은 수행하지 않았다.

## Evidence

명령: `go run ./cmd/moai spec lint SPEC-MOAI-GATEWAY-001`

```text
✓ No findings — all SPEC documents are valid
```

기본 ID 정규식 집합 및 RETIRED 차집합 출력:

```text
spec.md active 24 retired 2
acceptance.md active 24 retired 1
```

별도 실행 담당의 isolation-result.json을 읽어 대조한 출력:

```text
probe record: 0.154.0 0 1 ['MOAI_APPSERVER_ECHO_OK'] 1
```

순서는 Codex 버전, errors 수, tool_calls 수, final_texts, MCP startup event 수다.
이 문서 작성자는 live probe를 실행하지 않았고 기록을 읽었다.

## Baseline-attribution

2026-09-12, WT-unified-gateway / 81c1d58f9 / moai-proxy-unified 및 기존 미커밋 작업.
현 소스 go run으로 lint를 실행했다. 공식 기준 https://learn.chatgpt.com/docs/app-server 접근2026-09-12와
설치0.154.0 schema 차이를 명시했다. dynamicTools resume 문구를 설치 capability로 추정하지 않는다.

## Gaps

lint는 구조 검사다. 독립 의미 감사, 실제 Claude 도구 실행권 음성, 전체 deferred schema 캡처,
HTTP간 RPC·crash 복구·resume·fork·compact·managed/API·Windows의 모든 실제 제품 증거는 아직 미완료다.
단일 echo 기록에 MCP startup event1건이 남으므로 완전 격리 PASS가 아니다.

## Residual-risk

실험 API 버전 차이와 도구 schema 고정 목록이 핵심 위험이다. native schema N개를 dispatcher1개로 줄이는
대안을 묵시 채택하지 않으며 실패 때 지원 범위를 낮추어 완료로 표시하지 않는다.
과거 HISTORY·결정은 당시 근거로 유지했고 현재 직접 backend·token/opaque 복원 설계는 App Server로 대체했다.
