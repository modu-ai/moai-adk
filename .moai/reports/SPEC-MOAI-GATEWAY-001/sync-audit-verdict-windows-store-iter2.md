# Windows AUTH Store F1 독립 재감사 — iter2

## Claim

SPEC: SPEC-MOAI-GPT-AUTH-001 / 보고서 묶음 SPEC-MOAI-GATEWAY-001.
Overall Verdict: **PASS — F1 identity 수리의 로컬 재감사 범위에 한정. Windows native 수용은 PENDING.**
Overall Score: **N/A — Windows 실행 및 coverage 미관측.**
Iteration: 2.

이전 High/blocking F1의 사라진 후보 경로 의존성을 해결한 것으로 판정한다. 제품의 `privatePathInfo`는 metadata access와 read/write/delete sharing으로 연 같은 handle에서 reparse·파일 종류·owner·DACL을 검사하고 `f.Stat()`을 반환한다. 함수 반환 시 deferred Close가 실행되므로 `writePlatform`의 MoveFileEx 전에 후보 handle은 닫힌다. 현재 Go 1.26.8 Windows 구현에서 이 FileInfo는 파일 ID를 즉시 갖고 지연 조회용 path는 비어 있다.

후보뿐 아니라 root, state readback, broker scratch 및 broker auth의 SameFile 입력도 handle 기반 helper 또는 열린 파일의 Stat에서 얻는다. 따라서 이 변경 범위의 SameFile은 pinned directory가 열린 동안 파일 ID를 얻으려고 공유 모드 0으로 경로를 늦게 다시 여는 기존 stdlib 경로를 사용하지 않는다. 이는 제품 및 현재 stdlib 소스 확인과 보존된 semantic fixture의 결합 판정이다. Windows kernel의 실제 sharing 허용 결과를 관측한 것으로 확대하지 않는다.

## Dimension Scores

활성 default profile의 기준을 유지한다. 아래 PASS는 F1 delta 수용만 뜻하며 전체 Windows 기능·보안·coverage must-pass를 충족했다는 판정이 아니다.

| Dimension | Score | Verdict | Evidence |
|---|---:|---|---|
| Functionality (40%) | N/A | PASS, F1 로컬 delta / native UNVERIFIED | `identity_captured_before_rename=false SameFile=false` / `identity_captured_before_rename=true SameFile=true`; crosscompile exit 0 |
| Security (25%) | N/A | UNVERIFIED, native | `helper metadata access + read/write/delete sharing; no path-based Stat: PASS (source assertions)`; manifest diff 출력 없음, exit 0 |
| Craft (20%) | N/A | UNVERIFIED, native | `coverage: 29.5% of statements`는 선별된 macOS 회귀 시험만의 수치; 양 플랫폼 vet 출력 없음, exit 0 |
| Consistency (15%) | 100/100 | PASS, 변경 범위 | `$ gofmt -l internal/gateway/auth` 뒤 `exit=0`, 파일명 출력 없음 |

29.5%를 패키지 전체 coverage 기준 85%의 PASS로 사용하지 않는다. 선별 실행은 Windows 파일을 실행하지 않으므로 Windows coverage 실패를 측정한 것도 아니다. 이전 보고서의 88.4%나 작성자의 테스트 결과를 이번 측정값으로 재사용하지 않는다.

## Findings 및 회차 이력

- **F1 [High] [blocking → resolved for local delta] [confidence: High]** `internal/gateway/auth/store_windows.go:183,262`: 이전 `os.Stat(name)`의 지연 ID 조회를 검증된 handle의 `f.Stat()`으로 대체했다. 기존 semantic fixture의 loadFileId/sameFile 함수가 현재 stdlib 원문과 같은지 재확인했고, rename 이전 ID 확보 대조군 성공 및 미확보 대조군 실패를 재실행했다. 로컬에서 추가 제품 수리는 요구하지 않는다. 실제 Windows 정상 저장 수용은 아래 native 증거 확보 후 별도로 판정한다.
- 새로 기계적으로 확인된 blocking/optional 결함은 없다. native sharing·ACL의 미관측을 별도 제품 결함으로 꾸며 기록하지 않는다.
- Iter1: 지연 identity F1로 FAIL. Iter2: 해당 delta 로컬 PASS, native PENDING 유지.

새 `TestWindowsCandidateIdentitySurvivesRename`은 후보 identity 획득, 실제 rename, 최종 identity 비교, pinned root 검증 순서를 갖추었으며 Windows 시험 바이너리에 교차 컴파일되었다. 이 macOS 실행에서는 그 시험을 실행하지 않았다.

