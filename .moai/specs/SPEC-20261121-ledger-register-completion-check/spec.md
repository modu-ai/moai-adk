---
id: SPEC-20261121-ledger-register-completion-check
title: "The ledger-register pick has a fixed place inside the completion check of the PROJECT route"
status: draft
revision: 1
created: 2026-11-21
---

# The ledger-register pick has a fixed place inside the completion check of the PROJECT route

## Intent

The install-flow survey read at local v2026.11.0 on 2026-11-21 found G13: `skills/moai/workflows/project.md` says the card-ledger creation "waits for its own pick after the read-only `moai ledger-register-plan <project>` preview" but never says where in the route that pick is asked (`project.md:44`). A session following the text can ask at a different moment each run or skip the pick, and the survey's flow tree shows the pick floating without a step position.

The leader recorded the user's HOW pick on 2026-11-20T23:28Z: the card-ledger registration pick has a fixed place inside the completion check of `/moai project`. This change states that place in the workflow text, keeps byte for byte the sentence a route test pins, and pins the new place with a structure test in the route's own test file.

## Non-goals

- Rust behaviour: the pick is the session's ask, already shaped by `internal/questions`; no command, flag or JSON field changes.
- The UPDATE route's ledger-register sentences (`skills/moai/workflows/update.md`): unchanged; their pins stay as they are.
- What registering the ledger writes under the real HOME, and the guarded confirmation of that write: unchanged (that behaviour is named, not altered, by this text).

## Requirements

- REQ-1: The work-ledger bullet shall keep the sentence that names the read-only `moai ledger-register-plan <project>` preview verbatim and state right after it that the pick's fixed place is the completion check.
- REQ-2: The Completion Check section shall carry one fixed paragraph that names the card-ledger registration pick with its position in the check's offer (after the plan steps and before the closing-commit offer of step 8) and the rules that the read-only preview is shown first and the apply runs only after the user picks it.
- REQ-3: The Completion Check paragraph shall state when the pick joins the offer (the project holds no registered card ledger or an earlier run left the pick out) and that a project whose ledger is already registered adds nothing while a cancelled or unanswered pick is reported as open.
- REQ-4: A structure test shall pin the bullet's fixed-place sentence and the Completion Check paragraph's keys while the pinned preview sentence stays byte for byte what the UPDATE route's test asserts.
- REQ-5: The manifest entry of `workflows/project.md` shall carry the sha256 of the changed file and a reason line that records this change without any card number.

## Acceptance

- AC-1 (REQ-1): the work-ledger bullet holds the pinned preview sentence and after it the fixed-place sentence naming the completion check · verify: `cargo test --test project_flow_closing`
- AC-2 (REQ-2, REQ-3): the Completion Check section holds the fixed paragraph with the position in the offer, the preview-first rule, the pick-gated apply, the offer condition and the registered-ledger and unanswered cases · verify: `cargo test --test project_flow_closing`
- AC-3 (REQ-4): the new test fails on the base and passes on the head · verify: run on the base head then on the card head, both recorded in the card report
- AC-4 (REQ-5): the manifest check passes and the route tests that read `project.md` still pass · verify: `python3 -I tools/source_manifest.py check` and `cargo test --test update_route`

## Scope

- write: `skills/moai/workflows/project.md`, `skills/moai/references/source-manifest.json`, `tests/project_flow_closing.rs`, `.moai/specs/SPEC-20261121-ledger-register-completion-check/**`
- read-only: `skills/moai/workflows/update.md`, `tests/update_route.rs`, `tools/source_manifest.py`
- forbidden: any card number in the shipped text, the real HOME, Rust source changes

## Plan

RED first: the pin test reads `project.md` as text, in the style of this test file, and fails on the base because neither the fixed-place sentence nor the Completion Check paragraph exists. GREEN edits the two places in `project.md`, refreshes the manifest sha256 with `tools/source_manifest.py`, and the new pin and the existing route pins pass together.

## Risks

- The pinned sentence of the work-ledger bullet is asserted verbatim by `tests/update_route.rs`; the edit keeps it byte for byte and the route test runs in the same change.
- The Completion Check paragraph instructs the session only (an item of the check's offer), like the rest of the section; it promises no new command or flag.
- The manifest gate compares sha256 entries with the merged tree, so the entry must be refreshed in the same change or the merge gate refuses the head.
