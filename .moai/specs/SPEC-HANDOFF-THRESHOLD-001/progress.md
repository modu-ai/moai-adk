# Progress — SPEC-HANDOFF-THRESHOLD-001

> Lifecycle 3-phase: plan → run → sync. phase별 audit-ready signal + evidence 누적. §E.1은 plan-phase(manager-spec), §E.2/§E.3은 run-phase(manager-develop), §E.4는 sync-phase(manager-docs) 소관.

## §E.1 Plan-phase Audit-Ready Signal

- 작성일: 2026-07-06 (Epic Handoff-v2 M4/4, 마지막 마일스톤)
- 산출물 (Tier M 3-artifact + design/research + progress skeleton = 5 + 1):
  - `spec.md` — frontmatter 12-field + tier:M/era:V3R6/related_specs, GEARS REQ-THRESHOLD-001..018, Out-of-Scope h3 6개
  - `plan.md` — M1(config+stage)/M2(state-file)/M3(독트린+화해) milestone-split + 6 blocker 해소(4 C-axis + D1/D2) + AC 바인딩, Out-of-Scope h3
  - `acceptance.md` — AC-THRESHOLD-001..018 (REQ 1:1), edge cases, quality gate, DoD, 잔여 위험(concurrent empty-id 포함)
  - `design.md` — stage enum + 하드 상한 공식(clamp) + state-file 스키마(+writer_pid) + Guide/Mode 화해 + 호출부 + §D.4a/§D.4b concurrent-empty-id
  - `research.md` — autoCompactThreshold 위치/호출부/atomic 선례/template mirror drift 실측
- SPEC ID self-check: `decomposition: SPEC ✓ | HANDOFF ✓ | THRESHOLD ✓ | 001 ✓ → PASS`
- 2 LOCKED 결정 준수: 기존 HandoffConfig 필드만 소비 / 밴드 경계 defaults.go 상수 하드코딩
- 6 blocker 해소 (4 C-axis + plan-auditor iter-1 D1/D2):
  - B1 하드 상한 unreachable → `min(95, getAutoCompactThreshold()+10)` + `hard<soft` clamp + reachability 문서(REQ-005). autoCompactThreshold=memory.go 동일 패키지(open question 아님)
  - B2 write 호출부 → `builder.Build`(collectAll 직후, session_id+Memory 동시 스코프)
  - B3 session_id guard → last-writer-wins 스탬프 + reader 불일치 stale
  - B4 fallback-UUID → session_id 부재 시 captured_at freshness(single-session 생존)
  - **D1 (iter-1 SHOULD-FIX) template drift** → template mirror 256K 행 부재(실측 `grep -c`=0) vs LIVE 존재(=1). D4가 template에 256K 행 ADD(parity, REQ-017/AC-017) + Detection 절 section-level 편집(BOTH), full-sync 금지(LIVE 256K 삭제 회귀 방지). AC-016 `<doctrine>`=LIVE 명시.
  - **D2 (iter-1 SHOULD-FIX) concurrent empty-id hole** → B4 empty-id fallback이 B3 guard 재개방(UUID-less 2+ concurrent). `writer_pid` discriminator(REQ-018/AC-018) Go 헬퍼 레벨 기계 차단. 잔여: 독트린-only reader 미비교(보수적 폴백), 후속 Go reader가 완전 폐쇄. Tier M 비례(trigger 드묾).
- 실측 정정: Detection 독트린 **template mirror 존재**(`internal/template/templates/...`) → task 전제("template 밖") 오기. 256K 행 **LIVE만 존재(=1), template mirror 부재(=0)** → M1 drift, D4가 parity 회복.
- M1 무회귀 불변식 = AC-THRESHOLD-006 (statusline suffix config 무관, default guide=false에서 soft 유지)
- plan-auditor iter-1: PASS-WITH-DEBT 0.84 (4 MUST-PASS pass, no BLOCKING); D1/D2 SHOULD-FIX 반영 → REQ/AC 16→18
- plan_status: **audit-ready** (plan-auditor Tier M 임계 0.80; iter-1 0.84 ≥ 0.80, D1/D2 정정 완료)
- plan_complete_at: 2026-07-06

## §E.2 Run-phase Evidence

> cycle_type=tdd, per-milestone atomic GREEN 커밋(RED 커밋 없음). M1=eb23fc075(선행 landing), M2=2aaf43986, M3=0a843503e(로컬 worktree, push는 orchestrator 소관). 각 AC의 Actual Output은 이 run에서 실제 관측한 명령 출력.

