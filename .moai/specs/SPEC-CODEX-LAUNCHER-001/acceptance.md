# SPEC-CODEX-LAUNCHER-001 — 인수 기준

모든 기준은 기계 판정이다. codex 실 바이너리에 의존하는 항목은 없다 — 프로브 seam (`codexLookPath`, `codexCommandRunner`) 을 스텁해 판정한다.

## AC-CL-001 — 커맨드 등록 (REQ-CL-001)

- **Given** 빌드된 moai 바이너리
- **When** `moai --help` 를 실행하면
- **Then** LAUNCH COMMANDS 그룹에 `codex` 행이 나타나고, `moai codex --help` 가 rc 0 으로 도움말을 출력한다.

## AC-CL-002 — 동사 라우팅 (REQ-CL-002)

- **Given** 등록된 `codex` 커맨드
exec seam 은 호출 횟수뿐 아니라 **argv 전체와 cwd** 를 포착한다 — 명령 이름만 맞고 인자나 실행 위치가 틀린 구현이 통과하지 못하게.

- **When** 맨몸 / `status` / `cli` / `app` 각각을 호출하면
- **Then** 맨몸과 `status` 는 exec **0회**, `cli` 와 `app` 은 각 1회다.
- **And** `cli` 의 포착된 cwd 는 **호출자의 프로젝트 루트** 와 일치한다 (worktree 에서 부르면 그 worktree — 다른 트리에서 codex 가 뜨면 세션이 엉뚱한 코드를 본다).
- **And** `app` 의 포착 argv 는 정확히 `[codex, app]` 이다.
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
| `.codex/` 없음 | `not wired` + `moai init --agent codex` |
| hooks.json + config.toml 정상 | `wired` |
| hooks.json 이 화이트리스트 위반 키 보유 | `invalid` (건강 상태로 보고하지 않는다) |
| config.toml 은 있고 hooks.json 없음 | `partial` — 어느 쪽이 없는지 명시 |

- **And** 바이너리 경로 · 버전 행은 스텁이 공급한 값과 문자열 일치한다 (라벨만 있고 값이 비어도 통과하는 grep 판정을 쓰지 않는다).
- **And** 화이트리스트 판정은 `codexadapter.ValidateConfig` 를 호출해 얻는다 (AC-CL-007 이 재구현 부재를 강제).

## AC-CL-005 — CODEX_HOME 해석과 출처 표시 (REQ-CL-005)

- **Given** `CODEX_HOME=/tmp/xyz` 가 설정된 프로세스
- **When** 리드아웃을 조립하면
- **Then** 값은 `/tmp/xyz`, 출처는 `env` 로 보고된다.
- **And Given** `CODEX_HOME` 미설정
- **Then** 값은 `<home>/.codex`, 출처는 `default` 로 보고된다.

## AC-CL-006 — 배선 없는 프로젝트는 정보성 (REQ-CL-006)

- **Given** `.codex/` 가 없는 임시 프로젝트 루트
- **When** `moai codex status` 를 실행하면
- **Then** rc 는 0 이고, 배선 행은 미배선 상태로 적히며, 출력에 `moai init --agent codex` 문자열이 포함된다.

## AC-CL-007 — 분류 구현 단일성 (REQ-CL-007, REQ-CL-010)

텍스트 중복 부재만으로는 "공유 프로브를 실제로 쓴다" 를 증명하지 못한다. 호출로 판정한다:

호출 횟수 `≥1` 은 "불렀다" 만 증명하고 "그 값을 썼다" 는 증명하지 못한다 — 공유 함수를 부른 뒤 결과를 버리고 자체 분류를 하는 구현도 통과한다. **sentinel 전파** 로 판정한다:

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

**1단 — 순수 파일 판정 `classifyCodexAuthFile(raw []byte)` 을 직접 시험한다** (디스크 접근 없음). fixture 행렬:

| `auth.json` 내용 | 기대 `(provider, ok)` |
|---|---|
| `auth_mode=chatgpt` + `tokens` 객체 존재 | `("chatgpt", true)` |
| `auth_mode=chatgpt` + `tokens` 없음 | `("", false)` — 하강 |
| `auth_mode=chatgpt` + `tokens: null` | `("", false)` — 하강 |
| `auth_mode=apikey` + `OPENAI_API_KEY` 채워짐 | `("apiKey", true)` |
| `auth_mode=apikey` + `OPENAI_API_KEY: null` | `("", false)` — 하강 |
| `auth_mode=apikey` + `OPENAI_API_KEY: ""` | `("", false)` — 하강 |
| `auth_mode=totally-new-mode` | `("", false)` — 추측하지 않는다 |
| `{` (파싱 실패) | `("", false)` — 하강 |
| 빈 바이트 | `("", false)` — 하강 |

**비밀값 규율 — 타입으로 판정한다** (출력 grep 은 직접 누출만 잡으므로 보조 수단):

