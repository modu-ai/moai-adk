# t549 판정 — 루트를 받는 함수에 상태 디렉터리가 넘어가는 계층 어긋남

- card: t549 (Class B, SPEC 없음)
- lane: lane-10
- worktree: `.claude/worktrees/t549` · branch `WT-root-rederive`
- base: 로컬 develop `d3b7d438d`
- 판정: **결함** — 의도가 아니다. 수리 적용.

## 1. 주장 (Claim)

1. `homeTodoQueueRoot` 의 홈 미해소 분기가 ROOT 자리에 상태 디렉터리(`<base>/.moai/state/todo`)를 돌려주고, 소비자는 받은 값을 root 로 보고 `BacklogPathForRoot` 로 상태 디렉터리를 한 번 더 붙인다. 결과는 오류 없는 빈 큐다.
2. 이 값은 의도가 아니라 `ed021015f` 이관 때 층이 바뀌면서 남은 잔재다.
3. 기본 temp 루트 집합(스텁 없음)에서도 도달 가능하다.
4. 수리(`return base, false`) 후 두 리졸버 모두 기존 큐를 읽는다.

## 2. 증거 (Evidence)

### 2.1 수리 전 재현 — develop `d3b7d438d`

명령: `go test ./internal/kanban/ -run 'TestT549_' -count=1 -v` → `repro-develop-d3b7d438d.txt` (exit 1)

```
--- PASS: TestT549_BacklogPathForRootTreatsArgumentAsRoot
    BacklogPathForRoot(stateDir) = …/001/.moai/state/todo/.moai/state/todo/backlog.json
    pure: root=…/001/.moai/state/todo path=…/001/.moai/state/todo/.moai/state/todo/backlog.json items=0 err=<nil>
    pure resolver read 0 items through its root (err <nil>), want 3
    adopting: root=…/001/.moai/state/todo path=…/.moai/state/todo/.moai/state/todo/backlog.json items=0 err=<nil>
    adopting resolver read 0 items through its root (err <nil>), want 3
--- FAIL: TestT549_NoHomeFallbackReadsTheExistingQueue
```

대조군: 같은 픽스처의 canonical 경로(`BacklogPathForRoot(base)`) 읽기가 3건을 돌려줘야 본문 단언에 도달하도록 짰고, 도달했다(Fatalf 없음). 전제 5종(비 git · MOAI_HOME 비절대 · temp 판별 거짓 · `os.TempDir()` 내부 · 홈 미해소)도 모두 단언으로 확인했다. 따라서 0건은 읽을 수 없는 픽스처가 아니라 틀린 경로에서 나왔다.

### 2.2 도달성 탐침 — develop `d3b7d438d`

명령: `go test ./internal/kanban/ -run 'TestT549_Probe' -count=1 -v` → `probe-develop-d3b7d438d.txt` (exit 1)

```
pathInsideTempDir=true isTemp=false (reason "") git=false root=…/001/link/.moai/state/todo
--- FAIL: TestT549_ProbeSymlinkedTempBaseReachesNoHomeBranch
```

`TempRootsFn` 은 스텁하지 않았다(홈만 스텁). `$TMPDIR` 안의 심볼릭 링크가 비 temp·비 git 디렉터리(`/usr`)를 가리키면 `pathInsideTempDir` 는 참(어휘 비교), `TempOriginReason` 은 거짓(EvalSymlinks 후 비교)이 되어 fallback 분기에 들어간다. 여기에 HOME 미해소가 겹치면 결함 값이 반환된다.

### 2.3 경위 — 의도냐 결함이냐

| 시점 | no-home 반환 | 소비자가 root 에 붙이는 것 | 층 |
|---|---|---|---|
| `ed021015f^` (internal/cli/todo.go) | `Join(base, ".moai","state","kanban")` | `Join(root, "backlog.json")` | 일치 — root 가 상태 디렉터리 층 |
| `ed021015f` (2026-08-25, kanban 이관) | 같은 값 유지 | `BacklogPathForRoot(root)` = `Join(root, ".moai","state","kanban","backlog.json")` | **어긋남 시작** |
| `8910c337c` (2026-08-27, 디렉터리 개명) | `resolveStateDir(base, false)` | `BacklogPathForRoot` → `resolveStateDir(root)` | 어긋남 존속 |
| `db45d8209` (2026-09-08, t536) | 같은 값 | 같음 | 주석이 "이 파일의 유일한 층 어긋남 값"으로 명시, 수리는 안 함 |

