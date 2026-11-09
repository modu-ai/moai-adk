---
title: Codex
description: Inspect Codex instruction discovery and preview a moai skill projection into an explicit project target without installing or starting a model.
---

Rust `moai-cli` provides a read-only Codex doctor and a plan/apply interface for copying the common `moai` skill. File verification, Codex discovery, and actual execution are separate outcomes.

## Prerequisites and support

Complete [getting started](/en/guides/getting-started/) and retain `MOAI_CLI` and `MOAI_SOURCE` as absolute paths. Codex itself is a separate prerequisite for an eventual live session; none of the previews below installs or authenticates it.

From the root of the project you want to inspect:

```sh
export MOAI_PROJECT="$PWD"
"$MOAI_CLI" codex-doctor --help
"$MOAI_CLI" codex-projection-plan --help
"$MOAI_CLI" codex --help
```

The RC target is Darwin/arm64. The repository recorded `codex-cli 0.163.0`, but a version observation is not certification. Desktop, CLI, and IDE `$moai` discovery/invocation remain **`not_run`**. Linux, Windows, other architectures, and cloud/worktree host E2E are not certified by the local fixtures.

## Diagnose instruction loading

```sh
"$MOAI_CLI" codex-doctor "$MOAI_PROJECT" --cwd "$MOAI_PROJECT"
```

This reads files without starting Codex, calling a model, or writing configuration. Review these fields under `result`:

| Field | What to check |
| --- | --- |
| `agents.global` | Which global AGENTS file was found under Codex home. |
| `agents.project_chain` and `agents.included` | Root-to-working-directory instruction selection. |
| `agents.over_budget` | Whether the predicted instruction size exceeds the supplied/default budget. |
| `skills.moai` | Discovered project, user, or administrator copies and duplicate warnings. |
| `trust` | `not_evaluated` is not `trusted`. |
| `codex_layer` | Presence of project config/hooks, not proof that Codex loads them. |

Codex documents `AGENTS.override.md` before `AGENTS.md`, followed by configured fallback names for project directories, with nearer instructions later in the chain. The default combined project-document budget is 32 KiB. The doctor accepts `--fallback NAME` (repeatable) and `--max-bytes N` to model your known settings; **it does not read `config.toml`**. If you omit those options, its assumptions may differ from your actual configuration.

Untrusted projects can have their project `.codex/` configuration, hooks, and rules skipped by Codex. The doctor does not evaluate or change trust. Do not grant trust solely to clear a diagnostic. An instruction file's presence is not a sandbox, and it is not a permission grant.

## Preview a project skill projection

Use an explicit absolute target ending in `moai`, outside the source package. The following proposes a **project** location, but does not create it:

```sh
export MOAI_PROJECTION="$MOAI_PROJECT/.agents/skills/moai"
"$MOAI_CLI" codex-projection-plan "$MOAI_SOURCE/skills/moai" \
  --target "$MOAI_PROJECTION" \
  --allow-implicit-invocation false
```

`--allow-implicit-invocation false` plans a policy that requires explicit skill invocation; it does not disable all Codex automation or change host permissions. Both the target and this option affect the plan, so retain them if you later request an approved apply.

Inspect `package_digest`, `plan_digest`, `source_manifest`, `files`, and `preserved`:

- Every planned file has a digest and a target decision: `create`, `unchanged`, or `conflict`.
- `SKILL.md` remains byte-identical to the common package. The adapter also plans `agents/openai.yaml` metadata and the explicitly requested invocation policy.
- The source manifest comparison can report `verified`, `mismatch`, `missing`, or `unreadable`. A plan reporting a problem is not permission to install it.
- Unknown target files are preserved; conflicting target files are not an invitation to overwrite or delete them.
- No files or hooks are created by the plan. `host_loaded: false` and `execution_verified: false` do not become true because the plan succeeds.

The CLI deliberately has no default `--target`. Home versus project placement is not silently decided. This example neither changes `~/.agents/skills` nor resolves the repository's pending placement policy. A home copy and a project copy with the same name can both appear in selectors; the doctor warns rather than merging or selecting one for you.

