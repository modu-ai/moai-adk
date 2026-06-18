---
id: SPEC-V3R6-STATUSLINE-MILLION-BUDGET-001
title: "Statusline memory_test AutoCompactScaling model-env isolation — progress"
version: "0.2.0"
status: completed
created: 2026-06-18
updated: 2026-06-18
author: manager-develop
priority: P2
phase: "v3.0.0"
module: "internal/statusline"
lifecycle: spec-anchored
tags: "statusline, test-fixture, model-env-isolation, debt-cleanup"
---

# Progress — SPEC-V3R6-STATUSLINE-MILLION-BUDGET-001

## §E.1 Plan-phase Audit-Ready Signal

- Plan-auditor Phase 0.5 verdict: **PASS-WITH-DEBT 0.84** (Tier S threshold 0.75; skip-eligible 0.90 아님 → Phase 0.5 재실행 면제 불가지만 이미 실행 완료).
- Dimensions: Clarity 0.85 / Completeness 0.90 / Testability 0.80 / Traceability 1.00.
- Defects: **D1 SHOULD-FIX** (sibling-test line-citation drift `:276+/:350+/:415+/:450+` → 실제 `:268+/:336+/:361+/:423+` in spec.md §A/§G + plan.md §C/§F) — **sync-phase manager-spec re-delegation으로 이월**; D2-D5 MINOR.
- Implementation Kickoff Approval: **APPROVED** (사용자 Option A — scoped commit + 후속 SPEC).
- Technical thesis (Direction B / `t.Setenv("","")` / `memory.go:172-173` empty-skip / production 미변경) plan-auditor가 live code 대상 독립 검증 완료.

## §E — Phase 0.95 Mode Selection

- **Input parameters**: Tier S LEAN; scope = 1 file (`internal/statusline/memory_test.go`); domain count = 1 (statusline test); file language = 100% Go test; concurrency benefit = LOW (coding-heavy, single-file, no inter-file dependency); Agent Teams prereqs = N/A.
- **Decision**: **Mode 5 (sub-agent)** — single sequential `manager-develop` spawn.
- **Justification**: Tier S single-file coding-heavy work → Mode 5 is the default fallback per Anthropic's coding-task parallelism caveat ("most coding tasks involve fewer truly parallelizable tasks than research"). No fan-out benefit; one cohesive test-file edit. Mode 6 (workflow) requires ≥~30 files mechanical transform — not applicable.

## §E.2 Run-phase Evidence

**M1 — `isolateModelEnv(t)` helper applied to `TestCollectMemory_AutoCompactScaling` (Option M1-b).**

Change: `internal/statusline/memory_test.go` **+19 lines** (config package import + `isolateModelEnv(t *testing.T)` helper with SPEC-referencing doc-comment + single `isolateModelEnv(t)` call in the AutoCompactScaling runner body). Production `internal/statusline/memory.go` **UNCHANGED** (AC-SMB-003).

Baseline-attribution: 모든 측정은 동일 트리(HEAD `ce3f23dc3` 기반), 동일 세션, 유일 수정 파일 `internal/statusline/memory_test.go`.

**AC-SMB-001 proof — RED → GREEN, GLM env inline export (AP-SMB-005 anti-false-PASS gate):**

Command: `ANTHROPIC_DEFAULT_OPUS_MODEL=glm-5.2 ANTHROPIC_DEFAULT_SONNET_MODEL=glm-5.2 ANTHROPIC_DEFAULT_HAIKU_MODEL=glm-5.2 go test ./internal/statusline/ -run TestCollectMemory_AutoCompactScaling -v -count=1`

- Pre-fix (RED): 5/5 subtests FAIL — `TokensUsed/TokenBudget = 850000+, want 170000`.
- Post-fix (GREEN):
```
--- PASS: TestCollectMemory_AutoCompactScaling (0.00s)
    --- PASS: .../default_threshold_85%:_83%_used_→_~97%_display (0.00s)
    --- PASS: .../default_threshold_85%:_85%_used_→_100%_display (0.00s)
    --- PASS: .../threshold_90%:_83%_used_→_92%_display (0.00s)
    --- PASS: .../threshold_100%_(no_scaling) (0.00s)
    --- PASS: .../exceeded_threshold_capped_at_100% (0.00s)
PASS — ok github.com/modu-ai/moai-adk/internal/statusline 0.524s
```

**AC-SMB-002** — diff = `isolateModelEnv` helper + 1 call only; NO change to `wantUsed`/`wantBudget`/`wantPctApprox` table values (Direction A rejected). ✓

**AC-SMB-003** — `git diff internal/statusline/memory.go` → **EMPTY** (production byte-identical). ✓

**AC-SMB-004** — sibling override-priority tests under GLM-ambient (uncontaminated):
```
ok github.com/modu-ai/moai-adk/internal/statusline 0.386s
```
(`TestCollectMemory_GLMContextOverride|ExplicitOverride|LLMYAMLOverride|EnvOverridesLLMYAML` 모두 PASS — isolateModelEnv의 `t.Setenv` test-lifetime scoping이 sibling 테스트로 누수 없음.)
Clean-env full suite (`env -u ANTHROPIC_DEFAULT_*`): `ok ... 3.861s` — ALL PASS → 격리가 새로운 breakage를 도입하지 않음 확증.

**AC-SMB-005** — durability: 임시로 `glmContextWindows`에 `"glm-99.9": 2_000_000` 추가 + `ANTHROPIC_DEFAULT_OPUS_MODEL=glm-99.9` export → 5/5 PASS (격리가 모델명이 아닌 env-non-empty에 keyed). **REVERT 확인**: `git diff memory.go` commit 시점 empty. ✓

