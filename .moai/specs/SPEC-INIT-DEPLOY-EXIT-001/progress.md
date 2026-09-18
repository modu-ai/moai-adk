# SPEC-INIT-DEPLOY-EXIT-001 — 진행 기록

카드: **t931**

## §E.1 Plan-phase Audit-Ready Signal

- plan-phase 산출물 작성 완료: `spec.md`, `plan.md`, `acceptance.md`, `progress.md`
- status: `draft`
- Tier M 산출물 집합 충족 (design.md / research.md 는 Tier L 전용이므로 미생성)

## §E.2 Run-phase Evidence

측정 기준: 워크트리 `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t931`, 브랜치 `WT-init-render-exit`,
base HEAD `3ff2e5dda`, 미커밋 작업 트리. 아래 출력은 모두 이번 실행·이 트리에서 관측한 것이다.

### E.2.1 RED — 계약 반전 3건이 구현 전에 실패함 (initializer 층)

```
$ go test ./internal/core/project/ -count=1 -run 'TestInit_WithDeployerError|TestInit_WithDeployerNoManifest|TestInit_DeployFailureKeepsResultWarnings|TestInit_NonDeployFailureStaysWarning|TestInit_NilDeployerFallbackStaysNonFatal' -v
=== RUN   TestInit_DeployFailureKeepsResultWarnings
    initializer_deploy_fatal_test.go:61: Init() returned nil on a deploy failure; deployment failure must be fatal
--- FAIL: TestInit_DeployFailureKeepsResultWarnings (0.02s)
=== RUN   TestInit_NonDeployFailureStaysWarning
--- PASS: TestInit_NonDeployFailureStaysWarning (0.02s)
=== RUN   TestInit_NilDeployerFallbackStaysNonFatal
--- PASS: TestInit_NilDeployerFallbackStaysNonFatal (0.00s)
=== RUN   TestInit_WithDeployerError
    initializer_test.go:619: Init() returned nil on a template deployment failure; deployment failure must be fatal
--- FAIL: TestInit_WithDeployerError (0.02s)
=== RUN   TestInit_WithDeployerNoManifest
    initializer_test.go:651: Init() returned nil when the deployer had no manifest manager; deployment failure must be fatal
--- FAIL: TestInit_WithDeployerNoManifest (0.00s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/core/project	0.450s
FAIL
```

**감추지 않고 기록한다**: 같은 회차에서 `TestInit_NonDeployFailureStaysWarning`(AC-IDE-004)과
`TestInit_NilDeployerFallbackStaysNonFatal`(acceptance §D.1)은 **첫 실행부터 통과**했다. 이 둘은 새
계약이 아니라 보존 가드이므로 초록이 정상이며, 이 SPEC 의 계약 반전을 관측한 것은 나머지 3건뿐이다.

### E.2.2 RED — CLI 층 (구현 전)

```
$ go test ./internal/cli/ -count=1 -run 'TestInit_DeployFailure' -v
init_deploy_exit_test.go:104: error does not state that the project tree is incomplete: initialization failed: initialization: template deployment: deploy templates: template render ".claude/hooks/moai/handle-agent-hook.sh.tmpl": unexpanded dynamic token detected: found "$REPO"
init_deploy_exit_test.go:107: error carries no next-action guidance: initialization failed: initialization: template deployment: ...
--- FAIL: TestInit_DeployFailureSuppressesSuccessCard (0.02s)
init_deploy_exit_test.go:127: warning summary absent from stderr: "· Initializing MoAI project...\n"
init_deploy_exit_test.go:130: executor result warning did not reach the summary: "· Initializing MoAI project...\n"
--- FAIL: TestInit_DeployFailurePreservesWarningSummary (0.02s)
    --- FAIL: TestInit_DeployFailureConsistentAcrossPaths/force (0.02s)
    --- FAIL: TestInit_DeployFailureConsistentAcrossPaths/codex-only (0.02s)
FAIL	github.com/modu-ai/moai-adk/internal/cli	0.860s
```

