# t649 F1·F2 수정 재감사

SPEC: SPEC-MOAI-GATEWAY-001  
Overall Verdict: **PASS — F1·F2 변경분 한정**  
범위 점수: **2/2 수정 확인**. 전체 SPEC의 네 차원 가중 점수는 **UNVERIFIED**이며 이 판정으로 대체하지 않는다.

## Claim

- **F1 [High, blocking → 해결 확인, 확신 높음]** 원래 실패했던 두 overlay 재현이 PASS다. reasoning → commentary message → reasoning → final message 순서 및 각각의 phase가 유지되며, 두 message 뒤 reasoning 변형도 다음 입력으로 복원된다. 현재 추가 회귀는 같은 message의 여러 text part, 명시 null phase 및 phase 부재를 구분한다.
- **F2 [Medium, blocking → 해결 확인, 확신 높음]** 원래 실패했던 adapter/TLS 송신 재현이 PASS다. 최종 AuthPKCE 송신 본문에서 max_output_tokens가 빠지고 stream:true가 추가된 뒤에도 reasoning item의 원문 공백을 포함한 바이트가 보존된다.
- **권한 보존 대조군 PASS:** 실제 receipt.Store에 게시한 carrier의 phase만 변경하면 거절한다. 기존 계정·모델 계열·stripping·닫힌 권한 거절과 승인된 fork 대조군도 PASS다. 이는 권한 전체 감사 점수를 의미하지 않는다.
- **신규 blocking 발견 없음 — 조사 범위 한정.** 관리 설정, reasoning 없는 출력 및 Windows 관련 기존 Gap은 아래에 유지한다.

코드 확인: opaque/codec.go의 Decode는 접두사와 version 일치를 요구하고 V2에는 비어 있지 않은 public_items를 요구한다. EncodeWithPublic은 위치 중복·빈 part·유효하지 않은 phase를 거절하며, Decode의 canonical 재직렬화 비교는 알려지지 않은 필드·대체 표현을 거절한다. receipt_history.go는 envelope 전체 digest를 관측에 담고 공개 prefix를 launcher family·credential scope·검증된 model family에 결합한다. openai.go의 marshalRawObject는 json.Valid로 검사한 원시 값을 그대로 기록한다.

## Evidence

실행 1 — 원본 overlay 그대로:
```sh
go test -overlay .moai/reports/t649/independent-evidence/overlay.json ./internal/gateway/translate ./internal/gateway -run '^(TestAuditReasoning|TestAuditNativeEgress)' -count=1 -v -timeout=60s
```
```text
# github.com/modu-ai/moai-adk/internal/gateway [github.com/modu-ai/moai-adk/internal/gateway.test]
./.moai/reports/t649/independent-evidence/audit_gateway_test.go:3:6: auditHistory redeclared in this block
	internal/gateway/replay_egress_test.go:14:6: other declaration of auditHistory
./.moai/reports/t649/independent-evidence/audit_gateway_test.go:4:19: method auditHistory.Check already declared at internal/gateway/replay_egress_test.go:16:21
./.moai/reports/t649/independent-evidence/audit_gateway_test.go:5:19: method auditHistory.Publish already declared at internal/gateway/replay_egress_test.go:17:21
=== RUN   TestAuditReasoningPublicItemOrderAndPhase
    t649_audit_overlay_test.go:15: observed upstream replay: {"input":[{"content":[{"text":"hello","type":"input_text"}],"role":"user"},{"encrypted_content":"cipher_rs_1","id":"rs_1","summary":[],"type":"reasoning"},{"content":[{"text":"commentary","type":"output_text"}],"phase":"commentary","role":"assistant"},{"encrypted_content":"cipher_rs_2","id":"rs_2","summary":[],"type":"reasoning"},{"content":[{"text":"final","type":"output_text"}],"phase":"final_answer","role":"assistant"},{"content":[{"text":"continue","type":"input_text"}],"role":"user"}],"max_output_tokens":100,"model":"gpt-5.6-sol","store":false}
--- PASS: TestAuditReasoningPublicItemOrderAndPhase (0.00s)
=== RUN   TestAuditReasoningAfterTwoMessages
--- PASS: TestAuditReasoningAfterTwoMessages (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/gateway/translate	0.388s
FAIL	github.com/modu-ai/moai-adk/internal/gateway [build failed]
FAIL

```

위 실행은 F1 PASS이나 F2 패키지 빌드 실패다. 구현자가 회귀 시험에 원본과 동일한 auditHistory 이름을 추가하여 충돌했다. 이를 생산 결함 또는 PASS로 해석하지 않았다. 원본 F2의 assertion/body는 그대로 두고 증거 디렉터리 복사본의 helper 이름만 auditHistory → reauditHistory로 바꾸었다. 원본 증거와 소스 트리는 수정하지 않았다.

