# t422 verdict — todoFixture fail-loud 가드

카드: t422 (Tier S) · 브랜치: WT-todo-fail-loud (base: origin/develop 9145806d8)
운영자 배차 승인 2026-09-01 · lane 실행 2026-09-02

## Claim

`runTodo` / `runTodoWithClosedStdin` 테스트 헬퍼가 `todoFixture(t)` 없이 실행되어 큐 루트가
실제 primary 체크아웃으로 풀린 상태면, Execute 전에 즉시 `t.Fatalf`로 실패한다 (에러 메시지에
`todoFixture` 사용 안내 포함). 프로덕션 CLI 실행 경로는 무변경.

## Evidence

### 1. 가드 구현 (표면: 테스트 헬퍼 2개 — 프로덕션 파일 미수정)

- `internal/cli/todo_queue_root_test.go` — `liveTodoQueueRootReason()` (해석된 큐 루트가
  `os.TempDir()` 아래(양쪽 심링크 정규화, `/var/folders`↔`/private/var/folders`)면 `""`,
  아니면 todoFixture 안내 문장), `queueRootInsideTemp()`, `underDir()` (Rel 기반 — 형제
  접두사 오탐 면역) + 관측 테스트 3개
- `internal/cli/todo_test.go:35` — `runTodo` 게이트 (`t.Fatalf`, Execute 전)
- `internal/cli/todo_relate_test.go:315` — `runTodoWithClosedStdin` 게이트 (동일)
- 가드 발동 메시지 예: `todo queue isolation guard: queue root "/Users/goos/MoAI/moai-adk-go"
  is the live repository, not an isolated fixture — running todo commands now would read or
  mutate the operator's real backlog (the t394 incident). Call todoFixture(t) before any todo
  command: it points CLAUDE_PROJECT_DIR at a committed temp repo so the queue resolves there.`

### 2. 관측 테스트 3/3 PASS

```
=== RUN   TestTodoQueueRootGuard_FiresOnLiveRepository   (cwd 폴백 → 실제 리포 → 가드 발동 + todoFixture 명명)
--- PASS: TestTodoQueueRootGuard_FiresOnLiveRepository (0.03s)
=== RUN   TestTodoQueueRootGuard_SilentOnFixture          (todoFixture → 침묵)
--- PASS: TestTodoQueueRootGuard_SilentOnFixture (0.09s)
=== RUN   TestTodoQueueRootGuard_SilentOnHomeFallbackFixture (userHomeDirFn 폴백 격리 → 침묵)
--- PASS: TestTodoQueueRootGuard_SilentOnHomeFallbackFixture (0.04s)
```

### 3. 가드 채택 즉시 실화 — 미격리 잔존 테스트 3건 발견·수리

가드 활성 상태로 기존 스위트(`-run Todo`)를 돌리자 **가드가 실제 미격리 테스트 3건을 잡았다**
(전부 `t.TempDir()` + `t.Setenv("CLAUDE_PROJECT_DIR", …)`만 하고 git repo로 만들지 않은
모양 — git 풀이 실패 → 실제 home 폴백 큐로 풀림):

| 테스트 | 위험도 | 수리 |
|---|---|---|
| `TestTodoBareInvocationLists` | **높음** — 매 실행마다 폴백 큐에 실제 `add` 쓰기 | `todoFixture(t)` |
| `TestTodoSingleWordNaturalLanguageStillErrors` | 낮음 — 오류 거절만 (키 `001-3f74353b`의 base "001" = 자기 temp dir basename → 자기 키로 실제 home에 폴백 큐 생성) | `todoFixture(t)` |
| `TestTodoUnknownSubcommandStillErrors` | 낮음 — 오류 거절만 | `todoFixture(t)` |

수리 후 `-run Todo` 전체 GREEN (`ok github.com/modu-ai/moai-adk/internal/cli 66.361s`).

### 4. 양방향 뮤턴트 관측 (임시 테스트, 커밋·오염 없음 — 뮤턴트 경로는 list 읽기 전용)

- **(a) 가드 활성 + fixture 없음**: `-run TestT422MutantObservation` → **FAIL 0.03s** —
  `queue root "/Users/goos/MoAI/moai-adk-go" is the live repository …` (Execute 전 차단;
  워크트리에서 실행했음에도 t106 불변대로 실제 primary로 풀리는 것도 동시 관측)
- **(b) 가드 제거 뮤턴트** (runTodo 게이트 임시 주석): 동일 테스트 → **PASS, `err=<nil>`** —
  조용히 실행되며 출력에 운영자 라이브 큐 실카드 t10의 고유 텍스트 `Vet cross-platform`
  존재 (`grep -c` = 1) → fixture 없는 runTodo가 실제 라이브 큐로 흐름 확정 (add도 같은
  `todoBacklogPath(resolveTodoQueueRoot())` 해석 공유 → t394 재발 경로 그대로)
- **복원**: 가드 주석 해제 후 동일 테스트 다시 FAIL — 양방향 확인. 임시 테스트 파일·캡처 파일
  삭제 완료.

### 5. 프로덕션 경로 무변경

- 변경 파일은 `*_test.go` 3개뿐 — `todo.go` 등 프로덕션 파일 diff 0 (아래 커밋 참조)
- `go vet ./internal/cli/` → rc=0

## Baseline-attribution

- 모든 측정은 워크트리 `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t422`에서,
  브랜치 `WT-todo-fail-loud` @ `9145806d8`(= origin/develop) 트리에 대해 이 run에서 실행.
- 명령: `go vet ./internal/cli/`, `go test ./internal/cli/ -run 'Todo' -count=1`,
  `go test ./internal/cli/ -run 'TestTodoQueueRootGuard' -v -count=1`,
  뮤턴트 관측 `go test ./internal/cli/ -run 'TestT422MutantObservation' -v -count=1` (가드 활성/제거/복원 3회).

## Gaps

- 전체 스위트 판정 (완료 후 기록): `go test ./internal/cli/ ./internal/kanban/ -count=1`
  → `ok … internal/cli 364.742s` + `ok … internal/kanban 17.484s`, exit 0 — 가드 활성
  상태에서 오탐 0.
- 가드 표면은 `runTodo` / `runTodoWithClosedStdin` 2개 헬퍼. 새 테스트 헬퍼가 `newTodoCmd()`
  직접 Execute로 우회하면 가드가 없다 (기존 직접 사용처 4곳은 확인: 2곳은 env-격리된 자식
  프로세스 헬퍼, 2곳은 Execute 없는 커맨드 트리 검사 — 우회 없음).

## Residual-risk

- dev 에이전트가 `todoFixture`도 `runTodo`도 거치지 않고 `kanban.NewBacklogStore`를
  primary 경로로 직접 만들면 가드 밖이다 — 그런 흐름은 큐 접근 경로 자체가 달라 본 카드의
  표면(runTodo 경유)과 다르며, 관측된 사고 경로(t394)는 runTodo 경유였다.
- `queueRootInsideTemp`의 temp-tree 판정은 OS temp 루트 의존 — 테스트 환경이 `TMPDIR`을
  비-temp 경로로 덮어쓰는 극단적 설정이면 오탐(fail-loud 방향) 가능.
