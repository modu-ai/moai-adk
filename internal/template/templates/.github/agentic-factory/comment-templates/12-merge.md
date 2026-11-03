<!-- moai-template:12-merge@1.0.0; machine metadata is correlation only, never authorization -->
# Merge — {{ Issue title }}

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

## Pull or merge request URL and trusted author

{{ Write from evidence. State anything unverified or not run explicitly. }}

## Latest head, base and protection revision

{{ Write from evidence. State anything unverified or not run explicitly. }}

## Required checks on the current head

{{ Write from evidence. State anything unverified or not run explicitly. }}

## Review and independent audit PASS receipt

{{ Write from evidence. State anything unverified or not run explicitly. }}

## Writer and coordinator handoff

{{ Write from evidence. State anything unverified or not run explicitly. }}

## Evidence of no HOLD and no unresolved threads

{{ Write from evidence. State anything unverified or not run explicitly. }}

## Conditional merge effect id

{{ Write from evidence. State anything unverified or not run explicitly. }}

## Lookup after an unknown response, and the actual merge SHA

{{ Write from evidence. State anything unverified or not run explicitly. }}

## Closure responsibility

The leader performs the merge only when every item above is observed on the current head. A merge does not by itself prove release or acceptance.

## Publishing rules

- Redaction rule: never write tokens, API keys, credentials or cookies; personal data (email addresses, personal names, user-identifying local paths); absolute home-directory paths (use repository-relative paths); raw logs, environment dumps or stack traces (write a short summary and link an accessible artifact); or content copied from private remotes or private repositories. If sensitive content was already posted, stop and report it to the operator instead of silently editing history.
- Comment style: summary and delta only. Say what changed since the previous revision, link the evidence, and name the next action. Do not repeat the full Issue body, earlier comments or logs, and post nothing when nothing substantive changed.
