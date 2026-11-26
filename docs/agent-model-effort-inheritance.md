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

Limits of these runs:
- One role, one model, one run; Claude Code only at effort `high`.
- Not measured: a child started on a model other than the parent's.
- The Claude transcript fields are not in the product's official documentation; they are what
  the files held on the day.

## Where it can break (from the products' documentation, read 2026-11-16)

Claude Code (code.claude.com/docs/en/sub-agents, model-config, env-vars):
- The model order is: the `model` parameter of the Agent call, the `model` field of the
  definition (including `inherit`), `CLAUDE_CODE_SUBAGENT_MODEL`, then the parent's model.
  `CLAUDE_CODE_SUBAGENT_MODEL_FORCE=1` makes the variable win over the first two.
- A definition without `effort` follows the session effort only on the parent's model; on
  another model the documentation gives no fallback. `CLAUDE_CODE_EFFORT_LEVEL` outranks both
  the call and the definition.

## What the repository does

- A rendered agent file holds only `name`, `description` and the host's own fields (`tools`,
  `disallowedTools`, `isolation`, `developer_instructions`). The test
  `rendered_agent_files_name_no_model_and_no_effort` (`tests/agent_defs_package.rs`) fails when
  a rendered file gains any other key, so a model or effort key cannot be added unseen.
- The launcher passes `--model` and `--effort` to a Claude parent.
- Installing agent files into a real home and any user setting stay outside this page.
