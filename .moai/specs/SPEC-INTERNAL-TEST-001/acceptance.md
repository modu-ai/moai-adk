---
id: SPEC-INTERNAL-TEST-001
version: "0.1.0"
status: in-progress
created: 2026-07-08
updated: 2026-07-08
---

# SPEC-INTERNAL-TEST-001 — 인수 기준 (acceptance)

## §D AC Matrix — 전 항목 기계 검증 가능 (test-suite output 기준)

| AC ID | REQ | 검증 명령 (verbatim) | PASS 판정 |
|-------|-----|----------------------|-----------|
| AC-TEST-001 | REQ-TEST-001 | `go test ./internal/... -count=1` | exit 0; 출력에 `FAIL` 라인 0건, `panic:` 0건 (§D.3 flake 프로토콜 적용) |
| AC-TEST-002a | REQ-TEST-002 | `go test -run TestRunHookEvent_ReadInputError -count=1 ./internal/cli/` | exit 0, PASS — 테스트가 fail-open 계약(`err == nil` + default `HookOutput`)을 단언 |
| AC-TEST-002b | REQ-TEST-002 | `go test -cover -count=1 ./internal/cli/` | `ok ... coverage: N% of statements` 출력 — SIGSEGV 없이 커버리지 수치 보고 |
| AC-TEST-003a | REQ-TEST-003 | `go test -run TestAuthoringDocHasEffortMatrix -count=1 ./internal/cli/agentlint/` | exit 0, PASS |
| AC-TEST-003b | REQ-TEST-003 | `awk '/expectedAgents := \[\]string\{/,/\}/' internal/cli/agentlint/agent_lint_test.go \| grep -cE 'expert-\|researcher\|manager-strategy\|manager-quality\|manager-project\|manager-cycle\|builder-platform'` | `0` (expectedAgents 블록 내 archived/stale 명칭 0건) |
| AC-TEST-003c | REQ-TEST-003 | `diff internal/template/templates/.claude/rules/moai/development/agent-authoring.md .claude/rules/moai/development/agent-authoring.md` | 출력 없음 (양 트리 동일) — agent-authoring.md가 편집된 경우에만 적용 |
| AC-TEST-004a | REQ-TEST-004 | `go test -run TestCloseSubjectDoctrineAmendment -count=1 ./internal/spec/` | exit 0, 두 subtest(lifecycle-sync-gate.md / spec-frontmatter-schema.md) 모두 PASS |
| AC-TEST-004b | REQ-TEST-004 (C-2) | `ls internal/template/templates/.claude/rules/moai/workflow/lifecycle-sync-gate.md` | `No such file or directory` (템플릿 미러 미생성 유지) |
| AC-TEST-005a | REQ-TEST-005 | `go test -cover -count=1 ./internal/migration/` | `ok`, coverage ≥ 85.0% |
| AC-TEST-005b | REQ-TEST-005 | `grep -rn 't.Skip("waiting for migration package implementation")' internal/migration/` | 매치 0건 |
| AC-TEST-006a | REQ-TEST-006 | `ls internal/constitution/canary_test.go internal/constitution/contradiction_test.go` | 두 파일 모두 존재 |
| AC-TEST-006b | REQ-TEST-006 | `go test -cover -count=1 ./internal/constitution/` | `ok`, coverage ≥ 85.0% |

## §D.1 Severity / Traceability

- **Binding gate (MUST-PASS)**: AC-TEST-001 (headline). 이 항목이 FAIL이면 다른 AC 점수와 무관하게 SPEC 미완.
- **Per-finding gates (MUST-PASS)**: AC-TEST-002a..006b — 각 발견사항 F1-F5에 1:1 traceable (spec.md §A 표 참조).
- **Constraint gates**: AC-TEST-003c (Template-First), AC-TEST-004b (no-mirror) — 위반 시 해당 milestone 재작업.

## §D.2 Given-When-Then 시나리오

### 시나리오 1 — headline green (REQ-TEST-001)

