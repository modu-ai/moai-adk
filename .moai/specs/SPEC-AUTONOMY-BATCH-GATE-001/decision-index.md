# SPEC-AUTONOMY-BATCH-GATE-001 — Decision Index

Carrier for decisions surfaced during plan-phase assembly that the operator has not settled in an interview (`interview.decision_gate: on`, `.moai/config/sections/interview.yaml`). Rows state what is unresolved and why; no row carries a recommendation, a preferred answer, or a labelled option (`interview.recommendation_mode: pull`). Where the draft requirements currently read one way, the row says so as a fact about the draft, not as a preference. `Operator verdict:` lines are empty at authoring.

### Q1: Is the batch membership key the gate row, the shared decision subject, or both?

Label: FOUNDER
Authority anchor: (none — `.claude/rules/moai/workflow/auto-semantics.md` §9 defines gate rows but fixes no batching key; no prior SPEC row decides it.)
Why unresolved: the draft carries two separate clauses — REQ-BGS-001 groups rows by gate row into one summary, REQ-BGS-003 asks a shared judgment once across cards. Keying on gate row alone drops the one-question-for-many case (the five-cards-one-judgment instance in the source report). Keying on decision subject alone drops the multi-row summary for same-gate cards with different evidence. Keeping both widens what a reviewer must hold in view at once and doubles the surface the guard test pins.
Operator verdict:

### Q2: May a reserved (keep-set or authority-reserved) judgment with an identical decision subject across several cards be asked once, naming every card?

Label: FOUNDER
Authority anchor: (none — `auto-semantics.md` §9 says keep-set cases "keep the operator answer"; it does not say whether one answer may name several cards.)
Why unresolved: the card says the grade-3 exceptions keep individual approval. Reading that as one approval token per card means a reserved judgment shared by N cards is asked N times. Reading it as the operator's answer staying the operator's — while one question names all N cards and each card still gets its own record — leaves the answer with the operator but reduces the asks. Both readings keep a human answer; they differ on what "individual" means. The draft's REQ-BGS-003 excludes reserved rows from the ask-once rule, so it currently follows the first reading until ruled.
Operator verdict:

### Q3: Does the summary also list reserved and blocked rows for visibility, or only the approvable rows?

