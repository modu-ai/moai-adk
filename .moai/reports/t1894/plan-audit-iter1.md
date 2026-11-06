auditor-model: claude-sonnet-5-5
verdict: FAIL
overall_score: 0.79
must_pass_failed: 0
blocking_count: 4
plan_artifact_hash: 74d418b30f1f2ba8ae90cb22f41e21e4eafbe76ebea879994576e956f19b60e7
audited_sha: 4834700468ed2311e3efc4e12bb920db03db54e2
required_backend: codex fail
convergence_overall: fail
scope: delta
fix_scope: acceptance.md#AC-MJ-005 acceptance.md#E-005d acceptance.md#AC-MJ-006 acceptance.md#E-006 acceptance.md#AC-MJ-007 acceptance.md#AC-MJ-008 acceptance.md#E-002a acceptance.md#E-009 design.md#D-10 research.md#11 spec.md#6-open-items plan.md#C-order-and-dependencies acceptance.md#4-quality-gates-and-definition-of-done
defect_class: D16=ac-wording D17=traceability D18=format D19=other

# SPEC Review Report: SPEC-CARD-MODE-JUDGMENT-001 (card t1894), iteration 1 (delta audit after correction round 1)

Auditor: plan-auditor, independent second pass. Tier L (declared, spec.md:L14), PASS threshold 0.85. Reasoning context from the author: none was passed beyond the artifact files and the lane's task text; the author's closure claims in progress.md were re-executed, not accepted (M1 Context Isolation: only spec.md, plan.md, acceptance.md, design.md, research.md, decision-index.md, progress.md and the iteration 0 report were read).

Judging builds: tree HEAD 4834700468ed2311e3efc4e12bb920db03db54e2 (clean, unpushed; `git status --short` printed nothing). Every `moai` CLI observation (decision read, spec lint, verify audit-plan) was produced by the installed build moai_cp/20261007_195927; the audit_multi server reports build commit 7705b365e2, an ancestor of this HEAD (binary lag: CLI output describes older code). No Go command, build or test was run.

## Result summary

The correction round is real work: all nine blocking defects of the iteration 0 report are closed or reduced to a residual, the commit-shape conflict (MP-9) is gone, the traceability verb is clean (17 definitions, no UNCOVERED, no ORPHAN), and the four authority quotations still match the board character for character. The aggregate rose clearly (no regression, no STOP). The SPEC still fails: the required codex gate failed again (receipt below), and four blocking defects stand, three of them found independently here and by codex: a stem-less widening of the Routing bullet passes every release-blocking criterion (D16), two of the ten cron sites named by REQ-MJ-010 have no verifying check (D17), three release-blocking evidence commands are refused by the worktree guard as written (D18), and the SPEC states two incompatible things about what the open decision row Q7 gates (D19).

## Must-Pass Results

- [PASS] MP-1 REQ numbers: `python3` trace over spec.md printed `COLLECTED: 17 REQ definitions`; REQ-MJ-001..017 sequential, no gaps or duplicates (spec.md:L74-L102).
- [PASS] MP-2 GEARS (judged on the REQ layer in spec.md only; ACs are Given-When-Then in acceptance.md and were graded under Testability): every REQ-MJ entry is a shall clause in a Ubiquitous/While/When/Where form (spec.md:L74-L102). REQ-MJ-004 is now `When ... shall` throughout (spec.md:L77).
- [PASS] MP-3 frontmatter: 12 canonical fields present with correct types (spec.md:L2-L13); the unused `card` key is gone; `tier`, `related_specs` are outside the 12-field table and accepted. `moai spec lint --strict spec.md` printed `0 error(s), 0 warning(s)` plus one INFO `OwnershipTransitionUnmeasured`, lint_exit 0 (progress.md:L14 still says lint was NOT RUN; that line is stale, see D22).
- [N/A] MP-4: the SPEC edits repository process documents; no language tool is named.
- [PASS] MP-5 D7 cross-SPEC: referenced SPECs (spec.md:L15) all exist with status completed (iteration 0 measurement; no reference was added in the correction round: `git diff --stat` shows only the seven artifacts changed).
- [PASS] MP-6 D8: `grep -c syscall spec.md` printed `0`.
- [PASS] MP-7: `git grep -c -E 'NEEDS CLARIFICATION' -- plan.md research.md` printed nothing, exit 1.
- [PASS with a recorded refusal] MP-8 RED-now re-execution: 34 of the 42 ledger cells were re-run and reproduced exactly (table below); 3 were refused (below) and 5 known-failing-input control cells were not re-run (G9). Three release-blocking cells (E-002a, E-006, E-009) were REFUSED by the worktree guard as written. Per the execution discipline they are not recorded as passes; each was re-run in a guard-friendly form (the alternation moved to `awk -v re=...`) and the RED reproduced, but that is a different command, so D18 stands. No cell failed to reproduce.
- [PASS] MP-9 cross-artifact ordering: the CN-4 awk verb was not runnable under the guard (Gap G3); `grep -c -i -E '^(\*\*)?exit' plan.md` printed `0` (the plan binds no criterion through an `Exit:` line, so no mechanical CONFLICT can exist), and all 28 ordering candidates in acceptance.md (words before/after/first/prior to/pre-change) were read by hand. None binds a milestone against the plan's M1 → M2 → M3 → M4 → M5 order. The iteration 0 conflict pair is gone (see closure table D1). Run in full on this iteration, as the Retry Loop Contract requires.

