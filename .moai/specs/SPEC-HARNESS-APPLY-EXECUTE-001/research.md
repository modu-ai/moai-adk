# Research — SPEC-HARNESS-APPLY-EXECUTE-001

manager-spec가 코드를 직접 탐색하여 확보한 ground-truth. run-phase 진입 시 line 번호는 코드 이동 가능성이 있으므로 재확인 권장.

## §A. dual-apply-path 아키텍처 (현황)

harness apply에는 두 경로가 존재:
- **Path S (skill Edit)** — Claude Edit tool이 skill-workflow에서 직접 수행. 현재 모든 실제 apply가 이 경로. Go 파이프라인 미경유.
- **Path G (Go pipeline)** — `Applier.Apply()` 진입점. **프로덕션 caller 0** (dormant).

본 SPEC은 Path G에 첫 프로덕션 caller(opt-in execute verb)를 추가한다. Path S는 default 유지. dual-apply-path 중 어느 것을 canonical로 할지의 아키텍처 결정은 별도 후속 SPEC (predecessor 메모리 `L_harness_apply_pipeline_orphaned_blocks_phase5` 참조).

## §B. 핵심 코드 fact (검증됨)

### Applier 파이프라인 (internal/harness/applier.go)
- `NewApplierWithRegressionGate(manifestPath, baselinePath string) *Applier` (L84) — 프로덕션 seam, goMeasurer + BaselineStore + lineage 배선.
- `(*Applier).WithOutcomeObserver(obs *Observer) *Applier` (L102) — fluent, outcome 캡처 배선.
- `(*Applier).Apply(proposal, evaluator, snapshotBase, sessions) error` (L247):
  - Step1 `evaluator.Evaluate()` → switch Decision.Kind: Rejected(lineage "rejected"+return err) / PendingApproval(return `*ApplyPendingError`) / Approved(continue).
  - Step2 `createSnapshot`.
  - `gateActive()` (measurer+baselineStore 둘 다 non-nil) → `applyWithRegressionGate` (measure→apply→measure→compare→keep/rollback + recordOutcome).
  - 프로덕션 caller = 0 (grep: 정의/주석/테스트만).

### safety Pipeline (internal/harness/safety/pipeline.go)
- `NewPipeline(cfg PipelineConfig) *Pipeline` (L55). `PipelineConfig{ViolationLogPath, RateLimitPath, AutoApply bool}` (L42~52).
- `Evaluate()` (L89): L1 FrozenGuard → L2 Canary → L3 Contradiction → L4 RateLimiter → L5 Oversight.
- **L5 (L147)**: `!autoApply` → 항상 `DecisionPendingApproval` + OversightProposal; `autoApply` → `DecisionApproved`.
- 프로덕션 caller = 0 (`constitution.NewPipeline`은 다른 패키지 — 무관).

### outcome 캡처 (internal/harness/outcome.go)
- `(*Observer).RecordOutcome(rec OutcomeRecord) error` (L57~73) → `Event{EventType: EventTypeApplyOutcome, ...}` → `RecordExtendedEvent` append.
- **타겟 경로 = usage-log.jsonl** (RecordExtendedEvent append path). 별도 outcome 로그 없음 — usage-log.jsonl에 `apply_outcome` line으로 기록.
- recordOutcome seam은 applier.go L454(rolled-back) / L473(kept) 두 terminal branch.

### regression gate (internal/harness/regression_gate.go)
- baseline 파일 = `.moai/harness/measurements-baseline.yaml` (L98 주석 + L117 NewBaselineStore).
- `BaselineStore.Load()` (L124): 파일 부재 → `(zero, false, nil)` (first run, no block, REQ-RG-005).
- `goMeasurer.Measure` (L222): `go test -json -coverprofile` + `go vet` → MetricTriple. build-fail/timeout → fail-closed wrapped error.
- `MetricTriple.Regressions(candidate)` (L50): tests↓ / coverage↓ / lint↑ → regressed dims.
- `ApplyRegressionError` (L71) — `ApplyPendingError`와 구별되는 타입.

### canary nil-safe (internal/harness/safety/canary.go)
- `baselineScore([])` → 0 (L33~35).
- `defaultProjectedScorer(meaningful, baseline)` → baseline+0.02 (L56~62).
- nil sessions → drop=0 → 미reject (drop ≥ 0.10 거짓).

## §C. CLI verb-factory 패턴 (internal/cli/harness/)

