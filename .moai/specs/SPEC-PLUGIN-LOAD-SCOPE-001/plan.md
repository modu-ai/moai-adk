# Implementation Plan — SPEC-PLUGIN-LOAD-SCOPE-001

## 1. Overview

A measurement, not a build. The run phase builds throwaway fixture plugins in a scratch project and
measures them on **two routes**:

- **Static route (isolated).** validate, marketplace add, install, details, plugin test and the Codex
  marketplace add, plugin add and plugin list run under **scratch config homes** (`CLAUDE_CONFIG_DIR`
  and `CODEX_HOME` pointed at empty directories). The iteration-1 audit observed this work for Claude
  and that the scratch home gains `plugins/`, `settings.json`, `.claude.json`. Credentials are never
  copied into a scratch home (out of bounds, spec §7).
- **Runtime route (resolved real profile, observe-only).** `claude -p` needs the login, which a scratch
  home lacks (`"loggedIn": false`), so runtime observation runs against the **resolved real profile**
  (`$CLAUDE_CONFIG_DIR`; in this lane `/Users/goos/.moai/claude-profiles/moai-adk`, not `~/.claude`).
  Only `claude -p` sessions in the scratch project (a plugin run with `--plugin-dir` and a control run
  without it, each with `--no-session-persistence`) and the read-only verbs `--version`, `auth status`,
  `login status`, `plugin list` run there. Install, uninstall, marketplace add/remove/update,
  enable/disable, configure, plugin update, Codex plugin add/remove, `codex exec`, and any other
  registry- or settings-writing command are **forbidden against any real home**.

Contamination sources of the real-profile route are named, not hidden (spec §4): the profile's
`enabledPlugins` (gopls-lsp, vercel, typesafe), `pluginConfigs."agents-md@builtin".options.instructionFiles`,
account-synced plugins under `plugins/synced/` and their MCP servers, `remoteControlAtStartup: true`,
inherited `CLAUDE_CODE_*` variables and the messaging socket/token, project `.mcp.json`/hook approval in a
non-interactive session, and per-run model cost. Whether a launch option removes the ambient plugins is
observed at M1, not assumed.

Methodology is measurement/analysis (no `cycle_type` code implementation). Run-phase artifacts live under
`.moai/reports/t1434/` and are **local-only** (`.gitignore:235` is `.moai/reports/*`; `git check-ignore -v`
and `git status --porcelain --ignored` observed). The **durable carrier** the follow-up design cards read is
the 14-row verdict table plus an evidence index written into this SPEC's `progress.md §E.2` at the end of
the run phase (REQ-001, REQ-016). That block, the `§E.3` signal and the `status:`/`updated:` lines of
`spec.md` are the only tracked writes the run makes, in one commit written by manager-develop (§7).

Tier M. Ordering follows the card (fixture design, Claude static, Claude runtime, Codex, synthesis) and,
inside the plan, reversibility: M1 settles the config-home routing, the protected set and the checker
instruments, which are the decisions most likely to change; everything after depends on them.

**Proportionality decision (Agent Core Behaviors 4 and 5).** The deliverable is a 14-row table, but each
row holds two independently verdicted cells that a follow-up card cites as `[cell:<ID>-<tool>]`. Per-cell
evidence cards are therefore kept (28 cards, one per cell), authored from a single template that `probe.sh`
scaffolds, because a per-row block would put two independent claims under one citation. Iteration 3 trimmed
the machinery on the leader's decision: no composite fixture, no fixture-count criterion, no mutants for the
evidence-mode checkers. Budgets: 16 requirements, 15 acceptance criteria, Tier M kept.

## 2. Milestones (priority-ordered, no time estimates)

### M1 — Environment record, routing, protected set, instruments (Priority: High)

Most reversible decisions live here: which home each command class touches, the protected set, the scrub
list, the checker contracts.

- Create the scratch root (`$SCRATCH`, unique name under the session scratchpad) with `config-claude/`,
  `config-codex/`, `project/`, `marketplace/`, `marketplace-codex/`, and `ctl-tree/` (manifest-builder
  control). Record paths in `evidence/env.txt`.
- **First measured act, before anything else:** write `evidence/env.txt` with `echo $CLAUDE_CONFIG_DIR`,
  `echo $CODEX_HOME`, `claude auth status` (real profile, read-only), a one-line `codex login status`,
  the live `CLAUDE_CODE_*` and `MOAI_*` **names** (names only; the value of any name matching
  TOKEN, KEY or SECRET is never written), and from the resolved profile `settings.json` the non-secret
  values of `enabledPlugins`, `pluginConfigs."agents-md@builtin".options.instructionFiles` and
  `remoteControlAtStartup`. Then write `evidence/version-start.txt` with the readings of `claude --version`
  and `codex --version` (M5 writes `version-end.txt`; AC-015 compares the two).
