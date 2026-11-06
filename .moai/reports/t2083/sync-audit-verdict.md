auditor-model: claude-opus-5-5

# t2083 sync 감사 판정 — TestAgenticTrunkStatuslineBuilderExecCount 콜드 메모 수리

verdict: FAIL
audited_sha: dd5d0674e13d168954f93bb3ef8d53085ce6541c

- 감사 대상: `dd5d0674e1` (부모 `9b632c1eb3`), 변경 파일 `internal/statusline/agentic_trunk_test.go` 1개 (+18 / -5)
- 감사자: sync-auditor 단독, 1회차 (반복 이력 없음)
- 내용 판정(merits): **PASS, 점수 0.94** — 원인 주장 재현 확인, 기대값 약화 없음, 제품 결함 은폐 없음
- 최종 판정: **FAIL (게이트 미충족, fail-closed)** — 이 트리의 감사 계획이 codex를 required로 지정했는데, 배차 지시("no paid or cross-model calls")에 따라 codex 감사를 돌리지 않았다. 필수 게이트를 돌리지 않은 상태로 PASS를 낼 수 없으므로 FAIL로 둔다. 코드에서 찾은 blocking 결함은 없다.

## 차원별 점수

| 차원 | 점수 | 판정 | 근거 |
|---|---|---|---|
| Functionality (40%) | 95/100 | PASS | 수리 전 테스트를 심볼릭 링크 없는 TMPDIR에서 돌리면 CI 실패가 그대로 재현되고(`ran 3 ... want exactly 4`), 수리 후에는 통과. 기대값 1·4 그대로 |
| Security (25%) | 100/100 | PASS | 테스트 전용 변경. 새 입력 경로·비밀값·외부 호출 없음. 기존 셸 shim 패턴 재사용 |
| Craft (20%) | 90/100 | PASS | `go vet` 0, `gofmt -l` 출력 없음, 패키지 전체 시험 통과(두 TMPDIR 모두). 커버리지는 측정하지 않음(Gaps) |
| Consistency (15%) | 92/100 | PASS | 같은 파일의 픽스처·헬퍼(`statuslineGitStrategyFixture`, `t.Helper`)를 그대로 따름. 커밋 타입만 `fix`(테스트 전용이면 `test`가 더 정확, 선택 사항) |

필수 통과 차원(Functionality·Security)은 모두 기준을 넘는다. FAIL 사유는 차원 점수가 아니라 필수 교차 모델 게이트다.

## 결함 목록

- **F1** [High] [blocking] [게이트, 위치 무관] 감사 계획의 필수 백엔드 codex 미실행.
  - 근거: `moai verify audit-plan` 출력 `"cross_model_required": true`, `"enforced_required": ["codex"]`.
  - 실행하지 않은 이유: 배차 지시가 교차 모델·유료 호출을 금지했다.
  - Required fix: 리더가 `mcp__moai__codex_audit`(`target: baseBranch`, `project_root` = 이 워크트리)를 돌려 receipt를 받거나, 운영자가 이 카드의 codex 게이트 면제를 기록한다. codex 결과에 blocking 지적이 없으면 내용 판정(PASS 0.94)이 그대로 최종 판정이 된다.
- **F2** [Low] [optional] [touched-file] 메모가 워밍된 상태의 동작(같은 경로를 두 번째로 수집할 때 큐 루트 프로브를 건너뛰어 3회)을 고정하는 statusline 테스트는 이제 없다.
  - 메모 적중 자체는 `internal/homestate/location_memo_test.go`의 `TestLocationMemo_RepeatedLookupsSpawnNoGit`가 검증하므로 공백은 아니다. 추가는 재량이다.
- **F3** [Low] [optional] [outside] 메모 키가 경로 표기 그대로라 같은 디렉터리를 다른 표기로 부르면 적중하지 않는다(macOS `/var/...` 대 `/private/var/...`).
  - 이 경우 git을 한 번 더 부를 뿐 답은 맞으므로 성능 문제이지 정합성 문제는 아니다. 판정에 반영하지 않는다.
- **F4** [Info] [optional] [in-diff] 커밋 제목이 `fix(statusline)`인데 변경은 테스트 파일뿐이다. Conventional Commits로는 `test(statusline)`가 더 정확하다. 다시 쓸 필요는 없다.

