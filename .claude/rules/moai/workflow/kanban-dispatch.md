# Kanban Dispatch Protocol

How the **lead** session of Kanban Mode moves a card across the board: what admits work, who is told to do it, how completion is judged, and when the operator is asked to `/clear`.

> **Loading scope**: Intentionally always-loaded. A session learns it is the kanban lead from the SessionStart context, not from a file path, so a `paths:`-restricted rule would never reach the session that needs it.

> **Detail companion**: `kanban-dispatch-detail.md` carries this rule's long tables (board columns, card classes, review lenses), the dispatch-cycle walkthrough, incident narratives, and worked rationale. This stub keeps every [HARD] rule, every prohibition, and every cross-reference pointer. Load the companion when moving a card between columns, classifying a card, or choosing review lenses.

## Scope — when this rule is live

This rule binds a session whose SessionStart context declares **Kanban Mode** with the `lead` role. In every other session it is inert: a companion session (`plan` / `run` / `review` / `sync`) receives instructions, it does not dispatch them, and a session outside Kanban Mode has no board to move.

Kanban Mode is entered with `moai cc -k` (or `moai glm -k`), which elects one lead and prints one launch command per companion role. Companion sessions are launched **by hand, one per terminal** — a session cannot launch another session, and no mechanism to spawn a peer exists or is wanted.

One boundary on the mode itself: nudge delivery rides on cross-session messaging, which does not exist on native Windows and is off under some providers, versions, and flag settings — see `cross-session-messaging.md` § Availability constraints. An absent channel fails quietly, so the lead surfaces it to the operator; the queue itself keeps working without it.

## The board

Six columns, fixed and ordered: `backlog → plan → run → review → sync → done`. `backlog` and `done` have no owning session; the four columns between them each map to exactly one companion role (`plan` / `run` / `review` / `sync`), which is what makes dispatch a lookup rather than a decision. Column-by-role table: `kanban-dispatch-detail.md` § The board.

## Entry into the board is an operator act

`backlog` has no owning session, so no completion report exists for the lead to react to, and a lead that admitted cards on its own initiative would be **generating** work rather than scheduling it. Every card's origin is therefore the operator's request.

[HARD] **The lead is the queue's sole producer.** The operator asks; the lead turns the request into a card with `moai todo add "<description>"` (`moai todo` alone lists the queue). Production is the one queue mutation the lead performs on its own authority — and it is translation, not invention: nothing enters the queue the operator did not ask for.

[HARD] **Promotion is the operator's act, always.** After a `/clear`, the lead presents the queued cards through `AskUserQuestion` and the operator picks; only then does the lead dispatch to `plan`. The lead never picks for the operator, never reorders by inferred priority, and never silently promotes a backlog item. An empty queue is a state to report, not a prompt to invent work.

## Card classes — not every card needs four columns

The lead classifies each card as it leaves `backlog` and names the class in the dispatch: **A — direct close** (one file, one line, no design judgement, CI catches the regression; one session carries the card to a pull request, `plan` and `review` skipped), **B — defect, cause unknown** (`run → review → sync`; `plan` is skipped, so no SPEC exists), **C — design change** (a decision, or spans subsystems; all four columns). Full table and the ceremony-cost rationale: `kanban-dispatch-detail.md` § Card classes.

[HARD] **Class A is admitted on checked evidence, not on an assertion.** Two of its three properties are mechanically checkable, so they are checked and the check is cited: the diff is measured (`git diff --stat` against the base, showing the one file) and CI is observed green **on the head that will merge**. The third — no design judgement in it — is a judgement rather than a measurement, so it is stated in the dispatch where the operator can disagree with it. A card that cannot cite both measurements is not Class A, and it takes the `review` column. (Same shape as the CodeRabbit section below: a class that skips review on a claim nobody checked is the unobserved-claim hazard.)

The justification is never "it is faster". Speed is the effect of skipping the columns, not the reason for it, and a card justified by speed alone is a Class C card being rushed.

