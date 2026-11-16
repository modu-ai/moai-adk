# Agent model and effort: what a child starts on

Issue #3096 S7 (user instruction, #1975 6904941797 §B). The base agents (`agents/*.md`,
rendered by `src/agent_defs.rs`) name no model and no effort, so a child agent starts on the
parent session's model and effort. This page records what was measured and where the
inheritance can break. It changes no user setting.

## Measured 2026-11-16 (same machine, one run each)

| Product | Parent | Child (`moai-research`, no model/effort key) | Read from |
| --- | --- | --- | --- |
| Claude Code 2.1.325 | `claude-sonnet-5-5`, effort `high` | same, two children | `model` and `effort` of the assistant records in the session transcript |
| Codex 0.169.0, parent run with `-c model_reasoning_effort="low"` | `gpt-6.1-terra`, `low` | same | `turn_context` of the new rollout files (child marked by `session_meta.source.subagent.thread_spawn`) |
| Codex 0.169.0, parent run with `-c model_reasoning_effort="high"` | `gpt-6.1-terra`, `high` | same | same |

Not measured: OMP (no live session); a child started on a model other than the parent's.
The Claude transcript fields and the Codex rollout fields are not in the products' official
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
  the new model's default.

## What the repository does

- A rendered agent file holds only `name`, `description` and the host's own fields (`tools`,
  `disallowedTools`, `isolation`, `developer_instructions`). The test
  `rendered_agent_files_name_no_model_and_no_effort` (`tests/agent_defs_package.rs`) fails when
  a rendered file gains any other key, so a model or effort key cannot be added unseen.
- The launcher passes `--model` and `--effort` to a Claude parent and `-m` to a Codex parent;
  it does not pass an effort to Codex (`NOTE_EFFORT_NOT_PASSED_CODEX`). The child follows its
  parent either way, so the launcher is unchanged.
- Installing agent files into a real home and any user setting stay outside this page.
