# SPEC-INIT-QUIET-WIZARD-001 — 설계

이 문서는 run 단계가 따를 설계 방향을 제안한다. 함수·파일·테스트 이름은 제안이며, run 단계가 기존 코드 관례에 맞춰 확정한다. 이름이 바뀌면 acceptance.md 의 선택자도 함께 고친다. 인용 행번호는 워크트리 HEAD `120436f58` 기준이다.

## §1 목표와 제약

- init 은 4문항만 묻고, 나머지는 기본값에 맡긴다.
- reconfigure(`moai update -c`)의 질문은 한 문항도 바뀌지 않는다(D1).
- 대화형 사용자가 오늘 받는 MCP 프로비저닝 동작을 보존한다(D2).
- 위저드 렌더링·그룹 구성·스타일은 건드리지 않는다(t586).
- "미설정 = 기본값" 을 실행으로 증명하되, 실제 홈에는 아무것도 쓰지 않는다.
- 셸 설정 단계의 도달은 실행으로 관측한다. 소스 문자열 검사는 증거로 쓰지 않는다(리드 결정).

## §2 질문 생성자 분할

### §2.1 현재 구조

```
DefaultQuestions (5)  ── conversation_language, user_name, project_name, model_policy, report_format
GitQuestions (7)
Page3Questions (13)
ReconfigureQuestions = DefaultQuestions + GitQuestions (report_format 뒤에 끼움)   ← update_wizard.go:64
InitQuestions        = DefaultQuestions + Page3Questions                          ← wizard.go:39 RunWithDefaults
```

`DefaultQuestions` 가 init 과 reconfigure 의 공통 기반이다(`questions.go:269`, `:297`). 여기서 세 문항을 빼면 reconfigure 에서도 빠진다.

### §2.2 변경 뒤 구조

```
DefaultQuestions (5)   변경 없음
GitQuestions (7)       변경 없음
ReconfigureQuestions   변경 없음 (12)
Page3Questions (2)     agent_wiring, autonomy_tier  — 리터럴 내용·그룹 라벨 그대로
InitQuestions  (4)     pick(DefaultQuestions, "conversation_language", "user_name") + Page3Questions
```

- `pick` 은 ID 목록 순서대로 질문을 골라 새 슬라이스로 돌려주는 패키지 내부 도우미다. 없는 ID 를 만나면 조용히 건너뛰지 말고 테스트가 잡을 수 있게 한다. AC-IQW-001 이 4개 전부와 순서를 고정하므로, 도우미가 하나를 놓치면 그 테스트가 실패한다.
- 질문 리터럴을 복제하지 않는다. `conversation_language`·`user_name` 의 정의, 프로필 미리 채움(`prefillIdentityDefaults`), 번역이 한 곳에만 있다.
- `RunWithDefaults`(`wizard.go:35-59`)는 계속 `InitQuestions` 를 부른다. 이 함수는 손대지 않는다.

### §2.3 기각한 대안

| 대안 | 기각 이유 |
|---|---|
| `DefaultQuestions` 에서 세 문항을 직접 삭제 | reconfigure 에서도 사라짐 — D1 위반 |
| init 전용 생성자에 네 리터럴을 복사 | 정의·미리 채움·번역 대상이 두 벌이 되어 갈라짐 |
| 제거 문항에 `Condition: false` 를 달아 숨김 | 묻지 않는 질문이 결과 구조체와 매핑에 남음. "숨겨졌지만 연결된" 상태가 바로 F1 을 만든 모양이다. 스테퍼 분모 계산도 조건 필터에 계속 의존 |
| reconfigure 쪽을 새 생성자로 분리하고 `DefaultQuestions` 를 4문항 기반으로 축소 | reconfigure 진입 파일(`update_wizard.go`)과 그 고정 테스트까지 바뀜 — "reconfigure 불변" 을 증명할 비교 기준이 사라짐 |

## §3 대화형 MCP 기본값

### §3.1 현재 흐름

