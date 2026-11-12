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

Read revisions safely: write a revision path as `"${rev}:path"` (zsh reads `$rev:s…` as a modifier), compare full SHAs or `origin/<branch>` only, never a bare `HEAD` of a shared checkout, and run audit commands inside a detached audit worktree at the exact head. For the mechanical evidence of one two-parent merge, run `moai-cli audit-delta [PROJECT] --pass-head <audited-full-SHA> --tip <integrated-tip-full-SHA> --head <candidate-full-SHA>` (PROJECT defaults to `.`). It reads local commits only, requires the actual ordered parents `[pass_head, tip]`, and replays in an external temporary bare repository without changing the source, including its object database. Record the report: only `pure_merge` exits 0; `manual_delta`, `replay_conflict`, and `refused` return a computed `ok: true` report with exit 1. `ok` means computation, not AUDIT PASS; input/Git/I/O errors use the normal error envelope and exit 1, and usage errors exit 2. No text hunks alone does not prove tree equality (binary and mode-only changes are separate). Non-merges, octopus merges and mismatched/reversed parents are refused. Keep human content review, claim reservation, exact-head CI checks and verdict publication separate. Remote CI/PR mode, multiple-merge first-parent chains and patch-id comparison are not supported by this command: outside its one-merge scope, compare changes file by file manually; when a whole-diff patch-id differs because the base also changed a file, compare that file's added and removed line sets and name the files that differ only in context. Re-read the pull request head and the base tip right before posting a verdict or reporting a CI failure; if either moved, the verdict names the new values or waits.

## Re-audit PR body check

Save the PR body as UTF-8 `BODY.md` and run `python3 -I tools/check_pr_self_audit.py --reaudit --file BODY.md`. Standard input is also supported: `python3 -I tools/check_pr_self_audit.py --reaudit < BODY.md`. Re-audits must explicitly select `--reaudit`; template text and claim epochs never select it automatically. For an initial audit, omit the two re-audit-only fields and use the existing command.

Use exactly one visible `## 독립 review` section; multiple visible review sections are ambiguous and fail. Inside that section, before the next visible Markdown heading, include each of the following fields exactly once as visible text. HTML comments and fenced or indented code examples do not establish a section, end it or supply its fields. Fill the repair scope with the actual repaired path, boundary and affected scope for F1.

Comment delimiter text inside an inline code span (for example, `<!--`) remains visible explanation text, not an HTML comment opener. Code spans close with a backtick run of the same length as their opener; shorter or longer runs inside them are literal text. An unmatched backtick run does not hide an actual HTML comment, and comments outside a closed code span still cannot supply required fields.

```text
- 수리 확인 범위: F1 repaired path, boundary and affected scope
- 새 지적 분류: 없음
```

The finding classification is `없음` (none), `차단: concrete reason` (blocking), or `후속 Issue #123: concrete reason` (follow-up). Replace the follow-up number with the actual positive ASCII integer, and give a concrete reason for blocking or follow-up findings. Contract violations, data damage and security findings are blocking; other new findings belong in a follow-up Issue.

Missing or duplicate visible fields, empty or punctuation-only values, and unfinished markers such as `TODO`, `TBD`, `not_run` and `<path>` fail. Completed explanations may cite source paths such as `src/todo.rs`, inline code such as `ci: not_run`, and comparison operators; those references alone are not unfinished markers. This checks text completeness only: well-formed text does not prove its truth, an actual repair, contract compliance or the existence of the remote Issue. The checker makes no network requests and never posts comments or creates Issues. The existing verdict first-line format, verdict/comment fields and CI/check-run rules remain unchanged.

Bare markers followed only by punctuation or Markdown decoration also fail (for example, repair scope `TODO.` or `TBD-`, and finding classification `차단: TODO.` or `차단: TBD-`). This does not reject actual filenames such as `TODO.md` or `TBD-related.md`.

The #2388 marker repair preserves the reported audit inputs unchanged:

| Field value | Audit-reported exit before repair | Required exit after repair |
| --- | --- | --- |
| Repair scope `TODO.` | 0 | 1 |
| Repair scope `TBD-` | 0 | 1 |
| Finding classification `차단: TODO.` | 0 | 1 |
| Finding classification `차단: TBD-` | 0 | 1 |

## Exceptions and Permissions

Self-assessment in the same context and a judge's recommendation do not replace the required independent audit. If a reviewer cannot be created or the evidence is insufficient, the result is blocked or unknown.

## Completion Check

Bind the output to the task and the exact snapshot. Never report a verification or external effect that was not performed as complete.