이 회차에서 성공 카드 부재 단언은 이미 통과했다(오류 반환 분기가 카드 인쇄보다 앞에 있기 때문).
붉었던 것은 오류 문안(불완전 트리 · 다음 행동)과 경고 요약 보존 두 축이다.

### E.2.3 GREEN — 영향 패키지 전량

```
$ go test ./internal/core/project/... ./internal/cli/... -count=1 -timeout 30m
ok  	github.com/modu-ai/moai-adk/internal/core/project	2.409s
ok  	github.com/modu-ai/moai-adk/internal/cli	1185.847s
ok  	github.com/modu-ai/moai-adk/internal/cli/agentlint	1.101s
ok  	github.com/modu-ai/moai-adk/internal/cli/harness	2.601s
ok  	github.com/modu-ai/moai-adk/internal/cli/pr	0.258s
ok  	github.com/modu-ai/moai-adk/internal/cli/preference	1.598s
ok  	github.com/modu-ai/moai-adk/internal/cli/printer	0.901s
ok  	github.com/modu-ai/moai-adk/internal/cli/ptycaptest	4.356s
ok  	github.com/modu-ai/moai-adk/internal/cli/specid	0.236s
ok  	github.com/modu-ai/moai-adk/internal/cli/taskledger	0.628s
ok  	github.com/modu-ai/moai-adk/internal/cli/uikit	0.681s
ok  	github.com/modu-ai/moai-adk/internal/cli/update	0.613s
ok  	github.com/modu-ai/moai-adk/internal/cli/update/backup	1.471s
ok  	github.com/modu-ai/moai-adk/internal/cli/update/deploy	1.516s
ok  	github.com/modu-ai/moai-adk/internal/cli/update/merge	1.549s
ok  	github.com/modu-ai/moai-adk/internal/cli/update/plan	0.241s
ok  	github.com/modu-ai/moai-adk/internal/cli/update/report	0.242s
ok  	github.com/modu-ai/moai-adk/internal/cli/wizard	6.487s
ok  	github.com/modu-ai/moai-adk/internal/cli/worktree	5.496s
[exited with code 0]
```

위 전량 회차는 lint 후속 편집(테스트 헬퍼 반환 순서 정리) **이전** 트리 측정이다. 편집 후 재측정:

```
$ go test ./internal/cli/ -count=1 -run 'TestInit' -timeout 30m
ok  	github.com/modu-ai/moai-adk/internal/cli	21.923s

$ go test ./internal/core/project/ -count=1 -timeout 10m
ok  	github.com/modu-ai/moai-adk/internal/core/project	5.110s
```

### E.2.4 품질 게이트

```
$ go vet ./internal/core/project/... ./internal/cli/...
(출력 없음 — exit 0)

$ golangci-lint run ./internal/core/project/... ./internal/cli/...
0 issues.

$ go build ./...
build exit: 0
```

lint 1차 회차에서 지적 1건(`ST1008: error should be returned as the last argument`,
`internal/cli/init_deploy_exit_test.go:50`)이 나왔고, 테스트 헬퍼 반환 순서를
`(string, string, error)` 로 고쳐 해소했다. 위 `0 issues.` 는 수정 후 재측정값이다.

### E.2.5 변경 파일

```
$ git diff --stat -- internal/
 internal/cli/init.go                              | 22 +++++++++++-
 internal/core/project/initializer.go              | 16 +++++++--
 internal/core/project/initializer_persist_test.go |  6 ++--
 internal/core/project/initializer_test.go         | 42 ++++++++++-------------
 internal/core/project/phase.go                    |  6 +++-
 5 files changed, 61 insertions(+), 31 deletions(-)

$ git status --short
 M internal/cli/init.go
 M internal/core/project/initializer.go
 M internal/core/project/initializer_persist_test.go
 M internal/core/project/initializer_test.go
 M internal/core/project/phase.go
?? .moai/specs/SPEC-INIT-DEPLOY-EXIT-001/
?? internal/cli/init_deploy_exit_test.go
?? internal/cli/init_executor_seam.go
?? internal/core/project/initializer_deploy_fatal_test.go
```

