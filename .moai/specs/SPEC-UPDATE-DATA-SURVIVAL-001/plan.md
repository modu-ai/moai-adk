# SPEC-UPDATE-DATA-SURVIVAL-001 — Implementation Plan

> Ordered by decision-reversibility. §B (the failure contract) and §C (the destructive-target
> registry) are the decisions most likely to change under review and are stated first. The
> mechanical work — guard rewrites, error-branch coverage — sits at the bottom.

## §A Context

- **Repository**: `/Users/goos/MoAI/moai-adk-go`, branch `plan/epic-update-config-audit`, worktree
  HEAD `a8b42e112`, whose **code baseline is `d5336214e`** (the `origin/main` merge). Every commit
  between the two changes `.moai/specs/**` documents only —
  `git diff --name-only d5336214e..HEAD | grep -v '^\.moai/'` returns nothing — so no Go source
  differs and every code baseline recorded in `acceptance.md` reproduces at either commit.
- **Tier**: M — 300-1000 LOC across 5-15 files. Justification in §H.
- **Depends on**: `SPEC-UPDATE-REINSTALL-LOOP-002` (E1). Only REQ-RIL2-015/016
  (backup-before-delete for `defs.DeprecatedPaths`) is inherited; every milestone below is
  otherwise independent of E1's merge and can proceed in parallel.
- **Development mode**: TDD (`quality.yaml` `constitution.development_mode`). Each milestone writes
  its guard first, observes it FAIL, then implements.

### A.1 PRESERVE list (do not modify)

- `internal/template/templates/**` — 16-language distribution surface.
- The `defs.DeprecatedPaths` slice body and `internal/defs/dirs_test.go` count assertions. Only the
  Category D **comment prose** changes (M2).
- `internal/cli/update/backup/merge.go` and the YAML merge semantics — owned by
  `SPEC-UPDATE-YAML-PRESERVE-001`.
- `probeVersionSignal`, `detectV2Fingerprint`, `preserveInventoryRoots`, and the `--dry-run` branch
  — owned by E1.
- The success-path outcome of an update (NFR-UDS-005): a run that succeeds today must produce the
  same tree afterwards.

## §B The F5 decision — report and assisted restore, not automatic rollback

**Decision: report + an explicitly-invoked restore command. No automatic rollback.**

### B.1 Rationale

1. **An automatic rollback runs the least-tested code in the worst state.** By the time rollback
   would fire, the tree is in an unknown partial state — some templates deployed, some paths
   removed, `.moai/config` possibly absent. A rollback is itself a multi-step destructive
   filesystem operation (delete the partial deploy, restore N backup trees, re-merge configs). A
   rollback that fails half-way leaves the tree strictly worse than the failure it was recovering
   from. This is the failure shape `runtime-recovery-doctrine.md` §3 invariant 3 names: the
   recovery action triggering the very class of error it exists to recover from. The doctrine's
   prescribed response is the last-resort escape (abort + preserve), not a retry loop.

2. **The backups already exist; only the pointer and the applicator are missing.** Once M3 lands,
   every destroyed file has a disk artifact. What a stranded user lacks is (a) which backup
   directory belongs to the failed run and (b) a command that applies it. Those are cheap, and
   both are outside the failure path.

3. **A user-invoked restore is inspectable and re-runnable.** The user can list the backup
   directory before applying it, apply it twice safely (REQ-UDS-023), and abort. An automatic
   rollback offers none of that and fires exactly when the operator is least able to supervise it.

4. **Cheapest-first (`runtime-recovery-doctrine.md` §2).** Reporting is rung-4 "abort + preserve":
   persist state, name it, stop. Automatic rollback is a heavier rung attempted before the cheaper
   one has been shown insufficient.

### B.2 What "report" concretely means

On any error after the first destructive step, on either path:

- Write a **recovery manifest** into the run-scoped backup directory recording: the failed step
  name, the error, the backup directory's absolute path, and the exact restore command.
- Print the manifest to the user's terminal, not only to a file.
- Return the original error wrapped, so exit status is unchanged.

### B.3 The lockout escape

The restore entry point must run on a tree whose `system.yaml` is gone — that absence is the damage
it repairs. So the project-marker gate at `internal/cli/update.go:236` is bypassed **for the restore
entry point only** (REQ-UDS-022/025). The ordinary update path keeps the gate unchanged, and an AC
asserts that a bare `moai update` on a marker-less tree still fails with the existing message.

Refusal criterion (REQ-UDS-024): the restore entry point accepts a directory only when it carries
the backup's own marker file. It never copies from an arbitrary directory.

### B.4 Rejected alternative

