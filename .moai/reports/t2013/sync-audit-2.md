auditor-model: claude-opus-5-5
verdict: FAIL
audited_sha: ff2158a312658cd56329e0c51f1e42272a735d4d
score: 0.64

# Sync 감사 2회차 — 카드 t2013 (SPEC-FACTORY-COMPLETION-RECOVERY-001 개정)

- 트리: `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t2013`, 브랜치 `WT-agent-trunk-factory`
- 감사 대상: `git rev-parse ff2158a312` → `ff2158a312658cd56329e0c51f1e42272a735d4d` (= 감사 시작 시 `git rev-parse HEAD`), 작업 트리 깨끗(`git status --short` 출력 없음)
- 범위: `cb30e05b3e..ff2158a312` (`git merge-base HEAD main` → `cb30e05b3e5f64774b8dabd87e80c24332f4551d`). 106커밋, 그중 카드 비병합 커밋 89건. 범위에는 origin/main 흡수 병합(68ea579390, 1e7fb0515b)으로 들어온 타 카드 변경(#1956, #1946 등)도 섞여 있어 diff stat은 282파일이다. 이 감사의 판단 대상은 지시받은 카드 범위(아래 9개 항목)이고, codex 다리는 diff 전체를 봤다.
- 승인: decision board `d-20261102T200838Z-5cb1`의 단일 승인 감사 1회. 재감사 없음.
- 감사계획(`moai verify audit-plan --project-root <tree>`): `config_status: ok`, `model: codex`, codex `required`, claude `off`, glm `off`, `cross_model_required: true`, `enforced_required: ["codex"]`.

## 종합 판정

**FAIL (0.64).** 판정을 가른 것:

1. required codex 다리가 `fail`을 냈다(receipt `rcpt-809e05e68eaf48b16036000a`). required 다리의 FAIL은 PASS를 막는다.
2. 그 P1을 소스로 직접 확인했다. 새 BranchGuard 대상 판정(`explicitGitTarget`)은 명령 맨 앞 `cd <abs> &&`의 경로를 git 실행 위치로 믿는데, 뒤이은 `cd`·환경변수가 실제 실행 위치를 바꿔도 그대로 믿는다. 그래서 훅 cwd가 주 체크아웃이고 `deny_commits_on: [main]`인 상태에서 `cd /wt && cd /primary && git commit -m x` 같은 명령이 **수정 전에는 거절되던 것이 이제 허용된다**. 코드 자체 주석(branch_guard.go:456 "cwd judgment can only deny, never newly allow")이 약속한 불변식을 깨는 회귀다(F1, blocking).
3. codex P2 두 건(checkpoint의 행 선택·작업트리 대체)도 소스로 확인했다. 잘못된 카드 상태·다른 체크아웃의 git 정보를 재개 상태로 기록할 수 있다(F2·F3, blocking — correctness).
4. SPEC progress.md Binding run conditions 가운데 D15·D16·D17·D18·D20 5건이 미처분이다. D15/D17/D18/D20은 리더 승인(D10)으로 "final SPEC close sync"까지 이월된 기록이 있지만, 규칙상 미처분은 미처분으로 보고하며 must-pass 실패다.

Must-pass firewall: Functionality FAIL → 다른 점수와 무관하게 전체 FAIL.

## 차원별 점수

| 차원 | 점수 | 판정 | 근거 |
|---|---|---|---|
| Functionality (40%) | 55/100 | FAIL | F1(BranchGuard 회귀), F2·F3(checkpoint 오기록), required codex 다리 fail. 9개 항목 중 owner 재임대·런처 3건·wait-aware stop-goal·pr-open 수리 재개·wake 3사건·`--ack`·PR readiness는 소스 판독상 의도대로 동작. 이 감사는 Go 시험을 돌리지 못함(Gaps G1). |
| Security (25%) | 72/100 | PASS(경계) | 새 SQL은 모두 `?` 자리표시자(card_wake_events.go, factory_checkpoint.go:259-261), 문자열 결합 SQL 0건. 비밀값 없음. 감점: F1은 주 체크아웃 보호(문서상 협조 가드, opt-in·fail-open)의 우회 경로를 새로 연다 — 무결성 문제로 Functionality에서 차단 처리. |
| Craft (20%) | 65/100 | PASS(경계) | RED→GREEN 커밋 쌍 유지, gofmt 깨끗. 감점: 대상 판정 시험이 복합 `cd`·`GIT_DIR` 경우를 다루지 않음, checkpoint 시험에 작업트리 부재·중복 run 행 경우가 없음, `--ack`가 최신 wait에만 적용(레인 Gap 5), unverified base 거절 전에 merge-check가 먼저 기록(레인 Gap 4). |
| Consistency (15%) | 78/100 | PASS | CHANGELOG `[Unreleased]` Added 첫 항목·Fixed 첫 항목이 실제 파일·동작과 일치(CHANGELOG.md:12, :24 판독). §E.4 개정 블록이 `sync_status: complete-pending-audit`, `spec_md_status: in-progress (unchanged)`로 정직하게 기록. 감점: D16 장부 행이 갱신되지 않음, 규칙 템플릿 미러 대조는 가드 거절로 미측정(G3). |

가중 조화평균: 0.40/55 + 0.25/72 + 0.20/65 + 0.15/78 = 0.015745 → 63.5 → **0.64**.

## Codex 다리 (mcp__moai__codex_audit, 1회)

- 호출: `codex_audit(project_root=/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t2013, target=baseBranch, mode=native)`
- 결과: `verdict: fail`, `audit_receipt: rcpt-809e05e68eaf48b16036000a`, `build_commit: b1806498fe`, `review_base: main (merge base cb30e05b3e5f64774b8dabd87e80c24332f4551d)`. 별도 job id 필드는 결과에 없음 — receipt id가 식별자다.
- `gate_unmet`: 결과에 필드 없음(verdict 자체가 `fail`).
- codex 지적과 이 감사의 독립 확인:

| codex | 내용 | 이 감사의 판단 |
|---|---|---|
| P1 `internal/hook/branch_guard.go:463` | 복합 명령에서 첫 `cd /wt`를 대상으로 채택하지만 실제 실행은 뒤 `cd /primary` | **확인(소스 판독)** → F1 blocking |
| P2 `internal/cli/factory_checkpoint.go:350` | `UpdatedAt` 최대 행 선택이 rebind로 abandoned된 source 행을 고를 수 있음 | **확인(소스 판독)** → F2 blocking |
| P2 `internal/cli/factory_checkpoint.go:274` | 작업트리 부재 시 `base = root`로 다른 체크아웃의 HEAD·브랜치·dirty·unpushed를 기록 | **확인(소스 판독)** → F3 blocking |

이 감사는 `audit_multi`를 호출하지 않았다(지시가 codex_audit 단일 호출을 명시했고 승인된 감사는 1회). 따라서 audit_multi digest(`overall_verdict`/`plan_source`/`per_backend_verdicts`)는 없다. 오케스트레이터가 `moai verify audit-plan --result-file`로 확인할 경우 넣을 값은 per-backend {backend: codex, gate: required, verdict: fail}뿐이며, 그 확인이 끝나기 전까지 이 판정은 최종이 아니다.

## Findings (structured defect-list)

- **F1 [High] [blocking] `internal/hook/branch_guard.go:447,458-474,545,1111`** — `explicitGitTarget`은 `^\s*cd\s+(/…)\s*&&` 한 번만 떼어 내고 나머지 명령을 다시 보지 않는다. git 호출이 하나이기만 하면 그 경로를 `gitContextCwd`로 쓴다(545, 1111행). 실제 실행 위치를 바꾸는 후속 토큰 — 두 번째 `cd /primary &&`, `cd ..`, `pushd`, `(cd …; git …)`, `cd /x; git …`, `GIT_DIR=`/`GIT_WORK_TREE=` 접두 — 이 있어도 `/wt`(비주 체크아웃)로 판정해 `checkBranchState`와 `checkProtectedCommit` 둘 다 허용한다. 수정 전(cb30e05b3e)에는 같은 명령이 훅 cwd(주 체크아웃) 판정으로 거절됐으므로 순수 회귀다. 456행 주석 "falling back to the cwd judgment can only deny, never newly allow"는 폴백 쪽만 말하고, 대상 채택 쪽이 새 허용을 만드는 경로는 막지 않는다. 시험(`branch_guard_target_t2013_test.go:28-107`)에 이런 복합 형태가 없다. Required fix: 선행 `cd` 대상은 나머지 명령에 추가 디렉터리 변경(`cd`, `pushd`, `popd`, 괄호 서브셸, `;`/`||` 뒤 `cd`)이나 저장소 지정 환경변수·옵션(`GIT_DIR`, `GIT_WORK_TREE`, `--git-dir`, `--work-tree`)이 없을 때만 채택하고, 아니면 기존 cwd 판정으로 떨어뜨린다. 위 형태마다 주 체크아웃 cwd에서 거절을 고정하는 시험을 추가한다.
- **F2 [Medium] [blocking] `internal/cli/factory_checkpoint.go:344-354`** — `ListCards(ctx, "")`로 모든 run의 행을 읽고 `UpdatedAt`이 가장 큰 행을 현재 카드로 고른다. `cards`의 키는 `(run_id, card_id)`(`internal/homestate/factory.go:83`)라 한 카드가 여러 run에 행을 가질 수 있고, `abandonWithPointer`는 source 행을 `abandoned`로 바꾸면서 `updated_at`을 지금으로 갱신한다(`internal/homestate/factory_run_rebind.go` 517행대 `UPDATE cards SET state=?, version=version+1, updated_at=?`). 살아남은 target 행이 그 뒤 갱신되지 않았으면 abandoned 행이 선택되어 상태·작업트리·다음 조치가 틀리게 기록된다. 시험(`factory_checkpoint_t2013_test.go`)에 중복 행 경우가 없다. Required fix: dispatch 바인딩이 가리키는 run의 행을 읽거나, 최소한 종단 상태(`done`/`failed`/`abandoned`) 행을 후보에서 빼고 비종단 행을 우선한다. 중복 행 fixture 시험을 추가한다.
- **F3 [Medium] [blocking] `internal/cli/factory_checkpoint.go:273-305,358-364`** — 기록된 `WorktreePath`가 없거나 디렉터리가 아니면 `base = root`로 바꾼 뒤 그 위치에서 `rev-parse HEAD`·`branch --show-current`·`status`·`rev-list @{u}..HEAD`를 실행해 **주 체크아웃의** 값을 카드 값으로 기록하고, checkpoint 파일도 주 체크아웃 `.moai/reports/<card>/`에 쓴다. 이 checkpoint가 재개 판단의 근거(`TestFactoryCheckpointResumeFromFileAlone`)이므로 오정보가 다음 조치로 이어진다. Required fix: 저장 위치 폴백과 git 조회 위치를 분리하고, 카드 작업트리에서 읽지 못한 Branch/Head/Dirty/Unpushed는 `unknown`으로 남긴다. 작업트리 부재 시험을 추가한다.
- **F4 [—] [blocking] Binding run conditions** — D15·D16·D17·D18·D20 미처분(아래 표). must-pass 조건.
- **F5 [Low] [optional] `internal/cli/factory_wait.go` `factoryAckWake`** — `--ack`는 카드의 최신 wait(`waits[len(waits)-1]`, `LoadCardWaits`는 `ORDER BY seq`)에만 적용된다. 최신 wait가 아직 배달 전이고 이전 wait가 배달된 경우 순서 거절로 끝난다. 레인 Gap (5)와 같음. 선택자(`--wait <seq>`) 추가는 선택 사항.
- **F6 [Low] [optional] `internal/cli/factory_card_pr.go:281-291`** — unverified base 거절 전에 `RecordMergeCheckRun`이 먼저 실행되어 거절된 readiness가 기록으로 남는다. 레인 Gap (4)와 같음.
- **F7 [Low] [optional] `internal/cli/factory_card_pr.go:256-262`** — `behind N` 안내는 출력만 하고 ready 전환·CI 재실행을 막는 코드가 없다(레인 Gap 3). 문구가 "waits until the head contains it"라고 하지만 강제는 없다. 문구를 사실에 맞추거나 기록 필드로 남길 것.
- **F8 [Low] [optional] `internal/cli/hook_stop_goal_wait.go:83-87`** — 열린 wait에 `next=`만 있으면 그 사건이 영영 오지 않아도 목표 차단이 꺼진다. 상한(TurnsUsed·wall-clock·stagnation)은 그대로 작동하므로 무한 정지는 아니다. 관찰 기록만.

## Binding run conditions (SPEC progress.md 장부 재확인)

| 부채 | 상태 | 증거 |
|---|---|---|
| D15 | 미처분 | 장부 행 그대로, D10으로 final SPEC close sync 이월 |
| D16 | 미처분(부분 진행) | `grep -rn TestLeaderReceiptGateSelfIssued internal` → `internal/cli/leader_approval_gate_test.go:958` 주석만 남고 함수는 `TestLeaderReceiptGateIssuerIndependence`로 개명됨(963행). 그러나 `grep -c TestLeaderReceiptGateSelfIssued acceptance.md` → `1`, `grep -c TestLeaderReceiptGateIssuerIndependence acceptance.md` → `0`이라 acceptance.md 인용이 낡았고 장부 행도 갱신 안 됨. 1회차 감사의 "개명 안 됨" 판단은 이번 측정으로 정정한다. D10 이월 목록에도 없음. |
| D17 | 미처분 | D10 이월 |
| D18 | 미처분 | D10 이월 |
| D20 | 미처분 | D10 이월 |
| D22 / D23 / D24 | 처분 | 장부 "처분됨" 기록 |

## 범위별 점검 (소스 판독)

| 항목 | 판단 |
|---|---|
| owner 만료 임대 재임대 (`card_lease_expired_owner.go`, `factory_card.go` `factoryNextReacquire`) | 조건(leased·만료·holder=owner=label)과 한 트랜잭션(expiry→T3→`lease.reacquired`) 확인. 문제 없음. |
| 런처: teammateMode (`launcher.go:341-363`) | GLM env 잔재가 있을 때만 삭제, 잠긴 재판독에서도 재확인. 문제 없음. |
| 런처: 절대경로 `-w` (`launch_view.go:232-302`, `cc.go:304-310`, `glm.go:349-353`) | `git worktree list --porcelain` 기록과 심볼릭 해석 경로 비교, prunable 제외, 쓰기 전 거절. 문제 없음. |
| 런처: 죽은 라벨 재합류 (`factory.go:393-404, 949-996`) | 정규 `lane-<n>`만 허용, live면 claim 거절. 임대 무접촉. 문제 없음. |
| wait-aware stop-goal (`hook_stop_goal_wait.go`, `goal/evaluate.go:467-476`) | 모든 상한 판정 뒤에만 참조, 읽기 전용 DB, 실패 시 차단 유지. F8 관찰만. |
| factory checkpoint | F2·F3 |
| BranchGuard 동사 경계 (`branch_guard.go:440`) | `(?:[^\w-]|$)` 종결로 `commit-tree`/`commit-graph` 제외, `;`·`&&`·`|`·`)`·공백 뒤는 유지. 문제 없음. |
| BranchGuard 명시 대상 | F1 |
| pr-open 같은 owner 수리 재개 (`card_pr_open_repair.go`, `factory_card.go:1941-1947, 2079-2084`) | owner만, pr-open→assigned(run)→leased 한 트랜잭션, 재완료 시 기존 PR 재사용(`factory_card_pr.go:484-498`), 비강제 push. 문제 없음. |
| wake 3사건 · `--ack` (`card_wake_events.go`, `card_wait.go:172-175`, `factory_wait.go`) | delivered는 settle 트랜잭션 안, received/started는 순서 강제·중복 무쓰기. F5. |
| PR readiness (`factory_card_pr.go:234-291`) | fetch 실패 + 추적 ref 존재 → `stale-base` 거절, 추적 ref 없음 → unverified 거절, `gitio.IsAncestor(wt, ref, HEAD)` 인자 순서 올바름(gitio.go:73-77). F6·F7. |
| CHANGELOG · §E.4 | 일치. |

## Evidence

| Claim | 명령 | 관측 출력 (verbatim) |
|---|---|---|
| 감사 SHA | `git rev-parse ff2158a312` | `ff2158a312658cd56329e0c51f1e42272a735d4d` |
| 범위 기준 | `git merge-base HEAD main` | `cb30e05b3e5f64774b8dabd87e80c24332f4551d` |
| 범위 크기 | `git diff --stat cb30e05b3e..ff2158a312 \| tail -1` | `282 files changed, 35155 insertions(+), 662 deletions(-)` |
| 감사계획 | `moai verify audit-plan --project-root <tree>` | `"config_status": "ok"`, `"model": "codex"`, codex `"gate": "required"`, `"enforced_required": ["codex"]`, `exit=0` |
| codex 다리 | `mcp__moai__codex_audit(target=baseBranch, mode=native)` | `"verdict":"fail"`, `"audit_receipt":"rcpt-809e05e68eaf48b16036000a"`, `"build_commit":"b1806498fe"`, `"review_base":"main (merge base cb30e05b3e5f64774b8dabd87e80c24332f4551d)"` |
| F1 회귀 대상 코드 | `git diff cb30e05b3e..ff2158a312 -- internal/hook/branch_guard.go` | `+	leadingCdTargetPattern   = regexp.MustCompile(`^\s*cd\s+(/[^\s;&\|"'()<>]+)\s*&&\s*`)` / `+	if m := leadingCdTargetPattern.FindStringSubmatchIndex(rest); m != nil {` / `+		target, rest = rest[m[2]:m[3]], rest[m[1]:]` / `+	if target != "" {` / `+		gitContextCwd = target` |
| F1 시험 공백 | `grep -n 'cd \|func Test' internal/hook/branch_guard_target_t2013_test.go` | `37:		"cd " + fx.primary + " && git commit -m x",` / `55:		"cd " + fx.release + " && git commit -m x",` / `87:		`moai todo add "cd ` + fx.release + ` && git commit"`,` (복합 `cd` 형태 없음) |
| F2 행 키 | `grep -n 'PRIMARY KEY' internal/homestate/factory.go` | `83:  PRIMARY KEY(run_id, card_id)` |
| F2 abandon이 updated_at 갱신 | `sed -n 505,560p internal/homestate/factory_run_rebind.go \| grep -n 'updated_at\|UPDATE'` | `13:	r, err := tx.ExecContext(ctx, `UPDATE cards SET state=?, version=version+1, updated_at=? WHERE run_id=? AND card_id=? AND version=?`,` |
| F3 폴백 | Read `factory_checkpoint.go:273-305` | `base := card.WorktreePath` / `if st, err := os.Stat(base); base == "" \|\| err != nil \|\| !st.IsDir() {` / `base = root` / `if head, ok := checkpointGit(base, "rev-parse", "HEAD"); ok && head != "" {` |
| checkpoint 시험 공백 | `grep -n 'WorktreePath\|abandon\|Rebind' internal/cli/factory_checkpoint_t2013_test.go` | (출력 없음) |
| 서식 | `gofmt -l <카드 운영 파일 13개>` | (출력 없음) `gofmt_exit=0` |
| SQL 결합 탐침 | `grep -nE 'Sprintf\|\+ *"' internal/homestate/card_wake_events.go \| grep -iE 'select\|update\|insert'` | (출력 없음) `sql_concat_probe_exit=1` |
| 비밀값 탐침 | 카드 diff에서 `password\|secret\|token *=` 추가 줄 | `10525:+	const token = "v9.9.9"` (시험용 버전 문자열, 비밀 아님) |
| D16 | `grep -c TestLeaderReceiptGateSelfIssued .../acceptance.md` / `grep -c TestLeaderReceiptGateIssuerIndependence .../acceptance.md` | `1` / `0` |
| 부하(시험 보류 근거) | `sysctl -n vm.loadavg` (3회) | `{ 100.90 90.95 60.82 }` / `{ 89.26 88.86 60.60 }` / `{ 76.79 85.55 60.99 }`, 다른 Go 프로세스 `6` |

## Baseline-attribution

모든 측정은 이번 실행, 트리 `ff2158a312658cd56329e0c51f1e42272a735d4d`(깨끗)에서. 회귀 판단(F1)의 이전 동작은 `cb30e05b3e` 쪽 diff 제거 줄(`-	suffix, matched := matchBranchStateCommand(command)` 등)과 수정 전 `gitContextCwd := resolveProjectRootFromInputOrEnv(...)`만 쓰던 경로를 소스로 대조했다. codex 다리는 MCP 서버 빌드 `b1806498fe`(설치본, 트리 HEAD의 조상)가 실행했다 — 판정 도구의 빌드가 트리보다 오래됐다는 §2.2 사실을 기록한다(정적 리뷰라 결과 자체엔 영향 없음).

## Gaps

- **G1 Go 시험 미실행.** 지시는 범위 지정 Go 시험을 허용했으나, 측정한 부하가 76~101(다른 Go 프로세스 6개)로 이 저장소의 자원 진입 기준(load ≤ 32)을 넘어 한 건도 돌리지 않았다. Functionality의 동작 증거는 레인 기록(`.moai/reports/t2013/progress.md` 86행, HEAD 040ecb5c2d 시점)뿐이며 이 감사가 재관측하지 않았다. golangci-lint·`go vet`·Windows 빌드·커버리지도 미측정.
- **G2 정규식 에뮬레이션 거절.** F1을 실행으로 확인하려고 스크래치 디렉터리에서 정규식 사본을 `go run`하는 명령이 worktree 격리 가드에 거절됐다("too complex to verify that it stays inside the worktree"). 다른 형태로 재실행하지 않았고, F1은 소스 판독(결정적 정규식)으로만 확인했다.
- **G3 규칙 미러·CHANGELOG diff 거절.** `cmp`로 라이브/템플릿 규칙 3개를 대조하고 `git diff 2513910f84..ff2158a312 -- CHANGELOG.md`를 함께 본 복합 명령이 가드에 거절됐다. 재실행하지 않았다. CHANGELOG는 Read 도구로 12·24행을 직접 읽어 대체했고(판독, 측정 아님), 미러 동일성은 미측정이다.
- **G4 audit_multi 미호출.** 지시대로 `codex_audit` 단일 호출만 했다. audit_multi digest와 card second-review 기록은 없다.
- **G5 범위 혼입.** codex 다리는 main 흡수분까지 포함한 diff 전체를 봤다. 이 감사의 독립 판독은 카드 9개 항목에 한정했고 흡수된 타 카드 코드(jev, jevledger, release.sh 등)는 보지 않았다.
- **G6 레인 Gap 승계.** 레인이 기록한 하위 에이전트의 가드 거절 명령 재실행(progress.md 88행 (1))과 040ecb5c2d 시점 리터럴 검사 미검증((2))은 이 감사도 재검증하지 않았다.

## Residual-risk

- F1 외에 같은 계열의 기존 우회(`git -C <primary>` 뒤 동사, `--git-dir`, `update-ref`, `push . sha:main`)는 수정 전부터 가드 밖이었다(레인 기록 (e)). F1 수리로 닫히지 않는다.
- pr-open 수리 재개는 "PR을 연 뒤 큐 항목이 `picked`로 남는다"는 가정(레인 기록 87행)에 기대며 모든 종료 경로를 추적하지 않았다.
- 시험을 돌리지 않았으므로, 레인 보고 이후(040ecb5c2d → ff2158a312, 문서 2커밋) 사이 회귀가 없다는 판단은 그 두 커밋이 Go 파일을 건드리지 않았다는 diffstat 판독(`.moai/reports/t2013/progress.md`, SPEC progress.md, CHANGELOG.md)에 기댄다.
- 판정 도구 빌드(b1806498fe)가 트리보다 오래됐다.

## Iteration history

| 회차 | audited_sha | 판정 | codex receipt |
|---|---|---|---|
| 1 | 2513910f8423dc907b69f6faf3606418af1b29e2 | FAIL 0.68 | rcpt-1adafac80e1966d1648a5c24 (fail) |
| 2 | ff2158a312658cd56329e0c51f1e42272a735d4d | FAIL 0.64 | rcpt-809e05e68eaf48b16036000a (fail) |

1회차 지적 F1(dirty 트리)·F3(착지된 선행 대기)·F4(실제 충돌 도달)·F5(testdata 매핑)·F6(hoist)은 progress.md 26행의 RED/GREEN 커밋으로 처리됐다고 기록되어 있다. 이번 감사는 F1(dirty 관문, `factory_card.go:3349-3356`)과 F3(`factory_wait.go` `factoryWakeIfPredecessorLanded`)의 코드 존재만 판독으로 확인했다.
