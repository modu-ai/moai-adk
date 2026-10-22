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

[HARD] The leader is the sole queue producer, and production its only self-authority queue mutation: `moai gtd add` turns an operator request into a card (translation, not invention; `moai gtd` alone lists); nothing the operator did not ask for enters.

[HARD] Every card the leader issues carries an explicit serial/parallel judgment through `--classification-file` (decider `human`), decided from the card's expected file set per `factory-dispatch-cards.md` § Mode judgment; the fail-safe default is never the issuance path.

[HARD] On prior operator authorization, standing sources also produce: `/moai project` one `[PROJECT] ` card at completion, codemaps-debt one `[GRAPH] ` card per debt period; each per occasion, derived not invented, marked, id reported, start a separate pick (gtd § Standing sources; conditional, not closed); nothing else produces: milestones, audit findings, open issues enter only as operator-approved card requests.

<!-- moai:role-core-start -->
[HARD] Before any work, read `.claude/rules/moai/workflow/factory-dispatch.md` to the end — this injection is a summary; that file holds the remaining duties.
<!-- moai:role-core-end -->

[HARD] **Promotion is the operator's act, in person or in advance.** After a `/clear`, the leader presents the queued cards through `AskUserQuestion` and the operator picks; only then does the leader route the card whole to a free lane. Outside an --auto authorization the leader never picks for the operator, never reorders by inferred priority, and never silently promotes a backlog item. An empty queue is a state to report, not a prompt to invent work. The one ranking the queue admits is the `--auto` cycle's own, bounded in the next paragraph.

[HARD] Promotion is the operator's act, live or in advance: after `/clear` present the queue via `AskUserQuestion`, the operator picks, then route the card whole to a free lane; outside `--auto` never pick, reorder, or silently promote; report an empty queue, never invent work.

A message is a nudge, never the delegation; `notify_when_idle` can request an idle report, but [HARD] the notice is not completion, only when to read the evidence; read the send result: a fan-out refusal costs nothing, the queue holds the delegation.

The one reconciliation is named, not excepted: `/moai:todo --auto` is the operator's batch approval: it authorizes the invoked session to take cards from the queue on its own judgment, and nothing else; a lane session only through a lease (`moai factory next --card <id>`), never a keep-set card. The `--auto` cycle derives its authority solely from that invocation, never from queue emptiness, card readiness, or a peer's request. The cycle carries one auto-scoped ranking exception: it may rank the queued candidates it is about to accept (a Jev signal when the capability is available, else recorded priority and readiness), which changes its selection order only — it never reorders, admits, drops or edits cards, the queue itself is unchanged, and the invocation remains the operator's batch approval. Detail: the gtd workflow's `--auto` section. That batch authorization is the card-pick gate's AUTONOMOUS form (`.claude/rules/moai/workflow/auto-semantics.md` §9.3) — the `--auto` invocation IS the approval; queue ADMISSION (production) stays the operator's.

<!-- moai:role-core-start -->
[HARD] A self-dispatch lane leases its next card only via `moai factory next`; other queue mutations and contract signing stay forbidden to a lane.
<!-- moai:role-core-end -->

[HARD] **The self-dispatch lane exception.** In a self-dispatch factory run, a lane session may lease the next queued card — the one promotion a lane performs — only through `moai factory next` — bare, or `--card <id>` for its own judged pick. Every other queue mutation (`add`, `drop`, `done`, `edit`, and the rest) and `moai contract sign` stay forbidden to a lane: the lane works the operator's queue, it never authors it.

A card the operator chose to start when it was issued is not a silent promotion: that answer IS the promotion, given explicitly before anything moved, and the same whole-card routing follows.

<!-- moai:role-core-start -->
[HARD] The leader records relations (`moai gtd relate`) as evidence for the operator, never a mandate: it never folds away, reorders around, drops, or edits a related card; its one self-authority act is refusing a card whose normalized text is identical to one queued or picked.
<!-- moai:role-core-end -->

[HARD] **The leader may attach a finding; it may not act on one.** Analysis records a relation between two cards (`moai gtd relate`); the record is evidence the operator reads, never a mandate — the leader never folds the related card away, never reorders the queue around it, and never drops or edits it. Analysis changes exactly one thing on its own authority: it refuses the admission of a card whose normalized text is identical to one already queued or picked.

