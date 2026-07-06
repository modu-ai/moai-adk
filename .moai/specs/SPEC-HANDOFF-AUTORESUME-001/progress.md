# Progress — SPEC-HANDOFF-AUTORESUME-001

> Lifecycle 3-phase: plan → run → sync. 본 파일은 phase별 audit-ready signal + evidence를 누적한다. §E.1은 plan-phase(manager-spec), §E.2/§E.3은 run-phase(manager-develop), §E.4는 sync-phase(manager-docs) 소관.

## §E.1 Plan-phase Audit-Ready Signal

- 작성일: 2026-07-05 (worktree HEAD `97723664c`, clean detached base)
- 산출물 (Tier L 5-artifact + progress skeleton):
  - `spec.md` — frontmatter 12-field + tier:L/era:V3R6/related_specs, GEARS REQ-AUTORESUME-001..019, Out-of-Scope h3 6개
  - `plan.md` — M1(config)/M2(save CLI)/M3(SessionStart 소비) milestone-split + AC 바인딩
  - `acceptance.md` — AC-AUTORESUME-001..019 (REQ 1:1), edge cases, quality gate, DoD
  - `design.md` — 경로 분리 verdict, 4-source×mode branch table, nonce fallback, HandoffConfig, i18n degrade
  - `research.md` — registry accumulate-all 실측 반증, SessionStart matcher already-clear 실측 반증, config 패턴 미러
- SPEC ID self-check: `decomposition: SPEC ✓ | HANDOFF ✓ | AUTORESUME ✓ | 001 ✓ → PASS`
- 확정 사용자 결정 준수: mode default=manual / directive degrade-to-guidance / M1-M2-M3 split
- 실측 근거: registry.go:208-215 accumulate-all, settings.json:5 + .tmpl:6 이미 `startup|resume|clear|compact`
- plan-auditor iter-1 PASS-WITH-DEBT 0.85 → 정정 반영: D1(TTL auto-only, manual pure no-op) · D2(REQ-007 split → REQ-019, milestone 경계) · D3(Consume 필드 YAGNI 제거) · D4(branch table 8-cell + guide 양분기 AC) · D5(CLI 등록/writer 패키지 확정) · D6(rename-fail errno-무관 fail-open) · D7(nonce filename shape AC + 충돌-불가 논증 design prose). REQ/AC 18→19.
- plan-auditor iter-2 PASS-WITH-DEBT 0.89 (D1-D7 all RESOLVED on normative surfaces) → 최종 polish 반영(prose-level, REQ/AC 19 불변): N1(REQ-019 ⟩ REQ-010 stale-precedence 문장 — spec §C + design §C.2 + AC-010 live-scope + AC-019 no-hint sub-case) · N2(design §A 다이어그램 "ENOENT" → "rename failure — errno-agnostic" §C.3 정합) · N3(design §C.4 `ts` = 소비시각 정수 `UnixNano()` 명시, RFC3339 `saved_at` 문자열 아님 → AC-014 `^\d+-` 정규식 성립).
- plan_status: **audit-ready** (iter-2 0.89 ≥ Tier L 임계 0.85; D1-D7 + N1-N3 clean)
- plan_complete_at: 2026-07-05

## §E.2 Run-phase Evidence

- 구현 방식: TDD (RED→GREEN→REFACTOR working-tree, milestone별 atomic GREEN 커밋). 격리 worktree 브랜치, no push (orchestrator landing 예정).
- cycle_type: tdd. Tier L. plan-auditor iter PASS-WITH-DEBT (progress §E.1: iter-2 0.89; kickoff prompt: iter-3 0.90) ≥ Tier L 임계 0.85 → 구현 착수 승인 granted.

### AC PASS/FAIL Matrix (19/19 PASS)

