# Receipt core 수리분 독립 재감사 — iter2

## Claim

SPEC: SPEC-MOAI-GATEWAY-001. Overall Verdict: **FAIL — RC-F1의 마지막 precommit 취소 경계 미해결**. 평가 점수: **80/100**.

RC-F2·RC-F3·RC-F4는 이번 수리 범위에서 해결로 판정한다. RC-F1의 nonce 취소는 해결되었으나 마지막 파일 검증 guard에서 발생한 취소는 rename 전에 다시 확인하지 않는다. 제품 파일·SPEC를 수정하지 않았으며 실제 provider·기존 credential Store에는 접근하지 않았다.

| Dimension | Score | Verdict | Evidence (관측 원문) |
|---|---:|---|---|
| Functionality (40%) | 75/100 | FAIL | `final_guard_cancel err=<nil> ctx=context canceled candidates=1` |
| Security (25%) | 100/100 | PASS, 수리 delta 한정 | `guard=3 A rejected=true B preserved=true` |
| Craft (20%) | 50/100 | UNVERIFIED, 선택 시험 분모 | `coverage: 74.2% of statements` |
| Consistency (15%) | 100/100 | PASS | `gofmt exit=0 output=''` |

기본 evaluator profile을 사용했다. Functionality 실패가 전체 FAIL을 강제한다. 이전 전체 패키지의 89.1%나 구현자의 수리 후 88.9%를 이번 선택 시험의 수치로 대신하지 않았다.

### Findings 및 이전 결함 판정

- **RC-F1 [Medium/P2] [blocking] [confidence high] — PARTIALLY RESOLVED, 아직 FAIL.** `internal/gateway/receipt/store.go:255`의 ctx.Err 검사 뒤 `:258`의 guard가 root/lock 파일 검증을 수행한다. 그 guard에서 취소하면 `:261`의 Rename 앞에 추가 context 확인이 없어 후보가 저장되고 성공이 반환된다. 실제 transaction/write를 그대로 사용하여 세 번째 precommit guard에 취소를 넣었고 `err=nil`, `ctx=context canceled`, 후보 1개를 관측했다. **Required fix:** 마지막 guard 완료 뒤, Rename 직전에 context를 다시 확인한다. 같은 probe에서 context.Canceled와 이전 snapshot 보존을 확인한다. rename 이후의 취소에서 이미 완료한 기록을 보존하는 기존 계약은 유지한다. 취소와 syscall이 완전히 원자적으로 결합된다는 보장을 요구하는 것이 아니라, 취소 검사를 통과한 뒤 수행한 파일 I/O 동안의 취소를 놓치지 않도록 요구한다.
- **RC-F2 [기존 Medium/P2, blocking] — RESOLVED, confidence high, 관측된 interleaving 한정.** 이전 nonce lock 교체 probe는 A를 거절하고 B를 보존했다. 추가로 write guard 1·2·3에서 lock inode를 교체하고 열린 Store B로 발행했다. 세 경우 모두 A 오류와 최종 유일 후보가 정확한 B임을 확인했다. snapshot 내용만 nonce 시점에 교체한 기존 회귀 시험도 통과했다. 같은 UID의 모든 임의 변경이나 단일 원자 CAS를 보장한다고 확대하지 않는다.
- **RC-F3 [기존 Medium/P2, blocking] — RESOLVED, confidence high.** 같은 Prefix/Previous의 OpenAI required=true와 Anthropic required=false 혼재에서 빈 envelope가 거절되었다. plain foreign 후보 단독 및 정확한 OpenAI envelope는 유지되고 provider가 다른 opaque는 거절되는 회귀 시험도 통과했다.
- **RC-F4 [기존 Low/P3, optional] — RESOLVED, confidence high.** zero-value parent의 Fork가 ErrInvalid로 거절된다.

### Recommendations

추가 수리는 RC-F1의 마지막 guard와 Rename 사이 context 재검사로 제한한다. 재감사에는 추가 final-guard probe와 pre/post-commit 취소 대조군, 늦은 lock 교체의 B 보존 시험을 사용한다. cap/native·Windows·root 활성화 작업은 이번 범위에 합치지 않는다.

