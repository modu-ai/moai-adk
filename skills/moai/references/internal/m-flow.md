# Internal: m-flow

When to read: when a leader or a lane integrates, hands off or promotes work, and when `/moai project` records the project's collaboration flow. M-Flow is the default collaboration flow for a new project; GitHub Flow, GitLab Flow, Git Flow and a project's own flow stay valid choices, and an explicit choice already recorded in the project profile is never replaced by this default. SKILL.md, MODE-LEADER and MODE-AGENT point here instead of repeating these rules, so a change made here reaches all three.

## Roles

1. Only the leader orders the integration and arbitrates between lanes. The leader does not repair a conflict: it sends the conflict back to the lane that wrote the change.
2. Each lane works in its own worktree, in parallel with the others. A lane hands over a fixed commit SHA and its test results, and never writes to the integration reference itself.

## Where work is integrated

3. The integration point is a clean, separate local main, created from the Rust upstream (for the MoAI ADK itself, `origin/v2026.11.0`). The leader is its only writer, and the tests run on the real combined tree.
4. Keep three references apart: the Rust upstream, the local integration reference and the promotion reference. When two of them do not share a history, stop with an error: never force an integration, and never push straight to the Go `origin/main`.

## With and without a remote

5. A project with no GitHub or GitLab connection proceeds on local Git alone, and only the remote features wait. Once a remote is connected, push often to an approved working branch of the same lineage; a protected remote target takes a final pull request or merge request with the required security checks and an independent audit.

## Candidates and releases

6. A local macOS candidate and a public multi-OS release are separate: a public release waits for the Linux, macOS and Windows verification of one fixed SHA.

## What stays as it is

7. The current GEARS and acceptance criteria, ownership by claim comment, independent audit, the pull request and CI gate of a shared branch, and the protection rules all stay. No sentence of this flow weakens one of them.
