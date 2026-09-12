# t652 소스 이식 기준 측정

## Claim
기존 gateway 소스 103개와 픽스처 9개를 원본 해시 그대로 옮겼다. App Server 구현을 추가하기 전 기준 시험은 부모 프로세스 종료 감지 2건에서 실패했다.

## Evidence
명령: `go test ./internal/gateway/... -timeout=90s`

```text
--- FAIL: TestSupervisorExecRegressionSupervisorSurvivesActualExecAndStopsAfterParentExit (3.07s)
    supervisor_exec_test.go:93: reaped parent identity=indeterminate fingerprint="" signal=os: process already finished matchesESRCH=false matchesProcessDone=true
    supervisor_exec_test.go:104: child listener survived parent exit
--- FAIL: TestSupervisorParentProcessDeath (1.01s)
    supervisor_test.go:271: parent death not observed
FAIL
FAIL github.com/modu-ai/moai-adk/internal/gateway 9.866s
ok github.com/modu-ai/moai-adk/internal/gateway/auth 11.583s
ok github.com/modu-ai/moai-adk/internal/gateway/conversation 6.166s
ok github.com/modu-ai/moai-adk/internal/gateway/opaque 1.135s
ok github.com/modu-ai/moai-adk/internal/gateway/receipt 4.940s
ok github.com/modu-ai/moai-adk/internal/gateway/translate 4.136s
FAIL
```

출력의 탭은 공백으로 표시했다. 종료 코드 1.

## Baseline-attribution
작업 트리 t652, HEAD 28b9988d76c349eae7f32c846195af821f318a2c. 원본 gateway는 legacy HEAD 81c1d58f9에서 미추적 상태인 파일을 보존한 스냅샷이다. 각 파일 해시는 source-import-manifest.json에 있다.

## Gaps
App Server 새 adapter, 실제 Claude 왕복, Windows 런타임은 이 명령의 검증 범위가 아니다.

## Residual-risk
기존 체크아웃의 tracked dependency 변경은 일괄 복사하지 않았다. 이번 실패는 homestate/pid_state_unix.go의 os.ErrProcessDone 처리 누락과 일치하며, 해당 수정 이식 후 별도 시험이 필요하다.
