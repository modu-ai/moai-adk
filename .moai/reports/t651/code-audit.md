# t651 도구 등록 패키지 독립 감사

SPEC: SPEC-MOAI-GATEWAY-001 0.11.0의 AS-004, AS-006 중 등록 패키지 책임
Overall Verdict: PASS (범위: internal/codextools)
Score: 100/100 (아래 네 차원의 제한된 패키지 검사 범위)
전체 SPEC 및 실제 Claude 왕복 판정: UNVERIFIED — 이번 감사의 판정 대상이 아니다.

## Claim

초기 도구의 native 등록, 후발 도구의 dispatcher 전용 연결, 현재 요청의 허용 목록 및 전체 schema 검사, 소유권·호출 재사용 방지에 대해 이번 범위에서 차단 결함을 발견하지 않았다. 실제 포착한 Claude 도구 정의도 초기 등록과 후발 발견에 사용할 수 있었다. 프로덕션 코드는 수정하지 않았다.

| Dimension | Score | Verdict | Evidence |
|---|---:|---|---|
| Functionality (40%) | 100 | PASS, 패키지 범위 | `ok  github.com/modu-ai/moai-adk/internal/codextools  1.391s`; `initial=12 followup=13 native=12 captured schema registry accepted` |
| Security (25%) | 100 | PASS, 등록·인수·소유권 경계 | `--- PASS: TestAuditDuplicateBeginConcurrent (0.00s)`; `--- PASS: TestAuditAtomicDiscoveryAndForeignBindings (0.00s)`; `--- PASS: TestAuditSchemas (0.00s)` |
| Craft (20%) | 100 | PASS | `ok  github.com/modu-ai/moai-adk/internal/codextools  0.352s  coverage: 87.5% of statements`; `go vet` 출력 없음 |
| Consistency (15%) | 100 | PASS | `gofmt -l internal/codextools` 출력 없음; `internal/codextools/json.go:1:package codextools` |

점수는 전체 제품 완성도나 알려지지 않은 결함의 부재를 뜻하지 않는다. 기본 평가 프로필의 coverage >=85%, Critical/High 결함 없음 기준을 적용했다. HTTP/PTY 기능은 이 모듈의 차단 사유로 확대하지 않았다.

## Evidence

