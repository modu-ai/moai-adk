---
id: SPEC-INTERNAL-TEST-002
title: "internal/ 사전 부채 3종 청소 — TEST-001 DEBT-TEST-001/002/003 후속 (ARCH-001 M0 prerequisite)"
version: "0.1.0"
status: draft
created: 2026-07-09
updated: 2026-07-09
author: manager-spec
priority: P1
phase: "v3.0.0"
module: "internal/"
lifecycle: spec-anchored
tags: "test, ci, baseline, golden, coverage, internal, debt-cleanup"
tier: M
era: V3R6
depends_on: [SPEC-INTERNAL-TEST-001]
related_specs: [SPEC-INTERNAL-ARCH-001, SPEC-INTERNAL-TEST-001, SPEC-INTERNAL-PERF-001]
---

# SPEC-INTERNAL-TEST-002 — internal/ 사전 부채 3종 청소

## HISTORY

| 버전 | 날짜 | 변경 내용 | 작성자 |
|------|------|-----------|--------|
| 0.1.0 | 2026-07-09 | 최초 작성 (draft). SPEC-INTERNAL-TEST-001 §E.4 `follow_up_spec_scope_summary`가 명명한 debt-owner SPEC. 3개 부채(DEBT-TEST-001/002/003)를 GEARS 요구사항으로 변환. 저작 시점 HEAD에서 3 부채의 root cause를 독립 감사(verification-claim-integrity §1.1 surface 3 준수). 저작 동기: SPEC-INTERNAL-ARCH-001 plan-audit iter-1 FAIL(0.76, Tier L 임계 0.85)의 BLOCKING 결함 D1이 본 SPEC M1의 선행 완료를 요구. | manager-spec |

## §A 배경 및 목표 (Why)

SPEC-INTERNAL-TEST-001은 3개 residual debt를 남기며 PASS-WITH-DEBT로 3-phase close 되었다 (`progress.md §E.4` `residual_debt` 블록 및 `follow_up_spec_scope_summary` 참조). 본 SPEC은 그 3개 부채의 정식 debt-owner로서, 각 부채의 root cause를 기계적 도구로 재검증한 뒤 GEARS 요구사항으로 변환한다.

### 저작 시점 독자적 root-cause 재검증 (2026-07-09, verification-claim-integrity §1.1 surface 3 준수)

3개 부채 각각에 대해 전용 도구(`go test -v`, `go test -coverprofile`, `git show HEAD:<path>`)를 사용하여 현재 HEAD 베이스라인에서 root cause를 재측정했다. TEST-001 §E.4 기술 대비 **2개의 주요 정정사항**이 발견되었으며, 이를 투명하게 문서화한다:

| ID | TEST-001 §E.4 기술 (2026-07-08) | 2026-07-09 재검증 결과 | 정정 여부 |
|----|--------------------------------|----------------------|----------|
| DEBT-TEST-001 (부채 a) | 6× internal/cli TestDoctor/TestStatus TUI 렌더링 FAIL (Current_Light/Dark/NoColor). SIGSEGV 수정으로 노출. | **(a-stale) stale-golden 확정**. `got:` vs `want:` 단일 행 차분 결과 렌더러 코드는 정상, 6개 golden testdata 파일이 `v3.0.0-rc6`에 고정된 채 `pkg/version/version.go:8`이 `v3.0.0-rc8`로 올라간 것이 유일 원인. 문자 1자(`6`→`8`) 차이. | 원인 재확인 (stale-golden 가설 확인) |
| DEBT-TEST-002 (부채 b) | internal/statusline `TestBuild_WritesContextUsageWithSessionID` env-flake FAIL. `context_window_size=1000000, want 256000`. deterministic flake, 1회 재실행으로 PASS하지 않음. | **(b-already-resolved) 이미 해결됨**. commit `794bb4f84` (2026-07-09 00:41:12 +0900, "test(statusline): isolate context-usage test from ambient GLM env")가 TEST-001 sync(`80dea9684`, 2026-07-08 03:37:36 +0900) 이후에 landed. `builder_test.go:1815-1818`의 4개 `t.Setenv` 호출이 ambient GLM env를 차단. `-count=10` 연속 실행 10/10 PASS, GLM env 주입 상태에서도 3/3 PASS. | **정정 요약**: TEST-001 §E.4 기술 시점에는 deterministic FAIL이었으나 현재 HEAD에서는 이미 해결됨. 본 SPEC M2는 verify-only(evidence 재생성만, 코드 변경 없음). |
| DEBT-TEST-003 (부채 c) | internal/constitution package coverage 67.5% < 85%. pipeline.go 8함수 0% (integration-level) + human_oversight.go 일부. | **(c-confirmed) 확인**. `go tool cover -func` 측정 결과 pipeline.go 8함수(NewPipeline/Execute/createLogEntry/applyAmendment/acquireLock/releaseLock/updateSourceFile/updateRegistryClause) 전부 0% + human_oversight.go `printDiff` 0% / `Approve` 14.3%. package 67.5% → 85% 상승 필요 integration test 추가. | 원인 재확인 |