`codex-projection-apply` is separate. It requires the same package, target, metadata/policy options, the reviewed `--plan-digest`, and an approved `--permission-reference`. It re-plans, refuses a non-verified source manifest or any target conflict, and creates missing files exclusively. A failure after some creates can be partial; this is not an all-or-nothing installation. Its `verified` result compares installed bytes, **not** live host loading. This guide does not invoke apply or remove existing copies. See [safety and permissions](/en/reference/safety/) before considering it.

## Preview the launcher

```sh
"$MOAI_CLI" codex --dry-run
```

Read `argv`, `notes`, `executed`, and `host_loaded`. The launcher composes a plain session using Codex `workspace-write` sandbox mode and `on-request` approvals. It starts no process and sends no prompt in `--dry-run` mode. Composition is not a probe of authentication, model availability, or sandbox effectiveness.

Removing `--dry-run` starts a real Codex CLI in the current directory. That host can write within its allowed workspace, use existing configuration/hooks, send context externally, and incur costs. Only start it after reviewing the target, existing configuration, and approvals. The launcher does not install skills or make the project trusted. Claude-only advisor and effort controls are not passed through to Codex.

Do not use deprecated MoAI launcher `-l`/`-a` roles for new sessions. The `-a on-request` inside the composed **Codex** argv is Codex's approval policy, not the deprecated MoAI role flag.

## Use a discovered skill

Once a reviewed skill is genuinely discoverable, type `$` or use `/skills` in Codex CLI to select it. Enter a read-only request **inside Codex**, not a shell where `$moai` would be expanded as a variable:

```text
$moai Summarize this repository's current state. Do not change files, claim an Issue, or publish anything.
```

The request is guidance, not enforced isolation. For approved workflow participation, the current role form is `$moai --mode agent`. Agent mode begins a self-claim round and can post a GitHub claim, so do not use it as a read-only discovery test. Leader mode is Claude Code-only in this RC; Codex must not enter it. Codex sessions have no RC cross-session `ListAgents`/`SendMessage` channel; durable Issue comments and Markdown handoffs are separate, authorized effects.

## Troubleshooting and next steps

- **Duplicate skills:** inspect the paths reported by the doctor; do not delete an existing copy without an ownership and scope decision.
- **Manifest or target conflict:** resolve the reviewed source/target mismatch and make a fresh plan. Do not weaken the check or invent approval metadata.
- **Unexpected instructions:** compare the actual cwd, override files, fallback names, and byte budget with the doctor's inputs.
- **Config absent from the session:** project trust may be relevant, but the doctor cannot determine it. This CLI has no Codex config plan/apply subcommand; the project's `codex_config` implementation is a library surface, not a command to invent.
- **Skill files exist but no `$moai`:** check discovery in the real host and its restart guidance. A successful projection is not a live invocation receipt.

Continue with the [CLI reference](/en/reference/cli/), [safety reference](/en/reference/safety/), or [Claude Code guide](/en/guides/claude-code/).

## Source notes

These instructions use the pinned [CLI declarations](https://github.com/modu-ai/moai-adk/blob/cf33967c42c31f020fd0e35dc07b6b82093119f1/src/main.rs), [launcher](https://github.com/modu-ai/moai-adk/blob/cf33967c42c31f020fd0e35dc07b6b82093119f1/src/session.rs), [Codex adapter contract](https://github.com/modu-ai/moai-adk/blob/cf33967c42c31f020fd0e35dc07b6b82093119f1/docs/codex-adapter-contract.md), [mode workflow](https://github.com/modu-ai/moai-adk/blob/cf33967c42c31f020fd0e35dc07b6b82093119f1/skills/moai/workflows/mode-agent.md), and [support matrix](https://github.com/modu-ai/moai-adk/blob/cf33967c42c31f020fd0e35dc07b6b82093119f1/docs/support-matrix.md).

The contract records official [skill](https://learn.chatgpt.com/docs/build-skills), [AGENTS](https://learn.chatgpt.com/docs/agent-configuration/agents-md), and [configuration](https://learn.chatgpt.com/docs/config-file/config-reference) documentation read on 2026-11-08/09. Documented host behavior is not a new measurement performed for this site.
