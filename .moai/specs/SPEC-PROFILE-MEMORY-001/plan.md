# SPEC-PROFILE-MEMORY-001 — 구현 계획

> Tier **L** (0.2.0에서 M→L 승격, 0.2.1에서 iter-2 델타 반영, 0.2.2에서 iter-3 잔여 debt 5건 해소; REQ 24 / AC 21; 재감사 임계 **0.85**).
> 순서는 **되돌리기 어려운 결정 우선**이다. 데이터 모델 → 공개 API 형태 → 호출 순서 재배열 → 사용자 노출 문구 → 테스트 격리 순으로 배치했다. 위쪽일수록 검토 가치가 크다.

---

## §A 컨텍스트

- 리포: `/Users/goos/MoAI/moai-adk-go`, 기준 커밋 `c907db541` (main).
- 언어: Go. 문서 한국어, 코드·식별자·커밋 영어.
- 접촉 허용 범위: `internal/profile/`, `internal/cli/launcher.go`, `internal/cli/web.go` 1줄, `internal/web/app.go` 1지점, 신규·수정 테스트 파일.
- **줄 번호로 앵커하지 말 것.** 아래 인용은 모두 함수명 + 관측 가능한 동작 기준이며, 착수 시 Read/Grep으로 재확인한다.
- 설계 근거는 `design.md`, 코드 실측 원장은 `research.md` 참조.

---

## §B 실측으로 확인한 사실 (착수 전 재확인 대상)

| 사실 | 확인 방법 |
|------|-----------|
| `GetBaseDir()` 는 `BaseDirOverride` 또는 `~/.moai/claude-profiles` 만 반환, cwd 무관 | `internal/profile/profile.go` 판독 |
| 원장 키는 `lastProfileKey = "last_profile"` 단일 | 동 파일 상수 블록 |
| `RecordLastUsedProfile` 은 `isValidProfileName` 만 검사, `os.Stat` 없음 | 동 함수 본문 |
| `ResolveLaunchProfile` 은 읽기 측에서 `os.Stat` + `IsDir` 로 stale 필터 | 동 함수 본문 |
| `EnsureDir` 는 `MkdirAll` + `Setenv` 뿐, 시드·경고 없음 (멱등) | 동 함수 본문 |
| `ResolveLaunchProfile` 호출 지점은 `unifiedLaunchDefault` 단 1곳 | `grep -n ResolveLaunchProfile internal/cli/` |
| 기록 호출은 `unifiedLaunchDefault` 5단계, 디렉터리 생성은 `launchClaudeDefault` 1단계 | `internal/cli/launcher.go` |
| `unifiedLaunchDefault` 은 `findProjectRootFn()` 실패 시 즉시 에러 반환 | 동 함수 3단계 |
| 4단계 `applyGLMMode(root, profileName)` / `applyCGMode(root, profileName)` 는 해석된 `profileName` 을 소비 | 동 함수 4단계 |
| POSIX에서 `execOrSpawnClaude` 이후 코드 없음 | `launchClaudeDefault` 말미 주석 |
| 웹 쓰기 배선은 `newApp` 의 `recordLastProfile` 1지점, 핸들러는 에러 무시(advisory) | `internal/web/app.go`, `handlers.go` |
| `runWeb` 은 `projectRoot` 를 이미 보유하고 `ProfileName: profile.GetCurrentName()` 를 넘긴다 | `internal/cli/web.go` |
| `internal/cli` 는 `TestMain` + `sandboxProfileBaseDir()` 로 샌드박스됨 | `internal/cli/main_test.go` |
| **`internal/profile` 에 `TestMain` 없음** | `ls internal/profile/` — `main_test.go` 부재 |
| **`internal/web` 에 `TestMain` 없음** | `grep -rn "func TestMain" internal/web/` → 0건 |
| `TestGetBaseDir_Default` 는 `BaseDirOverride = ""` 설정 후 **복원하지 않음** (패키지 내 유일한 미복원 사례) | `internal/profile/profile_test.go` |
| `internal/web` 다수 테스트가 `newApp` 을 직접 호출 (board / agentfm_policy / agentfm_polish / m5_agentfm / schema_sections / projectnested / projectconfig) | `grep -rn "newApp(" internal/web/*_test.go` |
| 이 머신의 `grep` 은 ugrep 7.5.0 (`\s` 를 관대하게 처리) | `grep --version` |

---

## §C 결정 완료 항목

### RESOLVED-1 — 프로젝트 인지 범위는 런치 경로 + `moai web` (2026-08-02)

