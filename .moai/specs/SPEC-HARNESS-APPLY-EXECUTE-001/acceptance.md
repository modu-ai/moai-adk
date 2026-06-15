# Acceptance Criteria — SPEC-HARNESS-APPLY-EXECUTE-001

GEARS-format ACs. grep idiom 규약:
- **comment-line 필터 (canonical)**: `grep -n` 출력은 각 줄에 `N:` prefix를 붙이므로, comment 제외 필터는 **반드시 `-n`-aware 형태** `^[^:]*:[0-9]*:[[:space:]]*//`를 사용한다. column-0 형태 `^[[:space:]]*//` (및 non-greedy `^\s*//`)는 `-n` prefix 때문에 **inert** (comment 줄이 `-v`를 통과해 살아남음 — SEC-HARDEN line 반복 BLOCKING 결함 + 본 SPEC plan-audit iter1 D2 실증). 본 파일 전체 comment-filter grep은 AC-AEX-016과 동일한 `-n`-aware 형태로 통일한다.
- **test-name 검증**: trailing-anchored `$` (예: `TestX$`).
- **토큰 검증**: exact-token matcher (`grep -nE` + 정확 토큰).

## §D. AC Matrix

| AC ID | REQ | Given-When-Then (요약) | 검증 idiom (severity) |
|-------|-----|------------------------|----------------------|
| AC-AEX-001 | REQ-AEX-001 | execute.go가 `internal/cli/harness/`에 존재하고 boundary guard가 스캔함 | MUST |
| AC-AEX-002 | REQ-AEX-002 | `NewExecuteCmd()`가 `newHarnessRouterCmd()`에 등록됨 | MUST |
| AC-AEX-003 | REQ-AEX-003 | `--execute` 부재 시 payload-only 보존, 존재 시 Go 경로 | MUST |
| AC-AEX-004 | REQ-AEX-004 | proposal ID → `.moai/harness/proposals/<id>.json` 로드 | MUST |
| AC-AEX-005 | REQ-AEX-005 | Pipeline이 `AutoApply: true`로 구성됨 | MUST |
| AC-AEX-006 | REQ-AEX-006 | harness.yaml 디스크 `auto_apply: false` 불변 | MUST (FROZEN) |
| AC-AEX-007 | REQ-AEX-007 | AutoApply=true 하 `*ApplyPendingError` 발생 안 함 (contract) | MUST |
| AC-AEX-008 | REQ-AEX-008 | Applier가 regression gate + outcome observer 배선됨 | MUST |
| AC-AEX-009 | REQ-AEX-009 | snapshotBase/manifestPath/baselinePath/usageLogPath가 project root 상대 canonical 경로로 resolve됨 | MUST |
| AC-AEX-010 | REQ-AEX-010 | 정상 apply 후 `usage-log.jsonl`에 `apply_outcome` 1건 | MUST |
| AC-AEX-011 | REQ-AEX-011 | nil/empty sessions 전달 시 L2 미reject | MUST |
| AC-AEX-012 | REQ-AEX-012 | proposal 파일 부재 OR L1~L4 rejection → exit 1 | MUST |
| AC-AEX-013 | REQ-AEX-013 | `*ApplyRegressionError` → exit 1 | SHOULD |
| AC-AEX-014 | REQ-AEX-014 | `*ApplyPendingError` (AutoApply=true) → exit 2 (invariant) | MUST |
| AC-AEX-015 | REQ-AEX-015 | measurement-exec error → exit 2 | SHOULD |
| AC-AEX-016 | REQ-AEX-016 | `internal/cli/harness/`에 `AskUserQuestion(` 부재 | MUST (C-HRA-008) |

## §D.1 상세 시나리오 (Given-When-Then)

### AC-AEX-001 — execute.go 위치 + boundary guard 커버
- **Given** `internal/cli/harness/` 디렉터리
- **When** `ls internal/cli/harness/execute.go` 실행
- **Then** 파일이 존재하고 package 선언이 `package harness`이며 `TestPropose_NoAskUserQuestion`이 스캔 대상에 포함
- **검증**:
```bash
test -f internal/cli/harness/execute.go && echo PRESENT
go test -run 'TestPropose_NoAskUserQuestion$' ./internal/cli/harness/...
```

