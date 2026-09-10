# t556 — 병합 트리 재측정 기록 (미완결)

브랜치 `WT-goal-cond-classify` · 흡수 후 HEAD `9f89296a9` · 흡수 대상 **로컬 develop `d60b2956b`**

배차문은 "origin/develop 흡수" 였으나 로컬 develop 을 흡수했다. 병합 대상이 로컬 develop 이고 그쪽이 origin 보다 4 앞서 있어, origin 을 흡수하면 실제로 착지할 트리가 아닌 것을 재게 된다. 리드가 이 정정을 접수했다.

## 통과한 축 (병합 트리에서 측정)

| 검사 | 결과 |
|---|---|
| `go test ./internal/goal/... -count=1` | `ok … 1.292s` |
| `go vet ./internal/cli/` | exit 0 |
| `gofmt -l internal/cli/` | 출력 없음 |
| `golangci-lint run ./internal/cli/...` | `0 issues.` |
| `GOOS=windows GOARCH=amd64 go build ./...` | exit 0 |
| `make build` | exit 0 · `git status --short` 무출력(catalog 드리프트 없음) |

## `internal/cli` — 미완결. 완주한 실행이 없다

| 회차 | 명령 | 관측 |
|---|---|---|
| 1 | `go test ./internal/cli/... -count=1 \| grep -v "^ok" \| head -30` | **판정 없음** — 아래 계측 결함 참조 |
| 2 | 같은 명령(흡수 후) | `FAIL … 603.707s` (기본 600초 문턱) |
| 3 | `go test ./internal/cli/ -count=1 -timeout 20m` (전체 출력 보존) | `FAIL … 1201.266s` (20분 문턱) |

### [정정] 1회차는 green baseline 이 아니었다

앞선 보고에서 "흡수 전 같은 명령은 통과했다(exit 0, FAIL 줄 없음)" 라고 적었다. **거짓이다.** 파이프의 `head -30` 이 판정 줄이 나오기 전에 스트림을 잘랐고, 출력 파일 33줄에 `ok` / `FAIL` / `--- FAIL` 이 **0건**이다. 읽은 exit 0 은 go test 가 아니라 `head` 의 종료 코드였다.

그러므로 이 카드는 `internal/cli` 의 green baseline 을 가진 적이 없다. "흡수 후에 빨개졌다" 는 인상은 존재하지 않는 baseline 을 근거로 만들어진 것이며, 계측 결함의 산물이다.

계측 결함은 두 번 반복됐다 — 1회차 `head -30`(앞을 자름), 2회차 `tail -20`(panic 헤더와 걸린 테스트 이름을 버림). 3회차에서 전체 출력을 파일로 보존하고서야 진단이 가능해졌다.

### 3회차 타임아웃의 성질 — 정지가 아니라 예산 소진

```
panic: test timed out after 20m0s
	running tests:
		TestTodoExportJSON_FailurePathsSurface (0s)
		TestTodoExportJSON_FailurePathsSurface/unwritable_queue_directory_is_reported,_not_swallowed (0s)
```

알람 시점에 돌고 있던 테스트는 하나뿐이고 경과가 **(0s)** 다. 그 테스트가 멈춘 것이 아니라, 스위트 전체가 20분 예산을 다 쓰고 그 시점에 막 시작한 테스트가 찍혔다. 해당 테스트는 t306(`a8099e43e`)에서 들어온 기존 것이다.

### 실행 조건 오염 — 동시 실행

3회차의 마지막 약 13분 동안 같은 패키지의 전체 스위트가 **하나 더** 돌고 있었다(lane-7, `go test ./internal/cli/ -timeout 3600s`, pid 69925). 리드 실측 + 본 세션 `pgrep` 으로 양쪽에서 확인했다. 측정 중 실측 부하: `load average 21.86 / 15.01 / 15.49` → 이후 `11.17`.

리드 통지 누락으로 lane-7 에게 직렬화가 전달되지 않아 발생했다. 판독 규칙(리드 지정, lane-7 보강): **경합은 타임아웃과 플레이크를 만들지 거짓 통과를 만들지 않는다.** 따라서 비대칭이다 — 초록은 오염된 조건에서도 유효하고, 빨강만 판정이 아니다.

