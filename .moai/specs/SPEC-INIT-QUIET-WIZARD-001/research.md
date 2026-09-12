# SPEC-INIT-QUIET-WIZARD-001 — 조사

## §0 출처와 재확인 범위

- 원 조사: `.moai/reports/t583/research-explore.md` (읽기 전용 판독, 2026-09-10). 재현 판정: `.moai/reports/t583/verdict.md`.
- 이 문서의 file:line 은 2026-09-11 에 워크트리 HEAD `120436f58` 에서 `sed`/`grep` 으로 다시 읽어 확인한 것이다. 확인하지 못한 인용은 각 표에 "원 조사 인용(미재확인)" 으로 표시한다.
- 읽기 전용 확인만 했다. 컴파일·테스트·바이너리 실행은 하지 않았다. 여기 적은 동작 주장은 F1·F4 재현(verdict.md)을 빼면 모두 코드 판독이며 측정이 아니다.

## §1 init 질문 집합 — 18문항

| 사실 | 근거(재확인) |
|---|---|
| 위저드는 비대화형이 아니고 stdin 이 TTY 일 때만 돈다 | `internal/cli/init.go:694`, 호출 `:704` |
| `runWizardFn` → `wizard.RunWithDefaults` | `internal/cli/init_update_notice.go:69-71` |
| `RunWithDefaults` 가 `InitQuestions` 를 쓰고, 묻지 않는 4개 설정을 true 로 시드 | `internal/cli/wizard/wizard.go:35-59`(시드 46-52) |
| `InitQuestions` = `DefaultQuestions` + `Page3Questions` | `internal/cli/wizard/questions.go:296-303` |
| `DefaultQuestions` 5문항 | `questions.go:48`, ID 63·80·90·100·130 |
| `GitQuestions` 7문항(reconfigure 전용) | `questions.go:162`, ID 166~250 |
| `ReconfigureQuestions` = `DefaultQuestions` + Git(report_format 뒤) | `questions.go:268-289` |
| `Page3Questions` 13문항 | `questions.go:349`, ID 353·368·380·391·403·419·434·447·460·473·490·503·513 |
| init 질문에는 `Condition` 없음 | `expansion_test.go:11-69` 의 `gated: false` 표 |

질문별 기본값과 결과 필드:

| ID | 기본값 | `WizardResult` 필드(`types.go`) |
|---|---|---|
| conversation_language (유지) | `en`, 프로필 로케일로 미리 채움 | `ConversationLang` (15) |
| user_name (유지) | 빈 값, 프로필로 미리 채움 | `UserName` (16) |
| project_name | 디렉터리 이름 | `ProjectName` |
| model_policy | medium | `ModelPolicy` |
| report_format | html+md | `ReportFormat` |
| project_mode | personal (`questions.go:353` 문항) | `ProjectMode` (46) |
| worktree_auto_create | `"false"` (`questions.go:373`) | `WorktreeAutoCreate bool` (55) |
| todo_enabled | true | `TodoEnabled *bool` (62) |
| feedback_auto_submit | false | `FeedbackAutoSubmit *bool` (68) |
| project_continuation | card | `ProjectContinuation` (77) |
| audit_model | claude | `AuditModel` (88) |
| audit_gate_claude / codex / glm | required / required / advisory | `AuditGateClaude`·`AuditGateCodex`·`AuditGateGLM` (89-91) |
| codex_audit_enabled | false | `CodexAuditEnabled` (92) |
| agent_wiring (유지) | claude (`questions.go:490` 문항) | `AgentWiring` (100) |
| mcp_provision | `"true"` (`questions.go:503` 문항) | `MCPProvision bool` (93) |
| autonomy_tier (유지) | semi-auto, 그룹 Autonomy (`questions.go:513` 문항) | `AutonomyTier` (82) |

