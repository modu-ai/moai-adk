# SPEC-CODEX-LAUNCHER-001 — 인수 기준

모든 기준은 기계 판정이다. codex 실 바이너리에 의존하는 항목은 없다 — 프로브 seam (`codexLookPath`, `codexCommandRunner`, `codexLoginStatusRunner`) 을 스텁해 판정한다.

## 판정 어휘 (전 AC 공통)

이 문서에서 다음 낱말은 정해진 형태의 단언만 가리킨다. 산문으로 기대를 적으면 "그럴듯하면 통과" 가 되므로 형태를 못 박는다.

- **정확 일치** — 시험 패키지에 선언된 **이름 붙은 상수** 와의 문자열 동등(`==`) 비교. 부분 일치(`strings.Contains`)·정규식·grep 은 정확 일치가 아니다.
- **폐집합 단언** — "무엇이 없다" 를 금지 항목 열거로 세지 않고 "허용된 것의 집합과 같다" 로 센다. 열거식 금지는 열거되지 않은 형태를 놓치므로 이 SPEC 은 원칙적으로 폐집합을 쓴다.
- **포착 seam** — 이 SPEC 이 추가하는 시험 하네스는 프로세스를 띄우는 두 자리(현재 셸에서 자식을 띄우는 **직접 기동 경로** 와 tmux 새 창을 여는 `spawnLaunch` 경로) **모두** 를 한 기록기로 감싼다. 기록 항목은 `(program, argv, cwd, 형태, stdin, stdout, stderr)` 일곱 가지이며, "기동 0회" 류의 단언은 언제나 **두 자리의 합** 을 센다. 한쪽만 세는 단언은 다른 쪽으로 새는 구현을 못 잡는다.
  - 뒤 세 항목은 **실제로 `exec.Cmd` 에 대입된 값** 이어야 한다 — 기록기는 기동 직전의 `*exec.Cmd` 의 `Stdin` / `Stdout` / `Stderr` 필드를 그대로 읽는다. 별도로 계산한 값을 기록하고 `exec.Cmd` 에는 다른 것을 대입하는 구현은 이 규정으로 배제된다.
  - 이 SPEC 은 **프로세스를 교체하지 않는다.** 두 자리 모두 자식 프로세스를 띄우고 그 종료코드를 전파한다 (plan §C.5).
- **이 SPEC 이 추가한 파일** — `internal/cli/codex_launcher.go`, `internal/cli/codex_readiness.go`, 그리고 `internal/cli/mcp_codex.go` 의 이 SPEC diff 부분. 정적 판정의 범위는 이 집합이며, 시험이 파일 목록을 상수로 들고 있는다.

## AC-CL-001 — 커맨드 등록 (REQ-CL-001)

`codex` 행이 도움말 어딘가에 있다는 것만으로는 "`cc` / `glm` / `cg` 의 형제" 가 아니다. 제목이 같은 **별도 그룹** 을 만들어도 같은 낱말이 화면에 나오기 때문이다.

- **Given** 빌드된 moai 바이너리
- **When** `moai --help` 를 실행하면
- **Then** LAUNCH COMMANDS 그룹 헤딩의 출현 횟수가 **정확히 1** 이다.
- **And** 그 헤딩 다음 줄부터 첫 빈 줄까지가 그룹 블록이며, 그 블록 각 줄의 첫 토큰을 모은 집합이 `{cc, glm, cg, codex}` 를 **포함** 한다 (네 이름이 같은 블록에 있다).
- **And** 심볼로 직접 비교한다 — `codexCmd.GroupID == ccCmd.GroupID` 이고 `codexCmd.Parent() == rootCmd` 다 (그룹 ID 문자열을 시험에 다시 적지 않는다; 다시 적으면 두 곳이 함께 틀릴 수 있다).
- **And** `moai codex --help` 가 rc 0 으로 도움말을 출력한다.

## AC-CL-002 — 동사 라우팅 (REQ-CL-002)

포착 seam 은 호출 횟수뿐 아니라 **argv 전체와 cwd** 를 포착한다 — 명령 이름만 맞고 인자나 실행 위치가 틀린 구현이 통과하지 못하게.

