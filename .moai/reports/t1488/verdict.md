# t1488 — `moai codex -l` 레인 루프가 첫 카드 뒤에 멈추는 결함

브랜치 `WT-codex-lane-loop-no-exec`, 기준 develop `45ed3d29d`. 결함 근거: t1440 P-1 측정(`.claude/worktrees/t1440/.moai/reports/t1440/p1-syscall-exec.md`).

## Claim

POSIX 기본 직접 실행(`defaultCodexDirectLaunch`)이 레인 루프의 카드 세션까지 `syscall.Exec`로 띄워 런처 프로세스가 교체되고, 루프가 두 번째 카드를 임대하지 못했다. 수리 후 카드 세션(카드 식별 변수 `MOAI_KANBAN_CARD`가 있는 실행 — `codexCardLaunchEnv`만 설정하고 직접 경로는 스크럽한다)은 자식 프로세스로 띄우고 기다린다. 일회성 직접 실행의 `syscall.Exec`는 그대로다.

## Evidence

수리: Windows 쪽 start-and-wait 본문을 플랫폼 중립 `internal/cli/codex_direct_wait.go`의 `codexStartAndWait`로 옮기고(Windows 기본값은 이를 호출, 동작 불변), POSIX 기본값은 카드 변수가 있으면 이를 호출한다. 루프 영역(`codex_launcher.go`)은 손대지 않았다 — 루프 주석 "no process replacement"는 이제 사실이다.

RED (수리 전, 같은 테스트):

```
--- FAIL: TestCodexLaneLoopDefaultLaunchContinuesAfterFirstCard (1.76s)
    codex_lane_loop_no_exec_test.go:39: child pid=67706 err=<nil>
        fake-codex log:
        card=t1,pid=67706
    codex_lane_loop_no_exec_test.go:44: the lane loop did not return in the child (process replaced?):
        === RUN   TestCodexLaneLoopDefaultLaunchContinuesAfterFirstCard
    codex_lane_loop_no_exec_test.go:47: fake codex invoked 1 times, want 2 (one per queued card): ["card=t1,pid=67706"]
FAIL
```

GREEN:

```
=== RUN   TestCodexLaneLoopDefaultLaunchContinuesAfterFirstCard
    codex_lane_loop_no_exec_test.go:39: child pid=77130 err=<nil>
        fake-codex log:
        card=t1,pid=77610
        card=t2,pid=77983
--- PASS: TestCodexLaneLoopDefaultLaunchContinuesAfterFirstCard (4.93s)
```

- `go test -count=1 -race -run 'Codex' ./internal/cli/` (env 스크럽 한 호출): FAIL 1건 `TestCodex1718Fixtures_WidenedContent` 뿐 — 기준 트리(수리 파일을 `45ed3d29d`판으로 되돌린 상태)에서도 같은 테스트가 `--- FAIL` 재현, 이 카드와 무관(리뷰 finding 파싱).
- 이동한 리터럴 grep 적중 테스트: `TestRestampSeamIsCalledAtEveryNonReplaceCallSite`가 소스 수준으로 `codex_direct_windows.go`의 `stampFactoryRunOwner(` 호출을 요구해 적색 → 호출 지점이 옮겨진 파일(`codex_direct_wait.go`)을 가리키도록 갱신. 재실행 `-race -run 'Restamp|RunOwner|PaneDoor|DirectPOSIX|SD_AC003|LaneLoop'` → `ok ... 22.934s`.
- `go build ./...`, `GOOS=windows go build ./...`, `go vet ./internal/cli/`, `GOOS=windows go vet ./internal/cli/` 통과; `golangci-lint v2.1.6 run ./internal/cli/` → `0 issues.`

## Baseline-attribution

모두 이 실행, 이 워크트리(HEAD `45ed3d29d` + 본 변경)에서 측정. go test가 트리를 빌드하므로 설치된 moai 바이너리는 관여하지 않는다.

## Gaps

- Windows 런타임 실행 미측정(교차 빌드·vet만). 새 회귀 테스트는 `!windows` 빌드 태그로 POSIX 한정.
- `-run 'Factory|Managed'`까지 넓힌 -race 실행은 기본 10분 제한에 걸려 판정 불가(panic: test timed out) — 좁힌 셀렉터로 재측정했다. 전 패키지 판정은 CI 몫.
- codex 리뷰 지적 전, 새 테스트의 자식이 실제 `~/.moai/run/`에 run 상태를 썼을 수 있다(HOME 미격리 상태로 2회 실행). 이후 자식 HOME을 임시 디렉터리로 격리했다. 잔여물 여부는 확인하지 않았다.

## Residual-risk

- 카드 세션 판별이 env 키(`MOAI_KANBAN_CARD`) 존재에 기댄다. 다른 직접 실행 경로가 나중에 이 키를 싣게 되면 그 경로도 자식 실행으로 바뀐다(현재는 스크럽 목록 `codex_launcher.go:331`이 막는다).
- POSIX 카드 세션의 앵커 pid는 이제 런처 pid가 아니라 자식 pid로 레인 클레임에 찍힌다(Windows와 같은 형태). 레인 클레임 소비자가 런처 pid를 가정했다면 영향이 있을 수 있다 — 관련 테스트는 초록.
