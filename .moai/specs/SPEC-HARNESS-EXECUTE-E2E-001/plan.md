---
id: SPEC-HARNESS-EXECUTE-E2E-001
title: "Harness execute regression-gate 측정 root 결함 수정 + e2e 재현 테스트 — 구현 계획"
version: "0.2.0"
status: draft
created: 2026-06-15
updated: 2026-06-15
author: manager-spec
priority: P2
phase: "v3.0.0"
module: "internal/cli/harness, internal/harness"
lifecycle: spec-anchored
tags: "harness, telemetry, e2e, test, tdd, bugfix, regression-gate"
tier: M
---

# SPEC-HARNESS-EXECUTE-E2E-001 — 구현 계획 (plan.md)

> Tier M. cycle_type=tdd. Reproduction-First Bug Fixing (CLAUDE.md Rule 4).
> production fix(applier.go + execute.go) + e2e 재현 테스트. STILL-FROZEN 5 파일 diff 0.

## §A — Context (위치 + 분기 + 산출물)

- **작업 위치**: `/Users/goos/MoAI/moai-adk-go` (project root absolute)
- **현재 baseline**: HEAD `bf01fed74` (run-phase 진입 시 재확인)
- **SPEC 산출물**: `.moai/specs/SPEC-HARNESS-EXECUTE-E2E-001/{spec,plan,acceptance}.md`
- **FIX 대상 (수정)**: `internal/harness/applier.go`, `internal/cli/harness/execute.go`
- **신규 테스트 파일 후보**: `internal/cli/harness/execute_e2e_test.go` (package `harness`,
  `RunExecute`/`buildExecutePipelineConfig`/`resolveExecutePaths`/`writeProposalFixture`
  same-package 접근) + `internal/harness/applier_projectroot_test.go` (또는 기존
  `regression_gate_test.go`에 set/unset 단위 테스트 추가 — package `harness`)
- **STILL-FROZEN (diff 0)**: §A.5

### §A.5 STILL-FROZEN / PRESERVE list (diff 0 의무)

| 파일 | 비고 |
|------|------|
| `internal/harness/safety/pipeline.go` | L1-L5 Evaluate — 동결 |
| `internal/harness/outcome.go` | `RecordOutcome` — 동결 |
| `internal/harness/observer.go` | `NewObserver` — 동결 |
| `internal/harness/regression_gate.go` | `goMeasurer`/`Measure` 올바름 — 동결 |
| `.moai/config/sections/harness.yaml` | 디스크 `auto_apply: false` 불변 |
| 기존 테스트 `internal/cli/harness/execute_test.go` | 무수정 (신규 파일에 추가) |
| 기존 테스트 `internal/harness/execute_integration_test.go` | 무수정 (stub-measurer 컴포넌트 테스트는 그대로 유지) |

## §B — Known Issues (Tier M — 관련 카테고리)

- **B3. C-HRA-008 / Subagent Boundary**: 새 테스트 파일은 `AskUserQuestion(` 미도입.
  검증: `grep -c 'AskUserQuestion(' internal/cli/harness/execute_e2e_test.go` → 0.
  패키지 가드 `TestPropose_NoAskUserQuestion`이 자동 스캔.
- **B8. Working Tree Hygiene**: 실제 레포 `.moai/harness/usage-log.jsonl`,
  `.moai/harness/measurements-baseline.yaml`, `.moai/harness/regression-coverage.out`,
  `.moai/state/`, `.moai/cache/` 변경 절대 금지. 모든 write는 `t.TempDir()` 내부.
- **B10. Untouched Paths PRESERVE**: §A.5 STILL-FROZEN list + FIX 2파일 외 working
  tree 변경 금지. 병렬 세션의 다른 SPEC 디렉터리 / `.moai/research/*` 손대지 말 것.
- **B1. Cross-platform**: FIX 코드는 `filepath` API만 사용 (이미 그러함). 새 테스트는
  `t.TempDir()` 반환값(절대) + `filepath.Join`만 사용.
  `GOOS=windows GOARCH=amd64 go build ./...` 통과 의무.
- **B2. Cross-SPEC 정책 충돌**: `internal/harness` retired/superseded 마커 사전 확인.
  `grep -r "Retired\|superseded\|deprecation-marker" internal/harness internal/cli/harness`.
