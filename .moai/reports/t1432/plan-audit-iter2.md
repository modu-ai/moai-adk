auditor-model: claude-sonnet-5-5[1m]

verdict: FAIL
audited_sha: 6003074c07141b2564b93d13ae4b64e1f665e1f6

# SPEC Review Report: SPEC-HARNESS-RETENTION-HARDEN-001

Iteration: 2/3
Verdict: FAIL
Overall Score: 0.87 (above the Tier M threshold 0.80 on its own; the FAIL is carried by three blocking findings, one of them an unsatisfiable ordering obligation treated as MP-9-class — see Must-Pass Results and Defects)
Plan Artifact Hash: not computed (the run-gate hash is a runtime record; this audit read the committed files at the SHA above)
Auditor Version: plan-auditor/v-unversioned (model line above)

Reasoning context ignored per M1 Context Isolation. The iteration-1 report was read only after the independent pass below.

## Claim

1. Iteration-1 blocking findings: D1 fixed, D3 fixed, D5 fixed, D7 fixed, D10 fixed, D4 fixed (MP-6 clean), D2 partly fixed (the tracked-test witness is right, the mechanical check written for it cannot be satisfied — new finding B1).
2. All release-blocking RED-now cells reproduce on the current tree, with a non-vacuous swept count.
3. The relayed operator decisions are honoured and recorded as relayed, except that the ownership pin leaves the "link's own owner" semantic untested (B2), and one premise the REQ text rests on is false when measured (B3).

## Tree and tool attribution

- Toplevel `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1432`, branch `WT-harness-retention-debt`, HEAD `6003074c07141b2564b93d13ae4b64e1f665e1f6`, `git status --short` empty before and after (all probes ran in the session scratch directory through `go test -overlay`; nothing was written in the tree except this file).
- Code under audit is identical to base: `git diff --quiet 1e2151a38 HEAD -- internal cmd` exit 0.
- Judging build: `moai version` -> `v3.2.0-rc.26 archive/t1401-293-g45600e4ee built 2026-10-02T14:42:32Z`. `git merge-base --is-ancestor 45600e4ee HEAD` exit 0 (a strict ancestor: the installed build predates HEAD). `git diff --quiet 45600e4ee HEAD -- internal/spec` exit 0, so the lint code the build carries equals HEAD's; the lag does not affect the lint result. The moai MCP server in this session is rc.25 (`802a72235`); no MCP audit tool was called (no `audit_model` key exists under `.moai/config`, `grep -rln audit_model .moai/config` printed nothing), so `receipts=none`.
- Platform of every run: darwin, uid 501, go as installed (`go1.26.x`; the ledger says `go1.26.8`).

## Must-Pass Results

- [PASS] MP-1 REQ numbers: `grep -n -E "^- \*\*REQ-HRH-[0-9]+\*\*" spec.md` -> REQ-HRH-001..015 at spec.md:155-169, sequential, no gap, no duplicate. AC-HRH-001..014 at acceptance.md:39-143.
- [PASS] MP-2 GEARS (judged on the `REQ-XXX` requirement layer in spec.md only; the Given-When-Then entries in acceptance.md are the verification layer and were not graded here): REQ-001..006 and 013 are event-driven (`When ... shall`), REQ-007..012, 014, 015 ubiquitous. REQ-012 packs four checkable clauses into one sentence (non-blocking style note).
- [PASS] MP-3 frontmatter: spec.md:2-17 carries all 12 canonical fields (id, title, version "0.2.0", status draft, created, updated, author, priority P2, phase "v3.2.0 target", module, lifecycle spec-anchored, tags string) plus `tier: M`, `card`, `related_specs`. `moai spec lint SPEC-HARNESS-RETENTION-HARDEN-001` -> `✓ No findings — all SPEC documents are valid`. No `status:` field in plan.md, acceptance.md or decision-index.md (the only `--- FAIL` hits are test output inside the ledger).
- [N/A] MP-4: single-language (Go) SPEC, internal package.
- [PASS] MP-5 D7: referenced SPEC ids: itself and `SPEC-AGENT-TEAM-RETIRE-001`, whose `status:` is `completed` (not retired, superseded or archived); no REVIEW line.
- [PASS] MP-6 D8: `grep -n syscall` finds no hit in spec.md or plan.md; the three hits in acceptance.md (lines 168, 278, 349) sit in the `## Evidence ledger` section, whose first paragraph carries the `//go:build !windows` / `//go:build windows` obligation (acceptance.md:168). The per-section awk program of the D8 verb was not run; the check was done by grep and reading (stated as a substitution, not a refusal).
- [PASS] MP-7: `grep -n -E "NEEDS CLARIFICATION" plan.md` prints nothing; research.md does not exist (Tier M).
- [PASS, with non-blocking finding O2] MP-8 RED-now re-execution: see the table in Evidence. Every release-blocking criterion's RED reproduces on the current tree with the ledger's verbatim stdout (timings and temp-directory names differ) and the ledger's exit code.
- [FAIL] MP-9 cross-artifact ordering: the CN-4 verb ran in full (`bash <scratch>/cn4.sh`) and printed `COLLECTED: 7 milestones in plan order (M0 M1 M2 M3 M4 M5 M6), 0 exit bindings, 26 ordering candidates` and no `CONFLICT:` line; I read all 26 candidates and none orders work across milestones. The FAIL is an auditor-confirmed ordering obligation that cannot be satisfied by the commit graph, found by reading plan.md and acceptance.md against `git log` (B1 below); it is not a verb output. The aggregate score does not absorb it.

