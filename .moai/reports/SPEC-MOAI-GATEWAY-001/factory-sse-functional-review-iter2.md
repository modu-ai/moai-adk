# Factory·SSE F1/F2 독립 재감사 — iter2

## Claim

SPEC: SPEC-MOAI-GATEWAY-001
Overall Verdict: **PASS — F1/F2 수리 delta 한정**, 점수 **98/100**.

이전 `factory-sse-functional-review.md`의 F1(CRLF 실제 byte 계수)과 F2(private JSON alias 덮어쓰기)를 독립 시험으로 재검증하여 해결로 판정한다. 제품 코드·SPEC는 수정하지 않았다. 범위 밖의 factory/native policy/receipt/Windows/전체 커버리지 게이트를 이번 PASS로 대체하지 않는다.

| Dimension | Score | Verdict | Evidence (관측 원문) |
|---|---:|---|---|
| Functionality (40%) | 100/100 | PASS, F1/F2 한정 | `wire boundary matrix cases=72` |
| Security (25%) | 100/100 | PASS, F2 및 수리분 한정 | `--- PASS: TestGatewayFactoryAuditCaseAlias (0.00s)` |
| Craft (20%) | 90/100 | PASS, 수정 블록 한정 | `F1/F2 contained coverage blocks: 8/9 = 88.89%` |
| Consistency (15%) | 100/100 | PASS, 검사 파일 한정 | `gofmt exit=0 output=''` |

기본 evaluator profile의 Functionality/Security 필수 통과 조건을 delta에 적용했다. 이전 전체 변경의 Craft UNVERIFIED를 전체 PASS로 승격한 것이 아니다. 아래 커버리지 분모는 정확한 수정 범위 안에 포함되는 블록만 세었고, 전체 함수·패키지는 별도 한계로 둔다.

### Findings

- **F1 [기존 Medium/P2, blocking] — RESOLVED, confidence high.** `internal/gateway/translate/stream.go:77`의 split 함수가 `ScanLines`의 원문 advance를 계수한다. 이전 1447/1446 CRLF 실패 probe가 이제 성공 terminal 없이 `stream exceeds output byte limit`으로 끝났다. LF·CRLF/일반·구독 모드와 1·2·3·7·64·4096 byte 수신 분할, 실제 길이 대비 -1/0/+1 상한의 72개 사례를 확인했다. 개행 없는 부분 줄의 상한 초과도 같은 오류로 끝났다.
- **F2 [기존 Medium/P2, blocking] — RESOLVED, confidence high.** `internal/cli/gateway_factory.go:86`에서 최상위 키를 decode된 정확한 세 키로 제한한다. 이전 version/session_token/model_ids 대소문자 alias 실패 probe가 모두 거절된다. 단독 alias·앞뒤 중복·동일 중복의 Store 개방 전 거절 및 canonical payload의 resource 양성 대조군도 통과했다. 추가 독립 시험에서 escaped canonical key는 받아들이고, 같은 decode 결과를 갖는 escape 중복·비정규 대소문자·긴 s 문자는 거절했다.

새 blocking/optional 결함은 이번 delta 범위에서 재현되지 않았다. Required fix: 없음. 제품 경로 소유권을 돌려주며, 범위 밖 기능 판단은 기존 별도 게이트를 따른다.

## Evidence

### Functionality·Security: 이전 overlay와 추가 경계 시험

명령 (exit_code: 0):

```sh
unset ANTHROPIC_BASE_URL ANTHROPIC_API_KEY ANTHROPIC_AUTH_TOKEN Z_AI_API_KEY && GOCACHE=/tmp/gateway-foundation-cache go test -overlay=/var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/factory-sse-reaudit-5cboytmz/overlay.json -race ./internal/cli ./internal/gateway/translate -run 'TestGatewayFactoryAudit|TestSubscriptionStreamAudit|TestAuditIter2|TestGatewayFactoryExactPayloadKeysBeforeStore|TestStreamRawWireByteBoundaries' -count=1 -timeout 30s -v -coverprofile=/var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/factory-sse-reaudit-5cboytmz/coverage.out
```

관측 원문:

