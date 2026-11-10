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
5. First action: one self-claim round, steps 1 to 7 of SPEC §3, with the failure branch of each step. Act as the owner only on a `claimed` outcome; `uncertain` means no write (SPEC §4, item 4). A lock-busy answer (`ok: false`, kind `unavailable`, `retryable: true`, exit 1 — Issue #2345) is retried by the caller: wait a bounded interval (a fixed backoff under a hard cap of attempts), then resend the unchanged command; the tool itself never waits, and with the cap out the round reports `unavailable` and stops.
6. Before every commit and every push, run `moai-cli claim-verify` for the claimed Issue. Write only on the `owner` verdict. Any other verdict (`not_owner`, `stale`, `closed`, `no_claim`, `uncertain`) stops the write and is reported as it is. An expired lease is `stale`: never take the Issue over by time alone; wait for a handoff line or the operator. Where the host hook set is applied, the `PreToolUse` guard (`claim-verify --hook`) blocks a refused `git commit` or `git push` in either host; without it, this step is the boundary. When the claim was posted for you (no claim record in this clone), the guard reads the Issue from the branch `issue/<n>-…` and the claim from `Moai-Claim: <n>#<id>` in the commit message, so put the token in the `-m` message (a message from `-F` or an editor is not read) and make no commit before you have that `<id>`. A takeover (epoch n+1) needs the handoff line `moai-claim-handoff: comment=<earlier claim id> to=<session>` on the Issue; without it the earlier claim still holds and the verdict is `not_owner`. After the handoff, commits written under the earlier epoch's token still count on push, so the new claim needs no new base.
7. After the hand-off, start the next round without waiting for the merge. `moai-cli pr-confirm` moves the claim record to `pr_ready`; `moai-cli queue-next` then skips that card (reported as `skipped` with `reason: handed_off`) even though its Issue stays open until the merge, and claims the next free one. Leave the `pr_ready` record as it is: do not release it. It stays active so a repair handoff, worktree operations and `scope-check` still see the card as yours. After the merge the Issue is closed and no longer a candidate; the record itself is not changed by the merge. When no candidate is left, record the idle line once and stop (SPEC §3, step 2).
8. If the user wants the session to keep going, feed the claimed card's state to `tools/gen_goal_block.py` and print the generator's `agent` goal lines — card-scoped `/goal` lines, one per open Issue of the claimed card — for the user to paste (SPEC §7). The line comes from the generator, not from the model's own wording: never run `/goal` yourself and never hand-edit the generated lines.

## Exceptions and Permissions

Ownership comes only from the claim comment (SPEC §4). Labels, messages and an assignment hint never grant it. Never decide a HOLD, merge, or post an audit status; those belong to the leader and the audit gate.

## Completion Check

A hand-off is complete when the pull request is open, the durable report is sent, nothing is uncommitted, and the remote branch equals the local head. Never report a verification or external effect that was not performed as complete.
