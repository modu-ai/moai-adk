# claude-code-model-proxy 분석과 MoAI 적용 기록

## Claim

사용자가 지정한 저장소는 `jclab-joseph/claude-code-model-proxy`다. 동명 ccproxy 후보는 적용 근거에서 제외했다. 분석 커밋은 `112588175eb2b3b693a5bd23d58d8c501e0ae406`이며, 요청 변환·응답 SSE·구독 전송·모델 선택과 해당 시험을 확인했다.

MoAI의 실제 HTTP 400은 이번 안전 진단에서 `output_config.effort=medium`을 로컬 변환기가 거절한 것으로 확인됐다. 지정 소스의 명시 effort 전달 방식을 참고하여 GPT medium을 그대로 전달하도록 보완한 빌드에서 실제 Astra 답변을 확인했다. 이 결과만으로 도구·재개·전체 카드 완료를 주장하지 않는다.

## 소스 대조와 적용

| 경계 | 지정 소스의 처리 | MoAI 적용 판단 |
|---|---|---|
| effort | `internal/router/codex.go:30`에서 모델명 지정, 클라이언트 선택, 모델 기본값 순으로 결정한다. `internal/translate/claude_to_codex.go:470`은 명시 output_config.effort를 읽는다. | 직접 관측한 medium과 기존 high를 그대로 Responses reasoning.effort로 보존한다. 누락을 임의 기본값으로 채우거나 미지원 값을 조용히 낮추는 동작은 이번 수정에 포함하지 않는다. |
| 구독 요청 | `claude_to_codex.go:47`에서 stream=true, store=false, reasoning.encrypted_content include와 instructions 필드를 구성한다. | 기존 stream/store/출력 상한 계약과 대조했다. 동일한 구독 요청 제약을 시험하되, 서버의 출력 정책을 사용한다. |
| reasoning 완료 | `codex_to_claude.go:123`에서 added 값은 잠정 상태로 보고 done 값을 사용한다. 요약이 없어도 암호화된 상태를 signature로 전달한다. | 완료 item을 사용하고 요약 없는 reasoning도 보존해야 한다는 회귀 관점으로 참고한다. MoAI는 기존 opaque/receipt와 계정·family 결합을 유지한다. |
| 도구 스트림 | 같은 파일은 병렬 call별 상태를 저장하고 순차적인 Messages block으로 내보내며 terminal에만 있는 항목도 처리한다. | 이벤트 순서·최종 항목·도구 ID 왕복 시험의 대조군으로 삼는다. |
| 모델 선택 | `cmd/ccmproxy/main.go:504`가 modelPicker 설정을 출력한다. `internal/router/bootstrap.go:14`의 bootstrap 응답 수정은 실험 기능이다. | MoAI의 세션 modelPicker를 유지한다. 사용자 전역 설정 덮어쓰기나 bootstrap 응답 가로채기는 채택하지 않는다. |
| reasoning 소유 | `claude_to_codex.go:183`은 자체 signature marker가 없는 thinking을 건너뛴다. | 이 동작은 차용하지 않는다. MoAI의 세션·계정·receipt 검증과 유실·변조 거절을 유지한다. |

