# SPEC-PROFILE-MEMORY-001 — 코드 실측 원장

> Tier L 산출물. 기준 커밋 `c907db541` (main). 모든 항목은 파일 판독 또는 명령 실행으로 확인했다.
> **줄 번호는 기록하지 않는다** — 드리프트하기 때문이다. 앵커는 함수명과 관측 가능한 동작이다.

---

## §A 프로필 서브시스템 현황

### A.1 `internal/profile/profile.go`

| 심볼 | 확인된 동작 |
|------|-------------|
| `profilesDir` | `".moai/claude-profiles"` 상수 |
| `launchLedgerFile` | `"launch.yaml"` 상수 |
| `lastProfileKey` | `"last_profile"` 상수 — **원장의 유일한 MoAI 관리 키** |
| `BaseDirOverride` | 패키지 변수. 비어있지 않으면 `GetBaseDir()` 가 이를 반환 |
| `GetBaseDir()` | `BaseDirOverride` 또는 `~/.moai/claude-profiles`. **cwd 를 전혀 참조하지 않는다** → 결함 1의 직접 원인 |
| `GetCurrentName()` | `CLAUDE_CONFIG_DIR` 미설정 시 `ResolveLaunchProfile("")` 로 폴백, 결과 없으면 `"default"` |
| `isValidProfileName()` | 슬래시·역슬래시·선행 점·절대경로 거부. **디렉터리 존재는 검사하지 않는다** |
| `EnsureDir(name)` | `os.MkdirAll` + `os.Setenv("CLAUDE_CONFIG_DIR", ...)` **뿐**. 시드·경고 없음. 멱등 |
| `ResolveLaunchProfile(name)` | 명시 `-p` 우선 → `MOAI_NO_PROFILE_FALLBACK` 옵트아웃 → 원장 읽기 → `isValidProfileName` → `os.Stat` + `IsDir` (stale 가드) → 이름 반환. **읽기 측에만 존재 검증이 있다** |
| `RecordLastUsedProfile(name)` | `""`/`"default"` 거부 → `isValidProfileName` → read-modify-write(`map[string]any`) → `os.CreateTemp` + `Chmod(0o644)` + `os.Rename`. **디렉터리 존재 검사 없음** → 결함 2 |

**결함 2의 비대칭 구조**: 쓰기는 형태만, 읽기는 형태+존재를 검사한다. 그래서 유령 이름이 원장에 영구 잔류하면서도 해석 시 조용히 무시되어 사용자가 알 방법이 없다.

### A.2 결함 3 — 인증 상태 운반체 실측

| 관측 | 결과 |
|------|------|
| macOS Keychain 서비스 `Claude Code-credentials` | 존재. 서비스 스코프 전역이라 `CLAUDE_CONFIG_DIR` 변경으로 사라지지 않는다 |
| `~/.claude.json` | 존재(홈 루트, 약 200KB). `oauthAccount` / `hasCompletedOnboarding` / `userID` 보유 |
| `~/.claude/.claude.json` | **부재** |
| 명명 프로필 **4개** 의 `<profileDir>/.claude.json` | **보유 2개** (`mo.ai.kr`, `moai-adk` — 둘 다 `oauthAccount` 보유) / **부재 2개** (`copy-that`, `moai-cowork`) |
| 부재 2개의 의미 | 저장소의 절반이 **지금 이 순간 결함 3의 실사례**다. 다음 런치에서 로그인/온보딩 화면을 만난다 |
| `internal/web/profile_crud.go` `createProfileDir` | `GetProfileDir` + `os.MkdirAll` **뿐** — `.claude.json` 을 쓰지 않는다. 「콘솔에서 프로필 생성 → 저장(원장 기록) → bare `moai cc`」 가 정확히 이 상태를 만든다 |
| `internal/statusline/usage.go` 읽기 우선순위 주석 | macOS Keychain → `~/.claude/.credentials.json` → `~/.claude/credentials.json` |
| `internal/`·`cmd/`·`pkg/` 전역에서 자격증명 복사 로직 | **0건** |

**정정된 기전**: 종전 기록은 "`.credentials.json` 이 `~/.claude` 에만 있어서"로 적었으나 macOS에서는 불완전하다. 실제 승계되지 않는 것은 **`.claude.json`(계정 상태)** 이며, Keychain 토큰은 살아 있다. Linux/WSL2 에서는 파일이 실제 운반체이므로 `.credentials.json` 부재가 직접 영향을 준다. 이 플랫폼 비대칭이 자동 시드를 범위 밖에 두는 근거다.

---

## §B 런처 현황 — `internal/cli/launcher.go`

### B.1 `unifiedLaunchDefault` 현행 단계