**Class B skips `plan`, not `review`** — an unestablished cause is precisely what review catches, so it is the last column to drop. The `run` session owns both the investigation and the fix. Before the card leaves `run`, the evidence that established the cause — the command that reproduced the defect and what it printed — is written into the card's progress record, and the run session names that path in its completion report so the lead reads the cause rather than taking it on trust.

**Work in progress: one card per column, and only when each card occupies a different worktree.** Two cards sharing one worktree run serially whatever columns they sit in, because they share a working tree and a branch. (Where the parallelism comes from per class, and why research fan-out during `plan` is Class-C-only: `kanban-dispatch-detail.md` § Card classes.)

## The dispatch cycle

`[operator picks a card] → plan → run → review → sync → [lead marks done]` — each arrow is one dispatch from the lead to one companion session. Dispatch is addressed by session name via `SendMessage({to: "<name>", message: "…"})`; `ListAgents` reports the live set. Each instruction carries, at minimum: the card, the SPEC ID once one exists, the phase command to run, and the completion signal to write. Keep it a pointer, not a copy — the companion reads the SPEC artifacts itself rather than receiving them inline.

**`sync → done` is the same act with the dispatch removed.** No session occupies `done`, so the lead reads the sync session's completion evidence and records the terminal transition itself. Session-naming rules and the address-block walkthrough: `kanban-dispatch-detail.md` § The dispatch cycle.

### The delegation channel is the queue

[HARD] Work is delegated through the queue on disk, not through messages. The queue file resolves against the primary checkout from every linked worktree — one repository, one queue — so a card admitted from anywhere is visible everywhere, and that single-file visibility is what makes the queue a channel rather than a shared opinion.

A cross-session message is a nudge, never the delegation itself: delivery is not guaranteed, and a delivered message consumes the recipient's quota like a typed prompt. No dispatch may depend on a message arriving. The disk record — card admitted, picked, done — is what moves the board; a message that goes unanswered changes nothing.

### Dispatch language

[HARD] A dispatch is written in the operator's `conversation_language`. The operator watches it scroll past, which makes it user-facing output rather than internal agent traffic. The carve-out is narrow, and the boundary is **who reads it** — an `Agent()` subagent prompt reaches no human and stays English. (Why this classifies the dispatch under the language rules rather than exempting it: `kanban-dispatch-detail.md` § Dispatch language.)

What stays verbatim in every language: SPEC IDs, command names and their flags, file paths, session names, and technical identifiers. Those are addresses rather than prose, and a translated address does not resolve.

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

- `card`, `cmd`, `wt`, and `evidence` are always present. `spec` joins once a SPEC exists; a Class B card, which skips `plan`, carries none, and its `evidence` names whatever record the lead will read instead. `lens` appears only in a `review` dispatch — the lenses from the Review lens selection table, where the choice itself is the reason, stated as an address instead of a sentence.
- **No explanatory prose.** Procedure, background, and justification live in the card text and the SPEC artifacts this block points at; a dispatch that restates them makes the operator read the same thing twice. If something does not fit a field, it belongs in the card, not around the block.
- **Ceiling: the block is at most 10 lines.** A dispatch that does not fit is trying to be a handoff; move the payload into the card and send the block.

## Completion is read, never trusted

[HARD] The lead advances a card on **evidence it read**, not on a companion's reply. Reply routing between sessions is not guaranteed to arrive, and a reply is a claim rather than an observation.

Before moving a card out of a working column, the lead reads the card's `progress.md` (and, where the phase declares one, the verification evidence path it cites) and confirms the phase actually closed. A missing, unreadable, or stale evidence file is a **gap** — the card stays where it is and the lead reports why. Absence of a failure signal is not a pass.

This applies equally to the operator: when the lead reports a column advanced, it names what it read.

