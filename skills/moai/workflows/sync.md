# SYNC

## Input

A candidate that passed, the latest base, document and Issue revisions, and the allowed effects.

## Output

A document diff, an Issue report, an integration receipt, and a closure. The receipt's field contract is below. In 2026.11.0-rc.1 it is produced by this procedure with `git` and `gh`, not by a dedicated command, and it is carried in the PR body and the Issue report rather than in a file; every field that was not measured is written as `not_run`. Per-file apply receipts are a different record and stay under the project state directory's `receipts` folder.

## States

`ready → reconciling → document_proposed → authorized → publishing → confirmed → closed`. A run stops in `needs_rebase` (the candidate does not contain the latest base), `needs_amendment` (the SPEC itself must change), `conflict`, `blocked` (a gate, permission or independent verdict is missing) or `publish_unknown` (an external write ended without a confirmed result). A stop state is reported as it is; it is never rounded up to a pass.

## Procedure

1. **Base.** Fetch the integration branch and test whether the candidate contains its tip:
   `git fetch origin v2026.11.0` then `git merge-base --is-ancestor "$(git rev-parse origin/v2026.11.0)" <candidate_sha>`. Exit 0: up to date. Exit 1: `needs_rebase` — merge the base into the candidate branch (never rebase a published branch), resolve conflicts keeping both sides' intent, and treat the result as a new candidate. Any other exit is an error, not a result. A PASS measured on an earlier tree is never copied to the new one.
2. **Verify the exact tree.** On the candidate tree that will be pushed, run the repository's full check set and record each command with its exit code (the RC1 set: `cargo test --offline --locked`, `cargo clippy --offline --locked --all-targets -- -D warnings`, `cargo fmt --check`, `python3 -m unittest discover -s tests -p '*_test.py'`, `git diff --check`, `git diff --cached --check`). A check that did not run is `not_run`; a check that could not run is `blocked` with the reason.
3. **Classify differences** between the implementation and the approved SPEC: an *omission* (the SPEC asks for something the candidate lacks: implement it or report it), an *amendment* (the requirement itself has to change: write an amendment record with the original SPEC digest, the change and "re-approval required", and never lower an acceptance criterion quietly), or a *document mismatch* (only the text is behind: diff the managed documents, preserving user edits).
4. **Effects, one permission each.** Push, PR creation, Issue comment, merge and release are separate effects. Each has its own permission check and operation ID, and approving one never approves another. Record each external effect with a stable operation marker and an outbox: `effect-intend` before the send, `effect-confirm` after the response. The marker's project slot is the repository id (`gh api repos/<owner>/<repo> --jq .id`), never the repository name, which a rename or transfer changes. When a forge is connected for the first time and local TODOs are moved over, preview the list, the visibility scope and the secret-free body, and promote only the items the user selects.
5. **Pull request.** Render the body with `moai-cli pr-body-render` and open it with `gh pr create --base v2026.11.0 --body-file <file>`; the base is always named, because without `--base` gh uses the repository default branch. The body carries `Refs #<n>` (never `Closes`, `Fixes` or `Resolves`: on a non-default base GitHub ignores them, so closing an Issue is the integration report's job), the exact base and candidate SHAs, the checks with exit codes, the independent review slot, what was not run, and the operation marker `<!-- moai-op pr_create:<repository_id>:<task_key>:<epoch> -->` (`docs/pr-contract.md`; for example `<!-- moai-op pr_create:123456789:08-A:1 -->`). Check the opened PR: `gh pr view <n> --json baseRefName,headRefOid,body` must show `v2026.11.0`, the candidate SHA, and no closing keyword before `#`.
6. **CI evidence.** Query the checks recorded for the exact candidate: `gh api repos/<owner>/<repo>/commits/<candidate_sha>/check-runs --jq '.check_runs[] | {name,status,conclusion,head_sha}'`. A `pull_request` run builds the PR merge ref, so record both the head SHA and the merge SHA the run reports. When a merge conflict kept the run from starting, the CI field is `not_run`, not a pass.
7. **Independent review.** Hand a separate session or subagent the approved SPEC digest, the exact candidate SHA and diff, the check results and the questions to answer. Without a verdict from a separate context the state is `blocked` and nothing after the PR is published. A verdict binds only the SHA it names.
8. **Integration (the integration owner, not the author).** Read the verdict's first line, confirm the PR head equals the audited SHA and the merge base equals the base tip (otherwise update the branch and re-audit), post the `moai/audit` status on that head, then merge pinned to it: `gh pr merge <n> --auto --merge --match-head-commit <candidate_sha>`. A new push invalidates the status. Right before merging, run step 1 again; a moved base makes the receipt stale.
9. **Reconcile, then report.** Read every external result back (PR state and head, comment text, merge commit) before announcing it. After a timeout, do not repeat the write: run `effects-reconcile`, which looks the operation marker up and moves the uncertain effect to confirmed or not_performed without ever re-creating it; only a not_performed effect may be attempted again, as a new intent. Report successes, failures, `not_run`, `blocked`, the remaining work and artifact retention.

## Integration receipt fields

One JSON object, `schema_version` 1: `task_id`, `attempt`, `base_sha`, `candidate_sha`, `merge_base`, `spec_digest` (or `null` with "unapproved draft"), `audit_verdict_ref`, `local_checks[]` (`argv`, `exit_code`, `started`, `ended`), `ci_checks[]` (`name`, `head_sha`, `merge_sha`, `status`, `conclusion`, or `not_run`), `effects[]` (`kind`, `operation_id`, `status`), `status` (one of the states above). Generating it twice from the same inputs gives the same bytes and repeats no effect.

## Exceptions and Permissions

SYNC does not merge, release, change branch protection or rulesets, or approve deployment by itself. Repository settings are admin effects and stay with the user.

## Completion Check

Bind the output to the task and the exact snapshot. Never report a verification or external effect that was not performed as complete.
