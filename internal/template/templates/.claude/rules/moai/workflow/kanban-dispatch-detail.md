---
description: "Detail companion for kanban-dispatch.md — terminology, board table, card classes, dispatch-cycle naming, sync-gate review-lens table, /clear message structure, isolation rationale, verification-load incident record, sub-agent-first design intent, manager-lead working mode, per-card fan-out, Factory in-lane 3-stage, pre-dispatch cross-check rationale, PR-title carrier measurements"
paths: "**/kanban-dispatch*.md,**/.claude/agents/moai/manager-lead.md,**/.claude/skills/moai/workflows/todo.md"
---

# Kanban Dispatch — Detail Companion

> Detail companion of `kanban-dispatch.md` (the always-loaded stub). The stub keeps every [HARD] rule, prohibition, and cross-reference; this file owns the long tables, the dispatch-cycle walkthrough, incident narratives, and worked rationale. Load when moving a card between columns, classifying a card, or choosing review lenses for a sync dispatch.

## Design intent — sub-agent-first token discipline

The lead and lane sessions keep **only orchestration** in their context windows. Every unit of real work — research, authoring, implementation, verification sweeps — is delegated to `Agent()` sub-agents; the verbose output (tool results, file dumps, test logs) stays in the sub-agent's window, and only summaries return to the session. This is the token rationale for the whole architecture: a session that survives an entire card — or, in Factory Mode, an entire batch — would otherwise accumulate every card's raw output on top of the always-loaded prefix. The sections below (the manager-lead working mode, per-card fan-out, Factory in-lane 3-stage) are this intent expressed structurally.

## Terminology — the board vocabulary

`kanban-dispatch.md`, `sprint-round-naming.md`, and the operating notes share a working vocabulary that previously had no definition anywhere. Each term gets one definition and one example; the sections below assume these meanings.

| Term | Definition | Example |
|---|---|---|
| **lane** | One parallel work stream that carries a card end to end: one session paired with one worktree. A lane is a swimlane — a band reserved for one stream of work so parallel streams never interleave, and never share a working tree. "Lane-local verification" = that lane runs only the tests its own change can affect. | The `run` session working in worktree `.claude/worktrees/t0` is one lane. |
| **card** | One unit of work on the board, entered by the operator via `/moai todo "<description>"` and referred to by a short id. A card owns one worktree, one progress record, and its completion evidence. | `t0` — a one-line fix card. |
| **column** | One stage of the board, in fixed order `backlog → plan → run → sync → done`. The three working columns each map to exactly one companion role; the review verdict lives inside the sync gate. | `/moai run <SPEC-ID>` happens in the `run` column. |
| **backlog** | The entry queue of the board. No session owns it by design — work enters only when the operator puts it there. | `/moai todo "rename hint is stale"` appends a card to the backlog. |
| **lead** | The single coordinating session (`moai cc -k`). Moves cards between columns on evidence it read itself, asks the operator to `/clear` companions between phases, never writes code. | The session that dispatched a card with its worktree instruction. |
| **companion** | A worker session launched by hand, one terminal at a time (`moai cc -k --name <role>`), owning one column's work at a time. Named by its bare role; a second live session claiming the same role takes the next free number. | `plan`, `run`, `sync`. |
| **run-id** | The short identifier the lead prints at launch. It lives in `MOAI_KANBAN_ID` and the lead socket path — no session name carries it, the lead's included (t133): every session is named by its role, and a second live claim on a role takes the next free number. | `a1b2c3` — printed in the lead's bootstrap notice; the session itself is named `lead`. |
| **worktree** | The isolated checkout where a card's work happens, entered through the launcher (`moai cc -w <name>` / `EnterWorktree`), never raw `git worktree add`. The directory carries the card id; the branch carries a descriptive slug, `WT-<slug>` (≤3 hyphen tokens, ≤24 chars, no card id — `kanban-dispatch.md` § Isolation is entered, never provisioned). A worktree outlives a phase: one spans run through sync. | `.claude/worktrees/t0` on branch `WT-todo-queue`. |
| **dispatch** | The lead's instruction to one companion: a pointer (card id, SPEC id, phase command, completion signal), never a copy of the work. Written in the operator's conversation_language. | "card: t0 — wt: EnterWorktree(t0) … evidence: .moai/reports/t0/". |

