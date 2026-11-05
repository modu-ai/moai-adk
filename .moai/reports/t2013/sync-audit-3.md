auditor-model: claude-opus-5-5
verdict: FAIL
audited_sha: 6cb0c11f325fdabbb9b80e1bfabf0df92c251ed0
score: 0.76

# Sync 감사 3회차(확인 감사) — 카드 t2013 (SPEC-FACTORY-COMPLETION-RECOVERY-001 개정)

- 트리: `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t2013`, 브랜치 `WT-agent-trunk-factory`
- 감사 대상: `git rev-parse HEAD` → `6cb0c11f325fdabbb9b80e1bfabf0df92c251ed0`. 작업 트리에는 `.moai/reports/t2013/progress.md` 수정 1건만 있음(보고서 파일, Go 코드 아님).
- 1차 범위(델타): `ff2158a312..6cb0c11f32` 6커밋 — cc7c613122(RED)·98586ee059(GREEN) checkpoint, 733d74617c(RED)·f722ee92c5(GREEN) BranchGuard, c4f8c6d99a(progress 기록), 6cb0c11f32(acceptance·D16).
- 2차 범위: 카드 범위 `cb30e05b3e..6cb0c11f32` 중 위 수리로 생긴 회귀만.
- 승인: sync-audit-2 FAIL 뒤의 사전 승인 확인 감사 1회. 이후 감사 없음.
- 감사계획(`moai verify audit-plan --project-root <tree>`): `config_status: ok`, `model: codex`, codex `required`, claude `off`, glm `off`, `cross_model_required: true`, `enforced_required: ["codex"]`.

## 종합 판정

**FAIL (0.76).**

1. **지난 감사의 지적 F1~F4는 모두 해소됐고, 수리가 만든 회귀는 찾지 못했다.** 시험을 직접 돌려 확인했다(hook 218건 PASS·FAIL 0, checkpoint 8건 PASS, 개명된 receipt 시험 PASS).
2. 그런데 **required codex 다리가 다시 `fail`을 냈다**(receipt `rcpt-7ca077fcb8808436b4e45880`). 이번 지적 4건은 모두 수리 커밋이 건드리지 않은 카드 범위의 기존 코드에 대한 것이고, 앞선 두 회차에서 나오지 않았던 새 지적이다. required 다리의 FAIL은 PASS를 막는다.
3. 4건을 소스로 직접 대조했다. 2건(C1: 착지 뒤 단계로 넘어간 선행 카드에 대한 대기, C3: go.mod만 바뀐 경우 lint·test 생략)은 정확성 결함으로 확인되어 blocking으로 분류했다. 나머지 2건(C2 카드 임대 갱신, C4 lint 도구 버전)도 확인됐지만 안전한 거절이거나 영향이 작아 optional로 분류했다.

Must-pass firewall: Functionality가 blocking 2건으로 FAIL이므로 다른 점수와 상관없이 전체 FAIL이다.

## 이전 지적 처리 결과 (F1~F4)