- Build `evidence/env-scrub.txt` from the live names (not from a fixed list); build the single
  `unset <names> &&` prefix from it. `CLAUDE_CONFIG_DIR` and `CODEX_HOME` are never scrubbed: they are set
  to the scratch home (static route) or left as resolved (real route). If scrubbing a name breaks
  authentication on the real route, keep that name, list it with the observed reason, and re-run.
- Define the **protected set** from the resolved homes, with `~/.claude` and `~/.codex` as secondary
  read-only watches when they differ. Rows (named so the checker can require them):
  `claude:installed_plugins.json`, `claude:known_marketplaces.json`, `claude:settings.json`,
  `claude:plugins/cache`, `claude:plugins/marketplaces`, `claude:plugins/data`, `claude:plugins/synced`
  (recursive file list plus hash), and `codex:config.toml#plugins-tables` (the `[plugins]` and marketplace
  tables extracted from `config.toml`, hashed as text), `codex:plugins-tree`. M1 writes the first manifest
  of them as `evidence/protected-t0.sha256`. `evidence/protected-excluded.txt` names, each with a reason,
  `.claude.json` (rewritten by any session start), `sessions/`, `session-env/`, `debug/`, `file-history/`,
  `projects/` transcripts, `.codex-global-state.json`, Codex sqlite state.
- Write the **manifest builder** (inside `probe.sh`, git-free). Volatility is **not** classified from two
  back-to-back manifests (the profile drift is intermittent). Instead the probe takes a manifest immediately
  before and after each real-home command (`evidence/manifest/<cmd-id>-before.sha256`, `-after.sha256`,
  `manifest-diff-<cmd-id>.txt`) and once around the whole static phase (`static-window`), and classifies each
  changed row `AMBIENT` or `LEAK` by whether its changed content names a fixture plugin (`p-*`), the scratch
  marketplace, the scratch path, or a `SENTINEL_` string. `AMBIENT` is not "written by another lane"; the
  verdict Gaps say so. Nothing is restored; nothing the probe did not create is touched. At the first `LEAK`
  the probe runs no further real-home command and records the attributing command (§7).
- **Static isolation proof:** list each scratch home before (must be empty) and after one trivial
  registry-writing command per tool under its scratch home (a marketplace add of a trivial marketplace tree:
  `claude plugin marketplace add` for Claude, `codex plugin marketplace add` for Codex; a rejected Codex add
  is retried with a manifest corrected from the rejection message, at most 3 attempts, before the verdict is
  written), recorded as `evidence/scratch-home-before.txt` and `scratch-home-after.txt`, per tool, with the
  output kept as raw files (`claude-isolation-probe.txt`, `codex-isolation-probe.txt`). If a tool does not
  honor its variable the after-listing stays empty: write `ISOLATION-FAILED <tool> raw=<file>` with the
  proof, run no further registry-writing command for that tool, and carry its static cells as `UNOBSERVED`.
  The Codex registry-writing commands of M4 are separate commands that run after this verdict.
- **Observations the plan does not assume** (each recorded in a raw file, none asserted here; the M1
  `claude -p` runs count against the run cap of M3):
  (1) `claude plugin test` on an empty mod folder (the empty-sweep analogue: its output and exit code);
  (2) whether `claude -p --output-format stream-json --verbose` lists loaded skills, agents, plugins or MCP
  servers in its init event (a candidate deterministic channel, run once with `--plugin-dir` on one fixture);
  (3) whether `claude -p` loads `--plugin-dir` mod modules; (4) what `CLAUDE_CODE_PLUGIN_DIRS` does, or
  `UNOBSERVED(plugin-dirs-env-unmeasured)`; (5) `claude plugin details` with a bare name;
  (6) whether the launch options `--safe-mode`, `--bare`, `--setting-sources`, `--strict-mcp-config` and
  `--include-hook-events` remove the profile's ambient plugins and MCP servers for a `--plugin-dir` session
  (init listing with and without each option; status: assumption until measured; `--no-session-persistence`
  is on in every run, as AC-004 requires, and is therefore not toggled), which also says whether an isolated-default configuration for R11 exists; for Codex,
  `--ignore-user-config` and `--ephemeral` are read from `codex exec --help` only, because no real-home
  `codex exec` is allowed and a scratch home has no login.
- Generate the **14 fixtures** (one per row, §3) with their sentinels, and write `evidence/fixture-manifest.txt`
  (sorted `sha256  relpath` of the whole fixture tree) and the `fixture_sha256` stamp. The mod fixture holds
  exactly one test per `*.test.ts` file.
