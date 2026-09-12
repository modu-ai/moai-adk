# SPEC-INIT-QUIET-WIZARD-001 — 구현 계획

## §A 맥락

init 위저드를 18문항에서 4문항으로 줄이고, 제거한 키는 기본값에 맡긴다. 그 성질을 실제 `runInit` 실행 테스트로 봉인한다. 실행 테스트가 실제 홈의 셸 설정 파일에 쓰지 않도록, init 흐름의 셸 설정 단계를 교체 가능한 테스트 시접을 거쳐 실행하게 하는 변경을 가장 먼저 착지시킨다. 테스트는 이 시접에 스파이를 끼워 쓰기를 막는 동시에 단계 도달을 관측한다. init 실행 테스트를 돌리는 모든 run 슬롯은 실행 전후에 실제 홈 지문을 잰다(spec.md §4.3). 근거 조사는 research.md, 설계 대안은 design.md 에 있다. 1회차 plan 감사 결과는 `.moai/reports/t583/plan-audit.md` 다.

## §B 확정 결정 — 가장 바뀌기 쉬운 결정부터

리뷰는 이 절에 먼저 집중한다. 아래 결정이 뒤집히면 마일스톤 구성이 바뀐다.

| # | 결정 | 근거 | 뒤집힐 때 영향 |
|---|---|---|---|
| B1 | 질문 생성자를 나눈다(D1). `DefaultQuestions` 는 손대지 않고, `InitQuestions` 가 ID 로 골라 4문항을 조립한다 | reconfigure 12문항 보존 | 합치면 reconfigure 에서 3문항이 빠진다 |
| B2 | 대화형 MCP 기본값은 `internal/cli/init.go` 대화형 블록에서 `opts.MCPProvision = true` 로 둔다(D2). `WizardResult.MCPProvision` 필드는 지운다 | 기본값이 위저드 결과에 기대면, 시접을 주입한 테스트와 운영 경로가 갈라진다 | 위저드 시드로 옮기면 주입 테스트가 운영 동작을 재지 못한다 |
| B3 | 셸 설정 단계 시접은 `internal/core/project` 의 내보낸 함수 변수 하나다(`ConfigureShellEnvFn` 형태, 이름은 run 단계 확정). Step 6 이 이 변수를 거쳐 셸 설정을 기록하고, 운영 기본값은 현재 `configureShellEnv` 본문과 같은 함수다. 사용자 플래그·환경 변수로 만들지 않는다 | 리드 결정: 도달은 실행으로 관측한다. 스파이를 끼운 실제 `runInit` 이 단계 호출 횟수를 직접 보여 준다 | 소스 문자열 가드로 돌아가면 우회 변형을 막지 못한다 |
| B4 | 도달 불가가 되는 항목은 §G 규칙(다른 호출자가 닿으면 유지, 아니면 삭제)으로 판정한다 | 제거 선례 `question_removal_test.go`(REQ-WIZ-014·017) | 전부 유지로 가면 죽은 기록기와 테스트가 남는다 |
| B5 | 남는 4문항의 그룹 라벨은 현행 유지(Basic·Basic·Quality & Workflow·Autonomy) | 렌더링은 t586 소관 | 라벨을 바꾸면 t586 과 충돌 |
| B6 | `saveBoolAnswer` 와 `buildConfirmField` 는 남긴다. `saveBoolAnswer` 는 분기가 모두 사라져 빈 함수가 된다 | `QuestionTypeConfirm` 열거값이 남아 있고, 이를 지우는 일은 렌더링 구간(t586)에 닿는다. 선례 주석(`wizard.go:458-466`)도 같은 이유로 유지했다 | 지우면 t586 변경 구간과 겹친다 |
| B7 | 홈 안전 점검표는 새로 쓰거나 본문을 다시 쓰는 init 실행 테스트의 코드에만 넣는다. 필드 참조만 지우는 기존 테스트는 기존 HOME 헬퍼를 유지하고, 모든 실행 슬롯은 실제 홈 지문 절차로 덮는다(리드 결정 (a), 2026-09-11) | 공용 헬퍼를 고치면 변경 파일이 크게 늘어난다. 기존 헬퍼의 안전성은 판독뿐이므로 실행 쪽에서 잰다 | 전부 옮기면 공용 헬퍼와 그 호출 테스트 전체가 이 카드 범위에 들어온다 |
| B8 | `internal/cli` 쪽 셸 설정 억제 시접(`runInit` 이 `SkipShellConfig` 를 켜게 하는 변수)은 두지 않는다. B3 의 스파이 시접 하나가 쓰기 억제와 도달 관측을 모두 맡는다 | 스파이가 실제 기록 함수를 대신하므로 쓰기가 일어날 수 없고, 호출 횟수로 도달이 보인다. 두 번째 시접은 `runInit` 에 운영 코드 변경만 늘린다 | 두 시접으로 가면 `init.go` 배선 변경과 그 배선을 증명할 추가 관측이 필요하다 |

## §C 착수 전 점검 (run 단계 시작 시)

