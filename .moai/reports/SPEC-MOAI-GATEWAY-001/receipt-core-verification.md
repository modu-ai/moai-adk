# Receipt core 구현 검증

## Claim

승인된 receipt 변경분 계획(iter2 PASS 1.00)의 재사용 가능한 hash-only core와 POSIX manifest 저장소를 구현했다. 전체 reasoning 활성화나 gateway 통합 PASS가 아니다. 소유 범위는 새 `internal/gateway/receipt/` 패키지 8개 파일과 이 보고서뿐이다. SPEC/CLI/translate/auth 파일은 수정하지 않았다. cycle_type=tdd.

- `New/Publish/Check/Marshal/Parse`: 사전 허가 UUID의 digest, 공개 prefix·이전 prefix, provider, ordered opaque digest·item 수, required/complete만 저장한다. 저장 자료는 4096개 후보와 8MiB 입력/직렬화 한도를 적용한다. 고정 길이 필드와 제한된 provider 값 때문에 정상 4096개 자료는 byte 한도보다 먼저 후보 한도에 도달한다.
- 정상 A/B 후보 각각 허용, 동일 중복 idempotent, 혼합/변조/미등록/개수 불일치/필수 전체 삭제 거절, required=false만 있을 때 빈 envelope 허용, required=true와 공존할 때 빈 envelope 거절을 검증했다. 조회 실패는 거절하고 현재 입력에 없는 미소비 응답을 강제 삽입하지 않는다.
- `CanonicalPrefixes`는 사전 content/opaque 검증을 마친 message 배열을 받는다. strict tool-ID decoder를 주입해야 한다. object key/공백/숫자 의미/같은 text의 문자열·단일 block만 정규화하며 block cache_control과 thinking/opaque를 제외한다. message 안의 system, text, 배열, 원래 ID, 도구 인수/결과/오류를 보존한다. 중복 JSON 필드, 비정상 UTF-8, 짝 없는 UTF-16 escape, 과도한 depth/number exponent, 알 수 없는 block type은 거절한다. SHA-256 상태 snapshot으로 이전 history의 반복 해싱을 피한다.
- `Manifest.Fork`는 child UUID에 귀속된 독립 hash snapshot을 반환한다. 이후 parent 변경이 child를 바꾸지 않는다.
- `OpenStore(ctx, existingAuthorizedRoot, UUID, create)`는 절대 private root만 사용한다. resume에서는 root/lock/manifest를 새로 만들지 않는다. POSIX owner/mode/regular file/nlink/no-follow/identity 확인, OS process lock, temporary file fsync → root-relative rename → directory fsync → manifest readback을 수행한다. store Close는 실행 중 operation과 동기화한다. 완료 이후 API가 반환하며 실제 stream terminal 전달은 통합 caller 소관이다.
- Windows persistence는 `ErrUnsupported`를 유지한다. 부모가 전달한 새 Windows 계약은 인지했으나 별도 primitive를 복제하지 않았다.

## Evidence

### TDD RED

`go test ./internal/gateway/receipt` (구현 파일 생성 전):

```text
# github.com/modu-ai/moai-adk/internal/gateway/receipt [github.com/modu-ai/moai-adk/internal/gateway/receipt.test]
internal/gateway/receipt/core_test.go:8:49: undefined: Candidate
internal/gateway/receipt/core_test.go:8:85: undefined: Hash
internal/gateway/receipt/core_test.go:10:9: undefined: New
internal/gateway/receipt/core_test.go:11:49: undefined: Hash
internal/gateway/receipt/core_test.go:12:19: undefined: Candidate
internal/gateway/receipt/core_test.go:14:19: undefined: Candidate
internal/gateway/receipt/core_test.go:14:61: undefined: Observation
internal/gateway/receipt/core_test.go:15:19: undefined: Observation
internal/gateway/receipt/core_test.go:15:153: undefined: Hash
internal/gateway/receipt/core_test.go:15: undefined: Observation
internal/gateway/receipt/core_test.go:15: too many errors
FAIL	github.com/modu-ai/moai-adk/internal/gateway/receipt [build failed]
FAIL
```

