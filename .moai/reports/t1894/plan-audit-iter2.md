auditor-model: claude-sonnet-5-5
verdict: FAIL
overall_score: 0.88
must_pass_failed: 0
blocking_count: 1
plan_artifact_hash: 5bda514f1639f572749e57a956229ae8b5efb0ce579fc74206681f723ba2bfca
audited_sha: cb71f2b475263dc8b0921e37836dcc8208a62f71
required_backend: codex fail
convergence_overall: fail
scope: delta
fix_scope: acceptance.md#AC-MJ-020 acceptance.md#E-020b acceptance.md#E-020d acceptance.md#AC-MJ-006 spec.md#7-known-limits-and-residual-risk design.md#D-10 design.md#D-11 closed-world-check.py#cmd_paths
defect_class: D23=design

# SPEC Review Report: SPEC-CARD-MODE-JUDGMENT-001 (card t1894), iteration 2 (final delta audit after correction round 2)

Reasoning context ignored per M1 Context Isolation: only the SPEC directory artifacts (spec, plan, acceptance, design, research, decision-index, progress, the two new helper files) and the iteration 0 and iteration 1 reports were read. Tier L (spec.md:L14), PASS threshold 0.85 (spec-workflow.md tier table), ceilings 25 REQ / 25 AC (17 REQ, 20 AC present).

Judging builds. Tree: HEAD cb71f2b475263dc8b0921e37836dcc8208a62f71 (`git status --short` printed nothing; branch WT-fail-safe-serial, unpushed). Every `moai` CLI observation (decision read, spec lint, verify audit-plan) was produced by the installed build `moai_cp/20261007_195927`, which the audit_multi server reports as an ancestor of this HEAD (server build commit 7705b365e2; binary lag: CLI output describes older code). The `grep` of the Claude Code Bash tool is a shell function over ugrep 7.8.4; `/usr/bin/grep` is BSD grep 2.6.0-FreeBSD (used where stated). No Go command, build or test was run. The documents under the SPEC (`internal`, `.claude`, `.agents`, `.codex`, `AGENTS.md`) are byte-unchanged between the iteration 1 audited commit 4834700468 and this HEAD (`git diff --stat 4834700468ed2311e3efc4e12bb920db03db54e2 HEAD -- internal .claude .agents .codex AGENTS.md` printed nothing); only `.moai/` artifacts changed.

## Result summary

The correction round is real work. D17 through D22 are closed, the 17 REQ / 20 AC traceability is clean, the awk cells now pass the worktree guard, the Q7 gate statement is identical in four documents, every RED-now cell reproduces, and the closed-world check does close the defect it was written for: the four stem-less added-line mutants of the iteration 1 report are red in AC-MJ-020. The SPEC still fails, for two reasons that are independent of each other. (1) The required codex gate failed (receipt below) on a P1 that I reproduced independently: the path half of AC-MJ-020 exempts everything under `.moai/`, so a widening placed in a tracked `.moai/policies/` copy of the very documents the SPEC changes passes every release-blocking criterion, while four sentences of the SPEC claim the opposite (D23, blocking). (2) The order-blind line comparison admits a structural mutant (D25, optional) that codex also reproduced. The aggregate is above the Tier L threshold and higher than the iteration 1 aggregate (no STOP signal), but a failed required gate and one blocking defect decide the outcome.

## Must-Pass Results

