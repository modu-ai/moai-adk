# Kanban Dispatch Protocol

How the **lead** session of Kanban Mode moves a card across the board: what admits work, who is told to do it, how completion is judged, and when the operator is asked to `/clear`.

> **Loading scope**: Intentionally always-loaded. A session learns it is the kanban lead from the SessionStart context, not from a file path, so a `paths:`-restricted rule would never reach the session that needs it.

> **Lead-only detail**: The board, Entry into the board is an operator act, Card classes, The dispatch cycle, Review lens selection, The `/clear` handoff between phases — moved to `kanban-dispatch-detail.md`; load when moving a card between columns, classifying a card, or choosing review lenses.

## Scope — when this rule is live

This rule binds a session whose SessionStart context declares **Kanban Mode** with the `lead` role. In every other session it is inert: a companion session (`plan` / `run` / `review` / `sync`) receives instructions, it does not dispatch them, and a session outside Kanban Mode has no board to move.

Kanban Mode is entered with `moai cc -k` (or `moai glm -k`), which elects one lead and prints one launch command per companion role. Companion sessions are launched **by hand, one per terminal** — a session cannot launch another session, and no mechanism to spawn a peer exists or is wanted.

## Completion is read, never trusted

[HARD] The lead advances a card on **evidence it read**, not on a companion's reply. Reply routing between sessions is not guaranteed to arrive, and a reply is a claim rather than an observation.

Before moving a card out of a working column, the lead reads the card's `progress.md` (and, where the phase declares one, the verification evidence path it cites) and confirms the phase actually closed. A missing, unreadable, or stale evidence file is a **gap** — the card stays where it is and the lead reports why. Absence of a failure signal is not a pass.

**Class B skips `plan`, not `review`** — an unestablished cause is precisely what review catches, so it is the last column to drop. The `run` session owns both the investigation and the fix. Before the card leaves `run`, the evidence that established the cause — the command that reproduced the defect and what it printed — is written into the card's progress record, and the run session names that path in its completion report so the lead reads the cause rather than taking it on trust.

This applies equally to the operator: when the lead reports a column advanced, it names what it read.

### CodeRabbit is not read from `gh pr checks`

[HARD] A `gh pr checks` row naming CodeRabbit is not evidence that a review happened. CodeRabbit reports through a commit status whose `state` is `success` **even when no review ran**, so the row prints `pass` in both cases and is byte-identical between them. The only field that separates a reviewed pull request from an unreviewed one is the status description.

A CodeRabbit row counts as evidence only when BOTH of these hold:

1. The status is `success` **and** its description reads `Review completed`:

    ```bash
    gh api "repos/$REPO/commits/$HEAD_SHA/status" \
      --jq '.statuses[] | select(.context == "CodeRabbit" and .state == "success") | .description'
    ```

    Both halves are required, and neither is sufficient alone. `success` is not sufficient — that is this whole section's point, since it appears on unreviewed heads too. But it is still necessary: without the state filter a `failure` or `error` status carrying a `Review completed` description would read as a pass, which inverts the same mistake in the other direction.

    This predicate assumes the **combined** status endpoint, `/commits/{sha}/status`, which returns only the most recent status per context — measured on this repository, exactly one CodeRabbit entry per head. That assumption is the load-bearing part, so it is stated rather than left implicit: do not substitute the plural `/commits/{sha}/statuses`, which returns the full history newest-first. Measured on one head there: five CodeRabbit entries running from `Review queued` through `Review completed`, so a positional pick on that endpoint is wrong in one direction or the other — `last` selects the oldest. Where history is genuinely wanted, select by maximum `created_at` rather than by position.

1. A `Merge Risk:` line exists whose `` up to `<prefix>` `` matches the current `headRefOid`, so the verdict covers the head being merged rather than an earlier commit.

Anything else is a gap, not a pass. `Review rate limited` in particular means the review never started, and a card carrying it does not leave `review` or `sync`.

Branch protection is not the lever here. The status state is `success` in precisely the failing case, so adding CodeRabbit to the required contexts would admit the unreviewed pull request just as readily. The distinction lives in the description, and only a read of the description surfaces it — which is why an automated merge gate closing this hole does not close it on the path a human merges by hand.

## Isolation is entered, never provisioned

[HARD] A card's work happens inside a worktree, and that worktree is **entered through the launcher** — never created with a bare `git worktree add`.

| Need | Form |
|---|---|
| Work inside the worktree in this session | `moai cc -w <name>` |
| Open it in a new window, keeping this session | `moai cc -w <name> --spawn` |
| Re-enter one from the current session | `EnterWorktree(<path>)` |
| Leave it | `ExitWorktree` |
| Dispose it once the card's pull requests have merged | `moai worktree done` |

`moai worktree` deliberately carries no creation verb — its own help states that entering is the launcher's job. A tree made with a raw `git worktree add` is one git knows about and MoAI does not, so `done`, `clean`, and `recover` have nothing to close, and orphaned trees accumulate until someone reconciles them by hand.

The lead dispatches this rather than assuming it. Each instruction names the worktree the companion is to work in and says to drive it with `git -C <path>` rather than `cd` — a `cd` inside a compound command changes the directory for that invocation only, so the next command silently reads the wrong tree.

Two properties make the shared checkout the wrong place for a card:

