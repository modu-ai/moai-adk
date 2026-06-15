---
id: SPEC-HARNESS-APPLY-EXECUTE-001
title: "Harness Apply 실행 동사 — Applier.Apply() 첫 프로덕션 caller로 P2 observer/gate 활성화"
version: "0.1.0"
status: draft
created: 2026-06-15
updated: 2026-06-15
author: manager-spec
priority: P1
phase: "v3.0.0"
module: "internal/cli/harness, internal/harness"
lifecycle: spec-anchored
tags: "harness, apply, activation, observer, regression-gate, autoapply, telemetry"
era: V3R6
---

# SPEC-HARNESS-APPLY-EXECUTE-001 — Harness Apply 실행 동사

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-06-15 | manager-spec | 최초 plan-phase 산출물. Self-Harness 로드맵 P2 "observer/gate activation" 1차 항목. user decision A (opt-in Go execute verb) 확정. |

## §A. Context (배경)

### §A.1 문제 정의

Harness learning subsystem은 완전히 구축된 Apply 파이프라인을 보유한다 — 5-Layer safety pipeline (`safety.Pipeline`), in-Apply 비회귀 게이트 (`NewApplierWithRegressionGate`), apply-outcome 캡처 (`WithOutcomeObserver`), M6 lineage manifest. 그러나 이 파이프라인의 단일 진입점 `Applier.Apply()`는 **프로덕션 caller가 0개**다 (grep 검증: 정의/주석/테스트만 매칭). 실제 harness apply는 전부 skill-workflow Edit 경로(Path S — Claude Edit tool)가 수행하며 Go 파이프라인을 절대 경유하지 않는다 (dual-apply-path 아키텍처).

그 결과:
- regression gate가 단 한 번도 active 상태로 실행된 적이 없다.
- apply-outcome telemetry (`usage-log.jsonl`의 `apply_outcome` line)가 단 1건도 생성된 적이 없다.
- `safety.NewPipeline`의 프로덕션 caller도 0개다 (L1~L5가 실제 proposal에 대해 한 번도 평가된 적 없음).

이는 Phase 5 (downstream 학습 분석: failure-signature clustering, canary-effectiveness)가 의존할 **첫 outcome 데이터를 생성할 길이 없음**을 의미한다.

### §A.2 정직한 가치 framing (HONEST FRAMING — 과대광고 금지)

[HARD] 본 SPEC의 가치를 정직하게 기술한다. 현재 harness write surface는 markdown-only FROZEN allowlist이므로, regression gate의 측정 delta는 사실상 항상 Δ=0이다 (`go test`/`go vet` 결과가 markdown frontmatter 수정 전후로 동일). 즉:

- 본 SPEC은 회귀를 **"방지"하지 않는다**. markdown-only 표면에서 gate는 always-pass다.
- 본 SPEC의 **실질 가치는 단 하나**: `Applier.Apply()`의 첫 프로덕션 caller가 되어, **첫 apply-outcome telemetry를 생성**하는 것. 이 telemetry가 Phase 5 분석의 입력 substrate가 된다.
- 부차 가치: regression gate가 dormant defense-in-depth 안전망으로 실제 배선됨 — FROZEN allowlist가 향후 확대되거나 applier 결함이 allowlist 밖 Go/template 코드를 쓰면 그때 비로소 fire한다.

regression_gate.go / outcome.go의 기존 §A.2 HONEST FRAMING 주석과 정합한다.

### §A.3 user decision A (확정)

사용자는 사전 세션에서 **opt-in Go execute verb를 통한 점진적 활성화**를 결정했다 (full switchover 아님). Path S (skill Edit)는 default로 유지된다. Go execute 경로는 **opt-in 전용** — `--execute` 의도가 명시될 때만 `Applier.Apply()`를 호출한다.

### §A.4 predecessor 계보

- SPEC-HARNESS-LOOP-CLOSURE-001 (P0, completed) — lineage + human-gate
- SPEC-HARNESS-REGRESSION-GATE-001 (P1, completed) — 비회귀 게이트 scaffold
- SPEC-HARNESS-OUTCOME-CAPTURE-001 (P2 1차, completed) — outcome 캡처 enabler (dormant scaffold, 프로덕션 caller 0)
- SPEC-HARNESS-OUTCOME-ERRJOIN-001 (P2 후속, completed) — rolled-back branch errors.Join 신호 보존
- **SPEC-HARNESS-APPLY-EXECUTE-001 (본 SPEC, P2 activation 1차)** — 위 모든 scaffold를 처음으로 live 경로에 배선

## §B. autoApply Contract (핵심 설계 결정)

### §B.1 L5 재pending 회피 메커니즘

`safety.Pipeline.Evaluate()`는 L5 (Human Oversight)에서 다음과 같이 동작한다 (`safety/pipeline.go:147`):
- `autoApply == false` → **항상** `DecisionPendingApproval` 반환
- `autoApply == true` → `DecisionApproved` 반환

