# Receipt core·POSIX Store 독립 기능 감사

## Claim

SPEC: SPEC-MOAI-GATEWAY-001. Overall Verdict: **FAIL — Functionality 필수 통과 조건 미충족**. 평가 점수: **71.75/100**.

새 receipt 패키지에서 blocking 3건과 optional 1건을 재현했다. 제품 파일은 수정하지 않았다. 이 판정은 core/POSIX Store에 한정하며 실제 gateway 통합·native resume·Windows를 승인하지 않는다.

| Dimension | Score | Verdict | Evidence (관측 원문) |
|---|---:|---|---|
| Functionality (40%) | 50/100 | FAIL | `first_err=<nil> second_err=<nil> persisted_candidates=1` |
| Security (25%) | 75/100 | PASS, Critical/High 미발견 | `required_openai_plus_plain_anthropic empty_envelope_err=<nil>` — Medium 결함 |
| Craft (20%) | 90/100 | PASS, receipt 패키지 한정 | `coverage: 89.1% of statements` |
| Consistency (15%) | 100/100 | PASS | `gofmt exit=0 output=''` |

기본 evaluator profile의 필수 통과 조건을 적용했다. 기능 실패가 전체 FAIL을 강제하며, 점수는 저장소 전체의 품질 수치가 아니다.

### Findings

- **RC-F1 [Medium/P2] [blocking] [confidence high] `internal/gateway/receipt/store.go:175` — temporary write 이전의 취소를 놓친다.** 실제 nonce 생성 Reader에서 context를 취소했는데 Publish가 nil을 반환하고 완료 후보 1개를 저장했다. design receipt 항목 5의 취소·불완전 응답은 완료 receipt를 만들지 않는 경계에 어긋난다. **Required fix:** write에 context를 전달하고 nonce 이후·temporary write 전과 rename commit 전의 취소를 판정한다. commit 전 취소는 오류와 이전 snapshot 보존으로 끝나야 한다. durable commit 뒤 client가 취소된 경우의 기록 보존은 별개다.
- **RC-F2 [Medium/P2] [blocking] [confidence high] `internal/gateway/receipt/store.go:141`, `:215` — lock inode 교체 중 성공 publication을 잃는다.** A가 lock을 검증한 뒤 nonce 생성 중 `.lock`을 새 inode로 바꾸고 B Store가 발행하도록 했다. 두 Publish가 모두 nil인데 최종 후보는 1개였다. **Required fix:** 잠근 lock identity를 commit/readback까지 보존·검증하고 변경이 있으면 stale snapshot으로 덮어쓰거나 성공을 반환하지 않는다. 필요하면 현재 manifest generation/snapshot도 대조한다. 동일 probe에서 두 성공 후보가 모두 남거나 충돌 발행이 명시 실패해야 한다. 같은 UID의 완전한 snapshot rollback까지 막으라는 요구로 확대하지 않는다.
- **RC-F3 [Medium/P2] [blocking] [confidence high] `internal/gateway/receipt/core.go:106` — provider 선필터로 빈 envelope의 혼재 검사를 우회한다.** 같은 Prefix/Previous에 OpenAI required=true A와 Anthropic required=false를 저장하고 provider=anthropic/빈 envelope를 주면 Check가 nil이다. 승인된 **RPA-2**는 같은 공개 prefix의 required=true/false 혼재에서 빈 envelope를 모호성 오류로 판정하며, 같은 provider라는 제한은 없다. **Required fix:** 빈 envelope의 required 존재 판정은 동일 Prefix/Previous 전체에서 수행하여 provider 선택으로 누락을 정당화하지 못하게 한다. nonempty의 정확한 provider/digest/item 수 검사와 정상 foreign receipt 지원은 유지한다.
- **RC-F4 [Low/P3] [optional] [confidence high, 영향 제한] `internal/gateway/receipt/core.go:124` — 초기화하지 않은 parent의 Fork를 수용한다.** `var zero Manifest; zero.Fork(validUUID)`가 오류 없이 빈 child manifest를 만든다. 정상 Store Snapshot 경로에서 zero parent를 받은 사례는 관측하지 않아 이 항목만으로 FAIL을 내리지 않는다. **Optional fix:** 초기화된 parent인지 확인하고 zero-value Fork를 거절한다.

