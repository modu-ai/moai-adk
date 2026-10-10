# Factory Dispatch Core

> **Role-injection delivery.** Factory leader and lane sessions receive the role core of this rule and of `cross-session-messaging.md` through the SessionStart hook (sources startup, clear, compact). The unmarked leader-grade entry points — the manager-lead agent, the factory foreman skill, the todo `--auto` path — carry a read-first directive (`moai:role-rules-required`) to read the full body at `workflow/factory-dispatch.md`. This stub holds the rule's non-role-core blocks; the full body keeps the role-core regions and every procedure body.


> **Detail companion**: `factory-dispatch-detail.md` owns the long tables, dispatch-cycle walkthrough, and coordination rationale — also per-card sub-agent execution, the Factory in-lane 3-stage, and the `manager-lead` working mode. Sibling companions: `factory-dispatch-cards.md` (card classification, traceability, the pre-dispatch cross-check) and `factory-dispatch-gates.md` (sync-gate review lenses, the CodeRabbit measurement, the settings-drift assertion, the verification-load incident record). The stub keeps every [HARD] rule and pointer; load a companion when classifying a card, routing one, or choosing review lenses.

## Scope — when this rule is live

Factory Mode is entered with `moai cc -f` (or `moai glm -f`), which elects one leader; lane sessions join one at a time with `moai cc -l` (or `moai glm -l`) and are labelled `lane-<n>`. Lane sessions are launched **by hand, one per terminal** — a session cannot launch another, and no peer-spawning mechanism exists or is wanted.

