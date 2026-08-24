# SPEC-CODEX-LAUNCHER-001 — 구현 계획

## §A. Pre-flight (측정 완료)

| 항목 | 결과 | 근거 |
|---|---|---|
| 트리 | `WT-codex-launcher` (분기 지점 `9280c96b3`) | `git branch --show-current` |
| t88 M4 포함 | 포함 (`7b217da7c` 조상) | `git merge-base --is-ancestor` rc 0 |
| `moai codex` 부재 | 확인 | `.moai/reports/t197/measurement.md` M-1 |
| auth `unknown` 재현 | 재현 + 원인 확정 | 같은 문서 M-2 |
| 빌드 | green | `go build ./cmd/moai` rc 0 |

## §B. 결정 기록 — 맨몸 `moai codex` 의 의미 (해소됨)

두 해석이 모두 방어 가능했다:

- **(a) 기동** — `moai cc` 가 claude 를 띄우듯 맨몸이 Codex CLI 를 띄우고, 준비 상태는 기동 직전 stderr 한 블록으로 흘린다. 런처 패밀리 대칭에 부합.
- **(b) 리드아웃 + 명시 기동** — 맨몸은 리드아웃만, 기동은 `moai codex cli` / `moai codex app` 으로 명시. 카드 원문("설정 상태 확인 + 앱/CLI 기동 **안내**")에 더 가깝고, 실수로 세션이 교체될 위험이 없다.

**판정: (b)** — 리드 판정 (2026-08-24). REQ-CL-002 를 그 방향으로 재기술하고 AC-CL-002 의 조건절을 해소했다. 파생 결과 하나: 맨몸이 진단 표면이 되므로 codex 바이너리가 없어도 리드아웃은 성공해야 한다 (REQ-CL-012 를 launch 동사에만 적용하도록 좁힘, AC-CL-011 에 대응 절 추가).

## §C. 설계

### C.1 파일 배치

| 파일 | 성격 | 내용 |
|---|---|---|
| `internal/cli/codex_launcher.go` | 신규 | `codexCmd` 코브라 정의 (`GroupID: "launch"`), 세 동사 라우팅, `--spawn` 처리 |
| `internal/cli/codex_readiness.go` | 신규 | 리드아웃 조립 — `ProbeCodexSetup` + `codexwiring` + `CODEX_HOME` 해석을 한 구조체로 |
| `internal/cli/mcp_codex.go` | 수정 | auth 분류 재설계 — 3-seam 분해 (§C.2) |
| `internal/cli/codex_init.go` | 신규 | 미배선 시 초기화 제안 + 기존 생성기 호출 + 지시 계약 확보 (§C.6) |
| `internal/cli/codex_launcher_test.go` 외 | 신규 | 아래 §D 시험 |

`unifiedLaunch` 는 건드리지 않는다 — Claude 전용 경로이고 (§A.6) codex 는 별도 exec 이다. 공유하는 것은 `spawnLaunch` 하나다.

### C.2 auth 분류 재설계 (핵심 결함)

감사 지적 2건이 최초 설계를 뒤집었다. 순서대로:

**(1) 산문 파싱을 1순위에서 내린다.** 최초안은 "합친 스트림에서 `Logged in using ChatGPT` 를 부분 일치" 였는데, 현행 분류기(`mcp_codex.go:1331`)의 부분 일치는 **오류 문구에도 걸린다** — `API key missing` 은 `api key` 를 포함하므로 `apiKey` 로, `provider configuration unreadable` 은 `provider` 로 분류된다. 여기에 "rc 비영이어도 출력이 있으면 분류" 를 얹으면 **오류를 인증 성공으로 읽는다**. REQ-CL-009 의 "판정 불가는 gap 이지 판정이 아니다" 와 정면으로 어긋난다.

**(2) 구조화된 원천이 존재한다** (M-2b 실측). `codex doctor` 가 `stored auth mode: chatgpt` 를 알고 있고 그 원천은 `<CODEX_HOME>/auth.json` 의 `auth_mode` 필드다. doctor 자체는 46초라 런처가 부를 수 없지만(실측), 파일은 즉시 읽힌다.

따라서 **2단 사다리**로 간다:

| 단 | 원천 | 판정 |
|---|---|---|
| 1 | `<CODEX_HOME>/auth.json` 의 `auth_mode` + 그 모드가 함의하는 자격 재료 존재 여부 | 알려진 모드 + 재료 갖춤일 때만 판정 |
| 2 | `codex login status` 의 stdout+stderr 결합 | **전체 행 문법에 맞는 행** 의 캡처값만 |
| — | 그 외 전부 | `unknown` + 조치 안내 |