## Category Scores

| Dimension | Score | Band | Evidence |
|---|---|---|---|
| Clarity | 0.85 | 0.75-1.0 | REQ-003 (spec.md:157) claims a FIFO "skips and returns an error" and D4.a (spec.md:113) says "a FIFO ... keeps today's behaviour: skip, return the error"; measured behaviour is a blocking hang (B3). REQ-002 (spec.md:156) says "cannot open for reading and writing" while plan.md:69 heals only on "a permission error". |
| Completeness | 0.95 | 1.0 | HISTORY, WHY/A, decisions B, REQ C, constraints D, `### Out of Scope — ...` H3s at spec.md:180,187,196 each with `-` bullets; Tier M ceilings 16/16 respected with 15 REQ / 14 AC. |
| Testability | 0.78 | 0.75 | RB cells are real and re-executable; but the DoD/M0 ordering check is unsatisfiable (B1), the ownership pin cannot kill a target-owner mutant (B2), AC-009 cannot see syscalls added in the pre-lock helper functions (O1). |
| Traceability | 0.93 | 0.75-1.0 | `trace.sh` (the AC-4/AC-5 verb): `COLLECTED: 15 REQ definitions (acceptance input: read)`, no `UNCOVERED:`, no `ORPHAN:`; REQ-HRH-008 maps only to AC-HRH-009, an indirect pin (O1). |

## Evidence (commands run, observed output)

### 1. Iteration-1 blocking findings

| Iter-1 | Status | Evidence |
|---|---|---|
| D1 M0 build break | FIXED | plan.md:46 keeps `retention_owner_test.go` and `retention_stampwrite_test.go` out of M0. The committed drafts `red-now-drafts/zz_*_test.go` use only symbols that exist at base (`writeStaleLog`, `logSubjectsForUnit`, `readLogLines`, `marshalEvent`, `logHasSubject` at `retention_stamp_test.go:19,30,35`, `retention_unparsed_test.go:29,38`; `stampSuffix` at `retention_stamp_test.go:16`). Through `go test -overlay` (scratch overlay built from the committed copies, `ov.json`) all of them compile and run RED against base: see the table below. |
| D2 gitignored witness | PARTLY FIXED | The witness is now the tracked tests, not the ignored report (plan.md:9, :51-53; acceptance.md:387). `git ls-files .moai/reports/t1432` lists the committed drafts, so force-add precedent holds. The mechanical check as written does not work: finding B1. |
| D3 unterminated tail | FIXED | REQ-007 (spec.md:161), AC-008 (acceptance.md:95-101), D3 rule (spec.md:101-106). RED cells E-004 and E-005 reproduce (below). One corner of the one-interval claim is overstated (O5). |
| D4 MP-6 build tags | FIXED | REQ-012 (spec.md:166); AC-011; D8 clean (MP-6). Controls reproduced: `GOOS=windows go vet -overlay <untagged>` -> `undefined: syscall.Mkfifo`, exit 1; the tagged overlay -> empty, exit 0; `GOOS=windows go vet ./internal/harness/ ./internal/lockfile/` at HEAD -> empty, exit 0. |
| D5 heal-burst claim | FIXED | spec.md:127-137 (conditional removal), :206 ("not measured"), AC-006 proof by deterministic interleaving. |
| D7 RED-now cells | FIXED | acceptance.md:166-375 ledger (command, stdout, exit, tree) plus classes RB/RG and a flips-at column. |
| D10 figure misquote | FIXED | spec.md:50 now cites `pruneExclusive 85.0%` (sync-audit.md:119) and `87.5%` (sync-audit-delta.md:100); both quoted figures match those files. |

