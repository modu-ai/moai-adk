# t2027 card-review (advisory)

- tool: `codex_review` scope=card, backend codex, project_root = the t2027 tree; run once (no re-review)
- base (merge base the tool computed): 59464c5f37201f17e712a250caf6a46c508a35cf
- reviewed tree: /Users/goos/.moai/worktrees/projects/moai-adk-go/t2027 at HEAD 8aed1a8293 (fix commit on top of merge 9fc89e246e)
- tool verdict: `fail` (advisory, non-binding; the reviewer states it read statically and ran no test, build, or reproduction)

## Findings (verbatim titles) and disposition
| # | Severity | Title | File | Disposition |
|---|---|---|---|---|
| 1 | P1 | 중간 디렉터리 변경까지 확인한 뒤 Git 대상 경로를 정하라 | internal/hook/branch_guard.go:463 | not adopted: file is not in this card's diff |
| 2 | P2 | 모듈 의존성 변경도 검사 대상에 포함하라 | internal/cli/factory_release_route.go:252 | not adopted: file is not in this card's diff |
| 3 | P2 | 변경 목록 조회 실패를 빈 변경 목록으로 처리하지 마라 | internal/cli/factory_release_route.go:280 | not adopted: file is not in this card's diff |

Reason for "not adopted": this card's own commit (8aed1a8293) changes only internal/homestate/location_memo.go and internal/homestate/location_memo_symlink_test.go. The scope=card base is the card's branch point, so the diff the reviewer saw also contains the merged `release/main-batch` (f9a4efd964, merge 9fc89e246e); all three findings sit in Go code from that batch, which this card neither wrote nor was dispatched to change (the card targets mergeq.go, location_memo.go and af.go). They are the same three findings the t2022 reviews reported. Carried to the leader as information; not recorded as a pass.

## Scope note
- This review covers HEAD 8aed1a8293 only. The later commits 8dab263d8b, 7bcadec9c3 and 83f81f3042 (mergeq refusal of tracked gate changes, plus the cli pin test) were NOT reviewed by the tool: the leader forbade a re-review for that item.

## Gaps
- The tool output names no finding on the changed homestate files, and none on the three files the card targets; it does not show whether it read them in detail.
- The review is static; no test or build was run by the reviewer.
- This is not an independent audit and does not replace the leader's evidence read.

## Residual risk
- The `fail` verdict stays on record for the three batch findings; whether the leader reads it as card-blocking is a leader call, since none touches this card's diff.
