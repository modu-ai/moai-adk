# t41 GREEN — `moai worktree done` 조용한 실패 3결함, 수정 후 통과 증거

## Claim (주장)

1. **(a)** git subprocess 실패의 stderr가 사용자에게 그대로 도달한다 — `execGit`이 `*exec.ExitError`를 chain에 노출하지 않고 `CommandError`(stderr 원문 + ExitStatus 데이터)로 감싸, 구조적 ExitCoder 부합(무출력 + git 원 종료코드 관통)이 사라졌다. 실패 시 fang이 error를 출력하고 main이 1로 종료.
2. **(b)** locked tree 실패가 `ErrWorktreeLocked` sentinel으로 매핑되고(단일 `--force`로도 매핑 유지), done이 복사-붙여넣기 가능한 안내(`git worktree unlock <path>` / `git worktree remove -f -f <path>`)를 붙인다. done은 스스로 `-f -f`로 승격하지 않는다.
3. **(b2)** removal deadline이 10s → 2분으로 상향(raw 명령과의 구조적 발산 해소), deadline kill 시 error가 deadline을 명시.
4. **(c)** `--auto`는 성공 출력만 억제한다 — 실패(List/Remove/DeleteBranch/nil provider)는 error로 나와 non-zero exit. 의도된 2개 비-에러 종료(대상 worktree 없음 = 완료, anchored-session skip = t46)는 유지.

## Evidence (증거)

### G-1. 단위 테스트 (구현 후, 동일 워크트리)

```
go test ./internal/cli/worktree/ ./internal/core/git/ -count=1 -timeout 540s
```

출력 (전문):

```
ok  	github.com/modu-ai/moai-adk/internal/cli/worktree	6.872s
ok  	github.com/modu-ai/moai-adk/internal/core/git	89.626s
```

신규/갱신 테스트 (같은 패키지 `-v` 실행에서):

```
--- PASS: TestExecGit_FailureKeepsStderrVisible            (신규, (a) — 수정 전에도 PASS: text는 원래 있었음)
--- PASS: TestExecGit_FailureDoesNotSatisfyStructuralExitCoder (신규, (a) — RED→GREEN)
--- PASS: TestExecGit_DeadlineKillIsLegible                (신규, (b2) deadline 명시)
--- PASS: TestWorktreeRemove_LockedTreeMapsToSentinel      (신규, (b) — force 유무 무관 매핑 + lock reason 노출)
--- PASS: TestWorktreeRemoveTimeoutBudget                  (신규, (b2) — ≥ 2m)
--- PASS: TestRunDone_RemoveFailureKeepsGitStderr          (신규, (a) done 래핑 보존)
--- PASS: TestRunDone_LockedTreeGivesActionableGuidance    (신규, (b) — unlock/-f -f/path 안내 + 비승격)
--- PASS: TestRunDoneAuto_RemoveFailureIsNonZero           (신규, (c) — RED→GREEN)
--- PASS: TestRunDoneAuto_ListFailureIsNonZero             (신규, (c))
--- PASS: TestRunDoneAuto_NoWorktreeIsStillSuccess         (신규 — 의도된 grace 보존)
--- PASS: TestRunDoneAuto_SuccessStaysOutputSilent         (신규 — 성공 무출력 계약 보존)
--- PASS: TestRunDoneWithAutoMode_NoProvider               (갱신 — 삼킴→error, t41 c)
--- PASS: TestRunDoneWithAutoMode_SkipsRemovalWhenAnchoredSessionLive (기존 t46 — 시그니처 갱신만, 의미 불변)
--- PASS: TestRunDone_RefusesWhenAnchoredSessionLive       (기존 t46)
--- PASS: TestRunDone_ForceOverridesAnchorWarning          (기존 t46)
--- PASS: TestRunDone_RemovesWhenNoAnchoredSession         (기존)
--- PASS: TestRunDoneWithAutoMode_AfterMerge               (기존, 시그니처 갱신)
--- PASS: TestRunDoneWithAutoMode_DeleteBranch             (기존, 시그니처 갱신)
--- PASS: TestWorktreeRemove_Success / _NotFound / _Force  (기존)
```

패키지 전체(`-v` 요약)는 71건 전수 PASS — `internal/cli/worktree` 71/71 (FAIL 0), `internal/core/git` 전체 `ok` (89.6s).

교차 확인 — 의도적 ExitCoder carrier 보존 (CLI 경계 변경의 회귀 점검):

```
go test ./internal/cli/ -run 'TestFang' -count=1 -timeout 120s
--- PASS: TestFangExitCoderCharacterization (0.00s)
ok  	github.com/modu-ai/moai-adk/internal/cli	1.766s
```

