# t583 plan 조사 결과 — init 위저드 질문 표면 (Explore, 2026-09-10)

읽기 전용 코드 판독만 했다(실행·컴파일 없음). 경로는 워크트리 `.claude/worktrees/t583` 기준. "바이트 동일" 같은 주장은 측정이 아니라 판독이다. SPEC 작성자는 이 파일을 research.md 의 근거로 삼는다.

## 1. 현재 init 질문 집합 — 18개 (보고서의 16개가 아님)

- 위저드는 `!nonInteractive && isInteractiveStdin()` 일 때만 돈다(`internal/cli/init.go:694`, 호출 `:704`). `runWizardFn` → `wizard.RunWithDefaults`(`init_update_notice.go:69-71`) → `InitQuestions(projectRoot)`(`wizard.go:39`).
- `InitQuestions` = `DefaultQuestions`(5) + `Page3Questions`(13) (`questions.go:296-303`). init 질문에는 Condition 이 없다. Git 질문 7개는 reconfigure 전용(`questions.go:192-259,268-287`).
- 예전 질문 4개(lsp_enabled·enforce_quality·design_enabled·claude_design_enabled)는 묻지 않고 true 로 시딩(`wizard.go:46-52`).

| ID | 기본값 | WizardResult 필드 |
|---|---|---|
| conversation_language (유지) | `en`, 프로필 로케일로 미리 채움 (`questions.go:74`) | ConvLang |
| user_name (유지) | 빈 값, 프로필로 미리 채움 (`questions.go:85`) | UserName |
| project_name | 디렉터리 이름, `.`·`/` 는 `my-project` (`questions.go:50-53,95`) | ProjectName |
| model_policy | medium (`questions.go:122`) | ModelPolicy |
| report_format | html+md (`questions.go:139`) | ReportFormat |
| project_mode | personal (`questions.go:362`) | ProjectMode |
| worktree_auto_create | false (`questions.go:373`) | WorktreeAutoCreate bool |
| todo_enabled | true (`questions.go:385`) | TodoEnabled *bool |
| feedback_auto_submit | false (`questions.go:396`) | FeedbackAutoSubmit *bool |
| project_continuation | card (`questions.go:413`) | ProjectContinuation |
| audit_model | claude (`questions.go:430`) | AuditModel |
| audit_gate_claude / codex / glm | required / required / advisory (`questions.go:444,457,470`) | AuditGate* |
| codex_audit_enabled | false (`questions.go:478`) | CodexAuditEnabled |
| agent_wiring (유지) | claude (`questions.go:500`) | AgentWiring |
| mcp_provision | true (`questions.go:508`) | MCPProvision 평범한 bool |
| autonomy_tier (유지) | semi-auto, 그룹 "Autonomy" (`questions.go:523`) | AutonomyTier |

그룹은 같은 Group 라벨의 연속 질문으로 묶이고(`wizard.go:176-189`), 진행 표시 분모는 동적이다(`wizard.go:236-238`). 유지할 4개는 현재 세 라벨(Basic·Quality & Workflow·Autonomy)에 걸쳐 있다.

## 2. 제거 질문의 행선지와 "안 물었을 때"

"안 물음" = 현재 `--non-interactive` 경로. `wizardResult` 는 0값(`init.go:692`), `applyWizardPage3ToOpts` 는 실행되지 않는다(`init.go:754`).

