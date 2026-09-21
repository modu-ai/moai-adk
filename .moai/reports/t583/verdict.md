# t583 재현 판정서 — init 정온화 선행 재현 (F1·F4)

- 카드: t583 (Class C, Tier M~L)
- 워크트리: `.claude/worktrees/t583` · 브랜치 `WT-init-quiet-wizard`
- 기준 트리: origin/develop `d060e0d13` 에서 생성 → `git merge --no-ff 93182d137` → 병합 커밋 `120436f58` (`HEAD^2` = `93182d137` 확인)
- 근거 보고서 발췌: `.moai/reports/t583/audit-excerpt.md` (원본 sha256 `bde708a712fa37b7efb849f6ffa35ca366fafe260da08e60845f8fa0830d8f84`)
- 재현 코드·출력: `.moai/reports/t583/repro/t583_repro_test.go`, `.moai/reports/t583/repro/go-test-t583repro.txt`

## 1. 재현 방식과 그렇게 한 이유

리드 지시에 따라 바이너리로 `moai init` 을 실행하지 않았다. init 꼬리의 `ensureGlobalSettingsEnv` 가 `/tmp` 프로젝트에서 돌려도 실제 `~/.claude/settings.json` 을 정리하고 `~/.claude/hooks/moai` 를 지우며, `HOME=` 격리는 세션 가드가 거부하기 때문이다.

대신 `internal/cli` 패키지 안에서 실제 코드 경로를 도는 Go 테스트 3개를 한 번 실행했다(리드 승인 슬롯 1회).

- 홈 우회: `userHomeDirFn` 교체 + `profile.BaseDirOverride` + `MOAI_HOME` 를 모두 `t.TempDir()` 로. `t.Setenv("HOME")` 은 쓰지 않았다.
- 실행 전 자기 가드: 세 우회 경로가 실제 `os.UserHomeDir()` 안으로 풀리면 `runInit` 전에 `t.Fatal`.
- 실행 전 백업: 실제 `~/.claude/settings.json` 을 세션 스크래치로 읽기 복사. `~/.claude/hooks/moai` 는 실행 전부터 없었다.
- 실행 전후 대조: 테스트 안 sha256 비교와, 실행 뒤 셸에서 따로 잰 sha256.
- 실행 뒤 테스트 파일은 `internal/cli` 에서 보고서 디렉터리로 옮겼다. 트리에 RED 테스트는 남아 있지 않다(`git status --porcelain` → `?? .moai/reports/t583/` 한 줄).

## 2. 판정

| 결함 | 보고서 주장 | 판정 | 핵심 관측 |
|---|---|---|---|
| F1 | 위저드의 워크트리 "예" 답변이 버려진다 | **재현됨** | 좁은 경로·전체 runInit 경로 모두 `auto_create: false` 로 배포 |
| F4 | `--non-interactive` 이면 MCP 가 설치되지 않는다 | **재현되지 않음 — 전제가 틀림** | 배포된 `.mcp.json` 의 `mcpServers` 키가 `[moai context7]` |

### F1 — 재현됨

**주장**: 위저드에서 `worktree_auto_create` 에 "예"를 골라도 `workflow.yaml` 에 기록되지 않는다.

**증거** (명령: `go test ./internal/cli -run 'TestT583Repro' -count=1 -v -timeout 590s`, 종료 코드 1 — F1 두 테스트의 의도된 FAIL):

```
=== RUN   TestT583Repro_F1_Narrow
    t583_repro_test.go:138: after applyWizardPage3ToOpts: WorktreeAutoCreate=true WorktreeAutoCreateSet=false
    t583_repro_test.go:151: F1 reproduced (narrow): wizard answered yes, workflow.yaml auto_create is not true
--- FAIL: TestT583Repro_F1_Narrow (0.00s)
=== RUN   TestT583Repro_F1_RunInit
    t583_repro_test.go:161: deployed workflow.yaml line: "        auto_create: false"
    t583_repro_test.go:165: F1 reproduced (runInit): wizard answered yes, deployed auto_create is not true
--- FAIL: TestT583Repro_F1_RunInit (1.11s)
```

**원인 (코드 판독, 위 관측과 일치)**:
- `internal/cli/init.go:305-307` — 위저드 답을 `opts.WorktreeAutoCreate` 에만 넣고 `opts.WorktreeAutoCreateSet` 은 켜지 않는다.
- `internal/core/project/initializer_workflow_toggles.go:38-54` — `*Set` 추적자가 켜진 키만 쓴다. 추적자를 켜는 곳은 플래그 경로 `internal/cli/init_workflow_flags.go:40-44` 뿐이다.
- 기존 회귀 테스트 `internal/cli/init_workflow_wiring_test.go:100-130` 은 위저드 답을 `false` 로만 넣는다. `false` 는 템플릿 기본값과 같아서, 답이 버려져도 통과한다. 보고서의 "테스트가 false 만 재어 미탐지"와 일치한다.

보고서의 행번호(`init_workflow_flags.go:41-44`)는 현재 트리에서도 거의 같은 자리다.

### F4 — 재현되지 않음

**주장(보고서)**: `--non-interactive` init 은 MCP 를 프로비저닝하지 않는다.

**증거** (같은 실행):

```
=== RUN   TestT583Repro_F4_NonInteractive
    t583_repro_test.go:175: ensure-entry announcement on stdout: false
    t583_repro_test.go:192: .mcp.json mcpServers keys: [moai context7]
--- PASS: TestT583Repro_F4_NonInteractive (0.94s)
```

**해석**:
- `internal/template/templates/.mcp.json` 에 이미 `mcpServers.moai` (`command: moai`, `args: [mcp-server]`) 가 들어 있고, 비대화형 init 도 이 파일을 그대로 배포한다. 따라서 신규 프로젝트에서 "MCP 미설치"는 일어나지 않는다.
- 실제로 건너뛰는 것은 `internal/cli/init.go:1004-1011` 의 ensure-entry 호출과 그 안내 문구뿐이다(`opts.MCPProvision` 의 유일한 쓰기가 대화형 블록 안 `init.go:337` 이라 비대화형에서는 0값 false → `mcpDeclined=true`).
- 이 동작은 기존 테스트 `internal/cli/init_agent_wizard_test.go:172-205` (`TestRunInit_FlagAbsentNonInteractivePreservesCodexAbsence`) 가 **의도된 동작으로 고정**하고 있다. 같은 주석이 "파일 안 moai 항목 존재"를 AC-CW-004 로 별도 주장한다.

보고서가 틀렸다기보다, 호출이 생략된다는 코드 판독에서 "설치되지 않는다"는 결과까지 건너뛴 추론이다. 실행으로 재보니 결과는 달랐다.

## 3. 기준(Baseline) 귀속

- 모든 관측은 이번 실행, 워크트리 HEAD `120436f58` 트리에서 잰 것이다.
- 실제 홈 대조:
  - 실행 전: `shasum -a 256 ~/.claude/settings.json` → `86e2d9b63abd4d027b4f65bcdc3b41e127b55c9d65c5c8061d17f36b247cc5eb`, `~/.claude/hooks/moai` 없음
  - 테스트 안 기록: `settings.json sha256 before=86e2d9b6… after=86e2d9b6…; hooks/moai before=false after=false` (runInit 을 도는 두 테스트 모두)
  - 실행 뒤 셸 재측정: `86e2d9b63abd4d027b4f65bcdc3b41e127b55c9d65c5c8061d17f36b247cc5eb`, `hooks/moai absent`
- 백업 사본: 세션 스크래치 `t583-home-backup/settings.json` (sha256 동일). 복구가 필요한 상황은 생기지 않았다.

## 4. 미검증 (Gaps)

- 바이너리 실행 경로(실제 TTY 위저드, huh 렌더)는 재지 않았다 — 리드 지시로 금지. 위저드 결과는 `runWizardFn` 주입으로 대체했다.
- F1 전체 경로 테스트에는 "기록이 되면 `auto_create: true` 를 읽어 낸다"는 같은 실행 안의 양성 대조군이 없다. 쓰기 경로 자체는 기존 `TestRunInit_WorkflowToggleFlagsPersist` 가 덮지만, 이번 실행에서 그 테스트는 돌리지 않았다.
- F4 는 **신규** init 만 쟀다. 이미 `.mcp.json` 이 있는 디렉터리(`--force` 재초기화, 사용자가 만든 `.mcp.json`)에서 배포기가 파일을 덮는지·건너뛰는지는 재지 않았다. ensure-entry 호출의 생략이 실제 차이를 낳을 수 있는 곳은 그쪽이다.
- 홈 우회는 `runInit` 이 닿는 쓰기 경로 중 실행 당시 찾은 것(`userHomeDirFn`·`profile.GetBaseDir`·`MOAI_HOME`)만 막았다. 읽기 전용 해석(`initializer.go:394` `os.UserHomeDir`, `root.go:55` `paths.Home`)은 실제 홈을 읽는다. 실제 홈의 다른 파일(`~/.moai/**` 등)은 전후 대조하지 않았다.
- **사후에 찾은 누출 경로 — 셸 rc 추가 쓰기**: 실행 뒤 조사에서 `internal/core/project/initializer.go:333-347` 의 `configureShellEnv` 가 `SkipShellConfig` 가 꺼져 있으면(`runInit` 은 이 값을 켜지 않는다) `$HOME` 기준 셸 설정 파일에 줄을 **덧붙인다**는 것을 확인했다(`internal/shell/detect.go:128-145`: zsh·darwin → `~/.zshenv`). 이 경로는 세 우회 장치 어느 것에도 막히지 않는다. 사후 확인으로 실행 시각과 rc 파일 수정 시각을 비교했다.
  - 명령: `stat -f '%m %Sm %N' .moai/reports/t583/repro/go-test-t583repro.txt ~/.zshenv ~/.zshrc ~/.zprofile ~/.profile ~/.bashrc ~/.bash_profile ~/.config/fish/config.fish`
  - 출력: 실행 기록 파일 `Sep 10 23:19:26 2026`, `~/.zshenv`·`~/.zshrc` `Aug 16 17:02:58 2026`, `~/.zprofile` `Aug 31 15:46:57 2026`, `~/.profile`·`~/.bashrc`·`~/.bash_profile` `May 24 13:27:24 2026`, fish 설정은 없음
  - 판독: 모든 rc 파일의 수정 시각이 실행보다 앞선다. 이번 실행이 rc 파일을 고치지 않았다는 사후 근거다. 다만 줄이 이미 있어서 건너뛴 결과일 수 있으며, 줄이 없는 홈(CI·새 홈)에서는 같은 테스트가 실제 rc 에 쓴다. 이후 "미설정=기본값" 실행 테스트는 셸 설정 쓰기를 막는 시접이 선행 조건이다.

## 5. 잔여 위험

- 실제 `~/.claude/settings.json` 을 다른 세션이 같은 시각에 고쳤다면 해시 대조가 누출을 가리거나 거짓 경보를 낼 수 있다. 이번에는 전후 해시가 같았다.
- F1 원인은 코드 판독과 관측이 맞아떨어지지만, 수리 방식(위저드 경로에서 추적자를 켤지, 질문 자체를 없앨지)은 설계 판단이라 여기서 정하지 않았다.

