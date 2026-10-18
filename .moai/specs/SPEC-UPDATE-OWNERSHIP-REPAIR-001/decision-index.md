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
Options (factual; no recommendation):
- **(가)** Stamp `deployed_version` on each entry at deploy time, refresh the manifest-level `Version`/`DeployedAt` at the update's deploy-stage save, and write `<runDir>/manifest.prior.json` in the Backup stage. Evidence-complete unregistered files can then be judged `managed`, and a restore can put the manifest back with the files. This is the authored text of REQ-OWR-015 and plan.md M1b.
- **(나)** Produce no new evidence. Every unregistered file stays `unregistered` in production; the `managed` (attested) path exists in the verdict function but is reachable only from a fixture that supplies the evidence. REQ-OWR-015 and the manifest-copy half of REQ-OWR-010 are dropped, and AC-OWR-05, AC-OWR-06 and AC-OWR-15 shrink to their fixture-only halves.
Operator verdict:

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
Operator verdict:
