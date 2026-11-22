# NeXT-Flow git rule

- Work in a worktree under the project's `.claude/worktrees`, on its own branch; never in the primary checkout.
- A lane only commits: no push, no pull request, and no merge of a moved integration tip. A background session never pushes and never opens a pull request.
- A lane submits two things and nothing else: its head commit and the base its checks ran on. Add one `moai card note` whose first line is `LANE-READY-INTEGRATION <card> head=<40 hex>` and name the base in the lines below it.
- A card lands exactly once, by the leader: `moai card land <card> --head <40 hex> --by leader`. It moves the integration branch by one compare-and-swap and closes the card; a rerun never moves it twice.
- A protection hook refuses calls that move, copy, rename, delete or push the integration branch. Land with `moai card land`; never look for a way around the hook.
- After a landing the lander pushes the integration branch to the remote, fast-forward only.
- Remote checks never block a landing; a failed remote run becomes a repair card.

## Compact Instructions

When this conversation is summarized, keep:

- the card id and its generation
- the branch and the worktree path
- the last READY head and its base
- the open reservation: the paths held and when the lease ends
- the audit requests and decision notes still waiting for an answer

The project's own CLAUDE.md wins where it conflicts with this file.

Never force-push, run `git reset --hard` or `git clean -f`, pass `--no-verify`, or stage with `git add -A`.