- `internal/core/project/initializer.go` — 배포 실패를 `return result, fmt.Errorf("template deployment: %w", err)` 로 치명화. result 를 오류와 함께 반환하는 것이 핵심(실패 전에 기록된 skill-mirror 통지 보존).
- `internal/core/project/phase.go` — `Init` 오류 시 result 를 함께 전파.
- `internal/cli/init.go` — 오류 경로에서 `result.Warnings` 수집 + 불완전 트리 문안(파일 수 없음) 부착.
- `internal/cli/init_executor_seam.go` (신규) — `newInitPhaseExecutorFn` 주입 지점. 기본값은 실제 `project.NewPhaseExecutor`.
- `internal/core/project/initializer_test.go` — `TestInit_WithDeployerError` / `TestInit_WithDeployerNoManifest` 계약 반전.
- `internal/core/project/initializer_deploy_fatal_test.go` (신규) — AC-IDE-003/004 + §D.1 경계.
- `internal/cli/init_deploy_exit_test.go` (신규) — AC-IDE-005/006/007.
- `internal/core/project/initializer_persist_test.go` — "배포 실패는 비치명적 경고" 스테일 주석 정정(M5).

범위 규율: `initializer.go` 의 나머지 일곱 저하 지점(줄 216·236·255·272·293·304·396)은 diff 에
전혀 나타나지 않는다. `internal/template/templates/` 변경 0건.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-09-18
run_commit_sha: pending-backfill
run_status: complete
ac_total_count: 8
ac_pass_count: 7          # AC-IDE-001~007
ac_fail_count: 0
ac_gap_count: 1           # AC-IDE-008 — e2e J1/J1b 미실행, 관측 없음
preserve_list_post_run_count: 7   # initializer.go 의 나머지 저하 지점, 전부 미변경
l44_pre_commit_fetch: not-run     # 커밋은 리드 소관
l44_post_push_fetch: not-run
new_warnings_or_lints_introduced: 0
cross_platform_build:
  darwin: pass        # go build ./... exit 0
  windows: not-run    # GOOS=windows 미측정 — CI 몫
