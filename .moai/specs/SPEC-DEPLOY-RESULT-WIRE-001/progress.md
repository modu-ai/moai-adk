# progress — SPEC-DEPLOY-RESULT-WIRE-001

## §E.1 Plan-phase Audit-Ready Signal

- Tier: **M** (spec.md + plan.md + acceptance.md). 근거: 예상 변경 파일 6-8개 · 예상 LOC 150-350 — Tier S 상한(5 files / 300 LOC)의 경계에 걸치고, 세 호출부가 두 패키지(`internal/cli`, `internal/core/project`)에 흩어져 판정 팔이 3개로 갈라진다.
- 요구사항 **10개** / 판정 기준 **9개** — Tier M 상한 16/16 이내 (iter-1 은 9/8).
- iter-3 (v0.3.0): plan-audit iteration 2 **FAIL(0.75, 반복 상한 2/2)** 후 리드가 **정정 후 PASS-with-debt** 로 수용한 5건 대응. 요구사항 10 / 판정 9 **불변**. **N2**(구현자 함정) — `AC-DRW-004` 2번 팔이 올바른 구현에서 `1+2=3 ≠ 1+3=4` 로 붉어졌다. 낮은 표본을 상한 위(**4**)로 올리고 상한 판정을 **3번 팔**로 분리. **N4** — `AC-DRW-009` 에 3번 단언 추가(두 `CleanReinstallOptions` 리터럴의 `ErrOut: cmd.ErrOrStderr()` 주입을 `go/parser` 로 확인) → `AC-DRW-003` 의 재현 불가능하던 위증 검사가 실재하게 됨. **주장 하향은 채택하지 않았다.** **N5** — 부채로 미루지 않고 `AC-DRW-009` 1번에 검증 수단(AST) 명시 + 2번을 런타임 단언으로 분리; AST 가드가 런타임 값을 증명하지 않는다는 한계만 `acceptance.md` §D.6 에 잔존으로 기록. **N1** — spec §D 의 `:196` → `:205`. **N6** — plan §E·§H 의 `..008` → `..009`. **N3** — §A.4 호출 행 `:625` → `:624`.
- iter-3 실측 기록: (a) `sed -n '624,628p' internal/cli/update.go` → 624행이 `if _, runErr := runCleanReinstall(planCtx, cwd, CleanReinstallOptions{`, `Out:` 은 627 — 감사의 `:624` 가 맞다. (b) `grep -n "196\|:205" spec.md` → §D(139행)가 여전히 `:196` 이었음을 확인, §A.6(78행)만 고쳐져 문서가 자기모순이었다. (c) 소스 스캔 가드 선례 확인 — `internal/template/agent_askuser_audit_test.go` 가 `filepath.WalkDir` + `os.ReadFile` 로 정적 검사를 수행한다(AC-DRW-009 의 AST 가드가 이 저장소에서 이질적이지 않다는 근거).
- iter-3 자기 지적: v0.2.0 은 D1 에서 「판정은 프로덕션 배선에서, 각 판정은 어떤 잘못된 구현에서 붉어지는지를 적는다」는 규칙을 뽑아 **형식으로는** 9/9 에 적용했으나(위증 검사 절), 그 위증 검사가 **자기 AC 안에서 재현 가능한지**는 보지 않았다 — 규칙을 그 규칙의 출처 한 층 위에 적용하지 못했다. 그것이 N4 이며, plan AP-11 로 규율화했다.
- iter-2 (v0.2.0): plan-audit iteration 1 **FAIL(0.75)** 의 결함 7건 대응. **D1(critical)** — §A.4 스트림 실측이 틀렸다: clean-reinstall 의 `nil ⇒ os.Stderr` 는 프로덕션에서 실행되지 않는 분기이고, 호출부 두 곳이 stdout 을 주입한다. 표·결론 정정 + M3 을 `CleanReinstallOptions.ErrOut` 주입으로 재배선 + plan R4("기본 경로에서 판정") 은퇴 + AC-DRW-003 을 **주입된 프로덕션 배선**에 재고정. **D2** — `REQ-DRW-007` 비비례성을 통지 전체로 확장, `failed` 를 개수 요약 + 예시 3건 상한으로 고정, AC-DRW-004 에 `failed` 팔 추가. **D3** — AC-DRW-003 을 stdout 위험 두 호출부에 명시 고정. **D4** — 오귀속 문구 행 `196` → `205`(재측정). **D5** — `REQ-DRW-010` / `AC-DRW-009` 신설(주입 seam 기본값 가드). **D6** — §A.7 신설(`failed` 문구 3-of-3 실측)로 1-of-3 일반화 대체. **D7** — 전 AC 에 위증 검사 절 추가.
- iter-2 실측 기록: (a) `grep -n "Out:" internal/cli/update.go` → `425`, `627`; 두 자리의 `out` 은 각각 `runUpdate`(`:138`)의 `:154 out := cmd.OutOrStdout()` 와 `emitDryRunReinstallPlan`(`:592`)의 `out` 인자이며 후자는 `:362` 에서 같은 `out` 을 받는다 → **둘 다 stdout**. (b) `grep -n "non-symlink entry already exists" internal/template/skill_mirror.go` → **205**. (c) `MirrorModeFailed` 생산 지점 3곳 → `:217`, `:226`, `:243`.
- iter-2 자기 지적: iter-1 은 **기본값 선언을 읽고 스트림을 판정**했다. 기본값의 존재는 그 기본값이 쓰인다는 근거가 아니다. D1·D5·D7 이 같은 형태이며, 「판정은 프로덕션 배선에서, 각 AC 는 위증 검사를 본문에 적는다」로 일괄 대응했다.
- iter-2 감사 쟁점: 반박한 지적 없음 — D1·D4 는 독립 재측정으로 재현했고 나머지 5건도 전부 수용했다. init 범위 결정은 감사가 §A.3 인용 4건 + 추가 연결 2건(`PhaseExecutor` 통과 · `p.Collect` → stderr)을 확인해 유지됐다.
- SPEC ID 정규식 검사: `SPEC-DEPLOY-RESULT-WIRE-001` → **`PASS`** (Bash `[[ "$ID" =~ ^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$ ]]` 실행, verbatim 출력 `PASS`).
- 중복 검사: `ls -d .moai/specs/SPEC-DEPLOY*` → `no matches found`; `grep -rl "SPEC-DEPLOY-RESULT-WIRE-001" .moai/specs/` → 무출력. 신규 ID 확정.
- 착수 시점 실측(작성 시점, 워크트리 `t176` / 브랜치 `WT-deploy-result-wire`):
  - `deployer.Deploy(` 프로덕션 호출부 **3곳** — `internal/cli/update_template_sync.go:323`, `internal/cli/update_clean_install.go:439`, `internal/core/project/initializer.go:356`.
  - `ResultDeployer` 프로덕션 소비자 **0건**(sync-audit F2 재확인).
  - `InitResult.Warnings []string` 존재 — `internal/core/project/initializer.go:97`, 기존 append 6곳, 표시부 `internal/cli/init.go:706`.
  - 출력 스트림 독트린 — `internal/cli/CLAUDE.md:14` "stderr = human progress messages, warnings, errors. Never mix."