- [PASS] MP-1 REQ numbers. `python3 -I` port of the Group 4 verb printed `COLLECTED: 17 REQ definitions`; `grep -o` over spec.md lists REQ-MJ-001..017 in order, no gap, no duplicate (spec.md:L75-L103).
- [PASS] MP-2 GEARS, judged on the REQ layer of spec.md only (the ACs are Given-When-Then in acceptance.md and were graded under Testability). Every REQ-MJ entry is a shall clause in a Ubiquitous / While / When / Where form (spec.md:L75-L103); REQ-MJ-004 is `When ... shall` throughout (spec.md:L78).
- [PASS] MP-3 frontmatter. 12 canonical fields with correct types (spec.md:L2-L13); `version: "0.1.3"` quoted; `status: draft`; `tier`, `related_specs` are outside the 12-field table. `moai spec lint --strict` on spec.md printed `0 error(s), 0 warning(s)` plus one INFO `OwnershipTransitionUnmeasured`, exit 0.
- [N/A] MP-4. The SPEC edits repository process documents; no language tool is named.
- [PASS] MP-5 D7. Referenced SPECs: ALWAYS-LOADED-BUDGET, FACTORY-LEASE-RENEW, ROLE-INJECTION-BUDGET, TODO-CLASSIFY-DISPATCH all exist with status completed. SPEC-FACTORY-DEFAULT-MODE-001 is not found in `.moai/specs/`; it occurs only inside the verbatim 02f4 board quotation (spec.md:L47) and names card t1737's future SPEC: SHOULD severity under D7-5 (D27), not BLOCKING (no retired or superseded SPEC is referenced).
- [PASS] MP-6 D8. `grep -c syscall spec.md` printed `0`.
- [PASS] MP-7. `grep -rn -E 'NEEDS CLARIFICATION' plan.md research.md` printed nothing, exit 1.
- [PASS] MP-8 RED-now re-execution. Every release-blocking criterion (AC-MJ-001..009, -016, -018, -020) has RED-now cells; all were re-run on this tree and reproduced stdout and exit code (table below). The three cells refused by the guard in the iteration 1 audit (E-002a, E-006, E-009) now run as written. No cell failed to reproduce its RED.
- [PASS] MP-9 cross-artifact ordering. The CN-4 awk verb is refused by the guard (a `|` after printf somewhere in its text); a python port printed `COLLECTED: 5 milestones in plan order (M1 M2 M3 M4 M5), 0 exit bindings, 44 ordering candidates`, no CONFLICT. With zero `Exit:` bindings no mechanical CONFLICT can exist, so all 44 candidates were read: those naming a milestone (acceptance.md:L51 `M1 adds one new line directly after line 30`, L224 `M2 edits K ... M3 edits F`, L250, L777) order lines inside a file or state a commit shape that plan.md section E and section C state the same way. The iteration 0 conflict pair stays gone. Run in full this iteration, as the Retry Loop Contract requires.

## Category Scores

| Dimension | Score | Rubric band | Evidence |
|---|---|---|---|
| Clarity | 0.85 | 0.75-1.0 | the Q7 gate is one sentence, identical in spec.md:L142, plan.md:L71, acceptance.md:L776, decision-index.md:L84/L98; REQ-MJ-007 lists exactly seven limits (spec.md:L84); residual ambiguity: four statements about what AC-MJ-020 covers overreach (D23) |
| Completeness | 0.90 | 0.75-1.0 | all sections and 12 fields present; Out of Scope has three H3 groups with specific bullets (spec.md:L107, L114, L122); ten cron sites each with a check (research.md section 11); helper files carry a usage docstring |
| Testability | 0.80 | 0.75 | every AC binary-testable and RED cells reproduce; the closed-world check removed the iteration 1 mutants; remaining: path exemption (D23), order blindness (D25), one environment-dependent control cell (D24) |
| Traceability | 0.95 | 1.0 band, one indirect mapping | 17 collected, no UNCOVERED, no ORPHAN, 20 AC headers each naming full REQ ids; indirect: AC-MJ-020 cites REQ-MJ-006/-007/-012/-017 though it bounds all eight documents |

Aggregate (mean of four) 0.875, written as 0.88: above the Tier L threshold 0.85 and above the iteration 1 aggregate, so no STOP. It does not offset the failed required backend or the blocking defect (Must-Pass Firewall and the gate rule are independent of the aggregate).

## Closure table, D1 to D22 (each row re-executed on HEAD cb71f2b475)

