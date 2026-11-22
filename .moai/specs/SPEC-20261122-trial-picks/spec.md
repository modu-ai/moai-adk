---
id: SPEC-20261122-trial-picks
title: "Six Claude Code trial options are offered in the project and update previews, each off until the user picks it, each resting on a documented line"
status: draft
revision: 1
created: 2026-11-22
parent: null
---

# Trial options behind the user's pick

## Intent

Claude Code 공식 문서 전체 읽기 보고서(2026-11-22) 16장의 28~33번 여섯 가지를 "시험 선택지(TRIAL)"로 내놓아요. 여섯 모두 기본은 꺼짐이고, `/moai project` 나 `moai update` 의 미리보기에 한 줄씩 보이며, 사용자가 고른 것만 적용해요. 밖으로 데이터를 보내는 일(OTel)은 매번 사용자가 골라요. 이 조각은 SPEC 만 써요.

여섯 가지와 moai 가 하는 일:

1. 계획 모드의 계획 파일 자리(`plansDirectory`): 프로젝트 설정에 프로젝트 안 상대 경로를 적어요. 계획 파일은 압축 뒤에 다시 붙어서 긴 작업이 계획을 잃지 않아요.
2. 자동 모드 분류기의 `autoMode.hard_deny` 에 NeXT 한 줄: 통합 브랜치를 착지 명령 밖에서 옮기거나 지우지 않는다. 이 키는 사용자·관리 설정이나 `--settings` 에서만 읽혀서, 고른 설정 폴더의 사용자 설정에 적어요. 기본 규칙을 지키려고 `$defaults` 를 함께 적어요.
3. OTel 자원 속성 `moai.spec`, `moai.lane`: moai 가 띄우는 세션의 환경에 속성만 더해요. 내보내기를 켜는 변수, 주소, 내용 기록 변수는 moai 가 절대 설정하지 않아요. 내보내기는 사용자가 이미 켠 경우에만 일어나요.
4. `claude plugin eval` 로 `/moai` 스킬 회귀 평가: 고를 때만 돌리고, 보고서 게시를 막는 `--no-publish` 와 비용 상한을 붙여요.
5. 보안 플러그인(`security-guidance`) 제안: 웹·보안에 민감한 프로젝트의 하네스 미리보기에 내놓아요. 설치와 비용(턴마다·커밋마다 Opus 검토)을 미리 말해요. `/claude-security` 는 필요할 때 부르는 명령이라는 안내 한 줄만 둬요.
6. 공식 문서 MCP 서버를 moai-research 용으로 추가: 고른 설정 폴더마다 문서의 `claude mcp add` 명령을 돌려요. moai 가 `.claude.json` 을 직접 쓰지 않아요.

리더의 적용 범위 노트(A39)가 더한 다른 TRIAL 항목들(instructionFiles, 하위 프로세스 환경 정리, 업데이트 채널과 최소 판, StopFailure·PreModelSwitch, REPL 매처, `paths:` SPEC 도우미 스킬, 프로젝트 verify 스킬, Remote Control 알림, 채널, 관리형 Code Review, devcontainer, 화면 읽기 글, cacheTtl, 비용 항목)은 다음 개정에서 같은 틀에 행으로 더해요.

### Evidence (공식 문서 원문, 2026-11-22T07:33Z~07:36Z 읽음, `https://code.claude.com/docs/en/<page>.md`)

