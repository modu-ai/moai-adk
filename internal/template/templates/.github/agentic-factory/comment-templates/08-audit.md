<!-- moai-template:08-audit@1.0.0; machine metadata is correlation only, never authorization -->
# Independent Audit — {{ Issue title }}

Reusable comment source. GitHub reads no repository file for comment templates: copy the body below into a comment.

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

## Selected auditor and independent session

{{ Write from evidence. State anything unverified or not run explicitly. }}

## Evidence that writer and permissions are independent

{{ Write from evidence. State anything unverified or not run explicitly. }}

## Same head and same SPEC and policy revision

{{ Write from evidence. State anything unverified or not run explicitly. }}

## Required checks and evidence inventory

{{ Write from evidence. State anything unverified or not run explicitly. }}

## PASS, FAIL or UNKNOWN per acceptance criterion

{{ Write from evidence. State anything unverified or not run explicitly. }}

## Omissions and counter-evidence

{{ Write from evidence. State anything unverified or not run explicitly. }}

## Final verdict and its limits

{{ Write from evidence. State anything unverified or not run explicitly. }}

## Audit receipt and next action

{{ Write from evidence. State anything unverified or not run explicitly. }}

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