### AC-AEX-003 — payload-only 보존 (회귀) vs Go 경로 분기
- **Given** pending proposal 1건이 있는 프로젝트
- **When** `moai harness apply` (no `--execute`) 실행
- **Then** 기존 payload-only 동작 (JSON stdout 출력, `Applier.Apply()` 미호출) 불변
- **When** `moai harness apply --execute --id <id>` 실행
- **Then** Go execute 경로(`Applier.Apply()`)가 호출됨
- **검증** (test 함수 trailing-anchored `$`):
```bash
go test -run 'TestApply_PayloadOnly_PreservedWithoutExecuteFlag$' ./internal/cli/harness/...
go test -run 'TestApply_DelegatesToGoPath_WithExecuteFlag$' ./internal/cli/harness/...
```

### AC-AEX-005 — Pipeline AutoApply=true 구성
- **Given** execute 경로
- **When** Pipeline을 구성
- **Then** `safety.PipelineConfig{AutoApply: true}`로 생성됨 (L1~L4 강제, L5 auto-approve)
- **검증** (exact-token, `-n`-aware comment filter `^[^:]*:[0-9]*:[[:space:]]*//` — `grep -n`이 `N:` prefix를 붙이므로 column-0 `^[[:space:]]*//`는 inert):
```bash
grep -nE 'AutoApply:[[:space:]]*true' internal/cli/harness/execute.go | grep -v '^[^:]*:[0-9]*:[[:space:]]*//'
```

### AC-AEX-006 — harness.yaml 디스크 불변 (FROZEN)
- **Given** execute 경로 실행
- **When** apply 완료
- **Then** `harness.yaml`의 `auto_apply` 값은 디스크에서 `false` 그대로
- **검증** (execute 경로가 harness.yaml에 write하지 않음 — config write API 미호출; `-n`-aware comment filter):
```bash
# execute.go가 harness.yaml/config write를 수행하지 않음
grep -nE '(harness\.yaml|WriteHarnessConfig|SaveHarnessConfig|auto_apply)' internal/cli/harness/execute.go | grep -v '^[^:]*:[0-9]*:[[:space:]]*//' || echo "NO_CONFIG_WRITE"
```
- **통합 검증**: execute 실행 전후 `harness.yaml` byte-identical (테스트에서 read-compare).

### AC-AEX-007 — autoApply contract: production wiring (non-vacuous)
[D3 분리] 이 AC는 **프로덕션 배선**이 Pipeline을 `AutoApply: true`로 구성함을 증명한다. "AutoApply=true 하 Pending 발생 안 함"을 직접 단언하지 **않는다** — 그것은 `pipeline.go:147-157`에서 항상 참인 tautology (stub만 증명하고 통합을 증명하지 못함). 대신 RunExecute가 실제로 production Pipeline 생성 시 `AutoApply: true`를 전달하는지를 검증한다 (AC-AEX-005 grep과 상호보강: grep은 소스 토큰, 이 AC는 런타임 구성값).
- **Given** RunExecute가 (실제 또는 가시화된) Pipeline 생성 지점
- **When** Pipeline을 구성
- **Then** 전달된 `PipelineConfig.AutoApply == true` (테스트가 생성 지점을 관측 가능하도록, RunExecute는 Pipeline 생성을 `PipelineConfig` 값으로 노출하거나 seam을 통해 검증 가능해야 함 — design.md §F 테스트 seam 결정 참조)
- **검증**:
```bash
go test -run 'TestExecute_ProductionPipeline_UsesAutoApplyTrue$' ./internal/cli/harness/...
```