## 질문별 판단

### 1. 원인 주장이 맞는가 — 맞다 (재현 확인)

**Claim**: 두 하위 테스트가 저장소 경로 하나를 공유하는 바람에 첫 하위 테스트가 homestate 위치 메모를 데웠다. 그래서 둘째 하위 테스트는 큐 루트 프로브를 건너뛰어 3회가 나왔다. macOS에서는 경로 표기 차이로 메모를 우연히 빗나가 문제가 가려졌다.

**Evidence**:

수리 전 파일은 `git show 9b632c1eb3:internal/statusline/agentic_trunk_test.go`로 스크래치에 꺼냈다. 작업 트리는 건드리지 않고 `go test -overlay`로 그 파일을 끼워 넣었다. 심볼릭 링크 없는 TMPDIR = `/private/tmp/claude-501/.../scratchpad/tmpd`(저장소와 `$HOME` 바깥).

(a) 수리 전 + 심볼릭 링크 없는 TMPDIR → `exit=1`
```
TMPDIR=<scratch>/tmpd timeout 600 go test -count=1 -overlay=<scratch>/overlay.json -run TestAgenticTrunkStatuslineBuilderExecCount -v ./internal/statusline/
    agentic_trunk_test.go:351: collect pass ran 3 git resolution(s) for the CWD-only payload, want exactly 4 (...)
    --- PASS: TestAgenticTrunkStatuslineBuilderExecCount/anchor_from_stdin:_stateanchor_consumers_add_no_walk-up (0.10s)
    --- FAIL: TestAgenticTrunkStatuslineBuilderExecCount/anchor_from_the_git_walk-up:_goal_and_label_share_one_resolution (0.16s)
FAIL	github.com/modu-ai/moai-adk/internal/statusline	0.980s
```
(b) 수리 전 + 같은 TMPDIR + 둘째 하위 테스트만 실행 → `exit=0`. 하위 테스트 사이의 상호작용이라는 증거다.
```
go test ... -overlay=<scratch>/overlay.json -run 'TestAgenticTrunkStatuslineBuilderExecCount/anchor_from_the_git_walk-up' -v ./internal/statusline/
    --- PASS: TestAgenticTrunkStatuslineBuilderExecCount/anchor_from_the_git_walk-up:_goal_and_label_share_one_resolution (0.29s)
ok  	github.com/modu-ai/moai-adk/internal/statusline	0.916s
```
(c) 수리 전 + 기본 macOS TMPDIR(`/var/folders/kt/.../T/`) → `exit=0`. 로컬에서 문제가 가려진 경위와 같다.
```
    --- PASS: TestAgenticTrunkStatuslineBuilderExecCount/anchor_from_the_git_walk-up:_goal_and_label_share_one_resolution (0.64s)
ok  	github.com/modu-ai/moai-adk/internal/statusline	1.459s
```
(d) 수리 후 + 심볼릭 링크 없는 TMPDIR → `exit=0`. (e) 수리 후 + 기본 TMPDIR → `exit=0`.
```
(d) --- PASS: .../anchor_from_the_git_walk-up:_goal_and_label_share_one_resolution (0.54s)   ok ... 1.254s
(e) --- PASS: .../anchor_from_the_git_walk-up:_goal_and_label_share_one_resolution (1.13s)   ok ... 2.654s
```

**메커니즘을 코드로 대조**: `internal/statusline/backlog.go:47`의 `resolveBacklogCounts` → `factory.BacklogCountsForRoot`. `internal/homestate/paths.go:74`의 `CanonicalProjectRoot`가 `memoLocation("canonical", abs, ...)`로 감싸여 있다. `internal/homestate/location_memo.go:124`의 키는 `locationKey{op, dir, env}`이고, 지문이 같으면(`:131`) git을 부르지 않는다. 키가 경로 문자열 그대로이므로 공유 경로는 Linux에서 적중하고 macOS `/var` 대 `/private/var` 표기에서는 빗나간다. 관측한 (a)~(c) 패턴과 정확히 맞는다.

