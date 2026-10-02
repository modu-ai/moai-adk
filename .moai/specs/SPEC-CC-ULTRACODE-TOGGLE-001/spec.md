---
id: SPEC-CC-ULTRACODE-TOGGLE-001
title: "ultracode wording correction — /effort ultracode is an independent on/off toggle, not an xhigh-coupled effort level (rule source + template mirror + docs-site 4 locales)"
version: "0.1.2"
status: completed
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
- 2026-10-02 v0.1.2 — plan-audit iteration 2 repair (`.moai/reports/t1416/plan-audit-iter2.md`, FAIL 0.86 on unresolved D2): AC-001's closed verb-list pin replaced by a structural `xhigh` count guard (exactly one `xhigh` in each rule file, inside the launch-flag clause) with a rewording probe recorded; REQ-003 names the `/effort ultracode` command form (N2); W1 bounded to one clause and U1 made label-independent with a page-level `xhigh` line count (N3); vacuity wording corrected (N4).
- 2026-10-02 v0.1.1 — plan-audit iteration 1 repair (`.moai/reports/t1416/plan-audit.md`, FAIL 0.75): D1-D3 pin gaps closed (per-locale positive pins, zero pins, RED-now values with exit codes in an evidence ledger), D4 settings-reference facts corrected (OQ-2 resolved, OQ-1 narrowed, REQ-011 decided), D5 ledger + literal paths + AC classes, D6 cross-references, D7 REQ-007/AC-006 reconciled and slider check made a command, D8 REQ definition form, D9 REQ-002/003 made ubiquitous, D10 out-of-scope notes, D11 hugo control observed.

## Context (WHY)

Claude Code v2.1.284 changed `ultracode` into its own toggle in `/effort`. MoAI's rule source and docs-site still describe `/effort ultracode` as a mode that "combines `xhigh` reasoning with automatic workflow orchestration" and tell the reader to "step back with `/effort high`". Both statements are false from v2.1.284 onward: ultracode no longer forces `xhigh`, it stays on at any effort level, and it is turned off with `/effort ultracode off`, not by picking another level.

### Verified upstream facts (primary sources read 2026-10-02; do not re-litigate)

| ID | Fact | Source (fetched successfully) |
|----|------|-------------------------------|
| F1 | v2.1.284 changelog: "Changed Ultracode into its own toggle in `/effort` (Tab, or `/effort ultracode [on\|off]`): it no longer forces xhigh effort and stays on at any effort level" (independently re-read by plan-audit iteration 1 with raw `curl`, line 406 of the changelog) | `https://raw.githubusercontent.com/anthropics/claude-code/main/CHANGELOG.md` § 2.1.284 |
| F2 | "Ultracode is a Claude Code setting rather than a model effort level … at whichever effort level the session runs at." "Turning ultracode on or off with `/effort` or the `ultracode` setting leaves the effort level unchanged." | `https://code.claude.com/docs/en/model-config` § Adjust effort level |
| F3 | Exception: "The `--effort ultracode` flag and the Agent SDK `effortLevel: "ultracode"` value turn it on and also set the level to `xhigh`." Picking a level in the `/effort` slider or `/model` picker leaves ultracode as it was. | same page |
| F4 | `/effort ultracode off` turns it off; the `off` form, the slider toggle, and staying on at non-`xhigh` levels need v2.1.284+. Before v2.1.284: turning it on set `xhigh`, picking another level turned it off. | same page |
| F5 | `/effort ultracode` "lasts for the current session; to have every session start with it, set the `ultracode` setting" (`"ultracode": true` in a settings file). | `https://code.claude.com/docs/en/workflows` § Let Claude decide with ultracode |
| F6 | The persisted `effortLevel` setting and `CLAUDE_CODE_EFFORT_LEVEL` do not accept `ultracode`; if either or an effort cap sets the level, ultracode stays on at that level. | model-config |
| F7 | `ultracode` setting entry (Boolean, any scope, default unset): "The key doesn't change the session's effort level: ultracode runs at whichever level the session uses. Claude Code reads this key but never writes it: `/effort ultracode` turns ultracode on for the current session only." "`--effort ultracode` … at `xhigh` effort … requires Claude Code v2.1.203 or later." "This and the `/effort ultracode off` form require Claude Code v2.1.284 or later. Before v2.1.284, `ultracode: true` ran the session at `xhigh` effort, and an effort cap below `xhigh` kept ultracode off." | `https://code.claude.com/docs/en/settings-reference` § `ultracode` (raw HTML read via `curl`, tag-stripped, 2026-10-02) |

### Decisions

- DEC-1 (was OQ-2, resolved): the `ultracode` settings key DOES carry a version qualifier (F7): before v2.1.284 the key ran sessions at `xhigh`. The edited text therefore states the settings-key route together with `v2.1.284` on the same line, so it never reintroduces the removed coupling for an older Claude Code. This is a deliberate wording choice, pinned by AC-001/AC-004 (joint `v2.1.284` + `"ultracode": true` pin).

