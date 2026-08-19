# Kanban Dispatch Protocol

How the **lead** session of Kanban Mode moves a card across the board: what admits work, who is told to do it, how completion is judged, and when the operator is asked to `/clear`.

> **Loading scope**: Intentionally always-loaded. A session learns it is the kanban lead from the SessionStart context, not from a file path, so a `paths:`-restricted rule would never reach it.

> **Detail companion**: `kanban-dispatch-detail.md` owns the long tables, dispatch-cycle walkthrough, incident narratives, and rationale — now also per-card fan-out, Factory in-lane 3-stage, and the `manager-lead` working mode. The stub keeps every [HARD] rule and pointer; load the companion when moving or classifying a card, or choosing review lenses.

## Scope — when this rule is live

This rule binds a session whose SessionStart context declares **Kanban Mode** with the `lead` role. In every other session it is inert: a companion session (`plan` / `run` / `sync`) receives instructions, it does not dispatch them, and a session outside Kanban Mode has no board to move.

Kanban Mode is entered with `moai cc -k` (or `moai glm -k`), which elects one lead and prints one launch command per companion role. Companion sessions are launched **by hand, one per terminal** — a session cannot launch another, and no peer-spawning mechanism exists or is wanted.

The lead session works through the `manager-lead` agent: it holds the operator dialogue (the session's `AskUserQuestion` stays the user channel) while dispatching parallel work in the background — neither blocks the other. Lead and lane sessions orchestrate only; real work runs in sub-agents (design intent + spawn rules: `kanban-dispatch-detail.md` § Design intent, § The lead works through manager-lead).

One boundary: nudge delivery rides on cross-session messaging, absent on native Windows and off under some providers, versions, and flags (`cross-session-messaging.md` § Availability constraints) — an absent channel fails quietly; the lead surfaces it, and the queue keeps working without it.

## The board

Five columns, fixed and ordered: `backlog → plan → run → sync → done`. `backlog` and `done` have no owning session; the three working columns each map to exactly one companion role (`plan` / `run` / `sync`), which is what makes dispatch a lookup rather than a decision. There is no `review` column — the sync gate absorbs the review verdict and runs the lenses itself. Column-by-role table: `kanban-dispatch-detail.md` § The board.

## Entry into the board is an operator act

`backlog` has no owning session, and a lead that admitted cards on its own initiative would be **generating** work rather than scheduling it. Every card's origin is therefore the operator's request.

[HARD] **The lead is the queue's sole producer.** The operator asks; the lead turns the request into a card with `moai todo add "<description>"` (`moai todo` alone lists the queue). Production is the one queue mutation the lead performs on its own authority — translation, not invention: nothing enters the queue the operator did not ask for.

[HARD] **Promotion is the operator's act, always.** After a `/clear`, the lead presents the queued cards through `AskUserQuestion` and the operator picks; only then does the lead dispatch to `plan`. The lead never picks for the operator, never reorders by inferred priority, and never silently promotes a backlog item. An empty queue is a state to report, not a prompt to invent work.

## Report milestones ↔ queue cards

[HARD] **A milestone-bearing report under `.moai/reports/` carries a `## Card Cross-Check` section** — one table row per milestone, a `card` column holding the delivering card id or an explicit new-card marker. A mapping claim is verified against the queue (`moai todo`), never remembered. Before the lead turns a report into card requests, the request message states the full comparison — `N milestones → N cards` — naming every milestone with no card in the live queue. Detail: `kanban-dispatch-detail.md` § Report milestones ↔ queue cards.

## Card classes — not every card needs every column

The lead classifies each card as it leaves `backlog` and names the class in the dispatch: **A — direct close** (one file, one line, no design judgement, CI catches the regression; one session carries the card to a pull request, `plan` skipped), **B — defect, cause unknown** (`run → sync`; `plan` is skipped, so no SPEC exists), **C — design change** (a decision, or spans subsystems; all three working columns). Full table and rationale: `kanban-dispatch-detail.md` § Card classes.

[HARD] **Class A is admitted on checked evidence, not on an assertion.** Two of its three properties are mechanically checkable, and are checked and cited: the diff is measured (`git diff --stat` against the base, showing the one file) and CI is green **on the head that will merge**. The third — no design judgement in it — is a judgement, stated in the dispatch where the operator can disagree with it. A card that cannot cite both measurements is not Class A.

The justification is never "it is faster": speed is the effect of skipping the columns, not the reason for it — a card justified by speed alone is a Class C card being rushed.

**Class B skips `plan`, not the sync gate's review** — an unestablished cause is precisely what a review catches. The `run` session owns both the investigation and the fix; before the card leaves `run`, the cause-establishing evidence (the reproducing command + what it printed) is written into the card's progress record, and the run session names that path in its completion report so the lead reads the cause rather than taking it on trust.

**Work in progress: one card per worktree — two cards sharing a worktree run serially whatever columns they sit in, because they share a working tree and a branch.** A lane holding several cards in one column does it by per-card fan-out: parallel `Agent()` workers whose writes each stay inside their own card directory (`kanban-dispatch-detail.md` § Per-card fan-out; parallelism per class: § Card classes).

## The dispatch cycle

`[operator picks a card] → plan → run → sync → [lead marks done]` — each arrow is one dispatch from the lead to one companion session, addressed by name (`SendMessage`; `ListAgents` lists live sessions and says when it could not check them all). Each instruction carries, at minimum: the card, the SPEC ID once one exists, the phase command, and the completion signal to write. Keep it a pointer, not a copy — the companion reads the SPEC artifacts itself.

**`sync → done` is the same act with the dispatch removed.** No session occupies `done`, so the lead reads the sync session's completion evidence and records the terminal transition itself. Session-naming rules and the address-block walkthrough: `kanban-dispatch-detail.md` § The dispatch cycle.

### The delegation channel is the queue

[HARD] Work is delegated through the queue on disk, not through messages. The queue file resolves against the primary checkout from every linked worktree — one repository, one queue — so a card admitted from anywhere is visible everywhere, and that single-file visibility is what makes the queue a channel rather than a shared opinion.

A cross-session message is a nudge, never the delegation itself: delivery is not guaranteed, and a delivered message consumes the recipient's quota like a typed prompt. No dispatch depends on a message arriving. The disk record — card admitted, picked, done — moves the board; an unanswered message changes nothing.

Two properties of that nudge channel bear on dispatch. **A lane can be asked to report when it next goes idle** (`SendMessage` `notify_when_idle`, opt-in and one-shot), which spares the lead a polling loop — but [HARD] the notice is not the completion signal. A lane goes idle when it finishes, when it stops at a permission prompt, and when it dies; the notice cannot separate those, so it tells the lead *when to read the evidence* and nothing about what the evidence says. The card still advances on the evidence, per § Completion is read, never trusted. **And a nudge can be refused outright under fan-out**: nudging every lane inside one turn is a rapid burst, and the runtime refuses past the inbox's capacity rather than dropping silently (`cross-session-messaging.md` § Configuration surface). Read the send result; a refusal costs the board nothing, because the queue already carries the delegation.

### Dispatch language

[HARD] A dispatch is written in the operator's `conversation_language` — the operator watches it scroll past, which makes it user-facing output rather than internal agent traffic. The boundary is **who reads it**: an `Agent()` subagent prompt reaches no human and stays English. (Why this classifies rather than exempts: `kanban-dispatch-detail.md` § Dispatch language.)

What stays verbatim in every language: SPEC IDs, command names and their flags, file paths, session names, and technical identifiers. They are addresses rather than prose; a translated address does not resolve.

### Dispatch format

[HARD] A dispatch is a fixed-field address block, not prose. The fields:

```
card: <id>
spec: <SPEC-ID>
cmd: /moai run <SPEC-ID>
wt: .claude/worktrees/<name>
evidence: .moai/specs/<SPEC-ID>/progress.md
lens: --security --deep
```

- `card`, `cmd`, `wt`, and `evidence` are always present. `spec` joins once a SPEC exists; a Class B card, which skips `plan`, carries none, and its `evidence` names whatever record the lead will read instead. `lens` appears only in a `sync` dispatch — the review lenses the sync gate will run, the choice itself stated as an address rather than a sentence.
- `wt` names the new card's worktree, never a previous card's tree; where the lane may still be anchored elsewhere, it carries the exit-first instruction (`ExitWorktree` → `EnterWorktree(<card-id>)` → `git branch -m WT-<slug>`). The tree keeps the card id; the branch takes a descriptive slug — see the naming rule below.
- **No explanatory prose.** Procedure, background, and justification live in the card text and the SPEC artifacts the block points at; a dispatch that restates them makes the operator read the same thing twice. What does not fit a field belongs in the card, not around the block.
- **Ceiling: the block is at most 10 lines.** A dispatch that does not fit is trying to be a handoff; move the payload into the card and send the block.
- **[HARD] The send is read, not assumed.** A `routing` object on the result means an in-process mailbox took the block and it is lost; re-send to `name [ref]`. Conditional: `cross-session-messaging.md`.

## Completion is read, never trusted

[HARD] The lead advances a card on **evidence it read**, not on a companion's reply. Reply routing is not guaranteed to arrive, and a reply is a claim rather than an observation.

Before moving a card out of a working column, the lead reads the card's `progress.md` (and the verification evidence path the phase declares, if any) and confirms the phase closed. A missing, unreadable, or stale evidence file is a **gap** — the card stays put and the lead reports why. Absence of a failure signal is not a pass.

This applies equally to the operator: when the lead reports a column advanced, it names what it read.

**The final PASS/FAIL verdict is the lead's**, read from the evidence on disk and never delegated to the lane that produced the work — the executor judging its own output is the failure shape this section exists to prevent. Why the division is structural: `kanban-dispatch-detail.md` § The verdict's home.

### CodeRabbit is not read from `gh pr checks`

[HARD] A `gh pr checks` row naming CodeRabbit is not evidence that a review happened: the status is `success` **even when no review ran**, and the row prints `pass` byte-identically in both cases — only the description separates reviewed from unreviewed. A row counts only when BOTH hold:

1. The **combined** endpoint `/commits/{sha}/status` shows `state == "success"` **and** description `Review completed`:

    ```bash
    gh api "repos/$REPO/commits/$HEAD_SHA/status" \
      --jq '.statuses[] | select(.context == "CodeRabbit" and .state == "success") | .description'
    ```

2. A `Merge Risk:` line exists whose `` up to `<prefix>` `` matches the current `headRefOid`.

Anything else is a gap, not a pass. `Review rate limited` means the review never started; a card carrying it does not leave `sync`. (Endpoint choice + why branch protection is not the lever: `kanban-dispatch-detail.md` § CodeRabbit endpoint measurement.)

## Review lens selection

`review` is not one thing. The lead picks the lenses from what the card actually changed and states the choice in the `sync` dispatch, so the gate runs that review rather than re-deriving it. Lens table: `kanban-dispatch-detail.md` § Review lens selection.

`--deep --patch` is opt-in twice over: `--patch` drafts a fix and is absent unless the operator asked for it. Do not add it on the lead's own initiative.

## The `/clear` handoff between phases

[HARD] A companion session does not carry one card's context into the next card. When a phase completes and the lead has read its evidence, the lead **asks the operator to `/clear` that session** — `/clear` is a user-typed command and cannot be sent as an instruction. The lead's message states, in order: what closed (card, phase, evidence read), which session to `/clear` (by name), and what happens next (the next column, and which session is instructed once the clear is done).

Where the next phase reuses a just-cleared session, the lead re-sends the full pointer instruction rather than assuming the session remembers.

The lead's own session is cleared the same way, between cards rather than phases: once a card reaches `done`, the operator is asked to `/clear` the lead session, and the next turn presents the queue again.

## Isolation is entered, never provisioned

[HARD] A card's work happens inside a worktree, and that worktree is **entered through the launcher** — never created with a bare `git worktree add`.

| Need | Form |
|---|---|
| Work inside the worktree in this session | `moai cc -w <name>` |
| Open it in a new window, keeping this session | `moai cc -w <name> --spawn` |
| Re-enter one from the current session | `EnterWorktree(<path>)` |
| Leave it | `ExitWorktree` |
| Dispose it once the card's work has merged on the remote | L2 tree (`~/.moai/worktrees/…`) only: `moai worktree done`. An L1 tree (`.claude/worktrees/…`) is disposed via the session-end keep/remove prompt — `moai worktree` never registers it |

`moai worktree` deliberately carries no creation verb — entering is the launcher's job. A tree made with a raw `git worktree add` is one git knows about but MoAI does not: `done`, `clean`, and `recover` have nothing to close, and orphans accumulate until reconciled by hand.

[HARD] **`moai worktree done` closes L2 trees only.** A worktree entered by short name (`moai cc -w <name>` → `.claude/worktrees/<name>/`) is L1 and is never in `moai worktree`'s registry — `done` on it is a category error, not a disposal. L1 disposal is the session-end keep/remove prompt, or `git worktree unlock` + `git worktree remove` once the session is done. The full L1/L2 boundary lives in `worktree-integration.md` § Terminology Glossary.

[HARD] **The card's branch is unpushed, so its worktree is the work's only instance.** Dispose of no worktree — L1 or L2 — until the lead has integrated the branch and the remote merge has landed; disposal before that destroys the only copy.

[HARD] **A new card starts in a new worktree — exit any previous one first.** `EnterWorktree(<card-id>)` cannot run from inside a worktree session: a lane still anchored in the previous card's tree MUST `ExitWorktree` back to the primary checkout before creating the new one, or it does the new card's work on the old card's branch. The fresh tree is created from the remote default branch rather than reused — reuse without a `/clear` in between carries the old card's context and untracked artifacts into the new card. Where the new card depends on a prior card's unmerged code, merge the prior branch inside the new worktree; a dependency is a reason to merge, never to reuse the tree.

[HARD] **Card worktree branches carry the `WT-` prefix and a descriptive slug — never the card id.** `EnterWorktree(<name>)` auto-names its branch `worktree-<name>`, which is unwieldy and invisible to the worktree lifecycle tooling. Immediately after creating a card worktree, rename in place with `git branch -m WT-<slug>`: safe inside a worktree (tree, lock, anchoring unaffected); `moai cc -w <name>` re-entry resolves by tree name, not branch name, and the rename switches the disposal path (`worktree-integration.md` § Terminology Glossary).

The slug is derived from what the card **does**, so a reader of `git branch` or a pull-request list learns the change without a lookup — `WT-t0` says nothing, `WT-branch-naming` says what landed. Its shape:

| Property | Rule |
|---|---|
| Source | The card's title, not its id |
| Tokens | At most 3, hyphen-separated |
| Length | At most 24 characters (the slug alone; `WT-` brings the branch to at most 27) |
| Alphabet | Lowercase `a-z`, `0-9`, and `-` |
| Card id | MUST NOT appear — not as a prefix, a suffix, or a token |

The **worktree directory keeps the card id** (`.claude/worktrees/<card-id>`). Only the branch takes the slug; the tree path is what the disposal tooling and the evidence path key on.

[HARD] **Dropping the id from the branch moves traceability onto three other carriers, and all three are mandatory.** The branch name no longer answers "which card was this?", so nothing may rely on reading it back:

- The dispatch's `card:` field carries the card id — it is the address, and it is never omitted.
- Every commit on the branch names the card id in its message, so `git log` recovers the card without the branch name.
- The evidence path keeps the card id (`.moai/reports/<card-id>/verdict.md`).

A lane that reports a branch name without also reporting its card id has not reported the card. Merges reference the `WT-` name; the lead maps it back through the `card:` field it dispatched.

The lead dispatches this rather than assuming it: each instruction names the worktree and says to drive it with `git -C <path>` rather than `cd` — a `cd` inside a compound command lasts for that invocation only, so the next command silently reads the wrong tree. A companion reporting it worked in the shared checkout is a fault to report, not a detail to tidy up (rationale: `kanban-dispatch-detail.md` § Isolation rationale).

## Verification load is lane-local

[HARD] **Lane-local verification is scoped to the card.** A lane runs the tests its own change can affect, then pushes and lets CI run the full suite — the better evidence: the full suite, in a clean environment, against the actual pull-request head. A full-suite run on a loaded developer machine measures the machine, not the code. (Incident record — load 413, orphaned spin loops, the contention↔flakiness loop: `kanban-dispatch-detail.md` § Verification load incident record.)

[HARD] **Never spawn background load.** Where a verification genuinely needs contention, the load must be cleanup-guaranteed — kills registered with the test framework's cleanup hook, or a `timeout` wrapper that bounds the process from outside. A trailing `kill` is not cleanup; it is a line the process may never reach, and every path that ends early leaves the load running.

**A verification recipe that spawns processes is itself a hazard, and gets reviewed as one.** The fault belongs to the dispatcher who wrote and approved the recipe, not to the lane that ran it as given.

### The env-isolated verification form

[HARD] Inside a worktree, an environment-scrubbed verification runs as one compound `unset … && <command>` invocation:

```bash
unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test ./...
```

The `env -u VAR <command>` form is rejected — the Claude Code guard rejects shell structures it cannot statically track, `env` among them; `unset … && <command>` stays visible to static analysis. Two load-bearing properties, easy to "simplify" away: **one invocation** (each Bash call is a fresh process — an `unset` issued as its own call does not carry into the next command; the scrub and the command travel together or the scrub does nothing) and **no subshell** (`( unset …; <command> )` is rejected the same way).

Moving the command into a script file is not a workaround — the guard cannot read inside a script, so every check is bypassed for that payload. Where a verification cannot be expressed as one compound invocation, reduce the verification rather than route it around the guard.

## Integration into the release branch is self-served

[HARD] A lane whose card has passed verification does not wait for the lead to integrate it: the lane merges its own branch into the batch's release branch (`release/vX.Y.Z`) itself. The lead provisions the release branch and its worktree at batch start; from then on, every integration is lane work.

Two measured constraints — git checks one branch out in exactly one worktree, and the worktree-session guard refuses a cross-tree `git -C` — make the lane enter the release worktree rather than drive it remotely.

- **One integration surface.** The release branch lives in exactly one worktree — the one the lead provisioned; a lane never checks it out in its own tree.
- **Enter, do not redirect.** The lane switches into the release worktree with `EnterWorktree(<release-worktree-path>)` and runs the merge there as a plain `git merge --no-ff <WT-branch>`. A cross-tree `git -C <release-worktree> merge` from inside the lane's own worktree is refused by the worktree-session guard; entering is the sanctioned path.
- **Return the same way.** `ExitWorktree` returns the session to the primary checkout, not to the lane's own worktree — after the merge, the lane re-enters its card worktree with `EnterWorktree(<own-path>)` before continuing card work.
- **One integrating session at a time.** The release worktree is the serialization point. On entry, confirm no merge is in progress — `git rev-parse -q --verify MERGE_HEAD` must print nothing; a lane that finds a merge in progress exits, waits, and retries.
- **Conflicts belong to the lane that owns the change.** The integrating lane resolves the conflicts its own merge raises. A conflict it cannot resolve — a semantic clash with another lane's already-merged change — is a blocker report to the lead, not a forced merge.
- **Push the release branch; the batch pull request stays with the lead.** After its merge, the lane pushes the release branch (`git push origin release/vX.Y.Z`). A rejected push means another lane pushed first — fetch, integrate the fetched release branch, and push again; never force. Until the pushed release branch's batch PR merges, the disposal rule above still binds.

The completion signal is the branch name, merge SHA, and evidence path.

## Factory Mode — the card travels whole

`moai cc -f <N>` launches one lead plus lane sessions `lane-1..lane-N` ("lane" is the user-facing term). No per-column companions: the lead routes each card WHOLE to a free lane, which carries it `plan → run → sync` in-session — serial stages, each stage's execution spawned as sub-agents — and owns it end to end. A/B/C collapse into the lane (the class still names which ceremonies are skipped — `plan` for A and B — but no card changes sessions). Queue, evidence-reading, integration, and disposal rules are unchanged. Mechanics: `kanban-dispatch-detail.md` § Factory in-lane 3-stage.

## Boundaries — what this protocol does not do

- **No board state store.** The queue is a plain file; column position is held by the lead within a card's run and re-derived from SPEC status after a clear. Persistent board state, per-card worktree lifecycle, WIP limits, and card/frontmatter reconciliation are separate work, not assumed here.
- **No session spawning.** The lead addresses sessions the operator launched — it never creates one (sub-agents are not sessions).
- **No gate bypass.** Kickoff approval before run-phase entry, and every other approval gate, is unchanged by being inside a dispatch cycle.
- **No question delegation.** Companion sessions return blocker reports; the operator is asked by the lead, through `AskUserQuestion`.
- **A role with no live session is a fault, not a wait.** The lead reports the empty role and the waiting card; silently holding it presents as a hang, the most expensive failure shape to diagnose. Empty means a complete listing showed none; an incomplete one is a gap, not an empty role (detail companion).

## Cross-references

- `.claude/rules/moai/core/askuser-protocol.md` — the question channel the lead uses for card selection and `/clear` prompts
- `.claude/rules/moai/core/verification-claim-integrity.md` — why completion is read rather than trusted
- `.claude/rules/moai/core/agent-common-protocol.md` § Blocker Report Format — what a companion returns when it cannot proceed
- `.claude/rules/moai/workflow/worktree-integration.md` — the L1/L2 worktree tiers, their lifetimes, and the disposal contract
- `.claude/skills/moai/workflows/todo.md` — the backlog queue surface
- `.claude/agents/moai/manager-lead.md` — the coordination agent the lead session works through

---

Classification: Evolvable operational rule — applies to the lead session of Kanban Mode. Detail companion: `kanban-dispatch-detail.md` (stub + lazy-companion split).
