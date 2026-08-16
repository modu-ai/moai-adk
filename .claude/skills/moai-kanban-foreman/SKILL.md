---
name: moai-kanban-foreman
description: >
  One unattended kanban foreman iteration: watch the backlog queue, dispatch
  the next operator-picked card to an isolated worker, collect completion
  evidence on read (not on claims), and report. This is the body the
  project's loop.md driver invokes each iteration of a bare /loop; it can
  also be invoked directly to test one cycle by hand.

when_to_use: >
  Use when a bare /loop kanban foreman iteration fires (the loop.md driver
  points here), or when the operator asks for a single manual foreman pass
  over the backlog queue.

license: Apache-2.0
compatibility: Designed for Claude Code
allowed-tools: Read, Grep, Glob, Bash(moai todo:*), Bash(git status:*), Bash(git log:*), Bash(git rev-parse:*), Bash(git diff:*), Bash(git show:*)
disallowed-tools: AskUserQuestion
user-invocable: false
metadata:
  version: "1.0.0"
  category: "workflow"
  status: "active"
  tags: "kanban, foreman, loop, backlog, dispatch, unattended"

# MoAI Extension: Progressive Disclosure
progressive_disclosure:
  enabled: true
---

# Kanban Foreman Loop Iteration

One unattended pass of the kanban foreman: watch the backlog queue, dispatch
the next operator-picked card to an isolated worker, collect completion
evidence, report. The queue surface is `moai todo`; the dispatch protocol and
card classes live in the kanban dispatch rule (`.claude/rules/moai/workflow/kanban-dispatch.md`).

## Running unattended

`AskUserQuestion` is removed from the tool pool while this skill is active —
that is the mechanical guarantee that the loop cannot stop and ask. Anything
that would have been a question becomes a line in the iteration report, and
anything that genuinely needs the operator's decision becomes a blocked card
under Boundaries below.

Deployment: start a session in the project, run bare `/loop`, then
background the session — loop tasks carry over to the background session and
keep running without a terminal. `Esc` cancels the pending wakeup of a
waiting loop. A recurring loop expires seven days after creation; restart it
when the board still needs a foreman. Background monitors do not survive a
session resume — the first iteration after a resume re-arms the queue watch.

The session's permission settings must already allow what this loop uses
(queue reads, git inspection, the worker spawn). A permission prompt that
surfaces while unattended stalls the iteration until someone attaches;
pre-approving that surface in project settings is the operator's setup step,
not something this loop can do for itself.

## Boundaries (hard)

1. **The operator admits and picks work.** Only backlog items whose state is
   already `picked` are dispatchable. Never run `moai todo add`; never run
   `moai todo next <n>` — that mutation is the operator's pick. Never invent,
   reword, or reorder cards. An empty queue is a legitimate state: say so and
   idle.
2. **No approval gate is answered on the operator's behalf.** When a card's
   next step needs a human decision that is not already recorded as made
   (plan-to-run kickoff approval, a review severity call, a scope choice),
   do not proceed. Leave the card `picked`, name it blocked-for-operator in
   the report together with the decision it waits on, and move on.
3. **One write-capable worker at a time.** While a worker is in flight the
   iteration only reads. Never run two write-capable agents concurrently.
4. **Every worker runs in its own worktree** (`isolation: "worktree"` on the
   spawn; relative paths in the prompt — the worker's CWD is its worktree
   root). Nothing writes to the shared checkout.
5. **No integration actions.** No push, no pull request, no merge, no branch
   deletion, no worktree disposal. The card's branch is unpushed and its
   worktree is the work's only instance; both stay until the operator
   integrates them. The report names the branch and the worktree path.
6. **Verification is lane-local.** The worker runs only the checks its own
   change can affect; the full suite belongs to CI. Never spawn background
   CPU load — the queue watch below is the only long-running process this
   loop arms.
7. **Completion is read, never trusted.** A card advances only on evidence
   this iteration actually read.

## The iteration

1. **Queue watch.** If no backlog monitor is live (first iteration, or after
   a resume), arm one persistent Monitor on the queue file:

   - `command`:

     ```sh
     f=.moai/state/kanban/backlog.json
     last=init
     while true; do
       if [ -f "$f" ]; then cur=$(cksum "$f"); else cur=missing; fi
       if [ "$cur" != "$last" ]; then
         [ "$last" != init ] && echo "backlog changed"
         last=$cur
       fi
       sleep 5
     done
     ```

   - `persistent: true`
   - `description: backlog queue watch`

   The queue file is replaced atomically on every mutation, so tools that
   follow a single file handle are the wrong shape here; the checksum poll
   emits one line per change and costs one tiny read every five seconds. Do
   not tighten the interval, and do not arm a second watcher. Each emitted
   line, like each scheduled wakeup, is a prompt to run this same idempotent
   iteration — an iteration that finds nothing to do ends quickly.

2. **Read the queue.** `moai todo list --json` (lock-free). A missing queue
   file is an empty queue, never an error. Records carry `id`, `text`,
   `spec_id`, and `state` (`queued` | `picked` | `dropped`).

3. **Collect before dispatch.** If a worker dispatched by an earlier
   iteration has returned, read its evidence file first (step 6). If a
   worker is still running, end the iteration with a one-line status.

4. **Choose the dispatchable card.** The oldest `picked` item with no live
   worker and no recorded blocker. `queued` items are not yours to pick.

5. **Dispatch one worker** with the Agent tool, `isolation: "worktree"`. The
   dispatch is a fixed-field address block — a pointer, not a copy, ten
   lines at most:

   ```
   card: <id>
   spec: <SPEC-ID>            # only when the card carries one
   cmd: <the card's work in one line; the phase command when a SPEC is attached>
   evidence: .moai/reports/<card-id>/evidence.md
   ```

   Route the spawn to the agent the card's work matches under the standard
   delegation rules; where no specialist matches, a general-purpose spawn
   with a domain whitelist. The worker prompt carries the block plus these
   standing orders: rename the worktree branch to `WT-<card-id>` first;
   implement the card; verify lane-locally; write the evidence file with
   decisions, verbatim output tails of the checks run, explicit gaps, and
   residual risk; commit by explicit pathspec; never push.

6. **Collect on evidence.** When the worker returns, read the evidence file
   it names. Advance the card — `moai todo done <t-id>` — only when the
   evidence shows the work complete: verbatim passing output present, gaps
   named. A missing, unreadable, or stale evidence file is a gap: the card
   stays `picked`, the report says why, and the card is not re-dispatched
   this iteration. Absence of a failure signal is not a pass.

7. **Report.** Close with two to six lines: queue summary, what was
   dispatched or collected, which evidence was read, blocked cards and the
   decisions they wait on. This report is what the operator reads on
   reattach — name what you read, not what you were told.

## Factory seam (reserved, not implemented)

The single-worker dispatch above is the only mode. Fanning a card out to
numbered factory worker lanes — the multi-lane launcher surface — is
separate work; when the foreman grows that routing, it lands here as a
second dispatch mode chosen per card. Until then this loop spawns one
subagent per card, reads no factory state, and launches no lanes.

## Failure handling

Stop the loop (`ScheduleWakeup` with `stop: true`), with a one-line reason,
when the loop cannot do its job: the queue file is repeatedly unreadable,
the queue watch cannot be armed, or this skill's own surface is broken. A
transient worker failure is not a loop failure — record it on the card and
let the next iteration decide whether to re-dispatch.