- 카드 전제 정정 2건: (1) 범위가 "internal/cli (update 경로)" 로 적혀 세 번째 호출부를 덮지 않았다 — 세 곳 전부를 범위에 넣었다. (2) init 배선을 "materially larger change" 로 적었으나 **되돌릴 통로(`InitResult.Warnings`)가 이미 있고 `deployTemplates` 가 `result` 를 이미 인자로 받는다** — 비용이 update 호출부와 같은 급이다.
- 설계 결정 기록: `skipped` 모드 경고(sync-audit F4 오귀속)는 사용자에게 **올리지 않는다**(REQ-DRW-009 / AC-DRW-008). 문구 수정은 승계 카드 `t173` 소관이며, 이 SPEC 은 문구를 고치지 않는다.
- 잔존 고지: 폴백 플랫폼 2회차 이후 실행은 모드가 `skipped` 라 통지가 나가지 않는다(sync-audit F1). spec §D 본문에 명시.

## §E.2 Run-phase Evidence

착수 시점 §C 사전 점검 5건 전부 재측정 — 값이 스테일이 아니었다: 호출부 3곳 동일(`update_template_sync.go:323` · `update_clean_install.go:439` · `initializer.go:356`), `ResultDeployer` 프로덕션 소비자 0건, `InitResult.Warnings` 존재, `grep -n "Out:" internal/cli/update.go` → `425` · `627` (**둘 다 stdout**, §A.4 유효).