1. 분기 재측정: `git fetch origin develop && git rev-list --count --left-right origin/develop...HEAD`. plan 시점 첫 측정은 `17 1` 이었고 변경 구간의 `git diff --stat HEAD...origin/develop -- internal/cli/wizard internal/cli/init.go internal/cli/update_wizard.go internal/core/project internal/shell` 은 비어 있었다. v0.1.3 직전 재측정(2026-09-11)은 `118 1` 이었고, 같은 변경 구간에서 develop 이 바꾼 파일은 `internal/cli/update_wizard.go` 하나(t587 `c4990eea7`)였다(둘 다 progress.md §E.1). 흡수는 통합 창에서 한다. 흡수 뒤에도 AC-IQW-003 의 카드 범위 판정은 흡수한 `develop` 과의 merge-base 부터 재므로 그 변경 때문에 빨개지지 않는다. <!-- moving-ref-ok: pre-flight divergence check measures the current mainline itself (subject, not anchor); the plan-time value is a dated reference in progress.md §E.1 -->
2. t586(lane-2) 진행 상태 확인: 같은 위저드 파일을 다루므로, §F.1 변경 구간을 lane-2 에 알린다(리드 경유).
3. §G 의 각 삭제 항목에 대해 표에 적힌 `git grep` 을 다시 돌려, 삭제 직전에도 호출자가 없음을 확인한다.
4. 실제 홈 지문 절차 준비(spec.md §4.3, REQ-IQW-014·016): `./internal/cli` 또는 `./internal/core/project/...` 에 대한 `go test` 호출은 모두 슬롯이다. 슬롯마다 progress.md §E.2 슬롯 표에 `SLOT-<번호>` 행을 먼저 선언하고, acceptance.md AC-IQW-015 의 기준 명령으로 실행 전 지문을 뜨고, 테스트를 돌리고, 같은 명령 본문으로 실행 후 지문을 뜬 뒤 diff 한다. 두 지문·명령·종료 코드는 그 행에 남긴다. 차이가 나거나 표준 오류가 비어 있지 않으면 멈추고 리드에게 보고하며, 실제 홈 파일을 되돌리지 않는다. M1 의 첫 `go test` 가 첫 슬롯이다.

## §D 제약

- 검증은 영향 패키지로 한정한다: `./internal/cli/wizard/...`, `./internal/cli`(`-timeout 600s`), `./internal/core/project/...`. 전체 스위트는 CI 에 맡긴다.
- 바이너리로 `moai init`·`moai update` 를 실행하지 않는다.
- 이 SPEC 이 새로 쓰거나 본문을 다시 쓰는 init 실행 테스트는 spec.md §4.2 점검표를 코드에 모두 갖추고, `t.Parallel` 을 쓰지 않는다.
- 제거된 필드 참조만 지우는 기존 init 실행 테스트는 기존 `t.Setenv("HOME")` 헬퍼를 그대로 두며, 헬퍼를 고치지 않는다(§N 후속 후보).
- `./internal/cli` 또는 `./internal/core/project/...` 를 대상으로 하는 모든 `go test` 슬롯은 §C-4 의 실제 홈 지문 절차를 따른다. 실행 전후 지문은 같은 명령 본문으로 뜨고 표준 오류를 분리한다.
- 새 테스트를 겨누는 `go test -run` 은 `-v` 로 돌리고, 지목한 테스트마다 `--- PASS:` 줄이 있고 `[no tests to run]` 이 없는지 확인한다(`.claude/rules/moai/development/verification-completeness.md` §1.1).
- 셸 설정 단계 시접을 우회해 실제 기록 함수를 직접 부르는 뮤턴트는 실제 셸 설정 파일에 쓰므로 만들지도 돌리지도 않는다.
- 위저드 파일 변경은 §F.1 구간 밖으로 나가지 않는다.

## §E 자기 검증 명령

```bash
go test ./internal/cli/wizard/... -count=1
go test ./internal/core/project/... -count=1
go test ./internal/cli -count=1 -timeout 600s
go vet ./internal/cli/... ./internal/core/project/...
golangci-lint run ./internal/cli/... ./internal/core/project/...
```

둘째·셋째 명령은 §C-4 슬롯이다(선언과 지문 전후 채집). AC 별 명령은 acceptance.md 에 있다.

## §F 마일스톤 (의존 순서)

**착지 순서.** M1 의 셸 설정 단계 시접과 그 관측 테스트는 M2 의 첫 init 실행 테스트보다 먼저 커밋한다. 이 순서는 REQ-IQW-011 의 선행 조건이며, 이 계획의 마일스톤 순서로 지킨다(AC 증거로 쓰지 않는다).

### M1 — 셸 설정 단계 시접 + 스파이 관측 + 홈 안전 헬퍼 (우선순위 High, 선행 조건)

- [MODIFY] `internal/core/project/initializer.go`: Step 6(`:333-347`)이 셸 설정을 기록할 때 내보낸 함수 변수(`ConfigureShellEnvFn` 형태)를 거쳐 호출하게 한다. 운영 기본값은 현재 `configureShellEnv`(`:677`) 본문과 같은 함수(`shell.NewEnvConfigurator(logger).Configure(...)`)여서 동작은 바뀌지 않는다. `!opts.SkipShellConfig` 게이트는 그대로 둔다. 형태와 이유는 design.md §4.
- [NEW] `internal/core/project` 테스트(예: `initializer_shell_seam_test.go`):
  - 보조 — 시접의 기본값이 운영 기본 함수와 같은 함수임을 직접 읽어 확인(`reflect` 함수 포인터 비교).
  - 게이트 — 스파이를 끼운 실제 `Init` 실행에서 `SkipShellConfig=true` 면 스파이 호출 0회, `false` 면 1회. 시접을 바꾸기 전에 `t.Setenv` 로 `MOAI_HOME` 을 `t.TempDir()` 아래로 돌리고, `t.Cleanup` 으로 원복하며, `t.Parallel` 을 쓰지 않는다(design.md §4.2).
