# t641 판정 — `Status()` 가 인덱스 쓰기 락을 잡지 않게 한다

- 워크트리: `.claude/worktrees/t641`, 브랜치 `WT-status-lock-free`
- 커밋: `3f7a92a17`(재현 baseline) → `7937b4b69`(RED) → `5b42b3301`(GREEN) → 이 판정서를 담은 기록 커밋
- 측정 트리: `internal/core/git/manager.go` blob `75503e276f98560d000e63a7c569bf7f45d437cf` (GREEN 커밋의 blob 과 같다 — `mutant-restore.txt`)

## Claim

1. `gitManager.Status()` 는 `git --no-optional-locks status --porcelain --branch` 를 실행하며, stat 정보가 낡은 저장소에서도 `.git/index` 를 다시 쓰지 않는다.
2. 새 테스트 `TestStatusDoesNotRewriteIndex` 는 수정 전 코드와 플래그를 뺀 뮤턴트에서 실패하고, 수정 후 코드에서 통과한다. 같은 조건으로 준비한 대조 저장소에서는 플래그 없는 `git status` 가 인덱스를 다시 쓰는 것을 매번 확인하므로, 통과가 공허하지 않다.
3. 기존 파싱 동작(Modified/Staged/Untracked/Ahead/Behind)은 바뀌지 않았다. `internal/core` 와 `internal/statusline` 테스트가 모두 통과한다.
4. 바뀐 제품 코드는 `Status()` 의 인자 한 곳과 그 설명 주석뿐이다. `isWorkingTreeClean` 은 그대로 두었다.

## Evidence

### 재현 수치 (baseline 커밋 `3f7a92a17` 에 커밋된 산출물)

플래그 없는 status 루프를 동시에 돌리면서 add/commit 150회를 두 번 시도했다(`.moai/reports/t641/repro/`).

| 모드 | 회차 | add 락 실패 | commit 락 실패 | 합계 | status 실행 수 |
|---|---|---|---|---|---|
| 플래그 없음 | 1 (`plain.json`) | 8 | 4 | 12/150 | 362 |
| 플래그 없음 | 2 (`plain2.json`) | 6 | 2 | 8/150 | 369 |
| `--no-optional-locks` | 1 (`nol.json`) | 0 | 0 | 0/150 | 380 |
| `--no-optional-locks` | 2 (`nol2.json`) | 0 | 0 | 0/150 | 379 |

플래그 없음은 300회 중 20회가 `index.lock: File exists` 로 실패했고, 플래그를 붙이면 300회 중 0회였다. 실패 표본은 모두 exit 128 과 `Unable to create '.../index.lock': File exists` 이다. 측정 당시 부하 평균은 7.5~11.9 였다(`plain-load.txt`, `nol-load.txt`, `round2-load.txt`). 이 수치는 이번 세션에서 다시 잰 값이 아니라 baseline 커밋에 커밋된 기록이다.

`index-rewrite-probe.txt`: stat 정보가 낡은 상태에서 플래그 없는 status 는 인덱스의 sha 와 mtime 을 모두 바꿨고(`index_sha_changed=True index_mtime_changed=True`), 플래그를 붙인 쪽은 둘 다 그대로였다.

### RED — 커밋 `7937b4b69`, `.moai/reports/t641/red.txt`

```
control: plain status rewrote .git/index: sha dd0dbce4... -> 78734cf1...
Status() rewrote .git/index: sha 4d5eb421... -> 8c52dd90..., mtime 2026-09-10T22:05:26.018949738+09:00 -> 2026-09-10T22:05:28.984408439+09:00
--- FAIL: TestStatusDoesNotRewriteIndex (3.94s)
EXIT=1
```

실패 원인이 의도한 것(인덱스 재작성)이며, 대조군이 먼저 통과했으므로 stat 정보가 실제로 낡아 있었다.

### GREEN — 커밋 `5b42b3301`, `.moai/reports/t641/green.txt`

```
control: plain status rewrote .git/index: sha c2f9a06c... -> 027c393b...
--- PASS: TestStatusDoesNotRewriteIndex (3.84s)
EXIT=0
```

### 뮤턴트 — `.moai/reports/t641/mutant.txt`, `mutant-diff.txt`, `mutant-restore.txt`

`mutant-diff.txt` 는 `"--no-optional-locks",` 인자 하나만 지운 diff 다. 결과:

