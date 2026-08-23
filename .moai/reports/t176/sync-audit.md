# sync-audit — SPEC-DEPLOY-RESULT-WIRE-001 (카드 t176)

## 고정 상태 (감사 시작·종료 시점 동일)

| 항목 | 값 | 명령 |
|---|---|---|
| HEAD | `ba08d8694` | `git log --oneline -1` |
| 브랜치 | `WT-deploy-result-wire` | `git branch --show-current` |
| 워크트리 | `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t176` | `git rev-parse --show-toplevel` |
| 트리 | clean (무출력) | `git status --short` |

감사 종료 시점 재측정 결과 HEAD·트리 동일 — **감사 도중 개정 없음**. 선행 카드에서 감사를 오염시킨 형태는 재발하지 않았다.

## 판정

**Overall Verdict: PASS — 91/100**

| Dimension | Score | Verdict | Evidence (verbatim) |
|---|---|---|---|
| Functionality (40%) | 95/100 | PASS | `go test -count=1 ./internal/mirrornotice/...` → `ok  github.com/modu-ai/moai-adk/internal/mirrornotice  0.369s`<br>`go test -count=1 -timeout 300s ./internal/core/project/...` → `ok  github.com/modu-ai/moai-adk/internal/core/project  1.463s`<br>`go test -count=1 -timeout 900s ./internal/cli/` → `ok  github.com/modu-ai/moai-adk/internal/cli  325.510s` (4회 중 3회; F-01 참조) |
| Security (25%) | 92/100 | PASS | `grep -rn "\.Deploy(ctx" internal/ --exclude=*_test.go` → 2행(`initializer.go:375` fallback, `mirror_notice.go:37` fallback) — 미배선 프로덕션 호출부 0건. 포맷 문자열 주입 없음: `fmt.Fprintln(w, line)` / `fmt.Fprintf(&b, "\n  %d. %s", i+1, msg)` 모두 인자 전달 |
| Craft (20%) | 82/100 | PASS | `go test -cover ./internal/mirrornotice/` → `coverage: 91.7% of statements`<br>`golangci-lint run ./internal/cli/... ./internal/core/project/... ./internal/mirrornotice/...` → `0 issues.`<br>`go vet` (동일 범위) → `VET_EXIT=0`; `GOOS=windows go vet` → `WIN_VET_EXIT=0` |
| Consistency (15%) | 93/100 | PASS | `git diff --stat 9e5c4981b..ba08d8694 -- internal/template/ \| wc -l` → `0` (선행 SPEC 표면 무변경). seam 관용·AST 가드 선례·에러 래핑·주석 영어 모두 저장소 관행과 일치 |

가중 조화평균 = 91.1. must-pass 2종(Functionality / Security) 각각 독립 통과.

**blocking 결함 0건.** 아래 결함은 전부 optional 이며, 그중 F-01 은 이 카드 소관이 아닐 가능성이 높다.

## 독립 재도출 vs 보고 인용

**독립 재도출(9/9)** — AC-DRW-001..009 전부 소스·테스트를 직접 읽고 산술·의미를 재계산했다. 위증 검사 13건 중 **연역으로 재검증한 것은 산술이 결정적인 F1·F2·FB·FK·FJ 계열**이고, 나머지는 코드 구조로 확인했다(읽기 전용 제약상 구현을 실제로 깨뜨리지는 않았다 — §Gaps).

**보고 인용(재실행하지 않음)**: `internal/core/project` 88.4% · `internal/cli` 78.5% 커버리지, RED 증거 3건, 선행 `AC-CSC-010` 재실행 결과. `internal/mirrornotice` 91.7% 는 독립 재측정으로 일치 확인.

## 리드가 지목한 4건에 대한 판정

### (a) `AC-DRW-009` 3번 단언 — 개수 고정이 아닌 전수 수집인가 → **PASS**

