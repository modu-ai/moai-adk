---
id: SPEC-UPDATE-REINSTALL-LOOP-002
title: "moai update — end the release-user clean-reinstall loop (semver-normalized v2/v3 detection, deprecated-path exclusion from PRESERVE, residue cleanup, dry-run reachability)"
version: 0.3.0
status: draft
created: 2026-07-31
updated: 2026-07-31
author: manager-spec
priority: high
phase: plan
module: cli
lifecycle: spec-anchored
era: V3R6
tier: M
tags: "cli, update, clean-reinstall, v2-detection, semver, deprecated-paths, preserve, dry-run, regression"
issue_number: 1243
related_specs: [SPEC-UPDATE-REINSTALL-LOOP-001, SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001, SPEC-V3R6-V2-V3-CLEAN-REINSTALL-002, SPEC-UPDATE-YAML-PRESERVE-001]
---

# SPEC-UPDATE-REINSTALL-LOOP-002

## HISTORY

| Version | Date | Change |
|---------|------|--------|
| 0.1.0 | 2026-07-31 | Initial draft. Four-lens audit of `moai update` / `.moai/config`; Epic SPEC 1 of 6. |
| 0.2.0 | 2026-07-31 | Plan-audit revision (D1-D5). REQ-RIL2-003 narrowed to prefixed forms so it no longer conflicts with NFR-RIL2-001; §A Defect 1 gains the residue-free widening table and the `2.5.0` / `V2.5.0` rows; §G risk row 2 rewritten; REQ-RIL2-021 and NFR-RIL2-004 gain binding AC coverage. |
| 0.3.0 | 2026-07-31 | Plan-audit iter3 revision (PASS 0.88), documentation-only — no requirement or code change. **D12 (major)** closed: `AC-RIL2-014` command (a) was satisfiable by a comment alone; it now strips leading-comment lines and the AC additionally requires `TestUpdateDryRun_ZeroMutation` to assert at runtime that the fixture's `permissions.deny` array carries a `retiredV2DenyEntries` literal, with the residual trailing-comment gap stated explicitly. **D9 (minor)** disposition split: `acceptance.md:37` corrected six → four (its referent is the nine-row AC-RIL2-001 table); `progress.md:37` retains "six" as an accurate record of the v0.1.0 seven-row matrix and was only disambiguated — the audit's prescription to change it is rejected and the rejection recorded. **D15 (minor)**: the `SPEC-UPDATE-DOC-DRIFT-001` co-edit constraint (no concurrent edits to the `--dry-run` branch; M4 degrades to no-op verification if the sibling lands first) mirrored into `plan.md` §E M4. |

## §A Problem / Motivation

`SPEC-UPDATE-REINSTALL-LOOP-001` closed ONE cause of the perpetual `moai update` clean-reinstall loop: a `defs.DeprecatedPaths` entry that the v3 template also shipped, so removal was undone by redeployment. Its guard `TestDeprecatedPaths_NoTemplateCollision` (`internal/cli/deprecated_paths_collision_test.go:53`) keeps that intersection empty.

The loop nevertheless persists for **released** users, driven by two further causes that the existing guard structurally cannot see, plus a mirror-image false-negative and a diagnostic blind spot.

### Defect 1 — version detection is a literal `"v3."` prefix test

`probeVersionSignal` (`internal/cli/v2_detection.go:197`) confirms v3 only when `moai.version` literally begins with `v3.`:

```go
case strings.HasPrefix(v, "v3."):
```

Observed matrix (probe run against a tree carrying one `.agency/` directory, varying only the version string; `dep=true` is the pre-existing residue, not a new signal). `Signal 1` is `probeVersionSignal`'s first return value; `IsV2` is the composed fingerprint verdict:

| `moai.version` | `Signal 1` | `IsV2` | `V3VersionConfirmed` |
|---|---|---|---|
| `v3.0.1` | false | false | true |
| `3.0.1` | false | **true** | false |
| `V3.0.1` | false | **true** | false |
| `v4.0.0` | false | **true** | false |
| `4.0.0` | false | **true** | false |
| `""` | true | true | false |
| `v2.5.0` | true | true | false |
| `2.5.0` | false | true | false |
| `V2.5.0` | false | true | false |

#### The residue-free tree is where the widening hazard is visible

