# t1923 — split proposal (option b), SPEC-FACTORY-BOOT-MINIMAL-001

Status: PROPOSAL ONLY, written for the leader. No SPEC was edited for it; no audit was run for it.

## Why a split is on the table

Plan audit round 1 FAIL 0.69 (`plan-audit.md`), round 2 FAIL 0.76 (`plan-audit-iter1.md`, rcpt-87831eae3fa19ae402a21281, audited dce3279bcd). The Tier M ceiling (2 rounds) is reached; `moai spec ceiling` reads "ceiling HIT at 2 rounds, disposition hold, split required" (`plan_audit_ceiling_policy.on_final_hit=hold-and-split`). D1-D5 of round 1 were reproduced as closed; round 2 left four blocking defects:

| Defect | Where | Belongs to |
|---|---|---|
| N1 lane-rule keeps-check missing (a safety sentence outside R1-R5 can be deleted unseen) | AC-FBM-003 + `TestT1923RelocationProbe` | lane notice / relocation |
| N2 plan §C grep finds 28 test files, §A classifies 14 | plan §A/§C, AC-FBM-009 list | notice-pinning tests (hook/cli) |
| N3 `agents_md_check.py` matches markers, not obligation direction | AC-FBM-004 + `OBLIGATIONS` | root `AGENTS.md` |
| N4 AC-FBM-001 probe misses text appended after the pointer line | AC-FBM-001 + `t1923NoticeViolations` | notice content probe |

## Proposed split (by the audit's own anchors)

- **Part B — AGENTS.md byte headroom** (own SPEC): REQ-FBM-010, AC-FBM-004, `probe/agents_md_check.py`, defect N3. Root `AGENTS.md` only, no Go code, no hook tests, independent of every notice probe. The Codex ceiling tests are green now (24,501 B vs 24,576 B, margin 75 B, measured at 5995962990), so this is a headroom goal (target at or under 24,276 B) with the operator patch to be re-applied by hand (it does not apply as-is: `git apply --check` exit 1). Smallest of the parts; fresh audit rounds for it do not depend on the notice work.
- **Part L — lane notice + lane-rule relocation + budget guards** (own SPEC): REQ-FBM-002, -003, -004, -005, -006, -007, -008, -009, AC-FBM-002, -003, -009, -010, -011 (+ the regression guards), defects N1, N2. This is the part with the actual risk: the English lane composite is 9,946 against the 10,000 limit (54 characters of margin; a non-factory-allowance probe reached 10,006). Carries R1-R5 relocations, the fixed-part ceiling 2,845 and the breach overlay.
- **Part N — leader notice** (optional, can be dropped or deferred): REQ-FBM-001 and the leader half of AC-FBM-001, defect N4 (the pointer-line tail). The leader notice is already under the fixed-part ceiling (English 2,573 vs 2,845), so minimizing it is not needed for the budget; the card names it, but the SPEC already preserves the elements other completed SPECs pin (SPEC-AUTONOMY-BATCH-GATE-001 REQ-BGS-016/017). Dropping or deferring it removes one of the four defect sites and the only place that collides with completed-SPEC contracts.

REQ-FBM-011 (live six-lanes-plus-leader observation) is the completion check of Part L (and N if kept). REQ/AC ids above are mapped from the ids and titles as read in `spec.md` / `acceptance.md`; the splitting SPEC author should re-verify the per-id mapping before writing the new SPECs.

## What a split does not remove

N1, N2 and N4 all sit in the notice/relocation parts, so the lane part still has to pass an audit on three of the four open findings; the split mostly isolates N3 and the leader-notice risk and gives each new SPEC its own round count (the round count is SPEC-scoped evidence). The three open audit fixes are small anchor-level edits (see the `fix_scope` line of `plan-audit-iter1.md`).

## Records and state to keep in mind

- The ceiling-exception correction round (`manager-spec` for N1-N4) was started on the leader's earlier message and **stopped** when the leader retracted it. Its partial edits are uncommitted in the worktree (probe files only: `probe/zz_t1923_boot_probe_test.go.txt`, two `red-now-*` files, plus new `probe/lane-rule-retained-5995962990.json` and `probe/research/gen-overlay.json`) and saved as a patch at `/private/tmp/claude-501/-Users-goos-MoAI-moai-adk-go/e88b03dd-22ba-461d-a9a1-4a610a50f193/scratchpad/t1923-stopped-round-wip.patch`. They are unreviewed and not part of any audited tree.
- `progress.md` §G of the SPEC carries a line "운영자 결정 2026-10-23: 천장 예외 1회 ..." written (commit 65cff3e812) BEFORE the retraction arrived. It is now stale: that exception is not approved. It was NOT edited or deleted; the leader decides whether to append a retraction line.
- Existing verdict and wait records are preserved as they were.
