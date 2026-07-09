---
id: SPEC-INTERNAL-TEST-002
version: "0.1.0"
status: in-progress
created: 2026-07-09
updated: 2026-07-09
---

# SPEC-INTERNAL-TEST-002 — progress

## §E.1 Plan-phase Audit-Ready Signal

- 2026-07-09 (iter-1): plan-phase 산출물 4종 저작 완료 (manager-spec, `status: draft`). 3개 부채(TEST-001 §E.4 DEBT-TEST-001/002/003) root cause 재검증. plan-auditor iter-1 verdict: **FAIL (0.82, BLOCKING D1)**.
- 2026-07-09 (iter-2): D1-D5 defects 해소 (orchestrator independently verified D1+D2; manager-spec resolved D1-D5 per **Option A** — TEST-002 keeps per-debt scope, web-i18n 2건 deferred to future SPEC TEST-003, NO M4, NO scope expansion):
  - **D1 (BLOCKING) 해소**: spec.md §A "ARCH-001 선행 관계" internal contradiction 제거. ARCH-001 re-entry = 본 SPEC M1 **necessary not sufficient** + 후속 web-i18n SPEC(TEST-003) 완료. ARCH-001 AC-ARCH-001(`go test ./... exit 0`, whole-repo headline)가 잔여 2개 internal/web i18n 실패로 본 SPEC만으로 미달성. Epic 확장 5/5 → 6/6.
  - **D2 (SHOULD-FIX) 해소**: "13 FAIL" phantom arithmetic 정정. 실측 `go test ./internal/... -count=1` = **8 FAIL = 6(debt a, in scope) + 2(internal/web i18n, Out of Scope)**. debt b = 0(794bb4f84 resolved), debt c = 0(coverage gap, test FAIL 없음). Phantom "2(debt c 영향)" 제거.
  - **D3 (SHOULD-FIX) 해소**: acceptance.md AC-TEST-007 Then-1 regen-blindness 폐쇄. `grep -l "v3.0.0-rc8" ... | wc -l → 6` 기계 검증 + byte-level sanity diff 추가.
  - **D4 (MINOR) 해소**: "paused" → formal status `draft` (spec.md 2곳, plan.md 2곳). 8-value status enum 준수.
  - **D5 (MINOR) 해소**: acceptance.md AC-TEST-008 Then-3 commit-scope imprecision 정정. `HEAD~..HEAD` → `<M2-commit>^..<M2-commit>` (progress.md §E.2 SHA) 또는 run-phase range form.
- 3개 부채 재검증 결과 (iter-1에서 확정, iter-2 변경 없음):
  - **DEBT-TEST-001 (부채 a)**: byte-level diff로 stale-golden 확정 (`rc6`→`rc8` 단일 문자 차이, 6 golden testdata 파일). renderer 정상. M1은 `UPDATE_GOLDEN=1` 단일 명령으로 청산 가능.
  - **DEBT-TEST-002 (부채 b)**: commit `794bb4f84` (2026-07-09)가 이미 HEAD에서 해결. `-count=10` + GLM env 주입 상태에서 모두 PASS. M2는 verify-only (코드 변경 없음, evidence 재생성만).
  - **DEBT-TEST-003 (부채 c)**: `go tool cover -func`로 pipeline.go 8함수 + human_oversight.go 2함수 0%/near-0% 확인. package 67.5% → 85% 상승 필요 integration test 추가.
- 부가 발견: `go test ./internal/... -count=1` 관측 시 `internal/web` 2개 i18n 실패(`TestDataI18nKeysSubsetOfDictionary`, `TestI18nKeySetParity`)가 HEAD에 존재하나 TEST-001 §E.4가 명명하지 않은 부채이므로 본 SPEC Out of Scope로 명시적 제외 (spec.md §C + acceptance.md §D FL-3). 후속 web-i18n SPEC(TEST-003) 소관.
- ARCH-001 선행 관계 SSOT: 본 SPEC §A — M1 necessary + web-i18n SPEC sufficient for ARCH-001 AC-ARCH-001.
- Epic framing: 5/5 → 6/6 (TEST-002 본 SPEC 5번째 + future web-i18n SPEC TEST-003 6번째).
- REQ count = 3 (REQ-TEST-007/008/009), AC count = 3 (AC-TEST-007/008/009) — Option A, 변경 없음.
- plan-auditor iter-2 verdict: **PASS-WITH-DEBT (0.96, skip-eligible)** — orchestrator-spawned plan-auditor가 D1-D5를 독립 검증하여 RESOLVED 확인 (2026-07-09). 신규 MINOR D6 (working-tree count drift "18+5"→실측 "12+9", C-6 pathspec은 count 무관) + D7 (AC-TEST-009 Then-5 + §D.4에 `HEAD~..HEAD` 잔존, D5와 동일 패턴, M3 단일 commit이라 latent). 보고서: `.moai/reports/plan-audit/SPEC-INTERNAL-TEST-002-2026-07-09.md`. **residual risk**: `pkg/version/version.go` 자체가 dirty(rc7→rc8, uncommitted) — manager-develop는 rc8을 hardcode하지 말고 version.go 실측값에서 rc-token을 re-derive.