### AC PASS/FAIL 매트릭스

| AC | Status | 검증 명령 | 관측 출력 |
|---|---|---|---|
| AC-DRW-001 | PASS | `go test ./internal/mirrornotice/ -run TestLines_CopyFallbackCarriesCount` · `go test ./internal/cli/ -run 'MirrorNoticeGoesToStderr'` | `--- PASS: TestLines_CopyFallbackCarriesCount` / `--- PASS: TestCleanReinstall_MirrorNoticeGoesToStderrNotStdout (0.02s)` · `--- PASS: TestTemplateSync_MirrorNoticeGoesToStderrNotStdout (0.03s)` — 통지 존재 + 개수(11 / 7) 단언 포함 |
| AC-DRW-002 | PASS | `go test ./internal/mirrornotice/ -run TestLines_NoFallbackEmitsNothing` · `go test ./internal/cli/ -run 'NoFallbackEmitsNothing'` | `--- PASS: TestLines_NoFallbackEmitsNothing` / `--- PASS: TestCleanReinstall_NoFallbackEmitsNothing (0.02s)` · `--- PASS: TestTemplateSync_NoFallbackEmitsNothing (0.03s)` — stdout·stderr 양쪽 무출력 |
| AC-DRW-003 | PASS | `go test ./internal/cli/ -run 'MirrorNoticeGoesToStderr'` | `--- PASS` ×2. **두 호출부 모두 프로덕션 배선**: clean-reinstall 은 `Out` 에 stdout 버퍼를 **주입**한 상태(`update.go:425`·`:627` 과 동일), template sync 는 `out = cmd.OutOrStdout()` 그대로. 각 팔이 stderr 양성 + stdout 음성 2단언 |
| AC-DRW-004 | PASS | `go test ./internal/mirrornotice/ -run 'NotProportional\|CappedAtThree'` | `--- PASS: TestLines_CopyNoticeIsNotProportional` (2 vs 34 줄 수 동일) · `--- PASS: TestLines_FailedNoticeIsNotProportional` (**4** vs 34 동일 — 낮은 표본이 상한 위) · `--- PASS: TestLines_FailedExamplesAreCappedAtThree` (N=2 → 2건, N=34 → 3건). 3팔 독립 |
| AC-DRW-005 | PASS | `go test ./internal/cli/ -run 'PlainDeployerStillDeploys'` · `go test ./internal/core/project/ -run TestInit_PlainDeployerStillDeploys` | `--- PASS` ×3. 이중체(`plainDeployer` / `capturingDeployer`)는 `DeployWithResult` 를 **정의하지 않는다**; 각 테스트가 `any(dep).(template.ResultDeployer)` 로 확장 부재를 먼저 확인한 뒤 오류·panic·통지 부재를 단언 |
| AC-DRW-006 | PASS | `go test ./internal/cli/ -run TestCleanReinstall_FailedMirrorsReportedAndFailOpen` | `--- PASS (0.02s)` — 개수 9 도달 + 경고 인용 1~3건 + `runCleanReinstall` 무오류(오류 시 `t.Fatalf`) |
| AC-DRW-007 | PASS | 위 세 팔 각각 | init 팔 `--- PASS: TestInit_MirrorCopyFallbackReachesWarnings` (도달 지점 `InitResult.Warnings`), update 두 팔 `--- PASS` (도달 지점 캡처된 stderr) |
| AC-DRW-008 | PASS | `go test ./internal/mirrornotice/ -run SkippedWarningsAreNotForwarded` · `-run TestCleanReinstall_SkippedWarningsNotSurfaced` · `-run TestInit_SkippedWarningsAreNotForwarded` | `--- PASS` ×3 — `a non-symlink entry already exists` 가 어느 스트림·어느 슬라이스에도 없음 |
| AC-DRW-009 | PASS | `go test ./internal/cli/ -v -run 'TestSeamDefault\|TestCleanReinstallOptionsLiterals'` | `--- PASS: TestSeamDefaultIsTheProductionDeployer` (AST: seam 기본값이 `NewDeployerWithRendererAndForceUpdate(embedded, renderer, true)`) · `--- PASS: TestSeamDefaultSatisfiesResultDeployer` (런타임 `_, ok := dep.(template.ResultDeployer)`) · `--- PASS: TestCleanReinstallOptionsLiteralsInjectErrOut` (AST: `update.go` 의 `CleanReinstallOptions` 복합 리터럴을 **전부 수집**해 각각 `ErrOut: cmd.ErrOrStderr()` 확인 — 착수 시점 2건, 개수로 고정하지 않음) |

