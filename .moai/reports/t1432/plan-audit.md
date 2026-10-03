auditor-model: claude-sonnet-5-5[1m] (Sonnet 5.5)

verdict: FAIL
audited_sha: 69ddcc73ae1b6e776d2cc170c3814160788dd1ef

# SPEC Review Report: SPEC-HARNESS-RETENTION-HARDEN-001 (card t1432)
Iteration: 1/3
Verdict: FAIL
Overall Score: 0.75 (Tier M threshold 0.80; four must-pass-class findings below force FAIL independently of the score)
Plan Artifact Hash: not computed with the runtime `ComputeHash`; per-file SHA-256 prefixes of the audited bytes: spec.md `bbdaa345cd77c97c`, plan.md `153c52b1f55b965b`, acceptance.md `6e8da8421d78f07f` (decision-index.md and progress.md also read)
Auditor Version: plan-auditor (no version string available in this run)

Reasoning context ignored per M1 Context Isolation (none was passed; only the committed artifacts and the code named in the dispatch were read).

## Tree attribution

| Command | Observed |
|---|---|
| `git rev-parse --show-toplevel` | `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1432` |
| `git branch --show-current` | `WT-harness-retention-debt` |
| `git rev-parse --short HEAD` / `git status --short \| wc -l` | `69ddcc73a` / `0` (committed state audited; HEAD~1 = `1e2151a38`, the stated base) |
| `moai version` | `v3.2.0-rc.26   archive/t1401-293-g45600e4ee   built 2026-10-02T14:42:32Z` |
| `git merge-base --is-ancestor 45600e4ee 69ddcc73a` / `git rev-list --count 45600e4ee..69ddcc73a` | exit 0 / `127` |

The installed `moai` build (`45600e4ee`) is a strict ancestor of the tree HEAD, 127 commits behind (verification-claim-integrity §2.2). Its `moai spec lint` result is therefore a measurement by an older build, not by a build of this tree.

## Must-Pass Results

- [PASS] MP-1 REQ number consistency: `grep -n '^- \*\*REQ-HRH' spec.md` lists REQ-HRH-001 .. REQ-HRH-012 on spec.md:110-121, sequential, no gap, no duplicate, uniform zero padding. AC-HRH-001 .. 015 likewise (acceptance.md traceability table).
- [PASS] MP-2 GEARS compliance (judged on the `REQ-XXX` requirement layer in spec.md §C; ACs are Given-When-Then and were NOT penalized): REQ-001/002/003/004/010 Event-driven "When ..., the ... shall"; REQ-005..009, 011, 012 Ubiquitous "The ... shall". No informal modality. Clarity notes (not MP-2): REQ-010/011/012 take a test as subject and REQ-009 takes "the change set" (verification-layer and process content sitting in the requirement layer); REQ-004 embeds mechanism ("final reading of the log's size").
- [PASS] MP-3 frontmatter: all 12 canonical fields present with correct types (spec.md:2-13: id, title, version "0.1.0", status draft, created/updated ISO, author, priority P2, phase "v3.2.0 target", module, lifecycle spec-anchored, tags string); extra `tier`, `card`, `related_specs` accepted. `moai spec lint SPEC-HARNESS-RETENTION-HARDEN-001` printed `✓ No findings — all SPEC documents are valid` (judged by the build named above). plan.md and acceptance.md carry no `status:` frontmatter (statelessness holds).
- [N/A] MP-4 language neutrality: single-language (Go, `internal/harness`) SPEC.
- [PASS] MP-5 D7: the only other SPEC id referenced in spec.md is SPEC-AGENT-TEAM-RETIRE-001; its `status:` is `completed` (spec.md:5 of that SPEC), not retired/superseded/archived, so no REVIEW line applies; no BLOCKING.
- [FAIL] MP-6 D8 cross-platform discipline: the D8 verb is an `awk` program and the worktree guard refused to run it (see Gaps). Re-done by hand: `grep -n 'syscall' spec.md` -> lines 47, 85, 95 in three different sections (§A Background, D3, D4); `grep -n '//go:build\|cross-platform exemption\|EXCL' spec.md` -> no match (rc=1). Per D8-3 each section references `syscall` with neither a `//go:build` constraint nor an exemption clause = BLOCKING. Substance behind the literal result: the three mentions are descriptive (a stack frame name, "no syscall" in a cost cell, `syscall.O_NOFOLLOW` undefined on Windows), but the SPEC also prescribes NEW FIFO tests (AC-HRH-005, edit of the existing killed test for AC-HRH-012), which need `syscall.Mkfifo`, and acceptance.md:3 states only "FIFO-based tests skip on Windows" (a runtime skip). A runtime skip does not compile on Windows; `GOOS=windows go vet ./internal/harness/` (AC-HRH-011) compiles test files, so the missing `//go:build !windows` obligation for new test files is a lesson-#21-class gap the SPEC never states (the existing `retention_killed_test.go:1` carries `//go:build !windows`). See D4.
- [PASS] MP-7: `grep -rn 'NEEDS CLARIFICATION' plan.md` -> no match (rc=1); research.md absent (Tier M), nothing to scan there.
- [N/A] MP-8 RED-now cell re-execution: no acceptance criterion is classified release-blocking (`grep -rn -i 'release-blocking\|regression-guard' .moai/specs/SPEC-HARNESS-RETENTION-HARDEN-001/` -> no match), so the criterion's scope is empty. This is an N/A by classification, not a pass: the RED-now cells are in fact missing the four §2.1 elements and the cited commands cannot reproduce a RED today (observed below, D7).
- [FAIL] MP-9 cross-artifact ordering conflict: the CN-4 verb is an `awk` program and was refused by the worktree guard; read by hand. Plan M0 step 1 (plan.md:38) and acceptance DoD #1 (acceptance.md:144) cannot both hold together with AC-HRH-014. Quoted side by side below (D1). COLLECTED/CONFLICT lines were not produced mechanically; the finding is the auditor's own, per MP-9's candidate rule.

