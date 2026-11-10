# AUDIT

## Input

The approved contract, the exact diff, runner receipts (the internal `audit-reduce` operation re-runs the verification and uses only its own receipt; storing verdicts is not implemented in 2026.11.0-rc.1), and findings.

## Output

Findings from an independent context and an evidence-bound verdict.

## Procedure

1. Give only the necessary review package to a host-native context kept separate from the implementation: the approved SPEC digest, the exact candidate commit and diff, the verification commands, and an instruction to stay read-only.
2. Cross-check sources, digests, test traces, unrun checks, and the allowed scope directly.
3. Record functional, security, consistency, and maintainability findings, with the conditions for fixing each.
4. Have a deterministic admission step confirm the must-pass conditions and their evidence.
5. If the reviewer changes code, verify the result again as a new candidate.
6. Which checks the auditor runs locally and when an exact-head CI run may be cited follow the check-run rules in the VERIFY workflow (section "Check-run rules"); a cited run id appears in the verdict.

The review package and the way a verdict reaches integration are described in the SYNC workflow (procedure steps 7 and 8).

Reserve before starting. Before an audit, a re-audit or a delta audit, read the pull request's comments and post `moai-audit-claim: lane=<lane> head=<40-hex head SHA>` only when no other lane's line names the same head — gate the post on that check in one command, never as two unconditional steps — then read the comments back. For one head, the line with the lowest comment id wins; a new head needs a new line. A session whose claim token (`Moai-Claim`) appears in a commit between base and head is an author of that head and does not reserve it.

Record interim results where the next auditor can find them: a partial check or provisional finding goes on the pull request as a comment without a verdict line, so it survives a reassignment.

Judge behaviour from the code and the test that pins it. A module comment, a README or a design note is a hypothesis until code and test agree with it.

Make every mutation real: after applying it, `git diff -U0` shows exactly the intended change on a code line (not a comment or another occurrence of the same text); record that one-line diff with the result, and confirm `git diff --quiet HEAD` after restoring, before the next mutation. A change that narrows what a shared judgement matches carries, in the same commit, literal tests in the miss direction and a before/after table of the judgement over the same inputs. Take grammar tables of external tools (options, subcommands) from the installed version's own documentation and note the version.

Read revisions safely: write a revision path as `"${rev}:path"` (zsh reads `$rev:s…` as a modifier), compare full SHAs or `origin/<branch>` only, never a bare `HEAD` of a shared checkout, and run audit commands inside a detached audit worktree at the exact head. In a merge-only delta, compare the candidate's change with the audited one file by file; when a whole-diff patch-id differs because the base also changed a file, compare that file's added and removed line sets and name the files that differ only in context. Re-read the pull request head and the base tip right before posting a verdict or reporting a CI failure; if either moved, the verdict names the new values or waits.

## 재감사 PR 본문 검사

PR 본문을 UTF-8 `BODY.md`로 저장하고 `python3 -I tools/check_pr_self_audit.py --reaudit --file BODY.md`를 실행한다. 표준 입력도 지원한다: `python3 -I tools/check_pr_self_audit.py --reaudit < BODY.md`. 재감사는 반드시 `--reaudit`를 명시하며, 템플릿 문구나 claim epoch로 자동 추론하지 않는다. 최초 감사에서는 재감사 전용 두 필드를 삭제하거나 생략하고 기존 명령을 사용한다.

본문의 정확한 `## 독립 review` 절 안(다음 Markdown 제목 전)에 다음 두 줄을 각각 정확히 한 번 적는다. 수리 확인 범위는 F1의 수리 경로·경계와 영향받은 범위를 실제 내용으로 채운다.

```text
- 수리 확인 범위: F1 수리 경로·경계와 영향받은 범위
- 새 지적 분류: 없음
```

새 지적 분류는 `없음`, `차단: 구체적 이유`, `후속 Issue #123: 구체적 이유` 중 하나다. 후속 Issue 번호는 실제 양의 ASCII 정수로 바꾸고 차단·후속 분류에는 구체적 이유를 적는다. 계약 위반·데이터 손상·보안 문제는 차단하며, 그 밖의 새 지적은 후속 Issue로 분류한다.

누락·중복·빈 값·구두점만 있는 값과 `TODO`·`TBD`·`not_run`·`<...>` 같은 미완성 값은 실패한다. 이 검사는 텍스트 완전성만 확인한다. 거짓이지만 형식이 맞는 내용의 진실성, 실제 수리, 계약 준수나 원격 Issue 존재를 증명하지 않는다. 네트워크 요청이나 댓글 게시·Issue 생성 등 외부 효과는 없다. 기존 판정 첫 줄 형식, 판정·댓글 필드 및 CI/check-run 규칙은 그대로 유지한다.

## Exceptions and Permissions

Self-assessment in the same context and a judge's recommendation do not replace the required independent audit. If a reviewer cannot be created or the evidence is insufficient, the result is blocked or unknown.

## Completion Check

Bind the output to the task and the exact snapshot. Never report a verification or external effect that was not performed as complete.
