---
id: SPEC-20261124-bundle-tip-doc-refusal
title: "bundle operations page names the local integration tip, and the missing-tip refusal names both refs"
status: approved
revision: 1
created: 2026-11-24
approval:
  by: leader
  at: 2026-11-23
  basis: "리더 결정 노트 7703 (2026-11-23T16:57:18Z, 사용자 위임 2026-11-23T13:57Z, 리더 판단, 노트 7702 의 (a)): 판 1 과 D-1·D-2 승인; 리더가 노트에 적었고 레인은 그 노트를 인용만 해요"
  digest: "sha256:4a15bccaa646a835785ee100b0712cfe2a7f9e3ddf5dce7bf46e437fc4d6fd8a"
---

# bundle operations page names the local integration tip, and the missing-tip refusal names both refs

## Intent

Follow-up of `SPEC-20261124-bundle-local-tip` (card C-408, AUDIT PASS with two non-blocking notes). That change made the default tip of `moai bundle` the local `refs/heads/<work_flow.integration_branch>`, with `origin/<branch>` only when the local branch is missing, and a profile without an m-flow integration branch keeping `origin/v2026.11.0`. Two places still say the old thing.

1. N2: the operations page `docs/design-review/appendix/operations/bundle.md` (the `bundle ready` section, line 9) still says the default tip of an m-flow profile is `origin/<branch>`.
2. N3: when neither the local branch nor `origin/<branch>` exists, the refusal in `tip_of` (`src/bundle.rs`) names only `origin/NeXT`, so a reader cannot tell that the local branch was checked too.

Readers: the lander and the auditors who run `moai bundle` and read its page and its refusal.

## Interview

- Q1: Does the refusal need to know that the tip was defaulted? -> A: No (D-1). `tip_of` is reached by `ready`, `check`, `body`, `build` and `verify`; a flag through five signatures costs more than one true extra fact in the message.
- Q2: Does the operations page need a new pinned test? -> A: No (D-2). The three tests that read it pin other words and are held by C-414; a new reader needs a `tools/repo_scans.json` entry.
- Q3: Where does the new test live? -> A: `tests/bundle_tip_refusal_cli.rs`; its first word `bundle` is an area of `tests/areas.toml`, which `test_area_layout` checks.

## Non-goals

- Changing which ref is the default tip or the order of the two names: that is C-408's behavior and stays.
- Changing the refusal of a head that is missing (`head_missing`, `fetch first` on a head), which is another message.
- A test that reads the operations page for the new sentence: the three tests that read the page (`bundle_body_cli`, `bundle_build_cli`, `bundle_verify_cli`) pin unrelated words, are held by another card, and a new reader of the page would need a `tools/repo_scans.json` entry owned by other cards (D-2).

## Requirements

- REQ-1: The operations page of `bundle` shall say that without `--tip` an m-flow profile with a recorded integration branch judges the local `refs/heads/<branch>` when it is a commit, else `origin/<branch>`, and that a profile without one keeps `origin/v2026.11.0`.
- REQ-2: When the tip that `bundle` judges is `origin/<name>` and neither it nor `refs/heads/<name>` is a commit in the repository, the tool shall refuse and name both `origin/<name>` and `refs/heads/<name>`.
- REQ-3 (REQ-2): IF the tip is not of the form `origin/<name>`, THEN the tool shall refuse as before and name only that value.
- REQ-4 (REQ-2): IF the local `refs/heads/<name>` is a commit, THEN the tool shall refuse a missing `origin/<name>` naming only that value.

## Acceptance

- AC-1 (REQ-1): the operations page names `refs/heads/<branch>` before `origin/<branch>` for the m-flow default and keeps `origin/v2026.11.0` for a profile without a recorded branch · verify: `git grep -n "refs/heads/<branch>" docs/design-review/appendix/operations/bundle.md`
- AC-2 (REQ-2): in a temporary repository whose profile names an m-flow branch `NeXT`, with neither a local `NeXT` nor `origin/NeXT`, `bundle check` without `--tip` exits 1 and the refusal names `origin/NeXT` and `refs/heads/NeXT` · verify: `cargo nextest run --locked --offline --test bundle_tip_refusal_cli`
- AC-3 (REQ-3): the same command with `--tip no-such-branch` exits 1 and the refusal names `no-such-branch` and no `refs/heads/` · verify: `cargo nextest run --locked --offline --test bundle_tip_refusal_cli`
- AC-4 (REQ-4): with a local `NeXT` and no `origin/NeXT`, an explicit `--tip origin/NeXT` refuses and names `origin/NeXT` but not `refs/heads/NeXT` · verify: `cargo nextest run --locked --offline --test bundle_tip_refusal_cli`

## Scope

- write: `.moai/specs/SPEC-20261124-bundle-tip-doc-refusal/spec.md`, `docs/design-review/appendix/operations/bundle.md`, `src/bundle.rs`, `tests/bundle_tip_refusal_cli.rs`
- read-only: `src/ops/bundle.rs` (`resolve_tip`, `DEFAULT_TIP`), `tests/bundle_check_cli.rs`, `tests/bundle_body_cli.rs`, `tests/bundle_build_cli.rs`, `tests/bundle_verify_cli.rs`
- forbidden: `tools/repo_scans.json` and the test files above while another card holds them

## Plan

1. RED: `tests/bundle_tip_refusal_cli.rs` with a temporary HOME, a repository with a commit, an m-flow profile and no `NeXT` anywhere; `bundle check C-1=<40 hex>` without `--tip` must name both refs and fails today.
2. `tip_of` in `src/bundle.rs`: when the tip starts with `origin/` and `refs/heads/<rest>` is not a commit either, the refusal also names `refs/heads/<rest>`.
3. The `bundle ready` paragraph of the operations page says what C-408 changed.

## Risks

- A refusal with two names for an explicit `--tip origin/x`: the second name is a true fact (the local branch is not a commit either) and not a claim that it was looked up; AC-4 pins that a present local branch changes nothing.
- The card named `src/bundle.rs`, the page and `tests/bundle_check_cli.rs`; that test file is held by C-414, so the new test lives in its own file (D-2).

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | The refusal names both refs inside `tip_of`, for any tip of the form `origin/<name>`, instead of threading a "defaulted" flag through five actions | card C-413 scope names `src/bundle.rs`; the statement is true for explicit tips too | revert the one `tip_of` change |
| D-2 | The new test is a new file, and no new test reads the operations page | `tests/bundle_check_cli.rs` and the three page readers are in C-414's write scope (lease to 2026-11-23T20:28Z); a new page reader needs a `tools/repo_scans.json` entry, which other cards write in turn | add a pinned line to a page reader after C-414 lands |

## Result

Not started.