```text
=== RUN   TestAuditIter2DecodedKeyIdentity
--- PASS: TestAuditIter2DecodedKeyIdentity (0.00s)
=== RUN   TestGatewayFactoryAuditCaseAlias
--- PASS: TestGatewayFactoryAuditCaseAlias (0.00s)
=== RUN   TestGatewayFactoryAuditFrozenDependenciesAndSubset
--- PASS: TestGatewayFactoryAuditFrozenDependenciesAndSubset (0.00s)
=== RUN   TestGatewayFactoryAuditConcurrentClose
--- PASS: TestGatewayFactoryAuditConcurrentClose (0.00s)
=== RUN   TestGatewayFactoryAuditRejectBeforeGPTStore
--- PASS: TestGatewayFactoryAuditRejectBeforeGPTStore (0.00s)
=== RUN   TestGatewayFactoryExactPayloadKeysBeforeStore
--- PASS: TestGatewayFactoryExactPayloadKeysBeforeStore (0.00s)
PASS
coverage: 5.9% of statements
ok  	github.com/modu-ai/moai-adk/internal/cli	3.287s	coverage: 5.9% of statements
=== RUN   TestAuditIter2WireLimitChunkMatrix
    audit_iter2_boundary_test.go:9: wire boundary matrix cases=72
--- PASS: TestAuditIter2WireLimitChunkMatrix (0.11s)
=== RUN   TestAuditIter2UnterminatedLineEnforcesRawBudget
--- PASS: TestAuditIter2UnterminatedLineEnforcesRawBudget (0.00s)
=== RUN   TestSubscriptionStreamAuditCRLFByteLimit
    sse_independent_audit_test.go:5: wire_bytes=1447 limit=1446 terminal=false err=upstream SSE read: stream exceeds output byte limit
--- PASS: TestSubscriptionStreamAuditCRLFByteLimit (0.00s)
=== RUN   TestSubscriptionStreamAuditStrictStateAndJSON
=== RUN   TestSubscriptionStreamAuditStrictStateAndJSON/wrong_response_id
=== RUN   TestSubscriptionStreamAuditStrictStateAndJSON/output_object
=== RUN   TestSubscriptionStreamAuditStrictStateAndJSON/duplicate_terminal_field
--- PASS: TestSubscriptionStreamAuditStrictStateAndJSON (0.00s)
    --- PASS: TestSubscriptionStreamAuditStrictStateAndJSON/wrong_response_id (0.00s)
    --- PASS: TestSubscriptionStreamAuditStrictStateAndJSON/output_object (0.00s)
    --- PASS: TestSubscriptionStreamAuditStrictStateAndJSON/duplicate_terminal_field (0.00s)
=== RUN   TestSubscriptionStreamAuditWriterErrorClosesUpstream
--- PASS: TestSubscriptionStreamAuditWriterErrorClosesUpstream (0.00s)
=== RUN   TestSubscriptionStreamAuditCRLFBaselineControl
    sse_independent_audit_test.go:19: subscription=false wire_bytes=1598 limit=1597 terminal=false err=upstream SSE read: stream exceeds output byte limit
    sse_independent_audit_test.go:20: subscription=false LF negative control err=upstream SSE read: stream exceeds output byte limit
    sse_independent_audit_test.go:19: subscription=true wire_bytes=1447 limit=1446 terminal=false err=upstream SSE read: stream exceeds output byte limit
    sse_independent_audit_test.go:20: subscription=true LF negative control err=upstream SSE read: stream exceeds output byte limit
--- PASS: TestSubscriptionStreamAuditCRLFBaselineControl (0.01s)
=== RUN   TestStreamRawWireByteBoundaries
--- PASS: TestStreamRawWireByteBoundaries (0.05s)
PASS
coverage: 61.3% of statements
ok  	github.com/modu-ai/moai-adk/internal/gateway/translate	1.583s	coverage: 61.3% of statements
```

기존 감사 source를 덮어쓰지 않고 그대로 포함한 overlay에 추가 경계 source 두 개를 붙였다. 새 검사만 실행하여 과거 실패를 생략하지 않았다. 신규 제품 회귀 시험도 같은 실행에 포함했다. 실제 provider·기존 credential Store 접근은 없었다.

### Craft: 변경 블록 실행과 정적 검사

명령: `python3 /var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/factory-sse-reaudit-5cboytmz/coverage_delta.py`

```python
from pathlib import Path
p=Path(__file__).parent
covered=total=0
for line in (p/'coverage.out').read_text().splitlines()[1:]:
 loc,n,c=line.split();f,region=loc.rsplit(':',1);start=int(region.split('.')[0]);end=int(region.split(',')[1].split('.')[0]);n=int(n);c=int(c)
 if f.endswith('stream.go') and start>=77 and end<=84 or f.endswith('gateway_factory.go') and start>=87 and end<=96:
  print(line);total+=n;covered+=n*(c>0)
print(f'F1/F2 contained coverage blocks: {covered}/{total} = {100*covered/total:.2f}%')
```

