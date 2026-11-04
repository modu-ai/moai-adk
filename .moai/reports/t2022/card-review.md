# t2022 card-review (advisory)

- tool: `codex_review` scope=card, backend codex, project_root = the t2022 tree; run once (no re-review)
- base (merge base the tool computed): 59464c5f37201f17e712a250caf6a46c508a35cf
- reviewed tree: /Users/goos/.moai/worktrees/projects/moai-adk-go/t2022 at HEAD fb6ed75ce4 (docs commit on top of merge 742626a572)
- tool verdict: `fail` (advisory, non-binding; the reviewer states it read statically and ran no tests, build, or reproduction)

## Findings (verbatim titles) and disposition
| # | Severity | Title | File | Disposition |
|---|---|---|---|---|
| 1 | P1 | 복합 명령의 실제 실행 위치로 BranchGuard를 판정하라 | internal/hook/branch_guard.go:463 | not adopted: file is not in this card's diff |
| 2 | P2 | 실제 ledger reader를 웹 콘솔에 연결하라 | internal/cli/web.go:152 | not adopted: file is not in this card's diff |
| 3 | P2 | 변경 파일 목록을 NUL 구분자로 읽어라 | internal/cli/factory_release_route.go:280 | not adopted: file is not in this card's diff |

Reason for "not adopted": the card's own commit (fb6ed75ce4) changes only README.md, README.ko/ja/zh.md and docs-site/content pages (26 files). The scope=card base is the card's branch point, so the diff the reviewer saw also contains the `release/main-batch` merge (5a3ccd1a4d) brought into the tree at 742626a572; all three findings sit in Go code from that merged batch, which this card neither wrote nor may change (card scope: no code). The findings are carried to the leader as information about the rc.40 candidate, not repaired here.

## Gaps
- No finding was reported on the docs files themselves; whether the reviewer read them in detail is not shown in its output (the summary lists only the three Go files).
- The review is static; no test or build was run by the reviewer.
- This is not an independent audit and does not replace the leader's evidence read.

## Residual risk
- The `fail` verdict stays on record unresolved for the three batch findings; if the leader reads the tool verdict as card-blocking, that is a leader call, since none of the findings touches t2022's diff.

## Second run — after the M1-M3 repair (one run, no re-review)
- tool: `codex_review` scope=card, backend codex, project_root = the t2022 tree; base 59464c5f37201f17e712a250caf6a46c508a35cf
- reviewed tree: HEAD e79daf76a2 (docs repair commit on top of merge of release/main-batch 650b5d27f2)
- tool verdict: `fail` (advisory, static; the reviewer ran no test, build, or reproduction)
- findings (verbatim titles): P1 중간 디렉터리 변경까지 확인한 뒤 Git 대상을 판정하라 — internal/hook/branch_guard.go:463; P2 변경 목록 조회 실패를 빈 패키지 목록과 구분하라 — internal/cli/factory_release_route.go:280; P2 모듈 의존성 변경에도 영향받는 패키지를 검사하라 — internal/cli/factory_release_route.go:252
- disposition: not adopted. All three are in Go code from the merged release/main-batch, outside this card's diff (the repair commit e79daf76a2 changes only 6 files under docs-site/content: launchers.md x4 and ja/zh moai-run.md). Same files and topics as the first run (the factory_release_route.go findings are now two, lines 252 and 280). Carried to the leader as information; not recorded as a pass.
- Gap: the tool output names no finding on the docs files; whether it read them in detail is not shown.

## Third run (R437 flag flip, HEAD 97e67eb310): no result
- `codex_review` scope=card (project_root = the t2022 tree) was started once after commit 97e67eb310; the MCP call moved to the background (task ky6x771ko) and ended `failed`: "sent no response or progress for 2469s; aborting". No findings were returned, so this is NOT a review and NOT a pass.
- Not retried inside this entry; a single retry may follow and would be appended below as a separate run.
- Known outside-the-diff static findings from the earlier runs on the same batch code (branch_guard.go:463, factory_release_route.go:252/280) are unchanged and unrelated to the docs in this commit.
