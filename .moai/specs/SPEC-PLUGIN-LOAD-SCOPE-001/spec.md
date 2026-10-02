---
id: SPEC-PLUGIN-LOAD-SCOPE-001
title: "Measure what a Claude Code plugin and a Codex plugin can carry versus what must stay in the project scaffold"
version: "0.3.0"
status: in-progress
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
| 0.2.0 | 2026-10-02 | manager-spec | Iteration 2 of the plan-audit loop (`harness.yaml` `plan_audit_tier_ceilings.M` is 2). Answers `.moai/reports/t1434/plan-audit.md` (FAIL 0.69, threshold 0.80, defects D1-D21). Config-home model restated: static measurement runs under scratch homes, runtime observation runs through the resolved real profile observe-only. REQ-003 rewritten observe-only (no restore), per-command manifest attribution replaces the volatile-row device, checker criteria made able to fail on every counter, yield floor added, R04 runtime channel, Codex marketplace manifest, nested-session contamination statement, R11 configuration scope, durable carrier named. Requirements 14 to 16, acceptance criteria 13 to 16 (machinery criteria merged, tier kept at M). |
| 0.3.0 | 2026-10-02 | manager-spec | Iteration 3 of the plan-audit loop. Tier M plan-audit ceiling (2) exceeded with leader approval (2026-10-02). Answers `.moai/reports/t1434/plan-audit-iter2.md` (FAIL 0.73, threshold 0.80, findings F1-F17) and applies the leader's scope trim. Deleted: the composite (15th) fixture, the old AC-006 (fixture counts and sentinel uniqueness), the mutants of the `check-evidence.sh` modes and the `negative-controls` self-mutant. REQ-006 stays, reworded without the composite and the uniqueness counter, and is now covered by AC-006 (static evidence: fixture count equal to the row count, validate, install and details on every fixture), AC-013 (Codex manifests accepted) and AC-015 (fixture hash recomputed). The mutant guarantee stays for every counter of every `check-verdict.sh` mode. Changed: R05, R12, R13 are `static` rows (10 runtime rows, run cap 60); the control session is an allowed real-home class; floor cards carry a `STATIC-LINE` quote tied to a row token; AC-001 is pinned to a run-start SHA; the verb lists and the live name re-enumeration moved into the checker; manager-develop is named as the writer of `progress.md §E.2`/`§E.3` and of the status flip; the blocker and LEAK contracts are stated; the R08 final arguments are fixed; requirement lines are written as `- REQ-001: (Pattern) ...`. Acceptance ids were renumbered (old 007-016 are now 006-015). Requirements 16, acceptance criteria 15. |

## 1. Problem Statement

`moai init` copies 601 template files into every project (311 files under `.claude/skills`, of which 37
are `SKILL.md`; 92 rules; 48 hooks; counts measured by the plan-auditor at tree 676293144, report § E4).
Later design cards propose moving part of that payload into a Claude Code plugin (and a Codex plugin)
distributed through a marketplace, and shrinking `moai init` accordingly. Whether that is possible
depends on one question that nobody in this repository has answered by observation: **which component
kinds does a plugin actually deliver into a live session, and which only work when they sit in the
project's own `.claude/` / `.codex/` / `AGENTS.md` tree?**

