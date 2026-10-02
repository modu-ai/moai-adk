---
id: SPEC-CC-ULTRACODE-TOGGLE-001
title: "ultracode wording correction — /effort ultracode is an independent on/off toggle, not an xhigh-coupled effort level (rule source + template mirror + docs-site 4 locales)"
version: "0.1.0"
status: draft
created: 2026-10-02
updated: 2026-10-02
author: manager-spec
priority: P2
phase: "v3.2.0 target"
module: "docs-site/content"
lifecycle: spec-anchored
tags: "claude-code-upstream, wording-correction, ultracode, effort, docs-site, i18n, template-parity"
tier: M
---

# SPEC-CC-ULTRACODE-TOGGLE-001 — ultracode wording correction

## HISTORY

- 2026-10-02 v0.1.0 — Plan-phase artifacts authored by manager-spec (card t1416, Class C wording correction, worktree `.moai/worktrees/t1416`, branch `WT-ultracode-toggle-wording`, base `c50da9c2f`). Evidence basis: release-update sweep item U3 (2026-10-02).

## Context (WHY)

Claude Code v2.1.284 changed `ultracode` into its own toggle in `/effort`. MoAI's rule source and docs-site still describe `/effort ultracode` as a mode that "combines `xhigh` reasoning with automatic workflow orchestration" and tell the reader to "step back with `/effort high`". Both statements are false from v2.1.284 onward: ultracode no longer forces `xhigh`, it stays on at any effort level, and it is turned off with `/effort ultracode off`, not by picking another level.

### Verified upstream facts (primary sources fetched 2026-10-02; do not re-litigate)

| ID | Fact | Source (fetched successfully) |
|----|------|-------------------------------|
| F1 | v2.1.284 changelog: "Changed Ultracode into its own toggle in `/effort` (Tab, or `/effort ultracode [on\|off]`): it no longer forces xhigh effort and stays on at any effort level" | `https://raw.githubusercontent.com/anthropics/claude-code/main/CHANGELOG.md` § 2.1.284 |
| F2 | "Ultracode is a Claude Code setting rather than a model effort level … at whichever effort level the session runs at." "Turning ultracode on or off with `/effort` or the `ultracode` setting leaves the effort level unchanged." | `https://code.claude.com/docs/en/model-config` § Adjust effort level |
| F3 | Exception: "The `--effort ultracode` flag and the Agent SDK `effortLevel: "ultracode"` value turn it on and also set the level to `xhigh`." Picking a level in the `/effort` slider or `/model` picker leaves ultracode as it was. | same page |
| F4 | `/effort ultracode off` turns it off; the `off` form, the slider toggle, and staying on at non-`xhigh` levels need v2.1.284+. Before v2.1.284: turning it on set `xhigh`, picking another level turned it off. | same page |
| F5 | Persistence: `/effort ultracode` "lasts for the current session; to have every session start with it, set the `ultracode` setting" (`"ultracode": true` in any settings file). | `https://code.claude.com/docs/en/workflows` § Let Claude decide with ultracode; `https://code.claude.com/docs/en/settings-reference` § `ultracode` (Boolean, default unset, any scope) |
| F6 | The persisted `effortLevel` setting and `CLAUDE_CODE_EFFORT_LEVEL` do not accept `ultracode`; if either sets the level, ultracode stays on at that level. | model-config |

### Open questions (evidence gaps — not asserted anywhere in this SPEC's required wording)

- OQ-1 — Does flipping the slider's Ultracode toggle (Tab, then Enter) persist across sessions like Enter on an effort level does? The fetched pages state only that the `/effort` route is for the current session and that the `ultracode` setting is the persistent route; none says the slider toggle writes the setting. Closing it needs a live observation (toggle in a fresh session, inspect settings files, start a new session). Until then the required wording says exactly what the docs say and makes no slider-specific claim (REQ-011).
- OQ-2 — The `ultracode` settings entry in the fetched settings reference names no minimum version. The required wording names no version for the key (REQ-011).
- OQ-3 — The local changelog snapshot `.moai/research/cc-changelog-snapshot-2.1.233.md` predates 2.1.282/2.1.284 (its only ultracode hits are the older keyword rename and fixes), so the 2.1.284 fact rests on the web changelog read through a fetch tool. F1 is cross-corroborated by F4 ("Before v2.1.284 ...") from a second page.

Evidence baseline pinned to tree `c50da9c2f` (measured 2026-10-02 in this worktree): see `acceptance.md` RED-now cells.

## 1. Requirements (GEARS)

