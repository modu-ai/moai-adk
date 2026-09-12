# t597 — 팀 디렉터리 시각 오독으로 인한 살아있는 팀 삭제 (hooks 감사 H02)

- 카드: t597
- 워크트리: `.claude/worktrees/t597` · 브랜치 `WT-team-dir-mtime`
- 기준 트리: `eabce74448e094dd1a4393044138a04b2016af99` (origin/develop = local develop)
- 대상: `internal/hook/session_end.go` — `garbageCollectStaleTeams`

---

## Claim

1. 결함은 기준 트리 `eabce7444`에 살아 있었고, 형제 카드 t598의 수리는 이 결함에 닿지 않았다.
2. 원인은 디렉터리 inode 시각(mtime)을 팀의 활동 시각으로 읽은 것이다.
3. 수리 후 살아있는 팀(내용이 최근에 갱신된 팀)은 보존되고, 완전히 조용해진 팀은 여전히 수거된다.
4. 측정 불가(ReadDir/Stat 실패)는 보존 쪽으로 떨어진다.

## Evidence

### E1 — t598 영향 판정 (주장 1)

```
$ git diff 626fb4a9b..eabce7444 -- internal/hook/session_end.go
```
출력: `garbageCollectOrphanedTasks` 의 주석·본문만 변경(나이 게이트 추가 + Stat 오류 구분).
`garbageCollectStaleTeams` 구간은 diff에 나타나지 않음 — 한 줄도 바뀌지 않았다.

### E2 — 재현 (주장 1·2), 수리 전

```
$ go test ./internal/hook/ -run 'TestGarbageCollectStaleTeams_(KeepsLiveTeamWithOldDirMtime|StillCollectsFullyQuietTeam)' -count=1 -v
```
```
--- PASS: TestGarbageCollectStaleTeams_StillCollectsFullyQuietTeam (0.00s)
--- FAIL: TestGarbageCollectStaleTeams_KeepsLiveTeamWithOldDirMtime (0.00s)
    session_end_stale_team_repro_test.go:57: live team directory was deleted: a stale directory mtime is not evidence the owning session ended
    session_end_stale_team_repro_test.go:60: live task directory was deleted alongside the live team directory
FAIL	github.com/modu-ai/moai-adk/internal/hook	0.869s
```

RED 의 이유: 재현군의 팀 디렉터리는 mtime 만 25시간 전으로 돌렸고, 그 안의 `config.json`(다른 세션 소유, `leadSessionId: other-live-session`)과 짝 작업 디렉터리의 `task-1.json`(`status: in_progress`)은 방금 쓴 파일이다. 디렉터리 mtime 은 항목의 생성·삭제·개명에만 반응하고 기존 파일의 제자리 덮어쓰기에는 반응하지 않으므로, 오래 사는 팀의 디렉터리 시계는 태어난 시각에 멈춘다.

대조군은 같은 트리에서 PASS — 즉 이 RED 는 "수거 자체가 동작하지 않아서"가 아니다.

### E3 — 수리 후 (주장 3·4)

```
$ go test ./internal/hook/ -run 'TestGarbageCollect|TestSessionEnd' -count=1 -v | grep -cE "^--- PASS"
29
$ go test ./internal/hook/ -run 'TestGarbageCollect|TestSessionEnd' -count=1
ok  	github.com/modu-ai/moai-adk/internal/hook	4.451s
```

쓸어담은 건수 29건(빈 선택 아님). 포함된 테스트:

- `TestGarbageCollectStaleTeams_KeepsLiveTeamWithOldDirMtime` — 재현군, RED→GREEN
- `TestGarbageCollectStaleTeams_StillCollectsFullyQuietTeam` — 대조군, 계속 GREEN
- `TestGarbageCollectStaleTeams_KeepsTeamWhenTaskActivityUnmeasurable` — 측정 불가 시 보존 (작업 디렉터리를 `chmod 000` 으로 봉인, root 에서는 skip)
- 기존 `TestGarbageCollectStaleTeams*` / `TestGarbageCollectOrphanedTasks*` 전부

