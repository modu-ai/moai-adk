---
id: SPEC-UPDATE-MIGRATION-FIX-001
title: "Update migration follow-ups from the mo.ai.kr production run: verify the two reported defects against the current tree, pin regressions, and add the version-match integrity probe"
version: "0.1.0"
status: completed
created: 2026-10-09
updated: 2026-10-10
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: "internal/cli (update), internal/template (deploy surfaces), internal/userassets"
lifecycle: spec-anchored
tags: "update, migration, deny-specifiers, user-assets, integrity-probe, settings-purity, card-t1578"
tier: M
related_specs: [SPEC-USER-ASSET-INSTALL-001]
---

# SPEC-UPDATE-MIGRATION-FIX-001 — Update Migration Follow-ups (card t1578)

## HISTORY

- 2026-10-09: v0.1.0 created by manager-spec (card t1578, plan phase, Tier M).
  Card P1/P2 evidence originated in a real mo.ai.kr production run
  (v3.2.0-rc.26, `moai update --force --yes`, exit 0, 2026-10-07 14:52 KST,
  `/tmp/moaikr-force-update.log`). Plan-phase research found BOTH reported
  defect mechanisms already repaired upstream on this tree; the scope was
  re-authored accordingly (verification + regression pinning + one new
  integrity probe). Research record: research.md. Scope decisions for the
  operator's two review items (settings purity, sync-skip integrity check)
  are settled in Section 3 and decision-index.md.

## A. Background and Evidence Baseline

### A.1 The production observation (card evidence)

On 2026-10-07 14:52 KST, `moai update --force --yes` (v3.2.0-rc.26, deploy
mode: plugin) ran on the mo.ai.kr project and exited 0 while reporting
"220 managed re-deployed". Two anomalies were measured afterwards:

1. The legacy root-denial deny specifiers (colon-star family:
   `Bash(rm -rf /:*)`, `Bash(rm -rf ~:*)`, `Bash(rm -rf C:/:*)`,
   `Bash(del /S /Q C:/:*)`, `Bash(rmdir /S /Q C:/:*)`) survived in
   settings.json unmigrated.
2. `.claude/skills/moai-lane-watchdog` was an empty directory (0 files)
   while the `.agents/skills` catalog carried contents.

### A.2 Plan-phase findings (measured on this tree, baseline 2aab5f797 after a mid-research re-cut from 81786284e; anchors re-verified — acceptance.md EV-6)

Finding F1 (card item 1 — premise already repaired). The migration the card
asks to extend landed in the interval between the observation and this
card's dispatch:

- `normalizeLegacyRootDenySpecifiers` was born complete in commit
  `87da06367` / PR #1792 (card t1569 M2), merged 2026-10-07 05:22:36 UTC —
  30 minutes BEFORE the rc.26 run at 05:52:41 UTC, but the rc.26 binary was
  built earlier and carries none of it.
- The `legacyRootDenyNormalizeMap` (internal/cli/update_deny_migration.go)
  contains all 9 legacy forms INCLUDING the 5 colon-star forms the card
  names as surviving; canonical targets are the bare/star pairs matching
  the shipped template deny set exactly.
- Call sites cover both the plain v3 update path (before the version-match
  short-circuit, so a same-version update still migrates) and the
  clean-reinstall path.
- Measured: `go test ./internal/cli/ -run
  'TestRunUpdate_V3Path_NormalizesLegacyRootDenyEntries' -count=1` →
  `ok  github.com/modu-ai/moai-adk/internal/cli  2.866s`, exit 0 on the
  re-cut baseline 2aab5f797 (first run `ok ... 7.715s` straddled the
  re-cut; see acceptance.md EV-1/EV-6 attribution note).

This is precisely the "Deployment" quiet case of the verification-
completeness doctrine: a fix landed while the installed binary predated it,
so the pre-fix behavior was reproduced live and a card was issued for a
defect that was already repaired.

