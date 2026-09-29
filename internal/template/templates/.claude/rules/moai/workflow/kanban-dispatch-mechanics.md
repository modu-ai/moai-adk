---
description: Board-and-lens mechanics companion to the kanban dispatch protocol
paths: "**/kanban-dispatch*.md,**/.claude/agents/moai/manager-lead.md,**/.moai/state/integration/**"
---

# Kanban Dispatch — Board and Lens Mechanics

> Owns: the board mechanics, dispatch-cycle walkthrough, review-lens pointer, heavy-run lease, Factory Mode mechanics, the remaining boundaries and cross-references — and, relocated from the stub by the always-loaded diet, § Isolation (launcher table, worktree tiers, `WT-` branch naming, the traceability carriers, the worktree-guard refusal shapes), § Verification load is lane-local (the env-isolated variants and measurements), and § Integration into the release branch is self-served (the full window procedure). `kanban-dispatch.md` is the SSOT and keeps every binding clause as a one-liner; sibling companion `kanban-dispatch-detail.md` keeps the long tables and narratives.

## The board

Five columns, fixed and ordered: `backlog → plan → run → sync → done`. `backlog` and `done` have no owning session; the three working columns each map to exactly one companion role (`plan` / `run` / `sync`), which is what makes dispatch a lookup rather than a decision. There is no `review` column — the sync gate absorbs the review verdict and runs the lenses itself. Column-by-role table: `kanban-dispatch-detail.md` § The board.

## Review lens selection

`review` is not one thing. The factory leader picks the lenses from what the card actually changed and states the choice in the `sync` dispatch, so the gate runs that review rather than re-deriving it. Lens table: `kanban-dispatch-detail.md` § Review lens selection.

`--deep --patch` is opt-in twice over: `--patch` drafts a fix and is absent unless the operator asked for it. Do not add it on the leader's own initiative.

### Serializing a heavy run across lanes

Where two lanes would otherwise start the same heavy verification at once, a lane takes a lease on a named resource first: `moai slot acquire --resource <name> --max-duration <bound>`, and `moai slot release --resource <name>` when the run ends. `moai slot status` says who holds it. This is the same record-and-liveness shape as the integration window, kept deliberately separate from it — its own record, its own lock, its own config key — because a merge window and a heavy-run window are different resources and one window for both makes each wait on the other.

The bound is the holder's own declaration: past it another lane may take the resource over without `--force`, so a forgotten release costs the bound and not the batch. A PreToolUse guard that refuses a matching command while another live session holds the resource exists, and is **opt-in and off by default** (`workflow.slot_lease.enabled`); the verbs work whatever that flag says. Full surface: `.claude/rules/moai/workflow/resource-slot-lease.md`.

## The dispatch cycle walkthrough

`[operator picks a card] → plan → run → sync → [leader marks done]` — each arrow is one dispatch from the leader to one companion session, addressed by name (`SendMessage`; `ListAgents` lists live sessions and says when it could not check them all). Each instruction carries, at minimum: the card, the SPEC ID once one exists, the phase command, and the completion signal to write. Keep it a pointer, not a copy — the companion reads the SPEC artifacts itself.

**`sync → done` is the same act with the dispatch removed.** No session occupies `done`, so the leader reads the sync session's completion evidence and records the terminal transition itself. Session-naming rules and the address-block walkthrough: `kanban-dispatch-detail.md` § The dispatch cycle.

## Factory Mode mechanics

`moai cc -f <N>` launches one leader plus lane sessions labelled `lane-1..lane-N` (`lane-<n>` is the session label, and `-f lane` joins as the next free one). No per-column companions: the leader routes each card WHOLE to a free lane, or a lane receives the card by its own self-promotion of an already-queued card; either way the lane carries it `plan → run → sync` in-session — serial stages, each stage's execution spawned as sub-agents — and owns it end to end. A/B/C collapse into the lane (the class still names which ceremonies are skipped — `plan` for A and B — but no card changes sessions). Queue, evidence-reading, integration, and disposal rules are unchanged. Mechanics: `kanban-dispatch-detail.md` § Factory in-lane 3-stage.

## Isolation — launcher, tiers, branch naming, and the traceability carriers

Relocated from the stub (always-loaded diet): the stub keeps each [HARD] clause as a one-liner; this section owns the tables and the measured detail. Trigger affinity: the launcher/tier/naming decisions happen when provisioning or disposing a card worktree — the `manager-lead.md` and `kanban-dispatch*` surfaces this file is scoped to.

The launcher table (per-harness entry forms):