| 지적 | 수리 커밋 | 판정 | 근거 |
|---|---|---|---|
| F1 [High] BranchGuard 명시 대상 채택 회귀 | 733d74617c / f722ee92c5 | **해소** | `explicitGitTarget`(branch_guard.go:472-498): 선행 `cd <abs> &&` 뒤에 `cd`/`pushd`/`popd`/`(`나 `-C`가 있으면 대상으로 채택하지 않음(479-486). 명령 어디든 `GIT_DIR=`/`GIT_WORK_TREE=`/`--git-dir`/`--work-tree`가 있거나 `-C`가 둘 이상이면 `redirected`(477). 이 경우 git 전역 옵션을 지운 명령으로 훅 cwd 판정에 넘겨 동사 매처가 그대로 본다(488). 새 시험 `TestBranchGuardExplicitTargetNotAdoptedWhenLaterSegmentRedirects`가 8가지 형태(두 번째 cd, GIT_DIR, GIT_WORK_TREE, pushd, 서브셸, --git-dir, --work-tree, -C 두 개) 모두 주 체크아웃에서 거절되는지 확인하며 PASS. `cd /wt && git commit`·`git -C /wt commit` 허용(`…AllowsWorktree` PASS), 주 체크아웃 commit/revert/cherry-pick 거절(`…DeniesPrimary`, `TestProtectedCommit*` PASS). |
| F2 [Medium] checkpoint 행 선택 | cc7c613122 / 98586ee059 | **해소** | `checkpointRowRank`/`checkpointRowBetter`(factory_checkpoint.go:394-415): 비종단 행 → 종단(done/failed) 행 → abandoned 행 순서이고, 같은 순위 안에서만 `updated_at`을 비교. `terminalStates = {CardDone, CardFailed, CardAbandoned}`(card_record.go:74). 시험 `TestFactoryCheckpointPrefersLiveRowOverNewerAbandoned`(더 최근 abandoned 행 주입) PASS. |
| F3 [Medium] 작업트리 부재 시 다른 체크아웃의 git 상태 기록 | cc7c613122 / 98586ee059 | **해소** | 작업트리가 없으면 `gitBase = ""`이고, `checkpointGit`은 빈 base에서 `("", false)`를 돌려 git을 실행하지 않음(258-260). Branch/Head는 `unknown`, Dirty/Unpushed는 `-1`, `worktree: absent` 줄이 추가됨. 파일 저장 위치만 root로 남음. 시험 `TestFactoryCheckpointAbsentWorktreeReportsNoForeignGitState`(root를 별도 dirty 저장소로 구성) PASS. |
| F4 / D16 장부 | 6cb0c11f32 | **해소(D16)** | acceptance.md:444가 `TestLeaderReceiptGateIssuerIndependence`를 인용하고, 이 이름의 시험이 실제로 있으며(leader_approval_gate_test.go:963) 이번 실행에서 PASS. progress.md D16 행은 "처분됨"으로 바뀜. |
| F4 / D15·D17·D18·D20 | — | **기록된 이월** | SPEC progress.md:30 "D10 이월 승인 기록"(리더 승인 2026-10-26 15:4x KST, final SPEC close sync로 이월)과 :208-211 표. 지시에 따라 새 결함이 아니라 기록된 이월로 처리했다. 기록과 어긋나는 사실은 찾지 못함. |

## 차원별 점수

| 차원 | 점수 | 판정 | 근거 |
|---|---|---|---|
| Functionality (40%) | 68/100 | FAIL | F1~F3 해소를 시험으로 확인(hook `ok … 30.976s`, PASS 218건·FAIL 0, checkpoint `ok … 14.648s`). 감점: required codex 다리 fail, blocking으로 확인된 C1·C3(카드 범위 기존 코드). |
| Security (25%) | 85/100 | PASS | F1 우회 경로를 닫음(8형태 거절 시험). 새 SQL·문자열 결합 없음(수리 diff는 정규식과 순위 함수뿐). 감점: O1(아래) — `cd /wt && git -C /primary commit`을 worktree cwd에서 실행하면 허용되는데, 기준점(cb30e05b3e)과 같은 동작이라 회귀는 아님. |
| Craft (20%) | 80/100 | PASS | RED→GREEN 커밋 쌍 유지, `gofmt -l` 출력 없음, `go vet ./internal/hook/ ./internal/cli/` 출력 없음. 시험은 지난 감사가 지적한 형태를 하나씩 고정함. 감점: golangci-lint·커버리지 미측정(Gaps), C4(검증 캐시 도구 식별). |
| Consistency (15%) | 85/100 | PASS | 새 정규식·상수·주석이 파일의 기존 관례를 따름(`checkpointUnknown` 옆의 `checkpointWorktreeAbsent`, 기존 `parseCheckpoint`/`render` 키 처리 방식). D16 장부 행과 acceptance 인용이 일치. |

가중 조화평균: 0.40/68 + 0.25/85 + 0.20/80 + 0.15/85 = 0.005882 + 0.002941 + 0.002500 + 0.001765 = 0.013088 → 76.4 → **0.76**.

## Codex 다리 (mcp__moai__codex_audit, 1회)

