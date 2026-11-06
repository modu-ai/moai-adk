# t2135 — classification of the "other" Windows failures (report only, no product change)

## Claim
Of the 1335 categorized Windows-race failure rows of PR #1969, 201 are `OTHER` and 117 are `NOMSG`. The 117 NOMSG rows are parent rows of failed subtests (110 of them per the first audit: SQLITE 40, OTHER 33, DIRSYNC 18, ...), so 201 + 117 = 318 is NOT a count of independent failures. The 201 `OTHER` rows are accounted for below by a DISJOINT set of rules (sum = 201). Attributions are from failure messages, test names and code reading; nothing was reproduced (no Windows host), so every proposed card starts with a reproduction step.

## Evidence
Source table: `categorized.tsv` of the PR #1969 CI analysis (columns: shard, category, package, test, first failure line; 1335 rows; machine-local scratchpad of another session).
Category counts: SQLITE_URI 474, DIRSYNC_ACCESS_DENIED 326, OTHER 201, NOMSG 117, CLI_WORKTREE 60, HOMEGUARD_SEAM_REALHOME 54, POSIX_MODE_PERM 46, OWNERSHIP_EPOCH 26, BIN_SH_MISSING 19, PRECOMMIT_E2E 12 (`cut -f2 | sort | uniq -c`).
OTHER by package: internal/cli 134, factory 12, cli/update/skillresidue 7, session 5, homestate 5, graph 5, template 4, statusline 4, agenticfactory/claim 4, auditreceipt 3, astgrep 2, mergeq 2, 14 single-row packages (`awk '$2=="OTHER"{print $3}' | sort | uniq -c`; the single-row count is 14, not 8).
Disjoint accounting of the 201 OTHER rows (first matching rule wins, in this order; `awk` over the table, run in this session):

| # | Group | Rows | Rule |
|---|---|---|---|
| G2 | Stop-goal-wait hook | 15 | test name `TestStopGoalWait*`, `TestStopParseCap*`, `TestFactoryWaitResolveCloses*` |
| G3 | Empty-target guidance fixtures | 8 | test name `TestEmptyTarget*` |
| G4 | Todo store | 12 | `TestTodoStaleStoreDisclosure` 6 + `TestTodoPR_QueueDirUnchanged` 6 (the 6 `backlog.lock` message rows are inside these, not extra) |
| G5 | Skill mirror / residue | 10 | package `skillresidue`, or message about `skill.md`, the skill folder, the skills mirror |
| G6 | Environment/fixture | 5 | `installed version`, `state file unreadable` (3 rows), `TestCheckMCPProviderDuplicat*` |
| G7 | Path spelling (Windows path in message) | 61 | message carries a `C:` path, `RUNNER~1`, spelling/case-fold/raw-path text, OR a backslash in the message. Read without the backslash clause the rule gives G7 54 / G9 97; the clause also pulls in 4 rows that are not path spelling (an `\n` escape 1, awk escape warnings 2, CRLF line ends 1). Both are probed together by card 1, so the 7-row difference does not change the plan |
| G9 | Unassigned singles | 90 | everything else |
| | Total | 201 | |

G1 (SQLite URI / dirsync derivatives) is not an OTHER group: it lives in the NOMSG parent rows (agenticfactory/ledger 37, e2e 8, runner 7, claim 3 = 55 of the 117) and in the SQLITE_URI / DIRSYNC categories themselves. Of the other 62 NOMSG rows, most are parents of failures in other groups, but NOT all (second audit, R2-1): 6 are independent failures with no subtests — `internal/cli` `TestHookSink*` / `TestHookPathWarnRecordReachesSink` (5 rows) and `internal/homestate` `TestOpenFactoryReadonlyNeverCheckpointsWAL` (1 row) — and 4 more are parents of the G1 category (SQLite 3, dirsync 1). The 6 independent rows are proposed as card 9 below (no group covers them).

