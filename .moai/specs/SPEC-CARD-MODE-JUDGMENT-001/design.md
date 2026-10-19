# design.md — SPEC-CARD-MODE-JUDGMENT-001

The HOW layer for a documents-only change. Everything here is a measured constraint or a candidate wording; the normative text is `spec.md`, the observable checks are `acceptance.md`. Site ids S1-S17 and the guard names are defined in `research.md` §4 and §6.

## D-1 Tier L, and why the count is honest

`spec-workflow.md` § SPEC Complexity Tier: Tier M is "5 - 15 files affected", Tier L is "> 15 files or constitutional". Files this change must touch, counted without discounting mirrors:

| Group | Count | Files |
|---|---|---|
| Template documents | 8 | C, K, F, D, M, A, G, W (the rule files and two skill files; `research.md` §4) |
| Deployed mirrors of those | 8 | the same eight under `.claude/` |
| Generated or fixture files that must move with them | 2 | `internal/template/catalog.yaml` (two stored sha256 values: `moai`, `moai-lane-watchdog`) and `internal/template/testdata/binding_ledger.json` (rows whose `after_text` is rewritten, at least ALB-0276) |
| **Total** | **18** | > 15 |

Counting only the eight source documents gives 8-10 files (Tier M). The mirrors and fixtures are files the author must edit and a guard test reads, so they count; Tier L follows from the table, and `design.md` + `research.md` are not ceremony here — the pinning map (§D-3), the budget arithmetic (§D-4) and the sweep definition (`research.md` §7) are decisions a reviewer needs to see. Recorded as decision-index row Q6 (FOUNDER, implementation-level, `DEFAULT-APPLIED` Tier L; the cited rule table is guidance, not an operator setting, so no policy coverage is claimed). The Tier judgment question that `spec-assembly.md` poses to the operator could not be asked from this subagent; that is Gap G-1 in `progress.md`. A fact from the same table bounds the options: 17 requirements and 19 criteria exceed the Tier M ceilings of 16 and 16.

LOC is small (well under 300 added lines of prose); the tier is driven by file count and by the pinning-test density, not by size. The REQ ceiling (25) and AC ceiling (25) are respected: 17 REQ, 19 AC (sub-criteria `a`/`b`/`c`/`d` of one AC are counted once). The correction round added REQ-MJ-017 and AC-MJ-018 / AC-MJ-019 and changed no file count: the same eight template documents, their eight deployed mirrors, `catalog.yaml` and `binding_ledger.json` (now with rows ALB-0276 and ALB-0257 rewritten).

## D-2 Single homes (REQ-MJ-012)

| Rule | Home | Anchor token (exact, grep-able) | Other sites carry |
|---|---|---|---|
| Mode predicate (same files or shared state as an in-flight card; slot-gated surface), reason format, Class C `--spec` at pick, "explicit classification, never the default" | C `### Mode judgment` | `slot-gated surface`, `cannot be added after assign` | a pointer by section name: G:52, F (new [HARD] line), G (the `spec_id` bullet and the pick step), nothing else |
| Free lane; free ≠ live session | D, Routing bullet of `## Factory in-lane 3-stage` (parent line plus two sub-bullets, all pinned) | `no card is actually executing`, `not a free lane` | K:68 and F:227 keep "free lane" unchanged; M cites D |
| Delegated-dispatch promotion (the case in which the leader promotes queued `parallelizable` cards without a per-card pick) and, labelled as a consequence, the assignment input | D `### Delegated-dispatch promotion`, after the last bullet of `## Factory in-lane 3-stage` | `while the operator has delegated dispatch`; the eight whole lines of `acceptance.md` E-006 | F: one pointer line after F:40 (`acceptance.md` E-018a); G: one pointer paragraph after the [HARD] pick rule (E-018d); M: one row in the leader-duty enumeration (E-018e); nothing else |
| Lease renewal (executing card only), expiry harmless, no polling cron for slot/trigger waits, leader notice, *executing* defined | M `### Lease renewal` (four pinned lines) | `only the card the lane is actually executing`, `expiry is harmless` | the scope fragments of `acceptance.md` E-008b in K, F (two lines), A §5.1 and §14, W (constraint bullet), plus the pinned scope line in W §0a |

AC-MJ-016 enforces the table with one `grep -l`: the six home tokens appear in exactly C, D, M. A paraphrase restated in a second file would evade a verbatim-token check; that residual is read by the audit, not claimed closed.

## D-3 Pinning-test map: what each guard permits

The change is not free-form editing. Each row is a constraint measured in `research.md` §6.