### Open questions (evidence gaps — not asserted anywhere in the required wording)

- OQ-1 — Does flipping the slider's Ultracode toggle (Tab, then Enter) persist across sessions? Evidence so far (F7): Claude Code "reads this key but never writes it" and "`/effort ultracode` turns ultracode on for the current session only", which argues against slider persistence through the `ultracode` key; but no fetched page names the slider toggle's persistence or rules out another store (the slider's Enter on an effort level does save a default under `modelSettings`). Closing it needs a live observation (toggle in a fresh session, `grep -n ultracode` the settings files, start a new session). The required wording makes no slider claim (REQ-011) and the edited pages do not mention the slider at all (AC-008 slider pin).
- OQ-3 — The local changelog snapshot `.moai/research/cc-changelog-snapshot-2.1.233.md` predates 2.1.282/2.1.284, so the 2.1.284 fact rests on the web changelog; F1 is cross-corroborated by F4 ("Before v2.1.284 ...") and F7 from two further pages and was independently re-read by the plan-audit.

Evidence baseline: measured on this worktree at HEAD `0e7b6af5b` (scope files byte-identical to the original pin `c50da9c2f`: `git diff --stat c50da9c2f HEAD -- .claude internal docs-site` printed nothing). Cells are in `acceptance.md` § Evidence ledger.

## 1. Requirements (GEARS)

- **REQ-001**: The rule-source bullet in `.claude/rules/moai/workflow/dynamic-workflows.md` that distinguishes the `ultracode` per-prompt trigger from the session-wide mode shall describe `/effort ultracode` as an `independent on/off toggle` introduced in Claude Code `v2.1.284` that `leaves the effort level unchanged` (the three backticked phrases appear verbatim), and shall not couple ultracode to `xhigh` in any wording (forces, combines, pairs with, lifts to, defaults to, and every other phrasing); the rule file shall contain `xhigh` only inside the launch-flag clause of REQ-002.
- **REQ-002**: The same bullet shall name `/effort ultracode off` as the off route in place of the "step back with `/effort high`" instruction, and shall state the launch-time exception in this exact form: the `--effort ultracode` launch flag (and the Agent SDK `effortLevel: "ultracode"` value) also starts the session at `xhigh` (the words "starts the session at", followed by `xhigh` in backticks, appear verbatim). That clause shall be the only place in the rule file where `xhigh` appears (exactly one occurrence).
- **REQ-003**: The same bullet shall state that a toggle made with the `/effort ultracode` command lasts for the `current session` and must be re-issued in a new session (the slider toggle is not covered by this sentence; its persistence is OQ-1); shall name the `"ultracode": true` settings key as the documented route to start every session with ultracode on, with `v2.1.284` on the same line (DEC-1); and shall retain the statement that the `ultrathink.` opener of a resume message does not restore ultracode.
- **REQ-004**: The template mirror `internal/template/templates/.claude/rules/moai/workflow/dynamic-workflows.md` shall carry the same edit, such that the local copy and the mirror remain byte-identical.
- **REQ-005**: In each of the four locales (ko, en, ja, zh) the `/effort ultracode` row of the table in `docs-site/content/<loc>/claude-code/agentic/workflows.md` shall state the facts of REQ-001 to REQ-003 — toggle, effort level unchanged, `/effort ultracode off`, current-session scope retained, `"ultracode": true` with `v2.1.284` on the same line, and the `--effort ultracode` exception — whose `xhigh` mention is the row's only `xhigh` and sits in one clause with the `--effort ultracode` literal, with no table-cell break, period, semicolon, or comma (`. 。 ; ； , ， 、`) between the literal and `xhigh` — shall remain a single table row with the same column count, and shall not contain the `/effort high` return phrasing.
- **REQ-006**: In each of the four locales the code-block comment on the `/effort ultracode` line in `docs-site/content/<loc>/multi-llm/_index.md` shall not state `xhigh`, and the line shall remain a one-line comment inside the same code block.
- **REQ-007**: In each of the four locales `docs-site/content/<loc>/advanced/ultracode-workflows.md` shall not state, in a list bullet or in prose, that `/effort ultracode` raises or sets reasoning effort to `xhigh` (the page's `xhigh` mention in the effects list is removed and none is added: its count of `xhigh` lines falls by exactly one), shall not instruct the reader to return with `/effort high` (ko only carries that phrase today), shall name `/effort ultracode off` as the off route, and (ko only) shall not introduce the effects as "three things change together"; the existing session-boundary callout and every heading shall be preserved.
- **REQ-008**: In each of the four locales `docs-site/content/<loc>/claude-code/foundations/commands.md` shall describe `ultracode` as a toggle — a line containing `ultracode` and the locale's toggle word (en `toggle`, ko `토글`, ja `トグル`, zh `开关`) — in place of the sentence that calls it "simultaneously an `/effort` level"; and in ko and en the L137 sentence that lists `ultracode` among the `/effort` levels shall no longer do so.
- **REQ-009**: **When** the docs-site edits are made, ko shall be authored first as the canonical locale and en, ja, and zh derived from it in the same change set, with `v2.1.284` and the literal commands `/effort ultracode off`, `--effort ultracode`, and `"ultracode": true` kept verbatim and untranslated (per `hns-oss-docs-i18n-rules` §1-§2).
- **REQ-010**: The change shall not introduce new headings, new URLs outside the docs-site's allowed domain rule, emoji in body text, or Mermaid `LR`/`RL` directions in any edited docs-site page, and shall leave `docs-site/.locale-parity-baseline` untouched (the pages `advanced/ultracode-workflows.md` and `claude-code/foundations/commands.md` are both already listed there as divergent; the edit must not add divergence).
- **REQ-011**: The edited text shall not assert that the `/effort` slider toggle persists across sessions and shall not mention the slider (OQ-1), and shall state the settings-key route only together with `v2.1.284` (DEC-1).
- **REQ-012**: Lines and files outside the change map (§3) shall remain untouched, including every surface classified out of scope in §4.

## 2. Non-Goals

- No change to Go code, hooks, or the handoff injector: `internal/hook/handoff_inject_render.go` renders `/effort ultracode ← restore workflow fan-out` guidance, which stays accurate.
- No new SPEC for the open questions; they are carried as OQ-1 and OQ-3.

## 3. Change Map (WHAT is edited, by classification)

| Surface | Files | Class |
|---------|-------|-------|
| Rule source + template mirror | `.claude/rules/moai/workflow/dynamic-workflows.md` (bullet at L111) + `internal/template/templates/.claude/rules/moai/workflow/dynamic-workflows.md` | IN SCOPE — card-named |
| docs-site workflows table row | `docs-site/content/{ko,en,ja,zh}/claude-code/agentic/workflows.md` (ko L113, en/ja/zh L105) | IN SCOPE — card-named |
| docs-site multi-llm code comment | `docs-site/content/{ko,en,ja,zh}/multi-llm/_index.md` (ko L96, en L101, ja L98, zh L88) | IN SCOPE — card-named |
| docs-site ultracode-workflows page | `docs-site/content/{ko,en,ja,zh}/advanced/ultracode-workflows.md` (ko L65 + L67 + L71 prose, en L112, ja L109, zh L109) | IN SCOPE — same defect found by sweep (`xhigh` coupling + `/effort high` return) |
| docs-site commands page | `docs-site/content/{ko,en,ja,zh}/claude-code/foundations/commands.md` (ko L79 + L137, en L79 + L137, ja L80, zh L80) | IN SCOPE — secondary: same upstream change, "is an effort level" claim contradicted by F2; lane may split to a follow-up card if it wants strictly card-named scope |

## 4. Out of Scope

### Out of Scope — Surfaces that remain accurate after the correction

- `.claude/rules/moai/workflow/session-handoff.md`, `session-handoff-examples.md`, `session-handoff-format.md`, `.claude/output-styles/moai/moai.md` and their template mirrors: they describe the paste-time bare `ultracode` keyword versus the `/effort ultracode` session-persistence variant and state that `ultrathink.` does not restore ultracode. All of that stays true (F5, F7: Claude Code never writes the key, `/effort ultracode` is current-session only). `session-handoff-examples.md` L93 phrase "drops to non-ultracode effort" is loose wording, not a false claim; flagged as a follow-up candidate, not edited here.
- `.moai/docs/session-handoff-appendix.md` (L67) and its mirror `internal/template/templates/.moai/docs/session-handoff-appendix.md` (L67): the sentence says effort keywords (`ultrathink` / `ultracode`) inside a command argument are not documented to fire; it makes no claim about `xhigh` or the toggle model and stays true.
- `docs-site/content/*/multi-llm/model-policy.md` (lists `/effort low|medium|high|xhigh|max|ultracode|auto` as accepted slash-command values): the command syntax is still accepted (`/effort ultracode [on|off]`); no `xhigh` coupling is stated.
- `docs-site/content/*/cli-reference/handoff.md` and `internal/cli/handoff.go` (`--ultracode` records a restoration directive): unaffected.
- `docs-site/content/*/claude-code/agentic/sub-agents.md` ("`ultracode` sessions exempt" from the concurrent-subagent cap): unaffected.
- `internal/template/templates/AGENTS.md.tmpl` ("Workflow scripts (`ultracode`)" capability name): unaffected.

### Out of Scope — Historical and dated records

- `CHANGELOG.md` entries: historical statements about what was true at the time; not rewritten.
- `.moai/docs/harness-delivery-strategy.md` (header carries `작성: 2026-06-03`, L53 states the old coupling) and `.moai/docs/autonomous-workflow-strategy.md` (status header "전략 제안"; L574 mentions ultracode only in a cost-risk row): strategy proposal records, not user-facing documentation; not rewritten.

### Out of Scope — Verification of unresolved facts

- OQ-1 (slider-toggle persistence) live observation: not performed here; the SPEC wording avoids the claim and the slider itself instead.
