# Research — SPEC-UPDATE-REINSTALL-LOOP-001

> Grounding record for the `moai update` clean-reinstall loop (GitHub issue #1084, Critical).
> Every claim below cites a directly-read source anchor (content token + approximate line).
> Line numbers drift — prefer the named symbol / literal string as the durable anchor.

## A. The verified loop mechanism (root cause)

### A.1 The collision (single, mechanically confirmed)

`.claude/rules/moai/design` is BOTH:

1. A `DeprecatedPaths` entry — `internal/defs/dirs.go`, in the Category B block
   (`DeprecatedSince: "SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001"`, comment "design rule
   directory", ~line 284).
2. Shipped by the v3 embedded template — `internal/template/templates/.claude/rules/moai/design/constitution.md`
   exists (22,813 bytes, added ~2026-07-11).

Ground-truth intersection (executed): iterating **all 41** `Path:` entries in `dirs.go`
against `internal/template/templates/` shows **exactly one** collision:
`.claude/rules/moai/design`. No other DeprecatedPaths entry resolves to an existing
template path. Command run:

```
for each Path in dirs.go DeprecatedPaths:
    if exists(internal/template/templates/<Path>): report collision
→ total entries: 41  |  collisions: 1  |  COLLISION: .claude/rules/moai/design
```

### A.2 How the collision drives an infinite loop

1. `probeDeprecatedPathSignal` (`internal/cli/v2_detection.go`, ~line 238) `os.Stat`s each
   `DeprecatedPaths` entry and returns positive on the first hit. Because the v3 template
   ships `.claude/rules/moai/design/`, this Signal 3 is **always positive** on any v3 project.
2. `detectV2Fingerprint` aggregates (`v2_detection.go`, ~line 142):
   `IsV2 = !V3VersionConfirmed && (Signal1 || Signal2 || Signal3)`.
   When the v3-version override does **not** fire (see §A.3), `IsV2` becomes true.
3. `runUpdate` (`internal/cli/update.go`, ~line 284) calls `runCleanReinstall` when
   `IsV2 && isMoAIProject(cwd)`.
4. `runCleanReinstall` Step 4 (`internal/cli/update_clean_install.go`, ~line 229–240)
   `os.RemoveAll`s every deprecated path (removes `.claude/rules/moai/design`), and
   Step 5 (~line 285) redeploys the embedded templates (recreates `.claude/rules/moai/design`).
5. Net-zero tree change, but the deprecated path is present again → the next `moai update`
   re-triggers from step 1. This matches the reporter's log:
   `signals: version=false agency=false deprecated=true` + `Removed 1 deprecated paths`
   (the `[clean-reinstall] Removed %d deprecated paths` format string is `update_clean_install.go` ~line 240).

### A.3 The existing partial fix — the v3-version negative-override (FRAGILE)

`probeVersionSignal` (`v2_detection.go`, ~line 174–211) + the aggregation
(`~line 142`) already carry a **v3-version negative-override** (REQ-CRR-001, added by
`SPEC-V3R6-V2-V3-CLEAN-REINSTALL-002` for #1084). It forces `IsV2 = false` **only when**
`.moai/config/sections/system.yaml` `moai.version` **starts with the literal prefix `v3.`**
(`strings.HasPrefix(v, "v3.")`, ~line 199).

Consequences (this is the key nuance the loop analysis must carry):

- On THIS repo the local `.moai/config/sections/system.yaml` has `version: v3.0.0-rc13`
  → override fires → `IsV2 = false` → **no loop here**. The bug does not reproduce on a
  well-formed v3 project.
- The override is **version-string-format-dependent**. If `moai.version` is empty, missing,
  or in any non-`v3.`-prefixed format (e.g. a bare `3.0.0` with no `v`, or a differently
  shaped pre-release), `V3VersionConfirmed` stays false, the override does NOT fire, and the
  collision re-enters the loop. The reporter's `version=false` + `deprecated=true` is
  consistent with an "other"-format version string (not `v2.`, not `v3.`, not empty) or a
  build predating the override.
- `internal/template/templates/` ships **no** `system.yaml` (confirmed: no
  `sections/system.yaml` under templates). A project's `system.yaml` is generated at
  `moai init` time; the version-string format at generation governs whether the override
  protects it. (Resolved 2026-07-17 — the init generator's version format was NOT traced and
  is treated as informational-only; scope was NOT expanded. See plan.md §B "RESOLVED — init
  system.yaml version format" and spec.md §C. The evidence prose above is retained unchanged.)

**Conclusion**: the override is a band-aid whose protection is conditional on the version
string. Removing the collision entry (R1) is the **version-format-independent, regression-proof**
fix — it eliminates the loop's fuel regardless of whether the override fires.

## B. Count / test-guard facts (atomic-update requirement)

`internal/defs/dirs.go` `@MX:REASON` (~line 50) states the 41-entry total + 9/29/3 category
split is governed by `SPEC-DEPRECATEDPATHS-RECONCILE-001` and "Modifications MUST update both
this slice and internal/defs/dirs_test.go atomically."

`internal/defs/dirs_test.go` guards (read directly):

- `TestDeprecatedPathsTotalCount`: `const want = 41`.
- `TestDeprecatedPathsCategorySplit`: `wantCategoryA = 9`, `wantCategoryB = 29`, `wantCategoryC = 3`.

Removing `.claude/rules/moai/design` (a Category B entry) requires, in the same change:
total `41 → 40`, Category B `29 → 28`. The comment strings in `dirs.go` (@MX:REASON, ~line 50)
and `v2_detection.go` (header comment "41 entries", ~line 32) that recite the count must also
be updated to avoid stale-count drift.

## C. Secondary defect verification

### C.1 R3 — settings.json + user.yaml overwritten on clean-reinstall (CONFIRMED)

- **PRESERVE inventory does NOT cover them.** `buildPreserveInventory`
  (`internal/cli/update_preserve_inventory.go`, ~line 84) composes `collectUserOwnedFiles`
  (scans `.claude/skills`, `.claude/agents`, `.moai/harness` per its header) with
  `preserveInventoryRoots = {.moai/specs, .moai/project, .claude/commands}` (~line 66).
  Neither `.claude/settings.json` nor `.moai/config/sections/user.yaml` is in any scan root.
- **Clean-reinstall bypasses the normal 3-way merge.** `runUpdate` returns early after a
  successful `runCleanReinstall` (`update.go`, ~line 322), so the normal-path merge never runs.
  The normal path DOES protect them: `collectMergeableFiles` (~line 799) includes
  `.claude/settings.json` and `.moai/status_line.sh`, and `.moai/config/sections/*.yaml`
  (incl. `user.yaml`) are 3-way merged by `restoreMoaiConfig` (comment ~line 792).
- **Force deploy overwrites wholesale.** Clean-reinstall Step 5 builds the deployer via
  `NewDeployerWithRendererAndForceUpdate(embedded, renderer, true)` (`update_clean_install.go`,
  ~line 256). In `forceUpdate` mode the deployer SKIPS the user-file-protection guard
  (`internal/template/deployer.go`, ~line 168–185: "Skip this check in forceUpdate mode")
  and goes straight to `atomicWriteFile` (~line 201). The template ships
  `.claude/settings.json.tmpl` (16,269 bytes) and `.moai/config/sections/user.yaml.tmpl`
  (`name: "{{.UserName}}"`) — so on every clean-reinstall, `settings.json` is replaced with
  template defaults (dropping effortLevel/teammateMode/permissions/non-MoAI hooks, forcing
  outputStyle + model) and `user.yaml` `name` is re-rendered from `{{.UserName}}` (blank when
  the update template context carries no operator name).

**Verdict: R3 CONFIRMED.** The clean-reinstall path drops the 3-way merge protection the
normal update path has.

### C.2 R2 — "destructive-before-validation" claim (PARTIALLY REFUTED — read carefully)

The reporter claimed symlink / nested-`.git` aborts fire AFTER the destructive phase, leaving
a half-migrated project. Against the **current** canonical 7-step order
(`update_clean_install.go`, `runCleanReinstall` ~line 129–327) this is **not** what happens:

Canonical order: Step 2 inventory → Step 3 backup → **Step 4 REMOVE** → Step 5 reinstall →
Step 6 merge-back → Step 7 integrity. Backup precedes removal.

- **Symlink source**: `copyFile` refuses a symlink source
  (`internal/cli/update_archive.go`, ~line 367–368:
  `"refusing to dereference symlink source"`). A symlink under a PRESERVE root
  (`.moai/specs`/`.moai/project`/`.claude/commands`) is enumerated by the `WalkDir` in
  `buildPreserveInventory` (~line 108, `WalkDir` does not follow symlinks, `IsDir()==false`),
  so it reaches `copyFile` during **Step 3 snapshot** and aborts **BEFORE the destructive Step 4**.
  Symlinks under `.claude/skills` are silently skipped by `collectUserOwnedFiles`
  (`update_namespace_protect.go`, ~line 122) — a different data-loss risk (not backed up),
  not an abort.
- **Nested `.git`**: a `WalkDir` permission error inside a PRESERVE root propagates out of
  `buildPreserveInventory` (~line 108–137) → abort at **Step 2, BEFORE destruction**.

So on the current code both triggers are **fail-closed before any destruction**; the
"half-migrated, PRESERVE snapshot had 3 of 444 files" narrative does not reproduce with the
7-step order for these two triggers (the partial snapshot would be pre-destruction, leaving the
project intact — deprecated paths NOT yet removed).

**Residual R2 value (real, but reframed as hardening, not a bug fix):**
- There is no EXPLICIT pre-flight scan that emits an actionable, named-path error before Step 2;
  the abort surfaces as an opaque mid-walk (`WalkDir permission denied`) or mid-copy
  (`refusing to dereference symlink source`) error. R2 should add a clear pre-flight validation.
- A genuine half-migration is still theoretically possible if Step 5 deploy or Step 6 merge-back
  fails after Step 4 (no rollback of Step 4). REQ-RIL-006 targets this invariant.

### C.3 R4 — `"model": "sonnet"` template pin (CONFIRMED present; policy OPEN)

`internal/template/templates/.claude/settings.json.tmpl` line ~398: `"model": "sonnet",`.
On the clean-reinstall force-deploy this replaces any user-configured model. On the normal
path the `.claude/settings.json` 3-way merge (C.1) preserves the user's model, so the
downgrade is specific to the clean-reinstall bypass. Whether the template SHOULD pin a model
at all is a product decision — RESOLVED 2026-07-17 as option (b) merge-preserving: the pin is
KEPT and the user's model survives via the clean-reinstall settings merge (R3). See plan.md §B
"RESOLVED — model pin policy" and spec.md §B.4 REQ-RIL-010. (The `outputStyle: MoAI-Easy` pin is
a deliberate product default per CLAUDE.local.md §22.6 and is OUT OF SCOPE.)

## D. Gaps (explicitly NOT verified in this pass)

- The `moai init` system.yaml **generator's version-string format** (does a fresh v3 project
  always get a `v3.`-prefixed `moai.version`?) was not traced. This determines how often the
  override silently fails in the wild. Deferred to plan.md clarification.
- Whether the reporter's rc12 build actually contained the REQ-CRR-001 override (git archaeology
  on rc12) was not performed; the orchestrator baseline states rc12 vs main `dirs.go` are
  identical, but that does not settle the version-string format on the reporter's machine.
- The exact set of user keys the settings.json 3-way merge preserves on the NORMAL path
  (effortLevel/teammateMode/permissions/hooks) was inferred from `collectMergeableFiles`, not
  from reading the merge engine body.

## E. Cross-references

- `internal/defs/dirs.go` — `DeprecatedPaths` SSOT (41 entries), `@MX:REASON` count contract.
- `internal/defs/dirs_test.go` — `TestDeprecatedPathsTotalCount` / `TestDeprecatedPathsCategorySplit`.
- `internal/cli/v2_detection.go` — `detectV2Fingerprint`, `probeVersionSignal` (v3 override),
  `probeDeprecatedPathSignal`.
- `internal/cli/update_clean_install.go` — `runCleanReinstall` 7-step order.
- `internal/cli/update_preserve_inventory.go` — `buildPreserveInventory`, `preserveInventoryRoots`.
- `internal/cli/update_namespace_protect.go` — `collectUserOwnedFiles` (skips symlinks).
- `internal/cli/update_archive.go` — `copyFile` (`refusing to dereference symlink source`).
- `internal/cli/update.go` — `runUpdate` clean-reinstall trigger + early return, normal-path merge.
- `internal/template/deployer.go` — `Deploy` forceUpdate guard skip.
- `internal/template/templates/.claude/settings.json.tmpl` — `"model": "sonnet"` pin.
- `internal/template/templates/.moai/config/sections/user.yaml.tmpl` — `name: "{{.UserName}}"`.