```
위저드 mcp_provision 확인(기본 true) → saveBoolAnswer → WizardResult.MCPProvision
  → applyWizardPage3ToOpts: opts.MCPProvision = result.MCPProvision (init.go:337)
  → mcpDeclined := !opts.MCPProvision; codex → true, both → false (init.go:1004-1011)
  → provisionMCPEntryUnlessDeclined(...)
비대화형: 위 매핑이 돌지 않아 opts.MCPProvision = false → 보장 호출 생략
```

### §3.2 변경 뒤 흐름

```
대화형 블록(init.go:694~) 안: opts.MCPProvision = true   ← 코드 기본값
  → mcpDeclined 규칙 그대로 (codex 건너뜀, both 실행)
비대화형: 변경 없음 (0값 false → 생략)
```

- 기본값을 위저드 결과가 아니라 `internal/cli/init.go` 대화형 블록에 둔다. 테스트가 `runWizardFn` 으로 결과를 주입해도 운영 경로와 같은 값이 나온다.
- `WizardResult.MCPProvision` 은 채우는 질문이 없어지므로 지운다. `InitOptions.MCPProvision` 은 남긴다. `init_mcp_provision_test.go:105` 의 기존 소스 도달성 가드가 `opts.MCPProvision` 문자열을 요구하고, 비대화형 비대칭을 표현하는 자리이기도 하다. 그 기존 가드는 이 SPEC 이 새로 만드는 증거가 아니며, 이 SPEC 의 새 도달 증거는 §4 의 실행 관측이다.
- `init.go:1004` 바로 위 주석 블록(`:999` "Accepted cost (plan.md §B Decision B1)", `:1003` `@MX:SPEC`)은 "codex 를 골라도 `mcp_provision` 을 묻는다(결정 B1)" 는 서술을 담고 있다. 이 서술을 "대화형은 기본 프로비저닝, 하네스 규칙이 덮어씀, 비대화형은 생략" 으로 고친다.

### §3.3 기각한 대안

| 대안 | 기각 이유 |
|---|---|
| `RunWithDefaults` 시드에 `MCPProvision: true` 추가 | 기본값이 위저드 결과에 기대게 되어, 주입 테스트가 운영 동작을 재지 못함. `wizard.go` 시드 구간은 t586 과 겹침 |
| 보장 호출을 대화형·비대화형 모두에서 무조건 실행 | 비대화형 고정 테스트(`init_agent_wizard_test.go:172-205`) 위반 — D2 위반 |
| 보장 호출 자체를 삭제(템플릿 `.mcp.json` 에만 의존) | `--force` 재초기화나 사용자 `.mcp.json` 에서 결과가 달라질 수 있는데 측정하지 않았음. `both` 강제 규칙도 깨짐 |

## §4 셸 설정 단계 시접

### §4.1 누출 경로와 호출 구조

```
runInit → project.InitOptions{...}                       (init.go:592-617, SkipShellConfig 미설정)
  → project.NewInitializer / NewPhaseExecutor            (init.go:832-833, 인라인 생성)
  → executor.Execute(ctx, opts)                          (init.go:867)
    → initializer.Init → Step 6: !opts.SkipShellConfig   (initializer.go:333-347)
      → i.configureShellEnv()                            (initializer.go:677)
        → shell.NewEnvConfigurator(i.logger).Configure(...)
          → selectConfigFile: os.Getenv("HOME")          (shell/detect.go:128-145)
          → os.OpenFile(..., O_APPEND|O_CREATE|O_WRONLY) (shell/config.go:122, :199)
          → 줄이 이미 있으면 Skipped                       (shell/config.go:93, :167)
```

- `userHomeDirFn`·`profile.BaseDirOverride`·`MOAI_HOME` 어느 것도 이 경로를 막지 못한다. `t.Setenv("HOME")` 은 금지돼 있다.
- `runInit` 은 initializer 와 executor 를 인라인으로 만들어 `opts` 를 곧바로 넘기므로, `internal/cli` 테스트가 `InitOptions` 를 중간에서 볼 틈이 없다.
- 운영 코드에서 셸 설정 기록을 부르는 곳은 이 경로와 `moai update` 의 `internal/cli/update.go:824` 두 곳이다. 이 SPEC 은 init 경로만 다룬다.

