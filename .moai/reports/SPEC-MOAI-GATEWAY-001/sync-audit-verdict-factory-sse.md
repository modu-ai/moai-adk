# Private factory·구독 SSE 독립 기능 감사

## Claim

SPEC: SPEC-MOAI-GATEWAY-001 (private child factory 및 design §4.3 구독 SSE 변경 한정)
Overall Verdict: **FAIL — Functionality 필수 통과 조건 미충족**.
평가 점수: **66.25/100**. 이 점수는 범위 한정 감사 평가이며 제품 전체 품질·커버리지 수치가 아니다.

검증된 blocking 결함 2건이 있다. 범용 Stream에도 재현되는 CRLF byte 계수 오류가 새 SubscriptionStream 경로에 남아 있으며, 새 private factory는 대소문자만 다른 JSON 필드를 같은 필드의 덮어쓰기로 받아들인다. 제품 파일은 수정하지 않았다.

| Dimension | Score | Verdict | Evidence (관측 원문) |
|---|---:|---|---|
| Functionality (40%) | 50/100 | FAIL | `wire_bytes=1447 limit=1446 terminal=true err=<nil>` |
| Security (25%) | 85/100 | PASS, Critical/High 미발견 | `manifest diff lines: 0` 및 아래 인증·설정 경계 재현 출력 |
| Craft (20%) | 50/100 | UNVERIFIED, 전체 변경 커버리지 판정 유보 | `openGPTAuthStore 0.0%` (아래 go cover 원문) |
| Consistency (15%) | 100/100 | PASS, 검사 파일 한정 | `gofmt exit=0 output=''` |

기본 evaluator profile을 사용했다. Functionality 실패가 전체 FAIL을 강제한다. Craft는 좁은 선택 시험의 낮은 패키지 분모를 전체 구현 결함으로 해석하지 않으며, 추출된 helper를 이번 명령으로 실행하지 못한 점을 UNVERIFIED로 둔다.

### Findings

- **F1 [Medium/P2] [blocking] [confidence: high] `internal/gateway/translate/stream.go:86` — CRLF를 실제 수신 byte보다 적게 계산한다.** `Scanner.Text()`는 줄 끝의 CRLF를 제거하는데 계수는 `len(line)+1`만 더한다. 1447 byte인 유효한 구독 SSE에 1446 byte 상한을 주어도 err=nil이고 `message_stop`이 나온다. 같은 실행의 LF 초과 대조군은 오류로 끝났다. design §4.3의 크기 검증 및 구현 보고서의 byte limit 유지 주장을 충족하지 못한다. **Required fix:** 정규화 전 수신 byte를 세는 제한 reader 또는 동등한 정확 계수로 상한을 검사하고, LF·CRLF 각각의 상한 직전/정확/초과를 시험한다. 상한을 초과하면 성공 terminal을 내보내지 않아야 한다. outgoing token 상한의 정책·값을 바꾸는 수정은 이 결함의 해법이 아니다.
- **F2 [Medium/P2] [blocking] [confidence: high] `internal/cli/gateway_factory.go:85` — private payload의 의미상 중복 필드를 허용한다.** `DisallowUnknownFields`를 사용해도 encoding/json은 대소문자별 필드명을 같은 struct 필드에 매칭한다. `version:0` 뒤 `Version:1`, `session_token:first` 뒤 `SESSION_TOKEN:last`, `model_ids:[unknown]` 뒤 `MODEL_IDS:[claude-opus-5]`가 모두 유효 handler를 만든다. strict private payload의 정확한 세 필드·잘못된 version/model·중복 거부 경계를 우회한다. 다만 뒤의 값도 approved catalog 검사를 받으므로 임의 upstream이나 미승인 모델 접근을 입증한 것은 아니다. **Required fix:** struct decode 전에 최상위 키를 정확히 `version`, `session_token`, `model_ids`로 제한하고 다른 철자·대소문자 alias를 거부한다. 또는 같은 의미의 중복을 명시적으로 거부하는 동등한 최소 변경을 적용한다. 세 alias 사례와 단독 비정규 키를 resource open 이전에 거절하는 회귀 시험을 둔다.

### Recommendations

수정 범위는 F1의 수신 byte 계수와 F2의 private payload 키 검증으로 제한한다. 재감사는 두 실패 probe와 영향을 받는 기존 시험으로 진행한다. root 활성화, 출력 cap 정책 변경, Windows 수용 결정은 이번 수리 범위에 포함하지 않는다.

