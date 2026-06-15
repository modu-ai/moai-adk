---
id: SPEC-HARNESS-EXECUTE-E2E-001
title: "Harness execute regression-gate가 잘못된 측정 root(snapshot base)를 사용해 항상 fail-close하는 production 결함 수정 + e2e 재현 테스트"
version: "0.2.0"
status: implemented
created: 2026-06-15
updated: 2026-06-15
author: manager-spec
priority: P2
phase: "v3.0.0"
module: "internal/cli/harness, internal/harness"
lifecycle: spec-anchored
tags: "harness, telemetry, e2e, test, tdd, bugfix, regression-gate"
era: V3R6
tier: M
---

# SPEC-HARNESS-EXECUTE-E2E-001 — Harness execute regression-gate 측정 root 결함 수정 + e2e 재현 테스트

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-06-15 | manager-spec | plan-phase 산출물 최초 작성 (Tier S, test-only). 선행 SPEC-HARNESS-APPLY-EXECUTE-001 sync-auditor Defect 2 (telemetry happy-path e2e 미검증) 해소 목표. |
| 0.2.0 | 2026-06-15 | manager-spec | **scope 전환: test-only → bug-fix SPEC (Tier M, cycle_type=tdd, Reproduction-First Rule 4).** plan-auditor가 happy-path AC를 empirical probe로 검증한 결과 FAIL 0.58 (Testability 0.35) — 해당 AC가 FROZEN 제약 하에서 **달성 불가**임이 드러났고, 그 원인이 **실제 production 결함**(regression gate가 측정 root로 snapshot base를 사용해 모든 genuine proposal에서 fail-close)임을 2회 empirical 확인. e2e 테스트를 재현 테스트로 삼고 production fix를 동반한다. |

## §A — 배경 / 동기 (WHY)

### §A.1 선행 SPEC가 남긴 보증 갭 (원 동기)

선행 SPEC `SPEC-HARNESS-APPLY-EXECUTE-001` (completed, origin/main `5a22483e4`)은
`moai harness apply --execute --id <id>` verb로 `Applier.Apply()`의 첫 프로덕션
caller를 배선했다. 그 sync-auditor 보고서(PASS-WITH-DEBT)는 단 하나의 진짜 보증
갭을 deferred SHOULD-FIX **Defect 2 (Craft)**로 식별했다 — telemetry happy-path가
end-to-end로 검증되지 않음. `RunExecute` → 실제 `safety.Pipeline(AutoApply:true)`
→ `DecisionApproved` → `Apply` → usage-log.jsonl 경로가 한 번도 통째로 exercise
되지 않았다는 것.

### §A.2 plan-audit가 드러낸 진실: AC는 달성 불가였고, 원인은 production 결함이다

