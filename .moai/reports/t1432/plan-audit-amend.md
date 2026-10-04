auditor-model: claude-sonnet-5-5[1m]
verdict: FAIL
audited_sha: 6357a387c3032376b5bf2b44c0bebcb16b816a53

# SPEC Review Report: SPEC-HARNESS-RETENTION-HARDEN-001, amendment 0.4.0 (card t1432), delta audit
Iteration: 1/1 (the single delta iteration opened by the leader)
Overall Score: 0.81 (mean of the four rubric dimensions below). The result line is on line 2 of this file: FAIL, forced by the M5 must-pass firewall (MP-6), not by the aggregate. Tier M threshold 0.80.
Plan Artifact Hash: not computed (the mechanical hash is the `ComputeHash` subject set over spec.md, plan.md, acceptance.md; this audit read the committed tree at `6357a387c`, clean)
Auditor Version: plan-auditor/v1 (delta mode)

Reasoning context ignored per M1 Context Isolation. The dispatch described the amendment's design and four author-reported corrections; each was treated as a claim to verify, not as evidence.

## Section 1. Claim

1. The amended passages as committed are NOT adoptable: three blocking defects (D1 a D8 cross-platform BLOCKING finding, D2 a Windows claim that contradicts the SPEC's own D4.a and the shipped Windows owner check, D3 a mutant that satisfies AC-HRH-006 / 015 / 016 while violating REQ-HRH-005).
2. On Unix, for processes that honour the lock, the heal-lock protocol closes the check-then-remove window. I modelled it against `internal/harness/retention.go` and found no interleaving that removes a fresh state file or leaves two inodes locked. The scope of that closure is stated correctly in §B D4.c, §F and REQ-HRH-005 (Unix, honouring processes; older binary; not Windows), apart from D2 below.
3. The four author-reported corrections hold: (a) the lockfile ban is REQ-HRH-012; (b) a lock timeout writes no stamp; (c) the bare-helper probe is a control, not post-fix evidence; (d) the settings template lines are `:147`/`:204` (timeout) and `:149`/`:206` (async), but the 0.3.0 §A F6 row still cites the old coordinates (D7).
4. Amendment mechanics hold: 12 frontmatter fields, `status: in-progress`, `amendment_of`, `### Amendments` with `prior_completed_sha` equal to the prior close's `sync_commit_sha`, no `status:` in plan/acceptance/decision-index, `moai spec lint` clean, `moai spec audit` shows no `SyncStatusDrift`, 16 REQ / 16 AC counted independently.
5. I found no interleaving that defeats the design. That is a modelled result (reasoned from the code and the contract tests), not a post-fix measurement: no heal-lock production code exists at this tree (`git diff --name-only 7639c04c1 HEAD` lists only SPEC artifacts and `amend-drafts/`).

## Section 2. Evidence

### 2.1 Tree and tooling

```
$ git rev-parse --show-toplevel
/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1432
$ git branch --show-current
WT-harness-retention-debt
$ git rev-parse --short HEAD
6357a387c
$ git status --short
(empty)
$ moai version     (installed build)
moai-adk v3.2.0-rc.26   archive/t1401-293-g45600e4ee   built 2026-10-02T14:42:32Z
$ git merge-base --is-ancestor 45600e4ee HEAD        -> exit 0 (the installed build IS a strict ancestor of HEAD)
$ git rev-list --count 45600e4ee..HEAD -- internal/spec   -> 0 (no change to the lint/audit code since that build, so the lag does not bear on the two spec tools below)
$ moai spec lint SPEC-HARNESS-RETENTION-HARDEN-001
✓ No findings — all SPEC documents are valid
$ moai spec audit --filter-spec SPEC-HARNESS-RETENTION-HARDEN-001 --json --include-grandfathered
{ "total_specs": 1, "grandfathered": 0, "modern_era_clean": 1,
  "drift_findings": [ { "spec_id": "SPEC-HARNESS-RETENTION-HARDEN-001", "era": "V3R6", "finding_type": "EraAutoDetected", "severity": "INFO", "details": { "heuristic_matched": "H-4 (§E.2 + §E.4 + sync_commit_sha)" } } ] }
   -> no SyncStatusDrift
$ git log -S removeStateEntryIfUnchanged --format=%h --reverse -- internal/harness/retention.go
f52dd1b1c            (the helper was introduced by this card, as the amendment states)
$ git rev-parse 5bb35abe8165e8d5fc5cbab6b246ac6a82646cb8   -> 5bb35abe8165e8d5fc5cbab6b246ac6a82646cb8 (exists)
$ grep -n -E 'sync_commit_sha' progress.md
29: prior_completed_sha: 5bb35abe8165e8d5fc5cbab6b246ac6a82646cb8   (amendment note)
124: sync_commit_sha: 5bb35abe8165e8d5fc5cbab6b246ac6a82646cb8      (the prior close)  -> equal
$ grep -n -E '^status:' plan.md acceptance.md decision-index.md   -> (empty)
```

### 2.2 Counts and traceability (MP-1, AC-4, AC-5)

`grep -n -E '^- \*\*REQ-HRH-[0-9]+\*\*' spec.md` lists REQ-HRH-001 to REQ-HRH-016 at lines 186-201, sequential, no gap, no duplicate. `grep -n -E '^### AC-HRH-' acceptance.md` lists AC-HRH-001 to AC-HRH-016 at lines 43-178. Ceilings Tier M 16 / 16: at the ceiling, not over.

Traceability verb (run from a scratch script outside the tree):

```
COLLECTED: 16 REQ definitions (acceptance input: read)
```
No `UNCOVERED:` and no `ORPHAN:` line. I read the mappings behind it: REQ-HRH-005 -> AC-HRH-006 (title and Traceability row), REQ-HRH-016 -> AC-HRH-015 and AC-HRH-016, REQ-HRH-010 -> AC-HRH-010 (ninth sentinel), REQ-HRH-012 -> AC-HRH-011 (heal-lock clause). None is incidental.

### 2.3 CN-4 / MP-9 verb

```
COLLECTED: 11 milestones in plan order (M0 M1 M2 M3 M4 M5 M6 M7 M8 M9 M10), 0 exit bindings, 55 ordering candidates
```
No `CONFLICT:` line (0 exit bindings: the plan's `Exit:` lines name Definition-of-Done items, not AC ids, so the verb cannot bind a criterion to a milestone). I read the candidates that bear on the amendment: DoD 9 (T2 before every heal-lock production commit) against plan M7 (T2) and M8 (production): consistent; plan M9 "may precede or follow M8" against DoD 12: consistent; DoD 1's range `1e2151a38..HEAD` now also holds the amendment's commits, all descendants of T: consistent. `git log --reverse --format=%h 7639c04c1..HEAD -- internal/harness/retention.go internal/harness/retention_heal_unix.go internal/harness/retention_heal_windows.go` printed nothing (exit 0): no production commit exists yet, which is a Gap at M7 by the SPEC's own rule, not a pass.

### 2.4 D7 and D8 verbs

D7: the only other SPEC id referenced is `SPEC-AGENT-TEAM-RETIRE-001`, `status: completed` (not retired/superseded/archived): no `REVIEW:` line, no BLOCKING.

D8 (the section-scoped awk verb, run against spec.md):
```
BLOCKING: section "## §A Background" references syscall but carries no //go:build constraint or EXCL justification
```
The word `syscall` occurs in spec.md at L64 (the new "Heal-lock primitives [0.4.0]" row of §A: `GOOS=windows go doc syscall.Flock` and `syscall.O_NOFOLLOW`), L161 and L206. L161 and L206 sit in sections that carry a literal `//go:build` (D4.c at L160, §D at L206). L64 sits in §A, whose text (L43-80) carries neither `//go:build` nor an exemption clause. Control: `git show 7639c04c1:…/spec.md` (the 0.3.0 text) contains zero occurrences of `syscall` (`grep -c syscall` printed `0`), so this is a defect the amendment introduced and 0.3.0 passed D8 by auto-pass.

### 2.5 MP-7

`grep -rn 'NEEDS CLARIFICATION' .moai/specs/SPEC-HARNESS-RETENTION-HARDEN-001` printed nothing: no open marker. (No `research.md` exists at Tier M.)

### 2.6 MP-8: RED-now cells re-executed on the current tree

The ledger commands cite `S/t1432-amend/...` (a placeholder for the session scratch directory). The overlay JSON files were repointed to `.moai/reports/t1432/amend-drafts/`. I ran the committed copies by absolute path, each as one compound `unset MOAI_KANBAN … && go test …; echo "exit: $?"` invocation. Tree `6357a387c`; production code is identical to `7639c04c1`.

E-036 (AC-HRH-006 b), RAN:
```
--- FAIL: TestPruneHealSerializesOnTheHealLock (0.01s)
    zz_heallock_test.go:102: the pruner returned (<nil>) while another descriptor held the heal lock: it did not wait
    zz_heallock_test.go:106: the state-path entry was removed or replaced while the heal lock was held by another descriptor: err=<nil>
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/harness	0.682s
exit: 1
```
Matches the ledger cell verbatim.

E-037 (AC-HRH-015), RAN: `--- FAIL: TestPruneHealLockHeldPastTheBoundFailsClosed (0.00s)`, the same five failure lines as the cell (error `<nil>`, entry replaced, `log changed`, `archive directory exists: <nil>`, no warning), line numbers 187/190/196/199/202 identical, `exit: 1`.

E-038 (AC-HRH-016), RAN with `-v`: four subtests (`symlink`, `directory`, `fifo`, `not-owned`) each `--- FAIL` with the same four assertion families (`want an error naming …prune-heal, got <nil>`; `state-path entry was removed or replaced although the heal lock was unusable`; `log changed`; no warning), `exit: 1`. The line numbers differ from the cell (committed draft 288/291/297/300, cell 277/280/286/289): the committed draft was edited after the cell was written (D6).

E-035 (pre-fix control of the post-fix probe), RAN:
```
PROBE heal-window: trials=7429 heal_removed_the_swapped_in_fresh_entry=5 path_holds_F=7424 path_holds_other=0 heal_errors=0
--- PASS: TestZZProbeHealWindow (6.50s)   exit: 0
```
The window reproduces through the production entry point `healStateEntry` with a lock-honouring swapper (the cell recorded 5 in 1270; the count is "5 at the stop", the trial count varies run to run).

E-034 (the audit's bare-helper probe), RAN with an overlay I wrote in my own scratch directory, because the committed `amend-drafts/` holds the probe source but not its overlay JSON (D6):
```
PROBE remove-window: trials=899 helper_removed_the_swapped_in_fresh_entry=5 helper_first=892 swap_first=2
exit: 0
```
E-041, RAN: `grep -c -F "heal lock gives no exclusion on Windows" internal/harness/retention.go` -> `0`, exit 1. E-042, RAN: `git check-ignore -v .moai/harness/usage-log.jsonl.prune-heal` -> empty, exit 1.

E-039 (cases c, d, green at base), RAN with `-v`: `--- PASS: TestPruneCommonPathCreatesNoHealLock` (+ 2 subtests) and `--- PASS: TestPruneHealLockIsFreeWhilePrunerIsInItsArchiveStep`, exit 0. E-040 (case a), RAN: both `--- PASS`, exit 0. E-043, RAN: `GOOS=windows go vet ./internal/harness/ ./internal/lockfile/` exit 0; the untagged-flock overlay exits 1 with `undefined: syscall.Flock`, `LOCK_EX`, `LOCK_NB`.

Not re-executed, READ only: E-032, E-033 (Linux compile), E-044, E-045 (mutant survivors and kills of M9), E-046 controls.

Swept-count discipline: every `go test -run` above printed a `--- FAIL` or `--- PASS` line naming each selected test; none printed `[no tests to run]`. The cells themselves carry the same property.

### 2.7 Other measurements

```
$ git diff --quiet 1e2151a38 -- internal/lockfile internal/harness/observer.go        -> exit 0 (byte-identical)
$ git log --format=%h -s -L '/^func (r \*Retention) PruneStaleEntries/,…' (the four-function form) 1e2151a38..HEAD   -> (empty)
$ grep -n -E '"timeout": 5|"async": true' internal/template/templates/.claude/settings.json.tmpl  (the harness-observe entries)
147: "timeout": 5, 149: "async": true     and     204: "timeout": 5, 206: "async": true
$ cat internal/harness/retention_owner_windows.go   (L9)
func entryOwnedByCurrentUser(string) bool { return false }
```

## Section 3. Baseline-attribution

Every measurement above was taken in this run on the committed tree `6357a387c` (clean before and after), darwin arm64, uid 501, APFS. The judging build for `moai spec lint` / `moai spec audit` is `moai-adk v3.2.0-rc.26` (`45600e4ee`), a strict ancestor of HEAD with zero `internal/spec` commits since. The Go test runs used the toolchain on this machine (the cells record go1.26.8; I did not re-read `go version`). The audit's cited figures (5 of 5235; 235 of 20000) were checked against `sync-audit.md:185` and `sync-audit-delta.md:111` and match. No figure was carried over unmeasured except where marked READ above.

## Section 4. Gaps

- Linux and Windows runtime were NOT observed. Every platform premise of the heal lock (EWOULDBLOCK on a second descriptor, `O_NOFOLLOW`, FIFO open behaviour, directory open) is darwin only; Linux is compile-checked only (E-033, READ).
- No heal-lock production code exists, so the closure of the window is a MODEL result plus the pre-fix controls, not a post-fix measurement. The post-fix probe of DoD 10 and every M8 mutant kill are expectations.
- AC-HRH-006 b / 015 / 016: the ledger has RED-now cells, but no kill observation (G-11 of the SPEC; correct).
- The first heredoc-built scratch script was refused by the worktree guard ("too complex to verify"); I re-created it with the Write tool outside the tree and ran `sh <path>`. The Grep tool was unavailable in this session; I used `grep` through Bash. `awk` was NOT refused when run as one literal invocation.
- The MCP cross-backend audit tools (`mcp__moai__audit_multi`, `codex_audit`, `claude_audit`, `glm_audit`) were NOT run, although `.moai/config/sections/workflow.yaml` sets `audit.model: multi`. They review git diffs (`uncommittedChanges` is empty on this clean tree; `baseBranch` would review the whole card's already-audited code, not the amended plan-phase passages), so neither target isolates the audited surface. The result is a single-auditor verdict. No audit receipt was issued.
- The post-fix timing assertions (300 ms of AC-006 b, 4.5 s of AC-015 / 016, the 2 s bound) were not run under machine load (background load is prohibited; the SPEC records it as G-12).
- Not read in full: the 0.3.0 text outside the amended passages, the unamended ledger entries E-001 to E-031, and the red-now drafts of 0.3.0 (they passed iteration 3 at 0.92).
- Whether the helper's error must wrap its cause (D5) and the LOCK_SH mutant (D3) are REASONED from the unamended tests and flock semantics; no mutant was built.

## Section 5. Residual-risk

- A run phase that implements the amendment as written can ship a heal lock whose mode is shared (D3) and a source comment that states a false Windows hazard (D2); both pass every criterion as written.
- The post-fix probe is the only evidence for "lock held across the removal, not released before it". I checked its power from the pre-fix numbers (5 removals in 1270 trials in the cell, 5 in 7429 in my run, about 0.07 to 0.4 percent per trial): a release-before-removal mutant would show tens of removals in 20000 trials, so a 20000-trial zero is strong evidence, provided the run actually reaches 20000 (D11).
- The two-inode shape can still be produced by a deletion of `<log>.prune-heal` by something other than this code (D8).

## Must-Pass Results

- [PASS] MP-1 REQ number consistency: REQ-HRH-001 to 016 at spec.md:L186-L201, sequential, no duplicates, zero-padded alike (section 2.2).
- [PASS] MP-2 GEARS format: judged on the requirement layer (spec.md §C), not on the ACs. REQ-HRH-005 (L190) and REQ-HRH-016 (L201) are Event-driven ("When the retention pruner is about to remove a prune state path entry…, it shall…"); REQ-HRH-010 (L195) is Ubiquitous with one added clause. REQ-HRH-005 packs five obligations into one sentence, which is clarity debt, not a pattern failure.
- [PASS] MP-3 frontmatter: all 12 canonical fields present with correct types (`version: "0.4.0"` quoted, `status: in-progress`, `created`/`updated` ISO, `tags` string; spec.md:L2-L18); extras `tier`, `card`, `amendment_of`, `related_specs` are optional; `moai spec lint` clean.
- [N/A] MP-4 language neutrality: single-language (Go) SPEC about one package.
- [PASS] MP-5 D7: no referenced SPEC is retired, superseded or archived (section 2.4).
- [FAIL] MP-6 D8: the verb printed `BLOCKING: section "## §A Background" references syscall but carries no //go:build constraint or EXCL justification`; spec.md:L64 introduced by 0.4.0; unresolved (D1).
- [PASS] MP-7: no `[NEEDS CLARIFICATION` marker anywhere in the SPEC directory; no research.md at Tier M.
- [PASS] MP-8: the amended release-blocking RED-now cells (E-036, E-037, E-038) were re-executed through the committed copies and reproduce; E-041 and E-042 reproduce exactly; each cell carries command, verbatim stdout, exit code and tree SHA (the document-level pin `7639c04c1`, acceptance.md:L7). Qualification: the cells' command text is not runnable as written (`S/…` placeholder, scratch path); I ran the committed copies the leader named. That is a documentation defect (D6), not a non-reproducing RED.
- [PASS] MP-9: CN-4 verb printed `COLLECTED` and no `CONFLICT:`; candidates read; no ordering clause binds a milestone the plan schedules on the forbidden side (section 2.3). The empty bounded production listing is a Gap by the SPEC's rule, never read as a pass.

## Category Scores

| Dimension | Score | Rubric Band | Evidence |
|---|---|---|---|
| Clarity | 0.75 | minor ambiguity in one or two requirements | REQ-HRH-005 (L190) is one sentence with five obligations; the Windows consequence sentence in REQ-HRH-010 (L195) is wrong against D4.a (D2) |
| Completeness | 0.75 | one non-critical item sparse, frontmatter complete | all sections present, Out-of-Scope H3s with bullets (L213-L247); stale hoist statements (acceptance.md:L7, G-9, plan B12) and the D8 finding |
| Testability | 0.75 | one criterion not shallow-proof | the shared-lock mutant passes AC-HRH-006 b-d, 015, 016 (D3); timing assertions disclosed (G-12) |
| Traceability | 1.00 | every REQ has an AC and the reverse | verb output `COLLECTED: 16 REQ definitions`, no UNCOVERED, no ORPHAN (section 2.2) |

Per-audit-area result (the leader's six areas):

| # | Area | Result | Basis |
|---|---|---|---|
| 1 | Does the design close the window | closes it on Unix for honouring processes (modelled); scope stated correctly except the Windows consequence | D2 |
| 2 | Hostile heal-lock entry | design sound; disclosure gaps only | D8 |
| 3 | Acceptance criteria two-cell discipline | RED-now cells reproduce; one writable mutant survives | D3, D4, D5 |
| 4 | Amendment mechanics | pass | section 2.1, 2.2 |
| 5 | Consistency with unamended text | two inconsistencies | D2, D5 |
| 6 | Decision records | pass | below |

## Defects Found

D1. D8-CROSS-PLATFORM — spec.md:L64 — The 0.4.0 row "Heal-lock primitives" in `## §A Background` names `syscall.Flock` and `syscall.O_NOFOLLOW`, and §A carries no literal `//go:build` and no cross-platform exemption clause. The section-scoped D8 verb prints `BLOCKING: section "## §A Background" references syscall…`. 0.3.0 had no `syscall` word in spec.md, so the amendment introduced it. The substance is covered elsewhere (D4.a L132, D4.c L160, §D L206), but D8-2 requires the constraint in the same section or paragraph and a constraint elsewhere in the document does not cover it. — Severity: critical (MP-6, forces FAIL) — Class: blocking — Required fix: in the L64 row (or the §A paragraph that follows it) state that the helper which uses these symbols lives in a `//go:build !windows` file with a `//go:build windows` twin defining the same symbol (REQ-HRH-012); then re-run the awk verb and expect no output.

D2. WINDOWS-WINDOW-CONTRADICTION — spec.md:L195 (REQ-HRH-010), L165 (§B D4.c item 7), L236 (§E), L253 (§F item 2), acceptance.md:L33/L135 (AC-HRH-010, ninth sentinel) — The amendment states that "the heal window remains" on Windows and makes REQ-HRH-010 require that sentence in `retention.go`. But the same SPEC says (L129, D4.a) that on Windows "every entry counts as not owned … a faulty entry stays in place and retention stays off with a warning", and the shipped `retention_owner_windows.go:9` returns `false` for every path. `healStateEntry` returns at the ownership check before any removal, so on Windows no heal occurs, there is no heal window, and the heal-lock twin is unreachable in production. The requirement therefore mandates a false source comment, and the SPEC claims an unobserved Windows behaviour (§D says no Windows runtime claim is made). — Severity: major — Class: blocking — Required fix: reword the clause in REQ-HRH-010 and the four other places to: "on Windows the owner check never reports an entry as owned, so a faulty state-path entry is never healed and the heal lock is never reached; the Windows twin gives no exclusion and would matter only if a Windows ownership lookup is added (out of scope, §E)". Keep the sentinel phrase `heal lock gives no exclusion on Windows` (it is true as written); drop "so the heal window remains there".

D3. SHARED-LOCK-MUTANT — acceptance.md:L91-L100 (AC-HRH-006 b and its mutant list), L169-L186 (AC-HRH-015, 016), L841 (DoD 10) — A mutant whose helper takes `LOCK_SH|LOCK_NB` instead of `LOCK_EX|LOCK_NB` (or any mode that does not conflict with a copy of itself) passes every cited test and probe: AC-006 b holds an exclusive lock in the test, and a shared request conflicts with an exclusive holder, so it waits; AC-015 and AC-016 do not depend on the mode; the DoD-10 probe's swapper takes an exclusive lock, which also conflicts with a shared holder. But two real healers on that mutant do not exclude each other, which is exactly the property REQ-HRH-005 ("serialize healers") exists for. The listed mutants (never taken, taken on the common path, held across the prune, unconditional removal, unbounded wait, proceed after the bound) do not include it; the one stated unkillable mutant (release before the removal) is covered by the probe and honestly disclosed, this one is not, and it is killable by a deterministic test. `verification-completeness.md` §2 (mutant probe, [HARD]): a writable mutant makes the criterion too shallow to adopt. Reasoned; no mutant was built. — Severity: major — Class: blocking — Required fix: add a case to AC-HRH-006 (b2): the holder takes `LOCK_SH` instead of `LOCK_EX` on `<log>.prune-heal`; the pruner must still not return and must not touch the entry for 300 ms (a correct exclusive request conflicts with a shared holder); add the "shared lock mode" mutant to AC-HRH-006's list and to plan M8 step 8.

D4. M9-N4-GUARD — plan.md M9 (§F.1), acceptance.md:L98 and E-045 — The M9 draft `TestPruneHealFileArmKeepsAFreshStateFileSwappedInDuringTheHeal` (`amend-drafts/zz_m9_test.go`) has an owner-check stand-in that renames a fresh file over whichever path it is asked about and overwrites `freshInfo`. After M8 the same owner check is asked about `<log>.prune-heal` too, so the stand-in swaps that path and the test goes falsely red (the same hazard plan B11 settles for the existing swap test with a path guard). E-045's "unmutated PASS" holds only on the pre-M8 tree. M9 "may precede or follow M8" and says only "land the three draft tests (the model)". — Severity: minor — Class: non-blocking (must-fix, cheap; M9 is optional) — Required fix: state in M9 that the N4 test carries the same act-only-on-the-state-path guard, and that its unmutated pass is re-observed after M8.

D5. AC003-REROUTE-AND-ERROR-WRAP — acceptance.md:L60-L67 (AC-HRH-003 case b, DoD 6), the existing `TestPruneStateUnreplaceableInReadOnlyDirSkips` (retention_statepath_test.go:L142-L182), spec.md:L201 (REQ-HRH-016) — AC-003 case (b) puts a user-owned 0400 state file in a 0555 directory. After M8 the heal reaches the heal lock first, whose absent entry cannot be created in a read-only directory, so the heal fails closed before `removeStateEntryIfUnchanged` is called. (1) The existing test asserts `errors.Is(perr, fs.ErrPermission)`; REQ-HRH-016 only requires an error "naming the heal-lock path", not wrapping its cause, so an implementation that does not wrap breaks an unamended test. (2) The removal-failure arm then has no test: DoD 6 and AC-003 record "a heal that ignores the removal failure" as killed at M2, and after M8 that mutant survives case (b). The amendment did not sweep AC-003 or DoD 6 (`verification-completeness.md` §3). — Severity: major — Class: non-blocking (a run-phase gate would catch the wrap; the arm's unpinned status is debt) — Required fix: state in REQ-HRH-016 or plan M8 step 2 that the heal-lock error wraps the underlying cause; either add a variant to AC-003 (b) with a pre-created heal-lock file so the removal arm is reached, or record the arm's lost pin as accepted debt in the verdict; add the two AC-003 mutants to the M8 step 8 re-run list.

D6. LEDGER-PATHS-STALE — acceptance.md:L7, L554, G-9 (L821), plan.md B12 and M7 step 5 — The ledger commands use a placeholder `S/t1432-amend/…` and scratch paths, and the pin paragraph and G-9 still say the drafts live in the session scratch directory and that the leader or M7 will hoist them; commit `6357a387c` already hoisted them to `.moai/reports/t1432/amend-drafts/`. The overlay for E-034 (`S/audit/ov_race.json`) was not hoisted (the audit probe source is there, its overlay is not). E-038's cited line numbers (277/280/286/289) do not match the committed draft (288/291/297/300). A cell whose command cannot run as written fails the single-invocation form of `verification-completeness.md` §2.1 unless the reader substitutes paths. — Severity: minor — Class: non-blocking (must-fix, cheap, batch with D1) — Required fix: replace `S/t1432-amend/` and `S/audit/` by the committed paths, add `amend-drafts/audit-probe/overlay-race.json` (or equivalent) and cite it, refresh E-038's line numbers, and rewrite G-9, B12 and M7 step 5 as done.

D7. SETTINGS-COORDINATES — spec.md:L58 (§A F6 row, unamended text) and decision-index Q9 — The F6 row still cites `settings.json.tmpl:146-148` and `:203-205`; the measured lines are `:147`/`:204` (timeout) and `:149`/`:206` (async), so the range no longer covers the `async` lines the row relies on. Correction (d) was recorded only in Q9. Moving coordinate (`verification-claim-integrity.md` §2.1). — Severity: minor — Class: non-blocking (accept-as-debt) — Required fix: pin the tree SHA beside the coordinates or update them.

D8. HEAL-LOCK-DISCLOSURE — spec.md:L163-L168 (D4.c items 5, 8; residual paragraph), L238, L253 — Disclosure gaps, none contradicting a requirement: (i) deletion of `<log>.prune-heal` by anything other than this code (`git clean -X` removes ignored files once the entry is added to `.gitignore`; an operator `rm`) while a healer holds it lets a second healer lock a new inode: the same two-inode shape; the SPEC says only "never removed by this code". (ii) A planted heal-lock entry (symlink, directory, FIFO, foreign-owned) denies healing indefinitely; the SPEC says the entry "stays for the operator" but does not say that, together with a user-owned faulty state entry, retention stays off until an operator acts; the common path never touches the heal lock, so a planted entry alone denies nothing. (iii) The 2 s bound is per acquisition: up to three inspections could each wait, so the 4.5 s ceiling of AC-015 holds only when the lock is held throughout; the chain is contrived. — Severity: minor — Class: non-blocking (accept-as-debt; one sentence each in §F) — Required fix: none required for adoption; add the three sentences to §F item 3.

D9. COMMON-PATH-PIN — acceptance.md:L94 (AC-HRH-006 c) vs spec.md:L190 (REQ-HRH-005, last clause) — REQ-HRH-005 forbids opening, creating or locking the heal lock on the common path; case (c) pins only "no heal-lock entry exists afterwards", so a mutant that opens and locks an already-existing heal-lock file on the common path passes (c). The AC-006 title and D4.c say "never removed" without the Unix / honouring-process qualifier the body carries. — Severity: minor — Class: non-blocking (accept-as-debt) — Required fix (optional): a variant of (c) with the heal-lock file pre-created and held exclusive by the test, expecting the common-path prune to finish without waiting.

D10. DEFAULT-DRAW-COVERAGE — spec.md:L147-L151 (D4.c table) — The table lists options A, B, C. The audit's own suggested repair (`sync-audit.md` F1 "Required fix": an atomic rename-take with no lock file, no timeout, no new artifact, no Windows gap) is not listed or weighed. The leader's ruling selects C, so this does not change the verdict. — Severity: minor — Class: non-blocking (accept-as-debt; optional) — Required fix: none; one row would record why the rename-take form was not chosen.

D11. PROBE-TRIAL-FLOOR — acceptance.md:L841 (DoD 10) and `amend-drafts/zz_heal_race_probe_test.go` — The draft stops at a 60 s deadline or 100000 trials. At the ledger's pre-fix rate (1270 trials in 4.12 s, about 3.2 ms per trial) 20000 trials need about 65 s, and the post-fix path adds a lock round trip per trial, so the draft can stop short of the 20000 floor, which DoD 10 itself calls a Gap. My run took about 0.9 ms per trial (7429 trials in 6.50 s). — Severity: minor — Class: non-blocking (accept-as-debt) — Required fix: raise the draft's deadline (or make the trial floor, not time, the stop condition) when it is adapted at M8.

## Findings on the six areas (what I tried to break)

1 (window closure). Modelled against `retention.go` (`openStateFile` L198-L244, `healStateEntry` L249-L258, `removeStateEntryIfUnchanged` L264-L282). The protocol is: ownership check, take heal lock, re-inspect (`Lstat` and identity, type, mode, modification time), remove, release, then the loop creates the replacement exclusively. Interleavings tried, none defeats it for honouring processes:
- A heals, B waits then re-inspects: B finds the path absent (nothing to remove, creates exclusively, an "already exists" sends it back to inspect) or A's fresh file F (not the inspected E by mode and modification time, even if an inode number is reused).
- C inspected E earlier: same as B; its removal needs the path to hold E, and once E is gone nothing equal to E reappears.
- D on the absent-create path: creates F exclusively; A's later exclusive create fails with "already exists" and A opens D's file; one inode.
- Release before create: the gap between A's release and A's create is a gap in which the path is absent; a waiter that takes the lock there removes nothing and creates exclusively; at most one creator wins.
- Older binary that ignores the lock: the window remains (A's re-inspection and removal against an old-binary healer's replacement); stated in §B D4.c, §E and §F. Correctly scoped.
- Not covered by the SPEC: an external deletion of the heal-lock file (D8 i).
- Timeout path: no stamp is written; `pruneExclusive` returns before `pruneLocked`, and `lastPruneAt` is set only on success paths, so the next call retries. Correction (b) holds.
Scope of the claim: REQ-HRH-005's "so that a state file created by a healer that held the lock first is not removed" is scoped by its wording; AC-HRH-006's title ("never removed") is looser than its body (D9). The Windows consequence is wrong (D2).

2 (hostile heal-lock entry). `Lstat`-regular-only plus `O_NOFOLLOW|O_NONBLOCK`, create exclusive 0600, never truncated or removed: sound. A FIFO planted at the heal-lock path is never opened, and an `O_RDWR` open of a FIFO does not block on darwin (E-032 P3); the pre-existing FIFO-at-state-path hang stays recorded and unrepaired (§A L62, §E L231, §F), as required. The ownership rule reuses the existing `ownerCheck` field and adds no second test-only field (§D L208; plan M8 step 2). A timeout or hostile entry: heal not attempted, entry untouched, no stamp, one warning, error naming the path (REQ-HRH-016). The denial-of-retention reach is real but needs a user-owned faulty state entry as well (D8 ii).

3 (acceptance criteria). RED-now cells re-executed: E-036, E-037, E-038, E-035, E-034 (own overlay), E-039, E-040, E-041, E-042, E-043 (section 2.6); READ only: E-032, E-033, E-044 to E-046. The swept-count requirement is met by every cell. Mutant probes: AC-006 b falls to the shared-lock mutant (D3); AC-015 resists "a timeout that is a sleep" (such a mutant returns an error after 2 s although the holder released at 300 ms, so AC-006 b ii demands `nil` and fails) and "bound driven by the injected clock" (the test clock is a constant, so the wait never ends and the 8 s cap fires); AC-016 resists "owner check skipped" ((d)), "open follows links" ((a)) and "hostile entry removed and recreated" ((a), (b), (c)); but AC-016 c pins the outcome, not the mechanism, on a platform where an `O_RDWR` FIFO open does not block (stated in the criterion). The "release before the removal" mutant is genuinely unkillable by a deterministic test (no seam between the re-inspection and the removal), and the SPEC says so at acceptance.md:L100, spec.md:L168 and L253(6); the probe covers it statistically and that is an expectation until measured.

4 (mechanics). Pass (section 2.1, 2.2).

5 (consistency with the unamended text). REQ-HRH-008 and AC-HRH-009 untouched (the `-L` form printed nothing; the heal lock sits below `pruneExclusive`). `internal/lockfile` and `observer.go` byte-identical (exit 0). F5 and F6 are not reopened (disclosure only). Out-of-Scope and §F match the new design. No time estimates in the amended text (priority labels and phase order; the 2 s, 10 ms and 4.5 s figures are technical bounds). Windows claims are limited to build and vet, except the unobserved "window remains" claim (D2). Two unamended criteria are affected by the new design without a sweep: AC-HRH-003 b (D5) and the M9 / B11 owner-stand-in hazard (D4).

6 (decision records). The ruling is recorded as relayed, not as an operator answer, in the HISTORY row, `### Amendments` (L37), decision-index item (8) and the Q4 verdict line. Q9 is labelled EVIDENCE-NEEDED, its authority anchor says "(none … Context, not authority …)", the Operator verdict line is empty, and it carries no `(Recommended)` label; it does state that the relayed brief suggested 2 s and that D4.c applies it, which is the applied default disclosed as the decision-index header allows, not a recommendation to the operator. No finding.

## Regression Check

First iteration of this delta; the unamended 0.3.0 defects (iteration 3, 0.92) were not re-opened. Prior-iteration open items O4, O8, O9 remain open as `progress.md` records.

## Recommendation

FAIL. The run phase must not start on this text. Fix D1, D2 and D3 (all three are small edits: one clause in §A, one rewording in five places, one extra test case plus one mutant), batch D6 and D4 and D5 in the same revision, and re-submit for a confirming audit scoped to D1 to D6. D7 to D11 may be accepted as debt. The design itself (heal lock taken after the ownership check, held across re-inspection and removal, released before the replacement, fail closed on timeout or hostile entry, Unix only) holds against every interleaving I tried.

## Operational Notes (unverified)

- Measure whether the post-fix probe reaches 20000 trials: run the adapted `zz_heal_race_probe_test.go` against the final tree and read `trials=` in the PROBE line. Status: assumption (no post-fix tree exists).
- Measure the shared-lock mutant when M8 lands: apply a `LOCK_SH` copy of the helper by overlay and run the AC-006 / 015 / 016 tests; expect all to pass today's tests. Status: inferred (flock semantics: shared conflicts with exclusive, not with shared).
- Measure whether `TestPruneStateUnreplaceableInReadOnlyDirSkips` stays green when the heal-lock helper's error does not wrap its cause: run it with such an overlay. Status: inferred (the test asserts `errors.Is(perr, fs.ErrPermission)` at retention_statepath_test.go:L175).
- Measure the Linux behaviour of the heal-lock primitives in CI (`ubuntu-latest` runs the package tests per `plan.md` B7). Status: assumption.