### ARCH-001 선행 관계 (plan→run gate)

SPEC-INTERNAL-ARCH-001 plan-audit iter-1이 **BLOCKING 결함 D1**(score 0.76, Tier L 임계 0.85 미달)로 FAIL했다. D1의 내용: "ARCH-001 REQ-ARCH-002가 리팩터링 대상으로 삼는 `internal/cli`(update.go/hook.go split)이 **6개의 사전 RED 테스트**를 가지고 있어 M0 게이트(`go test ./... exit 0`, AC-ARCH-001)가 통과 불가" — 이 6개가 본 SPEC M1의 청소 대상이다.

**ARCH-001 re-entry는 본 SPEC M1 완료 + 후속 web-i18n SPEC(TEST-003 또는 특화 SPEC) 완료에 의존한다.** M1은 necessary not sufficient — ARCH-001 AC-ARCH-001(`go test ./... exit 0`, whole-repo headline gate)이 잔여 2개 `internal/web` i18n 실패(`TestDataI18nKeysSubsetOfDictionary`, `TestI18nKeySetParity`)로 인해 본 SPEC만으로는 달성되지 않는다. 본 SPEC M1이 6개 internal/cli FAIL을 청소한 뒤에도 whole-repo `go test ./...`는 여전히 2개 internal/web i18n FAIL로 non-green이며, ARCH-001 re-entry는 이 2건을 청소할 후속 web-i18n SPEC의 완료를 추가로 요구한다. 이 선행 관계는 ARCH-001 SPEC 산출물이 아닌 **본 SPEC §A가 SSOT**다 (ARCH-001 산출물은 본 SPEC이 소유하지 않음 — `status: draft`, blocked on TEST-002 M1 + web-i18n SPEC per plan-audit iter-1 BLOCKING D1).

### Epic 확장 (5/5 → 6/6)

본 SPEC 신규 편성 + 후속 web-i18n SPEC으로 internal/ 독립 감사 Epic이 5/5에서 6/6으로 확장된다: TEST-001 → SECURITY-001 → PERF-001 → ARCH-001(`status: draft`, 보류) → **TEST-002(본 SPEC)** → **web-i18n SPEC(TEST-003 또는 특화 SPEC, 미저작 future SPEC)** → ARCH-001 재진입. Epic scope SSOT는 본 SPEC plan.md §A.1.

### Headline gate 범위 (TEST-001과의 구분)

