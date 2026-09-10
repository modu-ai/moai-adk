# SPEC-CODEX-LAUNCHER-001 — AC 뮤턴트 분석 (t197)

대상: `.moai/specs/SPEC-CODEX-LAUNCHER-001/acceptance.md` AC-CL-001..016 **전수 16건**, 표본 없음.
방법: 각 AC 에 대해 "그 AC 가 **문자 그대로** 요구하는 모든 단언을 만족하면서, 그 AC 가 인용하는 REQ 를 위반하는" 구체적 구현(뮤턴트)을 실제로 작성 시도했다. 작성에 실패한 경우에만 MUTANT-FREE 로 적는다.
읽는 법: 각 절의 `강도` 는 뮤턴트가 실제 구현에서 나올 법한 정도다 — 강(자연스럽게 그렇게 짜게 됨) / 중(게으른 구현) / 약(작위적이지만 통과함).

---

## AC-CL-001
- **상태**: MUTANT-WRITABLE (강도: 중)
- **mutant**: `codexCmd` 를 `cc`/`glm`/`cg` 가 속한 기존 그룹 ID 가 아니라 **새 그룹**(`GroupID: "launch-codex"`, `Title: "LAUNCH COMMANDS"`)에 등록한다. cobra 는 그룹 제목의 중복을 검사하지 않으므로 `moai --help` 에 `LAUNCH COMMANDS` 블록이 **두 개** 렌더링되고, 두 번째 블록에 `codex` 한 줄만 들어간다. `moai codex --help` 는 정상 rc 0.
- **위반되는 REQ 조항**: "The system shall provide a top-level `moai codex` command registered in the `launch` command group, so it appears alongside `cc` / `glm` / `cg` in `moai --help`."
- **왜 AC 가 못 잡는가**: AC 는 "LAUNCH COMMANDS 그룹에 `codex` 행이 나타나고" 만 요구한다 — 도움말 전문에 대한 부분 문자열/블록 grep 으로 충족된다. `codex` 가 `cc`/`glm`/`cg` 와 **같은** 블록에 있는지, 블록이 하나인지는 어디에도 단언돼 있지 않다.
- **조이는 방법**: 도움말에서 `LAUNCH COMMANDS` 헤딩 출현 횟수 == 1 을 단언하고, 그 단일 블록 안에서 `cc`/`glm`/`cg`/`codex` **네 이름이 모두** 나오는지 확인한다. 더 강하게는 `codexCmd.GroupID == ccCmd.GroupID` 를 심볼로 직접 비교한다(문자열 리터럴 재기입이 아니라).

## AC-CL-002
- **상태**: MUTANT-WRITABLE (강도: 강)
- **mutant**: 동사 라우팅을 allowlist 가 아니라 **부정 분기**로 짠다. `switch args[0] { case "", "status": readout(); case "app": exec("codex","app"); default: exec(append([]string{"codex"}, args...)...) }` — 즉 `cli` 를 포함해 **인식하지 못한 모든 토큰**이 CLI 기동으로 떨어진다. cwd 는 `os.Getwd()` 를 그대로 쓴다.
  - 파생 뮤턴트(같은 절 안에서 별도로 성립): cwd 를 `os.Getwd()` 로 잡으므로 프로젝트 **하위 디렉터리**에서 호출하면 그 하위 디렉터리에서 codex 가 뜬다.
- **위반되는 REQ 조항**: "The bare `moai codex` command shall print the readiness readout and exec nothing; launching shall require an explicit verb — `cli` (Codex CLI in the current project directory) or `app` (desktop app)."
- **왜 AC 가 못 잡는가**: AC 가 도는 입력은 정확히 네 개(맨몸 / `status` / `cli` / `app`)다. 네 칸 모두 기대대로 동작하므로 exec 횟수 단언(0/0/1/1)·argv 단언·`--` 꼬리 단언이 전부 통과한다. "명시 동사를 **요구**한다" 는 부정 방향(오타 토큰 `moai codex --model o3`, `moai codex cl` 이 **기동해서는 안 된다**)은 한 칸도 돌지 않는다. cwd 단언도 "호출자의 프로젝트 루트" 인데 시험은 프로젝트 루트에서 호출하므로 `os.Getwd()` 뮤턴트와 구별되지 않는다 — 하위 디렉터리 축이 없다.
- **조이는 방법**: (1) 미지 토큰 행 추가 — `moai codex bogus` / `moai codex --model o3` 는 exec **0회** + 비영 rc + 사용법 진단. (2) cwd 축을 교차 — 프로젝트 루트의 **하위 디렉터리**에서 `cli` 를 호출하고, 포착 cwd 가 하위 디렉터리가 아니라 루트임을 단언한다(worktree 루트 판정과 함께 2칸).

