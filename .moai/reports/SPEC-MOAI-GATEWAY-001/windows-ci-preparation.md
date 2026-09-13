# Windows CI 실행 증거 gate 준비

## Claim

사용자의 Windows 시험은 GitHub CI에서 수행한다는 결정을 따라 기존 Release PR Multi-OS Verification에 native gateway 증거 gate를 추가했다. 새 broad workflow나 중복 Go suite는 만들지 않았다. 기존 Windows full-suite의 test-stream.json을 압축 전에 검사하며 AUTH ACL 후보 3개와 ordinary supervisor 수명 시험 4개의 정확한 package/test 이름에 run→pass가 있고 해당 package가 pass로 끝나야 한다. missing·skip·fail·중복/비정상 종료·잘린/깨진 stream은 성공이 아니다.

Windows artifact 부재는 warn 대신 error다. 새 script/manifest 변경도 기존 paths filter의 matrix 실행 대상으로 추가했다. 플랫폼 Store gate를 없애거나 native 실행을 주장하지 않았다. Push/PR/dispatch 없음.

## Evidence

새 fixture 검증을 checker 구현 전에 실행한 RED:

```text
bash: /Users/goos/MoAI/moai-adk-go/.claude/worktrees/moai-proxy-unified/scripts/ci-census/windows-gateway-evidence.sh: No such file or directory
```

최초 jq 구현은 reduce 괄호 syntax error였으며 수정 후 아래 fixture 판정을 실행했다.

`bash scripts/ci-census/windows-gateway-evidence-test.sh`

```text
"PASS Windows gateway native evidence: 2 required tests"
REJECTED missing
REJECTED skipped
REJECTED failed
REJECTED pass_without_run
REJECTED truncated
REJECTED malformed
REJECTED absent_stream
```

이는 **합성 JSON fixture 두 시험의 checker 단위 검증**이다. Windows native 테스트 2개를 실행했다는 뜻이 아니다. 실제 CI manifest는 아래 7개다.

```text
TestWindowsPrivateStorageCandidate internal/gateway/auth/private_windows_test.go
TestWindowsPrivateStorageRejectsBroadACL internal/gateway/auth/private_windows_test.go
TestWindowsPrivateStorageRejectsWrongTypeAndRelative internal/gateway/auth/private_windows_test.go
TestSupervisorChildStartsAndStops internal/gateway/supervisor_test.go
TestSupervisorRunnerLifetimeClosesPortAndOwnedOverlay internal/gateway/supervisor_test.go
TestSupervisorParentProcessDeath internal/gateway/supervisor_test.go
TestSupervisorCancelledStopStillJoinsReaper internal/gateway/supervisor_test.go
MANIFEST_SOURCE_MATCH 7
```

이 목록은 Python으로 각 함수 정의가 정확히 한 파일에 존재하고 !windows build constraint 파일이 아님을 확인한 출력이다. 실제 실행은 아니다. helper만 호출하고 반환하는 시험은 manifest에 넣지 않았다.

`bash -n scripts/ci-census/windows-gateway-evidence.sh scripts/ci-census/windows-gateway-evidence-test.sh`: exit 0, 출력 없음.

`/Users/goos/go/bin/actionlint .github/workflows/release-pr-multi-os.yml`: exit 0, 출력 없음.

현재 workflow diff를 직접 읽었다. 변경은 paths filter 1행, Windows always 증거 검사 step, Windows artifact 누락 error다. 다른 OS 실행·Go suite·release trigger·권한을 바꾸지 않았다. 기존 census가 이미 jq를 사용하므로 추가 runtime dependency는 없다.

## Baseline-attribution

2026-09-11, WT moai-proxy-unified / WT-unified-gateway / HEAD 81c1d58f9. 기존 release-pr-multi-os.yml 전체 test/upload/gate 구간과 AUTH 후보·supervisor 함수 정의를 읽었다. 실행은 로컬 합성 JSON·shell syntax·actionlint뿐이다. 실제 Claude/Codex/provider를 실행하지 않았다. 새 파일은 scripts/ci-census/windows-gateway-evidence.sh, windows-gateway-evidence-test.sh, windows-gateway-required.json이다.

## Gaps

GitHub Windows run ID/URL·native test-stream·artifact는 아직 없다. manifest의 candidate3 pass가 나와도 Windows Store 전체 활성화 조건은 충족되지 않는다. OpenStore는 현재 ErrPlatformUnsupported이고 native helper는 production 미연결이다.

향후 필요한 구현/시험:

- Windows Store의 directory/lock/state/scratch/auth.json 생성·읽기에 native private ACL 검사를 연결한다. chmod나 mode bits는 ACL 증거가 아니다.
- 상위 경로 reparse/identity 경쟁 방어, write/flush/rename 실패 및 원자 state readback, crash/내구성 검증과 소유 후보 cleanup을 완성한다.
- 별도 Windows OS 프로세스 LockFileEx 경합·취소·crash 해제, late refresh/logout 세대 보호, 다른 계정 접근 거절을 실행한다.
- 실제 Windows에서는 일반 AUTH 시험이 현재 OpenStore gate와 충돌할 수 있다. 이 환경에서 관측한 실패는 아니므로 가능성으로 기록한다. 시험을 skip하거나 지원 gate를 제거하여 녹색으로 만들지 않는다.
- 최종 Windows 지원 판정은 해당 run의 test-stream-release-verify-windows-latest artifact와 정확한 test pass 이벤트로 기록한다. CI 전체 green 또는 후보7개만으로 모든 AUTH/SPEC 완료를 주장하지 않는다.

## Residual-risk

manifest는 의도적으로 고정 이름을 사용하므로 시험을 rename/remove하면 CI는 실패한다. 이름 변경 시 대응과 검증을 함께 갱신해야 한다. 기존 workflow는 release/* PR 또는 수동 dispatch에만 해당하며 일반/문서 전용 skipped matrix는 native evidence가 아니다. 별도 dispatch를 이 작업에서 시작하지 않았다. 전원 장애 내구성을 통상 GitHub runner만으로 완전히 증명할 수 있는지는 시험 설계에서 분리해야 한다. 저장소 전체 판정은 실제 통합/release CI 소관이며 현재 PENDING이다.