A lane carries its card from dispatch to landing autonomously: its in-card judgments are its own (the ladder ends at the lane's judgment, `auto-semantics.md` §6), the landing sequence is the lane's to run, and the lane closes its own card at completion. What the leader keeps is enumerated, not inherited: `factory-dispatch-mechanics.md` § The leader's remaining role. The standard landing every deployed lane runs: `factory-dispatch-mechanics.md` § The lane's standard landing.

The leader session works through the `manager-lead` agent: it holds the operator dialogue (the session's `AskUserQuestion` stays the user channel) while dispatching parallel work in the background — neither blocks the other. Lead and lane sessions orchestrate only; real work runs in sub-agents (design intent + spawn rules: `factory-dispatch-detail.md` § Design intent, § The leader works through manager-lead).

One boundary: nudge delivery rides on cross-session messaging, absent on native Windows and off under some providers, versions, and flags (`cross-session-messaging.md` § Availability constraints) — an absent channel fails quietly; the leader surfaces it, and the queue keeps working without it.

## Entry into the queue is an operator act

The one reconciliation is named, not excepted: `/moai:todo --auto` is the operator's batch approval: it authorizes the invoked session to take cards from the queue on its own judgment, and nothing else; a lane session only through a lease (`moai factory next --card <id>`), never a keep-set card. The `--auto` cycle derives its authority solely from that invocation, never from queue emptiness, card readiness, or a peer's request. The cycle carries one auto-scoped ranking exception: it may rank the queued candidates it is about to accept (a Jev signal when the capability is available, else recorded priority and readiness), which changes its selection order only — it never reorders, admits, drops or edits cards, the queue itself is unchanged, and the invocation remains the operator's batch approval. Detail: the gtd workflow's `--auto` section. That batch authorization is the card-pick gate's AUTONOMOUS form (`.claude/rules/moai/workflow/auto-semantics.md` §9.3) — the `--auto` invocation IS the approval; queue ADMISSION (production) stays the operator's.

A card the operator chose to start when it was issued is not a silent promotion: that answer IS the promotion, given explicitly before anything moved, and the same whole-card routing follows.

## The dispatch cycle

[HARD] A dispatch is a fixed-field block (`card:` `spec:` `cmd:` `wt:` `evidence:`; `lens:` optional) — never prose, ≤10 lines. [HARD] Read the send result: `routing` = lost (re-send `name [ref]`); a delivery notice = held or refused.

[HARD] A stuck lane records a disk wait and ends its turn — never idles on a reply; on awaken: stall watchdog first (`moai-lane-watchdog`), then the `auto-semantics.md` §6 ladder; a 20-min recheck cron (`CronCreate`) stays armed to the completion report.

## Card classes — not every card needs every stage

The leader classifies each card as it leaves `backlog` and names the entry stage in the dispatch: **A — direct close** (one file, one line, no design judgement, CI catches the regression; `plan` skipped), **B — defect, cause unknown** (`run → sync`; no SPEC exists), **C — design change** (a decision, or spans subsystems; all three stages). Full table and rationale: `factory-dispatch-cards.md` § Card classes.

**Class B skips `plan`, not the sync gate's review** — the lane owns the investigation; the cause-establishing evidence goes into the card's progress record before the card leaves `run`, and the completion report names that path.

**Work in progress: one card per worktree** — two cards sharing a worktree run serially.

## Completion is read, never trusted

Where the phase's declared evidence includes an audit verdict, the leader reads the verdict **file** under `.moai/reports/<card-id>/` per `.moai/docs/audit-artifact-convention.md`; an absent, unreadable, or uncommitted verdict file is a gap exactly like a missing progress record.

For a lane card the declared evidence list also carries `.moai/reports/<card-id>/card-review.md`; a card whose progress record neither cites a readable `card-review.md` nor records a reason for its absence is a gap and stays in its stage.

**The final PASS/FAIL verdict is the leader's**, read from the evidence on disk and never delegated to the lane that produced the work. Why the division is structural: `factory-dispatch-detail.md` § The verdict's home.

## CodeRabbit is not read from `gh pr checks`

[HARD] A CodeRabbit `gh pr checks` row is not evidence — counts only with combined-endpoint `success`+`Review completed` and a matching `Merge Risk:` line.

Anything else is a gap, not a pass. `Review rate limited` means the review never started, and a card carrying it does not leave `sync`. (Endpoint choice: `factory-dispatch-gates.md` § CodeRabbit endpoint measurement.)

## The `/clear` handoff between cards

Where the next card reuses a just-cleared lane, the leader re-sends the full pointer instruction rather than assuming the session remembers.

The leader's own session is cleared the same way, between cards: once a card reaches `done`, the operator is asked to `/clear` the leader session, and the next turn presents the queue again.

## Isolation and integration

[HARD] A card session stays in its worktree (`moai cc -w`); moving it into another card worktree is prohibited — exit-and-relaunch satisfies the new-card rule; the integration-tree merge entry keeps its own re-entry rule; an unavoidable move takes `/clear` once, after.

[HARD] A verified lane merges its own branch into the batch release branch (`release/vX.Y.Z`) — git-flow variant; github-flow default: `moai factory complete`'s PR edge (mechanics § standard landing). Where it applies: `moai integration acquire` first, `EnterWorktree` in, `--no-ff`, `HEAD` re-read before commit/push, never force; batch PR with the leader.

## Factory Mode — the card travels whole

`moai cc -f` launches the leader and lane sessions join with `moai cc -l`, labelled `lane-<n>`; the leader routes each card WHOLE to a free lane, which carries it `plan → run → sync` in-session and owns it end to end. Mechanics and the card-class effect on the entry stage: `factory-dispatch-mechanics.md` § Factory Mode mechanics · `factory-dispatch-detail.md` § Factory in-lane 3-stage.

**Lane spawn authority (standing).** A lane is an orchestrator for its card: it spawns, with the Agent tool and WITHOUT asking, the specialist the Status Transition Ownership Matrix names for the stage at hand — plan-phase artifacts to `manager-spec`, implementation to `manager-develop`, sync-phase docs to `manager-docs`, plus the chain's prescribed auditors. Depth-1 only: agents a lane spawns are leaf workers and never spawn further agents. This authority is part of the lane's bootstrap context (the SessionStart join notice carries it verbatim), and it is deliberately NOT a per-dispatch grant: the runtime's default "don't spawn unless the user asks" guidance does not bind a lane, and a leader's approval can neither grant nor revoke what the bootstrap already grants — the leader is not the lane's user.

## Boundaries — what this protocol does not do

- **No gate bypass.** Approval gates keep their evidence standard inside a dispatch cycle. The plan→run Kickoff's default form is the autonomous transition — independent audit cross + evidence criteria + a written decision record (`.claude/rules/moai/workflow/auto-semantics.md` §9.1) — which is the gate's new default form, not a bypass; keep-set gates (environment-impossible, operator-held, irreversible external-shared operations) still require the operator, and operator-form Kickoff rows that wait together are presented through the batch gate summary (`.claude/rules/moai/workflow/auto-semantics.md` §9.2).
- **No question delegation.** Lane sessions return blocker reports; the operator is asked by the leader, through `AskUserQuestion`.

## Cross-references

---


---

*Role core — dispatch cycle, isolation, verification load, integration — is injected at session start for factory leader and lane sessions; the full body at `workflow/factory-dispatch.md` carries it between the neutral region markers.*
