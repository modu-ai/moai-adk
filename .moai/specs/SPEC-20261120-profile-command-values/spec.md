---
id: SPEC-20261120-profile-command-values
title: "Profile validation accepts command values and every refusal names the field"
status: draft
revision: 1
created: 2026-11-20
---

## Intent

The profile interview records the project's check and build commands as `decisions` entries
(the question-bank rows `test_command` and `build_command`), and users write free-form
decision values. Validation rejects every `/` in those values today (`portable()` treats any
slash as a path), so ordinary commands such as `go test ./...` or `./gradlew build` cannot be
saved (survey report update-project-logic-2026-11-20.md §5 items 5–6). At the same time the
one validation error, "invalid or nonportable profile; no values were logged", names no
field, so a refused profile is unactionable: the user cannot tell which entry failed or why.

This SPEC splits decision values from path-like values and rebuilds validation as
per-field checks: command-like free text keeps its guardrails (no absolute or home-relative
path, no secret shape, no control characters) while relative slashes pass, and every refusal
names the field id — never the value.

This document and the plan note are committed first; the implementation waits until the
`src/profile.rs` reservation of the older card frees (leader reclaim planned at
2026-11-20T11:52Z), then lands as the second commit pair on this branch.

## Non-goals

- The skill question-bank text (`internal/interview.md`, `questions.md`) and the workflow
  guidance: they are the skill-text card's rewrite, not this SPEC.
- The secret-shape filter (`crate::secret_shaped`), the size and control-character limits,
  and the portable rule for path-like fields — all unchanged.
- `read_profile`'s managed-key filtering and the lossless-editing HOLD: untouched.

## Requirements

- REQ-1: While a `decisions` entry's value is a relative command (its slash characters never start the value), profile validation shall accept the entry. · failure: none — the acceptance is pinned by AC-1.
- REQ-2: While a `decisions` entry's value starts with a slash (an absolute path), holds a backslash or a drive-letter path prefix, starts with a tilde-slash (a home-relative path), holds a control character, is secret-shaped, or exceeds the value length limit, validation shall refuse the entry. · failure: none — the refusals are pinned by AC-2.
- REQ-3: While validation refuses a profile, the error shall name the failing field's id (`profile field <id>` for key-level fields, `decisions[<entry id>].value` and the same shape for `topic` and `scope` for entry fields) without the refused value or any other profile value. · failure: none — the wording is pinned by AC-3.
- REQ-4: While a path-like field (`purpose`, a `tools` or `development_methods` entry, `interview_revision`, `versioning.format`, or a `decisions` entry's `topic` or `scope`) holds a slash, a backslash, a tilde-slash or a drive-letter path prefix, validation shall refuse it exactly as today. · failure: none — the existing behaviour is pinned by AC-4.

## Acceptance

- AC-1 (REQ-1): a v2 profile whose decisions carry `test_command` = `go test ./...` and
  `build_command` = `./gradlew build` (relative slashes) reads, validates and plans.
  verify: `cargo test --locked --offline --test profile_command_values`
- AC-2 (REQ-2): the same entry shape with a value that starts with a slash, a
  home-relative value, a backslash-holding value and a secret-shaped value each refuse.
  verify: `cargo test --locked --offline --test profile_command_values`
- AC-3 (REQ-3): each refusal's error text contains the field id and no value substring of
  the refused profile.
  verify: `cargo test --locked --offline --test profile_command_values`
- AC-4 (REQ-4): a `/` in `purpose`, a tool name, `versioning.format` and a decisions
  `topic` or `scope` still refuse (the in-file unit tests keep passing unchanged).
  verify: `cargo test --locked --offline --lib profile`
- AC-5 (REQ-1, REQ-2, REQ-3, REQ-4): neighbours — the repository-wide scans and the SPEC
  format.
  verify: `python3 -m unittest tests.repo_scans_test`; the built binary's
  `spec check SPEC-20261120-profile-command-values`

## Scope

- write: `src/profile.rs`, `tests/profile_command_values.rs`, this SPEC file
- read-only: `src/secret_shape.rs`, the question-bank skill text, `tools/repo_scans.json`
- forbidden: changing the portable rule of path-like fields, echoing profile values in
  errors, editing the skill question-bank text

## Plan

1. Base: this branch (created from the integration tip). This commit carries the SPEC only.
2. RED: `tests/profile_command_values.rs` — the `go test ./...` profile (AC-1) fails first,
   because validation refuses the slash today.
3. `src/profile.rs`: add the free-text check for decision values (REQ-1/2) and rebuild
   `validate`'s boolean into per-field checks that return field-named errors (REQ-3);
   path-like fields keep `portable` (REQ-4).
4. The in-file unit tests keep passing (AC-4); add the AC-3 wording assertions.
5. Neighbours (AC-5) and the implementation result recorded here.

## Risks

- The rebuilt per-field errors change error strings that a consumer might match on: no
  caller matches on the message today (the in-file tests use `is_err`), and the gate reads
  `local-check` evidence, not profile errors.
- A decision value that is genuinely a path (a user writing `tools/` as a value) now passes
  where it failed: that is the intent — values are free text with guardrails, and the
  secret-shape and absolute-path refusals stay.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D1 | Decision values get their own check: relative slashes allowed, absolute or home-relative paths and secret shapes refused | Survey §5 item 5: commands like `go test ./...` must save; the survey's refusal examples keep absolute and home paths out | Fold values back into `portable` |
| D2 | Every refusal names the field id and never a value | Survey §5 item 6: an unnamed refusal is unactionable; values stay out of logs by the same rule as everywhere | Revert to the single boolean error |
| D3 | `topic` and `scope` and the key-level path-like fields keep `portable` | They are labels and paths, not commands; narrowing the change to values keeps the surface small | Extend the free-text check to them |

## Result

구현 에이전트가 채워요. 이 커밋은 SPEC·계획까지예요 — `src/profile.rs` 예약이 풀린 뒤 같은
가지의 다음 커밋 쌍으로 구현이 와요.