## Category Scores

| Dimension | Score | Rubric band | Evidence |
|---|---|---|---|
| Clarity | 0.80 | between 0.75 and 1.0 | REQ-MJ-002/-003 restore "same files or shared state" (spec.md:L75); REQ-MJ-007 lists exactly seven limits (spec.md:L83); one inconsistency in what Q7 gates (D19, spec.md:L139 versus decision-index.md:L84) |
| Completeness | 0.85 | 0.75-1.0 | all sections and 12 fields present; Out of Scope has three H3 groups with specific bullets (spec.md:L106, L113, L121); site list now complete (ten cron sites, research.md:L211-L228) |
| Testability | 0.65 | between 0.50 and 0.75 | stem-less Routing sub-bullet passes E-005d and E-006 (D16, acceptance.md:L367, L376); two cron sites unchecked (D17, acceptance.md:L106-L115); three cells refused by the guard (D18) |
| Traceability | 0.85 | 0.75-1.0 | verb port: 17 collected, no UNCOVERED, no ORPHAN; full REQ ids in every AC header (acceptance.md:L20-L224); residual: REQ-MJ-004's reissue half and REQ-MJ-003's "name the conflicting card" carry no token (D20) |

Aggregate (mean of four): 0.79, below the Tier L threshold 0.85. Higher than the iteration 0 report's aggregate: no STOP signal.

## Closure table, D1 to D15 (each row re-executed on HEAD 4834700468)

