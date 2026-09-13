# 실계정 시험 전 AC 연결 지도

## Claim

active AC 60개의 기존 로컬 증거와 남은 구현·실측·플랫폼 조건을 정리했다. MG-002는 폐기 묘비라 제외한다. 전체 PASS나 독립 감사 판정이 아니다. 아래 명령·출력은 기존 보고서의 기록이며 이번 읽기 전용 작업에서 재실행한 시험이 아니다.

## Evidence — 먼저 연결할 생산 경로

- root.go:153은 newGatewayChildCommand(nil)이다. child RunE는 factory nil이면 즉시 오류다. M0 후 private payload에서 catalog/credential/adapter/Server를 조립할 factory가 필요하다.
- launcher.go:123의 unifiedLaunchWithGateway에는 nil binding이 전달된다. gpt.go:54는 transport 검증 대기 오류다. 세 launcher 실제 normal/continue/resume/spawn/worktree/factory 경로에 검증된 binding을 연결해야 한다.
- NewSessionCatalog는 ContextTokens를 0으로 둔다. OpenAI/Anthropic adapter는 양의 ContextTokens와 MeasureInput을 요구한다. 실제 모델별 한도와 로컬 측정 factory가 아직 없다.
- raw thinking/context_management/output_config·opaque는 일반 변환에서 거절된다. R은 tool marker/opaque hash 왕복 후보일 뿐 구현·실측 PASS가 아니다.
- PICKER Run은 pending이다. CLI 초기 모델 준비 외 modelPicker·설정 저장 격리·Default·fallback의 실제 연결이 필요하다.
- Store.Refresh production 검색은 정의만 반환했다. Login/Logout/Status CLI 밖에 만료 시 공식 refresh와 공급자 verifier를 호출하는 연결이 필요하다.
- gpt Logout은 Store.Logout만 호출하고 remote revocation 미요청을 출력한다. AUTH plan B.5의 소유 snapshot remote 시도/미지원 판정을 별도로 연결해야 한다.
- TEAMMATE는 설치 binary seam 읽기만 있다. named teammate override 실측 후 single-use private bootstrap·A/B/C 소유권·수명 구현이 필요하다.
- Windows OpenStore gate는 유지된다. native helper 후보의 compile를 ACL/내구성 런타임 증거로 대신할 수 없다.

## 기존 증거 색인

