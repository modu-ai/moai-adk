# t598 — 팀 디렉터리 부재만으로 타 세션 작업 목록을 즉시 삭제하는 결함 (hooks 감사 H03)

- 카드: t598 (홈 저장소 정본 — `moai todo pr t598`)
- 대상: `internal/hook/session_end.go` `garbageCollectOrphanedTasks`
- 기준 트리: `9935e4e3e1067cc5d5d4576f8ae4fd5410d838d6` (로컬 develop == origin/develop)
- 워크트리: `.claude/worktrees/t598`, 브랜치 `WT-session-end-task-gc`
- 출처: `reports/hooks-audit-20260911-01a08e35/hooks-audit.md` — H03

## 배차 전제의 정정

배차문 최초본은 `moai integration` 의 settings 드리프트 원장 비대칭을 지시했으나, 큐의 카드 t598
본문은 hooks 감사 H03 이었다. 원인은 백로그 큐 저장소의 분기 — 프로젝트 `.moai/state/todo/backlog.db`
(105건, 최대 t661) 와 홈 `~/.moai/db/moai-adk-go-1bd3d038/todo/backlog.db` (93건, 최대 t654) 가 같은
id 에 다른 카드를 담고 있었다. 운영자 결정으로 **홈 저장소가 정본**(= `moai todo` 가 읽는 곳)이며,
본 카드는 그 본문으로 재배차되어 수행되었다.

## Claim

1. H03 결함은 기준 트리 `9935e4e3e` 에서 살아 있었다 — 감사 당시 복제본(main `2213871af`)의
   소견이 낡지 않았다.
2. 결함은 두 축이다: (a) 나이 기준 부재로 방금 생성된 타 세션 작업 목록이 즉시 삭제된다,
   (b) `os.Stat` 오류가 부재로 읽혀, 존재하지만 읽을 수 없는 팀 디렉터리가 부재로 판정된다.
3. 수리 후 두 축 모두 차단되며, 기존 수거 동작(오래된 진짜 고아 디렉터리 회수)은 보존된다.
4. 패키지 전체 실행의 실패 2건은 본 변경에 귀속되지 않는다 — 기준 트리에서도 동일하게 실패한다.

## Evidence

### E1 — 재현 (수리 전, 기준 트리 `9935e4e3e`)

`go test ./internal/hook/ -run 'TestGarbageCollectOrphanedTasks_(KeepsFreshStandaloneTask|CollectsStaleStandaloneTask|KeepsTaskWhenTeamStatFails)$' -count=1 -v`

```
    session_end_task_gc_test.go:33: freshly created standalone task list was deleted; it must be kept until it is provably finished
    session_end_task_gc_test.go:96: task directory was deleted on an unreadable team directory; a Stat error is not proof of absence
--- PASS: TestGarbageCollectOrphanedTasks_CollectsStaleStandaloneTask (0.00s)
--- FAIL: TestGarbageCollectOrphanedTasks_KeepsFreshStandaloneTask (0.00s)
--- FAIL: TestGarbageCollectOrphanedTasks_KeepsTaskWhenTeamStatFails (0.00s)
FAIL	github.com/modu-ai/moai-adk/internal/hook	0.660s
```

RED 2건은 결함의 두 축에 각각 대응한다. 대조군 `CollectsStaleStandaloneTask` 는 수리 전에도
초록 — 테스트가 전면 실패가 아니라 결함 지점만 짚고 있음을 보인다.

### E2 — 수리 후

`go test ./internal/hook/ -run 'TestGarbageCollect' -count=1`

```
ok  	github.com/modu-ai/moai-adk/internal/hook	0.880s
```

`go test ./internal/hook/ -race -run 'TestGarbageCollect' -count=1`

```
ok  	github.com/modu-ai/moai-adk/internal/hook	2.317s
```

`gofmt -l internal/hook/` → 출력 없음. `go vet ./internal/hook/` → 출력 없음, exit 0.
`golangci-lint run ./internal/hook/...` → `0 issues.`

### E3 — 뮤턴트 3종 (전부 죽음)

| 뮤턴트 | 되돌린 것 | 죽인 테스트 |
|---|---|---|
| M1 | `if false && !info.ModTime().Before(cutoff)` — 나이 기준 무력화 | `KeepsFreshStandaloneTask` **만** |
| M2 | `!os.IsNotExist(err)` → `err == nil` — Stat 오류 구분 제거 | `KeepsTaskWhenTeamStatFails` **만** |
| M3 | `cutoff` 를 `-staleDuration * 1000` 으로 — 수거 경로 도달 불가 | `CollectsStaleStandaloneTask` + `TestGarbageCollectOrphanedTasks` 수거 케이스 2건 (+ 형제 함수 테스트 4건) |

M1·M2 는 복합 조건의 각 절반을 독립적으로 겨냥한다 — 한 뮤턴트가 정확히 한 테스트만 죽이므로
두 축이 서로를 가리지 않는다. M3 은 **도달성 대조군**이다: 수거를 막았을 때 수거를 주장하는
테스트가 빨개지므로, 그 테스트들의 초록은 삭제 경로에 실제로 도달한 결과이지 공허한 초록이 아니다.
(M3 의 `sed` 는 같은 이름의 상수를 쓰는 형제 함수 `garbageCollectStaleTeams` 의 `cutoff` 도 함께
바꾸므로 그쪽 테스트 4건도 죽었다. 의도한 범위보다 넓으나 도달성 근거로는 유효하다.)