## Evidence

아래 appendix에 실제 명령·전체 관측 출력·검사 source·baseline hash를 보존했다. 이전 독립 overlay는 그대로 포함했고 별도 추가 source에서 final guard 취소 및 guard 1·2·3 lock 교체를 시험했다. 모든 interleaving은 실제 Store transaction/write와 실제 temporary filesystem을 사용한다. guard wrapper는 filesystem 판정 결과를 위조하지 않고 해당 시점의 취소나 lock 교체만 수행한 뒤 원 guard를 실행한다.

기존 nonce 취소·A/B 유실·교차 provider 혼재·zero Fork 재현은 통과했다. 새 final guard 취소만 실패했다. 늦은 A 발행이 성공한 B를 덮어쓰지 않는지를 후보 개수뿐 아니라 후보 값으로 확인했다.

## Baseline-attribution

- root: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/moai-proxy-unified`
- branch: `WT-unified-gateway`
- HEAD: `81c1d58f9cf7045594ee61d5e4ff380948ce9eba`
- 이 HEAD 위 미커밋 수리본을 직접 측정했고 receipt 파일의 시작/종료 hash가 같았다. 다른 writer의 auth/Windows/native policy 파일은 측정하거나 변경하지 않았다.
- 근거로 `receipt-core-repair-verification.md`와 실제 core.go/store.go 및 새 회귀 시험을 읽었다. 구현자의 GREEN을 독립 PASS로 사용하지 않았다.

### 반복 이력

iter1의 RC-F1~F3 FAIL 및 RC-F4 optional을 보존한다. 이번 iter2는 이전 overlay와 수리 회귀 시험을 직접 재실행하고 precommit guard 경계를 추가했다. RC-F2~F4는 해결되었고 RC-F1은 마지막 guard 취소 사례가 남아 FAIL을 유지한다. 감사에서 제품 수리는 수행하지 않았다.

## Gaps

- 전체 receipt 패키지와 cross-process lock 전체 시험을 다시 실행하지 않고 결함 delta와 영향을 받는 경계만 시험했다. 전체 85% 커버리지 기준을 이번 결과로 재판정하지 않는다.
- Check의 모든 observation을 검증된 history에서 순서대로 구성하는 caller 전제는 그대로다. 역순 observation 수용을 전체 provenance PASS로 해석하지 않는다.
- 임의의 같은 UID filesystem 조작과 완전히 일관된 snapshot rollback을 모두 막는다고 주장하지 않는다. guard 검사와 rename은 단일 원자 CAS가 아니다.
- post-rename directory fsync 실패, 실제 provider/native resume, root 연결, Windows native 실행, 전체 CI는 이번 범위 밖이다.
- 실제 credential Store나 다른 세션의 live 자료를 사용하지 않았다. 커밋·push·PR·배포·SPEC 변경을 하지 않았다.

## Residual-risk

이번 수리의 lock/root/snapshot 확인은 관측된 stale publication을 막는다. 다만 마지막 guard의 filesystem I/O 중 context가 취소되는 경계를 Rename 앞에서 다시 확인해야 precommit 취소 계약에 맞는다. durable commit 이후 client 취소와 commit 이전 취소를 혼동하여 이미 완료한 receipt를 삭제해서는 안 된다. 이번 FAIL은 남은 한 경계에 한정한다.

## Evidence appendix

### Independent delta execution

```sh
unset ANTHROPIC_BASE_URL ANTHROPIC_API_KEY ANTHROPIC_AUTH_TOKEN Z_AI_API_KEY && GOCACHE=/tmp/gateway-foundation-cache go test -overlay=/var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/receipt-reaudit-g200solm/overlay.json -race ./internal/gateway/receipt -run '^TestReceiptAudit|^TestWriteCancellation|^TestPrecommitCancellation|^TestLockReplacement|^TestSnapshotReplacement|^TestEmptyEnvelopeAmbiguity|^TestZeroParent' -count=1 -timeout 30s -v -coverprofile=/var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/receipt-reaudit-g200solm/coverage.out
```

exit_code: 1

```text
=== RUN   TestReceiptAuditIter2FinalGuardCancellation
    receipt_audit_iter2_test.go:7: final_guard_cancel err=<nil> ctx=context canceled candidates=1
    receipt_audit_iter2_test.go:7: cancellation during final precommit guard still published
