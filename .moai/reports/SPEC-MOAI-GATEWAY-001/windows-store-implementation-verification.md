# Windows AUTH Store 구현과 로컬 검증

## Claim

승인된 Windows API 저장 계약을 실제 Store 경로에 연결했다. `OpenStore`의 Windows 미지원 분기를 제거하고 플랫폼별 open/lock/write를 분리했다. Windows는 보호된 current-user 루트와 scratch를 직접 생성하고, 조상·루트·사용 중 scratch의 디렉터리 handle을 delete-sharing 없이 유지하여 경로 교체를 막는다. Broker가 만든 파일은 current-user owner와 단일 유효 full-access ACE를 검증하며, 검증한 부모에서 상속된 ACE를 허용한다. NULL/없는 DACL, broad ACE, 잘못된 owner, reparse, 잘못된 파일 종류를 거절한다. 공개 파일의 ACL을 수정해 채택하지 않는다.

Windows 쓰기는 같은 디렉터리의 새 private candidate에 쓰기와 FlushFileBuffers를 완료한 뒤 `MoveFileEx(REPLACE_EXISTING | WRITE_THROUGH)`를 호출한다. COPY_ALLOWED는 없다. 최종 owner/ACL·부모 identity·candidate와 final 파일 identity·내용 readback을 통과해야 성공한다. API 오류나 교체 후 검증 실패는 `ErrAuthCommitUncertain`과 `ErrAuthState`를 함께 반환하고 해당 Store를 사용 불가 상태로 만든다. 이는 rollback 주장 없이 확인되지 않은 자격의 후속 송신을 차단한다. 다시 열 때에도 모든 경로·권한·상태 검증을 거친다.

POSIX는 기존 atomicfile.Replace와 file/directory sync를 유지한다. Broker/Logout의 private scratch 준비를 플랫폼 helper로 연결했다. Unix permission-bit 전용 시험 준비를 플랫폼별 ACL/permission 검사로 바꾸고, cleanup 실패 시험은 Windows의 delete-sharing 거부 handle로 실제 실패를 유도한다. 기존 시험 이름·시나리오를 삭제하거나 Windows skip으로 바꾸지 않았다. 기존 private candidate 시험의 미지원 기대만 승인된 Store 활성화 기대로 바꾸었다.

이 보고서는 **로컬 구현·교차 컴파일·macOS 회귀 검증**에 관한 것이다. Windows native 성공 또는 전체 AUTH 수용 PASS가 아니다.

## Evidence

이 실행에서 수행한 명령과 원문 출력이다. CWD는 아래 Baseline 경로이다.

```text
$ git fetch -q origin main
(exit 0, no output)
$ git rev-list --count --left-right origin/main...HEAD
0	2879
```

새 native 시험을 먼저 작성했을 때의 Windows 컴파일 RED:

```text
$ GOOS=windows GOARCH=amd64 go test -c ./internal/gateway/auth -o /tmp/gateway-windows-auth-red.exe
# github.com/modu-ai/moai-adk/internal/gateway/auth [github.com/modu-ai/moai-adk/internal/gateway/auth.test]
internal/gateway/auth/store_windows_test.go:26:9: s.validatePlatformRoot undefined (type *Store has no field or method validatePlatformRoot)
internal/gateway/auth/store_windows_test.go:35:9: s.validatePlatformRoot undefined (type *Store has no field or method validatePlatformRoot)
internal/gateway/auth/store_windows_test.go:44:10: undefined: setWindowsTestBroadACL
(exit 1)
```

후속 readback 최종화 시험의 컴파일 RED:

```text
internal/gateway/auth/store_windows_test.go:113:9: s.finishWindowsCommit undefined (type *Store has no field or method finishWindowsCommit)
(exit 1)
```

