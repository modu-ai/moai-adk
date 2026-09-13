# t649 독립 의미 검토

SPEC: SPEC-MOAI-GATEWAY-001 0.10.0  
판정: **FAIL — F1·F2 blocking**  
범위: reasoning/receipt 왕복·제공자 경계의 한정 검토. 전체 SPEC 점수는 **UNVERIFIED**다. 네 차원 전체 점수나 보안 전수 PASS를 부여하지 않는다.

## Claim

- **F1 [High] [blocking] [확신 높음]** `internal/gateway/translate/response.go:93`, `:168` 및 `internal/gateway/translate/request.go:310`, `:417`: 공개 message item의 경계를 잃은 뒤 reasoning의 원래 output index로 재삽입한다. 합성 응답 `reasoning → commentary message → reasoning → final message`는 첫 응답 성공 후 다음 요청에서 `reasoning → [commentary+final 단일 message] → reasoning`으로 바뀐다. 두 공개 message 다음에 reasoning이 있는 변형은 첫 응답 성공 후 재개 요청을 `reasoning output position mismatch`로 거절한다. 이는 AC-MG-009의 원래 순서 보존을 위반한다. 관측된 phase 손실도 기록하되, design의 기존 phase 생략 조항과 충돌하므로 이를 별도 독립 차단 근거로 삼지는 않았다. **필수 수정:** carrier/receipt에 원래 공개 item 경계와 reasoning의 상대 위치를 복원할 충분한 정보를 결합하고, 같은 message 내부 여러 text part와 서로 다른 message를 구별하여 후속 요청을 복원한다. 공개 metadata/phase 보존 계약도 기존 생략 조항과 일치시킨다.
- **F2 [Medium] [blocking] [확신 높음]** `internal/gateway/openai.go:132` 및 `:143`: translator가 복원한 reasoning item 원문을 AuthPKCE 전송 준비에서 `json.Marshal`하여 공백을 없앤다. 로컬 TLS upstream에서 실제 adapter 송신 본문을 캡처해 item 전체 바이트 불일치를 재현했다. **암호문 문자열 자체는 같고 JSON 의미도 같다.** 복호화 실패·서비스 장애를 주장하지 않는다. 차단 근거는 AC-MG-009와 REQ-MG-015의 item 원문 바이트 보존 계약이다. **필수 수정:** 최종 upstream 송신 단계까지 원문을 보존하며, max_output_tokens 제거와 stream 강제 이후의 raw body를 대조한다.
- **G1 [High 추정] [인수 차단 Gap] [미검증]** `internal/cli/gateway_session.go:98`: modelPicker/availableModels overlay는 생성되지만 managed 설정과 실제 effective 설정 충돌을 생산 시작 전에 판정한 실행 증거가 없다. 구현 인계도 미구현으로 명시한다. AC-MG-001/003은 managed 충돌·빈 목록의 명시 실패를 요구한다. 이번 검토는 managed UI를 직접 실행하지 않았으므로 실제 외부 제공자 송신이나 정책 우회를 입증한 결함으로 분류하지 않는다. 서버 catalog 경계의 단위·통합 대조군은 아래와 같이 PASS다.

F1·F2는 현재 트리에서 재현한 별개 결함이다. 어느 것도 root의 실제 GPT HTTP 400 원인으로 입증되지 않았다. 실제 계정 E2E 결과는 담당자의 보고서와 분리한다.

## Evidence

재현 파일은 `independent-evidence/audit_test.go`, `audit_gateway_test.go`, `overlay.json`에 있다. 소스 트리에 시험 파일을 넣지 않고 Go overlay로 실행했다. 합성 암호문과 로컬 TLS 서버만 사용했다.

F1 실행:
```sh
go test -overlay /tmp/t649-independent-review/overlay.json ./internal/gateway/translate -run '^TestAuditReasoning' -count=1 -v -timeout=45s
```
```text
=== RUN   TestAuditReasoningPublicItemOrderAndPhase
    t649_audit_overlay_test.go:15: observed upstream replay: {"input":[{"content":[{"text":"hello","type":"input_text"}],"role":"user"},{"encrypted_content":"cipher_rs_1","id":"rs_1","summary":[],"type":"reasoning"},{"content":[{"text":"commentary","type":"output_text"},{"text":"final","type":"output_text"}],"role":"assistant"},{"encrypted_content":"cipher_rs_2","id":"rs_2","summary":[],"type":"reasoning"},{"content":[{"text":"continue","type":"input_text"}],"role":"user"}],"max_output_tokens":100,"model":"gpt-5.6-sol","store":false}
    t649_audit_overlay_test.go:16: public output item boundaries lost: got 5 total input items, want 6
--- FAIL: TestAuditReasoningPublicItemOrderAndPhase (0.00s)
=== RUN   TestAuditReasoningAfterTwoMessages
    t649_audit_overlay_test.go:26: previous successful output cannot be replayed: reasoning output position mismatch
--- FAIL: TestAuditReasoningAfterTwoMessages (0.00s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/gateway/translate	0.504s
FAIL
```