```
control: plain status rewrote .git/index: sha 70ef27b9... -> 78b51e43...
Status() rewrote .git/index: sha 70d083bd... -> ad80cf5c..., mtime ...22:10:28.986151507 -> ...22:10:30.294592046
--- FAIL: TestStatusDoesNotRewriteIndex (3.24s)
EXIT=1
```

복원 후 `git diff --stat` 출력은 비었고(`EXIT=0` 만 남음), 작업 트리의 `manager.go` 해시와 `HEAD:internal/core/git/manager.go` 가 모두 `75503e276f98560d000e63a7c569bf7f45d437cf` 로 같다.

### 테스트 · 린트 · vet

| 명령 | 결과 | 파일 |
|---|---|---|
| `go test ./internal/core/... -count=1 -v` | `--- PASS` 481줄, `--- FAIL` 0줄, `EXIT=0`. `TestStatus_Clean`, `TestStatus_Dirty`, `TestStatus_NoUpstream`, `TestStatus_AheadBehind`, `TestIsClean_True`, `TestIsClean_False`, `TestStatusBranchHeaderShapes`, `TestParseStatusBranchHeader`, `TestStatusCategorization`, `TestStatusAheadBehindFromHeader`, `TestStatusDoesNotRewriteIndex` 가 실행되어 통과 | `test-core.txt` |
| `go test ./internal/statusline/ -count=1 -v` | `--- PASS` 779줄, `--- FAIL` 0줄, `EXIT=0` | `test-statusline.txt` |
| `gofmt -l internal/core/` | 출력 없음, `EXIT=0` (대조: 일부러 어긋나게 쓴 파일은 목록에 나왔다) | `gofmt.txt` |
| `go vet ./internal/core/...` | `EXIT=0` | `vet.txt` |
| `GOOS=windows go vet ./internal/core/...` | `EXIT=0` (테스트 파일까지 windows 로 컴파일) | `vet-windows.txt` |
| `golangci-lint run ./internal/core/...` | `0 issues.`, `EXIT=0` | `lint.txt` |

패키지 경로는 배차문의 `./internal/core/git/` 대신 `./internal/core/...` 를 썼다. 워크트리 가드가 `git` 이라는 경로 조각을 담은 go 명령줄을 거부했기 때문이다. `internal/core` 아래 테스트가 있는 패키지는 `git`, `project`, `quality` 셋이고(`integration`, `migration` 은 `.gitkeep` 뿐), 셋 모두 실행됐다. `-run` 을 쓴 RED/GREEN/뮤턴트 실행에서 `project`, `quality` 는 `[no tests to run]` 으로 나오며, 판정 대상 테스트는 `-v` 출력에서 이름으로 확인했다.

### `isWorkingTreeClean` 호출자 확인 — `.moai/reports/t641/isworkingtreeclean-callers.txt`

```
internal/core/git/manager.go:461:// isWorkingTreeClean is a package-level helper ...
internal/core/git/manager.go:462:func isWorkingTreeClean(ctx context.Context, dir string) (bool, error) {
internal/core/git/branch.go:75:	clean, err := isWorkingTreeClean(ctx, b.root)
EXIT=0
```

유일한 호출자는 `branchManager.Switch` 의 브랜치 전환 전 깨끗함 검사다.

## Baseline-attribution

- 재현 수치, `accuracy.json`, `index-rewrite-probe.txt` 는 baseline 커밋 `3f7a92a17` 에 들어 있는 기록이다. 이번 세션은 그 기록을 읽었을 뿐 재현 스크립트를 다시 돌리지 않았다.
- RED 는 `3f7a92a17` 의 `manager.go`(플래그 없음)에 새 테스트만 더한 트리에서, GREEN·전체 테스트·gofmt·vet·lint 는 blob `75503e27…` 의 `manager.go` 가 들어간 트리에서 재었다. 전체 테스트와 vet·lint 는 GREEN 커밋 직전에 실행했지만, 그 사이 `manager.go` 편집은 없었고 커밋된 blob 과 해시가 같다.
- 뮤턴트는 GREEN 커밋 `5b42b3301` 위에서 인자 하나를 지운 작업 트리로 재었고, 복원 후 blob 동일성을 다시 확인했다.
- 환경: `MOAI_*`, `CLAUDE_PROJECT_DIR`, 칸반 변수를 한 명령 안에서 `unset` 한 뒤 실행. 셸의 `GIT_*` 변수는 `GIT_EDITOR` 하나였고 `GIT_OPTIONAL_LOCKS` 는 없었다. git 은 `2.50.1 (Apple Git-155)`, macOS.

## Gaps