**결정**: 프로젝트 스코프 기억은 `moai cc | glm | cg` 런치 경로와 `moai web`(REQ-PM-024)에 적용한다. `moai profile current`(`internal/cli/profile.go`), `moai update`(`internal/cli/update.go` 2지점), `moai init`(`internal/cli/init.go`)는 무변경 — 무변경 래퍼를 통해 전역 `last_profile` 폴백을 그대로 쓴다.

**사용자 채택 근거**: 기존 사용자가 이미 보고 있는 출력을 바꾸지 않는다.

**`moai web` 만 예외인 이유 (감사 D6)**: RESOLVED-4의 클로저가 웹의 **쓰기** 를 project-scoped 로 만든다. 웹의 **읽기** 를 전역으로 두면 콘솔이 프로필 A를 표시하면서 프로젝트 B의 항목을 덮어쓰는, 클로저가 막으려던 바로 그 버그의 거울상이 생긴다. 읽기·쓰기 비대칭이 그 자체로 새 결함이므로 읽기도 함께 고친다.

**수용된 귀결 (한계 L-001)**: 프로젝트 스코프 항목이 전역값과 다른 프로젝트에서 `moai profile current` 는 **그 프로젝트에서 bare `moai cc` 가 실제로 띄울 프로필과 다른 이름을 표시할 수 있다.** `spec.md` §C L-001에 명시 기록했다. `moai web` 은 포함되지 않는다.

### RESOLVED-2 — 고아 `projects:` 항목은 완전 침묵 (2026-08-02)

**결정**: REQ-PM-008의 조용한 건너뛰기만 한다. `prune` 서브커맨드 없음, 런치 시점 고지 없음.

**사용자 채택 근거**: 현 코드베이스 어디에도 정리 메커니즘이 없으므로 침묵이 일관된 선택이다. 치유책 없는 고지는 매 런치마다 반복되어 사용자가 경고를 무시하도록 학습시킨다.

**수용된 귀결 (한계 L-002)**: 원장의 `projects:` 항목은 **무제한 단조 증가**하고 고아 항목이 영구 잔류한다.

### RESOLVED-3 — root 획득 실패 시 런치는 기존대로 중단 (2026-08-02, 감사 D1b)

**결정**: `unifiedLaunchDefault` 의 `findProjectRootFn()` 실패 시 즉시 에러를 반환하는 **현행 동작을 유지**한다. 해석을 root 획득 이후로 옮겨도(§D4) 이 조기 반환은 그대로 둔다.

**근거**: 오늘도 root 실패는 런치를 중단시킨다. "root 없이도 전역 폴백으로 계속"으로 바꾸면 기존 에러 의미를 바꾸는 스코프 확장이고, 모드 설정(4단계)이 `root` 를 필요로 하므로 어차피 진행할 수 없다.

**REQ-PM-007 과의 정합 (감사 D13)**: 따라서 REQ-PM-007 은 런치 경로에서 발화하지 않으며, **무변경 래퍼 계약**(`projectRoot == ""`)을 규정한다. 판정은 함수 직접 호출(AC-PM-005)로 한다. 이 사실을 `spec.md` §B.1 주석으로 명시했다.

### RESOLVED-4 — 웹 쓰기 배선은 클로저 (2026-08-02)

**결정**: `internal/web/app.go` `newApp` 에서 `recordLastProfile` 을 클로저로 배선한다. 필드 타입 `func(name string) error` 무변경.

---

## §D 설계 결정

### D1 — 원장 스키마: 단일 파일 + `projects:` 맵

```yaml
bypass: true                 # 레거시 키 보존 (read-modify-write)
model: claude-opus-4-6
last_profile: moai-adk       # 전역 폴백 — 유지
projects:                    # 신규
  /Users/goos/MoAI/moai-adk-go: moai-adk
  /Users/goos/MoAI/mo.ai.kr: mo.ai.kr
```

파일을 쪼개지 않는 이유: 현 기록기는 이미 `map[string]any` 로 read-modify-write 하므로 중첩 맵 추가가 레거시 키 보존과 원자 쓰기를 그대로 승계한다.

**키 정규화 (REQ-PM-009)**: 쓰기·읽기 양쪽에서 `filepath.EvalSymlinks` 를 best-effort로 적용하고 실패 시 `filepath.Clean` 로 폴백한다. macOS `/var` ↔ `/private/var` 비대칭이 테스트에서 즉시 물리는 지점이라 양쪽 동일 정규화가 필수다.

