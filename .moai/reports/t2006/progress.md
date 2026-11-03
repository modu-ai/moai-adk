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

### Handoff record (PR #1954; supersedes the "Pre-push conditions" list above, which predates the push)

decision record: decided_by=lane-7 (claude, lane orchestrator) evidence_refs=leader message (hand-off record, no code change),this section ladder_path=gate boundary (leader ruling, auto-semantics.md §6)

- Code head before this record: `e739a8bd36d9d8d1e4f84c3c6cfb4ee0df9822b1`; the branch is already pushed (the earlier "3 ahead of origin" no longer holds). This record is the only change in its commit (no code change), pushed once without force.
- State: PR #1954 draft, no hold. CI at e739a8bd36 (run 41332759419): the only failing jobs are Test (ubuntu-latest) and Race Test (internal/cli 1 of 4), both on `TestCleanReportsArchive_MovesOnlyEligibleCandidates` (the #1956 calendar defect, non-required per the leader). The card's own failure, `TestBindingLedgerIntegrity`, is fixed and passes in CI. Required-check rows (Lint, Build, Analyze) were not individually re-read in this record.
- origin/main was not absorbed into this branch; no independent audit exists for this PR (not requested by the leader at this point).
- Open items: the Gaps and Residual risk lists above; the ledger text and the rule file must stay byte-equal.
- Ready-for-review and merge are the leader's.

### Post-handoff record (R156/R158/R180/R182)

decision record: decided_by=lane-7 (claude, lane orchestrator) evidence_refs=leader messages R158,R180,R182; run 41364659599 failed-job logs ladder_path=gate boundary (leader ruling, auto-semantics.md §6)