## Baseline-attribution

실측 WT: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/moai-proxy-unified`.
시작과 검사 후 HEAD: `81c1d58f9cf7045594ee61d5e4ff380948ce9eba`, branch: `WT-unified-gateway`.
미커밋 변경물이므로 아래 다섯 파일 SHA-256과 전후 일치를 함께 고정했다. 제품·SPEC·workflow·manifest·사용자 자격 저장소를 수정하지 않았다. 이번 소유 범위는 이 보고서 및 `sync-audit-verdict-windows-store-iter2.md`뿐이다.

`source_session_id=01a08e7b-6aa0-7361-ab7e-ea8da1f02228`은 부모/developer가 지정한 출처이며 CLI로 새로 측정한 UUID가 아니다. 현재 WT CLI `--show-fallback`의 사용 불가 상태 역시 부모 전달 사항이며 이번 실행에서 호출하지 않았다.

## Evidence

명령은 각 호출에서 자격 관련 환경을 해제하고 GOCACHE를 지정했다. 모든 프로세스가 종료되었으며 백그라운드 부하나 live provider를 사용하지 않았다. semantic fixture의 syscall shim은 host 경로 열기를 사용한다. Windows kernel API를 실행하지 않는다.

### semantic

CWD: `/var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/windows-identity-audit-rl8lj_xv`

```text
$ unset CODEX_HOME OPENAI_API_KEY OPENAI_BASE_URL ANTHROPIC_BASE_URL ANTHROPIC_API_KEY ANTHROPIC_AUTH_TOKEN Z_AI_API_KEY && GOCACHE=/tmp/gateway-foundation-cache go test -v -count=1 -timeout=10s .
=== RUN   TestCandidateIdentityAfterRename
    identity_test.go:79: identity_captured_before_rename=false SameFile=false
    identity_test.go:79: identity_captured_before_rename=true SameFile=true
--- PASS: TestCandidateIdentityAfterRename (0.00s)
PASS
ok  	identityaudit	0.343s
exit=0
```

### posix

CWD: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/moai-proxy-unified`

```text
$ unset CODEX_HOME OPENAI_API_KEY OPENAI_BASE_URL ANTHROPIC_BASE_URL ANTHROPIC_API_KEY ANTHROPIC_AUTH_TOKEN Z_AI_API_KEY && GOCACHE=/tmp/gateway-foundation-cache go test -race ./internal/gateway/auth -run '^TestStoreLoginAtomicPrivateAndLogoutGeneration$|^TestStoreRejectsPartialFailureAndLateLoginAfterLogout$|^TestStoreRejectsSymlinkAndPublicPermissions$|^TestStoreRejectsCorruptionPermissionsAndCanceledMutations$|^TestBrokerCannotReplaceScratchHomeWithSymlink$|^TestScratchCleanupFailureDoesNotPublish$|^TestLogoutScratchCleanupFailurePreservesCanonical$' -count=1 -timeout=30s -v -coverprofile=/var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/windows-identity-iter2-ei29uexc/posix.cover
=== RUN   TestStoreRejectsCorruptionPermissionsAndCanceledMutations
=== RUN   TestStoreRejectsCorruptionPermissionsAndCanceledMutations/{
=== RUN   TestStoreRejectsCorruptionPermissionsAndCanceledMutations/{"generation":0}
=== RUN   TestStoreRejectsCorruptionPermissionsAndCanceledMutations/{"generation":1,"tombstone":true,"auth":{}}
=== RUN   TestStoreRejectsCorruptionPermissionsAndCanceledMutations/{"generation":1,"auth":{}}
--- PASS: TestStoreRejectsCorruptionPermissionsAndCanceledMutations (0.03s)
    --- PASS: TestStoreRejectsCorruptionPermissionsAndCanceledMutations/{ (0.00s)
    --- PASS: TestStoreRejectsCorruptionPermissionsAndCanceledMutations/{"generation":0} (0.00s)
    --- PASS: TestStoreRejectsCorruptionPermissionsAndCanceledMutations/{"generation":1,"tombstone":true,"auth":{}} (0.00s)
    --- PASS: TestStoreRejectsCorruptionPermissionsAndCanceledMutations/{"generation":1,"auth":{}} (0.00s)
=== RUN   TestBrokerCannotReplaceScratchHomeWithSymlink
--- PASS: TestBrokerCannotReplaceScratchHomeWithSymlink (0.00s)
=== RUN   TestScratchCleanupFailureDoesNotPublish
--- PASS: TestScratchCleanupFailureDoesNotPublish (0.01s)
=== RUN   TestLogoutScratchCleanupFailurePreservesCanonical
=== RUN   TestLogoutScratchCleanupFailurePreservesCanonical/tombstone
    logout_fault_test.go:98: actual scratch cleanup failure returned local completed, cleanup_failed, remote unknown; canonical bytes preserved
=== RUN   TestLogoutScratchCleanupFailurePreservesCanonical/newer_login
    logout_fault_test.go:98: actual scratch cleanup failure returned local completed, cleanup_failed, remote unknown; canonical bytes preserved
--- PASS: TestLogoutScratchCleanupFailurePreservesCanonical (0.05s)
    --- PASS: TestLogoutScratchCleanupFailurePreservesCanonical/tombstone (0.02s)
    --- PASS: TestLogoutScratchCleanupFailurePreservesCanonical/newer_login (0.03s)
=== RUN   TestStoreLoginAtomicPrivateAndLogoutGeneration
--- PASS: TestStoreLoginAtomicPrivateAndLogoutGeneration (0.02s)
=== RUN   TestStoreRejectsPartialFailureAndLateLoginAfterLogout
--- PASS: TestStoreRejectsPartialFailureAndLateLoginAfterLogout (0.02s)
=== RUN   TestStoreRejectsSymlinkAndPublicPermissions
--- PASS: TestStoreRejectsSymlinkAndPublicPermissions (0.00s)
PASS
coverage: 29.5% of statements
ok  	github.com/modu-ai/moai-adk/internal/gateway/auth	1.547s	coverage: 29.5% of statements
exit=0
```