The grounding design notes (the card's artifact link) name the open questions themselves: (1) can a
plugin carry always-loaded rules and settings-style hooks the way moai delivers them today, and (2)
can the mod pieces (`hooks/hooks.json` with a `{"modules": [...]}` payload) ride in a plugin. Those
notes are documentation reading. This SPEC plans the measurement that replaces reading with
observation.

This SPEC plans a **measurement and a verdict table**, not a product feature. It changes no moai
product code, no template, no `.mcp.json`, and no settings.

## 2. Observed Premises

Rows marked *author* were observed by a command the plan author ran against tree 802a72235 (iteration 1)
or tree 676293144 (iteration 2). Rows marked *auditor* were observed by the plan-auditor in
`.moai/reports/t1434/plan-audit.md` (iteration 1) or `.moai/reports/t1434/plan-audit-iter2.md`
(iteration 2), section named in brackets, and are carried here as that auditor's observation; rows
marked *orchestrator* were re-measured by the orchestrator in this run and are carried as stated. None of
these rows is a verdict cell.

| Observation | Command | Verbatim output (deciding line) | Source |
|-------------|---------|---------------------------------|--------|
| Claude Code version | `claude --version` | `2.1.287 (Claude Code)` | author, iteration 1 |
| Codex version | `codex --version` | `codex-cli 0.160.0` | author, iteration 1 |
| Claude install takes a marketplace id, not a path | `claude plugin install --help` | `Install a plugin from available marketplaces (use plugin@marketplace for specific marketplace)` | author, iteration 1 |
| Claude can add a path-sourced marketplace | `claude plugin marketplace --help` | `add [options] <source>      Add a marketplace from a URL, path, or GitHub repo` | author, iteration 1 |
| Claude can load a plugin directory for one session | `claude --help` | `--plugin-dir <path>                   Load a plugin from a directory or .zip` | author, iteration 1 |
| Claude has a static validator with a strict mode | `claude plugin validate --help` | `--strict    Treat warnings as errors (exit 1).` | author, iteration 1 |
| Claude has an inventory command | `claude plugin details --help` | `Show a plugin's component inventory and projected token cost` | author, iteration 1 |
| Claude has a mod test runner | `claude plugin test --help` | `Run a mod's tests` (runs `*.test.ts`, exits 1 when a test fails) | auditor [E1, D7] |
| Codex has no validator | `codex plugin --help` | commands listed: `add`, `list`, `marketplace`, `remove`, `help` | author, iteration 1 |
| Codex can add a local-path marketplace | `codex plugin marketplace add --help` | `Add a local or Git marketplace to the configured marketplace sources` | author, iteration 1 |
| The real SessionStart hook entry in the template | `Read internal/template/templates/.claude/settings.json.tmpl` lines 14-19 | an object with `"command": "bash"`, `"args": ["-c", <wrapper script>, "${CLAUDE_PROJECT_DIR}/.claude/hooks/moai/handle-session-start-navigator.sh"]`, `"timeout": 5`, `"type": "command"`; no template directive on those lines | author, iteration 2 |
| The lane session's Claude config home is the shared profile | `echo $CLAUDE_CONFIG_DIR` | `/Users/goos/.moai/claude-profiles/moai-adk` (not `~/.claude`); `CODEX_HOME` is unset, so Codex uses `~/.codex` | orchestrator |
| The session environment carries lane and session variables | `env` (names only) | 13 `CLAUDE_CODE_*` names (BRIDGE_SESSION_ID, CHILD_SESSION, ENABLE_TODO_TOOLS, ENTRYPOINT, EXECPATH, EXPERIMENTAL_AGENT_TEAMS, MAX_CONCURRENT_SUBAGENTS, MESSAGING_SOCKET, MESSAGING_TOKEN, NO_FLICKER, SESSION_ATTENDED, SESSION_ID, STOP_HOOK_BLOCK_CAP) and 14 `MOAI_*` names (AUTONOMY_TIER, CONFIG_SOURCE, FACTORY_AUTO_DISPATCH, FACTORY_CLEAR_POLICY, FACTORY_ROLE, FACTORY_WORKER, FACTORY_WORKERS, KANBAN_BACKEND, KANBAN_ID, KANBAN_SETTINGS_INJECTED, LAUNCH_PROVIDER, PROFILE_LEASE_TOKEN, PROJECT_DIR, SESSION_PID) | orchestrator |
| Authentication does not follow a fresh config home | `CLAUDE_CONFIG_DIR=<scratch> claude auth status` | `"loggedIn": false, "authMethod": "none"` | auditor [E3] |
| Static commands work under a scratch config home | validate, `marketplace add`, `install`, `details` under `CLAUDE_CONFIG_DIR=<scratch>` | exit 0 each; the scratch home gained `plugins/`, `settings.json`, `.claude.json` | auditor [E3] |
| `validate` prints a context-file message | `claude plugin validate <plugin with CLAUDE.md>` | `CLAUDE.md at the plugin root is not loaded as project context. To ship context with your plugin, use a skill ...` | auditor [E3] |
| `details` has a fixed category set | `claude plugin details p-rules@audit-mkt` | `Skills (0) Agents (0) Hooks (0) MCP servers (0) LSP servers (0)` and an `Always-on` token line; no rules, commands, output styles, or mods category | auditor [E3] |
| The profile carries plugin-affecting settings | read of the profile `settings.json` | `enabledPlugins` (gopls-lsp, vercel, typesafe), `pluginConfigs."agents-md@builtin".options.instructionFiles = "claude-md-or-agents-md"`, `remoteControlAtStartup: true`; account-synced plugins sit under `plugins/synced/` | auditor [D1, E2] |
| The default home drifts for reasons unrelated to this lane | `ls -la ~/.claude/plugins` | `known_marketplaces.json` and `.last_inuse_sweep` carry current-day modification times, written by a session that is not this lane | auditor [E2] |
| A Codex marketplace uses its own manifest path | `find ~/.codex/.tmp/bundled-marketplaces -maxdepth 4` | `openai-bundled/.agents/plugins/marketplace.json` | auditor [E6] |
| Report paths are local-only | `git check-ignore -v .moai/reports/t1434/verdict.md` | `.gitignore:235:.moai/reports/*	.moai/reports/t1434/verdict.md`; `git status --porcelain --ignored -- .moai/reports/t1434` prints `!! .moai/reports/t1434/` | author, iteration 2 |
| Launch options exist that may change what a nested session loads | `claude --help`; `codex exec --help` | claude lists `--safe-mode`, `--bare`, `--setting-sources`, `--strict-mcp-config`, `--settings`, `--no-session-persistence`, `--include-hook-events`; codex exec lists `-c`, `--ignore-user-config`, `--ephemeral`, `--skip-git-repo-check` | auditor [iteration 2, E6] |

Consequences the plan encodes:

- A fixture plugin cannot be installed from a bare path. The fixture needs a **local marketplace**
  wrapper for `claude plugin install`, and `--plugin-dir` is the install-free alternative that writes
  no registry.
- **Two routes, because authentication does not follow a scratch home.** Static measurement
  (validate, marketplace add, install, details, plugin test, and the Codex equivalents) runs under
  scratch config homes and is isolated by construction. Runtime observation (`claude -p`) needs the
  login and therefore runs through the **resolved real profile**, observe-only: only `claude -p`
  sessions (a plugin run and a control run) and read-only commands, never a registry- or
  settings-writing command (REQ-003, REQ-004).
- The real profile is shared with other sessions and drifts without any probe action. A byte-identity
  claim over it cannot be made. The plan observes instead: a manifest immediately before and after
  each real-home command, with every changed row attributed or declared ambient (REQ-005).
- Report-directory artifacts are **local-only** (`.gitignore:235`). The durable carrier the follow-up
  design cards read is the verdict table plus an evidence index written into this SPEC's
  `progress.md §E.2` at the end of the run phase (REQ-001, REQ-016).

Documented but **not yet observed** (the bundled `plugin-authoring` skill text, iteration 1): a mod is
a plugin whose `hooks/hooks.json` is `{ "modules": ["./register.tsx"] }` with a `hooks/register.tsx`
module, validated by `claude plugin validate <mod folder>` and testable by `claude plugin test <mod
folder>`; hot reload is off under `claude -p` (nobody can be asked); `CLAUDE_CODE_PLUGIN_DIRS` is named
as a way to point a session at plugin folders. `claude plugin test` is now observed to exist (table
above); everything else in this paragraph remains a hypothesis for the run phase, and
`CLAUDE_CODE_PLUGIN_DIRS` is recorded UNOBSERVED unless measured.

## 3. Requirements (GEARS)

- REQ-001: (Ubiquitous) The measurement shall build its fixture plugin tree and scratch project under the session scratchpad or a uniquely named `/tmp` directory, and shall write into the repository tree only (a) files under `.moai/reports/t1434/`, which are local-only because `.gitignore:235` ignores `.moai/reports/*`, and (b) in the one run-phase commit, written by manager-develop as a non-cycle evidence-writing delegation (the owner of `progress.md §E.2`, `§E.3` and of the `draft → in-progress` transition in the ownership matrix of `spec-frontmatter-schema.md`), the `§E.2` block (the durable verdict table and evidence index) and the `§E.3` block (the run-phase audit-ready signal) of this SPEC's `progress.md` and the `status:` and `updated:` frontmatter lines of this SPEC's `spec.md`; it shall not modify moai product code, templates, `.mcp.json`, repository settings, any other line of this SPEC's artifacts, or the payload of any plugin already installed in a real config home.
- REQ-002: (Ubiquitous) Before any other measured command the probe shall record in `evidence/env.txt` the values of `CLAUDE_CONFIG_DIR` and `CODEX_HOME`, the output of `claude auth status`, and the names (never the values) of every `CLAUDE_CODE_*` and `MOAI_*` environment variable live at that moment; every measured command shall then run as one compound invocation `unset <those names> && <command>`, with the name list enumerated from the live environment by the probe at run time and not fixed in this SPEC, and shall append one `CMD:` line carrying its id, its home class (`scratch` or `real`), and the scrub list to `evidence/commands.log`, with output longer than 50 lines redirected to a file under `evidence/raw/`; the checker shall enumerate those names again from its own environment at check time and require each of them to appear in `evidence/env-scrub.txt`.
- REQ-003: (Capability gate) **Where** a tool honors a config-home environment variable (`CLAUDE_CONFIG_DIR` for Claude Code, `CODEX_HOME` for Codex), the probe shall run every registry- or settings-writing command (validate, marketplace add, install, details, enable, plugin test, and the Codex marketplace add, plugin add and plugin list) only with that variable pointing at an initially empty scratch home, shall show the isolation by a before and after file listing of the scratch home going from empty to non-empty, and shall not copy credentials, keychain entries, or login state into any scratch home.
- REQ-004: (State-driven) **While** a measured command runs against a resolved real config home (the profile `CLAUDE_CONFIG_DIR` names, or `~/.claude` when unset; `CODEX_HOME` or `~/.codex`), because runtime observation needs the login a scratch home lacks, the probe shall run there only `claude -p` sessions in the scratch project (a plugin run with `--plugin-dir` and a control run without it, each with `--no-session-persistence`) and the read-only commands `--version`, `auth status`, `login status` and `plugin list`, and shall not run install, uninstall, marketplace add, remove or update, enable, disable, configure, plugin update, Codex plugin add or remove, `codex exec`, or any other registry- or settings-writing command against any real home; the forbidden and the read-only verb lists shall be literal text inside the checker script, equal to the lists in this requirement, and not a file the probe writes.
- REQ-005: (Ubiquitous) The probe shall derive the protected set from the resolved real homes (with `~/.claude` and `~/.codex` as secondary read-only watches when they differ), shall take a manifest of it immediately before and after each real-home command and once around the whole static phase, shall report every changed row as `AMBIENT` or `LEAK` (a row is `LEAK` when its changed content names a fixture plugin, the scratch marketplace, the scratch path, or a fixture sentinel; `AMBIENT` means only that it names none of these, never that another lane wrote it), shall stop running further real-home commands at the first `LEAK` and record the command whose manifest pair shows it, shall compare for Codex the `[plugins]` and marketplace tables of `config.toml` and not the whole-file hash, shall declare `.claude.json`, `sessions/`, `session-env/`, `debug/`, `file-history/`, `.codex-global-state.json`, and Codex sqlite state excluded with a stated reason, and shall restore nothing that the probe did not itself create.
- REQ-006: (Ubiquitous) The fixture set shall contain one minimal single-component plugin per row of §5 (no composite plugin), each carrying a sentinel string of the form `SENTINEL_<ID>_<random>` placed where a listing or a marker prints it, a Claude manifest and a Codex manifest, together with a Codex marketplace manifest at `<root>/.agents/plugins/marketplace.json` whose exact schema the probe discovers from rejection messages, and the whole fixture tree shall be covered by one hashed file manifest.
- REQ-007: (Event-driven) **When** a fixture is measured statically, the probe shall run `claude plugin validate` (plain, and with `--strict --json`), `claude plugin marketplace add` plus `claude plugin install`, `claude plugin details`, and, for every fixture holding `*.test.ts` files, `claude plugin test`, each under a scratch home, recording the exit code, the verbatim deciding output lines, and for `plugin test` the number of tests the runner reports executing, a zero count being an empty sweep and not a pass.
- REQ-008: (Event-driven) **When** a mechanical runtime observation channel exists for a component kind (a hook that writes a marker file carrying a per-run token, a skill, agent, command or MCP listing, or the init event of `claude -p --output-format stream-json --verbose` if M1 observes that it lists loaded components), the probe shall load the fixture in a scratch-project session, record the channel output verbatim together with a positive control showing the same channel fires for the same component placed in the project tree, and for a channel that relies on a model-quoted sentinel shall run it three times and accept an absence claim only when all three runs agree, the total number of `claude -p` runs (the M1 observation runs included) staying within a cap the probe derives and prints as 10 runtime rows times 2 (plugin and control) times 3 runs, which is 60, the runtime rows being the rows whose `channels` column in §5 is `static+runtime`.
- REQ-009: (Event-driven) **When** a runtime or static channel is blocked (no authentication, a refused command, a version gap, a control that did not fire, the run cap), the probe shall mark the affected cell `UNOBSERVED` with a reason that cites a raw evidence file and quotes the line of that file proving the block, and shall not infer the cell from validate output, install output, or documentation; for a row whose `channels` column in §5 is `static`, no runtime channel was ever designed, so its Claude cell is the static judgment (`PLUGIN-OK` only where `claude plugin details` lists the component, otherwise `PARTIAL(static-only: <what the tool reported>; runtime=UNOBSERVED(no-runtime-channel))`) and the text `runtime=UNOBSERVED(no-runtime-channel)` needs no raw quote.
- REQ-010: (Event-driven) **When** the settings-hook equivalence row (R08) is measured, the probe shall place a byte-for-byte copy of the template's SessionStart entry (the object at `internal/template/templates/.claude/settings.json.tmpl` whose script argument ends in `handle-session-start-navigator.sh` and whose `timeout` is 5), with only its final script-path argument changed, in a plugin `hooks/hooks.json` with the final argument `${CLAUDE_PLUGIN_ROOT}/hooks/marker.sh` and in the project `settings.json` with the final argument `${CLAUDE_PROJECT_DIR}/marker.sh`, where each marker script prints the argument it received (`$0`) as `ARG0=<value>`, and shall record for each copy whether the command executes and what its argument expanded to (`expanded`, `literal` or `empty`, recomputed from `ARG0`), and for the plugin copy whether a command that outlives 5 seconds is cut off, a cutoff meaning that the end marker `END` is absent and the session end reading minus `START` exceeds 5, and the observed seconds being `END` minus `START`, defined only when `END` exists.
- REQ-011: (Event-driven) **When** the rules and instruction rows (R09, R10, R11) are measured, the probe shall observe whether a plugin-shipped always-loaded rules directory, a plugin-shipped `paths:`-scoped rule (at session start and after a matching path is touched), and a plugin-shipped `CLAUDE.md` or `AGENTS.md` equivalent are loaded, shall accept `PROJECT-ONLY` only when the plugin sentinel was absent in three of three runs while the project-tree control sentinel was present, shall record the profile's `pluginConfigs."agents-md@builtin".options.instructionFiles` value as the configuration scope of R11, and shall record the verbatim `claude plugin validate` context-file message as a corroborating observation that decides no cell.
- REQ-012: (Event-driven) **When** the mods row (R04) is measured, the probe shall build a mod fixture whose hook module writes a marker through `$.fs` on `session.start` and which carries a `*.test.ts` holding exactly one test, shall load it with `--plugin-dir` in the session type the probe can drive, shall run the same marker logic as an ordinary plugin hook in the same session as a control that `session.start` hooks fire there, and shall mark the R04 runtime result `UNOBSERVED` or `PARTIAL(static-only)` with a quoted reason when the control fired and the mod marker did not.
- REQ-013: (Event-driven) **When** the Codex surface is measured, the probe shall build a `.codex-plugin` fixture per component kind, run `codex plugin marketplace add` (local path), `codex plugin add`, and `codex plugin list` only under a scratch `CODEX_HOME`, count a fixture as added only when `plugin add` exited 0 and `plugin list` names it, observe runtime activity through `codex exec` only under a scratch `CODEX_HOME` and only where a route exists that writes no registry and copies no credential, and mark every remaining Codex cell `UNOBSERVED` with the reason of REQ-009, there being no Codex validator.
- REQ-014: (Ubiquitous) Every verdict cell shall have an evidence card carrying five labeled fields (Claim, Evidence naming a file under `evidence/raw/` and quoting its deciding line, Baseline-attribution stating the claude and codex versions, the date and the fixture sha256, Gaps, Residual-risk), the quoted deciding line shall not be a stamp line or an `EXIT=` trailer and, for a Claude cell that is not `UNOBSERVED`, shall name the row's sentinel prefix `SENTINEL_<ID>_`, its fixture plugin name, or its details token as declared in plan.md, every raw file shall begin with a stamp line carrying those same versions and hash, and the readings of `claude --version` and `codex --version` taken at the start and at the end of the run shall be equal or the batch is invalid.
- REQ-015: (Ubiquitous) The verdict shall carry, for every row of §5 marked `floor=yes`, a Claude cell that is not `UNOBSERVED` and whose card carries a `STATIC-LINE:` quote of a line of that row's own static raw file (validate, details, or install output) that names the row's sentinel prefix, fixture plugin name or details token.
- REQ-016: (Ubiquitous) `verdict.md` shall present one row per row of §5 with the cell vocabulary `PLUGIN-OK`, `PROJECT-ONLY`, `PARTIAL(condition)`, `UNOBSERVED(reason)` for the Claude plugin and the Codex plugin, a recommended-home column drawn from `plugin`, `project-scaffold`, `split(part)`, `undetermined(reason)` and consistent with the Claude cell, a consequence column naming the effect on the marketplace card and on the init-shrink card, and the line `Non-goal: user-global install mode removal`, and the same 14-row table with the sha256 of `verdict.md` and an evidence index shall be written into `progress.md §E.2` as the durable carrier.

## 4. Constraints

- Plugin and mod surfaces are new and move between versions; every claim carries the version stamps `claude 2.1.287` and `codex-cli 0.160.0` (or whatever `--version` reports at run start) and is not carried forward to another version.
- `claude plugin install` consumes a marketplace id; the fixture is therefore installed through a local marketplace wrapper whose `.claude-plugin/marketplace.json` lives in the scratch tree.
- The scrub list is enumerated from the live environment by the probe at run time (REQ-002). The names in §2 are an illustration of this session's environment, not the list; a name the probe keeps because scrubbing it breaks authentication is listed in `evidence/env-scrub.txt` with the observed reason.
- **Nested-session contamination** is named here and not mitigated away. A nested `claude -p` against the resolved real profile is not a clean room: (a) inherited `CLAUDE_CODE_*` variables (session id, bridge session id, messaging socket and token, entrypoint, child-session flag) would let the child join cross-session messaging and the live-session registry, which the scrub of REQ-002 removes; (b) the profile's `enabledPlugins`, the account-synced plugins under `plugins/synced/`, and the MCP servers they carry load into the child with their skills and hooks, so the probe counts only lines carrying its own per-run token; whether any launch option (`--safe-mode`, `--bare`, `--setting-sources`, `--strict-mcp-config`, `--no-session-persistence`, `--include-hook-events`) removes those ambient components for a `--plugin-dir` session is not known and is observed at M1 (status: assumption until measured), and for Codex `--ignore-user-config` and `--ephemeral` are read from `codex exec --help` only; (c) `remoteControlAtStartup: true` in the profile may register the child as a Remote Control session; (d) project `.mcp.json` and project hooks in the scratch project may need an approval a non-interactive session cannot give, and a control that fails to fire for that reason is a blocked channel (REQ-009), not a negative result; (e) every run costs model tokens, bounded by the derived run cap of REQ-008.
- Every nested `claude -p` runs with `--no-session-persistence`. That option does not stop every profile write (project trust in `.claude.json`, history), so those writes are named in the verdict Gaps; a changed protected row that names no fixture identifier is `AMBIENT`, which does not mean another lane wrote it.
- The real profile is also shared with other lanes and the Codex app, so the probe never restores, rewrites, or removes anything there; a changed protected row is reported, attributed `AMBIENT` or `LEAK`, and left alone.
- Running `probe.sh` executes commands the worktree guard does not read one by one; the probe's commands are therefore constrained by REQ-003 to REQ-005 on their own terms and audited by the checker from `commands.log`, and the verdict's Gaps record that the guard did not mediate them. No `git init` is run; a hook, rule, or skill observation under `claude -p` needs no repository, and if a later observation proves to need one the probe script, not an interactive Bash call, creates it and the verdict names the fact.
- Probe processes that may hang (a hook sleeping past its timeout, a session waiting on a prompt) run under an external `timeout` wrapper.
- No `go test ./...` or other repository-wide verification is run; this SPEC touches no Go code.
- No time estimates; ordering is by priority and milestone sequence.
- Tier M is kept: 16 requirements (the ceiling) and 15 acceptance criteria (`spec-workflow.md` § SPEC Complexity Tier). Machinery criteria were merged in iteration 2 and trimmed in iteration 3 rather than tiering up, because no `design.md` or `research.md` would carry information not already in `plan.md`.

## 5. Component Matrix and Verdict Table Schema

Rows (ids are fixed so a checker can count them). The Claude cell and the Codex cell of each row are
independent verdicts; a row whose component has no Codex counterpart is measured by looking for one
and, finding none, marked `UNOBSERVED` with a reason citing the search output. The `floor` column marks
the documented standard components whose Claude cell must not be `UNOBSERVED` (REQ-015); the `channels`
column says whether a runtime channel exists for the row: `static+runtime` rows (R01-R04, R06-R11, ten
rows) are swept at runtime, `static` rows (R05, R12, R13, R14) have no runtime channel and their Claude
cell is the static judgment of REQ-009.

| ID | Component | Claude measurement focus | Codex measurement focus | floor | channels |
|----|-----------|--------------------------|-------------------------|-------|----------|
| R01 | skills | plugin `skills/` vs project `.claude/skills` | plugin skills vs `.agents/skills` mirror | yes | static+runtime |
| R02 | agents | plugin `agents/` vs project `.claude/agents` | plugin agents vs `.codex` agent surface | yes | static+runtime |
| R03 | commands | plugin `commands/` vs project `.claude/commands` | equivalent surface, if any | yes | static+runtime |
| R04 | mods | `hooks/hooks.json` `{"modules": [...]}` payload; `plugin test`; marker through `$.fs` | equivalent surface, if any | no | static+runtime |
| R05 | output styles | plugin `output-styles/` vs project `.claude/output-styles` | equivalent surface, if any | yes | static |
| R06 | MCP servers | plugin `.mcp.json` / manifest vs project `.mcp.json` | plugin MCP vs `.codex/config.toml` | yes | static+runtime |
| R07 | plugin hooks | plugin `hooks/hooks.json` event hooks | plugin hooks vs `.codex/hooks.json` | yes | static+runtime |
| R08 | settings-hook equivalence | the template SessionStart entry, `${CLAUDE_PROJECT_DIR}` expansion, 5s timeout, via a plugin | settings-style wiring via a plugin | no | static+runtime |
| R09 | always-loaded rules | plugin rules dir vs project `.claude/rules/**` | rules/instructions equivalent | no | static+runtime |
| R10 | path-scoped rules | plugin rules with `paths:` frontmatter vs project path-scoped rules | equivalent, if any | no | static+runtime |
| R11 | instructions file | plugin `CLAUDE.md` vs project `CLAUDE.md`, scoped to the profile's `instructionFiles` setting | plugin `AGENTS.md` vs project `AGENTS.md` | no | static+runtime |
| R12 | plugin settings keys | plugin `settings.json` keys | plugin config keys, if any | no | static |
| R13 | userConfig | manifest `userConfig` options, `claude plugin configure` (observed under a scratch home only) | manifest equivalent, if any | no | static |
| R14 | manifest fields | `.claude-plugin/plugin.json` fields, validate strictness | `.codex-plugin` manifest fields, install acceptance | yes | static |

Cell vocabulary: `PLUGIN-OK`, `PROJECT-ONLY`, `PARTIAL(<condition>)`, `UNOBSERVED(<reason>; raw=<file>; quote="<line>")`.
A cell is classified by the vocabulary token before its first `(`.
`PLUGIN-OK` needs an inventory or runtime observation; a cell resting on static acceptance alone is
`PARTIAL(static-only: <what the tool reported>)`; for a `static` row the text
`runtime=UNOBSERVED(no-runtime-channel)` is appended inside the `PARTIAL(...)` condition (REQ-009).
Each cell cites its evidence card as `[cell:<ID>-<claude|codex>]`. The recommended-home column takes
`plugin`, `project-scaffold`, `split(<which part>)`, or `undetermined(<reason>)`; consistency with the
Claude cell is fixed by REQ-016 (`plugin` needs `PLUGIN-OK`, `project-scaffold` needs `PROJECT-ONLY`,
`split` needs `PARTIAL`, `UNOBSERVED` needs `undetermined`). The consequence column states, as
`marketplace: ...` and `init-shrink: ...`, what the row means for each follow-up card.

## 6. Non-Functional Constraints

- A measurement in which a protected row of any real config home changes and the changed content names a fixture, the scratch marketplace, the scratch path, or a sentinel (a `LEAK`) stops further real-home commands at once and records the attributing command. A `LEAK` produced by a forbidden verb fails the SPEC regardless of how the verdict table reads; a `LEAK` produced by a permitted verb is a measured finding, written into the verdict Gaps as `LEAK-FINDING: <cmd-id>`, and does not fail the SPEC by itself. A changed row that names none of these is reported `AMBIENT` and fails nothing.
- Blocked outcomes: a floor row without a non-`UNOBSERVED` Claude cell, or every runtime row blocked, is reported by the lane's blocker report to the leader; the card then reports `blocked-on-measurement` with the partial table and the evidence path, and the leader chooses PASS-with-debt, re-plan or abandon; no criterion is relaxed to avoid it.
- The evidence directory holds the deciding lines verbatim; raw output over 50 lines is stored as a file and referenced, not pasted.
- The deliverable shall be re-runnable: the probe script rebuilds the fixtures from scratch in a new unique directory and reproduces the command sequence.
- Forgery of a raw file by hand is not defended mechanically; the checker catches vacuous, orphaned, mismatched and irrelevant evidence, not a deliberately fabricated file (named in Residual-risk of the verdict).

## 7. Out of Scope

### Out of Scope — Why the February 2026 user-global install mode was removed

- This is a history question about a past product decision, not measurable on this machine with these tools; `verdict.md` lists it as an explicit non-goal and answers nothing about it.

### Out of Scope — Marketplace publishing and init shrink

- No marketplace is published and no `moai init` behavior, template, or distribution manifest changes; those are later design cards that consume this verdict.

### Out of Scope — Mod implementation

- No real moai mod is written; the mod fixture carries only a sentinel-bearing minimal module sufficient to measure load scope (card 2 onward owns real mods).

### Out of Scope — Product code, templates, and settings

- No change to moai Go code, `internal/template/templates/`, `.mcp.json`, or any settings file; no push, pull request, deletion, or external issue (those remain the leader's).

### Out of Scope — Real-home writes and scratch-home credentials

- No registry- or settings-writing command runs against a real Claude or Codex home, and no credential, keychain entry, or login state is copied into a scratch home; a runtime cell that cannot be observed without either is `UNOBSERVED`.

### Out of Scope — Cross-version generalization

- Verdicts hold for the stamped versions only; no claim is made for any other Claude Code or Codex version.

## 8. Dependencies, Assumptions, and Prior Art

- Prior art (read-only): the real dual-manifest plugins from the moai-cowork marketplace show what shipped plugins carry; per the plan-auditor [D20] they sit under the profile's `plugins/synced/` (account-synced, absent from `installed_plugins.json`) and under `~/.codex/plugins/cache/moai-cowork/`, not under `~/.claude/plugins/`. They are ground truth for layout only and are never modified.
- Assumption A1 (resolved by the iteration-1 audit, re-observed in M1): `CLAUDE_CONFIG_DIR` redirects the Claude registry writes of the static commands; M1 observes the same for `CODEX_HOME` and, where Codex does not honor it, takes the `ISOLATION-FAILED` route (all Codex registry-writing commands are then not run and the Codex cells are `UNOBSERVED` with the proof).
- Assumption A2 (resolved): authentication does not follow a scratch home, so runtime observation rides the resolved real profile observe-only; M1 re-observes it with `claude auth status` under a scratch home and a one-line `codex login status`.
- Assumption A3: `claude plugin details` takes `name@marketplace` (observed by the auditor); the bare-name form is not yet tried and whichever form works is recorded.
- Assumption A4: `claude plugin test` exists in 2.1.287 (observed); whether `claude -p` loads mod modules, whether `claude -p --output-format stream-json --verbose` lists loaded components in its init event, which launch options remove the profile's ambient plugins, and what `CLAUDE_CODE_PLUGIN_DIRS` does are observed in M1 and M3, and `CLAUDE_CODE_PLUGIN_DIRS` stays `UNOBSERVED` unless measured.
- Related: the grounding design notes for card t1434 (external artifact, not committed here).
