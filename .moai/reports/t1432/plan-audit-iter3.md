auditor-model: claude-sonnet-5-5[1m]

verdict: PASS
audited_sha: c2cc320805f5bd7a96df38e8e7cf1b44a361fbec

# SPEC Review Report: SPEC-HARNESS-RETENTION-HARDEN-001

Iteration: 3/3 (final)
Verdict: PASS (no must-pass failure; no blocking finding; seven non-blocking findings, each classified accept-as-debt or must-fix below)
Overall Score: 0.92 (Tier M threshold 0.80)
Plan Artifact Hash: not computed (the run-gate hash is a runtime record; this audit read the committed files at the SHA above)
Auditor Version: plan-auditor/v-unversioned (model line above)

Reasoning context ignored per M1 Context Isolation. The two earlier reports were read only after the independent pass over the committed artifacts.

## Claim

1. Iteration-2 blocking findings B1, B2 and B3 are fixed (details under Regression Check).
2. Every release-blocking criterion whose RED-now cell was re-run reproduces on the current tree with the ledger's failing lines and the same exit code; the swept counts are non-vacuous.
3. The seven in-scope items are covered and nothing outside the binding scope is added; the five relayed decisions plus the three accepted process points are honoured and recorded as relayed, not as operator answers.
4. The revision introduces no new blocking defect. It does carry stale or loose statements (N1-N7 below), none of which makes a criterion impossible, vacuous or untraceable.

## Tree and tool attribution

- Toplevel `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1432`, branch `WT-harness-retention-debt`, HEAD `c2cc320805f5bd7a96df38e8e7cf1b44a361fbec` (`c2cc32080`), `git status --short` empty before and after (all probes ran through `go test -overlay` against committed sources; the only file written in the tree is this report; scratch files went to the session scratch directory).
- Code under audit equals base: `git diff --stat 1e2151a38 HEAD -- internal cmd` printed nothing; `git diff --stat 1e2151a38 HEAD -- internal/lockfile internal/harness/observer.go` printed nothing.
- Judging build: `moai version` printed `moai-adk v3.2.0-rc.26 ... archive/t1401-293-g45600e4ee built 2026-10-02T14:42:32Z`. `git merge-base --is-ancestor 45600e4ee HEAD` exits 0 (strict ancestor: the build predates HEAD); `git diff --stat 45600e4ee HEAD -- internal/spec` printed nothing, so the lint code that build carries equals HEAD's and the lag does not affect the `moai spec lint` result. No MCP audit tool was called (no `audit_model` key was observed in iteration 2 either), so `receipts=none`.
- Platform of every run: darwin, uid 501, go toolchain `go1.26.8` (path seen in the mD stack trace).

## Must-Pass Results