| AC | Milestone | Status | Verification (command) | Actual Output |
|----|-----------|--------|------------------------|---------------|
| AC-001 | M1 | PASS | `go test ./internal/config/ -run TestNewDefaultHandoffConfig` | Mode=="manual" ∧ Guide==false; no Consume 필드 — ok |
| AC-002 | M1 | PASS | `go test ./internal/config/ -run TestHandoffRegistered` | IsRegisteredOrException("handoff")==true, registry["handoff"]=="HandoffConfig" — ok |
| AC-003 | M1 | PASS | `go test ./internal/config/ -run TestLoadHandoffSection_PartialOverride` | `mode: auto`만 명시 → Mode=auto ∧ Guide=false (default 유지) — ok |
| AC-004 | M1 | PASS | `grep -cE 'SPEC-\|REQ-' handoff.yaml` = 0; `grep -c 'consume' handoff.yaml` = 0 | 0 / 0 — 중립 template (mode/guide만), internal_content_leak_test ok |
| AC-005 | M2 | PASS | `go test ./internal/cli/ -run TestHandoffSave_WritesJSONNotMarkdown` + `./internal/hook/handoff/ -run TestSavePending_WritesJSONNotMarkdown` | handoff/pending.json valid JSON; session-handoff/pending.md decoy mtime 불변 — ok |
| AC-006 | M2 | PASS | `go test ./internal/cli/ -run TestHandoffSave_Schema` | schema_version/body(verbatim)/directives.ultrathink/conversation_language/saved_at 존재 — ok |
| AC-007 | M2 | PASS | `go test ./internal/cli/ -run TestHandoffClear` | pending.json 제거 + decoy mtime 불변 — ok |
| AC-008 | M3 | PASS | `go test ./internal/hook/ -run TestBranchTable_AutoMode` (4 source, live) | clear만 inject+consumed(1); startup/resume/compact 보존(0 consumed) — ok |
| AC-009 | M3 | PASS | `go test ./internal/hook/ -run 'TestManualMode_NoOp\|TestManualMode_StalePendingPreserved'` | manual 4 source 바이트 불변; stale sub-case도 불변 — ok |
| AC-010 | M3 | PASS | `go test ./internal/hook/ -run TestNonClearSource_NoticeOnly` (source×guide, live) | 미소비+미주입; guide==true stderr 힌트 1건, guide==false 억제 — ok |
| AC-011 | M3 | PASS | `go test ./internal/hook/ -run TestDegradeToGuidance` | additionalContext NotContains("xhigh") ∧ Contains(ultrathink 안내 "입력") — ok |
| AC-012 | M3 | PASS | `go test ./internal/hook/ -run TestClaimThenInject_AuditPreserved` | pending 부재, consumed/<ts>-<nonce>.json 내용==원본, memory project_*.md 보존 — ok |
| AC-013 | M3 | PASS | `go test -race ./internal/hook/ -run 'TestConcurrentConsume_SingleWinner\|TestRenameFailure_FailOpen'` | 2 goroutine → 정확히 1 inject+1 consumed; injected rename err(비-ENOENT) → skip+정상반환 — ok |
| AC-014 | M3 | PASS | `go test ./internal/hook/ -run TestNonceFallback_FilenameShape` | consumed 파일명 `^\d+-[0-9a-f]{8}\.json$` 매칭 (empty session → crypto/rand) — ok |
| AC-015 | M3 | PASS | `go test ./internal/hook/ -run TestInjectionHeader_I18n` | ko/en/ja/zh header 각 언어; fr→en fallback — ok |
| AC-016 | M3 | PASS | `grep -rn 'AskUserQuestion\|mcp__askuser' internal/hook/ \| grep -v _test.go \| grep -v '// '` | 0 matches; `TestNoUserInteraction` ok |
| AC-017 | M3 | PASS | `go test ./internal/hook/ -run TestFailOpen_CorruptPending` | 손상 JSON → allow 반환, pending 보존(rename 안 함), consumed 0 — ok |
| AC-018 | M3 | PASS | `go test ./internal/hook/ -run TestThreeHandlerCoexist` | 3-handler registry merged additionalContext = sessionStart attribution ∧ handoff 안내 (accumulate-all) — ok |
| AC-019 | M3 | PASS | `go test ./internal/hook/ -run TestStaleTTLCleanup_AutoOnly` | auto+stale → 조용히 제거, no inject, no consumed, N1 hint 억제(guide==true여도); manual+stale 불변 대조 — ok |

### Invariant Evidence