### §4.2 제안 모양 — `internal/core/project` 의 내보낸 함수 변수 하나

```go
// internal/core/project/initializer.go (제안)

// ConfigureShellEnvFn performs the Step 6 shell-config write. Tests swap it for a
// spy so they can observe that Step 6 was reached without writing real rc files.
var ConfigureShellEnvFn = defaultConfigureShellEnv

func defaultConfigureShellEnv(logger *slog.Logger) (*shell.ConfigResult, error) {
	return shell.NewEnvConfigurator(logger).Configure(shell.ConfigOptions{
		AddClaudeWarningDisable: true,
		AddLocalBinPath:         true,
		AddGoBinPath:            true,
		PreferLoginShell:        true,
	})
}

func (i *projectInitializer) configureShellEnv() (*shell.ConfigResult, error) {
	return ConfigureShellEnvFn(i.logger)
}
```

- **운영 동작 불변.** 기본값 `defaultConfigureShellEnv` 는 현재 `configureShellEnv` 본문(`:678-685`)과 같다. Step 6 게이트 `!opts.SkipShellConfig` 도 그대로 둔다.
- **위치와 가시성.** 이 저장소의 시접 관례(`userHomeDirFn`, `runWizardFn`, `isInteractiveStdin`)를 따라 함수 변수로 두되, 기록 지점이 `internal/core/project` 안에 있으므로 거기에 둔다. `internal/cli` 테스트가 바꿔 끼워야 하므로 내보낸 이름이 필요하다. `internal/` 아래라 이 모듈 밖에서는 import 할 수 없어 사용자 표면이 생기지 않는다. `internal/core/project` 에는 아직 이런 함수 변수가 없다(`git grep` 0건) — 이 변수가 첫 사례다.
- **스파이.** 테스트용 스파이는 호출 횟수를 세고 `&shell.ConfigResult{Skipped: true}, nil` 을 돌려준다. 실제 기록 함수를 대신하므로 스파이가 끼워진 동안에는 셸 설정 파일에 쓸 수 없다.
- **시접을 바꾸는 테스트의 의무 (리드 조건, 2026-09-11).** 시접을 바꾸는 테스트·헬퍼는 모두 (1) `t.Parallel` 을 쓰지 않고, (2) `t.Cleanup` 으로 원래 값을 되돌리며, (3) 시접을 바꾸기 전에 `t.Setenv` 를 한 번 이상 부른다. (3) 은 Go testing 이 `t.Setenv` 와 `t.Parallel` 의 병용을 panic 으로 막는 성질(go1.26.8 `src/testing/testing.go:1752`, `:1765-1766`, `:1830-1838`)을 빌려, 병렬 금지를 텍스트가 아닌 실행으로도 드러내기 위한 것이다. `internal/cli` 헬퍼는 `MOAI_HOME` 우회(§5 3번)로, `internal/core/project` 게이트 테스트는 `MOAI_HOME` 을 `t.TempDir()` 아래로 돌리는 호출로 충족한다. 이 리드 조건은 acceptance.md AC-IQW-016 으로 판정한다.
- **기본값 직접 읽기(보조 관측).** 함수 값은 `==` 로 비교할 수 없으므로, 같은 패키지 테스트에서 `reflect.ValueOf(ConfigureShellEnvFn).Pointer() == reflect.ValueOf(defaultConfigureShellEnv).Pointer()` 로 확인한다.

### §4.3 시접 하나로 두 목적을 맡는 이유

두 목적이 있다: (1) 테스트가 실제 셸 설정 파일에 쓰지 않게 막기, (2) `runInit` 이 셸 설정 단계까지 실제로 도달하는지 실행으로 관측하기.