| AC | REQ | M | Actual Output (관측 증거) | Status |
|----|-----|---|---------------------------|--------|
| AC-001 | REQ-001 | M1 | `TestHandoffGuideStage_WrapperEquivalence` PASS — `shouldShowHandoffGuide == (handoffGuideStage != none)` 전 케이스 성립, 기존 M1 테스트 무손상 | PASS |
| AC-002 | REQ-002 | M1 | `TestRenderBarsInline_TwoStageSuffix` PASS — none→suffix無 / soft→`(⚠️/clear)` / hard→`(🛑/clear!)` | PASS |
| AC-003 | REQ-003 | M1 | `TestHardCeiling_Formula`(hardCeilingPct(256K)=95, 1M=95) + `TestHardCeiling_ClampBelowSoft`(env=60→clamp hard=soft=90) PASS | PASS |
| AC-004 | REQ-004 | M1 | `TestHandoffThresholdConstants` PASS(5 상수 값 검증); `grep -cE '500_000\|500000' renderer.go` = **0**(inline literal 부재, config 상수 참조) | PASS |
| AC-005 | REQ-005 | M3 | `grep -icE 'auto-compact.*pre-empt\|rarely fires\|frequently pre-empted'` = **2**(LIVE)/2(template); `grep -ic 'always fires'` = **0**/0 (verification-claim-integrity — stage-2 always fires 미주장) | PASS |
| AC-006 | REQ-006 | M1 | **INVARIANT** — `TestNoM1Regression_SuffixUnconditional` PASS(default guide=false에서 soft suffix 유지); static `var _ func(*StatusData) handoffStage = handoffGuideStage`; `grep -c 'HandoffConfig' renderer.go` = **0**(suffix 게이팅 부재) | PASS |
| AC-007 | REQ-007 | M2 | `TestWriteContextUsage_ConfigIndependent` PASS; static `var _ func(string,string,int,MemoryData,handoffStage) = writeContextUsage`(HandoffConfig 파라미터 없음 — 컴파일 보증) | PASS |
| AC-009 | REQ-009 | M2 | `TestWriteContextUsage_Atomic`(temp+rename valid JSON, .tmp 잔존 없음) + `TestWriteContextUsage_SilentFail`(빈 projDir / Available=false / cwSize≤0 / MkdirAll 실패 → panic·error 없음) PASS | PASS |
| AC-010 | REQ-010 | M2 | `TestContextUsage_Schema` PASS — 9필드(schema_version/session_id/writer_pid/captured_at/context_window_size/tokens_used/raw_pct/stage/band) 존재 | PASS |
| AC-011 | REQ-011 | M2 | `TestBuild_WritesContextUsageWithSessionID` PASS — Build 후 파일 session_id="sess-build-011" ∧ context_window_size=256000 ∧ tokens_used=230400; `grep -c writeContextUsage builder.go`=**1**, renderer.go=**0** | PASS |
| AC-012 | REQ-012 | M2 | `TestWriteContextUsage_ThrottleSkipUnchanged` PASS — 동일 payload 재write는 mtime 불변(skip, writer_pid 22 변경에도 throttle 유지), tokens 변경 시 write 발생 | PASS |
| AC-013 | REQ-013 | M2 | `TestSessionGuard_MismatchStale` PASS — session_id A vs B → stale(false), A vs A → valid(true), UUID 경로 writer_pid 무관 | PASS |
| AC-014 | REQ-014 | M2 | `TestFallbackUUID_FreshnessValidation` PASS — 빈 session_id도 write 발생, 양측 "" + fresh + 동일 writer → valid, expired → stale, UUID×"" mix → 보수적 stale | PASS |
| AC-015 | REQ-015 | M3 | `grep -c 'context-usage.json'` = **2**(LIVE)/2(template); Detection §1 state-file 우선 + §3 fallback 4-signal 보존 서술 | PASS |
| AC-016 | REQ-016 | M3 | LIVE·template §Detection 모두 context-usage.json 반영(section-level); `grep -c '256,000' <LIVE>` = **1**(중복 없음, Targets 표 무접촉); template internal-content-leak CI 가드 `ok` | PASS |
| AC-017 | REQ-017 | M3 | `grep -c '256,000' <template>` = **1**(누락 행 ADD) ∧ `grep -c '256,000' <LIVE>` = **1** → 두 사본 parity(drift 제거); `make build` exit 0 | PASS |
| AC-018 | REQ-018 | M2 | `TestConcurrentEmptyID_WriterPIDGuard` PASS(table: match+fresh→valid, mismatch+fresh→stale(cross-read 차단), match/mismatch+stale→stale) + UUID 경로 writer_pid 무관 유지(AC-013 보존) | PASS |
| AC-008 | REQ-008 | M3 | `grep -c 'statusline.*미소비\|does not read HandoffConfig' design.md` = **2**; `git diff --stat eb23fc075..HEAD -- internal/hook/handoff_inject.go` = **empty**(M3 auto-resume 무접촉) | PASS |