**정규화 폴백은 별도 키 네임스페이스다 (감사 D8)**: `EvalSymlinks` 성공 시 `/private/var/…`, `Clean` 폴백 시 `/var/…` 로 서로 다른 문자열이 나온다. 한 맵 안에 두 네임스페이스가 섞이면 한 프로젝트가 최악의 경우 **두 개의 키를 차지**할 수 있다(예: 프로젝트 디렉터리가 일시적으로 부재했다가 재생성된 경우). `EvalSymlinks` 는 대소문자도 정규화하지 않으므로 대소문자 무시 APFS에서 `/Users/goos/MoAI` 와 `/Users/goos/moai` 는 서로 다른 키다. 이 이중 키는 동작상 무해하며(둘 다 유효한 항목으로 해석됨) 한계 L-002의 고아 항목과 동일 분류로 수용한다.

### D2 — API 형태: 신규 project-aware 변종 + 기존 함수는 얇은 래퍼

```go
func RecordLastUsedProfileForProject(projectRoot, name string) error
func RecordLastUsedProfile(name string) error            // = ...ForProject("", name)

func ResolveLaunchProfileForProject(projectRoot, profileName string) string
func ResolveLaunchProfile(profileName string) string     // = ...ForProject("", profileName)

func GetCurrentNameForProject(projectRoot string) string
func GetCurrentName() string                             // = ...ForProject("")
```

`projectRoot == ""` 의 의미는 **프로젝트 스코프 건너뜀**이다: 기록 시 `projects:` 를 건드리지 않고 `last_profile` 만 쓰며, 해석 시 전역 폴백만 본다. 이것이 REQ-PM-007의 구현이다.

**기존 호출자 5곳 (감사 D6 — 종전 4곳 기재는 오류였다)**:

| 호출자 | 처리 |
|--------|------|
| `internal/cli/update.go` (2지점) | 무변경 — 래퍼가 전역 폴백 유지 |
| `internal/cli/profile.go` | 무변경 — `moai profile current` 출력 보존(L-001) |
| `internal/cli/web.go` | **변경** — `GetCurrentNameForProject(projectRoot)` (REQ-PM-024) |
| `internal/cli/init.go` | 무변경 |
| `internal/web/profile_crud.go` | 무변경 — 삭제 가드용 비교, 프로젝트 무관 |

시그니처를 확장하지 않고 변종을 추가하는 이유: 웹 패키지가 `recordLastProfile func(name string) error` 를 필드 타입으로 들고 있고 핸들러·테스트 시드가 그 형태에 묶여 있다.

### D3 — 웹 쓰기 배선: 시그니처 무변경 + 클로저

```go
// internal/web/app.go, newApp 내부
recordLastProfile: func(name string) error {
    return profile.RecordLastUsedProfileForProject(cfg.ProjectRoot, name)
},
```

### D4 — 해석·기록 순서 재배열 (감사 D1 — 종전 "기계적" 규정은 과소평가였다)

현 순서와 그 문제:

```
unifiedLaunchDefault (현행)
  2단계  resolved := ResolveLaunchProfile(profileName)   ← root 아직 없음. 전역만 읽음
  3단계  root, err := findProjectRootFn()                ← 여기서 root 생성. 실패 시 즉시 return
  4단계  applyGLMMode(root, profileName) / applyCGMode(root, profileName)
  5단계  RecordLastUsedProfile(originalProfile)          ← 디렉터리 아직 없음
  6단계  launchClaude(profileName, extraArgs)
           └─ launchClaudeDefault 1단계  profile.EnsureDir(profileName)   ← 여기서 생성
           └─ 말미  execOrSpawnClaude → syscall.Exec (POSIX, 이후 코드 없음)
```

두 개의 독립 문제가 겹쳐 있다. (a) 해석이 2단계에 있어 project-scoped 해석에 필요한 `root` 를 아직 못 얻는다 — 기록만 바꾸면 `projects:` 는 **쓰이기만 하고 런치에서 읽히지 않는다**(REQ-PM-003 미구현). (b) 기록이 디렉터리 생성보다 앞선다.

**채택 순서**:

```
unifiedLaunchDefault (변경 후)
  1단계   mode := resolveMode(modeOverride)
  2단계   root, err := findProjectRootFn()               ← 구 3단계를 앞으로 (RESOLVED-3: 실패 시 기존대로 즉시 return)
  3단계   originalProfile := profileName
          resolved := ResolveLaunchProfileForProject(root, profileName)
          resolved != profileName && resolved != "" 이면 안내 후 profileName = resolved
  4단계   applyGLMMode(root, profileName) / applyCGMode(root, profileName) / applyCCMode(root)
          ← 재배열 이후의 profileName(=resolved)을 소비한다. 재배열 전과 동일한 값 의미를 유지한다.
  4.5단계 (신규) **profileName**(해석 후 값) 이 named 이면:
            profile.EnsureDir(profileName)
            !profile.HasClaudeConfig(profileName) 이면 고지 1회 (D5/D6)
  5단계   originalProfile 이 named 이면:
            recordLastProfileFn(root, originalProfile)   ← 디렉터리 존재 확정. 시임 경유(D7)
  6단계   launchClaude(profileName, extraArgs)           ← EnsureDir 재호출은 no-op
```

핵심은 **구 2단계(해석)와 구 3단계(root 획득)의 자리 교환**이다. 4단계가 소비하는 `profileName` 은 재배열 이후에도 "해석 완료된 값"이라는 의미가 그대로 유지되어야 한다 — 이것이 재배열의 정확성 조건이다.

**4.5단계와 5단계의 게이트 변수가 다르다 (감사 iter-2 N1)**. 4.5단계는 **`profileName`**(해석 후 값), 5단계는 **`originalProfile`**(사용자가 `-p` 뒤에 친 값)을 본다. 두 변수를 같은 것으로 다루면 치명적 공백이 생긴다: bare `moai cc` 가 `projects:` 를 통해 `proj-one` 으로 해석되는 경우 `originalProfile == ""` 이므로, 4.5단계를 `originalProfile` 로 게이트하면 통째로 건너뛴다. 그러면 `EnsureDir` 는 6단계의 `launchClaudeDefault` 안에서 실행되는데 D5 결정에 따라 그쪽은 **침묵하는 멱등 재호출**이므로 — **결함 3을 재현하는 바로 그 경로에서 고지 횟수가 0** 이 된다. 도달 가능성은 가설이 아니다: 현재 프로필 저장소 4개 중 2개(`copy-that`, `moai-cowork`)가 `.claude.json` 부재 상태이며, 웹 콘솔의 `createProfileDir` 는 `MkdirAll` 만 하고 `.claude.json` 을 쓰지 않으므로 「콘솔에서 프로필 생성 → 저장(원장 기록) → 그 프로젝트에서 bare `moai cc`」 가 정확히 이 상태로 떨어진다.

REQ-PM-016 의 주어는 "a launch targets a named profile" 이고, `projects:` 를 통해 해석된 프로필도 **named profile 을 target 한 것**이다. 따라서 고지 게이트는 해석 후 값이어야 한다. REQ-PM-016 을 `-p` 명시 경로로 좁히는 대안은 이 SPEC이 없애려는 침묵을 그대로 복원하므로 기각한다.

**REQ-PM-015 불변식은 영향받지 않는다.** `originalProfile` 과 `profileName` 이 갈라지는 경우는 `originalProfile == ""` 일 때뿐이고, 그때는 5단계 기록이 아예 일어나지 않는다. 기록이 일어나는 모든 경우에는 여전히 두 값이 같다. `design.md` §D.3 의 불변식 서술은 유효하며 수정 불요다.

`EnsureDir` 는 `MkdirAll` + `Setenv` 로 멱등이므로 `launchClaudeDefault` 의 기존 호출은 그대로 두고 무해한 재호출이 된다.

**기각한 대안**: `launchClaude` 시그니처에 `projectRoot` 를 추가해 기록을 그 안으로 옮기는 안. `launchClaudeFunc` 는 테스트 시임이라 시그니처 변경이 테스트 다수로 파급되고, `launchClaudeDefault` 는 `EnsureDir` 이후에도 조기 반환 경로(`exec.LookPath("claude")` 실패, `--continue` 재개 실패)를 여러 개 가져 "어디에 넣어도 안전한 한 지점"이 없다.

**부수 효과 2건**: `CLAUDE_CONFIG_DIR` 이 조금 이르게 설정된다(4단계 모드 설정 이후이므로 GLM/CG env 주입과 무충돌). `EnsureDir` 실패가 더 이른 시점에 반환된다(어차피 `launchClaude` 에서 같은 에러로 실패할 경로).

**불변식 (REQ-PM-015)**: 기록이 일어나는 모든 경우 `originalProfile == profileName` 이다. `ResolveLaunchProfileForProject` 는 `profileName != ""` 이면 그대로 반환하고, 기록은 `originalProfile != ""` 일 때만 일어나기 때문이다. `@MX:ANCHOR` 로 고정한다.

