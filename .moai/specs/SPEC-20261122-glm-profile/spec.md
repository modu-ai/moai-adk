---
id: SPEC-20261122-glm-profile
title: "The GLM profile: a start-time settings object for GLM lanes only, the key only through the key-helper command, the real window, the not-available list"
status: draft
revision: 4
created: 2026-11-22
parent: null
---

# GLM profile: one start-time settings object for GLM lanes, never a key value

## Intent

GLM 레인은 지금 게이트웨이 값을 실행기 환경으로만 받아요(`src/session.rs:1069` 의 `command.envs(GLM_ENV)`). 공식 문서는 셸로 내보낸 게이트웨이 값이 "doesn't reliably reach background agents hosted by the supervisor" 이라고 적고, 해법으로 "Use a settings file for any gateway that background agents must always route through" 를 줘요(`llm-gateway-connect`, 2026-11-28T02:40Z 다시 읽음). 시작 때 주는 `--settings` 값은 배경 세션으로 이어지고(`agent-view` 498·760행, 아래 근거 표) 한 세션 동안만 적용되며 어떤 파일에도 쓰이지 않아요(`settings` 269행). 이 SPEC 은 GLM 레인 시작 때 그 값 하나를 만들어 줘요.

사용자 결정 두 개가 설계를 정해요.
- 결정 6141(2026-11-22): 게이트웨이 값은 GLM 레인에만 시작 때 `--settings` 로 주고, 공유 설정 루트와 사용자 설정과 사용자 세션은 first-party Claude 로 둬요. 그래서 세션 메시지가 모든 레인에 닿아요. 키는 apiKeyHelper 자리에만 있어요.
- 결정 6310(2026-11-22, 모든 OS): 키는 새 `moai key-helper glm` 명령이 moai 금고에서 읽어 터미널이 아닌 stdout 에만 값을 줘요. 값·길이·지문은 어떤 답이나 로그에도 없어요.

개정 4(2026-11-28, tip `b662181ea9`)가 바로잡는 것: 개정 3 은 REQ-1·AC-1·D-1·D-7 만 결정에 맞췄고 제목, Intent, Scope, Plan, Risks 에 공유 루트 계획이 남아 있었어요. 또 옛 코드 가지(`spec/20261122-glm-profile-code`)의 생성기는 별칭을 `glm-4.6`, 창을 200000 으로 적었는데 tip 은 사용자 결정 `SPEC-20261123-glm-54-model` 로 `glm-5.4[1m]` 와 1000000 이에요. 문서에 따르면 모르는 id 에 `[1m]` 이 붙으면 1M 창으로 가정하고 1M 보다 작은 변수 값은 따로 적용되지 않으므로(`model-config` 957~961행) 200000 선언은 이 별칭에서 효과가 없어요. 마지막으로 인증 우선순위는 `ANTHROPIC_AUTH_TOKEN` 이 `apiKeyHelper` 보다 앞이라(`authentication` 268~276행; `settings-reference` 5275행: 헬퍼 재실행 조건은 `ANTHROPIC_AUTH_TOKEN` 이 없을 때만) 실행기가 지금처럼 금고 키를 `ANTHROPIC_AUTH_TOKEN`으로 넘기면 헬퍼는 쓰이지 않아요. 이 SPEC 은 그 충돌을 REQ-2 로 풀어요.

베타 끄기(`CLAUDE_CODE_DISABLE_EXPERIMENTAL_BETAS`)는 tip 의 `GLM_ENV` 가 기본으로 켜 둬요(2026-11-10 사용자 결정, `src/session.rs:331`). 이 SPEC 은 그 기본을 바꾸지 않고, 바꿀지는 열린 질문 Q2 예요. 안 되는 기능 목록은 프로필 글에, 상태줄 5H·7D 는 GLM 페이로드에서 absent 로 나오는 golden 으로 고정해요.

### What exists (tip `b662181ea9`)