Label: FOUNDER
Authority anchor: (none — no committed policy fixes the report's row population.)
Why unresolved: listing reserved and blocked rows in a separate section gives the operator one view of everything waiting, at the cost of a longer report and a risk of reading the single approval as covering them. Listing only approvable rows keeps the report short and the approval scope unambiguous, but the operator then sees reserved and blocked cards only through separate messages. The draft's REQ-BGS-005 requires the report to state that rows added later, reserved rows, and blocked rows are not covered; it does not require listing them, so it holds under either answer. REQ-BGS-004 carries a per-row field with the keep-set and leader-held-power check result and its basis, so the classification is visible for the listed rows whichever way this is ruled.
Operator verdict:

### Q4: Is reconciling the stale per-card "mandatory" Kickoff wording (spec-assembly.md:202-208, moai.md:144,240) in scope of this SPEC?

Label: FOUNDER
Authority anchor: (related precedent, not the identical question: `.moai/specs/SPEC-LANE-STALL-WATCHDOG-001/spec.md` §A.7 — its amendment list and its reviewed-no-change list name neither file; `grep -n "spec-assembly"` and `grep -n "workflows/moai.md"` on that spec each returned no hit, with `grep -n "kanban-dispatch.md"` on the same file as the positive control.)
Why unresolved: the default-autonomous Kickoff (`auto-semantics.md` §9.1) and these two skill files disagree on whether the per-card operator question is mandatory. Including the reconciliation here makes the batch doctrine and the surrounding wording agree, but widens the change to two more mirrored files, one at 597 of 600 lines, with pinned wording. Leaving it out keeps the change small and the conflict in place, possibly as a separate card. REQ-BGS-001..013 hold either way; only REQ-BGS-014 and plan milestone M4 depend on this.
Operator verdict: DECIDE — in scope (operator answer at Decision Point 1, 2026-10-02). SPEC-level verdict (scope of this SPEC), not product-level, so no `.moai/project/product.md` reconciliation is triggered.

### Q5: Is the launcher injection text (SessionStart notice) in scope of this SPEC?

Label: FOUNDER
Authority anchor: (none — the card names "launcher injection text" as a surface, but no committed policy says whether the leader notice must carry gate guidance; the current injected strings carry no approval or Kickoff wording, checked in `internal/hook/session_start_factory_i18n.go`, `session_start_kanban_i18n.go`, `lane_spawn_authority.go`.)
Why unresolved: the leader already loads `kanban-dispatch.md` every turn, so an injected sentence would repeat a rule the leader reads anyway. Adding it costs one sentence in four locales plus the pinned-string tests (plan milestone M5). Omitting it leaves the card's third named surface unaddressed. A lane notice is not a candidate because a lane holds one card.
Operator verdict: DECIDE — in scope (operator answer at Decision Point 1, 2026-10-02). SPEC-level verdict (scope of this SPEC), not product-level, so no `.moai/project/product.md` reconciliation is triggered.

### Q6: Where does the grade-3 individual-approval rule live in shipped text?

Label: FOUNDER
Authority anchor: (none — the grade doctrine is in `.moai/docs/jev-local-operations.md`, a local maintainer document that is not template-shipped; `scripts/jev/route.sh` is absent from the tree. Shipped counterparts exist: the keep-set definition in `auto-semantics.md` §9 and the leader-retained powers clause at `kanban-dispatch.md:106`.)
Why unresolved: template text cannot cite a document users do not receive. Candidates: express the exception by the keep-set categories plus the authority-gate invariant only; or also name the leader-retained powers list and point at the kanban clause; or add a new shipped definition. Each changes how closely the shipped wording tracks the local grade doctrine, and whether the two can drift apart. The draft's REQ-BGS-010 reads the second candidate: it lists the three keep-set categories of §9 and the six powers `kanban-dispatch.md:106` keeps with the leader session (final PASS/FAIL verdicts, final merge approval, operator gates, card issuance and `done`, CodeRabbit slot-wait adjudication, cross-session dispute coordination). The phrase "user-facing behavior changes" appears only in the local doctrine, not in shipped text, and is not in the requirement. REQ-BGS-010, `design.md` §D.5, and this row name the same source.
Operator verdict:

### Q7: Is the counter-evidence source list closed or open?

Label: FOUNDER
Authority anchor: (none — `auto-semantics.md` §10 has no counter-evidence field; no committed policy defines its sources.)
Why unresolved: three forms exist. A closed list (audit warnings or debt, margin to threshold, unresolved decision-index rows, divergent audit-cross opinions, open blockers or wait records, path overlap between rows): the field is checkable against fixed sources and the `none searched=` claim is auditable; a risk the list did not foresee has no place to be recorded. An open list ("the strongest evidence against"): it fits any case, including a risk nobody foresaw; the field cannot be checked mechanically and a `none` claim has no fixed set to be audited against. A closed list with an explicit "other" entry: the fixed sources stay checkable and an unforeseen risk has a place; the "other" entry cannot itself be checked mechanically, and its presence changes what the guard test can pin. The draft's REQ-BGS-007 reads the closed list; the draft's reason for that reading is in `design.md` §D.4.
Operator verdict:

### Q8: Does the batch mechanism need mechanical enforcement in Go, or does it stay document-only?

Label: FOUNDER
Authority anchor: (none — `.moai/specs/SPEC-LANE-STALL-WATCHDOG-001/spec.md` §A.4 kept its own doctrine zero-Go, which is a related precedent, not a ruling on this question.)
Why unresolved: the existing multi-card `moai factory decide` already applies one gate and choice to several cards but writes no decision record (`internal/cli/factory_card.go:1469-1491`). Document-only relies on the orchestrator to write per-row records and is checked after the fact by the sync audit's re-read (detection, not prevention). Adding Go (an automatic record writer, or a refusal when a row lacks one) prevents a missing record but changes a verb with other callers and adds a surface to maintain. This SPEC's scope currently includes four product Go files — the SessionStart leader-notice strings in `internal/hook/session_start_{kanban,factory}{,_i18n}.go` — and three test Go files; none of the seven enforces the batch mechanism. Whether the batch mechanism itself (record writing, summary validation) is enforced in Go is the open question, and it stays outside this SPEC's scope.
Operator verdict:

### Q9: What outcome metric and baseline define success, given that the gate-round baseline does not exist?

Label: EVIDENCE-NEEDED
Authority anchor: (none — the data does not exist: the source report's per-gate breakdown of its 176 rounds is absent, and an ad hoc count over this machine's transcripts gave 124 all-question rounds for 09-26..09-29 and 17 / 27 for 09-30 / 10-01, neither gate-classified; commands and limits are in `research.md` R3.0.)
Why unresolved: a reduction claim needs a gate-classified count before the change, over a defined scope (which transcript directories, which window, how a question is classed as a gate round), and a count after. Until the run phase records that baseline (REQ-BGS-015, at a tracked path inside the SPEC directory), "176 to 10-20" cannot be verified, and the size of the population left after the autonomous transition is unknown.
Operator verdict:

### Q10: Which term names the consolidated presentation?

Label: FOUNDER
Authority anchor: (none — "batch approval" is already used for the `/moai:todo --auto` queue-consumption authorization at `kanban-dispatch.md:31`, `gtd.md:334`, `auto-semantics.md:169`; no committed policy names the new concept.)
Why unresolved: three forms exist. Reusing "batch approval": it matches how the card and the source report speak, and it collides with the queue authorization, which has a different owner and different authority. Keeping "batch gate summary" (the term the draft uses): it is distinct from the queue authorization and adds a new term; "batch" alone also means a dispatch batch and the `release/vX.Y.Z` batch PR in `kanban-dispatch.md`. Another name, or the same name with a defined scope: it resolves whichever collision the ruling targets and changes every occurrence in the doctrine text, the leader-notice sentence, and the guard test anchors. The draft's reason for its choice is in `design.md` §D.7.
Operator verdict:

### Q11: Is Tier M the correct classification?

Label: POLICY-COVERED
Authority anchor: `.claude/rules/moai/workflow/spec-workflow.md` § SPEC Complexity Tier (classification table: M = 300–1000 LOC guidance, 5–15 files, 3-file artifact set, 16 requirement / 16 acceptance ceilings; L = more than 1000 LOC or constitutional, more than 15 files, 5-file artifact set, 25 / 25 ceilings, plan-audit PASS threshold 0.85).
Why unresolved: resolved by committed policy applied to measured facts, recorded here as an orchestrator ruling (2026-10-02, in response to plan-audit iteration 1 defect D6). Counting rule: files are counted by path, each live copy and each template-mirror copy separately, excluding paths under `.moai/specs/` and `.moai/reports/`. Measured count: 17 planned files (5 doctrine/skill pairs = 10, 1 template guard test, 4 product Go files, 2 hook test files); the plan-audit report's own recount reached 17–18. 17 exceeds Tier M's 15 and lies in Tier L's more-than-15 band; the table does not state that a mirror pair counts once. Requirements stood at 16 of Tier M's 16. The LOC axis is an estimate with no diff behind it (UNVERIFIED) and does not decide the tier. The recount command and plan-time measurement are in `spec.md` §A.5 and `research.md` R5.
Operator verdict: orchestrator ruling (applies the committed table; not an operator verdict): Tier L. `spec.md` carries `tier: L`; the artifact set is `spec.md`, `plan.md`, `acceptance.md`, `design.md`, `research.md`; the ceilings are 25 requirements and 25 acceptance criteria (19 and 17 used); the plan-audit PASS threshold is 0.85.

### Q12: Are the other files that carry the same stale plan→run Kickoff wording in scope of this SPEC?

Label: FOUNDER
Authority anchor: (related precedent, not the identical question: Q4's verdict above names `spec-assembly.md` and `workflows/moai.md` only — operator answer at Decision Point 1, 2026-10-02, a SPEC-level scope verdict.)
Why unresolved: the sweep in `research.md` R4.4 read the Kickoff lines of the 38 files that mention the word. Four files assert in the present tense that the operator answers every card or that a skip-eligible audit score cannot substitute for the operator question: `.claude/agents/harness/workflow-specialist.md:72-77`, `.claude/skills/hns-moaiadk-patterns/SKILL.md:150`, `.claude/rules/moai/workflow/session-handoff-examples.md:237`, `.claude/skills/moai/workflows/project/doc-generation.md:384,392`. Five files carry only the "HUMAN GATE" or "mandatory human gate" label or a condition-order sentence: `.claude/skills/moai/workflows/plan.md:53,73`, `.claude/skills/moai-workflow-spec/SKILL.md:173,191`, `.claude/skills/moai/workflows/design.md:48,52`, `session-handoff-examples.md:242,250`, `.claude/skills/moai/SKILL.md:166,372`. The classification is a reading, not a mechanical check. Including these files makes the Kickoff wording agree across the document set but widens the change to more files, some of them paths-scoped, with pinned wording, or with live-versus-mirror differences. Leaving them out keeps the change small and the disagreement in place, possibly as a separate card. REQ-BGS-001..013 hold either way; REQ-BGS-014 covers two files only and says so.
Operator verdict:
