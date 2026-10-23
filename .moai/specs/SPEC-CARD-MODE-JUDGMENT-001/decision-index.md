# decision-index.md — SPEC-CARD-MODE-JUDGMENT-001

Stateless on the status axis (no `status:` field; the lifecycle lives in `spec.md`). Seven decisions the plan phase surfaced. Q3 has been settled by an operator ruling (recorded in its row); Q7 was answered by the operator on 2026-10-23, relayed by the leader (recorded in its row); the others carry the stamps shown. Each row states what was unresolved and why; none carries a recommendation for a judgment call. Where the published Default rule ranks the options (the option that preserves current behavior; else the one whose undo is a single revert of this SPEC's own commits; else the smaller user-visible surface), the row carries `Default:` and `Alternate:`, and an `implementation-level` row with a `Default:` is stamped `DEFAULT-APPLIED` at plan close. Every other FOUNDER row is left empty and goes to the operator.

### Q1: Do `manager-todo.md` and `manager-lead.md` get an issuance pointer?

Label: FOUNDER
Class: implementation-level
Authority anchor: (none — the card text names "the leader, manager-todo, and the todo workflow skill"; no committed clause decides it)
Why unresolved: Measured on tree `74ad197380a0711cdb9816d4e37c94d29af730ea`: `manager-todo.md` has no issuance step (it says the `--auto` cycle "never self-promotes, admits, drops, or edits cards") and `manager-lead.md:235` only lists `moai gtd` add among queue mutations a deputy may not perform. The leader's issuance clause is `factory-dispatch.md:30` and the todo skill's is `gtd.md:52`. Adding a pointer to either agent makes both TOMLs under `.codex/agents/moai/` regenerate (`make agents-emit`) and changes their catalog hashes.
Default: leave both agent files untouched (rule: preserves current behavior)
Alternate: add one pointer sentence to each agent to `factory-dispatch-cards.md` § Mode judgment, then run `make agents-emit` and the emit check
Operator verdict: DEFAULT-APPLIED 2026-10-19T02:28:48Z lane-31 manager-spec (plan author)
Record note: the operator ruling d-20261019T023441Z-d314 (read under Q3 below) ends with "Q1 default (manager-todo untouched) and Q4 (laneRecheckRule moved to t1695) stand." The Q1 default therefore stands as stamped; no agent file is edited.

### Q2: Is the standing-source card of `doc-generation.md` in scope?

Label: FOUNDER
Class: implementation-level
Authority anchor: (none — board 0c51's predicate is "every card the leader issues or picks"; a standing-source `[PROJECT]` card is issued by the project workflow, not the leader)
Why unresolved: `internal/template/templates/.claude/skills/moai/workflows/project/doc-generation.md:348` issues `moai gtd add "[PROJECT] <goal, one line, bounded by scope>"` with no classification, so it is the one remaining bare issuance outside sweep set R. A `[PROJECT]` card is a goal, not a file set, so the mode predicate of 0c51 does not obviously apply; which mode such a card should declare is not stated anywhere.
Default: exclude it from this SPEC (rule: preserves current behavior)
Alternate: include it, which also requires an operator decision on the mode a standing-source card declares
Operator verdict: DEFAULT-APPLIED 2026-10-19T02:28:48Z lane-31 manager-spec (plan author)
Class note (correction round 1): this row's question is scope — whether this SPEC edits that one line — and that is implementation-level. The follow-on question, which mode a standing-source card declares, is a separate product-level question that this row does not decide and the Alternate would raise.
Consequence of the default (stated, not hidden): after this SPEC lands, `doc-generation.md:348` stays a bare `moai gtd add` with no `--classification-file`. If a standing-source card counts as issued by the leader, 0c51 ("No card is issued on the fail-safe serial default") forbids exactly that line, so the SPEC leaves one known 0c51 site open. It is recorded as residual risk R-7 in `spec.md`; a follow-up card for it is proposed in the correction-round report, and the operator decides whether such a card is in scope and which mode it declares.

### Q3: Does "assignment input = picked + queued parallelizable" change who promotes a card?

Label: FOUNDER
Class: product-level
Authority anchor: (none for the new input; the existing promotion clause is `factory-dispatch.md` § Entry into the queue is an operator act — "Promotion is the operator's act, live or in advance" — and the lane lease exception)
Why unresolved: The card defines "the assignment input as picked + queued parallelizable". The board records 0c51 and a672 do not mention it. Read as a description of the candidate set a free lane may be handed, nothing about promotion changes. Read as letting the leader hand a queued parallelizable card to a free lane without an operator pick, it contradicts the promotion clause. The two readings change a shipped coordination behavior differently, so the wording of REQ-MJ-007 and AC-MJ-006 cannot be settled by the planner.
Default: the sentence names the candidate set and leaves promotion exactly as the existing clauses state it (rule: preserves current behavior)
Alternate: the leader may promote queued `parallelizable` cards to free lanes (requires amending the promotion clause and its ledger row)
Operator verdict: met 2026-10-19T02:34:41Z — decided_by=operator (AskUserQuestion 10-19), board record `d-20261019T023441Z-d314`, scope=card:t1894, kind=ruling, evidence_refs=`lane-31 plan 55d4f1486a decision-index Q3; factory-dispatch.md:38`. The time is the one the record id carries. Read with `moai decision read --scope card:t1894 --project-root /Users/goos/MoAI/moai-adk-go` (exit 0, first line `board=ok unparseable=0 records=141`) on tree 55d4f1486a454ff446be6f0f0645d4cea76bf0cc, judged by the installed build `moai_cp/20261007_195927`; re-read in correction round 1 from the full board (`moai decision read`, saved to a scratch file, grepped by id). The record, verbatim as printed:

```text
decision record: decided_by=operator (AskUserQuestion 10-19) evidence_refs=lane-31 plan 55d4f1486a decision-index Q3; factory-dispatch.md:38 ladder_path=operator verdict Q3 (FOUNDER row) id=d-20261019T023441Z-d314 scope=card:t1894 kind=ruling
  body: t1894 Q3 = leader auto-promotion is stated explicitly: while the operator has delegated dispatch (operator-away / standing delegation), when a live lane is free (no actively executing card, per a672) the leader itself promotes queued cards classified parallelizable (explicit judgment per 0c51) to picked and assigns them, without a per-card operator pick. Ordering: queue priority/readiness first; keep-set, held and [보류-marked cards are never auto-promoted; serial cards follow the slot rules. Lane self-dispatch is unchanged. Q1 default (manager-todo untouched) and Q4 (laneRecheckRule moved to t1695) stand.
```

Row status: met. The verdict is neither the row's Default nor its Alternate as worded: it permits leader promotion, under the limits above, while the operator has delegated dispatch. `spec.md` REQ-MJ-007 and REQ-MJ-017 implement it and nothing wider; the ruling's seven limits are the list in REQ-MJ-007, and the one sentence that is a consequence of them (the assignment input) is labelled as a consequence there.

### Q4: Is the Go SessionStart lane notice in scope?

Label: FOUNDER
Class: implementation-level
Authority anchor: (none — card t1894 says "Guidance change" and "Code (the default and the reclassify verb) is t1737's")
Why unresolved: `internal/hook/session_start_factory.go` and `session_start_factory_i18n.go` (`laneRecheckRule`, four locales) inject the unscoped rule "arm ONE recurring `CronCreate` recheck before a card's first stage … `CronDelete` it when you send the completion report", pinned by `internal/hook/lane_renew_interval_test.go`. The scoped documents and that injected notice will disagree until the notice changes. Changing it is Go code with i18n renderings and tests, which this documents-only card excludes.
Default: out of scope; the disagreement is recorded as residual risk R-1 and a follow-up card is proposed (rule: preserves current behavior)
Alternate: include the Go change (hook source, i18n table, the pinning test) in this card
Operator verdict: DEFAULT-APPLIED 2026-10-19T02:28:48Z lane-31 manager-spec (plan author)
Record note: the operator ruling d-20261019T023441Z-d314 (read under Q3 above) states "Q4 (laneRecheckRule moved to t1695)". The Go SessionStart lane notice is therefore owned by card t1695 and stays out of this SPEC's scope; the disagreement it leaves is residual risk R-1 in `spec.md` §7.

### Q5: Which cards count as "in flight", and which paths make a surface "slot-gated"?

Label: FOUNDER
Class: implementation-level
Authority anchor: (none — board 0c51 gives the categories "factory/todo core, release pipeline, shared migrations" and the phrase "in-flight card" without membership rules or paths)
Why unresolved: Whether a merge-ready or held card still counts as in flight for the file-overlap test (its files are unlanded) is not stated; neither is the path set behind each of the three categories. Both affect how consistently different issuers judge, not what the documents must contain.
Default: keep the board's words verbatim in the documents and leave membership and paths to the issuer's recorded reason (rule: preserves current behavior)
Alternate: define the membership and the path globs in the documents, which commits the documents to a list that code and layout changes will drift from
Operator verdict: DEFAULT-APPLIED 2026-10-19T02:28:48Z lane-31 manager-spec (plan author)
Consequence of the default (stated, not hidden): with "in flight" undefined, two issuers can judge the same merge-ready or held card differently, and at the time of the plan 25 merge-ready cards sat in the queue (board a672 evidence), so whether a merge-ready card counts as in flight decides most outcomes of the new judgment. The documents mitigate it only by requiring the reason to state the file set and the conflict check, so the judgment is auditable and disputable; they do not make it uniform. This is recorded as residual risk R-2 in `spec.md`. Defining "in flight" (for example as "any card not yet landed") is an operator call this row leaves open; the Alternate is the route for it.

### Q6: Which Tier is this SPEC?

Label: FOUNDER
Class: implementation-level
Authority anchor: (none — `.claude/rules/moai/workflow/spec-workflow.md` § SPEC Complexity Tier is a rule-table guidance whose own text says the Tier is judged by a Socratic question; it is not an operator setting in `.moai/config/sections/*.yaml` or a constitution clause, so it does not cover the question as written and this row does not claim policy coverage)
Why unresolved: The table is guidance on file count, and whether deployed mirrors and generated files count as affected files is the plan author's reading. Counted without discounting (8 template documents + 8 deployed mirrors + `catalog.yaml` + `binding_ledger.json` = 18) the "Files affected" column gives Tier L; counting only the eight source documents gives Tier M. The operator Tier question of `spec-assembly.md` could not be asked from a subagent. A fact that bounds the options, read from the same table: the SPEC carries 17 requirements and 20 acceptance criteria, which exceed the Tier M ceilings of 16 and 16, so Tier M is reachable only by cutting content.
Default: Tier L (rule: preserves current behavior — the SPEC is already authored, measured and audited as Tier L)
Alternate: Tier M, after cutting the requirements and the criteria to the Tier M ceilings (3 artifacts)
Operator verdict: DEFAULT-APPLIED 2026-10-19T03:35:14Z lane-31 manager-spec (plan author)

### Q7: What establishes "the operator has delegated dispatch"?

Label: FOUNDER
Class: product-level
Authority anchor: (none — the ruling d-20261019T023441Z-d314 conditions the leader's promotion on "while the operator has delegated dispatch (operator-away / standing delegation)" and names no mechanism, record, or expiry; no committed clause decides it)
Why unresolved: The documents carry the ruling's words (`(operator away, or a standing delegation)`) and, by design, no mechanism. An actor that treats an unstated delegation as given widens the exception in practice, and no command in this SPEC can see that (spec R-5). Which signal counts — nothing beyond the operator's own words in the session, a decision-board record the leader cites, a configuration key, or a time-bounded delegation — changes who may dispatch and when, and so changes a shipped coordination behavior. The options visible to the planner are listed unranked; the published Default rule does not rank them, because each one changes the user-visible behavior differently.
Operator verdict: met 2026-10-23 — decided_by=operator, relayed by the leader (cross-session message of 2026-10-23; this lane did not receive the answer from the operator directly). Answer: "the verdict's phrase is the whole condition" (Korean original: "판정문 구절이 전체 조건") — the phrase `(operator away, or a standing delegation)` of the ruling d-20261019T023441Z-d314 is the whole condition, and no further mechanism, record format or expiry is required. evidence_refs=the leader's cross-session message of 2026-10-23; `.moai/reports/t1894/lane8-wait.md` § "Kickoff (2026-10-23, lane-8)". No decision-board record is cited: `moai decision read --scope card:t1894 --project-root /Users/goos/MoAI/moai-adk-go` (exit 0, first line `board=ok unparseable=0 records=146`) printed no record id dated 2026-10-23.
Gate note: No requirement, criterion or milestone of this SPEC waits on Q7; REQ-MJ-007 (a) is text-final because it carries the ruling's words and no mechanism. Per the decision-gate rule the empty FOUNDER row Q7 goes to the operator and blocks the autonomous form of the plan→run Kickoff until the operator answers it or states that the ruling's phrase is the whole condition. State on 2026-10-23: the operator stated exactly that (relayed by the leader, verdict above), so the row is no longer empty and Q7 no longer blocks the autonomous form of the Kickoff.

Row status: met. The answer adds nothing to the documents: REQ-MJ-007 (a) already carries the ruling's words and no mechanism, and residual risk R-5 (the practice of treating an unstated delegation as given) stands as stated.

## Gate (state on 2026-10-23)

| Row | Label / class | State |
|---|---|---|
| Q1 | FOUNDER / implementation-level | `DEFAULT-APPLIED`; the operator ruling d-20261019T023441Z-d314 confirms the default |
| Q2 | FOUNDER / implementation-level | `DEFAULT-APPLIED` (unchanged); consequence now stated (R-7) |
| Q3 | FOUNDER / product-level | **met** by the operator ruling d-20261019T023441Z-d314 (verdict recorded above) |
| Q4 | FOUNDER / implementation-level | `DEFAULT-APPLIED`; the same ruling moves the notice to card t1695 |
| Q5 | FOUNDER / implementation-level | `DEFAULT-APPLIED` (unchanged); consequence now stated (R-2) |
| Q6 | FOUNDER / implementation-level | `DEFAULT-APPLIED` (re-labelled from POLICY-COVERED, which the cited anchor did not support) |
| Q7 | FOUNDER / product-level | **met** — answered 2026-10-23 (relayed by the leader): the ruling's phrase is the whole condition; no longer blocks the autonomous form of the Kickoff; no requirement, criterion or milestone waits on it |

No FOUNDER row carries an empty verdict. No requirement, criterion or milestone of this SPEC waits on Q7; REQ-MJ-007 (a) is text-final because it carries the ruling's words and no mechanism. Per the decision-gate rule the empty FOUNDER row Q7 goes to the operator and blocks the autonomous form of the plan→run Kickoff until the operator answers it or states that the ruling's phrase is the whole condition. Q7 was answered on 2026-10-23 (relayed by the leader) with exactly that statement, so it no longer blocks the autonomous Kickoff form.

## Milestone-to-row map

| Milestone (`plan.md`) | Rows it depends on | Note |
|---|---|---|
| M1 mode judgment and issuance | Q1 (agents untouched), Q2 (`doc-generation.md` untouched), Q5 (board words verbatim) | all defaults |
| M2 lease renewal and scoped standing cron | Q4 (Go notice is card t1695's; residual risk R-1) | nothing to decide here |
| M3 free lane and delegated-dispatch promotion | Q3 (met), Q7 (answered 2026-10-23, relayed by the leader; no text waits on it, and it no longer blocks the autonomous Kickoff form) | REQ-MJ-006, REQ-MJ-007, REQ-MJ-017 are written against the ruling, not against a planner default |
| M4 confirmation pass | Q1 (if flipped: agent emission) | defaults hold |
| M5 verification | none | |
