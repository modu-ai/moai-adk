# Kanban Dispatch Protocol

How the **lead** session of Kanban Mode moves a card across the board: what admits work, who is told to do it, how completion is judged, and when the operator is asked to `/clear`.

> **Loading scope**: Intentionally always-loaded. A session learns it is the kanban lead from the SessionStart context, not from a file path, so a `paths:`-restricted rule would never reach it. Cost to a session that never dispatches: the stub below, restated once per turn; every procedure body lives in the lazy companions.

> **Detail companion**: `kanban-dispatch-detail.md` owns the long tables, dispatch-cycle walkthrough, incident narratives, and rationale — now also per-card fan-out, Factory in-lane 3-stage, and the `manager-lead` working mode. The stub keeps every [HARD] rule and pointer; load the companion when moving or classifying a card, or choosing review lenses.

> **Mechanics companion**: `kanban-dispatch-mechanics.md` owns the board-and-lens bodies relocated from this file — § The board (the five fixed columns and their owning roles) · § The dispatch cycle walkthrough · § Review lens selection · § Serializing a heavy run across lanes (`moai slot` lease) · § Factory Mode mechanics · § Isolation (launcher table, worktree tiers, `WT-` branch naming, the traceability carriers) · § Verification-load detail · § Integration into the release branch (the self-serve window procedure) · § Boundaries · § Cross-references. Load it when moving or classifying a card, provisioning or disposing a card worktree, running lane-local verification, or entering the integration window.

## Scope — when this rule is live

This rule binds a session whose SessionStart context declares **Kanban Mode** with the `lead` role. In every other session it is inert: a companion session (`plan` / `run` / `sync`) receives instructions, it does not dispatch them, and a session outside Kanban Mode has no board to move.

Kanban Mode is entered with `moai cc -k` (or `moai glm -k`), which elects one lead and prints one launch command per companion role. Companion sessions are launched **by hand, one per terminal** — a session cannot launch another, and no peer-spawning mechanism exists or is wanted.

