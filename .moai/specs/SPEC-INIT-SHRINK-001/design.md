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
- `scripts/check-bare-name-resolution.sh` (+ its self-test script) — the REQ-008 measurement.
- `deployment_mode` config key (OD-5 default) beside `llm.harness`.

## 2. The deploy-mode split

### 2.1 Option shape

The deployer option carries one value with three spellings resolved at the call site:

```
DeployMode: plugin | local
```

- `plugin` — the default path: the deploy walk skips `.claude/skills/**` and `.claude/commands/**`
  (REQ-001); the skill mirror is disabled (OD-6 default); the project `.mcp.json` render carries no
  `moai` entry (OD-1 default).
- `local` — the `--no-plugin` and `--all` paths: today's payload byte-for-byte (REQ-003, REQ-007).

Slim-mode composition: slim/full governs which catalog entries the **local** deploy carries (core
vs all tiers, P-05); on the plugin path slim/full is moot for the excluded roots and keeps governing
the rest. `--all` resolves to `local` + full-tier (OD-7 default), which is exactly today's
`shouldDistributeAll` path plus the mode record.

Codex-only: the harnessFS re-homing of P-12 composes with the mode — `local` mode re-homes to
`.agents/skills` as today; `plugin` mode deploys no skills at all, so there is nothing to re-home
and the mirror stays off (OD-6 default).

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
path the ensure-entry call is skipped for the `moai` server only — the template `.mcp.json` render
that the deploy wrote already lacks the entry (the exclusion is a render-time filter, matching how
the file is deployed), so the provision call's only remaining work on the default path is
`context7` + `staggeredStartup` repair. On `local` mode everything is as today. The `--llm gpt`
project-entry decline is unaffected (OD-1); the Codex `config.toml` wiring is untouched on every
path.

## 3. Migration pipeline

Ordered inside the update flow's existing step table; the Clean step stays the removal executor.

1. **Trigger** (REQ-015, OD-4 default): the version check that already gates the sync
   (`update_template_sync.go:180-189`) has passed (the post-shrink binary bumped the version). The
   mode record is read: present → no migration (the deployer split of §2.2 governs). Absent → the
   migration path runs: install step fail-open under the opt-out (the t1435 step, sequenced after
   its landing), then classification.
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
   template render are the recovery source); `foreign` items are never touched. The removal runs
   through the Clean step's machinery so the existing progress lines, crash-window guards, and
   recovery manifest apply; the migration hands Clean a deploy-set-scoped template FS (the
   post-shrink deployer's file set, not the raw embedded tree) so the P-08 backup exemption cannot
   silently apply to dropped components. Alternative accepted at design time: pre-remove the
   modified set through the migration's own guarded path before Clean runs — chosen only if the
   scoped-FS hand-off proves invasive; the abort-before-removal contract is identical either way.
5. **Record write**: the migration ends by writing the mode record (`plugin`, or `local` under the
   opt-out — in which case steps 2-4 are skipped entirely).
6. **Idempotence** (REQ-014): a migrated project has a record, so step 1 short-circuits; a
   record-bearing project re-running the classifier sees an empty removed-root delta and prints
   zero counts.

Mirror entries: symlink mirror entries are removed as part of the dropped roots under OD-6 (a);
they are never dereferenced (P-11's constraint; the archive contract already refuses them).

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