- Several sessions read it at once, so a branch switch, a `git stash`, or a `git add -A` there sweeps another session's uncommitted work into a commit that was never meant to carry it.
- A card outlives a phase. Its worktree spans run through sync, which is why disposal is triggered by the merge rather than by the phase finishing.

Where a companion reports having worked in the shared checkout instead, that is a fault to report, not a detail to tidy up afterwards.

## Verification load is lane-local

Sessions share one machine as surely as they share one checkout, and verification is where that sharing goes wrong. Measured on a day when it did: load average reached 413, a neighbouring workspace's build took two and a half minutes, and its browser tests timed out — not because anything was wrong with them, but because four lanes were each running the full test suite at once and a full suite there takes five to ten minutes.

[HARD] **Lane-local verification is scoped to the card.** A lane runs the tests its own change can affect, then pushes and lets CI run the full suite.

CI is not the fallback here; it is the better evidence. It runs the full suite in a clean environment against the actual pull-request head. A full-suite run on a loaded developer machine measures the machine — a test that fails in a crowded batch and passes alone has told you about contention, not about the code, and reading it as a code signal sends the lane chasing a defect that is not there.

[HARD] **Never spawn background load.** The same incident had a second cause: a verification recipe started eight spin loops to test behaviour under CPU contention and placed its kill line *after* the long test command. The agent finished before reaching it, and twelve spinners ran orphaned for thirty-seven minutes.

Where a verification genuinely needs contention, the load must be cleanup-guaranteed — kills registered with the test framework's cleanup hook, or a `timeout` wrapper that bounds the process from outside. A trailing `kill` is not cleanup; it is a line the process may never reach, and every path that ends early leaves the load running.

**A verification recipe that spawns processes is itself a hazard, and gets reviewed as one.** The fault above belongs to the dispatcher who wrote and approved the recipe, not to the lane that ran it as given — a lane executing an approved recipe is doing what it was told, and a rule that blames the executor teaches the wrong actor to be careful.

The same day supplied the reason contention and flakiness feed each other: a failing test left an unbounded spin-loop goroutine running, which burned a core for the remainder of that package's run and slowed every test after it. Load makes tests fail; failing tests can generate load.

### The env-isolated verification form

[HARD] Inside a worktree, an environment-scrubbed verification runs as one compound `unset … && <command>` invocation:

```bash
unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test ./...
```

The `env -u VAR <command>` form is rejected. The guard doing the rejecting lives in the Claude Code binary rather than in MoAI, and the rejection is an argument-boundary misparse rather than a rule against `-u`: when `argv[0]` is `env`, the guard scans the whole remaining argv as env's own flags, so a flag belonging to the **inner** command — `-run`, `-count`, `-race` — is reported as an unmodelled `env` flag. `unset` is not in the guard's wrapper set, so no flag scan opens and the inner command's flags are never inspected.

Two properties of the form are load-bearing, and both are easy to "simplify" away:

- **One invocation.** Each Bash call is a fresh process, so an `unset` issued as its own call does not carry into the next command. The scrub and the command travel together or the scrub does nothing.
- **No subshell.** Wrapping it as `( unset …; <command> )` trips the guard's "too complex to verify" refusal instead — a different rejection with the same effect.

Moving the command into a script file is not a workaround — not as a standing pattern and not as a one-off. The guard cannot read inside a script, so the entire check set is bypassed for that payload, including the git-redirect checks that are the guard's actual purpose and the part worth keeping.

Reading the script first does not restore them. A human read is not the guard running: the checks stay bypassed however carefully the file was reviewed, and a review performed once says nothing about what the file contains the next time it is invoked. Where a verification cannot be expressed as one compound invocation, that is a signal to reduce the verification, not to route it around the guard.

## Boundaries — what this protocol does not do

- **No board state store.** The queue is a plain file; column position is held by the lead within a card's run and re-derived from SPEC status after a clear. Persistent six-column state, per-card worktree lifecycle, WIP limits, and card/frontmatter consistency reconciliation are separate work and are not assumed here.
- **No spawning.** The lead addresses sessions the operator launched. It never creates one.
- **No gate bypass.** Kickoff approval before run-phase entry, and every other approval gate, is unchanged by being inside a dispatch cycle.
- **No question delegation.** Companion sessions return blocker reports; the operator is asked by the lead, through `AskUserQuestion`.
- **A role with no live session is a fault, not a wait.** Where the role owning a dispatchable card's column has no occupant, the lead dispatches to nobody, and reports both the empty role and the waiting card. Silently holding the card presents as a hang, which is the most expensive failure shape to diagnose.

## Cross-references

- `.claude/rules/moai/core/askuser-protocol.md` — the question channel the lead uses for card selection and `/clear` prompts
- `.claude/rules/moai/core/verification-claim-integrity.md` — why completion is read rather than trusted
- `.claude/rules/moai/core/agent-common-protocol.md` § Blocker Report Format — what a companion returns when it cannot proceed
- `.claude/rules/moai/workflow/worktree-integration.md` — the L1/L2 worktree tiers, their lifetimes, and the disposal contract
- `.claude/skills/moai/workflows/todo.md` — the backlog queue surface
- `.claude/agents/moai/manager-kanban.md` — the coordination agent, including its kanban-lead role

---

Classification: Evolvable operational rule — applies to the lead session of Kanban Mode.
Version: 1.1.0 — split into stub + detail companion
