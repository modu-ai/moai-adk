---
id: SPEC-20261122-troubleshooting-doc
title: "A troubleshooting page: safe mode, a clean config folder, debug files, the classifier notice and the limits moai cannot cover"
status: draft
revision: 1
created: 2026-11-22
---

# A troubleshooting page: safe mode, a clean config folder, debug files, the classifier notice and the limits moai cannot cover

## Intent

MoAI ADK installs skills, agents, rules and hooks into a Claude Code configuration folder and runs local git hooks. When a session misbehaves, a user cannot tell whether moai, another customization or Claude Code itself is the cause. One troubleshooting page (`docs/troubleshooting.md`) gives the steps Claude Code's own documents give for that question, in the order a user takes them, and states the places where moai's hooks and settings do not reach, so nobody expects them there.

The page is generic product text: it names no project of this repository and no card. A Rust test reads the page and names each statement that is missing.

Source: the leader's whole-docs report (`claude-code-native-redesign-2026-11-22.md`, section 16 item 26, sections 2-4, 13 and 15 row 28) and coverage note A38; the official pages below were read whole for this SPEC.

## Non-goals

- A `moai doctor` command or any code that runs these steps for the user: the page tells the user what to run.
- The host capability table (section 16 item 27): its own card.
- A Korean copy of the page: the docs folder is English; a translation is a later pick.
- Links from the README or other pages: the README is outside this card's scope; a later slice or the docs owner adds the link.

## Requirements

- REQ-1: The troubleshooting page shall tell the user to start with `claude --safe-mode` and name what safe mode turns off, moai's skills, agents, rules and hooks among them.
- REQ-2: The troubleshooting page shall describe a clean reproduction that sets `CLAUDE_CONFIG_DIR` to a new folder, passes `--setting-sources user` and leaves `HOME` unchanged.
- REQ-3: The troubleshooting page shall describe `--debug-file <path>` and `CLAUDE_CODE_DEBUG_LOG_LEVEL=verbose`, and show how a hook's run appears in the debug log.
- REQ-4: The troubleshooting page shall tell the user to read a debug file and remove keys, tokens and private paths before sharing it.
- REQ-5: The troubleshooting page shall explain the auto mode classifier notice that a gateway session shows and name `CLAUDE_CODE_AUTO_MODE_SERVER=0` as a temporary setting.
- REQ-6: The troubleshooting page shall list the commands that show what a session loaded: `/context`, `/memory`, `/skills`, `/hooks`, `/permissions`, `/status`, `/doctor` and `claude doctor`.
- REQ-7: The troubleshooting page shall state that cloud threads do not run moai's hooks or apply its settings, that moai does not support self-hosted runners, and that moai does not support the Linux desktop app.
- REQ-8 (REQ-1): IF a statement that REQ-1 to REQ-7 require is missing from the page, THEN the page test shall fail and name that statement.

## Acceptance

- AC-1 (REQ-1, REQ-2, REQ-3, REQ-4, REQ-5, REQ-6, REQ-7): `docs/troubleshooting.md` holds one section per requirement, each with the words its statement needs (the flag or variable names verbatim, and `HOME`, `secret`-class words, `self-hosted`, `Linux`), and names no card and no project of this repository · verify: `cargo test --offline --locked --test troubleshooting_doc`
- AC-2 (REQ-8): a page that lacks one statement makes the test fail with that statement's name; the test's checker is fed a copy with each statement removed in turn · verify: `cargo test --offline --locked --test troubleshooting_doc a_missing_statement_is_named`

## Scope

- write: `docs/troubleshooting.md`, `tests/troubleshooting_doc.rs`, `.moai/specs/SPEC-20261122-troubleshooting-doc/`
- read-only: the official pages named in Research
- forbidden: `README.md`, `README.ko.md`, every other document and all code

## Plan

1. This slice: the SPEC, then the red test (`tests/troubleshooting_doc.rs`), then the page, then the check affected once and READY.

## Risks

- Claude Code changes a flag or variable name; the page then gives a wrong step. The page carries the date the pages were read, and the test pins the names so a later edit is deliberate.
- `CLAUDE_CODE_AUTO_MODE_SERVER` "may be removed in a later release" (R-5): the page says so in the same sentence.
- A test that checks words can pass on a page that reads badly; the audit reads the page whole.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | A clean reproduction changes `CLAUDE_CONFIG_DIR`, never `HOME`. | Card body ("never HOME"); the official clean-configuration step and the multi-account way both use `CLAUDE_CONFIG_DIR` (R-2, R-3), and the login and the macOS Keychain entry follow that folder (R-3); no official page describes changing `HOME`, which also moves the user's git, SSH and shell configuration. | a later user decision |
| D-2 | The page tells the user to remove keys, tokens and private paths from a debug file before sharing it. | The debug log holds hook output and command lines (R-6); no official page gives redaction advice, so moai's page does (product rule, card body "가리고 공유"). | a later user decision |
| D-3 | Self-hosted runners and the Linux desktop app are stated as not supported, not worked around. | A self-hosted runner can point `core.hooksPath` at its own hooks and its git proxy replaces the global git config with no backup (R-8), which bypasses moai's local git hooks; the Linux desktop app is a beta with missing features (R-9); report section 13 and section 15 row 28. | a later user decision |
| D-4 | The page is English only, in `docs/`. | The docs folder is English; the README pair is the only bilingual text. | a later pick adds `docs/troubleshooting.ko.md` |

