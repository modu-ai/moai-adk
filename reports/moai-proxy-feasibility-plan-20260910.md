# moai-proxy 도입 타당성 및 실행 계획

- 작성일: 2026-09-10 (Asia/Seoul)
- 보고 모드: plan
- 독자 수준: basic (`MoAI-Easy` 설정에서 도출)
- MoAI 기준선: `main`, `93299267a`
- ccmproxy 기준선: `112588175eb2b3b693a5bd23d58d8c501e0ae406`
- 판정: **조건부 GO — GPL 별도 배포물 또는 별도 허락을 전제로 진행**

## 1. 결론

기술적 가능성은 실측으로 확인됐다. Claude Code 2.1.267을 로컬 프록시에 연결해 `gpt-6-astra`가 실제로 응답했고, `/model` 선택기에 GPT-6가 표시됐으며, Read 도구 호출도 두 턴에 걸쳐 성공했다. 기존 `moai glm` 또한 `glm-5.3-flash[1m]` 실호출과 Read 도구 왕복에 성공했다.

그러나 `ccmproxy` 코드를 현재 Apache-2.0인 `moai-adk-go` 안에 직접 복사하는 방안은 승인할 수 없다. 상류는 GPL-3.0이다. 최신 main 소스에서 Go 1.26.1 표준 라이브러리의 호출 가능한 취약점 15건을 검출했고, 별도로 확인한 최신 릴리스 바이너리도 Go 1.26.1로 빌드됐다. 공식 Codex와 모델 메타데이터·OAuth refresh·세션 헤더 계약도 이미 어긋나 있다.

따라서 권고안은 다음과 같다.

1. `moai-proxy`를 GPL-3.0 독립 저장소·독립 실행 파일로 포크한다.
2. Apache-2.0 `moai`는 공개 CLI와 loopback HTTP 경계로 설치·실행·상태·정리만 담당한다.
3. 기본 연결 방식은 시스템 CA와 `/etc/hosts`를 건드리지 않는 `ANTHROPIC_BASE_URL=http://127.0.0.1:<port>`로 한다.
4. GLM은 현재 검증된 `moai glm` 직결 경로를 유지한다. 같은 Claude Code 세션에서 GPT↔GLM 전환이 제품 요구로 확정될 때만 `moai-proxy` provider로 추가한다.
5. Apache-only 단일 바이너리가 필수라면 코드 차용을 중단하고 상류 저작권자에게 별도 라이선스를 받거나 독립된 clean-room 구현으로 전환한다.

이 라이선스 판단은 기술적 배포 검토이며 법률 자문은 아니다.

## 2. 실측 기준선

| 대상 | 이번 실행에서 관측한 값 |
|---|---|
| 작업공간 | `/Users/goos/MoAI/moai-adk-go` |
| MoAI Git | `main` / `93299267a`; `origin/main...HEAD` = `1 1`; 기존 변경 약 353개 경로 |
| 설치 MoAI | `v3.2.0-rc.5`, commit `84fa4ece4` |
| Claude Code | `2.1.267` |
| Codex CLI | `0.153.4` |
| ccmproxy 최신 main | `112588175eb2b3b693a5bd23d58d8c501e0ae406`, 2026-09-08 |
| ccmproxy 릴리스 | `v0.1.0`, tag commit `2f7f68362198cd08d275c9b59ea85989c141b2bf` |
| 공식 Codex 비교 | `openai/codex` commit `5a9eb145c4c05fcfc7158d7c25b80e1322eccae1` |
| 실행 플랫폼 | macOS arm64 |
| 보고 세션 | `01a088fd-62fb-7cd1-b1ad-3dfabafc7474` |

공유 checkout은 원격과 `1 1`로 갈라져 있고 기존 변경이 많았다. 이번 작업은 새 보고서 두 파일만 추가했으며, 코드·브랜치·기존 파일은 변경하지 않았다.

## 3. 갭 실측 결과

