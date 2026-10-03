---
description: "Detail companion for factory-dispatch.md — terminology, card classes, dispatch-cycle naming, sync-gate review-lens table, /clear message structure, isolation rationale, verification-load incident record, sub-agent-first design intent, manager-lead working mode, sub-agent execution within a lane, Factory in-lane 3-stage, pre-dispatch cross-check rationale, PR-title carrier measurements, pre-merge settings-drift assertion"
paths: "**/factory-dispatch*.md,**/.claude/agents/moai/manager-lead.md,**/.claude/skills/moai/workflows/gtd.md"
---

# Factory Dispatch — Detail Companion

> Detail companion of `factory-dispatch.md` (the always-loaded stub). The stub keeps every [HARD] rule, prohibition, and cross-reference; this file owns the long tables, the dispatch-cycle walkthrough, incident narratives, and worked rationale. Load when classifying a card, routing a card to a lane, or choosing review lenses for a card's sync gate.

## Design intent — sub-agent-first token discipline

The factory leader and lane sessions keep **only orchestration** in their context windows. Every unit of real work — research, authoring, implementation, verification sweeps — is delegated to `Agent()` sub-agents; the verbose output (tool results, file dumps, test logs) stays in their window, and only summaries return to the session. The token rationale: a session surviving an entire card — or a Factory Mode batch — would otherwise accumulate every card's raw output on the always-loaded prefix. The sections below express this structurally.

## Terminology — the factory vocabulary

`factory-dispatch.md`, `sprint-round-naming.md`, and the operating notes share a working vocabulary. Each term gets one definition and one example; the sections below assume these meanings.

| Term | Definition | Example |
|---|---|---|
| **lane** | One parallel work stream that carries a card end to end: one session paired with one worktree. A lane is a swimlane — a band reserved for one stream of work; parallel streams never interleave and never share a working tree. "Lane-local verification" = that lane runs only the tests its own change can affect. | `lane-2` working in worktree `.moai/worktrees/t0` is one lane. |
| **card** | One unit of work in the queue, entered by the operator via `/moai gtd "<description>"` and referred to by a short id. A card owns one worktree, one progress record, and its completion evidence. | `t0` — a one-line fix card. |
| **backlog** | The entry queue. No session owns it by design — work enters only when the operator puts it there. | `/moai gtd "rename hint is stale"` appends a card to the backlog. |
| **leader** | The single coordinating session (`moai cc -f`). Routes cards to lanes on evidence it read itself, asks the operator to `/clear` a lane between cards, never writes code. | The session that dispatched a card with its worktree instruction. |
| **run-id** | The short identifier the leader prints at launch. It lives in `MOAI_KANBAN_ID` and the leader socket path — no session name carries it, the leader's included (t133): a lane is named `lane-<n>`, and a second live claim on a label takes the next free number. | `a1b2c3` — printed in the leader's bootstrap notice; the session itself is named `leader`. |
| **worktree** | The isolated checkout where a card's work happens — created by `moai worktree new` or a supported launcher and entered through one, never a raw worktree add. The directory carries the card id; the branch carries `WT-<slug>` (shape: the stub § Isolation is provisioned by MoAI). A worktree outlives a stage: one spans plan through sync. | `.claude/worktrees/t0` on branch `WT-todo-queue`. |
| **dispatch** | The leader's instruction to one lane: a pointer (card id, SPEC id, entry command, completion signal), never a copy of the work. Written in the operator's conversation_language. | "card: t0 — wt: EnterWorktree(t0) … evidence: .moai/reports/t0/". |

Factory Mode lanes are labelled `lane-<n>`; a lane owns a card end to end (§ Factory in-lane 3-stage).

## Report milestones ↔ queue cards

The stub's [HARD] rule — a `## Card Cross-Check` section per milestone-bearing report, mapping claims verified against the queue — exists because report→card linkage used to live in one person's memory: a report declared milestones, cards were issued separately, and nothing reconciled the two — milestones surfaced with no card, cards claimed milestones already landed.

