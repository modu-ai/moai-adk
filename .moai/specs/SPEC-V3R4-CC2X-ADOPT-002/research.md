# CC Upstream Change Analysis — 2.1.237 → 2.1.239 (Umbrella Research)

> **Umbrella SPEC**: SPEC-V3R4-CC2X-ADOPT-002 · created by `/harness:release-update`
> 2026-08-23 (Phase 5 gate resolved as **Option C** — plan + docs sync + child stubs).
> Child stubs: `SPEC-V3R4-CC2X-ADOPT-002-MSGR-001` (Windows messaging doctrine
> refresh), `SPEC-V3R4-CC2X-ADOPT-002-MCP-001` (CC 2.1.238/239 MCP behavior notes).
> Canonical dev-only copy of this research also lives at
> `.moai/research/cc-update-2.1.237-to-2.1.239.md` (untracked).

- **Run date**: 2026-08-23 (hns-release-update-specialist, Phases 0–5 + 7.5)
- **Since version**: 2.1.237 (from `.moai/state/last-cc-version.json`, verified this run)
- **Latest analyzed**: 2.1.239 (CHANGELOG head at fetch time)
- **Source**: `https://raw.githubusercontent.com/anthropics/claude-code/main/CHANGELOG.md` (verbatim via webReader, 2026-08-23)
- **Counts**: total_items = 98 (2.1.238: 39, 2.1.239: 59) · tier1_count = 18 · tier2_count = 21 · tier3_count = 59
- **Plan classification**: tier1+tier2 = 39 ≥ 10 → umbrella SPEC (this SPEC). Impact-bearing subset ≈ 8 items.

## Executive Summary

Two new CC releases since the last sweep (2.1.236..2.1.237, 2026-08-20): **2.1.238** and **2.1.239**. The bulk of both is Remote Control / cloud-session / TUI polish (59 Tier-3 items). The genuinely MoAI-relevant changes cluster into four groups:

1. **Cross-session messaging availability + fidelity (HIGH — doctrine drift).** 2.1.239 ships Windows availability ("Windows: cross-session messaging is now available … as on macOS and Linux"). `.claude/rules/moai/workflow/cross-session-messaging.md` § Availability constraints still states "macOS and Linux only … not on native Windows", and `kanban-dispatch.md` repeats "absent on native Windows". Both are now stale. Supporting fidelity fixes (2.1.238): refused inbound now reports `refused` to the sender (was silent success), inbox drops (rate limit / full queue) now notify the sender, `ListAgents`/`SendMessage` reach Remote Control hosts, and the pre-warmed idle worker no longer appears in listings. 2.1.239 adds: `ListAgents` tells a session its own name, `SendMessage` to self says so, `ListAgents` lists live teammates, and sessions titled `/` are addressable again.
2. **Harness-loadable `.md` robustness (VERIFIED CLEAN).** 2.1.239 fixes agents/skills/commands whose `.md` starts with a UTF-8 BOM being silently ignored. Measured this run: `internal/template/templates/**/*.md` = 343 files, **0 carry a BOM** (`head -c3 | od` scan) — no exposure; no action beyond noting the failure mode for future template authoring.
3. **MCP stdio protocol ordering + trust (MEDIUM).** 2.1.238 fixes stdio servers receiving `server/discover` before `initialize` (forced lazy backends to start every session) — directly relevant to `moai mcp-server` (`internal/cli/mcp_server.go`), a stdio server. Same release tightens `headersHelper` trust requirements (project `.mcp.json` servers now require the folder trust dialog, also under `claude -p`) — MoAI's template ships `.mcp.json` wiring `moai mcp-server`, so headless `-p` flows on fresh clones may now hit a trust prompt where they previously did not.
4. **Session/resume + worktree hygiene (LOW–MEDIUM).** 2.1.239 fixes `claude -c`/resume cross-directory collisions (paths differing only by `_`, `-`, `.`), `/resume` into deleted directories (removed worktrees), `.worktreeinclude` `**/` patterns, and the Linux sandbox breaking every sandboxed git command in repos with `extensions.worktreeConfig` set — the last two bear on MoAI's worktree doctrine (though MoAI does not set `extensions.worktreeConfig` itself; CC 2.1.207 notes it can be left behind by CC's own `worktree.sparsePaths`).

No removals or breaking changes touch surfaces MoAI depends on (no hook-schema, agent-frontmatter, permissions-model, or settings-precedence changes). No version-floor references in the template require bumping.

## Tier 1 — Critical (hooks / agents / skills / plugins / MCP / permissions / settings schema) — 18 items

