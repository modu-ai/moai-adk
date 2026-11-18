# Research: the question tools of Codex CLI and OMP

SPEC-20261118-codex-question-tool. Facts behind the Codex CLI and OMP rows of `skills/moai/references/internal/questions.md`: each fact has its quote, source and the UTC time it was read; `kind` says whether a fact comes from documentation or from source code. A fact not found is marked `unverified`.

## Codex CLI

Source kind: the official Codex documentation at `learn.chatgpt.com/docs` (the old `developers.openai.com/codex/...` addresses redirect there) does not mention the tool at all: a search of its config reference, developer-commands and build-skills pages for `user_input` and `default_mode` found nothing (read 2026-11-18T06:15:45Z). Every fact below therefore rests on the open-source repository at the tag of the installed version, `rust-v0.170.0` (latest stable release, published 2026-11-18T03:40:48Z; read 2026-11-18T06:14:38Z from `https://api.github.com/repos/openai/codex/releases/latest`). The base URL `…` is `https://raw.githubusercontent.com/openai/codex/rust-v0.170.0/codex-rs`. Each fact says `source`. The facts were read by a research agent (not this session) and the read times are the ones it measured with `date -u`.

### Tool name and the async sibling

- quote: `pub const REQUEST_USER_INPUT_TOOL_NAME: &str = "request_user_input";`
- url: https://raw.githubusercontent.com/openai/codex/rust-v0.170.0/codex-rs/core/src/tools/handlers/request_user_input_spec.rs
- read: 2026-11-18T06:14:42Z
- kind: source
- quote: `const TOOL_NAME: &str = "request_user_input_async";` and the result `{"accepted":true}` — it returns at once and the reply arrives later as a new user message; it is registered only when the model catalog advertises it and the thread is the root thread.
- url: https://raw.githubusercontent.com/openai/codex/rust-v0.170.0/codex-rs/core/src/tools/handlers/request_user_input_async.rs
- read: 2026-11-18T06:14:46Z
- kind: source

`request_user_input` is the tool a step can wait on. `request_user_input_async` is a different tool that waits for nobody, which is why `questions.md` forbids it for a question a step waits on.

### Questions, options, "Other"

- quote: `"Request user input for one to three short questions and wait for the response. This tool is only available in {allowed_modes}."` and `"Provide 2-3 mutually exclusive choices. Put the recommended option first and suffix its label with \"(Recommended)\". Do not include an \"Other\" option in this list; the client will add a free-form \"Other\" option automatically."`
- url: https://raw.githubusercontent.com/openai/codex/rust-v0.170.0/codex-rs/core/src/tools/handlers/request_user_input_spec.rs
- read: 2026-11-18T06:14:42Z
- kind: source
- note: the counts (1 to 3 questions, 2 to 3 options, `header` of 12 characters or fewer) are guidance in the schema text; the code enforces only that every question has a non-empty option list. The tool marks each question `is_other = true`, so the client adds the free-text choice itself.

### Modes and the feature flag

- quote: `pub const fn allows_request_user_input(self) -> bool { matches!(self, Self::Plan) }`
- url: https://raw.githubusercontent.com/openai/codex/rust-v0.170.0/codex-rs/protocol/src/config_types.rs
- read: 2026-11-18T06:15:06Z
- kind: source
- quote: `mode.allows_request_user_input() || (features.enabled(Feature::DefaultModeRequestUserInput) && *mode == ModeKind::Default)`
- url: https://raw.githubusercontent.com/openai/codex/rust-v0.170.0/codex-rs/tools/src/tool_config.rs
- read: 2026-11-18T06:15:03Z
- kind: source
- quote: `key: "default_mode_request_user_input", stage: Stage::UnderDevelopment, default_enabled: false` — the same values on the `main` branch (read 2026-11-18T06:15:52Z).
- url: https://raw.githubusercontent.com/openai/codex/rust-v0.170.0/codex-rs/features/src/lib.rs
- read: 2026-11-18T06:14:49Z
- kind: source

So the key is `features.default_mode_request_user_input`, default: false, stage: under development. The tool is registered in every mode; a call from a mode that is not allowed gets an error such as `request_user_input is unavailable in Default mode`.

### No answer, and what the Default-mode text tells the model

- quote: `"If `request_user_input` returns no answers, continue with best judgment instead of asking again or treating the turn as blocked."` and `"In Default mode, strongly prefer making reasonable assumptions and executing the user's request rather than stopping to ask questions."` and `"Use the `request_user_input` tool only when it is listed in the available tools for this turn."`
- url: https://raw.githubusercontent.com/openai/codex/rust-v0.170.0/codex-rs/collaboration-mode-templates/templates/default.md
- read: 2026-11-18T06:14:59Z
- kind: source
- quote: `"request_user_input was cancelled before receiving a response"` — and a client or connection error makes the app server submit an empty answer map, which the tool returns as a success with no answers; the core has no timeout of its own.
- url: https://raw.githubusercontent.com/openai/codex/rust-v0.170.0/codex-rs/core/src/session/mod.rs
- read: 2026-11-18T06:15:15Z
- kind: source

