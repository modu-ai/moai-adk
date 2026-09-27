# t1267 sync-audit — signtest 스냅숏 경합 (Class B, Tier S, SPEC 없음)

- 감사 대상: `git diff adf909ca9 601e1beed` (브랜치 `WT-signtest-snapshot-race`, 워크트리 `.claude/worktrees/t1267`)
- 변경 파일 2개: `internal/contract/sign/signtest/signtest.go` (+3), `internal/contract/sign/signtest/signtest_maintenance_test.go` (신규 40행)
- 감사자: sync-auditor (독립 감사, 소스 읽기 전용)
- 평가 프로필: SPEC 없음 → 내장 기본 프로필 (Functionality 40 / Security 25 / Craft 20 / Consistency 15, must-pass = Functionality + Security)

## 판정

**PASS** — 가중 점수 94.5 / 조화평균 93.6. 차단(blocking) 결함은 없고, 선택(optional) 항목 3건과 정보 항목 1건이 있다.

| 차원 | 점수 | 판정 | 근거 |
|---|---|---|---|
| Functionality (40%) | 95/100 | PASS | 수리 전 소스에 회귀 테스트를 걸면 RED, 수리 후 GREEN. 대상 패키지 race 테스트 전부 통과 (아래 E1–E4) |
| Security (25%) | 100/100 | PASS | 테스트 지원 코드에 로컬 git 설정 한 줄을 추가했을 뿐, 입력·비밀·외부 호출 표면에 변화 없음 (E8) |
| Craft (20%) | 90/100 | PASS | `go vet`·`gofmt`·`golangci-lint` 모두 깨끗함. 회귀 테스트는 타이밍과 무관한 결정적 검사이고 양성 대조를 갖춤 (E5, E6) |
| Consistency (15%) | 90/100 | PASS | 저장소에 이미 쓰이는 `maintenance.auto false` 관례를 따름. 다만 형제 픽스처가 함께 거는 `gc.auto 0` 은 생략함 (F1) |

## 질문별 판단

### (1) 원인은 확립됐는가, 그럴듯할 뿐인가

인과 사슬을 고리마다 따로 측정했다. 직접 관측하지 못한 고리는 CI에서 일어난 실제 타이밍 교차 하나뿐이다.

1. CI 오류가 가리키는 파일은 `.git/objects/maintenance.lock` 이다 (E7, 원문 인용).
2. 이 잠금 파일은 `git maintenance run` 이 소유한다. 프로브로 `objects/maintenance.lock` 을 미리 만들어 둔 뒤 `git maintenance run --no-detach --no-quiet` 를 실행하자 `warning: lock file '.git/objects/maintenance' exists, skipping maintenance` 가 출력됐다. 같은 조건에서 `git gc --auto` 는 아무것도 출력하지 않았다 (gc 는 `gc.pid` 를 쓰고 이 잠금은 보지 않는다) (E6).
3. 픽스처의 `git commit` 은 매번 `git maintenance run --auto --quiet --detach` 를 분리 실행한다. GIT_TRACE 로 확인했다 (E5, E6).
4. `maintenance.auto=false` 로 두면 이 분리 실행이 사라진다 (E5 GREEN, E6 세 번째 시나리오).

**`gc.auto=0` 을 추가해야 하는가**: 필요 없다. 프로브에서 `gc.auto=0` 만 걸어도 `git maintenance run --auto --quiet --detach` 는 그대로 실행됐다 (E6). 분리 실행 여부는 git 이 `maintenance.auto` 하나로 결정하고, 작업이 있는지는 분리된 프로세스가 잠금을 잡은 **뒤에** 판단한다. 그래서 오브젝트가 몇 개 없는 저장소에서도 잠금 파일은 생겼다가 사라진다. 픽스처가 실행하는 git 명령은 `init`/`config`/`add`/`commit`/`rev-parse` 이고, 서명 코드(`internal/contract/sign/defaults.go` `runGit`)는 `config`·`rev-parse HEAD` 만 부른다. 이 가운데 자동 maintenance 를 거는 명령은 `commit` 하나이고, 이 경로는 수리로 막혔다.