- Write `probe.sh`, `check-evidence.sh` and `check-verdict.sh` (both checkers git-free; contracts in §4 below;
  the verb lists and the live-name enumeration live in `check-evidence.sh`, not in files the probe writes),
  and run the **manifest-builder control** on `$SCRATCH/ctl-tree`: build, change one file's bytes (same path,
  same length), build again, plant a fixture identifier in a copy; record `evidence/negative-control-manifest.txt`.

Exit: env record, scrub list, resolved routing, `version-start.txt`, protected-set manifest
(`protected-t0.sha256`) and builder control, isolation verdict per tool, the six observations, fixtures
generated and hashed, instruments written.

### M2 — Claude static measurement, scratch home (Priority: High)

- For every fixture (R01-R14, one single-component plugin each): `claude plugin validate <path>` plain
  (`claude-validate-<ID>`) and `--strict --json` (`claude-validate-strict-<ID>`); wrap the fixtures in a
  local marketplace; `claude plugin marketplace add`; `claude plugin install <name>@<marketplace>`
  (`claude-install-<ID>`); `claude plugin details` (`claude-details-<ID>`) for every fixture that installed;
  `claude plugin test` (`claude-test-<ID>`) for every fixture holding `*.test.ts` (the mod fixture only),
  recording the runner's own executed-test count. All with `home=scratch`.
- Record the R11 corroborating line: the verbatim `CLAUDE.md at the plugin root is not loaded as project
  context ...` message from `claude-validate-R11`; it decides no cell (REQ-009 stays: no cell is inferred from
  validate output alone).
- The static rows R05, R12, R13, R14 are decided here: their Claude cell is the static judgment of REQ-009
  (`PLUGIN-OK` only where `details` lists the component, otherwise `PARTIAL(static-only: ...;
  runtime=UNOBSERVED(no-runtime-channel))`).
- Read, read-only, one or two real installed dual-manifest plugins (layout reference only; they sit under the
  profile `plugins/synced/` and `~/.codex/plugins/cache/moai-cowork/`).
- Wrap the whole static phase in one before/after manifest pair of the real profile (`static-window`).

### M3 — Claude runtime observation, real profile, observe-only (Priority: High)

Depends on the M1 authentication and routing record. Every command here carries `home=real`, runs only as a
`claude -p --no-session-persistence` session (a plugin run with `--plugin-dir`, or a control run without it)
or a read-only verb, and is bracketed by a before/after manifest pair. The ten runtime rows are R01-R04 and
R06-R11; R05, R12, R13 and R14 have no runtime channel and run nothing here.

- Scratch project with a project-tree **control** for each channel: marker-file hook in
  `.claude/settings.json`, sentinel rule in `.claude/rules/`, sentinel `CLAUDE.md`, `.mcp.json`. No `git init`.
  Every marker line carries a per-run token so ambient plugin output cannot be mistaken for a sentinel, and
  every runtime raw file carries the generic lines `PLUGIN-SENTINEL: <PRESENT|ABSENT> <sentinel>`,
  `PROJECT-CONTROL-SENTINEL: <PRESENT|ABSENT> <control sentinel>` and `CONTROL-FIRED: <value with the run
  token>`; the R04 and R08 raw files carry these in addition to their own token families.
- Prefer mechanical channels: hook-written marker files, and, if M1 observed it, the `stream-json` init-event
  listing for R01, R02, R03, R06. Model-quoted sentinel channels (R09-R11) run **three times**; an absence
  claim needs 3 of 3 agreeing runs, otherwise the cell is `PARTIAL`. Run cap printed by `probe.sh` as
  `RUNTIME-RUN-CAP=60` (10 runtime rows x 2 x 3, counted from the `channels` column of spec §5); a cell the
  cap would starve is `UNOBSERVED(cost-cap; raw; quote)`.
- **R08**: copy the template's SessionStart entry verbatim (the entry whose script argument ends in
  `handle-session-start-navigator.sh` and whose `timeout` is 5; at tree 676293144 these are lines 14-19 of
  `internal/template/templates/.claude/settings.json.tmpl`, identified by content, with no template directive
  on those lines) into the fixture plugin `hooks/hooks.json` and the project `settings.json`, changing only the
  final args element: `${CLAUDE_PLUGIN_ROOT}/hooks/marker.sh` in the plugin copy, `${CLAUDE_PROJECT_DIR}/marker.sh`
  in the project copy. `evidence/raw/r08-entry-diff.txt` shows each copy differs from the template entry in
  that element only. Each marker script prints `ARG0=$0` (the argument it received, not the environment
  variable) and a line `PLUGIN-HOOK-EXEC: PRESENT <token>` / `PROJECT-HOOK-EXEC: PRESENT <token>`; the probe
  writes `ABSENT` for a marker file that does not exist after the session. The plugin copy's marker also
  writes `START=<epoch>`, sleeps 7 seconds, then writes `END=<epoch>`; the probe records `SESSION-END=<epoch>`
  and `END=ABSENT` when no `END` line exists. The checker recomputes the expansion from `ARG0`, the cutoff
  (`END` absent and `SESSION-END - START > 5`) and the seconds (`END - START`, defined only when `END`
  exists). The project copy is the control: an `ABSENT` plugin line is a valid reading only when the project
  line is `PRESENT`.
