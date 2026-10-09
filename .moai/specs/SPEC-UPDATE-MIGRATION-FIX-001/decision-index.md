# SPEC-UPDATE-MIGRATION-FIX-001 — Decision Index

Decision gate: on (`.moai/config/sections/interview.yaml` interview.decision_gate).
One row per decision surfaced during plan assembly that the operator did
not settle in an interview. Labels: DECIDED / POLICY-COVERED /
EVIDENCE-NEEDED / FOUNDER.

### Q1: Should the version-match integrity probe fail the update when it finds damage, or warn and continue?

Label: FOUNDER
Authority anchor: (none — no completed SPEC HISTORY or Amendments row
decides the failure disposition of a NEW advisory step on this path; the
update-path precedent steps cite SPEC-FEEDBACK-PARTICIPATION-001
REQ-ANON-004 for their own warn-and-continue disposition, which is context,
not authority for this question)
Why unresolved: the card directs adding the check but does not state its
failure semantics; warn-and-continue changes what a shipped command prints
(fail would additionally change its exit semantics).
Class: implementation-level
Default: warn and continue, exit code unchanged (rule: preserves current
behavior — nothing on the skip path fails the update today, so warn keeps
the update's exit semantics and the undo is a single revert of this
SPEC's commits)
Alternate: fail the update (non-zero exit) when damage is found
Operator verdict: DEFAULT-APPLIED 2026-10-09T03:49:48Z manager-spec (plan close, card t1578 lane)

### Q2: Which paths does the probe's fixed representative set cover, and does it include the user-folder install roots?

Label: FOUNDER
Authority anchor: (none — no completed SPEC decides the composition of a
managed-surface integrity probe; the probe itself is new capability from
card item 4)
Why unresolved: the card says "representative file existence probe" without
naming the set; the set is revisable and its composition decides both the
probe's value and its maintenance cost.
Class: implementation-level
Default: minimal project-surface set — `.claude/settings.json` (exists and
parses as JSON), `.moai/config/sections/system.yaml` (exists, non-empty),
`.moai/manifest.json` (exists and parses) (rule: smaller user-visible
surface — the fewest new warning lines that still detect the mo.ai.kr
damage class; revisit trigger: any future managed-path loss that this set
misses)
Alternate: extend the set to spot-check user-folder install roots (one
representative installed file per root)
Operator verdict: DEFAULT-APPLIED 2026-10-09T03:49:48Z manager-spec (plan close, card t1578 lane)

### Q3: Should the project migrate user values out of settings.json into settings.local.json (the pure-template model), and does this SPEC carry that migration?

Label: EVIDENCE-NEEDED
Authority anchor: (interim disposition recorded in spec.md Section C.1;
the migration DECISION waits on data that does not exist yet)
Why unresolved: deciding whether a local-file migration is warranted needs
a measurement that has not been taken — how many deployed projects carry
user-authored keys mixed into settings.json, and which keys. The current
tree's 3-way merge model already preserves user values (measured: 32 keys
preserved on the mo.ai.kr force update; merge behavior pinned by
TestRunUpdate_V3Path_CleanSettingsUntouched), so the interim disposition
is no-migration, out of scope for this SPEC, with a follow-up card
candidate ("Settings purity survey") named in spec.md C.1. The operator
delegated the item-(3) disposition decision to the plan phase (card t1578
approval); the eventual migration funding decision stays with the
operator.
Operator verdict:
