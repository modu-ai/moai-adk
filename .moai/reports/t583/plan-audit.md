# SPEC Review Report: SPEC-INIT-QUIET-WIZARD-001
Iteration: 1/3
Verdict: FAIL
Overall Score: 0.79

작성자 추론 맥락은 M1 Context Isolation 에 따라 무시했다. 판정은 SPEC 산출물 5종(spec.md, plan.md, acceptance.md, design.md, research.md)과 워크트리 트리(HEAD `120436f58`, `internal/` 미수정 — `git status --short -- internal` 출력 없음)만 근거로 한다.

감사 경로: Claude 단독. `.moai/config/sections/` 어디에도 `audit_model` 키가 없고 컴파일 기본값이 `AuditModelClaude`(`internal/config/defaults.go` `Audit: AuditConfig{Model: AuditModelClaude, ...}`)이므로 MCP 교차 백엔드는 호출하지 않았다.

## Must-Pass Results

- [PASS] MP-1 REQ 번호 일관성: 요구사항 계층은 `spec.md:L75-88` 의 `REQ-IQW-001` ~ `REQ-IQW-014` 14개다. 3자리 채움이 일정하고 빈 번호·중복이 없다.
- [PASS] MP-2 EARS/GEARS 형식 (요구사항 계층 기준으로 판정): 14개 모두 다섯 GEARS 패턴 중 하나로 구조가 맞는다.
  - Ubiquitous: 001, 004, 008, 009
  - Unwanted `shall not`: 002, 012, 013
  - Event-driven `When`: 003, 010, 014
  - State-driven `While`: 005, 006
  - Where: 007, 011
  - Given-When-Then 은 `acceptance.md` 의 AC 계층에만 있어 이 기준의 대상이 아니다. 패턴 사용의 질에 관한 결함(D6·D7)은 MP 위반이 아닌 선택 결함으로 분류했다.
- [PASS] MP-3 YAML frontmatter: `spec.md:L2-13` 에 정규 12필드가 모두 있다. `version: "0.1.1"` 은 따옴표 semver, `created`/`updated` 는 ISO 날짜, `priority: P1`, `lifecycle: spec-anchored`, `phase: "v3.2.0 target"`(금지값 아님), `tags` 는 쉼표 문자열이다. `moai spec lint .moai/specs/SPEC-INIT-QUIET-WIZARD-001` 출력은 `✓ No findings — all SPEC documents are valid`.
- [N/A] MP-4 언어 중립성: 이 SPEC 은 템플릿에 바인딩되지 않는다. `module: "internal/cli/wizard, internal/cli"`(`spec.md:L11`)이고, §7 변경 대상에 `internal/template/templates/**` 가 없다. Go 구현 내부 전용이다.
- [PASS] MP-5 D7 교차 SPEC: 5개 산출물에서 추출한 참조 9건이 모두 존재하고, 상태는 retired·superseded·archived 가 아니다.
  - completed: CLI-WIZARD-RESTRUCTURE-001, FEEDBACK-AUTO-SUBMIT-001, INIT-HARNESS-PROMPT-001, INIT-WIZARD-REPAIR-001, MCP-DEFAULT-ON-001, MOAI-MCP-SERVER-001, PROJECT-CONTINUATION-KEY-001, TODO-ENABLE-FLAG-001
  - implemented: V3R5-INIT-WIZARD-EXPANSION-001
  - 부분 대체 관계는 `spec.md:L121-135` §5 표가 명시적으로 조정한다. BLOCKING 없음.
- [PASS] MP-6 D8 교차 플랫폼: 7개 파일 모두 `grep -c syscall` = 0. 자동 PASS.
- [PASS] MP-7 clarification gate: `grep -rn '\[NEEDS CLARIFICATION' plan.md research.md` → 매치 없음(`mp7-exit=1`). `plan.md:L215-217` §L 도 "남은 확인 항목 없음".

## Category Scores (0.0-1.0, rubric-anchored)