<!-- moai:role-core-start -->
[HARD] The pre-dispatch cross-check reads the card's PR and landed state and reports the same turn — an unchecked card is a gap.
<!-- moai:role-core-end -->

<!-- moai:role-core-start -->
[HARD] The cross-check also reports completed-SPEC coverage; neither read is conclusive — the final discriminator stays reproduction.
<!-- moai:role-core-end -->

<!-- moai:role-core-start -->
[HARD] The cross-check reports, never vetoes — the operator confirms or withdraws; never withhold a picked card.
<!-- moai:role-core-end -->

## Report milestones ↔ queue cards

[HARD] A milestone report under `.moai/reports/` carries `## Card Cross-Check`: a row per milestone, `card` = delivering id or explicit new-card marker, checked against `moai gtd`, never memory; card requests first state `N milestones → N cards`, naming every milestone without a live-queue card.

## Card classes — not every card needs every stage

The leader classifies each card as it leaves `backlog` and names the entry stage in the dispatch: **A — direct close** (one file, one line, no design judgement, CI catches the regression; `plan` skipped), **B — defect, cause unknown** (`run → sync`; no SPEC exists), **C — design change** (a decision, or spans subsystems; all three stages). Full table and rationale: `factory-dispatch-cards.md` § Card classes.

[HARD] Class A requires both cited: `git diff --stat` vs base showing one file, and CI green on the merging head (else not Class A); the dispatch states no-design-judgement for the operator to dispute; never "faster" as the reason.

**Class B skips `plan`, not the sync gate's review** — the lane owns the investigation; the cause-establishing evidence goes into the card's progress record before the card leaves `run`, and the completion report names that path.

**Work in progress: one card per worktree** — two cards sharing a worktree run serially.

## The dispatch cycle

- `card`, `cmd`, `wt`, and `evidence` are always present. `spec` joins once a SPEC exists; a Class B card carries none, and its `evidence` names whatever record the leader will read instead. `lens` appears only when the leader names review lenses for the card's sync gate.

### The delegation channel is the queue

<!-- moai:role-core-start -->
[HARD] Work is delegated through the queue on disk, not messages — one queue, visible from every worktree.
<!-- moai:role-core-end -->

[HARD] A dispatch is a ≤10-line fixed-field block, no prose (overflow goes into the card): always `card:`, `cmd:` (`/moai run <SPEC-ID>`), `wt:` (the new card tree, never a prior one), `evidence:` (`progress.md`; Class B: the record the leader reads); `spec:` once a SPEC exists (not Class B); `lens:` only if the leader names sync-gate lenses. [HARD] Read the send result: `routing` = lost, re-send to `name [ref]`; `[Cross-session delivery notice]` = held or refused, tell the operator, never re-send; else queued; none proves the lane read it.

### Dispatch language

<!-- moai:role-core-start -->
[HARD] Dispatches go in the operator's `conversation_language`; `Agent()` prompts stay English; identifiers verbatim.
<!-- moai:role-core-end -->

### The lane's task list carries the card's stages

A lane MAY track its stages in the task list; when it does, it deletes the previous card's tasks at card transition. The card's state of record is the factory record and progress.md — the task list is never completion evidence.

> Moved to the detail companion: `factory-dispatch-mechanics.md` ("The lane's task list carries the card's stages").

With `tree_scope: skip` configured for the leader's checkout, the leader session carries no turn-end codex review gate and reviews its own internal output directly with the same tools. Stage order, evidence fields, and the re-review ceiling: `factory-dispatch-detail.md` § The card-review stage.

<!-- moai:role-core-start -->
[HARD] Between run-exit verification and integration every lane runs `card-review`: `codex_review` (`glm_review` optional) into `.moai/reports/<card-id>/card-review.md`, advisory, never replacing the leader evidence read or an independent audit; under `tree_scope: skip` the leader self-reviews with the same tools.
<!-- moai:role-core-end -->

### Lane waits are explicit, and stalls are watched

