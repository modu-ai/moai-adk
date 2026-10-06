# Card t1544 — card-review (run→sync stage, advisory)

Reviewer: codex (mcp__moai__codex_review, scope=card, project_root=<this
worktree>) + lane-side disposition of every finding.
Reviewed tree: `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1544`
(branch `WT-audit-receipt-guard`).
Review verdict as returned: **fail** — disposition below; the one in-diff
finding was repaired and re-verified in this tree; the remaining findings are
outside this card's diff (base mis-resolution) and are handed to the leader.

## Scope note — the review base resolved to develop, not main

The resolver computed merge base `a158b4b5f` (the develop tip; this repo cut
over to GitHub Flow with main as base and develop left behind), so the
"card diff" the reviewer saw spans every main-side commit between
`a158b4b5f` and this branch's point `10df085da` — dozens of files this card
never touches. Evidence: `git diff --name-only a158b4b5f -- internal/` lists
~30 files across `internal/cli` (factory_*, todo*, memory_fold*, graph,
codex_review_*); this card's actual diff is exactly 3 files:

```
internal/auditreceipt/store.go
internal/hook/audit_receipt_guard.go
internal/hook/audit_receipt_guard_test.go
```

A control run of the same tool with `project_root` = the parent checkout
returned `inconclusive` ("this tree is not one: no card branch: main") — the
worktree root is the correct invocation; only its BASE is skewed by the
cutover.

## Finding 1 (P2) — IN DIFF — repaired

> 다른 감사 인스턴스의 시작 마커를 삭제하지 마세요 —
> internal/hook/audit_receipt_guard.go

Two hazards named, both real on the first fix version:

1. Two concurrent same-role background auditors share the derived
   `bg_<session>_<agent_type>` marker; the first accepted PASS deleted it,
   leaving the second instance's valid PASS refused `start marker missing`.
2. An agent-id-carrying auditor's FAIL cleared keys including the derived
   one, deleting a concurrent background auditor's marker.

**Repair (era-anchor semantics for derived keys):**
- `recordAuditorStart` writes a derived-keyed marker keep-earliest: an
  existing marker is not overwritten, so a second concurrent start does not
  move `StartedAt` past receipts the first instance will cite.
- `readStartMarker` now returns the marker AND the key it was found under;
  `consumeStartMarker` deletes only that found key, and never a derived key —
  a derived marker is a session-era anchor that no single stop may destroy.
  Agent-id-keyed markers keep the exact per-instance consume semantics.
- Added tests: `TestSubagentStop_ConcurrentBackgroundAuditorsShareEraAnchor`
  (both concurrent PASSes accepted, keep-earliest pinned) and
  `TestSubagentStop_AgentIDFailKeepsBackgroundMarker`; the accepted-PASS
  assertion in `TestSubagentStop_BackgroundSpawnAuditorPassIsProvable` now
  pins retention. Family run green (see verdict file).

Residual (accepted): derived markers outlive their session as inert orphans
(keyed by session id, never read again; `.moai/state` is machine-local).
Anti-recycling under a derived key is session-scoped rather than
instance-scoped — the finest identity an agent-id-less payload carries.

## Findings 2–8 — OUT OF DIFF — handed to the leader

All cite files this card never touched (`internal/cli/factory_card.go`,
`internal/cli/todo.go`, `internal/cli/todo_issuance.go`,
`internal/web/screens.templ`); they describe main-side code between the
resolver's base and this branch point, not this card's change:

- P1 hub-hint preservation on re-`assign` (factory_card.go:1310)
- P2 picked-candidate hub-predecessor skip (factory_card.go:821)
- P2 explicit `--files` in overlap computation (todo.go:940)
- P2 issuance guidance for `add --pick` (todo.go:873)
- P2 engage guidance self-comparison (todo_issuance.go:274)
- P2 SPEC-file read deadline on the I/O itself (todo_issuance.go:152)
- P2 SVG colors for the relation graph (web/screens.templ:358)

These belong to the leaders'/operators' queue triage, not this card.

## Also observed (out of card, measured on this machine)

`TestStaleRunNoticeFactoryLegacyLabel` (internal/hook/stale_run_m1_test.go)
fails **pre-existing on mainline**: reproduced on the primary checkout at
`ec13872f3` (main, none of this card's changes) — expects
`moai factory relaunch --provider cc` while the notice emits `--provider glm`.
Not caused by this card; flagged for the leader.

## Verdict

In-diff finding repaired and regression-pinned; remaining findings out of
diff. This card's diff is judged sound for merge by the lane, subject to the
leader's independent evidence read.