## Category Scores

| Dimension | Score | Rubric band | Evidence |
|---|---|---|---|
| Clarity | 0.75 | minor ambiguity in a few requirements | REQ-004 "after the retention pruner finished reading it" vs the prefix-measurement semantics of plan M1 (events appended during the read are not addressed by the REQ text); "at most one duplicate prune" (acceptance.md:132) versus the mechanism in D5 below; tail terminator undefined (D3). |
| Completeness | 0.75 | one non-critical gap | All sections present (HISTORY spec.md:21, background §A, decisions §B, REQs §C, constraints §D, `### Out of Scope — <topic>` x3 with bullets spec.md:132-149, risks §F; ACs in acceptance.md). Gaps: TOCTOU window and N-process heal burst not in §F (D5, D12); DoD #1 relies on an ignored path (D2). |
| Testability | 0.50 | several ACs need judgment or are vacuous today | RED commands return `ok ... [no tests to run]`, exit 0 on this tree (observed, D7); AC-008/009/010 grep sentinels admit trivial mutants (D8); AC-006(a) "equals what the baseline pruner writes" has no executable comparison; AC-014 binds a mechanism (D6). |
| Traceability | 1.00 | every REQ has an AC, no orphan | `grep -n '^| AC-HRH' acceptance.md`: AC-001..014 map 1:1 to REQ-001..012 (two ACs each for REQ-001 and REQ-004), AC-015 maps `REQ-HRH-001..004, -010`; `grep -o 'REQ-HRH-0[0-9][0-9]' acceptance.md \| sort -u` yields all 12 ids; no id outside 001-012. The traceability awk verb was refused (Gaps); this is a manual equivalent. AC-015's range notation is mapping-only; the REQs it names each have their own AC. |

Arithmetic mean of the four dimensions: (0.75 + 0.75 + 0.50 + 1.00) / 4 = 0.75, below the Tier M threshold 0.80.

## Premise honesty (task item 1) — which premises I re-observed and which I only read

I wrote throwaway probes outside the tree (scratchpad `probe/zz_probe_test.go`) and injected them with `go test -overlay` (nothing written into the tree; `git status --short | wc -l` = 0 afterwards).

