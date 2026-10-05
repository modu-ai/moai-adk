# MoAI MCP Claude 독립 감사 구현·테스트 체크리스트

> 상태: 구현 완료 · 영향 범위 검증 GREEN · 외부 모델 본응답 acceptance 일부 미관찰
> 구현 대상: `claude_audit`와 source-aware `audit_multi`
> 구현 기준선: `.claude/worktrees/develop`, `develop@ba60eb6d5`의 미커밋 작업 트리
> 작성일: 2026-09-14
> 원칙: 실행하지 않은 검사는 PASS로 표기하지 않고, 외부 provider capacity는 제품 성공이나 제품 결함으로 바꾸어 해석하지 않는다.

---

## 1. 결과 요약

Claude 독립 감사 기능의 제품 코드, 설정, MCP catalog, settings UI, canonical rule·skill·agent, template mirror를 구현했다. GPT와 GLM origin은 호출자가 넣은 `claude_verdict`를 무시하고 공식 Claude Code subscription backend를 실행한다. Claude origin은 기존 대화 세션의 anchor를 재사용할 수 있으며, 두 경로는 provenance의 `source`로 구분된다.

영향 범위 단위·통합·race·vet·Windows cross-compile·template target parity 검사는 통과했다. 실제 Claude Code subscription도 GPT origin과 GLM origin에서 first-party transport까지 도달했고, 도구 비활성화·세션 비저장·gateway 환경 제거·fixture 불변을 확인했다. 다만 provider가 두 호출에 HTTP 429 capacity 상태를 반환했으므로 실제 review verdict, resolved Claude model, token usage는 관찰하지 못했다.

판정 용어는 다음과 같다.

- `PASS`: 이 실행에서 기계적으로 관찰한 요구사항
- `PARTIAL`: 일부 기계적 계약은 통과했으나 외부 runtime 증거가 남은 요구사항
- `NOT-OBSERVED`: 실행되지 않았거나 외부 상태 때문에 결과를 관찰하지 못한 항목
- `BASELINE GAP`: 이번 변경과 무관한 기존 결손 때문에 전체 검사가 GREEN이 아닌 항목

---

## 2. Acceptance Criteria 판정

