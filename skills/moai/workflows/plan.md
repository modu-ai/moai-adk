# PLAN

## Input

The user's goal, a repository snapshot, existing Issues, and constraints.

## Output

A SPEC file as `internal/sdd` says — GEARS requirements, acceptance criteria with their verify commands, the user's recorded approval and its digest — and a task order.

## Procedure

1. Separate established facts from the assumptions that are needed.
2. Write or update the SPEC (`moai spec new <slug> --title <title> --minimal` for a new one): the GEARS conditions, states, events, actors, and actions, with an IF…THEN failure path for every `When` requirement, and `[NEEDS CLARIFICATION: …]` where an answer is missing. Read legacy EARS through a meaning-preserving import.
3. Link each AC to its actual means of verification, risks, write scope, and dependencies.
4. Distinguish an Issue's draft from its approved snapshot.
5. Confirm scope, cost, data, and the required safety conditions. Run `moai spec check`, record the user's selection as the approval with the digest `moai spec show` prints, and hand only that approved revision to RUN.
6. Record each judgment you make (intent, next step) with the internal `decision-record` operation: `decision_source: native`, `authority: none`. A judgment is a candidate next step; when it points past the authorized permissions, cost, external transmission, or destructive effects, ask the user instead of acting.

Before reading the code a plan touches, check the code atlas as `internal/code-atlas` says: use it only when `moai index-query` reports it `fresh` (or for the files a `stale` answer does not name), search and read the files otherwise, and never build it here.

When the plan hinges on how a stateful part behaves, why a pipeline is slow, what of an old part is worth keeping or what a refactor could break, read `internal/lenses` and apply the lens that fits.

## Base agents

Skip this section when you are a base agent yourself: base agents never start other agents.

`moai` installs five base agents into the host's user folder: `moai-plan`, `moai-research`, `moai-impl`, `moai-audit` and `moai-docs`. When this route hands a step to one of them:

- Give it one block, `task`, `scope`, `base` and `refs`, and nothing else; it assumes no conversation context.
- Decide on its first line, `RESULT <role> <ok|blocked|not_run> <target>`, then check the rest yourself: the files it lists and the commands and exit codes it reports. An agent's `ok` is not evidence.
- Start agents only from this session, one level deep.
- Use only what the host itself offers. Never stand in for a missing feature with extra sessions, folders or scripts: name the feature as not supported on this host, with the reason, and use the fallback given below.
- When the agent is not installed on this host, do the step in this session and say so.

PLAN may hand the draft to one `moai-plan` at a time, read-only, and may first fan out `moai-research` for open facts as the READ workflow describes. The agent drafts steps and an AC table; approval and the decisions in steps 4 and 5 stay with this session and the user. Read-only is enforced only where the installed agent file can carry it (not on Codex CLI); confirm afterwards that nothing changed.

## Exceptions and Permissions

Treat new requirements as amendments. Never mark a hard-to-test requirement or an unresolved safety boundary as approved on your own.

## Completion Check

Bind the output to the task and the exact snapshot. Never report a verification or external effect that was not performed as complete.
