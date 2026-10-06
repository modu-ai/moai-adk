# Card t1544 — verdict (run phase)

Card: t1544 (Class B, P1) — audit receipt guard deadlock for background
Agent() spawns.
Tree: `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1544`, branch
`WT-audit-receipt-guard`, point `10df085da` (origin/main at intake) + this
card's diff. Run: tmf011, lane-26.

## Claim

1. The defect is real and reproduced RED first: a SubagentStart payload
   without agent_id wrote no auditor start marker, and the auditor's
   SubagentStop was refused `start marker missing` even with a valid receipt —
   the structural refusal behind the t1509 phase-entry deadlock.
2. The fix (card direction B — session_id+agent_type marker key) repairs it:
   the background ceremony now closes — start marker recorded under the
   derived key, proven PASS accepted, marker semantics safe for concurrent
   same-role instances (card-review P2 repair).
3. Scoped verification is green; the 9 failures in the full
   `internal/hook` package run are pre-existing on mainline, reproduced on
   the primary checkout at `ec13872f3` without this card's changes.

## Evidence

RED (pre-fix, this tree, this run — the defect):

```
--- FAIL: TestSubagentStart_BackgroundSpawnWithoutAgentIDRecordsMarker (0.20s)
    audit_receipt_guard_test.go:418: plan-auditor: no start marker under the background key "bg_sess-bg-1_plan-auditor": ... no such file or directory
--- FAIL: TestSubagentStop_BackgroundSpawnAuditorPassIsProvable (0.33s)
    audit_receipt_guard_test.go:452: decision = "block", want none — ... (reason "AUDIT_RECEIPT_VIOLATION: plan-auditor reported PASS but the audit could not be corroborated — start marker missing. ...")
```

(also RED: the asymmetric id-at-stop case, and the recycled-receipt test
blocked `start marker missing` instead of reaching
`receipt created before the auditor started`.)

GREEN (post-fix, this tree, this run):

```
$ go test ./internal/hook/ -run '<guard family + 8 new tests>' -count=1
ok  	github.com/modu-ai/moai-adk/internal/hook	13.915s

$ go test ./internal/auditreceipt/... -count=1
ok  	github.com/modu-ai/moai-adk/internal/auditreceipt	1.388s

$ go vet ./internal/hook/... ./internal/auditreceipt/...
(exit 0, no output)

$ golangci-lint run ./internal/hook/... ./internal/auditreceipt/...
0 issues.

$ gofmt -l <3 changed files>   → (empty)
```

Full package (this tree, `-timeout 30m`, 776s):

```
FAIL	github.com/modu-ai/moai-adk/internal/hook	776.049s
--- FAIL: TestStaleRunNoticeFactoryLegacyLabel
--- FAIL: TestUserPromptSubmitHandler_MultipleSpecs (+7 more of the same two families)
```

Pre-existence control (primary checkout, main `ec13872f3`, NONE of this
card's changes, this run):

```
$ go test ./internal/hook/ -run 'TestUserPromptSubmitHandler|TestBuildSessionTitle' -count=1
FAIL	github.com/modu-ai/moai-adk/internal/hook	0.891s
$ go test ./internal/hook/ -run TestStaleRunNoticeFactoryLegacyLabel -count=1
FAIL	github.com/modu-ai/moai-adk/internal/hook	1.568s
$ go test ./internal/hook/ -run TestBuildSessionTitle_NilConfig -count=1
--- FAIL: TestBuildSessionTitle_NilConfig (0.00s)
    user_prompt_submit_test.go:732: buildSessionTitle() with nil config = "leader", want derived title
```

All 9 failing tests fail identically on mainline without this card's diff →
out of card. (Run 1 of the suite additionally hit the 10m default timeout;
the repo doctrine form `-timeout 30m` was used for the verdict run.)

Card review: `.moai/reports/t1544/card-review.md` — codex scope=card verdict
`fail`; the single in-diff finding (P2 shared-marker deletion) repaired and
pinned by two new tests; 7 findings out of diff (review base resolved to the
develop tip `a158b4b5f` behind main — cutover residue), handed to the leader.

## Baseline-attribution

Every figure above was measured in this run, on this tree
(`10df085da` + diff), with the lane env scrubbed
(`unset MOAI_KANBAN_ID MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED`)
for test runs; the pre-existence controls were measured on the primary
checkout at `ec13872f3` in the same run. Tool provenance: the suite runs
used the `go` toolchain directly; the card review ran through
`mcp__moai__codex_review` with `project_root` = this worktree (the parent
root returned `inconclusive`).

## Gaps

- The exact runtime shape of a background spawn's SubagentStop payload
  (whether it carries agent_id, agent_transcript_path) was not directly
  observed in this run; the fix covers every combination of id-at-start /
  id-at-stop through ordered key lookup, and the tests pin both symmetric
  and asymmetric shapes.
- CI has not run on this branch yet (PR just opened); the full-suite
  mainline verdict belongs to CI.
- The codex review's 7 out-of-diff findings were not reproduced or
  repaired here (they live in main-side code outside this card's diff).

## Residual-risk

- Derived-key markers are session-scoped anchors: anti-recycling under them
  is per-session-per-role, not per-instance (the finest identity an
  id-less payload carries). A receipt minted after the session's first
  same-role background auditor started verifies for any of them.
- Derived markers outlive their session as inert orphans (keyed by session
  id, never read again) in machine-local `.moai/state`.
- The serial-slot nominated-lease asymmetry observed during intake
  (`factory next --card` counts merely-assigned siblings as holders while
  bare arm-(a) does not) is a separate design wrinkle in
  `internal/cli/factory_card.go`, not touched by this card.
