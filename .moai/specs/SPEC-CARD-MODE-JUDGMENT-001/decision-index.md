# decision-index.md — SPEC-CARD-MODE-JUDGMENT-001

Stateless on the status axis (no `status:` field; the lifecycle lives in `spec.md`). Six decisions the plan phase surfaced. Q3 has since been settled by an operator ruling (recorded in its row); the others carry the stamps shown. Each row states what was unresolved and why; none carries a recommendation. Where the published Default rule ranks the options (the option that preserves current behavior; else the one whose undo is a single revert of this SPEC's own commits; else the smaller user-visible surface), the row carries `Default:` and `Alternate:`, and an `implementation-level` row with a `Default:` is stamped `DEFAULT-APPLIED` at plan close. Every other FOUNDER row is left empty and goes to the operator.

### Q1: Do `manager-todo.md` and `manager-lead.md` get an issuance pointer?

Label: FOUNDER
Class: implementation-level
Authority anchor: (none — the card text names "the leader, manager-todo, and the todo workflow skill"; no committed clause decides it)
Why unresolved: Measured on tree `74ad197380a0711cdb9816d4e37c94d29af730ea`: `manager-todo.md` has no issuance step (it says the `--auto` cycle "never self-promotes, admits, drops, or edits cards") and `manager-lead.md:235` only lists `moai gtd` add among queue mutations a deputy may not perform. The leader's issuance clause is `factory-dispatch.md:30` and the todo skill's is `gtd.md:52`. Adding a pointer to either agent makes both TOMLs under `.codex/agents/moai/` regenerate (`make agents-emit`) and changes their catalog hashes.
Default: leave both agent files untouched (rule: preserves current behavior)
Alternate: add one pointer sentence to each agent to `factory-dispatch-cards.md` § Mode judgment, then run `make agents-emit` and the emit check
Operator verdict: DEFAULT-APPLIED 2026-10-19T02:28:48Z lane-31 manager-spec (plan author)
Record note (correction round): the operator ruling d-20261019T023441Z-d314 (read under Q3 below) ends with "Q1 default (manager-todo untouched) and Q4 (laneRecheckRule moved to t1695) stand." The Q1 default therefore stands as stamped; no agent file is edited.

### Q2: Is the standing-source card of `doc-generation.md` in scope?

Label: FOUNDER
Class: implementation-level
Authority anchor: (none — board 0c51's predicate is "every card the leader issues or picks"; a standing-source `[PROJECT]` card is issued by the project workflow, not the leader)
Why unresolved: `internal/template/templates/.claude/skills/moai/workflows/project/doc-generation.md:348` issues `moai gtd add "[PROJECT] <goal, one line, bounded by scope>"` with no classification, so it is the one remaining bare issuance outside sweep set R. A `[PROJECT]` card is a goal, not a file set, so the mode predicate of 0c51 does not obviously apply; which mode such a card should declare is not stated anywhere.
Default: exclude it from this SPEC (rule: preserves current behavior)
Alternate: include it, which also requires an operator decision on the mode a standing-source card declares (a product-level question)
Operator verdict: DEFAULT-APPLIED 2026-10-19T02:28:48Z lane-31 manager-spec (plan author)

### Q3: Does "assignment input = picked + queued parallelizable" change who promotes a card?

Label: FOUNDER
Class: product-level
Authority anchor: (none for the new input; the existing promotion clause is `factory-dispatch.md` § Entry into the queue is an operator act — "Promotion is the operator's act, live or in advance" — and the lane lease exception)
Why unresolved: The card defines "the assignment input as picked + queued parallelizable". The board records 0c51 and a672 do not mention it. Read as a description of the candidate set a free lane may be handed, nothing about promotion changes. Read as letting the leader hand a queued parallelizable card to a free lane without an operator pick, it contradicts the promotion clause. The two readings change a shipped coordination behavior differently, so the wording of REQ-MJ-007 and AC-MJ-006 cannot be settled by the planner.
Default: the sentence names the candidate set and leaves promotion exactly as the existing clauses state it (rule: preserves current behavior)
Alternate: the leader may promote queued `parallelizable` cards to free lanes (requires amending the promotion clause and its ledger row)
Operator verdict: met 2026-10-19T02:34:41Z — decided_by=operator (AskUserQuestion 10-19), board record `d-20261019T023441Z-d314`, scope=card:t1894, kind=ruling, evidence_refs=`lane-31 plan 55d4f1486a decision-index Q3; factory-dispatch.md:38`. The time is the one the record id carries. Read in this correction round with `moai decision read --scope card:t1894 --project-root /Users/goos/MoAI/moai-adk-go` (exit 0, first line `board=ok unparseable=0 records=141`) on tree 55d4f1486a454ff446be6f0f0645d4cea76bf0cc, judged by the installed build `moai_cp/20261007_195927`. The record, verbatim as printed:

```text
decision record: decided_by=operator (AskUserQuestion 10-19) evidence_refs=lane-31 plan 55d4f1486a decision-index Q3; factory-dispatch.md:38 ladder_path=operator verdict Q3 (FOUNDER row) id=d-20261019T023441Z-d314 scope=card:t1894 kind=ruling
  body: t1894 Q3 = leader auto-promotion is stated explicitly: while the operator has delegated dispatch (operator-away / standing delegation), when a live lane is free (no actively executing card, per a672) the leader itself promotes queued cards classified parallelizable (explicit judgment per 0c51) to picked and assigns them, without a per-card operator pick. Ordering: queue priority/readiness first; keep-set, held and [보류-marked cards are never auto-promoted; serial cards follow the slot rules. Lane self-dispatch is unchanged. Q1 default (manager-todo untouched) and Q4 (laneRecheckRule moved to t1695) stand.
```

Row status: met. The verdict is neither the row's Default nor its Alternate as worded: it permits leader promotion, under the limits above, while the operator has delegated dispatch. `spec.md` REQ-MJ-007 and REQ-MJ-017 implement it and nothing wider.

### Q4: Is the Go SessionStart lane notice in scope?

Label: FOUNDER
Class: implementation-level
Authority anchor: (none — card t1894 says "Guidance change" and "Code (the default and the reclassify verb) is t1737's")
Why unresolved: `internal/hook/session_start_factory.go` and `session_start_factory_i18n.go` (`laneRecheckRule`, four locales) inject the unscoped rule "arm ONE recurring `CronCreate` recheck before a card's first stage … `CronDelete` it when you send the completion report", pinned by `internal/hook/lane_renew_interval_test.go`. The scoped documents and that injected notice will disagree until the notice changes. Changing it is Go code with i18n renderings and tests, which this documents-only card excludes.
Default: out of scope; the disagreement is recorded as residual risk R-1 and a follow-up card is proposed (rule: preserves current behavior)
Alternate: include the Go change (hook source, i18n table, the pinning test) in this card
Operator verdict: DEFAULT-APPLIED 2026-10-19T02:28:48Z lane-31 manager-spec (plan author)
Record note (correction round): the operator ruling d-20261019T023441Z-d314 (read under Q3 below) states "Q4 (laneRecheckRule moved to t1695)". The Go SessionStart lane notice is therefore owned by card t1695 and stays out of this SPEC's scope; the disagreement it leaves is residual risk R-1 in `spec.md` §7.

### Q5: Which cards count as "in flight", and which paths make a surface "slot-gated"?

Label: FOUNDER
Class: implementation-level
Authority anchor: (none — board 0c51 gives the categories "factory/todo core, release pipeline, shared migrations" and the phrase "in-flight card" without membership rules or paths)
Why unresolved: Whether a merge-ready or held card still counts as in flight for the file-overlap test (its files are unlanded) is not stated; neither is the path set behind each of the three categories. Both affect how consistently different issuers judge, not what the documents must contain.
Default: keep the board's words verbatim in the documents and leave membership and paths to the issuer's recorded reason (rule: preserves current behavior)
Alternate: define the membership and the path globs in the documents, which commits the documents to a list that code and layout changes will drift from
Operator verdict: DEFAULT-APPLIED 2026-10-19T02:28:48Z lane-31 manager-spec (plan author)

### Q6: Which Tier is this SPEC?

Label: POLICY-COVERED
Authority anchor: `.claude/rules/moai/workflow/spec-workflow.md` § SPEC Complexity Tier (S/M/L) — the "Files affected" column: Tier L is "> 15 files"
Why unresolved: The table is guidance on file count, and whether deployed mirrors and generated files count as affected files is the plan author's reading. Counted without discounting (8 template documents + 8 deployed mirrors + `catalog.yaml` + `binding_ledger.json` = 18) the table gives Tier L; counting only the eight source documents gives Tier M. The operator Tier question of `spec-assembly.md` could not be asked from a subagent.
Operator verdict:

## Gate (state after the correction round)

| Row | Label / class | State |
|---|---|---|
| Q1 | FOUNDER / implementation-level | `DEFAULT-APPLIED`; the operator ruling d-20261019T023441Z-d314 confirms the default |
| Q2 | FOUNDER / implementation-level | `DEFAULT-APPLIED` (unchanged) |
| Q3 | FOUNDER / product-level | **met** by the operator ruling d-20261019T023441Z-d314 (verdict recorded above) |
| Q4 | FOUNDER / implementation-level | `DEFAULT-APPLIED`; the same ruling moves the notice to card t1695 |
| Q5 | FOUNDER / implementation-level | `DEFAULT-APPLIED` (unchanged) |
| Q6 | POLICY-COVERED | anchor present (unchanged) |

No FOUNDER row carries an empty verdict, so no row blocks the autonomous Kickoff and no requirement or criterion is gated on a decision.

## Milestone-to-row map

| Milestone (`plan.md`) | Rows it depends on | Note |
|---|---|---|
| M1 mode judgment and issuance | Q1 (agents untouched), Q2 (`doc-generation.md` untouched), Q5 (board words verbatim) | all defaults |
| M2 lease renewal and scoped cron | Q4 (Go notice is card t1695's; residual risk R-1) | nothing to decide here |
| M3 free lane and delegated-dispatch promotion | Q3 (met) | REQ-MJ-006, REQ-MJ-007, REQ-MJ-017 are written against the ruling, not against a planner default |
| M4 mechanical sync | Q1 (if flipped: agent emission) | defaults hold |
| M5 verification | none | |