- **F** [foundation-verification.md](foundation-verification.md): 명령·원문 관측은 해당 보고서 Evidence 참조
- **I** [ingress-verification.md](ingress-verification.md): 명령·원문 관측은 해당 보고서 Evidence 참조
- **S** [supervisor-verification.md](supervisor-verification.md): unset ANTHROPIC_BASE_URL ANTHROPIC_API_KEY ANTHROPIC_AUTH_TOKEN Z_AI_API_KEY && GOCACHE=/tmp/gateway-foundation-cache go test ./internal/gateway ./internal/gateway/auth -race -timeout 60s
- **L** [cli-integration-verification.md](cli-integration-verification.md): unset ANTHROPIC_BASE_URL ANTHROPIC_API_KEY ANTHROPIC_AUTH_TOKEN Z_AI_API_KEY && MOAI_HOME=/tmp/gateway-cli-verification-home GOCACHE=/tmp/gateway-foundation-cache go test -race ./internal/cli ./internal/hook -run 'TestGateway|TestGPT' -timeout 90s
- **A** [auth-implementation-verification.md](auth-implementation-verification.md): $ go test ./internal/gateway/auth -race -count=1 -timeout=30s
- **AD** [auth-supervisor-code-audit-iter2.md](auth-supervisor-code-audit-iter2.md): unset ANTHROPIC_BASE_URL ANTHROPIC_API_KEY ANTHROPIC_AUTH_TOKEN Z_AI_API_KEY && MOAI_HOME=/tmp/gateway-auth-audit-home GOCACHE=/tmp/gateway-auth-audit-cache go test -race ./internal/gateway/auth -run 'TestStore|TestRefresh|TestCrossProcess|TestAuthStore|TestCodexBroker|TestSend|TestEarlyResponse|TestHeaderBoundary|TestCredentialReferences|TestBroker|TestScratch|TestExplicitAPI' -count=1 -timeout 30s
- **T** [translation-verification.md](translation-verification.md): $ go test -race ./internal/gateway/translate -count=1
- **TD** [protocol-functional-review-iter2.md](protocol-functional-review-iter2.md): $ GOCACHE=/tmp/gateway-protocol-review-cache go test -race -overlay .moai/state/verify/01a08e7b-6aa0-7361-ab7e-ea8da1f02228/functional-review-iter2/overlay.json ./internal/gateway/translate -run '^TestFunctionalDelta|^TestFunctionalEveryTruncated|^TestFunctionalSameItem|^TestStreamBuffered|^TestStreamDeferred|^TestStreamCanonicalOffset' -count=1 -v -timeout 15s
- **O** [openai-adapter-verification.md](openai-adapter-verification.md): $ GOCACHE=/tmp/gateway-translation-cache go test -race ./internal/gateway -run TestOpenAI -count=1 -timeout 30s -coverprofile=/tmp/gateway-openai-adapter-cover.out
- **N** [m6-local-adapters-verification.md](m6-local-adapters-verification.md): $ GOCACHE=/tmp/gateway-translation-cache go test -race ./internal/gateway -run 'TestAnthropicNative|TestGLMNative' -count=1 -timeout 30s -coverprofile=/tmp/gateway-m6-native-cover.out
- **CM** [cg-migration-verification.md](cg-migration-verification.md): `<SCRUB> <ENV> go test -race ./internal/cli ./internal/config -run TestCG -timeout 60s`
- **CR** [cg-retirement-runtime-verification.md](cg-retirement-runtime-verification.md): `<S> <E> go test -race ./internal/cli -run 'CG|Cg|HelpGroup|Fang|FactoryEntry|FactoryWorkerEntry|ExistingBranch' -timeout 90s`
- **CT** [cg-template-verification.md](cg-template-verification.md): `go test ./internal/template -run TestCGEmbeddedRetirement -timeout 30s`
- **CD** [cg-docs-verification.md](cg-docs-verification.md): hugo --source /Users/goos/MoAI/moai-adk-go/.claude/worktrees/moai-proxy-unified/docs-site --destination /tmp/cg-docs-render-20260911 --minify --cacheDir /tmp/cg-docs-hugo-cache-20260911 > /tmp/cg-docs-build-20260911.log 2>&1
- **CB** [cg-docs-browser-verification.md](cg-docs-browser-verification.md): 명령·원문 관측은 해당 보고서 Evidence 참조
- **CF** [cg-functional-consistency-review.md](cg-functional-consistency-review.md): $ unset ANTHROPIC_BASE_URL ANTHROPIC_API_KEY ANTHROPIC_AUTH_TOKEN Z_AI_API_KEY && MOAI_HOME=/tmp/gateway-cg-functional-review-home GOCACHE=/tmp/gateway-protocol-review-cache go test ./internal/cli -run '^TestCGMigrationPreviewApplyAndGuards$|^TestCGRetired|^TestCGRetirement' -count=1 -v -timeout 30s
- **W** [windows-auth-verification.md](windows-auth-verification.md): `GOOS=windows GOARCH=amd64 GOCACHE=/tmp/gateway-foundation-cache go test ./internal/gateway/auth -c -o /tmp/gateway-auth-windows.test.exe`
- **P** [provider-policy-preflight.md](provider-policy-preflight.md): 명령·원문 관측은 해당 보고서 Evidence 참조
- **R** [reasoning-binding-candidate.md](reasoning-binding-candidate.md): 명령·원문 관측은 해당 보고서 Evidence 참조
- **TM** [teammate-seam-preflight.md](teammate-seam-preflight.md): 명령·원문 관측은 해당 보고서 Evidence 참조

## 코어 — active 24개

