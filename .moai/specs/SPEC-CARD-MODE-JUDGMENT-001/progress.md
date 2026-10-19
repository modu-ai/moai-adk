# progress.md — SPEC-CARD-MODE-JUDGMENT-001

> Phase progress record. Only `§E.1` is filled by the plan phase (manager-spec); `§E.2` and `§E.3` belong to the run phase (manager-develop), `§E.4` to the sync phase (manager-docs).

## §E.1 Plan-phase Audit-Ready Signal

plan_status: audit-ready
audit_ready: true
plan_complete_at: 2026-10-19T05:17:41Z
tier: L
artifacts: spec.md v0.1.2 (REQ 17) · plan.md (M1-M5) · acceptance.md (AC 19: 42 ledger entries) · design.md (D-1..D-10, D-8a) · research.md (§1-§12) · decision-index.md (Q1-Q7) · progress.md
measured_tree: b7b44433c312da73de9f329ab08c8fa6c5aa4de0 (branch WT-fail-safe-serial, unpushed) for correction round 1; earlier pins 74ad197380a0711cdb9816d4e37c94d29af730ea and 55d4f1486a454ff446be6f0f0645d4cea76bf0cc (the document trees are identical between the three pins)
open_clarifications: 1 — decision-index Q7 (FOUNDER, product-level, empty verdict): what establishes "the operator has delegated dispatch" (spec R-5). It blocks the autonomous Kickoff by the decision-index rule and goes to the operator. Q3 was met by the operator ruling d-20261019T023441Z-d314 and is unchanged. Q1, Q2, Q4, Q5 carry `DEFAULT-APPLIED`; Q6 is FOUNDER implementation-level with `DEFAULT-APPLIED 2026-10-19T03:35:14Z` (Tier L). No requirement, criterion or milestone depends on Q7; it gates only the Kickoff form.
spec_lint: NOT RUN in the final state of correction round 1 — the Bash tool refused every command late in the round ("working directory resolved to the shared checkout"), including a bare `cd`, `git -C` and `echo`. The last lint result (`0 error(s), 0 warning(s)`, exit 0, one INFO) predates the correction round and does not describe spec.md v0.1.2. See Gap G-16.
spec_audit: `mcp__moai__spec_audit` (filter SPEC-CARD-MODE-JUDGMENT-001, project_root = this worktree), run in the correction round: one INFO `EraAutoDetected` with `heuristic_matched` "H-5 (modern phase or created date)", `era` V3R6, `modern_era_clean` 1, no drift finding.
plan_artifact_hash: not recomputed here — the earlier plan audit was stopped before it wrote a verdict, so no hash covers any version of these files; the lane recomputes the hash after the correction commit.

### Correction round 1 (after plan audit iteration 0, FAIL 0.65, defects D1-D9)

Each defect, its evidence command before and after (all cells re-run on `b7b44433c3` except where stated; the "after" values were observed on a scratch simulation of the post-run text, G-15, and on the real tree for the RED-now cells):

| Defect | Before | After |
|---|---|---|
| D1 commit ordering (Template-First unverifiable in one commit) | REQ-MJ-013 claimed an edit order the commit graph cannot witness | REQ-MJ-013 claims only per-milestone-commit coherence; both stub edits, catalog hashes and ledger rows ride the commit that changes what they pin; AC-MJ-011 adds a `git diff-tree` check (run-phase) |
| D2 AC-MJ-006 scan | printed `9 9` on the good copy and on all three escape mutants | bounded by the next `^## `, end marker required, 8 whole lines pinned, stem scan outside the subsection; prints `8 8 1 1`; 16 of 17 mutants red (N9, a stem-less Routing widening, is caught by AC-MJ-005) |
| D3 unscoped cron statements | 7 unscoped keep/arm lines per tree (K:28, F:119, F:122, A:117, A:118, A:396, W:244) | ten sites enumerated (research.md §11, REQ-MJ-010); AC-MJ-008 sweep (a) + per-site fragment counts K1/F2/A2/W1 + renewal wording + `when the lane stops executing the card` |
| D4 predicate narrowing | `shared state` only in the spec.md quote | REQ-MJ-002/-003 and the AC-MJ-001 token restored to "same files or shared state" (6 tokens) |
| D5 traceability | trace verb printed UNCOVERED for REQ-MJ-002, -003, -009 | full REQ ids in every AC header; re-run: `COLLECTED 17`, no UNCOVERED line |
| D6 vacuous checks | three mutants passed the old checks | AC-MJ-004 whole lines, AC-MJ-008 site fragments, AC-MJ-009 pinned scope line; each mutant observed red |
| D7 REQ-MJ-007 | nine lines, one a non-limit | exactly the seven limits of d314 plus a labelled consequence = eight pinned lines |
| D8 leader-duty enumeration | the M table of retained duties would disagree with the new subsection | REQ-MJ-017 (d), AC-MJ-018 (e): one row by pointer; research.md §12 |
| D9 card-id guard | diff-based `-G` fired on a harmless re-wrap | AC-MJ-013 uses a portable working-tree ERE excluding the five baselined ids; positive control prints 8; mutant `t1500` fires |

