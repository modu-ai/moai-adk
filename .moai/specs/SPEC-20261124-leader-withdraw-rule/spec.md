---
id: SPEC-20261124-leader-withdraw-rule
title: "Leader guidance: HOLD cards leave ready through card withdraw"
status: approved
revision: 1
created: 2026-11-24
approval:
  by: leader
  at: 2026-11-24
  basis: "리더 결정 노트 8238 (2026-11-24T05:33:07Z, 사용자 위임 2026-11-23T13:57Z, 리더 판단, 사용자 직접 선택 아님, 노트 8237 에 답함): 판 1(커밋 3f30135d7c) 승인; 레인은 그 노트를 인용만 해요"
  digest: "sha256:a784cea65786cc640c41e71d698b5c095b7e031c0a6200d27b78495a44c24bc3"
---

# Leader guidance: HOLD cards leave ready through card withdraw

## Intent

A card that carries a HOLD or an unmet prerequisite and stays `ready` is taken by a lane and given back again; card C-3 went through that three times. `moai card withdraw` exists and keeps such a card out of `ready` (`take` refuses it naming the reason, `card list` hides it), but no leader guidance names it, so the leader has no written rule to withdraw instead of leaving the card ready. This change adds that one rule to `docs/leader-rules.md` and pins it with a test.

Readers: the leader session that refills the ready list, and the auditor of this change.

## Interview

- Q1: Does the stale reclaim sentence of the earlier card still need a fix? -> A: No (D-1). The file that held it, `skills/moai/workflows/mode-leader.md`, left the package in `eb5a74a0c9`, and `docs/leader-rules.md` already names `moai card reclaim` in its lease line.
- Q2: Does the rule tell lanes to push an unpushed branch before a reclaim? -> A: No (D-2). A lane only commits and never pushes (`CLAUDE.md` Flow 6), so that sentence of the earlier card does not hold today and is not carried over.
- Q3: Which document carries the rule? -> A: `docs/leader-rules.md`, as one more numbered item of "What a round does" (D-3). The item "4. Refill" is pinned by `tests/followup_per_spec_rules.rs` and stays as it is.

## Non-goals

- Any change of `moai card withdraw`, `card take` or the ledger.
- The reclaim procedure and the lane rules (`docs/lane-rules.md`).
- `skills/moai/workflows/leader.md`: the route text is the subject of `SPEC-20261124-moai-leader-route`.

## Requirements

- REQ-1: The leader rules in `docs/leader-rules.md` shall tell the leader to withdraw a card that carries a HOLD or an unmet prerequisite with `moai card withdraw` instead of leaving it ready.
- REQ-2: The test `tests/leader_reclaim_guidance.rs` shall fail when the line of REQ-1 is missing from `docs/leader-rules.md` or lacks the command or the condition.

## Acceptance

- AC-1 (REQ-1): the item of `docs/leader-rules.md` that starts `6. Keep HOLD cards out of ready` holds `` `moai card withdraw` `` and `HOLD or an unmet prerequisite` · verify: `cargo nextest run --locked --offline --test leader_reclaim_guidance`
- AC-2 (REQ-2): on a copy of `docs/leader-rules.md` without that item the test fails with a non-zero exit, and on the committed file it passes · verify: `cargo nextest run --locked --offline --test leader_reclaim_guidance`

## Scope

- write: `.moai/specs/SPEC-20261124-leader-withdraw-rule/spec.md`, `docs/leader-rules.md`, `tests/leader_reclaim_guidance.rs`
- read-only: `CLAUDE.md`, `docs/lane-rules.md`, `src/ops/card.rs`, `tests/followup_per_spec_rules.rs`, `tests/repo_claude_md.rs`, `tools/repo_scans.json`
- forbidden: `tools/repo_scans.json`

## Plan

1. Write `tests/leader_reclaim_guidance.rs` first and run it red on the unchanged file.
2. Add item 6 to "What a round does" in `docs/leader-rules.md` and run the test green.
3. Run `moai check affected` once at READY and write the READY note with the receipt and the guard-mutant line.

## Risks

- A text pin proves the sentence exists, not that a leader follows it; the rule is a prompt to the leader and changes no command. The pins are the words of REQ-1 and the item's start, so a reworded line must update the test in the same change.
- A new item 6 shifts nothing: items 1 to 5 keep their numbers, and `tests/followup_per_spec_rules.rs` reads "4. Refill" by its start only.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | The stale reclaim sentence is not rewritten | its file left the package in `eb5a74a0c9`; `docs/leader-rules.md` already names `moai card reclaim` | add a reclaim item in a later SPEC |
| D-2 | No "push the holder's branch first" step | a lane only commits and never pushes (`CLAUDE.md` Flow 6) | add it in a later SPEC if the flow changes |
| D-3 | The rule is item 6 of "What a round does" in `docs/leader-rules.md` | the file already holds the leader's round steps; item 4 is pinned by another test | move the line |

## Result

Draft revision 1: not built yet.
