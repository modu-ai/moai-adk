# SPEC-CODEX-LAUNCHER-001 — 구현 계획

## §A. Pre-flight (측정 완료)

| 항목 | 결과 | 근거 |
|---|---|---|
| 측정 대상 트리 | `1ed61e4ac` | 전사본 L24-26 |
| 측정 대상 브랜치 | `WT-codex-launcher` | 전사본 L283-285 |
| t88 M4 포함 | 포함 (`7b217da7c` 조상) | `git merge-base --is-ancestor` rc 0 |
| `moai codex` 부재 | 확인 | `.moai/reports/t197/measurement.md` M-1 |
| auth `unknown` 재현 | 재현 + 원인 확정 | 같은 문서 M-2 |
| 빌드 | green | `go build ./cmd/moai` rc 0 |

## §B. 결정 기록 — 맨몸 `moai codex` 의 의미 (해소됨)

두 해석이 모두 방어 가능했다:

- **(a) 기동** — `moai cc` 가 claude 를 띄우듯 맨몸이 Codex CLI 를 띄우고, 준비 상태는 기동 직전 stderr 한 블록으로 흘린다. 런처 패밀리 대칭에 부합.
- **(b) 리드아웃 + 명시 기동** — 맨몸은 리드아웃만, 기동은 `moai codex cli` / `moai codex app` 으로 명시. 카드 원문("설정 상태 확인 + 앱/CLI 기동 **안내**")에 더 가깝고, 실수로 현재 세션이 codex 에 넘어갈 위험이 없다.

**판정: (b)** — 리드 판정 (2026-08-24). REQ-CL-002 를 그 방향으로 재기술하고 AC-CL-002 의 조건절을 해소했다. 파생 결과 하나: 맨몸이 진단 표면이 되므로 codex 바이너리가 없어도 리드아웃은 성공해야 한다 (REQ-CL-012 를 launch 동사에만 적용하도록 좁힘, AC-CL-011 에 대응 절 추가).

## §C. 설계

### C.1 파일 배치

| 파일 | 성격 | 내용 |
|---|---|---|
| `internal/cli/codex_launcher.go` | 신규 | `codexCmd` 코브라 정의 (`GroupID: "launch"`), 세 동사 라우팅, `--spawn` 처리 |
| `internal/cli/codex_readiness.go` | 신규 | 리드아웃 조립 — `ProbeCodexSetup` + `codexwiring` + `CODEX_HOME` 해석을 한 구조체로 |
| `internal/cli/mcp_codex.go` | 수정 | auth 분류 재설계 — 3-seam 분해 (§C.2) |
| `internal/cli/codex_launcher_test.go` 외 | 신규 | 아래 §D 시험 |

`unifiedLaunch` 는 건드리지 않는다 — Claude 전용 경로이고 (§A.6) codex 는 별도 기동 경로다. 공유하는 것은 `spawnLaunch` 하나다.

### C.2 auth 분류 재설계 (핵심 결함)

감사 지적 2건이 최초 설계를 뒤집었다. 순서대로:

**(1) 산문 파싱을 1순위에서 내린다.** 최초안은 "합친 스트림에서 `Logged in using ChatGPT` 를 부분 일치" 였는데, 현행 분류기(`mcp_codex.go:1331`)의 부분 일치는 **오류 문구에도 걸린다** — `API key missing` 은 `api key` 를 포함하므로 `apiKey` 로, `provider configuration unreadable` 은 `provider` 로 분류된다. 여기에 "rc 비영이어도 출력이 있으면 분류" 를 얹으면 **오류를 인증 성공으로 읽는다**. REQ-CL-009 의 "판정 불가는 gap 이지 판정이 아니다" 와 정면으로 어긋난다.

