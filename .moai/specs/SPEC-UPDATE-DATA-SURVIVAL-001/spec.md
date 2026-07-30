---
id: SPEC-UPDATE-DATA-SURVIVAL-001
title: "moai update — user-data survival: on-disk backup before every destructive step, a failure contract with an escape from the marker-gate lockout, and non-vacuous safety guards"
version: "0.1.0"
status: draft
created: 2026-07-31
updated: 2026-07-31
author: manager-spec
priority: P0
phase: "v3.0.2"
module: "internal/cli"
lifecycle: spec-anchored
era: V3R6
tier: M
tags: "cli, update, backup, rollback, data-loss, crash-window, safety-test, mutation-testing, regression"
related_specs: [SPEC-UPDATE-REINSTALL-LOOP-002, SPEC-UPDATE-YAML-PRESERVE-001, SPEC-V3R6-UPDATE-NAMESPACE-PROTECT-001, SPEC-V3R6-UPDATE-ARCHIVE-CONTRACT-001]
depends_on: [SPEC-UPDATE-REINSTALL-LOOP-002]
---

# SPEC-UPDATE-DATA-SURVIVAL-001

## HISTORY

| Version | Date | Change |
|---------|------|--------|
| 0.1.0 | 2026-07-31 | Initial draft. Four-lens audit of `moai update` / `.moai/config`; Epic SPEC 2 of 6. |

## §A Problem / Motivation

`moai update` destroys and redeploys large parts of a user's project tree. The destruction is
deliberate; the problem is that several destructive steps have **no artifact on disk** to restore
from, and there is **no defined behaviour** when a run dies between destruction and restoration.

Six defects, each observed against this tree.

### Defect 1 — `.claude/settings.json` is destroyed with only an in-memory backup

The normal path's first destructive step is `deploy.CleanMoaiManagedPaths`
(`internal/cli/update/deploy/deploy.go:28`), whose first `cleanTarget` entry is
`.claude/settings.json` (`deploy.go:40-42`, removed in the loop at `deploy.go:105`). Its only
backup is a `[]byte` captured into an `updatemerge.FileBackup` slice at
`internal/cli/update_template_sync.go:384-390` and re-applied by `MergeUserFiles` at
`update_template_sync.go:426-430`. The clean-reinstall path builds the identical in-memory slice at
`internal/cli/update_clean_install.go:312-317`.

Observed after running the two backup steps followed by the clean step on a fixture tree:

```
live .claude/settings.json              exists=false
backup artifact: .moai-backups/<ts>/sections/user.yaml
backup artifact: .moai-backups/<ts>/backup_metadata.json
   (plus .template-defaults/sections/*.yaml — no settings.json entry)
```

`BackupMoaiConfig` (`internal/cli/update/backup/backup.go:27`) walks `.moai/config` only
(`backup.go:33-34`), so it cannot cover a `.claude/` path by construction. A grep of the backup
package for any disk write of `settings.json` returns nothing.

Between the `RemoveAll` and the merge-back, the user's `permissions.allow`, `env`, and
`outputStyle` exist **only in the process's heap**. A crash, an ENOSPC, a SIGKILL, or Ctrl+C in
that window loses them permanently with no recovery path.

Two further files share the identical defect shape and the identical window:
`.moai/status_line.sh` (same `mergeableBackups` slice) and `.gitignore` (the `gitignoreBackup`
`[]byte` at `update_template_sync.go:379-383`).

### Defect 2 — destructive targets that appear in no protection set

Three protection sets exist:

| Set | Covers | Defined at |
|---|---|---|
| PRESERVE inventory | `.moai/specs`, `.moai/project`, `.claude/commands` | `update_preserve_inventory.go:66-70` |
| `BackupMoaiConfig` | `.moai/config` only | `backup/backup.go:27-34` |
| Namespace backup | `.claude/skills`, `.claude/agents`, `.moai/harness` | `update_namespace_protect.go:39-43` |

`MigrateLegacyMemoryDir` (`deploy.go:143`) removes `.moai/memory/` outright when both the legacy
and the new directory exist (`deploy.go:176`). That path is in none of the three sets and has no
backup of any kind.

