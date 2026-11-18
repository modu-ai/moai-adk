# Research: the question tools of Codex CLI and OMP

SPEC-20261118-codex-question-tool. Facts behind the Codex CLI and OMP rows of `skills/moai/references/internal/questions.md`: each fact has its quote, source and the UTC time it was read; `kind` says whether a fact comes from documentation or from source code. A fact not found is marked `unverified`.

## Codex CLI

Source kind: the official Codex documentation at `learn.chatgpt.com/docs` (the old `developers.openai.com/codex/...` addresses redirect there) does not mention the tool at all: a search of its config reference, developer-commands and build-skills pages for `user_input` and `default_mode` found nothing (read 2026-11-18T06:15:45Z). Every fact below therefore rests on the open-source repository at the tag of the installed version, `rust-v0.170.0` (latest stable release, published 2026-11-18T03:40:48Z; read 2026-11-18T06:14:38Z from `https://api.github.com/repos/openai/codex/releases/latest`). The base URL `…` is `https://raw.githubusercontent.com/openai/codex/rust-v0.170.0/codex-rs`. Each fact says `source`. Each source entry ends with the `sha256` of the bytes the author fetched again for this note (`fetched` is that UTC time; GitHub `main` URLs and the docs page can change later, the tag URLs cannot), as the repository rule on official sources asks (`AGENTS.md`, Authority 4). The facts were read by a research agent (not this session) and the read times are the ones it measured with `date -u`.

### Tool name and the async sibling

- quote: `pub const REQUEST_USER_INPUT_TOOL_NAME: &str = "request_user_input";`
- url: https://raw.githubusercontent.com/openai/codex/rust-v0.170.0/codex-rs/core/src/tools/handlers/request_user_input_spec.rs
- read: 2026-11-18T06:14:42Z
- sha256: 86ef5b005211ea6cc14c5a99821357a4484cccbb4756c911bbeba3913315d16f (fetched 2026-11-18T06:24:38Z, 4661 bytes)
- kind: source
- quote: `const TOOL_NAME: &str = "request_user_input_async";` and the result `{"accepted":true}` — it returns at once and the reply arrives later as a new user message; it is registered only when the model catalog advertises it and the thread is the root thread.
- url: https://raw.githubusercontent.com/openai/codex/rust-v0.170.0/codex-rs/core/src/tools/handlers/request_user_input_async.rs
- read: 2026-11-18T06:14:46Z
- sha256: f89c9aa13b95456301e28a5c6e04d78dd368f2453eab7a47b82d47429e1a738d (fetched 2026-11-18T06:24:38Z, 6238 bytes)
- kind: source

`request_user_input` is the tool a step can wait on. `request_user_input_async` is a different tool that waits for nobody, which is why `questions.md` forbids it for a question a step waits on.

### Questions, options, "Other"

- quote: `"Request user input for one to three short questions and wait for the response. This tool is only available in {allowed_modes}."` and `"Provide 2-3 mutually exclusive choices. Put the recommended option first and suffix its label with \"(Recommended)\". Do not include an \"Other\" option in this list; the client will add a free-form \"Other\" option automatically."`
- url: https://raw.githubusercontent.com/openai/codex/rust-v0.170.0/codex-rs/core/src/tools/handlers/request_user_input_spec.rs
- read: 2026-11-18T06:14:42Z
- sha256: 86ef5b005211ea6cc14c5a99821357a4484cccbb4756c911bbeba3913315d16f (fetched 2026-11-18T06:24:38Z, 4661 bytes)
- kind: source
- note: the counts (1 to 3 questions, 2 to 3 options, `header` of 12 characters or fewer) are guidance in the schema text; the code enforces only that every question has a non-empty option list. The tool marks each question `is_other = true`, so the client adds the free-text choice itself.

### Modes and the feature flag

- quote: `pub const fn allows_request_user_input(self) -> bool { matches!(self, Self::Plan) }` (the source breaks this over several lines; the line breaks are joined here)
- url: https://raw.githubusercontent.com/openai/codex/rust-v0.170.0/codex-rs/protocol/src/config_types.rs
- read: 2026-11-18T06:15:06Z
- sha256: ce9b26223db7cc4b5cf474e571e4509e8cc52536e159510f153f94de0fcd37e8 (fetched 2026-11-18T06:24:38Z, 31651 bytes)
- kind: source
- quote: `mode.allows_request_user_input() || (features.enabled(Feature::DefaultModeRequestUserInput) && *mode == ModeKind::Default)` (the source breaks this over several lines; the line breaks are joined here)
- url: https://raw.githubusercontent.com/openai/codex/rust-v0.170.0/codex-rs/tools/src/tool_config.rs
- read: 2026-11-18T06:15:03Z
- sha256: 1925968949c32190288dbc3a1473520ced9c66da425996a0652dd572b1341aae (fetched 2026-11-18T06:24:38Z, 1004 bytes)
- kind: source
- quote: `key: "default_mode_request_user_input", stage: Stage::UnderDevelopment, default_enabled: false` — the same values on the `main` branch (read 2026-11-18T06:15:52Z).
- url: https://raw.githubusercontent.com/openai/codex/rust-v0.170.0/codex-rs/features/src/lib.rs
- read: 2026-11-18T06:14:49Z
- sha256: fae6cbf9cc83ae604d329deaaf5b79e806afebe8390d21e91bfee76322014a20 (fetched 2026-11-18T06:24:38Z, 86879 bytes)
- kind: source

