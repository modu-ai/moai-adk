# t1296 — CI 임시 Git 저장소 정리 실패 조사

## Claim

CI에서 관측된 `.git/objects` 정리 실패의 원인은 아직 확인되지 않았다. 다음 재발 때 Git 자식 프로세스의 시작·종료 기록을 볼 수 있도록 해당 테스트에만 Trace2 진단을 추가했다. 카드는 미완료다.

## Evidence

CI run `36361758033`, Ubuntu job `108740239618`의 실패 로그:

```text
FAILED        github.com/modu-ai/moai-adk/internal/cli  TestStopChainGPTProfileNoClaudeDependency
testing.go:1464: TempDir RemoveAll cleanup: unlinkat /tmp/TestStopChainGPTProfileNoClaudeDependency4291530304/002/tier-proj/.git/objects: directory not empty
--- FAIL: TestStopChainGPTProfileNoClaudeDependency (0.57s)
```

현재 브랜치 `WT-gpt-tempdir-trace`의 계측 실행:

```text
$ go test ./internal/cli -run '^TestStopChainGPTProfileNoClaudeDependency$' -count=3 -timeout 180s
ok  	github.com/modu-ai/moai-adk/internal/cli	5.340s
$ go test ./internal/cli -run '^TestStopChainGPTProfileNoClaudeDependency$' -count=1 -v -timeout 120s
=== RUN   TestStopChainGPTProfileNoClaudeDependency
    codex_stop_chain_test.go:94: git trace active after fixture setup: 48420 bytes
sync gate: go vet failed: fake vet
codex review: fail: - [P1] found issues
- [P2] more
--- PASS: TestStopChainGPTProfileNoClaudeDependency (1.52s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	2.219s
```

`GIT_TRACE2_EVENT`은 테스트의 별도 `t.TempDir` 파일에만 기록한다. 실패 시 프로젝트 임시 디렉터리의 정리가 끝난 뒤 최대 12000바이트를 테스트 로그에 남기고, 마지막에 trace 디렉터리를 정리한다.

## Baseline-attribution

실패는 `517ec51ba`의 해당 CI 로그에서, 계측 활성화와 로컬 통과는 이 브랜치 시작점 `cee197917`에 계측 변경을 얹어 이번 실행에서 확인했다.

## Gaps

로컬 실행에서는 정리 실패가 재현되지 않았다. CI에서 계측이 적용된 뒤 같은 실패가 다시 나타난 로그를 아직 보지 못했다. 따라서 어떤 Git 명령 또는 다른 프로세스가 `.git/objects`를 썼는지는 미확정이다.

## Residual-risk

Git 이외의 작성자라면 Trace2만으로 식별되지 않을 수 있다. 실패 로그에서 Trace2 자식 종료와 `.git/objects`의 활동을 대조한 뒤에만 수리 방향을 정한다.