**2단의 규칙은 앵커가 아니라 전체 행 allowlist 다.** 최초 수정안은 `^logged in` 으로 시작하는 행 **안에서 다시 토큰을 찾는** 것이었는데, 그것도 뚫린다:

```
Logged in state unavailable: API key missing   →  (앵커 통과) → "api key" 부분 일치 → apiKey
```

같은 오분류가 한 겹 안쪽에서 재현된다. 따라서 **행 전체를 문법으로 고정하고 캡처값만 매핑** 한다:

```
(?i)^logged in using (chatgpt|api key)$
```

캡처 그룹이 곧 provider 이고, 캡처되지 않은 행은 어떤 토큰을 품고 있든 분류에 쓰이지 않는다. 문법에 없는 새 provider 문구가 나오면 `unknown` 으로 내린다 — 추측하지 않는다. **rc 비영일 때는 이 문법에 맞는 행이 있을 때만 분류하고, 없으면 `unknown`.**

**1단도 값만 보지 않는다 — 최소 구조 조건을 함께 본다.** `auth_mode` 만 읽으면 stale·불완전한 파일이 긍정 provider 를 만든다:

"존재" 가 아니라 **비어 있지 않은 값** 을 요구한다 — 객체가 있는 것과 자격 재료가 있는 것은 다르다:

| 파일 상태 | 판정 |
|---|---|
| `auth_mode=chatgpt` + `tokens` 중 값이 비지 않은 항목 ≥1 | `chatgpt` |
| `auth_mode=chatgpt` + `tokens` 가 `{}` | 하강 |
| `auth_mode=chatgpt` + `tokens` 값이 전부 `null` 또는 `""` | 하강 |
| `auth_mode=chatgpt` + `tokens` 없음 / `null` | 하강 |
| `auth_mode=apikey` + 키 필드가 비지 않음 | `apiKey` |
| `auth_mode=apikey` + 키 필드 `null` / `""` | 하강 |
| 알 수 없는 `auth_mode` 값 | 하강 |
| 파일 없음 / 파싱 실패 | 하강 |

1단 실패는 오류가 아니라 하강이다.

**비밀값 규율은 grep 이 아니라 타입으로 건다 — 그리고 그 타입은 값을 보존하지 않는다.** 앞선 안은 "비밀 필드 없는 구조체" 라 선언해 놓고 `APIKey string` 으로 **키 전문을 역직렬화** 했다. 문자열로 받는 순간 그 값은 메모리에 존재하고 오류·로그 경로로 샐 수 있다. `json.RawMessage` 도 원문을 보존하므로 같은 문제다.

값을 보존하지 않는 타입을 쓴다:

```go
// nonEmpty 는 "비어 있지 않은 값이 있었다" 는 사실만 남기고 값은 버린다.
type nonEmpty bool

func (n *nonEmpty) UnmarshalJSON(b []byte) error {
    s := strings.TrimSpace(string(b))
    *n = nonEmpty(s != "" && s != "null" && s != `""` && s != "{}")
    return nil   // b 는 여기서 끝난다 — 어디에도 저장하지 않는다
}

type codexAuthFile struct {
    AuthMode string   `json:"auth_mode"`   // 열거값이며 비밀이 아니다
    APIKey   nonEmpty `json:"OPENAI_API_KEY"`
    Tokens   tokenSet `json:"tokens"`      // 아래 — 값 없이 "비지 않은 항목 수" 만
}
```

`tokenSet` 도 같은 방식으로 각 항목의 비어있지 않음만 세고 값은 버린다. 결과적으로 이 타입 집합에는 `string` / `[]byte` / `json.RawMessage` 형태의 자격 필드가 **하나도 없다** — AC-CL-008 이 리플렉션으로 그 부재를 판정한다. 오류를 감쌀 때도 원본 JSON 본문을 넣지 않는다 (경로와 사유만).

**seam — 인터페이스를 넓히지 않되, 시험이 닿을 수 있는 층에 둔다.** 최초안은 `codexCommandRunner` 에 `runCombined` 를 추가하는 것이었고, 그 근거로 "구현체는 프로덕션 1 + 시험 1" 이라 적었다. 실측 결과 **틀렸다** (M-7): 시험 더블이 둘(`fakeCodexRunner`, `stubCodexRunner`)이라 넓히면 둘 다 깨진다.

두 번째 안(`codexAuthProbe` 가 최종 provider 만 반환)도 **틀렸다** — 그러면 시험이 stdout/stderr/rc 를 주입할 지점이 없어져, 최종값을 스텁하면 분류기를 건너뛰고 기존 러너를 스텁하면 stderr 경로에 닿지 못한다. 핵심 회귀 시험이 명목 시험으로 퇴행한다.