- **R09/R10/R11**: a rules fixture inside the plugin, always-loaded and `paths:`-scoped; ask the session to
  quote the sentinel at start (`PHASE=start`) and again after touching a matching path (`PHASE=after-touch`);
  `PROJECT-ONLY` only with the project control observed present. R11 is **configuration-scoped**: the verdict
  states the recorded `instructionFiles` value. The isolated-default configuration cannot run an authenticated
  session on a scratch home; it is measured only if observation (6) of M1 finds a launch option that gives an
  authenticated session an isolated configuration without writing any real home, otherwise it is recorded as
  a Gap.
- **R04 (mods), concrete channel.** Fixture `hooks/hooks.json` = `{ "modules": ["./register.tsx"] }`;
  `hooks/register.tsx` writes a marker through `$.fs` on `session.start`; `hooks/register.test.ts` holds
  exactly one test, for `claude plugin test`. Load with `--plugin-dir` in `claude -p` (hot reload is off under
  `-p`, so load-at-start is what is tested). Control: the same marker logic as an ordinary plugin hook in the
  same session (R07 fixture), written as `HOOK-CONTROL: <token>`; the mod marker is `MOD-MARKER: <token>`.
  Result table, fixed here so it cannot be argued afterwards: control fired and mod marker present,
  `PLUGIN-OK` allowed; control fired and mod marker absent, the R04 runtime result is
  `UNOBSERVED(modules-not-loaded-under-claude-p; raw; quote)` and the Claude R04 cell is at most
  `PARTIAL(static-only: validate and test results)`; control did not fire, `UNOBSERVED(control-not-fired)`.
  An interactive or SDK session that does load modules cannot be driven by the probe (no terminal) and is
  named in the card's Gaps.

### M4 — Codex measurement (Priority: Medium)

- `.codex-plugin` fixtures per Codex-applicable row, and a Codex marketplace root with
  `.agents/plugins/marketplace.json` (the observed Codex layout, e.g.
  `~/.codex/.tmp/bundled-marketplaces/openai-bundled/.agents/plugins/marketplace.json`, read-only reference).
  The exact schema is discovered by rejection messages: each failed `codex plugin marketplace add` output is
  saved as `evidence/raw/codex-marketplace-schema-<n>.txt` and the manifest is corrected until it is accepted
  or the attempts are exhausted, in which case the Codex cells are `UNOBSERVED` with the last message quoted.
- `codex plugin marketplace add <path>`, `codex plugin add`, `codex plugin list` under scratch `CODEX_HOME`
  only; a fixture counts as added only when `plugin add` exited 0 and `codex-plugin-list` names it. There is
  no `codex plugin validate`, so static evidence is install and list output.
- Runtime: no `codex exec` session runs against a real Codex home (not on the REQ-004 allowlist), and a
  scratch `CODEX_HOME` has no login; by default Codex runtime cells are therefore `UNOBSERVED` with the
  `codex plugin --help` and `codex exec --help` output quoted as the proof, unless M4 finds a route that
  writes no registry and copies no credential. Project-tree controls (`.codex/config.toml`,
  `.codex/hooks.json`, `AGENTS.md`, `.agents/skills`) are still written for the cells that do get a route.
- Rows with no Codex counterpart: a documented search (schema probe by rejection messages), then
  `UNOBSERVED(no-equivalent-surface-found; raw; quote)` only when the search output shows nothing.

### M5 — Verdict synthesis and durable carrier (Priority: High)

- Scaffold the 28 evidence cards (`evidence/cells/R01-claude.md` through `R14-codex.md`) from the template
  (Claim, Evidence with `DECIDING-LINE: "<quote>"` and, on the Claude cards of the seven floor rows, `STATIC-LINE: "<quote>"`,
  Baseline-attribution, Gaps, Residual-risk), fill them, then write `verdict.md` with the §5 table, the
  recommended-home and consequence columns, the stamps, the version agreement, the `LEAK-FINDING:` lines (if
  any) and the non-goal line. The Residual-risk of the verdict states that the `check-evidence.sh` modes were
  not observed failing (§6).
- Run `check-verdict.sh cells`, `check-verdict.sh table`, and each AC checker on the real evidence; build the
  seven base inputs and the 40 mutant inputs of §6 under `evidence/mutants/` and run every base and mutant to
  observe the bases pass and the mutants fail; record `evidence/negative-control-checker.txt`.