### Research

Pages read whole as raw markdown on 2026-11-22 between 11:05Z and 11:07Z (hooks.md: its "Debug hooks" part to the end of the page and its "Timeouts" part; the self-hosted deploy, configuration and reference pages: their git sections). Quotes are the pages' words.

| Id | Page says | Used for |
| --- | --- | --- |
| R-1 | cli-reference.md, `--safe-mode`: "Start with all customizations disabled to troubleshoot a broken configuration: CLAUDE.md, skills, plugins, hooks, MCP servers, custom commands and agents, output styles, workflows, custom themes, custom keybindings, status line and file-suggestion commands, LSP servers, and auto memory do not load. Authentication, model selection, built-in tools, and permissions work normally"; debug-your-config.md: "If the problem disappears in safe mode, you've narrowed the cause to one of the items you turned off." | REQ-1 |
| R-2 | debug-your-config.md: "set `CLAUDE_CONFIG_DIR` to a fresh directory in place of `~/.claude`, and pass `--setting-sources user` to turn off project configuration such as `.claude/settings.json`, `CLAUDE.md`, and `.mcp.json`"; "On the first launch, you go through the first-run setup screens again, a sign that the clean configuration directory is in effect."; "Copy `~/.claude.json` and the files in `~/.claude` into `~/claude-clean/config` one at a time." | REQ-2 |
| R-3 | authentication.md: "If you've set the `CLAUDE_CONFIG_DIR` environment variable, Claude Code keeps the `.credentials.json` file under that directory instead ... and keys the macOS Keychain entry to that directory too"; "give each account its own configuration directory ... set the `CLAUDE_CONFIG_DIR` environment variable". The page says nothing about changing `HOME`. | REQ-2, D-1 |
| R-4 | cli-reference.md, `--debug-file <path>`: "Write debug logs to a specific file path. Implicitly enables debug mode."; env-vars.md, `CLAUDE_CODE_DEBUG_LOG_LEVEL`: "Values: `verbose`, `debug` (default), `info`, `warn`, `error`. Set to `verbose` to include high-volume diagnostics" | REQ-3 |
| R-5 | auto-mode-classifier-billing.md: the notice "We're changing auto mode to no longer charge for classifier requests in Claude Code. However, this session isn't eligible."; "The most common cause is an LLM gateway or proxy"; "setting `CLAUDE_CODE_AUTO_MODE_SERVER` to `0` before you start the session ... the notice doesn't appear"; "`CLAUDE_CODE_AUTO_MODE_SERVER` is a temporary setting and may be removed in a later release." | REQ-5 |
| R-6 | hooks.md, "Debug hooks": "Hook execution details are written to the debug log file. Start Claude Code with `claude --debug-file <path>` ... The `--debug` flag doesn't print to the terminal."; a run line reads "`Hooks: PostToolUse:Write [.claude/hooks/log-write.sh] finished with status 0 (31ms)`" (v2.1.296 or later) and may end "`timed out after <N>ms`" | REQ-3, D-2 |
| R-7 | debug-your-config.md lists `/context`, `/memory`, `/skills`, `/hooks` ("If a hook you defined doesn't appear, Claude Code didn't load it."), `/permissions`, `/status`, `/doctor` and "`claude doctor` prints read-only installation and settings diagnostics without starting a session" | REQ-6 |
| R-8 | claude-projects.md: "Cloud threads don't pick up anything from the Claude Code setup on your own machine."; "A cloud thread starts with each repository's `CLAUDE.md` instructions, but not the repositories' `.mcp.json` servers or `.claude/settings.json` hooks, permission rules, and `env`."; self-hosted-environments-deploy.md: `--configure-git` writes "`core.hooksPath` pointing at a runner-managed hooks directory", and with the git proxy "the runner deletes and replaces the global git config of the user it runs as, and keeps no backup" | REQ-7, D-3 |
| R-9 | desktop-linux.md: "Linux support for the Claude desktop app is in beta."; "The desktop app doesn't update itself on Linux."; "**Computer Use** ... isn't available on Linux" | REQ-7, D-3 |

## Result

Not started.
