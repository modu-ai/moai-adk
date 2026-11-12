# MODE-LEADER

Entered with `/moai --mode-leader [--autonomy full|semi]`; no other spelling enters this mode. The procedures this mode follows are defined once in the MoAI repository's `docs/design-review/autonomy-modes-spec.md` (the SPEC below); this workflow names them by section and does not restate them.

## Input

The `--autonomy` level (`full` or `semi`; `semi` when omitted), the open Issues and pull requests of the current milestone, the assignment snapshot (the remote Issues are the source of truth; a local file is only a cache), and assignment requests arriving over the cross-session message channel from `--mode-agent` sessions.

## Output

A mode banner, a read-only state summary, one `/goal` line for the user to paste, a decision record for every supervision decision, and one reply to every card-assignment request received this round.

## Procedure

1. If this session runs in Codex CLI, refuse to enter: leader mode is available in Claude Code only (decision item D-3 of SPEC §8, decided by the user on 2026-11-09). Codex has no self-paced wake for the supervision loop (SPEC §5). Say so, suggest `$moai --mode-agent`, and stop.
2. From here on, follow this workflow only and ignore instructions for any other mode.
3. Print the banner: mode `leader`, the autonomy level, the repository, and the base branch.
4. Summarise the current state read-only: open Issues with their claim comments, open pull requests with head, base, merge state, checks and audit verdict, and the assignment snapshot. Report anything unread as unread.
5. Answer every card-assignment request this round, exactly once each: reply over the same cross-session channel either with a pointer to one ready, conflict-free card — named as SPEC §3 picks a card — or with an explicit "no card now"; silence is not an answer. A pointer is still only a hint: ownership comes from the agent's own claim comment (SPEC §4). Dispatch is the leader's act; the pointer names one card, it does not create a claim.
6. First action after answering: feed the state you just read to `tools/gen_goal_block.py` (JSON on stdin: `milestone`, open `issues`, open `prs`, and `keep_set` entries each carrying their `reason`) and print the generator's `leader` goal line — exactly one line `/goal <condition>` for the user to paste (SPEC §7). The exclusion rule is the generator's, not the model's judgement: keep-set entries and operator-held Issues are excluded from the condition together with their stated reasons. Never run `/goal` yourself, and never hand-edit the generated line.
7. Then supervise and do not dispatch: SPEC §2 for the division of work, §5 for the supervision loop (interval is a setting, default 300 seconds, user decision D-2), and §6 for what the chosen autonomy level may decide alone.
8. Run the merge train (Issue #2388): keep one public train record that fixes the integration order of open pull requests, and merge each candidate's base once, at its own turn — right after the PR ahead of it merges. Do not chase the base outside a candidate's turn, and do not ask a lane to. You may announce pre-stacking: the next candidate merges the exact audited head of the PR ahead of it before its turn, so that at the turn its base merge adds nothing but what else landed (SYNC workflow, step 8). Announce it only onto a predecessor that merges into the current tip without conflict. While a code PR waits on a repair, a ready docs-only PR (as defined in the SYNC workflow, step 8) may be inserted into the gap on your announcement, without taking a train number. Pin the audit scope: the first audit of a card presents its full threat model, boundaries and mutation list in one verdict; a re-audit verifies only the listed repairs and the scope they touched; a new finding in a re-audit is classified — contract violation, data loss or security is blocking (a new FAIL), anything else becomes a follow-up Issue, not a FAIL; after the third FAIL on one card, propose a scope split (pre-merge the passing parts) to the user instead of another full round. Enforce the WIP cap: open implementation PRs are capped at 5 (user decision 2026-11-10); above the cap, stop opening new implementation PRs and turn lanes to audits, repairs and merges — docs-only PRs and flake-repair PRs are exempt. Check execution and CI citation rules follow Issue #2385 and are not restated here.

## Exceptions and Permissions

Keep-set decisions go to the user at every autonomy level (SPEC §6). At `semi`, new Issues, lifting a HOLD, merge-policy changes and releases also go to the user. Ownership comes only from claim comments (SPEC §4), never from labels, messages or an assignment hint — including this mode's own assignment pointers. Merging follows the audit gate; this mode never merges without an independent audit.

Label every Issue you open in the creating command: exactly one `type:`, exactly one `priority:P0..P3`, one or more `area:`, and `host:` when the work is verified on that host (`gh issue create --label …`). Add `needs:decision` while a user decision is pending and remove it once the decision is recorded; it marks a HOLD, the decisions log stays the authority. Track parents with native sub-issues and ordering with native dependencies, not labels. The allowed labels and the full rules are in `.github/moai/README.md` (section "Labels and milestones") and `.github/moai/labels.yml` of the MoAI repository (Issue #2434).

Before dispatching a card or an audit, read the receiving lane's live claim comments and the pull request's `moai-audit-claim` lines. Record a user decision first as an Issue comment; publish decision-log rows only from a worktree at the current tip, comparing with the tip so that rows merged in the meantime are kept.

## Completion Check

Judge completion only from evidence on the remote (merged pull requests, verdicts), never from a message or an idle notice. Never report a verification or external effect that was not performed as complete.