원문 링크: [요청 변환](https://github.com/jclab-joseph/claude-code-model-proxy/blob/112588175eb2b3b693a5bd23d58d8c501e0ae406/internal/translate/claude_to_codex.go), [응답 변환](https://github.com/jclab-joseph/claude-code-model-proxy/blob/112588175eb2b3b693a5bd23d58d8c501e0ae406/internal/translate/codex_to_claude.go), [구독 전송](https://github.com/jclab-joseph/claude-code-model-proxy/blob/112588175eb2b3b693a5bd23d58d8c501e0ae406/internal/codex/responses.go), [effort 결정](https://github.com/jclab-joseph/claude-code-model-proxy/blob/112588175eb2b3b693a5bd23d58d8c501e0ae406/internal/router/codex.go).

저장소 LICENSE는 GPL v3다. 소스 파일을 MoAI에 복사하지 않았으며, 프로토콜 처리 방식과 검증 사례를 참고하여 기존 MoAI 코드에 한정된 변경을 작성했다.

## Evidence

지정 소스 작업 위치: `/tmp/moai-t649-claude-code-model-proxy`.

```text
git rev-parse HEAD
112588175eb2b3b693a5bd23d58d8c501e0ae406
```

실행한 변환 시험:

```sh
go test ./internal/translate -run 'Test(EffortSelection|EffortFrom|OutputConfigEffortWinsOverThinkingBudget|FixedUpstreamFields|StreamThinkingCarriesMarkedSignature|StreamReasoningWithoutSummaryStillDeliversSignature|StreamParallelToolCallsDoNotInterleave|StreamAdoptsToolCallsOnlyPresentInTerminalEvent)$' -count=1 -v -timeout=60s
```

```text
--- PASS: TestEffortSelection (0.00s)
--- PASS: TestFixedUpstreamFields (0.00s)
--- PASS: TestOutputConfigEffortWinsOverThinkingBudget (0.00s)
--- PASS: TestEffortFrom (0.00s)
--- PASS: TestStreamThinkingCarriesMarkedSignature (0.00s)
--- PASS: TestStreamReasoningWithoutSummaryStillDeliversSignature (0.00s)
--- PASS: TestStreamParallelToolCallsDoNotInterleave (0.00s)
--- PASS: TestStreamAdoptsToolCallsOnlyPresentInTerminalEvent (0.00s)
PASS
ok github.com/jclab-joseph/claude-code-model-proxy/internal/translate 0.329s
```

router의 세 시험도 실행했다.

```sh
go test ./internal/router -run 'Test(ClientEffortIsForwarded|ClientEffortIsClampedToWhatTheModelSupports|TheEffortSentUpstreamIsTheOneResolved)$' -count=1 -v -timeout=60s
```

```text
--- PASS: TestClientEffortIsForwarded (0.01s)
--- PASS: TestClientEffortIsClampedToWhatTheModelSupports (0.00s)
--- PASS: TestTheEffortSentUpstreamIsTheOneResolved (0.01s)
PASS
ok github.com/jclab-joseph/claude-code-model-proxy/internal/router 0.390s
```

실제 MoAI 오류의 안전 진단 범주:

```text
T649_DIAG effort_medium unclassified
T649_DIAG request_translation unsupported effort
T649_DIAG local_400 unclassified
```

수정 후 실제 결과는 [answer-repaired.json](e2e/answer-repaired.json)에 있다.

```json
{"exit_code":0,"assistant_models":["gpt-6-astra"],"marker_in_final_result":true,"success":true}
```

## Baseline-attribution

2026-09-12 이 세션에서 지정 소스의 시험 11개를 직접 실행했다. 소스 자체의 시험은 합성/로컬 시험이며 공개 GPT 서비스 성공의 증거가 아니다. 실제 MoAI 답변의 기준은 WT-unified-gateway, HEAD 81c1d58f9와 미커밋 변경, `/tmp/moai-t649-repaired` SHA-256 `5b259dbbaee35c325d3ff8dd959d7192ff9a6ba91edb2dabab5a18bbc8a792d7`, Claude Code 2.1.268이다.

## Gaps

지정 소스 전체 시험·실계정 실행·전체 보안 감사를 수행하지 않았다. 후속 MoAI 검증에서는 [Read 도구 왕복](e2e/tool-repaired-corrected.json), [도구 세션의 새 프로세스 재개](e2e/resume-tool-repaired.json), [rc.8 대화형 모델 선택과 실제 답변](e2e/interactive-rc8.json)이 통과했다. managed 설정 전체 출처, reasoning 없는 응답의 공개 message/phase 보존, Windows CI는 이 기록으로 통과시키지 않는다.

## Residual-risk

참조 구현의 permissive 변환·모델별 effort clamp·서명 marker만으로의 판단은 MoAI의 엄격한 계약과 다르다. 동일한 코드 형태를 복사한다고 MoAI의 정확성이나 보안이 확보되는 것은 아니다. 실제 핵심 사용 검증과 F1·F2 독립 재감사 후 로컬 시험용 rc.8을 설치했다. [설치 기록](rc8-build.json)에 정확한 바이너리 해시와 rc.7 백업 경로를 남겼으며, 전체 SPEC·CI 완료 판정은 하지 않았다.