**불변식(invariant) row**:

| Invariant | 관측 증거 | Status |
|-----------|-----------|--------|
| M1 무회귀 (AC-006) | default guide=false에서 soft suffix 유지, statusline이 HandoffConfig 미소비(컴파일 보증) | PASS |
| M1 band 로직 verbatim 승계 (B10 PRESERVE) | `softThresholdPct` = M1 ≥cutoff→50 / else→90 로직 그대로(리터럴만 config 상수화) | PASS |
| M3 handoff_inject.go 무접촉 (B10) | `git diff eb23fc075..HEAD internal/hook/handoff_inject.go` empty | PASS |
| 4-signal 휴리스틱 fallback 보존 (AC-015) | Detection §3에 4-signal 그대로 유지 | PASS |
| context-usage.json 커밋 미포함 (B8) | 런타임 생성물 — 코드만 커밋; TestMain이 test side-effect .moai/state 정리 | PASS |

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-07-06
run_commit_sha: 0a843503e   # M-final(M3) 로컬 worktree SHA; push 시 orchestrator rebase 가능 → backfill 대상
run_status: complete        # 18/18 AC PASS, 0 FAIL
ac_pass_count: 18
ac_fail_count: 0
preserve_list_post_run_count: 5   # M1 band verbatim / handoff_inject 무접촉 / M1 테스트 green / 4-signal fallback / LIVE 256K 무중복 — 5개 전부 준수, 위반 0
l44_pre_commit_fetch: deferred-to-orchestrator   # push 미수행(task 지시) — 병렬 statusline 세션 레이스는 orchestrator가 pre-push 처리
l44_post_push_fetch: deferred-to-orchestrator    # run-phase에서 push 없음
new_warnings_or_lints_introduced: 0   # golangci-lint 0 issues, go vet clean (statusline + config)
cross_platform_build:
  darwin_amd64: ok      # go build ./... exit 0
  windows_amd64: ok     # GOOS=windows GOARCH=amd64 go build ./... exit 0
  make_build: ok        # template embed 재컴파일 exit 0
coverage:
  statusline: 85.4%     # 신규 context_usage.go per-func avg ~96.5%
  config: 80.3%         # 패키지 pre-existing baseline — M2/M3는 config 무접촉, M1 상수는 compile-time(statement 커버리지 영향 없음), 회귀 아님
total_run_phase_files: 10   # M1 4(renderer.go/defaults.go/handoff_stage_test.go/handoff_thresholds_test.go) + M2/M3 6(context_usage.go/context_usage_test.go/builder.go/builder_test.go/LIVE doctrine/template doctrine)
m1_to_mN_commit_strategy: per-milestone-atomic-green   # origin/main landed: M1 eb23fc075(feat) → M2 318db2617(feat) → M3 8020e877d(docs); rebase 전 로컬 SHA는 2aaf43986/0a843503e (병렬 web-console 위 rebase로 rewrite), RED 커밋 없음
known_baseline_failure: internal/cli/TestRunHookEvent_ReadInputError   # coverage_test.go:77 pre-existing panic, base eb23fc075에서도 fail, internal/cli는 M4 무접촉(diff empty) → M4 무관
```

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_status: complete
sync_complete_at: 2026-07-06
sync_commit_sha: <pending-backfill>          # 이 sync 커밋 SHA — 다음 백필 커밋에서 기록 (spec.md era:V3R6 H-override라 classify 안전)
lifecycle_transition: in-progress → completed # 3-phase close (SPEC-V3R6-LIFECYCLE-REDESIGN-001)
run_landed_commits: eb23fc075(M1) 318db2617(M2) 8020e877d(M3)  # origin/main rebase 후 실제 landed SHA (병렬 web-console a73ceedf6 위 선형)
ac_matrix: 18/18 PASS (AC-THRESHOLD-001..018)
plan_auditor: iter-1 PASS-WITH-DEBT 0.84 (Tier M thresh 0.80); D1(template drift)/D2(B4 concurrent empty-id) SHOULD-FIX 해소
independent_verification: orchestrator 7/7 PASS  # go build+GOOS=windows / go test 89 ok(cli TestRunHookEvent_ReadInputError baseline 제외) / make build exit0 / 256K parity LIVE=1 template=1 / template §25 internal-token 0 / AC-006 불변 statusline HandoffConfig 시그니처 0 / 런타임 아티팩트 누출 0
sync_method: orchestrator-direct  # manager-docs CHANGELOG(7000+줄) autocompact thrashing 회피 (memory 교훈), frontmatter+CHANGELOG+§E.4 직접 편집
known_baseline_failure: internal/cli/TestRunHookEvent_ReadInputError  # M4 무관 pre-existing (base eb23fc075도 fail)
```