## Evidence

### Functionality·Security: 독립 adversarial overlay

명령 (exit_code: 1):

```sh
unset ANTHROPIC_BASE_URL ANTHROPIC_API_KEY ANTHROPIC_AUTH_TOKEN Z_AI_API_KEY && GOCACHE=/tmp/gateway-foundation-cache go test -overlay=/var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/factory-sse-audit-znc5n4it/overlay.json -race ./internal/cli ./internal/gateway/translate -run 'TestGatewayFactoryAudit|TestSubscriptionStreamAudit' -count=1 -timeout 30s -v
```

관측 원문:

```text
=== RUN   TestGatewayFactoryAuditCaseAlias
    factory_independent_audit_test.go:5: semantic duplicate JSON field accepted: {"version":0,"Version":1,"session_token":"private","model_ids":["claude-opus-5"]}
    factory_independent_audit_test.go:5: semantic duplicate JSON field accepted: {"version":1,"session_token":"first","SESSION_TOKEN":"last","model_ids":["claude-opus-5"]}
    factory_independent_audit_test.go:5: semantic duplicate JSON field accepted: {"version":1,"session_token":"private","model_ids":["unknown"],"MODEL_IDS":["claude-opus-5"]}
--- FAIL: TestGatewayFactoryAuditCaseAlias (0.00s)
=== RUN   TestGatewayFactoryAuditFrozenDependenciesAndSubset
--- PASS: TestGatewayFactoryAuditFrozenDependenciesAndSubset (0.00s)
=== RUN   TestGatewayFactoryAuditConcurrentClose
--- PASS: TestGatewayFactoryAuditConcurrentClose (0.00s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	1.488s
=== RUN   TestSubscriptionStreamAuditCRLFByteLimit
    sse_independent_audit_test.go:5: wire_bytes=1447 limit=1446 terminal=true err=<nil>
    sse_independent_audit_test.go:5: CRLF wire-byte limit exceeded but success emitted
--- FAIL: TestSubscriptionStreamAuditCRLFByteLimit (0.00s)
=== RUN   TestSubscriptionStreamAuditStrictStateAndJSON
=== RUN   TestSubscriptionStreamAuditStrictStateAndJSON/output_object
=== RUN   TestSubscriptionStreamAuditStrictStateAndJSON/duplicate_terminal_field
=== RUN   TestSubscriptionStreamAuditStrictStateAndJSON/wrong_response_id
--- PASS: TestSubscriptionStreamAuditStrictStateAndJSON (0.00s)
    --- PASS: TestSubscriptionStreamAuditStrictStateAndJSON/output_object (0.00s)
    --- PASS: TestSubscriptionStreamAuditStrictStateAndJSON/duplicate_terminal_field (0.00s)
    --- PASS: TestSubscriptionStreamAuditStrictStateAndJSON/wrong_response_id (0.00s)
=== RUN   TestSubscriptionStreamAuditWriterErrorClosesUpstream
--- PASS: TestSubscriptionStreamAuditWriterErrorClosesUpstream (0.00s)
=== RUN   TestSubscriptionStreamAuditCRLFBaselineControl
    sse_independent_audit_test.go:19: subscription=false wire_bytes=1598 limit=1597 terminal=true err=<nil>
    sse_independent_audit_test.go:20: subscription=false LF negative control err=stream exceeds output byte limit
    sse_independent_audit_test.go:19: subscription=true wire_bytes=1447 limit=1446 terminal=true err=<nil>
    sse_independent_audit_test.go:20: subscription=true LF negative control err=stream exceeds output byte limit
--- PASS: TestSubscriptionStreamAuditCRLFBaselineControl (0.00s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/gateway/translate	0.432s
FAIL
```

같은 묶음에서 dependency snapshot 불변, catalog 부분집합, 미선택 GPT Store 미개방, 동시 request/Close 경쟁, malformed JSON·terminal ID·output 타입 거절, downstream writer 실패 때 upstream Close는 통과했다. F1과 F2 때문에 묶음 전체는 실패했다. baseline control의 PASS는 오류 수정 완료가 아니라 일반 Stream에도 같은 계수 오류가 있음을 기록하는 관측 시험의 성공이다.

### Resource 개방 전 검증: 양성 대조군 포함

명령 (exit_code: 0):