| 항목 | 인용 | 출처 |
| --- | --- | --- |
| `plansDirectory` | "Choose where Claude Code stores the plan files it writes in plan mode. Claude Code resolves the path relative to the project root." / Scope "Any file" / Default unset, so Claude Code uses the `plans` folder of the user's config folder / 프로젝트 밖으로 풀리는 경로(`"../plans"`)는 기본 자리로 돌아가요 | settings-reference.md L2595-2610 |
| 계획 파일과 압축 | "Claude Code re-injects the plan file after each compaction, so the plan survives where conversation history may not" | large-codebases.md L472 |
| `hard_deny` | "`hard_deny` rules block unconditionally. User intent and `allow` exceptions don't apply." / 각 규칙은 산문 / "Include `"$defaults"` in every list you set" | auto-mode-config.md L231, L223, L269 |
| `autoMode` 자리 | "The classifier doesn't read `autoMode` from project settings in `.claude/settings.json` or `.claude/settings.local.json`." / Scope "User or managed"; `--settings` 도 읽음 | auto-mode-config.md L72, L64-70; settings-reference.md L854 |
| 분류기의 한계 | "Auto mode reduces permission prompts but does not guarantee safety." / 패턴 거절은 `permissions.deny` 가 분류기보다 먼저 | permission-modes.md L338; auto-mode-config.md L57 |
| OTel 켜기 | "OpenTelemetry export to your backend is opt-in and requires explicit configuration." / `CLAUDE_CODE_ENABLE_TELEMETRY` "(required)" | monitoring-usage.md L1849, L110 |
| 자원 속성 | `OTEL_RESOURCE_ATTRIBUTES` 는 쉼표로 나눈 `key=value`, 공백 금지, 특수 문자는 퍼센트 인코딩, 값은 펼치지 않음 / "attaches these values as attributes on every metric datapoint and event record" / 내장 키를 덮지 않음 | monitoring-usage.md L445-495, L458 |
| 저장소 설정의 OTel | "Claude Code ignores the OpenTelemetry exporter variables in a repository's `.claude/settings.json` and `.claude/settings.local.json`" / 하위 프로세스에 `OTEL_*` 를 넘기지 않음 | monitoring-usage.md L69, L71 |
| 내용 기록 | 프롬프트·도구 입력·출력·API 본문 기록은 각자의 `OTEL_LOG_*` 변수로만 켜지고 기본은 꺼짐 | monitoring-usage.md L124-131 |
| plugin eval | "Every eval run and every judge grader is a real model call on your account" / 기본 세 번 실행과 플러그인 없는 비교 실행 / claude.ai 로그인이면 보고서를 비공개 아티팩트로 게시, "Pass `--no-publish` to keep it local." / v2.1.269 이상 / 대상은 매니페스트가 있는 플러그인 폴더나 skills-directory 플러그인 | plugin-evals.md L23, L44-46, L448, L29, L31 |
| security-guidance | 편집마다 패턴 검사(모델 호출 없음), 턴 끝 모델 검토, 커밋·푸시마다 에이전트 검토 / "By default, both reviews run on a Claude Opus model." / "None of the layers block writes or commits." / Python 3.7+(커밋 검토 3.10+), 첫 실행에 `pip` 와 네트워크 / 설치 `claude plugin install security-guidance@claude-plugins-official --scope project` 와 `enabledPlugins` | security-guidance.md L63-67, L174, L110, L17-20, L49-57 |
| claude-security | "Patches are never applied automatically." / 스캔마다 사용량, 유료 플랜이나 API 필요 | claude-security.md L90, L13, L21 |
| 문서 MCP | `claude mcp add --scope user --transport http claude-code-docs https://code.claude.com/docs/mcp` / 인증 필요 없음 / 사용자 범위는 `.claude.json` 의 `mcpServers`, `CLAUDE_CONFIG_DIR` 가 있으면 그 폴더 안 | mcp-quickstart.md L145, L28, L126-128 |

확인되지 않은 것: `plansDirectory` 의 "project root" 가 git 루트인지 시작 폴더인지, 사용자 지정 자리의 계획 파일도 압축 뒤 다시 붙는지(문서는 "the plan file"), 계획 모드에서 분류기가 `hard_deny` 를 쓰는지, "skills-directory plugin" 이 매니페스트 없는 스킬 폴더를 뜻하는지, 문서 MCP 서버의 도구 이름(문서에는 없고 실측으로만 `search_claude_code_docs`, `query_docs_filesystem_claude_code_docs`). 각 코드 조각이 해당 줄을 쓰기 전에 실측해요.

## Non-goals