The pair most easily confused: a **column** names a phase of the work (`run`); a **lane** names who carries one card through those phases (the `run` session in `.claude/worktrees/t0`). One is a stage on the board, the other is a stream through the stages.

Factory Mode companions are named `lane-1..lane-N` — `lane` is the user-facing term. A factory lane owns a card end to end rather than one column (§ Factory in-lane 3-stage); in Kanban Mode a lane is one column's session carrying its column's cards.

## The board

Five columns, fixed and ordered:

```
backlog → plan → run → sync → done
```

`backlog` and `done` have no owning session. The three working columns between them each map to exactly one companion role, which is what makes dispatch a lookup rather than a decision. There is no `review` column: the review verdict is absorbed by the sync gate, which runs the review lenses itself (§ Review lens selection).

| Column | Owning role | What happens there |
|---|---|---|
| `backlog` | *none* — a queue | Work waits. Entry is an operator act (see the stub). |
| `plan` | `plan` | SPEC authored (`/moai plan`), then plan-audited. |
| `run` | `run` | Implementation (`/moai run <SPEC-ID>`). |
| `sync` | `sync` | Review verdict (lenses per card), docs, CHANGELOG, PR (`/moai sync <SPEC-ID>`). |
| `done` | *none* — terminal | Card closed. Nothing is dispatched here. |

## Report milestones ↔ queue cards

The stub's [HARD] rule — a `## Card Cross-Check` section per milestone-bearing report, mapping claims verified against the queue — exists because report→card linkage used to live in one person's memory: a report declared milestones, cards were issued separately, and nothing reconciled the two. Milestones surfaced with no card, and cards claimed milestones that had already landed.

The mechanical check runs the same comparison the lead states by hand:

```
moai graph build && moai graph query --milestones-no-card
```

It writes the report-milestone and milestone-card edges from each report's Card Cross-Check table and lists every milestone whose claimed card is missing from the live queue (queued/picked; dropped does not qualify). "Not in live queue" covers both completed and never-issued cards — resolve each flag with `git log --oneline --grep 'merge: <card-id>'` before issuing a new card.

## Card classes — not every card needs every column

Most of what accumulates in the backlog is chores: a one-line fix, a stale reference, a renamed flag. Sending those through `plan → run → sync` costs more in ceremony than the change is worth. The lead classifies each card as it leaves `backlog` and names the class in the dispatch.

| Class | Shape | Path |
|---|---|---|
| A — direct close | The change is one file and one line, there is no design judgement in it, and CI catches the regression | One session carries the card through to a pull request; `plan` is skipped |
| B — defect, cause unknown | Something is wrong and the cause has not been established | `run → sync`; `plan` is skipped, so no SPEC exists |
| C — design change | The change contains a decision, or spans subsystems | All three working columns |

The Class-A evidence rule is the same shape as the CodeRabbit section of the stub: a class that skips review on a claim nobody checked is exactly the unobserved-claim hazard this rule forbids everywhere else; writing the justification down is not the same as verifying it.

For Class A this inverts where the parallelism comes from. Handing three sessions a whole card each puts three cards in flight; pipelining one card through three columns puts one. Pipelining repays its handoff cost only when each column does substantial work, which is the Class C case — and research fan-out during `plan` is reserved for Class C for the same reason. Per-card fan-out (§ Per-card fan-out and sub-agent execution) is the across-cards axis layered on top of this; Factory Mode replaces the between-session pipelining entirely (§ Factory in-lane 3-stage).

## The dispatch cycle

Each arrow below is one dispatch from the lead to one companion session:

```
[operator picks a card]  →  plan  →  run  →  sync  →  [lead marks done]
```