| Version | Category | Summary | Impact on moai-adk-go | Doc |
|---|---|---|---|---|
| 2.1.239 | agents/skills/commands | `.md` files starting with a UTF-8 BOM were silently ignored; now load | Templates verified clean (343 files, 0 BOM — measured this run). Note for future authoring. | code.claude.com/docs/en/skills (SKILL.md format unchanged; stable_signature: name ≤64 chars, description ≤1024) |
| 2.1.239 | plugins | claude.ai-synced plugins show as `name@synced`; `claude plugin enable/disable @synced`; never override a same-named local plugin | None (MoAI ships no claude.ai-synced plugins) | code.claude.com/docs/en/plugins |
| 2.1.239 | skills (bundled) | `/claude-api upgrade` migration for anthropic 0.x→1.x + skill reference update | None (bundled Anthropic skill, not MoAI surface) | — |
| 2.1.239 | MCP (UI) | Elicitation forms taller than the terminal no longer clipped in fullscreen | Low (moai mcp-server exposes no elicitation) | code.claude.com/docs/en/mcp |
| 2.1.239 | MCP | Remote MCP servers no longer stay failed after transient 5xx on mid-session reconnect (`setMcpServers()` / cloud) | Low–mid (benefits moai MCP consumers on reconnect) | code.claude.com/docs/en/mcp |
| 2.1.239 | plugins | Marketplace `metadata.pluginRoot` now effective for bare source names | None | code.claude.com/docs/en/plugins |
| 2.1.239 | hooks/telemetry | OTel trace fragmentation fixed for tool executions deferred by a `PreToolUse` hook | Low (cleaner traces for MoAI's PreToolUse hooks) | docs.anthropic.com/en/docs/claude-code/hooks (stable_signature: Configuration Safety — hooks snapshot at startup) |
| 2.1.239 | hooks | Hooks failing `posix_spawn ENOENT` after cwd deleted; now run from project root/home | Mid (robustness for MoAI's hook wrappers in worktree-removal flows) | hooks doc, Hook Execution Details |
| 2.1.239 | settings/rules | `claudeMdExcludes` now excludes a symlinked `.claude/rules` file when the pattern names the rules dir or the symlink | Low (MoAI does not ship `claudeMdExcludes`; symlinked-rules users benefit) | settings doc |
| 2.1.238 | settings schema | New `keybindingFlavor` setting (`classic` default, `readline` = Bash-style Ctrl+W) | None (additive key; not MoAI-managed) | settings doc (key not yet in the settings table at fetch time) |
| 2.1.238 | plugins/MCP | Marketplace `headersHelper` mints HTTP headers for catalog/archive fetches | None | plugins doc |
| 2.1.238 | plugins/permissions | Catalog-entry `headersHelper` gated to install/update with `[y/N]` prompt | None | plugins doc |
| 2.1.238 | MCP/permissions | `headersHelper` + inline MCP servers in project `.mcp.json` / agent files now require folder trust acceptance (also under `claude -p`) | Mid for headless: template `.mcp.json` consumers on fresh clones may see a trust prompt | code.claude.com/docs/en/mcp (stable_signature: "Claude Code prompts for approval before using project-scoped servers from `.mcp.json`") |
| 2.1.238 | MCP/security | `headersHelper` from project/plugin/agent scope runs without inherited credential env vars; user/managed/claude.ai helpers run from config dir | None directly (defense in depth) | mcp doc |
| 2.1.238 | MCP protocol | stdio servers no longer receive `server/discover` before `initialize` (lazy backends stopped starting every session open) | Mid — `moai mcp-server` is a stdio server; ordering fix may alter its startup request sequence expectations | mcp doc |
| 2.1.238 | MCP CLI | `claude mcp list`/`get` show disabled servers as `⊘ Disabled` without health-checking them | Low (CLI UX only) | mcp doc |
| 2.1.238 | skills (bundled) | claude-api skill updated for Managed Agents Aug 19 release | None | — |
| 2.1.238 | permissions | Improved Bash permission checking for zsh-specific syntax in shell conditionals | Low–mid (MoAI Bash guards rely on CC's analyzer; fewer false results) | settings doc, permissions section |

## Tier 2 — Important (TUI / CLI / statusline / worktree / headless / session / memory) — 21 items

| Version | Category | Summary | Impact on moai-adk-go | Doc |
|---|---|---|---|---|
| 2.1.239 | messaging/session | **Windows: cross-session messaging now available** (`SendMessage`/`ListAgents` across machines, as on macOS/Linux) | **HIGH — doctrine drift**: `cross-session-messaging.md` § Availability constraints ("not on native Windows") and `kanban-dispatch.md` ("absent on native Windows") are stale; both need updates (template + local mirrors) | — |
| 2.1.239 | messaging | `ListAgents` lists live teammates (previously only subagents/sessions — a reachable teammate looked absent) | Mid — MoAI `cross-session-messaging.md` documents teammate listing behavior; `ListAgents` row needs a refresh | — |
| 2.1.239 | messaging | `ListAgents` tells a session its own name; `SendMessage` to your own name says so | Low — kanban lead self-identification gets easier | — |
| 2.1.239 | messaging | Sessions titled `/` addressable via `SendMessage` again (were "(untitled)") | Low | — |
| 2.1.239 | session | `claude -c`/resume no longer picks sessions from a different directory whose path differed only by `_`, `-`, `.` | Mid — MoAI worktree names (`.claude/worktrees/<card-id>`) and session anchoring benefit | — |
| 2.1.239 | session | Custom session titles no longer disappear from `/resume` after ~64 KB post-rename | Low | — |
| 2.1.239 | session | `/resume` no longer marks sessions "recently changed" on mere reopen/touch | Low | — |
| 2.1.239 | session/worktree | `/resume` all-projects no longer tells you to `cd` into a deleted directory (removed worktree); resumes in cwd | Mid — aligns with MoAI worktree-disposal flows | — |
| 2.1.239 | worktree | `.worktreeinclude` patterns starting with `**/` now match targets inside gitignored directories | Low (MoAI does not ship `.worktreeinclude` defaults) | — |
| 2.1.239 | worktree/sandbox | Linux sandbox no longer breaks all sandboxed git in repos with `extensions.worktreeConfig` set (nonexistent `.git/config.worktree` unreadable) | Low–mid — CC itself can leave `extensions.worktreeConfig` behind (2.1.207 note); MoAI worktree-heavy repos on Linux sandboxed sessions benefit | — |
| 2.1.239 | session/hooks | Remote sessions send keep-alives while a long `SessionStart`/`Setup` hook runs (no idle-reap mid-hook) | Low (remote-only; MoAI SessionStart hooks are non-trivial) | hooks doc |
| 2.1.239 | session loop | `/goal`: repeat check-ins back off 30 min → 1 h → 2 h | Low — native `/goal`; MoAI's `/moai goal` is its own engine (goal-directive.md unaffected) | — |
| 2.1.239 | session loop | `/goal`: resume from the `claude --resume` picker restores the active goal | Low (same scope) | — |
| 2.1.239 | session/TUI | Esc-with-queued-prompt race fixed (next turn finishing early; later resubmit could repeat actions) | Low–mid (turn-integrity for long MoAI sessions) | — |
| 2.1.239 | CLI tool | WebFetch no longer retains expired page content for the whole session (intended 15-min TTL now enforced) | Low — MoAI's URL-verification protocol benefits (`CLAUDE_CODE_WEBFETCH_CACHE_TTL_MS`, 2.1.233, now behaves as documented) | — |
| 2.1.239 | TUI/statusline | Cost estimates (`/cost`, statusline, `--max-budget-usd`) include the 1.1× US-only-inference premium | Low (statusline cost accuracy only) | — |
| 2.1.238 | memory/subagents | Unbounded memory growth in long interactive sessions fixed (subagent tool results released after leaving the display window) | Mid — long MoAI orchestration sessions directly benefit | — |
| 2.1.238 | output styles | Custom/project/plugin output styles no longer drift back to default mid-session | Mid — MoAI ships persona output styles (`.claude/output-styles/moai/`); drift fix improves persona stability | — |
| 2.1.238 | worktree | Worktree-isolation Bash refusals no longer tell you to remove a redirect the command had none | Low — error-message quality on MoAI worktree guard hits | — |
| 2.1.238 | messaging | Sending to a machine-local session that refuses inbound (`crossSessionInbound: "refuse"`) now reports `refused` instead of silent success | Mid — MoAI doctrine documents send-result reading (`kanban-dispatch.md`); "refused" outcome should be reflected | — |
| 2.1.238 | messaging | A session whose inbox drops messages (rate limit/full queue) now tells the sender | Mid — extends the 2.1.236 up-front burst-refusal behavior already in `cross-session-messaging.md` § Configuration surface | — |

## Tier 3 — Minor — 59 items (summary)

Not individually tabled; grouped:

- **Remote Control (16)**: per-task Stop, invalid-role exits, session env inheritance, crashed-process reuse, mid-turn message loss, phone model picks, login-expired retry, sign-out message, RC-host `ListAgents`/`SendMessage` reachability, 403 tolerance, clearer disabled message, runaway title-sync fix, images with file path, web proxy for anthropic hosts, Chrome `/clear` tab groups.
- **Cloud/platform (8)**: fullscreen renderer offer on Bedrock/Vertex/Foundry, cloud plan-mode resume, Bedrock proxy streaming double-billing, Bedrock SSO proxy hang, musl native add-ons, Windows VSCode banner, macOS startup, self-hosted-runner flags ×2 + slow-poll removal.
- **TUI/cosmetic (19)**: dark-ansi theme, fullscreen renderer prompt, mouse-in-browser-terminals text, theme badge colors, vim agent view, `selection:copy`, shell-mode Tab `./`, fullscreen click focus, slash-command panels, `/workflows` dialog overflow, Ctrl+W/U placeholder, masked inputs, Ctrl+Backspace, long-path truncation, Ctrl+L/Cmd+K repaint (`/clear` double-press removed in fullscreen), Backspace Ctrl+H, permission-diff wrapping, Ctrl+Z bracketed paste, keybindingFlavor readline word keys.
- **Misc (16)**: usage-limit message, crash dump on deleted cwd, JetBrains 5 s pause, `/insights` tags, org-policy re-send, compaction skill-args reminder, retry-watchdog spend-limit, `/tmp/claude-*-cwd` cleanup, proxy refusal naming, `/model`/`/effort` cache-miss warning, prompt-suggestion flag, startup update-check delay, claude-api skill updates ×2.

## Cross-Cutting Concerns

1. **Doctrine drift is the real deliverable.** The only MoAI-owned text now factually wrong is the cross-session-messaging availability claim (Windows). Affected surfaces: `.claude/rules/moai/workflow/cross-session-messaging.md` (§ Availability constraints, § Configuration surface — refused/inbox-drop outcomes), `.claude/rules/moai/workflow/kanban-dispatch.md` (§ Scope — "absent on native Windows"), both template + local mirrors. Template edits require their own SPEC → child `…-MSGR-001`. docs-site pages (`multi-llm/kanban-mode.md` ×4 locales) were corrected in this sweep's PR (docs-site is not template-managed).
2. **MCP stdio ordering.** `moai mcp-server` should be smoke-tested against 2.1.239 (lazy-start no longer forced). No code change indicated — the fix relaxes a client-side behavior — but a confirmation note in `.claude/rules/moai/core/moai-mcp-tools.md` may be warranted if behavior differs → child `…-MCP-001`.
3. **Trust gating under `claude -p`.** Headless MoAI flows on fresh clones now require the project-folder trust dialog to have been accepted for `.mcp.json` servers. Docs-site (if it documents headless setup) should mention it → child `…-MCP-001`.
4. **docs-site sync.** No CC-version floors changed beyond the messaging-availability topic; corrected in this sweep (OS bullet, 4 locales, Windows floor v2.1.239 noted).

## Recommended Child SPEC Decomposition

- Child 1 — `SPEC-V3R4-CC2X-ADOPT-002-MSGR-001`: doctrine refresh — cross-session messaging availability + send-result outcomes (`cross-session-messaging.md`, `kanban-dispatch.md`; template + local mirrors).
- Child 2 — `SPEC-V3R4-CC2X-ADOPT-002-MCP-001`: MCP notes — stdio discover-ordering note in `moai-mcp-tools.md` + headless trust-prompt caveat (docs-site 4-locale if user-facing).
- (No child needed for BOM — verified clean this run.)

## References

- CHANGELOG (verbatim): https://raw.githubusercontent.com/anthropics/claude-code/main/CHANGELOG.md (fetched 2026-08-23)
- hooks: https://docs.anthropic.com/en/docs/claude-code/hooks (canonical: code.claude.com/docs/en/hooks)
- sub-agents: https://docs.anthropic.com/en/docs/claude-code/sub-agents (canonical: code.claude.com/docs/en/sub-agents)
- skills: https://docs.anthropic.com/en/docs/claude-code/skills (canonical: code.claude.com/docs/en/skills)
- plugins: https://docs.anthropic.com/en/docs/claude-code/plugins (canonical: code.claude.com/docs/en/plugins)
- mcp: https://docs.anthropic.com/en/docs/claude-code/mcp (canonical: code.claude.com/docs/en/mcp)
- settings: https://docs.anthropic.com/en/docs/claude-code/settings (canonical: code.claude.com/docs/en/settings)
- Prior sweep: `.moai/research/cc-update-2.1.236-to-2.1.237.md`

## Open Questions

1. **Umbrella vs plan-only**: resolved 2026-08-23 — operator selected Option C (umbrella + child stubs + docs sync).
2. **Windows messaging floors**: does Windows availability also carry the v2.1.225/232/236 version floors (cross-machine open, @mentions, `notify_when_idle` 2.1.236) on Windows, or a subset? Unverifiable from the changelog text alone ("as on macOS and Linux" suggests parity); child `…-MSGR-001` should verify before finalizing doctrine wording.
3. **`moai mcp-server` lazy-start**: worth a live smoke test under CC 2.1.239 before closing child `…-MCP-001` (verification-claim discipline — this research asserts no observed behavior).
