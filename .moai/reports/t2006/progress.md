# t2006 progress record (lane-7, Class B, wording-only)

card: t2006 | branch: WT-cache-aware-execution | run: tnsei4 | stop: draft PR, no merge / auto-merge / factory complete

## Intake (2026-11-01)

- Base: tree HEAD was cb30e05b3e (14 commits behind origin/main). `git merge --ff-only origin/main` -> HEAD 5c3f0ebf27584d2c45deb2d00f9d69a4eea932c4; `git merge-base --is-ancestor origin/main HEAD` -> exit 0 (contains).
- Mirror parity before edit: `cmp` live vs `internal/template/templates/` for cache-aware-execution.md (5039 B) and cache-aware-execution-reference.md (7170 B) -> identical.
- Trees: `grep -rl idleCompaction .claude internal .codex` -> 0 files (premise "0 hits in rules/code/templates" re-measured true).
- Read-only reuse: .moai/worktrees/t1755 SPEC-CACHE-IDLE-COMPACT-001 and .moai/reports/t1755/hold-record.md (policy "keep default, set no key" never faulted; audit defects were acceptance-check strength, not facts).

## Upstream re-verification (fetched this run, 2026-11-01)

| Claim | Source | Observed |
|---|---|---|
| 2.1.299 added terminal idle compaction: 200K+ conversation idle ~54 min compacted before 1h cache expires; added `CLAUDE_CODE_IDLE_COMPACT_MIN_TOKENS` and `"idleCompaction": false`, both need v2.1.290+ | code.claude.com/docs/en/changelog (2.1.299 entry, offset 100000) | quoted verbatim |
| 2.1.302 added `/config` "Idle compaction" row | changelog | quoted verbatim |
| 2.1.314 added idle compaction to VS Code and Desktop Code tab; changed to also run for telemetry-off subscribers and on Bedrock, Vertex, Foundry, LLM gateways "once the one-hour prompt cache is on" | changelog | quoted verbatim |
| idleCompaction: default true; 6 conditions (session, autoCompactEnabled, one-hour TTL, feature flags fetchable, size >= 200,000, ~1h idle + awake); `DISABLE_AUTO_COMPACT` turns it off too; key needs v2.1.290+; `/config` row needs v2.1.302+ | settings-reference#idlecompaction (offset 200000) | quoted verbatim |
| Threshold env: default 200000, accepts positive integer, below 100000 raised to 100000 | env-vars row | quoted verbatim |
| TTL: one hour only on subscription within plan usage for main conversation (+ server-controlled helpers); everything else 5 min; `promptCacheTtl` / `CLAUDE_CODE_PROMPT_CACHE_TTL` (v2.1.242+) choose per bucket; gateways must forward `anthropic-beta` | prompt-caching#which-ttl-each-request-gets | quoted verbatim |

## Gaps (not observed)

- Whether a GLM / z.ai gateway session is on the one-hour TTL or runs idle compaction: NOT measured; the wording must assert nothing about it.
- `DISABLE_AUTO_COMPACT` env-vars row itself not fetched (only the settings-reference "Per-session overrides" statement).
- Windows / Codex runtime behaviour: not applicable, not measured.

## Decision records

- D1 (lane judgment): keep the `idleCompaction` default; MoAI sets no key, no code, no config (card text + t1755 policy). Reason: card forbids pinning policy in code/config.
- D2 (lane judgment): reference TTL table rows 19-20 and the "authentication-mode defaults" paragraph contradict the prompt-caching page (TTL is per request bucket and billing, not auth mode); they are corrected in the same edit because directive 9's reference text cites them.

## Checkpoint (2026-11-01, after leader message: operator priority change, resources go to AF)

decision record: decided_by=lane-7 (claude, lane orchestrator) evidence_refs=leader cross-session message (priority change, not an approval cancellation, no kill/reset/delete),PR#1954 ladder_path=gate boundary (leader ruling, auto-semantics.md §6)