- **Given** 라우팅 표를 심볼로 직접 읽으면
- **Then** 기동을 일으키는 동사 집합이 `{cli, app}` 과 **같고**, 리드아웃 형태를 만드는 토큰 집합이 `{"" (맨몸), status}` 와 **같다** — 폐집합 단언이다. 미지 토큰을 `default` 분기로 기동에 떨어뜨리는 구현은 이 등식에서 걸린다.
- **When** 맨몸 / `status` / `cli` / `app` 각각을 호출하면
- **Then** 맨몸과 `status` 는 기동 **0회**, `cli` 와 `app` 은 각 1회다 (두 자리의 합).
- **And Given** 미지 토큰 여섯 개 — `bogus` · `cl` · `CLI` · `Cli` · `--model` · `-x`
- **Then** 여섯 칸 모두 기동 **0회**, rc 비영, 그리고 진단이 사용법 상수와 **정확 일치** 한다. 명시 동사를 실제로 *요구* 하는지는 이 부정 방향 칸에서만 관측된다.
- **And** `cli` 의 포착 cwd 를 **세 위치에서 교차** 한다: (1) 프로젝트 루트에서 호출 → 루트, (2) 프로젝트 루트의 **하위 디렉터리**(`internal/cli/`) 에서 호출 → **여전히 루트**, (3) worktree 의 하위 디렉터리에서 호출 → 그 **worktree 루트**. 세 칸 모두 정확 일치. (2) 가 없으면 `os.Getwd()` 를 그대로 쓰는 구현과 구별되지 않는다.
- **And** `app` 의 포착 argv 는 정확히 `[codex, app]` 이다.
- **And Given** 직접 기동 seam 이 종료코드 `0` · `1` · `2` · `126` · `127` 을 각각 돌려주는 다섯 칸
- **Then** `cli` · `app` 각각에서 moai 의 rc 가 그 다섯 값과 **각각 같다**. 성공만 0 으로 맞추고 실패를 한 값으로 뭉개는 구현은 세 비영 값이 서로 다른 이 칸에서 갈린다.
- **And** `cli` · `app` 의 직접 기동 칸에서 포착된 `stdin` / `stdout` / `stderr` 세 값이 각각 그 시험 프로세스 자신의 `os.Stdin` / `os.Stdout` / `os.Stderr` 와 **항등** 이다 (`==` 로 비교하는 인터페이스 값 동등이며, "같은 종류" 나 "nil 아님" 이 아니다). 파이프 · `bytes.Buffer` · `io.MultiWriter` · `io.Discard` · `nil` 어느 것으로 바꿔도 이 칸에서 걸린다. 대화형 tty 상속은 기계적으로 이 항등에 달려 있으므로, 관측 불가능한 tty 대신 그 **선행 조건** 을 판정한다 (tty 왕복 자체는 「판정 제외」의 Gap).
- **And Given** `moai codex cli -- --model o3 "a b" '$x' --flag=v` 처럼 `--` 뒤에 공백·인용·`$`·`=` 를 포함한 인자를 주면
- **Then** 포착 argv 의 `codex` 뒤 꼬리가 이 토큰들과 **정확히 일치** 한다 (개수·순서·문자 모두; 셸 재해석이나 재인용으로 변형되지 않는다).
- **근거**: 맨몸의 의미는 리드 판정으로 (b) "리드아웃 + 명시 기동" 으로 확정됐다 (plan §B).

## AC-CL-003 — `--spawn` 패리티 (REQ-CL-003)

- **Given** tmux 존재를 흉내 낸 스텁 spawn seam
- **When** `moai codex cli --spawn` 과 `moai codex app --spawn` 을 각각 실행하면
- **Then** 둘 다 `spawnLaunch` 를 정확히 1회 호출하고, **직접 기동 경로의 호출은 0회** 다 (성공 경로 — 실패만 시험하면 아무것도 못 띄우는 구현이 통과한다). 두 자리를 함께 세므로 `--spawn` 을 무시하고 현재 셸에서 띄워 버리는 구현이 여기서 걸린다.
- **And Given** `moai codex cli --spawn -- --model o3 "a b" '$x' --flag=v` — AC-CL-002 와 **같은 꼬리 입력**
- **Then** spawn 에 포착된 `(program, argv)` 가 같은 입력의 **직접 기동 경로 포착값과 토큰 단위로 동일** 하다. 두 포착을 서로 비교하는 것이 요점이다 — 꼬리 단언이 직접 기동 경로에만 걸려 있으면 spawn 에서 꼬리를 버리는 구현이 그대로 통과한다. 여기서 `program`/`argv` 는 tmux 가 아니라 **새 창에서 실행될 대상**(codex) 을 가리킨다 — 기록 대상이 tmux 면 이 등식은 성립할 수 없다.
- **And Given** tmux 부재를 흉내 낸 환경
- **Then** codex 쪽 진단 바이트가 같은 조건의 `moai cc --spawn` 진단 바이트와 **정확 일치** 하고, 기동은 0회다 ("동일 계열" 이라는 산문 대신 바이트 동등).
- **And** 그 진단 문안의 문자열 리터럴은 비시험 Go 원본 전체에서 **정확히 1회** 나타난다 — 같은 바이트를 자기 파일에 복사한 두 번째 사본은 여기서 걸린다. 진단은 공유 상수 하나여야 한다.
- **And** 맨몸 / `status` 에 `--spawn` 을 주면 rc 비영으로 거부하고 기동 0회다 — 리드아웃은 새 창에서 띄울 대상이 아니다.

## AC-CL-004 — 리드아웃이 실제 값을 보고한다 (REQ-CL-004)

라벨 존재가 아니라 **값** 을 판정한다. 각 칸의 배선 행 전문을 이름 붙은 기대 상수와 **정확 일치** 로 비교한다 — `partial` 이라는 낱말이 들어갔는지를 세는 판정은 쓰지 않는다. 어느 파일이 없는지 틀리게 적어도 통과하기 때문이다.

| 배선 상태 (fixture) | 기대 배선 행 |
|---|---|
| `.codex/` 없음 | `not wired` + 조치 문구 |
| **`.codex/` 는 있고 비어 있음** | **`not wired`** + 조치 문구 — 디렉터리 존재만 보는 구현이 떨어지는 행 |
| `hooks.json` 만 있음 | `partial` + **없는 쪽** `.codex/config.toml` 을 명명 + 조치 문구 |
| `config.toml` 만 있음 | `partial` + **없는 쪽** `.codex/hooks.json` 을 명명 + 조치 문구 |
| 둘 다 정상 | `wired` — 조치 문구 **없음** |
| 둘 다 있고 `hooks.json` 이 화이트리스트 위반 키 보유 | `invalid` + 조치 문구 |

- **And** 두 `partial` 칸은 서로에 대해 **배타** 다: `hooks.json` 만 있는 칸의 배선 행은 `config.toml` 을 포함하고 `hooks.json` 을 **포함하지 않으며**, 반대 칸은 정확히 반대다. 상태와 무관하게 한 문안을 고정 출력하는 구현은 두 칸 중 하나에서 반드시 깨진다.
- **And** 배선 행의 상태 토큰은 폐집합 `{not wired, partial, wired, invalid}` 의 원소다 — 이 넷 밖의 어떤 낱말도 상태 자리에 오지 않는다.