- [PASS] MP-1 REQ numbers: `grep -n -o -E "^- \*\*REQ-HRH-[0-9]+\*\*" spec.md` printed REQ-HRH-001..015 on spec.md lines 157-171, sequential, no gap, no duplicate; AC-HRH-001..014 on acceptance.md lines 39-145.
- [PASS] MP-2 GEARS, judged on the REQ layer in `spec.md` only (the Given-When-Then entries in `acceptance.md` are the verification layer and were not graded here): REQ-001..006 and 013 are event-driven `When ..., the ... shall`; REQ-007..012, 014, 015 are ubiquitous `The ... shall`. REQ-012 packs four checkable clauses into one sentence (style note only). REQ-008 names four function identifiers in the requirement text (RQ-4 note, optional).
- [PASS] MP-3 frontmatter: spec.md:2-16 carries id, title, version "0.3.0", status draft, created, updated, author, priority P2, phase "v3.2.0 target", module, lifecycle, tags (string), plus `tier: M`, `card`, `related_specs`. `moai spec lint SPEC-HARNESS-RETENTION-HARDEN-001` printed `✓ No findings — all SPEC documents are valid`. `grep -n "^status:" plan.md acceptance.md` found none (the `---` hits are test-output lines in the ledger).
- [N/A] MP-4: single-language (Go) internal package.
- [PASS] MP-5 D7: the referenced SPEC ids are itself and `SPEC-AGENT-TEAM-RETIRE-001`, whose `status:` is `completed` (spec.md:5 of that SPEC, v0.1.2); not retired, superseded or archived, so no REVIEW line.
- [PASS] MP-6 D8: `grep -c syscall` printed spec.md 0, plan.md 0, acceptance.md 3 (the three ledger hits, in the `## Evidence ledger` section whose preamble states the `//go:build !windows` and `//go:build windows` twin obligation). The per-section awk verb was not run (the worktree guard refuses awk programs; see Gaps); D8-4 applies to the spec body, which has no hit.
- [PASS] MP-7: `grep -rn "NEEDS CLARIFICATION"` over the SPEC directory printed nothing; `research.md` does not exist (Tier M).
- [PASS] MP-8 RED-now re-execution: see the table under Evidence. Every release-blocking criterion's RED reproduces on the current tree.
- [PASS] MP-9 ordering: the CN-4 verb (run from the iteration-2 auditor's script copy, see Gaps) printed `COLLECTED: 7 milestones in plan order (M0 M1 M2 M3 M4 M5 M6), 0 exit bindings, 36 ordering candidates` and no `CONFLICT:` line. I read all 36 candidates: none binds a criterion to a milestone against the plan order. The one ordering obligation that matters, DoD 1 (acceptance.md:517) against plan M0 Exit (plan.md:54), is jointly satisfiable and now carries positive controls (B1 re-verification below). One drafting looseness in DoD 1 is recorded as N1, not as a conflict.

## Category Scores (rubric-anchored)

| Dimension | Score | Band | Evidence |
|---|---|---|---|
| Clarity | 0.90 | 0.75-1.0 | REQ-001/003/004 state the link-owner rule in the requirement text (spec.md:157, :159, :160); REQ-002 is narrowed to permission errors (spec.md:158); FIFO claims removed from REQ-003 and D4.a (spec.md:159, :115). Looseness: DoD 1 "every other `.moai/reports/t1432/*` file ... in commit T" (N1); two stale statements (N2, N3). |
| Completeness | 0.95 | 1.0 | HISTORY (spec.md:21-27), Background (§A), Decisions (§B), Requirements (§C), Constraints (§D), `### Out of Scope — ...` H3s with `-` bullets (spec.md:183, :190, :199, :204), Risks (§F). 15 REQ / 14 AC inside the Tier M ceilings 16 / 16; five new test files plus one edited, as plan.md §A says. |
| Testability | 0.88 | 0.75-1.0 | RB cells reproduce; B1 ordering check has positive controls that fire; AC-HRH-005 now kills the Stat-based and every-link-owned mutants at the check level. Remaining: O4, O8, O9 left open by design; AC-005 (e) may be a recorded skip on Linux; plan-phase mutant copies are base-bound (N6). |
| Traceability | 0.95 | 1.0 | The AC-4/AC-5 verb printed `COLLECTED: 15 REQ definitions (acceptance input: read)` and no `UNCOVERED:` or `ORPHAN:` line. Traceability table (acceptance.md:20-35) maps each REQ to at least one AC; REQ-001/003/004 also reach AC-005. The 'another tool's silence' rule is not used: the lint tool was not cited as corroboration of traceability. |

Overall 0.92 (mean of the four, rounded). Iteration trend: 0.75 (iteration 1), 0.87 (iteration 2), 0.92 (iteration 3): no regression, no STOP.

## Evidence (commands run, observed output)

All commands plain single invocations from the worktree root unless stated. Exit codes: `go` commands surface a non-zero exit as `Exit code N`; the Bash tool does NOT surface exit 1 for `grep` or `git diff` (observed: `git diff --exit-code --name-only HEAD~1 HEAD -- spec.md` printed the file name with no exit marker), so exit codes of those entries are read from their output, not observed (Gap G-A).

### B1 re-verification (ordering check)

| Command | Observed |
|---|---|
| `git log --reverse --format=%h 1e2151a38 -- internal/harness/retention.go` | `68f023289 fe0901cdc fe211e9c9 987dcc61b 8172295ca 035a7bc31 fe4e8f44e` (E-021: seven commits, all older than the base) |
| `git log --reverse --format=%h 1e2151a38..2ebc10f8f -- internal/harness/retention.go` | empty (E-022: bounded listing, no production commit yet; the SPEC reads this as a Gap, never a pass) |
| `git log --reverse --format=%h fe211e9c9~1..fe211e9c9 -- internal/harness/retention.go` | `fe211e9c9` (E-023: the bounded form lists a hash on a range with a change) |
| `git merge-base --is-ancestor 2ebc10f8f 68f023289` | `Exit code 1` (E-024: the check CAN fail on a mis-ordered pair) |
| `git merge-base --is-ancestor 68f023289 2ebc10f8f` | exit 0, no output (E-024: it passes on the right order) |

Conclusion: the unbounded form cannot pass (T is a descendant of the base), the bounded `1e2151a38..HEAD` form can, and both directions of the test are observed. FIXED. Residual (optional): the Exit names `retention_statepath_test.go` as the witness file for T; the tail and stamp-bytes tests are required to be in the same commit by plan.md M0 step 6 but the Exit line does not mechanically check that (accept as debt; the test files are listed together in M0).

### B2 re-verification (owner pin)

- REQ wording: spec.md:157 (REQ-001) "the owner recorded on the link itself, read without following it, never the owner of the entry it points at"; spec.md:159 (REQ-003) "for a symbolic link: because the link's own owner is not the current user, whatever its target's owner"; spec.md:160 (REQ-004) same parenthesis.
- AC-HRH-005 (acceptance.md:73-79) has (d) a user-created link to `/` expected "owned" and (e) a foreign-owned link expected "not owned", and names three mutants: always-owned (fails (b), (e)); follow-the-link read (fails (d)); every-link-owned (fails (e), passes (d), which is why (e) exists). DoD 6 (acceptance.md:522) and plan M2 step 6 (plan.md:78) list all three.
- Premises re-observed this run: `stat -f 'link-own: %Sp uid=%u'` on a user-created link to `/` printed `link-own: lrwxr-xr-x uid=501`; `stat -L -f 'target-own: ...'` on it printed `target-own: drwxr-xr-x uid=0`; `id -u` printed `501` (E-026 reproduces). `stat -f '%Sp %u %Su %N' /var /tmp /etc /bin` printed `lrwxr-xr-x 0 root` for `/var`, `/tmp`, `/etc` and `drwxr-xr-x 0 root /bin` (E-027 reproduces; `/bin` is a directory on darwin and is skipped by the "first candidate that is a foreign-owned link" rule).
- Exactly one test-only field: plan.md:76 (M2 step 4) names one function-valued unexported field on `Retention`, installed by `NewRetention`; AC-006's helper split (plan.md:75), AC-014's parameter seam (spec.md:147) and AC-007's FIFO use no field. spec.md:178 and acceptance.md DoD 7 state the allowance was spent once.
- Conclusion: a Stat-based owner read and an every-link-is-owned check are both killed at the check level by AC-005 (d) and (e). FIXED. Residual: no criterion drives the pruner end to end with a user-owned link to a root-owned target, so a pruner that resolved the link before calling the check would pass AC-001..006 (N5, accept as debt).

### B3 re-verification (FIFO hang)

- `go test -count=1 -v -overlay .moai/reports/t1432/red-now-drafts/overlay-b3.json -run '^TestZZProbeB3FifoAtStatePath$' ./internal/harness/` printed `zz_fifo_state_test.go:29: PROBE-B3 BLOCKED: PruneStaleEntries did not return within 3s on a FIFO state path` and `--- PASS: TestZZProbeB3FifoAtStatePath (3.00s)` (E-025 reproduces: the hang is real on the base tree).
- `grep -n -i -E "fifo|special file|named pipe"` over the five files: every state-path FIFO mention is a statement that it is pre-existing, unrepaired, outside this card (spec.md:50, :65, :115, :157, :159, :199-202, :218; acceptance.md:62, :157; plan.md:25, :73, :115; progress.md:12). No REQ, AC or plan step promises a skip, an error or a pin for a FIFO at the state path. The FIFO uses that remain (acceptance.md:91 archive path as FIFO for AC-007; retention_killed_test.go for N1) are archive-path and test-only FIFOs, not the state path.
- AC-HRH-003 stays decidable: both cases are non-FIFO (a directory; a regular file in a read-only directory, acceptance.md:62). `TestPruneStateUnreplaceableInReadOnlyDirSkips` re-ran: `--- PASS` (E-012 reproduces).
- Conclusion: FIXED; the hang is recorded and routed to the leader (plan.md B9, progress.md:12).

### Iteration-2 optional findings O1-O10

| # | Status | Evidence |
|---|---|---|
| O1 | fixed | AC-HRH-009 now scopes four functions with `-L` (acceptance.md:110). Re-ran: four-function form on `1e2151a38..HEAD` printed nothing (E-028); on `fe211e9c9~1..fe211e9c9` printed `fe211e9c9` (E-029); out-of-order options printed `fatal: -L parameter '^func readStamp(' starting at line 195: regexec() failed to match` with `Exit code 128` (E-030); the `PruneStaleEntries`-only controls printed `fe211e9c9` then nothing (E-015). |
| O2 | partly fixed | Placeholders are gone; every ledger command is a single invocation against committed overlay JSON. The JSON files still hold absolute paths into the plan session's scratch directory (acceptance.md:174, G-3); the commands ran here only because that directory exists. Left to the leader or M0 with the precondition stated (N7). |
| O3 | fixed | REQ-002 reads "cannot open ... because of a permission error" (spec.md:158); REQ-003 sends other open failures to skip (spec.md:159). |
| O4 | left open, reason stated | progress.md:18: inode-reuse false red, APFS observed not to reuse, Linux unobserved. Accept as debt (failure direction is a false red, not a missed defect). |
| O5 | fixed | spec.md:106 states a final unterminated line is classified by the first later prune that finds it terminated, and names the no-rewrite case; acceptance.md:161 repeats it. |
| O6 | fixed | DoD 6 lists the AC-003 mutants (acceptance.md:522). |
| O7 | fixed | acceptance.md:102 "four top-level `--- PASS` lines (... also prints one indented `--- PASS` line per subtest)". |
| O8 | left open, reason stated | progress.md:19: late line compared by substring, so a re-encoding mutant passes. Accept as debt. |
| O9 | left open, reason stated | progress.md:20: the phrase naming the missing archive step is judged by reading. Accept as debt. |
| O10 | fixed | plan.md:75 names the inspection step / removal step split for AC-006. |

Iteration-1 findings D1-D14: none regressed (D1 M0 compile order: plan.md:47 keeps the symbol-dependent tests out of M0; D2 tracked witness: the three test files are the witness; D3 terminator: REQ-007; D4 build tags: REQ-012; D5 heal burst: REQ-005 and D4.c; D7 RED-now cells: the ledger; D10 figure: spec.md:52 states `pruneExclusive 85.0%` is the first audit's and `87.5%` the delta audit's).

### MP-8 RED-now re-execution (single invocations, current tree; Go code equals base and `2ebc10f8f`)

Re-executed by me (observed output, not read):

| AC (class) | Ledger | Command (as cited) | Observed on this tree | Matches ledger |
|---|---|---|---|---|
| AC-001 (RB) | E-001 | `go test -count=1 -overlay .../overlay-drafts.json -run '^TestPruneStateSymlinkReplacedTargetUntouched$' ./internal/harness/` | `--- FAIL ...` `victim changed: err=<nil> content="2026-10-02T00:00:00Z"`, `state path is not a regular file: err=<nil>`, `FAIL`, `Exit code 1` | yes |
| AC-002 (RB) | E-002 | same form, `TestPruneStateUnwritableFileReplaced` | five failure lines incl. `first prune returned retention: prune state open failed: open ...: permission denied, want nil`, `Exit code 1` | yes (temp dir name differs) |
| AC-007 (RB) | E-003 | `TestPruneCarriesLateEvents` | `late-event count = 0, want 1`, `order wrong: [...fresh...]`, `Exit code 1` | yes |
| AC-008 b (RB) | E-004 | `TestPruneTailPartialLineCarriedAndTerminated` | `partial fragment occurs 0 times, want 1`, `Exit code 1` | yes |
| AC-008 c (RB) | E-005 | `TestPruneStaleFinalLineWithoutNewlineWaitsOneInterval` | `stale-final count after first prune = 0, want 1 ...`, `Exit code 1` | yes |
| AC-013 survivor (RB) | E-006 | `go test -count=1 -overlay .../overlay-mB.json ./internal/harness/` | `ok  github.com/modu-ai/moai-adk/internal/harness  0.955s` (no `[no tests to run]`; whole package swept) | yes |
| AC-013 kill (RB) | E-007 | `... overlay-mB-newtest.json -run '^TestPruneStampShorterOverLongerIsExact$'` | `state file = "2026-10-02T00:00:00Z123456789Z", want exactly "2026-10-02T00:00:00Z"`, `Exit code 1` | yes |
| AC-013 unmutated | E-020 | `... overlay-drafts.json -run '^TestPruneStampShorterOverLongerIsExact$'` | `ok ... 0.415s` with no `[no tests to run]` (the draft is in the overlay, so it ran) | yes |
| AC-014 survivor (RB) | E-008 | `... overlay-mC.json ./internal/harness/` | `ok ... 0.923s` | yes |
| AC-012 (RB) | E-009 | `go test -count=1 -timeout 40s -overlay .../overlay-mD.json -run '^TestPruneStamp_StampExistsBeforeTheWork$' ./internal/harness/` (background) | `panic: test timed out after 40s`, `running tests: TestPruneStamp_StampExistsBeforeTheWork (40s)`, goroutine in `syscall.Open` at `retention_killed_test.go:43`, `FAIL ... 40.583s`, `[exited with code 1]` | yes |
| AC-010 (RB) | E-010 | `grep -c -F -e "residual window" -e ... -e "F6: not reproduced, not measured" internal/harness/retention.go` | `0` | yes (exit 1 inferred, G-A) |
| AC-010 (RB) | E-011 | `grep -c -F "events other hooks append in that window are lost" internal/harness/retention.go` | `1` | yes |
| AC-010 guard | n/a | `grep -c -E 't1[0-9]{3}' internal/harness/retention.go` | `0` | consistent with acceptance.md:119 |
| AC-003 b (RG) | E-012 | `... overlay-unreplaceable.json -run '^TestPruneStateUnreplaceableInReadOnlyDirSkips$' -v` | `--- PASS: TestPruneStateUnreplaceableInReadOnlyDirSkips (0.00s)`, `ok` | yes |
| AC-009 (RG) | E-013 | `git diff --stat 1e2151a38 HEAD -- internal/lockfile internal/harness/observer.go` (stat form, because the exit of `--quiet` is not surfaced) | empty | yes |
| AC-009 control | E-014 | `git diff --exit-code --stat 6dc31a727~1 6dc31a727 -- internal/harness/observer.go` | ` internal/harness/observer.go | 4 +++-` (a real diff, so the `--quiet` form exits 1) | yes (exit inferred, G-A) |
| AC-011 (RG) | E-016 | `GOOS=windows go vet ./internal/harness/ ./internal/lockfile/` | empty, exit 0 | yes |
| AC-011 control | E-017 | `GOOS=windows go vet -overlay .../overlay-untagged.json ./internal/harness/` | `zz_untagged_fifo_test.go:11:20: undefined: syscall.Mkfifo`, `Exit code 1` | yes |
| AC-011 control | E-018 | `GOOS=windows go vet -overlay .../overlay-drafts.json ./internal/harness/` | empty, exit 0 | yes |
| vacuous form | E-019 | `go test -count=1 -run '^TestPruneStateSymlinkReplacedTargetUntouched$' ./internal/harness/` | `ok  github.com/modu-ai/moai-adk/internal/harness  0.557s [no tests to run]` (a vacuous green, correctly excluded by the swept-count rule) | yes |
| B1 | E-021..E-024 | see B1 table | match | yes |
| B3 probe | E-025 | see B3 | match | yes |
| AC-005 premises | E-026, E-027 | see B2 | match | yes |
| AC-009 | E-028..E-030 | see O1 | match | yes |
| mutant diffs | E-006/E-008/E-009 notes | `diff internal/harness/retention.go .../retention_mB.go` printed `210,212d209` (the `Truncate(0)` block); `..._mC.go` printed `142,144c142 ... _ = writeStamp(sf, now)`; `..._mD.go` printed `146c146 ... err = nil` | the notes are accurate |

Read, not re-executed: E-015 second command and E-031 (the latter re-observed through the `git diff --stat 1e2151a38 HEAD -- internal cmd` empty output), the `release-pr-multi-os.yml` line citations, and every Linux or Windows statement (no such runtime here). Not re-executed because it is a RED that cannot exist before the run phase: AC-004, -005 (adoption proofs), -006, and AC-014's kill (Gap G-1, G-2 in the SPEC, accepted as RG or survivor-RED by the ledger's own classes).