## 6. 후속 카드 전제에 미치는 영향

- **t583 자체 범위 — 카드 전제 정정**: F4 는 재현되지 않았다. 신규 init 도 템플릿 `.mcp.json` 으로 `moai`·`context7` 를 배포한다(§2 F4 증거: 배포 파일 측정, `init_agent_wizard_test.go:172-205` 의 의도 고정). 기존 `.mcp.json` 이 있는 경우는 미측정이며 §4 Gaps 에 남긴다.
- **운영자 결정 (2026-09-10, 리드 세션에서 운영자가 직접 답한 것을 리드가 전달)**: `--no-mcp` 는 범위에서 뺀다. 이에 따라 SPEC 범위는 **16→4 질문 축소 + F1 수리** 로 확정했고, `--no-mcp` 에 대한 NEEDS CLARIFICATION 표식은 해소됐다. 이후 테스트에도 이번과 같은 홈 안전장치(시접 우회·실행 전 자기 가드·전후 sha256 대조)를 유지한다.
- **F1**: 운영자 결정은 워크트리 질문을 "F1 수리 후 web 으로" 옮기는 것이다. 질문을 없애면 위저드 경로의 버그 자체가 사라지므로, 수리 대상은 web 쪽 기록 경로와 update 재구성 경로(`runWorkflowConfigStep`)의 동등성 검증으로 옮겨 간다.
- **t584 (자율 모드 재정의)**: 전제 변화 없음. 이번 재현은 자율성 등급 경로를 건드리지 않았다.
- **t585 (하네스 3-way)**: `.mcp.json` 이 하네스 선택과 무관하게 템플릿으로 무조건 배포된다는 보고서 핵심 발견 5 가 이번 관측과 일치한다. codex 단독 필터 설계에서 이 파일을 빼야 한다는 전제는 그대로 유효하다.
- **t588 (정합성)**: 보고서의 결함 목록에서 F4 를 "비대화형은 ensure-entry 안내가 없다" 수준으로 낮춰 적어야 한다.

## 7. plan 조사 뒤 확정된 전제 (2026-09-10, 리드 결정)

조사 원문: `.moai/reports/t583/research-explore.md` (읽기 전용 판독, 실행 없음).

### 7.1 전제 정정

- init 질문은 16개가 아니라 **18개**다(`internal/cli/wizard/questions.go:296-303` — `DefaultQuestions` 5 + `Page3Questions` 13). 이 카드는 18→4, 14개를 제거한다. 남기는 4개: `conversation_language`·`user_name`·`autonomy_tier`·`agent_wiring`.
- F1 은 워크트리 질문을 없애면 init 경로에서 증상 자체가 사라진다. "F1 수리"의 실체는 질문 제거, 서로 모순되는 문서 3곳(`internal/core/project/initializer.go:57-62`·`internal/cli/init.go:300-304`·`internal/cli/wizard/questions.go:372`) 정리, 위저드 생략 경로 실행 테스트로 재발 봉인이다.

### 7.2 리드 결정

- **D1 = (a)**: 질문 생성자를 나눠 `moai update -c` reconfigure 는 현행 질문을 유지한다. 이 카드는 init 만 바꾼다.
- **D2 = (a)**: `mcp_provision` 질문을 없앤 뒤에도 대화형 경로는 코드에서 프로비저닝 기본값을 true 로 둬, 오늘 "기본값 수락" 사용자의 동작을 보존한다. 비대화형은 `internal/cli/init_agent_wizard_test.go:172-205` 가 고정한 현행 동작을 유지한다. `--force` 재초기화와 moai 항목 없는 기존 `.mcp.json` 은 미측정 Gap 으로 SPEC 에 남긴다.
- **안전 (승인)**: 셸 설정 쓰기를 막는 시접을 run 단계 선행 조건으로 SPEC 에 넣는다. 모든 init 실행 테스트는 다음을 갖춘다 — 실행 전 읽기 백업, 시접이 돌려주는 홈 ≠ 실제 `os.UserHomeDir()` 단언(같으면 `t.Fatal`), 실행 전후 대조 목록: `~/.claude/settings.json` sha256, `~/.claude/hooks/moai` 존재, `~/.zshenv`·`~/.zshrc`·`~/.zprofile`·`~/.profile`·`~/.bashrc` 의 mtime 과 sha256.

### 7.2.1 리드 결정 — 홈 안전 점검표 적용 범위 (2026-09-11)

- **결정 (a)**: 점검표를 코드로 넣는 대상은 이 SPEC 이 새로 쓰거나 본문을 다시 쓰는 init 실행 테스트뿐이다. 제거 필드 참조만 지우는 기존 테스트(`internal/cli/init_agent_wizard_test.go:64-160`, `internal/cli/doctor_codex_e2e_test.go`, `internal/cli/init_workflow_wiring_test.go:100-134` 등)는 기존 `t.Setenv("HOME")` 헬퍼를 그대로 쓴다.
- **실행 절차는 소급한다**: 이 SPEC 의 run 단계에서 기존 init 실행 테스트를 하나라도 돌리는 슬롯은, 실행 전후에 실제 홈 지문을 뜬다.
  - 대상: `~/.claude/settings.json` sha256, `~/.claude/hooks/moai` 존재, rc 6종(`~/.zshenv`·`~/.zshrc`·`~/.zprofile`·`~/.profile`·`~/.bashrc`·`~/.bash_profile`)의 mtime 과 sha256
  - 전후에 **같은 명령**을 쓰고 stderr 를 분리한다. 카드 t661 에서 전후 도구가 달라 가짜 diff 가 난 전례가 있다.