Dropped or only partly closed findings: D13 (fast path unpinned) is only partly closed — see O1.

### 2. MP-8 re-execution (document pin `db6d88a2a`; code identical to HEAD)

Overlay JSON files in the repo name the original scratch directory; I rebuilt them from the committed copies in `.moai/reports/t1432/red-now-drafts/` (scratch `audit2/ov*.json`). `<A>` = `-overlay /private/tmp/.../scratchpad/audit2/ov.json`.

| Ledger | Cited criterion | Command re-run (substituted form) | Observed here | Matches ledger |
|---|---|---|---|---|
| E-001 | AC-001 | `go test -count=1 <A> -run '^TestPruneStateSymlinkReplacedTargetUntouched$' ./internal/harness/` | `--- FAIL ... victim changed: err=<nil> content="2026-10-02T00:00:00Z"` / `state path is not a regular file`, `FAIL ... 0.683s`, exit 1 | yes |
| E-002 | AC-002 | same, `TestPruneStateUnwritableFileReplaced` | five failure lines, `prune state open failed: ... permission denied`, exit 1 | yes |
| E-003 | AC-007 | `TestPruneCarriesLateEvents` | `subjects after prune: map[fresh:1]`, `late-event count = 0, want 1`, exit 1 | yes |
| E-004 | AC-008 b | `TestPruneTailPartialLineCarriedAndTerminated` | `partial fragment occurs 0 times, want 1`, exit 1 | yes |
| E-005 | AC-008 c | `TestPruneStaleFinalLineWithoutNewlineWaitsOneInterval` | `stale-final count after first prune = 0, want 1`, exit 1 | yes |
| E-006 | AC-013 survivor | `go test -count=1 -overlay <mB> ./internal/harness/` | `ok ... 1.646s`, exit 0 | yes |
| E-007 | AC-013 kill | mB plus the draft test | `state file = "2026-10-02T00:00:00Z123456789Z", want exactly ...`, exit 1 | yes |
| E-008 | AC-014 survivor | `go test -count=1 -overlay <mC> ./internal/harness/` | `ok ... 2.316s`, exit 0 | yes |
| E-009 | AC-012 | `go test -count=1 -timeout 40s -overlay <mD> -run '^TestPruneStamp_StampExistsBeforeTheWork$' ./internal/harness/` (output to a file) | `panic: test timed out after 40s` (line 1), `FAIL ... 40.686s`, exit 1 | yes |
| E-010 | AC-010 | the eight-sentinel `grep -c -F -e ...` | `0` | stdout yes; the exit-code field was not separately observable (the shell tool shows no exit label for a no-match grep; semantics: 1) |
| E-011 | AC-010 | `grep -c -F "events other hooks append in that window are lost" internal/harness/retention.go` | `1` | yes |
| E-012 | AC-003 b pin | `-run '^TestPruneStateUnreplaceableInReadOnlyDirSkips$' -v` with the draft | `--- PASS: ... (0.01s)`, exit 0 | yes |
| E-013 | AC-009 | `git diff --quiet 1e2151a38 -- internal/lockfile internal/harness/observer.go` | empty, exit 0 | yes |
| E-016/17/18 | AC-011 | Windows vet at HEAD; untagged overlay; tagged overlay | exit 0 / exit 1 (`undefined: syscall.Mkfifo`) / exit 0 | yes |
| E-020 | AC-013 green | `-run '^TestPruneStampShorterOverLongerIsExact$' -v` unmutated | `--- PASS`, exit 0 | yes |

