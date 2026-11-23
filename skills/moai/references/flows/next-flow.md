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
- The READY note carries one line per new guard test in its body: `guard-mutant: <test> | <code removed or flipped> | exit <n>`, where `<n>` is the non-zero exit code the test gave on a copy that lacks the guarded code, or `guard-mutant: none | <reason>` when the change adds no guard test. A guard test that passes with its code removed is not a guard (four times on 2026-11-23); the lane runs that mutant once per new guard test before it reports READY. `moai bundle check` lists a missing, unreadable or `exit 0` line as a warning; it cannot tell whether a line is true, so the auditor reads it.
- A decision the lane cannot make goes on the card as a blocked note (`workflows/work.md`); with no registered leader the lane takes the next card, and with a registered leader it waits for that leader's dispatch.
- A `LANE-READY-INTEGRATION`, READY, REPAIRED or MERGE-READY note and a `LANE-DONE` line on the screen hand a card back; none of them is done in the ledger. How the session is cleared and the next card starts goes in two stages.

### Completion message (a lane to the leader)

- A lane writes its READY or REPAIRED note and gives the card back before it sends any message.
- When a registered leader holds a live lease, the lane then sends that leader one cross-session message that names the card and the head.
- The lane reads the leader's session id with `moai session-register --show --name leader` and addresses exactly that session; it never sends to a row picked by the name `leader`, because another project's session can carry that name. The lane matches that id to a row of the host's session list; with no matching row it sends nothing.
- Without a registered leader, a live lease or a cross-session channel, or when the send is refused or held, the lane writes the ledger note only; the leader's ledger watch picks the note up. One message per finished card, no retry loop. `LANE-DONE` is the lane's last line, after the message when one is sent.
- The message is a notice, not proof: the ledger decides.
- A lane never sends `/clear` to itself, and never takes its next card on its own while a leader is registered.
- No `moai` command sends this message yet: `moai host-wake` addresses a session by name, which the address rule forbids, so a lane uses the host's own message tool (SPEC-20261122-session-messages).

### Transition (the leader and a lane-terminal manager today)

- When the leader gets the message, it reads the ledger first, and sends `/clear` and then the next card's dispatch only when the ledger shows the card done or released and nothing is open.
- A message the ledger does not back starts nothing: the leader sends no `/clear` and no dispatch, and waits for the ledger.
- The lane's message and the leader's dispatch, answer and line after `/clear` are written in the user's conversation language; only a prompt to a subagent is English, and commands, paths, card ids and protocol lines stay verbatim.
- The leader sends `/clear` to the exact existing window only when nothing is open: no card in review, no live reservation, no run of the lane's own, no unsaved change, no input or selection prompt on the screen and nothing unknown.
- A fixed delay after `/clear` proves nothing. The leader confirms that the new session is ready, sends the next line once, and confirms that session's reservation of the new card (the ACK). When the ACK is not seen it inspects the window and does not send again.
- A new window is opened only when the existing one cannot be recovered.

### Final flow

- `/clear` and a lane-terminal manager are not conditions of dispatch. The start signal is the card verified and landed on the local integration branch, then done with its reservation released in the ledger.
- With no registered leader the worker reserves the next eligible card once, and an automatic restart runs only through the existing HOLD, ownership, priority and safety checks.
- Waiting for an audit and the hand-back of a card are separate from done and start nothing. A worker clearing its own session is not documented as a guarantee.

## Landing

- A card lands exactly once, by the lander: today the leader session, the only caller the command accepts. The command is `moai card land <card> --head <40 hex> --by leader`.
- When the tip is not on the first-parent line of the head (the tip moved after the lane's base), the command refuses with `stale_base` and changes nothing. The lander then merges the lane's head onto the tip once, in a worktree of its own, with the tip as the first parent, and lands that merge. The lane does nothing.
- Audit brief: name the scan tests `moai repo-scans --refs <tip> <head>` picks and run them on the merge tree; record each as a receipt step `scan:<id>` of the PASS, which `moai card land` reads.
- The protection hook is git's `reference-transaction` hook, which runs `moai card guard`. It refuses every create, move and delete of the integration branch that the landing command did not announce, a rename or copy onto its name included. Land with `moai card land`; never look for a way around the hook.
- After a landing the landing command pushes the integration branch to the remote, fast-forward only, and only when the user has turned the remote copy on: the `remote_copy` setting, off by default, a key of the project profile. While it is off nothing is pushed and the remote is not read. A copy that could not finish waits for the next landing, and a remote whose branch went another way is held as it is; neither undoes the landing.
- Remote checks never block a landing; a failed remote run becomes a repair card.
- Nobody moves the integration branch by hand, by a reset or by a push to it, and a landing never names a head nobody audited.