TEST-001은 headline REQ(`AC-TEST-001`: `go test ./internal/... exit 0`)을 소유하고 PASS-WITH-DEBT로 재해석했다. **본 SPEC은 headline gate를 소유하지 않는다.** 이유: 저작 시점 HEAD 베이스라인 `go test ./internal/... -count=1` 관측 결과 **8 FAIL = 6(debt a, 본 SPEC scope) + 2(internal/web i18n, Out of Scope)**. 0(debt b 이미 해결, commit `794bb4f84`) + 0(debt c는 coverage gap으로 test FAIL 없음 — 이전 "2(debt c 영향)" 항은 phantom term으로 제거). 잔여 2개 `internal/web` i18n 실패(`TestDataI18nKeysSubsetOfDictionary`, `TestI18nKeySetParity`)는 TEST-001 §E.4가 명명하지 않은 부채로 본 SPEC scope 밖이며, 후속 web-i18n SPEC(TEST-003 또는 특화 SPEC) 소관이다. 따라서 본 SPEC은 3개 per-debt REQ(REQ-TEST-007/008/009)와 3개 per-debt AC(AC-TEST-007/008/009)만 소유하며, headline `go test ./internal/... exit 0` 및 whole-repo `go test ./... exit 0`(ARCH-001 AC-ARCH-001)은 후속 web-i18n SPEC + Epic 차원의 관심사로 남긴다.

## §B GEARS 요구사항

### REQ-TEST-007 — debt (a): internal/cli golden testdata 재생성 (ARCH-001 M0 prerequisite, Event-driven)

**When** `go test ./internal/cli/ -run 'TestDoctor_Current_Light|TestDoctor_Current_Dark|TestDoctor_NoColor|TestStatus_Current_Light|TestStatus_Current_Dark|TestStatus_NoColor' -count=1` is executed at repository HEAD, all 6 tests **shall** PASS and the 6 golden testdata files (`internal/cli/testdata/{doctor,status}-{light,dark,nocolor}.golden`) **shall** carry the current `pkg/version/version.go` Version string (replacing the stale `v3.0.0-rc6` fixture with the current `v3.0.0-rc8` value). **Where** the regeneration mechanism uses the existing `UPDATE_GOLDEN=1` env flag already wired at `doctor_golden_test.go:14` and `status_golden_test.go:11`, the remediation **shall** invoke `UPDATE_GOLDEN=1 go test ./internal/cli/ -run 'TestDoctor_Current_|TestStatus_Current_' -count=1` to regenerate all 6 fixtures in one pass (no manual byte editing required, no test framework change required).

### REQ-TEST-008 — debt (b): statusline env-isolation already-resolved 상태 검증 (verify-only, Event-driven + Where)

**When** `go test -run TestBuild_WritesContextUsageWithSessionID -count=10 ./internal/statusline/` is executed, the test **shall** PASS 10 consecutive runs at repository HEAD. **Where** commit `794bb4f84` (2026-07-09, "test(statusline): isolate context-usage test from ambient GLM env") already applied the fix — 4 `t.Setenv` calls at `internal/statusline/builder_test.go:1815-1818` isolating `MOAI_STATUSLINE_CONTEXT_SIZE` / `ANTHROPIC_DEFAULT_OPUS_MODEL` / `ANTHROPIC_DEFAULT_SONNET_MODEL` / `ANTHROPIC_DEFAULT_HAIKU_MODEL` from ambient dev-machine GLM env — this requirement's verification is **evidence-regeneration only**, and **no production or test code change shall be made** under this REQ. The test's design (synthetic 256K window → expect 256000 write) is correct as-is; the GLM-1M override is the ambient-env hazard the isolation repels.

### REQ-TEST-009 — debt (c): internal/constitution pipeline.go integration coverage 67.5% → 85% (Ubiquitous)

The `internal/constitution` package **shall** gain executable integration tests covering the 8 `pipeline.go` functions currently at 0% statement coverage — `NewPipeline`, `Execute`, `createLogEntry`, `applyAmendment`, `acquireLock`, `releaseLock`, `updateSourceFile`, `updateRegistryClause` — plus `human_oversight.go::printDiff` (0%) and `human_oversight.go::Approve` (14.3%), raising package statement coverage from 67.5% to ≥ 85%. **While** the pipeline functions are integration-level (filesystem lock acquisition, source-file amendment, registry-clause update), the tests **shall** use `t.TempDir()` isolation per CLAUDE.local.md §6 and **shall not** mutate the real `.claude/rules/moai/core/` registry.