- REQ-001 — The rule-source bullet in `.claude/rules/moai/workflow/dynamic-workflows.md` that distinguishes the `ultracode` per-prompt trigger from the session-wide mode shall describe `/effort ultracode` as an independent on/off toggle introduced in Claude Code v2.1.284 that leaves the session's effort level unchanged, and shall not state that it combines or forces `xhigh`.
- REQ-002 — **When** that bullet names the ways to turn the session-wide mode on or off, it shall give `/effort ultracode off` as the off route in place of the "step back with `/effort high`" instruction, and shall state the launch-time exception: `--effort ultracode` (and the Agent SDK `effortLevel: "ultracode"` value) also set the level to `xhigh`.
- REQ-003 — **While** describing persistence, that bullet shall state that a toggle made through `/effort` lasts for the current session and must be re-issued in a new session, and that the `ultracode` settings key (`"ultracode": true`) is the documented route to start every session with it on. The existing statement that the `ultrathink.` opener of a resume message does not restore ultracode shall be retained and shall remain consistent with the settings-key route.
- REQ-004 — The template mirror `internal/template/templates/.claude/rules/moai/workflow/dynamic-workflows.md` shall carry the same edit such that the local copy and the mirror remain byte-identical.
- REQ-005 — In each of the four locales (ko, en, ja, zh) the `/effort ultracode` row of the table in `docs-site/content/<loc>/claude-code/agentic/workflows.md` shall state the same facts as REQ-001 to REQ-003 (toggle, effort level unchanged, `/effort ultracode off`, current-session scope, `ultracode` setting for persistence, `--effort ultracode` exception, v2.1.284), shall remain a single table row with the same column count, and shall not contain the `xhigh` coupling or the `/effort high` return phrasing.
- REQ-006 — In each of the four locales the code-block comment on the `/effort ultracode` line in `docs-site/content/<loc>/multi-llm/_index.md` shall no longer state `xhigh`, and the line shall remain a one-line comment inside the same code block.
- REQ-007 — In each of the four locales `docs-site/content/<loc>/advanced/ultracode-workflows.md` shall no longer state that `/effort ultracode` raises or sets reasoning effort to `xhigh`, shall no longer instruct the reader to return with `/effort high`, and (ko only) shall no longer introduce the effects as "three things change together"; the existing session-boundary callout and every heading shall be preserved.
- REQ-008 — In each of the four locales `docs-site/content/<loc>/claude-code/foundations/commands.md` shall no longer describe `ultracode` as an `/effort` level (the sentence that calls it "simultaneously an `/effort` level" and, in ko and en, the sentence that lists it among the levels), and shall describe it as a toggle offered in `/effort`.
- REQ-009 — **When** the docs-site edits are made, ko shall be authored first as the canonical locale and en, ja, and zh derived from it in the same change set, with `v2.1.284` and the literal commands `/effort ultracode off`, `--effort ultracode`, and `"ultracode": true` kept verbatim and untranslated (per `hns-oss-docs-i18n-rules` §1-§2).
- REQ-010 — The change shall not introduce new headings, new URLs outside the docs-site's allowed domain rule, emoji in body text, or Mermaid `LR`/`RL` directions in any edited docs-site page, and shall leave `docs-site/.locale-parity-baseline` untouched (the page `advanced/ultracode-workflows.md` is already listed there as divergent; the edit must not add divergence).
- REQ-011 — The edited text shall not assert that the `/effort` slider toggle persists across sessions, and shall not name a minimum version for the `ultracode` settings key (OQ-1, OQ-2).
- REQ-012 — Lines and files outside the change map (§3) shall remain untouched, including every surface classified out of scope in §4.

## 2. Non-Goals

- No change to Go code, hooks, or the handoff injector: `internal/hook/handoff_inject_render.go` renders `/effort ultracode ← restore workflow fan-out` guidance, which stays accurate.
- No new SPEC for the open questions; they are carried as OQ-1..OQ-3.

## 3. Change Map (WHAT is edited, by classification)

| Surface | Files | Class |
|---------|-------|-------|
| Rule source + template mirror | `.claude/rules/moai/workflow/dynamic-workflows.md` (bullet at L111) + `internal/template/templates/.claude/rules/moai/workflow/dynamic-workflows.md` | IN SCOPE — card-named |
| docs-site workflows table row | `docs-site/content/{ko,en,ja,zh}/claude-code/agentic/workflows.md` (ko L113, en/ja/zh L105) | IN SCOPE — card-named |
| docs-site multi-llm code comment | `docs-site/content/{ko,en,ja,zh}/multi-llm/_index.md` (ko L96, en L101, ja L98, zh L88) | IN SCOPE — card-named |
| docs-site ultracode-workflows page | `docs-site/content/{ko,en,ja,zh}/advanced/ultracode-workflows.md` (ko L65 + L67 + L71 prose, en L112, ja/zh L109) | IN SCOPE — same defect found by sweep (`xhigh` coupling + `/effort high` return) |
| docs-site commands page | `docs-site/content/{ko,en,ja,zh}/claude-code/foundations/commands.md` (ko L79+L137, en L79+L137, ja L80, zh L80) | IN SCOPE — secondary: same upstream change, "is an effort level" claim contradicted by F2; lane may split to a follow-up card if it wants strictly card-named scope |

## 4. Out of Scope

### Out of Scope — Surfaces that remain accurate after the correction

- `.claude/rules/moai/workflow/session-handoff.md`, `session-handoff-examples.md`, `session-handoff-format.md`, `.claude/output-styles/moai/moai.md` and their template mirrors: they describe the paste-time bare `ultracode` keyword versus the `/effort ultracode` session-persistence variant and state that `ultrathink.` does not restore ultracode. All of that stays true (F5). `session-handoff-examples.md` L93 phrase "drops to non-ultracode effort" is loose wording, not a false claim; flagged as a follow-up candidate, not edited here.
- `docs-site/content/*/multi-llm/model-policy.md` (lists `/effort low|medium|high|xhigh|max|ultracode|auto` as accepted slash-command values): the command syntax is still accepted; no `xhigh` coupling is stated.
- `docs-site/content/*/cli-reference/handoff.md` and `internal/cli/handoff.go` (`--ultracode` records a restoration directive): unaffected.
- `docs-site/content/*/claude-code/agentic/sub-agents.md` ("`ultracode` sessions exempt" from the concurrent-subagent cap): unaffected.
- `internal/template/templates/AGENTS.md.tmpl` ("Workflow scripts (`ultracode`)" capability name): unaffected.

### Out of Scope — Historical and dated records

- `CHANGELOG.md` entries and `.moai/docs/harness-delivery-strategy.md` L53 / `.moai/docs/autonomous-workflow-strategy.md` (dated 2026-06 strategy records): historical statements about what was true at the time; not rewritten.

### Out of Scope — Verification of unresolved facts

- OQ-1 (slider-toggle persistence) live observation: not performed here; the SPEC wording avoids the claim instead.
