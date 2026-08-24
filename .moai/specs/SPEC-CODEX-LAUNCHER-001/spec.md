---
id: SPEC-CODEX-LAUNCHER-001
title: "Codex 전용 런처 — moai codex: 배선·CODEX_HOME·auth 상태 확인과 앱/CLI 기동"
version: "0.7.0"
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
| 0.3.0 | 2026-08-24 | 교차모델 감사 1차(codex 백엔드) 지적 4건 반영. auth 분류를 `auth_mode` 파일 1순위 + 앵커된 긍정 행 2순위의 2단 사다리로 재설계(REQ-CL-008/009 재기술) — 부분 일치가 오류 문구를 인증 성공으로 읽는 결함을 설계 단계에서 제거. 공유 러너 인터페이스 확장 철회(REQ-CL-010 후단) — 구현체 수 측정이 틀렸음이 드러남(M-7). AC 5건을 명목 커버리지에서 행동 판정으로 교체. 근거 문서를 축약 없는 실측본으로 재작성 |
| 0.4.0 | 2026-08-24 | 교차모델 감사 2차 지적 5건 반영. (1) `^logged in` 앵커도 뚫린다는 반례(`Logged in state unavailable: API key missing`)를 받아 전체 행 문법 + 캡처값 매핑으로 교체(REQ-CL-009). (2) 파일 1단에 모드별 최소 구조 조건 추가 — 자격 재료 없는 stale 파일은 하강(REQ-CL-008), 비밀값 규율을 출력 grep 에서 타입(비밀 필드 없는 구조체)으로 격상. (3) seam 을 3분해 — 최종 provider 만 반환하는 단일 seam 으로는 stdout/stderr/rc 주입 지점이 없어 핵심 회귀 시험이 명목화됨. (4) AC-CL-002 에 cwd·argv 전달 판정, AC-CL-007 에 sentinel 전파 판정 추가. (5) 근거 문서 재작성 + 에이전트 TOML 수를 12→**11** 로 정정(`ls | wc -l` 이 별칭 긴 형식의 `total` 행을 함께 셌다) |
| 0.5.0 | 2026-08-24 | 3차 감사 지적 5건 반영 + 운영자 신규 요구 접기. (1) 파일 1단이 `tokens:{}` 를 통과하던 구멍 폐쇄 — '존재' 가 아니라 '비어 있지 않은 값' 요구(REQ-CL-008). (2) 비밀값 규율을 실제로 타입에 걸었다 — 앞선 안은 `APIKey string` 으로 키 전문을 역직렬화하면서 '비밀 필드 없음' 이라 선언했다. (3) 실행 seam 이 결합된 바이트를 반환해 프로덕션 stderr 결함이 회귀 시험을 우회하던 문제 — 두 스트림을 분리 반환하고 결합 규칙 자체를 시험 대상으로. (4) 작성 불가능했던 AC 해소 — `classifyCodexAuthFile` 에 `err` 추가, 오류 문안 판정은 경로를 아는 층으로 이동. (5) 근거 문서를 전사(transcript) 방식으로 전환 — 스크립트와 그 출력을 커밋하고 문서는 줄 범위를 인용한다. 신규: 미배선 프로젝트 초기화 제안 + `AGENTS.md`↔`CLAUDE.md` 지시 계약(REQ-CL-015/016). AC 2건 통합(공유 러너 무회귀→게이트, 앱 위임→동사 라우팅)으로 Tier M 상한 16/16 유지 |
| 0.6.0 | 2026-08-24 | 최종 확인 감사 지적 3건 반영. (1) `nonEmpty` 판정을 원문 바이트 비교에서 **JSON 타입** 판정으로 교체 — 감사 실측이 `{ }`·`false`·`0`·`[]` 가 전부 '비어 있지 않음' 으로 통과함을 보였다. 자격 재료는 비어 있지 않은 JSON 문자열이어야 하고 다른 타입은 전부 부재(REQ-CL-008), AC 표에 12행 추가. (2) 근거 스크립트를 자기완결로 — 측정 대상 바이너리와 doctor JSON 을 스스로 만들고, 선행 조건 실패 시 중단하며, 인용하는 모든 수치를 전사본 안에서 잡는다. 'read-only' 주장 철회(빌드 + codex 의 PATH 별칭 시도는 실제 변경). (3) 배선 판정을 디렉터리 존재에서 **파일 집합** 으로 — 빈 `.codex/`·한쪽 파일만 있는 상태를 미배선으로 판정(REQ-CL-006/015), AC 에 5종 상태 행렬. 지시 계약은 import 줄 형태(`@AGENTS.md`)를 고정하고 멱등을 2회 실행 바이트 비교로 판정하며 로컬 지시 파일을 `CLAUDE.local.md` 로 명명(REQ-CL-016) |
| 0.7.0 | 2026-08-24 | 범위 조정 — 운영자 판정. 초기화 요구(구 REQ-CL-015/016 + AC)를 `SPEC-CODEX-INIT-001` 로 분리하고 14 REQ 로 복귀. 풀린 예산으로: 상한 맞추려 통합했던 AC 2건 복원(공유 러너 무회귀·앱 위임), REQ-CL-011 명시 참조 회복(14/14 전수), 4차 감사 커버리지 지적 중 런처 소속 흡수 — `tokens` **키 의미 검증**(무관 키·계정 메타데이터는 자격 재료 아님, 원시 타입 거부만으로는 `{"irrelevant":"x"}` 가 통과)과 리드아웃 배선 6상태 행렬. 산문-전사본 드리프트 정정(낡은 핀 `6bfb076bc`, 46초·롤아웃 31,525 잔존 수치) |

