# WORK

`/moai work` takes one card at a time through the whole flow and repeats until no card is left. Each stage is its own route; this text orders the stages and names the card commands between them.

## Input

The project's card ledger, read with `moai card next`, the SPEC each card names and the worktree each card is worked in. Read the flow reference `flows/next-flow` before the first take: how a lane works, what it submits and how a card lands.

## Output

For every card: the take outcome, what each stage produced, and either the landed head or the blocked note, each recorded as a note on the card (`moai card note`) and on the screen.

## Procedure

1. **take** — Run `moai card next` and take the card its `next` names: a card in `mine` first (it is already yours), else the first `ready` one. Reserve it with `moai card take <card> --scope <p> --worktree <w> --branch <b>` and keep the `generation` it answers; every later write of the card gives `--generation <g>`. Read the state afresh each time (`moai card next`) — never what you remember, so a restart, a usage-limit stop or a compacted conversation misses nothing. Ownership comes only from your own reservation; an empty or failed output is not a count. Read the project's flow and the integration branch from its profile with `moai work-flow-plan --project .` (its `profile` holds both; a profile with no `work_flow` names none) and name both in the first note you write on the card; never type their values from memory.
2. **plan** — Pick the PLAN route for the card: its SPEC, the acceptance criteria and the task order. A card whose SPEC the user has not approved goes no further than the blocked note below.
3. **run** — Pick the RUN route in the card's worktree: the test that fails first, then the change and its checks. Before every commit run `moai card check <card> --generation <g> --path <p>` with every path the change touches; write only when it answers `allowed: true`. Any other answer (another holder, a stale generation, an ended lease, a path outside the scope) stops the write and is reported as it is.
4. **sync** — Pick the SYNC route on the same branch: the SPEC Result and the documents, and each criterion of the card marked with `moai card criterion <card> <n> --generation <g> --met --evidence <ref>`, or `--withdraw --reason <text>` for one that no longer applies.
5. **audit** — Pick the AUDIT route in a separate session on the exact head: the author of a change never audits it. A verdict other than a pass sends the card back to the run step.
6. **land** — Land the audited head with `moai card land <card> --head <40 hex>` and `--by leader`. The land command accepts only `--by leader` today, so the session that lands must be the one the project allows to land; any other session stops here and leaves the card with a note that names the head; that card is handed back, not done.
7. **repeat** — Go back to the take step until no card is left: `moai card next` names no card you can take. A READY note or a `LANE-DONE` line is not the start signal: with no registered leader, take the next card once, when the ledger shows the previous card done or released (`flows/next-flow`); with a registered leader, wait for that leader's dispatch. Then report the landed cards and the blocked ones, each with its reason, and end.

When a card is handed back, write the READY or REPAIRED note and give the card back first; only then, with a registered leader, send it one cross-session message (`flows/next-flow`, "Completion message"). Read its session id with `moai session-register --show --name leader`, match that id to a row of the host's session list and address exactly that row, never a row picked by name; with no matching row, send nothing. With no leader, no live lease or no channel, write the ledger note only. Send one message per finished card, never a retry; `LANE-DONE` is your last line, after the message when one is sent. The message is a notice, not proof: never send `/clear` to yourself and never take the next card on your own while a leader is registered.

Before you write the READY, REPAIRED or blocked note, read `git status --short` in the worktree: it must print nothing, so commit what is uncommitted first and the head the note names holds it. After a READY or REPAIRED note you make no commit, because the note names the head its checks ran on. After the note, giving the card back and the message to the leader (when one is sent), leave the card's worktree with `ExitWorktree(keep)`, and never remove the worktree or move out of it with a shell `cd`. Write `LANE-DONE` as the last line of the turn, after that tool call. The next card starts outside every card worktree: enter the worktree the leader made with `EnterWorktree` and its `path`, or leave the current one first before you make a new one.

Where the project offers lane commands that run a card in a lane of its own, a step may run there; if the lane commands are missing, say so and run the step in the current session.

## Exceptions and Permissions

A card that needs a user decision goes to the ledger: write `LANE-BLOCKED kind=decision <card> <the decision, its options and the default>` as a note on the card, never a question to the user. A session that works a card never opens a selection prompt for the person; whatever it needs decided goes in that note. With no registered leader, take the next card; with a registered leader, leave the blocked card as its note and wait for that leader's dispatch. A HOLD or a `guarded` choice is reported the same way and never decided here. Without the `moai` command there is no card ledger: take nothing and report it once.

## Completion Check

Every card taken in this run is landed with its head named, handed back with the note that names its head (not called done), or carries its blocked note with the reason. Never report a stage, a landing or a check that was not performed as complete.