The mechanical check runs the same comparison the leader states by hand:

```
moai graph build && moai graph query --milestones-no-card
```

It writes the report-milestone and milestone-card edges from each report's Card Cross-Check table and lists every milestone whose claimed card is missing from the live queue (queued/picked; dropped does not qualify). "Not in live queue" covers both completed and never-issued cards — resolve each flag with `git log --oneline --grep 'merge: <card-id>'` before issuing a new card.

## Card classes — not every card needs every stage

Most of what accumulates in the backlog is chores: a one-line fix, a stale reference, a renamed flag. Sending those through `plan → run → sync` costs more in ceremony than the change is worth. The leader classifies each card as it leaves `backlog` and names the entry stage in the dispatch.

| Class | Shape | Path |
|---|---|---|
| A — direct close | The change is one file and one line, there is no design judgement in it, and CI catches the regression | One lane carries the card through to a pull request; `plan` is skipped |
| B — defect, cause unknown | Something is wrong and the cause has not been established | `run → sync`; `plan` is skipped, so no SPEC exists |
| C — design change | The change contains a decision, or spans subsystems | All three stages |

The Class-A evidence rule is the same shape as the CodeRabbit section of the stub: a class that skips review on a claim nobody checked is exactly the unobserved-claim hazard this rule forbids everywhere else; writing the justification down is not the same as verifying it.

The parallelism comes from the across-cards axis: each lane carries a whole card, so three lanes put three cards in flight (§ Sub-agent execution within a lane for the within-card axis; § Factory in-lane 3-stage for the stages). Research fan-out during `plan` is Class-C-only.

## The dispatch cycle

Each dispatch is one send from the leader to one lane session, and the lane carries the card from there:

```
[operator picks a card]  →  leader routes WHOLE to a free lane  →  lane: plan → run → sync  →  [leader reads evidence, marks done]
```

Dispatch is addressed by session name. Lanes are named `lane-<n>`; a label held by a live session is bumped to the next free number, and no run id travels in any session name, the leader's included (one run per machine; the id lives in `MOAI_KANBAN_ID` and the leader socket). `ListAgents` lists live sessions and says when it could not check them all; send with `SendMessage({to: "<name>", message: "…"})`, using the short reference the listing prints when a bare name is ambiguous.

A bare name fails in two different ways, and only one of them announces itself. The refusal case is the harmless one: the runtime cannot pick between same-named sessions, says so, and the send is re-issued with the short reference the listing prints. The other case says nothing at all.

**The silent case — an in-process mailbox takes the name.** While the team namespace is active (`CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS=1`, which ships enabled in settings and in the distributed template), a name carried by both a live lane session and an in-process teammate mailbox resolves to the **mailbox**. Nobody reads it, the lane never learns a dispatch existed, and the send reports success. The result's shape is the only separator:

| Result shape | Where the dispatch went |
|---|---|
| no `routing` object — the result names the peer (`… (another Claude session on this machine)`) | the lane session — **delivered** |
| a `routing` object, e.g. `routing: {sender: "team-lead", target: "@<name>"}` | an in-process team mailbox — **lost** |

So read a `routing` object as a failure signal and re-send to the `name [ref]` the listing printed. The two rows are not equally measured: the absent-`routing` row is directly observed — dispatches arriving normally, repeatedly — while the present-`routing` row rests on one reported contrast experiment. That is enough to act on (re-sending a delivered message costs one duplicate; missing a lost one stalls a run), but it is one measurement rather than a pattern.

Detecting the collision is the second line of defence. The first is not creating it: the leader spawns its coordination agent unnamed precisely so no in-process teammate ever carries a name a lane session also answers to (§ The leader works through manager-lead). Where that holds, this section never fires.

**The collision is conditional, not universal.** A name no in-process teammate shares still delivers on the bare form — so "always address by reference" would be a false rule, and the binding one is *read the result*. Where a spawned teammate plausibly shares a lane's label, address by `name [ref]` from the first send rather than after a lost one.

