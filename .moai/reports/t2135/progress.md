# t2135 — classification of the "other" Windows failures (report only, no product change)

## Claim
Of the 1335 categorized Windows-race failure rows of PR #1969, 201 are `OTHER` and 117 carry no message (`NOMSG`). They are not one defect: ~55 of the NOMSG rows are probable derivatives of the SQLite-URI / directory-fsync groups that sibling cards already cover; the 201 `OTHER` rows split into about eight root-cause groups of which two are repeated enough to be bounded cards (stop-goal-wait hook, empty-target guidance fixtures) and one (Windows path spelling) is a sweep that needs per-site attribution. Every attribution below is an INFERENCE from the failure message and test names; nothing was reproduced (no Windows host), so each proposed card starts with a reproduction step.

## Evidence
Source table: `categorized.tsv` of the PR #1969 CI analysis (`/private/tmp/claude-501/-Users-goos-MoAI-moai-adk-go/43472e68-3604-44b1-a008-c14cce281ea1/scratchpad/categorized.tsv`, 1335 rows; columns: shard, category, package, test, first failure line).
Category counts (`cut -f2 | sort | uniq -c`): SQLITE_URI 474, DIRSYNC_ACCESS_DENIED 326, OTHER 201, NOMSG 117, CLI_WORKTREE 60, HOMEGUARD_SEAM_REALHOME 54, POSIX_MODE_PERM 46, OWNERSHIP_EPOCH 26, BIN_SH_MISSING 19, PRECOMMIT_E2E 12.
OTHER by package (`awk '$2=="OTHER"{print $3}' | uniq -c`): internal/cli 134, internal/factory 12, cli/update/skillresidue 7, session 5, homestate 5, graph 5, template 4, statusline 4, agenticfactory/claim 4, auditreceipt 3, astgrep 2, mergeq 2, and 8 single-test packages.
NOMSG by package: internal/cli 38, agenticfactory/ledger 37, agenticfactory/e2e 8, cli/worktree 7, agenticfactory/runner 7, agenticfactory/claim 3, template 2, homestate 2, others.
OTHER message signatures (normalized, top): `want a block decision, got stdout=""` 10; `want success containing "re-measure recorded"` 4; `target worktree removed: worktree <path>` 3; `stat backlog.lock: GetFileAttributesEx <path>` 3; `no dispatch operation recorded (... mission supervisor: governance)` 3; the four `...refusal must contain ".moai/config/sections/git-strategy.yaml", got:` families 8 (message empty); skill-residue `skill.md = "" ... want the user's bytes back` 2 plus `move the skill folder aside: rename <path>` 2; `installed version = "unknown"` 2; `Claude Code state file unreadable; not checked` 2; remaining ~120 are single signatures (path spelling `C:\...` vs `RUNNER~1`, `.agents\skills` vs `/`, case folding, `stored CWD ... differs from the caller spelling`, `underPath does not fold case on a folding volume`, `t16's edge file ... want the raw path "...\\..."`).

## Groups (root cause is inferred; counts are rows of the table)

| # | Group | Rows | Attribution (inferred) | Proposed card |
|---|---|---|---|---|
| G1 | SQLite URI / dirsync derivatives sitting in NOMSG: agenticfactory/ledger 37, e2e 8, runner 7, claim 3 | ~55 | product: cards t2118 (URI) and t2119 (dirsync) | none now — re-measure after t2118 + t2119 + t2123 land |
| G2 | Stop-goal-wait hook: `TestStopGoalWait_*`, `TestStopParseCap_*`, `TestFactoryWaitResolveClosesOperatorWaitAndStopGoalBlocksAgain`: hook returns `stdout=""` where a block decision is expected | ~17 | product or fixture (hook subprocess/path or DB access on Windows); cause unknown | reproduce-first card, class B, one package |
| G3 | Empty-target guidance (`TestEmptyTarget*NamesTheFix/{no,unparseable}_git-strategy.yaml`): refusal text empty | 8 | test fixture (the two rows that build a missing / unparseable `git-strategy.yaml`); cause unknown | small bounded card |
| G4 | Todo store on Windows: `TestTodoStaleStoreDisclosure` 6, `TestTodoPR_QueueDirUnchanged` 6, `stat backlog.lock: GetFileAttributesEx` 3 | ~15 | product-or-test (file attribute / lock path handling on Windows) | one card, reproduce-first |
| G5 | Skill residue and mirror: `cli/update/skillresidue` 7, `skill.md = ""` 2, `move the skill folder aside: rename` 2, `.agents\skills mirror absent` | ~12 | test: rename/replace of directories in use and separator spelling | one card |
| G6 | Environment tests that need the real Claude CLI/state: `installed version = "unknown"`, `Claude Code state file unreadable`, `TestCheckMCPProviderDuplicat*` | ~7 | test: legitimate skip when the tool is absent | quick skip card (like t2123/t2130) |
| G7 | Path spelling: short names (`RUNNER~1` vs long), `\` vs `/`, case folding, `stored CWD ... differs from the caller spelling`, `tree-scope cwd`, `thread/start cwd`, `worktree = <path>`, `t16's edge file ... raw path` | ~40-60 | mixed product/test; the largest family, needs per-site attribution | a sweep split by surface (cwd/worktree identity; graph edge paths; session) — first a probe card to attribute each site |
| G8 | Timing/behaviour: `want exactly 1 winner, got 0`, `recorded stat calls = []`, `status="armed" turns=0`, `re-measure recorded` 4 | ~10 | undetermined; possibly flaky or subprocess related | classify during reproduction |
| G9 | Unclassified singles | ~40 | n/a | roll into G7's probe |

Rows already assigned elsewhere and NOT part of this report: CLI_WORKTREE 60 (t2085), HOMEGUARD 54 (t2133), PRECOMMIT_E2E 12 (t2134), BIN_SH 19 (t2123), POSIX_MODE_PERM 46 (t2130/2131/2132), SQLITE_URI 474 (t2118), DIRSYNC 326 (t2119), OWNERSHIP_EPOCH 26 (re-check after t2119).

## Baseline-attribution
Counts come from this run of `cut | sort | uniq -c` and `awk` over the table above (1335 rows); the table itself is a CI artifact of one run (head SHA recorded in the t2077 analysis), not of this tree.

## Gaps
- Nothing was reproduced; no Windows host. Group causes are inferred from messages and test names only; group sizes for G2-G9 use signature and test-name prefixes, so they are approximate (±10 rows each).
- `categorized.tsv` lives in another session's scratchpad (machine-local); it is not a durable citation target. The per-group counts above are carried here.
- Rows past the first failure line per test are invisible: a test can fail again after its first failure is fixed.
- No attempt was made to check whether cards already landed on main (t2063-t2091) fixed some of these rows.

## Residual-risk
Proposed cards may overlap with t2085 (cli/worktree) and t2092 (umbrella analysis) at the edges; the leader's issue triage should de-duplicate. G1 and the OWNERSHIP_EPOCH rows may vanish after t2118/t2119; opening cards for them before a re-measure would waste a slot.