### AC-AEX-014 — autoApply contract invariant: error-classification branch (non-vacuous)
[D3 핵심] 이 AC는 **error-classification 분기**를 증명한다 (AutoApply=true가 Pending을 안 낸다는 tautology가 아님). 테스트는 `Applier.Apply()`의 2번째 파라미터(`SafetyEvaluator`, injectable)에 **`DecisionPendingApproval`을 반환하는 stub evaluator를 주입**하여 `*ApplyPendingError`를 강제 발생시키고, RunExecute의 error→exit-code 매핑이 이를 **INVARIANT VIOLATION (exit 2)**으로 분류하는지 검증한다.
- **Given** `*ApplyPendingError`를 반환하도록 stub `SafetyEvaluator`(`DecisionPendingApproval` 반환)를 주입한 RunExecute 경로
- **When** RunExecute가 `Applier.Apply()`로부터 `*ApplyPendingError`를 수신
- **Then** verb는 exit 2로 종료 + "INVARIANT VIOLATION: autoApply contract — Pending under AutoApply=true" diagnostic을 stderr에 출력 (단순 user-error exit 1이 아님)
- **And** 분류는 `errors.As(err, &*ApplyPendingError{})`로 수행 (errjoin walk 가능)
- **검증** (error-classification 분기를 실증하는 stub 주입 테스트):
```bash
go test -run 'TestExecute_PendingErrorUnderAutoApply_ClassifiedAsInvariantExit2$' ./internal/cli/harness/...
```
- 참고: `Applier.Apply`의 evaluator 파라미터는 `SafetyEvaluator` 인터페이스(applier.go:194~196)이므로 stub 주입에 추가 seam이 불필요하다 — RunExecute가 evaluator를 주입받거나 내부 생성 지점을 테스트 가능하게 노출하면 충분 (design.md §F 결정).

### AC-AEX-008 — Applier 배선 (regression gate + outcome observer)
- **Given** execute 경로
- **When** Applier 생성
- **Then** `NewApplierWithRegressionGate(...).WithOutcomeObserver(NewObserver(...))` 체인 사용
- **검증** (exact-token, `-n`-aware comment filter):
```bash
grep -nE 'NewApplierWithRegressionGate' internal/cli/harness/execute.go | grep -v '^[^:]*:[0-9]*:[[:space:]]*//'
grep -nE 'WithOutcomeObserver' internal/cli/harness/execute.go | grep -v '^[^:]*:[0-9]*:[[:space:]]*//'
```

### AC-AEX-009 — canonical 경로 resolve (project root 상대)
- **Given** execute 경로가 project root `<root>`을 가진 상태
- **When** Applier + Pipeline + Observer를 위한 경로들을 구성
- **Then** 다음 4개 경로가 `<root>` 상대 canonical 경로로 resolve됨 (design.md §B Wiring Recipe):
  - `snapshotBase` = `<root>/.moai/harness/learning-history/snapshots`
  - `manifestPath` = `<root>/.moai/harness/learning-history/manifest.jsonl`
  - `baselinePath` = `<root>/.moai/harness/measurements-baseline.yaml`
  - `usageLogPath` = `<root>/.moai/harness/usage-log.jsonl`
- **And** user-supplied `--project-root`는 `filepath.Abs`로 절대화됨 (internal/cli/CLAUDE.md 절대경로 규칙 — `filepath.Join(cwd, userPath)` 금지)
- **검증** (table-driven test로 4개 경로 join 검증 + 절대경로 규칙):
```bash
go test -run 'TestExecute_ResolvesCanonicalHarnessPaths$' ./internal/cli/harness/...
```
- 추가 (exact-token, `-n`-aware comment filter — 경로 const 사용 확인):
```bash
grep -nE 'measurements-baseline\.yaml|learning-history|usage-log\.jsonl' internal/cli/harness/execute.go | grep -v '^[^:]*:[0-9]*:[[:space:]]*//'
```

### AC-AEX-010 — first apply-outcome telemetry
- **Given** baseline 파일 부재 (first run), 통과하는 proposal
- **When** execute 경로로 정상 apply
- **Then** `<root>/.moai/harness/usage-log.jsonl`에 `apply_outcome` event line 1건 추가됨 (first-run baseline 채택, verdict="kept")
- **검증** (통합 테스트 — t.TempDir() 프로젝트):
```bash
go test -run 'TestExecute_FirstApply_WritesApplyOutcomeTelemetry$' ./internal/cli/harness/...
```
- 추가: outcome line이 `EventTypeApplyOutcome` (`apply_outcome`) 타입 + `OutcomeVerdict:"kept"` 포함.

