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
  Only `--plugin-dir` sessions and read-only commands run there. Install, uninstall, marketplace
  add/remove/update, enable/disable, configure, plugin update, Codex plugin add/remove, and any other
  registry- or settings-writing command are **forbidden against any real home**.

Contamination sources of the real-profile route are named, not hidden (spec §4): the profile's
`enabledPlugins` (gopls-lsp, vercel, typesafe), `pluginConfigs."agents-md@builtin".options.instructionFiles`,
account-synced plugins under `plugins/synced/` and their MCP servers, `remoteControlAtStartup: true`,
inherited `CLAUDE_CODE_*` variables and the messaging socket/token, project `.mcp.json`/hook approval in a
non-interactive session, and per-run model cost.

Methodology is measurement/analysis (no `cycle_type` code implementation). Run-phase artifacts live under
`.moai/reports/t1434/` and are **local-only** (`.gitignore:235` is `.moai/reports/*`; `git check-ignore -v`
and `git status --porcelain --ignored` observed this iteration). The **durable carrier** the follow-up design
cards read is the 14-row verdict table plus an evidence index written into this SPEC's `progress.md §E.2`
at the end of the run phase (REQ-001, REQ-016). That is the one tracked write the run makes.

Tier M. Ordering follows the card (fixture design, Claude static, Claude runtime, Codex, synthesis) and,
inside the plan, reversibility: M1 settles the config-home routing, the protected set and the checker
instruments, which are the decisions most likely to change; everything after depends on them.

**Proportionality decision (Agent Core Behaviors 4 and 5).** The deliverable is a 14-row table, but each
row holds two independently verdicted cells that a follow-up card cites as `[cell:<ID>-<tool>]`. Per-cell
evidence cards are therefore kept (28 cards, one per cell), authored from a single five-field template that
`probe.sh` scaffolds, because a per-row block would put two independent claims under one citation. Machinery
criteria were merged instead (manifest builder, per-command attribution and excluded-set into one criterion;
both checkers' negative controls into one), which holds both budgets at 16/16 and keeps Tier M.

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
  `remoteControlAtStartup`.
- Build `evidence/env-scrub.txt` from the live names (not from a fixed list); build the single
  `unset <names> &&` prefix from it. `CLAUDE_CONFIG_DIR` and `CODEX_HOME` are never scrubbed: they are set
  to the scratch home (static route) or left as resolved (real route). If scrubbing a name breaks
  authentication on the real route, keep that name, list it with the observed reason, and re-run.
- Define the **protected set** from the resolved homes, with `~/.claude` and `~/.codex` as secondary
  read-only watches when they differ. Rows (named so the checker can require them):
  `claude:installed_plugins.json`, `claude:known_marketplaces.json`, `claude:settings.json`,
  `claude:plugins/cache`, `claude:plugins/marketplaces`, `claude:plugins/data`, `claude:plugins/synced`
  (recursive file list plus hash), and `codex:config.toml#plugins-tables` (the `[plugins]` and marketplace
  tables extracted from `config.toml`, hashed as text), `codex:plugins-tree`.
  `evidence/protected-excluded.txt` names, each with a reason, `.claude.json` (rewritten by any session
  start), `sessions/`, `session-env/`, `debug/`, `file-history/`, `projects/` transcripts,
  `.codex-global-state.json`, Codex sqlite state.
- Write the **manifest builder** (inside `probe.sh`, git-free). Volatility is **not** classified from two
  back-to-back manifests (the profile drift is intermittent). Instead the probe takes a manifest immediately
  before and after each real-home command (`evidence/manifest/<cmd-id>-before.sha256`, `-after.sha256`,
  `manifest-diff-<cmd-id>.txt`) and once around the whole static phase (`static-window`), and classifies each
  changed row `AMBIENT` or `LEAK` by whether its changed content names a fixture plugin (`p-*`), the scratch
  marketplace, the scratch path, or a `SENTINEL_` string. Nothing is restored; nothing the probe did not
  create is touched.
- **Static isolation proof:** list each scratch home before (must be empty) and after the first
  registry-writing command (`evidence/scratch-home-before.txt`, `scratch-home-after.txt`, per tool). If a
  tool does not honor its variable the after-listing stays empty: write `ISOLATION-FAILED <tool> raw=<file>`
  with the proof, run no registry-writing command for that tool, and carry its static cells as `UNOBSERVED`.
- **Observations the plan does not assume** (each recorded in a raw file, none asserted here):
  (1) `claude plugin test` on an empty mod folder (the empty-sweep analogue: its output and exit code);
  (2) whether `claude -p --output-format stream-json --verbose` lists loaded skills, agents, plugins or MCP
  servers in its init event (a candidate deterministic channel, run once with `--plugin-dir` on one fixture);
  (3) whether `claude -p` loads `--plugin-dir` mod modules; (4) what `CLAUDE_CODE_PLUGIN_DIRS` does, or
  `UNOBSERVED(plugin-dirs-env-unmeasured)`; (5) `claude plugin details` with a bare name.
- Write `probe.sh`, `check-evidence.sh` and `check-verdict.sh` (both checkers git-free; contracts in §4 below),
  and run the **manifest-builder control** on `$SCRATCH/ctl-tree`: build, change one file's bytes (same path,
  same length), build again, plant a fixture identifier in a copy; record `evidence/negative-control-manifest.txt`.