`update_mirror_notice_test.go` `TestCleanReinstallOptionsLiteralsInjectErrOut` 는 `ast.Inspect` 로 `CompositeLit` 를 순회하며 타입 이름이 `CleanReinstallOptions` 인 것을 **슬라이스에 누적**한다. 고정 개수 비교는 어디에도 없고, `len(literals) == 0` 일 때 `t.Fatal` 로 비어 있는 순회(가드가 엉뚱한 파일을 보는 상태)를 막는다. `&CleanReinstallOptions{...}` 형태도 `UnaryExpr` 안의 `CompositeLit` 로 잡힌다. 네 번째 호출부가 `update.go` 에 생기면 자동으로 판정 대상이 된다.

**단, 잔존이 하나 있다**: 가드는 `parseCLIFile(t, "update.go")` 로 **`update.go` 한 파일만** 판다. 실측상 프로덕션 리터럴은 현재 `update.go:418` · `update.go:632` 두 곳뿐이고(`grep -rn "CleanReinstallOptions{" internal/` — 나머지 25건 전부 `_test.go`), AC 본문도 "`internal/cli/update.go` 안의" 로 파일을 명시하므로 **구현은 AC 를 정확히 따랐다**. 결함은 AC 쪽 범위 설정에 있고, 다른 파일에 세 번째 프로덕션 호출부가 생기면 가드 밖으로 빠진다. → F-02.

### (b) `AC-DRW-004` 세 팔 — 각각 다른 잘못된 구현에서 붉어지는가, 공허하지 않은가 → **PASS**

구현의 산술(`internal/mirrornotice/notice.go`): 복사 요약 1줄 고정, `failed` 요약 1줄 + 예시 `min(failed, 3)` 줄.

| 잘못된 구현 | 1팔 (copy 2 vs 34) | 2팔 (failed 4 vs 34) | 3팔 (상한 2→2 / 34→3) |
|---|---|---|---|
| 복사를 스킬별 1줄씩 출력 | **RED** (2 vs 34) | green | green |
| `failed` 예시 상한 제거 | green | **RED** (5 vs 35) | **RED** (34 ≠ 3) |
| 상한 = 4 | green | green (5 vs 5) | **RED** (4 ≠ 3) |
| 상한 = 1 | green | green (2 vs 2) | **RED** (1 ≠ 2) |
| `failed` 예시를 아예 안 냄 | green | green (1 vs 1) | **RED** (0 ≠ 2) |

세 팔 각각 **자기만 잡는 잘못된 구현**을 갖는다(1팔: 복사 비례화 / 2팔: 상한 제거 / 3팔: 상한 4). 공허한 팔 없음 — 어느 팔도 모든 구현에서 항상 참이 아니다.

**N2 가 지목한 함정이 실제로 회피됐음을 확인**: 2팔의 낮은 표본이 `entries(MirrorModeFailed, 4)` 로 상한 3 **위**에 있어 올바른 구현이 `1+3 = 1+3` 으로 통과한다. 테스트 주석이 그 이유(`at 2 a correct implementation legitimately emits fewer lines than at 34 (1+2 vs 1+3)`)를 명시해 다음 편집자가 2로 되돌리는 것을 막는다.

3팔의 예시 개수 세는 방식(`strings.Contains(l, "permission denied")`)은 요약 줄과 예시 줄을 정확히 가른다 — 요약 문구 두 종("copied instead of linked", "could not be mirrored")에 그 부분식이 없다.

### (c) `AC-DRW-005` — 맨 타입 단언을 잡는가 → **PASS**

세 팔 전부 `ok` 없는 단언에서 panic 으로 붉어진다:

- `internal/cli/mirror_notice.go:35` `rd, ok := dep.(template.ResultDeployer)` → 맨 단언으로 바꾸면 `TestCleanReinstall_PlainDeployerStillDeploys` · `TestTemplateSync_PlainDeployerStillDeploys` 가 panic.
- `internal/core/project/initializer.go:364` `if rd, ok := i.deployer.(template.ResultDeployer); ok` → `TestInit_PlainDeployerStillDeploys` 가 panic.

