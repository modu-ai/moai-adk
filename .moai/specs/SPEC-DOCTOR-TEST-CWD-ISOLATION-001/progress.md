# 진행 기록: SPEC-DOCTOR-TEST-CWD-ISOLATION-001

## §E.1 Plan-phase Audit-Ready Signal

plan_status: audit-ready
plan_complete_at: 2026-09-13
baseline: `WT-doctor-red@dd235a66b1145922565841d33acafef0d1ded6a8` (RED ledger re-measured 2026-09-18, spec v0.3.0; raw output `.moai/reports/t675/red/`; prior baseline `74d872aafbd90235e67163a5bc233f7c8a934491`)
tier: S
artifacts: `spec.md` + `plan.md` — canonical Tier S set; inline ACs in `spec.md §3`
plan_artifact_hash (sha256, spec/plan v0.3.1 — wording-only corrections O1-O3 after plan-audit iteration 2 PASS 0.92):
- `spec.md`: `40d89d3e1dd0ba07c2b837c769caa53a57cd9f2ed8966a089d321a35e3198b32`
- `plan.md`: `bbdf56897cf72cb1d2592b1fb81d5776f142fff5f74db6e154d166fc64208ec5`
scope_decision: Operator-directed Tier S CWD isolation supersedes the card-origin embedded-C1 hypothesis and Tier M Class B classification.
red_baseline: With `MOAI_EMBED_CHECK_BIN=/usr/bin/false`, the exact nine-test doctor selection exits 1 with nine failures on the pinned baseline; the representative single test also exits 1.
checks:
- `moai spec lint SPEC-DOCTOR-TEST-CWD-ISOLATION-001 --strict --json` → exit 0 (2026-09-18, v0.3.0), one `info` finding `OwnershipTransitionUnmeasured` (commit `49bf74a82` carries no Authored-By-Agent trailer; not clearable without history rewrite)
- `moai spec audit --base-dir /Users/goos/MoAI/moai-adk-go/.claude/worktrees/t675 --filter-spec SPEC-DOCTOR-TEST-CWD-ISOLATION-001 --strict --json` → exit 0, `modern_era_clean: 1`, only `EraAutoDetected` INFO via H-5
prior_gateway_blocker: Preserved at `.moai/reports/t675/gateway-502-20260913.md`; the previous manager-spec delegation ended in HTTP 502 three times before this recovery session.
next_action: Hand off the unchanged Tier S plan artifacts to `plan-auditor`; no implementation has started.

## §E.1a 재배차 재측정 (2026-09-18)

- 흡수 후 트리: `WT-doctor-red@ae6ad726b` (로컬 develop `f67d2193f` 병합)
- 자연 상태 9건: `go test -count=1 -v -run 'TestRunDoctor_|TestDoctorCmd_' ./internal/cli/` → `ok ... 262.827s`. 이 selector 는 27개 테스트를 고르므로(plan-audit iter2 O3), 그 실행에서 확인된 것은 9개 대상 테스트가 각각 `--- PASS` 를 찍었다는 사실까지다.
- 자연 상태 9건 재측정(2026-09-18, spec v0.3.1): HEAD `8831e4297`(Go 변경 없음)에서 정확한 앵커 selector `-run '^(TestRunDoctor_WithExport|…|TestDoctorCmd_VerboseExecution)$'` → exit 0, `--- PASS:` 정확히 9줄, `--- FAIL` 0줄, `ok  	github.com/modu-ai/moai-adk/internal/cli	106.622s` (원본 `.moai/reports/t675/red/natural-9-anchored.txt`)
- 오염 주입 재현: `MOAI_EMBED_CHECK_BIN=/usr/bin/false go test -count=1 -v -run '^TestDoctorCmd_Execution$' ./internal/cli/` → `doctor command RunE error: doctor: 1 check(s) failed`, `--- FAIL`
- 자연 통과의 귀속: 이 트리에 `bin/moai` 가 없어 Agent Emit Embed 가 판정 대상 없음(정보성 건너뜀)으로 끝난다. 그 분기(`f6c027fa0`, SPEC-CI-DOCTOR-BIN-001)는 카드 실측 base `5ddccacc9` 에 이미 있었으므로, 카드 당시의 적색은 로컬에 빌드된 노후 `bin/moai` 가 있을 때만 나는 **환경 상태** 였다고 판단한다 (노후 바이너리 자체로 재현하지는 않았다 — 주입 대조로만 확인).
- 등급: Class B, Tier S 유지. CI 는 `bin/moai` 를 빌드하지 않으므로 배치 push 적색 요인이 아니다 → 우선순위 High 에서 강등 근거. 테스트가 주변 작업 폴더에 결과를 맡기는 결함은 남아 있어 plan 을 이어 간다.
- plan-audit(이전 세션): FAIL 0.67 — D1-D5 수리 후 재감사 필요.