- **판독의 실행 확인**: "기존 HOME 헬퍼는 셸 감지(`internal/shell/detect.go:128`)도 임시 홈을 따르므로 셸 설정 누출이 없다"는 지금 코드 판독일 뿐이다. run 단계의 첫 실행 지문이 이 판독의 실행 확인이 된다.
- **후속 카드 후보 (미발행)**: 기존 init 테스트 헬퍼(`runInitForAutonomyAtHomeCapturingOut`, `internal/cli/init_autonomy_wiring_test.go:40`)의 `t.Setenv("HOME")` 방식을 시접 기반 홈 안전 헬퍼로 옮기는 정리. 큐가 복구된 뒤 리드가 발행한다(이 레인은 `moai todo` 를 쓰지 않는다).

### 7.2.2 가설 — `internal/core/project` 테스트의 실제 홈 셸 rc 쓰기 (판독 근거, 실행 확인 없음)

**주장 (가설)**: `internal/core/project` 패키지 테스트는 `SkipShellConfig` 없이 `Init()` 을 부르므로, 해당 줄이 없는 홈에서 실제 셸 설정 파일에 줄을 덧붙인다.

**근거 (판독, 이번 트리 HEAD `120436f58`)**:
- `Init()` 을 부르는 테스트: `internal/core/project/initializer_test.go:82·113·166·193·229`, `initializer_audit_wiring_test.go:38·63`, `initializer_persist_test.go:109` 등.
- `grep -rln 'SkipShellConfig' internal/core/project/*_test.go; echo "skip-exit=$?"` → `skip-exit=1`(참조 0건). 같은 패키지 테스트에서 `HOME` 문자열을 찾는 grep 은 한 줄도 출력하지 않았다.
- `Init()` 의 Step 6(`internal/core/project/initializer.go:333-347`)은 `SkipShellConfig` 가 false 면 `configureShellEnv`(`:677-686`)를 부르고, `internal/shell/env.go:124-175` 가 `$HOME` 기준 rc 파일에 `CLAUDE_DISABLE_PATH_WARNING`·`$HOME/.local/bin`·`$HOME/go/bin` 을 덧붙인다. 줄이 이미 있으면 `ErrAlreadyConfigured` 로 건너뛴다.
- 이 머신: `grep -c 'CLAUDE_DISABLE_PATH_WARNING\|\.local/bin\|go/bin' "$HOME/.zshenv"` → `4`.

**"이미 있으면 건너뜀"이 뜻하는 것**: 이 머신에서 가설이 드러나지 않는 이유가 바로 이 건너뜀이다. 같은 이유로, 이 머신의 로컬 초록은 CI 러너나 새 머신에서 실제 rc 가 안전하다는 것을 말해 주지 않는다. 줄이 없는 홈에서만 쓰기가 일어나기 때문이다.

**처분 (리드, 2026-09-11)**: 후속 카드로 받는다. t661(`internal/cli` 홈 격리)과 같은 격리 결함 계열이며, 큐 저장소 이전 뒤 발행 목록에 올랐다. t583 범위에서는 수리하지 않는다. run 단계에서 `./internal/core/project/...` 를 돌리는 슬롯의 **첫 실행 전후 홈 지문**(REQ-IQW-014 / AC-IQW-015 절차)을 이 가설의 실측 근거로 이 문서에 남긴다.

### 7.3 t588 입력 — reconfigure 가 report_format 답을 저장하지 않는다

**주장**: `moai update -c` 의 위저드는 `report_format` 을 묻지만(`ReconfigureQuestions` = `DefaultQuestions` + `GitQuestions`, `questions.go:268-295`), 그 답을 디스크에 쓰는 코드가 없다.

**증거 (코드 판독)**:

```
$ grep -n 'ReportFormat\|ProjectName' internal/cli/update_wizard.go; echo "grep-exit=$?"
grep-exit=1

$ grep -rn 'ReportFormat' --include='*.go' internal/cli | grep -v _test.go
internal/cli/init.go:748:		if opts.ReportFormat == "" && result.ReportFormat != "" {
internal/cli/init.go:749:			opts.ReportFormat = result.ReportFormat
internal/cli/wizard/types.go:30:	ReportFormat string // Report output format: html+md, md
internal/cli/wizard/wizard.go:417:		result.ReportFormat = value
```

`WizardResult.ReportFormat` 를 읽는 비테스트 코드는 init 경로(`init.go:748-749`)뿐이다. reconfigure 의 적용 함수 `applyWizardConfig`(`update_wizard.go:133-373`)는 이 필드를 한 번도 읽지 않는다. `ProjectName` 도 같다.

**Gap**: 실행 재현은 하지 않았다. `moai update -c` 는 홈에 쓰므로 바이너리 실행은 금지다. t588 의 재현은 `applyWizardConfig` 를 `t.TempDir()` 프로젝트에 대해 `ReportFormat: "md"` 로 호출한 뒤 `.moai/config/sections/report.yaml` 이 `html+md` 로 남는지 관측하는 Go 테스트가 알맞다(이 문서의 홈 안전장치 그대로).

### 7.4 t586 과의 충돌 회피 — wizard 변경 예정 구간

