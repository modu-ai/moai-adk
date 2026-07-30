# SPEC-UPDATE-REINSTALL-LOOP-002 — Implementation Plan

Version: 0.1.0 · Status: draft · Tier: M

> Sections are ordered by decision-reversibility. §A carries the two decisions most likely to change under review; §B..§D are the mechanical consequences; milestones (§E) follow execution order.

## §A Decisions requiring review first

### D1 — What counts as "v3-confirmed"

**Decision.** Replace the literal `strings.HasPrefix(v, "v3.")` test with a normalize-then-compare-major rule:

1. `strings.TrimSpace`
2. strip at most one leading `v` or `V`
3. parse the leading run of digits as the major version
4. classify: `major >= 3` → v3-confirmed · `major == 2` → Signal 1 positive · any other parseable major → negative, no override · unparseable / empty / absent → Signal 1 positive, no override

**Why `>= 3` and not `== 3`.** A `v4.0.0` project is not a v2 project. Pinning equality reproduces the identical bug one major version later — the matrix in spec.md §A shows `v4.0.0` and `4.0.0` both classifying `IsV2=true` today. `>= 3` is the forward-safe form.

**Why major-only, ignoring prerelease and build metadata.** The signal being answered is "is this project on the v3 template era", which is a major-version question. `3.0.0-rc13` is a v3 project. Full semver comparison would add ordering semantics no caller needs.

**Why an unparseable version stays Signal-1 positive.** This preserves the existing broader-detection posture (an absent or drifted `system.yaml` is evidence of a partial migration). Flipping it to negative would be a silent behaviour widening in the false-negative direction, which NFR-RIL2-001 forbids.

**Why a parseable non-2, non-3+ major is negative-with-no-override.** It reproduces today's `default:` branch exactly, and it leaves a version string available for tests that need Signal 1 neutral so they can isolate Signals 2/3 — which REQ-RIL2-009 depends on.

**Reversibility.** This is the decision most likely to be revised. Alternatives considered and rejected: (a) full semver parse via a dependency — adds a module for a major-digit extraction; (b) normalising at the producer (`pkg/version`, `.goreleaser.yml`) — out of scope per spec.md §C, and would leave already-installed projects carrying the bare form unfixed; (c) accepting both prefixes with a two-branch `HasPrefix` — restates the bug for `V`, `4.`, and every future major.

### D2 — How to handle the false-negative direction (Defect 3)

**Decision: narrow the override's consequence, do not defer.**

The override keeps its current effect on `IsV2` — a v3-confirmed project must not enter the destructive full reinstall (REQ-RIL2-018). What changes is that a v3-confirmed project carrying Signal-2 or Signal-3 residue is routed to a **residue-cleanup path**: back up the deprecated paths, remove them, and let the existing `.agency/` migration pre-step run. No PRESERVE snapshot, no tree wipe, no forced redeploy (REQ-RIL2-020).

**Why not defer.** Deferring is only defensible while the override is rare. It is rare today *because of Defect 1* — released users never reach it. M1 makes `V3VersionConfirmed` true for the entire released population, so the unconditional short-circuit stops being an edge case and becomes the default. Shipping M1 without D2 would convert issue #1243's "loops forever" into "residue is never cleaned", including the half-migrated freeze described in spec.md §A Defect 3. The two changes are coupled and belong in one SPEC.

**Why this is a small change, not a new subsystem.** Half of it already exists: `internal/cli/update.go:384-390` already fires `runAgencyMigrationAdapter` independently when `fpErr == nil && !fingerprint.IsV2 && isMoAIProject(cwd)`. That block is precisely the "v3 project with residue" branch. D2 extends it with the deprecated-path half, reusing `scanDeprecatedPaths` and `backupDeprecatedPaths`, both of which already exist in `internal/cli/update_cleanup.go`.

**Reversibility.** The alternative — leaving the override unconditional and documenting the freeze as accepted debt — remains available and is a one-block revert. It is rejected here because the freeze is unrecoverable by any user-facing command.

## §B Consequence: the exclusion cannot ship without the backup

Excluding deprecated paths from the PRESERVE inventory (REQ-RIL2-010) removes them from the Step 3 snapshot as well as from the Step 6 restore. Verification found that clean-reinstall Step 4 (`update_clean_install.go:265-275`) calls `os.RemoveAll` directly and never calls `backupDeprecatedPaths` — a grep of lines 250-276 for that identifier returns 0, despite the file header at `:12` documenting Step 4 as "scanDeprecatedPaths + backupDeprecatedPaths".

Today the 9 intersecting paths are incidentally protected by the resurrection bug itself. Removing the resurrection without adding the backup would turn a no-op into unbacked-up deletion of user-authored `.claude/commands/agency/*.md`. The two halves ship together in M2.

## §C Known issues in the current tree (verified 2026-07-31)

| Claim | Location | Verified |
|---|---|---|
| Literal `"v3."` prefix test | `internal/cli/v2_detection.go:197` | yes |
| Preserve roots = 3 entries | `internal/cli/update_preserve_inventory.go:66-70` | yes |
| `DeprecatedPaths` ∩ preserve roots = 9 | probe over `defs.DeprecatedPaths` (40 entries) | yes |
| Post-REMOVE re-scan precedes merge-back | `update_clean_install.go:278` vs Step 6 | yes |
| Dry-run returns before v2 detection | `update.go:294-304` vs `:306` | yes |
| `DryRun:` passed at `update.go:338` is always false | consequence of the above | yes |
| Step 4 never calls `backupDeprecatedPaths` | `update_clean_install.go:250-276`, grep count 0 | yes |
| GoReleaser injects bare version | `.goreleaser.yml:22` | yes |
| `GetFullVersion` applies no normalization | `pkg/version/version.go:29-30` | yes |
| Banner always renders a `v` | `internal/cli/uikit/banner.go:100` | yes |