Dispatch is addressed by session name. Companions are named by their bare role; a name held by a live session is bumped to the next free number, and no run id travels in any session name, the lead's included (one run per machine; the id lives in `MOAI_KANBAN_ID` and the lead socket path). `ListAgents` lists live sessions and says when it could not check them all; send with `SendMessage({to: "<name>", message: "…"})`, using the short reference the listing prints when a bare name is ambiguous.

A bare name fails in two different ways, and only one of them announces itself. The refusal case is the harmless one: the runtime cannot pick between same-named sessions, says so, and the send is re-issued with the short reference the listing prints. The other case says nothing at all.

**The silent case — an in-process mailbox takes the name.** While the team namespace is active (`CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS=1`, which ships enabled in settings and in the distributed template), a name carried by both a live companion session and an in-process teammate mailbox resolves to the **mailbox**. Nobody reads it, the companion never learns a dispatch existed, and the send reports success. The result's shape is the only separator:

| Result shape | Where the dispatch went |
|---|---|
| no `routing` object — the result names the peer (`… (another Claude session on this machine)`) | the companion session — **delivered** |
| a `routing` object, e.g. `routing: {sender: "team-lead", target: "@<name>"}` | an in-process team mailbox — **lost** |

So read a `routing` object as a failure signal and re-send to the `name [ref]` the listing printed. The two rows are not equally measured, and the table should not be read as if they were: the absent-`routing` row is directly observed — dispatches arriving normally, repeatedly — while the present-`routing` row rests on a single reported contrast experiment that sent both forms in one turn. That is enough to act on, because the cost of re-sending a delivered message is one duplicate and the cost of missing a lost one is a stalled board, but it is one measurement rather than a pattern.

Detecting the collision is the second line of defence. The first is not creating it: the lead spawns its coordination agent unnamed precisely so no in-process teammate ever carries a name a companion session also answers to (§ The lead works through manager-lead). Where that holds, this section never fires.

**The collision is conditional, not universal.** A name no in-process teammate shares still delivers on the bare form, and bare-name dispatch is observed arriving normally in runs with no such teammate — so "always address by reference" would be a false rule, and the binding one is *read the result*. Companion names are role-shaped, and a spawned teammate can carry the same shape; where that overlap is plausible, address by `name [ref]` from the first send rather than after a lost one.

This does not soften what moves the board. The queue on disk is still the delegation, evidence on disk is still what advances a card, and a dispatch that silently missed shows up as a card that never progressed — the message was only ever a nudge. Mechanism from the messaging side: `cross-session-messaging.md` § Addressing, sending, and replying.

