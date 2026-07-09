---
id: SPEC-INTERNAL-TEST-002
version: "0.1.0"
status: completed
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

**Baseline**: HEAD `f3f348ff5` (synced origin/main, divergence 0 0). Run-phase executed in worktree `agent-aed48595b57db9c20` (pristine checkout of HEAD). **Worktree `version.go = v3.0.0-rc7`** (the HEAD committed value); the main-repo working tree carries an uncommitted `rc8` (release-process-owned, PRESERVE). Per the delegation prompt's anti-hardcode directive ("goldens follow version.go"), the observed worktree value `rc7` was used for M1 regen — this produces HEAD consistency (version.go=rc7 + goldens=rc7 → 6 tests PASS at HEAD), which is the correct outcome for unblocking ARCH-001's M0 gate.

| AC | Status | Verification Command | Actual Output (this run) |
|----|--------|---------------------|--------------------------|
| AC-TEST-007 (M1) | **PASS** | `UPDATE_GOLDEN=1 go test ./internal/cli/ -run 'TestDoctor_Current_Light\|TestDoctor_Current_Dark\|TestDoctor_NoColor\|TestStatus_Current_Light\|TestStatus_Current_Dark\|TestStatus_NoColor' -count=1` | regen exit=0; `git diff --stat internal/cli/testdata/` = 6 files, 6 insertions(+), 6 deletions(-); byte-level diff = only `rc6→rc7` (2 line patterns, no layout/emoji change → C-1 renderer untouched); `grep -l "v3.0.0-rc7" …{doctor,status}-{light,dark,nocolor}.golden \| wc -l` → **6**; re-run WITHOUT UPDATE_GOLDEN → `ok github.com/modu-ai/moai-adk/internal/cli 0.416s` (exit 0, 6/6 PASS) |
| AC-TEST-008 (M2) | **PASS** (verify-only) | `go test -run TestBuild_WritesContextUsageWithSessionID -count=10 ./internal/statusline/` + GLM-env `ANTHROPIC_DEFAULT_{OPUS,SONNET,HAIKU}_MODEL=glm-5.2 … -count=3` | 10x consecutive → `ok … 0.393s` (exit 0); GLM-env 3x → `ok … 0.236s` (exit 0). Zero code change (commit `794bb4f84` already-resolved — C-2 respected). |
| AC-TEST-009 (M3) | **PASS** | `go test -coverprofile=cover.out ./internal/constitution/` + `go tool cover -func` | coverage: **85.1%** (≥85% floor, up from 67.5%); all 10 named functions escaped 0%: NewPipeline 100%, Execute 82.2%, createLogEntry 100%, applyAmendment 35.7% (capped by documented `updateSourceFile` stub — unreachable lines after the stub error cannot be covered without modifying production, C-3), acquireLock 76.9%, releaseLock 100%, updateSourceFile 100%, updateRegistryClause 100%, Approve 85.7%, printDiff 100%. Production code (pipeline.go / human_oversight.go) UNCHANGED — C-3 verified via empty `git diff`. |

### E-deliverables (this run)

- **E2 Cross-platform build**: `go build ./...` exit 0; `GOOS=windows GOARCH=amd64 go build ./...` exit 0.
- **E3 Coverage**: internal/constitution **85.1%**; internal/cli + internal/statusline PASS (no coverage regression from M1/M2).
- **E4 Subagent boundary**: N/A (no harness/hook code touched — SPEC scope).
- **E5 Lint**: `golangci-lint run --timeout=2m` → `0 issues.` (0 NEW vs pre-flight baseline 0).
- **E6 Push**: see §E.3 (`l44_post_push_fetch`).

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: "2026-07-09"
run_start_commit_sha: "f3f348ff5"        # HEAD before M1 (synced origin/main, divergence 0 0)
m1_commit_sha: "0f1f07349"               # M1: 6 goldens (rc6→rc7) + 4 SPEC draft→in-progress
run_commit_sha: "<pending backfill — M3 commit SHA recorded post-commit>"
run_status: "PASS (3/3 AC: AC-TEST-007/008/009 all PASS)"
ac_pass_count: 3
ac_fail_count: 0
preserve_list_post_run_count: 0          # C-4/C-6: no PRESERVE files entered any commit
l44_pre_commit_fetch: "origin/main == f3f348ff5 (divergence 0 0 before run-phase)"
l44_post_push_fetch: "<pending push — orchestrator-side>"
new_warnings_or_lints_introduced: 0      # golangci-lint 0 issues (baseline 0)
cross_platform_build:
  darwin_amd64: "exit 0"
  windows_amd64: "exit 0"
