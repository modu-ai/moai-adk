# Design — SPEC-HARNESS-APPLY-EXECUTE-001

## §A. 개요

`Applier.Apply()`(프로덕션 caller 0)를 처음으로 live 경로에 배선하는 opt-in Go execute verb의 설계. autoApply contract(§B.1 spec.md)가 핵심 — L5 재pending을 in-memory `AutoApply: true`로 회피하면서 harness.yaml 디스크 default를 보존한다.

## §B. Wiring Recipe (RunExecute 본문 — conceptual)

```go
// 경로 resolve (project root 상대, 절대경로 규칙 per internal/cli/CLAUDE.md)
proposalPath  := filepath.Join(root, ".moai", "harness", "proposals", id+".json")
snapshotBase  := filepath.Join(root, ".moai", "harness", "learning-history", "snapshots")
manifestPath  := filepath.Join(root, ".moai", "harness", "learning-history", "manifest.jsonl")
baselinePath  := filepath.Join(root, ".moai", "harness", "measurements-baseline.yaml")
usageLogPath  := filepath.Join(root, ".moai", "harness", "usage-log.jsonl")
violationLog  := filepath.Join(root, ".moai", "harness", "frozen-guard-violations.jsonl")
rateLimitPath := filepath.Join(root, ".moai", "harness", "rate-limit-state.json")

proposal := loadProposalByID(proposalPath)  // → harness.Proposal (REQ-AEX-004); 부재 → exit 1

// autoApply contract: AutoApply=true (in-memory ONLY — harness.yaml 디스크 불변)
// L1~L4 강제, L5 auto-approve (오케스트레이터가 C-HRA-008 경계에서 이미 L5 승인 획득)
pipeline := safety.NewPipeline(safety.PipelineConfig{
    ViolationLogPath: violationLog,
    RateLimitPath:    rateLimitPath,
    AutoApply:        true,                 // ← L5 재pending 회피 (the contract)
})

applier := harness.NewApplierWithRegressionGate(manifestPath, baselinePath).
    WithOutcomeObserver(harness.NewObserver(usageLogPath))

var sessions []harness.Session             // nil (first run, recent-session metrics 없음 — REQ-AEX-011)

err := applier.Apply(proposal, pipeline, snapshotBase, sessions)
// err 분기 → §D Error mapping
```

## §C. autoApply contract 흐름 (정상 경로)

```
RunExecute(--execute --id X)
  └─ pipeline = NewPipeline(AutoApply: true)
  └─ applier.Apply(proposal, pipeline, snapshotBase, nil)
       └─ pipeline.Evaluate()
            L1 FrozenGuard   → pass (proposal.TargetPath가 FROZEN 아님)
            L2 Canary        → pass (sessions nil → baseline 0, drop 0)
            L3 Contradiction → pass (Phase 3 empty report)
            L4 RateLimiter   → pass
            L5 Oversight     → AutoApply=true → DecisionApproved  ★ (재pending 회피)
       └─ createSnapshot
       └─ gateActive()=true → applyWithRegressionGate
            (1) baseline measure (markdown-only → Δ=0)
            (2) applyFileModification (frontmatter write)
            (3) candidate measure
            (4) compare → first-run baseline 채택 (no prior baseline)
            (5) keep + baselineStore.Save + writeLineage("approved")
            (6) recordOutcome("kept", ...) → usage-log.jsonl에 apply_outcome 1건  ★ (first telemetry)
  └─ err = nil → exit 0
```

## §D. Decision 표 (verb shape 최종 결정)

| 옵션 | 설명 | 장점 | 단점 | 채택 |
|------|------|------|------|------|
| (a) `apply` verb에 `--execute`/`--id` flag만 추가 (harness.go, package cli) | 기존 verb 확장 | UX 단순 (`apply --execute`) | harness.go는 boundary guard 밖 디렉터리(package cli) → execute 로직이 guard 커버 안 됨 (AP-2) | 부분 |
| **(b) 신규 `execute.go` verb-factory (internal/cli/harness/, package harness)** | propose.go/install.go 미러 | **boundary guard 커버**, exit-code 패턴 재사용, ANCHOR 격리 | 신규 파일 1개 | **✅ 채택** |

