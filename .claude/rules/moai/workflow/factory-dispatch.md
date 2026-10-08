---
description: "Role-gated dispatch protocol for Factory Mode. The full body stays in place for role-injection delivery: factory leader and lane sessions receive the role core through the SessionStart hook, and the unmarked leader-grade entry points carry a read-first directive for this file. The always-loaded stub is factory-dispatch-core.md"
paths: "**/factory-dispatch*.md,**/.claude/agents/moai/manager-lead.md,**/.claude/skills/moai-factory-foreman/SKILL.md,**/.claude/skills/moai/workflows/gtd.md"
---

# Factory Dispatch Protocol

> Moved to the detail companion: `factory-dispatch-mechanics.md` ("Factory Dispatch Protocol").

> **Loading scope**: Role-gated — delivered by injection, not session-start loading. Factory leader and lane sessions receive this rule's role core through the SessionStart hook (sources startup, clear, compact); the unmarked leader-grade entry points — the manager-lead agent, the factory foreman skill, the todo `--auto` path — carry a read-first directive for this full body. The always-loaded stub is `factory-dispatch-core.md`; the top-level `paths:` key is a non-delivery placement.

> **Detail companion**: `factory-dispatch-detail.md` owns the long tables, dispatch-cycle walkthrough, and coordination rationale — also per-card sub-agent execution, the Factory in-lane 3-stage, and the `manager-lead` working mode. Sibling companions: `factory-dispatch-cards.md` (card classification, traceability, the pre-dispatch cross-check) and `factory-dispatch-gates.md` (sync-gate review lenses, the CodeRabbit measurement, the settings-drift assertion, the verification-load incident record). The stub keeps every [HARD] rule and pointer; load a companion when classifying a card, routing one, or choosing review lenses.

## Scope — when this rule is live

> Moved to the detail companion: `factory-dispatch-mechanics.md` ("Scope — when this rule is live").

Factory Mode is entered with `moai cc -f` (or `moai glm -f`), which elects one leader; lane sessions join one at a time with `moai cc -l` (or `moai glm -l`) and are labelled `lane-<n>`. Lane sessions are launched **by hand, one per terminal** — a session cannot launch another, and no peer-spawning mechanism exists or is wanted.