- 게이트웨이 값의 지금 길: `src/session.rs:337` `GLM_ENV`(9쌍: 기본 URL, 타임아웃, 별칭 4개, 창 변수 2개, 베타 끄기)를 `launch` 가 `command.envs` 로 호스트 프로세스에 줘요(`:1069`). `compose` 는 같은 이름을 `env_keys` 로 알려요(`:1039`).
- 키의 지금 길: 키는 환경 `ANTHROPIC_AUTH_TOKEN` 에서, 없으면 moai 금고에서 와요(`glm_token`, `:401`). 값은 `ANTHROPIC_AUTH_TOKEN`(`:1075`)과 `Z_AI_API_KEY`(`:1076`, z.ai MCP 서버가 읽는 이름)로 호스트 프로세스에만 가요(`tests/launcher_glm_key_file.rs`). 금고 범위는 프로젝트 항목 다음 Shared 항목이에요(`SPEC-20261124-glm-key-scope`, 승인됨).
- `key-helper` 명령도 `src/glm_settings.rs` 도 tip 에 없어요. 옛 가지의 커밋 4개(`a765636b85`, `4ad95b0eea`, `9f94c4d8cf`, `3dec1444db`)는 합쳐지지 않았고, 이 SPEC 은 그 내용을 현재 tip 위에서 다시 쓰는 것을 전제로 해요(옛 `helper_line` 은 프로젝트가 풀리지 않으면 Shared 로 폴백했는데, 런처의 `read_here` 는 그 경우를 거절하므로 둘이 어긋나요: REQ-9).
- `--settings` 한 개를 만드는 선례: `src/headless_run.rs:225` 가 인라인 JSON 한 개를 만들어요.
- 상태줄 입력은 `rate_limits` 를 Option 으로 파싱하고(`src/statusline/input.rs`), 있을 때만 상태에 복사해요(`src/statusline/state.rs`). 일반 absent 시험은 있어요(`tests/statusline_counts_refresh.rs`, `tests/usage_limit_cli.rs`). GLM 모양 페이로드를 이름으로 겨냥한 golden 은 없어요.

### Evidence (official pages, read 2026-11-28)