`.moai/db` is a **`defs.DeprecatedPaths` entry** (`internal/defs/dirs.go:313`, under the
`brand + db directories` grouping) rather than a bespoke removal, so it is deleted by the generic
deprecated-path loop at `update_clean_install.go:271`. It is likewise in none of the three
protection sets — but its backup is already required by `SPEC-UPDATE-REINSTALL-LOOP-002`
REQ-RIL2-015 (backup before every deprecated-path deletion), so this SPEC depends on that
requirement rather than restating it.

Separately, the `defs.DeprecatedPaths` Category D comment (`dirs.go:344-351`) states that
"the `.moai/project/db/` docs (a preserve root, manual deletion per CHANGELOG)" are excluded from
active removal. No `.moai/project/db` entry is registered; the registered directory entry is
`.moai/db` (`dirs.go:313`). Comment and code name different paths, so a reader auditing db-removal
coverage is pointed at the wrong path.

### Defect 3 — the deletion radius of `~/.claude/hooks/moai` is unpinned

`ensureGlobalSettingsEnv` (`internal/cli/update.go:733`) removes a directory inside the user's real
HOME:

```go
globalHooksDir := filepath.Join(homeDir, defs.ClaudeDir, "hooks", "moai")   // update.go:742
if _, err := os.Stat(globalHooksDir); err == nil {
    _ = os.RemoveAll(globalHooksDir)                                        // update.go:744
}
```

A mutation lens widened the target from `~/.claude/hooks/moai` to `~/.claude/hooks` — deleting the
user's non-MoAI global hooks — and ran the suite through an overlay:

```
$ go test -overlay=overlay5.json ./internal/cli/ -run 'GlobalSettings|Update|Hook'
ok  github.com/modu-ai/moai-adk/internal/cli  18.305s
```

Everything passed. Re-verified on this tree: `grep -rn 'globalHooksDir' internal/cli/ --include='*_test.go'`
returns **0** lines — the identifier appears only at `update.go:742-744`. This deletion is outside
the project root, unbacked, and unrecoverable, and nothing detects a widening of its radius.

### Defect 4 — `TestMoaiUpdate_PreservesUserArea` executes zero production code

`internal/cli/update_safety_test.go:59` claims to enforce that `moai update` does not touch user
areas. It calls `simulateMoaiUpdate` (`update_safety_test.go:43-57`), a fake defined in the same
test file whose entire body writes one managed file under `.claude/skills/moai/` and comments
`// Note: we intentionally do NOT touch user areas.` The assertions at `update_safety_test.go:95-102`
compare pre/post hashes of three user directories that the fake never visits, so they hold by
construction.

Re-verified on this tree: a grep of that file for any production symbol
(`runUpdate|RunUpdate|CleanMoaiManagedPaths|BackupMoaiConfig|runCleanReinstall|buildPreserveInventory|deploy\.|backup\.`)
returns no matches. The `CLAUDE.local.md` §24 contract — `moai update` never deletes user-authored
harness agents or `hns-*` skills — therefore has no executable defence.

### Defect 5 — no failure contract, and the repair tool locks itself out

Normal path: `CleanMoaiManagedPaths` (invoked at `update_template_sync.go:243-247`) removes
`.moai/config` wholesale (`deploy.go:121`); when the subsequent `Deploy Templates` step
(`update_template_sync.go:249`) fails, the function returns and the `"Restore Settings"` case at
`update_template_sync.go:391` is never reached. Clean path: a Step 5 failure returns before
Steps 5.5 and 6.

The backups survive on disk, but nothing tells the user which backup belongs to the failed run,
and no command applies one — the only restore entry points are `backup.RestoreMoaiConfig`
(`backup/restore.go:40`, called mid-run only) and `moai migrate restore-skill`
(`migrate_restore_skill.go:90`, a different subsystem).

Worse, the failure is self-sealing. Once `.moai/config/sections/system.yaml` is gone, the project
marker check at `internal/cli/update.go:236` rejects the next invocation:

```
not a moai project: .moai/config/sections/system.yaml not found in the current directory
```

The tool that would repair the damage refuses to run on the damaged tree.

### Defect 6 — `mergeBackPreserveInventory` partial-restore branches are uncovered

`mergeBackPreserveInventory` (`internal/cli/update_preserve_inventory.go:330`) restores the PRESERVE
inventory file-by-file. Its three failure returns — the non-`ErrNotExist` stat error
(`:346`), the `MkdirAll` error (`:350`), and the `copyFile` error (`:354`) — each return
immediately from inside the loop. A failure at file *k* leaves files `0..k-1` restored, files
`k..n` absent, and no report of where the boundary fell. The measured statement coverage of the
function is recorded in `plan.md §C`; the three error branches are the uncovered portion.