## §A. 측정 전제 (Verified baseline)

> 근거: `.moai/reports/t197/` — 자기완결 측정 스크립트 `probe.sh`, 그 1회 실행 무편집 전사본 `probe-output.txt` (**측정 대상 트리 `1ed61e4ac`** — 전사본 L24-26 이 스스로 찍은 값), 그리고 전사본의 줄 범위를 인용하며 해석하는 `measurement.md`. 미관측 항목은 그 문서 말미에 명시돼 있다.
>
> 인용은 `citation-sweep.sh` 가 **판정한다** — 통과면 rc 0, 드리프트면 rc 1. 다섯 가지를 본다: 산문의 커밋 핀이 전사본이 스스로 찍은 핀과 같은가, 인용한 줄 범위가 전사본 길이 안에 있는가, 그 인용이 `citation-manifest.txt` 의 행에 묶여 있는가, **그 줄 범위가 실제로 그 주장의 근거를 담고 있는가**, 그리고 더 이상 인용되지 않는 manifest 행이 남아 있지 않은가. 네 번째가 핵심이다 — 줄 번호는 맞는데 그 줄이 주장과 무관한 형태를 잡는다.
> 산문이 그 근거의 **올바른 해석인지** 는 판정하지 않는다. 그건 사람이 읽는다.
> 이 게이트 자체는 `gate-selftest.sh` 가 검증한다 — 다섯 검사마다 통과해서는 안 되는 입력을 하나씩 주입해 실제로 rc 1 을 관측한다. 주입은 작업 트리가 아니라 **코퍼스 사본**(`MOAI_CITATION_ROOT`)에 하고, 실행 후 `git status --porcelain` 이 실행 전과 같은지로 트리 무변경을 확인한다 — 스크립트가 기억하는 파일 목록이 아니라 git 에게 묻는다.
> t88 (M4) 산출물 `7b217da7c` 가 이 트리의 조상임을 확인했다.

### §A.1 `moai codex` 는 없다 (카드 전제 성립)