### 위증 검사 실행 결과 (각 AC 가 명시한 잘못된 구현을 실제로 만들어 붉어지는지 관측)

이 계보가 다섯 번 만들어 낸 결함이 **검사할 수 없는 것을 검사한다고 주장하는 판정**이므로, 초록만으로 끝내지 않고 각 판정을 실제로 깨뜨려 확인했다. **13건** 전부 예상대로 붉어졌다(아래 표 행 수 = 13 = `falsification_checks_run`). 초판은 이 숫자를 `11` 로 적었다 — 관측은 옳고 **라벨만 틀린** 형태이며, 이 계보의 반복 결함(주장이 실제 검사한 것과 어긋남)의 작은 사례다.

| # | 주입한 잘못된 구현 | 붉어진 판정 (관측) |
|---|---|---|
| F1 | `failed` 예시 상한 제거(항목별 전량 출력) | `TestLines_FailedNoticeIsNotProportional` (`4 skills → 5 lines, 34 skills → 35 lines`) · `TestLines_FailedExamplesAreCappedAtThree` (`34 → 34건`) · `TestLines_FailedCountAndWarningsReachTheUser` |
| F2 | 상한을 3 → 1 | `TestLines_FailedExamplesAreCappedAtThree` (`2 failed → 1건, want 2`) — 3번 팔이 상한을 단독으로 판정함을 확인 |
| F3 | `skipped` 를 `failed` 와 함께 forward (AP-1 형태) | `TestLines_SkippedWarningsAreNotForwarded` · `TestLines_MixedRunReportsCopyAndFailedOnly` |
| F4 | 결과 내용을 보지 않고 무조건 통지 | `TestLines_NoFallbackEmitsNothing` · `TestLines_SkippedWarningsAreNotForwarded` — AC-DRW-001 은 이 구현에서 **초록**이므로 AC-DRW-002 가 유일한 검출자임이 실측 확인됨 |
| F5 | 통지에서 개수 제거 | `TestLines_CopyFallbackCarriesCount` · `TestLines_MixedRunReportsCopyAndFailedOnly` (AP-6 방어 확인) |
| FA | init 소비부 제거 | `TestInit_MirrorCopyFallbackReachesWarnings` |
| FB | init 에서 타입 단언 없이 무조건 승격 | `TestInit_PlainDeployerStillDeploys` — `panic: invalid memory address or nil pointer dereference` |
| FG | template sync 통지를 `out`(stdout)에 실음 (iter-1 M3 이 지시했던 형태) | `TestTemplateSync_MirrorNoticeGoesToStderrNotStdout` — stderr 부재 + **stdout 존재** 두 단언 모두 붉어짐 |
| FH | `ErrOut` 필드는 추가했으나 호출부 한 곳에서 주입 누락 | `TestCleanReinstallOptionsLiteralsInjectErrOut` (`Literal fields: [DryRun Force Out RunMigrateAgency]`) — **N4 가 실재하게 만든 위증 검사가 실제로 작동함을 확인** |
| FI | seam 기본값을 비프로덕션 생성자로 교체 | `TestSeamDefaultIsTheProductionDeployer` |
| FJ | seam 기본값이 프로덕션 생성자를 **부르지만** `ResultDeployer` 를 만족하지 않는 래퍼 반환 | `TestSeamDefaultSatisfiesResultDeployer` **붉음** / `TestSeamDefaultIsTheProductionDeployer` **초록** — 1번·2번 단언을 분리한 N5 결정이 실측으로 정당화됨 |
| FK | `deployWithMirrorNotice` 에서 무조건 승격 | `TestCleanReinstall_PlainDeployerStillDeploys` — `panic: interface conversion: *cli.plainDeployer is not template.ResultDeployer` |
| FL | 소비부 제거(항상 `Deploy` 만 호출) | `TestCleanReinstall_MirrorNoticeGoesToStderrNotStdout` · `TestTemplateSync_MirrorNoticeGoesToStderrNotStdout` 동시 붉음 |