- 호출: `codex_audit(project_root=/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t2013, target=baseBranch, mode=native)`
- 결과: `verdict: fail`, `audit_receipt: rcpt-7ca077fcb8808436b4e45880`, `build_commit: b1806498fe`, `review_base: main (merge base cb30e05b3e5f64774b8dabd87e80c24332f4551d)`. 결과에 `gate_unmet` 필드는 없음(verdict 자체가 `fail`).
- 지난 회차의 P1(F1)·P2 두 건(F2·F3)은 이번 결과에 다시 나오지 않았다.
- `audit_multi`는 호출하지 않았다(지시가 codex_audit 단일 호출을 명시함). 오케스트레이터가 `moai verify audit-plan --result-file`로 확인할 때 넣을 값은 per-backend `{backend: codex, gate: required, verdict: fail}`이다. 그 확인이 끝나기 전까지 이 판정은 최종이 아니다.

| codex | 내용 | 이 감사의 소스 대조 | 분류 |
|---|---|---|---|
| C1 P2 `internal/cli/factory_wait.go:256` | `factoryLandedPredecessor`는 `factoryLandedState`(factory_wait.go:35-37: `merged-local` 또는 `merged-pr`만 참)로 판정한다. 선행 카드가 `pushed`/`ci-green`/`done`으로 넘어간 뒤에 기록한 대기는 깨어나지 않고, `--next`가 있으면 stop-goal 차단도 계속 꺼진다. | **확인.** 1회차 F3("이미 착지한 선행 카드") 수리가 착지 직후 상태만 다룬다. | blocking (Medium, 정확성) |
| C2 P2 `internal/cli/factory_card.go:3363` | 빠른 필수 검사가 끝난 뒤 통합 window 임대만 갱신한다(3363-3368). merge 단계의 카드 읽기(`integrationReadMergeCardForRun`, integration_merge.go:189-193)는 카드 임대가 만료되지 않았는지(`LeaseUnexpired`) 확인한다. | **확인.** 검사가 카드 임대 잔여 시간보다 길면 merge가 거절된다. 다만 거절은 카드 상태를 바꾸지 않는 안전한 실패이고, `verify run` 재사용과 owner 재임대 경로(57d469a679)로 다시 시도할 수 있다. | optional (Medium, 가용성) |
| C3 P2 `internal/cli/factory_release_route.go:252` | 바뀐 경로를 `.go` 파일이 있는 가장 가까운 상위 디렉터리로 매핑한다(294-322). 저장소 루트에는 `.go` 파일이 없다(`find <tree> -maxdepth 1 -name '*.go'` 출력 없음). 그래서 `go.mod`/`go.sum`만 바뀌면 패키지 목록이 비고, `go build ./...`만 돌며 lint·test는 빠진다(254-256). | **확인.** required 검사가 의존성 변경이라는 한 부류에서 빨간불을 낼 수 없다. | blocking (Medium, 정확성) |
| C4 P2 `internal/cli/factory_release_route.go:228` | `golangci-lint` 단계도 `toolVersionCmd: go version`만 기록한다. | **확인.** lint 바이너리를 바꿔도 TTL 동안 이전 통과 결과를 재사용할 수 있다. AGENTS.md §4는 "결과가 의존하는 툴체인 식별자"를 적으라고 한다. 실제 영향 범위는 작다. | optional (Low) |

## Findings (structured defect-list)