## §E.2 Run-phase Evidence

- 측정일: 2026-09-18 · 편집 전 트리 HEAD `aead782bc` · 구현 커밋 M1 `85f1ba686` · 로컬 develop = merge-base = `a851b205ccb525db537733111a3938ddd74ff801`
- 테스트 결과는 M1 편집 직후의 워킹 트리에서 측정했고, 그 트리를 수정 없이 `85f1ba686` 으로 커밋했다. diff 기준 AC 는 커밋 뒤 `85f1ba686` 에서 측정했다.
- 원본 출력: `.moai/reports/t675/run/` (`red-e8-nine.txt`, `green-nine-poisoned.txt`, `green-nine-natural.txt`, `ac-001-002-003-004-005-static.txt`)

### E8 RED (편집 전, HEAD `aead782bc`)

`MOAI_EMBED_CHECK_BIN=/usr/bin/false go test -count=1 -v -run '^(TestRunDoctor_(WithExport|WithFix|Verbose|AllFlags|VerboseAndDetail|ExportMode)|TestDoctorCmd_(Execution|ExportFlag|VerboseExecution))$' ./internal/cli/` → exit 1. `--- FAIL` 9줄이 대상 9개 테스트를 각각 가리키고, 모두 `doctor: 1 check(s) failed`. AllFlags 추적에 `✗ Error: could not extract embedded artifacts from /usr/bin/false: false init: exit status 1 ()` 가 찍혔다. 마지막 줄은 `FAIL	github.com/modu-ai/moai-adk/internal/cli	132.787s`.

### AC 판정표

| AC | 상태 | 명령 | 실제 출력 |
|----|------|------|-----------|
| AC-DTC-001 | PASS | `MOAI_EMBED_CHECK_BIN=/usr/bin/false go test -count=1 -v -run '^TestDoctorCmd_Execution$' ./internal/cli/` | exit 0 · `--- PASS: TestDoctorCmd_Execution (11.11s)` · `ok  	github.com/modu-ai/moai-adk/internal/cli	11.723s` · `--- FAIL` 0줄 |
| AC-DTC-002 | PASS | E-RED-002 명령을 그대로 재실행 + `go test ./internal/cli -list '^(TestRunDoctor_(…)|TestDoctorCmd_(…))$'` | exit 0 · `--- PASS:` 9줄(대상 9개) · `--- FAIL` 0줄 · `ok  	github.com/modu-ai/moai-adk/internal/cli	113.530s` · AllFlags 에서 `✓ Agent Emit Embed`; `-list` 는 exit 0 으로 9개 이름을 모두 나열 |
| AC-DTC-003 | PASS | `git diff --name-only develop...HEAD -- '*.go'` | exit 0 · `internal/cli/coverage_improvement_test.go` / `internal/cli/doctor_test.go` / `internal/cli/integration_test.go` (정확히 3줄) |
| AC-DTC-004 | PASS | `git diff -U0 --no-color develop...HEAD -- internal/cli/coverage_improvement_test.go internal/cli/doctor_test.go internal/cli/integration_test.go` | exit 0 · 제거된 내용 줄 0개 · 추가된 내용 줄 정확히 9개, 모두 `+	t.Chdir(t.TempDir())` |
| AC-DTC-005 | PASS | AC-DTC-004 와 같은 명령 | `os.Chdir(` 추가 줄 0개. hunk 헤더가 `func TestRunDoctor_WithExport` · `WithFix` · `Verbose` · `AllFlags` · `VerboseAndDetail` · `ExportMode` · `TestDoctorCmd_Execution` · `ExportFlag` · `VerboseExecution` 을 각각 한 번씩 가리킨다 |