| Id | Status | Evidence (before = earlier reports, after = this run) |
|---|---|---|
| D1 MP-9 commit shape | closed | CN-4 python port: 0 exit bindings, no CONFLICT; plan.md:L93 and acceptance.md:L777 state the same graph-verifiable shape; `pairs` command observed on known inputs (E-011g `9579676013 pairs 0 unpaired 7`, exit 1; E-011h `1969709ce6 pairs 2 unpaired 0 expected 2`, exit 0) |
| D2 AC-MJ-006 scan | closed | the 22 earlier mutant files re-scanned with the E-006 command: good `8 8 1 1`; held_removed, nocond, serial_promoted, trailing_space, appended_line1, consequence_relabelled `7 8 1 1`; dup_line, extra_bullet `8 9 1 1`; h3/h4 widening `8 10 1 1`; routing_parent_stem, routing_sub_stem `8 8 1 2`; extra_h2_before_review `8 8 0 2`; endmarker_renamed `8 8 0 1`; heading_renamed `0 0 0 6`; heading_deleted `0 0 0 5`; the five stem-less files still `8 8 1 1` (now red in AC-MJ-020, see D16) |
| D3 cron sites | closed | ten sites: K:28, F:119, F:122, A intro, A fire-and-renew, A delete, A section 14 carrier, W:244, M:26, M Lease renewal; `grep -r -n` for arm/keep statements over `.claude/rules` and `.claude/skills` returns 8 hits, all inside that set or the two name-only sentences (A:103, A:419 version line); E-008a prints the 14 lines (7 per tree) |
| D4 predicate narrowing | closed | E-001 `0` `0` exit 1 with the eight-token list including `same files or shared state` |
| D5 traceability | closed | python port of the Group 4 verb: `COLLECTED: 17 REQ definitions`, no UNCOVERED, no ORPHAN |
| D6 shallow mutants | closed | E-004c, E-004e `:0`/`:0` exit 1; E-008b eight `:0` exit 1; E-009 `0 1 0` x2 exit 1 |
| D7 REQ-MJ-007 additions | closed | the eight E-006 pinned lines are rows of the manifest (checked mechanically, below); d314 body is a substring of spec.md and decision-index.md |
| D8 leader-duty enumeration | closed | E-018e `:0`/`:0` exit 1; the pinned row is a manifest row of M |
| D9 AC-MJ-013 wrong-reason red | closed | E-013a eight `:0` exit 1; E-013b `8` exit 0 under both greps; probe file: 21 of 30 lines matched under both greps; see D24 for the control cell E-013c |
| D10 Q7 row | closed | decision-index.md:L77-L84 |
| D11 Q6 label | closed | decision-index.md:L67-L75 |
| D12 Q2 and Q5 consequences | closed | decision-index.md:L26, L65; spec.md R-7, R-2 |
| D13 absolute sentence plus exception | closed (accepted as disclosed) | spec.md R-6 (L151) |
| D14 portability | closed | no PCRE, no `\b` in the cells; E-012b `OVER 25803` exit 1 |
| D15 housekeeping | closed | no card key, 02f4 verbatim (board quote check below) |
| D16 stem-less added lines | closed for the four reported mutants and for every added or removed line in the sixteen files; residual under D23 (paths outside the sixteen files) and D25 (order) | before: E-005d `3` and E-006 `8 8 1 1`, exit 0 on four stem-less additions. After: AC-MJ-020 first number `1` on those additions (rebuilt as `extra_both_sides`, `text_in_blank_line`, `append_to_listed_line`, below); the checker, the manifest (73 `+`/`-` rows by the author's count: 48 added, 25 removed lines) and the pinned cells are consistent (section below) |
| D17 two cron sites unchecked | closed | E-008e `0`/`0` exit 1 (A fire-and-renew two whole lines) and E-008f `0`/`0` exit 1 (M:26 opening) reproduce RED; both pinned lines are manifest rows; research.md section 11 rows 5 and 9 now name them |
| D18 three cells refused | closed | E-002a printed the `gtd.md:52` line, exit 1; E-006 `0 0 0 1` x2, exit 1; E-009 `0 1 0` x2, exit 1; each run as written through the guard, none refused |
| D19 Q7 gate stated two ways | closed | identical sentence in spec.md:L142, plan.md:L71, acceptance.md:L776, decision-index.md:L84 and L98; no board record answers Q7 (board grep) |
| D20 reissue and named card tokens | closed | E-001 now counts eight tokens (`names the conflicting card`, `reissued with an explicit classification` added); the manifest C paragraph carries both |
| D21 commit-shape listing | closed | `pairs` mode decides by exit status and expected count; observed red (E-011g) and green (E-011h) |
| D22 stale progress lines | closed | progress.md:L14 records the lint run; the "+3" item carries a SUPERSEDED marker (progress.md:L75, L116) |

## Evidence of the closed-world check (D16, the main subject of this iteration)

The helper files. `closed-world-check.py` (138 lines) and `closed-world-manifest.txt` (113 lines, 8 `file:` rows, 48 `+` rows and 25 `-` rows by the author's count; I did not recount) were read in full. `lines` compares, per file and per side, the multiset of non-blank lines of the working copy with the multiset at the base commit d31fb64239a688c053e455000cc9fe37f540b790 and prints four numbers; `paths` lists changed or new paths outside `.moai/` that the manifest does not name; `pairs` decides the milestone commit shape.

Re-execution on the real tree: `lines` printed the eight rows of E-020a exactly (`C template 0 2 0 0 deployed 0 2 0 0` ... `W template 0 2 0 2 deployed 0 2 0 2`), exit 1; first and third numbers are 0 on all sixteen files, so the red is the unmade work and nothing else. `paths` printed `unexpected-paths 0`, exit 0; against the old base commit 3642a3a9b9cf101863656153d2d99b69fcb66a05 it printed `unexpected-paths 175`, exit 1 (E-020c).

Consistency of the manifest with the pinned cells (mechanical): `pin_vs_manifest` extracted the pinned strings of E-004c, E-004e, E-005d, E-006 (eight lines), E-007 (four lines), E-008e, E-009 and tested each against the `+` rows: all are exact whole-line rows (E-008f is a substring of a row, as designed). The planned core stub, rebuilt from the manifest K rows, measures 9,997 UTF-16 units against 9,999 today (`current 9999 planned 9997 delta -2`), the figure plan.md:L31 states.

Mutants run through the checker. I rebuilt the good copy from the manifest (base lines minus the `-` rows plus the `+` rows, `0 0 0 0` on all sixteen files, exit 0) and ran 21 mutants, both sides unless noted (results saved under the scratch directory `/private/tmp/lxq/aud/`, `cw_results.txt`):

| Mutant | Result |
|---|---|
| added line equal to a line already present elsewhere in the file (multiset) | red: D first number 1 |
| extra stem-less sentence on both sides | red: first number 1 |
| text placed in a blank-only line | red: first number 1 |
| text appended to a listed line | red: first and second numbers 1 |
| ZWSP-only line | red: first number 1 |
| trailing space on a base line | red: first and third numbers 1 |
| sentence smuggled behind a CR inside the heading line | red: first and second numbers 1 |
| deployed-only extra line | red: deployed side first number 1 |
| base line deleted and a listed line duplicated | red: first and third numbers 1 |
| Cyrillic homoglyph inside a pinned line | red: first and second numbers 1 |
| 17th file in the sixteen-file tree (lines mode) | not seen by `lines` by design; `paths` flags it (probe: manager-todo.md and AGENTS.md give exit 1) |
| listed heading moved to the top of the file; two listed bullets swapped; two base lines reordered; a Routing sub-bullet re-parented into another section | NOT seen: `0 0 0 0`, exit 0 (order is ignored; disclosed at acceptance.md:L245, L253) |
| NBSP-only line, tab-only line, CRLF over the whole file | not seen, harmless (blank by definition) |
| existing closing code fence of C moved behind the new Mode judgment subsection | NOT seen: `0 0 0 0`, exit 0, and E-001 still prints `8` `8` (D25; reproduced after codex reported it) |

The older 22-file AC-MJ-006 family and the four stem-less files are in the closure table. Judgment: the added-line and removed-line escape is closed inside the sixteen documents; what remains is outside that set (D23) and order (D25).

The author's stated deviation from the lane brief (the manifest is frozen in the plan phase instead of an allowlist the run phase fills in; a wording change in the run becomes a blocker report to manager-spec). Judged sound: an allowlist written by the party the check constrains proves nothing, and `verification-claim-integrity.md` section 2.3 asks for exactly this graph order (list first, lines later); the manifest is committed in the plan-phase correction commit that precedes any document commit. It does not make the run phase infeasible or the criterion unsatisfiable by correct work: a run that copies the manifest rows prints `0 0 0 0` by construction (my rebuilt good copy), the `+` rows are the very lines the pinned cells E-004c/e, E-005d, E-006, E-007, E-008e, E-009 name, `paths` has only the two named non-document files (`gen-catalog-hashes.go` writes only `catalog.yaml`, read at lines 211-273), and a necessary wording change has a defined owner and path (mid-run authority, manager-spec; the same blocker-report rule `spec-frontmatter-schema.md` states for run-phase body changes). Cost, not defect: the planned wording was never run through the Go guards (progress.md G-15), so a guard that rejects a planned line costs a blocker round trip; and if a landing race forces the lane to merge newer main into the card tree, `lines` and `paths` go red on lines that are not the lane's (recorded in Residual-risk).

Can the helper files change without the audit noticing? They are not in the plan-artifact hash subject set (`planArtifactNames` in `internal/runtime/audit_cache.go`: acceptance, decision-index, design, plan, research, spec, tasks). I reproduced the hash with a python port of `ComputeHash` (`name:len\0` + bytes + `\0` per present file): it printed `5bda514f1639f572749e57a956229ae8b5efb0ce579fc74206681f723ba2bfca`, equal to the lane's runtime value. A helper-only edit therefore leaves the hash and a cached PASS valid; it is noticed at run time, because acceptance.md E-020d pins both SHA-256 values and AC-MJ-020 (c) is part of an R criterion. Verified: E-020d printed `same same`, exit 0 (computed `c2b0ef17...2c08` for the manifest and `c82828d4...5b63` for the script, equal to the pinned values); E-020e (one hex digit changed) printed `DIFFERENT same`, exit 1. To change a helper and keep E-020d green the pin in acceptance.md must change, which changes the hash. design.md D-11 words this too broadly (D26).

## RED-now re-execution (MP-8)

Cells run from the worktree root on HEAD cb71f2b475; each stdout and exit code below equals the ledger entry.

| Cell | Observed | Match |
|---|---|---|
| E-001 | `0` `0`, exit 1 | yes |
| E-002a / b / c / d | gtd.md:52 line exit 1 / same line at the pinned old commit / six paths / `:0` x2 exit 1 | yes |
| E-003 | `:0` x2, exit 1 | yes |
| E-004a / b / c / d / e | `:0` exit 1 / `:1` exit 0 / `:0` exit 1 / `:1` exit 0 / `:0` exit 1 | yes |
| E-005a / b / c / d | `:0` exit 1 / `:1` exit 0 / `:0` exit 1 / `:0` exit 1 | yes |
| E-006 | `0 0 0 1` x2, exit 1 | yes |
| E-007 | `0 0 0` x2, exit 1 | yes |
| E-008a / b / c / d / e / f | 14 lines exit 1 / eight `:0` exit 1 / `:1` x4 exit 0 / `:0` x2 exit 1 / `0` `0` exit 1 / `0` `0` exit 1 | yes |
| E-009 | `0 1 0` x2, exit 1 | yes |
| E-010a / b / c | empty exit 1 / `:1` x2 exit 0 / `1` exit 0 | yes |
| E-011 x7, E-011f | seven `cmp` exit 0; factory-dispatch.md `differ: char 23127, line 222`, exit 1 | yes |
| E-011g / h | `pairs 0 unpaired 7` exit 1 / `pairs 2 unpaired 0 expected 2` exit 0 | yes |
| E-012a / b | `OK 9999` exit 0 / `OVER 25803` exit 1 | yes |
| E-013a / b | eight `:0` exit 1 / `8` exit 0 | yes |
| E-013c | `0` x3 exit 1 under the shell-function grep; `2` `2` `1` exit 0 under `/usr/bin/grep` | differs by grep (D24) |
| E-016 | empty, exit 1 | yes |
| E-017 | ` 1 file changed, 1 insertion(+), 1 deletion(-)` | yes |
| E-018a / b / c / d / e | `:0` x2 exit 1 / `:1` x4 exit 0 / `:0` x4 exit 1 / `:0` x2 exit 1 / `:0` x2 exit 1 | yes |
| E-019a / b / c | `:1` x2 exit 0 / `0` exit 1 / `:1` x2 exit 0 | yes |
| E-020a / b / c / d / e | eight rows exit 1 / `unexpected-paths 0` exit 0 / `unexpected-paths 175` exit 1 / `same same` exit 0 / `DIFFERENT same` exit 1 | yes |

Board quotations (`moai decision read` saved to a scratch file, exit 0, `board=ok unparseable=0 records=842`): d314 (board line 40-41), 0c51 (57), a672 (67), 02f4 (63) bodies are substrings of spec.md; d314 is also a substring of decision-index.md. No board record answers Q7 (grep for `delegated dispatch` and `t1894` finds only d314 and two unrelated lane-state rulings). Eight template/deployed pairs: seven `cmp` exit 0, factory-dispatch.md exit 1 with the single declared fork line (`git diff --no-index --shortstat`: 1 insertion, 1 deletion). The 20 named Go tests exist (`grep -r -l -E '^func <name>\('` found a file for each); ledger rows ALB-0254, -0256, -0257, -0265, -0268, -0276, -0304 exist in `binding_ledger.json`; `manager-todo.md` has no `classification-file`, `moai todo add` or `moai gtd add` (count 0, exit 1).

## DEFAULT-APPLIED stamps (item 3)

| Row | Class / stamp | Allowed for its class? | Silently decides an operator matter? |
|---|---|---|---|
| Q1 | implementation-level, DEFAULT-APPLIED | yes (preserves behavior; confirmed by d314 "manager-todo untouched") | no |
| Q2 | implementation-level, DEFAULT-APPLIED | yes for the scope question; the follow-on product question (mode of a standing-source card) is declared separate | no: leaves one known 0c51 site open, stated as the consequence (decision-index.md:L26, spec R-7), not hidden |
| Q3 | product-level, met by the operator record d314 | yes: real operator record, verbatim, re-read from the board | no |
| Q4 | implementation-level, DEFAULT-APPLIED | yes; d314 moves the notice to card t1695 | no |
| Q5 | implementation-level, DEFAULT-APPLIED | formally yes ("how consistently ... not what the documents must contain") | no, but the row's own consequence says whether a merge-ready card counts "decides most outcomes" (decision-index.md:L65), which strains the implementation-level class; it is disclosed (R-2). Optional observation, not a defect |
| Q6 | implementation-level, DEFAULT-APPLIED | yes; the 17 REQ / 20 AC exceed the Tier M ceilings, stated; no policy coverage claimed | no |
| Q7 | product-level, open, unstamped FOUNDER | correct class, correctly unstamped | no: blocks only the autonomous Kickoff form; the four documents say so identically. An open Q7 is not a defect and is not counted as one |

## Defects Found

D23. BLOCKING. The path half of AC-MJ-020 exempts everything under `.moai/`, so a widening placed in a tracked projection of the changed documents passes every release-blocking criterion, while the SPEC states it would be caught — closed-world-check.py:96 (`p not in allowed and not p.startswith('.moai/')`), acceptance.md:L239 (title: "and no other path changes"), acceptance.md:L92 (AC-MJ-006: "one written in any other file is an unexpected path there"), spec.md:L153 (R-8: "no other path may change"), design.md D-10 row "A changed path outside the eight documents" — evidence: `git ls-files` shows tracked copies `.moai/policies/workflow/factory-dispatch-detail.md` (which contains the old sentence `free means the lane`), `.moai/policies/workflow/factory-dispatch.md`, `-cards.md`, `-mechanics.md`, `auto-semantics.md`, and `.moai/workflows/gtd.md`; spec.md:L116 names the `.moai/policies/**` and `.moai/workflows/gtd.md` copies as not to be hand-edited. Probe (the checker's `cmd_paths` with the git call replaced by a fixed diff list): `.moai/policies/workflow/factory-dispatch-detail.md`, `.moai/workflows/gtd.md`, `.moai/docs/x.md`, `.moai/config/sections/workflow.yaml` each printed `unexpected-paths 0`, exit 0; `.claude/agents/moai/manager-todo.md` and `AGENTS.md` printed `unexpected-paths 1`, exit 1. The independent codex run reproduced it on the real checker (a sentence added to `.moai/policies/card-dispatch.md` printed `unexpected-paths 0`, exit 0, tracked and untracked; the same sentence at the root exited 1). No other criterion reads these paths (AC-MJ-016 lists only the eight template files). This is the iteration 1 defect D16 in a place the closed-world check does not look, and the correction round again states coverage the code does not give; criterion (b) itself says "outside `.moai/`" (acceptance.md:L242), so the title, L92, R-8 and D-10 contradict it. Severity: BLOCKING (a writable mutant passes the release-blocking set and four statements are false). Class: blocking. Required fix: limit the exemption to the paths a run legitimately writes (`.moai/specs/SPEC-CARD-MODE-JUDGMENT-001/`, `.moai/reports/`, `.moai/state/`, `.moai/logs/`) in `cmd_paths`, add a known failing input cell (a `.moai/policies/` edit reads `unexpected-paths 1`, exit 1) and re-pin both SHA-256 values in E-020d; or, if the exemption is kept, rewrite acceptance.md:L239, L92, spec.md R-8 and design.md D-10 to say that paths under `.moai/` are not covered and name the tracked projections as the open residual.

D24. MINOR. E-013c, a control cell of AC-MJ-013, does not reproduce in the lane's own tool — acceptance.md E-013c (stdout `2` `2` `1`) versus the Claude Code Bash tool, where `grep` is a shell function over ugrep 7.8.4: the same command printed `0` for all three files, exit 1, while `/usr/bin/grep` (BSD 2.6.0-FreeBSD) printed `2` `2` `1`, exit 0. The five-way alternation `t(133|224|696|1339|1393)` is mishandled by ugrep (`t(1339|1393)` matches, `t(133|1393)` prints 0). The leak pattern itself (E-013a, E-013b) printed the same counts under both greps, and a 30-line probe file matched 21 lines under both, so the instrument is sound; only the control reads dead in the tool. acceptance.md:L10 says cells were run through `sh` with the system grep, so the cause is known but not recorded for the next runner. Class: optional. Required fix: add one line to section 0 saying cells are judged by `/usr/bin/grep`, or give the alternation in a form both engines agree on.

D25. SHOULD-FIX. The order-blind comparison admits a structural mutant that breaks a changed document while every criterion stays green — closed-world-check.py:76 (multiset difference), acceptance.md:L245, L253 and design.md D-10 (disclosed limit: "a line moved without change is not seen") — evidence: in a copy built from the manifest, moving the existing closing code fence of C (line 18, `` ``` ``) behind the new Mode judgment paragraph puts the new subsection inside a code block; `lines` printed `0 0 0 0` on all sixteen files, exit 0, and E-001 printed `8` `8`, exit 0 (my reproduction of codex finding P2). Moving a Routing sub-bullet or a listed heading is likewise unseen. Lexical limits are acknowledged by the SPEC (R-8), so this is a narrowing hole rather than a false claim. Class: optional. Required fix: add a fifth number to `lines` (count of fence lines starting with three backticks per file against the base count) or compare each `+` row with its preceding base line; observe the fence mutant red.

D26. MINOR. design.md D-11 "Freezing" overstates the hash coverage — design.md D-11 ("any change to either file changes the plan-artifact hash through acceptance.md") versus audit_cache.go `planArtifactNames` (the helpers are not subjects): an edit to a helper alone leaves the plan-artifact hash, and so a cached PASS, valid; E-020d catches it only when the run executes. acceptance.md:L247 states it correctly ("requires a matching change here"). Class: optional. Required fix: reword D-11 to "a helper edit is caught by E-020d at run time; keeping E-020d green requires changing acceptance.md, which changes the hash".

D27. MINOR. spec.md:L47 quotes SPEC-FACTORY-DEFAULT-MODE-001 (inside the verbatim 02f4 body), which is not in `.moai/specs/` (D7-5 SHOULD). It is card t1737's future SPEC; the quote must stay verbatim. Class: optional. Required fix: none; optionally add "(future SPEC of card t1737)" next to the quotation outside the blockquote.

Stagnation and regression: none. The iteration 1 defects were each worked and none repeats unchanged; the aggregate rose (0.79 to 0.88); the new blocking defect (D23) is a residual of the fix for D16, not a repeat.

## Recommendation (final round: the ceiling policy applies)

1. D23: narrow the `.moai/` exemption in `cmd_paths` (or correct the four statements), add the negative-control cell, re-pin E-020d. One scoped edit set: closed-world-check.py, acceptance.md AC-MJ-020 / E-020b / E-020d, and the wording of AC-MJ-006 L92, spec.md R-8, design.md D-10/D-11.
2. Optional D24-D27 are surfaced for the orchestrator's discretion; they do not by themselves justify the failure. D25 is the one worth taking in the same edit (one more number in `lines`).
3. The defect is anchored to named files and anchors (fix_scope above), so the final-hit admission ladder can route it to a split or a debt rather than a blind hold; the decision is the ladder's, not this report's.

## Five-section evidence record

**Claim.** D17-D22 are closed, D1-D15 hold, and D16 is closed for the stem-less added-line mutants; the SPEC still fails because the required codex gate failed on a P1 I reproduced (D23: the `.moai/` exemption of the path check, contradicted by four statements of coverage) and one blocking defect stands.

**Evidence.** Commands run from /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1894 (exit codes as observed):
- `git rev-parse HEAD` printed cb71f2b475263dc8b0921e37836dcc8208a62f71; `git status --short` printed nothing.
- `python3 -I .moai/specs/SPEC-CARD-MODE-JUDGMENT-001/closed-world-check.py lines <manifest>`: eight rows as in E-020a, exit 1. `... paths <manifest>`: `unexpected-paths 0`, exit 0; with the old base `3642a3a9b9...`: `unexpected-paths 175`, exit 1. `... pairs <manifest> 957967601351...`: `pairs 0 unpaired 7`, exit 1; with `1969709ce693...:2`: `pairs 2 unpaired 0 expected 2`, exit 0. Hash cells E-020d `same same` exit 0, E-020e `DIFFERENT same` exit 1.
- Mutants through the checker: scratch builds from the manifest, 21 cases (table above), saved under /private/tmp/lxq/aud/ (`cw_results.txt`, `paths_probe.py`, `fence_mutant.py`); fence move: `0 0 0 0` exit 0 and E-001 `8` `8`.
- 22-file AC-MJ-006 family: output `8 8 1 1;8 8 1 1;8 8 1 1;8 8 1 1;8 8 1 1;7 8 1 1;7 8 1 1;8 9 1 1;8 8 0 1;8 9 1 1;8 8 0 2;8 8 1 1;8 10 1 1;8 10 1 1;0 0 0 5;0 0 0 6;7 8 1 1;7 8 1 1;8 8 1 2;8 8 1 2;7 8 1 1;7 8 1 1`, exit 1 (one red among them by design). E-005d on the same files: good `3`, N9 `2`, the four stem-less `3`.
- E-002a, E-006, E-009, E-007, E-008a-f, E-001 run as written through the guard, none refused (outputs in the table). The Group 6 CN-4 awk verb was refused (a `|` after printf in its text); a python port ran instead.
- `moai verify audit-plan --project-root <toplevel>`: config_status ok, model codex, codex gate required, claude and glm off, enforced_required [codex]. `mcp__moai__audit_multi` called once (target baseBranch, delta_base 4834700468ed2311e3efc4e12bb920db03db54e2, project_root = this tree): overall verdict fail; codex (required) fail with two findings (P1 the `.moai/` exemption at closed-world-check.py:96, P2 order-blind line moves at closed-world-check.py:76), both reproduced here (D23, D25); receipt rcpt-55ffb196ab252308a7c4a3be; server build commit 7705b365e2, an ancestor of HEAD cb71f2b475 (binary lag stated by the tool). After the call, `moai verify audit-plan ... --result-file .moai/state/audit-plan-result.json`: `convergence_check` ok true, unmet [] (the required backend answered; its answer was a fail, a reviewed result and not an unmet gate).
- `moai spec lint --strict spec.md`: `0 error(s), 0 warning(s)`, one INFO, exit 0. Trace port: 17 REQ, no UNCOVERED, no ORPHAN, 20 AC headers. Plan-artifact hash by python port of `ComputeHash`: `5bda514f...bfca`, equal to the lane value.

**Baseline-attribution.** Every observation was measured in this run against HEAD cb71f2b475263dc8b0921e37836dcc8208a62f71. CLI results were judged by the installed build moai_cp/20261007_195927 (older than HEAD); the codex leg ran in the audit_multi server (build commit 7705b365e2, also older than HEAD) and its findings were accepted only after reproduction here. Scratch mutants were built in this run from the manifest rows and are cited only through printed outputs. The pinned comparison base of the checker (d31fb64239a688c053e455000cc9fe37f540b790) is an earlier commit of the same branch; the sixteen documents are byte-identical between it and HEAD (`git diff --stat` over the document trees printed nothing).

**Gaps (not observed).**
- G1: the MCP `audit_cache` tool is not in this agent's tool list, so `compute_hash` was not called; the hash was reproduced with a python port of `ComputeHash` read from `internal/runtime/audit_cache.go` and equals the lane's value. The Go implementation was not executed.
- G2: the Group 6 CN-4 awk verb was refused by the guard; the python port reproduces its logic by my reading, not by running the original. A mechanical CONFLICT from the original cannot be strictly excluded, though the plan binds no criterion through an `Exit:` line (0 bindings).
- G3: no Go test, build or make target was run (docs-only plan audit). What the named guards pin comes from the file and function names and the ledger rows; `TestRoleCoreStubGate` and its neighbours were not executed against the planned text, which exists only as manifest rows and the author's scratch simulation (progress.md G-15, not available to this audit). The 74-mutant claim of AC-MJ-020 was not re-run in full: 21 of my own mutants plus the 22-file family were.
- G4: the manifest rows were not recounted line by line; row counts come from the author and from the checker's own per-file numbers (C 2, K 2, F 5, D 12, M 7, A 8, G 10, W 2 additions per side).
- G5: the positioned post-run text of the other seven documents was not reconstructed (only C, for the fence mutant); the green values of the pinned cells on a positioned copy of D were reasoned from the pinned lines being exact manifest rows, and the earlier iteration's reconstruction.
- G6: git cells with a trailing exit probe are refused by the guard; the exit codes of E-002b, E-002c, E-017 are stated from git's contract. Inline `python3 -c` naming `git` was refused once, so the stub arithmetic was run from a file reading the working copy (equal to the base for that file).
- G7: tool-environment difference: the shell-function grep (ugrep 7.8.4) mishandles one alternation (D24); all other grep cells agreed with the ledger under it, and the cells that matter were also run with `/usr/bin/grep` where stated.
- G8: sibling-SPEC precedents for DEFAULT-APPLIED were not re-read this iteration (judged against the stamp rule and the iteration 1 reading); research.md and design.md were read as diffs from the iteration 1 commit plus the sections cited.

**Residual-risk.** (1) Even after D23 and D25, the criteria are lexical (R-8): a contradiction inside a line that also carries a pinned token, or a move that changes meaning without changing a line, is seen only by the sync audit's reading of the manifest rows. (2) Frozen-manifest cost: the planned wording has not met the Go guards, and if a landing race makes the lane merge newer main into the card tree, `lines` and `paths` go red on lines that are not the lane's (the base is a pinned card-branch commit, not a merge-base); the defined path is a blocker report. (3) A helper-only edit does not move the plan-artifact hash; it is caught by E-020d only when the run executes it. (4) The audit_multi server and the installed moai predate HEAD; the codex leg's scratch mutants ran in its own environment, and I reproduced both of its findings here. (5) Q7 stays open: the autonomous Kickoff form is blocked until the operator answers, independently of this report.

## Operational Notes (unverified)

- Measure the D23 fix before landing it — command: `python3 -I /private/tmp/lxq/aud/paths_probe.py` run from the worktree root (it replaces the git call with a fixed diff list; expected after the fix: the `.moai/policies` and `.moai/workflows` lines exit 1, the SPEC directory and `.moai/reports` lines exit 0). Status: measured before the fix (all four `.moai/` lines exit 0), unmeasured after.
- Measure the D25 fix — command: `python3 -I /private/tmp/lxq/aud/fence_mutant.py`; expected after the fix: exit 1. Status: measured before the fix (`0 0 0 0`, exit 0), unmeasured after.
- Run every ledger cell once in the lane's own tool before M1 and compare with the ledger, with `/usr/bin/grep` where D24 applies. Status: inferred (the shell-function grep differed for one cell here).
- Q7 can be closed by one operator sentence: a named mechanism, or "the ruling's phrase is the whole condition". Status: assumption (decision-gate rule read from the agent contract, not executed).

AUDIT-VERDICT: FAIL spec=SPEC-CARD-MODE-JUDGMENT-001 receipts=rcpt-55ffb196ab252308a7c4a3be