| AC | 검증 대상 | 판정 | 이번 실행의 근거 | 남은 항목 |
|---|---|---|---|---|
| AC-CLA-001 | MCP 공개 표면 | PASS | `claude_audit` registration·catalog count·이름 일치 테스트 통과 | 없음 |
| AC-CLA-002 | 입력 schema | PASS | `target`, `focus`, `model`, `effort`, `project_root` schema 및 closed enum 테스트 통과 | evolving model ID는 의도적으로 free text 허용 |
| AC-CLA-003 | read-only 보장 | PASS | exact argv에 `--safe-mode`, `--restricted`, `--tools ""`, strict MCP, no persistence 포함; live fixture hash 불변 | 실제 모델 본응답은 미관찰 |
| AC-CLA-004 | shell injection 방지 | PASS | `exec.CommandContext` argv·stdin 경계, focus·model·diff 및 `;$(...)` 포함 project path의 literal 처리 테스트 통과; production shell 호출 검색 0건 | 없음 |
| AC-CLA-005 | provider 격리 | PASS | 모든 `ANTHROPIC_*`, `CLAUDE_CODE_*`, `CLAUDECODE`와 지정 gateway 변수 scrub negative tests 통과 | 새 provider 변수 추가 시 scrub 정책 회귀 가능 |
| AC-CLA-006 | subscription provenance | PASS | sanitized `claude auth status --json`에서 `loggedIn=true`, `authMethod=claude.ai`, `apiProvider=firstParty`, subscription 확인; GPT/GLM live 호출도 auth gate 통과 | PII 값은 의도적으로 기록하지 않음 |
| AC-CLA-007 | model provenance | PARTIAL | requested/resolved model, effort, usage null 의미론, strict structured schema, alias fallback·non-Claude mismatch, binary-missing 기본 provenance 단위 테스트 통과 | HTTP 429로 실제 resolved model·usage는 NOT-OBSERVED |
| AC-CLA-008 | review material | PASS | 공통 `collectReviewDiff()` 사용, 빈 diff면 auth/model 실행 없이 inconclusive 처리 테스트 통과 | 없음 |
| AC-CLA-009 | GPT main 교차 감사 | PASS | GPT origin이 caller anchor를 무시하고 Claude backend를 호출하는 통합 테스트 및 live transport 통과 | 실제 verdict는 capacity로 NOT-OBSERVED |
| AC-CLA-010 | GLM main 교차 감사 | PASS | GLM origin이 caller anchor를 무시하고 Claude backend를 호출하는 통합 테스트 및 live transport 통과 | 실제 verdict는 capacity로 NOT-OBSERVED |
| AC-CLA-011 | Claude main 호환 | PASS | `moai cc` launcher가 `MOAI_LAUNCH_PROVIDER=claude`를 child에 실제 주입하고 valid caller anchor만 `source=in_session_anchor`로 사용하는 테스트 통과 | 없음 |
| AC-CLA-012 | required gate | PASS | GPT/GLM origin의 Claude inconclusive가 overall fail과 `gate_unmet=claude`를 만드는 테스트 통과 | 없음 |
| AC-CLA-013 | 독립성 | PASS | backend별 입력을 독립 생성하고 다른 verdict를 prompt에 넣지 않는 participant 테스트 통과 | 없음 |
| AC-CLA-014 | cancellation | PARTIAL | Unix 실제 자식 process tree 종료 테스트 통과; Windows Job Object 구현 cross-compile 통과 | Windows 호스트 runtime 테스트 NOT-RUN |
| AC-CLA-015 | secret·PII 비기록 | PASS | auth PII 비복사, raw capacity 결과 비노출, 환경 sentinel scrub, prompt·diff 비지속 테스트 통과 | 운영 로그 집계 계층은 이번 범위 밖 |
| AC-CLA-016 | 설정 보존 | PASS | config defaults, model·effort 독립 override, effort-only project pin, key inventory, settings schema, web i18n/policy 검사 통과 | 없음 |
| AC-CLA-017 | 문서·배포 parity | PARTIAL | 변경 대상 rule·skill exact mirror, agent MCP Audit Tools section parity, 세 target catalog hash 검사 통과 | 전체 manifest는 기존 `moai`, `manager-git` hash 불일치로 BASELINE GAP |
| AC-CLA-018 | 범위 비회귀 | PASS | Codex/GLM audit, 실제 `moai cc`·`moai glm` origin marker, convergence, gate, build identity를 포함한 scoped regression 3회 반복과 race 통과 | full repository suite·CI NOT-RUN |

---

## 3. TDD RED → GREEN 원장

