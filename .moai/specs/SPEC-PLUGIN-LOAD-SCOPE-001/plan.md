# Implementation Plan — SPEC-PLUGIN-LOAD-SCOPE-001

## 1. Overview

A measurement, not a build. The run phase builds throwaway fixture plugins in a scratch project,
drives `claude` and `codex` against them under isolated config homes, and writes a verdict table
backed by observed command output. Methodology is measurement/analysis (no `cycle_type` code
implementation); run-phase evidence is `.moai/reports/t1434/verdict.md` plus raw evidence under
`.moai/reports/t1434/evidence/`.

Tier M. Ordering follows the card: fixture design, Claude static, Claude runtime, Codex, synthesis.
Within that order the decisions most likely to change come first: M1 settles whether config-home
isolation works at all, and everything after it depends on the answer.

## 2. Milestones (priority-ordered, no time estimates)

### M1 — Probe fixture design, isolation proof, snapshot protocol (Priority: High)

Most reversible decisions live here: isolation channel, fixture layout, snapshot set.

- Create the scratch root (`$SCRATCH` under the session scratchpad, unique name) and the scratch project inside it; record the paths in `evidence/env.txt`.
- Enumerate lane environment variable names from the live environment (`MOAI_KANBAN*`, `MOAI_FACTORY_*`) into `evidence/env-scrub.txt`; build the single `unset … &&` prefix from that list.
- Define the **protected set**: for Claude, `~/.claude/plugins/installed_plugins.json`, `known_marketplaces.json`, a recursive file-list-with-hash of `plugins/cache` and `plugins/marketplaces`, and `~/.claude/settings.json`; for Codex, `~/.codex/config.toml` and the `~/.codex/plugins` tree listing. Excluded by declaration: session transcripts and history (they change whenever any session runs, including this one).
- Take two manifests of the protected set at T0 and T0' before any probe command; rows that differ are **volatile-by-observation**, recorded with their diff and excluded from the byte-identity check (the real directory already shows runtime-written files with current-day modification times).
- Prove isolation (REQ-002): run `claude plugin marketplace add <scratch marketplace>` with `CLAUDE_CONFIG_DIR=<scratch home>` and show the scratch home receives the registry file while the real protected-set manifest is unchanged; repeat for Codex with `CODEX_HOME`. If isolation fails for a tool, switch that tool to the real-config protocol (before-snapshot, probe, restore, verify) and say so in `evidence/00-isolation.txt`.
- Observe, at the same time, the authentication question: one trivial `claude -p` under the isolated home, bounded by `timeout`. Result decides whether runtime observation (M3) uses the isolated home, `--plugin-dir` against the real login, or is UNOBSERVED.
- Write the fixture generator (probe script) and the negative control for the restoration diff (a deliberately altered copy of a manifest, shown to make `diff` exit 1).

Exit: isolation verdict per tool, protected-set manifests, volatile-row list, authentication verdict, fixtures generated and hashed.

### M2 — Claude static measurement (Priority: High)

- For each of R01–R14 single-component fixtures plus the composite: `claude plugin validate <path>` plain and with `--strict --json`; wrap the fixtures in a local marketplace; `claude plugin marketplace add`; `claude plugin install <name>@<marketplace>`; `claude plugin details <name>` (inventory and projected token cost); `claude plugin test` for the mod fixture. Record exit codes and deciding lines.
- Observe the A3/A4 assumptions (`details` argument form, `plugin test`, `CLAUDE_CODE_PLUGIN_DIRS`) and record misses.
- Read, read-only, one or two real installed dual-manifest plugins (layout reference only) to cross-check which component directories real plugins ship.
- Output: one raw file per command under `evidence/raw/`, one stamp header per file.

### M3 — Claude runtime observation (Priority: High)

Depends on the M1 authentication verdict.

- Scratch project with a project-tree **control** for each channel: marker-file hook in `.claude/settings.json`, sentinel rule in `.claude/rules/`, sentinel `CLAUDE.md`.
- Load each fixture with `--plugin-dir` (no registry write) or from the isolated installed state, run a bounded `claude -p` session, and read the mechanical channel: marker files written by hooks, the skill/agent/command listing, MCP tool listing, sentinel quotation for rules and instructions.
- R08 settings-hook equivalence: same moai-shaped `bash -c` wrapper (`${CLAUDE_PROJECT_DIR}`-prefixed script path) in a plugin `hooks.json` and in project `settings.json`; marker files record expanded `CLAUDE_PROJECT_DIR` / `CLAUDE_PLUGIN_ROOT`; a 7-second sleeping command under a 5-second timeout, wrapped in `timeout`, shows whether the cutoff applies.
- R09/R10: a rules fixture inside the plugin, always-loaded and `paths:`-scoped; ask the session to quote the sentinel at start and again after touching a matching path; PROJECT-ONLY only with the positive control observed.
- Run each sentinel-quoting channel twice; a model-dependent channel that disagrees with itself is recorded PARTIAL or UNOBSERVED, never PLUGIN-OK.

