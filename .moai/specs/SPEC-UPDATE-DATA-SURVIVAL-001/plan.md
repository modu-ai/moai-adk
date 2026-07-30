# SPEC-UPDATE-DATA-SURVIVAL-001 — Implementation Plan

> Ordered by decision-reversibility. §B (the failure contract) and §C (the destructive-target
> registry) are the decisions most likely to change under review and are stated first. The
> mechanical work — guard rewrites, error-branch coverage — sits at the bottom.

## §A Context

- **Repository**: `/Users/goos/MoAI/moai-adk-go`, branch `main`, HEAD `1d4e4f7da` at authoring time.
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

| # | Target | Site | Protection today | Action |
|---|---|---|---|---|
| 1 | `.claude/settings.json` | `deploy.go:40-42` → `:105` | **in-memory only** (`update_template_sync.go:384-390`) | M3 adds disk backup |
| 2 | `.moai/status_line.sh` | same mergeable set | **in-memory only** | M3 adds disk backup |
| 3 | `.gitignore` | `update_template_sync.go:379-383` | **in-memory only** | M3 adds disk backup |
| 4 | `.claude/commands/moai` | `deploy.go:44-46` → `:105` | template-managed, regenerated | exempt (recorded) |
| 5 | `.claude/agents/moai` | `deploy.go:48-50` → `:105` | template-managed, regenerated | exempt (recorded) |
| 6 | `.claude/skills/moai*` (glob) | `deploy.go:52-55` → `:83` | template-managed | exempt, **with a noted hazard** — see §C.1 |
| 7 | `.claude/rules/moai` | `deploy.go:57-59` → `:105` | template-managed | exempt (recorded) |
| 8 | `.claude/output-styles/moai` | `deploy.go:61-63` → `:105` | template-managed | exempt (recorded) |
| 9 | `.claude/hooks/moai` | `deploy.go:65-67` → `:105` | template-managed | exempt (recorded) |
| 10 | `.moai/config` (wholesale) | `deploy.go:121` | `BackupMoaiConfig` (`backup.go:27`) | covered |
| 11 | `.moai/memory/` | `deploy.go:176` | **none** | M2 adds backup (REQ-UDS-008) |
| 12 | `defs.DeprecatedPaths` incl. `.moai/db` | `update_clean_install.go:271` | **none today** | inherited from E1 REQ-RIL2-015; registry records the cross-SPEC assignment |
| 13 | `~/.claude/hooks/moai` | `update.go:744` | **none**, outside project | M4 pins the radius; backup out of scope (outside project root) |

Rows 1-3, 11, 12, 13 are the complete set of destructive targets in **no** protection set. Of those,
row 12 is E1's; rows 1-3 are M3; row 11 is M2; row 13 is M4 (radius pinning rather than backup,
because a HOME-scoped backup would itself write outside the project).

### C.1 Noted hazard, not a deliverable

Row 6's glob `.claude/skills/moai*` also matches a user-authored `moai-my-notes` skill — precisely
the ambiguous case `collectUserOwnedFilesConservative`
(`update_namespace_protect.go:167`) was introduced to back up. The conservative namespace backup
therefore already covers it, so this SPEC records the interaction in the registry and does **not**
narrow the glob. Narrowing it is a behaviour change belonging to a namespace SPEC.

### C.2 Measured baseline for M7

`mergeBackPreserveInventory` statement coverage was measured on this tree with
`go test -covermode=set -coverprofile=... ./internal/cli/` followed by `go tool cover -func`. The
observed figure is recorded verbatim in `acceptance.md` AC-UDS-013 as that AC's baseline; the three
uncovered blocks are the failure returns at `update_preserve_inventory.go:346`, `:350`, and `:354`.

## §D Constraints

- `t.TempDir()` only; no test writes to the operator's real `~/.claude`.
- HOME redirection goes through the existing `userHomeDir` indirection (`update.go:734`), not
  `t.Setenv("HOME", …)` — per `CLAUDE.local.md` §13, a process-wide HOME mutation pollutes parallel
  tests.
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

Introduces the registry table of §C as code, its drift guard, the missing `.moai/memory/` backup,
and the `dirs.go` Category D comment fix.

- Covers REQ-UDS-006 through REQ-UDS-010.

### M3 — On-disk backup for the three in-memory-only files

Writes `.claude/settings.json`, `.moai/status_line.sh`, and `.gitignore` to the run-scoped backup
directory before the first destructive step, on both paths, retaining the in-memory merge-back.

- Covers REQ-UDS-001 through REQ-UDS-005.

### M4 — HOME deletion-radius pinning

Extracts the removal target to a named symbol, adds the radius guard, and demonstrates its failure
against a widened radius via `go test -overlay`.

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
