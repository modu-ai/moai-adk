# SYNC

## Input

A candidate that passed, the latest base, document and Issue revisions, and the allowed effects.

## Output

A document diff, an Issue report, an integration receipt (planned; not implemented in 2026.11.0-rc.1), and a closure.

## Procedure

1. Confirm whether the candidate must be re-integrated onto the latest base and whether the exact candidate needs re-verification.
2. Classify differences between the implementation and the SPEC as omissions, amendments, or document mismatches.
3. Diff only managed documents, and preserve user edits.
4. Confirm the target and permission for each of push, merge, report, and release separately. When a forge is connected for the first time and local TODOs are moved over, preview the list, the visibility scope, and the secret-free body, and promote only the items the user selects to Issues. Open a pull request from the body `moai-cli pr-body-render` produces (`gh pr create --body-file`): it carries `Refs #<n>`, the exact base and candidate SHAs, the checks run, the independent review slot and what was not run, and never a closing keyword (`docs/pr-contract.md`).
5. Record external effects with a stable operation marker and an outbox (`effect-intend` before the send, `effect-confirm` after the response), then query the actual result again (`effects-reconcile` moves an uncertain effect to confirmed or not_performed and never re-creates it).
6. Report successes, failures, what was not run, the remaining work, and artifact retention details.

## Exceptions and Permissions

Right after a timeout, do not post again; reconcile instead. SYNC does not include automatic deployment or release approval.

## Completion Check

Bind the output to the task and the exact snapshot. Never report a verification or external effect that was not performed as complete.
