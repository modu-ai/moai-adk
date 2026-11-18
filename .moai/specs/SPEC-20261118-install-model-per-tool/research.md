# Research: models and effort levels per tool (card C-47, AC seq 2)

Read 2026-11-18 between 06:14Z and 06:16Z by three read-only research agents; every time below is a `date -u` taken in the same step as the read. Web pages came through a fetch tool that returns an extract, so a web quote is an extract, not raw page bytes. Nothing here was run in a live tool session: a value in a catalog is "listed", not "accepted at runtime".

## Claude Code (installed 2.1.327)

| Name | `--model` value | Efforts | Source |
| --- | --- | --- | --- |
| Opus 5.5 | `claude-opus-5-5` (alias `opus`) | low, medium, high, xhigh, max | model-config page, models overview |
| Sonnet 5.5 | `claude-sonnet-5-5` (alias `sonnet`) | low, medium, high, xhigh, max | same |
| Haiku 5.5 | `claude-haiku-5-5` (alias `haiku`) | low, medium, high, xhigh, max | same |
| Fable 5.2 | `claude-fable-5-2` (alias `fable`) | low, medium, high, xhigh, max | same |

- Start: `claude --model claude-opus-5-5 --effort high` (aliases work too).
- Help text (`claude --help`, 06:14:36Z): `--effort <level>  Effort level for the current session (low, medium, high, xhigh, max)`.
- The CLI reference page also lists `ultracode` for `--effort`; `claude --help` does not. Not used here.
- `max` is not accepted in the persisted `effortLevel` setting, only as `--effort max`.
- Precedence: the env var `CLAUDE_CODE_EFFORT_LEVEL` wins over `--effort`. A launcher that passes `--effort` must remove that variable from the child environment, or its value is ignored.
- Claude Code's own default for Sonnet 5.5 is `medium`; the API models page says `high` for the API. Our default `high` therefore has to be passed explicitly.
- Minimum versions: Opus 5.5 v2.1.280, Sonnet 5.5 v2.1.284, Haiku 5.5 v2.1.293.
- Sources: https://code.claude.com/docs/en/cli-reference, https://code.claude.com/docs/en/model-config, https://platform.claude.com/docs/en/about-claude/models/overview, `claude --help`.
- Not verified: a live `claude --model … --effort …` run; the last 35 000 characters of the model-config page.

## Codex CLI (installed codex-cli 0.170.0)

| Display name | Model id (`-m`) | Efforts listed | Default in catalog |
| --- | --- | --- | --- |
| GPT-6.1-Sol | `gpt-6.1-sol` | low, medium, high, xhigh, max, ultra | low |
| GPT-6.1 Terra | `gpt-6.1-terra` | low, medium, high, xhigh, max, ultra | low |
| GPT-6.2-Sol | `gpt-6.2-sol` | low, medium, high, xhigh, max, ultra | low |
| GPT-6-Luna | `gpt-6-luna` | low, medium, high, xhigh, max | medium |

- The card's `gpt-6.1 sol` is not an id; the id is `gpt-6.1-sol`. The card's "GPT-6.1 Terra" is the display name of `gpt-6.1-terra`.
- Start: `codex -m gpt-6.1-sol -c 'model_reasoning_effort="high"'`. There is no dedicated effort flag; `-c key=value` is the generic override (`codex --help`). `config.toml` keys: `model`, `model_reasoning_effort`.
- Docs: "GPT-6.1 Terra requires Codex CLI 0.159.0 or later. To use it, run `codex -m gpt-6.1-terra`." Config reference: "Reasoning effort advertised by the selected model, such as `low`, `medium`, `high`, `xhigh`, `max`, or `ultra`. Available levels depend on the model and client."
- Question tool (for card C-52): `request_user_input`, "This tool is only available in Plan mode"; in Default mode it is behind the feature flag `default_mode_request_user_input` (under development, off). No official documentation of the tool was found; the evidence is the local prompt and catalog.
- Sources: `codex --help`, `codex exec --help`, `codex debug models` (06:15:12Z and 06:15:16Z), `codex debug prompt-input` (06:15:30Z), `codex features list` (06:15:37Z), https://learn.chatgpt.com/docs/models and https://learn.chatgpt.com/docs/config-file/config-reference (reached by redirect from developers.openai.com; the host is server-supplied and was not independently confirmed as OpenAI's).
- Not verified: runtime acceptance of `max` or `ultra`; hidden catalog entries.

## OMP (Oh My Pi 18.18.7) with GLM models

| Model | OMP id | Efforts offered by OMP | Z.ai documents the levels? |
| --- | --- | --- | --- |
| GLM-5.3 | `zai/glm-5.3` | low, high, max (default max; thinking cannot be disabled) | yes: "supports three reasoning effort levels: `low`, `high`, and `max`" |
| GLM-5.3-Flash | `zai/glm-5.3-flash` | low, high, max (OMP catalog) | no: the page lists no levels |

- Start: `omp --model zai/glm-5.3 --thinking high`, or the suffix form `--model zai/glm-5.3:high`. `--thinking` accepts `off, minimal, low, medium, high, xhigh, max, auto`; only `low`, `high`, `max` fit GLM-5.3. A bare `glm-5.3` is fuzzy-matched across providers, so the provider-qualified id is used.
- `--thinking off` on GLM-5.3 is clamped to `low` (`thinking-requires-effort`).
- Config: `~/.omp/agent/config.yml` keys `defaultThinkingLevel` and `modelRoles`. Key env var: `ZAI_API_KEY` for `zai`.
- Question tool (for card C-52): the built-in `ask` tool, interactive mode only; one `<form>` with radio or checkbox inputs, a recommended value preselected with `checked`. A headless fork gets an error.
- Sources: OMP installed source and bundled docs under `@oh-my-pi/pi-catalog` and `@oh-my-pi/pi-coding-agent` (files named in the research agent's report: `glm.kdl`, `zai.kdl`, `model-thinking.ts`, `ask.ts`, `models.md`, `settings.md`, `cli-reference.md`), `omp --help`, `omp models zai`; https://docs.z.ai/guides/llm/glm-5.3, https://docs.z.ai/guides/llm/glm-5.3-flash, https://docs.z.ai/guides/capabilities/thinking-mode (extracts).
- Not verified: how `defaultThinkingLevel: max` behaves on a model without `max` (OMP steps down to the nearest lower level it offers); the `zhipu-coding-plan` provider's ids.
- `moai glm` is configured like OMP with the same two models (card text); its transport is not researched here.