| 항목 | 실측 | 상태 | 도입 조건 |
|---|---|---|---|
| Claude Code에서 GPT-6 응답 | `gpt-6-astra`가 `MOAI_PROXY_GPT6_OK`, exit 0 | 닫힘 | 현재 계약을 canary로 고정 |
| `/model` 선택기 | `GPT-6 Astra (Codex)` 행을 실제 TUI에서 확인 | 닫힘 | 안정된 `modelPicker.options` 사용 |
| GPT 도구 사용 | Read 도구로 파일 토큰 회수, 2턴, exit 0 | 닫힘 | 도구·결과 golden test 유지 |
| 시스템 CA 없는 연결 | 평문 loopback gateway로 `MOAI_HTTP_GATEWAY_OK` | 닫힘 | 이를 기본값으로 채택 |
| Claude 쪽 비필수 트래픽 | 기본 실행에서 eval/event logging을 관측; privacy env 적용 후 해당 호출 소멸 | 부분 닫힘 | `/api/hello` 로컬 응답과 egress allow-list 필요 |
| GLM 실사용 | `moai glm`이 `glm-5.3-flash[1m]` 응답, exit 0 | 닫힘 | 기존 직결 경로 유지 |
| GLM 도구 사용 | Read 도구 왕복 성공, 2턴 | 닫힘 | 회귀 E2E로 유지 |
| 단위·race·vet | `go test`, `go test -race`, `go vet` 모두 exit 0 | 닫힘 | PR CI로 승격 |
| coverage | Go toolchain 1.26.1/1.26.0 불일치로 전체 명령 실패 | 열림 | 깨끗한 Go 1.26.6+ 환경에서 재측정 |
| 취약점 | `govulncheck`가 호출 가능한 표준 라이브러리 취약점 15건 검출 | 차단 | Go 1.26.6+로 올리고 0건 확인 |
| 릴리스 재현성 | darwin/arm64 v0.1.0을 같은 명령으로 다시 빌드해 binary SHA256 일치 | 닫힘 | provenance까지 추가 |
| 릴리스 무결성 | archive checksum 일치; attestation 404, tag/commit unsigned | 부분 | 서명·attestation·SBOM 필수 |
| 플랫폼 | 6개 target cross-compile 성공; 실제 실행은 macOS arm64와 상류 Linux amd64 smoke뿐 | 열림 | 6개 target 실제 runtime smoke |
| Codex 최신 계약 | 모델·effort·OAuth refresh·session header 드리프트 확인 | 차단 | compatibility profile 재작성 |
| credential 동시 갱신 | 프로세스 간 lock/CAS/atomic rename 부재 확인 | 차단 | 전용 저장소 또는 lock+reread+atomic replace |
| GPL 경계 | ccmproxy GPL-3.0, MoAI Apache-2.0 | 차단 | 별도 GPL 실행물·법률 검토 또는 별도 허락 |
| 실제 장문맥 | `[1m]`에서 Claude client는 1,000,000 context로 인식 | 부분 | 272k 초과 실제 backend 호출은 별도 승인형 시험 |
| 실제 비용·할당량 | Claude client 추정값만 관측 | 열림 | provider dashboard readback 필요 |
| OAuth refresh 회전 | 유효 토큰을 강제로 만료시키지 않음 | 열림 | 격리 계정 canary에서 검증 |
| 혼합 세션 GPT↔GLM↔Claude | 구현 전이므로 미관측 | 열림 | provider 도입 시 상태·signature 격리 시험 |

### 최신 Codex와의 계약 차이

아래 값은 ccmproxy `1125881…`와 공식 Codex `5a9eb14…` 소스를 같은 날 대조한 결과다. “공식 현재값”은 공개 OpenAI API 전체의 영구 계약이 아니라 해당 Codex commit의 구현 기준선이다.

| 계약 항목 | ccmproxy 현재값 | 공식 Codex 현재값 | moai-proxy 목표 | 불일치 시 동작 |
|---|---|---|---|---|
| Astra/Sol 기본 effort | `medium` | `low` | pinned compatibility profile의 값 | freshness CI 실패, 릴리스 차단 |
| Astra/Sol/Terra effort | `ultra` 미지원 | `ultra` 지원 | 모델별 허용 집합을 그대로 보존 | 요청 거절; 조용한 `max` 강등 금지 |
| Sol/Terra/Luna context | `372000` | 표준 `272000`, experimental max `872000` | standard/max를 별도 필드로 보존 | catalog fixture 실패 |
| Spark 노출 | `gpt-5.3-codex-spark` 노출 | visible catalog에 없음 | entitlement 조회 결과에만 노출 | picker에서 숨김 |
| authorize scope | `openid email profile offline_access`; `originator` query 없음 | 여기에 connector read/invoke scope와 `originator` 포함 | 공식 client profile과 fixture 일치 | OAuth 시작 전 진단 실패 |
| refresh body | form-urlencoded, refresh scope 포함 | JSON, scope 없음 | 공식 client profile과 fixture 일치 | token write 금지, 재로그인 안내 |
| session header | `Session_id` | `session-id`, `thread-id` | 매 세션 UUID를 두 hyphen header로 전달 | contract test 실패 |
| originator 기본값 | `codex-tui` | `codex_cli_rs` | profile 값으로 버전 관리 | canary 경고 뒤 릴리스 차단 |

근거 위치:

- ccmproxy 모델: `internal/codex/models.go:31-106`
- ccmproxy OAuth: `internal/codex/oauth.go:22-88,136-220`
- ccmproxy 요청 헤더: `internal/codex/responses.go:15-108`
- 공식 Codex 모델: `codex-rs/models-manager/models.json`
- 공식 Codex OAuth: `codex-rs/login/src/server.rs:576-612`, `auth/manager.rs:1597-1617`
- 공식 Codex 헤더: `codex-rs/codex-api/src/requests/headers.rs:5-13`

## 4. 실제 동작 증거

### GPT-6 기본 응답

```text
Claude Code 2.1.267
model: gpt-6-astra
result: MOAI_PROXY_GPT6_OK
exit: 0
proxy route: POST https://chatgpt.com/backend-api/codex/responses -> 200
```