```sh
unset ANTHROPIC_BASE_URL ANTHROPIC_API_KEY ANTHROPIC_AUTH_TOKEN Z_AI_API_KEY && GOCACHE=/tmp/gateway-foundation-cache go test -overlay=/var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/factory-sse-audit-znc5n4it/overlay.json ./internal/cli -run '^TestGatewayFactoryAuditRejectBeforeGPTStore$' -count=1 -timeout 30s -v
```

관측 원문:

```text
=== RUN   TestGatewayFactoryAuditRejectBeforeGPTStore
--- PASS: TestGatewayFactoryAuditRejectBeforeGPTStore (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	1.009s
```

GPT 행을 실제로 선택한 설정에서 보통의 잘못된 version·unknown field·중복 model은 OpenStore 호출 0, 유효 payload는 모의 OpenStore 호출 1이라는 양성 대조군을 사용했다. 실제 credential Store나 broker에는 접근하지 않았다.

### 기존 집중 시험 및 Craft

명령 (exit_code: 0):

```sh
unset ANTHROPIC_BASE_URL ANTHROPIC_API_KEY ANTHROPIC_AUTH_TOKEN Z_AI_API_KEY && GOCACHE=/tmp/gateway-foundation-cache go test -race ./internal/cli ./internal/gateway ./internal/gateway/translate -run '^TestGatewayFactory|^TestGatewayChild|^TestSubscriptionStream|^TestOpenAISubscription|^TestStream' -count=1 -timeout 45s -coverprofile=/var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/factory-sse-audit-znc5n4it/coverage.out
```

관측 원문:

```text
ok  	github.com/modu-ai/moai-adk/internal/cli	3.474s	coverage: 6.0% of statements
ok  	github.com/modu-ai/moai-adk/internal/gateway	2.011s	coverage: 10.8% of statements
ok  	github.com/modu-ai/moai-adk/internal/gateway/translate	1.575s	coverage: 77.0% of statements
```

기존 시험에는 private startup 실패 정리, 실제 mock Store 소유권/초기화 실패 정리/로그인 부재 무송신, TLS mock OAuth/GLM 고정 목적지, sparse terminal 재구성/일반 Stream 비확장/잘못된 상태/UTF-8/취소가 포함된다. 이 기존 시험의 통과가 두 독립 재현 실패를 상쇄하지 않는다.

명령:

```sh
unset ANTHROPIC_BASE_URL ANTHROPIC_API_KEY ANTHROPIC_AUTH_TOKEN Z_AI_API_KEY && GOCACHE=/tmp/gateway-foundation-cache go tool cover -func=/var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/factory-sse-audit-znc5n4it/coverage.out | rg 'gateway_factory.go|openGPTAuthStore|openAIResponseMedia|SubscriptionStream|stream.go.*(stream|event|doneItem|layout)'
unset ANTHROPIC_BASE_URL ANTHROPIC_API_KEY ANTHROPIC_AUTH_TOKEN Z_AI_API_KEY && GOCACHE=/tmp/gateway-foundation-cache go vet ./internal/cli ./internal/gateway ./internal/gateway/translate
```

관측 원문:

```text
github.com/modu-ai/moai-adk/internal/cli/gateway_factory.go:40:			newGatewayHandlerFactory		87.2%
github.com/modu-ai/moai-adk/internal/cli/gateway_factory.go:178:		ServeHTTP				100.0%
github.com/modu-ai/moai-adk/internal/cli/gateway_factory.go:194:		Close					100.0%
github.com/modu-ai/moai-adk/internal/cli/gpt_auth.go:19:			openGPTAuthStore			0.0%
github.com/modu-ai/moai-adk/internal/gateway/openai.go:210:			openAIResponseMedia			100.0%
github.com/modu-ai/moai-adk/internal/gateway/translate/stream.go:58:		SubscriptionStream			100.0%
github.com/modu-ai/moai-adk/internal/gateway/translate/stream.go:62:		stream					92.3%
github.com/modu-ai/moai-adk/internal/gateway/translate/stream.go:206:		event					94.0%
github.com/modu-ai/moai-adk/internal/gateway/translate/stream.go:437:		doneItem				86.7%
github.com/modu-ai/moai-adk/internal/gateway/translate/stream.go:496:		layout					100.0%
go vet exit=0
```

### Security·Consistency·baseline 기계 검사

명령: `python3 /var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/factory-sse-audit-znc5n4it/mechanical.py`

