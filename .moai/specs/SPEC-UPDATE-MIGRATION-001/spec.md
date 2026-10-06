---
id: SPEC-UPDATE-MIGRATION-001
title: "Preservation-based moai update migration for existing projects"
version: "0.1.0"
status: in-progress
created: 2026-10-07
updated: 2026-10-07
author: GOOS (via manager-spec)
priority: P1
phase: "v3.2.0"
module: "internal/cli/update"
lifecycle: spec-anchored
tags: "update, migration, preservation, template, conflict, deploy"
tier: L
---

# SPEC-UPDATE-MIGRATION-001 — Preservation-based moai update migration

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-10-07 | GOOS | Initial draft. Card t1547. Supersedes the withdrawn t1541 `.claude-plugin` distribution-path premise (operator decision 2026-10-06). |

## A. Background

The mods plugin install plan was withdrawn (operator decision 2026-10-06). All installs are
confirmed as **copying into the user folder + the project** — there is no plugin distribution
path, so `moai update` on an EXISTING project is the only mechanism that keeps a project's
`.claude/` / `.agents/` / `.moai/` structure current. This card redefines what that mechanism
means: update is a **migration of the existing structure**, not wipe-and-redeploy.

Today's behavior is destructive by construction. `CleanMoaiManagedPaths`
(`internal/cli/update/deploy/deploy.go:107`) deletes the managed roots wholesale —
`.claude/settings.json`, `.claude/{commands,agents,hooks}/moai`, `.claude/skills/moai*` globs,
`.claude/rules/moai`, `.claude/output-styles/moai`, `.moai/config` — before redeploying only
what the embedded template carries. Measured consequences (`.moai/docs/update-local-file-survival.md`,
card t750; 2026-08-15 incident):

- Local-only files under managed roots are silently destroyed (12 files on 2026-08-15). The
  t111 pre-clean backup catches template-absent files, but the recovery is manual and the
  destruction is still the design.
- The `Updated N files` summary counts only managed-root-OUTSIDE files
  (`internal/cli/update/plan/plan.go:73` skips `IsMoaiManaged` paths) and never shows
  deletions (reported 32, actual 175 on 2026-08-15).
- `git-strategy.yaml` local keys (`workflow: git-flow`, `worktree_base_branch: develop`) do not
  exist in the template default, so the wipe-and-redeploy reverts them on every update; the
  2026-09-24 incident cut 6 card worktrees from `main` instead of `develop` (card t1159) before
  the manual re-apply rule existed.

The new principle (operator decision 2026-10-06): update must **preserve user-owned/local
content, refresh template-managed content, and report instead of overwrite where they
collide** — the same conflict semantics t1509 installs under on the install side
("user-created same-name files are not overwritten, only reported").

The codebase already carries most of the machinery this SPEC wires together:

- `internal/manifest` — per-file provenance (`template_managed` / `user_modified` /
  `user_created`) with `CurrentHash`, the discriminator between refreshed and user-modified.
- `internal/cli/update/migrate_classify.go` — the identical/modified/foreign three-way
  classifier with template-carriage as the class gate (SPEC-INIT-SHRINK-001).
- `internal/cli/update/merge` + `internal/merge` — the 3-way merge engine with per-file
  strategy selection (`plan.DetermineStrategy`: CLAUDE.md section merge, .gitignore entry
  merge, JSON merge, YAML deep, line merge).
- `internal/cli/update/plan.IsUserOwnedNamespace` — the user-owned namespace predicate.
- `internal/cli/update/deploy` — the t111 pre-clean backup and the REQ-UDS-008
  backup-before-destroy ordering.

What is missing is the wiring: the update-on-existing-project flow still runs the wholesale
wipe first and consults none of the classification when deciding what to remove.

## B. Requirements (GEARS)

REQ numbers are block-scoped per subsection (B.1: 001-004, B.2: 010-016, B.3: 020-021,
B.4: 030-032, B.5: 040); the gaps between blocks are reserved ranges, not omissions.

### B.1 Ownership classification

- REQ-UPM-001: The update flow shall classify every regular file under the managed roots into
  exactly one of four classes BEFORE any removal or overwrite: `template-owned` (template
  carries the path and the manifest record is healthy with content equal to the tracked
  state), `user-modified` (template carries the path but the content diverges from the
  tracked state, or the record is absent or stale), `user-owned` (the template does not carry
  the path), and `stale` (a prior template carried the path and the current template does
  not).
- REQ-UPM-002: The classification gate shall be template carriage — managed-name matching or
  the clean-target list alone shall never route a file into a removal-eligible class
  (the D-15 loss path).
