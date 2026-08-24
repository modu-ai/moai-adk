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

**판정: (b)** — 리드 판정 (2026-08-24). REQ-CL-002 를 그 방향으로 재기술하고 AC-CL-002 의 조건절을 해소했다. 파생 결과 하나: 맨몸이 진단 표면이 되므로 codex 바이너리가 없어도 리드아웃은 성공해야 한다 (REQ-CL-012 를 launch 동사에만 적용하도록 좁힘, AC-CL-013 에 대응 절 추가).

## §C. 설계

### C.1 파일 배치

| 파일 | 성격 | 내용 |
|---|---|---|
| `internal/cli/codex_launcher.go` | 신규 | `codexCmd` 코브라 정의 (`GroupID: "launch"`), 세 동사 라우팅, `--spawn` 처리 |
| `internal/cli/codex_readiness.go` | 신규 | 리드아웃 조립 — `ProbeCodexSetup` + `codexwiring` + `CODEX_HOME` 해석을 한 구조체로 |
| `internal/cli/mcp_codex.go` | 수정 | auth 분류 재설계 — 3-seam 분해 (§C.2) |
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

| 파일 상태 | 판정 |
|---|---|
| `auth_mode=chatgpt` + `tokens` 존재 | `chatgpt` |
| `auth_mode=chatgpt` + `tokens` 없음/null | 1단 실패 → 2단 하강 |
| `auth_mode=apikey` + API 키 필드 채워짐 | `apiKey` |
| `auth_mode=apikey` + 키 필드 null/빈 값 | 1단 실패 → 2단 하강 |
| 알 수 없는 `auth_mode` 값 | 1단 실패 → 2단 하강 |
| 파일 없음 / 파싱 실패 | 1단 실패 → 2단 하강 |

1단 실패는 오류가 아니라 하강이다.

**비밀값 규율은 grep 이 아니라 타입으로 건다.** 출력에서 키 문자열을 찾는 검사는 직접 누출만 잡고 오류 메시지·로그 경로를 통제하지 못한다. 대신 역직렬화 대상을 비밀 필드가 **없는** 최소 구조체로 고정한다:

```go
type codexAuthFile struct {
    AuthMode string `json:"auth_mode"`
    APIKey   string `json:"OPENAI_API_KEY"` // 존재 여부 판정에만 쓰고 값은 어디에도 싣지 않는다
}
```

`tokens` 는 필드가 없으므로 구조체에 들어오지 않는다. `tokens` 의 존재 여부만 필요할 때는 `json.RawMessage` 길이로 판정하고 내용은 보지 않는다. 오류를 감쌀 때 원본 JSON 본문을 오류 문자열에 넣지 않는다 (경로와 사유만).

**seam — 인터페이스를 넓히지 않되, 시험이 닿을 수 있는 층에 둔다.** 최초안은 `codexCommandRunner` 에 `runCombined` 를 추가하는 것이었고, 그 근거로 "구현체는 프로덕션 1 + 시험 1" 이라 적었다. 실측 결과 **틀렸다** (M-7): 시험 더블이 둘(`fakeCodexRunner`, `stubCodexRunner`)이라 넓히면 둘 다 깨진다.

두 번째 안(`codexAuthProbe` 가 최종 provider 만 반환)도 **틀렸다** — 그러면 시험이 stdout/stderr/rc 를 주입할 지점이 없어져, 최종값을 스텁하면 분류기를 건너뛰고 기존 러너를 스텁하면 stderr 경로에 닿지 못한다. 핵심 회귀 시험이 명목 시험으로 퇴행한다.

셋으로 나눈다:

```go
// (1) 저수준 실행 seam — 시험이 stdout/stderr/rc 를 여기서 주입한다.
var codexLoginStatusRunner = defaultLoginStatusRunner
    // func(ctx, binaryPath) (combined []byte, exitCode int, err error)

// (2) 순수 파서 — 프로세스도 파일도 만지지 않는다. 단위 시험의 주 대상.
func parseCodexAuthLine(combined []byte, exitCode int) string

// (3) 순수 파일 판정 — 바이트를 받아 판정한다. 디스크 접근은 호출자 몫.
func classifyCodexAuthFile(raw []byte) (provider string, ok bool)
```

`classifyCodexAuth` 는 (3) → (1)+(2) 순으로 부르는 얇은 조립부가 되고, `codexCommandRunner` 인터페이스와 그 세 구현체는 무변경 → `--version` 경로도 무변경.

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

## §D. 마일스톤

| M | 내용 | 산출 |
|---|---|---|
| M1 | auth 분류 2단 사다리 + 회귀 시험 | 3-seam 분해(`codexLoginStatusRunner` / `parseCodexAuthLine` / `classifyCodexAuthFile`), `classifyCodexAuth` 는 얇은 조립부로 (`codexCommandRunner` 무변경) |
| M2 | CODEX_HOME 해석 + 리드아웃 조립 | `codex_readiness.go` + 단위 시험 |
| M3 | 커맨드 표면 + 동사 라우팅 + `--spawn` | `codex_launcher.go` + 등록 시험 |
| M4 | 도움말·예시 문안 + 중립성 통과 | help 텍스트, 필요 시 템플릿 문서 |

M1 은 독립적으로 가치가 있다 (`moai web` · MCP 도구의 오표시가 그 자체로 해소된다). M2→M3 는 순차, M4 는 M3 이후.

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