세 번째 안(`codexLoginStatusRunner` 가 **이미 결합된** `combined []byte` 를 반환)도 **틀렸다** — 스텁이 결합된 바이트를 돌려주면 "stdout 은 비고 stderr 에만 있다" 는 상태를 표현할 수 없고, 프로덕션이 여전히 stderr 를 버려도 시험은 전부 통과한다. **이 SPEC 이 고치려는 결함이 정확히 회귀 시험을 우회한다.**

두 스트림을 **분리해서** 넘긴다:

```go
// (1) 저수준 실행 seam — 두 스트림을 나눠 반환한다. 시험이 여기서 "stdout 비고 stderr 만" 을 표현한다.
var codexLoginStatusRunner = defaultLoginStatusRunner
    // func(ctx, binaryPath) (stdout, stderr []byte, exitCode int, err error)

// (2) 순수 결합 — 프로덕션의 결합 규칙 자체가 시험 대상이 된다.
func combineCodexStreams(stdout, stderr []byte) []byte

// (3) 순수 파서 — 프로세스도 파일도 만지지 않는다.
func parseCodexAuthLine(combined []byte, exitCode int) string

// (4) 순수 파일 판정 — 바이트를 받아 판정한다. 디스크 접근은 호출자 몫.
//     err 는 파싱 실패 사유이며, ok=false 와 함께 반환된다 (판정은 하강).
func classifyCodexAuthFile(raw []byte) (provider string, ok bool, err error)

// (5) 파일 읽기 조립부 — 경로를 아는 유일한 층. 오류 문안 시험의 대상.
func readCodexAuthFile(codexHome string) (provider string, ok bool, err error)
```

`(4)` 에 `err` 를 붙인 것은 AC 와 시그니처가 어긋나 있었기 때문이다 — 앞선 안의 AC 는 "반환된 오류의 `Error()`" 를 검사하는데 시그니처에는 오류도 경로도 없어 **작성 불가능한 AC** 였다. 오류 문안(경로·사유만, 본문 미포함) 판정은 경로를 아는 `(5)` 의 계약으로 옮긴다.

**프로덕션 결합 경로도 한 번은 실제로 시험한다.** `defaultLoginStatusRunner` 를 **fixture 실행 파일** — `testdata/` 에 커밋된, stdout 을 비우고 stderr 로 한 줄만 쓰는 스크립트 — 에 대고 돌려 두 스트림이 실제로 분리 수집되는지 본다. 세 가지를 지킨다:

- fixture 는 `exec` 로 **직접** 실행한다. 셸을 경유하지 않으므로 인자 해석·인용 문제가 없다.
- 자기 자신을 재실행하는 헬퍼 프로세스 방식은 쓰지 않는다 — `go test` 에서 자기 바이너리를 다시 부르면 수트가 재귀 실행된다.
- Windows 에서는 이 한 건만 skip 한다. 나머지 순수 함수 시험((2)(3)(4))은 전 플랫폼에서 돈다.

`classifyCodexAuth` 는 (5) → (1)+(2)+(3) 순으로 부르는 얇은 조립부가 되고, `codexCommandRunner` 인터페이스와 그 세 구현체는 무변경 → `--version` 경로도 무변경.

### C.3 CODEX_HOME 해석

```
CODEX_HOME 환경변수 설정 → 그 값, source="env"
미설정                    → filepath.Join(os.UserHomeDir(), ".codex"), source="default"
```

존재 여부도 함께 본다 (`os.Stat`). 없으면 `missing` 으로 적고 조치는 `codex login` — 디렉터리를 만들지 않는다 (REQ-CL-013).

### C.4 리드아웃 형태 (고정 6행)

```
codex     0.149.0  (/Users/…/bin/codex)
home      ~/.codex            (default)
auth      chatgpt
wiring    wired               (.codex/hooks.json, .codex/config.toml)
agents    11 TOML
harness   moai hook --harness codex
```

각 행은 실패해도 `unknown` 으로 적고 다음 행으로 넘어간다. `codex doctor` 를 부르지 않는다 — 그건 codex 의 진단이고, 필요하면 마지막 줄에서 명령만 안내한다.

### C.5 기동

- 맨몸 / `status`: 리드아웃만. exec 0회.
- `cli`: `exec` 로 `codex` 를 프로젝트 루트에서 교체 실행. `--` 뒤 인자는 그대로 전달.
- `app`: `codex app` 위임.
- 두 launch 동사는 codex 바이너리 미해결 시 exec 전에 단일 진단으로 종료 (REQ-CL-012). 리드아웃은 이 경우에도 rc 0 으로 나온다.

### C.6 미배선 초기화 + 지시 계약

**런처는 생성기를 부를 뿐 생성 로직을 갖지 않는다.** `.codex/hooks.json` / `.codex/config.toml` 을 무엇으로 채울지는 SPEC-CODEX-WIRING-001 이 이미 정했고, 여기서 두 번째 구현을 만들면 두 판본이 갈라진다. 런처가 하는 일은 셋뿐이다: 배선 부재 판정 → 제안 → 수락 시 `moai init --agent codex` 경로 호출.

