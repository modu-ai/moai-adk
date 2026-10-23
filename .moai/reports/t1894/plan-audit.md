auditor-model: claude-sonnet-5-5
verdict: FAIL
overall_score: 0.65
must_pass_failed: 1
blocking_count: 9
plan_artifact_hash: 1af447e8c0afd5d2f8ce2560ea8c647adabe1bdd3eb67982c496b16e33ed6fe6
audited_sha: cc109a75b6bf9c06ce86d811a8f0068551f02484
required_backend: codex fail
convergence_overall: fail
scope: full
fix_scope: acceptance.md#E-006 acceptance.md#AC-MJ-006 acceptance.md#AC-MJ-008 acceptance.md#AC-MJ-004 acceptance.md#AC-MJ-009 acceptance.md#AC-MJ-005 acceptance.md#AC-MJ-007 acceptance.md#AC-MJ-013 acceptance.md#4-quality-gates-and-definition-of-done plan.md#E-commits plan.md#M2 spec.md#REQ-MJ-002 spec.md#REQ-MJ-007 spec.md#REQ-MJ-010 research.md#10
defect_class: D1=design D2=ac-wording D3=traceability D4=design D5=traceability D6=ac-wording D7=design D8=design D9=ac-wording

# SPEC Review Report: SPEC-CARD-MODE-JUDGMENT-001 (card t1894), iteration 0

Auditor: plan-auditor, independent first audit. Tier L (declared), PASS threshold 0.85, aggregate 0.65 (below threshold). Reasoning context from the author: none was passed beyond the artifact files; the author's disclosures in the prompt were judged, not accepted (M1 Context Isolation: only spec.md, plan.md, acceptance.md, design.md, research.md, decision-index.md, progress.md were read).

Judging builds: the tree is HEAD cc109a75b6bf9c06ce86d811a8f0068551f02484 (clean, unpushed). Every `moai` CLI observation (spec lint, decision read, todo next/add help, verify audit-plan) was produced by the installed build moai_cp/20261007_195927; the audit_multi server reports build commit 7705b365e2, an ancestor of this HEAD (binary lag, so CLI output describes older code). No Go command, build or test was run (documents-only plan audit).

## Result summary

The required codex gate failed (receipt below), one must-pass criterion fails (MP-9, a cross-artifact ordering conflict confirmed by reading both texts), and nine blocking defects stand. The SPEC is close in structure and its authority quotations are faithful, but its release-blocking criteria can be satisfied by widening or reversing the very rules they guard, and it omits a site and a requirement half. See Defects.

## Must-Pass Results

- [PASS] MP-1 REQ number consistency: 17 REQ definitions collected, REQ-MJ-001..017 sequential, no gaps or duplicates (python trace of `^- **REQ-MJ-nnn**` lines in spec.md:L69-L97; sequence check printed True).
- [PASS] MP-2 GEARS format (requirement layer, spec.md only; ACs are Given-When-Then in acceptance.md and were graded under Testability): every REQ-MJ entry carries a shall clause in a Ubiquitous/While/When/Where form. One REQ mixes a non-shall clause (spec.md:L72, "A card that reached the queue on the default is reissued ..."), recorded as MINOR D15.
- [PASS] MP-3 frontmatter: all 12 canonical fields present with correct types (spec.md:L2-L13: id, title, version "0.1.1" quoted, status draft, created/updated ISO dates, author, priority P1, phase "v3.2.0 target" a release label, module, lifecycle spec-anchored, tags string). `moai spec lint --strict` printed `0 error(s), 0 warning(s)` plus one INFO OwnershipTransitionUnmeasured, lint_exit 0. Extra keys `tier`, `card`, `related_specs` are outside the 12-field table; lint accepted them. plan/acceptance/design/research/decision-index carry no status field.
- [N/A] MP-4 language neutrality: the SPEC edits repository process documents, not multi-language tooling; no language tool is named.
- [PASS] MP-5 D7 cross-SPEC: referenced SPECs SPEC-ALWAYS-LOADED-BUDGET-001, SPEC-FACTORY-LEASE-RENEW-001, SPEC-ROLE-INJECTION-BUDGET-001, SPEC-TODO-CLASSIFY-DISPATCH-001 all exist with status completed (none retired/superseded/archived); no REVIEW line.
- [PASS] MP-6 D8: `grep -c syscall spec.md` printed 0, so D8 auto-passes.
- [PASS] MP-7 clarification gate: `grep -rn '\[NEEDS CLARIFICATION' plan.md research.md` printed nothing (exit 1, no marker).
- [PASS] MP-8 RED-now re-execution: every release-blocking criterion's cell was re-executed on this tree and reproduced (table below). The three git-based cells (E-010a, E-013a, E-017) reproduced stdout; their exit codes could not be captured because the worktree guard refuses a git command with a trailing `; echo` (Gap G3).
- [FAIL] MP-9 cross-artifact ordering consistency: see D1. The CN-4 awk verb was refused by the worktree guard (Gap G2); the plan carries no `Exit:` lines, so no mechanical CONFLICT could be emitted, and the ordering candidates were read by hand. One candidate is a confirmed conflict (D1).

