# SPEC-CODEX-LAUNCHER-001 — 인수 기준

모든 기준은 기계 판정이다. codex 실 바이너리에 의존하는 항목은 없다 — 프로브 seam (`codexLookPath`, `codexCommandRunner`) 을 스텁해 판정한다.

## AC-CL-001 — 커맨드 등록 (REQ-CL-001)

- **Given** 빌드된 moai 바이너리
- **When** `moai --help` 를 실행하면
- **Then** LAUNCH COMMANDS 그룹에 `codex` 행이 나타나고, `moai codex --help` 가 rc 0 으로 도움말을 출력한다.

## AC-CL-002 — 동사 라우팅 (REQ-CL-002)

- **Given** 등록된 `codex` 커맨드
- **When** 맨몸 / `status` / `cli` / `app` 각각을 스텁 exec seam 으로 호출하면
- **Then** 맨몸과 `status` 는 exec 을 **0회** 수행하고 리드아웃만 내며, `cli` 는 `codex` 를, `app` 은 `codex app` 을 각각 정확히 1회 exec 한다.
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

- **Given** `ProbeCodexSetup` 과 `codexadapter.ValidateConfig` 각각의 호출 횟수를 세는 seam
- **When** 리드아웃을 1회 조립하면
- **Then** 두 함수의 호출 횟수가 각각 ≥1 이다 (런처가 자체 분류·자체 검증을 하지 않았다는 실행 증거).
- **And** `grep -rn "login status" internal/ --include="*.go" | grep -v _test` 의 히트는 공유 분류기 한 곳뿐이다.
- **And** `internal/web` 에는 auth 분류 로직이 여전히 0건이다.
- **And** `git diff` 에 `codexCommandRunner` 인터페이스 선언의 변경이 없다 (REQ-CL-010 후단).

## AC-CL-008 — auth 분류 2단 사다리 (REQ-CL-008) [핵심]

**1단 (파일):**

- **Given** 임시 CODEX_HOME 에 `auth.json` 을 두고 `auth_mode` 를 `chatgpt` / `apikey` / (알 수 없는 값) 로 각각 채운 fixture
- **When** 분류하면
- **Then** 각각 `chatgpt` / `apiKey` / `unknown` 이다. 알 수 없는 값은 추측하지 않는다.
- **And** 어떤 경우에도 토큰 값(`tokens.*`)이나 API 키 값이 읽히거나 출력에 나타나지 않는다 (출력 전문을 키 문자열로 grep → 0건).

**2단 (하강):**

- **Given** `auth.json` 이 없고, stdout 은 비우고 stderr 에만 `Logged in using ChatGPT` 를 쓰는 스텁
- **When** `ProbeCodexSetup` 을 호출하면
- **Then** `AuthProvider == "chatgpt"` 이다.
- **기준선 근거**: 이 절은 수정 전 트리에서 반드시 실패해야 한다 (현행은 `unknown` — M-2 실측). 실패를 먼저 관측한 뒤 수정한다.

## AC-CL-009 — 오류 문구를 인증 성공으로 읽지 않는다 (REQ-CL-009) [핵심]

부분 일치의 오분류를 막는 음성 시험이다. `auth.json` 부재 상태에서, 아래 각 출력을 스텁으로 공급한다:

| 스텁 출력 (rc) | 기대 |
|---|---|
| `error: API key missing` (rc 1) | `unknown` — `apiKey` 아님 |
| `provider configuration unreadable` (rc 1) | `unknown` — `provider` 아님 |
| `failed to reach ChatGPT backend` (rc 1) | `unknown` — `chatgpt` 아님 |
| `Logged in using ChatGPT` (rc 0) | `chatgpt` |
| `Logged in using ChatGPT` (rc 1) | `chatgpt` — 긍정 행이 있으므로 rc 만으로 버리지 않는다 |
| 두 스트림 모두 비어 있음 (rc 0) | `unknown` |

- **근거**: 현행 분류기(`mcp_codex.go:1331`)는 부분 일치라 위 1~3행을 각각 `apiKey` / `provider` / `chatgpt` 로 잘못 분류한다. 앵커된 긍정 행 매칭이 이 여섯 칸을 동시에 만족시키는 유일한 형태다.

## AC-CL-010 — 판정 불가는 조치와 함께 보고 (REQ-CL-009)

- **Given** `auth.json` 부재 + 두 스트림 모두 비어 있는 스텁
- **When** 맨몸 `moai codex` 를 실행하면
- **Then** auth 행은 `unknown` 이고, 출력에 `codex login status` 문자열이 포함되며, 로그아웃 단정 문구는 없다.

## AC-CL-011 — 공유 러너 무회귀 (REQ-CL-010)

- **When** 기존 codex 관련 시험을 실행하면 (`go test ./internal/cli/... -run Codex -timeout 600s`)
- **Then** 전부 통과한다.
- **And** `codexCommandRunner` 인터페이스 선언과 그 세 구현체(`realCodexRunner` / `fakeCodexRunner` / `stubCodexRunner`) 어느 것도 이 SPEC 의 diff 에 나타나지 않는다 — 새 seam 은 별도 변수이므로 (M-7, plan §C.2).

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