**이 수치를 인용하는 사람은 이 동시 실행 조건을 함께 읽어야 한다.**

## 어설션 실패 4건 — 타임아웃과 별개 축, 전부 이 카드의 파일이 아님

| 테스트 | 관측 |
|---|---|
| `TestHomeStateChangedSurfaceCoverageConsumesFreshProfile` | `audited production file changed after coverage tip: internal/cli/launcher.go` |
| `TestChangedProductionFilesDerivesCurrentHeadDiffAndPlatformDisposition` | 같은 에러 |
| `TestAuditLagUsesBinlagSeam` | `home_state_coverage.go:243/251 MISSING`; `245/253` REQ-ABI-006 위반 |
| `TestHomeStateChangedSurfaceCoverageRunsBoundedFocusedSuite` | 자식 `go test` 가 `exit status 1` — 위 셋의 연쇄 |

### 보고되는 파일명은 무작위다 — 귀속 근거로 쓸 수 없다

3회차 로그에서는 앞의 둘이 `internal/cli/mcp_server.go`(이 카드가 바꾼 파일)를 지목했고, 직후 격리 재실행에서는 `internal/cli/launcher.go`(건드리지 않은 파일)를 지목했다. 에러는 `home_state_coverage.go:298` 에서 `expectedBlobs` **map 순회**의 첫 적중을 뱉으므로 파일명이 실행마다 바뀐다.

"내 파일이 지목됐다" 도, "내 파일이 지목되지 않았다" 도 귀속 근거가 되지 못한다.

### 귀속 — blob 동일성으로부터의 유도 (직접 측정 아님)

지목된 파일 전부가 `develop` 과 이 브랜치 HEAD 사이에서 blob 동일이다:

```
$ git rev-parse HEAD:<file> develop:<file>
internal/cli/launcher.go            6fb3ac394d9d47c10207bde63d2aa3a8bb8c992d  (양쪽 동일)
internal/cli/home_state_coverage.go 7de85890c954af1ae9a56dcafe886251c8168a94  (양쪽 동일)
internal/cli/mcp_build_identity_test.go 9b7c853c8474e217dfda91808b57c0fb4b52d419  (양쪽 동일)
```

이 검사들의 입력은 evidence-chain marker 커밋의 blob 과 HEAD blob 이고, 이 카드의 커밋은 marker 를 추가하지 않는다. 따라서 이 파일들에 대한 판정은 develop 과 이 브랜치에서 동일하며, **develop 자체가 이미 이 셋에서 빨갛다**는 결론이 따른다.

[HARD] 이것은 develop 트리에서 직접 돌려 본 **측정이 아니라 유도**다. 그렇게 읽어야 한다.

