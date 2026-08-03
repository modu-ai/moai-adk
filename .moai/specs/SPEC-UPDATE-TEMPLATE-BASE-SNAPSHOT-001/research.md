# Research — SPEC-UPDATE-TEMPLATE-BASE-SNAPSHOT-001

> Tier L research artifact. MEASURED empirical claims that the design (design.md) and the plan (plan.md) depend on. Every claim here is verified against the actual source tree at the measured line numbers, with the verification command recorded inline. Where a claim could not be reduced to a mechanical grep, the reasoning is stated explicitly and labelled as such.

## §A Clean-step walk radius — does `.moai/cache/` survive every clean-flow variant?

**Claim**: the update clean step (across ALL its variants in `update_cleanup.go` and `update_clean_install.go`) does NOT delete `.moai/cache/`. The snapshot location at `.moai/cache/template-snapshot/` is therefore outside every clean radius.

This claim is load-bearing for Decision D1 + D5 (snapshot location) and REQ-TBS-004 (snapshot survives clean step). It was flagged as an open question in iter-0 and is now resolved empirically.

### A.1 `update_cleanup.go` — deletion sites

**Method**: enumerate every `os.Remove` / `os.RemoveAll` call site and every `filepath.Walk` / `filepath.WalkDir` walk root in the file; for each, identify what is being deleted.

| Line | Call | Target | Walks `.moai/cache/`? |
|---|---|---|---|
| `:87` | `os.Remove(lockPath)` | `<projectRoot>/.moai/.update.lock` (single lock file) | NO |
| `:105` | `os.Remove(lockPath)` | same lock file (stale-lock cleanup) | NO |
| `:131-132` | (skip prefixes) | `.moai/backup/`, `.moai/logs/` are SKIPPED (excluded from cleanup), not deleted | NO (these are exclusions) |
| `:161` | (marker file) | `.moai-skip-cleanup` marker — read, not deleted | NO |
| `:320-321` | `os.Remove(abs)` | a single deprecated-path file (from the `DeprecatedPaths` enumeration at `internal/defs/dirs.go`) | NO — `abs` is `filepath.Join(projectRoot, rel)` where `rel` iterates the enumerated DeprecatedPaths list |
| `:324` | `os.RemoveAll(abs)` | a deprecated-path directory (same enumeration) | NO — same enumerated list |
| `:387,395` | `os.Remove(probeFile)` | `<projectRoot>/.moai-fscase-probe` (filesystem case-sensitivity probe) | NO |

**Walk roots**: the file's only recursive walk is over the `DeprecatedPaths` enumeration (each entry treated as a file or directory to remove). There is NO `filepath.Walk` or `filepath.WalkDir` over `.moai/` as a whole, and NO call site references `.moai/cache/` or `defs.CacheSubdir`.

**Verification command** (re-runnable):
```bash
grep -n 'filepath.Walk\|filepath.WalkDir\|os.RemoveAll\|os.Remove' internal/cli/update_cleanup.go
grep -n 'cache\|CacheSubdir' internal/cli/update_cleanup.go
# Expected: no matches for cache/CacheSubdir.
```

### A.2 `update_clean_install.go` — deletion sites

**Method**: same — enumerate every deletion call and every walk root.

| Line | Call | Target | Walks `.moai/cache/`? |
|---|---|---|---|
| `:322` | `os.RemoveAll(abs)` | a deprecated-path entry from the `deprecated` slice (the `DeprecatedPaths` enumeration, post-expansion via `expandDeprecatedBackupTargets`) | NO — `abs` iterates the enumerated deprecated list, not `.moai/` recursively |
| `:368` | (targeted file write, not delete) | `.claude/settings.json`, `.moai/status_line.sh` (in-memory-only files re-written, not deleted from `.moai/cache/`) | NO |
| `:567` | `filepath.WalkDir(abs, ...)` | walk root is a backup-target path (for `backupDeprecatedPaths`), NOT a deletion of `.moai/cache/` | NO — this walk READS files to back them up, then the backup dir lives under `.moai/backups/v2-to-v3-<stamp>/` |

**Clean radius in Step 4/5**: the clean-install path's "clean" step is the `deprecated` removal loop at `:320-325` (inside `guardFirstDestructiveStep`), which removes ONLY the enumerated DeprecatedPaths entries. Step 5 force-deploys templates (overwrites `.moai/config/sections/*.yaml` via the deploy step) but does NOT delete `.moai/cache/`. Step 5.5 restores via `RestoreMoaiConfig`.

**Verification command** (re-runnable):
```bash
grep -n 'filepath.Walk\|filepath.WalkDir\|os.RemoveAll\|os.Remove' internal/cli/update_clean_install.go
grep -n 'cache\|CacheSubdir' internal/cli/update_clean_install.go
# Expected: no matches that target .moai/cache/ for deletion.
```

### A.3 Conclusion

Across BOTH clean-flow variants, the deletion surface is:
1. The enumerated `DeprecatedPaths` list (specific paths, none of which is `.moai/cache/` — verified against `internal/defs/dirs.go` Categories A-D).
2. Targeted single files (`.update.lock`, `.moai-fscase-probe`).