**git 버전 범위**: `maintenance.auto` 설정은 git 2.29 에서 도입됐다. 분리 실행(`--detach`)은 그보다 최근 버전의 동작이며, 경합이 가능한 것도 이 때문이다. 실측한 버전은 로컬 `git version 2.54.0 (Apple Git-157)` 과 CI Race Test 러너(`Image: ubuntu-24.04`) `git version 2.55.0` 이다 (E7). 둘 다 이 설정을 따른다. macOS·Windows 러너의 git 버전은 재지 않았다 (Gaps 참고).

### (2) Snapshot/AssertUnchanged 설계에 남은 경합

- 픽스처의 git 호출(`Project.Git`)과 서명 경로의 `runGit` 은 모두 동기 `cmd.Run()` 이다. 수리 뒤에는 명령이 반환된 다음 `.git` 에 쓰는 자식 프로세스가 없다.
- 서명 경로의 `runGit` 은 전역/시스템 git 설정을 끄지 않는다. 그래도 실행하는 명령이 `config`·`rev-parse` 뿐이어서, 전역 설정에 `core.fsmonitor` 가 켜져 있어도 인덱스를 갱신하지 않으므로 fsmonitor 데몬을 띄우지 않는다. 로컬 설정이 전역보다 우선하므로 개발자의 전역 `maintenance.auto` 가 픽스처 설정을 뒤집을 수도 없다.
- `internal/cli/contract.go` 에는 `exec.Command` 가 없다 (grep 0행).
- 구조적 잔여 위험: `Snapshot` 은 `.git` 까지 통째로 걷기 때문에 앞으로 백그라운드 writer 가 새로 생기면 같은 부류의 실패가 다시 난다 (F3, optional).

### (3) `t.Setenv(GIT_TRACE)` 안전성과 비공허성

- `internal/contract/sign/signtest/` 에는 `t.Parallel` 이 없다 (grep 0행). 게다가 Go 는 병렬 테스트에서 `t.Setenv` 를 부르면 panic 을 내므로 안전성을 기계적으로 보장받는다. 영향 범위는 같은 패키지 안의 이 테스트 하나다.
- `ScrubbedEnv()` 가 `os.Environ()` 에서 GIT_TRACE 를 넘기기 때문에 `New()` 안의 커밋까지 추적된다.
- 양성 대조 `built-in: git commit` 은 추적이 실제로 켜졌는지를 확인한다. 수리 전 소스(`adf909ca9`)를 `-overlay` 로 끼워 넣고 돌리면 이 테스트가 실패하므로, 테스트가 공허하지 않다는 것도 실측으로 확인했다 (E5).
- 한계: 이 대조는 「추적이 살아 있다」는 것만 보증한다. git 이 앞으로 `run_command` 추적 줄의 형식을 바꾸면 음성 검사가 소리 없이 공허해질 수 있다 (F2, optional).

### (4) signtest.New 소비자

소비자 7개 파일(`internal/cli/contract_ac_test.go`, `internal/cli/contract_helpers_test.go`, `internal/config/autonomy_contract_test.go`, `internal/contract/sign/{ac_contract_016,helpers,sign_edge,sign}_test.go`)을 grep 한 결과, 어디에도 `.Git(` 호출이나 `exec.Command` 가 없다. 저장소를 따로 만들거나 픽스처 설정 밖에서 커밋하는 소비자는 없고, 모두 `New()` 를 거치므로 수리가 그대로 적용된다. 실패 지점 `contract_ac_test.go:318` 은 `newContractProject` (→ `signtest.New` → 커밋) 직후의 `p.Snapshot()` 이며, 진단과 들어맞는다.

## Findings (구조화 결함 목록)

