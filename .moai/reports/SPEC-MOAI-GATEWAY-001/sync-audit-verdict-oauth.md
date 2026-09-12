# OAuth passthrough 독립 기능 감사

## Claim

SPEC: SPEC-MOAI-GATEWAY-001, REQ-MG-016 및 요청 인증·provider 격리의 OAuth 변경 범위.
Overall Verdict: **PASS — OAuth 구성품 한정**, 점수 **99/100**.

이번 판정은 새 요청 단위 OAuth credential, router 연결, 별도 adapter 생성자와 관련 ingress 경계의 독립 감사다. 전체 SPEC 수용·native Claude 정책 지원·root factory 활성화·제품 출시를 판정하지 않는다. 기본 evaluator profile의 네 차원과 Functionality/Security 필수 통과 조건을 이 한정 범위에 적용했다. SPEC에 evaluator_profile은 없으며 `.moai/config/evaluator-profiles/default.md`를 읽었다. 별도 audit_model 설정은 `.moai/config/sections` 검색에서 찾지 못하여 외부 backend 판정은 사용하지 않았다.

| Dimension | Score | Verdict | Evidence (관측 원문) |
|---|---:|---|---|
| Functionality (40%) | 100/100 | PASS, 한정 범위 | `--- PASS: TestOAuthAuditRequestCompletionLifetime (0.00s)` |
| Security (25%) | 100/100 | PASS, 한정 범위 | `--- PASS: TestOAuthAuditWireDuplicatesAndAuthBeforeBody (0.00s)` |
| Craft (20%) | 95/100 | PASS, 아래 분모 한정 | `OAuth credential + router credential + OAuth constructor statements: 62/65 = 95.38%` |
| Consistency (15%) | 100/100 | PASS, 변경 파일 한정 | `gofmt exit=0 output=''` |

표의 시간 포함 출력은 아래 실행 로그 원문과 함께 해석한다. 점수는 범위 내 감사 평가이며 저장소 전체 품질 수치가 아니다. Craft 분모는 새 credential 파일 전체, router credential 함수 전체, OAuth 생성자다. 기존 adapter의 전체 parser·stream·policy와 server/catalog 전체 분모는 포함하지 않았다.

다음 동작을 실제 시험에서 확인했다.

- 요청의 헤더를 복사한 뒤 원본을 바꾸어도 Bearer가 바뀌지 않는다. 동시 두 요청의 자격 증명이 섞이지 않는다.
- 실제 HTTP 수신에서 대소문자가 다른 Authorization 중복과 x-api-key 동시 입력을 거절한다. 별도 session header 중복도 401로 끝난다.
- session 인증 실패 시 Content-Length만 전송하고 본문은 보내지 않아도 401을 반환한다. body를 기다리거나 adapter에 도달하지 않는다.
- 실제 HTTP 요청 종료 후 보관해 둔 credential의 Generation과 Apply가 context.Canceled로 실패한다. 다음 요청에 Bearer가 없으면 이전 토큰을 재사용하지 않는다.
- GPT·GLM·Anthropic API key 경로는 inbound Bearer가 여러 개여도 저장소 resolver의 참조를 사용하며, adapter에 인증 헤더를 전달하지 않는다.
- OAuth adapter까지 잇는 시험에서 upstream 401은 401/authentication_error가 되고 한 번만 송신한다. 307은 502로 바뀌며 redirect를 따라가지 않고 응답의 인증·Location·비공개 본문을 전달하지 않는다.
- 전송 중 취소는 upstream 연결 종료로 이어진다. 정확한 endpoint 이외의 port 표기·fragment·userinfo·suffix host·escaped path·다른 method·nil 요청은 Apply에서 거절한다. 출력 헤더의 대소문자별 Authorization/x-api-key도 하나의 Bearer로 정리된다.

### Findings

제품 코드에서 이번 범위를 막는 결함을 재현하지 못했다. blocking 0, optional 0이다. 검증하지 않은 전체 기능은 아래 Gaps에 남겼으며 결함으로 추정하지 않았다.

### Recommendations

후속 제품 통합 판정은 factory와 native 정책 지원의 별도 게이트 증거를 갖춘 뒤 수행한다. 이번 구성품 PASS를 그 판정에 대신 사용하지 않는다.

## Evidence

### 기능 및 보안 경계 — race 포함 집중 시험

명령:

```sh
unset ANTHROPIC_BASE_URL ANTHROPIC_API_KEY ANTHROPIC_AUTH_TOKEN Z_AI_API_KEY && GOCACHE=/tmp/gateway-foundation-cache go test -overlay=/var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/oauth-independent-audit-6yil1ch3/overlay.json -race ./internal/gateway/auth ./internal/gateway -run 'TestAnthropicOAuth|TestOAuth|TestSessionAuthPrecedesBodyAndPath|TestIngressGenerationAndProviderGuard' -count=1 -timeout 30s -v -coverprofile=/var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/oauth-independent-audit-6yil1ch3/coverage.out
```