### Recommendations

수리는 RC-F1~F3에 한정하고 RC-F4는 선택 사항으로 처리한다. 아래 nonce interleaving overlay를 그대로 보존했으며 수리 후 같은 재현과 영향받는 기존 시험으로 재감사한다.

## Evidence

명령과 관측 원문은 아래 Evidence appendix에 함께 보존한다. 기존 package 시험은 PASS였지만 독립 API probe에서 RC-F1~F3가 실패했다.

독립 시험은 실제 package API와 감사 소유 temporary directory를 사용했다. `crypto/rand.Reader`를 합성 callback으로 감싸 정확한 nonce 생성 지점에 취소·파일 교체·다른 Store 발행을 넣었다. 난수 자체는 기존 Reader에서 읽으며 t.Cleanup으로 원 Reader를 복원했다. product write/lock 함수를 fake로 바꾸지 않았고 background load를 만들지 않았다.

정규화 matrix에서는 큰 인접 정수 구분, 의미가 같은 소수·지수 정규화, Unicode composed/decomposed 구분, role·배열·tool ID·error flag 변경과 decoder 오류를 확인했다. 부분 manifest 손실은 거절되었고 일관된 전체 snapshot rollback은 명시된 한계대로 수용되었다. 발행 직전 root 교체는 거절되며 이전 manifest가 보존되었다.

기존 receipt 패키지 시험에는 A/B·동일 중복·required=false·동일 provider 혼재·불완전 후보·4096개/8MiB, checksum·fork, Unicode/JSON, 12개 동시 발행, private permission/symlink/hardlink/root replacement, process lock 취소·강제 종료 회수, write 실패의 이전 상태 보존이 포함된다. process helper는 CommandContext와 Kill/Wait cleanup으로 제한되어 있다.

## Baseline-attribution

- root: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/moai-proxy-unified`
- branch: `WT-unified-gateway`
- HEAD: `81c1d58f9cf7045594ee61d5e4ff380948ce9eba`
- 이 HEAD 위 미커밋 receipt 8개 파일을 측정했다. 시작/종료 hash는 같았으며 아래 appendix에 기록했다.
- nonce overlay: `/var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/receipt-independent-audit-9wppdm_t/overlay.json`
- 실제 읽은 계획 근거: core design receipt/RPA-2/RPA-A1와 `receipt-plan-audit-iter2.md`. 구현 근거: `receipt-core-verification.md`. 이전 실행 수치를 이번 측정치로 사용하지 않았다.

### 반복 이력

첫 독립 명령은 overlay 경로를 찾는 shell substitution에서 지연되어 시험 결과가 나오지 않았다. 감사 소유 shell과 자식만 종료한 뒤 정확한 overlay 경로로 다시 실행했다. 이를 제품 실패나 성공 근거로 세지 않았다. 최종 독립 명령에서 RC-F1~F3가 실패했고 제품 파일 변경 없이 이 판정을 기록한다.

## Gaps

- **Caller provenance:** CanonicalPrefixes는 사전 semantic/opaque 검증과 strict tool ID decoder를 요구한다. Check는 caller가 제공한 모든 Prefix/Previous·Provider·Opaque·Items를 신뢰한다. 실제로 역순 Observation 두 개도 nil을 반환했다. 이는 history 원문을 다시 검증하는 API가 아니라는 관측이다. 통합 caller가 모든 assistant 경계를 순서대로 빠짐없이 검증된 codec 결과와 연결해야 한다. 이 전제 자체를 전체 제품 보장으로 사용하지 않는다.
- private bootstrap UUID/metadata 비교, owned index, retained profile/family lease, native transcript identity, fork Store 초기화, terminal 이전 Publish 연결, foreign stripping을 감사하지 않았다.
- RC-F2는 같은 UID가 private lock 파일을 교체할 수 있는 조건이다. 다른 UID의 root 접근 가능성을 입증하지 않는다. 가능한 모든 filesystem race를 전수 검사한 것은 아니다.
- post-rename directory fsync/readback 실패는 독립 fault injection으로 재현하지 않았다. rename 뒤 오류에서 파일이 이미 교체되었을 수 있다는 제한은 유지한다.
- Windows는 ErrUnsupported 코드 경계만 읽었다. 이번에는 cross-compile도 반복하지 않았으며 native ACL/lock/durability PASS를 주장하지 않는다.
- 실제 provider/native resume/기존 credential Store/사용자 대화 원문/전체 repository CI·CVE·배포는 실행하거나 접근하지 않았다.

## Residual-risk

receipt는 hash 일관성과 유실 탐지 자료다. 공급자 서명이나 같은 UID의 완전한 snapshot 교체 방어가 아니다. 이 한계가 취소 전 완료 저장·두 성공 publication의 유실·RPA-2의 빈 envelope 혼재 수용을 정당화하지는 않는다. 세 blocking 항목을 수리한 뒤 실제 caller 연결은 별도 판정을 받아야 한다.

## Evidence appendix

### Independent probes

```sh
unset ANTHROPIC_BASE_URL ANTHROPIC_API_KEY ANTHROPIC_AUTH_TOKEN Z_AI_API_KEY && GOCACHE=/tmp/gateway-foundation-cache go test -overlay=/var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/receipt-independent-audit-9wppdm_t/overlay.json -race ./internal/gateway/receipt -run '^TestReceiptAudit' -count=1 -timeout 30s -v
```

exit_code: 1

```text
=== RUN   TestReceiptAuditCancelBeforeNonceWrite
    receipt_independent_audit_test.go:8: publish_err=<nil> ctx_err=context canceled persisted_candidates=1
    receipt_independent_audit_test.go:8: cancellation before temporary write still published completion
