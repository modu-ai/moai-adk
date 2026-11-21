---
id: SPEC-20261121-stage-order-once
title: "The skill states the stage order once, in the SDD reference, and every other document points to it"
status: draft
revision: 1
created: 2026-11-21
---

# The skill states the stage order once, in the SDD reference, and every other document points to it

## Intent

The order PLAN → RUN → VERIFY → SYNC → AUDIT → merge is written out in full in six places of the shipped skill: `references/internal/sdd.md` (step 8, with what each stage does), `references/chains/index.md` (the rule list), and the workflows `audit.md`, `sync.md`, `verify.md` and `run.md`. Six copies drift: the next change to one stage has to find and edit all of them, and the one that is missed tells a session the wrong order. The skill is structured in four roles (the entry `SKILL.md` and the route table pick the route, the chains pick the agents and their order, a workflow holds the inputs, outputs, steps, blocks and resume of one route, a reference is read only when needed), and the order of the stages belongs to one reference, `internal/sdd`, that the others cite. This SPEC keeps the full sentence in `internal/sdd` only, turns the other five copies (`chains/index.md`, `audit.md`, `sync.md`, `verify.md` and `run.md`, called the five pointer files below) into pointers, and replaces the test that asked every document to repeat the order with a test that asks for one copy and five pointers. It changes skill text and tests only.

## Non-goals

- The round order of a lane in `workflows/mode-agent.md` (`take → PLAN → RUN … → merge → done`), which is a loop, not the stage order, and the test that pins it.
- The route lists of the chain table in `references/chains/index.md` (the request rows keep their stage lists) and the test that pins them.
- Slimming `SKILL.md`, limiting `/loop` to a confirmed leader session, and any change to what a stage does.
- Editing the earlier SPEC `SPEC-20261117-sdd-guidance`: this SPEC replaces its REQ-3 and leaves that file as it is.

## Requirements

- REQ-1: The skill shall state the order PLAN → RUN → VERIFY → SYNC → AUDIT → merge in full only in `references/internal/sdd.md`.
- REQ-2: Each of the five pointer files shall say that the order of the stages is the one `internal/sdd` gives, without writing the order out.
- REQ-3: Each stage workflow shall keep the sentences that say what its own stage reads, writes or refuses.
- REQ-4: The chain table shall keep its request rows with their stage lists.
- REQ-5: The round order of `workflows/mode-agent.md` shall stay as it is.
- REQ-6: The skill's source manifest shall match the files after the change.

## Acceptance

- AC-1 (REQ-1, REQ-2): the full order is in `sdd.md` only, and each of the five other files names `internal/sdd` as the owner of the order.
  verify: `cargo test --locked --offline --test sdd_guidance stage_order_is_written_once`
- AC-2 (REQ-3): each workflow still says what its stage reads, writes or refuses.
  verify: `cargo test --locked --offline --test sdd_guidance each_stage_workflow_keeps_its_own_duty`
- AC-3 (REQ-4, REQ-5): the chain rows and the lane round order are unchanged.
  verify: `cargo test --locked --offline --test agent_defs_package` and `cargo test --locked --offline --test mode_workflow_steps`
- AC-4 (REQ-6): the manifest matches the package.
  verify: `python3 -I tools/source_manifest.py check`

## Scope

- write: `.moai/specs/SPEC-20261121-stage-order-once/**`, `tests/sdd_guidance.rs`, `skills/moai/references/internal/sdd.md`, `skills/moai/references/chains/index.md`, `skills/moai/workflows/audit.md`, `skills/moai/workflows/sync.md`, `skills/moai/workflows/verify.md`, `skills/moai/workflows/run.md`, `skills/moai/references/source-manifest.json` (regenerated last, after its other holders release it)
- read-only: `skills/moai/workflows/mode-agent.md`, `tests/mode_workflow_steps.rs`, `tests/agent_defs_package.rs`
- forbidden: a card, Issue or pull request number in the sentences this change adds, and any edit of a file another card holds before that card has released it or agreed in a pair line

## Plan