- [NEW] `internal/cli/init_home_guard_test.go`: 세 홈 우회(`MOAI_HOME` 은 `t.Setenv`) + 셸 설정 단계 스파이 설치·원복(설치 전에 원래 값과 같은 함수인지 `reflect` 포인터로 단언) + 실행 전 자기 가드 + 전후 대조 목록 8항목(`~/.claude/settings.json` sha256, `~/.claude/hooks/moai` 존재, `~/.zshenv`·`~/.zshrc`·`~/.zprofile`·`~/.profile`·`~/.bashrc`·`~/.bash_profile` 의 mtime·sha256 — spec.md §4.2). 가드 판정은 오류를 돌려주는 함수로 두어 가드 자체를 테스트한다.
- [NEW] `internal/cli` 주 관측 테스트(예: `init_shell_seam_test.go`): 홈 안전 헬퍼를 갖춘 채 스파이를 끼우고 실제 `runInit` 을 실행해 스파이 호출이 정확히 1회임을 단언.
- [RECORD] 배선 제거 뮤턴트 2종(게이트 반전, Step 6 호출 삭제)을 적용해 주 관측 테스트가 RED 가 되는 출력을 progress.md §E.2 에 남긴다. 뮤턴트는 커밋하지 않는다.
- 파일: `internal/core/project/initializer.go`, `internal/core/project/initializer_shell_seam_test.go`(새), `internal/cli/init_home_guard_test.go`(새), `internal/cli/init_shell_seam_test.go`(새). `internal/cli/init.go` 는 M1 에서 바꾸지 않는다.
- AC: AC-IQW-004, AC-IQW-005, AC-IQW-015(이 마일스톤의 `go test` 슬롯), AC-IQW-016(시접 대입 테스트의 병렬 금지·원복, 뮤턴트 C1·C2·D 기록).

### M2 — RED: init 실행 테스트 (우선순위 High)

- [NEW] `internal/cli/init_quiet_wizard_test.go`: 유지 4문항 답 + 운영 시드만 주입한 대화형 `runInit`, 비대화형 `runInit` 기준 실행, 해석값 표 대조, 섹션 파일 바이트 비교, MCP 기본값(claude·codex·both), `--project-mode` 플래그 기록, 관측기 음성 대조군. 모든 실행에 홈 안전 헬퍼(스파이 포함)를 쓴다.
- 현재 코드에서 기대되는 RED: 대화형 실행이 `workflow.audit` 아래 `model`·`gates` 키와 `workflow.todo.enabled` 를 써서 섹션 파일이 비대화형과 달라진다. 주입 결과에 `MCPProvision` 이 없으면 안내 문구가 나오지 않는다. RED 출력은 progress.md §E.2 에 그대로 남긴다.
- AC: AC-IQW-006~009, AC-IQW-011(신규 테스트 부분), AC-IQW-015.

### M3 — 위저드 패키지: 생성자 분할과 질문 제거 (우선순위 High)

- [MODIFY] `InitQuestions` 가 `DefaultQuestions` 에서 `conversation_language`·`user_name` 을 ID 로 골라 담고, 이어 `Page3Questions` 를 붙인다. `Page3Questions` 는 `agent_wiring`·`autonomy_tier` 만 돌려준다(리터럴은 그대로 둔다).
- [REMOVE] 페이지 3 전용 11문항의 정의, `saveAnswer`·`saveBoolAnswer` 의 해당 분기, `WizardResult` 의 해당 필드, ko/ja/zh 번역 33항목.
- [MODIFY] 고정 테스트 갱신(§H 위저드 부분). 제거 선례에 따라 제거 목록 테스트를 넓힌다.
- AC: AC-IQW-001~003.

### M4 — init 배선: 매핑 삭제, 대화형 MCP 기본값, 주석 정리 (우선순위 High)

- [MODIFY] `applyWizardPage3ToOpts`(`internal/cli/init.go:277-339`)에서 제거 질문의 매핑(`project_mode`, 워크트리, todo, feedback, continuation, audit 6개와 `AuditConfigSet`, `MCPProvision`)을 지운다. LSP·품질·디자인 시드 매핑은 남긴다.
- [MODIFY] 위저드 결과 적용 블록(`init.go:728-750`)의 `ProjectName`·`ModelPolicy`·`ReportFormat` 매핑을 지운다. init 위저드가 더는 이 값을 채우지 않는다.
- [MODIFY] 대화형 블록에 `opts.MCPProvision = true` 를 둔다. `init.go:1004-1011` 의 `codex`/`both` 규칙과 그 위 주석(결정 B1 서술)을 새 규칙에 맞게 고친다.
- [MODIFY] 워크트리 기록 규칙 주석 정리: `internal/core/project/initializer.go:57-62`, `internal/cli/init.go:300-304`(매핑과 함께 사라짐), `questions.go:372`(질문과 함께 사라짐). 남는 주석은 REQ-IQW-008 의 한 규칙만 서술한다.
- [MODIFY] CLI 고정 테스트 갱신(§H CLI 부분). M2 가 GREEN 이 된다.
- AC: AC-IQW-006~012, AC-IQW-015.

### M5 — `internal/core/project` 죽은 기록기 정리 (우선순위 Medium)

- [REMOVE] §G 에서 "삭제"로 판정한 기록기·`InitOptions` 필드·호출부·그 단위 테스트.
- AC: AC-IQW-013, AC-IQW-015.

### M6 — 마감 검증과 뮤턴트 기록 (우선순위 Medium)

- §E 명령 전부, 코드 뮤턴트 2종(acceptance.md AC-IQW-007b) 실행과 원복, 결과를 progress.md §E.2 에 그대로 기록. 뮤턴트 실행과 마감 `go test` 도 모두 §C-4 슬롯이다.
- AC: AC-IQW-007b, AC-IQW-014, AC-IQW-015, AC-IQW-016(마감 시 대입 파일 재스윕과 두 실행 관측).

### §F.1 위저드 변경 구간 — t586 충돌 회피용 명시 목록

이 SPEC 이 `internal/cli/wizard/` 에서 건드리는 곳은 아래뿐이다. 이 밖은 손대지 않는다. v0.1.2 의 셸 설정 단계 시접은 `internal/core/project` 와 `internal/cli` 테스트 파일에만 닿으므로 이 구간은 바뀌지 않았다.