### compile

CWD: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/moai-proxy-unified`

```text
$ unset CODEX_HOME OPENAI_API_KEY OPENAI_BASE_URL ANTHROPIC_BASE_URL ANTHROPIC_API_KEY ANTHROPIC_AUTH_TOKEN Z_AI_API_KEY && GOCACHE=/tmp/gateway-foundation-cache GOOS=windows GOARCH=amd64 go test -c ./internal/gateway/auth -o /var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/windows-identity-iter2-ei29uexc/auth.test.exe
(stdout/stderr empty)
exit=0
```

### vetwindows

CWD: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/moai-proxy-unified`

```text
$ unset CODEX_HOME OPENAI_API_KEY OPENAI_BASE_URL ANTHROPIC_BASE_URL ANTHROPIC_API_KEY ANTHROPIC_AUTH_TOKEN Z_AI_API_KEY && GOCACHE=/tmp/gateway-foundation-cache GOOS=windows GOARCH=amd64 go vet ./internal/gateway/auth
(stdout/stderr empty)
exit=0
```

### vetposix

CWD: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/moai-proxy-unified`

```text
$ unset CODEX_HOME OPENAI_API_KEY OPENAI_BASE_URL ANTHROPIC_BASE_URL ANTHROPIC_API_KEY ANTHROPIC_AUTH_TOKEN Z_AI_API_KEY && GOCACHE=/tmp/gateway-foundation-cache go vet ./internal/gateway/auth
(stdout/stderr empty)
exit=0
```

### mechanical

CWD: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/moai-proxy-unified`