- **B5. CI 3-tier**: spec-lint / golangci-lint / Test(per OS) 각각 fail 가능. FIX는
  exported API(`WithProjectRoot`) 추가이므로 godoc 의무.

## §C — Pre-flight Check List (착수 전)

```bash
# 1. baseline 확인
git rev-parse HEAD            # expect bf01fed74... (또는 run 진입 시점 HEAD 재기록)
git branch --show-current

# 2. cross-platform 빌드 사전 확인
go build ./...
GOOS=windows GOARCH=amd64 go build ./...

# 3. 패키지 테스트 + 커버리지 baseline (비회귀 비교 기준)
go test -cover ./internal/cli/harness/ ./internal/harness/

# 4. 결함 위치 + measurer 계약 재확인
sed -n '479,490p' internal/harness/applier.go     # measurementRoot 결함
sed -n '397,399p' internal/harness/applier.go     # applyWithRegressionGate root 호출
sed -n '222,235p' internal/harness/regression_gate.go  # testCmd.Dir = projectRoot
sed -n '137,143p' internal/cli/harness/execute.go      # RunExecute fluent chain 배선점

# 5. blast-radius 재확인 (§D.1)
grep -rn 'measurementRoot' internal/ --include="*.go"
grep -rn 'NewApplierWithRegressionGate\|applyWithRegressionGate' internal/ --include="*.go"
grep -rn 'measurer.Measure' internal/ --include="*.go" | grep -v "_test.go"
```

## §D — Constraints (DO NOT VIOLATE) + Blast Radius

### §D.1 Blast Radius 분석 (manager-spec 사전 grep, run 진입 시 재확인)

| 심볼 | production caller | 결론 |
|------|-------------------|------|
| `measurementRoot` | **applier.go:398 (applyWithRegressionGate) 단 1곳** | 다른 패키지 0. 안전하게 변경 가능. |
| `applyWithRegressionGate` | **applier.go:291 (Apply) 단 1곳** | 내부. |
| `NewApplierWithRegressionGate` | **execute.go:140 (RunExecute) 단 1곳** + 테스트 | RunExecute만 opt-in 대상. |
| `goMeasurer.Measure` | applier.go:413, 424 (gate 내부) | measurer는 변경 안 함 (root만 바로잡음). |
| 기존 gate-active 테스트 | `applier_test.go`/`outcome_test.go`/`execute_integration_test.go`/`regression_gate_test.go` — 전부 **stub measurer 주입** (`&Applier{measurer: stubMeasurer{...}}`) | stub은 `projectRoot` 인자 무시 → fallback이든 set이든 **무영향**. Approach A는 `projectRoot` unset 시 기존 동작 보존 → 회귀 0. |

핵심 안전 논거: (1) `measurementRoot` 호출자가 1곳뿐, (2) 모든 기존 gate-active
테스트는 stub measurer라 root를 무시, (3) Approach A는 unset 시 기존 fallback 보존.
따라서 fix는 `RunExecute` 경로에만 새 동작을 주입하고 다른 모든 경로는 byte-동일.

### §D.2 금지/의무

- §A.5 STILL-FROZEN 5 파일 + 2 기존 테스트 diff 0
- 실제 레포 runtime-managed 파일(`.moai/harness/*`, `.moai/state/*`) 변경 금지
- `measurer`(goMeasurer/Measure) 로직 변경 금지 — root만 바로잡음
- Conventional Commits + `🗿 MoAI` trailer (`feat(SPEC-HARNESS-EXECUTE-E2E-001): M{N} ...`)
- `--no-verify` / `--amend` / force-push 금지
- C-HRA-008: 새 테스트에 `AskUserQuestion(` 0건
- 새 exported API(`WithProjectRoot`)는 godoc 필수 + 기존 `WithOutcomeObserver` 빌더 스타일 미러링

## §E — Self-Verification Deliverables

### E1. AC Binary PASS/FAIL Matrix
acceptance.md AC-E2E-001~009 각 행을 검증 커맨드 + 실제 출력으로 채움.

### E2. Cross-Platform Build
```
$ go build ./...                          → exit 0
$ GOOS=windows GOARCH=amd64 go build ./... → exit 0
```

### E3. Coverage (비회귀)
```
$ go test -cover ./internal/cli/harness/ ./internal/harness/   # baseline 대비 동등 이상
```

