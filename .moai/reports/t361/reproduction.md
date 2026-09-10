# t361 재현 보고 — deferred-scan opt-out 의 패키지 경계

- 카드: t361
- 워크트리: `.claude/worktrees/t361` · 브랜치 `WT-deferred-scan-seam`
- 기준선: `origin/develop` = `ad272be20` (측정 시점 `git rev-list --count --left-right origin/develop...HEAD` = `0 0`)
- 측정 환경: darwin/arm64 로컬. CI 원 관측은 linux/amd64.

## Claim

카드가 서술한 기제 — "`deferredScansAsync` 스위치가 `internal/hook` 미노출 변수라 `internal/cli`
테스트에서 손이 닿지 않는다" — 는 **기준선 `ad272be20` 에서 더 이상 참이 아니다.** 카드가 관측한
실패는 관측 시점 이후 착지한 t352 수리가 이미 닫았다.

단, 카드가 별도로 제기한 설계 우려(방향 (a) 는 호출 패키지마다 각자 끄기를 기억해야 해서 같은 사각이
세 번째로 복제될 수 있다)는 **여전히 유효한 잔여 질문**이다. 현재 인벤토리는 0 간극이지만, 그것을
지키는 기계적 장치는 없다.

## Evidence

### E1 — 관측 head 와 수리 커밋의 순서

```
$ git log -1 --format='%h %ad %s' --date=iso 48d8ef4be
48d8ef4be 2026-08-28 22:28:45 +0900 docs(SPEC-GUARD-LIVENESS-001): integration backfill — completed, sync_commit_sha (t333)

$ git log -1 --format='%h %ad %s' --date=iso 410f6241d
410f6241d 2026-08-28 22:45:36 +0900 fix(SPEC-TEMPDIR-CLEANUP-RACE-001): M1 opt-in synchronous deferred scans (t352)

$ git merge-base --is-ancestor 410f6241d 48d8ef4be
NO: fix landed AFTER the observed failing head
```

카드가 인용한 CI run `33175578950` 의 head `48d8ef4be` 는 수리보다 **17분 앞선다.** 즉 그 실패는
수리 이전 트리의 실패다.

### E2 — 수리가 바로 이 테스트를 대상으로 한다

`internal/cli/binary_lag_test.go:56-63` (기준선 `ad272be20`):

```go
projectDir := t.TempDir()
// WithSynchronousDeferredScans: this test owns projectDir and t.TempDir
// deletes it the moment the test body returns. Without the option Handle's
// deferred MX cold-start scan writes .moai/state/mx-index.json from a
// goroutine that can outrun the join bound, and the write races that
// deletion ("unlinkat ... .moai/state: directory not empty").
// SPEC-TEMPDIR-CLEANUP-RACE-001 REQ-TCR-003.
out, err := hook.NewSessionStartHandler(nil, hook.WithSynchronousDeferredScans()).Handle(...)
```

주석이 인용한 오류 문자열이 카드가 기록한 실패 메시지와 같다.

### E3 — 스위치가 패키지 경계를 넘는다

`internal/hook/session_start.go`:

```
40:	// syncDeferredScans records that this handler's constructor was given
83:// WithSynchronousDeferredScans makes every deferred step of Handle run inline,
110:func WithSynchronousDeferredScans() Option {
125:// asyncDeferredScans reports whether THIS handler's deferred steps run in a
130:func (h *sessionStartHandler) asyncDeferredScans() bool {
131:	return deferredScansAsyncEnabled() && !h.syncDeferredScans
```

`h.asyncDeferredScans()` 소비 지점 4곳: `session_start.go:174` (guardLivenessRefresh),
`:357`, `:602` (binaryLagAdvisory), `:615` (guardLivenessAdvisory). 즉 guard-liveness 경로와
binary-lag 경로 **양쪽 모두** 노출된 옵션을 따른다 — 카드가 "선례의 사각까지 복제됐다"고 지목한
`session_start_binary_lag.go:54` 계열도 포함된다.

### E4 — 기준선에서 초록 (재현 시도 결과: 실패 재현 안 됨)