- R156 rerun (`gh run rerun 41364659599 --failed`, once): the same two jobs failed again on `TestCleanReportsArchive_MovesOnlyEligibleCandidates`; a rerun reuses the original merge commit, which lacks the #1956 fix.
- R158 absorb: `git merge --no-ff origin/main` (no conflicts, no rebase), scoped `go test ./internal/cli -run TestCleanReportsArchive` passed, pushed once without force: `9f737ae253` -> `0249200318`.
- R180 finding (source read only, not reproduced in CI): `Capture` abandons its worker after the 50ms time box (`internal/config/defaults.go` DefaultBugreportCaptureTimeBox), and `TestCaptureSpoolsMoaiAndAmbiguous` read the spool count immediately after `Capture`, so a slow `-race` runner sees 0 lines. An ordering race, not a data race.
- R182 repair (main-side test hardening, outside this PR's scope; no new card): RED `e56fe1eb7c` (stub write held until after `Capture` returns; 5/5 runs fail "spooled 0 line(s), want 1"), GREEN `bb763bfae6` (channel wait with 2s Fatal bound, `t.Cleanup` restore; 20/20 pass, `go test ./internal/bugreport -race -count=1` ok). Production code untouched; no sleep, skip or registry.
- Gaps: the CI remainder-packages race leg at the pushed head is not yet read; the HOME/TMP-dependence hypothesis was not tested.

### Sync-phase record

decision record: decided_by=manager-docs (sync phase, spawned by lane-7) evidence_refs=commands below ladder_path=disk-evidence (auto-semantics.md §6 step 1). Class B, no SPEC directory: no spec.md/plan.md/acceptance.md exist, so no B12 AC-count self-test applies (ac_source=none); test 1 (pre-emission grep) and test 3 (path existence) were run.

**Docs checked**

| Path | Result |
|---|---|
| `docs-site/content/en/claude-code/context-memory/prompt-caching.md` | Read via grep. Describes the 5-minute idle lifetime and an "Auth method / Default TTL" table (subscription 1 hour, API key and third party 5 minutes, `promptCacheTtl` noted). Does not mention directive 9, session length or idle compaction. Not contradicted by the new directive 9 wording. Unchanged. |
| `docs-site/content/en/cost-optimization/prompt-caching.md` | Same table and `promptCacheTtl` paragraph (lines 142-147); no idle compaction mention. Not contradicted. Unchanged. |
| `docs-site/content/en/advanced/tokenomics-overview.md` | Line 67/89: 5-minute idle lifetime and stagger-spawn/deferred-edit summary; no directive 9 claim. Not contradicted. Unchanged. |
| `docs-site/content/ko/.../prompt-caching.md`, `docs-site/content/ko/cost-optimization/prompt-caching.md` | Matched by the cache-aware grep; not opened, mirrors of the en pages above. ja and zh: no match. Not verified line by line (Gap). |
| `README.md`, `README.ko.md`, `README.ja.md`, `README.zh.md`, `docs/` | `grep -rlE "cache-aware-execution\|Cache-Aware Execution\|idleCompaction\|idle compaction"` returned only the en prompt-caching page; no README or docs/ hit. |
| `CHANGELOG.md` [Unreleased] | No cache-aware directive 9 entry before this pass (`grep -c 't2006' CHANGELOG.md` -> `0`). |

**Docs changed**

- `CHANGELOG.md`: one entry added under `[Unreleased]` / `### Changed` (card t2006, Class B, no SPEC). Convention check: [Unreleased] and 3.2.0 carry Class B no-SPEC entries (e.g. lines 101-107 of the pre-edit file: `card t1873, Class B, no SPEC` ... `card t1854, Class B defect fix, no SPEC`), so an entry is expected.
- docs-site pages: none changed. Reason: no contradiction found; a 4-locale docs-site edit is outside a wording-only rule card.

**Audit-ready signal (E.4 style)**

```yaml
sync_complete_at: 2026-11-03
sync_commit_sha: pending-backfill
sync_status: complete-pending-audit
gap_named: ko docs-site pages not read line by line; no independent sync audit run (forbidden in this pass)
changelog_entry_position: "[Unreleased] / ### Changed, appended after the SPEC-CI-CODEX-REVIEW-001 entry"
frontmatter_status_transitions: none (no SPEC)
b12_self_test_a: pass (grep -c t2006 CHANGELOG.md -> 0 before edit)
b12_self_test_b: not applicable (no acceptance.md)
b12_self_test_c: pass (ls of every path named in the entry succeeded)
```

`sync_status` is not `complete` because the audit step was out of scope for this pass and the ko pages were not read; every other listed check was performed.

**Claim / Evidence / Baseline-attribution**

| Claim | Evidence (command -> observed) | Baseline |
|---|---|---|
| Sync ran at the stated HEAD | `git rev-parse --short HEAD` -> `83e4f0c3e8` | this tree, this run |
| No t2006 CHANGELOG entry existed | `grep -c 't2006' CHANGELOG.md` -> `0` | this tree, before the edit |
| Live and template rule files are byte-identical | `cmp .claude/rules/moai/workflow/cache-aware-execution.md internal/template/templates/.claude/rules/moai/workflow/cache-aware-execution.md` -> no output; same for `cache-aware-execution-reference.md` -> no output | this tree |
| The card's own change set is 4 rule files, the ledger and 3 report files | `git diff --name-only 5c3f0ebf27 e739a8bd36` -> the 2 live rule files, 2 template mirrors, `internal/template/testdata/binding_ledger.json`, `.moai/reports/t2006/{card-review,progress,verdict}.md` | base 5c3f0ebf27 to code head e739a8bd36 |
| Entry paths exist | `ls` of the 2 live rule files, 2 mirrors, `binding_ledger.json`, `progress.md` -> all listed, exit 0 | this tree |
| Directive 9 wording the docs were compared against | `grep` of `cache-aware-execution.md` line 23: "5 minutes by default; one hour for a subscription's main conversation within plan usage, or with `promptCacheTtl` ... Claude Code may compact an idle 200K+ conversation at about 54 minutes (`idleCompaction`, v2.1.290+); MoAI keeps that default and ships no key" | this tree |

**Gaps (not observed)**

- Korean docs-site pages not read line by line; ja/zh have no cache-aware match.
- No independent sync audit and no `go test` run in this pass (scoped tests not needed for a CHANGELOG and report edit; nothing in the CHANGELOG is parsed by the touched Go tests, which was not verified).
- The en docs-site TTL table keeps the "Auth method" framing that D2 corrected in the rule reference; read as imprecise-but-not-contradicting, not as a verified error.
- The idle-compaction facts in the CHANGELOG entry are the quoted upstream values recorded under "Upstream re-verification" above, not re-fetched in this pass.
- The bugreport test hardening commits (`e56fe1eb7c`, `bb763bfae6`) and the main merge `0249200318` are not described in the CHANGELOG entry (out of the card's wording scope).

**Residual-risk**

- The docs-site prompt-caching pages say nothing about idle compaction; a reader of those pages learns only the 5-minute baseline. Left for a separate docs card (4-locale same-PR obligation).
- The binding ledger rows ALB-0195/0197 and the rule text must stay byte-equal; a later directive 1-10 edit reopens them.