뮤턴트 적용 전 원본을 `/tmp/t598_session_end.orig.go` 로 보존하고, 각 뮤턴트 후 복원하여
`diff -q` 로 동일성을 확인했다.

### E4 — 기존 표 테스트의 전제 수정

`TestGarbageCollectOrphanedTasks` (session_end_test.go) 는 **갓 만든** 디렉터리를 놓고 "고아면
지워진다"고 단언하여, 결함 동작 자체를 정답으로 못박고 있었다. 이 테스트의 원래 의도(팀 디렉터리
유무가 판별식)는 보존하고, 생성한 작업 디렉터리 전부를 48시간 전으로 `os.Chtimes` 하여 나이 축을
설정에서 제거했다. 나이 축은 E1 의 신규 테스트 2건이 전담한다.

### E5 — 슬롯 불요 확인

`go list -deps ./internal/hook/... | grep -c "moai-adk-go/internal/cli"` → `0`

## Baseline-attribution

모든 측정은 워크트리 `.claude/worktrees/t598` 에서, `git rev-parse HEAD` =
`9935e4e3e1067cc5d5d4576f8ae4fd5410d838d6` 인 트리 위에서 이번 실행으로 수행했다.
`git rev-parse --show-toplevel` = `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t598`.
로컬 `develop` 과 `origin/develop` 이 모두 동일 SHA 임을 착수 시 확인했다.

패키지 전체 실패 2건의 귀속은 **같은 트리에서 변경을 걷어내고** 재측정했다:
추적 파일 2개를 `git restore --source=HEAD` 로 되돌리고 신규 테스트 파일을 트리 밖으로 옮긴 뒤

`go test ./internal/hook/ -run 'TestSessionStart_DeferredScan' -count=3`

```
--- FAIL: TestSessionStart_DeferredScanDoesNotBlockReturn (0.66s)
--- FAIL: TestSessionStart_DeferredScanDoesNotBlockReturn (0.63s)
--- FAIL: TestSessionStart_DeferredScanDoesNotBlockReturn (0.78s)
--- FAIL: TestSessionStart_DeferredScanJoinsWithinBound (1.60s)
    --- FAIL: TestSessionStart_DeferredScanJoinsWithinBound/slow_scan_drops_advisory (0.98s)
FAIL	github.com/modu-ai/moai-adk/internal/hook	6.452s
```

기준 상태에서도 동일하게 실패하므로 **물려받은 빨강**이다. 측정 후 백업본(`/tmp/t598-backup/`)에서
3개 파일을 복원하고 `gofmt`/`go vet`/대상 테스트 재통과를 확인했다.

## Gaps

- **패키지 전체(`go test ./internal/hook/`)는 초록이 아니다.** 위 2건이 남아 있으며 본 카드 범위 밖이다.
  전 패키지·크로스플랫폼 판정은 CI 몫이다(CLAUDE.local.md §4).
- `TestSessionStart_DeferredScan*` 의 **근본 원인은 조사하지 않았다.** 시간 의존 테스트라는 점과
  기준 트리에서도 실패한다는 사실만 관측했다. 별도 카드 소관으로 본다.
- 카드의 완료 조건 문구 중 "**종료가 입증된** 소유 작업만 삭제"는 소유권 기록 도입 없이는 문자 그대로
  충족되지 않는다. 운영자 결정으로 나이 기준을 종료 증명의 대리 지표로 채택했다 — 같은 파일의 형제
  함수 `garbageCollectStaleTeams` 가 이미 쓰는 판별식이다. 소유권 기록 방식은 채택하지 않았다.
- `os.Chmod(dir, 0o000)` 기반 Stat-오류 재현은 **POSIX 권한에 의존**한다. root 실행 시 건너뛰도록
  가드를 두었으나 Windows 에서의 거동은 측정하지 않았다.
- `entry.Info()` 는 `ReadDir` 시점에 캐시된 정보를 반환할 수 있다. 형제 함수와 동일한 관례를 따랐으며,
  이 차이가 실환경에서 문제가 되는지는 측정하지 않았다.
- **push 하지 않았고 CI 판정을 받지 않았다.** develop 병합은 리드의 창 지명 이후다.

## Residual-risk

- 24시간 넘게 조용한 타 세션의 작업 목록은 여전히 수거 대상이다. 형제 함수와 동일한 성질이며,
  카드가 막으려 한 "방금 만든 것을 즉시 지움"은 차단되지만 "오래 조용한 남의 것"은 차단되지 않는다.
  진정한 소유권 판정이 필요하다면 별도 카드가 필요하다.
- 나이 기준 도입으로 고아 디렉터리의 **회수가 최대 24시간 지연**된다. 디스크 점유가 그만큼 늦게
  풀리지만, 데이터 보존을 우선한 의도된 맞교환이다.
- `session_end.go` 는 H02·H04 카드와 공유 수정 경로다. 본 변경은 `garbageCollectOrphanedTasks`
  한 함수와 그 godoc 에 한정되며 다른 함수는 건드리지 않았으나, 병합 순서에 따라 충돌 가능성이 있다.
