# decision-index.md — SPEC-TODO-UNIFY-001

`interview.decision_gate: on`. Nine decisions surfaced while assembling the plan for card t1641 that the operator has not settled. This file carries no status field; the SPEC's lifecycle lives in `spec.md` alone. `Operator verdict` is empty at authoring and is read by the Kickoff gate. Label vocabulary: DECIDED, POLICY-COVERED, EVIDENCE-NEEDED, FOUNDER. Every row states what is unresolved and why; none carries a preferred answer for a judgment call. A `Default:` line is the option the published default rule ranks first (preserves current behavior; else the option whose undo is a single revert of this SPEC's own commits; else the smaller user-visible surface) and is a policy application, not a recommendation.

No row can be labelled DECIDED or POLICY-COVERED: the only places that decide these questions are the live queue card (not in the committed tree) and the untracked 3.2 plan and redesign reports under `.moai/reports/`, which are never authority. Where committed material bears on a row it is named under `Authority anchor` as related context, not as authority.

At plan close the implementation-level rows that carry a `Default:` and an empty verdict are filled with `DEFAULT-APPLIED`. Product-level rows stay empty and block the autonomous Kickoff; they go to the operator.

## Q1: What does "clean up the 6 gtd subverbs" mean — absorb all six into `moai todo`, absorb the used ones and delete the rest, or leave them under `moai gtd`?

- Label: FOUNDER
- Class: product-level
- Authority anchor: none. Related committed context: `AGENTS.md` §11 (`moai todo` canonical, `moai gtd` a fully supported alias); `.claude/skills/moai/workflows/gtd.md` lines 150, 161, 173, 185, 194, 211 (the guidance names the six as `moai todo <verb>`).
- Why unresolved: the card and its two absorbed cards disagree. Card t1590 fixes "register the six on `moai todo`". Card t1602 says the GTD-stage verbs "stay under `moai gtd`". The redesign's D3 sentence says absorb the used ones and delete the rest. Reading (c) cannot coexist with "gtd warns, then delegates" and removal after one release for those six verbs; reading (b) removes a user-facing feature. Measured use of the CLI verbs in the tree is nil (library functions are called directly by `goal.go`); the verbs are documented in the workflow skill, four READMEs, and four docs-site locales.
- Default: (a) absorb all six (rule: preserves current behavior)
- Alternate: (b) absorb the verbs with measured use, delete the rest
- Operator verdict:

## Q2: Does the warning name a literal removal release or say "the next release"?

- Label: FOUNDER
- Class: implementation-level
- Authority anchor: none.
- Why unresolved: the card says "one release of grace" without naming a number; the release plan's number is not in a committed artifact the plan can cite. Both options satisfy the tests (the warning contains a non-empty constant); they differ in whether a release number must be kept current.
- Default: (b) the phrase "the next release" (rule: smaller user-visible surface)
- Alternate: (a) a literal release taken from the release plan
- Operator verdict: DEFAULT-APPLIED 2026-10-18T06:33:24Z manager-spec (plan owner, card t1641)

## Q3: Is "rename gtd to todo everywhere" the user-facing spelling only, or also the internals (Go identifiers, `gtd_*` tables, the `gtd-<16 hex>` item id, `.moai/gtd/answers`)?

- Label: FOUNDER
- Class: product-level
- Authority anchor: none. Related committed context: the doc comment on `NewGTDCommand` (`internal/cli/gtd.go`) states that internal package names and the on-disk location "deliberately remain unchanged".
- Why unresolved: the operator's order reads "rename gtd to todo everywhere". Renaming persisted formats needs a data migration and a dual-read of the old item id and touches about 47 files; renaming only the spelling leaves identifiers carrying the old name. Neither reading is stated in a committed artifact.
- Default: (a) user-facing spelling only (rule: preserves current behavior)
- Alternate: (b) everything, as a separate SPEC with its own migration
- Operator verdict:

## Q4: Is the landing-to-done machine transition with three observers, and automatic receipt issuing, in this card's scope or card t1639's?

- Label: FOUNDER
- Class: product-level
- Authority anchor: none. Related committed context: `SPEC-FACTORY-RECORD-REPAIR-001` spec.md §G (moved items to "3.2-2-2, t1639").
- Why unresolved: the master plan lists the machine transition and automatic receipts under t1639 (on hold) and only "lane own-card done" and "T18/T19 removal" under t1641, while t1641's completion criterion reads "done is automatic for the most recent 10 landed cards". Delivering the observers here duplicates t1639 and adds a SessionEnd hook that must satisfy the advisory-check discipline.
- Default: (a) deliver the lane own-card path and the receipt seam only (rule: smaller user-visible surface)
- Alternate: (b) also deliver the SessionEnd and `factory complete` observers
- Operator verdict:

