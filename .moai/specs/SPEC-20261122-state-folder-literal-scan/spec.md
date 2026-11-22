---
id: SPEC-20261122-state-folder-literal-scan
title: "A repository scan keeps tests from spelling the state folder name beyond a counted baseline"
status: draft
revision: 1
created: 2026-11-22
---

# A repository scan keeps tests from spelling the state folder name beyond a counted baseline

## Intent

The user-approved redesign (`~/.moai/reports/moai-adk/agentic-redesign-2026-11-22.md` §3E: "no state folder name literals in tests, enforced by a scan rule; old-layout fixtures only as exceptions") and audit-landing-process-review B2 record the cost: the path-root change took three landing rounds. Tests written meanwhile expected the earlier project folder for a fresh project, and one Python install test still did after the landing. The rule "tests do not spell the state folder name" was only text. This SPEC makes it a scan. The folder for a fresh project or HOME comes from the resolvers (`state_root`, `home_root`, or the Python helper that mirrors them). A literal folder name is right only where a test builds an earlier layout on purpose: an old-layout fixture, a Go-generation folder, a migration.

Today 175 Rust test files spell the earlier name on 1,363 lines, and 141 spell the new one. A text scan cannot tell a fixture from a fresh-project expectation line by line. So it ratchets per file: a counted baseline file lists every test file allowed to spell either name and how many times. The scan fails when a file outside the list spells a name, or when a listed file spells one more often than its baseline. A count may fall without editing the baseline, so fixing a test never touches a shared file. A new fixture that really needs the literal raises its count in the baseline, in the same change, where review sees it. Readers are lane authors and auditors; the failure text names the file, the name, both counts and the two ways out.

## Non-goals

- Rewriting the existing literals (that is per-file work, owned by the cards that touch those files).
- Scanning product code under `src/` (the resolvers are the rule there), documents, or `tests/fixtures/**` data trees.
- Telling a fixture line from an expectation line by meaning.

## Requirements

- REQ-1: The scan shall read every `.rs` and `.py` file under `tests/` outside `tests/fixtures/`, and count per file the earlier folder name and the new folder name where each stands as a whole path segment.
- REQ-2: If a scanned file spells either name and is not in the baseline file, then the scan shall fail and name that file.
- REQ-3: If a scanned file spells a name more often than its baseline count, then the scan shall fail and name the file, the name and both counts.
- REQ-4: The scan shall pass a file whose counts are at or below its baseline counts.
- REQ-5: The scan shall run as a Rust test listed in `tools/repo_scans.json`, so a change to any test file selects it.
- REQ-6: Where `MOAI_STATE_LITERALS_WRITE=1` is set, the scan shall write the baseline file from the current counts.

## Acceptance

- AC-1 (REQ-1, REQ-3): a listed old-layout fixture text that gains a line expecting the earlier folder's `receipts` for a fresh project fails, with the file, the name and both counts in the message.
  verify: `cargo nextest run --locked --offline --test state_folder_literal_scan`
- AC-2 (REQ-4): the same listed fixture text unchanged passes, and with one literal removed it still passes.
  verify: `cargo nextest run --locked --offline --test state_folder_literal_scan`
- AC-3 (REQ-1, REQ-2): an unlisted Python test text that spells the new folder name fails; text that only names `moai-cli` as a program, `moai-cli-v1.2.3` as a tag or `mo.ai.kr` counts nothing.
  verify: `cargo nextest run --locked --offline --test state_folder_literal_scan`
- AC-4 (REQ-1..REQ-5): the repository's own tests pass against the baseline file, and `tools/repo_scans.json` lists the scan with `tests/` `.rs` and `.py` in its scope.
  verify: `cargo nextest run --locked --offline --test state_folder_literal_scan` and `python3 -m unittest tests.repo_scans_test`
- AC-5 (REQ-6): with the switch set, a temporary tree's baseline is rewritten to its counts, sorted, and the judge then passes it.
  verify: `cargo nextest run --locked --offline --test state_folder_literal_scan`

## Scope

- write: `.moai/specs/SPEC-20261122-state-folder-literal-scan/spec.md`, `tests/state_folder_literal_scan.rs` (new), `tests/fixtures/state_folder_literals/allow.tsv` (new), `tools/repo_scans.json` (one entry)
- read-only: `tests/**/*.rs`, `tests/**/*.py` (as the scan's input)
- forbidden: any other test file, `src/`

## Decisions

- D-1: per-file counts, not per-line markers: marking 1,500 lines across 300 files would touch files other cards hold.
- D-2: counts may fall silently. Requiring exact counts would make every lane that fixes a literal edit the shared baseline, the hot-file conflict this redesign removes.
- D-3: the scan file spells neither name; it builds both at run time, so it needs no baseline row of its own.
- D-4: the area-crate test merge moves test files; it regenerates the baseline with REQ-6 in the same change.

## Plan

1. RED: `tests/state_folder_literal_scan.rs` with the judge cases (AC-1..AC-3, AC-5) and the repository case (AC-4), against an empty baseline file: the repository case fails.
2. GREEN: generate `allow.tsv` with the switch from a tip that holds the install-test repair; add the `repo_scans.json` entry.
3. `moai check affected` once, HANDOFF READY.

## Risks

- A lane whose branch adds a literal to a listed file gets a failure at its READY receipt: that is the point. The text says how to fix it with the resolver, or how to raise the count for a real fixture.
- Branches that landed after the baseline was generated can push a count above it on the tip: the scan then fails on the tip, and the next landing raises or fixes it. The leader's landing check runs the scan on the merged tree.