실행 2 — helper 이름만 다른 원본 재현:
```sh
go test -overlay .moai/reports/t649/reaudit-evidence/overlay.json ./internal/gateway/translate ./internal/gateway -run '^(TestAuditReasoning|TestAuditNativeEgress)' -count=1 -v -timeout=60s
```
```text
=== RUN   TestAuditReasoningPublicItemOrderAndPhase
    t649_audit_overlay_test.go:15: observed upstream replay: {"input":[{"content":[{"text":"hello","type":"input_text"}],"role":"user"},{"encrypted_content":"cipher_rs_1","id":"rs_1","summary":[],"type":"reasoning"},{"content":[{"text":"commentary","type":"output_text"}],"phase":"commentary","role":"assistant"},{"encrypted_content":"cipher_rs_2","id":"rs_2","summary":[],"type":"reasoning"},{"content":[{"text":"final","type":"output_text"}],"phase":"final_answer","role":"assistant"},{"content":[{"text":"continue","type":"input_text"}],"role":"user"}],"max_output_tokens":100,"model":"gpt-5.6-sol","store":false}
--- PASS: TestAuditReasoningPublicItemOrderAndPhase (0.00s)
=== RUN   TestAuditReasoningAfterTwoMessages
--- PASS: TestAuditReasoningAfterTwoMessages (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/gateway/translate	0.181s
=== RUN   TestAuditNativeEgressPreservesOpaqueItemBytes
    t649_audit_overlay_test.go:17: observed subscription request: {"input":[{"content":[{"text":"hello","type":"input_text"}],"role":"user"},{ "type":"reasoning", "id":"rs_audit", "summary":[], "encrypted_content":"synthetic_cipher" },{"content":[{"text":"answer","type":"output_text"}],"role":"assistant"},{"content":[{"text":"continue","type":"input_text"}],"role":"user"}],"model":"gpt-5.6-sol","store":false,"stream":true}
--- PASS: TestAuditNativeEgressPreservesOpaqueItemBytes (0.05s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/gateway	0.484s

exit_code=0
```

실행 3 — 현재 경계·phase·receipt·V2 회귀:
```sh
go test ./internal/gateway/translate ./internal/gateway/opaque -run '^(TestRegressionReasoning|TestReasoning|TestStrictJSONAndReasoningSubfields|TestPublicLayoutV2CanonicalAndBounded|TestReceiptHistory)' -count=1 -v -timeout=60s
```
```text
=== RUN   TestReasoningRoundTripUsesFinalEncryptedItem
--- PASS: TestReasoningRoundTripUsesFinalEncryptedItem (0.00s)
=== RUN   TestReasoningBufferedStreamUsesNativeDeltas
--- PASS: TestReasoningBufferedStreamUsesNativeDeltas (0.00s)
=== RUN   TestReasoningToolRoundTripAndPublicationFailure
--- PASS: TestReasoningToolRoundTripAndPublicationFailure (0.00s)
=== RUN   TestReasoningCarrierPreservesOriginalItemBytes
--- PASS: TestReasoningCarrierPreservesOriginalItemBytes (0.00s)
=== RUN   TestReasoningWithoutPublicCompletionFailsClosed
--- PASS: TestReasoningWithoutPublicCompletionFailsClosed (0.00s)
=== RUN   TestReceiptHistoryRejectsStrippingForeignScopeAndFamily
--- PASS: TestReceiptHistoryRejectsStrippingForeignScopeAndFamily (0.05s)
=== RUN   TestReceiptHistoryAuthorizedForkRetainsScope
--- PASS: TestReceiptHistoryAuthorizedForkRetainsScope (0.04s)
=== RUN   TestReceiptHistoryRejectsMalformedAndClosedAuthority
--- PASS: TestReceiptHistoryRejectsMalformedAndClosedAuthority (0.03s)
=== RUN   TestReceiptHistoryRejectsPublicLayoutMutation
--- PASS: TestReceiptHistoryRejectsPublicLayoutMutation (0.02s)
=== RUN   TestRegressionReasoningPublicItemOrderAndPhase
    replay_boundaries_test.go:34: observed upstream replay: {"input":[{"content":[{"text":"hello","type":"input_text"}],"role":"user"},{"encrypted_content":"cipher_rs_1","id":"rs_1","summary":[],"type":"reasoning"},{"content":[{"text":"commentary","type":"output_text"}],"phase":"commentary","role":"assistant"},{"encrypted_content":"cipher_rs_2","id":"rs_2","summary":[],"type":"reasoning"},{"content":[{"text":"final","type":"output_text"}],"phase":"final_answer","role":"assistant"},{"content":[{"text":"continue","type":"input_text"}],"role":"user"}],"max_output_tokens":100,"model":"gpt-5.6-sol","store":false}
--- PASS: TestRegressionReasoningPublicItemOrderAndPhase (0.00s)
=== RUN   TestRegressionReasoningAfterTwoMessages
--- PASS: TestRegressionReasoningAfterTwoMessages (0.00s)
=== RUN   TestRegressionReasoningMultipartAndNullPhase
--- PASS: TestRegressionReasoningMultipartAndNullPhase (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/gateway/translate	0.502s
=== RUN   TestReasoningShapesAndOrder
--- PASS: TestReasoningShapesAndOrder (0.00s)
=== RUN   TestStrictJSONAndReasoningSubfields
--- PASS: TestStrictJSONAndReasoningSubfields (0.00s)
=== RUN   TestPublicLayoutV2CanonicalAndBounded
--- PASS: TestPublicLayoutV2CanonicalAndBounded (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/gateway/opaque	0.522s

exit_code=0
```

