# SPEC-CODEX-LAUNCHER-001 — 구현 계획

## §A. Pre-flight (측정 완료)

| 항목 | 결과 | 근거 |
|---|---|---|
| 트리 | `WT-codex-launcher` @ `9280c96b3` | `git branch --show-current` |
| t88 M4 포함 | 포함 (`7b217da7c` 조상) | `git merge-base --is-ancestor` rc 0 |
| `moai codex` 부재 | 확인 | `.moai/reports/t197/measurement.md` M-1 |
| auth `unknown` 재현 | 재현 + 원인 확정 | 같은 문서 M-2 |
| 빌드 | green | `go build ./cmd/moai` rc 0 |

## §B. 열린 결정 1건

[NEEDS CLARIFICATION: 맨몸 `moai codex` 의 의미]
두 해석이 모두 방어 가능하며, Implementation Kickoff Approval 에서 운영자가 정한다.

- **(a) 기동** — `moai cc` 가 claude 를 띄우듯 맨몸 `moai codex` 가 Codex CLI 를 띄우고, 준비 상태는 기동 직전 stderr 한 블록으로 흘린다. "moai cc 수준의 완성도" 라는 운영자 기준과 런처 패밀리 대칭에 부합. 본 SPEC 의 REQ-CL-002 는 현재 이 안으로 기술돼 있다.
- **(b) 확인** — 맨몸은 리드아웃만 출력하고 기동은 `moai codex cli` / `moai codex app` 으로 명시. 카드 원문("설정 상태 확인 + 앱/CLI 기동 **안내**")에 더 가깝고, 실수로 세션을 교체할 위험이 없다.

(a) 로 확정되면 REQ 수정 없음. (b) 로 확정되면 REQ-CL-002 한 줄과 관련 AC 2건만 바뀐다 — 나머지 설계는 두 안에서 동일하다.

## §C. 설계

### C.1 파일 배치

| 파일 | 성격 | 내용 |
|---|---|---|
| `internal/cli/codex_launcher.go` | 신규 | `codexCmd` 코브라 정의 (`GroupID: "launch"`), 세 동사 라우팅, `--spawn` 처리 |
| `internal/cli/codex_readiness.go` | 신규 | 리드아웃 조립 — `ProbeCodexSetup` + `codexwiring` + `CODEX_HOME` 해석을 한 구조체로 |
| `internal/cli/mcp_codex.go` | 수정 | auth 스트림 오독 수정 (§C.2) |
| `internal/cli/codex_launcher_test.go` 외 | 신규 | 아래 §D 시험 |

`unifiedLaunch` 는 건드리지 않는다 — Claude 전용 경로이고 (§A.6) codex 는 별도 exec 이다. 공유하는 것은 `spawnLaunch` 하나다.

### C.2 auth 스트림 수정 (핵심 결함)

현행:

```go
cmd.Stderr = &bytes.Buffer{}   // 버려짐
if err := cmd.Run(); err != nil { return "", err }
return out.String(), nil       // stdout 만
```

`run()` 의 계약을 바꾸면 `--version` 호출부까지 영향을 받으므로, **러너 인터페이스에 스트림을 합쳐 돌려주는 두 번째 메서드를 추가** 하고 `classifyCodexAuth` 만 그쪽을 쓴다.

- `codexCommandRunner` 에 `runCombined(ctx, bin, args, stdin) (string, error)` 추가 — stdout + stderr 를 순서대로 이어붙여 반환.
- `realCodexRunner.runCombined` 는 두 스트림을 각각 버퍼에 받아 결합한다 (`cmd.CombinedOutput` 은 쓰지 않는다 — t227 이 기록한 배너/JSON 혼입 함정과 같은 계열이며, 여기서는 결합 순서를 우리가 통제하는 편이 낫다).
- `classifyCodexAuth` 는 `runCombined` 를 호출하고, **rc 비영일 때도 캡처된 출력이 있으면 분류를 시도** 한다. 현행은 `err != nil` 이면 즉시 `unknown` 인데, 로그아웃 상태의 codex 가 비영 rc + 안내문을 낼 수 있다.
- `--version` 호출부는 `run()` 유지 — 계약 변경 0.

기존 테스트의 스텁 러너는 새 메서드를 구현해야 하므로, 스텁에 `runCombined` 를 추가하고 기본 동작을 `run` 위임으로 둔다 (기존 케이스 무변경).

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

- 맨몸 (또는 `cli`): `exec` 로 `codex` 를 프로젝트 루트에서 교체 실행. `--` 뒤 인자는 그대로 전달.
- `app`: `codex app` 위임.
- 어느 경로든 codex 바이너리 미해결 시 exec 전에 단일 진단으로 종료 (REQ-CL-012).

## §D. 마일스톤

| M | 내용 | 산출 |
|---|---|---|
| M1 | auth 스트림 수정 + 회귀 시험 | `runCombined` 추가, `classifyCodexAuth` 전환, 스텁 갱신 |
| M2 | CODEX_HOME 해석 + 리드아웃 조립 | `codex_readiness.go` + 단위 시험 |
| M3 | 커맨드 표면 + 동사 라우팅 + `--spawn` | `codex_launcher.go` + 등록 시험 |
| M4 | 도움말·예시 문안 + 중립성 통과 | help 텍스트, 필요 시 템플릿 문서 |

M1 은 독립적으로 가치가 있다 (`moai web` · MCP 도구의 오표시가 그 자체로 해소된다). M2→M3 는 순차, M4 는 M3 이후.

## §E. 위험

| 위험 | 완화 |
|---|---|
| codex 버전에 따라 `login status` 문구/스트림이 다시 바뀐다 | 결합 스트림을 읽으므로 스트림 이동에 둔감. 문구 변화는 `unknown` + 조치 안내로 안전 하강 (REQ-CL-009) |
| 스텁 러너 인터페이스 변경이 기존 codex 시험을 깬다 | 새 메서드는 추가만 — 기존 `run` 계약 무변경, 스텁 기본 동작을 위임으로 |
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
