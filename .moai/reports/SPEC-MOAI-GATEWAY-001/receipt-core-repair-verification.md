# Receipt core 독립 감사 결함 수리

## Claim

RC-F1~F3 blocking과 선택 RC-F4를 receipt 패키지 안에서 수리했다. 감사관의 보존 overlay를 수정 전후 같은 경로로 실행했다. 이는 구현자의 재현·수리 검증이며 독립 재감사 PASS가 아니다.

- RC-F1: write에 context를 전달하고 nonce 생성 뒤 temporary write 전, snapshot 확인 뒤 rename 직전에 취소를 판정한다. precommit 취소는 오류로 끝내며 이전 manifest를 보존한다. rename 뒤 취소는 이미 완료한 기록을 삭제하지 않는다.
- RC-F2: transaction이 잠근 descriptor와 경로의 inode/owner/mode 및 root identity를 검증하는 guard를 commit/readback까지 유지한다. nonce 뒤·rename 전·readback 전후에 재검사한다. publication을 시작할 때 읽은 전체 manifest digest도 rename 전 현재 snapshot과 대조한다. lock 교체와 별개로 다른 내용의 snapshot이 들어오면 stale overwrite를 거절한다.
- RC-F3: 빈 envelope의 required 존재는 같은 Prefix/Previous 전체에서 확인한다. provider 선택으로 필수 opaque의 유실을 정당화할 수 없다. 실제 비어 있지 않은 envelope는 provider/digest/item 수가 정확해야 하며, plain foreign receipt 단독 양성군은 유지한다.
- RC-F4: 초기화하지 않은 parent의 Fork를 거절한다.

제품 수정은 `core.go`·`store.go` 두 파일이다. 기존 permission failure 시험의 internal write 호출을 새 context/guard 서명에 맞췄고, `repair_test.go`·`repair_posix_test.go`에 회귀 시험을 추가했다. 다른 agent 소유 auth/Windows/CLI/translate/SPEC에는 쓰지 않았다.

## Evidence

### 수정 전 동일 overlay RED

```sh
unset ANTHROPIC_BASE_URL ANTHROPIC_API_KEY ANTHROPIC_AUTH_TOKEN Z_AI_API_KEY && go test -overlay=/var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/receipt-independent-audit-9wppdm_t/overlay.json ./internal/gateway/receipt -run '^TestReceiptAudit' -count=1 -timeout 30s -v
```

exit_code: 1. 관측 핵심 원문:

```text
publish_err=<nil> ctx_err=context canceled persisted_candidates=1
first_err=<nil> second_err=<nil> persisted_candidates=1
required_openai_plus_plain_anthropic empty_envelope_err=<nil>
```

같은 단계에 영구 회귀 시험도 먼저 기록하고 실행했다.

```text
--- FAIL: TestPrecommitCancellationPreservesSnapshot (0.03s)
    repair_posix_test.go:44: precommit cancellation <nil>
--- FAIL: TestLockReplacementRejectsStalePublisher (0.03s)
    repair_posix_test.go:77: stale publisher acknowledged
--- FAIL: TestSnapshotReplacementRejectsStalePublisher (0.02s)
    repair_posix_test.go:102: changed snapshot overwritten
--- FAIL: TestEmptyEnvelopeAmbiguityAcrossProviders (0.00s)
    repair_test.go:24: provider selected away required receipt
--- FAIL: TestZeroParentCannotCreateFork (0.00s)
    repair_test.go:38: zero parent fork accepted
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/gateway/receipt	0.415s
FAIL
```

### 수리 후 동일 overlay 및 commit 경계 GREEN

```sh
go test -overlay=/var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/receipt-independent-audit-9wppdm_t/overlay.json -race ./internal/gateway/receipt -run '^TestReceiptAudit|^TestWriteCancellation' -count=1 -timeout 30s -v
```

exit_code: 0. 관측 원문:

```text
=== RUN   TestReceiptAuditCancelBeforeNonceWrite
    receipt_independent_audit_test.go:8: publish_err=context canceled ctx_err=context canceled persisted_candidates=0
--- PASS: TestReceiptAuditCancelBeforeNonceWrite (0.05s)
=== RUN   TestReceiptAuditLockReplacementNoLostPublication
    receipt_independent_audit_test.go:13: first_err=invalid or inaccessible private receipt state second_err=<nil> persisted_candidates=1
--- PASS: TestReceiptAuditLockReplacementNoLostPublication (0.03s)
=== RUN   TestReceiptAuditCanonicalMeaningMatrix
--- PASS: TestReceiptAuditCanonicalMeaningMatrix (0.00s)
=== RUN   TestReceiptAuditForkZeroAndCallerBoundary
    receipt_independent_audit_test.go:24: zero-value parent Fork rejected err=invalid conversation receipt
    receipt_independent_audit_test.go:26: caller-supplied reverse observation order check_err=<nil> (caller owns prefix derivation)
--- PASS: TestReceiptAuditForkZeroAndCallerBoundary (0.00s)
=== RUN   TestReceiptAuditPartialManifestLossAndRollback
    receipt_independent_audit_test.go:30: partial corruption rejected; internally consistent full snapshot rollback accepted as documented
--- PASS: TestReceiptAuditPartialManifestLossAndRollback (0.03s)
=== RUN   TestReceiptAuditRootSwapBeforeWriteFails
--- PASS: TestReceiptAuditRootSwapBeforeWriteFails (0.02s)
=== RUN   TestReceiptAuditCrossProviderRequiredAmbiguity
    receipt_independent_audit_test.go:38: required_openai_plus_plain_anthropic empty_envelope_err=invalid conversation receipt
--- PASS: TestReceiptAuditCrossProviderRequiredAmbiguity (0.00s)
=== RUN   TestWriteCancellationBeforeAndAfterCommit
=== RUN   TestWriteCancellationBeforeAndAfterCommit/before-rename
=== RUN   TestWriteCancellationBeforeAndAfterCommit/after-rename
--- PASS: TestWriteCancellationBeforeAndAfterCommit (0.04s)
    --- PASS: TestWriteCancellationBeforeAndAfterCommit/before-rename (0.01s)
    --- PASS: TestWriteCancellationBeforeAndAfterCommit/after-rename (0.02s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/gateway/receipt	1.517s
```

추가 commit 경계 시험은 실제 Store transaction/write를 사용하고 guard 시점에 context만 취소한다. nonce callback 회귀는 실제 crypto/rand 원 Reader를 보존하며 t.Cleanup으로 전역 Reader를 복원한다. 병렬 실행하지 않는다. 실제 provider·자격 저장소를 사용하지 않는다.

### 최종 범위 검증 병렬 묶음

```text
$ go test -race -coverprofile=/tmp/gateway-receipt-repair-cover.out ./internal/gateway/receipt -count=1 -timeout 45s
ok  	github.com/modu-ai/moai-adk/internal/gateway/receipt	5.693s	coverage: 88.9% of statements
exit_code: 0

$ go vet ./internal/gateway/receipt
output: <empty>
exit_code: 0

$ GOOS=windows GOARCH=amd64 go test -c -o /tmp/gateway-receipt-repair-windows.test.exe ./internal/gateway/receipt
output: <empty>
exit_code: 0

$ gofmt -l internal/gateway/receipt
output: <empty>
exit_code: 0
```

coverage profile statement 합산:

```text
core.go 99/114 86.84%
platform_posix.go 20/23 86.96%
projection.go 177/189 93.65%
store.go 151/177 85.31%
```

## Baseline-attribution

- root: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/moai-proxy-unified`
- 최종 `git branch --show-current`: `WT-unified-gateway`
- 최종 `git rev-parse --short HEAD`: `81c1d58f9`
- 기준은 해당 HEAD의 현재 미커밋 receipt 패키지다. 다른 병행 작업으로 전체 tree가 정지해 있다고 주장하지 않는다.
- source_session_id는 부모/개발자 귀속 `01a08e7b-6aa0-7361-ab7e-ea8da1f02228`이며 이번 수리에서 CLI 재측정 UUID라고 주장하지 않는다.

- `core.go`: `81f81009bf65030145f8a1dec5084e36a1cf68bf7643f36da83c8f89d0f18cfd`
- `store.go`: `2c1effbfec0e202bec93f590debfdcd9edb80119cab023fb2e2cd2dfeb3627ff`
- `lock_posix_test.go`: `fbe45b46148d9898ecc42f0deb85aff5d2872709f00d7178bde12e9fec1bc796`
- `repair_test.go`: `d02141f5a1ade3b2067dc3b967c0447ba13a5b02eea829781e2f9d9ea0035e07`
- `repair_posix_test.go`: `ec87ff4894dcb9bdf02269283e27522129d60580033be1e975b2e40ca41983d3`

## Gaps

독립 재감사, actual native/GPT 통합, Windows native 실행은 남아 있다. Windows는 기존 ErrUnsupported를 유지하며 cross-compile만 통과했다. Check의 모든 assistant observation을 검증된 history에서 순서대로 구성하는 책임은 caller에 있다. 전체 repository 판정은 integration branch CI 책임이며 이 작업에서 remote/commit/push를 수행하지 않아 run ID 없이 PENDING이다. LSP snapshot은 별도 측정하지 않았으며 범위 compiler/race/vet 결과만 제시한다.

## Residual-risk

검사와 rename은 같은 UID의 임의 filesystem 변경에 대해 단일 원자 compare-and-swap이 아니다. 이번 수리는 관측된 nonce interleaving의 stale publication을 거절하며, 가능한 모든 same-UID 조작이나 완전한 일관 snapshot rollback을 방어했다고 확장하지 않는다. rename 뒤 sync/readback 오류에서는 파일이 이미 교체되었을 수 있다. caller는 오류를 성공 terminal로 바꾸거나 자동 재시도하지 않아야 한다. Windows의 승인된 새 durability 계약 구현은 담당 agent의 후속 작업이다.