| Dimension | Score | Rubric Band | Evidence |
|-----------|-------|-------------|----------|
| Clarity | 0.80 | 0.75 | 인용 file:line 을 재측정했더니 전부 일치했다(아래 "재측정 확인" 참조). 모호한 곳은 세 군데다. REQ-IQW-005 "기본으로"(`spec.md:L79`)는 무엇이 기본을 덮는지 말하지 않는다. REQ-IQW-012 "어떤 파일도"(`spec.md:L86`)는 대조 8항목보다 넓다(D2). §6 "블록이 없고"(`spec.md:L141`)는 사실과 다르다(D5). |
| Completeness | 0.95 | 1.0 | HISTORY `spec.md:L18`, 배경 §1 `L23`, 범위 §2 `L36`, 요구사항 §3 `L73`, 제약 §4 `L90`, Exclusions §8 `L161` 이 모두 있다. `### Out of Scope — …` H3 7개에 각각 `-` 항목이 있다. Tier L 산출물 5종이 모두 있다. 미측정 목록 `plan.md:L203-213` 이 명시돼 있다. |
| Testability | 0.65 | 0.50–0.75 | AC-IQW-004 는 적힌 대로는 구현할 수 없다(D1). 새 테스트 이름을 겨누는 `-run` 선택자 7건에 스윕 수 확인이 없어 공허 통과 위험이 있다(D3). AC-IQW-015 마지막 판정은 0건만 걸러낸다(D8). |
| Traceability | 0.80 | 0.75 | 모든 REQ 에 AC 가 있고(`acceptance.md:L16-33`), 모든 AC 가 실재하는 REQ 를 가리킨다. 다만 AC-IQW-008 은 대응 REQ 보다 강한 성질을 단언하고(D4), REQ-IQW-012 는 AC 가 증명하는 범위보다 넓다(D2). |

집계는 조화평균(0.80, 0.95, 0.65, 0.80) ≈ **0.79** 로, Tier L PASS 기준 0.85 에 미달한다. must-pass 는 전부 통과했지만 blocking 결함이 있고 점수가 기준에 못 미쳐 FAIL 이다.

## Defects Found (structured defect-list)

D1. AC-004-UNOBSERVABLE — acceptance.md:L78-91 — AC-IQW-004 는 세 가지를 요구한다: 시접을 건드리지 않은 상태에서 `runInit` 이 만드는 `InitOptions` 의 `SkipShellConfig` 가 false 로 전달됨을 관측할 것, 시접을 켜면 true 로 전달됨을 관측할 것, 그리고 시접 도입 커밋이 테스트 파일 추가 커밋의 조상일 것. 그러나 `runInit` 은 이 값을 밖에서 볼 틈을 주지 않는다. — Severity: major — Class: blocking — Required fix: 아래 둘 중 하나로 고친다.
  - 관측 불가인 이유:
    - `internal/cli/init.go:832-833` 이 `project.NewInitializer`·`project.NewPhaseExecutor` 를 인라인으로 만들고 `:867` 에서 `executor.Execute(ctx, opts)` 로 넘긴다.
    - `internal/cli` 의 함수 변수 시접 목록(`runWizardFn`·`userHomeDirFn` 등)에 opts 를 가로채는 시접이 없다. design.md §4.2(L95-103)와 plan.md M1 파일 목록(L53-56)도 `skipShellConfigForInit` 하나만 선언한다.
    - 간접 관측(실제 셸 설정 파일 불변)은 공허해질 수 있다. `internal/shell/config.go:69,93,167` 이 이미 설정된 줄이 있으면 `Skipped` 로 건너뛰므로, 이미 설정된 머신에서는 시접 여부와 무관하게 불변이 나온다.
    - 기본값(false) 전달을 실행으로 관측하면 실제 `~/.zshenv` 에 쓴다. `internal/shell/detect.go:128` 이 `os.Getenv("HOME")` 을 읽는데, §4.2 1항은 새 테스트의 `t.Setenv("HOME")` 을 금지한다. SPEC 스스로의 안전 불변식과 충돌한다.
    - 조상 판정 `git merge-base --is-ancestor A B` 는 A=B(같은 커밋)여도 0 을 돌려준다. 따라서 "먼저 착지"를 증명하지 못한다.
  - 수정 (a): opts 관측 시접(예: 실행기 생성 함수 변수)을 design.md §4.2, plan.md M1 파일 목록, spec.md §7 [NEW] 행에 추가한다.
  - 수정 (b): AC-IQW-004 를 두 절로 다시 쓴다.
    - 소스 도달성 가드: `init_mcp_provision_test.go:105-124` 선례처럼 `init.go` 에 `SkipShellConfig: skipShellConfigForInit()` 문자열이 있음을 단언한다.
    - 시접 기본값 직접 읽기: `skipShellConfigForInit()==false`.
    - 각 절이 실행 관측인지 소스 관측인지 명시한다.
  - 두 경우 모두 조상 판정에는 "두 해시가 다름" 조건을 더한다.