**The final PASS/FAIL verdict is the lead's**, read from the evidence on disk and never delegated to the lane that produced the work — the executor judging its own output is the failure shape this section exists to prevent. The division is structural, not ceremonial: where the board's lanes run on a different backend than the lead, the lane sessions cannot commission judgment work onto the lead's backend, so the verdict has a home in the lead even when the execution has none.

### CodeRabbit is not read from `gh pr checks`

[HARD] A `gh pr checks` row naming CodeRabbit is not evidence that a review happened. CodeRabbit reports through a commit status whose `state` is `success` **even when no review ran**, so the row prints `pass` in both cases and is byte-identical between them. The only field that separates a reviewed pull request from an unreviewed one is the status description.

A CodeRabbit row counts as evidence only when BOTH of these hold:

1. The status is `success` **and** its description reads `Review completed`:

    ```bash
    gh api "repos/$REPO/commits/$HEAD_SHA/status" \
      --jq '.statuses[] | select(.context == "CodeRabbit" and .state == "success") | .description'
    ```

    Both halves are required, and neither is sufficient alone — without the state filter, a `failure`/`error` status carrying a `Review completed` description would read as a pass. The predicate assumes the **combined** status endpoint `/commits/{sha}/status` (most recent status per context); do NOT substitute the plural `/commits/{sha}/statuses` (full history newest-first — a positional pick selects the wrong entry). Measurement narrative: `kanban-dispatch-detail.md` § CodeRabbit endpoint measurement.

1. A `Merge Risk:` line exists whose `` up to `<prefix>` `` matches the current `headRefOid`, so the verdict covers the head being merged rather than an earlier commit.

Anything else is a gap, not a pass. `Review rate limited` in particular means the review never started, and a card carrying it does not leave `review` or `sync`. Branch protection is not the lever here — the status state is `success` in precisely the failing case; only a read of the description surfaces it.

## Review lens selection

`review` is not one thing. The lead picks the lenses from what the card actually changed, and states the reason in the dispatch so the review session does not re-derive it. Lens table (which lenses per touched surface): `kanban-dispatch-detail.md` § Review lens selection.

`--deep --patch` is opt-in twice over: `--patch` drafts a fix and is absent unless the operator asked for it. Do not add it on the lead's own initiative.

## The `/clear` handoff between phases

[HARD] A companion session does not carry one card's context into the next card. When a phase completes and the lead has read its evidence, the lead **asks the operator to `/clear` that session** — `/clear` is a user-typed command and cannot be sent as an instruction. The lead's message states, in order: what closed (card, phase, evidence read), which session to `/clear` (by name), and what happens next (the column the card moves to, and which session is instructed once the clear is done).

Where the next phase reuses a just-cleared session, the lead re-sends the full pointer instruction rather than assuming the session remembers the card.

The lead's own session is cleared the same way, between cards rather than phases: once a card reaches `done`, the operator is asked to `/clear` the lead session, and the next turn begins by presenting the queue again.

## Isolation is entered, never provisioned

[HARD] A card's work happens inside a worktree, and that worktree is **entered through the launcher** — never created with a bare `git worktree add`.

| Need | Form |
|---|---|
| Work inside the worktree in this session | `moai cc -w <name>` |
| Open it in a new window, keeping this session | `moai cc -w <name> --spawn` |
| Re-enter one from the current session | `EnterWorktree(<path>)` |
| Leave it | `ExitWorktree` |
| Dispose it once the card's work has merged on the remote | L2 tree (`~/.moai/worktrees/…`) only: `moai worktree done`. An L1 tree (`.claude/worktrees/…`) is disposed via the session-end keep/remove prompt — `moai worktree` never registers it |

`moai worktree` deliberately carries no creation verb — its own help states that entering is the launcher's job. A tree made with a raw `git worktree add` is one git knows about but MoAI does not, so `done`, `clean`, and `recover` have nothing to close and orphans accumulate until reconciled by hand.