### AC-AEX-011 — nil/empty sessions L2 nil-safe
- **Given** recent-session metrics 없음
- **When** nil/empty `[]harness.Session`으로 `Applier.Apply()` 호출
- **Then** L2 Canary가 reject하지 않음 (baselineScore=0, drop=0)
- **검증**:
```bash
go test -run 'TestExecute_NilSessions_CanaryDoesNotReject$' ./internal/cli/harness/...
```

### AC-AEX-012 / AC-AEX-013 / AC-AEX-015 — error → exit code
- **Given** 다양한 에러 시나리오
- **When**/**Then**:
  - proposal 파일 부재 OR L1~L4 rejection → exit 1
  - `*ApplyRegressionError` → exit 1
  - measurement-exec wrapped error → exit 2
- **검증** (`errors.As` 타입 분기 — errjoin walk 가능):
```bash
go test -run 'TestExecute_MissingProposal_ExitsUserError$' ./internal/cli/harness/...
go test -run 'TestExecute_RegressionError_ExitsUserError$' ./internal/cli/harness/...
go test -run 'TestExecute_MeasurementExecError_ExitsSystemError$' ./internal/cli/harness/...
```

### AC-AEX-016 — C-HRA-008 boundary
- **Given** `internal/cli/harness/` 전체 소스
- **When** boundary guard 스캔
- **Then** `AskUserQuestion(` substring 부재 (execute.go 포함)
- **검증**:
```bash
grep -rn 'AskUserQuestion\|mcp__askuser' internal/cli/harness/ | grep -v '_test.go' | grep -v '^[^:]*:[0-9]*:[[:space:]]*//' || echo "BOUNDARY_CLEAN"
go test -run 'TestPropose_NoAskUserQuestion$' ./internal/cli/harness/...
```

## §D.2 Edge Cases

- EC-1: proposal ID에 경로 traversal (`../`) 시도 → loader가 base 디렉터리 내로 제약 (절대경로 규칙, exit 1).
- EC-2: baseline 파일 존재 + markdown-only 변경 → Δ=0 → kept (regression 없음, first-run 아님).
- EC-3: proposals 디렉터리 자체 부재 → exit 1 (user error, "no proposal").
- EC-4: Windows 빌드 (`GOOS=windows`) — execute.go는 cross-platform (syscall.Exec/tmux 미사용, 순수 파일 IO).

## §D.3 Definition of Done

- [ ] M1~M5 전 AC PASS (총 16개: MUST 14개 + SHOULD 2개 — §D AC Matrix가 SSOT)
- [ ] `go test -count=1 ./internal/cli/harness/... ./internal/harness/...` green
- [ ] `GOOS=windows GOARCH=amd64 go build ./...` exit 0
- [ ] `golangci-lint run` 0 finding (대상 패키지)
- [ ] C-HRA-008 boundary guard PASS
- [ ] coverage: `internal/cli/harness/execute.go` ≥ 85%, 패키지 비회귀
- [ ] FROZEN 파일 (applier/regression_gate/outcome/pipeline/canary/lineage) diff = 0
- [ ] harness.yaml diff = 0
- [ ] CHANGELOG framing = telemetry (NOT "prevents regressions")

## §D.4 Quality Gate Criteria

- 모든 MUST AC PASS (must-pass — 하나라도 실패 시 reject).
- HONEST FRAMING 준수 (§A.2): sync-auditor가 "회귀 방지" 과대광고 적발 시 SHOULD-FIX.
- autoApply contract invariant (AC-AEX-007/014)는 sync-auditor adversarial probe 대상. AC-AEX-007은 프로덕션 배선이 `AutoApply: true`를 전달함을 검증(non-vacuous); AC-AEX-014는 stub evaluator로 `*ApplyPendingError`를 강제 발생시켜 error-classification 분기가 exit 2 (INVARIANT VIOLATION)로 매핑함을 실증(tautology 아님). sync-auditor는 두 AC가 stub만 증명하지 않고 통합/분기를 증명하는지 확인한다.
