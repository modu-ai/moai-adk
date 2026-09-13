# t652 보존 gateway의 PID 판정 의존 변경

## Claim

보존 gateway baseline을 새 워크트리에 가져오자 부모 종료 supervisor 시험 2개가 실패했다. 기존 legacy 트리의 단일 조건 변경을 이식하여 `os.ErrProcessDone`도 dead PID로 판정한다. App Server 구현 변경은 아직 없다.

## Evidence

명령: `go test ./internal/gateway -run 'TestSupervisorExecRegressionSupervisorSurvivesActualExecAndStopsAfterParentExit|TestSupervisorParentProcessDeath' -count=1 -timeout=30s`

```text
--- FAIL: TestSupervisorExecRegressionSupervisorSurvivesActualExecAndStopsAfterParentExit (3.08s)
    supervisor_exec_test.go:93: reaped parent identity=indeterminate fingerprint="" signal=os: process already finished matchesESRCH=false matchesProcessDone=true
    supervisor_exec_test.go:104: child listener survived parent exit
--- FAIL: TestSupervisorParentProcessDeath (1.01s)
    supervisor_test.go:271: parent death not observed
FAIL
FAIL github.com/modu-ai/moai-adk/internal/gateway 4.407s
```

수정은 internal/homestate/pid_state_unix.go의 조건 한 줄이다.

```go
if errors.Is(err, syscall.ESRCH) || errors.Is(err, os.ErrProcessDone) {
```

수정 후 명령과 출력:

```text
go test ./internal/gateway -run TestSupervisor -count=1 -timeout=60s
ok  github.com/modu-ai/moai-adk/internal/gateway 6.419s

go test ./internal/homestate -run 'PID|Process|Fingerprint|ProfileLease' -count=1 -timeout=60s
ok  github.com/modu-ai/moai-adk/internal/homestate 1.596s
```

## Baseline-attribution

대상 t652 / WT-gateway-appserver-turns / HEAD 28b9988d7 및 오케스트레이터가 source-import-manifest.json으로 확인해 가져온 gateway baseline. 편집 전 원격 main 비교 0 3304. legacy source는 moai-proxy-unified의 기존 미커밋 변경이며 새로운 의미 확장이 아니다.

## Gaps

변경 대상은 !windows 파일이다. Windows 런타임을 검증한 것은 아니다. App Server 인증 seam·HTTP 연계는 아직 구현하지 않았다. 저장소 전체 verdict는 integration branch CI 소관이고 PENDING이다.

## Residual-risk

이번 관측은 macOS의 이미 종료된 Process.Signal(0) 반환값이다. 모든 운영체제의 PID 재사용·fingerprint 경로를 전수 실행한 것은 아니다. 기존 supervisor와 homestate 관련 시험 범위에서 성공했다.