Pins: ledger entries pin `2ebc10f8f` or `db6d88a2a` (commit SHAs, not branch names); `git cat-file -t` printed `commit` for both; both carry Go code identical to base and to HEAD (`git diff --stat 1e2151a38 HEAD -- internal cmd` empty).

Swept-count check: every `go test -run` RED above printed a `--- FAIL` or `--- PASS` line for a named test, or an `ok` for a whole-package run with no `[no tests to run]`. No cited RED is a vacuous `ok [no tests to run]`.

## Defects Found

No blocking finding. Non-blocking findings, in order of weight. Class for each: optional (M6), with a disposition for the final iteration.

N1. DOD1-IN-COMMIT-T — acceptance.md:517 — "`.moai/reports/t1432/red-baseline.md` and every other `.moai/reports/t1432/*` file are committed with `git add -f` ... in commit T": read literally this puts files that do not exist at T (the completion `verdict.md`, plan.md §A line 9 names it as card evidence, written at M6) and files already committed earlier (the plan-audit reports and the drafts, tracked since `db6d88a2a`..`c2cc32080`) into T. Not an MP-9 conflict because the sensible reading (only `red-baseline.md` is added in T; later evidence is force-added in the commit that produces it) is unambiguous to a reader of plan.md M0 step 6, and the Exit witness does not depend on it. — Severity: minor — Class: optional — Disposition: accept-as-debt; if the leader wants it closed, reword to "`red-baseline.md` is committed with `git add -f` in commit T; every later `.moai/reports/t1432/*` file is force-added in the commit that produces it".

