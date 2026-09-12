# Receipt 최종 guard 수리 독립 재감사 — iter3

## Claim

SPEC: SPEC-MOAI-GATEWAY-001. Overall Verdict: **PASS — RC-F1 최종 guard 수리 delta 한정**, 평가 점수 **100/100**.

iter2에 남은 RC-F1의 마지막 filesystem guard 중 취소를 그대로 재현하여 `context canceled`와 후보 0개를 확인했다. RC-F2·RC-F3·RC-F4의 해결 상태도 같은 명령에서 유지되었다. 새 결함 탐색이나 더 넓은 감사 반복을 수행하지 않았다.

| Dimension | Score | Verdict | Evidence (관측 원문) |
|---|---:|---|---|
| Functionality (40%) | 100/100 | PASS, RC-F1 delta | `final_guard_cancel err=context canceled ctx=context canceled candidates=0` |
| Security (25%) | 100/100 | PASS, 해결 항목 회귀 한정 | `guard=3 A rejected=true B preserved=true` |
| Craft (20%) | 100/100 | PASS, 새 검사 2개 statement 한정 | `store.go:262.2,262.29 1 24` 및 `store.go:262.29,264.3 1 2` |
| Consistency (15%) | 100/100 | PASS | `gofmt exit=0 output=''` |

기본 evaluator profile의 필수 통과 조건을 수리 delta에 적용했다. 이 점수와 PASS는 전체 receipt 패키지 커버리지, 제품 통합, root 활성화, Windows 또는 실제 provider 수용을 뜻하지 않는다. 이전 넓은 범위의 Gaps를 이번 판정으로 제거하지 않는다.

### Findings 및 이전 항목 판정

- **RC-F1 — RESOLVED, confidence high.** `internal/gateway/receipt/store.go:262`가 마지막 guard 뒤 Rename 직전에 취소를 다시 확인한다. 이전 final-guard probe가 오류 없이 완료하던 상태에서 context.Canceled·후보 0개로 바뀌었다. nonce 시점과 일반 precommit 취소도 이전 snapshot을 보존하고, after-rename 취소는 이미 완료한 기록을 유지했다.
- **RC-F2 — RESOLVED 유지, confidence high, 관측된 interleaving 한정.** nonce 및 guard 1·2·3의 lock 교체에서 stale A는 오류이고 최종 후보 값은 정확히 B였다. snapshot 교체 회귀 시험도 통과했다. 모든 같은 UID의 임의 filesystem 조작이나 원자 CAS 방어로 확대하지 않는다.
- **RC-F3 — RESOLVED 유지, confidence high.** 동일 공개 prefix의 교차 provider required=true/false 혼재는 빈 envelope를 거절한다. plain foreign 단독 및 정확한 OpenAI opaque의 양성 대조군은 유지된다.
- **RC-F4 — RESOLVED 유지, confidence high.** 초기화하지 않은 parent의 Fork가 ErrInvalid로 거절된다.

이번 delta에서 남은 blocking/optional 발견 사항은 없다. Required fix: 없음. 제품 경로 소유권을 돌려준다.

## Evidence

iter2의 정확한 overlay 경로를 재사용했다. 제품·SPEC·overlay source를 수정하지 않았고 기존 수리 회귀 시험을 함께 실행했다. 아래 appendix에 명령과 전체 관측 출력, 검사 source, 변경하지 않은 source hash를 기록한다. 구현자가 작성한 `receipt-final-guard-repair.md`의 GREEN을 이번 독립 실행으로 대신하지 않았다.

새 product 변경의 두 statement는 coverage profile에서 각각 24회와 2회 실행되었다. 전체 선택 시험 커버리지는 74.3%다. 새 두 statement의 100%를 전체 package의 85% 충족으로 해석하지 않는다.

## Baseline-attribution