- REQ-UPM-003: The update flow shall treat `IsUserOwnedNamespace` paths as `user-owned`
  regardless of any manifest state.
- REQ-UPM-004: The classifier shall use Lstat semantics for every entry it walks and shall
  record symlinks in the classification output without dereferencing them (the
  REQ-CSL-001/REQ-013 rule carried forward).

### B.2 Reconciliation semantics per class

- REQ-UPM-010: The update flow shall refresh `template-owned` files in place from the current
  template render without archiving them (the template is the recovery source).
- REQ-UPM-011: **While** a `user-modified` file has a mergeable format (CLAUDE.md, .gitignore,
  JSON, YAML), the update flow shall attempt a 3-way merge using the existing merge engine and
  strategy selection, and shall write the merged result only when the merge completes without
  conflict.
- REQ-UPM-012: **When** a merge conflict is detected in a `user-modified` file, the update
  flow shall preserve the on-disk file byte-for-byte, shall write the template render to a
  sibling `<path>.moai-new` file, and shall report the conflict in the update summary — never
  a silent overwrite and never a silent delete. **While** that sibling name is already
  occupied — a leftover sidecar from a prior unresolved conflict, or an independently created
  user file of that name — the update flow shall write the first unused numbered sibling
  (`<path>.moai-new.2`, `<path>.moai-new.3`, …), shall report the collision in the update
  summary, and shall never overwrite the existing sibling.
- REQ-UPM-013: The update flow shall preserve `user-owned` files untouched — no backup, no
  archive, no rewrite — and shall list each preserved path in the update summary.
- REQ-UPM-014: The update flow shall archive-then-remove `stale` files (a prior template
  carried the path; the current one does not) into the migration archive root
  (`.moai/archive/files/<tag>/`, the existing `ArchiveFilesRoot()` layout; copies are written
  with the package default file mode — the archive is a recovery surface, not a
  confidentiality boundary, and this SPEC defines no retention/cleanup policy for it), and
  shall list each removal in the update summary.
- REQ-UPM-015: The update flow shall not delete any file under a managed root whose class is
  `user-owned` or `user-modified-unresolved`, in any code path, including the wholesale
  cleanup path retained for legacy flows.
- REQ-UPM-016: **When** any removal or overwrite is about to run, the update flow shall have
  completed its preservation copy first, and a backup failure shall abort before the
  destruction it was taken to survive (the REQ-UDS-008 rule, carried forward unchanged).

### B.3 The `.moai/config` case

- REQ-UPM-020: The update flow shall not delete `.moai/config/` wholesale on an
  existing-project update; it shall deep-merge each `sections/*.yaml` file so that
  template-carried keys are updated to current template values and user-added keys are
  preserved verbatim.
- REQ-UPM-021: **When** `.moai/config/sections/git-strategy.yaml` carries operator-set values
  for keys the template also carries with neutral defaults (`git_strategy.worktree_base_branch:
  develop` where the template default is empty; `git_strategy.manual.workflow: git-flow` where
  the template default is `github-flow`), the update flow shall preserve the operator-set
  values across the update, so that no post-update manual re-apply step is required.

### B.4 Summary reporting

- REQ-UPM-030: The update summary shall report, with per-path listings: refreshed files,
  merged files, conflicts (preserved local + `.moai-new` sidecar), preserved user-owned
  files, and archived-removed stale files.
- REQ-UPM-031: The update summary shall include deletions of managed files; a run that
  removes N managed files shall state so — the summary shall never report a deletion-free run
  while the removal machinery deleted files.
- REQ-UPM-032: The summary file-count fix shall not paint update output in a form that
  pre-commits the t1527 UX overhaul (banners, progress, severity prefixes): the reconciliation
  outcome rides the existing report/outcome structure with plain counts and path lists.

### B.5 Legacy-path guard

- REQ-UPM-040: **Where** the legacy v1→v2 incompatible-config path still requires a fresh
  config install, the update flow shall perform the wholesale `.moai/config` removal only
  after a full pre-clean backup of the directory, and shall state the fresh-install reason in
  the progress output.

## C. Success Criteria (Tier L — see acceptance.md for the full matrix)

1. Characterization suite first: the current `CleanMoaiManagedPaths` behavior is pinned by
   tests BEFORE the pipeline changes, so the fix is provable against the deletion hazard.
2. The hazard test: a local-only file under a managed root survives an update over a fixture
   project, byte-for-byte, with no manual restore.