| Site | Pinned by | Constraint on the edit |
|---|---|---|
| S6 F:30 | ledger row ALB-0254 (`rule-body:`, `binding`) | do not edit the line; add the explicit-classification obligation as a NEW [HARD] line directly after it |
| K:32, K:36, K:68 | ledger rows ALB-0265, -0268, -0304 (`always:` core, `normative`) | do not edit; mode judgment is reached through C's `### Mode judgment`, and the existing pointer "Full table and rationale: `factory-dispatch-cards.md` § Card classes" already leads there |
| S7 F:119 | ledger row ALB-0276 (`role-core:`, `processing: rewrite`, `after_text` = the compressed line) | rewriting it requires updating that row's `after_text` in the same change (REQ-MJ-016); keep `moai-lane-watchdog`, `CronCreate`, `recheck`, `disk evidence` |
| S8 F:122, S10 K:28 | `TestFactoryDispatchBodyCarriesRenewIntervalLinkage` | the line keyed `A lane that cannot proceed records an explicit wait on disk` (F) / `A stuck lane records a disk wait` (K) must contain ``every renewal period of `workflow.factory_lease.renew_interval` `` on the same physical line and must not contain `every 20 minutes` or `20-min` |
| F `### Lane waits are explicit, and stalls are watched` | `TestLaneRecheckDoctrine` | keep tokens `CronCreate`, `recurring`, `disk evidence` in that section |
| A `## 5.` (spans §5.1) | `TestLaneRecheckDoctrine` | keep `standing`, `CronCreate`, `recurring`, `renew_interval`, `before the first stage`, `CronList`, `CronDelete` |
| S15 W §0a | `TestLaneWatchdogRenewsFirst`, `TestLaneWatchdogRecordsRefusalAndMeasures` | the first procedural step keeps "Renew" in its heading, `moai factory renew` in its body, precedes "Measure progress", and keeps the tokens `--json`, `result`, `renew_result`, `renewed`, `refused-expired`, `refused-not-holder`, `error`, `does not end the pass`; the scope clause is additive and must not extend the closed set of `renew_result` values — a pass that does not renew carries the previous value forward (the same carry-forward language §1 already uses for `last_change_at`) |
| S18 F:38 and F:40 | F:38: `TestAutoRankDoctrineAmendment` (prohibition list, lines 113-127), `TestAutoPickDocDoctrine` (`contain` / `absent`), `TestAutoRankMirrorParity` (the passage starts at F:38's opening words); F:40: ledger row ALB-0256 (`rule-body:`, `after_text` presence) | edit neither; the exception is a new pointer line after F:40 inside the mirror-parity passage (identical in both trees, neutral); the new text never contains `Promotion is the operator's act, always.` or `The leader never picks for the operator` |
| S19 F:44 and K:20 opening clause | ledger row ALB-0257 (`always:` core, verbatim); the paragraph's literals `auto-scoped ranking exception` / `selection order only` and the pinned sentences stay in one paragraph (`TestAutoRankDoctrineAmendment`, `TestAutoRankMirrorParity`) | change only `The one reconciliation is named, not excepted:` → `A reconciliation is named, not excepted:`, identically in F and K, and update ALB-0257's `after_text` to the new K:20 text in the same change |
| S21 G:307-310 | `autoDocGTDProhibitions` ("The pick is the operator's. Do not preselect, …") | keep the sentence; add the pointer as a separate paragraph after the [HARD] paragraph; outside the `--auto` passage of `TestAutoRankMirrorParity` |
| M `## The leader's remaining role` table | no ledger row, no Go test pins the table (`git grep` over `internal` found only a comment in `batch_gate_summary_doctrine_test.go` and the ledger row of the stub line that points at it) | add one row by pointer (REQ-MJ-017 (d)); the stub's pointer sentence (`factory-dispatch-core.md:12`) stays true and is not edited |
| the card-id literals in the changed files | `card_id_leak_test.go` baseline: `t1339`, `t1393` (A), `t133`, `t224` (D), `t696` (G); `TestCardIDBaselineHasNoStaleEntries` | keep the sentences that carry them (A:126, A:146, D:24, D:135, G:258); AC-MJ-013 excludes exactly these five literals and fires on any other |
| any ledger-covered line of ≥ 40 UTF-16 units | AC-ALB-021(2) fragment rule | never copy such a line verbatim into a `paths:`-scoped rule or a skill file; new text is worded fresh |
| C | `TestAutoPickDocDoctrine` | keep the sentence "a pull request or landed state is a skip input for a queued candidate the session chose and is report-only for an operator-picked card" (it is in C's `## The pre-dispatch cross-check`, not in the section being edited) |
| G | `TestTodoSkillDocumentsClassification` | keep `--classification-file`, `` `serial` ``, `` `parallelizable` ``, `` `high` ``, `` `normal` ``, `` `low` ``, `sorted`; the template copy carries no SPEC id, REQ token, `20YY-MM-DD` date, or 9-40 hex run (note: this applies to the new G text too) |
| all templates | `TestTemplateNoInternalContentLeak`, card-id leak tests | no card id (`t` + digits), board id (`a672`/`0c51`/`02f4`), SPEC id, date, or SHA in any template text. The SPEC may cite them; the shipped documents may not — they state the rule, and the board record stays in `.moai/specs` |

## D-4 The core-stub budget (REQ-MJ-014)

`factory-dispatch-core.md` measures **9,999** UTF-16 code units against a budget of 10,000. Two core edits are needed, and both land in the M2 commit with their template and deployed twins (no separate trim step):

| Edit | Result | Delta |
|---|---|---|
| K:28 tail replaced by `is kept while the lane executes the card.` (drops "to the completion report": the scope marker now states the end) | scoped, keeps the pinned wording | **+4** |
| K:20 opening clause `A reconciliation is named, not excepted:` (was `The one reconciliation is named, not excepted:`) | drops the exhaustiveness claim | **−6** |
| Net | 9,999 → **9,997** | **−2** |

A pointer form for K:28 (scope marker plus a cite of `factory-dispatch-mechanics.md` § Lease renewal) measured **+115** and is rejected: the stub cannot carry it. The promotion pointer is deliberately not in the stub: the stub holds no promotion clause today (F:38 and F:40 are body lines outside it). Because the net is negative, no offset trim of any other core line is planned or needed, and no ledger-covered line other than ALB-0257 (K:20) and the K:28 row changes. AC-MJ-012 is the mechanical check (it prints `OK 9997` on the planned wording and exits 1 on `OVER`); `TestRoleCoreStubGate` is the real gate. The pointer to § Lease renewal lives in F's two lines (path-scoped) instead of the stub.

## D-5 The scope marker and the per-line trap

REQ-MJ-010 is checked with one marker phrase, `while the lane executes the card`, on every line that arms or keeps the cron. Two traps:

- **Per physical line.** AC-MJ-008a's regex works line by line. `auto-semantics.md` §5.1 is hard-wrapped near 80 columns, and today the phrase "kept until the completion report." sits alone on line 118. The scoped sentence must keep the marker and the `kept` / `keeps` / `stays armed` … `completion report` words on one physical line, or the sentence must be reflowed. A mutant that scopes the sentence but leaves a wrapped tail line unscoped is exactly what the regex catches.
- **Scoping is not deleting.** The cron's purpose — waking a lane stopped by an API error or a delegate report that never comes — is kept for an executing card (REQ-MJ-010). AC-MJ-008c keeps the renewal-period wording on K and F; the doctrine tokens in §D-3 keep A and F's section content. Deleting the sentence would pass AC-MJ-008a and fail these.

## D-6 What the mode judgment can and cannot be checked for

REQ-MJ-001..003 specify what the document must say and what a reason must contain. No command here can establish that a card's stated file set is its true file set, or that "overlaps an in-flight card" was judged correctly; a reason field is human or LLM text. The documents therefore ask for the *evidence of the judgment* (the file set and the conflict check, named), which a reviewer or the next audit can read and dispute, and the SPEC claims nothing mechanical beyond that (spec.md R-2). The decision-index rows Q5 record the two places the board's words leave judgment open ("in-flight" membership; the paths behind "factory/todo core, release pipeline, shared migrations").

## D-7 Class C: `--spec` at pick

0c51: "Class C cards get --spec at pick time (cannot be added after assign)." `gtd.md` today tells the reader (lines 267-270) that an empty `spec_id` on a picked card "is the normal state" and (330-333) to "Attach the SPEC only when one exists and its identifier is known". For a Class C card that is the opposite of the board rule, so S4 and S5 are qualified, not just supplemented: the sentence "An empty `spec_id` on a picked card is" must no longer stand unqualified (AC-MJ-004b), and C's `### Mode judgment` carries the rule once (home token `cannot be added after assign`). The `--spec` flag exists on `moai todo next` and not on `add` (`moai todo next --help`, build `moai_cp/20261007_195927`).

## D-8 Free lane versus live session (P8)

Two independent conditions decide whether a lane can take a card: the lane has a live session (a session occupies the label; `factory-dispatch-mechanics.md:193`, kept) AND no card is actually executing on it (new). A lane failing the first is a fault the leader reports; a lane failing the second is busy. D's Routing bullet states the second and cites M § Boundaries for the first, so a reader cannot read "free" as "no session". The sentence in D:133 ("free means the lane's previous card reached `done` and its evidence was read") is the one that conflicts and must go (AC-MJ-005b).

## D-8a Delegated-dispatch promotion (the operator ruling d314)

The card's phrase "assignment input = picked + queued parallelizable" read two ways — a description of the candidate set, or a licence for the leader to promote. The operator ruling d314 chose the second, with limits (`spec.md` §2 quotes it). Three design consequences:

1. **The old sentences stay; the exception stands beside them.** F:38, F:40 and the `--auto` paragraph are pinned by Go doctrine tests and a ledger row (D-3, rows S18-S21); this card changes no Go. Folding the exception into them would turn those guards red. The exception is therefore one home subsection in D plus pointers — which is also the "each rule once" requirement (REQ-MJ-012). Read together, F:38 says the leader never picks outside an `--auto` authorization, and the pointer line right after F:40 names the one further stated case; the heading of F:38 already speaks of promotion "in advance", which is how the two sit together without a contradiction (a reading, not an added rule — the rule added is the ruling's, in the home subsection).
2. **Only one assertion changes meaning, and it is rewritten minimally.** "The one reconciliation is named" claimed there was a single reconciliation; with a second stated case it is false, so S19 drops the exhaustiveness (identical in F and K).
3. **Every limit is a pinned whole line.** The ruling is narrow and mostly negations; `acceptance.md` E-006 pins eight whole lines (`AC-MJ-006`; seven limits plus the heading's rule sentence set, with the assignment-input sentence labelled a consequence, not a further limit) and scans for the stems `promot`, `queued`, `parallelizable` outside the subsection, so a widening cannot hide inside a sentence that keeps the right words. The price is that D's wording is bound by the criterion; an improvement amends the criterion with it.

Not decided here, and not invented: what establishes "the operator has delegated dispatch" (R-5). The subsection carries the ruling's words, `(operator away, or a standing delegation)`, and nothing more.

## D-9 Candidate wording (not normative; the anchor tokens and constraints above are)

Mode judgment, for C (neutral, no ids):

> ### Mode judgment — serial or parallelizable
>
> Class names the stages a card needs; mode names whether it may run beside other cards. Decide the mode at design time, from the card's expected file set — never from the queue's default. A card is `serial` only when it touches the same files or shared state as an in-flight card, or lies on a slot-gated surface (factory/todo core, release pipeline, shared migrations); every other card is `parallelizable`. The reason recorded with the classification states the file set and the conflict check, and for a `serial` card names the conflicting card or surface. Supply the judgment with `--classification-file` (decider `human`). A card that reached the queue on the fail-safe default is reissued with an explicit classification. For a Class C card, give `--spec` when the card is picked: it cannot be added after assign.

Leader line for F after line 30, G:52, M `### Lease renewal`, D Routing bullet: wording in `plan.md` M1-M3 steps. The renewal home must contain the tokens `only the card the lane is actually executing`, `merge-ready`, `held`, `serial-waiting`, `expiry is harmless`, `notice`, `factory next --card`, and the marker `while the lane executes the card`. D's Routing bullet must contain `no card is actually executing` and `not a free lane`. The delegated-dispatch subsection of D is **not** candidate wording: its eight lines are pinned verbatim by `acceptance.md` E-006, the F pointer line by E-018a and the G pointer paragraph by E-018d; this file does not repeat them (one copy, so no drift).

## D-10 What the acceptance family can see, and what it cannot

| Residual | Why no command closes it |
|---|---|
| A paraphrased restatement of a rule in a second file | AC-MJ-016 matches verbatim home tokens only |
| An issuance instruction in prose with no command text | V (`research.md` §7) keys on the command string; AC-MJ-002/-003 add positive presence checks |
| Whether the lane actually follows the new wording | AC-MJ-015 is an operational check |
| The Go SessionStart notice still saying the old thing | out of scope (Q4 moved to card t1695 by d314, R-1) |
| The exception written in another file under a paraphrase that carries no home token | AC-MJ-016 and AC-MJ-018 match verbatim tokens and whole lines; the sync audit reads for it |
| An actor treating an unstated delegation as given | the ruling names no mechanism (R-5); AC-MJ-006 pins text, not practice |
| The 0-hit absence checks of `TestAutoPickDocDoctrine` (`absent` strings) | no live control exists in a single read-only command; left to the Go run of AC-MJ-014 |
| Wording drift between template and deployed | AC-MJ-011 (`cmp`, plus the per-milestone `git diff-tree` coherence check) and the Go parity guards, run-phase |
| A stem-less widening of the Routing bullet (no `promot`/`queued`/`parallelizable` word) | caught by AC-MJ-005 (E-005d pins the three Routing lines), not by AC-MJ-006 (mutant N9, recorded in `progress.md`) |
| A sentence contradicting a pinned line while also carrying the marker phrase | per-site fragment counts in E-008b bound the marker to the six named sites; a contradiction elsewhere is read by the audit |
| The simulated post-run text used to prove the checks red/green | it is not evidence about the real tree (Gap G-15); the run-phase repeats each RED-now/green pair on the real tree |