So the key is `features.default_mode_request_user_input`, default: false, stage: under development. The tool is registered in every mode; a call from a mode that is not allowed gets an error such as `request_user_input is unavailable in Default mode`.

### No answer, and what the Default-mode text tells the model

- quote: `"If `request_user_input` returns no answers, continue with best judgment instead of asking again or treating the turn as blocked."` and `"In Default mode, strongly prefer making reasonable assumptions and executing the user's request rather than stopping to ask questions."` and `"Use the `request_user_input` tool only when it is listed in the available tools for this turn."`
- url: https://raw.githubusercontent.com/openai/codex/rust-v0.170.0/codex-rs/collaboration-mode-templates/templates/default.md
- read: 2026-11-18T06:14:59Z
- sha256: 1042cc643eb0147ca1039b19287c7462ceb297502f7f310d9664ac323a12feca (fetched 2026-11-18T06:24:38Z, 1288 bytes)
- kind: source
- quote: `"{REQUEST_USER_INPUT_TOOL_NAME} was cancelled before receiving a response"` (a format string over the tool-name constant, so the model reads `request_user_input was cancelled before receiving a response`)
- url: https://raw.githubusercontent.com/openai/codex/rust-v0.170.0/codex-rs/core/src/tools/handlers/request_user_input.rs
- read: 2026-11-18T06:22:50Z
- sha256: 8276c328843a34b08540211f79fecc3da35217a297828b459f165546071d0751 (fetched 2026-11-18T06:24:38Z, 6154 bytes)
- kind: source
- quote: `rx_response.await.ok()` (line 3720) — the core waits for the answer with no timeout of its own.
- url: https://raw.githubusercontent.com/openai/codex/rust-v0.170.0/codex-rs/core/src/session/mod.rs
- read: 2026-11-18T06:22:50Z
- sha256: 0777e54dcd4d2db1708dfd0836911f2a67af4b2853311dd2b3a934866cea32a9 (fetched 2026-11-18T06:24:38Z, 253966 bytes)
- kind: source
- quote: `error!("request failed with client error: {err:?}"); let empty = CoreRequestUserInputResponse { answers: HashMap::new(), };` (lines 1722-1727; the `error!` call and the struct literal are joined here from separate lines) — when the client request fails, the app server submits an empty answer map.
- url: https://raw.githubusercontent.com/openai/codex/rust-v0.170.0/codex-rs/app-server/src/bespoke_event_handling.rs
- read: 2026-11-18T06:22:51Z
- sha256: 74a17ccb69fd166069730ab89bc2f8ded7000cbb433e2bdf5c5edd5fa0d1c5f0 (fetched 2026-11-18T06:24:38Z, 161908 bytes)
- kind: source

This is why a case-(2) question must not go on by best judgment when nothing comes back (`questions.md` rule 3): Codex's own text asks for the opposite.

### Skill instructions, `codex exec`, subagents

- quote: the skills documentation never names `request_user_input`; its only tool hook is `"to declare tool dependencies for a more seamless experience with using the skill."` (`dependencies: tools:` in `agents/openai.yaml`, MCP type in the example). A skill therefore cannot make the call; it works only when the tool is in the turn's tool list.
- url: https://learn.chatgpt.com/docs/build-skills
- read: 2026-11-18T06:15:45Z
- sha256: b238cf9d75afa8818aa104b88f6ef15891f6e7a5a7183c2b36d2920b23f751ca (fetched 2026-11-18T06:24:38Z, 443673 bytes)
- kind: docs
- quote: `"request_user_input is not supported in exec mode for thread `{}`"` (error code -32000)
- url: https://raw.githubusercontent.com/openai/codex/rust-v0.170.0/codex-rs/exec/src/lib.rs
- read: 2026-11-18T06:15:15Z
- sha256: 05490e995c4672cbe585cffb8514cb1df0fe7b5be452b947782c86efbb10cd72 (fetched 2026-11-18T06:24:38Z, 96042 bytes)
- kind: source
- quote: `"request_user_input can only be used by the root thread"` (line 69; the restriction flag `root_capabilities_restricted` is set for non-root agents in `core/src/session/mod.rs:1010`)
- url: https://raw.githubusercontent.com/openai/codex/rust-v0.170.0/codex-rs/core/src/tools/handlers/request_user_input.rs
- read: 2026-11-18T06:22:50Z
- sha256: 8276c328843a34b08540211f79fecc3da35217a297828b459f165546071d0751 (fetched 2026-11-18T06:24:38Z, 6154 bytes)
- kind: source

