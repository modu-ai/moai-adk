auditor-model: claude-opus-5-5

verdict: PASS
audited_sha: edb1f763e21cf0893269ab6667145afc3a6619e0
score: 0.94
blocking_count: 0

# t2062 sync 감사 (iter1) — codemaps 신선도 복구

- 카드: t2062 (`[GRAPH] codemaps described-source-diff >= 40`)
- 트리: `/Users/goos/.moai/worktrees/projects/moai-adk-go/t2062`, 브랜치 `WT-codemaps-described-sourc-2`
- 감사 대상 HEAD: `edb1f763e21cf0893269ab6667145afc3a6619e0`, 기준 main `cf4ad8510d8ba9aabe6a804cadf043611390b31b`
- 감사 방식: Claude 단독(리더 지시). codex·GLM·audit_multi는 호출하지 않았다(Gaps 참고).
- 판정 빌드: 이 트리에서 `go build -ldflags "-X github.com/modu-ai/moai-adk/pkg/version.Commit=edb1f763e2" -o <scratch>/t2062-audit/moai ./cmd/moai`로 만든 바이너리를 경로로 호출. `moai version` 출력 `v3.2.0   edb1f763e   built unknown`. 판정 빌드 커밋 = 트리 HEAD이므로 HEAD의 엄격한 조상이 아니다. 설치본은 쓰지 않았다.

## 결론

블로킹 결함 없이 PASS. 변경 범위는 codemaps 문서 세 개·provenance.json·카드 보고서뿐이고, `described-source-diff`는 이 트리 바이너리로 0을 재현했으며 독립 재계산도 0이다. 폴드 가드 테스트 셋은 손대지 않은 채 통과하고, Phase 4 awk 점검은 COLLECTED 줄만 낸다. 새로 쓴 사실 30건을 소스와 대조해 30건 모두 맞았다. 스탬프 커밋은 main의 조상이다.

## 차원별 점수

| 차원 | 점수 | 판정 | 근거 |
|---|---|---|---|
| Functionality (40%) | 95/100 | PASS | 범위·신선도·폴드 가드·스탬프 조건 6개 모두 충족(아래 1~3, 5, 6). 감점은 창 안의 일부 변경 패키지에 새 서술이 없다는 점(O2) |
| Security (25%) | 100/100 | PASS | 문서·JSON 전용 변경. Go 소스·설정·의존성 변경 0건(아래 1). 추가 줄에 비밀값 없음(diff 전수 열람) |
| Craft (20%) | 90/100 | PASS | 가드 테스트 3개 통과, awk 점검 무결, 사실 정확도 30/30. 감점은 증분 갱신 결정이 레인 자체 판단이라는 점과 범위 밖 서술 미검증(O2) |
| Consistency (15%) | 92/100 | PASS | 기존 판 형식(「현재 변경 경계」 절, 「card … 이 더한 것」 표, 직전 판 강등) 그대로. 새 import 간선 서술이 비테스트 간선 관례를 따름(O3) |

조화평균: 4/(1/0.95+1/1.00+1/0.90+1/0.92) ≈ 0.94. 필수 통과 차원(Functionality·Security) 모두 임계 이상.

## 검증 명령과 관측 출력

### 1. 범위

명령: `git -C <tree> diff --name-only main..HEAD`
```
.moai/project/codemaps/data-flow.md
.moai/project/codemaps/entry-points.md
.moai/project/codemaps/modules.md
.moai/project/codemaps/provenance.json
.moai/reports/t2062/card-review.md
.moai/reports/t2062/graph-check-after.txt
.moai/reports/t2062/graph-check-before.txt
.moai/reports/t2062/progress.md
```
명령: `git -C <tree> diff --name-only main..HEAD -- internal cmd pkg go.mod go.sum .moai/config .claude Makefile` → 출력 없음.
`git rev-parse main` → `cf4ad8510d8ba9aabe6a804cadf043611390b31b`. `git status --short` → 출력 없음(깨끗함).

### 2. 신선도 재현

