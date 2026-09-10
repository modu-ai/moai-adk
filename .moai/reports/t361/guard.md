# t361 정적 가드 — 뮤테이션 증거

운영자 판정 (b)(2026-09-02, lead-1 경유)에 따라 카드를 "cross-package `Handle` 호출 테스트의
옵션 누락을 막는 정적 가드 1건"으로 재정의한 뒤의 구현 증거.

- 기준선: `origin/develop` = `ad272be20`
- 추가 파일: `internal/hook/session_start_deferred_optout_guard_test.go` (테스트 1개, 신규)
- 선례: `internal/cli/harness/propose_boundary_test.go` `TestPropose_NoAskUserQuestion`

## Claim

가드 `TestDeferredScanOptOut_CrossPackageTestCallersOptOut` 는 두 축 모두에서 살아 있다.
① 규약 위반(옵션 누락)을 잡고 ② 자기 사정거리가 끊기면 스스로 실패한다.

## 선례가 자기 유효성을 증명하는 방식

`TestPropose_NoAskUserQuestion` 은 마지막에 `if scanned == 0 { t.Fatal(...) }` 를 둔다. 훑은
모집단이 비면 초록이 아무것도 뜻하지 않으므로, 빈 모집단 자체를 실패로 만든다. 이 가드도 같은
성질을 `callSites == 0` 으로 가져왔다.

선례와 다른 점 하나: 선례는 부분문자열 스캔이지만 이 가드는 `go/ast` 파싱이다. 대상 호출이
여러 줄에 걸쳐 있어서다 — 예: `graph_deferred_refresh_test.go:67` 은 생성자·옵션·옵션이 4줄에
나뉜다. 부분문자열로는 "같은 호출에 옵션이 붙었는가"를 판정할 수 없다.

## Evidence — 뮤테이션 2건

### M1 — 규약 위반을 잡는가

`internal/cli/binary_lag_test.go:63` 에서 `hook.WithSynchronousDeferredScans()` 제거:

```
$ go test -count=1 -run 'TestDeferredScanOptOut_CrossPackageTestCallersOptOut' ./internal/hook/
--- FAIL: TestDeferredScanOptOut_CrossPackageTestCallersOptOut (0.07s)
    session_start_deferred_optout_guard_test.go:167: internal/cli/binary_lag_test.go:63 calls NewSessionStartHandler without WithSynchronousDeferredScans.
        A test outside internal/hook owns the directory it hands Handle, and t.TempDir removes it the moment the test body returns. [...]
FAIL	github.com/modu-ai/moai-adk/internal/hook	0.512s
```

파일·행이 정확히 지목된다. 원복 후 `git status --porcelain` 로 무변경 확인.

### M2 — 사정거리가 끊기면 스스로 실패하는가

`handlerCtor` 상수를 아무것도 매칭하지 않는 이름으로 바꿈:

```
$ go test -count=1 -run 'TestDeferredScanOptOut_CrossPackageTestCallersOptOut' ./internal/hook/
--- FAIL: TestDeferredScanOptOut_CrossPackageTestCallersOptOut (0.08s)
    session_start_deferred_optout_guard_test.go:189: scanned 0 cross-package NewSessionStartHandlerZZZ test call sites; the guard is not reaching its subjects and its pass asserts nothing
FAIL	github.com/modu-ai/moai-adk/internal/hook	0.685s
```

생성자 이름 변경·패키지 이동·워크 파손 어느 쪽이든 초록으로 통과하지 않는다. 원복 확인.

### 기준선 초록 + 패키지 재측정

```
$ gofmt -l internal/hook/session_start_deferred_optout_guard_test.go   # 출력 없음
$ go vet ./internal/hook/                                              # exit 0
$ go test -count=1 ./internal/hook/         → ok  35.651s
$ go test -count=1 -race ./internal/hook/   → ok  37.510s
```

## 설계 결정 3건

1. **가드의 거처는 `internal/hook`.** 규약의 주체가 이 패키지의 노출 옵션이므로 소유 패키지가
   자기 계약을 지킨다. 스캔 범위는 거처와 무관하게 모듈 전체다.
2. **루트는 `go.mod` 를 거슬러 찾는다** — `git rev-parse` 가 아니다. git 이 없거나 shallow
   clone 인 환경에서 가드가 스스로를 건너뛰면, 재지 않은 통과를 통과로 보고하게 된다.
3. **프로덕션 호출자는 범위 밖.** 프로덕션은 async 경로를 원하고 그게 deferred 단계의 존재
   이유다. 의무는 `_test.go` 에만 붙는다 — `t.TempDir` 이 테스트 종료 시점에 사라지는 쪽.

## Gaps

- 예외 통로(escape hatch)를 두지 않았다. 패키지 밖에서 **일부러** async 경로를 시험하려는 테스트가
  앞으로 생기면 이 가드가 막는다. 선례도 예외 통로가 없으므로 같은 자세를 택했고, 필요해지는 시점에
  실패를 보고 판단하는 편이 지금 쓰지 않을 통로를 미리 파는 것보다 낫다고 봤다.
- t361 원 실패의 로컬 재현은 여전히 미달성이다(`reproduction.md` § Gaps). 이 가드는 그 실패를
  재현하지 않으며, 재발 경로를 기계적으로 막을 뿐이다.
- 로컬 초록은 CI 판정이 아니다. 기준선 `ad272be20` 의 develop 은 별건(`internal/web` i18n)으로
  적색이라고 리드가 통지했다.