```python
from pathlib import Path
import json,hashlib,re,subprocess
p=Path(__file__).parent;baseline=json.loads((p/'baseline.json').read_text())
for f in ['internal/cli/gateway_factory.go','internal/cli/gateway_child.go','internal/gateway/openai.go','internal/gateway/translate/stream.go']:
 for label,pattern in [('shell/storage/log',r'\b(?:exec\.Command|os\.ReadFile|os\.Getenv|log\.|slog\.|fmt\.Print)')]:
  hits=[str(i) for i,l in enumerate(Path(f).read_text().splitlines(),1) if re.search(pattern,l)];print(f'{f} {label}: '+(','.join(hits) if hits else 'no matches'))
print('manifest diff lines:',len(subprocess.check_output(['git','diff','--','go.mod','go.sum'],text=True).splitlines()))
fmt=subprocess.run(['gofmt','-l']+[f for f in baseline if f.endswith('.go')],capture_output=True,text=True);print(f'gofmt exit={fmt.returncode} output={fmt.stdout!r}')
print('baseline hashes unchanged='+str(all(hashlib.sha256(Path(f).read_bytes()).hexdigest()==h for f,h in baseline.items())))
for i,l in enumerate(Path('internal/cli/root.go').read_text().splitlines(),1):
 if 'newGatewayChildCommand' in l:print(f'root.go:{i}: {l.strip()}')
for cmd in [['git','rev-parse','HEAD'],['git','branch','--show-current']]:print(subprocess.check_output(cmd,text=True).strip())
```

관측 원문:

```text
internal/cli/gateway_factory.go shell/storage/log: no matches
internal/cli/gateway_child.go shell/storage/log: no matches
internal/gateway/openai.go shell/storage/log: no matches
internal/gateway/translate/stream.go shell/storage/log: no matches
manifest diff lines: 0
gofmt exit=0 output=''
baseline hashes unchanged=True
root.go:156: rootCmd.AddCommand(newGatewayChildCommand(nil))
81c1d58f9cf7045594ee61d5e4ff380948ce9eba
WT-unified-gateway
```

무검출은 해당 파일·해당 패턴의 결과다. 전체 저장소 비밀·CVE 또는 전체 OWASP 준수 판정이 아니다. 모델·헤더·Store 소유권 경계는 위 실행 시험으로 따로 확인했다. `gpt_auth.go`는 기존 broker 실행을 포함하므로 새 helper의 파일 경로 추출과 함께 소스만 읽었고, 전체 broker 동작을 이번 변경의 새 송신 경로로 오인하지 않았다.

## Baseline-attribution

