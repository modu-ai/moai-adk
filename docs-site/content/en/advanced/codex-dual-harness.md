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

The two harnesses' hook surfaces are nearly but not exactly the same. Measurement (against codex-cli 0.147.0) found exactly two divergences: the **event name** the harness passes, and **three output keys** codex declares but does not act on (`systemMessage`, `continue`, `stopReason`). Everything else measured identical, so `internal/codexadapter` is a thin translation layer that sits **in front of** the dispatcher — nothing under `internal/hook` is modified.

### The 11-event table

| Codex event | MoAI dispatcher arg | Adapted this milestone? |
|---|---|---|
| PreToolUse | `pre-tool` | yes |
| PostToolUse | `post-tool` | yes |
| SessionStart | `session-start` | yes |
| SessionEnd | `session-end` | yes |
| Stop | `stop` | yes |
| UserPromptSubmit | `user-prompt-submit` | yes |
| PreCompact | `compact` | no — unmeasured |
| PostCompact | `post-compact` | no — unmeasured |
| PermissionRequest | `permission-request` | no — unmeasured |
| SubagentStart | `subagent-start` | no — unmeasured |
| SubagentStop | `subagent-stop` | no — measured NOT to fire |

All eleven events have a dispatcher counterpart. Excluding an event from adaptation is a scoping decision about measurement coverage, never an absence of a counterpart. SubagentStop is the special case: it was **measured never to fire** — under codex, delegation surfaces as a PostToolUse whose tool name begins "collaboration", so mapping it would wire a path nothing ever flows through.

Unadapted events are not silently ignored — they are **refused**. An unknown event (a typo) and a recognized-but-unadapted event (a scoping decision) return distinct errors, so an operator can tell a mistake from a decision. The config validator collects **every** unknown-key violation instead of stopping at the first.

### Nothing invokes it yet

{{< icon warning warn >}} This package shipped as a library and **nothing calls it yet**. The `--agent` config generator — the follow-up card that wires the adapter into a generated codex config — completes the connection. As of now this page is a map of where the wiring will land, not the manual for a switched-on feature.

## Next steps

- [Multi-model Audit Convergence](/en/advanced/multi-model-audit/) — the path where the codex backend already participates in audits today
- [moai update](/en/cli-reference/update/) — the skill mirror's symlink/copy deployment and its notice
- [Agent Guide](/en/advanced/agent-guide/) — the roles of the eleven agents being dual-published