```text
$ unset CODEX_HOME OPENAI_API_KEY OPENAI_BASE_URL ANTHROPIC_BASE_URL ANTHROPIC_API_KEY ANTHROPIC_AUTH_TOKEN Z_AI_API_KEY && GOCACHE=/tmp/gateway-foundation-cache python3 /var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/windows-identity-iter2-ei29uexc/mechanical.py
store.go c88e1ad176efee0bf6efd18365f51212166513eee9bd2ee91ae5f78383832112 unchanged=True
store_posix.go af6ecdb361ea2c3c275c35a8f51843c968bfa5607608aa36a78e8fe4b2da9742 unchanged=True
store_windows.go 6b992461b805090eb15a2eddb06c82bfdda8f1daf767c0783d32ed4d92cbd479 unchanged=True
store_windows_test.go 6c0eba02806cdf780dacfe34b7ef59c86a5c700f19798cc496bf62622bcd84e6 unchanged=True
private_windows.go b2cfdb971647b695e64b95b2e57357b492a63a4320122ee72dbd6f9e166ea6be unchanged=True
helper validated handle -> f.Stat -> deferred close before MoveFileEx: PASS (source assertions)
helper metadata access + read/write/delete sharing; no path-based Stat: PASS (source assertions)
func (fs *fileStat) loadFileId() error { fixture verbatim equality: PASS
func sameFile(fs1, fs2 *fileStat) bool { fixture verbatim equality: PASS
Windows handle constructor eager ID + empty path: PASS (current stdlib source assertions)
$ gofmt -l internal/gateway/auth
exit=0
$ git diff --numstat -- go.mod go.sum
exit=0
$ git rev-parse HEAD
81c1d58f9cf7045594ee61d5e4ff380948ce9eba
exit=0
$ git branch --show-current
WT-unified-gateway
exit=0
$ go env GOROOT GOOS GOARCH
/Users/goos/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.26.8.darwin-arm64
darwin
arm64
exit=0
store_windows.go:33:// Pin each existing ancestor without FILE_SHARE_DELETE. Reparse points are
store_windows.go:90:windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE, nil, windows.OPEN_EXISTING,
store_windows.go:96:if windows.GetFileInformationByHandle(h, &info) != nil || info.FileAttributes&windows.FILE_ATTRIBUTE_REPARSE_POINT != 0 || info.FileAttributes&windows.FILE_ATTRIBUTE_DIRECTORY == 0 || private && validateWindowsPrivateHandle(h, true) != nil {
store_windows.go:107:info, e := privatePathInfo(s.dir, true)
store_windows.go:108:if e != nil || !os.SameFile(s.platform.identity, info) {
store_windows.go:141:return validateWindowsPrivateHandle(windows.Handle(f.Fd()), false)
store_windows.go:160:windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE, nil, windows.OPEN_EXISTING, windows.FILE_FLAG_OPEN_REPARSE_POINT, 0)
store_windows.go:183:candidate, e := privatePathInfo(name, false)
store_windows.go:236:if e != nil || !os.SameFile(candidate, info) {
store_windows.go:243:final, e := privatePathInfo(path, false)
store_windows.go:244:if e != nil || !os.SameFile(info, final) || validateWindowsPrivatePath(path, false) != nil {
store_windows.go:259:// privatePathInfo captures identity eagerly from a validated handle. Windows
store_windows.go:260:// path-based os.Stat/FileInfo defers file-ID lookup until SameFile, which is too
store_windows.go:262:func privatePathInfo(path string, directory bool) (os.FileInfo, error) {
store_windows.go:268:windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE, nil,
store_windows.go:275:if validateWindowsPrivateHandle(h, directory) != nil {
store_windows.go:281:func (s *Store) stateInfo() (os.FileInfo, error) {
store_windows.go:282:return privatePathInfo(filepath.Join(s.dir, "state.json"), false)
store_posix.go:57:if e = atomicfile.Replace(name, filepath.Join(s.dir, "state.json")); e != nil {
store_posix.go:60:if e = syncDirectory(s.dir); e != nil {
store_posix.go:93:func privatePathInfo(path string, directory bool) (os.FileInfo, error) { return os.Lstat(path) }
store_posix.go:95:func (s *Store) stateInfo() (os.FileInfo, error) { return s.root.Lstat("state.json") }
No secret values inspected or printed.
exit=0
```

보고서 쓰기 전 공유 상태 점검:

```text
$ git fetch -q origin main && git rev-list --count --left-right origin/main...HEAD
0	2879
exit=0
```

## Gaps

- 실제 Windows의 Store login/refresh/logout, 새 rename identity 시험, pinned sharing, owner·ACL·reparse, MoveFileEx/LockFileEx, crash 경계 실행과 GitHub run ID/artifact는 관측하지 않았다. **Windows native CI는 PENDING**이다.
- 선별 macOS 시험 일곱 개는 정상 저장, 늦은 로그인, 손상·권한·취소, scratch symlink, cleanup 실패 보존을 확인한다. 실제 계정 로그인·refresh, 전체 AUTH suite, 전체 저장소 suite는 수행하지 않았다.
- security 소스 경계 검색 및 manifest 변경 없음은 취약점 DB 조회나 OWASP 전체 검증이 아니다.
- Win32 전원 장애 내구성이나 POSIX directory fsync 동등성을 주장하지 않는다. receipt 및 native policy, 새 opaque package는 감사하지 않았다.

## Residual-risk

파일 ID를 늦게 가져오는 F1 원인은 제거되었으나, Win32의 실제 access/share 호환성과 ACL 상속·소유자 의미는 로컬 semantic fixture와 교차 컴파일로 보장할 수 없다. 이미 승인된 Windows 필수 시험 및 새 identity 회귀 시험의 실제 run/pass를 해당 통합 HEAD에서 확보해야 Windows 수용을 판정할 수 있다. 기존 uncertain commit 차단과 POSIX atomic replace·file/directory sync 경계는 소스와 선별 회귀 검사에서 유지되었다. 이 보고서는 push·PR·merge·원격 실행을 수행하거나 수용한 기록이 아니다.

