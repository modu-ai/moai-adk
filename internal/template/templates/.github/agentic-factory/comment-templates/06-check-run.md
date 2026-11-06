<!-- moai-template:06-check-run@1.0.0; machine metadata is correlation only, never authorization -->
# Check Run — {{ Issue title }}

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

## Target head, base and tree state

{{ Write from evidence. State anything unverified or not run explicitly. }}

## Check name and original run URL

{{ Write from evidence. State anything unverified or not run explicitly. }}

## Command or workflow revision

{{ Write from evidence. State anything unverified or not run explicitly. }}

## Environment and input digest

{{ Write from evidence. State anything unverified or not run explicitly. }}

## Start and end time and exit status

{{ Write from evidence. State anything unverified or not run explicitly. }}

## Success, failure, skipped, cancelled and missing are different results

{{ Write from evidence. State anything unverified or not run explicitly. }}

## Log summary and preserved original

{{ Write from evidence. State anything unverified or not run explicitly. }}

## Publishing rules

- Redaction rule: never write tokens, API keys, credentials or cookies; personal data (email addresses, personal names, user-identifying local paths); absolute home-directory paths (use repository-relative paths); raw logs, environment dumps or stack traces (write a short summary and link an accessible artifact); or content copied from private remotes or private repositories. If sensitive content was already posted, stop and report it to the operator instead of silently editing history.
- Comment style: summary and delta only. Say what changed since the previous revision, link the evidence, and name the next action. Do not repeat the full Issue body, earlier comments or logs, and post nothing when nothing substantive changed.
