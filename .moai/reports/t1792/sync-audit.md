auditor-model: claude-opus-5-5

# t1792 sync-audit — GH PR #1820 gomod-minor-patch 8개 모듈 갱신

verdict: PASS
audited_sha: d25e6bf9dbb0a4ec118ba3ef4a7ab17963b2e534
required_backend: codex=pass receipt=rcpt-58e600355a1d88fa8185dfe5,rcpt-396d0079c53fe96d7e305b04

- 카드: t1792 (Class A, SPEC 없음), 브랜치 `WT-gh-pr-1820`
- 감사 트리: `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1792`, HEAD `d25e6bf9d`, merge base `0f6060e0edb263945467132533f5deb85aa7d4d2`
- 감사 회차: 1회차 (이전 sync-audit 없음)
- 점수: Functionality 95 / Security 90 / Craft 90 / Consistency 95 → 조화평균 약 92.4, 필수 통과 차원(Functionality, Security) 모두 통과

## 교차 모델 감사 (audit plan)

`moai verify audit-plan --project-root <tree>` 결과: `config_status: ok`, `model: codex`, `cross_model_required: true`, `enforced_required: ["codex"]` (claude·glm 은 off).

`mcp__moai__audit_multi` (target=baseBranch, gates 인자 없음) 결과 다이제스트:

- `overall_verdict`: pass
- `gate_unmet`: (없음)
- `plan_source`: config
- `per_backend_verdicts`: backend=codex, gate=required, verdict=pass
- `audit_receipt`: rcpt-58e600355a1d88fa8185dfe5
- `build_commit`: ee981a932, `tree_root`: 위 감사 트리

이 판정은 오케스트레이터가 위 다이제스트로 `moai verify audit-plan --result-file` 확인을 통과시킨 뒤에야 최종이다.

## Claim / Evidence / Baseline-attribution

### C1. go.mod/go.sum 변경은 dependabot 8개 모듈과 그 간접 의존성 변동뿐이다

- 증거: dependabot 커밋 `fbeeb67be` 의 go.mod 증감 줄과 카드 전체(merge base..HEAD) go.mod 증감 줄을 비교 → `diff` 출력 없음, `IDENTICAL_GOMOD_DELTA`.
- 직접 의존성 8개: `charm.land/bubbletea/v2 v2.0.10→v2.1.0`, `github.com/a-h/templ v0.3.1020→v0.3.1070`, `github.com/mark3labs/mcp-go v1.1.1→v1.2.1`, `github.com/mattn/go-runewidth v0.0.30→v0.0.31`, `golang.org/x/net v0.59.0→v0.60.0`, `golang.org/x/sync v0.23.0→v0.24.0`, `golang.org/x/sys v0.48.0→v0.49.0`, `golang.org/x/tools v0.50.0→v0.51.0`.
- 간접 변동 4개: `andybalholm/brotli v1.2.1→v1.2.6`, `charmbracelet/ultraviolet`(pseudo-version 갱신), `charmbracelet/x/ansi v0.11.8→v0.11.9`, `xo/terminfo → v1.0.0`.
- go.sum: 위 12개 모듈만 각 2줄(-2/+2) 교체. 그 밖의 모듈 줄 변동 없음.
- 결론: 범위 외 의존성 변경 없음.

### C2. 생성된 `*_templ.go` 는 templ v0.3.1070 기준이며 재생성은 no-op 이다

- 증거: `grep -h 'templ: version' internal/web/*_templ.go | sort | uniq -c` → `8 // templ: version: v0.3.1070`. `go list -m github.com/a-h/templ` → `v0.3.1070`.
- 셸 PATH 의 `templ` 은 `v0.3.1020` 이라 쓰지 않았다. go.mod 고정 버전으로 `go build -o <scratch>/templ1070 github.com/a-h/templ/cmd/templ` 를 빌드(`templ1070 version` → `v0.3.1070`)한 뒤, `internal/web` 사본을 scratch 디렉터리에서 재생성:
  - `(✓) Complete [ updates=0 duration=125.817417ms ]`, exit 0
  - `diff -r internal/web <scratch>/web` → 출력 없음, `REGEN_NOOP`