`codex exec` and a subagent cannot use the dialog; `questions.md` sends those to the plain-text fallback.

### Not confirmed

- unverified: what a refused `codex exec` call looks like to the model (read from the code, not run); what Esc in the TUI returns to the model (the interrupt path was not followed); whether the model catalog of a given model advertises `request_user_input_async`.

## OMP

Source kind: the installed package `@oh-my-pi/pi-coding-agent` 18.18.7 (`omp --version` printed `omp/18.18.7`, run 2026-11-18T06:14:21Z; the package ships its source but no markdown docs) and the upstream repository `https://github.com/can1357/oh-my-pi` (named by the installed `package.json`), whose `docs/` were read from the `main` branch, which may be newer than 18.18.7. A `url:` that starts with `/` is a file of the installed package, relative to `/Users/goos/.bun/install/global/node_modules/@oh-my-pi/pi-coding-agent`. No live `omp` session was run. The facts were read by a research agent (not this session) and the read times are the ones it measured with `date -u`.

### Tool name and parameters

- quote: `readonly name = "ask";` and `const askSchema = arkType({ html: arkType("string").describe("one HTML <form>") });` with `readonly strict = true;` — one parameter, `html`, not a list of options.
- url: /src/tools/ask.ts
- read: 2026-11-18T06:14:28Z
- sha256: 890c65a5a0eee2f79a14b7f3eacf2620fb66a9a68ce96ba2964ba7b6e07bc9f4 (fetched 2026-11-18T06:24:38Z, 19004 bytes)
- kind: source
- quote: `"Preselect your recommendation with `checked`/`selected`/`value`."` and `"One-of: `<input type="radio">`; many-of: `type="checkbox"`; long lists: `<select>`."`
- url: /src/prompts/tools/ask.md
- read: 2026-11-18T06:14:28Z
- sha256: bbe02b1a49b9f5764ccf05aea132caf97f8456fdd441ca346535daaa7e5a9a63 (fetched 2026-11-18T06:24:38Z, 2000 bytes)
- kind: source

### "Other" and buttons

- quote: `"Radio and checkbox groups automatically get an "Other" free-text answer: its text arrives as the radio's value, or as an extra checkbox array entry. NEVER add your own "Other" option or a text field for it."` and `"NEVER add buttons (`<button>`, `type="submit"`): the UI supplies Submit and Cancel."`
- url: /src/prompts/tools/ask.md
- read: 2026-11-18T06:14:28Z
- sha256: bbe02b1a49b9f5764ccf05aea132caf97f8456fdd441ca346535daaa7e5a9a63 (fetched 2026-11-18T06:24:38Z, 2000 bytes)
- kind: source
- quote: `"Asks the interactive user to fill in one HTML form and returns the submitted form data."`
- url: https://raw.githubusercontent.com/can1357/oh-my-pi/main/docs/tools/ask.md
- read: 2026-11-18T06:16:52Z
- sha256: f72045b671cb289c43cc721c5c588c89159d84124a9558472a44bbf0200a0f4f (fetched 2026-11-18T06:24:39Z, 9192 bytes)
- kind: docs

### The `ask.timeout` setting

- quote: `id: "ask.timeout", type: "number", default: 0` — 0 disables it; plan mode always disables it (`const timeout = planModeEnabled ? undefined : askTimeoutMs(this.session.settings);`).
- url: /src/modes/settings.ts
- read: 2026-11-18T06:14:28Z
- sha256: eeabc90cad6cac1e0337889b6a507907b1561c95fc2577a2f76fd4bbcd9eb4c4 (fetched 2026-11-18T06:24:39Z, 46088 bytes)
- kind: source
- quote: `"Submit an `ask` form's current values (the model's preselected answers) after this many seconds; `0` disables it. Also bounds `cfg://` approval prompts; an unanswered approval denies the write instead of auto-submitting."`
- url: https://raw.githubusercontent.com/can1357/oh-my-pi/main/docs/settings.md
- read: 2026-11-18T06:16:55Z
- sha256: 619bda937cba036da0d65b4671958bab0d77112d4eaa59d5333f22bb83c5f4be (fetched 2026-11-18T06:24:39Z, 258832 bytes)
- kind: docs
- quote: `"A timeout submits the form's current values: in the fallback the unanswered fields keep their defaults and a checkbox group keeps what was checked so far."`
- url: https://raw.githubusercontent.com/can1357/oh-my-pi/main/docs/tools/ask.md
- read: 2026-11-18T06:16:52Z
- sha256: f72045b671cb289c43cc721c5c588c89159d84124a9558472a44bbf0200a0f4f (fetched 2026-11-18T06:24:39Z, 9192 bytes)
- kind: docs

