# SPEC-UPDATE-MIGRATION-FIX-001 — Implementation Plan

Card t1578. Tier M. Order inside Section F follows decision-reversibility:
the probe's behavior contract (most likely to change under review) leads
the implementation milestone; the mechanical guard pins close the plan.

## A. Context

- Worktree: `.moai/worktrees/t1578`, branch `WT-update-migration-fixes`.
  BASELINE RE-CUT: the worktree fast-forwarded from 81786284e to
  2aab5f797 (main absorbed, 155 commits) mid-research; the canonical
  baseline for all anchors is 2aab5f797 and every research conclusion was
  re-verified on it (acceptance.md EV-6). Relevant re-cut deltas: the
  deny-migration map/test and internal/template/deployer_mode.go are
  byte-unchanged; update.go's 57-line delta renames the cleanup preview to
  reconciliation preview and relocates the migrateProjectCommonAssets
  removal arm INTO the sync flow AFTER the confirmation gate (gated on
  `userAssetsInstalled`) — M1-b's sync fixture must keep that ordering in
  mind, and the skip-path block still carries no integrity probe
  (re-read verbatim-identical).
- Card evidence: `/tmp/moaikr-force-update.log` (still present at plan
  time; run phase should copy it to `.moai/reports/t1578/` before it can
  be cleaned) + the research.md divergence record.
- Both reported defects verify as already repaired upstream (F1/F2 in
  spec.md A.2); the only NEW implementation is the version-match integrity
  probe (K3). Item (3) is out of scope with rationale (spec.md C.1).

## B. Known Issues Carried Into Run

