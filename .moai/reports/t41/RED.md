# t41 RED — `moai worktree done` 조용한 실패 3결함, 수정 전 실패 증거

카드: t41 — `moai worktree done`이 git stderr를 전달하지 않고 조용히 실패한다 (lock 감지·`--auto` exit code 포함). Tier S.

## Premise check (사전 판정) — 수정 전 이 트리에서의 항목별 판정

base: `release/v3.1.1` 병합 완료 (`be03419b7`, t40/t42/t43/t46/t53/t62/t73/t74 포함). t46/t73/t74(`cf749fafe` 등)의 anchor guard는 **이미 존재**하며 본 카드와 무관한 경로(anchored session 거절, `ANCHORED_SESSIONS_PRESENT` sentinel — loud)에서 동작한다.

| 항목 | 판정 | 근거 |
|---|---|---|
| (a) git stderr 전달 | **미해결 — 이 트리에서 재현** | 실측 재현: locked tree에서 `done` → **rc=128, stderr 0줄, stdout 0줄**. 다만 원인은 "stderr가 error에 없다"가 아니라 **구조적 ExitCoder 매칭이 출력을 억제**하는 것 — `execGit`은 stderr를 error message에 이미 포함(`TestExecGit_FailureKeepsStderrVisible`는 수정 전에도 PASS). `*exec.ExitError`가 `ExitCode() int`를 갖고 있어 `cmd/moai/main.go`의 `ExitCoder` 인터페이스와 `internal/cli/fang.go`의 `moaiErrorHandler`(ExitCoder carrier silent)에 **우연히** 부합 → error text 미출력 + git 종료코드(128) 관통. |
| (b) lock 감지 안내 | **미해결 — 코드에 전혀 없음** | `worktreeManager.Remove`의 substring 매핑은 `is not a working tree` / `contains modified or untracked files` 2종뿐. `cannot remove a locked working tree` 미매핑 → default 경로(=침묵 경로)로 떨어짐. |
| (b2) lock 없는데 무출력 실패 | **양분** | 침묵 메커니즘은 (a)와 동일(모든 git 실패가 default 경로에서 무출력). done이 raw 명령과 달리 실패하는 **구조적 발산원**은 `Remove`의 `context.WithTimeout(..., 10*time.Second)` — raw `git worktree remove`에는 deadline이 없음. 큰 트리/부하 상황에서 10s 내 완료 실패 시 git이 kill되고(부분 삭제) 이후 raw 명령은 rc=0으로 마무리 — 카드 관측과 부합. lock 없는 removal 실패 자체는 오늘 재현 못함(clean tree → done으로 제거 성공, dirty tree → loud `ErrWorktreeDirty`). |
| (c) `--auto` exit code | **미해결 — 이 트리에서 재현** | 실측 재현: locked tree에서 `done --auto` → **rc=0, 출력 없음**. `runDoneWithAutoMode`의 모든 실패 경로가 `return false, nil`. |

측정 메모: git 2.50.1(Apple Git-155)에서 lock 파일명은 `.git/worktrees/<n>/lock`이 **아니라** `locked`다. `lock`으로 수동 생성 시 git은 lock 없이 제거해버린다 — lock 재현은 반드시 `git worktree lock --reason ...`으로 할 것(이 파일명 차이가 처음에 재현을 놓치게 했다).

## Claim (주장)

수정 전에는:

1. (a) git subprocess 실패 error chain이 구조적 ExitCoder를 만족시켜 CLI 경계에서 무출력+원 종료코드 관통 — 이를 잡는 테스트가 런타임에 실패한다.
2. (b) `ErrWorktreeLocked` sentinel과 lock 안내(`git worktree unlock` / `-f -f`)가 존재하지 않아, 이를 참조하는 테스트는 컴파일부터 실패한다.
3. (b2) `removeTimeout` 상수가 존재하지 않아(10s 하드코딩) budget 테스트가 컴파일부터 실패한다.
4. (c) `--auto` 실패가 non-zero여야 한다는 테스트가 런타임에 실패한다(현재 rc=0 삼킴).

## Evidence (증거)

### E-1. 바이너리 실물 재현 (수정 전 빌드, 카드 관측의 원 재현)

준비 (임시 repo, `/tmp/t41-repro.0IPTad`):

```
git worktree add -b feature/w4 .../wt4
git worktree lock --reason "claude session (pid 12345)" .../wt4
git worktree remove .../wt4            → rc=128  "fatal: cannot remove a locked working tree, lock reason: claude session (pid 12345)" + "use 'remove -f -f' to override or unlock first"
git worktree remove --force .../wt4    → rc=128  (동일 메시지 — force 1회로는 못 넘음; 카드 서술과 일치)
```