Exit: env record, scrub list, resolved routing, protected-set manifests and builder control, isolation
verdict per tool, the five observations, fixtures generated and hashed, instruments written.

### M2 — Claude static measurement, scratch home (Priority: High)

- For every fixture (R01-R14 single-component plus composite): `claude plugin validate <path>` plain
  (`claude-validate-<ID>`) and `--strict --json` (`claude-validate-strict-<ID>`); wrap the fixtures in a
  local marketplace; `claude plugin marketplace add`; `claude plugin install <name>@<marketplace>`
  (`claude-install-<ID>`); `claude plugin details` (`claude-details-<ID>`) for every fixture that installed;
  `claude plugin test` (`claude-test-<ID>`) for every fixture holding `*.test.ts`, recording the runner's own
  executed-test count. All with `home=scratch`.
- Record the R11 corroborating line: the verbatim `CLAUDE.md at the plugin root is not loaded as project
  context ...` message from `claude-validate-R11`; it decides no cell (REQ-009 stays: no cell is inferred from
  validate output alone).
- Read, read-only, one or two real installed dual-manifest plugins (layout reference only; they sit under the
  profile `plugins/synced/` and `~/.codex/plugins/cache/moai-cowork/`).
- Wrap the whole static phase in one before/after manifest pair of the real profile (`static-window`).

### M3 — Claude runtime observation, real profile, observe-only (Priority: High)

Depends on the M1 authentication and routing record. Every command here carries `home=real`, runs only as a
`--plugin-dir` session (or a read-only verb), and is bracketed by a before/after manifest pair.

- Scratch project with a project-tree **control** for each channel: marker-file hook in
  `.claude/settings.json`, sentinel rule in `.claude/rules/`, sentinel `CLAUDE.md`, `.mcp.json`. No `git init`.
  Every marker line carries a per-run token so ambient plugin output cannot be mistaken for a sentinel.
- Prefer mechanical channels: hook-written marker files, and, if M1 observed it, the `stream-json` init-event
  listing for R01, R02, R03, R06. Model-quoted sentinel channels (R09-R11) run **three times**; an absence
  claim needs 3 of 3 agreeing runs, otherwise the cell is `PARTIAL`. Run cap printed by `probe.sh` as
  `RUNTIME-RUN-CAP=<13 x 2 x 3>`; a cell the cap would starve is `UNOBSERVED(cost-cap; raw; quote)`.
- **R08**: copy the template's SessionStart entry verbatim (the entry whose script argument ends in
  `handle-session-start-navigator.sh` and whose `timeout` is 5; at tree 676293144 these are lines 14-19 of
  `internal/template/templates/.claude/settings.json.tmpl`, identified by content, with no template directive
  on those lines) into the fixture plugin `hooks/hooks.json` and the project `settings.json`, changing only the
  final args element to a marker script. `evidence/raw/r08-entry-diff.txt` shows the template entry and the
  fixture entry differ in that element only. The marker script writes `PLUGIN-HOOK-EXEC: <token>` /
  `PROJECT-HOOK-EXEC: <token>`, the expanded `CLAUDE_PROJECT_DIR` and `CLAUDE_PLUGIN_ROOT` values, and for the
  timeout case `START=<epoch>`, a 7-second sleep, then `END=<epoch>`; the probe records `SESSION-END=<epoch>`.
  The cutoff and the observed seconds are recomputed by the checker from the epochs.