- HEAD ae1035f53955e2b5721a452d6248073c71aae4f0 on WT-cache-aware-execution, working tree clean (`git status --short` empty), pushed to origin.
- PR https://github.com/modu-ai/moai-adk/pull/1954: state OPEN, isDraft true, autoMergeRequest null, headRefOid equals HEAD (gh pr view, this run).
- Remaining work: none on the lane side. Merge waits for AF completion, then the same-SHA CI, native Codex receipt and an independent audit, read by the leader. CI was `pending` at the last read; not read to terminal state.
- Work stopped here by the leader's ruling: no further push, no new card. Tree .moai/worktrees/t2006 preserved (branch exists on origin; remote merge not landed). Standing recheck cron already deleted; lease may expire (harmless; re-lease with `factory next --card t2006`).

## Ledger repair (leader resume message: fix own defect locally; no push / ready / new PR)

decision record: decided_by=lane-7 (claude, lane orchestrator) evidence_refs=leader resume message,CI failure ALB-0195/ALB-0197,internal/template/testdata/binding_ledger.json ladder_path=disk-evidence (auto-semantics.md §6 step 1)

- Attempt 1 of 3. Checkpoint committed locally: 9d17e333ca (progress.md; `git add` needed `-f` because `.moai/reports/t2006` is gitignored, file already tracked).
- Reproduced (HEAD 9d17e333ca): `go test -count=1 -run '^TestBindingLedgerIntegrity$' ./internal/template/` -> exit 1; `after-text absent for row ALB-0195 ... (obligation lost)` and `... row ALB-0197 ... (obligation lost)`; `[after-text presence] ... failures=2`.
- Cause (read from ledger + sources): ALB-0195 after_text = the reference-companion intro sentence, which my edit reworded; ALB-0197 after_text = the directives block, whose directive 9 my edit reworded. The ledger's ALB-0197 after_text also still carried an older directive 5 ("Inherit the session model on spawns"), while the base commit 5c3f0ebf27 already has "Per-role model pins split the cache — accepted"; I did not run the test at base, so whether base was red on that line is NOT measured.
- Repair: after_text of both rows set to the deployed text (`internal/template/testdata/binding_ledger.json`, 2 insertions/2 deletions, before_text and all checks untouched; no test weakened). Done by a script over the exact JSON string (`python3 -I`, scratchpad); a first heredoc-form invocation was refused by the worktree guard and re-run as a plain command (Gap: refusal, no substitution of result).
- Re-run, same command -> exit 0, `ok github.com/modu-ai/moai-adk/internal/template 1.353s`. Wider scoped run `-run 'Mirror|Rule|Embed|Manifest|DateProvenance|BindingLedger'` -> exit 0, `ok ... 6.243s`.
- Commit 75c24561b3 (local only, NOT pushed). PR #1954 still head ae1035f539 on origin, draft, unchanged.
- Gaps: CI not re-run on the new commit; `TestCleanReportsArchive` untouched per leader; judging build = go test from this tree's source, no installed moai build involved.

## CI failure vs local repair read-through (leader read-only assignment; no push, no CI/test rerun)

decision record: decided_by=lane-7 (claude, lane orchestrator) evidence_refs=gh api jobs 124780129207/124780129202/124780129389 logs,git show origin/main:internal/cli/clean_reports_archive_test.go,git show 72c02a4d9d,git show 5c3f0ebf27 ladder_path=disk-evidence (auto-semantics.md §6 step 1)

Claim 1 - the three failed jobs at head ae1035f539 fail on exactly two things.
- `gh api repos/modu-ai/moai-adk/actions/jobs/<id>` -> all three `conclusion: failure`, `head_sha: ae1035f53955e2b5721a452d6248073c71aae4f0`: 124780129207 = `Test (ubuntu-latest)`, 124780129202 = `Race Test (internal/cli 1 of 4)`, 124780129389 = `Race Test (remainder packages)`.
- Log greps (`--- FAIL|obligation lost|failures=`): 124780129207 -> `TestCleanReportsArchive_MovesOnlyEligibleCandidates` (`clean_reports_archive_test.go:121: no new YYYY-MM shard created under archive/`) AND `TestBindingLedgerIntegrity` (`failures=2`; `after-text absent for row ALB-0195 ... (obligation lost)`, same for ALB-0197). 124780129202 -> only TestCleanReportsArchive. 124780129389 -> only TestBindingLedgerIntegrity (same two rows).
- Not read: the other CI jobs and their logs.

