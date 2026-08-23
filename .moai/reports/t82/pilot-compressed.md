# M1 pilot rewrite — compressed clause set

Pilot files per `plan.md` §E M1: `kanban-dispatch.md` (the high end — 23 markers) and
`native-idiom-and-register.md` (the low end — 2 markers). Only the Codex-relevant blocks are
rewritten; a Claude-only block never enters `AGENTS.md` and so has no compression to measure.

Each block below is the `AGENTS.md`-bound form. Delimiters are machine-readable so
`pilot_measure.py` can tally before/after bytes per block against `blocks.json`.

<!--B064-->
A `gh pr checks` row naming CodeRabbit is not evidence that a review ran: the status reads
`success` and prints `pass` identically whether or not one did. Count the row only when BOTH hold:
(1) `gh api "repos/$REPO/commits/$HEAD_SHA/status"` reports the `CodeRabbit` context with
`state == "success"` and description `Review completed`; (2) a `Merge Risk:` line exists whose
commit prefix matches the current `headRefOid`.
<!--/B064-->

<!--B066-->
Work inside a worktree, and enter it through the launcher (`moai cc -w <name>`, `EnterWorktree`);
never create one with a bare `git worktree add`. Leave with `ExitWorktree`. Dispose an L2 tree
(`~/.moai/worktrees/…`) with `moai worktree done`; an L1 tree (`.claude/worktrees/…`) is disposed
at session end and is never registered.
<!--/B066-->

<!--B067-->
`moai worktree done` closes L2 trees only. A tree under `.claude/worktrees/` is L1, is absent from
the registry, and is disposed by the session-end prompt or by `git worktree unlock` +
`git worktree remove`.
<!--/B067-->

<!--B068-->
A card's branch is unpushed, so its worktree holds the only copy of the work. Dispose no worktree —
L1 or L2 — until the branch is integrated and the remote merge has landed.
<!--/B068-->

<!--B069-->
Start a new card in a new worktree. Exit any previous worktree back to the primary checkout first,
or the new card's work lands on the old card's branch. Create the fresh tree from the remote
default branch; never reuse the previous card's tree. Where the new card depends on a prior card's
unmerged code, merge that branch inside the new worktree.
<!--/B069-->

<!--B070-->
Card worktree branches carry the `WT-` prefix and a descriptive slug, never the card id. Rename in
place immediately after creating the tree: `git branch -m WT-<slug>`. Re-entry resolves by tree
name, so the rename is safe.
<!--/B070-->

<!--B071-->
Because the branch name no longer identifies the card, three carriers are mandatory: the dispatch's
`card:` field, the card id in every commit message on the branch, and the card id in the evidence
path (`.moai/reports/<card-id>/verdict.md`).
<!--/B071-->

<!--B072-->
Scope verification to the change: run the tests the change can affect, then push and let CI run the
full suite. A full-suite run on a loaded developer machine measures the machine, not the code.
<!--/B072-->

<!--B073-->
Never spawn background load. Where a verification needs contention, the load must be
cleanup-guaranteed — kills registered with the test framework's cleanup hook, or a `timeout`
wrapper bounding the process from outside. A trailing `kill` is not cleanup.
<!--/B073-->

<!--B074-->
Inside a worktree, run an environment-scrubbed verification as one compound
`unset <VARS> && <command>` invocation; a separate `unset` call does not carry into the next
command.
<!--/B074-->

<!--B034-->
When `conversation_language ≠ en`, every user-facing surface — chat, reports, README, docs, and
generated sites — must read as natural native prose, not English mapped word-for-word.
Translation-style calques are prohibited; native idiom is required.
<!--/B034-->
