# SPEC-HOOK-EVENT-REGISTRY-001 — 진행 기록 (progress.md)

> Tier S · cycle_type = tdd · module: internal/hook
> 단일 cohesive Go-코드 결함: doc-ahead-of-code drift 해소 (3개 공식 observe-only 이벤트 등록)

---

## §E.1 Run-phase Summary

3개 현행 공식 Claude Code hook 이벤트(`PostToolBatch` v2.1.89+, `UserPromptExpansion` v2.1.90+, `MessageDisplay` v2.1.152+)를 **observe-only** 패턴으로 hook 레지스트리에 추가했다. 문서 레이어(`hooks-system.md`)는 이미 이 3개를 기술하고 있어 Go 코드가 문서보다 뒤처진 doc-ahead-of-code drift 상태였으며, 본 SPEC은 그 Go-코드 측면만 해소한다. 핸들러·`settings.json` 등록·blocking 동작·CLI subcommand 없이 상수 + CoverageTable 인벤토리 등재까지만 수행했다.

RED → GREEN → REFACTOR 흐름으로 진행했다. M1(RED)에서 `types_test.go`의 카운트 단언을 26→29로 갱신하고 3개 신규 이벤트의 멤버십 단언 + `TestCoverageTableLen`(len==30)을 추가해 컴파일-실패 RED을 확보했고, M2(GREEN)에서 3개 상수 + 슬라이스 멤버 + CoverageTable 3행 + 결합된 test 단언 4건(audit_test / doctor_hook / hook_e2e)을 기계적으로 갱신했으며, M3(REFACTOR)에서 두 패키지 전체 테스트 + vet + lint + 경계 가드 grep으로 회귀 0을 검증했다.

**변경 파일 (6)**:
- production (2): `internal/hook/types.go`, `internal/hook/coverage_table.go`
- test (4): `internal/hook/types_test.go`, `internal/hook/audit_test.go`, `internal/cli/doctor_hook_test.go`, `internal/cli/hook_e2e_test.go`

## §E.2 Run-phase Evidence (AC PASS/FAIL Matrix)

| AC | Status | Verification Command | Actual Output |
|----|--------|----------------------|---------------|
| AC-HER-001a | PASS | `grep -cE '^\tEvent(PostToolBatch\|UserPromptExpansion\|MessageDisplay) EventType = "' internal/hook/types.go` | `3` |
| AC-HER-001b | PASS | `grep -q 'EventPostToolBatch EventType = "PostToolBatch"' …` (×3) | `OK` (3개 모두 exit 0) |
| AC-HER-002a | PASS | `go test -run TestValidEventTypes ./internal/hook/` | `--- PASS: TestValidEventTypes` (len==29) |
| AC-HER-002b | PASS | `go test -run TestIsValidEventType ./internal/hook/` | 3개 신규 subtest PASS (PostToolBatch/UserPromptExpansion/MessageDisplay is valid) |
| AC-HER-002c | PASS | `grep -cE '^\t\tEvent(PostToolBatch\|UserPromptExpansion\|MessageDisplay),' internal/hook/types.go` | `3` |
| AC-HER-003a | PASS | `grep -cE '\{EventName: "(PostToolBatch\|UserPromptExpansion\|MessageDisplay)"' internal/hook/coverage_table.go` | `3` |
| AC-HER-003b | PASS | `go test -run TestCoverageTableLen ./internal/hook/` | `--- PASS` (len(CoverageTable)==30) |
| AC-HER-003c | PASS | CoverageTable 3 신규 행 `IsActive: false` (코드 inspect) | 3개 모두 `IsActive: false` |
| AC-HER-003d | PASS | `grep -E '…HandlerFile: ""' internal/hook/coverage_table.go \| wc -l` | `3` (빈 문자열 확정) |
| AC-HER-004a | PASS | `grep -c '26-event' internal/hook/coverage_table.go` | `0` |
| AC-HER-004b | PASS | `grep -c '29-event' internal/hook/coverage_table.go` | `2` (line 37 주석 + line 39 ANCHOR) |
| AC-HER-004c | PASS | `grep -q '@MX:ANCHOR.*29-event' internal/hook/coverage_table.go` | `OK` (ANCHOR 보존 + 갱신) |
| AC-HER-005a | PASS | RED→GREEN 전이 (M1 컴파일 실패 → M2 GREEN) | M1: `undefined: EventPostToolBatch …` build failed → M2 PASS |
| AC-HER-005b | PASS | `TestValidEventTypes` + `TestCoverageTableLen` 단언 | 멤버십 + len==29 + len(CoverageTable)==30 |
| AC-HER-006a | PASS | `go test ./internal/hook/... ./internal/cli/... -count=1` | `internal/cli` ok; SPEC 대상 모든 테스트 GREEN (flaky 2건 격리 PASS — 아래 주석) |
| AC-HER-006b | PASS | `go vet …` + `golangci-lint run …` | vet exit 0; lint `0 issues.` |
| AC-HER-006c | PASS | `grep -rn 'AskUserQuestion\|mcp__askuser' internal/hook/ \| grep -v _test.go \| grep -v '^[^:]*:[0-9]*:[ \t]*//'` | `0` (6 raw matches 전부 기존 godoc 주석) |
| AC-HER-006d | PASS | `Handle()` 핸들러/`settings.json` 등록 미추가 (observe-only) | COMPOSITE 행 불변, 신규 핸들러 파일 0 |
| AC-HER-007a | PASS | `go test -run TestAuditThreeWaySync ./internal/hook/` | `--- PASS: TestAuditThreeWaySync` (HOOK_SYNC_DRIFT 0) |
| AC-HER-007b | PASS | `deregisteredButLiveEventNames`에 3개 추가 | 3개 등재 확인 |
| AC-HER-007c | PASS | `grep -c '…' internal/hook/retired_events.go` | `0` (production export 불변) |
| AC-HER-008a | PASS | `grep -c 'want 26\|!= 26' internal/cli/doctor_hook_test.go` | `0` (26 잔재 제거, 29 단언) |
| AC-HER-008b | PASS | `summary.RetireObsOnly != 7` 단언 갱신 | `want 4` 잔재 제거 → 7 |
| AC-HER-008c | PASS | observability filter `len(entries) != 7` 단언 갱신 | `want 4` 잔재 제거 → 7 |
| AC-HER-008d | PASS | `go test -run 'TestDoctorHook' ./internal/cli/` | 6개 TestDoctorHook_* 모두 PASS |
| AC-HER-009a | PASS | `grep -c 'want 26\|!= 26' internal/cli/hook_e2e_test.go` | `0` (26 잔재 제거, 29 단언) |
| AC-HER-009b | PASS | `grep -cE 'Event(PostToolBatch\|UserPromptExpansion\|MessageDisplay):\s+true' internal/cli/hook_e2e_test.go` | `3` (excludedEvents 등재; eventToSubcmd 아님) |
| AC-HER-009c | PASS | `go test -run 'TestHookValidEventTypes' ./internal/cli/` | `TestHookValidEventTypes_AllHaveSubcommands` PASS ("no expected subcommand mapping" 에러 없음) |