The lead session works through the `manager-lead` agent: it holds the operator dialogue (the session's `AskUserQuestion` stays the user channel) while dispatching parallel work in the background — neither blocks the other. Lead and lane sessions orchestrate only; real work runs in sub-agents (design intent + spawn rules: `kanban-dispatch-detail.md` § Design intent, § The lead works through manager-lead).

One boundary: nudge delivery rides on cross-session messaging, absent on native Windows and off under some providers, versions, and flags (`cross-session-messaging.md` § Availability constraints) — an absent channel fails quietly; the lead surfaces it, and the queue keeps working without it.

## Entry into the board is an operator act

`backlog` has no owning session, so a lead admitting cards on its own initiative would be **generating** work rather than scheduling it. Every card's origin is the operator's request.

[HARD] **The lead is the queue's sole producer.** The operator asks; the lead turns the request into a card with `moai gtd add "<description>"` (`moai gtd` alone lists the queue). Production is the one queue mutation the lead performs on its own authority — translation, not invention: nothing enters the queue the operator did not ask for.

[HARD] **Standing sources are the other producers, and they produce on the operator's prior authorization.** `/moai project` issues one card when it completes (`[PROJECT] ` prefix); the codemaps-debt trigger in `moai integration release` issues one per debt period (`[GRAPH] ` prefix). The full conditions — one per occasion, derived not invented, marked, id reported, starting still a separate pick — are the SSOT at `.claude/skills/moai/workflows/gtd.md` § Standing sources, and that set is **conditional, not closed**. Nothing else produces; a report milestone, an audit finding, or an open issue still reaches the queue only as a card request the operator approves.

[HARD] **Promotion is the operator's act, always.** After a `/clear`, the lead presents the queued cards through `AskUserQuestion` and the operator picks; only then does the lead dispatch according to the card class. The lead never picks for the operator, never reorders by inferred priority, and never silently promotes a backlog item. An empty queue is a state to report, not a prompt to invent work.

A card the operator chose to start when it was issued is not a silent promotion: that answer IS the promotion, given explicitly before anything moved, and the same class-based entry follows.

[HARD] **The lead may attach a finding; it may not act on one.** Analysis records a relation between two cards (`moai gtd relate`); the record is evidence the operator reads, never a mandate — the lead never folds the related card away, never reorders the queue around it, and never drops or edits it. Analysis changes exactly one thing on its own authority: it refuses the admission of a card whose normalized text is identical to one already queued or picked.

[HARD] **The pre-dispatch PR cross-check.** Before dispatching a card out of `backlog`, the lead reads that card's pull-request and landed state (`moai gtd pr <id>`; by hand `gh pr list` plus a `git log` against the integration branch) and reports what it read in the same turn. An unchecked card is a gap, not a clean card (§ Completion is read, never trusted).

[HARD] **The cross-check also asks whether a completed SPEC already covers the work.** A card id cannot answer "has someone else already done this" — where the card names an issue or a subsystem, the lead also reads whether a covering SPEC is already `completed` and reports that alongside the PR and landed state. Neither read is conclusive: the final discriminator stays reproduction.

[HARD] **The cross-check reports; it never vetoes.** Where the card carries an open pull request or is already landed, the lead surfaces that and the operator **confirms or withdraws** it. The lead never withholds a picked card on its own authority. Why the wording is the only available control: `kanban-dispatch-detail.md` § The pre-dispatch cross-check.

## Report milestones ↔ queue cards

[HARD] **A milestone-bearing report under `.moai/reports/` carries a `## Card Cross-Check` section** — one table row per milestone, a `card` column holding the delivering card id or an explicit new-card marker. A mapping claim is verified against the queue (`moai gtd`), never remembered. Before the lead turns a report into card requests, the request message states the full comparison — `N milestones → N cards` — naming every milestone with no card in the live queue. Detail: `kanban-dispatch-detail.md` § Report milestones ↔ queue cards.

## Card classes — not every card needs every column

The lead classifies each card as it leaves `backlog` and names the class in the dispatch: **A — direct close** (one file, one line, no design judgement, CI catches the regression; `plan` skipped), **B — defect, cause unknown** (`run → sync`; no SPEC exists), **C — design change** (a decision, or spans subsystems; all three working columns). Full table and rationale: `kanban-dispatch-detail.md` § Card classes.

[HARD] **Class A is admitted on checked evidence, not on an assertion.** Two of its three properties are mechanically checked and cited: the diff is measured (`git diff --stat` against the base, showing the one file) and CI is green **on the head that will merge**. The third — no design judgement in it — is a judgement, stated in the dispatch where the operator can disagree with it. A card that cannot cite both measurements is not Class A. The justification is never "it is faster".

**Class B skips `plan`, not the sync gate's review** — the `run` session owns the investigation; the cause-establishing evidence goes into the card's progress record before the card leaves `run`, and the completion report names that path.

**Work in progress: one card per worktree** — two cards sharing a worktree run serially whatever columns they sit in. A lane holding several cards in one column does it by per-card fan-out (`kanban-dispatch-detail.md` § Per-card fan-out).

## The dispatch cycle

### The delegation channel is the queue

[HARD] Work is delegated through the queue on disk, not through messages. The queue file resolves against the primary checkout from every linked worktree — one repository, one queue — so a card admitted from anywhere is visible everywhere, and that single-file visibility is what makes the queue a channel rather than a shared opinion.

A cross-session message is a nudge, never the delegation itself. **A lane can be asked to report when it next goes idle** (`SendMessage` `notify_when_idle`), but [HARD] the notice is not the completion signal — it says *when to read the evidence*, nothing about what it says. **A nudge can be refused outright under fan-out**; read the send result, and a refusal costs the board nothing because the queue already carries the delegation (`cross-session-messaging.md` § An idle notice is a scheduling hint, § Configuration surface).

### Dispatch language

[HARD] A dispatch is written in the operator's `conversation_language` — the operator watches it scroll past, which makes it user-facing output rather than internal agent traffic. The boundary is **who reads it**: an `Agent()` subagent prompt reaches no human and stays English. What stays verbatim in every language: SPEC IDs, command names and their flags, file paths, session names, and technical identifiers. (Classification rationale: `kanban-dispatch-detail.md` § Dispatch language.)

### Dispatch format

[HARD] A dispatch is a fixed-field address block, not prose. The fields:

```
card: <id>
spec: <SPEC-ID>
cmd: /moai run <SPEC-ID>
wt: .moai/worktrees/<name>
evidence: .moai/specs/<SPEC-ID>/progress.md
lens: --security --deep
```

- `card`, `cmd`, `wt`, and `evidence` are always present. `spec` joins once a SPEC exists; a Class B card carries none, and its `evidence` names whatever record the lead will read instead. `lens` appears only in a `sync` dispatch.
- `wt` names the new card's worktree, never a previous card's tree. The tree keeps the card id; the branch takes a descriptive slug (§ Isolation below).
- **No explanatory prose.** Procedure, background, and justification live in the card text and the SPEC artifacts the block points at.
- **Ceiling: the block is at most 10 lines.** A dispatch that does not fit is trying to be a handoff; move the payload into the card and send the block.
- **[HARD] The send is read, not assumed.** A `routing` object on the result means an in-process mailbox took the block — lost (re-send to `name [ref]`); a following `[Cross-session delivery notice]` means the lane's permission policy is holding or refused it (surface it to the operator, do not re-send); anything else queued it. None of the three establishes that the lane's Claude read it. Full shape table: `cross-session-messaging.md` § A send result has three shapes.

## Deputy dispatch surface

[HARD] **The deputy is resident, not optional.** Before the batch's first lane dispatch, the `-k`/`-f` lead session spawns exactly one UNNAMED background `Agent()` running manager-lead as its **coordination deputy** and keeps it for the batch. Its delegable/retained matrix lives in the agent itself (`manager-lead.md` § Deputy dispatch surface); this stub carries only the boundary that binds the board.

[HARD] **A completion report reaches the lead as a `RECOMMEND:` summary, not as raw reading batches** — turn occupancy moves; the lead's own evidence-read before advancing a card does not (§ Completion is read, never trusted).

[HARD] **The deputy never holds a power of consequence.** Final PASS/FAIL verdicts, final merge approval (`LEAD-MERGE-APPROVED`), operator gates, card issuance and `done` (`moai gtd` mutations), CodeRabbit slot-wait adjudication, and cross-session dispute coordination stay with the lead session. Round-report measurement and drafting are the deputy's; the asserted figures are the lead's — every figure names its measurer. Nothing structural moves with the delegation: the queue stays the channel, completion stays evidence the lead read, and the verdict's home stays the lead.

## Completion is read, never trusted

[HARD] The lead advances a card on **evidence it read**, not on a companion's reply. Before moving a card out of a working column, the lead reads the card's `progress.md` and the verification evidence path the phase declares; a missing, unreadable, or stale evidence file is a **gap** — the card stays put and the lead reports why. Absence of a failure signal is not a pass.

Where the phase's declared evidence includes an audit verdict, the lead reads the verdict **file** under `.moai/reports/<card-id>/` per `.moai/docs/audit-artifact-convention.md`; an absent, unreadable, or uncommitted verdict file is a gap exactly like a missing progress record.

**The final PASS/FAIL verdict is the lead's**, read from the evidence on disk and never delegated to the lane that produced the work. Why the division is structural: `kanban-dispatch-detail.md` § The verdict's home.

### CodeRabbit is not read from `gh pr checks`

[HARD] A `gh pr checks` row naming CodeRabbit is not evidence that a review happened: the status is `success` **even when no review ran**, and the row prints `pass` byte-identically in both cases. A row counts only when BOTH hold:

1. The **combined** endpoint `/commits/{sha}/status` shows `state == "success"` **and** description `Review completed`:

    ```bash
    gh api "repos/$REPO/commits/$HEAD_SHA/status" \
      --jq '.statuses[] | select(.context == "CodeRabbit" and .state == "success") | .description'
    ```

2. A `Merge Risk:` line exists whose commit prefix matches the current `headRefOid`.

Anything else is a gap, not a pass. `Review rate limited` means the review never started, and a card carrying it does not leave `sync`. (Endpoint choice: `kanban-dispatch-detail.md` § CodeRabbit endpoint measurement.)

## The `/clear` handoff between phases

[HARD] A companion session does not carry one card's context into the next card. When a phase completes and the lead has read its evidence, the lead **asks the operator to `/clear` that session** — `/clear` is a user-typed command and cannot be sent as an instruction. The lead's message states, in order: what closed (card, phase, evidence read), which session to `/clear` (by name), and what happens next.

Where the next phase reuses a just-cleared session, the lead re-sends the full pointer instruction rather than assuming the session remembers.

The lead's own session is cleared the same way, between cards rather than phases: once a card reaches `done`, the operator is asked to `/clear` the lead session, and the next turn presents the queue again.

## Isolation is provisioned by MoAI, then entered through a launcher

[HARD] A card's work happens inside a worktree, and that worktree is **entered through the launcher** — never created with a bare `git worktree add`. Launcher table (per-harness forms): `kanban-dispatch-mechanics.md` § Isolation.

[HARD] A Codex factory agent must never call `moai cc -w`, `EnterWorktree`, or `ExitWorktree`; those are Claude Code entry mechanisms.

[HARD] **`moai worktree done` closes L2 trees only** — project trees under `.claude/worktrees/` and `.moai/worktrees/` are L1 and absent from its registry. The full L1/L2 boundary lives in `worktree-integration.md` § Terminology Glossary.

[HARD] **The card's branch is unpushed, so its worktree is the work's only instance.** Dispose of no worktree — L1 or L2 — until the lead has integrated the branch and the remote merge has landed.

[HARD] **A new card starts in a new worktree** — a lane anchored in the previous card's tree MUST `ExitWorktree` before entering the next one; the fresh tree is created from the configured base, never reused. Where the new card depends on a prior card's unmerged code, merge that branch inside the new worktree.

[HARD] **Card worktree branches carry the `WT-` prefix and a descriptive slug — never the card id** (rename in place with `git branch -m WT-<slug>`); the worktree directory keeps the card id, and traceability rests on the dispatch `card:` field, the commit message, the evidence path, and the PR title — all mandatory. Slug shape and carrier detail: `kanban-dispatch-mechanics.md` § Isolation · `kanban-dispatch-detail.md` § The PR-title carrier.

[HARD] **Inside a worktree session that `<path>` is the worktree's own absolute path** — the guard refuses `-C .`, relative paths, runtime-computed paths, and paths outside this worktree; plain `git`, `git -C <own absolute path>`, and `--git-dir=<own .git>` pass.

## Verification load is lane-local

[HARD] **Lane-local verification is scoped to the card.** A lane runs the tests its own change can affect, then pushes and lets CI run the full suite. (Incident record: `kanban-dispatch-detail.md` § Verification load incident record.)

[HARD] **Never spawn background load.** Where a verification genuinely needs contention, the load must be cleanup-guaranteed — kills registered with the test framework's cleanup hook, or a `timeout` wrapper bounding the process from outside.

### The env-isolated verification form

[HARD] Inside a worktree, an environment-scrubbed verification runs as one compound `unset … && <command>` invocation — each Bash call is a fresh process, so a separate `unset` does not carry into the next command; the scrub and the command travel together, or the scrub does nothing:

```bash
unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test ./...
```

Subshell and `env -u` variants, their measured refusal shapes, and the script-file bypass hazard: `kanban-dispatch-mechanics.md` § Verification load is lane-local.

## Integration into the release branch is self-served

[HARD] A lane whose card has passed verification does not wait for the lead to integrate it: the lane merges its own branch into the batch's release branch (`release/vX.Y.Z`) itself. The window is taken with `moai integration acquire --name <lane> --card <card-id>` BEFORE entering the release worktree, released after the completion report is sent; the lane enters the release worktree with `EnterWorktree` (a cross-tree `git -C` is refused), merges `--no-ff`, re-reads `HEAD` before the commit and again before the push, pushes `release/vX.Y.Z` (never force), and leaves the batch pull request with the lead. The full window procedure: `kanban-dispatch-mechanics.md` § Integration into the release branch is self-served · the `acquire` settings-drift assertion: `kanban-dispatch-detail.md` § The pre-merge settings-drift assertion.

## Factory Mode — the card travels whole

`moai cc -f <N>` launches one lead plus lane sessions labelled `worker-1..worker-N`; the lead routes each card WHOLE to a free lane, which carries it `plan → run → sync` in-session and owns it end to end. Mechanics, label convention, and the A/B/C collapse: `kanban-dispatch-mechanics.md` § Factory Mode mechanics · `kanban-dispatch-detail.md` § Factory in-lane 3-stage.

**Lane spawn authority (standing).** A lane is an orchestrator for its card: it spawns, with the Agent tool and WITHOUT asking, the specialist the Status Transition Ownership Matrix names for the stage at hand — plan-phase artifacts to `manager-spec`, implementation to `manager-develop`, sync-phase docs to `manager-docs`, plus the chain's prescribed auditors. Depth-1 only: agents a lane spawns are leaf workers and never spawn further agents. This authority is part of the lane's bootstrap context (the SessionStart join notice carries it verbatim), and it is deliberately NOT a per-dispatch grant: the runtime's default "don't spawn unless the user asks" guidance does not bind a lane, and a lead's approval can neither grant nor revoke what the bootstrap already grants — the lead is not the lane's user. The same authority binds kanban companion sessions.

## Boundaries — what this protocol does not do

- **No gate bypass.** Kickoff approval before run-phase entry, and every other approval gate, is unchanged by being inside a dispatch cycle.
- **No question delegation.** Companion sessions return blocker reports; the operator is asked by the lead, through `AskUserQuestion`.

The three remaining boundaries — no board state store, no session spawning, and an empty role being a fault rather than a wait: `kanban-dispatch-mechanics.md` § Boundaries.

## Cross-references

- `.claude/rules/moai/core/askuser-protocol.md` — the question channel the lead uses for card selection and `/clear` prompts
- `.claude/rules/moai/core/verification-claim-integrity.md` — why completion is read rather than trusted
- The remaining four (blocker-report format, worktree tiers, the queue surface, `manager-lead`): `kanban-dispatch-mechanics.md` § Cross-references

---

Classification: Evolvable operational rule — applies to the lead session of Kanban Mode. Detail companion: `kanban-dispatch-detail.md` (stub + lazy-companion split).