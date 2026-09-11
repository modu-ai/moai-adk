# t647: Todo 소비 화면 및 문서 개선

## Claim

SURF-01~05를 구현했다. 늦게 생성된 프로젝트 로컬 및 홈 대기열 디렉터리를 감시 대상으로 다시 등록하고 화면 갱신 신호를 보낸다. 읽을 수 없는 대기열은 `/todo`와 Overview에서 0건이나 빈 대기열로 표시하지 않는다. 오류 안내는 영어·한국어·일본어·중국어를 제공하며 원시 오류 문자열을 전달하지 않는다. 4개 언어 문서의 SQLite 저장 설명과 Class별 절차를 수정하고 설치된 todo 워크플로와 배포 템플릿의 경로를 일치시켰다.

## Evidence

수정 전 RED 명령:

```text
go test ./internal/web -run 'TestTodoUnreadableDoesNotClaimEmpty|TestTodoWatcherRegistersLateDirectories' -count=1 -v
```

관측된 실패 부분:

```text
    todo_recovery_test.go:21: /todo lacks unreadable diagnostic
    todo_recovery_test.go:22: /todo claims empty/counts for unreadable queue
    todo_recovery_test.go:21: / lacks unreadable diagnostic
    todo_recovery_test.go:22: / claims empty/counts for unreadable queue
--- FAIL: TestTodoUnreadableDoesNotClaimEmpty (0.14s)
    todo_recovery_test.go:49: late todo directory never refreshes
    todo_recovery_test.go:49: late todo directory never refreshes
--- FAIL: TestTodoWatcherRegistersLateDirectories (9.30s)
    --- FAIL: TestTodoWatcherRegistersLateDirectories/project-local (4.60s)
    --- FAIL: TestTodoWatcherRegistersLateDirectories/home (4.69s)
FAIL
FAIL    github.com/modu-ai/moai-adk/internal/web    10.107s
FAIL
```

마지막 GREEN 명령과 출력:

```text
go test ./internal/web -run 'TestTodo|TestConsoleRoutesLeaveBacklogUntouched|TestBacklogReadSeam|TestResolvedWatchPaths|TestOverviewPrimarySurface' -count=1
ok      github.com/modu-ai/moai-adk/internal/web    6.785s
```

같은 선택 범위의 계측 명령:

```text
go test ./internal/web -run 'TestTodo|TestConsoleRoutesLeaveBacklogUntouched|TestBacklogReadSeam|TestResolvedWatchPaths|TestOverviewPrimarySurface' -count=1 -coverprofile=/tmp/todo-surface-audit.IYzgx9/coverage.out
ok      github.com/modu-ai/moai-adk/internal/web    6.791s    coverage: 43.2% of statements
```

`go tool cover -func`의 주요 함수 계측값:

```text
Watch                   85.3%
resolvedWatchPaths      100.0%
eventFor                100.0%
readTodoQueue           92.9%
buildTodo               100.0%
todoStateCount          100.0%
boundedTodoItems        75.0%
```

43.2%는 선택한 테스트가 전체 web 패키지 문장 중 실행한 비율이며, 전체 테스트 실행의 커버리지가 아니다. 기존 디렉터리의 실제 파일 이벤트를 먼저 확인한 후 새 디렉터리를 생성하므로 감시가 시작되지 않은 상태를 결함 증거로 삼지 않는다. 테스트의 goroutine은 stop 채널과 종료 대기로 정리한다.

추가 확인: `templ generate -f screens.templ`을 `internal/web`에서 실행해 생성 파일을 갱신했다. `node --check internal/web/assets/i18n.js`, `git diff --check`, 설치·배포 todo 워크플로의 `cmp`는 모두 출력 없이 성공했다.

## Baseline-attribution