[HARD] **`moai worktree done` closes L2 trees only.** A worktree entered by short name (`moai cc -w <name>` → `.claude/worktrees/<name>/`) is L1 and is never in `moai worktree`'s registry — `done` on it is a category error, not a disposal. L1 disposal is the session-end keep/remove prompt, or `git worktree unlock` + `git worktree remove` once the session is done. The full L1/L2 boundary lives in `worktree-integration.md` § Terminology Glossary.

[HARD] **The card's branch is unpushed, so its worktree is the work's only instance.** Dispose of no worktree — L1 or L2 — until the lead has integrated the branch and the remote merge has landed; disposal before that destroys the only copy.

[HARD] **A new card starts in a new worktree.** A companion session taking a new card creates a fresh worktree with `EnterWorktree(<card-id>)` (based on the remote default branch) rather than reusing the previous card's tree — reuse without a `/clear` in between carries the previous card's context and untracked artifacts into the new card. Where the new card depends on a prior card's unmerged code, the dependency is pulled in inside the new worktree with `git merge <prior-branch>`; a dependency is a reason to merge, never a reason to reuse the tree.

[HARD] **Card worktree branches carry the `WT-` prefix.** `EnterWorktree(<name>)` auto-names its branch `worktree-<name>` — unwieldy for long names, invisible to the worktree lifecycle tooling. Immediately after creating a card worktree, rename in place with `git branch -m WT-<name>`: safe inside a worktree (tree, lock, session anchoring unaffected), and `moai cc -w <name>` re-entry resolves by tree name, not branch name. The rename also switches the eventual disposal path (`worktree-integration.md` § Terminology Glossary). Dispatches, merges, and completion evidence reference the `WT-` name.

The lead dispatches this rather than assuming it: each instruction names the worktree and says to drive it with `git -C <path>` rather than `cd` — a `cd` inside a compound command lasts for that invocation only, so the next command silently reads the wrong tree. Why the shared checkout is the wrong place for a card (concurrent readers, card-outlives-phase): `kanban-dispatch-detail.md` § Isolation rationale. Where a companion reports having worked in the shared checkout instead, that is a fault to report, not a detail to tidy up afterwards.

## Verification load is lane-local

[HARD] **Lane-local verification is scoped to the card.** A lane runs the tests its own change can affect, then pushes and lets CI run the full suite. CI is not the fallback here; it is the better evidence — it runs the full suite in a clean environment against the actual pull-request head. A full-suite run on a loaded developer machine measures the machine, not the code. (Incident record — load 413, orphaned spin loops, the contention↔flakiness loop: `kanban-dispatch-detail.md` § Verification load incident record.)

[HARD] **Never spawn background load.** Where a verification genuinely needs contention, the load must be cleanup-guaranteed — kills registered with the test framework's cleanup hook, or a `timeout` wrapper that bounds the process from outside. A trailing `kill` is not cleanup; it is a line the process may never reach, and every path that ends early leaves the load running.

**A verification recipe that spawns processes is itself a hazard, and gets reviewed as one.** The fault belongs to the dispatcher who wrote and approved the recipe, not to the lane that ran it as given — a lane executing an approved recipe is doing what it was told, and a rule that blames the executor teaches the wrong actor to be careful.

### The env-isolated verification form

[HARD] Inside a worktree, an environment-scrubbed verification runs as one compound `unset … && <command>` invocation:

```bash
unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test ./...
```

The `env -u VAR <command>` form is rejected. The guard doing the rejecting lives in the Claude Code binary rather than in MoAI, and this cannot be turned off: it rejects shell structures it cannot statically track, and `env` is one. The `unset … && <command>` form keeps the command visible to static analysis, which is why it is the prescribed form.

Two properties of the form are load-bearing, and both are easy to "simplify" away:

- **One invocation.** Each Bash call is a fresh process, so an `unset` issued as its own call does not carry into the next command. The scrub and the command travel together or the scrub does nothing.
- **No subshell.** Wrapping it as `( unset …; <command> )` is another structure the guard cannot statically track, and is rejected the same way.