### RED 증거 (구현 전 실패 출력)

- M1: `go test ./internal/mirrornotice/...` → `no non-test Go files in .../internal/mirrornotice` / `FAIL ... [build failed]`
- M2: `go test ./internal/core/project/... -run TestInit_Mirror...` → `--- FAIL: TestInit_MirrorCopyFallbackReachesWarnings (0.02s)` / `mirror notice absent from InitResult.Warnings — the init path is still silent about a copy fallback` / `warnings: []`
- M3: `go test ./internal/cli/ -run ...` → `update_mirror_notice_test.go:237:21: undefined: newTemplateSyncDeployer` / `FAIL github.com/modu-ai/moai-adk/internal/cli [build failed]`

### 회귀 (M4)

- 선행 `AC-CSC-010`: `go test ./internal/template/ -run TestSkillMirror -count=1 -v` → `--- PASS: TestSkillMirror_ClaudePathUnchangedByMirror (0.01s)` 포함 10/10 PASS.
- `internal/template` 무변경: `git status --porcelain internal/template/` → 무출력.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-08-22
run_commit_sha: ba08d8694
run_status: audit-ready
ac_pass_count: 9
ac_fail_count: 0
falsification_checks_run: 13
falsification_checks_reddened_as_predicted: 13
preserve_list_post_run_count: 0
new_warnings_or_lints_introduced: 0
cross_platform_build:
  darwin_vet: "go vet ./internal/cli/... ./internal/core/project/... ./internal/mirrornotice/... → exit 0"
  windows_vet: "GOOS=windows GOARCH=amd64 go vet (동일 범위) → exit 0 — 컴파일만 증명하며 Windows 동작 근거가 아니다"
coverage:
  internal/mirrornotice: "91.7% of statements"
  internal/core/project: "88.4% of statements"
  internal/cli: "78.5% of statements — 패키지 기존 baseline(이 카드가 낮춘 값이 아니다). 이 카드가 더한 코드의 판정은 위 매트릭스"
