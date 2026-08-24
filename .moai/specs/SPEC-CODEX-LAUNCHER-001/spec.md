---
id: SPEC-CODEX-LAUNCHER-001
title: "Codex 전용 런처 — moai codex: 배선·CODEX_HOME·auth 상태 확인과 앱/CLI 기동"
version: "0.2.0"
status: draft
created: 2026-08-24
updated: 2026-08-24
author: manager-spec
priority: P1
phase: "v3.2 target"
module: internal/cli
lifecycle: spec-anchored
era: V3R6
tier: M
tags: "codex, launcher, codex-home, auth, dual-harness, cli"
depends_on: [SPEC-CODEX-WIRING-001]
related_specs: [SPEC-CODEX-DUAL-AGENTS-001, SPEC-CODEX-SKILLS-CANONICAL-001, SPEC-CODEX-HOOK-ADAPTER-001]
---

# SPEC-CODEX-LAUNCHER-001 — Codex 전용 런처

## HISTORY

| 버전 | 날짜 | 변경 |
|---|---|---|
| 0.1.0 | 2026-08-24 | 최초 작성 (plan-phase, 카드 t197). 리드 배차 시 추가된 운영자 기준 3항(런처 동사 / auth 상태 노출 / t88 M4 정합) 반영 |
| 0.2.0 | 2026-08-24 | 맨몸 `moai codex` 의미 확정 — 리드 판정 (b) "리드아웃 + 명시 기동". REQ-CL-002 재기술, AC-CL-002 조건절 해소, plan §B 결정 기록으로 전환 |

## §A. 측정 전제 (Verified baseline)

> 근거: `.moai/reports/t197/measurement.md` (worktree `WT-codex-launcher` @ `9280c96b3`, 2026-08-24 실측).
> t88 (M4) 산출물 `7b217da7c` 가 이 트리의 조상임을 확인했다.

### §A.1 `moai codex` 는 없다 (카드 전제 성립)

빌드한 바이너리의 `--help` 에 LAUNCH COMMANDS 는 `cc` / `glm` / `cg` 셋뿐이고, `codex` 를 `Use:` 로 갖는 최상위 커맨드는 0건이다 (`codex-review-gate` 는 `moai hook` 하위 커맨드). Codex 사용자는 moai 가 깐 배선이 실제로 살아 있는지 확인할 진입점이 없다.

### §A.2 auth 상태가 항상 `unknown` — 원인은 스트림 오독

로컬은 로그인된 상태(`codex login status` → `Logged in using ChatGPT`, rc 0)인데 `codex_setup` 프로브는 `auth_provider: unknown` 을 반환한다. 스트림을 갈라 재면 이 문구는 **stdout 0바이트 / stderr 24바이트** 로, 전량 stderr 로 나간다. `realCodexRunner.run` (`internal/cli/mcp_codex.go:273`) 은 `cmd.Stderr = &bytes.Buffer{}` 로 stderr 를 버리므로 `classifyCodexAuth` (:1325) 는 빈 문자열을 받고 fail-open 으로 `unknown` 을 낸다.

분류 패턴의 문제가 아니라 **읽는 스트림이 틀렸다**. 같은 오독이 `moai web` 콘솔의 Codex 카드에도 그대로 나타난다 (`internal/web/codex_state.go` — `AuthProvider == "unknown"` 분기).

### §A.3 codex CLI 가 이미 진단을 제공한다

`codex doctor` 는 설치·런타임·디스크·git·`CODEX_HOME` 여유 공간까지 진단하고, `codex app` 은 데스크톱 앱을 (없으면 설치 관리자를) 띄운다. 따라서 moai 런처가 재구현할 영역은 **moai 쪽 프로젝트 배선 상태** 로 한정된다.

### §A.4 moai 쪽 Codex 배선 표면 (t88 M4)

| 표면 | 위치 |
|---|---|
| 생성기 | `moai init --agent claude\|codex\|both` (`internal/cli/init.go:393-400`) |
| 산출물 | `.codex/hooks.json`, `.codex/config.toml` (`internal/codexwiring/codexwiring.go:29,31`), 신뢰 사이드카 |
| 에이전트 | `.codex/agents/moai/*.toml` 12종 (템플릿) |
| 진단 | `moai doctor` 의 `Codex Wiring` 체크 (`internal/cli/doctor_codex.go`) |
| 런타임 | `moai hook --harness codex` (`internal/cli/hook_harness_codex.go`) |

이 저장소 자체에는 `.codex/` 가 없다 — 배선 미적용이 기본 상태다.

### §A.5 `CODEX_HOME` 을 읽는 런타임 코드는 0건

`grep -rn "CODEX_HOME" internal/ --include="*.go"` → 0건. 문서·SPEC·`settings.local.json` 허용 목록에만 존재하는, 런처가 새로 도입해야 할 표면이다.

### §A.6 기존 런처 3종의 공통 경로는 재사용 불가

`runCC` / `runCG` / `runGLM` 은 모두 `unifiedLaunch` 로 수렴하고 그 본체는 `.claude/settings.local.json` 변형 + Claude 프로필 해석 + `claude` exec 이다. codex 는 다른 바이너리이므로 이 경로에 얹을 수 없다. 공유 가능한 것은 `--spawn` (tmux 새 창) 정도다.

## §B. 문제 진술

