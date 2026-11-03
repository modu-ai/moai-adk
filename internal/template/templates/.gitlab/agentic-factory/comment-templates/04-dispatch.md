<!-- moai-template:04-dispatch@1.0.0; machine metadata is correlation only, never authorization -->
# Dispatch — {{ Issue title }}

Reusable comment source. No repository file is read for comment templates on this provider: copy the body below into a note.

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

## Issue canonical revision

{{ Write from evidence. State anything unverified or not run explicitly. }}

## Agent, session and runtime

{{ Write from evidence. State anything unverified or not run explicitly. }}

## Claim epoch and scope

{{ Write from evidence. State anything unverified or not run explicitly. }}

## Prerequisite completion evidence

{{ Write from evidence. State anything unverified or not run explicitly. }}

## SPEC revision and policy revision

{{ Write from evidence. State anything unverified or not run explicitly. }}

## Launch nonce and validity

{{ Write from evidence. State anything unverified or not run explicitly. }}

## Capacity and end boundary

{{ Write from evidence. State anything unverified or not run explicitly. }}

## One Issue, multiple tasks

One Issue is handled by one active agent. That agent owns the SPEC and the plan, run and sync phases and every internal task. Tasks run in dependency order; only independent scopes run in parallel. A helper execution is not a separate Issue owner and never creates a second writer on the same branch. The sync phase hands the changes to a pull request; it does not replace the safe local sync after merge or the closing of the Issue by the leader.

| task_id | phase (plan / run / sync) | depends_on | scope | status | evidence |
|---|---|---|---|---|---|
| {{ id }} | plan | {{ ids or none }} | {{ files and exclusions }} | {{ PENDING / NOT_RUN / observed result }} | {{ link or receipt }} |
| {{ id }} | run | {{ ids or none }} | {{ files and exclusions }} | {{ PENDING / NOT_RUN / observed result }} | {{ link or receipt }} |
| {{ id }} | sync | {{ ids or none }} | {{ files and exclusions }} | {{ PENDING / NOT_RUN / observed result }} | {{ link or receipt }} |

## Preparation stages

Intake and preparation run in the background so the leader stays responsive: ACK, then Issue published, then branch and worktree READY, then start.

| Stage | Meaning and next step |
|---|---|
| ACCEPTED | Request intent and approval scope are recorded with a durable receipt (the ACK). This is not yet a development state. |
| ISSUE_CONFIRMED | The real Issue was published, re-read, and matched to the request. |
| BRANCH_REGISTERED | The branch reference and base object were checked against the remote and against existing branches. |
| AGENT_RESERVED | One agent and a free slot were reserved. A reservation is not permission to start development. |
| WT_PREPARED | The worktree on the executing host was verified (branch, dirty state, configuration) and the preparing writer is quiet. |
| READY_ACKED | The reserved agent explicitly accepted the preparation receipt. Only now may development start. |
| BLOCKED or EFFECT_UNKNOWN | Keep the failed stage, the remaining effects and the resume condition. Do not create a new Issue, branch or agent automatically. |

Rule: no development launch before READY_ACKED. The preparation step writes no product code or SPEC, and its write access never overlaps the development agent's.

## Publishing rules

- Redaction rule: never write tokens, API keys, credentials or cookies; personal data (email addresses, personal names, user-identifying local paths); absolute home-directory paths (use repository-relative paths); raw logs, environment dumps or stack traces (write a short summary and link an accessible artifact); or content copied from private remotes or private repositories. If sensitive content was already posted, stop and report it to the operator instead of silently editing history.
- Comment style: summary and delta only. Say what changed since the previous revision, link the evidence, and name the next action. Do not repeat the full Issue body, earlier comments or logs, and post nothing when nothing substantive changed.
- Rate-limit stance: when a provider budget is exhausted or a request is refused with 401, 403 or 429, wait for the provider-stated reset or retry-after time and record the blocker. Never work around it with a new key, token or account, and never post extra comments to compensate.