Moving the command into a script file is not a workaround — the guard cannot read inside a script, so the entire check set is bypassed for that payload. Reading the script first does not restore them; a human read is not the guard running. Where a verification cannot be expressed as one compound invocation, that is a signal to reduce the verification, not to route it around the guard.

## Integration into the release branch is self-served

[HARD] A lane whose card has passed verification does not wait for the lead to integrate it: the lane merges its own branch into the batch's release branch (`release/vX.Y.Z`) itself. The lead provisions the release branch and its worktree at batch start; from then on, every integration is lane work.

The mechanism respects two measured constraints — git checks one branch out in exactly one worktree, and the worktree-session guard refuses a cross-tree `git -C` — so the lane enters the release worktree rather than driving it remotely:

- **One integration surface.** The release branch lives in exactly one worktree — the one the lead provisioned. A lane never checks the release branch out in its own tree.
- **Enter, do not redirect.** The lane switches into the release worktree with `EnterWorktree(<release-worktree-path>)` and runs the merge there as a plain `git merge --no-ff <WT-branch>`. A cross-tree `git -C <release-worktree> merge` from inside the lane's own worktree is refused by the worktree-session guard; entering is the sanctioned path.
- **Return the same way.** `ExitWorktree` returns the session to the primary checkout, not to the lane's own worktree — after the merge, the lane re-enters its card worktree with `EnterWorktree(<own-path>)` before continuing card work.
- **One integrating session at a time.** The release worktree is the serialization point. On entry, confirm no merge is in progress — `git rev-parse -q --verify MERGE_HEAD` must print nothing; a lane that finds a merge in progress exits, waits, and retries.
- **Conflicts belong to the lane that owns the change.** The integrating lane resolves the conflicts its own merge raises. A conflict it cannot resolve — a semantic clash with another lane's already-merged change — is a blocker report to the lead, not a forced merge.
- **Push the release branch; the batch pull request stays with the lead.** After its merge, the lane pushes the release branch (`git push origin release/vX.Y.Z`). A rejected push means another lane pushed first — fetch, integrate the fetched release branch, and push again; never force. Until the pushed release branch's batch PR merges, the disposal rule above still binds.

The completion signal is the branch name, merge SHA, and evidence path.

## Boundaries — what this protocol does not do

- **No board state store.** The queue is a plain file; column position is held by the lead within a card's run and re-derived from SPEC status after a clear. Persistent six-column state, per-card worktree lifecycle, WIP limits, and card/frontmatter consistency reconciliation are separate work and are not assumed here.
- **No spawning.** The lead addresses sessions the operator launched. It never creates one.
- **No gate bypass.** Kickoff approval before run-phase entry, and every other approval gate, is unchanged by being inside a dispatch cycle.
- **No question delegation.** Companion sessions return blocker reports; the operator is asked by the lead, through `AskUserQuestion`.
- **A role with no live session is a fault, not a wait.** The lead reports the empty role and the waiting card; silently holding it presents as a hang, the most expensive failure shape to diagnose.

## Cross-references

- `.claude/rules/moai/core/askuser-protocol.md` — the question channel the lead uses for card selection and `/clear` prompts
- `.claude/rules/moai/core/verification-claim-integrity.md` — why completion is read rather than trusted
- `.claude/rules/moai/core/agent-common-protocol.md` § Blocker Report Format — what a companion returns when it cannot proceed
- `.claude/rules/moai/workflow/worktree-integration.md` — the L1/L2 worktree tiers, their lifetimes, and the disposal contract
- `.claude/skills/moai/workflows/todo.md` — the backlog queue surface
- `.claude/agents/moai/manager-kanban.md` — the coordination agent, including its kanban-lead role

---

Classification: Evolvable operational rule — applies to the lead session of Kanban Mode. Detail companion: `kanban-dispatch-detail.md` (stub + lazy-companion split).