Stub measurement: before `OK 9999`; after the planned wording (K:28 tail +4, K:20 clause −6) `OK 9997`; mutants: pointer form `OVER 10162`, K:20 edit omitted `OVER 10003`, marker phrase pasted `OVER 10027` (all on the scratch simulation).

Optional defects D10-D15:

- **D10** added decision-index Q7 (FOUNDER, product-level, empty verdict) — flagged above, blocks the autonomous Kickoff.
- **D11** Q6 relabelled FOUNDER implementation-level with `DEFAULT-APPLIED 2026-10-19T03:35:14Z` (Tier L); the table it cited is guidance, not an operator setting.
- **D12** consequences of the Q2 and Q5 defaults written into the rows and into residual risks R-7 and R-2.
- **D13** R-6 now records that a later card folds the exception into the pinned Go doctrine tests.
- **D14** every `-P` and `\b` removed; AC-MJ-012 exits 1 on OVER; the `Tools` convention line added.
- **D15** only board ids are cited; 02f4 is quoted verbatim (checked as a substring of the saved board file); the scope items are listed in §1; REQ-MJ-004 is a `shall` clause; the `card:` frontmatter key was dropped (the card id stays in `tags`).

### Correction round (before the first plan audit; a second commit, not an amendment)

What changed, and why:

1. **Decision-index Q3 settled.** The operator ruling `d-20261019T023441Z-d314` (verbatim in the Q3 row and in `spec.md` §2) lets the leader promote queued `parallelizable` cards to `picked` and assign them while the operator has delegated dispatch, under seven limits. The Q3 row carries the verdict and is marked met; its Gate section and the milestone-to-row map say so; no text calls REQ-MJ-007 / AC-MJ-006 / M3 gated any more. Q4 and Q1 carry a record note quoting the ruling's last sentence; Q2, Q5, Q6 are untouched.
2. **REQ-MJ-007 and AC-MJ-006 rewritten** to implement exactly the ruling: one home subsection in D, nine whole lines pinned by AC-MJ-006 (RED-now `0 0` / `0 0`, positive control and six mutants observed in a scratch copy, all red).
3. **REQ-MJ-017, AC-MJ-018, AC-MJ-019 added** for the sentences that said promotion is the operator's alone. `research.md` §10 measures every such sentence (commands, stdout, exit codes, tree) and decides it: F:38 and F:40 kept verbatim plus one pointer line (pinned by three Go doctrine-test families and ledger row ALB-0256); the opening clause "The one reconciliation is named" rewritten identically in F:44 and K:20 (ledger row ALB-0257; stub −6 units); the todo skill's pick rule kept plus a pointer; the rest kept, with reasons.
4. **Counts:** REQ 16 → 17; AC 17 → 19 (R 10 → 11, G 6 → 7, O 1); milestones M1-M5 unchanged; files affected 18 (unchanged). Both ceilings (25) hold, and `spec.md`, `acceptance.md`, `plan.md` and `design.md` D-1 agree.
5. **Stub budget re-measured:** `OK 9999`, template and deployed stubs `cmp`-equal (research.md §8); the net of the planned K edits is +3, so a ≥ 2-unit offset is required (plan M2 step 2, AC-MJ-012).

Rows the ruling affects, listed and not decided here:

- **Q5** (in-flight membership; paths of the three slot-gated surfaces): the ruling's "serial cards follow the slot rules" and "keep-set, held and [보류-marked cards" lean on the same undefined membership; its default (board words verbatim) is unchanged, and the documents cite the existing keep-set (`auto-semantics.md` §9.3) rather than restate it.
- **A possible new row, not authored:** what establishes "the operator has delegated dispatch" (spec R-5). The ruling gives no mechanism; if the operator wants one, it is a FOUNDER, product-level question.
- **Doctrine homes the ruling does not mention** (left unedited, reasons in `research.md` §10): the `auto-semantics.md` §9 gate inventory and the §9.3 card-pick authorization, `contract-autonomy.md:92`, the foreman skill. They describe other paths; whether the inventory should name the new case is for the operator or a later card.

### Premise record (card text and premise-check.md P1-P12, re-measured; commands and outputs in research.md §1 and acceptance.md §2)