| AC | 증거/로컬 범위 | 남은 조건 |
|---|---|---|
| AC-MG-001 | L: 초기 argv seam | 구현+실측: 실제 binding과 3launcher 첫 turn 원문 |
| AC-MG-003 | F/I: 2양성·9변형·raw 음성/로컬 응답 | 실측: 실제 gateway 연속 전환·401/404 /model 거절·s 첫 turn |
| AC-MG-004 | T/O/N: 공개 도구 schema·ID 역매핑/TLS | 구현+실측: raw 정책과 3provider Read/Write/Bash |
| AC-MG-005 | L: 설정 merge/private overlay fixture | 실측: 세 launcher 실제 파일 hash/환경·쓰기 대상 |
| AC-MG-006 | S/AD: helper actual exec·parent death·포트 종료 | 실측: Claude exit N/INT/STOP-fg/kill; 플랫폼: Windows release |
| AC-MG-007 | T/TD: 교차 tool/content index 순서 | 로컬 범위 근거; 실제 adapter 조합 확인 |
| AC-MG-008 | T/TD/O/N: EOF·429·취소·성공 terminal 없음 | 로컬 범위 근거; 실제 raw client 중계 |
| AC-MG-009 | T/R: 공개 pair·미완료 거절; carrier 후보 | 구현+실측: opaque provenance/resume/foreign strip/전체 유실; AC BLOCKED |
| AC-MG-010 | O/N: 합성 context 초과·image/PDF 거절 | 구현: 실제 MeasureInput/ContextTokens |
| AC-MG-011 | F/I: 동시 요청 exact routing | 실측: 실제 subagent/background slot 표시·route |
| AC-MG-012 | L: hook guard/settings 보존 | 실측: MCP·승인·compact·GLM web 인증 |
| AC-MG-013 | I/S/O/N: auth/path/model/gzip/redirect/loopback | 최종 통합 로그·비밀표지·listener 확인; 전체 보안 PASS 아님 |
| AC-MG-014 | L: GPT help/backend seam | 실측: factory/kanban4분기 전 구간; badge/상수 최종 증거 연결 |
| AC-MG-015 | L: help/command tree 대조 | 최종 빌드 help 대조 |
| AC-MG-016 | L/AD: POSIX Exec/helper 실제 exec | 구현: production gateway 선행분기 연결 후 확인 |
| AC-MG-017 | S: 127.0.0.1:0 실제 handoff | 최종 CLI/점유 포트 대조 |
| AC-MG-018 | L: 14키/ZAI분리/GLM4슬롯/settings/hook | 실측: 오염 tmux·settings 우선순위·공유경쟁·in-process |
| AC-MG-019 | L: config guard 관련 기록 | 최종 두 named env guard 실행 결과 필요; build로 대체 금지 |
| AC-MG-020 | A/L: 소유 저장소·사용자 경로 fixture | 실측: 실제 GPT 세션 Codex auth hash+mtime 불변 |
| AC-MG-021 | F/I/N: GPT4 exact/unknown 거절·OAuth 경로 부재 | 실측: GPT4 접근·T09 양성 전 passthrough 금지 |
| AC-MG-022 | I/O/N/P: no fallback mock | 구현+실측: client fallbackModel/CLI flag·설정 우회 |
| AC-MG-023 | L: 닫힌 동사·auth 처리기 | 실측: 공식 broker; 최종 command 오류 출력 |
| AC-MG-024 | I: estimate/목록/미등록path | 최종 factory/catalog endpoint 연결 |
| AC-MG-025 | F/L/N: tier4/저장 GLM 키/beta제거 | 실측: ZAI 요청; setup/status/tools 이전 대비 출력·종료코드 |

## AUTH — active 10개

| AC | 증거/로컬 범위 | 남은 조건 |
|---|---|---|
| AC-GA-001 | A/L: mock login→Store/CLI | 실측: 공식 새 로그인·구독 요청 |
| AC-GA-002 | A: operation/loginId/cancel | 실측: 설치 RPC/callback; broker 내부 PKCE 시험 아님 |
| AC-GA-003 | A/AD: timeout/late/강제종료/세대보존 | 실측: 실제 listener 종료·추가교환0 |
| AC-GA-004 | A/AD/W: private state/cleanup·Windows 후보 | 플랫폼: ACL/내구성/Store 연결·native 실행 |
| AC-GA-005 | A: OS process refresh1회/late logout/lock crash | 구현+실측: 생산 refresh/verifier; Windows lock |
| AC-GA-006 | A/AD/O: socket write barrier/logout/Owns | 생산 Store 연결·in-flight 취소 실측; remote 별도 |
| AC-GA-007 | A/O: endpoint/generation/redirect/no fallback | 구현: refresh resolver; 공급자 실패 실측 |
| AC-GA-008 | A/O: 구독/API endpoint 분리 | 실측: 구독 방식/실제 경로 일치 |
| AC-GA-009 | A/L: scratch/CODEX_HOME 격리 | 실측: 네 실제 동작 hash/권한/read-copy 계수·remote scratch |
| AC-GA-010 | T/O: 공개 변환/TLS 기반 | 구현+실측: AUTH+PICKER 실제 GPT4 회상/tool/stream |