F2 실행:
```sh
go test -overlay /tmp/t649-independent-review/overlay.json ./internal/gateway -run '^TestAuditNativeEgress' -count=1 -v -timeout=45s
```
```text
=== RUN   TestAuditNativeEgressPreservesOpaqueItemBytes
    t649_audit_overlay_test.go:17: observed subscription request: {"input":[{"content":[{"text":"hello","type":"input_text"}],"role":"user"},{"type":"reasoning","id":"rs_audit","summary":[],"encrypted_content":"synthetic_cipher"},{"content":[{"text":"answer","type":"output_text"}],"role":"assistant"},{"content":[{"text":"continue","type":"input_text"}],"role":"user"}],"model":"gpt-5.6-sol","store":false,"stream":true}
    t649_audit_overlay_test.go:18: opaque item original bytes lost at subscription egress
--- FAIL: TestAuditNativeEgressPreservesOpaqueItemBytes (0.02s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/gateway	0.423s
FAIL
```

최초 실행 결과는 두 log의 `FAIL`이다. 실행 wrapper가 이어서 `cat`을 실행한 탓에 wrapper 종료 코드는 0이었으나, 이를 Go 시험 PASS로 해석하지 않는다. 보존된 overlay로 재현할 때는 `-overlay .moai/reports/t649/independent-evidence/overlay.json`을 사용한다.

보존/차단 대조군 실행:
```sh
go test ./internal/gateway/translate ./internal/gateway/auth ./internal/gateway ./internal/cli -run '^(TestReceiptHistoryRejectsStrippingForeignScopeAndFamily|TestReceiptHistoryAuthorizedForkRetainsScope|TestReplayScopeSurvivesRefreshButRejectsReplacement|TestNativePolicyWireAndProviderIsolation|TestGatewayProviderContractRejectsForeignModel|TestGatewayProviderContractPickerAndSlots|TestNativeProviderProductionBindingKeepsCatalogAndFamilySeparate)$' -count=1 -v -timeout=60s
```
```text
=== RUN   TestReceiptHistoryRejectsStrippingForeignScopeAndFamily
--- PASS: TestReceiptHistoryRejectsStrippingForeignScopeAndFamily (0.04s)
=== RUN   TestReceiptHistoryAuthorizedForkRetainsScope
--- PASS: TestReceiptHistoryAuthorizedForkRetainsScope (0.04s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/gateway/translate	0.418s
=== RUN   TestReplayScopeSurvivesRefreshButRejectsReplacement
--- PASS: TestReplayScopeSurvivesRefreshButRejectsReplacement (0.04s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/gateway/auth	0.671s
=== RUN   TestNativePolicyWireAndProviderIsolation
--- PASS: TestNativePolicyWireAndProviderIsolation (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/gateway	0.962s
=== RUN   TestGatewayProviderContractRejectsForeignModel
--- PASS: TestGatewayProviderContractRejectsForeignModel (0.00s)
=== RUN   TestGatewayProviderContractPickerAndSlots
--- PASS: TestGatewayProviderContractPickerAndSlots (0.00s)
=== RUN   TestNativeProviderProductionBindingKeepsCatalogAndFamilySeparate
--- PASS: TestNativeProviderProductionBindingKeepsCatalogAndFamilySeparate (0.04s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	1.232s
```

## Baseline-attribution

```text
git rev-parse --short HEAD
81c1d58f9
git branch --show-current
WT-unified-gateway
```

실행 위치는 `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/moai-proxy-unified`다. HEAD만의 결과가 아니라 기존 미커밋 구현을 포함한 디스크 트리에서 실행했다. 관련 파일 SHA-256은 `independent-evidence/baseline.json`에 보존했다. 생산 코드·기존 시험·SPEC·큐·브랜치는 수정하지 않았다.

## Gaps

- 실제 GPT/Claude/GLM 계정 요청·Claude PTY·도구·resume/fork E2E는 이 검토가 실행하지 않았다.
- F1 테스트의 HistoryAuthority는 허용 stub이다. 목적은 receipt 보안 전체 검증이 아니라, 정상 승인된 응답의 변환·다음 입력 복원을 분리해 판정하는 것이다. 실 receipt의 계정/계열/변조 거절은 별도 대조군만 실행했다.
- F2는 실제 OpenAIAdapter와 합성 자격증명 저장소·로컬 TLS upstream을 사용하지만, 공개 인터넷 송신은 없다.
- managed server/MDM/registry/file 출처 충돌, effective context, 전체 MCP schema 동작, Windows GitHub CI는 미실행이다.
- 전체 SPEC AC matrix, 취약 의존성 감사, 전체 변경 coverage와 lint는 이번 한정 검토 밖이다. 85% 충족이나 전체 네 차원 PASS를 주장하지 않는다.

## Residual-risk

단일 reasoning+단일 text 또는 단일 tool 응답의 실사용 성공만으로 F1의 여러 공개 item 경우를 덮을 수 없다. F2의 공백 차이는 실제 서버에서 의미상 허용될 수 있으나 현재 엄격한 원문 계약의 증거는 충족하지 못한다. 수정 후 이 두 재현군과 기존 대조군을 다시 실행하고 실제 E2E 실패 원인을 별도로 좁혀야 한다.

## Iteration history

1. 이번 독립 검토: F1 두 재현 FAIL, F2 송신 바이트 재현 FAIL, 범위 한정 대조군 7개 PASS. 변경 전 시험 증거이며 수정 후 재감사는 아직 없다.