| 질문 | 쓰기 경로 | 안 물으면 | 오늘 기본값을 받아들인 사용자와 디스크가 달라지나 |
|---|---|---|---|
| project_name | `applyDetectedDefaults` 가 `filepath.Base` 로 채움(`phase.go:196-199`) | 디렉터리 이름 | 아니오. 파일시스템 루트 `/` 에서만 `my-project` vs `/` |
| model_policy | `ApplyPerformanceTier` 는 값이 있을 때만(`init.go:925-933`), `ApplyProfile` 은 항상(빈 값 → medium, `init.go:940-953`) | `profile: medium`, `performance_tier` 는 템플릿 그대로 | 바이트만: 대화형은 `performance_tier: medium`, 비대화형은 `"medium"`(따옴표) |
| report_format | `writeReportConfig` 항상 실행, 빈 값 → html+md (`initializer.go:253,579-593`) | html+md | 아니오 |
| project_mode | `writeProjectModeYAML` 항상 실행, 빈 값 → personal (`initializer_expansion.go:34,205-232`) | personal | 아니오 |
| worktree_auto_create | 추적자 켜진 키만 씀(`initializer_workflow_toggles.go:38-54`), 추적자는 플래그만 켬 | no-op | 아니오(이미 F1 로 버려짐). `initializer.go:57-62` 주석은 "위저드 답은 참고용"이라 적어 `init.go:300-304`·`questions.go:372` 와 모순 |
| todo_enabled | nil 이면 안 씀(`initializer_expansion.go:175-201`) | 템플릿에 키 없음 | **바이트 달라짐**: 오늘 대화형은 `workflow.todo.enabled: true` 를 넣고(키가 없어 yamlpatch 재인코딩 경로), 제거 뒤엔 키 없음. 런타임 값은 둘 다 true(`todo_enabled.go:29-34`) |
| feedback_auto_submit | nil 이면 안 씀(`initializer_expansion.go:126-152`) | 템플릿 그대로 | 아니오(판독) |
| project_continuation | 빈 값이면 안 씀(`initializer_expansion.go:83-108`) | 템플릿 그대로 | 아니오(판독) |
| audit_model + 게이트 3 | 대화형은 `AuditConfigSet = true` 무조건(`init.go:338`) → `writeWorkflowAuditYAML` 이 model·gates 4키를 **삽입**(`initializer_audit.go:37-79,184-214`) | no-op | **예: 대화형만 audit 블록을 씀.** 해석 값은 컴파일 기본값과 같다(`defaults.go:966-973`, `loader.go:223-237`) |
| codex_audit_enabled | true 일 때만 씀(`initializer_audit.go:67-73`) | no-op | 아니오 |
| mcp_provision | `init.go:337` 이 유일한 쓰기, `mcpDeclined := !opts.MCPProvision`(`init.go:1004-1011`) | false → ensure-entry 호출 생략(agent_wiring=both 제외) | 신규 init: 디스크 차이 없음(템플릿 `.mcp.json` 에 moai 포함, ensure-entry 는 멱등 건너뜀 `mcp_server.go:864-904`). 안내 문구만 다름. **`--force` 재초기화나 moai 항목 없는 기존 `.mcp.json` 에서는 결과가 달라짐**(미측정) |

## 3. 컴파일 기본값 vs 템플릿 기본값

- 부분 재정의 계약: `Load` 는 `NewDefaultConfig()` 에서 출발(`loader.go:28-36`). llm·workflow·feedback 로더는 기본값 시딩(`loader.go:190-201,223-237,276-290`). `project.yaml`·`report.yaml` 은 로더 소비자가 없는 고아 섹션.
- workflow.todo.enabled: 템플릿에 키 없음, 런타임 true(의도).
- workflow.audit.model/gates.*: 템플릿에 키 없음, 컴파일 기본값이 공급.
- llm.performance_tier: 따옴표 차이만. `LLMConfig.Profile` 의 컴파일 기본값은 찾지 못함.
- 나머지(worktree.auto_create·feedback.auto_submit·project.continuation·codex.review_gate.enabled·project.mode·report.format)는 차이 없음.
- `.mcp.json`: SPEC-MCP-DEFAULT-ON-001 REQ-A-1 은 "항목 정확히 1개"라 적지만 템플릿은 moai·context7 2개.

## 4. 현재 질문 집합을 고정한 테스트

- 집합·순서·개수 고정: `questions_test.go:87-128,316-334`, `expansion_test.go:11-69`, `restructure_test.go:68-100`, `unified_form_test.go:87-131,242-262`, `wizard_test.go:624-672`(분모 19), `agent_wiring_question_test.go:65-93`(**mcp_provision 이 agent_wiring 바로 뒤여야 함 — 제거와 정면 충돌**).
- 질문별로 무의미해지는 테스트: `questions_test.go`(report_format·model_policy), `wizard_test.go` 여러 곳, `coverage_boost_test.go`(project_name 픽스처·ko 번역 의존), `model_policy_default_test.go`, `model_policy_matrix_agreement_test.go`, `translations_completeness_test.go:21-83`, `expansion_test.go:71-263`, `restructure_test.go:125`, `worktree_test.go`, `todo_enabled_test.go`, `feedback_auto_submit_test.go`, `project_continuation_test.go`, `mcp_audit_test.go`.
- 제거 선례 패턴: `question_removal_test.go:22-56` — `removedInM3` 에 ID 를 넣으면 ko/ja/zh 번역과 `saveAnswer`/`saveBoolAnswer` 분기 삭제를 강제(REQ-WIZ-014/017).
- CLI 쪽: `init_audit_test.go:20-66`(AuditConfigSet=true 단언), `init_audit_wiring_test.go`, `init_workflow_wiring_test.go:100-134`(F1 사각지대), `init_flag_precedence_test.go`, `init_agent_wizard_test.go:64-218`, 픽스처 `init_wizard_identity_test.go`·`init_gitdetect_test.go`·`init_update_notice_test.go`. 헬퍼 다수가 `t.Setenv("HOME")` 사용.
- 번역(ko/ja/zh만, 영어는 질문 원문): 제거 대상 14개 ID 의 항목이 `translations.go:31-536` 에 있다. project_name·report_format·model_policy 는 reconfigure 가 유지하면 살아 있다. `development_mode` 는 이미 고아.