그룹 묶기: 같은 `Group` 라벨이 연속한 무조건 질문끼리 한 huh 그룹이 된다(`wizard.go:158` `buildFormGroups` 본문). 스테퍼 분모는 `stepperDenominator`(`wizard.go:236`)가 보이는 질문 수로 동적으로 계산한다.

## §2 답이 디스크로 가는 길

| 질문 | 매핑(`init.go`) | 기록기 | 묻지 않을 때 |
|---|---|---|---|
| project_name | `:729-731` 빈 opts 일 때만 | `applyDetectedDefaults`(`internal/core/project/phase.go:195-199`) | 디렉터리 이름 |
| model_policy | `:739-741` | 성능 등급은 값이 있을 때만(`init.go:925-933`), 프로필은 항상 기록하고 빈 값은 medium(`init.go:940-953`) | `llm.profile: medium`, `performance_tier` 는 템플릿 `"medium"`(`llm.yaml:48`) |
| report_format | `:748-750` | `writeReportConfig`, 빈 값 → html+md(`initializer.go:253` 호출, `:579-593`) | html+md |
| project_mode | `:278-280` (플래그 미지정일 때) | `writeProjectModeYAML`(`initializer_expansion.go:34` 호출, `:205`) | 템플릿 `mode: personal`(`project.yaml.tmpl:14`) |
| worktree_auto_create | `:305-307` — 값만 넣고 추적자는 켜지 않음 | `WriteWorkflowTogglesYAML` 는 추적자가 켜진 키만(`initializer_workflow_toggles.go:38`, 호출 `initializer.go:289`) | 템플릿 `auto_create: false`(`workflow.yaml:57`) |
| todo_enabled | `:313` | `writeWorkflowTodoYAML`, nil 이면 no-op(`initializer_expansion.go:175-178`, 호출 `:50`) | 키 없음, 로더 해석 true(`internal/config/todo_enabled.go:29`) |
| feedback_auto_submit | `:319` | `writeFeedbackAutoSubmitYAML`, nil 이면 no-op(`initializer_expansion.go:126-127`, 호출 `:53`) | 템플릿 `auto_submit: false`(`feedback.yaml:13`) |
| project_continuation | `:325` | `writeWorkflowProjectContinuationYAML`, 빈 값이면 no-op(`initializer_expansion.go:83-84`, 호출 `:56`) | 템플릿 `continuation: card`(`workflow.yaml:47`) |
| audit_model + 게이트 3 | `:332-335`, 그리고 `:338` `AuditConfigSet = true` 무조건 | `writeWorkflowAuditYAML`, `AuditConfigSet` false 면 no-op(`initializer_audit.go:37-38`, 호출 `initializer.go:306`) | 블록 없음, 컴파일 기본값 claude / required / required / advisory(`internal/config/defaults.go:966-973`) |
| codex_audit_enabled | `:336` | 같은 기록기, true 일 때만(`initializer_audit.go:67`) | 템플릿 `review_gate`(`workflow.yaml:98`) |
| mcp_provision | `:337` (유일한 쓰기) | `mcpDeclined := !opts.MCPProvision`, codex → true, both → false(`init.go:1004-1011`), `provisionMCPEntryUnlessDeclined`(`init.go:252`) | 0값 false → 보장 호출 생략 |

`opts` 구성은 `init.go:592-617` 한 곳이며, 비테스트 코드에서 `project.InitOptions{` 를 만드는 곳은 여기뿐이다(`git grep` 결과 1건). 따라서 `InitOptions` 의 todo·feedback·continuation·audit 필드를 채우는 곳은 `applyWizardPage3ToOpts`(`init.go:277-339`) 뿐이다.

비테스트 코드에서 제거 대상 `WizardResult` 필드를 읽는 곳은 `init.go:278-337` 매핑과 `wizard.go` 의 답 저장 분기뿐이다. `update_wizard.go` 는 `ModelPolicy` 만 읽는다(`:307-310`).

## §3 F1 원인과 모순된 서술

