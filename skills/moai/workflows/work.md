# WORK

`/moai work` takes one card at a time through the whole flow and repeats until no card is left. Each stage is its own route; this text orders the stages and names the card commands between them.

## Input

The project's card ledger, read with `moai card next`, the SPEC each card names and the worktree each card is worked in.

## Output

For every card: the take outcome, what each stage produced, and either the landed head or the blocked note, each recorded as a note on the card (`moai card note`) and on the screen.

## Procedure

1. **take** — Run `moai card next` and take the card its `next` names: a card in `mine` first (it is already yours), else the first `ready` one. Reserve it with `moai card take <card> --scope <p> --worktree <w> --branch <b>` and keep the `generation` it answers; every later write of the card gives `--generation <g>`. Read the state afresh each time (`moai card next`) — never what you remember, so a restart, a usage-limit stop or a compacted conversation misses nothing. Ownership comes only from your own reservation; an empty or failed output is not a count.
2. **plan** — Pick the PLAN route for the card: its SPEC, the acceptance criteria and the task order. A card whose SPEC the user has not approved goes no further than the blocked note below.
3. **run** — Pick the RUN route in the card's worktree: the test that fails first, then the change and its checks. Before every commit run `moai card check <card> --generation <g> --path <p>` with every path the change touches; write only when it answers `allowed: true`. Any other answer (another holder, a stale generation, an ended lease, a path outside the scope) stops the write and is reported as it is.
4. **sync** — Pick the SYNC route on the same branch: the SPEC Result and the documents, and each criterion of the card marked with `moai card criterion <card> <n> --generation <g> --met --evidence <ref>`, or `--withdraw --reason <text>` for one that no longer applies.
5. **audit** — Pick the AUDIT route in a separate session on the exact head: the author of a change never audits it. A verdict other than a pass sends the card back to the run step.
6. **land** — Land the audited head with `moai card land <card> --head <40 hex>` and `--by leader`. The land command accepts only `--by leader` today, so the session that lands must be the one the project allows to land; any other session stops here and leaves the card with a note that names the head.
7. **repeat** — Go back to the take step until no card is left: `moai card next` names no card you can take. Then report the landed cards and the blocked ones, each with its reason, and end.

Where the project offers lane commands that run a card in a lane of its own, a step may run there; if the lane commands are missing, say so and run the step in the current session.

## Exceptions and Permissions

A card that needs a user decision never stops the loop: write `LANE-BLOCKED kind=decision <card> <the decision, its options and the default>` as a note on the card, never a question to the user, and take the next card. A HOLD or a `guarded` choice is reported the same way and never decided here. Without the `moai` command there is no card ledger: take nothing and report it once.

## Completion Check

Every card taken in this run is landed with its head named, or carries its blocked note with the reason. Never report a stage, a landing or a check that was not performed as complete.
