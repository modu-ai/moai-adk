# Agent model and effort: what a child starts on

Issue #3096 S7 (user instruction, #1975 6904941797 §B). The base agents (`agents/*.md`,
rendered by `src/agent_defs.rs`) name no model and no effort. In the runs below, a child of
such a file showed its parent's model and effort; that is all that was measured. This page
records those runs and where the products' documentation says inheritance can break. It
changes no user setting.

## Measured 2026-11-16 (same machine, one run each, role `moai-research` only)

| Product | Parent | Child (`moai-research`, no model/effort key) | Read from |
| --- | --- | --- | --- |
| Claude Code 2.1.325 | `claude-sonnet-5-5`, effort `high` | same, two children | `model` and `effort` of the assistant records in the session transcript |
| Codex 0.169.0, parent run with `-c model_reasoning_effort="low"` | `gpt-6.1-terra`, `low` | same | `turn_context` of the new rollout files (child marked by `session_meta.source.subagent.thread_spawn`) |
| Codex 0.169.0, parent run with `-c model_reasoning_effort="high"` | `gpt-6.1-terra`, `high` | same | same |

Limits of these runs:
- One role, one model per product, one run per row; Claude Code only at effort `high`.
- The Codex child rollout holds a copy of the parent's records, so the child's own
  `turn_context` could not be told apart from the copy. A differing value would have shown as
  a separate line; every line was of one kind. This is not a proof of the child's own values.
- The Codex parent got its effort from `-c model_reasoning_effort=…`. The launcher does not
  start Codex that way (it passes `-m` only, no effort), so the launcher's own path, and a
  parent whose effort nobody set, were not measured.
- Not measured: OMP (no live session); a child started on a model other than the parent's.
- The Claude transcript fields and the Codex rollout fields are not in the products' official
  documentation; they are what the files held on the day.

## Where it can break (from the products' documentation, read 2026-11-16)

Claude Code (code.claude.com/docs/en/sub-agents, model-config, env-vars):
- The model order is: the `model` parameter of the Agent call, the `model` field of the
  definition (including `inherit`), `CLAUDE_CODE_SUBAGENT_MODEL`, then the parent's model.
  `CLAUDE_CODE_SUBAGENT_MODEL_FORCE=1` makes the variable win over the first two.
- A definition without `effort` follows the session effort only on the parent's model; on
  another model the documentation gives no fallback. `CLAUDE_CODE_EFFORT_LEVEL` outranks both
  the call and the definition.

Codex (developers.openai.com/codex/subagents, config-reference):
- `[agents] default_subagent_model` and `default_subagent_reasoning_effort` in `config.toml`,
  and the `model` and `reasoning_effort` arguments of `spawn_agent`, replace the parent's
  value; an agent file that sets a key wins over both.
- An agent file that sets only `model` keeps the effort resolved before it; it does not take
  the new model's default. Source: the subagents page for the behaviour and `child_config.rs`
  on the `main` branch of github.com/openai/codex, read 2026-11-16; that `main` matches the
  installed 0.169.0 is unverified.

## What the repository does

- A rendered agent file holds only `name`, `description` and the host's own fields (`tools`,
  `disallowedTools`, `isolation`, `developer_instructions`). The test
  `rendered_agent_files_name_no_model_and_no_effort` (`tests/agent_defs_package.rs`) fails when
  a rendered file gains any other key, so a model or effort key cannot be added unseen.
- The launcher passes `--model` and `--effort` to a Claude parent and `-m` to a Codex parent;
  it does not pass an effort to Codex (`NOTE_EFFORT_NOT_PASSED_CODEX`). This PR leaves the
  launcher unchanged. The Codex runs above did not use the launcher's path, so what a child
  shows under it is not measured here; the Codex launcher effort is a separate piece (#3096 S7b).
- Installing agent files into a real home and any user setting stay outside this page.