- **R09/R10/R11**: a rules fixture inside the plugin, always-loaded and `paths:`-scoped; ask the session to
  quote the sentinel at start (`PHASE=start`) and again after touching a matching path (`PHASE=after-touch`);
  `PROJECT-ONLY` only with the project control observed present. R11 is **configuration-scoped**: the verdict
  states the recorded `instructionFiles` value. The isolated-default configuration cannot run an authenticated
  session on a scratch home; it is measured only if M1 finds a way to run an authenticated session with an
  overriding settings layer that writes no real home, otherwise it is recorded as a Gap.
- **R04 (mods), concrete channel.** Fixture `hooks/hooks.json` = `{ "modules": ["./register.tsx"] }`;
  `hooks/register.tsx` writes a marker through `$.fs` on `session.start`; `hooks/register.test.ts` exists for
  `claude plugin test`. Load with `--plugin-dir` in `claude -p` (hot reload is off under `-p`, so load-at-start
  is what is tested). Control: the same marker logic as an ordinary plugin hook in the same session
  (R07 fixture). Result table, fixed here so it cannot be argued afterwards: control fired and mod marker
  present, `PLUGIN-OK` allowed; control fired and mod marker absent, the R04 runtime result is
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
- Runtime: `codex exec` runs only if M4 finds a way to load a plugin that writes no registry in the real
  Codex home; the observed `codex plugin --help` lists no such flag, so by default Codex runtime cells are
  `UNOBSERVED` with the help output quoted as the proof. Project-tree controls (`.codex/config.toml`,
  `.codex/hooks.json`, `AGENTS.md`, `.agents/skills`) are still written for the cells that do get a route.
- Rows with no Codex counterpart: a documented search (schema probe by rejection messages), then
  `UNOBSERVED(no-equivalent-surface-found; raw; quote)` only when the search output shows nothing.

### M5 — Verdict synthesis and durable carrier (Priority: High)

- Scaffold the 28 evidence cards (`evidence/cells/R01-claude.md` through `R14-codex.md`) from the five-field
  template, fill them, then write `verdict.md` with the §5 table, the recommended-home and consequence columns,
  the stamps, the version agreement and the non-goal line.
- Run `check-verdict.sh cells`, `check-verdict.sh table`, and each AC checker on the real evidence; build the
  mutant inputs under `evidence/mutants/` and run every mutant to observe it fail; record
  `evidence/negative-control-checker.txt` and `evidence/negative-control-evidence.txt`.
- Take the final `--version` readings into `evidence/version-end.txt` and compare with `version-start.txt`.
- Write the **durable carrier**: the 14-row table, a line `verdict_sha256: <sha256 of verdict.md>`, and the
  evidence index (`evidence index:` followed by `RAW-FILES=<n>` and the file names) into `progress.md §E.2`.
  Then run `check-evidence.sh stamps`.

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
      p-hook/ p-sethook/ p-rules/ p-rules-paths/ p-claudemd/
      p-pluginsettings/ p-userconfig/ p-manifest/ p-composite/
        .claude-plugin/plugin.json  .codex-plugin/plugin.json
        <component dirs>  with one SENTINEL_<kind>_<random> each
  marketplace-codex/
    .agents/plugins/marketplace.json            Codex marketplace manifest (schema found by rejection)
  evidence-staging/            raw outputs before copy
