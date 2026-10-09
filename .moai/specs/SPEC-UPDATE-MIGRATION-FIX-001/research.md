# SPEC-UPDATE-MIGRATION-FIX-001 — Research Record

Purpose: the card text's assumptions diverged from the codebase reality in
four material ways. This file records what was verified, with the commands
and outputs, so the plan and the run phase reason from the tree rather than
from the card prose. Canonical baseline: worktree
`.moai/worktrees/t1578` at HEAD 2aab5f797, 2026-10-09 — the tree
fast-forwarded from 81786284e to 2aab5f797 mid-research; every
conclusion below was re-verified on 2aab5f797 (acceptance.md EV-6: the
deny-migration file pair, deployer_mode.go, and
internal/userassets/install.go are byte-unchanged across the re-cut;
update.go's delta does not touch any conclusion).

## D1 — Card item 1: the "missing" colon-star forms are already in the migration map

Card claim: "the legacy specifier pattern set (9 forms) misses the
colon-star family ... '/:*' · '~:*' · 'C:/:*' (3 forms) plus del/rmdir
'C:/:*' (2 forms) survive unmigrated. Fix direction: add the colon-star
forms to the pattern set AND complete the bare/star pairs for the regular
6 forms."

Tree reality:

```
$ grep -n "normalizeLegacyRootDenySpecifiers" internal/cli/*.go
internal/cli/update.go:432:         if normErr := normalizeLegacyRootDenySpecifiers(cwd, out); normErr != nil {
internal/cli/update_clean_install.go:568: ... (same call on the clean-reinstall path)
internal/cli/update_deny_migration.go:146: func normalizeLegacyRootDenySpecifiers(...)
```

`legacyRootDenyNormalizeMap` (internal/cli/update_deny_migration.go:121)
carries exactly 9 keys: the 5 colon-star forms the card names as surviving
(`Bash(rm -rf /:*)`, `Bash(rm -rf ~:*)`, `Bash(rm -rf C:/:*)`,
`Bash(del /S /Q C:/:*)`, `Bash(rmdir /S /Q C:/:*)` — plus
`Bash(Remove-Item -Recurse -Force C:/:*)`) and the 3 escaped-star forms;
values are the bare/star canonical pairs matching the shipped template
deny set (`settings.json.tmpl` lines 560-568: 6 Bash rm forms + 3 Windows
star forms). The call site in update.go sits BEFORE the version-match
short-circuit, so a same-version update also migrates (comment + test
placement assertion). The test
`TestRunUpdate_V3Path_NormalizesLegacyRootDenyEntries` pins all 9 forms
through the real update path with test-owned specifier lists.

Measured: EV-1 (acceptance.md) — `ok ... 7.715s`, exit 0.

Timeline reconciliation (why the card saw survivors):

```
$ git show 87da06367 -s --format='%ci | %s'
2026-10-07 05:22:36 +0000 | fix(update): migrate legacy root-denial settings rules to canonical forms on update (card t1569 M2) (#1792)
```

- The migration was born complete in ONE commit (7a0fd5be1, merged
  #1792) — there was never a smaller map that "missed" the colon-star
  family (`git show 7a0fd5be1:internal/cli/update_deny_migration.go`
  shows the identical 9-key map).
- The mo.ai.kr run used the rc.26 binary (log line 1), built before that
  merge; its run timestamp (05:52:41 UTC from the backup dir name) is 30
  minutes after the merge but the binary predates it. rc.26 carried no
  root-denial migration at all.
- Classification: the verification-completeness doctrine's "Deployment"
  quiet case — "a fix landed and the installed binary was built from an
  older commit ... a card was issued for a defect that was already
  repaired."

## D2 — Card item 2: the suspected deploy-path defect mechanism no longer exists

Card claim: "Claude-side skill re-deploy creates directories but omits
contents ... Suspected file-write defect in the Claude-side deploy path
(deployer / bundle code under internal/template/ or internal/userassets/)."

Tree reality: the rc.26-era deployer branched on deploy mode, and plugin
mode excluded Claude-side skill files from the project payload — the
plausible producer of dir-skeleton-without-files. That entire branch is
retired by SPEC-USER-ASSET-INSTALL-001 (status: completed; PR #1772 — the
merge commit of this worktree's base):

```
$ sed -n '5,17p' internal/template/deployer_mode.go
(see acceptance.md EV-3 — "The plugin-mode split is RETIRED with its
carrier: the plugin payload constant, the mirror policy, the exclusion
walk branch, ... are all gone")
```

`isCommonAssetRoot` now excludes `.claude/skills/`, `.claude/agents/moai/`,
`.agents/skills/`, `.codex/agents/moai/` from the project payload in every
mode; common assets install through `internal/userassets` (Installer.Install,
install.go) into per-user folders. The template tree still carries
`internal/template/templates/.claude/skills/moai-lane-watchdog/SKILL.md`
(11,959 bytes) as the installer's source content, and the current
`.moai/manifest.json` lists
`.claude/skills/moai-lane-watchdog/SKILL.md`.

Consequence: the observed defect is not reachable through the old branch on
this tree; whether the NEW installer can produce an empty directory target
is a measurement (plan M1-b), not an assumption. No test in the update
test families currently asserts the absence of empty directories under
managed skill roots (precise grep over internal/cli/update*_test.go: the
only "empty directory" assertions concern the backup-dir cleanup, not
managed roots).

Re-cut addendum (observed in the 81786284e..2aab5f797 delta): the
`migrateProjectCommonAssets` removal arm was relocated INSIDE the template
sync flow, AFTER the confirmation gate, and is gated on
`userAssetsInstalled` (a failed or skipped user-side install must leave
every project-side asset in place; cancelling the update leaves managed
assets byte-intact). This strengthens the K2 story — the project-side
asset lifecycle now has an explicit ordering contract — and M1-b's sync
fixture must exercise it with the user-asset phase outcome as the gate
input, per plan.md A.

## D3 — Card item 4: the sync-skip path is verifiably content-blind

"Up to date · Skipping sync" is the `if syncSkipped {` early return in
runUpdate (acceptance.md EV-4, verbatim block). It runs two side-steps
(retired model-key strip per card t1527 D5; participation ask per
SPEC-FEEDBACK-PARTICIPATION-001 REQ-ANON-004) and returns. Both share the
established disposition "A failure warns; it never fails the update" — the
pattern the new probe follows (REQ-UMF-002).

Note: `syncSkipped` conflates two entry conditions (version match, and a
user-cancelled merge — the helper re-evaluates the predicate, per the
in-block comment). REQ-UMF-003 requires M3 to distinguish them so a
cancelled run is not told its sync "completed with integrity OK".

## D4 — Card item 3: the settings model is already merge-based, not pure-template

The review item hypothesizes the model "settings.json = pure template (old
backed up then deleted, new copied) + user values in settings.local.json"
and asks whether a local-file migration is needed. Tree reality: settings.json
is a 3-way MERGE target — the merge derives its base so template keys reach
existing projects while user keys survive
(TestRunUpdate_V3Path_CleanSettingsUntouched, its idempotency arm, and the
in-test comment "settings.json is a merge target"). The mo.ai.kr
production run itself preserved 32 user settings keys through a force
update (log lines 42-49). The purity premise does not hold on this tree;
the disposition is out-of-scope + follow-up card candidate (spec.md C.1,
decision-index Q3).

## Gaps (explicitly not verified)

- mo.ai.kr's actual settings.json deny list was never re-measured (remote
  production project, not reachable from this worktree). F1's
  classification rests on the tree + test evidence, not on the production
  file. Residual: an exotic JSON-escaped variant of a legacy specifier
  would not exact-match the map keys — unknowable from here; the migration
  is exact-string by design and reports its count.
- The rc.26 binary's embedded manifest was not inspected (no artifact
  available); the D2 mechanism account is reconstructed from the retired
  branch's own comment plus the production log, not from the rc.26 tree.
- `/tmp/moaikr-force-update.log` is volatile storage; plan.md C2 directs
  the run phase to copy it into `.moai/reports/t1578/` first.

## Residual risks

- The probe's representative file set could grow into a maintenance
  burden; bounded by spec.md E (fixed small set, constant cost).
- If the user-folder installer CAN leave an empty directory target (M1-b
  measures), the mo.ai.kr empty-dir class has a second, current producer —
  spec.md R1 names the escalation branch.