## Category Scores

| Dimension | Score | Band | Evidence |
|---|---|---|---|
| Clarity | 0.70 | between 0.50 and 0.75 | REQ-MJ-002 rests on undefined "in-flight" membership (decision-index Q5 default, spec.md:L116) and omits "shared state" (D4); REQ-MJ-007's condition "operator has delegated dispatch" is unobservable (R-5, spec.md:L138); spec headings cite "scope item 1..4" (spec.md:L67, L75, L87, L95) that the SPEC never lists |
| Completeness | 0.75 | 0.75 | all sections and 12 fields present; Out of Scope has three H3 groups with specific bullets (spec.md:L101-L117); site list omits auto-semantics.md:L395-L396 (D3) and the leader-duty enumeration in factory-dispatch-mechanics.md (D8) |
| Testability | 0.50 | 0.50 | AC-MJ-006 widening passes outside the pinned subsection (D2); AC-MJ-004(b), AC-MJ-008, AC-MJ-009 each admit a writable mutant (D6); AC-MJ-013 can go red on pre-existing text (D9) |
| Traceability | 0.65 | between 0.50 and 0.75 | 17 REQ, 19 AC, no orphan; the verb-form trace prints UNCOVERED for REQ-MJ-002, -003, -009 because the mapping is written as hyphen shorthand ("REQ-MJ-001, -002, -003") that is not machine-resolvable, and two requirement halves (REQ-MJ-006 inclusion clause, REQ-MJ-009 no-polling clause) have no verifying token (D5) |

Aggregate (mean of four): 0.65. Below the Tier L threshold 0.85.

## Authority fidelity (the card-specific check)

Board read: `moai decision read > /private/tmp/lxq/aud/board.txt` (1898 lines, exit 0), then grep by id.

| Authority | On the board (board.txt line) | In the SPEC | Judgment |
|---|---|---|---|
| 0c51 d-20261019T020637Z-0c51, decided_by operator, standing | L24-L25, body "Card issuance must carry ... (cannot be added after assign)." | spec.md:L36 quote | verbatim, identical text |
| a672 d-20261019T014804Z-a672, decided_by leader under operator instruction, standing | L34-L35 | spec.md:L40 quote | verbatim |
| 02f4 d-20261019T015400Z-02f4, operator, scope card t1737 | L30-L31 | spec.md:L42 | paraphrased under a heading that says "quoted, not paraphrased" (MINOR, D15); it creates no policy |
| d314 d-20261019T023441Z-d314, decided_by operator, scope card t1894, kind ruling | L4-L5 | spec.md:L46, decision-index.md Q3 fenced block | record line and body match the board character for character |

