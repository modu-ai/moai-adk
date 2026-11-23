---
id: SPEC-20261124-bundle-local-tip
title: "bundle check judges the local integration branch tip, and each audit keeps its own scratch folder"
status: draft
revision: 1
created: 2026-11-24
---

# bundle check judges the local integration branch tip, and each audit keeps its own scratch folder

## Intent

One problem of two kinds came up twice on the night of 2026-11-23 (the recurring-problem rule of `CLAUDE.md`, card C-408).

1. `moai bundle` without `--tip` judged `origin/<work_flow.integration_branch>` (`src/ops/bundle.rs` `resolve_tip`). The integration branch is the local `NeXT`; `origin` is behind it (user decision 2026-11-23, `CLAUDE.md` "Repository and preservation boundaries"). The answer was `bundle refused: the tip `origin/NeXT` is not a commit here (fetch first)`, so the audits of C-354 slice 5 and C-318 slice 2 could not read the tier floor.
2. Two auditors wrote the same folder name `merged/` in the shared scratchpad and overwrote each other's files (the same two audits; each repaired its own).

Fix: the default tip is the local `refs/heads/<branch>`, and `origin/<branch>` only when the local branch does not exist; the audit contract says that scratch space is one folder per audit, named for it.

Readers: the lander and the auditors, who run `moai bundle check` and keep scratch files during an audit.

## Non-goals

- Changing which flow or profile key drives the default tip: only the m-flow integration branch of the profile does, as today.
- Fetching, or reading the remote, to decide the default tip: the tool never fetches.
- A tool that creates or checks the scratch folders: the line is a rule in the audit contract with a pinned text.
- The rule-provenance table `docs/claude-md-rule-provenance.md`: outside this card's write scope; the cause line goes to a scope-expansion request.

## Requirements

- REQ-1: When `moai bundle` judges a tip without `--tip` and the profile names an m-flow integration branch, the tool shall answer the local `refs/heads/<branch>` as the tip when that branch is a commit in the repository.
- REQ-2 (REQ-1): IF the local `refs/heads/<branch>` is not a commit in the repository, THEN the tool shall answer `origin/<branch>` as the tip.
- REQ-3: When `--tip` is given, the tool shall judge that value unchanged.
- REQ-4: The `--tip` help text of every bundle action shall name the local branch first and `origin/<branch>` as the fallback.
- REQ-5: The audit contract shall state in one line that an audit keeps scratch files only in a folder of its own name and neither reads nor writes another audit's folder.
- REQ-6: The source manifest entry of `references/audit.md` shall carry the checksum of the changed file, written by the manifest tool.
- REQ-7 (REQ-3): IF the given `--tip` is not a commit in the repository, THEN the tool shall refuse and name that value.

## Acceptance

- AC-1 (REQ-1, REQ-3): in a temporary repository with a profile that names an m-flow integration branch, a local branch of that name and no `origin/<branch>`, `bundle check` without `--tip` answers the tier floor with the tip named `refs/heads/<branch>`, and with `--tip` it still judges the given value · verify: `cargo nextest run --locked --offline --test bundle_check_cli`
- AC-2 (REQ-2, REQ-7): in the same repository without the local branch but with `origin/<branch>`, `bundle check` answers with the tip `origin/<branch>`; with neither it refuses and names the tip it looked for; a `--tip` that is no commit is refused with its value named · verify: `cargo nextest run --locked --offline --test bundle_check_cli`
- AC-3 (REQ-4): the `--tip` help text of `ready`, `check`, `body`, `build` and `verify` names the local branch and says `origin/<branch>` only when it is missing · verify: `cargo nextest run --locked --offline --test bundle_check_cli`
- AC-4 (REQ-5, REQ-6): `audit.md` carries the scratch-folder line within the 6144-byte cap of the reference files, a pinned test fails when the line is removed, and the manifest checksum equals the file's · verify: `cargo nextest run --locked --offline --test bundle_check_cli --test next_flow_skill_scan`

## Scope

- write: `.moai/specs/SPEC-20261124-bundle-local-tip/spec.md`, `src/ops/bundle.rs`, `src/bundle.rs`, `tests/bundle_check_cli.rs`, `skills/moai/references/audit.md`, `skills/moai/references/source-manifest.json`
- read-only: `tests/next_flow_skill_scan.rs` (the size cap and manifest pins), `src/audit_delta.rs` (the scrubbed git command)
- forbidden: `docs/claude-md-rule-provenance.md` and `CLAUDE.md` until a card whose write scope names them (request below)

## Plan

1. RED: `tests/bundle_check_cli.rs` gets a repository with a profile, a local integration branch and no `origin/<branch>`; `bundle check` without `--tip` fails today with the fetch-first refusal.
2. `src/bundle.rs` gets one public function that resolves the default tip of a branch in a repository (the local ref when it is a commit, else `origin/<branch>`) through the existing scrubbed git helper; `src/ops/bundle.rs` calls it with the `--dir` repository and its help lines change.
3. RED then GREEN for the help text and the audit contract line: a pinned test on the line, the line added to `audit.md` with room made under the 6144-byte cap, and the manifest checksum rewritten by the tool.
4. Request the scope expansion for the rule-provenance cause line as a `LANE-BLOCKED kind=scope` note.

## Risks

- `audit.md` has 38 bytes of room under its 6144-byte cap; the new line costs more, so an existing sentence is shortened without changing a pinned phrase (`tests/core_rules_2876.rs`, `tests/next_flow_skill_scan.rs`). The size test names the failure.
- The default tip is judged in the `--dir` repository, not the `--project` folder; a test whose project folder and repository differ guards this.
- Tests that read repository files by `CARGO_MANIFEST_DIR` are the subject of C-398; the pinned test reads `audit.md` only as the existing skill-scan tests do.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | The default tip is the local integration branch, `origin/<branch>` only as the fallback | user decision 2026-11-23 (the local `NeXT` is the integration branch, `origin` is behind) and card C-408 | restore `resolve_tip` to return `origin/<branch>` |
| D-2 | The scratch rule is a line in `references/audit.md` with a pinned test, no new tool | card C-408 asks for one line and a pinned test | remove the line and its test |
