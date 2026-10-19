---
id: SPEC-CARD-MODE-JUDGMENT-001
title: "Force the serial/parallel judgment at card design time — mode-judgment step, explicit classification at issuance, free-lane definition, executing-card-only lease renewal"
version: "0.1.3"
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
related_specs: [SPEC-TODO-CLASSIFY-DISPATCH-001, SPEC-FACTORY-LEASE-RENEW-001, SPEC-ALWAYS-LOADED-BUDGET-001, SPEC-ROLE-INJECTION-BUDGET-001]
---

# SPEC: Force the serial/parallel judgment at card design time

## HISTORY

- 0.1.0 — 2026-10-19 — Plan-phase artifacts for card t1894 (Class C, P1, Template-First, operator instruction 10-19). Every figure below was measured on tree `74ad197380a0711cdb9816d4e37c94d29af730ea` (branch `WT-fail-safe-serial`, clean, unpushed) with the commands recorded in `research.md` and `progress.md` §E.1; nothing is carried from the card text. Code (the default mode and the reclassify verb) belongs to card t1737 and is not touched here.
- 0.1.1 — 2026-10-19 — Correction round before the first plan audit (a second commit, not an amendment). The operator ruling `d-20261019T023441Z-d314` settled decision-index row Q3: the leader may itself promote queued `parallelizable` cards to `picked` and assign them while the operator has delegated dispatch, under stated limits. REQ-MJ-007 and AC-MJ-006 are rewritten to implement exactly that; REQ-MJ-017, AC-MJ-018 and AC-MJ-019 are added for the sentences that stated the old operator-only rule (measured in `research.md` §10); the Go SessionStart notice moves to card t1695 (Q4) and stays out of scope. An earlier plan audit was stopped before it wrote a verdict, so no audit hash covers the 0.1.0 text; the lane recomputes the plan-artifact hash after this commit.
- 0.1.2 — 2026-10-19 — Correction round 1 after plan audit iteration 0 (FAIL 0.65, nine blocking defects D1-D9, receipt rcpt-39ba8947c14db9823fd47f7e). Changes: REQ-MJ-002/-003 restore the board's "same files/shared state" (D4); REQ-MJ-007 states exactly the seven limits of the ruling d314 and labels the assignment-input sentence as a consequence (D7); REQ-MJ-010 names the full measured set of standing-cron sites (D3); REQ-MJ-013 no longer claims an authoring order the commit graph cannot witness (D1); REQ-MJ-017 gains clause (e) for the leader-duty enumeration that declares itself the boundary (D8); REQ-MJ-004 is a shall clause throughout, 02f4 is quoted verbatim, the scratch-file citation is gone, the unused frontmatter key `card` is dropped (D15); decision-index Q6 is relabelled and Q7 added (D10, D11). Requirement count unchanged at 17; criteria unchanged at 19.
- 0.1.3 — 2026-10-19 — Correction round 2 after plan audit iteration 1 (FAIL 0.79, four blocking defects D16-D19 and three minor D20-D22, receipt rcpt-2adbd9564960a6f7da294042; D1-D15 closed). Changes: acceptance gains AC-MJ-020, a closed-world check of every line added to or removed from the eight changed document pairs against a pinned manifest, plus a check that no other path changes (D16); AC-MJ-008 gains checks (e) and (f) for the two cron sites that had none (D17); the cells E-002a, E-006 and E-009 are rewritten in the awk form the worktree guard accepts (D18); what Q7 does and does not gate is one sentence, the same in every document (D19); AC-MJ-001 gains tokens for the reissue half of REQ-MJ-004 and the named conflicting card of REQ-MJ-003 (D20); the commit-shape check of AC-MJ-011 decides by exit status (D21). Requirement count unchanged at 17; criteria 20 (R 12, G 7, O 1).

## 1. Context and problem

Cards added without an explicit classification fall on the fail-safe default (`mode=serial`, `decider=default`; the code constant is the notice text at `internal/cli/todo_classify.go:41`, the default decider is `factory.DefaultCardDecider{}` at line 36). On 2026-10-19 that default had piled up 29 cards on the single serial slot, and lanes ran lease-renewal loops in front of that one slot (board records a672, 02f4, 0c51 cite the counts). The guidance that tells the leader how to issue a card never asks for a serial/parallel judgment, and the lane guidance tells a lane to keep a recurring recheck cron armed "to the completion report" whatever state its card is in. This SPEC changes **documents only**: it forces the judgment at issuance, fixes the renewal and free-lane wording, and states the one case in which the leader may promote queued cards without a per-card operator pick (the operator ruling d314 of §2).

