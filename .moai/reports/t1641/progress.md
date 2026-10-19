# Card t1641 — progress record

Card: t1641 (run tn1p4e, Class C, P1, operator-approved). Branch `WT-todo-unify`. SPEC: `SPEC-TODO-UNIFY-001`.

## Stage: plan (authored 2026-10-18T06:37Z, tree `026f48658f`; revised for audit iteration 1; revised again for audit iteration 2)

- Artifacts: `.moai/specs/SPEC-TODO-UNIFY-001/` holds spec.md, plan.md, acceptance.md, design.md, research.md, decision-index.md, progress.md. Tier L; 25 requirements, 25 criteria (revisions 2 and 3; revision 3 added rows, not identifiers).
- Plan audit iteration 1: FAIL 0.62 (MP-7 failed, ten blocking findings; receipt `rcpt-90e345b1d5488c65cc03140e`; report `.moai/reports/t1641/plan-audit.md`). Revision 2 repaired D1 to D15.
- Plan audit iteration 2: FAIL 0.80 against the Tier L threshold 0.85 (must-pass all OK, blocking_count 7, codex required leg `fail`, receipt `rcpt-717ae7d0bd489fdeced99ba8`; report `.moai/reports/t1641/plan-audit-iter2.md`). Revision 3 (this record) repairs D16 to D22 and the cheap optional items; it needs plan-audit iteration 3, the last under the Tier L ceiling (a third FAIL goes to hold-and-split).
- Kickoff: no open decision row remains. Q5 (the automatic receipt) is the one trust-model decision, marked `lane-judgment, leader-confirmed 10-18 (conditional)`, shape unchanged in revision 3.

## Revision 3: where each iteration-2 defect is closed, and the criterion whose mutant fails

All line numbers are in the files as committed by the revision-3 commit (`grep -n` of the final files). "Mutant" is the wrong implementation that the named criterion row fails.

| Defect | Closed at | Criterion and failing mutant |
|--------|-----------|------------------------------|
| D16 stale PASS at an early commit, verdict read from the lane-writable worktree, TOCTOU | spec.md:110-116 (REQ-TU-017 facts 1 to 3, one transaction); design.md:85-105 (where facts are read, table :89-95, why 2c :97-99, TOCTOU :103); research.md §R.14 (:254-279, measured on eight merges); decision-index.md:71 (Q5 decision line); plan.md M2a (:83-89) | AC-TU-016 (acceptance.md:82-93) rows (c) stale evidence and (d) worktree-only verdict: the revision-2 ancestor rule and a worktree reader fail them; row (g) rollback: a separate-transaction issuer fails it; row (b) with the absorbed-main shape: a reports-only rule fails it |
| D17 re-measure record never compared with the merge | spec.md:114 (fact 3); design.md:95, :101; plan.md M2a GREEN (`factory.VerifyRemeasureBinding`) | AC-TU-016 row (f) (acceptance.md:88): a callback that validates without comparing `tree` and `base` fails both rows |
| D18 exact-32 verb pin vs the `spec` verb (33) | acceptance.md:39-43 (AC-TU-004); design.md:46, :72, :166 (D.8 mutant row), plan.md M3b, M4 (:107-123); research.md E-19, E-20 (:39-40) | AC-TU-004: a build with 32 (stage verbs without `spec`) or 27 (`spec` without stage verbs) fails the exact-33 set |
| D19 warning precedes required-flag and flag-group refusals | spec.md:78, :80 (REQ-TU-002/003); design.md:28 (the run-step wrapper); research.md §R.11 (:242-244) and E-22, E-23 (:42-43); plan.md M5 (:124-128); acceptance.md:30-33 (AC-TU-002) | AC-TU-002 rows (4) `gtd capture x`, (5) `gtd --auto --auto-lane`, (6) lane guard: a warning printed from the pre-run hook (the iteration-2 design) fails rows (4) and (5) |
| D20 setter refuses the state of the wedge it was written for | spec.md:128, :130, :132 (REQ-TU-022/023/024); design.md:131 (which states); research.md §R.15 (:281-288); decision-index.md:110 (Q8) | AC-TU-021 (acceptance.md:118-121): a setter with the revision-2 rule fails the `leased`, `plan`, `plan-audit`, `kickoff` rows; AC-TU-022 (:123-125): the wedge fixture and its control |
| D21 half-written setter overtaken by the lease edge | spec.md:128 (completing repeat in every non-terminal state); design.md:135-137; plan.md M3a, M3b | AC-TU-025 (acceptance.md:127-129): a repeat that applies the revision-2 state rule fails the lease-between-attempts rows; one that applies the first-fill set fails the advance-to-`run` row; completing a pair on a terminal card fails the abandoned row |
| D22 two clauses without a criterion row | spec.md:122 (log-row issuer moved into REQ-TU-020, where AC-TU-014 maps); acceptance.md:73-76 (AC-TU-014); design.md:109-111 | AC-TU-014: a close that logs the path but no issuer fails the issuer assertion; one that logs `auto` for a leader-issued receipt fails the leader-issuer row; one that moves a `pushed` row fails byte-identity |
| D23 (optional) REQ-FCR-014 reversal | research.md §R.8 (:223); design.md:107; spec.md:34 (HISTORY), :116 (REQ-TU-017), :161 (Out of Scope); plan.md §C item 3 (:57), M2a files, §H (:190) | no criterion (naming only); M2a pre-flight blocks if the draft's status changed |
| D24 (optional) E-15 pathspec | research.md:35 (E-15 over `internal/cli`, exit 1 observed); acceptance.md:27 | AC-TU-001 RED-now proxy flips wherever the phrase lands |
| D25 (optional) `TestNoResidualGTDSpelling` red between M7 and M8 | plan.md M7 RED (:141), M8 RED (:150); acceptance.md:67 | AC-TU-010 |
| D26 (optional) T29 deciders | spec.md:120 (REQ-TU-019 clause c); design.md:119; plan.md M1 (:75-81); acceptance.md:99-102 (AC-TU-018) | AC-TU-018 decider row: a T29 that admits any decider fails the empty-decider and other-string rows; one admitting only `human` fails the `audit` row |
| D28 (optional) REQ order | spec.md:134-136 (REQ-TU-025 moved to §3.4, after REQ-TU-024) | none (format) |
| D27, D28 second half | not done | M2a (about 8 files / 470 lines) and M2b stay above the plan's own 400-line figure at estimate level, each with a split clause (M2a's names M2c); the identifiers T18 and T29 stay inside REQ-TU-018/019 |

## Decisions made without asking (recorded, per the decision ladder; evidence in `decision-index.md`)

- Q1 absorb all six stage verbs on `moai todo`, delete none. Q2 the warning states removal one release after the release that first carries it. Q3 user-facing spelling only; the workflow document keeps its path. Q4 lane own-card path plus the automatic receipt inside the two close paths plus edge T29; no observers. Q6 remove T18 and the T19 reservation now; declare T29. Q7 both records end `done`, factory edge first, queue archive second, with a stated recovery. Q8 a new leader-side `moai todo spec` verb, set-once per carrier (revision 3: first fill from `picked` through `kickoff`, completing repeat in any non-terminal state). Q9 leader only.
- Q5 (leader-confirmed, conditional): the receipt is issued automatically only when three machine facts all hold; issuer `auto`; revocable; fail-closed to the leader's manual approval. Revision 3 changed only how facts 2 and 3 are bound (design.md §D.5).
- Revision-3 design choices recorded in the SPEC with their rejected alternatives (design.md §D.9, :209-214): print the warning from the run step instead of narrowing REQ-TU-003; the verdict-corroboration fact not taken (a fourth fact outside the leader-confirmed shape); the card-tip-equals-audited-commit rule and the reports-only rule rejected on measured data; setter first fill limited to states before `run`.

## Verification this round (command, then what was observed; all plain commands in this worktree)