--- FAIL: TestReceiptAuditCancelBeforeNonceWrite (2.73s)
=== RUN   TestReceiptAuditLockReplacementNoLostPublication
    receipt_independent_audit_test.go:13: first_err=<nil> second_err=<nil> persisted_candidates=1
    receipt_independent_audit_test.go:13: two acknowledged publications lost a candidate after lock replacement
--- FAIL: TestReceiptAuditLockReplacementNoLostPublication (4.58s)
=== RUN   TestReceiptAuditCanonicalMeaningMatrix
--- PASS: TestReceiptAuditCanonicalMeaningMatrix (0.00s)
=== RUN   TestReceiptAuditForkZeroAndCallerBoundary
    receipt_independent_audit_test.go:24: zero-value parent Fork accepted candidates=0
    receipt_independent_audit_test.go:26: caller-supplied reverse observation order check_err=<nil> (caller owns prefix derivation)
--- PASS: TestReceiptAuditForkZeroAndCallerBoundary (0.00s)
=== RUN   TestReceiptAuditPartialManifestLossAndRollback
    receipt_independent_audit_test.go:30: partial corruption rejected; internally consistent full snapshot rollback accepted as documented
--- PASS: TestReceiptAuditPartialManifestLossAndRollback (2.28s)
=== RUN   TestReceiptAuditRootSwapBeforeWriteFails
--- PASS: TestReceiptAuditRootSwapBeforeWriteFails (3.03s)
=== RUN   TestReceiptAuditCrossProviderRequiredAmbiguity
    receipt_independent_audit_test.go:38: required_openai_plus_plain_anthropic empty_envelope_err=<nil>
    receipt_independent_audit_test.go:38: same-prefix mixed required=true/false accepted empty envelope after provider selection