N2. STALE-UNTRACKED-CLAIM — acceptance.md:512 (G-7) and plan.md:21 (B5) — both say the B3 probe files are "untracked at this revision; the leader force-adds them". `git ls-files .moai/reports/t1432` lists `overlay-b3.json`, `probe-b3.out` and `zz_b3_fifo_state_test.go`, and `git show --stat c2cc32080` shows all three added in the revision's own commit. A statement about repository state that the commit it sits in contradicts. — Severity: minor — Class: optional — Disposition: accept-as-debt (no criterion depends on it; the leader's force-add action is already done).

N3. CI-WINDOWS-CLAIM-FALSE — acceptance.md:127 (AC-HRH-011) and plan.md:23 (B7) — "its Windows leg is a cross-compile of the binary (`go build`, test files not compiled, `:509`, `:545`)" and "the local Windows vet is the only per-card Windows guard". `.github/workflows/ci.yml:524-536` runs `go vet ./...` with `GOOS=windows`/`GOARCH=amd64` in the same `build` job on every PR, and its comment says it exists because `go build` does not compile `_test.go` files. The SPEC's READ-labelled statement is therefore false, and its "only guard" conclusion overstated. Effect: benign (the local guard is still correct and cheaper, and CI is a second guard); no criterion is affected. Iteration 2 marked the related iteration-1 D14 resolved on the strength of the citation without reading that step. — Severity: minor — Class: optional — Disposition: accept-as-debt; correct the sentence in the run-phase verdict or by an inline fix.

