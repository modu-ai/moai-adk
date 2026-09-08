# t463 — Disposal Evidence (lead-approved 3-tree disposal)

Date: 2026-09-03 · Executor: lane-13 (now in worktree t463, branch WT-lane-push-doctrine)

## Claim
Lead-approved disposal of worktrees t440 / t442 / t449 executed after independent re-measurement (lead's relayed values were not trusted blindly).

## Evidence (commands + verbatim outputs, this run)
- `git fetch origin develop` → rc=0
- `git rev-parse origin/develop` → `d592b0551eeb731e5bbd3ef330bf71b21c0822c9`
- `git status --porcelain` (t440, direct read while anchored there) → 0 lines
- `git merge-base --is-ancestor HEAD origin/develop` (t440 HEAD `b80cc9cf1`) → rc=0
- `git merge-base --is-ancestor 57d07998e origin/develop` (t442) → rc=0
- `git merge-base --is-ancestor 7e6701cfd origin/develop` (t449) → rc=0
- `git worktree remove <t442>` → rc=0 · `<t449>` → rc=0 · `<t440>` → rc=0 (no `--force`; git's own dirty-refusal is the cleanliness net)
- `git worktree list | grep -E '/t440|/t442|/t449'` → no output (grep rc=1 — no residue)
- Branches kept per lead instruction: `WT-delivery-notice-docs`@`b80cc9cf1` · `WT-glm-effort-measured`@`57d07998e` · `WT-integration-lock-record`@`7e6701cfd`

## Baseline attribution
All commands run in this session on 2026-09-03: measurements from the t440 worktree before anchor move; removals from the t463 worktree after `ExitWorktree(keep)` → `EnterWorktree(t463)`; against `origin/develop` = `d592b0551` after a fresh fetch.

## Gaps
- t442/t449 pre-removal `git status --porcelain` was NOT directly read: the worktree-session guard refused `git -C` cross-tree reads from the isolated session. Cleanliness rests on (a) git's remove-time dirty-check (refuses dirty trees without `--force`; all three returned rc=0) and (b) the lead's independent dirty-0 reading.
- `moai session list --json` returned `[]` — the registry provided no live-anchor signal for t442/t449; it neither proved nor disproved a live session anchored there.

## Residual risk
An untracked non-ignored file in t442/t449 would have made `git worktree remove` refuse (rc≠0) — none did. Ignored scratch (`.moai/state`, cache) was intentionally disposed with the trees. Branch residue cleanup remains a separate matter per lead.

## Card-premise measurement (t463 target file)
- `CLAUDE.local.md` IS tracked on `origin/develop` (ls-tree: blob `d77aa86e`, mode 100644); NOT gitignored (`git check-ignore -v` rc=1); tracked in the card worktree index (`git ls-files --error-unmatch` rc=0).
- Lead's fix items ①②③ are ALREADY satisfied on develop @ `d592b0551`:
  - ① lane window procedure ("창을 받으면:") contains no push step — `git push` appears only inside the :349 prohibition text and the :378 lead-batch procedure;
  - ② "**운영 절차**" bash block carries `# push는 창 밖 — 리드가 레인 병합 SHA를 모아 일괄로 한다 (아래 절차)`;
  - ③ lane completion-report items (~:346) already list "로컬 병합 SHA".
- REMAINING DELTA = item ④ only: ~:349 second sentence "카드가 마감되면 로컬 develop 병합(창 경유 `git push origin develop`)이 **유일한** 공개 경로다" — parenthetical places push inside the window and names no actor (readable as lane-subject).
- The lead's cited coordinates :305/:321 match the STALE primary checkout (`main` @ `7ad9f8534`), not develop — stale-primary-copy misread; premise correction recorded here and to be carried into the SPEC.