### E4. Subagent Boundary Grep (C-HRA-008)
```
$ grep -c 'AskUserQuestion(' internal/cli/harness/execute_e2e_test.go   # 0
$ go test -run TestPropose_NoAskUserQuestion ./internal/cli/harness/    # PASS
```

### E5. STILL-FROZEN diff 0
```
$ git diff --stat <baseline> HEAD -- \
    internal/harness/safety/pipeline.go internal/harness/outcome.go \
    internal/harness/observer.go internal/harness/regression_gate.go \
    .moai/config/sections/harness.yaml
  # (no output — 0 diff lines)
```

### E6. 실제 레포 harness runtime 파일 미변경
```
$ git status --porcelain .moai/harness/   # empty
```

### E7. RED→GREEN 재현 증거 (Rule 4)
- RED: fix 미적용(또는 stash) 상태에서 재현 테스트 실행 → fail-close 에러 +
  telemetry 0건 (출력 캡처)
- GREEN: fix 적용 후 동일 테스트 → `verdict="kept"` 1건 PASS (출력 캡처)

## §F — Milestone Breakdown (Tier M)

### §F.1 Fix Approach 선정 (Approach A 채택, B 기각)

| Approach | 변경 | 장점 | 단점 | 판정 |
|----------|------|------|------|------|
| **A. project root threading** | `Applier.projectRoot` 필드 + `WithProjectRoot(root)` 빌더; `applyWithRegressionGate`가 set 시 사용, unset 시 `measurementRoot(snapshotDir)` fallback; `RunExecute`가 `.WithProjectRoot(root)` 호출 | 기존 호출자/테스트 무영향(unset fallback 보존); cross-package 경로 지식 누수 없음; `WithOutcomeObserver` 빌더와 대칭 | 필드 1개 + 빌더 1개 추가 | **채택** |
| B. `measurementRoot` 4단계 strip | `measurementRoot`가 `execSnapshotBaseRel` 깊이만큼 strip하여 root 복원 | 호출부 무변경 | applier.go가 cli 패키지의 경로 상수 깊이(4)를 하드코딩 → 계층 누수, snapshotBase 구조 변경 시 silently 깨짐 | **기각** |

### M1 — RED: 재현 테스트 작성 (미수정 코드에서 fail-close 입증)

1. `internal/cli/harness/execute_e2e_test.go` 생성 (package `harness`).
2. `t.TempDir()` 프로젝트 root를 **real 최소 Go 모듈**로 구성:
   - `<root>/go.mod` 작성 (예: `module e2efixture\n\ngo 1.23\n`).
   - `<root>/fixture_test.go` 작성 — 통과하는 trivial 테스트 1건
     (`func TestFixturePass(t *testing.T) {}`). `go test ./...`가 project root에서
     성공(빌드 OK, 1 package, 0 fail)하도록.
   - 비-FROZEN markdown 타겟 `<root>/docs/e2e-sample.md` (frontmatter + body) 작성.
   - synthetic proposal: same-package `writeProposalFixture(t, root, id, targetAbs,
     "description", "<newValue>")` 재사용 (targetAbs = docs/e2e-sample.md 절대 경로).
3. `RunExecute(ExecuteOptions{ID: id, ProjectRoot: root})` 호출.
4. **RED 어서션 (미수정 코드)**: 반환 error != nil 이고 "measurement" /
   "fail-closed" 부류 에러이며, `<root>/.moai/harness/usage-log.jsonl`에 apply_outcome
   라인이 **0건**임을 확인. (이 시점에서 production 미수정 → 측정 root가 snapshot
   base → fail-close → telemetry 미기록. 재현 성공.)
5. RED 상태 커밋 (Rule 4: reproduction test first). 커밋 메시지에 재현 증거 요약.

### M2 — GREEN(fix): regression gate가 project root 측정하도록 수정

1. `internal/harness/applier.go`:
   - `Applier` struct에 `projectRoot string` 필드 추가 (godoc: set 시 measurer
     root override, unset 시 measurementRoot fallback).
   - `WithProjectRoot(root string) *Applier` 빌더 추가 (receiver 반환, fluent;
     `WithOutcomeObserver` 미러링).
   - `applyWithRegressionGate`의 `projectRoot := measurementRoot(snapshotDir)`를
     `projectRoot := a.measurementProjectRoot(snapshotDir)`(또는 inline 조건)로
     교체: `if a.projectRoot != "" { return a.projectRoot }; return measurementRoot(snapshotDir)`.
   - `measurementRoot` godoc의 intent-vs-behavior 불일치 주석 정정 (snapshot base를
     반환함을 정확히 기술 + fallback 용도 명시).