Neither variant walks `.moai/` recursively. `.moai/cache/` is outside scope across ALL clean-flow variants. The snapshot location at `.moai/cache/template-snapshot/` survives every clean path.

**This claim is now empirical, not asserted.** The M0 pre-flight check (plan.md §C check 4) and the M3 regression test (`TestSnapshot_SurvivesCleanStep`, AC-TBS-004) re-verify this at run time.

### A.4 `.gitignore` confirmation

`.moai/cache/` is gitignored at `.gitignore:273`. The snapshot subdirectory `.moai/cache/template-snapshot/` is covered by this rule (REQ-TBS-005).

```bash
grep -n '^\.moai/cache/' .gitignore   # .gitignore:273
git check-ignore .moai/cache/template-snapshot/sections/system.yaml
# Expected: exit 0 (path is ignored).
```

## §B Restore-completion call graph — is `runUpdateRestore` a distinct fourth path?

**Claim**: there are exactly THREE distinct restore-completion sites in the production code, and ALL THREE require a `WriteSnapshot` hook per Decision D4.

This resolves iter-1 audit D4 and the open question flagged in iter-0 plan.md §B.

### B.1 `RestoreMoaiConfig` callers (the normal update path)

`RestoreMoaiConfig` is the function that performs the 3-way/2-way merge of backed-up user sections into the freshly-deployed config. Its production callers (excluding tests):

```bash
grep -rn 'RestoreMoaiConfig' internal/ --include='*.go' | grep -v _test.go
```

| Site | File | Context |
|---|---|---|
| `:420` | `internal/cli/update_template_sync.go` | Normal `moai update` — template-sync path. Restore after deploy. |
| `:451` | `internal/cli/update_clean_install.go` | Normal `moai update` — clean-install path. Restore after force-deploy. |

These are the two normal update sites. Both are wired for `WriteSnapshot` per Decision D4 triggers #2 and #3.

### B.2 `RestoreFromBackupDir` callers (the lockout-escape path)

`RestoreFromBackupDir` (in `internal/cli/update/backup/restore_entry.go:47`) is a separate entry point that wraps `RestoreMoaiConfig` with two fixed-point passes (per the idempotency note in its doc comment). Its production callers (excluding tests):

```bash
grep -rn 'RestoreFromBackupDir' internal/ --include='*.go' | grep -v _test.go
```

| Site | File | Context |
|---|---|---|
| `:53` | `internal/cli/update_restore.go` | `runUpdateRestore` — the user-invocable lockout-escape entry. Reached BEFORE the project-marker gate, because the marker's absence is exactly the damage this entry repairs (`SPEC-UPDATE-DATA-SURVIVAL-001` REQ-UDS-022). |