- `internal/cli/wizard/questions.go`: `Page3Questions`(349-527) 중 `agent_wiring`·`autonomy_tier` 외 전부 제거 대상, `InitQuestions`(296-303) 구성 변경, D1(a)에 따라 init 전용 생성자 신설(`DefaultQuestions`·`ReconfigureQuestions` 는 유지).
- `internal/cli/wizard/wizard.go`: `saveAnswer`/`saveBoolAnswer` 분기(397-482), `RunWithDefaults` 시드(35-52), 그룹·진행 표시 분모(176-238)에 영향.
- `internal/cli/wizard/translations.go`: init 전용이 되는 11개 ID(`project_mode`·`worktree_auto_create`·`todo_enabled`·`feedback_auto_submit`·`project_continuation`·`audit_model`·`audit_gate_claude`·`audit_gate_codex`·`audit_gate_glm`·`codex_audit_enabled`·`mcp_provision`) × ko/ja/zh. `project_name`·`model_policy`·`report_format` 는 reconfigure 가 유지하므로 번역도 유지.
- `internal/cli/init.go`: `applyWizardPage3ToOpts`(270-339), 위저드 결과 적용 블록(728-754), MCP 기본값(1004).
- 확정본은 `.moai/specs/SPEC-INIT-QUIET-WIZARD-001/plan.md` §F.1 이다. 2026-09-11 lane-2(t586)에 그대로 전달했고 수신 확인을 받았다. lane-2 는 t583 병합 전에는 `questions.go`·`wizard.go`·`types.go`·`translations.go`·`init.go` 를 건드리지 않는다. "Quality & Workflow" 그룹에 `agent_wiring` 하나만 남는 그룹 정리는 t586 범위로 가져갔다.

## 8. plan 감사 처분 — PASS-with-debt (2026-09-11, 리드·운영자 결정)

### 8.1 감사 추이

| 회차 | 보고서 | 판정 | 점수 | 남은 blocking |
|---|---|---|---|---|
| 1 | `.moai/reports/t583/plan-audit.md` | FAIL | 0.79 | D1~D6 |
| 2 | `.moai/reports/t583/plan-audit-iter2.md` | FAIL | 0.86 | D9·D10 |
| 3 (최종) | `.moai/reports/t583/plan-audit-iter3.md` | FAIL | 0.87 | D17 |

판정·점수는 각 보고서 머리(2~4행)를 레인이 직접 읽어 옮겼다. 세 회차 모두 must-pass 는 통과했다(MP-4 해당 없음). 3회차는 Tier L 상한이라 재감사하지 않는다.

### 8.2 부채 — D17

- **내용**: AC-IQW-005 의 대상 목록은 "카드가 추가한 파일" 단위다. REQ-IQW-012 의 코드 의무는 "이 SPEC 이 새로 쓰거나 본문을 다시 쓴 테스트" 단위라 AC 가 요구보다 좁다(`plan-audit-iter3.md` L37).
- **통과할 수 있는 변형(감사자 판독, 실행 확인 없음)**: 기존 파일 안에 새 init 실행 테스트를 추가, 기존 테스트 본문을 다시 씀, `git add` 만 하고 커밋하지 않은 새 파일.
- **해악 범위**: 실제 홈 누출은 AC-IQW-015 슬롯 지문이 따로 덮는다. 뚫려도 코드 점검표를 갖추지 않은 테스트가 판정을 통과하는 데서 그친다.

### 8.3 부채를 받치는 조건 (run 단계 의무)

1. run 위임 프롬프트에 제약 "새 init 실행 테스트는 새 파일에만 추가한다"를 넣는다.
2. M6 마감 때, 기존 `internal/cli` 테스트 파일에 추가된 `^+func Test` 줄 수가 0 임을 명령과 출력 그대로 progress.md §E.2 에 기록한다.

### 8.4 선택 결함 D18~D20

run 단계에서 반영 여부를 이 절에 기록한다. SPEC 문서는 plan 단계에서 더 고치지 않는다.

| 결함 | 내용 | run 중 처분 |
|---|---|---|
| D18 | AC-IQW-016 L436 이 "AC-005 는 세 파일만 본다"는 v0.1.3 이전 전제를 적음 | 미정 |
| D19 | §5 완료 정의(L515)·plan.md M6 에 뮤턴트 E·F 기록 누락 | 미정 |
| D20 | REQ-IQW-012 문장과 §4.2 머리말의 게이트 테스트 범위 문구 차이 | 미정 |

### 8.5 미검증과 다음 단계

- **Gap**: 세 회차의 `moai spec lint` 종료 코드 0 은 설치본 `ed71054d3-dirty` 가 냈다. 이 빌드는 워크트리 HEAD `120436f58` 과 양방향 모두 조상 관계가 아니다(progress.md §E.1 기록). lint 판정은 이 트리로 만든 빌드에 귀속되지 않는다.
- **다음 단계**: Implementation Kickoff Approval 은 운영자에게 따로 묻는 중이다. 답이 오기 전에는 run 단계에 들어가지 않는다. 승인 뒤 develop 흡수는 로컬 develop tip(리드 전달 시점 `81c1d58f9`)을 대상으로 하고, 흡수 전에 tip 을 다시 잰다.

### 8.6 조건 2 의 판정 명령 후보 (run 위임문에 실을 것)

progress.md 는 조건 2 를 문장으로만 적었고 acceptance.md 에는 명령이 없다. 작업자마다 세는 방식이 달라지지 않도록 아래 명령을 run 위임문에 그대로 싣는다. 흡수 뒤, develop 병합 전에만 유효하다(범위 판정은 흡수한 로컬 develop 과의 merge-base 기준, gitflow-lane-protocol §8).

```bash
# 대조군: 카드가 수정한 기존 internal/cli 최상위 테스트 파일 수 — 1 이상이어야 판정이 성립(제거된 필드 참조를 지우므로)
git diff --name-only --diff-filter=M develop...HEAD -- ':(glob)internal/cli/*_test.go' > .moai/state/verify/t583/debt-d17-modified.txt; echo "exit=$?"
wc -l < .moai/state/verify/t583/debt-d17-modified.txt
# 판정: 커밋분 + 미커밋분에서 기존 파일에 추가된 테스트 함수 줄
git diff --diff-filter=M develop...HEAD -- ':(glob)internal/cli/*_test.go' > .moai/state/verify/t583/debt-d17-committed.diff; echo "exit=$?"
git diff --diff-filter=M HEAD -- ':(glob)internal/cli/*_test.go' > .moai/state/verify/t583/debt-d17-uncommitted.diff; echo "exit=$?"
command grep -c '^+func Test' .moai/state/verify/t583/debt-d17-committed.diff .moai/state/verify/t583/debt-d17-uncommitted.diff
```