| 단계 | RED에서 관찰한 실패 | 구현 후 GREEN |
|---|---|---|
| Claude core | `performClaudeAuditWith`, request·transport·schema·error 상수 미정의로 compile 실패 | Claude core test package `ok` |
| MCP registration·fan-out | GPT/GLM이 caller anchor 사용, Claude origin source 없음, unknown origin refusal 없음, catalog 29개 | source-aware fan-out·단일 registration·catalog 30개로 `cli`, `mcp` 통과 |
| 실제 runner | helper JSON이 stdout에 잡히지 않음 | Go 평가 순서에 따른 capture bug 수정 후 실제 subprocess runner 통과 |
| settings | `workflow.audit.claude.model` field 없음 | Claude model·effort schema, defaults, inventory, i18n 추가 후 통과 |
| defaults | Claude model과 effort가 빈 값 | `sonnet/high` 기본값 추가 후 통과 |
| bounded output | `CLAUDE_OUTPUT_TRUNCATED` 미정의 | 1 MiB bounded stdout/stderr와 truncation sentinel 추가 후 통과 |
| 오류 분류 | auth/model/binary 오류 상수 미정의 | fail-closed taxonomy 구현 후 통과 |
| 환경 격리 | `CLAUDECODE`, 미래 `CLAUDE_CODE_*` 변수가 child env에 남음 | exact `CLAUDECODE`와 prefix scrub으로 통과 |
| web | `claude_audit` enable title·description 누락 | 네 locale 추가 후 web 검사 통과 |
| widget policy | Claude model이 closed-domain 필드로 오인됨 | evolving official model ID를 허용하는 명시 whitelist로 통과 |
| live 최초 | GPT/GLM 모두 `CLAUDE_OUTPUT_MALFORMED`로 분류 | JSON envelope의 `api_error_status=429`를 capacity로 분류하고 subscription provenance 유지 후 live test 통과 |
| model fallback | `opus` 요청이 `claude-sonnet-*` 응답으로 바뀌어도 pass로 수용 | requested alias와 resolved model family가 다르면 `CLAUDE_PROVIDER_MISMATCH`로 fail-closed |
| strict result schema | summary 누락, unknown field, invalid severity가 수용됨 | required pointer field·severity enum·unknown-field 거부로 통과 |
| 실제 launcher origin | `moai cc`와 legacy `moai glm` child에서 `MOAI_LAUNCH_PROVIDER`가 비어 있음 | 공통 launcher wrapper가 각각 `claude`, `glm`을 child에 전달하고 반환 경로에서 환경을 복원 |
| anchor 유효성 | `overall_verdict=banana`인 Claude anchor가 수렴에 참여 | `pass|fail|inconclusive`만 valid anchor로 인정하고 나머지는 독립 backend 실행 |
| exit-zero capacity | `api_error_status=429` envelope가 process exit 0이면 malformed로 분류 | exit code와 무관하게 envelope status를 검사해 `CLAUDE_CAPACITY_UNAVAILABLE`로 분류 |
| effort-only pin | project pin이 effort만 지정하면 기본 `high`로 덮임 | model과 effort를 독립 해석해 `sonnet/xhigh` 보존 |
| binary-missing provenance | Claude binary 탐색 실패 시 requested model·effort가 빈 값 | 실행 파일 탐색 전에 default/project/explicit request를 해석해 `sonnet/high` 기록 |
| 파일 책임 분리 | `mcp_claude.go`가 564줄로 실행·protocol 책임을 함께 가짐 | core 303줄, runner 77줄, protocol 207줄로 분리하고 집중 회귀 통과 |

초기 Claude core RED의 실제 compiler 출력:

```text
undefined: performClaudeAuditWith
undefined: claudeAuditRequest
undefined: claudeAuditTransport
undefined: claudeAuditSchema
FAIL
```

기본 설정 RED의 실제 출력:

```text
default Claude.Model = "", want sonnet
default Claude.Effort = "", want high
```

---

## 4. 실행 증거

### 4.1 공식 Claude Code subscription 상태

민감 필드를 제외한 read-only 확인:

```text
$ claude auth status --json | jq '{loggedIn,authMethod,apiProvider,subscriptionType}'
{
  "loggedIn": true,
  "authMethod": "claude.ai",
  "apiProvider": "firstParty",
  "subscriptionType": "max"
}
```

CLI 기준선:

```text
$ claude --version
2.1.270 (Claude Code)
```

### 4.2 영향 범위 회귀

```text
$ go test ./internal/cli -run '^(TestClaudeAudit|TestClaudeRealRunner|TestParseClaudeAudit|TestResolveClaudeAudit)' -count=3
ok github.com/modu-ai/moai-adk/internal/cli 11.539s

$ go test ./internal/cli -run '^Test(AuditMulti|RunMultiAudit|Converge|DefaultBackendCaller|GateOr|AuditVerdict|PersistedConvergence|BuildIdentity|AuditCompletes|AuditLag)' -count=3
ok github.com/modu-ai/moai-adk/internal/cli 39.378s

$ go test ./internal/cli -run '^(TestCCCmd_Execution_NoDeps|TestGLMCmd_AddsModelOverrides)$' -count=3
ok github.com/modu-ai/moai-adk/internal/cli 0.934s

$ go test ./internal/config ./internal/mcp ./internal/settings -run '^(TestAudit|TestMoaiMCP|TestClaude)' -count=1 -v
ok github.com/modu-ai/moai-adk/internal/config 0.298s
ok github.com/modu-ai/moai-adk/internal/mcp 0.419s
ok github.com/modu-ai/moai-adk/internal/settings 0.803s

$ go test ./internal/web -run '^(TestClosedSetFieldsAreClosedWidgets|TestClosedSetRejectsOutOfSetValue|TestFreeTextWhitelist)$' -count=1 -v
ok github.com/modu-ai/moai-adk/internal/web 0.870s
```