## §C 제약사항 (Constraints)

- **C-1 (unwanted)**: REQ-TEST-007 remediation **shall not** alter renderer code in `internal/cli/doctor.go`, `internal/cli/status.go`, or any production path — the 6 FAILs are stale-fixture failures, NOT renderer bugs (root cause confirmed by byte-level `got:` vs `want:` diff showing only the version-string character difference). Test-fixture-only remediation.
- **C-2 (unwanted)**: REQ-TEST-008 **shall not** modify `internal/statusline/builder_test.go`, `internal/statusline/renderer.go`, or any statusline file — the fix is already in HEAD via commit `794bb4f84`. This REQ owns only evidence regeneration (test-run logs cited in acceptance.md).
- **C-3 (unwanted)**: REQ-TEST-009 remediation **shall not** modify `internal/constitution/pipeline.go` or `internal/constitution/human_oversight.go` production code — tests-only. 실제 production 결함 발견 시 silent fix 대신 blocker report 반환 (TEST-001 C-3 패턴 동일 적용).
- **C-4 (working-tree preservation)**: 본 SPEC run-phase는 현재 working tree의 18 modified + 5 untracked 파일(IGGDA Path-B retire 잔류 — orchestration-mode-selection.md, template iggda-*.sh deletion, hook/session_start.go, statusline/renderer.go 💾→♻️ emoji 변경, moai-easy.md, credentials.yml)을 **절대 touch/commit/stash하지 않는다**. 특히 `internal/statusline/renderer.go`와 `internal/statusline/cache_hit_test.go`는 working tree에서 이미 수정 상태이나 이는 본 SPEC scope 밖이며 REQ-TEST-008은 HEAD 베이스라인에서만 검증한다.
- **C-5**: REQ-TEST-009 테스트 작성 시 filesystem interaction은 전부 `t.TempDir()` isolation 의무 (CLAUDE.local.md §6).
- **C-6 (pathspec discipline)**: 본 SPEC의 모든 commit은 specific pathspec(`git add .moai/specs/SPEC-INTERNAL-TEST-002/ internal/cli/testdata/ internal/constitution/*_test.go` 등)만 사용 — `git add -A` / `git add .` 금지. 다른 Claude session(1d3c155b quiescent)과의 race 회피.
- **C-7**: 커버리지 플로어는 CLAUDE.local.md 기준 준수 — internal/constitution package 85% 최소.

## Out of Scope (제외 범위)

다음 항목은 본 SPEC의 out of scope로 명시적으로 제외한다.

### Out of Scope — internal/web i18n 2개 신규 실패

- `go test ./internal/... -count=1` 관측 결과 `internal/web` 패키지의 `TestDataI18nKeysSubsetOfDictionary` 및 `TestI18nKeySetParity` 2개 FAIL이 HEAD 베이스라인에 존재한다. 이는 TEST-001 §E.4가 명명하지 않은 부채(2026-07-09 재검증 시 신규 발견)로, **본 SPEC scope 밖**이다. 후속 SPEC(SPEC-INTERNAL-TEST-003 또는 web-i18n 특화 SPEC) 소관.

### Out of Scope — headline `go test ./internal/... exit 0`

- 본 SPEC은 3개 per-debt REQ(REQ-TEST-007/008/009)와 3개 per-debt AC(AC-TEST-007/008/009)만 소유한다. headline gate `go test ./internal/... exit 0`는 위 2개 `internal/web` i18n 실패로 인해 본 SPEC이 머지되어도 달성되지 않으며, 이를 본 SPEC이 claim하지 않는다 (과거 TEST-001의 headline reinterpretation 패턴 반복 회피). Epic 차원의 관심사.

### Out of Scope — SPEC-INTERNAL-ARCH-001 산출물 수정