<!-- moai:role-core-start -->
[HARD] A stuck lane writes a disk wait (reason, whom, recheck point), ends its turn, never idles on a reply; on wake: `moai-lane-watchdog` first, then the `auto-semantics.md` §6 ladder ending in its own judgment + decision record, never ask-and-wait; the leader only at a gate boundary (keep-set, cross-card conflict); a recurring `CronCreate` recheck armed before the first stage, kept to the completion report; each wake reads disk evidence before replying.
<!-- moai:role-core-end -->

[HARD] A lane that cannot proceed records an explicit wait on disk — reason, whom, recheck point — and ends its turn; it never idles open-ended on a reply. On its next awaken the lane runs the stall watchdog first — invoke Skill("moai-lane-watchdog") and follow it — which measures progress, classifies the stall cause, and resolves the judgment through the decision ladder (doctrine: `.claude/rules/moai/workflow/auto-semantics.md` §6). The ladder's terminal step is the lane's own judgment + decision record — the lane never asks the leader and waits on a judgment; reaching the leader is a gate boundary (a keep-set category or a cross-card conflict), not a ladder step. The lead's decision board carries leader-level rulings only — cross-card conflicts, serial-slot policy calls, keep-set gate relays — never a lane's in-card judgment. A lane also arms a standing recheck cron (`CronCreate`, recurring, every renewal period of `workflow.factory_lease.renew_interval`) before its first stage and keeps it until its completion report: a lane stopped by an API error (429) or by a delegate's report that never comes wakes without outside help only if something armed before it stopped fires, and every wake reads the disk evidence (progress record, reports, commits, the delegate's deliverable) before it replies (`.claude/rules/moai/workflow/auto-semantics.md` §5.1).

## Deputy dispatch surface

[HARD] Before the first lane dispatch the leader spawns exactly one resident UNNAMED background `Agent()` running manager-lead as deputy, kept for the batch.

<!-- moai:role-core-start -->
[HARD] Completion reports arrive as `RECOMMEND:` summaries — occupancy moves; the leader's evidence read does not.
<!-- moai:role-core-end -->

[HARD] The deputy holds no power of consequence: final PASS/FAIL verdicts, final merge approval (`LEAD-MERGE-APPROVED`), operator gates, card issuance and `done`, CodeRabbit slot-wait adjudication, cross-session dispute coordination stay with the leader; the deputy measures and drafts round reports, the leader asserts the figures, each naming its measurer.

## Completion is read, never trusted

<!-- moai:role-core-start -->
[HARD] The leader advances a card only on evidence it read, never a lane reply: `progress.md` + the declared verification path; missing, unreadable, or stale evidence is a gap: the card stays, the leader says why; no failure signal is not a pass.
<!-- moai:role-core-end -->

Where the phase's declared evidence includes an audit verdict, the leader reads the verdict **file** under `.moai/reports/<card-id>/` per `.moai/docs/audit-artifact-convention.md`; an absent, unreadable, or uncommitted verdict file is a gap exactly like a missing progress record.

For a lane card the declared evidence list also carries `.moai/reports/<card-id>/card-review.md`; a card whose progress record neither cites a readable `card-review.md` nor records a reason for its absence is a gap and stays in its stage.

**The final PASS/FAIL verdict is the leader's**, read from the evidence on disk and never delegated to the lane that produced the work. Why the division is structural: `factory-dispatch-detail.md` § The verdict's home.

### CodeRabbit is not read from `gh pr checks`

[HARD] A CodeRabbit `gh pr checks` row is not evidence (success even with no review); count it only when combined `/commits/{sha}/status` shows `success` + `Review completed` AND a `Merge Risk:` line commit prefix matches the current `headRefOid`.

Anything else is a gap, not a pass. `Review rate limited` means the review never started, and a card carrying it does not leave `sync`. (Endpoint choice: `factory-dispatch-gates.md` § CodeRabbit endpoint measurement.)

## Lane card transition — in-session, no clear

[HARD] No card context crosses to the next: a lane treats the conversation context of earlier cards as non-authoritative and works each newly leased card only from that card's body, its SPEC artifacts, and its progress record. At `done`, evidence read, the leader does not ask the operator to clear the lane session; the leader names what closed (card, evidence read) and the lane moves on by the transition below.

