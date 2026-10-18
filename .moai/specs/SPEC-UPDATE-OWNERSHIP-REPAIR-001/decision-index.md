# SPEC-UPDATE-OWNERSHIP-REPAIR-001 — Decision index

Stateless on the status axis: the lifecycle lives in spec.md. Each row follows Detect → Explain → Ask. No row carries an embedded recommendation; a `Default:` is a policy application of the published Default rule, not a recommendation. Authority anchors are committed artifacts; an anchor that cannot be verified routes the row to FOUNDER. Route labels in use: DECIDED, POLICY-COVERED, EVIDENCE-NEEDED, FOUNDER.

### Q1: Does a profile-mapped project keep its shipped common-asset copies regardless of whether a user-root counterpart exists?

Label: DECIDED
Authority anchor: `.moai/specs/SPEC-UPDATE-REPAIR-RC30-001/decision-index.md` § Q2 (operator decision d-20261010T060848Z-dc3b; the SPEC's status is `completed`, `spec.md:5`). Resolve with `git show HEAD:.moai/specs/SPEC-UPDATE-REPAIR-RC30-001/decision-index.md`.
Why unresolved: Not unresolved. The identical question under identical conditions (a project the launch ledger maps to a profile) is answered there: option (c), the project keeps its shipped copies and the kept copy carries a provenance mark so a later doctor can identify and reclaim it. REQ-OWR-004 follows it and adds only a recorded counterpart state; the `KeptBy` mark's meaning is unchanged.
Operator verdict:

### Q2: May the update stamp the deploying binary's version on every manifest entry it writes, and retain the manifest as loaded at the start of each run in that run's backup directory?

Label: FOUNDER
Class: product-level
Why unresolved: The three-part evidence rule (deployed version, original deployed path, prior manifest record) cannot be met by any file today. The manifest records no per-file deployed version, and `Manifest.Version` is written only at project initialization (`internal/core/project/initializer.go:607-608`); no prior manifest generation is retained anywhere (research.md D(1), D(7)). Producing that evidence adds a field to `.moai/manifest.json`, a file users commonly commit, and adds a file to each run's backup directory, so it changes what a shipped command writes. No committed record or setting decides it. The published Default rule is not applied to this row: its first-ranked option (keep current behavior, option (나)) leaves the card's own required positive fixture — an evidence-complete file judged managed — unrealizable in production, so ranking it would be ranking an option the card excludes.
Options (factual; no recommendation; the row is decided as (가) below — the option text stays as the record):
- **(가)** Stamp `deployed_version` on each entry at deploy time, refresh the manifest-level `Version`/`DeployedAt` at the update's deploy-stage save, and write `<runDir>/manifest.prior.json` in the Backup stage. Evidence-complete unregistered files can then be judged `managed`, and a restore can put the manifest back with the files. **What it changes in the field today: nothing for the 313 existing `.agents/skills` files** — the evidence (deployed version, prior record) exists only for files deployed after the change, and the copy-fallback mirrors are never registered at all, so those 313 stay `unregistered` and kept; the positive path becomes reachable only for a file that a later run deployed under a retained generation and whose entry was then lost. What it does change from the next run on: entries that run deploys carry the deploying version, the manifest-level Version is refreshed, every run backup holds one `manifest.prior.json`, and a run whose backup fails removes no file. This is the authored text of REQ-OWR-015, REQ-OWR-016 and plan.md M1b. A failed `manifest.prior.json` write blocks removal for that run (fail-closed).
- **(나)** Produce no new evidence. Every unregistered file stays `unregistered` in production; the `managed` (attested) path exists in the verdict function but is reachable only from a fixture that supplies the evidence.

Consequences of (나), complete (plan.md § G.1 carries the same table; the SPEC is built so that (나) is a pure subtraction of the folded items and a reduction of the reduced ones):
- Folded away entirely: REQ-OWR-015 and REQ-OWR-016; AC-OWR-06 and AC-OWR-20; milestone M1b (the `deployed_version` field, the manifest-level `Version`/`DeployedAt` refresh, `manifest.prior.json`, the prior-generation reader, the backup-failure hand-off).
- Reduced: REQ-OWR-010 to the file-bytes half (no manifest copy); REQ-OWR-008 to the "this run emptied it" half, which t1668's `pruneEmptiedSkillDir` already does, so milestone M5 adds no production code; REQ-OWR-013 loses its emptied-directory clause for directories an earlier run emptied; AC-OWR-15 loses its manifest-copy assertions; AC-OWR-13 loses cases A and C and keeps B, D, E, F as regression guards of t1668 behavior; AC-OWR-18 loses sub-case (c); AC-OWR-17 loses the `DeployedAt` mutant.
- Survives unchanged: REQ-OWR-001 to -007, -009, -011, -012, -014; AC-OWR-01 to -04, -07 to -12, -14, -16, -19; and AC-OWR-05 at function level, because the evidence interpreter and its interface live in M1a, not in M1b.
- Operator-visible: the field symptom of three emptied directories left by earlier runs stays unrepaired in production, and the positive path of card requirement (6), an evidence-complete file judged `managed`, is unreachable in production and exercised only from a fixture. Both card items (5) and (6) are then delivered in part by design.
Operator verdict: DECIDED-BY-LEADER (a), CONDITIONAL — board d-20261018T040918Z-def2, 2026-10-18, source: leader, from the operator's design recommendation checked against decision-index Q2. A design decision only: NOT a user security approval, NOT approval to touch the real home, and it does NOT pre-approve Kickoff (Kickoff is decided after the independent plan audit). Conditions folded into REQ-OWR-015 to -017, design.md D-3, AC-OWR-06, -20, -21: (1) `deployed_version` only on entries this run deployed, never back-filled; manifest-level Version/DeployedAt refreshed at the deploy-stage save. (2) The load-time manifest's original bytes go to `<runDir>/manifest.prior.json`; if any backup (file bytes or manifest copy) fails, that run removes zero files. (3) Atomic creation (temp + rename), manifest.Save's temp + rename kept, a partial or failed write never reads as a valid generation, a corrupt or unreadable generation is no evidence. (4) `priorManifestGenerations` = 8 is a lookup bound only, no new auto-delete or shortened retention. (5) Old JSON still reads, same-version repeat leaves bytes unchanged, an older binary that drops the field weakens the evidence so the file is preserved. (7) Anything that mixes in deletion, protection relaxation, a separate security permission, `removalEnabled` activation, or card t1702 scope is not carried.

### Q3: What form does the final remaining-file list take?

Label: FOUNDER
Class: implementation-level
Why unresolved: The card asks for a final remaining list with a reason per file. The operator ruling d-20261012T221654Z-9ced (cited in `internal/cli/update_dryrun_preview.go:37-39`) shows preserved files as a count by default and as paths only under `--verbose`, so 313 rows cannot be a default print. No committed record decides whether the list must also persist beyond the run's terminal output.
Options (factual; no recommendation):
- **(가)** The verdict pass returns the rows as a value; the update renders them through the existing terminal reference rows, grouped by reason by default and per file under `--verbose`. No new artifact on disk.
- **(나)** As (가), and also write the rows as a sorted JSON file in the run's backup directory.
Default: (가) (rule: the option with the smaller user-visible surface — no new artifact; the repeated-update comparison reads the same rows through the returned value and the verbose output).
Alternate: (나)
Operator verdict: DEFAULT-APPLIED 2026-10-18T03:07:32Z manager-spec (card t1865, plan close)

### Q4: How far does the update go when a path differs from its manifest key only in letter case?

Label: FOUNDER
Class: implementation-level
Why unresolved: The shipped rename of a lowercase `skill.md` to `SKILL.md` already exists in `internal/cli/update/skillresidue` (shipped skills under `.claude/skills` only, held back until the user confirms). The migration walk reads the on-disk spelling and so misses the registered key (research.md D(4)). No committed record says whether the migration may also rename.
Options (factual; no recommendation):
- **(가)** The verdict function resolves the spelling difference to the registered entry on a case-insensitive filesystem and reports it; the migration performs no rename.
- **(나)** The migration also renames the file to the registered spelling, including copies under `.agents/skills` that `skillresidue` does not cover.
Default: (가) (rule: the option that preserves current behavior — renames stay with `skillresidue`; the alternative adds a second rename owner).
Alternate: (나)
Operator verdict: DEFAULT-APPLIED 2026-10-18T03:07:32Z manager-spec (card t1865, plan close)

### Q5: May the update read a profile-side counterpart (a file under the profile's own skills or agents directory) to decide keep versus migrate for a profile-mapped project?

Label: EVIDENCE-NEEDED
Authority anchor: none — an evidence-needed row carries no anchor.
Why unresolved: Reading a profile-side counterpart means touching a profile root. Whether that read is safe depends on card t1702's open protection verdict: its sync audit ended FAIL on an unresolved P1, namely that the `user-root:claude-skills/` match in `internal/config/protected_zone.go` does not protect `claude-skills@<profile>/` paths. The leader's binding constraint (board d-20261018T020724Z-48fb) keeps this SPEC off profile roots until that evidence exists. Until then this SPEC records the profile-side state as `not-checked` and keeps every profile-mapped copy (Q1).
Leader ruling (2026-10-18, source: leader ruling): resolved as the conservative reading. Profile roots are not read at all, the profile-side state is recorded as `not-checked`, and profile-mapped copies are kept unconditionally. This SPEC is written to that reading (spec.md § B, REQ-OWR-004 and REQ-OWR-005, AC-OWR-07 and AC-OWR-08). The route label stays EVIDENCE-NEEDED: the board line itself and its digest prefix were not available to this author, so the row is not relabeled DECIDED (an anchor that cannot be pinned never upgrades a row). The row is not a FOUNDER row, so its empty operator line does not block the Kickoff.
Operator verdict:

## Accepted debt (plan audit iteration 1)

The following findings are accepted rather than fixed, each with its reason. They are not decision rows.
- Iteration 2 (leader to accept or reject by name): the profile non-access check stays an inference tool. A read through a path built from fragments that no listed symbol or token names, with the result discarded, is undetectable without an access-recording seam (acceptance.md AC-OWR-08 limit (1) and (3), design.md D-8). The seam was not adopted. Decision requested only if the leader wants the seam; otherwise this stands as accepted debt.
- Iteration 2: REQ-OWR-016 requires the migration removal arm to be split into collect, verify every backup, then remove; this widens the merge-conflict surface in `migrate_project_assets.go` beyond the call-line changes (design.md D-3 item 4, acceptance.md G-L).
- Iteration 2: the version-match gate is not changed; recovery and the earlier-run emptied-directory case are reached only by `--force` or a changed template/version run (plan.md OQ-6, acceptance.md G-M).
- D13, remainder: REQ-OWR-005 names the two profile-root directories and the two protected files. They are the leader's binding constraint text, so the requirement keeps them; the other implementation details in requirement wording were removed (REQ-OWR-008, REQ-OWR-015) and the AC-OWR-01 constant-count assertion was replaced by a matrix over every combination.
- D15: `update_template_sync.go` takes call-line changes in M1b, M3 and conditionally M6, and `migrate_project_assets.go` in M1a, M2, M3 and M5, so landing after t1668 and a later merge of main may conflict in those two files. Mitigated by keeping new judgment code in new files and by removing the production change from M8; the residual conflict risk is accepted.
- D2 limits: the source scan misses a path built without any listed token, and the differential canary misses a profile-side read whose result is discarded; both are stated in design.md D-8. A file-access recording seam was considered and not adopted because it would reshape the judgment code and widen the conflict surface in the two files above.
- D9 limit: if the stale-record variant of the commands fixture cannot be driven through an existing seam, the M6 measurement records that and the commands criterion stays a regression guard (plan.md M6).