This does not soften what moves a card. The queue on disk is still the delegation, evidence on disk is still what advances a card, and a dispatch that silently missed shows up as a card that never progressed — the message was only ever a nudge. Messaging-side mechanism: `cross-session-messaging-detail.md` § Addressing, sending, and replying.

**A listing is not always complete, and an incomplete one cannot prove absence.** From Claude Code 2.1.234 the listing says when your account's session list was too long to check completely, rather than leaving unseen sessions to read as absent. That disclosure is what the stub's fault clause turns on: the leader concludes a lane is empty from a listing that checked everywhere, and a listing that reports it could not is a **gap** — it re-checks and reports the gap, never a fault. Concluding otherwise reports a running lane as dead, which is the failure the upstream change exists to prevent, and is an unobserved absence claim under `verification-claim-integrity.md` §1 (the absence of a signal is not evidence of its subject's absence). This binds only the *absence* direction: a session the listing DOES show is present whatever else it could not reach.


Each instruction carries, at minimum: the card, the SPEC ID once one exists, the entry command to run, and the completion signal to write. Keep it a pointer, not a copy — the lane reads the SPEC artifacts itself rather than receiving them inline.

**`sync → done` is the same act with the dispatch removed.** No session occupies `done`, so the leader reads the lane's completion evidence and records the terminal transition itself.

### Dispatch language — classification rationale

[HARD] A dispatch is written in the operator's `conversation_language` (normative statement lives in the stub). This is a classification, not an exemption from the language rules, and it needs no change to either of them:

- `agent-common-protocol.md` § Language Handling already opens with the opposite of an English default — agents receive and respond in the configured `conversation_language`. What it fixes to English is code, identifiers, and names. A dispatch is prose, so the rule was never against it; it simply did not name cross-session messages in its list.
- `moai-constitution.md` § Response Language reserves English for internal agent communication, but the axis that clause sits on is stated one line above it: user-facing responses go in the operator's language. A message a human reads is user-facing by that rule's own criterion, so putting a dispatch in the operator's language applies the constitution rather than carving an exception out of it.

### Dispatch format — rationale

The address-block format (normative statement lives in the stub) is the "pointer, not a copy" rule above made mechanical: every field is an address the lane resolves by reading what it names. It also settles the Dispatch language rule by construction — a block of pure addresses has nothing to translate, while the leader's reports to the operator (progress notes, `/clear` requests) remain in the operator's `conversation_language`.

## The leader works through manager-lead

The factory leader session does not draft its own coordination: it spawns the `manager-lead` agent and works through it. The split:

- **Dialogue.** manager-lead holds the operator conversation — card selection, `/clear` prompts, blocker surfacing. The blocker-report discipline is unchanged (`agent-common-protocol.md` § Blocker Report Format): manager-lead returns blocker reports, and the session's `AskUserQuestion` remains the user channel. The rename moves no part of the user channel into the agent.
- **Dispatch.** manager-lead routes cards and reads evidence in the background while the dialogue continues — neither the conversation nor the lane coordination blocks the other, because sub-agents run in the background by default and the session stays free between their returns.

[HARD] **Spawn manager-lead UNNAMED.** Under `CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS` a named spawn converts to an in-process teammate, and that path carries two measured hazards: an observed output-loss discrepancy (one same-version session watched a named spawn become an output-less in-process teammate — the two-sided evidence is recorded in CLAUDE.md §15), and in-process teammates cannot spawn background subagents, which forces foreground-only execution and breaks the dialogue-never-blocks property above. An unnamed spawn keeps it a plain subagent (related: `orchestration-mode-selection.md` §C.1).

### Deputy mode — background coordination off the leader's turn

The leader's turn loop is the scarcest surface in a factory run: a dispatch send, a CI watch, and a CodeRabbit poll each occupy it serially, and an occupied leader judges nothing else. The deputy exists to move that occupancy. The leader session (still through manager-lead, still UNNAMED) delegates coordination duties to a background deputy instance whose charter — the delegable/retained matrix, the delivery-shape verification protocol, the standing messaging hazards — is codified in the agent itself (`.claude/agents/moai/manager-lead.md` § Deputy dispatch surface).

**Why the deputy is resident rather than optional.** An optional mechanism is one a loaded leader never reaches for: the spawn's turn is the same turn the queue waits on, so the delegation is deferred exactly when it would pay most. Making the spawn a batch-start obligation (stub § Deputy dispatch surface) removes the decision from the moment of pressure. The cost is one background agent per batch whether or not it is needed — an unused deputy costs one spawn; an unspawned one costs every dispatch after it.

**Reading the completion report.** A lane's completion report names evidence paths; reading those paths is several tool batches, and every one of them sits in the leader's turn. The deputy performs that read and returns a `RECOMMEND:` summary naming what it read. What moves is the occupancy, never the obligation: the leader still reads the evidence before advancing the card, because a deputy report is one more claim until the evidence under it has been read (stub § Completion is read, never trusted). A deputy that returns a conclusion without naming the paths it read has returned nothing usable.

**Round-report drafting.** The measurement batches and the table scaffolding are mechanical and belong to the deputy; the figures the leader will personally assert are re-authored by the leader, because a cited figure carries its measurer's attribution and an unattributed one is a defect (`verification-claim-integrity.md` §2). In the report the two attributions sit side by side — deputy-measured values naming the deputy and the path, leader-asserted values naming the leader. Per-round files plus an index is the structural half: a single file rewritten from its header every round makes each round's authoring cost grow with the batch's age — a drafting cost the delegation cannot remove.

**Watching without polling.** Repeated listing rounds to learn whether a lane has finished are pure turn occupancy that returns no information most of the time. One `notify_when_idle` request replaces the loop (`cross-session-messaging.md` § An idle notice is a scheduling hint) — opt-in, one-shot, so a second notice needs a second request. Its boundary is inherited by citation and not restated: the notice says *when to go look* and nothing about what the evidence says — a session goes idle on finish, on a permission prompt, and on death, and the notice cannot separate those. Advancing a card on the notice alone is an unobserved completion claim.

**What the deputy does in the background:**

- **Dispatch sends with delivery-shape verification.** The deputy sends the fixed-field address blocks for ALREADY-PICKED cards and reads every send result. A `routing` object on the result means an in-process mailbox took the block — lost, not delivered — and the deputy re-sends to the `name [ref]` form the `ListAgents` listing printed (§ The dispatch cycle). A rapid-burst refusal is read and reported; the queue already carries the delegation, so a refused nudge stalls nothing.
- **Bounded CI-watch polls.** The deputy polls a card's checks to terminal states and returns those states — the states, never a judgement about them.
- **CodeRabbit two-condition reads.** The deputy reads the combined-status `CodeRabbit` entry (state `success` AND description `Review completed`) plus the `Merge Risk:` line matching the current `headRefOid`, and REPORTS both conditions to the leader. It never adjudicates the slot-wait outcome: a card carrying `Review rate limited` is reported as exactly that, and the leader decides what it means.
- **First-pass evidence reading.** The deputy reads a card's completion evidence and returns findings as recommendations, each prefixed `RECOMMEND:` — never a verdict token.
- **Summary reporting.** The deputy folds its observations into a single report addressed to the leader.

**What returns to the leader's turn:** `RECOMMEND:`-prefixed recommendations, terminal CI states, delivery confirmations and refusals, and the two-condition CodeRabbit read — the deputy's report is input to the leader's judgement, never a substitute for it.

**What does not change — the structural principles:**

- **The verdict's home.** The final PASS/FAIL remains the leader's, read from evidence on disk (§ The verdict's home). A deputy recommendation promoted to a verdict is exactly the adjudication-promotion failure the report/verdict split exists to prevent.
- **Completion is read, never trusted.** The deputy's first-pass read does not discharge the leader's own evidence-read obligation. The leader advances a card on what the leader read; a deputy report is one more claim until the evidence under it has been read.
- **The queue is the channel.** The deputy's `SendMessage` is a nudge, never the delegation (stub § The delegation channel is the queue); card advancement never depends on a message arriving, whoever sent it.

The retained powers — final verdicts, final merge approval, operator gates, card issuance and `done`, CodeRabbit adjudication, cross-session dispute coordination — are enumerated under the `DEPUTY-RETAINED-BY-LEAD` marker in the agent charter; the stub carries the [HARD] boundary clauses.

Messaging stays a nudge (stub § The delegation channel is the queue, § Completion is read, never trusted). The rename changes who drafts the coordination, not what counts as delegated or as done.

## Sub-agent execution within a lane

The lane session orchestrates only: each stage's execution — plan authoring, run implementation, sync sweeps — is spawned as `Agent()` sub-agents whose output stays in their windows. The lane merges and inspects results through the evidence they write, not by absorbing their transcripts. Across cards the parallelism is the lanes themselves: one card per lane, one worktree per card.

Discipline:

- **Ceiling: 10 concurrent agents per lane.** The launcher injects `CLAUDE_CODE_MAX_CONCURRENT_SUBAGENTS` at that value; a lane does not raise it.
- **Parallel fan-out is for read-heavy work plus isolated writes.** Never run multiple write-capable agents on the SAME card concurrently — one card, one writer at a time; stages of one card in series. (Intra-card research fan-out during `plan` remains Class-C-only — § Card classes.)
- **Write-capable sub-agents spawned in parallel MUST carry `isolation: "worktree"`** (the Agent tool's isolation parameter), so concurrent file writes cannot collide even outside the card directories: each write agent works in its own worktree copy, and the lane integrates via evidence and merge. Read-only fan-out (investigation, audits) stays unisolated — worktree setup cost buys nothing there. L1 `Agent(isolation: "worktree")` semantics, the relative-path prompt rule, and the lifecycle are owned by `worktree-integration.md` and are cross-referenced, not restated here; the stub's "dispose only after the remote merge lands" and "exit the previous worktree before a new card's" rules remain the SSOT for card worktrees.
- **Verification stays lane-local.** The full suite is CI's job (stub § Verification load is lane-local).
- **Stagger same-type spawns**: spawn one worker first and the rest once it has started producing, so the later spawns read the first one's prompt cache (`cache-aware-execution.md` directive 2).
- **Sub-agents are not sessions.** The stub's "launched by hand, no peer-spawning" boundary governs sessions; an `Agent()` worker inside a lane is ordinary in-session orchestration, and the lane remains the card's owner of record.

## Factory in-lane 3-stage

Factory Mode (`moai cc -f` / `moai glm -f` for the leader, `-l` for each lane) rests on whole-card ownership: one leader plus `lane-<n>` sessions, each lane owning one card end to end. Lanes are launched by hand; the leader keeps the run-id, the queue, and the verdict.

- **Routing.** The leader routes a card WHOLE to a free lane — free means the lane's previous card reached `done` and its evidence was read. A lane busy on a card is not addressed; with every lane busy, the card waits in the queue rather than being dispatched. The address block carries `cmd`, which names the entry stage the class prescribes (`/moai plan` for C, `/moai run` for B, the direct close for A), and the lane proceeds through the remaining stages without further dispatches.
- **Serial stages, sub-agent execution.** Plan completes before run begins, run before sync — a lane never runs two stages of the same card concurrently. Within a stage it fans out sub-agents per § Sub-agent execution within a lane.
- **Standing spawn authority (card t224's normative home).** The lane's right to those sub-agent spawns is standing, not per-dispatch: the SessionStart join notice carries the authority sentence verbatim, and the runtime's default "don't spawn unless the user asks" guidance does not bind a lane (the leader is not the lane's user, so leader approval can neither grant nor revoke what the bootstrap grants). Scope is the Status Transition Ownership Matrix's specialist for the stage at hand; depth is one — spawned agents are leaves. Observed defect this closes: two tk8hce lanes refused to spawn `manager-spec` and edited SPEC bodies directly, routing artifact writes around the ownership matrix.
- **Classes name the entry stage.** A/B/C name which ceremonies a card skips (B skips `plan` and carries no SPEC; A goes straight to the close); no card changes sessions. Every lane runs the serial stages for whatever its card still needs.
- **The `/clear` boundary is between cards.** A factory lane is cleared once its card reaches `done`, before the next card is routed to it, exactly as the stub's `/clear` rule requires.
- **Evidence, verdict, integration unchanged.** The lane writes the same completion signals; the leader still reads evidence and owns the final PASS/FAIL; release-branch integration is still lane work under the stub's rules. The deputy surface applies too: the factory leader may delegate dispatch sends and watches to the coordination deputy (§ The leader works through manager-lead → Deputy mode), while the leader keeps the verdict and the batch pull request.

## The card-review stage

A lane closes its card through an ordered list of stages. The list fixes where the card-scope self-review sits, whichever factory lane carries the card:

1. `[run-exit]` Run convergence and lane-local verification (stub § Verification load is lane-local).
2. `[card-review]` The card-scope self-review described below.
3. `[integration]` The integration window and the merge into the batch's release branch (stub § Integration into the release branch is self-served).
4. `[report]` The completion report naming every evidence path, then the `/clear` the stub requires.

Running `[card-review]`:

- **Call.** The lane, or a sub-agent it spawns, calls `codex_review` (the required backend) with `scope: card` and `project_root` set to its own `git rev-parse --show-toplevel`; `glm_review` takes the same arguments and adds a second opinion from the GLM backend. `scope: card` resolves through the scope resolver the turn-end gate uses and recomputes the merge base on every call. On a tree that is not a card worktree the tool returns `inconclusive` and reviews nothing.
- **Fallback.** Where the running MCP server predates the tools (they are absent from its tool list), the codex leg runs as `moai verify codex-review --project-root <tree>` and the GLM leg is recorded as unavailable. At card intake the lane compares the running MCP server's build (the server instructions line, or `moai doctor`) with `moai version`; on a mismatch, or a required tool absent from the server's tool list, it records `mcp: server=<build> cli=<build> fallback=CLI` in the card's progress record and uses the CLI equivalent for each affected tool.
- **Evidence.** The result is written to `.moai/reports/<card-id>/card-review.md` with the backend, the base commit, the verdict, the findings, and a disposition for each finding (fixed, carried over, or not adopted with the reason). The card's progress record cites that path.
- **Ceiling.** After a repair the lane re-reviews at most 2 times, the same ceiling as the run-exit verify gate's re-entry limit. A finding still open at the ceiling is written down with its disposition and raised to the leader; the lane does not widen the ceiling.
- **Advisory.** The result carries no authority: the card's PASS/FAIL stays with the leader's evidence read and the independent audit. A missing reviewer or an `inconclusive` result is recorded as "review not performed", never as a pass. The stage is a card step rather than a turn-end hook, so it never blocks a turn.
- **Continued firing.** A stage that silently stops being run shows at the leader's completion read: the declared evidence list names `card-review.md`, and a card whose progress record neither cites it nor records why it is absent stays in its stage (stub § Completion is read, never trusted). No hook enforces the file; the leader's read is the answer to what would look different if the stage stopped.
- **One review per diff.** A repository that also enables the card-scope turn-end gate for its lanes reviews the same card diff twice; enable one of the two.
- **The leader.** A leader reviews its own internal output with the same tools at `scope: uncommitted`. In the primary checkout that scope covers the whole shared working tree, which may include another session's work.

## The verdict's home

The stub keeps the norm — the final PASS/FAIL verdict is the leader's, read from evidence on disk, never delegated to the lane that produced the work. The division is structural, not ceremonial: where the lanes run on a different backend than the leader, the lane sessions cannot commission judgment work onto the leader's backend, so the verdict has a home in the leader even when the execution has none.

## CodeRabbit endpoint measurement

The CodeRabbit predicate in the stub assumes the **combined** status endpoint, `/commits/{sha}/status`, which returns only the most recent status per context — measured on this repository, exactly one CodeRabbit entry per head. That assumption is the load-bearing part, so it is stated rather than left implicit: do not substitute the plural `/commits/{sha}/statuses`, which returns the full history newest-first. Measured on one head there: five CodeRabbit entries running from `Review queued` through `Review completed`, so a positional pick on that endpoint is wrong in one direction or the other — `last` selects the oldest. Where history is genuinely wanted, select by maximum `created_at` rather than by position.

Branch protection is not the lever: the status state is `success` precisely in the failing case, so adding CodeRabbit to the required contexts admits the unreviewed pull request just as readily. The distinction lives in the description — which is why an automated merge gate closing this hole does not close the path a human merges by hand.

## Review lens selection

`review` is not one thing. The leader picks the lenses from what the card actually changed, and states the choice in the dispatch so the sync gate runs that review rather than re-deriving it:

| Card touched | Lenses to instruct |
|---|---|
| Auth, session handling, input parsing, external calls, secrets, file/path handling | `--security` (add `--deep` when the surface is reachable by untrusted input) |
| Non-trivial logic across several files | `--deep` (adversarially verified multi-phase scan) |
| Whole-tree sweep rather than a diff | add `--repo` |
| Suspected over-engineering | `--lean` (advisory only; applies no fixes) |
| UI or design-system surface | `--design`, and `--critique` after the build |
| Small, local, low-risk diff | no flag — the default 4-perspective pass is enough |

## The `/clear` handoff between cards — message structure

The leader's message to the operator states three things, in this order:

1. **What closed** — the card and the evidence that was read.
2. **Which session to `/clear`** — by name, so the operator clears the right terminal.
3. **What happens next** — the next card the lane is routed, and which lane will be instructed once the clear is done.

## Isolation rationale

Two properties make the shared checkout the wrong place for a card:

- Several sessions read it at once, so a branch switch, a `git stash`, or a `git add -A` there sweeps another session's uncommitted work into a commit never meant to carry it.
- A card outlives a stage — its worktree spans plan through sync, which is why disposal triggers on the merge rather than the stage finishing.

## The pre-merge settings-drift assertion

The stub's `acquire` clause has a narrow subject: the **tracked** `.claude/settings.json` in the tree the lane is standing in when it takes the window. Twice, a card worktree has been found with that file modified in the working copy and no author anyone could name. Neither was caught by the merge window — the first surfaced because a merge happened to be refused, the second in a full sweep of every worktree nine days later. Authorship was closed as unattributable, which leaves detection as the only end that can be fixed: an unnoticed modification cannot be traced, but it can be caught before it rides a merge.

`acquire` is the right point because of cwd. The procedure has a lane run it from its own card worktree and only then enter the release worktree, so the tree standing under `acquire` is the tree about to be merged. Code does not enforce that ordering, so the report always names the tree it measured: measuring a different tree is acceptable, measuring one quietly is not.

The predicate:

```
git --no-optional-locks status --porcelain -- .claude/settings.json
```

Three properties of it are load-bearing.

- **`--no-optional-locks` is mandatory.** A plain `git status` takes an index WRITE lock for tens of milliseconds. A check that runs immediately before a merge, on a machine with several lanes on it, would otherwise manufacture the contention it exists to protect against.
- **The verdict is the match count, never an exit code.** Zero lines is a pass, one or more is drift.
- **A failed measurement is its own state.** `clean` / `drift` / `undetermined` are three states, not two with a fallback: an unmeasured tree that reports `clean` is precisely the nine-day silence, and the absence of a signal is not evidence of cleanliness.

On a hit: the working copy is preserved under the primary checkout's state directory (visible to the leader, who is not in the lane's tree), one row goes into `ledger.jsonl` beside it, and the path plus its sha256 and size are reported. Contents are never printed and never committed — the file is runtime-written and can carry tokens, absolute paths, and pane ids; promoting it into history needs a human secret-scan first.

**Nothing is ever restored, reverted, or deleted.** For the same reason: an automatic restore of a file holding machine-specific values is itself data destruction. Disposal is a human decision.

Two surfaces, and neither is redundant. Without the `acquire` precondition, running the check would be a social protocol — the same gap the announcement rule named about itself. Without `moai integration preflight [path]`, there would be no way to ask the question without taking a window.

Refusal is the only part that is opt-in (`workflow.settings_drift_gate.enabled`, default off, the same posture as the repository's other guards). Detection, preservation, the ledger row and the report run on every `acquire` regardless — the observed failure was nine days of nobody looking, not nine days of nothing being blocked, so gating the observation would remove the very thing the default-off posture rests on. Where refusal is on, `--allow-settings-drift` records the window anyway and stamps the bypass into the lock record; `--force` does not and must not, because it answers a different question ("take the window from a live holder") and one flag carrying both leaves the record unable to say which was meant. An `undetermined` verdict does not refuse — it is reported loudly and the window is recorded, matching every other guard's fail-open posture on uncertainty.

## Verification load incident record

Sessions share one machine as surely as they share one checkout, and verification is where that sharing goes wrong. Measured on a day when it did: load average reached 413, a neighbouring workspace's build took two and a half minutes, and its browser tests timed out — not because anything was wrong with them, but because four lanes were each running the full test suite at once and a full suite there takes five to ten minutes.

The never-spawn-background-load rule (normative statement lives in the stub) comes from the same incident's second cause: a verification recipe started eight spin loops to test behaviour under CPU contention and placed its kill line *after* the long test command. The agent finished before reaching it, and twelve spinners ran orphaned for thirty-seven minutes.

The same day supplied the reason contention and flakiness feed each other: a failing test left an unbounded spin-loop goroutine running, which burned a core for the remainder of that package's run and slowed every test after it. Load makes tests fail; failing tests can generate load.

## The pre-dispatch cross-check

The stub's two [HARD] clauses are the rule; this section is why each is worded the way it is.

**The failure was not a misread — it was that nothing required looking.** Several cards sat `queued` while each already carried an open pull request, and one sat `queued` while its fix was already an ancestor of the integration branch. The second was discovered only after a lane had started work on it, which cost the whole lane — the only sub-case in the record with a quantified cost. A leader with perfect tooling available would still have dispatched blind, because reading was optional.

**Why "reports, never vetoes" is a separate clause with its own criterion.** The obligation to look and the prohibition on acting are different rules, and the second is the fragile one. The operator has already picked the card; a leader that then withholds the dispatch because it found an open pull request has overridden an operator act rather than informing one — the same de-facto-authority hazard the read-only ruling on the queue tooling exists to prevent.

The hazard is invisible after the fact. A clause requiring the leader to read and report, and a clause authorizing the leader to refuse, produce identical transcripts up to the moment the card does not move; nothing downstream distinguishes "the operator withdrew it" from "the leader declined to send it". No mechanical check can separate the two readings, so the wording is the only control there is — which is why the literal `confirms or withdraws` is pinned by its own acceptance criterion rather than folded into the obligation clause.

**The tooling is a convenience, not the obligation.** `moai gtd pr <id>` answers both halves in one read, but the clause is satisfiable by hand (`gh pr list`, then `git log` against the integration branch) and was written to be: the doctrine landed before the tooling, and the interval between them is a real operating condition rather than a paper one.

**Why the read can be a skip input for one card and report-only for another.** Put plainly, a pull request or landed state is a skip input for a queued candidate the session chose and is report-only for an operator-picked card: the operator already chose the second, and passing it over would override that act; the session chose the first itself, so skipping it overrides nobody. The two clauses govern different cards, so "reports, never vetoes" is untouched.

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
