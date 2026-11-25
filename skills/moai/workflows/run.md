# RUN

## Input

An approved SPEC, a plan, a base, a write scope and an owner. RUN follows the SPEC as `internal/sdd` says: only its Scope, and never while a `[NEEDS CLARIFICATION …]` question is open. The order of the stages is the one `internal/sdd` gives.

For the checks: an exact tree, the SPEC and policy digests, and a check inventory.

## Output

A patch, receipts from actual verification, and the remaining gaps. Each check in a receipt is marked passed, failed, not_run, blocked, stale or unknown.

## Procedure

1. Preflight: run `moai card check <card> --generation <g> --path <p>` for each path to write, and write only on `allowed: true`. Read the base and `git status`. Builds, slots and cleanup: `internal/build-slots`.
2. Read only the sources the task needs: gather them with `context-bundle` under a budget, re-check with `context-verify` before citing, and read a `stale` range again. A web page: `internal/web-pages`.
3. Write the test first and run it red for the reason you expect (`internal/test-design`), recording runs as `internal/tdd-results` says. For a change to existing code (preserve-then-change), first commit characterization tests that pass on the base and pin the behaviour the change must keep, then make the change and show they still pass. Add a test to a file of its own under `tests/`. A check, gate or criterion added or judged: `internal/check-completeness`. A boundary between two parts crossed: `internal/boundary-checks`. Authentication, input, secrets, outgoing requests or an interface others call: `internal/security-checklist`.
4. Hand each planned change to `moai-impl`, one at a time; change only the allowed scope.
5. Check commands come from `internal/quality`. Run every `verify:` command of the SPEC's acceptance criteria on the exact tree, recording command, start, end and exit code; SYNC writes them into the SPEC's `## Result` before the audit. Who runs which check, and citing a remote run: `internal/check-rules`.
6. Run each line of `internal/self-check` that fits; any other goes under `Not run` with the reason.
7. Run `moai check affected --card <base>..<head>` once, at READY, on the checkout at the head; never run it again on a tree that already passed it; the wide set (a bare `moai check affected`) belongs to the lander's `moai card land-batch`.
8. Hand the verified exact head to SYNC on the same branch, and only then to a separate AUDIT.

## When a step stops

| What happened | Next action |
| --- | --- |
| A check or test is red | Read its log, fix, run again. Fails once, passes on a rerun of the same head: flaky, not fixed (`internal/check-rules`). |
| A run selected no test (`0 passed`) | `not_run`, never a pass. |
| A tier 2 guard has no mutant | Add its `guard-mutant:` line to the READY note (`internal/self-check`). |
| `moai card check` does not answer `allowed: true` | Write nothing and report the answer; when the path is needed, add a `LANE-BLOCKED` note (`moai card note`). |
| A slot exits 75 | `not_run`; report it to the leader, never lower the limit (`internal/build-slots`). |
| The receipt fails | Read the failing step's log, fix on the same branch, commit, run the receipt once on the new head. |
| The base moved | Keep the base your checks ran on; the lander deals with it once. |
| `moai-impl` ends `blocked` or its `base` differs from `git rev-parse HEAD` | Read its note, fix the cause, hand the change again. |
| An attempt stalled or was interrupted | Send nothing again: read owner and effects back with `claim-show` and `effects-reconcile`; resume only on what the records prove. |

## Base agents

Skip this section when you are a base agent yourself: base agents never start other agents.

`moai` installs five base agents into the host's user folder: `moai-plan`, `moai-research`, `moai-impl`, `moai-audit` and `moai-docs`. When this route hands a step to one of them:

- Give it one block, `task`, `scope`, `base` and `refs`, plus `head` where the route's own text below asks for it, and nothing else; it assumes no conversation context.
- Decide on its first line, `RESULT <role> <ok|blocked|not_run> <target>`, then check the rest yourself: the files it lists and the commands and exit codes it reports. An agent's `ok` is not evidence.
- Start agents only from this session, one level deep.
- Use only what the host itself offers. Never stand in for a missing feature with extra sessions, folders or scripts: name the feature as not supported on this host, with the reason, and use the fallback given below.
- When the agent is not installed on this host, do the step in this session and say so.

RUN hands every planned change to `moai-impl` with the host's agent tool; do not write the change yourself when the agent is installed. Start the next change only after its `RESULT` line is in. Never run two `moai-impl` at once; only `moai-research` runs several at once.

- Claude Code: the installed agent file sets `isolation: none`, so it works in this worktree. `isolation: worktree` would make a second worktree from the remote default branch and build it from nothing. Pass `base`; the agent confirms `git rev-parse HEAD` equals it before writing and ends `blocked` when it does not.
- Claude Code runs a spawned agent in the background and sends a notice when it finishes; wait for that notice before you read its result or start the next step.
- Cleanup: the agent deletes only the build output it made in the worktree it works in, from inside that worktree and by relative path (`rm -rf target/debug`), never by an absolute path, and reports in its `Cleanup` section what it could not delete.

## Resume

Read the state afresh: `git status` and `git log`, `moai card show <card>` for the reservation and its `generation`, and the SPEC's acceptance criteria. Continue at the first criterion whose `verify:` command has no exit code on the current tree. Receipts of an older tree, base or SPEC are stale; nothing already sent is sent again.

## Exceptions and Permissions

Never lower acceptance to hide a failure. An agent's statement or another worktree's result is no evidence. The session that wrote a change never audits it. The auditing session records its own verdict with `moai card verdict`; the author never records it.

Before adding a dependency, look its exact name and version up in its registry and cite the output. An automatic CI repair runs at most three rounds as new commits without a force-push; it never patches a semantic failure and never edits secret, credential or CI-definition files.

A command that was not run is `not_run`, never a pass. A missing required guard is `blocked`. A HOLD or a `guarded` choice is reported as blocked and never decided by a lane. Filter one target per `cargo test` call: a test-name filter applies to every target.

A refusal by a guard or a hook, and platform-specific code, checker changes and manifest conflicts, follow `internal/shell-guards`.

## Completion Check

Bind the output to the task and the exact snapshot. Never report a verification or external effect that was not performed as complete.