## §B Goals

- No destructive step runs before the data it destroys exists on disk outside the process.
- Every destructive target is either covered by a named protection set or explicitly and
  deliberately exempt, with the exemption recorded.
- The HOME-scoped deletion cannot widen without a test failing.
- The user-area safety contract is defended by a test that drives real production code.
- A run that dies mid-flight leaves the user a named backup, a stated tree state, and a command
  that restores it — and that command works on a tree whose project marker was destroyed.
- A partial restore reports precisely which files were restored and which were not.

## §C Scope Exclusions

### Out of Scope — v2/v3 detection, PRESERVE∩Deprecated, dry-run

- Semver-normalized v2/v3 version detection, the `preserveInventoryRoots` ∩ `defs.DeprecatedPaths`
  exclusion, the residue-cleanup path, and `--dry-run` reachability. All owned by
  `SPEC-UPDATE-REINSTALL-LOOP-002`.
- **Backup-before-delete for `defs.DeprecatedPaths` entries** (including `.moai/db`) is owned by
  that SPEC's REQ-RIL2-015/016. This SPEC declares `depends_on: [SPEC-UPDATE-REINSTALL-LOOP-002]`
  and treats deprecated-path backup as an inherited precondition, never re-specifying it. The
  dependency is load-bearing in the reverse direction too: E1's PRESERVE-exclusion fix removes the
  incidental protection that `.claude/commands/agency/*.md` enjoys today, which is why E1 binds
  `backupDeprecatedPaths` to its own milestone.

### Out of Scope — YAML merge fidelity

- Comment, key-order, and scalar-style preservation across `.moai/config/sections/*.yaml` merges,
  and the old-only-key drop. Owned by `SPEC-UPDATE-YAML-PRESERVE-001`. This SPEC cares that the
  bytes reach disk before deletion, not how they are later merged.

### Out of Scope — merge tiers, dead keys, CI gates, doc drift

- Merge-tier semantics, dead `.moai/config` keys, CI gate additions, and documentation drift.
  Owned by Epic SPECs 3 through 6.

### Out of Scope — DeprecatedPaths catalogue membership

- Adding, removing, or re-categorising entries in `defs.DeprecatedPaths`. This SPEC amends only the
  Category D **comment prose** so it names the registered path; the 40-entry slice is untouched and
  its count invariant (`internal/defs/dirs_test.go`) is preserved.

### Out of Scope — template content

- Any edit under `internal/template/templates/**`. Fixtures live in `_test.go` files only.

### Out of Scope — transactional filesystem rollback

- An automatic, atomic, whole-tree rollback of a failed update. §D.5 records the decision to
  provide an explicitly-invoked restore command instead, with rationale.

## §D Requirements (GEARS)

### D.1 On-disk backup before destruction

- **REQ-UDS-001** — **When** the update reaches its first destructive step on either path, every
  file whose sole backup is currently in-process memory (`.claude/settings.json`,
  `.moai/status_line.sh`, `.gitignore`) shall already exist as a byte-identical copy under a
  backup directory on disk.
- **REQ-UDS-002** — The on-disk backup shall be written under a single run-scoped backup directory
  whose absolute path is reported to the caller, so a later restore can name it.
- **REQ-UDS-003** — **When** writing any of those on-disk backups fails, the update shall abort
  before the first destructive step and shall surface a grep-able sentinel naming the file that
  could not be backed up.
- **REQ-UDS-004** — The in-memory merge-back path shall be retained unchanged; the on-disk copy is
  additive insurance for the crash window, not a replacement for the merge.
- **REQ-UDS-005** — The normal path and the clean-reinstall path shall use the same on-disk backup
  mechanism, so neither path can drift into being the unprotected one.

### D.2 Destructive-target coverage

- **REQ-UDS-006** — Every destructive filesystem operation in the update subsystem shall be
  enumerated in a single in-repo registry naming, per target, the protection set that covers it or
  the recorded reason it is exempt.
- **REQ-UDS-007** — A regression guard shall assert that every destructive target in the registry
  carries either a protection-set assignment or an explicit exemption, and shall fail when a target
  carries neither.
