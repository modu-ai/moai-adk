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