**A listing is not always complete, and an incomplete one cannot prove absence.** From Claude Code 2.1.234 the listing says when your account's session list was too long to check completely, rather than leaving unseen sessions to read as absent. That disclosure is what the stub's fault clause turns on: the lead concludes a role is empty from a listing that checked everywhere, and a listing that reports it could not is a **gap** — it re-checks and reports the gap, never a fault. Concluding otherwise reports a running companion as dead, which is the failure the upstream change exists to prevent, and is an unobserved absence claim under `verification-claim-integrity.md` §1 (the absence of a signal is not evidence of its subject's absence). This binds only the *absence* direction: a session the listing DOES show is present whatever else it could not reach.


Each instruction carries, at minimum: the card, the SPEC ID once one exists, the phase command to run, and the completion signal to write. Keep it a pointer, not a copy — the companion reads the SPEC artifacts itself rather than receiving them inline.

**`sync → done` is the same act with the dispatch removed.** No session occupies `done`, so the lead reads the sync session's completion evidence and records the terminal transition itself.

### Dispatch language — classification rationale

[HARD] A dispatch is written in the operator's `conversation_language` (normative statement lives in the stub). This is a classification, not an exemption from the language rules, and it needs no change to either of them:

- `agent-common-protocol.md` § Language Handling already opens with the opposite of an English default — agents receive and respond in the configured `conversation_language`. What it fixes to English is code, identifiers, and names. A dispatch is prose, so the rule was never against it; it simply did not name cross-session messages in its list.
- `moai-constitution.md` § Response Language reserves English for internal agent communication, but the axis that clause sits on is stated one line above it: user-facing responses go in the operator's language. A message a human reads is user-facing by that rule's own criterion, so putting a dispatch in the operator's language applies the constitution rather than carving an exception out of it.

### Dispatch format — rationale

The address-block format (normative statement lives in the stub) is the "pointer, not a copy" rule above made mechanical: every field is an address the companion resolves by reading what it names. It also settles the Dispatch language rule by construction — a block of pure addresses has nothing to translate, while the lead's reports to the operator (progress notes, `/clear` requests) remain in the operator's `conversation_language`.

## The lead works through manager-lead

The `-k` / `-f` lead session does not draft its own coordination: it spawns the `manager-lead` agent (renamed from `manager-kanban`) and works through it. The split:

- **Dialogue.** manager-lead holds the operator conversation — card selection, `/clear` prompts, blocker surfacing. The blocker-report discipline is unchanged (`agent-common-protocol.md` § Blocker Report Format): manager-lead returns blocker reports, and the session's `AskUserQuestion` remains the user channel. The rename moves no part of the user channel into the agent.
- **Dispatch.** manager-lead routes cards and reads evidence in the background while the dialogue continues — neither the conversation nor the lane coordination blocks the other, because sub-agents run in the background by default and the session stays free between their returns.

[HARD] **Spawn manager-lead UNNAMED.** Under `CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS` a named spawn converts to an in-process teammate, and that path carries two measured hazards: an observed output-loss discrepancy (one same-version session watched a named spawn become an output-less in-process teammate — the two-sided evidence is recorded in CLAUDE.md §15), and in-process teammates cannot spawn background subagents, which forces foreground-only execution and breaks the dialogue-never-blocks property above. An unnamed spawn keeps it a plain subagent (related: `orchestration-mode-selection.md` §C.1).

### Deputy mode — background coordination off the lead's turn

The lead's turn loop is the scarcest surface on the board: a dispatch send, a CI watch, and a CodeRabbit poll each occupy it serially, and while it is occupied the lead cannot judge anything else. The deputy exists to move that occupancy. The lead session (still through manager-lead, still UNNAMED) delegates coordination duties to a background deputy instance whose charter — the delegable/retained matrix, the delivery-shape verification protocol, the standing messaging hazards — is codified in the agent itself (`.claude/agents/moai/manager-lead.md` § Deputy dispatch surface). This section adds only what the board sees of it.

**What the deputy does in the background:**

- **Dispatch sends with delivery-shape verification.** The deputy sends the fixed-field address blocks for ALREADY-PICKED cards and reads every send result. A `routing` object on the result means an in-process mailbox took the block — lost, not delivered — and the deputy re-sends to the `name [ref]` form the `ListAgents` listing printed (§ The dispatch cycle). A rapid-burst refusal is read and reported; the queue already carries the delegation, so a refused nudge stalls nothing.
- **Bounded CI-watch polls.** The deputy polls a card's checks to terminal states and returns those states — the states, never a judgement about them.
- **CodeRabbit two-condition reads.** The deputy reads the combined-status `CodeRabbit` entry (state `success` AND description `Review completed`) plus the `Merge Risk:` line matching the current `headRefOid`, and REPORTS both conditions to the lead. It never adjudicates the slot-wait outcome: a card carrying `Review rate limited` is reported as exactly that, and the lead decides what it means.
- **First-pass evidence reading.** The deputy reads a card's completion evidence and returns findings as recommendations, each prefixed `RECOMMEND:` — never a verdict token.
- **Summary reporting.** The deputy folds its observations into a single report addressed to the lead session.

**What returns to the lead's turn:** `RECOMMEND:`-prefixed recommendations, terminal CI states, delivery confirmations and refusals, and the two-condition CodeRabbit read. That is the entire return surface — the deputy's report is input to the lead's judgement, never a substitute for it.

**What does not change — the structural principles:**

- **The verdict's home.** The final PASS/FAIL remains the lead's, read from evidence on disk (§ The verdict's home). A deputy recommendation promoted to a verdict is exactly the adjudication-promotion failure the report/verdict split exists to prevent.
- **Completion is read, never trusted.** The deputy's first-pass read does not discharge the lead's own evidence-read obligation. The lead advances a card on what the lead read; a deputy report is one more claim until the evidence under it has been read.
- **The queue is the channel.** The deputy's `SendMessage` is a nudge, never the delegation (stub § The delegation channel is the queue); card advancement never depends on a message arriving, whoever sent it.

The retained powers — final verdicts, final merge approval, operator gates, card issuance and `done`, CodeRabbit adjudication, cross-session dispute coordination — are enumerated under the `DEPUTY-RETAINED-BY-LEAD` marker in the agent charter; the stub carries the [HARD] boundary clauses.

The queue on disk remains the delegation channel and completion remains evidence-read (stub § The delegation channel is the queue, § Completion is read, never trusted) — messaging stays a nudge. The rename changes who drafts the coordination, not what counts as delegated or as done.

## Per-card fan-out and sub-agent execution

Two distinct fan-out axes exist; do not conflate them:

- **Across cards (per-card fan-out).** A lane holds several cards in one column by giving each card's work to a parallel `Agent()` worker. The plan lane authors SPECs for several cards at once; each worker writes only inside its own card directory (`.moai/specs/<SPEC-ID>/`), which is why parallel writes cannot collide. Run and sync lanes use the same pattern for parallelizable per-card work.
- **Within a card (stages as sub-agents).** The lane session orchestrates only: each stage's execution — plan authoring, run implementation, sync sweeps — is spawned as `Agent()` sub-agents whose output stays in their windows. The lane merges and inspects results through the evidence they write, not by absorbing their transcripts.

Discipline:

- **Ceiling: 10 concurrent agents per lane.** The launcher injects `CLAUDE_CODE_MAX_CONCURRENT_SUBAGENTS` at that value; a lane does not raise it.
- **Parallel fan-out is for read-heavy work plus per-card-isolated writes.** Never run multiple write-capable agents on the SAME card concurrently — one card, one writer at a time; cards in parallel, stages of one card in series. (Intra-card research fan-out during `plan` remains Class-C-only — § Card classes.)
- **Write-capable sub-agents spawned in parallel MUST carry `isolation: "worktree"`** (the Agent tool's isolation parameter), so concurrent file writes cannot collide even outside the card directories: each write agent works in its own worktree copy, and the lane integrates via evidence and merge. Read-only fan-out (investigation, audits) stays unisolated — worktree setup cost buys nothing there. L1 `Agent(isolation: "worktree")` semantics, the relative-path prompt rule, and the lifecycle are owned by `worktree-integration.md` and are cross-referenced, not restated here; the stub's "dispose only after the remote merge lands" and "exit the previous worktree before a new card's" rules remain the SSOT for card worktrees.
- **Verification stays lane-local.** The full suite is CI's job (stub § Verification load is lane-local).
- **Stagger same-type spawns**: spawn one worker first and the rest once it has started producing, so the later spawns read the first one's prompt cache (`cache-aware-execution.md` directive 2).
- **Sub-agents are not sessions.** The stub's "launched by hand, no peer-spawning" boundary governs sessions; an `Agent()` worker inside a lane is ordinary in-session orchestration, and the lane remains the card's owner of record.

## Factory in-lane 3-stage

Factory Mode (`moai cc -f <N>` / `moai glm -f <N>`) trades the per-column board for whole-card ownership: one lead plus `lane-1..lane-N` sessions, each lane owning one card end to end. Lanes are launched by hand like kanban companions; the lead keeps the run-id, the queue, and the verdict.

- **Routing.** The lead routes a card WHOLE to a free lane — free means the lane's previous card reached `done` and its evidence was read. A lane busy on a card is not addressed; with every lane busy, the card waits in the queue rather than being dispatched. The address block is unchanged; `cmd` names the entry stage the class prescribes (`/moai plan` for C, `/moai run` for B, the direct close for A), and the lane proceeds through the remaining stages without further dispatches.
- **Serial stages, sub-agent execution.** Plan completes before run begins, run before sync — a lane never runs two stages of the same card concurrently. Within a stage it fans out sub-agents per § Per-card fan-out and sub-agent execution.
- **Classes collapse into the lane.** A/B/C still name which ceremonies a card skips (B skips `plan` and carries no SPEC; A goes straight to the close), but the "wholesale vs serial hand-off" distinction between sessions disappears — there are no per-column sessions to hand off to. Every lane runs the serial stages for whatever its card still needs.
- **The `/clear` boundary is between cards, not phases.** The between-phases hand-off disappears (that is the point of the mode); the between-cards one does not — a factory lane is cleared once its card reaches `done`, before the next card is routed to it, exactly as the stub's `/clear` rule requires.
- **Evidence, verdict, integration unchanged.** The lane writes the same completion signals; the lead still reads evidence and owns the final PASS/FAIL; release-branch integration is still lane work under the stub's rules. The deputy surface applies unchanged too: a factory lead may delegate dispatch sends and watches to the coordination deputy exactly as a kanban lead does (§ The lead works through manager-lead → Deputy mode), while the lead keeps the verdict and the batch pull request.

## The verdict's home

The stub keeps the norm — the final PASS/FAIL verdict is the lead's, read from evidence on disk, never delegated to the lane that produced the work. The division is structural, not ceremonial: where the board's lanes run on a different backend than the lead, the lane sessions cannot commission judgment work onto the lead's backend, so the verdict has a home in the lead even when the execution has none.

## CodeRabbit endpoint measurement

The CodeRabbit predicate in the stub assumes the **combined** status endpoint, `/commits/{sha}/status`, which returns only the most recent status per context — measured on this repository, exactly one CodeRabbit entry per head. That assumption is the load-bearing part, so it is stated rather than left implicit: do not substitute the plural `/commits/{sha}/statuses`, which returns the full history newest-first. Measured on one head there: five CodeRabbit entries running from `Review queued` through `Review completed`, so a positional pick on that endpoint is wrong in one direction or the other — `last` selects the oldest. Where history is genuinely wanted, select by maximum `created_at` rather than by position.

Branch protection is not the lever here. The status state is `success` in precisely the failing case, so adding CodeRabbit to the required contexts would admit the unreviewed pull request just as readily. The distinction lives in the description, and only a read of the description surfaces it — which is why an automated merge gate closing this hole does not close it on the path a human merges by hand.

## Review lens selection

`review` is not one thing. The lead picks the lenses from what the card actually changed, and states the choice in the `sync` dispatch so the sync gate runs that review rather than re-deriving it:

| Card touched | Lenses to instruct |
|---|---|
| Auth, session handling, input parsing, external calls, secrets, file/path handling | `--security` (add `--deep` when the surface is reachable by untrusted input) |
| Non-trivial logic across several files | `--deep` (adversarially verified multi-phase scan) |
| Whole-tree sweep rather than a diff | add `--repo` |
| Suspected over-engineering | `--lean` (advisory only; applies no fixes) |
| UI or design-system surface | `--design`, and `--critique` after the build |
| Small, local, low-risk diff | no flag — the default 4-perspective pass is enough |

## The `/clear` handoff between phases — message structure

The lead's message to the operator states three things, in this order:

1. **What closed** — the card, the phase, and the evidence that was read.
2. **Which session to `/clear`** — by name, so the operator clears the right terminal.
3. **What happens next** — the column the card moves to, and which session will be instructed once the clear is done.

## Isolation rationale

Two properties make the shared checkout the wrong place for a card:

- Several sessions read it at once, so a branch switch, a `git stash`, or a `git add -A` there sweeps another session's uncommitted work into a commit that was never meant to carry it.
- A card outlives a phase. Its worktree spans run through sync, which is why disposal is triggered by the merge rather than by the phase finishing.

## Verification load incident record

Sessions share one machine as surely as they share one checkout, and verification is where that sharing goes wrong. Measured on a day when it did: load average reached 413, a neighbouring workspace's build took two and a half minutes, and its browser tests timed out — not because anything was wrong with them, but because four lanes were each running the full test suite at once and a full suite there takes five to ten minutes.

The never-spawn-background-load rule (normative statement lives in the stub) comes from the same incident's second cause: a verification recipe started eight spin loops to test behaviour under CPU contention and placed its kill line *after* the long test command. The agent finished before reaching it, and twelve spinners ran orphaned for thirty-seven minutes.

The same day supplied the reason contention and flakiness feed each other: a failing test left an unbounded spin-loop goroutine running, which burned a core for the remainder of that package's run and slowed every test after it. Load makes tests fail; failing tests can generate load.

## The pre-dispatch cross-check

The stub's two [HARD] clauses are the rule; this section is why each is worded the way it is.

**The failure was not a misread — it was that nothing required looking.** Several cards sat `queued` while each already carried an open pull request, and one sat `queued` while its fix was already an ancestor of the integration branch. The second was discovered only after a lane had started work on it, which cost the whole lane — the only sub-case in the record with a quantified cost. A lead with perfect tooling available would still have dispatched blind, because reading was optional.

**Why "reports, never vetoes" is a separate clause with its own criterion.** The obligation to look and the prohibition on acting are different rules, and the second is the fragile one. The operator has already picked the card; a lead that then withholds the dispatch because it found an open pull request has overridden an operator act rather than informing one — the same de-facto-authority hazard the read-only ruling on the queue tooling exists to prevent.

The hazard is invisible after the fact. A clause requiring the lead to read and report, and a clause authorizing the lead to refuse, produce identical transcripts up to the moment the card does not move; nothing downstream distinguishes "the operator withdrew it" from "the lead declined to send it". No mechanical check can separate the two readings, so the wording is the only control there is — which is why the literal `confirms or withdraws` is pinned by its own acceptance criterion rather than folded into the obligation clause.

**The tooling is a convenience, not the obligation.** `moai todo pr <id>` answers both halves in one read, but the clause is satisfiable by hand (`gh pr list`, then `git log` against the integration branch) and was written to be: the doctrine landed before the tooling, and the interval between them is a real operating condition rather than a paper one.

## The PR-title carrier

**Why the id leaves the branch name and lands on the PR title.** Neither name can serve both readers. The branch name is read by a human scanning `git branch` or a pull-request list, who learns nothing from an opaque card id and everything from a descriptive slug. The PR title is read by a resolver mapping pull requests back to cards, which needs a token it can match exactly. The branch-name rule and the title rule therefore assign different jobs to different names, and a reader meeting both [HARD] clauses cold will suspect a contradiction where there is none — which is why the stub states the non-contradiction outright instead of leaving it to be inferred.

**The carrier measurements.** Scanning a set of open pull requests for card tokens, three carriers behave differently:

| carrier | recall | precision | verdict |
|---|---|---|---|
| PR title | ~64% | every token present named the delivering card | precise, incomplete |
| PR body | complete | poor — one pull request carried five tokens for one card | complete, noisy |
| commit messages | high, and **wrong** on the worst case | worst of the three | unusable for attribution |

The commit carrier deserves its own note, because it is the one the isolation rule already makes [HARD]. A branch that merges the integration branch inherits every other card's commits, so its token set scales with integration rather than with the card: in the measured worst case a pull request carried a dozen-plus card tokens and its own delivering card was **not among them**. Being mandated as a traceability carrier does not make it a usable attribution index.

**The fix for ambiguous parsing is a naming convention, not a smarter parser.** The title carrier is already the precise one — where a token is present, it names the delivering card. It is incomplete only because nothing required it. Requiring it is what turns a resolver from a heuristic into a lookup, and no amount of parser sophistication substitutes for the missing token.

**Why a token in a batch pull request's title would be worse than none.** A release or batch pull request delivers no single card. A card token in such a title would name a card the pull request does not deliver, and a resolver reading titles would attribute confidently and wrongly — a silent error, where the absent token merely leaves the card unresolved and visible as such. The scope restriction is therefore load-bearing, not politeness.