1. Rewrite the test first: `tests/sdd_guidance.rs` asserts the full order in `sdd.md` only and a pointer to `internal/sdd` in the five others; it fails on the tip because all six files carry the order.
2. Replace the order in `chains/index.md`, `audit.md`, `sync.md`, `verify.md` and `run.md` by a pointer, keeping every sentence about the stage's own duty.
3. Regenerate the manifest row with `python3 -I tools/source_manifest.py write`; run `sdd_guidance`, `agent_defs_package`, `mode_workflow_steps`, `product_cli skill_package` and every other test that names these files.
4. Ask for an independent audit; the implementer does not audit.

## Risks

- A neighbouring card edits the same workflow (the guidance sweep, the M-Flow guidance, the branch and base text): the pair line names the exact sentence each adds, and whoever lands second merges by sentence.
- A test other than `sdd_guidance` pins the sentence this SPEC removes: a search of `tests/` for the order before the edit lists them, and each one is moved to the pointer or kept with a reason.
- The source manifest is held by other cards at the time of writing: the manifest row is written last, after they release it.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | `internal/sdd` keeps the one full copy of the order | It already carries what each stage does, so the order sits with its meaning | Move the sentence and its test |
| D-2 | This SPEC replaces REQ-3 of `SPEC-20261117-sdd-guidance` and does not edit that file | Both are drafts; the newer test is the live one once this SPEC merges | Restore the earlier test sentence |
| D-3 | The lane round order and the chain rows stay | They are not the stage order and their own tests pin them | Add them to the pointer rule later |

## Result

Filled in by the implementation lane on 2026-11-21; the audit checks it again. The SPEC is `186c3aad4d`; red `fae26aa734`, green `3fd0306b99`; the local tip `42f5af0800` is merged in as `0d74395003` (the earlier tips as `c4ce5f1c4b` and `3145f287fb`, no conflict each time, `git merge-tree` exit 0); the five manifest rows are `d48ffb4b42`; the formatting commit is `56506f1712`. Evidence files are in this worktree's `target/notes/`.

Red (`red1.out`, before the text): `every_place_states_one_order` and `stage_order_is_written_once` FAILED on the tip text, 10 passed and 2 failed in `tests/sdd_guidance.rs`; `each_stage_workflow_keeps_its_own_duty` passed on the tip because it guards what must stay.

- AC-1: met — `cargo test --locked --offline --test sdd_guidance` (`green1.out`, and again in `affected/06-test_targets.log`): 12 passed, 0 failed; `stage_order_is_written_once` is one of them and finds the full order only in `references/internal/sdd.md` and the pointer in the five files.
- AC-2: met — same run, `each_stage_workflow_keeps_its_own_duty`.
- AC-3: met — `cargo test --locked --offline --test agent_defs_package --test mode_workflow_steps` (`green1.out`): 19 and 43 passed; `agent_defs_package` 19 passed again in `affected/06-test_targets.log`.
- AC-4: met — `python3 -I tools/source_manifest.py check` prints `{"conflicted": false, "changed": [], "dropped": [], "ok": true}` on the final head; `product_cli` 79 passed in `affected/06-test_targets.log`.

### Checks

`moai check affected 42f5af0800..0d74395003` (the branch's own build of the binary, `--max 2`), summary line as printed: `check-affected base=42f5af0800fb head=0d743950031d targets=43 py=8 scans=listed:42 diff_check=0 fmt=1 clippy=0 test_lib=0 test_bins=0 test_targets=0 build_cli=0 python=0 result=FAIL`. The one failing step was `fmt` (`affected/02-fmt.log`): `tests/sdd_guidance.rs` was not rustfmt-clean. It is fixed in `56506f1712` (formatting only); afterwards `rustfmt --edition 2021 --check` is clean on every changed Rust file and `cargo test --locked --offline --test sdd_guidance` (`fmtfix.out`) passes 12. The whole `check affected` was not repeated on the final head.

Mutants (five, on the committed files, each restored, `restored: clean = True`; `mutate.out`): the full order written out again in `run.md`, the pointer dropped from `verify.md`, the `sync` duty sentence dropped, the full order removed from the reference, the chain rule writing the order out — each killed by a named test of `sdd_guidance`.

### Not run

- `moai check affected` on the final head `56506f1712`: not run again (a formatting-only commit follows the run; its own check is the `fmt --check` and the target above).
- `mode_workflow_steps` inside the `check affected` run: it was not in the affected set; its 43 passes are from `green1.out` on the green commit.
- A second run of the same tests on the same tree: not repeated.