**(2) 구조화된 원천이 존재한다** (M-2b 실측). `codex doctor` 가 `stored auth mode: chatgpt` 를 알고 있고 그 원천은 `<CODEX_HOME>/auth.json` 의 `auth_mode` 필드다. doctor 자체는 전사본 실측 67초라 런처가 부를 수 없지만, 파일은 즉시 읽힌다.

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
| `auth_mode=chatgpt` + `tokens` 중 **인정된 키**(`access_token`/`id_token`/`refresh_token`)의 값이 비지 않은 것 ≥1 | `chatgpt` |
| `auth_mode=chatgpt` + `tokens` 에 무관한 키만 (`{"irrelevant":"x"}`) | 하강 — 키를 안 보는 구현이 뚫리는 지점 |
| `auth_mode=chatgpt` + `tokens` 에 `account_id` 만 | 하강 — 계정 메타데이터는 자격 재료가 아니다 |
| `auth_mode=chatgpt` + `tokens` 가 `{}` | 하강 |
| `auth_mode=chatgpt` + `tokens` 값이 전부 `null` 또는 `""` | 하강 |
| `auth_mode=chatgpt` + `tokens` 없음 / `null` | 하강 |
| `auth_mode=apikey` + 키 필드가 비지 않음 | `apiKey` |
| `auth_mode=apikey` + 키 필드 `null` / `""` | 하강 |
| 알 수 없는 `auth_mode` 값 | 하강 |
| 파일 없음 / 파싱 실패 | 하강 |

1단 실패는 오류가 아니라 하강이다.

**비밀값 규율은 grep 이 아니라 타입으로 건다 — 그리고 그 타입은 값을 보존하지 않는다.** 앞선 안은 "비밀 필드 없는 구조체" 라 선언해 놓고 `APIKey string` 으로 **키 전문을 역직렬화** 했다. 문자열로 받는 순간 그 값은 메모리에 존재하고 오류·로그 경로로 샐 수 있다. `json.RawMessage` 도 원문을 보존하므로 같은 문제다.

값을 보존하지 않는 타입을 쓰되, **원문 바이트를 문자열로 비교하지 않는다.** 감사가 이 함정을 실측으로 보여줬다 — 원문 비교판(`s != "{}" && s != "null" && …`)은 `{ }`(공백 하나), `false`, `0`, `[]` 를 전부 "비어 있지 않음" 으로 통과시킨다:

```
'{}'  -> False      '{ }' -> True   ← 같은 뜻인데 갈린다
'null'-> False      'false'-> True  ← 타입이 아예 다른데 통과
'""'  -> False      '0'   -> True
                    '[]'  -> True
```

원문이 아니라 **타입** 으로 판정한다. 자격 재료는 "비어 있지 않은 JSON 문자열" 이어야 하고, 다른 JSON 타입은 전부 부재로 친다:

```go
// nonEmptyString 은 "비어 있지 않은 문자열이 있었다" 는 사실만 남기고 값은 버린다.
// 문자열이 아닌 어떤 JSON 타입(null/false/0/[]/{})도 부재로 판정된다.
type nonEmptyString bool

func (n *nonEmptyString) UnmarshalJSON(b []byte) error {
    var s string
    if err := json.Unmarshal(b, &s); err != nil {
        *n = false            // 문자열이 아니면 자격 재료가 아니다
        return nil            // 파일 전체를 깨뜨리지 않고 이 필드만 부재 처리
    }
    *n = nonEmptyString(strings.TrimSpace(s) != "")
    return nil                // s 는 여기서 끝난다 — 어디에도 저장하지 않는다
}

// chatgptCredentialKeys 는 로그인 자격을 실제로 담는 키다. account_id 같은
// 계정 메타데이터는 여기 없다 — 그것만 있는 파일은 로그인 상태가 아니다.
var chatgptCredentialKeys = map[string]bool{
    "access_token": true, "id_token": true, "refresh_token": true,
}

// tokenSet 은 인정된 키 중 값이 비어 있지 않은 문자열인 것만 센다. 객체가 아니면 0.
type tokenSet struct{ credentialCount int }

func (t *tokenSet) UnmarshalJSON(b []byte) error {
    var m map[string]nonEmptyString
    if err := json.Unmarshal(b, &m); err != nil {
        t.credentialCount = 0   // 객체가 아니면(배열·문자열·불리언·수·null) 자격 재료 없음
        return nil
    }
    for k, v := range m {
        if chatgptCredentialKeys[k] && bool(v) {
            t.credentialCount++   // 키를 안 보면 {"irrelevant":"x"} 가 인증으로 통과한다
        }
    }
    return nil
}

type codexAuthFile struct {
    AuthMode string         `json:"auth_mode"`   // 열거값이며 비밀이 아니다
    APIKey   nonEmptyString `json:"OPENAI_API_KEY"`
    Tokens   tokenSet       `json:"tokens"`
}
```