## AC-CL-003
- **상태**: MUTANT-WRITABLE (강도: 강)
- **mutant**: `--spawn` 경로에서 `spawnLaunch(verb)` 만 호출하고 `--` 뒤 사용자 인자를 **버린다**(꼬리를 seam 에 넘기지 않음). tmux 부재 시에는 자체 문안 `codex: tmux not found` 로 비영 종료한다(`moai cc` 가 쓰는 공유 진단 상수를 쓰지 않는다).
- **위반되는 REQ 조항**: "Where the operator passes `--spawn`, the system shall run the launch in a new tmux window instead of replacing the current process, matching the `moai cc --spawn` contract, and shall fail with the same diagnostic when tmux is absent."
- **왜 AC 가 못 잡는가**: AC 는 spawn 인자에 대해 "그 인자에 각각 `cli` / `app` 이 보존" 만 요구한다 — 동사만 보고 인자 꼬리는 보지 않는다. `--` 꼬리 정확 일치 단언은 AC-CL-002 의 **exec 경로**에만 걸려 있어 spawn 경로와 교차되지 않는다(형태 × 인자 교차 누락). tmux 진단은 "동일 계열" 이라는 산문이라 `tmux` 단어만 들어가면 통과한다.
- **조이는 방법**: (1) `moai codex cli --spawn -- --model o3 "a b"` 를 돌려 spawn 에 포착된 인자 꼬리가 exec 경로와 **같은 토큰 열**임을 단언. (2) tmux 부재 진단이 `moai cc --spawn` 이 쓰는 것과 **같은 상수/같은 바이트**임을 문자열 동등으로 단언한다.

## AC-CL-004
- **상태**: MUTANT-WRITABLE (강도: 강)
- **mutant**: `partial` 문안을 상태와 무관하게 고정한다 — `hooks.json` 만 있든 `config.toml` 만 있든 항상 `partial (.codex/hooks.json missing)` 를 찍는다. auth 행은 프로브 결과와 무관하게 항상 `unknown` 을 찍는다.
- **위반되는 REQ 조항**: "The readiness readout shall report, as discrete rows: codex binary path, codex version, resolved `CODEX_HOME`, auth provider, and project wiring state (`.codex/hooks.json` + `.codex/config.toml` presence and whitelist validity)."
- **왜 AC 가 못 잡는가**: 표의 두 `partial` 행은 기대값을 "`partial`(어느 쪽이 없는지 명시)" 라는 **산문**으로만 적는다. 두 행 모두 `partial` 을 포함하고 파일 이름 하나를 명시하므로, "명시했는지" 를 존재 검사로 구현한 판정은 두 칸 다 통과한다 — **틀린 파일 이름**을 잡는 단언이 없다. auth 행은 이 AC 의 어떤 줄도 값을 요구하지 않는다(값 일치 단언은 바이너리 경로·버전 두 행에만 걸려 있다).
- **조이는 방법**: 두 `partial` 행을 서로에 대해 배타로 단언한다 — `config.toml` 만 있는 칸의 출력은 `hooks.json` 을 포함하고 `config.toml` 은 **포함하지 않는다**(반대 칸은 반대로). auth 행에도 스텁이 공급한 값(예: `stub-provider`)과의 문자열 일치 단언을 건다.

## AC-CL-005
- **상태**: MUTANT-WRITABLE (강도: 강)
- **mutant**: `v, ok := os.LookupEnv("CODEX_HOME"); if ok { return v, "env" }` — 즉 **설정됐지만 빈 문자열**(`CODEX_HOME=`)일 때 값 `""` / 출처 `env` 를 보고하고 폴백하지 않는다. 폴백 경로는 `filepath.Join(os.Getenv("HOME"), ".codex")` 로 짠다.
- **위반되는 REQ 조항**: "The system shall resolve `CODEX_HOME` from the `CODEX_HOME` environment variable, falling back to `~/.codex`, and shall report which of the two supplied the value."
- **왜 AC 가 못 잡는가**: AC 의 두 칸은 "설정됨(`/tmp/xyz`)" 과 "미설정" 뿐이다. **설정됐으나 빈 값**이라는 세 번째 상태가 없어, `LookupEnv` 의 `ok` 만 보는 구현과 값의 비어있음까지 보는 구현이 구별되지 않는다. 폴백 칸의 기대값 `<home>/.codex` 도 `HOME` 기반 구현과 `os.UserHomeDir` 기반 구현이 POSIX 시험 환경에서 같은 문자열을 내므로 갈리지 않는다(§E 크로스플랫폼 요구 미관측).
- **조이는 방법**: (1) `CODEX_HOME=""` 행 추가 — 기대는 폴백 + 출처 `default`. (2) 공백만 있는 값 행 추가. (3) `HOME` 을 지운 상태에서도 폴백이 홈을 해석하는지(= `os.UserHomeDir` 사용) 단언하거나 홈 해석 seam 을 스텁해 호출을 확인한다.