명령: `<scratch>/t2062-audit/moai graph check` (작업 디렉터리 = 트리, exit 1)
```
codemaps  metric=described-source-diff value=0 threshold=40 verdict=fresh
mx-index  metric=inventory-content-diff value=0 threshold=1 verdict=absent  (mx-index absent (untracked runtime artifact — fresh worktree state))
edges     metric=source-fingerprint-mismatch value=0 threshold=0 verdict=absent  (edges.jsonl absent (untracked derived artifact — fresh worktree state))
citations metric=positive-cited-path-absence value=0 threshold=0 verdict=fresh
```
exit 1은 mx-index·edges가 absent여서이며 codemaps·citations 행과 무관하다.

독립 재계산(스탬프 → HEAD):
`git diff --name-only cf4ad8510d… HEAD -- 'internal/**.go' 'cmd/**.go' 'pkg/**.go' ':(exclude,glob)**/*_test.go' ':(exclude,glob)**/testdata/**' | wc -l` → `0`
필터 없는 `git diff --name-only cf4ad8510d… HEAD -- internal cmd pkg` → 출력 없음.

직전 창 확인(67b4734d6b → cf4ad8510d, 같은 필터): 전체 `115`, `--diff-filter=A` `18`. 카드의 "115 = 추가 18·수정 97" 주장과 일치.

### 3. 폴드 가드

명령: `go test -count=1 -v -run TestCodemapsFoldPreservationGuard ./internal/graph/`
```
=== RUN   TestCodemapsFoldPreservationGuard
--- PASS: TestCodemapsFoldPreservationGuard (0.01s)
ok  	github.com/modu-ai/moai-adk/internal/graph	1.284s
```
명령: `go test -count=1 -v -run TestCodemapsFoldGuardFixtures ./internal/graph/` (실행 테스트 1개, 하위 6개)
```
--- PASS: TestCodemapsFoldGuardFixtures (0.55s)
    --- PASS: TestCodemapsFoldGuardFixtures/intact_copy_passes (0.02s)
    --- PASS: TestCodemapsFoldGuardFixtures/short_form_core/git_is_not_a_violation (0.14s)
    --- PASS: TestCodemapsFoldGuardFixtures/injected_fold_token_fails_naming_unit_and_document (0.32s)
    --- PASS: TestCodemapsFoldGuardFixtures/lost_floor_line_fails_naming_the_unit (0.00s)
    --- PASS: TestCodemapsFoldGuardFixtures/missing_generator_document_fails_naming_the_document (0.04s)
    --- PASS: TestCodemapsFoldGuardFixtures/missing_record_fails (0.01s)
ok  	github.com/modu-ai/moai-adk/internal/graph	1.321s
```
추가: `go test -count=1 -v -run TestCodemapsFoldGuardFloorCoverage ./internal/graph/` → `--- PASS: TestCodemapsFoldGuardFloorCoverage (0.00s)`.

테스트 파일 불변: `git diff --stat main..HEAD -- internal/graph` → 출력 없음(가드 테스트는 `internal/graph/codemaps_fold_guard_test.go`).

변이 탐침: (a) 가드 자체의 `injected_fold_token_fails…` 하위 테스트가 현재 트리의 codemaps 문서를 /tmp 사본으로 복사해 폴드 토큰을 주입하고 가드가 실패함을 확인한다 — 위에서 PASS. (b) Phase 4 awk에 스크래치 문서 하나(`| \`internal/atomicfile/owner.go\` | mutation probe … |`)를 더해 돌린 결과:
```
COLLECTED: fold=53 omission=29
FOLD-PROSE: internal/atomicfile/owner.go 1
```
트리 안의 파일은 건드리지 않았다.

### 4. 새 서술의 정확도 — 30건 대조, 적중 30 / 실패 0

