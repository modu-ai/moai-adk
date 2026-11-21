---
id: SPEC-20261121-stage-order-test-wide
title: "The test that keeps the stage order in one place reads every skill file and every base agent definition"
status: draft
revision: 1
created: 2026-11-21
---

# The test that keeps the stage order in one place reads every skill file and every base agent definition

## Intent

`SPEC-20261121-stage-order-once` moved the full order PLAN → RUN → VERIFY → SYNC → AUDIT → merge into `references/internal/sdd.md` and left a pointer in the five other files. Its test, `stage_order_is_written_once` in `tests/sdd_guidance.rs`, finds a second copy only in the files it names: the route references, five internal references, the chain table and `SKILL.md`. The audit of that SPEC found the gap: a copy in `references/internal/review.md`, in any other reference, in a workflow that is not a route reference, or in an agent definition under `agents/` would pass. This SPEC widens the test to every markdown file of the skill package and to the five base agent definitions, so the next copy fails the test on the day it is written. It changes one test file and a SPEC.

## Non-goals

- Changing a skill text, an agent definition or the route table: the full order is already only in `references/internal/sdd.md` on the tip, so the wider test passes on it.
- A test that walks the repository: such a test would have to be listed in `tools/repo_scans.json`; this test reads the list of files the source manifest already keeps.
- The pointer test of the five pointer files, which stays as it is.

## Requirements

- REQ-1: The test `stage_order_is_written_once` shall read every markdown file that the skill's source manifest lists and find the full stage order in `references/internal/sdd.md` only.
- REQ-2: The test `stage_order_is_written_once` shall read each of the five base agent definitions under `agents/` and find no copy of the full stage order in them.
- REQ-3: The test shall take its list of skill files from the source manifest and not from a walk of a folder.
- REQ-4: The skill's source manifest shall match the files after the change.

## Acceptance

- AC-1 (REQ-1, REQ-2, REQ-3): the test reads the manifest's markdown files and the five agent definitions, and finds the order only in `references/internal/sdd.md`; a copy placed in `references/internal/review.md` and a copy placed in `agents/audit.md` each make it fail.
  verify: `cargo test --locked --offline --test sdd_guidance stage_order_is_written_once`
- AC-2 (REQ-3): the repository-scan list still holds: no walking test is added.
  verify: `python3 -I -m unittest tests.repo_scans_test`
- AC-3 (REQ-4): the manifest matches the package.
  verify: `python3 -I tools/source_manifest.py check`

## Scope

- write: `.moai/specs/SPEC-20261121-stage-order-test-wide/**`, `tests/sdd_guidance.rs`
- read-only: `skills/moai/references/source-manifest.json`, `skills/moai/references/internal/sdd.md`, `agents/*.md`, `tests/repo_scans_test.py`
- forbidden: a card, Issue or pull request number in the sentences this change adds, a walk of a folder in the test, and an edit of a file another card holds

## Plan

1. Show the gap first: put a copy of the full order into `references/internal/review.md` and run the narrow test of the tip (it passes: the mutant survives); the same copy in `agents/audit.md` survives too.
2. Widen the test: take the markdown files from the manifest's `files` keys, add the five agent definitions, assert the order only in `sdd.md`.
3. Run `sdd_guidance` without a name filter and the same two copies against the new test (both make it fail), then restore the files.
4. Run `tests.repo_scans_test` and `check affected`; ask for an independent audit; the implementer does not audit.

## Risks

- The manifest lists a file that no longer exists: the test reads each listed path with `read`, which panics with the path, so the failure names the file.
- A later SPEC adds an agent definition: the test names the five that exist today; a sixth is not read until its name is added, and the Result says so.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | The list comes from the source manifest | The manifest already names every package file and is checked on every change; a folder walk would need a repository-scan entry | Walk the folder and list the test in the scan table |
| D-2 | The five agent definitions are named in the test | They live outside `skills/moai/`, so the manifest does not list them | Read the folder and list the test in the scan table |

## Result

(written after the checks)