`go test ./internal/gateway/receipt -run TestStore` (store 구현 전):

```text
internal/gateway/receipt/store_test.go:26:14: undefined: OpenStore
internal/gateway/receipt/store_test.go:26:92: undefined: ErrUnsupported
internal/gateway/receipt/store_test.go:33:10: undefined: OpenStore
internal/gateway/receipt/store_test.go:37:12: undefined: OpenStore
internal/gateway/receipt/store_test.go:40:14: undefined: OpenStore
internal/gateway/receipt/store_test.go:67:16: undefined: OpenStore
internal/gateway/receipt/store_test.go:94:12: undefined: OpenStore
internal/gateway/receipt/store_test.go:104:10: undefined: OpenStore
internal/gateway/receipt/store_test.go:132:12: undefined: OpenStore
internal/gateway/receipt/store_test.go:140:12: undefined: OpenStore
internal/gateway/receipt/store_test.go:140:12: too many errors
FAIL	github.com/modu-ai/moai-adk/internal/gateway/receipt [build failed]
FAIL
```

추가 malformed projection 시험은 짝 없는 surrogate 3건과 미초기화 Manifest 수용으로 RED가 났고, 구조 필드 시험은 아래 RED가 났다. 각각 구현 보강 후 같은 시험이 통과했다.

```text
--- FAIL: TestMalformedProjectionAndManifest (0.00s)
    core_test.go:200: uninitialized manifest accepted
--- FAIL: TestManifestRequiresEveryStructuralField (0.00s)
    core_test.go:256: missing count accepted
```

프로세스 잠금 helper의 최초 `select {}`는 Go deadlock 종료로 잠금을 풀어 `lock cancellation <nil>`을 만들었다. 이는 fixture 오류였다. helper를 timer 대기로 바꾸고, 부모 CommandContext 10초와 t.Cleanup Kill/Wait로 수명을 제한한 뒤 실제 contention/cancel/강제 종료 회수 시험이 통과했다. public mutation 시험도 `tool_result` type까지 바꾼 fixture 문자열을 content 필드만 바꾸도록 좁힌 뒤 통과했다. 이 두 수정은 제품 결함의 GREEN 증거로 세지 않는다.

### 최종 독립 명령 병렬 묶음

```text
$ go test -race -coverprofile=/tmp/gateway-receipt-cover.out ./internal/gateway/receipt
ok  	github.com/modu-ai/moai-adk/internal/gateway/receipt	8.145s	coverage: 89.1% of statements
exit_code: 0

$ go vet ./internal/gateway/receipt
output: <empty>
exit_code: 0

$ GOOS=windows GOARCH=amd64 go test -c -o /tmp/gateway-receipt-windows.test.exe ./internal/gateway/receipt
output: <empty>
exit_code: 0

$ gofmt -l internal/gateway/receipt
output: <empty>
exit_code: 0
```

coverage profile을 statement 수로 합산한 실제 파일별 결과:

```text
core.go 95/110 86.36%
platform_posix.go 20/23 86.96%
projection.go 177/189 93.65%
store.go 127/148 85.81%
```

12개 병렬 Publish를 두 Store 인스턴스에서 수행한 뒤 close/reopen에서 12개 모두 남았다. 별도 프로세스가 LOCKED를 출력한 뒤 snapshot의 deadline cancellation을 확인하고, 그 process를 Kill/Wait한 뒤 snapshot이 다시 성공했다. manifest 삭제/손상/공개 권한/심볼릭 링크/하드링크/외부 UUID/root replacement/닫힌 Store를 거절했다. 현재 macOS 비특권 계정에서 root를 0500으로 바꾼 publication은 오류였고, 권한 복원 후 이전 빈 manifest가 유지됐다. privileged runner에서는 같은 시험의 closed-filesystem-handle 분기로 이전 disk 상태 보존을 확인하도록 했다. 그 분기는 이 실행에서 관측하지 않았다.

