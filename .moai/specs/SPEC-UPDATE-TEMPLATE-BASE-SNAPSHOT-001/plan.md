# Plan — SPEC-UPDATE-TEMPLATE-BASE-SNAPSHOT-001

Tier **L** (large). Justification in §A.

## §A Context

### Tier assessment

Tier **L**. Complexity Estimator reasoning:

- **Scope (LOC)**: estimated 400-800 LOC production. New snapshot package (write + read + fallback), two write-time hook sites (`moai init` end, `moai update` restore end), one read-time change (`BackupMoaiConfig` prefers snapshot), migration/fallback path, plus lifecycle and provenance tests. This sits in the upper-M / lower-L band; the new persisted-artifact dimension pushes it firmly to L.
- **Files affected**: estimated 10-14 files. New: `internal/cli/update/backup/snapshot.go`, `snapshot_test.go`, `base_loader.go`, `base_loader_test.go`, a provenance-focused test, a `quality.yaml`-focused test. Modified: `backup.go` (BackupMoaiConfig base-source switch), `init.go` (snapshot-write hook), `update_template_sync.go` (snapshot-write hook after restore), `update_clean_install.go` (snapshot-write hook after restore). Indirectly-touched: existing `backup_test.go` (add snapshot-present cases). This exceeds the Tier M ceiling of 15 files only on a generous count; the *constitutional* axis (new on-disk data model + lifecycle + migration) is the deciding factor and places the SPEC at L per the Tier guidance "> 1000 LOC or constitutional".
- **Blast radius**: spans three subsystems — `moai init`, the `moai update` clean-install path, and the `moai update` template-sync path. Two distinct write triggers (init-end, update-restore-end) and one read-site change. This is not a single-file fix; it is a new persisted artifact with a lifecycle.
- **Risk**: medium-high. The change is behaviour-visible on every `moai update`, and the snapshot must survive the update clean step (`update_cleanup.go` / `update_clean_install.go`) which deletes `.moai/config/`. The backward-compat surface (pre-existing installs with no snapshot) is non-trivial and carries its own design question (Decision D6 below).
- **Not Tier M**: a Tier M SPEC would not introduce a new persisted on-disk artifact with its own lifecycle and migration semantics. The prior SPEC (`SPEC-UPDATE-YAML-PRESERVE-001`) was Tier M precisely because it was a representation fix with no new persisted data model — this SPEC is the provenance fix the prior SPEC deferred *because* it needs the new persisted data model.
- **Not Tier S**: out of the question — multi-file, multi-subsystem, new artifact.

plan-auditor PASS threshold at Tier L: **0.85**.

### Verified baseline (measured this revision, 2026-08-03, against the worktree tree)

| Claim | Location | Status |
|---|---|---|
| `SaveTemplateDefaults` reads NEW embedded template raw | `backup.go:147-192` | CONFIRMED — `template.EmbeddedTemplates()` at `:150`, `fs.ReadFile(embedded, prefix+name)` at `:175`, raw bytes written at `:186` |
| `.tmpl` stripped, placeholders left intact | `backup.go:180-186` (comment rationalizes this as "correct behavior") | CONFIRMED — the rationalization IS the defect |
| Per-backup `.template-defaults/sections/` write site | `backup.go:114-118` | CONFIRMED |
| `MergeYAML3Way` signature `([]byte, []byte, []byte) ([]byte, error)` | `merge.go:25` | CONFIRMED — frozen by REQ-TBS-008 |
| BASE read at restore | `restore.go:118` (`basePath`), `restore.go:121` (`MergeYAML3Way` call) | CONFIRMED — `basePath := filepath.Join(templateDefaultsDir, "sections", relPath)` at line 118 |
| 2-way fallback at restore | `restore.go:139` (`MergeYAMLDeep`) | CONFIRMED — frozen by REQ-TBS-009 |
| `BackupMoaiConfig` call sites | `update_template_sync.go:210,345`, `update_clean_install.go:359` | CONFIRMED — 3 production sites |
| `RestoreMoaiConfig` call sites | `update_template_sync.go:420`, `update_clean_install.go:451` | CONFIRMED — 2 production sites |
| `moai init` deploy success return | `init.go:723` (`return nil`) | CONFIRMED — snapshot-write hook point (Decision D4) |
| `.moai/cache/` gitignored | `.gitignore:273` | CONFIRMED |
| `.moai/state/` gitignored | `.gitignore:207,275` | CONFIRMED |
| `defs.MoAIDir` / `ConfigSubdir` / `SectionsSubdir` | `internal/defs/dirs.go:14,96-99` | CONFIRMED |
| `template.Renderer.Render(templateName, data)` exists | `internal/template/renderer.go:55-75` | CONFIRMED — available if re-render is needed, but Decision D3 chooses on-disk copy instead |