| Invariant | Status | Evidence |
|-----------|--------|----------|
| 기존 테스트 무손상 (config/hook/handoff/cli-handoff) | PASS | config/hook/handoff 패키지 green; cli handoff 6 test green |
| session-handoff/pending.md 무접촉 (경로 격리) | PASS | production code(pending.go/handoff.go/handoff_inject*.go) session-handoff write 0; decoy mtime 불변 test |
| settings.json matcher 무변경 (assertion only) | PASS | `git diff` .claude/settings.json + .tmpl = empty; matcher `startup\|resume\|clear\|compact` 존재 확인 |
| verification-claim-integrity (주입 xhigh 미주장) | PASS | AC-011 NotContains("xhigh"); i18n disclaimer "확장 추론 모드를 자동으로 활성화하지 않습니다" |
| cross-platform (windows MoveFileEx fail-open) | PASS | `GOOS=windows GOARCH=amd64 go build ./...` exit 0; rename errno-무관 fail-open(AC-013) |

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-07-06
run_commit_sha: "382b72ae9"   # 마지막 run-phase code 커밋(lint fix); progress.md evidence 커밋은 후속 (self-referential, backfill 가능)
run_status: PASS
ac_pass_count: 19
ac_fail_count: 0
preserve_list_post_run_count: 3   # session-handoff/pending.md · settings.json · settings.json.tmpl 무접촉 확인
l44_pre_commit_fetch: "N/A — 격리 worktree(9fbc7a2be detached base), orchestrator가 landing 시 pre-spawn fetch 수행"
l44_post_push_fetch: "N/A — no push (커밋 정책: worktree-branch commits only, orchestrator lands)"
new_warnings_or_lints_introduced: "0 net — 3건 도입(errcheck×2 + staticcheck×1) 후 382b72ae9에서 전부 수정"
cross_platform_build:
  host_darwin_arm64: "exit 0"
  windows_amd64: "exit 0 (GOOS=windows GOARCH=amd64 go build ./...)"
total_run_phase_files: 18   # config 4 mod + config test 1 + template yaml 1 + live yaml 1 + hook/handoff pending.go + pending_test.go + cli handoff.go + handoff_test.go + hook handoff_inject.go + _render.go + _test.go + _cover_test.go + cli deps.go mod + spec.md/plan.md frontmatter
m1_to_mN_commit_strategy: "M별 atomic GREEN (RED 커밋 없음): M1 30840b276 · M2 b165fa0ff · M3 a2d65265d · cover 6e0332236 · lint 382b72ae9. no push."
coverage:
  config_new_funcs: "NewDefaultHandoffConfig 100% · loadHandoffSection 100%"
  hook_handoff_pkg: "86.6% (pending.go: ConsumedDir/ClearPending/ReadPending 100%, SavePending 86.7%)"
  hook_handoff_inject: "renderHandoffContext/handoffLocale*/handoffConfig/handoffStale/isHex8 100% · Handle 95.2% · claimAndInject 80% · consumeNonce 85.7%"
pre_existing_baseline_failure: "internal/cli TestRunHookEvent_ReadInputError (coverage_test.go:77) — HEAD 9fbc7a2be에서 이미 panic(nil deref, hook post-tool RunE); handoff 무관, 본 SPEC 범위 밖"
```

## §E.4 Sync-phase Audit-Ready Signal

- 작성일: 2026-07-06 (orchestrator-direct sync — manager-docs가 CHANGELOG 7720줄 read로 autocompact thrashing 실패하여 전환; `feedback_glm_orchestrator_direct_sync_mx` fallback 패턴)
- Run-phase landing: origin/main `2b0efe516` (6 atomic GREEN 커밋 M1 `30840b276` / M2 `b165fa0ff` / M3 `a2d65265d` / cover `6e0332236` / lint `382b72ae9` / docs `2b0efe516`, L1 격리 worktree → FF push, 병렬 세션 로컬 main 86건 무접촉)
- 독립 재검증 (orchestrator, 8/8 PASS): `go test ./internal/config/... ./internal/hook/...` ok / `go test -race ./internal/hook/...` clean(AC-013) / AC-016 경계 grep 0 matches / `go build ./...` + `GOOS=windows` exit 0 / `golangci-lint` 0 issues / `go vet` exit 0 / `moai spec lint` spec+plan No findings / coverage handoff 86.6% (≥85%)
- plan-auditor iter-3 PASS-WITH-DEBT 0.90 (Tier L 임계 0.85; 0.85→0.89→0.90 단조 증가); D1 MINOR(AC-008 GIVEN live-scope 한정어 누락, REQ-level 모순 없음) → run-phase test-authoring guidance로 carry(AC-008 fixture는 LIVE non-stale pending 사용, §E.2 AC-008 row 확인)
- 19/19 AC PASS (AC-AUTORESUME-001..019, REQ 1:1)
- 경로 격리 확인: `session-handoff/pending.md` 무접촉, settings.json 무변경 (matcher 이미 `startup|resume|clear|compact`, assertion only)
- frontmatter: spec.md status in-progress → completed + updated 2026-07-06 (era V3R6, H-4 predicate)
- CHANGELOG: [Unreleased] `### Added` 엔트리 추가 (dup grep = 1)
- pre-existing baseline: `internal/cli` `TestRunHookEvent_ReadInputError` nil-deref panic — base `9fbc7a2be`(handoff refs 0)에서도 동일 재현, SPEC scope 밖 (manager-develop 결백)
- sync_commit_sha: b1ea0b9f9