- **REQ-UDS-008** — **When** `MigrateLegacyMemoryDir` removes `.moai/memory/` because both the
  legacy and the new directory exist, it shall back up `.moai/memory/` before removal.
- **REQ-UDS-009** — The `defs.DeprecatedPaths` Category D comment shall name the path that is
  actually registered, so that comment and code agree.
- **REQ-UDS-010** — The registry shall record `defs.DeprecatedPaths` deletion as covered by
  `SPEC-UPDATE-REINSTALL-LOOP-002` REQ-RIL2-015 rather than duplicating that coverage.

### D.3 Deletion-radius pinning

- **REQ-UDS-011** — The HOME-scoped removal in `ensureGlobalSettingsEnv` shall be covered by a test
  that fails when the removal target is widened to any ancestor of `~/.claude/hooks/moai`.
- **REQ-UDS-012** — **While** that test runs, it shall not read from or write to the operator's real
  home directory.
- **REQ-UDS-013** — The removal target shall be derived from a single named symbol so the guard has
  one thing to assert against, rather than re-deriving the path independently.
- **REQ-UDS-014** — The guard shall be shown to FAIL against a deliberately widened radius before it
  is accepted.

### D.4 Non-vacuous user-area safety

- **REQ-UDS-015** — The user-area preservation guard shall drive a real production entry point of
  the update subsystem, not a fake defined in the test file.
- **REQ-UDS-016** — The replacement guard shall assert that user-owned harness agents and `hns-*`
  skills are byte-identical before and after that entry point runs.
- **REQ-UDS-017** — The replacement guard shall be shown to FAIL when the production code under it
  is mutated to touch a user-owned path.
- **REQ-UDS-018** — The `simulateMoaiUpdate` fake shall be removed, so no later reader mistakes it
  for coverage.

### D.5 Failure contract and lockout escape

- **REQ-UDS-019** — **When** any step of either update path returns an error after the first
  destructive step has run, the update shall write a recovery manifest recording the failed step,
  the run-scoped backup directory, and the restore command, and shall print that manifest.
- **REQ-UDS-020** — The update shall not attempt an automatic rollback of the tree. The decision and
  its rationale are recorded in `plan.md §B`.
- **REQ-UDS-021** — A restore entry point shall accept a backup directory and apply it to the
  project tree.
- **REQ-UDS-022** — **When** the restore entry point runs on a project whose
  `.moai/config/sections/system.yaml` is absent, it shall proceed rather than reject the tree,
  because that absence is the damage it exists to repair.
- **REQ-UDS-023** — The restore entry point shall be idempotent: running it twice against the same
  backup directory shall leave the same tree state as running it once.
- **REQ-UDS-024** — The restore entry point shall refuse a backup directory that is not a
  recognisable update backup, rather than copying arbitrary content into the project.
- **REQ-UDS-025** — The project-marker gate at the update entry point shall remain in force for the
  ordinary update path; only the restore entry point is exempt.

### D.6 Partial-restore reporting

- **REQ-UDS-026** — **When** `mergeBackPreserveInventory` fails part-way through the inventory, its
  error shall name the file at which it stopped and the count of files already restored.
- **REQ-UDS-027** — The three failure returns of `mergeBackPreserveInventory` shall each be covered
  by a test that reaches them deterministically.
- **REQ-UDS-028** — The stat, `MkdirAll`, and `copyFile` failure branches shall be distinguishable
  from one another in the returned error.

## §E Non-Functional Constraints

- **NFR-UDS-001** — Crash-window defects are probabilistic on the success path. Tests shall reach
  the failure window deterministically by invoking the backup step and the destructive step as
  separate calls and asserting on the filesystem between them, or by injecting a failing step —
  never by racing a real crash.
- **NFR-UDS-002** — All tests shall use `t.TempDir()`. A test exercising HOME-scoped behaviour shall
  redirect the home lookup through the existing `userHomeDir` indirection rather than mutating the
  process environment, so parallel tests cannot observe each other's HOME.
- **NFR-UDS-003** — No file under `internal/template/templates/**` shall be added or modified.
- **NFR-UDS-004** — Go conventions: `snake_case.go` filenames, `fmt.Errorf("…: %w", err)` wrapping,
  English comments and godoc.
- **NFR-UDS-005** — Added backup writes shall not change the success-path outcome of an update: a
  run that succeeds today shall produce the same tree after the change.