--- FAIL: TestReceiptAuditIter2FinalGuardCancellation (0.05s)
=== RUN   TestReceiptAuditIter2LateLockReplacementPreservesB
=== RUN   TestReceiptAuditIter2LateLockReplacementPreservesB/1
    receipt_audit_iter2_test.go:12: guard=1 A rejected=true B preserved=true
=== RUN   TestReceiptAuditIter2LateLockReplacementPreservesB/2
    receipt_audit_iter2_test.go:12: guard=2 A rejected=true B preserved=true
=== RUN   TestReceiptAuditIter2LateLockReplacementPreservesB/3
    receipt_audit_iter2_test.go:12: guard=3 A rejected=true B preserved=true
--- PASS: TestReceiptAuditIter2LateLockReplacementPreservesB (0.08s)
    --- PASS: TestReceiptAuditIter2LateLockReplacementPreservesB/1 (0.02s)
    --- PASS: TestReceiptAuditIter2LateLockReplacementPreservesB/2 (0.03s)
    --- PASS: TestReceiptAuditIter2LateLockReplacementPreservesB/3 (0.03s)
=== RUN   TestReceiptAuditCancelBeforeNonceWrite
    receipt_independent_audit_test.go:8: publish_err=context canceled ctx_err=context canceled persisted_candidates=0
--- PASS: TestReceiptAuditCancelBeforeNonceWrite (0.01s)
=== RUN   TestReceiptAuditLockReplacementNoLostPublication
    receipt_independent_audit_test.go:13: first_err=invalid or inaccessible private receipt state second_err=<nil> persisted_candidates=1
--- PASS: TestReceiptAuditLockReplacementNoLostPublication (0.02s)
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
=== RUN   TestWriteCancellationBeforeAndAfterCommit/after-rename
--- PASS: TestWriteCancellationBeforeAndAfterCommit (0.04s)
    --- PASS: TestWriteCancellationBeforeAndAfterCommit/before-rename (0.02s)
    --- PASS: TestWriteCancellationBeforeAndAfterCommit/after-rename (0.02s)
