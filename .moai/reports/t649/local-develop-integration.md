# Gateway / App Server 로컬 develop 통합

## Claim
사용자가 승인한 게이트웨이 및 App Server 작업 트리 4개의 모든 미커밋 변경을 각 브랜치에 커밋하고 로컬 develop에 병합했다. 원격 push 및 작업 트리 제거는 하지 않았다. App Server 전체 운영 완료나 카드 완료를 주장하지 않는다.

## Baseline-attribution
- 대상: /Users/goos/MoAI/moai-adk-go/.claude/worktrees/develop
- 시작 develop: d9eac8264b39ceec18032db379e6f65098851d27
- 네 브랜치 병합 후: 89c7c0ef09bb1cc32740c15345825badb6cca34f
- gateway 소스: ce79ef7caf758a957fe77c279c633bc00c95905b (t649)
- App Server core 증거: d73b24b393914597a3669e5648c201b85b623e44 (t650)
- App Server tools 증거: 1a8e3aa30d806ef24a66a1ab3d063e29fb5c2053 (t651)
- App Server turn bridge: 530bd73305e9cfb44d531d13fe1d468d394c4fd1 (t652)

## Evidence
`git merge-base --is-ancestor <각 위 source SHA> develop`: 네 번 모두 exit 0.
`git status --porcelain -uall`: 네 source 작업 트리 모두 빈 출력. develop에는 시작 시부터 있던 미추적 보고서 7개가 남아 있으며 이번 커밋에 포함하지 않았다.

병합 트리에서 실행:
`go test -p 1 ./internal/codexapp ./internal/codexbridge ./internal/gateway/... -count=1 -timeout=120s`
```text
ok  github.com/modu-ai/moai-adk/internal/codexapp 0.469s
ok  github.com/modu-ai/moai-adk/internal/codexbridge 12.198s
ok  github.com/modu-ai/moai-adk/internal/gateway 8.634s
ok  github.com/modu-ai/moai-adk/internal/gateway/auth 3.115s
ok  github.com/modu-ai/moai-adk/internal/gateway/conversation 1.907s
ok  github.com/modu-ai/moai-adk/internal/gateway/opaque 0.414s
ok  github.com/modu-ai/moai-adk/internal/gateway/receipt 1.273s
ok  github.com/modu-ai/moai-adk/internal/gateway/translate 0.731s
```
`go test -p 1 ./internal/cli -run 'TestGateway|TestGPT|TestDefaultLaunch|TestCG|TestLaunchClaude|TestGLM' -count=1 -timeout=120s`
```text
ok  github.com/modu-ai/moai-adk/internal/cli 19.557s
```
공통 환경변수 상수로 통합 수리 후:
`go test -p 1 ./internal/config ./internal/cli -run 'TestNoBareAnthropicEnvVarLiteralsInProduction|TestGateway|TestGPT|TestDefaultLaunch' -count=1 -timeout=120s`
```text
ok  github.com/modu-ai/moai-adk/internal/config 1.004s
ok  github.com/modu-ai/moai-adk/internal/cli 3.756s
```
`go test -p 1 ./internal/kanban -run 'TestWriteThenReadRoundTripsEveryField|TestWrittenJSONCarriesTheDocumentedKeys|TestNewRecordStampsAnRFC3339EnteredAt|TestRecordWithoutSpecIDRoundTrips' -count=1 -timeout=30s`
```text
ok  github.com/modu-ai/moai-adk/internal/kanban 0.489s
```
템플릿 해시 및 CG 안내 문구 동기화 후:
`go test -p 1 ./internal/template -run 'TestOutputStylesTemplateLiveParity|TestCatalogHashCoversSkillSubfiles|TestManifestHashFormat' -count=1 -timeout=60s`
```text
ok  github.com/modu-ai/moai-adk/internal/template 0.522s
```
`go build -p 1 -o /tmp/moai-develop-integration-check ./cmd/moai`: 출력 없음, exit 0 (89c7c0ef0의 코드, 카탈로그 해시 갱신 전).

## 통합 수리
- AGENTS 템플릿: develop worktree 명령 표를 유지하고 cc/glm/gpt 및 CG migration을 반영.
- catalog: [1m] 경로와 App Server 인증 방식을 함께 유지.
- router/server: App Server 관리 인증을 연결.
- translate: 최신 is_error=true 처리 및 검증을 유지.
- CLI: develop의 공통 환경변수 상수 규칙에 맞게 기존 상수 사용.
- moai-learn: 배포 템플릿과 live 원본의 CG 제거 문구를 일치시킴.
- catalog.yaml: 공식 생성기로 변경된 항목 6개의 해시를 재계산.

## Gaps
추가로 실행한 `go test -p 1 ./internal/config ./internal/homestate ./internal/kanban ./internal/template/... -count=1 -timeout=120s`는 최초 환경변수 상수/템플릿 일치 검사 실패 및 kanban 패키지 시간 초과로 exit 1이었다. 환경변수·템플릿 실패는 위 대상 재검사로 통과했다. homestate는 12.417s에 통과했다. kanban 전체는 통과하지 않았다:
```text
panic: test timed out after 2m0s
running tests:
    TestForemanQueueWatch_DBOnlyTargetMissesWALDeferral (14s)
FAIL github.com/modu-ai/moai-adk/internal/kanban 120.586s
```
시간 초과가 이번 병합으로 생긴 회귀인지는 입증하지 않았다. 전체 저장소/Windows CI/실제 App Server 모든 사용자 여정은 검증하지 않았다. 기존 로컬 설치 바이너리는 이전 rc.8이며 이번 통합 develop 빌드로 재배포하지 않았다.

## Residual-risk
로컬 통합은 완료했지만 CI 및 App Server 후속 운영 검증은 별도다. 과거 raw 증거 로그의 공백은 보존했다. source 트리는 통합되었으나 제거하지 않는다. primary main 체크아웃의 다른 작업은 변경하지 않는다.