| Id | Status | Evidence (before = iteration 0 report, after = this run) |
|---|---|---|
| D1 MP-9 commit shape | closed | before: plan.md "one commit per milestone" vs DoD "template edit precedes the deployed edit". After: plan.md:L7 and L93 state "together with its deployed twin ... never a template-only commit followed by a deployed-only one"; acceptance.md:L148 and L658 state the same graph-verifiable shape and say the order inside a commit is not claimed; spec.md:L95 (REQ-MJ-013) says the same. design.md carries no ordering clause (grep of design/research/progress/decision-index/spec for template-first/precede/diff-tree: only L116 of design.md, consistent). The `git diff-tree --no-commit-id --name-only -r <sha>` command is a real read-only single invocation (run on HEAD: exit 0, seven paths). It lists, it does not decide (D21, optional). |
| D2 AC-MJ-006 scan | closed for the three escapes; residual under D16 | before: `9 9` on the good copy and on three escapes. After: 22 scratch files built in /private/tmp/lxq/aud/m2/ and scanned with the E-006 command (-v form): good `8 8 1 1`; `####` widening `8 10 1 1`; separate `###` widening `8 10 1 1`; Routing parent line with a stem `8 8 1 2`. The author's 16 red mutants and the good copy reproduce exactly (numbers in the Evidence block); N9 (stem-less edit of the Routing parent line) prints `8 8 1 1` for E-006 and `2` for E-005d (d): caught by AC-MJ-005 as claimed. Claim "16 of 17 red, N9 by AC-MJ-005" verified. NEW mutants pass: see D16. |
| D3 cron sites | partly | the set of ten is complete: `git grep -n -i -E` over the template tree (18 hits) and the deployed tree, read line by line, finds no unscoped arm/keep statement beyond K:28, F:119, F:122, A:117-118, A:127-129, A:130-135, A:396, W:244, M:26 (plus the new M subsection); A:103 and A:201 only name the cron. REQ-MJ-010 (spec.md:L89), plan M2 steps 1-5 and AC-MJ-008 name that set. But two of the ten have no verifying check: D17. |
| D4 predicate narrowing | closed | spec.md:L75 "same files or shared state"; AC-MJ-001 token list (acceptance.md:L23) carries `same files or shared state`; design.md:L101 candidate wording; E-001 re-run: `0`/`0` exit 1. |
| D5 traceability | closed | python port of the Group 4 verb (the awk form was refused, G3): `COLLECTED: 17 REQ definitions`, no UNCOVERED, no ORPHAN, exit 0. The REQ-MJ-006 inclusion clause is pinned by E-005d sub-bullet 1 and the REQ-MJ-009 no-polling clause by E-007 pinned line 4 (acceptance.md:L367, L385). |
| D6 shallow mutants | closed | AC-MJ-004 whole-line pins: E-004c, E-004e `:0`/`:0` exit 1, E-004d `:1`/`:1`; AC-MJ-008 per-site fragments: E-008b eight `:0` exit 1 and the E-008a awk printed fourteen lines (K:28, F:119, F:122, A:117, A:118, A:396, W:244 on both trees); AC-MJ-009 pinned line: `0 1 0` x2 (in -v form). Lexical limits stated (R-8). |
| D7 REQ-MJ-007 additions | closed | spec.md:L83 lists limits (a)-(g) and labels the assignment-input sentence a consequence; d314 body (board.txt:L23) contains each limit; "no other actor" and "nothing else" are gone; E-006 pinned lines 1-8 (acceptance.md:L376) match. |
| D8 leader-duty enumeration | closed | REQ-MJ-017 (d) (spec.md:L102), AC-MJ-018 (e); E-018e `:0`/`:0` exit 1; the table exists at factory-dispatch-mechanics.md:L169-L182 and still lacks the row, which is the stated red. |
| D9 AC-MJ-013 wrong-reason red | closed | portable ERE; probe file /private/tmp/lxq/aud/leak_in.txt: excluded t133, t224, t696, t1339, t1393 not matched; t134, t1394, t1338, t1340, t1392, t225, t223, t697, t695, t1500, t1894, a672, 0c51, 02f4, a SPEC id, a d-2026 id, a date and a 16-hex run matched; E-013a eight `:0` exit 1 on the real tree. |
| D10 Q7 row | closed as a row; contradiction in spec.md remains | decision-index.md:L77-L84 adds Q7 (FOUNDER, product-level, empty verdict, no recommendation). Whether the SPEC states honestly what it gates: see D19. |
| D11 Q6 label | closed | decision-index.md:L67-L75: FOUNDER, implementation-level, `Default:` Tier L, stamp `DEFAULT-APPLIED 2026-10-19T03:35:14Z lane-31 manager-spec (plan author)`; the anchor text no longer claims policy coverage. See stamp judgment below. |
| D12 Q2 and Q5 consequences | closed | decision-index.md:L26 and L65 state the consequences; spec.md R-7 and R-2 (spec.md:L144, L149). |
| D13 absolute sentence plus exception | closed (accepted as disclosed) | spec.md R-6 (spec.md:L148) records that a later card folds the exception into the Go doctrine tests. |
| D14 portability | closed for the grep/PCRE part; new related defect D18 | `grep -P` and `\b` are gone from the cells (acceptance.md:L10 states the tool convention); E-012b exits 1 on OVER (re-run: `OVER 25803`, exit 1). The awk cells are a different portability problem: D18. |
| D15 housekeeping | closed | no `/private/tmp` citation in spec.md (grep count 0); no `card:` key (grep count 0); 02f4 quoted verbatim (python check: body is a substring of spec.md); §1 lists the four scope items (spec.md:L30); REQ-MJ-004 is shall-form. |

## DEFAULT-APPLIED stamps, judged against manager-spec.md:L121 and the sibling precedents

The published rule allows `DEFAULT-APPLIED <UTC> <runner+role>` only on an implementation-level FOUNDER row that carries `Default:` chosen by the ranked rule and an empty verdict at close. The sibling decision indexes under the t1721 worktree (SPEC-AUDIT-CEILING-001, SPEC-ASIDE-BROWSER-001) predate the stamp and leave their FOUNDER rows empty or operator-answered; they are no precedent for or against the stamp, and I found no sibling that stamped a product-level row.