기대: 대조군 1 이상, 두 diff 파일 모두 `^+func Test` 계수 0. 대조군이 0 이면 "측정 불가"로 보고한다.

plan 시점 관측(2026-09-11, HEAD `120436f58`): 대조군 명령의 파이프 형태 `git diff --name-only --diff-filter=M develop...HEAD -- ':(glob)internal/cli/*_test.go' | wc -l` 출력 `0` — 카드 커밋이 아직 없으니 옳은 이유로 측정 불가 상태다. 판정식이 실제로 빨개지는지(기존 파일에 `func Test` 를 더한 입력)는 실행하지 않았다(Gap).

## 9. run 착수 전 흡수와 인용 좌표 재측정 (2026-09-11)

- **Kickoff**: 운영자 승인(리드 전달), 대상 SPEC v0.1.4, D17 PASS-with-debt.
- **plan 산출물 커밋**: `f92c04952` (16 파일, 경로 명시 스테이징, 트레일러 `Authored-By-Agent: manager-spec` 파싱 확인).
- **흡수 전 신선도**: `git fetch origin develop -q` 뒤 `git rev-list --count --left-right origin/develop...develop` → `0	0`. `git rev-parse develop` → `4c99d973e92a8c74ab16136ce9e2c49ec4e2636c`(리드 전달값과 일치).
- **흡수**: `git merge --no-ff develop` → 병합 `2723447be`, `git rev-parse HEAD^2` → `4c99d973e…`, 이후 `git status --porcelain` 무출력.
- **SPEC 대상 경로 변화**: `git diff --stat 120436f58 HEAD -- internal/cli/wizard internal/cli/init.go internal/cli/update_wizard.go internal/cli/update.go internal/core/project internal/shell internal/cli/init_workflow_flags.go internal/template/templates/.moai/config/sections/workflow.yaml` → `update.go | 33`, `update_wizard.go | 17` 두 파일뿐.
- **인용 경로 교집합**: SPEC 산출물이 인용한 코드 경로 36개(`grep -rhoE '(internal|pkg|cmd)/…' | sort -u`)와 흡수로 바뀐 코드 파일 목록(`git diff --name-only 120436f58 HEAD -- internal pkg cmd`)의 교집합은 위 두 파일뿐이다. 나머지 인용 좌표는 해당 파일이 바이트 불변이라 그대로 유효하다.

| 인용 | 흡수 전(`120436f58`) | 흡수 뒤(`2723447be`) | 판정 |
|---|---|---|---|
| `internal/cli/update.go:824` (spec.md:204, design.md:101, research.md:109, spec-compact.md:46) | `configurator := shell.NewEnvConfigurator(nil)` | 같은 줄이 **797행** (`grep -n 'NewEnvConfigurator(' internal/cli/update.go` → `797:`) | 줄 번호만 이동. t587 이 legacy skill archive 블록 두 곳(483·548 부근)을 걷어냄 |
| 운영 코드의 셸 설정 기록 호출 "두 곳" | 두 곳 | `git grep -n 'NewEnvConfigurator(' -- internal ':!*_test.go'` → `update.go:797`, `initializer.go:678` (+ 정의 `env.go:32`, 내부 `:197`) | 주장 유지 |
| `update_wizard.go:64` | `wizard.ReconfigureQuestions(cwd)` | 같음 | 유지 |
| `update_wizard.go:133` | `func applyWizardConfig(…)` | 같음 | 유지 |
| `update_wizard.go:307-310` | `ModelPolicy` 적용 | 같음 | 유지 |
| `update_wizard.go:133-373` (verdict.md:161) | 함수 끝 `373` | 함수 끝 **`382`** | 끝 줄 이동(t587 +13/−4) |

SPEC 문서는 고치지 않는다(리드 지시). `update.go` 경로는 이 SPEC 범위 밖이라 797행 이동이 요구·AC 판정에 영향을 주지 않는다. run 위임문에는 흡수 뒤 좌표를 싣는다.

---

## 10. run·sync 단계 기록 (2026-09-12)

plan 단계 판정서(§1~§9)에 run 과 sync 에서 새로 관측된 것을 잇는다. 수치는 모두 `.moai/specs/SPEC-INIT-QUIET-WIZARD-001/progress.md` §E.2·§E.3 에 기록된 실측이며, 이 절은 인용한다. 재측정하지 않았다.

### 10.1 커버리지 — 세 줄