So a form left unanswered is submitted with whatever the model preselected: that is why a case-(2) question presets nothing (`questions.md` rule 3). The installed settings label for `ask.timeout` still says "Auto-select the recommended ask option"; the code and the upstream docs say "submit the current values".

### Print mode, RPC and subagents

- quote: `const isInteractive = !parsedArgs.print && !autoPrint && parsedArgs.mode === undefined;` — in print mode there is no UI.
- url: /src/main.ts
- read: 2026-11-18T06:14:28Z
- sha256: 98c5c0616d0c85417eea797ebecbeeecebbf432ce3f5a6535604b7aba08af254 (fetched 2026-11-18T06:24:39Z, 122638 bytes)
- kind: source
- quote: `return (session.canPromptUser ?? session.hasUI) || session.advertiseHeadlessAsk ? new AskTool(session) : null;` (line 371) — without a UI the tool is not offered; a call without one throws `ToolAbortError("Ask tool requires interactive mode")` (line 402).
- url: /src/tools/ask.ts
- read: 2026-11-18T06:14:28Z
- sha256: 890c65a5a0eee2f79a14b7f3eacf2620fb66a9a68ce96ba2964ba7b6e07bc9f4 (fetched 2026-11-18T06:24:38Z, 19004 bytes)
- kind: source
- quote: `"In `rpc-ui`, tool UI such as `ask` still sends `extension_ui_request` frames for the host to answer."`
- url: https://raw.githubusercontent.com/can1357/oh-my-pi/main/docs/cli-reference.md
- read: 2026-11-18T06:17:02Z
- sha256: 699102a24b406d31aa3def22df9730a22c6e658b05b18fd83275917780135543 (fetched 2026-11-18T06:24:39Z, 68127 bytes)
- kind: docs
- quote: child sessions of the task tool are created with `hasUI: false` (`src/task/executor.ts:5005`), so `AskTool.createIf` returns null for a subagent.
- url: /src/task/executor.ts
- read: 2026-11-18T06:14:28Z
- sha256: 80841ac94f463426c67201151fdb0a3e413b235c383560a04427d43a6d446efd (fetched 2026-11-18T06:24:39Z, 226331 bytes)
- kind: source

### Not confirmed

- unverified: no live `omp` print-mode run or subagent run was made; the upstream docs state no sentence about `ask` in subagents (read from source, so inferred); whether `docs/tools/ask.md` on `main` differs from the 18.18.7 tag.

## Codex run

- status: not_run — a real `$moai project` run in Codex, with the questions arriving through the dialog, was not made
- reason: the run needs a Codex session that calls a model with the user's own account (the prompt leaves this machine) and a person who picks in the Plan-mode dialog; a lane may not decide the first, and `codex exec` cannot show the dialog (see the Codex CLI section), so a scripted run would prove nothing.

What was measured instead is read-only, calls no model and writes nothing outside this worktree's ignored `target/` folder. The same installed binary renders what the model is given, so the claims of the table row can be checked against it. The commands ran in an empty temporary git project.

- command: `codex --version` and `codex features list`
- read: 2026-11-18T06:17:47Z
- output: `codex-cli 0.170.0` and `default_mode_request_user_input          under development  false`

- command: `codex debug prompt-input "hi"` (exit 0), the tool `request_user_input` in the rendered tool list
- read: 2026-11-18T06:17:51Z
- output: `Request user input for one to three short questions and wait for the response. This tool is only available in Plan mode.`

- command: `codex debug prompt-input --enable default_mode_request_user_input "hi"` (exit 0), the same tool
- read: 2026-11-18T06:17:51Z
- output: `Request user input for one to three short questions and wait for the response. This tool is only available in Default or Plan mode.`

What this shows: the tool is in the model's tool list under the name `request_user_input`; with the flag off, as shipped in `codex-cli 0.170.0`, it is for Plan mode only; with the flag on, Default mode is allowed too. The rendered schema asks for options labelled "(Recommended)" first, 2 or 3 per question, no "Other" of the model's own. The rendered Default-mode instruction says `If `request_user_input` returns no answers, continue with best judgment`. The rendered input also lists `request_user_input_async`, which `questions.md` forbids for a question a step waits on. What it does not show: the dialog on screen, a person's pick, or the model's use of the tool while following the MoAI route.

- next: a person runs `/plan` in Codex in a temporary project and enters `$moai project`; the questions must arrive as a dialog with the recommended option first. Capture the transcript as the evidence this section lacks.