- **(1) 억제.** 스파이가 기록 함수 자체를 대신하므로 쓰기가 일어날 수 없다. 이전 설계의 `internal/cli` 억제 변수(`runInit` 이 `SkipShellConfig` 를 켜게 하는 방식)보다 약하지 않다 — 그 방식도 "테스트가 시접을 켜야 안전" 이라는 같은 전제에 기댔다.
- **(2) 관측.** 주 관측 테스트는 스파이를 끼운 채 실제 `runInit` 을 돌려 호출이 정확히 1회임을 본다. `runInit` 은 `SkipShellConfig` 를 켜지 않으므로, 1회는 기본 게이트가 Step 6 까지 실제로 전달됐다는 뜻이다. 게이트가 켜진 경우의 0회는 `runInit` 에 그 경로가 없으므로 `internal/core/project` 게이트 테스트가 실제 `Init` 을 `SkipShellConfig=true` 로 돌려 본다.
- **두 번째 시접을 두지 않는 이유.** `internal/cli` 억제 변수를 따로 두면 `runInit` 의 운영 코드에 배선이 하나 늘고, 그 배선이 실제로 전달되는지를 또 관측해야 한다. 스파이 시접 하나로 두 목적이 모두 충족되므로 더할 이유가 없다.
- **증명하지 못하는 것.** 시접을 우회해 `configureShellEnv` 가 `shell.NewEnvConfigurator` 를 직접 부르는 회귀는 스파이 호출 0회로 RED 가 되지만, 그 실행 중에 실제 셸 설정 파일에 쓸 수 있다. 이 우회를 뮤턴트로 만들어 돌리지 않으며(plan.md §D), 구현 회귀는 실행 슬롯 지문(spec.md §4.3)과 멈춤 규칙이 받친다.

### §4.4 배선 제거 뮤턴트

주 관측 테스트가 실제로 빨개질 수 있음을 기록한다(acceptance.md AC-IQW-004).

| 뮤턴트 | 바꾸는 곳 | 기대 |
|---|---|---|
| A — 게이트 반전 | `initializer.go:333` `if !opts.SkipShellConfig` → `if opts.SkipShellConfig` | `runInit` 은 게이트를 켜지 않으므로 Step 6 건너뜀 → 스파이 0회 → 주 관측 RED. 게이트 테스트도 RED |
| B — 호출 삭제 | Step 6 블록의 시접 호출 제거 | 스파이 0회 → 주 관측 RED |
| (금지) 시접 우회 | `configureShellEnv` 가 기록 함수를 직접 호출 | 실제 셸 설정 파일에 쓰므로 만들지 않음 |

### §4.5 기각한 대안

| 대안 | 기각 이유 |
|---|---|
| 소스 문자열 도달성 가드(`init.go` 에 `SkipShellConfig:` 문자열이 있는지 검사) | 텍스트 패턴 추론이라 배선을 우회하는 변형도 통과할 수 있음 — 리드 결정으로 기각 |
| `internal/cli` 억제 변수 + `InitOptions` 관측 시접(실행기 생성 함수 변수) | 운영 코드에 시접이 둘 생기고, 그래도 셸 설정 단계 자체의 도달은 보지 못함 |
| `InitOptions` 에 기록 함수 필드 추가 | 모든 호출자가 쓰는 구조체의 표면이 커지고, 필드를 채우는 `internal/cli` 쪽 시접이 또 필요 |
| `NewInitializer` 생성자 인자로 주입 | `runInit` 이 인라인으로 생성하므로 테스트가 닿으려면 생성자를 감싸는 시접이 또 필요 |
| "실제 셸 설정 파일이 바뀌지 않았다" 로 간접 관측 | `shell/config.go:93`, `:167` 의 건너뛰기 때문에 이미 설정된 머신에서는 시접과 무관하게 불변. 기본 게이트 전달을 실제 기록으로 보면 실제 `~/.zshenv` 에 씀 |
| `--skip-shell-config` CLI 플래그 | 사용자 표면이 생김(도움말·문서·템플릿). 필요한 곳은 테스트뿐 |
| 환경 변수(`MOAI_SKIP_SHELL_CONFIG` 등) | 운영에서도 도달 가능해지고 `envkeys.go` 상수·문서가 필요. 테스트 환경 변수가 부모 셸로 새면 사용자 init 이 조용히 셸 설정을 건너뜀 |
| 테스트에서 `t.Setenv("HOME")` | 리드 지시로 금지 |
| `internal/shell` 에 홈 시접 추가 | `os.Getenv("HOME")` 과 `os.UserHomeDir()` 두 해석을 모두 막아야 해서 변경이 더 크고, 도달 관측은 여전히 불가 |