보강: `launcher.go` 를 마지막으로 바꾼 `6010d5d82`(#1697, t593)는 이미 develop 조상이다(`git merge-base --is-ancestor 6010d5d82 develop` exit 0). 이 계열은 이 배치 이전에 착지했다.

**독립 확인 (다른 경로, 다른 레인).** lane-2 가 t554 창에서 자기 수리를 되돌린 베이스라인에서 같은 4건이 **바이트 동일한 메시지로** 그대로 실패하는 것을 측정으로 세웠다(리드 전달). 서로 다른 두 경로 — 이쪽은 blob 동일성 유도, 저쪽은 되돌린 베이스라인 측정 — 가 같은 4건을 가리키므로 귀속이 선다. 이 카드는 develop 트리 직접 측정 순번을 따로 쓰지 않는다.

## [HARD] 배치 기준선 — `internal/cli` 의 기지(旣知) 적색

리드가 배치 정책으로 고정했다. develop 의 `internal/cli` 는 아래 4건에서 이미 빨갛다:

```
TestHomeStateChangedSurfaceCoverageConsumesFreshProfile
TestChangedProductionFilesDerivesCurrentHeadDiffAndPlatformDisposition
TestAuditLagUsesBinlagSeam
TestHomeStateChangedSurfaceCoverageRunsBoundedFocusedSuite   (위 셋의 연쇄)
```

레인의 병합 트리 재측정 통과 기준: 실패 집합이 위 4건의 **부분집합**이고, 자기 변경이 추가한 테스트가 통과할 것. **5번째 실패가 나오면 그 레인은 병합하지 않는다.**

기준선을 고정하는 이유는 하나다 — 상시 적색인 검사는 새 적색과 기존 적색을 구분하지 못한다. 이 4건의 소관은 카드 **t600** 이다.

## 이 카드의 변경분

제 테스트 3건은 격리 실행에서 매번 통과한다(`ok … 1.2s`), 뮤턴트 가드로 실제로 코드에 걸려 있음도 확인했다(prose 분기 비활성화 시 CLI·MCP 양쪽 RED).

다만 이 변경이 `internal/cli/mcp_server.go` 를 audited 집합에 하나 더 넣는 것은 사실이다. 지금은 `launcher.go` 가 같은 검사를 무조건 실패시켜 red→red 라 판정을 바꾸지 않지만, launcher 쪽이 정리되면 이 카드의 파일이 드러난다. → 잔여 위험.

## Gaps — 관측하지 않은 것

- ~~`internal/cli` 를 완주한 실행이 한 번도 없다~~ → **아래 4회차에서 닫혔다.** 1~3회차 시점의 기록으로 남긴다.
- develop 트리에서 위 4건을 직접 돌려 보지 않았다(위 귀속은 유도). 독립 확인은 lane-2 의 되돌린-베이스라인 측정이 맡았다.
- 전체 스위트(`go test ./...`)는 로컬에서 돌리지 않았다(CLAUDE.local.md §4). CI 몫이다.
- `internal/hook/stop-goal` 평가자 경로는 건드리지 않았고 실행하지도 않았다. 이 카드는 arm 시점만 바꾼다.

## 4회차 — 완주. 판정이 선다

```
$ go test ./internal/cli/ -count=1 -timeout 60m > cli-final.log 2>&1
$ echo "GOTEST_EXIT=$?" >> cli-final.log

FAIL	github.com/modu-ai/moai-adk/internal/cli	1125.213s
FAIL
GOTEST_EXIT=1
```

**타임아웃 없이 완주했다** (1125s < 60m). 앞선 3회차까지의 미측정 구간이 닫혔다.

`--- FAIL` 계수 **4**, 실패 집합:

```
--- FAIL: TestHomeStateChangedSurfaceCoverageConsumesFreshProfile (3.19s)
--- FAIL: TestHomeStateChangedSurfaceCoverageRunsBoundedFocusedSuite (120.89s)
--- FAIL: TestChangedProductionFilesDerivesCurrentHeadDiffAndPlatformDisposition (2.99s)
--- FAIL: TestAuditLagUsesBinlagSeam (0.78s)
```

**기지 적색 4건과의 차집합 = 공집합.** 5번째 실패 없음 → 배치 정책의 레인 통과 기준을 충족한다.

### 이 카드의 테스트 3건 — 통과 (명시 증거)

```
$ go test ./internal/cli/ -count=1 -v -run 'TestProseShapedCommand|TestGoalArm_ProseShape|TestMCPGoalArm_ProseShape'

--- PASS: TestGoalArm_ProseShapedConditionRefused (0.01s)
--- PASS: TestGoalArm_ProseShapeExemptions (0.01s)
        (cmd-prefix / model-prefix / real-command 3 서브테스트)
--- PASS: TestMCPGoalArm_ProseShapedConditionRefused (0.01s)
--- PASS: TestProseShapedCommand (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	1.032s
EXIT=0
```

### 4회차의 실행 조건 — 부분 오염, 그러나 판정은 선다

실행 시작 시점은 조용했다(실측: `internal/cli` 실행 0개, load 5.43). 시작 76초 후 lane-7 이 같은 패키지에 진입했다:

```
34958  etime 01:16  go test ./internal/cli/ -count=1 -timeout 60m   ← 이 카드
39387  etime 00:18  go test ./internal/cli/ -run 'TestHomeState…|TestChanged…|TestAuditLag…' -v -timeout 900s
39906  etime 00:08  go test ./internal/cli -run ^(TestHomeState.*|…)$   ← 39387 의 자식
```

부모 체인으로 귀속: 39387 의 조부는 pid 14968 = lane-7 세션(`3e7d359d-…`). 자식 39906 포함 둘 다 lane-7 소유다. 앞선 3회차의 오염(전체 스위트 하나 더)보다 가볍고, load 는 8.8 수준에 머물렀다.

판독 규칙(리드 지정, lane-7 보강): **경합은 타임아웃과 플레이크를 만들지 거짓 통과를 만들지 않는다.** 4회차는 완주했고 실패 집합이 기지 4건의 부분집합이므로, 오염 하에서도 이 결과는 유효하다 — 오히려 더 강한 근거다.

lane-7 의 그 실행은 사후적으로 불필요한 것으로 확인됐다(lane-7 자신의 완주 실행이 이미 기지 4건과 일치했고, "대조군 필요 없음" 메시지가 엇갈렸다). 리드가 중단을 지시했다.

## 계측 결함 세 건 — 모두 조용히 틀린 값을 돌려줬다

| 회차 | 형태 | 잃은 것 |
|---|---|---|
| 1 | `\| head -30` | 판정 줄(`ok`/`FAIL`)을 **앞에서** 자름 → 존재하지 않는 green baseline |
| 2 | `\| tail -20` | panic 헤더와 `running tests:` 를 **뒤에서** 자름 → 걸린 테스트 이름 소실 |
| — | `pgrep -f 'go test ./internal/cli'` | **자기 명령줄을 셈** → 카운트가 구조적으로 0 이 될 수 없음 |

셋의 공통 성질: **에러를 내지 않고 그럴듯한 값을 돌려준다.** 앞의 둘은 "파이프 금지 + 파일 리다이렉트" 계약으로 걸리지만, 셋째는 걸리지 않는다 — 프로세스 계수는 패턴이 아니라 **실행 파일 정체**로 걸러야 한다(`ps -eo comm,args | awk '$1=="go"'`). 리드가 배차문 계약에 별도 항으로 넣기로 했다.

## 2차 흡수 (`c9a9e8866`) — 위 판정을 무효화한다

위 4회차 완주는 develop `d60b2956b` 를 흡수한 트리에서 잰 것이다. 그 뒤 develop 이 두 번 움직였고(`c8203fbf3` t550 → `c9a9e8866` t555), 통합 창 안에서 최신 tip 을 흡수했다.

```
흡수 전 HEAD 9f89296a9  →  흡수 후 HEAD 0e47f8ca7   (develop c9a9e8866, develop 대비 3 앞섬)

$ git diff --stat 9f89296a9..HEAD -- internal/cli/
 internal/cli/todo.go                | 151 +++++++++++++++++++++++++++++---
 internal/cli/todo_verb_leak_test.go | 168 ++++++++++++++++++++++++++++++++++++
 2 files changed, 305 insertions(+), 14 deletions(-)

CLI_FILES_IN_DELTA=2
```

### 사전 판독은 낡았다 — 예상으로 표시해 둔 것이 맞았다

창 대기 중 `c8203fbf3` 기준으로 미리 읽은 값은 `CLI_FILES_IN_DELTA = 0` 이었다. 그 사이 lane-7 의 t555 가 들어와 tip 이 `c9a9e8866` 이 되면서 **0 → 2** 로 바뀌었다. 흡수 전 읽기는 근거가 아니라 예상이며, 인용은 흡수 후 실측으로만 한다.

### "파일이 델타에 있다" ≠ "판정이 뒤집힌다" — 두 델타의 성질은 반대다

| | t550 신규 테스트 | t555 델타 |
|---|---|---|
| 위치 | `internal/hook/quality/gate_oxlint_lint_test.go` — **다른 패키지** | `internal/cli/todo.go` 수정 + `internal/cli/todo_verb_leak_test.go` 신규 — **같은 패키지** |
| `internal/cli` 실패 집합에 5번째를 더할 수 있는가 | **원리적으로 불가능** | **구조적으로 가능** (신규 테스트가 그 자체로 5번째가 될 수 있고, `todo.go` 수정이 기존 테스트를 깨뜨릴 수 있다) |
| 결론 | 재측정 불필요가 **판정** | 재봐야 **알 수 있다** |

lane-7 이 자기 병합 트리에서 초록을 봤을 가능성은 높다. 그러나 그것은 lane-7 의 트리이지 이 카드의 병합 트리가 아니다 — **두 변경이 각각 초록인 것과 합쳐서 초록인 것은 다른 주장이고, 이 배치가 카드별로 검증해 병합하는 구조라 정확히 그 구멍이 생긴다.** 재측정은 그 구멍을 메우는 것이다.

따라서 위 4회차의 "차집합 공집합" 판정은 **이 tip 에서는 성립이 확인되지 않았다.** 재측정 대기 중.

## 5회차 — `c9a9e8866` 흡수본 완주. 병합 대상 트리의 판정이 선다

### 측정 대상과 흡수 델타

```
HEAD      0e47f8ca7f84f062279d528a882fbe049045a3b6   (c9a9e8866 흡수본 — t555 의 internal/cli 적중 2 포함)
develop   d3b7d438d   (측정 시점 tip)
worktree  ?? .moai/reports/t556/remeasure.md 만 미추적, 코드 무변경
```

측정 시점 develop 은 흡수한 `c9a9e8866` 보다 더 나아간 `d3b7d438d`(t576)였다. 흡수 없이 잰 근거:

```
$ git diff --name-only c9a9e8866..d3b7d438d -- internal/cli/
internal/cli/update/merge/conflict_blind_breadth_test.go
internal/cli/update/merge/conflict_blind_repro_test.go
EXIT=0
```

경로 적중은 2 이지만 둘 다 (a) `internal/cli/update/merge` — **다른 Go 패키지**이고 (b) `_test.go` 다. 다른 패키지의 테스트 파일은 `./internal/cli/` 테스트 바이너리에 컴파일되지 않고, 테스트가 아닌 파일은 0 이라 의존 코드 변화도 없다. **본체 패키지 적중 0.** "경로상 `internal/cli/` 아래" 와 "`internal/cli` 패키지" 는 다르다.

병합 창에서 tip 을 흡수한 뒤 서브트리 해시와 `go list -deps` 로 이월을 보인다(리드 지시).

### 결과

```
$ go test ./internal/cli/ -count=1 -timeout 60m > cli-final2.log 2>&1
$ echo "GOTEST_EXIT=$?" >> cli-final2.log

FAIL	github.com/modu-ai/moai-adk/internal/cli	972.536s
GOTEST_EXIT=1

$ grep -c "test timed out" cli-final2.log   → 0
$ grep -c "^panic:" cli-final2.log          → 0
```

**타임아웃·panic 없이 완주.** `--- FAIL` 계수 **4**:

```
--- FAIL: TestHomeStateChangedSurfaceCoverageConsumesFreshProfile (2.86s)
--- FAIL: TestHomeStateChangedSurfaceCoverageRunsBoundedFocusedSuite (113.10s)
--- FAIL: TestChangedProductionFilesDerivesCurrentHeadDiffAndPlatformDisposition (3.49s)
--- FAIL: TestAuditLagUsesBinlagSeam (0.77s)
```

**기지 적색 4건과의 차집합 = 공집합.** t555 가 들인 `internal/cli/todo.go` 수정과 `todo_verb_leak_test.go` 신규 테스트는 5번째 실패를 만들지 않았다 — 구조적으로 가능했던 것이 실제로는 일어나지 않았음을 이 트리에서 측정으로 확인했다.

### 이 카드의 테스트 3건 — 이 트리에서 통과

```
$ go test ./internal/cli/ -count=1 -v -run 'TestProseShapedCommand|TestGoalArm_ProseShape|TestMCPGoalArm_ProseShape'
--- PASS: TestGoalArm_ProseShapedConditionRefused (0.00s)
--- PASS: TestGoalArm_ProseShapeExemptions (0.01s)
--- PASS: TestMCPGoalArm_ProseShapedConditionRefused (0.01s)
--- PASS: TestProseShapedCommand (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	1.045s
EXIT=0
```

`--- PASS` 계수 4(최상위 테스트 4개 — 3건 중 하나가 서브테스트 3개를 가진 구조), 셀렉터 0매치 아님.

### 실행 조건 — 부분 오염, 판정은 선다

시작 조건: `internal/cli` 실행 0, load 7.48. 1분 간격 샘플러(v2)가 기록한 동시 `go test`:

| 시각 | 겹친 실행 | load (1분) |
|---|---|---|
| 15:33–15:34 | `internal/kanban/` | 13.31 |
| 15:34 | `internal/cli/ -run ^(TestCodexSkillPathDotDot…)$` — **같은 패키지 스코프 실행** | **22.57** |
| 15:35 | `internal/web/` | 19.04 |
| 15:36–15:38 | `internal/cli/ -run Todo -timeout 30m` — **같은 패키지 스코프 실행, 약 3분** | 21.37 |
| 15:38–15:39 | `internal/cli -run ^(TestHomeState.*|…)$ -coverpkg=… -coverprofile=…` · `internal/session/...` | 8.95 |
| 15:40–종료 | 겹침 없음 | 6.67–8.62 |

15:38–15:39 의 `-coverpkg … -coverprofile=/var/folders/…/moai-home-state-coverage-*.out.cli` 실행은 이 카드 실행 안의 `TestHomeStateChangedSurfaceCoverageRunsBoundedFocusedSuite` 가 띄우는 자식과 명령 형태가 같고, 그 테스트의 소요(113.10s)와 관측 구간도 겹친다. **그러나 부모 pid 를 기록하지 않았으므로 이 카드의 자식이라고 귀속하지 않는다** — 가능성이 높다는 데서 멈춘다.

판독 규칙(리드 지정): 경합은 거짓 통과를 만들지 않는다. 5회차는 완주했고 실패 집합이 기지 4건과 같으므로 오염 하에서도 유효하다.

### 계기 결함 넷째 — 필드 번호가 어긋난 필터

겹침 샘플러 v1 은 `ps -eo pid,etime,comm,args | awk '$3=="go" && $4=="test"'` 로 걸렀다. 실측한 필드 배치는 `$3=comm(go)`, `$4=go`, `$5=test` 였다 — **`$4=="test"` 는 구조적으로 참이 될 수 없다.** 에러 없이 빈 기록을 남기므로 그대로 두었다면 실행 내내 "동시 실행 없음" 이라는 거짓 증거가 쌓였을 것이다.

잡은 방법: 첫 샘플에 **이 카드 자신의 실행조차 없었다.** 루프 조건은 그 실행을 보고 있었으므로(그래서 끝나지 않았음) 기록 필터 쪽이 틀린 것이다. 수정한 v2 는 먼저 **자기 실행(pid 3767)이 잡히는지** 확인한 뒤에야 신뢰했다 — 자기 자신을 잡아야 "다른 줄이 없음" 이 진짜 부재를 뜻한다.

v1 가동 구간(약 15:31:35–15:32:24)은 샘플이 무효이며, 그 사이 수동 관측(실행 경과 01:09 시점)은 이 카드 실행 1개뿐이었다.

## 결론 (4회차 시점 — `d60b2956b` 흡수본 기준)

`internal/cli` **완주** · `--- FAIL` **4** · 기지 적색 4건과의 **차집합 공집합** · 이 카드의 테스트 3건 **전부 통과**. 그 트리에서는 배치 정책의 레인 통과 기준을 충족했다.

[HARD] 이 판정은 `d60b2956b` 흡수본에 귀속된다. 병합 대상 트리(`c9a9e8866` 흡수본, HEAD `0e47f8ca7`)의 판정은 재측정 결과로 갱신한다.

남은 절차는 통합 창 재요청 → develop 워크트리 진입 → HEAD 재판독 → `--no-ff` 병합 → release → 로컬 병합 SHA 보고다. push 는 리드 소관이다.

기지 4건 자체의 소관은 카드 **t600** 이며 이 카드의 범위가 아니다. 실행 순번 표면의 부재는 카드 **t607** 로 발행됐다.
