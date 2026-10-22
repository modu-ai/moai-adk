# progress.md — SPEC-CARD-MODE-JUDGMENT-001

> Phase progress record. Only `§E.1` is filled by the plan phase (manager-spec); `§E.2` and `§E.3` belong to the run phase (manager-develop), `§E.4` to the sync phase (manager-docs).

## §E.1 Plan-phase Audit-Ready Signal

plan_status: audit-ready
audit_ready: true
plan_complete_at: 2026-10-19T13:47:03Z
tier: L
artifacts: spec.md v0.1.4 (REQ 17) · plan.md (M1-M5) · acceptance.md (AC 20: 53 ledger entries) · design.md (D-1..D-11, D-8a) · research.md (§1-§12) · decision-index.md (Q1-Q7) · closed-world-manifest.txt · closed-world-check.py · progress.md
measured_tree: d31fb64239a688c053e455000cc9fe37f540b790 (branch WT-fail-safe-serial, unpushed) for correction round 2; earlier pins b7b44433c312da73de9f329ab08c8fa6c5aa4de0, 74ad197380a0711cdb9816d4e37c94d29af730ea and 55d4f1486a454ff446be6f0f0645d4cea76bf0cc (the document trees are identical between the four pins: the diff between b7b44433c3 and d31fb64239 over internal, .claude, .agents and .codex printed nothing); the ceiling-exception round measured on HEAD f80aef625e with the SPEC-directory edits uncommitted in the working tree, the sixteen documents byte-identical to the pin (the checker's first, third and fifth numbers are 0 on every row)
open_clarifications: 0 — Q7 answered 2026-10-23 (relayed by the leader), see decision-index Q7 (FOUNDER, product-level, verdict met: the ruling's phrase is the whole condition): what establishes "the operator has delegated dispatch" (spec R-5). No requirement, criterion or milestone of this SPEC waits on Q7; REQ-MJ-007 (a) is text-final because it carries the ruling's words and no mechanism. Per the decision-gate rule the empty FOUNDER row Q7 goes to the operator and blocks the autonomous form of the plan→run Kickoff until the operator answers it or states that the ruling's phrase is the whole condition. Q3 was met by the operator ruling d-20261019T023441Z-d314 and is unchanged. Q1, Q2, Q4, Q5 carry `DEFAULT-APPLIED`; Q6 is FOUNDER implementation-level with `DEFAULT-APPLIED 2026-10-19T03:35:14Z` (Tier L).
spec_lint: `moai spec lint --strict .moai/specs/SPEC-CARD-MODE-JUDGMENT-001/spec.md` run by this lane in correction round 2 on spec.md v0.1.3: `0 error(s), 0 warning(s)`, one INFO `OwnershipTransitionUnmeasured` (the plan-phase commit carries no Authored-By-Agent trailer), exit 0; re-run in the ceiling-exception round on spec.md v0.1.4: `0 error(s), 0 warning(s)`, the same single INFO, exit 0. The lint was NOT RUN at the end of correction round 1 (the Bash tool refused every command late in that round, Gap G-16); the lane ran it after that round and the iteration 1 audit reproduced `0 error(s), 0 warning(s)`, exit 0, on spec.md v0.1.2. Judging build: the installed `moai_cp/20261007_195927`, which predates HEAD (G-14).
spec_audit: `mcp__moai__spec_audit` (filter SPEC-CARD-MODE-JUDGMENT-001, project_root = this worktree), run again in correction round 2 (`audited_at` 2026-10-19T05:56:04Z): one INFO `EraAutoDetected` with `heuristic_matched` "H-5 (modern phase or created date)", `era` V3R6, `modern_era_clean` 1, `total_specs` 1, no drift finding.
plan_artifact_hash: not recomputed by this agent (it has no audit_cache tool). The iteration 1 report recorded `74d418b30f1f2ba8ae90cb22f41e21e4eafbe76ebea879994576e956f19b60e7` for the tree before round 2; the plan artifacts changed in round 2, so the lane recomputes the hash after the correction commit. `closed-world-manifest.txt` and `closed-world-check.py` are not in the hash subject set; their SHA-256 values are pinned in `acceptance.md` E-020d, which is.

### Correction round 2 (after plan audit iteration 1, FAIL 0.79, defects D16-D22; the last round of the Tier L ceiling)

The iteration 1 report is `.moai/reports/t1894/plan-audit-iter1.md` (receipt rcpt-2adbd9564960a6f7da294042; the required codex gate failed; D1-D15 closed). Each defect, its evidence before and after, on tree `d31fb64239a688c053e455000cc9fe37f540b790` for the real-tree cells and on a scratch simulation of the post-run text (G-15, rebuilt in this round) for the green paths and mutants.

| Defect | Before | After |
|---|---|---|
| D16 (blocking) stem-less added line passes AC-MJ-005 and AC-MJ-006 | four mutants — an extra Routing sub-bullet (two wordings), a sentence in the card-review section, a sentence before the subsection heading — printed E-005d `:3` / `:3` and E-006 `8 8 1 1` / `8 8 1 1`, exit 0 (reproduced here on the simulation); design.md line 117 said AC-MJ-005 catches them | new AC-MJ-020 with a pinned manifest of every allowed added and removed line (73 rows — 48 added lines, 25 removed — over the 8 documents, applied to both trees, 16 files) and a read-only check: RED-now on the real tree `C 0 2 0 0 / K 0 2 0 2 / F 0 5 0 3 / D 0 12 0 1 / M 0 7 0 1 / A 0 8 0 8 / G 0 10 0 8 / W 0 2 0 2` (both sides), exit 1; the simulation prints `0 0 0 0` on all sixteen files, exit 0; the four mutants print `D template 1 0 0 0 deployed 1 0 0 0`, exit 1; 74 distinct mutants (the 21 of the AC-MJ-006 family, 37 of the earlier cross-criterion list, 16 further) all red; design.md D-10 states what each criterion catches and D-11 describes the check |
| D17 two of the ten cron sites have no check | with the A §5.1 fire-and-renew sentence and M:26 left unedited and everything else as planned, the old cells printed E-007 `4 4 1` / `4 4 1`, E-008a empty, E-008b K 1 F 2 A 2 W 1, E-008c and E-008d `1`, all exit 0 | new E-008e (two whole lines of the fire-and-renew sentence) and E-008f (the opening of M:26): RED-now `0` / `0` and `0` / `0`, exit 1; the simulation prints `2` / `2` and `1` / `1`, exit 0; the unedited-sentence mutants print `0` / `0`, exit 1; one reworded line prints `1` / `1`, exit 1 |
| D18 E-002a, E-006, E-009 refused by the guard | refused as written (iteration 1 audit) | the alternation moved to `-v re=` in each; each cell was run through the guard in this round and printed the recorded RED values (E-002a the `gtd.md:52` line, E-006 `0 0 0 1` ×2, E-009 `0 1 0` ×2), exit 1; §0 Tools names the restriction; E-001, E-008e, E-008f and the python cells were also run through the guard |
| D19 two statements about what Q7 gates | spec.md §6 "nothing is gated on it", plan.md §C "gates no text", acceptance.md §4 "No requirement or criterion is gated on a decision" against decision-index.md and progress.md "blocks the autonomous Kickoff" | one statement everywhere (spec.md §6 below the table, plan.md §C, acceptance.md §4, decision-index.md Q7 gate note and Gate section, this record): no requirement, criterion or milestone of this SPEC waits on Q7; REQ-MJ-007 (a) is text-final; per the decision-gate rule the empty FOUNDER row Q7 goes to the operator and blocks the autonomous form of the plan→run Kickoff until the operator answers it or states that the ruling's phrase is the whole condition; the Q3 row of spec.md §6 now says "no text waits on it" |
| D20 two requirement halves carry no token | a `### Mode judgment` subsection without the reissue sentence and without the conflicting-card clause still printed `6` | AC-MJ-001 has eight tokens, adding `names the conflicting card` and `reissued with an explicit classification`; RED-now `0` / `0`, exit 1; simulation `8` / `8`; mutants: conflicting-card clause deleted `7` / `7`, reissue sentence deleted `7` / `7`, predicate narrowed `7` / `7`, heading renamed `0` / `0`, deployed only `8` / `0` |
| D21 commit-shape check lists, does not decide | a bare `git diff-tree` listing exits 0 whatever it lists | the `pairs` mode of `closed-world-check.py` prints `<sha> pairs <n> unpaired <m> expected <n>` and exits 1 on any unpaired document or on a different pair count; known inputs from history: a deployed-only commit prints `pairs 0 unpaired 7`, exit 1 (E-011g), a two-pair commit with expected count 2 prints `pairs 2 unpaired 0 expected 2`, exit 0 (E-011h); the run-phase obligation names the expected count per milestone (M1 3, M2 5, M3 4) |
| D22 stale lines in this record | `spec_lint: NOT RUN` and the item "net of the planned K edits is +3" | `spec_lint` now records the lint run by the lane after round 1 and the run in this round on spec.md v0.1.3 (`0 error(s), 0 warning(s)`, exit 0); the "+3" item carries a `[SUPERSEDED by correction round 1 …]` note; stub re-measured: real `OK 9999`, simulated post-run text `OK 9997` (no core-stub wording changed in round 2) |

Design choice stated, with its reason. The audit's wording was that the allowlist is "an artifact the run phase fills in". The manifest is instead written and frozen by this plan phase and the run phase never edits it, because a list the run phase writes for itself cannot be an independent witness of what the run phase did; the cost is that a wording change during the run is a blocker report for manager-spec (the same regime the pinned cells of AC-MJ-006, -007, -009 and -018 already have). The mid-run authority route of the SPEC artifact ownership section (D-NEW-1) is the amendment path.

Regression check of the closed defects D1-D15 after the edits (cross-layer sweep: every revised criterion's requirement and plan items were swept in the same pass — AC-MJ-001 against REQ-MJ-001 to REQ-MJ-004 and plan M1 step 1, AC-MJ-008 against REQ-MJ-010 and plan M2 steps 1 and 4, AC-MJ-011 against REQ-MJ-013 and plan §E, AC-MJ-020 against REQ-MJ-006, -007, -012, -017 and plan §A, §B, §E, §D):

- All 51 ledger entries were re-run on `d31fb64239a688c053e455000cc9fe37f540b790` (50 single-command cells through a comparison script and the seven `cmp` calls of E-011 one by one): 0 mismatches between recorded and observed stdout and exit code.
- Trace verb (a Python port of the Group 4 verb; the awk form is refused by the guard): 17 REQ definitions, 20 AC headers, no UNCOVERED and no ORPHAN requirement; classes R 12, G 7, O 1. Counts agree across spec.md (HISTORY 0.1.3), acceptance.md, plan.md, design.md D-1, decision-index.md Q6 and this record: REQ 17 (ceiling 25), AC 20 (ceiling 25).
- D1 (commit shape): plan §E, acceptance.md §4 and AC-MJ-011 state the same graph-verifiable shape; the manifest is frozen in the plan-phase commit and never edited by the run phase, so it adds no run-phase commit and no ordering clause. D2 and D7: E-006 still prints `8 8 1 1` on the simulation. D3: every one of the ten cron sites now has a check (K:28, F:119, F:122, A intro, A fire-and-renew (e), A delete (d), A §14, W:244, M:26 (f), M's subsection by AC-MJ-007). D4, D5, D6, D8, D9: cells unchanged and re-run (0 mismatches). D10-D15: rows and quotations untouched.

### Ceiling-exception round (after plan audit iteration 2, FAIL 0.88, defects D23 blocking, D25 SHOULD-FIX, D24 MINOR; operator record d-20261019T062011Z-8108)

The iteration 2 report is `.moai/reports/t1894/plan-audit-iter2.md` (receipt rcpt-55ffb196ab252308a7c4a3be; the required codex gate failed on D23; D1-D22 closed). The operator record granted exactly this round: fix D23, D25 and D24, then one final plan audit (codex required). Every cell below was re-run in this round on HEAD `f80aef625e` with the SPEC-directory edits uncommitted in the working tree; the sixteen documents are byte-identical to the round-2 pin. Scratch probes lived under `/private/tmp/d25probe/` and were removed after the measurements.

| Defect | Before (iteration 2 audit) | After (this round) |
|---|---|---|
| D23 (blocking) the path half of AC-MJ-020 exempts all of `.moai/` | `cmd_paths` with a fixed diff list: `.moai/policies/workflow/factory-dispatch-detail.md`, `.moai/workflows/gtd.md`, `.moai/docs/x.md`, `.moai/config/sections/workflow.yaml` each printed `unexpected-paths 0`, exit 0, while four SPEC sentences claim a projection edit is caught | the exemption narrows to the four run-writable prefixes in `cmd_paths` (closed-world-check.py:96-100). Measured after: `.moai/policies/workflow/factory-dispatch-detail.md` → `unexpected-paths 1` + the path line, exit 1 (new cell E-020f); `.moai/workflows/gtd.md` → `unexpected-paths 1` + the path line, exit 1 (new cell E-020g); simulated paths under the four writable prefixes (`.moai/specs/SPEC-CARD-MODE-JUDGMENT-001/`, `.moai/reports/`, `.moai/state/`, `.moai/logs/`) → `unexpected-paths 0`, exit 0; `AGENTS.md` → `unexpected-paths 1`, exit 1. E-020b on the real tree still `unexpected-paths 0`, exit 0 (no untracked non-ignored `.moai/` file exists and no tracked `.moai/` path outside the four prefixes changed against the base). E-020c re-measured: `unexpected-paths 200`, exit 1 (was 175 under the whole-`.moai/` exemption). The four coverage statements (acceptance.md AC-MJ-020 title, AC-MJ-006 Evades paragraph, spec.md R-8, design.md D-10 row) were re-verified against the implemented code and hold as written; none was rewritten. The When/Then clauses of AC-MJ-020 and the design.md D-11 bullets now describe the narrowed behavior. E-020d re-pinned: the check's SHA-256 is `4b7586be370413555fa409542a93f8fc443f8e5fdcc13a5d581647610fa32514`, the manifest's unchanged — `same same`, exit 0; E-020e `DIFFERENT same`, exit 1 |
| D25 (SHOULD-FIX) the order-blind line comparison misses a moved code fence | the fence-move mutant (the existing closing fence of C moved behind the new Mode judgment paragraph) printed `0 0 0 0`, exit 0, and E-001 still `8`/`8` | `lines` prints a fifth number per side: code-fence lines whose nearest preceding non-blank line, walking past the manifest's listed rows, differs from the base. A plain fence count was measured blind to the mutant (2 → 2 on the rebuilt copy), so the fifth number compares fence context, not count; listed rows added or removed anywhere leave every walked predecessor unchanged, so correct work prints 0 whatever position the rows land in. Measured: copies of all sixteen files rebuilt from the manifest rows alone, listed rows at a non-planned position, print `0 0 0 0 0` on every row, exit 0 (green path); the fence-move mutant tree prints `C template 0 0 0 0 1`, exit 1; a stem-less addition re-run through the five-number checker prints `D template 1 0 0 0 0`, exit 1 (the four-number semantics unchanged). E-020a re-recorded in the five-number form (fifths 0, seconds and fourths as before, exit 1); mutant count 74 → 75, all red; design.md D-10/D-11 and the AC prose state the moved-fence coverage and its non-fence limit |
| D24 (MINOR) E-013c dead under the lane's shell-function grep | E-013c printed `0`/`0`/`0`, exit 1 under ugrep 7.8.4 and `2` `2` `1`, exit 0 under `/usr/bin/grep` | §0 states the judging grep: `/usr/bin/grep` (BSD 2.6.0-FreeBSD) judges every cell; the shell-function grep (ugrep) mishandles the five-way alternation of E-013c, and a cell the two engines disagree on is judged by `/usr/bin/grep`. Re-measured this round: E-013a eight `:0` exit 1 and E-013b `8` exit 0 (both engines agreed); E-013c `2` `2` `1`, exit 0 |

Regression of the closed defects D1-D22, re-run in this round, each matching its pinned expectation under `/usr/bin/grep`: E-001 `0`/`0` exit 1; E-004c and E-004e `:0` ×2 exit 1; E-005d `:0` ×2 exit 1; E-006 `0 0 0 1` ×2 exit 1; E-007 `0 0 0` ×2 exit 1; E-008e `0`/`0` exit 1; E-008f `0`/`0` exit 1; E-009 `0 1 0` ×2 exit 1; E-011g `pairs 0 unpaired 7` exit 1 and E-011h `pairs 2 unpaired 0 expected 2` exit 0 (through the changed checker; the `pairs` mode is untouched); E-013a/b/c as above; the Q7 gate-statement identity check: the sentence occurs once in spec.md, once in plan.md, once in acceptance.md and twice in decision-index.md, exit 0; the AC-MJ-020 first-number red re-observed through the new checker on a stem-less addition mutant. `moai spec lint --strict .moai/specs/SPEC-CARD-MODE-JUDGMENT-001/spec.md` on spec.md v0.1.4: `0 error(s), 0 warning(s)`, one INFO `OwnershipTransitionUnmeasured`, exit 0.

Gaps of this round: (G-17) the plan-artifact hash was not recomputed by this agent (no audit_cache tool); the lane recomputes it after the correction commit. (G-18) the negative-control cells E-020f/E-020g drive `cmd_paths` through the explicit path list after `--`, not through a real edit to a tracked projection (a real edit would dirty the shared tree); the git-query path of `cmd_paths` is exercised by E-020b and E-020c, and the narrowed predicate is the same line for both inputs. (G-19) the fifth number's green path was observed on copies of all sixteen files rebuilt from the manifest rows, not on the plan-phase simulation of the planned wording (which lives outside the repository, G-15); the rebuilt copies place the listed rows at a non-planned position, which the fifth number is invariant to by construction.

### Correction round 1 (after plan audit iteration 0, FAIL 0.65, defects D1-D9)

Each defect, its evidence command before and after (all cells re-run on `b7b44433c3` except where stated; the "after" values were observed on a scratch simulation of the post-run text, G-15, and on the real tree for the RED-now cells):

| Defect | Before | After |
|---|---|---|
| D1 commit ordering (Template-First unverifiable in one commit) | REQ-MJ-013 claimed an edit order the commit graph cannot witness | REQ-MJ-013 claims only per-milestone-commit coherence; both stub edits, catalog hashes and ledger rows ride the commit that changes what they pin; AC-MJ-011 adds a `git diff-tree` check (run-phase) |
| D2 AC-MJ-006 scan | printed `9 9` on the good copy and on all three escape mutants | bounded by the next `^## `, end marker required, 8 whole lines pinned, stem scan outside the subsection; prints `8 8 1 1`; 16 of 17 mutants red (N9, a stem-less Routing widening, is caught by AC-MJ-005) |
| D3 unscoped cron statements | 7 unscoped keep/arm lines per tree (K:28, F:119, F:122, A:117, A:118, A:396, W:244) | ten sites enumerated (research.md §11, REQ-MJ-010); AC-MJ-008 sweep (a) + per-site fragment counts K1/F2/A2/W1 + renewal wording + `when the lane stops executing the card` |
| D4 predicate narrowing | `shared state` only in the spec.md quote | REQ-MJ-002/-003 and the AC-MJ-001 token restored to "same files or shared state" (6 tokens; 8 since correction round 2, D20) |
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
5. **Stub budget re-measured:** `OK 9999`, template and deployed stubs `cmp`-equal (research.md §8). [SUPERSEDED by correction round 1: this item once read "the net of the planned K edits is +3, so a ≥ 2-unit offset is required". The planned K edits (K:28 tail +4, K:20 clause −6) net −2, measured 9,997 on the simulated text (design.md §D-4, plan M2 step 2, AC-MJ-012); no offset trim is planned. Re-measured in correction round 2: the real stub `OK 9999`, the simulated stub `OK 9997`.]

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
- **G-17** (correction round 2) No Go command ran. The manifest rows are the planned post-run text; what the Go guards (`TestBindingLedgerIntegrity`, `TestRoleCoreStubGate`, the doctrine and mirror-parity tests) make of them was read from source, never executed (same as G-2). The simulated stub measures 9,997 units; the Go gate renders the template first.
- **G-18** (correction round 2) The worktree guard refused one command in this round: a Python program fed through a heredoc whose text named git commands ("a worktree-isolated session's git operations must target its own worktree"). It was re-issued as separate edits with the editing tool; no result was inferred from the refused command (`verification-claim-integrity.md` §3.1). The new python cells call git inside the program and were accepted as plain invocations.
- **G-19** (correction round 2) The mutant harnesses, the simulation (`sim`) and the 74 mutant trees live in the session scratchpad and are not cited as evidence; they are a known loss. The mutant counts and numbers in `acceptance.md` come from runs of those harnesses on this tree and are reproducible from the manifest and the planned text; the RED-now cells run on the real tree.
- **G-20** (correction round 2) The `pairs` instrument was observed on two historical commits (a deployed-only commit and a two-pair commit). No milestone commit exists, so the run-phase obligation of AC-MJ-011 is not recorded as a pass.
- **G-21** (correction round 2) `closed-world-check.py` compares the working tree with the base commit (an uncommitted edit is seen), compares non-blank lines as multisets (order and blank lines are not seen), and does not judge a listed line. The `paths` mode reads `git diff --name-only <base>` and the untracked files; it does not see an ignored file.
- **G-22** (correction round 2) The 51-entry ledger comparison used the verbatim stdout and the exit code of each command run through `sh -c` from a script; only the cells that were new or rewritten in this round (E-001, E-002a, E-006, E-008e, E-008f, E-009 and the python cells) were also run directly through the worktree guard. The `moai spec lint` result is judged by the installed build, which predates HEAD (G-14).

### Residual risk

- The manifest rows are wording no command can judge, and the run phase may not change them; a needed wording change is a blocker report for manager-spec (accepted friction, round 2). A paraphrase hidden inside a listed line, a line moved without change and a blank-line change are not seen by AC-MJ-020.
- The mode judgment is text a human or an LLM writes; no command here proves a stated file set is the real one (spec R-2).
- Until the Go SessionStart notice changes (Q4), a lane reads two instructions about the standing cron (spec R-1).
- The ruling d314 names no mechanism that establishes "the operator has delegated dispatch" (spec R-5). AC-MJ-006 pins the text of the exception, not whether an actor treats an unstated delegation as given.
- AC-MJ-006 pins D's wording line by line; any wording improvement during the run phase must amend the criterion with it. A pure paraphrase of the exception placed in a second file evades AC-MJ-016 / AC-MJ-018 (verbatim tokens and whole lines); the sync audit reads for it.
- Three Go doctrine-test families and one ledger row pin the old operator-only sentences; the plan keeps them verbatim and adds the exception beside them, so a reader meets both an absolute sentence and a pointer to its one stated exception. That is deliberate (no Go change in this card) and is the price of "each rule stated once".
- The `internal/cli` Go sources (`todo.go:6`, `:238`, `:1576`) still say the next pick is the operator's act; they describe the operator's own `next` pick, which the ruling does not change.
- A paraphrased restatement of a rule in a second file evades AC-MJ-016; the sync audit reads for it.

## §E.2 Run-phase Evidence

Run-phase status: **complete** (AC-MJ-014 was blocked at M5 on `770b5ae654`, resolved by the mid-run amendment `8b9f610b91` and the correction `05bcdc3a8f`; see the verdict §6). A second correction, `45270e3ef3` (D transition boundary, after sync-audit finding S1 and the amendment `0fca8e0e90`), follows; see the verdict §7. A third, evidence-only step (verdict §8) re-ran the ledger after the amendment `80a7dc041b` that fixed the AC-MJ-020 (b) scope. Measured on HEAD `80a7dc041b`, clean tree; the Go results were produced at `45270e3ef3` and no Go or template file changed since. The commands are the single-line cells of `acceptance.md` §2 (ids in the verdict); the decisive output lines are in `.moai/reports/t1894/verdict.md` §2, which this table cites.

| AC | Status | Actual Output (decisive) |
|----|--------|--------------------------|
| AC-MJ-001 | PASS | `8` / `8` |
| AC-MJ-002 | PASS | E-002a empty exit 0; E-002d `:1` / `:1` |
| AC-MJ-003 | PASS | `:1` / `:1` |
| AC-MJ-004 | PASS | (a) `:1` (b) `:0` (c) `:2` (d) `:1` (e) `:2`, both trees |
| AC-MJ-005 | PASS | (a) `:1` (b) `:0` (c) `:1` (d) `:3`, both trees; (e) E-005e `0 1` / `0 1` |
| AC-MJ-006 | PASS | `8 8 1 1` / `8 8 1 1` |
| AC-MJ-007 | PASS | `4 4 1` / `4 4 1` |
| AC-MJ-008 | PASS | (a) empty; (b) core 1, leader body 2, doctrine 2, watchdog 1; (d) 1; (e) 2; (f) 1 |
| AC-MJ-009 | PASS | `1 1 0` / `1 1 0` |
| AC-MJ-010 | PASS | (a) empty; (b) `:1` / `:1` |
| AC-MJ-011 | PASS | seven `cmp` exit 0; `pairs 3/5/4/1/1 unpaired 0` for `2fb0be5ed8`, `89eb51de2f`, `770b5ae654`, `05bcdc3a8f`, `45270e3ef3` |
| AC-MJ-012 | PASS | `OK 9997` |
| AC-MJ-013 | PASS | `:0` on all eight template files |
| AC-MJ-014 | PASS | `ok  	github.com/modu-ai/moai-adk/internal/template	6.155s` (14 of 14 named guards, `TestBindingLedgerIntegrity` PASS, `violations=0`); `ok  	github.com/modu-ai/moai-adk/internal/cli	0.926s` (6 of 6); no `[no tests to run]` in either log |
| AC-MJ-015 | not mine | recorded by the leader |
| AC-MJ-016 | PASS | the three homes: cards, detail, mechanics |
| AC-MJ-017 | PASS | ` 1 file changed, 1 insertion(+), 1 deletion(-)` |
| AC-MJ-018 | PASS | (a) `:1`/`:1` (b) `:0` x4 (c) `:1` x4 (d) `:1`/`:1` (e) `:1`/`:1` |
| AC-MJ-019 | PASS | `:1` / `:1` |
| AC-MJ-020 | PASS | (a) E-020a and E-020i: all sixteen rows `0 0 0 0 0`; (b) E-020b `unexpected-paths 0`, E-020j `unexpected-paths 0`, controls E-020c `unexpected-paths 199`, E-020f/g `unexpected-paths 1`, E-020k `unexpected-paths 1` (`CHANGELOG.md.bak`), E-020l `unexpected-paths 1` (`docs/CHANGELOG.md`); (c) E-020d `same same`, E-020e `DIFFERENT same`; E-020h `C good 0 0 pair-move 44 44`. The earlier red (E-020b `unexpected-paths 1` = the sync-phase `CHANGELOG.md`, audit S2) is resolved by amendment `80a7dc041b` |

| Invariant | Status | Actual Output |
|-----------|--------|---------------|
| Stub budget | PASS | `OK 9997` (limit 10,000; was 9,999) |
| Ledger rows ALB-0276 / ALB-0257 updated in M2 | applied | rows rewritten in `89eb51de2f`; the correction `05bcdc3a8f` changed no ledger or catalog file |
| Catalog hashes | applied | `moai` in M1 and M3, `moai-lane-watchdog` in M2; `TestCatalogHashCoversSkillSubfiles` PASS |
| Agents, Go source, other SPEC dirs untouched | PASS | `git diff --name-only` of this run's commits names only the 16 documents, `catalog.yaml`, `binding_ledger.json` and SPEC-directory files; the `paths` check reads `unexpected-paths 0` at `80a7dc041b` (the checker now admits the exact repo-root `CHANGELOG.md`) |

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-10-23
run_commit_sha: 45270e3ef3      # last run-phase document commit (the D correction); the evidence commit that carries this block follows it
run_status: complete
ac_pass_count: 19               # AC-MJ-001..014, 016..020, all verified at HEAD 80a7dc041b (Go results from 45270e3ef3, no Go or template file changed since)
ac_fail_count: 0
open_finding: none              # AC-MJ-020 (b) / audit S2 resolved by amendment 80a7dc041b; verdict.md §8
ac_not_owned: 1                 # AC-MJ-015, the leader's
preserve_list_post_run_count: not-measured
l44_pre_commit_fetch: not-run
l44_post_push_fetch: not-run    # nothing was pushed
new_warnings_or_lints_introduced: not-measured   # documents only; no lint run
cross_platform_build:
  linux_darwin: not-applicable  # no Go source changed
  windows: not-applicable
total_run_phase_files: 19       # product files: 8 templates, 8 deployed, catalog.yaml, binding_ledger.json, spec.md (the correction re-touched two of the 16 documents)
m1_to_mN_commit_strategy: "M1 2fb0be5ed8, M2 89eb51de2f, M3 770b5ae654, one commit each, template and deployed twins together; M4 no commit; M5 green after the correction 05bcdc3a8f (mid-run amendment 8b9f610b91)"
blocker: none                   # resolved; verdict.md §6
sync_heads_up: "acceptance.md holds 21 distinct AC-looking identifiers (AC-ALB-021 is named in an edge case); the SPEC's own count stays 20 (AC-MJ-001..020)"
```

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: 2026-10-23
sync_commit_sha: pending-backfill   # corrective sync commit; the first sync commit ea7b93c085 was audited and failed (sync-audit.md, 0.80)
sync_status: complete   # corrective sync after the iter-1 audit FAIL
ac_source: .moai/specs/SPEC-CARD-MODE-JUDGMENT-001/acceptance.md   # tier L
b12_self_test_a: pass   # pre-emission grep -c 'SPEC-CARD-MODE-JUDGMENT-001' CHANGELOG.md printed 0 before the entry was written
b12_self_test_b: pass   # REQ-MJ-001..017 = 17 distinct ids in spec.md; AC-MJ-001..020 = 20 distinct ids in acceptance.md; the entry states 17 and 20 (AC-MJ-015 named as the leader's criterion). The default-prefix AC- sweep reads 21 only because an edge case cites AC-ALB-021; not used for the count
b12_self_test_c: pass   # every path named in the entry was checked with ls (8 template documents and their deployed twins, catalog.yaml, binding_ledger.json, verdict.md)
changelog_entry_position: "[Unreleased] > Added, first entry"
frontmatter_status_transitions:
  spec_md: "in-progress -> implemented -> completed, on this single sync commit; updated: 2026-10-23"
  plan_md: "no frontmatter block; nothing refreshed"
  acceptance_md: "no frontmatter block; nothing refreshed"
canary_compliance_check: not-applicable   # this SPEC defines no forward-looking policy that its own sync tests
mx_tag_validation: "not-applicable - documents only, no source file changed, no @MX tag added or changed"
codemaps_readme_docs_site: "deliberately not touched - no user-facing feature and no code changed"
corrective_sync:
  s1: "fixed by the run phase in 45270e3ef3 (D transition boundary follows the free-lane definition), after SPEC amendments 8b9f610b91, 0fca8e0e90, 80a7dc041b"
  s2: "resolved - AC-MJ-020(b) is green with the sync CHANGELOG.md admitted by SPEC amendment 80a7dc041b; run evidence ff41cde331 shows all 19 owned ACs green, E-020b unexpected-paths 0"
  s3: "corrected - the CHANGELOG entry no longer claims a mode-judgment pointer in factory-dispatch-core.md; it names the pointers in factory-dispatch.md and the gtd workflow, the core stub's reconciliation clause and cron-scope tail, and the three mid-run amendments; counts stay 17 REQ and 20 AC-MJ"
  s4: "left as the pre-disclosed R-9 operator-held residual (polling-cron wording versus auto-semantics section 14)"
  s5: "minor, accepted"
evidence: .moai/reports/t1894/verdict.md
gaps:
  - "AC-MJ-015 (lane acknowledgement) is the leader's and is not recorded here"
  - "sync-auditor verdict not part of this commit; the orchestrator owns the sync audit"
  - "repository-wide CI verdict pending; only the 14 + 6 named Go guards were run in the run phase"
```

## §F Phase 4 Mode Selection

- Inputs: tier L (docs-only, Template-First); 16 files (8 template + 8 deployed twins) across milestones M1-M5; domains: rules / skills / agents docs (markdown), 1 Python checker (unchanged); coding-heavy: no; concurrency benefit: low (twins must stay byte-identical, one writer per tree).
- Evaluation: direct - not selected (not trivial); serial - selected; fanout - not selected (not multi-domain research); sweep - not selected (not >= ~30 files, edits are not one uniform rule).
- Decision: serial
- Justification: one `manager-develop` (cycle_type=tdd, effort high per Class C) works milestone by milestone in the card worktree; the listed manifest rows are the contract, the closed-world checker measures the result. Kickoff met in its autonomous form on 2026-10-23: plan-audit-iter5 PASS 0.91 (rcpt-e6173f9d83a33a2cf3be2185, hash 68c195f0..., audited 21974dddf2), Q7 answered (leader-relayed), `moai factory decide --gate kickoff --choice approve --decider audit` recorded (card v13, state run).

## §G Override and Refusal Record

- 2026-10-22T18:54:09Z SPEC-CARD-MODE-JUDGMENT-001 ceiling-refusal outcome=hold reasons="plan-audit ceiling reached (round count 6 >= tier ceiling 3); the verdict matches no admitting arm and holds, entry blocked (REQ-ACE-006) — release path: the split/new-SPEC route of REQ-ACE-005 or an operator decision recorded in progress.md §G" evidence=/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1894/.moai/reports/t1894/plan-audit-iter1.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1894/.moai/reports/t1894/plan-audit-iter2.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1894/.moai/reports/t1894/plan-audit-iter3.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1894/.moai/reports/t1894/plan-audit-iter4.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1894/.moai/reports/t1894/plan-audit-iter5.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1894/.moai/reports/t1894/plan-audit.md
- 2026-10-22T18:57:41Z SPEC-CARD-MODE-JUDGMENT-001 ceiling-outcome outcome=pass-through reasons="plan-audit ceiling reached (round count 6 >= tier ceiling 3 + 1 delta rounds); the verdict is admission-clean and admits without a question (REQ-ACE-013)" evidence=/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1894/.moai/reports/t1894/plan-audit-iter1.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1894/.moai/reports/t1894/plan-audit-iter2.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1894/.moai/reports/t1894/plan-audit-iter3.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1894/.moai/reports/t1894/plan-audit-iter4.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1894/.moai/reports/t1894/plan-audit-iter5.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1894/.moai/reports/t1894/plan-audit.md
- 2026-10-22T18:58:01Z SPEC-CARD-MODE-JUDGMENT-001 ceiling-outcome outcome=pass-through reasons="plan-audit ceiling reached (round count 6 >= tier ceiling 3 + 1 delta rounds); the verdict is admission-clean and admits without a question (REQ-ACE-013)" evidence=/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1894/.moai/reports/t1894/plan-audit-iter1.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1894/.moai/reports/t1894/plan-audit-iter2.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1894/.moai/reports/t1894/plan-audit-iter3.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1894/.moai/reports/t1894/plan-audit-iter4.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1894/.moai/reports/t1894/plan-audit-iter5.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1894/.moai/reports/t1894/plan-audit.md