exit_code: 0. 관측 원문:

```text
=== RUN   TestAnthropicOAuthRequestBoundary
--- PASS: TestAnthropicOAuthRequestBoundary (0.00s)
=== RUN   TestAnthropicOAuthMalformedPaddingAndContext
--- PASS: TestAnthropicOAuthMalformedPaddingAndContext (0.00s)
PASS
coverage: 6.5% of statements
ok  	github.com/modu-ai/moai-adk/internal/gateway/auth	1.384s	coverage: 6.5% of statements
=== RUN   TestOAuthAuditProviderIsolation
=== RUN   TestOAuthAuditProviderIsolation/gpt-audit
=== RUN   TestOAuthAuditProviderIsolation/glm-audit
=== RUN   TestOAuthAuditProviderIsolation/key-audit
--- PASS: TestOAuthAuditProviderIsolation (0.00s)
    --- PASS: TestOAuthAuditProviderIsolation/gpt-audit (0.00s)
    --- PASS: TestOAuthAuditProviderIsolation/glm-audit (0.00s)
    --- PASS: TestOAuthAuditProviderIsolation/key-audit (0.00s)
=== RUN   TestOAuthAuditWireDuplicatesAndAuthBeforeBody
--- PASS: TestOAuthAuditWireDuplicatesAndAuthBeforeBody (0.00s)
=== RUN   TestOAuthAuditRequestCompletionLifetime
--- PASS: TestOAuthAuditRequestCompletionLifetime (0.00s)
=== RUN   TestOAuthAuditAdapterEndToEnd401AndRedirect
=== RUN   TestOAuthAuditAdapterEndToEnd401AndRedirect/401
=== RUN   TestOAuthAuditAdapterEndToEnd401AndRedirect/307
--- PASS: TestOAuthAuditAdapterEndToEnd401AndRedirect (0.02s)
    --- PASS: TestOAuthAuditAdapterEndToEnd401AndRedirect/401 (0.01s)
    --- PASS: TestOAuthAuditAdapterEndToEnd401AndRedirect/307 (0.01s)
=== RUN   TestOAuthAuditCancellationInFlight
--- PASS: TestOAuthAuditCancellationInFlight (0.01s)
=== RUN   TestOAuthAuditApplyDestinationAndHeaderOverwrite
--- PASS: TestOAuthAuditApplyDestinationAndHeaderOverwrite (0.00s)
=== RUN   TestOAuthRequestIsolationAnd401
--- PASS: TestOAuthRequestIsolationAnd401 (0.00s)
=== RUN   TestOAuthNative401AndHeaders
--- PASS: TestOAuthNative401AndHeaders (0.01s)
=== RUN   TestOAuthDefaultCatalogAfterMeasuredM0
--- PASS: TestOAuthDefaultCatalogAfterMeasuredM0 (0.00s)
=== RUN   TestSessionAuthPrecedesBodyAndPath
--- PASS: TestSessionAuthPrecedesBodyAndPath (0.00s)
=== RUN   TestIngressGenerationAndProviderGuard
--- PASS: TestIngressGenerationAndProviderGuard (0.00s)
PASS
coverage: 28.2% of statements
ok  	github.com/modu-ai/moai-adk/internal/gateway	1.627s	coverage: 28.2% of statements
```

### Craft — 함수 커버리지와 정적 검사

명령:

```sh
unset ANTHROPIC_BASE_URL ANTHROPIC_API_KEY ANTHROPIC_AUTH_TOKEN Z_AI_API_KEY && GOCACHE=/tmp/gateway-foundation-cache go tool cover -func=/var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/oauth-independent-audit-6yil1ch3/coverage.out | rg 'credential_anthropic_oauth.go|router.go|NewAnthropicOAuthAdapter|total:'
unset ANTHROPIC_BASE_URL ANTHROPIC_API_KEY ANTHROPIC_AUTH_TOKEN Z_AI_API_KEY && GOCACHE=/tmp/gateway-foundation-cache go vet ./internal/gateway/auth ./internal/gateway
```

관측 원문:

```text
github.com/modu-ai/moai-adk/internal/gateway/anthropic.go:44:				NewAnthropicOAuthAdapter	100.0%
github.com/modu-ai/moai-adk/internal/gateway/auth/credential_anthropic_oauth.go:12:	NewAnthropicOAuth		100.0%
github.com/modu-ai/moai-adk/internal/gateway/auth/credential_anthropic_oauth.go:56:	Provider			100.0%
github.com/modu-ai/moai-adk/internal/gateway/auth/credential_anthropic_oauth.go:60:	Generation			100.0%
github.com/modu-ai/moai-adk/internal/gateway/auth/credential_anthropic_oauth.go:66:	Redacted			100.0%
github.com/modu-ai/moai-adk/internal/gateway/auth/credential_anthropic_oauth.go:67:	String				100.0%
github.com/modu-ai/moai-adk/internal/gateway/auth/credential_anthropic_oauth.go:68:	GoString			100.0%
github.com/modu-ai/moai-adk/internal/gateway/auth/credential_anthropic_oauth.go:69:	Apply				100.0%
github.com/modu-ai/moai-adk/internal/gateway/router.go:30:				credential			84.2%
total:											(statements)			19.0%
go vet exit=0
```

go vet 자체 출력은 비어 있었으며 종료 코드를 shell에서 별도로 표시했다.

### Security·Consistency·baseline — 기계 검사

명령: `python3 /var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/oauth-independent-audit-6yil1ch3/mechanical.py`

검사 스크립트 원문:

```python
from pathlib import Path
import subprocess,json,hashlib,re,os
p=Path(__file__).parent
files=list(json.loads((p/'baseline.json').read_text()))
prod=[f for f in files if f.endswith('.go') and not f.endswith('_test.go')]
for label,pattern in [('command/injection',r'os/exec|exec\.Command|sql\.|unsafe|InsecureSkipVerify\s*:\s*true'),('credential storage/log sinks',r'\b(?:os\.Getenv|os\.ReadFile|os\.WriteFile|log\.|slog\.|fmt\.Print)')]:
 hits=[f'{f}:{i}' for f in prod for i,line in enumerate(Path(f).read_text().splitlines(),1) if re.search(pattern,line)]
 print(label+': '+(', '.join(hits) if hits else 'no matches in scoped production files'))
print('manifest diff lines:',len(subprocess.check_output(['git','diff','--','go.mod','go.sum'],text=True).splitlines()))
fmt=subprocess.run(['gofmt','-l']+[f for f in files if f.endswith('.go')],capture_output=True,text=True);print('gofmt exit='+str(fmt.returncode)+' output='+repr(fmt.stdout))
print('baseline hashes unchanged='+str(all(hashlib.sha256(Path(f).read_bytes()).hexdigest()==h for f,h in json.loads((p/'baseline.json').read_text()).items())))
covered=total=0
for line in (p/'coverage.out').read_text().splitlines()[1:]:
 loc,n,c=line.split();filename,region=loc.rsplit(':',1);start=int(region.split('.')[0]);end=int(region.split(',')[1].split('.')[0]);n=int(n);c=int(c)
 if filename.endswith('credential_anthropic_oauth.go') or filename.endswith('router.go') or filename.endswith('anthropic.go') and start>=44 and end<=46:
  total+=n;covered+=n*(c>0)
print(f'OAuth credential + router credential + OAuth constructor statements: {covered}/{total} = {100*covered/total:.2f}%')
for cmd in [['git','rev-parse','HEAD'],['git','branch','--show-current']]:print(subprocess.check_output(cmd,text=True).strip())
```

관측 원문:

```text
command/injection: no matches in scoped production files
credential storage/log sinks: no matches in scoped production files
manifest diff lines: 0
gofmt exit=0 output=''
baseline hashes unchanged=True
OAuth credential + router credential + OAuth constructor statements: 62/65 = 95.38%
81c1d58f9cf7045594ee61d5e4ff380948ce9eba
WT-unified-gateway
```

검색 무검출은 지정 파일의 해당 패턴에만 해당한다. 전체 OWASP 또는 의존성 취약점 데이터베이스 조회를 대신하지 않는다. 최초 탐색의 `log\.` 정규식은 `Catalog`의 일부를 잘못 잡았으며 단어 경계를 추가한 위 검사에서 다시 확인했다. 제품 보안 결함으로 계산하지 않았다.

### 반복 이력

1. 최초 overlay 시험에서 취소용 upstream fixture가 요청 본문을 소진하지 않은 채 context 종료를 기다렸다. client Send 취소 뒤에도 fixture 종료 대기가 끝나지 않아 전체 시험이 외부 Go timeout으로 종료되었다. 원문: `oauth_independent_audit_test.go:57: upstream not cancelled`, `panic: test timed out after 45s`, `FAIL`. 이 결과를 제품 결함이나 PASS 근거로 사용하지 않았다.
2. 감사 소유 fixture에서 본문 소진과 최대 대기 제한을 추가했다. 같은 제품 파일로 집중 시험이 통과했다.
3. 목적지 변형과 출력 헤더 대소문자 정리 시험을 추가한 뒤 위 최종 명령을 실행했다. 모든 선택 시험이 통과했다. 제품 구현은 수정하지 않았다.

