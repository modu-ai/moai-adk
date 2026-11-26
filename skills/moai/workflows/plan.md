# PLAN

## Input

The user's goal, a repository snapshot, the card and its notes, and constraints.

## Output

A SPEC file as `internal/sdd` says — GEARS requirements, acceptance criteria with their verify commands, the user's recorded approval and its `approval_digest` — and a task order.

## Procedure

1. Separate established facts from the assumptions that are needed.
2. Write or update the SPEC (`moai spec new <slug> --title <title> --minimal` for a new one): the GEARS conditions, states, events, actors, and actions, with an IF…THEN failure path for every `When` requirement, and `[NEEDS CLARIFICATION: …]` where an answer is missing. Read legacy EARS through a meaning-preserving import.
3. Link each AC to its actual means of verification, risks, write scope, and dependencies.
4. Distinguish a draft SPEC from its approved revision.
5. Confirm scope, cost, data, and the required safety conditions. Run `moai spec check`, record the user's selection as the approval with the `approval_digest` that `moai spec show` prints, and hand only that approved revision to RUN.
6. Record each judgment you make (intent, next step) with the internal `decision-record` operation: `decision_source: native`, `authority: none`. A judgment is a candidate next step; when it points past the authorized permissions, cost, external transmission, or destructive effects, ask the user instead of acting.

Before reading the code a plan touches, check the code atlas as `internal/code-atlas` says: use it only when `moai index-query` reports it `fresh` (or for the files a `stale` answer does not name), search and read the files otherwise, and never build it here.

When the plan hinges on how a stateful part behaves, why a pipeline is slow, what of an old part is worth keeping or what a refactor could break, read `internal/lenses` and apply the lens that fits.

## Card branches

Read this section when the plan is for a card (`moai card show <card>`); a question that changes nothing needs neither branch.

1. **Sort.** Read the card's Write set and put the card in a tier by the audit contract (the `audit` reference): tier 2 when a path starts with a prefix of its floor list or the change deletes, changes permissions, gates or HOME; tier 0 or 1 otherwise. When the tier cannot be told from the Write set, treat the card as tier 2 and write why in the SPEC.
2. **Tier 0 or 1: no SPEC.** The card's body and criteria are the design, and its criteria are the acceptance criteria. Write no SPEC and ask for no approval. Reserve the card on a branch whose name does not start with `spec/`, `card/<slug>-<yyyymmdd>`, the slug naming the work and never a card number: a missing `Moai-Spec` trailer fails `moai trailer-check` only on a `spec/` branch. The card lands alone with `moai card land`; `moai card land-batch` does not take a card that names no SPEC.
3. **Tier 2: SPEC and approval.** Follow the Procedure above: write the SPEC (`moai spec new`) with an `## Interview` of the choices the user made, put every open choice with its options and default in its `## Decisions`, and hand only the approved revision to RUN. A lane files the SPEC's digest and the open choices as a `LANE-BLOCKED kind=decision` note on the card and waits; the approval is the user's selection and is never recorded by the lane.
4. **Planning agent output.** When `moai-plan` returns, its `Write set` and `Tests that read changed text` become the card's scope (`moai card take --scope`, or `moai card scope` for a card already held), each of its `Shared spots` is written under `moai card line-lease`, and its `Current holders` name the cards to ask first. Each of its `Open questions` becomes an open-question marker (`internal/sdd` item 5) in the SPEC of a tier 2 card, and a `LANE-BLOCKED kind=decision` note on the card in tier 0 or 1 when nobody has answered it.

## Base agents

Skip this section when you are a base agent yourself: base agents never start other agents.

`moai` installs five base agents into the host's user folder: `moai-plan`, `moai-research`, `moai-impl`, `moai-audit` and `moai-docs`. When this route hands a step to one of them:

- Give it one block, `task`, `scope`, `base` and `refs`, plus `head` where the route's own text below asks for it, and nothing else; it assumes no conversation context.
- Decide on its first line, `RESULT <role> <ok|blocked|not_run> <target>`, then check the rest yourself: the files it lists and the commands and exit codes it reports. An agent's `ok` is not evidence.
- Start agents only from this session, one level deep.
- Use only what the host itself offers. Never stand in for a missing feature with extra sessions, folders or scripts: name the feature as not supported on this host, with the reason, and use the fallback given below.
- When the agent is not installed on this host, do the step in this session and say so.

PLAN hands the draft to one `moai-plan` with the host's agent tool; do not draft it yourself when the agent is installed. First fan out `moai-research` for every open fact, several at once in one message and one question each, as the READ workflow describes, and wait until every part's `RESULT` line is in. Then start one `moai-plan`, read-only, and wait for its `RESULT` line before RUN starts: the plan comes first, `moai-impl` starts after it and never beside it. The agent drafts steps and an AC table; approval and the decisions in steps 4 and 5 stay with this session and the user. Read-only is enforced only where the installed agent file can carry it; confirm afterwards that nothing changed.

## Resume

Read the SPEC file afresh, never from memory: run `moai spec check` and `moai spec show <id>`, and continue at the first step of the procedure whose output is missing. The recorded approval still covers the SPEC while its digest matches; only a change to the requirements, the acceptance criteria or the `forbidden:` entries needs a new approval.

## Exceptions and Permissions

Treat new requirements as amendments. Never mark a hard-to-test requirement or an unresolved safety boundary as approved on your own.

## Completion Check

Bind the output to the task and the exact snapshot. Never report a verification or external effect that was not performed as complete.