근거 명령:
- `git grep -n -B2 -A6 -E '"\.moai", "todo"' ed021015f^ -- internal/cli` → 원형의 `Join(root, "backlog.json")`
- `git grep -n -E 'backlog\.json"|BacklogPathForRoot\(' ed021015f -- internal/cli/todo.go internal/web/todo_queue_read.go` → 이관 시점 소비자가 `BacklogPathForRoot(root)`
- `git log -S 'resolveStateDir(base, false)'` → `8910c337c`, `db45d8209`

결론: 원형에서는 root 가 "backlog.json 이 들어 있는 디렉터리"였고 그 층에서는 이 값이 맞았다. 이관이 root 의 뜻을 "프로젝트 루트"로 옮기면서 홈 분기(`~/.moai/todo/<key>`)와 소비자는 새 층으로 갔지만 no-home 값만 옛 층에 남았다. 함수 자신의 주석도 "in-project root 를 돌려준다"고 적고 있었다. 그래서 수리는 반환값을 base 로 되돌리는 한 줄로 충분하다.

### 2.4 수리 후 — 같은 브랜치, 미커밋 작업 트리

| 명령 | 결과 | 파일 |
|---|---|---|
| `go test ./internal/kanban/ -run 'TestT549_' -count=1 -v` | exit 0, `--- PASS` 3개 (이름: `…TreatsArgumentAsRoot`, `…NoHomeFallbackReadsTheExistingQueue`, `…SymlinkedTempBaseWithDefaultRootsReturnsBase`) | `t549-tests-postfix.txt` |
| `go test ./internal/kanban/ -count=1` | exit 0, `ok … 155.211s` | `kanban-postfix.txt` |
| `go test ./internal/web/ -count=1` | exit 0, `ok … 14.177s` | `web-postfix.txt` |
| `gofmt -l internal/kanban/` · `go vet ./internal/kanban/` | 출력 없음, exit 0 / exit 0 | `vet-kanban.txt` |
| `go test ./internal/cli/ -run 'Todo' -count=1 -v -timeout 30m` (15:36:07–15:39:25, 리드 슬롯 승인 1회) | exit 0, `ok … 188.352s` · 최상위 `--- PASS` 161 / `--- FAIL` 0 / `--- SKIP` 1 | `cli-todo-postfix.txt`, `cli-todo-postfix.meta` |

셀렉터 대조: 정적 목록(`cli-todo-selector-static.txt`)은 165개였지만 그중 3개(`TestSaveBoolAnswerTodoEnabled`, `TestTodoEnabledQuestion`, `TestTodoEnabledTranslationsExist`)는 하위 패키지 `internal/cli/wizard/todo_enabled_test.go` 소속이다. git pathspec 의 `*` 가 디렉터리를 넘어 매칭해 과다 집계됐다. `./internal/cli/` 패키지 기준 기대치는 162이고, 실행 이름 집합과 `comm` 대조 결과 양방향 차이 0이다. SKIP 1개는 `TestAxisACanaryHomeSweep_TodoFamily` — `MOAI_AXIS_A_CANARY_SWEEP=1` 을 켜야 자식 `go test` 를 띄우는 opt-in canary 라 의도적으로 돌리지 않았다.

## 3. 스윕 — 「루트를 받는 함수가 그것을 무엇으로 취급하는가」

production Go(`*_test.go` 제외)의 해당 함수 전수. 명령: `git grep -n -E 'resolveStateDir\(|BacklogPathForRoot|homeTodoQueueRoot|fallbackTodoQueueRoot|ResolveTodoQueueRoot|StateDirForRoot\(|RuntimeStateDirForRoot\(' -- '*.go' ':!*_test.go'`

| 함수 | 받는 값 | 돌려주는 값 | 판정 |
|---|---|---|---|
| `StateDirForRoot` · `projectStateDirForRoot` · `LegacyStateDirForRoot` · `RuntimeStateDirForRoot` | root | 상태 디렉터리 | 일치 |
| `resolveStateDir` | root | 상태 디렉터리 | 일치 |
| `BacklogPathForRoot` · `BacklogPathForRootAdopting` | root | 큐 파일 | 일치 — 인자가 root 라는 계약을 테스트로 고정 |
| `BacklogCountsForRoot` · `QueuedBacklogCountForRoot` | root | 집계 | 일치 |
| `tempOriginSubstituteRoot` | base | base(root) | 일치 |
| `homeTodoQueueRoot` (ok) | base | `~/.moai/todo/<key>` (root 로 소비됨) | 층은 일치, 단 §5 Gap 1 |
| `homeTodoQueueRoot` (no-home) | base | ~~상태 디렉터리~~ → base | **결함 → 수리** |
| `fallbackTodoQueueRoot` · `ResolveTodoQueueRoot` · `ResolveTodoQueueRootAdopting` | base | root (no-home 값을 그대로 전달) | 위 수리로 함께 해소 |
| `adoptLocalTodoQueue` | base, fallbackRoot | — (둘 다 `BacklogPathForRoot` 로 확장) | 일치 |