**조치 문구** 는 `moai init --agent codex` 이며, REQ-CL-006 이 요구하는 대로 **불완전한 다섯 상태 전부** 에 나타나야 한다 — 부재 상태에만 붙이는 구현은 빈·부분·불량 배선에서 사용자에게 다음 행동을 알려주지 않는다.

- **And** 맨몸과 `status` 두 형태가 여섯 행에서 **같은 값** 을 보고한다 (형태에 따라 판정이 갈리지 않는다).
- **And** 이 판정은 `SPEC-CODEX-INIT-001` 이 소비하는 단일 정의다 — 그 SPEC 은 자체 파일 검사를 하지 않는다 (AC-CI-002).
- **And** 바이너리 경로 · 버전 · auth 세 행은 스텁이 공급한 값(`/sentinel/path/codex` · `SENTINEL-VER-9x9` · `sentinel-provider`)과 각각 **정확 일치** 한다. auth 행에 값 단언이 없으면 프로브 결과와 무관하게 `unknown` 을 찍는 구현이 통과한다.
- **And** 화이트리스트 판정은 `codexadapter.ValidateConfig` 를 호출해 얻는다 (AC-CL-007 이 재구현 부재를 강제).

## AC-CL-005 — CODEX_HOME 해석과 출처 표시 (REQ-CL-005)

출처 라벨은 폐집합 `{env, default}` 의 원소이며 각 칸에서 **정확 일치** 로 본다.

| `CODEX_HOME` 상태 | 기대 값 | 기대 출처 |
|---|---|---|
| `/tmp/xyz` | `/tmp/xyz` | `env` |
| 미설정 | `<home>/.codex` | `default` |
| **설정됐으나 빈 문자열** | `<home>/.codex` | **`default`** — `LookupEnv` 의 `ok` 만 보는 구현이 죽는 칸 |
| **공백뿐(`"   "`)** | `<home>/.codex` | **`default`** |

- **And Given** 홈 해석 seam 을 스텁해 **끝에 구분자가 붙은** 경로(`/tmp/h/`)를 돌려주게 하면
- **Then** 폴백 값은 `filepath.Join` 결과(`/tmp/h/.codex`)와 정확 일치한다 — 문자열 접합판(`home + "/.codex"`)은 구분자가 겹쳐 이 칸에서 갈린다.
- **And** 그 스텁이 정확히 1회 호출되고, `HOME` 을 지운 상태에서도 위 네 칸이 같은 결과를 낸다 — 홈 해석이 `HOME` 직독이 아니라 seam(`os.UserHomeDir` 계열)을 지난다는 뜻이다 (§E 크로스 플랫폼).

## AC-CL-006 — 배선 없는 프로젝트는 정보성 (REQ-CL-006)

AC-CL-004 의 여섯 상태 중 **불완전한 다섯 상태 전부** 를, **맨몸과 `status` 두 형태 각각** 으로 돈다 (10칸).

- **When** 각 칸을 실행하면
- **Then** rc 는 0 이고, 배선 행이 AC-CL-004 의 기대 상수와 정확 일치하며, 출력에 `moai init --agent codex` 문자열이 포함된다.
- **And** 열 칸 모두에서 리드아웃 전문은 **stdout 으로만** 나가고 **stderr 는 0바이트** 다 — rc 만 재는 판정은 오류 스트림으로 보고하는 구현을 못 잡는다.
- **And** 열 칸의 출력에서 대소문자 무시 `error` · `failed` · `fatal` · `broken` · `cannot` · `unable` 여섯 낱말의 히트 합이 **0** 이다 (정보성이라는 요구의 어휘 측면). 상태 자리 자체는 AC-CL-004 의 폐집합이 이미 고정한다.
- **And** 배선 완료 상태에서는 `moai init --agent codex` 가 **나타나지 않는다** — 조치가 필요 없는데 조치를 권하지 않는다.

## AC-CL-007 — 분류 구현 단일성 (REQ-CL-007, REQ-CL-010)

텍스트 중복 부재만으로는 "공유 프로브를 실제로 쓴다" 를 증명하지 못한다. 호출 횟수 `≥1` 로도 부족하다 — "불렀다" 만 증명하고 "그 값을 썼다" 는 증명하지 못한다. **sentinel 전파** 로 판정하되, 런처 한 표면이 아니라 **소비자 축으로 교차** 한다:

- **Given** 공유 프로브가 실제 값과 구별되는 sentinel 을 돌려주도록 스텁 (버전 `SENTINEL-VER-9x9`, 바이너리 경로 `/sentinel/path/codex`, auth `sentinel-provider`)
- **When** 같은 스텁 하나로 (a) 런처 리드아웃 조립, (b) `codex_setup` MCP 도구 응답, (c) `moai web` 콘솔 Codex 카드 렌더를 각각 1회 수행하면
- **Then** 세 표면 **모두** 에서 `sentinel-provider` 가 그대로 나타난다 — 한 표면이라도 빠지면 그 표면은 공유 프로브가 아닌 다른 경로에서 값을 얻고 있다는 뜻이다. 런처에만 거는 단언은 web · MCP 소비자를 한 번도 돌리지 않는다.
- **And** 런처 리드아웃에서는 세 sentinel 값이 모두 나타난다.
- **And Given** `codexadapter.ValidateConfig` 가 sentinel 위반을 돌려주도록 스텁
- **Then** 배선 행이 그 위반을 반영한다 (호출만 하고 자체 검증 결과를 쓰지 않았다면 실패한다).
- **And** **분류 원천의 폐집합 단언**: `internal/` 하위 비시험 `*.go` 중 provider 리터럴(`"chatgpt"` 또는 `"apiKey"`)을 담은 파일의 집합이 `{internal/cli/mcp_codex.go}` 와 **같다**. 두 번째 분류 경로는 — 명령 문구를 쓰든 `auth.json` 을 직접 열든 — 어딘가에서 이 두 리터럴 중 하나를 만들어야 하므로 이 등식에서 걸린다.
- **And** 보조 확인: `grep -rn "login status" internal/ --include="*.go" | grep -v _test` 의 히트는 공유 분류기 한 곳뿐이다.
- **And** `git diff` 에 `codexCommandRunner` 인터페이스 선언의 변경이 없다 (REQ-CL-010 후단).