위 RED는 **컴파일 의존성 실패**이며 native 동작 실패를 관측한 증거가 아니다. 파일을 먼저 작성했지만 macOS에서 Windows 실행을 못 했으므로 native RED/GREEN 주기는 CI에서 아직 입증해야 한다.

최종 범위 검증:

```text
$ GOOS=windows GOARCH=amd64 go test -c ./internal/gateway/auth -o /tmp/gateway-windows-auth.exe
(exit 0, no output)
$ GOOS=windows GOARCH=amd64 go vet ./internal/gateway/auth
(exit 0, no output)
$ go vet ./internal/gateway/auth
(exit 0, no output)
$ go test -race ./internal/gateway/auth -count=1 -coverprofile=/tmp/gateway-windows-integration-posix.cover
ok  	github.com/modu-ai/moai-adk/internal/gateway/auth	5.646s	coverage: 88.4% of statements
$ go tool cover -func=/tmp/gateway-windows-integration-posix.cover | tail -1
total: (statements) 88.4%
```

마지막 cover 출력은 열 공백을 정리했다. 88.4%는 macOS가 빌드한 AUTH 패키지의 실행 범위다. Windows 파일 coverage 또는 모든 수정 파일 85% 달성으로 사용하지 않는다. `readPrivateBrokerFile`의 일부 실패 분기와 POSIX 파일 시스템 오류 경로의 개별 coverage는 85% 미만이다. LSP 서버 실행 대신 양 플랫폼 `go vet`와 컴파일을 관측했으며 LSP baseline은 미수집이다.

```text
$ bash scripts/ci-census/windows-gateway-evidence-test.sh
"PASS Windows gateway native evidence: 2 required tests"
REJECTED missing
REJECTED skipped
REJECTED failed
REJECTED pass_without_run
REJECTED truncated
REJECTED malformed
REJECTED absent_stream
```

위 checker 출력의 두 시험은 합성 fixture이며 Windows native 결과가 아니다. 실제 manifest는 기존 9개를 보존하고 새 Store 시험 및 기존 세대/로그아웃/송신/cleanup/broker 시험을 추가하여 **23개**다. Python으로 중복 및 실제 test 함수 정의를 검사했고 이전 22개 시점에 `unique True`, `missing_definitions []`를 관측했다. 마지막 scratch pin 시험을 추가한 후 23개를 다시 확인했다.

새 native 시험은 login/refresh/logout, 루트 rename 거절, broker broad ACL 거절과 기존 세대 보존, post-replace readback 불일치와 credential Apply 차단, flush 전후 프로세스 강제 종료 후 상태 회복, NULL ACL 거절, MoveFileEx 공유 위반 시 canonical 보존, broker scratch rename 거절을 포함한다. 프로세스 fixture는 외부 context timeout·cleanup kill/wait로 종료를 제한한다. 전원 차단 시험을 주장하지 않는다.

## Baseline-attribution

```text
$ git rev-parse --short HEAD
81c1d58f9
$ git branch --show-current
WT-unified-gateway
```

