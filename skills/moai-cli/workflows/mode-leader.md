# MODE-LEADER

Entered with `/moai-cli --leader-mode [--autonomy full|semi]` (the legacy alias `--mode leader` selects this route too). The procedures this mode follows are defined once in the MoAI repository's `docs/design-review/autonomy-modes-spec.md` (the SPEC below); this workflow names them by section and does not restate them.

## Input

The `--autonomy` level (`full` or `semi`; `semi` when omitted), the open Issues and pull requests of the current milestone, the assignment snapshot (the remote Issues are the source of truth; a local file is only a cache), and assignment requests arriving over the cross-session message channel from `--agent-mode` sessions.

## Output

A mode banner, a read-only state summary, one `/goal` line for the user to paste, a decision record for every supervision decision, and one reply to every card-assignment request received this round.

## Procedure

1. If this session runs in Codex CLI, refuse to enter: leader mode is available in Claude Code only (decision item D-3 of SPEC §8, decided by the user on 2026-11-09). Codex has no self-paced wake for the supervision loop (SPEC §5). Say so, suggest `$moai-cli --agent-mode`, and stop.
2. From here on, follow this workflow only and ignore instructions for any other mode.
3. Print the banner: mode `leader`, the autonomy level, the repository, and the base branch.
4. Summarise the current state read-only: open Issues with their claim comments, open pull requests with head, base, merge state, checks and audit verdict, and the assignment snapshot. Report anything unread as unread.
5. Answer every card-assignment request this round, exactly once each: reply over the same cross-session channel either with a pointer to one ready, conflict-free card (its Issue number, the claim comment id to build `Moai-Claim` from, and the worktree rule) or with an explicit "no card now" — silence is not an answer. A pointer is still only a hint: ownership comes from the agent's own claim comment (SPEC §4). Dispatch is the leader's act; the pointer names one card, it does not create a claim.
6. First action after answering: compose the completion condition from that state and print exactly one line `/goal <condition>` for the user to paste (SPEC §7). Never run `/goal` yourself.
7. Then supervise and do not dispatch: SPEC §2 for the division of work, §5 for the supervision loop (interval is a setting, default 300 seconds, user decision D-2), and §6 for what the chosen autonomy level may decide alone.

## Exceptions and Permissions

Keep-set decisions go to the user at every autonomy level (SPEC §6). At `semi`, new Issues, lifting a HOLD, merge-policy changes and releases also go to the user. Ownership comes only from claim comments (SPEC §4), never from labels, messages or an assignment hint — including this mode's own assignment pointers. Merging follows the audit gate; this mode never merges without an independent audit.

## Completion Check

Judge completion only from evidence on the remote (merged pull requests, verdicts), never from a message or an idle notice. Never report a verification or external effect that was not performed as complete.