## AC-CL-008 — auth 분류 2단 사다리 (REQ-CL-008) [핵심]

**1단 — 순수 파일 판정 `classifyCodexAuthFile(raw []byte) (provider, ok, err)` 을 직접 시험한다** (디스크 접근 없음). 판정 기준은 **비어 있지 않은 값** 이지 객체의 존재가 아니다:

| `auth.json` 내용 | 기대 `(provider, ok)` |
|---|---|
| `{"auth_mode":"chatgpt","tokens":{"access_token":"x"}}` | `("chatgpt", true)` |
| **`{"auth_mode":"chatgpt","tokens":{}}`** | **`("", false)`** — 빈 객체는 자격 재료가 아니다 |
| `{"auth_mode":"chatgpt","tokens":{"access_token":null,"id_token":null}}` | `("", false)` |
| `{"auth_mode":"chatgpt","tokens":{"access_token":"","id_token":""}}` | `("", false)` |
| `{"auth_mode":"chatgpt"}` (tokens 없음) | `("", false)` |
| `{"auth_mode":"chatgpt","tokens":null}` | `("", false)` |
| `{"auth_mode":"apikey","OPENAI_API_KEY":"x"}` | `("apiKey", true)` |
| `{"auth_mode":"apikey","OPENAI_API_KEY":null}` | `("", false)` |
| `{"auth_mode":"apikey","OPENAI_API_KEY":""}` | `("", false)` |
| **`{"auth_mode":"apikey","OPENAI_API_KEY":"   "}`** | **`("", false)`** — 공백뿐. 모드마다 다른 잣대를 쓰는 구현이 죽는 칸 |
| `{"auth_mode":"totally-new-mode","tokens":{"access_token":"x"}}` | `("", false)` — 추측하지 않는다 |
| **`{"auth_mode":123,"tokens":{"access_token":"x"}}`** | **`("", false)`** — 모드가 문자열이 아니다 |
| **`{"auth_mode":["chatgpt"],"tokens":{"access_token":"x"}}`** | **`("", false)`** — 모드가 문자열이 아니다 |
| **`{"auth_mode":"CHATGPT","tokens":{"access_token":"x"}}`** | **`("", false)`** — 모드는 알려진 값과 정확히 같아야 한다 |
| `{` (파싱 실패) | `("", false)` + `err != nil` |
| 빈 바이트 | `("", false)` |
| **`{"auth_mode":"chatgpt","tokens":{ }}`** (공백 하나) | **`("", false)`** — 원문 문자열 비교판이 뚫리는 지점 |
| `{"auth_mode":"chatgpt","tokens":[]}` | `("", false)` — 객체가 아니다 |
| `{"auth_mode":"chatgpt","tokens":"x"}` | `("", false)` — 객체가 아니다 |
| `{"auth_mode":"chatgpt","tokens":{"access_token":false}}` | `("", false)` — 문자열이 아니다 |
| `{"auth_mode":"chatgpt","tokens":{"access_token":0}}` | `("", false)` — 문자열이 아니다 |
| `{"auth_mode":"chatgpt","tokens":{"access_token":"   "}}` | `("", false)` — 공백뿐 |
| **`{"auth_mode":"chatgpt","tokens":{"irrelevant":"x"}}`** | **`("", false)`** — 무관한 키는 자격 재료가 아니다 |
| `{"auth_mode":"chatgpt","tokens":{"account_id":"x"}}` | `("", false)` — 계정 메타데이터만으로는 로그인이 아니다 |
| `{"auth_mode":"chatgpt","tokens":false}` | `("", false)` |
| `{"auth_mode":"chatgpt","tokens":0}` | `("", false)` |
| **`{"auth_mode":"apikey","OPENAI_API_KEY":false}`** | **`("", false)`** — 문자열이 아니다 |
| `{"auth_mode":"apikey","OPENAI_API_KEY":0}` | `("", false)` |
| `{"auth_mode":"apikey","OPENAI_API_KEY":[]}` | `("", false)` |
| `{"auth_mode":"apikey","OPENAI_API_KEY":{}}` | `("", false)` |

굵은 행들이 가르는 층위:

1. `{ }` — 원문 바이트를 문자열과 비교하는 구현이 뚫린다.
2. `false` — JSON 타입을 안 보는 구현이 뚫린다.
3. `{"irrelevant":"x"}` — **타입은 보는데 키를 안 보는** 구현이 뚫린다.
4. `apikey` × 공백뿐 — **모드마다 다른 잣대** 를 쓰는 구현이 뚫린다 (공백 제거를 chatgpt 쪽에만 거는 형태).
5. `auth_mode` 의 타입 오염(`123` / `["chatgpt"]`)과 대소문자(`"CHATGPT"`) — 모드 값을 느슨하게 받는 구현이 뚫린다.

이 층위를 동시에 만족시키는 규칙은 하나뿐이다: **자격 재료는 인정된 키 집합에 속하는, 비어 있지 않은 JSON 문자열이어야 하고, 모드는 알려진 값과 정확히 같은 JSON 문자열이어야 한다.** ChatGPT 모드가 인정하는 키는 로그인 자격을 실제로 담는 것들(`access_token` / `id_token` / `refresh_token`)이고, `account_id` 같은 계정 메타데이터는 자격 재료가 아니다.

**비밀값 규율 — 값을 보존할 수 있는 어떤 형태도 두지 않는다.**