## PICKER — active 9개

| AC | 증거/로컬 범위 | 남은 조건 |
|---|---|---|
| AC-GP-001 | L: explicit/default/empty→sol argv | 구현+실측: 첫 turn/표시 |
| AC-GP-002 | L: mode default/env scrub | 실측: 저장 모델 오염·Opus5/Sonnet5·타provider0 |
| AC-GP-003 | L: continue/fallback 준비 | 실측: saved continue/resume×명시모델·재시작 |
| AC-GP-004 | F/I/P: catalog4·probe 준비 | 구현+실측: modelPicker와 /model·s 각4ID |
| AC-GP-005 | L/S: settings/cleanup 후보 | 구현+실측: 실제2세션 persistence/hash/권한 |
| AC-GP-006 | L: InitialModel 값 | 구현+실측: Default 표시/다음 turn 일치 |
| AC-GP-007 | T/O: 합성 tool/stream | 실측: 같은 세션 GPT4 회상·도구 |
| AC-GP-008 | I: credential없는 검증/turn401 | 실측: /model 이전선택 유지 vs s 첫오류/no fallback |
| AC-GP-009 | F/L: GLM슬롯·생산 gate 닫힘 | 구현: PICKER 출시 연결; AUTH 실측 없이 완료 금지 |

## CG-RETIRE — active 8개

| AC | 증거/로컬 범위 | 남은 조건 |
|---|---|---|
| AC-CR-001 | CR/CF: root 제거/help/error/spawn counter | 후속 executable RED→GREEN으로 진단 수리; 아래 Evidence 참조 |
| AC-CR-002 | CM/CR/CF: legacy guard/hash/정상대조 | 생산 gateway 연결 후 guard 순서 |
| AC-CR-003 | CM/CF: preview/apply/lock/hash/실패/재실행 | claude-only 로컬; hybrid actual gate 유지 |
| AC-CR-004 | CR/CF: 주입철거/historical reader | 생산 gateway provider 연결은 코어 |
| AC-CR-005 | CR/CF: 3launcher shape/counter | 실제 정상 WT/factory 전 구간은 별도 |
| AC-CR-006 | CT/CD/CB/CF: 4locale build/대상링크/render/help/명령 | 역사6표본 외 전체 hash 미관측; 후속 executable 진단 수리 증거는 아래 참조 |
| AC-CR-007 | CR/CF: retired/정상대조 counter | 오류객체 시험은 실제 stderr출력을 보장하지 않았음 |
| AC-CR-008 | CM/TM: hybrid failclosed/seam후보 | 구현+실측: TEAMMATE 동등 역할/auth/lifetime |

## TEAMMATE — active 9개

| AC | 증거/로컬 범위 | 남은 조건 |
|---|---|---|
| AC-GT-001 | TM/I: binary seam·일반 sessionauth | 구현+실측: A/B/C private pane·교차거절 |
| AC-GT-002 | TM: helper 관측 계획 | 구현+실측: single-use/비밀0/cleanup |
| AC-GT-003 | L: lead env scrub만 | 구현+실측: stale pane bootstrap/원본tmux/profile |
| AC-GT-004 | S/AD: lead gateway 수명만 | 구현+실측: A pane 종료·B/C 유지 |
| AC-GT-005 | F/I: exact registry 기반 | 구현+실측: 실제 teammate 전모델/unknown0 |
| AC-GT-006 | TM: seam 후보 | 구현+실측: 재사용/동시소비/late완료/누수 |
| AC-GT-007 | TM: 2.1.268 override 코드 존재 | 실측 먼저: named Agent→helper argv/env/paneID/수명 |
| AC-GT-008 | L/P: lead settings·fallback 계획 | 구현+실측: pane permission/MCP/profile/fallback503 |
| AC-GT-009 | TM: 실제 통합 증거 없음 | 구현+실측: AUTH/PICKER A/B/C 전모델 tool/stream/종료 |

## Baseline-attribution

2026-09-11 WT `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/moai-proxy-unified`. 실제 HEAD 조회 출력은 `81c1d58f9`, 현재 브랜치 조회 출력은 `WT-unified-gateway`다. acceptance 5종, 증거 색인의 보고서, 지정 연결 함수 본문을 읽었다. 긴 출력의 주요 factory/refresh/최신 delta 문맥은 따로 읽었다. 기존 보고서 당시 출력은 동일 HEAD라도 이후 uncommitted 변경을 자동 포괄하지 않는다.