이중체가 확장 부재 상태를 **실제로 재현**함도 확인했다: `plainDeployer`(`update_mirror_notice_test.go:66`)와 `capturingDeployer`(`initializer_test.go:45`) 모두 `DeployWithResult` 메서드를 **정의하지 않는다**(`grep -rn "func (.*capturingDeployer)"` → `Deploy` / `ExtractTemplate` / `ListTemplates` / `ValidateAll` 4개뿐). 두 테스트 모두 `any(dep).(template.ResultDeployer)` 로 그 사실을 **런타임에 먼저 단언**하므로, 나중에 누가 `DeployWithResult` 를 추가해 판정을 조용히 공허하게 만드는 것도 막힌다. `acceptance.md` 의 `[HARD]`("컴파일 시점에 만족하지 않아야 한다")는 런타임 단언으로 충족됐다 — 효력 동일.

### (d) 신규 패키지 `internal/mirrornotice` — 인가된 범위인가 → **인가됨 (plan 이 아니라 spec §D 로)**

- **plan 은 인가하지 않았다.** plan M1 은 "순수 함수 하나" 라고만 적고 배치처를 말하지 않는다.
- **spec §D 가 인가한다.** `Out of Scope — 구현 세부` 가 "헬퍼 함수 이름과 **배치**는 run-phase 판단" 이라고 명시한다. 패키지 선택은 배치 결정이므로 spec 이 명시적으로 run-phase 에 넘긴 사항이다.
- **구조적으로 강제된다**(리드의 추정 확인): `internal/cli` 가 `internal/core/project` 를 import 하므로(`update_template_sync.go` 의 `project.ProgressReporter`), 역방향 import 는 순환이다. 세 호출부가 같은 문구를 쓰려면 두 패키지가 함께 의존할 수 있는 제3의 패키지가 **필요하다**.

따라서 미신고 범위 확장이 아니다. 다만 plan 의 Tier M 예상(변경 파일 6-8개)에 대해 실제 run-phase 파일은 9개다 → F-03.

## 그 밖의 결함 계열 (리드 지목)

### 스트림 규율 → **PASS**

`update.go` 의 프로덕션 리터럴 **둘 다** `ErrOut: cmd.ErrOrStderr()` 를 싣는다(`:425` 인접 / `:632` 인접, diff 실측). `Out` 의 의미는 바뀌지 않았다(둘 다 `out` = `cmd.OutOrStdout()` 유지). init 경로는 새 출력 표면을 만들지 않고 기존 `InitResult.Warnings` → `p.Collect`(`init.go:707`) → `emitSummary(cmd.ErrOrStderr())` 통로를 그대로 쓴다 — `warnCollector.Collect` 는 즉시 재출력하지 않고 요약 패널 1회 방출만 하므로(`init_warnings.go:52`), 통지가 stdout 에 새는 경로가 없다.

`emitDryRunReinstallPlan` 의 시그니처가 `(ctx, cwd, force, out, th)` → `(cmd, cwd, force, th)` 로 바뀌었다. 동작은 동등하다(`ctx := cmd.Context()`, `out := cmd.OutOrStdout()` 가 호출부에서 넘기던 값과 같은 값). 다만 이 함수는 전용 테스트가 없어(`grep -rn "emitDryRunReinstallPlan" internal/cli/` → `update.go` 외 참조 0건) 변경이 컴파일·vet 로만 커버된다 → F-04.

### 비비례성 실전 확인 → **PASS**

34건 fixture 로 실제 줄 수를 셌다(`TestLines_*NotProportional`, `entries(mode, 34)`). 사용자가 보는 최대 줄 수 = 복사 요약 1 + `failed` 요약 1 + 예시 3 = **4줄**, 스킬 수와 무관. init 경로에서는 이 4줄이 `emitSummary` 의 번호 매김 항목 4개가 된다 — 여전히 유한 상한. 헬퍼 반환값이 아니라 사용자 도달 지점(캡처된 stderr / `InitResult.Warnings` 슬라이스)에서 판정된다.