```
$ go test -count=1 -v -run 'TestBinaryLag_OneSeamServesBothSurfaces|TestDeferredEdgesRefresh|TestSessionStartDeferred' ./internal/cli/
--- PASS: TestBinaryLag_OneSeamServesBothSurfaces (0.06s)
--- PASS: TestDeferredEdgesRefresh_StaleRefreshesAndStagesNothing (0.34s)
--- PASS: TestDeferredEdgesRefresh_FreshNoRewrite (0.20s)
--- PASS: TestDeferredEdgesRefresh_BudgetOverrunWarns (0.08s)
--- PASS: TestSessionStartDeferredWriteDoesNotOutliveHandle (2.13s)
ok  	github.com/modu-ai/moai-adk/internal/cli	4.023s

$ go test -count=1 -race -run '<같은 셀렉터>' ./internal/cli/
ok  	github.com/modu-ai/moai-adk/internal/cli	5.600s
```

셀렉터 공허 초록이 아님을 확인: `grep -c '^=== RUN'` = 5.

### E5 — 뮤테이션 (옵션 제거) 로도 로컬에서는 RED 가 서지 않음

`binary_lag_test.go:63` 을 `hook.NewSessionStartHandler(nil)` 로 되돌린 뒤:

```
$ go test -count=15 -run 'TestBinaryLag_OneSeamServesBothSurfaces' ./internal/cli/
ok  	github.com/modu-ai/moai-adk/internal/cli	1.867s

$ go test -count=30 -race -run 'TestBinaryLag_OneSeamServesBothSurfaces' ./internal/cli/
ok  	github.com/modu-ai/moai-adk/internal/cli	4.836s
```

이후 원본 복원, `git status --porcelain` 빈 출력으로 트리 무변경 확인.

### E6 — 호출자 인벤토리 (기준선 전수)

```
$ grep -rn 'NewSessionStartHandler' --include='*.go' internal/ pkg/ cmd/ | grep -v '^internal/hook/'
internal/cli/deps.go:225                    (프로덕션 — 옵션 없음이 정상)
internal/cli/graph_deferred_refresh_test.go:67   WithSynchronousDeferredScans() 있음
internal/cli/graph_deferred_refresh_test.go:134  WithSynchronousDeferredScans() 있음
internal/cli/session_start_deferred_write_test.go:115  있음
internal/cli/binary_lag_test.go:63                     있음
```

패키지 밖 테스트 호출자 4/4 가 옵션을 넘긴다. 현재 간극 0.

## Baseline-attribution

모든 측정은 이 워크트리, HEAD `ad272be20`, `origin/develop` 과 `0 0`. E1 의 SHA 비교는 같은
트리의 `git log` / `git merge-base` 출력. E4/E5 는 이 실행에서 직접 돌린 `go test` 출력.

## Gaps

- **원 실패를 재현하지 못했다.** E5 의 뮤테이션은 darwin/arm64 에서 45회(15 plain + 30 race)
  전부 초록이었다. 원 관측은 linux/amd64 CI 의 `Test (ubuntu)` + `Race Test` 두 잡. 즉 카드가
  요구한 "심어보고 확인"은 **로컬에서 미달성**이며, 기제 확정은 E1-E3 의 문서·순서 근거에 의존한다.
- 리눅스 컨테이너 재현은 시도하지 않았다(레인은 CI 를 직접 요청하지 않는다).
- t352 SPEC 의 "observation 1" 은 그 SPEC 에서 명시적으로 미재현·범위 밖으로 남겨져 있다.
  t361 이 그 observation 1 과 같은 것인지는 별도 판정이 필요하다 — 본 보고는 t361 이 인용한
  실패(observation 2 형태, `.moai/state` 정리 실패)만 다룬다.

## Residual-risk

- **옵션 누락을 막는 기계적 장치가 없다.** 새 cross-package `Handle` 호출 테스트가
  `WithSynchronousDeferredScans()` 를 잊으면 같은 사각이 세 번째로 복제된다. 현재는 4/4 가
  맞지만 그것을 지키는 것은 규율뿐이다. 이 리포에는 같은 모양의 정적 가드 선례가 있다
  (`internal/cli/worktree/new_test.go` `TestNew_NoAskUserQuestion`).
- 로컬 초록이 CI 초록을 뜻하지 않는다. 기준선의 CI 판정은 별도.