## Gaps

이 작성 중 새 시험·제품·SPEC·카드·git mutation·실제 client 호출은 없다. 원본 보고서 출력/해시는 링크를 따른다. 전체 저장소 판정은 통합 브랜치 CI가 소유하며 PENDING이다. 이 표는 실제 계정 미검증을 coverage/lint/delta audit PASS로 바꾸지 않는다. Sonnet4.5 역사 캡처는 향후 Opus5·Sonnet5 실측 대체물이 아니다.

## Residual-risk

여러 writer가 같은 WT에서 진행하므로 읽은 순간의 지도다. 부모가 보고한 cg 실행 진단의 무출력 exit1은 아직 이 작성자가 재현하기 전이며 다음 승인된 수리에서 직접 재현한다. 해당 발견을 기존 error-object 시험 PASS로 덮지 않는다. 새로운 독립 보안 감사는 하지 않았다.

## 상태 metadata 관측

5개 SPEC의 spec.md는 모두 `status: draft`다. progress.md에는 YAML status 앞머리 없이 Run Evidence 섹션이 있다. 코어/AUTH/CG의 구현 보고서가 존재하므로 draft 표기를 구현 미착수 증거로 사용하지 않았다. 이 작업에서는 상태를 바꾸지 않았다.

읽은 `/Users/goos/MoAI/moai-adk-go/.agents/skills/moai-run/SKILL.md`는 moai skill에 run 인수를 넘기는 wrapper이며 자체 상태 전이 규칙은 없다. 현재 loaded manager-develop 계약은 첫 M1 run commit에서 네 plan-phase artifact(spec/plan/acceptance/progress)의 draft→in-progress와 updated 갱신을 구현 담당에게 허용하고 implemented/completed는 manager-docs에게 둔다. commit 금지 상태에서 자동 전이를 만들지 않고 부모의 후속 소유권 판단에 넘긴다.

## 후속 승인 수리 — 실제 cg 진단 출력

부모 관측을 별도 test executable로 직접 재현했다. cmd/moai/main.go와 같은 Execute→ResolveExitCode→os.Exit 경계를 사용하고 임시 HOME/MOAI_HOME, context timeout, Command.Run의 Wait 회수로 세 조합을 시험했다. 원래 error 객체만 검사한 시험과 다르다.

환경 정리 및 GOCACHE/MOAI_HOME은 L과 같은 단일 invocation으로 적용했다.

`go test ./internal/cli -run '^TestCGRetiredExecutableDiagnostic$' -count=1 -timeout 20s`

```text
--- FAIL: TestCGRetiredExecutableDiagnostic (0.32s)
    cg_retirement_test.go:217: [cg] stdout="" stderr=""
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	1.253s
FAIL
```

root.go의 pre-Cobra CG guard에서 기존 moaiErrorHandler를 한 번 호출하고 동일 errCGRetired를 반환하도록 수정했다. main.go·dependency 초기화·provider·기존 guard 순서를 바꾸지 않았다. 신규 시험은 cg, cg --help, cg --spawn -f 2 각각 exit1/stdout0/stderr의 이전 명령 1회를 판정한다.

`go test ./internal/cli -run '^TestCGRetired|^TestCGRetirement|^TestCGMigrationPreviewApplyAndGuards$' -count=1 -timeout 30s`

```text
ok  	github.com/modu-ai/moai-adk/internal/cli	3.929s
```

CR-001/006/007의 진단 공백은 위 main-equivalent executable 경계에서 해결했다. 최종 실제 제품 바이너리 재빌드 smoke는 부모 소관이다. hybrid capability gate를 재현할 입력은 `migrate cg --target claude-glm --apply`이며 claude-only 전용 accept-role-change를 섞지 않는다.

추가 범위 검증:

`go test -race ./internal/cli -run '^TestCGRetiredExecutableDiagnostic$|^TestCGRetiredTopLevelDiagnostic$' -count=1 -timeout 20s`

```text
ok  	github.com/modu-ai/moai-adk/internal/cli	3.303s
```

`go vet ./internal/cli`: exit 0, stdout/stderr 없음.
