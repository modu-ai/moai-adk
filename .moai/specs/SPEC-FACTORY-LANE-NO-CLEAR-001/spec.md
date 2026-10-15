---
id: SPEC-FACTORY-LANE-NO-CLEAR-001
title: "Lane card transition without /clear — in-session exit, lease, and enter"
version: "0.1.0"
status: in-progress
created: 2026-10-15
updated: 2026-10-15
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: "internal/cli, internal/hook, internal/config, internal/template/templates, .claude/rules/moai/workflow, .moai/policies/workflow, .claude/skills/moai/workflows, .claude/agents/moai"
lifecycle: spec-anchored
tags: "factory, lane, worktree, session-transition, clear-policy, template-first"
tier: M
related_specs: [SPEC-FACTORY-SELF-DISPATCH-001, SPEC-FACTORY-LANE-AUTONOMY-001, SPEC-FACTORY-LANE-WORKTREE-HANDOFF-001, SPEC-TODO-CLASSIFY-DISPATCH-001, SPEC-WORKTREE-EXIT-RETURN-001]
---

# SPEC-FACTORY-LANE-NO-CLEAR-001 — Lane card transition without /clear

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-10-15 | manager-spec | Initial plan-phase draft (card t1793, Class C, Template-First, P1, Tier M). Recorded eight open decision rows in `decision-index.md`. |

Basis: the requesting card (operator instruction of 2026-10-14 and 2026-10-15, recorded on the decision board as d-20261014T230709Z-5347). That board record was not pinned by digest and is not cited as an authority; the card is the instruction source for this SPEC. The release label `v3.2.0 target` follows the label carried by the current v3.2 SPEC series; no later release label is implied.

## §A Problem Statement

A factory lane (a session joined with `moai cc -l`) carries one card at a time. When a lane finishes its card, the doctrine today requires the operator to clear that lane session with `/clear` before the lane takes its next card. A lane cannot run `/clear` on itself, so every card transition waits on the operator. The operator instruction removes that wait: the lane leaves its finished card worktree with an in-session exit that keeps the tree, requests its next card itself from the primary checkout, and works the new card without carrying the previous card's conversation context.

Measured surface (research step of this plan phase; counts are literal `/clear` matches in the named file, observed in this worktree):

- The factory rule family carries the transition doctrine: `factory-dispatch.md` (7), `factory-dispatch-core.md` (3), `factory-dispatch-detail.md` (6), `factory-dispatch-cards.md` (4), `factory-dispatch-mechanics.md` (2), and `auto-semantics.md` (1). The same family is mirrored in `internal/template/templates/`, and a policy copy set lives in `.moai/policies/workflow/`.
- The lane-facing output: the clear-policy line that lane completion prints comes from one printer reached from three completion call sites. The SessionStart lane-join rule carries the next-card instruction and a recurring stall-recheck instruction in four conversation languages.
- The leader-facing text: `manager-lead.md` names `/clear` in four places, one of them the Role-B return template line that requests an operator `/clear {session-name}`.
- The operator-session serial `moai todo --auto` cycle emits per-card `/clear` guidance. The lane cycle (`moai todo --auto` inside a lane) contains no `/clear` text.
- The two surfaces disagree on isolation: the factory rule forbids moving a running card session into another card worktree and names relaunch as the rule's answer, while the in-session exit-then-enter sequence is the path the operator instruction requires.

The full location ledger, with a disposition per location, is in `plan.md` § Location Ledger.

Why this matters: the card transition is the one step in a lane's autonomous loop that still needs a human. Removing it lets a lane take the next dispatched card without waiting for the operator, and removes a context-carry dependency that the lane should not have.

Premise carried, not proven here: the in-session sequence (lease from the primary checkout, exit keeping the tree, enter the leased tree, verify the toplevel) is recorded in session memory as observed once on Claude Code 2.1.300. Session memory is not a committed authority. REQ-LNC-003 and AC-LNC-003 make the run phase re-observe it on the deployed runtime, and decision Q7 records the open question.

## §B Requirements (GEARS)

Scope of the transition rule: a Claude Code factory lane (a lane launched with the Claude harness, including the GLM-backed Claude-harness lane). Codex lanes keep their own transition (decision Q6).