| Page | Read (UTC) | sha256 | Read how |
| --- | --- | --- | --- |
| model-config | 02:40:58Z | `5af716d1c4a3db4f932fc9dd5400403f50f335d67a737ff1134dac8757eab0da` | whole |
| llm-gateway-connect | 02:40:58Z | `cf561ea099b6611d0590e564535133f7e16a2649ec86532fccafa2cf0958e4d3` | whole |
| llm-gateway-protocol | 02:40:58Z | `179829f8dde2f4575a7937f89f9b167176c06224aad84747ce85ae107b24752d` | whole |
| env-vars | 02:40:59Z | `30cd22fe95a40ced8d175d368d5e915441014255f40965e681d2a3dcbdff317e` | whole |
| settings | 02:40:59Z | `7587a3251756cd151948718b70fff20fd26e3a4b545858d172d0941d57ab8f10` | whole |
| authentication | 02:41:04Z | `b7f53a69dadcbea9010e802460e34782074fac94a42dbfcd5ce5a70d62be0fc0` | whole (per the research agent's report) |
| cli-reference | 02:41:05Z | `6f14dee1b0d15f83910653e82fe01c11cf688dc6d3c31790c3a4cd187bac367a` | whole (per the research agent's report) |
| settings-reference | 02:42:13Z | `f87b28f014d323f1d99655f4dd8f329693e8295eec0f1cecf8af0bcc534a5249` | partial: lines 1-60, 200-359, 2000-2001, 2440-2503, 5225-5334, 5690-5895 |
| agent-view | 02:42:14Z | `12c79547e5ac5cbd0d554ac98ddbea6e8c4bf3903bb7247f65eff1b5d14abac1` | partial: lines 1-40, 395-532, 640-783 |
| errors | 02:42:14Z | `e4d74f5e0dc5453ab6b522faabb541e3062722a6ff93912a2fb1cf3944a7a83c` | partial: lines 1-100, 1190-1409 |

Verified by the reads: all ten variable names in `GLM_ENV` and `ANTHROPIC_AUTH_TOKEN` are still named as they are (`env-vars`); `--settings` takes a path or an inline JSON string, ranks above the user, project and local files, lasts one session and writes no file (`cli-reference` 132, `settings` 160·231·269); the `env` block is read from `--settings` (`settings-reference` 2469); `apiKeyHelper` has scope "Any file" (`settings-reference` 5258) and the helper must print nothing but the key (`llm-gateway-connect` 319, `errors` 1241·1257); `--settings` is carried to a background session (`agent-view` 498·760).
Not stated in any page read: that `apiKeyHelper` may be given through `--settings` (only `env` is named there), and whether the helper's stdout is a terminal or a pipe. Both are unverified; see Risks and Q3.

## Interview

- Q1: GLM 값을 어디에 주나요? → A: GLM 레인에만, 시작 때 `--settings` 로. 공유 설정 루트와 사용자 세션은 first-party Claude 로 둬요(사용자 결정 2026-11-22, 카드 C-340 노트 6141).
- Q2: 키는 어떻게 주나요? → A: `moai key-helper glm` 이 금고에서 읽어 터미널이 아닌 stdout 에만 값을 줘요(사용자 결정 2026-11-22, 노트 6310, 추천 옵션, 모든 OS).

## Non-goals

- 키 값의 저장·출력·대신 쓰기: 어떤 조각에서도 하지 않아요(금지). `moai key-set` 이 쓰는 금고 항목은 그대로예요.
- 공유 설정 루트의 사용자 settings 에 쓰기, 프로젝트 `.claude/settings.json` 이나 `.claude/settings.local.json` 에 쓰기: 결정 6141 과 문서 경고("Don't put the credential in a project's `.claude/settings.json`")로 범위 밖.
- `Z_AI_API_KEY` 전달 방식 바꾸기: 열린 질문 Q1.
- 환경 쌍 주입(`command.envs(GLM_ENV)`) 없애기: 이 SPEC 은 그대로 둬요(D-9). 없앨지는 헬퍼 경로가 실제 `claude` 에서 검증된 뒤의 별도 개정이에요.
- 상태줄 진행 순서 재설계(다른 카드), 원장·흐름 규칙(다른 카드), 데스크톱 앱 경로(`desktop.md`: 데스크톱은 `ANTHROPIC_BASE_URL` 을 스스로 정해요).

## Requirements

- REQ-1: When moai starts a GLM lane, the launcher shall pass that lane one inline `--settings` JSON object holding an env block built from the launcher's GLM pair list and the apiKeyHelper entry naming the key-helper command by absolute path, and no credential value · failure: none — the shared config root's user settings and the user's own session stay first-party Claude (the user decisions on this card), and nothing is written to any settings file.
- REQ-2: When the moai key vault holds the glm key, the launcher shall give the GLM lane no `ANTHROPIC_AUTH_TOKEN` variable, because that variable outranks the helper.
- REQ-3: The window declaration in the settings object shall be the value of the launcher's GLM pair list and no second copy of it · failure: none — a changed pair changes the object and the lane's environment together.
- REQ-4: The profile text shall name the side effects of the beta-disable variable (context management, beta tool-schema fields, the structured-output format field, MCP tool search) next to the entry that sets it · failure: none — a missing name is a documentation defect, not a runtime failure.
- REQ-5: The profile text shall carry the not-available list of the gateway route (artifact publishing, claude.ai connectors, Code Mode, advisor, web-fetch subagents, gateway WebSearch, Remote Control, channels, Chrome, voice dictation) with the measured offered-tool versus usable-path distinction · failure: none — a missing item in the list is a documentation defect, not a runtime failure.
- REQ-6: The status line renderer shall show the five-hour and seven-day windows as absent when the payload carries no rate_limits object, pinned by a golden test on a GLM-shaped payload · failure: none — a payload with rate_limits renders exactly as before.
- REQ-7: The `moai glm --dry-run` notes shall name a credential variable in the caller's environment that outranks the helper or that bills unattended headless runs · failure: none — a note changes nothing itself.
- REQ-8: The `moai key-helper` command shall print the stored key and nothing else, and only to a stdout that is not a terminal · failure: a terminal stdout, a missing entry or a blocked entry — one plain line on stderr and a non-zero exit, with no value, length or fingerprint in any answer or log.
- REQ-9: The `moai key-helper` command shall resolve the entry through the function the launcher uses to read the key · failure: none — the launcher and the helper cannot name different entries for one folder.
- REQ-10 (REQ-2): IF the key is absent from both the vault and the caller's environment, THEN the launcher shall answer `key_absent` as it does today and pass a caller-exported `ANTHROPIC_AUTH_TOKEN` unchanged.

## Acceptance

- AC-1 (REQ-1, REQ-3): `moai glm --dry-run` under a temporary HOME answers argv with one `--settings` followed by one JSON object whose env block equals the launcher's pair list (names and values) and whose `apiKeyHelper` is the running binary's absolute path plus `key-helper glm`; the object holds no credential value; no file under the temporary HOME's config folder is created or changed · verify: `cargo test --locked --offline --test glm_profile_cli ac1_`
- AC-2 (REQ-2, REQ-7, REQ-10): with a fake key in the temporary vault, the lane environment carries no `ANTHROPIC_AUTH_TOKEN`; with the variable exported by the caller it is passed and the dry-run notes name it as outranking the helper; with no key anywhere the launcher answers `key_absent` · verify: `cargo test --locked --offline --test glm_profile_cli ac2_`
- AC-3 (REQ-4): the profile text names every listed side effect of the beta-disable variable next to its entry · verify: `cargo test --locked --offline --test glm_profile_cli ac3_`
- AC-4 (REQ-5): the profile text carries every item of the not-available list and the offered-versus-usable distinction · verify: `cargo test --locked --offline --test glm_profile_cli ac4_`
- AC-5 (REQ-6): a GLM-shaped payload without rate_limits renders both windows as absent in the golden, and a payload with them renders unchanged · verify: `cargo test --locked --offline --test usage_limit_cli glm_absent`
- AC-6 (REQ-8, REQ-9): with a fake key in a temporary vault, the binary prints exactly the key to a pipe and exits 0, refuses a terminal stdout (pseudo-terminal) with a plain line and no value, fails closed for a missing entry, and reads the same entry the launcher reads for the same folder · verify: `cargo test --locked --offline --test key_helper_cli`

## Scope

- this slice, write: `.moai/specs/SPEC-20261122-glm-profile/**`
- later slices, write: `src/glm_settings.rs` (new) and its `src/lib.rs` line, the `key-helper` arm in `src/main.rs` and its rows in `src/help.rs` (shared hubs: through the owning cards' give-back or a line lease), `src/key_vault.rs`, `src/session.rs`, `src/launcher_terminal.rs`, `src/statusline/input.rs`, `src/statusline/state.rs`, `tests/glm_profile_cli.rs`, `tests/key_helper_cli.rs`, `tests/usage_limit_cli.rs`, a GLM-shaped fixture under `tests/fixtures/statusline*/`, `docs/design-review/appendix/operations/key-helper.md`
- forbidden in every slice: writing the real HOME, the real vault or any real config folder; a real GLM request; printing or storing a credential value; a project settings file as the profile target

## Plan

1. (done) SPEC slices of revisions 1 and 3; revision 4 (this one) is the draft for approval.
2. After approval, in this order: (a) `key-helper` command with its red tests first (REQ-8, REQ-9); (b) `src/glm_settings.rs` builds the object from `GLM_ENV` (REQ-1, REQ-3) and the launcher passes it as one `--settings` argument (AC-1); (c) the launcher drops `ANTHROPIC_AUTH_TOKEN` for a vault key (REQ-2, AC-2); (d) the dry-run note (REQ-7).
3. Status line: the absent-5H/7D golden on a GLM-shaped fixture (REQ-6). It touches no key or settings path and is independent of the approval.
4. Capability list and beta side effects as profile text (REQ-4, REQ-5), shared with the doctor capability table when that card lands.
5. Before step 2(c) merges, one check in a throwaway HOME with the installed `claude`: whether an `apiKeyHelper` given through inline `--settings` is accepted and how the helper's stdout is connected (Q3). It makes no GLM request.

## Risks

- `apiKeyHelper` through `--settings` is not stated in the pages read (only `env` is). If Claude Code ignores it there, the lane has no key and fails with "Your apiKeyHelper script is failing"-style errors; the only other documented place for the helper is a user or project settings file, which decision 6141 forbids. That fallback is the user's choice (Q3), never this SPEC's.
- The helper's stdout may be a pipe or a terminal-like device; no page says. The command refuses a terminal by design (decision 6310), so a host that attaches a pseudo-terminal would fail closed. The throwaway-HOME check in Plan step 5 settles it.
- The key reaches Claude Code twice today (`ANTHROPIC_AUTH_TOKEN` and the MCP name `Z_AI_API_KEY`). REQ-2 removes the first for a vault key; the second stays (Q1), so the key is still in the lane process environment until the user picks otherwise.
- A leftover `ANTHROPIC_AUTH_TOKEN` or `ANTHROPIC_API_KEY` in the caller's environment silently outranks the helper (precedence: cloud variables, token, key, helper, OAuth); REQ-7 names it, nothing more. A key saved by `/login` is used instead of these sources unless `CI` is set (`authentication` 302).
- The helper output is cached five minutes by default (`CLAUDE_CODE_API_KEY_HELPER_TTL_MS`) and a run slower than ten seconds shows a warning; reading the vault is local and fast, but a vault in a slow HOME would surface that notice.
- The window value must be the real one per model. Per `model-config` 957~961 a `[1m]` id with a smaller declaration ignores the declaration; the object copies the pair list and refuses to guess.
- A settings-file `env` value beats a shell export (documented, "in most sessions"). While the launcher keeps injecting the pair into the environment (D-9) the two sources carry the same values from one list, so they cannot disagree; a later removal of the injection needs the Plan step 5 check first.
- Remote Control and voice are unavailable on the gateway route regardless of this profile (claude.ai identity needed); the not-available list prevents rediscovering this per lane.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | The gateway values reach a GLM lane through one start-time `--settings` object that lane alone reads, never the shared config root's user settings. | User decision 6141; the background-agent warning and the committed-file warning from `llm-gateway-connect`. | Write the shared user settings instead. |
| D-2 | The credential path is `apiKeyHelper` naming `moai key-helper glm`; moai names the slot and never writes a value. | User decision 6310; `authentication` 230-235. | Read the token from the vault into the environment as today. |
| D-3 | The beta-disable pair stays in the pair list as the tip has it (2026-11-10 user decision); the profile text names its side effects. | `src/session.rs:331`; `llm-gateway-connect` 514 names it as the repair, `llm-gateway-protocol` 285-291 its side effects; Claude Code v2.1.304 or later retries without the context-management pair on its own (`llm-gateway-protocol` 221). Whether to turn it into a repair-only entry is Q2. | Ship it as a repair only. |
| D-4 | The window is declared with the documented variables from the one pair list, not by renaming the alias. | `model-config` 957-961 and 1056-1058; the user decision of `SPEC-20261123-glm-54-model`. | Map the alias instead. |
| D-5 | The not-available list is profile text with the offered-versus-usable distinction, not a runtime check. | The lane measured tools offered while the paths are unavailable; a list is cheap to keep true. | Turn it into a doctor check later. |
| D-6 | The absent 5H/7D is pinned as a golden on a GLM-shaped fixture. | Card criterion 2; the renderer already models the absence structurally. | Leave it unpinned. |
| D-7 | Every lane still runs under the shared config root so the reachable-session listing sees them; only the routing rides the lane's own start-time object. | User decision 6141. | A shared-root user-settings env block. |
| D-8 | The object is inline JSON in the argument, the way `src/headless_run.rs:225` builds its `--settings`: no file under HOME, no "Settings file not found" case, no cleanup. | `cli-reference` 132 accepts a path or an inline string; the existing launcher shape. | A temporary settings file. |
| D-9 | The environment pair injection stays in this revision; only the token variable stops (REQ-2). | A smaller change; a settings-file value beats a shell export, so the same list in both cannot disagree. | Remove the injection after the Plan step 5 check. |
| D-10 | The helper command is named by absolute path (`current_exe()`), not a bare `moai`. | The helper runs through `/bin/sh` with the dispatching shell's `PATH` (`settings-reference` 5256, `agent-view` 697); another `moai` can come first on `PATH`. | A bare command name. |
| D-11 | The helper reads the entry through the launcher's own read function, so its scope rule is the launcher's (project entry, then Shared, a refusal when the identity cannot be resolved). | `SPEC-20261124-glm-key-scope` (approved); the old `helper_line` fell back to Shared where the launcher refuses. | The helper's own resolution. |

### Open questions (the user's pick before the code slices)

- Q1 — `Z_AI_API_KEY`: the z.ai MCP server reads the key from this environment name, which the helper cannot feed. Options: (a) keep passing it from the vault into the lane environment as today (default; the key stays in the lane process environment); (b) stop passing it and let the user wire the MCP server's own key source. Default (a); this revision changes nothing about it.
- Q2 — beta-disable default: (a) keep `CLAUDE_CODE_DISABLE_EXPERIMENTAL_BETAS=1` in the pair list as the tip has it (default, the 2026-11-10 decision); (b) move it to a repair-only entry offered after a 400 `Extra inputs` error, as the card body first asked (it also turns off MCP tool search). Default (a).
- Q3 — if the Plan step 5 check shows that an `apiKeyHelper` given through inline `--settings` is ignored or that the helper runs on a terminal-like stdout: options are (a) keep `ANTHROPIC_AUTH_TOKEN` in the lane environment as today and drop the helper (default; no key moves into any settings file), (b) place the helper in the user settings (against decision 6141). Default (a).

## Result

(Filled when the work is done. Revision 4 awaits the user's approval; no approval block is recorded. Only the status line slice, which touches no key or settings path, ran ahead of it.)

- AC-5: met — `glm_absent_windows_are_left_out_and_present_windows_render` in `tests/usage_limit_cli.rs` runs the real binary (plain output, the hidden clock and render-budget seams, temporary HOME, TMPDIR and project). A GLM-shaped payload (model `glm-5.4[1m]`, a declared 1M window, no `rate_limits`) renders `🧠 glm-5.4`, the version line and `🔋 CW █░░░░░░░░░ 12%` with no 5H and no 7D (the renderer leaves an absent window out, it does not print `n/a`); the same payload with both windows renders `🔋 5H ██░░░░░░░░ 25% (1h 0m) | 🔋 7D ████░░░░░░ 40% (3d 0h)` after the CW part, and nothing is written under the temporary HOME. Run: `cargo nextest run --locked --offline --no-fail-fast --test usage_limit_cli` 2 passed (slot `--max 2`). The version line is compared as `<binary-version>`, as the goldens of `tests/fixtures/statusline-claude` do. No product code changed. Guard mutant: none — `src/statusline/render.rs` is outside this card's write scope; both directions are pinned (absent and present), so a change that shows or drops a window fails one of the two assertions.
- AC-1 to AC-4 and AC-6: not_run (code slices wait for the approval).
