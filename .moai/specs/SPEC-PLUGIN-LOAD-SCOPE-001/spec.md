---
id: SPEC-PLUGIN-LOAD-SCOPE-001
title: "Measure what a Claude Code plugin and a Codex plugin can carry versus what must stay in the project scaffold"
version: "0.1.0"
status: draft
created: 2026-10-02
updated: 2026-10-02
author: manager-spec
priority: P1
phase: "v3.2.0"
module: ".moai/reports/t1434"
lifecycle: spec-anchored
tier: M
tags: "plugin,codex,measurement,load-scope,scaffold,mods,design-card-1"
---

# SPEC-PLUGIN-LOAD-SCOPE-001 — moai plugin load-scope measurement

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-10-02 | manager-spec | Initial plan-phase draft (card t1434, Class C, Mods design card 1 of 6, operator-directed 2026-10-02) |

## 1. Problem Statement

`moai init` copies roughly 500 template files into every project (311 skills, 92 rules, 48 hooks,
per the card). Later design cards propose moving part of that payload into a Claude Code plugin
(and a Codex plugin) distributed through a marketplace, and shrinking `moai init` accordingly.
Whether that is possible depends on one question that nobody in this repository has answered by
observation: **which component kinds does a plugin actually deliver into a live session, and
which only work when they sit in the project's own `.claude/` / `.codex/` / `AGENTS.md` tree?**