### GPT-6 모델 선택기

```text
Header: GPT-6 Astra (Codex) with low effort
/model row: 5. GPT-6 Astra (Codex) ✔ OpenAI GPT-6 Astra through a local Codex proxy.
```

### GPT-6 도구 왕복

```text
fixture: MOAI_PROXY_TOOL_ROUNDTRIP_OK
allowed tool: Read
result: exact fixture token returned
turns: 2
permission denials: 0
exit: 0
```

### GLM 실호출과 도구 왕복

```text
moai glm model: glm-5.3-flash[1m]
result: MOAI_GLM_LIVE_OK
exit: 0

Read fixture: MOAI_PROXY_TOOL_ROUNDTRIP_OK
turns: 2
permission denials: 0
exit: 0
```

### 상류 품질 검사

```text
go test ./... -count=1          -> 10 packages ok, 1 package no test files
go test -race ./... -count=1    -> all packages pass
go vet ./...                    -> exit 0
gofmt -l .                      -> no output, exit 0
govulncheck ./...               -> 15 reachable standard-library vulnerabilities
go test ./... -count=1 -cover   -> failed: compiler/tool version mismatch
```

### 릴리스 검사

```text
archive: ccmproxy_v0.1.0_darwin_arm64.tar.gz
SHA256: 345e55a4bb2df6d512f2f2a432ee8c00515fbbfca75eb6e89dac899858a915c7
checksum match: yes
binary rebuild SHA256 match: yes
GitHub attestation: not found (HTTP 404)
binary signature: ad-hoc; TeamIdentifier not set
```

### 재현용 증거 원장

민감한 token 값과 임시 디렉터리 절대경로만 `<REDACTED>`로 바꿨다. 결과 토큰, 모델, status, exit code는 관측값을 그대로 적었다.

```text
Command:
git -C <TEMP>/repo rev-parse HEAD
Output:
112588175eb2b3b693a5bd23d58d8c501e0ae406

Command:
git -C <TEMP>/repo show -s --format='%H%n%cI%n%s' HEAD
Output:
112588175eb2b3b693a5bd23d58d8c501e0ae406
2026-09-08T10:16:22+09:00
docs: lead with the /etc/hosts setup and show the model picker
```

```text
Command:
go test ./... -count=1
Output:
ok github.com/jclab-joseph/claude-code-model-proxy/cmd/ccmproxy
ok github.com/jclab-joseph/claude-code-model-proxy/internal/ca
ok github.com/jclab-joseph/claude-code-model-proxy/internal/codex
ok github.com/jclab-joseph/claude-code-model-proxy/internal/config
ok github.com/jclab-joseph/claude-code-model-proxy/internal/httplog
?  github.com/jclab-joseph/claude-code-model-proxy/internal/logx [no test files]
ok github.com/jclab-joseph/claude-code-model-proxy/internal/proxy
ok github.com/jclab-joseph/claude-code-model-proxy/internal/resolve
ok github.com/jclab-joseph/claude-code-model-proxy/internal/router
ok github.com/jclab-joseph/claude-code-model-proxy/internal/tokenize
ok github.com/jclab-joseph/claude-code-model-proxy/internal/translate
Exit code: 0

Command:
govulncheck ./...
Output excerpt:
Your code is affected by 15 vulnerabilities from the Go standard library.
This scan also found 5 vulnerabilities in packages you import and 8
vulnerabilities in modules you require, but your code doesn't appear to call
these vulnerabilities.
Exit code: 3
Target: ccmproxy main 112588175eb2b3b693a5bd23d58d8c501e0ae406
```

```text
Command:
go test -race ./... -count=1
Output:
ok github.com/jclab-joseph/claude-code-model-proxy/cmd/ccmproxy 1.870s
ok github.com/jclab-joseph/claude-code-model-proxy/internal/ca 3.982s
ok github.com/jclab-joseph/claude-code-model-proxy/internal/codex 5.997s
ok github.com/jclab-joseph/claude-code-model-proxy/internal/config 6.295s
ok github.com/jclab-joseph/claude-code-model-proxy/internal/httplog 6.989s
?  github.com/jclab-joseph/claude-code-model-proxy/internal/logx [no test files]
ok github.com/jclab-joseph/claude-code-model-proxy/internal/proxy 6.659s
ok github.com/jclab-joseph/claude-code-model-proxy/internal/resolve 8.649s
ok github.com/jclab-joseph/claude-code-model-proxy/internal/router 12.337s
ok github.com/jclab-joseph/claude-code-model-proxy/internal/tokenize 8.462s
ok github.com/jclab-joseph/claude-code-model-proxy/internal/translate 8.104s
Exit code: 0

Command: go vet ./...
Output: <empty>
Exit code: 0

Command: gofmt -l .
Output: <empty>
Exit code: 0

Command: go test ./... -count=1 -cover
Output excerpt:
compile: version "go1.26.1" does not match go tool version "go1.26.0"
ok github.com/jclab-joseph/claude-code-model-proxy/internal/ca 1.326s coverage: 76.8% of statements
ok github.com/jclab-joseph/claude-code-model-proxy/internal/codex 1.723s coverage: 67.7% of statements
ok github.com/jclab-joseph/claude-code-model-proxy/internal/router 3.467s coverage: 84.3% of statements
Exit code: 1
Target: ccmproxy main 112588175eb2b3b693a5bd23d58d8c501e0ae406
```