2. `internal/cli/harness/execute.go`:
   - `RunExecute`의 fluent chain(line ~140)에 `.WithProjectRoot(root)` 추가
     (root는 이미 `filepath.Abs`로 절대화된 project root).
3. **GREEN 어서션**: M1 재현 테스트가 이제 통과 — 반환 error == nil + apply_outcome
   라인 정확히 1건 + `outcome_verdict == "kept"`. (수정 후 측정 root = project root
   = fixture 모듈 → `go test ./...` 통과 → Δ=0 → kept.)

### M3 — fix 단위 검증 + 회귀 0 확인

1. `WithProjectRoot` set/unset 단위 테스트 추가 (package `harness`):
   - set: `NewApplierWithRegressionGate(...).WithProjectRoot("/x")` → 내부적으로
     `/x`가 measurer에 전달됨을 stub measurer로 관측 (measurer가 받은 root 캡처).
   - unset: `WithProjectRoot` 미호출 → 기존 `measurementRoot(snapshotDir)` fallback
     경로 사용됨을 관측.
2. 전체 패키지 회귀 0: `go test ./internal/cli/harness/ ./internal/harness/` —
   특히 기존 gate-active stub-measurer 테스트(`applier_test.go`/`outcome_test.go`/
   `execute_integration_test.go`/`regression_gate_test.go`) 전부 PASS 확인 (stub은
   root 무시 → 무영향).
3. STILL-FROZEN diff 0 + cross-platform build + 커버리지 비회귀 + C-HRA-008 검증.

> **non-vacuous 보장 (D4 해소)**: non-vacuous proof는 scratch mis-assertion이 아니라
> **genuine RED→GREEN 전이**다 — M1에서 미수정 코드는 실제로 fail-close하고(telemetry
> 0건), M2 fix 후에야 kept telemetry가 기록된다. 재현 테스트가 production glue 전체
> (real pipeline + real measurer)를 구동하므로 stub-bypass tautology가 아니다.

## §G — Anti-Patterns (회피)

- measurer를 stub하여 happy-path를 통과시키기 — 정확히 sync-auditor WI-3 GAP +
  plan-auditor FAIL 패턴. 재현 테스트는 real measurer를 real 모듈에서 구동.
- `&Applier{...}` 직접 구성 / `runExecuteWith`에 stub evaluator 주입 — 갭을 닫지
  못함. 반드시 `RunExecute` 진입점.
- `measurementRoot`에 cli 패키지 경로 깊이 하드코딩 (Approach B) — 계층 누수, 기각.
- 실제 레포 `.moai/harness/*`에 쓰기 — `t.TempDir()` + ProjectRoot override 필수.
- regression_gate.go(measurer) 수정 — measurer는 올바름, 동결. root만 바로잡음.
- D3 패키지 90% 커버리지 scope creep — 명시적 out of scope (spec.md §D.1).
- fixture를 go.mod 없는 빈 디렉터리로 만들기 — `go test ./...`가 build-fail →
  GREEN 불가. 반드시 real 최소 Go 모듈.

### §G.1 Out of Scope (plan-level)

- STILL-FROZEN 5 파일 수정 (§A.5 diff 0).
- D3 패키지 커버리지 90% 절대 바 달성 (spec.md §D.1 — 별도 follow-up SPEC).
- rolled-back / regression-blocked verdict 변형 테스트 (kept happy-path + RED만).
- measurer 로직 변경 / harness write surface 확장 / propose verb genuine 생성기화.

## §H — Cross-References

- spec.md §B (FROZEN/FIX 구분) / §C (GEARS 요구사항) / §D (Exclusions) / §B.5 (COST NOTE)
- acceptance.md (AC-E2E-001~009 상세 GWT + RED→GREEN)
- 결함: `internal/harness/applier.go:484` `measurementRoot` + `:398`
- 측정기(동결): `internal/harness/regression_gate.go:222` `goMeasurer.Measure`
- 배선점: `internal/cli/harness/execute.go:140`
- CLAUDE.md Rule 4 (Reproduction-First Bug Fixing) + §16 (Multi-File Decomposition)