- **C1 [Medium] [blocking] `internal/cli/factory_wait.go:35-37, 250-261`**: 대기 해제 판정이 선행 카드의 *현재* 상태가 `merged-local`/`merged-pr`일 때만 참이다. 착지를 지나 `pushed`/`ci-green`/`done`에 이른 선행 카드를 대상으로 기록한 대기(재기록 포함)는 영구히 열린 채 남는다. Required fix: 착지 판정에 착지 이후 상태(`pushed`, `ci-green`, `done`; abandoned/failed는 제외)를 넣거나, 카드의 착지 이력 이벤트를 읽는다. 각 착지 이후 상태로 선행 카드를 고정한 시험을 추가한다.
- **C3 [Medium] [blocking] `internal/cli/factory_release_route.go:248-262, 294-322`**: 패키지 소유자가 없는 Go 입력(`go.mod`, `go.sum`, 루트의 빌드 입력)이 바뀌면 빠른 필수 검사가 build만 돌린다. Required fix: `go.mod`/`go.sum`(및 `go.work`) 변경이 있으면 `./...`로 넓히거나, 범위를 정하지 못하면 검사를 보류·거절한다. go.mod만 바뀐 fixture 시험을 추가한다.
- **C2 [Medium] [optional] `internal/cli/factory_card.go:3363-3368`**: 검사 통과 뒤에 카드 임대도 같은 소유 확인 아래에서 갱신하는 편이 좋다. 지금도 안전하게 거절되므로 정확성 결함은 아니다.
- **C4 [Low] [optional] `internal/cli/factory_release_route.go:225-230`**: 단계별 실제 도구 버전(`golangci-lint --version`)을 tool identity에 넣는다.
- **O1 [Low] [optional] `internal/hook/branch_guard.go:477-489`**: 수리 전(ff2158a312)에는 `cd /wt && git -C /primary commit`을 선행 cd 채택 뒤 `-C /primary`로 다시 판정해 거절했다. 수리 뒤에는 `redirected`가 되어 훅 cwd 판정으로 넘어가므로, 훅 cwd가 worktree이면 허용된다. 같은 이유로 `git -C /primary commit -C HEAD`(commit의 `-C <commit>` 옵션)도 worktree cwd에서는 허용된다. 두 경우 모두 기준점 cb30e05b3e와 같은 동작이고(그때는 훅 cwd만 봄), "폴백은 거절만 늘린다"는 불변식은 기준점 대비로 지켜진다. 카드 중간 상태 대비로는 거절 하나가 줄었다. 개선안: `redirected`일 때도 절대 경로 `-C`가 하나라도 주 체크아웃을 가리키면 거절 쪽으로 판정한다. 같은 계열의 기존 우회(`git -c … -C`, 셸 래퍼)는 지난 회차 Residual-risk와 같다.

## Binding run conditions (SPEC progress.md 재확인)

| 부채 | 상태 | 증거 |
|---|---|---|
| D16 | 처분 | acceptance.md:444 인용이 새 이름, 시험 PASS, 장부 행 "처분됨" |
| D15 / D17 / D18 / D20 | 기록된 이월(리더 승인 D10) | progress.md:30, :208-211 |
| D22 / D23 / D24 | 처분 | 장부 기록(지난 회차와 같음) |

## Evidence