**Baseline-attribution**: 위 명령 모두 이 실행에서 이 트리(HEAD `dd5d0674e1`)에 대해 돌렸다. 수리 전 상태는 부모 커밋 `9b632c1eb3`의 테스트 파일 하나만 overlay로 바꾼 것이고, 나머지 소스는 HEAD와 같다(이 커밋이 바꾼 파일은 그 하나뿐이다).

### 2. 테스트가 약해졌는가 — 아니다

`git diff HEAD~1 HEAD`에서 바뀐 것은 저장소 생성이 `newRoot` 클로저로 옮겨 가고 각 하위 테스트가 `root := newRoot(t)`를 부르는 부분뿐이다. 판정문 `got != 1`, `got != 4`, `WorkflowLabel != "AF-Flow"`과 오류 메시지는 한 글자도 바뀌지 않았다. 클로저는 하위 테스트의 `t`를 받으므로 `t.TempDir()` 정리도 하위 테스트마다 따로 된다.

### 3. 제품 결함이 가려졌는가 — 아니다

메모가 없는 상태에서 CWD-only 수집은 큐 루트 프로브 1회와 stateanchor walk-up 3회, 합해서 4회다. 이는 수리 전 코드로 둘째 하위 테스트만 돌린 (b)가 4로 통과한 것과 일치한다. 같은 프로세스에서 같은 경로를 두 번째로 수집할 때 큐 루트 프로브가 빠지는 것은 Issue #1948 메모의 설계 목적 그대로다(`location_memo.go` 머리 주석: "keeps an answer for as long as the filesystem it was read from looks the same"). 지문이 파일 시스템 변화를 반영하므로 정합성 위험은 없다. 테스트가 콜드 메모를 암묵적으로 가정하고 있던 것이 결함이었다. 표기 별칭에서 메모가 빗나가는 것은 성능 문제일 뿐이다(F3).

### 4. 같은 가정을 숨긴 다른 테스트가 있는가 — 찾지 못했다

```
grep -rln -E 'ShimCalls|WalkUpCount|git-calls\.log|gitCallCount|countGit|spawnCount|execCount|ExecCount' internal/statusline internal/homestate --include='*_test.go'
→ internal/statusline/agentic_trunk_test.go   (이 파일 하나뿐)
```

- 같은 파일의 `TestAgenticTrunkStatuslineNoExtraGit`: 기대값이 0회다. 메모는 횟수를 줄이기만 하므로 0 판정이 가려질 수 없다.
- `internal/statusline/landed_test.go`의 `TestRefreshLandedCounts_OneInvocationRegardlessOfCardCount`: 반복마다 `t.TempDir()`를 새로 만들고 `landedGitRunner` 스텁으로 센다. 메모 경로를 거치지 않는다.
- `internal/homestate/location_memo_test.go`의 메모 시험들: 각자 `resetLocationMemo()`를 부르고 `lmDir`로 심볼릭 링크를 푼 새 경로를 쓴다.

grep 수준의 조사이므로 다른 패키지(`internal/factory` 등)는 범위 밖이다(Gaps).

### 5. `TestResolveBacklogCounts_LatencyBudget` — 이 변경과 무관하다는 데 동의

이 시험은 `internal/statusline/backlog_sqlite_test.go:187`에 있고, 자체 `t.TempDir()` 루트 두 개(`dbRoot`, `jsonRoot`)로 p95를 잰다. shim도, 바뀐 테스트의 상태도 공유하지 않는다. 공유하는 것은 프로세스 전역 메모뿐인데, 경로마다 키가 따로이고 이번 변경이 늘린 항목은 1개다. 내가 돌린 두 번의 패키지 전체 실행(부하 평균 45~65)에서는 모두 통과했다. 연결 고리를 찾지 못했으므로 카드에 불리하게 반영하지 않는다.

## 실행한 명령과 출력