| # | 주장 | 근거(명령 → 관측) | 결과 |
|---|---|---|---|
| 1 | `internal/sqliteuri`의 `File(path, query)` | grep → `sqliteuri.go:23:func File(path string, query url.Values) string {` | 적중 |
| 2 | Windows 드라이브 경로를 `file:///C:/…`로, 공백·`#`·`%`·`?` 인코딩 | `sqliteuri.go:5-6` 주석 "emits the root-absolute form file:///C:/… … percent-encodes spaces, '#', '%' and '?'" | 적중 |
| 3 | sqliteuri 소비자 = ledger·discovery·factorymsg·homestate·cli | `go list -f '{{.ImportPath}} {{.Imports}}'`: ledger·discovery·factorymsg·homestate 모두 `internal/sqliteuri` 포함; `git grep -l` HEAD: `internal/cli/factory_card.go`, `migrate_home_state.go` | 적중 |
| 4 | `moai jev cost [--json]` | `jev_cost.go:58 Use: "cost"`, `:71 Flags().Bool("json"…)`, `:72 jevCmd.AddCommand(costCmd)` | 적중 |
| 5 | jev cost가 호출 수·추정 입력 토큰·정답 결정당 토큰을 집계 | `jev_cost.go:78-81` 필드 `calls`·`estimated_input_tokens`·`correct`·`estimated_input_tokens_per_correct_decision` | 적중 |
| 6 | `moai verify clear-rejection <agent-type> <spec-id> --evidence <path>`, 종료 0/1/2 | `verify_clear_rejection.go:67 Use: "clear-rejection <agent-type> <spec-id>"`, `:68 Short: "… (human path only; exit 0 cleared, 1 refused, 2 usage)"`, `:84 "evidence"` | 적중 |
| 7 | `moai todo add --base-sha <40hex>` | `todo.go:961` 플래그, `:1089 "todo add: --base-sha must be a full 40-hex commit SHA"` | 적중 |
| 8 | `moai factory wait --run <id>` | `factory_wait.go:173 cmd.Flags().StringVar(&run, "run", …)` | 적중 |
| 9 | `ClearRejectionOnApproval`이 `approvals.jsonl`에 승인 기록, 없으면 `ErrMarkNotFound` | `auditreceipt/store.go:1493 approvalsFile = "approvals.jsonl"`, `:1498 ErrMarkNotFound`, `:1529 func ClearRejectionOnApproval` | 적중 |
| 10 | 판정 라벨 `INCONCLUSIVE`(`VerdictInconclusive`) | `auditreceipt/store.go:1062 const VerdictInconclusive = "INCONCLUSIVE"` | 적중 |
| 11 | 닫힌 오류 분류 quota·timeout·auth·unavailable·other | `auditverdict/verdict.go:96 ErrorClasses = []string{"quota", "timeout", "auth", "unavailable", "other"}`; `store.go:196 ErrorClass string` | 적중 |
| 12 | `required_backend` 문법과 `Fields.BlockingDefects` | `verdict.go:77 BlockingDefects []string`, `:193 case "required_backend":` | 적중 |
| 13 | runtime의 `IsInconclusiveReport` | `runtime/audit_counter.go:85` | 적중 |
| 14 | `SecondReviewRecord` 델타 필드 5개 | `closure/records.go:67-71` Delta·DeltaBase·ExtendsReceipt·ConfirmedIDs·VerdictDigests | 적중 |
| 15 | factorylane `absorbOnlyChange` | `factorylane/merge_specless.go:219` | 적중 |
| 16 | `mergeq.Land`가 낡은 보유자 쓰기를 `ErrStaleRow`로 거절 | `mergeq.go:56 var ErrStaleRow`, `:222 func (m *Merger) Land`, `:277` 주석 "… gets ErrStaleRow instead" | 적중 |
| 17 | config `effectiveWorkflow` — 미설정 workflow의 -f/-l 세션은 af-flow | `loader_workflow_disposition.go:50-55` `if factorySession && strings.TrimSpace(profile.Workflow) == "" { return WorkflowAFFlow }` | 적중 |
| 18 | `ModeProfile.ReleaseRoute`(기본 false)·`ValidCodexAuditEfforts`·`ParseInstructionImports` | `types.go:146 ReleaseRoute bool … // af-flow only; default false`, `closed_sets.go:121`, `token_budget_guard.go:315` | 적중 |
| 19 | `worktreeroot.CanonicalSpelling`을 훅 레지스트리 재키잉이 사용 | `worktreeroot.go:552`; `hook/worktree_registry.go:188,190` | 적중 |
| 20 | `claim.OwnerState.Refusal()` 공개 메서드 | `claim/presence.go:52 func (o OwnerState) Refusal() error` | 적중 |
| 21 | 사라진 간선 runner → homestate | `go list` runner Imports에 homestate 없음; 옛 스탬프 `67b4734d6b:internal/agenticfactory/runner/runner.go:24` 에 import 존재 | 적중(비테스트 기준, O3) |
| 22 | 새 간선 checkpoint·ctxmodel → gitenv | `go list` 두 패키지 모두 `internal/gitenv`; 옛 스탬프에서는 `main_test.go`만 import | 적중 |
| 23 | 새 간선 cli/worktree → atomicfile, hook → worktreeroot, cli → jevledger | `go list`·HEAD `git grep` 존재; 옛 스탬프 `git grep -l` 세 건 모두 출력 없음 | 적중 |
| 24 | ledger 의존 = core와 sqliteuri | `go list` ledger Imports: `agenticfactory/core`, `internal/sqliteuri` (그 밖은 표준·modernc) | 적중 |
| 25 | `PushGateTarget`·`NewestVerdictFile`·`CardWorktreeBindings` 공개 | `card_evidence.go:71`, `card_evidence_readers.go:249`, `card_worktree.go:211` | 적중 |
| 26 | factory 백로그 기록의 `BaseSHA` | `factory/backlog_store.go:153 BaseSHA string \`json:"base_sha,omitempty"\`` | 적중 |
| 27 | 창 115개 = 추가 18 | 위 2절 재계산 `115`, `18` | 적중 |
| 28 | 패키지 199개 | `go list ./internal/... ./cmd/... ./pkg/... | wc -l` → `199` | 적중 |
| 29 | cli 새 파일(audit_delta_*·audit_error_class·audit_rerun_legs·jev_cost+플랫폼 쌍·verify_clear_rejection·update_progress·filemode_portable) | 창 `--name-status`에 `A` 로 모두 존재, 플랫폼 쌍은 `jev_cost_open_unix.go`·`jev_cost_open_windows.go` | 적중 |
| 30 | `spec/lint_ownership_batch.go`, `mission/filemode.go` 신규 | 창 `--name-status`: `A internal/spec/lint_ownership_batch.go`, `A internal/mission/filemode.go` | 적중 |