| 파일 | 건드리는 곳 | 건드리지 않는 곳 |
|---|---|---|
| `questions.go` | `InitQuestions`(296-303) 본문, `Page3Questions`(349-527) 에서 11문항 리터럴과 딸린 주석 삭제, ID 로 고르는 작은 도우미 추가 | `DefaultQuestions`(48), `GitQuestions`(162), `ReconfigureQuestions`(268-289), `FilteredQuestions`·`TotalVisibleQuestions`·`QuestionByID`, 남는 두 리터럴(`agent_wiring`·`autonomy_tier`)의 내용과 그룹 라벨 |
| `wizard.go` | `saveAnswer` 의 `project_mode`(434)·`project_continuation`(440)·`audit_model`·`audit_gate_*`(442-449) 분기, `saveBoolAnswer`(467) 본문 분기와 그 위 주석(458-466) | `RunWithDefaults`(35-59, 시드 46-52 포함), `buildFormGroups`(158~) 와 그룹 묶기, 스테퍼(`stepperDenominator` 236, `visibleQuestionIndex` 242), `buildField`·`buildConfirmField`(487~), 모든 스타일·렌더링 코드 |
| `types.go` | `WizardResult` 의 제거 필드 11개(`ProjectMode`, `WorktreeAutoCreate`, `TodoEnabled`, `FeedbackAutoSubmit`, `ProjectContinuation`, `AuditModel`, `AuditGateClaude`, `AuditGateCodex`, `AuditGateGLM`, `CodexAuditEnabled`, `MCPProvision`) | 나머지 필드 |
| `translations.go` | ko(32~)·ja(201~)·zh(369~) 표에서 11개 ID 항목 삭제 | `project_name`·`model_policy`·`report_format`·Git·`agent_wiring`·`autonomy_tier` 항목, 548 이후 UI 문자열 표 |
| 위저드 테스트 | §H 에 나열한 파일 | 렌더링·색·폼 스냅숏 테스트 |

## §G 유지·삭제 판정표

판정 규칙: CLI 플래그나 reconfigure 등 다른 호출자가 닿으면 **유지**, 아니면 **삭제**. 각 항목은 착수 직전 표의 명령으로 다시 확인한다(§C-3). 명령은 셸 `grep` 래퍼 대신 `git grep` 을 쓴다.