No claimed file:line failed to match.

## §D Constraints

- No edits under `internal/template/templates/**` (NFR-RIL2-003). All fixtures in `_test.go`.
- `t.TempDir()` only; never the project root (NFR-RIL2-002).
- `snake_case.go`, `%w` wrapping, English comments (NFR-RIL2-004).
- Cross-platform path handling; verify with `GOOS=windows GOARCH=amd64 go build ./...` (NFR-RIL2-005).
- **`git stash` is prohibited in this work.** The checkout is shared with concurrent sessions and `git stash` is repository-global; it silently absorbs other sessions' uncommitted changes, and `git stash push` without `-u` refuses untracked files outright. Falsification steps use a scratch `git worktree` instead (see acceptance.md §C).

## §E Milestones (execution order)

### M1 — Version-signal normalization

Touches `internal/cli/v2_detection.go` (`probeVersionSignal`), plus a new table-driven test file.

- Implement the D1 rule. Keep the returned diagnostic strings informative — they surface in telemetry and `--dry-run`.
- Add the seven-row table-driven test (REQ-RIL2-008) asserting both `IsV2` and `V3VersionConfirmed` per row.
- Migrate the `TestUpdate_ZeroNetChange_DesignDirNoLongerTriggersLoop` fixture (REQ-RIL2-009): its current `3.0.0-rc13` string becomes v3-confirmed under D1, so the test would pass through the override instead of isolating Signal 3. Replace with a version whose major is neither 2 nor ≥3.
- Add the monotonicity assertion for NFR-RIL2-001.

Priority: High. Blocks M3.

### M2 — Deprecated-path exclusion + backup-before-delete

Touches `internal/cli/update_preserve_inventory.go` (`buildPreserveInventory`), `internal/cli/update_clean_install.go` (Step 4), plus a new guard test file.

- Add the exclusion predicate to the preserve walk (REQ-RIL2-010). It must match both exact paths and nested children, on slash-normalised paths.
- Wire `backupDeprecatedPaths` into Step 4 before the `os.RemoveAll` loop (REQ-RIL2-015), aborting with a grep-able sentinel on backup failure (REQ-RIL2-016).
- Add `TestDeprecatedPaths_NoPreserveInventoryCollision` — the guard symmetric to `TestDeprecatedPaths_NoTemplateCollision`, asserting the intersection with the **built inventory** is empty (REQ-RIL2-012). It must not assert emptiness against the static root prefixes, which is 9 by design (REQ-RIL2-014).
- Add the poisoned-input negative-path test (REQ-RIL2-013), modelled on the existing `TestDeprecatedPaths_CollisionGuardDetectsReinsertion`.
- Preserve the filesystem-diff removal count (REQ-RIL2-017).

Priority: High. The two halves must land in one commit — see §B.

### M3 — Residue cleanup for v3-confirmed projects

Touches `internal/cli/update.go` (the `fpErr == nil && !fingerprint.IsV2 && isMoAIProject(cwd)` block at `:384-390`).

- Extend that block with a deprecated-path sweep: `scanDeprecatedPaths` → `backupDeprecatedPaths` → remove (REQ-RIL2-019).
- Do not add snapshot, wipe, or redeploy (REQ-RIL2-020); leave the agency migration pre-step in place (REQ-RIL2-021).
- Add the two-consecutive-run idempotence test (REQ-RIL2-022).

Priority: High. Depends on M1 (the override population widens) and on M2 (reuses the backup wiring).

### M4 — `--dry-run` reachability

Touches `internal/cli/update.go` (dry-run branch ordering).

- Move the dry-run early return to after the v2-detection block, or hoist detection above it, so `runCleanReinstall`'s dry-run branch and the residue-cleanup plan are both reachable (REQ-RIL2-024, REQ-RIL2-025).
- Assert zero mutation via a tree-hash comparison (REQ-RIL2-026).
- Keep the legacy-skill archive summary and worktree advisory (REQ-RIL2-027).

Priority: Medium. Depends on M2 and M3 for the plan lines it must print.

## §F Anti-patterns to avoid

- Asserting a guard with `go test -run <pattern>` alone: a zero-match pattern exits 0. Every AC pairs `-run` with a `--- PASS: <exact name>` assertion.
- Re-implementing the deprecated-path exclusion predicate inside the guard test. The guard must call the production `buildPreserveInventory`, otherwise it proves only that the test's own copy of the rule is self-consistent.
- Using `git stash` for falsification (see §D).
- Treating the 9-entry root intersection as the defect. The defect is the *inventory* intersection; the root intersection is the legitimate input.

## §G Cross-references

- spec.md §A (defect evidence), §D (requirements), §G (risk register)
- acceptance.md §C (falsification procedure)
- `SPEC-UPDATE-REINSTALL-LOOP-001` — the sibling cause and the symmetric template guard
- `SPEC-UPDATE-YAML-PRESERVE-001` — Epic sibling; disjoint scope (YAML merge)