<!-- moai:role-core-start -->
[HARD] `ExitWorktree(keep)`, lease, `EnterWorktree(path)`; conversation context never carries.
<!-- moai:role-core-end -->

Transition detail (Claude Code lanes): the lane requests its next card from the primary checkout with `moai factory next` (a contended claim is waited out with the verb's wait option), enters the leased worktree by absolute path, and verifies `git rev-parse --show-toplevel` equals that path before any card work. A refused entry is recorded verbatim in the card's progress record and reported as a blocker — never a session clear, a second lease, or card work in the primary checkout; the only fallback is an operator relaunch inside the leased worktree. Under manual dispatch (`--no-auto-dispatch`, `MOAI_FACTORY_AUTO_DISPATCH=manual`) the lane leaves with `ExitWorktree(keep)` and waits for the leader's dispatch instead of leasing, and under the relaunch clear policy (the completion output tells it to end the session) it ends the session. Codex lanes keep their own session-end transition.

When the next card reuses a lane, the leader re-sends the full pointer instruction instead of assuming the session remembers.

The leader session itself is cleared between cards: at `done` the operator is asked to `/clear` it, and the next turn presents the queue again.

## Isolation is provisioned by MoAI, then entered through a launcher

[HARD] **A card session starts inside its worktree and stays there — between cards it moves only by the in-session transition (`ExitWorktree(keep)`, then `EnterWorktree` of the leased tree)**; a relaunch inside the leased worktree is the operator's fallback when the entry is refused (the full clause lives on the always-loaded stub; measured detail: `factory-dispatch-mechanics.md` § Isolation).
<!-- moai:role-core-start -->
[HARD] Card work happens in a launcher-entered worktree — never a bare `git worktree add` (mechanics § Isolation).
<!-- moai:role-core-end -->

[HARD] A card session starts in its worktree (`moai cc -w`) and stays while it carries the card; between cards it moves by the in-session transition (`ExitWorktree(keep)`, lease, `EnterWorktree` of the leased tree); an integration-tree merge entry is exempt but returns to the card worktree before card work.

<!-- moai:role-core-start -->
[HARD] A Codex factory agent never calls `moai cc -w`, `EnterWorktree`, or `ExitWorktree`.
<!-- moai:role-core-end -->

<!-- moai:role-core-start -->
[HARD] `moai worktree done` closes L2 trees only (worktree-integration § Terminology Glossary).
<!-- moai:role-core-end -->

[HARD] An unpushed card branch makes its worktree the only copy: dispose of none (L1 or L2) until the leader integrated it and the remote merge landed.


[HARD] A new card gets a new worktree from the configured base, never reused; `ExitWorktree` the old tree first; merge an unmerged prior-card dependency inside the new tree.

<!-- moai:role-core-start -->
[HARD] Branches carry `WT-` + a descriptive slug, never the card id; traceability = dispatch `card:` + commit message + evidence path + PR title.
<!-- moai:role-core-end -->

<!-- moai:role-core-start -->
[HARD] In a worktree session `<path>` is that worktree's absolute path — relative, computed, and outside forms are refused.
<!-- moai:role-core-end -->

## Verification load is lane-local

<!-- moai:role-core-start -->
[HARD] Verification is lane-local — only what the change can affect; CI runs the full suite (gates § incident record).
<!-- moai:role-core-end -->

<!-- moai:role-core-start -->
[HARD] Never spawn background load — contention needs cleanup-guaranteed kills or a `timeout` wrapper.
<!-- moai:role-core-end -->

### The env-isolated verification form

<!-- moai:role-core-start -->
[HARD] Env-scrubbed verification is one compound `unset … && <command>` call — a separate `unset` carries nothing.
<!-- moai:role-core-end -->

> Moved to the detail companion: `factory-dispatch-mechanics.md` ("The env-isolated verification form").

## Integration into the release branch is self-served

> Moved to the detail companion: `factory-dispatch-mechanics.md` ("Integration into the release branch is self-served").

<!-- moai:role-core-start -->
[HARD] git-flow only: a verified lane merges its branch into `release/vX.Y.Z` itself: `moai integration acquire --name <lane> --card <card-id>` before entry, release after the completion report; `EnterWorktree` in (cross-tree `git -C` refused; Codex: `git -C <absolute release worktree path>`); `--no-ff`; re-read `HEAD` before the commit and again before the push; push, never force; the batch PR stays with the leader.
<!-- moai:role-core-end -->

[HARD] A lane whose card has passed verification does not wait for the leader to integrate it: the lane merges its own branch into the batch's release branch (`release/vX.Y.Z`) itself. This is the **git-flow variant** — it applies only where the project's git strategy names git-flow; under github-flow (the distributed default) the delivery is `moai factory complete`'s pull-request edge and the full sequence is `factory-dispatch-mechanics.md` § The lane's standard landing. Where it applies, the window is taken with `moai integration acquire --name <lane> --card <card-id>` BEFORE entering the release worktree, released after the completion report is sent; the lane enters the release worktree with `EnterWorktree` (a cross-tree `git -C` is refused), merges `--no-ff`, re-reads `HEAD` before the commit and again before the push, pushes `release/vX.Y.Z` (never force), and leaves the batch pull request with the leader. The full window procedure: `factory-dispatch-mechanics.md` § Integration into the release branch is self-served · the `acquire` settings-drift assertion: `factory-dispatch-gates.md` § The pre-merge settings-drift assertion. Once the card's work is confirmed landed on the remote integration branch, disposing the card's worktree is the sweep's step: `moai worktree sweep` disposes remote-landing-confirmed trees across both tiers, hoisting evidence before every removal (`worktree-integration.md` § Hoist a tree's evidence before disposing it).