| 항목 | 판정 | 근거(plan 시점 판독) | 착수 직전 재확인 |
|---|---|---|---|
| 질문 정의 `project_name`·`model_policy`·`report_format` | 유지 | `ReconfigureQuestions` 가 사용(`update_wizard.go:64`) | `git grep -n 'ReconfigureQuestions(' -- internal ':!*_test.go'` |
| 번역 `project_name`·`model_policy`·`report_format` | 유지 | reconfigure 렌더링 | 위와 같음 |
| `saveAnswer` 분기 `project_name`·`model_policy`·`report_format` | 유지 | reconfigure 가 답을 받음. `ModelPolicy` 는 `update_wizard.go:307-310` 이 읽음 | `git grep -n 'result.ModelPolicy' -- internal ':!*_test.go'` |
| `WizardResult.ProjectName`·`ModelPolicy`·`ReportFormat` | 유지 | reconfigure 가 채움(`ProjectName`·`ReportFormat` 을 저장하지 않는 결함은 t588) | — |
| 질문 정의·번역 11개(페이지 3 전용) | 삭제 | `InitQuestions` 외 사용처 없음 | `git grep -n 'Page3Questions(' -- internal ':!*_test.go'` → `questions.go` 내부만 |
| `saveAnswer` 분기 `project_mode`·`project_continuation`·`audit_model`·`audit_gate_*` | 삭제 | 제거 질문 전용 | — |
| `saveBoolAnswer` 분기 5개 | 삭제 | 제거 질문 전용 | — |
| `saveBoolAnswer` 함수, `buildConfirmField` | 유지(빈 함수) | §B B6 | `git grep -n 'saveBoolAnswer(\|buildConfirmField(' -- internal/cli/wizard ':!*_test.go'` |
| `WizardResult` 제거 필드 11개 | 삭제 | 비테스트 독자가 `init.go:278-337` 매핑뿐 | `git grep -nE 'result\.(ProjectMode\|WorktreeAutoCreate\|TodoEnabled\|FeedbackAutoSubmit\|ProjectContinuation\|AuditModel\|AuditGate\|CodexAuditEnabled\|MCPProvision)' -- internal cmd ':!*_test.go'` |
| `init.go` 매핑 `project_mode`(278-280) | 삭제 | 플래그 경로는 `init.go:609` 에서 직접 `opts.ProjectMode` 를 채움 | — |
| `InitOptions.ProjectMode` + `writeProjectModeYAML` | 유지 | `--project-mode` 플래그(`init.go:92`)가 닿음 | `git grep -n '"project-mode"' -- internal/cli/init.go` |
| `init.go` 워크트리 매핑(305-307) | 삭제 | 위저드 답 소멸 | — |
| `InitOptions.WorktreeAutoCreate`·`WorktreeAutoCreateSet` + `WriteWorkflowTogglesYAML` | 유지 | `--worktree-auto-create`(`init_workflow_flags.go:40-44`) | `git grep -n 'WorktreeAutoCreateSet = true' -- internal ':!*_test.go'` |
| `runWorkflowConfigStep` | 유지 | reconfigure 단계(`update_wizard.go`) | — |
| `init.go` 매핑 todo·feedback·continuation(313·319·325) | 삭제 | 위저드 전용, 플래그 없음 | — |
| `InitOptions.TodoEnabled` + `writeWorkflowTodoYAML` + 호출부(`initializer_expansion.go:50`) | 삭제 | 필드를 채우는 곳이 `init.go:313` 뿐 | `git grep -n 'TodoEnabled' -- internal/core/project internal/cli ':!*_test.go'` (statusline 의 동명 필드는 다른 타입) |
| `InitOptions.FeedbackAutoSubmit` + `writeFeedbackAutoSubmitYAML` + 호출부(`:53`) | 삭제 | 필드를 채우는 곳이 `init.go:319` 뿐 | `git grep -n 'FeedbackAutoSubmit' -- internal ':!*_test.go'` |
| `InitOptions.ProjectContinuation` + `writeWorkflowProjectContinuationYAML` + 호출부(`:56`) | 삭제 | 필드를 채우는 곳이 `init.go:325` 뿐 | `git grep -n 'ProjectContinuation' -- internal ':!*_test.go'` |
| `init.go` audit 매핑 6개 + `AuditConfigSet = true`(332-338) | 삭제 | 위저드 전용 | — |
| `InitOptions` audit 필드 6개 + `AuditConfigSet` + `writeWorkflowAuditYAML` + 호출부(`initializer.go:306`) | 삭제 | `AuditConfigSet` 을 켜는 곳이 `init.go:338` 뿐 | `git grep -n 'AuditConfigSet' -- internal ':!*_test.go'` |
| `InitOptions.MCPProvision` | 유지 | 대화형 기본값을 싣는 곳, `init.go:1004` 가 읽음, `init_mcp_provision_test.go:105` 소스 도달성 가드가 `opts.MCPProvision` 문자열을 요구 | `git grep -n 'opts.MCPProvision' -- internal/cli` |
| `provisionMCPEntryUnlessDeclined` | 유지 | 대화형·`both` 경로 | — |
| `init.go` `ProjectName`·`ModelPolicy`·`ReportFormat` 매핑(728-750) | 삭제 | init 위저드가 더는 채우지 않음. 기본값 경로(`phase.go:195-199`, `init.go:940-953`, `initializer.go:579-593`)가 같은 값을 공급 | — |
| `DevelopmentMode` 매핑과 `saveAnswer` `development_mode` 분기 | 손대지 않음 | 이미 고아였던 기존 항목. 이 카드 범위 밖 | — |
| 번역 완결성 테스트의 옵션 면제 목록(`translations_completeness_test.go:13` `optionTranslationExemptIDs`) 중 제거 ID 항목(있다면) | 삭제 | 면제 목록이 제거 ID 를 가리키면 고아가 됨 | `sed -n '/^var optionTranslationExemptIDs/,/^}/p' internal/cli/wizard/translations_completeness_test.go` |
| 기존 테스트 헬퍼 `runInitForAutonomyAtHomeCapturingOut`(`init_autonomy_wiring_test.go:40`) | 유지(수정 안 함) | 필드 참조만 지우는 기존 테스트가 계속 사용(§B B7). 이관은 §N 후속 후보 | — |
| 셸 설정 단계 시접(`ConfigureShellEnvFn` 형태, 새로 추가) | 유지(새 항목) | Step 6 의 유일한 호출 경로. 운영 기본값이 현재 동작 | `git grep -n 'ConfigureShellEnvFn' -- internal/core/project ':!*_test.go'` (이름 확정 후) |
| `InitOptions.SkipShellConfig` 필드와 Step 6 게이트(`initializer.go:44`, `:333`) | 유지 | 게이트는 AC-IQW-004 게이트 테스트가 실행으로 확인. `runInit` 은 이 필드를 켜지 않으며 켜게 바꾸지도 않는다(§B B8) | `git grep -n 'SkipShellConfig' -- internal ':!*_test.go'` |

## §H 의도적으로 갱신할 고정 테스트

현재 질문 집합을 고정한 테스트는 목적을 확인한 뒤 새 집합에 맞춰 고치거나, 대상이 사라진 경우 지운다. 조용한 삭제는 금지하고, 커밋 메시지에 테스트별 처분을 적는다.

처분의 두 부류를 구별한다. **본문 재작성**은 테스트가 무엇을 단언하는지를 바꾸는 경우이고, 그 테스트가 init 을 실행한다면 spec.md §4.2 점검표를 코드에 갖춘다. **필드 참조 제거**는 사라진 필드를 가리키는 줄만 지우고 단언을 그대로 두는 경우이며, 기존 HOME 헬퍼를 유지한다. 두 부류 모두 실행 슬롯에서는 §C-4 지문 절차를 따른다.

**위저드 패키지 (`internal/cli/wizard/`)** — 이 패키지 테스트는 init 을 실행하지 않는다.

