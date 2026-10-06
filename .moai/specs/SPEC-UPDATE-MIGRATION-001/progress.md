# progress.md — SPEC-UPDATE-MIGRATION-001 (card t1547)

## §E.1 Plan-phase Audit-Ready Signal

plan_status: audit-ready
plan_complete_at: 2026-10-07
tier: L
artifacts: spec.md, plan.md, acceptance.md, design.md, research.md (5) + progress.md

## §Assumptions

Recorded per the card's autonomous-lane instruction (contract-free run; assumptions recorded, not asked). Each is falsifiable at run-phase pre-flight.

1. **A1 — Manifest presence is not guaranteed.** Many existing projects lack a healthy
   `.moai/manifest` (feature is newer than most installs). Design assumes manifest-absent
   projects classify conservatively as `user-modified` with prior-render merge base (spec R-1).
   If run-phase finds manifest coverage is actually universal, the fallback path is dead code —
   keep it anyway (cheap, tested).
2. **A2 — `internal/manifest` enum names.** research.md cites `template_managed` /
   `user_modified` / `user_created` provenance values and `manifest.HashBytes` from the
   migrate_classify source; exact symbol names to be re-verified at M3 (read-only check, not a
   redesign risk).
3. **A3 — `.moai-new` sidecar naming is new surface.** No precedent found in-tree; the name
   `<path>.moai-new` is chosen for greppability and adjacency. If t1527 later defines a
   conflict-reporting vocabulary, the sidecar name is trivially renamable (single constant).
4. **A4 — Template content needs no change.** The reconciliation is orchestration-side; no
   `internal/template/templates/` edit is planned. If M4/M5 find a template gap (e.g. a
   missing `.gitignore`-style marker convention for YAML), it is escalated as a blocker, not
   improvised (Template-First + §25 neutrality would apply).
5. **A5 — `CleanMoaiManagedPathsWithTargets` callers.** The legacy/migration callers of the
   WithTargets form are assumed compatible with the REQ-UPM-015 guard (they pass classified
   removal lists already). Re-verify call sites at M4; a caller that legitimately needs to
   remove user-owned paths would be a design contradiction requiring escalation.
6. **A6 — v1→v2 fresh-install path scope.** REQ-UPM-040 assumes the legacy incompatible-config
   path is reachable only from version-class detection (v1 projects). If no such path is
   reachable in the current tree, the requirement reduces to "no wholesale path remains except
   tests" — a simpler close, recorded at M4.
7. **A7 — Fixture git repos in tests.** AC-UPM-040 runs `git status` inside a `t.TempDir()`
   fixture repo. Assumes `git init` in tests is acceptable in this repo's CI (existing suites
   do not appear to depend on git binaries in unit tests — verify; if not, AC-UPM-040 falls
   back to filesystem-only assertions and the git-status check runs as a scripted verification
   step instead).

## Mode Selection

Deferred to run-phase Phase 4 (owned by the orchestrator/manager-develop flow, not this plan phase). Expected: `serial` (coding-heavy, brownfield, single-package cluster).

## §E.2 Run-phase Evidence

Duty mirror (plan-audit r2 claim-discrepancy note): the M2-before-M4 RED recording
duty and the auto-demotion clause for AC-UPM-020/021 live in acceptance.md:31 and
acceptance.md:37 — both hazard RED captures land HERE with all four §2.1 elements
before M4, or both criteria auto-demote to regression-guard (binding, not discretionary).

### Milestone log

- **M1 — characterization pin (GREEN-BASE, no behavior change)** — commit _pending_.
  `internal/cli/update/deploy/deploy_characterization_test.go` (new): 4 tests pinning
  current `CleanMoaiManagedPaths` behavior — local-only file backed-up-then-deleted;
  user-modified template file deleted WITHOUT backup (silent-overwrite hazard);
  `.moai/config` wiped wholesale; managed roots fully emptied (wipe-first end state).
  Exit: green on the UNMODIFIED tree (HEAD 80cfe7c0b).

## §E.3 Run-phase Audit-Ready Signal

_pending run-phase_

## §E.4 Sync-phase Audit-Ready Signal

_pending sync-phase_
