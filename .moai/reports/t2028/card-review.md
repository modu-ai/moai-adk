# t2028 card-review (advisory)

- tool: `codex_review` scope=card, backend codex, project_root = the t2028 tree; run once (no re-review)
- base (merge base the tool computed): 59464c5f37201f17e712a250caf6a46c508a35cf
- reviewed tree: /Users/goos/.moai/worktrees/projects/moai-adk-go/t2028 at HEAD d7c11ab7d3
- tool verdict: `fail` (advisory, non-binding; the reviewer states it read statically and ran no test, build, or reproduction)

## Findings (verbatim titles) and disposition
| # | Severity | Title | File | Location tag | Disposition |
|---|---|---|---|---|---|
| 1 | P1 | 복합 명령의 마지막 작업 디렉터리를 확인하라 | internal/hook/branch_guard.go:463 | outside | not adopted: file is not in this card's diff (docs/agent definitions only) |
| 2 | P1 | 게이트 이후 변경을 고정된 merge SHA와 비교하라 | internal/agenticfactory/mergeq/gitcli.go:238 | outside | not adopted here; carried to the leader as information (see below) |

Reason: this card changes documentation, agent definitions and one test-data file (`git diff --stat fe4634ff43 -- '*.go'` is empty; no Go file in the card's own commits). The scope=card base is the card's branch point, so the reviewed diff also contains the merged `release/main-batch`; both findings sit in Go code from that batch.

## Notes for the leader
- Finding 1 repeats the branch_guard.go:463 finding already reported by the t2022 and t2027 reviews.
- Finding 2 is new in wording: it is about `TrackedChanges` in `mergeq/gitcli.go` (added by the t2027 fix, commit 7bcadec9c3): `git diff --name-only -z HEAD --` compares against the check worktree's current HEAD, so a gate that commits its edit or checks out another commit would read as clean while `landUnderLock` lands the earlier `cand.Result`. This matches the residual-risk class t2027 recorded; it was neither reproduced nor refuted here (static claim only). Whether it blocks RC40 is a leader call; the fix would belong to a t2027 follow-up, not to this card.

## Gaps
- The review is static; no test or build was run by the reviewer.
- The tool output names no finding on the doc, agent-definition or skill files this card changed, and does not show whether it read them in detail.
- This is not an independent audit and does not replace the leader's evidence read.

## Residual risk
- The `fail` verdict stays on record for the two batch findings; whether it is card-blocking is a leader call, since neither touches this card's diff.