### 5. 폴드 위생 (Phase 4 awk)

skill 본문의 awk 프로그램을 인라인으로, 상위 문서 6개(data-flow·dependencies·docs-truth·entry-points·modules·overview)에 대해 실행:
```
COLLECTED: fold=53 omission=29
```
UNCOVERED·FOLD-PROSE·GAP 줄 없음.

### 6. 스탬프 생존성

- `provenance.json` `commit_sha`: `cf4ad8510d8ba9aabe6a804cadf043611390b31b` (= `git rev-parse main`, 브랜치 지역 커밋 아님)
- `git merge-base --is-ancestor cf4ad8510d… main; echo "exit=$?"` → `exit=0`
- `git merge-base --is-ancestor cf4ad8510d… origin/main; echo "exit=$?"` → `exit=1`
- `git branch -r --contains cf4ad8510d…` → `origin/release/main-batch-20261106`
- `git rev-list --count --left-right origin/main...main` → `0	1621`

리더 조건(main의 조상, 지역 커밋 아님)은 충족. origin/main에는 아직 없다(O1).

## 블로킹 결함

없음.

## 선택 사항(블로킹 아님)

- O1 [low] [optional] [outside] 스탬프 `cf4ad8510d`는 로컬 main과 `origin/release/main-batch-20261106`에만 있고 origin/main에는 없다. 배치 PR이 squash로 병합되면 스탬프가 원격 main에서 도달 불가가 된다(SPEC-STAMP-REACHABILITY-001 우려). 머지 커밋 방식이면 문제없다. 필요한 조치: 배치 PR 병합 방식을 확인하고, squash라면 병합 뒤 스탬프를 다시 찍는다.
- O2 [low] [optional] [touched-file] 창의 변경 패키지 중 서술이 새로 붙지 않은 곳이 있다(`internal/factorymsg`의 run_state·store, `internal/discovery`, `internal/testguard`, `internal/web`, `internal/mission` 영수증 계열, `internal/hook`의 branch_guard·commit_identity_guard 등). 지표는 신선도만 보므로 해당 패키지의 기존 서술이 이 창에서 틀어졌는지는 검증하지 않았다. 필요한 조치: 없음(후속 갱신에서 확인 권장).
- O3 [info] [optional] [in-diff] "사라진 간선 runner → homestate"는 비테스트 간선 기준이다. `runner/takeover_presence_test.go`는 여전히 homestate를 import 한다. 문서 전반의 비테스트 간선 관례와 일치하므로 수정 불요.
- O4 [info] [optional] [in-diff] 레인은 `--force` 전체 재생성 대신 증분 갱신을 택했다(진행 기록 결정 1, 스스로 공개). 리더 지시 "도구로 재생성, 손으로 땜질 금지"와의 관계: 이 저장소의 codemaps 문서 생성기는 codemaps skill 자체이며 skill이 증분 갱신을 지원하고, 스탬프는 `generated_by: codemaps-gen` 형식, Phase 4 점검·폴드 가드가 모두 통과했다. 리더가 정책상 전체 재생성만 인정한다면 이 항목은 리더 판단 사안이다.
- O5 [info] [optional] [in-diff] `provenance.json`의 `tree_root`가 기계 로컬 절대경로를 싣는다. 직전 스탬프와 같은 형식이라 신규 문제는 아니다.

