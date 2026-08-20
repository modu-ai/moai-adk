# t103 review verdict — statusline segment notation + session-line placement

- Reviewer: lead session (review hub; operator-confirmed direct verdict)
- Card: t103 · Worktree: `.claude/worktrees/t103` · Branch: `WT-t103` @ `062e05017`
- Delta reviewed: `0ede5db6a..062e05017` (3 files, +55/−22 — renderer.go + 2 test files). NOTE: the dispatch named base `c2a520cc9`; the correct review delta starts at `0ede5db6a` (see §1 focus-3).
- Lens: default 4-perspective (local diff, no flags)
- Evidence read: `.claude/worktrees/t103/.moai/reports/t103/run-evidence.md` (46 lines, 5-section)

## Verdict: PASS — with one notation flag for the operator (§F)

## 1. Dispatch focus items (all three verified)

| # | Focus | Check performed | Result |
|---|-------|-----------------|--------|
| 1 | Session line moved last; regression test real & run; no-identity sessions keep 3 lines | Code: `renderSessionLine` moved after L3 with the identical `!= ""` conditional guard (position-only change, behavior parity). Test: `TestRenderDefaultV3_SessionLineClosesTheStatusline` exists (`session_identity_test.go`), asserts last-line placement AND absence of 🏷️/🔄 TODO: on prior lines. Evidence: `go test ./internal/statusline/` ok 15.041s with the new test included. Precision note: the 3-line preservation holds when identity, backlog, AND GitHub segments are all absent — this guard shape is pre-existing and unchanged by t103 | PASS |
| 2 | Numeric meaning (first=picked, second=queued) grounded in code | Two-point verification: (a) renderer format `fmt.Sprintf("🔄 TODO: %d/%d", data.Backlog.Picked, data.Backlog.Queued)` — order explicit; (b) data source `internal/statusline/backlog.go resolveBacklogCounts` counts by state (`picked` → Picked, `queued` → Queued), fail-open on absent/unreadable/corrupt | PASS |
| 3 | Base mismatch (dispatch c2a520cc9 vs tip 062e05017) | `git log c2a520cc9..062e05017` = t106 draft + rework + merge + t103. `git log --stat -- internal/statusline/` in that range: ONLY 062e05017 touches statusline — the t106 commits are statusline-unrelated, as claimed. Correct review delta therefore `0ede5db6a..062e05017` (what this verdict binds) | PASS |

## 2. Other claims (evidence §1 items 3-4)

- repo 🔀→📡: already landed at `caf435ec4`; lane's sweep grep (repo-icon misuse) 0 hits; the icon-distinction test asserts 📡 repo vs 🔀 PRs — accepted.
- Separator │→| item: **out of verdict** per the dispatch (closed as "no change exists"; operator intent query in progress elsewhere) — honored, not judged here.

## 3. Baseline attribution

Code reads @ `062e05017` (WT-t103 ref verified). Test outputs lane-attributed (statusline suite ok 15.041s, vet/lint clean); CI owns the full matrix.

## 4. Gaps

- No on-screen rendering check (statusline displays in the operator's terminal; unit tests pin the strings and ordering).
- Full suite not run locally (lane-local discipline).

## 5. Residual risks

- The disabled-GitHub test still asserts ⚠️ absence (now vacuously true) — cosmetic test-tightening candidate, noted by the lane.
- `resolveBacklogCounts` reads `.moai/state/kanban/backlog.json` under its own `boardRoot` — the same worktree-split family as t106's CLI fix and its `session_start_kanban.go:202` gap. Out of t103's scope (notation-only card); added to the follow-up candidate list.

## F. Notation flag (operator confirm — one-string fix if the card text stands)

The queued card text records the operator's 3rd confirmation as "숫자 슬래시 양쪽 공백('4 / 20', '2 / 1')" — spaces around the slash on BOTH segments. The implementation renders:
- backlog: `🔄 TODO: 12/26` — **no spaces** around the slash
- github: `🔀 7 / 3` — spaces around the slash

The two segments in the same line are therefore notationally inconsistent with each other, and the backlog form does not match the queue-text example '4 / 20'. Not a FAIL — cosmetic, single string.

**RESOLVED (operator, same session as this verdict): 공백 복원 채택.** The lane is instructed to change the backlog format string to `"🔄 TODO: %d / %d"` and update the 2 test expectations (`🔄 TODO: 12/26` → `🔄 TODO: 12 / 26` in github_test.go and session_identity_test.go) BEFORE the t103 merge. The PASS binds with this rider.