D2. REQ012-SCOPE-EXCEEDS-VERIFICATION — spec.md:L86 (REQ-IQW-012), spec.md:L106 (§4.2 5항), acceptance.md:L97 — REQ-IQW-012 는 새 init 실행 테스트가 "실제 사용자 홈의 어떤 파일도 바꿔서는 안 된다"고 절대형으로 요구한다. 그러나 이를 증명하는 AC-IQW-005·015 는 8항목만 대조한다. plan.md §K(`L207-208`)는 `NewEnvConfigurator` 본문과 지연 업데이트 확인 경로를 미추적으로 인정하므로, 8항목 밖의 홈 쓰기는 요구 위반인데도 어떤 AC 로도 잡히지 않는다. — Severity: major — Class: blocking — Required fix: REQ-IQW-012 를 "§4.2 5항의 8항목을 바꿔서는 안 된다"로 좁혀 AC 가 증명하는 범위와 맞춘다. 또는 8항목 밖 경로를 대조 목록에 넣고 AC-IQW-005 기대값을 갱신한다. 좁힐 경우 나머지 경로는 plan.md §K 미측정으로 남는다는 문장을 REQ 옆에 둔다.

D3. EMPTY-SWEEP-UNGUARDED — acceptance.md:L44, L54, L85, L100, L131, L141, L169, L179, L208 — 새로 작성할 테스트 이름을 겨누는 `go test -run` 선택자에 스윕 수 확인이 없다. 대상은 `TestInitQuestions_QuietSet`, `TestRemovedQuestions*`, `TestInitShellConfigSeam`, `TestHomeGuard`, `TestRunInit_QuietWizard*` 이다. 이 이름들은 design.md 가 "제안"이라 하고 spec.md §4.1(L94)이 "run 단계가 확정한다"고 한 이름이다. run 단계에서 이름이 조금만 달라져도 `[no tests to run]` 과 함께 `ok`·종료 코드 0 이 나오고, 그 줄은 모든 테스트가 통과한 출력과 구별되지 않는다. `.claude/rules/moai/development/verification-completeness.md §1.1`("A pass whose swept set is empty asserts nothing")에 해당한다. 기존 테스트를 겨누는 선택자는 재측정에서 모두 정의가 있었다(`TestSeedMirrorsProductionLSPSeed` 등 11개, 각 defs≥1). — Severity: major — Class: blocking — Required fix: 새 테스트 선택자 AC 마다 `-v` 를 붙이고, 기대값에 두 조건을 추가한다: (1) 선택자가 지목한 각 테스트 이름의 `--- PASS: <이름>` 줄이 출력에 있을 것, (2) 출력에 `[no tests to run]` 이 없을 것. AC-IQW-004·005·014 에도 같은 조건을 넣는다.

D4. AC008-STRONGER-THAN-REQ — acceptance.md:L26, L162-170 — AC-IQW-008 은 대화형과 비대화형 섹션 파일 5개의 **바이트 동일**을 단언하며 REQ-IQW-003 에 매핑돼 있다. 그런데 REQ-IQW-003(`spec.md:L77`)이 요구하는 것은 **해석된 값**의 동일성이고, 디스크 모양 변화는 spec.md §6(`L137-145`)이 "받아들인 변화"로 서술할 뿐 규범 요구로 두지 않았다. 대응하는 REQ 가 없어서 뒤에 누가 이 AC 를 완화하거나 삭제해도 요구사항 계층에서 막을 근거가 없다. — Severity: minor — Class: blocking — Required fix: §6 의 모양 동일성을 요구사항으로 올린다. 예: REQ-IQW-015 (Event-driven) "When 대화형 init 이 유지 4문항만 답하고 끝나면, 섹션 파일 5개는 플래그 없는 비대화형 init 의 것과 바이트 동일해야 한다". 그 뒤 AC-IQW-008 을 그 REQ 에 매핑한다. 반대로 AC-008 을 해석값 수준으로 약화해도 된다.

D5. SPEC6-AUDIT-BLOCK-MISSTATEMENT — spec.md:L141, acceptance.md:L166 — spec.md §6 은 "정리 뒤에는 [workflow.audit] 블록이 없고", AC-IQW-008 은 "`audit:` 블록 ... 이 삽입되지 않는다"고 적었다. 그러나 템플릿 `internal/template/templates/.moai/config/sections/workflow.yaml:85-91` 은 이미 `audit:` 키(`codex`/`glm` 의 model·effort 핀)를 싣고 배포된다. 사라지는 것은 `model`·`gates` 하위 키일 뿐 `audit:` 키 자체가 아니다. run 단계가 이 문장을 그대로 믿고 `audit:` 부재를 단언하면 올바른 구현에서도 테스트가 빨개진다. — Severity: minor — Class: blocking — Required fix: spec.md:L141 을 "`workflow.audit` 아래 `model`·`gates` 키가 없고(템플릿이 싣는 `codex`/`glm` 핀은 남음)"로, acceptance.md:L166 을 "`audit.model`·`audit.gates` 키와 `todo:` 키가 삽입되지 않는다"로 고친다.