`SPEC-HARNESS-EXECUTE-E2E-001` v0.1.0 (test-only)을 plan-auditor가 감사한 결과
**FAIL 0.58 (Testability 0.35)**. 핵심 happy-path AC("`RunExecute`가 비-FROZEN
proposal에 대해 `verdict="kept"` apply_outcome 라인을 기록")가 FROZEN 제약 하에서
**empirically UNSATISFIABLE**임이 드러났다. plan-auditor의 empirical probe가
밝힌 production 결함을 본 작성자가 **2회 독립 재현**했다.

**Root cause (production 결함)**:

`internal/harness/applier.go` `measurementRoot(snapshotDir)` (line 484)이
`filepath.Dir(snapshotDir)`를 반환한다. `snapshotDir`은
`<snapshotBase>/<ISO-DATE>/` 이므로 그 부모는 **snapshot BASE 디렉터리**
(`<root>/.moai/harness/learning-history/snapshots`)이지 **project root가 아니다**.
applier.go:482 godoc은 "derives the project root for measurement … snapshotBase's
grandparent"라고 적혀 있으나, 실제 코드는 (a) grandparent가 아닌 parent를
반환하고 (b) 진짜 project root는 snapshotBase로부터 **4단계 위**다 — 즉 intent와
behavior가 불일치하며 어느 쪽으로 읽어도 project root에 도달하지 못한다.

`goMeasurer.Measure` (regression_gate.go:222)는 `testCmd.Dir = projectRoot`로
`go test -count=1 -json -coverprofile ./...`를 실행하고 stdout만 캡처한다
(`testCmd.Output()`). 측정 root가 snapshot base이면:

- **CASE 1 (fresh tempdir, go.mod 없음)**: `go test ./...` → `"Action":"build-fail"`
  → fail-closed 분기(`bytes.Contains(testOut, "build-fail")`, regression_gate.go:250).
- **CASE 2 (실제 moai-adk-go 모듈 내부의 빈 디렉터리 = 실제-프로젝트 snapshot base
  모사)**: stdout = **0 bytes** ("matched no packages / no packages to test"가
  STDERR로 가고 `.Output()`이 버림) → `len(testOut)==0` → fail-closed 분기
  (regression_gate.go:242, "go test produced no output").

**결론**: 출시된 `moai harness apply --execute --id <id>` verb는 **어떤 컨텍스트의
어떤 genuine proposal에서도 fail-close한다** — testable Go 패키지를 절대 보유하지
않는 디렉터리를 측정하기 때문이다. 지금까지 surface되지 않은 이유는 (1) genuine
proposal이 존재하지 않고(propose가 no-op) (2) 컴포넌트 테스트가 measurer를 stub
했기 때문이다 (이것이 정확히 sync-auditor WI-3 GAP).

### §A.3 이 SPEC의 가치 (bug-fix)

CLAUDE.md Rule 4 (Reproduction-First Bug Fixing)에 따라 e2e 테스트를 **재현
테스트**로 삼는다. 미수정 코드에서 fail-close(telemetry 0건)로 RED, 수정 후
`verdict="kept"` telemetry 1건으로 GREEN. fix는 regression gate가 **실제 project
root**를 측정하도록 만든다. 이로써 선행 SPEC의 보증 갭(Defect 2)을 닫는 동시에
출시된 verb의 결함을 제거한다.

## §B — 범위 (WHAT) + FROZEN 제약

### §B.1 In Scope — production fix + 재현 테스트

1. **FIX**: regression gate가 측정하는 root를 snapshot base가 아닌 **실제 project
   root**로 바로잡는다. plan.md §F.1에서 두 접근을 평가하고 1개를 채택:
   - **(채택) Approach A — 명시적 project root threading**: `Applier`에
     `projectRoot` 필드 추가 + `WithProjectRoot(root)` 빌더(기존
     `WithOutcomeObserver` 빌더 미러링). `RunExecute`(execute.go:140 fluent chain)가
     `.WithProjectRoot(root)` 호출. `applyWithRegressionGate`가 `a.projectRoot`가
     set이면 그것을, unset이면 기존 `measurementRoot(snapshotDir)` fallback을 사용.
     → 기존 gate 호출자/테스트 무영향, `RunExecute`만 opt-in.
   - **(기각) Approach B — `measurementRoot` 자체를 4단계 strip**: applier.go에
     cross-package 경로 깊이(`execSnapshotBaseRel`) 지식을 하드코딩하게 되어
     계층 누수. plan.md §F.1 trade-off에서 기각.
2. **REPRODUCTION TEST (Rule 4)**: e2e 테스트가 재현 테스트다. RED = `RunExecute`를
   real 최소 Go 모듈인 `t.TempDir()` 프로젝트(`go.mod` + project root에 통과하는
   trivial 테스트 1건 + proposal용 비-FROZEN markdown 타겟)에서 구동 → 미수정
   코드는 fail-close(regression-gate baseline/candidate measurement error, telemetry
   0건). GREEN = 수정 후 `RunExecute` → 실제 `safety.Pipeline(AutoApply:true)` →
   `DecisionApproved` → `Apply` → (이제 올바른) project root에서 real measurer 실행
   (`go test` 통과, Δ=0) → `verdict="kept"` apply_outcome 라인이 temp
   usage-log.jsonl에 기록.

### §B.2 FROZEN 제약 (수정/동결 구분 — HARD)

**FIX TARGETS (NOT frozen — 이 SPEC이 수정)**:

| 파일 | 수정 내용 |
|------|-----------|
| `internal/harness/applier.go` | `projectRoot` 필드 + `WithProjectRoot(root)` 빌더 추가; `applyWithRegressionGate`가 set 시 `a.projectRoot` 사용, unset 시 `measurementRoot(snapshotDir)` fallback. `measurementRoot` godoc의 intent-vs-behavior 불일치 주석 정정. |
| `internal/cli/harness/execute.go` | `RunExecute`의 fluent chain(line ~140)에 `.WithProjectRoot(root)` 추가 (root는 이미 절대화된 project root). |

**STILL FROZEN (diff 0 의무)**:

| 파일 | 비고 |
|------|------|
| `internal/harness/safety/pipeline.go` | L1-L5 Evaluate — 결함과 무관, 동결 |
| `internal/harness/outcome.go` | `RecordOutcome` — 동결 |
| `internal/harness/observer.go` | `NewObserver` — 동결 |
| `internal/harness/regression_gate.go` | `goMeasurer`/`Measure` 로직은 **올바르다** — 주어진 root를 측정할 뿐. 결함은 전달된 root가 틀린 것이지 measurer가 아니다. 동결. |
| `.moai/config/sections/harness.yaml` | 디스크 `auto_apply: false` 불변 — AutoApply는 in-memory 유지 |

### §B.3 telemetry 오염 금지 불변식 (HARD)

모든 synthetic proposal·go.mod·trivial 테스트·usage-log write는 엄격히
`t.TempDir()` 내부에서만 발생한다. 실제 레포의
`.moai/harness/usage-log.jsonl` 및 `.moai/harness/measurements-baseline.yaml`,
`.moai/harness/regression-coverage.out`은 테스트에 의해 절대 touch되지 않는다.
재현 테스트는 `ExecuteOptions.ProjectRoot`를 `t.TempDir()`로 지정하여 격리한다.

### §B.4 subagent boundary 보존 불변식 (HARD)

새 테스트 파일은 `AskUserQuestion(` 호출을 일절 도입하지 않는다. 패키지 가드
`TestPropose_NoAskUserQuestion` (`internal/cli/harness/propose_boundary_test.go`)이
패키지 디렉터리를 스캔하므로 새 테스트 파일도 자동으로 커버된다.

### §B.5 COST NOTE (결함 아님 — 명시)

`measurementRoot`를 실제 project root로 바로잡으면, 대형 실제 프로젝트에서
`--execute`는 baseline + candidate 각각 `go test ./...`를 1회씩, 총 **2회**(각
5분 timeout) 실행한다. 이는 regression gate의 목적(apply 전후 project-health
측정)에 내재된 비용이며 opt-in verb에 대해 **수용 가능**하다 — 결함이 아니다.
재현 **테스트**는 fixture 모듈이 극소(go.mod + trivial 테스트 1건)이므로 빠르게
유지된다. 이 사실을 명시해 향후 결함 오인을 방지한다.

## §C — GEARS 요구사항 (REQUIREMENTS)

GEARS 표기. `<subject>`는 일반화 명사(test, applier, regression gate 등).

### §C.1 production fix 요구사항

- **REQ-E2E-001 (Ubiquitous)**: The regression gate, when active, **shall**
  measure project-health at the actual project root — NOT the snapshot base
  directory.

- **REQ-E2E-002 (Capability gate — Where)**: **Where** an explicit project root
  is wired via `WithProjectRoot(root)`, the applier **shall** pass that root to
  the measurer; **where** it is unset, the applier **shall** fall back to the
  pre-existing `measurementRoot(snapshotDir)` derivation so that existing gate
  callers and tests are unaffected.

- **REQ-E2E-003 (Event-driven)**: **When** `RunExecute` constructs the
  gate-active applier (execute.go), it **shall** wire the already-absolutized
  project root via `.WithProjectRoot(root)`.

### §C.2 재현 테스트 요구사항 (RED→GREEN)

- **REQ-E2E-004 (Event-driven — RED)**: **When** the e2e reproduction test drives
  `RunExecute` against the UNFIXED code with a valid non-FROZEN proposal in a
  `t.TempDir()` minimal Go module, the apply **shall** fail closed (regression-
  gate baseline/candidate measurement error) and write NO `apply_outcome`
  telemetry line — demonstrating the production defect.

- **REQ-E2E-005 (Event-driven — GREEN)**: **When** the same test drives
  `RunExecute` against the FIXED code, the production pipeline **shall** evaluate
  L1-L5 with real implementations, reach `DecisionApproved`, run the real
  measurer at the corrected project root (the fixture's `go test ./...` passes,
  Δ=0), and write exactly one `apply_outcome` line with `outcome_verdict ==
  "kept"` to the temp `usage-log.jsonl`.

- **REQ-E2E-006 (Ubiquitous)**: The reproduction test **shall** call the
  production entry point `RunExecute` directly — NOT a hand-constructed
  `&Applier{...}` value, NOT `applier.Apply` directly, NOT `runExecuteWith` with
  a stub evaluator, and NOT a stub measurer.

### §C.3 fixture / 격리 / 동결 요구사항

- **REQ-E2E-007 (Capability gate — Where)**: **Where** the test fixture must
  reach `DecisionApproved` and a real measurer pass, the `t.TempDir()` project
  **shall** be a real minimal Go module — a `go.mod` plus one trivial passing
  Go test at the project root plus a non-FROZEN markdown target file (with
  frontmatter) for the proposal — so `go test ./...` succeeds at the corrected
  project root.

- **REQ-E2E-008 (Ubiquitous)**: All test writes (go.mod, trivial test, proposal,
  target file, usage-log, baseline, coverage profile) **shall** occur inside
  `t.TempDir()` via `ExecuteOptions.ProjectRoot`; the real repository's
  `.moai/harness/*` files **shall not** be written.

- **REQ-E2E-009 (Ubiquitous)**: This SPEC **shall not** modify any FROZEN file
  (the §B.2 STILL-FROZEN set); `pipeline.go`, `outcome.go`, `observer.go`,
  `regression_gate.go`, and `harness.yaml` **shall** have diff 0.

- **REQ-E2E-010 (Event-detected — replaces unwanted)**: **When** a future change
  introduces an `AskUserQuestion(` call into the new test file, the package guard
  `TestPropose_NoAskUserQuestion` **shall** detect it.

- **REQ-E2E-011 (State-driven — While)**: **While** the execute path passes
  `nil` sessions to the pipeline (first run, no recent-session metrics), the L2
  Canary check **shall not** reject the happy-path proposal, so that reaching
  `DecisionApproved` is genuine (canary.go nil-safe).

## §D — Exclusions (What NOT to Build)

[HARD] 다음은 이 SPEC의 범위 밖이며 명시적으로 제외한다.

### §D.1 Out of Scope

- **measurer(goMeasurer/Measure) 로직 변경 금지.** 결함은 전달된 root가 틀린 것이지
  measurer가 아니다. measurer는 STILL FROZEN(§B.2).
- **pipeline / outcome / observer / regression_gate / harness.yaml 변경 금지**
  (§B.2 STILL FROZEN, diff 0).
- **measurer를 stub하여 happy-path를 통과시키는 것 금지.** 그것이 정확히 sync-auditor
  WI-3 GAP이며 plan-auditor가 FAIL시킨 패턴이다. 재현 테스트는 real measurer를
  real 최소 Go 모듈에서 구동한다.
- **D3 패키지 커버리지 90% 절대 바 달성 제외.** 선행 SPEC sync-auditor Defect 3은
  `internal/cli/harness` 패키지 커버리지 77.9% < 90%를 baseline-inherited debt로
  분류했다. 이 SPEC은 비회귀(파일·패키지 커버리지 동등 이상)만 gating하며 90% 절대
  바 달성은 별도 follow-up SPEC 결정으로 남긴다.
- **rolled-back / regression-blocked verdict 변형 테스트 제외.** markdown-only
  FROZEN write surface에서 Δ=0 always-pass이므로 gate-active 정상 경로에서 rolled-back
  은 자연 발생하지 않는다. 이 SPEC은 `verdict="kept"` happy-path와 RED fail-close만
  다룬다.
- **CHANGELOG / docs-site / 4-locale 문서 동기화 제외** (sync-phase manager-docs 소유).
- **harness write surface(FROZEN allowlist) 확장 제외** — 이 SPEC은 측정 root 결함만
  고친다.
- **propose verb를 genuine proposal 생성기로 만드는 작업 제외** — 별도 SPEC 후보.

### §D.2 Out of Scope (성능 / 비용 측면 — 비요구)

- **`--execute`의 2회 `go test ./...` 실행 비용 최적화 제외.** §B.5에 명시한 대로
  이는 regression gate 목적에 내재된 비용이며 결함이 아니다. 비용 최적화(증분 측정,
  캐싱)는 Phase5 deferral.

## §E — 인수 기준 요약 (상세는 acceptance.md)

| AC | 요약 | 검증 |
|----|------|------|
| AC-E2E-001 | 재현 테스트가 `RunExecute`를 직접 호출 (no `&Applier{}`/`applier.Apply`/`runExecuteWith`+stub/stub measurer) | `grep -c` + 본문 read |
| AC-E2E-002 | RED: 미수정 코드에서 fail-close + telemetry 0건 (재현 입증) | RED 커밋 / `go test` 출력 |
| AC-E2E-003 | GREEN: 수정 후 `verdict="kept"` apply_outcome 라인 정확히 1건 + nil error | `go test -run <Name>` PASS |
| AC-E2E-004 | fix: `WithProjectRoot` set 시 measurer가 project root 측정, unset 시 fallback | unit test (set/unset 두 경로) |
| AC-E2E-005 | fixture가 real Go 모듈(go.mod + trivial 통과 테스트 + 비-FROZEN 타겟) | 본문 read + GREEN 통과 |
| AC-E2E-006 | STILL-FROZEN 5 파일 diff 0 | `git diff --stat <baseline> HEAD -- <files>` |
| AC-E2E-007 | 실제 레포 `.moai/harness/*` 미변경 | `git status --porcelain .moai/harness/` |
| AC-E2E-008 | C-HRA-008: 새 테스트에 `AskUserQuestion(` 0건 | `grep -c` + 가드 테스트 PASS |
| AC-E2E-009 | DoD: `internal/cli/harness` + `internal/harness` 패키지 커버리지 비회귀 | `go test -cover` before/after |

`<baseline>` = `bf01fed74` (이 SPEC 작업 시작 시 HEAD; run-phase 진입 시 재확인).

## §F — Definition of Done

- production fix(applier.go + execute.go) 적용, regression gate가 project root 측정
- 재현 e2e 테스트가 RED(미수정 fail-close)→GREEN(수정 후 verdict="kept") 전이 입증
- STILL-FROZEN 5 파일 diff 0
- 실제 레포 `.moai/harness/*` 미변경
- C-HRA-008 가드 clean
- `go test ./internal/cli/harness/ ./internal/harness/` PASS (회귀 0 — 특히 기존
  gate-active stub-measurer 테스트 전부 PASS)
- `internal/cli/harness` + `internal/harness` 패키지 커버리지 비회귀
- `go build ./...` + `GOOS=windows GOARCH=amd64 go build ./...` exit 0
- `go vet ./internal/cli/harness/ ./internal/harness/` exit 0
- `golangci-lint run` clean (변경 파일)

## §G — Cross-References

- 선행 SPEC: `.moai/specs/SPEC-HARNESS-APPLY-EXECUTE-001/`
- 권위 보고서: `.moai/reports/sync-audit/SPEC-HARNESS-APPLY-EXECUTE-001-2026-06-15.md`
  (WI-3 GAP, Defect 2)
- 결함 위치: `internal/harness/applier.go` `measurementRoot` (line 484) +
  `applyWithRegressionGate` (line 397-398)
- 측정기(올바름, 동결): `internal/harness/regression_gate.go` `goMeasurer.Measure`
  (line 222, `testCmd.Dir = projectRoot`, `.Output()` stdout-only)
- fix 배선점: `internal/cli/harness/execute.go` `RunExecute` (line 140 fluent chain)
- CLAUDE.md Rule 4 (Reproduction-First Bug Fixing)