**Cross-platform (E6)**: `GOOS=windows GOARCH=amd64 go build ./...` → exit 0. ✓

### ⚠️ Gaps (verification-claim-integrity — 명시적으로 PASS로 관측되지 않은 항목)

**OUT-OF-SCOPE 사전존재 실패: baseline `TestCollectMemory` (접미사 없는 원본) under GLM-ambient.** 본 SPEC 변경으로 인한 것이 아님 (fix 전 clean main HEAD `ce3f23dc3`에서 동일하게 재현 — orchestrator 독립 관측 확증). 3개 서브테스트가 `TokenBudget = 1000000, want 200000`로 FAIL:
- `TestCollectMemory/zero_values_-_session_start_state` FAIL
- `TestCollectMemory/used_percentage_takes_priority` FAIL (`TokensUsed = 250000, want 50000` + `TokenBudget = 1000000, want 200000`)
- `TestCollectMemory/current_usage_calculation` FAIL

**근거 원인**: AutoCompactScaling과 동일 — `resolveContextWindowOverride()`가 ambient `ANTHROPIC_DEFAULT_*_MODEL`을 읽어 glm match 시 1M을 반환, baseline 테스트의 200K 기대값을 override. baseline 테스트는 의도적 override-priority 테스트가 아님; SPEC §F의 "sibling tests" 제외가 이것을 override-priority 테스트와 동일시하는 scope-defect.

**처리 (사용자 Option A)**: 후속 Tier S SPEC `SPEC-V3R6-STATUSLINE-BASELINE-ISOLATION-001`로 이월 — 동일 `isolateModelEnv(t)` 1-liner를 `TestCollectMemory` runner body에 적용. GLM-ambient statusline suite는 해당 후속 SPEC이 land할 때까지 **PARTIAL GREEN**.

## §E.3 Run-phase Audit-Ready Signal

- run_status: **audit-ready**
- run_commit_sha: _(M1 commit — subject: `test(statusline): isolate AutoCompactScaling from ambient GLM model env (SPEC-V3R6-STATUSLINE-MILLION-BUDGET-001)`)_
- in-scope 5/5 ACs (AC-SMB-001/002/003/004/005) verbatim evidence로 PASS 검증.
- Production code unchanged (AC-SMB-003 byte-identical).
- baseline `TestCollectMemory` gap 정직하게 기록 + 후속 SPEC queue.
- D1 SHOULD-FIX (line-citation drift) sync-phase manager-spec re-delegation으로 이월.

## §E.4 Sync-phase Audit-Ready Signal

- sync_status: audit-ready
- sync_commit_sha: a7737e9db
- CHANGELOG.md `[Unreleased]` → `### Changed` entry added — AC-HNS-011 퇴차 명시 + baseline gap forward-gap(`BASELINE-ISOLATION-001`).
- 사용자 가시 docs 변경 없음 (test-only; README/docs-site 영향 없음).
- Frontmatter `in-progress → implemented` (spec.md/plan.md/progress.md).
- sync는 orchestrator-direct 수행 (GLM backend — manager-docs/sync-auditor spawn context-limit fallback per `feedback_glm_orchestrator_direct_sync_mx`).
- **D1 SHOULD-FIX (sibling-test line-citation drift) 후속 `BASELINE-ISOLATION-001` plan-phase로 이월** — moving-target rationale: 본 run-phase +19-line edit + 향후 baseline-isolation edit가 모두 `memory_test.go` 라인 번호를 shift시키므로, 모든 statusline test isolation 완료 후 1회 재측정이 효율적.

## §E.5 Mx-phase Audit-Ready Signal

- mx_status: audit-ready
- mx_commit_sha: 130e3efd7
- **4-phase lifecycle complete**: plan (`aaf556119`) → run (`1ae603ff8`) → sync (`a7737e9db`) → Mx (this commit).
- @MX scan: `isolateModelEnv` helper fan_in=1 (< 3 ANCHOR threshold) → 신규 @MX tag 불필요; 상세 doc-comment가 intent + SPEC 참조 전달 (NOTE 역할 충족). Production 코드 미변경이므로 @MX 대상 영역 없음.
- Frontmatter `implemented → completed` (spec.md/plan.md/progress.md).
- Mx orchestrator-direct (GLM; ownership matrix `implemented→completed` = manager-docs OR orchestrator 허용).
- sync-auditor 4-dim 자체평가(GLM fallback): Functionality HIGH(5/5 AC PASS) / Security N/A(test-only) / Craft HIGH(최소 helper) / Consistency HIGH(컨벤션 준수).
- **전방 forward-gaps (별도 SPEC)**: (1) `SPEC-V3R6-STATUSLINE-BASELINE-ISOLATION-001` — baseline `TestCollectMemory` 격리(동일 isolateModelEnv 1-liner); (2) lint-rule SPEC — `ClassifyPRTitle` feat→implemented plan-phase false positive (`StatusGitConsistency` WARNING 근원, `transitions.go:110` case 2).

## HISTORY

- **2026-06-18** (v0.2.0, manager-develop): M1 run-phase complete. `isolateModelEnv` helper (Option M1-b) 적용. in-scope 5/5 ACs PASS. baseline `TestCollectMemory` gap run-phase 검증 중 발견 + 사용자 Option A로 후속 SPEC 이월. plan-auditor Phase 0.5 PASS-WITH-DEBT 0.84 통과.