- CI 단계(`.github/workflows/ci.yml:798` `go run github.com/a-h/templ/cmd/templ generate -path ./internal/web` + `git diff --exit-code`)와 같은 버전으로 판정했다.

### C3. `@` 이스케이프는 렌더링 결과를 바꾸지 않는다

- 변경: `internal/web/fieldsets.templ` 225행 `Model @ effort` → `Model &#64; effort`.
- 증거: 생성 코드 `internal/web/fieldsets_templ.go:793` 의 정적 문자열에 `<span data-i18n="sec.agenttiers.col.pair">Model &#64; effort</span>` 가 그대로 들어간다(이중 이스케이프 `&amp;#64;` 아님). merge base 의 같은 문자열은 `Model @ effort`. `&#64;` 는 HTML 숫자 문자 참조로 브라우저에서 `@` 로 표시되므로 사용자에게 보이는 텍스트는 같다. 바이트 단위 HTML 출력은 달라진다(아래 F2).

### C4. 영향 패키지에서 이 카드로 인한 회귀가 없다

- 범위: 갱신 모듈(templ, mcp-go, bubbletea, go-runewidth)을 직접 import 하는 패키지는 `go list` 로 `internal/cli`, `internal/cli/ptycaptest`, `internal/cli/update`, `internal/cli/wizard`, `internal/tui/internal`, `internal/web` 로 확인.
- 실행(스크립트 한 번, `TMPDIR=GOTMPDIR=/private/tmp/t1792s`, `GIT_CONFIG_GLOBAL=/private/tmp/t1792s/gc`, `MOAI_FACTORY_*` 5개 unset, `-count=1`):
  - `go test ./internal/web/... ./internal/tui/... ./internal/cli/wizard/... ./internal/cli/update/...` → exit 0. `ok internal/web 33.972s`, `ok internal/tui`, `ok internal/tui/golden`, `ok internal/tui/internal`, `ok internal/cli/wizard`, `ok internal/cli/update` 및 하위 5개 패키지 모두 ok.
  - `go build ./...` → exit 0, `go vet ./internal/web/...` → exit 0.
  - `go test -run 'TestCheckCodexWiring|TestCountCodexAgentTOMLs|MCP|Mcp' ./internal/cli/` → exit 1, 실패 10건(C5 참고). 그 밖의 실패 없음.

### C5. 레인의 기존 실패 귀속(A/B)은 타당하다

- `internal/hook` `TestNoNewEnvLiteralsInDiff`: 재실행 결과 `diff introduces env-name literals ... [MOAI_KANBAN_CARD]`. 이 테스트는 `internal/hook`, `internal/factorymsg`, `internal/cli` 의 비테스트 파일만 diff 하는데, 카드 diff(`--stat`)는 go.mod, go.sum, `internal/web/*`, `.moai/reports/t1792/*` 만 건드린다. `git log -S'MOAI_KANBAN_CARD' 0f6060e0e..HEAD` 는 보고서 커밋 `d25e6bf9d` 만 나온다(코드 커밋 없음). 레인이 지목한 `9c8447f9b`(card t1776)는 `git merge-base --is-ancestor 9c8447f9b main` 로 main 에 포함됨을 확인. → 카드와 무관.
- `internal/cli` 10건: A(카드 트리)와 B(`/Users/goos/MoAI/moai-adk-go/.claude/worktrees/main`, `go -C <main> test` 로 같은 `-run` 집합·같은 격리 환경) 모두 같은 10개(`TestCheckCodexWiring_*` 8개, `TestCountCodexAgentTOMLs`, `TestCountCodexAgentTOMLs_BadPatternRoot`)가 실패. 실패 메시지(`count on an unwired root = 6, want 0`)는 실제 사용자 홈의 codex 배치를 읽는 환경 누수 양상으로, 의존성 갱신과 무관하다. → 기존 실패이며 카드와 무관.

## Gaps (관측하지 않은 것)