- Take the final `--version` readings into `evidence/version-end.txt` and compare with `version-start.txt`.
- Write the **durable carrier**: the 14-row table, a line `verdict_sha256: <sha256 of verdict.md>`, and the
  evidence index (`evidence index:` followed by `RAW-FILES=<n>` and the file names), then run
  `check-evidence.sh stamps`. The tracked write itself (the `§E.2` block, the `§E.3` signal, the `spec.md`
  `status:` flip) is delegated to manager-develop in one non-cycle evidence-writing delegation (§7), which
  receives the verified table and index as input and writes nothing else.

## 3. Fixture Layout (scratch, never committed)

```
$SCRATCH/                      unique dir under the session scratchpad
  config-claude/               CLAUDE_CONFIG_DIR for the static route (starts empty)
  config-codex/                CODEX_HOME for the static route (starts empty)
  ctl-tree/                    manifest-builder control tree (3 small files)
  project/                     fresh scratch project, NOT a git repository
    .claude/{settings.json,rules/,skills/}      project-tree controls
    CLAUDE.md  AGENTS.md  .mcp.json  .codex/
  marketplace/
    .claude-plugin/marketplace.json             lists every fixture plugin
    plugins/
      p-skill/ p-agent/ p-command/ p-mod/ p-outstyle/ p-mcp/
      p-hook/ p-sethook/ p-rules/ p-pathrules/ p-claudemd/
      p-pluginsettings/ p-userconfig/ p-manifest/
        .claude-plugin/plugin.json  .codex-plugin/plugin.json
        <component dirs>  with one SENTINEL_<ID>_<random> each
  marketplace-codex/
    .agents/plugins/marketplace.json            Codex marketplace manifest (schema found by rejection)
  evidence-staging/            raw outputs before copy
```

Fourteen plugin directories: R01-R14 map one to one onto them (no composite). Each fixture carries its
sentinel `SENTINEL_<ID>_<random>` where a listing or a marker prints it (a skill, agent or command
description, an MCP server name, the hook marker text, a rule body, the mod marker, a manifest field). The
`fixture_sha256` stamped into every raw file is the sha256 of `evidence/fixture-manifest.txt`, the sorted
`sha256  relpath` listing of `$SCRATCH/marketplace`, `$SCRATCH/marketplace-codex` and the project controls.
The scratch tree is kept until M5 ends so the checker can re-hash it; its removal is the leader's.

Everything under `.moai/reports/t1434/` is **local-only** (gitignored). The only tracked writes are the ones
named in §1 and §7.

## 4. Evidence Files, Command Line Format, and Instrument Contracts

**`CMD:` line** (one per measured command, appended to `evidence/commands.log`):

```
CMD: id=<n> home=<scratch|real> unset <NAME1> <NAME2> ... && [CLAUDE_CONFIG_DIR=<scratch>/config-claude | CODEX_HOME=<scratch>/config-codex] timeout <secs> <command ...>
```

A `home=scratch` line carries the override for its tool; a `home=real` line carries none and is either a
`claude -p --no-session-persistence` session (with `--plugin-dir` for a plugin run, without it for a control
run) or one of the read-only verbs. The forbidden-verb list (the registry- and settings-writing verbs of
REQ-004) and the read-only list are **literal text in `check-evidence.sh`**, copied from spec REQ-004; the
probe writes no verb list. `probe.sh` increments a counter per command and prints it as `CMD-COUNT=<n>` in
`evidence/probe-counts.txt`, together with `FIXTURES=`, `ROWS=` (rows counted from §5 of `spec.md`),
`REAL-CMDS=`, `ROWS-MANIFEST=` and `RUNTIME-RUN-CAP=`.

**Raw files** under `evidence/raw/`: `claude-validate-<ID>`, `claude-validate-strict-<ID>`,
`claude-install-<ID>`, `claude-details-<ID>`, `claude-test-<ID>`, `claude-runtime-<ID>` (R01-R04, R06-R11),
`claude-isolation-probe`, `codex-isolation-probe`, `codex-marketplace-add`, `codex-marketplace-schema-<n>`,
`codex-plugin-add-<ID>`, `codex-plugin-list`, `codex-exec-<ID>` (only if a route exists), `r08-entry-diff`
(all `.txt`). Each begins with `# stamp: claude=<v> codex=<v> fixture_sha256=<h> date=<d>` and
`# CMD-ID: <n>` and ends with `EXIT=<code>`. A line beginning `#` and an `EXIT=` trailer are never a
deciding line.

