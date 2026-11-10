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

The review package and the way a verdict reaches integration are described in the SYNC workflow (procedure steps 7 and 8).

Reserve before starting. Before an audit, a re-audit or a delta audit, read the pull request's comments and post `moai-audit-claim: lane=<lane> head=<40-hex head SHA>` only when no other lane's line names the same head — gate the post on that check in one command, never as two unconditional steps — then read the comments back. For one head, the line with the lowest comment id wins; a new head needs a new line. A session whose claim token (`Moai-Claim`) appears in a commit between base and head is an author of that head and does not reserve it.

Record interim results where the next auditor can find them: a partial check or provisional finding goes on the pull request as a comment without a verdict line, so it survives a reassignment.

Judge behaviour from the code and the test that pins it. A module comment, a README or a design note is a hypothesis until code and test agree with it.

Make every mutation real: after applying it, `git diff -U0` shows exactly the intended change on a code line (not a comment or another occurrence of the same text); record that one-line diff with the result, and confirm `git diff --quiet HEAD` after restoring, before the next mutation. A change that narrows what a shared judgement matches carries, in the same commit, literal tests in the miss direction and a before/after table of the judgement over the same inputs. Take grammar tables of external tools (options, subcommands) from the installed version's own documentation and note the version.

Read revisions safely: write a revision path as `"${rev}:path"` (zsh reads `$rev:s…` as a modifier), compare full SHAs or `origin/<branch>` only, never a bare `HEAD` of a shared checkout, and run audit commands inside a detached audit worktree at the exact head. In a merge-only delta, compare the candidate's change with the audited one file by file; when a whole-diff patch-id differs because the base also changed a file, compare that file's added and removed line sets and name the files that differ only in context. Re-read the pull request head and the base tip right before posting a verdict or reporting a CI failure; if either moved, the verdict names the new values or waits.

## Exceptions and Permissions

Self-assessment in the same context and a judge's recommendation do not replace the required independent audit. If a reviewer cannot be created or the evidence is insufficient, the result is blocked or unknown.

## Completion Check

Bind the output to the task and the exact snapshot. Never report a verification or external effect that was not performed as complete.