A lane carries its card from dispatch to landing autonomously: its in-card judgments are its own (the ladder ends at the lane's judgment, `auto-semantics.md` §6), the landing sequence is the lane's to run, and the lane closes its own card at completion. What the leader keeps is enumerated, not inherited: `factory-dispatch-mechanics.md` § The leader's remaining role. The standard landing every deployed lane runs: `factory-dispatch-mechanics.md` § The lane's standard landing.

The leader session works through the `manager-lead` agent: it holds the operator dialogue (the session's `AskUserQuestion` stays the user channel) while dispatching parallel work in the background — neither blocks the other. Lead and lane sessions orchestrate only; real work runs in sub-agents (design intent + spawn rules: `factory-dispatch-detail.md` § Design intent, § The leader works through manager-lead).

One boundary: nudge delivery rides on cross-session messaging, absent on native Windows and off under some providers, versions, and flags (`cross-session-messaging.md` § Availability constraints) — an absent channel fails quietly; the leader surfaces it, and the queue keeps working without it.

## Entry into the queue is an operator act

> Moved to the detail companion: `factory-dispatch-mechanics.md` ("Entry into the queue is an operator act").

<!-- moai:role-core-start -->
[HARD] **The leader is the queue's sole producer.** The operator asks; the leader turns the request into a card with `moai gtd add "<description>"` (`moai gtd` alone lists the queue). Production is the one queue mutation the leader performs on its own authority — translation, not invention: nothing enters the queue the operator did not ask for.
<!-- moai:role-core-end -->

<!-- moai:role-core-start -->
[HARD] **Standing sources are the other producers, and they produce on the operator's prior authorization.** `/moai project` issues one card when it completes (`[PROJECT] ` prefix); the codemaps-debt trigger in `moai integration release` issues one per debt period (`[GRAPH] ` prefix). The full conditions — one per occasion, derived not invented, marked, id reported, starting still a separate pick — are the SSOT at `.claude/skills/moai/workflows/gtd.md` § Standing sources, and that set is **conditional, not closed**. Nothing else produces; a report milestone, an audit finding, or an open issue still reaches the queue only as a card request the operator approves.
<!-- moai:role-core-end -->

<!-- moai:role-core-start -->
[HARD] **Promotion is the operator's act, in person or in advance.** After a `/clear`, the leader presents the queued cards through `AskUserQuestion` and the operator picks; only then does the leader route the card whole to a free lane. Outside an --auto authorization the leader never picks for the operator, never reorders by inferred priority, and never silently promotes a backlog item. An empty queue is a state to report, not a prompt to invent work. The one ranking the queue admits is the `--auto` cycle's own, bounded in the next paragraph.
<!-- moai:role-core-end -->

The one reconciliation is named, not excepted: `/moai:todo --auto` is the operator's batch approval: it authorizes the invoked session to take cards from the queue on its own judgment, and nothing else; a lane session only through a lease (`moai factory next --card <id>`), never a keep-set card. The `--auto` cycle derives its authority solely from that invocation, never from queue emptiness, card readiness, or a peer's request. The cycle carries one auto-scoped ranking exception: it may rank the queued candidates it is about to accept (a Jev signal when the capability is available, else recorded priority and readiness), which changes its selection order only — it never reorders, admits, drops or edits cards, the queue itself is unchanged, and the invocation remains the operator's batch approval. Detail: the gtd workflow's `--auto` section. That batch authorization is the card-pick gate's AUTONOMOUS form (`.claude/rules/moai/workflow/auto-semantics.md` §9.3) — the `--auto` invocation IS the approval; queue ADMISSION (production) stays the operator's.

<!-- moai:role-core-start -->
[HARD] **The self-dispatch lane exception.** In a self-dispatch factory run, a lane session may lease the next queued card — the one promotion a lane performs — only through `moai factory next` — bare, or `--card <id>` for its own judged pick. Every other queue mutation (`add`, `drop`, `done`, `edit`, and the rest) and `moai contract sign` stay forbidden to a lane: the lane works the operator's queue, it never authors it.
<!-- moai:role-core-end -->

A card the operator chose to start when it was issued is not a silent promotion: that answer IS the promotion, given explicitly before anything moved, and the same whole-card routing follows.

<!-- moai:role-core-start -->
[HARD] **The leader may attach a finding; it may not act on one.** Analysis records a relation between two cards (`moai gtd relate`); the record is evidence the operator reads, never a mandate — the leader never folds the related card away, never reorders the queue around it, and never drops or edits it. Analysis changes exactly one thing on its own authority: it refuses the admission of a card whose normalized text is identical to one already queued or picked.
<!-- moai:role-core-end -->

<!-- moai:role-core-start -->
[HARD] **The pre-dispatch PR cross-check.** Before dispatching a card out of `backlog`, the leader reads that card's pull-request and landed state (`moai gtd pr <id>`; by hand `gh pr list` plus a `git log` against the integration branch) and reports what it read in the same turn. An unchecked card is a gap, not a clean card (§ Completion is read, never trusted).
<!-- moai:role-core-end -->

<!-- moai:role-core-start -->
[HARD] **The cross-check also asks whether a completed SPEC already covers the work.** A card id cannot answer "has someone else already done this" — where the card names an issue or a subsystem, the leader also reads whether a covering SPEC is already `completed` and reports that alongside the PR and landed state. Neither read is conclusive: the final discriminator stays reproduction.
<!-- moai:role-core-end -->

<!-- moai:role-core-start -->
[HARD] **The cross-check reports; it never vetoes.** Where the card carries an open pull request or is already landed, the leader surfaces that and the operator **confirms or withdraws** it. The leader never withholds a picked card on its own authority. Why the wording is the only available control: `factory-dispatch-cards.md` § The pre-dispatch cross-check.
<!-- moai:role-core-end -->

## Report milestones ↔ queue cards

<!-- moai:role-core-start -->
[HARD] **A milestone-bearing report under `.moai/reports/` carries a `## Card Cross-Check` section** — one table row per milestone, a `card` column holding the delivering card id or an explicit new-card marker. A mapping claim is verified against the queue (`moai gtd`), never remembered. Before the leader turns a report into card requests, the request message states the full comparison — `N milestones → N cards` — naming every milestone with no card in the live queue. Detail: `factory-dispatch-cards.md` § Report milestones ↔ queue cards.
<!-- moai:role-core-end -->

## Card classes — not every card needs every stage

The leader classifies each card as it leaves `backlog` and names the entry stage in the dispatch: **A — direct close** (one file, one line, no design judgement, CI catches the regression; `plan` skipped), **B — defect, cause unknown** (`run → sync`; no SPEC exists), **C — design change** (a decision, or spans subsystems; all three stages). Full table and rationale: `factory-dispatch-cards.md` § Card classes.

<!-- moai:role-core-start -->
[HARD] **Class A is admitted on checked evidence, not on an assertion.** Two of its three properties are mechanically checked and cited: the diff is measured (`git diff --stat` against the base, showing the one file) and CI is green **on the head that will merge**. The third — no design judgement in it — is a judgement, stated in the dispatch where the operator can disagree with it. A card that cannot cite both measurements is not Class A. The justification is never "it is faster".
<!-- moai:role-core-end -->

**Class B skips `plan`, not the sync gate's review** — the lane owns the investigation; the cause-establishing evidence goes into the card's progress record before the card leaves `run`, and the completion report names that path.

**Work in progress: one card per worktree** — two cards sharing a worktree run serially.

## The dispatch cycle

### The delegation channel is the queue

<!-- moai:role-core-start -->
[HARD] Work is delegated through the queue on disk, not through messages. The queue file resolves against the primary checkout from every linked worktree — one repository, one queue — so a card admitted from anywhere is visible everywhere, and that single-file visibility is what makes the queue a channel rather than a shared opinion.
<!-- moai:role-core-end -->

<!-- moai:role-core-start -->
A cross-session message is a nudge, never the delegation itself. **A lane can be asked to report when it next goes idle** (`SendMessage` `notify_when_idle`), but [HARD] the notice is not the completion signal — it says *when to read the evidence*, nothing about what it says. **A nudge can be refused outright under fan-out**; read the send result, and a refusal costs the run nothing because the queue already carries the delegation (`cross-session-messaging.md` § An idle notice is a scheduling hint, § Configuration surface).
<!-- moai:role-core-end -->

### Dispatch language

<!-- moai:role-core-start -->
[HARD] A dispatch is written in the operator's `conversation_language` — the operator watches it scroll past, which makes it user-facing output rather than internal agent traffic. The boundary is **who reads it**: an `Agent()` subagent prompt reaches no human and stays English. What stays verbatim in every language: SPEC IDs, command names and their flags, file paths, session names, and technical identifiers. (Classification rationale: `factory-dispatch-detail.md` § Dispatch language.)
<!-- moai:role-core-end -->

### Dispatch format

<!-- moai:role-core-start -->
[HARD] A dispatch is a fixed-field address block, not prose. The fields:

```
card: <id>
spec: <SPEC-ID>
cmd: /moai run <SPEC-ID>
wt: .moai/worktrees/<name>
evidence: .moai/specs/<SPEC-ID>/progress.md
lens: --security --deep
```

- `card`, `cmd`, `wt`, and `evidence` are always present. `spec` joins once a SPEC exists; a Class B card carries none, and its `evidence` names whatever record the leader will read instead. `lens` appears only when the leader names review lenses for the card's sync gate.
- `wt` names the new card's worktree, never a previous card's tree. The tree keeps the card id; the branch takes a descriptive slug (§ Isolation below).
- **No explanatory prose.** Procedure, background, and justification live in the card text and the SPEC artifacts the block points at.
- **Ceiling: the block is at most 10 lines.** A dispatch that does not fit is trying to be a handoff; move the payload into the card and send the block.
- **[HARD] The send is read, not assumed.** A `routing` object on the result means an in-process mailbox took the block — lost (re-send to `name [ref]`); a following `[Cross-session delivery notice]` means the lane's permission policy is holding or refused it (surface it to the operator, do not re-send); anything else queued it. None of the three establishes that the lane's Claude read it. Full shape table: `cross-session-messaging.md` § A send result has three shapes.
<!-- moai:role-core-end -->

### The lane's task list carries the card's stages

<!-- moai:role-core-start -->
[HARD] A lane session tracks its card's execution on the session's task tools. At card intake it registers the card's execution stages via `TaskCreate` before beginning the first stage; at every stage transition it keeps the list current via `TaskUpdate` so the list always shows the stage in flight; and it reports completion only while the list reflects the end state — or carries an explicit annotation naming why it does not. A task list that contradicts its completion report is the same gap as a missing evidence file (§ Completion is read, never trusted).
<!-- moai:role-core-end -->

> Moved to the detail companion: `factory-dispatch-mechanics.md` ("The lane's task list carries the card's stages").

<!-- moai:role-core-start -->
[HARD] Between run-exit verification and integration every lane runs a `card-review` stage: a card-scope self-review through `codex_review` (`glm_review` optional) whose result is written to `.moai/reports/<card-id>/card-review.md` — advisory only, never a replacement for the leader's evidence read or an independent audit. With `tree_scope: skip` configured for the leader's checkout, the leader session carries no turn-end codex review gate and reviews its own internal output directly with the same tools. Stage order, evidence fields, and the re-review ceiling: `factory-dispatch-detail.md` § The card-review stage.
<!-- moai:role-core-end -->

### Lane waits are explicit, and stalls are watched

<!-- moai:role-core-start -->
[HARD] A lane that cannot proceed records an explicit wait on disk — reason, whom, recheck point — and ends its turn; it never idles open-ended on a reply. On its next awaken the lane runs the stall watchdog first — invoke Skill("moai-lane-watchdog") and follow it — which measures progress, classifies the stall cause, and resolves the judgment through the decision ladder (doctrine: `.claude/rules/moai/workflow/auto-semantics.md` §6). The ladder's terminal step is the lane's own judgment + decision record — the lane never asks the leader and waits on a judgment; reaching the leader is a gate boundary (a keep-set category or a cross-card conflict), not a ladder step. The lead's decision board carries leader-level rulings only — cross-card conflicts, serial-slot policy calls, keep-set gate relays — never a lane's in-card judgment. A lane also arms a standing recheck cron (`CronCreate`, recurring, every 20 minutes) before its first stage and keeps it until its completion report: a lane stopped by an API error (429) or by a delegate's report that never comes wakes without outside help only if something armed before it stopped fires, and every wake reads the disk evidence (progress record, reports, commits, the delegate's deliverable) before it replies (`.claude/rules/moai/workflow/auto-semantics.md` §5.1).
<!-- moai:role-core-end -->

## Deputy dispatch surface

<!-- moai:role-core-start -->
[HARD] **The deputy is resident, not optional.** Before the batch's first lane dispatch, the factory leader session spawns exactly one UNNAMED background `Agent()` running manager-lead as its **coordination deputy** and keeps it for the batch. Its delegable/retained matrix lives in the agent itself (`manager-lead.md` § Deputy dispatch surface); this stub carries only the boundary that binds the queue.
<!-- moai:role-core-end -->

<!-- moai:role-core-start -->
[HARD] **A completion report reaches the leader as a `RECOMMEND:` summary, not as raw reading batches** — turn occupancy moves; the leader's own evidence-read before advancing a card does not (§ Completion is read, never trusted).
<!-- moai:role-core-end -->

<!-- moai:role-core-start -->
[HARD] **The deputy never holds a power of consequence.** Final PASS/FAIL verdicts, final merge approval (`LEAD-MERGE-APPROVED`), operator gates, card issuance and `done` (`moai gtd` mutations), CodeRabbit slot-wait adjudication, and cross-session dispute coordination stay with the leader session. Round-report measurement and drafting are the deputy's; the asserted figures are the leader's — every figure names its measurer. Nothing structural moves with the delegation: the queue stays the channel, completion stays evidence the leader read, and the verdict's home stays the lead.
<!-- moai:role-core-end -->

## Completion is read, never trusted

<!-- moai:role-core-start -->
[HARD] The leader advances a card on **evidence it read**, not on a lane's reply. Before moving a card out of a working stage, the leader reads the card's `progress.md` and the verification evidence path the phase declares; a missing, unreadable, or stale evidence file is a **gap** — the card stays put and the leader reports why. Absence of a failure signal is not a pass.
<!-- moai:role-core-end -->

Where the phase's declared evidence includes an audit verdict, the leader reads the verdict **file** under `.moai/reports/<card-id>/` per `.moai/docs/audit-artifact-convention.md`; an absent, unreadable, or uncommitted verdict file is a gap exactly like a missing progress record.

For a lane card the declared evidence list also carries `.moai/reports/<card-id>/card-review.md`; a card whose progress record neither cites a readable `card-review.md` nor records a reason for its absence is a gap and stays in its stage.

**The final PASS/FAIL verdict is the leader's**, read from the evidence on disk and never delegated to the lane that produced the work. Why the division is structural: `factory-dispatch-detail.md` § The verdict's home.

### CodeRabbit is not read from `gh pr checks`

<!-- moai:role-core-start -->
[HARD] A `gh pr checks` row naming CodeRabbit is not evidence that a review happened: the status is `success` **even when no review ran**, and the row prints `pass` byte-identically in both cases. A row counts only when BOTH hold:

1. The **combined** endpoint `/commits/{sha}/status` shows `state == "success"` **and** description `Review completed`:

    ```bash
    gh api "repos/$REPO/commits/$HEAD_SHA/status" \
      --jq '.statuses[] | select(.context == "CodeRabbit" and .state == "success") | .description'
    ```

2. A `Merge Risk:` line exists whose commit prefix matches the current `headRefOid`.
<!-- moai:role-core-end -->

Anything else is a gap, not a pass. `Review rate limited` means the review never started, and a card carrying it does not leave `sync`. (Endpoint choice: `factory-dispatch-gates.md` § CodeRabbit endpoint measurement.)

## The `/clear` handoff between cards

<!-- moai:role-core-start -->
[HARD] A lane session does not carry one card's context into the next card. When a card reaches `done` and the leader has read its evidence, the leader **asks the operator to `/clear` that lane session** — `/clear` is a user-typed command and cannot be sent as an instruction. The leader's message states, in order: what closed (card, evidence read), which session to `/clear` (by name), and what happens next.
<!-- moai:role-core-end -->

<!-- moai:role-core-start -->
[HARD] On a card transition inside a lane session, `/clear` happens exactly once, **after** the session has moved into the next card's worktree — not before the move. The count is not reduced: one card transition, one `/clear`.
<!-- moai:role-core-end -->

Where the next card reuses a just-cleared lane, the leader re-sends the full pointer instruction rather than assuming the session remembers.

The leader's own session is cleared the same way, between cards: once a card reaches `done`, the operator is asked to `/clear` the leader session, and the next turn presents the queue again.

## Isolation is provisioned by MoAI, then entered through a launcher

<!-- moai:role-core-start -->
[HARD] A card's work happens inside a worktree, and that worktree is **entered through the launcher** — never created with a bare `git worktree add`. Launcher table (per-harness forms): `factory-dispatch-mechanics.md` § Isolation.
<!-- moai:role-core-end -->

<!-- moai:role-core-start -->
[HARD] A Codex factory agent must never call `moai cc -w`, `EnterWorktree`, or `ExitWorktree`; those are Claude Code entry mechanisms.
<!-- moai:role-core-end -->

<!-- moai:role-core-start -->
[HARD] **`moai worktree done` closes L2 trees only** — project trees under `.claude/worktrees/` and `.moai/worktrees/` are L1 and absent from its registry. The full L1/L2 boundary lives in `worktree-integration.md` § Terminology Glossary.
<!-- moai:role-core-end -->

<!-- moai:role-core-start -->
[HARD] **The card's branch is unpushed, so its worktree is the work's only instance.** Dispose of no worktree — L1 or L2 — until the leader has integrated the branch and the remote merge has landed.
<!-- moai:role-core-end -->

<!-- moai:role-core-start -->
[HARD] **A card session starts inside its worktree and stays there.** The standard is a lane session launched inside the card's worktree through the launcher (`moai cc -w`); it starts with that tree's project skill set. Moving a running card session from one card worktree into another card worktree is prohibited — that transition attaches a second skill listing on top of the session's first, and that duplication is what this rule exists to prevent. Two composition rules bound the prohibition rather than leaving it implicit: (a) the "A new card starts in a new worktree" clause below composes with it — its exit-first requirement is satisfied by ending the lane session and launching the next session inside the new tree, never by moving the running session; (b) the brief entry into the release/integration worktree that the same files mandate for merging a card branch is NOT covered by the prohibition — it keeps its own re-entry rule (return to the card worktree before continuing card work). Where a worktree move is unavoidable anyway, `/clear` is issued exactly once, after the move — never before it.
<!-- moai:role-core-end -->

<!-- moai:role-core-start -->
[HARD] **A new card starts in a new worktree** — a lane anchored in the previous card's tree MUST `ExitWorktree` before entering the next one; the fresh tree is created from the configured base, never reused. Where the new card depends on a prior card's unmerged code, merge that branch inside the new worktree.
<!-- moai:role-core-end -->

<!-- moai:role-core-start -->
[HARD] **Card worktree branches carry the `WT-` prefix and a descriptive slug — never the card id** (rename in place with `git branch -m WT-<slug>`); the worktree directory keeps the card id, and traceability rests on the dispatch `card:` field, the commit message, the evidence path, and the PR title — all mandatory. Slug shape and carrier detail: `factory-dispatch-mechanics.md` § Isolation · `factory-dispatch-cards.md` § The PR-title carrier.
<!-- moai:role-core-end -->

<!-- moai:role-core-start -->
[HARD] **Inside a worktree session that `<path>` is the worktree's own absolute path** — the guard refuses `-C .`, relative paths, runtime-computed paths, and paths outside this worktree; plain `git`, `git -C <own absolute path>`, and `--git-dir=<own .git>` pass.
<!-- moai:role-core-end -->

## Verification load is lane-local

<!-- moai:role-core-start -->
[HARD] **Lane-local verification is scoped to the card.** A lane runs the tests its own change can affect, then pushes and lets CI run the full suite. (Incident record: `factory-dispatch-gates.md` § Verification load incident record.)
<!-- moai:role-core-end -->

<!-- moai:role-core-start -->
[HARD] **Never spawn background load.** Where a verification genuinely needs contention, the load must be cleanup-guaranteed — kills registered with the test framework's cleanup hook, or a `timeout` wrapper bounding the process from outside.
<!-- moai:role-core-end -->

### The env-isolated verification form

<!-- moai:role-core-start -->
[HARD] Inside a worktree, an environment-scrubbed verification runs as one compound `unset … && <command>` invocation — each Bash call is a fresh process, so a separate `unset` does not carry into the next command; the scrub and the command travel together, or the scrub does nothing:

```bash
unset MOAI_KANBAN_ID MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test ./...
```
<!-- moai:role-core-end -->

> Moved to the detail companion: `factory-dispatch-mechanics.md` ("The env-isolated verification form").

## Integration into the release branch is self-served

> Moved to the detail companion: `factory-dispatch-mechanics.md` ("Integration into the release branch is self-served").

<!-- moai:role-core-start -->
[HARD] A lane whose card has passed verification does not wait for the leader to integrate it: the lane merges its own branch into the batch's release branch (`release/vX.Y.Z`) itself. The window is taken with `moai integration acquire --name <lane> --card <card-id>` BEFORE entering the release worktree, released after the completion report is sent; the lane enters the release worktree with `EnterWorktree` (a cross-tree `git -C` is refused), merges `--no-ff`, re-reads `HEAD` before the commit and again before the push, pushes `release/vX.Y.Z` (never force), and leaves the batch pull request with the leader. The full window procedure: `factory-dispatch-mechanics.md` § Integration into the release branch is self-served · the `acquire` settings-drift assertion: `factory-dispatch-gates.md` § The pre-merge settings-drift assertion. Once the card's work is confirmed landed on the remote integration branch, disposing the card's worktree is the sweep's step: `moai worktree sweep` disposes remote-landing-confirmed trees across both tiers, hoisting evidence before every removal (`worktree-integration.md` § Hoist a tree's evidence before disposing it).
<!-- moai:role-core-end -->

## Factory Mode — the card travels whole

`moai cc -f` launches the leader and lane sessions join with `moai cc -l`, labelled `lane-<n>`; the leader routes each card WHOLE to a free lane, which carries it `plan → run → sync` in-session and owns it end to end. Mechanics and the card-class effect on the entry stage: `factory-dispatch-mechanics.md` § Factory Mode mechanics · `factory-dispatch-detail.md` § Factory in-lane 3-stage.

**Lane spawn authority (standing).** A lane is an orchestrator for its card: it spawns, with the Agent tool and WITHOUT asking, the specialist the Status Transition Ownership Matrix names for the stage at hand — plan-phase artifacts to `manager-spec`, implementation to `manager-develop`, sync-phase docs to `manager-docs`, plus the chain's prescribed auditors. Depth-1 only: agents a lane spawns are leaf workers and never spawn further agents. This authority is part of the lane's bootstrap context (the SessionStart join notice carries it verbatim), and it is deliberately NOT a per-dispatch grant: the runtime's default "don't spawn unless the user asks" guidance does not bind a lane, and a leader's approval can neither grant nor revoke what the bootstrap already grants — the leader is not the lane's user.

## Boundaries — what this protocol does not do

- **No gate bypass.** Approval gates keep their evidence standard inside a dispatch cycle. The plan→run Kickoff's default form is the autonomous transition — independent audit cross + evidence criteria + a written decision record (`.claude/rules/moai/workflow/auto-semantics.md` §9.1) — which is the gate's new default form, not a bypass; keep-set gates (environment-impossible, operator-held, irreversible external-shared operations) still require the operator, and operator-form Kickoff rows that wait together are presented through the batch gate summary (`.claude/rules/moai/workflow/auto-semantics.md` §9.2).
- **No question delegation.** Lane sessions return blocker reports; the operator is asked by the leader, through `AskUserQuestion`.

> Moved to the detail companion: `factory-dispatch-mechanics.md` ("Boundaries — what this protocol does not do").

## Cross-references

> Moved to the detail companion: `factory-dispatch-mechanics.md` ("Cross-references").

---