- HEAD and tree: `git rev-parse HEAD` printed `3dc1052f6faf0883ceebd747a8af5979c1508d93`, branch `WT-todo-unify`, `git status --short` empty before editing. `git diff --stat 23e40eb7c4 3dc1052f6f -- . ':!.moai/specs/SPEC-TODO-UNIFY-001' ':!.moai/reports/t1641'` printed nothing, exit 0.
- Judging build: `go build -o <scratchpad>/moai-t1641-iter3 ./cmd/moai` exit 0 (plain build, no Makefile LDFLAGS; its version stamp prints `moai-adk v3.1.3`, so the tree HEAD is its identity).
- Verb count (D18): `moai todo --help` and `moai gtd --help` from that build, COMMANDS blocks counted by `python3 -I count_verbs.py <help file>`: `26 add list done ... triage` and `32 add list ... triage capture clarify organize reflect engage answer`. Expected after the SPEC: 26 + `spec` + six = 33 on each tree.
- Cobra (D19): `grep -n` of `command.go` in the module cache printed `919 ParseFlags`, `935 flag.ErrHelp`, `968 ValidateArgs`, `985 PersistentPreRunE`, `999 PreRunE`, `1007 ValidateRequiredFlags`, `1010 ValidateFlagGroups`, `1014 RunE`; `go.mod:23` pins `cobra v1.10.2`. `git grep` printed `gtd.go:98,121,158 MarkFlagRequired` and `todo.go:356 MarkFlagsMutuallyExclusive`; `git grep -n -w -F 'Run:'` over the constructor files printed only two unrelated `factory.GitLandedQuerier{Run: ...}` lines (my first check with a line-leading regex was too weak and was replaced), and the `RunE:` positive control printed `gtd.go:5`, `gtd_answer.go:1`, `todo.go:6`.
- Board (D20): `moai decision read --scope card:t1641` printed `board=ok unparseable=0 records=108`; record `d-20261010T001732Z-d4e8` predicate "a card of run tmnboq leased without a spec_id whose SPEC was authored after the lease", body "The leader stamps spec_id on a leased card row ...". The readers of the identifier: `card_evidence_readers.go:185-188`, `card_audit_kickoff.go:31`.
- Landing flow (D16, D17), `git rev-list --parents`, `git ls-tree`, `git show`, `git diff --name-only`, `git merge-base --is-ancestor`, `cat` of `.moai/state/remeasure/<tree>.json`, over eight merges (research.md §R.14): five carry no committed report files in the card tip; three do (t1856, t1837, t1812). For all three the merge tree equals the second parent's tree and the record's `tree` equals it and its `base` equals the merge's first parent. For the audited commit versus the card tip, t1856 and t1837 differ by two report files each; t1812 differs by 254 paths (main absorbed after the audit). `ad_rule.py` (scratchpad, `git -C` subprocess) printed for t1812 `R1 FAIL` (252 violations) and `R2 PASS` (0), and R1 PASS and R2 PASS for the other two. R2 is the rule written into REQ-TU-017.
- Ledger re-runs with exit codes shown: E-15 exit 1, E-16 exit 0, E-17 exit 1, E-18 exit 1, E-21 exit 1, E-23 exit 0.
- Lint: `moai spec lint SPEC-TODO-UNIFY-001` (by path, the binary above) printed `0 error(s), 0 warning(s)` and one INFO `OwnershipTransitionUnmeasured` (commit `63829fe20a` has no `Authored-By-Agent` trailer); `lint exit=0`. `moai spec lint --strict SPEC-TODO-UNIFY-001` printed the same and `strict exit=0`.
- Counts: `git grep -n -c -E '^\*\*REQ-TU-'` over spec.md printed 25; `git grep -n -c -E '^\*\*AC-TU-'` over acceptance.md printed 25; `git grep -n -c -F 'NEEDS CLARIFICATION'` over the SPEC directory printed nothing (no marker).

## Gaps at the end of this stage

- No `go test` run (card constraint, and the heavy slot was not taken): every class B criterion is unadopted until its RED is recorded, and every claim about the behaviour of the designed code (the in-transaction issuance and its rollback, the RunE wrapper's refusal rows, the setter state table and interleavings, the 33-name enumeration, the `VerifyRemeasureBinding` comparison) is a design, not an observation.
- The cobra order was read and its anchors grepped, not executed with a throwaway program.
- The commit sample behind the choice of rule R2 is eight merges, not a census; the recency pattern of committed report files (the three that have them are the most recent audited cards) is a reading. R2 over-refuses two shapes by design (a file both the card and main changed in different hunks after the audit; verdict files numbered `-r<N>` rather than `-iter<N>`), recorded in design.md:99.
- Whether the merge verb's own step (REQ-MWQ-018 cause 4) compares the record's `tree` before T16 was not read; the T16 callback itself was read and does not.
- The lag of the installed `moai` (used for `moai decision read` only) was not re-measured this round; the leader's confirmation of Q5 is the coordinator's message, not a board record.
- Three commands were refused by the worktree guard this round (a `for` loop over `git rev-list`; `python3` with a shell variable in the script path; a `cat` whose path held a `$(git ...)` substitution). Each was re-run as plain separate commands or with literal absolute paths; every result quoted above is from those runs. No `sh`, `awk` pipeline, or `python3 -c` was attempted.
- The docs-site build, the board mod's JavaScript tests, and `sh` availability on every CI runner remain unexamined, as in revision 2.
- `ad_rule.py` and `count_verbs.py` live in the session scratchpad and are not part of the commit; their logic is stated in design.md §D.4 and §D.5 and research.md §R.14 so a reader can redo the measurements with plain git.

## Audit ceiling hold (lane-22, 2026-10-18)

wait record: reason = plan audit iteration 3 of 3 returned FAIL (`.moai/reports/t1641/plan-audit-iter3.md`: overall_score 0.86 against the 0.85 threshold, must_pass_failed 0, blocking_count 3, plan_artifact_hash cbe6819defad8c2e664b58a7b275f3b973601f0005b119ed30de4275ad69e97a, required codex leg `fail` on both runs, receipts rcpt-9557497ba4bcc4b92912703c and rcpt-01bf6dad688922d13c1931d1); the audit ceiling is exhausted, and a further revision or a split is the leader's decision, not the lane's.
Open blockers (all in the iteration-3 text; the requirement and criterion identifier sets need not change): D29 the push-gate caller `factory_card.go:3238` passes only the configured integration target, so in a no-remote repository without git-strategy config T29 refuses where T18 admitted (fix anchor: add `factory_card.go` to M1 and a CLI-level no-config row); D30 `moai factory approve --issuer auto` writes issuer `auto` with the leader role into a store keyed `(card_uuid, run_id)`, so a manual receipt is indistinguishable from an automatic one and revocation could hit a leader receipt (fix anchor: a column or reserved string for the automatic kind, plus a same-card manual row); D31 the design reads every fact through `gitRead`, which trims whitespace and does not disable replacement objects, so the claim that a lane cannot change what a fact reads is false as stated (fix anchor: a raw reader with replacement objects off, plus two rows).
Auditor's reading: the SPEC is not yet sound enough for run phase; the three blockers are narrow; nothing points to a split; a delta re-audit scoped to D29-D31 plus the full must-pass set is the natural next step. Informational (D35): 65 of the 83 recent integration merges on local main commit no sync-audit file, so the ten-card automatic-done criterion reads low until lanes commit their evidence.
whom: leader. recheck: each cron wake reads the leader's reply and the decision board; the lane revises or splits only on a recorded ruling.

## Post-ceiling correction, revision 4 (2026-10-18T09:01Z, manager-spec, under the leader's standing rule C)

Authority: the leader's standing rule C as quoted in the dispatch (board record `d-20261018T084540Z-325a`; the board was not read by this session, so the record id is the leader's citation): after the audit ceiling, one correction commit and one delta audit; a FAIL then means hold and shrink. Input: `.moai/reports/t1641/plan-audit-iter3.md` (FAIL 0.86, must_pass_failed 0, blocking_count 3: D29, D30, D31). Scope: those three only; D32 to D37 are untouched and D35 is out of scope as ordered. Start state: HEAD `f40740ea61`, branch `WT-todo-unify`, clean. The delta audit, `plan-audit-iter4.md`, is the auditor's to write; this session wrote no `plan-audit*.md`. Q5's shape and its operator-verdict line are unchanged, and the identifier sets are 25 requirements and 25 criteria as before.