## Baseline-attribution

측정 위치: /Users/goos/MoAI/moai-adk-go/.claude/worktrees/moai-proxy-unified  
이번 실행의 세션 조회:
```text
moai session current
01a090e5-0bef-7de3-b680-18be96c58267
git branch --show-current
WT-unified-gateway
git rev-parse HEAD
81c1d58f9cf7045594ee61d5e4ff380948ce9eba
```
기존 미커밋 구현을 포함한 디스크 트리의 결과다. 관련 파일의 이번 SHA256:
```text
36281d08763434acfd2c442497cb268b1039b60fd524ad0b3c4aa043ecc1b55c  internal/gateway/opaque/codec.go
7baf9dc7258ba2d59b280f2618dff4ecdde99516eef1ff5c7c43367dff16beda  internal/gateway/openai.go
bec5be01b66254d70cbce32754f057c85f36ad2e23c492ebb8545ee691e06865  internal/gateway/translate/request.go
dccd3c5291e45fcdf844d86098d50125fbb335c70790745dc354b153f1817497  internal/gateway/translate/response.go
d03a02457bdffcbaaa25cfe598b7f7e3aa3a73e51853e17d08dd22965528f746  internal/gateway/translate/receipt_history.go
```
생산 코드·기존 시험·SPEC·큐·git 상태를 수정하지 않았다. 보고서와 재실행 overlay 증거만 추가했다.

## Gaps

- G1 managed effective modelPicker 충돌 preflight는 이번 재감사에서 구현·실행 확인하지 않았다. 기존 인수 Gap은 유지한다.
- reasoning이 없는 응답의 여러 public item 경계·phase 복원은 이번 F1 변경분 판정에 포함하지 않았다.
- Windows GitHub CI, 전체 SPEC AC, 전체 변경 coverage 85%, 취약 의존성 전수 검사는 미실행이다. 따라서 전체 Functionality/Security/Craft/Consistency 점수는 부여하지 않는다.
- 실제 계정 요청은 실행하지 않았다. Root의 GPT6 답변·Read·새 프로세스 resume 관측과 다른 담당의 실제 picker 결과는 그 담당 증거로 판단해야 한다.
- 원본 F1/F2의 HistoryAuthority stub은 replay 변환을 분리하는 목적이다. 실제 receipt 권한 변조 시험은 별도로 실행했지만 두 재현을 production store로 합친 모든 조합을 전수 검증하지 않았다.

## Residual-risk

이 PASS는 원래 F1/F2 실패를 재현한 두 입력 경계 사례와 최종 송신 바이트 사례, 관련 현재 회귀에서 수정을 확인한 결과다. 실제 서버의 모든 reasoning 형태·native transcript의 모든 편집·managed source 우선순위까지 보장하지 않는다. 로컬 rc.8 시험 설치를 위한 이 두 수리 게이트는 해소되었으나 전체 카드 완료나 릴리스 인수 PASS로 확대하면 안 된다.

## Iteration history

1. independent-review.md: F1 두 재현 FAIL, F2 실제 adapter 송신 바이트 재현 FAIL.
2. 이번 재감사: 원본 overlay F1 두 개 PASS; F2 helper 충돌로 첫 빌드 실패.
3. helper 이름만 바꾼 동일 F2 assertion PASS; 원본 F1 재실행 PASS; 현재 15개 경계·reasoning·authority·strict V2 회귀 PASS.

