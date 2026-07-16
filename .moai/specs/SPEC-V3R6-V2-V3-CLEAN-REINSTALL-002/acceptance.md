# SPEC-V3R6-V2-V3-CLEAN-REINSTALL-002 — Acceptance Criteria

**Status**: draft (plan-phase)
**Parent SPEC**: SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001 (status: completed, 2026-05-26, phase v3.0.0-rc2)
**REQ token namespace**: `REQ-CRR-NNN` (Clean-Reinstall-Repair) — aligned with `spec.md §B` and `plan.md §F` (distinct from the parent's `REQ-VVCR-NNN` to avoid collision with the parent's 35 REQ tokens).

## §D AC Matrix

Each AC is a Given-When-Then scenario tracing to one or more `REQ-CRR-NNN` tokens defined in `spec.md §B`. The AC ID namespace, severities, and traceability rows are aligned verbatim with `spec.md §D Acceptance Index`.

Severity: S1 = must-pass (blocks release), S2 = should-pass (debt if missed), S3 = nice-to-have.

### AC-CRR-001 — V3-version negative-override fires [S1]

**Traceability**: REQ-CRR-001 (restores parent REQ-VVCR-027 idempotency).

**GIVEN** a fixture project with `.moai/config/sections/system.yaml` carrying `moai.version: v3.0.0-rc2`, AND a lingering `.agency/` legacy directory present, AND any non-empty subset of the 43 DeprecatedPaths still on disk
**WHEN** `detectV2Fingerprint` is invoked
**THEN** all of the following hold:
- (a) the heuristic returns `IsV2: false`;
- (b) the returned reason string names the v3-version negative-override;
- (c) the return is issued regardless of Signal 2 (`.agency/`) and Signal 3 (DeprecatedPaths) state — both are short-circuited.

### AC-CRR-002 — Loop terminates on v3 project (reproduction test) [S1]

**Traceability**: REQ-CRR-002 (loop termination invariant), REQ-CRR-008 (reproduction test shape per CLAUDE.md §7 Rule 4).

**GIVEN** a fixture v3 project (`system.yaml moai.version: v3.0.0-rc2` + user-modified `language.yaml` with `conversation_language: ko`)
**WHEN** `moai update` is invoked twice in succession on the same fixture
**THEN** all of the following hold:
- (a) the second invocation does NOT enter the clean-reinstall code path;
- (b) no backup directory is created at `.moai/backups/v2-to-v3-*-{stamp}/` on the second invocation;
- (c) no REMOVE-phase log is emitted on the second invocation;
- (d) `language.yaml` is byte-identical (`conversation_language: ko`) across both invocations.

The reproduction test MUST fail on the pre-fix implementation (loop re-triggers on the second run) and pass on the post-fix implementation.

### AC-CRR-003 — `language.yaml` user customization survives consecutive runs [S1]

**Traceability**: REQ-CRR-003 (user-modified config preservation during file-level sync), REQ-CRR-010 (non-weakening of parent HARD-6).

**GIVEN** a v3 project with `.moai/config/sections/language.yaml` containing `conversation_language: ko` (user-modified from the template baseline `en`)
**WHEN** the file-level sync code path (REQ-CRR-002) runs on consecutive `moai update` invocations
**THEN** all of the following hold:
- (a) the SHA-256 hash diff classifies `language.yaml` as user-modified (per parent REQ-VVCR-007);
- (b) the user's `ko` value is preserved byte-identical at the canonical path `.moai/config/sections/language.yaml`;
- (c) when the sync encounters a template drift, the template baseline is diverted to `.moai/backups/v2-to-v3-{ISO}/config-conflicts/language.yaml.template-baseline`;
- (d) the user version is NOT overwritten, removed, relocated, or dropped.

### AC-CRR-004 — Positive project-marker precondition [S1]

**Traceability**: REQ-CRR-004 (restores parent REQ-VVCR-001 from false-positive regression).

**GIVEN** `moai update` invoked in any directory
**WHEN** the v2 fingerprint heuristic is evaluated
**THEN** all of the following hold:
- (a) the heuristic first checks for the presence of `.moai/config/sections/system.yaml` as a regular file;
- (b) when that file is absent, the heuristic returns `IsV2: false` immediately WITHOUT evaluating Signal 1/2/3, regardless of `.agency/` presence or any other legacy marker;
- (c) the return reason names the missing project marker.