```yaml
plan_complete_at: "2026-07-09"
plan_status: "audit-ready (iter-2)"
plan_artifacts:
  - spec.md (12 canonical frontmatter fields + era: V3R6 + tier: M + depends_on: [SPEC-INTERNAL-TEST-001])
  - plan.md (§A-H, Tier M template, §A.1 Epic table 6 rows)
  - acceptance.md (§D + §D.1-§D.7, 3 ACs: AC-TEST-007/008/009)
  - progress.md (본 파일, §E.1 iter-2 + skeleton)
tier: M
req_count: 3
ac_count: 3
iter_1_verdict: "FAIL (0.82, BLOCKING D1)"
iter_2_verdict: "PASS-WITH-DEBT (0.96, skip-eligible, monotonic increase from iter-1)"
iter_2_defects_resolved: [D1, D2, D3, D4, D5]
iter_2_new_defects: [D6 MINOR working-tree count drift, D7 MINOR HEAD~..HEAD residual AC-TEST-009/§D.4]
option: "A (per-debt scope, NO M4, web-i18n deferred to TEST-003)"
epic_framing: "5/5 → 6/6"
authoring_session_id: pending
```

## §E.2 Run-phase Evidence

_<pending run-phase — manager-develop이 AC-TEST-007/008/009 evidence + commit SHA 기록>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase — manager-develop이 run_status + AC PASS/FAIL matrix 기록>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase — manager-docs가 sync_commit_sha + 3-phase close 기록>_

## §F Phase 0.95 Mode Selection

### Input parameters

- tier: M
- scope (file count): M1 = 6 testdata files (`internal/cli/testdata/{doctor,status}-{light,dark,nocolor}.golden`); M2 = 0 code files (verify-only); M3 = 1-2 new `_test.go` files (`internal/constitution/`) → 총 ~8 files
- domain count: 2-3 (internal/cli testdata · internal/constitution test · statusline verify) — 그러나 sequential dependency
- file language mix: Go test data + Go test code (100% Go)
- concurrency benefit: LOW — coding-heavy (Anthropic coding-task parallelism caveat) + M1→M2→M3 sequential dependency + M1은 ARCH-001 re-entry critical-path gate(P0)
- Agent Teams prereqs: NOT met (`workflow.team.enabled` default false; `CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS` unset)

### Mode evaluation

| Mode | 선택 | 근거 |
|------|------|------|
| 1 trivial | 아니오 | multi-file, M3 integration test 실제 저작 |
| 2 background | 아니오 | write work, not read-only |
| 3 agent-team | 아니오 | prereqs 미충족 + coding-heavy |
| 4 parallel | 아니오 | domains < 3, coding-heavy, M1→M2→M3 sequential dependency → fan-out 이익 제한 |
| 5 sub-agent | **YES** | Tier M + coding-heavy sequential DDD; M3 integration test 저작 + M1 golden regen 메커니즘 |
| 6 workflow | 아니오 | < 30 files, semantic(신규 test 저작) not mechanical |

### Decision: sub-agent (Mode 5)

### Justification

Tier M + coding-heavy → Mode 5 (sub-agent sequential) per Anthropic's coding-task parallelism caveat ("most coding tasks involve fewer truly parallelizable tasks than research"). M1(golden regen, P0) → M2(verify-only, P2) → M3(integration test, P1)는 sequential dependency — M1은 ARCH-001 re-entry의 critical-path gate. M1/M3는 파일 scope가 disjoint (`internal/cli/testdata/` vs `internal/constitution/*_test.go`)하나, M2가 verify-only(코드 변경 0)이고 전체를 단일 `manager-develop` DDD cycle이 처리하는 것이 token 효율적 + sequential 의존성 보존에 부합. `cycle_type=ddd` (기존 코드에 대한 characterization/integration test 추가; coverage baseline 67.5%). Phase 0.5 plan-auditor verdict = PASS-WITH-DEBT 0.96 (skip-eligible) — Implementation Kickoff Approval(사용자 HUMAN GATE)은 별도 mandatory.

## §G IGGDA Kickoff Predicate

_<pending run-phase>_

- (a) intent clarity 100%: _<pending — 본 SPEC 위임 프롬프트가 M1/M2/M3 + 3 부채 + Milestone 순서 명시>_
- (b) plan-auditor PASS: **PASS (PASS-WITH-DEBT 0.96, iter-2, 2026-07-09, orchestrator-spawned)** — Phase 0.5 gate cleared
- (c) Tier M: PASS (plan.md §A.3)
- (d) no dangerous keywords: PASS (database migration 아님; `internal/migration` 패키지 본 SPEC scope 밖)