| Need | Form |
|---|---|
| Create a harness-neutral L1 worktree | `moai worktree new <name>` |
| Start Claude Code in a MoAI tree | `moai cc -w <absolute-worktree-path>` (`--spawn` for a new window; Claude may ask to approve an external worktree path) |
| Start a Claude-native worktree | `moai cc -w <name>` creates or enters `.claude/worktrees/<name>` |
| Re-enter or leave in the current Claude Code session | `EnterWorktree(<path>)` / `ExitWorktree` |
| Start a Codex app chat in a new tree | Select Worktree and the starting branch in the new chat |
| Start a Codex CLI session in an existing tree | `moai codex -w <name-or-absolute-path>`; the flag never creates a tree |
| Work in a tree from the current Codex session | `git -C <absolute-worktree-path>` and direct file operations there |
| Dispose it once the card's work has merged on the remote | L2 tree (`~/.moai/worktrees/…`) only: `moai worktree done`. Project L1 trees under `.claude/worktrees/` or `.moai/worktrees/` are removed only after their sessions end |

`moai worktree new <name>` creates a harness-neutral L1 tree at `.moai/worktrees/<name>` without entering it. A Codex factory lane uses it for each card, then launches the interactive card session with `codex -C <absolute-worktree-path>`; it does not use `moai codex -w` to create the tree. The Codex app's Worktree selector creates a Codex-managed tree separately. A raw `git worktree add` bypasses MoAI's name validation, base selection, and post-create Git configuration.

[HARD] **`moai worktree done` closes L2 trees only.** `done` refuses both project roots. Remove an L1 tree only after its session ends and its branch is integrated, using the session-end prompt where available or `git worktree unlock` + `git worktree remove`.

The slug says what the card **does**, so a reader of `git branch` or a pull-request list learns the change without a lookup — `WT-t0` says nothing, `WT-branch-naming` says what landed. Its shape:

| Property | Rule |
|---|---|
| Source | The card's title, not its id |
| Tokens | At most 3, hyphen-separated |
| Length | At most 24 characters (the slug alone; `WT-` brings the branch to at most 27) |
| Alphabet | Lowercase `a-z`, `0-9`, and `-` |
| Card id | MUST NOT appear — not as a prefix, a suffix, or a token |

The **worktree directory keeps the card id** (`.moai/worktrees/<card-id>` for new MoAI trees; `.claude/worktrees/<card-id>` for existing Claude-native trees) — only the branch takes the slug, and the tree path is what the disposal tooling and the evidence path key on.

[HARD] **Dropping the id from the branch moves traceability onto three other carriers, and all three are mandatory.** The branch name no longer answers "which card was this?", so nothing may rely on reading it back:

- The dispatch's `card:` field carries the card id — it is the address, and it is never omitted.
- Every commit on the branch names the card id in its message, so `git log` recovers the card without the branch name.
- The evidence path keeps the card id (`.moai/reports/<card-id>/verdict.md`).

A lane reporting a branch name without its card id has not reported the card. Merges reference the `WT-` name; the lead maps it back through the dispatched `card:` field.

[HARD] **A card-delivering pull request's PR title MUST carry the delivering card id** — the branch name is read by a human scanning `git branch` and wants a slug; the PR title is read by a machine and wants the id. Traceability rests on **four** carriers — the dispatch `card:` field, the commit message, the evidence path, and the PR title. It binds card-delivering pull requests only, and only those opened after it landed; nothing is retitled. Rationale and carrier measurements: `kanban-dispatch-detail.md` § The PR-title carrier.

The lead dispatches this rather than assuming it: each instruction names the worktree and says to drive it with `git -C <path>` rather than `cd` — a `cd` inside a compound command lasts for that invocation only, so the next command silently reads the wrong tree. A companion reporting it worked in the shared checkout is a fault to report (rationale: `kanban-dispatch-detail.md` § Isolation rationale).

[HARD] **Inside a worktree session that `<path>` is the worktree's own absolute path**, and the dispatch writes it that way. Measured on Claude Code 2.1.275: the guard refuses `-C .`, a relative path, a path computed at runtime, and a path outside this worktree — three distinct refusal messages, none of them a runtime defect. Plain git (pipes and `&&` chains included), `git -C <own absolute path>`, and `--git-dir=<own .git>` pass; so does `cd <own worktree> && git …`, which the rule above still advises against for the reason it gives. The refusal is git-scoped: a command carrying no git passes with substitution, loops, redirects, or a heredoc body naming a git command.

## Verification load is lane-local

Relocated from the stub: the variants and the measured refusal shapes behind the stub's compound-invocation rule. Trigger affinity: consulted while composing or reviewing a lane verification recipe — the lead/lane work surfaces this file is scoped to.

`env -u VAR <command>` scrubs identically and is **not** refused; it is excluded to keep one recipe across lanes, which is a convention rather than a runtime constraint. Do not rewrite the standard form into it, and do not cite a guard as the reason — the reason is uniformity.