Finding F2 (card item 2 — suspected mechanism retired). The rc.26-era
deployer branched on deploy mode; in plugin mode the project payload
excluded Claude-side skill files. That entire surface is retired by
SPEC-USER-ASSET-INSTALL-001 (status: completed; PR #1772, merged into this
tree's base): `deployer_mode.go` states "The plugin-mode split is RETIRED
with its carrier", the deployer carries a single project payload shape, and
`isCommonAssetRoot` excludes `.claude/skills/`, `.claude/agents/moai/`,
`.agents/skills/`, `.codex/agents/moai/` from the project payload — common
skills and agents now install through the per-user-folder installer
(internal/userassets). The empty-dir defect class is therefore not
reachable through the old branch; whether the new installer can produce an
empty directory is what M1 must measure, not assume.

Finding F3 (card item 4 — the gap that made the damage invisible). The
version-match short-circuit in `runUpdate` ("Up to date · Skipping sync")
returns after two side-steps (retired model-key strip, participation ask)
with no content check: a damaged managed surface stays damaged through any
number of non-force updates. This is the one genuinely new implementation
item.

Finding F4 (card item 3 — purity model is already merge-based). The current
settings model is a 3-way merge target (user values preserved; the mo.ai.kr
run itself preserved 32 user settings keys), not the pure-template model
the review item hypothesizes. See Section 3 for the disposition.

### A.3 Known-Issues summary

| ID | Card item | Status on this tree | This SPEC's work |
|----|-----------|---------------------|------------------|
| K1 | (1) P1 colon-star migration | Repaired by t1569 M2 (#1792) | M1 verification record + M2 guard confirmation |
| K2 | (2) P2 empty skill dir | Old branch retired by SPEC-USER-ASSET-INSTALL-001 | M1 fixture measurement + M2 no-empty-dir guard |
| K3 | (4) review: integrity check | Absent (new capability) | M3 implement probe on the skip path |
| K4 | (3) review: settings purity | Merge model already preserves user values | Out of scope (Section 3), follow-up card candidate |

## B. Requirements (GEARS)

### B.1 Integrity probe on the version-match short-circuit (new capability)

- REQ-UMF-001 (event-driven): When an update run takes the version-match
  short-circuit (`syncSkipped` == true from a version match, not a user
  cancellation), the system shall run the managed-surface integrity probe —
  a fixed, documented set of representative managed project paths — before
  returning, and shall emit one greppable warning line naming each
  representative path found missing or damaged.
- REQ-UMF-002 (unwanted): The integrity probe shall not fail the update
  run; a probe internal error is a warning line, never a non-zero exit.
- REQ-UMF-003 (state-driven): While the update run is a user-cancelled
  merge (the same `syncSkipped` return), the system shall not report
  integrity findings as if the run had completed a sync — the probe
  distinguishes the two entry conditions or is skipped on cancellation, per
  the helper contract fixed in plan.md Section F (M3).

### B.2 Regression guards on the two repaired mechanisms

- REQ-UMF-004 (event-driven): When a settings.json deny list carries any of
  the 9 legacy root-denial specifiers, the plain update path shall replace
  each with its canonical template form. (Guard: behavior exists today via
  card t1569 M2; M1 records the measurement, M2 confirms the test family
  remains the pin. No new migration logic is authorized by this SPEC.)
- REQ-UMF-005 (event-driven): When a template sync completes in a fixture
  project, the system shall have left no zero-file directory under a
  managed skill or agent root for which the template catalog carries
  content.
- REQ-UMF-006 (capability-gated): Where the per-user-folder installer
  installs a directory target from the common-asset catalog, the installer
  shall verify the target received at least one file and shall report any
  empty install target rather than counting it as installed.

## C. Scope Decisions (operator review items — settled)

### C.1 Item (3) settings purity model — OUT OF SCOPE, follow-up card candidate

The review item asks whether a migration is needed to move user values
that old runtimes mixed into settings.json into settings.local.json, under
the model "settings.json = pure template + user values in
settings.local.json". Disposition:

- The current tree enforces a 3-way MERGE model for settings.json (the
  merge derives its base so template keys reach existing projects while
  user keys survive — pinned by TestRunUpdate_V3Path_CleanSettingsUntouched
  and its idempotency arm). The mo.ai.kr run measured 32 user settings
  keys preserved through a force update. The purity premise ("user values
  are stranded in settings.json with no protection") does not hold.
- Migrating to pure-template + settings.local.json is an architecture-level
  change to the settings merge engine (touching init, update, doctor, the
  merge-back preserve inventory, and every settings test family) with a
  broad regression surface and a real user-communication cost (existing
  projects' settings.json would change shape on update). It deserves its
  own operator decision with survey data, not a rider on this card.
- Interim disposition recorded here: no migration. Follow-up card candidate
  text: "Settings purity survey: measure how many deployed projects carry
  user-authored keys in settings.json, then decide whether a
  settings.local.json migration SPEC is warranted" (EVIDENCE-NEEDED, see
  decision-index.md Q3).

### C.2 Item (4) integrity check — IN SCOPE

Brought in as REQ-UMF-001..003 / M3. Rationale: small bounded surface (one
probe on one early-return path) and explicitly the operator's review
direction. The probe is warn-and-continue, matching every existing
advisory step on the update path.

Wording-honesty scope note (plan-audit D5): the minimal 3-file probe set
detects NEITHER observed production damage class — the mo.ai.kr empty
skill directory lies outside the probed set, and the legacy-specifier
survival behind K1 is valid JSON that parses cleanly. What the probe
actually indicates is gross structural loss of the managed project core
(a missing or unparseable settings.json / system.yaml / manifest.json) —
a canary, not a damage-class detector. A damage-class-targeted probe set
is an OPERATOR DECISION, recorded as a follow-up card candidate (see
decision-index.md Q2); this SPEC does not re-decide the set.

## D. Acceptance Criteria

Release-blocking and regression-guard criteria are enumerated in
acceptance.md in TWO-CELL form (RED-now cell + green-path cell), with the
plan-phase evidence ledger carrying command, verbatim output, exit code,
and tree SHA (canonical baseline 2aab5f797; per-entry attributions
corrected for the mid-research re-cut) for every cell measured during this
plan phase.

- AC-UMF-001 (release-blocking): the version-match skip path runs the
  integrity probe (RED: structural — the skip-path block contains no probe
  call today).
- AC-UMF-002 (release-blocking): the probe reports a damaged representative
  path by name and exits 0 (RED: same structural evidence).
- AC-UMF-003 (regression-guard): the 9-form root-denial normalization stays
  covered and green on the plain update path. (Historical RED is not
  re-executable on this tree — undecidable disposition per
  verification-completeness §2.1; classified regression-guard, never
  recorded as a fix.)
- AC-UMF-004 (regression-guard): a full sync cycle leaves no zero-file
  directory under a managed skill/agent root that the catalog ships with
  content; the user-folder installer reports empty directory targets.
- AC-UMF-005 (plan-gate): this SPEC records the item-(3)
  out-of-scope disposition with rationale (satisfied by Section C.1;
  classification per acceptance.md — no RED-now cell applies).

## E. Constraints

- Go-side changes only (internal/cli, internal/userassets). No template
  tree changes are authorized, so no `make build` template-regeneration
  step is required by this SPEC; the run phase re-measures if that
  assumption breaks.
- The probe set is a FIXED, small, documented list — it must stay
  constant-cost (advisory-check discipline) and must not grow into a full
  catalog diff.
- Fail-open: probe errors warn; the update exit code is unchanged by probe
  outcomes (REQ-UMF-002).
- English artifacts and code comments; conventional commits carrying the
  card id t1578.
- No code changes in plan phase; artifacts only.

## F. Out of Scope

### Out of Scope — settings purity / settings.local.json migration

- No migration of user values out of settings.json into
  settings.local.json is authorized by this SPEC (rationale: Section C.1;
  the current merge model already preserves user values).
- No change to the settings 3-way merge engine, the merge-back preserve
  inventory, or doctor's settings handling.

### Out of Scope — re-implementing the two repaired mechanisms

- No new legacy-denial-specifier migration logic, no changes to
  `legacyRootDenyNormalizeMap`, and no changes to the t1569 M2 call-site
  placement (already complete on this tree; M2 only confirms the guard
  tests).
- No resurrection or re-implementation of the retired plugin deploy mode,
  the `.agents/skills` mirror policy, or any other carrier retired by
  SPEC-USER-ASSET-INSTALL-001.
- No repair of the remote mo.ai.kr project itself — its settings.json
  heals on its next update with a build carrying t1569 M2; its stale skill
  dirs are addressed by the installer's own contract, not by this SPEC.

### Out of Scope — probe scope growth

- No full catalog-vs-disk diff, no content hashing of deployed files, and
  no probe of user-folder install roots in M3's minimal set (named as the
  decision-index Q2 alternate for a future round).

## G. Risks

- R1: M1's fixture measurement contradicts Finding F2 (the new installer
  CAN leave an empty directory). Then M2 escalates from guard to repair and
  REQ-UMF-006 becomes the fix target. Branch is explicit in plan.md M1.
- R2: The probe's representative set drifts out of date as the template
  evolves. Mitigation: the set lives beside the probe with a comment
  binding it to the template's stable project files, and M2's guard test
  asserts the set members are template-managed paths.
- R3: Residual unknown — mo.ai.kr's actual settings.json deny list was
  never re-measured remotely; the F1 classification rests on the tree
  evidence and the test measurement, not on the production file.

## H. Cross-references

- research.md — the full divergence record (commands + outputs).
- acceptance.md — two-cell AC enumeration + evidence ledger.
- decision-index.md — Q1 probe failure disposition, Q2 representative file
  set, Q3 purity-migration funding.
- Card t1569 / PR #1792 — the upstream repair this SPEC verifies (K1).
- SPEC-USER-ASSET-INSTALL-001 — the architecture change that retired the
  K2 mechanism; its M2 installer is the surface REQ-UMF-006 guards.
- `.claude/rules/moai/development/verification-completeness.md` §1.3 — the
  Deployment quiet case that classifies K1/K2.