## Q5: What satisfies the leader-approval receipt gate when a lane closes its own card?

- Label: FOUNDER
- Class: product-level
- Authority anchor: none. Related committed context: `internal/cli/leader_approval_gate.go` header (issuance is leader-only; lane sessions are refused at the boundary); draft `SPEC-FACTORY-COMPLETION-RECOVERY-001` REQ-FCR-002.
- Why unresolved: the gate binds every close of a factory-linked card and only the leader can mint a receipt. The options differ in who is trusted: landing evidence alone (a), a receipt issued automatically from the verdict file, landing evidence, and the remeasure hash (b), or a receipt the leader already issued (c). Under (c) the automatic-done criterion cannot hold without a leader act.
- Default: (c) the lane's `done` is admitted but still needs a leader-issued receipt (rule: preserves current behavior)
- Alternate: (a) landing evidence substitutes for the receipt for that card only
- Operator verdict:

## Q6: Are T18 and the T19 reservation removed now, given the GitHub Flow transition is not complete in this repository?

- Label: FOUNDER
- Class: product-level
- Authority anchor: none. Related committed context: `.moai/config/sections/git-strategy.yaml` (`mode: manual`, `manual.workflow: git-flow`); `SPEC-GITHUB-FLOW-DEFAULT-001` `status: in-progress`; draft `SPEC-FACTORY-COMPLETION-RECOVERY-001` REQ-FCR-010 (would implement T19); `internal/homestate/card_evidence.go` `PushGateTarget` (returns `done` for a repository with no remote).
- Why unresolved: the decision text removes the git-flow edges "with the completion of the GitHub Flow transition", which has not happened here. T18 is today the only completion edge for a card whose repository has no remote, so removing it removes a user-facing completion path unless a replacement is named. Two draft SPECs edit the same rows in opposite directions.
- Default: (c) defer the removal until `SPEC-GITHUB-FLOW-DEFAULT-001` closes (rule: preserves current behavior)
- Alternate: (b) remove only the T19 reservation and keep T18 guarded for no-remote repositories
- Operator verdict:

## Q7: Which record carries "done" in the ten-card criterion, and what counts as "landed"?

- Label: FOUNDER
- Class: product-level
- Authority anchor: none. Related committed context: `SPEC-TODO-LAND-AUTO-DONE-001` (the auto-done scan reads the resolved landed ref and closes the queue row); `SPEC-LOCAL-MAIN-FLOW-001` (in the local-main flow the merge is on local `main` while `origin/main` advances by batch).
- Why unresolved: the queue (archive) and the factory card row each have a `done`, closed by different paths today. "Landed" differs between the resolved landed ref and local `main`. A card done in one record and open in the other is an observed wedge.
- Default: record: queue only; landed: reachable from the resolved landed ref (rule: preserves current behavior)
- Alternate: record: both queue and factory row
- Operator verdict:

## Q8: What shape does the post-pick spec_id setter take?

- Label: FOUNDER
- Class: implementation-level
- Authority anchor: none. Related committed context: `internal/homestate/card_picked.go` (the refusal "its hints and pointers change only while picked"); `internal/cli/contract_report.go` `cardToSpecID` (reads the queue spec_id).
- Why unresolved: three shapes exist. A new verb that writes both carriers set-once (a); widening `factory assign --spec` to set-if-empty after assignment, which fixes the factory row only (b); letting `next <n> --spec` re-enter a picked card, which removes the positive pick gate (c).
- Default: (a) a new `moai todo spec <n> <SPEC-ID>` verb writing both carriers set-once (rule: preserves current behavior; every existing verb is unchanged)
- Alternate: (b) widen `factory assign --spec`
- Operator verdict: DEFAULT-APPLIED 2026-10-18T06:33:24Z manager-spec (plan owner, card t1641)

## Q9: Who may call the spec_id setter — the leader only, or a lane for its own card?

- Label: FOUNDER
- Class: implementation-level
- Authority anchor: none. Related committed context: `todoLaneReadOnlyVerbs` in `internal/cli/todo.go` (a lane may run seven read-only verbs; every other verb is refused).
- Why unresolved: a lane learns the SPEC id at plan phase, so a leader-only setter forces a message to the leader; admitting a lane for its own card widens the lane allowlist a second time (the first is M2's own-card done).
- Default: (a) leader only (rule: preserves current behavior)
- Alternate: (b) a lane, for its own card, only while the spec_id is empty
- Operator verdict: DEFAULT-APPLIED 2026-10-18T06:33:24Z manager-spec (plan owner, card t1641)