| Row | Class / stamp | Allowed for its class? | Silently decides an operator matter? |
|---|---|---|---|
| Q1 | implementation-level, stamped | yes (rule 1, preserves behavior); d314 additionally confirms ("manager-todo untouched", board.txt:L23) | no |
| Q2 | implementation-level, stamped | formally yes; the class note (decision-index.md:L25) separates the scope question (implementation-level) from the mode question (not decided) | no: the 0c51 consequence is disclosed in the row and R-7; the follow-up card is only "proposed" in a report outside the tree |
| Q3 | product-level, met by operator record | yes: a real operator record, quoted verbatim, verified against `moai decision read` (board.txt:L22-L23) | no |
| Q4 | implementation-level, stamped | yes; d314 moves the notice to card t1695 | no |
| Q5 | implementation-level, stamped | formally yes | no, but it leaves "in flight" undefined; disclosed (R-2) |
| Q6 | implementation-level, stamped, re-labelled from POLICY-COVERED | yes: the label no longer claims a covering setting, and the row states the fact that bounds the options (17 REQ and 19 AC exceed the Tier M ceilings of 16 and 16, spec-workflow.md ceilings table). The stated rule reason "preserves current behavior" is thin for a Tier choice (circular: the tier is whatever was authored), but the ceilings make Tier L the only reading without cutting content. The stamp time 03:35:14Z matches the file mtime of decision-index.md (12:35 local) and follows the iteration 0 commit (12:12 local), so it is not backdated. | no: the planner default is recorded as a planner default, not an operator verdict (G-1 states the Tier question was not asked) |
| Q7 | product-level, open, empty verdict | correct class and correctly unstamped | the row is honest; spec.md:L139 is not (D19) |

## RED-now re-execution (MP-8) and mutant tests

Cells re-run from the worktree root on HEAD 4834700468 (documents identical to the pinned trees per acceptance.md:L7: the correction commit changed only the seven SPEC artifacts, `git diff --stat b7b44433c3 4834700468`).

| Cell | Observed (stdout, exit) | Matches cell |
|---|---|---|
| E-001 | `0` `0`, exit 1 | yes |
| E-002a | refused by the guard as written; -v form printed `gtd.md:52 ...`, exit 1 | refused (D18) |
| E-002d | `:0` x2, exit 1 | yes |
| E-003 | `:0` x2, exit 1 | yes |
| E-004a / b / c / d / e | `:0` exit 1 / `:1` exit 0 / `:0` exit 1 / `:1` exit 0 / `:0` exit 1 | yes |
| E-005a / b / c / d | `:0` exit 1 / `:1` exit 0 / `:0` exit 1 / `:0` exit 1 | yes |
| E-006 | refused as written; -v form `0 0 0 1` x2, exit 1 | refused (D18) |
| E-007 | `0 0 0` x2, exit 1 | yes |
| E-008a / b / c / d | fourteen lines exit 1 / eight `:0` exit 1 / `:1` x4 exit 0 / `:0` x2 exit 1 | yes |
| E-009 | refused as written; -v form `0 1 0` x2, exit 1 | refused (D18) |
| E-010a / b | empty exit 1 / `:1` x2 exit 0 | yes |
| E-011 x7, E-011f | seven exit 0; fork pair `differ: char 23127, line 222`, exit 1 | yes |
| E-012a / b | `OK 9999` exit 0 / `OVER 25803` exit 1 | yes |
| E-013a / c | eight `:0` exit 1 / `2` `2` `1` exit 0 | yes |
| E-016 | empty, exit 1 | yes |
| E-017 | ` 1 file changed, 1 insertion(+), 1 deletion(-)` | yes (exit not capturable, G2) |
| E-018a / b / c / d / e | `:0` x2 exit 1 / `:1` x4 exit 0 / `:0` x4 exit 1 / `:0` x2 exit 1 / `:0` x2 exit 1 | yes |
| E-019a / c | `:1` x2, exit 0 | yes |