## AC-CL-006
- **상태**: MUTANT-WRITABLE (강도: 중)
- **mutant**: 불완전 배선을 **오류 스타일**로 보고한다 — `stderr` 에 `ERROR: .codex wiring broken (hooks.json missing)` 를 쓰고 이어서 `Run: moai init --agent codex` 를 쓴 뒤 rc 0 으로 종료한다.
- **위반되는 REQ 조항**: "Where the project's `.codex/` wiring is incomplete — the directory absent, present but empty, or missing either generated file — the readout shall report it as an informational state (not an error) and name `moai init --agent codex` as the action, mirroring the fail-open stance of the `moai doctor` Codex Wiring check."
- **왜 AC 가 못 잡는가**: 세 단언이 전부 (a) rc == 0, (b) 배선 행이 상태를 적음, (c) 출력에 `moai init --agent codex` 포함이다. "정보성이지 오류가 아니다" 가 rc 로만 대리 측정되고 있어, rc 0 을 유지하면서 오류 어휘·오류 스트림을 쓰는 구현이 그대로 통과한다.
- **조이는 방법**: 불완전 5상태 × 2형태 10칸 각각에서 (1) 배선 행이 `stdout` 으로 나오는지(스트림 분리 단언), (2) 출력에 `error` / `ERROR` / `failed` 어휘가 0건인지를 함께 단언한다.

## AC-CL-007
- **상태**: MUTANT-WRITABLE (강도: 강)
- **mutant**: 런처는 공유 프로브를 제대로 쓴다(그래서 sentinel 3종이 전부 나온다). 그러나 `internal/web/codex_state.go` 는 고치지 않고, 대신 그 파일에 **파일 기반 자체 분류기**를 새로 넣는다 — `<CODEX_HOME>/auth.json` 을 직접 열어 `auth_mode` 를 스위치한다. 이 코드에는 `login status` 문자열이 없고 프로세스도 띄우지 않는다.
- **위반되는 REQ 조항**: "The classification correction shall apply to every consumer of the shared probe (the `codex_setup` MCP tool, the web console Codex card, and this launcher), with no second classification path introduced and no change to the shared `codexCommandRunner` interface."
- **왜 AC 가 못 잡는가**: 중복 부재 판정이 **두 개의 텍스트 grep** 이다 — `grep "login status"` 히트가 한 곳뿐인지, `internal/web` 에 "auth 분류 로직이 여전히 0건" 인지. 앞의 grep 은 명령 프로브 문구를 쓰지 않는 파일 기반 사본을 보지 못하고, 뒤의 단언은 어떤 패턴으로 0건을 잰다는 정의가 없다(측정 불가능한 단언 → 실제 시험에서는 임의의 좁은 패턴이 된다). sentinel 전파 단언은 **런처 표면에만** 걸려 있어 web/MCP 소비자는 한 번도 돌지 않는다.
- **조이는 방법**: sentinel 전파를 소비자 축으로 교차한다 — 같은 sentinel 프로브 스텁 하나로 (a) 런처 리드아웃, (b) `codex_setup` MCP 도구 응답, (c) web 콘솔 Codex 카드 렌더 결과 **셋 모두**에서 `sentinel-provider` 가 나오는지 단언한다. 텍스트 grep 은 보조로만 남기고, "0건" 단언에는 스캔한 패턴과 경로를 명시한다.