## §5 홈 안전 헬퍼

재현 프로브(`.moai/reports/t583/repro/t583_repro_test.go:24-121`)의 패턴을 테스트 헬퍼로 옮긴다.

```
prepareSafeInitHome(t) (fakeHome string, spy *shellSeamSpy)
  1. 실제 홈 지문 채집: settings.json sha256, hooks/moai 존재, 셸 설정 6개(.zshenv·.zshrc·.zprofile·.profile·.bashrc·.bash_profile) mtime+sha256
  2. t.Cleanup: 같은 지문 재채집 → 하나라도 다르면 t.Errorf
  3. userHomeDirFn, profile.BaseDirOverride → t.TempDir() 아래, MOAI_HOME → t.Setenv 로 t.TempDir() 아래
  4. project.ConfigureShellEnvFn 가 테스트 패키지 초기화 때 잡아 둔 원래 값과 같은 함수인지
     reflect 함수 포인터로 단언 (다르면 t.Fatal — 앞선 테스트가 원복하지 않았다는 뜻)
     → 호출을 세는 스파이로 교체, t.Cleanup 으로 원복
  5. checkSeamHomes(realHome, seamPaths) error  ← 판정만 하는 순수 함수
     오류면 t.Fatal (runInit 호출 전)
```

- 5번을 오류를 돌려주는 함수로 분리해, "실제 홈 아래 경로면 거부한다" 는 음성 사례와 임시 디렉터리 수용을 단위 테스트로 잰다(AC-IQW-005).
- 헬퍼는 스파이를 돌려주어, 주 관측 테스트(AC-IQW-004)가 호출 횟수를 읽을 수 있게 한다.
- 지문의 부재 파일은 `absent` 로 기록해, "없던 파일이 생김" 도 변화로 잡는다.
- 헬퍼를 쓰는 테스트는 `t.Parallel` 을 쓰지 않는다. 3번의 `t.Setenv` 때문에 헬퍼를 부르는 테스트에 `t.Parallel` 을 더하면 Go testing 이 panic 하고, 4번의 진입 단언은 원복 누락을 같은 프로세스의 다음 실행에서 드러낸다. 두 성질은 acceptance.md AC-IQW-016 이 잰다.
- 기존 헬퍼 `runInitForAutonomyAtHomeCapturingOut`(`init_autonomy_wiring_test.go:40`)는 `t.Setenv("HOME")` 을 쓰므로 새 테스트에서 재사용하지 않는다. 리드 결정 (a)(2026-09-11)에 따라 제거된 필드 참조만 지우는 기존 테스트는 이 헬퍼를 그대로 쓰고, 시접 기반 헬퍼로의 이관은 범위 밖 후속 후보다(plan.md §N).
- 테스트 코드 안의 대조와 별개로, run 단계의 모든 `go test` 슬롯은 테스트 프로세스 밖에서 같은 8항목을 같은 명령으로 전후 채집한다(spec.md §4.3, acceptance.md AC-IQW-015). 코드 안 대조는 새 테스트만 덮고, 슬롯 지문은 기존 헬퍼 테스트까지 덮는다.

## §6 실행 테스트 설계

### §6.1 주입 결과의 모양

```go
&wizard.WizardResult{
    ConversationLang: "en", UserName: "tester", AgentWiring: "claude", AutonomyTier: "semi-auto",
    // RunWithDefaults 시드와 같은 값 (wizard.go:46-52)
    LSPEnabled: true, EnforceQuality: true, CoverageExemptionsEnabled: false,
    DesignEnabled: true, ClaudeDesignEnabled: true,
}
```

시드를 빼면 `applyWizardPage3ToOpts` 가 LSP·품질·디자인을 false 로 옮겨, 운영에 없는 상태를 재게 된다. 시드 일치는 기존 `TestSeedMirrorsProductionLSPSeed`(`init_flag_precedence_test.go:312`)와 함께 돌려 확인한다. 필드 이름은 `types.go` 에서 확인했다(`ConversationLang`:15, `UserName`:16, `LSPEnabled`:47 등).