- 패키지 `internal/cli/harness/` (sub-dir, package `harness`) — boundary guard `TestPropose_NoAskUserQuestion`(propose_boundary_test.go)이 디렉터리 내 모든 비-test `.go`에서 `AskUserQuestion(` 부재 검증.
- `propose.go` `NewProposeCmd()` + `install.go` `NewInstallCmd()` — export 팩토리, `newHarnessRouterCmd()`(harness_route.go, package cli)에서 등록.
- install.go가 exit-code/절대경로/flag 패턴 reference (RunInstall + NewInstallCmd, projectRoot resolve, MarkFlagRequired).
- 기존 `apply` verb (`newHarnessApplyCmd` harness.go:203 → `runHarnessApply`:219)는 **package cli**(boundary guard 디렉터리 밖), payload-only.

### 경로 const (internal/cli/harness.go)
- `harnessDefaultLogPath = ".moai/harness/usage-log.jsonl"` (L34)
- `harnessDefaultSnapshotBase = ".moai/harness/learning-history/snapshots"` (L37)
- `harnessDefaultProposalDir = ".moai/harness/proposals"` (L40)
- `harnessConfigPath = ".moai/config/sections/harness.yaml"` (L43)

### Proposal struct (internal/harness/types.go:268)
- 필드: ID / TargetPath / FieldKey / NewValue / PatternKey / Tier / ObservationCount / CreatedAt.

## §D. proposal 데이터 상태

- `.moai/harness/learning-history/` 존재. `applied/`·`snapshots/` 하위 부재 (0 applies). `tier-promotions.jsonl` + `usage-log.jsonl` 존재.
- 첫 `--execute` 실행이 첫 snapshot + lineage + apply_outcome 생성.
- run-phase 테스트는 t.TempDir() 프로젝트에 fixture proposal + measurer stub로 통합 검증 (실제 `go test` 재귀 호출 회피 — Measurer 인터페이스 stub 주입).

## §E. predecessor 메모리 계보

- P0: HARNESS-LOOP-CLOSURE-001 (lineage + human-gate).
- P1: HARNESS-REGRESSION-GATE-001 (gate scaffold, honest framing MUST-PASS).
- P2 1차: HARNESS-OUTCOME-CAPTURE-001 (outcome enabler, F2 SHOULD-FIX dormant scaffold).
- P2 후속: HARNESS-OUTCOME-ERRJOIN-001 (rolled-back errors.Join 신호 보존).
- 본 SPEC: P2 activation 1차 — 모든 scaffold를 처음 live 배선 (`Applier.Apply()` 첫 caller).

핵심 발견 (메모리): `Applier.Apply()` 0 production caller → observer/gate activation이 dual-apply-path 아키텍처 결정에 blocked. 본 SPEC은 user decision A(opt-in)로 그 blocker를 우회하여 첫 telemetry 생성.

## §F. 해소된 설계 결정 (plan-phase 확정 — run-phase 이연 아님)

1. **verb 표면 (해소)**: `apply --execute` UX vs `execute` sub-verb — design.md §D는 둘 다 노출(하이브리드)로 확정. 실제 로직은 `internal/cli/harness/execute.go`에 위치하고 `apply --execute`는 단순 위임. run-phase는 이 결정을 따르며 변경하지 않는다.
2. **테스트 seam (해소 — plan-audit iter1 D4)**: design.md §F.1에서 **2-tier 분할로 확정**. T1(white-box, `package harness` 내부)은 same-package `stubMeasurer`(regression_gate_test.go:174, 이미 존재하는 `Measurer` 구현체)로 gate-active Applier를 직접 구성하여 apply-outcome telemetry 통합(AC-AEX-010)을 검증 — 실제 `go test` 재귀 없음. T2(black-box, `internal/cli/harness`)는 exported `SafetyEvaluator` 인터페이스(applier.go:194)에 stub evaluator를 주입하여 verb-shape + error→exit-code 매핑(AC-AEX-012/013/014/015)을 검증. RunExecute는 evaluator/applier 주입 가능한 내부 함수로 분해(execute.go 내부 한정, applier.go/pipeline.go/FROZEN 파일 무수정 → C2/C3 보존). **coin-flip 아님** — ground-truth(`stubMeasurer` same-package 존재 + `SafetyEvaluator` exported)에 근거한 확정 seam. 상세는 design.md §F.1 표 참조.

> §D 67행의 "run-phase 테스트는 ... stub 주입" 기술은 위 2-tier 분할(T1 white-box stubMeasurer)로 구체화되었다.
