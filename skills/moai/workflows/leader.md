# LEADER

`/moai leader` makes the window that types it the registered coordination window of the project; the window holds no standing role, and when it ends the lanes go on from the ledger alone. With no registered leader the flow ends at the submission: nothing is audited or landed, and no lane takes that over. Round after round it reads the ledger, answers the lanes, sends audits and lands audited heads.

`/moai leader` sets a mode and is not a command: text after it is never run, and the window says in one line, in the user's conversation language, that instructions come as a message. A window whose name does not already start with `leader-` (it was not started with `moai cc -l`) gives the user the one line `/rename leader-<project name>` once, in the user's conversation language, and reads its own name from the first line of `ListAgents` before it goes on. If that name is not the wanted one after the user's reply, or the user gives none, go on under the name the window has, address the lanes and be addressed by the short id, and write one ledger note that says so.

## Input

The project's card ledger (`moai card inbox`, `moai card wait`, `moai card next`), the registry entry of the leader (`moai session-register --show --name leader`) and the lanes' notes. Read the flow reference `flows/next-flow` before the first round: its "Completion message" and "Transition" sections say what a lane sends and what this window does with it.

## Output

For every round: the entry this window holds, the notes it answered, the heads it sent to audit and the heads it landed, each recorded as a note on the card (`moai card note`) and on the screen.

## Procedure

1. **register** — Read the entry with `moai session-register --show --name leader` first. When it answers `free`, register this session with `moai session-register . --name leader --mode leader` and keep the `instance` the answer returns. When it answers `held` and its `session` is this window's own host session id, register nothing again: renew that entry as step 2 says, with the `instance` the answer shows. When another session holds the entry, or the registration is refused with `conflict`, stop with one line that names the holder's lease end and register nothing; a refusal of any other kind is an error to report, not an answer to who leads.
2. **round** — Once per round, before you read the ledger, renew the registration with `moai session-register . --name leader --mode leader --instance <instance>`; the mode is stored only when given, so pass `--mode leader` with every renewal. Read the ledger with `moai card inbox`, `moai card wait` and `moai card next`; a lane's completion message is a notice that you check against the ledger before you act, as the "Completion message" and "Transition" sections of `flows/next-flow` say.
3. **dispatch** — Send the next card's dispatch with the host's message tool, or with `moai lane-send` when the lanes run in terminals that command can reach: it reads the screen first and sends only to a lane that is working or idle; no `moai` command sends a cross-session message yet. The dispatch is the `DISPATCH <card> ready head=<tip>` line and the scope (`flows/team-messages`); address a lane by its listed name, or by the short id when two rows share the name. When the host's channel is missing, refused or held, use the lane-terminal manager's messaging (its adapter reference names the commands), and when neither exists write the ledger note alone. Write each dispatch and answer in the user's conversation language; a message is a notice, so the receiving lane reads the ledger before it acts.
4. **audit** — Send the audit of a head to a session that wrote none of the change (the AUDIT route); the author of a change never audits it.
5. **land** — Land that head with `moai card land <card> --head <40 hex> --by leader` only after a verdict `PASS` for that exact head; the command accepts only `--by leader` today, so this window is the caller. A head that is not on the tip answers `stale_base`, which `flows/next-flow` ("Landing") tells how to merge once.
6. **end** — When the route ends, release the registration with `moai session-register . --name leader --release --instance <instance>`. A window that is closed releases nothing: its lease ends by itself (four hours by default), and the lanes then write the ledger note only. Until then the entry can be freed with `moai session-register . --name leader --release --by leader --instance <instance>`, taking the `instance` from `moai session-register --show --name leader`.

## Exceptions and Permissions

A `LANE-BLOCKED kind=decision` note is sorted in the order of the skill entry's paragraph that begins "Sort an open question", case by case: this window decides and records as a note on the card what that order leaves to the leader, and what it sends to the user goes to the user as a selection with the recommended option first and waits for the pick. A timer that repeats the round is armed only as `internal/loop` says. Without the `moai` command there is no registry and no ledger: register nothing and report it once.

## Completion Check

Every round names what it read, answered, sent and landed. A head is called landed only when `moai card land` answered so, and a message from a lane is never taken for a ledger state.
