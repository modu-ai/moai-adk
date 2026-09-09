---
title: "Codex Dual Harness — AGENTS.md, Dual Agent Publication, Hook Adapter"
weight: 31
draft: false
added_in: "v3.1.3"
description: "The four artifacts that let codex-cli read MoAI-ADK — the root AGENTS.md standing contract, dual agent TOML publication, the .agents/skills mirror, and the internal/codexadapter hook adapter library."
---

MoAI-ADK's primary harness (the runtime that actually drives agents) is Claude Code, but as of v3.1.3 it carries a **dual surface that codex-cli can read too**. None of it changes what Claude Code does — the rules and agent definitions that already existed are simply published a second time, in the locations and formats codex looks for. This page covers the four artifacts and the problem each one solves.

## The root AGENTS.md — a standing contract for any harness

The root `AGENTS.md` is a standing contract that binds a turn **regardless of which agent harness drives it** — not a Claude-specific file. It is one file because of how codex reads: it consumes project instructions under a byte cap and **drops the overflow silently, with no warning and exit 0**. A contract that does not fit reports as complete. Fitting under the ceiling is therefore itself a requirement, and a build guard (a check run at build time that the file stays under it) holds it there.

Making room took eleven always-loaded documents down to stubs (short summaries) pointing at eight lazy companions (detail documents read only on demand). **What moved was the prose explaining each obligation, never the obligation itself** — the source of truth remains `.claude/rules/moai/**` and `CLAUDE.md`, and `AGENTS.md` carries it in a harness-neutral form.

{{< callout type="info" >}}
A personal `~/.codex/AGENTS.md` joins the same merged chain and is consumed **before** this file, narrowing what the project's contract can carry. Overflow is dropped from the tail silently — which is why the clauses in this file are ordered most-critical-first.
{{< /callout >}}

## Dual agent publication — eleven TOMLs

The 11 retained agents are published in two forms: `.claude/agents/moai/*.md` for Claude Code (the source) and `.codex/agents/moai/*.toml` for codex (the derivation). The TOML is not hand-written — `internal/template/agentemit` generates it **deterministically** (same input, same output, every time) from the markdown source, and the generated file's header says "regenerate, do not edit".

Three guards keep source and derivation from drifting apart: a golden-file comparison (against expected output), an embed check (against the templates compiled into the binary), and a deploy check (against what lands in user repositories). Edit the markdown and the TOML follows; edit only the TOML and the guards catch it.

## `.agents/skills` — the skill mirror

codex-cli does not read Claude Code's `.claude/skills/`, so skills are deployed as a **mirror** under `.agents/skills`. The mirror list is not hand-maintained — it is derived from the actual skill set at deploy time, so it cannot go stale as skills come and go. The directory is a deployment artifact aimed **outside user repositories** and is never committed; it prefers a symbolic link, falling back to a copy where links cannot be created (the `moai init` / `moai update` completion summaries say so — see the [moai update](/en/cli-reference/update/) page).

## `internal/codexadapter` — the hook adapter library

The two harnesses' hook surfaces are nearly but not exactly the same. Measurement (against codex-cli 0.153.4) found three divergences: the **event name** the harness passes, **three output keys** codex declares but does not act on (`systemMessage`, `continue`, `stopReason`), and the **PreToolUse decision contract** — the codex parser rejects `permissionDecision:allow` without `updatedInput` and rejects `permissionDecision:ask` outright. `internal/codexadapter` is a thin translation layer that sits **in front of** the dispatcher (nothing under `internal/hook` is modified); refused decision shapes (allow, ask, defer) degrade to the no-opinion `{}`, handing the choice to codex's own approval flow, each drop is announced through the discard sink, and a blank-reason deny gains a default reason.

### The 12-event table

| Codex event | MoAI dispatcher arg | Adapted this milestone? |
|---|---|---|
| PreToolUse | `pre-tool` | yes |
| PostToolUse | `post-tool` | yes |
| SessionStart | `session-start` | yes |
| SessionEnd | `session-end` | yes |
| Stop | `stop` | yes |
| UserPromptSubmit | `user-prompt-submit` | yes |
| PreCompact | `compact` | no — compaction never triggered non-interactively |
| PostCompact | `post-compact` | no — compaction never triggered non-interactively |
| PermissionRequest | `permission-request` | no — approval request never raised non-interactively |
| SubagentStart | `subagent-start` | yes |
| SubagentStop | `subagent-stop` | yes |
| Interrupt | — (no counterpart) | no — fires on SIGINT; adaptation needs a new dispatcher subcommand (follow-up) |

Twelve events are covered: eleven have a dispatcher counterpart, and `Interrupt` — the 12th officially documented event — is recognized and refused with a distinct no-counterpart message until a dispatcher subcommand exists. On codex-cli 0.153.4, `SubagentStart` and `SubagentStop` were **measured to fire** (reversing an earlier 0.147.0 observation that SubagentStop never fired) and are now adapted: `RenderHooks` installs `moai hook subagent-start --harness codex` and `moai hook subagent-stop --harness codex` into the user's `.codex/hooks.json`. The compact and permission events are held back on honest evidence, not assumption: non-interactive `codex exec` runs never reached compaction (best effort 264,808 input tokens against a measured 1,048,576-character input cap) and never raised an approval request — recorded as trigger-not-achieved, never as "does not fire".

Unadapted events are not silently ignored — they are **refused**. An unknown event (a typo) and a recognized-but-unadapted event (a scoping decision or a missing counterpart, as with `Interrupt`) return distinct errors, so an operator can tell a mistake from a decision. The config validator collects **every** unknown-key violation instead of stopping at the first.

### What invokes it now

The adapter is no longer call-free: `RenderHooks` writes the two adapted subagent hook lines (`moai hook subagent-start --harness codex`, `moai hook subagent-stop --harness codex`) into the user's `.codex/hooks.json`. The remaining events still await the `--agent` config generator — the follow-up card that wires the adapter into a generated codex config — so for those this page remains a map of where the wiring will land.

## Next steps

- [Multi-model Audit Convergence](/en/advanced/multi-model-audit/) — the path where the codex backend already participates in audits today
- [moai update](/en/cli-reference/update/) — the skill mirror's symlink/copy deployment and its notice
- [Agent Guide](/en/advanced/agent-guide/) — the roles of the eleven agents being dual-published