## AC-CL-008
- **상태**: MUTANT-WRITABLE (강도: 강, 뮤턴트 5종)
- **mutant**:
  1. **apikey 공백 뮤턴트** — `nonEmptyString` 판정을 chatgpt 경로에만 `strings.TrimSpace` 로 걸고 `OPENAI_API_KEY` 는 `s != ""` 로 본다. `{"auth_mode":"apikey","OPENAI_API_KEY":"   "}` 가 `apiKey` 로 분류된다.
  2. **비밀 보존 뮤턴트** — `type tokenSet struct { credentialCount int; raw map[string]string }` 로 원문 토큰을 맵에 보관한다. 필드 타입이 `map[string]string` 이라 `string`/`[]byte`/`json.RawMessage` 어느 것도 아니다.
  3. **로깅 뮤턴트** — `readCodexAuthFile` 이 반환 오류에는 경로만 싣되, 직전에 `log.Printf("codex auth raw: %s", raw)` 로 본문을 로그에 남긴다.
  4. **결합 규칙 뮤턴트** — `combineCodexStreams(stdout, stderr []byte) []byte { return stderr }` — stdout 을 버린다.
  5. **1단 사문화 뮤턴트** — `classifyCodexAuthFile` 은 표대로 정확히 구현하되 `ProbeCodexSetup` 이 그것을 **호출하지 않고** 항상 명령 프로브로 간다(파일 판정 함수는 시험에서만 불리는 죽은 코드).
- **위반되는 REQ 조항**: "The auth classifier shall determine the provider from `<CODEX_HOME>/auth.json` only when the file's `auth_mode` is a known value AND the credential material that mode implies is present under a key that mode recognizes, as a non-empty JSON string ... as shall an empty or whitespace-only string and a container holding no such value. ... It shall deserialize through types that record only whether each credential field was non-empty, never the value itself, so no credential is retained, logged, or wrapped into an error."
- **왜 AC 가 못 잡는가**: (1) 26행 표에 `apikey` × **공백뿐** 칸이 없다 — 공백 행은 chatgpt 쪽에만 있다(모드 × 값-형태 교차 누락). (2) 리플렉션 단언이 **금지 타입 열거**(`string`/`[]byte`/`json.RawMessage`)라, 값 보존이 가능한 다른 종류(`map[string]string`, `[]string`, `fmt.Stringer` 구현체)는 전부 통과한다 — REQ 는 "값 자체를 절대 기록하지 않는다" 인데 AC 는 세 타입 이름만 센다. (3) 비밀 유출 판정이 **반환 오류의 `Error()` 전문**에만 걸려 있어 로그·stdout·패닉 메시지 채널이 비어 있다(REQ 는 `retained, logged, or wrapped` 셋을 요구). (4) 결합 규칙 fixture 가 `stdout 0바이트 / stderr 1줄` **한 축**뿐이라 stdout 을 버리는 구현과 둘을 합치는 구현이 구별되지 않는다 — 이 SPEC 이 고치려는 결함의 정확한 거울상이다. (5) 통합 1건의 given 이 "`auth.json` 부재" 라 1단을 통과시키지 않는다 — 1단이 조립부에 실제로 배선됐는지 어느 칸도 관측하지 않는다.
- **조이는 방법**: (1) 표에 `{"auth_mode":"apikey","OPENAI_API_KEY":"   "}` → `("",false)` 행과 `auth_mode` 타입 오염 행(`123` / `["chatgpt"]` / `"CHATGPT"`) 추가. (2) 리플렉션 단언을 **화이트리스트**로 뒤집는다 — 허용 필드는 `auth_mode` 를 받는 `string` 하나와 `bool`/`int` 뿐이고 그 외 어떤 kind 도 0건. (3) 비밀 sentinel 을 오류 전문뿐 아니라 **캡처한 stdout+stderr+로그 싱크 전체**에서 0건으로 단언. (4) fixture 실행 파일을 3종(stderr-only / stdout-only / 양쪽)으로 늘리고 세 칸 모두에서 결합 결과가 그 줄을 담는지 단언. (5) 통합 칸을 2개로 — `auth.json` **존재+유효** 일 때 명령 러너 스텁이 **0회** 호출되는지(1단 우선), 부재일 때 1회 호출되는지.

