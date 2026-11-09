# PLAN

## Input

The user's goal, a repository snapshot, existing Issues, and constraints.

## Output

A GEARS SPEC, REQ/AC traces, an approval digest, and a task DAG.

## Procedure

1. Separate established facts from the assumptions that are needed.
2. Write the GEARS conditions, states, events, actors, and actions. Read legacy EARS through a meaning-preserving import.
3. Link each AC to its actual means of verification, risks, write scope, and dependencies.
4. Distinguish an Issue's draft from its approved snapshot.
5. Confirm scope, cost, data, and the required safety conditions, and hand only an approved revision and digest to RUN.
6. Record each judgment you make (intent, next step) with the internal `decision-record` operation: `decision_source: native`, `authority: none`. A judgment is a candidate next step; when it points past the authorized permissions, cost, external transmission, or destructive effects, ask the user instead of acting.

## Exceptions and Permissions

Treat new requirements as amendments. Never mark a hard-to-test requirement or an unresolved safety boundary as approved on your own.

## Completion Check

Bind the output to the task and the exact snapshot. Never report a verification or external effect that was not performed as complete.
