---
id: SPEC-HARNESS-CLI-COVERAGE-001
title: "internal/cli/harness 패키지 테스트 커버리지 ≥90% 상향 (test-only)"
version: "0.1.0"
status: implemented
created: 2026-06-15
updated: 2026-06-16
author: manager-spec
priority: P2
phase: "v3.0.0"
module: "internal/cli/harness"
lifecycle: spec-anchored
tags: "harness, coverage, test, tdd, self-harness"
era: V3R6
tier: S
---

# SPEC-HARNESS-CLI-COVERAGE-001 — internal/cli/harness 테스트 커버리지 ≥90% 상향

## §A — 컨텍스트 & 동기 (WHY)

### A.1 Self-Harness 로드맵 D3

본 SPEC은 Self-Harness 로드맵의 **D3 coverage** 항목이다. `internal/cli/harness` 패키지는
`moai harness <verb>` CLI 표면(`execute` / `install` / `propose`)을 호스팅하며, 최근
SPEC-HARNESS-APPLY-EXECUTE-001 / SPEC-HARNESS-EXECUTE-E2E-001을 통해 `Applier.Apply()`의
첫 프로덕션 caller가 배선되었다. 이 패키지는 사용자-facing CLI 경계이므로 회귀가 발생하면
discovery / scripting / CI 통합 / Claude Code hook 호출이 모두 깨진다. 현재 statement
coverage는 **77.9%**로, 핵심 cobra RunE 클로저(`NewInstallCmd` 33.3%, `NewExecuteCmd` 46.2%)와
non-dry-run write 경로가 미커버 상태다.

### A.2 측정 가능한 목표

`internal/cli/harness` statement coverage를 현재 **77.9%**에서 **≥ 90.0%**로 상향한다.
**test-only**: 프로덕션 `.go` non-test 파일은 원칙적으로 무수정(§C HARD 제약 C1). 패키지 경로는
`github.com/modu-ai/moai-adk/internal/cli/harness`이며 대상 소스는 `execute.go` / `install.go` /
`propose.go`다.

### A.3 가치 framing (정직)

이 SPEC은 새 기능을 추가하지 않는다. 가치는 (1) 사용자-facing CLI 경계의 회귀 안전망 강화,
(2) `NewInstallCmd` / `NewExecuteCmd`의 cobra RunE 클로저가 `cmd.Execute()`로 실제 통과되는지를
검증하여 flag 배선 / 절대경로 처리 / error 전파를 non-vacuous하게 고정하는 것이다. 커버리지
ceiling(≥90%)은 **원칙적**이며 accidental하지 않다 — os.Exit 분기 + MarkFlagRequired panic 분기는
표준 Go 테스트로 도달 불가한 documented residual로 명시된다(§D.4).

---

## §B — Ground-truth 커버리지 baseline (측정값, 재도출 금지)

오케스트레이터가 측정한 baseline. 본 SPEC의 모든 커버리지 산술의 SSOT다.

### B.1 함수별 gap (statement coverage <100%)

| 파일:라인 | 함수 | 현재 커버리지 | 목표 |
|-----------|------|--------------:|------|
| `install.go:98` | `NewInstallCmd` | 33.3% | ≥87.5% (reachable ceiling; 3 residual) |
| `install.go:60` | `RunInstall` | 80.0% | 100% |
| `execute.go:298` | `NewExecuteCmd` | 46.2% | ≥69.2% (reachable ceiling; os.Exit body 3 + panic 1 residual) |
| `execute.go:355` | `runExecuteCommand` | 62.5% | ≥90% |
| `execute.go:112` | `RunExecute` | 87.5% | ≥90% |
| `execute.go:155` | `runExecuteWith` | 90.0% | ≥90% (유지/개선) |
| `propose.go:78` | `runPropose` | 78.1% | ≥90% |
| **total** | — | **77.9%** | **≥90.0%** |

### B.2 미커버 블록 (count=0, `file:startLine.col,endLine.col`)

**install.go**
- `61.28-63.3` — `RunInstall` empty-ProjectRoot error 분기
- `74.77-76.3` — `RunInstall` `ScaffoldHarnessDir` 실패 error 분기
- `125.52-155.14` — `NewInstallCmd` RunE 클로저 **전체** (getwd 분기 / filepath.Abs 분기 / opts+RunInstall 호출 / success stdout emit)
- `168.56-169.67` — `MarkFlagRequired` panic 분기 (방어적 도달 불가)

**execute.go**
- `116.17-118.4` + `124.17-126.4` — `RunExecute` error 분기
- `159.17-161.4` — `runExecuteWith` error 분기
- `327.52-335.14` — `NewExecuteCmd` RunE: success 호출 + **`os.Exit(ExitCodeForError(err))` 분기 (329-334)** + return nil
- `344.51-345.67` — `MarkFlagRequired` panic 분기 (방어적 도달 불가)
- `358.65-360.4` — `runExecuteCommand` InheritedFlags lookup 분기
- `366.2-369.12` — `runExecuteCommand` success stdout emit

**propose.go**
- `80-95` — `runPropose` default-flag fallback (InputPath/OutputDir/Limit 미지정 시 기본값 대입)
- `125.78-137.3` — non-dry-run `WriteProposals` 경로 (candidates>0 분기)

### B.3 기존 재사용 가능 테스트 인프라 (재발명 금지)

| 헬퍼/픽스처 | 위치 | 용도 |
|-------------|------|------|
| `writeProposalFixture(t, root, id, targetPath, fieldKey, newValue)` | `execute_test.go:27` | t.TempDir() root에 pending proposal JSON 작성 |
| `writeE2EFixtureModule(t, root) (targetAbs string)` | `execute_e2e_test.go:47` | t.TempDir()를 real 최소 Go 모듈로 구성 + 비-FROZEN 타겟 |
| `countApplyOutcomeKept(t, usageLogPath)` | `execute_e2e_test.go:77` | usage-log.jsonl apply_outcome(kept) 라인 카운트 |
| `newGateInactiveApplier()` | `execute_test.go:120` | gate-inactive Applier 구성 |
| `writeHarnessFile(t, path, body)` | `install_test.go:22` | 부모 디렉터리 포함 파일 작성 |
| `stubEvaluator{decision, err}` | `execute_test.go:108` | 주입 가능한 SafetyEvaluator stub |
| `cobraRequiredAnnotation` const | `execute_test.go:295` | required flag annotation key |

기존 분기 테스트: `TestExecute_MissingProposal/EmptyID/PathTraversal/MalformedProposalJSON/ProposalPathIsDirectory/...`,
`TestNewExecuteCmd_FlagWiring`, `TestRunInstall_*`, `TestPropose_*`, `TestPropose_NoAskUserQuestion`(경계 가드).

---

## §C — GEARS 요구사항 (REQ-HCC-*)

> GEARS notation. `<subject>`는 일반화된 명사(테스트 스위트 / 패키지 / 빌드)를 사용한다.

### C.1 커버리지 목표 (Ubiquitous)

- **REQ-HCC-001**: `internal/cli/harness` 테스트 스위트는 패키지 statement coverage를 **≥ 90.0%**로
  달성한다(`go test -coverprofile` + `go tool cover -func` total 행 기준).
- **REQ-HCC-002**: `internal/cli/harness` 테스트 스위트는 `NewInstallCmd`(install.go:98) 함수
  커버리지를 **≥ 87.5% (reachable ceiling)**로 달성한다. (24 stmts 중 3개는 test-only로 도달 불가한
  documented residual: os.Getwd 실패 129-131 / filepath.Abs 실패 137-139 / MarkFlagRequired panic
  168-169 — §D.4 / EX-2 참조.)
- **REQ-HCC-003**: `internal/cli/harness` 테스트 스위트는 `runExecuteCommand`(execute.go:355) 함수
  커버리지를 **≥ 90%**로 달성한다.
- **REQ-HCC-004**: `internal/cli/harness` 테스트 스위트는 `RunInstall`(install.go:60) 함수
  커버리지를 **100%**로 달성한다(2개 error 분기 + happy path 모두 도달).
- **REQ-HCC-005**: `internal/cli/harness` 테스트 스위트는 `runPropose`(propose.go:78) 함수
  커버리지를 **≥ 90%**로 달성한다(default-flag fallback + non-dry-run write 경로 도달).
- **REQ-HCC-006**: `internal/cli/harness` 테스트 스위트는 `NewExecuteCmd`(execute.go:298) 함수
  커버리지를 **≥ 69.2% (reachable ceiling; os.Exit/panic residual)**로 달성한다. (13 stmts 중 4개는
  test-only로 도달 불가한 documented residual: os.Exit body 329-334 = 3 stmts + MarkFlagRequired
  panic 344-345 = 1 stmt — §D.4 / EX-2 참조. RunE success path(327 진입 + 335 return-nil)는 커버됨.)

