# Design — SPEC-INIT-SHRINK-001

## 1. Architecture overview

Three seams already exist and this design extends them rather than adding a fourth mechanism:

1. **The deployer family** (P-05): `ClaudeHarness(Slim)` / `DualHarness(Slim)` /
   `CodexOnly` deployers are selected by harness and slim mode at `init.go:738-767`, and update
   rebuilds the same family in `newTemplateSyncDeployer` (harness-aware, force-update). A deploy
   **mode** is a second axis on the same seam: the file-set split is an option on the deployer, not
   a new deployer type.
2. **The clean step** (P-06..P-08): update already removes managed roots before re-deploying, with
   backup semantics keyed to the raw template FS. The migration reuses the removal, and fixes the
   one hazard: the backup exemption must key to the **deploy-mode render**, not the raw template.
3. **The archive contract** (P-09/P-10): `archiveSkill` already implements idempotent,
   drift-checked, symlink-refusing backup with a versioned archive root and a guard test for list
   freshness. The migration is a second consumer of the same contract with its own classified list.

New units, deliberately few:

- `internal/cli/update/migrate_classify.go` — the classifier (REQ-010) and its three classes.
- `internal/cli/update_migrate.go` — the trigger wiring in the update flow (REQ-015).
- the post-install list-surface probe (§2.4) — the tri-state install-outcome signal this card adds
  beside the t1435 step (never a change to the step's return contract; consumed by init's MCP
  policy and the migration trigger — REQ-005, REQ-015).
- `scripts/check-bare-name-resolution.sh` (+ its self-test script) — the REQ-008 measurement.
- `deployment_mode` config key (OD-5 settled (a) 2026-10-03) beside `llm.harness`.

## 2. The deploy-mode split

### 2.1 Option shape

The deployer option carries one value with three spellings resolved at the call site:

```
DeployMode: plugin | local
```

- `plugin` — the default path: the deploy walk skips `.claude/skills/**` and `.claude/commands/**`
  (REQ-001); the skill mirror is disabled once the Codex actual-execution verification is produced
  (OD-6 settled (a) + condition — where the verification cannot be produced, the mirror stays
  deployed **with its entries re-homed to real directory copies rendered from the embedded
  template tree** (P-11's existing copy fallback), never left as the symlinks into
  `../../.claude/skills/<name>` that would dangle once the plugin path stops deploying
  `.claude/skills/**`; §3 mirror entries); the project `.mcp.json` render carries no `moai` entry
  when the install-outcome probe reads `confirmed` (OD-1 settled (c): the entry is written as the
  fallback carrier on `failed-or-skipped` and on `opted-out`; §2.4).
- `local` — the `--no-plugin` and `--all` paths: today's payload byte-for-byte (REQ-003, REQ-007).

Slim-mode composition: slim/full governs which catalog entries the **local** deploy carries (core
vs all tiers, P-05); on the plugin path slim/full is moot for the excluded roots and keeps governing
the rest. `--all` resolves to `local` + full-tier (OD-7 default), which is exactly today's
`shouldDistributeAll` path plus the mode record.

Codex-only: the harnessFS re-homing of P-12 composes with the mode — `local` mode re-homes to
`.agents/skills` as today; `plugin` mode deploys no skills at all, so there is nothing to re-home
and the mirror stays off (OD-6 primary path; under the no-verification fallback the mirror stays
here too, as re-homed real copies per §2.1).

### 2.2 Where the flip lands

- `internal/template/deployer.go`: the option plumbed through the constructor family
  (`WithDeployMode`), consumed by the walk that builds `ListTemplates` and by `mirrorSkills`.
- `internal/cli/init.go`: the selection switch gains the mode resolution — opt-out or `--all` →
  `local`, else `plugin` — and writes the record (REQ-009) beside the `ApplyHarness` call
  (`init.go:947`).
- `internal/cli/update_template_sync.go`: `newTemplateSyncDeployer` reads the record (REQ-016);
  the absent-record branch is §3.

### 2.3 The MCP entry

`provisionMCPEntryUnlessDeclined` (P-15) keeps its signature and its decline table; on the plugin
path the ensure-entry call for the `moai` server is skipped only when the install-outcome probe
(§2.4) reads `confirmed` (OD-1 settled (c) — the render-time filter stands on the confirmed path);
on `failed-or-skipped`, and on `opted-out` (the local-path semantics of REQ-003/REQ-005), the
provision call writes the project `moai` entry as the fallback carrier, which requires the
provision call to run after the install step and read the probe's outcome. On `local` mode
everything is as today. The `--llm gpt`
project-entry decline is unaffected (OD-1); the Codex `config.toml` wiring is untouched on every
path.

### 2.4 The install-outcome signal — the post-install list-surface probe

The t1435 install step is fail-open by contract and returns nil in every case: a failure, a
timeout, an absent tool, and the opt-out skip are indistinguishable to its caller, and the outcome
exists only as prose on the step's output writer (`internal/cli/plugin_install.go:187-202` —
"runPluginInstallStep acts on opts.Tools in order and returns nil in every case"). That return
contract is t1435's — completed — and is not changed here. This card adds its own observational
signal beside the step, consumed by §2.3 (the MCP-entry policy) and §3 (the migration trigger):
the **post-install list-surface probe**. The probe runs after the install step's process work
ends, starts no install action, and never re-runs or repairs the step.