관측 원문:

```text
github.com/modu-ai/moai-adk/internal/gateway/translate/stream.go:77.67,79.69 2 73824
github.com/modu-ai/moai-adk/internal/gateway/translate/stream.go:79.69,81.4 1 39
github.com/modu-ai/moai-adk/internal/gateway/translate/stream.go:82.3,83.29 2 73785
github.com/modu-ai/moai-adk/internal/cli/gateway_factory.go:87.42,89.4 1 0
github.com/modu-ai/moai-adk/internal/cli/gateway_factory.go:90.3,90.27 1 30
github.com/modu-ai/moai-adk/internal/cli/gateway_factory.go:90.27,91.15 1 75
github.com/modu-ai/moai-adk/internal/cli/gateway_factory.go:92.49,92.49 0 51
github.com/modu-ai/moai-adk/internal/cli/gateway_factory.go:93.12,94.34 1 24
F1/F2 contained coverage blocks: 8/9 = 88.89%
```

분모는 split 함수 내부와 exact-key 분기의 범위 안에 완전히 들어오는 profile block이다. 새 Map unmarshal 오류 return 1개는 실행되지 않았다. 전체 함수 커버리지는 이 수치와 다르며 newGatewayHandlerFactory 73.9%, stream 76.8%였다. 이번 명령이 모든 기존 분기를 선택하지 않았으므로 이전 측정치를 대신 넣지 않았다.

명령 (exit_code: 0):

```sh
unset ANTHROPIC_BASE_URL ANTHROPIC_API_KEY ANTHROPIC_AUTH_TOKEN Z_AI_API_KEY && GOCACHE=/tmp/gateway-foundation-cache go vet ./internal/cli ./internal/gateway/translate
```

관측 원문 (go vet의 stdout/stderr는 비었고 종료 코드만 shell에서 표시):

```text
go vet exit=0
```

### Security·Consistency·baseline 기계 검사

명령: `python3 /var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/factory-sse-reaudit-5cboytmz/mechanical.py`

```python
from pathlib import Path
import json,hashlib,re,subprocess
p=Path(__file__).parent;baseline=json.loads((p/'baseline.json').read_text())
for f in ['internal/cli/gateway_factory.go','internal/gateway/translate/stream.go']:
 pattern=r'\b(?:exec\.Command|os\.ReadFile|os\.Getenv|log\.|slog\.|fmt\.Print)'
 hits=[str(i) for i,l in enumerate(Path(f).read_text().splitlines(),1) if re.search(pattern,l)]
 print(f'{f} shell/storage/log: '+(','.join(hits) if hits else 'no matches'))
print('manifest diff lines:',len(subprocess.check_output(['git','diff','--','go.mod','go.sum'],text=True).splitlines()))
fmt=subprocess.run(['gofmt','-l']+[f for f in baseline if f.endswith('.go')],capture_output=True,text=True);print(f'gofmt exit={fmt.returncode} output={fmt.stdout!r}')
print('baseline hashes unchanged='+str(all(hashlib.sha256(Path(f).read_bytes()).hexdigest()==h for f,h in baseline.items())))
for i,l in enumerate(Path('internal/cli/root.go').read_text().splitlines(),1):
 if 'newGatewayChildCommand' in l:print(f'root.go:{i}: {l.strip()}')
for cmd in [['git','rev-parse','HEAD'],['git','branch','--show-current']]:print(subprocess.check_output(cmd,text=True).strip())
```

마지막 관측 원문:

```text
internal/cli/gateway_factory.go shell/storage/log: no matches
internal/gateway/translate/stream.go shell/storage/log: no matches
manifest diff lines: 0
gofmt exit=0 output=''
baseline hashes unchanged=True
root.go:156: rootCmd.AddCommand(newGatewayChildCommand(nil))
81c1d58f9cf7045594ee61d5e4ff380948ce9eba
WT-unified-gateway
```

패턴 무검출은 검사 파일·패턴 범위에 한정한다. 전체 OWASP/CVE/로그 안전을 추가로 승인한 것은 아니다.

## Baseline-attribution

