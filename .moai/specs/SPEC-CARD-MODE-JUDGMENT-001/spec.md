---
id: SPEC-CARD-MODE-JUDGMENT-001
title: "Force the serial/parallel judgment at card design time — mode-judgment step, explicit classification at issuance, free-lane definition, executing-card-only lease renewal"
version: "0.1.0"
status: draft
created: 2026-10-19
updated: 2026-10-19
author: MoAI
priority: P1
phase: "v3.2.0 target"
module: "internal/template/templates/.claude/rules/moai/workflow, internal/template/templates/.claude/skills, .claude/rules/moai/workflow, .claude/skills"
lifecycle: spec-anchored
tags: "factory, card-classification, serial-parallel, issuance-procedure, lease-renewal, free-lane, guidance, template-first, card-t1894"
tier: L
card: t1894
related_specs: [SPEC-TODO-CLASSIFY-DISPATCH-001, SPEC-FACTORY-LEASE-RENEW-001, SPEC-ALWAYS-LOADED-BUDGET-001, SPEC-ROLE-INJECTION-BUDGET-001]
---

# SPEC: Force the serial/parallel judgment at card design time

## HISTORY

- 0.1.0 — 2026-10-19 — Plan-phase artifacts for card t1894 (Class C, P1, Template-First, operator instruction 10-19). Every figure below was measured on tree `74ad197380a0711cdb9816d4e37c94d29af730ea` (branch `WT-fail-safe-serial`, clean, unpushed) with the commands recorded in `research.md` and `progress.md` §E.1; nothing is carried from the card text. Code (the default mode and the reclassify verb) belongs to card t1737 and is not touched here.

## 1. Context and problem

Cards added without an explicit classification fall on the fail-safe default (`mode=serial`, `decider=default`; the code constant is the notice text at `internal/cli/todo_classify.go:41`, the default decider is `factory.DefaultCardDecider{}` at line 36). On 2026-10-19 that default had piled up 29 cards on the single serial slot, and lanes ran lease-renewal loops in front of that one slot (board records a672, 02f4, 0c51 cite the counts). The guidance that tells the leader how to issue a card never asks for a serial/parallel judgment, and the lane guidance tells a lane to keep a recurring recheck cron armed "to the completion report" whatever state its card is in. This SPEC changes **documents only**: it forces the judgment at issuance and fixes the renewal and free-lane wording.

## 2. Authorities (quoted, not paraphrased)

The operator's standing rule and the leader's standing rules are the policy. This SPEC cites them and adds none. Read with `moai decision read` on the pinned tree; quoted from the saved output (`/private/tmp/lxq/board-spec.txt`).

**0c51** (`decided_by=operator`, `scope=standing`, `kind=standing-rule`, `d-20261019T020637Z-0c51`):

> Card issuance must carry an explicit serial/parallel judgment. The leader adds every card with --classification-file (decider=human) whose mode is decided from the expected file set: parallelizable unless it edits the same files/shared state as an in-flight card or a slot-gated surface (factory/todo core, release pipeline, shared migrations), in which case serial with the conflicting card named in reason. The reason field states the file set and the conflict check. No card is issued on the fail-safe serial default; a card that slipped through is reclassified (t1737 verb once landed; until then reissue via classification-file, as t1888→t1892, t1889→t1893). Class C cards get --spec at pick time (cannot be added after assign).

**a672** (`decided_by=leader` under operator instruction, `scope=standing`, `kind=standing-rule`, `d-20261019T014804Z-a672`):

> Renew only the card the lane is actively executing. Do NOT renew leases on (a) merge-ready cards — expiry is harmless: at the landing slot the leader reassigns or the lane re-leases with factory next --card, stage preserved (verified t1876 reassign 01:4xZ, stage=merge-ready); (b) held/operator-wait cards; (c) serial cards waiting for the slot (assigned, no lease). No polling crons for slot or trigger waits — the leader sends a notice when a slot or landing turn opens. A lane whose only cards are (a)-(c) counts as free for dispatch.

**02f4** (`decided_by=operator`, `scope=card:t1737`): a one-time serial-slot exception for t1737 only. It sets no policy for this card and is cited only to bound the scope of code (t1737 owns the default and the reclassify verb).

## 3. Definitions