**Transactional rollback via a staged tree** — deploy into a shadow directory and swap on success.
Rejected: the swap is not atomic across the several roots the update touches (`.claude/`,
`.moai/`, repo-root files), cross-device renames are not guaranteed, and the change would rewrite
the entire deploy pipeline — far outside this SPEC's blast radius. Recorded so it is not silently
re-proposed.

## §C The destructive-target registry

The registry is the new shared artifact M2 introduces: a single in-repo table naming every
destructive filesystem operation in the update subsystem and its protection assignment. Below is
the enumeration derived from this tree; the implementation encodes it and guards it.

### C.0 Scan scope — what the guard enumerates, and why it is wider than §C.1's target list

REQ-UDS-007 requires the drift guard to enumerate destructive **call sites** by static source scan,
independently of the registry. That scan is mechanical, so its scope is fixed by grep semantics
rather than by judgement, and it finds materially more than the user-data targets §C.1 tabulates:

```
$ grep -rn 'os\.RemoveAll(\|os\.Rename(' internal/cli/update/ internal/cli/update*.go \
    --include='*.go' | grep -v '_test.go' | wc -l
      17
```

**17 call sites across 10 (file, function) pairs.** An earlier draft of this plan implied the scan
scope was 7 sites (`deploy.go`'s five plus `update_clean_install.go:271` and `update.go:766`); that
figure omitted `update_archive.go`, `update/backup/backup.go`, `update_cleanup.go`, and
`update_namespace_protect.go` entirely. The registry must therefore carry a row per scanned site,
not per user-data target, or the multiset comparison fails on first run.

| # | File | Function | Sites | Class |
|---|---|---|---|---|
| 1 | `update/deploy/deploy.go` | `CleanMoaiManagedPaths` | 3 (`:83`, `:105`, `:121`) | user data — §C.1 rows 1-10 |
| 2 | `update/deploy/deploy.go` | `MigrateLegacyMemoryDir` | 2 (`:169`, `:176`) | user data — §C.1 row 11 |
| 3 | `update_archive.go` | `archiveSkill` | 1 (`:92`) | user data — removes the archive destination before re-archiving |
| 4 | `update_archive.go` | `archiveLegacySkills` | 1 (`:304`) | user data — renames a skill dir into a backup location |
| 5 | `update/backup/backup.go` | `BackupMoaiConfig` | 3 (`:107`, `:135`, `:140`) | **exempt** — unwinds the backup dir this call just created, on its own error paths |
| 6 | `update/backup/backup.go` | `CleanupOldBackups` | 1 (`:259`) | **exempt** — retention pruning of moai-authored backup dirs |
| 7 | `update_clean_install.go` | `runCleanReinstall` | 1 (`:271`) | user data — §C.1 row 12 (E1-owned) |
| 8 | `update_cleanup.go` | `removeDeprecatedFile` | 1 (`:324`) | user data — deprecated-path removal |
| 9 | `update.go` | `ensureGlobalSettingsEnv` | 1 (`:766`) | user data, outside project root — §C.1 row 13 |
| 10 | `update_namespace_protect.go` | `backupUserOwnedNamespace` | 3 (`:225`, `:233`, `:243`) | **exempt** — defensive cleanup of the namespace backup dir this call created (`EC-UNP-007`) |

The three exempt rows (5, 6, 10) remove only directories the same run authored; no user data
predating the run is at risk. Each records that reason in the registry per REQ-UDS-007 rather than
carrying a protection-set assignment. Rows 3, 4, and 8 are destructive sites the original §C.1
enumeration missed; M2 assigns them protection or exemption during implementation, and §C.1 below
is the user-data target view that remains the design rationale.

### C.1 Destructive targets (user-data view)

| # | Target | Site | Protection today | Action |
|---|---|---|---|---|
| 1 | `.claude/settings.json` | `deploy.go:40-42` → `:105` | **in-memory only** (`update_template_sync.go:384-390`) | M3 adds disk backup |
| 2 | `.moai/status_line.sh` | same mergeable set | **in-memory only** | M3 adds disk backup |
| 3 | `.gitignore` | `update_template_sync.go:379-383` | **in-memory only** | M3 adds disk backup |
| 4 | `.claude/commands/moai` | `deploy.go:44-46` → `:105` | template-managed, regenerated | exempt (recorded) |
| 5 | `.claude/agents/moai` | `deploy.go:48-50` → `:105` | template-managed, regenerated | exempt (recorded) |
| 6 | `.claude/skills/moai*` (glob) | `deploy.go:52-55` → `:83` | template-managed | exempt, **with a noted hazard** — see §C.1a |
| 7 | `.claude/rules/moai` | `deploy.go:57-59` → `:105` | template-managed | exempt (recorded) |
| 8 | `.claude/output-styles/moai` | `deploy.go:61-63` → `:105` | template-managed | exempt (recorded) |
| 9 | `.claude/hooks/moai` | `deploy.go:65-67` → `:105` | template-managed | exempt (recorded) |
| 10 | `.moai/config` (wholesale) | `deploy.go:121` | `BackupMoaiConfig` (`backup.go:27`) | covered |
| 11 | `.moai/memory/` | `deploy.go:176` | **none** | M2 adds backup (REQ-UDS-008) |
| 12 | `defs.DeprecatedPaths` incl. `.moai/db` | `update_clean_install.go:271` | **none today** | inherited from E1 REQ-RIL2-015; registry records the cross-SPEC assignment |
| 13 | `~/.claude/hooks/moai` | `update.go:766` | **none**, outside project | M4 pins the radius; backup out of scope (outside project root) |

Rows 1-3, 11, 12, 13 are the complete set of destructive targets in **no** protection set. Of those,
row 12 is E1's; rows 1-3 are M3; row 11 is M2; row 13 is M4 (radius pinning rather than backup,
because a HOME-scoped backup would itself write outside the project).