## AC-CL-009
- **상태**: MUTANT-WRITABLE (강도: 중)
- **mutant**: 문법 정규식 대신 **표 자체를 하드코딩한 룩업**으로 구현한다 — 각 줄을 소문자화한 뒤 `map[string]string{"logged in using chatgpt":"chatgpt","logged in using api key":"apiKey"}` 에 **완전 일치**로 조회. 표의 11칸을 전부 맞힌다. 그러나 실제 codex 출력에 흔한 `Logged in using ChatGPT\r`(CRLF), 후행 공백, 토큰 뒤 탭 등은 전부 `unknown` 으로 떨어진다.
- **위반되는 REQ 조항**: "When classifying from command output, the system shall accept a provider only from a status line matching a fixed whole-line grammar, mapping the captured provider term and nothing else"
- **왜 AC 가 못 잡는가**: 11행이 전부 **고정 입력 → 고정 출력** 예시라서, 문법을 구현한 것과 예시를 외운 것을 구별하지 못한다. 정규화 축(CRLF / 후행 공백 / 탭 / `api KEY` 같은 토큰 대소문자 혼합)이 한 칸도 없고, 서로 다른 provider 를 말하는 일치 행이 둘 이상일 때의 기대값도 없다.
- **조이는 방법**: (1) 긍정 두 행에 `\r`·후행 공백·선행 공백·토큰 대소문자 변형을 붙인 파생 케이스를 기계 생성해 **여전히 같은 provider** 를 내는지 단언. (2) 속성 단언 — 임의 입력에 대해 결과가 비어 있지 않으려면 시험이 독립적으로 가진 참조 문법 `(?i)^logged in using (chatgpt|api key)$` 에 맞는 행이 반드시 존재해야 한다. (3) 충돌 행(두 provider 동시 등장)의 기대값을 못 박는다.

## AC-CL-010
- **상태**: MUTANT-WRITABLE (강도: 중)
- **mutant**: auth 행을 `auth  unknown  (no credentials found — run: codex login status)` 로 찍고, 러너가 **오류를 반환**하는 경우(`err != nil`)에는 `auth  logged out` 으로 단정한다.
- **위반되는 REQ 조항**: "... and shall otherwise report `unknown` together with the action `codex login status` — an unreadable probe is a gap, never a verdict."
- **왜 AC 가 못 잡는가**: "로그아웃 단정 문구는 없다" 는 특정 문자열의 부재 검사로 구현될 수밖에 없는데 `no credentials found` 는 실질적으로 로그아웃 단정이면서 그 문자열이 아니다. 더 결정적으로 given 이 "`auth.json` 부재 + 두 스트림 모두 비어 있는 스텁" **한 칸**뿐이라, 러너 오류·비영 rc·파싱 실패 같은 다른 판정 불가 경로는 한 번도 돌지 않는다 — `logged out` 단정 분기가 미도달이다.
- **조이는 방법**: (1) auth 행 전문을 고정 상수와 **문자열 동등** 비교(부재 grep 대신 동등). (2) 판정 불가 원인을 축으로 교차 — 스트림 공백 / 러너 오류 / 비영 rc + 문법 불일치 / `auth.json` 파싱 실패 4칸에서 같은 행·같은 조치 문구가 나오는지.

## AC-CL-011
- **상태**: MUTANT-WRITABLE (강도: 강)
- **mutant**: 바이너리 미해결 시 리드아웃을 **조기 반환**한다 — 바이너리 행 `not found` 한 줄만 찍고 rc 0 으로 끝낸다(home / auth / wiring 행을 만들지 않는다). launch 동사에서는 진단을 **세 줄**(경로 탐색 실패 / 설치 안내 / 사용법)로 찍고 비영 종료한다.
- **위반되는 REQ 조항**: "Where the codex binary is absent from PATH, the launch verbs (`cli`, `app`) shall fail with a single diagnostic naming the install action and shall exec nothing; the readout form shall still succeed, reporting the binary row as not found — a diagnostic that refuses to run when the thing it diagnoses is missing is useless exactly when it is needed."
- **왜 AC 가 못 잡는가**: 리드아웃 쪽 단언이 "rc 0 + 바이너리 행 `not found`" 두 개뿐이라 나머지 네 행이 사라져도 통과한다 — 바이너리 부재 × 배선 상태 교차가 없어 §E fail-open("어느 프로브가 실패해도 나머지를 계속 보고")이 관측되지 않고, REQ 가 말하는 "필요할 때 정확히 쓸모없어지는 진단" 이 그대로 재현된다. 진단 쪽은 "설치 조치를 명명한다" 만 요구하므로 **단일** 진단 요구는 세지 않는다.
- **조이는 방법**: (1) 바이너리 부재 상태에서 리드아웃이 여전히 **6행 전부**를 내고 wiring 행이 fixture 상태(예: `wired`)를 정확히 보고하는지 단언(바이너리 부재 × 배선 2상태 교차). (2) 진단 줄 수 == 1 을 단언한다.

