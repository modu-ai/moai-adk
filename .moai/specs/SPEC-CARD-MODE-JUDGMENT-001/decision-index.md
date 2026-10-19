# decision-index.md — SPEC-CARD-MODE-JUDGMENT-001

Stateless on the status axis (no `status:` field; the lifecycle lives in `spec.md`). Six decisions the plan phase surfaced and the operator has not settled. Each row states what is unresolved and why; none carries a recommendation. Where the published Default rule ranks the options (the option that preserves current behavior; else the one whose undo is a single revert of this SPEC's own commits; else the smaller user-visible surface), the row carries `Default:` and `Alternate:`, and an `implementation-level` row with a `Default:` is stamped `DEFAULT-APPLIED` at plan close. Every other FOUNDER row is left empty and goes to the operator.

### Q1: Do `manager-todo.md` and `manager-lead.md` get an issuance pointer?

Label: FOUNDER
Class: implementation-level
Authority anchor: (none — the card text names "the leader, manager-todo, and the todo workflow skill"; no committed clause decides it)
Why unresolved: Measured on tree `74ad197380a0711cdb9816d4e37c94d29af730ea`: `manager-todo.md` has no issuance step (it says the `--auto` cycle "never self-promotes, admits, drops, or edits cards") and `manager-lead.md:235` only lists `moai gtd` add among queue mutations a deputy may not perform. The leader's issuance clause is `factory-dispatch.md:30` and the todo skill's is `gtd.md:52`. Adding a pointer to either agent makes both TOMLs under `.codex/agents/moai/` regenerate (`make agents-emit`) and changes their catalog hashes.
Default: leave both agent files untouched (rule: preserves current behavior)
Alternate: add one pointer sentence to each agent to `factory-dispatch-cards.md` § Mode judgment, then run `make agents-emit` and the emit check
Operator verdict: DEFAULT-APPLIED 2026-10-19T02:28:48Z lane-31 manager-spec (plan author)

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
Operator verdict:

### Q4: Is the Go SessionStart lane notice in scope?

Label: FOUNDER
Class: implementation-level
Authority anchor: (none — card t1894 says "Guidance change" and "Code (the default and the reclassify verb) is t1737's")
Why unresolved: `internal/hook/session_start_factory.go` and `session_start_factory_i18n.go` (`laneRecheckRule`, four locales) inject the unscoped rule "arm ONE recurring `CronCreate` recheck before a card's first stage … `CronDelete` it when you send the completion report", pinned by `internal/hook/lane_renew_interval_test.go`. The scoped documents and that injected notice will disagree until the notice changes. Changing it is Go code with i18n renderings and tests, which this documents-only card excludes.
Default: out of scope; the disagreement is recorded as residual risk R-1 and a follow-up card is proposed (rule: preserves current behavior)
Alternate: include the Go change (hook source, i18n table, the pinning test) in this card
Operator verdict: DEFAULT-APPLIED 2026-10-19T02:28:48Z lane-31 manager-spec (plan author)

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
