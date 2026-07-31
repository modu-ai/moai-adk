# SPEC-UPDATE-REINSTALL-LOOP-002 — Implementation Plan

Version: 0.2.0 · Status: draft · Tier: M

> Sections are ordered by decision-reversibility. §A carries the two decisions most likely to change under review; §B..§D are the mechanical consequences; milestones (§E) follow execution order.

## §A Decisions requiring review first

### D1 — What counts as "v3-confirmed"

**Decision.** Replace the literal `strings.HasPrefix(v, "v3.")` test with a normalize-then-compare-major rule:

1. `strings.TrimSpace`
2. strip at most one leading `v` or `V`
3. parse the leading run of digits as the major version
4. classify, in this order:
   - `major >= 3` → v3-confirmed (prefix irrelevant — `v3.0.1`, `3.0.1`, `V3.0.1`, `v4.0.0`, `4.0.0` all qualify)
   - `major == 2` **and the original string carried a leading `v` or `V`** → Signal 1 positive, no override
   - `system.yaml` absent, unreadable, or unparseable **as YAML**, or `moai.version` empty → Signal 1 positive, no override
   - everything else in a well-formed file — a parseable major that is neither 2 nor ≥3, a **bare** (unprefixed) major-2 string, or a string with no leading numeric run at all — → Signal 1 negative, no override

**Why `>= 3` and not `== 3`.** A `v4.0.0` project is not a v2 project. Pinning equality reproduces the identical bug one major version later — the matrix in spec.md §A shows `v4.0.0` and `4.0.0` both classifying `IsV2=true` today. `>= 3` is the forward-safe form.

**Why major-only, ignoring prerelease and build metadata.** The signal being answered is "is this project on the v3 template era", which is a major-version question. `3.0.0-rc13` is a v3 project. Full semver comparison would add ordering semantics no caller needs.

**Why the bare `2.x` form is excluded from the Signal-1-positive branch.** This is the one asymmetry in the rule and it is deliberate. On a residue-free project, `2.5.0` and `V2.5.0` are `IsV2=false` today (observed — spec.md §A, residue-free table); admitting them on major alone flips them to `true` and routes a project with no legacy residue into the destructive full reinstall. NFR-RIL2-001 forbids exactly that movement and takes no exception. The v3 side has no such constraint, which is why REQ-RIL2-007 *does* accept bare `3.0.1`: widening v3-confirmed narrows the destructive path. The practical cost is nil — a genuine v2 project carries v2 residue by definition, so Signals 2/3 catch it whatever its version string says.

**Why "unparseable" means the file, not the major digits.** An absent, unreadable, or YAML-invalid `system.yaml` stays Signal-1 positive: that preserves the existing broader-detection posture, since a drifted config file is itself evidence of a partial migration. A *well-formed* file carrying an unrecognized string (`abc`, `3`, `1.9.0`) is a different case and stays **negative**, exactly as the current `default:` branch does. Conflating the two would make `abc` Signal-1 positive and widen a residue-free project from `IsV2=false` to `true` — the same NFR-RIL2-001 violation as the bare-`2.x` case, arriving through a wording ambiguity rather than a design choice. AC-RIL2-003 carries `abc` in its input set for this reason.

**Why a parseable non-2, non-3+ major is negative-with-no-override.** It reproduces today's `default:` branch exactly, and it leaves a version string available for tests that need Signal 1 neutral so they can isolate Signals 2/3 — which REQ-RIL2-009 depends on.

**Reversibility.** This is the decision most likely to be revised. Alternatives considered and rejected: (a) full semver parse via a dependency — adds a module for a major-digit extraction; (b) normalising at the producer (`pkg/version`, `.goreleaser.yml`) — out of scope per spec.md §C, and would leave already-installed projects carrying the bare form unfixed; (c) accepting both prefixes with a two-branch `HasPrefix` — restates the bug for `V`, `4.`, and every future major.

### D2 — How to handle the false-negative direction (Defect 3)

**Decision: narrow the override's consequence, do not defer.**

The override keeps its current effect on `IsV2` — a v3-confirmed project must not enter the destructive full reinstall (REQ-RIL2-018). What changes is that a v3-confirmed project carrying Signal-2 or Signal-3 residue is routed to a **residue-cleanup path**: back up the deprecated paths, remove them, and let the existing `.agency/` migration pre-step run. No PRESERVE snapshot, no tree wipe, no forced redeploy (REQ-RIL2-020).

**Why not defer.** Deferring is only defensible while the override is rare. It is rare today *because of Defect 1* — released users never reach it. M1 makes `V3VersionConfirmed` true for the entire released population, so the unconditional short-circuit stops being an edge case and becomes the default. Shipping M1 without D2 would convert issue #1243's "loops forever" into "residue is never cleaned", including the half-migrated freeze described in spec.md §A Defect 3. The two changes are coupled and belong in one SPEC.