| Claim | 명령 | 관측 출력 (verbatim) |
|---|---|---|
| 감사 SHA | `git -C <tree> rev-parse --show-toplevel HEAD` | `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t2013` / `6cb0c11f325fdabbb9b80e1bfabf0df92c251ed0` |
| 델타 커밋 | `git -C <tree> log --oneline ff2158a312..6cb0c11f32` | 6줄: `6cb0c11f32 … D16 resolved`, `c4f8c6d99a …`, `f722ee92c5 [t2013] fix(hook): adopt an explicit cd or -C git target only when no later segment can redirect the command`, `733d74617c [t2013] test(hook): RED - …`, `98586ee059 [t2013] fix(cli): checkpoint prefers the live row …`, `cc7c613122 [t2013] test(cli): RED - …` |
| 델타 파일 | `git -C <tree> diff --stat ff2158a312..6cb0c11f32` | `7 files changed, 198 insertions(+), 10 deletions(-)` (Go: factory_checkpoint.go, factory_checkpoint_rows_t2013_test.go, branch_guard.go, branch_guard_target_t2013_test.go) |
| 범위 기준 조상 | `git -C <tree> merge-base --is-ancestor cb30e05b3e HEAD` | 출력 없음, 종료 0 |
| 감사계획 | `moai verify audit-plan --project-root <tree>` | `"config_status": "ok"`, `"model": "codex"`, codex `"gate": "required"`, `"cross_model_required": true`, `"enforced_required": ["codex"]` |
| hook 시험 | `go test -count=1 -v -run '^(TestBranchGuard.*\|TestProtectedCommit.*\|TestMatchProtectedCommitCommand.*)$' ./internal/hook/` | `PASS` / `ok  	github.com/modu-ai/moai-adk/internal/hook	30.976s`; 결과 파일에서 `grep -c -- '--- PASS'` → `218`, `grep -E '…\|--- FAIL'` → FAIL 줄 없음 |
| F1 대상 시험 | 같은 결과 파일에서 `grep -E -- '--- (PASS\|FAIL): TestBranchGuardExplicit…'` | `--- PASS: TestBranchGuardExplicitTargetDeniesPrimary (0.57s)` / `--- PASS: TestBranchGuardExplicitTargetAllowsWorktree (0.60s)` / `--- PASS: TestBranchGuardExplicitTargetUnchangedForeignForms (0.89s)` / `--- PASS: TestBranchGuardExplicitTargetNotAdoptedWhenLaterSegmentRedirects (0.81s)` |
| checkpoint 시험 | `go test -count=1 -v -run '^TestFactoryCheckpoint' ./internal/cli/` | `--- PASS: TestFactoryCheckpointPrefersLiveRowOverNewerAbandoned (1.69s)` / `--- PASS: TestFactoryCheckpointAbsentWorktreeReportsNoForeignGitState (1.13s)` / `--- PASS: TestFactoryCheckpointLegacyTolerant (0.00s)` / `--- PASS: TestFactoryCheckpointResumeFromFileAlone (4.63s)`(하위 3건 PASS) / `--- PASS: TestFactoryCheckpointZeroWriteWhenUnchanged (2.41s)` / `--- PASS: TestFactoryCheckpointVerbPrintsFileContent (2.17s)` / `--- PASS: TestFactoryCheckpointMissingCardWritesNothing (1.22s)` / `--- PASS: TestFactoryCheckpointNextActionTable (0.00s)` / `ok  	github.com/modu-ai/moai-adk/internal/cli	14.648s` |
| D16 시험 | `go test -count=1 -v -run '^TestLeaderReceiptGateIssuerIndependence$' ./internal/cli/` | `--- PASS: TestLeaderReceiptGateIssuerIndependence (3.50s)` / `ok  	github.com/modu-ai/moai-adk/internal/cli	4.886s` |
| D16 인용 | `grep -rn 'TestLeaderReceiptGateSelfIssued\|TestLeaderReceiptGateIssuerIndependence' <tree>/internal <acceptance.md>` | `leader_approval_gate_test.go:957`·`:958`(주석), `:963:func TestLeaderReceiptGateIssuerIndependence`, `acceptance.md:444:- green(M1): … '^TestLeaderReceiptGateIssuerIndependence$'` |
| 서식 | `gofmt -l <수리 Go 파일 4개>` | 출력 없음 |
| vet | `go vet ./internal/hook/ ./internal/cli/` | 출력 없음 |
| 종단 상태 | `grep -n 'terminalStates *=' … card_record.go` | `74:var terminalStates = []string{CardDone, CardFailed, CardAbandoned}` |
| C1 근거 | `grep -rn 'func factoryLandedState' -A 10 <tree>/internal/cli/` | `factory_wait.go:36:	return state == homestate.CardMergedLocal \|\| state == homestate.CardMergedPR` |
| C2 근거 | Read integration_merge.go:189-193 | `if card.LeaseHolder == lane && card.LeaseExpiresAt != "" {` / `leaseUnexpired = time.Now().UTC().Before(expires)` |
| C3 근거 | `find <tree> -maxdepth 1 -name '*.go'` + Read factory_release_route.go:252-256 | 출력 없음 / `pkgs := factoryChangedGoPackages(cardTree, releaseTip)` / `steps := [][]string{{"go", "build", "./..."}}` / `if len(pkgs) == 0 {` / `return steps, nil` |
| C4 근거 | Read factory_release_route.go:227-228 | `envNames:           []string{"GOFLAGS", "GOOS", "GOARCH", "CGO_ENABLED"},` / `toolVersionCmd:     []string{"go", "version"},` |
| codex 다리 | `mcp__moai__codex_audit(target=baseBranch, mode=native)` | `"verdict":"fail"`, `"audit_receipt":"rcpt-7ca077fcb8808436b4e45880"`, `"build_commit":"b1806498fe"`, `"review_base":"main (merge base cb30e05b3e5f64774b8dabd87e80c24332f4551d)"` |

