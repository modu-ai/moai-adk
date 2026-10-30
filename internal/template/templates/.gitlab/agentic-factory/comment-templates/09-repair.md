<!-- moai-template:09-repair@1.0.0; machine metadata is correlation only, never authorization -->
# Repair — {{ Issue title }}

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

## Failure family and cumulative cycle count

{{ Write from evidence. State anything unverified or not run explicitly. }}

## Classification: first occurrence, recurrence, patch-induced or environment

{{ Write from evidence. State anything unverified or not run explicitly. }}

## Reproduction input and cause hypothesis

{{ Write from evidence. State anything unverified or not run explicitly. }}

## Invariants and impact on all callers

{{ Write from evidence. State anything unverified or not run explicitly. }}

## Approved integrated fix scope

{{ Write from evidence. State anything unverified or not run explicitly. }}

## Normal, boundary, error and resource-limit tests

{{ Write from evidence. State anything unverified or not run explicitly. }}

## SHA before and after, with evidence

{{ Write from evidence. State anything unverified or not run explicitly. }}

## Remaining budget and HOLD conditions

{{ Write from evidence. State anything unverified or not run explicitly. }}

## Execution logic diagram

The Mermaid source below is a structure template. Complete it with the real branches, failures and resume conditions of this work before it is rendered. A platform-rendered Mermaid fence is not the pinned renderer's output.

```mermaid
flowchart TD
 N0[Record failure family]
 N1[Cause and full impact]
 N2[Approved repair]
 N3[Normal boundary and error tests]
 N4[Same-head audit or diagnostic HOLD]
 N0 --> N1
 N1 --> N2
 N2 --> N3
 N3 --> N4
```

- source SHA-256: {{ hash of the final Mermaid source }}
- Render status: {{ PASS, or RENDER_BLOCKED; never record NOT_RUN as PASS }}
- Pele version / bundle hash: {{ pinned version and bundle hash }}
- light / dark SVG SHA-256 and URI: {{ actual render results }}
- Pele SVG: not generated (renderer unavailable) — see GAP-PELE-SVG
- Render, sanitize, contrast and layout QA: {{ each result; NOT_RUN is never PASS }}
- Text description: {{ flow, failures, approvals, recovery and completion conditions }}

## Publishing rules

- Redaction rule: never write tokens, API keys, credentials or cookies; personal data (email addresses, personal names, user-identifying local paths); absolute home-directory paths (use repository-relative paths); raw logs, environment dumps or stack traces (write a short summary and link an accessible artifact); or content copied from private remotes or private repositories. If sensitive content was already posted, stop and report it to the operator instead of silently editing history.
- Comment style: summary and delta only. Say what changed since the previous revision, link the evidence, and name the next action. Do not repeat the full Issue body, earlier comments or logs, and post nothing when nothing substantive changed.
