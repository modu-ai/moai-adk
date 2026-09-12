# Windows AUTH Store 독립 구현 감사

## Claim

SPEC: SPEC-MOAI-GPT-AUTH-001, Windows 저장 계약 변경분만 평가한다.
Overall Verdict: **FAIL — F1 Functionality 차단 결함**
Overall Score: **N/A — Windows native 미관측으로 전체 점수를 산출하지 않는다.**
Iteration: 1. 기존 windows-api-plan-audit.md는 계획 변경분 판정이며 이 구현 판정의 이전 회차가 아니다.

로컬 macOS AUTH 회귀 검증은 통과했다. 그러나 Windows writer는 rename 전 후보 파일의 identity를 경로 기반 `os.Stat`으로 보관한다. 이 Go 버전의 Windows 구현은 파일 ID를 첫 `os.SameFile`까지 지연 조회하므로, rename 후 사라진 후보 경로에서 조회에 실패한다. 따라서 이 구현을 Windows 완료 또는 정상 성공 경로가 준비된 상태로 수용할 수 없다. F1을 수정한 뒤 해당 HEAD의 실제 Windows CI에 진행해야 한다. 진단 목적으로 현재 코드를 CI에 보내는 것 자체를 금지하는 판정은 아니다.

## Dimension Scores

| Dimension | Score | Verdict | Evidence |
|---|---:|---|---|
| Functionality (40%) | 50/100 | FAIL | `identity_captured_before_rename=false SameFile=false` / `identity_captured_before_rename=true SameFile=true` |
| Security (25%) | N/A | UNVERIFIED | `internal/gateway/auth/broker.go` / `internal/gateway/auth/private_windows.go` — 보안 경계 검색 원문, native 미관측 |
| Craft (20%) | N/A | UNVERIFIED | `ok  	github.com/modu-ai/moai-adk/internal/gateway/auth	5.458s	coverage: 88.4% of statements` — macOS 한정 |
| Consistency (15%) | 100/100 | PASS, 범위 한정 | `gofmt -l internal/gateway/auth` / `go vet ./internal/gateway/auth`: 출력 없음, exit 0 |

활성 profile은 `.moai/config/evaluator-profiles/default.md`, harness.default_profile=default이며 hierarchical 설정은 없다. Functionality의 모든 AC 통과 및 Security의 Critical/High 부재가 must-pass다. F1로 Functionality가 실패하므로 다른 점수에 관계없이 전체 FAIL이다. native API와 Windows coverage는 점수를 추정하지 않았다.

## Findings

- **F1 [High] [blocking] [confidence: High] `internal/gateway/auth/store_windows.go:183` 및 `:236` — rename 뒤 후보 identity 조회가 사라진 이전 경로를 연다.** `os.Stat(name)`은 현재 Go 1.26.8 Windows의 일반 파일 경로에서 `GetFileAttributesEx` 정보를 얻고 `saveInfoFromPath(name)`으로 경로를 저장한다. 파일 ID는 이때 확보하지 않는다. `MoveFileEx` 성공 뒤 `verifyWindowsCommit`에서 `os.SameFile(candidate, info)`를 처음 호출하면 `loadFileId`가 이전 후보 이름을 `CreateFile(OPEN_EXISTING)`로 열려고 한다. 후보 이름은 이미 `state.json`으로 이동했으므로 `SameFile`은 false, finalization은 `ErrAuthCommitUncertain` 및 Store 사용 차단이 된다. 로그인·refresh·logout의 정상 저장 성공 경로를 막는다. **Required fix:** rename 전에 검증된 후보 handle의 `f.Stat()` 또는 동등한 즉시 handle identity를 확보한다. rename 전에 그 handle은 Windows 공유 계약에 맞게 닫고, 최종 handle identity와 비교한다. `os.SameFile(candidate,candidate)`로 경로 캐시를 우연히 채우는 방식보다 명시적인 handle identity가 의도를 보존한다. 기존 native `TestWindowsStoreLoginRefreshLogout` 및 관련 23개 필수 시험을 유지하고 실제 Windows run/pass로 확인한다.

