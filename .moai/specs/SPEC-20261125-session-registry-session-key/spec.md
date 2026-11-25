---
id: SPEC-20261125-session-registry-session-key
title: "Session registry keyed by session id, lane registration step, one name for both -a forms"
status: draft
revision: 1
created: 2026-11-25
---

# Session registry keyed by session id, lane registration step, one name for both -a forms

## Intent

User decision 2026-11-25 (decisions log line `11:27Z(11-25)`): a session registry entry is keyed by the host session id, so one session never owns two entries and a session of another name never replaces a live entry; a lane registers, renews and releases through its route text so the work-in-progress limit counts live lanes; and `-a` treats `agent-<name>` and `<project>-agent-<name>` as one name. The limit number (lanes times 2) does not change.

## Requirements

- REQ-1: When a session with a host session id registers, the registry shall write its entry to the file named by that session id, with the lane name kept inside the entry. · failure: REQ-2
- REQ-2 (REQ-1): IF another live session already holds the registered name, THEN the registry shall refuse the registration and leave both entries as they were.
- REQ-3: When a session registers again under another name, the registry shall replace that session's own entry so that one entry is left for the session. · failure: none — the replaced entry is the session's own, and no other session's entry is touched.
- REQ-4: The registry shall keep reading an entry whose file is named by the lane name. · failure: none — an entry that cannot be read is an error as it is today.
- REQ-5: When a lookup by name runs (`--show`, `--release`, `--lane-check` or a liveness lookup), the registry shall find the entry of either file form. · failure: none — a name with no entry answers `free` or `not_lane` as it does today.
- REQ-6: When the work-in-progress limit counts lanes, it shall count each live registered session once, whatever the form of its entry file. · failure: none — the multiple of 2 per lane and the limit off at 0 lanes stay as they are.
- REQ-7: The lane route text shall order the session registration, its renewal inside the lease and its release at the end as steps of the lane. · failure: none — a lane that skips a step is counted as before, by its live entries only.
- REQ-8: When `-a` picks a lane name, the launcher shall treat a live lock or a live lease of `agent-<name>` and of `<project>-agent-<name>` as one taken name. · failure: none — with every name taken the pick refuses as it does today.

## Acceptance

- AC-1 (REQ-1, REQ-2, REQ-3, REQ-4, REQ-5): a Rust test registers sessions with distinct host session ids in a temporary clone and shows the entry file name, the refusal of a name held by another live session with both entries unchanged, one entry after a rename, and the name lookups reading an entry of each form · verify: `cargo test --locked --test session_registry_session_key`
- AC-2 (REQ-6, REQ-7): a Rust test shows the limit counting a session once for two entry forms and the lane route text holding the register, renew and release steps, with the multiple unchanged · verify: `cargo test --locked --test session_registry_session_key`
- AC-3 (REQ-8): a Rust test shows the `-a` pick skipping a name taken in either form and returning the first name free in both · verify: `cargo test --locked --test lane_team_names`

## Scope

- write: `.moai/specs/SPEC-20261125-session-registry-session-key/**`
- write: `src/lane_registry.rs`
- write: `src/session.rs`
- write: `src/card/wip_limit.rs`
- write: `src/ops/session_register.rs`
- write: `skills/moai/workflows/work.md`
- write: `tests/session_registry_session_key.rs`
- write: `tests/session_registry_cli.rs`
- write: `tests/lane_team_names.rs`

## Plan

1. Write the failing tests: two live sessions of one name, a rename of one session, both name forms of `-a`, the limit count.
2. Key the entry file by the host session id, keep the name-keyed read path, and make the name lookups scan both forms.
3. Count distinct live sessions in the limit and make the `-a` pick read both name forms.
4. Add the register, renew and release steps to the lane route text.