```
$ unset MOAI_KANBAN ... && go test -count=1 -overlay <scratch>/overlay.json -run '^TestZZProbe' -v ./internal/harness/
PROBE-F4 err=<nil> victim_changed=true victim="2026-10-02T00:00:00Z"
PROBE-F7 uid=501 err=retention: prune state open failed: open .../usage-log.jsonl.prune-state: permission denied stale_still_in_log=true
PROBE-LOSS err=<nil> late_event_present=false fresh_present=true
PROBE-UNTERM err=<nil> ends_with_newline=true tail="...\"subject\":\"unterminated\"\n"
ok  	github.com/modu-ai/moai-adk/internal/harness	0.809s
```

| Premise | Label in SPEC | My status |
|---|---|---|
| F4 symlink followed and target truncated | OBSERVED | RE-OBSERVED (own probe, output identical in substance); cause READ at retention.go:121 |
| F7 unwritable state file disables retention | OBSERVED | RE-OBSERVED (uid 501, mode 0400, `permission denied`, stale event still in log); discard READ at observer.go:93 and :141 |
| Event loss during prune | OBSERVED | RE-OBSERVED (FIFO-blocked archive step, `late_event_present=false`) |
| F5 Windows lock in-process only | NOT REPRODUCED; shape READ | READ only (`lockfile_windows.go:26-41`, per-path `sync.Mutex` map). Label truthful. |
| F6 waiter has no timeout | NOT REPRODUCED; shape READ | READ only (`lockfile_unix.go:23-25` blocking `LOCK_EX`, no try-lock). Label truthful. |
| N1 FIFO drain hang | AUDIT-REPORTED; open READ | READ only (`retention_killed_test.go:43` blocking `O_RDONLY`); audit text confirmed at sync-audit-delta.md:82. NOT re-run. |
| N2 mB / mC survive | AUDIT-REPORTED | NOT re-run. One figure is misquoted: spec.md:48 says `pruneExclusive 85.0%`; the audit says `pruneExclusive 87.5%` (sync-audit-delta.md:100 and :164). See D10. |

No AC or REQ asserts a Windows runtime observation: REQ-HRH-007 labels the burst "not reproduced, not measured"; AC-HRH-011 states "Verification level: build and vet only". I also ran `GOOS=windows go doc syscall.O_NOFOLLOW` -> `doc: no symbol O_NOFOLLOW in package syscall` (confirms the spec.md:95 claim) and `GOOS=windows go vet ./internal/harness/ ./internal/lockfile/` -> no output (clean) on the unmodified tree.

Bounding facts (task item 2), each checked against the tree:
- (a) seven non-test importers of `internal/lockfile`: CONFIRMED (`grep -rl 'internal/lockfile"' --include='*.go'` minus tests: cli/glm_tools.go, cli/settings.go, cli/taskledger/taskledger.go, contract/receipt/store.go, escalation/lock.go, harness/retention.go, hook/agentmemory.go). `@MX:ANCHOR` in `lockfile_unix.go:17-20` names three.
- (b) REQ-ATR-001 / no silent LockFileEx upgrade: CONFIRMED (`lockfile_windows.go:13-16` `@MX:NOTE ... do NOT silently "upgrade" this to LockFileEx`; SPEC-AGENT-TEAM-RETIRE-001 `status: completed`, REQ-ATR-001 at its spec.md:90).
- (c) hook timeout 5 s, async: CONFIRMED at settings.json.tmpl (`"timeout": 5 ... "async": true` for the harness-observe blocks around :146-149 and :203-206). Not stated in the SPEC: these blocks sit inside `{{ if .HookOptIn.Enabled }}` (opt-in observer hooks).
- (d) append before prune: CONFIRMED (observer.go:86 `f.Write(data)`, :93 `PruneStaleEntries`; same at :135/:141).

## Design defaults (task item 3)