**Why this is a small change, not a new subsystem.** Half of it already exists: `internal/cli/update.go:406-412` already fires `runAgencyMigrationAdapter` independently when `fpErr == nil && !fingerprint.IsV2 && isMoAIProject(cwd)`. That block is precisely the "v3 project with residue" branch. D2 extends it with the deprecated-path half, reusing `scanDeprecatedPaths` and `backupDeprecatedPaths`, both of which already exist in `internal/cli/update_cleanup.go`.

**Reversibility.** The alternative — leaving the override unconditional and documenting the freeze as accepted debt — remains available and is a one-block revert. It is rejected here because the freeze is unrecoverable by any user-facing command.

## §B Consequence: the exclusion cannot ship without the backup

Excluding deprecated paths from the PRESERVE inventory (REQ-RIL2-010) removes them from the Step 3 snapshot as well as from the Step 6 restore. Verification found that clean-reinstall Step 4 (`update_clean_install.go:265-275`) calls `os.RemoveAll` directly and never calls `backupDeprecatedPaths` — a grep of lines 250-276 for that identifier returns 0, despite the file header at `:12` documenting Step 4 as "scanDeprecatedPaths + backupDeprecatedPaths".

Today the 9 intersecting paths are incidentally protected by the resurrection bug itself. Removing the resurrection without adding the backup would turn a no-op into unbacked-up deletion of user-authored `.claude/commands/agency/*.md`. The two halves ship together in M2.

## §C Known issues in the current tree

Verified 2026-07-31. **Baseline**: code baseline `d5336214e`; worktree HEAD `5468a4afc` (branch `plan/epic-update-config-audit`), a descendant of `d5336214e` that changes SPEC documents only — `git merge-base --is-ancestor d5336214e HEAD` exits 0, and `git rev-list --count HEAD..origin/main` is `0`. Every row below was re-run against `5468a4afc`.

| Claim | Location | Verified |
|---|---|---|
| Literal `"v3."` prefix test | `internal/cli/v2_detection.go:197` | yes |
| Preserve roots = 3 entries | `internal/cli/update_preserve_inventory.go:66-70` | yes |
| `DeprecatedPaths` ∩ preserve roots = 9 | probe over `defs.DeprecatedPaths` (40 entries) | yes |
| Post-REMOVE re-scan precedes merge-back | `update_clean_install.go:278` vs Step 6 | yes |
| Dry-run returns before v2 detection | `update.go:294-304` vs `:328` | yes |
| `DryRun:` passed at `update.go:360` is always false | consequence of the above | yes |
| Step 4 never calls `backupDeprecatedPaths` | `update_clean_install.go:250-276`, grep count 0 | yes |
| GoReleaser injects bare version | `.goreleaser.yml:22` | yes |
| `GetFullVersion` applies no normalization | `pkg/version/version.go:29-30` | yes |
| Banner always renders a `v` | `internal/cli/uikit/banner.go:100` | yes |
| Agency-migration pre-step block sits at `:406-412` (not `:384-390`) | `grep -n 'fpErr == nil && !fingerprint.IsV2 && isMoAIProject' internal/cli/update.go` → `406:` | yes |
| Deny-rule migration (`stripRetiredV2DenyEntries`) sits between the dry-run return and the v2 block | `update.go:306-326`, call at `:323`; dry-run return `:293-304`; v2 block `{`…`}` at `:334-413` | yes |
| Residue-free `2.5.0` / `V2.5.0` classify `IsV2=false` today | probe over `probeVersionSignal` + `detectV2Fingerprint` | yes |
| Residue-free `abc` classifies Signal 1 **negative** today | same probe | yes |
| Residue-carrying fixture yields `IsV2=true` for every non-v3-confirmed row | same probe, `residue=true` arm | yes |
| `moai spec lint` reports 0 findings for this SPEC | `moai spec lint` → `0 error(s), 62 warning(s)`, none naming `SPEC-UPDATE-REINSTALL-LOOP-002` | yes |