```

Fifteen plugin directories: R01-R14 map one to one onto fourteen of them, plus `p-composite`.
`fixture_sha256` is the sha256 of `evidence/fixture-manifest.txt`, the sorted `sha256  relpath` listing of
`$SCRATCH/marketplace`, `$SCRATCH/marketplace-codex` and the project controls; it is stamped into every raw
file. The scratch tree is kept until M5 ends so the checker can re-hash it; its removal is the leader's.

Everything under `.moai/reports/t1434/` is **local-only** (gitignored). The only tracked write is the
`progress.md §E.2` block of M5.

## 4. Evidence Files, Command Line Format, and Instrument Contracts

**`CMD:` line** (one per measured command, appended to `evidence/commands.log`):

```
CMD: id=<n> home=<scratch|real> unset <NAME1> <NAME2> ... && [CLAUDE_CONFIG_DIR=<scratch>/config-claude | CODEX_HOME=<scratch>/config-codex] timeout <secs> <command ...>
```

A `home=scratch` line carries the override for its tool; a `home=real` line carries none and is either a
`--plugin-dir` session or a read-only verb from `evidence/readonly-verbs.txt`. `evidence/forbidden-verbs.txt`
lists the registry- and settings-writing verbs of REQ-004. Both lists are written by `probe.sh` from the spec
text. `probe.sh` increments a counter per command and prints it as `CMD-COUNT=<n>` in
`evidence/probe-counts.txt`, together with `FIXTURES=`, `ROWS=` (rows counted from §5 of `spec.md`),
`REAL-CMDS=`, `ROWS-MANIFEST=` and `RUNTIME-RUN-CAP=`.

**Raw files** under `evidence/raw/`: `claude-validate-<ID>`, `claude-validate-strict-<ID>`,
`claude-install-<ID>`, `claude-details-<ID>`, `claude-test-<ID>`, `claude-runtime-<ID>` (R01-R13),
`codex-marketplace-add`, `codex-marketplace-schema-<n>`, `codex-plugin-add-<ID>`, `codex-plugin-list`,
`codex-exec-<ID>`, `r08-entry-diff` (all `.txt`). Each begins with
`# stamp: claude=<v> codex=<v> fixture_sha256=<h> date=<d>` and `# CMD-ID: <n>` and ends with `EXIT=<code>`.

**Checkers** (git-free, print counters, exit non-zero on any failure and on an empty sweep):

| Script and mode | Prints | Reads |
|-----------------|--------|-------|
| `check-evidence.sh commands` | `CMDS SCRUB-MATCH ENV-NAMES-MATCH CMD-COUNT-MATCH SECRET-VALUES RESULT` | `env.txt`, `env-scrub.txt`, `commands.log`, `probe-counts.txt` |
| `check-evidence.sh isolation` | `TOOLS ISOLATED FAILED BEFORE-EMPTY AFTER-NONEMPTY CRED-FILES RESULT` | `scratch-home-*.txt` |
| `check-evidence.sh observe-only` | `REAL-CMDS FORBIDDEN-REAL REAL-OFF-ALLOWLIST SCRATCH-NO-OVERRIDE REAL-WITH-OVERRIDE RESULT` | `commands.log`, verb lists |
| `check-evidence.sh manifests` | `ROWS REQUIRED-ROWS PAIRS REAL-CMDS LEAK EXCLUDED-DECLARED BUILDER-CONTROL RESULT` | `protected-*`, `manifest/`, `negative-control-manifest.txt` |
| `check-evidence.sh fixtures` | `FIXTURES ROWS SENTINELS SENTINELS-UNIQUE CLAUDE-MANIFESTS CODEX-MANIFESTS CODEX-MARKETPLACE SHA-RECOMPUTED RESULT` | `fixture-*.txt`, scratch tree |
| `check-evidence.sh static` | `FIXTURES VALIDATE-PLAIN VALIDATE-STRICT INSTALL-FILES INSTALL-OK DETAILS TEST-FIXTURES TESTS-MATCH RESULT` | raw files, `fixture-manifest.txt` |
| `check-evidence.sh stamps` | `RAW STAMPED VERSION-EQUAL SHA-RECOMPUTED NONGOAL CARRIER-ROWS CARRIER-SHA RESULT` | raw files, version files, `verdict.md`, `progress.md` |
| `check-evidence.sh negative-controls` | `MODES MUTANTS MUTANTS-RED RESULT` | `evidence/mutants/**`, re-executed |
| `check-verdict.sh cells <file>` | `CELLS CARDLESS MISSING-FIELDS ORPHAN-RAW QUOTE-MISMATCH EMPTY-REASON RESULT` | `verdict.md`, cards, raw files |
| `check-verdict.sh runtime <file>` | `RUNTIME-ROWS SWEPT NOT-SWEPT-UNOBSERVED BAD-CONTROL BAD-RUNS RESULT` | `verdict.md`, `claude-runtime-*` |
| `check-verdict.sh table <file>` | `ROWS HEADER VOCAB-BAD HOME-BAD INCONSISTENT CONSEQ-BAD FLOOR FLOOR-MET RESULT` | `verdict.md`, `spec.md` §5 |
| `check-verdict.sh r08`, `r04`, `rules`, `codex` | row-specific counters as named in the criteria | the row's raw files |

`FLOOR` is not a typed number: the checker counts the rows marked `floor=yes` in §5 of `spec.md`. A card's
deciding line is the field `DECIDING-LINE: "<quote>"` inside its Evidence field; the checker confirms the
quote as a fixed-string match in the file the card cites. A mutant file under `evidence/mutants/` exists for
every counter and every mode, so `check-evidence.sh negative-controls` re-executes them rather than reading a
recorded result.