**게이트 변수 표 (구현 시 혼동 방지)**:

| 단계 | 게이트 변수 | 이유 |
|------|-------------|------|
| 4.5 (EnsureDir + 고지) | `profileName` (해석 후) | 실제로 런치될 프로필이 판정 대상. bare 런치의 해석 결과도 포함해야 한다 |
| 5 (기록) | `originalProfile` (사용자 입력) | 해석 결과를 되기록하면 전역 폴백이 프로젝트 항목으로 승격되는 원치 않는 쓰기가 생긴다 |

> **`named` 의 정의 (감사 iter-3 NEW-5)**: 두 게이트의 `named` 는 `name != "" && name != "default"` 를 뜻한다. `internal/profile/profile.go` 의 기존 용법(`RecordLastUsedProfile` 이 `""` 와 `"default"` 를 거부)과 동일하다. `"default"` 를 통과시키면 `GetProfileDir("default") == ""` 이므로 `HasClaudeConfig("default") == false` 가 되어 **모든 `-p default` 런치가 잘못된 고지를 낸다.**

> **게이트를 `profileName` 으로 넓혀도 안전한 이유 (감사 iter-3 NEW-3)**: 해석기(M1-3)는 후보 이름마다 `os.Stat` + `IsDir` 로 디렉터리 존재를 확인한 뒤에만 그 이름을 반환한다(현행 `ResolveLaunchProfile` 의 stale 가드와 동일 계약). 따라서 해석으로 얻은 프로필에 대한 `EnsureDir` 는 항상 **기존 디렉터리에 대한 no-op** 이며, 사용자가 요청하지 않은 디렉터리가 새로 생기는 일은 없다. 실제 생성이 일어나는 경우는 명시 `-p <new>` 뿐이고 그것은 REQ-PM-013이 요구하는 동작이다.

### D5 — 고지 발화 지점은 4.5단계 단 하나 (감사 D5)

`EnsureDir` 호출 지점은 재배열 후에도 **둘**이다: 신규 4.5단계와 기존 `launchClaudeDefault` 1단계. "`EnsureDir` 직후"를 두 곳 모두에 적용하면 고지가 **두 번** 나간다.

**결정**: 고지는 **`unifiedLaunchDefault` 4.5단계에서만** 발화한다. `launchClaudeDefault` 의 `EnsureDir` 는 고지 없는 멱등 재호출로 남긴다.

부수 이점: 고지 로직을 `EnsureDir` 자체에 넣지 않음으로써 `moai web` / `moai update` 등 비런치 경로가 `EnsureDir` 를 부르더라도 고지가 새지 않는다(§G 안티패턴과 정합).

판정은 함수 단위(AC-PM-011)가 아니라 **런치 1회 실행에서 고지 문자열이 정확히 1회**(AC-PM-018)로 한다. 함수를 1회 호출해 1회 출력을 보는 것은 이중 발화를 구조적으로 탐지할 수 없다.

### D6 — 결함 3 탐지 술어와 문구

```go
// internal/profile/profile.go
const claudeConfigStateFile = ".claude.json"

// HasClaudeConfig reports whether the named profile directory already carries
// Claude Code account state. Decides solely on the presence of
// claudeConfigStateFile — consults no platform credential store.
// Pure predicate: writes nothing, prints nothing.
func HasClaudeConfig(name string) bool
```

술어는 profile 패키지(지식 소재지), **출력은 CLI 계층**에 둔다. 탐지 기준은 플랫폼 중립이다 — "프로필 디렉터리에 Claude 설정 상태 파일이 없다". macOS Keychain / Linux 파일이라는 인증 운반체 차이는 **자동 시드를 범위 밖으로 두는 근거**로만 쓰이고 술어에는 들어가지 않는다(REQ-PM-018).

문구 초안 (REQ-PM-020 — 존재하지 않는 명령을 광고하지 않는다):

```
Notice: profile "work" has no Claude Code configuration yet.
  Claude Code will show the login / onboarding screen on this launch.
  Account state is not inherited between profiles; sign in once and it
  persists for this profile.
```

### D7 — 기록 호출을 패키지 시임으로 (감사 D4)

```go
// internal/cli/launcher.go — findProjectRootFn / launchClaudeFunc / unifiedLaunchFunc 와 동일 패턴
var recordLastProfileFn = profile.RecordLastUsedProfileForProject
```