N4. AC005-LINUX-SKIP — acceptance.md:77, :79 — (e) is a recorded skip where no candidate among `/var`, `/tmp`, `/etc`, `/bin` is a foreign-owned link (darwin: `/var` qualifies, observed; Linux not observed; `/bin` is a root-owned link only on a merged-usr layout). If the skip is implemented as a top-level `t.Skip`, the test reports `--- SKIP` and steps (a)-(d) are not credited either; if the CI runner skips (e), the every-link-owned mutant (3) survives there. Stated as a limitation in the SPEC and covered locally by the M2 mutation run on darwin. — Severity: minor — Class: optional — Disposition: accept-as-debt; run phase should make each step its own subtest so a skipped step does not hide the others.

N5. WIRING-UNPINNED — acceptance.md:156 (Edge cases) vs AC-HRH-001..006 — the Edge cases section claims a user-owned link to a root-owned directory such as `/` is healed, but no criterion drives `PruneStaleEntries` with such a link; AC-005 (d) pins the check alone and AC-001 uses a user-owned target. A pruner that passed the resolved target path to the check would satisfy AC-001..006 and violate REQ-001's link-owner rule. A cheap strengthening exists (an AC-001 case whose link points at `/`, constructible by a non-root user, observed in E-026). — Severity: minor — Class: optional — Disposition: accept-as-debt; mention in the run-phase delegation as a recommended extra case.