- **F1** [Low, 확신 중간] [optional] `internal/contract/sign/signtest/signtest.go:172` — 형제 픽스처(`internal/graph/t688_push_guard_test.go:191-192`, `internal/kanban/status_read_test.go:41-42`, `internal/hook/pre_tool_test.go:1471-1472`)는 `gc.auto 0` 과 `maintenance.auto false` 를 함께 건다. 이 변경은 뒤쪽만 건다. 프로브에서 `gc.auto=0` 이 분리 실행을 막지 못했고 `maintenance.auto=false` 만으로 충분했으므로 정확성 문제는 아니다. 권장 조치: 관례를 맞추고 싶다면 `p.Git("config", "gc.auto", "0")` 한 줄을 추가한다. 하지 않아도 무방하다.
- **F2** [Low, 확신 중간] [optional] `internal/contract/sign/signtest/signtest_maintenance_test.go:35-38` — 음성 검사가 GIT_TRACE 의 `run_command` 줄 텍스트(`maintenance run`)에 의존한다. 양성 대조는 추적 활성만 보증하므로, 추적 형식이 바뀌면 음성 검사가 공허해질 수 있다. 권장 조치(선택): `trace: run_command:` 접두가 한 줄 이상 존재함을 단언하거나, 수리하지 않은 저장소에서 같은 프로브가 분리 실행을 관측하는지 확인하는 대조군을 추가한다.
- **F3** [Info, 확신 높음] [optional] `internal/contract/sign/signtest/signtest.go` `Snapshot` — `.git` 전체를 걷는 설계는 새로 생기는 백그라운드 writer 에 취약하다. 현재 알려진 writer 는 이번 수리로 모두 없어졌으므로 지금 하드닝할 필요는 없다(도달 불가 상태를 막는 방어 코드가 된다). 재발하면 그때 `.git` 을 제외하는 방안이나 잠금 파일의 ENOENT 를 허용하는 방안을 검토한다.
- **F4** [Info] 증거 공백 — 로컬에서 잠금 경합 자체는 재현되지 않았다(레인 보고와 같음). 원인은 사슬의 고리별 측정으로 확립했고, 실제 CI 타이밍 교차는 관측하지 못했다. 수리 후 develop CI Race Test 가 초록이면 이 공백이 닫힌다.

## Claim / Evidence / Baseline / Gaps / Residual-risk

### Claim

1. 수리는 픽스처 커밋이 띄우던 분리 maintenance 프로세스를 없애고, 회귀 테스트는 수리 전에 RED, 수리 후에 GREEN 이다.
2. `objects/maintenance.lock` 은 `git maintenance run` 이 소유하며, `gc.auto=0` 으로는 분리 실행을 막지 못한다.
3. 대상 패키지(`./internal/contract/...`, `./internal/config/`, `./internal/cli/` 의 TestAC_CONTRACT)가 race 모드에서 통과한다.
4. vet·format·lint 가 깨끗하다.

### Evidence

모든 명령은 `unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED GIT_DIR GIT_WORK_TREE GIT_INDEX_FILE && ` 를 앞에 붙인 한 번의 호출로 실행했다.

**E1** `go test -race -count=1 ./internal/contract/... ./internal/config/` → exit=0

```
ok  	github.com/modu-ai/moai-adk/internal/contract	1.957s
ok  	github.com/modu-ai/moai-adk/internal/contract/sign	13.091s
ok  	github.com/modu-ai/moai-adk/internal/contract/sign/signtest	1.248s
ok  	github.com/modu-ai/moai-adk/internal/config	6.866s
```

**E2** `go test -race -count=2 -run TestAC_CONTRACT ./internal/cli/` → exit=0

```
ok  	github.com/modu-ai/moai-adk/internal/cli	30.106s
```

**E3** `go vet ./internal/contract/...` → `exit=0` (출력 없음). `gofmt -l internal/contract/sign/signtest/` → 출력 없음. `golangci-lint run ./internal/contract/sign/signtest/...` → `0 issues.`

**E4** 커버리지(참고용, Tier S 테스트 지원 패키지): `go test -count=1 -cover -coverpkg=./internal/contract/sign/signtest ./internal/contract/sign/signtest/ ./internal/contract/sign/`

```
ok  	github.com/modu-ai/moai-adk/internal/contract/sign/signtest	0.424s	coverage: 49.2% of statements in ./internal/contract/sign/signtest
ok  	github.com/modu-ai/moai-adk/internal/contract/sign	18.928s	coverage: 88.2% of statements in ./internal/contract/sign/signtest
```

**E5** RED/GREEN. 수리 전 `signtest.go` 는 `git show adf909ca9:...` 로 스크래치에 꺼냈고, 소스는 건드리지 않은 채 `go test -overlay` 로 끼워 넣었다.

RED — `go test -count=1 -overlay <scratch>/overlay.json -run TestFixtureGitStartsNoBackgroundMaintenance -v ./internal/contract/sign/signtest/` → exit=1