## Attribution, corrected by the audit (code reading, not reproduced)
- G6 is NOT "needs the real Claude CLI, so skip": `TestInstalledVersionFromResolvedPath` builds a symlink fixture under `t.TempDir()` and replaces the PATH lookup by a seam; the likely cause is `versionSegmentFromPath` looking for a slash-fixed `/versions/` prefix (product, path separator). `TestCheckMCPProviderDuplicates_*` writes `C:\...` paths into a JSON string without escaping, which breaks the JSON and yields "unreadable" (test-fixture defect, escape the path). A skip card would hide two fixable defects.
- G3's "refusal message is empty" was a misreading of the table: the test prints `got:\n%s`, so the first line always ends at `got:` and the real refusal text is on the second line, which the table does not carry. The likely cause is a product path built with `filepath.Join(root, filepath.FromSlash(...))`, which yields backslashes on Windows while the test expects slashes; that also explains why only the "unreadable" rows fail. Needs reproduction.
- G4: the `StaleStore` rows look like path spelling rather than locking (audit F7); `backlog.lock` is a stat/attribute path issue (6 rows, not 3).
- G8 ("timing") is dropped: its rows are path-absoluteness (stat calls, 3 rows) and "worktree dirty" issues, folded into G7/G9.

## Proposed follow-up cards (counts are the disjoint ones above)
1. G7 path-spelling: first a PROBE card to attribute the 61 rows to test vs product by surface (cwd/worktree identity, graph edge paths, session, update moves), then fix cards per surface. Class B, reproduce-first.
2. G9 90 singles: classified inside the same probe (they are single-test packages and singles in cli).
3. G2 stop-goal-wait hook (15): reproduce first (hook subprocess/DB/path), one package, class B.
4. G3 empty-target fixtures (8): confirm the `filepath.FromSlash` hypothesis, then fix the expectation or the product spelling.
5. G4 todo store (12): path spelling + stat on `backlog.lock`; one card. Its scope also takes the 2 `TestDoctorTodoStoreDivergence` rows that the disjoint rules leave in G9 (R2-3).
6. G5 skill mirror / residue (10): rename of directories in use and separator spelling; one card.
7. G6 (5): two defects (a product `versionSegmentFromPath` separator; a test JSON-escape) — NOT a skip card.
8. G1 + OWNERSHIP_EPOCH 26: no card now; re-measure after t2118, t2119 and t2123 land.
9. Six independent NOMSG failures with no subtests (R2-1): `internal/cli` `TestHookSink*` / `TestHookPathWarnRecordReachesSink` (5) and `internal/homestate` `TestOpenFactoryReadonlyNeverCheckpointsWAL` (1): class B, reproduce-first, first step reads the CI test stream for their messages (the table has none).
Rows assigned elsewhere and not part of this report: CLI_WORKTREE 60 (t2085), HOMEGUARD 54 (t2133), PRECOMMIT_E2E 12 (t2134), BIN_SH 19 (t2123), POSIX_MODE_PERM 46 (t2130/t2131/t2132), SQLITE_URI 474 (t2118), DIRSYNC 326 (t2119). The card ids were checked against `moai todo` by the audit.

## Baseline-attribution
Counts come from `cut | sort | uniq -c` and `awk` over the table, run in this session against the 1335-row file; the table is a CI artifact of the PR #1969 run, not of this tree. The CI head SHA is recorded in the t2077-tree analysis `.moai/reports/pr-sweep-20261106/pr1969-ci-analysis.md` (the audit could not confirm that t2077 is the right owner of that document; the SHA itself should be taken from `gh pr view 1969 --json headRefOid`).

## Gaps
- Nothing was reproduced; no Windows host. G2-G9 attributions are inferences from messages, test names and the audit's code reading.
- The table keeps only the first failure line per test: a test can fail again after its first failure is fixed.
- 90 rows are "unassigned singles"; they are classified only as a pool.
- The G7 rule as written in words does not reproduce 61 / 90 without the backslash clause (see the G7 row); the counts are reproducible only with that clause.
- The CI head SHA is not recorded here; take it from `gh pr view 1969 --json headRefOid` (R2-4).
- Whether cards that landed on main after the CI run (t2063-t2091) already fixed some rows was not checked.
- Codex was not approved for this card: Claude audit only.

## Residual-risk
Proposed cards may overlap with t2085 and t2092 at the edges. G1 and OWNERSHIP_EPOCH may vanish after t2118/t2119; opening cards for them before a re-measure would waste a slot. The G7 probe is the largest unknown (61 + 90 rows).

## Audit round 1 (FAIL 0.77 @665a7c5673) and its repair
F1 (G6 attribution contradicted by code) corrected above; F2 (G3 table misread) corrected above; F3 counts corrected (backlog.lock 6, state-file-unreadable 3, single-row packages 14, G4 12 not ~15); F4 group arithmetic replaced by the disjoint accounting that sums to 201 and the NOMSG explanation; F5 G8 dropped; F7 noted under G4; F8 head-SHA source caveat in Baseline-attribution. F6 (POSIX exec bits 3, shell digest 2) is inside G9 and not separately classified.