On the residue-carrying fixture above, `IsV2` is `true` for every non-v3-confirmed row, because Signals 2/3 dominate. That column therefore **cannot** discriminate a Signal-1 change, and a monotonicity guard built on that fixture is vacuous. The same probe run against an otherwise-empty project (no `.agency/`, no deprecated paths — only `system.yaml`) isolates Signal 1:

| `moai.version` | `Signal 1` | `IsV2` (today) | `IsV2` if `major == 2` alone made Signal 1 positive |
|---|---|---|---|
| `v2.5.0` | true | true | true (unchanged) |
| `2.5.0` | **false** | **false** | **true — widened** |
| `V2.5.0` | **false** | **false** | **true — widened** |
| `abc` | false | false | **true — widened**, if "unparseable" is read as "major digits unparseable" |
| `1.9.0` | false | false | false |
| `3` | false | false | false |

The last two columns are why REQ-RIL2-003 is scoped to `v`/`V`-prefixed forms (see §D.1). A bare `2.5.0` on a residue-free tree is `IsV2=false` today; making the bare form Signal-1 positive would flip it to `true` and send a project with no legacy residue into the destructive full reinstall — precisely what NFR-RIL2-001 forbids.

Release binaries record the **no-`v`** form. `.goreleaser.yml:22` injects `-X …/pkg/version.Version={{.Version}}`; GoReleaser's `.Version` is the git tag with the leading `v` stripped. Issue #1243's reporter pasted `moai-adk: 3.0.1 (b7a98fb88, built …)` — the raw `GetFullVersion()` output (`pkg/version/version.go:29-30`), which applies no normalization. Local `make build` derives the version from `git describe`, which **keeps** the `v`, so maintainers never reproduce it.

The `moai version` banner is not a usable signal either: `internal/cli/uikit/banner.go:100` renders `"v" + strings.TrimPrefix(version, "v")`, so a `v` is always displayed regardless of what is stored.

Consequence: released users take the `default:` branch — no override — so Signals 2 and 3 decide alone, and a single stale `.agency/` directory or one deprecated path forces a full clean-reinstall on **every** update, forever. A future `v4.0.0` triggers this for every user in either prefix form.

### Defect 2 — deprecated paths live inside PRESERVE roots and are resurrected

`preserveInventoryRoots` (`internal/cli/update_preserve_inventory.go:66-70`) is `.moai/specs`, `.moai/project`, `.claude/commands`. Its intersection with `defs.DeprecatedPaths` (40 entries) is **9**:

```
.claude/commands/agency/{agency,brief,build,evolve,learn,profile,resume,review}.md
.moai/project/brand
```

The clean path adds them to the PRESERVE inventory (Step 2), backs them up (Step 3), deletes them (Step 4), and **restores them** (Step 6 `mergeBackPreserveInventory`). Observed on a synthetic tree:

```
PRESERVE inventory CONTAINS victim: .claude/commands/agency/brief.md
scanDeprecatedPaths(before) = [.claude/commands/agency/brief.md .agency]
```

The log still prints `[clean-reinstall] Removed N deprecated paths` because the post-REMOVE re-scan runs at `update_clean_install.go:278` — immediately after Step 4 and **before** Step 6 undoes it. Net effect is zero, the signal stays positive, the next update loops.

### Defect 3 — the false-negative direction

The override short-circuits Signals 2 and 3 unconditionally (`v2_detection.go:140-142`), so a tree with full v2 residue but a v3 version string is never cleaned. A run interrupted after templates rendered `system.yaml` to the new version freezes the project half-migrated, permanently.

This defect **worsens** once Defect 1 is fixed: normalization makes `V3VersionConfirmed` true for the entire released population, so the unconditional short-circuit stops being a rare edge and becomes the default. Fixing Defect 1 alone converts "loops forever" into "never cleans residue".

### Defect 4 — `--dry-run` cannot preview any of this

`internal/cli/update.go:294-304` returns from the `--dry-run` branch **before** the v2-detection block at `:328-413`. The dry-run early return inside `runCleanReinstall` (`update_clean_install.go:186-198`) is therefore unreachable from the CLI, and the `DryRun:` field passed at `update.go:360` always arrives false. A user running `moai update --dry-run` on a v2 tree sees only the legacy-skill archive summary.

### Defect 5 — clean-reinstall Step 4 removes without its own backup (discovered during verification)