```text
Command shape (temporary paths and dummy credential redacted):
env HTTPS_PROXY=http://127.0.0.1:18080 HTTP_PROXY=http://127.0.0.1:18080 \
  NODE_EXTRA_CA_CERTS=<REDACTED>/ca.crt ANTHROPIC_API_KEY=<DUMMY> \
  claude --model gpt-6-astra --permission-mode default --safe-mode \
  --restricted --strict-mcp-config --tools '' --effort low \
  --no-session-persistence --output-format json --print '<FIXED_TOKEN_PROMPT>'
Observed JSON fields:
result: MOAI_PROXY_GPT6_OK
modelUsage.gpt-6-astra.inputTokens: 1006
modelUsage.gpt-6-astra.outputTokens: 11
modelUsage.gpt-6-astra.contextWindow: 200000
modelUsage.gpt-6-astra.maxOutputTokens: 32000
Exit code: 0
Proxy log excerpt:
codex messages model=gpt-6-astra -> gpt-6-astra effort=low
POST https://chatgpt.com/backend-api/codex/responses -> 200
```

```text
Command shape (temporary settings and fixture path redacted):
env ANTHROPIC_BASE_URL=http://127.0.0.1:18081 ANTHROPIC_API_KEY=<DUMMY> \
  CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC=1 DISABLE_TELEMETRY=1 \
  DISABLE_ERROR_REPORTING=1 claude --settings <REDACTED>/picker-settings.json \
  --model 'gpt-6-astra[1m]' --permission-mode default --safe-mode --restricted \
  --strict-mcp-config --tools Read --effort low --no-session-persistence \
  --output-format json --print '<READ_FIXTURE_PROMPT>'
Observed fields:
result: MOAI_PROXY_TOOL_ROUNDTRIP_OK
num_turns: 2
permission_denials: []
modelUsage.gpt-6-astra.contextWindow: 1000000
modelUsage.gpt-6-astra.maxOutputTokens: 64000
Exit code: 0
```

```text
Command:
gtimeout 90 moai glm --permission-mode default --safe-mode --restricted \
  --strict-mcp-config --tools '' --effort low --no-session-persistence \
  --output-format json --print '<FIXED_TOKEN_PROMPT>'
Observed fields:
result: MOAI_GLM_LIVE_OK
modelUsage.glm-5.3-flash[1m].inputTokens: 1900
modelUsage.glm-5.3-flash[1m].outputTokens: 13
modelUsage.glm-5.3-flash[1m].contextWindow: 1000000
modelUsage.glm-5.3-flash[1m].maxOutputTokens: 32000
Exit code: 0
```

```text
Command shape (temporary fixture path redacted):
gtimeout 90 moai glm --permission-mode default --safe-mode --restricted \
  --strict-mcp-config --tools Read --add-dir <REDACTED> --effort low \
  --no-session-persistence --output-format json --print '<READ_FIXTURE_PROMPT>'
Observed fields:
result: MOAI_PROXY_TOOL_ROUNDTRIP_OK
num_turns: 2
permission_denials: []
modelUsage.glm-5.3-flash[1m].inputTokens: 2641
modelUsage.glm-5.3-flash[1m].cacheReadInputTokens: 2496
modelUsage.glm-5.3-flash[1m].outputTokens: 66
Exit code: 0
```

```text
Interactive picker procedure:
1. Start HTTP-only ccmproxy on 127.0.0.1:18081.
2. Start Claude Code 2.1.267 with isolated CLAUDE_CONFIG_DIR,
   picker-settings.json, ANTHROPIC_BASE_URL=http://127.0.0.1:18081,
   privacy environment variables, --model gpt-6-astra, --safe-mode.
3. Enter /model.
Observed terminal rows:
GPT-6 Astra (Codex) with low effort
6. GPT-6 Astra (Codex) ✔ OpenAI GPT-6 Astra through a local Codex proxy.
Observation termination: Esc, then Ctrl-C twice; interactive exit was not used as a pass criterion.
Proxy cleanup: Ctrl-C; port 18081 no longer LISTEN.
```

```text
Command shape — HTTP-only gateway:
env ANTHROPIC_BASE_URL=http://127.0.0.1:18081 ANTHROPIC_API_KEY=<DUMMY> \
  claude --model gpt-6-astra --permission-mode default --safe-mode \
  --restricted --strict-mcp-config --tools '' --effort low \
  --no-session-persistence --output-format json --print '<FIXED_TOKEN_PROMPT>'
Observed output:
result: MOAI_HTTP_GATEWAY_OK
Exit code: 0
Proxy log excerpt:
HEAD /api/hello -> Anthropic 200
POST https://chatgpt.com/backend-api/codex/responses -> 200

Repeat with --model 'gpt-6-astra[1m]':
result: MOAI_HTTP_1M_OK
contextWindow: 1000000
maxOutputTokens: 64000
Exit code: 0
```

