# SPEC-CODEX-LAUNCHER-001 — 인수 기준

모든 기준은 기계 판정이다. codex 실 바이너리에 의존하는 항목은 없다 — 프로브 seam (`codexLookPath`, `codexCommandRunner`) 을 스텁해 판정한다.

## AC-CL-001 — 커맨드 등록 (REQ-CL-001)

- **Given** 빌드된 moai 바이너리
- **When** `moai --help` 를 실행하면
- **Then** LAUNCH COMMANDS 그룹에 `codex` 행이 나타나고, `moai codex --help` 가 rc 0 으로 도움말을 출력한다.

## AC-CL-002 — 동사 라우팅 (REQ-CL-002)

exec seam 은 호출 횟수뿐 아니라 **argv 전체와 cwd** 를 포착한다 — 명령 이름만 맞고 인자나 실행 위치가 틀린 구현이 통과하지 못하게.

- **When** 맨몸 / `status` / `cli` / `app` 각각을 호출하면
- **Then** 맨몸과 `status` 는 exec **0회**, `cli` 와 `app` 은 각 1회다.
- **And** `cli` 의 포착된 cwd 는 **호출자의 프로젝트 루트** 와 일치한다 (worktree 에서 부르면 그 worktree — 다른 트리에서 codex 가 뜨면 세션이 엉뚱한 코드를 본다).
- **And** `app` 의 포착 argv 는 정확히 `[codex, app]` 이다 — 앱 경로 탐색·설치 시도 코드는 실행되지 않는다 (`/Applications` 류 하드코딩 grep 0건).
- **And Given** `moai codex cli -- --model o3 "a b" '$x' --flag=v` 처럼 `--` 뒤에 공백·인용·`$`·`=` 를 포함한 인자를 주면
- **Then** 포착 argv 의 `codex` 뒤 꼬리가 이 토큰들과 **정확히 일치** 한다 (개수·순서·문자 모두; 셸 재해석이나 재인용으로 변형되지 않는다).
- **근거**: 맨몸의 의미는 리드 판정으로 (b) "리드아웃 + 명시 기동" 으로 확정됐다 (plan §B).

## AC-CL-003 — `--spawn` 패리티 (REQ-CL-003)

- **Given** tmux 존재를 흉내 낸 스텁 spawn seam
- **When** `moai codex cli --spawn` 과 `moai codex app --spawn` 을 각각 실행하면
- **Then** 둘 다 `spawnLaunch` 를 정확히 1회 호출하고, 그 인자에 각각 `cli` / `app` 이 보존되며, 현재 프로세스는 exec 으로 교체되지 않는다 (성공 경로 — 실패만 시험하면 아무것도 못 띄우는 구현이 통과한다).
- **And Given** tmux 부재를 흉내 낸 환경
- **Then** 둘 다 `moai cc --spawn` 과 동일 계열의 진단으로 실패하고 exec 은 0회다.
- **And** 맨몸 / `status` 에 `--spawn` 을 주면 거부한다 — 리드아웃은 새 창에서 띄울 대상이 아니다.

## AC-CL-004 — 리드아웃이 실제 값을 보고한다 (REQ-CL-004)

라벨 존재가 아니라 **값** 을 판정한다. 다음 상태 행렬을 스텁으로 구성하고 각 칸의 보고값을 확인한다:

| 배선 상태 (fixture) | 기대 보고 |
|---|---|
| `.codex/` 없음 | `not wired` + 조치 문구 |
| **`.codex/` 는 있고 비어 있음** | **`not wired`** + 조치 문구 — 디렉터리 존재만 보는 구현이 떨어지는 행 |
| `hooks.json` 만 있음 | `partial`(어느 쪽이 없는지 명시) + 조치 문구 |
| `config.toml` 만 있음 | `partial`(어느 쪽이 없는지 명시) + 조치 문구 |
| 둘 다 정상 | `wired` — 조치 문구 **없음** |
| 둘 다 있고 `hooks.json` 이 화이트리스트 위반 키 보유 | `invalid`(건강 상태로 보고하지 않는다) + 조치 문구 |

**조치 문구** 는 `moai init --agent codex` 이며, REQ-CL-006 이 요구하는 대로 **불완전한 다섯 상태 전부** 에 나타나야 한다 — 부재 상태에만 붙이는 구현은 빈·부분·불량 배선에서 사용자에게 다음 행동을 알려주지 않는다.