**최종 verb shape (하이브리드 — 두 UX 모두 노출)**:
1. **Primary 로직**: `internal/cli/harness/execute.go`의 `RunExecute` + `NewExecuteCmd()` (boundary guard 커버 — REQ-AEX-001/002). `moai harness execute --id <id>` UX 제공.
2. **`apply --execute` UX 위임**: 기존 `newHarnessApplyCmd`(harness.go)에 `--execute`+`--id` flag 추가. flag 존재 → `harnesscli.RunExecute(opts)` 위임 (resume intent의 `apply --execute <id>` UX 보존); flag 부재 → 기존 payload-only (REQ-AEX-003, C4).

이로써 (a)의 `apply --execute` UX와 (b)의 boundary-guard 커버리지를 모두 확보한다. execute의 **모든 실제 로직은 internal/cli/harness/에 위치**하여 guard 커버; harness.go의 `--execute` 분기는 단순 위임 1줄(로직 없음)이므로 guard 밖이어도 안전.

## §E. Error surface → exit code mapping

`errors.As` 타입 분기 (errjoin 선례 — `*ApplyRegressionError`가 `errors.Join`으로 감싸여도 walk 가능):

| 반환 에러 | 분기 | exit | 진단 메시지 | REQ |
|-----------|------|------|------------|-----|
| `nil` | 성공 | 0 | success + telemetry 기록 알림 | AEX-010 |
| proposal 파일 부재 | loader | 1 | "proposal not found: <id>" | AEX-012 |
| rejection error (L1~L4) | `Apply` 반환 | 1 | "proposal rejected (L<n>)" | AEX-012 |
| `*ApplyRegressionError` | `errors.As` | 1 | "regression gate rolled back: regressed=<dims>" | AEX-013 |
| `*ApplyPendingError` (AutoApply=true) | `errors.As` | **2** | "INVARIANT VIOLATION: autoApply contract — Pending under AutoApply=true" | AEX-014 |
| measurement-exec wrapped error | else (system) | 2 | "measurement failed (fail-closed): <err>" | AEX-015 |

분기 순서: `*ApplyPendingError`(invariant, 가장 먼저 명시 검출) → `*ApplyRegressionError` → 기타 rejection(string match "rejected") → 기타(system).

## §F. sessions sourcing (L2 Canary nil-safe 분석)

`canary.go` 분석 결과 (코드 읽고 확인):
- `baselineScore([])` → `n==0` → return 0 (canary.go:33~35).
- `EvaluateCanaryWithScorer`: `baseline=0`, `projected=defaultProjectedScorer(proposal, 0)`.
- meaningful proposal (TargetPath/NewValue 비어있지 않음) → `projected = 0 + 0.02 = 0.02`.
- `drop = baseline - projected = 0 - 0.02 = -0.02` → `drop >= 0.10` 거짓 → **미reject**.

**결론**: first execute run에서 nil/empty sessions는 안전 — L2가 reject하지 않는다 (REQ-AEX-011). 향후 usage-log에서 session 로드는 Phase 5 deferral; 본 SPEC은 nil 전달.

## §F.1 [D4 결정] 테스트 seam — 확정 (run-phase 이연 아님)

**문제**: `NewApplierWithRegressionGate`는 `newGoMeasurer()`를 하드와이어(applier.go:88)하므로 실제 `go test`/`go vet`를 실행한다. `WithMeasurer` seam은 존재하지 않고, 추가하는 것은 C2/C3 FROZEN 위반(applier.go 로직 변경 금지)이다. 따라서 verb를 "있는 그대로" 통합 테스트하면 테스트 내부에서 실제 `go test`로 재귀한다.

**확정 결정 (2-tier 테스트 분할 — coin-flip 아님, run-phase에서 변경 불가)**:

| Tier | 테스트 위치 (package) | 커버 대상 | measurer / evaluator 처리 | 실제 go test 재귀? |
|------|----------------------|-----------|---------------------------|---------------------|
| T1 (white-box, harness 내부) | `package harness` (`internal/harness/execute_integration_test.go` 신규) | Apply-outcome 통합 (gate-active Apply → `apply_outcome` telemetry 기록, AC-AEX-010) | **same-package 접근**으로 `Applier{measurer: stubMeasurer{...}, baselineStore: NewBaselineStore(t.TempDir())}`를 직접 구성 (regression_gate_test.go:174 `stubMeasurer`가 이미 same-package `Measurer` 구현체로 존재) | **아니오** — stub measurer가 고정 triple 반환, 실제 toolchain 미실행 |
| T2 (black-box, cli 경계) | `package harness` (`internal/cli/harness/execute_test.go` 신규) | verb-shape + flag delegation (AC-AEX-001/002/003) + error→exit-code 매핑 (AC-AEX-012/013/014/015) | `Applier.Apply`의 2번째 파라미터 `SafetyEvaluator`(exported interface, applier.go:194)에 **cross-package stub evaluator 주입** (`DecisionPendingApproval`/rejection/Approved 반환). RunExecute는 evaluator + applier를 주입 가능한 형태로 구성 | **아니오** — evaluator stub이 결정을 좌우, gate-active Apply 미도달 (또는 stub measurer 주입한 applier 함께 전달) |