## Claim / Evidence / Baseline-attribution / Gaps / Residual-risk

- Claim: t2062 변경은 문서 전용이고, codemaps 신선도(값 0)·인용 신선도(값 0)를 복구하며, 폴드 가드와 Phase 4 점검을 통과하고, 표본 사실 30건이 소스와 일치하며, 스탬프는 main의 조상이다.
- Evidence: 위 1~6절의 명령과 원문 출력.
- Baseline-attribution: 트리 `edb1f763e2`(이 감사 시작 시 HEAD, 감사 중 이동 없음), 기준 main `cf4ad8510d`, 직전 스탬프 `67b4734d6b`. 도구 측정은 이 트리에서 빌드한 바이너리(커밋 `edb1f763e`)와 같은 트리에서 실행한 `go test`·`go list`가 냈다.
- Gaps:
  - 디스크 비상(리더 통지, 여유 공간 3 GiB 이하)은 이 감사의 모든 빌드·테스트·`go list` 실행이 끝나고 판정 파일을 쓴 뒤에 도착했다. 위에 인용한 출력은 모두 끝난 실행의 것이며, 통지 이후 새 빌드·테스트는 시작하지 않았다. 감사용 스크래치 캐시(`gocache-t2062-audit` 489M, `t2062-audit` 94M)는 지우지 않고 남겨 두었다.
  - 감사 계획(`moai verify audit-plan`, 트리 빌드)은 `model: codex`, `cross_model_active: true`, `cross_model_required: false`, codex gate `advisory`, claude·glm `off`를 보고했다. 리더 지시에 따라 codex·GLM·audit_multi를 호출하지 않았다. required 게이트가 없어 PASS를 막지 않지만, advisory codex 다리는 실행되지 않았다. 영수증은 발급되지 않았다.
  - 워크트리 가드 거절 2건: (a) `awk -f <scratch 파일>` — "what it runs cannot be shown not to be git" 사유로 거절, 같은 프로그램을 인라인 인자로 다시 실행해 결과를 얻었다. (b) 글롭이 든 `git ls-files internal/cli/jev_cost*` — 거절, 대신 창 `--name-status` 목록으로 파일 존재를 확인했다. 또 zsh가 따옴표 없는 `--include=*.go`를 글롭으로 풀어 grep 4건이 실행되지 않았고 따옴표를 붙여 재실행했다.
  - 폴드 가드의 변이 탐침은 가드 자체 픽스처 하위 테스트(현재 문서 사본 + 주입)와 awk 스크래치 탐침으로 갈음했다. modules.md 전체 사본을 따로 변형해 Go 가드를 돌리지는 않았다.
  - `go test ./internal/graph/` 패키지 전체·저장소 전체 테스트는 실행하지 않았다(문서 전용 변경, 지정 가드 3개만).
  - 새 서술 외의 기존 서술(O2)의 정확도는 검증하지 않았다.
  - mx-index·edges 층은 이 트리에 없어(absent) 측정하지 않았다. Windows 미실행.
- Residual-risk:
  - O1의 squash 병합 시 스탬프 고아화.
  - main이 배치 PR 착지 전에 더 움직이면 `described-source-diff`가 다시 오른다.
  - 원시 `go test`·`go list` 전체 출력은 스크래치에만 있고, 판정을 결정한 줄만 이 파일에 옮겼다. 스크래치 내용은 인용 근거가 아니다.
