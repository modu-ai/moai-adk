# plan.md — SPEC-CARD-MODE-JUDGMENT-001

Documents-only change, Template-First, Tier L. Milestones are ordered by decision reversibility — the wording most likely to be disputed first, the mechanical confirmation last — and by priority, never by duration. Requirements: `spec.md` §4. Criteria: `acceptance.md`. Constraints and rationale: `design.md`. Exact wording of every added line: `closed-world-manifest.txt`. Measurements and line numbers: `research.md`.

## A. Technical approach

1. **Template-First is an authoring discipline, and the commit graph cannot witness it.** Edit the template side, then bring the deployed copy (`.claude/...`) to the same bytes, inside one milestone. Which of the two was edited first cannot be read from a commit that carries both (`verification-claim-integrity.md` §2.3), so no criterion claims it. What the graph does show, and what AC-MJ-011 checks, is that every milestone commit carries a changed template document **together with** its deployed twin. A deployed-only edit still fails AC-MJ-001..009, which read both sides.
2. **One home per rule.** M1 writes C's `### Mode judgment`; M2 writes M's `### Lease renewal`; M3 writes D's Routing sub-bullets and `### Delegated-dispatch promotion`. Other sites receive a marker or a pointer, not the rule (REQ-MJ-012, `design.md` §D-2).
3. **Do not edit a ledger-pinned line that has no reason to change** (`design.md` §D-3): F:30, K:32/36/68 stay verbatim. Lines whose scope must change (F:119, F:122, K:28, A §5.1 and §14, W:244, M:26, the K:20/F:44 clause) are rewritten, and each ledger row whose `after_text` is rewritten (ALB-0276 for F:119; ALB-0257 for K:20) is updated in the same commit. The card-id literals `t1393`/`t1339` (A), `t133`/`t224` (D), `t696` (G) stay: `TestCardIDBaselineHasNoStaleEntries` fails when one leaves its file.
4. **Worded fresh.** New text copies no ledger-covered line of ≥ 40 units into a `paths:`-scoped file (AC-ALB-021(2)), and carries no card id, board id, SPEC id, date, or long hex run (AC-MJ-013). Pinned whole lines (AC-MJ-004, -005, -006, -007, -009, -018) are copied from `acceptance.md`, not re-worded, and each sits on one physical line. **Every line the run adds or removes is a row of `closed-world-manifest.txt`** (AC-MJ-020; the manifest also carries the free-wording lines of C, K, F, A, G, W and M, which were generated from the planned text): copy the rows, with their line breaks, and add nothing. The run phase does not edit the manifest; a wording change is a blocker report for manager-spec.
5. **Class vs mode.** Say once, in C, that class (A/B/C) names the stages and mode names whether a card may run beside others; the two axes are independent.

## B. Milestones

### M1 — Mode judgment and explicit issuance (Priority High; highest change-likelihood: user-facing procedure)

Files (template and deployed in one commit): C, G, F, plus the `moai` entry of `internal/template/catalog.yaml` (G lives in that skill directory: `go run ./internal/template/scripts/gen-catalog-hashes.go --entry moai`).

1. C: add `### Mode judgment — serial or parallelizable` after the parallelism paragraph (line 34), before `## Card size ceiling`. Content: the two rows of manifest section C (the heading and the one paragraph), which carry the file-set predicate with "the same files or shared state", the slot-gated-surface trigger, the reason format, the explicit `--classification-file` (decider `human`), the reissue sentence and the Class C `--spec` sentence containing `cannot be added after assign` (`design.md` §D-9 gives the rationale, not the wording).
2. G: rewrite line 52 so the issuance instruction sits on one physical line carrying `--classification-file <path|->` and citing `factory-dispatch-cards.md` § Mode judgment; rewrite the `spec_id` bullet (lines 266-271) and the pick step (lines 328-329) to the pinned lines of AC-MJ-004 (c) and (e). Leave line 56's factual sentence and the `--force` row alone (REQ-MJ-015).
3. F: add one new [HARD] line directly after line 30 (the explicit-classification obligation, citing C). Leave line 30 verbatim.
4. Check: AC-MJ-001, -002, -003, -004 green; AC-MJ-010, -011, -013 and -017 still green; AC-MJ-020 (a) shows `0` in its first and third column for every file and `0 0 0 0` for C (the rows of G and F that belong to M2 and M3 are still missing, so their second and fourth columns are not yet `0`); the `pairs` command on the M1 commit prints `pairs 3 unpaired 0` (C, G, F). `TestTodoSkillDocumentsClassification` neutrality: no SPEC id / date / hex in G.

