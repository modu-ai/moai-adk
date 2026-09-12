# Windows Store identity 지연 조회 수정

## Claim

독립 감사 F1의 지연 FileInfo identity 문제를 수정했다. Windows 후보 파일의 `os.Stat(name)`을 검증된 handle의 `f.Stat()`으로 바꾸었다. 새 `privatePathInfo`는 reparse·owner·DACL·종류를 같은 handle에서 확인한 뒤 identity가 즉시 채워진 FileInfo를 반환하고 handle을 닫는다. 따라서 rename은 후보 handle이 닫힌 뒤 수행되며 사라진 후보 이름에서 identity를 다시 조회하지 않는다.

루트·broker scratch·broker auth 파일·state readback의 SameFile 피연산자도 Windows에서는 이 handle 기반 helper를 사용한다. helper는 metadata access와 read/write/delete sharing으로 열어 기존 pinned directory의 sharing 계약과 충돌하는 지연 경로 재열기를 피한다. POSIX state 조회는 기존 `s.root.Lstat`를 유지하며 나머지 POSIX helper도 기존 Lstat 동작을 유지한다. ACL·부모/파일 identity·내용 readback·불확실한 자격 차단 계약은 유지했다.

새 native `TestWindowsCandidateIdentitySurvivesRename`은 후보 identity 확보 후 실제 rename하고 이전 경로가 사라진 상태에서 final identity와 비교한다. 기존 native 정상 Store 시험·23개 필수 manifest는 보존했다. 이 수정을 Windows native PASS로 표시하지 않는다.

## Evidence

먼저 새 시험을 작성한 컴파일 RED:

```text
$ GOOS=windows GOARCH=amd64 go test -c ./internal/gateway/auth -o /tmp/gateway-windows-identity-red.exe
# github.com/modu-ai/moai-adk/internal/gateway/auth [github.com/modu-ai/moai-adk/internal/gateway/auth.test]
internal/gateway/auth/store_windows_test.go:323:14: undefined: privatePathInfo
internal/gateway/auth/store_windows_test.go:326:19: undefined: privatePathInfo
(exit 1)
```

이는 컴파일 RED이며 Windows 동작을 실행한 결과가 아니다.

최종 명령 및 관측:

```text
$ GOOS=windows GOARCH=amd64 go test -c ./internal/gateway/auth -o /tmp/gateway-windows-identity.exe
(exit 0; no output)
$ GOOS=windows GOARCH=amd64 go vet ./internal/gateway/auth
(exit 0; no output)
$ go vet ./internal/gateway/auth
(exit 0; no output)
$ go test -race ./internal/gateway/auth -count=1 -timeout=90s
ok  	github.com/modu-ai/moai-adk/internal/gateway/auth	5.495s
$ git rev-parse --short HEAD
81c1d58f9
$ git branch --show-current
WT-unified-gateway
```

이번 실행에서 Go 1.26.8 `src/os/types_windows.go:46–80`과 `stat_windows.go:117`을 직접 읽었다. `newFileStatFromGetFileInformationByHandle`은 `GetFileInformationByHandle`의 VolumeSerialNumber/FileIndexHigh/FileIndexLow를 즉시 채우고 지연 조회용 path를 비운다. 실제 새 helper는 `os.NewFile`로 만든 handle의 `f.Stat()`을 반환한다. 이 소스 확인은 Windows kernel 실행 증거가 아니다. 감사자가 작성한 별도 semantic fixture를 이번 실행에서 재실행했다고 주장하지 않는다.

## Baseline-attribution

WT `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/moai-proxy-unified`, branch `WT-unified-gateway`, HEAD `81c1d58f9`. 부모가 제공한 source_session_id `01a08e7b-6aa0-7361-ab7e-ea8da1f02228`에 귀속한다. 미커밋 변경 4개 파일의 최종 SHA-256:

| AUTH 파일 | SHA-256 |
|---|---|
| `store.go` | `c88e1ad176efee0bf6efd18365f51212166513eee9bd2ee91ae5f78383832112` |
| `store_posix.go` | `af6ecdb361ea2c3c275c35a8f51843c968bfa5607608aa36a78e8fe4b2da9742` |
| `store_windows.go` | `6b992461b805090eb15a2eddb06c82bfdda8f1daf767c0783d32ed4d92cbd479` |
| `store_windows_test.go` | `6c0eba02806cdf780dacfe34b7ef59c86a5c700f19798cc496bf62622bcd84e6` |

## Gaps

Windows native 실행·새 identity 시험의 실제 run/pass·native coverage·GitHub run ID는 PENDING이다. 저장소 전체 verdict는 통합 브랜치 GitHub CI run의 소유이며 아직 PENDING이다. 기존 Windows Store 감사 F1의 독립 재감사도 이 보고서 작성 시점에는 미실행이다. 이 변경에서 SPEC·workflow·manifest·remote 상태를 수정하지 않았다.

## Residual-risk

정상/오류 공유 모드, ACL 상속, root/scratch pin 및 실제 rename 후 identity 비교는 GitHub Windows runner에서 확인해야 한다. 명시 handle identity로 F1의 사라진 후보 경로 의존성을 제거했지만 crosscompile만으로 native 전체 수용을 입증하지 않는다. 기존 승인된 Windows API 저장 완료 기준과 전원 장애 보장 제외는 그대로다.