N6. MUTANT-COPIES-BASE-BOUND — acceptance.md:134 (`<mD overlay>`), plan.md:50, :100 — the committed `retention_m*.go` files are copies of the BASE `retention.go` plus one edit. An overlay that replaces the whole file at M4 or M6 reverts M1-M3's production changes and breaks compilation of the new test files that name the new symbols, giving a build failure (a wrong-reason red) instead of a mutant run. AC-012's text defines mD by behaviour ("the pruner stamps and then returns without pruning"), so the criterion is sound; the run phase must regenerate each mutant against the then-current `retention.go`. — Severity: minor — Class: optional — Disposition: accept-as-debt; state it in the M4 delegation.

N7. OVERLAY-PATHS-ABSOLUTE — acceptance.md:174 (G-3), plan.md:21 (B5) — the committed overlay JSON files name the plan session's scratch directory; the ledger re-executes only where it exists (it did here). The SPEC states the precondition and assigns regeneration to the leader or M0. E-009's stdout is elided with `...` (the full output is tracked in `red-now-drafts/mD-out.txt`), which is less than the verbatim stdout §2.1 asks for. — Severity: minor — Class: optional — Disposition: must-do before M0 for the regeneration (one setup step, owned by the leader per G-3); the elision is accept-as-debt.

Also carried open from iteration 2 and judged optional: O4, O8, O9 (accept-as-debt, reasons at progress.md:18-20).