- **When** 역직렬화 대상 타입 집합(`codexAuthFile` 및 그 **중첩 타입 전부**, 재귀)의 필드를 리플렉션으로 열거하면
- **Then** 그 필드 kind 의 집합이 **폐집합** `{string, bool, int, struct}` 안에 들어가고, `string` 인 필드는 `auth_mode` 태그를 가진 **하나뿐** 이다. `map` · `slice` · `array` · `interface` · `pointer` 어느 것도 0건이다. 금지 타입을 열거하는 대신 허용 집합을 고정하므로, 열거에 없던 보존 형태(`map[string]string` · `[]string` · `fmt.Stringer` 구현체)도 여기서 걸린다.
- **And Given** 가짜 토큰 문자열 `SENTINEL-TOKEN-9x9` 를 심은 fixture 를 `readCodexAuthFile` (경로를 아는 유일한 층) 에 먹이고, 그 파일을 파싱 실패하도록 만들면
- **When** 그 호출 동안의 **네 채널을 모두 포착** 하면 — 반환 오류의 `Error()` 전문, 포착한 stdout, 포착한 stderr, `log` 출력 싱크
- **Then** 네 채널 전부에서 `SENTINEL-TOKEN-9x9` 히트가 0 이다. REQ 는 `retained, logged, or wrapped` 셋을 요구하므로, 오류 전문만 보는 판정은 로그로 새는 구현을 못 잡는다.
- **And** 같은 sentinel 을 심은 fixture 로 리드아웃을 끝까지 조립했을 때도 네 채널 히트가 0 이다.

**2단 — 순수 파서 `parseCodexAuthLine(combined, exitCode)` 을 직접 시험한다** (프로세스 없음). 표는 AC-CL-009.

**결합 규칙 — 프로덕션 경로를 실제로 시험한다** (이 SPEC 이 고치려는 결함이 여기 있으므로 스텁으로 대신하지 않는다). fixture 실행 파일을 **3종** `testdata/` 에 커밋하고 세 칸 모두 돈다 (셸 경유 없이 `exec` 로 직접 실행):

| fixture | stdout | stderr |
|---|---|---|
| stderr 전용 | 0바이트 | `Logged in using ChatGPT` |
| **stdout 전용** | `Logged in using ChatGPT` | 0바이트 |
| 양쪽 | `noise from stdout` | `Logged in using ChatGPT` |

- **When** `defaultLoginStatusRunner` 를 각 fixture 에 대고 호출하면
- **Then** 반환된 두 스트림이 위 표대로 **분리 수집** 된다 (각 칸 정확 일치).
- **And** `combineCodexStreams(stdout, stderr)` 의 결과가 세 칸 모두에서 그 칸의 **비어 있지 않은 줄을 전부** 담는다 — 양쪽 칸에서는 두 줄이 다 담기며 줄 수는 2 다. 한쪽 스트림을 버리는 구현은 stdout 전용 칸이나 양쪽 칸에서 반드시 깨진다. fixture 가 stderr 한 축뿐이면 그 구현과 올바른 구현이 구별되지 않는다 — 이 SPEC 이 고치려는 결함의 정확한 거울상이다.
- **플랫폼**: 이 세 칸만 Windows 에서 skip. 순수 함수 시험은 전 플랫폼.

**통합 3건** — 사다리가 실제로 이어지는지. 최종값 스텁이 아니라 **저수준 러너 스텁** 으로 돈다:

| given | 기대 |
|---|---|
| `auth.json` **존재 + 유효**(`auth_mode=chatgpt` + 유효 토큰) | `AuthProvider == "chatgpt"` **이고 러너 스텁 호출 0회** — 1단이 조립부에 실제로 배선됐다는 뜻. 1단 함수를 만들어 놓고 부르지 않는 구현이 여기서 죽는다 |
| `auth.json` 부재 + 러너가 `stdout=[]`, `stderr=[Logged in using ChatGPT]`, rc 0 | `AuthProvider == "chatgpt"`, 러너 호출 1회 |
| `auth.json` 부재 + 러너가 `stdout=[Logged in using ChatGPT]`, `stderr=[]`, rc 0 | `AuthProvider == "chatgpt"` — 프로덕션 조립부가 두 스트림을 다 읽는지 |

- **기준선 근거**: 두 번째 칸은 수정 전 트리에서 반드시 실패해야 한다 (현행은 `unknown` — M-2 실측). 실패를 먼저 관측한 뒤 수정한다.

## AC-CL-009 — 오류 문구를 인증 성공으로 읽지 않는다 (REQ-CL-009) [핵심]

순수 파서 `parseCodexAuthLine(combined, exitCode)` 을 직접 시험한다. 각 행이 하나의 케이스다:

| 입력 (rc) | 기대 | 무엇을 막는가 |
|---|---|---|
| `Logged in using ChatGPT` (0) | `chatgpt` | — (긍정 기준선) |
| `Logged in using API key` (0) | `apiKey` | — (긍정 기준선) |
| `Logged in using ChatGPT` (1) | `chatgpt` | rc 만으로 버리지 않는다 |
| `error: API key missing` (1) | `unknown` | 부분 일치 |
| `provider configuration unreadable` (1) | `unknown` | 부분 일치 |
| `failed to reach ChatGPT backend` (1) | `unknown` | 부분 일치 |
| **`Logged in state unavailable: API key missing`** (1) | **`unknown`** | **앵커만 거는 안이 뚫리는 지점** — 이 행은 `logged in` 으로 시작한다 |
| `Logged in using ChatGPT (session expired)` (1) | `unknown` | 전체 행 문법 불일치 — 추측하지 않는다 |
| `Logged in using Acme SSO` (0) | `unknown` | 문법에 없는 새 provider 는 추측하지 않는다 |
| `warning: cache stale\nLogged in using ChatGPT` (0) | `chatgpt` | 잡음 행이 섞여도 긍정 행은 찾는다 |
| 빈 입력 (0) | `unknown` | — |
| **`Logged in using ChatGPT\nLogged in using API key`** (0) | **`unknown`** | 서로 다른 provider 를 말하는 일치 행이 둘 — 충돌은 판정하지 않는다 |
| `Logged in using ChatGPT\nLogged in using ChatGPT` (0) | `chatgpt` | 같은 provider 반복은 충돌이 아니다 |

