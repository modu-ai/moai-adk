# Landing and release

When to read: when you, as the lander, land an audited head, open or merge a release pull request, record an external effect or write the integration receipt. A lane's part is in the SYNC workflow, not here.

## Base and `needs_rebase`

The base depends on the flow. In NeXT-Flow the integration branch is the local branch of `work_flow.integration_branch`: a lane fetches nothing, never chases its tip and never merges a moved tip (`flows/next-flow`). The lander deals with a moved tip once, at landing. The fetch below is for a release pull request, whose base is the remote integration branch.

Fetch the integration branch (`<integration>`: the target saved at first setup) and test whether the candidate contains its tip: `git fetch origin <integration>` then `git merge-base --is-ancestor "$(git rev-parse origin/<integration>)" <candidate_sha>`. Exit 0: up to date. Exit 1: `needs_rebase`; merge the base into the candidate branch (never rebase a published branch), resolve conflicts keeping both sides' intent, and treat the result as a new candidate. Any other exit is an error, not a result. A PASS measured on an earlier tree is never copied to the new one.

## States

`ready → reconciling → document_proposed → authorized → publishing → confirmed → closed`. A run stops in `needs_rebase` (the candidate does not contain the latest base), `needs_amendment` (the SPEC itself must change), `conflict`, `blocked` (a gate, permission or independent verdict is missing) or `publish_unknown` (an external write ended without a confirmed result). A stop state is reported as it is; it is never rounded up to a pass.

## Effects, one permission each

Push, the release pull request's creation, Issue comment, merge and release are separate effects. Each has its own permission check and operation ID, and approving one never approves another. Repository settings, branch protection and rulesets are admin effects and stay with the user; the lander never changes them or approves a deployment by itself.

When a release or release candidate is prepared, run `internal/versioning` `## At every release` before asking for the release effect.

Record each external effect with a stable operation marker and an outbox. Before the send run `moai effect-intend --task <task> --epoch <epoch> --operation <operation> --project-id <repository id> --local-id <local id> --attempt <n> --kind <kind> <project>` with the payload JSON on stdin; after the response run `moai effect-confirm --task <task> --operation-id <operation id> --observed-state <state> <project>` (`uncertain` marks an interrupted write). The marker's project slot is the repository id (`gh api repos/<owner>/<repo> --jq .id`), never the repository name, which a rename or transfer changes.

The effects stay with the lander's own session: a base agent writes files and reports, and never sends anything. The work ledger is the project's card ledger, read and written with `moai card` as the WORK workflow does; Issues are not the work ledger. Without the `moai` command there is no card ledger: the ledger is reported as unavailable (`blocked`) instead of falling back to another store.

## Landing

A card lands as the flow reference says (`flows/next-flow`). A lane only commits, on its own branch in its own worktree, and adds its ready note with its head and the base its checks ran on; it pushes nothing, opens no release pull request and merges no moved tip. Only after a passing verdict is recorded on the card (the audit contract, its section "Who records the verdict") does the lander land the head, once, with `moai card land <card> --head <40 hex> --by leader`, which moves the integration branch and closes the card in the same call. When the tip moved after the lane's base, the lander merges the head onto the tip once in a worktree of its own and lands that merge. Builds for local testing and installation are made from the integration branch tip; a broken tip is repaired with a revert commit, landed the same way.

## The release pull request

A release pull request is the only pull request, and none goes per card or per lane: only for a real deployment, the lander creates `release/v{head}.{yearweek}.{build}` from the integration branch and opens one release pull request to the remote's integration branch (full CI on Linux, macOS and Windows, independent audit by a session with no commits in the range).

Whoever opens the release pull request renders the body with `moai pr-body-render` and opens it with `gh pr create --base <integration> --body-file <file>`; the base is always named, because without `--base` gh uses the repository default branch. The body of the release pull request carries `Refs #<n>` (never `Closes`, `Fixes` or `Resolves`: on a non-default base GitHub ignores them, so closing an Issue is the integration report's job), the exact base and candidate SHAs, the checks with exit codes, the independent review slot, what was not run, and the operation marker `<!-- moai-op pr_create:<repository_id>:<task_key>:<epoch> -->` (for example `<!-- moai-op pr_create:123456789:08-A:1 -->`). Check the opened release pull request: `gh pr view <n> --json baseRefName,headRefOid,body` must show `<integration>`, the candidate SHA, and no closing keyword before `#`.

## CI evidence of the release pull request

Query the checks recorded for the exact candidate of the release pull request: `gh api repos/<owner>/<repo>/commits/<candidate_sha>/check-runs --jq '.check_runs[] | {name,status,conclusion,head_sha}'`. A `pull_request` run builds the merge ref of the release pull request, so record both the head SHA and the merge SHA the run reports. When a merge conflict kept the run from starting, the CI field is `not_run`, not a pass. A lane's head has no CI run to cite; a merge-only delta of the release pull request may cite its exact-head CI run only under `internal/check-rules`, section `## Citing a remote run`.

## Independent review of the release pull request

Hand a separate session or subagent the approved SPEC digest, the exact candidate SHA and diff, the check results and the questions to answer. The review goes to `moai-audit` as the AUDIT workflow describes, never to the agent or session that wrote the change. Without a verdict from a separate context the state is `blocked`: no card lands and no release pull request is published. A verdict binds only the SHA it names.

## Release pull request merge

The integration owner, not the author, merges the release pull request. A card is integrated by its landing; this section is for the release pull request alone. Read the verdict's first line and confirm the head of the release pull request equals the audited SHA. The release pull request merges when its current head has a successful `ci`, read from the head's check runs, and a non-author PASS that cites that run, with no conflict; merge it with the audited head as the expected head, so that a newer push is refused. A new push needs its own CI and a delta re-audit. Never merge the tip into it just to follow it: that invalidates the audit and buys nothing.

## Reconcile, then report

Read every external result back (the state and head of the release pull request, comment text, merge commit) before announcing it. After a timeout, do not repeat the write: run `moai effects-reconcile --task <task> <project>`, which looks the operation marker up and moves the uncertain effect to confirmed or not_performed without ever re-creating it; only a not_performed effect may be attempted again, as a new intent. An external write whose result was not read back is reconciled this way, never repeated. Report successes, failures, `not_run`, `blocked`, the remaining work and artifact retention.

## Integration receipt fields

The receipt is produced by this procedure with `git` and `moai card`, not by a dedicated command, and it is carried in a note on the card rather than in a file; every field that was not measured is written as `not_run`. Per-file apply receipts are a different record and stay under the project state directory's `receipts` folder.

One JSON object, `schema_version` 1: `task_id`, `attempt`, `base_sha`, `candidate_sha`, `merge_base`, `spec_digest` (or `null` with "unapproved draft"), `audit_verdict_ref`, `local_checks[]` (`argv`, `exit_code`, `started`, `ended`), `ci_checks[]` (`name`, `head_sha`, `merge_sha`, `status`, `conclusion`, or `not_run`), `effects[]` (`kind`, `operation_id`, `status`), `status` (one of the states above). Generating it twice from the same inputs gives the same bytes and repeats no effect.