판정: `apikey` 는 `APIKey == true`, `chatgpt` 는 `Tokens.credentialCount >= 1` 일 때만 확정하고, 나머지는 전부 하강.

결과적으로 이 타입 집합에는 `auth_mode` 하나를 빼면 `string` / `[]byte` / `json.RawMessage` 형태의 필드가 **하나도 없다** — AC-CL-008 이 리플렉션으로 (중첩 타입까지 재귀해) 그 부재를 판정한다. 오류를 감쌀 때도 원본 JSON 본문을 넣지 않는다 (경로와 사유만).

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

**프로덕션 결합 경로도 실제로 시험한다.** `defaultLoginStatusRunner` 를 **fixture 실행 파일 3종** — `testdata/` 에 커밋된 stderr 전용 / stdout 전용 / 양쪽 스크립트 — 에 대고 돌려 두 스트림이 실제로 분리 수집되는지, 그리고 결합 결과가 세 칸 모두에서 그 줄들을 담는지 본다 (한 축만 두면 한쪽 스트림을 버리는 구현과 올바른 구현이 구별되지 않는다 — AC-CL-008). 세 가지를 지킨다:

- fixture 는 `exec` 로 **직접** 실행한다. 셸을 경유하지 않으므로 인자 해석·인용 문제가 없다.
- 자기 자신을 재실행하는 헬퍼 프로세스 방식은 쓰지 않는다 — `go test` 에서 자기 바이너리를 다시 부르면 수트가 재귀 실행된다.
- Windows 에서는 이 세 칸만 skip 한다. 나머지 순수 함수 시험((2)(3)(4))은 전 플랫폼에서 돈다 — GOOS skip 은 이 셋뿐이며 그 개수를 AC-CL-014 가 센다.

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

- 맨몸 / `status`: 리드아웃만. 기동 0회.
- `cli`: `os/exec` 로 `codex` 를 프로젝트 루트에서 **자식 프로세스로 띄우고**, 그 프로세스가 끝날 때까지 기다린 뒤 그 종료코드를 moai 의 종료코드로 삼는다. `--` 뒤 인자는 그대로 전달.
- `app`: `codex app` 위임. 같은 자식-프로세스 형태다.
- 두 launch 동사는 codex 바이너리 미해결 시 기동 전에 단일 진단으로 종료 (REQ-CL-012). 리드아웃은 이 경우에도 rc 0 으로 나온다.

**프로세스 교체(`syscall.Exec`)를 쓰지 않는다.** 그 호출은 unix 전용이라 Windows 에 대응 심볼이 없고, 그러면 플랫폼별 파일을 갈라야 한다. 이 SPEC 은 OS 빌드 태그 0건과 `GOOS=windows` 컴파일 통과를 함께 요구하므로(AC-CL-014) 교체와는 양립하지 않는다. 런처가 할 일은 codex 를 **띄우는 것**이지 프로세스 정체성을 물려주는 것이 아니므로, 전 플랫폼 공통의 `os/exec` 를 쓰고 교체가 공짜로 주던 성질은 아래처럼 각각 대체한다.