**D3 tail-carry.** The default does close the observed `late-event` loss (the late event lands after the measured prefix and is carried). What remains: events appended between the final size reading and the rename; events by a writer holding the old inode across the rename; stale-timestamped late events (stay until the next interval). REQ-006 discloses the window truthfully and the old "whole interval" sentence is the right sentence to remove. BUT the boundary rule as specified is wrong in one case: see D3 below (unterminated tail glues the next append). The measured baseline always ends the rewritten log with a newline (probe `ends_with_newline=true`), because `partitionEvents` + `overwriteWithEvents` normalize terminators; a verbatim tail copy removes that guarantee.

**D4 replace-on-fault.** Deleting the stamp file is within the card scope (it is a ~35-byte derived file in the project's own directory) and does not hit the "never simplify away safety" carve-out (removing a link, not its target, is security-positive). Findings: (1) the audit's own suggested fixes were non-destructive ("Lstat ... refuse a non-regular file", "read-only stamp check plus a logged warning", sync-audit.md:222,225) and the SPEC's default goes further than the audit asked while its own decision-index row Q4 (FOUNDER, anchor "(none)") says it "deletes a file the current user did not necessarily create" yet D4 is `operator-held: no`; (2) the heal-burst bound is not correct (D5); (3) a TOCTOU window exists between `Lstat` and the following open on the regular-file path and is undisclosed (D12); (4) one unverified premise is dressed as a reason: "Lstat and Remove are portable" (Windows symlink/junction semantics were not observed; only build and vet exist).

**D5 test seam.** A production-struct field added only for a test is the larger of two possible seams. Observed: a read-only handle makes the stamp write fail on darwin without any struct field (`os.Open(p).Truncate(0)` -> `invalid argument`; `WriteAt` -> `bad file descriptor`; scratch program, outside the tree). Extracting the locked phase of `pruneExclusive` into a function that takes the opened `*os.File` would let the test pass a read-only handle. The SPEC's option table omits this and states "no file-system fault available" as the reason for a field; that premise is incomplete (D6). The field is otherwise minimal and scoped (unexported, nil falls back), so the choice is defensible, just not shown to be the smallest.

## Acceptance criteria / two-cell discipline (task item 4)

Observed on this tree (the AC-HRH-001 verify command, run as written):

```
$ unset MOAI_KANBAN ... && go test -count=1 -run '^TestPruneStateSymlinkTargetUntouched$' ./internal/harness/
ok  	github.com/modu-ai/moai-adk/internal/harness	0.813s [no tests to run]
exit=0
```

Every "RED anchor" in the traceability table is either probe output (spec.md §A) or audit-quoted text; none is the AC's own command with its stdout, exit code and tree SHA. `acceptance.md` pins `1e2151a38` only in AC-007 and AC-011. The plan's M0 would create the cells, but M0's recording instruction (plan.md:38 "record verbatim command and output") omits exit code and tree SHA and no green-path cell names a flipping milestone per AC (D7).

Mutant probe results: AC-001 alone admits "skip the prune when the state path is a link" (caught by AC-002); AC-003 admits `chmod 0600` instead of replacement (acceptable, arguably better); AC-007 passes a mutant that adds an `os.Lstat` at the top of `PruneStaleEntries` (every hook gains a syscall without touching observer.go); AC-008/009/010 pass a comment containing only the sentinel strings; AC-012 requires applying a hand-made mutant overlay that is described, not committed.

Vacuous/impossible/wrong-reason red check: AC-014's RED as planned is wrong-reason (compile failure, D1). I confirmed the mechanism: a scratch test referencing a not-yet-existing field fails the whole package build (`r.writeStampFn undefined (type *Retention has no field or method writeStampFn)`, `FAIL ... [build failed]`).

## Scope and tier (task item 5)

12 REQ / 15 AC against the Tier M ceiling 16 / 16: within budget. Seven card items each have at least one REQ and AC (F4: REQ-001, AC-001/002; F7: REQ-002/003, AC-003/004; event loss: REQ-004/005/006, AC-005..008; F5: REQ-007, AC-009; F6: REQ-008, AC-010; N1: REQ-010, AC-012; N2: REQ-011/012, AC-013/014; REQ-009 spans F5/F6). Nothing is outside the card scope. Exclusions are consistent with the REQs (REQ-009 and the lockfile exclusion spec.md:141; REQ-005 and the observer exclusion spec.md:143-144). One residual: F7's core complaint ("silently disables retention") stays silent in the un-healable case (directory not writable); disclosed in §F (spec.md:156).

## Decision index (task item 7)

All six rows carry an authority anchor or an honest "(none ...)" (Q1-Q5 "(none — ...)", Q6 anchored to spec-workflow.md § SPEC Complexity Tier). Pull-mode neutrality is imperfect (D11): Q3 describes appender locking with consequences ("can get an append killed ... before it writes") and tail-carry with benefits ("no change to the append path"); Q4 carries a mitigating parenthetical ("a 35-byte derived stamp; for a symbolic link only the link is removed"); Q5 uses "leaves N2 half-closed" for the alternative. Plan B4 and progress.md name only D1-C, D2-B/C, D3-A as liftable options, although Q4 and Q5 are also open; an operator choice of D4-B would require rewriting REQ-001..003 with no amendment path stated.

## Defects Found

D1. M0-COMPILE-ORDER — plan.md:38 / acceptance.md:119-120,144 / spec.md:99-106 — Plan M0 step 1 puts "mutant-killing tests for AC-HRH-013, -014 (test files only; production code untouched)" into the baseline commit and DoD #1 requires that commit to precede every commit touching `retention.go`; AC-HRH-014 requires "a pruner whose stamp writer is replaced", which exists only after the M3 production change (plan.md:65 adds the field to `Retention`). Quoted pair: plan.md:38 "test files only; production code untouched" versus plan.md:65 "add an unexported function-valued field on `Retention`" and acceptance.md:144 "committed in a commit that precedes every commit touching `internal/harness/retention.go`". Observed consequence: a test referencing the absent field makes the whole package fail to build, so every other M0 RED (AC-001/002/003/005/013) would be a build failure instead of the observed behavioural RED (wrong-reason red), and plan.md:30 ("RED observed in M0 against the unmodified retention code") is unsatisfiable for AC-014. — Severity: critical — Class: blocking — Required fix: take the AC-014 test out of the M0 baseline commit; M0 records only compilable tests (AC-001/002/003/005/013) plus the surviving-mutant observation for mC against the existing suite; the AC-014 test lands in M3 together with the seam, with the mC survivor as its RED cell. Restate plan.md:38/:40 and acceptance.md:144 accordingly. (MP-9.)

D2. REPORT-PATH-IGNORED — plan.md:38,42,44 / acceptance.md:144 / spec.md:125 — The baseline-first witness (M0 Exit, DoD #1, §D) is a commit containing `.moai/reports/t1432/red-baseline.md`, but `.gitignore:235` is `.moai/reports/*` ("Card/audit reports are local-only artifacts ... never on the remote"; the verdict-file exception was withdrawn). `git check-ignore -v .moai/reports/t1432/red-baseline.md` -> `.gitignore:235:.moai/reports/*`. The commit graph cannot witness the ordering of an ignored file, and `git add -f` contradicts the operator directive in that file. verification-claim-integrity §2.3 requires the graph to be the witness. — Severity: critical — Class: blocking — Required fix: make the graph witness a tracked artifact: a tests-only commit (the failing tests, their observed RED output in the commit message and in `progress.md` §E.2) that is an ancestor of the first commit touching `retention.go` (witness: `git merge-base --is-ancestor <tests-commit> <fix-commit>` exit 0). Keep `red-baseline.md` as a local companion if wanted, and rewrite M0 Exit and DoD #1 to name the tracked witness.

D3. TAIL-NEWLINE — spec.md:113 (REQ-HRH-004), :153 / acceptance.md:64-66 (AC-HRH-006 b) — The boundary rule carries a tail "verbatim ... after the kept lines" and AC-006(b) pins "the unterminated line appears exactly once, whole, in the replacement". A rewritten log that ends without a newline makes the next observer append (`json.Marshal` + `'\n'`, O_APPEND, observer.go:70-87) land on the same line and corrupt the NEXT valid event. The baseline never leaves that state (observed: probe `PROBE-UNTERM ends_with_newline=true`). The AC does not pin a terminator, so a verbatim copy satisfies it. — Severity: major — Class: blocking (REQ and AC encode a regression against measured baseline behaviour) — Required fix: add to REQ-HRH-004 and AC-HRH-006(b) that the replacement ends with a newline (terminate an unterminated carried tail), plus an AC that a following append lands on its own line; state in §B D3 that an unterminated final line is no longer classified (a stale unterminated event is archived one interval later) and that malformed-JSON tail lines keep "carried verbatim" except for the terminator.

D4. D8-SYSCALL — spec.md:47,85,95 / acceptance.md:3 — D8 literal result: three sections reference `syscall` with no `//go:build` or exemption clause (verb run by hand, awk refused). Substantive gap: new FIFO tests need `syscall.Mkfifo` and the SPEC only states a runtime skip. — Severity: critical (MP-6 BLOCKING) — Class: blocking — Required fix: state in acceptance.md (and plan M0/M4) that every new or edited FIFO test file carries `//go:build !windows` like `retention_killed_test.go:1`, and either rephrase the three descriptive mentions ("system call", "O_NOFOLLOW is undefined on Windows") or add an explicit cross-platform exemption clause in each affected section.

D5. HEAL-BURST-OVERCLAIM — spec.md:94,155 / acceptance.md:132 / plan.md:57 — "one possible duplicate prune when two processes heal the same fault at the same instant" and "at most one duplicate prune (spec §F); the stamp then bounds repeats" are unverified (§F admits "Not measured") and the mechanism contradicts them: `os.Remove(path)` is not inode-conditional, so a late process deletes the state file another process just created, creates its own, and each locks a different inode (the stamp written into an unlinked inode is lost). The burst is N hook processes, not two, and concurrent pruners re-open the t1425 hazard (concurrent appends to one `<YYYY-MM>.jsonl.gz`, the archive corruption t1425 fixed). The retry bound in plan.md:57 ("exclusive create reports the path already exists") does not cover the delete-the-winner case. — Severity: major — Class: blocking (an unobserved premise stated as a bound) — Required fix: restate the bound as "up to the number of hooks that observed the fault in the same instant, once per fault healed"; require removal only when `os.SameFile` of the re-`Lstat` result matches the inode inspected first, add an AC for two concurrent healers, and keep the "not measured" label.

D6. SEAM-OPTIONS-INCOMPLETE — spec.md:99-106 / acceptance.md:115-120 — D5 states no portable file-system fault exists and offers only a struct field or a recorded survivor. A read-only `*os.File` makes `writeStamp` fail on darwin (observed above; Windows behaviour unobserved), enabling a parameter seam (extract the locked phase) with no struct field. AC-014 hard-codes the field mechanism ("stamp writer replaced"). — Severity: minor — Class: optional — Required fix: add the extraction option to D5 with its cost, and word AC-014 on the outcome (a stamp-write failure returns an error and leaves the log unmodified) rather than on the mechanism.

D7. RED-NOW-CELLS — acceptance.md:7-23, plan.md:38 — No AC carries the four §2.1 elements; the cited commands print `ok [no tests to run]`, exit 0 today (observed); M0's recording instruction omits exit code and tree SHA; no green-path cell names a flipping milestone per AC; the table's label "AUDIT-REPORTED, re-observed in M0" (AC-012/013/014 rows) asserts a re-observation that has not happened. — Severity: major — Class: blocking (§2 adoption is [HARD]) — Required fix: classify each behaviour-changing AC explicitly (release-blocking vs regression-guard); add to M0 the mandatory per-AC record `command / verbatim stdout / exit code / tree SHA` and a "flips at Mx" column; reword the "re-observed in M0" labels as "to be observed in M0".

D8. WEAK-DOC-ACS — acceptance.md:76-93 — AC-008/009/010 verify with `grep -c -F` sentinels; AC-010 greps 2 of the 5 obligations in REQ-HRH-008 (nothing for "event appended before the wait", "waiter delay not reproduced", the design-question clause); AC-009 does not check the "no cross-process exclusion" or "burst" statements. A comment holding just the sentinels passes. — Severity: minor — Class: optional — Required fix: pin each REQ-HRH-007/008 obligation with its own sentinel phrase or name the reviewer-read as the pass condition.

D9. HISTORY-IN-SOURCE — spec.md:117 (REQ-HRH-008) — requires the production comment to say "the earlier lock-behaviour design question did not carry the hook-timeout fact". The audit asked that "the record" say it (sync-audit.md:224), i.e. the decision record, not source. Card/process history in source comments is out of place. — Severity: minor — Class: optional — Required fix: move that clause to `progress.md` or the decision record; keep only the technical facts (blocking wait, 5 s timeout, event appended first) in REQ-HRH-008.

D10. FIGURE-MISQUOTE — spec.md:48 — "`pruneExclusive 85.0%`" is labelled AUDIT-REPORTED; the audit reports 87.5% (sync-audit-delta.md:100, :164). — Severity: minor — Class: blocking-lite for premise honesty, cheap fix — Required fix: correct to 87.5% (or drop the figure).

D11. DECISION-INDEX-FRAMING — decision-index.md:19-23, :28-30, :33-37, plus plan.md:18 and progress.md:10 — loaded wording in a pull-mode neutral question (Q3, Q4, Q5), and inconsistent liftable-option set (Q4/Q5 open, but only D1-C, D2-B/C, D3-A are named as liftable). — Severity: minor — Class: optional — Required fix: state each option's cost and benefit in parallel form; either list D4-B and D5-B as liftable in plan B4/progress or mark Q4/Q5 as non-blocking for Kickoff.

D12. TOCTOU-UNDISCLOSED — spec.md:110 (REQ-HRH-001), :151-158 — REQ-001 is absolute ("shall neither open, truncate nor write the link's target") while the specified `Lstat` then open sequence leaves a window on the regular-file branch (exclusive create does not follow links; the plain read-write open does). §F lists no such residual. — Severity: minor — Class: optional — Required fix: add the residual to §F and, if cheap, open the regular-file branch with a post-open `os.SameFile` check against the `Lstat` result.

D13. FASTPATH-UNPINNED — acceptance.md:69-74 — AC-007 pins only `observer.go` bytes; the observer calls `PruneStaleEntries` inline on every event (observer.go:93), so a syscall added to the lock-free fast path in `retention.go` violates REQ-005's spirit and the §D constraint ("No requirement here may add work to the append path") without failing AC-007. — Severity: minor — Class: optional — Required fix: add an AC that the fresh-stamp path (`TestPruneStamp_FreshStampNeedsNoLock`) stays unmodified and green, and place the heal helper only inside `pruneExclusive`.

D14. UNCHECKED-CI-CLAIM — acceptance.md:99 — "the CI release matrix compiles the Windows target" was not verified in this run. — Severity: minor — Class: optional — Required fix: cite the workflow file and job, or drop the sentence.

## Regression Check

Iteration 1: not applicable.

## Evidence-bearing report (five sections)

### Claim
The SPEC is not adoptable at this iteration: the baseline-first mechanism (M0/DoD #1) is unsatisfiable on two independent grounds (D1 compile ordering, D2 ignored path), a measured-baseline regression is encoded in the tail rule (D3), D8 returns BLOCKING (D4), and the heal-burst bound is an unobserved premise (D5). Its premise labels are truthful (three OBSERVED premises re-observed; F5/F6 honestly not reproduced). Score 0.75 < 0.80.

### Evidence
Commands and verbatim outputs are in the sections above: tree identity table; `moai spec lint` -> `✓ No findings`; overlay probe output (F4, F7, loss, unterminated-tail baseline); the vacuous `[no tests to run]` run; the build-failure line for a test referencing a missing field; `git check-ignore -v` -> `.gitignore:235:.moai/reports/*`; read-only handle probe; `GOOS=windows go doc syscall.O_NOFOLLOW`; `GOOS=windows go vet` (no output).

### Baseline-attribution
All measurements were taken in this run against the tree at HEAD `69ddcc73ae1b6e776d2cc170c3814160788dd1ef` (retention.go and lockfile files unchanged from base `1e2151a38`, `git status --short` empty before and after). Tool-provenance: `moai spec lint` was judged by the installed build `v3.2.0-rc.26 ... g45600e4ee`, a strict ancestor of HEAD (127 commits behind); `go` commands built the tree directly. The mutant/probe results are scratch-overlay runs; nothing was written into the tree.

### Gaps
- Refused tool calls (verification-claim-integrity §3.1): the AC-4/AC-5 traceability verb, the CN-4 ordering verb and the D8 verb are `awk` programs; the worktree guard refused the first ("this command runs awk with a program that can execute commands or that it cannot read in a plain command"). One compound probe command was also refused as "too complex to verify", and a `git check-ignore ... ; echo` compound was refused; each was re-issued as plain separate commands. Substitutions: traceability by `grep` on REQ and AC ids, D8 by `grep -n 'syscall'` and `grep -n '//go:build'`, CN-4 by reading plan.md and acceptance.md by hand. These are inferred-by-reading, not the verbs' own output; no `COLLECTED/CONFLICT` lines exist for CN-4.
- Not run: N1 mutant mD hang, N2 mutants mB/mC (taken from the t1425 audit text), any Windows runtime, `go test -race ./internal/harness/` on the unmodified tree, the CI matrix claim (D14), concurrent heal behaviour (D5 reasoning from the code, not observed), the `Lstat`/`Remove` Windows semantics.
- Not compared to the SPEC's own scratch probe (it lives outside the tree); my probe is an independent reproduction.

### Residual-risk
Even after D1-D5 are fixed: the tail-carry window (final size reading to rename, and old-inode writers) is real and only disclosed; the Lstat-then-open TOCTOU; un-healable F7 (unwritable directory) stays silent; F5/F6 stay unreproduced hypotheses, so the disclosure may be wrong in either direction; MP-8 stays N/A until ACs are classified.

## Recommendation

FAIL. Revise through manager-spec, then re-audit scoped to D1-D5 and D7 (plus the full CN-4 ordering check):
1. D1: split M0 so only compilable tests are in the baseline; move the AC-014 test and the seam together into M3 (plan.md:38,40,65; acceptance.md:144).
2. D2: replace the ignored `red-baseline.md` witness with a tracked tests-only ancestor commit (plan.md:44; acceptance.md:144; spec.md:125).
3. D3: require a terminated tail in REQ-HRH-004 / AC-HRH-006(b); add the following-append AC.
4. D4: add the `//go:build !windows` obligation for FIFO test files and clear the three `syscall` sections.
5. D5: correct the heal-burst bound; require inode-conditional removal and an AC for concurrent healers.
6. D7: classify the ACs; define the per-AC RED-now record (command, stdout, exit code, tree SHA) and a "flips at" cell.
7. Fix D10 (87.5%) in the same pass; D6, D8, D9, D11-D14 are optional per M6 and do not by themselves justify a FAIL.

## Operational Notes (unverified)

- Measure whether a dedicated test-only tracked path exists for the baseline evidence — command: `git -C <tree> ls-files .moai/specs/SPEC-HARNESS-RETENTION-HARDEN-001/progress.md` (status: assumption; progress.md is tracked at HEAD per the plan commit, not re-checked here).
- Measure the N-process heal burst before accepting any bound — command: spawn N helper processes on one 0400 state file with the existing helper-process pattern in `retention_concurrent_test.go:50` and count archive appends (status: inferred from the Remove-not-inode-conditional reading; not measured).
- Measure whether `GOOS=windows go vet ./internal/harness/` fails when a new FIFO test lacks the build tag — command: add the file via `go vet -overlay` outside the tree (status: inferred from `syscall.Mkfifo` being undefined on Windows; not run).

AUDIT-VERDICT: FAIL spec=SPEC-HARNESS-RETENTION-HARDEN-001 receipts=none