### 오귀속 문구 보류 → **PASS**

`Lines` 의 `switch` 가 `MirrorModeCopy` / `MirrorModeFailed` 만 계수하고 `MirrorModeSkipped` 는 **어느 분기도 타지 않는다** — `Warnings()` 를 통째로 forward 하는 형태가 아니다. 3팔(`mirrornotice` · clean-reinstall · init)이 `a non-symlink entry already exists` 부재를 각각 단언한다. 문구 원본은 `skill_mirror.go:205` 에 그대로 있고(재측정: `grep -n "non-symlink entry already exists"` → `205`), 이 카드가 고치지 않았다 — spec §D 의 결정과 일치.

### 범위 규율 → **PASS**

`Deploy` 시그니처 무변경, `ResultDeployer` 필수화 없음(`deployer.go:64` 인터페이스 정의 무변경), `internal/template` diff 0줄, 미러 로직·정리 경로(t173 소관) 무변경. `git diff --stat 9e5c4981b..ba08d8694 -- internal/template/` → 무출력.

### 기록된 부채(§D.6) → **여전히 정확**

"AST 가드는 리터럴을 증명하지 런타임 값을 증명하지 않는다" 는 구현 후에도 참이다. 완화 3종이 실제로 서 있다: 2번 단언은 런타임(`newTemplateSyncDeployer(fstest.MapFS{})` 를 실제로 호출), `AC-DRW-003` 이 같은 배선을 행동으로 판정, 3번 단언이 리터럴을 본다. 남는 구멍은 §D.6 이 적은 그대로 "cobra 의 `ErrOrStderr()` 가 stderr 가 아닌 경우" 뿐이다. **부채 서술 갱신 불필요.**

## Findings