| 테스트 | 현재 고정 내용 | 처분 |
|---|---|---|
| `expansion_test.go:11-69` `TestPage3QuestionsStructure` | 페이지 3 13문항 순서 | 2문항(`agent_wiring`, `autonomy_tier`)으로 갱신 |
| `expansion_test.go:71-263` | 페이지 3 개별 질문 | 제거 질문 부분 삭제 |
| `restructure_test.go:68-90` `TestInitPages_Membership` | 페이지별 구성원 | Basic 2, Quality & Workflow 1(`agent_wiring`), Autonomy 1 로 갱신. Model & Report 페이지는 init 에서 비어 있음을 단언 |
| `restructure_test.go:92-124` `TestInitPages_MergeIntoOneGroupPerPage` | Model & Report 포함 페이지 순회 | init 에 남는 페이지만 순회하도록 갱신 |
| `restructure_test.go:125` `TestPage3_NoModeGate` | `project_mode` 무조건 노출 | 대상 소멸 — 삭제 또는 남는 문항 무조건 노출로 갱신 |
| `wizard_test.go:624-672` `TestStepperTotal_DynamicDenominator` | 분모 19(6 + 페이지 3 13) | 6 + 2 = 8 로 갱신. reconfigure 분모 사례(6·9·10)는 그대로 |
| `agent_wiring_question_test.go:65-93` `TestAgentWiringQuestion_PrecedesMCPProvision` | `mcp_provision` 이 바로 뒤 | `mcp_provision` 부재와 `autonomy_tier` 가 바로 뒤임을 단언하도록 갱신 |
| `mcp_audit_test.go` | audit·codex·MCP 질문 노출 | 삭제 |
| `worktree_test.go`, `todo_enabled_test.go`, `feedback_auto_submit_test.go`, `project_continuation_test.go` | 해당 질문 존재·기본값 | 삭제(대상 소멸) |
| `question_removal_test.go:15-18` `removedInM3` | 선례 제거 목록 | 이번 11개 ID 를 별도 목록으로 추가해 부재·번역 고아·답 저장 분기를 같은 방식으로 단언. 공유 3문항은 init 부재 + reconfigure 존재를 단언 |
| `translations_completeness_test.go:95` `TestWizardQuestionTranslationCompleteness` | `InitQuestions` 전 문항 번역 | 코드 변경 없이 통과해야 함(갱신 대상 아님) |
| `questions_test.go:87`·`:184`·`:316` | `DefaultQuestions` 5문항·reconfigure 12문항 | 변경 금지(AC-IQW-003) |
| `autonomy_test.go` | `InitQuestions` 에 `autonomy_tier` | 변경 없이 통과해야 함 |

**CLI 패키지 (`internal/cli/`)**

| 테스트 | 현재 고정 내용 | 처분 |
|---|---|---|
| `init_audit_test.go:20-66` `TestApplyWizardPage3ToOpts_AuditSelection` | 위저드 audit 선택 → opts, `AuditConfigSet=true` | 삭제 |
| `init_audit_test.go` `…AuditConfigSetFalseByDefault` | `AuditConfigSet` 기본 false | 필드 삭제로 대상 소멸 — 삭제 |
| `init_audit_wiring_test.go:29` `TestRunInit_WizardAuditSelectionPersists` | 위저드 audit 블록 기록 | 삭제 |
| `init_workflow_wiring_test.go:100-134` `TestRunInit_WorktreeAutoCreateFlagBeatsWizard` | 플래그 vs 위저드 답 | 필드 참조 제거 — 주입 결과의 `WorktreeAutoCreate` 필드만 지우고 단언은 그대로. 기존 HOME 헬퍼 유지 |
| `init_workflow_wiring_test.go:83` `TestRunInit_WorkflowToggleFlagsAbsentByteIdentical` | 비대화형 바이트 동일 | 변경 금지(AC-IQW-010) |
| `init_flag_precedence_test.go:60` `TestFlagBeatsWizard_Page3Settings` | `project-mode`·워크트리 위저드 답 | 제거 필드 부분 삭제, LSP·품질·디자인 우선순위는 유지(init 실행 없음) |
| `init_agent_wizard_test.go:64-160` | `WizardResult.MCPProvision` 사용 | 필드 참조 제거. 대화형 기본값으로 같은 결과가 나와야 함. 기존 HOME 헬퍼 유지 |
| `init_agent_wizard_test.go:172-205` | 비대화형 보장 호출 생략 | 변경 금지(AC-IQW-010) |
| `doctor_codex_e2e_test.go:54`·`:86` | `MCPProvision: true` 픽스처 | 필드 참조 제거. 기존 HOME 헬퍼 유지 |
| `init_mcp_provision_test.go:105` `TestRunInit_CallsMCPProvisioning` | 소스에 `opts.MCPProvision` 문자열 | 변경 없이 통과해야 함 |
| 픽스처 사용 파일 `init_wizard_identity_test.go`·`init_gitdetect_test.go`·`init_update_notice_test.go` | 제거 필드를 쓰는 경우 | 컴파일 오류가 나면 필드 참조만 제거. 기존 헬퍼 유지 |

**코어 패키지 (`internal/core/project/`)** — §G 삭제 기록기의 단위 테스트: `initializer_audit_test.go`, `initializer_audit_wiring_test.go`, `initializer_todo_test.go`, `initializer_feedback_test.go`, `project_continuation_write_test.go` 삭제. `initializer_expansion_test.go`·`initializer_persist_test.go` 에서 제거 필드를 쓰는 부분만 정리. `initializer_workflow_toggles_test.go` 는 변경 금지.

## §I mx_plan

| 대상 | 태그 | 내용 |
|---|---|---|
| `internal/core/project/initializer.go` 셸 설정 단계 시접(`ConfigureShellEnvFn` 형태) | `@MX:NOTE` + `@MX:SPEC: SPEC-INIT-QUIET-WIZARD-001` | Step 6 의 셸 설정 기록은 이 변수를 거친다. 운영 기본은 실제 기록. 테스트는 스파이로 바꿔 끼워 실제 셸 설정 파일에 쓰지 않고 호출을 관측한다 |
| 같은 시접 | `@MX:WARN` + `@MX:REASON` | 패키지 전역 변수를 테스트가 바꾼다. 바꾸는 테스트는 `t.Parallel` 금지, `t.Cleanup` 으로 원복. 시접을 우회해 기록 함수를 직접 부르면 테스트가 실제 홈에 쓴다 |
| `internal/cli/init.go` 대화형 `opts.MCPProvision = true` | `@MX:NOTE` + `@MX:SPEC` | 위저드 질문이 없는 대화형 기본값. 비대화형은 0값 false 로 보장 호출을 생략하며 이 비대칭은 의도다 |
| `internal/cli/init.go:1003` 기존 `@MX:SPEC: SPEC-INIT-HARNESS-PROMPT-001` | `@MX:SPEC` 추가 | 새 SPEC 을 함께 가리키고, 주석의 결정 B1 서술을 대체 사실로 고침 |
| `internal/cli/wizard/questions.go` `InitQuestions` | `@MX:NOTE` + `@MX:SPEC` | init 4문항은 ID 로 골라 조립한다. reconfigure 는 `DefaultQuestions` 를 그대로 쓴다(D1) |
| `internal/core/project/initializer.go:57-62` 워크트리 필드 주석 | `@MX:NOTE` | init 기록은 플래그 추적자만. 대화형 변경은 reconfigure·web |
| 홈 안전 헬퍼(테스트 파일) | `@MX:WARN` + `@MX:REASON` | 패키지 변수·환경을 바꾸므로 `t.Parallel` 금지. 스파이를 빠뜨리면 실제 셸 설정 파일에 쓴다 |