## Baseline-attribution

- 실제 root: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/moai-proxy-unified`
- branch: `WT-unified-gateway`
- HEAD: `81c1d58f9cf7045594ee61d5e4ff380948ce9eba`
- 이번 감사는 이 HEAD 위의 **미커밋 구현 파일**을 측정했다. HEAD만으로 구현을 재현할 수 없다.
- 보고서 작성 전 `git fetch origin main` 출력: `From https://github.com/modu-ai/moai-adk` / `* branch main -> FETCH_HEAD`; `git rev-list --count --left-right origin/main...HEAD` 출력: `0	2879`.
- 구현·시험·manifest SHA-256은 아래와 같고 최종 재측정에서 모두 유지되었다.

```json
{
  "internal/gateway/auth/credential_anthropic_oauth.go": "333da1af5efc7f43284a79d7370bd15f4f67cf51496f13334a8ff159fdd0fbc0",
  "internal/gateway/auth/credential_anthropic_oauth_test.go": "ac093fe803a99ca79b1f74817361d69f921e7a99c4d0d8bb2a5cd5945aa741b8",
  "internal/gateway/router.go": "035732220f4eaf3039931d1c5f1a88cafad6f71b92fafdec0c969a841c0ab2df",
  "internal/gateway/server.go": "dadc61d7746557afbda8532e4aa76f8aaff05f484bc22c1467467a78a2fd097e",
  "internal/gateway/catalog.go": "de0174dba7fd55ce2331c4d95b73656343437170a1dd67f45b931f8aa44baa52",
  "internal/gateway/anthropic.go": "1b981ff3aeb8a14288778c43df75fb4dbf2ab1b3e974d6e1a294fbc157bdff73",
  "internal/gateway/oauth_test.go": "7a92f251707060d9c808ec1fceecdc737681c6f1f1bd1f48275711a1fa4cd860",
  "go.mod": "530783a72da61a17948ed525b77d8d62bf1b387c3cbe3ce2bef59aa549030eb0",
  "go.sum": "334c75e10a71d3931a8e084a3136907ecfb92099fe52acafd5883350bb5ed730"
}
```

독립 시험은 제품 폴더를 수정하지 않는 Go overlay로 추가했다. source: `/var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/oauth-independent-audit-6yil1ch3/oauth_audit_test.go`, SHA-256 `6dc911c62ee48eaa644ad2d0a78b6c3a0eb2606c8ae82eb3ad70fca29048bf52`. 재현 자료와 최초 실패 로그는 이 감사 소유 임시 디렉터리에 보존했다.

M0 선행 근거는 `m0-refresh-transport-observation.md`와 구현 보고서 `oauth-passthrough-verification.md`를 이번에 읽었다. 실제 token POST 200→반환 토큰의 후속 사용 관측은 그 선행 실행의 증거이며 이번 감사에서 재실행한 사실이 아니다. SPEC `REQ-MG-016` 및 design §2.1·§2.2의 측정 전 활성화 금지 경계에 한해 대조했다.

## Gaps

- 전체 SPEC AC, root factory, 설치 Claude의 native 정책·context·thinking·tool 조합, 실계정 갱신을 이번 구성품 시험으로 검증하지 않았다.
- 선택 시험의 패키지 전체 커버리지는 auth 6.5%, gateway 28.2%, 합계 19.0%다. 이는 좁은 시험 선택의 분모다. 전체 패키지 Craft 85% 충족을 주장하지 않는다.
- 고정된 TLS mock을 사용했다. 실제 Anthropic 서비스에 연결하거나 실제 자격 증명 저장소를 열지 않았다.
- 악성 caller가 custom Transport/DialTLSContext를 제공하는 상태, 전체 call-site의 raw-header logging, 의존성 전체 CVE, 다른 플랫폼은 감사 범위 밖이다.
- Go overlay source는 임시 디렉터리 자료다. 저장소에 영구 regression test로 편입하지 않았다.
- 커밋·push·PR·배포 및 queue 상태 변경을 수행하지 않았다.

## Residual-risk

허용된 endpoint·provider·요청 문맥 경계를 확인했어도 gateway 전체 제품 경로의 지원 상태가 되는 것은 아니다. Bearer는 요청 수명 동안 Go 메모리에 존재하며 취소가 바이트 zeroization을 보장하지 않는다. 이번 시험은 인터페이스 사용 가능 여부와 송신 경계를 검증한다. 실제 플랫폼의 다음 Claude Code 버전에서 OAuth 헤더·갱신 계약이 달라질 가능성은 별도 live 게이트에서 다루어야 한다.
