# research.md — SPEC-CC2219-UPSTREAM-ALIGN-001 (thin pointer)

The research artifact for this SPEC is the external harness report — the SSOT, not duplicated here:

- **`.moai/research/cc-update-2.1.207-to-2.1.219.md`** (2026-07-25, `hns-release-update-specialist`) — §2 version inventory, §3 tier table, §4 per-cluster detail (GD-1..GD-9), §5 verified NO-OPs, §6 verification statement + gaps, §7 decomposition/recommendation, §9 consolidated file inventory.

Supplementary evidence not in the report:

- **2026-07-25 nesting probe** (post-report, same day): `sync-auditor` with `CLAUDE_CODE_MAX_SUBAGENT_SPAWN_DEPTH` verifiably unset spawned an `Explore` child returning "NESTING-OK" on the installed 2.1.219 binary — resolves report §6 Gaps 1-2 (nesting default-ON; changelog authoritative, doc lags). Caveats: single trial, depth-1 only, env propagation not separately verified. Encoded in spec.md §A.1.
- **PR #1146** (commit 714270085, merged to main): GD-3 fork-matcher hotfix — basis of the GD-3 Out of Scope exclusion.

Note: the report is a gitignored dev-only artifact; this pointer file is the in-SPEC record of its identity and role.