## AC-CL-012
- **상태**: MUTANT-WRITABLE (강도: 강, 뮤턴트 2종)
- **mutant**:
  1. **캐시 뮤턴트** — 리드아웃 결과를 `~/.moai/cache/codex-readiness.json` 에 캐시로 쓴다(스냅샷되는 세 트리 **밖**).
  2. **디렉터리 생성 뮤턴트** — `CODEX_HOME` 이 존재하지 않을 때 `os.MkdirAll(codexHome, 0o700)` 로 만들고 `missing` 대신 `empty` 로 보고한다.
- **위반되는 REQ 조항**: "The system shall not mutate `.claude/settings.local.json`, Claude profile state, or any file under `CODEX_HOME` on any verb — the launcher reads state and execs; it does not write."
- **왜 AC 가 못 잡는가**: (1) 무쓰기 판정 범위가 **열거된 세 트리**(임시 프로젝트 루트 / 임시 CODEX_HOME / 임시 Claude 프로필)로 한정돼 있다. `HOME` 을 격리하지 않으면 그 밖 어디에 쓰든 세 스냅샷이 동일하다 — REQ 후단 "it does not write" 는 열거된 세 대상보다 넓은데 판정은 열거뿐이다. (2) `CODEX_HOME` 스냅샷은 **존재하는** 임시 디렉터리를 전제로 잡으므로 "CODEX_HOME 자체가 없는" 상태가 fixture 축에 없다 — 디렉터리를 만들어 버리는 구현이 관측되지 않는다(§C.3 이 명시적으로 금지한 동작).
- **조이는 방법**: (1) 격리된 `HOME`/`XDG_*` 아래 전체를 스냅샷 대상에 넣거나, 읽기 전용 임시 홈으로 프로세스 전체의 쓰기를 막아 판정한다. (2) fixture 에 "CODEX_HOME 경로 부재" 칸을 추가하고 실행 후에도 그 경로가 **여전히 존재하지 않음**을 단언한다.

## AC-CL-013
- **상태**: MUTANT-WRITABLE (강도: 강)
- **mutant**: `codexCmd.Long` 을 한국어로 쓰고 예시에 `~/MoAI/moai-adk-go` 경로와 `자세한 근거는 REQ-CL-011 참조` 를 넣는다. `Short` 필드와 `--spawn` 플래그 usage 문자열에는 `t197 런처` 를 남긴다.
- **위반되는 REQ 조항**: "Help text, examples, and any template-side documentation shall stay language-neutral and free of internal identifiers, satisfying the template neutrality guard."
- **왜 AC 가 못 잡는가**: 판정 대상이 "`codexCmd` 의 생성된 도움말 문자열(`Long` + 예시)" 로 못 박혀 있어 `Short`·플래그 usage·하위 커맨드 도움말은 스캔 밖이다. 금지 패턴도 네 종(`SPEC-` / `t197` / 내부 날짜 / 커밋 SHA) 열거라 `REQ-` 토큰, macOS 편향 절대 경로, `CLAUDE.local` 참조는 걸리지 않는다(§25.1 C 클래스 중 일부만 열거). "language-neutral" 은 어떤 단언에도 대응되지 않는다.
- **조이는 방법**: (1) 스캔 범위를 `codexCmd` 와 그 하위 커맨드의 **모든** 문자열 필드(`Use`/`Short`/`Long`/`Example`/모든 플래그 usage)로 확장. (2) 금지 패턴을 `.moai/docs/template-internal-isolation-doctrine.md §25.1` 의 C 클래스에서 가져와 `REQ-`·macOS 홈 경로·`CLAUDE.local` 포함. (3) 비-ASCII 문자 0건(또는 로케일 중립 검사)으로 language-neutral 을 기계화.