D6. REQ003-ABSOLUTE-VS-KNOWN-EXCEPTION — spec.md:L77, acceptance.md:L295 — REQ-IQW-003 은 "해석된 값은 오늘 모든 질문에 기본값을 받아들인 사용자가 얻는 값과 같아야 한다"고 예외 없이 요구한다. 그런데 acceptance.md §3 은 파일시스템 루트에서 init 할 때 `project_name` 이 오늘의 `my-project` 치환과 달라질 수 있음을 인정하고 측정하지 않는다. 요구와 인정된 사례가 모순된다. — Severity: minor — Class: blocking — Required fix: REQ-IQW-003 에 예외절을 명시한다(예: "단, 프로젝트 루트가 파일시스템 루트인 경우의 `project_name` 은 제외"). 또는 해당 사례를 측정하는 AC 를 추가한다.

D7. REQ-LAYER-HOW-LEAKAGE — spec.md:L83, L85, L88 — 요구사항 계층에 구현·검증 방법이 섞였다(RQ-3/RQ-4).
  - REQ-IQW-011: "함수 변수 형태"라는 구현 형식을 지정한다. 조건 "테스트가 init 을 실행하면"은 GEARS `Where`(능력 게이트·기능 플래그·정적 설정)가 아니라 상황 조건이다.
  - REQ-IQW-009: "위저드 시접에 ... 주입해 실제 `runInit` 을 돌리고"라는 테스트 절차를 요구문에 담았다.
  - REQ-IQW-014: 두 개의 `When` 의무(채집·기록, 멈춤·보고)를 한 ID 에 묶었다.
  - — Severity: minor — Class: optional — Required fix: REQ-011 은 "셸 설정 쓰기를 억제할 수 있는 테스트 전용 수단을 제공해야 한다"로 무엇만 남기고 형식은 design.md §4.2 로 옮긴다. REQ-009 의 절차는 AC-IQW-006 으로 옮긴다. REQ-014 는 채집 의무와 멈춤 의무 두 REQ 로 나눈다(Tier L 상한 25 안에 여유 있음).

D8. AC015-WEAK-TERMINAL-CHECK — acceptance.md:L276-279 — 마지막 판정은 "`command grep -c 'home-diff-exit=' progress.md` 가 기록된 슬롯 수 이상"인데, "기록된 슬롯 수"를 독립적으로 세는 수단이 없다. 실제로 걸러내는 경우는 0건뿐이다. 슬롯 하나를 빠뜨려도 통과한다. — Severity: minor — Class: optional — Required fix: 슬롯 목록(슬롯 이름 열거)을 progress.md §E.2 표로 먼저 선언하게 한다. 그리고 `home-<slot>-after.out` 파일 수 = 선언 슬롯 수 = `home-diff-exit=` 기록 수 삼자 일치를 판정식으로 둔다.

## 재측정 확인 (PASS 근거로 쓴 관측)

기준 트리: HEAD `120436f58`(`git rev-parse --short HEAD`), `internal/` 미수정.