- **When** 역직렬화 대상 구조체의 필드 집합을 리플렉션으로 열거하면
- **Then** 토큰 계열 필드(`id_token` / `access_token` / `refresh_token`)에 대응하는 필드가 **하나도 없다**.
- **And Given** 파싱이 실패하도록 만든 fixture 에 가짜 토큰 문자열을 심고
- **When** 반환된 오류의 `Error()` 전문을 검사하면
- **Then** 그 문자열이 포함되지 않는다 (오류는 경로와 사유만 싣는다).

**2단 — 순수 파서 `parseCodexAuthLine(combined, exitCode)` 을 직접 시험한다** (프로세스 없음). 표는 AC-CL-009.

**통합 1건** — 두 단이 실제로 이어지는지:

- **Given** `auth.json` 부재 + `codexLoginStatusRunner` 가 stdout 을 비우고 stderr 로만 `Logged in using ChatGPT` (rc 0) 를 돌려주도록 스텁
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

## AC-CL-011 — 공유 러너 무회귀 (REQ-CL-010)

- **When** 기존 codex 관련 시험을 실행하면 (`go test ./internal/cli/... -run Codex -timeout 600s`)
- **Then** 전부 통과한다.
- **And** `codexCommandRunner` 인터페이스 선언과 그 세 구현체(`realCodexRunner` / `fakeCodexRunner` / `stubCodexRunner`) 어느 것도 이 SPEC 의 diff 에 나타나지 않는다 — 새 seam 은 별도 변수 + 순수 함수 둘이므로 (M-7, plan §C.2).
- **And** 순수 함수 두 개(`parseCodexAuthLine` / `classifyCodexAuthFile`)는 시험에서 프로세스를 띄우지 않고 직접 호출된다 — 시험 파일에 이 둘의 직접 호출이 각각 ≥1건 존재한다.

## AC-CL-012 — 데스크톱 앱 위임 (REQ-CL-011)

- **Given** exec seam 스텁
- **When** `moai codex app` 을 실행하면
- **Then** 정확히 `codex app` 이 호출되고, 앱 경로 탐색이나 설치 시도 코드는 실행되지 않는다 (`grep` 으로 `/Applications` 류 하드코딩 0건).

## AC-CL-013 — codex 부재 시 exec 없음 (REQ-CL-012)

- **Given** `codexLookPath` 가 실패하도록 스텁된 상태
- **When** `cli` / `app` 을 실행하면
- **Then** 둘 다 비영 rc 로 종료하고, exec 호출 횟수는 0 이며, 진단은 설치 조치를 명명한다.
- **And** 맨몸 / `status` 는 rc 0 으로 리드아웃을 내되 바이너리 행을 `not found` 로 적는다 — 진단 명령을 못 쓰게 만드는 것은 REQ-CL-006 의 fail-open 취지에 어긋난다.

## AC-CL-014 — 쓰기 없음 (REQ-CL-013)

- **Given** 세 트리의 파일 목록·mtime 스냅샷 — 임시 프로젝트 루트, 임시 CODEX_HOME, **그리고 임시 Claude 프로필 디렉터리**(REQ-CL-013 이 명시적으로 보호하는 세 번째 대상)
- **When** 네 형태(맨몸 / `status` / `cli` / `app`)를 (exec 스텁 상태로) 각각 실행한 뒤 다시 스냅샷하면
- **Then** 세 스냅샷 모두 동일하다 (`.claude/settings.local.json` 무변경, CODEX_HOME 하위 신규 파일 0, 프로필 상태 무변경).

## AC-CL-015 — 중립성 (REQ-CL-014)

- **When** 템플릿 중립성 가드를 실행하면 (`MOAI_TEMPLATE_LEAK_STRICT=1 go test ./internal/template/...`)
- **Then** 통과한다.
- **And Given** 템플릿 가드는 `internal/cli` 를 보지 않으므로 도움말은 따로 판정한다
- **When** `internal/cli` 에서 `codexCmd` 의 생성된 도움말 문자열(`Long` + 예시)을 직접 취해 검사하면
- **Then** SPEC ID(`SPEC-`) · 카드 id(`t197`) · 내부 날짜 · 커밋 SHA 패턴이 각각 0건이다.

## AC-CL-016 — 게이트 (전 REQ)

- **When** `go build ./...` · `go vet ./...` · `GOOS=windows go vet ./...` · `golangci-lint run` 을 실행하면
- **Then** 전부 rc 0 이다.

---

## 판정 제외 (근거 명시)

- **실제 Codex 앱 기동**: CI 러너에 데스크톱 환경이 없다. 운영자 수동 확인 항목으로 `progress.md` 에 남긴다 — 확인 방법은 배선된 프로젝트에서 `moai codex app` 실행 후 앱 전면 등장 여부.
- **실 바이너리 auth 왕복**: 로그인 상태는 머신 상태에 의존한다. 기계 판정은 스텁으로 하고, 실 바이너리 확인은 `moai codex status` 출력 1회를 `progress.md` 에 붙이는 것으로 갈음한다 (§A.2 의 기준선 측정과 대칭).