### C.1a Noted hazard, not a deliverable

Row 6's glob `.claude/skills/moai*` also matches a user-authored `moai-my-notes` skill — precisely
the ambiguous case `collectUserOwnedFilesConservative`
(`update_namespace_protect.go:167`) was introduced to back up. The conservative namespace backup
therefore already covers it, so this SPEC records the interaction in the registry and does **not**
narrow the glob. Narrowing it is a behaviour change belonging to a namespace SPEC.

### C.2 Measured baseline for M6

`mergeBackPreserveInventory` statement coverage was measured on this tree with
`go test -covermode=set -coverprofile=... ./internal/cli/` followed by `go tool cover -func`,
yielding `64.3%` (package total `75.7%`). The figure is recorded verbatim in `acceptance.md`
**AC-UDS-016** as that AC's baseline; the three uncovered blocks are the failure returns at
`update_preserve_inventory.go:346`, `:350`, and `:354`. (Two stale cross-references corrected here:
the milestone is M6, not M7 — §F defines M1-M6 only — and the coverage baseline lives in
AC-UDS-016, not AC-UDS-013, which covers HOME-test isolation.)

## §D Constraints

- `t.TempDir()` only; no test writes to the operator's real `~/.claude`.
- HOME redirection goes through the injectable seam `userHomeDirFn` (`glm_tools.go:123`), not
  `t.Setenv("HOME", …)` — per `CLAUDE.local.md` §13, a process-wide HOME mutation pollutes parallel
  tests. **M4 must first route `ensureGlobalSettingsEnv` through that seam** (REQ-UDS-013): it
  currently calls the plain function `userHomeDir()` at `update.go:756`, which is not reassignable,
  so no redirection point exists today. An earlier draft of this constraint asserted the
  indirection already existed; it does not — see `spec.md` §A Defect 3.
- No `internal/template/templates/**` edit.
- Falsification uses a scratch `git worktree` regenerated from a pre-fix ref and driven with
  `go -C`, or `go test -overlay`. **`git stash` is prohibited** — it refuses untracked files without
  `-u`, and with `-u` it is repository-global and can swallow a concurrent session's work.
- Conventional Commits; `fmt.Errorf("…: %w", err)`; English comments and godoc.

## §E Self-Verification

Every milestone reports the Section E matrix from
`.claude/rules/moai/development/manager-develop-prompt-template.md`: the AC PASS/FAIL table with
verbatim command output, cross-platform build result, coverage, lint delta, and the RED failing
output captured before GREEN.

## §F Milestones

### M1 — Failure contract: recovery manifest + restore entry point

Implements §B. Highest reversibility — the shape of the manifest and the restore command's surface
are the decisions most likely to be revised under review, so they land first and are reviewed
before anything depends on them.

- Recovery-manifest writer invoked from both paths' error returns after the first destructive step.
- Restore entry point accepting a backup directory; marker-file validation (REQ-UDS-024);
  idempotent application (REQ-UDS-023).
- Project-marker gate bypass scoped to the restore entry point (REQ-UDS-022/025).
- Covers REQ-UDS-019 through REQ-UDS-025.

### M2 — Destructive-target registry + `.moai/memory/` backup + comment reconciliation

Introduces the registry of §C.0 as code (10 rows / 17 sites, not the 7 an earlier draft assumed),
its drift guard, the missing `.moai/memory/` backup, and the `dirs.go` brand+db group comment fix.

- The guard enumerates sites by static source scan, independently of the registry (REQ-UDS-007).
  Its independence is proven by the §C.4 falsification in `acceptance.md`, which injects an
  unregistered destructive site and requires an observed `--- FAIL`. A guard that passes that
  procedure is enumerating from the registry and must be rewritten.