수락 전에는 아무것도 쓰지 않는다. 거절하면 기동도 하지 않는다 — 미배선 프로젝트로 들어가면 훅이 하나도 안 붙은 채 세션이 열리고, 그게 조용히 잘못된 상태다. 맨몸 / `status` 는 어느 상태에서도 제안하지 않는다 (읽기는 쓰지 않는다).

**지시 계약**: Codex 는 `AGENTS.md` 를 읽고 Claude 는 `CLAUDE.md` 를 읽는다. 두 하네스가 같은 지시를 보게 하려면 원본을 하나로 두고 나머지가 그것을 가리켜야 한다 — 이 저장소가 t82 에서 택한 구조 그대로다. 초기화가 확보하는 상태:

| 프로젝트 상태 | 하는 일 |
|---|---|
| 둘 다 없음 | 둘 다 만들고 `CLAUDE.md` 에 import 줄 |
| `AGENTS.md` 만 있음 | 원본 **무변경**, `CLAUDE.md` 만 만들어 import |
| `CLAUDE.md` 만 있음 | 기존 내용 보존 + import 줄만 추가, `AGENTS.md` 생성 |
| 둘 다 있고 import 줄도 있음 | 무변경 (멱등) |

로컬 전용 지시 파일이 함께 있으면 그 내용도 계약에 반영한다. 원칙은 하나: **기존 사용자 내용은 보존하고 없는 연결만 채운다.** 재실행해도 import 줄은 1건을 넘지 않는다.

## §D. 마일스톤

| M | 내용 | 산출 |
|---|---|---|
| M1 | auth 분류 2단 사다리 + 회귀 시험 | 3-seam 분해(`codexLoginStatusRunner` / `parseCodexAuthLine` / `classifyCodexAuthFile`), `classifyCodexAuth` 는 얇은 조립부로 (`codexCommandRunner` 무변경) |
| M2 | CODEX_HOME 해석 + 리드아웃 조립 | `codex_readiness.go` + 단위 시험 |
| M3 | 커맨드 표면 + 동사 라우팅 + `--spawn` | `codex_launcher.go` + 등록 시험 |
| M4 | 미배선 초기화 + 지시 계약 | `codex_init.go` — 생성기 호출 위임 + `AGENTS.md`↔`CLAUDE.md` 계약, 멱등 |
| M5 | 도움말·예시 문안 + 중립성 통과 | help 텍스트, 필요 시 템플릿 문서 |

M1 은 독립적으로 가치가 있다 (`moai web` · MCP 도구의 오표시가 그 자체로 해소된다). M2→M3 는 순차, M4 는 M3 이후 (기동 경로가 있어야 그 앞에 제안을 붙일 수 있다), M5 는 마지막.

## §E. 위험

| 위험 | 완화 |
|---|---|
| codex 버전에 따라 `login status` 문구/스트림이 또 바뀐다 | 1단(`auth_mode` 파일)이 산문에 의존하지 않는다. 2단은 결합 스트림을 읽어 스트림 이동에 둔감하고, 문구가 바뀌면 문법 불일치 → `unknown` + 조치 안내로 안전 하강 (REQ-CL-009) |
| `auth.json` 의 형식·경로가 바뀐다 | 1단 실패는 오류가 아니라 2단으로의 하강이고, 2단도 실패하면 `unknown`. 어느 단도 파일을 쓰지 않는다 |
| 인터페이스 확장이 기존 시험 더블을 깬다 | **설계를 바꿔 회피했다** — `codexCommandRunner` 무변경, 저수준 실행 seam + 순수 함수 둘로 분해 (§C.2). 최초안의 "구현체 2개" 측정이 틀렸다는 것이 계기다 (M-7) |
| exec 교체가 칸반/팩토리 세션의 훅 환경을 끌고 들어간다 | 이번 범위에서 `-k` / `-f` 를 다루지 않는다 (§C 제외). 환경 주입 없음 |
| CI 러너에 codex 바이너리가 없다 | 모든 시험이 `codexLookPath` / 러너 스텁 seam 을 통과하도록 작성 — 실 바이너리 의존 0 (SPEC-CODEX-SKILLS-CANONICAL-001 이 세운 선례) |

## §F. 검증 (run-phase)

```
go build ./...
go vet ./...
go test ./internal/cli/... -run 'Codex' -timeout 600s
golangci-lint run
```

`internal/cli` 전체 스위트는 로컬 하한 600s (메모리 규율) — 타깃 실행 후 push 하고 전체 판정은 CI 로 넘긴다.