### M4 — Codex measurement (Priority: Medium)

- `.codex-plugin` fixtures per Codex-applicable row; `codex plugin marketplace add <path>`, `codex plugin add`, `codex plugin list` under scratch `CODEX_HOME`; there is no `codex plugin validate`, so static evidence is install and list output only.
- Runtime: `codex exec` in the scratch project where authentication allows, with project-tree controls (`.codex/config.toml`, `.codex/hooks.json`, `AGENTS.md`, `.agents/skills`); otherwise UNOBSERVED with reason.
- Rows with no Codex counterpart get a documented search (manifest schema probe via install rejection messages) and `UNOBSERVED(no-equivalent-surface-found)` only when the search found nothing.

### M5 — Verdict synthesis (Priority: High)

- Write the 28 evidence cards (`evidence/cells/R01-claude.md` … `R14-codex.md`), then `verdict.md` with the table of SPEC §5, the recommended-home and consequence columns, the stamps, the start/end version agreement, and the non-goal line.
- Write the checker (`check-verdict.sh`, no git invocation) and run it on the real verdict and on a mutated copy (one evidence reference removed) to observe it fail.
- Run the restoration proof (AC-003/AC-004) last: end manifests, `diff` against T0 minus volatile rows, negative control.
- Final: re-read `claude --version` and `codex --version`, compare with the M1 stamps.

## 3. Fixture Layout (scratch, never committed)

```
$SCRATCH/                      unique dir under the session scratchpad
  config-claude/               CLAUDE_CONFIG_DIR for the isolated runs
  config-codex/                CODEX_HOME for the isolated runs
  project/                     fresh scratch project (git init allowed here only)
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
  evidence-staging/            raw outputs before copy
```

Committed (the only repository writes): `.moai/reports/t1434/{verdict.md, probe.sh, check-verdict.sh,
evidence/{commands.log, env.txt, env-scrub.txt, 00-isolation.txt, raw/*, cells/*}}`. Fixture
`sha256` is computed over the sorted file manifest of `$SCRATCH/marketplace` and stamped into every evidence file.

## 4. Risks and Mitigations

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| Isolated `CLAUDE_CONFIG_DIR` loses the keychain login, so `claude -p` cannot run | High | Runtime rows blocked | Decide at M1 with one bounded `claude -p`; fall back to `--plugin-dir` against the real login (no registry write) under the REQ-003 snapshot protocol; still blocked → UNOBSERVED, stated in the cell |
| Real `~/.claude/plugins` drifts during the run (other sessions, runtime sweeps), making a byte-identity diff fail or pass vacuously | High | False FAIL or false PASS | Two T0 manifests derive the volatile rows; the diff covers non-volatile rows only; row count must be non-zero; negative control proves the diff can fail |
| CLI version changes mid-run (auto-update) | Medium | Cells stamped inconsistently | Read both versions at start and end; mismatch invalidates the batch; disable nothing, only detect |
| A hook sleeping past its timeout, or a session blocked on a prompt, hangs the probe | Medium | Stuck run | External `timeout` wrapper on every runtime command; sleep fixtures bounded |
| Plugin install writes cache and registry files in the scratch home and, if isolation fails, in the real home | Medium | Global state change | Isolation proof first; protected-set diff last; failed isolation switches to snapshot-restore |
| Session transcripts and history written under the real config home by non-isolated runs | Medium | Unrestorable residue | Declared excluded set with reason; isolated home preferred; residual named in Residual-risk, never claimed byte-identical |
| Sentinel-quoting channels depend on the model | Medium | Nondeterministic cells | Prefer mechanical channels (marker files, listings); run model channels twice; disagreement downgrades the cell |
| Lane environment variables falsify env-reading behavior | Medium | Wrong hook/env observations | One compound `unset … &&` per command, enumerated from the live environment |
| Plugin `hooks.json` and settings `hooks` differ in event names or env expansion | Medium | R08 mis-measured | Measure both with the same marker wrapper and the project-tree control side by side |
| The mod surface is new and its manifest or validator text changes | Medium | Stale cell | Version-stamp every cell; record validator text verbatim |

## 5. Verification Approach

Each acceptance criterion in `acceptance.md` is a plain command plus an expected observable. The two
mechanical deliverable checks (restoration diff, verdict checker) each carry a negative control,
following `.claude/rules/moai/development/verification-completeness.md` §1.1. The RED-now ledger in
`acceptance.md` pins the starting observation to tree 802a72235.

## 6. Hand-off Notes for the Run Phase

- Run phase is a lane-authored measurement; `manager-develop` is not asked to write product code. The orchestrator may run it inline in the card worktree.
- `progress.md §E.2` receives the evidence index; nothing in `.moai/specs/` besides `progress.md` changes during run.
- Anything that cannot be observed is recorded as UNOBSERVED; no cell is promoted by documentation or by passing `validate` alone.