**Row tokens.** A quote that decides a non-`UNOBSERVED` Claude cell (the card's `DECIDING-LINE` and, on a
floor row, its `STATIC-LINE`) must contain at least one token of its row: the sentinel prefix
`SENTINEL_<ID>_`, the fixture plugin name, or the details token. The `STATIC-LINE` is checked as a fixed
string in that row's own `claude-validate-<ID>`, `claude-validate-strict-<ID>`, `claude-details-<ID>` or
`claude-install-<ID>` file. On a floor row whose only static evidence is validate or install output, the row
token is the fixture plugin name (the verbatim `Successfully installed plugin <name>@...` line carries it).

| ID | Fixture plugin | Details token (where `details` has a category) |
|----|----------------|-----------------------------------------------|
| R01 | `p-skill` | `Skills (1)` |
| R02 | `p-agent` | `Agents (1)` |
| R03 | `p-command` | (none: `details` has no commands category) |
| R04 | `p-mod` | (none) |
| R05 | `p-outstyle` | (none: `details` has no output-styles category) |
| R06 | `p-mcp` | `MCP servers (1)` |
| R07 | `p-hook` | `Hooks (1)` |
| R08 | `p-sethook` | (none) |
| R09 | `p-rules` | (none) |
| R10 | `p-pathrules` | (none) |
| R11 | `p-claudemd` | (none) |
| R12 | `p-pluginsettings` | (none) |
| R13 | `p-userconfig` | (none) |
| R14 | `p-manifest` | (none) |

**Checkers** (git-free, print counters, exit non-zero on any failure and on an empty sweep). `check-evidence.sh`
is a **measurement helper**: its modes carry no mutants of their own, and a wrong reading there is caught by
the cell-level cross-checks of `check-verdict.sh` and by the red input each criterion states. `check-verdict.sh`
decides the deliverable and every one of its counters carries a mutant (§6).

| Script and mode | Prints | Reads |
|-----------------|--------|-------|
| `check-evidence.sh commands` | `CMDS SCRUB-MATCH ENV-NAMES-MATCH CMD-COUNT-MATCH SECRET-VALUES RESULT` | `env.txt`, `env-scrub.txt`, `commands.log`, `probe-counts.txt`, its own environment names |
| `check-evidence.sh isolation` | `TOOLS ISOLATED FAILED BEFORE-EMPTY AFTER-NONEMPTY CRED-FILES RESULT` | `scratch-home-*.txt` |
| `check-evidence.sh observe-only` | `REAL-CMDS FORBIDDEN-REAL REAL-OFF-ALLOWLIST SCRATCH-NO-OVERRIDE REAL-WITH-OVERRIDE FORBIDDEN-VERBS READONLY-VERBS RESULT` | `commands.log`, its own literal lists |
| `check-evidence.sh manifests` | `ROWS REQUIRED-ROWS PAIRS REAL-CMDS LEAK LEAK-FORBIDDEN LEAK-UNREPORTED LEAK-CONTINUED EXCLUDED-DECLARED BUILDER-CONTROL RESULT` | `protected-*`, `manifest/`, `negative-control-manifest.txt`, `commands.log`, `verdict.md` |
| `check-evidence.sh static` | `ROWS FIXTURES VALIDATE-PLAIN VALIDATE-STRICT INSTALL-FILES INSTALL-OK DETAILS TEST-FIXTURES TESTS-MATCH RESULT` | raw files, `fixture-manifest.txt`, spec §5 |
| `check-evidence.sh stamps` | `RAW STAMPED VERSION-EQUAL SHA-RECOMPUTED NONGOAL CARRIER-ROWS CARRIER-SHA RESULT` | raw files, version files, `verdict.md`, `progress.md` |
| `check-evidence.sh negative-controls` | `BASES-GREEN MUTANTS MUTANTS-RED RESULT` | `evidence/mutants/**` and the names in §6, re-executed |
| `check-verdict.sh cells <file>` | `CELLS CARDLESS MISSING-FIELDS ORPHAN-RAW QUOTE-MISMATCH EMPTY-REASON RESULT` | `verdict.md`, cards, raw files, §4 row tokens |
| `check-verdict.sh runtime <file>` | `RUNTIME-ROWS SWEPT NOT-SWEPT-UNOBSERVED BAD-CONTROL BAD-RUNS RUNS RUNTIME-RUN-CAP OVER-CAP RESULT` | `verdict.md`, `claude-runtime-*`, `commands.log` |
| `check-verdict.sh table <file>` | `ROWS HEADER VOCAB-BAD HOME-BAD INCONSISTENT CONSEQ-BAD FLOOR FLOOR-MET RESULT` | `verdict.md`, `spec.md` §5, cards |
| `check-verdict.sh r08 <file>` | `EXEC-BAD EXPANSION-BAD EPOCH-BAD ENTRY-DIFF-BAD` and the recomputed `EXPANSION`, `TIMEOUT-CUTOFF`, `SECONDS` | `claude-runtime-R08.txt`, `r08-entry-diff.txt`, `env.txt` |
| `check-verdict.sh r04 <file>` | `TESTS-EXECUTED TESTS-FOUND OUTCOME TESTS-GAP PAYLOAD-BAD CELL-MISMATCH` | `claude-test-R04.txt`, `claude-runtime-R04.txt`, `fixture-manifest.txt`, `verdict.md` |
| `check-verdict.sh rules <file>` | `RUNS-BAD CONTROL-BAD CONFIG-SCOPE-BAD VALIDATE-MSG-BAD` | `claude-runtime-R09/R10/R11.txt`, R11 card, `env.txt`, `claude-validate-R11.txt` |
| `check-verdict.sh codex <file>` | `ADDS-OK LISTED ADD-MISMATCH VALIDATE-LINES REAL-CODEX-WRITES` | `commands.log`, `codex-*` raw files, `verdict.md` |

