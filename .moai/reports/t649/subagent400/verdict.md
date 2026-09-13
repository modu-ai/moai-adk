# ToolSearch 이후 400 오류: 원인 재현과 중단 시점

## Claim

세 실패 에이전트는 ToolSearch 응답의 중첩 `tool_reference`를 받은 직후 다음 모델 요청에 실패했다. 합성 요청으로 두 독립적인 로컬 거절 경로를 재현했다. 이는 서버 약관이나 GPT 모델 자체의 실패로 입증된 것이 아니다.

- translator `textContent`는 tool result 안에서도 text 이외 타입을 거절했다.
- receipt `normalizeContent`는 tool result를 재귀 처리하면서 tool_reference를 거절했다.
- 기존 deferred 도구 생략 정책에는 참조된 도구를 다시 활성화하는 처리가 없었다.

## Evidence

명령: `go test ./internal/gateway/translate ./internal/gateway/receipt -run 'TestToolReference' -count=1`

```text
--- FAIL: TestToolReferenceActivatesOnlySelectedSchema (0.00s)
    tool_reference_test.go:13: only text is supported here
--- FAIL: TestToolReferenceCanonicalPrefixPreservesName (0.00s)
    tool_reference_test.go:11: invalid conversation receipt
```

위 출력은 수정 전 red.log에서 발췌했다. 원본 전체 로그도 보존했다.

수정 후 명령: `go test ./internal/gateway/translate ./internal/gateway/receipt -count=1`

```text
ok  github.com/modu-ai/moai-adk/internal/gateway/translate 0.582s
ok  github.com/modu-ai/moai-adk/internal/gateway/receipt 1.191s
```

`go vet ./internal/gateway/translate ./internal/gateway/receipt`: exit 0, 출력 없음.
추가 scoped.log는 receipt/replay/native-policy/history 관련 gateway 하위 패키지 필터 실행 결과이다. `[no tests to run]` 행은 해당 기능 검증 근거가 아니다.

## Baseline-attribution

워크트리: `.claude/worktrees/moai-proxy-unified`, 브랜치 `WT-unified-gateway`, HEAD `81c1d58f9` 및 기존 미커밋 변경. 편집 전 fetch 및 비교 결과 `origin/main...HEAD = 0 2879`. 세션 `01a08e7b-6aa0-7361-ab7e-ea8da1f02228`.

이번 작업이 작성한 파일:

- internal/gateway/translate/request.go
- internal/gateway/translate/tool_reference.go (신규)
- internal/gateway/translate/tool_reference_test.go (신규)
- internal/gateway/receipt/projection.go
- internal/gateway/receipt/tool_reference_test.go (신규)

수정은 도구 목록에 실제 제공된 참조만 활성화하고, 참조를 request-local 도구 이름으로 텍스트화하며, receipt 해시에는 원래 참조 이름을 포함한다. 기존 account/session/conversation/opaque 검사는 유지했다.

## Gaps

사용자가 Codex App Server 전면 설계로 우선순위를 변경하여 위 최소 패치 직후 추가 변경을 중단했다. 설치·커밋·실제 Claude Code E2E를 하지 않았다. 런타임 오류의 정확한 분기 추적 로그는 확보하지 않았고, 합성 구조 재현으로 두 로컬 거절을 입증했다. 변경 파일 전체 85% 커버리지와 LSP 진단은 측정하지 않았다. 저장소 전체 테스트 판정은 integration branch CI 소관이며 PENDING이다.

## Residual-risk

이 패치는 중단 시점 작업이며 새 App Server 설계의 완료 증거가 아니다. 도구 검색 → 참조 반환 → 스키마 활성화 → 호출 → 결과 → 후속 응답은 새 설계의 실제 E2E 인수 조건으로 유지해야 한다. error tool result 등 별도 미지원 타입은 이번 범위에서 변경하지 않았다.