Policy statements checked against the quotations:
- REQ-MJ-002/-003 (spec.md:L70-L71): 0c51 says the mode is serial when the card "edits the same files/shared state as an in-flight card or a slot-gated surface". The REQ and the design wording say only "overlaps the file set". The words "shared state" were dropped. A card that shares state without sharing a file is parallelizable under the REQ and serial under the operator rule. Defect D4.
- REQ-MJ-007 (spec.md:L78) against the d314 body: the seven limits in d314 are the delegation condition, a free live lane (a), the parallelizable classification (b), the priority/readiness order (e), never keep-set/held/[보류 (c), serial cards follow the slot rules (d), lane self-dispatch unchanged (g). All seven are present. Two additions are not in d314: limit (f) "the leader's act is promotion to `picked` and assignment, nothing else" and the lead words "the leader — and no other actor". Both only restrict, but the REQ claims "every limit of d314 and no other". Defect D7.
- REQ-MJ-008/-009 and the a672 (a)-(c) list: faithful. REQ-MJ-010 (scoping the standing recheck cron, keeping its stall-recovery purpose) extends a672's "no polling crons for slot or trigger waits" by inference; reasonable, and it creates one site gap (D3).
- REQ-MJ-004/-005/-015: faithful to 0c51; the permanent "reissue" wording drops 0c51's "until then" qualifier, which is intended (the documents must not name the unlanded verb) and is recorded here only as residual (no owner is named for the later edit when the verb lands).
- The SPEC adds no other policy: Q1 and Q4 are confirmed by the ruling's last sentence (d314 body, "Q1 default (manager-todo untouched) and Q4 (laneRecheckRule moved to t1695) stand"); card t1695 exists in the todo queue (`moai todo list --json` printed it).

## DEFAULT-APPLIED and POLICY-COVERED stamps (judged against manager-spec.md:L121 and the sibling first audits)

The published rule (manager-spec.md:L121) allows DEFAULT-APPLIED only on an implementation-level FOUNDER row that carries `Default:` chosen by the ranked rule and an empty verdict. Precedents: the first audits of the sibling plans (t1721 plan-audit.md D2, t1722 plan-audit.md D1/D3) did not fail the stamp format itself; they failed on gate state restated three ways, on an unreconciled start rule and on a requirement contradicting a sibling SPEC.

| Row | Class / stamp | Allowed? | Does it silently decide an operator matter? |
|---|---|---|---|
| Q1 agents untouched | implementation-level, DEFAULT-APPLIED | yes (rule 1: preserves current behavior; format `DEFAULT-APPLIED <UTC> <runner+role>` correct) | no; d314 confirms the default, so the stamp is also covered by an operator verdict |
| Q2 doc-generation.md excluded | implementation-level, DEFAULT-APPLIED | formally yes; the row itself calls its Alternate "a product-level question" (decision-index.md:L23), and the excluded line `doc-generation.md:348` is a bare `moai gtd add` that 0c51 ("No card is issued on the fail-safe serial default") would forbid if standing-source cards are leader-issued; F:32 lists standing sources under the leader's production authority. The class label is doubtful and the exclusion leaves a known 0c51 site (D12) | partly: it keeps a violating site alive without an operator word |
| Q3 promotion | FOUNDER product-level, met by the operator ruling | yes; verdict is a real operator record, not a planner default | no; REQ-MJ-007 follows the ruling (with D7 additions) |
| Q4 Go notice out of scope | implementation-level, DEFAULT-APPLIED | yes, and moved to t1695 by the ruling | no |
| Q5 in-flight membership / slot-gated paths | implementation-level, DEFAULT-APPLIED | formally yes | it leaves REQ-MJ-002's core terms undefined while 25 merge-ready cards (a672 evidence) sit in the queue, so whether a merge-ready card is "in flight" decides most outcomes of the new judgment; a Clarity cost, not a stamp violation (D12) |
| Q6 Tier | POLICY-COVERED | no: the label requires an operator setting in `.moai/config/sections/*.yaml` or a constitution clause covering the question as written (manager-spec.md:L121 steps); the anchor is a workflow rule table whose own text says Tier is judged by a Socratic AskUserQuestion, and the row states "the plan author's reading" and that the question "could not be asked". Tier L is nevertheless robust: 17 REQ and 19 AC exceed the Tier M ceilings of 16/16 (spec-workflow.md ceilings table). Defect D11 |

## RED-now re-execution (MP-8) and mutant tests

Commands were run from the worktree root against HEAD cc109a75b6 (documents identical to the pinned trees per research.md L3). Ledger-cell comparison:

| Cell | Observed stdout | Matches cell |
|---|---|---|
| E-001 (awk, C both sides) | `0` / `0`, exit 0 | yes |
| E-002a (git grep V sweep) | one line `gtd.md:52 ...` | yes |
| E-002d | `gtd.md:0` ×2, exit 1 | yes |
| E-003 | `:0` ×2, exit 1 | yes |
| E-004a / E-004b | `:0` ×2 exit 1 / `:1` ×2 exit 0 | yes |
| E-005a / b / c | `:0` exit 1 / `:1` exit 0 / `:0` exit 1 | yes |
| E-006 (awk, nine lines) | `0 0` / `0 0`, exit 0 | yes |
| E-007 | `0` / `0`, exit 0 | yes |
| E-008a / b / c | K 1, F 2, A 1 (both trees), exit 0 / all eight `:0`, exit 1 / `:1` ×4, exit 0 | yes |
| E-009 | `0` / `0`, exit 0 | yes |
| E-010a | empty (git grep -c) | yes (exit not captured) |
| E-011 (cmp ×7) | all seven exit 0 | yes |
| E-011f | cmp on factory-dispatch.md differs at char 23127, line 222, exit 1 | yes |
| E-012a | `OK 9999`, exit 0 | yes |
| E-013a | empty | yes (exit not captured) |
| E-016 | empty, exit 1 | yes |
| E-017 | ` 1 file changed, 1 insertion(+), 1 deletion(-)` | yes (exit not captured) |
| E-018a / b / c / d | `:0` ×2 exit 1 / `:1` ×4 exit 0 / `:0` ×4 exit 1 / `:0` ×2 exit 1 | yes |
| E-019a / c | `:1` ×2, exit 0 | yes |

Specific re-checks asked for:
- (a) Eight pairs: seven `cmp` exit 0; factory-dispatch.md differs only at line 222 and `git diff --no-index --shortstat` prints 1 insertion, 1 deletion. Confirmed.
- (b) manager-todo.md (template): no `classification-file`, `moai todo add`, or issuance step; line 92 says "issue an audit" (unrelated) and line 54 "self-promotes, admits, drops, or edits cards". Confirmed.
- (c) Core stub: `OK 9999`. Arithmetic checked: "The one reconciliation" (22 chars) to "A reconciliation" (16) is -6; the marker form on K:28 replaces " to the completion report" (25 characters) by "while the lane executes the card, " (34) which is +9; net +3; a trim of at least 2 units from an uncovered line is feasible (K:28, K:3, K:84 measured uncovered: no ledger `after_text` contains "A stuck lane records a disk wait"; grep of binding_ledger.json finds only ALB-0276's compressed F text). The offsetting trim is not yet chosen and no test other than the ledger was read for those lines (the author's own G-7). Feasible, unproven.
- (d) Standing-cron sites: the author's list (K:28, F:119, F:122, A:117-118, M:26, W §0a) is not complete: `auto-semantics.md:395-396` says "the standing carrier of §5.1 stays armed" while a wait on the leader is open (an explicit non-executing state); research.md:L61 (S14) identifies it ("also 392-396"), but REQ-MJ-010, plan.md M2 step 4 and AC-MJ-008 do not. Also unscoped arm/keep statements: `auto-semantics.md:103,201`, `moai-lane-watchdog/SKILL.md:244` ("armed at card intake"). Defect D3.
- (e) Tests and ledger rows: all 19 named test functions exist (`grep -rn '^func Test...' internal`: TestAutoPickMirrorParity, TestAutoRankDoctrineAmendment, TestAutoPickDocDoctrine, TestAutoRankMirrorParity, TestAutoRankMarkerDisclosure, TestTodoSkillDocumentsClassification, TestRuleTemplateMirrorDrift, TestDeclaredRuleMirrorForks, TestReviewOwnership_CardReviewDoctrine, TestFactoryDispatchBodyCarriesRenewIntervalLinkage, TestLaneRecheckDoctrine, TestLaneWatchdogRenewsFirst, TestLaneWatchdogRecordsRefusalAndMeasures, TestBindingLedgerIntegrity, TestRoleCoreStubGate, TestDeployedAlwaysLoadedCharBudget, TestWorkflowRulePathsPinned, TestTemplateNoInternalContentLeak, TestCatalogHashCoversSkillSubfiles), `autoDocGTDProhibitions` exists in todo_auto_doc_test.go, and ledger rows ALB-0254/-0256/-0257/-0265/-0268/-0276/-0304 exist with the claimed locations (ALB-0257 `always:` core verbatim, ALB-0276 `role-core:` rewrite). Verified by reading the test sources: row coverage compares `before_text` with the pinned anchor blob (binding_ledger_test.go:L519), so rewriting live text does not break `before_text`; after-text presence is checked against the deployed location (L824-L892), so the plan's same-commit `after_text` updates are required and sufficient. TestAutoRankMirrorParity compares the F passage from the "Promotion is the operator's act, in person or in advance." marker to "The self-dispatch lane exception." byte-for-byte and the neutral regex covers SPEC/REQ/date/hex only; the planned pointer line passes it. The "renew-interval wording" test is TestFactoryDispatchBodyCarriesRenewIntervalLinkage (renew_interval_body_test.go:L28). TestAutoPickDocDoctrine `absent` strings are not written by the planned text. I did not run any of them.
- (f) AC-MJ-006 mutants, run with the E-006 command on files under /private/tmp/lxq/aud/m/ (good copy built from the nine pinned lines):

| Mutant | Output | Caught |
|---|---|---|
| good copy | `9 9` | control |
| remove "a held card," from the never-promoted line | `8 9` | yes |
| remove the delegation condition from line 1 | `8 9` | yes |
| replace the serial line with "Serial cards: promoted too." | `8 9` | yes |
| trailing space on the order line | `8 9` | yes (brittle by design) |
| duplicate a pinned line | `9 10` | yes |
| extra authority under a `####` sub-heading right after the nine lines ("Held cards and serial cards may also be promoted while dispatch is delegated.") | `9 9` | NO |
| extra authority under a separate `###` heading ("Held cards are promoted too.") | `9 9` | NO |
| widening placed in the Routing bullet of the same file | `9 9` | NO |

  The three escapes are inside file D, which AC-MJ-016 already expects to list, and none contains the exact token "while the operator has delegated dispatch". The criterion's claim "makes a widening visible however it is written" (acceptance.md:L76) is false. The independent codex run reproduced the sub-heading escape. Defect D2.
- (g) REQ-MJ-007: carries all seven d314 limits, plus two additions (D7).

Per release-blocking criterion mutant reasoning (beyond the above): AC-MJ-001 passes a reversed predicate (declared residual R-2); AC-MJ-002 passes an instruction written with single quotes or without a quoted description (V keys on `add "`); AC-MJ-003 passes a [HARD] line saying the flag is optional; AC-MJ-004(b) passes a reordered sentence (D6); AC-MJ-005 passes "a lane holding merge-ready cards is not a free lane; no card is actually executing means busy" (D5); AC-MJ-007 passes the seven tokens in a contrary paragraph (declared residual, but REQ-MJ-009's prohibition has no token, D5); AC-MJ-008 passes when "completion report" is deleted from unscoped sentences while one marker per file exists (D6); AC-MJ-009 passes a clause that renews every card (tokens `executing`, `merge-ready` present, polarity unchecked; D6); AC-MJ-016 passes a paraphrase (declared); AC-MJ-018 is whole-line and holds.

## Defects Found

D1. MP-9 ordering conflict, plan commit scheme versus Definition of Done — plan.md:L91 and acceptance.md:L570 — plan.md:L91: "One commit per milestone ... template and deployed files of one milestone together"; acceptance.md:L570: "Template-First order held: every template edit precedes the deployed edit (visible in the commit order of M1-M3)." Template and deployed edits of a milestone land in one commit, so the commit graph cannot show one before the other; the DoD clause cannot be satisfied by following the plan (verification-claim-integrity.md §2.3: only the commit graph witnesses ordering). The independent codex run found the same pair. — Severity: BLOCKING (critical, must-pass) — Class: blocking — Required fix: either split each milestone into a template commit followed by a deployed commit in plan.md §E, or rewrite the DoD clause to what the graph can verify (deployed byte-identical to template at each milestone commit, AC-MJ-011), and say so in both documents.

D2. AC-MJ-006 passes a widened exception placed outside the pinned subsection — acceptance.md:L306 (E-006) and L68-L78 — the awk ends the subsection at any `^##` line, so extra authority under a `####` or `###` heading, or in the Routing bullet, still prints `9 9` (nine mutant runs above; three escapes). The text at L76 claims a widening is visible "however it is written". REQ-MJ-007's "every limit and no other" is not enforced for the file as a whole. — Severity: BLOCKING — Class: blocking — Required fix: scan from the subsection heading to the next heading of level `##` or `###` that is not part of the nine lines and fail on any heading nested or adjacent that carries a promotion-authority sentence; add the three escape mutants to the mutant list and observe them red; or count lines from the heading to the next `^## ` only and require the next heading to be exactly `## The card-review stage`.

D3. A site that keeps the standing recheck cron is omitted from the scoping work — auto-semantics.md:L395-L396, research.md:L61, plan.md:L33, spec.md:L84, acceptance.md:L93 — "the standing carrier of §5.1 stays armed" is stated while a wait on the leader is open, so after the planned scoping §5.1 and §14 disagree and REQ-MJ-009 (no cron for a trigger wait) is contradicted. Research S14 lists it; REQ-MJ-010, plan M2 step 4 and AC-MJ-008(a) (which needs "completion report" on the same line) do not. Also unscoped arming statements at auto-semantics.md:L103 and L201 and moai-lane-watchdog/SKILL.md:L244. — Severity: SHOULD-FIX — Class: blocking — Required fix: name §14 (and the other arm statements, or state why they stay) in REQ-MJ-010 and plan M2, and add a check that fails when a statement keeping the cron during a non-executing wait remains.

D4. REQ-MJ-002 narrows operator rule 0c51 — spec.md:L70 and design.md:L98 versus board.txt:L25 — the rule says "edits the same files/shared state as an in-flight card"; the REQ says the file set overlaps. A card sharing state without a shared file is judged parallelizable under the SPEC and serial under the rule. AC-MJ-001 carries no token for it. — Severity: SHOULD-FIX — Class: blocking — Required fix: restore "files or shared state" in REQ-MJ-002/-003, the design candidate wording and the AC-MJ-001 token list (or record an operator verdict for the narrowing).

D5. Requirement halves and traceability notation without a verifying criterion — spec.md:L77 (REQ-MJ-006: a lane holding only merge-ready, held, or serial-waiting cards is free), L83 (REQ-MJ-009: no polling cron for a slot or trigger wait), acceptance.md:L58-L66 (AC-MJ-005 tokens: `no card is actually executing`, absence of the old sentence, `not a free lane`), L80-L88 (AC-MJ-007: seven tokens, title says "five facts"; none about a polling cron), L16, L80 (hyphen shorthand) — the inclusion clause of REQ-MJ-006 has no token and the required token `not a free lane` can be satisfied by a sentence saying a merge-ready lane is not free; REQ-MJ-009's prohibition has no token; the trace verb prints UNCOVERED for REQ-MJ-002, -003, -009 because the mapping is written "REQ-MJ-001, -002, -003" and "REQ-MJ-008, -009" (a hyphen number is not an expansion the verb performs). — Severity: SHOULD-FIX — Class: blocking — Required fix: write the full ids in each AC header, add a positive token for the inclusion clause (for example `merge-ready` and `free` in the Routing bullet) and for the no-polling clause (for example `polling`), and fix the "five facts" title.

D6. Mutant probe shallow on three release-blocking criteria — acceptance.md:L48-L56 (AC-MJ-004b), L90-L98 (AC-MJ-008), L100-L108 (AC-MJ-009) — (i) AC-MJ-004(b) is a negative fragment: the sentence reordered ("On a picked card an empty `spec_id` is the normal state") passes while violating REQ-MJ-005, and S5 (gtd.md:330-333, "Attach the SPEC only when one exists") is checked by nothing; (ii) AC-MJ-008: deleting the words "completion report" from the unscoped sentences while one marker per file exists satisfies (a), (b) and (c) — the mutant list tries only the deletion of the whole sentences and a wrapped tail; (iii) AC-MJ-009: a §0a that renews every card passes (`executing`, `merge-ready` present; no polarity token). verification-completeness.md §2: a criterion with a writable mutant is too shallow to adopt. — Severity: SHOULD-FIX — Class: blocking — Required fix: for each, add a token or whole-line check that fails on the mutant (for example a required sentence "does not renew" inside §0a; a per-sentence check that each cron-keeping sentence carries the marker, not a per-file count) and record the mutant as observed red.

D7. REQ-MJ-007 claims "every limit of d314 and no other" but adds two — spec.md:L78, acceptance.md:L306 (pinned lines 1 and 7) — "the leader — and no other actor" and limit (f) "the leader's act is promotion to `picked` and assignment, nothing else" are not in the d314 body (board.txt:L5). They restrict, so they are not widening, but the SPEC states it adds no policy of its own (spec.md:L30). — Severity: SHOULD-FIX — Class: blocking — Required fix: delete (f) and "no other actor", or obtain the operator's word for them and cite it; keep the pinned lines in step.

D8. The new leader duty is not in the leader-duty enumeration the documents declare to be the boundary — factory-dispatch-mechanics.md:L169-L177 versus spec.md:L78 and research.md §10 — the section says "What the leader keeps is enumerated, and the enumeration is the boundary — a duty not listed is a lane's"; promotion of queued cards is not listed, yet the new subsection gives it to the leader and pinned line 5 of E-006 cites this very section. The §10 sweep searched for operator-only wording and missed this restriction. — Severity: SHOULD-FIX — Class: blocking — Required fix: add the delegated promotion to the enumeration by pointer (home D), or state in D that it is an addition to that table, and extend REQ-MJ-017 and the sweep to this section.

D9. AC-MJ-013 can go red on pre-existing text — acceptance.md:L141-L150, auto-semantics.md:L126 and L146 — the `-G` pattern fires on any added or removed line that matches; auto-semantics.md:L126 ("observed on card t1393") and L146 ("cards t1393 and t1339") sit in the §5.1 block (L113-L146) that plan.md M2 step 4 reflows, so a re-wrap that moves those lines makes the criterion fail for text the work did not introduce, the "wrong-reason red" of verification-completeness.md §2. The cell is green only by emptiness today. — Severity: SHOULD-FIX — Class: blocking — Required fix: measure added lines only (for example `git diff -U0` filtered to lines beginning `+`) or allow-list the two existing ids, and observe a mutant that adds a new id red.

D10. Open product question surfaced but given no decision row — spec.md:L138 (R-5), progress.md:L31 — "what establishes 'the operator has delegated dispatch'" is described as a possible FOUNDER, product-level question and "not authored". manager-spec.md:L121 asks for one row per surfaced decision. The ruling's own words are carried verbatim, so the REQ is faithful; the gap is that an actor can treat an unstated delegation as given and no command sees it. — Severity: SHOULD-FIX — Class: optional — Required fix: add a decision-index row (FOUNDER, product-level, empty verdict) or record why the operator's phrase is the whole condition.

D11. Q6 is labelled POLICY-COVERED without a covering setting or constitution clause — decision-index.md:L64-L69 — see the stamp table. The outcome (Tier L) is robust because REQ 17 and AC 19 exceed the Tier M ceilings. — Severity: SHOULD-FIX — Class: optional — Required fix: relabel FOUNDER (implementation-level, Default Tier L) or cite a setting that covers the question as written.

D12. Q2 and Q5 defaults leave consequences unremarked — decision-index.md:L16-L24, L54-L62 — Q2: the row's Alternate is "a product-level question" yet the row is implementation-level; the excluded doc-generation.md:348 line is a bare issuance 0c51 forbids. Q5: leaving "in flight" undefined (merge-ready and held cards) lets two issuers judge the same card differently, which is the problem the card exists to remove; the SPEC admits it (R-2). — Severity: SHOULD-FIX — Class: optional — Required fix: state in REQ-MJ-004 or the open items that the doc-generation site remains a known 0c51 gap with an owner; consider defining "in flight" as "any card not yet landed" by operator verdict.

D13. Coexisting absolute sentence and exception — factory-dispatch.md:L38 and L40 (kept verbatim) versus the planned pointer line (acceptance.md E-018a) — a reader meets "Outside an --auto authorization the leader never picks ... never silently promotes" and, two lines later, "may promote queued cards in exactly one further case". Disclosed (spec.md R-6); the operator asked for the case to be "stated explicitly" and the pinned sentences make the statement an exception to an absolute. — Severity: MINOR — Class: optional — Required fix: none required now; record that the Go doctrine tests must be amended in a later card to fold the exception into the sentence.

D14. Verification-form and environment weaknesses — acceptance.md:L333-L342 (E-008a), L435-L440 (E-012a), L451-L455 (E-013a) — `grep -c -P` and `git grep -P` need a PCRE-capable grep: the independent codex run got `invalid option -- P`, exit 2, in its environment, while it ran here (the shell's grep resolves to ugrep 7.8.4); the cells state no tool dependency. E-012a prints OVER with exit 0, so a gate reading the exit code is blind (report-not-verdict, verification-completeness.md §1.1). — Severity: SHOULD-FIX — Class: optional — Required fix: name the required grep/PCRE support in the conventions, or use a portable awk form; make E-012a exit non-zero on OVER.

D15. Housekeeping — spec.md:L32 cites `/private/tmp/lxq/board-spec.txt` (machine-local scratch) as the source of the quotes; spec.md:L42 paraphrases 02f4 under "quoted, not paraphrased"; spec.md:L67, L75, L87, L95 cite "scope item 1..4" that the SPEC never lists; spec.md:L72 mixes a non-shall clause into REQ-MJ-004; frontmatter carries `card` and `related_specs` outside the optional-field table. — Severity: MINOR — Class: optional — Required fix: cite the board ids only; list the four scope items in §1; rewrite the reissue clause as a shall clause.

Stagnation / regression: first iteration; none.

## Recommendation

Fix the blocking set before re-audit (delta scope = the fix_scope list; the confirming re-audit re-runs the ordering check in full):
1. D1: separate template and deployed commits in plan.md §E, or restate the DoD ordering clause as a graph-verifiable one in both plan.md and acceptance.md.
2. D2: make AC-MJ-006 detect authority placed anywhere in file D; observe the three escape mutants red.
3. D3: add auto-semantics.md §14 (and the other arming sites or a stated reason) to REQ-MJ-010, plan M2 step 4 and AC-MJ-008.
4. D4: restore "shared state" in REQ-MJ-002/-003, design D-9, AC-MJ-001.
5. D5: write full REQ ids in AC headers; add tokens for the REQ-MJ-006 inclusion clause and the REQ-MJ-009 no-polling clause; fix the "five facts" title.
6. D6: strengthen AC-MJ-004(b), AC-MJ-008 and AC-MJ-009 against the stated mutants.
7. D7: remove limit (f) and "no other actor" or cite an operator word.
8. D8: reconcile factory-dispatch-mechanics.md "The leader's remaining role" with the new leader duty.
9. D9: measure added lines only in AC-MJ-013.
The optional findings D10-D15 are surfaced for the orchestrator's discretion and do not by themselves justify a FAIL.

## Five-section evidence record

**Claim.** The SPEC fails its plan audit: one must-pass (MP-9) fails, the required codex gate failed, nine blocking defects stand, aggregate 0.65 against 0.85.

**Evidence.** Commands run from /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1894 (exit codes as observed; "n/c" = guard refused a trailing echo, output matched):
- `git rev-parse HEAD` printed cc109a75b6bf9c06ce86d811a8f0068551f02484; `git status --short` printed nothing.
- `moai decision read > /private/tmp/lxq/aud/board.txt` exit 0, 1898 lines; grep by id found d314 (L4), 0c51 (L24), 02f4 (L30), a672 (L34).
- `moai verify audit-plan --project-root <toplevel>`: config_status ok, model codex, codex required (enforced_required [codex]), claude and glm off, cross_model_required true. After the audit, `moai verify audit-plan ... --result-file .moai/state/audit-plan-result.json`: `convergence_check` ok true, unmet [] (the required backend answered; its answer was a fail, a reviewed result and not an unmet gate).
- `cmp` ×8 pairs: seven exit 0, factory-dispatch.md exit 1 (char 23127, line 222).
- python UTF-16 count of factory-dispatch-core.md: `OK 9999`, exit 0.
- All MP-8 cells as in the table; mutant runs on /private/tmp/lxq/aud/m/*.md as in the table.
- `moai spec lint --strict .../spec.md`: 0 errors, 0 warnings, 1 INFO, exit 0.
- `grep -rn '\[NEEDS CLARIFICATION' plan.md research.md`: no match, exit 1. `grep -c syscall spec.md`: 0. Referenced SPEC statuses: completed ×4.
- Python trace of REQ/AC: 17 REQ, 19 AC, sequential, no orphan; UNCOVERED by the verb's rule: REQ-MJ-002, -003, -009 (shorthand).
- `mcp__moai__audit_multi` called once (target baseBranch, project_root = this tree): overall result fail; codex (required) fail with four findings (P1 AC-MJ-006 sub-heading widening; P2 standing-cron site auto-semantics.md:396; P2 DoD commit order versus plan.md:91; P2 `grep -P` not executable in its environment). Receipt rcpt-39ba8947c14db9823fd47f7e; tool build commit 7705b365e2, ancestor of HEAD. The first, second and third codex findings were reproduced here independently (D2, D3, D1); the fourth is environment-dependent (D14).

**Baseline-attribution.** Every observation above was measured in this run against HEAD cc109a75b6 (documents identical to the pinned trees 74ad197380 / 55d4f1486a per research.md L3). CLI results were judged by the installed build moai_cp/20261007_195927, which predates HEAD (the audit_multi note says the installed binary was built from 7705b365e). Mutant files were constructed in this run under /private/tmp/lxq/aud/m/ and are scratch, not cited as evidence beyond the outputs printed above.

**Gaps (not observed).**
- G1: the plan artifact hash was not recomputed here (no audit_cache tool in this agent); the value written is the lane's runtime value for the directory at HEAD cc109a75b6 (`git status --short .moai/specs` printed nothing, so the directory is unchanged since commit).
- G2: the CN-4 ordering awk verb and `awk -f` forms were refused by the worktree guard ("too complex to verify"); a python/grep substitute plus a manual read of the ordering candidates was used instead (plan has no `Exit:` lines: grep exit 1).
- G3: git commands with a trailing `; echo $?` were refused; exit codes of E-010a, E-013a and E-017 were not captured (stdout matched the cells). A shell `for` loop around git and two compound git forms were also refused and re-issued as plain commands.
- G4: the Grep tool is unavailable in this session; `grep` via Bash resolves to ugrep 7.8.4 (a shell function), so `-P` worked here.
- G5: no Go test, build or `make` target was run; every statement about what a guard test pins comes from reading its source (binding_ledger_test.go L480-L540 and L824-L892, todo_auto_doc_test.go L92-L260, todo_auto_pick_doc_test.go L30-L150, renew_interval_body_test.go L28-L60). Sections of those files not read: the mutation fixtures of the ledger test and the bodies of TestAutoPickDocDoctrine and TestAutoPickMirrorParity beyond the surface table.
- G6: research.md was read at L1-L135 and the §10 table; the remaining §10 evidence blocks were skimmed, not re-run.
- G7: whether card t1695 truly owns laneRecheckRule was not verified beyond the ruling's wording and the card's presence in the queue.

**Residual-risk.** (1) The nine blocking defects were found by reading and by mutants of the criteria; further escapes of the same kind may exist in criteria not mutated here (AC-MJ-001, -002, -003, -016). (2) The auditor's judgment on D4 depends on reading "shared state" as broader than "files"; the operator may have meant the same thing. (3) The CN-4 verb was not executed; a mechanical CONFLICT line elsewhere cannot be excluded, though the plan binds no criterion to a milestone through an `Exit:` line. (4) The audit_multi server and the installed moai predate HEAD.

## Operational Notes (unverified)

- Measure the escape rate of AC-MJ-006 after the fix — command: run the E-006 awk on /private/tmp/lxq/aud/m/mut_subsub.md, mut_h3.md, mut_routing.md; expected all three print a count other than `9 9`. Status: measured before the fix (all three printed `9 9`), unmeasured after.
- Measure whether the standing carrier sentence is the only unscoped keep statement in auto-semantics.md — command: `grep -n -i 'stays armed' .claude/rules/moai/workflow/auto-semantics.md`. Status: inferred (one hit at L396 seen by reading).
- Check the offsetting stub trim before the run phase — command: grep the repository for tests naming K lines 3, 26, 28, 48, 54, 62, 64 and 84. Status: assumption (the author read only the binding ledger).

AUDIT-VERDICT: FAIL spec=SPEC-CARD-MODE-JUDGMENT-001 receipts=rcpt-39ba8947c14db9823fd47f7e