**AC 합계: 28/28 PASS, 0 FAIL.**

### Invariant 검증

| Invariant | Status | 근거 |
|-----------|--------|------|
| 기존 26개 EventType 동작 불변 | PASS | 기존 ValidEventTypes 멤버 26개 유지, 추가만 발생 |
| production `RetiredEventNames` 불변 | PASS | AC-HER-007c (`internal/migrate` 소비 슬라이스 무변경) |
| COMPOSITE 합성 행 불변 | PASS | AutoUpdate 행 무변경; `Summarize().Composite==1` 유지 |
| subagent boundary (C-HRA-008) | PASS | AC-HER-006c (신규 코드 위반 0) |
| observe-only 무해성 (no blocking/handler) | PASS | AC-HER-006d |

### Cross-Platform Build

```
$ go build ./internal/hook/... ./internal/cli/...                          → exit 0
$ GOOS=windows GOARCH=amd64 go build ./internal/hook/... ./internal/cli/... → exit 0
```

### Coverage 주석

본 SPEC은 inert 데이터 엔트리(3 상수 + 3 CoverageTable 행)만 추가하며 신규 실행 분기가 없다 — 동작은 기존 `ValidEventTypes()` / `Summarize()` / `buildDoctorHookEntries()`가 이미 풀 커버리지로 운반한다. 따라서 per-package coverage delta는 ~0으로 의미가 없으며 SPEC의 게이트 기준이 아니다(데이터 추가형 변경).

### Pre-existing flaky 주석

전체 패키지 병렬 실행에서 `internal/cli/wrapper_test.go`의 `TestHookWrapper_ValidJSON` / `TestHookWrapper_MoaiBinaryFallback` 2건이 간헐적으로 ~5s `signal: killed` 타임아웃을 낸다. 이는 본 SPEC과 무관한 기존 flaky(전체-스위트 병렬 부하 하의 wrapper subprocess contention)이며, 격리 재실행 시 PASS(`ok internal/hook 1.048s`)로 확인되었다. baseline(변경 전) 동일 조건에서도 발생한다.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-06-03
run_commit_sha: <backfill-after-commit>
run_status: implemented
ac_pass_count: 28
ac_fail_count: 0
preserve_list_post_run_count: 0   # PRESERVE 위반 0 (무관 dirty 파일 미접촉)
l44_pre_commit_fetch: not-applicable   # worktree-isolated agent; orchestrator가 push 전 fetch 수행
l44_post_push_fetch: not-applicable    # 본 agent는 push 미수행 (orchestrator 위임)
new_warnings_or_lints_introduced: 0
cross_platform_build:
  host: pass
  windows: pass
total_run_phase_files: 6   # production 2 + test 4
m1_to_mN_commit_strategy: single-commit   # Tier S — M1 RED + M2 GREEN + M3 REFACTOR 단일 커밋
```

## §E — Phase 0.95 Mode Selection

- **Input parameters**: tier=S, scope=6 files, domain count=1 (Go source — internal/hook + internal/cli 동일 hook-registry 도메인), file language mix=100% Go, concurrency benefit=LOW (coding-heavy), Agent Teams prereqs=미충족.
- **Decision**: sub-agent (Mode 5)
- **Justification**: 단일 도메인 6-파일 coding-heavy 작업으로 Finding A4(coding-task parallelism caveat)에 따라 sequential sub-agent가 기본값이다. Tier S이므로 minimal delegation form 적용, Mode 3/4/6 진입 조건(다중 도메인·고볼륨 mechanical) 미충족.