## §J 위험

| # | 위험 | 대응 |
|---|---|---|
| R1 | t586 이 같은 위저드 파일을 동시에 고쳐 병합 충돌 | §F.1 구간 공유, 렌더링 코드 미접촉, 통합 창 순서는 리드가 정함 |
| R2 | `saveBoolAnswer` 가 빈 함수가 되어 린터(`unparam` 등)가 경고 | M6 에서 린트 결과 확인. 경고가 나면 결과를 기록하고 리드에 보고(억제 주석을 임의로 달지 않음) |
| R3 | 새 프로젝트의 디스크 모양이 달라져 `moai update` 3-way 병합에서 기존 프로젝트와 다르게 다뤄짐 | spec.md §6 에 기록. 해석값 동일성은 AC-IQW-006, 모양 동일성은 REQ-IQW-015·AC-IQW-008 로 봉인. 병합 영향 측정은 미측정 목록 |
| R4 | 기존 테스트 헬퍼 `runInitForAutonomyAtHomeCapturingOut`(`init_autonomy_wiring_test.go:40`)가 `t.Setenv("HOME")` 을 써서 새 테스트가 무심코 재사용 | 새 테스트는 M1 헬퍼만 사용. AC-IQW-005 가 새 테스트 파일에서 `t.Setenv("HOME"` 부재를 검사. 필드 참조만 지우는 기존 테스트는 헬퍼를 유지하고(§B B7) 실행 슬롯 지문으로 덮음 |
| R5 | 스파이가 막지 못하는 다른 홈 쓰기 경로 | 코드 안 전후 대조(새 테스트)와 슬롯 지문(모든 실행)이 8항목의 변화를 잡음. 8항목 밖 경로는 §K 미측정 |
| R6 | 주입한 위저드 결과가 운영 시드(`RunWithDefaults` 46-52)와 달라 테스트가 운영 경로를 재지 못함 | 주입 결과에 시드 5개를 똑같이 싣고, 기존 `TestSeedMirrorsProductionLSPSeed`(`init_flag_precedence_test.go:312`)를 함께 돌림 |
| R7 | core/project 기록기 삭제가 목록에 없는 테스트를 깨뜨림 | §G 재확인 명령, M5 에서 `./internal/core/project/...` 전체 실행 |
| R8 | 완료된 SPEC 9개의 본문이 코드와 어긋난 채 남음 | spec.md §5 관계표, sync 단계에서 HISTORY 개정 |
| R9 | 슬롯 지문이 거짓 차이를 냄 — 앞뒤 명령이 다르거나(카드 t661 전례), 다른 세션이 같은 시각에 `~/.claude/settings.json` 을 고침 | 같은 명령 본문 강제, 표준 오류 분리, 차이가 나면 되돌리지 않고 멈춘 뒤 리드가 원인 판정 |
| R10 | 시접을 우회해 기록 함수를 직접 부르는 구현 회귀가 생기면, 주 관측 테스트는 호출 0회로 RED 가 되지만 그 실행 중에 실제 셸 설정 파일에 쓸 수 있다 | 그런 뮤턴트는 만들지 않는다(§D). 구현 회귀는 슬롯 지문과 REQ-IQW-016 멈춤 규칙이 받친다. 셸 설정 줄이 이미 있는 머신에서는 기록이 건너뛰어져 지문이 불변일 수 있음(`internal/shell/config.go:93`, `:167`) |
| R11 | 시접을 바꾸는 테스트와 같은 패키지의 병렬 테스트가 전역 변수를 동시에 읽어 데이터 경합이 난다. 또는 원복되지 않은 스파이가 뒤 테스트로 새어 관측을 오염시킨다 | 리드 조건(2026-09-11): 시접을 바꾸는 테스트·헬퍼는 `t.Parallel` 을 쓰지 않고 `t.Cleanup` 으로 원복한다. 검증 시점에 대입 파일을 패턴으로 다시 뽑고(빈 스윕은 실패), 병렬 금지는 텍스트(`t.Parallel` 부재)와 실행(`t.Setenv` 병렬 충돌 panic)으로, 원복은 텍스트(대입 줄 수·`t.Cleanup`)와 실행(`-count=2` 재진입)으로 판정하며 뮤턴트 C1·C2·D 를 기록한다(acceptance.md AC-IQW-016). CI 경합 검출은 보조 신호일 뿐이다 — 병렬 테스트가 하나뿐이면 `t.Parallel` 이 더해져도 조용하다 |

## §K 미측정 (Gaps)

