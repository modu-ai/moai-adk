# t654 A5-M6 — Windows GitHub CI 로컬 준비 증거 (AS-022 / AC-MG-006)

## Claim (이 창의 몫 — 로컬 준비)

release-pr-multi-os.yml windows-latest 레그가 읽을 "이름을 정한 Windows 시험"이 존재하고,
Windows에서 skip 없이 실행 가능한 형태로 명명돼 있으며, 변경 트리가 Windows cross-compile과
시험 바이너리 컴파일을 통과한다. **실제 GitHub CI 실행 증거(`as5-windows-ci-verdict.md`)는 이
창에서 만들지 않는다** — push와 workflow_dispatch는 리드 소관(운영자 판정, 창 대기).

## Evidence

명령: `GOOS=windows GOARCH=amd64 go build ./...` →
```
windows-build-exit=0
```
명령: `GOOS=windows GOARCH=amd64 go vet ./internal/cli/ ./internal/gateway/...` →
```
windows-vet-exit=0
```
명령: `GOOS=windows GOARCH=amd64 go test -c -o /dev/null ./internal/gateway/auth/` 및
`GOOS=windows GOARCH=amd64 go test -c -o /dev/null ./internal/cli/` →
```
windows-test-compile-exit=0
```
명령: `grep -c "t.Skip" internal/gateway/auth/store_windows_test.go internal/gateway/auth/lock_windows_test.go internal/cli/codex_contract_fixture_windows_test.go` →
```
internal/gateway/auth/lock_windows_test.go:0
internal/cli/codex_contract_fixture_windows_test.go:0
internal/gateway/auth/store_windows_test.go:0
```
(skip 0건 — Windows에서 묵살되는 시험 없음)

### 이름을 정한 시험 목록 (release 아티팩트 판독 시 이 이름들과 대조)

`internal/gateway/auth` (Windows 저장소 내구성·권한·크래시 복구 — AS-016의 process
종료·재개·flush/원자교체·권면 표면): TestWindowsLockProcess, TestWindowsLockCancelledBeforeAcquisition,
TestWindowsLockContentionCancellationAndCrashRelease, TestWindowsStoreLoginRefreshLogout,
TestWindowsStoreRejectsRootReplacement, TestWindowsStoreRejectsInheritedBroadBrokerFile,
TestWindowsStoreUncertainCommitBlocksCredential, TestWindowsStoreCrashFixture,
TestWindowsStoreProcessCrashRecovery, TestWindowsStoreRejectsNullACL,
TestWindowsStoreReplacementFailurePreservesCanonical, TestWindowsStorePinsBrokerScratch,
TestWindowsCandidateIdentitySurvivesRename, TestWindowsPrivateStorageCandidate,
TestWindowsPrivateStorageRejectsBroadACL, TestWindowsPrivateStorageRejectsWrongTypeAndRelative.

`internal/cli`: TestWindows… 명명 상수 시험은 이 패키지에 없고, supervisor/launcher의 Windows
spawn-and-wait 경로(`launch_exec_windows.go`, `supervisor_windows.go`)는 위 auth 패키지 시험과
플랫폼 중립 시험(`TestGatewaySessionRealSupervisorHandoffAndCleanup` 등 — windows-latest 레그가
`-tags=integration` 없이 `./...`를 실행하며 함께 돈다)으로 덮인다.

## Baseline-attribution

2026-09-14, worktree `.claude/worktrees/t654`, branch `WT-gateway-launchers` 최종 코드 상태에서
직접 실행.

## Gaps (창 대기 — 리드 실행)

AC-MG-006/AS-022의 PASS 근거가 될 GitHub CI 실행은 이 창에서 수행하지 않는다. 리드의 push 뒤
실행할 명령:

1. `gh workflow run release-pr-multi-os.yml --ref develop`(또는 release PR에서 자동 실행) —
   windows-latest 레그가 `-tags=integration` 없이 `./...` 실행.
2. 실행 완료 후 아티팩트 판독:
   `gh run download <run-id> -n test-stream-release-verify-windows-latest` →
   `test-stream.json.gz`에서 위 이름 각각의 종료 이벤트 `"Action":"pass"` 단언.
3. 판독 결과를 `.moai/reports/t654/as5-windows-ci-verdict.md`에 기록.
   시험 부재·skip·아티팩트 부재는 PASS가 아니다. cross-compile exit 0만으로 PASS 세지 않는다.

## Residual-risk

- Windows 회귀는 release 시점에야 드러난다(카드·develop CI는 windows Go-test 레그 없음 —
  `research.md` §20 실측). 이는 결정 4가 남긴 잔여 위험이다.
- 대화형 TTY·job control(Ctrl-Z 후 fg)의 Windows 동작은 CI runner에서 재현되지 않는다 — 미리
  선언된 Gap(AC-MG-006).