- K1 (card item 1): ALREADY REPAIRED by t1569 M2 (#1792, 87da06367). No
  migration code changes authorized. Work = measurement + guard
  confirmation.
- K2 (card item 2): suspected mechanism (plugin exclusion walk) retired by
  SPEC-USER-ASSET-INSTALL-001. Work = fixture measurement + guard. If M1
  measures an actual empty-dir producer on the current tree, escalate M2 to
  repair per spec.md R1.
- K3 (card item 4): the skip path returns with no content check — the one
  implementation item.
- K4 (card item 3): out of scope (spec.md C.1). Run phase does nothing for
  it beyond not regressing the merge tests.

## C. Pre-flight

- C1: `git rev-parse --short HEAD` → expect `2aab5f797` (record the
  measured value even if it moved again; re-pin RED cells to the measured
  tree).
- C2: copy `/tmp/moaikr-force-update.log` to `.moai/reports/t1578/` (it is
  the only surviving copy of the production evidence).
- C3: confirm the plan-phase baseline still holds:
  `go test ./internal/cli/ -run 'TestRunUpdate_V3Path_NormalizesLegacyRootDenyEntries' -count=1`
  → ok, exit 0.

## D. Constraints (binding on every milestone)

- Go-only: `internal/cli`, `internal/userassets`. Template tree untouched;
  if a milestone discovers a template-side need, STOP and report — that is
  a scope change (spec.md E).
- No new CLI flags, no config keys. The probe is always-on advisory.
- Probe cost constant and small: a fixed file set, existence + (for
  settings.json) parse check. No walking, no hashing.
- Tests join the existing families: update-path fixtures reuse the
  `runUpdateInFixture` / `writeV3ProjectFixture` harness of
  internal/cli/update_deny_migration_test.go; installer tests join
  internal/userassets/installer_test.go patterns (t.TempDir, no parallel
  chdir fixtures in cli).
- Every AC flips through its acceptance.md two-cell record; the probe test
  must be observed RED-adjacent first (the structural evidence in EV-4 is
  the plan-phase RED), never adopted green-only.

## E. Self-Verification Plan (maps to progress.md §E)

- E1 (per milestone): the milestone's own test commands, verbatim output
  into §E.2 evidence.
- E2 (run exit): scoped families green —
  `go test ./internal/cli/ -run 'TestRunUpdate_V3Path|TestStripRetiredV2Deny|TestUpdate.*Integrity' -count=1`,
  `go test ./internal/userassets/ -count=1`, `gofmt -l` on touched files,
  `golangci-lint run` on touched packages.
- E3: the skip-path structural evidence RE-MEASURED after M3 (same
  command as EV-4) showing the probe call inside the `syncSkipped` block —
  this is AC-UMF-001's flip witness.
- E4: no template parity obligations (no template changes) — record the
  `git diff --stat` scope proof instead.

## F. Milestones

### M1 — RED-first reproduction attempt + classification (Priority High; first)

Deliverable: a measurement record (progress.md §E.2 seed + a findings
section) answering, with observed output, exactly three questions. This
milestone writes NO implementation code.

- M1-a (K1 reproduction attempt): run the 9-form normalization through the
  real update path (C3 command). Expected per research: GREEN (already
  repaired). Record: classify K1 "repaired-by #1792; guard exists" — or,
  if RED, classify K1 as live and hand M2 the failing output.
- M1-b (K2 reproduction attempt, in-repo): build a fixture exercising a
  full template-sync cycle (the `runUpdateInFixture` harness) and assert,
  post-run, that no directory under the fixture's managed skill/agent
  roots is zero-file while the catalog ships content for it. Expected per
  research: the project payload excludes those roots entirely, so the
  assertion holds vacuously-by-design — record that the OBSERVABLE defect
  surface moved to the user-folder installer, and run the installer
  equivalent: install the common-asset selection into a temp user root and
  assert every installed directory target received at least one file
  (REQ-UMF-006's probe). If this measures an empty-directory producer,
  M2 becomes a repair milestone for the measured path (spec.md R1) and
  the classification record says so; otherwise record "mechanism retired;
  guard to be pinned in M2".
- M1-c (K3 RED confirmation): re-run the EV-4 structural command
  (acceptance.md evidence ledger) on the current HEAD and paste verbatim
  output — the skip-path block with no probe call. This is the RED anchor
  AC-UMF-001/002 flip from.

### M2 — Regression pins for the repaired mechanisms (Priority Medium)

Deliverable: tests that keep K1/K2 from regressing silently. Implementation
surface: test files only, unless M1-b escalated to repair.

- M2-a (K1): confirm the existing pin —
  `TestRunUpdate_V3Path_NormalizesLegacyRootDenyEntries` already carries
  all 9 forms in its own (implementation-independent) specifier lists.
  Run-phase work is verification + a one-line progress note; add code only
  if the pin is found weaker than described (then extend the fixture, do
  not touch the migration map).
- M2-b (K2): add the two guards as tests:
  - `TestTemplateSync_LeavesNoEmptyManagedSkillDirs` (internal/cli): the
    M1-b fixture, assertion promoted to a permanent test. The assertion
    must count what it swept and fail on a zero-count sweep (no empty
    catalog roots in a fixture = fixture problem, not a pass).
  - `TestInstaller_RejectsEmptyDirectoryTargets` or equivalent
    (internal/userassets): the M1-b installer probe as a permanent test,
    plus — only if M1-b measured a real producer — the REQ-UMF-006
    reporting behavior (installer result counts an empty dir target as
    failed-with-reason, not installed).
- M2-c: `go test ./internal/cli/ ./internal/userassets/ -count=1` scoped
  green; record verbatim output.

### M3 — Version-match integrity probe (Priority High)

Deliverable: REQ-UMF-001..003 implemented. File-level change map:

- `internal/cli/update_integrity_probe.go` (new):
  `runManagedSurfaceIntegrityProbe(out io.Writer, projectRoot string)`
  — walks a fixed representative set, defined in this file as a named
  slice with a comment binding each entry to why it is representative
  (plan default, decision-index Q2): `.claude/settings.json` (exists AND
  parses as JSON), `.moai/config/sections/system.yaml` (exists,
  non-empty), `.moai/manifest.json` (exists, parses as JSON). Output: one
  `[update] Integrity: <path> (<reason>)` warning line per damaged entry
  via the existing `tui.CheckLine("warn", ...)` channel so it renders
  through the standard update UX; internal errors degrade to a single
  warning line (REQ-UMF-002, fail-open).
- `internal/cli/update.go` (one insertion): inside the
  `if syncSkipped {` block, before `stripRetiredModelConfigOnVersionMatch`
  or immediately after it (order: probe first — it is pure observation;
  the strip is mutation), guarded to the version-match entry condition
  only: the syncSkipped flag currently conflates version-match and
  user-cancelled merge, so M3 threads the distinction the helper already
  re-evaluates (REQ-UMF-003) — the probe runs on version-match, is
  skipped on cancellation.
- `internal/cli/update_integrity_probe_test.go` (new):
  - `TestRunUpdate_VersionMatch_RunsIntegrityProbe`: fixture (the
    runUpdateInFixture harness), version-matched double update; second
    run prints no integrity warnings on an intact fixture (false-positive
    guard) and the probe fires when a representative file is deleted
    (assert the warning names the path; exit stays 0 — REQ-UMF-002).
  - `TestRunUpdate_UserCancelled_SkipsIntegrityProbe`: the cancelled-merge
    entry prints no integrity line (REQ-UMF-003).
  - `TestIntegrityProbe_FailOpen`: a representative path made unreadable
    (chmod 000 where permitted, or a directory-as-file) yields one
    warning line, no error return.
- Ordering rationale (reversibility): the probe's FILE SET and line format
  are the decisions most likely to be revised — they sit at the top of
  this file with the named-slice constant so a revision is one edit; the
  update.go insertion is mechanical and last.

### M4 — Convergence verification (Priority High)

- Full scoped batch (E2), the re-measured EV-4 flip witness (E3), scope
  proof (`git diff --stat` — Go-only), and the §E.2 evidence write-up.
- Re-run C3 once more as the K1 guard's final witness.

## G. Anti-Patterns (named refusals)

- Do NOT "fix" the already-green 9-form migration by editing
  `legacyRootDenyNormalizeMap` — that is scope K1 reverting to
  re-implementation (spec.md Out of Scope).
- Do NOT implement the probe as a blocking gate or a new exit code path —
  REQ-UMF-002 forbids it.
- Do NOT widen the probe into a catalog diff or content hash — spec.md
  Out of Scope (probe scope growth).
- Do NOT create settings.local.json migration scaffolding "while in the
  area" — K4 is out of scope (agent-common scope discipline).
- Do NOT adopt a green test whose selector matched zero tests — count the
  swept set in every new test's failure messages.

## H. Cross-references

- spec.md (requirements REQ-UMF-001..006, scope decisions, Out of Scope)
- acceptance.md (two-cell ACs + evidence ledger EV-1..EV-4)
- research.md (divergence record)
- decision-index.md (Q1..Q3)
- internal/cli/update_deny_migration_test.go (fixture harness to reuse)
- internal/userassets/install.go (Installer.Install / applyTarget — M2-b
  surface)
