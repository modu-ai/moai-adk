# SPEC-WEB-CONSOLE-REDESIGN-001 — Progress

## §E.1 Plan-phase Audit-Ready Signal

plan_status: audit-ready
plan_complete_at: 2026-08-08
tier: L
artifacts: spec.md, plan.md, acceptance.md (Tier L 편차 — plan.md §G.1 참조)
open_clarifications: 0 (D3=(c) 제거, D4=(i) 2-option radio + __present 보존 확정)

## §E.2 Run-phase Evidence

M1-M3 완료. M4(서드파티 LLM 탭) / M5(프로필 UI) / M6(autonomy stub) / M7(횡단 마감)은
별도 위임 소관이라 이 구간에서 착수하지 않았다.

| AC | 상태 | 검증 명령 | 실측 |
|----|------|-----------|------|
| AC-WCR-001 | PASS | `go test ./internal/web/ -run TestDeadConfigAbsentFromSchema` | ok — `token_budget.{plan,run,sync}` 부재 |
| AC-WCR-002 | PASS | `go test ./internal/web/ -run TestDeadConfigAbsentFromSchema` | ok — `auto_clear.*` 4종 부재 |
| AC-WCR-003 | PASS | `go test ./internal/config/ -run TestWorkflowBackwardCompat` | ok — 두 블록 담은 yaml 로드 성공, 접근자 4종 디스크 값 반환 |
| AC-WCR-004 | PASS | `go test ./internal/web/ -run TestGitStrategyRendered` | ok — mode + 3× merge_method 컨트롤 렌더 |
| AC-WCR-010 | PASS | `go test ./internal/web/ -run 'TestConsoleTabsOrder\|TestTabPanelRenderOrderMatchesTabs'` | ok — 9탭 순서 + 패널 렌더 순서 일치 |
| AC-WCR-011 | PASS | `go test ./internal/web/ -run TestWorkflowTabRetainsProseConsumedFields` | ok — 산문 소비 6필드 잔류 |
| AC-WCR-012 | PASS | `go test ./internal/web/ -run TestGitWorktreeTabFields` | ok — git_strategy 4 + worktree 4 + branch_guard 1 |
| AC-WCR-013 | PASS | `go test ./internal/web/ -run TestAuditTabFields` | ok — 4필드 렌더 + Section/Persist 무변경 |
| AC-WCR-020 | PASS | `go test ./internal/web/ -run 'TestBoolFieldsRenderAsRadio\|TestCountCheckboxInputsIsFalsifiable'` | ok — checkbox 0개, bool마다 radio 2개. 검출기 falsifiability는 합성 fixture로 별도 증명 |
| AC-WCR-021 | PASS | `go test ./internal/web/ -run 'TestSchemaTogglePresentCompanion\|TestParseSchemaFormBoolSemanticsUnchanged'` | ok — `__present` 보존, 파서 3분기 동작 동일 |
| AC-WCR-022 | PASS | `go test ./internal/web/ -run 'TestClosedSetFieldsAreClosedWidgets\|TestClosedSetRejectsOutOfSetValue'` | ok — 14필드 select/radio + Validate, 집합 밖 값 거부 |
| AC-WCR-023 | PASS-WITH-DEBT | `go test ./internal/web/ -run 'TestFreeTextWhitelist\|TestM4PendingFreeTextStillPending'` | ok — 잔류 자유 텍스트는 화이트리스트 + M4 미착수 4종(`llm.glm.models.*`). 화이트리스트에 SPEC 열거 누락분 2종(`user_name`, `security.sandbox.docker_image`) 추가 — 둘 다 진정 열린 값 |
| AC-WCR-060 | PASS | `go test ./internal/web/ -run TestI18n -v` + `git diff --stat <base>..HEAD -- internal/web/i18n_untranslated_allowlist_test.go` | 14/14 PASS, allowlist diff 0바이트 (신규 예외 0건). 신규 키 44종/로케일 |
| AC-WCR-061 | 공백 통과 | `git diff --name-only <base>..HEAD -- internal/template/ .moai/config/sections/` | 출력 없음 — yaml/템플릿 편집 0건이므로 미러·`make build` 의무 미발생 |
| AC-WCR-062 | 공백 통과 | (AC-WCR-061과 동일 근거) | 템플릿 미러 미갱신 → 중립성 가드 대상 없음 |
| AC-WCR-063 | FAIL (M7 소관) | `go test -cover ./internal/web/...` | 65.4% (< 90.0%). 베이스라인 62.2% 대비 +3.2pp. 90% 달성은 M7 횡단 마감 항목 |