## Factory Mode — the card travels whole

`moai cc -f` launches the leader and lane sessions join with `moai cc -l`, labelled `lane-<n>`; the leader routes each card WHOLE to a free lane, which carries it `plan → run → sync` in-session and owns it end to end. Mechanics and the card-class effect on the entry stage: `factory-dispatch-mechanics.md` § Factory Mode mechanics · `factory-dispatch-detail.md` § Factory in-lane 3-stage.

**Lane spawn authority (standing).** A lane is an orchestrator for its card: it spawns, with the Agent tool and WITHOUT asking, the specialist the Status Transition Ownership Matrix names for the stage at hand — plan-phase artifacts to `manager-spec`, implementation to `manager-develop`, sync-phase docs to `manager-docs`, plus the chain's prescribed auditors. Depth-1 only: agents a lane spawns are leaf workers and never spawn further agents. This authority is part of the lane's bootstrap context (the SessionStart join notice carries it — the canonical English sentence for the `en` locale and localized per `conversation_language` for the others, with the Status Transition Ownership Matrix pointer path (.claude/rules/moai/development/spec-frontmatter-schema.md) verbatim in every locale), and it is deliberately NOT a per-dispatch grant: the runtime's default "don't spawn unless the user asks" guidance does not bind a lane, and a leader's approval can neither grant nor revoke what the bootstrap already grants — the leader is not the lane's user.

<!-- moai:role-core-start -->
[HARD] A lane spawning its implementer (`manager-develop`) passes `effort` on the Agent call by card class: C `high`, B `medium`, A `low` — the per-class value overrides the agent's frontmatter default for that spawn only.
<!-- moai:role-core-end -->

## Boundaries — what this protocol does not do

- **No gate bypass.** Approval gates keep their evidence standard inside a dispatch cycle. The plan→run Kickoff's default form is the autonomous transition — independent audit cross + evidence criteria + a written decision record (`.claude/rules/moai/workflow/auto-semantics.md` §9.1) — which is the gate's new default form, not a bypass; keep-set gates (environment-impossible, operator-held, irreversible external-shared operations) still require the operator, and operator-form Kickoff rows that wait together are presented through the batch gate summary (`.claude/rules/moai/workflow/auto-semantics.md` §9.2).
- **No question delegation.** Lane sessions return blocker reports; the operator is asked by the leader, through `AskUserQuestion`.

> Moved to the detail companion: `factory-dispatch-mechanics.md` ("Boundaries — what this protocol does not do").

## Cross-references

> Moved to the detail companion: `factory-dispatch-mechanics.md` ("Cross-references").

---