# SPEC-CODEX-LAUNCHER-001 — 구현 계획

## §A. Pre-flight (측정 완료)

| 항목 | 결과 | 근거 |
|---|---|---|
| 트리 | `WT-codex-launcher` @ `9280c96b3` | `git branch --show-current` |
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
| `internal/cli/mcp_codex.go` | 수정 | auth 스트림 오독 수정 (§C.2) |
| `internal/cli/codex_launcher_test.go` 외 | 신규 | 아래 §D 시험 |

`unifiedLaunch` 는 건드리지 않는다 — Claude 전용 경로이고 (§A.6) codex 는 별도 exec 이다. 공유하는 것은 `spawnLaunch` 하나다.

### C.2 auth 분류 재설계 (핵심 결함)

감사 지적 2건이 최초 설계를 뒤집었다. 순서대로:

**(1) 산문 파싱을 1순위에서 내린다.** 최초안은 "합친 스트림에서 `Logged in using ChatGPT` 를 부분 일치" 였는데, 현행 분류기(`mcp_codex.go:1331`)의 부분 일치는 **오류 문구에도 걸린다** — `API key missing` 은 `api key` 를 포함하므로 `apiKey` 로, `provider configuration unreadable` 은 `provider` 로 분류된다. 여기에 "rc 비영이어도 출력이 있으면 분류" 를 얹으면 **오류를 인증 성공으로 읽는다**. REQ-CL-009 의 "판정 불가는 gap 이지 판정이 아니다" 와 정면으로 어긋난다.

**(2) 구조화된 원천이 존재한다** (M-2b 실측). `codex doctor` 가 `stored auth mode: chatgpt` 를 알고 있고 그 원천은 `<CODEX_HOME>/auth.json` 의 `auth_mode` 필드다. doctor 자체는 46초라 런처가 부를 수 없지만(실측), 파일은 즉시 읽힌다.

따라서 **2단 사다리**로 간다:

| 단 | 원천 | 판정 |
|---|---|---|
| 1 | `<CODEX_HOME>/auth.json` 의 `auth_mode` (+ `OPENAI_API_KEY` 비-null 여부) | `chatgpt` / `apiKey` / `provider` — 알려진 값만 |
| 2 | `codex login status` 의 stdout+stderr 결합 | **앵커된 긍정 행** 에서만 분류 |
| — | 그 외 전부 | `unknown` + 조치 안내 |

2단의 앵커 규칙: 행 단위로 보되 `^logged in` (대소문자 무시) 으로 시작하는 행에서만 provider 토큰을 읽는다. 오류 진단문에 우연히 `provider` / `API key` / `ChatGPT` 가 섞여 있어도 그 행은 긍정 행이 아니므로 분류에 쓰이지 않는다. **rc 비영일 때는 이 앵커 행이 있을 때만 분류하고, 없으면 `unknown`** — 최초안의 "출력만 있으면 분류" 는 폐기한다.

비밀값 규율: `auth.json` 에서 읽는 것은 `auth_mode` 문자열과 `OPENAI_API_KEY` 의 **존재 여부(bool)** 뿐이다. 토큰 값·키 값은 읽지도 출력하지도 않는다.

**seam — 인터페이스를 넓히지 않는다.** 최초안은 `codexCommandRunner` 에 `runCombined` 를 추가하는 것이었고, 그 근거로 "구현체는 프로덕션 1 + 시험 1" 이라 적었다. 실측 결과 **틀렸다** (M-7): Go 인터페이스는 구조적이라 이름 grep 이 암묵 구현을 놓쳤고, 실제로는 시험 더블이 둘(`fakeCodexRunner`, `stubCodexRunner`)이다. 인터페이스를 넓히면 둘 다 깨진다.

대신 **좁은 별도 seam** 을 둔다:

```go
// 패키지 수준 변수 — 시험이 이것만 갈아끼운다.
var codexAuthProbe = defaultCodexAuthProbe   // func(ctx, codexHome, binaryPath) (provider string)
```

`codexCommandRunner` 는 무변경 → 기존 시험 더블 2개 모두 무변경, `--version` 경로도 무변경.

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
agents    12 TOML
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
| M1 | auth 분류 2단 사다리 + 회귀 시험 | `codexAuthProbe` seam 신설, `classifyCodexAuth` 를 앵커 매칭으로 교체 (`codexCommandRunner` 무변경) |
| M2 | CODEX_HOME 해석 + 리드아웃 조립 | `codex_readiness.go` + 단위 시험 |
| M3 | 커맨드 표면 + 동사 라우팅 + `--spawn` | `codex_launcher.go` + 등록 시험 |
| M4 | 도움말·예시 문안 + 중립성 통과 | help 텍스트, 필요 시 템플릿 문서 |

M1 은 독립적으로 가치가 있다 (`moai web` · MCP 도구의 오표시가 그 자체로 해소된다). M2→M3 는 순차, M4 는 M3 이후.

## §E. 위험

| 위험 | 완화 |
|---|---|
| codex 버전에 따라 `login status` 문구/스트림이 또 바뀐다 | 1단(`auth_mode` 파일)이 산문에 의존하지 않는다. 2단은 결합 스트림을 읽어 스트림 이동에 둔감하고, 문구가 바뀌면 앵커 불일치 → `unknown` + 조치 안내로 안전 하강 (REQ-CL-009) |
| `auth.json` 의 형식·경로가 바뀐다 | 1단 실패는 오류가 아니라 2단으로의 하강이고, 2단도 실패하면 `unknown`. 어느 단도 파일을 쓰지 않는다 |
| 인터페이스 확장이 기존 시험 더블을 깬다 | **설계를 바꿔 회피했다** — `codexCommandRunner` 무변경, 좁은 `codexAuthProbe` 변수 seam 신설 (§C.2). 최초안의 "구현체 2개" 측정이 틀렸다는 것이 계기다 (M-7: 구조적 인터페이스라 이름 grep 이 `stubCodexRunner` 를 놓쳤다) |
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
