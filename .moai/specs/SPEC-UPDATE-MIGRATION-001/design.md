# design.md — SPEC-UPDATE-MIGRATION-001 (card t1547)

## 1. Pipeline shape

The update-on-existing-project flow becomes a five-stage pipeline replacing the
classify-nothing / wipe-first sequence:

```
backup (unchanged) → classify (NEW, read-only, whole managed-root set)
  → reconcile per class (NEW) → refresh from template (existing deploy)
  → summary (EXTENDED outcome report)
```

`CleanMoaiManagedPaths` is NOT deleted. The update flow stops calling it for the
default (existing-project) path; the wholesale path remains for the legacy
v1→v2 fresh-install case (REQ-UPM-040) and keeps its t111 pre-clean backup.

## 2. Per-path ownership classification (spec §B.1 realized)

New function (working name) `ReconcilePlan(projectRoot, tmplFS, render, manifest) ReconciliationPlan`
in `internal/cli/update` (package location decided in run phase; deploy stays a leaf, so the
orchestration layer holds it and passes dispositions down):

| Class | Decision rule (ordered) | Disposition |
|---|---|---|
| `template-owned` | template carries path AND manifest record healthy (`template_managed`, hash matches) AND content == render | refresh in place |
| `user-modified` | template carries path AND (record absent/stale OR content != render) | 3-way merge → merged write, or conflict → preserve + `.moai-new` sidecar + report |
| `user-owned` | template does NOT carry path (covers `IsUserOwnedNamespace` by construction) | preserve untouched, list in summary |
| `stale` | manifest (or prior template) carried path AND current template does not | archive-then-remove, list in summary |

Notes:
- The class gate is template carriage (REQ-UPM-002). `IsUserOwnedNamespace` paths are forced
  `user-owned` (REQ-UPM-003) as defense-in-depth even where template carriage might disagree.
- Manifest-absent project: every template-carried file reads `user-modified` (R-1). Merge base
  falls back to the previous template render when no manifest base exists; a merge whose
  "user" side equals that base collapses to a plain refresh.
- Symlinks: carried over from the existing Lstat-first branch; never classified, never merged;
  existing dispositions unchanged.

## 3. Reconciliation dispositions

- **refresh**: current behavior, minus the pre-delete. Deployment overwrites the file from the
  render. No backup needed (template is the recovery source).
- **merge**: `plan.DetermineStrategy` picks the strategy; `internal/merge` engine runs
  base=(manifest-tracked or prior-render) / ours=disk / theirs=new render. Clean result →
  write, report `merged`. The `MergeGitignoreFile` marker-carry behavior is the model.
- **conflict**: on-disk file untouched byte-for-byte; template render written to
  `<path>.moai-new` (occupied name → first unused numbered sibling `.moai-new.2`, `.moai-new.3`,
  … — the existing sibling is never overwritten, REQ-UPM-012); conflict listed in summary with
  both paths including the collision. No auto-resolution (R-2).
- **preserve**: no write, no backup, no archive; listed in summary. Includes every file under
  `.moai/config/sections/` the template does not carry and all local-only files under managed
  roots.
- **archive-remove**: `stale` files copied to the migration archive root
  (`.moai/archive/files/<tag>/`, the existing `ArchiveFilesRoot()` layout) before removal;
  archive failure aborts the removal (REQ-UDS-008 ordering, reused). Archive copies are
  written with the package default file mode (`defs.FilePerm`, 0644) via the existing
  `copyRegularFile` — source modes are NOT inherited, so the archive is a recovery surface,
  not a confidentiality boundary; this SPEC defines no retention/cleanup policy (stale
  archive content persists until an operator cleans it).

## 4. The `.moai/config` case (REQ-UPM-020/021)

Replace wipe-then-restore with a **base-aware 3-way merge** per `sections/*.yaml` — NOT
template-wins layering. The merge base is the manifest-tracked (or prior-render) state; ours
is the on-disk file; theirs is the current template render. Per-key resolution:

- ours changed from base AND theirs did not → **ours wins** (the operator value survives).
  This is the incident clause: `git_strategy.worktree_base_branch: develop` survives a
  template whose own value for that key is `""`, and `manual.workflow: git-flow` survives a
  template default of `github-flow`. The template carries both keys with neutral defaults
  (`git-strategy.yaml.tmpl`: `worktree_base_branch: ""`, `workflow: github-flow`) — the
  2026-09-24 hazard was VALUE reversion, and any literal template-wins layering re-creates it.
- theirs changed from base AND ours did not → theirs wins (the template update lands).
- both changed → conflict → preserve ours + report (the standard REQ-UPM-012 conflict path).
- key absent from the disk file → the template fills it (new template keys land).

The engine is `mrg.YAMLDeep` driven with the base in hand (the same 3-way shape the merge
package already applies to user-customized files), not a fresh implementation. Files the
template does not carry under `sections/` are user-owned (preserve). The template-carries
check applies per section FILE, not per key.

## 5. Summary reporting (REQ-UPM-030/031/032)

Extend `internal/cli/update/report/outcome.go` with reconciliation outcome categories:
`refreshed`, `merged`, `conflicted`, `preserved`, `archived-removed`. Managed files join the
counted set (the `plan.go:73` exclusion is removed with the wipe). Output stays plain counts +
path lists on the existing report structures — no new banner/progress vocabulary (t1527
boundary, C-4).

## 6. Ordering and abort semantics (NFR-UPM-002)

1. Classify the whole tree (read-only; failure → abort, nothing changed).
2. Take the run backup (existing Backup step, unchanged).
3. Reconcile per path in deterministic order; every removal is preceded by its own
   preservation copy; any failure aborts remaining work with the tree intact for
   unprocessed paths (already-processed paths stay in their post-step state — each step is
   individually backed-up/reversible).
4. Deploy template renders (existing deploy machinery, minus the clean call).
5. Emit the extended summary.

## 7. Milestone-to-package map

See plan.md §F. Go packages touched: `internal/cli/update` (orchestration), `internal/cli/update/deploy`
(target-list plumbing only), `internal/cli/update/plan` (count exclusion removal),
`internal/cli/update/report` (outcome categories), `internal/cli/update/backup` (verify
restore-path interplay), possibly `internal/manifest` (read-only; changes only if a gap
surfaces). No template content change is required by this design; if one emerges it follows
Template-First (C-1).

## 8. Alternatives rejected

- **Delete `CleanMoaiManagedPaths` outright** — breaks legacy fresh-install path and
  `WithTargets` callers; the guard test (REQ-UPM-015) covers the retained path instead.
- **Conflict → inline markers in the user file** — mutates user content; violates
  preserve-untouched.
- **Manifest-only classification (no template carriage gate)** — repeats the D-15 loss path;
  manifest can be absent or stale exactly when it is most needed.
- **Diff-based (git) reconciliation** — projects are not guaranteed git repos at update time;
  classification must not depend on git.
