---
id: SPEC-V3R6-V2-V3-CLEAN-REINSTALL-002
title: "Repair v2→v3 clean-reinstall regression (#1084 silent data loss loop, #1086 arbitrary-directory pollution) — Implementation Plan"
version: "0.1.0"
status: draft
created: 2026-07-16
updated: 2026-07-16
author: manager-spec
priority: P1
phase: "v3.0.0-rc-stabilization"
module: "internal/cli"
lifecycle: spec-anchored
tags: "moai-update, v2-v3-migration, regression-repair, fingerprint-convergence, idempotency"
tier: M
depends_on: [SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001]
related_specs: [SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001, SPEC-V3R3-UPDATE-CLEANUP-001, SPEC-V3R6-UPDATE-NAMESPACE-PROTECT-001]
---

# Implementation Plan — SPEC-V3R6-V2-V3-CLEAN-REINSTALL-002

## §A Context

### §A.1 Regression Root-Cause Analysis (3 causes)

The regression introduced by `SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001`'s implementation has three distinct root causes that share the `runUpdate → detectV2Fingerprint → runCleanReinstall` code path:

**Root Cause A — Fingerprint non-convergence (parent REQ-VVCR-001 / REQ-VVCR-027 violation).**

The v2-fingerprint heuristic combines three OR'd signals (Signal 1 version, Signal 2 `.agency/`, Signal 3 DeprecatedPaths). The parent's REQ-VVCR-027 idempotency contract requires that on a v3 project, the heuristic returns `IsV2: false` so `runUpdate` routes to file-level sync. The implementation violates this because:

- (A.i) Signal 1's `Option α broader detection policy` treats `moai.version` empty/missing as a POSITIVE v2 signal. This is correct for v2 projects that predate `system.yaml`, but it fails to converge on v3 projects where `system.yaml` legitimately exists with `moai.version: v3.*`. The Option α policy lacks a v3-version negative-override.
- (A.ii) Signal 3 enumerates 43 DeprecatedPaths (parent §A.4 expansion). After REMOVE removes all 43 and reinstall does NOT recreate them, the next run's Signal 3 SHOULD find 0 paths. BUT — the 43 entries include v.2.x-era paths (Category B) and rc1-stage paths (Category C) that overlap with paths the v3 template still ships. If the reinstall phase recreates any of these (a subtle violation of parent REQ-VVCR-019), Signal 3 stays positive → loop.

The fix strategy is **NOT** to prune the DeprecatedPaths list (HARD-3 forbids that), but to add a **v3-version negative-override** (REQ-CRR-001): when Signal 1 reads `v3.*`, the heuristic returns `IsV2: false` regardless of Signal 2/3 state. A v3 version field is proof of prior migration; any lingering Signal 2/3 evidence is user-retained legacy, not v2-zombie evidence.

**Root Cause B — Phantom `Removed N deprecated paths` log.**

The REMOVE phase logs `Removed N deprecated paths` where N is the **planned-list length** (43 entries, or a filtered subset like 10), not the **actual filesystem removal count**. On the first clean-reinstall, N matches reality. On the second run (post-convergence), REMOVE finds 0 paths to remove (all 43 are already absent) but the log still emits N=10 (or 43). Issue #1084 confirms: actual `git diff` is 0 paths removed, but the log says 10.

The fix is **actual-removal-count log gating** (REQ-CRR-006): compute the removal count as `(paths existing pre-REMOVE) - (paths existing post-REMOVE)`, and emit the log only when this count > 0.