| Term | Meaning in this SPEC |
|---|---|
| mode | The classification field with the closed value set `serial` / `parallelizable` (`internal/factory/classification.go`). Independent of the card class (A / B / C). |
| slot-gated surface | The surfaces named by 0c51: factory/todo core, release pipeline, shared migrations. The board gives categories, not paths (open item OI-5). |
| executing | The lane is carrying out a stage of the card — plan, run, sync, the card-review, or the landing sequence — and is not waiting on a slot, a trigger, or the operator. |
| free lane | A lane on which no card is actually executing. A lane that holds only merge-ready, held (operator-wait), or serial-waiting cards is free. |
| live session | A running session occupies the lane label. A different notion from *free*: a lane with no live session is a fault, not a free lane (`factory-dispatch-mechanics.md` § Boundaries, unchanged). |
| template side / deployed side | `internal/template/templates/.claude/...` is the source; `.claude/...` is the deployed copy of this repository (Template-First). |

## 4. Requirements (GEARS)

Home documents (each rule is stated once, in its home, and cited by section name elsewhere — REQ-MJ-012): **C** = `factory-dispatch-cards.md`, **D** = `factory-dispatch-detail.md`, **M** = `factory-dispatch-mechanics.md`. Site list and line numbers: `research.md` § Site map.

### 4.1 Mode judgment and issuance (scope items 1 and 2)

- **REQ-MJ-001** The card-classification section of C (§ Card classes) shall carry a mode-judgment step that decides a card's execution mode from the card's expected file set, as an axis independent of the card class.
- **REQ-MJ-002** While the expected file set of a card overlaps the file set of an in-flight card or lies on a slot-gated surface (factory/todo core, release pipeline, shared migrations), the mode-judgment step shall assign `serial`; in every other case it shall assign `parallelizable`.
- **REQ-MJ-003** The mode-judgment step shall require the recorded reason to state the file set and the conflict check and, when the mode is `serial`, to name the conflicting card or slot-gated surface.
- **REQ-MJ-004** When a card is issued, the issuance procedure shall require an explicit classification supplied through `--classification-file` with decider `human`, and no statement of the procedure in the leader's queue-entry clause (`factory-dispatch.md`), the todo workflow skill (`gtd.md`), or the other documents of the sweep set R (`research.md` § Sweep set R) shall present issuing on the default decider as an option. A card that reached the queue on the default is reissued with an explicit classification; the documents shall not name a reclassification verb (REQ-MJ-015).
- **REQ-MJ-005** Where a card is Class C, the issuance procedure shall state that `--spec` is given when the card is picked and cannot be added after assign, and the todo skill shall not describe an empty `spec_id` on a picked Class C card as the normal state.

### 4.2 Free lane and assignment input (scope item 3)

- **REQ-MJ-006** The documents shall define a free lane as a lane on which no card is actually executing, shall state that a lane holding only merge-ready, held, or serial-waiting cards is free, and shall keep that definition separate from the live-session condition; the sentence "free means the lane's previous card reached `done`" shall not remain.
- **REQ-MJ-007** Where the assignment input is described, the documents shall define it as the picked cards plus the queued cards classified `parallelizable`, and shall leave promotion authority (the operator's pick, the `--auto` batch approval, a lane's lease) exactly as the existing clauses state it. *Held for the operator verdict of open item OI-3 before run-phase milestone M3 writes this sentence.*

### 4.3 Lease renewal (scope item 4)

- **REQ-MJ-008** A lane shall renew the lease of the card it is actually executing and of no other card; a lease on a merge-ready, held (operator-wait), or serial-waiting card may expire, expiry is harmless, and at the landing slot the leader reassigns the card or the lane re-leases it with `factory next --card`, the stage preserved.
- **REQ-MJ-009** While a lane waits on a slot or a trigger, the lane shall not arm or keep a polling cron for that wait, and the leader shall send a notice when a slot or a landing turn opens.
- **REQ-MJ-010** Each statement that arms or keeps the lane's standing recheck cron to the completion report (the always-loaded core stub, the role-core and long-form lane-wait clauses of `factory-dispatch.md`, `auto-semantics.md` §5.1, and the renew paragraph of M) shall be scoped to the time the lane is executing the card, and shall keep the stall-recovery purpose of that cron — waking a lane stopped by an API error or by a delegate report that never comes — for an executing card.
- **REQ-MJ-011** The lane-watchdog skill's first-step lease renewal (§0a) shall renew the card the lane is actually executing and shall not renew a merge-ready, held, or serial-waiting card, keeping its order (renew before measuring) for an executing card.

### 4.4 Cross-cutting