- 기본으로 켜는 것: 여섯 모두 꺼진 채로 나가요.
- OTel 내보내기 자체(수집기, 주소, 헤더, 내용 기록): 사용자의 몫이고 moai 는 안내 글만 보여요.
- 플러그인과 언어 서버의 설치 엔진: 설치 명령을 돌리는 것까지예요. `enabledPlugins` 값 검사는 규모 설정 SPEC 의 쓰기 검사를 써요.
- `.claude.json` 쓰기: 문서의 명령이 써요.
- A39 의 나머지 TRIAL 항목: 다음 개정.

## Requirements

- REQ-1: The project and update previews shall list the six trial options, each unpicked by default, each with its documented gate, its cost, its risk and the source line it rests on.
- REQ-2: When an apply runs without a trial pick, the apply shall write nothing and run nothing for any trial option. · failure: none — a preview that lists the options changes no file.
- REQ-3: When the user picks plan files, the plan shall set plansDirectory in the project settings to a folder inside the project. · failure: none — a path that leaves the project or holds a backslash is refused before any write, because Claude Code would fall back to its default place.
- REQ-4: When the user picks the NeXT classifier rule, the plan shall add one hard_deny line that forbids moving, resetting, force-pushing or deleting the recorded integration branch outside the landing command, together with the $defaults entry, to the user settings of each config folder the user picks. · failure: none — no project settings file receives an autoMode key, and a hard_deny list that already holds the line is left as it is.
- REQ-5: When the user picks the OTel attributes, the sessions moai starts shall carry moai.spec and moai.lane as percent-encoded entries appended to the user's own OTEL_RESOURCE_ATTRIBUTES value. · failure: none — moai sets no variable that turns telemetry on, chooses an exporter, an endpoint or headers, or logs prompt, tool or response content.
- REQ-6: The OTel pick shall never carry over: every new preview offers it unpicked, and an apply without the pick leaves no moai attribute in place.
- REQ-7: When the user picks the skill evaluation, moai shall run claude plugin eval on the shipped suite with --no-publish and the cost ceiling the user gave. · failure: none — a run without a ceiling or with a Claude Code older than 2.1.269 is refused before any model call, and publishing the report needs its own pick.
- REQ-8: When the user picks the security plugin, the plan shall list its enabledPlugins entry for the scope the user picks and the install command, and say that the reviews run on an Opus model each turn and each commit, need Python and never block a write. · failure: none — the install command runs only after the pick, and a project without git is told that the reviews skip there.
- REQ-9: When the user picks the documentation server, moai shall run the documented claude mcp add command at user scope with CLAUDE_CONFIG_DIR set to each config folder the user picks. · failure: none — a folder that already holds a server of that name keeps it and the preview names it, and moai writes no .claude.json itself.
- REQ-10: Every trial row and every apply answer shall name the documentation page and line it rests on and the date the line was read.

## Acceptance

- AC-1 (REQ-1, REQ-2, REQ-10): in a temporary project and an empty temporary HOME, both previews list six rows, each unpicked, each with gate, cost, risk and source; an apply with no pick leaves every file byte-identical and starts no process.
  verify: `cargo test --locked --offline --test trial_picks_cli list_`
- AC-2 (REQ-3): the plan-files pick writes `plansDirectory` with a project-relative path; `../plans` and a backslash path are refused before any write.
  verify: `cargo test --locked --offline --test trial_picks_cli plans_`
- AC-3 (REQ-4): the classifier pick writes the line and `$defaults` into the picked config folder's user settings only; a project settings file never receives `autoMode`; a second apply adds nothing.
  verify: `cargo test --locked --offline --test trial_picks_cli hard_deny_`
- AC-4 (REQ-5, REQ-6): a `--dry-run` session start with the pick carries the two attributes appended to a planted user value, percent-encoded, and none of the telemetry switch, exporter, endpoint, header or content variables; the next preview shows the pick unpicked, and an apply without it removes the attributes.
  verify: `cargo test --locked --offline --test trial_picks_cli otel_`
- AC-5 (REQ-7): with a stand-in `claude`, the eval pick runs `claude plugin eval` with `--no-publish` and the given `--max-cost-usd`; without a ceiling or with a stand-in version below 2.1.269 nothing runs.
  verify: `cargo test --locked --offline --test trial_picks_cli eval_`