- **`opted-out`**: the t1435 opt-out is set (`--no-plugin` or `MOAI_SKIP_PLUGIN_INSTALL=1|true`,
  t1435 OD-5). Decided before the step runs; the step is not attempted; the outcome is
  `opted-out` without probing.
- **`confirmed`**: the step ran, and for every tool it acted on, that tool's installed-plugin
  list surface names `moai@moai-adk` — `claude plugin list` for the Claude tool (a measured
  read-only verb: the t1434 probe's read-only list, SPEC-PLUGIN-LOAD-SCOPE-001 REQ-004), and
  `codex plugin list --json` for the Codex tool (the t1435 REQ-021 route, whose read pattern
  `internal/cli/doctor_plugin_version.go` `probeCodexPluginVersion` already implements). The
  probe starts its list commands through the same REQ-017 runner seam the step and the doctor
  probe use, bounded by the doctor probe's timeout-constant class. The confirmed verdict requires
  every acted tool's surface to list the plugin.
- **`failed-or-skipped`**: everything else — the tool was absent from PATH, a step command failed
  or timed out, the runner refused, a list surface does not name the plugin, or the probe could
  not read a surface. Only a listed plugin confirms; every unconfirmed outcome resolves here
  (conservative: unconfirmed means failed).

Failure modes, named: a tool list that lags the install reads unconfirmed and parks the project
in `local` — the conservative direction OD-4's amendment wants, with the documented init re-entry
(REQ-018) as the recourse; a list-surface format change degrades to unconfirmed (the probe
matches only the plugin ref string and the JSON entry — no coupling to the tool's internal
install layout, the coupling t1435 REQ-021 deliberately avoids); a partial multi-tool landing
(one surface lists, the other does not) is unconfirmed. The probe's own errors are never fatal to
the flow — they resolve to `failed-or-skipped`.

Arm mapping: `confirmed` is the only outcome that may record `plugin` or omit the project `moai`
entry; `failed-or-skipped` writes no plugin record anywhere, removes nothing, keeps
`deployment_mode: local`, and writes the project `moai` entry as the fallback carrier;
`opted-out` deploys the full local payload and writes the `moai` entry (REQ-003/REQ-005's
local-path semantics). In verification the probe is inert by the same REQ-017 refusal — a test
that exercises an arm injects a runner whose list output names (or omits) the plugin, or sets the
opt-out; no test reaches a real tool.

## 3. Migration pipeline

Ordered inside the update flow's existing step table; the Clean step stays the removal executor.

1. **Trigger** (REQ-015, OD-4 default): the version check that already gates the sync
   (`update_template_sync.go:180-189`) has passed (the post-shrink binary bumped the version). The
   mode record is read: present → no migration (the deployer split of §2.2 governs). Absent → the
   migration path runs: install step fail-open under the opt-out (the t1435 step, sequenced after
   its landing), its outcome read through the post-install list-surface probe (§2.4); a
   `failed-or-skipped` or `opted-out` probe outcome ends the migration there with the record
   written `local` and steps 2-4 unexecuted (OD-4 settled (a, amended)) — on `opted-out` the full
   local payload deploys as today; on probe `confirmed`, then classification.
2. **Classification** (REQ-010): for every file under the dropped roots
   (`.claude/skills/**`, `.claude/commands/**`, the mirror), compare the on-disk content with the
   template **render for this project's context** — the same render the deployer would write. The
   manifest record (P-13) is the fast path and the freshness check: managed + hash-equal to the
   render → `identical`; managed + different, or no usable record → `modified`; not managed
   (P-19/P-21) → `foreign`. Absent/stale records conservatively route to `modified` (RK-9).
3. **Archive** (REQ-012): `modified` items are archived first — skill directories through
   `archiveSkill` (which already refuses symlinks and aborts the run's drift contract safely), and
   standalone files (commands, mirror files) through the same layout with a file-level archive
   root. **Any archive failure aborts before the removal step runs** — the P-08 REQ-UDS-008 rule.
4. **Removal** (REQ-011/REQ-013): `identical` items are removed (no archive — the plugin and the
   template render are the recovery source); `foreign` items and symlinks are never in the
   removal list. The binding invariant: **the removal executor's scope is exactly the classified
   removal list** — the identical set plus the modified set, the latter only once REQ-012's batch
   archive has fully succeeded — never the global managed-roots walk. Primary form: the migration
   passes that list to the Clean machinery as its target list, and Clean processes exactly that
   list with its existing progress lines, crash-window guards, and backup semantics; the
   non-dropped managed roots (settings, agents/moai, rules/moai, output-styles/moai, hooks/moai,
   `.moai/config`) keep their normal Clean treatment in the same run. Clean's default global walk
   over `ManagedCleanTargets` (P-06/P-07) does not run over the dropped roots in a migration run.
   Accepted alternative, same invariant: the migration removes through its own guarded path
   reusing `backupThenRemove`, and the update flow skips Clean's dropped roots entirely for that
   run. Why the earlier scoped-FS hand-over idea is dead: `tmplFS` scopes only the backup (P-08) —
   the walk itself has no classification gate and no user-owned skip and removes every glob match,
   so no template-FS scoping could preserve a foreign `moai-custom` skill the `.claude/skills/moai*`
   glob hits, or a symlink entry. Backup disposition in the migration path: modified items carry
   their own step-3 archive before they enter the removal list (the archive IS their backup), and
   identical items are removed without backup per OD-3 (a) — the P-08 exemption is moot by
   construction, not by FS scoping. The abort-before-removal contract is unchanged. On a
   `local`-outcome migration run (`failed-or-skipped` / `opted-out`) nothing is removed by the
   migration and the normal Clean walk runs as today — the local deploy redeploys the full
   payload.
5. **Record write**: the migration ends by writing the mode record (`plugin` after a confirmed
   install, or `local` under the opt-out — in which case steps 2-4 are skipped entirely — or
   `local` when the install step failed, OD-4's amended condition).
6. **Idempotence** (REQ-014): a migrated project has a record, so step 1 short-circuits; a
   record-bearing project re-running the classifier sees an empty removed-root delta and prints
   zero counts.

Mirror entries: symlink mirror entries are removed as part of the dropped roots under OD-6
settled (a) + condition — the removal happens only after Codex is verified to actually execute
plugin-borne skills. Where the verification cannot be produced the mirror stays — re-homed, not
kept as links: mirror entries are relative symlinks into `../../.claude/skills/<name>` (P-11) and
the plugin path deploys no `.claude/skills/**` (REQ-001), so a kept symlink would dangle; the
migration re-homes kept mirror entries to real directory copies rendered from the embedded
template tree (P-11's existing copy fallback) BEFORE the dropped-root removal runs, and a fresh
plugin deploy under the fallback renders the mirror as real copies from the start. The entries
are never dereferenced (P-11's constraint; the archive contract already refuses them).

## 4. The resolution gate

`scripts/check-bare-name-resolution.sh <fixture-dir>` measures the three questions of REQ-008 with
the runtime routes t1434 validated:

- **Claude skill resolution**: a local fixture plugin (built by the script from the tree's rendered
  command/skill sources, or a checked-in minimal fixture) carrying one skill and one command whose
  body calls the bare name; a `claude -p --plugin-dir <fixture> --no-session-persistence` session
  under a scratch `CLAUDE_CONFIG_DIR` with **no scaffold copies in the project tree** is asked to
  invoke the component by its bare name; the session writes a marker file through a tool when the
  invocation fires. Marker present → bare name resolved.
- **Plugin command body**: the same session runs the plugin command and the marker records which
  skill the body's `Skill("moai")` resolved to (or that it failed).
- **Codex naming**: the fixture is added under a scratch `CODEX_HOME`; the render scan
  (`codex debug prompt-input`, the t1434 route) records the listed names for the plugin-borne
  components.
- **Hermeticity** (REQ-020): live-enumerated scrub of `MOAI_*`/`CLAUDE_*`/`CODEX_*` except the
  scratch homes it sets; scratch working directory; before/after protected-set hash with directory
  entries; `isolation-*` negative-control cases (scrub disabled → FAIL) in
  `scripts/test-bare-name-resolution.sh`, mirroring the t1435 harness shape.
- **Verdict routing**: the script prints `RESULT pass=<n> fail=<m>`; the run phase records the
  verdict in `progress.md` §E.2 before the M2 flip merges. A `fail` verdict is not a blocker for
  the harness — it is the OD-2 input; the blocker is merging the flip without a recorded verdict.

## 5. Alternatives and where each Open Decision lands

Settled 2026-10-03 (verbatim rulings in `decision-index.md` Q1..Q8; branch summary in `spec.md` §5
Settlement): OD-1 (c), OD-2 (a) + condition, OD-3 (a) + three conditions, OD-4 (a, amended),
OD-5 (a) + condition, OD-6 (a) + condition, OD-7 (a), OD-8 (a). The "Where the default lands"
column below describes the authoring-time defaults; the taken branches supersede it where they
differ — for OD-1 the (c) alternative is now the landing (the provision call runs after the
install step and observes its outcome, §2.3).

| OD | Design pressure | Where the default lands | Where the alternatives would land |
|----|-----------------|-------------------------|-----------------------------------|
| OD-1 | R06 makes a project copy inert-but-duplicated; the decline table is harness-keyed, not plugin-aware | §2.3: render-time exclusion of the `moai` entry on the plugin path | (b) no exclusion — one line removed from the filter; (c) the exclusion becomes conditional on the install step's outcome, requiring the provision call to run after the install step (ordering change in `init.go`) |
| OD-2 | The gate's verdict routes the reference form | §4 records the verdict; M2 ships bare-form if resolved | (a) the renderer grows a mode variable and the instruction templates grow conditionals (template-render work in M2); (b) the exclusion set narrows (deployer option constant changes); (c) M2's flip is deferred, M1+M3 ship alone |
| OD-3 | Archive cost vs safety | §3 step 3/4: archive modified only | (b) archive `identical` too (step 3 gains a second class; larger archives); (c) steps 3/4 are deferred behind a report step |
| OD-4 | Old projects never saw the install step | §3 step 1: install-then-classify | (b) absent record → `local` + guidance (no install call in update); (c) the wiring moves to a new verb and update only prints its name |
| OD-5 | The harness key is the persistence precedent | §2.2: `deployment_mode` in `llm.yaml` beside `llm.harness` | (b) a new section file (same reader pattern); (c) inference from the manifest (a reader-only change, no writer) |
| OD-6 | R01-codex is render-level evidence; the local mirror is cheap insurance | §2.1: mirror off on `plugin`, on for `local` | (b) the mirror option stops being mode-keyed; (c) the mirror machinery retires and the codex-only re-homing loses its local counterpart |
| OD-7 | Optional packs have no plugin home (t1435 OD-8 pin) | §2.1: `--all` = `local` + full tier | (b) `--all` = `plugin` + optional packs deployed locally (a mixed file set the docs must describe); (c) the flag becomes a deprecation notice |
| OD-8 | t1466 owns broader instruction drift | REQ-008's resolvability scope only | (b) the full reference sweep joins M2/M4; (c) only mechanically-forced text changes |

## 6. Not designed here

- The plugin payload, manifests, marketplace, install verb, and doctor check (t1435).
- Whether plugin-borne components behave identically to scaffold copies beyond what t1434 measured
  (double-firing hooks, shadowing) — t1435's gaps, not this card's.
- The broader deployed-template instruction drift (t1466).
- Branch-flow narratives (t1453).