Open items touching M1: none blocking. Q1 default (agents untouched) and Q2 default (`doc-generation.md` untouched) apply; Q5 default keeps the board's words verbatim.

### M2 — Executing-card-only lease renewal, scoped standing cron, and the stub's two edits (Priority High; most pinned)

Files (template and deployed in one commit): M, K, F, A, W, plus `internal/template/testdata/binding_ledger.json` rows ALB-0276 and ALB-0257 and the `moai-lane-watchdog` entry of `catalog.yaml` (`--entry moai-lane-watchdog`).

1. M: add `### Lease renewal — only the card being executed` with the four pinned lines of AC-MJ-007 under `## Factory Mode mechanics`, after the renew paragraph and before `## Isolation`; rewrite line 26 to start "While it executes the card," and point at it (the `moai factory renew` verb and `factory_lease.duration` text stay). *Executing* is defined once, in the first pinned line.
2. K: **both stub edits in this one commit** — K:28 tail becomes `is kept while the lane executes the card.` (+4 units) and the K:20 clause becomes `A reconciliation is named, not excepted:` (−6 units); the stub nets −2 (9,999 → 9,997, measured on the simulated text, `research.md` §8) and is never over budget at any commit. The pinned renew-interval wording stays on K:28. The clause edit is REQ-MJ-017 (b) and rides here for that reason.
3. F: line 119 (the ledger-pinned compressed line) takes `kept while the lane executes the card;`, line 122 takes `keeps it while the lane executes the card:`, line 44 takes the same clause edit as K:20; update ledger rows ALB-0276 (`after_text` = the new line 119) and ALB-0257 (`after_text` = the new K:20 text) in this commit. Keep the `### Lane waits are explicit, and stalls are watched` tokens (`CronCreate`, `recurring`, `disk evidence`).
4. A: rewrite on single physical lines the §5.1 intro (`kept while the lane executes the card.`), the fire-and-renew sentence (renews the lease of the card the lane is executing), the Keep bullet (takeover also at resumed execution; `Delete it … when the lane stops executing the card, and at the latest when the completion report is sent`), and the §14 carrier sentence (`is kept only while the lane executes the card`). Leave A:103 and A:201 (they only name the cron) and the one-shot wait recheck of §14. Keep the §5 tokens of `design.md` §D-3 and the lines that carry `t1393` / `t1339`.
5. W: the §5 constraint bullet takes `armed at card intake and kept while the lane executes the card` on one physical line; §0a gains the pinned scope line of AC-MJ-009, keeping the heading word `Renew`, `moai factory renew`, the closed `renew_result` set, "does not end the pass", and order (Renew before Measure). A pass that does not renew carries the previous `renew_result` forward.
6. Check: AC-MJ-007, -008 (a)-(f), -009, -012 green; AC-MJ-020 (a) first and third column `0` for every file; the `pairs` command on the M2 commit prints `pairs 5 unpaired 0` (M, K, F, A, W); AC-MJ-018 (b) and (c) green; AC-MJ-011 and -013 still green. Ledger and catalog rows are in this commit, so `TestBindingLedgerIntegrity` and the catalog hash audit can be run on it.

### M3 — Free lane, delegated-dispatch promotion, and the pointers to it (Priority Medium; nothing is held: the operator ruling d314 settled Q3)

Files (template and deployed in one commit): D, F, G, M, plus the `moai` entry of `catalog.yaml` (G again).

1. D Routing bullet (line 133): replace the parent line with the pinned parent line of AC-MJ-005 (the "free means the lane's previous card reached `done`" clause removed) and add the two pinned sub-bullets (executing-card definition with the merge-ready / held / serial-waiting inclusion; free versus live with `not a free lane`). Line 35 (the diagram) keeps the term unchanged.
2. D: add `### Delegated-dispatch promotion` after the last bullet of `## Factory in-lane 3-stage` and before `## The card-review stage`, consisting of exactly the eight lines pinned in `acceptance.md` E-006 — copy them, do not re-word: AC-MJ-006 compares whole lines. Re-measure D's size afterwards (27,611 units today; the per-file figure that applies is 40,000).
3. Pointers (REQ-MJ-017, AC-MJ-018, AC-MJ-019):
   a. F: insert the exact pointer line of E-018a directly after F:40 as its own paragraph. Leave F:38 and F:40 verbatim.
   b. G: add the exact pointer paragraph of E-018d as one physical line after the [HARD] pick rule (the paragraph at `gtd.md:307-310` whose last words are `invent work.`). Leave that paragraph verbatim.
   c. M: add the exact enumeration row of E-018e after the `Serial-slot policy` row of `## The leader's remaining role`.
   d. New text avoids the strings `Promotion is the operator's act, always.` and `The leader never picks for the operator`, and carries no SPEC id, REQ token, date, board or card id, or hex run (the neutrality tests).
