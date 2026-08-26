# Audit → Commit-Window Correlation Recipe

> SPEC-TEAMMATE-REVIVAL-GUARD-001 · run-phase M3 deliverable, placed at `.moai/docs/`
> at sync (card t267). Direction-3 (visibility) deliverable of the mechanism: the
> audit rows exist so a revival window is provable after the fact.

## Purpose

When a stopped teammate is found to have run again (the revival incident shape), bound
the revival window from the stop-guard audit trail and attribute the commits inside it
post hoc.

## Steps

1. **Bound the window.**
   `grep '"name":"<teammate>"' .moai/logs/agent-stop-audit.jsonl`
   Window = `stop_recorded` (row N) → the earliest of `respawn_cleared`,
   `session_cleared`, or the next `stop_recorded` for that name. Each row carries a
   UTC RFC3339 timestamp, session_id, name, agent_id, and decision — enough to key the
   window to one session.
2. **Attribute commits inside the window.**
   `git log --since='<window start UTC>' --until='<window end UTC>' --format='%h %cI %an%n%b'`
   Match author/committer identity to the teammate where possible. Where the revived
   agent committed under a shared identity, the audit row's session_id plus the window
   boundaries are the attribution evidence.
3. **Classify.** Commits inside a stop→window-end span with no intervening
   `respawn_cleared` are revival-window output — re-verify them independently. The
   originating incident precedent: the revived output itself passed re-verification;
   the defect was ownership (an unowned writer), not artifact quality. The risk the
   next revival carries is that it passes WITHOUT re-verification.
4. **When the deny layer is enabled**, step 2 usually returns an empty result for the
   window — the send was refused. An empty result is the guard working; the
   `send_denied` row is the record that the attempt was made and when.

## Measured worked example (this card's worktree, live rows)

Session `703df7e1-d40e-420d-8f08-0b4f795068dd` stopped `t267-ep1-probe2` at
`2026-08-26T17:12:48Z`; two sends were denied (`17:12:53Z` lead-issued,
`17:13:10Z` teammate-issued — the incident vector); a same-name spawn cleared the
entry at `17:13:21Z`.

```
git log --since='2026-08-26T17:12:48Z' --until='2026-08-26T17:13:21Z' --format='%h %cI %an%n%b'
→ 0 commits
```

No revival contamination; the deny held for the whole window. (The run's own commits
`f7bd5bdc7` / `70541af5c` sit at 02:04 / 02:07 +09:00, before the window opened.)

## Operational note

A registry file whose session ended before the SessionEnd cleanup existed (or on a
binary predating it) is inert: per-session scoping means no other session's sends
consult it. Such files are manual-cleanup garbage, not a correctness hazard.
