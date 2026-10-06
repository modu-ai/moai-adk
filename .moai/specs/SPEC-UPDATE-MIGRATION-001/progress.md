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

- **M1 — characterization pin (GREEN-BASE, no behavior change)** — commit `e9ed0ef0f`.
  `internal/cli/update/deploy/deploy_characterization_test.go` (new): 4 tests pinning
  current `CleanMoaiManagedPaths` behavior — local-only file backed-up-then-deleted;
  user-modified template file deleted WITHOUT backup (silent-overwrite hazard);
  `.moai/config` wiped wholesale; managed roots fully emptied (wipe-first end state).
  Exit: green on the UNMODIFIED tree (HEAD 80cfe7c0b).
- **M2 — RED hazard tests** — commit `c244fcdaf` (tests + M4-shaped reconciliation
  seam whose bodies delegate to the current wipe-first flow). 6 tests RED:
  `TestUpdate_LocalOnlyFileSurvives`, `TestUpdate_GitStrategyValuesSurvive`,
  `TestUpdate_NoGitDeletionsForLocalOnlyFiles`, `TestUpdate_ConflictPreservesFileAndWritesSidecar`,
  `TestUpdate_ConflictSidecarCollisionUsesFirstUnusedNumber`, `TestUpdate_StaleFileArchivedAndRemoved`.
  Fixtures are in-code (`t.TempDir()` per the package's `newClassifyFixture` convention)
  rather than `testdata/` files — the plan's testdata note is satisfied in spirit; no
  golden files are involved.

### RED evidence — AC-UPM-020 / AC-UPM-021 (§2.1 four elements, MP8-RED-M2)

Captured BEFORE M4 (this evidence commit precedes any pipeline-change commit).
Both measurements ran on the M2 commit; the code paths exercised are committed
there (`c244fcdaf`), not an uncommitted working copy.

**LEDGER-RED-UPM-020** (cited by acceptance.md:31):

- command: `go test -run TestUpdate_LocalOnlyFileSurvives -count=1 ./internal/cli/update/`
- stdout (verbatim):

```
--- FAIL: TestUpdate_LocalOnlyFileSurvives (0.01s)
    reconcile_test.go:107: local-only file did not survive update: open /var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/TestUpdate_LocalOnlyFileSurvives3579795044/001/.claude/rules/moai/dev-only-rule.md: no such file or directory
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli/update	0.116s
FAIL
```

- executed-test count: 1 (`--- FAIL` names the executed test; this is NOT `ok ... [no tests to run]`)
- exit code: `1`
- tree: `c244fcdaf` (this run, this tree)

**LEDGER-RED-UPM-021** (cited by acceptance.md:37):

- command: `go test -run TestUpdate_GitStrategyValuesSurvive -count=1 ./internal/cli/update/`
- stdout (verbatim):

```
--- FAIL: TestUpdate_GitStrategyValuesSurvive (0.00s)
    reconcile_test.go:174: worktree_base_branch = "", want "develop" (VALUE reversion — the 2026-09-24 hazard)
    reconcile_test.go:177: manual.workflow = "github-flow", want "git-flow" (VALUE reversion — the 2026-09-24 hazard)
    reconcile_test.go:180: summary neither preserved nor merged .moai/config/sections/git-strategy.yaml: {Refreshed:[] Merged:[] Conflicts:[] Preserved:[] ArchivedRemoved:[]}
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli/update	0.116s
FAIL
```

- executed-test count: 1 (`--- FAIL` names the executed test; this is NOT `ok ... [no tests to run]`)
- exit code: `1`
- tree: `c244fcdaf` (this run, this tree)

**Debt disposition:** MP8-RED-M2 is DISPOSED — both release-blocking criteria keep
their RED-now cells with all four §2.1 elements; the auto-demotion clause did NOT
execute. The RED-for-the-right-reason check: both failures name the exact hazard
the criterion asserts (file destroyed / operator VALUES reverted), not an unrelated
pre-existing breakage.

### Card-scope review findings (leader relay, 4 P1 + 2 P2) — reproduction and repair

Each finding was reproduced in this session before repair (VCI §1: a codex claim is
a hypothesis until measured). All six verified against this tree; regression tests
added for each.

1. **P1 classifier misroute (reconcile_classify.go)** — REPRODUCED: a pristine
   prior-render file (healthy manifest record, hash matches disk) whose NEW render
   differs classified `user-modified`; with no merge base the conflict disposition
   then reverted the template's own update. REPAIR: the template-owned decision now
   reads the TRACKED state (REQ-UPM-001's own wording) — healthy record + hash match
   → template-owned regardless of the new render. Test: `TestClassifyHealthyRecordPristineContentRefreshes`.
   The failing `TestUpdateForce_TemplateChangedKeyStillPropagates` family is green again.
2. **P1 double processing (update_template_sync.go)** — REPRODUCED (same failing
   family): restore-merged section files and the mergeable set were re-captured into
   `reconPending`; the merge phase then diffed the operator's pre-deploy bytes
   against the OTHER step's output and could revert its delivered updates. REPAIR:
   `ReconcileOptions.Exclude` — the cli wiring excludes `.moai/config/sections/*`
   (restore-handled) and the `collectMergeableFiles` set (mergeable-handled). Tests:
   `TestUpdate_ExcludedPathsNotPending` + the green YAML propagation family.
3. **P1 symlink regression (reconcile.go)** — REPRODUCED: a symlinked managed root
   recorded-but-left let the deploy write THROUGH the link to an external directory
   (the wholesale clean this replaces used to remove link entries). REPAIR:
   `deploy.DisposeSymlinks` applies the existing link-dedicated dispositions to the
   classifier's recorded links BEFORE the deploy; the reconcile calls it as a
   destructive step inside the guard window. Test: `TestUpdate_SymlinkedRootDisposedBeforeDeploy`
   (external sentinel byte-identical, link removed, deploy writes a real directory).
4. **P1 archive escape (reconcile.go)** — REPRODUCED shape: an archive destination
   under a symlink would route the copy outside the project and then remove the
   operator's file. REPAIR: `ensureNoSymlinkPath` Lstats every path component from
   the project root to the destination and refuses before any write; the stale file
   stays in place. Test: `TestUpdate_ArchiveRefusesSymlinkDestination`.
5. **P2 archive overwrite across runs** — REPRODUCED by construction: the archive
   root was run-invariant, so a re-update overwrote the previous recovery copy.
   REPAIR: run-scoped archive root `<tag>/<timestamp>/<rel>`. Test:
   `TestUpdate_ArchiveRunScopedNoOverwrite` (run-1 copy survives run-2 byte-identical).
6. **P2 outcome accounting (update_template_sync.go)** — REPRODUCED: the outcome
   note still counted the whole pre-clean snapshot as removals on the default path,
   reporting preserved files as deleted. REPAIR: the default path's removal
   accounting reads the reconciliation's actual dispositions
   (`len(ArchivedRemoved)`; RemovedLocalOnly stays 0 — every removal carries a
   recovery copy); the wholesale branches keep the t40 accounting.

Environment note: `TestGuardBypassMutant_ObserveHomePollution` reads the lane
factory env; from this lane session it false-reds on `MOAI_FACTORY_WORKER=lane-11`
(the t1350 class). With the env scrubbed the test passes — the red is session
environment, not code.

### Milestone log (continued)

- **M4 — reconciliation pipeline + wiring + guard** — commit `59fc3be1e` (with the
  six review findings repaired in the same milestone, per the leader relay; the
  per-finding reproduction record is above). The six M2 hazard tests GREEN:
  `go test -count=1 ./internal/cli/update/...` all packages ok (this run, tree
  59fc3be1e). Broad cli update-flow suite (env-scrubbed, ~30 test-name families):
  `ok github.com/modu-ai/moai-adk/internal/cli 118.005s`.
- **M5 — summary + counting** — `plan.AnalyzeFiles` managed exclusion removed
  (REQ-UPM-032; the wipe premise is gone), `report.RenderReconciliation` added on
  the existing plain-text outcome structure, cli outcome gains the reconciliation
  rows (counts + conflict/preserved/archived path lists; pill no longer adds
  `ManagedRedeployed` — the caller's count already includes managed files, and the
  breakdown states the inclusion), `--dry-run` previews the reconciliation plan via
  the read-only classifier (`previewReconciliation`; the cleanup-deletion preview's
  subject no longer exists). Tests: `TestUpdate_SummaryHonesty` (AC-UPM-032, all
  five categories per-path), preview + outcome test families updated to the new
  contracts.

## §E.3 Run-phase Audit-Ready Signal

_pending run-phase_

## §E.4 Sync-phase Audit-Ready Signal

_pending sync-phase_