5단계는 `profile.RecordLastUsedProfileForProject(...)` 를 직접 부르지 않고 이 시임을 경유한다.

**이유**: REQ-PM-014(기록 실패해도 런치 계속)를 파일시스템으로 유도하려면 유일하게 동작하는 조합이 「프로필 디렉터리 선생성 → base 를 `chmod 0o500` → `MkdirAll` 은 기존 디렉터리라 nil → `os.CreateTemp(baseDir, …)` 가 EACCES」 하나뿐이고, 이마저 root 로 실행하면 chmod 가 무시된다. 시임 주입은 이 파일시스템 의존을 없애고 Windows CI 에서도 판정 가능하게 만든다.

기각한 조합 2건(둘 다 감사 D4에서 사용 불가로 판정):
- "프로필 디렉터리 부재" — 4.5단계가 기록 직전에 생성하므로 불가능한 전제다.
- "base 읽기 전용" 단독 — 4.5단계의 `MkdirAll` 이 먼저 실패해 `unifiedLaunchDefault` 가 조기 반환하므로, AC 가 요구하는 "fake `launchClaude` 가 호출됨" 전제가 성립하지 않는다.

### D8 — 테스트 격리 범위 (감사 D9, D12)

**`internal/web` 도 `TestMain` 대상이다 (REQ-PM-022)**. 이 SPEC이 `newApp` 을 수정하고, `internal/web` 에는 `TestMain` 이 없으며, 다수 테스트가 `newApp` 을 직접 호출한다(§B 표). `newTestApp` 이 시임을 스텁하는 경로만으로는 직접 호출 테스트가 커버되지 않는다. 따라서 AC-PM-014(2) 명령에 `./internal/web/` 를 포함한다.

**후행 가드가 필요하다 (감사 D12)**. Go 는 테스트 파일을 사전순으로 실행하므로 `main_test.go` < `preferences_test.go` < `profile_test.go` 이고, `main_test.go` 안의 `TestProfileBaseDirIsSandboxed` 는 **오염보다 먼저** 실행되어 항상 통과한다. 즉 M5-2(`TestGetBaseDir_Default` 복원)를 통째로 생략해도 그 가드는 GREEN 이다. 따라서 `profile_test.go` 보다 뒤에 정렬되는 파일(`zz_sandbox_guard_test.go`)에 후행 가드를 추가해 오염을 실제로 판정한다(AC-PM-021).

---

## §E 제약

| # | 제약 |
|---|------|
| C1 | `internal/template/templates/` 무접촉 |
| C2 | 하드코딩 금지 — env는 `envkeys.go`, YAML 키·파일명은 `lastProfileKey` 옆 상수 |
| C3 | 원장 쓰기는 read-modify-write + 원자 쓰기(temp + `os.Rename`) 유지 |
| C4 | 모든 신규 테스트는 `profile.BaseDirOverride` + `t.TempDir()` 로 격리, 실제 홈 무접촉 |
| C5 | 크로스 플랫폼 — `filepath.Join`, `EvalSymlinks` 실패 시 `Clean` 폴백, 파일 권한에 의존하는 테스트 금지 |
| C6 | 지나가는 리팩터링 금지 |
| C7 | 기록 실패는 런치를 막지 않는다 (stderr 경고 후 계속) |
| C8 | AC 명령의 정규식은 POSIX ERE 로만 — `\s` 금지, `[[:space:]]` 사용 (이 머신 `grep` 은 ugrep 이라 `\s` 가 통과하지만 stock BSD grep 에서는 리터럴 `s` 로 해석되어 0매치가 된다) |

---

## §F 마일스톤

### M1 — 원장 스키마 + 해석기/기록기 (되돌리기 가장 어려움)

`internal/profile/profile.go`:

1. 상수 추가: `projectsKey = "projects"`, `claudeConfigStateFile = ".claude.json"`.
2. `normalizeProjectKey(projectRoot string) string` — `EvalSymlinks` best-effort + `Clean` 폴백. 빈 문자열은 빈 문자열로.
3. `ResolveLaunchProfileForProject(projectRoot, profileName string) string` — 명시 `-p` 우선 → 옵트아웃 검사 → 원장 로드 → `projects[key]` → `last_profile` → `""`. 각 후보마다 `isValidProfileName` + `os.Stat`/`IsDir` 검증(REQ-PM-008).
4. `RecordLastUsedProfileForProject(projectRoot, name string) error` — 기존 이름 형태 검사에 **디렉터리 존재 검사 추가**(REQ-PM-011), 그 다음 `last_profile` 과 (projectRoot 비어있지 않으면) `projects[key]` 를 함께 기록. read-modify-write + 원자 쓰기 유지(REQ-PM-005).
5. `GetCurrentNameForProject(projectRoot string) string`.
6. `HasClaudeConfig(name string) bool` — 순수 술어(REQ-PM-018).
7. 기존 3함수를 래퍼로 축소.