## 5. Risks and Mitigations

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| The "real home" is mis-identified (this lane runs on a profile, not `~/.claude`) | High | Protection watches the wrong tree | M1 records `$CLAUDE_CONFIG_DIR`, `$CODEX_HOME`, auth status first; protected set derived from the resolved homes; `~/.claude`, `~/.codex` as secondary watches |
| Runtime needs the login a scratch home lacks | High | Runtime rows blocked | Runtime runs on the real profile observe-only (`--plugin-dir` only); a blocked cell is `UNOBSERVED` with the quoted proof |
| Real profile drifts or is written concurrently by other lanes | High | A diff that is not the probe's | Per-command before/after manifests with `AMBIENT`/`LEAK` attribution by fixture identifiers; nothing restored |
| A probe command writes a real registry despite the plan | Low | Global state change | Forbidden-verb list audited from `commands.log` (AC-004); `home=real` lines restricted to `--plugin-dir` or read-only verbs; probe script itself is guard-unmediated (Gaps) |
| Nested `claude -p` joins messaging or loads the profile's plugins and MCP servers | High | Contaminated observations | Scrub of every live `CLAUDE_CODE_*` name; per-run token on marker lines; contamination sources listed in the verdict Gaps |
| Project `.mcp.json` or hooks need non-interactive approval, control does not fire | Medium | False negative | A control that does not fire makes the channel `UNOBSERVED(control-not-fired)`, never a negative |
| Model-quoted sentinel channels are nondeterministic | Medium | Wrong absence claim | Three runs, 3 of 3 agree for an absence claim, otherwise `PARTIAL`; mechanical channels preferred |
| `claude -p` does not load mods | Medium | R04 runtime blocked | Result table fixed in M3; static-only `PARTIAL` with the control as proof |
| Codex ignores `CODEX_HOME` or has no marketplace manifest we can satisfy | Medium | Codex cells blocked | `ISOLATION-FAILED` route runs no writing command; rejection messages saved; cells `UNOBSERVED` with quote |
| CLI version changes mid-run (auto-update) | Medium | Cells stamped inconsistently | Start and end version readings must be equal (AC-016); mismatch invalidates the batch |
| A hook sleeping past its timeout, or a prompt, hangs the probe | Medium | Stuck run | External `timeout` wrapper on every runtime command |
| Run-phase evidence is gitignored and never reaches the remote | Certain | Follow-up cards cannot read it | Durable carrier in `progress.md §E.2` with `verdict_sha256` and an evidence index (AC-016) |
| Per-run model cost | Medium | Spend | Run cap derived as 13 x 2 x 3 and printed; starved cells are `UNOBSERVED(cost-cap)` |
| A raw file is fabricated by hand | Low | Vacuous green | Not defended mechanically; named in Residual-risk |

## 6. Verification Approach

Each acceptance criterion in `acceptance.md` is a plain command (`sh .moai/reports/t1434/check-*.sh <mode>`
or a git form) plus the counters it must print. Every checker and the manifest builder carries a mutant that
turns it red, observed failing and re-executable (`.claude/rules/moai/development/verification-completeness.md`
§1.1, §1.2). `acceptance.md` pins a keyed RED-now cell and a keyed green-path cell per criterion (§2), each
measured on tree 676293144; the RED cells read run-phase files that are local-only, so the pin names the
tracked tree and the measurement is of the worktree at that time (stated in the ledger). Continued firing
(§1.3): the checkers print their swept counts, so an instrument that stops sweeping prints zero and fails.

## 7. Hand-off Notes for the Run Phase

- Run phase is a lane-authored measurement; `manager-develop` is not asked to write product code. The orchestrator may run it inline in the card worktree.
- The tracked write is `progress.md §E.2` only; nothing else under `.moai/specs/` changes during run, and nothing outside `.moai/reports/t1434/` and that block is written in the repository.
- Anything that cannot be observed is recorded as `UNOBSERVED` with a raw file and a quoted line; no cell is promoted by documentation or by passing `validate` alone. If every runtime row is blocked, or a floor row cannot reach a non-`UNOBSERVED` Claude cell honestly, the corresponding criteria fail and the run returns a blocker report; the criteria are not relaxed.
