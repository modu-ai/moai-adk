# t1293 — Stop 체인 예산 경계 수리와 중복 변경 정리

## Claim

Race 잡의 컷오프 판정 경쟁은 `budgeted`와 `advisoryMember`가 결과 채널과 기한 완료 채널 중 임의의 준비된 쪽을 선택하던 경계에서 발생했다. 원래 담당자의 수리는 결과 전송 시각을 실제 기한과 비교한다. 별도 작업자가 중복으로 넣었던 테스트 대체와 검증되지 않은 Git 유지보수 설정은 이 병합에서 되돌린다. Ubuntu의 GPT 임시 디렉터리 정리 실패는 t1296에서 별도로 조사한다.

## Evidence

원래 담당자 브랜치 `WT-stopchain-runner-flake`의 수리 커밋 `303e943e1`, 후속 테스트 정리 `cf6e665b9`, `develop` 병합 `733254ef5`. `internal/cli/codex_stop_chain.go`의 `budgeted`와 `advisoryMember`는 결과의 `sentAt`이 컨텍스트 기한보다 이르면 결과를 쓰고, 그렇지 않으면 컷오프를 쓴다. 담당자의 경계 테스트 두 건을 포함한 재측정:

```text
$ go test -race ./internal/cli -run '^(TestStopChainGateCutOffNeverAllows|TestStopChainBudgetedCutOffBeatsADeadlineEdgeMember|TestStopChainAdvisoryCutOffBeatsADeadlineEdgeMember|TestStopChainGPTProfileNoClaudeDependency)$' -count=2 -timeout 180s
ok  	github.com/modu-ai/moai-adk/internal/cli	10.558s
```

`git diff develop HEAD --name-status`는 이 정리 브랜치에서 `codex_stop_chain_test.go`, `codex_stop_fixture_test.go`, 이 보고서 세 경로만 다른 것으로 나타났다. 두 테스트 파일의 차이는 중복 커밋 `f98fab331`의 테스트 대체와 `gc.auto=0`·`maintenance.auto=false`를 되돌리는 내용이다.

## Baseline-attribution

위 테스트는 원래 담당자 병합 `733254ef5`를 흡수한 `WT-stop-chain-budget` 작업 트리에서 이번 실행에 측정했다. 원래 담당자의 생산 코드와 새 경계 테스트는 유지한 상태다. CI의 최종 판정은 아직 받지 않았다.

## Gaps

Ubuntu 임시 디렉터리 정리 실패의 작성자와 재현 조건은 확인되지 않았다. t1296의 CI 계측·재발 포획 전까지 원인을 확정하지 않는다. `733254ef5` 이후 통합 트리의 새 CI 판정도 아직 없다.

## Residual-risk

원래 `1ns` 예산 테스트는 시스템 부하에 민감한 입력이지만, 결과가 기한 안에 전송됐는지를 코드가 별도로 판별한다. 로컬 race 2회 통과를 반복 CI의 대체 근거로 쓰지 않는다.