`update_clean_install.go:12` documents Step 4 as "scanDeprecatedPaths + backupDeprecatedPaths", but the Step 4 body (`:265-275`) calls `os.RemoveAll` directly and **never** calls `backupDeprecatedPaths`. A grep of lines 250-276 for that identifier returns 0.

Today the 9 intersecting paths are incidentally protected by the PRESERVE snapshot — the very resurrection bug of Defect 2. Removing that resurrection (Defect 2's fix) therefore **opens an unbacked-up deletion path** for user-authored `.claude/commands/agency/*.md`. Defect 5 must be closed in the same change set, not after it.

## §B Goals

- A released binary reporting `3.0.1` classifies identically to a source build reporting `v3.0.1`.
- A deprecated path is never both removed and restored in one clean-reinstall cycle.
- A v3-confirmed project carrying genuine legacy residue has that residue cleaned, without a destructive full reinstall.
- Every deprecated-path deletion is preceded by a backup.
- `moai update --dry-run` previews the clean-reinstall plan without mutating the tree.

## §C Scope Exclusions

### Out of Scope — YAML merge and comment preservation

- Comment / key-order / scalar-style preservation in `.moai/config/sections/*.yaml` merges. Owned by `SPEC-UPDATE-YAML-PRESERVE-001`.
- The `map[string]any` round-trip in `internal/cli/update/backup/merge.go`.

### Out of Scope — DeprecatedPaths catalogue membership

- Adding, removing, or re-categorising entries in `defs.DeprecatedPaths` (`internal/defs/dirs.go`). This SPEC changes how the existing 40 entries are *processed*, never which paths are listed.
- The `DeprecatedPaths` ↔ embedded-template intersection invariant, owned by `SPEC-UPDATE-REINSTALL-LOOP-001`.

### Out of Scope — template content

- Any edit under `internal/template/templates/**`. Fixtures live in `_test.go` files only.
- The 16-language template neutrality contract is a constraint on this work, not a deliverable of it.

### Out of Scope — version string production

- Changing what `.goreleaser.yml` injects, what `pkg/version` stores, or what the banner renders. This SPEC makes the *consumer* tolerant; it does not normalise at the producer.

### Out of Scope — update subsystem redesign

- The 7-step canonical clean-reinstall order, the backup-coverage gate, the update lock, and the archive/legacy-skill pipeline retain their current structure.

## §D Requirements (GEARS)

### D.1 Version-signal normalization

- **REQ-RIL2-001** — The version probe shall normalize `moai.version` before comparison by trimming surrounding whitespace, removing at most one leading `v` or `V`, and parsing the leading numeric component as the major version.
- **REQ-RIL2-002** — **Where** the normalized major version is 3 or greater, the version probe shall report the project as v3-confirmed.
- **REQ-RIL2-003** — **Where** the normalized major version is exactly 2 **and** the original string carried a leading `v` or `V`, the version probe shall report Signal 1 positive and shall not report v3-confirmed.
  - **Prefix scope is deliberate, not an oversight.** A bare `2.x` (no prefix) is Signal-1 **negative** today and stays negative — see REQ-RIL2-006. Making the bare form positive would flip a residue-free project from `IsV2=false` to `IsV2=true`, which NFR-RIL2-001 forbids (§A, residue-free table). The asymmetry against REQ-RIL2-007 (which does accept the bare `3.0.1`) is intentional: widening *v3-confirmed* narrows the destructive path, whereas widening *Signal 1* broadens it. Only the second direction is constrained.
  - **Cost of the narrowing is nil in practice.** A genuine v2 project is v2 precisely because it carries v2 residue, so Signals 2/3 detect it regardless of the version string. The narrowing removes only the case "bare `2.x` version with no residue whatsoever", which is not a v2 project in any observable sense.
  - **Rejected alternative:** granting NFR-RIL2-001 an explicit exception for `major == 2` and letting the bare form be Signal-1 positive. Rejected because it converts a monotonicity invariant into a case-by-case judgement, and because the input it admits (residue-free bare `2.x`) is the exact shape of the false positive that triggers a destructive reinstall.
- **REQ-RIL2-004** — **When** `system.yaml` is absent, unreadable, or unparseable as YAML, or carries an empty `moai.version`, the version probe shall report Signal 1 positive and shall not report v3-confirmed, preserving the existing broader-detection behaviour. "Unparseable" here means the *file* fails to parse; a well-formed file carrying an unrecognized version string is governed by REQ-RIL2-006, not by this requirement.
- **REQ-RIL2-005** — **Where** the version string carries a prerelease or build suffix (`3.0.1-rc13`, `3.0.0+build.5`), the version probe shall classify it by its major component alone.
- **REQ-RIL2-006** — **When** a well-formed `system.yaml` carries a non-empty `moai.version` that REQ-RIL2-002 and REQ-RIL2-003 do not claim — a parseable major that is neither 2 nor 3-or-greater, a bare (unprefixed) major-2 string, or a string whose leading component is not numeric at all — the version probe shall report Signal 1 negative and shall not report v3-confirmed, so that Signals 2 and 3 decide. This reproduces the current `default:` branch of `probeVersionSignal` (`internal/cli/v2_detection.go:205-207`, inside the `switch` at `:192-208`) exactly, and is the clause that keeps NFR-RIL2-001 satisfiable without exception.
- **REQ-RIL2-007** — The version probe shall treat `v3.0.1`, `3.0.1`, `V3.0.1`, `v4.0.0`, and `4.0.0` as equivalent v3-confirmed inputs.
- **REQ-RIL2-008** — The regression suite shall exercise the classification as a table-driven test over at least the nine version strings enumerated in the §A Defect 1 residue-carrying matrix, asserting `Signal 1`, `IsV2`, and `V3VersionConfirmed` per row.
- **REQ-RIL2-009** — The fixture of `TestUpdate_ZeroNetChange_DesignDirNoLongerTriggersLoop` shall be migrated to a version string that still yields a negative Signal 1 without a v3 override, so the test continues to isolate Signal 3 rather than passing through the widened override.

### D.2 Deprecated-path exclusion from PRESERVE

- **REQ-RIL2-010** — The PRESERVE inventory builder shall exclude every file whose project-relative path equals, or is nested under, a `defs.DeprecatedPaths` entry.
- **REQ-RIL2-011** — The clean-reinstall shall not restore any deprecated path during the merge-back step.
- **REQ-RIL2-012** — A regression guard named for the preserve-inventory intersection shall assert that the intersection of `defs.DeprecatedPaths` and the **built** PRESERVE inventory is empty.
- **REQ-RIL2-013** — The guard shall be accompanied by a negative-path test that reintroduces an un-excluded deprecated path into the inventory and asserts the guard detects it, proving the guard is not a vacuous pass.
- **REQ-RIL2-014** — The guard shall not assert that the intersection of `defs.DeprecatedPaths` and the static `preserveInventoryRoots` prefixes is empty; that intersection is 9 entries by design and is the condition the exclusion exists to handle.

### D.3 Backup before deletion

- **REQ-RIL2-015** — **When** the clean-reinstall reaches the deprecated-path removal step, it shall back up every path it is about to delete before deleting it.
- **REQ-RIL2-016** — **When** the backup of a deprecated path fails, the clean-reinstall shall abort before deleting that path and shall surface a grep-able sentinel naming the path.
- **REQ-RIL2-017** — The removal-count reporting shall continue to be derived from the filesystem diff rather than from the planned-list length.

### D.4 Residue cleanup on v3-confirmed projects

- **REQ-RIL2-018** — **Where** a project is v3-confirmed, the v2 fingerprint shall continue to resolve `IsV2` false, so the destructive full reinstall is not triggered.
- **REQ-RIL2-019** — **When** a v3-confirmed project carries deprecated-path residue, `moai update` shall remove that residue through a residue-cleanup path that backs up before deleting.
- **REQ-RIL2-020** — The residue-cleanup path shall not perform the PRESERVE snapshot, the tree wipe, or the forced template redeployment of the full clean-reinstall.
- **REQ-RIL2-021** — **While** the residue-cleanup path is active, the existing independent `.agency/` migration pre-step shall continue to fire under its current conditions.
- **REQ-RIL2-022** — **When** `moai update` runs twice consecutively on a v3-confirmed project with residue, the second run shall remove zero deprecated paths.
- **REQ-RIL2-023** — The residue-cleanup path shall run only on a genuine MoAI project as determined by the existing project marker.

### D.5 Dry-run reachability

- **REQ-RIL2-024** — **When** `moai update --dry-run` runs on a project whose fingerprint would trigger the clean-reinstall, the command shall emit the clean-reinstall plan including the count of paths that would be removed.
- **REQ-RIL2-025** — **When** `moai update --dry-run` runs on a project whose fingerprint would trigger residue cleanup, the command shall emit the residue-cleanup plan.
- **REQ-RIL2-026** — `moai update --dry-run` shall not mutate the project tree.
- **REQ-RIL2-027** — `moai update --dry-run` shall continue to emit the legacy-skill archive summary and the worktree advisory it emits today.

## §E Non-Functional Constraints

- **NFR-RIL2-001** — Classification changes shall not widen the destructive path: no input that classifies as non-v2 today shall classify as v2 after the change. **This constraint admits no exception**; where it collided with REQ-RIL2-003, the requirement was narrowed rather than the constraint relaxed (§D.1). The comparison shall be evaluated on a **residue-free** project fixture, because a fixture carrying `.agency/` or deprecated paths yields `IsV2=true` under both the old and the new rule for every non-v3-confirmed input, making the comparison vacuous (§A, residue-free table).
  - **One-directional by design.** The reverse movement is permitted and is the point of REQ-RIL2-002 / REQ-RIL2-007: an input that classifies as v2 today MAY classify as non-v2 after the change.
- **NFR-RIL2-002** — All new tests shall use `t.TempDir()` and shall not write into the project root.
- **NFR-RIL2-003** — No file under `internal/template/templates/**` shall be added or modified.
- **NFR-RIL2-004** — Go conventions: `snake_case.go` filenames, `fmt.Errorf("…: %w", err)` wrapping, English comments and godoc.
- **NFR-RIL2-005** — Path handling shall remain correct on linux, darwin, and windows.

## §F Success Criteria

- The nine-row version matrix classifies as specified, verified by an executed table-driven test, and the residue-free monotonicity comparison reports no widened input.
- A clean-reinstall on a tree containing `.claude/commands/agency/brief.md` leaves that path absent after the cycle completes, with a backup copy on disk.
- Two consecutive `moai update` runs on a residue-carrying v3 project report zero removals on the second.
- `moai update --dry-run` prints a removal-count line and leaves the tree hash unchanged.
- `go build ./...` and `go test ./internal/cli/...` pass.

## §G Risk Register

| Risk | Direction | Bound |
|---|---|---|
| Normalization misclassifies a genuine v2 project as v3 | False negative — needed migration skipped | REQ-RIL2-003 pins prefixed major 2; the table-driven test asserts the `v2.5.0` row. A bare `2.x` deliberately falls to REQ-RIL2-006 (Signal 1 negative), where Signals 2/3 still detect the residue every genuine v2 project carries |
| Normalization misclassifies a v3 project as v2 | False positive — destructive reinstall | NFR-RIL2-001 monotonicity assertion, evaluated on a **residue-free** fixture (AC-RIL2-003). The residue-carrying matrix rows do NOT bound this risk — every non-v3-confirmed row is `IsV2=true` there under both rules, so the assertion would be vacuous. The input set explicitly includes `2.5.0`, `V2.5.0`, and `abc`, the three shapes that a careless `major == 2` or "unparseable ⇒ positive" reading would widen |
| Exclusion opens unbacked-up deletion of user commands | Data loss | REQ-RIL2-015/016 make backup a precondition of deletion in the same change set |
| Residue cleanup becomes a second destructive path | Data loss | REQ-RIL2-020 forbids snapshot/wipe/redeploy; REQ-RIL2-019 mandates backup |
| Widened override silently voids an existing test's intent | Undetected regression | REQ-RIL2-009 migrates the fixture explicitly |
| Dry-run reachability changes live-run ordering | Behaviour drift | REQ-RIL2-027 pins existing dry-run output; live-path ordering untouched |

## §H Cross-References

- `internal/cli/v2_detection.go` — `probeVersionSignal`, `detectV2Fingerprint`
- `internal/cli/update_preserve_inventory.go` — `preserveInventoryRoots`, `buildPreserveInventory`, `mergeBackPreserveInventory`
- `internal/cli/update_clean_install.go` — `runCleanReinstall` Steps 2/3/3.5/3.9/4/6
- `internal/cli/update_cleanup.go` — `scanDeprecatedPaths`, `backupDeprecatedPaths`
- `internal/cli/update.go` — dry-run branch, v2-detection block, agency-migration pre-step
- `internal/cli/deprecated_paths_collision_test.go` — the symmetric template-collision guard
- `internal/defs/dirs.go` — `DeprecatedPaths` (40 entries)
