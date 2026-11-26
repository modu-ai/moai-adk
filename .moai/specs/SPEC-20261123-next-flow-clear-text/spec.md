---
id: SPEC-20261123-next-flow-clear-text
title: "NeXT-Flow lane text: when a lane session is cleared and how the next card starts"
status: draft
revision: 1
created: 2026-11-23
---

# NeXT-Flow lane text: when a lane session is cleared and how the next card starts

## Intent

User pick of 2026-11-22 (a selection in the leader session, after a peer relayed the wording): the flow reference states, in two stages, when a lane session is cleared and how its next card starts. The transition stage is what the leader and Orca do today; the final flow is what the product does when `/clear` and Orca are no condition. The instruction file and the WORK workflow must not contradict it: a READY, REPAIRED or MERGE-READY note and a `LANE-DONE` line on the screen are hand-backs, never done in the ledger. Evidence: on 2026-11-22T22:41Z four lanes released their cards (ready and released, not done), were cleared and reserved their next cards, and the reservations were read back. Docs only: no structure, install or freeze change. The shipped flow text says "a lane-terminal manager" where this SPEC says Orca, because a shipped text names Orca only in its adapter (`tests/orca_optional_3005.rs`); `CLAUDE.md` is not shipped and names Orca.

## Requirements

- REQ-1: The `A lane` section of `skills/moai/references/flows/next-flow.md` shall name two stages, Transition and Final flow.
- REQ-2: The Transition stage shall let the leader send `/clear` to the exact existing window only when the ledger shows the lane's card done or released and nothing is open: no card in review, no live reservation, no run of the lane's own, no unsaved change, no input or selection prompt on the screen and nothing unknown.
- REQ-3: When a window has been cleared, the Transition stage shall have the leader confirm that the new session is ready without taking a fixed delay as proof, send the next line once and confirm that session's reservation of the new card as the ACK.
- REQ-4 (REQ-3): IF the ACK is not seen, THEN the leader shall inspect the window instead of sending the line again.
- REQ-5: The Transition stage shall open a new window only when the existing one cannot be recovered.
- REQ-6: In the Final flow `/clear` and Orca shall not be conditions of dispatch.
- REQ-7: In the Final flow the start signal for the worker's next reservation shall be a card verified and landed on the local integration branch and then done with its reservation released in the ledger.
- REQ-8: In the Final flow the worker shall reserve the next eligible card once.
- REQ-9: In the Final flow an automatic restart shall run only through the existing HOLD, ownership, priority and safety checks.
- REQ-10: The Final flow text shall say that waiting for an audit and the hand-back are separate from done, and that a worker clearing its own session is not documented as a guarantee.
- REQ-11: `CLAUDE.md` `Clear after every card` and `skills/moai/workflows/work.md` shall not contradict the two stages or call a READY note or a `LANE-DONE` line done.
- REQ-12: The source manifest shall carry the sha256 of each changed shipped file.

## Acceptance

- AC-1 (REQ-1, REQ-2, REQ-5): `next-flow.md` names the two stages and the Transition conditions, within its 6144-byte limit · verify: `cargo test --locked --offline --test next_flow_skill_scan`
- AC-2 (REQ-3, REQ-4): the Transition text names the readiness check, the next line sent once, the ACK read back and the inspection when the ACK is missing · verify: `grep -n "ACK" skills/moai/references/flows/next-flow.md`
- AC-3 (REQ-6, REQ-7, REQ-8, REQ-9, REQ-10): the Final flow text names the start signal, the one reservation, the existing checks, the separate wait for an audit and the self-clear that is no guarantee · verify: `grep -n "Final flow" skills/moai/references/flows/next-flow.md`
- AC-4 (REQ-11): `CLAUDE.md` stays inside its 44 KiB budget and `work.md` keeps its steps, and neither calls a READY note or a `LANE-DONE` line done · verify: `cargo test --locked --offline --test repo_claude_md --test next_flow_skill_scan --test mode_workflow_steps`
- AC-5 (REQ-12): the manifest sha256 of each changed shipped file equals the file's · verify: `cargo test --locked --offline --test source_manifest_test`

## Scope

- write: this SPEC folder, `skills/moai/references/flows/next-flow.md`, `skills/moai/workflows/work.md`, `CLAUDE.md`, `skills/moai/references/source-manifest.json`
- read-only: everything else. No test, code, structure, install or freeze change.

## Plan

1. State the Transition and Final flow stages in the `A lane` section of the flow reference, in the conditions the card gives.
2. Change `CLAUDE.md` `Clear after every card` and the hand-off pointer, and the `work.md` land and repeat steps and its completion check, only where they contradict the stages.
3. Update the source manifest sha256 of the changed shipped files; run the tests that read these files and one receipt run at READY.