## Baseline-attribution

모든 측정은 이번 실행에서 트리 `6cb0c11f325fdabbb9b80e1bfabf0df92c251ed0`을 대상으로 했다. Go 시험과 vet은 트리의 소스를 `go test`/`go vet`이 그 자리에서 빌드해 판정했으므로, 판정한 빌드가 곧 트리다. codex 다리는 MCP 서버 빌드 `b1806498fe`(설치본)가 실행했다. 판정 도구 빌드가 트리 HEAD보다 오래됐다는 점(§2.2)을 기록해 둔다. 정적 리뷰라 결과 자체에는 영향이 없다. O1의 회귀 비교는 `git diff ff2158a312..6cb0c11f32 -- internal/hook/branch_guard.go`의 제거·추가 줄과 수리 전 코드 경로(선행 cd 채택 뒤 `explicitGitTargetPattern` 재판정)를 소스로 대조한 것이다. 실행으로 재현하지는 않았다.

## Gaps

- **G1 첫 hook 시험 형태 거절.** 출력을 스크래치 파일로 보내고 `echo exit=$?`를 붙인 형태가 worktree 격리 가드에 거절됐다("construct too complex to verify"). 같은 형태로 다시 돌리지 않고, 지시대로 일반 단일 명령으로 실행했다. 출력은 도구 결과 파일에 저장됐고 그 파일에 `grep`을 돌려 셌다.
- **G2 golangci-lint·커버리지·Windows 빌드 미측정.** `go vet`과 `gofmt`만 확인했다.
- **G3 O1·C1~C4는 실행으로 재현하지 않았다.** 소스 판독으로만 확인했다(결정적 정규식·분기라 판독으로 결론이 나지만, 재현 시험은 없다).
- **G4 audit_multi 미호출.** 지시대로 `codex_audit` 단일 호출만 했다. audit_multi digest와 card second-review 기록이 없다.
- **G5 2차 범위 한정.** 카드 범위의 나머지는 수리로 생긴 회귀만 봤다. codex의 새 지적 4건 말고 카드 기존 코드를 처음부터 다시 감사하지는 않았다.
- **G6 작업 트리 미커밋 변경.** `.moai/reports/t2013/progress.md` 수정이 커밋되지 않은 채 있다. 보고서 파일이라 판정 대상 코드에는 영향이 없다고 보고 내용은 읽지 않았다.

## Residual-risk

- 선행 `cd`/`-C` 판정은 정규식 기반이다. 셸 래퍼(`bash -c`), `git -c … -C`처럼 전역 옵션이 앞에 오는 형태, 변수 확장 경로는 기존부터 가드 밖이다(fail-open 설계).
- C1·C3는 수리 전부터 카드 범위에 있던 결함이다. 1·2회차 감사가 잡지 못했으므로, 같은 기존 코드에 codex가 아직 내놓지 않은 지적이 더 있을 수 있다.
- 판정 도구 빌드(b1806498fe)가 트리보다 오래됐다.

## Iteration history

| 회차 | audited_sha | 판정 | codex receipt |
|---|---|---|---|
| 1 | 2513910f8423dc907b69f6faf3606418af1b29e2 | FAIL 0.68 | rcpt-1adafac80e1966d1648a5c24 (fail) |
| 2 | ff2158a312658cd56329e0c51f1e42272a735d4d | FAIL 0.64 | rcpt-809e05e68eaf48b16036000a (fail) |
| 3 | 6cb0c11f325fdabbb9b80e1bfabf0df92c251ed0 | FAIL 0.76 | rcpt-7ca077fcb8808436b4e45880 (fail) |

3회차는 2회차 지적 F1~F4가 모두 해소됐음을 확인했다. FAIL이 된 이유는 required codex 다리의 새 지적(C1·C3 blocking) 때문이다. 이 회차가 사전 승인된 마지막 감사이므로, C1·C3를 이 카드 안에서 고칠지 후속 카드로 넘길지는 리더가 결정할 일이다.