| 단계 | 동작 | 이 SPEC과의 관계 |
|------|------|------------------|
| 1 | `resolveMode(modeOverride)` | 무관 |
| 2 | `resolved := profile.ResolveLaunchProfile(profileName)` | **root 를 못 얻은 상태** → project-scoped 해석 불가 (감사 D1의 근거) |
| 3 | `root, err := findProjectRootFn()`; 실패 시 즉시 `return` | root 생성 지점. RESOLVED-3의 대상 |
| 4 | `applyGLMMode(root, profileName)` / `applyCGMode(root, profileName)` / `applyCCMode(root)` | **해석된 `profileName` 을 소비** → 재배열 시 순서 보존 필요 |
| 5 | `profile.RecordLastUsedProfile(originalProfile)` | 디렉터리 생성 전 → 결함 2 가드와 충돌 |
| 6 | `launchClaude(profileName, extraArgs)` | 이후 코드 없음(POSIX exec) |

`ResolveLaunchProfile` 호출 지점은 이 2단계 **단 하나**다(`grep -n ResolveLaunchProfile internal/cli/`).

### B.2 `launchClaudeDefault` 조기 반환 경로

`EnsureDir` 이후에도 여러 조기 반환이 있다: `exec.LookPath("claude")` 실패, `--continue` 재개 실패(비-exit-1). 말미에서 `execOrSpawnClaude` 가 POSIX `syscall.Exec` 로 프로세스를 대체하므로 그 이후 코드는 실행되지 않는다.

→ 기록을 `launchClaudeDefault` 안으로 옮기는 안이 기각된 이유: `EnsureDir` 직후와 exec 직전 사이에 "어디에 넣어도 안전한 한 지점"이 존재하지 않는다.

### B.3 기존 시임 패턴

`unifiedLaunchFunc`, `launchClaudeFunc`, `findProjectRootFn`, `newDetectorFn`, `injectTmuxSessionEnvFn` — 모두 패키지 레벨 `var` 로 선언된 테스트 시임이다. 신규 `recordLastProfileFn` 은 이 확립된 패턴을 따른다.

### B.4 주입 `io.Writer` 경고 선례

`warnNoModelResolved(w io.Writer, profileName string)` 이 이미 존재한다. 고지 함수는 이 선례를 그대로 따라 테스트에서 버퍼로 단언 가능하게 만든다.

---

## §C 호출자 전수 조사

`grep -rn "RecordLastUsedProfile\|ResolveLaunchProfile\|GetCurrentName" --include="*.go" internal cmd pkg | grep -v _test.go` 결과(주석 행 제외):

| 파일 | 심볼 | 처리 |
|------|------|------|
| `internal/cli/launcher.go` | `ResolveLaunchProfile`, `RecordLastUsedProfile` | **변경** (M3) |
| `internal/web/app.go` | `recordLastProfile: profile.RecordLastUsedProfile` | **변경** — 클로저 배선 (M2) |
| `internal/web/handlers.go` | `a.recordLastProfile(selected)` | 무변경 — 시임 경유, 에러 무시(advisory) |
| `internal/cli/web.go` | `ProfileName: profile.GetCurrentName()` | **변경** — REQ-PM-024 (M2) |
| `internal/cli/profile.go` | `GetCurrentName()` | 무변경 — `moai profile current` 출력 보존(L-001) |
| `internal/cli/update.go` (2지점) | `GetCurrentName()` | 무변경 |
| `internal/cli/init.go` | `GetCurrentName()` | 무변경 |
| `internal/web/profile_crud.go` | `GetCurrentName()` | 무변경 — 삭제 가드 비교, 프로젝트 무관 |

**iter-1 정정 (감사 D6)**: 종전 plan.md 는 무변경 호출자를 4곳으로 적었으나 실제는 5곳이며, 누락된 `internal/cli/web.go` 가 바로 사용자 결정으로 **변경** 대상이 된 항목이다.

`internal/cli/web.go` `runWeb` 은 함수 앞부분에서 이미 `projectRoot, err := findProjectRootFn()` 를 수행하므로, project-scoped 변경은 인자 추가 없이 1줄 교체다.

---

## §D 테스트 인프라 현황

| 관측 | 결과 |
|------|------|
| `internal/cli/main_test.go` | `TestMain` + `sandboxProfileBaseDir()` + `TestProfileBaseDirIsSandboxed` 가드 존재 |
| `internal/profile/main_test.go` | **부재** (`ls internal/profile/`) |
| `internal/web` `func TestMain` | **0건** (`grep -rn "func TestMain" internal/web/`) |
| `internal/profile/profile_test.go` `TestGetBaseDir_Default` | `BaseDirOverride = ""` 설정 후 **복원 없음**. 패키지 내 유일한 미복원 사례 (다른 테스트는 모두 `defer`/`t.Cleanup` 사용) |
| `internal/web` 에서 `newApp` 직접 호출 테스트 | `board_test.go`(2), `agentfm_policy_test.go`, `agentfm_polish_test.go`, `m5_agentfm_test.go`, `schema_sections_test.go`, `projectnested_test.go`, `projectconfig_test.go`, `profile_crud_test.go`(3), `profile_traversal_test.go`, `handlers_test.go` |
| `internal/web` 시임 스텁 | `handlers_test.go` `newTestApp` 이 `recordLastProfile` 을 스텁하나, 위 직접 호출 경로는 커버하지 않는다 |

