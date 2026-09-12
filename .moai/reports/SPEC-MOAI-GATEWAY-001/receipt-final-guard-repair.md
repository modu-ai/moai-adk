# Receipt 마지막 guard 취소 수리

## Claim

iter2에 남은 RC-F1만 수리했다. `store.go`의 마지막 filesystem guard가 끝난 뒤 Rename 바로 앞에 ctx.Err 검사를 추가했다. `repair_posix_test.go`의 기존 commit 경계 표에 final-guard 취소 행을 추가했다. RC-F2/F3/F4와 receipt 설계는 변경하지 않았다. 독립 재감사 판정은 별도다.

## Evidence

수정 전 지정 overlay와 새 영구 회귀 행을 먼저 실행하여 RED를 확인했다.

```text
$ go test -overlay=/var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/receipt-reaudit-g200solm/overlay.json ./internal/gateway/receipt -run '^TestReceipt' -count=1 -timeout 30s
--- FAIL: TestReceiptAuditIter2FinalGuardCancellation (0.05s)
    receipt_audit_iter2_test.go:7: final_guard_cancel err=<nil> ctx=context canceled candidates=1
    receipt_audit_iter2_test.go:7: cancellation during final precommit guard still published
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/gateway/receipt	0.537s
FAIL

$ go test ./internal/gateway/receipt -run '^TestWriteCancellationBeforeAndAfterCommit/final-guard$' -count=1
--- FAIL: TestWriteCancellationBeforeAndAfterCommit (0.03s)
    --- FAIL: TestWriteCancellationBeforeAndAfterCommit/final-guard (0.03s)
        repair_posix_test.go:148: precommit cancellation <nil>
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/gateway/receipt	0.381s
FAIL
```

수리 후 동일 overlay와 이전 회귀 묶음:

```sh
go test -overlay=/var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/receipt-reaudit-g200solm/overlay.json -race ./internal/gateway/receipt -run '^TestReceiptAudit|^TestWriteCancellation|^TestPrecommitCancellation|^TestLockReplacement|^TestSnapshotReplacement|^TestEmptyEnvelopeAmbiguity|^TestZeroParent' -count=1 -timeout 30s -v
```

exit_code: 0. 핵심 관측 원문:

```text
=== RUN   TestReceiptAuditIter2FinalGuardCancellation
    receipt_audit_iter2_test.go:7: final_guard_cancel err=context canceled ctx=context canceled candidates=0
--- PASS: TestReceiptAuditIter2FinalGuardCancellation (0.03s)
    receipt_audit_iter2_test.go:12: guard=1 A rejected=true B preserved=true
    receipt_audit_iter2_test.go:12: guard=2 A rejected=true B preserved=true
    receipt_audit_iter2_test.go:12: guard=3 A rejected=true B preserved=true
=== RUN   TestWriteCancellationBeforeAndAfterCommit
=== RUN   TestWriteCancellationBeforeAndAfterCommit/before-rename
=== RUN   TestWriteCancellationBeforeAndAfterCommit/final-guard
=== RUN   TestWriteCancellationBeforeAndAfterCommit/after-rename
--- PASS: TestWriteCancellationBeforeAndAfterCommit (0.06s)
    --- PASS: TestWriteCancellationBeforeAndAfterCommit/before-rename (0.02s)
    --- PASS: TestWriteCancellationBeforeAndAfterCommit/final-guard (0.02s)
    --- PASS: TestWriteCancellationBeforeAndAfterCommit/after-rename (0.03s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/gateway/receipt	1.674s
```

최종 범위 시험과 vet를 병렬 실행했다.

```text
$ go test -race -coverprofile=/tmp/gateway-receipt-final-guard-cover.out ./internal/gateway/receipt -count=1 -timeout 45s
ok  	github.com/modu-ai/moai-adk/internal/gateway/receipt	5.692s	coverage: 88.9% of statements
exit_code: 0

$ go vet ./internal/gateway/receipt
output: <empty>
exit_code: 0
```

## Baseline-attribution

WT `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/moai-proxy-unified`; 이번 시작 `git rev-parse --short HEAD` 출력 `81c1d58f9`. 해당 HEAD 위의 현재 receipt 미커밋 파일을 측정했다. 다른 writer 파일을 수정하지 않았다. source session은 개발자 제공 `01a08e7b-6aa0-7361-ab7e-ea8da1f02228`이며 이번 UUID CLI 재측정으로 주장하지 않는다.

- `store.go`: `4a62e6a191620654da954e2bf460d8b17d3b85cf89ea2c56cdde354974a5a5f7`
- `repair_posix_test.go`: `f103a8ca74de3a8fecc4b32cda57dfe7a361b0f36909f4175ca7f2eec6f49e1d`

## Gaps

독립 iter3 재감사, 실제 provider/native 통합, Windows native 및 repository integration CI는 이 변경의 완료 주장에 포함하지 않는다. Integration branch CI run은 미실행으로 PENDING이며 remote/commit/push는 수행하지 않았다. LSP/cross-compile은 이번 3줄 product 변경에서 반복하지 않았다.

## Residual-risk

ctx 검사와 syscall이 원자적으로 결합된다고 주장하지 않는다. 파일 검증 I/O 이후 마지막 취소를 다시 확인하는 경계를 보강했다. rename 뒤의 취소는 이미 완료한 기록을 유지하며, 이후 fsync/readback 오류에서 파일이 이미 바뀌었을 수 있다는 기존 제한은 유지한다.
