# t1476 card-review (advisory)

- Tool: `mcp__moai__codex_review` scope=card, project_root = this worktree, base `2b9e4a4d0`, reviewed HEAD `93cd34131`, backend codex.
- Verdict: **fail** (1 finding, P2).

## Finding P2 — design.md §G addendum overstated the prune bound

Codex: `PruneObservationLogs` ages out the audit file whole, by mtime; every append refreshes the mtime, so on an actively used project the file never ages out (its isolated repro: 60-day rows survived a 30-day cutoff, file 146→292 bytes). The addendum's "live growth bound" claim was wrong.

Confirmed by reading: `internal/hook/prune_logs.go:169-171` — `if name == agentModelAuditFileName { ... if info.ModTime().Before(cutoff) {`.

Disposition: **fixed once** — design.md §G addendum, the REQ-AMI-009 v0.9.0 sub-bullet, and the Amendments scope row 3 now state that the age-out removes only an idle file and that growth under continuous use is an open residual. Codex's repro itself was not re-run here (Gap).

No re-review run (single fix round per dispatch).