### 4.3 race·vet·portable build

```text
$ go test -race ./internal/cli -run '^Test(ClaudeAudit|ClaudeRealRunner|ParseClaudeAudit|ResolveClaudeAudit|AuditMulti|RunMultiAudit|Converge|DefaultBackendCaller|GateOr|CCCmd_Execution_NoDeps|GLMCmd_AddsModelOverrides|AuditVerdict|PersistedConvergence|BuildIdentity|AuditCompletes|AuditLag)' -count=1
ok github.com/modu-ai/moai-adk/internal/cli 20.977s

$ go test ./internal/cli -run '^TestClaudeAudit_ProjectPathIsLiteralData_AC_CLA_004$' -count=1 -v
--- PASS: TestClaudeAudit_ProjectPathIsLiteralData_AC_CLA_004 (0.38s)
ok github.com/modu-ai/moai-adk/internal/cli 1.313s

$ go vet ./internal/cli ./internal/config ./internal/mcp ./internal/settings ./internal/web ./internal/template
[exit 0, no output]

$ /Users/goos/go/bin/staticcheck ./internal/cli ./internal/config ./internal/mcp ./internal/settings ./internal/template ./internal/web
[exit 0, no output]

$ GOOS=windows GOARCH=amd64 go test -c -o <temporary-output> ./internal/cli
windows_compile=PASS
```

새 Claude 구현 파일에 대한 statement coverage 집계:

```text
claude_audit_statement_coverage=200/229 87.3%
```

이 수치는 `internal/cli` 전체 package coverage가 아니라 이번에 추가한 Claude 구현 파일의 profile statement를 집계한 값이다.

### 4.4 template·agent 배포 표면

```text
$ make agents-emit-check
[exit 0]

$ go test ./internal/template -run '^TestClaudeAuditTemplateSurfacesAndCatalogHash$' -count=1
ok github.com/modu-ai/moai-adk/internal/template 0.637s
```

대상 검사는 다음을 확인한다.

- canonical/template MCP rules와 cross-model audit skill mirror
- plan/sync auditor의 `## MCP Audit Tools` section
- 생성된 Codex agent TOML
- `moai-ref-cross-model-audit`, `plan-auditor`, `sync-auditor` catalog hash

### 4.5 실구독 GPT·GLM origin read-only acceptance

```text
$ MOAI_LIVE_CLAUDE_AUDIT=1 go test ./internal/cli -run '^TestLiveClaudeAudit_GPTAndGLMOriginsUseSubscriptionReadOnly' -count=1 -v
gpt origin reached verified Claude subscription transport; provider capacity was unavailable
glm origin reached verified Claude subscription transport; provider capacity was unavailable
PASS
ok github.com/modu-ai/moai-adk/internal/cli 5.318s
```

이 PASS가 의미하는 범위:

- GPT·GLM origin이 caller anchor 대신 실제 Claude CLI 경로를 실행함
- first-party subscription auth gate 통과
- provider/gateway secret sentinel이 child로 전달되지 않음
- Claude tool surface 없음, session persistence 없음
- fixture 파일의 감사 전후 hash 동일
- HTTP 429를 `CLAUDE_CAPACITY_UNAVAILABLE`과 `inconclusive`로 fail-closed 처리

이 PASS가 의미하지 않는 범위:

- Claude가 실제 diff를 판정했다는 주장
- resolved Claude model이 실제 응답에서 확인됐다는 주장
- 실제 token usage가 수집됐다는 주장