Stub arithmetic (design.md:L56-L64): old tail `stays armed to the completion report.` (37 units) to `is kept while the lane executes the card.` (41) is +4; `The one reconciliation is named, not excepted:` to `A reconciliation is named, not excepted:` is -6 (python `len` difference); net -2, 9,999 to 9,997 on the planned wording. Today `OK 9999`. The author's 9,997 is a planned-wording figure, not a measured one on the real tree (G-15); the arithmetic is verified, the file is not yet edited.

## Defects Found

D16. BLOCKING. A stem-less widening placed in or beside the Routing bullet passes every release-blocking criterion — acceptance.md:L367 (E-005d), L376 (E-006), design.md:L117 (D-10 row) — AC-MJ-005 (d) counts that the three pinned Routing lines are present (`grep -c -x -F` counts matching lines; an added fourth line is not counted) and AC-MJ-006 stem-scans only for `promot|queued|parallelizable`. Scratch mutants in /private/tmp/lxq/aud/m2/ (results in /private/tmp/lxq/aud/ac6_results.txt and ac5_results.txt): an extra Routing sub-bullet "A held card also counts as a free lane candidate while the operator is away." and "Held cards are assigned to a free live lane too." printed E-006 `8 8 1 1` and E-005d `3`, (a) `1`, (b) `0`, (c) `1`; the same for a stem-less sentence added to the card-review section and one before the subsection heading. AC-MJ-016 and AC-MJ-018 key on verbatim home tokens and whole lines, so none of them moves. The independent codex run reproduced the sub-bullet escape (finding P1: "The leader may take any backlog item ... even when the operator has retained dispatch" kept E-005d at `3` and E-006 at `8 8 1 1`, exit 0). design.md:L117 says such a widening is "caught by AC-MJ-005": true only for an edit of the three pinned lines, false for an added line; AC-MJ-006's own residual (acceptance.md:L89) covers "elsewhere in D" but not the Routing block it points to AC-MJ-005 for. verification-completeness.md §2: a writable mutant means the criterion is too shallow to adopt. Class: blocking. Required fix: bound the Routing block (from the parent line to the next blank line or heading) and require exactly three lines, or count all non-empty lines between the Routing parent and the next `###`/`##` heading; add the extra-sub-bullet mutant as a negative control observed red; correct design.md D-10 and the AC-MJ-006 "Evades, stated" sentence to say what is and is not caught.

D17. SHOULD-FIX. Two of the ten cron sites named by REQ-MJ-010 have no verifying check — spec.md:L89, research.md:L221 and L225 (rows 5 and 9), acceptance.md:L91-L101 (AC-MJ-007), L103-L116 (AC-MJ-008) — row 5 (A §5.1 fire-and-renew sentence, auto-semantics.md:L127-L129) is said to be "scoped by the count of `when the lane stops executing the card`", but that literal pins row 6 (the delete sentence, L134) only; row 9 (the renew paragraph, factory-dispatch-mechanics.md:L26) is pointed at the new subsection but AC-MJ-007 scans the `### Lease renewal` subsection only. E-008a printed no line for A:127-129 or M:26 today (they carry no keep-word), so they cannot go red or green. A run that leaves either sentence in its present unscoped wording keeps AC-MJ-007 at `4 4 1` and AC-MJ-008 (a)-(d) at their green values. The independent codex run reproduced both (finding P2). AC-MJ-008's title claims "site by site". Class: blocking. Required fix: pin the new wording of both sentences as whole lines (a whole-line `grep -c -x -F` per tree, RED `:0` today) and add the unedited-sentence mutant as a negative control; or remove the two sites from REQ-MJ-010's list and say why.

D18. SHOULD-FIX. Three release-blocking evidence commands are refused by the worktree guard as written — acceptance.md:L248 (E-002a), L376 (E-006), L450 (E-009) — E-006 and E-009 were refused with the message that an awk program has a `|` after print/printf ("a | in a regex ... at such a spot is refused too: pass the pattern with -v re='a|b'"); E-002a was refused as an awk program "that can execute commands or that it cannot read in a plain command". Every lane session runs under the same guard, and the SPEC's own Gap G-4 (progress.md:L91) records refusals of this kind. Re-run in the form the guard message prescribes, the REDs reproduce: `awk -v re='promot|queued|parallelizable' ... !f&&(tolower($0) ~ re){k++}` printed `0 0 0 1` / `0 0 0 1`, exit 1; `awk -v re='every card|each card' ...` printed `0 1 0` / `0 1 0`, exit 1; `awk -v re='moai (todo|gtd) add "' ...` printed the gtd.md:52 line, exit 1. Under the §2.1 execution discipline a refused cell is not a pass and the criterion takes the demotion, which would strip three of the SPEC's central gates of release-blocking standing. acceptance.md:L10 says the cells "were run through `sh`"; the lane under audit cannot do that. Class: blocking. Required fix: move every `|` alternation out of the awk program text into `-v re=...` for E-002a, E-006, E-009 (and E-007, E-008a, E-001 are fine as run), re-run each cell and its mutants in that form, and add one line to §0 Conventions naming the guard's awk restriction.