| # | Premise | Result |
|---|---|---|
| P1 | template = deployed `factory-dispatch-cards.md` | holds (`cmp` exit 0). Seven of the eight target pairs are byte-identical; `factory-dispatch.md` is **not** (`cmp` exit 1, one declared live-only line, `TestAutoPickMirrorParity`) — premise P1 is false for that file and it is handled as AC-MJ-017 |
| P2 | no mode-judgment step | holds |
| P3 | `--classification-file` documented only in `gtd.md` | holds; additionally `manager-todo.md` and `manager-lead.md` describe **no** issuance step (Q1) |
| P4 | candidate issuance documents | resolved: only `gtd.md:52` and `factory-dispatch.md:30` describe issuing; `doc-generation.md:348` is a third, bare, out of scope by default (Q2) |
| P5 | `gtd.md:56` states the code default | holds; kept (REQ-MJ-015) |
| P6 | `todo_classify.go:41` | notice constant; default decider at line 36 |
| P7 | three renewal sites contradict a672 | holds and widens: five document sites (`factory-dispatch-core.md:28`, `factory-dispatch.md:119` and `:122`, `auto-semantics.md` §5.1, `factory-dispatch-mechanics.md:26`) plus `moai-lane-watchdog/SKILL.md` §0a, plus the Go SessionStart notice (out of scope, Q4) |
| P8 | free-lane wording | holds; `…-detail.md:133` conflicts, `…-mechanics.md:193` is the separate live-session notion |
| P9 | authorities 0c51, a672, 02f4 | read with `moai decision read`; quoted in spec.md §2 |
| P10 | code is t1737's | holds; `SPEC-FACTORY-DEFAULT-MODE-001` is not in this tree |
| P11 | Template-First build chain | read; no Go run; `make build` not required for a documents-only change |
| P12 | `spec_id` set on the card | holds |

### What this plan measured beyond the premise check (summary; detail in research.md)

- Mirror and sync mechanism: Template-First; `*-check` make targets are read-only drift detectors; the real parity guards are `TestRuleTemplateMirrorDrift`, `TestAutoPickMirrorParity`, `TestReviewOwnership_CardReviewDoctrine`, `TestTodoSkillDocumentsClassification`. The `.moai/policies/**` and `.moai/workflows/gtd.md` copies are stale deploy-time projections, not guards.
- Pinning tests that constrain the wording: the binding ledger (rows ALB-0254, -0265, -0268, -0276, -0304 touch the target lines), the renew-interval token test, `TestLaneRecheckDoctrine`, `TestLaneWatchdogRenewsFirst` / `…RecordsRefusalAndMeasures`, the card-id/internal-content leak tests, the AC-ALB-021(2) fragment rule.
- Budget: `factory-dispatch-core.md` measures 9,999 UTF-16 units against 10,000 — one unit of headroom; the two planned stub edits (K:28 +4, K:20 −6) net −2 (`OK 9997` simulated), so no offset trim is needed.
- Catalog: skill-directory hashes in `internal/template/catalog.yaml` change for `moai` (it contains `workflows/gtd.md`) and `moai-lane-watchdog`.
- Sweep for "no statement relies on the serial default when issuing": violating sentence V defined once (research.md §7); fires on exactly `gtd.md:52`; positive control observed on the base commit.

### Gaps (not observed in this phase)