1. **실측**: `internal/cli` 커버리지는 카드 트리에서 **82.4%**(SLOT-29, `go test ./internal/cli -cover -count=1 -timeout 1500s`, tree HEAD `4079087ab`, `m6-cover-cli.txt`), 카드 base 트리에서 **82.3%**(SLOT-30, base `ee99507fbe3b4a22c6a0a74815723d222dfdc04d`, `m6-cover-cli-base.txt`) 였다. 차이는 **+0.1pp** 다.
2. **대조군의 한계 — 유효숫자를 주장하지 않는다**: base 트리는 `git archive | tar -x` 로 추출해 `.git` 이 없다. 저장소 이력을 읽는 테스트가 환경 때문에 실패하므로 base 의 실패 6건(`TestBuildIdentity_VersionDerivationUnchanged`, `TestBuildIdentity_IsMonotoneAcrossAnAncestorRelation`, `TestPreCommitLegacyNoRecord`, `TestHomeStateValidationCommandWrappersAndHelperFailures`, `TestTodoHistoryNeverPrompts`, `TestVersionStampRegistry`)은 base 트리의 결함이 아니라 추출 방식의 산물이다. **두 실패 집합은 직접 비교 대상이 아니고**, 커버리지 수치도 그 6건이 실패냐 통과냐에 따라 조금 움직인다. 그래서 `+0.1pp` 에 유효숫자를 주장하지 않는다 — 이 대조군이 세우는 것은 "카드가 커버리지를 **낮추지 않았다**" 까지다.
3. **85% 목표는 양쪽 트리 모두에서 미달이다**(`.moai/config/sections/quality.yaml` `test_coverage_target: 85`). 카드가 만든 상태가 아니라 패키지 수준의 선재 상태이며, 끌어올리는 일은 이 카드 범위 밖이다 — 후속 후보다(리드 판정, 2026-09-12).

### 10.2 뮤턴트 F 의 한계 — 판별식이 호출 모양을 본다

AC-IQW-005 스윕의 판별식은 `runInit[A-Za-z]*\(|prepareSafeInitHome\(` 로, **호출 모양의 텍스트**를 찾는다. 그래서 호출 없이 쓴 우회는 스윕을 빠져나가며 겉보기에 생존자로 읽힌다.

이 카드에서 실제로 관측했다(progress.md §E.2 뮤턴트 F 항목). 첫 시도는 `_ = runInit` 을 괄호 없이 담아 판별식에 걸리지 않았고, 스윕 결과는 targets 3 · 금지 토큰 교집합 0 — **깨끗한 트리와 완전히 같은 출력**이었다. 고친 뮤턴트(실제 `runInit(nil, nil)` 호출 포함)는 targets `3 → 4`, 금지 토큰 교집합 `0 → 1` 로 제대로 빨개졌다.

읽는 법: 이 스윕의 0 은 "init 을 실행하는 파일이 없다"가 아니라 **"init 을 이 이름으로 호출하는 파일이 없다"** 이다. 다른 철자로 init 을 실행하는 파일은 보이지 않는다.

### 10.3 AC-IQW-016 문구 정정 — 왜 필요한가

acceptance.md AC-IQW-016 은 뮤턴트 C1·C2 를 "원복 대입을 **삭제**한다"로 적었다. 문자 그대로 적용하면 Go 가 파일을 거부한다 — `declared and not used: origSeam`, `[build failed]`, 테스트는 한 줄도 실행되지 않는다(SLOT-12). 이것은 **실행 RED 가 아니라 도구 실패**이고, 도구 실패는 가드가 살아 있다는 증거가 되지 못한다. 둘은 관측적으로 구분돼야 한다.

그래서 관측은 컴파일되는 형태(`_ = origSeam`)로 다시 잡았다 — SLOT-12R·SLOT-13 이며, 둘 다 round 2 에서 기대한 실행 RED 를 냈다. SLOT-12 와 SLOT-12R 의 차이가 이 정정의 근거다.

**정정은 이미 반영됐다 — 커밋 `1c16e4227`.** manager-spec 재위임으로 acceptance.md:451-453 을 고쳤다: C1·C2 를 "줄 삭제"에서 "`_ = origSeam` 치환"으로 바꾸고, 그렇게 적는 이유(그냥 지우면 잡아 둔 변수가 쓰이지 않아 Go 가 컴파일을 거부하고 도구 실패가 된다)를 SLOT-12 vs SLOT-12R 근거와 함께 한 줄 추가했다. 기존 기대치(대입 줄 수 1 감소, C1 판정은 실행 RED, C2 는 텍스트+실행 RED, 뮤턴트 미커밋)는 전부 그대로다.

**[정정 기록]** 이 절의 종전 문장은 "정정은 아직 반영되지 않았다. 문구 반영은 manager-spec 재위임으로 남긴다" 였다. 이는 거짓이었다 — sync 를 수행한 manager-docs 가 acceptance.md 본문을 읽지 않고 progress.md §E.2 에 남아 있던 **정정 이전에 작성된 sync 의무 문장**을 근거로 판단한 결과다. 레인이 sync 커밋 직후 `sed -n '451,453p' acceptance.md` 와 `git log -1 -- acceptance.md`(→ `1c16e4227`)로 실물을 확인해 이 문장과 §10.6 (e) 를 정정했다. 교훈: 부채 해소 여부는 그 부채를 기록한 문서가 아니라 **부채의 대상 파일**에서 확인한다.

### 10.4 `init_workflow_wiring_test.go` — 주석만 바뀐 편집

plan.md:179 는 `TestRunInit_WorkflowToggleFlagsAbsentByteIdentical`(`init_workflow_wiring_test.go:83`)을 **변경 금지**로 적었다(AC-IQW-010). 이 카드는 그 함수의 **앞 주석**을 고쳤다 — 본문은 건드리지 않았다.

근거는 AC-IQW-010 의 본문 추출 비교다: `sed -n '/^func TestRunInit_WorkflowToggleFlagsAbsentByteIdentical/,/^}/p'` 로 base(`120436f58`)와 head 에서 각각 뽑아 `diff` → `byte-identical-diff-exit=0`, base 추출 12줄(progress.md §E.2 AC-IQW-010 본문 보존 절). 즉 함수 본문은 바이트 동일하고, 바뀐 것은 추출 범위 **밖**의 주석이다.