### 일반 환경 회귀 (오염 주입 없음)

`go test -count=1 -v -run '^(…9개 앵커…)$' ./internal/cli/` → exit 0 · `--- PASS:` 9줄 · `--- FAIL` 0줄 · `ok  	github.com/modu-ai/moai-adk/internal/cli	107.697s`

### 정적 검사

- `gofmt -l <세 파일>` → exit 0, 출력 없음
- `go vet ./internal/cli/` → exit 0, 출력 없음
- `golangci-lint run ./internal/cli/` → exit 1, staticcheck 2건: `internal/cli/gtd_answer.go:84:15 S1038`, `internal/cli/launcher.go:811:3 S1021`. 둘 다 이 카드의 diff 밖에 있는 파일이므로(AC-DTC-003) 기존 결함이며, **새로 생긴 lint 는 0건**이다.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-09-18
run_commit_sha: 85f1ba686
run_status: audit-ready
ac_pass_count: 5
ac_fail_count: 0
preserve_list_post_run_count: 0   # 비테스트 Go 파일 변경 0건 (AC-DTC-003)
l44_pre_commit_fetch: not-run     # 리드 일괄 push 체계라 레인은 fetch/push 하지 않는다
l44_post_push_fetch: not-applicable
new_warnings_or_lints_introduced: 0
cross_platform_build:
  darwin: not-run-separately      # 세 test 파일만 바뀌었고, go vet 로 패키지 컴파일을 확인했다
  windows: not-run                # CI(origin/develop) 소관
total_run_phase_files: 4          # test 파일 3개 + spec.md 프런트매터
m1_to_mN_commit_strategy: "M1 = test isolation + status draft->in-progress; evidence commit follows"
```

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: 2026-09-18
sync_commit_sha: pending-backfill-sync
sync_status: audit-ready
spec_status_transition: "in-progress -> implemented -> completed"   # spec.md status 만 변경, updated 는 이미 2026-09-18
b12_self_test_a: "grep -c 'SPEC-DOCTOR-TEST-CWD-ISOLATION-001' CHANGELOG.md -> 0 (emission 전)"
b12_self_test_b: "grep -oE 'AC-([A-Z0-9]+-)*[0-9]+' spec.md | sort -u | wc -l -> 5 (AC-DTC-001..005; Tier S 는 AC 가 spec.md §3 인라인) == CHANGELOG 5"
b12_self_test_c: "CHANGELOG 인용 경로 3개 test 파일 + spec.md 모두 존재 확인"
changelog_entry_position: "[Unreleased] > ### Fixed 첫 항목"
docs_sync: "README / docs-site 변경 없음 — test-only, 사용자 대면 동작 변화 없음"
mx_validation: "프로덕션 코드 변경 0; 세 test 파일의 추가 9줄은 t.Chdir(t.TempDir()) 뿐이라 추가할 @MX 태그 없음"
evidence_paths:
  - .moai/reports/t675/run/red-e8-nine.txt
  - .moai/reports/t675/run/green-nine-poisoned.txt
  - .moai/reports/t675/run/green-nine-natural.txt
  - .moai/reports/t675/run/ac-001-002-003-004-005-static.txt
tests_rerun_in_sync: false   # run-phase 증거(§E.2)와 오케스트레이터의 229971c1c 재검증을 인용
```