---

## 5. 정적 보안 체크

- [x] production `mcp_claude*.go`에서 `sh -c`, `/bin/sh`, `dangerously-skip-permissions`, `bypassPermissions`, `--bare` 사용 없음
- [x] 명령은 `exec.CommandContext`와 고정 argv를 사용하고 prompt·diff는 stdin으로 전달
- [x] stdout·stderr 각각 1 MiB 상한과 truncation 오류 존재
- [x] MCP context 취소·timeout이 process와 Unix 자식 process group을 종료
- [x] Windows Job Object가 close 시 자식 process 종료 정책으로 컴파일됨
- [x] 모든 `ANTHROPIC_*`, 모든 `CLAUDE_CODE_*`, exact `CLAUDECODE` 제거
- [x] `Z_AI_API_KEY`, `MOAI_BACKUP_AUTH_TOKEN`, `MOAI_LAUNCH_PROVIDER`, `ENABLE_TOOL_SEARCH`, `API_TIMEOUT_MS` 제거
- [x] auth JSON의 email·organization 값을 result에 복사하지 않음
- [x] raw prompt·diff·CLI envelope를 파일에 저장하는 production 경로 없음
- [x] capacity 오류에서 raw provider message를 result에 노출하지 않음

---

## 6. 기준선에서 발견된 별도 GAP

### 6.1 무필터 `internal/cli` baseline

착수 전 다음 범위를 실행했다.

```text
$ go test ./internal/cli ./internal/config ./internal/mcp ./internal/settings
ok github.com/modu-ai/moai-adk/internal/config
ok github.com/modu-ai/moai-adk/internal/mcp
ok github.com/modu-ai/moai-adk/internal/settings
FAIL github.com/modu-ai/moai-adk/internal/cli
moai embeds stale agent-emit artifacts: manager-git.toml, manager-spec.toml
[terminated after approximately 601s]
```

이 결과는 이번 Claude 변경의 회귀로 분류하지 않았다. 실패가 구현 전 작업 트리의 다른 generated artifact 상태를 지목했고, Claude 영향 범위 회귀는 별도로 GREEN을 관찰했기 때문이다. 반대로 이 판단은 전체 `internal/cli`가 통과했다는 뜻도 아니다.

### 6.2 전체 template catalog·manifest hash

```text
$ go test ./internal/template -run '(Claude|Audit|Catalog|ManifestHash|Widget)' -count=1
TestCatalogHashCoversSkillSubfiles: CATALOG_HASH_SKINNY: moai stored 8f... computed 94...
CATALOG_HASH_UNSTABLE: moai stored 8f... computed 94...
CATALOG_HASH_UNSTABLE: manager-git stored 676... computed 395...
FAIL
```

이번 변경 대상 세 entry의 hash는 계산값과 일치하지만, 기존 `moai`, `manager-git` 불일치는 남아 있다. generator가 이 두 값을 바꾸려 한 결과는 다른 세션 소유 변경으로 판단해 흡수하지 않고 원래 상태로 복원했다.

---

## 7. 변경 표면 체크리스트

### 제품 코드

- [x] `internal/cli/mcp_claude.go`
- [x] `internal/cli/mcp_claude_runner.go`
- [x] `internal/cli/mcp_claude_protocol.go`
- [x] `internal/cli/mcp_claude_process_unix.go`
- [x] `internal/cli/mcp_claude_process_windows.go`
- [x] `internal/cli/launcher.go`와 `internal/config/envkeys.go` origin marker
- [x] `internal/cli/mcp_server.go`
- [x] `internal/cli/mcp_audit_multi.go`
- [x] `internal/cli/mcp_convergence.go`
- [x] `internal/cli/mcp_codex.go` 공통 provenance 확장
- [x] `internal/config/audit_models.go`
- [x] `internal/config/defaults.go`
- [x] `internal/mcp/catalog.go`
- [x] `internal/settings/schema_sections.go`
- [x] `internal/web/assets/i18n.js`

### 테스트