Claim 2 - TestCleanReportsArchive is a date-dependent fixture defect outside this PR's diff.
- `git show origin/main:internal/cli/clean_reports_archive_test.go` (origin/main = 5c3f0ebf27): fixture writes `archive/2026-08/t-older/evidence.md` and the assertion skips the shard named `2026-08`; the candidate is backdated `DefaultReportsArchiveRetentionDays+1` days (`internal/config/defaults.go:522` = 90, so 91).
- Arithmetic (not executed): run date 2026-11-01 minus 91 days = 2026-08-02, so the new shard is also `2026-08` and the "new shard" search finds none. This matches the log line and the commit message of 72c02a4d9d.
- `git show 72c02a4d9d`: renames the fixture shard to `2000-01` in 5 places of `internal/cli/clean_reports_archive_test.go` only (5+/5-); assertions unchanged. Read only, not applied.
- `git branch -a --contains 72c02a4d9d` -> `WT-agentic-factory-core-claim-runner`, `WT-agentic-factory-integration` only; `git branch -r --contains` empty, so it is not on origin/main (local remote-tracking state; no fetch run this pass).
- This PR's diff touches no `internal/cli` file (4 rule files, ledger, report files), so the failure is not caused by it. Whether origin/main CI itself is red was not read.

Claim 3 - local 75c24561b3 covers the TestBindingLedgerIntegrity failure; nothing from the CI logs remains for that test.
- CI names only ALB-0195 and ALB-0197 (`failures=2` of 363 rows); local repair changes exactly those two rows' after_text (earlier entry: 2+/2-); the earlier run of the CI command gave exit 0 and `ok ... 1.353s`, and the wider scoped run exit 0 `ok ... 6.243s`. Not rerun in this pass (forbidden), so these are the earlier observations. CI on the new SHA does not exist yet.
- Remaining CI failure after the repair: TestCleanReportsArchive only (Claim 2), which this PR does not own.

Claim 4 - the earlier observation "the ledger's ALB-0197 directive 5 was already older than base" is REFUTED; correction of the Ledger repair section above.
- `git diff --stat 5c3f0ebf27 ae1035f539 -- internal/template/testdata/binding_ledger.json` -> empty (the PR did not touch the ledger before the repair).
- Script over `git show 5c3f0ebf27:...binding_ledger.json` and `...cache-aware-execution.md`: `ALB-0197 base after_text == base rule block: True`; the base ledger text contains `Per-role model pins` and not `Inherit the session model`.
- So at base the ledger and rule agreed on directive 5; the two failing rows were caused by this PR's directive 9 and intro-sentence rewording alone. Base test state itself still not run (the text equality is the evidence, not a test result).

Pre-push conditions (list for the leader; none executed):
1. Push approval is the leader's / operator's call (keep-set); local HEAD is 3 ahead of origin (9d17e333ca, 75c24561b3, 9dad956c08, plus this record commit).
2. Before the push: `git fetch origin main`, then the divergence count against origin/main (pre-spawn/pre-edit sync pattern); re-read `git rev-parse --short HEAD` and the branch.
3. TestCleanReportsArchive will still fail on the new SHA unless the 72c02a4d9d fix reaches origin/main and this branch absorbs it (or the leader rules the red as a known main defect). The lane does not touch it.
4. After the push: read same-SHA CI to terminal state, the native Codex review receipt, and an independent audit before any ready/merge (leader reads).
5. PR stays draft, no auto-merge, until AF completes (operator priority ruling).

Gaps: logs of the other CI jobs not read; no fetch, no test, no CI rerun in this pass; origin/main CI state not read; the date arithmetic is hand-computed, not run.
Residual risk: the repaired ledger text and the rule file must stay byte-equal to each other; any later wording edit to directives 1-10 or the reference intro sentence re-opens ALB-0195/0197.