품질 게이트:

- `go build ./...` → exit 0
- `GOOS=windows GOARCH=amd64 go build ./...` → exit 0
- `go vet ./...` → exit 0
- `golangci-lint run --timeout=5m` → `0 issues.`
- `go test ./...` → `internal/template` 3건 실패(`TestLateBranchTemplateMirror` /
  `TestRuleTemplateMirrorDrift` / `TestSanitizedPairParity`). 전부 PRE-EXISTING:
  변경분을 `git stash` 한 베이스라인에서 동일하게 실패함을 실측 확인. 신규 실패 0건
- `templ generate` 재실행 후 `*_templ.go` 해시 무변경 (생성 산출물 동기화 완료)

### §E.2.1 `workflow.execution_mode` 닫힌 집합 정정 (M3 잔여 항목 해소)

M3는 `config.ValidExecutionModes()`를 `{auto, solo, team}`으로 선언했으나 이 집합은
**추론**이었고 SSOT 근거가 없었다. M4-M6 구간에서 실측한 결과 `cg`가 누락된 4번째
값임이 확인됐다.

| 근거 | 관측 |
|------|------|
| `.moai/config/sections/harness.yaml` `mode_defaults` 키 | `solo` / `team` / `cg` (3개) |
| `internal/config/types.go:868` `ModeDefaults` 주석 | "default harness level map per execution mode (solo/team/cg)" |
| `SPEC-V3R2-HRN-001` REQ-HRN-001-014 | "**While** `workflow.yaml execution_mode` is `auto`, the router **shall** consult `cfg.ModeDefaults.solo\|team\|cg`" |
| Go 런타임 reader | 0건 (`ExecutionMode` struct 멤버 + 기본값만; 소비는 산문) |

M3가 이 필드를 닫힌 위젯으로 전환하면서 집합 밖 값의 저장이 거부되므로, 누락은
표시 문제가 아니라 **문서화된 모드 하나를 사용자가 설정할 수 없게 만드는 회귀**였다.

처분: `ValidExecutionModePins()` = `{solo, team, cg}`를 신설하고
`ValidExecutionModes()` = `auto` + pins로 구성했다. `harness.mode_defaults.*`
FieldDef 목록도 같은 접근자에서 파생하도록 바꿔 두 집합이 구조적으로 드리프트할 수
없게 했다 (`modeDefaultFields()`). `internal/config/execution_modes_test.go`가 배포되는
`harness.yaml`의 `mode_defaults` 키를 실제로 파싱해 pin 집합과 대조한다 — 추론이 아닌
산출물 대조다.

## §E.3 Run-phase Audit-Ready Signal

run_status: partial (M1-M3 완료, M4-M7 미착수)
run_complete_at: 2026-08-08
run_commit_sha: c9406abf5
ac_pass_count: 13
ac_fail_count: 1 (AC-WCR-063 커버리지 — M7 소관)
ac_vacuous_count: 2 (AC-WCR-061 / AC-WCR-062 — yaml·템플릿 편집 0건)
preserve_list_post_run_count: 5/6 (plan.md §A.3). 6번 항목("미렌더 섹션 중
  git_strategy 외 나머지의 FieldDef 무접촉")은 의도적 편차다: REQ-WCR-022가
  `harness.*`를 닫힌 집합 전환 대상으로 명시하므로 harness FieldDef 8종의
  Type/Options/Validate를 변경했다. 이름·섹션·영속화 경로는 무변경이고 harness
  섹션은 여전히 미렌더다. 나머지 5항목(workflow.yaml 키, config struct+접근자,
  `parseSchemaForm` bool 분기, glmkey.go 계약, board.*)은 diff 0라인으로 실측 확인.
new_warnings_or_lints_introduced: 0
cross_platform_build: darwin/arm64 PASS, windows/amd64 PASS
total_run_phase_files: 15
m1_to_mN_commit_strategy: 밀스톤당 1커밋 (M1 / M2 / M3), push 없음

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