**정규화 축 — 파생 케이스를 기계 생성한다.** 위 두 긍정 기준선 행 각각에 다음 변형의 **모든 조합** 을 적용해 케이스를 만든다: 후행 `\r` / 후행 공백 / 선행 공백 / 후행 탭 / provider 토큰의 대소문자 뒤섞기(`ChatGPT` → `chatGPT`, `API key` → `api KEY`).

- **Then** 생성된 모든 파생 케이스가 원본과 **같은 provider** 를 낸다. 고정 입력 11칸을 외운 룩업 표는 이 축에서 죽는다 — 현장 출력에는 CRLF 와 후행 공백이 흔하다.

**속성 단언 — 표를 외우는 구현을 원천 차단한다.**

- **Given** 고정 seed 로 1,000건의 임의 입력을 생성한다 (긍정 문구 조각·오류 문구 조각·공백·개행을 섞는다)
- **Then** 각 입력에 대해 `parseCodexAuthLine` 이 `unknown` 이 **아닌** 값을 내는 것과, 그 입력 안에 시험이 독립적으로 들고 있는 참조 문법 `(?i)^[ \t]*logged in using (chatgpt|api key)[ \t]*\r?$` 에 맞는 행이 **정확히 한 종류의 provider 로** 존재하는 것이 **동치** 다 (양방향). 참조 문법은 구현에서 import 하지 않고 시험이 따로 적는다 — 같은 상수를 공유하면 아무것도 검증하지 못한다.
- **근거**: 현행 분류기(`mcp_codex.go:1332-1338`)는 부분 일치라 4~6행을 각각 `apiKey` / `provider` / `chatgpt` 로 잘못 분류한다. 7행은 `^logged in` **앵커만** 거는 중간안도 뚫린다는 것을 고정한다.

## AC-CL-010 — 판정 불가는 조치와 함께 보고 (REQ-CL-009)

- **Given** 판정 불가 원인 **네 축** — (1) `auth.json` 부재 + 두 스트림 모두 비어 있음, (2) `auth.json` 부재 + 러너가 오류를 반환, (3) `auth.json` 부재 + 비영 rc + 문법 불일치 행, (4) `auth.json` 파싱 실패
- **When** 네 칸 각각에서 맨몸 `moai codex` 를 실행하면
- **Then** 네 칸 모두 auth 행 전문이 **같은 이름 붙은 상수와 정확 일치** 한다. 그 상수는 `unknown` 과 조치 `codex login status` 를 담는다. 원인마다 다른 문안을 내거나 러너 오류를 로그아웃으로 단정하는 분기는 여기서 갈린다 — 원인 축이 한 칸뿐이면 그 분기는 한 번도 실행되지 않는다.
- **And** 네 칸의 출력 전문에서 대소문자 무시 폐집합 `{logged out, logged-out, not logged in, no credentials, signed out}` 의 히트 합이 **0** 이다. 특정 한 문구의 부재만 보는 판정은 같은 뜻의 다른 문구를 놓친다.

## AC-CL-011 — codex 부재 시 exec 없음 (REQ-CL-012)

- **Given** `codexLookPath` 가 실패하도록 스텁된 상태
- **When** `cli` / `app` 을 실행하면
- **Then** 둘 다 비영 rc 로 종료하고, 기동 횟수는 0 이며(두 자리의 합), stderr 로 나간 진단의 **줄 수가 정확히 1** 이고 그 줄이 설치 조치 상수와 **정확 일치** 한다 (REQ 의 "single diagnostic" — 줄 수를 안 세면 세 줄짜리 진단도 통과한다).
- **And Given** 바이너리 부재 × 배선 2상태(`wired` / `not wired`) 교차
- **When** 맨몸 / `status` 를 실행하면 (총 4칸)
- **Then** rc 는 0 이고, 리드아웃의 행 라벨 집합이 폐집합 `{codex, home, auth, wiring, agents, harness}` 과 **같으며**, 바이너리 행은 `not found` 이고 배선 행은 그 칸 fixture 상태에 대한 AC-CL-004 기대 상수와 정확 일치한다. 행 라벨 집합을 등식으로 재는 것이 요점이다 — 바이너리 행 하나만 확인하면 나머지 다섯 행을 통째로 생략하는 조기 반환이 통과하고, §E fail-open("어느 프로브가 실패해도 나머지를 계속 보고")이 한 번도 관측되지 않는다.

## AC-CL-012 — 쓰기 없음 (REQ-CL-013)

범위: **모든 형태, 예외 없음.** 초기화는 `SPEC-CODEX-INIT-001` 로 분리됐으므로 런처에는 쓰기가 허용되는 경로가 없다 (REQ-CL-013).

- **Given** 격리된 임시 홈 — `HOME` · `XDG_CONFIG_HOME` · `XDG_CACHE_HOME` · `XDG_DATA_HOME` · `TMPDIR` 을 전부 그 아래로 돌리고, 임시 프로젝트 루트 · 임시 CODEX_HOME · 임시 Claude 프로필 디렉터리도 그 안에 둔다. 스냅샷 대상은 **격리 홈 트리 전체** 의 파일 목록·mtime 이다. 열거된 세 트리만 재면 `~/.moai/cache/` 같은 트리 밖 쓰기가 보이지 않는데, REQ 후단 "it does not write" 는 열거보다 넓다.
- **When** 네 형태(맨몸 / `status` / `cli` / `app`)를 (기동 스텁 상태로) 각각 실행한 뒤 다시 스냅샷하면
- **Then** 스냅샷이 동일하다.
- **And Given** `CODEX_HOME` 이 **가리키는 경로 자체가 없는** 칸을 추가해 같은 네 형태를 돌리면
- **Then** 실행 후에도 그 경로가 **여전히 존재하지 않고**(`os.Stat` 이 `IsNotExist`), 리드아웃은 그 상태를 보고한다. 존재하는 디렉터리를 전제한 스냅샷만으로는 없으면 만들어 버리는 구현(§C.3 이 금지)이 관측되지 않는다.