**TDD 순서 (RED first)**: 이 리포는 `quality.yaml` `development_mode: tdd` 이며 manager-develop §E8 은 GREEN 이전의 RED 출력 verbatim 을 요구한다. 이 마일스톤의 AC(AC-PM-001~008, 011, 019, 020)를 덮는 테스트 — `TestRecordForProject_PreservesLegacyKeys`, `TestResolveForProject_ProjectScopeWinsOverGlobal`, `TestResolveForProject_LegacyLedgerUnchanged`, `TestResolveForProject_OptOutDisablesBothLookups`, `TestForProject_EmptyRootFallsBackToGlobal`, `TestResolveForProject_StaleProjectEntrySkipped`, `TestProjectKey_NormalizationSymmetric`, `TestRecordForProject_RejectsMissingDirectory`, `TestHasClaudeConfig_DecidesOnClaudeJSONAlone`, `TestGetCurrentNameForProject_ProjectScoped`, `TestRecordForProject_NoPartialStateOnFailure` — 를 **먼저 RED 로 작성**하고 실패 출력을 확보한 뒤 구현한다.

### M2 — 호출자 배선 (D3, RESOLVED-1)

1. `internal/web/app.go` `newApp` 의 `recordLastProfile` 을 클로저로 교체 (REQ-PM-010).
2. `internal/cli/web.go` `runWeb` 의 `ProfileName` 을 `profile.GetCurrentNameForProject(projectRoot)` 로 교체 (REQ-PM-024). `projectRoot` 는 같은 함수 앞부분에서 이미 얻은 값을 재사용한다.
3. 나머지 4개 호출자(§D2 표)는 무변경 확인만 한다.

### M3 — 해석·기록 순서 재배열 + 시임 (D4, D7)

1. `internal/cli/launcher.go` 에 `var recordLastProfileFn = profile.RecordLastUsedProfileForProject` 시임 추가.
2. `unifiedLaunchDefault` 을 §D4 「채택 순서」로 재배열: root 획득을 해석보다 앞으로, 해석을 `ResolveLaunchProfileForProject(root, profileName)` 로 교체, 4.5단계 신설, 기록을 시임 경유로 교체.
3. 4단계가 재배열 이후의 `profileName` 을 소비함을 확인한다.
4. `@MX:ANCHOR` 로 순서 불변식(root → resolve → mode → EnsureDir → record → exec) 고정.

**TDD 순서 (RED first)**: `TestUnifiedLaunch_UsesProjectScopedResolution`(AC-PM-017), `TestUnifiedLaunch_FirstTimeNewProfileIsRecorded`(AC-PM-009), `TestUnifiedLaunch_RecordFailureDoesNotBlockLaunch`(AC-PM-010) 를 먼저 RED 로 작성한다. 특히 AC-PM-017 은 현행 코드에서 `"global-one"` 을 받아 반드시 실패해야 하며, 그 실패 출력이 D1 결함의 존재 증거다.

### M4 — 새 프로필 고지 (D5, D6)

1. 고지 함수를 주입 가능한 `io.Writer` 로 받는다. **`warnNoModelResolved` 선례는 함수 시그니처만 공급한다** — 그 호출 지점은 `warnNoModelResolved(os.Stderr, profileName)` 로 `os.Stderr` 를 하드코딩하므로 함수는 테스트 가능하지만 **런치 경로 출력은 테스트할 수 없다.** AC-PM-018이 단언하는 것은 바로 그 런치 경로 출력이다.
2. 따라서 **stderr 시임을 별도로 선언**한다(감사 iter-3 NEW-1):
   ```go
   // internal/cli/launcher.go — 기존 5개 시임과 동일한 패키지 레벨 var 패턴
   var launcherStderr io.Writer = os.Stderr
   ```
   `internal/cli` 에는 현재 `io.Writer` 시임이 하나도 없다(`grep -rn '^var .* io\.Writer' internal/cli/` → 0건). 4.5단계의 고지와 5단계의 기록 실패 경고(`"Warning: failed to record last-used profile: %v"`) **둘 다** 이 시임으로 출력한다 — AC-PM-010(c)와 AC-PM-018 세 케이스가 모두 이 주입에 의존한다.