- The `dirs.go` change is confined to the `// brand + db directories` group comment. Category D's
  comment is **accurate** and is not touched — AC-UDS-007 guards its `.moai/project/db/` contrast
  clause against deletion. See `spec.md` §A Defect 2's retracted finding.
- Covers REQ-UDS-006 through REQ-UDS-010.

### M3 — On-disk backup for the three in-memory-only files

Writes `.claude/settings.json`, `.moai/status_line.sh`, and `.gitignore` to the run-scoped backup
directory before the first destructive step, on both paths, retaining the in-memory merge-back.

- Includes the **failure path**: a backup write that fails aborts before the first destructive step
  and surfaces the `backup-write-failed:` sentinel naming the file (REQ-UDS-003, AC-UDS-020). This
  path had no AC before the iteration-1 revision; it is the one branch where the SPEC's central
  promise is actually load-bearing, so the abort must be verified by a call-recording spy proving
  `CleanMoaiManagedPaths` was never invoked — not inferred from an unchanged tree.
- Covers REQ-UDS-001 through REQ-UDS-005.

### M4 — HOME deletion-radius pinning

Routes `ensureGlobalSettingsEnv`'s HOME lookup through the injectable `userHomeDirFn` seam, extracts
the removal target to a named symbol, adds the radius guard, and demonstrates its failure against a
widened radius via `go test -overlay`.

- **The seam substitution is a precondition, not a nicety.** `update.go:756` calls the plain
  `userHomeDir()`; with no injection point and `t.Setenv("HOME", …)` forbidden (NFR-UDS-002), the
  radius guard cannot be written at all. Do this first, then the guard.
- Covers REQ-UDS-011 through REQ-UDS-014.

### M5 — Non-vacuous user-area safety guard

Replaces `TestMoaiUpdate_PreservesUserArea` with a guard driving a real production entry point,
deletes `simulateMoaiUpdate`, and demonstrates the replacement failing under a mutation that
touches `.claude/agents/harness/`.

- Covers REQ-UDS-015 through REQ-UDS-018.

### M6 — `mergeBackPreserveInventory` partial-restore reporting and coverage

Adds the restored-count and stopping-file detail to the three failure returns and covers each
branch deterministically.

- Covers REQ-UDS-026 through REQ-UDS-028.

## §G Anti-Patterns

- Re-specifying deprecated-path backup. It is E1's REQ-RIL2-015; duplicating it creates two owners
  for one behaviour.
- Testing the crash window by racing a real crash. Reach it deterministically (NFR-UDS-001).
- Using `git stash` for falsification. Prohibited by §D.
- Widening the marker-gate bypass beyond the restore entry point.
- Asserting the new guards pass without first observing them fail against unfixed code.
- Narrowing the `moai*` glob as a side effect of M2. It is a recorded hazard, not a deliverable.
- Changing the `defs.DeprecatedPaths` entry count while fixing its comment.
- Building the drift guard so it enumerates from the registry. That yields `count == count`, passes
  forever, and reproduces in M2 exactly the vacuous-guard defect §A Defect 3/4 exist to remove.
- Deleting the Category D `.moai/project/db/` contrast prose. It is accurate; the retracted
  iteration-1 finding that called it a mismatch is refuted in `spec.md` §A Defect 2.
- Satisfying AC-UDS-014 by renaming `simulateMoaiUpdate` instead of deleting it and driving real
  production code.

## §H Tier Justification

Tier **M**. Six milestones across roughly ten files: `deploy.go`, `update_template_sync.go`,
`update_clean_install.go`, `update_preserve_inventory.go`, `update.go`, `dirs.go`, plus four to five
new `_test.go` files and one new registry source file. No file exceeds a few hundred lines of
change; the estimate sits inside the 300-1000 LOC / 5-15 file band. It is not Tier S: it spans two
execution paths, adds a new user-facing entry point, and carries a design decision (§B). It is not
Tier L: the design decision is a single binary choice with a recorded rejected alternative, not a
design space warranting `design.md`, and the codebase analysis that would populate `research.md` is
already inlined in §C as a verified enumeration.

## §I Cross-References

- `.moai/specs/SPEC-UPDATE-REINSTALL-LOOP-002/spec.md` — the inherited deprecated-path backup
  requirement and the scratch-worktree falsification precedent (`acceptance.md` §C.2).
- `.claude/rules/moai/workflow/runtime-recovery-doctrine.md` §2, §3 — the cheapest-first ladder and
  the invariant behind §B.1.
- `.claude/rules/moai/development/manager-develop-prompt-template.md` §E — the self-verification
  matrix each milestone reports.
- `CLAUDE.local.md` §13 (HOME in tests), §24 (user-owned namespace contract), §25 (template
  neutrality).