This is why a case-(2) question must not go on by best judgment when nothing comes back (`questions.md` rule 3): Codex's own text asks for the opposite.

### Skill instructions, `codex exec`, subagents

- quote: the skills documentation never names `request_user_input`; its only tool hook is `"to declare tool dependencies for a more seamless experience with using the skill."` (`dependencies: tools:` in `agents/openai.yaml`, MCP type in the example). A skill therefore cannot make the call; it works only when the tool is in the turn's tool list.
- url: https://learn.chatgpt.com/docs/build-skills
- read: 2026-11-18T06:15:45Z
- kind: docs
- quote: `"request_user_input is not supported in exec mode for thread `{}`"` (error code -32000)
- url: https://raw.githubusercontent.com/openai/codex/rust-v0.170.0/codex-rs/exec/src/lib.rs
- read: 2026-11-18T06:15:15Z
- kind: source
- quote: `"request_user_input can only be used by the root thread"`
- url: https://raw.githubusercontent.com/openai/codex/rust-v0.170.0/codex-rs/core/src/session/mod.rs
- read: 2026-11-18T06:15:04Z
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
- kind: source
- quote: `"Preselect your recommendation with `checked`/`selected`/`value`."` and `"One-of: `<input type="radio">`; many-of: `type="checkbox"`; long lists: `<select>`."`
- url: /src/prompts/tools/ask.md
- read: 2026-11-18T06:14:28Z
- kind: source

### "Other" and buttons

- quote: `"Radio and checkbox groups automatically get an "Other" free-text answer: its text arrives as the radio's value, or as an extra checkbox array entry. NEVER add your own "Other" option or a text field for it."` and `"NEVER add buttons (`<button>`, `type="submit"`): the UI supplies Submit and Cancel."`
- url: /src/prompts/tools/ask.md
- read: 2026-11-18T06:14:28Z
- kind: source
- quote: `"Asks the interactive user to fill in one HTML form and returns the submitted form data."`
- url: https://raw.githubusercontent.com/can1357/oh-my-pi/main/docs/tools/ask.md
- read: 2026-11-18T06:16:52Z
- kind: docs

### The `ask.timeout` setting

- quote: `id: "ask.timeout", type: "number", default: 0` — 0 disables it; plan mode always disables it (`const timeout = planModeEnabled ? undefined : askTimeoutMs(this.session.settings);`).
- url: /src/modes/settings.ts
- read: 2026-11-18T06:14:28Z
- kind: source
- quote: `"Submit an `ask` form's current values (the model's preselected answers) after this many seconds; `0` disables it. Also bounds `cfg://` approval prompts; an unanswered approval denies the write instead of auto-submitting."`
- url: https://raw.githubusercontent.com/can1357/oh-my-pi/main/docs/settings.md
- read: 2026-11-18T06:16:55Z
- kind: docs
- quote: `"A timeout submits the form's current values: in the fallback the unanswered fields keep their defaults and a checkbox group keeps what was checked so far."`
- url: https://raw.githubusercontent.com/can1357/oh-my-pi/main/docs/tools/ask.md
- read: 2026-11-18T06:16:52Z
- kind: docs

So a form left unanswered is submitted with whatever the model preselected: that is why a case-(2) question presets nothing (`questions.md` rule 3). The installed settings label for `ask.timeout` still says "Auto-select the recommended ask option"; the code and the upstream docs say "submit the current values".

### Print mode, RPC and subagents

- quote: `const isInteractive = !parsedArgs.print && !autoPrint && parsedArgs.mode === undefined;` and `return (session.canPromptUser ?? session.hasUI) || session.advertiseHeadlessAsk ? new AskTool(session) : null;` — in print mode the tool is not offered; a call without a UI throws `ToolAbortError("Ask tool requires interactive mode")`.
- url: /src/main.ts
- read: 2026-11-18T06:14:28Z
- kind: source
- quote: `"In `rpc-ui`, tool UI such as `ask` still sends `extension_ui_request` frames for the host to answer."`
- url: https://raw.githubusercontent.com/can1357/oh-my-pi/main/docs/cli-reference.md
- read: 2026-11-18T06:17:02Z
- kind: docs
- quote: child sessions of the task tool are created with `hasUI: false` (`src/task/executor.ts:5005`), so `AskTool.createIf` returns null for a subagent.
- url: /src/task/executor.ts
- read: 2026-11-18T06:14:28Z
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