### AC-CRR-005 — Non-project directory rejection (reproduction test) [S1]

**Traceability**: REQ-CRR-005 (non-project cwd rejection), REQ-CRR-009 (reproduction test shape per CLAUDE.md §7 Rule 4).

**GIVEN** `moai update` invoked in a fixture cwd with NO `.moai/` directory at all (e.g. `/tmp/some-random-dir`)
**WHEN** the invocation completes
**THEN** all of the following hold:
- (a) no `.moai/` or `.claude/` directory is created in the cwd;
- (b) no template files are written anywhere under the cwd;
- (c) a structured `not a moai project` error is emitted, naming the missing marker file and directing the user to `moai init`;
- (d) the command exits non-zero;
- (e) `runCleanReinstall` is never invoked.

The reproduction test MUST fail on the pre-fix implementation (full template tree installed in the cwd) and pass on the post-fix implementation.

### AC-CRR-006 — Actual-removal-count log gating (phantom eliminated) [S1]

**Traceability**: REQ-CRR-006 (subsidiary repair for #1084 phantom log).

**GIVEN** a clean-reinstall invocation where the REMOVE phase enumerates the 43-entry DeprecatedPaths list (parent §A.4)
**WHEN** the REMOVE phase completes
**THEN** all of the following hold:
- (a) the removal count is computed as `(paths existing pre-REMOVE) - (paths existing post-REMOVE)`, NOT the planned-list length (43 or any filtered subset like 10);
- (b) when the actual removed count is zero, no `Removed N deprecated paths` message is emitted;
- (c) when the actual removed count is zero, a `No deprecated paths found to remove` informational line MAY be emitted in its place;
- (d) when the actual removed count is greater than zero, the `Removed N deprecated paths` message reports the actual count.

### AC-CRR-007 — Independent `.agency/` migration on v3 project [S1]

**Traceability**: REQ-CRR-007 (restores parent REQ-VVCR-025 user-asset migration contract).

**GIVEN** a v3 project (per REQ-CRR-001 negative-override) with a lingering `.agency/` legacy directory present
**WHEN** `moai update` is invoked
**THEN** all of the following hold:
- (a) `runMigrateAgency` fires independently of the v2 fingerprint verdict;
- (b) `.agency/` content is migrated into `.moai/` per the parent REQ-VVCR-025 contract;
- (c) the full clean-reinstall code path is NOT activated;
- (d) no backup directory is created at `.moai/backups/v2-to-v3-*-{stamp}/`;
- (e) no REMOVE-phase log is emitted.

### AC-CRR-008 — Non-weakening of PRESERVE / namespace protection [S1]

**Traceability**: REQ-CRR-010 (parent §C HARD-1 through HARD-6 all preserved).

**GIVEN** the implementation of REQ-CRR-001..009 applied
**WHEN** the run-phase author compares the parent SPEC's user-asset preservation surfaces against the post-fix codebase
**THEN** ALL parent invariants are intact:
- (a) PRESERVE inventory enumeration (parent REQ-VVCR-005/006) unchanged;
- (b) SHA-256 hash-diff detection of user-modified configs (parent REQ-VVCR-007/008) unchanged;
- (c) backup directory scheme `.moai/backups/v2-to-v3-{ISO-8601-UTC}/` (parent REQ-VVCR-009/010) unchanged;
- (d) MERGE-back path restoration (parent REQ-VVCR-013..016) unchanged;
- (e) namespace-protected path set (parent REQ-VVCR-005/006) unchanged;
- (f) 43-entry DeprecatedPaths list (parent §A.4) unchanged (HARD-3 forbids pruning).

Verification obligation: `git diff internal/cli/update_preserve_inventory.go` yields 0 lines changed.

### AC-CRR-009 — Three-run idempotency on v3 project [S1]

**Traceability**: REQ-CRR-011 (restores parent REQ-VVCR-027 across multiple runs).

**GIVEN** a fixture v3 project with user-modified `language.yaml` containing `conversation_language: ko`
**WHEN** `moai update` is invoked three times in succession on the same fixture
**THEN** all of the following hold:
- (a) `language.yaml` is byte-identical (`conversation_language: ko`) across all three runs;
- (b) no backup directory is created on runs 2 and 3;
- (c) no REMOVE-phase log is emitted on runs 2 and 3;
- (d) run 1 MAY legitimately perform file-level sync if template content drifted (this is NOT a regression — it is the normal v3 sync path).

### AC-CRR-010 — Cross-platform parity (macOS/Linux/Windows) [S2]

**Traceability**: REQ-CRR-001..011 (SHOULD-1 cross-platform constraint in `spec.md §C`).

**GIVEN** the fingerprint predicate change (REQ-CRR-001 v3-version negative-override + REQ-CRR-004 positive-marker precondition) applied
**WHEN** the reproduction tests (AC-CRR-002, AC-CRR-005) and the idempotency test (AC-CRR-009) run on macOS, Linux, and Windows fixture runners
**THEN** all of the following hold:
- (a) the fingerprint verdict is identical across all three OS matrices;
- (b) the non-project-directory rejection behavior is identical across all three OS matrices;
- (c) the `os.Stat` semantics for `.moai/config/sections/system.yaml` detection produce the same positive/negative verdict on all three OS.

## §D.1 Severity Classification

| AC | Severity | Blocks release? |
|---|---|---|
| AC-CRR-001 (v3-version negative-override) | S1 | yes |
| AC-CRR-002 (loop terminates, reproduction test) | S1 | yes |
| AC-CRR-003 (language.yaml preservation) | S1 | yes |
| AC-CRR-004 (positive project-marker precondition) | S1 | yes |
| AC-CRR-005 (non-project rejection, reproduction test) | S1 | yes |
| AC-CRR-006 (actual-removal-count log gating) | S1 | yes |
| AC-CRR-007 (independent `.agency/` migration) | S1 | yes |
| AC-CRR-008 (non-weakening of PRESERVE) | S1 | yes |
| AC-CRR-009 (three-run idempotency) | S1 | yes |
| AC-CRR-010 (cross-platform parity) | S2 | no (debt if missed) |

## §D.2 Traceability Matrix

| REQ | Primary AC | Secondary AC |
|---|---|---|
| REQ-CRR-001 (v3-version negative-override) | AC-CRR-001 | AC-CRR-002, AC-CRR-010 |
| REQ-CRR-002 (loop termination invariant) | AC-CRR-002 | AC-CRR-009 |
| REQ-CRR-003 (user-modified config preservation during file-level sync) | AC-CRR-003 | — |
| REQ-CRR-004 (positive project-marker precondition) | AC-CRR-004 | AC-CRR-005, AC-CRR-010 |
| REQ-CRR-005 (non-project cwd rejection) | AC-CRR-005 | — |
| REQ-CRR-006 (actual-removal-count log gating) | AC-CRR-006 | — |
| REQ-CRR-007 (independent `.agency/` migration) | AC-CRR-007 | — |
| REQ-CRR-008 (reproduction test: fingerprint non-convergence) | AC-CRR-002 | — |
| REQ-CRR-009 (reproduction test: non-project directory pollution) | AC-CRR-005 | — |
| REQ-CRR-010 (user-asset preservation non-weakening) | AC-CRR-008 | AC-CRR-003 |
| REQ-CRR-011 (three-run idempotency) | AC-CRR-009 | — |

## §D.3 Indirect Verification

The following surfaces have no dedicated AC but are covered transitively:

- **HARD-1 (parent SPEC authority preserved)** — verified indirectly: `spec.md §A.3 Non-Goals` and `plan.md §H Cross-References` both name the parent as authoritative for design intent; AC-CRR-008 verifies the parent's preservation surfaces are untouched.
- **HARD-2 (PRESERVE inventory non-weakening)** — verified indirectly via AC-CRR-008 (a)(b)(c)(d)(e); the obligation `git diff internal/cli/update_preserve_inventory.go yields 0 lines changed` is the mechanical check.
- **HARD-3 (43-entry DeprecatedPaths list frozen)** — verified indirectly via AC-CRR-008 (f); the table is read-only input to the heuristic, and the reproduction test fixture (AC-CRR-002) uses at least one entry, so any change to the table size surfaces as fixture drift.
- **HARD-4 (Reproduction-First binding)** — verified indirectly via AC-CRR-002 and AC-CRR-005 (both bind REQ-CRR-008/009 reproduction tests that MUST fail pre-fix before any production fix is applied).
- **HARD-5 (no code changes in plan-phase)** — verified at plan-phase gate: this plan-phase artifact set carries no code changes under `internal/`, `pkg/`, `cmd/`, or template source. Run-phase is a separate delegation.
- **HARD-5 (parent SPEC backup-before-removal)** — verified indirectly via AC-CRR-002 step 2 fixture assertion that the backup directory exists on the first run (the first run is the only run where clean-reinstall legitimately fires).
- **HARD-6 (parent SPEC user-modified-config diversion to `config-conflicts/`)** — verified indirectly via AC-CRR-003 (c).

## §D.4 Closure Gates (Definition of Done)

- [ ] all S1 AC pass (AC-CRR-001, -002, -003, -004, -005, -006, -007, -008, -009).
- [ ] AC-CRR-010 (cross-platform parity, S2) either passes OR is recorded as debt with rationale.
- [ ] reproduction tests (AC-CRR-002 binds REQ-CRR-008; AC-CRR-005 binds REQ-CRR-009) are green on the run-phase branch AND were confirmed failing on the pre-fix codebase first (Reproduction-First per CLAUDE.md §7 Rule 4).
- [ ] parent SPEC HARD-1..HARD-6 invariants verified unchanged (AC-CRR-008).
- [ ] `internal/cli/update_preserve_inventory.go` has 0 lines changed (AC-CRR-008 verification obligation).
- [ ] no edits under `internal/template/templates/` (CLAUDE.local.md §25 template-internal-isolation preserved).
- [ ] no new top-level `moai` CLI command (parent constraint preserved; fix is in-place predicate + log gating, not a new command).
- [ ] no new DeprecatedPaths entries (HARD-3; the 43-entry list is frozen).
- [ ] 3-phase close: plan → run → sync per the V3R6 lifecycle (no separate Mx-phase commit).

## §D.5 Forward-Looking Checks (run-phase obligations)

- The run-phase author MUST execute the Pre-Flight Audit in `plan.md §C` (6 commands including parent SPEC existence check, affected-files existence check, DeprecatedPaths 43-entry count verification, baseline test capture, and `migrateLegacyMemoryDir` precedent grep) before any code change.
- The run-phase author MUST run the reproduction tests FIRST (Reproduction-First per CLAUDE.md §7 Rule 4) and confirm they FAIL on the pre-fix codebase before implementing the M1+M2+M3 fixes. Only after confirming failure may the fix be applied; the reproduction tests must then PASS.
- The run-phase author MUST verify fingerprint convergence on the SAME fixture used for the first invocation (no test-fixture drift between first and second invocation).
- The run-phase author MUST NOT weaken PRESERVE-inventory enumeration, SHA-256 hash-diff detection, namespace protection, backup directory scheme, MERGE-back restoration, or the 43-entry DeprecatedPaths list to make any AC pass (HARD-2 + HARD-3 + AC-CRR-008).
- The run-phase author MUST NOT modify `internal/cli/update_preserve_inventory.go` (HARD-2 / anti-pattern AP-CRR-004).
- The run-phase author MUST NOT redesign the parent's 7-step canonical clean-reinstall workflow (REQ-VVCR-004) — the fix is a predicate change + log gating + invocation-point relocation, not a workflow redesign (anti-pattern AP-CRR-005).
- The run-phase author MUST NOT treat this SPEC as an in-place amendment of the parent (anti-pattern AP-CRR-006) — the parent remains `status: completed` and authoritative; this is a `-002` sequel.
- The run-phase author MUST report self-verification per the 5-section Evidence-Bearing format (Claim / Evidence / Baseline-attribution / Gaps / Residual-risk) per `plan.md §E` and `.claude/rules/moai/core/verification-claim-integrity.md` §3.

## §D.6 Edge Cases

- **Edge-1 (v3 project, user-modified file with coincidental hash-match)**: a user-touched file whose hash coincidentally equals the template baseline (no-op hash-diff) — preservation is still required because the user may revert later; the marker records "user-touched" not "user-differing". Covered by AC-CRR-003 invariant (a) which classifies via the SHA-256 diff plus the user-touched flag, not by hash alone.
- **Edge-2 (partial v3 markers, mid-migration)**: a project cwd with partial v3 markers (some v3 paths present, some v2 paths present) on a v2 project MID-MIGRATION — fingerprint must return positive to complete the migration, not get stuck mid-state. This is a v2-project case (system.yaml carries v2.* version), so REQ-CRR-001 negative-override does NOT fire.
- **Edge-3 (`.agency/` in non-project cwd)**: `.agency/` directory present in a non-project cwd (e.g. a tool config cache) — must NOT trigger clean-reinstall. Covered by AC-CRR-004 + AC-CRR-005 guard (the positive-marker precondition refuses before any Signal evaluation).
- **Edge-4 (interrupted clean-reinstall on a v2 project)**: interrupted clean-reinstall (process killed mid-MERGE-back) on a v2 project — next invocation must detect incomplete migration and resume, NOT treat it as v3-reached. The v3-version negative-override does NOT mask this (the v2 project's system.yaml still carries v2.*).
- **Edge-5 (`system.yaml` symlink)**: v3 project where `.moai/config/sections/system.yaml` is a SYMLINK to a regular file — `os.Stat` (not `os.Lstat`) treats symlink-to-regular-file as regular, so both the symlink and the regular-file forms satisfy the positive-marker precondition (AC-CRR-004). Malicious symlink loops are rejected by `os.Stat` returning an error, which the heuristic treats as "marker absent".
- **Edge-6 (Windows path separators)**: Windows path separator differences in the positive-marker check — MUST use `filepath.Join("config", "sections", "system.yaml")` over string concatenation, so the predicate resolves identically on macOS/Linux/Windows (cross-platform parity AC-CRR-010).
- **Edge-7 (moai project where `system.yaml` exists but is empty)**: a moai project where `system.yaml` exists but is empty (zero bytes or YAML parse-fails) — the positive-marker precondition is satisfied (file exists as regular file), but Signal 1 version read returns empty. In this case the v3-version negative-override does NOT fire (no `v3.*` prefix), so the heuristic falls through to Signal 2/3 evaluation. This is correct: a moai project with a corrupted `system.yaml` is treated as a candidate for v2 detection (not silently rejected).

## §D.7 Quality Gate Criteria (TRUST 5)

- **Tested**: reproduction tests (AC-CRR-002 binds REQ-CRR-008, AC-CRR-005 binds REQ-CRR-009) green; characterization tests for fingerprint predicate (REQ-CRR-001/004), non-project rejection (REQ-CRR-005), `.agency/` decoupling (REQ-CRR-007), phantom-log gating (REQ-CRR-006), and three-run idempotency (REQ-CRR-011) all green; coverage on touched files `internal/cli/v2_detection.go`, `internal/cli/update_clean_install.go`, `internal/cli/update.go` >= 85% per the project target.
- **Readable**: clear naming for the v3-version negative-override and the positive-marker precondition (the M1 decision-heaviest milestone per `plan.md §F`); comments in English per `language.yaml code_comments: en`; error messages name the missing marker file and link the remedy (`moai init`).
- **Unified**: `gofmt` + `golangci-lint run --timeout=2m` clean on touched files; no NEW lint findings beyond baseline (distinguish NEW from pre-existing per `plan.md §B B5`).
- **Secured**: input validation on project-marker detection (AC-CRR-004 must NOT be bypassable by a crafted cwd path, symlink escape, or path traversal); no path traversal in the backup-diversion path `.moai/backups/v2-to-v3-{ISO}/config-conflicts/`; the structured `not a moai project` error does NOT leak absolute cwd paths or environment details.
- **Trackable**: Conventional Commit subject `fix(SPEC-V3R6-V2-V3-CLEAN-REINSTALL-002): repair #1084 #1086 regressions` per CLAUDE.local.md §4 Git Workflow + `manager-develop-prompt-template.md §B9` (Hybrid Trunk 1-person OSS, direct-to-main push for Tier M); `🗿 MoAI` trailer NOT required for plan-phase artifact commit (this is plan-phase, not code).

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-07-16 | manager-spec | Initial plan-phase draft. 10 AC-CRR-NNN tokens (9 S1 + 1 S2 cross-platform) aligned to `spec.md §D Acceptance Index` and `plan.md §F` milestone decomposition. Traceability matrix covers all 11 REQ-CRR-NNN requirements. Reproduction-First binding (CLAUDE.md §7 Rule 4) encoded in AC-CRR-002 (REQ-CRR-008) and AC-CRR-005 (REQ-CRR-009). Parent SPEC HARD-1..HARD-6 invariants preserved (AC-CRR-008 + §D.3 indirect verification). |