t88 (M4) 이 Codex 쪽 배선을 깔았지만, 그 배선을 **확인하고 그 위에서 Codex 를 띄우는 진입점** 이 없다. 운영자는 (1) 배선이 살아 있는지, (2) `CODEX_HOME` 이 어디를 가리키는지, (3) 로그인이 되어 있는지를 각각 다른 도구로 확인해야 하고, 그중 (3) 은 moai 가 보여주는 값 자체가 틀려 있다.

## §C. 범위

### 포함

- `moai codex` 최상위 커맨드 (LAUNCH COMMANDS 그룹, `cc` / `glm` / `cg` 의 형제)
- 준비 상태 리드아웃: 바이너리 · 버전 · `CODEX_HOME` · auth · 프로젝트 배선 (`.codex/*`)
- auth 분류의 스트림 오독 수정 (§A.2) — 프로브를 공유하는 세 표면 전부에 반영
- Codex CLI / 데스크톱 앱 기동 위임
- `--spawn` (tmux 새 창) 패리티

### Out of Scope (제외)

- `codex doctor` 가 이미 하는 진단의 재구현 (위임한다)
- `.codex/` 배선 생성·수리 — 생성은 `moai init --agent`, 표류 보고는 `moai doctor` 소관 (SPEC-CODEX-WIRING-001)
- Codex 쪽 로그인 수행 (`codex login` 위임; moai 는 자격 증명을 만들거나 옮기지 않는다)
- Kanban / Factory 진입 토큰 (`-k` / `-f`) — Claude 세션 모델 전제이므로 이번 범위 밖
- `moai cg` 형 하이브리드 (Claude 리드 + Codex 팀메이트)

## §D. 요구사항 (GEARS)

### D.1 커맨드 표면

- **REQ-CL-001** — The system shall provide a top-level `moai codex` command registered in the `launch` command group, so it appears alongside `cc` / `glm` / `cg` in `moai --help`.
- **REQ-CL-002** — The bare `moai codex` command shall print the readiness readout and exec nothing; launching shall require an explicit verb — `cli` (Codex CLI in the current project directory) or `app` (desktop app). `status` shall be accepted as an explicit alias of the bare readout form.
- **REQ-CL-003** — Where the operator passes `--spawn`, the system shall run the launch in a new tmux window instead of replacing the current process, matching the `moai cc --spawn` contract, and shall fail with the same diagnostic when tmux is absent.

### D.2 준비 상태 리드아웃

- **REQ-CL-004** — The readiness readout shall report, as discrete rows: codex binary path, codex version, resolved `CODEX_HOME`, auth provider, and project wiring state (`.codex/hooks.json` + `.codex/config.toml` presence and whitelist validity).
- **REQ-CL-005** — The system shall resolve `CODEX_HOME` from the `CODEX_HOME` environment variable, falling back to `~/.codex`, and shall report which of the two supplied the value.
- **REQ-CL-006** — Where the project carries no `.codex/` wiring, the readout shall report it as an informational state (not an error) and name `moai init --agent codex` as the action, mirroring the fail-open stance of the `moai doctor` Codex Wiring check.
- **REQ-CL-007** — The readiness readout shall consume the existing `ProbeCodexSetup` / `codexwiring` surfaces and shall not fork a second auth-classification or wiring-validation implementation.

### D.3 auth 상태 (§A.2 결함 수정)

- **REQ-CL-008** — The auth classifier shall read the combined stdout and stderr of `codex login status`, so a codex build that writes the status line to stderr classifies correctly instead of degrading to `unknown`.
- **REQ-CL-009** — Where the classification still cannot be determined, the system shall report `unknown` together with the action `codex login status` rather than asserting a logged-out state — an unreadable probe is a gap, never a verdict.
- **REQ-CL-010** — The stream-reading correction shall apply to every consumer of the shared probe (the `codex_setup` MCP tool, the web console Codex card, and this launcher), with no second classification path introduced.

### D.4 기동 위임

- **REQ-CL-011** — The `app` verb shall delegate to `codex app` rather than reimplementing desktop-app discovery or installation.
- **REQ-CL-012** — Where the codex binary is absent from PATH, the launch verbs (`cli`, `app`) shall fail with a single diagnostic naming the install action and shall exec nothing; the readout form shall still succeed, reporting the binary row as not found — a diagnostic that refuses to run when the thing it diagnoses is missing is useless exactly when it is needed.
- **REQ-CL-013** — The system shall not mutate `.claude/settings.local.json`, Claude profile state, or any file under `CODEX_HOME` on any verb — the launcher reads state and execs; it does not write.

### D.5 배포 표면

- **REQ-CL-014** — Help text, examples, and any template-side documentation shall stay language-neutral and free of internal identifiers, satisfying the template neutrality guard.

## §E. 비기능 요구

- 진단 경로는 fail-open: 어느 프로브가 실패해도 런처는 그 행을 `unknown` 으로 적고 나머지를 계속 보고한다.
- 리드아웃은 bounded — 화면 한 폭 안에 들어가는 고정 행 집합이며, `codex doctor` 급 전체 덤프를 내지 않는다.
- 크로스 플랫폼: 경로 해석은 `os.UserHomeDir` 기반이며 macOS 편향 경로를 하드코딩하지 않는다.

## §F. 성공 판정

맨몸 `moai codex` 가 로그인된 머신에서 `auth: chatgpt` 를 보고하고 (현행은 `unknown`), 배선 없는 프로젝트에서 오류 없이 `wiring: not wired` + 조치 문구를 보고하며 아무것도 exec 하지 않는다. 기동은 `moai codex cli` 가 프로젝트 디렉터리에서 Codex CLI 를, `moai codex app` 이 데스크톱 앱을 띄운다.
