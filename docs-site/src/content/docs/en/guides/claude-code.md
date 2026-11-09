---
title: Claude Code
description: Diagnose Claude Code instruction loading and preview project settings and launcher commands without starting a session.
---

Use this guide to check what Rust `moai-cli` can prepare for Claude Code. Planning a file or predicting instruction loading is not evidence that a live session loaded the `moai` skill.

## Prerequisites and support

Complete [getting started](/en/guides/getting-started/) and keep its absolute `MOAI_CLI` and `MOAI_SOURCE` variables. For diagnostics, install Claude Code separately using its own official instructions. `claude-doctor` can inspect its version, but does not install or authenticate it.

From the project you want to inspect, set the target explicitly:

```sh
export MOAI_PROJECT="$PWD"
"$MOAI_CLI" claude-doctor --help
"$MOAI_CLI" cc --help
```

The RC target is Darwin/arm64. The repository recorded Claude Code **2.1.319** and static skill validation; live `/moai` discovery, invocation, and permission-denial E2E remain **`not_run`**, not certified support. Linux, Windows, other architectures, and cloud sessions are not established by that evidence.

## Diagnose instruction loading

```sh
"$MOAI_CLI" claude-doctor "$MOAI_PROJECT"
```

The doctor reads instruction files and relevant settings, and may run `claude --version`. It does not start a model session, create an import, delete a `CLAUDE.md`, or change settings. Inspect `result.agents_md_status`, `reason`, `load_path`, `basis`, and `suggestions`.

- Direct AGENTS loading needs Claude Code **2.1.277+**. An unavailable version yields uncertainty, not proof of support.
- Under the default instruction-file policy, a `CLAUDE.md`, `.claude/CLAUDE.md`, or `CLAUDE.local.md` in the inspected scope can suppress direct `AGENTS.md` loading.
- An existing `@AGENTS.md` import can provide an import path. The doctor proposes alternatives but applies neither.
- `basis: predicted` is important: managed policy and separate `--settings` files are not checked. Even a status named `loaded` is a prediction, not a live-session receipt.

Preserve your existing instructions. Review any proposed import or policy change with the project owner instead of deleting a conflicting file to make the doctor green.

## Preview project settings

The Rust settings interface plans against **project** `.claude/settings.json`. For example, preview selecting Claude's built-in Default output style:

```sh
printf '%s\n' '{"outputStyle":"Default"}' |
  "$MOAI_CLI" claude-settings-plan "$MOAI_PROJECT"
```

This writes nothing. Read the planned action, conflicts, expected/current digests, and `plan_digest`. A differing user-owned value can produce a conflict; do not work around it by deleting ownership records. The plan refuses a target that is HOME or whose `.claude` directory is `CLAUDE_CONFIG_DIR`.

`claude-settings-apply` is the separate mutating operation: it takes the same desired JSON, `--plan-digest`, and `--permission-reference`, re-plans, and writes project settings plus `.moai-cli/manifests/claude-settings.json`. This guide does not invoke it. The reference is not authorization; approve the exact diff first. Other user-owned keys must be preserved.

A settings merge is **not** an asset installer. The Rust replacement does not copy the provisional Python adapter's output-style files, `statusline.py`, or licenses. Do not point `statusLine.command` at a nonexistent script or assume an output style is present because its name appears in settings. The Rust binary does not call that Python adapter. Existing hooks and permissions still need review; this RC does not generate a MoAI hook set.

## Preview the launcher

The launcher is a repository-development convenience, not a skill installer. Preview a plain session with no MoAI role and no experimental advisor:

```sh
"$MOAI_CLI" cc --no-advisor --permissions --dry-run
```

Inspect `argv`, `notes`, `executed`, and `host_loaded`. `--dry-run` composes JSON only: it does not start Claude Code, send a prompt, or verify credentials. Plain `cc` sessions do not add a permission-skip flag. `--permissions` explicitly requests that the launcher not add one; it does **not** override permissive settings already configured in Claude Code.

The Claude launcher defaults are model `opus`, effort `medium`, and advisor `fable`. `--no-advisor` removes the advisor option, which the source marks experimental and Anthropic-API-only. Removing only `--dry-run` from the reviewed command starts a real host in your current directory. Do that only when you intentionally accept the host's authentication, hooks, external transmission, and costs. Review existing project, local, user, and managed permission settings first.

Do not use deprecated `-l`/`-a` MoAI launcher roles for new sessions. They retain legacy behavior; use a plain session and select a skill mode inside it instead.

## Use a discovered skill

Building `moai-cli` or opening `cc` does not make `/moai` appear. Claude Code must first discover a reviewed `moai` skill in one of its skill locations. This guide does not copy anything into `~/.claude/skills` or a project's `.claude/skills`; home/project installation and live loading remain separate steps with explicit approval.

If a reviewed skill is already available, enter a read-only request **in Claude Code**, not in your shell:

```text
/moai Summarize this repository's current state. Do not change files, start implementation, or publish anything.
```

That request is guidance, not a sandbox. For authorized workflow participation, the current forms are `/moai --mode leader` (default autonomy `semi`) or `/moai --mode agent`. Agent mode begins a self-claim round and can post a GitHub claim; it is not a read-only test. Leader mode is Claude-only in this RC and does not waive independent audit or user decisions. Neither mode grants permissions from repository text.

## Troubleshooting and next steps

- **AGENTS suppressed:** read the doctor's reason and preserve existing CLAUDE instructions; approve a proposed remedy separately.
- **Skill absent:** check the host's actual skill discovery, then follow its reload/restart guidance. Binary availability and settings success are not skill-loading evidence.
- **Settings conflict or stale plan:** inspect the changed bytes and re-plan; do not fabricate a permission reference or reuse an old digest.
- **Host not launched:** a successful `--dry-run` verifies composition only, not host installation, authentication, model availability, or E2E.

See [CLI reference](/en/reference/cli/), [safety and permissions](/en/reference/safety/), and the [Codex guide](/en/guides/codex/) for the other adapter.

## Source notes

Command forms and behavior are pinned to the [CLI declarations](https://github.com/modu-ai/moai-adk/blob/cf33967c42c31f020fd0e35dc07b6b82093119f1/src/main.rs), [launcher implementation](https://github.com/modu-ai/moai-adk/blob/cf33967c42c31f020fd0e35dc07b6b82093119f1/src/session.rs), [Claude adapter contract](https://github.com/modu-ai/moai-adk/blob/cf33967c42c31f020fd0e35dc07b6b82093119f1/docs/claude-adapter-contract.md), and [support matrix](https://github.com/modu-ai/moai-adk/blob/cf33967c42c31f020fd0e35dc07b6b82093119f1/docs/support-matrix.md). Current mode routing is in the [common skill](https://github.com/modu-ai/moai-adk/blob/cf33967c42c31f020fd0e35dc07b6b82093119f1/skills/moai/SKILL.md).

The repository contract records official [memory](https://code.claude.com/docs/en/memory) and [skills](https://code.claude.com/docs/en/skills) documentation read on 2026-11-08. Those documented capabilities are not new live-host measurements made for this site.