| 교체가 주던 성질 | 대체물 | 어디서 판정하는가 |
|---|---|---|
| 종료코드가 그대로 셸에 도달 | 부모가 자식의 종료코드를 받아 **그대로 전파** 한다 | AC-CL-016 의 rc 동등 칸 + AC-CL-002 의 `cli` rc 전파 칸 |
| 시그널이 codex 에 직접 전달 | 부모가 **중계** 한다 — 부모가 받은 시그널을 자식에게 넘기고, 자식이 끝날 때까지 자신은 죽지 않는다 | run-phase 구현 규약(§F). 시그널 왕복 자체는 CI 에서 관측하지 않는다 |
| 프로세스 트리 깊이가 늘지 않음 | **한 단 깊어진다.** 무해하다 — moai 부모는 대기만 하고 자식의 스트림·종료코드를 가로막지 않는다 | 판정 대상 아님 (관측 가능한 결과가 없다) |
| 대화형 tty 를 codex 가 그대로 물려받음 | 기동 seam 이 부모 자신의 `os.Stdin` / `os.Stdout` / `os.Stderr` **값 자체** 를 자식에게 넘긴다 (파이프·버퍼·`io.Discard` 로 바꾸지 않는다) | AC-CL-002 의 stdio 항등 칸. tty 왕복 실측은 Gap — acceptance.md 「판정 제외」 참조 |

## §D. 마일스톤

| M | 내용 | 산출 |
|---|---|---|
| M1 | auth 분류 2단 사다리 + 회귀 시험 | 3-seam 분해(`codexLoginStatusRunner` / `parseCodexAuthLine` / `classifyCodexAuthFile`), `classifyCodexAuth` 는 얇은 조립부로 (`codexCommandRunner` 무변경) |
| M2 | CODEX_HOME 해석 + 리드아웃 조립 | `codex_readiness.go` + 단위 시험 |
| M3 | 커맨드 표면 + 동사 라우팅 + `--spawn` | `codex_launcher.go` + 등록 시험 |
| M4 | 도움말·예시 문안 + 중립성 통과 | help 텍스트, 필요 시 템플릿 문서 |

M1 은 독립적으로 가치가 있다 (`moai web` · MCP 도구의 오표시가 그 자체로 해소된다). M2→M3 는 순차, M4 는 마지막. 미배선 초기화는 `SPEC-CODEX-INIT-001` 로 분리됐고 그쪽 M1 이 여기 M3(기동 경로)에 의존한다.

## §E. 위험

| 위험 | 완화 |
|---|---|
| codex 버전에 따라 `login status` 문구/스트림이 또 바뀐다 | 1단(`auth_mode` 파일)이 산문에 의존하지 않는다. 2단은 결합 스트림을 읽어 스트림 이동에 둔감하고, 문구가 바뀌면 문법 불일치 → `unknown` + 조치 안내로 안전 하강 (REQ-CL-009) |
| `auth.json` 의 형식·경로가 바뀐다 | 1단 실패는 오류가 아니라 2단으로의 하강이고, 2단도 실패하면 `unknown`. 어느 단도 파일을 쓰지 않는다 |
| 인터페이스 확장이 기존 시험 더블을 깬다 | **설계를 바꿔 회피했다** — `codexCommandRunner` 무변경, 저수준 실행 seam + 순수 함수 둘로 분해 (§C.2). 최초안의 "구현체 2개" 측정이 틀렸다는 것이 계기다 (M-7) |
| 기동한 codex 자식이 칸반/팩토리 세션의 훅 환경을 끌고 들어간다 | 이번 범위에서 `-k` / `-f` 를 다루지 않는다 (§C 제외). 환경 주입 없음 |
| CI 러너에 codex 바이너리가 없다 | 모든 시험이 `codexLookPath` / 러너 스텁 seam 을 통과하도록 작성 — 실 바이너리 의존 0 (SPEC-CODEX-SKILLS-CANONICAL-001 이 세운 선례) |

## §F. 검증 (run-phase)

```
go build ./...
go vet ./...
go test ./internal/cli/... -run 'Codex' -timeout 600s
golangci-lint run
```

`internal/cli` 전체 스위트는 로컬 하한 600s (메모리 규율) — 타깃 실행 후 push 하고 전체 판정은 CI 로 넘긴다.