3. The regression test for the 2026-09-24 incident: user-added `git-strategy.yaml` keys
   survive an update with zero manual re-apply.
4. `git status --porcelain | grep '^ D'` shows 0 deletions of tracked local-only files after
   an update over a fixture project (the §2.3 manual check, automated in acceptance).
5. Affected packages (`internal/cli/update/...`, `internal/manifest` if touched) at or above
   the 85% package coverage target, existing tests unmodified in intent.

## D. Constraints

- C-1: Template-First (AGENTS.local.md §2) — any template-side change lands in
  `internal/template/templates/` first, then `make build`; catalog hash regeneration rides
  `make build` / `make agents-emit`. No template content in this SPEC carries internal card
  ids, dates, or repo-private workflow values (§25 neutrality).
- C-2: The `internal/cli` boundary (CLAUDE.md, internal/cli) — exit-code discipline, stdout
  machine-readable / stderr progress separation, no interactive prompts (no
  AskUserQuestion-shaped code in the CLI).
- C-3: Cross-platform paths — `filepath` everywhere; no POSIX-only assumptions; verify with a
  windows GOOS build.
- C-4: No scope creep into t1509 (user-folder install: `~/.claude`, `~/.agents`, profile
  dirs) or t1527 (init/update UX overhaul). This SPEC touches the project-side update flow
  only.
- C-5: The deploy package stays a leaf (no new root-package imports).
- C-6: Tests use `t.TempDir()` isolation; no OTEL env vars in tests; no local full-suite runs
  (`go test ./internal/cli/update/...` scoped per §4 of AGENTS.local.md).
- C-7: Protected directories (AGENTS.local.md §2: `.claude/`, `.moai/project/`, `.moai/specs/`
  are never deleted during template sync) sit OUTSIDE the managed-root set and shall never
  enter the clean-target set; any future clean-target expansion keeps them out.

## E. Non-Functional Requirements

- NFR-UPM-001: Classification is read-only and deterministic — two runs over the same tree
  and template set produce byte-identical classification output.
- NFR-UPM-002: A run aborted mid-reconciliation leaves the project in its pre-run state for
  every not-yet-processed path (ordering: classify all → preserve/merge/archive per path →
  refresh; each destructive step is preceded by its own backup).
- NFR-UPM-003: Reconciliation of a clean project (no user modifications, no local-only files)
  produces the same end state as the current wipe-and-redeploy — the preservation pipeline is
  additive, not a behavior change for clean trees.

## F. Dependencies

- `internal/manifest` (provenance + hash) — consumed, not modified unless a gap is found.
- `internal/merge` engine, `internal/cli/update/{plan,merge,deploy,report}` — modified/wired.
- Related SPECs: SPEC-INIT-SHRINK-001 (classifier origin), SPEC-CLI-CLEAN-SYMLINK-001 (Lstat
  semantics), SPEC-UPDATE-DATA-SURVIVAL-001 (REQ-UDS-008 backup ordering), SPEC-CLI-TUX-V3-003
  (package decomposition). Boundary SPECs: t1509 (install side, queued), t1527 (UX, queued).

## G. Risks

- R-1: Manifest-absent projects (never tracked) classify conservatively as `user-modified` —
  every template-carried file becomes a merge candidate. Mitigation: merge with
  last-deploy-as-base fallback (template base when no manifest base exists); a clean merge of
  identical content is a refresh.
- R-2: `.moai-new` sidecar accumulation on never-resolved conflicts. Accepted for this SPEC:
  conflicts are rare, reported, and operator-resolved; no auto-resolution is attempted.
- R-3: The wholesale cleanup path is retained for legacy flows (REQ-UPM-040) — a divergence
  between the two paths could reintroduce the hazard. Mitigation: REQ-UPM-015 makes the
  guard test assert on both paths.

## H. Out of Scope

### Out of Scope — user-folder install (t1509)
- The `~/.claude`, `~/.agents`, and profile-directory install flow (the install side of the
  copy-only decision) is card t1509's scope. This SPEC notes the boundary only.

### Out of Scope — update/init UX overhaul (t1527)
- Banners, progress rendering, and severity prefixes are t1527's scope. This SPEC keeps
  reporting to plain counts and path lists inside the existing report structures.

### Out of Scope — plugin distribution path
- The t1541 `.claude-plugin` distribution premise is voided (operator decision 2026-10-06);
  no plugin-path work is in this SPEC.

### Out of Scope — template content changes
- New or redesigned template files are not in scope; this SPEC changes how the update flow
  reconciles whatever the template carries.
