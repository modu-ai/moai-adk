---
id: SPEC-20261122-trial-picks
title: "Claude Code trial options of the whole-docs report are offered in the project and update previews, each off until the user picks it, each resting on a cited line"
status: draft
revision: 2
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

Revision 2 (2026-11-23, the leader's coverage note on this SPEC): adds rows 7 to 38, the other TRIAL items of sections 1 to 15 of the same report, as user picks that are off by default, with REQ-11 to REQ-22, AC-8 to AC-19, the unsettled facts they need, decisions D-7 to D-12 and slices E to I; TRIAL items another SPEC owns are pointer rows only.

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

Unsettled for revision 2 (none of them is settled here; each is measured or re-read before the slice that needs it, and the slice records the answer with its page, line and read time):

- U-1 (row 7): the report gives the switch as `pluginConfigs["cc-plugin-agents-md@builtin"].options.instructionFiles` and as a `/config` choice with no line number; whether a user settings file accepts that exact path and value is unsettled.
- U-2 (row 8): which variables `CLAUDE_CODE_SUBPROCESS_ENV_SCRUB` removes; the report says only "credentials" in Bash, hook and MCP environments.
- U-3 (row 9): the accepted values of `autoUpdatesChannel` besides `"stable"`, and whether `minimumVersion` refuses older sessions or only stops downgrades.
- U-4 (row 10): the scope and the key location of `bashEditDiffEnabled`; the report gives the gate name only.
- U-5 (rows 11 and 12): the value form of the one-hour TTL in `promptCacheTtl`, `subagentPromptCacheTtl` and `experimental.cacheTtl`, and the rule that ignores it while usage credits are in use.
- U-6 (row 13): the unit and the default of `CLAUDE_CODE_WARN_DISK_FREE_GIB`.
- U-7 (row 15): the StopFailure payload fields that tell a usage limit from an authentication failure; the report cites only changelog 2.1.78.
- U-8 (row 16): whether a PreModelSwitch hook can block, and where `estimated_cache_write_usd` sits in its input.
- U-9 (row 17): the matcher name `REPL` and the shape of `tool_input.code` while Code Mode is experimental.
- U-10 (row 18): the InstructionsLoaded input fields beyond `file_path`, `memory_type` and `load_reason`.
- U-11 (row 19): how a skill's `paths:` list is matched (relative to which folder, and on which tools).
- U-12 (row 20): how Claude Code decides "changes to docs or tests" before it skips the `verify` skill.
- U-13 (row 27): the name of the Remote Control "push when action is needed" setting; the report does not give it.
- U-14 (row 28): the argument form of `--channels` and whether a channel server registers under the current MCP protocol without `MCP_PROTOCOL_NEGOTIATION=legacy`.
- U-15 (row 25): the full identifier of the devcontainer feature; the report gives `claude-code:1.0` only.
- U-16 (row 31): whether `claude --init-only` makes a network or model call.
- U-17 (row 33): which file "a theme file" means for artifact design tokens.
- U-18 (row 36): the current `srt` configuration format, which the report marks as changing.
- U-19 (rows 29 and 30): the setup a managed Code Review needs on the repository side, the form of its severity line, and the current Ultrareview price (the report says about $5 to $25 after three free runs).

### Revision 2 inventory

Every row is off by default and is applied only on the user's pick after the preview. `<config>` is each config folder the user picks (`CLAUDE_CONFIG_DIR`, else the default Claude folder of the HOME). "§" is the section of the whole-docs report of 2026-11-22; the page is under `https://code.claude.com/docs/en/`; the versions are the changelog versions the report cites.

| Row | Item | What moai writes, where, in which scope | Source (as the report gives it) | Default | Cost or transmission |
| --- | --- | --- | --- | --- | --- |
| 7 | AGENTS.md switch | `pluginConfigs["cc-plugin-agents-md@builtin"].options.instructionFiles` = `"claude-md"` in `<config>/settings.json`; user scope (project settings cannot set it) | §2-5; memory.md, settings-reference.md; 2.1.277 | off | No cost. No `AGENTS.md` is read in any project of that config folder, also where no `CLAUDE.md` exists |
| 8 | Subprocess environment scrub | `env.CLAUDE_CODE_SUBPROCESS_ENV_SCRUB` = `"1"` in `<config>/settings.json`; user scope | §3-2; env-vars.md | off | No cost. Credentials leave the Bash, hook and MCP environments, so a tool that reads a token from its environment stops working there |
| 9 | Update channel and floor | `autoUpdatesChannel` = `"stable"` and `minimumVersion` = the lowest Claude Code version moai's own checks ran on, in `<config>/settings.json`; user scope | §3-2; setup.md | off | No cost. Pins the updates of every session of that config folder |
| 10 | Bash edit diff | `bashEditDiffEnabled` = `true`; scope as the page gives it (U-4) | §13 #28; hooks.md (public beta) | off | No cost. "best effort ... not to enforce a policy": moai reads the changed-file list as an out-of-scope hint only |
| 11 | Prompt cache one hour | `promptCacheTtl` and `subagentPromptCacheTtl` = the one-hour value in `<config>/settings.json`; user scope | §11; prompt-caching.md, settings-reference.md | off | Cache writes cost more than the default |
| 12 | Audit agent cache one hour | `experimental.cacheTtl` = the one-hour value in the front matter of `<config>/agents/moai/moai-audit.md`; user scope | §13 #30; sub-agents.md, prompt-caching.md; 2.1.248 | off | Cache writes cost more; ignored while usage credits are in use |
| 13 | Disk warning | `env.CLAUDE_CODE_WARN_DISK_FREE_GIB` = the GiB value the user gives, in `<config>/settings.json`; user scope | §11; env-vars.md | off | None |
| 14 | moai-research memory | `memory: local` in the front matter of `<config>/agents/moai/moai-research.md`; Claude Code then keeps `.claude/agent-memory-local/moai-research/` in each project | §1-2, §5-1; sub-agents.md, claude-directory.md; 2.1.33 | off (moai-audit never gets it: §1-2 NO) | The first 200 lines or 25 KB of that memory join every moai-research prompt |
| 15 | StopFailure hook | `hooks.StopFailure` entry running `moai`, which appends one ledger note for a lane stopped by a usage limit or an authentication failure, in `<config>/settings.json` | §4-1; 2.1.78 | off | None; exits 0 outside a moai project |
| 16 | PreModelSwitch hook | `hooks.PreModelSwitch` entry running `moai`, which prints one warning with `estimated_cache_write_usd` before a model switch and never blocks, in `<config>/settings.json` | §4-1; hooks.md; 2.1.251 | off | None; exits 0 outside a moai project |
| 17 | REPL matcher | `hooks.PreToolUse` entry with matcher `REPL` that runs the existing pre-tool check over `tool_input.code`, in `<config>/settings.json`; offered only when the user's own settings turn Code Mode on (moai never does) | §4-5, §13 #1; code-mode.md; 2.1.323 (experimental) | off | None. The check sees the code text only, as the classifier "doesn't see values computed at run time" |
| 18 | InstructionsLoaded hook | `hooks.InstructionsLoaded` entry running `moai`, which appends `file_path`, `memory_type` and `load_reason` to a log in moai's state folder, in `<config>/settings.json` | §2-6; memory.md, hooks.md | off | None. Not called for an `AGENTS.md` read through the setting |
| 19 | SPEC helper skill with `paths:` | project `.claude/skills/spec-helper/SKILL.md` whose `paths:` names the project's SPEC folder; project scope, from `/moai project` only | §5-3; skills.md | off | Its description joins the skill listing budget (1% of the context) |
| 20 | Project `verify` skill | project `.claude/skills/verify/SKILL.md` that runs the light steps of `moai check`; project scope only, never a config folder | §5-4; skills.md; 2.1.286 | off | Runs "right before each commit, except for changes to docs or tests"; does not run when the built-in git instructions are off (report §17 ⑧) |
| 21 | Area guidance hook | project `.claude/settings.json` `hooks.SessionStart` entry that names the area's plugin from the start folder; monorepo projects only | §12-2; large-codebases.md | off | None; needs workspace trust like every project hook |
| 22 | CLAUDE.md update Stop hook | project `.claude/settings.json` `hooks.Stop` entry that suggests a `CLAUDE.md` update and reads `stop_hook_active` before it blocks | §12-3; large-codebases.md, hooks.md; 2.1.163 | off | One extra turn each time it blocks; overlaps auto memory |
| 23 | JetBrains edit deny | project `.claude/settings.json` `permissions.deny` entry `Edit(./.idea/**)`; offered only when `.idea/` exists | §13 #14; jetbrains.md (beta) | off | None |
| 24 | iOS simulator rule | project `.claude/rules/ios-simulator.md` with one line saying that lanes share one simulator; offered only when an Xcode project exists | §13 #23; desktop-ios-simulator.md (public beta) | off | None |
| 25 | Devcontainer | project `.devcontainer/devcontainer.json` with the Claude Code feature `claude-code:1.0`; offered only when no devcontainer file exists | §8-4; devcontainer.md | off | The container downloads Claude Code when it is built |
| 26 | Dynamic workflow | project `.claude/workflows/cross-audit.js` for a large cross audit; project scope | §8-4, §11; workflows.md | off (also off by default on Pro) | Model calls; "Large workflow" warning above 25 agents or 1.5M tokens |
| 27 | Remote Control notices | prints how to turn on "push when action is needed" in the user's own Remote Control setup; writes nothing | §7-3; remote-control.md; 2.1.196, 2.1.222 | off | External: while connected "the session transcript ... is stored on Anthropic servers"; not available behind a gateway |
| 28 | Channels | one profile key naming a channel server the user installed; interactive lanes moai starts then carry `--channels` with that server; only for remote-check failure notices | §7-4, §13 #5; channels.md, channels-reference.md (research preview) | off | External as the server sends it. "An ungated channel is a prompt injection vector"; permission relay is never offered; ignored with `-p`; needs claude.ai or Console sign-in and Bun |
| 29 | Managed Code Review | prints the setup steps for release pull requests only; moai's receipt reads a missing severity line as no verdict | §8-2; code-review-pull-requests.md | off | Paid per push; the code goes to the review service |
| 30 | Ultrareview | for an audit tier 2 area prints `/ultrareview` for the user to run; moai never runs it | §8-2, §13 #10; ultrareview.md (research preview) | off | Three free runs, then about $5 to $25 each; uploads a repository bundle |
| 31 | Installed-hook smoke run | `moai doctor` runs `claude --init-only` once in a temporary project with `CLAUDE_CONFIG_DIR` set to a picked config folder; writes nothing | §8-4; cli-reference.md | off | Runs only the Setup and SessionStart hooks; network use unsettled (U-16) |
| 32 | Screen reader text | one profile key; the card board and the progress board print `Header: value` lines instead of tables | §9-2; accessibility.md | off | None |
| 33 | Report colours from design tokens | write target unsettled (U-17); listed as not yet available | §9-2; artifacts.md | off | None |
| 34 | Claude in Chrome | prints `/chrome` as the web-check option of a lane; writes nothing (project settings cannot turn it on) | §13 #7; chrome.md; 2.1.198 | off | Needs claude.ai sign-in; its context is always loaded once it is on |
| 35 | Fullscreen renderer | prints `/tui fullscreen` as the user's terminal choice; writes nothing | §13 #12; fullscreen.md (research preview) | off | None; conflicts with `tmux -CC` |
| 36 | Sandbox runtime for audit reproductions | one profile key; an empty-tool-home audit reproduction runs under `srt` when it is on PATH; moai installs nothing | §13 #25; sandbox-environments.md (beta research preview) | off | None |
| 37 | OTel events and tracing beta | prints the event names `skill_activated`, `tool_decision`, `hook_execution_complete`, `subagent_completed` and `compaction` and the tracing beta for debugging; sets no variable | §11, §12-3, §13 #29; monitoring-usage.md | off | External only through the user's own collector; the tracing beta carries content attributes |
| 38 | Usage insights | prints one line naming `/insights`; writes nothing | §11; costs.md (the report marks it guidance; the coverage note lists it) | off | None |

### Rows other SPECs own (pointers only, not specified here)

| Row | Item | Source (report) | Owner |
| --- | --- | --- | --- |
| P-1 | SubagentStart and SubagentStop hooks | §4-1 | report §17 ④ (recommended: not used) |
| P-2 | Lane-closing Stop hook | §4-1 | `SPEC-20261122-nonstop-lane` (`/goal` per card on the user's pick); `SPEC-20261122-next-flow-hooks` keeps Stop hooks out of the moai hook set |
| P-3 | `claude --bg` and the agent view as lane host | §6-2, §13 #2 | `SPEC-20261122-bg-session-research`, `SPEC-20261122-autonomous-orchestration` |
| P-4 | Remote Control server mode `--spawn worktree` | §7-3, §13 #40 | `SPEC-20261122-bg-session-research` (user pick: not used); report §17 ⑥ (§7-3 calls it ⑤) |
| P-5 | `CLAUDE_CODE_DISABLE_EXPERIMENTAL_BETAS` | §7-5, §13 #34 | `SPEC-20261122-glm-profile` (REQ-4) |
| P-6 | GitLab CI/CD | §8-4, §13 #17 | `SPEC-20261122-project-ci-design` |
| P-7 | Status line `rate_limits.spend_limit` and `prompt_cache` fields | §9-1 | `SPEC-20261122-statusline-limits` (it names `spend_limit`; the `prompt_cache` fields are not named there yet) |
| P-8 | Themes and keybindings | §10-4 | the mods work of report §10 |
| P-9 | Plugin `experimental.themes`, `monitors` and `evals` | §13 #31 | report §10 and §17 ① (only when moai is bundled as a plugin) |
| P-10 | `claudeMdExcludes`, language server plugins, `sparsePaths` and `symlinkDirectories`, package skills, `additionalDirectories` | §12-2 | `SPEC-20261122-project-scale-settings` |
| P-11 | Advisor as a lane cost option | §13 #8 | the launcher's `--advisor` and `--no-advisor` options (`SPEC-20261122-next-flow-update-and-launcher`) |

## Non-goals

- 기본으로 켜는 것: 여섯 모두 꺼진 채로 나가요.
- OTel 내보내기 자체(수집기, 주소, 헤더, 내용 기록): 사용자의 몫이고 moai 는 안내 글만 보여요.
- 플러그인과 언어 서버의 설치 엔진: 설치 명령을 돌리는 것까지예요. `enabledPlugins` 값 검사는 규모 설정 SPEC 의 쓰기 검사를 써요.
- `.claude.json` 쓰기: 문서의 명령이 써요.
- The pointer rows P-1 to P-11: their owners specify them.
- Turning Code Mode, channels, Remote Control, Chrome or a collector on: rows 17, 27, 28, 34 and 37 only follow what the user turned on.
- Installing `srt`, a channel server, the Code Review service or a devcontainer runtime.

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
- REQ-11: The project and update previews shall list each of rows 7 to 38 unpicked with what it writes and where and in which scope and with its cost or transmission note and the report section and page or changelog version it rests on.
- REQ-12: The trial apply shall write exactly the listed key and value of each user-settings row the user picked after the preview (rows 7 to 11 and 13) into the user settings of each config folder the user picked and nothing for an unpicked row · failure: none — a key that already holds another value is shown as a conflict and kept.
- REQ-13: The trial apply shall add the listed front-matter field of a picked agent row only to the agent file that row names (moai-audit for row 12 and moai-research for row 14) in each picked config folder · failure: none — moai-audit never receives a memory field.
- REQ-14: The trial apply shall add one hook entry for each picked observing hook row (rows 15 and 16 and 18) to the user settings of each picked config folder whose command exits 0 outside a moai project and never blocks a call.
- REQ-15: The picked REPL entry of row 17 shall run the refusals of the existing pre-tool check over the code text it receives and be offered only while the user's own settings turn Code Mode on · failure: none — moai never sets the Code Mode switch.
- REQ-16: The project setup shall write each picked project row (rows 19 to 26) only at the project paths that row lists and only from `/moai project` · failure: none — `moai update` never offers these rows, a config folder never receives them, an existing file at a listed path is shown as a conflict and kept, and rows 21 and 23 to 25 are offered only when their detection holds.
- REQ-17: The trial apply shall print the listed notice of each picked notice row (rows 27 and 29 and 30 and 34 and 35 and 37 and 38) with its source and start no process and change no file.
- REQ-18: The trial apply shall store each picked lane option (rows 28 and 32 and 36) as one profile key that the commands moai runs later read · failure: none — row 28 adds only `--channels` with the server the user named and never the development-channels flag or a permission relay, and row 36 runs nothing when `srt` is not on PATH.
- REQ-19: The installed-hook smoke run of row 31 shall run `claude --init-only` once in a temporary project with CLAUDE_CONFIG_DIR set to a picked config folder and only on the user's pick · failure: none — a non-zero exit is reported with its output and changes no file.
- REQ-20: The trial apply shall remove only what its own receipt records for a row the user un-picks and add nothing when the same picks are applied again.
- REQ-21: Every row that sends data off the machine (rows 27 to 30 and 37) shall be offered unpicked in every preview and leave no moai key in place after an apply without its pick.
- REQ-22: The report-colour row 33 shall stay listed as not yet available and refuse a pick until a measured write target is recorded in this SPEC.

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
- AC-8 (REQ-11, REQ-2): in a temporary project and an empty temporary HOME, both previews list rows 7 to 38 unpicked, each with its scope, its default off, its cost or transmission note and its report section with page or version; no pointer row P-1 to P-11 is offered as a pick; an apply with no pick leaves every file byte-identical.
  verify: `cargo test --locked --offline --test trial_picks_cli r2_list_`
- AC-9 (REQ-12): picking row 8 for one of two planted config folders adds exactly `env.CLAUDE_CODE_SUBPROCESS_ENV_SCRUB` = `"1"` to that folder's user settings and leaves every other byte of both folders unchanged; a planted different value is reported as a conflict and kept; rows 7, 9, 10, 11 and 13 each write only their listed key and value.
  verify: `cargo test --locked --offline --test trial_picks_cli r2_settings_`
- AC-10 (REQ-13): picking row 14 adds `memory: local` to the moai-research agent file only and the moai-audit file gains no `memory` field; picking row 12 adds `experimental.cacheTtl` to the moai-audit file only.
  verify: `cargo test --locked --offline --test trial_picks_cli r2_agent_`
- AC-11 (REQ-14): each picked row of 15, 16 and 18 adds one entry for its event; run with a stand-in payload whose `cwd` is not a moai project, each command exits 0 with no output; with a moai project it never exits 2.
  verify: `cargo test --locked --offline --test trial_picks_cli r2_hooks_`
- AC-12 (REQ-15): row 17 is not offered when the planted user settings do not turn Code Mode on; with Code Mode on, the entry exits 2 for a stand-in `tool_input.code` that moves the integration branch and exits 0 for a harmless one; no apply writes the Code Mode switch.
  verify: `cargo test --locked --offline --test trial_picks_cli r2_repl_`
- AC-13 (REQ-16): `moai update` lists none of rows 19 to 26; `/moai project` with every row picked writes only the listed project paths and nothing under the config folder; a planted file at a listed path is kept and reported; row 23 is absent without `.idea/`, row 24 without an Xcode project, row 21 outside a monorepo fixture and row 25 when a devcontainer file exists; the row 22 entry reads `stop_hook_active`.
  verify: `cargo test --locked --offline --test trial_picks_cli r2_project_`
- AC-14 (REQ-17): picking every notice row prints each notice with its report section and page, the stand-in `claude` records no call, and every file stays byte-identical.
  verify: `cargo test --locked --offline --test trial_picks_cli r2_notice_`
- AC-15 (REQ-18): with row 28 picked, a `--dry-run` interactive lane start carries `--channels` with the named server and never `--dangerously-load-development-channels`, and a `-p` run carries neither; with row 36 picked and no `srt` on PATH, the audit reproduction runs as before; with row 32 picked, the card board prints `Header: value` lines and no table.
  verify: `cargo test --locked --offline --test trial_picks_cli r2_lane_`
- AC-16 (REQ-19): with row 31 picked, the stand-in `claude` receives `--init-only` exactly once with `CLAUDE_CONFIG_DIR` set to the picked folder and a temporary project as its working folder; a stand-in exit of 1 is reported with its output and no file changes.
  verify: `cargo test --locked --offline --test trial_picks_cli r2_init_only_`
- AC-17 (REQ-20): a second apply of the same picks leaves every file byte-identical; un-picking rows 8, 15 and 23 removes only the entries the receipt lists and keeps a key the user added beside them.
  verify: `cargo test --locked --offline --test trial_picks_cli r2_unpick_`
- AC-18 (REQ-21): after an apply that picked row 28, the next preview shows rows 27 to 30 and 37 unpicked, and an apply without row 28 leaves no channel key in the profile.
  verify: `cargo test --locked --offline --test trial_picks_cli r2_external_`
- AC-19 (REQ-22): row 33 is listed as not yet available with U-17 as its reason, and a pick of it is refused before any write.
  verify: `cargo test --locked --offline --test trial_picks_cli r2_tokens_`

## Scope

- this slice, write: `.moai/specs/SPEC-20261122-trial-picks/**`
- later slices, write: a new module `src/trial_picks.rs`, the preview and apply steps in `src/project_update.rs` and `src/update.rs`, the session environment in `src/session.rs`, `tests/trial_picks_cli.rs` with a stand-in `claude`, the eval suite folder of the skill package, the documentation-line data file
- later slices of revision 2, write: the base-agent renderer for rows 12 and 14, the hook entry renderer for rows 15 to 18 and 21 and 22, the project harness templates for rows 19 to 26, the card board and progress board output for row 32, the doctor command for row 31
- read-only: `src/claude.rs` (the settings writer), `.moai/specs/SPEC-20261122-settings-write-check/spec.md`, `.moai/specs/SPEC-20261122-project-scale-settings/spec.md`, `.moai/specs/SPEC-20261122-headless-run-contract/spec.md`
- read-only for revision 2: `.moai/specs/SPEC-20261122-next-flow-base-agents/spec.md`, `.moai/specs/SPEC-20261122-next-flow-hooks/spec.md`, `.moai/specs/SPEC-20261122-nonstop-lane/spec.md`, `.moai/specs/SPEC-20261122-glm-profile/spec.md`, `.moai/specs/SPEC-20261122-bg-session-research/spec.md`
- forbidden: the real HOME and the real config folders in tests, a real `claude` in tests, the network in tests, the integration branch, `AGENTS.md` and `AGENTS.local.md` of this repository

## Plan

1. (done) SPEC revision 1, with the documentation lines above.
2. Measure the unconfirmed facts in a temporary config folder before the code slice that needs them: the project root of `plansDirectory`, whether a custom plan file is re-injected after compaction, whether a bare skill folder is an eval target.
3. Slice A: the six rows in both previews and the no-pick rule (REQ-1, REQ-2, REQ-10); the documentation-line data file carries page, line, quote and read date.
4. Slice B: plan files and the classifier line through the settings writer with its backup and checks (REQ-3, REQ-4).
5. Slice C: the OTel attributes in the session environment and the never-carry-over rule (REQ-5, REQ-6).
6. Slice D: the eval run, the security plugin row and the documentation server (REQ-7, REQ-8, REQ-9).
7. (done) SPEC revision 2 (this slice): rows 7 to 38, the pointer rows and REQ-11 to REQ-22, with the report's citations.
8. Before each slice below, re-read the cited pages of its rows and measure the unsettled facts U-1 to U-19 it needs in a temporary config folder; record page, line and read time in the documentation-line data file (D-8).
9. Slice E: rows 7 to 11 and 13 in both previews, the user-settings writes, the receipt and the un-pick (REQ-11, REQ-12, REQ-20). First, because it reuses slice B's settings writer and the receipt every later slice removes by.
10. Slice F: the observing hooks of rows 15, 16 and 18 and the REPL entry of row 17 (REQ-14, REQ-15). After E, because hook entries go through the same user-settings writer and receipt.
11. Slice G: the agent rows 12 and 14 (REQ-13). After E, because it needs the receipt but touches the agent renderer, not the settings writer, so it stays a separate small slice.
12. Slice H: the project rows 19 to 26 in `/moai project` (REQ-16). After F, because rows 21 and 22 reuse slice F's hook entry renderer in the project file.
13. Slice I: the notice rows, the lane options, the smoke run, the external rule and the not-available row (REQ-17, REQ-18, REQ-19, REQ-21, REQ-22). Last, because these rows write least and the lane options need the lane start code that the other slices leave unchanged.

## Risks

- The classifier reads prose rules with a model; a hard_deny line is a second gate, not a guarantee. The row says so and names `permissions.deny` (the hooks SPEC's deny rules) as the gate that runs first.
- Writing a user settings file is a guarded write; it goes through the settings writer's backup, checks and rollback, and only on the user's pick.
- Custom resource attributes ride on every metric and event the user's exporter sends. The row says where they go: only to the endpoint the user configured.
- An eval run costs model calls (about six runs per case by default plus judge calls); the ceiling is required and the row shows the estimate formula.
- The security plugin's first run installs a Python environment from the network and runs Opus reviews; the row says so before the pick.
- The documentation server's tool names are not documented; moai-research text names the server, not its tools.
- Revision 2 rests on the report's citations, not on a fresh read; a page may have changed since 2026-11-22. Each slice re-reads its pages first (D-8), and a row whose cited line is gone is dropped from the preview, not guessed.
- Hook rows 15 to 18 live in the user settings and so run in every project of that config folder; each command exits 0 outside a moai project, and the preview says so.
- Row 8 can break a tool that reads a token from its environment inside Bash, hooks or MCP servers; the preview names that effect.
- Row 20 runs before every commit and costs a model turn each time; it also disappears when the built-in git instructions are off (report §17 ⑧).
- Rows 29 and 30 cost money and send the code off the machine; moai only prints the steps and the command, and the user runs them.
- Row 28 opens a path for outside events into a session; moai never offers the development-channels flag or a permission relay.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | All six start unpicked in every preview; the OTel pick is asked again every time. | The card: each stays off unless the user picks it; external transmission is the user's pick every time. | Remember picks. |
| D-2 | The classifier line goes into the user settings of picked config folders, never a project file. | The classifier ignores `autoMode` in project files (auto-mode-config.md L72). | Pass it with `--settings` per lane run only. |
| D-3 | moai adds only resource attributes; it never turns export on or picks a destination. | Export is the user's configuration; project files cannot set the exporter variables anyway (monitoring-usage.md L69). | Offer a full OTel setup. |
| D-4 | The eval run always passes `--no-publish` unless publishing is its own pick. | The default publishes a private artifact when signed in (plugin-evals.md L448). | Publish by default. |
| D-5 | moai runs the documented `claude mcp add` instead of writing `.claude.json`. | `.claude.json` is a separate global file outside the settings writer's rules. | Write the key directly. |
| D-6 | The default plan folder is `.moai/plans` inside the project. | It keeps plans beside the SPECs and inside the project as the setting requires; whether to commit them is the user's choice in the row. | Another folder. |
| D-7 | The source of each row 7 to 38 is the report's own citation (section, page and line or changelog version), read on 2026-11-22; this revision reads no official page again. | The lane that wrote revision 2 has no network. | Re-read every page now. |
| D-8 | Each slice that implements a row re-reads the cited page first and records page, line, quote and read time before it writes code. | A citation one day old is not a read of the current page (the read-first rule). | Trust the report's citation. |
| D-9 | TRIAL items another SPEC or a report §17 question owns are pointer rows P-1 to P-11 only. | One owner per change; the brief and the report's card column. | Specify them here too. |
| D-10 | Project rows 19 to 26 are offered by `/moai project` only and never written to a config folder. | The install-place decision of the report's preamble: a project harness is made only on the user's pick for that project. | Offer them in `moai update` too. |
| D-11 | A row moai wrote earlier shows as present with an un-pick choice; every other row starts unpicked, and the rows that send data off the machine start unpicked every time. | D-1 for unwritten rows; an un-pick needs to see what is present; external transmission is the user's pick every time. | Remember every pick. |
| D-12 | Row 38 (`/insights`) is included although the report marks it guidance rather than TRIAL. | The coverage note lists it among the cost items. | Drop the row. |

## Result

(Filled when the work is done. This is the SPEC-only slice.)