| 위치 | 서술 | 재확인 |
|---|---|---|
| `internal/core/project/initializer.go:54-62` | 워크트리 필드는 추적자가 켜졌을 때만 기록되고, "the wizard advisory alone is informational and leaves the deployed template default untouched" | 확인 |
| `internal/cli/init.go:300-304` | "wizard-only confirm, applies when the wizard ran AND --worktree-auto-create was not explicitly supplied" | 확인 |
| `internal/cli/wizard/questions.go:371-372` | 질문 설명 "When enabled, moai init / moai profile / moai web automatically enter a worktree" | 확인 |
| 추적자를 켜는 유일한 곳 | `internal/cli/init_workflow_flags.go:40-44` | 확인 |
| 사각지대 테스트 | `internal/cli/init_workflow_wiring_test.go:100` — 위저드 답을 `false` 로만 넣음 | 확인 |

## §4 F4 와 MCP 기본값

- 비대화형 신규 init 이 배포한 `.mcp.json` 의 `mcpServers` 키: `[moai context7]` (verdict.md §2 F4, 측정).
- 비대화형 보장 호출 생략은 `init_agent_wizard_test.go:172-205`(`TestRunInit_FlagAbsentNonInteractivePreservesCodexAbsence`, 함수 186) 가 의도로 고정. 안내 문구 상수는 `init_agent_wizard_test.go:31` `"Provisioned the moai MCP server entry in .mcp.json (default-on)."`.
- 소스 도달성 가드: `init_mcp_provision_test.go:105` `TestRunInit_CallsMCPProvisioning` 은 함수 본문에 `opts.MCPProvision` 문자열이 있어야 통과.
- 기존 `@MX:SPEC: SPEC-INIT-HARNESS-PROMPT-001` 태그: `init.go:1003`. 그 위 주석(`:999` "Accepted cost (plan.md §B Decision B1)")은 codex 선택자에게도 `mcp_provision` 을 묻는 결정을 설명한다.
- 보장 호출의 멱등 건너뜀(`mcp_server.go:864-904`): 원 조사 인용(미재확인).

## §5 홈 해석과 셸 설정 누출

| 해석 | 위치 | 쓰기 여부 | 세 우회로 막히나 |
|---|---|---|---|
| `userHomeDirFn` → `paths.Home`(HOME 우선) | 선언 `glm_tools.go:123-124`, 위임 `homedir.go`, 사용 `init.go:889` | 쓰기(전역 settings) | 예 |
| `ensureGlobalSettingsEnv` | `init.go:964` | 쓰기 | 예(원 조사 인용: `userHomeDirFn` 경유) |
| `homestate.EnsureProjectLayout`(MOAI_HOME) | `init.go:877` | 쓰기 | 예 |
| `os.UserHomeDir` in `deployTemplates` | `initializer.go:394` | 읽기(렌더 파일에 경로가 들어감) | 아니오 |
| 셸 설정: `SkipShellConfig` false 이면 `configureShellEnv` | 필드 `initializer.go:44`, 분기 `:333-347` | **쓰기** | **아니오** |
| 셸 파일 선택: `os.Getenv("HOME")` | `internal/shell/detect.go:128-145` | — | 아니오 |
| 셸 파일 추가: `O_APPEND|O_CREATE|O_WRONLY` | `internal/shell/config.go:122`, `:199`; `os.UserHomeDir` `:38`, `:223` | 쓰기 | 아니오 |
| `runInit` 이 `SkipShellConfig` 를 설정하는가 | `init.go:592-617` 구성에 없음, `internal` 전체에서 `SkipShellConfig` 는 `initializer.go:44`·`:333` 두 곳뿐 | — | — |
| `root.go:55` 루트 탐색 가드 | 원 조사 인용(미재확인), 도달 여부 미추적 | 읽기 | — |