Swept-count: every `-v` run printed a `--- PASS` or `--- FAIL` line for its named test; the vacuous form is real and documented (E-019: `ok ... [no tests to run]`, exit 0; read, not re-run). RED reasons match the stated reasons (the AC text names why each is red). Not re-executed: E-014, E-015 (positive controls of the AC-009 forms, read), E-019 (read). Recorded refusal: the first attempt, `cd <worktree> && unset ... && go test ... | sed`, was refused by the worktree guard ("construct too complex to verify"); I re-ran each command as a plain single `go test` invocation with no `cd`, `unset`, pipe or `&&`. The kanban variables were therefore not scrubbed in these runs; the runs read no kanban state (RED outputs identical to the ledger's, which was scrubbed).

### 3. Probes (scratch, outside the tree, darwin uid 501)

- P-FIFO (overlay, `zz_fifostate_test.go`): a FIFO at `<log>.prune-state`, base code, `PruneStaleEntries(30)`: `PROBE-FIFO blocked for 5s (hook would hang)`, `FAIL 5.921s`, exit 1. Cause READ at `retention.go:94` (`stampIsFresh(readStampFile(statePath), now)`; `readStampFile` calls `os.Open` read-only, which blocks on a FIFO with no writer) and `:187-189`.
- P-OWN: symbolic link owned by the test user pointing at `/`: `PROBE-OWN lstat uid: 501 stat uid: 0 euid: 501`.
- P-RO: `PROBE-RO truncate: ... invalid argument`, `writeat: ... bad file descriptor` (the AC-014 seam premise, darwin only).
- P-INODE: `PROBE-INODE same after remove+create: false` on APFS (so the inode-reuse concern in O4 is inferred, not observed).
- P-GIT: `git log --reverse --format=%h -- internal/harness/retention.go` prints 7 historical hashes (`68f023289 fe0901cdc fe211e9c9 987dcc61b 8172295ca 035a7bc31 fe4e8f44e`); `git log --reverse --format=%h 1e2151a38..HEAD -- internal/harness/retention.go` prints nothing; `git merge-base --is-ancestor 6003074c0 68f023289` exit 1 (a stand-in for T, any commit after the base), `git merge-base --is-ancestor 68f023289 6003074c0` exit 0.

### 4. Operator-decision fidelity

(a) No `internal/lockfile` change: REQ-012 (spec.md:166) + AC-011 `git diff --quiet 1e2151a38 -- internal/lockfile`; event loss D3-B plus residual disclosure (REQ-006, REQ-009); F5/F6 disclosure only (REQ-010, REQ-011) and no REQ, AC or milestone repairs them (plan.md:17, :101). Recorded as relayed: spec.md:26, decision-index.md:3, :12, :19, :26, :33, :40, progress.md:11 each say "relayed by the leader session ... not a direct operator answer".
(b) Ownership restriction: REQ-001..004; AC-004 pins the refusal with a foreign entry byte-identical, one `[WARN] harness/retention:` line, via the injected check; AC-005 pins the default check on a file the test made and on `/` (non-root, no root needed; skips under uid 0, stated). Link-owner semantic and Windows "not owned" are in spec.md:113 and plan.md:69-71, not in any REQ and not pinned for the link case: B2. The warning is written by the pruner, so `observer.go` stays byte-identical (AC-009).
(c) One field: counted in REQ/AC/plan text — the owner-check field on `Retention` (plan.md:71, spec.md:176, §B D5 owner-lookup option A); the stamp-write pin uses the extracted locked phase taking the opened file (plan.md:79, AC-014 worded on the outcome). Count is one; consistent with "exactly one".

### 5. Cross-artifact consistency

REQ 15 / AC 14 against the Tier M ceilings 16 / 16; files affected 9 against 5-15; the 0.1.0 HISTORY row ("12 REQ / 15 AC") is historical and the 0.2.0 row ("15 REQ / 14 AC") matches. The dispatch's "M0-M3" does not match plan.md, which has M0-M6 (informational). Flip-milestones in the acceptance table agree with plan.md's milestone contents (AC-001/002 -> M2, AC-007/008 -> M1, AC-010 -> M5, AC-012 -> M4, AC-013 -> M0, AC-014 -> M3). DoD item 6 lists the AC-004/005/006 mutants and omits the AC-003 mutants that plan.md:73 names (cosmetic, O6).

## Defects Found

B1. M0-EXIT-UNSATISFIABLE — plan.md:53 (M0 Exit) and acceptance.md:387 (DoD #1) — Both state: "`git log --reverse --format=%h -- internal/harness/retention.go` lists the commits touching production code, and `git merge-base --is-ancestor T <each>` exits 0 for every one" (plan.md:53: "for every commit F listed by `git log --reverse --format=%h -- internal/harness/retention.go`, `git merge-base --is-ancestor T F` exits 0"). The command carries no range, so it lists seven commits that precede the base (observed, P-GIT above); T, a descendant of the base, can never be their ancestor, so the check exits 1 for each (observed: `git merge-base --is-ancestor 6003074c0 68f023289` exit 1). The criterion is impossible as written, which is the "impossible" direction of `verification-completeness.md` §2; the ordering obligation of §D and `verification-claim-integrity.md` §2.3 is therefore unverifiable by the check the SPEC names. — Severity: critical (MP-9-class: plan M0 order against an unsatisfiable DoD ordering clause) — Class: blocking — Required fix: bound the listing to the card's own commits: `git log --reverse --format=%h 1e2151a38..HEAD -- internal/harness/retention.go` in both plan.md:53 and acceptance.md:387, add the positive control (the same form on a range that contains a change prints a hash; on `1e2151a38..HEAD` before any fix it prints nothing, so the loop body is read only when the listing is non-empty), and state that an empty listing means "no production commit yet", not a pass.

B2. OWNER-PIN-GAP — spec.md:113, plan.md:69-71, acceptance.md:63-77 — The operator decision requires deletion only of entries owned by the current user and the SPEC specifies the symbolic link's own owner (Lstat). No criterion pins that: AC-001 uses a link owned by the user pointing at a user-owned victim, AC-004 injects the answer through the seam, AC-005 tests a regular file and a directory (`/`), where Stat and Lstat agree. A mutant whose owner check reads the target's owner (`os.Stat`) passes AC-001..006: with a link owned by the user pointing at a root-owned target it refuses a healable link (violating REQ-001), and with a foreign-owned link pointing at a user-owned target it deletes a foreign entry (violating the operator restriction). Observed premise: P-OWN, link owner 501, target owner 0. Mutant probe per `verification-completeness.md` §2 fails, so the pin is too shallow to adopt. — Severity: major — Class: blocking — Required fix: add a case (d) to AC-HRH-005 (file `//go:build !windows`, skip under uid 0): a symbolic link the test creates, pointing at `/`, must be reported owned by the production check; name the Stat-based owner lookup as a fourth mutant that must fail, and add it to DoD #6 and plan.md M2 step 6. State the link-owner rule in a REQ (REQ-001 or REQ-003) rather than only in §B and the plan.

B3. FIFO-PREMISE-FALSE — spec.md:113 (D4.a "A directory, a FIFO or any other kind keeps today's behaviour: skip, return the error"), :157 (REQ-HRH-003, "neither a symbolic link nor a regular file ... shall skip the prune, return an error naming the failure"), :63 ("a FIFO as a state path ... a prune stops at the lock step") — Measured on the base tree (P-FIFO): `PruneStaleEntries` with a FIFO at the state path blocks (no return in 5 s), because the lock-free pre-check at `retention.go:94` reads the stamp with a blocking read-only `os.Open` before the lock step is reached. The premise "fails at the lock" is therefore unobserved for the production path and false as measured; REQ-003 cannot hold for a FIFO and AC-003 only covers a directory. A hostile symbolic link to a FIFO hangs the same pre-check before any heal runs, so REQ-001's "shall continue the prune" is also not unconditional. REQ-HRH-008 (spec.md:162) forbids adding a system call to the pre-lock path, so the natural code fix is out of scope by the SPEC's own constraint. — Severity: major — Class: blocking (an unobserved premise stated as fact, falsified by a measurement) — Required fix: correct the premise and scope: drop FIFO and "any other kind" from REQ-003 and from the D4.a "keeps today's behaviour" sentence (keep directory, which is tested), record the FIFO hang in §F and Out of Scope as a pre-existing, unrepaired condition of the lock-free pre-check (cite `retention.go:94`, `:187-189` and the probe result), and say in REQ-001 that its guarantee holds for a link whose target does not block a read-only open. If instead the hang is to be repaired, amend REQ-008 and add an AC; either is acceptable, the current text is not.

O1. AC-009-SHALLOW — acceptance.md:103-109 — REQ-HRH-008 defines the pre-lock path as "the in-memory interval check, the stamp read and the log existence check" (spec.md:162); the stamp read is `readStampFile` / `readStamp` / `stampIsFresh`, which sit outside `PruneStaleEntries`. AC-009's function-scoped `git log -L` on `PruneStaleEntries` and the unmodified `TestPruneStamp_FreshStampNeedsNoLock` (`retention_stamp_contract_test.go:112-149`, which checks for no blocking on a held lock, not for system calls) do not see a `Lstat` or `Open` added inside those helpers. A mutant adding one passes AC-009 and violates REQ-008. Iteration-1 D13 was closed only for the function body. — Severity: minor — Class: optional — Required fix: extend the `-L` check to `readStampFile`, `readStamp` and `stampIsFresh`, or state the reviewer-read of those three functions as the pass condition.

O2. LEDGER-FORM — acceptance.md:166-375 — Ledger commands carry `<S>` placeholders (angle brackets outside quotes, which is outside the single-invocation form of `verification-completeness.md` §2.1 and of the MP-8 form), point at an ephemeral session scratch path, and the committed overlay JSON files (`red-now-drafts/overlay-*.json`) still name that scratch path; G-3 says the hoist would regenerate them. I re-executed every cited command after substituting the path, so the RED cells are decided; the placeholder form is not mechanically re-executable by a later reader. — Severity: minor — Class: optional — Required fix: rewrite the ledger commands against a path under `.moai/reports/t1432/red-now-drafts/` (overlay JSON files need absolute paths; state the worktree root and the generation command, or generate them in M0 and cite the generated names).

O3. REQ-002-SCOPE — spec.md:156 versus plan.md:69 — REQ-002 heals a regular file the pruner "cannot open for reading and writing" (any cause); the plan heals only on a permission error. A transient open failure (`EMFILE`) would otherwise remove a healthy state file; a mutant healing on every open error passes AC-002. — Severity: minor — Class: optional — Required fix: word REQ-002 on "cannot open ... because of a permission error".

O4. AC-002-IDENTITY — acceptance.md:51 — "file identity differs from the 0400 file's" can be falsely red on a file system that reuses a freed inode number at once (inferred; Linux not observed; APFS observed `false`, P-INODE). The SPEC itself treats reuse as a hazard (spec.md:133). — Severity: minor — Class: optional — Required fix: have the test hold the old file open until after the assertion (the inode cannot be reused while open), or assert on the mode and an inode-independent marker.

O5. TAIL-BOUND-OVERSTATED — spec.md:104 — "a stale event in that position is archived by the next interval's prune" holds only if some prune rewrites the log; when that final unterminated line is the only stale event, `prune` returns at `len(stale) == 0` (retention.go:228-232) with no rewrite and no terminator added. In-flight appends complete on their own; only a damaged tail stays. — Severity: minor — Class: optional — Required fix: restate as "classified by the first later prune that sees the line terminated".

O6. DoD-6-INCOMPLETE — acceptance.md:392 — DoD #6 lists the AC-004/005/006 mutation runs only; plan.md:73 also names the AC-003 mutants (heal removing a directory, heal ignoring the removal failure). — Severity: minor — Class: optional.

O7. AC-008-COUNT — acceptance.md:100 — "prints four `--- PASS` lines": `TestPruneKeepsUnparsedLinesVerbatim` has five parallel subtests (retention_unparsed_test.go:21-27, :51-53), so `-v` prints nine `--- PASS` lines (four at column 0). — Severity: minor — Class: optional — Required fix: say "four top-level `--- PASS` lines".

O8. AC-007-VERBATIM — acceptance.md:90-92 — REQ-006 says verbatim and in order; the draft asserts `late-event` by `strings.Contains` and order by first/last line only. A mutant re-encoding a late line passes unless the late bytes differ from the encoder output. — Severity: minor — Class: optional — Required fix: append a late line with non-canonical spacing and compare bytes.

O9. AC-012-MESSAGE — acceptance.md:131 — "the output names the missing archive step" has no pinned phrase; judged by reading. — Severity: minor — Class: optional.

O10. HEAL-HELPER-API — acceptance.md:81-84 — AC-006 interposes a stand-in healer between the helper's inspection and its removal, which needs an inspect/remove split in the helper; plan.md:70 and :72 do not say so. — Severity: minor — Class: optional — Required fix: one sentence in plan.md M2 naming the split.

Optional findings are surfaced and not used to carry the verdict; the three blocking findings are what turn it.

## Regression Check (iteration 2)

- D1 RESOLVED (RED reproduces for the compilable tests, M0 excludes the symbol-dependent ones).
- D2 UNRESOLVED in mechanism (the check is unsatisfiable): B1.
- D3 RESOLVED. D4 RESOLVED. D5 RESOLVED. D7 RESOLVED. D10 RESOLVED.
- D6, D8, D9, D11, D12, D14 (optional in iteration 1): D6 resolved (parameter seam, AC-014 on the outcome); D8 resolved (per-obligation sentinels, reviewer-read stated); D9 resolved (card history moved to decision-index Q2 and progress.md, `grep -c -E 't1[0-9]{3}' retention.go` -> `0`); D11 resolved (decision-index rows are neutral and the relayed verdicts are recorded as relayed); D12 resolved (post-open identity check and the §F residual); D14 resolved (`ci.yml:125`, `release-pr-multi-os.yml:98` cited; READ, not run here). D13 only partly: O1.
- Score trend: iteration 1 0.75, iteration 2 0.87 (no regression, no STOP).

## Recommendation

Fix B1, B2 and B3 (each is a few lines in spec.md, plan.md and acceptance.md) and re-audit iteration 3 scoped to this list plus the full CN-4 verb. The optional findings may be taken or left; O1 and O3 are the cheapest to take.

1. B1: bound the commit listing in plan.md:53 and acceptance.md:387 to `1e2151a38..HEAD`, add a positive control.
2. B2: add the symlink-to-`/` case and the Stat-based mutant to AC-005, DoD #6 and plan.md M2; put the link-owner rule in a REQ.
3. B3: correct REQ-003, D4.a and spec.md:63 on the FIFO; record the hang as an unrepaired pre-existing condition (or amend REQ-008 and add an AC).

## Evidence-bearing sections (per `verification-claim-integrity.md` §3)

### Claim
The plan artifacts are internally consistent and every release-blocking RED reproduces, but three defects block adoption (B1, B2, B3).

### Evidence
Commands and observed outputs are in the sections above (tables in Evidence 2, probes in Evidence 3). The verb runs: `bash <scratch>/trace.sh` -> `COLLECTED: 15 REQ definitions (acceptance input: read)` and nothing else; `bash <scratch>/cn4.sh` -> `COLLECTED: 7 milestones in plan order (M0 M1 M2 M3 M4 M5 M6), 0 exit bindings, 26 ordering candidates`, no `CONFLICT:`, no `GAP:`, no `NONE:`.

### Baseline-attribution
Tree `6003074c07141b2564b93d13ae4b64e1f665e1f6` (code identical to `1e2151a38` and to `db6d88a2a`, the ledger pin); judging build `45600e4ee`, a strict ancestor of HEAD with identical `internal/spec`; measured in this run on darwin, uid 501. No figure is carried over from another run except where labeled READ.

### Gaps
- Linux and Windows behaviour: unobserved (every probe, mutant and re-execution ran on darwin).
- The `.exit` field of E-010: the shell tool does not label a no-match grep exit; stdout `0` reproduced.
- E-014, E-015, E-019: read, not re-executed. Kill observations for AC-004/005/006/014 and the mutants M1/M2 need run-phase code: not obtainable now (the SPEC states this as G-1, G-2).
- The first attempt to run the tests with `cd` + `unset` + a pipe was refused by the worktree guard (verification-claim-integrity §3.1); the re-runs used plain invocations without the kanban `unset`. The D8 per-section awk program was not run; the check was a grep plus a reading.
- No cross-backend audit (no `audit_model` configured); single-auditor opinion.

### Residual-risk
- The hang in B3 exists on the base tree and is not repaired by this SPEC under REQ-008; a hook is killed at 5 s, so the effect is a lingering hook, not a lost event.
- The timing-dependent late-event tests can pass vacuously on a slow runner (ledger G-4); the mutation runs at M1 are the adoption proof.
- Whether a concurrent append can be seen half-written, whether hook standard error reaches a user, and the heal-burst size remain unmeasured, as the SPEC states.

## Operational Notes (unverified)

- Measure whether `git log --reverse --format=%h 1e2151a38..HEAD -- internal/harness/retention.go` is empty before M0 and lists only fix commits after it — command: that line, run before and after commit T. Status: inferred (the empty result at HEAD is observed above; the post-M0 listing is not).
- Measure Linux read-only-handle truncate behaviour before M3 — command: the P-RO probe run on the CI runner image. Status: assumption.
- Measure inode-number reuse on the CI file system — command: the P-INODE probe on ubuntu-latest. Status: assumption.