`Applier.Apply()`가 `DecisionPendingApproval`을 받으면 `*ApplyPendingError`를 반환하고 **파일 수정 단계에 절대 도달하지 못한다**. 따라서 execute verb가 실제로 apply에 도달하려면 Evaluate()가 PendingApproval을 반환하면 안 된다.

오케스트레이터(MoAI)는 C-HRA-008 경계에서 **이미 사용자의 L5 승인을 획득한 상태**에서 execute verb를 호출한다. 그러므로 verb는 Pipeline을 `PipelineConfig{AutoApply: true}`로 생성한다 — L1~L4 (Frozen/Canary/Contradiction/RateLimit)는 **여전히 강제**되고 L5 (Human Oversight)만 CLI 레벨에서 auto-approve된다.

이것이 **autoApply contract**다: `--execute` 의도는 "오케스트레이터가 이미 사용자 L5 동의를 수집했다; 재pending하지 말라"는 단언이다.

### §B.2 [HARD] FROZEN 불변식 — harness.yaml 영속값 불변

[HARD][FROZEN — DO-NOT-MODIFY] `harness.yaml`의 영속된 `auto_apply: false`는 디스크에서 **절대 변경되지 않는다**. `AutoApply: true`는 `--execute`가 존재할 때에만 설정되는 **in-memory `PipelineConfig` 값**이다. config 파일을 mutate하면 안 된다.

근거: harness.yaml의 `auto_apply: false`는 "기본적으로 사람 승인 필요"라는 시스템 default 안전 정책이다. execute verb가 이를 디스크에서 true로 바꾸면, 이후 모든 harness 동작이 사람 승인 없이 auto-apply되어 안전 default가 무너진다. in-memory override는 단 1회 호출에만 국한되어 안전 default를 보존한다.

## §C. EARS/GEARS Requirements

### §C.1 Execute verb 표면 (Ubiquitous + Where)

- **REQ-AEX-001** (Ubiquitous): The execute verb shall live in package `internal/cli/harness` (sub-directory) as a new verb-factory file `execute.go`, mirroring `propose.go` / `install.go`, so the contained-directory boundary guard `TestPropose_NoAskUserQuestion` covers it.
- **REQ-AEX-002** (Ubiquitous): The execute verb shall be registered into `newHarnessRouterCmd()` (`harness_route.go`) via an exported factory `NewExecuteCmd()`.
- **REQ-AEX-003** (Where capability gate): Where the `--execute` flag is present, the existing `apply` verb shall delegate file-application to the Go execute path (`Applier.Apply()`); where `--execute` is absent, the `apply` verb shall preserve its existing payload-only behavior (read oldest pending proposal, print JSON, no `Applier.Apply()` call).
- **REQ-AEX-004** (Event-driven): When the execute path is invoked with a proposal ID, the verb shall load the proposal from `.moai/harness/proposals/<id>.json` into a `harness.Proposal` value.

### §C.2 autoApply contract (State-driven + Unwanted)

- **REQ-AEX-005** (Ubiquitous): The execute path shall construct the safety Pipeline with `safety.PipelineConfig{AutoApply: true}` so L1–L4 remain enforced and L5 is auto-approved at the CLI level.
- **REQ-AEX-006** (Unwanted behavior — FROZEN): The execute path shall not mutate the persisted `auto_apply: false` value in `harness.yaml` on disk.
- **REQ-AEX-007** (State-driven — invariant): While the execute path runs under `AutoApply: true`, `Applier.Apply()` shall never return `*ApplyPendingError`; its occurrence is an invariant violation.

### §C.3 Apply 파이프라인 배선 (Ubiquitous)

- **REQ-AEX-008** (Ubiquitous): The execute path shall construct the Applier via `harness.NewApplierWithRegressionGate(manifestPath, baselinePath).WithOutcomeObserver(harness.NewObserver(usageLogPath))`, wiring the regression gate AND the outcome observer.
- **REQ-AEX-009** (Ubiquitous): The execute path shall pass `snapshotBase = <root>/.moai/harness/learning-history/snapshots`, `manifestPath = <root>/.moai/harness/learning-history/manifest.jsonl`, `baselinePath = <root>/.moai/harness/measurements-baseline.yaml`, and `usageLogPath = <root>/.moai/harness/usage-log.jsonl`, resolved relative to the project root.
- **REQ-AEX-010** (Event-driven): When `Applier.Apply()` returns nil on the first gated apply, the execute path shall have produced the FIRST `apply_outcome` telemetry line in `usage-log.jsonl`.

### §C.4 sessions sourcing for L2 Canary (Where + Event-detected)

- **REQ-AEX-011** (Where): Where no recent-session metrics are available for the first execute run, the execute path shall pass a nil/empty `[]harness.Session` slice to `Applier.Apply()`; `baselineScore` over an empty slice returns 0, and `defaultProjectedScorer` returns baseline+0.02 for a meaningful proposal, so the canary drop is 0 and L2 does not reject (nil-safe per `canary.go`).