- **NFR-UDS-006** — Path handling shall remain correct on linux, darwin, and windows.
- **NFR-UDS-007** — The `defs.DeprecatedPaths` entry count and its guarding test shall be unchanged.

## §F Success Criteria

- After the two backup steps and before the clean step, `.claude/settings.json`,
  `.moai/status_line.sh`, and `.gitignore` each exist under the run-scoped backup directory —
  verified by an executed test that inspects the filesystem between the two calls.
- The destructive-target registry lists every `RemoveAll` / `Rename` site in the update subsystem,
  and its guard fails when a new site is added without a protection assignment.
- Widening the HOME-scoped removal to `~/.claude/hooks` makes the suite fail.
- The replacement user-area guard fails when production code is mutated to touch
  `.claude/agents/harness/`.
- A tree whose `system.yaml` was destroyed can be restored by the restore entry point.
- All three `mergeBackPreserveInventory` failure branches are exercised, and each error names its
  own cause.
- `go build ./...`, `GOOS=windows GOARCH=amd64 go build ./...`, and `go test ./internal/cli/...` pass.

## §G Risk Register

| Risk | Direction | Bound |
|---|---|---|
| On-disk backup writes slow every update or fill the disk | Performance / capacity regression | Backup set is three small files; the existing `CleanupOldBackups` retention applies to the same backup root |
| A restore command becomes a second destructive path | Data loss | REQ-UDS-024 refuses unrecognised directories; REQ-UDS-023 makes it idempotent; it never deletes, only writes |
| Marker-gate exemption widens into a general bypass | Safety regression | REQ-UDS-025 scopes the exemption to the restore entry point alone, asserted by an AC |
| Replacement safety guard is itself vacuous | Undetected regression | REQ-UDS-017 requires a demonstrated mutation failure via `go test -overlay` |
| Deletion-radius guard passes for the wrong reason | False confidence | REQ-UDS-014 requires a demonstrated failure against a widened radius |
| Registry drifts as new destructive sites are added | Silent coverage gap | REQ-UDS-007 makes the guard fail on an unassigned target |
| Depending on E1 blocks this SPEC | Schedule coupling | Only the deprecated-path backup is inherited; every other milestone here is independent of E1's merge |

## §H Cross-References

- `internal/cli/update/deploy/deploy.go` — `CleanMoaiManagedPaths` (`:28`), the `cleanTarget` list
  (`:39-73`), the removal loop (`:83`, `:105`), the `.moai/config` wipe (`:121`),
  `MigrateLegacyMemoryDir` (`:143`, `:169`, `:176`)
- `internal/cli/update_template_sync.go` — `gitignoreBackup` (`:379-383`), `mergeableBackups`
  (`:384-390`), `Clean Managed Paths` (`:243-247`), `Restore Settings` (`:391`),
  `MergeUserFiles` (`:426-430`)
- `internal/cli/update_clean_install.go` — Step 4 deprecated-path removal (`:265-275`), the
  in-memory mergeable set (`:312-317`), `BackupMoaiConfig` call (`:304`)
- `internal/cli/update_preserve_inventory.go` — `preserveInventoryRoots` (`:66-70`),
  `mergeBackPreserveInventory` (`:330`, failure returns `:346`/`:350`/`:354`)
- `internal/cli/update_namespace_protect.go` — `userOwnedScanRoots` (`:39-43`),
  `backupUserOwnedNamespace` (`:196`), `verifyNamespaceBackupCoverage` (`:293`)
- `internal/cli/update/backup/backup.go` — `BackupMoaiConfig` (`:27`), `CleanupOldBackups` (`:206`)
- `internal/cli/update.go` — project-marker gate (`:236`), `ensureGlobalSettingsEnv` (`:733`),
  `globalHooksDir` (`:742-744`)
- `internal/cli/update_safety_test.go` — `simulateMoaiUpdate` (`:43-57`),
  `TestMoaiUpdate_PreservesUserArea` (`:59`), tautological assertions (`:95-102`)
- `internal/defs/dirs.go` — `.moai/db` entry (`:313`), Category D comment (`:344-351`)
- `.moai/specs/SPEC-UPDATE-REINSTALL-LOOP-002/spec.md` — REQ-RIL2-015/016 (deprecated-path backup)
- `CLAUDE.local.md` §24 — the user-owned namespace contract Defect 4 leaves undefended