- 실제 root: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/moai-proxy-unified`
- branch: `WT-unified-gateway`
- HEAD: `81c1d58f9cf7045594ee61d5e4ff380948ce9eba`
- 이번 HEAD 위 미커밋 수리본을 직접 측정했다. 시작/종료 파일 hash가 같음을 확인했다.
- `git fetch -q origin main; git rev-list --count --left-right origin/main...HEAD` 관측: `0	2879`.
- root 등록의 실제 관측은 `newGatewayChildCommand(nil)`이다.

```json
{
  "internal/cli/gateway_factory.go": "15917f81924580527f468cb2e36509c0b786996b59a0677f8f563876b3ddcc8e",
  "internal/cli/gateway_factory_test.go": "f72c526235b1aa1d02e6ad181b6dffc957b3de4ace52b34e6277ea94aabbbf2b",
  "internal/gateway/translate/stream.go": "a9b1699eb6f624f56e9a9158c38f01ad6f3a29c899e41129cc5653536aa636c5",
  "internal/gateway/translate/subscription_stream_test.go": "cf695894c19046b2d6ff72c266064e3dbb2be50270b3d4d31a9989f1c9b2dbd4",
  "internal/cli/root.go": "b61caf1b37cca4b25490ba63940ad5577cf69aa93a604e935cf9f36f721b4456",
  "go.mod": "530783a72da61a17948ed525b77d8d62bf1b387c3cbe3ce2bef59aa549030eb0",
  "go.sum": "334c75e10a71d3931a8e084a3136907ecfb92099fe52acafd5883350bb5ed730"
}
```

추가 감사 source:

- `/var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/factory-sse-reaudit-5cboytmz/factory_boundary_test.go`, SHA-256 `f7f897af4467f6b23169ed9a6cc09395ce25133367f5dbf715dc96bd8160f09f`
- `/var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/factory-sse-reaudit-5cboytmz/sse_boundary_test.go`, SHA-256 `0b8948b9a97a867d03803972e542488bfd0caf0b3b01766d81618f723ef2e655`
- `/var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/factory-sse-reaudit-5cboytmz/overlay.json`은 이전 독립 source 경로와 위 추가 source 경로를 매핑한다.

### 반복 이력

1. iter1: 같은 HEAD의 수리 전 미커밋 구현에서 F1/F2를 재현하여 FAIL을 기록했다. 원문과 hash는 `factory-sse-functional-review.md`에 보존했다.
2. 수리 worker의 `factory-sse-repair-verification.md`를 읽고 수리한 두 구현 블록과 회귀 시험을 확인했다. 해당 보고서의 GREEN을 이번 감사 실행으로 주장하지 않았다.
3. iter2: 이전 overlay를 직접 재실행하고 새 72개 수신 분할/상한 경계, 부분 줄, decoded JSON key identity를 확인했다. 모두 통과했고 제품 source hash가 유지되어 F1/F2를 RESOLVED로 판정한다.

## Gaps

- 전체 factory/native policy/receipt 구현·root 연결·실제 provider·실계정 갱신·native title/CLI 지원을 감사하지 않았다. 다른 세션의 별도 실행이나 새로운 사용자 정책 결정을 이 delta 근거에 섞지 않았다.
- Windows native 검증과 전체 repository CI는 이번에 실행하지 않았다. 이전 별도 게이트의 판단을 변경하지 않는다.
- 실제 MoAI home의 openGPTAuthStore, 기존 credential Store, 별도 실계정 관측 자료에 접근하지 않았다.
- 전체 패키지 커버리지(cli 5.9%, translate 61.3%)는 선택 시험 분모다. 전체 85% 충족 또는 이전 Craft GAP 해소를 주장하지 않는다.
- CR-only 이벤트 구문, terminal 뒤 서비스 종료까지의 추가 byte 소비는 새 계약으로 추가하지 않았다.
- 커밋·push·PR·배포·queue 변경을 수행하지 않았다.

## Residual-risk

raw advance 계수는 parser가 소비한 줄까지를 세며 terminal 이후의 read-ahead를 별도 메시지 처리로 세지 않는다. 이번 경계 시험은 이 계약 안에서 실제 수신 길이를 검증한다. 출력 token 정책과 일반 capability 지원은 별도다. private payload의 key 검증이 caller가 제공하는 모든 정책 metadata의 실제 지원까지 증명하지는 않는다. 이번 PASS는 열거한 두 blocking 결함의 수리 수용으로만 사용한다.