### §C.5 Error surface → exit code mapping (Event-detected)

- **REQ-AEX-012** (Event-detected): When the proposal ID resolves to no file OR `Applier.Apply()` returns a rejection error (L1–L4 block), the verb shall exit 1 (user error) with a structured stderr diagnostic.
- **REQ-AEX-013** (Event-detected): When `Applier.Apply()` returns `*ApplyRegressionError`, the verb shall exit 1 (gate rolled back — user-actionable) with a stderr diagnostic naming the regressed dimensions; for the markdown-only surface this branch is not expected on a first run (baseline adopted) but must be handled.
- **REQ-AEX-014** (Event-detected — invariant violation): When `Applier.Apply()` returns `*ApplyPendingError` under `AutoApply: true`, the verb shall exit 2 (system error / invariant violation) with a diagnostic stating the autoApply contract was violated.
- **REQ-AEX-015** (Event-detected): When a measurement-exec failure (build error / timeout) is surfaced as a wrapped error, the verb shall exit 2 (system error).

### §C.6 C-HRA-008 subagent boundary (Unwanted)

- **REQ-AEX-016** (Unwanted behavior): No source file in `internal/cli/harness/` (including the new `execute.go`) shall invoke `AskUserQuestion` or `mcp__askuser__*`. The orchestrator owns user interaction; the verb takes positional/flag inputs and emits structured errors.

## §D. Exclusions (What NOT to Build)

[HARD] 본 SPEC의 명시적 비범위.

### Out of Scope — 명시적 비범위 항목

- **Path S 전환 — Out of Scope**: skill-workflow Edit 경로는 default로 그대로 유지. dual-apply-path 아키텍처 결정(어느 경로를 canonical로 할지)은 별도 후속 SPEC의 책임.
- **harness.yaml 디스크 mutation — Out of Scope**: `harness.yaml`의 `auto_apply` 디스크 값 변경 안 함 (§B.2 FROZEN 불변식).
- **FROZEN scaffold 알고리즘 수정 — Out of Scope**: regression gate / outcome / pipeline / lineage / canary 알고리즘 수정 안 함. 본 SPEC은 thin caller(배선)이며 기존 scaffold를 호출만 한다 — `applier.go`, `regression_gate.go`, `outcome.go`, `pipeline.go`, `canary.go`, `lineage.go`의 로직 변경 금지.
- **Phase 5 downstream consumer — Out of Scope**: failure-signature clustering, canary-effectiveness, held-out 검증, scorer-loop은 전부 downstream out-of-scope.
- **회귀 "방지" 주장 — Out of Scope**: markdown-only 표면에서 gate는 always-pass (§A.2 HONEST FRAMING). CHANGELOG/문서에 "prevents regressions" 류 과대광고 금지; "generates first apply-outcome telemetry"로만 framing.
- **auto-discovery / 자동 트리거 — Out of Scope**: execute는 명시적 opt-in 호출 전용. 자동 스케줄링/배치 적용 없음.
- **propose verb 수정 — Out of Scope**: proposal 생성은 SPEC-V3R6-HARNESS-PROPOSAL-GEN-001 소관. 본 SPEC은 이미 존재하는 proposal을 소비만 한다.
- **scorer/tier/classifier 수정 — Out of Scope**: `defaultProjectedScorer`, `ClassifyTier` 등 무수정.

## §E. MX Tag Targets

High fan_in danger zone (mx-tag-protocol.md 기준):
- `Applier.Apply` — 본 SPEC이 **첫 프로덕션 caller**가 되어 fan_in이 변동 → 기존 `@MX:ANCHOR` 유지 + caller 목록 갱신 대상 (NEVER auto-delete ANCHOR).
- `safety.NewPipeline` — 본 SPEC이 첫 프로덕션 caller → `@MX:ANCHOR` 후보 (fan_in이 test-only에서 production으로 전환).
- 신규 `runExecute` / `NewExecuteCmd` — autoApply contract 분기(§B.1)가 비자명 → `@MX:NOTE` (autoApply=true는 L5 재pending 회피 의도, harness.yaml 디스크 불변) + `@MX:WARN`(AutoApply=true가 in-memory 전용임을 강조, 디스크 mutate 금지) 후보. `@MX:WARN`은 `@MX:REASON` 필수.

## §F. Cross-References

- `.claude/rules/moai/workflow/mx-tag-protocol.md` — MX 태그 규약
- `.claude/rules/moai/quality/boundary-verification.md` — CLI↔harness pkg 경계 검증
- `internal/cli/CLAUDE.md` — exit code discipline (0/1/2), C-HRA-008 boundary, 절대경로 규칙
- `internal/harness/applier.go` (Apply seam), `safety/pipeline.go` (AutoApply L5), `outcome.go` (RecordOutcome), `regression_gate.go` (baseline const), `canary.go` (nil-safe sessions)