기준 작업 디렉터리: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t651`

### 기능 시험

명령: `go test -race ./internal/codextools`

```text
ok  	github.com/modu-ai/moai-adk/internal/codextools	1.391s
```

명령: `go test -cover ./internal/codextools`

```text
ok  	github.com/modu-ai/moai-adk/internal/codextools	0.352s	coverage: 87.5% of statements
```

명령: `go vet ./internal/codextools`, `gofmt -l internal/codextools`

두 명령 모두 출력 없음. 함께 실행한 명령 호출의 종료 코드 0.

### 보안·의존성·명명 검사

명령: `rg -n 'exec\.|net/http|os\.|Load\(|UseLoader|owner !=|strictJSON|schema.Validate|jsonschema' internal/codextools go.mod`

관측 출력 중 관련 줄:

```text
go.mod:89: github.com/santhosh-tekuri/jsonschema/v6 v6.0.2 // indirect
internal/codextools/registry.go:91:func (denyLoader) Load(string) (any, error) { return nil, ErrInvalid }
internal/codextools/registry.go:124: compiler.UseLoader(denyLoader{})
internal/codextools/registry.go:228: if r.owner.ThreadID == "" || owner != r.owner || len(refs) == 0 || len(refs) > maxTools {
internal/codextools/registry.go:326: if e.schema.Validate(args) != nil {
internal/codextools/registry.go:339: if r.owner.ThreadID == "" || owner != r.owner {
```

명령: `rg -n '^func Test|^package ' internal/codextools`

```text
internal/codextools/json.go:1:package codextools
internal/codextools/registry_test.go:1:package codextools
internal/codextools/registry_test.go:33:func TestHybridPositiveAndNoThreadRecreation(t *testing.T) {
internal/codextools/registry_test.go:56:func TestSixNegativeGroups(t *testing.T) {
internal/codextools/registry_test.go:145:func TestSnapshotRevocationAndDuplicateDiscovery(t *testing.T) {
internal/codextools/registry_test.go:161:func TestSchemaCanonicalizationAndLocalReferences(t *testing.T) {
internal/codextools/registry_test.go:182:func TestStrictArgumentsAndPlaceholder(t *testing.T) {
internal/codextools/registry_test.go:197:func TestNativeToolWireDiscriminator(t *testing.T) {
internal/codextools/registry_test.go:209:func TestRevocationBeforeInvocation(t *testing.T) {
internal/codextools/registry_test.go:217:func TestPrepareBeforeServerAllocatesThread(t *testing.T) {
internal/codextools/registry.go:3:package codextools
```

### 감사자가 추가한 임시 overlay 시험

부모의 별도 허용 후 `/tmp/t651-audit-*`에만 시험 파일과 Go overlay 매핑을 만들었다. `subprocess.run(..., timeout=55)`로 실행을 제한하고 `finally: shutil.rmtree(tmp)`로 삭제했다. 소스 트리에 시험을 추가하지 않았다.

명령 형태: `go test -race -overlay <temporary-overlay.json> -run TestAudit -v ./internal/codextools`

```text
=== RUN   TestAuditActualClaudeCapture
    audit_extra_test.go:5: json: cannot unmarshal string into Go struct field .requests.tool_references of type codextools.Reference
--- FAIL: TestAuditActualClaudeCapture (0.00s)
=== RUN   TestAuditDuplicateBeginConcurrent
--- PASS: TestAuditDuplicateBeginConcurrent (0.00s)
=== RUN   TestAuditAtomicDiscoveryAndForeignBindings
--- PASS: TestAuditAtomicDiscoveryAndForeignBindings (0.00s)
=== RUN   TestAuditSchemas
--- PASS: TestAuditSchemas (0.00s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/codextools	0.391s
FAIL
AUDIT_PROBE_EXIT=1
AUDIT_TEMP_CLEANED=true
```

첫 실패는 감사 시험이 포착 파일의 요약 문자열 배열을 typed reference 배열로 읽은 오류다. 제품 코드 결함으로 판정하지 않았다. 도구 배열만 읽고 명시적인 typed reference를 제공하도록 감사 시험을 바로잡은 뒤 다시 실행했다.

명령 형태: `go test -race -overlay <temporary-overlay.json> -run TestAuditActualClaudeCapture -v ./internal/codextools`

```text
=== RUN   TestAuditActualClaudeCapture
    audit_extra_test.go:10: initial=12 followup=13 native=12 captured schema registry accepted
--- PASS: TestAuditActualClaudeCapture (0.01s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/codextools	1.386s
AUDIT_PROBE_EXIT=0
AUDIT_TEMP_CLEANED=true
```

추가 시험의 단정:
- 같은 call ID로 8개 goroutine이 Begin을 호출해 정확히 하나만 승인된다.
- 발견 목록에 유효 도구와 미등록 도구를 함께 넣으면 전체 발견이 거절되고 일부 등록이 남지 않는다.
- 다른 thread 또는 account binding으로 발견·호출할 수 없다.
- oneOf, minimum, pattern 및 draft-07 schema에서 잘못된 인수를 거절하고 양성 인수는 승인한다.
- 기존 Claude 2.1.269 포착 파일의 도구 목록 12개에서 13개로 늘어도 native 목록은 dispatcher 포함 12개로 고정된다.

## Baseline-attribution

명령: `git -C /Users/goos/MoAI/moai-adk-go/.claude/worktrees/t651 rev-parse --short HEAD`

```text
dad5e9a9e
```

감사 시작과 종료에 세 파일의 SHA-256을 재측정했으며 값이 같았다.

```text
2358b4faa2392e0fa1daf04c66616b313a78b454ffe698ba035b872a6de9891f  internal/codextools/json.go
bb0b5a8fb2b0fc4437a55351f4cbaeca82328434f5891a93e12acd06aed56125  internal/codextools/registry.go
f9723552661572461ae718f38a1670205705ec5331a1fd4e4bc14dd4b4c16406  internal/codextools/registry_test.go
```

시작 상태는 `.moai/reports/t651/`, `internal/codextools/`가 untracked였다. 이 감사는 해당 미커밋 내용에 대한 판정이다. `git check-ignore .moai/reports/t651/code-audit.md` 출력 없음으로 보고서가 무시 규칙에 잡히지 않음을 확인했다.

## Findings

차단 결함 0건. 선택 결함 0건. 이번에 기계적으로 재현한 제품 결함은 없다.

## Gaps

- AS-004/AS-005/AS-006 전체의 실제 Claude 실행 횟수, HTTP 연결, GPT 최종 답변, image/tool failure 결과는 검사하지 않았다.
- 포착 파일은 이전에 수집한 실제 Claude 요청이다. 이번에는 해당 데이터를 현재 코드에 입력했으며 새 Claude 프로세스를 실행한 것은 아니다.
- 임시 감사 시험은 정리했으므로 저장소 회귀 시험으로 남지 않는다. 핵심 기존 회귀 시험은 registry_test.go에 있다.
- 의존성 이름·버전·사용 및 외부 참조 차단을 검사했다. 최신 취약점 데이터베이스 조회나 전 저장소 dependency audit는 수행하지 않았다.
- authenticated current snapshot 및 Binding이 신뢰할 수 있는 요청에서 유래하는지는 호출자 책임이다. 호출자 통합은 아직 이번 범위 밖이다.

## Residual-risk

이 패키지는 도구를 실행하지 않는다. Begin 이후 실제 Claude 실행까지의 시점 차이, 현재 요청의 인증, App Server thread/turn과 HTTP 대화의 연결, 프로세스 장애 복구가 상위 계층에서 올바르게 구현되어야 한다. 패키지 PASS를 전체 카드 완료 또는 배포 승인의 근거로 단독 사용해서는 안 된다.

## Iteration history

1. 기존 패키지 및 race/coverage/vet/format 검사 완료.
2. 추가 감사 시험 중 capture parser 오류를 확인하고 제품 결함과 분리.
3. 감사 parser 수정 후 포착 schema replay PASS. 프로덕션 소스 해시 변화 없음.