## AC-CL-014
- **상태**: MUTANT-WRITABLE (강도: 강)
- **mutant**: `codex_readiness.go` / `codex_launcher.go` 상단에 `//go:build !windows` 를 달고, Windows 용으로는 "unsupported" 를 찍고 rc 0 으로 끝내는 스텁 파일을 둔다. 경로 조립은 `home + "/.codex"` 처럼 `/` 를 하드코딩한다.
- **위반되는 REQ 조항**: (AC 가 "전 REQ" 를 인용하므로 대표로) "The readiness readout shall report, as discrete rows: codex binary path, codex version, resolved `CODEX_HOME`, auth provider, and project wiring state ..." — Windows 에서는 어느 행도 보고되지 않는다. 함께: §E "크로스 플랫폼: 경로 해석은 `os.UserHomeDir` 기반이며 macOS 편향 경로를 하드코딩하지 않는다."
- **왜 AC 가 못 잡는가**: 네 게이트가 전부 **컴파일/정적 분석**이다. `GOOS=windows go vet` 은 빌드 태그로 배제된 파일을 보지 않으며 스텁이 타입 검사만 통과하면 초록이다("vet 은 컴파일만 증명한다"). 어떤 게이트도 Windows 에서의 **동작**을 관측하지 않고, `/` 하드코딩은 vet 대상이 아니다.
- **조이는 방법**: (1) 이 SPEC 이 추가한 파일에 OS 빌드 태그가 0건임을 단언. (2) 경로 조립 단위 시험을 `filepath` 결과와 비교해 전 플랫폼에서 돌리고, CI 의 windows 잡에서 `go test ./internal/cli/... -run Codex` 를 **실행**한다(vet 이 아니라 test).

## AC-CL-015
- **상태**: MUTANT-WRITABLE (강도: 중, 뮤턴트 2종)
- **mutant**:
  1. **회귀 은폐 뮤턴트** — 새 분류기 때문에 깨지는 기존 시험 `TestCodexAuthClassification` 첫 줄에 `t.Skip("superseded by two-stage ladder")` 를 넣는다. `-run Codex` 는 여전히 "전부 통과" 한다.
  2. **명목 호출 뮤턴트** — 시험 파일에 `_ = combineCodexStreams(nil, nil)` / `_, _ = parseCodexAuthLine(nil, 0)` / `_, _, _ = classifyCodexAuthFile(nil)` 를 한 줄씩 넣어 "직접 호출 ≥1건" 을 충족시킨다(단언 없음). 버전 조회는 `codexCommandRunner.run` 을 부르되 반환값을 **버리고** `"unknown"` 을 찍는다.
- **위반되는 REQ 조항**: "The classification correction shall apply to every consumer of the shared probe ... with no second classification path introduced and no change to the shared `codexCommandRunner` interface." (무회귀 축) / REQ-CL-004 의 "codex version" 행 (버전 축)
- **왜 AC 가 못 잡는가**: "전부 통과한다" 는 skip 을 실패로 세지 않는다(go test 는 skip 을 ok 로 보고한다). "각각 ≥1건의 직접 호출이 존재한다" 는 **호출 존재**만 요구하고 결과에 대한 단언을 요구하지 않는다 — "호출은 세지만 결과가 쓰였는지는 안 본다" 는 알려진 실패 형태 그대로다. `--version` 축도 "호출 seam 으로 확인" 이라 호출 사실만 세고, 그 값이 리드아웃에 실렸는지는 이 AC 에 없다(AC-CL-004 의 버전 값 일치는 스텁을 어느 층에 두느냐에 따라 이 뮤턴트를 통과시킨다).
- **조이는 방법**: (1) `go test -json` 으로 skip 개수 == 0 을 단언하고, 기존 Codex 시험 함수 이름 목록을 diff 전후로 비교해 삭제·개명 0건 확인. (2) "직접 호출" 대신 각 순수 함수가 **기대값과 비교되는 단언**에 최소 N개 케이스로 등장함을 요구(AC-CL-008/009 표를 근거로 지목). (3) 버전 축은 러너 스텁이 공급한 sentinel 버전 문자열이 최종 출력에 나타나는지로 교체.

## AC-CL-016
- **상태**: MUTANT-WRITABLE (강도: 강)
- **mutant**: 먼저 `codex app` 을 exec seam 으로 시도하고, seam 이 **실패를 반환하면** macOS 폴백으로 `exec.Command("/usr/bin/open", "-b", "com.openai.codex")` 를 부른다. 또한 codex 의 stdout 을 캡처해 `install` 이 포함된 줄을 걸러내고 moai 자체 안내 문구로 치환한다.
- **위반되는 REQ 조항**: "The `app` verb shall delegate to `codex app` rather than reimplementing desktop-app discovery or installation."
- **왜 AC 가 못 잡는가**: (1) 시험이 **성공 경로**만 돈다 — exec 스텁이 성공하므로 폴백 분기는 실행되지 않고 포착 argv 는 정확히 `[codex, app]` 이다. (2) 금지 grep 이 `/Applications` · `open -a` · 설치 관리자 호출이라는 **열거**라, `open -b` + 번들 ID 형태는 세 패턴 어디에도 걸리지 않는다. (3) "codex 의 출력을 그대로 통과시키고 재해석하지 않는다" 에 대응하는 fixture 가 없다 — 프로세스를 교체하는 exec 스텁에서는 출력 경로 자체가 관측되지 않는다.
- **조이는 방법**: (1) exec seam 이 **실패를 반환하는** 칸을 추가하고, 그 칸에서 추가 프로세스 기동이 0회이며 rc 가 codex 의 것을 그대로 전파하는지 단언(성공 경로 단독 금지). (2) 금지 판정을 열거 grep 대신 "이 SPEC 이 추가한 파일에서 `os/exec` 로 기동되는 실행 파일 이름 집합 == {codex}" 라는 폐집합 단언으로 교체. (3) codex 출력에 `install` 문구를 담은 fixture 를 넣고 런처 출력이 그 바이트와 **동일**함을 단언(필터·치환 0건).