## AC-CL-013 — 중립성 (REQ-CL-014)

- **When** 템플릿 중립성 가드를 실행하면 (`MOAI_TEMPLATE_LEAK_STRICT=1 go test ./internal/template/...`)
- **Then** 통과한다.
- **And Given** 템플릿 가드는 `internal/cli` 를 보지 않으므로 도움말은 따로 판정한다
- **When** `codexCmd` **와 그 모든 하위 커맨드** 를 순회하며 리플렉션으로 각 `cobra.Command` 의 **모든 `string` · `[]string` 필드**(`Use` / `Short` / `Long` / `Example` / `Aliases` / `SuggestFor` 등)와 **모든 플래그의 usage 문자열**(로컬·영속 전부)을 모아 검사하면 — `Long` 과 예시만 보는 판정은 `Short` 와 플래그 usage 를 스캔 밖에 둔다
- **Then** 다음 패턴이 각각 0건이다: `SPEC-` · `REQ-` · 카드 id(`t` + 숫자) · ISO 날짜 · 7자리 이상 커밋 SHA · 절대 홈 경로(`/Users/` · `/home/`) · `CLAUDE.local` · `.moai/reports`. 목록의 출처는 `.moai/docs/template-internal-isolation-doctrine.md` §25.1 의 금지 클래스이며, 네 종만 세던 앞선 판정은 나머지를 통과시켰다.
- **And** 모아진 문자열 전체에서 **비-ASCII 문자 수가 0** 이다 — "language-neutral" 을 기계화한 형태다 (한국어 도움말이 여기서 걸린다).

## AC-CL-014 — 게이트와 크로스 플랫폼 (전 REQ)

- **When** `go build ./...` · `go vet ./...` · `GOOS=windows go vet ./...` · `golangci-lint run` 을 실행하면
- **Then** 전부 rc 0 이다.
- **And** `GOOS=windows go test -c ./internal/cli/` 가 rc 0 이다 — vet 보다 강하다 (시험 바이너리가 실제로 링크된다). 다만 이것도 **컴파일만** 증명한다.
- **And** 이 SPEC 이 추가한 파일에서 OS 빌드 태그(`//go:build` 줄 중 `windows` · `darwin` · `linux` · `unix` 토큰을 담은 것)가 **0건** 이고, `_windows.go` / `_unix.go` / `_darwin.go` 접미 파일도 0건이다. 빌드 태그로 Windows 판을 빈 스텁으로 갈아 끼우면 정적 게이트는 전부 초록이 되므로, 이 단언이 없으면 네 게이트는 Windows 동작에 대해 아무것도 말하지 않는다.
- **And** 이 SPEC 이 추가한 파일에서 **`syscall` 패키지의 import 가 0건** 이고, 프로세스 교체 계열 식별자(`syscall.Exec` · `unix.Exec` · `golang.org/x/sys/unix`)의 출현이 **0건** 이다. 위 태그 0건 단언이 성립하는 유일한 방법은 전 플랫폼 공통 API 로 기동하는 것이며, 이 단언이 그것을 못 박는다 — 교체를 쓰면서 태그를 안 갈면 Windows 컴파일이 깨지고, 태그를 갈면 앞 단언이 깨진다. 둘 다 피하려는 우회를 정면으로 막는다.
- **And** 이 SPEC 이 추가한 시험 중 `runtime.GOOS` 로 skip 하는 것은 AC-CL-008 의 fixture 실행 3칸 **뿐** 이다 (개수 3, 그 세 이름을 상수로 고정). 나머지는 전 플랫폼에서 실행된다.
- **And** 경로 조립은 `filepath.Join` 결과와의 비교로 판정한다 (AC-CL-005 의 끝-구분자 칸이 그 판정이며, `/` 를 하드코딩한 조립은 거기서 갈린다).
- **관측 범위 명시**: Windows **실행** 은 `release-pr-multi-os.yml` 의 windows 레그(`go test ./...`)가 이 시험들을 함께 돌 때 관측된다. 위 두 단언(빌드 태그 0건 · GOOS skip 3건)이 그 레그에 이 시험들이 실제로 실린다는 것을 보증한다. PR 단계 CI 는 ubuntu 만 돌므로 머지 전 Windows 실행 결과가 없으면 그것은 통과가 아니라 **Gap** 으로 적는다. PR 단계에 windows 잡을 새로 추가하는 것은 이 SPEC 범위 밖이다.

## AC-CL-015 — 공유 러너 무회귀 (REQ-CL-010)