### What this plan owes the prior SPEC

The prior SPEC's `plan.md:146` contract binds this SPEC: "the follow-up changes only *which bytes* are passed, not the merge's shape". This plan honours that contract by (a) leaving `MergeYAML3Way` byte-signature unchanged, (b) leaving `RestoreMoaiConfig`'s read path (per-backup `.template-defaults/sections/<name>`) unchanged, and (c) changing ONLY the source of the bytes that `BackupMoaiConfig` writes into that per-backup directory. The change is local to `BackupMoaiConfig` + a new snapshot package + three write-time hook points.

## §B Known Issues

- The existing `SaveTemplateDefaults` has 7+ direct test callers in the repo (`internal/cli/update/backup/backup_test.go`, `internal/cli/update_fileops_test.go`, `internal/cli/coverage_improvement_test.go`, `internal/cli/target_coverage_test.go`, `internal/cli/update/backup/backup_error_test.go`). Changing `SaveTemplateDefaults`'s signature ripples to all of them. Decision D2 chooses a non-breaking surface to avoid that ripple.
- `update_clean_install.go` runs `BackupMoaiConfig` → clean → deploy → `RestoreMoaiConfig` → ... in sequence. The snapshot write MUST happen AFTER restore completes (Decision D4), not before — otherwise the snapshot records the pre-clean (old) state and the next update's BASE is the very files being replaced.
- The snapshot directory (`.moai/cache/template-snapshot/`) is NOT under any preserve root the clean step honours, so the clean step will not delete it. But the clean step DOES touch `.moai/cache/` in some flows (per `update_cleanup.go`'s walk). Decision D5 verifies this empirically at M0 and pins the location accordingly.
- **Third restore-completion site (D4, iter-1 audit D4)**: `internal/cli/update_restore.go:53` calls `backup.RestoreFromBackupDir` from `runUpdateRestore` — the user-invocable lockout-escape entry point (referenced by `SPEC-UPDATE-DATA-SURVIVAL-001` plan.md §B.3). This is a DISTINCT path from the two normal update sites (`update_template_sync.go:420` and `update_clean_install.go:451`); it is NOT routed through either. After it runs, the on-disk config reflects the chosen backup. Decision D4 wires a third `WriteSnapshot` hook here so the snapshot invariant ("snapshot = what the current install just received") holds uniformly across all three restore-completion sites; the research.md call-graph work confirms this enumeration is exhaustive.

## §C Pre-Flight (Run-Phase Entry Checks)

```bash
# 1. Confirm the prior SPEC is merged (this SPEC depends on it)
git log --oneline -1 3ced5b152 2>/dev/null && echo "prior SPEC merged" || echo "BLOCKER: prior SPEC not on this branch"

# 2. Confirm the baseline suite is green before any edit
go test ./internal/cli/update/backup/... -count=1 2>&1 | tail -20
go test ./internal/cli/... -count=1 2>&1 | tail -20

# 3. Capture the pre-change package coverage as the baseline for the coverage constraint
go test -cover ./internal/cli/update/backup/

# 4. Confirm the clean step does NOT delete .moai/cache/ (research.md §A measures this exhaustively)
grep -n "filepath.Walk\|os.RemoveAll\|os.Remove" internal/cli/update_cleanup.go internal/cli/update_clean_install.go
# Expected: every RemoveAll/Remove site targets an enumerated DeprecatedPaths entry or a
# targeted lock/probe file; NONE walks .moai/ recursively and none references .moai/cache/.

# 5. Enumerate ALL restore-completion sites (research.md §B resolves the call graph)
grep -rn "RestoreMoaiConfig\|RestoreFromBackupDir" internal/ --include='*.go' | grep -v _test.go
# Expected production sites (3 distinct paths):
#   internal/cli/update_template_sync.go:420   (normal update — template-sync path)
#   internal/cli/update_clean_install.go:451   (normal update — clean-install path)
#   internal/cli/update_restore.go:53          (user-invocable runUpdateRestore / lockout-escape)
# All three require a WriteSnapshot hook per Decision D4.

# 6. Confirm the snapshot location candidate is gitignored
grep -n "\.moai/cache" .gitignore
```

## §D Constraints (Hard)

1. `MergeYAML3Way` signature frozen (REQ-TBS-008). No edit to `merge.go:25` beyond comments.
2. `RestoreMoaiConfig`'s read path frozen — it continues to read BASE from `backupDir/.template-defaults/sections/<name>` (`restore.go:118`). No edit to `restore.go` beyond comments.
3. 2-way fallback (`MergeYAMLDeep` at `restore.go:139`) MUST remain available for unparseable bases (REQ-TBS-009).
4. Snapshot MUST live under `.moai/cache/` (Decision D5) so it survives the clean step; `.moai/cache/` is already gitignored (NFR-TBS-004 / NFR-TBS-007).
5. Snapshot MUST NOT be widened beyond `.moai/config/sections/` (REQ-TBS-015).
6. Snapshot write MUST be best-effort non-blocking (REQ-TBS-014): a snapshot write failure MUST NOT fail the enclosing `moai init` / `moai update`.
7. No new dependency (NFR-TBS-003); no template-source edit (NFR-TBS-004).
8. Package coverage of `internal/cli/update/backup/` MUST NOT fall below the M0-captured baseline.
9. Backward-compat: pre-existing installs (no snapshot) MUST update cleanly via the REQ-TBS-007 fallback. Any regression here blocks merge.

## §E Self-Verification (Design Decisions)

### Decision D1 — snapshot location: `.moai/cache/template-snapshot/`

**Chosen**: `.moai/cache/template-snapshot/sections/`.

**Candidates considered**:
- `.moai/state/template-snapshot/` — also gitignored, also survives clean. Rejected: `state/` is semantically "session/runtime state" (per `defs.StateSubdir`), not "rendered template artifact". The snapshot is a cache of deployed-on-disk content; `cache/` is the honest name.
- `.moai/config/.template-snapshot/` (hidden inside config) — Rejected: the clean step deletes `.moai/config/` wholesale, so a snapshot there would be destroyed before the next update reads it. This is the load-bearing constraint.
- `.moai-backups/template-snapshot/` (sibling of `BackupsDir`) — Rejected: `.moai-backups/` is for rotating per-timestamp config backups (`CleanupOldBackups`); mixing a persistent snapshot into that namespace invites accidental pruning.

**Why `.moai/cache/` clears every constraint**:
- Survives the clean step — `update_cleanup.go` walks `.moai/config/`, not `.moai/cache/`. Verified at M0 pre-flight (§C check 4).
- Already gitignored — `.gitignore:273` (`.moai/cache/`).
- Semantically honest — the snapshot is a rendered copy of content whose source-of-truth lives in the embedded templates; that is precisely a cache.

### Decision D2 — non-breaking API: add `SaveTemplateBase(destDir, projectRoot)`, keep `SaveTemplateDefaults` as the embedded-raw fallback

**Chosen**: introduce a new function `SaveTemplateBase(destDir, projectRoot string) error` in the backup package. Its behaviour:
1. If `.moai/cache/template-snapshot/sections/` exists, copy each snapshot section into `destDir/sections/` (the per-backup BASE directory).
2. Else, delegate to the existing `SaveTemplateDefaults(destDir)` (current embedded-raw behaviour).

`BackupMoaiConfig` (`backup.go:114-118`) is changed to call `SaveTemplateBase(templateDefaultsDir, projectRoot)` instead of `SaveTemplateDefaults(templateDefaultsDir)`.

**Rejected — change `SaveTemplateDefaults`'s signature to accept `projectRoot`**: ripples to 7+ direct test callers across 5 files. The non-breaking surface keeps every existing `SaveTemplateDefaults` test valid as a test of the fallback path, which is exactly what those tests should assert.

**Rejected — read the snapshot directly inside `RestoreMoaiConfig`**: changes the read path, which the prior SPEC's plan.md:146 contract and REQ-TBS-008 forbid. The per-backup `.template-defaults/` directory remains the BASE carrier that `RestoreMoaiConfig` reads; only its *source* moves.

**Cascade**: zero production callers of `SaveTemplateDefaults` change (they were all via `BackupMoaiConfig`). All existing `SaveTemplateDefaults` tests remain valid as fallback-path tests. A new `SaveTemplateBase` test set covers both branches (snapshot-present, snapshot-absent).

### Decision D3 — snapshot content source: copy on-disk rendered files, not re-render from templates

**Chosen**: the snapshot writer reads `.moai/config/sections/<name>` from disk and copies the bytes verbatim into `.moai/cache/template-snapshot/sections/<name>`.

**Rejected — re-render each section via `template.Renderer.Render`**: produces the same bytes in principle, but requires reconstructing the `TemplateContext` (GoBinPath, HomeDir, ...) at snapshot-write time, introduces a second rendering site that could drift from the deploy-time render, and re-renders `system.yaml.tmpl`'s `version: {{.Version}}` against the *binary's* version — which may differ from the version recorded in the on-disk `system.yaml` if the user is mid-update. The on-disk file IS the rendered truth; copying it is both simpler and more correct.

**Edge case — files in `.moai/config/sections/` that are NOT template-sourced** (user-created custom sections, e.g. a project-local `foo.yaml`): the snapshot copies them too, because `SaveTemplateDefaults`'s current scope is "every file under sections/" (it walks the embedded FS, but the merge at restore time walks the *backup*, which includes user-created sections). Copying user-created sections into the snapshot is harmless — the 3-way merge's `old == base` test fires for them, and since they have no NEW template counterpart they route through the "target file does not exist" branch at `restore.go:89-95` anyway. No behaviour change for user-created sections.

### Decision D4 — write timing: end of successful init AND end of every successful restore-completion site (three sites)

**Chosen**: FOUR write triggers total — one at init end, three at restore-completion sites.

1. **`moai init` end** — after `init.go:697` (`ScaffoldEvolutionDir`) and before `init.go:723` (`return nil`). At this point templates are deployed and rendered on disk; the snapshot captures the fresh-install baseline.
2. **`moai update` restore end — template-sync path** — after `RestoreMoaiConfig` returns nil in `update_template_sync.go:420`. NEW templates deployed + merged; snapshot captures "what this update just produced".
3. **`moai update` restore end — clean-install path** — after `RestoreMoaiConfig` returns nil in `update_clean_install.go:451`. Same invariant.
4. **`runUpdateRestore` end — lockout-escape path** — after `backup.RestoreFromBackupDir` returns nil in `update_restore.go:53`. The user-invocable restore entry applies a chosen backup to the tree; the post-restore on-disk config IS the new baseline for the next update, so writing the snapshot here keeps the invariant uniform. (iter-1 audit D4: this is a DISTINCT third restore-completion site, not one of the two normal update sites; the research.md §B call graph confirms the enumeration is exhaustive.)

**Why wire all three restore sites, not just the two normal ones**: omitting the lockout-escape site would leave the snapshot stale (or absent) after a user repairs via `moai update --restore <dir>`. The next normal `moai update` would then read either no snapshot (REQ-TBS-007 fallback — today's wrong-base behaviour) or a pre-repair snapshot (wrong base). Wiring the third site costs one best-effort call and closes the gap.

**Rejected — write snapshot BEFORE restore (at backup time)**: would record the pre-update (old) state as the BASE for the next update, defeating the purpose. The merge needs BASE = what the user's *previous* update delivered, which is exactly the post-restore state of the prior cycle.

**Rejected — write snapshot only on `moai init`, not on `moai update`**: would make the snapshot stale after the first `moai update` — the BASE would forever reflect the install-time template version, not the most-recently-deployed version. Every subsequent template change would then misfire `old != base` for the same wrong reason as today.

**Error handling (REQ-TBS-014)**: all four write triggers call the snapshot writer with best-effort semantics — a non-nil error is logged to stderr and swallowed. The enclosing `init`/`update`/`update-restore` returns its original result. The next update falls back per REQ-TBS-007.

### Decision D5 — snapshot survives the clean step: verified by location, asserted by test

**Chosen**: pin the location at `.moai/cache/template-snapshot/` and add a regression test that runs a clean-step simulation and asserts the snapshot is preserved.

**Mechanism**: the clean step (`update_cleanup.go`, `update_clean_install.go`) walks and deletes `.moai/config/` (and a small set of other roots). `.moai/cache/` is not among them. The M0 pre-flight (§C check 4) grep-verifies this empirically; the M3 regression test simulates a clean step by calling the same cleanup routine against a `t.TempDir()` project with a populated snapshot and asserts the snapshot survives.

**Rejected — explicit snapshot-preservation list inside the clean step**: would couple the clean step to this SPEC. The location-based approach keeps the clean step ignorant of the snapshot; the survival is a property of the directory choice, not a special-case exemption.

### Decision D6 — backward compatibility: graceful fallback to embedded-raw on first run

**Chosen**: the first `moai update` after this feature ships encounters an absent snapshot. `SaveTemplateBase` (Decision D2) delegates to `SaveTemplateDefaults`, producing the same BASE bytes as today. The update completes with today's behaviour. At the end of the update's restore, the snapshot-write trigger (Decision D4) writes the first snapshot. The NEXT update then has a correct BASE.

**Migration cost**: exactly one update cycle of "still wrong base" for pre-existing installs. This is acceptable because (a) it is strictly better than today (which is always wrong base), (b) the harm ordering makes a stale value less damaging than a destroyed file, and (c) there is no way to reconstruct a historically-correct snapshot from the current on-disk state without making the very assumption this SPEC invalidates.

**Rejected — "seed the snapshot from embedded-raw on first run, before reading": would make the first update's BASE the NEW embedded template (today's wrong behaviour) AND overwrite any future-correct snapshot. No benefit over the fallback.

**Rejected — block the first update on snapshot creation: violates REQ-TBS-013 (update must complete cleanly). Updates MUST NOT depend on a snapshot existing.

### Decision D7 — the per-backup `.template-defaults/` directory: RETAINED as the BASE carrier

**Chosen**: the per-backup `.template-defaults/` directory (`backup.go:114-118`) is retained. `RestoreMoaiConfig` continues to read BASE from `backupDir/.template-defaults/sections/<name>` (`restore.go:118`). The persistent snapshot feeds the per-backup directory at backup time via `SaveTemplateBase`; the per-backup directory then travels with the backup as before.

**Why not replace the per-backup directory with a direct snapshot read**:
- `RestoreMoaiConfig`'s read path is frozen by the prior SPEC's contract (REQ-TBS-008 + plan.md:146). Pointing restore at the snapshot would change the read path.
- The per-backup directory captures BASE at backup time, which is a consistent point-in-time. A direct snapshot read would race with a concurrent `moai update` rewriting the snapshot.
- The per-backup directory is already exercised by every existing backup/restore test. Removing it would orphan that test surface.

So the persistent snapshot is a *source* that feeds the per-backup *carrier*; the carrier is what the merge reads. Two layers, each with a single responsibility.

### Decision D8 — `quality.yaml` correctness: dedicated AC, not a side-effect

**Chosen**: `quality.yaml` (the file the prior SPEC promoted from always-2-way to real-3-way) gets a dedicated acceptance criterion (AC-TBS-013) and a dedicated test (M4). The test constructs a snapshot with a rendered `quality.yaml` (placeholders resolved), a LOCAL with the same rendered values, and a NEW template with an updated default; it asserts the merge adopts the NEW default (because `old == base`) rather than misreading the placeholder keys as user edits.

**Why dedicated**: the prior SPEC's audit D9 explicitly named `quality.yaml` as the enlarged-blast-radius victim. A generic "placeholder keys are not misread" AC would cover `quality.yaml` implicitly, but the prior SPEC's audit trail warrants an explicit, named assertion (acceptance.md AC-TBS-013) so a regression on `quality.yaml` specifically is visible in the test output, not buried in a generic failure.

## §F Milestones (Ordered by Decision-Reversibility)

Highest-change-likelihood decisions first; mechanical revisions last.

### M1 — Snapshot package skeleton (the data-model decision)

**Files**: new `internal/cli/update/backup/snapshot.go`, `snapshot_test.go`.

Implement:
- `const SnapshotSubdir = "cache/template-snapshot"` (relative to `defs.MoAIDir`).
- `func SnapshotDir(projectRoot string) string` — `filepath.Join(projectRoot, defs.MoAIDir, SnapshotSubdir)`.
- `func WriteSnapshot(projectRoot string) error` — walks `<projectRoot>/.moai/config/sections/`, copies each `.yaml`/`.yml` file verbatim into `SnapshotDir(projectRoot)/sections/<relpath>`. Best-effort: returns error on total failure (no config dir), but individual copy errors are swallowed with a stderr warning (REQ-TBS-014).
- `func HasSnapshot(projectRoot string) bool` — true iff `SnapshotDir(projectRoot)/sections/` exists and is non-empty.

Unit tests: `TestWriteSnapshot_CopiesRenderedSections`, `TestWriteSnapshot_NoConfigDirIsBestEffort`, `TestHasSnapshot_True/False`, all using `t.TempDir()`.

**Exit**: `go test ./internal/cli/update/backup/ -run 'TestWriteSnapshot|TestHasSnapshot' -count=1` green.

### M2 — `SaveTemplateBase` base-loader (the behaviour decision)

**Files**: new `internal/cli/update/backup/base_loader.go`, `base_loader_test.go`; edit `backup.go:114-118`.

Implement:
- `func SaveTemplateBase(destDir, projectRoot string) error` — if `HasSnapshot(projectRoot)`, copy each file from `SnapshotDir(projectRoot)/sections/` into `destDir/sections/`; else delegate to `SaveTemplateDefaults(destDir)` (the fallback).
- Edit `BackupMoaiConfig` at `backup.go:114-118` to call `SaveTemplateBase(templateDefaultsDir, projectRoot)` instead of `SaveTemplateDefaults(templateDefaultsDir)`. The `projectRoot` is already a parameter of `BackupMoaiConfig`.

Unit tests: `TestSaveTemplateBase_PrefersSnapshot`, `TestSaveTemplateBase_FallsBackWhenSnapshotAbsent`, `TestSaveTemplateBase_FallbackMatchesSaveTemplateDefaultsBytes`.

**Exit**: `go test ./internal/cli/update/backup/ -run 'TestSaveTemplateBase' -count=1` green; `go build ./...` clean.

### M3 — Write-time hook points + clean-step survival (the lifecycle decision)

**Files**: edit `internal/cli/init.go`, `internal/cli/update_template_sync.go`, `internal/cli/update_clean_install.go`; new regression test `internal/cli/update/backup/snapshot_survival_test.go`.

Wire `WriteSnapshot(projectRoot)`:
- In `init.go` just before the final `return nil` at `:723` — wrap in best-effort (log + swallow).
- In `update_template_sync.go` after the `RestoreMoaiConfig` block at `:420-430` — wrap in best-effort.
- In `update_clean_install.go` after the `RestoreMoaiConfig` block at `:451` — wrap in best-effort.
- In `update_restore.go` after the `RestoreFromBackupDir` call at `:53` (the `runUpdateRestore` lockout-escape path) — wrap in best-effort. This closes the third restore-completion site per Decision D4 + iter-1 audit D4.

Regression test: `TestSnapshot_SurvivesCleanStep` — populate a `t.TempDir()` project with `.moai/config/sections/` + snapshot, run the clean-step routine in-process (or simulate by deleting `.moai/config/`), assert `.moai/cache/template-snapshot/sections/` is preserved.

**Exit**: `go test ./internal/cli/update/backup/ -run 'TestSnapshot_SurvivesCleanStep' -count=1` green; `go test ./internal/cli/... -count=1` fully green.

### M4 — Provenance + `quality.yaml` correctness tests (the new contract)

**Files**: new `internal/cli/update/backup/snapshot_provenance_test.go`, `snapshot_quality_yaml_test.go`.

**Provenance test** (`TestMerge_BaseFromSnapshot_NotFromEmbedded`): construct a snapshot with rendered values (`version: 3.0.1`), a LOCAL file matching the snapshot, and a NEW template with `version: 3.1.0`. Run `SaveTemplateBase` + `MergeYAML3Way`; assert the merged output carries `version: 3.1.0` (NEW template value adopted because `old == base`), NOT `3.0.1` (which would indicate the wrong-base misread). This is the load-bearing correctness assertion for REQ-TBS-010.

**`quality.yaml` test** (`TestMerge_QualityYaml_Real3Way_WithSnapshot`, AC-TBS-013): snapshot with rendered `quality.yaml` (placeholders resolved to real values), LOCAL matching the snapshot, NEW template with updated `enforce_quality` / `test_coverage_target`. Assert the merge adopts the NEW values. This is the dedicated target per Decision D8.

**Exit**: both tests green; provenance test RED against the wrong-base implementation (falsifiability AC-TBS-010 is verified in M5).

### M5 — Falsifiability verification (the anti-vacuity gate)

**Procedure** (per acceptance.md AC-TBS-010, mirroring the prior SPEC's AC-UYP-022 pattern):

```bash
# 1. Stash the implementation files
git stash push -u internal/cli/update/backup/snapshot.go \
  internal/cli/update/backup/base_loader.go \
  internal/cli/update/backup/snapshot_provenance_test.go 2>&1

# 2. Assert the stash actually reverted the implementation
git diff --exit-code HEAD -- internal/cli/update/backup/snapshot.go && \
  echo "FAIL: snapshot.go unexpectedly present" || echo "PASS: snapshot.go reverted"
test ! -f internal/cli/update/backup/base_loader.go && \
  echo "PASS: base_loader.go absent" || echo "FAIL: base_loader.go still present"
# (the modified backup.go reverts to SaveTemplateDefaults = wrong-base)

# 3. Run the provenance test — it MUST FAIL against the wrong-base implementation
go test ./internal/cli/update/backup/ -run TestMerge_BaseFromSnapshot_NotFromEmbedded -count=1
# Expected: FAIL (wrong base → version misread as user edit → assertion fails)

# 4. Restore the implementation
git stash pop

# 5. Re-run — MUST PASS now
go test ./internal/cli/update/backup/ -run TestMerge_BaseFromSnapshot_NotFromEmbedded -count=1
# Expected: PASS
```

Both observations (RED on stashed, GREEN on restored) are recorded verbatim in `progress.md §E.2`.

**Exit**: RED step documented; GREEN step green.

### M6 — Verification sweep

```bash
go test ./... -count=1
go test -cover ./internal/cli/update/backup/          # vs M0/pre-flight baseline
golangci-lint run ./internal/cli/update/backup/...
GOOS=windows GOARCH=amd64 go build ./...
grep -rn "AskUserQuestion\|mcp__askuser" internal/cli/update/backup/ | grep -v _test.go | grep -v "// "
# Expected: 0 matches (NFR-TBS-005)
grep -rn "SPEC-UPDATE-TEMPLATE-BASE-SNAPSHOT" internal/template/templates/
# Expected: 0 matches (NFR-TBS-004 template neutrality)
```

Plus an end-to-end check: simulate a two-cycle `moai update` in `t.TempDir()` — first update with no snapshot (fallback path, writes snapshot), second update with snapshot (provenance path, correct BASE) — and assert the second update adopts a NEW template default rather than misreading it as a user edit.

### M7 — Commit

Conventional commit, one logical unit. `fix(update): use rendered template snapshot as 3-way merge BASE (#<issue>)`.

## §G Anti-Patterns (Avoid)

- **Do not** change `MergeYAML3Way`'s signature or `RestoreMoaiConfig`'s read path. Both are frozen by REQ-TBS-008 + the prior SPEC's plan.md:146 contract.
- **Do not** widen the snapshot beyond `.moai/config/sections/` (REQ-TBS-015). A "snapshot everything" temptation is exactly the scope-creep the Out-of-Scope section exists to prevent.
- **Do not** make the snapshot write a blocking failure. REQ-TBS-014 makes it best-effort; a snapshot failure MUST NOT break init/update.
- **Do not** couple the clean step to the snapshot. Decision D5 pins the location so the clean step is *ignorant* of the snapshot; adding a special-case exemption would re-couple them.
- **Do not** re-render templates at snapshot-write time. Decision D3 chose on-disk copy for both correctness and simplicity; a second render site introduces drift risk.
- **Do not** bundle the `sync_commit_sha` backfill chore or any other prior-SPEC follow-up. Scope discipline (CLAUDE.local.md §16 Rule 5).
- **Do not** claim the provenance correctness without running the M5 falsifiability check. A green test on an unreverted tree does not distinguish "correct" from "vacuous" — this is the lesson the prior SPEC's plan-audit D1 / AC-UYP-022 codified.

## §H Cross-References

- `spec.md` §C — the 15 requirements this plan implements.
- `acceptance.md` §D — the AC matrix each milestone is verified against.
- `internal/cli/update/backup/backup.go:147-192` — `SaveTemplateDefaults` (defect site, retained as fallback).
- `internal/cli/update/backup/backup.go:114-118` — `BackupMoaiConfig` base-source switch site.
- `internal/cli/update/backup/restore.go:118-121` — frozen read path (`basePath` at `:118`, `MergeYAML3Way` call at `:121`; unchanged).
- `internal/cli/update/backup/merge.go:25` — frozen `MergeYAML3Way` signature (unchanged).
- `internal/cli/init.go:723` — init-end snapshot-write hook.
- `internal/cli/update_template_sync.go:420` — update-restore-end snapshot-write hook (template-sync path).
- `internal/cli/update_clean_install.go:451` — update-restore-end snapshot-write hook (clean-install path).
- `internal/defs/dirs.go:14,96-99` — `MoAIDir`, `SectionsSubdir` path constants.
- `.gitignore:273` — `.moai/cache/` gitignore (load-bearing for REQ-TBS-005).
- CLAUDE.local.md §2 (`.moai/cache/` gitignored), §6 (test isolation), §16 Rule 5 (scope discipline).
- `.claude/rules/moai/development/spec-frontmatter-schema.md` — SPEC frontmatter schema (this SPEC's frontmatter validated against it).
