# t657 MoAI Web UI/UX 개선 검증 판정

## Claim

- 기본 정보 구조를 `Overview`, `Todo`, `Settings` 3개 주 표면으로 재구성하였다.
- `Kanban`, `Specs`, `Monitor`는 주 메뉴와 Overview에서 노출하지 않고, 기존 route·handler·상세 화면은 보존하였다.
- Settings는 현행 14개 탭을 모두 유지하면서 각 탭에 설명, 적용 시점, 직접 진입 링크를 포함하는 상세 헤더를 추가하였다.
- Overview에는 작업 건강도, Todo 요약, 빠른 설정 진입, 현재 범위 정보를 배치하였다.
- console UI의 색상 토큰은 docs-site frozen token 체계에 맞추고, 반응형 간격·컨트롤 높이·키보드 focus·텍스트 줄바꿈 규칙을 보강하였다.

이 판정은 작업 브랜치의 구현과 검증을 대상으로 하며, 이 문서를 작성하는 시점에는 `develop` 병합 및 원격 push 완료를 주장하지 않는다.

## Evidence

### 카드와 작업 기준

실행 명령:

~~~
$ moai todo list --limit 0 | rg 't657|MoAI Web UI/UX'
todo: answered by the SQLite backlog store; the backlog.json beside it is NOT the queue — an export or a legacy leftover, whose contents can be arbitrarily stale
t657	queued	MoAI Web UI/UX 전면 개선: Kanban·Specs·Monitor는 1차 UX에서 숨기고 route·handler는 보존, Overview·Todo·Settings 3면 IA로 재구성, 현행 Settings 14개 탭을 상세 페이지로 구현, docs-site frozen token 유지, 검증 후 WT에서 develop 병합·push
~~~

### 템플릿 생성

실행 명령:

~~~
$ go run github.com/a-h/templ/cmd/templ generate -path ./internal/web
(✓) Post-generation event received, processing... [ updates=1 needsRestart=true needsBrowserReload=true ]
(✓) Complete [ updates=1 duration=66.223583ms ]
~~~

생성된 `root_templ.go`, `screens_templ.go`, `shell_templ.go`는 템플릿 변경으로부터 생성된 결과물이다.

### 자동 검증

핵심 UI 회귀 검증:

~~~
$ GOCACHE=/tmp/moai-web-t657-gocache go test ./internal/web -run 'Test(GoormSansCodeSelfHosted|DocsSiteColourSystem|PrimaryRailKeepsThreeSurfaceIA|SettingsPageHeaderKeepsAllFourteenTabs|OverviewPrimarySurfaceWiring|TodoNavRowIsSecondAndCurrent|ConsoleTabsOrder|TabPanelRenderOrderMatchesTabs)' -count=1
ok  	github.com/modu-ai/moai-adk/internal/web	0.983s
~~~

변경 범위를 포함한 나머지 `internal/web` 전체 검증:

~~~
$ GOCACHE=/tmp/moai-web-t657-gocache go test ./internal/web -count=1 -skip TestShutdown_RouteStopsServer > /tmp/moai-web-t657-full.log 2>&1
$ tail -n 6 /tmp/moai-web-t657-full.log
ok  	github.com/modu-ai/moai-adk/internal/web	16.658s
~~~

분리한 기존 graceful-shutdown 검사:

~~~
$ GOCACHE=/tmp/moai-web-t657-gocache go test ./internal/web -run TestShutdown_RouteStopsServer -count=1 -v
=== RUN   TestShutdown_RouteStopsServer
--- PASS: TestShutdown_RouteStopsServer (0.01s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/web	0.623s
~~~

핵심 view-model 및 탭 레이아웃 race 검증:

~~~
$ GOCACHE=/tmp/moai-web-t657-gocache go test ./internal/web -race -run 'Test(PrimaryRailKeepsThreeSurfaceIA|SettingsPageHeaderKeepsAllFourteenTabs|OverviewPrimarySurfaceWiring|ConsoleTabsOrder|TabPanelRenderOrderMatchesTabs)' -count=1
ok  	github.com/modu-ai/moai-adk/internal/web	2.323s
~~~

정적 분석:

~~~
$ GOCACHE=/tmp/moai-web-t657-gocache go vet ./internal/web
# 표준 출력 없음, exit code 0
~~~

공백·patch 오류 검사:

~~~
$ git -C /Users/goos/MoAI/moai-adk-go/.claude/worktrees/t657 diff --check
# 표준 출력 없음, exit code 0
~~~

### 브라우저 관측

작업 브랜치 소스로 `http://127.0.0.1:3047`를 기동하여 확인하였다.

- `/`의 주 rail에는 `Overview`, `Todo`, `Settings`만 표시되었다.
- Overview에는 `WORK HEALTH`, `TODO SNAPSHOT`, `QUICK ACTIONS`, `CURRENT SCOPE` 블록이 표시되었다.
- `/settings?tab=gate`에는 `Identity`, `Language`, `LLM`, `GLM Settings`, `Workflow`, `Git & Worktree`, `Audit`, `Codex`, `Agents`, `Report`, `MCP`, `Cross-Session`, `Feedback`, `Quality Gate` 14개 탭이 표시되었다.
- 같은 Settings 화면의 상세 헤더에는 `SETTINGS DETAIL`, `Quality Gate`, `Commit-time quality gate posture.`, `Pages 14`, `Per request`, `Direct link`가 표시되었다.
- Overview와 주 rail에는 `Kanban`, `Specs`, `Monitor` 링크가 표시되지 않았다.

## Baseline-attribution

모든 구현·검증은 다음 worktree와 기준점에서 수행하였다.

~~~
worktree: /Users/goos/MoAI/moai-adk-go/.claude/worktrees/t657
$ git rev-parse --short HEAD
d060e0d13
$ git branch --show-current
WT-moai-web-settings-ux
$ git rev-list --count --left-right origin/develop...HEAD
0	0
~~~

즉, 작업 브랜치는 확인 시점의 `origin/develop`와 동일한 `d060e0d13`에서 출발하였다.

## Gaps

- 기본 sandbox 권한으로 `go test ./internal/web -count=1`을 실행했을 때 기존 `TestShutdown_RouteStopsServer`가 listener bind timeout으로 실패하였다. listen 권한을 허용한 분리 재검증에서는 위와 같이 `PASS`였으며, 나머지 전체 검증도 `-skip TestShutdown_RouteStopsServer` 조건에서 `ok`였다.
- 저장소에 `/Users/goos/go/bin/moai e2e` 명령이 등록되어 있지 않아 자동 e2e 명령의 결과는 제시하지 않는다. 브라우저 확인은 로컬 수동 관측이다.
- 외부 브라우저 엔진별 시각 회귀, pixel diff, CI 전체 suite는 이 worktree에서 측정하지 않았다.

## Residual-risk

- 변경 후 원격 CI에서 전체 repository suite와 배포 환경의 실제 asset serving을 추가로 확인해야 한다.
- 브라우저 관측은 한 로컬 viewport와 기본 설정 상태에 한정되므로, 모바일·장문 번역·실제 사용자 설정값 조합에서의 시각적 줄바꿈은 잔여 위험으로 남는다.
- Kanban·Specs·Monitor의 호환 route는 코드상 보존했으나, 주 IA에서 숨긴 뒤의 직접 URL 접근과 기존 북마크 흐름은 원격 환경에서 별도 smoke 검증이 필요하다.