4. Check: AC-MJ-005, -006, -016, -017, -018, -019 green; AC-MJ-020 (a) `0 0 0 0` on every file and (b) `unexpected-paths 0` (the whole manifest is applied after M3); the `pairs` command on the M3 commit prints `pairs 4 unpaired 0` (D, F, G, M); AC-MJ-011 and -013 still green.

### M4 — Confirmation pass (Priority Medium; expected to produce no commit)

1. Re-run AC-MJ-013 (`E-013a`) now that milestone diffs exist; a hit names a leaking file.
2. Confirm that every rewritten covered line has its ledger row updated (M2: ALB-0276, ALB-0257) and that each skill directory changed by a milestone has its catalog hash regenerated in that milestone's commit (`moai` in M1 and M3, `moai-lane-watchdog` in M2). A missed row is repaired in a corrective commit named as such; it is not a normal step.
3. Agent files are not edited by default, so `make agents-emit` and the `.codex` TOMLs are untouched; if Q1 flips to include them, run `make agents-emit` and re-run the emit check.

### M5 — Verification and evidence (Priority High; the gate)

Read-only checks first, in one parallel batch (AC-MJ-001..013, -016, -017, -018, -019, -020 commands from `acceptance.md`), outputs redirected to files under the card's scratch directory with the exit code and a bounded tail surfaced; the lines that decide each claim go into `.moai/reports/t1894/verdict.md`.

Then the **scoped Go checks** (AC-MJ-014). Never `go test ./...`:

- `internal/template`, one invocation: `go test ./internal/template -count=1 -run '^(TestBindingLedgerIntegrity|TestRoleCoreStubGate|TestDeployedAlwaysLoadedCharBudget|TestRuleTemplateMirrorDrift|TestDeclaredRuleMirrorForks|TestReviewOwnership_CardReviewDoctrine|TestFactoryDispatchBodyCarriesRenewIntervalLinkage|TestLaneRecheckDoctrine|TestLaneWatchdogRenewsFirst|TestLaneWatchdogRecordsRefusalAndMeasures|TestWorkflowRulePathsPinned|TestTemplateNoInternalContentLeak|TestCardIDBaselineHasNoStaleEntries|TestCatalogHashCoversSkillSubfiles)$'`
- `internal/cli`, one invocation: `go test ./internal/cli -count=1 -run '^(TestAutoPickMirrorParity|TestAutoPickDocDoctrine|TestTodoSkillDocumentsClassification|TestAutoRankDoctrineAmendment|TestAutoRankMirrorParity|TestAutoRankMarkerDisclosure)$'`

Record, per verdict, the empty-sweep check (each `-run` must report a non-zero test count — `[no tests to run]` is a failure, `verification-completeness.md` §1.1) and the judging build (build from this tree and invoke it by path, or state that the installed build is not a strict ancestor — `verification-claim-integrity.md` §2.2).