`FLOOR` is not a typed number: the checker counts the rows marked `floor=yes` in §5 of `spec.md` (seven:
R01, R02, R03, R05, R06, R07, R14). A card's deciding line is the field `DECIDING-LINE: "<quote>"` inside
its Evidence field; the checker confirms the quote as a fixed-string match in the file the card cites.

## 5. Risks and Mitigations

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| The "real home" is mis-identified (this lane runs on a profile, not `~/.claude`) | High | Protection watches the wrong tree | M1 records `$CLAUDE_CONFIG_DIR`, `$CODEX_HOME`, auth status first; protected set derived from the resolved homes; `~/.claude`, `~/.codex` as secondary watches |
| Runtime needs the login a scratch home lacks | High | Runtime rows blocked | Runtime runs on the real profile observe-only (`claude -p` plugin and control sessions); a blocked cell is `UNOBSERVED` with the quoted proof |
| Real profile drifts or is written concurrently by other lanes | High | A diff that is not the probe's | Per-command before/after manifests with `AMBIENT`/`LEAK` attribution by fixture identifiers; nothing restored; `AMBIENT` is not read as "another lane wrote it" |
| A probe command writes a real registry despite the plan | Low | Global state change | Forbidden-verb list lives in the checker and is audited from `commands.log` (AC-004); `home=real` lines restricted to `claude -p` sessions or read-only verbs; first `LEAK` stops the probe's real-home commands; probe script itself is guard-unmediated (Gaps) |
| Nested `claude -p` joins messaging or loads the profile's plugins and MCP servers | High | Contaminated observations | Scrub of every live `CLAUDE_CODE_*` name; per-run token on marker lines; M1 observes the launch options; contamination sources listed in the verdict Gaps |
| Project `.mcp.json` or hooks need non-interactive approval, control does not fire | Medium | False negative | A control that does not fire makes the channel `UNOBSERVED(control-not-fired)`, never a negative |
| Model-quoted sentinel channels are nondeterministic | Medium | Wrong absence claim | Three runs, 3 of 3 agree for an absence claim, otherwise `PARTIAL`; mechanical channels preferred |
| `claude -p` does not load mods | Medium | R04 runtime blocked | Result table fixed in M3; static-only `PARTIAL` with the control as proof |
| Codex ignores `CODEX_HOME` or has no marketplace manifest we can satisfy | Medium | Codex cells blocked | `ISOLATION-FAILED` route runs no writing command; rejection messages saved; cells `UNOBSERVED` with quote |
| CLI version changes mid-run (auto-update) | Medium | Cells stamped inconsistently | Start and end version readings must be equal (AC-015); mismatch invalidates the batch |
| A hook sleeping past its timeout, or a prompt, hangs the probe | Medium | Stuck run | External `timeout` wrapper on every runtime command |
| Run-phase evidence is gitignored and never reaches the remote | Certain | Follow-up cards cannot read it | Durable carrier in `progress.md §E.2` with `verdict_sha256` and an evidence index (AC-015) |
| Per-run model cost | Medium | Spend | Run cap 60 = 10 x 2 x 3, printed and pinned by AC-007; starved cells are `UNOBSERVED(cost-cap)` |
| A raw file is fabricated by hand | Low | Vacuous green | Not defended mechanically; named in Residual-risk |
| A `check-evidence.sh` mode is wrong and reads green | Low | A measurement helper misreports | Not observed failing (no mutants by decision); caught by the cell-level cross-checks and the stated red input of each criterion; named in Residual-risk |

## 6. Verification Approach