| 명령 | 결과 |
|---|---|
| `TMPDIR=<scratch>/tmpd timeout 900 go test -count=1 ./internal/statusline/` | `exit=0` / `ok  	github.com/modu-ai/moai-adk/internal/statusline	72.197s` |
| `timeout 900 go test -count=1 ./internal/statusline/` (기본 TMPDIR) | `exit=0` / `ok  	github.com/modu-ai/moai-adk/internal/statusline	52.384s` |
| `go vet ./internal/statusline/` | `vet exit=0` (출력 없음) |
| `gofmt -l internal/statusline/agentic_trunk_test.go` | 출력 없음, `exit=0` |
| `moai verify audit-plan --project-root <this worktree>` | `"config_status": "ok"`, `"model": "codex"`, `"cross_model_active": true`, `"cross_model_required": true`, `"enforced_required": ["codex"]` |
| `uptime` (실행 중) | `load averages: 40.76 45.47 46.92` → `65.50 56.89 51.66` |
| `go version` | `go version go1.26.8 darwin/arm64` |

## Gaps (관측하지 않은 것)

- **codex 필수 게이트 미실행**: 배차 지시에 따라 `audit_multi`·`codex_audit`를 부르지 않았다. receipt도 없다. 이것이 F1이자 FAIL의 유일한 사유다.
- **Linux 미실행**: macOS에서 심볼릭 링크 없는 TMPDIR로 CI의 경로 표기를 대신했다. ubuntu에서 직접 돌린 결과는 없다.
- **커버리지 미측정**: 지시받은 범위가 `go test`와 `go vet`뿐이었다. 테스트 파일만 바뀌었으므로 제품 코드 커버리지가 줄 수는 없지만, 수치를 관측하지는 않았다.
- **golangci-lint 미실행**: 같은 이유다. progress.md가 적은 기존 진단(193행 `rangeint`)은 이번 diff 밖에 있다.
- **메모 원인을 직접 계측하지 않음**: `resetLocationMemo`는 패키지 밖에서 부를 수 없고, 파일 수정이 금지되어 메모를 끈 대조 실행은 하지 못했다. (a)~(c)의 패턴과 코드 대조로 추론했다. 다른 캐시가 원인일 여지는 (b)·(c)가 경로 표기에만 의존한다는 점에서 매우 낮다.
- **도구 거부 2건**: 워크트리 격리 가드가 셸 변수(`$S`)가 들어간 `git show ... > $S/...` 복합 명령과 `go test` 명령을 거부했다. 같은 명령을 변수 없이 절대 경로로 풀어서 다시 실행했으므로 측정 내용에는 차이가 없다.
- **도구 출처(§2.2)**: `moai verify audit-plan`은 설치된 `/Users/goos/go/bin/moai`(v3.2.0-rc.41)로 돌렸다. 빌드 커밋은 `moai version` 출력에서 확인하지 못했다. 이 명령은 설정을 읽기만 하고, 이번 diff는 그 코드를 바꾸지 않는다.
- 4번 질문의 grep은 `internal/statusline`과 `internal/homestate`만 대상으로 했다.

## Residual-risk (관측했어도 남는 위험)

- CI의 ubuntu가 macOS 대체 실행과 다른 이유(예: 다른 캐시, git 버전 차이)로 다른 횟수를 낼 가능성은 남는다. 다만 수리 후 각 하위 테스트가 새 디렉터리를 쓰므로 메모 적중 경로 자체가 없어진다.
- `TestResolveBacklogCounts_LatencyBudget`은 시간 기반 시험이라 높은 부하에서 간헐적으로 실패할 수 있다(progress.md 기록 2/4). 이번 변경과 무관하지만 CI 재실행에서 다시 보일 수 있다.
- 스크래치 로그(`/private/tmp/claude-501/.../scratchpad/*.log`)는 이 기기에만 남는다. 판정을 가른 출력 줄은 위에 옮겨 적었다.

## 반복 이력

- 1회차(이번): 내용 PASS 0.94, 최종 FAIL(F1 게이트). 이전 감사 없음.

## Operational Notes (unverified)

- `inferred`: F1을 해소한 뒤 다시 감사할 때는 범위를 F1 게이트 하나로 좁혀도 된다. 코드 diff에는 blocking 결함이 없기 때문이다.
- `measured`: 이 워크트리 세션에서 셸 변수를 넣은 `go`/`git` 명령은 격리 가드에 거부된다. 절대 경로를 풀어 쓰면 통과한다(위 Gaps의 거부 2건과 재실행 결과).