lint: "golangci-lint run --timeout=5m (동일 범위) → 0 issues"
tests: "go test -count=1 ./internal/cli/ ./internal/core/project/ ./internal/mirrornotice/ → 3/3 ok"
known_baseline_failure: "internal/cli/agentlint TestConstitutionCrossReference — moai-constitution.md 가 agent-authoring.md 를 상호참조하지 않는다는 문서 lint. 이 카드는 .claude/ 를 건드리지 않았다(git status --porcelain .claude/ 무출력, grep -c 'agent-authoring' → 0). 선행 결함"
total_run_phase_files: 9
m1_to_mN_commit_strategy: "마일스톤당 1커밋 (M1 mirrornotice · M2 init · M3 update 두 경로 · M5 CHANGELOG+progress). M4 는 검증 전용이라 diff 없음"
```

### 기록된 잔존 2건 (sync-audit 제기, 이 카드에서 고치지 않음)

**1. DoD 는 문자 그대로 충족되지 않았다 — 선행 baseline 예외.** `acceptance.md` §D.5 는 `go test ./internal/cli/...`(서브패키지 포함)를 요구하지만, `internal/cli/agentlint` 의 `TestConstitutionCrossReference` 가 붉다. 위 `known_baseline_failure` 가 그 내용이며, 감사가 독립적으로 재현해 **이 카드 소관이 아님**을 확인했다(이 카드는 `.claude/` 파일을 하나도 건드리지 않는다). 실행 범위를 `./internal/cli/`(서브패키지 제외)로 좁혀 초록을 얻었으므로, **"go test 전부 초록" 이 아니라 "선행 결함 1건을 제외하고 초록"** 이 정확한 서술이다. 나중에 읽는 사람이 이것을 무결점 실행으로 오독하지 않도록 여기 명시한다.

**2. AST 가드의 파일 범위 — 승계 작업.** `AC-DRW-009` 3번 단언은 `internal/cli/update.go` **한 파일만** 파싱한다. 리터럴 **개수**를 고정하지 않은 것은 의도대로이고 그 축의 우려는 해당되지 않지만, `CleanReinstallOptions` 복합 리터럴이 `internal/cli` 의 **다른 파일**에 새로 생기면 판정 밖으로 빠진다. 실측(`grep -rn "CleanReinstallOptions{" internal/`, 전수)으로 프로덕션 리터럴은 `update.go` 의 둘뿐이고 나머지는 전부 `_test.go` 이므로 **오늘 놓치는 것은 없다**. 승계 작업: 파싱 대상을 `internal/cli` 의 비-테스트 소스 전체로 넓히고, AC 문구를 "`internal/cli/update.go` 안의" → "`internal/cli` 프로덕션 소스의" 로 고친다. `acceptance.md` §D.6 의 기존 잔존(AST 가드는 리터럴을 증명하지 런타임 값을 증명하지 않는다)과 나란히 읽는다.

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: 2026-08-22
sync_commit_sha: pending-backfill-sync-SPEC-DEPLOY-RESULT-WIRE-001
sync_status: audit-ready
b12_self_test_a: "grep -c 'SPEC-DEPLOY-RESULT-WIRE-001' CHANGELOG.md → 0 (중복 없음, 방출 진행)"
b12_self_test_b: "grep -oE 'AC-([A-Z0-9]+-)*[0-9]+' acceptance.md | sort -u | wc -l → 10 (AC-DRW-001..009 9건 + 선행 회귀 참조 AC-CSC-010 1건). CHANGELOG 는 이 SPEC 소관인 9건을 적는다"
b12_self_test_c: "ls internal/mirrornotice/{notice.go,notice_test.go} · internal/cli/{mirror_notice.go,update.go,update_mirror_notice_test.go} → 전부 존재. grep -n 'ErrOut' internal/cli/update.go → :428 · :638 두 프로덕션 리터럴"
changelog_entry_position: "[Unreleased] → ### Added, SPEC-CODEX-DUAL-AGENTS-001 항목 바로 뒤 (SPEC 링크를 단 close 항목끼리 모은다)"
frontmatter_status_transitions:
  spec_md: "in-progress → completed (status: 만 변경, updated: 는 이미 2026-08-22)"
  plan_md: "해당 없음 — frontmatter 없음"
  acceptance_md: "해당 없음 — frontmatter 없음"
  progress_md: "해당 없음 — frontmatter 없음"
docs_surface: "README(4로케일) · docs-site 무변경. 근거: 두 표면 어디에도 `.agents/skills` 미러 기능 서술이 없다(grep 실측 — 유일한 매치 2건은 'agents/skills' 문자열이 다른 문맥에 쓰인 오탐). 이 카드는 문서화되지 않은 표면에 stderr 출력을 더할 뿐이고 새 커맨드·플래그가 없다"
changelog_reconciliation: "선행 카드가 적은 제약 문장 'The fallback warning does not currently reach you' 는 run-phase M5(ba08d8694)가 같은 자리에서 **치환**했다 — 제약과 그 해제가 나란히 놓이지 않는다. 나란히 남은 첫 번째 제약(2회차 이후 복사본 고착)은 여전히 참이며, 치환된 문장이 '2회차 이후는 mode=skipped 라 통지도 나오지 않는다'를 명시해 같은 후속 작업(t173)으로 두 항목을 묶는다"
mx_tag_validation: "신규 코드에 @MX 주석 추가 없음. `internal/mirrornotice` 는 순수 함수 2개(goroutine·전역 상태·fan_in ≥3 없음)이고, 세 호출부는 기존 함수 안 몇 줄 추가다 — @MX:WARN/@MX:ANCHOR 트리거 어느 것도 성립하지 않는다"
sync_audit: "PASS 91/100, blocking 0건 (Functionality 95 / Security 92 / Craft 82 / Consistency 93). 보고서 `.moai/reports/t176/sync-audit.md`"
audit_findings_disposition: "F-03(total_run_phase_files 8→9) · F-05(run_commit_sha 백필) · 산문 '11건'→'13건' 3건은 이 커밋에 반영. F-01(internal/cli 간헐 실패, 귀속 미확정) · F-02(AST 가드 파일 범위) · F-04 · F-06 · F-07 은 optional 이며 승계 대상 — §E.3 의 '기록된 잔존 2건' 참조"
```