- **REQ-MJ-012** Each rule of REQ-MJ-002/-003/-005 (home C), REQ-MJ-006/-007 (home D), and REQ-MJ-008/-009 (home M) shall be stated once in its home document; every other document that needs it shall carry a pointer by section name and no restatement.
- **REQ-MJ-013** Every changed document shall be edited on the template side first, and the deployed copy shall be byte-identical to it, except that the declared one-line fork of `factory-dispatch.md` (the worktree-sweep sentence) shall remain exactly one differing line.
- **REQ-MJ-014** Where the always-loaded core stub is edited, it shall stay at or under 10,000 UTF-16 code units (measured today at 9,999) and the always-loaded rules subtotal under 150,000; no changed template document shall carry a card id, board record id, SPEC id, internal date, or commit SHA.
- **REQ-MJ-015** The documents shall not describe a changed classification default or a reclassification verb as available, and shall not alter the factual default-decider sentence of the todo skill's verb reference (`gtd.md`, the `moai todo add` row).
- **REQ-MJ-016** Where a guard test pins a changed sentence or token (the renew-interval wording, the binding ledger rows, the doctrine-section tokens, the watchdog first-step order, the mirror and neutrality tests), the change shall keep that guard green; a ledger row whose `after_text` is deliberately rewritten shall be updated in the same change.

## 5. Out of scope

### Out of Scope — code and runtime

- The default mode and the reclassify verb (card t1737, lane-18): no Go change, no new verb, no change to `factory.DefaultCardDecider`, and no description of either as available (REQ-MJ-015).
- The SessionStart lane notice in `internal/hook/session_start_factory.go` / `session_start_factory_i18n.go` (`laneRecheckRule`), which injects the same "arm before the first stage, `CronDelete` at the completion report" rule in four locales and is pinned by `internal/hook/lane_renew_interval_test.go`. It is Go code; it keeps the unscoped wording until a follow-up card changes it (open item OI-4, residual risk R-1).

### Out of Scope — documents deliberately left alone

- `.moai/policies/**` and `.moai/workflows/gtd.md`: deploy-time projections of the rule tree (`internal/template/harness_fs.go` maps `.moai/policies` to `.claude/rules/moai`). Measured stale already on this tree (the card-size-ceiling section is absent from the policy copy of C), regenerated by the update path, not hand-edited.
- `internal/template/templates/.claude/skills/moai/workflows/project/doc-generation.md:348` (a standing-source `[PROJECT]` card issued with a bare `moai gtd add`): outside the card's named scope; open item OI-2.
- The agent files `manager-todo.md` and `manager-lead.md` and their Codex TOML emissions: measured to carry no issuance procedure (open item OI-1).
- `docs-site/**`, `AGENTS.md`, `AGENTS.local.md`, `.moai/docs/**`: no statement of the changed rules found there (`research.md` § Sweep).

### Out of Scope — policy not decided here

- Which path set counts as a slot-gated surface, and which card states count as "in flight" for the overlap test (open item OI-5).
- The lane's "acknowledgement of receipt" of the new guidance: an operational check the leader records, not an acceptance criterion (AC-MJ-015).

## 6. Open items (decision-index rows)

The decisions this plan could not settle are rows Q1-Q6 in `decision-index.md`; none is decided by the planner except where the published Default rule ranks the options (stamped `DEFAULT-APPLIED`). Summary:

| Row | Question | Label | Default / state |
|---|---|---|---|
| Q1 / OI-1 | Do `manager-todo.md` / `manager-lead.md` get an issuance pointer? | FOUNDER, implementation-level | Default: no (measured: neither describes issuance) |
| Q2 / OI-2 | Is the standing-source card of `doc-generation.md` in scope? | FOUNDER, implementation-level | Default: no |
| Q3 / OI-3 | Does "assignment input = picked + queued parallelizable" change who promotes? | FOUNDER, product-level | Empty — blocks the autonomous Kickoff; gates REQ-MJ-007 only |
| Q4 / OI-4 | Is the Go SessionStart lane notice in scope? | FOUNDER, implementation-level | Default: no (follow-up card) |
| Q5 / OI-5 | "In-flight" membership; paths of the three slot-gated surfaces | FOUNDER, implementation-level | Default: keep the board's words verbatim |
| Q6 | Tier | POLICY-COVERED | Tier L by the files-affected column (`design.md` D-1) |

## 7. Known limits and residual risk

- **R-1** Until OI-4 is acted on, a lane that reads its SessionStart notice is told to keep the recheck cron to the completion report whatever its card state; the scoped document wording and the injected notice disagree. Mitigation: the operational check AC-MJ-015 names the notice, and the follow-up card is proposed in the report.
- **R-2** The mode judgment is a human/LLM judgment recorded in a reason field; no command here verifies that a stated file set is the card's real file set (`design.md` D-6).
- **R-3** `moai todo next` / assign behaviour for Class C (`--spec` at pick) is taken from board 0c51 and from `moai todo next --help` (a `--spec` flag exists on `next`); the "cannot be added after assign" fact was not re-derived from code in this plan phase.
- **R-4** The installed `moai` build (`moai_cp/20261007_195927`) predates this tree's HEAD; every CLI observation in `research.md` names it as the judging build (verification-claim-integrity §2.2).