빌드한 바이너리의 `--help` 에 LAUNCH COMMANDS 는 `cc` / `glm` / `cg` 셋뿐이고, `codex` 를 `Use:` 로 갖는 최상위 커맨드는 0건이다 (`codex-review-gate` 는 `moai hook` 하위 커맨드). Codex 사용자는 moai 가 깐 배선이 실제로 살아 있는지 확인할 진입점이 없다.

### §A.2a 구조화된 auth 원천이 존재한다

`codex doctor` 는 auth 를 구조화해 알고 있고 (`stored auth mode: chatgpt`), 그 원천은 `<CODEX_HOME>/auth.json` 의 `auth_mode` 필드다. doctor 자체는 커밋된 전사본에서 **67초** 걸려 런처의 대화형 리드아웃에 쓸 수 없지만, 파일은 즉시 읽힌다. 산문 파싱보다 이쪽이 1순위다.

다만 이 머신에서 관측한 조합은 `auth_mode=chatgpt` 하나뿐이다 — `apikey` / `provider` 모드의 실제 파일 형태는 미관측이므로, 설계는 알려지지 않은 값을 추측하지 않고 명령 프로브로 하강한다.

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
| 에이전트 | `.codex/agents/moai/*.toml` 11종 (템플릿) |
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
- auth 분류 재설계 (§A.2 스트림 오독 + 부분 일치 오분류) — 프로브를 공유하는 세 표면 전부에 반영
- Codex CLI / 데스크톱 앱 기동 위임
- `--spawn` (tmux 새 창) 패리티

### Out of Scope (제외)

- `codex doctor` 가 이미 하는 진단의 재구현 (위임한다)
- 미배선 프로젝트의 **초기화** — 제안·생성기 호출·지시 계약은 `SPEC-CODEX-INIT-001` 로 분리됐다. 런처는 배선 상태를 **판정하고 보고** 하는 데까지다
- 배선 생성 로직과 이미 깔린 배선의 표류 수리 — `moai init --agent` / `moai doctor` 소관 (SPEC-CODEX-WIRING-001)
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
- **REQ-CL-006** — Where the project's `.codex/` wiring is incomplete — the directory absent, present but empty, or missing either generated file — the readout shall report it as an informational state (not an error) and name `moai init --agent codex` as the action, mirroring the fail-open stance of the `moai doctor` Codex Wiring check. The action shall accompany every incomplete state, not only an absent directory, and shall be absent when the wiring is complete. This test is the single definition of wiring completeness; `SPEC-CODEX-INIT-001` consumes it rather than restating it.
- **REQ-CL-007** — The readiness readout shall consume the existing `ProbeCodexSetup` / `codexwiring` surfaces and shall not fork a second auth-classification or wiring-validation implementation.

### D.3 auth 상태 (§A.2 결함 수정)

- **REQ-CL-008** — The auth classifier shall determine the provider from `<CODEX_HOME>/auth.json` only when the file's `auth_mode` is a known value AND the credential material that mode implies is present under a key that mode recognizes, as a non-empty JSON string — account metadata under an unrecognized key is not credential material; any other JSON type — a boolean, a number, an array, an object, or null — shall count as absent, as shall an empty or whitespace-only string and a container holding no such value. In every other case — an unknown mode, empty or missing credential material, an unparseable file, or no file at all — it shall fall back to the command probe rather than report a provider. It shall deserialize through types that record only whether each credential field was non-empty, never the value itself, so no credential is retained, logged, or wrapped into an error.
- **REQ-CL-009** — When classifying from command output, the system shall accept a provider only from a status line matching a fixed whole-line grammar, mapping the captured provider term and nothing else; a line that merely contains a provider term shall never classify, whatever it starts with. It shall classify a non-zero exit only when such a line is present, and shall otherwise report `unknown` together with the action `codex login status` — an unreadable probe is a gap, never a verdict.
- **REQ-CL-010** — The classification correction shall apply to every consumer of the shared probe (the `codex_setup` MCP tool, the web console Codex card, and this launcher), with no second classification path introduced and no change to the shared `codexCommandRunner` interface.

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