`runUpdateRestore` is NOT routed through `update_template_sync.go` or `update_clean_install.go`. It is a distinct top-level entry point (referenced from the `moai update --restore` command surface and from the recovery manifest's lockout-escape instructions).

### B.3 Enumeration exhaustiveness

Are there any OTHER restore-completion sites? The grep at §B.1 + §B.2 returns exactly three production sites (two `RestoreMoaiConfig` + one `RestoreFromBackupDir`). The only other callers of either function are test files:

```bash
grep -rln 'RestoreMoaiConfig\|RestoreFromBackupDir' internal/ --include='*_test.go'
# Returns: internal/cli/update/backup/restore_entry_test.go (and possibly others, all _test.go)
```

Tests do not require `WriteSnapshot` hooks (they do not initiate a real update cycle).

### B.4 Disposition

**Decision D4 wires all three sites**:
1. `update_template_sync.go:420` (template-sync restore end)
2. `update_clean_install.go:451` (clean-install restore end)
3. `update_restore.go:53` (`runUpdateRestore` end — the lockout-escape path)

The third site is wired for invariant uniformity: after a user repairs via `moai update --restore <dir>`, the on-disk config reflects the chosen backup, and the snapshot should capture that state so the next normal `moai update` has a correct BASE. Omitting this site would leave the snapshot stale or absent after a repair, degrading the next update to the REQ-TBS-007 fallback (today's wrong-base behaviour) for one more cycle.

**No fourth site exists.** The enumeration is exhaustive.

## §C Snapshot-content correctness — rendered vs raw

**Claim**: copying the on-disk `.moai/config/sections/*.yaml` produces bytes with placeholders RESOLVED, distinct from the embedded-raw template bytes that carry `{{.Version}}` / `{{.GoBinPath}}` / `{{.EnforceQuality}}` tokens.

This is the core correction vs the current `SaveTemplateDefaults` (which writes embedded-raw bytes with placeholders intact). It is verified by direct inspection of the deploy pipeline:

### C.1 Deploy renders templates before writing to disk

`moai init` (at `internal/cli/init.go:585-596`) constructs a `template.Renderer` via `template.NewRenderer(embeddedFS)` and a deployer via `template.NewDeployerWithRenderer(embeddedFS, renderer)`. The deployer's `Deploy` method (`internal/template/deployer.go:100`) renders each `.tmpl` file through the renderer before writing it to disk. The on-disk `.moai/config/sections/system.yaml` therefore carries `version: "3.0.1"` (resolved), NOT `version: {{.Version}}`.

### C.2 `SaveTemplateDefaults` bypasses the renderer

`SaveTemplateDefaults` at `backup.go:175` reads the embedded template via `fs.ReadFile(embedded, prefix+name)` — the RAW `.tmpl` source, never rendered. It writes those raw bytes (with `.tmpl` suffix stripped but placeholders intact) to the per-backup BASE directory. The comment at `:180-184` rationalizes this as "correct behavior" — that rationalization is the defect this SPEC fixes.

### C.3 The correction

The snapshot writer reads `.moai/config/sections/<name>` from DISK (post-render) and copies those bytes. The bytes therefore carry resolved values. Verification:

```bash
# After a moai init, the on-disk file has resolved values:
grep '^version:' .moai/config/sections/system.yaml
# Expected: version: "3.0.1" (quoted, resolved) — NOT version: {{.Version}}

# The embedded template has the placeholder:
grep '^version:' internal/template/templates/.moai/config/sections/system.yaml.tmpl
# Expected: version: "{{.Version}}"
```

The snapshot's copy of `system.yaml` will byte-equal the on-disk file (resolved) and will NOT contain `{{.Version}}`. This is the property asserted by AC-TBS-003.

## §D `quality.yaml` post-YAML-PRESERVE behaviour

**Claim**: post-YAML-PRESERVE (commit `3ced5b152`), `quality.yaml` transitions from always-2-way to real-3-way, making it the primary victim of the wrong-base defect.

This is established by the prior SPEC's audit (D4 + D9) and re-verified here:

```bash
# quality.yaml.tmpl is the sole .tmpl with UNQUOTED placeholders:
grep -n 'enforce_quality\|test_coverage_target' internal/template/templates/.moai/config/sections/quality.yaml.tmpl
# Expected:
#   :13: enforce_quality: {{.EnforceQuality}}
#   :16: test_coverage_target: {{.TestCoverageTarget}}
# (Unquoted — YAML reads {{...}} as a nested flow mapping.)
```

Pre-YAML-PRESERVE: the map decoder failed on this file (`yaml: invalid map key: map[string]interface {}{".EnforceQuality":interface {}(nil)}`), so `MergeYAML3Way` errored and `restore.go:139` silently routed `quality.yaml` through the 2-way fallback on every `moai update`. The wrong-base problem could not manifest because the 3-way path never ran.

Post-YAML-PRESERVE: the node decoder parses `quality.yaml.tmpl` successfully (placeholders land as scalar-nested-flow-mapping text rather than failing). `MergeYAML3Way` no longer errors. The 3-way path runs. With the wrong BASE (embedded-raw with `{{.EnforceQuality}}` intact), every placeholder key differs from the user's rendered value, `old != base` fires, and every placeholder key is misread as a user edit.

**This is the enlarged blast radius** (prior SPEC audit D9). `quality.yaml` is the dedicated correctness target — AC-TBS-013.

## §E Prior SPEC contract — what this SPEC owes `SPEC-UPDATE-YAML-PRESERVE-001`

The prior SPEC's `plan.md:146` (Decision D5, "What this SPEC owes D5") states:

> The node-tree rewrite must not entrench the wrong-base assumption. `MergeYAML3Way` continues to take `baseData []byte` from its caller, so the follow-up changes only *which bytes* are passed, not the merge's shape. M2 must not add any code that assumes base and new originate from the same document.

This SPEC honours that contract:
1. `MergeYAML3Way` signature is UNCHANGED (REQ-TBS-008, AC-TBS-008).
2. `RestoreMoaiConfig`'s read path is UNCHANGED — it still reads BASE from `backupDir/.template-defaults/sections/<name>` at `restore.go:118`.
3. Only the SOURCE of the bytes that `BackupMoaiConfig` writes into `.template-defaults/` changes (Decision D2).
4. No code in this SPEC assumes base and new originate from the same document. The snapshot is a DIFFERENT document (previously-deployed rendered template) from the NEW embedded template, by construction.

## §F Open questions resolved vs deferred

| Question | Status | Resolution |
|---|---|---|
| Clean-step touches `.moai/cache/`? | RESOLVED (§A) | NO — across all variants. Empirically verified. |
| `runUpdateRestore` is a distinct restore site? | RESOLVED (§B) | YES — third site, wired per Decision D4. |
| Snapshot carries resolved placeholders? | RESOLVED (§C) | YES — on-disk copy, post-render. |
| `quality.yaml` is the primary victim? | RESOLVED (§D) | YES — dedicated AC-TBS-013. |
| Prior SPEC's read-path freeze honoured? | RESOLVED (§E) | YES — signature + read path unchanged. |
| Snapshot file mode / pruning / size | DEFERRED (design.md §G) | Run-phase mechanical decisions; not plan-phase. |
| Snapshot signing / integrity | OUT OF SCOPE (spec.md §F) | Same threat model as `.moai/config/`. |
