---
name: moai-ref-aside-browser
description: >
  Safe-use reference for the optional Aside browser agent (the `aside` command-line
  tool and its MCP server), which works inside the operator's own logged-in browser.
  Covers when the orchestrator may use it, how its permission modes work, and the
  rules that keep it read-only by default. Operated by the orchestrator alone.
  NOT for: public pages that a plain fetch can read, subagent work, or CI runs.

when_to_use: >
  Use before any step that runs `aside` or the Aside MCP tools, when registering the
  optional Aside MCP server, or when `/moai e2e --tool aside` is requested explicitly.

user-invocable: false
metadata:
  version: "1.0.0"
  category: "domain"
  status: "active"
  tags: "aside, browser, mcp, e2e, reference"
---

# Aside Browser (optional)

## Purpose

Aside is an optional browser agent that acts inside the operator's own browser, so it sees every page the operator is signed in to, with the operator's credentials, cookies, and history. `aside mcp` starts an MCP server over stdio with two tools: `exec` runs the browser agent on logged-in sites, and `repl` runs Playwright-style JavaScript against open pages (120-second timeout; `page` starts unset, so call `openTab(url)` or `getTabs()` first). Flag and tool names here reflect one observed version; if a probe or a call fails, treat Aside as absent.

## When to use, and when not

- Use it for a task that needs a page behind the operator's login, which a fresh headless browser cannot reach, and only when the operator asked for it.
- Do not use it for public pages (a plain fetch or the platform default browser tooling reads those), in CI, or because it happens to be installed.

## Who operates it

Aside is operated by the orchestrator alone. A subagent never invokes Aside, neither the `aside` command nor the Aside MCP tools; a task that needs Aside returns a blocker report to the orchestrator, which runs the step itself or drops it.

## Safe-use rules

1. Permission. Never pass `--permission full-access`, and never request full-access by any other route (settings, prompts, or tool arguments).
2. Guard by omission. The permission flag takes one of three values; `ask` and `guard` are the same Guard mode, and omitting the flag also means Guard. Prefer to omit `--permission` entirely so the default Guard applies.
3. REPL limit. `aside repl` has no permission flag, so its read-only limit is discipline, not enforcement: restrict REPL work to navigation, reading, and screenshot operations.
4. Writes. Any step that changes state (submitting a form, posting, purchasing, deleting, sending, changing a setting, approving) needs explicit operator confirmation, obtained by the orchestrator through its question channel before the step runs. Never infer consent from the task text. Treat every `exec` run as state-changing, because an agent-driven run can click, type, and submit.
5. Install. Never install Aside and never run install tooling for it. If the operator wants Aside's own browser skill added to their coding agents, advise the operator to run `aside skills install`; the orchestrator does not run it.
6. Secrets. The session carries the operator's logins, so never place credentials, cookies, tokens, or session data in any output, log, file, report, or prompt. Cite screenshots by path and carry only the minimum page text a step needs.
7. Output size. Redirect verbose REPL output to a file and bring only the exit code and a bounded tail into context.

## Optional registration

Aside is never part of the default configuration. An operator who wants the MCP server registers it with one explicit command:

```text
moai mcp add aside --command aside --args mcp --scope user
```

Omit `--scope user` to register it for one project instead. The registration is optional and never a prerequisite of any workflow in this skill.

## Relation to `/moai e2e`

`/moai e2e --tool aside` is the only way Aside enters the e2e workflow: it is explicit-only, never auto-detected, never offered, and never recommended. The orchestrator loads this skill first and runs every Aside step itself. When Aside is absent, fails its availability probe (`aside --version`), is signed out, or is excluded (for example when `CI=true`), the workflow continues silently on the platform default toolchain, with no Aside-specific message, prompt, install attempt, or failure.