No claimed file:line failed to match. The single misattribution found by plan-audit (M3's `:384-390`) is corrected in §E M3.

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
- Add the nine-row table-driven test (REQ-RIL2-008) asserting `Signal 1`, `IsV2`, and `V3VersionConfirmed` per row. The `Signal 1` column is load-bearing: on the residue-carrying fixture the `IsV2` column is `true` for every non-v3-confirmed row, so `IsV2` alone cannot detect a Signal-1 regression.
- Migrate the `TestUpdate_ZeroNetChange_DesignDirNoLongerTriggersLoop` fixture (REQ-RIL2-009): its current `3.0.0-rc13` string becomes v3-confirmed under D1, so the test would pass through the override instead of isolating Signal 3. Replace with a version whose major is neither 2 nor ≥3.
- Add the monotonicity assertion for NFR-RIL2-001, on a **residue-free** fixture and over an input set that includes `2.5.0`, `V2.5.0`, and `abc`. Built on the residue-carrying fixture the assertion is vacuous; without those three inputs it cannot observe the two ways the rule can widen (bare major-2, and "unparseable" read as "major digits unparseable").

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

Touches `internal/cli/update.go` (the `fpErr == nil && !fingerprint.IsV2 && isMoAIProject(cwd)` block at `:406-412`).

> Locate the block by its condition token `fpErr == nil && !fingerprint.IsV2 && isMoAIProject`, not by line number — the line number is a convenience, the token is the anchor. Confirmed at `:406` on the current tree (`grep -n 'fpErr == nil && !fingerprint.IsV2 && isMoAIProject' internal/cli/update.go` → `406:`).

- Extend that block with a deprecated-path sweep: `scanDeprecatedPaths` → `backupDeprecatedPaths` → remove (REQ-RIL2-019).
- Do not add snapshot, wipe, or redeploy (REQ-RIL2-020); leave the agency migration pre-step in place (REQ-RIL2-021).
- Add the two-consecutive-run idempotence test (REQ-RIL2-022).

Priority: High. Depends on M1 (the override population widens) and on M2 (reuses the backup wiring).

### M4 — `--dry-run` reachability

Touches `internal/cli/update.go` (dry-run branch ordering).

**`--dry-run` must never mutate the filesystem. The reachability fix works by making the already-implemented non-mutating renderer reachable — never by relocating the early return.**

- **Hoist the detection above the dry-run branch.** Compute the v2 fingerprint (and the residue-cleanup decision) *before* the `--dry-run` early return at `:293-304`, then have that branch render the resulting plan and return. `runCleanReinstall` already contains the non-mutating renderer — `if opts.DryRun { … return result, nil }` at `update_clean_install.go:186-198`, which prints the plan (`scanDeprecatedPaths` at `:189`) and performs no mutation — and `update.go:360` already wires `DryRun: getBoolFlag(cmd, "dry-run")` into it. The renderer is unreachable today only because the CLI returns first. Nothing new needs to be written to satisfy REQ-RIL2-024 / REQ-RIL2-025; the plumbing has to reach the code that exists.
- **Rejected option — moving the dry-run early return to after the v2-detection block.** This was listed as an equal alternative in v0.1.0 and is now dropped as a confirmed defect. `update.go:306-326` sits between the dry-run early return and the v2-detection block, and calls `stripRetiredV2DenyEntries(cwd, out)`, which rewrites `settings.json`. Relocating the return past that block would make `moai update --dry-run` write to disk, violating REQ-RIL2-026 and AC-RIL2-014. The code's own comment at `:312-313` states the placement is deliberate — *"after the --binary / --dry-run early-returns (so a dry run never mutates)"* — so the rejected option reverses an invariant the source explicitly asserts. Any implementation that moves that return is wrong regardless of whether a test happens to catch it.
- Assert zero mutation via a tree-hash comparison (REQ-RIL2-026).
- Keep the legacy-skill archive summary and worktree advisory (REQ-RIL2-027). Both are emitted from the dry-run branch today — `emitWorktreeAdvisory(out, cwd)` at `update.go:302` and `dryRunArchiveLegacySkills(cwd, out)` at `:303` — so hoisting detection above the branch must not displace either call.

> Consistency note: the sibling `SPEC-UPDATE-DOC-DRIFT-001` settles its own `--dry-run` handling the same way (make the existing non-mutating renderer reachable — `spec.md:373` selects option B; `:376-378` keeps the early return in place). Neither SPEC may propose a `--dry-run` path that writes.
>
> **Sibling co-edit constraint (mirrored from `SPEC-UPDATE-DOC-DRIFT-001/progress.md` §E.1, "M1 versus E1").** Both SPECs touch the `--dry-run` branch of `internal/cli/update.go`. **The two MUST NOT edit that branch concurrently on this branch** — they are not a dependency edge in either frontmatter, so the ordering is honoured by sequencing, not by a `depends_on` gate or an `--ignore-deps` override. This SPEC (E1) owns the reachability change as REQ-RIL2-024 / REQ-RIL2-025; the sibling owns only the help-text contract. Either order is correct: if this SPEC runs first, M4 implements the hoist and the sibling's M1 verifies an already-landed fix; **if the sibling lands first, M4 degrades to a no-op verification** — confirm the hoist is present, confirm `--dry-run` still mutates nothing (AC-RIL2-014), and record that no code change was required, rather than authoring a competing hoist. The constraint is recorded here rather than only in the sibling because it changes what M4 does, not merely when it runs.

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