- 실제 root: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/moai-proxy-unified`
- 실제 branch: `WT-unified-gateway`
- 실제 HEAD: `81c1d58f9cf7045594ee61d5e4ff380948ce9eba`
- 기준은 이 HEAD 위의 미커밋 receipt 수리본이다. receipt 파일의 시작/종료 hash가 같았으며 다른 세션의 writer 파일을 근거에 합치지 않았다.
- source session: `01a08e7b-6aa0-7361-ab7e-ea8da1f02228` — **developer-pinned 값**이다. 이 실행에서 CLI가 반환한 UUID로 주장하지 않는다. WT CLI `--show-fallback` 경로를 사용할 수 없다는 상태는 위임 정보이며 이번 감사에서 그 CLI를 다시 실행하지 않았다.
- 정확한 재현 overlay: `/var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/receipt-reaudit-g200solm/overlay.json`

### 반복 이력

iter1에서 RC-F1~F3 blocking과 RC-F4 optional을 기록했다. iter2에서 RC-F2~F4를 해결로 판정했으나 마지막 guard의 취소로 RC-F1을 유지했다. iter3에서 같은 final-guard 재현이 취소 오류·후보 0개로 끝났고 이전 해결 항목의 회귀도 없었다. 따라서 마지막 수리 delta를 PASS로 판정한다.

## Gaps

- 전체 receipt 패키지 및 repository suite를 다시 실행하지 않고 final-guard와 이전 해결 항목의 지정 시험만 실행했다. package 전체 Craft 또는 integration CI를 재판정하지 않는다.
- caller가 검증된 history의 모든 assistant observation을 순서대로 구성해야 하는 전제는 그대로다. Check의 caller 경계 관측을 전체 provenance 보장으로 확대하지 않는다.
- 실제 provider·native resume·credential Store·Windows native·root 연결·live 자료를 사용하거나 검증하지 않았다.
- post-rename durability 실패, 같은 UID의 완전한 rollback, 임의의 모든 filesystem race는 이번 범위 밖이다.
- 커밋·push·PR·배포·SPEC·제품 source·다른 writer 파일을 변경하지 않았다. cap/native 후속 감사도 시작하지 않았다.

## Residual-risk

context 검사와 Rename syscall은 단일 원자 연산이 아니다. 이번 변경은 취소 검사 뒤 수행한 filesystem guard I/O 동안의 취소를 다시 확인한다. durable rename 이후의 기록 보존과 후속 sync/readback 오류에서 파일이 이미 바뀌었을 수 있다는 기존 제한은 그대로다. PASS는 마지막 precommit 취소 결함의 수리 수용에만 사용한다.

## Evidence appendix

### Independent delta execution

```sh
unset ANTHROPIC_BASE_URL ANTHROPIC_API_KEY ANTHROPIC_AUTH_TOKEN Z_AI_API_KEY && GOCACHE=/tmp/gateway-foundation-cache go test -overlay=/var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/receipt-reaudit-g200solm/overlay.json -race ./internal/gateway/receipt -run '^TestReceiptAudit|^TestWriteCancellation|^TestPrecommitCancellation|^TestLockReplacement|^TestSnapshotReplacement|^TestEmptyEnvelopeAmbiguity|^TestZeroParent' -count=1 -timeout 30s -v -coverprofile=/var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/receipt-iter3-audit-3aokjiev/coverage.out
```

exit_code: 0

```text
=== RUN   TestReceiptAuditIter2FinalGuardCancellation
    receipt_audit_iter2_test.go:7: final_guard_cancel err=context canceled ctx=context canceled candidates=0
--- PASS: TestReceiptAuditIter2FinalGuardCancellation (0.03s)
=== RUN   TestReceiptAuditIter2LateLockReplacementPreservesB
=== RUN   TestReceiptAuditIter2LateLockReplacementPreservesB/1
    receipt_audit_iter2_test.go:12: guard=1 A rejected=true B preserved=true
=== RUN   TestReceiptAuditIter2LateLockReplacementPreservesB/2
    receipt_audit_iter2_test.go:12: guard=2 A rejected=true B preserved=true
=== RUN   TestReceiptAuditIter2LateLockReplacementPreservesB/3
    receipt_audit_iter2_test.go:12: guard=3 A rejected=true B preserved=true