F1의 Windows API 전체 실행 실패를 이 macOS 환경에서 관측한 것은 아니다. 실제 제품 코드와 정확한 Go stdlib 경로를 읽고, 아래의 stdlib 함수 원문을 사용하는 제한된 semantic fixture에서 해당 지연 조회 실패를 기계적으로 재현했다. 이 구분을 native RED로 바꾸어 기록해서는 안 된다.

추가로 검증된 제품 결함은 없다. read handle이 열린 동안 다른 경로 기반 FileInfo의 지연 identity 조회와 Windows 공유 모드가 상호작용하는지, 잘못된 owner/reparse의 실제 거절, 각 crash 경계의 충분성은 아래 Gaps다. 텍스트만 보고 별도 결함으로 단정하지 않았다.

## Evidence

명령 CWD는 별도 표시가 없으면 Baseline의 WT다.

```text
$ unset CODEX_HOME OPENAI_API_KEY OPENAI_BASE_URL ANTHROPIC_API_KEY ANTHROPIC_AUTH_TOKEN ANTHROPIC_BASE_URL && go test -race ./internal/gateway/auth -count=1 -timeout=90s -coverprofile=/tmp/gateway-windows-audit-auth.cover
ok  	github.com/modu-ai/moai-adk/internal/gateway/auth	5.458s	coverage: 88.4% of statements
$ GOOS=windows GOARCH=amd64 go test -c ./internal/gateway/auth -o /tmp/gateway-windows-audit-auth.exe
(no output; exit 0)
$ GOOS=windows GOARCH=amd64 go vet ./internal/gateway/auth
(no output; exit 0)
$ go vet ./internal/gateway/auth
(no output; exit 0)
$ gofmt -l internal/gateway/auth
(no output; exit 0)
$ rg -l 'exec.Command|http.NewRequest|os.Getenv|unsafe.Pointer' internal/gateway/auth --glob '*.go' --glob '!**/*_test.go'
internal/gateway/auth/broker.go
internal/gateway/auth/private_windows.go
$ rg -n 'golang.org/x/sys ' go.mod
30:	golang.org/x/sys v0.47.0
```

위 grep은 입력/프로세스/unsafe 경계와 manifest의 감사 대상 식별용이다. 비밀값을 출력하지 않았으며 OWASP 전체 PASS나 dependency vulnerability scan PASS로 사용하지 않는다. `broker.go`의 명시 환경 구성·절대 실행 경로와 `private_windows.go`의 SID/ACE/handle 검사를 직접 읽었다.

```text
$ go env GOROOT GOOS GOARCH
/Users/goos/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.26.8.darwin-arm64
darwin
arm64
```

위 GOROOT의 `src/os/stat_windows.go:24–49`, `src/os/types_windows.go:287–367`을 `sed -n`으로 읽었다. `loadFileId`와 `sameFile` 함수 원문을 변경 없이 추출하여 독립 Go fixture를 만들고, syscall 경계만 host의 실제 파일 열기로 대체했다. `os.Rename` 전 identity를 확보하지 않은 대조군과 미리 확보한 대조군을 비교했다. 정확한 Windows 함수의 제어 흐름을 실행했지만 Windows kernel API를 실행한 것은 아니다.

```text
$ go test -v -count=1 -timeout=10s .
# CWD: /var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/windows-identity-audit-rl8lj_xv
=== RUN   TestCandidateIdentityAfterRename
    identity_test.go:79: identity_captured_before_rename=false SameFile=false
    identity_test.go:79: identity_captured_before_rename=true SameFile=true
--- PASS: TestCandidateIdentityAfterRename (0.00s)
PASS
ok  	identityaudit	0.328s
```

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

이는 합성 checker 시험이다. 현재 manifest JSON을 Python으로 읽어 `(Package,Test)` 중복과 Go test 함수 정의를 대조한 원문 출력:

```text
required 23 unique 23
missing_definitions []
```

기존 계획 감사의 9개 항목이 현재 manifest 안에 모두 남아 있음을 직접 비교했다. 새로운 14개는 Store 정상 흐름, root/scratch rename, broad/NULL ACL, readback 실패, 공유 위반 replace 실패, process crash, 기존 프로세스 세대/로그아웃/송신/cleanup/broker 흐름이다. 함수 존재나 합성 checker 통과만으로 native run/pass를 주장하지 않는다.

## Baseline-attribution