- **알려진 비용은 재지 않았다.** optional lock 을 끄면 status 는 새로 계산한 stat 정보를 인덱스에 저장하지 않는다. 그래서 다른 쓰기(add, commit, 플래그 없는 status 등)가 인덱스를 갱신하기 전까지는 다음 status 가 같은 파일을 다시 stat 하거나 해시하는 작업을 더 할 수 있다. 출력은 같다 — `accuracy.json` 에서 두 형식의 porcelain 출력(수정·스테이징·미추적 파일, ahead 1 behind 1)이 바이트 단위로 같았다(`byte_equal: true`). 추가 작업량이나 지연 시간은 측정하지 않았다.
- **실행 중인 세션에는 아직 효과가 없다.** statusline 은 설치된 `moai` 바이너리가 그린다. 이 수정은 배치 끝에 빌드하고 설치하기 전까지 살아 있는 세션의 동작을 바꾸지 않는다. 설치 후 실제 세션에서 락 충돌이 사라졌는지는 관측하지 않았다.
- **수정된 코드로 동시 쓰기 재현을 다시 돌리지 않았다.** baseline 재현은 git CLI 두 형식을 비교한 것이고, 새 Go 테스트는 인덱스 재작성이라는 부수 효과만 고정한다. `Status()` 를 동시 루프로 돌리며 add/commit 실패 횟수를 세는 측정은 하지 않았다.
- **windows 와 linux 에서는 실행하지 않았다.** windows 는 `go vet` 으로 컴파일만 확인했다. 이 전역 옵션을 받아들이는 git 의 최소 버전은 이번에 확인하지 않았고, 각 플랫폼의 git 버전에서 테스트를 돌린 것도 아니다(로컬 확인은 `2.50.1` 하나). 전체 판정은 CI 몫이다.
- **다른 레인의 index.lock 관측은 근거로 쓰지 않는다.** 오늘 다른 곳에서 기록된 관측은 다음과 같다. 이 판정서는 그 원인을 이 메커니즘으로 귀속하지 않으며, 증명으로 인용하지도 않는다.
  - `internal/kanban/settings_drift.go` 머리 주석 1항: 플래그 없는 `git status` 가 인덱스 쓰기 락을 수십 밀리초 잡는다는 이전 카드의 측정 기록
  - 프로젝트 메모리 색인의 "statusline status 가 인덱스 쓰기 락", "index lock 2차 사례", "index.lock 실패는 경로 유무가 판별식" 항목
  - 리드 배차문이 언급한 오늘 다른 레인들의 index.lock 관측(이 세션에서 원본을 읽지 않았다)

## Residual-risk

- **`isWorkingTreeClean` 은 여전히 플래그 없이 돈다.** 호출자가 브랜치 전환 직전의 검사 한 곳뿐이고 statusline 경로가 아니라서 범위 밖으로 두었다. 이 경로는 매 렌더마다 도는 statusline 과 달리 브랜치 전환 때만 실행되지만, 그 순간 같은 워크트리에서 다른 세션이 쓰고 있으면 같은 종류의 충돌이 날 수 있다. 이 경로의 충돌 빈도는 재지 않았다.
- **git 오류 메시지의 명령 이름이 바뀐다.** `execGit` 은 `CommandError.Op` 를 첫 인자로 채우므로(`firstArg`), `Status()` 가 실패하면 메시지가 `status: git status: ...` 가 아니라 `status: git --no-optional-locks: ...` 로 나온다. 바깥의 `status:` 접두어는 남아 있어 어떤 동작이 실패했는지는 읽을 수 있고, 이 문구에 기대는 테스트는 없었다(전체 테스트 통과). `firstArg` 는 공용 헬퍼라 이번 범위에서 고치지 않았다.
- **새 테스트는 파일시스템 시간 해상도에 기댄다.** 1.1초씩 두 번 기다려 racy-git 판정을 피하고, stat 정보가 낡았는지는 대조군으로 매번 확인한다. 대조군이 인덱스를 다시 쓰지 않는 환경에서는 테스트가 공허하게 통과하지 않고 `t.Fatalf` 로 멈춘다. 대신 그런 환경에서는 거짓 실패가 날 수 있다.
- **`Status()` 의 다른 소비자.** 제품 코드 호출자는 `internal/statusline/git.go:37` 과 같은 파일의 `IsClean()` 이다(`.Status()` 비테스트 grep 기준). 동작 차이는 인덱스 쓰기가 없어진다는 것 하나이며, 인덱스 갱신에 기대는 소비자는 찾지 못했다.