- AC-6 (REQ-8, REQ-9): the plugin pick lists the entry and the install command with the cost lines; the server pick runs the stand-in `claude mcp add` once per picked folder with `CLAUDE_CONFIG_DIR` set, and a folder holding the server keeps it.
  verify: `cargo test --locked --offline --test trial_picks_cli plugin_`
  verify: `cargo test --locked --offline --test trial_picks_cli docs_server_`
- AC-7: neighbouring checks.
  verify: `cargo fmt --check`
  verify: `cargo clippy --locked --offline --all-targets -- -D warnings`

## Scope

- this slice, write: `.moai/specs/SPEC-20261122-trial-picks/**`
- later slices, write: a new module `src/trial_picks.rs`, the preview and apply steps in `src/project_update.rs` and `src/update.rs`, the session environment in `src/session.rs`, `tests/trial_picks_cli.rs` with a stand-in `claude`, the eval suite folder of the skill package, the documentation-line data file
- read-only: `src/claude.rs` (the settings writer), `.moai/specs/SPEC-20261122-settings-write-check/spec.md`, `.moai/specs/SPEC-20261122-project-scale-settings/spec.md`, `.moai/specs/SPEC-20261122-headless-run-contract/spec.md`
- forbidden: the real HOME and the real config folders in tests, a real `claude` in tests, the network in tests, the integration branch, `AGENTS.md` and `AGENTS.local.md` of this repository

## Plan

1. (done) SPEC (this slice), with the documentation lines above.
2. Measure the unconfirmed facts in a temporary config folder before the code slice that needs them: the project root of `plansDirectory`, whether a custom plan file is re-injected after compaction, whether a bare skill folder is an eval target.
3. Slice A: the six rows in both previews and the no-pick rule (REQ-1, REQ-2, REQ-10); the documentation-line data file carries page, line, quote and read date.
4. Slice B: plan files and the classifier line through the settings writer with its backup and checks (REQ-3, REQ-4).
5. Slice C: the OTel attributes in the session environment and the never-carry-over rule (REQ-5, REQ-6).
6. Slice D: the eval run, the security plugin row and the documentation server (REQ-7, REQ-8, REQ-9).

## Risks

- The classifier reads prose rules with a model; a hard_deny line is a second gate, not a guarantee. The row says so and names `permissions.deny` (the hooks SPEC's deny rules) as the gate that runs first.
- Writing a user settings file is a guarded write; it goes through the settings writer's backup, checks and rollback, and only on the user's pick.
- Custom resource attributes ride on every metric and event the user's exporter sends. The row says where they go: only to the endpoint the user configured.
- An eval run costs model calls (about six runs per case by default plus judge calls); the ceiling is required and the row shows the estimate formula.
- The security plugin's first run installs a Python environment from the network and runs Opus reviews; the row says so before the pick.
- The documentation server's tool names are not documented; moai-research text names the server, not its tools.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | All six start unpicked in every preview; the OTel pick is asked again every time. | The card: each stays off unless the user picks it; external transmission is the user's pick every time. | Remember picks. |
| D-2 | The classifier line goes into the user settings of picked config folders, never a project file. | The classifier ignores `autoMode` in project files (auto-mode-config.md L72). | Pass it with `--settings` per lane run only. |
| D-3 | moai adds only resource attributes; it never turns export on or picks a destination. | Export is the user's configuration; project files cannot set the exporter variables anyway (monitoring-usage.md L69). | Offer a full OTel setup. |
| D-4 | The eval run always passes `--no-publish` unless publishing is its own pick. | The default publishes a private artifact when signed in (plugin-evals.md L448). | Publish by default. |
| D-5 | moai runs the documented `claude mcp add` instead of writing `.claude.json`. | `.claude.json` is a separate global file outside the settings writer's rules. | Write the key directly. |
| D-6 | The default plan folder is `.moai/plans` inside the project. | It keeps plans beside the SPECs and inside the project as the setting requires; whether to commit them is the user's choice in the row. | Another folder. |

## Result

(Filled when the work is done. This is the SPEC-only slice.)