**Root Cause C — `signals: version=true` false-positive in non-project directories (#1086).**

The Option α broader-detection policy treats `system.yaml missing` as a POSITIVE Signal 1. This is correct for v2 projects that predate `system.yaml` (Signal 1 = positive when the file is absent), but it ALSO fires in arbitrary non-project directories. Running `moai update` in `/tmp/some-random-dir` (which has no `.moai/` at all) triggers Signal 1 = positive → `IsV2: true` → `runCleanReinstall` installs the full template tree → cwd polluted with `.moai/`, `.claude/`, and the entire template namespace.

The parent's acceptance criterion EC-5 Scenario A explicitly documents this `system.yaml missing → positive Signal 1` branch as INTENDED behavior for v2 handling. The regression is that the branch lacks a **positive project-marker precondition**: `no system.yaml AND no positive moai-project marker → not a moai project, refuse`. The fix adds the precondition (REQ-CRR-004/005).

### §A.2 Shared Code Path → Unified Fix Strategy

The three root causes share `runUpdate → detectV2Fingerprint → runCleanReinstall`. The fix is unified:

1. **Fingerprint predicate tightening** (kills Root Cause A): add v3-version negative-override at the top of `detectV2Fingerprint`.
2. **Positive project-marker precondition** (kills Root Cause C): gate the entire heuristic behind a `.moai/config/sections/system.yaml` existence check; absence → `IsV2: false` + structured error.
3. **Actual-removal-count log gating** (kills Root Cause B): recompute N from filesystem diff, not from the planned list.
4. **`.agency/` migration decoupling** (preserves parent REQ-VVCR-025 user-asset contract): move `runMigrateAgency` invocation out of `runCleanReinstall` and into a pre-update step parallel to `migrateLegacyMemoryDir` at `internal/cli/update.go:1731`.

### §A.3 Affected Files (run-phase scope; NOT modified in plan-phase)

| File | Parent SPEC origin | Change scope |
|------|-------------------|--------------|
| `internal/cli/v2_detection.go` | Parent M3 (NEW) | Modify `detectV2Fingerprint` to add v3-version negative-override (REQ-CRR-001) and positive project-marker precondition (REQ-CRR-004). |
| `internal/cli/update_clean_install.go` | Parent M4 (NEW) | Modify REMOVE phase to compute actual removal count and gate the log (REQ-CRR-006). Move `runMigrateAgency` invocation out (REQ-CRR-007). |
| `internal/cli/update.go` | Parent M5 integration | Add non-project-directory rejection in `runUpdate` before fingerprint call (REQ-CRR-005). Add `runMigrateAgency` as independent pre-update step (parallel to `migrateLegacyMemoryDir` at line 1731). |
| `internal/cli/v2_detection_test.go` | Parent M3 tests | Add reproduction tests REQ-CRR-008/009 (failing-pre-fix, passing-post-fix). |
| `internal/cli/update_clean_install_test.go` | Parent M4 tests | Add idempotency tests REQ-CRR-011 (three-run invariant). |

### §A.4 PRESERVE Inventory (DO NOT TOUCH — HARD-2)

The following surfaces are OWNED by the parent SPEC and MUST NOT be altered by this SPEC's implementation:

- `internal/cli/update_preserve_inventory.go` — `buildPreserveInventory()`, `detectUserModifiedConfigs()` (SHA-256 hash diff), `snapshotPreserveInventory()`
- PRESERVE enumeration semantics (parent REQ-VVCR-005/006)
- SHA-256 hash-diff detection (parent REQ-VVCR-007/008)
- Backup directory scheme `.moai/backups/v2-to-v3-{ISO-8601-UTC}/` (parent REQ-VVCR-009/010)
- MERGE-back path restoration (parent REQ-VVCR-013..016)
- DeprecatedPaths 43-entry list (parent §A.4; HARD-3 forbids pruning)

## §B Known Issues / Risks (B1-B12 auto-injection, filtered to relevant)

**B1 (Cross-platform Build Tags)** — minimal risk. The fix uses `os.Stat` for the project-marker check (cross-platform safe). No `syscall` package usage introduced. Pre-flight: `GOOS=windows GOARCH=amd64 go build ./...` MUST still pass.

**B2 (Cross-SPEC Policy Conflict Pre-Scan)** — RELEVANT. The parent SPEC's design is the policy; this SPEC is a regression repair. Run `grep -r "Retired\|superseded" internal/cli/v2_detection.go internal/cli/update_clean_install.go` pre-flight; no retire/supersede markers expected in these files (they were NEW in the parent).

**B3 (C-HRA-008 Subagent Boundary)** — N/A. The fix is in CLI code, not subagent-domain code. No `AskUserQuestion` calls introduced.

**B5 (CI 3-tier Awareness)** — RELEVANT. spec-lint, golangci-lint, and per-OS tests can each fail. The reproduction tests (REQ-CRR-008/009) MUST fail pre-fix on ALL three OS matrices (macOS, Linux, Windows).

**B6 (spec-lint Heading Convention)** — RELEVANT for this plan-phase artifact set. The spec.md uses `### Out of Scope — <topic>` H3 sub-headings (not bare `## Out of Scope` H2).

**B7 (`input.CWD` resolution)** — RELEVANT. The non-project-directory rejection (REQ-CRR-004) depends on resolving the cwd correctly. Prefer `$CLAUDE_PROJECT_DIR` over `os.Getwd()` to avoid the leak hazard documented in B7.

**B10 (Untouched Paths PRESERVE)** — RELEVANT. The fix's PRESERVE list is §A.4 above. Notably, `internal/cli/update_preserve_inventory.go` is OFF-LIMITS (parent owns it).

**B11 (AskUserQuestion Prohibited)** — N/A (subagent boundary respected; this is plan-phase).

## §C Pre-Flight Audit (run-phase execution, NOT plan-phase)

The following commands MUST be executed by `manager-develop` at run-phase M1 start, before any code change:

```bash
# 1. Confirm parent SPEC files exist (the regression baseline)
ls -la .moai/specs/SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001/{spec,plan,acceptance,progress}.md

# 2. Confirm the affected files exist (parent M3/M4/M5 delivered them)
ls -la internal/cli/v2_detection.go internal/cli/update_clean_install.go
git grep -n "runCleanReinstall\|detectV2Fingerprint" internal/cli/update.go

# 3. Confirm the existing test baseline (so reproduction tests can extend, not replace)
ls -la internal/cli/v2_detection_test.go internal/cli/update_clean_install_test.go

# 4. Confirm the DeprecatedPaths 43-entry count (parent §A.4) is intact
git grep -c "deprecatedPaths\|DeprecatedPaths" internal/cli/update_cleanup.go internal/defs/

# 5. Capture pre-fix reproduction baseline (for B5 per-OS test matrices)
go test ./internal/cli/... 2>&1 | tail -20
GOOS=linux go test ./internal/cli/... 2>&1 | tail -5
GOOS=windows go test ./internal/cli/... 2>&1 | tail -5

# 6. Confirm migrateLegacyMemoryDir precedent exists (for runMigrateAgency decoupling)
git grep -n "migrateLegacyMemoryDir" internal/cli/update.go
```

## §D Constraints

(See spec.md §C — duplicated here for run-phase agent visibility.)

- HARD-1: Parent SPEC authority preserved (do NOT supersede design intent)
- HARD-2: PRESERVE inventory non-weakening (do NOT touch `update_preserve_inventory.go`)
- HARD-3: No new DeprecatedPaths entries (43-entry list FROZEN)
- HARD-4: Reproduction-First binding (tests authored run-phase M4, fail-pre-fix first)
- HARD-5: No code changes in plan-phase (this SPEC is plan-phase ONLY; run-phase is separate delegation)
- SHOULD-1: Cross-platform parity (macOS / Linux / Windows identical predicate behavior)

## §E Self-Verification Checklist (run-phase manager-develop E1-E7)

> Reported per the 5-section Evidence-Bearing format (Claim / Evidence / Baseline-attribution / Gaps / Residual-risk) — see `.claude/rules/moai/core/verification-claim-integrity.md` §3.

- **E1** — AC binary PASS/FAIL matrix (10 AC-CRR-001..010, each row: AC ID, status, verification command, actual output). MUST: 9/10 PASS (all MUST); SHOULD-1 (AC-CRR-010 cross-platform) target PASS but non-blocking.
- **E2** — Cross-platform build: `go build ./...` AND `GOOS=windows GOARCH=amd64 go build ./...` both exit 0.
- **E3** — Coverage on `internal/cli/v2_detection.go` and `internal/cli/update_clean_install.go` ≥85% (per project target).
- **E4** — Subagent boundary grep (C-HRA-008 family): `grep -rn 'AskUserQuestion' internal/cli/ | grep -v "_test.go" | grep -v "// "` yields 0 matches.
- **E5** — Lint: `golangci-lint run --timeout=2m` exits 0; report NEW issues separately from baseline.
- **E6** — Branch HEAD + push state: list of new commit SHAs; `git push origin main` result (Hybrid Trunk 1-person OSS, direct-to-main per CLAUDE.local.md §23).
- **E7** — Blocker report (if any): structured blocker, NEVER `AskUserQuestion`.

## §F Milestone Decomposition

> Ordered by **decision-reversibility** (highest-change-likelihood decisions first, mechanical steps last) per CLAUDE.md §7 Rule 1 Approach-First.

### M1 — Fingerprint predicate tightening (decision-heaviest)

**Why first**: The v3-version negative-override (REQ-CRR-001) and positive project-marker precondition (REQ-CRR-004) are the highest-impact behavioral changes. If these decisions are wrong, downstream milestones are wasted. Human review focuses here.

**Scope**:
- Modify `internal/cli/v2_detection.go` `detectV2Fingerprint`:
  - Add positive project-marker precondition at function entry: if `.moai/config/sections/system.yaml` does not exist as a regular file → return `IsV2: false, Reason: "no moai-project marker"` immediately.
  - Add v3-version negative-override after reading `moai.version`: if version starts with `v3.` → return `IsV2: false, Reason: "v3 project (negative-override)"` regardless of Signal 2/3 state.
- Author unit tests:
  - `TestDetectV2Fingerprint_V3VersionNegativeOverride` — covers REQ-CRR-001
  - `TestDetectV2Fingerprint_PositiveProjectMarkerPrecondition` — covers REQ-CRR-004
  - `TestDetectV2Fingerprint_LegacyV2HandlingUnchanged` — non-regression: legacy `v2.*` version, `.agency/`, DeprecatedPaths still detect correctly (parent REQ-VVCR-001 contract preserved)

**Reversibility**: HIGH. Predicate logic change; easy to revert if it breaks legacy v2 detection.

### M2 — Non-project directory rejection (#1086 surface)

**Why second**: Builds on M1's positive-marker precondition; adds the structured-error path in `runUpdate`.

**Scope**:
- Modify `internal/cli/update.go` `runUpdate`:
  - Before invoking `detectV2Fingerprint`, check for `.moai/config/sections/system.yaml` existence.
  - If absent → emit structured `not a moai project` error (naming the missing marker file + directing user to `moai init`), exit non-zero WITHOUT invoking `runCleanReinstall` or installing any template content.
- Author unit tests:
  - `TestRunUpdate_NonProjectDirectoryRejection` — covers REQ-CRR-005

**Reversibility**: HIGH. Adds a new early-exit path; existing legitimate-project behavior unchanged.

### M3 — `.agency/` migration decoupling (user-asset preservation defense)

**Why third**: Decouples `runMigrateAgency` from `runCleanReinstall` so v3 projects with lingering `.agency/` directory still get migration without triggering the loop.

**Scope**:
- Move `runMigrateAgency` invocation OUT of `runCleanReinstall` (`internal/cli/update_clean_install.go`) and INTO `runUpdate` (`internal/cli/update.go`), as a pre-update step parallel to `migrateLegacyMemoryDir` at line 1731.
- The decoupled invocation fires whenever `.agency/` is present, regardless of fingerprint verdict.
- Author integration test:
  - `TestRunUpdate_V3ProjectWithAgencyDir_MigratesIndependently` — covers REQ-CRR-007

**Reversibility**: MEDIUM. Invocation-point relocation; the function body of `runMigrateAgency` itself is unchanged (parent owns it).

### M4 — Phantom-log gating + reproduction tests (Reproduction-First per HARD-4)

**Why fourth**: Gates the REMOVE-phase log on actual removal count, and authors the failing-reproduction tests that HARD-4 binds.

**Scope**:
- Modify `internal/cli/update_clean_install.go` REMOVE phase:
  - Pre-scan: count paths in the planned list that exist on filesystem → `N_planned`.
  - Post-scan: after REMOVE, count paths in the planned list that still exist → `N_remaining`.
  - Actual removal count: `N_removed = N_planned - N_remaining`.
  - Emit `Removed N deprecated paths` ONLY when `N_removed > 0`. When `N_removed == 0`, emit `No deprecated paths found to remove` informational line instead.
- Author failing-reproduction tests (CLAUDE.md §7 Rule 4):
  - `TestReproduction_FingerprintNonConvergence_Issue1084` — covers REQ-CRR-008. The test MUST fail on the pre-fix implementation (clean-reinstall re-triggers on second run) and pass on the post-fix implementation.
  - `TestReproduction_NonProjectDirectoryPollution_Issue1086` — covers REQ-CRR-009. Same fail-pre/pass-post invariant.
- Run reproduction tests BEFORE applying M1+M2+M3 fixes; confirm they FAIL. Then re-run after fixes; confirm they PASS.

**Reversibility**: LOW. Log-message change is mechanical; reproduction tests are additive.

### M5 — Idempotency integration tests + cross-platform verification (mechanical)

**Why last**: Pure test-enumeration + CI verification; no decision content.

**Scope**:
- Author three-run idempotency integration test `TestRunUpdate_ThreeRunIdempotency_V3Project` — covers REQ-CRR-011. Fixture: v3 project + user-modified `language.yaml` (`conversation_language: ko`). Invoke `moai update` three times; assert: (a) `language.yaml` byte-identical across all three runs, (b) no backup directory created on runs 2/3, (c) no REMOVE-phase log on runs 2/3.
- Author cross-platform parity test (SHOULD-1) — runs the fixture on macOS/Linux/Windows; asserts identical fingerprint verdict + rejection behavior.
- Verify pre-existing parent tests still pass (non-regression on parent REQ-VVCR-001..027 contract).
- Run full `go test ./internal/cli/... ./internal/defs/...` and `golangci-lint run`; report exit codes.

**Reversibility**: NONE. Pure verification step.

## §G Anti-Patterns

- **AP-CRR-001** — Pruning the DeprecatedPaths 43-entry list to `fix the loop`. This violates HARD-3. The fix is the v3-version negative-override (REQ-CRR-001), not list pruning.
- **AP-CRR-002** — Weakening the PRESERVE inventory to `stop the overwrite`. This violates HARD-2. The fix tightens the fingerprint predicate so clean-reinstall does not activate on v3 projects; the PRESERVE inventory remains intact for legitimate v2 projects.
- **AP-CRR-003** — Authoring reproduction tests POST-fix only (skipping the failing-pre-fix step). This violates CLAUDE.md §7 Rule 4 (Reproduction-First). The tests MUST be authored and confirmed FAILING on the pre-fix implementation BEFORE the fix is applied.
- **AP-CRR-004** — Modifying `internal/cli/update_preserve_inventory.go`. This is HARD-2 territory. The fix's scope is `v2_detection.go`, `update_clean_install.go`, `update.go`, and their `_test.go` siblings — NOT the PRESERVE inventory module.
- **AP-CRR-005** — Redesigning the clean-reinstall 7-step canonical order (parent REQ-VVCR-004). Out of scope. The fix modifies predicate logic and log gating, not workflow structure.
- **AP-CRR-006** — Treating this SPEC as an in-place amendment of the parent. The parent remains `status: completed` and authoritative; this is a `-002` sequel that repairs implementation, not design.
- **AP-CRR-007** — Changing `system.yaml` detection to look at a different file (e.g., `.moai/manifest.json`). The fix uses `.moai/config/sections/system.yaml` (the SAME file the parent's Signal 1 reads) as the positive marker — anything else introduces signal-source drift.

## §H Cross-References

- **Parent SPEC**: `.moai/specs/SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001/spec.md` (design intent — DO NOT modify)
- **Parent plan.md M3-M5**: defines `internal/cli/v2_detection.go`, `internal/cli/update_clean_install.go`, `internal/cli/update_preserve_inventory.go`, `internal/cli/update.go` integration points
- **Parent acceptance.md AC-VVCR-001 + AC-VVCR-004 + AC-VVCR-015 + EC-5**: the regression surface (AC-VVCR-001 = fingerprint heuristic; AC-VVCR-004 = user-modified config preservation; AC-VVCR-015 = v3 idempotency; EC-5 = `system.yaml missing → positive Signal 1` Scenario A that #1086 exploits)
- **`.moai/project/tech.md`**: Go 1.26, standard `testing` (no testify), system Git via exec.Command (no go-git), custom YAML loader
- **CLAUDE.local.md §23**: Hybrid Trunk 1-person OSS, Tier M direct-to-main push
- **`.claude/rules/moai/development/manager-develop-prompt-template.md`**: Section A-E 5-section delegation template (Tier M form)
- **Epic scope**: This = ENTRY SPEC; SPEC-2 = #1087/#1088 (doctor false signals); SPEC-3 = #1090 (zone-registry template packaging); #1081 OUT (Linux dotfile, upstream track)

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-07-16 | manager-spec | Initial plan-phase draft. 5 milestones ordered decision-reversibility-first (M1 predicate → M2 non-project rejection → M3 `.agency` decoupling → M4 phantom-log + reproduction tests → M5 idempotency + cross-platform). PRESERVE inventory non-weakened (HARD-2). Reproduction-First bound (HARD-4). |