```text
Privacy comparison, same minimal GPT request:
Baseline without privacy variables, observed proxy log paths:
/api/eval/... -> 200
/api/event_logging/v2/batch -> 200 (request sizes observed: 174–213 KB)

With:
CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC=1
DISABLE_TELEMETRY=1
DISABLE_ERROR_REPORTING=1
Observed output:
result: MOAI_PRIVACY_ENV_OK
Exit code: 0
Observed proxy log paths in that run:
GET policy_limits/settings -> 401
POST https://chatgpt.com/backend-api/codex/responses -> 200
No /api/eval or /api/event_logging path appeared in that run.
Evidence boundary: original temporary proxy logs were not retained after the run;
these are the command transcript excerpts observed in this session.
```

```text
Command:
for target in darwin/amd64 darwin/arm64 linux/amd64 linux/arm64 windows/amd64 windows/arm64; do
  CGO_ENABLED=0 GOOS=<os> GOARCH=<arch> go build -trimpath \
    -ldflags '-s -w -X main.version=audit' ./cmd/ccmproxy
done
Output:
darwin/amd64 PASS
darwin/arm64 PASS
linux/amd64 PASS
linux/arm64 PASS
windows/amd64 PASS
windows/arm64 PASS
Exit code: 0
Target: ccmproxy main 112588175eb2b3b693a5bd23d58d8c501e0ae406
```

```text
Command: shasum -a 256 ccmproxy.tar.gz ccmproxy ccmproxy-rebuilt
Output:
345e55a4bb2df6d512f2f2a432ee8c00515fbbfca75eb6e89dac899858a915c7  ccmproxy.tar.gz
b7ca6bd502638f602e22555d7b990bc626e435bf9b09777d31a9ef443a1cc402  ccmproxy
b7ca6bd502638f602e22555d7b990bc626e435bf9b09777d31a9ef443a1cc402  ccmproxy-rebuilt
Exit code: 0

Command:
gh release download v0.1.0 --repo jclab-joseph/claude-code-model-proxy \
  --pattern checksums.txt --output /tmp/ccmproxy-checksums-audit.txt
rg 'ccmproxy_v0.1.0_darwin_arm64.tar.gz' /tmp/ccmproxy-checksums-audit.txt
Output:
345e55a4bb2df6d512f2f2a432ee8c00515fbbfca75eb6e89dac899858a915c7  ccmproxy_v0.1.0_darwin_arm64.tar.gz
Exit code: 0

Command: go version -m ccmproxy
Output excerpt:
ccmproxy: go1.26.1
path github.com/jclab-joseph/claude-code-model-proxy/cmd/ccmproxy
mod github.com/jclab-joseph/claude-code-model-proxy v0.1.0
build vcs.revision=2f7f68362198cd08d275c9b59ea85989c141b2bf
Exit code: 0

Command: gh attestation verify ccmproxy.tar.gz --repo jclab-joseph/claude-code-model-proxy
Output:
Error: HTTP 404: Not Found (.../attestations/sha256:345e55a4...)
Exit code: 1

Command: git tag -v v0.1.0
Output excerpt:
object 2f7f68362198cd08d275c9b59ea85989c141b2bf
type commit
tag v0.1.0
error: no signature found
Exit code: 1

Command: codesign -dv --verbose=4 ccmproxy
Output excerpt:
Format=Mach-O thin (arm64)
Signature=adhoc
TeamIdentifier=not set
Exit code: 0

Command: gh api repos/jclab-joseph/claude-code-model-proxy/commits/1125881... --jq '.commit.verification'
Output:
{"reason":"unsigned","signature":null,"verified":false}
Exit code: 0
```

상류 소스 결론의 직접 근거:

| 주장 | SHA 고정 근거 |
|---|---|
| `eval` 가능한 미인용 env 출력 | [`cmd/ccmproxy/main.go:680-706`](https://github.com/jclab-joseph/claude-code-model-proxy/blob/112588175eb2b3b693a5bd23d58d8c501e0ae406/cmd/ccmproxy/main.go#L680-L706) |
| 예측 가능한 `/tmp/ccmproxy.exe`, `setcap`, hosts 수정 | [`run.sh:1-42`](https://github.com/jclab-joseph/claude-code-model-proxy/blob/112588175eb2b3b693a5bd23d58d8c501e0ae406/run.sh#L1-L42) |
| credential 직접 쓰기 | [`internal/codex/store.go:152-292`](https://github.com/jclab-joseph/claude-code-model-proxy/blob/112588175eb2b3b693a5bd23d58d8c501e0ae406/internal/codex/store.go#L152-L292) |
| 고정 User-Agent·originator | [`internal/config/config.go:180-206`](https://github.com/jclab-joseph/claude-code-model-proxy/blob/112588175eb2b3b693a5bd23d58d8c501e0ae406/internal/config/config.go#L180-L206) |
| release Action major tag와 unsigned checksum | [`.github/workflows/release.yml:1-156`](https://github.com/jclab-joseph/claude-code-model-proxy/blob/112588175eb2b3b693a5bd23d58d8c501e0ae406/.github/workflows/release.yml#L1-L156) 및 GitHub attestation 404 실측 |

## 5. 권고 아키텍처

```text
Claude Code
  │ ANTHROPIC_BASE_URL=http://127.0.0.1:<ephemeral-port>
  ▼
moai-proxy (GPL-3.0, 별도 실행 파일)
  ├─ readiness/health + fail-closed router
  ├─ Anthropic Messages ↔ Codex Responses adapter
  ├─ Codex credential store 또는 읽기 전용 Codex auth adapter
  └─ provider registry
       ├─ GPT/Codex → ChatGPT Codex backend
       └─ GLM/Z.AI → 선택 기능; 기본은 기존 `moai glm` 직결

moai (Apache-2.0)
  ├─ `moai gpt setup`: 명시적 설치, digest 검증
  ├─ `moai gpt`: 자식 프로세스 감독, 환경 주입, 종료 정리
  ├─ `moai gpt status`: 마스킹된 상태
  └─ `moai doctor`: 선택 기능 진단
```

핵심 경계는 “같은 Go module에 GPL 코드를 포함하지 않고, 별도 실행 파일을 loopback HTTP와 공개 CLI로 관리”하는 것이다. 그래도 묶음 배포와 전용 통신의 결합성은 법률 검토 대상이다.

## 6. 구현 계획

### Phase 0 — 라이선스와 배포 경계

- `moai-proxy`를 GPL-3.0 독립 저장소·독립 릴리스로 확정한다.
- 원 저작권·GPL 사본·수정일·corresponding source·의존성 NOTICE 제공 방식을 문서화한다.
- `moai`와 `moai-proxy`의 패키징·업데이트·제거 경계를 법률 검토한다.
- Apache-only 요구가 나오면 직접 차용을 중단하고 별도 허락 또는 clean-room으로 전환한다.

완료 조건: 라이선스 선택, 소스 제공 URL, NOTICE, 배포 단위가 승인 기록에 남아 있다.

### Phase 1 — 상류 포크 하드닝

- Go 기준을 1.26.6 이상으로 올리고 `govulncheck` 0건을 강제한다.
- PR/main CI를 추가하고 Actions를 immutable SHA로 고정한다.
- signed tag, SBOM, provenance/attestation, checksum 서명을 추가한다.
- archive에 LICENSE와 NOTICE를 넣는다.
- `eval` 가능한 `ccmproxy env`를 제거한다.
- HTTP-only 모드에서는 CA를 만들지 않도록 lazy initialization한다.

완료 조건: `gofmt`, `vet`, unit, race, vulnerability, archive policy가 CI에서 모두 통과한다.

### Phase 2 — 호환성·보안 계약 재기준화

- 공식 Codex의 모델 metadata, effort, `session-id`, `thread-id`, originator를 기준으로 compatibility profile을 만든다.
- OAuth authorize/refresh를 현재 공식 구현과 fixture 수준으로 맞춘다.
- unknown model·unknown compression·의도하지 않은 provider 전달은 fail closed한다.
- credential은 proxy 전용 파일을 기본으로 하고, 0600·lock·reread·merge·fsync·atomic rename을 적용한다.
- 로그에서 credential, 이메일, prompt, tool payload, query secret을 차단한다.
- `127.0.0.1:0`을 지원하고 실제 포트와 readiness를 JSON 또는 FD로 부모에 전달한다.
- `/api/hello`를 로컬에서 처리하고 외부 egress allow-list를 문서화한다.

완료 조건: 실계정 없이 protocol golden, negative routing, concurrent refresh, cleanup 시험이 통과한다.

### Phase 3 — MoAI 명령 통합

- 사용자 여정은 **명시적 2단계**로 고정한다: 최초 1회 `moai gpt setup`, 매 사용 시 `moai gpt`.
- `moai gpt setup`: GPL 고지와 설치 대상 버전을 보여 주고 사용자의 대화형 동의를 받은 뒤 플랫폼 자산을 선택한다. 내장 digest와 대조하고 누락·불일치 시 실패한다.
- 설치 위치는 POSIX `~/.local/share/moai/bin/moai-proxy`, Windows `%LOCALAPPDATA%\MoAI\bin\moai-proxy.exe`로 고정한다.
- Codex auth가 없으면 setup 안에서 `moai-proxy login`을 실행한다. 사용자가 취소하면 내려받은 임시 archive와 부분 설치를 지우고 기존 버전은 보존한다.
- update는 새 바이너리를 sibling temp에 검증한 뒤 atomic rename하고, readiness 실패 시 직전 버전으로 되돌린다.
- `moai gpt`: proxy를 자식으로 시작하고 ready를 기다린 뒤 Claude Code를 자식으로 실행한다.
- POSIX의 기존 `syscall.Exec` 경로를 쓰지 않고, 신호 전달·정확한 PID 종료·Claude exit code 보존을 구현한다.
- 임시 Claude settings에 `modelPicker.options`를 넣고 `gpt-6-astra[1m]`을 선택 가능하게 한다.
- `CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC=1`, `DISABLE_TELEMETRY=1`, `DISABLE_ERROR_REPORTING=1`을 기본 주입한다.
- `moai gpt status`, `moai doctor`에는 바이너리 버전·digest·auth source·readiness를 마스킹해 표시한다.
- 초기판에서는 웹 설정 UI를 추가하지 않는다.

주요 수정 후보:

| 영역 | 후보 파일 |
|---|---|
| 명령 등록 | `internal/cli/root.go`, 신규 `internal/cli/gpt.go` |
| 자식 감독 | `internal/cli/launcher.go`, `internal/cli/codex_job_control.go` 패턴 재사용 |
| 설정 타입·기본값 | `internal/config/types.go`, `defaults.go`, `validation.go` |
| 배포 템플릿 | `internal/template/templates/.moai/config/sections/llm.yaml` |
| 설치·검증 | `internal/update`의 checksum/rollback 구조를 일반화하되 fail-closed 유지 |
| 진단 | `internal/cli/doctor.go`, `doctor_codex.go` 패턴 재사용 |

완료 조건: 설치 실패가 기존 설치를 훼손하지 않고, 정상·오류·Ctrl-C 종료 뒤 child/socket/temp settings가 남지 않는다.

### Phase 4 — GPT-6 E2E와 릴리스

- Claude Code 최신 지원 버전에서 단문, streaming, Read 도구, tool result, thinking continuation, picker를 시험한다.
- 272k 초과 장문맥은 별도 승인형·비용 상한 canary로 검증한다.
- provider dashboard에서 모델·계정·사용량을 읽어 client 추정치와 분리 기록한다.
- macOS arm64/amd64, Linux arm64/amd64, Windows arm64/amd64 실제 runtime smoke를 수행한다.
- 깨끗한 archive에서 설치→실행→업데이트 실패 rollback→제거까지 검증한다.

완료 조건: 모든 플랫폼 결과와 실계정 canary 증거가 릴리스 SHA에 귀속된다.

### Phase 5 — 선택적 GLM 통합

- 기본 제품은 이미 검증된 `moai glm` 직결 경로를 유지한다.
- 같은 세션의 `/model`에서 GPT↔GLM을 바꿔야 한다는 요구가 승인된 경우에만 provider interface를 연다.
- Z.AI의 Anthropic-compatible 요청을 불필요하게 Codex 형식으로 변환하지 않고 pass-through한다.
- GPT·GLM credential과 reasoning signature가 상대 provider로 절대 넘어가지 않는 negative test를 둔다.

완료 조건: 기존 `moai glm`보다 사용자 가치가 명확하고, provider 간 비밀·상태 격리 시험이 모두 통과한다.

## 7. 합격 기준

| ID | 우선순위 | 실행 가능한 합격 기준 |
|---|---|---|
| LIC-01 | P0 | 지정된 법률 검토자가 GPL 독립 배포·source URL·LICENSE·NOTICE·수정일 checklist를 서명하고, release archive 검사에서 네 항목이 모두 존재함 |
| SEC-01 | P0 | `claude-*`, `gpt-*`, `glm-*`, unknown model × Anthropic/Codex/Z.AI credential fixture 12개에서 의도한 provider 이외 Authorization header가 0건임 |
| SUP-01 | P0 | checksum·signature·release metadata 각각 missing/mismatch fixture 6개가 모두 non-zero exit이며 기존 설치 SHA가 유지됨 |
| VUL-01 | P0 | pinned Go ≥1.26.6에서 `govulncheck ./...` 출력이 `No vulnerabilities found.`이고 exit 0임 |
| INS-01 | P1 | 깨끗한 임시 HOME에서 `moai gpt setup`→`moai gpt`를 실행해 설치·login/readiness·고정 토큰 응답을 완주함; 취소 fixture에서는 설치 전후 파일 SHA가 같음 |
| SYS-01 | P1 | 실행 전후 `/etc/hosts` SHA, system trust 목록, privileged listener 목록이 동일함 |
| E2E-01 | P1 | 지원 행렬에 고정한 Claude Code 2.1.267에서 picker·단문·SSE stream·Read tool fixture가 모두 exit 0임 |
| CLN-01 | P1 | 정상·Ctrl-C·proxy crash·Claude 오류 4개 fixture 후 기록된 child PID가 없고 ready JSON의 port가 LISTEN하지 않으며 temp settings 경로가 없음 |
| DRF-01 | P1 | pinned official Codex SHA의 model JSON, OAuth request, header golden fixture가 byte/semantic 비교를 통과함; 차이 fixture는 CI non-zero임 |
| OS-01 | P1 | darwin/linux/windows × amd64/arm64 6개 runner에서 설치→version→loopback ready→shutdown smoke가 같은 release SHA로 통과함 |
| CTX-01 | P2 | 별도 승인과 비용 상한 아래 272k 초과 synthetic payload가 성공하거나 관측된 상한·오류를 기록하고 provider dashboard 사용량을 run ID에 연결함 |
| GLM-01 | P2 | 제품 책임자가 same-session 전환 요구를 승인한 뒤, GPT↔GLM tool/reasoning 전환과 SEC-01 격리 시험이 모두 통과해야 provider flag를 켬 |

## 8. 권고하지 않는 방안

- GPL 소스를 `internal/`로 복사한 뒤 Apache-2.0만 유지하는 방식
- `eval "$(moai-proxy env)"` 같은 shell 문자열 실행
- 예측 가능한 `/tmp/ccmproxy.exe`에 빌드한 뒤 privileged `setcap`을 부여하는 방식
- `/etc/hosts`와 system CA 설치를 기본값으로 삼는 방식
- `/etc/hosts` 전체를 임시 파일로 다시 쓰는 방식
- checksum을 받지 못했을 때 설치를 계속하는 방식
- 서명되지 않은 checksum과 mutable Action tag만 신뢰하는 방식
- 고정 포트를 먼저 고른 뒤 proxy를 띄우는 방식
- `pkill`이나 프로세스 이름으로 정리하는 방식
- unknown model을 조용히 Anthropic으로 넘기는 방식
- 유효한 Codex auth 파일을 직접 덮어쓰면서 프로세스 간 잠금을 생략하는 방식
- 모델 catalog와 Codex User-Agent를 코드에 장기 고정하는 방식
- GLM을 Codex translator에 억지로 통과시키는 방식

## 9. 미관측 범위와 잔여 위험

### Gaps

- 272k를 넘는 GPT-6 실제 backend 장문맥 경계는 비용·데이터 전송을 수반해 실행하지 않았다.
- provider dashboard의 실제 과금·할당량은 읽지 않았다. 보고된 `costUSD`는 Claude Code의 추정치다.
- OAuth refresh token 회전을 강제로 유발하지 않았다.
- 이번 머신에서는 macOS arm64만 실행했다. 상류 CI의 Linux amd64 smoke 성공은 별도로 확인했지만, 나머지 target의 실제 실행은 검증하지 않았다.
- MCP, subagent, image, PDF, structured output, remote control의 실계정 E2E는 수행하지 않았다.
- GPT↔GLM↔Claude 동일 세션 전환은 구현 전이라 검증할 수 없다.
- GPL 별도 실행 파일 경계의 최종 법적 효력과 서비스 약관 허용 여부는 확정하지 않았다.

### Residual risk

- ChatGPT subscription backend와 Claude Code picker는 공개된 안정 계약이 아니므로 양쪽 업데이트로 깨질 수 있다.
- Anthropic은 공식 문서에서 비-Claude 모델을 Claude Code에 라우팅하는 구성을 지원하지 않는다고 밝힌다.
- 별도 실행 파일이라도 결합 배포·전용 프로토콜의 밀접성에 따라 GPL 판단이 달라질 수 있다.
- Messages API와 Responses API의 의미론 차이 때문에 새 content block이나 tool 유형에서 정보 손실이 생길 수 있다.
- 실계정 canary는 credential·비용·조직 정책 위험을 가지므로 항상 opt-in과 비용 상한이 필요하다.

## 10. 출처

- ccmproxy 저장소: https://github.com/jclab-joseph/claude-code-model-proxy
- ccmproxy 기준 commit: https://github.com/jclab-joseph/claude-code-model-proxy/tree/112588175eb2b3b693a5bd23d58d8c501e0ae406
- 공식 Codex 저장소: https://github.com/openai/codex/tree/5a9eb145c4c05fcfc7158d7c25b80e1322eccae1
- OpenAI GPT-6 Astra 모델: https://developers.openai.com/api/docs/models/gpt-6-astra
- Anthropic 네트워크 설정: https://code.claude.com/docs/en/network-config
- Anthropic LLM gateway: https://code.claude.com/docs/en/llm-gateway
- Z.AI Claude Code 설정: https://docs.z.ai/devpack/tool/claude
- GNU GPL FAQ — aggregate: https://www.gnu.org/licenses/gpl-faq.html#MereAggregation

## 11. 최종 판정

`moai-proxy`는 만들 수 있고, Claude Code에서 GPT-6를 사용하는 핵심 경로는 이 머신에서 실제로 동작했다. GLM도 이미 `moai glm`으로 동작한다. 다만 제품화의 첫 작업은 translator를 복사하는 일이 아니라 **GPL 경계 확정, Go 1.26.6+ 보안 갱신, 최신 Codex 계약 재기준화, fail-closed 설치·supervision 설계**여야 한다.

권고 의사결정은 **조건부 GO**다. 독립 GPL companion 방식을 승인하면 Phase 0부터 진행할 수 있다. Apache-only 단일 배포가 필수라면 별도 라이선스 확보 전까지는 **NO-GO**다.
