# SYNC

## Input

A candidate that passed the checks of RUN, its SPEC, the base its checks ran on and the lane's receipt.

## Output

The SPEC's `## Result` and status, a document and `@cc:` comment diff, and a commit on the lane's branch. The work ledger is the project's card ledger, read and written with `moai card`; Issues are not the work ledger. A check that was not measured is `not_run`.

## States

`ready → reconciling → document_proposed → authorized → publishing → confirmed → closed`. A run stops in `needs_rebase` (the candidate does not contain the latest base), `needs_amendment` (the SPEC itself must change), `conflict`, `blocked` (a gate or independent verdict is missing) or `publish_unknown`. A stop state is reported as it is; it is never rounded up to a pass.

## Before the audit

The order of the stages is the one `internal/sdd` gives: SYNC writes, AUDIT checks. On the candidate branch, before the audit, SYNC makes the SPEC, the documents and the `@cc:` comments match the code: it writes `## Result` from the VERIFY results (each AC met with its evidence, withdrawn with the reason, or what is missing), sets the status, fixes the document mismatches of step 3, keeps the `@cc:` comments true to the code and adds those the change calls for (agents and tools add and link them automatically), checks the SPEC and commits (steps 5 and 6). A requirement without code or tests, or a change outside the SPEC's Scope, is reported and stops the step; it is never written as met. Only after the landing, which closes the card: the SPEC is archived and the release note written.

## Procedure

1. **Base.** Record the base the lane's checks ran on. A lane fetches nothing, never chases the integration tip and never merges a moved tip (`flows/next-flow`); the lander deals with a moved tip once, at landing.
2. **Cite the receipt.** SYNC runs no check set. Take each check and its exit code from the lane's receipt (step 7, the receipt step, of the RUN workflow); a check without a result is `not_run`, a receipt of another head `stale`.
3. **Classify differences** between the implementation and the approved SPEC: an *omission* (the SPEC asks for something the candidate lacks: implement it or report it), an *amendment* (the requirement itself has to change: write an amendment record with the original SPEC digest, the change and "re-approval required", and never lower an acceptance criterion quietly), or a *document mismatch* (only the text is behind: diff the managed documents, preserving user edits). Four further drift kinds are reported the same way when present — each as `(kind, ground, proposed_action)`, separately, and never resolved by weakening the approved SPEC (an approved intent change goes through a SPEC amendment with a new revision and a regression test; an unexplained difference is blocked for review):
   - `spec_oracle_conflict` — the SPEC wording and a test oracle contradict each other. Ground: quote both lines. Proposed action: a SPEC amendment decided by the host, never auto-edited.
   - `broken_reference` — a deletion or rename severed a reference (file path, Issue, release pull request or SPEC id, workflow name). Ground: the dangling target. Proposed action: re-point the reference or restore the target.
   - `done_but_stale_evidence` — an Issue or record claims completion while its cited evidence is missing or stale (the audited or receipt sha is no longer the current HEAD, or the cited check predates the latest change). Ground: the stale reference against the current state. Proposed action: re-run the check and re-attach the evidence.
   - `explicit_link_unreviewed` — an explicit link (References/Closes/same-as/supersession) exists whose meaning was never reconciled. Ground: the link and both ends. Proposed action: review and record the link's semantics before relying on it.
   An explicit REQ link is never, by itself, proof of meaning.
4. **Write `## Result`.** One `moai-docs` writes the files (see Base agents); for any user-facing document it reads `internal/doc-claims`.
5. **Check the SPEC.** Run `moai spec check` and fix what it reports.
6. **Commit** with the trailers `moai trailers --spec <SPEC id> --verified <kinds>` prints, `Moai-Spec` among them; `--verified` names only the kinds of check that ran, `none` when none did. A lane only commits: it pushes nothing and opens no release pull request.
7. **Hand over.** Quote the lane's receipt and the committed head in the card's note. Everything after the lane's commit (effects, landing, the release pull request, CI evidence, the independent review, the release merge, reconcile and the integration receipt) is the lander's, in `flows/landing-release`. When a release or release candidate is prepared, run `internal/versioning` `## At every release` first.

## Base agents

Skip this section when you are a base agent yourself: base agents never start other agents.

`moai` installs five base agents into the host's user folder: `moai-plan`, `moai-research`, `moai-impl`, `moai-audit` and `moai-docs`. When this route hands a step to one of them:

- Give it one block, `task`, `scope`, `base` and `refs`, plus `head` where the route's own text below asks for it, and nothing else; it assumes no conversation context.
- Decide on its first line, `RESULT <role> <ok|blocked|not_run> <target>`, then check the rest yourself: the files it lists and the commands and exit codes it reports. An agent's `ok` is not evidence.
- Start agents only from this session, one level deep.
- Use only what the host itself offers. Never stand in for a missing feature with extra sessions, folders or scripts: name the feature as not supported on this host, with the reason, and use the fallback given below.
- When the agent is not installed on this host, do the step in this session and say so.

SYNC hands the files of step 4 (the SPEC's `## Result` and status, the document fixes and the `@cc:` comments) to one `moai-docs` with the host's agent tool, with `scope` naming only those paths; do not write them yourself when the agent is installed. It starts after `moai-impl` has returned, runs in this worktree and never beside `moai-impl`, and you wait for its `RESULT` line before step 5. It writes files and reports, and never sends anything.

## Resume

Read the SPEC's `## Result`, `moai spec check` and the card's notes afresh, and continue at the first step of the Procedure whose result is not on the branch. Nothing already written or sent is written or sent again.

## Exceptions and Permissions

SYNC does not merge, release, change branch protection or rulesets, or approve deployment by itself. Repository settings are admin effects and stay with the user.

## Completion Check

Bind the output to the task and the exact snapshot. Never report a verification or external effect that was not performed as complete.