소비자: `internal/web/todo_queue_read.go:33-35`(`vm.Root` 에 리졸버 값, `BacklogPathForRoot(root)` 로 읽기), `internal/cli/todo.go:59,74`(Adopting 리졸버 + `BacklogPathForRootAdopting`), `internal/statusline/landed.go:210` · `backlog.go:48`(boardRoot = `resolveStateAnchor`, 리졸버 미경유), `internal/web/events.go:176`(`StateDirForRoot(root)`, 리졸버 미경유), `internal/kanban/record.go:189` · `internal/cli/kanban.go:330,368`(`RuntimeStateDirForRoot`). 리졸버 반환값을 받는 소비자는 web 과 cli 둘뿐이고, 둘 다 root 로 취급한다.

## 4. 변경 파일

- `internal/kanban/todo_root.go` — no-home 분기 `return base, false`, 함수 주석과 `tempOriginSubstituteRoot` 주석을 현재 상태에 맞춤
- `internal/kanban/todo_root_temp_guard_test.go` — 옛 no-home 값을 전제로 한 서브테스트 주석 갱신(단언 변경 없음)
- `internal/kanban/root_layer_repro_test.go` — 신규 회귀 테스트 3개

## 5. 미검증 (Gaps)

1. **홈 분기 root 의 재키잉 (카드 발행 대상, 이 카드에서 수리 안 함).** 로그 관측만 했다(`probe-develop-d3b7d438d.txt`, `TestT549_ProbeHomeRootRekeying`): `BacklogPathForRoot(homeRoot)` = `<home>/.moai/db/001-b770b8ec-f195ba1f/todo/backlog.json`, `BacklogPathForRoot(base)` = `<home>/.moai/db/001-b770b8ec/todo/backlog.json`. 즉 `~/.moai/todo/<key>` 가 다시 프로젝트 키로 쓰여 canonical 과 다른 db 디렉터리가 생긴다. 읽기·쓰기가 같은 함수를 거치므로 서로는 일관되지만, 이것이 사용자에게 보이는 오작동인지는 측정하지 않았다.
2. **fallback 의 이름 기반 stat 가설 (미측정).** `fallbackTodoQueueRoot` 와 `adoptLocalTodoQueue` 는 `os.Stat(BacklogPathForRoot(…))` 로 `backlog.json` 이름만 확인한다. 저장소가 `backlog.db` 만 가진 레이아웃이면 기존 큐를 못 보고 빈 홈 root 를 고를 수 있다는 가설이다. 코드 판독뿐이고 실행 관측은 없다.
3. 심볼릭 링크 도달성은 macOS(이 머신)에서만 관측했다. linux·windows 에서는 미관측이다(windows 는 `/usr` 부재로 skip 예상).
4. internal/cli 는 `-run 'Todo'` 스코프만 돌렸다. 패키지 전체와 `internal/cli/wizard`, `internal/statusline` 은 이 카드에서 돌리지 않았다(statusline 은 리졸버를 거치지 않아 영향 경로 밖이라고 판독했지만 실행 확인은 아니다). 전체 판정은 develop push 뒤 CI 몫이다.
5. 병합 트리 재측정은 아직 없다 — 창을 받은 뒤 `origin/develop` 흡수 트리에서 다시 잰다.

## 6. 잔여 위험 (Residual-risk)

- 심볼릭 링크 회귀 테스트는 `/usr` 를 읽기 전용 대상으로 쓴다. 대상이 git 저장소이거나 temp 루트 안에 있는 호스트에서는 skip 되어 그 호스트에서는 아무것도 단언하지 않는다.
- no-home 반환이 base 가 되면서 `ResolveTodoQueueRootAdopting` 의 no-home 경로가 `moai todo` 에서 `BacklogPathForRootAdopting(base)` 로 이어져 legacy 디렉터리 이관(adopt=true)이 이제 실제 프로젝트 경로에서 일어날 수 있다. 수리 전에는 두 겹 경로에서 일어났다(사실상 아무것도 이관되지 않음). 명령 경로에서 이관은 원래 의도된 동작이지만, 이 조합(HOME 미해소)에서 실행해 보지는 않았다.
- 동시 실행 조건: internal/cli 스코프는 lane-6 의 `go test ./internal/cli/ -count=1 -timeout 60m` 전체 스위트와 병행해 돌았다(`cli-todo-postfix.meta`). 부하로 인한 시간 초과·flaky 가 결과에 섞일 수 있다.