Each acceptance criterion in `acceptance.md` is a plain command (`sh .moai/reports/t1434/check-*.sh <mode>`
or a git form) plus the counters it must print. **One place defines which modes carry mutants: this
section.** Every counter of every `check-verdict.sh` mode carries a mutant, and each mode also carries a
`ZERO-SWEEP` mutant and a base input that must pass (`.claude/rules/moai/development/verification-completeness.md`
§1.1, §1.2): the manifest-builder control and these 40 mutants are observed failing and are re-executable.
The `check-evidence.sh` modes carry no mutants (see §4); they are measurement helpers.

| Mode of `check-verdict.sh` | Mutant names (one mutant per name; each mutant differs from the mode's base input by that single defect) | Names |
|----------------------------|-------------------------------------------------------------------------------------------------------------|-------|
| cells | `cells:CELLS` `cells:CARDLESS` `cells:MISSING-FIELDS` `cells:ORPHAN-RAW` `cells:QUOTE-MISMATCH` `cells:QUOTE-GENERIC` `cells:EMPTY-REASON` `cells:ZERO-SWEEP` | 8 |
| runtime | `runtime:SWEPT` `runtime:NOT-SWEPT-UNOBSERVED` `runtime:BAD-CONTROL` `runtime:BAD-RUNS` `runtime:OVER-CAP` `runtime:ZERO-SWEEP` | 6 |
| table | `table:ROWS` `table:HEADER` `table:VOCAB-BAD` `table:HOME-BAD` `table:INCONSISTENT` `table:CONSEQ-BAD` `table:FLOOR-MET` `table:ZERO-SWEEP` | 8 |
| r08 | `r08:EXEC-BAD` `r08:EXPANSION-BAD` `r08:EPOCH-BAD` `r08:ENTRY-DIFF-BAD` `r08:ZERO-SWEEP` | 5 |
| r04 | `r04:TESTS-GAP` `r04:PAYLOAD-BAD` `r04:CELL-MISMATCH` `r04:ZERO-SWEEP` | 4 |
| rules | `rules:RUNS-BAD` `rules:CONTROL-BAD` `rules:CONFIG-SCOPE-BAD` `rules:VALIDATE-MSG-BAD` `rules:ZERO-SWEEP` | 5 |
| codex | `codex:ADD-MISMATCH` `codex:VALIDATE-LINES` `codex:REAL-CODEX-WRITES` `codex:ZERO-SWEEP` | 4 |

Total 40 names (8 + 6 + 8 + 5 + 4 + 5 + 4), seven modes, seven bases. `QUOTE-MISMATCH` has two mutants
because it has two branches: an altered quote, and a quote that is a generic line (a stamp line or an `EXIT=`
trailer, or a line naming no row token) standing for a component.

`acceptance.md` pins a keyed RED-now cell and a keyed green-path cell per criterion (§2), each measured on
the committed revision `6d0d75af3`; the RED cells read run-phase files that are local-only, so the pin names
the committed tree and the measurement is of the worktree at that time (stated in the ledger). Continued
firing (§1.3): the checkers print their swept counts, so an instrument that stops sweeping prints zero and
fails.

## 7. Hand-off Notes for the Run Phase

- Run phase is a lane-authored measurement. The orchestrator runs M1-M5 inline in the card worktree for the
  local-only evidence (`probe.sh`, the checkers, the cards), which are not phase-owned artifacts. The one
  tracked write is **delegated to manager-develop** (non-cycle, evidence-writing; it writes no product code and
  no test): the `progress.md §E.2` block (the verified 14-row table and evidence index), the `§E.3` run-phase
  audit-ready signal, and the `status: draft -> in-progress` flip with `updated:` in `spec.md`, all in the
  one run-phase commit, per the ownership matrix of `spec-frontmatter-schema.md`. No earlier run-phase commit
  exists (M1-M4 write only ignored files), so this commit is the first run-phase commit. Nothing else under
  `.moai/specs/` changes during run, and nothing outside `.moai/reports/t1434/` and those lines is written
  in the repository. AC-001 reads its base from `run_start_sha` in `progress.md §E.1`.
- **Blocker contract.** A blocked floor row (no non-`UNOBSERVED` Claude cell) or an all-runtime-blocked
  outcome is reported by the lane's blocker report to the leader; the card then reports
  `blocked-on-measurement` with the partial table and the evidence path, and the leader chooses
  PASS-with-debt, re-plan or abandon. The criteria are not relaxed.
- **LEAK contract.** On the first `LEAK` the probe runs no further real-home command and records the
  attributing command. A `LEAK` from a permitted verb is a measured finding written into the verdict Gaps
  (`LEAK-FINDING: <cmd-id>`); a `LEAK` from a forbidden verb fails the SPEC.
- Anything that cannot be observed is recorded as `UNOBSERVED` with a raw file and a quoted line (a static
  row's missing runtime half is the one exception, REQ-009); no cell is promoted by documentation or by
  passing `validate` alone.