## 5. runInit 이 닿는 홈 해석

| 해석 | 위치 | 읽기/쓰기 | 세 우회로 막히나 |
|---|---|---|---|
| `userHomeDirFn` → `paths.Home`(HOME 우선) | 자율 번들 `~/.claude/settings.json`(`init.go:889-900`), `ensureGlobalSettingsEnv`(`init.go:964`) | 쓰기 | 예 |
| `paths.MoaiHome`(MOAI_HOME) | `homestate.EnsureProjectLayout`(`init.go:877`) | 쓰기 | 예 |
| `profile.GetBaseDir` | 런치 원장·IsSetup·ReadPreferences(+ 이관 rename) | 읽기, rename 가능 | 예 |
| `os.UserHomeDir` | `deployTemplates` 가 홈 경로를 렌더 파일에 굽고 go bin 탐지(`initializer.go:394-411`) | 읽기(실제 홈 경로가 배포 파일에 들어감) | 아니오 |
| `os.Getenv("HOME")`(`shell/detect.go:128-140`), `os.UserHomeDir`(`shell/config.go:37-45,223`) | `configureShellEnv`(`initializer.go:333-347`) → `~/.zshenv` 등에 **O_APPEND** | **쓰기** | **아니오 — 누출.** `runInit` 은 `SkipShellConfig` 를 켜지 않음(`init.go:592-617`). 줄이 이미 있으면 건너뜀 |
| `paths.Home` | `root.go:55` 루트 탐색 가드 | 읽기 | 아니오(도달 여부 미추적) |

결론: 세 우회는 셸 rc 추가 쓰기를 제외한 모든 홈 쓰기를 막는다. "미설정=기본값" 실행 테스트에는 셸 설정 쓰기를 막는 새 시접이 필요하다.

## 6. moai web 에서 편집 가능한가

- 표면 없음: project.name·project.mode(`project` 섹션 RouteExcluded, `sectionroute.go:21-22`), `.mcp.json` moai 항목(web MCP 는 `mcp.yaml` 도구별 토글만).
- 표면 있음: performance_tier(`settings_shell.go:145-146` 등), report.format(`schema_sections.go:543-560`), worktree.auto_create(`:351`), todo.enabled(`:367`), feedback.auto_submit(`:473`), project.continuation(`:376-378`), audit.model/gates(`:393-404`), codex.review_gate.enabled(`:429`).

## 7. `moai update -c` (reconfigure)

- `-c` → `runInitWizard(cmd, true)` → `ReconfigureQuestions` = `DefaultQuestions` + `GitQuestions`, 3페이지 없음(`update_wizard.go:64`, `questions.go:268-295`).
- **`DefaultQuestions` 에서 project_name·model_policy·report_format 을 빼면 reconfigure 에서도 빠진다** — 생성자를 나누지 않는 한.
- reconfigure 뒤 `runWorkflowConfigStep` 이 워크트리 자동 생성 포함 y/n 4개를 따로 묻고 저장한다(`init_workflow_flags.go:66-101`).
- 기존 결함: `applyWizardConfig` 가 `ReportFormat`·`ProjectName` 을 저장하지 않는다(`update_wizard.go:133-373`) — reconfigure 의 report_format 답이 버려진다(F1 과 같은 모양). model_policy 는 `llm.profile`·`system.yaml` 에만 쓰고 `performance_tier` 는 안 씀(`:307-332`).

## 8. 이 질문들을 가진 기존 SPEC