- **REQ-LNC-001** (Event-driven): When a Claude Code factory lane's card reaches a terminal state (landing reported, parked, or ended), the lane shall leave the card worktree by an in-session exit that keeps that worktree, and shall not clear its conversation session.
- **REQ-LNC-002** (Event-driven): When a Claude Code factory lane has left its card worktree, the lane shall request its next card itself from the primary checkout through the factory lease verb or the todo auto lane cycle, and shall wait out a claim held by contention through the verb's wait option rather than abandoning the request.
- **REQ-LNC-003** (Event-driven): When a card has been leased to a Claude Code factory lane, the lane shall enter the leased card worktree within the same session and shall verify that the entered worktree is the leased card's own before any card work begins.
- **REQ-LNC-004** (Event-driven, undesired): When the in-session entry into the leased worktree is refused, the lane shall record the verbatim refusal in its card progress record and report it to the leader as a blocker, and shall not substitute a session clear, a second lease, or card work in the primary checkout; the only fallback is a session relaunch that the operator performs inside the leased worktree.
- **REQ-LNC-005** (Ubiquitous): The Claude Code factory lane shall work each newly leased card only from that card's body, its SPEC artifacts, and its progress record, and shall treat the conversation context of earlier cards as non-authoritative.
- **REQ-LNC-006** (Ubiquitous): No Claude Code factory lane shall require an operator clear command at any point between two cards that the lane carries.
- **REQ-LNC-007** (Event-driven): When the factory completion verb records a lane's completion, including the merge-observation path that records a merged pull request, the completion output shall carry no instruction to clear the session and shall name the in-session transition step that follows.
- **REQ-LNC-008** (Event-driven): When the todo auto lane cycle hands over from a completed card to the next card, its completion output and its exit guidance shall carry no instruction to clear the session.
- **REQ-LNC-009** (Ubiquitous): The factory clear-policy setting shall not cause any lane completion output to instruct a clear of the session, for any of its accepted values; the relaunch value shall keep its meaning of an operator-chosen session end and shall not name a clear.
- **REQ-LNC-010** (Ubiquitous): The SessionStart lane-join guidance shall name the in-session card transition as the lane's only transition between cards, shall name card takeover as a trigger for the recurring stall-recheck, and shall carry the same meaning in every supported conversation language.
- **REQ-LNC-011** (Ubiquitous): The leader-side dispatch surfaces shall not request that the operator clear a lane session; the leader's own session-clear rules are unchanged by this SPEC.
- **REQ-LNC-012** (Ubiquitous): The operator-session serial `--auto` cycle shall keep its per-card clear guidance, every surface that states that guidance shall name the operator session as its only scope, and the guidance shall not be applied to factory lanes (decision Q1 in `decision-index.md`).
- **REQ-LNC-013** (Ubiquitous): The factory doctrine family (the always-loaded stub, the full body, and each companion) shall state one identical transition rule for Claude Code lanes, shall replace the isolation clause that forbids moving between card worktrees with the exit-then-enter sequence of REQ-LNC-001 through REQ-LNC-003, and shall leave Codex lane transitions to decision Q6.
- **REQ-LNC-014** (Ubiquitous): Every changed doctrine or agent artifact shall be edited in its template source first; its local copy shall be byte-identical to the template source; the policy copies shall carry the same transition passages as the rule files under the path transform; and the generated codex agent definitions shall match their emission.
- **REQ-LNC-015** (Ubiquitous): The role-core injection test and the factory lane rule tests shall assert the new transition rule, so that the SessionStart injection for lanes and its test expectations agree.
- **REQ-LNC-016** (Ubiquitous): Every `/clear` reference that remains in the scoped surfaces after the change shall belong to one of three kinds — the leader session's own clear, the operator-session serial cycle, or a generic session-lifecycle or context-threshold note — and each remaining reference shall be listed with its kind in the location ledger.

## §C Scope Decisions

- Eight open decision rows are recorded in `decision-index.md` (Q1 through Q8). Rows that change a shipped default or a user-visible output are product-level and are not applied by the authoring agent. Each row's Default is stated only where the published Default rule ranks its options.
- Where the operator instruction itself fixes an outcome, the SPEC encodes that outcome and does not open a row. The clear-when-full context-threshold request is removed from lane completion output because the instruction requires zero `/clear` instructions in lane-cycle output. The residual risk of that removal is recorded in § E.
- The `--clear-policy` launch flag and its accepted values stay in place (decision Q2 default). Only the text the lane is told changes.
- The Claude Code lane transition replaces the isolation rule that forbade moving into another card worktree. The relaunch path remains only as the fallback named in REQ-LNC-004.

## §D Acceptance Summary

Sixteen acceptance criteria, AC-LNC-001 through AC-LNC-016, each with a verification command and an expected observable result, are in `acceptance.md`. The REQ-to-AC matrix is in `acceptance.md` § D. AC-LNC-003 is a run-phase observation on the deployed runtime and cannot be discharged in the plan phase.

## §E Residual Risk

- A lane whose context grows past the context-window threshold receives no lane-side prompt after this change, because the card's acceptance removes the threshold request from lane completion output. The orchestrator's threshold rule in the context-window-management doctrine still governs the orchestrator, and the operator may still clear the session at any time.
- The in-session entry premise (decision Q7) rests on one recorded observation. A refusal on the deployed runtime sends the lane to the blocker path of REQ-LNC-004, not to a clear.
- The policy copies carry drift outside the transition passages. The surgical default (decision Q5) leaves that drift in place.
- Two policy-only kanban doctrine copies still carry lane clear text until the operator decides (decision Q8).

## Exclusions (What NOT to Build)

Anything not named as a requirement in § B, or as a ledger row in `plan.md` § F, is out of scope for this SPEC. The groups below name the exclusions that a reader is most likely to mistake for part of the change.

### Out of Scope — Leader-session and operator-session clears

- The leader session's own clear rules, including its post-clear opening move and its own session clear after a card reaches done. These references are kept with kind "leader session's own clear" (REQ-LNC-016).
- The operator-session serial `--auto` cycle's per-card clear guidance is kept and scoped (REQ-LNC-012); its behavior is not changed.

### Out of Scope — Context-window, handoff, and goal clears

- Context-threshold clear guidance, the paste-ready handoff, the handoff injector, and goal re-arm on clear. These are session-lifecycle mechanisms, not the lane card transition.
- Retirement of the `--clear-policy` flag and its values (decision Q2 default keeps them).

### Out of Scope — Codex lane transition

- Codex lanes keep their session-end and new-session rules; their transition is decision Q6.
- The Codex owned-card rule and the root instruction file's Codex wording are not edited by this SPEC.

### Out of Scope — Orphan and root dogfood copies

- The policy-only kanban doctrine copies (decision Q8 default leaves them untouched).
- The root dogfood `.codex/agents/moai/manager-lead.toml` (decision Q4 default leaves it untouched).
- Non-transition passages of the policy copies (decision Q5 default is surgical).

### Out of Scope — Integration, messaging, and queue mechanics

- Changes to the integration window, the worktree tiers and disposal rules, the cross-session messaging rules, the lease machinery, and the queue data model.
- Changes to the verification batch, the merge gate, or the CodeRabbit read.