### Where each blocker is closed (file:line in the committed files) and the row that fails its mutant

| Blocker | Closed at | Row that fails the mutant | Mutant |
|---------|-----------|---------------------------|--------|
| D29 push gate caller passes only the configured target, so T29 refuses in a no-remote repository without configuration | design.md:129-133 (§D.6 "The landing ref without a remote, and what the caller passes"), design.md:19 (push-gate row of the current-state map), spec.md:121 (REQ-TU-019 (a) and last sentence), plan.md:78-81 (M1 RED (4), GREEN, Files, Packages), acceptance.md:104-107 (AC-TU-018), research.md:45 (E-24) and :296 (§R.16) | AC-TU-018 (acceptance.md:104): the CLI-level no-configuration row (`TestPushGateNoRemoteNoConfig`), the own-branch row, and the configured-branch row; mutants at acceptance.md:107 | a T29 whose no-remote test needs a configured branch, or tests the card's own branch (what `factoryResolveIntegrationBranch` answers), refuses the no-configuration row; one that counts the card's own branch fails the own-branch row; an empty-value test applied when a branch is configured fails the configured-branch row |
| D30 issuer `auto` cannot tell an automatic receipt from a manual one, so revocation could hit a leader receipt | design.md:111 (reserved value, one writer) and :113 (revocation), design.md:14 (receipt-issuance row), spec.md:117 (REQ-TU-017 last sentence) and :137 (REQ-TU-025), plan.md:62 (pre-flight item 8), :87-88 (M2a GREEN, Files), :200 (risk row), acceptance.md:92 (AC-TU-016 row h, reserved-value clause), :109-111 (AC-TU-019), research.md:46 (E-25) and :298 | AC-TU-016 row (h) reserved-value subtest (acceptance.md:92, mutants :98); AC-TU-019 same-card leader-receipt row and the `--issuer auto` / `--issuer AUTO` rows (acceptance.md:109, mutants :111) | a manual path that accepts `auto` fails both the row (h) subtest and AC-TU-019's reserved-value rows; an exclusion that removes any receipt of the (card, run) pair fails AC-TU-019's same-card row |
| D31 `gitRead` trims whitespace and does not disable replacement objects, so "a lane cannot change what a fact reads" was false | design.md:87 (claim corrected), :92-95 (table), :97-99 (the reader), :107 (TOCTOU), spec.md:111 (REQ-TU-017 parenthesis), plan.md:79 and :86-87 (raw reader in M1 and M2a), acceptance.md:93-95 (rows i, j, k), :96 (Green), :98 (mutants), research.md:47 (E-26) and :300 | AC-TU-016 rows (i) leading-space path (acceptance.md:93), (j) replacement refs (:94), (k) graft and shallow files (:95); mutants at acceptance.md:98 | a trimming reader fails row (i); a reader with replacement objects enabled, or hardened for the diff reads only, fails row (j) (j2 reaches the merge commit's tree and parents); a reader that disables replacement objects and ignores graft and shallow files fails row (k) |

Known debt (the leader's premise rule, `.moai/reports` is gitignored): KD-1 stands in plan.md:202-204 (§H, "Known debt (revision 4)"), decision-index.md:132-140 (a `## Known debt` section, not a decision row, no Q number), and spec.md:164 (an Out of Scope bullet). Text: the automatic-receipt facts depend on lanes committing their evidence under the gitignored `.moai/reports`; where the evidence is not committed, fact 2 is absent and the automatic path simply does not issue (fail-closed, the leader approves by hand); the SPEC deliberately does not fix this by requiring report commits (operator policy), does not remove fact 2, and no criterion claims the automatic path works for uncommitted evidence (AC-TU-020 now says the opposite, acceptance.md:113). Measured: `git check-ignore -v --no-index .moai/reports/t1641/progress.md` printed `.gitignore:235:.moai/reports/*`; the operator order of 2026-09-14 is the leader's statement and was not read.

### Findings that differ from the audit's fix anchors (read the code, then chose)

1. **D29: the audit's first alternative does not work, so `factory_card.go` is not added to M1's edit list.** `factoryResolveIntegrationBranch` (`internal/cli/factory_card.go:2552-2557`) answers the card worktree's own branch when nothing is configured (`:2556`, source `BranchSourceCaller`), and `factory complete` itself refuses that source and the card's own branch (`:2146-2162`); the merge commit is a child of the card tip and is not in that branch. Resolving through it would still refuse where removed T18 admitted. The correction sits in T29's guard instead: exact reachability when a branch is named, and for an empty value "contained in a local branch other than the card's own"; the caller keeps forwarding the configured target (`factory_card.go:3238`, `:3394`), read and not edited (plan.md:80). This follows the audit's second alternative (restating the no-remote landing test), not the first. The empty-value test accepts one case the exact test would not (a merge dropped from the integration branch but still held by another local branch, observed in the throwaway repository); it is stated and bounded to the no-configuration case (design.md:133, plan.md:199).
2. **D31: a third rewriting mechanism, measured.** `--no-replace-objects` does not cover `info/grafts` or the `shallow` file (both honored under the option, both reachable from a linked worktree's common git directory). The design adds a refusal when either file exists, with row (k). This goes past the audit's two cases because without it the corrected claim would still be false.
3. **D30: mechanism chosen.** A reserved issuer value, not a column: one comparison in the one function that is the whole manual path (writers enumerated, E-25); a column needs a table migration and a flag reset on replace-on-conflict (design.md:241, §D.9).

### Verification this round (commands run, output seen; tree HEAD `f40740ea61`)

- Build: `go build -o <scratchpad>/moai-t1641-iter4 ./cmd/moai`, exit 0 (plain build, no Makefile LDFLAGS; the tree HEAD is the build's identity; code is identical to the audited revision, see below).
- Lint, by path from the tree directory, exit codes captured by a script (`lint_run.py` in the scratchpad), final run on the files as committed: `moai spec lint SPEC-TODO-UNIFY-001` printed one INFO `OwnershipTransitionUnmeasured` (commit `63829fe20a0f13fa93e974fe9331bb455ccb1009` has no `Authored-By-Agent` trailer) and `0 error(s), 0 warning(s)`, `exit=0`; `moai spec lint --strict SPEC-TODO-UNIFY-001` printed the same, `exit=0`.
- Identifier sets (`final_checks.py`): spec.md 25 REQ ids, unique 25, same ids in the same order as the previous revision at HEAD, contiguous 001..025; acceptance.md 25 AC ids, unique 25, same; `git grep -n -c -E '^\*\*REQ-TU-'` printed `spec.md:25` and `git grep -n -c -E '^\*\*AC-TU-'` printed `acceptance.md:25`, exit 0 each.
- Q5 marker: the line `- Operator verdict: lane-judgment, leader-confirmed 10-18 (conditional)` occurs once in decision-index.md now and once at HEAD, and the two lines are identical (`True`); the marker text occurs once in each.
- No clarification marker: `git grep -c -F 'NEEDS CLARIFICATION'` over the SPEC directory printed nothing, exit 1.
- Code unchanged: `git diff --stat 164050fdde f40740ea61 -- . ':!.moai/specs/SPEC-TODO-UNIFY-001' ':!.moai/reports/t1641'` printed nothing, exit 0; the same against the working tree printed nothing, exit 0; `git status --short` lists only the seven SPEC files before this record was added.
- Ledger greps with exit codes captured (research.md E-24 to E-26): the caller's value at `factory_card.go:3238`; the receipt writers `leader_approval.go:173`, `card_transition.go:849`, the one caller `leader_approval_gate.go:350`; no tracked writer of `Issuer: "auto"` (exit 1); `card_evidence.go:33` trims; no use of `--no-replace-objects` outside a parser skip-list (`commit_identity_guard.go:456`); `gitReadRaw` absent (exit 1).
- Throwaway-repository probes (git 2.54.0), two scratchpad scripts calling `git -C <repo>` and one ten-line Go program (`go run`, exit 0): leading-space path, replacement ref, graft file, shallow file, `for-each-ref --contains`, and the `man git` text; results verbatim in research.md §R.16 (:294-:308).
- Read, with line numbers quoted above: `factory_card.go:2146-2162, 2552-2557, 3238, 3304-3396`, `card_transition.go:722-747`, `card_evidence_readers.go:273-276, 290-318`, `card_evidence.go:19-34`, `leader_approval.go:149-182`, `leader_approval_gate.go:344-373`, `loader_integration_branch.go:91-129`, `loader_workflow_disposition.go:90-103`, `integration_merge.go:54-63`, `integration_merge_step.go:399`.

### Gaps (not observed)

- No `go test` was run (card constraint, heavy slot not taken): every new row of AC-TU-016 (i)-(k), AC-TU-018, and AC-TU-019 is a criterion with a mutant, not an observation; class B criteria stay unadopted until their RED is recorded at run time.
- `decide --gate push` was not run in a configuration-less repository: a built `moai` cannot reach a `merged-local` row without the whole factory chain, and no `go test`; D29 rests on reading the caller and guard code (quoted) and on the audit's own reading, and AC-TU-018's CLI-level row is its test.
- The number of rows already carrying the issuer `auto` in a live factory store was not counted (the store is under the operator's home and was not opened); M2a's pre-flight item 8 does it.
- The graft entry probed lists the tip alone (a parentless rewrite); the first-parent rewrite named in row (k1) was not constructed, its form is left to the run-phase fixture. One git version (2.54.0) only. The Go program reproduces the `TrimSpace` line at `card_evidence.go:33`; it does not call the unexported `gitRead`.
- The operator order of 2026-09-14 behind `.gitignore:235` and the board record `d-20261018T084540Z-325a` were taken from the leader's message and not read.
- D32 to D34, D36, D37 and the other optional items are untouched; D27 (M2a and M2b above the plan's own 400-line figure) is not fixed, and M2a's estimate grew to 520 lines with its split clause naming M2c; M1's grew to 9 files and 400 lines with a split clause naming M1b.
- Refusals by the worktree guard this round: none observed. Two self-inflicted command errors, neither a guard refusal: `git grep ... -A 22` after the pattern (git read `-A` as a revision, exit 128) and `git check-ignore -v` on a tracked path (exit 1 until `--no-index` was added); a quoting mistake in one `ledger_exit.py` command (a mangled `"T29"` grep) whose output was not used.
- The scratch scripts and probe repositories live in the session scratchpad and are not part of the commit; their logic is stated in research.md §R.16.

## Delta audit outcome and shrink proposal (lane-22, 2026-10-18)

wait record: reason = the one delta audit of standing rule C (board d-20261018T084540Z-325a) returned FAIL by the required-gate label (`.moai/reports/t1641/plan-audit-iter4.md`, audited 96e799fb3c8081597d88a36cbf67bff079ecc73f: overall_score 0.90, must_pass_failed 0, blocking_count 0, plan_artifact_hash b169e77e7e900ac758efbe059b79f61e2302f25eba2f14ddf76884812ae5f6da; the required codex leg returned `fail` on both runs, receipts rcpt-e6f0d86fa24a064dca480a46 and rcpt-6bf047de21e97b28beb61bc7). Rule C allows no further audit round after the delta: the outcome is hold and shrink. The `plan-audit → kickoff` edge needs a PASS-family verdict file, so the lane does not enter kickoff and does not work around the required gate.
What the delta found: D29, D30 and D31 are closed (the author's two deviations were verified against the code by the auditor). The codex failures sit on hardening gaps adjacent to the delta: D38 (SHOULD-FIX) T29's with-remote landing test still reads through the plain `gitRead`/`verifyPushed`, so a replacement ref can make an unlanded merge read as landed; D39 (SHOULD-FIX) in the no-configuration case any local branch other than the card's own satisfies the containment test, and a lane can create one; D40 (SHOULD-FIX) `diff.ignoreSubmodules=all` hides a gitlink change from the evidence diff (no `.gitmodules` here, so nothing is exposed today); D41, D42 minor. The auditor reads the SPEC as sound enough for run phase and D38-D40 as one sentence and one row each; the lane does not act on that reading, since the gate label decides.
Shrink proposal (RECORDED ONLY, NOT ISSUED; whether to issue it, and whether to change the card's completion criteria, is the leader's and operator's decision): every blocker of iterations 1-4 and the debt KD-1 sit in one cluster, the dependency chain `M1 → M2a → M2b` of plan.md §A (remove T18 and the T19 reservation and declare T29; the automatic receipt and its revocation; lane own-card done with the both-records close). The rest of the SPEC stands without it: M3a/M3b (the spec_id setter) have no dependency on that cluster, M4 → M5 → M6 → M7 → M8 (the six stage verbs, the warning alias, the slash command and skills, the substitution script and its application) form their own chain, and M9 (AGENTS.md, rule statements, closing guards) needs M8, M2b and M1 today, so it would have to drop the T18/T19 and lane-own-card wording. Proposed shape: t1641 keeps M3a, M3b, M4-M8 and a trimmed M9; a new card takes M1, M2a, M2b with D38-D40 added as one sentence and one row each. Cost to name: the card text's three completion items 'T18/T19 removal', 'lane own-card done exception' and 'done automatic for the most recent 10 landed cards' would move to the new card, and the ten-card criterion reads low until lanes commit their evidence (D35, `.moai/reports/*` is gitignored by operator order, KD-1).
whom: leader. recheck: each cron wake reads the leader's reply and the decision board; the lane changes nothing in the SPEC and opens no further audit round without a recorded ruling.

## Shrunk-scope revision 5 (2026-10-19, manager-spec, operator board d-20261018T214213Z-94a7)

Authority: the operator's ruling of 10-19 (AskUserQuestion; board record `d-20261018T214213Z-94a7`, scope `card:t1641`, read this round with `moai decision read --scope card:t1641`, which printed `board=ok unparseable=0 records=130` and the record verbatim). Scope: SPEC artifacts only — spec.md, plan.md, acceptance.md, decision-index.md revised in place; no code under internal/, scripts/, mods/, templates touched; no plan-audit*.md written (the delta audit is the auditor's; the lane runs it after this commit). Start state: HEAD `3c1809e2f1`, branch `WT-todo-unify`, clean (`git status --short` empty; `git diff --stat 3dc1052f6f 3c1809e2f1 -- . ':!.moai/specs/SPEC-TODO-UNIFY-001' ':!.moai/reports/t1641'` printed nothing). This section is committed by the revision-5 commit itself; the commit sha is read from the branch log after it lands (a commit cannot carry its own sha — the same convention as the revision-4 section, which records its start HEAD `f40740ea61` and not its own commit).

### Identifier outcome

- Requirements 25 → **15**: retired REQ-TU-002, REQ-TU-003 (the warning design) and REQ-TU-015 to REQ-TU-021, REQ-TU-025 (the done path). Surviving: REQ-TU-001, REQ-TU-004 to REQ-TU-014, REQ-TU-022, REQ-TU-023, REQ-TU-024.
- Criteria 25 → **15**: retired AC-TU-001, AC-TU-002 (warning) and AC-TU-014 to AC-TU-020, AC-TU-023 (done path). Surviving: AC-TU-003 to AC-TU-013, AC-TU-021, AC-TU-022, AC-TU-025.
- Numbering: identifiers kept stable, gaps stand (002, 003, 015-021, 025 in REQ; 001, 002, 014-020, 023 in AC). Chosen because `moai spec lint` carries no contiguity rule (checked in `internal/spec/lint.go`: format, uniqueness, and an advisory AC→REQ coverage rule only — `CoverageRule` emits `CoverageIncomplete` as a warning, never an error), because no retired identifier is reused with a new meaning (REQ-TU-001's revised text continues the same subject — the gtd/todo relationship — under the operator's reversal), and because renumbering 15 survivors would break every cross-reference in design.md/research.md/decision-index.md for zero lint benefit.
- Sets verified this round: `grep -c -E '^\*\*REQ-TU-[0-9]+ ' spec.md` printed 15; `grep -c -E '^\*\*AC-TU-[0-9]+ ' acceptance.md` printed 15 (after removing a leftover duplicate AC-TU-004 definition the first editing pass left; 16 was observed once and repaired); every surviving REQ is mapped by at least one AC and every AC maps surviving REQs (001←003, 004←004, 005←005, 006←006, 007←007, 008←008, 009+010←009, 011←010, 012←011, 013←012, 014←013, 022+023←021 and 025, 024←022). `grep -c 'lane-judgment, leader-confirmed 10-18 (conditional)' decision-index.md` printed 1 (the Q5 marker, byte-identical, untouched). `git grep -c 'NEEDS CLARIFICATION' -- .` printed nothing, exit 1.
- Dangling-reference sweep: every remaining mention of a retired identifier sits in an explicitly-marked context (spec.md:121 the [MOVED to t1878] note; plan.md:40 the Q2-superseded row and :76 the withdrawn-milestones note; acceptance.md:68 the Track-B-moved paragraph; decision-index.md Q2/Q4/Q5/Q6/Q7 rows carrying their Superseded/Moved markers, and the HISTORY table). No live requirement, criterion, milestone, or risk text cites a retired identifier.

### Where the four clusters landed (file:line as committed)

1. **Removed the M1 → M2a → M2b chain (moved to t1878):** spec.md:36 (HISTORY 0.5.0), :43 (§1.1 D4 bullet marked moved), :54 (§1.2 item 6 closing sentence), :60 (§1.3), :64-69 (§2 two tracks, counts 15/15), :101 (Track C renumbered 3.3→3.2 after 3.2/3.4 deletion), :119-121 (Out of Scope moved note `[MOVED to t1878: ...]` naming the new card, KD-1 and D38-D40), :131-135 (the FCR-014, fourth-fact, T16, observers, and committed-evidence bullets rewritten onto t1878); plan.md:3 (header revision-5 paragraph), :11-19 (§A table — ranks renumbered 1-7, the done-path cluster moved with Q4-Q7), :23-29 (graph M3a→M3b, M4→M5→M6→M7→M8→M9, arrow reasons rewritten), :74-76 (M1/M2a/M2b withdrawn note carrying D29-D31, D38-D40, KD-1 pointers), :40-47 (§B rows superseded/moved), :176 (§H known-debt pointer replacing KD-1's full text); acceptance.md:7 (revision-5 paragraph), :11 (numbering note extended), :66-68 (Track B section replaced by the withdrawn-criteria paragraph); decision-index.md:15 (revision-5 paragraph), :69 (Q4 moved), :83 (Q5 moved — marker line untouched), :97 (Q6 moved), :111 (Q7 moved), :161 (KD-1 moved).
2. **CLI-alias policy reversed (immediate removal, zero residual):** spec.md:77 (REQ-TU-001 rewritten: gtd tree removed, unknown command), :85 (REQ-TU-007 rewritten: /moai:gtd file deleted, moai-gtd skill leaves the publisher set, workflow document renamed to workflows/todo.md), :89 (REQ-TU-009 classes extended), :93 (REQ-TU-011 rewritten: the four-pattern grep prints nothing, no allowlist; baseline 397 lines / 101 files at `3c1809e2f1`), :95 (REQ-TU-012 narrowed), :99 (REQ-TU-014: no gtd spelling anywhere in AGENTS.md); plan.md:103-110 (M5 rewritten as deletion + caller migration), :112-119 (M6 carries the workflow-document rename), :121-127 (M7 script classes), :129-132 (M8 batches: fixture texts binding_ledger.json + roles tomls migrate, AGENTS.md pair waits for M9), :186 (§J anti-pattern: no warning alias, no grace); acceptance.md:29 (AC-TU-003 rewritten onto REQ-TU-001), :34-38 (AC-TU-004 rewritten: 33 on the single tree, parity window M4-M5), :40-43 (AC-TU-005 floors re-derived for todo-only tokens), :49 (AC-TU-007 rewritten), :61-64 (AC-TU-010 rewritten: the operator's grep, 0 target, this revision's 397-hit RED-now), :87 (AC-TU-011: binding_ledger.json unpinned, paths keys name the new path), :94 (AC-TU-013: no alias statement at all); decision-index.md:42 (Q2 superseded), :139-157 (Q10 row).
3. **M9 trimmed:** plan.md:134-138 (content = AGENTS.md row without alias language + the rule-family statements naming moai todo + policies projection; guards carry the last expected red; the T18/T19 and lane-own-card wording is gone).
4. **Operator ruling recorded:** decision-index.md:139-157 (Q10: label FOUNDER, operator-decided, board id, essence — shrink accepted + gtd immediate removal + one delta audit, rule C counter restarts; the `gtd\.md` consequence naming the workflow-document rename as a recorded consequence, with the persisted formats untouched).

### Verification this round (commands run, output seen; tree HEAD `3c1809e2f1`)

- Lint (installed binary per the dispatch's cheap alternative; `moai version` answered, so no build was run and no resource admission was needed): `moai spec lint SPEC-TODO-UNIFY-001` printed one INFO row — `INFO OwnershipTransitionUnmeasured ... commit 63829fe20a0f13fa93e974fe9331bb455ccb1009 (docs(t1641): SPEC-TODO-UNIFY-001 plan artifacts (card t1641)) has no Authored-By-Agent trailer — ownership transition unmeasured` — and the summary line `0 error(s), 0 warning(s)`, `lint-exit=0`; `moai spec lint --strict SPEC-TODO-UNIFY-001` printed the same INFO and `0 error(s), 0 warning(s)`, `strict-exit=0`.
- RED-now observation for the removal criterion (AC-TU-003): `moai gtd --help` through the installed build `ee981a932-dirty` (an ancestor, per research.md §R.1) exited 0 and printed the compatibility-alias help text — the tree still carries the alias, which is the red the revision removes.
- Residual-grep baseline re-measured with the completion judgment's own command: `git grep -E 'moai gtd|/moai:gtd|moai:gtd|gtd\.md' -- . ':!.moai/specs' ':!.moai/reports' ':!CHANGELOG.md' ':!*.jsonl'` printed **397 matching lines in 101 tracked files** at `3c1809e2f1`. The operator's ruling records 387 hits measured 10-19; the difference is not explained by this round (different tree state or exclusion shape) and is recorded, not reconciled: the target 0 is unchanged, and both figures are named in REQ-TU-011 and AC-TU-010 with the command verbatim so the run phase re-measures rather than re-cites.
- Frontmatter: spec.md `status:` untouched (`draft` as found); `updated:` refreshed to 2026-10-19; `version:` 0.4.0 → 0.5.0; title and tags revised to the shrunk scope. plan.md and acceptance.md carry no frontmatter (stateless per the frontmatter schema rule) — nothing to refresh there.

### Gaps (not observed)

- design.md and research.md were NOT edited this round (not in this delegation's write set): both still describe the moved M1 → M2a → M2b machinery in their §D.5/§D.6/§R.13-§R.16 regions and cite retired REQ identifiers there. The revised spec/plan/acceptance/decision-index carry the moved notes that govern them, and t1878 owns authoring the moved cluster's artifacts, but until t1878's SPEC exists the two artifacts describe more than this card delivers — the delta audit should read them through the moved notes.
- The installed `moai` used for lint is `ee981a932-dirty`, a strict ancestor 516 commits behind (research.md §R.1); the SPEC's DoD requires a tree-built binary at run time, so this lint result is the plan-phase check, not the DoD's.
- The AC-TU-005 floor argument (that the 326-line count survives M8 because substituted lines carry the todo spelling) is a reading of §R.5b's measurement, not a re-measurement of the post-substitution corpus.
- The M8 unit size (~380 lines / ~95 files) is derived from this round's grep; it is re-measured at pre-flight after absorbing main, and the file-level classification of all 101 files into the batch classes was spot-checked (binding_ledger.json 46, .moai/workflows/gtd.md 46, docs-site 4×12, rules/policies/manifest/tests for the rest), not enumerated file-by-file.
- Whether the publisher's exact set (17 names) loses exactly one member when the gtd command file is deleted was not measured (the publisher runs in the project build); AC-TU-007 states the equality-with-publisher-output form so the run phase measures it.
- The delta audit itself was not run (the lane runs it after this commit); no plan-audit*.md file was written by this round.
- The operator's 387-hit measurement could not be reproduced exactly here (397 with the command as written); recorded as an open count discrepancy, not resolved by adjusting the command.

## Delta audit outcome (2026-10-19, plan-auditor — the operator's one round)

Verdict file: `.moai/reports/t1641/plan-audit-delta.md` (gitignored, uncommitted, per the
standing verdict-file ruling). **PASS-WITH-DEBT 0.88** (Tier L threshold 0.85 met),
must_pass_failed 0, blocking_count 0. audited_sha `436e77f3485183b81586b6319d5b16b9d3585bde`
(tree HEAD verified unmoved, clean). plan_artifact_hash
`ffba0a85e253cf3e9185d4123c62b629872bb6c61e347c9585c4550b38b4f328` — from two agreeing
transcriptions of `internal/runtime/audit_cache.go:92-141` ComputeHash; the MCP `audit_cache`
tool is absent from this session, disclosed in the file.

All six delta-scope items CLEAN; every class-A RED re-executed and reproduced (397/101 at
both `3c1809e2f1` and HEAD; AC-TU-008 grep 8; AGENTS.md 1 hit / 23,780 bytes; `moai gtd
--help` exit 0 through the installed ancestor build; `moai todo reflect` exit 1). One AC-set
correction to the dispatch's summary: the surviving set includes AC-TU-024 (no-skip guard);
artifacts are consistent at 15/15.

Carried (run-phase obligations, none a blocker):
- D43 (SHOULD-FIX): M5/M6 file lists omit the two gtd slash-command file deletions
  (`internal/template/templates/.claude/commands/moai/gtd.md` + deployed
  `.claude/commands/moai/gtd.md`); M5 must delete both by explicit path.
- D44 (SHOULD-FIX): AC-TU-008's green-path cell names M7 while plan revision 5 moved those
  consumer edits to M5; plan M7's stale "Gap carried" note rides M5 now.
- CONV-1 (mechanical): the required codex leg did not run — the auditor session has no moai
  MCP audit tools; `required_backend: codex inconclusive` recorded, `convergence_overall`
  deliberately omitted. **Plan admission will refuse this verdict fail-closed on the
  receipt**; disposition is the leader's (bridge leg at kickoff, or the ceiling's
  label-only admission arm). The lane does not advance to kickoff on this file alone.

## Proxy codex leg outcome and repair round (2026-10-19, leader ruling)

The proxy audit (standing rule bc8c) returned **codex fail** — receipt `rcpt-54ca7c42603570ff69d6589e`,
bound to HEAD `0fa5dc2892` (SPEC artifacts byte-identical at the audited `436e77f348`; the
proxy verified the identity). The proxy appended its section to `plan-audit-delta.md` and
updated `required_backend: codex fail` + `audit_receipt` (all other keys byte-identical).
Substantive verdict: FAIL — kickoff closed. This consumed round 1 of the counter restarted
by operator act 94a7; a third FAIL in this counter holds the card (leader ruling, 10-19).

Blocking findings (codex, proxy-confirmed where marked):
- **P1** REQ-TU-011's grep-0 target is unsatisfiable together with source preservation:
  `internal/factory/gtd_engage.go:28` hashes the persisted card-UUID namespace string
  `"moai:gtd-card:"`, which the `moai:gtd` alternative matches; substituting it changes
  every derived card UUID, against REQ-TU-012 (read-confirmed by the proxy).
- **P2** plan.md M6 renames the workflow doc and rewrites `TestWorkflowRulePathsPinned`, but
  the rule-file `paths:` updates sit in M8a — M6's mandatory checks cannot pass on M6's own
  edits alone.
- **P3** (optional) M8's GREEN wording conflicts with AC-TU-010/M9 keeping
  `TestNoResidualGTDSpelling` RED until M9.

Leader-ordered repair (normal round; tree writes resumed):
1. REQ-TU-011 gains an explicit UUID-namespace preservation exception — a file:literal
   allowlist entry for `internal/factory/gtd_engage.go` / `"moai:gtd-card:"`, plus a
   criterion row (with mutant) that card-UUID derivation bytes stay unchanged.
2. The rule-file `paths:` edits move from M8a into M6 (renames and their consumer-path
   updates land in one milestone).
3. manager-spec commits the repair.
4. One delta audit with codex required, through a fresh bc8c proxy brief.

## Repair revision 6 (2026-10-19, manager-spec, leader ruling on rcpt-54ca7c42603570ff69d6589e)

Authority: the leader's repair ruling of 10-19 (decision-index.md Q11) on the proxy codex leg (standing rule bc8c, receipt `rcpt-54ca7c42603570ff69d6589e`, bound to HEAD `0fa5dc2892`; findings verbatim in `plan-audit-delta.md` § "Codex leg (proxy, 2026-10-19)"). Scope: the two blocking findings P1-P2 plus the optional P3 wording fix, exactly as ordered — spec.md, plan.md, acceptance.md, decision-index.md revised in place; no code, template, or non-SPEC file touched; no `plan-audit*.md` written (the next delta audit with codex required is the auditor's, through a fresh bc8c proxy brief). Start state: HEAD `01e876706f`, branch `WT-todo-unify`, clean (`git status --short` empty before editing). This section is committed by the revision-6 commit itself; the commit sha is read from the branch log after it lands (a commit cannot carry its own sha — the same convention as the revision-4 and revision-5 records, which record their start HEAD).

### Where P1 landed (file:line)

- spec.md:94 (REQ-TU-011): the grep-0 target restated as "nothing outside one named preservation exception"; the exception written as the file:literal pair `internal/factory/gtd_engage.go` × `"moai:gtd-card:"` — the card-UUID hash namespace (`gtdCardUUID` derives every persisted card UUID from `sha256("moai:gtd-card:" + itemID)`), which the `moai:gtd` alternative matches and REQ-TU-012 freezes — minimal (this file+literal only; any other matching line in the file, and any other file, stays residual), named as an exception with its reason, not a pattern carve-out; "grep-0 means zero residual hits outside the named exception".
- acceptance.md:63 (AC-TU-010): the completion grep gains exactly one path exclusion (`':!internal/factory/gtd_engage.go'`) plus the excepted-file row — a read of the excepted file finds no matching line other than the namespace literal itself; the two clauses together are the mechanical form of "zero hits outside the named exception". The RED-now cell keeps the original four-exclusion command's measured 397/101 and names the exception's line (`gtd_engage.go:28`) as one of them.
- acceptance.md:89-91 (AC-TU-011, extended — no new AC id): the byte-invariance row — given the `gtd_engage.go` source, the namespace input of `gtdCardUUID` is the byte literal `moai:gtd-card:` and the hash input is that literal concatenated with the raw item id; test `TestGTDCardUUIDNamespaceInvariant` (`internal/factory`, created at M5, class G) asserts literal + construction with a known-answer UUID vector; the mutant row names `moai:todo-card:`, a reordered/interpolated construction, or any edit to the literal as failing — even where the diff pin is blind because a later change moved the file.

### Where P2 landed (file:line)

- plan.md:111-117 (M6): title gains "rule paths"; RED adds the rewritten pin test's red reason (the rule sources still carry the old path) and the two document tests rewritten first to read `workflows/todo.md` (their current fixture paths read `workflows/gtd.md` — `todo_skill_doc_test.go:22-23`, `todo_skill_doc_parity_test.go:24-25`); GREEN carries the five rule sources' `paths:` key updates (measured at HEAD: line 3 of `cross-session-messaging.md`, `factory-dispatch.md`, `factory-dispatch-cards.md`, `factory-dispatch-detail.md`, `factory-dispatch-gates.md`, template source + deployed mirror each, glob-shape-preserving segment swap) plus the `factory-dispatch-mechanics.md:199` prose cross-reference, and states the two document tests' post-milestone form; Files grows to about 24 items with the M6b split seam named (rule keys + mechanics + doc-test rewrites; neither M7 nor M8 starts before both land).
- plan.md:122 (M7 GREEN): the path-string class loses its rule-source members (M6 updated them).
- plan.md:129 (M8a): the batch's rule-source entry narrowed to rule prose — the `paths:` keys are absent from this batch, already new-path from M6.
- plan.md:60 (pre-flight item 7) and plan.md:128 (M8 unit/size): the exception file `internal/factory/gtd_engage.go` is named as excluded from the M8 input set (preserved Go source, one line, no milestone edits it).
- plan.md:166 (§H risk row) and acceptance.md:109 (§3 edge case): the "M6/M8 together" wording replaced — rename + consumer edits + rule `paths:` keys in M6; residual path strings in test prose/policies/codemaps in M8's sweep.

### P3: taken

Contained wording fix, as the ruling allowed. plan.md:131 (M8 Mechanism): "a green `TestNoResidualGTDSpelling`" after M8 replaced by "zero replacements and no hit outside the always-loaded pair" — M8 closes with script idempotence observed and the residual test recorded RED (the two `AGENTS.md` files keep it red until M9); the full-residual GREEN is M9's completion condition (AC-TU-010 already stated the flip story; M9's guards already own the green). No expansion beyond that one sentence.

### decision-index.md and HISTORY

- decision-index.md:3 (intro names Q11), :17 (revision-6 paragraph: the two findings, the ordered repair, round accounting — the FAIL consumed round 1 of the 94a7 counter, this repair round is round 2, a third FAIL holds), :154-165 (Q11 row: decided_by leader, evidence the receipt + the re-measured `gtd_engage.go:28` line + the pin-test mechanics, essence P1 exemption + P2 M6/M8a reordering, alternates rejected, Operator verdict `LEADER-DECIDED 2026-10-19`).
- spec.md:4 (`version: "0.6.0"`) and spec.md:37 (HISTORY 0.6.0 row). `status:` untouched (`draft` as found); `updated:` already 2026-10-19, unchanged.

### Identifier outcome

No identifier added, removed, or renumbered: 15 REQ (spec.md) and 15 AC (acceptance.md), verified below; the P1 criteria extend AC-TU-010 and AC-TU-011 in place (no new id, so no 1:1 disturbance: AC-TU-010→REQ-TU-011, AC-TU-011→REQ-TU-012 unchanged). No machine-key-shaped line added anywhere (the receipt id appears only inside prose and `evidence_refs=`).

### Verification this round (commands run, output seen; tree HEAD `01e876706f`, working tree = the revision)

- P1 premise re-measured by this session, not cited from the proxy: `sed -n '20,34p' internal/factory/gtd_engage.go` shows `gtdCardUUID` with `sum := sha256.Sum256([]byte("moai:gtd-card:" + itemID))`; `git grep -n -E 'moai gtd|/moai:gtd|moai:gtd|gtd\.md' -- internal/factory/gtd_engage.go` prints `internal/factory/gtd_engage.go:28: sum := sha256.Sum256([]byte("moai:gtd-card:" + itemID))`, exit 0 — one hit, the line the exception shields.
- P2 premise re-measured: `git grep -n 'workflows/gtd.md' -- .claude/rules internal/template/templates/.claude/rules` prints the `paths:` key at line 3 of the five rule documents (deployed + template source = 10 lines) plus the `factory-dispatch-mechanics.md:199` prose line; `internal/template/workflow_rule_paths_pinned_test.go:24-58` read — the test compares the embedded tree's `paths:` value against the recorded expectation at `:31`, so a rewritten expectation cannot pass without the same-milestone rule edit; the two M6 doc tests' fixture paths confirmed at `todo_skill_doc_test.go:22-23` and `todo_skill_doc_parity_test.go:24-25`. (The dispatch said "six rule files' `paths:` keys"; the tree measures five `paths:` keys + the sixth file's prose cross-reference — both are named in M6, the census stated as measured.)
- Lint (installed binary, both modes, on the files as committed): `moai spec lint SPEC-TODO-UNIFY-001` printed one INFO row — `INFO OwnershipTransitionUnmeasured ... commit 63829fe20a0f13fa93e974fe9331bb455ccb1009 (docs(t1641): SPEC-TODO-UNIFY-001 plan artifacts (card t1641)) has no Authored-By-Agent trailer — ownership transition unmeasured` — and `0 error(s), 0 warning(s)`, `lint-exit=0`; `moai spec lint --strict SPEC-TODO-UNIFY-001` printed the same INFO and `0 error(s), 0 warning(s)`, `strict-exit=0`.
- Marker and counts: `git grep -c 'NEEDS CLARIFICATION' -- .moai/specs/SPEC-TODO-UNIFY-001` printed nothing, exit 1; `grep -c 'lane-judgment, leader-confirmed 10-18 (conditional)' decision-index.md` printed 1, and the line's md5 is identical at HEAD and in the working tree (`dadaf7ac41457efd0ace3ff487698a1e` both); `grep -c -E '^\*\*REQ-TU-[0-9]+ ' spec.md` printed 15; `grep -c -E '^\*\*AC-TU-[0-9]+ ' acceptance.md` printed 15.
- Scope: `git status --short` before this record listed exactly the four SPEC artifacts; no file outside `.moai/specs/SPEC-TODO-UNIFY-001/` and this record is staged.

### Gaps (not observed)

- No `go test` was run (card constraint): `TestGTDCardUUIDNamespaceInvariant` is a named criterion with a mutant, not an observation; its RED/GREEN evidence is M5's to record. Every claim about the pin-test mechanics is a static reading of `workflow_rule_paths_pinned_test.go`, not an executed selection.
- Lint is attributed to the installed ancestor build (research.md §R.1 lineage), per the dispatch's rule — the SPEC's DoD re-lints from a tree build at run time.
- The next delta audit (codex required, fresh bc8c proxy brief) is not this session's to run; no `plan-audit*.md` was written.
- design.md §D.3's substitution class table was not edited (not in this delegation's four-file write set); plan.md M7 now states the rule-source members' removal from the path-string class, and design.md's table remains read-through-the-plan text until its owner revises it.
- The dispatch's "six rule files" vs the measured five-`paths:`-keys census is recorded above, not reconciled by editing the plan to a bare "six" — the plan names the measured set.

## Round-2 proxy audit FAIL and final repair round (2026-10-19)

The round-2 proxy audit returned **FAIL 0.80** (`plan-audit-delta2.md`: verdict FAIL,
must_pass_failed 1, blocking_count 3, codex leg fail, receipt
`rcpt-03225d18a7dc4a018105dc3d`, audited_sha `a286e02696…`). Verified by the auditor: the
round-1 P1 and P2 repairs hold. New blockers (the auditor drafted the required fixes):
- **D1 (critical, MP-9)**: plan.md:115's M6 split seam (M6b) re-orders the rule `paths:`
  edits after the rename, against acceptance.md:109's same-milestone clause. Fix: keep
  rename + pin-test rewrite + five `paths:` keys + mechanics:199 cross-reference + two
  doc-test rewrites in ONE unit; if anything splits, split off the slash-command,
  router-skill and published-skills edits instead.
- **D2 (major)**: AC-TU-011's namespace test is placed at M5 by acceptance.md:90 but plan
  M5 carries none of it; two unnamed collisions — a test file under `internal/factory/gtd_*.go`
  trips AC-TU-011's own no-change guard, and a contiguous `"moai:gtd-card:"` literal in the
  test trips AC-TU-010's residual grep. Fix: add the test to M5 (Files/Packages
  `internal/factory`, name-exact selection), name a file outside the glob (e.g.
  `card_uuid_namespace_test.go`), require the literal assembled non-contiguously or matched
  against bytes read from `gtd_engage.go`.
- **D3 (major)**: `TestNoResidualGTDSpelling` (and any tracked planted fixture) contains its
  own four-pattern alternation, so the completion grep can never reach exit 1. Fix: state in
  AC-TU-010 and plan M8 RED that the test assembles its pattern from non-contiguous parts
  and writes its fixture to a temp directory at run time; optionally re-run the grep over
  the test file as a row.

Leader: the counter's round 3 is the last — a FAIL there holds the card.

## MCP reconnection confirmed (2026-10-19, board d-20261018T235443Z-39d8)

Per the leader's verification request: one `mcp__moai__session_list` call from this session
at 2026-10-19T00:1xZ UTC — **SUCCESS** (server responded; result 97,136 characters of
session-list JSON, saved to the tool-results file by the display-size wrapper). The moai MCP
server is reconnected in this lane session. Per standing rule bc8c the audit proxy ends
automatically: the lane runs the t1641 round-3 (final) delta audit directly through its own
plan-auditor spawn with in-session `audit_multi` (codex required), and t1862's landing can
use the MCP `factory_complete` when its turn arrives.

## Final repair revision 7 (2026-10-19, manager-spec)

Authority: the leader's final repair ruling of 10-19 on the round-2 proxy audit — receipt
`rcpt-03225d18a7dc4a018105dc3d`, bound to HEAD `a286e02696`, verdict FAIL 0.80, blocking D1-D3
(`plan-audit-delta2.md` § "Defects Found" and § "Recommendation") — as restated in the dispatch:
round 3 of the 94a7 counter, the LAST; a FAIL at the next audit holds the card. Scope: the
auditor's drafted fixes verbatim, no more — spec.md, plan.md, acceptance.md, decision-index.md
revised in place; no code or template touched; no `plan-audit*.md` written (the round-3 delta
audit is the auditor's, run after this commit). Start state: HEAD `f2ba34cdff`, branch
`WT-todo-unify`, clean (`git status --short` empty before editing). Mid-round, the concurrent
record commit `060f568e48` (this record file only — a 10-line MCP-reconnection note) landed on
the branch; `git diff --name-only f2ba34cdff 060f568e48` printed only
`.moai/reports/t1641/progress.md`, so the four SPEC artifacts are byte-identical between the two
and every measurement below stands against both. This section is committed by the revision-7
commit itself; the commit sha is read from the branch log after it lands (a commit cannot carry
its own sha — the same convention as the revision-4, revision-5, and revision-6 records, which
record their start HEAD).

### Per-fix file:line (working tree at this commit)

- **D1 — the M6 seam keeps one unit (plan.md:115, the seam sentence rewritten).** The first unit
  keeps, together, the workflow-document rename in its three locations, the
  `TestWorkflowRulePathsPinned` rewrite, the five rule `paths:` keys with their deployed mirrors,
  the `factory-dispatch-mechanics.md` cross-reference, the two document-test path rewrites, and
  the embed-manifest regeneration; M6b carries ONLY the slash-command, router-skill, and
  published-skills edits, landing immediately after the first unit and before M7; the golden test
  is red until M6b closes it; neither M7 nor M8 starts before M6b. plan.md:114 (M6 GREEN) carries
  the matching split note; acceptance.md:111 (§3 edge case) carries the revision-7 note that the
  same-milestone clause holds unchanged on the split branch; the auditor's D1 fallback (green at
  M6b's close + an amended acceptance clause) was not needed — one unit was feasible.
- **D2 — the namespace test planned in M5 with both collisions named (plan.md:105 RED item 5,
  :106 GREEN, :107 Files, :108 Packages, :109 name-exact selection).** File
  `internal/factory/card_uuid_namespace_test.go` — outside the `gtd_*.go` glob AC-TU-011's diff
  pin guards; package `internal/factory`; selector `TestGTDCardUUIDNamespaceInvariant` (created
  green, class G, not an expected red); expected literal assembled non-contiguously
  (`"moai:" + "gtd-card:"`) or compared against bytes read from `gtd_engage.go`. Acceptance side:
  acceptance.md:91 (AC-TU-011's revision-7 sentence) and :92 (the green cell names the file).
- **D3 — the residual test's self-collision removed (acceptance.md:65, AC-TU-010's revision-7
  constraints + the confirming row; plan.md:130, M8 RED).** The test source assembles its
  four-pattern alternation from non-contiguous parts (no line of its file matches any pattern);
  its planted-mutant fixture is written into a temporary directory at run time — no tracked
  fixture; the completion judgment adds the grep's re-run over the test's own file
  (`internal/template/no_residual_gtd_spelling_test.go`, the path plan.md M8 names) expecting
  nothing printed and exit 1. acceptance.md:66 (green cell) names the test file.
- **Bookkeeping.** spec.md:4 version `0.7.0`, spec.md:38 the HISTORY row (15 and 15, no
  identifier change); decision-index.md:3 header (a twelfth), :19 the Revision 7 paragraph, :168
  the Q12 row (decided_by leader 10-19, evidence `rcpt-03225d18a7dc4a018105dc3d`, final repair
  before the counter's last round); `updated:` stays 2026-10-19 (correct); `status:` untouched
  (draft); plan.md:195 §K names Q11 and Q12. Optional D4/D5/D6 were NOT applied (the
  Recommendation marks them optional; the dispatch scoped this round to D1-D3).

### Verification this round (commands run, output seen; working tree = the revision; HEAD `f2ba34cdff` at round start and `060f568e48` at the batch — the four SPEC artifacts byte-identical across both, see the start-state note)

- `moai spec lint SPEC-TODO-UNIFY-001` (installed binary): one INFO row — `INFO
  OwnershipTransitionUnmeasured ... commit 63829fe20a0f13fa93e974fe9331bb455ccb1009
  (docs(t1641): SPEC-TODO-UNIFY-001 plan artifacts (card t1641)) has no Authored-By-Agent trailer
  — ownership transition unmeasured` — and `0 error(s), 0 warning(s)`, `lint-exit=0`;
  `moai spec lint --strict SPEC-TODO-UNIFY-001` printed the same INFO and `0 error(s), 0
  warning(s)`, `strict-exit=0`.
- Marker and counts: `grep -rn 'NEEDS CLARIFICATION' plan.md research.md` printed nothing,
  exit 1; `grep -c 'lane-judgment, leader-confirmed 10-18 (conditional)' decision-index.md`
  printed 1, and the line's md5 is identical at HEAD and in the working tree
  (`dadaf7ac41457efd0ace3ff487698a1e` both) — byte-identical exactly once;
  `grep -c -E '^\*\*REQ-TU-[0-9]+ ' spec.md` printed 15; `grep -c -E '^\*\*AC-TU-[0-9]+ '
  acceptance.md` printed 15.
- Scope: only the four SPEC artifacts and this record are staged; `git status --short` before
  staging listed exactly them.

### Gaps (not observed)

- The round-3 delta audit with codex required is not this session's to run; no `plan-audit*.md`
  was written. A FAIL there holds the card.
- The confirming row and both construction constraints are named criteria, not executions: no
  `go test` was run (card constraint) — `TestGTDCardUUIDNamespaceInvariant` and
  `TestNoResidualGTDSpelling` remain class G/B criteria whose evidence is M5/M8's to record.
- Lint is attributed to the installed ancestor build (research.md §R.1 lineage), per the
  dispatch's rule — the SPEC's DoD re-lints from a tree build at run time.
- The commit sha of this revision is not in this section's text (a commit cannot carry its own
  sha); it is read from the branch log after this commit lands — the child of `060f568e48` on
  `WT-todo-unify`.