| SPEC (상태) | 16→4 축소가 대체·충돌하는 요구 |
|---|---|
| SPEC-CLI-WIZARD-RESTRUCTURE-001 (completed) | REQ-WIZ-001 3페이지, -003 page 1 프로젝트 이름, -004 모델·리포트 페이지, -015 page 3 저장, -016 reconfigure 순서. -014/-017 은 번역·저장 분기 삭제 의무가 됨 |
| SPEC-V3R5-INIT-WIZARD-EXPANSION-001 (implemented) | REQ-IWE-001 project.mode 질문 대체. -008 플래그는 유지 |
| SPEC-INIT-WIZARD-REPAIR-001 (completed) | REQ-005 무의미화. -006 추적자·-007 update 단계 유지. -008 audit 쓰기는 init 에서 도달 불가 |
| SPEC-INIT-HARNESS-PROMPT-001 (completed) | REQ-IHP-001 유지. REQ-IHP-009 "claude 는 mcp_provision 답을 유지" 대상 소멸, 결정 B1 "mcp_provision 무조건 질문"과 충돌 |
| SPEC-MCP-DEFAULT-ON-001 (completed) | REQ-A-3 기본 true 확인 질문 + 거절 경로, REQ-A-5 로케일 문자열 |
| SPEC-MOAI-MCP-SERVER-001 (completed) | REQ-MCP-015 "init 위저드가 audit·MCP 프로비저닝을 노출" — init 쪽과 정면 충돌(web 쪽은 충족) |
| SPEC-TODO-ENABLE-FLAG-001 / SPEC-FEEDBACK-AUTO-SUBMIT-001 / SPEC-PROJECT-CONTINUATION-KEY-001 (completed) | 각 위저드 질문 요구 |
| SPEC-AUTONOMY-TIERS-001 (completed) | REQ-001 유지 |
| SPEC-WT-DOC-001 (archived) | 위저드 요구 없음. 코드 주석이 여전히 인용(출처 드리프트) |
| SPEC-CLI-TUX-INIT-UPDATE-001 (completed) | 표시 전용, 충돌 없음 |

## 위험과 설계 결정 거리

1. 개수: 18→4 (14개 제거).
2. 공유 생성자: `DefaultQuestions` 는 init·reconfigure 공용 — reconfigure 에서 project_name·model_policy·report_format 을 유지할지 결정 필요.
3. mcp_provision: 평범한 bool 이라 질문을 없애면 모든 실행에서 "거절"로 읽힌다. 코드에서 기본 true 로 둘지, ensure-entry 호출을 없앨지 결정 필요. `--force`·기존 `.mcp.json` 에서 동작이 달라질 수 있다(미측정).
4. 기존 프로젝트는 위저드가 쓴 모양(audit 블록·todo.enabled·따옴표 없는 performance_tier)을 유지한다. 새 프로젝트는 컴파일 기본값에 기대므로 해석 값은 같지만 디스크 모양이 다르다 — update 3-way 병합 스냅숏에 영향.
5. init 에서 도달 불가해지는 코드: `applyWizardPage3ToOpts`, 워크플로 토글·todo·feedback·continuation·audit 쓰기의 위저드 분기, `saveBoolAnswer`, `buildConfirmField` — 플래그용으로 남길지 지울지 항목별 결정.
6. F1: 질문을 없애면 init 증상은 사라진다. 모순된 문서 3곳은 정리 필요.
7. "미설정=기본값" 테스트: 셸 rc 시접 필요, 환경 교체 때문에 `t.Parallel` 불가.
8. web 표면 공백: project.name·project.mode·`.mcp.json` 프로비저닝은 유일한 대화형 표면을 잃는다(CLI 플래그는 남음: `--name`, `--project-mode`, `--llm`).
9. 완료된 SPEC 최소 8개에 HISTORY 개정 필요.

## 확신도와 공백

- 높음: 질문 구성, opts·쓰기 매핑, 추적자·포인터 의미, web 라우팅, reconfigure 배선.
- 미측정: 같은 값 yamlpatch 줄 교체의 바이트 동일성, audit 블록 없이 `audit:` 만 있을 때 기본값 유지, 멱등 건너뜀 경로의 `.mcp.json.lock` 생성, `--force` 재초기화의 `.mcp.json` 동작.
- 미발견: `LLMConfig.Profile` 컴파일 기본값.
- 미추적: `defaultDeferredUpdateCheck` 내부, `NewEnvConfigurator` 본문, `root.go:55` 도달 여부.
- 미확인: `autonomy_test.go`, `pat_mask_test.go`, web 의 autonomy_tier 표면.