--- PASS: TestReceiptAuditIter2LateLockReplacementPreservesB (0.09s)
    --- PASS: TestReceiptAuditIter2LateLockReplacementPreservesB/1 (0.03s)
    --- PASS: TestReceiptAuditIter2LateLockReplacementPreservesB/2 (0.03s)
    --- PASS: TestReceiptAuditIter2LateLockReplacementPreservesB/3 (0.03s)
=== RUN   TestReceiptAuditCancelBeforeNonceWrite
    receipt_independent_audit_test.go:8: publish_err=context canceled ctx_err=context canceled persisted_candidates=0
--- PASS: TestReceiptAuditCancelBeforeNonceWrite (0.01s)
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
--- PASS: TestReceiptAuditRootSwapBeforeWriteFails (0.01s)
=== RUN   TestReceiptAuditCrossProviderRequiredAmbiguity
    receipt_independent_audit_test.go:38: required_openai_plus_plain_anthropic empty_envelope_err=invalid conversation receipt
--- PASS: TestReceiptAuditCrossProviderRequiredAmbiguity (0.00s)
=== RUN   TestPrecommitCancellationPreservesSnapshot
--- PASS: TestPrecommitCancellationPreservesSnapshot (0.01s)
=== RUN   TestLockReplacementRejectsStalePublisher
--- PASS: TestLockReplacementRejectsStalePublisher (0.02s)
=== RUN   TestSnapshotReplacementRejectsStalePublisher
--- PASS: TestSnapshotReplacementRejectsStalePublisher (0.02s)
=== RUN   TestWriteCancellationBeforeAndAfterCommit
=== RUN   TestWriteCancellationBeforeAndAfterCommit/before-rename
=== RUN   TestWriteCancellationBeforeAndAfterCommit/final-guard
=== RUN   TestWriteCancellationBeforeAndAfterCommit/after-rename
--- PASS: TestWriteCancellationBeforeAndAfterCommit (0.05s)
    --- PASS: TestWriteCancellationBeforeAndAfterCommit/before-rename (0.02s)
    --- PASS: TestWriteCancellationBeforeAndAfterCommit/final-guard (0.02s)
    --- PASS: TestWriteCancellationBeforeAndAfterCommit/after-rename (0.02s)
=== RUN   TestEmptyEnvelopeAmbiguityAcrossProviders
--- PASS: TestEmptyEnvelopeAmbiguityAcrossProviders (0.00s)
=== RUN   TestZeroParentCannotCreateFork
--- PASS: TestZeroParentCannotCreateFork (0.00s)
PASS
coverage: 74.3% of statements
ok  	github.com/modu-ai/moai-adk/internal/gateway/receipt	1.665s	coverage: 74.3% of statements
```

### Static checks

```sh
unset ANTHROPIC_BASE_URL ANTHROPIC_API_KEY ANTHROPIC_AUTH_TOKEN Z_AI_API_KEY && GOCACHE=/tmp/gateway-foundation-cache go vet ./internal/gateway/receipt
```

stdout/stderr: empty

```text
go vet exit=0
```

### Mechanical checks

```sh
python3 /var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/receipt-iter3-audit-3aokjiev/mechanical.py
```

```python
from pathlib import Path
import hashlib,json,subprocess,re
p=Path(__file__).parent;baseline=json.loads((p/'baseline.json').read_text())
print('baseline hashes unchanged='+str(all(hashlib.sha256(Path(f).read_bytes()).hexdigest()==h for f,h in baseline.items())))
prod=[f for f in baseline if not f.endswith('_test.go')]
for name,pattern in [('shell/env/log sinks',r'\b(?:exec\.Command|os\.Getenv|log\.|slog\.|fmt\.Print)'),('no-follow/private/lock/fsync',r'O_NOFOLLOW|Nlink == 1|s.Uid|Flock|\.Sync\(')]:
 hits=[f'{f}:{i}' for f in prod for i,l in enumerate(Path(f).read_text().splitlines(),1) if re.search(pattern,l)];print(name+': '+(', '.join(hits) if hits else 'no matches'))