**근거 (ground-truth 확인됨)**:
- `stubMeasurer` (regression_gate_test.go:174~180)는 `package harness` 내부 test 타입이며 `Measurer` 인터페이스(regression_gate.go:196)를 구현한다. white-box(T1) 테스트는 same-package이므로 unexported `measurer`/`baselineStore` 필드를 직접 세팅하여 gate-active Applier를 stub measurer로 구성할 수 있다 → 실제 `go test` 재귀 회피.
- `SafetyEvaluator`(applier.go:194~196)는 **exported interface**이므로 cross-package(T2, `internal/cli/harness`)에서도 stub evaluator 주입이 가능하다 → `*ApplyPendingError`/rejection/Approved 분기를 실제로 강제 발생시켜 RunExecute의 error→exit-code 매핑(AC-AEX-014 등)을 non-vacuous하게 검증.

**RunExecute 구성 제약 (테스트 가능성 — C2/C3 위반 없이)**:
- RunExecute는 `SafetyEvaluator`와 `*Applier`를 **주입받는 내부 함수**로 분해한다 (예: `runExecuteWith(opts, evaluator SafetyEvaluator, applier *Applier) error`). production `RunExecute`는 `safety.NewPipeline(AutoApply:true)` + `NewApplierWithRegressionGate(...).WithOutcomeObserver(...)`를 생성해 위 내부 함수로 위임한다.
- 이 분해는 `internal/cli/harness/execute.go`(신규 파일) 내부에서만 일어나며 `applier.go`/`pipeline.go`/FROZEN 파일을 일절 수정하지 않는다 → C2/C3 보존.
- T1은 `package harness` 내부의 신규 `_test.go`에서 직접 stub measurer로 Apply를 호출하므로 verb를 거치지 않는다 (telemetry 통합 검증의 SSOT).

**경계**: T2는 실제 measurer를 절대 활성화하지 않는다 (stub evaluator가 Approved를 반환하더라도 stub measurer를 함께 주입하거나, error-branch 검증의 경우 Apply가 gate 도달 전 반환). 실제 toolchain 측정은 T1의 stub measurer로만 대체되며, 어느 tier도 테스트 내부에서 실제 `go test`를 재귀 호출하지 않는다.

## §G. MX Tag 설계

| 대상 | 태그 | 내용 |
|------|------|------|
| `Applier.Apply` (applier.go:245 기존 ANCHOR) | `@MX:ANCHOR` 갱신 | caller 목록에 "harness execute CLI (RunExecute)" 추가 (첫 프로덕션 caller). NEVER auto-delete. |
| `safety.NewPipeline` (pipeline.go:55) | `@MX:ANCHOR` 후보 | 첫 프로덕션 caller → fan_in test-only→production 전환. `@MX:REASON` 필수. |
| `RunExecute` (신규) | `@MX:NOTE` | autoApply=true는 L5 재pending 회피 의도; 오케스트레이터가 C-HRA-008 경계에서 이미 L5 승인 획득. |
| `RunExecute` AutoApply 분기 (신규) | `@MX:WARN` + `@MX:REASON` | AutoApply=true는 in-memory PipelineConfig 전용 — harness.yaml 디스크 mutate 금지 (FROZEN 불변식). |

## §H. 경계 검증 (boundary-verification.md 적용)

CLI(internal/cli/harness) ↔ harness pkg 경계:
- proposal JSON 스키마 ↔ `harness.Proposal` struct 필드 일치 (TargetPath/FieldKey/NewValue/ID/PatternKey/Tier — types.go:268).
- exit code 계약 ↔ error 타입 (`*ApplyPendingError`/`*ApplyRegressionError`/rejection) 정합.
- usage-log.jsonl line 스키마 ↔ `EventTypeApplyOutcome` 필드 (outcome.go:57~73).