WT: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/moai-proxy-unified`. 이번 변경은 미커밋 상태다. 부모가 제공한 source_session_id는 `01a08e7b-6aa0-7361-ab7e-ea8da1f02228`이며 이 하위 실행의 `moai session current`는 runtime UUID를 제공하지 못하고 canonical unavailable fallback을 출력했다. 이를 fresh UUID 확인으로 기록하지 않는다.

이번 변경 파일의 SHA-256:

| AUTH 파일 | SHA-256 |
|---|---|
| `store.go` | `75b4e61c544814503b49a34117b2a01f607040f82478ab4e2a38a3d830d06ac9` |
| `store_posix.go` | `48f1e470a3c1828aca1760510f5f87175dd76fc3bcc12bceddd7157b512617f5` |
| `store_windows.go` | `f336cfbb3cd990a8951edce463882b609b07956d9a02fffc45a40be1bd386649` |
| `private_windows.go` | `b2cfdb971647b695e64b95b2e57357b492a63a4320122ee72dbd6f9e166ea6be` |
| `lock_windows.go` | `26d96e611b6d909c2858001be581ebe0af809da4484d1f42fc4b284b2c201e73` |
| `broker.go` | `cc84bc25f434291b44409cb61874602dbed163d83943e2a067c916cf4d56c61d` |
| `logout.go` | `f2d76991f40563c293bb4e81d3f9cd77ffc6be27d0947d37db0ae9ca106b318b` |
| `store_windows_test.go` | `8b5bbf204edf47ec092f9b08189ec0d4d3dc850d866f19c0e255f1ed1b73aefe` |
| `private_windows_test.go` | `0193f828ff3402bc7368b13c2e62dc89e6138b7921b49c8495b52e292fa2d2f1` |
| `platform_helpers_test.go` | `a4f9b7080e91768df2163676f0cac8afdf6d2a3a1fadf60e3811a1438beb55a4` |
| `platform_helpers_posix_test.go` | `46b8724199866781c3a9fd98b0f3ef5b65026b0acc4bf8767a4f48fce0e0a6be` |
| `store_test.go` | `70bd2f250992e747c8414feaeca6667607919512964c6033007277212c2cc67f` |
| `broker_test.go` | `e6af8e29e1ab79824df1b9a991a96b40d9dc4175560e926020e8e0d301e19a99` |
| `broker_regression_test.go` | `5f1bf58cfe418b63ef4f4afba9c9907c8177f08c2979743209b333b2aaa31805` |
| `boundary_test.go` | `27f039d44cdacb52c99d879ffe7fb4fb813e40cb6ce192de6f0e21c96a75b2ee` |
| `logout_broker_test.go` | `67cee75a3d8eeb11d8151e54abccc756c8c5e34c5549b7efc8a133009a06a2e2` |
| `logout_test.go` | `9a7ef01f17299abd2c84fc73991effb38fcb18f2bfe263ef4c7d998583d29d7b` |

별도 변경: `scripts/ci-census/windows-gateway-required.json`. 기존 다른 AUTH 작성자의 파일은 모두 보존했다. workflow, SPEC 본문, receipt, CLI, remote 상태는 편집하지 않았다. 부모가 실제 `gpt_auth.go`를 확인한 결과 final `gateway-auth` 디렉터리를 미리 만들지 않고 OpenStore에 넘기므로 해당 CLI 사전 생성 우려는 적용되지 않는다.

## Gaps

GitHub Windows native 실행·run ID·artifact·native coverage는 **PENDING**이다. 통합 브랜치의 GitHub CI run이 저장소 전체 test verdict의 소유자이며, 현재 run ID는 아직 없고 verdict는 **PENDING**이다. push/PR/workflow dispatch 권한을 추정해 실행하지 않았다. 독립 Windows Store 구현 감사도 아직 수행하지 않았다. 잘못된 owner/reparse는 구현 검증 경계가 있으나 새 Store 전용 실제 native adversarial fixture의 관측은 없다. 기존 심볼릭 링크 시험의 runner 권한과 broker child 기본 owner/상속 ACL은 실제 Windows에서 확인해야 한다. 실제 Codex Windows login/refresh는 수행하지 않았다.

## Residual-risk

Native API의 성공·ACL 상속 형태·sharing 동작·프로세스 종료 후 회복은 교차 컴파일로 판정할 수 없다. Windows의 API 오류는 변경 여부를 단정하지 않고 보수적으로 Store 사용을 차단한다. 생성 파일의 current-user owner 계약이 실제 elevated broker 기본 owner와 일치하는지는 runner에서 검증해야 한다. 승인된 계약은 API completion/readback/process crash이며 POSIX와 같은 갑작스러운 전원 장애 내구성 보장이 아니다. 해당 native gate를 통과하기 전에는 Windows 전체 수용 완료로 표시할 수 없다.