- **F-01** [medium] [optional] `internal/cli` 패키지 — `go test -count=1 -timeout 900s ./internal/cli/` 4회 실행 중 **1회 FAIL**(`FAIL github.com/modu-ai/moai-adk/internal/cli 323.122s`), 이후 3회 연속 `ok`(398.078s / 325.510s / 2.910s 부분 실행 포함). 신뢰도: 실패 자체는 확실(관측), **이 카드 귀속은 낮다**. 첫 실행에서 출력을 `tail -40` 으로 잘라 실패 테스트 이름을 못 건졌고, 이 카드가 더한 테스트만 `-count=10` 으로 돌린 결과는 `ok ... 2.910s`(10/10 안정). **Required fix**: 없음(이 카드에서 고칠 대상이 특정되지 않는다). 후속 조치로 `internal/cli` 전량을 `-count=1` 로 몇 회 돌려 실패 테스트 이름을 포획할 것 — 잡히면 별도 카드. 이 카드의 PASS 를 막지 않는 이유는 신규 테스트가 격리 반복에서 안정하고, 신규 코드가 goroutine·전역 상태를 만들지 않기 때문이다.
- **F-02** [low] [optional] `internal/cli/update_mirror_notice_test.go` `parseCLIFile(t, "update.go")` — AC-DRW-009 3번 가드가 `update.go` **한 파일 범위**다. 리터럴 개수는 고정하지 않았으므로 리드가 지목한 형태의 결함은 없으나, 프로덕션 `CleanReinstallOptions` 리터럴이 `internal/cli` 의 **다른 파일**에 생기면 판정 밖이다. **Required fix**(후속 카드 권장): `filepath.Glob("*.go")` 로 비-테스트 파일 전체를 순회하도록 넓히고, `acceptance.md` AC-DRW-009 3번의 "`internal/cli/update.go` 안의" 를 "`internal/cli` 프로덕션 소스의" 로 함께 고칠 것. 지금 고치지 않는 근거: 실측상 프로덕션 리터럴은 `update.go` 두 곳뿐(`grep -rn "CleanReinstallOptions{" internal/` 전수).
- **F-03** [low] [문서 정확성] `.moai/specs/SPEC-DEPLOY-RESULT-WIRE-001/progress.md` §E.3 `total_run_phase_files: 8` — 실측 **9**. `git diff --numstat 9e5c4981b..ba08d8694` 12파일 중 SPEC 문서 2(`spec.md`, `progress.md`)와 `CHANGELOG.md` 1 을 빼면 9(`notice.go`, `notice_test.go`, `mirror_notice.go`, `update.go`, `update_clean_install.go`, `update_template_sync.go`, `update_mirror_notice_test.go`, `initializer.go`, `initializer_mirror_notice_test.go`). **Required fix**: `8` → `9`.
- **F-04** [low] [optional] `internal/cli/update.go` `emitDryRunReinstallPlan` — 시그니처를 `(ctx, cwd, force, out, th)` → `(cmd, cwd, force, th)` 로 바꿨다. 동작 동등이나 이 카드가 요구한 최소 변경(파라미터 1개 추가)보다 넓고, 이 함수에는 전용 테스트가 없어 컴파일·vet 외 근거가 없다. **Required fix**: 없음(회귀 근거 없음). 기록만 남긴다 — 다음 편집자가 "cobra 커맨드를 헬퍼로 내리는 것" 을 이 저장소의 관행으로 오독하지 않도록.
- **F-05** [low] [optional] `progress.md` §E.3 `run_commit_sha: pending-backfill-run` — M5 커밋 `ba08d8694` 가 존재하는데 백필되지 않았다. **Required fix**: `ba08d8694` 로 치환.
- **F-06** [info] [optional] `internal/mirrornotice/notice_test.go` · `update_mirror_notice_test.go` 의 개수 단언이 `strings.Contains(joined, "12")` / `"11"` / `"9"` 형태다. 현 fixture 에서는 다른 출처의 숫자와 충돌하지 않음을 확인했으나(복사 모드는 예시 문구를 내지 않고, `failed` 예시의 스킬 인덱스가 개수와 겹치지 않는다), 문구가 바뀌면 조용히 약해질 수 있다. **Required fix**: 없음. 후속 편집 시 `fmt.Sprintf("%d skill(s)", n)` 부분식으로 좁힐 것.
- **F-07** [info] [optional] `internal/cli/update_mirror_notice_test.go` `runTemplateSyncCapturing` 이 `os.Chdir(tmpDir)` 로 **프로세스 전역 cwd** 를 바꾼다. 같은 파일의 `TestSeamDefaultIsTheProductionDeployer` · `TestCleanReinstallOptionsLiteralsInjectErrOut` 는 `t.Parallel()` 이면서 상대 경로(`./update.go`)를 판다. Go 의 스케줄링상 최상위 parallel 테스트는 직렬 패스가 끝난 뒤 재개되므로 현재 실제 겹침은 없고, `-count=10` 반복에서도 재현되지 않았다. **Required fix**: 없음. 다만 이 무언의 불변식이 깨지면(누가 chdir 하는 테스트에 `t.Parallel()` 을 붙이면) AST 가드가 `parse ./update.go: no such file` 로 붉어진다 — 그때 원인을 AST 가드에서 찾지 않도록 기록한다. F-01 의 후보 원인 하나이기도 하다(미확인).

## 구현 보고·SPEC·선행 감사와 어긋나는 실측