- 감사 기준 커밋: `ee99507fbe3b4a22c6a0a74815723d222dfdc04d`
- 구현 트리: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/todo-audit`
- 브랜치: `WT-todo-audit`
- 계측일: 2026-09-11
- TDD 지침을 읽고 RED → GREEN 순서로 진행했다. 문서 claim-check 스킬은 읽기 전용 평가용이므로 구현 작업에는 적용하지 않았다.

## Gaps

브라우저 전체 여정과 전체 web 테스트 묶음은 이 분담에서 실행하지 않았다. 실제 디렉터리 감시와 HTTP 렌더링을 테스트했다. UI 오류가 화면에서 구분되는지는 DOM 문자열로 검증했으며 스크린샷을 근거로 삼지 않았다. 기타 카드 상태·저장 계층의 정확성은 별도 분담에서 검증한다.

## Residual-risk

감시 복구는 1초 간격으로 등록되지 않은 경로를 확인한다. 운영체제의 알림 큐 자체가 넘치는 경우까지 재현하지 않았다. 생성된 templ 코드의 변경량에는 조건문 추가에 따른 지역 변수 번호와 원본 줄 번호 변경이 포함된다.

## 주변 문서·소비 경계의 실제 판독 범위

보고서 후보 217개 전체를 이 담당자가 모두 읽었다는 뜻은 아니다. 다음은 `rg`로 todo/backlog 언급을 찾고 실제 관련 구간을 읽은 범위다. 언어 집합은 `en`, `ko`, `ja`, `zh`다.

| 범위 | 실제 판독 구간과 확인 내용 |
|---|---|
| `docs-site/content/{언어}/advanced/factory-mode.md` | 카드 Class 설명, 이미 picked인 카드의 배차, 운영자 선택 문장(31·52~57행 부근). 자동 pick을 허용한다고 해석하지 않았다. |
| `docs-site/content/{언어}/advanced/kanban-mode.md` | todo/backlog 언급, 253~264행의 Class A/B/C 표와 Class B의 SPEC 생략. 보드 상태 저장소 설명은 검색해 발견했지만 runtime 전체 진위를 재판정하지 않았다. |
| `docs-site/content/{언어}/advanced/moai-web-console.md` | 88~114행 전체: todo 읽기 전용, 오류/빈 상태, SSE 감시 경로, 재연결. 아래 추가 drift를 수정했다. |
| `docs-site/content/{언어}/advanced/statusline.md` | backlog 토글과 200~203행 부근의 관측원 부재 시 숨김 설명. todo 저장 경로 리터럴은 없었다. |
| `docs-site/content/{언어}/advanced/config-sections.md` | 201~218행 todo.enabled 안내 억제·명시적 호출 유지·기본 켜짐 계약. |
| `kanban-dispatch.md` | 25~55행의 producer/promotion/PR 확인/Class 규칙. 31·33행의 무조건 plan 진입 문구와 47행의 Class별 예외가 충돌하여 해당 두 문장만 고쳤다. embedded counterpart에도 같은 두 문장만 반영했다. |
| `kanban-dispatch-detail.md` | 카드·컬럼 정의와 Class 표, 255~275행 부근의 검증 부하 사건 및 pre-dispatch PR 확인 설명. |
| `main-checkout-branch-guard-detail.md` | 110~126행: todo add의 인용된 설명 안에 있는 git 명령을 실제 실행 명령으로 오인하지 않는 범위. |
| `cadence-bridge.md` | 82~99행: 발견 기록은 TaskList 또는 reports/cadence에 남기며 todo 카드를 자동 발행하는 권한으로 해석하지 않는 경계. |
| `core/verification-claim-integrity.md` | 114~120행: 과거 카드가 등장하는 baseline ordering 사례. todo 런타임 기능 설명이 아니었다. |
| `core/moai-constitution-detail.md` | 50~54행: lessons inbox backlog. todo 카드 저장소와 별개인 기록이다. |

위 6개 관련 규칙 외 `workflow/mx-tag-protocol.md`도 검색에 걸렸다. 124~140·150~159행을 확인한 결과 `todo_per_file`은 MX 태그 제한이므로 todo 큐 소비 경계에서 제외했다.

코드 소비 경계의 판독:

- hook: `session_start_kanban.go` 179~212행의 enabled gate와 queued-only 요약, i18n 문구 및 disabled 테스트. 부팅 시 읽기 실패를 0으로 요약하는 기존 정책은 관측했지만 이번 UI 수정과 동일 결함으로 자동 확대하지 않았다.
- statusline: `backlog.go`, `builder.go` 254~282행, `renderer.go` 207~220행, SQLite reader 및 unavailable 테스트. 렌더 억제와 데이터 수집을 분리하는 명시적 계약을 확인했다.
- config/settings: `todo_enabled.go`와 관련 enabled 테스트, `schema_todo_test.go`의 nested bool 등록·저장 경로 검사.
- MCP: `internal/mcp/catalog.go`의 도구 목록과 catalog 테스트, CLI 등록 파일의 todo 참조 검색. 전용 todo MCP 도구는 이 목록에 없으며 그것 자체를 미구현 결함으로 판정하지 않았다.

## T16·T20 주변 문서 교정의 추가 근거

4개 언어 advanced web 문서에 unreadable도 empty라는 옛 문장과 프로젝트 로컬만 적힌 감시 표가 남아 있었다. 읽기 전용 Node probe의 수정 전 출력:

```text
en: unreadable_empty_claim=true; home_todo_watch_documented=false
ko: unreadable_empty_claim=true; home_todo_watch_documented=false
ja: unreadable_empty_claim=true; home_todo_watch_documented=false
zh: unreadable_empty_claim=true; home_todo_watch_documented=false
advanced_web_doc_drift_locales=4
```

현재 runtime 증거:

```text
go test ./internal/web -run 'TestTodoUnreadableDoesNotClaimEmpty|TestResolvedWatchPathsIncludeHomeTodoAndFactory' -count=1
ok      github.com/modu-ai/moai-adk/internal/web    2.537s
```

부모가 T16 문서 수정 범위를 승인한 뒤 4개 advanced web 페이지를 수정했다. HTTP 200은 유지하되 missing/valid empty와 unavailable을 구분하고, 홈 Todo 및 Factory 감시 경로와 늦게 생성된 디렉터리의 재등록을 설명했다.

T20 규칙의 수정 전 Node 계약 검사는 `unconditional_promotion_plan_clauses:2`, `class_b_skips_plan:true`를 출력하고 exit 1이었다. 부모 승인 뒤 설치·배포 규칙의 해당 두 문장을 Class A direct close, Class B run, Class C plan으로 수정했다. 기존 운영자 선택·발행·우선순위 권한은 바꾸지 않았다.

수정 후 Node assert 검사(옛 오류 문장 부재, 홈 Todo·Factory 감시 존재, HTTP 200 명시, 두 규칙의 Class별 진입 문구 각 2건)와 `git diff --check`가 성공했다.

```text
.claude/rules/moai/workflow/kanban-dispatch.md: class_entry_contract=PASS
internal/template/templates/.claude/rules/moai/workflow/kanban-dispatch.md: class_entry_contract=PASS
en: current_unreadable_and_watch_contract=PASS
ko: current_unreadable_and_watch_contract=PASS
ja: current_unreadable_and_watch_contract=PASS
zh: current_unreadable_and_watch_contract=PASS
```

추가 소비 경계 검증:

```text
go test ./internal/statusline ./internal/hook ./internal/config ./internal/mcp -run 'TestRendererBacklogSegmentGating|TestResolveBacklogCounts_UnreadableIsUnavailableNotZero|TestSessionStartKanbanRespectsTodoDisabled|TestTodoEnabled|TestMoaiMCP' -count=1
ok      github.com/modu-ai/moai-adk/internal/statusline    0.540s
ok      github.com/modu-ai/moai-adk/internal/hook    1.235s
ok      github.com/modu-ai/moai-adk/internal/config    1.351s
ok      github.com/modu-ai/moai-adk/internal/mcp    0.336s
```

이 추가 검토에서도 후보 문서 전체 내용의 무결함이나 런타임 전체의 완전성을 주장하지 않는다. 표에 적은 todo 관련 구간과 명시한 실행 범위만 근거다.

## T22: 읽기가 자체 갱신을 반복하는 문제

독립 동작 검토에서 실제 `Hub.Watch`와 `/todo` HTTP 읽기를 연결했다. seed 이후 카드 쓰기 없이 읽기 → kanban 이벤트 → 읽기를 4회 반복했고, 기준 develop `ee99507fb`와 수정 중인 트리 양쪽에서 동일한 반복을 관측했다. 따라서 이번 reader 변경으로 새로 생긴 회귀라고 판정하지 않았다. 대조군은 감시 준비를 기존 config 파일 이벤트로 확인한 뒤, 아무 작업 없는 600ms 동안 이벤트가 없음을 확인했다. 감시 goroutine과 원시 fsnotify observer는 종료 채널/Close와 join으로 정리했다.

수정 전 양쪽 트리에서 관측한 핵심 출력:

```text
idle control: no event
read 1 triggered kanban
read 2 triggered kanban
read 3 triggered kanban
read 4 triggered kanban
read-triggered kanban cycles=4/4
HTTP reads sustain todo refresh feedback without queue writes
```

명령은 각 트리에서 `go test -overlay /tmp/todo-surface-audit.IYzgx9/feedback-{current,baseline}.json ./internal/web -run TestPeerTodoReadWatcherFeedback -count=1 -v`이며, current 실행은 exit 1 / 3.889s, baseline 실행은 exit 1 / 4.290s였다. 원시 파일 이벤트를 별도 observer로 수집한 current 재실행에서는 다음을 관측했다.

```text
CREATE backlog.db-wal
CREATE backlog.db-shm
WRITE backlog.db-shm
CHMOD backlog.db-shm
REMOVE backlog.db-shm
REMOVE backlog.db-wal
```

쓰기 없는 HTTP 읽기가 SQLite의 임시 WAL/SHM 파일 생성·제거를 유발했다. 부모의 T22 수정 승인 뒤 `events.go`에서 `backlog.db-shm` 이벤트와 `backlog.db-wal`의 비쓰기 이벤트를 제외했다. WAL 쓰기는 유지하여 writer가 연결을 유지하고 메인 DB가 아직 checkpoint되지 않은 경우도 갱신된다. DB 및 JSON 파일 이벤트, 늦게 생성된 디렉터리 재등록은 그대로 유지했다.

영구 회귀 테스트의 수정 전 RED:

```text
go test ./internal/web -run 'TestTodoReadsDoNotRefreshThemselves|TestTodoCommitsStillRefresh' -count=1 -v
    todo_feedback_test.go:72: read 1 triggered kanban without any queue write