The card's scope has four items, and §4 groups the requirements by them: (1) a mode-judgment step in the card-classification guidance (§4.1, REQ-MJ-001 to REQ-MJ-003); (2) an explicit classification in the issuance procedure of the leader and the todo skill (§4.1, REQ-MJ-004, REQ-MJ-005); (3) a free-lane definition and the assignment input (§4.2, REQ-MJ-006, REQ-MJ-007, and §4.5 for the sentences that called promotion the operator's alone); (4) lease renewal for the executing card only (§4.3, REQ-MJ-008 to REQ-MJ-011). §4.4 holds the cross-cutting requirements.

## 2. Authorities (quoted, not paraphrased)

The operator's standing rule and the leader's standing rules are the policy. This SPEC cites them and adds none. Each record is read with `moai decision read` and cited by its board id; the quoted text is the board's text as printed.

**0c51** (`decided_by=operator`, `scope=standing`, `kind=standing-rule`, `d-20261019T020637Z-0c51`):

> Card issuance must carry an explicit serial/parallel judgment. The leader adds every card with --classification-file (decider=human) whose mode is decided from the expected file set: parallelizable unless it edits the same files/shared state as an in-flight card or a slot-gated surface (factory/todo core, release pipeline, shared migrations), in which case serial with the conflicting card named in reason. The reason field states the file set and the conflict check. No card is issued on the fail-safe serial default; a card that slipped through is reclassified (t1737 verb once landed; until then reissue via classification-file, as t1888→t1892, t1889→t1893). Class C cards get --spec at pick time (cannot be added after assign).

**a672** (`decided_by=leader` under operator instruction, `scope=standing`, `kind=standing-rule`, `d-20261019T014804Z-a672`):

> Renew only the card the lane is actively executing. Do NOT renew leases on (a) merge-ready cards — expiry is harmless: at the landing slot the leader reassigns or the lane re-leases with factory next --card, stage preserved (verified t1876 reassign 01:4xZ, stage=merge-ready); (b) held/operator-wait cards; (c) serial cards waiting for the slot (assigned, no lease). No polling crons for slot or trigger waits — the leader sends a notice when a slot or landing turn opens. A lane whose only cards are (a)-(c) counts as free for dispatch.

**02f4** (`decided_by=operator`, `scope=card:t1737`, `kind=ruling`, `d-20261019T015400Z-02f4`). A one-time serial-slot exception for t1737 only; it sets no policy for this card and is cited only to bound the scope of code (t1737 owns the default and the reclassify verb). Its body, verbatim:

> Operator granted t1737 (SPEC-FACTORY-DEFAULT-MODE-001, scope extended with an official reclassify verb) a one-time serial-slot exception: it takes the serial slot ahead of the waiting serial cards (t1888, t1889, t1877, t1880, t1881, t1883, t1673). Mechanics: the slot is held by any in-flight serial lease (factory_card.go:335/943), so t1882 (lane-44, plan-audit) pauses — renewals stop, its lease lapses, tree preserved — and t1737 is leased by lane-18. t1882 resumes when t1737 reaches merging. Once t1737 lands, the leader reclassifies parallel-safe serial-default cards (first t1888, t1889) with the new verb.

**d314** (`decided_by=operator (AskUserQuestion 10-19)`, `scope=card:t1894`, `kind=ruling`, `d-20261019T023441Z-d314`, `evidence_refs=lane-31 plan 55d4f1486a decision-index Q3; factory-dispatch.md:38`). Read with `moai decision read --scope card:t1894 --project-root /Users/goos/MoAI/moai-adk-go` (exit 0, `board=ok unparseable=0 records=141`) on tree `55d4f1486a454ff446be6f0f0645d4cea76bf0cc`, judged by the installed build `moai_cp/20261007_195927`:

> t1894 Q3 = leader auto-promotion is stated explicitly: while the operator has delegated dispatch (operator-away / standing delegation), when a live lane is free (no actively executing card, per a672) the leader itself promotes queued cards classified parallelizable (explicit judgment per 0c51) to picked and assigns them, without a per-card operator pick. Ordering: queue priority/readiness first; keep-set, held and [보류-marked cards are never auto-promoted; serial cards follow the slot rules. Lane self-dispatch is unchanged. Q1 default (manager-todo untouched) and Q4 (laneRecheckRule moved to t1695) stand.

## 3. Definitions

| Term | Meaning in this SPEC |
|---|---|
| mode | The classification field with the closed value set `serial` / `parallelizable` (`internal/factory/classification.go`). Independent of the card class (A / B / C). |
| slot-gated surface | The surfaces named by 0c51: factory/todo core, release pipeline, shared migrations. The board gives categories, not paths (open item OI-5). |
| in-flight card | The board's phrase (0c51). Its membership — whether a merge-ready or held card counts — is not defined by the board or by this SPEC (decision-index Q5, residual risk R-2). |
| executing | The lane is carrying out a stage of the card — plan, run, sync, the card-review, or the landing sequence — and is not waiting on a slot, a trigger, or the operator. |
| free lane | A lane on which no card is actually executing. A lane that holds only merge-ready, held (operator-wait), or serial-waiting cards is free. |
| live session | A running session occupies the lane label. A different notion from *free*: a lane with no live session is a fault, not a free lane (`factory-dispatch-mechanics.md` § Boundaries, unchanged). |
| free live lane | A lane that has a live session and on which no card is actually executing (d314: "a live lane is free (no actively executing card, per a672)"). |
| delegated dispatch | The operator's state of having handed dispatch to the leader: "operator-away / standing delegation" in d314. The ruling names no mechanism that establishes it; the documents carry its words and invent none (residual risk R-5, decision-index Q7). |
| promotion | A queued card becoming `picked`. Today it is the operator's act, in person or in advance, or the `--auto` batch approval, or a lane's own lease; d314 adds the one case of REQ-MJ-007. |
| keep-set | The cards a session never takes on its own judgment, as `auto-semantics.md` §9.3 and the `--auto` section of `gtd.md` define them. The documents cite that definition and do not restate it. |
| template side / deployed side | `internal/template/templates/.claude/...` is the source; `.claude/...` is the deployed copy of this repository (Template-First). |

## 4. Requirements (GEARS)

Home documents (each rule is stated once, in its home, and cited by section name elsewhere — REQ-MJ-012): **C** = `factory-dispatch-cards.md`, **D** = `factory-dispatch-detail.md`, **M** = `factory-dispatch-mechanics.md`. Site list and line numbers: `research.md` § Site map.

### 4.1 Mode judgment and issuance (scope items 1 and 2)

- **REQ-MJ-001** The card-classification section of C (§ Card classes) shall carry a mode-judgment step that decides a card's execution mode from the card's expected file set, as an axis independent of the card class.
- **REQ-MJ-002** While a card edits the same files or shared state as an in-flight card, or edits a slot-gated surface (factory/todo core, release pipeline, shared migrations), the mode-judgment step shall assign `serial`; in every other case it shall assign `parallelizable`.
- **REQ-MJ-003** The mode-judgment step shall require the recorded reason to state the file set and the conflict check and, when the mode is `serial`, to name the conflicting card; the conflict check shall name the shared state or the slot-gated surface it found.
- **REQ-MJ-004** When a card is issued, the issuance procedure shall require an explicit classification supplied through `--classification-file` with decider `human`, and no statement of the procedure in the leader's queue-entry clause (`factory-dispatch.md`), the todo workflow skill (`gtd.md`), or the other documents of the sweep set R (`research.md` § Sweep set R) shall present issuing on the default decider as an option. When a card has reached the queue on the default, the issuance procedure shall direct that it be reissued with an explicit classification, and the documents shall not name a reclassification verb (REQ-MJ-015).
- **REQ-MJ-005** Where a card is Class C, the issuance procedure shall state that `--spec` is given when the card is picked and cannot be added after assign, and the todo skill shall not describe an empty `spec_id` on a picked Class C card as the normal state.

### 4.2 Free lane and assignment input (scope item 3)

- **REQ-MJ-006** The documents shall define a free lane as a lane on which no card is actually executing, shall state that a lane holding only merge-ready, held, or serial-waiting cards is free, and shall keep that definition separate from the live-session condition; the sentence "free means the lane's previous card reached `done`" shall not remain.
- **REQ-MJ-007** While the operator has delegated dispatch, the documents shall authorize the leader to promote queued cards to `picked` and assign them, with no per-card operator pick, in one subsection of the detail companion (home D, `### Delegated-dispatch promotion`). The subsection shall state the limits of the ruling d314 and only those: (a) only while the operator has delegated dispatch (operator away, or a standing delegation); (b) only to a live lane on which no card is actually executing; (c) only queued cards explicitly classified `parallelizable`; (d) queue priority and readiness first; (e) a keep-set card, a held card, or a `[보류`-marked card is never promoted; (f) serial cards follow the slot rules; (g) lane self-dispatch is unchanged. The subsection shall also state, labelled as a consequence of the ruling and not as a further limit, that the assignment input of a free live lane is the picked cards plus, only while dispatch is delegated, the queued `parallelizable` cards.

### 4.3 Lease renewal (scope item 4)

- **REQ-MJ-008** A lane shall renew the lease of the card it is actually executing and of no other card; a lease on a merge-ready, held (operator-wait), or serial-waiting card may expire, expiry is harmless, and at the landing slot the leader reassigns the card or the lane re-leases it with `factory next --card`, the stage preserved.
- **REQ-MJ-009** While a lane waits on a slot or a trigger, the lane shall not arm or keep a polling cron for that wait, and the leader shall send a notice when a slot or a landing turn opens.
- **REQ-MJ-010** Each statement that arms or keeps the lane's standing recheck cron shall be scoped to the time the lane is executing the card, and shall keep the stall-recovery purpose of that cron — waking a lane stopped by an API error or by a delegate report that never comes — for an executing card. The statements are the ten sites of `research.md` §11: the always-loaded core stub; the compressed and the long-form lane-wait clauses of `factory-dispatch.md`; the intake sentence, the fire-and-renew sentence and the delete sentence of `auto-semantics.md` §5.1, and the carrier sentence of its §14 explicit-wait section, which keeps the standing cron armed during a wait on the leader; the constraint bullet of the lane-watchdog skill; the renew paragraph of M; and the cron sentence of M's own `### Lease renewal` subsection. A sentence that only names the cron and makes no claim about when it is armed or kept (`auto-semantics.md` §5 prompt introduction, §8 runner row) stays. The one-shot wait recheck of `auto-semantics.md` §14 is a different mechanism and stays as it is (residual risk R-9).
- **REQ-MJ-011** The lane-watchdog skill's first-step lease renewal (§0a) shall renew the card the lane is actually executing and shall not renew a merge-ready, held, or serial-waiting card, keeping its order (renew before measuring) for an executing card.

### 4.4 Cross-cutting

- **REQ-MJ-012** Each rule of REQ-MJ-002/-003/-005 (home C), REQ-MJ-006/-007 (home D), and REQ-MJ-008/-009 (home M) shall be stated once in its home document; every other document that needs it shall carry a pointer by section name and no restatement.
- **REQ-MJ-013** Every changed document shall be carried by the template side as the source and by the deployed copy as its byte-identical result, and each milestone commit shall carry a changed template document together with its deployed copy, except that the declared one-line fork of `factory-dispatch.md` (the worktree-sweep sentence) shall remain exactly one differing line. The order in which the two sides were edited inside a commit is an authoring discipline (Template-First) that the commit graph cannot witness; this SPEC claims only what the graph shows.
- **REQ-MJ-014** Where the always-loaded core stub is edited, it shall stay at or under 10,000 UTF-16 code units (measured today at 9,999) and the always-loaded rules subtotal under 150,000; no changed template document shall carry a card id, board record id, SPEC id, internal date, or commit SHA other than the card-id literals the leak-test baseline already lists for those files.
- **REQ-MJ-015** The documents shall not describe a changed classification default or a reclassification verb as available, and shall not alter the factual default-decider sentence of the todo skill's verb reference (`gtd.md`, the `moai todo add` row).
- **REQ-MJ-016** Where a guard test pins a changed sentence or token (the renew-interval wording, the binding ledger rows, the doctrine-section tokens, the watchdog first-step order, the `--auto` doctrine and mirror-parity tests, the neutrality tests, the card-id baseline), the change shall keep that guard green; a ledger row whose `after_text` is deliberately rewritten shall be updated in the same change.

### 4.5 Sites of the operator-act promotion rule (scope item 3, continued)

- **REQ-MJ-017** Where a document states that promotion is the operator's act, the documents shall carry the exception of REQ-MJ-007 by pointer to home D and not by restatement, and shall keep every sentence a doctrine guard pins exactly as pinned: (a) the leader's body (`factory-dispatch.md`) shall keep both forms of its promotion clause and gain one pointer line directly after them; (b) the clause that opens "The one reconciliation is named, not excepted" shall claim no exhaustiveness, with identical wording in the leader's body and in the always-loaded core stub (the stub's size stays governed by REQ-MJ-014); (c) the todo skill's pick rule (`gtd.md`) shall gain the same pointer and keep its pinned sentence; (d) the enumeration of the leader's retained duties (`factory-dispatch-mechanics.md` § The leader's remaining role), which declares itself the boundary of what the leader does, shall gain one row naming the delegated-dispatch promotion by pointer to home D; (e) the sites that `research.md` §10 decides to keep — agent files, the moai skill's index line, the `--auto` section, the pre-dispatch cross-check text, the lane watchdog's queue line, the card-pick gate of `auto-semantics.md` — shall stay unedited.

## 5. Out of scope

### Out of Scope — code and runtime

- The default mode and the reclassify verb (card t1737, lane-18): no Go change, no new verb, no change to `factory.DefaultCardDecider`, and no description of either as available (REQ-MJ-015).
- The SessionStart lane notice in `internal/hook/session_start_factory.go` / `session_start_factory_i18n.go` (`laneRecheckRule`), which injects the same "arm before the first stage, `CronDelete` at the completion report" rule in four locales and is pinned by `internal/hook/lane_renew_interval_test.go`. It is Go code; the operator ruling d314 moves it to card t1695, and it keeps the unscoped wording until that card changes it (decision-index row Q4, residual risk R-1).
- Go source text that states the operator-pick rule — `internal/cli/todo.go` lines 6, 238 and 1576 (a package comment, the `moai todo` help text "picking the next card is the operator's act too (next <n>)", and a verb comment): code, not documents, and each concerns the operator's own `next` pick, which the ruling does not change (`research.md` §10, sites P12-P14).
- The Go doctrine tests that pin the old operator-only sentences: they stay as they are in this card; folding the exception into the sentences they pin is a later card's work (residual risk R-6).

### Out of Scope — documents deliberately left alone

- `.moai/policies/**` and `.moai/workflows/gtd.md`: deploy-time projections of the rule tree (`internal/template/harness_fs.go` maps `.moai/policies` to `.claude/rules/moai`). Measured stale already on this tree (the card-size-ceiling section is absent from the policy copy of C), regenerated by the update path, not hand-edited.
- `internal/template/templates/.claude/skills/moai/workflows/project/doc-generation.md:348` (a standing-source `[PROJECT]` card issued with a bare `moai gtd add`): outside the card's named scope; open item OI-2, residual risk R-7.
- The agent files `manager-todo.md` and `manager-lead.md` and their Codex TOML emissions: measured to carry no issuance procedure (decision-index row Q1; the ruling d314 confirms "manager-todo untouched"). Their promotion wording — `manager-todo.md:54` "self-promotes, admits, drops, or edits cards" (the `--auto` cycle) and `manager-lead.md:235` (the deputy dispatches already-picked cards only) — stays: the ruling gives the leader the promotion, not the cycle and not the deputy.
- `docs-site/**`, `AGENTS.md`, `AGENTS.local.md`, `.moai/docs/**`: no statement of the changed rules found there (`research.md` § Sweep).
- The cron mentions that are not the lane's standing recheck cron: the foreman skill's queue Monitor, the cadence-bridge recipes, the harness schedule-registration steps, and the `settings.json.tmpl` permission list (`research.md` §11).

### Out of Scope — policy not decided here

- Which path set counts as a slot-gated surface, and which card states count as "in flight" for the overlap test (open item OI-5).
- What establishes "the operator has delegated dispatch" (decision-index Q7, residual risk R-5).
- The lane's "acknowledgement of receipt" of the new guidance: an operational check the leader records, not an acceptance criterion (AC-MJ-015).

## 6. Open items (decision-index rows)

The decisions the plan phase surfaced are rows Q1-Q7 in `decision-index.md`; none is decided by the planner except where the published Default rule ranks the options (stamped `DEFAULT-APPLIED`). Q3 is settled by the operator ruling quoted in §2. Q7 is open. Summary:

| Row | Question | Label | Default / state |
|---|---|---|---|
| Q1 / OI-1 | Do `manager-todo.md` / `manager-lead.md` get an issuance pointer? | FOUNDER, implementation-level | Default: no (measured: neither describes issuance) |
| Q2 / OI-2 | Is the standing-source card of `doc-generation.md` in scope? | FOUNDER, implementation-level | Default: no; the consequence (R-7) is stated |
| Q3 / OI-3 | Does "assignment input = picked + queued parallelizable" change who promotes? | FOUNDER, product-level | **Met** by the operator ruling d314 (leader promotion while dispatch is delegated, with limits); REQ-MJ-007 and REQ-MJ-017 implement it; no text waits on it |
| Q4 / OI-4 | Is the Go SessionStart lane notice in scope? | FOUNDER, implementation-level | Default: no; d314 states it is moved to card t1695 |
| Q5 / OI-5 | "In-flight" membership; paths of the three slot-gated surfaces | FOUNDER, implementation-level | Default: keep the board's words verbatim; the consequence (R-2) is stated |
| Q6 | Tier | FOUNDER, implementation-level | Default: Tier L (`design.md` D-1); no policy-coverage claim |
| Q7 | What establishes "the operator has delegated dispatch"? | FOUNDER, product-level | **Open**, empty verdict; goes to the operator; the statement below the table says what it does and does not gate |

No requirement, criterion or milestone of this SPEC waits on Q7; REQ-MJ-007 (a) is text-final because it carries the ruling's words and no mechanism. Per the decision-gate rule the empty FOUNDER row Q7 goes to the operator and blocks the autonomous form of the plan→run Kickoff until the operator answers it or states that the ruling's phrase is the whole condition.

## 7. Known limits and residual risk

- **R-1** Until card t1695 changes the Go notice, a lane that reads its SessionStart notice is told to keep the recheck cron to the completion report whatever its card state; the scoped document wording and the injected notice disagree. Mitigation: the operational check AC-MJ-015 names the notice; the owner is card t1695 by the operator ruling d314.
- **R-2** The mode judgment is a human/LLM judgment recorded in a reason field; no command here verifies that a stated file set is the card's real file set (`design.md` D-6). With "in flight" undefined (decision-index Q5) two issuers can judge the same merge-ready or held card differently; the documents make the judgment auditable (the reason states the file set and the conflict check) and not uniform.
- **R-3** `moai todo next` / assign behaviour for Class C (`--spec` at pick) is taken from board 0c51 and from `moai todo next --help` (a `--spec` flag exists on `next`); the "cannot be added after assign" fact was not re-derived from code in this plan phase.
- **R-4** The installed `moai` build (`moai_cp/20261007_195927`) predates this tree's HEAD; every CLI observation in `research.md` names it as the judging build (verification-claim-integrity §2.2).
- **R-5** The ruling d314 does not say what establishes "the operator has delegated dispatch". The documents carry its words ("operator away, or a standing delegation") and define no mechanism, record format or expiry; an actor that treats an unstated delegation as given widens the exception in practice, and no command here can see that. AC-MJ-006 pins the text, not the practice. The question is decision-index row Q7, open.
- **R-6** Two Go doctrine-test families pin the old operator-only sentences (`internal/cli/todo_auto_doc_test.go`, `internal/cli/todo_auto_pick_doc_test.go`); they were read, not run (progress.md G-2). The plan keeps the pinned sentences verbatim and adds the exception beside them, so those guards constrain the wording of this change and cannot be relaxed by it. A reader therefore meets the absolute sentence of `factory-dispatch.md` § Entry into the queue ("Outside an --auto authorization the leader never picks for the operator …") and, two lines below, a pointer line naming one further case. A later card amends the Go doctrine tests so that the exception is folded into the sentences they pin; until then the coexistence is deliberate.
- **R-7** `doc-generation.md:348` stays a bare `moai gtd add` (decision-index Q2 default). If a standing-source card counts as issued by the leader, 0c51 forbids that line, so one known 0c51 site is left open. A follow-up card is proposed in the correction-round report; whether such a card is in scope and which mode it declares is the operator's call.
- **R-8** The acceptance family is lexical. The per-criterion cells cannot see a widening or a contradiction written in words they do not match (a stem-less promotion sentence added to D passes AC-MJ-005 (d) and AC-MJ-006, observed in the iteration 1 audit and again in round 2). AC-MJ-020 closes the *added line* case: every line the run adds to or removes from the eight changed document pairs must be a row of the pinned manifest, whatever words it uses, and no other path may change. What stays open is the wording of the manifest rows themselves (the sync audit reads it), a sentence that carries a scope marker and contradicts it inside a listed line, and the practice of the lanes (AC-MJ-015). The pinned whole lines, the manifest and the sync audit's reading are the controls; none of them is a proof.
- **R-9** `auto-semantics.md` §14 keeps a one-shot recheck (`CronCreate` with `recurring: false`) while a wait on the leader is open. Whether a672's "no polling crons for slot or trigger waits" reaches that one-shot is not stated by the board; this SPEC leaves the one-shot unchanged and scopes only the standing cron. The operator may rule on it with Q7 or separately.