**두 개의 선행 하자**:
1. `internal/profile` 과 `internal/web` 모두 패키지 샌드박스가 없다. 이 SPEC이 두 패키지를 모두 수정하므로 둘 다 필요하다(REQ-PM-022).
2. `TestGetBaseDir_Default` 의 미복원은 패키지 `TestMain` 샌드박스를 그 이후 테스트에 대해 무력화한다. Go 의 사전순 파일 실행 때문에 `main_test.go` 안의 가드로는 이를 탐지할 수 없다 → 후행 가드 필요(감사 D12).

---

## §E 도구 환경 실측

| 항목 | 관측 |
|------|------|
| `grep --version` | `ugrep 7.5.0` — GNU/PCRE 확장(`\s` 등)을 관대하게 처리 |
| 영향 | AC 명령이 `\s` 를 쓰면 **이 머신에서만** 통과하고 stock BSD `/usr/bin/grep` 에서는 리터럴 `s` 로 해석되어 0매치 → 거짓 실패. 감사 D11의 근거이며 제약 C8로 코드화했다 |

`.claude/rules` 의 Tier 표 실측: Tier L = 5 아티팩트, PASS 임계 **0.85**, REQ/AC 상한 각 **25**.

---

## §F 감사 iter-1 대응 원장

| 감사 ID | 실측 재확인 결과 | 반영 위치 |
|---------|-----------------|-----------|
| D1 | `ResolveLaunchProfile` 호출은 런처 1곳뿐이며 전역만 읽음. 재배열 필요 확인 | plan.md §D4, AC-PM-017 |
| D2 | REQ-PM-018 이 어느 AC 행에도 없었음 확인 | acceptance.md §B 매트릭스, AC-PM-011/012 |
| D3 | Tier 표에서 M 상한 16 < REQ 23 확인 | frontmatter `tier: L`, design.md·research.md 신규 |
| D4 | `EnsureDir` 가 기록 직전이라 "디렉터리 부재" 전제 불가; `MkdirAll` 선행 실패로 "base RO" 단독 불가 확인 | plan.md §D7, AC-PM-010 |
| D5 | `EnsureDir` 호출 지점 2곳 확인 | plan.md §D5, AC-PM-018 |
| D6 | `internal/cli/web.go` 누락 확인, `projectRoot` 기보유 확인 | §C 표, REQ-PM-024, AC-PM-019 |
| D7 | 파일 잠금 부재 확인 (`RecordLastUsedProfile` 에 flock 없음) | spec.md L-003 |
| D8 | `EvalSymlinks` 성공/폴백이 다른 문자열 산출 확인 | plan.md §D1, L-002 |
| D9 | `internal/web` `TestMain` 0건 확인 | REQ-PM-022, AC-PM-014(2) |
| D10 | AC-PM-001 이 원자성을 판정하지 않음 확인 | AC-PM-020 신설 |
| D11 | `grep` 이 ugrep 임을 확인 | §E, 제약 C8, AC-PM-012/015 정규식 수정 |
| D12 | 사전순 실행으로 `main_test.go` 가드가 오염보다 선행함 확인 | plan.md §D8, AC-PM-021 |
| D13 | `findProjectRootFn` 실패 시 조기 반환 확인 | spec.md §B.1 주석, plan.md RESOLVED-3 |

## §G 감사 iter-2 대응 원장

| 감사 ID | 실측 재확인 결과 | 반영 위치 |
|---------|-----------------|-----------|
| N1 | 저장소 4개 중 2개가 `.claude.json` 부재(fresh) 확인. `createProfileDir` 가 `MkdirAll` 만 함을 확인 → bare 런치 해석 경로가 실제 도달 가능 | plan.md §D4 게이트 변수 표 + 4.5단계를 `profileName` 으로 변경, AC-PM-018 케이스 B 신설 |
| N2 | `RecordLastUsedProfileForProject` 흐름상 가드 3은 `os.CreateTemp` 에 도달하지 않음 확인 | acceptance.md AC-PM-020 (2) 유도 레시피 고정 + "CreateTemp 이후" 조건 + 미검증 표기 |
| N3 | AC-PM-001 행에서 REQ-PM-005 가 누락됐음 확인 (0.1.0 → 0.2.0 회귀) | acceptance.md §B 매트릭스 복원 |
| N4 | `quality.yaml` `development_mode: tdd` 확인 | plan.md M1 / M3 / M4 에 RED-first 절 추가 |
| N5 | AC-PM-014(1)·AC-PM-021 이 요구하는 테스트명이 계획 측에 없었음 확인 | plan.md M5-1 / M5-3 / M5-4 에 테스트 함수명 명시 |
| N6 | `GetCurrentName` 이 `CLAUDE_CONFIG_DIR` 을 먼저 읽고 설정 시 원장 미조회 확인 | spec.md REQ-PM-024 `Where` 절 + 발화 조건 주석 |
| N7 | 명명 프로필 4개 / 보유 2 / 부재 2 확인 (종전 "2개" 기재는 오독 유발) | research.md §A.2 정정 |