---

## 요약

| AC | 상태 | 한 줄 요약 |
|---|---|---|
| AC-CL-001 | MUTANT-WRITABLE | 같은 제목의 **별도 그룹**에 등록해도 "LAUNCH COMMANDS 에 codex 행" 은 통과 — 같은 블록·단일 블록 단언 없음 |
| AC-CL-002 | MUTANT-WRITABLE | 미지 토큰을 전부 `cli` 로 떨어뜨리는 default 분기 + cwd `os.Getwd()` — 네 칸만 도는 표가 둘 다 못 잡음 |
| AC-CL-003 | MUTANT-WRITABLE | spawn 경로에서 `--` 꼬리를 버려도 통과(꼬리 단언이 exec 경로에만); tmux 진단은 "동일 계열" 산문 |
| AC-CL-004 | MUTANT-WRITABLE | 두 `partial` 칸이 서로 배타로 단언되지 않아 **틀린 파일 이름 고정 출력**이 통과; auth 행 값 단언 없음 |
| AC-CL-005 | MUTANT-WRITABLE | `CODEX_HOME=""`(설정+빈값) 칸이 없어 `LookupEnv` 의 ok 만 보는 구현이 생존 |
| AC-CL-006 | MUTANT-WRITABLE | rc 0 만 재므로 오류 어휘·stderr 로 보고하는 구현이 "정보성" 요구를 우회 |
| AC-CL-007 | MUTANT-WRITABLE | sentinel 전파가 런처에만 걸려 web/MCP 소비자 미관측 — `login status` 를 안 쓰는 파일 기반 사본이 통과 |
| AC-CL-008 | MUTANT-WRITABLE | apikey×공백 칸 부재, 금지-타입 열거식 리플렉션(map 통과), 오류 전문만 보는 비밀 검사, stderr-only 단일 fixture, 1단 미배선 미관측 |
| AC-CL-009 | MUTANT-WRITABLE | 11행이 고정 예시라 **표를 외운 룩업**이 통과 — CRLF·공백 정규화 축 없음 |
| AC-CL-010 | MUTANT-WRITABLE | 특정 문구 부재만 검사해 `no credentials found` 형 단정이 통과; 판정 불가 원인 축이 1칸 |
| AC-CL-011 | MUTANT-WRITABLE | 바이너리 부재 시 나머지 행을 통째로 생략해도 통과(fail-open 미관측), 진단 줄 수 미측정 |
| AC-CL-012 | MUTANT-WRITABLE | 스냅샷 범위가 세 트리 열거라 `~/.moai/cache` 쓰기가 안 보임; CODEX_HOME 부재 칸이 없어 디렉터리 생성 미관측 |
| AC-CL-013 | MUTANT-WRITABLE | `Long`+예시만 스캔, 금지 패턴 4종 열거 — `Short`·플래그 usage·`REQ-`·macOS 경로·한국어 도움말이 통과 |
| AC-CL-014 | MUTANT-WRITABLE | 네 게이트가 전부 정적 — `//go:build !windows` 스텁으로 Windows 기능을 비워도 초록 |
| AC-CL-015 | MUTANT-WRITABLE | `t.Skip` 이 "전부 통과" 로 계수됨; "직접 호출 ≥1" 은 단언 없는 명목 호출로 충족 |
| AC-CL-016 | MUTANT-WRITABLE | 성공 경로만 돌아 `open -b` 폴백·출력 재해석이 미관측; 금지 grep 이 열거식 |

MUTANT-FREE 0 / 16 — 16개 AC 전부에 대해 구체적 뮤턴트를 실제로 작성했다.