**Resource-slot and admission note (run phase; not an action now).** Compiling the `internal/cli` test binary is the heavy step (the full `internal/cli` suite is on the repository's heavy-run list). Before the Go invocations the lane takes the resource lease (`moai slot acquire --resource <name> --max-duration <bound>` … `moai slot release --resource <name>`; `resource-slot-lease.md`), follows the leader's resource-entry rules (load and memory admission, staggered start, one Go invocation per lane at a time), and runs each verification as a single `unset <VARS> && <command>` call when it needs a scrubbed environment (`AGENTS.md` §4). `make build` (which also runs the agent- and command-emit drift checks) is not required for a documents-only change and is taken only if the leader asks for it, in the same lease.

## C. Order and dependencies

M1 → M2 → M3 → M4 → M5, with M5 running after M4. No milestone waits on a decision: Q3 is met by the operator ruling d314, and Q1, Q2, Q4, Q5, Q6 carry their stamps. No requirement, criterion or milestone of this SPEC waits on Q7; REQ-MJ-007 (a) is text-final because it carries the ruling's words and no mechanism. Per the decision-gate rule the empty FOUNDER row Q7 goes to the operator and blocks the autonomous form of the plan→run Kickoff until the operator answers it or states that the ruling's phrase is the whole condition. Q7 was answered on 2026-10-23 (relayed by the leader) with exactly that statement, so it no longer blocks the autonomous Kickoff form. The two stub edits both land in M2, so the stub is within budget at every commit; M3 does not touch K.

## D. Risks

| Risk | Where it shows | Mitigation |
|---|---|---|
| Core stub over 10,000 units | AC-MJ-012; `TestRoleCoreStubGate` | the `is kept while the lane executes the card.` tail (+4) and the K:20 clause (−6) in one commit, net −2; never the pointer form in K |
| A pinned old sentence folded into the exception | AC-MJ-019; `TestAutoRankDoctrineAmendment`, `TestAutoPickDocDoctrine` | edit F:38, F:40 and the pinned clauses of F:44 / K:20 not at all beyond the one clause; the exception is the pointer line plus home D |
| A widening of the exception in the home subsection | AC-MJ-006 (whole-line equality, line count, end marker, stem scan) | copy the eight lines from `acceptance.md` E-006; amend the criterion with any wording change |
| A widening in the Routing bullet, or any other added line | AC-MJ-005 (d) pins its three lines (it cannot see an added fourth); AC-MJ-006 stem scan (it cannot see a stem-less line); AC-MJ-020 lists every allowed added and removed line of the eight documents and reports any other | copy the pinned lines and the manifest rows, add nothing |
| An unstated delegation treated as given | spec R-5, decision-index Q7 | the documents carry the ruling's words and no mechanism; the sync audit reads for invented ones |
| Ledger row left stale | `TestBindingLedgerIntegrity` "obligation lost" | update the row in the same commit as the line |
| Catalog hash left stale | catalog hash audit | regenerate the changed skill entry in the commit that changes the skill directory |
| A pinned token lost in a scoped rewrite | `TestLaneRecheckDoctrine`, renew-interval test, watchdog tests | `design.md` §D-3 token list checked line by line |
| A baselined card id lost in a reflow of A §5.1 | `TestCardIDBaselineHasNoStaleEntries` | keep the lines carrying `t1393` / `t1339` |
| Unscoped tail line in hard-wrapped A | AC-MJ-008(a) | reflow so the scoped sentence sits on one physical line |
| Template-only edit | AC-MJ-001..009 show `:0`/`0` on the deployed operand; `TestAutoPickMirrorParity` | both sides in the same commit |
| `factory-dispatch.md` fork absorbed or duplicated | AC-MJ-017; `TestAutoPickMirrorParity` | apply the same edits to both copies; never copy the worktree-sweep sentence into the template |
| Injected SessionStart lane notice disagrees with the scoped rule | spec R-1 | out of scope (Q4); the leader's lane notice (AC-MJ-015) names it; a follow-up card is proposed in the report |

## E. Commits (run phase)

One commit per milestone M1, M2, M3, each naming the card id in its message, explicit pathspecs only (the primary-checkout sweep prohibition applies in spirit inside the worktree too). **Each of those commits carries every changed template document together with its deployed twin** — never a template-only commit followed by a deployed-only one — so the template-vs-deployed mirror guards are never asked to pass over a half-applied state. `catalog.yaml` and `binding_ledger.json` ride in the commit that changes what they pin (M1: the `moai` hash; M2: ALB-0276, ALB-0257 and the `moai-lane-watchdog` hash; M3: the `moai` hash again). M4 and M5 produce no commit unless a corrective one is needed.

How each commit stays coherent without running the Go guards per commit (they run once, at M5, on the final tree): the pairs are paired; a ledger row and a catalog hash are updated with the line they describe; both stub edits share a commit; and the read-only cells of the milestone (AC-MJ-001..013 as listed in each milestone's check step) are run on the commit before the next milestone starts. A shared-tree guard failure in an intermediate commit is therefore a mistake in that commit, not an expected state. The commit-graph evidence for this shape is the run-phase obligation of AC-MJ-011 (the `pairs` command, one call per milestone commit with its expected pair count). The allowed lines are fixed before the first document commit: `closed-world-manifest.txt` is written and frozen by the plan phase, so the commit graph orders the list of allowed lines ahead of the lines themselves (`verification-claim-integrity.md` §2.3). The plan-phase commits carry only `.moai/specs/SPEC-CARD-MODE-JUDGMENT-001/*`.