=== RUN   TestEmptyEnvelopeAmbiguityAcrossProviders
--- PASS: TestEmptyEnvelopeAmbiguityAcrossProviders (0.00s)
=== RUN   TestZeroParentCannotCreateFork
--- PASS: TestZeroParentCannotCreateFork (0.00s)
FAIL
coverage: 74.2% of statements
FAIL	github.com/modu-ai/moai-adk/internal/gateway/receipt	0.639s
FAIL
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
python3 /var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/receipt-reaudit-g200solm/mechanical.py
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
no-follow/private/lock/fsync: internal/gateway/receipt/store.go:227, internal/gateway/receipt/store.go:268, internal/gateway/receipt/platform_posix.go:15, internal/gateway/receipt/platform_posix.go:21, internal/gateway/receipt/platform_posix.go:27, internal/gateway/receipt/platform_posix.go:34, internal/gateway/receipt/platform_posix.go:50
manifest diff lines: 0
gofmt exit=0 output=''
core.go 83/114 72.81%
platform_posix.go 14/23 60.87%
projection.go 139/189 73.54%
store.go 137/177 77.40%
81c1d58f9cf7045594ee61d5e4ff380948ce9eba
WT-unified-gateway
```

### Baseline SHA-256

```json
{
  "internal/gateway/receipt/store_test.go": "0abe83fe53042fc2f6d4de670e0e6a195c47c0697dcecca041e0dae4a066ae0c",
  "internal/gateway/receipt/lock_posix_test.go": "fbe45b46148d9898ecc42f0deb85aff5d2872709f00d7178bde12e9fec1bc796",
  "internal/gateway/receipt/store.go": "2c1effbfec0e202bec93f590debfdcd9edb80119cab023fb2e2cd2dfeb3627ff",
  "internal/gateway/receipt/core_test.go": "fcd58ea16b6757e70cdfb0940ed2d6b18d2497633d3c375ed56c8233476c491d",
  "internal/gateway/receipt/repair_posix_test.go": "ec87ff4894dcb9bdf02269283e27522129d60580033be1e975b2e40ca41983d3",
  "internal/gateway/receipt/core.go": "81f81009bf65030145f8a1dec5084e36a1cf68bf7643f36da83c8f89d0f18cfd",
  "internal/gateway/receipt/repair_test.go": "d02141f5a1ade3b2067dc3b967c0447ba13a5b02eea829781e2f9d9ea0035e07",
  "internal/gateway/receipt/platform_windows.go": "97295c3aed156477bb0539d6e10100f621d19989c391c74a86313e8a38097f34",
  "internal/gateway/receipt/projection.go": "e2cae7cd97095b3dda24e8ccfbb2f87400042f0f0321019fed972031db571cf9",
  "internal/gateway/receipt/platform_posix.go": "0a10595022e38e84043f8074b9dd5c67b31ee8219bfac9e13ecd103d054f6486"
}
```

### Added reproducer

Source: `/var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/receipt-reaudit-g200solm/receipt_iter2_test.go`

SHA-256: `82675326affdb88ba00026935067265b59c79962fe3fea2a07826efb8dfd5e4c`

```go
//go:build !windows
package receipt
import("context";"testing";"errors";"os";"path/filepath";"fmt")
func TestReceiptAuditIter2FinalGuardCancellation(t *testing.T){
 ctx,cancel:=context.WithCancel(context.Background());defer cancel();s,e:=OpenStore(ctx,privateDir(t),testUUID,true);if e!=nil{t.Fatal(e)};defer s.Close()
 e=s.transaction(ctx,false,func(guard func()error)error{m,e:=s.read();if e!=nil{return e};raw,e:=m.Marshal();if e!=nil{return e};expected:=Hash(raw);m.Publish(candidate("p","prior","A"));calls:=0;return s.write(ctx,m,func()error{calls++;if calls==3{cancel()};return guard()},&expected)})
 m,readErr:=s.Snapshot(context.Background());if readErr!=nil{t.Fatal(readErr)};t.Logf("final_guard_cancel err=%v ctx=%v candidates=%d",e,ctx.Err(),len(m.Candidates()));if !errors.Is(e,context.Canceled)||len(m.Candidates())!=0{t.Fatal("cancellation during final precommit guard still published")}
}
func TestReceiptAuditIter2LateLockReplacementPreservesB(t *testing.T){
 for _,at:=range []int{1,2,3}{t.Run(fmt.Sprint(at),func(t *testing.T){ctx:=context.Background();dir:=privateDir(t);a,e:=OpenStore(ctx,dir,testUUID,true);if e!=nil{t.Fatal(e)};defer a.Close();b,e:=OpenStore(ctx,dir,testUUID,false);if e!=nil{t.Fatal(e)};defer b.Close();next:=candidate("p","prior","B")
 e=a.transaction(ctx,false,func(guard func()error)error{m,e:=a.read();if e!=nil{return e};raw,e:=m.Marshal();if e!=nil{return e};expected:=Hash(raw);m.Publish(candidate("p","prior","A"));calls:=0;return a.write(ctx,m,func()error{calls++;if calls==at{if e:=os.Rename(filepath.Join(dir,".lock"),filepath.Join(dir,".lock-old"));e!=nil{return e};if e:=os.WriteFile(filepath.Join(dir,".lock"),nil,0600);e!=nil{return e};if e:=b.Publish(ctx,next);e!=nil{return e}};return guard()},&expected)})
 m,readErr:=b.Snapshot(ctx);if readErr!=nil{t.Fatal(readErr)};if e==nil||len(m.Candidates())!=1||m.Candidates()[0]!=next{t.Fatalf("guard=%d firsterr=%v candidates=%d B preserved=false",at,e,len(m.Candidates()))};t.Logf("guard=%d A rejected=true B preserved=true",at)
 })}
}

```

### Remote comparison

```sh
git fetch -q origin main
git rev-list --count --left-right origin/main...HEAD
```

```text
0	2879
```
