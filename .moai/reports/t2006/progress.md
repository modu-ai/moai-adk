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