```text
$ git rev-parse HEAD
81c1d58f9cf7045594ee61d5e4ff380948ce9eba
$ git branch --show-current
WT-unified-gateway
```

WT: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/moai-proxy-unified`. 시작과 보고서 작성 직전 HEAD/branch 동일. 변경물은 미커밋 상태이며 AUTH 소스 hash를 아래에 고정한다. source_session_id `01a08e7b-6aa0-7361-ab7e-ea8da1f02228`은 부모가 전달한 값이며 이 하위 실행에서 관측한 새 UUID가 아니다. 부모에게서 AUTH 감사와 보고서 두 경로만 소유권을 받았다. 제품·SPEC·workflow·다른 작업자 파일·사용자 자격·원격 상태는 변경하지 않았다.

## Gaps

- Windows native Store/ACL/MoveFileEx/LockFileEx/프로세스 crash 실행, 해당 HEAD의 GitHub run ID/artifact와 Windows coverage는 미관측이다. 전체 AUTH 수용을 승인하지 않는다.
- 초기 write/flush 도중, 교체와 최종 readback 사이 등 모든 저장 경계의 강제 종료 fixture는 새 시험의 두 stage만으로 전수 입증되지 않는다. 현재 `flushed`는 직접 candidate helper를 실행하며 `committed`는 Logout 완료 후 종료한다.
- uncertain 시험은 finalization helper를 직접 호출하여 `CredentialRef.Apply`와 Status 차단을 검증하도록 작성돼 있다. 해당 실패를 실제 MoveFileEx 이후 유도하고 `SendAuthorized`의 wire 송신 0까지 관측한 native evidence는 없다.
- ACL wrong-owner/reparse, 조상 rename, inherited broker 자식의 실제 owner/ACE 형태 및 read handle과 지연 `SameFile` 조회의 공유 모드 상호작용은 native에서 확인해야 한다.
- macOS coverage 88.4%는 Windows 코드 coverage가 아니다. provider/Codex 실제 로그인·refresh, 릴리스/CI 전체 suite, 취약점 DB 대조, LSP baseline은 실행하지 않았다. 범위 밖의 receipt와 native policy 감사 결과를 합산하지 않았다.

## Residual-risk

F1 수정 후에도 Win32 공유·ACL 상속·소유자·프로세스 종료 의미는 교차 컴파일과 이 한정 semantic fixture가 보장하지 않는다. 보호된 루트 및 scratch에 삭제 공유 없는 handle을 유지하고, 파일 flush 후 동일 디렉터리 MoveFileEx를 호출하며, post-error Store 사용을 차단하는 현재 경계는 수정 중 보존해야 한다. POSIX writer의 atomicfile.Replace 및 file/directory sync를 유지한다. 프로세스 crash 기준을 전원 장애에서의 POSIX directory fsync 동등성으로 확대하지 않는다.

## Recommendations

F1을 좁게 수정한 뒤 후보 identity 회귀 대조군, 양 플랫폼 컴파일/vet, macOS 영향 범위 race 시험을 확인한다. 이후 실제 GitHub Windows CI의 23개 필수 run/pass 및 artifact를 확보하고 이 감사의 F1 delta와 native Gaps를 다시 판정한다. 이번 턴은 push/PR/merge/dispatch를 승인하거나 실행하지 않는다.

감사 대상 SHA-256:

```text
store.go 75b4e61c544814503b49a34117b2a01f607040f82478ab4e2a38a3d830d06ac9
store_windows.go f336cfbb3cd990a8951edce463882b609b07956d9a02fffc45a40be1bd386649
store_posix.go 48f1e470a3c1828aca1760510f5f87175dd76fc3bcc12bceddd7157b512617f5
private_windows.go b2cfdb971647b695e64b95b2e57357b492a63a4320122ee72dbd6f9e166ea6be
broker.go cc84bc25f434291b44409cb61874602dbed163d83943e2a067c916cf4d56c61d
logout.go f2d76991f40563c293bb4e81d3f9cd77ffc6be27d0947d37db0ae9ca106b318b
store_windows_test.go 8b5bbf204edf47ec092f9b08189ec0d4d3dc850d866f19c0e255f1ed1b73aefe
send.go d699b86d7b747d9320776d4a530258d9edff858de51d41280876bb009a43749e
```