### §6.2 관측기

- 입력: 프로젝트 경로. 출력: 불일치 목록(키, 기대, 실제).
- 파일 관측: `project.yaml`·`report.yaml`·`llm.yaml`·`feedback.yaml` 은 YAML 을 직접 읽는다(`project.yaml`·`report.yaml` 은 로더 소비자가 없는 섹션).
- 해석 관측: workflow 계열은 `config.NewLoader().Load(<project>/.moai)`(`internal/config/loader.go:31`)의 결과에서 읽는다. `todo.enabled` 는 디스크 키 부재와 해석값 true 를 둘 다 본다.
- `.mcp.json` 은 JSON 으로 읽어 `mcpServers.moai` 존재를 본다.

### §6.3 음성 대조군과 뮤턴트

- 테스트 안(AC-IQW-007a): 생성된 프로젝트를 임시 디렉터리로 복사해 두 키를 기본값이 아닌 값으로 고친 뒤 관측기를 돌려, 불일치가 정확히 그 두 키로 보고되는지 본다. 원본은 불일치 0건이어야 한다. 이것으로 "관측기가 원래 아무것도 못 잡는" 공허한 초록을 막는다.
- 코드 뮤턴트(AC-IQW-007b): 대화형 블록에 기본값이 아닌 쓰기를 넣으면 같은 테스트가 실제로 빨개지는지 run 단계가 기록한다. 기록기가 남는 키(`project.mode`, 워크트리 추적자)를 골랐다. 삭제되는 기록기(audit 등)는 뮤턴트로 되살릴 수 없기 때문이다.

### §6.4 두 실행 비교

대화형 4문항 실행과 플래그 없는 비대화형 실행을 **같은 디렉터리 이름**의 서로 다른 임시 루트에서 수행해, 섹션 파일 5개를 바이트 비교한다(REQ-IQW-015, AC-IQW-008). 이름이 같아야 `project.yaml` `name` 이 같다. `workflow.yaml` 비교는 `audit:` 키 자체가 아니라 그 아래 `model`·`gates` 하위 키의 삽입 여부로 갈린다 — 템플릿이 `workflow.yaml:85-91` 에서 `audit.codex`·`audit.glm` 핀을 싣고 배포하므로 두 쪽 모두 `audit:` 키를 가진다.

## §7 남는 4문항의 그룹

| 문항 | 그룹 라벨 | huh 그룹 |
|---|---|---|
| `conversation_language`, `user_name` | Basic | 1 |
| `agent_wiring` | Quality & Workflow | 2 |
| `autonomy_tier` | Autonomy | 3 |

같은 라벨이 연속한 무조건 문항끼리 묶이므로(`wizard.go:158` `buildFormGroups` 본문) huh 그룹은 3개가 된다. "Quality & Workflow" 에 하네스 질문 하나만 남아 라벨이 내용과 어긋나지만, 라벨 재구성은 렌더링 결정이라 t586 에 맡긴다. 이 SPEC 은 라벨 문자열을 바꾸지 않는다.

## §8 결정 요약

| 결정 | 선택 | 핵심 이유 |
|---|---|---|
| 생성자 분할 | ID 로 골라 조립 | 리터럴 단일화, reconfigure 불변 |
| MCP 기본값 위치 | init 대화형 블록 | 주입 테스트 = 운영 경로 |
| 셸 설정 단계 | `internal/core/project` 내보낸 함수 변수 하나 + 스파이 | 쓰기 억제와 도달 관측을 한 시접이 맡음, 사용자 표면 없음, 실행으로 증명 |
| 홈 안전 | 판정 함수 분리 헬퍼 + 스파이 반환 | 가드 자체를 음성 사례로 검증 |
| 증명 | 실행 + 관측기 대조군 + 코드·배선 뮤턴트 + 스윕 확인 | 목록·문자열 검사만으로는 F1 류와 빈 스윕을 못 잡음 |
| 그룹 라벨 | 현행 유지 | t586 경계 |
