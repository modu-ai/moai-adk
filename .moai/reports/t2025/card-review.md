# t2025 card-review (advisory)

- tool: `codex_review` scope=card, backend codex, project_root = the t2025 tree; run once (no re-review)
- base (merge base the tool computed): 59464c5f37201f17e712a250caf6a46c508a35cf
- reviewed tree: /Users/goos/.moai/worktrees/projects/moai-adk-go/t2025 at HEAD 4b0d275fe6 (test fix on top of merge 6b29bf9b80)
- tool verdict: `fail` (advisory, non-binding; the reviewer states it read statically and ran no test, build, or reproduction)

## Findings (verbatim titles) and disposition
| # | Severity | Title | File | Disposition |
|---|---|---|---|---|
| 1 | P1 | 복합 명령의 실제 Git 작업 디렉터리를 확인하라 | internal/hook/branch_guard.go:463 | not adopted: file is not in this card's diff |
| 2 | P2 | 패키지를 찾지 못했다고 필수 시험을 생략하지 마라 | internal/cli/factory_release_route.go:252 | not adopted: file is not in this card's diff |

Reason for "not adopted": the card's own commit (4b0d275fe6) changes only internal/cli/factory_age_alarm_test.go (24 insertions, 3 deletions). The scope=card base is the card's branch point, so the diff the reviewer saw also contains the `release/main-batch` merge (98d6257439) brought in at 6b29bf9b80; both findings sit in Go code from that merged batch, which this card neither wrote nor may change (card scope: only the test file). Finding 1 repeats the t2022 review's finding 1 (same file and topic). The findings are carried to the leader as information about the rc.40 candidate, not repaired here.

## Gaps
- No finding was reported on the changed test file; whether the reviewer read it in detail is not shown in its output.
- The review is static; no test or build was run by the reviewer.
- This is not an independent audit and does not replace the leader's evidence read.

## Residual risk
- The `fail` verdict stays on record unresolved for the two batch findings; whether the leader reads it as card-blocking is a leader call, since neither finding touches t2025's diff.