- ARCH-001은 `status: draft` (blocked on TEST-002 M1 + web-i18n SPEC per plan-audit iter-1 BLOCKING D1) 상태의 SPEC이다 — "paused"/"일시정지"는 8-value status enum에 존재하지 않는 비정형 용어이므로 사용하지 않는다. 본 SPEC은 ARCH-001의 spec/plan/acceptance/progress를 **절대 수정하지 않는다**. 본 SPEC §A가 ARCH-001 선행 관계의 SSOT다.

### Out of Scope — internal/statusline renderer.go / cache_hit_test.go working-tree 변경

- working tree의 `internal/statusline/renderer.go`(💾→♻️ emoji 변경) 및 `internal/statusline/cache_hit_test.go`(동일 emoji 변경)는 IGGDA Path-B retire 코호트의 unrelated work이며 본 SPEC scope 밖이다. REQ-TEST-008은 HEAD 베이스라인에서만 검증하며 dirty-tree 변경사항에 의존하거나 이를 머지하지 않는다.

### Out of Scope — internal/cli testdata golden 파일 양식(format) 재설계

- REQ-TEST-007은 stale version string(`rc6`→`rc8`) 갱신만 수행한다. golden 파일의 레이아웃/폭/이모지/구조 변경은 scope 밖이다 — renderer가 정상이므로 format 변경 이유가 없다.

### Out of Scope — CI workflow / `moai init` template 변경

- `.github/workflows/**` 및 template 트리(`internal/template/templates/**`) 편집 없음. 본 SPEC은 테스트 스위트 + golden testdata + 테스트 코드만 다룬다.

## 참조 (Cross-References)

- `pkg/version/version.go:8` — `Version = "v3.0.0-rc8"` (REQ-TEST-007 갱신 대상 버전 문자열 소스)
- `internal/cli/testdata/{doctor,status}-{light,dark,nocolor}.golden` (6개, REQ-TEST-007 갱신 대상 — stale `v3.0.0-rc6`)
- `internal/cli/doctor_golden_test.go:13-14, 21` / `internal/cli/status_golden_test.go:10-11, 95-102, 128-130` — `UPDATE_GOLDEN=1` regen mechanism (이미 구현됨, REQ-TEST-007 활용)
- `internal/statusline/builder_test.go:1802-1859` — `TestBuild_WritesContextUsageWithSessionID` (REQ-TEST-008 검증 대상, 이미 HEAD에서 PASS)
- `internal/statusline/memory.go:18-30, 128-172` — `glmContextWindows` + `resolveContextWindowOverride` (GLM 1M override 메커니즘; REQ-TEST-008이 검증하는 isolation이 차단하는 대상)
- commit `794bb4f84` "test(statusline): isolate context-usage test from ambient GLM env" (2026-07-09) — REQ-TEST-008 already-resolved 근거
- `internal/constitution/pipeline.go:29-264` — 8 functions at 0% coverage (REQ-TEST-009 대상)
- `internal/constitution/human_oversight.go:30, 62` — `Approve` 14.3% / `printDiff` 0% (REQ-TEST-009 대상)
- `.moai/specs/SPEC-INTERNAL-TEST-001/progress.md §E.4` `residual_debt` + `follow_up_spec_scope_summary` — 본 SPEC의 3개 부채 원본 명명
- `.moai/specs/SPEC-INTERNAL-ARCH-001/spec.md` REQ-ARCH-002 + plan.md AC-ARCH-001(M0 gate) — 본 SPEC M1을 선행 조건으로 참조하는 요구사항
- CLAUDE.local.md §2 (Template-First Rule — 본 SPEC은 template 미러 변경 없음), §6 (Test Isolation `t.TempDir()`, Coverage Targets 85%)
- verification-claim-integrity §1.1 surface 3 — 본 SPEC §A의 root-cause 재검증이 준수하는 "결함/부채 단언은 전용 도구로 검증" 규칙