- **And** 맨몸과 `status` 두 형태가 여섯 행에서 **같은 값** 을 보고한다 (형태에 따라 판정이 갈리지 않는다).
- **And** 이 판정은 `SPEC-CODEX-INIT-001` 이 소비하는 단일 정의다 — 그 SPEC 은 자체 파일 검사를 하지 않는다 (AC-CI-002).

- **And** 바이너리 경로 · 버전 행은 스텁이 공급한 값과 문자열 일치한다 (라벨만 있고 값이 비어도 통과하는 grep 판정을 쓰지 않는다).
- **And** 화이트리스트 판정은 `codexadapter.ValidateConfig` 를 호출해 얻는다 (AC-CL-007 이 재구현 부재를 강제).

## AC-CL-005 — CODEX_HOME 해석과 출처 표시 (REQ-CL-005)

- **Given** `CODEX_HOME=/tmp/xyz` 가 설정된 프로세스
- **When** 리드아웃을 조립하면
- **Then** 값은 `/tmp/xyz`, 출처는 `env` 로 보고된다.
- **And Given** `CODEX_HOME` 미설정
- **Then** 값은 `<home>/.codex`, 출처는 `default` 로 보고된다.

## AC-CL-006 — 배선 없는 프로젝트는 정보성 (REQ-CL-006)

AC-CL-004 의 여섯 상태 중 **불완전한 다섯 상태 전부** 를, **맨몸과 `status` 두 형태 각각** 으로 돈다 (10칸).

- **When** 각 칸을 실행하면
- **Then** rc 는 0 이고 (미배선은 오류가 아니다), 배선 행이 그 상태를 적으며, 출력에 `moai init --agent codex` 문자열이 포함된다.
- **And** 배선 완료 상태에서는 그 문자열이 **나타나지 않는다** — 조치가 필요 없는데 조치를 권하지 않는다.
- **And** 열 칸의 rc 가 모두 0 이다 — 어느 불완전 상태도 리드아웃을 실패시키지 않는다 (REQ-CL-006 fail-open).

## AC-CL-007 — 분류 구현 단일성 (REQ-CL-007, REQ-CL-010)

텍스트 중복 부재만으로는 "공유 프로브를 실제로 쓴다" 를 증명하지 못한다. 그렇다고 호출 횟수 `≥1` 로도 부족하다 — "불렀다" 만 증명하고 "그 값을 썼다" 는 증명하지 못한다 — 공유 함수를 부른 뒤 결과를 버리고 자체 분류를 하는 구현도 통과한다. **sentinel 전파** 로 판정한다:

- **Given** 공유 프로브가 실제 값과 구별되는 sentinel 을 돌려주도록 스텁 (버전 `SENTINEL-VER-9x9`, 바이너리 경로 `/sentinel/path/codex`, auth `sentinel-provider`)
- **When** 리드아웃을 1회 조립하면
- **Then** 세 sentinel 값이 **모두** 최종 출력 문자열에 그대로 나타난다 (하나라도 빠지면 그 항목은 공유 프로브가 아닌 다른 경로에서 왔다는 뜻).
- **And Given** `codexadapter.ValidateConfig` 가 sentinel 위반을 돌려주도록 스텁
- **Then** 배선 행이 그 위반을 반영한다 (호출만 하고 자체 검증 결과를 쓰지 않았다면 실패한다).
- **And** 두 함수의 호출 횟수는 각각 ≥1 이다.
- **And** `grep -rn "login status" internal/ --include="*.go" | grep -v _test` 의 히트는 공유 분류기 한 곳뿐이다.
- **And** `internal/web` 에는 auth 분류 로직이 여전히 0건이다.
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
| `{"auth_mode":"totally-new-mode","tokens":{"access_token":"x"}}` | `("", false)` — 추측하지 않는다 |
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

굵은 세 행이 세 층위를 가른다:

1. `{ }` — 원문 바이트를 문자열과 비교하는 구현이 뚫린다.
2. `false` — JSON 타입을 안 보는 구현이 뚫린다.
3. `{"irrelevant":"x"}` — **타입은 보는데 키를 안 보는** 구현이 뚫린다. 값이 비지 않은 문자열이기만 하면 세어버리므로, 무관한 메타데이터가 ChatGPT 인증으로 통과한다.