- 전체 저장소 테스트는 지시대로 돌리지 않았다(CI 담당). `internal/cli` 는 `-run` 부분집합만 실행했다. 이 패키지는 mcp-go 를 import 하므로, MCP 서버 경로 중 `MCP|Mcp` 이름 패턴에 걸리지 않는 테스트는 관측하지 않았다.
- `internal/cli/ptycaptest` 는 실행하지 않았다.
- B 측 main 트리의 HEAD SHA 는 기록하지 못했다(`.git` 파일만 읽었다). 로컬 `main` 은 지금 `629bb51cd` 로 merge base 보다 앞서 있다. 다만 `git diff --stat 0f6060e0e 629bb51cd -- go.mod go.sum internal/web` 출력이 없어 카드 경로와 겹치지 않는다.
- `internal/cli` 10건의 근본 원인은 조사하지 않았다(범위 밖, 기존 실패).
- govulncheck 등 취약점 스캔은 돌리지 않았다. 보안 판단은 신규 모듈이 없고 모두 기존 모듈의 minor/patch 갱신이라는 go.mod 관측에 근거한다.
- 원격 PR #1820 의 CI·CodeRabbit 은 읽지 않았다(운영자 보유 범위).
- `@` 를 이스케이프하지 않으면 새 파서가 실패한다는 커밋 메시지 주장은 직접 재현하지 않았다. 이 판정에는 영향이 없다.

## Residual-risk (관측했어도 남는 위험)

- bubbletea v2.1.0, mcp-go v1.2.1 은 minor 갱신이다. 실행한 테스트가 다루지 않는 런타임 동작(TUI 렌더링 세부, MCP 프로토콜 경계)의 변화는 CI 전체 실행에서만 드러날 수 있다.
- `internal/web` 테스트(33.97s)는 이전 레인 기록에 부하에 민감한 서버 flake 가 한 번 있었다. 이번 실행은 통과했다.
- main 이 계속 전진하므로 착지 시점에 go.mod/go.sum 충돌이 생기면 재생성·재검증이 필요하다.

## Findings

- F1 [Low] [optional] `internal/cli` doctor/codex 테스트 10건과 `internal/hook` `TestNoNewEnvLiteralsInDiff` 가 main 에서도 실패한다. 이 카드 탓은 아니지만 CI 를 빨갛게 만들 수 있다. 필요한 조치: 별도 카드로 분리(환경 누수 격리, `MOAI_KANBAN_CARD` 를 envkeys 상수로 이동). 이 카드에서는 고치지 않는다. 신뢰도 높음.
- F2 [Info] [optional] `internal/web/fieldsets.templ:225` 의 출력 HTML 바이트는 `@` 에서 `&#64;` 로 바뀐다. 표시 텍스트는 같다. 서버 쪽 HTML 을 바이트 그대로 비교하는 골든/스냅샷 검사가 있다면 영향을 받을 수 있으나, `internal/web` 테스트는 통과했다. 필요한 조치: 없음. 신뢰도 높음.
- F3 [Info] [optional] 셸 PATH 의 `templ` 은 v0.3.1020 이다. 로컬에서 맨 `templ generate` 를 실행하면 이전 버전 헤더로 되돌아간다. 필요한 조치: 재생성 시 `go run github.com/a-h/templ/cmd/templ` 를 쓰거나 설치본을 v0.3.1070 으로 갱신. 저장소 변경은 필요 없다. 신뢰도 높음.

차단(blocking) 결함 없음.

## 차원별 점수

| 차원 | 점수 | 판정 | 근거 |
|---|---|---|---|
| Functionality (40%) | 95 | PASS | 재생성 `updates=0` + `REGEN_NOOP`, 영향 패키지 `ok` 11개, build/vet exit 0 |
| Security (25%) | 90 | PASS | 신규 모듈 없음, minor/patch 8개 + 간접 4개만 변경(go.mod delta가 dependabot 커밋과 동일). 취약점 스캔은 Gap |
| Craft (20%) | 90 | PASS | 생성 코드 버전 헤더 8/8 v0.3.1070, 이스케이프 최소 변경(1줄) |
| Consistency (15%) | 95 | PASS | 생성 파일은 templ 출력 그대로, 커밋 메시지에 card id 포함 |