- [x] Claude runner·auth·schema·error·scrub 단위 테스트
- [x] Unix 실제 process-tree cancellation 테스트
- [x] GPT·GLM live subscription read-only 테스트
- [x] audit_multi source-aware fan-out·gate·participant 테스트
- [x] catalog·settings·default·inventory 테스트
- [x] canonical/template/agent/hash surface 테스트
- [x] build identity 공통 계약에 `claude_audit` 포함

### 문서·template

- [x] canonical MCP tool rule·catalogue
- [x] cross-model audit skill
- [x] plan/sync auditor
- [x] project와 template workflow YAML
- [x] template Claude surfaces
- [x] generated Codex plan/sync auditor TOML
- [x] 변경 대상 catalog hash

---

## 8. Claim

1. GPT와 GLM main session에서 Claude 감사는 caller가 꾸민 anchor가 아니라 공식 Claude Code subscription subprocess로 실행된다.
2. Claude main session은 기존 in-session anchor를 사용할 수 있고, 실제 subscription call과 provenance로 구분된다.
3. subprocess는 shell을 거치지 않고 도구·MCP·session persistence를 비활성화하며 gateway/provider 환경을 제거한다.
4. Claude inconclusive는 GPT/GLM required gate에서 성공으로 완화되지 않는다.
5. 영향 범위 검증은 GREEN이며, 실제 모델 본응답만 provider capacity로 미관찰이다.

## 9. Evidence

근거는 이 문서 3장의 RED→GREEN 원장, 4장의 명령·출력, 5장의 정적 보안 체크, 6장의 별도 baseline GAP에 기록했다. 제품 구현의 핵심 기계적 경계는 `mcp_claude.go`의 orchestration·environment scrub, `mcp_claude_runner.go`의 argument-vector 실행, `mcp_claude_protocol.go`의 subscription auth·strict structured output validation, `launcher.go`의 trusted origin marker, `mcp_convergence.go`의 source-aware fan-out에서 확인했다.

## 10. Baseline-attribution

모든 구현·검증 주장은 2026-09-14의 `.claude/worktrees/develop`, `develop@ba60eb6d5` 미커밋 작업 트리에 귀속한다. live transport 주장은 같은 실행의 `/Users/goos/.local/bin/claude` 2.1.270과 당시 first-party subscription 상태에만 귀속한다. CI, 원격 `develop`, release binary, 다른 OS runtime에는 귀속하지 않는다.

## 11. Gaps

- 실제 Claude review verdict, resolved model, token usage는 provider HTTP 429 capacity로 `NOT-OBSERVED`다.
- Windows process-tree cancellation은 compile PASS이며 runtime은 `NOT-RUN`이다.
- full repository suite와 CI는 `NOT-RUN`이다.
- 전체 template manifest에는 이번 변경과 무관한 기존 hash GAP 두 건이 남아 있다.
- commit, push, PR은 수행하지 않았다.

## 12. Residual-risk

- provider allowance가 회복된 뒤 같은 gated live test를 다시 실행해야 실제 model provenance acceptance를 닫을 수 있다.
- Claude Code CLI의 envelope나 option 계약이 바뀌면 parser가 fail-closed하면서 감사 가용성이 중단될 수 있다.
- Windows Job Object의 descendant termination은 Windows runtime에서 별도 관찰해야 한다.
- 통합 담당자가 dirty worktree 전체를 sweep-stage하면 다른 세션의 Gateway·SPEC·orchestration 변경이 섞일 수 있다. 명시 경로 staging 전 branch·HEAD·status를 다시 읽어야 한다.

---

## 13. 최종 판정

제품 구현과 로컬 영향 범위 검증은 완료됐다. GPT·GLM에서 Claude subscription 독립 감사로 진입하는 경로와 읽기 전용·fail-closed 안전성은 확인했다. 외부 provider capacity 때문에 실제 Claude 판정과 resolved model을 관찰하지 못했으므로 전체 운영 acceptance는 `PARTIAL`, 제품 구현 판정은 `PASS`다.