세 층위를 동시에 만족시키는 규칙은 하나뿐이다: **자격 재료는 인정된 키 집합에 속하는, 비어 있지 않은 JSON 문자열이어야 한다.** ChatGPT 모드가 인정하는 키는 로그인 자격을 실제로 담는 것들(`access_token` / `id_token` / `refresh_token`)이고, `account_id` 같은 계정 메타데이터는 자격 재료가 아니다.

**비밀값 규율 — 값을 보존하지 않는 타입으로 판정한다.**

- **When** 역직렬화 대상 타입 집합(`codexAuthFile` 및 그 중첩 타입)의 필드를 리플렉션으로 열거하면
- **Then** `auth_mode` 를 받는 필드 하나를 제외하고, `string` · `[]byte` · `json.RawMessage` 타입의 필드가 **하나도 없다** — 자격 재료는 값이 아니라 "비어 있지 않음" 이라는 bool 로만 남는다.
- **And Given** 가짜 토큰 문자열 `SENTINEL-TOKEN-9x9` 를 심은 fixture 를 `readCodexAuthFile` (경로를 아는 유일한 층) 에 먹이고, 그 파일을 파싱 실패하도록 만들면
- **When** 반환된 오류의 `Error()` 전문을 검사하면
- **Then** `SENTINEL-TOKEN-9x9` 가 포함되지 않는다 (오류는 경로와 사유만 싣는다).

**2단 — 순수 파서 `parseCodexAuthLine(combined, exitCode)` 을 직접 시험한다** (프로세스 없음). 표는 AC-CL-009.

**결합 규칙 — 프로덕션 경로를 실제로 시험한다** (이 SPEC 이 고치려는 결함이 여기 있으므로 스텁으로 대신하지 않는다):

- **Given** `testdata/` 에 커밋된 fixture 실행 파일 — stdout 은 비우고 stderr 로만 `Logged in using ChatGPT` 를 쓴다 (셸 경유 없이 `exec` 로 직접 실행)
- **When** `defaultLoginStatusRunner` 를 그 파일에 대고 호출하면
- **Then** 반환된 `stdout` 은 0바이트, `stderr` 는 그 한 줄이다 — 두 스트림이 분리 수집된다.
- **And** `combineCodexStreams(stdout, stderr)` 의 결과가 그 줄을 담는다.
- **플랫폼**: 이 절만 Windows 에서 skip. 순수 함수 시험은 전 플랫폼.

**통합 1건** — 사다리가 실제로 이어지는지:

- **Given** `auth.json` 부재 + `codexLoginStatusRunner` 가 `stdout=[]`, `stderr=[Logged in using ChatGPT]`, `exitCode=0` 을 돌려주도록 스텁
- **When** `ProbeCodexSetup` 을 호출하면
- **Then** `AuthProvider == "chatgpt"` 이다.
- **기준선 근거**: 이 절은 수정 전 트리에서 반드시 실패해야 한다 (현행은 `unknown` — M-2 실측). 실패를 먼저 관측한 뒤 수정한다.

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

- **근거**: 현행 분류기(`mcp_codex.go:1332-1338`)는 부분 일치라 4~6행을 각각 `apiKey` / `provider` / `chatgpt` 로 잘못 분류한다. 7행은 `^logged in` **앵커만** 거는 중간안도 뚫린다는 것을 고정한다 — 전체 행 문법 + 캡처값 매핑만이 11칸을 동시에 만족시킨다.

## AC-CL-010 — 판정 불가는 조치와 함께 보고 (REQ-CL-009)

- **Given** `auth.json` 부재 + 두 스트림 모두 비어 있는 스텁
- **When** 맨몸 `moai codex` 를 실행하면
- **Then** auth 행은 `unknown` 이고, 출력에 `codex login status` 문자열이 포함되며, 로그아웃 단정 문구는 없다.

## AC-CL-011 — codex 부재 시 exec 없음 (REQ-CL-012)

- **Given** `codexLookPath` 가 실패하도록 스텁된 상태
- **When** `cli` / `app` 을 실행하면
- **Then** 둘 다 비영 rc 로 종료하고, exec 호출 횟수는 0 이며, 진단은 설치 조치를 명명한다.
- **And** 맨몸 / `status` 는 rc 0 으로 리드아웃을 내되 바이너리 행을 `not found` 로 적는다 — 진단 명령을 못 쓰게 만드는 것은 REQ-CL-006 의 fail-open 취지에 어긋난다.