리드가 이 편집을 받아들였다(2026-09-12). 같은 파일의 다른 편집(`TestRunInit_WorktreeAutoCreateFlagBeatsWizard` 의 `wizard.WizardResult{WorktreeAutoCreate: false}` → `{}`)은 plan.md:178 이 명시한 "제거된 필드 참조 제거"라 변경 금지 대상이 아니다.

### 10.5 문서 영향 — 이 커밋에서 고치지 않았다

이 카드는 사용자 문서를 거짓으로 만든다. 고치지 않은 채 넘기므로 무엇이 왜 그런지 적는다.

- **README 4 로케일** — `README.md:288`, `README.ko.md:288`, `README.zh.md:288`, `README.ja.md:288` 이 대화형 위저드가 "모델 정책을 고른다"고 적는다. `model_policy` 는 이제 init 에서 묻지 않는다(`DefaultQuestions` 에 남아 reconfigure 경로 전용).
- **docs-site 4 로케일** — `docs-site/content/{en,ko,ja,zh}/getting-started/init-wizard.md`(각 235줄)이 고정 3-페이지 흐름과 페이지별 질문표를 서술한다. 4 로케일 모두 `3-페이지|3-page|Page 3|LSP` 계열 적중 4건.

고치지 않은 이유 셋:

1. **이 페이지는 이 카드 이전에 이미 낡아 있다.** Page 3 으로 적힌 질문들(LSP·품질 게이트·디자인)은 2026-08-03 에 이미 질문에서 빠졌다(`internal/cli/wizard/types.go` 주석). 지금 고치면 이 카드가 만들지 않은 드리프트까지 같은 diff 에 들어가 귀속이 불가능해진다.
2. **4 로케일이 구조적으로 평행하지 않다.** en 은 `Page 1/2/3` 골격, ko 는 `1단계/2단계` 골격에 산문 해설이 붙은 다른 편집 체제다. 미러 번역이 아니라 로케일마다 새로 쓰는 저작 작업이다.
3. **docs-site 의 Vercel 바인딩이 미검증이다.** `develop` 에 docs-site 변경이 들어갈 때 프리뷰/프로덕션 배포가 어떻게 반응하는지 확인되지 않았고, 저장소 규율은 docs-site 를 만지는 카드가 이 점을 별도로 확인하도록 요구한다(CLAUDE.local.md §4.1). 이 레인은 그 확인을 수행할 수 없다.

**처분은 리드 몫이다.** 위 8개 파일을 범위로 하는 별도 문서 카드를 제안한다. CHANGELOG 항목에도 미갱신 사실을 적었다.

### 10.6 잔여 위험

- **(a) 뮤턴트 C 생존 — 판별하는 테스트가 없다.** `init.go` 에서 `case agentWiringBoth: mcpDeclined = false` 분기를 지워도 실패가 나지 않았다(SLOT-27 4 PASS, 교정 선택자 SLOT-27b 5 PASS). 비대화형 `--llm both` 강제 분기를 판별하는 테스트가 없다. 리드 지시에 따라 이 카드는 그 테스트를 쓰지 않았다 — 후속 후보다.
- **(b) 대화형 경로 판별력 상실.** 대화형 MCP 기본값이 true 가 되면서 `internal/cli/init_agent_wizard_test.go:139`·`:159`("both 가 거절을 이긴다")는 대화형 경로에서 더 이상 판별하지 않는다 — 거절을 표현할 수단이 없어졌다. 두 단언은 여전히 통과하지만 이름이 주장하는 것을 재지 않는다.
- **(c) 기지 레드 4건, 병합 트리 재측정 대기.** `TestHomeStateChangedSurfaceCoverageConsumesFreshProfile`, `TestHomeStateChangedSurfaceCoverageRunsBoundedFocusedSuite`, `TestChangedProductionFilesDerivesCurrentHeadDiffAndPlatformDisposition`, `TestAuditLagUsesBinlagSeam`. base 트리에서도 같이 실패하므로 카드 귀속이 아니고, develop 의 `5b7927b15`(t600)·`92494400f`(t606)가 이 계열을 고친다. **귀속됐을 뿐 해소되지는 않았다** — 통합 창의 병합 트리 재측정이 판정한다.
- **(d) 미관측 Gap 은 통과가 아니다.** progress.md §E.3 Gaps 1~10 은 관측되지 않은 항목이며, 특히 AC-IQW-012·AC-IQW-013 은 이 카드가 수행한 **제거**를 단언하는 AC다. 그쪽 회귀는 실제로 돌린 어떤 검사에도 잡히지 않는다.
- ~~**(e) AC-IQW-016 문구 부채 미해소(§10.3).**~~ **해소됨** — 커밋 `1c16e4227` 에서 acceptance.md:451-453 이 `_ = origSeam` 치환 형태로 고쳐졌다. 종전 이 항목의 서술은 대상 파일을 읽지 않고 progress.md 의 옛 의무 문장을 근거로 쓴 것이며, §10.3 의 [정정 기록] 에 경위를 남겼다.

### 10.7 이 절이 재지 않은 것 (Gaps)

- 커버리지·슬롯·뮤턴트 수치를 **재측정하지 않았다**. 전부 progress.md §E.2·§E.3 인용이며, 이 sync 단계에서 `go test` 를 돌린 적이 없다(레인에 슬롯이 없다).
- 문서 영향은 **정적 grep 으로만** 판정했다(README 4건, docs-site 4건). 렌더된 페이지나 hugo 빌드로 확인하지 않았다.
- SPEC 본문(spec.md §A~§H, plan.md, acceptance.md, design.md, research.md)은 이 커밋에서 **고치지 않았다**. spec.md 프런트매터의 `status`·`updated` 두 필드만 3단계 마감으로 갱신했다.