--- FAIL: TestTodoReadsDoNotRefreshThemselves (1.64s)
--- PASS: TestTodoCommitsStillRefresh (1.90s)
    --- PASS: TestTodoCommitsStillRefresh/closed_writer (0.95s)
    --- PASS: TestTodoCommitsStillRefresh/held_WAL_writer (0.94s)
FAIL
FAIL    github.com/modu-ai/moai-adk/internal/web    4.221s
FAIL
```

수정 후 GREEN:

```text
go test ./internal/web -run 'TestTodoReadsDoNotRefreshThemselves|TestTodoCommitsStillRefresh|TestTodoWatcherRegistersLateDirectories' -count=1 -v
--- PASS: TestTodoReadsDoNotRefreshThemselves (3.41s)
--- PASS: TestTodoCommitsStillRefresh (1.90s)
    --- PASS: TestTodoCommitsStillRefresh/closed_writer (0.94s)
    --- PASS: TestTodoCommitsStillRefresh/held_WAL_writer (0.95s)
--- PASS: TestTodoWatcherRegistersLateDirectories (2.32s)
    --- PASS: TestTodoWatcherRegistersLateDirectories/project-local (1.12s)
    --- PASS: TestTodoWatcherRegistersLateDirectories/home (1.20s)
PASS
ok      github.com/modu-ai/moai-adk/internal/web    8.304s
```

원래 overlay 진단도 수정 후 재실행했다. 원시 CREATE/REMOVE/SHM WRITE는 계속 관측되었으나 Hub 갱신은 사라졌다.

```text
go test -overlay /tmp/todo-surface-audit.IYzgx9/feedback-current.json ./internal/web -run TestPeerTodoReadWatcherFeedback -count=1 -v
idle control: no event
read 1 produced no event
read 2 produced no event
read 3 produced no event
read 4 produced no event
read-triggered kanban cycles=0/4
--- PASS: TestPeerTodoReadWatcherFeedback (6.39s)
PASS
ok      github.com/modu-ai/moai-adk/internal/web    7.042s
```

관측은 현재 macOS 파일 알림 동작에 귀속한다. Linux·Windows 파일 이벤트 조합 및 브라우저 EventSource 전체 여정은 이 테스트에서 실행하지 않았다. 4회 반복을 기계적으로 확인했으며 무한 시간을 실제로 기다렸다고 주장하지 않는다.