### E4 — 정적 검사

```
$ gofmt -l internal/hook/          # 출력 없음
$ go vet ./internal/hook/          # 출력 없음
$ golangci-lint run ./internal/hook/...
0 issues.
```

## Baseline-attribution

위 명령은 모두 이 실행에서, 이 트리(`eabce7444` + 작업 중 변경분)에 대해 수행했다. E2 는 수리 적용 **전**, E3·E4 는 적용 **후**의 같은 워크트리다. 다른 패키지·다른 시점의 수치를 옮겨 오지 않았다.

## 수리 내용

`garbageCollectStaleTeams` 의 판정 근거를 바꿨다.

- 이전: 팀 디렉터리 자체의 `ModTime()` 하나.
- 이후: `newestActivity()` — 팀 디렉터리와 그 아래 전체, 그리고 짝 작업 디렉터리와 그 아래 전체에서 가장 최근 mtime. 그 값이 24시간 창 안이면 보존.
- 측정 실패(`os.Stat` / `filepath.WalkDir` 오류, ENOENT 제외)는 경고를 남기고 **보존**. 대상이 아예 없는 경우(ENOENT)만 "기여 없음"으로 넘어간다.

기존 테스트 `TestGarbageCollectStaleTeams_AlsoRemovesTaskDir` 의 셋업을 한 군데 고쳤다: "stale" 시나리오인데 짝 작업 디렉터리를 방금 만든 상태로 두고 있었다 — 그 모양이 정확히 이 결함이 삭제하던 대상이라, 작업 디렉터리도 같은 과거 시각으로 맞췄다. **단정문은 건드리지 않았다.**

## Gaps (관측하지 않은 것)

- `internal/hook` 패키지 **전체** 수트: 다른 레인들이 같은 패키지를 동시에 돌려 경합이 걸린 상태라 이 세션에서 완주 결과를 관측하지 못했다. 관측한 것은 위 29건의 대상 선택 실행뿐이다. 전 패키지 판정은 CI 몫이다 (CLAUDE.local.md §4).
- 실제 `~/.claude/teams` · `~/.claude/tasks` 에서의 동작은 관측하지 않았다. 재현·검증 전부 `t.TempDir()` 임시 홈에서 수행했고 실제 경로는 읽기만 했다(현재 `~/.claude/teams` 는 비어 있음).
- 다른 OS(linux/windows)에서의 동작은 관측하지 않았다. `chmod 000` 봉인 테스트는 root 실행 시 skip 하도록 했다.

## Residual-risk

- **살아 있지만 24시간 내내 아무것도 쓰지 않은 팀은 여전히 수거된다.** 운영자 응답을 기다리며 멈춰 있는 팀이 이 모양이다. 카드가 적은 "세션 생존 상태 / 명시적 임대 만료 기록" 은 이 호출 지점에서 근거가 없다 — `garbageCollectStaleTeams` 는 `homeDir` 만 받고, 세션 레지스트리는 프로젝트 상대 경로(`.moai/state/active-sessions.json`)라 다른 프로젝트의 세션은 원리상 보이지 않으며, 팀 `config.json` 에는 `leadSessionId` 뿐이라 pid 도 임대 만료 시각도 없다. 임대·하트비트 축은 리드가 별도 카드로 발행 목록에 올렸다.
- 활동 판정이 파일시스템 mtime 에 기댄다는 점은 그대로다. mtime 을 갱신하지 않는 저장 방식(예: 원자적 rename 이 아닌 mmap 쓰기)이 쓰이면 같은 계열의 오독이 재발할 수 있다. 현재 Claude Code 의 쓰기 방식은 관측하지 않았다.
- 트리 순회가 추가되어 팀 하나당 디렉터리 순회 비용이 늘었다. 팀 디렉터리는 작고 SessionEnd 훅은 5초 예산이라 문제 되지 않을 것으로 보지만, 대형 작업 목록에서의 실측은 하지 않았다.
