---
id: SPEC-20261118-skill-audit-followups
title: "skill audit follow-ups: session trailer when missing, auditor worktree, trailer-check in CI"
status: draft
revision: 1
created: 2026-11-18
---

# skill audit follow-ups: session trailer when missing, auditor worktree, trailer-check in CI

## Intent

Three non-blocking audit findings in `skills/moai` guidance: `internal/sdd` item 10 reads as if every commit carries
`Moai-Session` (SDD-5 audit N-a); the audit workflow does not say why the auditor works outside the implementer's
worktree (SDD-6 audit N2); `internal/commit-trailers` states `trailer-check --branch` for a detached checkout but no test
holds it (SDD-6 audit N1).

## Requirements

- REQ-1: The `internal/sdd` shared-places item shall say that `Moai-Session` is written only when the host gives a session id, and point to `internal/commit-trailers` for the case without one.
- REQ-2: The audit workflow shall say that the auditor works in its own worktree or a clone inside it, never in the implementer's, because the worktree key of `Moai-Session` tells author and auditor apart.
- REQ-3: The `internal/commit-trailers` reference shall say that `moai trailer-check` in CI or a detached checkout takes `--branch <name>`.

## Acceptance

- AC-1 (REQ-1): the item-10 line holds "when the host gives a session id" and "otherwise `Moai-Spec` only".
  verify: `cargo test --test sdd_guidance the_reference_states_the_rules_the_user_decided`
- AC-2 (REQ-2): `workflows/audit.md` holds the sentence, naming the worktree key and `internal/commit-trailers`.
  verify: `cargo test --test mode_workflow_steps the_auditor_works_outside_the_implementers_worktree`
- AC-3 (REQ-3): the `trailer-check` item of `internal/commit-trailers` names `--branch <name>` for a detached checkout and CI.
  verify: `cargo test --test sdd_guidance the_trailer_reference_names_the_branch_flag_for_ci`

## Scope

- write: `skills/moai/references/internal/sdd.md`, `skills/moai/workflows/audit.md`,
  `skills/moai/references/source-manifest.json`, `tests/sdd_guidance.rs`, `tests/mode_workflow_steps.rs`,
  `.moai/specs/SPEC-20261118-skill-audit-followups/spec.md`
- read-only: `skills/moai/references/internal/commit-trailers.md` (REQ-3 already holds; only a test is added)

## Plan

1. Tests first (RED for REQ-1 and REQ-2; REQ-3 is already written, so its test pins it).
2. One sentence each in `internal/sdd` and `workflows/audit.md`; manifest write and check.
3. Checks and the trailers from `moai trailers --spec`.

## Result

- AC-1: met — `tests/sdd_guidance.rs:24` `the_reference_states_the_rules_the_user_decided` (phrases at :58); sentence at `skills/moai/references/internal/sdd.md:16`.
- AC-2: met — `tests/mode_workflow_steps.rs:745` `the_auditor_works_outside_the_implementers_worktree`; sentence at `skills/moai/workflows/audit.md:49`.
- AC-3: met — `tests/sdd_guidance.rs:69` `the_trailer_reference_names_the_branch_flag_for_ci`; sentence at `skills/moai/references/internal/commit-trailers.md:17` (unchanged).
- N2 (a folder in no git worktree) is left to `internal/commit-trailers` item 2.