재현 프로브(`.moai/reports/t583/repro/t583_repro_test.go:24-121`)는 세 우회와 실행 전 자기 가드, 전후 sha256 대조를 이미 갖췄다. 셸 설정 파일은 대조 목록에 없었고, 사후 mtime 비교로만 이번 실행이 쓰지 않았음을 확인했다(verdict.md §4).

기존 테스트 헬퍼 `runInitForAutonomyAtHomeCapturingOut`(`init_autonomy_wiring_test.go:40`)는 `t.Setenv("HOME")` 을 쓴다. 이 방식에서는 `internal/shell/detect.go:128` 의 `os.Getenv("HOME")` 도 임시 홈을 따라가므로 셸 설정 누출이 없다고 판독된다. 이것은 코드 판독일 뿐 실행 확인이 아니다. 리드 결정 (a)(2026-09-11)에 따라 이 헬퍼를 쓰는 기존 테스트는 그대로 두므로, run 단계에서 이 헬퍼를 쓰는 테스트를 처음 돌리는 슬롯의 실제 홈 지문 대조(spec.md §4.3, acceptance.md AC-IQW-015)가 이 판독의 첫 실행 확인이 된다.

### §5.1 셸 설정 단계 시접 설계의 근거 (plan 감사 D1, 2026-09-11 재확인)