1. **`progress.md` 위증 검사 개수 — 본문 "11건" vs 표 13행 vs YAML `13`.** 표를 세면 F1·F2·F3·F4·F5 + FA·FB + FG·FH·FI·FJ·FK·FL = **13**. YAML `falsification_checks_run: 13` 과 일치하고, 산문 "11건 전부 예상대로 붉어졌다" 만 틀렸다. 관측 자체가 아니라 **라벨이 틀린 형태**다. → `progress.md` 산문의 `11` → `13`.
2. **`total_run_phase_files: 8` vs 실측 9** (F-03).
3. **`tests: "... 3/3 ok"` 는 그 실행에 한해 참이나 재현되지 않는다** — 내 4회 실행 중 `internal/cli` 가 1회 FAIL. 보고를 위증으로 보지는 않는다(그들이 관측한 실행은 실제로 초록이었을 것). 다만 "3/3 ok" 를 **패키지가 안정하다**로 읽으면 과대 인용이다 (F-01).
4. **`known_baseline_failure` 는 정확하다 — 독립 재현했다.** `go test -count=1 ./internal/cli/agentlint/` → `--- FAIL: TestConstitutionCrossReference (0.00s)` / `moai-constitution.md should cross-reference agent-authoring.md for effort matrix`. 이 카드가 `.claude/` 를 건드리지 않았음도 diff 로 확인(`git diff --stat` 12파일에 `.claude/` 없음) — **선행 결함 귀속 정확**. 다만 `acceptance.md` §D.5 의 완료 정의는 `go test ./internal/cli/...`(하위 패키지 포함) 통과를 요구하므로, **DoD 한 항목이 문자 그대로는 미충족**이다. 보고가 이를 은폐하지 않고 명시했으므로 결함으로 세지 않고 기록만 한다.
5. **선행 감사와의 불일치 없음.** spec §A.6 의 `:205`, §A.7 의 `failed` 생산 3지점, `InitResult.Warnings` 통로, 프로덕션 `CleanReinstallOptions` 리터럴 2건 — 전부 재측정해 일치했다. 이 계보에서 다섯 번 반복된 「검사할 수 없는 것을 검사한다고 주장하는 판정」의 **여섯 번째 사례는 발견되지 않았다**.

## Gaps (관측하지 않은 것)

- **위증 검사 13건을 실제로 주입해 재현하지 않았다.** 읽기 전용 제약(구현·테스트 수정 금지)상 산술과 코드 구조로 연역했다. F1·F2·FB·FK·FJ 는 산술이 결정적이라 확신이 높고, FG·FH·FL 은 코드 경로가 단선이라 확신이 높다. **재현 관측은 없다.**
- **`internal/cli` 78.5% · `internal/core/project` 88.4% 커버리지를 재측정하지 않았다.** `internal/mirrornotice` 91.7% 만 독립 확인했다.
- **F-01 의 실패 테스트 이름을 포획하지 못했다.** 첫 실행 출력을 `tail -40` 으로 잘랐고, 이후 3회가 초록이라 재현 실패.
- **실제 폴백 플랫폼(권한 없는 Windows)에서 통지를 육안 확인하지 않았다.** `acceptance.md` §D.4 가 이미 AC 밖으로 분리한 항목이다. `GOOS=windows go vet` 은 **컴파일만 증명**한다.
- **선행 `AC-CSC-010` 을 재실행하지 않았다** — `internal/template` diff 0 으로 갈음했다.
- **CI 매트릭스 판정 없음.** 로컬 darwin 단일 환경 측정이며, 미푸시·PR 없음.

## Residual risk

- **`internal/cli` 의 간헐 실패가 미해소다.** 이 카드가 원인이라는 근거도, 아니라는 근거도 없다 — 신규 테스트의 격리 반복이 안정하다는 것은 정황이지 증명이 아니다. PR CI 가 붉어지면 F-01·F-07 을 먼저 볼 것.
- **AST 가드 2건은 소스 텍스트를 본다.** §D.6 이 기록한 한계 그대로이며, 이 감사가 넓히지도 좁히지도 못했다.
- **2회차 이후 침묵은 설계상 남는다**(spec §A.5 / §D, 승계 카드 t173). `CHANGELOG` 문구가 이 경계를 정확히 적었음을 확인했다 — "The run in which the fallback happens now tells you so; later runs still do not."
- **`ErrOut` 을 주입하지 않는 기존 테스트 25건**이 `os.Stderr` 기본값으로 떨어진다. darwin 에서는 심볼릭 링크가 성공해 통지가 나지 않으므로 무해하나, 폴백이 발생하는 환경에서 CI 를 돌리면 그 테스트들이 실제 stderr 로 통지를 흘린다. `acceptance.md` §D.6 이 이미 이 상태를 의도된 것으로 기록했다.