print('manifest diff lines:',len(subprocess.check_output(['git','diff','--','go.mod','go.sum'],text=True).splitlines()))
r=subprocess.run(['gofmt','-l','internal/gateway/receipt'],capture_output=True,text=True);print(f'gofmt exit={r.returncode} output={r.stdout!r}')
counts={}
for line in (p/'coverage.out').read_text().splitlines()[1:]:
 loc,n,c=line.split();f=loc.rsplit(':',1)[0].rsplit('/',1)[-1];n=int(n);c=int(c);a,b=counts.get(f,(0,0));counts[f]=(a+n*(c>0),b+n)
for f,(a,b) in sorted(counts.items()):print(f'{f} {a}/{b} {100*a/b:.2f}%')
for cmd in [['git','rev-parse','HEAD'],['git','branch','--show-current']]:print(subprocess.check_output(cmd,text=True).strip())
```

```text
baseline hashes unchanged=True
shell/env/log sinks: no matches
no-follow/private/lock/fsync: internal/gateway/receipt/store.go:227, internal/gateway/receipt/store.go:272, internal/gateway/receipt/platform_posix.go:15, internal/gateway/receipt/platform_posix.go:21, internal/gateway/receipt/platform_posix.go:27, internal/gateway/receipt/platform_posix.go:34, internal/gateway/receipt/platform_posix.go:50
manifest diff lines: 0
gofmt exit=0 output=''
core.go 83/114 72.81%
platform_posix.go 14/23 60.87%
projection.go 139/189 73.54%
store.go 139/179 77.65%
81c1d58f9cf7045594ee61d5e4ff380948ce9eba
WT-unified-gateway
```

### New context check coverage

```sh
rg 'store.go:262' /var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/receipt-iter3-audit-3aokjiev/coverage.out
```

```text
github.com/modu-ai/moai-adk/internal/gateway/receipt/store.go:262.2,262.29 1 24
github.com/modu-ai/moai-adk/internal/gateway/receipt/store.go:262.29,264.3 1 2
```

### Baseline SHA-256

```json
{
  "internal/gateway/receipt/store_test.go": "0abe83fe53042fc2f6d4de670e0e6a195c47c0697dcecca041e0dae4a066ae0c",
  "internal/gateway/receipt/lock_posix_test.go": "fbe45b46148d9898ecc42f0deb85aff5d2872709f00d7178bde12e9fec1bc796",
  "internal/gateway/receipt/store.go": "4a62e6a191620654da954e2bf460d8b17d3b85cf89ea2c56cdde354974a5a5f7",
  "internal/gateway/receipt/core_test.go": "fcd58ea16b6757e70cdfb0940ed2d6b18d2497633d3c375ed56c8233476c491d",
  "internal/gateway/receipt/repair_posix_test.go": "f103a8ca74de3a8fecc4b32cda57dfe7a361b0f36909f4175ca7f2eec6f49e1d",
  "internal/gateway/receipt/core.go": "81f81009bf65030145f8a1dec5084e36a1cf68bf7643f36da83c8f89d0f18cfd",
  "internal/gateway/receipt/repair_test.go": "d02141f5a1ade3b2067dc3b967c0447ba13a5b02eea829781e2f9d9ea0035e07",
  "internal/gateway/receipt/platform_windows.go": "97295c3aed156477bb0539d6e10100f621d19989c391c74a86313e8a38097f34",
  "internal/gateway/receipt/projection.go": "e2cae7cd97095b3dda24e8ccfbb2f87400042f0f0321019fed972031db571cf9",
  "internal/gateway/receipt/platform_posix.go": "0a10595022e38e84043f8074b9dd5c67b31ee8219bfac9e13ecd103d054f6486"
}
```

### Unmodified overlay source identities

- `/var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/receipt-independent-audit-9wppdm_t/receipt_audit_test.go`: `70002ba979ad46fb18b11eadefcea2331b4ad3cbb215239c48868d59f8faa45c`
- `/var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/receipt-reaudit-g200solm/receipt_iter2_test.go`: `82675326affdb88ba00026935067265b59c79962fe3fea2a07826efb8dfd5e4c`

### Remote comparison

```sh
git fetch -q origin main
git rev-list --count --left-right origin/main...HEAD
```

```text
0	2879
```
