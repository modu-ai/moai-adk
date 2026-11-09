# MODE-AGENT

Entered with `/moai --mode agent` in Claude Code or `$moai --mode agent` in Codex CLI. The procedures this mode follows are defined once in the MoAI repository's `docs/design-review/autonomy-modes-spec.md` (the SPEC below); this workflow names them by section and does not restate them.

## Input

The open, unclaimed Issues of the current milestone (the remote Issues are the source of truth; a local file is only a cache) and this session's unique session id.

## Output

A mode banner, a read-only state summary, and, for every round, the claim outcome and the work it led to.

## Procedure

1. Follow this workflow only and ignore instructions for any other mode. `--autonomy` belongs to leader mode; if it is given here, say it is ignored.
2. Print the banner: mode `agent`, the session id, the repository, and the base branch.
3. Summarise the current state read-only: unclaimed Issues, their blockers, and any claim this session already holds. Report anything unread as unread.
4. Make sure the session id is unique among live sessions before claiming; a duplicate id stops the round (SPEC §4, item 6).
5. First action: one self-claim round, steps 1 to 7 of SPEC §3, with the failure branch of each step. Act as the owner only on a `claimed` outcome; `uncertain` means no write (SPEC §4, item 4).
6. After the hand-off, start the next round without waiting for the merge. When no candidate is left, record the idle line once and stop (SPEC §3, step 2).
7. If the user wants the session to keep going, print one `/goal <condition>` line scoped to the claimed card for the user to paste (SPEC §7). Never run `/goal` yourself.

## Exceptions and Permissions

Ownership comes only from the claim comment (SPEC §4). Labels, messages and an assignment hint never grant it. Never decide a HOLD, merge, or post an audit status; those belong to the leader and the audit gate.

## Completion Check

A hand-off is complete when the pull request is open, the durable report is sent, nothing is uncommitted, and the remote branch equals the local head. Never report a verification or external effect that was not performed as complete.