- **When** 기존 codex 관련 시험을 실행하면 (`go test -json ./internal/cli/... -run Codex -timeout 600s`)
- **Then** 전부 통과하고, `-json` 출력에서 `"Action":"skip"` 인 시험 함수 수가 **0** 이다 (Windows 에서만 skip 되는 AC-CL-008 fixture 3칸은 제외하며, 그 세 이름은 상수로 고정). `t.Skip` 은 `go test` 요약에서 `ok` 로 보고되므로 "전부 통과" 만으로는 깨지는 시험에 skip 을 붙여 숨기는 수정을 못 잡는다.
- **And** 수정 전후로 `-run Codex` 가 잡는 **시험 함수 이름 목록** 을 비교해 삭제·개명이 0건이다 (새 이름 추가는 허용).
- **And** 순수 함수 세 개(`combineCodexStreams` / `parseCodexAuthLine` / `classifyCodexAuthFile`)는 시험에서 프로세스를 띄우지 않고 직접 호출되며, **각 호출의 반환값이 기대값과 비교** 된다. 케이스 수의 하한을 못 박는다: `classifyCodexAuthFile` ≥ AC-CL-008 1단 표의 행 수, `parseCodexAuthLine` ≥ AC-CL-009 표의 행 수 + 파생 케이스 수, `combineCodexStreams` ≥ 3 (AC-CL-008 결합 3칸). "직접 호출 ≥1건" 은 단언 없는 명목 호출로 충족되므로 쓰지 않는다.
- **And** `go test -coverprofile` + `go tool cover -func` 에서 이 세 함수의 구문 커버리지가 각각 **100%** 다 — 표에 없는 분기를 몰래 넣는 구현이 여기서 드러난다.
- **And** `--version` 조회 경로는 기존 `codexCommandRunner.run` 을 그대로 쓴다: 러너 스텁이 공급한 `SENTINEL-VER-9x9` 가 최종 리드아웃 출력에 **정확 일치** 로 나타난다. 호출 사실만 세면 반환값을 버리고 `unknown` 을 찍는 구현이 통과한다.

## AC-CL-016 — 데스크톱 앱 위임 (REQ-CL-011)

- **Given** 기동 seam 스텁
- **When** `moai codex app` 을 실행하면
- **Then** 포착 argv 가 정확히 `[codex, app]` 이다.
- **And Given** 기동 seam 이 **실패를 반환하는** 칸 (codex 가 앱을 못 찾아 비영 rc 로 끝나는 상황)
- **Then** 그 뒤로 추가 프로세스 기동이 **0회** 이고, moai 의 rc 가 seam 이 돌려준 rc 와 **같다**. 성공 경로만 돌면 macOS 폴백 분기(`open -b <bundle-id>` 류)는 한 번도 실행되지 않는다.
- **And** **폐집합 단언**: 이 SPEC 이 추가한 파일이 기동하는 실행 파일 이름의 집합이 `{codex}` 와 **같다**. 두 축으로 잰다 — (1) 전 시험 칸에서 포착 seam 이 기록한 `program` 의 basename 합집합, (2) 이 SPEC 이 추가한 파일의 **프로세스 기동 원시 호출 전부** — `exec.Command` · `exec.CommandContext` · `os.StartProcess` — 의 호출 지점에서 0번 인자로 넘어가는 값이 codex 경로 변수 하나뿐임 (한 원시만 열거하면 다른 원시로 새는 구현이 정적 축을 빠져나간다). 금지 패턴을 열거하는 grep(`/Applications` · `open -a` 따위)은 열거되지 않은 형태(`open -b` + 번들 ID)를 놓친다.
- **And Given** stdout 에 `install the desktop app from ...` 를 쓰는 fixture 실행 파일
- **Then** 런처를 지난 출력 바이트가 그 fixture 출력과 **동일** 하다 — 필터·치환·재해석 0건. codex 의 안내를 moai 문구로 갈아 끼우는 구현이 여기서 걸린다.
- **And** 이 칸은 **직접 기동 경로와 spawn 경로 양쪽에서 각각** 돈다. 프로세스를 교체하지 않으므로 두 경로 모두에서 출력이 관측 가능하며, 한쪽만 돌리면 다른 쪽에서 출력을 가공하는 구현이 통과한다.

---

## 판정 제외 (근거 명시)

- **실제 Codex 앱 기동**: CI 러너에 데스크톱 환경이 없다. 운영자 수동 확인 항목으로 `progress.md` 에 남긴다 — 확인 방법은 배선된 프로젝트에서 `moai codex app` 실행 후 앱 전면 등장 여부.
- **실 바이너리 auth 왕복**: 로그인 상태는 머신 상태에 의존한다. 기계 판정은 스텁으로 하고, 실 바이너리 확인은 `moai codex status` 출력 1회를 `progress.md` 에 붙이는 것으로 갈음한다 (§A.2 의 기준선 측정과 대칭).
- **대화형 tty 왕복**: CI 러너의 시험 프로세스는 tty 를 붙이고 돌지 않으므로, codex 가 실제로 tty 를 물려받아 대화형으로 동작하는지는 **관측할 수 없다.** 그래서 이 SPEC 은 tty 동작을 단언하지 않고, 그것이 기계적으로 의존하는 **선행 조건** 만 판정한다 — 기동 seam 이 부모 자신의 `os.Stdin` / `os.Stdout` / `os.Stderr` 값을 그대로 자식에게 넘기는지(AC-CL-002 의 stdio 항등 칸). 왕복 자체는 **명시적 Gap** 이며 run-phase 의 운영자 수동 측정 항목으로 `progress.md` 에 남긴다 — 배선된 프로젝트에서 실제 터미널의 `moai codex cli` 를 실행해 codex 의 대화형 프롬프트가 뜨고 입력이 먹는지 확인한다. **그 측정이 깨진 것으로 나오면 빌드 태그 문제가 운영자에게 되돌아간다** — 그때는 프로세스 교체를 되살릴지(플랫폼별 파일 + AC-CL-014 면제 조항) 다른 stdio 배선으로 고칠지를 운영자가 판정해야 하며, 이 SPEC 은 그 결정을 선점하지 않는다.
- **PR 단계 Windows 실행**: PR CI 는 ubuntu 만 돈다 (AC-CL-014 의 관측 범위 참조). Windows 실행 관측은 릴리스 PR 의 multi-OS 레그 소관이며, 그 결과가 없는 시점에는 통과가 아니라 Gap 이다. PR 단계에 windows 잡을 추가하는 것은 이 SPEC 범위 밖이다.