- `--force` 재초기화, moai 항목이 없는 기존 `.mcp.json` 이 있는 디렉터리에서 대화형·비대화형 init 의 `.mcp.json` 결과.
- 바이너리 실행 경로(실제 TTY, huh 렌더)에서 4문항 화면. 테스트는 `runWizardFn` 주입으로 대체한다.
- `internal/cli/root.go:55` 루트 탐색 가드의 도달 여부 — 홈을 읽기만 하는지 확인하지 않았다.
- `deployTemplates` 가 `os.UserHomeDir()`(`initializer.go:394`)로 실제 홈 경로를 렌더 파일에 굽는 읽기 경로. 쓰기는 아니지만 테스트 산출물에 실제 홈 경로가 들어간다.
- 같은 값을 다시 쓰는 yamlpatch 줄 교체의 바이트 동일성, `audit:` 키만 있고 하위 키가 없는 파일에서의 로더 기본값 유지.
- 새 프로젝트 모양 변화가 `moai update` 3-way 병합에 주는 실제 영향.
- `LLMConfig.Profile` 의 컴파일 기본값 위치(조사에서 찾지 못함). AC 는 디스크 값으로 대조한다.
- **대조 8항목 밖의 홈 경로 (REQ-IQW-012 범위 밖).** `runInit` 이 닿는 다음 함수들이 실제 홈에 쓰는지는 추적하지 않았다: `startDeferredUpdateNotice`(`internal/cli/init_update_notice.go:77`), `installPrePushHookOptional`(`internal/cli/hook_install.go:247`), `ScaffoldEvolutionDir`(`internal/cli/update/deploy/deploy.go:572`), `NewEnvConfigurator`(`internal/shell/env.go:32`, 스파이가 대신하는 기본 함수 안에서만 호출). 이 경로들의 쓰기는 어떤 AC 로도 잡히지 않는다. 필요하면 대조 목록을 넓히는 별도 결정이 있어야 한다.
- **기존 HOME 헬퍼의 셸 설정 안전성은 코드 판독뿐이다.** `t.Setenv("HOME")` 을 쓰면 `internal/shell/detect.go:128` 의 `os.Getenv("HOME")` 도 임시 홈을 따라가므로 셸 설정 누출이 없다고 읽히지만, 실행으로 확인하지 않았다. 기존 헬퍼를 쓰는 테스트가 도는 첫 run 슬롯의 지문 대조(AC-IQW-015)가 첫 실행 확인이다.
- **`internal/core/project` 기존 테스트의 셸 설정 경로.** 이 패키지 테스트 중 `NewInitializer` 를 만드는 파일이 5개(`initializer_audit_wiring_test.go`, `initializer_mirror_notice_test.go`, `initializer_persist_test.go`, `initializer_test.go`, `phase_test.go`)이고, 이 패키지 어디에도 `t.Setenv("HOME")` 이 없으며, `SkipShellConfig` 는 `initializer.go:44`·`:333` 외에 나오지 않는다(2026-09-11 `git grep`). 이 테스트들이 `Init` 의 셸 설정 단계까지 도달하는지는 확인하지 않았다. 도달한다면 실제 홈 셸 설정에 쓸 수 있으므로 `./internal/core/project/...` 슬롯도 지문 절차 대상에 넣었다. 결함으로 단정하지 않는다. 새 시접은 이 테스트들에 스파이를 끼울 수단을 주지만, 끼우는 일은 이 SPEC 범위 밖이다.
- **셸 설정 줄이 이미 있는 머신에서의 지문.** `internal/shell/config.go:93`, `:167` 은 줄이 이미 있으면 기록을 건너뛴다. 그런 머신에서는 지문 불변이 시접의 효과를 증명하지 못한다. 시접의 효과는 스파이 호출 횟수(AC-IQW-004)로만 증명한다.

## §L 해소된 확인 항목

- **해소됨 — 리드 결정 (a), 2026-09-11.** 기존 테스트 소급 범위: 홈 안전 점검표는 이 SPEC 이 새로 쓰거나 본문을 다시 쓰는 init 실행 테스트의 코드에만 넣고, 필드 참조만 지우는 기존 테스트는 기존 HOME 헬퍼를 유지한다. 기존·신규 init 실행 테스트를 돌리는 모든 run 슬롯은 실제 홈 지문 절차를 따른다(spec.md §2.3·§4.2·§4.3, REQ-IQW-012·014·016, AC-IQW-005·015).
- **해소됨 — 리드 결정, 2026-09-11 (plan 감사 D1).** 셸 설정 단계 도달 증거: 소스 문자열 가드와 커밋 조상 판정 대신, `internal/core/project` 셸 설정 단계 시접에 스파이를 끼운 실제 실행 관측과 기록된 배선 제거 뮤턴트로 증명한다(REQ-IQW-011, AC-IQW-004, design.md §4). 남은 확인 항목 없음.

## §M 교차 참조

- 근거: `.moai/reports/t583/verdict.md`, `.moai/reports/t583/research-explore.md`, `.moai/reports/t583/audit-excerpt.md`, 재현 프로브 `.moai/reports/t583/repro/t583_repro_test.go`(홈 시접 패턴 원형)
- plan 감사 1회차: `.moai/reports/t583/plan-audit.md`(FAIL 0.79, D1~D8)
- 제거 선례: `internal/cli/wizard/question_removal_test.go`
- 스윕 확인 근거: `.claude/rules/moai/development/verification-completeness.md` §1.1
- 관련 카드: t584(자율 등급), t585(하네스 3-way), t586(위저드 렌더·i18n), t588(정합성, reconfigure 답 저장), t661(앞뒤 도구가 달라 거짓 차이를 낸 지문 전례)

## §N 범위 밖 후속 후보

- **기존 HOME 헬퍼의 시접 기반 이관.** `runInitForAutonomyAtHomeCapturingOut`(`internal/cli/init_autonomy_wiring_test.go:40`)와 그 호출 테스트를 `t.Setenv("HOME")` 에서 M1 의 시접 기반 홈 안전 헬퍼로 옮긴다. 이 SPEC 에서는 하지 않는다(spec.md §8). 카드는 만들지 않으며, 리드가 큐 복구 뒤 발행한다. 이관이 끝나면 슬롯 지문 절차의 대상이던 기존 테스트도 코드 점검표를 갖추게 된다.