**The subshell is the form that actually breaks, and only sometimes.** `( unset …; <command> )` runs when the command carries no git, and is refused when it does, because the worktree guard cannot statically verify a git call inside one. Since a verification recipe may acquire a git step later, the compound form above is the one that keeps working either way.

Measured on Claude Code 2.1.276, inside a worktree session: `env -u FOO echo ok` and `env -u FOO git rev-parse --short HEAD` both ran; `( unset FOO; echo ok )` ran; `( unset FOO; git rev-parse --short HEAD )` and `git -C . rev-parse --short HEAD` were both refused. The last one is the control — it shows the guard was live while `env` was passing.

Moving the command into a script file is not a workaround — the guard cannot read inside a script, so every check is bypassed for that payload. Where a verification cannot be expressed as one compound invocation, reduce the verification rather than route it around the guard.

**A verification recipe that spawns processes is itself a hazard, and gets reviewed as one.** The fault belongs to the dispatcher who wrote and approved the recipe, not to the lane that ran it as given.

## Integration into the release branch is self-served

Relocated from the stub: the full window procedure behind the stub's one-paragraph rule. Trigger affinity: executed inside the integration window — the `**/.moai/state/integration/**` trigger this file is scoped to is the window's own state directory.

Two measured constraints make the lane enter the release worktree rather than drive it remotely: git checks one branch out in exactly one worktree, and the worktree-session guard refuses a cross-tree `git -C`.

- **One integration surface.** The release branch lives in exactly one worktree — the one the lead provisioned; a lane never checks it out in its own tree.
- **Enter, do not redirect.** The lane switches in with `EnterWorktree(<release-worktree-path>)` and runs a plain `git merge --no-ff <WT-branch>` there. A cross-tree `git -C <release-worktree> merge` is refused by the worktree-session guard; entering is the sanctioned path.
- **Return the same way.** `ExitWorktree` returns to the primary checkout, not to the lane's own worktree — the lane re-enters its card worktree with `EnterWorktree(<own-path>)` before continuing.
- **One integrating session at a time — and an empty `MERGE_HEAD` does not establish that.** The release worktree is the serialization point. `git rev-parse -q --verify MERGE_HEAD` printing nothing is NECESSARY, never sufficient: it prints nothing just as readily while another lane is mid-resolution. Reading that silence as "the tree is free" is what lets two lanes overlap, invisibly until one commits.

    [HARD] **Serialize by the recorded hold and the announcement, not by probe.** `moai integration acquire` records the hold, `moai integration status` says who has it, `moai integration release` gives it back when the completion report is sent. Taking a live holder's window needs `--force`, which records what it displaced — deliberate, never quiet. The recorded hold is what the PreToolUse guard reads to refuse a second lane's `git merge`; the deny layer is opt-in (`workflow.integration_lock.enabled`, default off), the record works either way. The announcement to the lead rides alongside it. The probe stays — but it is the last check, never the first.

    [HARD] **`acquire` asserts the caller's tree first.** Detail: `kanban-dispatch-detail.md` § The pre-merge settings-drift assertion.

    [HARD] **Re-read `HEAD` immediately before the commit and again before the push.** `AGENTS.md` §2 binds this everywhere; the release worktree is where it has already earned its keep.

- **Conflicts belong to the lane that owns the change.** The integrating lane resolves what its own merge raises. A conflict it cannot resolve — a semantic clash with another lane's merged change — is a blocker report to the lead, not a forced merge.
- **Push the release branch; the batch pull request stays with the lead.** A rejected push means another lane pushed first — fetch, integrate, push again; never force. Until that branch's batch PR merges, the disposal rule above still binds.

The completion signal is the branch name, merge SHA, and evidence path.

## Boundaries

- **No board state store.** The queue is a plain file; column position is held by the leader within a card's run and re-derived from SPEC status after a clear. Persistent board state, per-card worktree lifecycle, WIP limits, and card/frontmatter reconciliation are separate work, not assumed here.
- **No session spawning.** The leader addresses sessions the operator launched — it never creates one (sub-agents are not sessions).
- **A role with no live session is a fault, not a wait.** The leader reports the empty role and the waiting card; silently holding it presents as a hang, the most expensive failure shape to diagnose. Empty means a complete listing showed none; an incomplete one is a gap, not an empty role (detail companion).

## Cross-references

- `.claude/rules/moai/core/agent-common-protocol.md` § Blocker Report Format — what a companion returns when it cannot proceed
- `.claude/rules/moai/workflow/worktree-integration.md` — the L1/L2 worktree tiers, their lifetimes, and the disposal contract
- `.claude/skills/moai/workflows/gtd.md` — the backlog queue surface
- `.claude/agents/moai/manager-lead.md` — the coordination agent the leader session works through