| 대조 대상 | 명령(요지) | 관측 | SPEC 기재 |
|---|---|---|---|
| AC-IQW-002 대조군 | `git grep -cE '"(project_mode|…|mcp_provision)"' 120436f58 -- internal/cli/wizard ':!*_test.go'` | questions.go:11, translations.go:33, wizard.go:11 | 일치 |
| AC-IQW-003 추출물 | `sed -n '/^func TestReconfigureQuestionsOrder/,/^}/p' … | wc -l` | 52 | 일치 |
| AC-IQW-005 대조군 | `git grep -c 't.Setenv("HOME"' 120436f58 -- internal/cli/init_agent_wizard_test.go` | 7 | 일치 |
| AC-IQW-010 추출물 | 두 함수 `sed` 추출 `wc -l` | 20, 12 | 일치 |
| AC-IQW-012 대조군 | `git grep -cE 'worktree_auto_create|WorktreeAutoCreate' 120436f58 -- …` + `sed` 범위 wizard 계수 | init.go:2, questions.go:2, translations.go:3, types.go:1, wizard.go:2; `wizard-mentions=2` | 일치 |
| AC-IQW-013 유지 항목 | `git grep -nE 'func writeProjectModeYAML|…|opts\.MCPProvision' 120436f58 -- internal ':!*_test.go'` | 5줄 (init.go:252·337·1004, initializer_expansion.go:205, initializer_workflow_toggles.go:38) | 일치 |
| F1 경로 | `init.go:305-307`, `init_workflow_flags.go` `applyWorkflowBranchGuardFlags` | 위저드는 값만, 추적자는 플래그 분기에서만 설정 | 일치 |
| MCP 규칙 | `init.go:999-1011` | 결정 B1 주석, `mcpDeclined := !opts.MCPProvision`, codex/both 스위치 | 일치 |
| reconfigure 순서 | `questions.go` `ReconfigureQuestions` | Git 을 `report_format` 뒤에 삽입 = Default 5 + Git 7 | REQ-IQW-004 와 일치 |
| AC-IQW-008 위험 요소 | `project.yaml.tmpl:17` `created_at: "{{.CreatedAt}}"` / `internal/template/context.go` | 운영 코드에 `WithCreatedAt` 호출 0건, `time.Now` 없음 → init 경로에서 빈 값 | 타임스탬프로 인한 바이트 불일치 없음 |
| codex/both 홈 쓰기 | `internal/codexwiring` 의 `os.UserHomeDir`/`HOME` 검색 | 0건 | 누출 경로 아님 |

## Gaps (관측하지 않은 것)

- 컴파일·테스트·바이너리 실행 없음(리드 지시). AC 명령의 실행 가능성은 판독으로만 판정했다.
- `startDeferredUpdateNotice`, `installPrePushHookOptional`, `ScaffoldEvolutionDir`, `NewEnvConfigurator` 가 실제 홈에 쓰는지 추적하지 않았다(D2 의 근거이며, SPEC 도 plan.md §K 에서 같은 공백을 인정한다).
- `internal/core/project` 기존 테스트가 셸 설정 단계까지 도달하는지 확인하지 않았다(SPEC 의 §K 공백과 동일).
- `.moai/reports/t583/verdict.md` 의 F1 재현·F4 반증 측정값은 다시 재지 않았다. SPEC 이 인용한 대로만 받았다.

## Residual-risk

- D1 을 (b) 소스 가드로 고치면, 시접이 실제로 셸 설정 쓰기를 막는지는 실행 슬롯 지문(AC-IQW-015)에만 기대게 된다. 셸 설정이 이미 되어 있는 머신에서는 그 지문도 불변으로 나오므로(idempotent skip), 이 머신의 슬롯 지문은 시접의 효과를 증명하지 못한다.
- AC-IQW-008 의 바이트 동일은 판독상 성립한다. autonomy 번들은 `settings.json`만, `SyncToProjectConfig` 는 `statusline.yaml`만 쓴다. 다만 LSP·품질·디자인 기록기가 비교 대상 5파일에 손대지 않는다는 점은 개별로 추적하지 않았다.

## Recommendation

1. **D1** — acceptance.md:L78-91 AC-IQW-004 를 고친다. opts 관측 시접을 설계와 파일 목록에 선언하거나, 소스 도달성 가드와 시접 기본값 직접 읽기로 다시 쓴다. 조상 판정에는 "두 해시가 다름"을 더한다.
2. **D2** — spec.md:L86 REQ-IQW-012 의 범위를 §4.2 5항 8항목으로 좁히거나, 대조 목록을 넓혀 AC-IQW-005·015 와 맞춘다.
3. **D3** — 새 테스트를 겨누는 `-run` AC 전부에 `-v` 를 붙이고, `--- PASS: <이름>` 존재와 `[no tests to run]` 부재를 기대값에 추가한다(verification-completeness.md §1.1).
4. **D4** — §6 모양 동일성을 새 REQ(예: REQ-IQW-015)로 올리고 AC-IQW-008 을 그 REQ 에 매핑한다.
5. **D5** — spec.md:L141 과 acceptance.md:L166 의 `audit:` 서술을 "`model`·`gates` 하위 키 부재"로 정정한다(템플릿 workflow.yaml:85-91 참조).
6. **D6** — spec.md:L77 REQ-IQW-003 에 파일시스템 루트 예외절을 둔다.
7. D7·D8 은 선택 사항이다. REQ-011/009 에서 구현 방법을 걷어내고 REQ-014 를 분할하는 일, AC-015 에 슬롯 목록 삼자 일치 판정을 두는 일은 오케스트레이터 재량이다.

D1~D6 을 반영하면 재감사는 이 결함 목록만 대상으로 한다(Retry Loop Contract, Tier L 상한 3회).