- 실제 root: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/moai-proxy-unified`
- branch: `WT-unified-gateway`
- HEAD: `81c1d58f9cf7045594ee61d5e4ff380948ce9eba`
- 이 HEAD 위의 미커밋 구현을 측정했다. 여러 세션의 변경이 존재하므로 HEAD만으로 재현할 수 없다. 감사 대상 파일의 시작/종료 hash는 일치했다.
- 작성 전 remote 비교: `git fetch origin main`; `git rev-list --count --left-right origin/main...HEAD`. 관측 비교: `0	2879`.
- F1은 새 SubscriptionStream과 같은 tree의 기존 Stream 모두에서 재현했다. 따라서 구독 모드에만 새로 생긴 회귀로 귀속하지 않는다. 과거 commit의 최초 도입 시점은 측정하지 않았다. 이번에 추가한 구독 경로도 현재 크기 검증 요구를 충족해야 하므로 blocking으로 분류한다.
- F2는 새 factory의 JSON decode에서 재현했다. resource 권한 상승·임의 endpoint 송신을 관측한 것은 아니다.

대상 SHA-256:

```json
{
  "internal/cli/gateway_factory.go": "1669b2fc87e83e8e0cc4de9f8079aebc74352c551c620aaaaf68b42235b71974",
  "internal/cli/gateway_factory_test.go": "faf340b08d61bf6c4104246afbe84acd9cbd38e38ee97fc46c05b00883dfb218",
  "internal/cli/gateway_child.go": "09f9e409814ac9e40b846b6f43bd1b6288cf03262f078e523741a022c5c09c31",
  "internal/cli/gpt_auth.go": "4faed304311daf6b07b7d834727bcf1135edae332a406831f089cb6f40ad6719",
  "internal/gateway/openai.go": "9274c055b50cdeb4cc21936d00b55aeac14796902fdeb59c44551e88ae69764e",
  "internal/gateway/openai_subscription_test.go": "1f8e7f14aea6ff5e334568f1a9b340430122ed3d1c3382ba49e7a22f850fa0c7",
  "internal/gateway/translate/stream.go": "34157d4336c22b285f183082ffe6e91eef290dbbc424b60b99d4d48804df4414",
  "internal/gateway/translate/subscription_stream_test.go": "4b6af6484b162f8e2e92fb64867036e28d5ba3e3340d48e5c363d9d821e13165",
  "internal/cli/root.go": "b61caf1b37cca4b25490ba63940ad5577cf69aa93a604e935cf9f36f721b4456",
  "go.mod": "530783a72da61a17948ed525b77d8d62bf1b387c3cbe3ce2bef59aa549030eb0",
  "go.sum": "334c75e10a71d3931a8e084a3136907ecfb92099fe52acafd5883350bb5ed730"
}
```

독립 probe는 Go overlay로만 넣었다. 제품 폴더의 Go 파일은 변경하지 않았다. 임시 자료:

- `/var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/factory-sse-audit-znc5n4it/overlay.json`
- `/var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/factory-sse-audit-znc5n4it/factory_audit_test.go` SHA-256 `5994628c0e663a0113ecfba764ed1a0c62849d1682bf1d8150f27e83fa2c67dd`
- `/var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/factory-sse-audit-znc5n4it/sse_audit_test.go` SHA-256 `046570e9f9970d74a225700a5ed66527a6c1b95d7570f1b0ead24c3c401471ed`

### 반복 이력

1. 기존 보고서 `child-factory-verification.md`, `subscription-sse-verification.md`와 참조 구현을 읽었다. 두 보고서의 이전 GREEN 수치를 이번 실행의 baseline으로 재사용하지 않았다.
2. 최초 독립 probe에서 F1/F2 외에 두 fixture 문제가 있었다. terminal ID 변형은 fixture의 실제 `resp-1` 대신 존재하지 않는 `r`을 치환하여 원본을 그대로 보냈다. 또 최상위 response의 추가 필드는 현재 계약에서 금지하지 않았는데 이를 실패로 단언했다. 전자는 올바른 ID를 변형하도록 수정했고 후자는 근거 없는 단언을 제거했다. 두 항목을 제품 결함으로 계산하지 않았다.
3. 최종 독립 probe에서 F1/F2만 실패했다. 일반 Stream/LF 음성 대조군, resource 개방 양성 대조군 및 기존 집중 시험을 추가 실행했다. 제품 변경 없이 이 FAIL 판정을 내보낸다.

## Gaps

- production root/launcher의 연결, 실제 native policy·beta·MeasureInput·opaque receipt, 실계정 GPT/OAuth 호출과 갱신은 실행하지 않았다.
- `openGPTAuthStore`의 실제 MoAI home 경로 helper는 이번 선택 시험에서 0.0%다. 기존 보고서의 77.8%를 재측정치로 주장하지 않는다. 기존 credential Store를 수정하지 않기 위해 이번 감사에서 실제 home을 열지 않았다.
- 패키지 커버리지(cli 6.0%, gateway 10.8%, translate 77.0%)는 선택 시험 분모이며 전체 변경·저장소의 85% 기준 충족을 주장하지 않는다.
- Windows native Store, 전체 저장소 CI, 의존성 CVE 조회, 장기 부하 시험은 수행하지 않았다.
- factory Close의 검증은 협조적 request context와 실제 child의 서버 종료 순서를 대상으로 했다. 임의 custom handler의 영구 블로킹까지 보장한다고 주장하지 않는다.
- outgoing max_output_tokens 계약·정책·상한은 변경하지 않았다. 전체 native readiness는 별도 감사 범위다.

## Residual-risk

F1은 수신 byte 예산 검사가 정규화된 줄 길이에 의존하는 문제다. 더 큰 설정값으로 우회하면 검증이 강화되지 않으므로 정확 계수로 고쳐야 한다. F2는 private 입력의 의미상 충돌이며 approved catalog 자체를 깨지는 못하지만 설정을 읽는 코드 사이의 해석을 다르게 만들 수 있다. 이번 판정은 이 두 경계의 수리와 확인을 요구하며 제품 활성화 승인을 제공하지 않는다.
