# t106 review verdict — todo queue root resolution (primary checkout + lossless fallback)

- Reviewer: lead session (operator-confirmed direct verdict)
- Card: t106 · Worktree: `.claude/worktrees/t106` · Branch: `WT-t106` @ `f5297037f` (base `5c3141372`; draft `6ba8ef90e` included — ancestry verified)
- Delta reviewed: `5c3141372..f5297037f` (8 files, +672/−9)
- Lens: `--deep` (path resolution + migration logic + file-manipulation surface)
- Evidence read: `.claude/worktrees/t106/.moai/reports/t106/evidence.md` (220 lines, 5-section, live measurements)

## Verdict: PASS

## 1. Claims reviewed (dispatch + evidence vs this review's direct reads)

| # | Claim | Check performed | Result |
|---|-------|-----------------|--------|
| 1 | Commit/lineage/ref | `git log -3`, `merge-base --is-ancestor 6ba8ef90e` → DRAFT-OK, `show-ref refs/heads/WT-t106` = `f5297037f19a…` | PASS |
| 2 | 8 files +672/−9 | `git diff --stat` — exact | PASS |
| 3 | Primary resolution via `gitcore.ResolveGitDirs` | `ResolveGitDirs` exists (`internal/core/git/checkout.go:54`); `resolveTodoQueueRoot` returns `filepath.Dir(CommonDir)` — the branch_guard discrimination precedent reused, no file moved | PASS |
| 4 | Fallback `~/.moai/todo/<base>-<sha256[:8]hex>/` | `fallbackTodoQueueRoot` + `todoQueueProjectKey` (`sum[:4]` hex = 8 chars) read in full; home-unresolvable degrades to project-local (no hard error); naming rationale (todo, not kanban — session-record scope) documented and honored | PASS |
| 5 | Adopt-not-shadow | `adoptLocalTodoQueue`: target-exists → no-op (idempotent, no re-adoption); same-volume atomic rename; cross-volume copy KEEPS the original (lossless — deletion forbidden); best-effort failure leaves local queue intact for a later retry | PASS |
| 6 | [HARD] ① worktree list == primary | Evidence: measured twice with direct python reads of the primary file (35→34 tracked exactly while the lead processed a card) — lane-attributed live measurement | PASS-attributed |
| 7 | [HARD] ② worktree add → primary | Evidence: `todo add` from the worktree landed in primary (direct read `found in PRIMARY: True`), cleaned up (`todo done`) — lane-attributed | PASS-attributed |
| 8 | [HARD] ③ adoption test | `TestTodoQueue_FallbackAdoptsExistingLocalQueue` exists (`todo_queue_root_test.go:152`), runs the REAL resolution path; `-count=1 -v` PASS 0.15s; no-re-adoption asserted. The manual out-of-tree repro was correctly REFUSED by the worktree guard (and not routed around — doctrine honored) | PASS |
| 9 | Stale-lock sweep | `NewBacklogStore` removes `legacyBacklogLockFileName` (`backlog.json.lock`) best-effort; live `backlog.lock` untouched; measured before/after on the primary (STALE LOCK STILL PRESENT → SWEPT) | PASS |
| 10 | Session records untouched (scope guard) | Diff file list contains no `<uuid>.json` path; evidence §Scope-guard confirms — 35 session records never read/moved | PASS |
| 11 | Single constructor | All 5 verb call sites (add/list/done/next/unpick) route through `newTodoStore()` — verified in diff | PASS |
| 12 | Skill doc + mirror | `todo.md` ↔ template shasum identical (`e804cf01b…`); template-side added lines neutrality grep → empty | PASS |
| 13 | Test coverage shape | 6 test functions: worktree-converges, primary-is-itself, subdir→repo-root, no-git fallback (+key shape), adoption, worktree-sees-primary | PASS |

## 2. Evidence (this review's commands)

- `git log / merge-base / diff --stat / --name-only / show-ref` @ `f5297037f`
- Full diff read: `internal/cli/todo.go` (resolution core), `internal/kanban/backlog_store.go` (lock sweep)
- `git show | shasum` mirror pair; template added-line neutrality grep
- `git grep 'func ResolveGitDirs'` — precedent existence
- Read: evidence.md in full

## 3. Baseline attribution

- Code reads: tree @ `f5297037f`. Live-queue measurements ([HARD] ①②, lock sweep) are the lane's, taken against the primary's real backlog.json that morning with the branch-built binary — attributed in evidence with verbatim outputs. Test-run outputs lane-attributed (same rationale as prior verdicts; CI owns the full matrix).

## 4. Gaps

- `internal/hook/session_start_kanban.go:202` reads the queue under `input.ProjectDir` — the same split class for a worktree session's bootstrap notice (unmeasured ProjectDir behavior; follow-up card recommended — recorded as candidate).
- Bare-repository launch contexts unhandled (Dir(CommonDir) would be the bare repo parent) — declared unsupported.
- Full-suite / cross-platform deferred to CI (darwin local only).

## 5. Residual risks

- No-git projects key the fallback by launch directory — a subdir launch gets a different key than the root launch. Pre-existing sensitivity class (the old code was equally CWD-dependent); git projects are now fully convergent. Worth documenting if no-git usage matters.
- Git submodule: CommonDir belongs to the submodule — queue roots at the submodule checkout (reasonable, unmeasured).
- Cross-volume adoption leaves an inert downgrade-era local copy; a later downgrade to pre-fallback binaries would read the stale local copy, not the live fallback queue (deliberate losslessness trade-off).