## Regression Check (iteration 3)

- B1 RESOLVED (positive controls observed in both directions).
- B2 RESOLVED at the check level (residual N5).
- B3 RESOLVED (premise corrected; hang recorded as pre-existing and unrepaired; AC-003 decidable).
- O1, O3, O5, O6, O7, O10 RESOLVED; O2 partly (N7); O4, O8, O9 left open with reasons (accepted).
- Iteration-1 D1-D14: no regression found; D14 re-opened as the narrower N3 (a cited line was read for its job name only).
- Score trend 0.75 -> 0.87 -> 0.92: no regression, no STOP.

## Scope and decision check (binding scope, leader-relayed 2026-10-03)

| Item | Where | Verdict |
|---|---|---|
| F4 symlink followed | REQ-001, AC-001, AC-005 (d) | covered |
| F5 Windows lock | REQ-010 disclosure only, REQ-012 lockfile byte-identical; D1 default A | covered, no code repair |
| F6 lock waiters | REQ-011 disclosure only; D2 default A | covered, no code repair |
| F7 unwritable state file | REQ-002, REQ-003, AC-002, AC-003 | covered |
| late-event loss | REQ-006, 007, 008, 009; AC-007, 008, 009, 010 | covered (tail-carry plus residual window) |
| N1 FIFO drain hang | REQ-013, AC-012 | covered |
| N2 mB and mC | REQ-014, 015; AC-013, 014 | covered |
| no change to `internal/lockfile` | REQ-012, AC-011, DoD 3, §E | honoured; recorded as relayed (decision-index.md:5, row Q1, Q2) |
| heal deletes only a current-user-owned entry; foreign never deleted, warn only, test-pinned | REQ-001..004, AC-004, AC-005, D4.a | honoured; recorded as relayed (Q4) |
| exactly one unexported test-only field | spec.md:178, D5, plan M2 step 4 | honoured (spent on the owner lookup) |
| intermediate red M0 commit, force-added card evidence, conservative Windows rule accepted; completion report says "Windows runtime not observed" and names the red SHA | spec.md:179 (§D), acceptance DoD 8, plan M6, decision-index.md:5 item 7 | honoured; stated as a leader-session acceptance, not an operator answer (HISTORY row 0.3.0, decision-index item 7) |
| no FIFO repair, no F5/F6 code repair | spec.md:199-202, plan.md:107, :115 | not contradicted anywhere (greps above) |

Nothing outside the seven items is added except what the relayed Q4 verdict itself requires (ownership check, warning, heal safety), and the pre-existing FIFO hang is recorded without a requirement.

## Recommendation

PASS. Proceed to Kickoff under the default autonomous form (verdict PASS; score 0.92 at or above the Tier M threshold 0.80; artifact set unchanged since this audit; no open blocker) with these carry-forwards, none blocking:

1. Leader or M0, before the first commit: regenerate the overlay JSON files in `red-now-drafts/` against the committed sources (N7, G-3); correct acceptance.md G-7 and plan.md B5 (N2) and the CI-Windows sentence in AC-HRH-011 and plan B7 (N3) if a one-line inline fix is wanted (accept-as-debt otherwise).
2. Run-phase delegation should carry: mutants regenerated against the then-current `retention.go` (N6); AC-005 steps as separate subtests (N4); an optional end-to-end case for a user-owned link to a root-owned target (N5); the completion report states `Windows runtime not observed` in Gaps and names the red commit T on one line.
3. DoD 1 wording (N1) is accept-as-debt.

## Evidence-bearing sections (per `verification-claim-integrity.md` §3)

**Claim.** PASS at 0.92 with seven non-blocking findings (N1-N7); B1-B3 fixed; MP-1..MP-9 pass or N/A.

**Evidence.** The commands and verbatim output lines are in the Evidence section above and in Must-Pass Results.

**Baseline-attribution.** Every measurement was taken in this run, against tree `c2cc320805f5bd7a96df38e8e7cf1b44a361fbec` (Go code equal to base `1e2151a38` and to `2ebc10f8f`), darwin, uid 501, go1.26.8; the judging `moai` build is `45600e4ee`-derived (`v3.2.0-rc.26`), a strict ancestor of HEAD with `internal/spec` unchanged between them.

**Gaps (explicitly unobserved).**
- G-A: the Bash tool does not surface exit status 1 for `grep` and `git diff`; the exit codes of E-010, E-011, E-013, E-014 were inferred from their output (and, for E-014, from the same range's `--exit-code --stat` showing a diff), not observed.
- G-B: refused commands (verification-claim-integrity §3.1). The inline CN-4 awk program, the D8 per-section awk program, and `awk -f <scratch file>` were refused by the worktree guard ("what it runs cannot be shown not to be git"). Substitution: `bash <scratch>/audit2/cn4.sh` and `bash <scratch>/audit2/trace.sh`, scripts the iteration-2 auditor wrote from the same verb text with the SPEC paths hard-coded; I did not diff them against the verb text byte for byte, and I read the candidate lines independently with `grep -n -i -E "before|after|first|prior to|pre-change" acceptance.md`. D8 was done by `grep -c syscall`, not by the section-scoped program. A `git diff --quiet` and `sh -c '...; echo rc=$?'` form were also refused or not informative.
- G-C: the kanban environment variables were not scrubbed (a separate or compound `unset` is refused by the guard); the tests read no kanban state and their failing lines equal the ledger's scrubbed runs.
- G-D: no Linux or Windows observation of anything (E-026, E-027, the read-only-handle seam, the FIFO and stderr tests, the `/bin` candidate); every run was darwin uid 501.
- G-E: the RG criteria AC-004, -005, -006 and AC-014's kill have no RED-now by construction (the SPEC's G-1, G-2); no mutation run was executed (the code does not exist).
- G-F: the `release-pr-multi-os.yml` citations (`:98`, `:152`, `:210`) were not read; the `ci.yml` citations were read (N3).
- G-G: no MCP audit tool (`audit_multi`, `claude_audit`, `codex_audit`, `glm_audit`) was called; no receipts issued.

**Residual-risk.** The late-event tests depend on a 300 ms bounded delay (a slow runner can make them pass vacuously, never fail falsely: G-4 of the ledger); the residual loss window, the heal-burst gap and the TOCTOU gap are disclosed in the SPEC as unmeasured and not closed; the pre-existing FIFO hang at the state path remains; N4 and N5 leave two mutant shapes unkilled in CI on some platforms.

## Operational Notes (unverified)

- Measure the CI Windows vet coverage before relying on N3's correction: `grep -n -B2 -A8 "Vet cross-platform" .github/workflows/ci.yml` (status: measured for the quoted lines, inferred for the claim that it runs on every PR; the job is named `Build (...)` and is "always runs — required check" per its header at ci.yml:490).
- Measure that M2 produced a Linux-qualifying (e) candidate on the first CI run: look for `--- SKIP` under `TestOwnerCheckDefault` in the `Test (ubuntu-latest)` log (status: assumption, no Linux observation).
- Measure the heal burst if it ever matters: N concurrent `PruneStaleEntries` processes against one faulty state entry, counting archive writes (status: assumption; the SPEC marks it unmeasured).

AUDIT-VERDICT: PASS spec=SPEC-HARNESS-RETENTION-HARDEN-001 receipts=none