total_run_phase_files: 8
m1_to_mN_commit_strategy: 미커밋 — 리드가 통합 창에서 커밋
```

### AC PASS/FAIL 행렬

| AC | 층 | 판정 | 근거(명령 + 관측) | 측정 주체 |
|---|---|---|---|---|
| AC-IDE-001 | 바이너리 프로브 | PASS | `$REPO` 주입 + `make build` + `./bin/moai init <dir> --llm claude --non-interactive` → **exit 1**, 47 파일, `MoAI project initialized` **0회**, stderr 에 실패 템플릿 경로 + `The project tree is INCOMPLETE — ...` | **orchestrator(lane session) 측정** — 이 에이전트가 돌린 것이 아님 |
| AC-IDE-002 | 바이너리 프로브(양성 대조) | PASS | 템플릿 무수정 + 같은 명령 → exit 0, **596 파일**, `MoAI project initialized` 1회, AGENTS.md 존재 | **orchestrator(lane session) 측정** |
| AC-IDE-003 | initializer (Go) | PASS | `TestInit_WithDeployerError`, `TestInit_DeployFailureKeepsResultWarnings` — RED 관측(§E.2.1) 후 GREEN(§E.2.3) | manager-develop(본 에이전트) |
| AC-IDE-004 | initializer (Go) | PASS | `TestInit_NonDeployFailureStaysWarning` — 셸 설정 seam 실패 시 `Init()` nil + `shell configuration` 경고. **첫 실행부터 초록**(보존 가드) | manager-develop |
| AC-IDE-005 | CLI (Go) | PASS | `TestInit_DeployFailureSuppressesSuccessCard` — RED 관측(§E.2.2) 후 GREEN. `RunE` non-nil 오류 + 성공 카드 부재 + 불완전 트리 문안 | manager-develop |
| AC-IDE-006 | CLI (Go) | PASS | `TestInit_DeployFailurePreservesWarningSummary` — stderr 에 `warning(s) during init:` + 통지, stdout 누출 0 | manager-develop |
| AC-IDE-007 | CLI (Go) | PASS | `TestInit_DeployFailureConsistentAcrossPaths/{force,codex-only}` — 두 경로 모두 non-nil 오류 + 카드 부재 | manager-develop |
| AC-IDE-008 | 회귀 | **GAP** | `e2e/cli/tux3_journeys.sh` J1/J1b 미실행 — 이 에이전트도 오케스트레이터도 측정하지 않음 | — |

**귀속 주의**: AC-IDE-001/002 는 본 에이전트가 관측한 값이 아니다. 배차 지시에 따라 M4 바이너리
프로브는 오케스트레이터(lane session)가 수행했고, 위 수치는 그 측정의 **보고값**이다. 측정 조건은
워크트리 t931 · 브랜치 `WT-init-render-exit` · base HEAD `3ff2e5dda` · 미커밋 작업 트리이며,
주입한 템플릿은 원복 완료되어 `git status` 에 `internal/template/templates/` 변경 0건이다.

### 결함 재현(수리 전 기준선) — orchestrator 측정

같은 주입 조건에서 수리 전 바이너리: **exit 0 · 77 파일 · 성공 카드 출력.** 이것이 SPEC §A.1 이
서술한 결함의 재현이며, 수리 후 같은 조건이 exit 1 · 카드 0회로 뒤집힌 것이 이 SPEC 의 효과다.

### Gaps (관측하지 않은 것)

- **AC-IDE-008** — e2e J1/J1b 시나리오 미실행.
- `go test ./...` 전량, `GOOS=windows GOARCH=amd64 go build ./...`, `make build` 는 이 에이전트가
  돌리지 않았다(전 패키지·크로스플랫폼 판정은 CI 몫이라는 배차 지시). `make build` 는 오케스트레이터가
  프로브 과정에서 수행했으나, 그 결과는 프로브 귀속이지 이 에이전트의 측정이 아니다.
- `internal/cli` 1185초 전량 통과는 lint 후속 편집 **이전** 트리 측정이다. 편집 후에는 `-run TestInit`
  범위로만 재측정했고, 그 밖의 `internal/cli` 테스트는 편집 후 트리에서 재측정하지 않았다.
- 실패 유도 시 배포 파일 수가 수리 전 77 → 수리 후 47 로 줄어든 이유는 **확인하지 않았다.** 배포 중단
  시점이 같더라도 후속 초기화 단계가 오류로 조기 종료되기 때문이라는 추정이 있으나, 검증하지 않았으므로
  사실로 취급하지 않는다.

### Residual-risk

- CLI 층 AC 는 주입한 executor 로 검증했으므로, 실제 deployer 가 `template deployment` 부분 문자열을
  포함한 오류를 낸다는 접합면은 initializer 층 테스트와 오케스트레이터 바이너리 프로브가 닫는다.
  이 문자열 매칭이 두 층을 잇는 유일한 접합면이다.
- `phase.go` 가 오류와 함께 result 를 반환하도록 바뀌었다. 현재 프로덕션 소비자는 `init.go` 하나뿐이고
  nil 가드를 넣었으나, 앞으로 "오류면 result 는 nil" 을 가정하는 소비자가 생기면 틀린다.
- 오류 문안의 `INCOMPLETE` 토큰에 테스트 3건이 의존한다(의도된 결합). 문안 변경 시 함께 붉어진다.
- 새로 도입한 `newInitPhaseExecutorFn` seam 은 테스트 주입점이다. 스왑한 테스트가 `t.Parallel` 을
  쓰거나 `t.Cleanup` 복원을 빠뜨리면 같은 패키지의 다른 init 테스트를 오염시킨다.

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