D19. SHOULD-FIX. The SPEC states two incompatible things about what the open row Q7 gates — spec.md:L139, plan.md:L71, acceptance.md:L657 versus decision-index.md:L84 and L96 and progress.md:L13 — spec.md:L139 says Q7 is "Open, empty verdict; goes to the operator; nothing is gated on it"; plan.md:L71 says "Q7 (open) gates no text"; acceptance.md:L657 says "No requirement or criterion is gated on a decision"; decision-index.md:L84 and L96 and progress.md:L13 say the empty product-level FOUNDER verdict "blocks the autonomous Kickoff". Both can be true (no text waits; the Kickoff form does), but a reader of spec.md alone is told the run may start, and REQ-MJ-007 (a) ("while the operator has delegated dispatch") is exactly the phrase Q7 asks about. The user-facing question was whether the SPEC states honestly which requirements it gates: the requirement texts are honest (they carry the ruling's words and R-5), the Kickoff effect is stated in only one of the documents. The independent codex run reproduced it (finding P2). Class: blocking (consistency of a stated gate). Required fix: say once, and the same way in spec.md §6, plan.md §C and acceptance.md §4: "no requirement or criterion text waits on Q7; REQ-MJ-007 (a) is text-final but its operational trigger is open; the autonomous Kickoff form is blocked until the operator answers Q7 or states that the ruling's phrase is the whole condition".

D20. MINOR. Two requirement halves carry no token — spec.md:L76-L77 (REQ-MJ-003 "name the conflicting card"; REQ-MJ-004 "shall direct that it be reissued") versus acceptance.md:L23 (the six AC-MJ-001 tokens: `expected file set`, `in-flight card`, `same files or shared state`, `slot-gated surface`, `conflict check`, `--classification-file`) — a `### Mode judgment` subsection that omits the reissue sentence and the named conflicting card still prints `6`. Class: optional. Required fix: add `reissue` (and a token for the named conflicting card) to the token list, with a mutant observed red.

D21. MINOR. The commit-shape check is a listing, not a verdict — acceptance.md:L148 — `git diff-tree --no-commit-id --name-only -r <sha>` exits 0 whatever it lists (run on HEAD: exit 0, seven paths); matching template paths to deployed twins is a human read. Class: optional. Required fix: none now (the criterion is declared run-phase and undecidable); in the run phase print the unmatched side in one decision line and exit 1 on a non-empty difference.

D22. MINOR. Stale statements in progress.md — progress.md:L14 (`spec_lint: NOT RUN` — `moai spec lint --strict` printed `0 error(s), 0 warning(s)`, exit 0 on spec.md v0.1.2 in this run), L53 ("the net of the planned K edits is +3, so a >= 2-unit offset is required" — contradicted by L82, design.md:L56-L64 and plan.md M2 step 2, net -2). The L53 text sits in the earlier correction-round record, so it reads as history, but a reader can take it for current. Class: optional. Required fix: mark both as superseded.

Stagnation / regression: none; the iteration 0 defects were each worked, none repeats unchanged. Not a regression of the aggregate.

## Recommendation (delta scope = fix_scope; the confirming re-audit re-runs the ordering check in full)

1. D16: bound the Routing block (exactly three lines) and add the extra-sub-bullet mutant as an observed-red negative control; correct design.md D-10.
2. D17: pin the new wording of A §5.1 fire-and-renew and of M:26 with whole-line checks that are red today, with unedited-sentence mutants.
3. D18: rewrite E-002a, E-006, E-009 with `-v re=` and re-observe RED, green path and mutants in that form.
4. D19: state the Q7 gate once, identically, in spec.md §6, plan.md §C and the acceptance.md DoD.
5. Optional D20-D22 are surfaced for the orchestrator's discretion and do not by themselves justify a FAIL.

## Five-section evidence record

**Claim.** The correction round closed the nine blocking defects of the iteration 0 report (D1, D4-D9 closed; D2 and D3 closed with residuals now filed as D16 and D17), but the SPEC still fails: the required codex gate failed, four blocking defects stand, the aggregate is 0.79 against 0.85.

**Evidence.** Commands run from /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1894 (exit codes as observed):
- `git rev-parse HEAD` printed 4834700468ed2311e3efc4e12bb920db03db54e2; `git status --short` printed nothing.
- 34 of 42 evidence-ledger cells re-run and reproduced as in the table (E-011 counted as one cell of seven `cmp` calls); the three refused cells (E-002a, E-006, E-009) re-run in `awk -v re=` form with the stdout shown in D18.
- AC-MJ-006 mutants: `awk -v re=... ` over 22 files, output saved to /private/tmp/lxq/aud/ac6_results.txt. Tail of that file: `N9_routing_parent_stemless.md 8 8 1 1`, `NEW_routing_extra_subbullet_stemless.md 8 8 1 1`, `NEW_routing_extra_subbullet_stemless2.md 8 8 1 1`, `NEW_review_section_stemless.md 8 8 1 1`, `NEW_before_head_stemless.md 8 8 1 1`; the 16 author mutants: `held_removed 7 8 1 1`, `nocond 7 8 1 1`, `serial_promoted 7 8 1 1`, `trailing_space 7 8 1 1`, `dup_line 8 9 1 1`, `h4_widening 8 10 1 1`, `h3_widening 8 10 1 1`, `routing_parent_stem 8 8 1 2`, `routing_sub_stem 8 8 1 2`, `appended_line1 7 8 1 1`, `extra_bullet 8 9 1 1`, `extra_h2_before_review 8 8 0 2`, `heading_renamed 0 0 0 6`, `heading_deleted 0 0 0 5`, `endmarker_renamed 8 8 0 1`, `consequence_relabelled 7 8 1 1`; good copy `8 8 1 1`.
- AC-MJ-005 (d) on the same files (saved to /private/tmp/lxq/aud/ac5_results.txt): good `3`; N9 parent-line edit `2`; the four stem-less additions `3` each.
- `moai decision read > /private/tmp/lxq/aud/board.txt`, exit 0, 1913 lines; python quote check: d314 (board.txt:L22-L23), 0c51 (L39-L40), a672 (L49-L50), 02f4 (L45-L46) bodies are substrings of spec.md, d314 also of decision-index.md. No board record answers Q7.
- `cmp` x8 pairs: seven exit 0, factory-dispatch.md exit 1 (char 23127, line 222); `git diff --no-index --shortstat` printed ` 1 file changed, 1 insertion(+), 1 deletion(-)`.
- `moai spec lint --strict .moai/specs/SPEC-CARD-MODE-JUDGMENT-001/spec.md`: `0 error(s), 0 warning(s)` plus one INFO, exit 0.
- Trace verb (python port): `COLLECTED: 17 REQ definitions`, no UNCOVERED, no ORPHAN.
- Template tree search for cron arm/keep statements: 18 hits (saved to /private/tmp/lxq/aud/cron_tmpl.txt); deployed tree and docs: 35 hits (cron_depl.txt); both read line by line.
- 20 named Go test functions exist (`git grep` count 20); ledger rows ALB-0254, -0256, -0257, -0265, -0268, -0276, -0304 exist (binding_ledger.json lines 4821, 4895, 4932, 5196, 5275, 5534, 6430); manager-todo.md has no `classification-file` or `moai todo add` (grep exit 1).
- `moai verify audit-plan --project-root <toplevel>`: config_status ok, model codex, codex required, claude and glm off, enforced_required [codex]. `mcp__moai__audit_multi` called once (target baseBranch, delta_base cc109a75b6bf9c06ce86d811a8f0068551f02484, project_root = this tree): overall_verdict fail; codex (required) fail with three findings (P1 Routing sub-bullet escape, P2 cron sites fire-and-renew and M:26 unchecked, P2 Q7 gating contradiction), all three reproduced independently here (D16, D17, D19); receipt rcpt-2adbd9564960a6f7da294042; server build commit 7705b365e2, an ancestor of HEAD 4834700468. After the call, `moai verify audit-plan ... --result-file .moai/state/audit-plan-result.json`: `convergence_check` ok true, unmet [] (the required backend answered; its answer was a fail, a reviewed result and not an unmet gate).

**Baseline-attribution.** Every observation was measured in this run against HEAD 4834700468ed2311e3efc4e12bb920db03db54e2. CLI results were judged by the installed build moai_cp/20261007_195927 (older than HEAD: `moai version` printed that build). Mutant files were built in this run under /private/tmp/lxq/aud/ from the pinned lines of acceptance.md E-006 and E-005d; they are scratch, cited only through the outputs printed above.

**Gaps (not observed).**
- G1: the plan artifact hash was not recomputed (this agent has no audit_cache tool); the value written is the lane's runtime value for the SPEC directory at HEAD (`git status --short` printed nothing, so the directory is unchanged since the commit).
- G2: git commands with a trailing exit probe are refused by the guard; the exit code of E-017 and of the E-002b/E-002c-style git cells is stated from git's contract, not probed.
- G3: the Group 4 trace verb and the Group 6 CN-4 verb (awk with `|` after print) were refused by the guard; a python port of the trace verb and a grep plus manual read of the ordering candidates were used. A mechanical CONFLICT line from the original CN-4 awk cannot be excluded, though the plan binds no criterion through an `Exit:` line (grep count 0).
- G4: the awk cells E-002a, E-006, E-009 were refused as written (D18); their reproduction is in a different form.
- G5: no Go test, build or make target was run; what a guard test pins comes from reading todo_auto_doc_test.go L170-L232 and the ledger rows named above. TestAutoRankMirrorParity compares each live document with its template mirror (not K with F); the clause `The one reconciliation is named` appears in no Go test, only in the ledger row ALB-0257 `before_text`/`after_text`.
- G6: the planned post-run text does not exist; green paths and mutants in the SPEC are from the author's scratch simulation (G-15), which this audit did not reproduce; the green path of AC-MJ-006 was reproduced only on my own reconstruction of the nine/eight pinned lines.
- G7: sibling-SPEC precedents for DEFAULT-APPLIED were read only for the decision indexes of SPEC-AUDIT-CEILING-001 and SPEC-ASIDE-BROWSER-001 (grep of headings), not for the t1722 SPEC.
- G8: research.md was read at §10-§12 and the headings; its other sections were not re-run.
- G9: the five known-failing-input control cells E-002b, E-002c, E-010c, E-013b and E-019b were not re-run (E-010c and E-013b read an untracked report under .moai/reports/t1894/; E-002b and E-002c are git cells whose exit the guard does not let a probe capture). Their absence does not change a RED-now judgment; it leaves the author's control observations unconfirmed here.

**Residual-risk.** (1) The criteria are lexical (R-8): after D16-D18 are fixed a widening written as a contradiction inside a line that also carries a pinned token, or as a paraphrase in another file, still evades; the sync audit's reading is the only control. (2) The mutants found here are those I could write; other lines of the same kind (AC-MJ-001, -002, -003, -016) were not mutated beyond the author's list. (3) The audit_multi server and the installed moai predate HEAD; the codex leg's scratch mutants ran in its own environment. (4) Three of the central gates (AC-MJ-002 c1, -006, -009) would be demoted by the §2.1 discipline until D18 is fixed.

## Operational Notes (unverified)

- Measure the Routing-block bound after the fix — command: re-run E-005d and E-006 (-v form) on /private/tmp/lxq/aud/m2/NEW_routing_extra_subbullet_stemless.md and NEW_routing_extra_subbullet_stemless2.md; expected: at least one count differs from the good copy's `3` / `8 8 1 1`. Status: measured before the fix (both unchanged), unmeasured after.
- Check how the run-phase lane will execute the awk cells — command: run each cell once under the lane's guard before M1. Status: inferred (the guard refused them here; the lane runs in a worktree session as well).
- Q7 can be closed by one operator sentence: either a named mechanism or "the ruling's phrase is the whole condition" — then the autonomous Kickoff form is open. Status: assumption (decision-gate rule read from manager-spec.md:L121; not executed).

AUDIT-VERDICT: FAIL spec=SPEC-CARD-MODE-JUDGMENT-001 receipts=rcpt-2adbd9564960a6f7da294042
