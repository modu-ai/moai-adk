# NeXT-Flow

When to read: before you take, run, audit or land a card, as the session that works through the cards (`workflows/work.md`) or as the session that lands them. The flow rule file in the config folder holds the short rules and points here; this is the one full description.

## The name

NeXT means New · engineering · eXperience · Together: people and agents build one product side by side, each in a place of their own. Its one line is "Work in parallel. Land once." Many lanes work at the same time; each change reaches the integration branch exactly once, through one lander.

## How the flow works

Four technical rules make the flow run. They describe how it works; they are not a reading of the name.

- **No tip chase.** A lane works from the base it started on. When the integration tip moves, the lane does not merge the new tip, rebase or start its checks again. The lander deals with a moved tip once, at landing.
- **Evidence before landing.** A head lands only with its evidence: the checks the lane ran on that head, with their exit codes, and a passing verdict from a session that did not write the change (`references/audit.md`). The landing command itself refuses a head whose SPEC check or card criteria do not pass at that head.
- **Exactly-once landing.** A card lands once. The landing command moves the integration branch by one compare-and-swap and closes the card in the same call; a rerun of the same call finds the head already inside the tip and never moves the branch twice.
- **A guarded tip.** The integration branch moves only through the landing command. A protection hook refuses every other change of it.

## A lane

- Work in a worktree under the project's `.claude/worktrees`, on its own branch; never in the primary checkout.
- A lane only commits: no push, no pull request and no merge of a moved integration tip. A background session never pushes and never opens a pull request.
- A lane submits two things and nothing else: its head commit and the base its checks ran on. It adds one `moai card note` whose first line is `LANE-READY-INTEGRATION <card> head=<40 hex>` and names the base in the lines below it.
- A decision the lane cannot make goes on the card as a blocked note (`workflows/work.md`); the lane takes the next card.
- A `LANE-READY-INTEGRATION`, READY, REPAIRED or MERGE-READY note and a `LANE-DONE` line on the screen hand a card back; none of them is done in the ledger. How the session is cleared and the next card starts goes in two stages.

### Transition (the leader and a lane-terminal manager today)

- The leader sends `/clear` to the exact existing window only after the ledger shows the lane's card done or released (a safe hand-back that keeps the checkpoint and the candidate head for the audit), and only when nothing is open: no card in review, no live reservation, no run of the lane's own, no unsaved change, no input or selection prompt on the screen and nothing unknown.
- A fixed delay after `/clear` proves nothing. The leader confirms that the new session is ready, sends the next line once, and confirms that session's reservation of the new card (the ACK). When the ACK is not seen it inspects the window and does not send again.
- A new window is opened only when the existing one cannot be recovered.

### Final flow

- `/clear` and a lane-terminal manager are not conditions of dispatch. The start signal is the card verified and landed on the local integration branch, then done with its reservation released in the ledger.
- The worker reserves the next eligible card once, and an automatic restart runs only through the existing HOLD, ownership, priority and safety checks.
- Waiting for an audit and the hand-back of a card are separate from done and start nothing. A worker clearing its own session is not documented as a guarantee.

## Landing

- A card lands exactly once, by the lander: today the leader session, the only caller the command accepts. The command is `moai card land <card> --head <40 hex> --by leader`.
- When the tip is not on the first-parent line of the head (the tip moved after the lane's base), the command refuses with `stale_base` and changes nothing. The lander then merges the lane's head onto the tip once, in a worktree of its own, with the tip as the first parent, and lands that merge. The lane does nothing.
- The protection hook is git's `reference-transaction` hook, which runs `moai card guard`. It refuses every create, move and delete of the integration branch that the landing command did not announce, a rename or copy onto its name included. Land with `moai card land`; never look for a way around the hook.
- After a landing the landing command pushes the integration branch to the remote, fast-forward only, and only when the user has turned the remote copy on: the `remote_copy` setting, off by default, a key of the project profile. While it is off nothing is pushed and the remote is not read. A copy that could not finish waits for the next landing, and a remote whose branch went another way is held as it is; neither undoes the landing.
- Remote checks never block a landing; a failed remote run becomes a repair card.
- After a card closes, the landing command cleans that card's worktrees through the same plan and apply `worktree-gc-plan` and `worktree-gc-apply` use, on the plan's own digest, and puts the outcome in the `cleanup` key of its answer. Nothing new becomes removable: the safe criteria, the preserved `.claude/worktrees` folder and the fixed slots hold, so a worktree there is listed with its reason and kept. A card that stays open is not cleaned, and a cleanup that cannot run never undoes the landing; read its `error`.
- When a branch is deleted, the guard hook only appends its ref to the queue `moai/worktree-cleanup-queue` in the common git folder. The next cleanup reads the queue, drops a ref that still exists and cleans the worktrees of the cards that held a ref that is gone. The verdict step has the same function (`worktree_cleanup::run` with `Trigger::Verdict`) and is not wired to a command yet.
- Nobody moves the integration branch by hand, by a reset or by a push to it, and a landing never names a head nobody audited.
