# t1296 — CI 임시 Git 저장소 정리 실패 조사

## Claim

로컬 Trace2에서 임시 저장소의 두 `git commit`이 각각 분리된 자동 maintenance와 `repack`을 시작하는 것을 확인했다. 이 작업이 테스트 종료 후 `.git/objects`를 쓸 수 있으므로 fixture의 Git 명령에만 `maintenance.auto=false`를 적용했다. 원래 실패한 CI 실행에는 Trace2가 없어 그 순간의 작성자를 직접 특정할 수 없으며, 수정 커밋의 CI 판정도 아직 없다.

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

현재 원격 `develop` `e7eb03517`의 같은 테스트에 읽기 전용 Go overlay를 씌워 성공 때도 전체 Trace2를 출력했다. 출력 860개 이벤트의 47개 Git 프로세스에서 두 fixture `commit`이 각각 `maintenance run --auto --quiet --detach`와 `repack -d -l --cruft --cruft-expiration=2.weeks.ago --quiet --write-midx`를 자식으로 시작했다. 모든 자식은 이번 로컬 실행에서 종료됐지만, [Git 공식 문서](https://git-scm.com/docs/git-maintenance)는 `maintenance.autoDetach`의 기본값이 분리 실행이며 `maintenance.auto=false`가 명령 뒤 자동 maintenance를 막는다고 설명한다.

fixture의 Git 명령에 `-c maintenance.auto=false`를 넣은 뒤 동일한 overlay로 다시 잰 결과:

```text
trace_events 656 git_process_starts 37 auto_maintenance_starts 0 repack_starts 0
test_result --- PASS: TestStopChainGPTProfileNoClaudeDependency (1.24s)
test_exit=0
```

관련 Stop-chain 범위와 race 재측정:

```text
$ timeout 180s go test ./internal/cli -run '^TestStopChain' -count=1 -timeout 160s
ok  \tgithub.com/modu-ai/moai-adk/internal/cli\t45.259s
test_exit=0
$ go test -race ./internal/cli -run '^TestStopChainGPTProfileNoClaudeDependency$' -count=2 -timeout 180s
ok  \tgithub.com/modu-ai/moai-adk/internal/cli\t5.584s
$ go vet ./internal/cli && git diff --check && gofmt -l internal/cli/codex_stop_fixture_test.go
(출력 없음, exit 0)
```

최신 로컬 `develop` `2e34b99b5`를 흡수한 뒤 재측정:

```text
$ go test -race ./internal/cli -run '^TestStopChainGPTProfileNoClaudeDependency$' -count=2 -timeout 180s
ok  	github.com/modu-ai/moai-adk/internal/cli	5.064s
$ go test ./internal/cli -run '^TestStopChainGPTProfileNoClaudeDependency$' -count=1 -v -timeout 120s | rg '^(=== RUN|    codex_stop_chain_test.go:|--- PASS|PASS|ok)'
=== RUN   TestStopChainGPTProfileNoClaudeDependency
    codex_stop_chain_test.go:94: git trace active after fixture setup: 64183 bytes
--- PASS: TestStopChainGPTProfileNoClaudeDependency (1.29s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	2.060s
```

## Baseline-attribution

실패는 `517ec51ba`의 해당 CI 로그에서 확인했다. 수리 전 Trace2는 `origin/develop` `e7eb03517`, 수리 후 Trace2는 `WT-gpt-tempdir-trace`에 이 fixture 수정만 얹은 워킹 트리에서 이번 실행에 측정했다. overlay는 로깅 조건만 바꾸고 저장소 소스는 바꾸지 않았다.

## Gaps

로컬 실행에서는 정리 실패가 재현되지 않았다. 실패한 CI 실행 자체에는 Trace2가 없으므로 그 순간의 정확한 작성자는 미확정이다. 수정 커밋의 통합 CI 결과도 아직 없다.

## Residual-risk

다른 프로세스가 `.git/objects`를 쓰는 별개 경로는 남을 수 있다. CI 재발 시 Trace2 상세로 작성자를 다시 대조해야 한다.