--- FAIL: TestReceiptAuditCrossProviderRequiredAmbiguity (0.00s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/gateway/receipt	13.746s
FAIL
```

### Existing package tests

```sh
unset ANTHROPIC_BASE_URL ANTHROPIC_API_KEY ANTHROPIC_AUTH_TOKEN Z_AI_API_KEY ANTHROPIC_AUTH_TOKEN Z_AI_API_KEY && GOCACHE=/tmp/gateway-foundation-cache go test -race -coverprofile=/var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/receipt-independent-audit-9wppdm_t/coverage.out ./internal/gateway/receipt -count=1 -timeout 45s
```

exit_code: 0

```text
ok  	github.com/modu-ai/moai-adk/internal/gateway/receipt	32.191s	coverage: 89.1% of statements
```

### Mechanical checks

```sh
python3 /var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/receipt-independent-audit-9wppdm_t/mechanical.py
```

exit_code: 0

```text
baseline hashes unchanged=True
shell/env/log sinks: no matches
no-follow/private/lock/fsync: internal/gateway/receipt/store.go:207, internal/gateway/receipt/store.go:223, internal/gateway/receipt/platform_posix.go:15, internal/gateway/receipt/platform_posix.go:21, internal/gateway/receipt/platform_posix.go:27, internal/gateway/receipt/platform_posix.go:34, internal/gateway/receipt/platform_posix.go:50
manifest diff lines: 0
gofmt exit=0 output=''
core.go 95/110 86.36%
platform_posix.go 20/23 86.96%
projection.go 177/189 93.65%
store.go 127/148 85.81%
81c1d58f9cf7045594ee61d5e4ff380948ce9eba
WT-unified-gateway
```

### Lint

```sh
unset ANTHROPIC_BASE_URL ANTHROPIC_API_KEY ANTHROPIC_AUTH_TOKEN Z_AI_API_KEY && GOCACHE=/tmp/gateway-foundation-cache go vet ./internal/gateway/receipt
```

stdout/stderr: empty

```text
go vet exit=0
```

### Mechanical check source

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

### Baseline SHA-256

```json
{
  "internal/gateway/receipt/store_test.go": "0abe83fe53042fc2f6d4de670e0e6a195c47c0697dcecca041e0dae4a066ae0c",
  "internal/gateway/receipt/lock_posix_test.go": "3ded2855d586c9bda61a402bae68a9a63a615823646a9818813bc4205f4e592f",
  "internal/gateway/receipt/store.go": "50932606f7df1fec46be36a24e94e995c7ed94c85d4d5baf592ff7652f5c642a",
  "internal/gateway/receipt/core_test.go": "fcd58ea16b6757e70cdfb0940ed2d6b18d2497633d3c375ed56c8233476c491d",
  "internal/gateway/receipt/core.go": "7e7ed0d906584dfdb886971f988133119504cec45dfaf12389967a867da551c3",
  "internal/gateway/receipt/platform_windows.go": "97295c3aed156477bb0539d6e10100f621d19989c391c74a86313e8a38097f34",
  "internal/gateway/receipt/projection.go": "e2cae7cd97095b3dda24e8ccfbb2f87400042f0f0321019fed972031db571cf9",
  "internal/gateway/receipt/platform_posix.go": "0a10595022e38e84043f8074b9dd5c67b31ee8219bfac9e13ecd103d054f6486"
}
```

### Reproducer

Source: `/var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/receipt-independent-audit-9wppdm_t/receipt_audit_test.go`

SHA-256: `70002ba979ad46fb18b11eadefcea2331b4ad3cbb215239c48868d59f8faa45c`

```go
//go:build !windows
package receipt
import("context";cryptorand "crypto/rand";"io";"os";"path/filepath";"testing";"strings";"encoding/json";"errors";"fmt")
type auditNonceHook struct{source io.Reader;callback func()}
func(r *auditNonceHook)Read(b []byte)(int,error){if r.callback!=nil{cb:=r.callback;r.callback=nil;cb()};return r.source.Read(b)}
func TestReceiptAuditCancelBeforeNonceWrite(t *testing.T){
 dir:=privateDir(t);s,e:=OpenStore(context.Background(),dir,testUUID,true);if e!=nil{t.Fatal(e)};defer s.Close();ctx,cancel:=context.WithCancel(context.Background());defer cancel();original:=cryptorand.Reader;t.Cleanup(func(){cryptorand.Reader=original});cryptorand.Reader=&auditNonceHook{source:original,callback:cancel}
 e=s.Publish(ctx,candidate("p","prior","opaque"));cryptorand.Reader=original;m,readErr:=s.Snapshot(context.Background());if readErr!=nil{t.Fatal(readErr)};t.Logf("publish_err=%v ctx_err=%v persisted_candidates=%d",e,ctx.Err(),len(m.Candidates()));if e==nil||len(m.Candidates())!=0{t.Fatal("cancellation before temporary write still published completion")}
}
func TestReceiptAuditLockReplacementNoLostPublication(t *testing.T){
 dir:=privateDir(t);s,e:=OpenStore(context.Background(),dir,testUUID,true);if e!=nil{t.Fatal(e)};defer s.Close();other,e:=OpenStore(context.Background(),dir,testUUID,false);if e!=nil{t.Fatal(e)};defer other.Close();original:=cryptorand.Reader;t.Cleanup(func(){cryptorand.Reader=original});var concurrentErr error
 cryptorand.Reader=&auditNonceHook{source:original,callback:func(){cryptorand.Reader=original;if e:=os.Rename(filepath.Join(dir,".lock"),filepath.Join(dir,".lock-replaced"));e!=nil{t.Fatal(e)};if e:=os.WriteFile(filepath.Join(dir,".lock"),nil,0600);e!=nil{t.Fatal(e)};concurrentErr=other.Publish(context.Background(),candidate("p","prior","B"))}}
 firstErr:=s.Publish(context.Background(),candidate("p","prior","A"));cryptorand.Reader=original;m,e:=s.Snapshot(context.Background());if e!=nil{t.Fatal(e)};t.Logf("first_err=%v second_err=%v persisted_candidates=%d",firstErr,concurrentErr,len(m.Candidates()));if firstErr==nil&&concurrentErr==nil&&len(m.Candidates())!=2{t.Fatal("two acknowledged publications lost a candidate after lock replacement")}
}
func TestReceiptAuditCanonicalMeaningMatrix(t *testing.T){
 decode:=func(s string)(string,error){if strings.HasPrefix(s,"bound:"){return strings.TrimPrefix(s,"bound:"),nil};if strings.HasPrefix(s,"invalid:"){return "",ErrInvalid};return s,nil}
 shape:=func(num,text,id string)string{return fmt.Sprintf(`[{"role":"user","content":%q},{"role":"assistant","content":[{"type":"tool_use","id":%q,"name":"echo","input":{"n":%s,"a":[true,null,"1"]}}]},{"role":"user","content":[{"type":"tool_result","tool_use_id":%q,"content":"out","is_error":false}]},{"role":"assistant","content":"done"}]`,text,id,num,id)}
 prefix:=func(raw string)Boundary{b,e:=CanonicalPrefixes([]byte(raw),decode);if e!=nil{t.Fatal(e)};return b[len(b)-1]}
 base:=prefix(shape("9007199254740993","é","id"));for _,raw:=range []string{shape("9007199254740993.0","é","bound:id"),shape("90071992547409930e-1","é","id")}{if prefix(raw)!=base{t.Fatal("equivalent public input diverged")}}
 for _,raw:=range []string{shape("9007199254740992","é","id"),shape("9007199254740993","é","id"),shape("9007199254740993","é","other"),strings.Replace(shape("9007199254740993","é","id"),`[true,null,"1"]`,`[null,true,"1"]`,1),strings.Replace(shape("9007199254740993","é","id"),`"is_error":false`,`"is_error":true`,1),strings.Replace(shape("9007199254740993","é","id"),`"role":"user"`,`"role":"system"`,1)}{if prefix(raw)==base{t.Fatal("changed public meaning collided")}}
 if _,e:=CanonicalPrefixes([]byte(shape("1","text","invalid:x")),decode);e==nil{t.Fatal("invalid tool ID decoder error ignored")}
}
func TestReceiptAuditForkZeroAndCallerBoundary(t *testing.T){
 var zero Manifest;fork,e:=zero.Fork(testUUID);if e==nil{t.Logf("zero-value parent Fork accepted candidates=%d",len(fork.Candidates()))}else{t.Logf("zero-value parent Fork rejected err=%v",e)}
 m,_:=New(testUUID);a:=candidate("A","start","a");b:=candidate("B","A","b");m.Publish(a);m.Publish(b);toObs:=func(c Candidate)Observation{return Observation{Prefix:c.Prefix,Previous:c.Previous,Provider:c.Provider,Opaque:c.Opaque,Items:c.Items}}
 e=m.Check(testUUID,[]Observation{toObs(b),toObs(a)});t.Logf("caller-supplied reverse observation order check_err=%v (caller owns prefix derivation)",e)
}
func TestReceiptAuditPartialManifestLossAndRollback(t *testing.T){
 dir:=privateDir(t);s,e:=OpenStore(context.Background(),dir,testUUID,true);if e!=nil{t.Fatal(e)};defer s.Close();s.Publish(context.Background(),candidate("A","start","opaqueA"));first,e:=os.ReadFile(filepath.Join(dir,"manifest.json"));if e!=nil{t.Fatal(e)};s.Publish(context.Background(),candidate("B","A","opaqueB"));latest,e:=os.ReadFile(filepath.Join(dir,"manifest.json"));if e!=nil{t.Fatal(e)}
 var fields map[string]json.RawMessage;json.Unmarshal(latest,&fields);fields["candidates"]=json.RawMessage("[]");bad,_:=json.Marshal(fields);os.WriteFile(filepath.Join(dir,"manifest.json"),bad,0600);if _,e=s.Snapshot(context.Background());e==nil{t.Fatal("partial manifest loss accepted")};os.WriteFile(filepath.Join(dir,"manifest.json"),first,0600);m,e:=s.Snapshot(context.Background());if e!=nil||len(m.Candidates())!=1{t.Fatal("full rollback documented boundary changed",e)};t.Log("partial corruption rejected; internally consistent full snapshot rollback accepted as documented")
 os.Remove(filepath.Join(dir,".lock"));if _,e=OpenStore(context.Background(),dir,testUUID,false);e==nil{t.Fatal("resume recreated missing lock")};if _,e=os.Lstat(filepath.Join(dir,".lock"));!errors.Is(e,os.ErrNotExist){t.Fatal("missing lock created")}
}
func TestReceiptAuditRootSwapBeforeWriteFails(t *testing.T){
 dir:=privateDir(t);s,e:=OpenStore(context.Background(),dir,testUUID,true);if e!=nil{t.Fatal(e)};defer s.Close();moved:=dir+"-audit-old";t.Cleanup(func(){os.RemoveAll(moved)});original:=cryptorand.Reader;t.Cleanup(func(){cryptorand.Reader=original});cryptorand.Reader=&auditNonceHook{source:original,callback:func(){if e:=os.Rename(dir,moved);e!=nil{t.Fatal(e)};if e:=os.Mkdir(dir,0700);e!=nil{t.Fatal(e)}}};e=s.Publish(context.Background(),candidate("A","start","a"));cryptorand.Reader=original;if e==nil{t.Fatal("root identity swap accepted")};raw,e:=os.ReadFile(filepath.Join(moved,"manifest.json"));if e!=nil{t.Fatal(e)};m,e:=Parse(raw,testUUID);if e!=nil||len(m.Candidates())!=0{t.Fatal("root swap changed prior manifest")}
}

func TestReceiptAuditCrossProviderRequiredAmbiguity(t *testing.T){
 m,_:=New(testUUID);required:=candidate("same-public","same-prior","opaqueA");plain:=required;plain.Provider="anthropic";plain.Required=false;plain.Opaque=Digest{};plain.Items=0;m.Publish(required);m.Publish(plain);e:=m.Check(testUUID,[]Observation{{Prefix:plain.Prefix,Previous:plain.Previous,Provider:plain.Provider}});t.Logf("required_openai_plus_plain_anthropic empty_envelope_err=%v",e);if e==nil{t.Fatal("same-prefix mixed required=true/false accepted empty envelope after provider selection")}
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