## AC-CL-012 — 쓰기 없음 (REQ-CL-013)

범위: **모든 형태, 예외 없음.** 초기화는 `SPEC-CODEX-INIT-001` 로 분리됐으므로 런처에는 쓰기가 허용되는 경로가 없다 (REQ-CL-013).

- **Given** 세 트리의 파일 목록·mtime 스냅샷 — 임시 프로젝트 루트, 임시 CODEX_HOME, **그리고 임시 Claude 프로필 디렉터리**(REQ-CL-013 이 명시적으로 보호하는 세 번째 대상)
- **When** 네 형태(맨몸 / `status` / `cli` / `app`)를 (exec 스텁 상태로) 각각 실행한 뒤 다시 스냅샷하면
- **Then** 세 스냅샷 모두 동일하다 (`.claude/settings.local.json` 무변경, CODEX_HOME 하위 신규 파일 0, 프로필 상태 무변경).

## AC-CL-013 — 중립성 (REQ-CL-014)

- **When** 템플릿 중립성 가드를 실행하면 (`MOAI_TEMPLATE_LEAK_STRICT=1 go test ./internal/template/...`)
- **Then** 통과한다.
- **And Given** 템플릿 가드는 `internal/cli` 를 보지 않으므로 도움말은 따로 판정한다
- **When** `internal/cli` 에서 `codexCmd` 의 생성된 도움말 문자열(`Long` + 예시)을 직접 취해 검사하면
- **Then** SPEC ID(`SPEC-`) · 카드 id(`t197`) · 내부 날짜 · 커밋 SHA 패턴이 각각 0건이다.

## AC-CL-014 — 게이트 (전 REQ)

- **When** `go build ./...` · `go vet ./...` · `GOOS=windows go vet ./...` · `golangci-lint run` 을 실행하면
- **Then** 전부 rc 0 이다.


## AC-CL-015 — 공유 러너 무회귀 (REQ-CL-010)

- **When** 기존 codex 관련 시험을 실행하면 (`go test ./internal/cli/... -run Codex -timeout 600s`)
- **Then** 전부 통과한다.
- **And** `codexCommandRunner` 인터페이스 선언과 그 세 구현체(`realCodexRunner` / `fakeCodexRunner` / `stubCodexRunner`) 어느 것도 이 SPEC 의 diff 에 나타나지 않는다 — 새 seam 은 별도 변수 + 순수 함수들이므로 (M-7, plan §C.2).
- **And** 순수 함수 세 개(`combineCodexStreams` / `parseCodexAuthLine` / `classifyCodexAuthFile`)는 시험에서 프로세스를 띄우지 않고 직접 호출된다 — 시험 파일에 각각 ≥1건의 직접 호출이 존재한다.
- **And** `--version` 조회 경로는 기존 `codexCommandRunner.run` 을 그대로 쓴다 (호출 seam 으로 확인).

## AC-CL-016 — 데스크톱 앱 위임 (REQ-CL-011)

- **Given** exec seam 스텁
- **When** `moai codex app` 을 실행하면
- **Then** 포착 argv 가 정확히 `[codex, app]` 이다 — 앱을 직접 찾아 띄우려는 시도가 없다.
- **And** 이 SPEC 이 추가하는 코드에 앱 번들 경로 탐색·설치 시도가 0건이다 (`/Applications` · `open -a` · 설치 관리자 호출 grep 각 0건).
- **And** codex 가 앱을 못 찾을 때의 처리(설치 관리자 안내 등)는 codex 의 출력을 그대로 통과시키고 재해석하지 않는다.

---

## 판정 제외 (근거 명시)

- **실제 Codex 앱 기동**: CI 러너에 데스크톱 환경이 없다. 운영자 수동 확인 항목으로 `progress.md` 에 남긴다 — 확인 방법은 배선된 프로젝트에서 `moai codex app` 실행 후 앱 전면 등장 여부.
- **실 바이너리 auth 왕복**: 로그인 상태는 머신 상태에 의존한다. 기계 판정은 스텁으로 하고, 실 바이너리 확인은 `moai codex status` 출력 1회를 `progress.md` 에 붙이는 것으로 갈음한다 (§A.2 의 기준선 측정과 대칭).