### C.2 커버리지 도달 경로 (Event-driven)

- **REQ-HCC-007**: **When** 테스트가 `NewInstallCmd().Execute()`를 `--spec-id` / `--domain` /
  `--project-root <t.TempDir()>` args로 호출하면, RunE 클로저는 `.moai/harness/main.md` + CLAUDE.md
  marker block을 생성하고 success stdout을 emit한다.
- **REQ-HCC-008**: **When** 테스트가 `--project-root` 없이(non-parallel subtest에서 os.Chdir 후)
  `NewInstallCmd().Execute()`를 호출하면, RunE 클로저는 `os.Getwd()` 기본-cwd 분기를 통과한다.
- **REQ-HCC-009**: **When** 테스트가 `runExecuteCommand`를 success 경로로 호출하면, success stdout
  emit 블록(366-369)이 도달된다.
- **REQ-HCC-010**: **When** 테스트가 부모 cobra command에 persistent `--project-root` flag를 세팅하고
  자식 execute에는 own flag 없이 `runExecuteCommand`를 호출하면, InheritedFlags lookup 분기
  (358-360)가 도달된다.
- **REQ-HCC-011**: **When** 테스트가 actionable 패턴을 담은 `tier-promotions.jsonl` 픽스처에 대해
  `--dry-run=false --output-dir <t.TempDir()>`로 `runPropose`를 호출하면, non-dry-run
  `WriteProposals` 경로(125-137)가 도달되고 proposal이 디스크에 작성된다.
- **REQ-HCC-012**: **When** 테스트가 모든 override flag를 생략하고 `runPropose`를 호출하면,
  default-flag fallback 블록(80-95: InputPath/OutputDir/Limit 기본값 대입)이 도달된다.

### C.3 error 분기 도달 (Event-detected)

- **REQ-HCC-013**: **When** `RunInstall`이 빈 `ProjectRoot`로 호출되면, error 분기(61-63)가
  도달되어 `"empty project root"` error를 반환한다.
- **REQ-HCC-014**: **When** `RunInstall`의 `harness` 디렉터리 경로가 기존 regular file과 충돌하면,
  `ScaffoldHarnessDir` 실패 error 분기(74-76)가 도달된다.
- **REQ-HCC-015**: **When** `RunExecute` / `runExecuteWith`의 잔존 error 분기(execute.go 116/124/159)가
  기존 분기 테스트로 도달되지 않으면, 본 SPEC의 테스트가 해당 경로를 도달시킨다.

### C.4 격리 & 경계 (State-driven / Unwanted)

- **REQ-HCC-016**: **While** 어떤 테스트가 실행 중이면, 모든 파일 write는 `t.TempDir()` 내부에서만
  발생한다(프로젝트 root mutation 금지).
- **REQ-HCC-017**: `internal/cli/harness` 테스트 스위트는 `AskUserQuestion` / `mcp__askuser` 호출을
  도입하지 **않는다**(C-HRA-008 경계; `TestPropose_NoAskUserQuestion` 가드 green 유지).
- **REQ-HCC-018**: `internal/cli/harness` 테스트는 `os.Chdir`를 사용하는 subtest를 parallel로
  실행하지 **않으며**, 종료 시 원래 cwd를 복원한다.

### C.5 무회귀 (Unwanted / Ubiquitous)

- **REQ-HCC-019**: SPEC-HARNESS-EXECUTE-E2E-001이 수정한 `--execute` fail-close 동작
  (measurementRoot threading)은 회귀하지 **않는다**(`execute_e2e_test.go` green 유지).
- **REQ-HCC-020**: 본 SPEC의 변경 후 `go test ./...`(패키지 단독이 아닌 전체)는 통과하고,
  `go vet ./...`는 clean하며, `golangci-lint run` baseline은 유지된다.
- **REQ-HCC-021**: **Where** test-only 제약(C1)이 유효한 동안, 프로덕션 non-test `.go` 파일은
  무수정 상태를 유지한다(`git diff --name-only`가 패키지 하위 `*_test.go`만 표시).

### C.6 seam 예외 (Where — 조건부)