| 사실 | 근거(재확인) | 설계에 주는 의미 |
|---|---|---|
| `runInit` 은 initializer·executor 를 인라인으로 만들고 `opts` 를 곧바로 넘긴다 | `internal/cli/init.go:832` `project.NewInitializer`, `:833` `project.NewPhaseExecutor`, `:867` `executor.Execute(ctx, opts)` | `internal/cli` 테스트는 `InitOptions` 를 중간에서 볼 수 없다 |
| Step 6 은 `configureShellEnv` 를 부르고, 이 함수는 `shell.NewEnvConfigurator(i.logger).Configure(...)` 한 호출뿐이다 | `internal/core/project/initializer.go:333-347`, `:677-685` | 이 한 호출을 함수 변수로 바꾸면 운영 동작 그대로 스파이를 끼울 수 있다 |
| 설정 줄이 이미 있으면 기록을 건너뛴다 | `internal/shell/config.go:93`(`already configured`), `:167`(`PATH ... already configured`) | 이미 설정된 머신에서 "셸 설정 파일 불변" 은 시접과 무관하게 나오므로 증거가 될 수 없다 |
| 운영 코드에서 셸 설정 기록을 부르는 곳은 두 곳 | `internal/core/project/initializer.go:678`, `internal/cli/update.go:824` (`git grep 'NewEnvConfigurator('`, 정의 `internal/shell/env.go:32` 와 내부 `:197` 제외) | init 경로만 시접을 거치게 하면 되고, update 경로는 범위 밖 |
| `internal/core/project` 에는 아직 함수 변수 시접이 없다 | `git grep -nE '^var [A-Za-z]+Fn = ' -- internal/core/project ':!*_test.go'` 출력 없음 | 새 시접이 이 패키지의 첫 사례다 |
| 기존 소스 도달성 가드 선례 | `internal/cli/init_mcp_provision_test.go:105-117` `TestRunInit_CallsMCPProvisioning` — `init.go` 를 읽어 문자열 포함만 검사 | 리드 결정으로 이 형태는 새 도달 증거로 쓰지 않는다(텍스트 추론) |
| 템플릿 `workflow.yaml` 이 이미 `audit:` 키를 싣는다 | `internal/template/templates/.moai/config/sections/workflow.yaml:85-91` (`audit.codex`·`audit.glm` 의 model·effort, 값 빈 문자열) | 새 프로젝트 모양 비교는 `audit:` 키가 아니라 `model`·`gates` 하위 키 기준이어야 한다(plan 감사 D5) |
| 파일시스템 루트의 `project_name` 기본값 | `internal/cli/wizard/questions.go:50-53` — 이름이 `.`·`/`·`\` 면 `my-project`; `internal/core/project/phase.go:195-199` — `filepath.Base` 그대로 | REQ-IQW-003 예외절의 근거(plan 감사 D6) |
| 대조 8항목 밖의 홈 쓰기 후보 | `internal/cli/init_update_notice.go:77` `startDeferredUpdateNotice`, `internal/cli/hook_install.go:247` `installPrePushHookOptional`, `internal/cli/update/deploy/deploy.go:572` `ScaffoldEvolutionDir` — 함수 존재만 확인, 홈 쓰기 여부는 추적하지 않음 | REQ-IQW-012 를 8항목으로 좁히고 나머지는 plan.md §K 미측정으로 둔 근거(plan 감사 D2) |

## §6 로더와 템플릿 기본값

- `Loader.Load(configDir)` 는 `.moai` 디렉터리를 받아 `config/sections` 를 읽고, `NewDefaultConfig()` 에서 출발한다(`internal/config/loader.go:31-38`).
- `Config.TodoEnabled()` 는 키가 nil 이면 true(`internal/config/todo_enabled.go:29`).
- audit 컴파일 기본값: `defaults.go:966-973`(모델 claude, 게이트 required/required/advisory — 967-971 확인).
- 템플릿 값: `project.yaml.tmpl:6` `name: "{{.ProjectName}}"`, `:14` `mode: personal`; `llm.yaml:41` `profile: "medium"`, `:48` `performance_tier: "medium"`; `feedback.yaml:13` `auto_submit: false`; `workflow.yaml:47` `continuation: card`, `:57` `auto_create: false`, `:98` `review_gate:`; `report.yaml` 26바이트.
- workflow·feedback·llm 로더의 기본값 시딩(`loader.go:190-201,223-237,276-290`): 원 조사 인용(미재확인).
- `project.yaml`·`report.yaml` 은 로더 소비자가 없는 섹션: 원 조사 인용(미재확인). 실행 테스트 관측기는 두 파일을 직접 읽으므로 이 주장에 의존하지 않는다.
- `LLMConfig.Profile` 의 컴파일 기본값: 원 조사에서 찾지 못함.

## §7 reconfigure (`moai update -c`)

- 진입: `update_wizard.go:64` `wizard.ReconfigureQuestions(cwd)` — 비테스트 코드에서 이 생성자를 부르는 유일한 곳.
- 적용: `applyWizardReconfigureSteps` → `applyWizardConfig`(`update_wizard.go:133`) → `runWorkflowConfigStep`(`init_workflow_flags.go`, 워크트리 포함 y/n 4개를 따로 묻고 변경분만 기록).
- `applyWizardConfig` 는 `ModelPolicy` 만 읽고(`:307-310`) `ProjectName`·`ReportFormat` 은 읽지 않는다(`git grep` 0건, verdict.md §7.3) — t588 소관.
- reconfigure 순서 고정 테스트: `questions_test.go:184` `TestReconfigureQuestionsOrder` (12문항 목록 + Git 연속 블록 단언). `DefaultQuestions` 5문항 고정: `questions_test.go:87` `TestQuestionOrder`, `:316` `TestQuestionsAllPresent`.

## §8 현재 집합을 고정한 테스트 (재확인한 것)

| 테스트 | 위치 | 고정 내용 |
|---|---|---|
| `TestPage3QuestionsStructure` | `expansion_test.go:11` | 페이지 3 13문항 순서·타입·조건 |
| `TestInitPages_Membership` | `restructure_test.go:68` | 페이지 구성원(Basic 3, Model & Report 2, Q&W 12, Autonomy 1) |
| `TestInitPages_MergeIntoOneGroupPerPage` | `restructure_test.go:92` | Model & Report 포함 3페이지 각각 연속 1회 |
| `TestPage3_NoModeGate` | `restructure_test.go:125` | `project_mode` 무조건 노출 |
| `TestStepperTotal_DynamicDenominator` | `wizard_test.go:624-672` | reconfigure 분모 6·9·10, 페이지 3 합산 19 |
| `TestAgentWiringQuestion_PrecedesMCPProvision` | `agent_wiring_question_test.go:65` | `agent_wiring` 바로 뒤 `mcp_provision` |
| `TestWizardQuestionTranslationCompleteness` | `translations_completeness_test.go:95` | `InitQuestions` 전 문항 ko/ja/zh 번역 (면제 목록 `:13`) |
| `removedInM3` 제거 선례 | `question_removal_test.go:15-18`, 테스트 22~ | 부재·번역 고아·답 저장 분기 |
| `TestApplyWizardPage3ToOpts_AuditSelection` | `init_audit_test.go:20` | audit 매핑과 `AuditConfigSet=true` |
| `TestRunInit_WizardAuditSelectionPersists` | `init_audit_wiring_test.go:29` | 위저드 audit 블록 기록 |
| `TestRunInit_WorkflowToggleFlagsAbsentByteIdentical` | `init_workflow_wiring_test.go:83` | 무플래그 비대화형 workflow.yaml 바이트 동일 |
| `TestRunInit_WorktreeAutoCreateFlagBeatsWizard` | `init_workflow_wiring_test.go:100` | 플래그 vs 위저드 답(false) |
| `TestFlagBeatsWizard_Page3Settings` | `init_flag_precedence_test.go:60` | 페이지 3 플래그 우선순위 |
| `TestSeedMirrorsProductionLSPSeed` | `init_flag_precedence_test.go:312` | 시드와 운영 값 일치 |
| `WizardResult.MCPProvision` 픽스처 | `doctor_codex_e2e_test.go:54`, `:86` | codex·claude 경로 |

제거 필드를 참조하는 테스트 파일(`git grep -lw`):

- `ProjectMode`: `init_audit_test.go`, `init_flag_precedence_test.go`, `wizard/question_removal_test.go`, `wizard/expansion_test.go`, `core/project/initializer_{todo,expansion,persist,feedback}_test.go`
- `TodoEnabled`: `wizard/todo_enabled_test.go`, `core/project/initializer_todo_test.go`
- `FeedbackAutoSubmit`: `wizard/feedback_auto_submit_test.go`, `core/project/initializer_feedback_test.go`
- `ProjectContinuation`: `core/project/project_continuation_write_test.go`
- `AuditModel`: `init_audit_wiring_test.go`, `init_audit_test.go`, `mcp_convergence_test.go`(설정 열거형 검사로 이 필드와 무관), `wizard/mcp_audit_test.go`, `core/project/initializer_audit{,_wiring}_test.go`
- `MCPProvision`: `doctor_codex_e2e_test.go`, `init_audit_test.go`, `init_agent_wizard_test.go`, `init_mcp_provision_test.go`, `core/project/initializer_audit_test.go`
- `WorktreeAutoCreate`: `init_audit_test.go`, `init_workflow_flags_test.go`, `init_workflow_wiring_test.go`, `wizard/worktree_test.go`, `core/project/initializer_workflow_toggles_test.go`
- `AuditConfigSet`: `init_audit_wiring_test.go`, `init_audit_test.go`, `init_workflow_wiring_test.go`, `core/project/initializer_audit{,_wiring}_test.go`

`core/project` 쪽 기록기 단위 테스트: `initializer_todo_test.go:38·85·111`, `initializer_feedback_test.go:30·75·101`, `initializer_audit_test.go:24·80·114·143·197·233`, `initializer_audit_wiring_test.go:23·58`, `initializer_workflow_toggles_test.go:51~173`(유지 기록기).

## §9 번역 표

`internal/cli/wizard/translations.go` 의 로케일 표 시작: ko `:32`, ja `:201`, zh `:369`. 제거 대상 11개 ID 는 세 표 모두에 있다(ko 108·116·120·124·137·146·156·165·174·183·196, ja 276·284·288·292·305·314·324·333·342·351·364, zh 444~). `project_name`·`model_policy`·`report_format` 은 reconfigure 가 계속 렌더링하므로 남는다. `development_mode` 항목(ko 46 등)은 이미 대응 질문이 없는 고아이며 이 SPEC 범위 밖이다.

## §10 CLI 플래그 표면

| 플래그 | 선언 | 이 SPEC 뒤 |
|---|---|---|
| `--project-mode` | `init.go:92` | 유지, `init.go:609` 에서 opts 로 |
| `--worktree-auto-create` | `init.go:119` | 유지, 추적자 경로 |
| `--autonomy-tier` | `init.go:127` | 유지 |
| `--llm` | `init.go:133` | 유지 |

todo·feedback·continuation·audit·MCP 프로비저닝에는 init 플래그가 없다(`initCmd.Flags()` 선언 검색 결과).

## §11 web 콘솔 표면 (원 조사 인용, 미재확인)

- 표면 없음: `project.name`·`project.mode`(`sectionroute.go:21-22` 에서 project 섹션 제외), `.mcp.json` moai 항목.
- 표면 있음: performance_tier, report.format, worktree.auto_create, todo.enabled, feedback.auto_submit, project.continuation, audit.model/gates, codex.review_gate.enabled(`internal/settings/schema_sections.go` 여러 행).
- web 쪽은 `internal/core/project` 기록기를 부르지 않는다: `writeWorkflowAuditYAML` 등 기록기의 비테스트 호출자는 `initializer.go`·`initializer_expansion.go` 뿐(재확인).

## §12 기존 SPEC 요구 (재확인한 원문 위치)

| SPEC | 요구 위치 |
|---|---|
| SPEC-CLI-WIZARD-RESTRUCTURE-001 | `spec.md:42`(REQ-WIZ-001), `:44`(003), `:45`(004), `:61`(014), `:65`(015), `:66`(016), `:67`(017) |
| SPEC-V3R5-INIT-WIZARD-EXPANSION-001 | `spec.md:58`(REQ-IWE-001), `:79`(008) |
| SPEC-INIT-WIZARD-REPAIR-001 | `spec.md:84`(REQ-005), `:85`(006), `:86`(007), `:87`(008) |
| SPEC-INIT-HARNESS-PROMPT-001 | `spec.md:94`(REQ-IHP-001), `:116`(009); `plan.md:25` 결정 B1 |
| SPEC-MCP-DEFAULT-ON-001 | `spec.md:65`(REQ-A-3), `:69`(REQ-A-5) |
| SPEC-MOAI-MCP-SERVER-001 | `spec.md:106`(REQ-MCP-015) |
| SPEC-TODO-ENABLE-FLAG-001 | `spec.md:122`(REQ-4) |
| SPEC-FEEDBACK-AUTO-SUBMIT-001 | `spec.md:193`(REQ-11) |
| SPEC-PROJECT-CONTINUATION-KEY-001 | `spec.md:89`(REQ-PCK-010) — 원문은 "reconfigure 위저드" 라고 적지만, 해당 질문은 `Page3Questions` 에만 있어 실제로는 init 전용이다 |

## §13 확신도와 공백

- 높음(재확인): 질문 구성과 순서, opts 매핑과 기록기 게이트, 셸 설정 누출 경로, reconfigure 진입과 고정 테스트, 제거 필드 참조 테스트 목록.
- 측정: F1 재현, F4 반증(verdict.md).
- 미재확인: web 라우팅 행번호, 로더 시딩 행번호, 보장 호출 멱등 경로, `root.go:55`.
- 미추적: `NewEnvConfigurator` 본문, `defaultDeferredUpdateCheck` 내부.
- 미측정: `--force` 재초기화와 기존 `.mcp.json`, 같은 값 yamlpatch 교체의 바이트 동일성, 모양이 달라진 새 프로젝트의 3-way 병합 영향.
