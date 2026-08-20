# t110 review verdict — lead-role discipline codification (kanban-dispatch.md)

- Reviewer: lead session (review hub)
- Card: t110 · Worktree: `.claude/worktrees/t110` · Branch: `WT-t110` @ `5c9a9715c` (base `0ede5db6a`)
- Delta reviewed: `0ede5db6a..5c9a9715c` (3 files: rule + template mirror + evidence)
- Lens: default 4-perspective + budget-discipline verification
- Evidence read: `.claude/worktrees/t110/.moai/reports/t110/evidence.md` (135 lines, 5-section)
- Note: verdict file lives in the lead's release-v311 tree (worktree isolation blocks card-worktree writes — same as t103).

## Verdict: PASS — with one cross-file residual flagged (§R, follow-up)

## 1. Dispatch focus items (all four verified)

| # | Focus | Check performed | Result |
|---|-------|-----------------|--------|
| 1a | Budget arithmetic 27,764 → 27,709 (−55B) | Reviewer-measured `wc -c` at base and tip via `git show`: 27764 / 27709 — exact. Evidence honestly records the rejected interim (+419) state | PASS |
| 1b | 8 compressions weaken no rule | Full diff read sentence-by-sentence. The two argument-heavy passages (dispatch-language rationale; CodeRabbit endpoint discipline) preserve every argument step — classification-not-exemption, who-reads-it boundary, verbatim-address rule; combined-vs-plural endpoint, both-halves-required, positional hazard, `created_at` selection. Several sites STRENGTHEN (three new [HARD] blocks). No weakening found | PASS |
| 2 | Mirror byte-identical | Reviewer shasum: local = template = `f1d96d233…`. Lane also ran `MOAI_TEMPLATE_LEAK_STRICT=1 go test ./internal/template/` ok 24.066s (lane-attributed) | PASS |
| 3 | Principle-① reconciliation complete | WITHIN kanban-dispatch.md: complete — the boundary note now reads "nudge delivery rides on cross-session messaging … the queue itself keeps working without it", and the new §"The delegation channel is the queue" codifies queue=channel / message=nudge. **BUT one residual survives in the sibling file** — see §R | PASS with flag |
| 4 | Budget-test overflow 666 attribution | Test exists (`internal/config/token_budget_guard_test.go`, `-run 'TestAlwaysLoadedTokenBudget$'`); lane output: 76,666 vs 76,000 = overflow 666, pre-existing (guard landed in base; lead measured ~680 over BEFORE this card). t110's only always-loaded edit is −55B ≈ −14 tokens — the overflow was equal or larger at base. Arithmetic verified; absolute resolution is t114's (out of scope, agreed) | PASS |

## 2. The three [HARD] codifications (content check)

1. **Queue = delegation channel, message = nudge** — correctly grounded on the t106-landed property (primary-checkout resolution, one repo one queue). No dispatch may depend on a message arriving. Sound.
2. **Lead = sole producer; promotion = operator's act** — production-is-translation-not-invention framing preserves the operator-origin principle and ADDS "empty queue is a state to report, not a prompt to invent work".
3. **Final PASS/FAIL verdict is the lead's** — never delegated to the producing lane, with the backend-neutral structural rationale (mixed-backend boards cannot commission judgment onto the lead's backend). This codifies what this round already practiced.

## 3. Baseline attribution

Sizes measured by the reviewer at both commits (object store). Test outputs (template suite, budget-guard FAIL output) lane-attributed on WT-t110's tree. The budget figure 76,666 is the guard's own output — authoritative over the 4 B/token approximation.

## 4. Gaps

- Byte→token conversion approximate (guard figure authoritative) — declared by the lane.
- Full-suite verdict deferred to CI (lane-local discipline).
- `chief` naming / `-k` entry guidance / SessionStart notice — deliberately NOT this card (t97 scope), no new term introduced. Verified: none appears.

## 5. Residual risks

- Sole-producer is a documented norm, not a mechanical gate — the operator's direct `moai todo add` still works (by design; the norm binds the lead).
- The mixed-backend verdict rationale should be re-read if the backend set changes.

## R. Cross-file residual (flagged — one sentence + mirror)

`cross-session-messaging.md:19` (local AND template mirror) still reads: "because Kanban Mode's lead–companion dispatch **rides entirely on this channel**, they bound where that mode can operate" — the exact absolute claim t110 reconciled in kanban-dispatch.md, in the very file kanban-dispatch.md:13 cross-references. Under the new doctrine the availability constraints bound NUDGE delivery (and thus dispatch latency), not the mode's operability — the queue keeps the board moving without the channel.

Disposition: NOT a t110 defect (its file is fully reconciled; its budget constraint was per-file and met). Route as a one-sentence rider to **t114** (the in-flight always-loaded diet card, already editing these files under the 666-overflow regime where any fix needs compensating trim) or a minimal follow-up card. Recorded in the follow-up candidate list as item ⑦.