3. 발화 지점은 4.5단계 **단 하나**, 게이트 변수는 `profileName`(해석 후 값). `launchClaudeDefault` 의 `EnsureDir` 는 고지 없이 둔다.

**TDD 순서 (RED first)**: `TestFreshProfileNotice_WriterContent`(AC-PM-011), `TestUnifiedLaunch_FreshProfileNoticeEmittedExactlyOnce`(AC-PM-018 — 명시 `-p` 케이스와 bare 런치 케이스 **둘 다**) 를 먼저 RED 로 작성한다. bare 런치 케이스는 4.5단계를 `originalProfile` 로 게이트한 구현에서 횟수 0 으로 실패해야 한다(N1 회귀 가드).

### M5 — 테스트 격리 정비 (D8, 선행 하자 제거)

1. `internal/profile/main_test.go` 신규 — `sandboxProfileBaseDir()` 헬퍼 + `func TestMain(m *testing.M)` + 가드 테스트 **`TestProfileBaseDirIsSandboxed`**(`internal/cli/main_test.go` 의 동명 테스트와 같은 이름·같은 단언). AC-PM-014(1)이 이 이름을 두 패키지에서 각 1회, 총 2회 관측한다.
2. `internal/profile/profile_test.go` `TestGetBaseDir_Default` 수정 — `orig` 저장 + `t.Cleanup` 복원.
3. `internal/profile/zz_sandbox_guard_test.go` 신규 — 후행 가드 테스트 **`TestSandboxSurvivesPackageRun`**. `profile_test.go` 보다 뒤에 정렬되며, M5-2를 생략하면 실패해야 한다(AC-PM-021).
4. `internal/web/main_test.go` 신규 — 동일 패턴의 `func TestMain(m *testing.M)` + 가드 테스트 **`TestProfileBaseDirIsSandboxed`**(REQ-PM-022). AC-PM-014(1)의 기대값 2를 구성하는 두 번째 인스턴스다.

### M6 — AC 검증

`acceptance.md` 의 AC 전항 실행 후 증거 기록.

---

## §G 안티패턴

- **기록만 project-aware 로 바꾸고 해석은 그대로 두기** — `projects:` 가 쓰이기만 하고 런치에서 읽히지 않는다(iter-1 감사 D1의 치명 결함). M3의 해석 교체와 M1의 기록 교체는 한 쌍이며 분리 머지 금지.
- **가드만 넣고 순서는 안 옮기기** — 최초 `-p <new>` 런치가 조용히 기록되지 않는다.
- **`projects:` 를 쓰면서 `last_profile` 은 안 쓰기** — 다운그레이드한 바이너리가 기억을 통째로 잃는다.
- **쓰기만 정규화하고 읽기는 안 하기** — macOS `/private/var` 비대칭으로 항목이 절대 매치되지 않는다.
- **웹 쓰기만 project-scoped 로 만들고 읽기는 전역으로 두기** — 콘솔이 A를 표시하며 B를 덮어쓰는 거울상 버그가 생긴다(감사 D6).
- **고지를 `EnsureDir` 안에 넣기** — 비런치 경로(`moai web`, `moai update`)로 새어 나가고, 두 호출 지점에서 이중 발화한다.
- **고지 판정을 함수 1회 호출로 하기** — 이중 발화를 구조적으로 탐지 못 한다. 런치 1회 실행 기준으로 센다.
- **파일 권한(chmod)으로 기록 실패를 유도하기** — root 실행에서 무력화되고 Windows CI 에서 판정 불가. 시임 주입을 쓴다.
- **테스트에서 실제 홈 읽기** — 이 리포에는 `go test` 가 실제 `launch.yaml` 을 덮어쓴 사고 기록이 있다.
- **AC 정규식에 `\s` 쓰기** — 이 머신은 ugrep 이라 통과하지만 stock BSD grep 에서 0매치가 된다.
- **줄 번호 앵커** — AC는 함수명·관측 동작으로만 앵커한다.

---

## §H 교차 참조

- `spec.md` — REQ 원장 (24건)
- `acceptance.md` — AC 행렬 (21건)
- `design.md` — 원장 스키마·API 표면·호출 순서 설계
- `research.md` — 코드 실측 원장 + 감사 iter-1 대응 표
- `internal/cli/main_test.go` — 샌드박스 `TestMain` 선례
- `internal/cli/launcher.go` `warnNoModelResolved` — 주입 `io.Writer` 경고 선례
