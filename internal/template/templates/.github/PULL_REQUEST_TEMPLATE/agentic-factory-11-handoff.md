<!-- moai-template:11-handoff@1.0.0; machine metadata is correlation only, never authorization -->
# Agent Handoff — {{ Issue title }}

| Common field | Value |
|---|---|
| actor / role / session_id | {{ authenticated actor, role and session }} |
| Issue / repo immutable ID | {{ canonical URL and immutable ID }} |
| template_version / policy_revision | 1.0.0 / {{ policy hash }} |
| source_event_id / source_revision | {{ source event and revision }} |
| effect_id / request_id / dedup_key | {{ publish intent ID and canonical payload hash }} |
| collector_cursor / adopted_revision / consumer_ack | {{ record collection, adoption and ACK separately; give the reason when not applicable }} |
| head_sha / base_sha / spec_hash | {{ value, or not applicable with the reason }} |
| observed_at / original_evidence | {{ RFC 3339 time with offset; original URL, log or receipt }} |
| current state / next action / owner | {{ observed state and the next trigger }} |

## Previous owner and next owner

{{ Write from evidence. State anything unverified or not run explicitly. }}

## Handoff target: Issue, head SHA and SPEC

{{ Write from evidence. State anything unverified or not run explicitly. }}

## Writer readiness receipt

{{ Write from evidence. State anything unverified or not run explicitly. }}

## Coordinator readiness receipt

{{ Write from evidence. State anything unverified or not run explicitly. }}

## Existing execution and remote-tool quiescence

{{ Write from evidence. State anything unverified or not run explicitly. }}

## Epoch and fence evidence

{{ Write from evidence. State anything unverified or not run explicitly. }}

## Uncommitted files and worktree preservation

The original worktree, branch and uncommitted changes are preserved, not deleted. The agent leaves the worktree safely and moves to a new Issue in a new worktree only after the handoff is accepted.

## Outstanding obligations and receiving ACK

{{ Write from evidence. State anything unverified or not run explicitly. }}

## Closure responsibility

Opening this pull request is not completion. After the handoff is accepted, the leader tracks checks, review, repair dispatch, merge, safe local sync and Issue closure, and the leader closes the Issue after acceptance evidence. Reference the Issue with a plain reference (Refs {{ issue }}) and do not use a closing keyword: a keyword would close the Issue at merge, before that evidence exists. Repairs return to the original Issue and branch under a single owner.

## One Issue, multiple tasks

One Issue is handled by one active agent. That agent owns the SPEC and the plan, run and sync phases and every internal task. Tasks run in dependency order; only independent scopes run in parallel. A helper execution is not a separate Issue owner and never creates a second writer on the same branch. The sync phase hands the changes to a pull request; it does not replace the safe local sync after merge or the closing of the Issue by the leader.

| task_id | phase (plan / run / sync) | depends_on | scope | status | evidence |
|---|---|---|---|---|---|
| {{ id }} | plan | {{ ids or none }} | {{ files and exclusions }} | {{ PENDING / NOT_RUN / observed result }} | {{ link or receipt }} |
| {{ id }} | run | {{ ids or none }} | {{ files and exclusions }} | {{ PENDING / NOT_RUN / observed result }} | {{ link or receipt }} |
| {{ id }} | sync | {{ ids or none }} | {{ files and exclusions }} | {{ PENDING / NOT_RUN / observed result }} | {{ link or receipt }} |

## Context ledger

The latest index lives in the Issue's canonical context block. A comment carries only the entries that changed: the item, the reason, the new revision, a short result and the link.

| entry_id | revision | kind | repo-relative path | commit-pinned link | artifact commit | artifact sha256 | target head | owner (role, session_id) | verification state | verified_at | accessibility |
|---|---|---|---|---|---|---|---|---|---|---|---|
| {{ stable id }} | {{ number }} | {{ spec / plan / decision / failure / regression / audit / pr_handoff / completion / other }} | {{ path }} | {{ link pinned to the full artifact commit }} | {{ full commit }} | {{ digest of the file bytes }} | {{ full head, or N/A with the reason }} | {{ role, immutable session_id }} | {{ PENDING / NOT_RUN / PASS / FAIL / BLOCKED / UNKNOWN / STALE / LINK_BLOCKED }} | {{ RFC 3339 time }} | {{ accessible / inaccessible-local / inaccessible-unpushed / unknown }} |

Ledger rules:

- A link is evidence only when it is a commit-pinned link: host, owner, repository, blob, the full artifact commit, then the repository-relative path. Branch-based links, short hashes and local file links are not evidence.
- Material that exists only on a local machine, or only in a commit that has not been pushed, is recorded with accessibility inaccessible-local or inaccessible-unpushed and verification state LINK_BLOCKED. It is never presented as team-accessible.
- An entry becomes STALE when the target head changes, the artifact sha256 no longer matches the file bytes, or the adopted contract revision changes. A STALE entry is kept with its original verdict preserved; it is never deleted.
- PENDING and NOT_RUN are not results, and an empty link is not completion evidence. A digest does not prove that the content is correct.
- The artifact commit (where the file lives) and the target head (the code the file describes or verifies) are separate fields.

## Publishing rules

- Redaction rule: never write tokens, API keys, credentials or cookies; personal data (email addresses, personal names, user-identifying local paths); absolute home-directory paths (use repository-relative paths); raw logs, environment dumps or stack traces (write a short summary and link an accessible artifact); or content copied from private remotes or private repositories. If sensitive content was already posted, stop and report it to the operator instead of silently editing history.
- Comment style: summary and delta only. Say what changed since the previous revision, link the evidence, and name the next action. Do not repeat the full Issue body, earlier comments or logs, and post nothing when nothing substantive changed.