- **Given** M1-M5 milestone이 모두 완료된 repository HEAD
- **When** `go test ./internal/... -count=1`을 실행하면
- **Then** exit code 0으로 완주하고, 출력에 `FAIL` 패키지와 `panic:` 스택 트레이스가 존재하지 않는다

### 시나리오 2 — fail-open 계약 단언 (REQ-TEST-002)

- **Given** `deps.HookProtocol.ReadInput`이 `errors.New("invalid JSON")`을 반환하도록 mock된 상태
- **When** `post-tool` 서브커맨드의 `RunE`가 실행되면
- **Then** 반환 오류는 `nil`이고(fail-open), default `HookOutput`(`{}`)이 출력 경로에 기록되며, 테스트는 이를 단언하고 PASS한다 (구 계약 `err != nil` 단언 및 무조건 `err.Error()` 호출은 제거됨)

### 시나리오 3 — AC-DLC-011 oracle 매치 (REQ-TEST-004)

- **Given** `lifecycle-sync-gate.md`에 close-subject full-ID mandate amendment가 삽입된 상태
- **When** `drift_doctrine_test.go`의 oracle 정규식(같은 줄 `(combined|abbreviated)[^\n]*scope` + prohibition token, `(?s).{0,400}?` 양방향)을 적용하면
- **Then** 최소 한 방향이 매치되어 subtest가 PASS한다

### 시나리오 4 — migration crash-recovery (REQ-TEST-005, edge)

- **Given** `t.TempDir()` 내에 in-flight marker가 존재하는(직전 크래시 시뮬레이션) 마이그레이션 상태
- **When** runner가 재실행되면
- **Then** `detectInFlightState`가 상태를 감지하고 `cleanupInFlightState` 경로가 수행되며, 최종 상태는 idempotent하게 수렴한다 (동일 실행 반복 시 결과 불변)

## §D.3 Edge Cases / 재실행 프로토콜 (statusline pre-existing flake)

- AC-TEST-001 실행 중 `internal/statusline` `TestBuild_WritesContextUsageWithSessionID`가 단독 FAIL하는 경우(`context_window_size=1000000, want 256000` — env/state flake, `project_ac_css_001_rescan_debt.md` 기록 pre-existing 부채):
  1. `go test -count=1 ./internal/statusline/`을 1회 재실행한다.
  2. 재실행 PASS → AC-TEST-001은 PASS로 판정하되, flake 발생 사실을 progress.md §E.2 증거에 residual debt로 기록한다.
  3. 재실행에서도 결정론적 FAIL → 신규 회귀 가능성으로 간주, blocker report 반환 (본 SPEC scope 재판단은 orchestrator 소관).
- migration 테스트는 Windows 분기(`version_windows.go`)를 포함하므로, run-phase에서 `GOOS=windows GOARCH=amd64 go build ./internal/migration/` 컴파일 무결을 함께 확인한다 (실행은 darwin/linux 한정).
- 커버리지 판정은 `go test -cover` 출력의 verbatim 수치를 인용한다 — 표만들기용 추정치 금지 (coverage-audit 교훈).

## §D.4 Definition of Done

- [ ] AC-TEST-001..006b 전 항목 PASS (검증 명령 verbatim 출력 증거를 progress.md §E.2에 기록)
- [ ] C-1..C-5 제약 위반 0건 (production 코드 diff 없음: `git diff --stat -- internal/cli/hook.go internal/migration/*.go internal/constitution/canary.go internal/constitution/contradiction.go` → 무변경; 단 `*_test.go` 제외)
- [ ] 신규/변경 테스트에 `t.Skip` placeholder 미도입
- [ ] lint clean: `golangci-lint run` 신규 위반 0건
- [ ] 커밋은 specific-path로 한정 (병렬 세션 레이스 방어)

## §D.5 Residual Risk

- statusline env-flake는 본 SPEC 이후에도 잔존 (out of scope — 별도 후속 소관)
- migration/constitution production 코드의 잠재 결함은 테스트 작성으로 **노출**될 수 있으나 수정은 후속 SPEC 소관 (C-3)