The grounding design notes (the card's artifact link) name the open questions themselves: (1) can a
plugin carry always-loaded rules and settings-style hooks the way moai delivers them today, and (2)
can the mod pieces (`hooks/hooks.json` with a `{"modules": [...]}` payload) ride in a plugin. Those
notes are documentation reading. This SPEC plans the measurement that replaces reading with
observation.

This SPEC plans a **measurement and a verdict table**, not a product feature. It changes no moai
product code, no template, no `.mcp.json`, and no settings.

## 2. Observed Premises (plan-phase measurements, this run, tree 802a72235)

Every line below was observed by a command run during plan authoring, against the installed
binaries named. They are premises the plan relies on; they are not verdict cells.

| Observation | Command | Verbatim output (deciding line) |
|-------------|---------|---------------------------------|
| Claude Code version | `claude --version` | `2.1.287 (Claude Code)` |
| Codex version | `codex --version` | `codex-cli 0.160.0` |
| Claude install takes a marketplace id, not a path | `claude plugin install --help` | `Install a plugin from available marketplaces (use plugin@marketplace for specific marketplace)` |
| Claude can add a path-sourced marketplace | `claude plugin marketplace --help` | `add [options] <source>      Add a marketplace from a URL, path, or GitHub repo` |
| Claude can load a plugin directory for one session | `claude --help` | `--plugin-dir <path>                   Load a plugin from a directory or .zip` |
| Claude has a static validator with a strict mode | `claude plugin validate --help` | `Validate a plugin or marketplace manifest, or the skills, agents, and commands in a directory` and `--strict    Treat warnings as errors (exit 1).` |
| Claude has an inventory command taking a name | `claude plugin details --help` | `Show a plugin's component inventory and projected token cost` (usage `details [options] <name>`) |
| Codex has no validator | `codex plugin --help` | commands listed: `add`, `list`, `marketplace`, `remove`, `help` |
| Codex can add a local-path marketplace | `codex plugin marketplace add --help` | `Add a local or Git marketplace to the configured marketplace sources` |
| moai's settings hook wiring shape | `grep -n -m6 CLAUDE_PROJECT_DIR internal/template/templates/.claude/settings.json.tmpl` | hook entries run `bash -c '[ -f "$0" ] && exec bash "$0"; ...'` with argument `"${CLAUDE_PROJECT_DIR}/.claude/hooks/moai/handle-session-start.sh"`; `"timeout": 5` present on one entry, `"timeout": 30` on others |
| The real plugin state is not quiescent | `ls -la ~/.claude/plugins` | `known_marketplaces.json` and `.last_inuse_sweep` carry modification times from the current day, written by the runtime rather than by this plan |

Two consequences follow, and the plan encodes both:

- A fixture plugin cannot be installed from a bare path. The fixture needs a **local marketplace**
  wrapper for `claude plugin install`, and `--plugin-dir` is the install-free alternative that writes
  no registry.
- The real `~/.claude/plugins` state drifts without any probe action. A "byte-identical after
  restoration" proof must therefore separate probe-caused change from background drift (REQ-003).

Documented but **not yet observed** (the bundled `plugin-authoring` skill text, this run): a mod is a
plugin whose `hooks/hooks.json` is `{ "modules": ["./register.tsx"] }` with a `hooks/register.tsx`
module, validated by `claude plugin validate <mod folder>` and testable by `claude plugin test
<mod folder>`; `CLAUDE_CODE_PLUGIN_DIRS` is named as a way to point a session at plugin folders.
These are hypotheses for the run phase to observe, not facts.

## 3. Requirements (GEARS)

- REQ-001 (Ubiquitous): The measurement shall run in a freshly created project directory and fixture plugin tree under the session scratchpad or a uniquely named `/tmp` directory, and shall write into the repository tree nothing outside `.moai/reports/t1434/`.
- REQ-002 (Capability gate): **Where** a tool honors a config-home environment variable (`CLAUDE_CONFIG_DIR` for Claude Code, `CODEX_HOME` for Codex), the probe shall run validate, marketplace, install, details, and enable commands against a scratch config home, and shall prove the isolation by observation: the scratch home receives the registry writes while the real config home's protected-set manifest is unchanged.
- REQ-003 (State-driven): **While** any measured command acts on a real user config home (for example because isolation removes authentication), the probe shall hold a before-snapshot of the protected set (copy plus sha256 manifest), shall derive volatile-by-observation rows from two manifests taken before any probe command, and shall restore and verify the non-volatile rows byte-identical afterwards; a run that leaves a non-volatile protected row changed fails the SPEC.
- REQ-004 (Ubiquitous): Every measured command shall run as one compound invocation `unset <lane environment variable names> && <command>`, shall use quiet or JSON output where the tool offers it, shall redirect output longer than 50 lines to a file, and shall append the exact command line to `evidence/commands.log` with the raw output saved under `evidence/raw/`.
- REQ-005 (Ubiquitous): The fixture set shall contain one minimal single-component plugin per component kind in the matrix of §5, plus one composite plugin carrying all of them, each component carrying a unique sentinel string, so that a failure of one kind cannot mask another and the fixture tree can be hashed.
- REQ-006 (Event-driven): **When** a fixture is measured statically, the probe shall run `claude plugin validate` (plain and `--strict --json`), `claude plugin marketplace add` plus `claude plugin install`, `claude plugin details`, and, for the mod fixture, `claude plugin test`, recording the exit code and the verbatim deciding output lines of each.
- REQ-007 (Event-driven): **When** a mechanical runtime observation channel exists for a component kind (a hook that writes a marker file, a sentinel the session is asked to quote, a skill/agent/command listing, an MCP tool listing), the probe shall load the fixture in a scratch-project session and record the channel output verbatim, together with a positive control showing the same channel fires for the same component placed in the project tree.
- REQ-008 (Event-driven): **When** a runtime channel is blocked (no authentication under the isolated config home, a refused tool, a version gap), the probe shall mark the affected cell UNOBSERVED with the reason, and shall not infer the cell from validate output, install output, or documentation.
- REQ-009 (Event-driven): **When** the settings-hook equivalence row is measured, the probe shall place moai's settings hook shape (a `bash -c` wrapper whose script argument is `${CLAUDE_PROJECT_DIR}`-prefixed, with a 5-second timeout) in a plugin `hooks/hooks.json` and in the project `settings.json`, and shall record for each whether the command executes, whether `CLAUDE_PROJECT_DIR` and `CLAUDE_PLUGIN_ROOT` expand, and whether a command that outlives 5 seconds is cut off.
- REQ-010 (Event-driven): **When** the rules and instruction rows are measured, the probe shall observe whether a plugin-shipped rules directory (always-loaded and `paths:`-scoped) and a plugin-shipped `CLAUDE.md` or `AGENTS.md` equivalent are loaded, and at what moment, and shall accept a PROJECT-ONLY verdict only when the sentinel was observed absent while the positive control (the same file in the project tree) was observed present.
- REQ-011 (Event-driven): **When** the Codex surface is measured, the probe shall build a `.codex-plugin` fixture per component kind, run `codex plugin marketplace add` (local path), `codex plugin add`, and `codex plugin list` against a scratch `CODEX_HOME`, observe runtime activity through `codex exec` where authentication allows, and mark the remaining cells UNOBSERVED with the reason, there being no Codex validator.
- REQ-012 (Ubiquitous): Every verdict cell shall have an evidence card carrying five labeled fields — Claim, Evidence (command plus verbatim output, citing a file under `evidence/raw/`), Baseline-attribution (claude and codex version, date, fixture sha256), Gaps, Residual-risk — and `claude --version` and `codex --version` shall be read at the start and at the end of the run and agree, or the batch is invalid.
- REQ-013 (Ubiquitous): `verdict.md` shall present one row per component with the cell vocabulary PLUGIN-OK, PROJECT-ONLY, PARTIAL(condition), UNOBSERVED(reason) for the Claude plugin and the Codex plugin, a recommended-home column, and a consequence column naming the effect on the follow-up cards (marketplace, init shrink), and shall list the user-global install-mode removal question as an explicit non-goal.
- REQ-014 (Unwanted behavior): The measurement shall not modify moai product code, templates, `.mcp.json`, or repository settings, and shall not modify the payload of any plugin already installed under `~/.claude/plugins/` or `~/.codex/` (those are read-only reference).

## 4. Constraints

- Plugin and mod surfaces are new and move between versions; every claim carries the version stamps `claude 2.1.287` and `codex-cli 0.160.0` (or whatever `--version` reports at run start) and is not carried forward to another version.
- `claude plugin install` consumes a marketplace id; the fixture is therefore installed through a local marketplace wrapper whose `.claude-plugin/marketplace.json` lives in the scratch tree.
- Lane environment variables (`MOAI_KANBAN`, `MOAI_KANBAN_ID`, `MOAI_KANBAN_LABEL`, `MOAI_KANBAN_LEAD_ADDR`, `MOAI_KANBAN_SETTINGS_INJECTED`, and every `MOAI_FACTORY_*` name present at run start) are scrubbed inside the same invocation as the measured command, because a separate `unset` call does not reach the next command.
- Probe processes that may hang (a hook sleeping past its timeout, a session waiting on a keychain prompt) run under an external `timeout` wrapper.
- No `go test ./...` or other repository-wide verification is run; this SPEC touches no Go code.
- No time estimates; ordering is by priority and milestone sequence.
- Tier M is proposed because the work produces three plan artifacts' worth of decisions and a bounded deliverable set (verdict, evidence, one probe script, one checker script) with no product code; it exceeds Tier S because the acceptance set has more than 8 criteria and one criterion guards user-global state, and it does not reach Tier L because no design.md or research.md carries information not already in plan.md.

## 5. Component Matrix and Verdict Table Schema

Rows (ids are fixed so a checker can count them). The Claude cell and the Codex cell of each row are
independent verdicts; a row whose component has no Codex counterpart is measured by looking for one
and, finding none, marked `UNOBSERVED(no-equivalent-surface-found)`.

| ID | Component | Claude measurement focus | Codex measurement focus |
|----|-----------|--------------------------|-------------------------|
| R01 | skills | plugin `skills/` vs project `.claude/skills` | plugin skills vs `.agents/skills` mirror |
| R02 | agents | plugin `agents/` vs project `.claude/agents` | plugin agents vs `.codex` agent surface |
| R03 | commands | plugin `commands/` vs project `.claude/commands` | equivalent surface, if any |
| R04 | mods | `hooks/hooks.json` `{"modules": [...]}` payload; `plugin test` | equivalent surface, if any |
| R05 | output styles | plugin `output-styles/` vs project `.claude/output-styles` | equivalent surface, if any |
| R06 | MCP servers | plugin `.mcp.json` / manifest vs project `.mcp.json` | plugin MCP vs `.codex/config.toml` |
| R07 | plugin hooks | plugin `hooks/hooks.json` event hooks | plugin hooks vs `.codex/hooks.json` |
| R08 | settings-hook equivalence | moai settings.json wiring shape, `${CLAUDE_PROJECT_DIR}` quoting, 5s timeout, via a plugin | settings-style wiring via a plugin |
| R09 | always-loaded rules | plugin rules dir vs project `.claude/rules/**` | rules/instructions equivalent |
| R10 | path-scoped rules | plugin rules with `paths:` frontmatter vs project path-scoped rules | equivalent, if any |
| R11 | instructions file | plugin `CLAUDE.md` vs project `CLAUDE.md` | plugin `AGENTS.md` vs project `AGENTS.md` |
| R12 | plugin settings keys | plugin `settings.json` keys | plugin config keys, if any |
| R13 | userConfig | manifest `userConfig` options, `claude plugin configure` | manifest equivalent, if any |
| R14 | manifest fields | `.claude-plugin/plugin.json` fields, validate strictness | `.codex-plugin` manifest fields, install acceptance |

Cell vocabulary: `PLUGIN-OK`, `PROJECT-ONLY`, `PARTIAL(<condition>)`, `UNOBSERVED(<reason>)`. Each
cell cites its evidence card as `[cell:<ID>-<claude|codex>]`. The recommended-home column takes
`plugin`, `project-scaffold`, or `split(<which part>)`. The consequence column states in one clause what the
row means for the marketplace card and for the init-shrink card.

## 6. Non-Functional Constraints

- A measurement that leaves the protected set of any real config home changed fails the SPEC regardless of how the verdict table reads.
- The evidence directory holds the deciding lines verbatim; raw output over 50 lines is stored as a file and referenced, not pasted.
- The deliverable shall be re-runnable: the probe script rebuilds the fixtures from scratch in a new unique directory and reproduces the command sequence.

## 7. Out of Scope

### Out of Scope — Why the February 2026 user-global install mode was removed

- This is a history question about a past product decision, not measurable on this machine with these tools; `verdict.md` lists it as an explicit non-goal and answers nothing about it.

### Out of Scope — Marketplace publishing and init shrink

- No marketplace is published and no `moai init` behavior, template, or distribution manifest changes; those are later design cards that consume this verdict.

### Out of Scope — Mod implementation

- No real moai mod is written; the mod fixture carries only a sentinel-bearing minimal module sufficient to measure load scope (card 2 onward owns real mods).

### Out of Scope — Product code, templates, and settings

- No change to moai Go code, `internal/template/templates/`, `.mcp.json`, or any settings file; no push, pull request, deletion, or external issue (those remain the leader's).

### Out of Scope — Cross-version generalization

- Verdicts hold for the stamped versions only; no claim is made for any other Claude Code or Codex version.

## 8. Dependencies, Assumptions, and Prior Art

- Prior art (read-only): the real dual-manifest plugins from the moai-cowork marketplace installed under `~/.claude/plugins/` (for example moai-analyst, moai-lawyer) show what shipped plugins carry; they are ground truth for layout only and are never modified.
- Assumption A1: `CLAUDE_CONFIG_DIR` and `CODEX_HOME` redirect all plugin registry writes. Unverified at plan time; REQ-002 makes it an observed result of the first run-phase step, with the real-config protocol of REQ-003 as the fallback.
- Assumption A2: a runtime observation under an isolated Claude config home may be blocked by authentication (the stored login may not follow a fresh config home). Where blocked, REQ-003's real-config protocol using `--plugin-dir` (no registry write) is the preferred route to a runtime observation, and a still-blocked cell is UNOBSERVED.
- Assumption A3: `claude plugin details <name>` requires an installed plugin rather than a path. Unverified; whichever form works is recorded.
- Assumption A4: the Claude-side `plugin test` and `CLAUDE_CODE_PLUGIN_DIRS` named by the bundled skill exist in 2.1.287; the run phase observes them, and a miss is recorded rather than worked around.
- Related: the grounding design notes for card t1434 (external artifact, not committed here).