```
=== RUN   TestFixtureGitStartsNoBackgroundMaintenance
    signtest_maintenance_test.go:37: a fixture git command started background maintenance:
        18:33:30.379036 run-command.c:672       trace: run_command: git maintenance run --auto --quiet --detach
--- FAIL: TestFixtureGitStartsNoBackgroundMaintenance (0.12s)
FAIL
```

GREEN — 같은 명령을 overlay 없이 실행 → exit=0

```
=== RUN   TestFixtureGitStartsNoBackgroundMaintenance
--- PASS: TestFixtureGitStartsNoBackgroundMaintenance (0.11s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/contract/sign/signtest	0.200s
```

**E6** git 동작 프로브(스크래치 Go 프로그램, 픽스처와 같은 환경: `GIT_CONFIG_GLOBAL=/dev/null`, `GIT_CONFIG_NOSYSTEM=1`)

```
git version 2.54.0 (Apple Git-157)
scenario=default                commit_traced=true background_children=["git maintenance run --auto --quiet --detach"]
scenario=gc.auto=0              commit_traced=true background_children=["git maintenance run --auto --quiet --detach"]
scenario=maintenance.auto=false commit_traced=true background_children=[]
maintenance-run-with-held-lock: err=<nil> out="warning: lock file '.git/objects/maintenance' exists, skipping maintenance"
gc-auto-with-held-maintenance-lock: err=<nil> out=""
```

**E7** CI 실패 원문과 러너 git 버전: `gh run view 36231073168 --job 108374290407 --log` 에서 grep

```
Race Test	Set up job	2026-09-26T08:52:51.0384804Z Image: ubuntu-24.04
Race Test	Checkout code	2026-09-26T08:52:52.2497059Z git version 2.55.0
Race Test	Run race detector across all packages	2026-09-26T09:11:06.8581286Z       contract_ac_test.go:318: signtest: snapshot: open /tmp/TestAC_CONTRACT_014human_path,_stdin_is_an_empty_pipe2339684247/001/.git/objects/maintenance.lock: no such file or directory
```

**E8** 보안 표면: diff 전문(3행 설정 추가 + 40행 테스트)을 직접 읽었다. 외부 입력, 자격증명, 네트워크, 파일 권한과 관련된 변경은 없다.

### Baseline-attribution

- 모든 측정은 이번 실행에서 워크트리 `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t1267`, HEAD `601e1beed`, `git status --short` 비어 있는 상태를 대상으로 했다(감사 전후 두 번 확인).
- RED 기준선은 커밋 `adf909ca9` 의 `signtest.go` 를 overlay 로 대체한 트리이며, 테스트 파일은 `601e1beed` 판이다.
- CI 기준선은 run 36231073168 / job 108374290407 의 로그를 이번 감사에서 직접 받아 grep 한 결과다.

### Gaps

- CI 에서 일어난 실제 잠금 경합의 타이밍 교차는 관측하지 못했다(로컬 재현 불가). 원인은 사슬의 고리별 측정으로 확립했다.
- macOS·Windows CI 러너의 git 버전은 재지 않았다. Race Test 잡(ubuntu-24.04, git 2.55.0)만 측정했다.
- 수리 후 develop CI Race Test 결과는 아직 없다(push 는 리드가 일괄로 한다).
- 교차 모델 감사(`mcp__moai__audit_multi`)는 실행하지 않았다. Tier S·Class B 이고 오케스트레이터가 요청하지 않았다.
- `go test ./...` 전체 스위트는 지시에 따라 실행하지 않았다.

### Residual-risk

- 앞으로 `.git` 에 비동기로 쓰는 다른 writer(새 git 기본 동작, 전역 설정에 의존하는 데몬)가 생기면 `Snapshot` 이 같은 부류의 실패를 다시 낼 수 있다(F3).
- git 추적 형식이 바뀌면 회귀 테스트의 음성 검사가 공허해질 수 있다(F2).
- git 2.29 미만에서는 `maintenance.auto` 가 무시된다. 다만 그런 버전에서는 커밋이 `gc --auto` 를 직접 호출하고, gc 는 `maintenance.lock` 을 쓰지 않으며 작업이 없으면 파일도 만들지 않는다. 현실적인 CI 러너는 이 범위 밖이다.