total_run_phase_files: 13                # 6 goldens + 4 SPEC frontmatter (M1) + 2 new _test.go + progress.md §E.2/§E.3 (M3)
m1_to_mN_commit_strategy: "2 commits — M1 (goldens + draft→in-progress frontmatter), then M2+M3 bundled (constitution integration tests + M2 verify-only evidence + progress.md §E.2/§E.3)"
version_token_observed: "rc7 (worktree version.go at HEAD f3f348ff5; main-repo working tree has uncommitted rc8, PRESERVE)"
```

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: "2026-07-09"
sync_commit_sha: "<pending — backfill via follow-up chore commit>"
sync_status: "completed (3-phase close plan→run→sync; completed transition rides single sync commit per SPEC-V3R6-LIFECYCLE-REDESIGN-001 merged 3-phase close / Status Transition Ownership Matrix)"
b12_self_test_a: "PASS — pre-emission grep -c 'SPEC-INTERNAL-TEST-002' CHANGELOG.md returned 1; context inspection (grep -B3 -A3) confirmed the single occurrence is an incidental forward-reference INSIDE the SPEC-INTERNAL-TEST-001 entry (line 58: '후속 SPEC-INTERNAL-TEST-002가 잔여 3종 부채 소유 예정'), NOT a duplicate entry from a parallel sync session. B12 halt condition (count≥1 from parallel BATCH-SYNC) does not apply — this is a cross-SPEC reference naming TEST-002 as future debt-owner, written when TEST-001 synced. No ### subsection entry for TEST-002 existed pre-emission."
b12_self_test_b: "PASS — AC count 3 (AC-TEST-007/008/009) per acceptance.md §D AC Matrix (SSOT), matches CHANGELOG entry '3/3 AC PASS(AC-TEST-007/008/009)'"
b12_self_test_c: "PASS — file paths verified via ls before commit: 6 goldens (internal/cli/testdata/{doctor,status}-{light,dark,nocolor}.golden all exist, all carry v3.0.0-rc7 token) + 2 new test files (internal/constitution/pipeline_test.go 15112B + human_oversight_test.go 3167B both exist)"
changelog_entry_position: "[Unreleased] > ### Fixed (inserted after SPEC-INTERNAL-TEST-001 entry, before ## [v3.0.0-rc7] version header)"
frontmatter_status_transitions:
  spec_md: "in-progress → completed (sync commit)"
  plan_md: "in-progress → completed (sync commit)"
  acceptance_md: "in-progress → completed (sync commit)"
  progress_md: "in-progress → completed (sync commit)"
  updated_field_refreshed: "2026-07-09 (all 4 artifacts; date unchanged — run-phase and sync-phase both 2026-07-09)"
canary_compliance_check:
  preserve_list_violations: 0           # C-4: no PRESERVE files entered sync commit
  pathspec_discipline: "PASS — git add .moai/specs/SPEC-INTERNAL-TEST-002/ CHANGELOG.md only (never -A/-.)"
  era_classification: "V3R6 H-4 (§E.2 present + §E.4 present + sync_commit_sha non-empty; spec.md frontmatter era: V3R6 explicit override)"
  body_content_untouched: true          # spec/plan/acceptance body sections unchanged (frontmatter status+updated only)
  e2_e3_ownership_respected: true       # §E.2 + §E.3 owned by manager-develop — not modified by manager-docs
  ac_count: 3                           # AC-TEST-007/008/009
  ac_pass_count: 3
  ac_fail_count: 0
run_phase_commit_shas:
  m1: "ffea91710"                       # 6 goldens rc6→rc7 + 4 SPEC draft→in-progress frontmatter (on origin/main)
  m3: "65c23fbf2"                       # 2 new _test.go + progress.md §E.2/§E.3 (on origin/main)
  m2_strategy: "verify-only, bundled with M3 commit (code change 0, commit 794bb4f84 already-resolved)"
  note: "progress.md §E.3 m1_commit_sha records stale worktree SHA 0f1f07349 (rebased to ffea91710 on origin/main); §E.3 is manager-develop-owned — not corrected here"
sync_phase_note: "3-phase close: plan(iter-2 PASS-WITH-DEBT 0.96) → run(M1 ffea91710 + M3 65c23fbf2, 3/3 AC PASS) → sync(this commit). era: V3R6. tier: M."
```

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