수정 전 `/tmp/t41-moai` (본 워크트리 소스 빌드) 실행:

```
$ /tmp/t41-moai worktree done feature/w4   2>stderr 1>stdout; echo rc=$?
rc=128        # stderr 0 bytes, stdout 0 bytes  ← 카드 (a) 그대로
$ /tmp/t41-moai worktree done feature/w4 --auto 2>stderr 1>stdout; echo rc=$?
rc_auto=0     # stderr 0 bytes, stdout 0 bytes  ← 카드 (c) 그대로
```

### E-2. 런타임 RED (구조적 ExitCoder — 기호만으로 작성, 구현 전)

```
go test ./internal/core/git/ -run 'TestExecGit_Failure' -count=1 -timeout 60s
```

출력 (전문):

```
--- FAIL: TestExecGit_FailureDoesNotSatisfyStructuralExitCoder (0.02s)
    exec_stderr_test.go:54: error chain must not satisfy the structural ExitCoder interface (it would print nothing and pass git's exit code through): git status: fatal: not a git repository (or any of the parent directories): .git: exit status 128
    exec_stderr_test.go:54: error chain must not satisfy the structural ExitCoder interface (it would print nothing and pass git's exit code through): remove worktree: git status: fatal: not a git repository (or any of the parent directories): .git: exit status 128
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/core/git	1.551s
```

(`TestExecGit_FailureKeepsStderrVisible`는 PASS — stderr text는 error 안에 원래 있었다. 결함은 출력 억제.)

### E-3. 컴파일 RED (신규 API — `ErrWorktreeLocked`, `removeTimeout`)

```
go test ./internal/core/git/ -run 'TestWorktreeRemove' -count=1 -timeout 120s
```

출력 (전문):

```
# github.com/modu-ai/moai-adk/internal/core/git [github.com/modu-ai/moai-adk/internal/core/git.test]
internal/core/git/worktree_locked_test.go:28:21: undefined: ErrWorktreeLocked
internal/core/git/worktree_locked_test.go:39:21: undefined: ErrWorktreeLocked
internal/core/git/worktree_locked_test.go:53:5: undefined: removeTimeout
internal/core/git/worktree_locked_test.go:54:96: undefined: removeTimeout
FAIL	github.com/modu-ai/moai-adk/internal/core/git [build failed]
FAIL
```

### E-4. 컴파일 RED (신규 API — `git.CommandError` in CLI 테스트)

```
go test ./internal/cli/worktree/ -run 'TestRunDone' -count=1 -timeout 120s
```

출력 (전문):

```
# github.com/modu-ai/moai-adk/internal/cli/worktree [github.com/modu-ai/moai-adk/internal/cli/worktree.test]
internal/cli/worktree/done_stderr_test.go:42:10: undefined: git.CommandError
internal/cli/worktree/done_stderr_test.go:80:15: undefined: git.ErrWorktreeLocked
internal/cli/worktree/done_stderr_test.go:110:10: undefined: git.CommandError
internal/cli/worktree/done_stderr_test.go:125:10: undefined: git.CommandError
FAIL	github.com/modu-ai/moai-adk/internal/cli/worktree [build failed]
FAIL
```

## Baseline-attribution (baseline 귀속)

- 워크트리: `.claude/worktrees/agent-ae3785cfe611541f0` (branch `WT-t41`, base = `be03419b7` = release/v3.1.1 병합 후, 구현 0커밋 상태).
- E-1의 바이너리는 이 트리의 수정 전 소스로 `go build -o /tmp/t41-moai ./cmd/moai` 빌드한 것. E-2~E-4는 같은 트리에서 신규 테스트 파일 추가 직후 실행.

## Gaps (미검증)

- (b2)의 ">10s removal" 시나리오를 실제 느린 트리로 재현하지 못함 — 대형 fixture가 필요해 로컬에서 비현실적. 발산원은 코드 검사(`Remove`의 10s ctx vs raw 무deadline)로 확정하고 budget 상향+deadline 명시로 대응.
- lock 없는 removal 실패(dirty 아님) 사례를 오늘 재현 못함 — clean tree는 done으로 제거되고 dirty는 loud `ErrWorktreeDirty`로 떨어짐. 카드의 issue-1467 사례 원인 특정은 10s kill이 유력 후보이나 워크플로우 로그 부재로 단정 불가.

## Residual-risk (잔여 위험)

- 구조적 ExitCoder 부합 제거는 CLI 전역에 적용되는 변경 — git 실패가 기존 128 등 원 코드 대신 1로 나간다. 의도적 carrier(`worktree verify` 0/1/2/3, hook exit 2 등)는 `TestFangExitCoderCharacterization`로 보존 확인(GREEN 참조).