## Baseline-attribution

- 측정 WT: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/moai-proxy-unified`
- `git branch --show-current` → `WT-unified-gateway`
- `git rev-parse --short HEAD` → `81c1d58f9`
- `moai session current` → `01a08e7b-6aa0-7361-ab7e-ea8da1f02228`
- 기준은 위 HEAD에 현재 새 미커밋 receipt 파일을 더한 상태다. 부모의 다른 파일은 병행 변경 중이므로 그 전체 작업의 테스트 결과로 확장하지 않는다.
- 초기 LSP diagnostics snapshot은 확보하지 않았다. 범위 compiler/race/vet 결과만 제시한다.

- `internal/gateway/receipt/core.go` — `7e7ed0d906584dfdb886971f988133119504cec45dfaf12389967a867da551c3`
- `internal/gateway/receipt/core_test.go` — `fcd58ea16b6757e70cdfb0940ed2d6b18d2497633d3c375ed56c8233476c491d`
- `internal/gateway/receipt/lock_posix_test.go` — `3ded2855d586c9bda61a402bae68a9a63a615823646a9818813bc4205f4e592f`
- `internal/gateway/receipt/platform_posix.go` — `0a10595022e38e84043f8074b9dd5c67b31ee8219bfac9e13ecd103d054f6486`
- `internal/gateway/receipt/platform_windows.go` — `97295c3aed156477bb0539d6e10100f621d19989c391c74a86313e8a38097f34`
- `internal/gateway/receipt/projection.go` — `e2cae7cd97095b3dda24e8ccfbb2f87400042f0f0321019fed972031db571cf9`
- `internal/gateway/receipt/store.go` — `50932606f7df1fec46be36a24e94e995c7ed94c85d4d5baf592ff7652f5c642a`
- `internal/gateway/receipt/store_test.go` — `0abe83fe53042fc2f6d4de670e0e6a195c47c0697dcecca041e0dae4a066ae0c`

## Gaps

- native CLI, private bootstrap, request metadata UUID 비교, owned session index, retained family/profile lease, actual transcript 확인, fork store의 원자 초기화 연결은 이 패키지에 없다. memory Fork snapshot을 제품 fork 성공으로 해석하지 않는다.
- opaque codec, reversible tool ID, foreign-provider stripping, 실제 stream terminal 이전 Store.Publish 연결은 caller에서 구현·검증해야 한다. CanonicalPrefixes가 opaque 원문의 유효성/공급자 서명을 판정한다고 주장하지 않는다.
- `Check`는 CanonicalPrefixes와 validated codec에서 만든 모든 assistant boundary를 받아야 한다. caller가 임의로 observation을 생략하거나 digest를 조작하도록 허용하면 이 API만으로 원본 history를 검증하지 못한다.
- Windows는 cross-compile뿐이다. native Windows ACL/lock/durability는 GitHub CI 책임이며 현재 PENDING이다.
- full repository test verdict는 프로젝트 integration branch의 CI run 책임이다. 이 작업에서 remote/commit/push를 수행하지 않아 run ID가 없고 판정은 PENDING이다.
- 독립 감사와 실제 provider/native 재개 시험은 아직 수행하지 않았다. 테스트는 임시 경로의 합성 자료이며 사용자 credential store에 접근하지 않았다.

## Residual-risk

동일 권한 공격자가 공개 history와 완전한 manifest를 일관되게 이전 snapshot으로 교체하는 rollback을 SHA-256만으로 막지 않는다. 저장 자료는 원문 복구가 아니라 일관성·유실 탐지에 쓰인다. POSIX rename 이후 directory sync/readback 실패는 오류로 반환하지만 파일이 이미 교체되었을 수 있으며 caller는 성공 terminal을 내보내면 안 된다. Manifest는 caller 소유의 비동시 객체이며 Store가 프로세스 간 publication을 직렬화한다. Windows API 계약과 native product 통합이 끝나기 전 reasoning 제품 gate를 열지 않는다.