- **REQ-HCC-022**: **Where** ≥90% 목표가 test-only로 진정 도달 불가함이 입증된 경우에 한해,
  최소 테스트 seam(예: injectable `exitFunc`)을 프로덕션 코드에 도입할 수 있다. 이 경우
  seam은 명시적 AC(§D AC-HCC-008b)로 노출되고 plan-auditor 정밀 검토 대상으로 flag된다.
  **기본 가정은 seam 불필요** — `NewInstallCmd` 단독으로 gap 대부분이 해소되며, 진정 차단된
  statement는 os.Exit 라인 ~3-4개 + MarkFlagRequired panic 2개뿐이다(§D.4 residual).

---

## §D — 성공 기준 요약

- 패키지 coverage ≥ 90.0% (REQ-HCC-001)
- 함수별 목표 충족 (reachable ceiling): `NewInstallCmd` ≥87.5% / `NewExecuteCmd` ≥69.2% (os.Exit body 3 + panic 1 residual) / `runExecuteCommand` ≥90% / `runPropose` ≥90% / `RunInstall` 100%
- documented acceptable residual 명시(os.Exit 분기 + 2 panic 분기) — accidental ceiling이 아님
- 프로덕션 non-test `.go` 무수정 (seam 예외 시 명시·flag)
- `go test ./...` 전체 green + vet clean + lint baseline 유지
- 기존 경계/회귀 가드(`TestPropose_NoAskUserQuestion`, `execute_e2e_test.go`) green 유지

상세 측정 기준은 `acceptance.md`의 AC-HCC-* 참조.

---

## §E — Exclusions (What NOT to Build)

### §E.1 Out of Scope — 명시적 비범위 항목

본 SPEC이 **명시적으로 다루지 않는** 범위(scope creep 방지):

- **EX-1 — 프로덕션 로직 변경 금지**: `execute.go` / `install.go` / `propose.go`의 알고리즘 동작을
  변경하지 않는다. 본 SPEC은 test-only 커버리지 상향이며, 동작 정정·기능 추가·리팩터링을 포함하지
  않는다. (seam 예외는 REQ-HCC-022의 조건부 최소 변경에 한함.)
- **EX-2 — os.Exit 분기 / MarkFlagRequired panic 커버 강제 금지**: `NewExecuteCmd`의
  `os.Exit(...)` 분기(execute.go:329-334)와 두 `MarkFlagRequired` panic 분기
  (execute.go:344-345, install.go:168-169)를 도달시키기 위해 테스트를 왜곡하지 않는다. 이는
  documented acceptable residual이다(§D.4 / AC-HCC-009).
- **EX-3 — 다른 패키지 커버리지 상향 금지**: `internal/harness`, `internal/harness/proposalgen`,
  `internal/harness/safety` 등 인접 패키지의 커버리지는 본 SPEC 범위 밖이다(D3 후속 분리 항목).
- **EX-4 — 100% 커버리지 추구 금지**: 목표는 ≥90%이며 residual을 강제로 0으로 만들지 않는다.
  미커버 잔존은 §D.4에 원칙적으로 열거된 항목에 한정한다.
- **EX-5 — e2e telemetry / Phase 5 분석 구현 금지**: SPEC-HARNESS-APPLY-EXECUTE-001의 deferred
  D2(e2e telemetry test)와 Phase 5 가치 검증은 본 SPEC과 무관하다.
- **EX-6 — CLI 동작/플래그 추가 금지**: 새 flag, 새 subcommand, 새 출력 포맷을 추가하지 않는다.
  기존 표면을 테스트로 고정할 뿐이다.

---

## §F — Cross-References

- 대상 패키지 규약: `internal/cli/CLAUDE.md`(절대경로 규칙 / 서브에이전트 경계 / exit code 규율)
- 테스트 격리: `CLAUDE.local.md` §6(t.TempDir() + filepath.Abs)
- 선행 SPEC: SPEC-HARNESS-APPLY-EXECUTE-001(execute verb 첫 배선), SPEC-HARNESS-EXECUTE-E2E-001
  (measurementRoot fail-close 정정 — REQ-HCC-019 무회귀 대상)
- 경계 가드 선례: `propose_boundary_test.go`(`TestPropose_NoAskUserQuestion`)
- Go 언어 규칙: `.claude/rules/moai/languages/go.md`(coverage ≥85% / table-driven / t.Parallel)