### G-2. 바이너리 실물 (구현 후 빌드, RED의 E-1과 동일 fixture·동일 명령)

수정 후 재빌드한 `/tmp/t41-moai`로 동일 locked tree(`git worktree lock --reason "claude session (pid 12345)"`)에서:

```
$ /tmp/t41-moai worktree done feature/w4 2>stderr 1>stdout; echo rc=$?
rc=1
```

stderr (전문 — fang error box):

```
   ERROR
  Remove worktree: remove worktree at "/private/tmp/t41-repro.0IPTad/wt4": git: worktree is locked (git worktree:
  fatal: cannot remove a locked working tree, lock reason: claude session (pid 12345)
  use 'remove -f -f' to override or unlock first)

  The worktree is locked — usually a live session is still using it:
    unlock and retry:  git worktree unlock /private/tmp/t41-repro.0IPTad/wt4
    remove anyway:     git worktree remove -f -f /private/tmp/t41-repro.0IPTad/wt4
  moai does not force a locked tree on its own.
```

→ (a) git stderr 원문(`fatal: ...` + git 자체 hint) 전달, (b) 안내 제공, 종료코드 128→1.

```
$ /tmp/t41-moai worktree done feature/w4 --auto 2>stderr 1>stdout; echo rc=$?
rc_auto=1        # 동일 stderr — 카드 (c) 해소 (수정 전 rc=0 무출력)
```

정상 경로 회귀 (lock 해제 후):

```
$ git worktree unlock .../wt4
$ /tmp/t41-moai worktree done feature/w4; echo rc=$?
rc=0             # 성공 카드 출력("Done: worktree for branch feature/w4 / Worktree removed.")

$ /tmp/t41-moai worktree done feature/happy --auto; echo rc=$?
rc_auto_happy=0  # stdout 0 bytes — --auto 성공 무출력 계약 유지
```

### G-3. 정적 게이트

```
go vet ./internal/cli/worktree/... ./internal/core/git/...   → exit 0 (출력 없음)
golangci-lint run ./internal/cli/worktree/... ./internal/core/git/... → 0 issues. (exit 0)
gofmt -l internal/cli/worktree/ internal/core/git/           → 출력 없음 (정상)
go build ./...                                               → exit 0 (출력 없음)
```

## 변경 파일

- `internal/core/git/manager.go` — `execGit` 실패 경로: `CommandError`(Op/Stderr/ExitStatus) 신설, `*exec.ExitError` 비-chain, deadline kill 명시
- `internal/core/git/errors.go` — `ErrWorktreeLocked` sentinel
- `internal/core/git/worktree.go` — lock substring 매핑 + `removeTimeout` 10s→2m
- `internal/cli/worktree/done.go` — `runDoneWorktreeCleanup`(auto 삼킴 제거), `doneRemoveError`/`lockGuidance`, interactive 경로 동일 적용, `--auto` flag help 정직화
- 테스트: `internal/core/git/exec_stderr_test.go`(신규), `internal/core/git/worktree_locked_test.go`(신규), `internal/cli/worktree/done_stderr_test.go`(신규), `done_test.go`/`done_anchor_test.go`(시그니처·기대치 갱신)

## Baseline-attribution (baseline 귀속)

- 위 전부 동일 워크트리(branch `WT-t41`, 구현 커밋 후 작업 트리)에서 이번 실행으로 관측. G-2 바이너리는 이 트리 소스로 `go build -o /tmp/t41-moai ./cmd/moai` 재빌드.

## Gaps (미검증)

- (b2) ">10s removal" 실제 대형 트리 재현 미실행(RED와 동일 사유) — budget 상향(2m)과 deadline 명시로 대응했으나 실측은 아님.
- `internal/cli` 전체 수트는 실행 안 함(카드 대상 패키지 + 소비 경계인 fang 특성화만) — 전 판정은 CI.
- 로컬 전체 저장소 수트(`go test ./...`)는 규율(CLAUDE.local.md §4)상 미실행.

## Residual-risk (잔여 위험)

- git 실패의 종료코드가 128 등 원 값에서 1로 바뀌었다(의도적). 원 코드에 의존하던 외부 자동화가 있다면 영향 — 저장소 내 의존 없음을 grep으로 확인(`exec.ExitError` consumer는 모두 자기 cmd.Run 직접 실행 경로).
- 2분 deadline 내에도 못 끝내는 트리는 여전히 실패하나, 이제 원인(deadline)과 대책(raw 명령)이 error에 명시된다.
- `git worktree remove` 외 경로(`clean`의 `gitWorktreeCmd` 등)는 본 변경의 적용 밖이나, `clean`은 이미 `Warning: could not remove ...`로 error text를 출력하고 있어 침묵 경로가 아님(조사 기록, 미변경).
