# Native 요청 정책 연결 관측

## Claim

현재 설치 Claude의 실제 로컬 picker 관측 요청 003~008을 현재 nativeRequest와 Responses Request에 넣으면 두 경계 모두 거절한다. 이는 전체 명령 연결 전에 정책 지원을 구현해야 한다는 실행 근거다. 거절을 제품 성공으로 판정하지 않는다.

## Evidence

임시 Go overlay는 기존 소스를 변경하지 않고 각 raw 요청을 두 함수에 입력했다. 오류 문자열이나 대화 본문은 출력하지 않고 SHA와 거절 여부만 기록했다.

명령:
```text
unset ANTHROPIC_BASE_URL ANTHROPIC_API_KEY ANTHROPIC_AUTH_TOKEN Z_AI_API_KEY && GOCACHE=/tmp/gateway-policy-cache go test -overlay /var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/gateway-native-policy-estcjx9e/overlay.json ./internal/gateway -run '^TestObservedNativePolicyReadiness$' -count=1 -v -timeout 30s
```
원문 출력:
```text
=== RUN   TestObservedNativePolicyReadiness
    observed_policy_readiness_test.go:9: request-003 sha=4406c0852222412720c83f5a44081e2e2c068b5f0fa77f013dc9fe8795c1082b native_rejected=true responses_rejected=true
    observed_policy_readiness_test.go:9: request-004 sha=ad2ef38e71be87d715dc2a76bde98d960b25866d0912b33b2655168e30408cd2 native_rejected=true responses_rejected=true
    observed_policy_readiness_test.go:9: request-005 sha=5bf97c3d4c391b64ef61b5f69acd5c3a3765c5e994c614361a904fef7c20d313 native_rejected=true responses_rejected=true
    observed_policy_readiness_test.go:9: request-006 sha=18b332bd8b0261faac8a8fb486a84b031b5d6dfe4408416ebc9892b578684913 native_rejected=true responses_rejected=true
    observed_policy_readiness_test.go:9: request-007 sha=191d552f2b1f92fbadf03190231a8de7b46fbdf804f7870ae026aa62cf5276ff native_rejected=true responses_rejected=true
    observed_policy_readiness_test.go:9: request-008 sha=803dbbb5c16af10a6dfef41801ae39e087982aecf16a3dd9f7a6c3c7703b40ca native_rejected=true responses_rejected=true
--- PASS: TestObservedNativePolicyReadiness (0.01s)
PASS
ok github.com/modu-ai/moai-adk/internal/gateway 0.405s
```

별도 Python json readback에서 request003은 output_config.effort=high와 title JSON schema, request004는 thinking.type=adaptive, output_config.effort=high, context_management.edits=[{type:clear_thinking_20251015,keep:all}]이었다. request003의 thinking/context_management는 부재이며 null이라는 요청 값을 뜻하지 않는다. 도구 필드 집합은 description/input_schema/name이었다.

## Baseline-attribution

2026-09-11, WT-unified-gateway / 81c1d58f9의 미커밋 트리. source_session_id=01a08e7b-6aa0-7361-ab7e-ea8da1f02228. 원본은 .moai/state/verify/동일 UUID/gateway-entry/picker-followup-20260911T102039Z-26558abd/raw이며 실제 upstream에 보내지 않았다. 현재 gateway 및 translate 요청 검증 함수만 직접 호출했다.

## Gaps

이 시험은 첫 거절 이후의 잠재적인 추가 필드 호환성, 실제 Anthropic/GPT 정책 수용, context 계량의 정확성, receipt 보존을 검증하지 않는다. title schema와 adaptive/effort/keep-all의 정확한 공급자별 계약을 확정·시험해야 한다. 출력 상한의 사용자 결정은 별도 대기이며 이 관측으로 변경하지 않는다.

## Residual-risk

키를 허용 목록에 추가하는 것만으로 의미 보존이 입증되지 않는다. native passthrough와 Responses 변환은 별개다. 공급자별 명시 계약, 정상 입력의 실제 수용, 잘못된 입력의 무송신 대조군과 reasoning 유지 검증이 필요하다.