- **G-1** The Tier question that `spec-assembly.md` poses to the operator could not be asked from a subagent; Tier L is the plan author's classification from the files-affected column (decision-index Q6, design.md D-1).
- **G-2** No Go command ran. Every statement about what a guard test pins comes from reading its source; none was executed. AC-MJ-014 takes the undecidable disposition (regression-guard, not recorded as a pass).
- **G-3** The binding ledger tests, the 10,000-unit core gate and the 150,000 subtotal were not run; the unit counts are a Python UTF-16 measurement of the files as stored, while the Go gates render the template first.
- **G-4** The worktree-isolation guard refused several read-only invocations during this phase: a loop that mixed `cmp` with a piped `git ls-files`, two `sed`/`grep` calls whose path was a shell variable, and three `git diff -G` calls whose pattern carried `{n}` quantifiers ("a form too complex to verify"). Each was re-issued as a plain command (one `git`, literal paths) or with the quantifier written out without braces. Recorded here per `verification-claim-integrity.md` §3.1; no result was inferred from a refused command. The same refusal will apply to a run-phase author who writes the AC-MJ-013 pattern with `{n}`.
- **G-5** Every `moai` CLI observation (`decision read`, `todo next --help`, `factory status`, `spec lint`) was judged by the installed build `moai_cp/20261007_195927`, which predates HEAD; `spec lint` in particular was not run from a build made from this tree.
- **G-6** "Class C `--spec` cannot be added after assign" is board 0c51's statement; this plan saw `--spec` on `moai todo next` and not on `add`, and did not read the assign path in code.
- **G-7** (superseded by correction round 1: the stub edits net −2, so no trim is planned) Which core lines were safe to trim was measured only against the binding ledger; other pinning tests (for example `role_entry_points_test.go`) were not read.
- **G-8** The Go SessionStart notice was read from source (`session_start_factory_i18n.go`), not observed at a live join.
- **G-9** `grep -c` printed per-file lines in an order different from the argument order; the acceptance expectations are therefore stated as sets.
- **G-10** `TestAutoPickMirrorParity` and `TestRuleTemplateMirrorDrift` were read in the regions that name the target files, not end to end.
- **G-11** (correction round) The `internal/cli` doctrine tests that pin the old operator-only sentences were read, not run: `todo_auto_doc_test.go` lines 85-420 and `todo_auto_pick_doc_test.go` lines 1-150. The bodies of `TestAutoPickDocDoctrine` and `TestAutoPickMirrorParity` (from line 181 and 279) were not read; their `contain` / `absent` / parity literals come from the surface table at lines 40-150 and from `research.md` §6. Of `binding_ledger_test.go` only lines 470-540, 815-892 and 985-1075 were read (row coverage, after-text presence, the AC-ALB-021(2) sweep): the claim that a `rule-body:` row (ALB-0256) is checked by `after_text` presence only, and that an `always:` row (ALB-0257) is checked by `after_text` presence in the deployed stub, rests on those lines; the mutation fixtures further down (lines 1196-1500) were not read.
- **G-15** (correction round 1) The green paths, positive controls and mutants were observed on a scratch simulation of the post-run text (eight documents plus a mutation harness under the session scratchpad). The simulation is evidence that the criteria are satisfiable and that each listed mutant is red; it is not evidence about the real tree after the run. Scratch files are not cited and are named here as a known loss.
- **G-16** (correction round 1) Late in the round the Bash tool refused every command (message: working directory resolved to the shared checkout), including `cd`, `git -C <worktree> rev-parse` and `echo`. Consequences, none inferred as passed: `moai spec lint --strict` was not run on spec.md v0.1.2; the final `git status`, the HEAD/branch re-read and the commit could not be made; `plan_complete_at` is the `audited_at` clock reading of `mcp__moai__spec_audit` (one INFO `EraAutoDetected`, `modern_era_clean` 1, no drift finding), the only clock source left; the final repo-wide sweeps for stale "nine lines" / "Template-First order" / "9,999" wording were done by reading design.md, research.md, progress.md and the spec artifacts, not by `grep`.
- **G-12** (correction round) The AC-MJ-006 positive control and its six mutants were observed against scratch copies of the pinned subsection, under `/private/tmp/claude-501/-Users-goos-MoAI-moai-adk-go/3485ddff-4832-470a-9a17-ee4dac7ad178/scratchpad/ac6/` (`good.md`, `mut_held.md`, `mut_nocond.md`, `mut_serial.md`, `mut_append.md`, `mut_dup.md`, `mut_extra.md`). Those files are scratch, not cited as evidence; they are named here as a known loss. They are reproducible from the nine lines of `acceptance.md` E-006. The RED-now cell itself runs on the real files.
- **G-13** (correction round) The worktree-isolation guard refused one compound command (a `cd` into the scratch directory followed by a heredoc and several `sed` calls). It was re-issued as separate plain commands and the `Write` tool; no result was inferred from the refused command (`verification-claim-integrity.md` §3.1).
- **G-14** (correction round) `moai decision read` and `moai spec lint` were judged by the installed build `moai_cp/20261007_195927`, which predates HEAD (`moai version` printed that build); the board record is data, so the lag does not change what was read, but `spec lint` was not run from a build made from this tree.
### Residual risk

- The mode judgment is text a human or an LLM writes; no command here proves a stated file set is the real one (spec R-2).
- Until the Go SessionStart notice changes (Q4), a lane reads two instructions about the standing cron (spec R-1).
- The ruling d314 names no mechanism that establishes "the operator has delegated dispatch" (spec R-5). AC-MJ-006 pins the text of the exception, not whether an actor treats an unstated delegation as given.
- AC-MJ-006 pins D's wording line by line; any wording improvement during the run phase must amend the criterion with it. A pure paraphrase of the exception placed in a second file evades AC-MJ-016 / AC-MJ-018 (verbatim tokens and whole lines); the sync audit reads for it.
- Three Go doctrine-test families and one ledger row pin the old operator-only sentences; the plan keeps them verbatim and adds the exception beside them, so a reader meets both an absolute sentence and a pointer to its one stated exception. That is deliberate (no Go change in this card) and is the price of "each rule stated once".
- The `internal/cli` Go sources (`todo.go:6`, `:238`, `:1576`) still say the next pick is the operator's act; they describe the operator's own `next` pick, which the ruling does not change.
- A paraphrased restatement of a rule in a second file evades AC-MJ-016; the sync audit reads for it.

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
